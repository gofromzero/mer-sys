package middleware

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// MerchantPermissionMiddleware 商户权限检查中间件
type MerchantPermissionMiddleware struct {
	userRepo *repository.UserRepository
}

// NewMerchantPermissionMiddleware 创建商户权限中间件
func NewMerchantPermissionMiddleware() *MerchantPermissionMiddleware {
	return &MerchantPermissionMiddleware{
		userRepo: repository.NewUserRepository(),
	}
}

// RequireMerchantPermission 需要商户权限检查的中间件
func (m *MerchantPermissionMiddleware) RequireMerchantPermission(permissions ...types.Permission) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := r.GetCtx()

		// 从上下文获取用户信息（应该由认证中间件设置）
		userID := r.GetHeader("X-User-ID")
		merchantIDHeader := r.GetHeader("X-Merchant-ID")

		if userID == "" {
			r.Response.WriteJsonExit(g.Map{
				"code":    401,
				"message": "未认证用户",
				"data":    nil,
			})
			return
		}

		if merchantIDHeader == "" {
			r.Response.WriteJsonExit(g.Map{
				"code":    403,
				"message": "缺少商户信息",
				"data":    nil,
			})
			return
		}

		userIDInt, err := strconv.ParseUint(userID, 10, 64)
		if err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    400,
				"message": "用户ID格式不正确",
				"data":    nil,
			})
			return
		}

		merchantID, err := strconv.ParseUint(merchantIDHeader, 10, 64)
		if err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    400,
				"message": "商户ID格式不正确",
				"data":    nil,
			})
			return
		}

		// 检查用户是否属于该商户
		user, err := m.userRepo.FindMerchantUserByID(ctx, userIDInt, merchantID)
		if err != nil {
			g.Log().Warningf(ctx, "商户用户验证失败: %v", err)
			r.Response.WriteJsonExit(g.Map{
				"code":    403,
				"message": "无权访问该商户资源",
				"data":    nil,
			})
			return
		}

		// 获取用户的商户角色
		userRoles, err := m.userRepo.GetMerchantUserRoles(ctx, userIDInt, merchantID)
		if err != nil {
			g.Log().Errorf(ctx, "获取用户商户角色失败: %v", err)
			r.Response.WriteJsonExit(g.Map{
				"code":    500,
				"message": "权限检查失败",
				"data":    nil,
			})
			return
		}

		// 检查权限
		if len(permissions) > 0 {
			hasPermission := m.checkUserMerchantPermissions(userRoles, permissions)
			if !hasPermission {
				r.Response.WriteJsonExit(g.Map{
					"code":    403,
					"message": "权限不足",
					"data":    nil,
				})
				return
			}
		}

		// 将商户用户信息注入到上下文
		ctx = context.WithValue(ctx, "merchant_user", user)
		ctx = context.WithValue(ctx, "merchant_id", merchantID)
		ctx = context.WithValue(ctx, "merchant_roles", userRoles)
		r.SetCtx(ctx)

		r.Middleware.Next()
	}
}

// RequireMerchantRole 需要特定商户角色的中间件
func (m *MerchantPermissionMiddleware) RequireMerchantRole(roleTypes ...types.RoleType) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := r.GetCtx()

		// 从上下文获取用户信息
		userID := r.GetHeader("X-User-ID")
		merchantIDHeader := r.GetHeader("X-Merchant-ID")

		if userID == "" || merchantIDHeader == "" {
			r.Response.WriteJsonExit(g.Map{
				"code":    401,
				"message": "认证信息不完整",
				"data":    nil,
			})
			return
		}

		userIDInt, err := strconv.ParseUint(userID, 10, 64)
		if err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    400,
				"message": "用户ID格式不正确",
				"data":    nil,
			})
			return
		}

		merchantID, err := strconv.ParseUint(merchantIDHeader, 10, 64)
		if err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    400,
				"message": "商户ID格式不正确",
				"data":    nil,
			})
			return
		}

		// 获取用户的商户角色
		userRoles, err := m.userRepo.GetMerchantUserRoles(ctx, userIDInt, merchantID)
		if err != nil {
			g.Log().Errorf(ctx, "获取用户商户角色失败: %v", err)
			r.Response.WriteJsonExit(g.Map{
				"code":    500,
				"message": "权限检查失败",
				"data":    nil,
			})
			return
		}

		// 检查角色
		hasRole := false
		for _, roleType := range roleTypes {
			for _, userRole := range userRoles {
				if userRole == roleType {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			r.Response.WriteJsonExit(g.Map{
				"code":    403,
				"message": "角色权限不足",
				"data":    nil,
			})
			return
		}

		// 将信息注入到上下文
		ctx = context.WithValue(ctx, "merchant_id", merchantID)
		ctx = context.WithValue(ctx, "merchant_roles", userRoles)
		r.SetCtx(ctx)

		r.Middleware.Next()
	}
}

// checkUserMerchantPermissions 检查用户是否拥有指定的商户权限
func (m *MerchantPermissionMiddleware) checkUserMerchantPermissions(userRoles []types.RoleType, requiredPermissions []types.Permission) bool {
	// 获取所有角色的权限
	rolePermissions := types.GetDefaultRoles()
	
	var userPermissions []types.Permission
	for _, roleType := range userRoles {
		if role, exists := rolePermissions[roleType]; exists {
			userPermissions = append(userPermissions, role.Permissions...)
		}
	}

	// 检查是否拥有所有必需权限
	for _, requiredPerm := range requiredPermissions {
		hasPermission := false
		for _, userPerm := range userPermissions {
			if userPerm == requiredPerm {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			return false
		}
	}

	return true
}

// ExtractMerchantIDFromRequest 从请求中提取商户ID
func ExtractMerchantIDFromRequest(r *ghttp.Request) (uint64, error) {
	// 首先尝试从URL参数获取
	merchantIDStr := r.Get("merchant_id").String()
	
	// 如果URL参数没有，尝试从请求体获取
	if merchantIDStr == "" {
		merchantIDStr = r.Get("merchant_id").String()
	}
	
	// 如果还是没有，尝试从路径参数获取
	if merchantIDStr == "" {
		merchantIDStr = r.GetRouter("merchant_id").String()
	}
	
	// 最后尝试从Header获取
	if merchantIDStr == "" {
		merchantIDStr = r.GetHeader("X-Merchant-ID")
	}

	if merchantIDStr == "" {
		return 0, fmt.Errorf("未找到商户ID")
	}

	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("商户ID格式不正确: %v", err)
	}

	return merchantID, nil
}

// GetMerchantUserFromContext 从上下文获取商户用户信息
func GetMerchantUserFromContext(ctx context.Context) (*types.User, bool) {
	if user, ok := ctx.Value("merchant_user").(*types.User); ok {
		return user, true
	}
	return nil, false
}

// GetMerchantIDFromContext 从上下文获取商户ID
func GetMerchantIDFromContext(ctx context.Context) (uint64, bool) {
	if merchantID, ok := ctx.Value("merchant_id").(uint64); ok {
		return merchantID, true
	}
	return 0, false
}

// GetMerchantRolesFromContext 从上下文获取商户角色
func GetMerchantRolesFromContext(ctx context.Context) ([]types.RoleType, bool) {
	if roles, ok := ctx.Value("merchant_roles").([]types.RoleType); ok {
		return roles, true
	}
	return nil, false
}

// IsMerchantAdmin 检查是否为商户管理员
func IsMerchantAdmin(ctx context.Context) bool {
	roles, ok := GetMerchantRolesFromContext(ctx)
	if !ok {
		return false
	}

	for _, role := range roles {
		if role == types.RoleMerchantAdmin {
			return true
		}
	}
	return false
}

// IsMerchantOperator 检查是否为商户操作员
func IsMerchantOperator(ctx context.Context) bool {
	roles, ok := GetMerchantRolesFromContext(ctx)
	if !ok {
		return false
	}

	for _, role := range roles {
		if role == types.RoleMerchantOperator {
			return true
		}
	}
	return false
}

// HasMerchantPermission 检查是否拥有指定的商户权限
func HasMerchantPermission(ctx context.Context, permission types.Permission) bool {
	roles, ok := GetMerchantRolesFromContext(ctx)
	if !ok {
		return false
	}

	rolePermissions := types.GetDefaultRoles()
	for _, roleType := range roles {
		if role, exists := rolePermissions[roleType]; exists {
			if role.HasPermission(permission) {
				return true
			}
		}
	}
	return false
}
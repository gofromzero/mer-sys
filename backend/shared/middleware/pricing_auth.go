package middleware

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// PricingPermission 定价权限常量
type PricingPermission string

const (
	PricingPermissionCreate PricingPermission = "product:pricing:create"
	PricingPermissionRead   PricingPermission = "product:pricing:read"
	PricingPermissionUpdate PricingPermission = "product:pricing:update"
	PricingPermissionDelete PricingPermission = "product:pricing:delete"
	RightsPermissionManage  PricingPermission = "product:rights:manage"
	PriceChangePermission   PricingPermission = "product:pricing:change"
)

// RequirePricingPermission 创建权限验证中间件
func RequirePricingPermission(permission PricingPermission) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		// 获取用户ID和权限信息
		userID := r.GetCtx().Value("user_id")
		if userID == nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    401,
				"message": "用户未认证",
			})
			return
		}

		// 获取用户权限列表（从上下文或数据库）
		userPermissions, err := getUserPermissions(r.GetCtx(), userID)
		if err != nil {
			g.Log().Error(r.GetCtx(), "获取用户权限失败", err)
			r.Response.WriteJsonExit(g.Map{
				"code":    500,
				"message": "权限验证失败",
			})
			return
		}

		// 检查是否具有所需权限
		if !hasPermission(userPermissions, string(permission)) {
			g.Log().Warning(r.GetCtx(), fmt.Sprintf("用户 %v 缺少权限 %s", userID, permission))
			r.Response.WriteJsonExit(g.Map{
				"code":    403,
				"message": "权限不足",
				"required_permission": string(permission),
			})
			return
		}

		// 权限验证通过，继续处理
		r.Middleware.Next()
	}
}

// getUserPermissions 获取用户权限列表
func getUserPermissions(ctx context.Context, userID interface{}) ([]string, error) {
	// TODO: 实现从数据库或缓存获取用户权限
	// 这里提供示例实现，实际应该从用户权限表获取
	
	// 临时实现：为演示目的返回一些默认权限
	// 在实际项目中，这应该查询用户角色和权限表
	return []string{
		"product:pricing:read",
		"product:pricing:create", 
		"product:pricing:update",
		"product:rights:manage",
	}, nil
}

// hasPermission 检查用户是否具有指定权限
func hasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission {
			return true
		}
		// 支持通配符权限，如 "product:*" 包含所有product相关权限
		if permission == "product:*" && 
		   (requiredPermission == "product:pricing:create" ||
		    requiredPermission == "product:pricing:read" ||
		    requiredPermission == "product:pricing:update" ||
		    requiredPermission == "product:pricing:delete" ||
		    requiredPermission == "product:rights:manage" ||
		    requiredPermission == "product:pricing:change") {
			return true
		}
	}
	return false
}

// LogPricingOperation 记录定价操作日志
func LogPricingOperation(operation string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		userID := r.GetCtx().Value("user_id")
		tenantID := r.GetCtx().Value("tenant_id")
		
		g.Log().Info(r.GetCtx(), fmt.Sprintf("定价操作: %s, 用户: %v, 租户: %v, IP: %s", 
			operation, userID, tenantID, r.GetClientIp()))
		
		r.Middleware.Next()
	}
}
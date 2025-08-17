package middleware

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// AuthMiddleware 认证中间件配置
type AuthMiddleware struct {
	jwtManager     *auth.JWTManager
	roleRepository repository.RoleRepository
	publicPaths    []string
	skipPaths      []string
}

// NewAuthMiddleware 创建认证中间件实例
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:     auth.NewJWTManager(),
		roleRepository: repository.NewRoleRepository(),
		publicPaths: []string{
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/health",
			"/api/v1/health/ready",
			"/api/v1/health/live",
			"/api/v1/health/simple",
			"/api/v1/health/component",
		},
		skipPaths: []string{
			"/favicon.ico",
			"/robots.txt",
			"/static",
			"/assets",
		},
	}
}

// SetPublicPaths 设置公开路径
func (am *AuthMiddleware) SetPublicPaths(paths []string) *AuthMiddleware {
	am.publicPaths = paths
	return am
}

// AddPublicPath 添加公开路径
func (am *AuthMiddleware) AddPublicPath(path string) *AuthMiddleware {
	am.publicPaths = append(am.publicPaths, path)
	return am
}

// SetSkipPaths 设置跳过认证的路径
func (am *AuthMiddleware) SetSkipPaths(paths []string) *AuthMiddleware {
	am.skipPaths = paths
	return am
}

// JWTAuth JWT认证中间件
func (am *AuthMiddleware) JWTAuth(r *ghttp.Request) {
	ctx := r.GetCtx()
	path := r.URL.Path

	// 检查是否为公开路径
	if am.isPublicPath(path) {
		r.Middleware.Next()
		return
	}

	// 检查是否为跳过路径
	if am.isSkipPath(path) {
		r.Middleware.Next()
		return
	}

	// 提取Token
	token := am.extractToken(r)
	if token == "" {
		am.respondWithError(r, 401, "缺少访问令牌")
		return
	}

	// 验证Token
	claims, err := am.jwtManager.ValidateToken(ctx, token)
	if err != nil {
		am.respondWithError(r, 401, "令牌验证失败："+err.Error())
		return
	}

	// 检查Token是否在黑名单中
	isBlacklisted, err := am.jwtManager.IsTokenBlacklisted(ctx, token)
	if err != nil {
		g.Log().Errorf(ctx, "检查令牌黑名单失败: %v", err)
		am.respondWithError(r, 500, "令牌验证失败")
		return
	}

	if isBlacklisted {
		am.respondWithError(r, 401, "令牌已失效")
		return
	}

	// 将用户信息和令牌信息添加到上下文
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
	ctx = context.WithValue(ctx, "username", claims.Username)
	ctx = context.WithValue(ctx, "roles", claims.Roles)
	ctx = context.WithValue(ctx, "permissions", claims.Permissions)
	ctx = context.WithValue(ctx, "token", token)
	ctx = context.WithValue(ctx, "token_type", claims.TokenType)

	// 更新请求上下文
	r.SetCtx(ctx)

	// 继续处理请求
	r.Middleware.Next()
}

// PermissionAuth 权限检查中间件
func (am *AuthMiddleware) PermissionAuth(requiredPermission types.Permission) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := r.GetCtx()
		path := r.URL.Path

		// 检查是否为公开路径
		if am.isPublicPath(path) {
			r.Middleware.Next()
			return
		}

		// 从上下文获取用户信息
		userID := am.getUserIDFromContext(ctx)
		tenantID := am.getTenantIDFromContext(ctx)

		if userID == 0 || tenantID == 0 {
			am.respondWithError(r, 401, "用户认证信息不完整")
			return
		}

		// 检查用户是否拥有所需权限
		hasPermission, err := am.roleRepository.HasPermission(ctx, userID, tenantID, requiredPermission)
		if err != nil {
			g.Log().Errorf(ctx, "权限检查失败: %v", err)
			am.respondWithError(r, 500, "权限检查失败")
			return
		}

		if !hasPermission {
			am.respondWithError(r, 403, "权限不足")
			return
		}

		// 继续处理请求
		r.Middleware.Next()
	}
}

// RoleAuth 角色检查中间件
func (am *AuthMiddleware) RoleAuth(requiredRole types.RoleType) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := r.GetCtx()
		path := r.URL.Path

		// 检查是否为公开路径
		if am.isPublicPath(path) {
			r.Middleware.Next()
			return
		}

		// 从上下文获取用户信息
		userID := am.getUserIDFromContext(ctx)
		tenantID := am.getTenantIDFromContext(ctx)

		if userID == 0 || tenantID == 0 {
			am.respondWithError(r, 401, "用户认证信息不完整")
			return
		}

		// 检查用户是否拥有所需角色
		hasRole, err := am.roleRepository.HasRole(ctx, userID, tenantID, requiredRole)
		if err != nil {
			g.Log().Errorf(ctx, "角色检查失败: %v", err)
			am.respondWithError(r, 500, "角色检查失败")
			return
		}

		if !hasRole {
			am.respondWithError(r, 403, "角色权限不足")
			return
		}

		// 继续处理请求
		r.Middleware.Next()
	}
}

// TenantIsolation 租户隔离中间件
func (am *AuthMiddleware) TenantIsolation(r *ghttp.Request) {
	ctx := r.GetCtx()
	path := r.URL.Path

	// 检查是否为公开路径
	if am.isPublicPath(path) {
		r.Middleware.Next()
		return
	}

	// 从多个来源尝试获取租户ID
	tenantID := am.extractTenantID(r)
	if tenantID == 0 {
		// 尝试从JWT Token中获取
		userTenantID := am.getTenantIDFromContext(ctx)
		if userTenantID > 0 {
			tenantID = userTenantID
		}
	}

	if tenantID == 0 {
		am.respondWithError(r, 400, "缺少租户标识")
		return
	}

	// 验证用户是否属于该租户
	userTenantID := am.getTenantIDFromContext(ctx)
	if userTenantID > 0 && userTenantID != tenantID {
		am.respondWithError(r, 403, "跨租户访问被拒绝")
		return
	}

	// 将租户ID添加到上下文
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	r.SetCtx(ctx)

	// 继续处理请求
	r.Middleware.Next()
}

// RequirePermissions 要求多个权限的中间件（AND关系）
func (am *AuthMiddleware) RequirePermissions(permissions ...types.Permission) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := r.GetCtx()
		path := r.URL.Path

		// 检查是否为公开路径
		if am.isPublicPath(path) {
			r.Middleware.Next()
			return
		}

		// 从上下文获取用户信息
		userID := am.getUserIDFromContext(ctx)
		tenantID := am.getTenantIDFromContext(ctx)

		if userID == 0 || tenantID == 0 {
			am.respondWithError(r, 401, "用户认证信息不完整")
			return
		}

		// 检查用户是否拥有所有所需权限
		for _, permission := range permissions {
			hasPermission, err := am.roleRepository.HasPermission(ctx, userID, tenantID, permission)
			if err != nil {
				g.Log().Errorf(ctx, "权限检查失败: %v", err)
				am.respondWithError(r, 500, "权限检查失败")
				return
			}

			if !hasPermission {
				am.respondWithError(r, 403, "权限不足：缺少"+string(permission))
				return
			}
		}

		// 继续处理请求
		r.Middleware.Next()
	}
}

// RequireAnyPermission 要求任意权限的中间件（OR关系）
func (am *AuthMiddleware) RequireAnyPermission(permissions ...types.Permission) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := r.GetCtx()
		path := r.URL.Path

		// 检查是否为公开路径
		if am.isPublicPath(path) {
			r.Middleware.Next()
			return
		}

		// 从上下文获取用户信息
		userID := am.getUserIDFromContext(ctx)
		tenantID := am.getTenantIDFromContext(ctx)

		if userID == 0 || tenantID == 0 {
			am.respondWithError(r, 401, "用户认证信息不完整")
			return
		}

		// 检查用户是否拥有任意一个所需权限
		for _, permission := range permissions {
			hasPermission, err := am.roleRepository.HasPermission(ctx, userID, tenantID, permission)
			if err != nil {
				g.Log().Errorf(ctx, "权限检查失败: %v", err)
				continue
			}

			if hasPermission {
				// 拥有其中一个权限即可通过
				r.Middleware.Next()
				return
			}
		}

		// 没有任何所需权限
		am.respondWithError(r, 403, "权限不足")
	}
}

// extractToken 从请求中提取Token
func (am *AuthMiddleware) extractToken(r *ghttp.Request) string {
	// 1. 从Authorization头提取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// 支持 "Bearer <token>" 格式
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		// 支持直接传递token
		return authHeader
	}

	// 2. 从X-Access-Token头提取
	if token := r.Header.Get("X-Access-Token"); token != "" {
		return token
	}

	// 3. 从查询参数提取（不推荐，但支持）
	if token := r.Get("token").String(); token != "" {
		return token
	}

	// 4. 从Cookie提取（如果启用了Cookie认证）
	if cookie := r.Cookie.Get("access_token"); cookie != nil {
		return cookie.String()
	}

	return ""
}

// extractTenantID 从请求中提取租户ID
func (am *AuthMiddleware) extractTenantID(r *ghttp.Request) uint64 {
	// 1. 从X-Tenant-ID头提取
	if tenantIDStr := r.Header.Get("X-Tenant-ID"); tenantIDStr != "" {
		return g.NewVar(tenantIDStr).Uint64()
	}

	// 2. 从查询参数提取
	if tenantID := r.Get("tenant_id").Uint64(); tenantID > 0 {
		return tenantID
	}

	// 3. 从路径参数提取（如果路由中包含）
	if tenantID := r.Get("tenantId").Uint64(); tenantID > 0 {
		return tenantID
	}

	return 0
}

// isPublicPath 检查是否为公开路径
func (am *AuthMiddleware) isPublicPath(path string) bool {
	for _, publicPath := range am.publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// isSkipPath 检查是否为跳过路径
func (am *AuthMiddleware) isSkipPath(path string) bool {
	for _, skipPath := range am.skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getUserIDFromContext 从上下文获取用户ID
func (am *AuthMiddleware) getUserIDFromContext(ctx context.Context) uint64 {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(uint64); ok {
			return id
		}
	}
	return 0
}

// getTenantIDFromContext 从上下文获取租户ID
func (am *AuthMiddleware) getTenantIDFromContext(ctx context.Context) uint64 {
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(uint64); ok {
			return id
		}
	}
	return 0
}

// respondWithError 统一错误响应
func (am *AuthMiddleware) respondWithError(r *ghttp.Request, code int, message string) {
	r.Response.Status = code
	r.Response.WriteJson(g.Map{
		"code": code,
		"msg":  message,
		"data": nil,
	})
	r.ExitAll()
}

// GetCurrentUser 获取当前用户信息的工具函数
func GetCurrentUser(ctx context.Context) *types.UserInfo {
	userID := ctx.Value("user_id")
	username := ctx.Value("username")
	tenantID := ctx.Value("tenant_id")

	if userID == nil || username == nil || tenantID == nil {
		return nil
	}

	return &types.UserInfo{
		ID:       userID.(uint64),
		Username: username.(string),
		TenantID: tenantID.(uint64),
	}
}

// GetCurrentUserPermissions 获取当前用户权限的工具函数
func GetCurrentUserPermissions(ctx context.Context) *types.UserPermissions {
	userID := ctx.Value("user_id")
	tenantID := ctx.Value("tenant_id")
	roles := ctx.Value("roles")
	permissions := ctx.Value("permissions")

	if userID == nil || tenantID == nil {
		return nil
	}

	userPermissions := &types.UserPermissions{
		UserID:   userID.(uint64),
		TenantID: tenantID.(uint64),
	}

	if roles != nil {
		if roleList, ok := roles.([]types.RoleType); ok {
			userPermissions.Roles = roleList
		}
	}

	if permissions != nil {
		if permList, ok := permissions.([]types.Permission); ok {
			userPermissions.Permissions = permList
		}
	}

	return userPermissions
}

// HasPermissionInContext 在上下文中检查权限
func HasPermissionInContext(ctx context.Context, permission types.Permission) bool {
	permissions := ctx.Value("permissions")
	if permissions == nil {
		return false
	}

	if permList, ok := permissions.([]types.Permission); ok {
		for _, perm := range permList {
			if perm == permission {
				return true
			}
		}
	}

	return false
}

// HasRoleInContext 在上下文中检查角色
func HasRoleInContext(ctx context.Context, role types.RoleType) bool {
	roles := ctx.Value("roles")
	if roles == nil {
		return false
	}

	if roleList, ok := roles.([]types.RoleType); ok {
		for _, r := range roleList {
			if r == role {
				return true
			}
		}
	}

	return false
}
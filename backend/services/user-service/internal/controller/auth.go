package controller

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/services/user-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// AuthController 认证控制器
type AuthController struct {
	authService *service.AuthService
	jwtManager  *auth.JWTManager
}

// NewAuthController 创建认证控制器
func NewAuthController() *AuthController {
	return &AuthController{
		authService: service.NewAuthService(),
		jwtManager:  auth.NewJWTManager(),
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" v:"required|length:3,50#用户名不能为空|用户名长度为3-50字符"`
	Password string `json:"password" v:"required|length:6,128#密码不能为空|密码长度为6-128字符"`
	TenantID uint64 `json:"tenant_id" v:"required|min:1#租户ID不能为空|租户ID必须大于0"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    int64           `json:"expires_in"` // 访问令牌过期时间（秒）
	TokenType    string          `json:"token_type"` // Bearer
	User         *types.UserInfo `json:"user"`
}

// LogoutRequest 登出请求结构
type LogoutRequest struct {
	Token        string `json:"token,omitempty"`         // 可选，从Header或Body获取
	RefreshToken string `json:"refresh_token,omitempty"` // 可选，同时撤销刷新令牌
}

// Login 用户登录
func (c *AuthController) Login(r *ghttp.Request) {
	ctx := r.GetCtx()

	var req LoginRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  fmt.Sprintf("请求参数错误: %v", err),
			"data": nil,
		})
		return
	}

	// 验证用户凭证
	user, userPermissions, err := c.authService.ValidateUserCredentials(ctx, req.Username, req.Password, req.TenantID)
	if err != nil {
		g.Log().Errorf(ctx, "登录验证失败 - 用户: %s, 租户: %d, 错误: %v", req.Username, req.TenantID, err)
		r.Response.WriteJsonExit(g.Map{
			"code": 401,
			"msg":  "用户名、密码或租户信息错误",
			"data": nil,
		})
		return
	}

	// 检查用户状态
	if user.Status != types.UserStatusActive {
		g.Log().Warningf(ctx, "用户状态异常 - 用户: %s, 状态: %s", req.Username, user.Status)
		r.Response.WriteJsonExit(g.Map{
			"code": 403,
			"msg":  "用户账户已被禁用或待激活",
			"data": nil,
		})
		return
	}

	// 生成访问令牌
	accessToken, err := c.jwtManager.GenerateTokenWithPermissions(ctx, user, userPermissions, "access")
	if err != nil {
		g.Log().Errorf(ctx, "生成访问令牌失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "登录失败，请稍后重试",
			"data": nil,
		})
		return
	}

	// 生成刷新令牌
	refreshToken, err := c.jwtManager.GenerateTokenWithPermissions(ctx, user, userPermissions, "refresh")
	if err != nil {
		g.Log().Errorf(ctx, "生成刷新令牌失败: %v", err)
		// 清理已生成的访问令牌
		c.jwtManager.RevokeToken(ctx, accessToken)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "登录失败，请稍后重试",
			"data": nil,
		})
		return
	}

	// 更新用户最后登录时间
	err = c.authService.UpdateLastLoginTime(ctx, user.ID)
	if err != nil {
		g.Log().Warningf(ctx, "更新用户最后登录时间失败: %v", err)
		// 不影响登录流程，继续执行
	}

	// 构建用户信息（不包含敏感信息）
	userInfo := &types.UserInfo{
		ID:         user.ID,
		UUID:       user.UUID,
		Username:   user.Username,
		Email:      user.Email,
		Phone:      user.Phone,
		Status:     user.Status,
		TenantID:   user.TenantID,
		MerchantID: user.MerchantID, // 包含商户ID
		Roles:      userPermissions.Roles,
		Profile:    user.Profile,
	}

	// 计算访问令牌过期时间（24小时）
	expiresIn := int64(24 * 60 * 60) // 24小时，单位：秒

	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		User:         userInfo,
	}

	g.Log().Infof(ctx, "用户登录成功 - 用户: %s, 租户: %d, 角色: %v",
		user.Username, user.TenantID, userPermissions.Roles)

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "登录成功",
		"data": response,
	})
}

// Logout 用户登出
func (c *AuthController) Logout(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 从请求中获取令牌
	var req LogoutRequest
	if err := r.Parse(&req); err != nil {
		g.Log().Warningf(ctx, "解析登出请求失败: %v", err)
		// 继续执行，尝试从Header获取令牌
	}

	// 从Authorization Header获取访问令牌
	token := req.Token
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "缺少访问令牌",
			"data": nil,
		})
		return
	}

	// 验证令牌并获取用户信息
	claims, err := c.jwtManager.ValidateToken(ctx, token)
	if err != nil {
		g.Log().Warningf(ctx, "登出时令牌验证失败: %v", err)
		// 即使令牌无效，也尝试撤销以确保安全
	}

	// 撤销访问令牌
	err = c.jwtManager.RevokeToken(ctx, token)
	if err != nil {
		g.Log().Errorf(ctx, "撤销访问令牌失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "登出失败，请稍后重试",
			"data": nil,
		})
		return
	}

	// 如果提供了刷新令牌，也一并撤销
	if req.RefreshToken != "" {
		err = c.jwtManager.RevokeToken(ctx, req.RefreshToken)
		if err != nil {
			g.Log().Warningf(ctx, "撤销刷新令牌失败: %v", err)
			// 不影响登出流程
		}
	}

	// 记录用户登出
	if claims != nil {
		g.Log().Infof(ctx, "用户登出成功 - 用户ID: %d, 租户: %d", claims.UserID, claims.TenantID)
	} else {
		g.Log().Infof(ctx, "用户登出完成 - 令牌已撤销")
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "登出成功",
		"data": nil,
	})
}

// RefreshToken 刷新访问令牌
func (c *AuthController) RefreshToken(r *ghttp.Request) {
	ctx := r.GetCtx()

	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" v:"required#刷新令牌不能为空"`
	}

	var req RefreshRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  fmt.Sprintf("请求参数错误: %v", err),
			"data": nil,
		})
		return
	}

	// 使用安全的令牌轮换刷新
	newAccessToken, newRefreshToken, err := c.jwtManager.RefreshTokenWithRotation(ctx, req.RefreshToken)
	if err != nil {
		g.Log().Errorf(ctx, "令牌刷新失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code": 401,
			"msg":  "令牌刷新失败，请重新登录",
			"data": nil,
		})
		return
	}

	// 计算新的过期时间
	expiresIn := int64(24 * 60 * 60) // 24小时

	response := g.Map{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    expiresIn,
		"token_type":    "Bearer",
	}

	g.Log().Infof(ctx, "令牌刷新成功")

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "令牌刷新成功",
		"data": response,
	})
}

// GetUserInfo 获取当前用户信息（需要认证）
func (c *AuthController) GetUserInfo(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 从上下文获取当前用户信息（由认证中间件注入）
	currentUser := ctx.Value("current_user")
	if currentUser == nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 401,
			"msg":  "未认证用户",
			"data": nil,
		})
		return
	}

	claims, ok := currentUser.(*auth.TokenClaims)
	if !ok {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "用户信息格式错误",
			"data": nil,
		})
		return
	}

	// 获取用户详细信息
	user, err := c.authService.GetUserByID(ctx, claims.UserID, claims.TenantID)
	if err != nil {
		g.Log().Errorf(ctx, "获取用户信息失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "获取用户信息失败",
			"data": nil,
		})
		return
	}

	// 构建响应数据（不包含敏感信息）
	userInfo := &types.UserInfo{
		ID:         user.ID,
		UUID:       user.UUID,
		Username:   user.Username,
		Email:      user.Email,
		Phone:      user.Phone,
		Status:     user.Status,
		TenantID:   user.TenantID,
		MerchantID: user.MerchantID, // 包含商户ID
		Roles:      claims.Roles,
		Profile:    user.Profile,
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "获取用户信息成功",
		"data": userInfo,
	})
}

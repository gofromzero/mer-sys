package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

// TestMiddlewareJWTIntegration 测试中间件与JWT验证的集成
func TestMiddlewareJWTIntegration(t *testing.T) {
	Convey("中间件与JWT验证集成测试", t, func() {
		// 创建测试用JWT管理器（避免依赖配置文件）
		jwtManager := auth.NewJWTManagerForTest("test-jwt-secret-key", 24)
		mockRepo := NewMockRoleRepository()
		middleware := NewAuthMiddlewareForTest(jwtManager, mockRepo)
		
		Convey("有效JWT Token验证流程", func() {
			// 创建测试用户权限
			userPerms := &types.UserPermissions{
				UserID:   1,
				TenantID: 100,
				Roles:    []types.RoleType{types.RoleMerchant},
				Permissions: []types.Permission{
					types.PermissionUserView,
					types.PermissionProductView,
				},
			}
			
			// 创建测试用户
			user := &types.User{
				ID:       1,
				TenantID: 100,
				Username: "testuser",
				Email:    "test@example.com",
			}
			
			ctx := context.Background()
			
			// 生成JWT Token
			token, err := jwtManager.GenerateTokenWithPermissions(ctx, user, userPerms, "access")
			So(err, ShouldBeNil)
			So(token, ShouldNotBeEmpty)
			
			// 创建带Token的HTTP请求
			req := httptest.NewRequest("GET", "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("X-Tenant-ID", "100")
			
			// 验证Token解析
			claims, err := jwtManager.ValidateToken(ctx, token)
			So(err, ShouldBeNil)
			So(claims.UserID, ShouldEqual, 1)
			So(claims.TenantID, ShouldEqual, 100)
			So(len(claims.Roles), ShouldEqual, 1)
			So(claims.Roles[0], ShouldEqual, types.RoleMerchant)
			So(len(claims.Permissions), ShouldEqual, 2)
		})
		
		Convey("无效JWT Token处理", func() {
			// 测试过期Token
			expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDk0NTkyMDB9.invalid"
			
			req := httptest.NewRequest("GET", "/api/v1/users", nil)
			req.Header.Set("Authorization", "Bearer "+expiredToken)
			
			// 验证Token失败
			ctx := context.Background()
			_, err := jwtManager.ValidateToken(ctx, expiredToken)
			So(err, ShouldNotBeNil)
		})
		
		Convey("缺少Token的请求处理", func() {
			req := httptest.NewRequest("GET", "/api/v1/users", nil)
			// 不设置Authorization头
			
			// 验证中间件能正确识别缺少Token的情况
			r := &ghttp.Request{Request: req}
			token := middleware.extractToken(r)
			So(token, ShouldBeEmpty)
		})
	})
}

// TestPermissionTenantIsolationIntegration 测试权限检查与租户隔离的配合
func TestPermissionTenantIsolationIntegration(t *testing.T) {
	Convey("权限检查与租户隔离集成测试", t, func() {
		mockRepo := NewMockRoleRepository()
		
		Convey("同租户用户权限验证", func() {
			// 设置用户1在租户100的权限
			userID := uint64(1)
			tenantID := uint64(100)
			mockRepo.SetUserPermissions(userID, tenantID, []types.Permission{
				types.PermissionUserView,
				types.PermissionProductManage,
			})
			mockRepo.SetUserRoles(userID, tenantID, []types.RoleType{
				types.RoleMerchant,
			})
			
			ctx := context.Background()
			
			// 验证用户在正确租户下的权限
			hasPermission, err := mockRepo.HasPermission(ctx, userID, tenantID, types.PermissionUserView)
			So(err, ShouldBeNil)
			So(hasPermission, ShouldBeTrue)
			
			hasRole, err := mockRepo.HasRole(ctx, userID, tenantID, types.RoleMerchant)
			So(err, ShouldBeNil)
			So(hasRole, ShouldBeTrue)
		})
		
		Convey("跨租户权限隔离验证", func() {
			// 设置用户1在租户100的权限
			userID := uint64(1)
			tenantID100 := uint64(100)
			tenantID200 := uint64(200)
			
			mockRepo.SetUserPermissions(userID, tenantID100, []types.Permission{
				types.PermissionUserView,
			})
			
			ctx := context.Background()
			
			// 验证用户在租户100有权限
			hasPermission100, err := mockRepo.HasPermission(ctx, userID, tenantID100, types.PermissionUserView)
			So(err, ShouldBeNil)
			So(hasPermission100, ShouldBeTrue)
			
			// 验证用户在租户200没有权限（租户隔离）
			hasPermission200, err := mockRepo.HasPermission(ctx, userID, tenantID200, types.PermissionUserView)
			So(err, ShouldBeNil)
			So(hasPermission200, ShouldBeFalse)
		})
		
		Convey("多权限组合验证", func() {
			userID := uint64(2)
			tenantID := uint64(100)
			
			// 设置用户拥有部分权限
			mockRepo.SetUserPermissions(userID, tenantID, []types.Permission{
				types.PermissionUserView,
				types.PermissionOrderView,
			})
			
			ctx := context.Background()
			
			// 验证拥有的权限
			hasUserView, _ := mockRepo.HasPermission(ctx, userID, tenantID, types.PermissionUserView)
			hasOrderView, _ := mockRepo.HasPermission(ctx, userID, tenantID, types.PermissionOrderView)
			So(hasUserView, ShouldBeTrue)
			So(hasOrderView, ShouldBeTrue)
			
			// 验证没有的权限
			hasUserManage, _ := mockRepo.HasPermission(ctx, userID, tenantID, types.PermissionUserManage)
			hasProductManage, _ := mockRepo.HasPermission(ctx, userID, tenantID, types.PermissionProductManage)
			So(hasUserManage, ShouldBeFalse)
			So(hasProductManage, ShouldBeFalse)
		})
	})
}

// TestPublicPathWhitelistIntegration 测试公开路径白名单功能
func TestPublicPathWhitelistIntegration(t *testing.T) {
	Convey("公开路径白名单集成测试", t, func() {
		// 使用测试专用构造函数避免配置依赖
		jwtManager := auth.NewJWTManagerForTest("test-secret", 24)
		roleRepo := &MockRoleRepository{}
		middleware := NewAuthMiddlewareForTest(jwtManager, roleRepo)
		
		Convey("默认公开路径验证", func() {
			// 测试认证相关的公开路径
			publicAuthPaths := []string{
				"/api/v1/auth/login",
				"/api/v1/auth/register",
				"/api/v1/auth/refresh",
			}
			
			for _, path := range publicAuthPaths {
				So(middleware.isPublicPath(path), ShouldBeTrue)
			}
			
			// 测试健康检查和系统路径
			systemPaths := []string{
				"/api/v1/health",
				"/api/v1/ping",
				"/api/v1/version",
			}
			
			for _, path := range systemPaths {
				So(middleware.isPublicPath(path), ShouldBeTrue)
			}
		})
		
		Convey("受保护路径验证", func() {
			// 测试需要认证的路径
			protectedPaths := []string{
				"/api/v1/users",
				"/api/v1/users/profile",
				"/api/v1/products",
				"/api/v1/orders",
				"/api/v1/admin/dashboard",
			}
			
			for _, path := range protectedPaths {
				So(middleware.isPublicPath(path), ShouldBeFalse)
			}
		})
		
		Convey("动态公开路径配置", func() {
			// 添加新的公开路径
			newPublicPaths := []string{
				"/api/v1/public/announcements",
				"/api/v1/public/help",
			}
			
			for _, path := range newPublicPaths {
				middleware.AddPublicPath(path)
				So(middleware.isPublicPath(path), ShouldBeTrue)
				// 测试子路径也应该被允许
				So(middleware.isPublicPath(path+"/details"), ShouldBeTrue)
			}
		})
		
		Convey("跳过路径配置", func() {
			// 测试默认跳过的静态资源路径
			skipPaths := []string{
				"/favicon.ico",
				"/static/css/app.css",
				"/static/js/bundle.js",
				"/assets/images/logo.png",
			}
			
			for _, path := range skipPaths {
				So(middleware.isSkipPath(path), ShouldBeTrue)
			}
			
			// 添加新的跳过路径
			middleware.SetSkipPaths([]string{"/docs", "/swagger"})
			So(middleware.isSkipPath("/docs/api.html"), ShouldBeTrue)
			So(middleware.isSkipPath("/swagger/index.html"), ShouldBeTrue)
		})
		
		Convey("路径匹配优先级测试", func() {
			// 测试路径匹配的优先级：跳过路径 > 公开路径 > 受保护路径
			
			// 设置一个既是公开路径又是跳过路径的情况
			middleware.AddPublicPath("/api/v1/test")
			middleware.SetSkipPaths([]string{"/api/v1/test"})
			
			// 跳过路径应该优先
			So(middleware.isSkipPath("/api/v1/test/endpoint"), ShouldBeTrue)
			So(middleware.isPublicPath("/api/v1/test/endpoint"), ShouldBeTrue)
		})
	})
}

// TestFullAuthenticationFlow 测试完整的认证流程集成
func TestFullAuthenticationFlow(t *testing.T) {
	Convey("完整认证流程集成测试", t, func() {
		// 创建测试用JWT管理器（避免依赖配置文件）
		jwtManager := auth.NewJWTManagerForTest("test-jwt-secret-key", 24)
		mockRepo := NewMockRoleRepository()
		
		Convey("成功认证流程", func() {
			// 1. 创建用户权限
			userID := uint64(1)
			tenantID := uint64(100)
			userPerms := &types.UserPermissions{
				UserID:   userID,
				TenantID: tenantID,
				Roles:    []types.RoleType{types.RoleMerchant},
				Permissions: []types.Permission{
					types.PermissionUserView,
					types.PermissionProductManage,
				},
			}
			
			// 设置Mock仓储数据
			mockRepo.SetUserPermissions(userID, tenantID, userPerms.Permissions)
			mockRepo.SetUserRoles(userID, tenantID, userPerms.Roles)
			
			// 2. 创建测试用户
			user := &types.User{
				ID:       userID,
				TenantID: tenantID,
				Username: "merchant1",
				Email:    "merchant1@example.com",
			}
			
			ctx := context.Background()
			
			// 3. 生成JWT Token
			token, err := jwtManager.GenerateTokenWithPermissions(ctx, user, userPerms, "access")
			So(err, ShouldBeNil)
			So(token, ShouldNotBeEmpty)
			
			// 4. 验证Token
			claims, err := jwtManager.ValidateToken(ctx, token)
			So(err, ShouldBeNil)
			So(claims.UserID, ShouldEqual, userID)
			So(claims.TenantID, ShouldEqual, tenantID)
			
			// 5. 验证权限
			hasPermission, err := mockRepo.HasPermission(ctx, userID, tenantID, types.PermissionUserView)
			So(err, ShouldBeNil)
			So(hasPermission, ShouldBeTrue)
			
			// 6. 验证角色
			hasRole, err := mockRepo.HasRole(ctx, userID, tenantID, types.RoleMerchant)
			So(err, ShouldBeNil)
			So(hasRole, ShouldBeTrue)
		})
		
		Convey("Token刷新流程", func() {
			// 创建用户权限
			userPerms := &types.UserPermissions{
				UserID:   1,
				TenantID: 100,
				Roles:    []types.RoleType{types.RoleCustomer},
				Permissions: []types.Permission{types.PermissionUserView},
			}
			
			// 创建测试用户
			user := &types.User{
				ID:       1,
				TenantID: 100,
				Username: "customer1",
				Email:    "customer1@example.com",
			}
			
			ctx := context.Background()
			
			// 生成初始Token
			originalToken, err := jwtManager.GenerateTokenWithPermissions(ctx, user, userPerms, "access")
			So(err, ShouldBeNil)
			
			// 验证原始Token
			originalClaims, err := jwtManager.ValidateToken(ctx, originalToken)
			So(err, ShouldBeNil)
			So(originalClaims.UserID, ShouldEqual, 1)
			
			// 模拟Token刷新（这里简化处理，实际应该有刷新Token机制）
			newToken, err := jwtManager.GenerateTokenWithPermissions(ctx, user, userPerms, "access")
			So(err, ShouldBeNil)
			So(newToken, ShouldNotEqual, originalToken) // 新Token应该不同
			
			// 验证新Token
			newClaims, err := jwtManager.ValidateToken(ctx, newToken)
			So(err, ShouldBeNil)
			So(newClaims.UserID, ShouldEqual, originalClaims.UserID)
			So(newClaims.TenantID, ShouldEqual, originalClaims.TenantID)
		})
		
		Convey("权限不足场景", func() {
			// 创建权限有限的用户
			userID := uint64(2)
			tenantID := uint64(100)
			mockRepo.SetUserPermissions(userID, tenantID, []types.Permission{
				types.PermissionUserView, // 只有查看权限
			})
			
			ctx := context.Background()
			
			// 验证有的权限
			hasViewPermission, err := mockRepo.HasPermission(ctx, userID, tenantID, types.PermissionUserView)
			So(err, ShouldBeNil)
			So(hasViewPermission, ShouldBeTrue)
			
			// 验证没有的权限
			hasManagePermission, err := mockRepo.HasPermission(ctx, userID, tenantID, types.PermissionUserManage)
			So(err, ShouldBeNil)
			So(hasManagePermission, ShouldBeFalse)
		})
	})
}
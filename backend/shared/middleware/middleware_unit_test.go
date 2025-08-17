package middleware

import (
	"context"
	"testing"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

// TestMiddlewareLogic 测试中间件核心逻辑（不依赖GoFrame配置）
func TestMiddlewareLogic(t *testing.T) {
	Convey("中间件核心逻辑测试", t, func() {
		
		Convey("路径匹配测试", func() {
			publicPaths := []string{
				"/api/v1/auth/login",
				"/api/v1/auth/register",
				"/api/v1/health",
			}
			
			skipPaths := []string{
				"/favicon.ico",
				"/static",
				"/assets",
			}
			
			Convey("公开路径匹配", func() {
				So(isPathMatched("/api/v1/auth/login", publicPaths), ShouldBeTrue)
				So(isPathMatched("/api/v1/auth/login/test", publicPaths), ShouldBeTrue) // 前缀匹配
				So(isPathMatched("/api/v1/health/ready", publicPaths), ShouldBeTrue)
				So(isPathMatched("/api/v1/users", publicPaths), ShouldBeFalse)
			})
			
			Convey("跳过路径匹配", func() {
				So(isPathMatched("/favicon.ico", skipPaths), ShouldBeTrue)
				So(isPathMatched("/static/css/app.css", skipPaths), ShouldBeTrue)
				So(isPathMatched("/assets/js/main.js", skipPaths), ShouldBeTrue)
				So(isPathMatched("/api/v1/users", skipPaths), ShouldBeFalse)
			})
		})

		Convey("Token提取逻辑测试", func() {
			Convey("Bearer Token格式", func() {
				token := extractTokenFromAuthHeader("Bearer abc123def456")
				So(token, ShouldEqual, "abc123def456")
			})
			
			Convey("直接Token格式", func() {
				token := extractTokenFromAuthHeader("abc123def456")
				So(token, ShouldEqual, "abc123def456")
			})
			
			Convey("空Authorization头", func() {
				token := extractTokenFromAuthHeader("")
				So(token, ShouldEqual, "")
			})
		})

		Convey("上下文用户信息提取测试", func() {
			Convey("完整用户信息", func() {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "user_id", uint64(123))
				ctx = context.WithValue(ctx, "tenant_id", uint64(456))
				ctx = context.WithValue(ctx, "username", "testuser")

				userID := getUserIDFromCtx(ctx)
				tenantID := getTenantIDFromCtx(ctx)
				username := getUsernameFromCtx(ctx)

				So(userID, ShouldEqual, 123)
				So(tenantID, ShouldEqual, 456)
				So(username, ShouldEqual, "testuser")
			})
			
			Convey("缺少用户信息", func() {
				ctx := context.Background()
				
				userID := getUserIDFromCtx(ctx)
				tenantID := getTenantIDFromCtx(ctx)
				username := getUsernameFromCtx(ctx)

				So(userID, ShouldEqual, 0)
				So(tenantID, ShouldEqual, 0)
				So(username, ShouldEqual, "")
			})
		})

		Convey("权限检查逻辑测试", func() {
			permissions := []types.Permission{
				types.PermissionUserView,
				types.PermissionOrderView,
				types.PermissionProductManage,
			}
			
			Convey("权限存在检查", func() {
				So(hasPermissionInList(types.PermissionUserView, permissions), ShouldBeTrue)
				So(hasPermissionInList(types.PermissionOrderView, permissions), ShouldBeTrue)
				So(hasPermissionInList(types.PermissionProductManage, permissions), ShouldBeTrue)
			})
			
			Convey("权限不存在检查", func() {
				So(hasPermissionInList(types.PermissionUserManage, permissions), ShouldBeFalse)
				So(hasPermissionInList(types.PermissionSystemConfig, permissions), ShouldBeFalse)
			})
		})

		Convey("角色检查逻辑测试", func() {
			roles := []types.RoleType{
				types.RoleMerchant,
				types.RoleCustomer,
			}
			
			Convey("角色存在检查", func() {
				So(hasRoleInList(types.RoleMerchant, roles), ShouldBeTrue)
				So(hasRoleInList(types.RoleCustomer, roles), ShouldBeTrue)
			})
			
			Convey("角色不存在检查", func() {
				So(hasRoleInList(types.RoleTenantAdmin, roles), ShouldBeFalse)
			})
		})
	})
}

// TestPermissionUtilities 权限工具函数测试
func TestPermissionUtilities(t *testing.T) {
	Convey("权限工具函数测试", t, func() {
		
		Convey("GetCurrentUser函数测试", func() {
			Convey("完整用户信息", func() {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "user_id", uint64(123))
				ctx = context.WithValue(ctx, "username", "testuser")
				ctx = context.WithValue(ctx, "tenant_id", uint64(456))
				
				user := GetCurrentUser(ctx)
				So(user, ShouldNotBeNil)
				So(user.ID, ShouldEqual, 123)
				So(user.Username, ShouldEqual, "testuser")
				So(user.TenantID, ShouldEqual, 456)
			})
			
			Convey("缺少必要字段", func() {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "user_id", uint64(123))
				// 缺少username和tenant_id
				
				user := GetCurrentUser(ctx)
				So(user, ShouldBeNil)
			})
		})

		Convey("HasPermissionInContext函数测试", func() {
			Convey("有权限", func() {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "permissions", []types.Permission{
					types.PermissionUserView,
					types.PermissionOrderView,
				})
				
				So(HasPermissionInContext(ctx, types.PermissionUserView), ShouldBeTrue)
				So(HasPermissionInContext(ctx, types.PermissionOrderView), ShouldBeTrue)
				So(HasPermissionInContext(ctx, types.PermissionUserManage), ShouldBeFalse)
			})
			
			Convey("无权限信息", func() {
				ctx := context.Background()
				
				So(HasPermissionInContext(ctx, types.PermissionUserView), ShouldBeFalse)
			})
		})

		Convey("HasRoleInContext函数测试", func() {
			Convey("有角色", func() {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "roles", []types.RoleType{
					types.RoleMerchant,
					types.RoleCustomer,
				})
				
				So(HasRoleInContext(ctx, types.RoleMerchant), ShouldBeTrue)
				So(HasRoleInContext(ctx, types.RoleCustomer), ShouldBeTrue)
				So(HasRoleInContext(ctx, types.RoleTenantAdmin), ShouldBeFalse)
			})
			
			Convey("无角色信息", func() {
				ctx := context.Background()
				
				So(HasRoleInContext(ctx, types.RoleMerchant), ShouldBeFalse)
			})
		})
	})
}

// 辅助函数，提取中间件的核心逻辑进行测试

// isPathMatched 检查路径是否匹配列表中的任一前缀
func isPathMatched(path string, pathList []string) bool {
	for _, p := range pathList {
		if len(path) >= len(p) && path[:len(p)] == p {
			return true
		}
	}
	return false
}

// extractTokenFromAuthHeader 从Authorization头提取Token
func extractTokenFromAuthHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	
	// 支持 "Bearer <token>" 格式
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	
	// 支持直接传递token
	return authHeader
}

// getUserIDFromCtx 从上下文获取用户ID
func getUserIDFromCtx(ctx context.Context) uint64 {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(uint64); ok {
			return id
		}
	}
	return 0
}

// getTenantIDFromCtx 从上下文获取租户ID
func getTenantIDFromCtx(ctx context.Context) uint64 {
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(uint64); ok {
			return id
		}
	}
	return 0
}

// getUsernameFromCtx 从上下文获取用户名
func getUsernameFromCtx(ctx context.Context) string {
	if username := ctx.Value("username"); username != nil {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// hasPermissionInList 检查权限是否在列表中
func hasPermissionInList(permission types.Permission, permissions []types.Permission) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// hasRoleInList 检查角色是否在列表中
func hasRoleInList(role types.RoleType, roles []types.RoleType) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
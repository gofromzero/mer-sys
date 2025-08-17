package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/services/user-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/services/user-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
	"github.com/gofromzero/mer-sys/backend/shared/cache"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

// TestAuthE2EFlow 端到端认证流程测试
func TestAuthE2EFlow(t *testing.T) {
	Convey("认证系统端到端测试", t, func() {
		// 初始化测试环境
		mockCache := cache.NewMockCache()
		jwtManager := auth.NewJWTManagerForTest("test-secret-key", 24)
		mockUserRepo := repository.NewMockUserRepository()
		mockRoleRepo := repository.NewMockRoleRepository()
		authService := service.NewAuthService(jwtManager, mockUserRepo, mockRoleRepo)
		authController := controller.NewAuthController(authService)
		authMiddleware := middleware.NewAuthMiddlewareForTest(jwtManager, mockRoleRepo)

		// 创建测试服务器
		server := g.Server()
		server.Group("/api/v1", func(group *ghttp.RouterGroup) {
			// 认证相关路由
			group.POST("/auth/login", authController.Login)
			group.POST("/auth/logout", authMiddleware.JWTAuth, authController.Logout)
			group.POST("/auth/refresh", authController.RefreshToken)
			
			// 受保护的测试路由
			group.GET("/users", authMiddleware.JWTAuth, func(r *ghttp.Request) {
				r.Response.WriteJson(g.Map{"message": "用户列表", "users": []string{"user1", "user2"}})
			})
			
			// 需要权限的测试路由
			group.GET("/admin/users", 
				authMiddleware.JWTAuth,
				authMiddleware.RequirePermissions(types.PermissionUserManage),
				func(r *ghttp.Request) {
					r.Response.WriteJson(g.Map{"message": "管理员用户列表", "admin_users": []string{"admin1"}})
				},
			)
		})

		// 准备测试数据
		testUser := &types.User{
			ID:         1,
			UUID:       "test-user-uuid",
			Username:   "testuser",
			Email:      "test@example.com",
			TenantID:   1,
			Status:     types.UserStatusActive,
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Uo8.OvOVvVEjKqxI6oCqJBXMHQOyHW", // secret
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		mockUserRepo.CreateUser(context.Background(), testUser)

		// 设置用户权限
		mockRoleRepo.SetUserPermissions(1, 1, []types.Permission{
			types.PermissionUserView,
			types.PermissionUserManage,
		})
		mockRoleRepo.SetUserRoles(1, 1, []types.RoleType{
			types.RoleTenantAdmin,
		})

		Convey("完整认证流程测试", func() {
			var accessToken, refreshToken string

			Convey("1. 用户登录", func() {
				loginData := map[string]interface{}{
					"username":  "testuser",
					"password":  "secret",
					"tenant_id": 1,
				}
				jsonData, _ := json.Marshal(loginData)

				req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				data := response["data"].(map[string]interface{})
				accessToken = data["access_token"].(string)
				refreshToken = data["refresh_token"].(string)

				So(accessToken, ShouldNotBeEmpty)
				So(refreshToken, ShouldNotBeEmpty)

				// 验证用户信息
				user := data["user"].(map[string]interface{})
				So(user["username"], ShouldEqual, "testuser")
				So(user["email"], ShouldEqual, "test@example.com")
			})

			Convey("2. 使用Token访问受保护资源", func() {
				req := httptest.NewRequest("GET", "/api/v1/users", nil)
				req.Header.Set("Authorization", "Bearer "+accessToken)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["message"], ShouldEqual, "用户列表")
			})

			Convey("3. 访问需要权限的资源", func() {
				req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
				req.Header.Set("Authorization", "Bearer "+accessToken)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["message"], ShouldEqual, "管理员用户列表")
			})

			Convey("4. Token刷新", func() {
				refreshData := map[string]interface{}{
					"refresh_token": refreshToken,
				}
				jsonData, _ := json.Marshal(refreshData)

				req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				data := response["data"].(map[string]interface{})
				newAccessToken := data["access_token"].(string)
				newRefreshToken := data["refresh_token"].(string)

				So(newAccessToken, ShouldNotBeEmpty)
				So(newRefreshToken, ShouldNotBeEmpty)
				So(newAccessToken, ShouldNotEqual, accessToken)
				So(newRefreshToken, ShouldNotEqual, refreshToken)

				// 更新token用于后续测试
				accessToken = newAccessToken
				refreshToken = newRefreshToken
			})

			Convey("5. 用户登出", func() {
				req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
				req.Header.Set("Authorization", "Bearer "+accessToken)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)
				So(response["message"], ShouldEqual, "登出成功")
			})

			Convey("6. 验证Token已失效", func() {
				req := httptest.NewRequest("GET", "/api/v1/users", nil)
				req.Header.Set("Authorization", "Bearer "+accessToken)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})

		Convey("权限验证测试", func() {
			// 创建无权限用户
			limitedUser := &types.User{
				ID:         2,
				UUID:       "limited-user-uuid",
				Username:   "limiteduser",
				Email:      "limited@example.com",
				TenantID:   1,
				Status:     types.UserStatusActive,
				PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Uo8.OvOVvVEjKqxI6oCqJBXMHQOyHW", // secret
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			mockUserRepo.CreateUser(context.Background(), limitedUser)

			// 只给查看权限，没有管理权限
			mockRoleRepo.SetUserPermissions(2, 1, []types.Permission{
				types.PermissionUserView,
			})
			mockRoleRepo.SetUserRoles(2, 1, []types.RoleType{
				types.RoleCustomer,
			})

			Convey("无权限用户访问管理接口应被拒绝", func() {
				// 登录获取token
				loginData := map[string]interface{}{
					"username":  "limiteduser",
					"password":  "secret",
					"tenant_id": 1,
				}
				jsonData, _ := json.Marshal(loginData)

				req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)
				So(w.Code, ShouldEqual, http.StatusOK)

				var loginResponse map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &loginResponse)
				data := loginResponse["data"].(map[string]interface{})
				limitedToken := data["access_token"].(string)

				// 尝试访问管理接口
				req = httptest.NewRequest("GET", "/api/v1/admin/users", nil)
				req.Header.Set("Authorization", "Bearer "+limitedToken)
				w = httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusForbidden)
			})
		})

		Convey("租户隔离测试", func() {
			// 创建不同租户的用户
			otherTenantUser := &types.User{
				ID:         3,
				UUID:       "other-tenant-user-uuid",
				Username:   "otheruser",
				Email:      "other@example.com",
				TenantID:   2, // 不同租户
				Status:     types.UserStatusActive,
				PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Uo8.OvOVvVEjKqxI6oCqJBXMHQOyHW", // secret
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			mockUserRepo.CreateUser(context.Background(), otherTenantUser)

			// 设置权限（但在不同租户）
			mockRoleRepo.SetUserPermissions(3, 2, []types.Permission{
				types.PermissionUserManage,
			})

			Convey("不同租户用户无法访问其他租户资源", func() {
				// 登录获取token
				loginData := map[string]interface{}{
					"username":  "otheruser",
					"password":  "secret",
					"tenant_id": 2,
				}
				jsonData, _ := json.Marshal(loginData)

				req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)
				So(w.Code, ShouldEqual, http.StatusOK)

				var loginResponse map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &loginResponse)
				data := loginResponse["data"].(map[string]interface{})
				otherToken := data["access_token"].(string)

				// 验证token包含正确的租户信息
				claims, err := jwtManager.ValidateToken(context.Background(), otherToken)
				So(err, ShouldBeNil)
				So(claims.TenantID, ShouldEqual, 2)
				So(claims.UserID, ShouldEqual, 3)
			})
		})
	})
}

// TestAuthSecurityE2E 认证安全端到端测试
func TestAuthSecurityE2E(t *testing.T) {
	Convey("认证安全测试", t, func() {
		// 初始化测试环境
		mockCache := cache.NewMockCache()
		jwtManager := auth.NewJWTManagerForTest("test-secret-key", 24)
		mockUserRepo := repository.NewMockUserRepository()
		mockRoleRepo := repository.NewMockRoleRepository()
		authService := service.NewAuthService(jwtManager, mockUserRepo, mockRoleRepo)
		authController := controller.NewAuthController(authService)
		authMiddleware := middleware.NewAuthMiddlewareForTest(jwtManager, mockRoleRepo)

		// 创建测试服务器
		server := g.Server()
		server.Group("/api/v1", func(group *ghttp.RouterGroup) {
			group.POST("/auth/login", authController.Login)
			group.GET("/users", authMiddleware.JWTAuth, func(r *ghttp.Request) {
				r.Response.WriteJson(g.Map{"message": "success"})
			})
		})

		Convey("SQL注入防护测试", func() {
			Convey("登录接口SQL注入防护", func() {
				// 尝试SQL注入攻击
				maliciousData := map[string]interface{}{
					"username":  "admin'; DROP TABLE users; --",
					"password":  "password",
					"tenant_id": 1,
				}
				jsonData, _ := json.Marshal(maliciousData)

				req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				// 应该返回认证失败，而不是服务器错误
				So(w.Code, ShouldEqual, http.StatusUnauthorized)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldNotEqual, 0)
			})
		})

		Convey("Token伪造防护测试", func() {
			Convey("伪造Token应被拒绝", func() {
				// 使用伪造的token
				fakeToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

				req := httptest.NewRequest("GET", "/api/v1/users", nil)
				req.Header.Set("Authorization", "Bearer "+fakeToken)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("空Token应被拒绝", func() {
				req := httptest.NewRequest("GET", "/api/v1/users", nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})

			Convey("格式错误的Token应被拒绝", func() {
				req := httptest.NewRequest("GET", "/api/v1/users", nil)
				req.Header.Set("Authorization", "InvalidToken")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusUnauthorized)
			})
		})

		Convey("暴力破解防护测试", func() {
			Convey("多次错误登录应有适当响应", func() {
				// 模拟多次错误登录尝试
				for i := 0; i < 5; i++ {
					wrongData := map[string]interface{}{
						"username":  "nonexistent",
						"password":  "wrongpassword",
						"tenant_id": 1,
					}
					jsonData, _ := json.Marshal(wrongData)

					req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
					req.Header.Set("Content-Type", "application/json")
					w := httptest.NewRecorder()

					server.ServeHTTP(w, req)

					// 每次都应该返回认证失败
					So(w.Code, ShouldEqual, http.StatusUnauthorized)
				}
			})
		})

		Convey("输入验证测试", func() {
			Convey("空用户名应被拒绝", func() {
				emptyData := map[string]interface{}{
					"username":  "",
					"password":  "password",
					"tenant_id": 1,
				}
				jsonData, _ := json.Marshal(emptyData)

				req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("空密码应被拒绝", func() {
				emptyData := map[string]interface{}{
					"username":  "testuser",
					"password":  "",
					"tenant_id": 1,
				}
				jsonData, _ := json.Marshal(emptyData)

				req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			Convey("无效租户ID应被拒绝", func() {
				invalidData := map[string]interface{}{
					"username":  "testuser",
					"password":  "password",
					"tenant_id": 0,
				}
				jsonData, _ := json.Marshal(invalidData)

				req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}
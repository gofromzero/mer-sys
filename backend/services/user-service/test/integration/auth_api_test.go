package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/services/user-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/services/user-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
	"github.com/gofromzero/mer-sys/backend/shared/config"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

// TestAPIResponse API响应结构
type TestAPIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// TestLoginData 登录响应数据结构
type TestLoginData struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    int64           `json:"expires_in"`
	TokenType    string          `json:"token_type"`
	User         *types.UserInfo `json:"user"`
}

// setupTestServer 设置测试服务器
func setupTestServer() *ghttp.Server {
	// 初始化数据库和Redis连接
	config.InitDatabase()
	config.InitRedis()

	s := g.Server("test")
	authController := controller.NewAuthController()

	// 配置路由
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Group("/auth", func(authGroup *ghttp.RouterGroup) {
			authGroup.POST("/login", authController.Login)
			authGroup.POST("/logout", authController.Logout)
			authGroup.POST("/refresh", authController.RefreshToken)
		})

		group.Group("/user", func(userGroup *ghttp.RouterGroup) {
			userGroup.GET("/info", authController.GetUserInfo)
		})
	})

	return s
}

// setupTestUser 创建测试用户
func setupTestUser(ctx context.Context, tenantID uint64) (*types.User, string, error) {
	userRepo := repository.NewUserRepository()
	
	// 创建测试用户
	testUser := &types.User{
		UUID:     fmt.Sprintf("test-user-%d-%d", tenantID, time.Now().UnixNano()),
		Username: fmt.Sprintf("testuser%d", tenantID),
		Email:    fmt.Sprintf("test%d@example.com", tenantID),
		Phone:    "13800138000",
		Status:   types.UserStatusActive,
		TenantID: tenantID,
	}

	password := "testpassword123"
	
	// 使用AuthService创建用户（自动处理密码加密）
	authService := service.NewAuthService()
	if err := authService.CreateUser(ctx, testUser, password); err != nil {
		return nil, "", fmt.Errorf("创建测试用户失败: %v", err)
	}

	// 查询创建的用户
	createdUser, err := userRepo.FindByUsernameAndTenant(ctx, testUser.Username, tenantID)
	if err != nil {
		return nil, "", fmt.Errorf("查询创建的用户失败: %v", err)
	}

	return createdUser, password, nil
}

// cleanupTestUser 清理测试用户
func cleanupTestUser(ctx context.Context, userID uint64) {
	userRepo := repository.NewUserRepository()
	userRepo.DeleteByID(ctx, userID)
}

// makeHTTPRequest 发送HTTP请求并返回响应
func makeHTTPRequest(server *ghttp.Server, method, path string, body interface{}, headers map[string]string) (*http.Response, *TestAPIResponse, error) {
	var requestBody []byte
	var err error

	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("序列化请求体失败: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(requestBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置自定义头部
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	resp := w.Result()

	// 解析响应体
	var apiResp TestAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return resp, nil, fmt.Errorf("解析响应体失败: %v", err)
	}

	return resp, &apiResp, nil
}

// TestAuthAPIIntegration 认证API集成测试
func TestAuthAPIIntegration(t *testing.T) {
	Convey("认证API集成测试", t, func() {
		server := setupTestServer()
		ctx := context.Background()
		tenantID := uint64(1)

		// 创建测试用户
		testUser, testPassword, err := setupTestUser(ctx, tenantID)
		So(err, ShouldBeNil)
		So(testUser, ShouldNotBeNil)

		// 测试完成后清理
		defer cleanupTestUser(ctx, testUser.ID)

		Convey("用户登录API测试", func() {
			Convey("成功登录", func() {
				loginReq := map[string]interface{}{
					"username":  testUser.Username,
					"password":  testPassword,
					"tenant_id": tenantID,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 200)
				So(apiResp.Msg, ShouldEqual, "登录成功")

				// 验证响应数据
				dataBytes, _ := json.Marshal(apiResp.Data)
				var loginData TestLoginData
				json.Unmarshal(dataBytes, &loginData)

				So(loginData.AccessToken, ShouldNotBeEmpty)
				So(loginData.RefreshToken, ShouldNotBeEmpty)
				So(loginData.TokenType, ShouldEqual, "Bearer")
				So(loginData.ExpiresIn, ShouldEqual, 86400) // 24小时
				So(loginData.User, ShouldNotBeNil)
				So(loginData.User.Username, ShouldEqual, testUser.Username)
				So(loginData.User.Email, ShouldEqual, testUser.Email)
				So(loginData.User.TenantID, ShouldEqual, tenantID)

				Convey("令牌验证", func() {
					// 验证访问令牌
					jwtManager := auth.NewJWTManager()
					claims, err := jwtManager.ValidateToken(ctx, loginData.AccessToken)
					So(err, ShouldBeNil)
					So(claims.UserID, ShouldEqual, testUser.ID)
					So(claims.TenantID, ShouldEqual, tenantID)
					So(claims.TokenType, ShouldEqual, "access")

					// 验证刷新令牌
					refreshClaims, err := jwtManager.ValidateToken(ctx, loginData.RefreshToken)
					So(err, ShouldBeNil)
					So(refreshClaims.UserID, ShouldEqual, testUser.ID)
					So(refreshClaims.TenantID, ShouldEqual, tenantID)
					So(refreshClaims.TokenType, ShouldEqual, "refresh")
				})

				Convey("用户登出API测试", func() {
					Convey("使用Authorization Header登出", func() {
						headers := map[string]string{
							"Authorization": fmt.Sprintf("Bearer %s", loginData.AccessToken),
						}

						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/logout", nil, headers)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 200)
						So(apiResp.Msg, ShouldEqual, "登出成功")

						Convey("登出后令牌应被撤销", func() {
							jwtManager := auth.NewJWTManager()
							_, err := jwtManager.ValidateToken(ctx, loginData.AccessToken)
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldContainSubstring, "token has been revoked")
						})
					})

					Convey("使用请求体中的令牌登出", func() {
						// 重新登录获取新令牌
						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
						So(err, ShouldBeNil)
						So(apiResp.Code, ShouldEqual, 200)

						dataBytes, _ := json.Marshal(apiResp.Data)
						var newLoginData TestLoginData
						json.Unmarshal(dataBytes, &newLoginData)

						logoutReq := map[string]interface{}{
							"token":         newLoginData.AccessToken,
							"refresh_token": newLoginData.RefreshToken,
						}

						resp, apiResp, err = makeHTTPRequest(server, "POST", "/api/v1/auth/logout", logoutReq, nil)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 200)
						So(apiResp.Msg, ShouldEqual, "登出成功")

						Convey("访问令牌和刷新令牌都应被撤销", func() {
							jwtManager := auth.NewJWTManager()
							
							_, err := jwtManager.ValidateToken(ctx, newLoginData.AccessToken)
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldContainSubstring, "token has been revoked")

							_, err = jwtManager.ValidateToken(ctx, newLoginData.RefreshToken)
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldContainSubstring, "token has been revoked")
						})
					})
				})

				Convey("令牌刷新API测试", func() {
					// 重新登录获取新令牌
					_, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
					So(err, ShouldBeNil)

					dataBytes, _ := json.Marshal(apiResp.Data)
					var newLoginData TestLoginData
					json.Unmarshal(dataBytes, &newLoginData)

					Convey("成功刷新令牌", func() {
						refreshReq := map[string]interface{}{
							"refresh_token": newLoginData.RefreshToken,
						}

						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/refresh", refreshReq, nil)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 200)
						So(apiResp.Msg, ShouldEqual, "令牌刷新成功")

						// 验证新令牌
						dataBytes, _ := json.Marshal(apiResp.Data)
						var refreshedData map[string]interface{}
						json.Unmarshal(dataBytes, &refreshedData)

						newAccessToken := refreshedData["access_token"].(string)
						newRefreshToken := refreshedData["refresh_token"].(string)

						So(newAccessToken, ShouldNotBeEmpty)
						So(newRefreshToken, ShouldNotBeEmpty)
						So(newAccessToken, ShouldNotEqual, newLoginData.AccessToken)
						So(newRefreshToken, ShouldNotEqual, newLoginData.RefreshToken)

						Convey("新令牌应该有效", func() {
							jwtManager := auth.NewJWTManager()
							claims, err := jwtManager.ValidateToken(ctx, newAccessToken)
							So(err, ShouldBeNil)
							So(claims.UserID, ShouldEqual, testUser.ID)
							So(claims.TenantID, ShouldEqual, tenantID)
						})

						Convey("旧令牌应该被撤销", func() {
							jwtManager := auth.NewJWTManager()
							_, err := jwtManager.ValidateToken(ctx, newLoginData.RefreshToken)
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldContainSubstring, "token has been revoked")
						})
					})

					Convey("使用无效刷新令牌", func() {
						refreshReq := map[string]interface{}{
							"refresh_token": "invalid.refresh.token",
						}

						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/refresh", refreshReq, nil)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 401)
						So(apiResp.Msg, ShouldEqual, "令牌刷新失败，请重新登录")
					})
				})
			})

			Convey("登录失败场景", func() {
				Convey("用户名错误", func() {
					loginReq := map[string]interface{}{
						"username":  "wronguser",
						"password":  testPassword,
						"tenant_id": tenantID,
					}

					resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, 200)
					So(apiResp.Code, ShouldEqual, 401)
					So(apiResp.Msg, ShouldEqual, "用户名、密码或租户信息错误")
				})

				Convey("密码错误", func() {
					loginReq := map[string]interface{}{
						"username":  testUser.Username,
						"password":  "wrongpassword",
						"tenant_id": tenantID,
					}

					resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, 200)
					So(apiResp.Code, ShouldEqual, 401)
					So(apiResp.Msg, ShouldEqual, "用户名、密码或租户信息错误")
				})

				Convey("租户ID错误", func() {
					loginReq := map[string]interface{}{
						"username":  testUser.Username,
						"password":  testPassword,
						"tenant_id": 999,
					}

					resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, 200)
					So(apiResp.Code, ShouldEqual, 401)
					So(apiResp.Msg, ShouldEqual, "用户名、密码或租户信息错误")
				})

				Convey("请求参数验证", func() {
					Convey("缺少用户名", func() {
						loginReq := map[string]interface{}{
							"password":  testPassword,
							"tenant_id": tenantID,
						}

						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 400)
						So(apiResp.Msg, ShouldContainSubstring, "用户名不能为空")
					})

					Convey("缺少密码", func() {
						loginReq := map[string]interface{}{
							"username":  testUser.Username,
							"tenant_id": tenantID,
						}

						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 400)
						So(apiResp.Msg, ShouldContainSubstring, "密码不能为空")
					})

					Convey("缺少租户ID", func() {
						loginReq := map[string]interface{}{
							"username": testUser.Username,
							"password": testPassword,
						}

						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 400)
						So(apiResp.Msg, ShouldContainSubstring, "租户ID不能为空")
					})

					Convey("用户名过短", func() {
						loginReq := map[string]interface{}{
							"username":  "ab",
							"password":  testPassword,
							"tenant_id": tenantID,
						}

						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 400)
						So(apiResp.Msg, ShouldContainSubstring, "用户名长度为3-50字符")
					})

					Convey("密码过短", func() {
						loginReq := map[string]interface{}{
							"username":  testUser.Username,
							"password":  "12345",
							"tenant_id": tenantID,
						}

						resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
						So(err, ShouldBeNil)
						So(resp.StatusCode, ShouldEqual, 200)
						So(apiResp.Code, ShouldEqual, 400)
						So(apiResp.Msg, ShouldContainSubstring, "密码长度为6-128字符")
					})
				})
			})

			Convey("用户状态验证", func() {
				// 创建待激活用户
				pendingUser := &types.User{
					UUID:     fmt.Sprintf("pending-user-%d", time.Now().UnixNano()),
					Username: "pendinguser",
					Email:    "pending@example.com",
					Status:   types.UserStatusPending,
					TenantID: tenantID,
				}

				password := "testpassword123"
				authService := service.NewAuthService()
				err := authService.CreateUser(ctx, pendingUser, password)
				So(err, ShouldBeNil)

				defer func() {
					userRepo := repository.NewUserRepository()
					createdUser, _ := userRepo.FindByUsernameAndTenant(ctx, pendingUser.Username, tenantID)
					if createdUser != nil {
						userRepo.DeleteByID(ctx, createdUser.ID)
					}
				}()

				Convey("待激活用户登录失败", func() {
					loginReq := map[string]interface{}{
						"username":  pendingUser.Username,
						"password":  password,
						"tenant_id": tenantID,
					}

					resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
					So(err, ShouldBeNil)
					So(resp.StatusCode, ShouldEqual, 200)
					So(apiResp.Code, ShouldEqual, 403)
					So(apiResp.Msg, ShouldEqual, "用户账户已被禁用或待激活")
				})
			})
		})

		Convey("登出API错误场景", func() {
			Convey("缺少令牌", func() {
				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/logout", nil, nil)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldEqual, "缺少访问令牌")
			})

			Convey("无效令牌格式", func() {
				headers := map[string]string{
					"Authorization": "Bearer invalid.token.format",
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/logout", nil, headers)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 200) // 即使令牌无效也应该返回成功
				So(apiResp.Msg, ShouldEqual, "登出成功")
			})
		})
	})
}

// TestAuthAPIPasswordVerificationIntegration 密码验证集成测试
func TestAuthAPIPasswordVerificationIntegration(t *testing.T) {
	Convey("密码验证和加密集成测试", t, func() {
		server := setupTestServer()
		ctx := context.Background()
		tenantID := uint64(2)

		Convey("密码加密一致性验证", func() {
			// 创建用户时的密码加密
			testUser := &types.User{
				UUID:     fmt.Sprintf("pwd-test-user-%d", time.Now().UnixNano()),
				Username: "passwordtestuser",
				Email:    "pwdtest@example.com",
				Status:   types.UserStatusActive,
				TenantID: tenantID,
			}

			password := "complex!Password123"
			authService := service.NewAuthService()
			err := authService.CreateUser(ctx, testUser, password)
			So(err, ShouldBeNil)

			defer func() {
				userRepo := repository.NewUserRepository()
				createdUser, _ := userRepo.FindByUsernameAndTenant(ctx, testUser.Username, tenantID)
				if createdUser != nil {
					userRepo.DeleteByID(ctx, createdUser.ID)
				}
			}()

			Convey("登录验证加密一致性", func() {
				loginReq := map[string]interface{}{
					"username":  testUser.Username,
					"password":  password,
					"tenant_id": tenantID,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 200)
				So(apiResp.Msg, ShouldEqual, "登录成功")
			})

			Convey("不同密码登录失败", func() {
				loginReq := map[string]interface{}{
					"username":  testUser.Username,
					"password":  "complex!Password124", // 最后一位不同
					"tenant_id": tenantID,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 401)
				So(apiResp.Msg, ShouldEqual, "用户名、密码或租户信息错误")
			})
		})
	})
}

// TestAuthAPIBoundaryConditions 边界条件测试
func TestAuthAPIBoundaryConditions(t *testing.T) {
	Convey("认证API边界条件测试", t, func() {
		server := setupTestServer()

		Convey("极长用户名和密码", func() {
			// 生成最大长度的用户名和密码
			longUsername := make([]byte, 50)
			for i := range longUsername {
				longUsername[i] = 'a'
			}

			longPassword := make([]byte, 128)
			for i := range longPassword {
				longPassword[i] = 'b'
			}

			loginReq := map[string]interface{}{
				"username":  string(longUsername),
				"password":  string(longPassword),
				"tenant_id": 1,
			}

			resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			// 应该返回401因为用户不存在，但不应该因为长度问题返回400
			So(apiResp.Code, ShouldEqual, 401)
		})

		Convey("超长用户名应被拒绝", func() {
			tooLongUsername := make([]byte, 51) // 超过50字符限制
			for i := range tooLongUsername {
				tooLongUsername[i] = 'a'
			}

			loginReq := map[string]interface{}{
				"username":  string(tooLongUsername),
				"password":  "validpassword",
				"tenant_id": 1,
			}

			resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			So(apiResp.Code, ShouldEqual, 400)
			So(apiResp.Msg, ShouldContainSubstring, "用户名长度为3-50字符")
		})

		Convey("超长密码应被拒绝", func() {
			tooLongPassword := make([]byte, 129) // 超过128字符限制
			for i := range tooLongPassword {
				tooLongPassword[i] = 'b'
			}

			loginReq := map[string]interface{}{
				"username":  "validuser",
				"password":  string(tooLongPassword),
				"tenant_id": 1,
			}

			resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, 200)
			So(apiResp.Code, ShouldEqual, 400)
			So(apiResp.Msg, ShouldContainSubstring, "密码长度为6-128字符")
		})

		Convey("租户ID边界值测试", func() {
			Convey("租户ID为0应被拒绝", func() {
				loginReq := map[string]interface{}{
					"username":  "validuser",
					"password":  "validpassword",
					"tenant_id": 0,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "租户ID必须大于0")
			})

			Convey("租户ID为负数应被拒绝", func() {
				loginReq := map[string]interface{}{
					"username":  "validuser",
					"password":  "validpassword",
					"tenant_id": -1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "租户ID必须大于0")
			})

			Convey("极大租户ID应被处理", func() {
				loginReq := map[string]interface{}{
					"username":  "validuser",
					"password":  "validpassword",
					"tenant_id": uint64(18446744073709551615), // uint64最大值
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				// 应该返回401因为用户不存在，但不应该因为租户ID过大而崩溃
				So(apiResp.Code, ShouldEqual, 401)
			})
		})
	})
}

// TestConcurrentAuthRequests 并发认证请求测试
func TestConcurrentAuthRequests(t *testing.T) {
	Convey("并发认证请求测试", t, func() {
		server := setupTestServer()
		ctx := context.Background()
		tenantID := uint64(3)

		// 创建测试用户
		testUser, testPassword, err := setupTestUser(ctx, tenantID)
		So(err, ShouldBeNil)
		defer cleanupTestUser(ctx, testUser.ID)

		Convey("并发登录请求", func() {
			concurrency := 10
			results := make(chan *TestAPIResponse, concurrency)

			loginReq := map[string]interface{}{
				"username":  testUser.Username,
				"password":  testPassword,
				"tenant_id": tenantID,
			}

			// 并发发送登录请求
			for i := 0; i < concurrency; i++ {
				go func() {
					_, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
					if err != nil {
						results <- &TestAPIResponse{Code: 500, Msg: err.Error()}
					} else {
						results <- apiResp
					}
				}()
			}

			// 收集结果
			successCount := 0
			for i := 0; i < concurrency; i++ {
				result := <-results
				if result.Code == 200 {
					successCount++
				}
			}

			// 所有请求都应该成功
			So(successCount, ShouldEqual, concurrency)
		})

		Convey("并发刷新令牌请求", func() {
			// 先登录获取令牌
			loginReq := map[string]interface{}{
				"username":  testUser.Username,
				"password":  testPassword,
				"tenant_id": tenantID,
			}

			_, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq, nil)
			So(err, ShouldBeNil)
			So(apiResp.Code, ShouldEqual, 200)

			dataBytes, _ := json.Marshal(apiResp.Data)
			var loginData TestLoginData
			json.Unmarshal(dataBytes, &loginData)

			// 并发刷新令牌
			concurrency := 5
			results := make(chan *TestAPIResponse, concurrency)

			for i := 0; i < concurrency; i++ {
				go func() {
					refreshReq := map[string]interface{}{
						"refresh_token": loginData.RefreshToken,
					}

					_, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/refresh", refreshReq, nil)
					if err != nil {
						results <- &TestAPIResponse{Code: 500, Msg: err.Error()}
					} else {
						results <- apiResp
					}
				}()
			}

			// 收集结果
			successCount := 0
			failCount := 0
			for i := 0; i < concurrency; i++ {
				result := <-results
				if result.Code == 200 {
					successCount++
				} else {
					failCount++
				}
			}

			// 由于令牌轮换机制，只有一个请求应该成功，其他的应该失败
			So(successCount, ShouldEqual, 1)
			So(failCount, ShouldEqual, concurrency-1)
		})
	})
}
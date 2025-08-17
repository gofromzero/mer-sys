package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/services/user-service/internal/controller"
	. "github.com/smartystreets/goconvey/convey"
)

// TestAPIResponse API响应结构
type TestAPIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// setupTestServer 设置测试服务器（不初始化数据库连接）
func setupTestServer() *ghttp.Server {
	s := g.Server("test-unit")
	authController := controller.NewAuthController()

	// 配置路由
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Group("/auth", func(authGroup *ghttp.RouterGroup) {
			authGroup.POST("/login", authController.Login)
			authGroup.POST("/logout", authController.Logout)
			authGroup.POST("/refresh", authController.RefreshToken)
		})
	})

	return s
}

// makeHTTPRequest 发送HTTP请求并返回响应
func makeHTTPRequest(server *ghttp.Server, method, path string, body interface{}) (*http.Response, *TestAPIResponse, error) {
	var requestBody []byte
	var err error

	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(requestBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	resp := w.Result()

	// 解析响应体
	var apiResp TestAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return resp, nil, err
	}

	return resp, &apiResp, nil
}

// TestAuthAPIValidation 认证API参数验证测试
func TestAuthAPIValidation(t *testing.T) {
	Convey("认证API参数验证测试", t, func() {
		server := setupTestServer()

		Convey("登录参数验证", func() {
			Convey("缺少用户名", func() {
				loginReq := map[string]interface{}{
					"password":  "testpassword",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "用户名不能为空")
			})

			Convey("缺少密码", func() {
				loginReq := map[string]interface{}{
					"username":  "testuser",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "密码不能为空")
			})

			Convey("缺少租户ID", func() {
				loginReq := map[string]interface{}{
					"username": "testuser",
					"password": "testpassword",
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "租户ID不能为空")
			})

			Convey("用户名过短", func() {
				loginReq := map[string]interface{}{
					"username":  "ab",
					"password":  "testpassword",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "用户名长度为3-50字符")
			})

			Convey("用户名过长", func() {
				longUsername := make([]byte, 51) // 超过50字符限制
				for i := range longUsername {
					longUsername[i] = 'a'
				}

				loginReq := map[string]interface{}{
					"username":  string(longUsername),
					"password":  "testpassword",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "用户名长度为3-50字符")
			})

			Convey("密码过短", func() {
				loginReq := map[string]interface{}{
					"username":  "testuser",
					"password":  "12345",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "密码长度为6-128字符")
			})

			Convey("密码过长", func() {
				tooLongPassword := make([]byte, 129) // 超过128字符限制
				for i := range tooLongPassword {
					tooLongPassword[i] = 'b'
				}

				loginReq := map[string]interface{}{
					"username":  "testuser",
					"password":  string(tooLongPassword),
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "密码长度为6-128字符")
			})

			Convey("租户ID为0", func() {
				loginReq := map[string]interface{}{
					"username":  "testuser",
					"password":  "testpassword",
					"tenant_id": 0,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "租户ID必须大于0")
			})

			Convey("租户ID为负数", func() {
				loginReq := map[string]interface{}{
					"username":  "testuser",
					"password":  "testpassword",
					"tenant_id": -1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "租户ID必须大于0")
			})
		})

		Convey("刷新令牌参数验证", func() {
			Convey("缺少刷新令牌", func() {
				refreshReq := map[string]interface{}{}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/refresh", refreshReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldContainSubstring, "刷新令牌不能为空")
			})

			Convey("无效刷新令牌格式", func() {
				refreshReq := map[string]interface{}{
					"refresh_token": "invalid.token.format",
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/refresh", refreshReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 401)
				So(apiResp.Msg, ShouldEqual, "令牌刷新失败，请重新登录")
			})
		})

		Convey("登出参数验证", func() {
			Convey("缺少访问令牌", func() {
				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/logout", nil)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldEqual, "缺少访问令牌")
			})
		})
	})
}

// TestAuthAPIResponseFormat 认证API响应格式测试
func TestAuthAPIResponseFormat(t *testing.T) {
	Convey("认证API响应格式测试", t, func() {
		server := setupTestServer()

		Convey("JSON响应格式标准化", func() {
			Convey("登录请求返回标准JSON格式", func() {
				loginReq := map[string]interface{}{
					"username":  "testuser",
					"password":  "testpassword",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				So(resp.Header.Get("Content-Type"), ShouldContainSubstring, "application/json")

				// 响应结构验证
				So(apiResp.Code, ShouldBeIn, []int{200, 400, 401, 403, 500})
				So(apiResp.Msg, ShouldNotBeEmpty)
				// Data字段可以为nil或包含数据
			})

			Convey("参数错误时返回400状态码", func() {
				loginReq := map[string]interface{}{
					"username": "ab", // 过短
				}

				_, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(apiResp.Code, ShouldEqual, 400)
				So(apiResp.Msg, ShouldNotBeEmpty)
				So(apiResp.Data, ShouldBeNil)
			})
		})

		Convey("错误消息本地化", func() {
			Convey("中文错误消息", func() {
				loginReq := map[string]interface{}{
					"username": "ab",
				}

				_, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(apiResp.Code, ShouldEqual, 400)
				// 验证返回中文错误消息
				So(apiResp.Msg, ShouldContainSubstring, "用户名")
			})
		})
	})
}

// TestAuthAPISecurityHeaders 认证API安全头测试
func TestAuthAPISecurityHeaders(t *testing.T) {
	Convey("认证API安全头测试", t, func() {
		server := setupTestServer()

		Convey("响应头安全性", func() {
			loginReq := map[string]interface{}{
				"username":  "testuser",
				"password":  "testpassword",
				"tenant_id": 1,
			}

			resp, _, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
			So(err, ShouldBeNil)

			// 验证Content-Type是JSON
			So(resp.Header.Get("Content-Type"), ShouldContainSubstring, "application/json")
		})
	})
}

// TestAuthAPIEdgeCases 认证API边界条件测试
func TestAuthAPIEdgeCases(t *testing.T) {
	Convey("认证API边界条件测试", t, func() {
		server := setupTestServer()

		Convey("特殊字符处理", func() {
			Convey("用户名包含特殊字符", func() {
				loginReq := map[string]interface{}{
					"username":  "test@user.com",
					"password":  "testpassword",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				// 这里应该返回401而不是400，因为参数格式正确但用户不存在
				So(apiResp.Code, ShouldEqual, 401)
			})

			Convey("密码包含特殊字符", func() {
				loginReq := map[string]interface{}{
					"username":  "testuser",
					"password":  "test!@#$%^&*()_+",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				// 密码格式正确，应该返回401（用户不存在）而不是400
				So(apiResp.Code, ShouldEqual, 401)
			})
		})

		Convey("UTF-8编码处理", func() {
			Convey("中文用户名", func() {
				loginReq := map[string]interface{}{
					"username":  "测试用户",
					"password":  "testpassword",
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				// 格式正确但用户不存在
				So(apiResp.Code, ShouldEqual, 401)
			})
		})

		Convey("大型数据处理", func() {
			Convey("最大长度用户名和密码", func() {
				maxUsername := make([]byte, 50)
				for i := range maxUsername {
					maxUsername[i] = 'a'
				}

				maxPassword := make([]byte, 128)
				for i := range maxPassword {
					maxPassword[i] = 'b'
				}

				loginReq := map[string]interface{}{
					"username":  string(maxUsername),
					"password":  string(maxPassword),
					"tenant_id": 1,
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				// 应该通过参数验证，返回401（用户不存在）
				So(apiResp.Code, ShouldEqual, 401)
			})
		})

		Convey("极值租户ID处理", func() {
			Convey("最大uint64租户ID", func() {
				loginReq := map[string]interface{}{
					"username":  "testuser",
					"password":  "testpassword",
					"tenant_id": uint64(18446744073709551615),
				}

				resp, apiResp, err := makeHTTPRequest(server, "POST", "/api/v1/auth/login", loginReq)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
				// 应该通过参数验证，返回401（用户不存在）
				So(apiResp.Code, ShouldEqual, 401)
			})
		})
	})
}
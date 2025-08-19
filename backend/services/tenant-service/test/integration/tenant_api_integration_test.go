package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/services/tenant-service/internal/controller"
)

// 模拟HTTP服务器用于测试
func setupTestServer() *ghttp.Server {
	s := g.Server()
	s.SetPort(8082) // 使用不同端口避免冲突

	// 设置路由
	tenantController := controller.NewTenantController()
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.POST("/tenants", tenantController.Create)
		group.GET("/tenants", tenantController.List)
		group.GET("/tenants/:id", tenantController.GetByID)
		group.PUT("/tenants/:id", tenantController.Update)
		group.PUT("/tenants/:id/status", tenantController.UpdateStatus)
		group.GET("/tenants/:id/config", tenantController.GetConfig)
		group.PUT("/tenants/:id/config", tenantController.UpdateConfig)
		group.GET("/tenants/:id/config/notifications", tenantController.GetConfigNotification)
	})

	return s
}

func TestTenantAPIIntegration(t *testing.T) {
	Convey("租户API集成测试", t, func() {
		server := setupTestServer()
		defer server.Shutdown()

		// 用于存储创建的租户ID
		var createdTenantID uint64

		Convey("创建租户接口测试", func() {
			Convey("成功创建租户", func() {
				reqData := map[string]interface{}{
					"name":           "集成测试租户",
					"code":           "integration-test",
					"business_type":  "ecommerce",
					"contact_person": "测试人员",
					"contact_email":  "integration@test.com",
					"contact_phone":  "13800138000",
					"address":        "测试地址",
				}

				jsonData, _ := json.Marshal(reqData)
				req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)
				So(response["message"], ShouldEqual, "租户创建成功")

				// 保存创建的租户ID用于后续测试
				data := response["data"].(map[string]interface{})
				createdTenantID = uint64(data["id"].(float64))
				So(createdTenantID, ShouldBeGreaterThan, 0)
			})

			Convey("参数验证失败", func() {
				reqData := map[string]interface{}{
					"name": "", // 空名称应该失败
					"code": "",
				}

				jsonData, _ := json.Marshal(reqData)
				req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldNotEqual, 0)
			})

			Convey("重复代码创建失败", func() {
				// 先创建一个租户
				reqData := map[string]interface{}{
					"name":           "重复测试租户1",
					"code":           "duplicate-test",
					"business_type":  "retail",
					"contact_person": "测试人员1",
					"contact_email":  "duplicate1@test.com",
				}

				jsonData, _ := json.Marshal(reqData)
				req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)
				So(w.Code, ShouldEqual, http.StatusOK)

				// 尝试创建相同代码的租户
				reqData["name"] = "重复测试租户2"
				reqData["contact_email"] = "duplicate2@test.com"

				jsonData, _ = json.Marshal(reqData)
				req = httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w = httptest.NewRecorder()

				server.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldNotEqual, 0)
				So(response["error"], ShouldContainSubstring, "租户代码已存在")
			})
		})

		Convey("查询租户接口测试", func() {
			// 确保有测试数据
			if createdTenantID == 0 {
				// 创建测试租户
				reqData := map[string]interface{}{
					"name":           "查询测试租户",
					"code":           "query-test",
					"business_type":  "service",
					"contact_person": "查询测试",
					"contact_email":  "query@test.com",
				}

				jsonData, _ := json.Marshal(reqData)
				req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				data := response["data"].(map[string]interface{})
				createdTenantID = uint64(data["id"].(float64))
			}

			Convey("获取租户列表", func() {
				req := httptest.NewRequest("GET", "/api/v1/tenants", nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				data := response["data"].(map[string]interface{})
				So(data["total"], ShouldBeGreaterThan, 0)
				So(data["tenants"], ShouldNotBeNil)
			})

			Convey("分页查询租户", func() {
				req := httptest.NewRequest("GET", "/api/v1/tenants?page=1&page_size=5", nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				data := response["data"].(map[string]interface{})
				So(data["page"], ShouldEqual, 1)
				So(data["size"], ShouldBeLessThanOrEqualTo, 5)
			})

			Convey("按状态筛选租户", func() {
				req := httptest.NewRequest("GET", "/api/v1/tenants?status=active", nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)
			})

			Convey("搜索租户", func() {
				req := httptest.NewRequest("GET", "/api/v1/tenants?search=测试", nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)
			})

			Convey("根据ID获取租户详情", func() {
				url := fmt.Sprintf("/api/v1/tenants/%d", createdTenantID)
				req := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				data := response["data"].(map[string]interface{})
				So(data["id"], ShouldEqual, createdTenantID)
			})

			Convey("获取不存在的租户", func() {
				req := httptest.NewRequest("GET", "/api/v1/tenants/99999", nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 404)
			})
		})

		Convey("更新租户接口测试", func() {
			Convey("更新租户基本信息", func() {
				reqData := map[string]interface{}{
					"name":           "更新后的租户名称",
					"business_type":  "manufacturing",
					"contact_person": "更新后的联系人",
				}

				jsonData, _ := json.Marshal(reqData)
				url := fmt.Sprintf("/api/v1/tenants/%d", createdTenantID)
				req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				data := response["data"].(map[string]interface{})
				So(data["name"], ShouldEqual, reqData["name"])
				So(data["business_type"], ShouldEqual, reqData["business_type"])
			})

			Convey("更新不存在的租户", func() {
				reqData := map[string]interface{}{
					"name": "不存在的租户",
				}

				jsonData, _ := json.Marshal(reqData)
				req := httptest.NewRequest("PUT", "/api/v1/tenants/99999", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldNotEqual, 0)
			})
		})

		Convey("租户状态管理接口测试", func() {
			Convey("更新租户状态", func() {
				reqData := map[string]interface{}{
					"status": "suspended",
					"reason": "集成测试暂停",
				}

				jsonData, _ := json.Marshal(reqData)
				url := fmt.Sprintf("/api/v1/tenants/%d/status", createdTenantID)
				req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				// 验证状态已更新
				getUrl := fmt.Sprintf("/api/v1/tenants/%d", createdTenantID)
				getReq := httptest.NewRequest("GET", getUrl, nil)
				getW := httptest.NewRecorder()

				server.ServeHTTP(getW, getReq)

				var getResponse map[string]interface{}
				json.Unmarshal(getW.Body.Bytes(), &getResponse)
				data := getResponse["data"].(map[string]interface{})
				So(data["status"], ShouldEqual, "suspended")
			})

			Convey("状态参数验证", func() {
				reqData := map[string]interface{}{
					"status": "invalid_status",
					"reason": "无效状态测试",
				}

				jsonData, _ := json.Marshal(reqData)
				url := fmt.Sprintf("/api/v1/tenants/%d/status", createdTenantID)
				req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldNotEqual, 0)
			})
		})

		Convey("租户配置管理接口测试", func() {
			Convey("获取租户配置", func() {
				url := fmt.Sprintf("/api/v1/tenants/%d/config", createdTenantID)
				req := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				data := response["data"].(map[string]interface{})
				So(data["max_users"], ShouldNotBeNil)
				So(data["max_merchants"], ShouldNotBeNil)
				So(data["features"], ShouldNotBeNil)
			})

			Convey("更新租户配置", func() {
				reqData := map[string]interface{}{
					"max_users":     200,
					"max_merchants": 100,
					"features":      []string{"basic", "advanced_report"},
					"settings": map[string]string{
						"theme": "dark",
						"lang":  "zh-CN",
					},
				}

				jsonData, _ := json.Marshal(reqData)
				url := fmt.Sprintf("/api/v1/tenants/%d/config", createdTenantID)
				req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				So(response["code"], ShouldEqual, 0)

				// 验证配置已更新
				getUrl := fmt.Sprintf("/api/v1/tenants/%d/config", createdTenantID)
				getReq := httptest.NewRequest("GET", getUrl, nil)
				getW := httptest.NewRecorder()

				server.ServeHTTP(getW, getReq)

				var getResponse map[string]interface{}
				json.Unmarshal(getW.Body.Bytes(), &getResponse)
				data := getResponse["data"].(map[string]interface{})
				So(data["max_users"], ShouldEqual, 200)
				So(data["max_merchants"], ShouldEqual, 100)
			})

			Convey("获取配置变更通知", func() {
				url := fmt.Sprintf("/api/v1/tenants/%d/config/notifications", createdTenantID)
				req := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()

				server.ServeHTTP(w, req)

				// 可能有通知也可能没有，都是正常的
				So(w.Code, ShouldEqual, http.StatusOK)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				So(err, ShouldBeNil)
				// code可能是0（有通知）或404（无通知）
				So(response["code"], ShouldBeIn, []interface{}{0, 404})
			})
		})
	})
}

func TestTenantAPIErrorHandling(t *testing.T) {
	Convey("租户API错误处理测试", t, func() {
		server := setupTestServer()
		defer server.Shutdown()

		Convey("无效JSON格式", func() {
			req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer([]byte("invalid json")))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response["code"], ShouldEqual, 400)
		})

		Convey("无效租户ID", func() {
			req := httptest.NewRequest("GET", "/api/v1/tenants/invalid_id", nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response["code"], ShouldEqual, 400)
			So(response["message"], ShouldContainSubstring, "租户ID格式错误")
		})

		Convey("空请求体", func() {
			req := httptest.NewRequest("POST", "/api/v1/tenants", nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			So(w.Code, ShouldEqual, http.StatusOK)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response["code"], ShouldNotEqual, 0)
		})
	})
}

func TestTenantAPIPerformance(t *testing.T) {
	Convey("租户API性能测试", t, func() {
		server := setupTestServer()
		defer server.Shutdown()

		Convey("批量创建租户性能", func() {
			start := context.Background()
			ctx, cancel := context.WithTimeout(start, 5*1000000000) // 5秒超时
			defer cancel()

			successCount := 0
			totalRequests := 10

			for i := 0; i < totalRequests; i++ {
				select {
				case <-ctx.Done():
					t.Logf("超时，成功创建 %d/%d 个租户", successCount, totalRequests)
					return
				default:
					reqData := map[string]interface{}{
						"name":           fmt.Sprintf("性能测试租户%d", i),
						"code":           fmt.Sprintf("perf-test-%d", i),
						"business_type":  "test",
						"contact_person": fmt.Sprintf("测试人员%d", i),
						"contact_email":  fmt.Sprintf("perf%d@test.com", i),
					}

					jsonData, _ := json.Marshal(reqData)
					req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(jsonData))
					req.Header.Set("Content-Type", "application/json")
					w := httptest.NewRecorder()

					server.ServeHTTP(w, req)

					if w.Code == http.StatusOK {
						var response map[string]interface{}
						json.Unmarshal(w.Body.Bytes(), &response)
						if response["code"] == float64(0) {
							successCount++
						}
					}
				}
			}

			So(successCount, ShouldBeGreaterThan, totalRequests/2) // 至少50%成功率
		})

		Convey("并发查询性能", func() {
			// 先创建一个测试租户
			reqData := map[string]interface{}{
				"name":           "并发测试租户",
				"code":           "concurrent-test",
				"business_type":  "test",
				"contact_person": "并发测试",
				"contact_email":  "concurrent@test.com",
			}

			jsonData, _ := json.Marshal(reqData)
			req := httptest.NewRequest("POST", "/api/v1/tenants", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			tenantID := uint64(data["id"].(float64))

			// 并发查询测试
			concurrency := 5
			done := make(chan bool, concurrency)
			successCount := 0

			for i := 0; i < concurrency; i++ {
				go func() {
					defer func() { done <- true }()

					url := fmt.Sprintf("/api/v1/tenants/%d", tenantID)
					req := httptest.NewRequest("GET", url, nil)
					w := httptest.NewRecorder()

					server.ServeHTTP(w, req)

					if w.Code == http.StatusOK {
						var response map[string]interface{}
						json.Unmarshal(w.Body.Bytes(), &response)
						if response["code"] == float64(0) {
							successCount++
						}
					}
				}()
			}

			// 等待所有协程完成
			for i := 0; i < concurrency; i++ {
				<-done
			}

			So(successCount, ShouldEqual, concurrency) // 所有请求都应该成功
		})
	})
}
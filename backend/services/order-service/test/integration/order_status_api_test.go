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

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOrderStatusAPI(t *testing.T) {
	Convey("订单状态管理API集成测试", t, func() {
		// 初始化HTTP服务器
		s := g.Server()

		// 注册中间件
		authMiddleware := middleware.NewAuthMiddleware()
		s.Group("/api/v1", func(group *ghttp.RouterGroup) {
			group.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)

			// 订单基础路由
			orderController := controller.NewOrderController()
			group.Group("/orders", func(orderGroup *ghttp.RouterGroup) {
				orderGroup.POST("/", orderController.CreateOrder)
				orderGroup.GET("/{id}", orderController.GetOrder)
				orderGroup.GET("/", orderController.ListOrders)
				orderGroup.PUT("/{id}/cancel", orderController.CancelOrder)

				// 新增的订单查询路由
				orderGroup.GET("/query", orderController.QueryOrders)
				orderGroup.GET("/{id}/detail", orderController.GetOrderDetail)
				orderGroup.GET("/search", orderController.SearchOrders)
				orderGroup.GET("/stats", orderController.GetOrderStats)
			})

			// 订单状态管理路由
			orderStatusController := controller.NewOrderStatusController()
			group.Group("/orders", func(orderGroup *ghttp.RouterGroup) {
				orderGroup.PUT("/{id}/status", orderStatusController.UpdateOrderStatus)
				orderGroup.GET("/{id}/status-history", orderStatusController.GetOrderStatusHistory)
				orderGroup.POST("/batch-update-status", orderStatusController.BatchUpdateOrderStatus)
			})

			// 订单超时管理路由
			orderTimeoutController := controller.NewOrderTimeoutController()
			group.Group("/orders/timeout", func(timeoutGroup *ghttp.RouterGroup) {
				timeoutGroup.POST("/process", orderTimeoutController.ProcessTimeoutOrders)
				timeoutGroup.GET("/statistics", orderTimeoutController.GetTimeoutStatistics)
				timeoutGroup.POST("/start-monitoring", orderTimeoutController.StartMonitoring)
				timeoutGroup.POST("/stop-monitoring", orderTimeoutController.StopMonitoring)
			})

			// 超时配置路由
			orderTimeoutConfigController := controller.NewOrderTimeoutConfigController()
			group.Group("/timeout-configs", func(configGroup *ghttp.RouterGroup) {
				configGroup.POST("/", orderTimeoutConfigController.CreateConfig)
				configGroup.GET("/{id}", orderTimeoutConfigController.GetConfig)
				configGroup.PUT("/{id}", orderTimeoutConfigController.UpdateConfig)
				configGroup.DELETE("/{id}", orderTimeoutConfigController.DeleteConfig)
				configGroup.GET("/", orderTimeoutConfigController.ListConfigs)
			})
		})

		testServer := httptest.NewServer(s.Handler())
		defer testServer.Close()

		Convey("订单状态管理API测试", func() {
			token := generateTestJWT(1, 1)

			// 首先创建一个测试订单
			orderReqBody := map[string]interface{}{
				"merchant_id": 1,
				"items": []map[string]interface{}{
					{
						"product_id": 1001,
						"quantity":   1,
					},
				},
			}

			body, _ := json.Marshal(orderReqBody)
			orderReq, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders", bytes.NewBuffer(body))
			orderReq.Header.Set("Authorization", "Bearer "+token)
			orderReq.Header.Set("X-Tenant-ID", "1")
			orderReq.Header.Set("Content-Type", "application/json")

			orderResp, err := http.DefaultClient.Do(orderReq)
			So(err, ShouldBeNil)
			So(orderResp.StatusCode, ShouldEqual, 200)

			var orderResult map[string]interface{}
			json.NewDecoder(orderResp.Body).Decode(&orderResult)
			orderData := orderResult["data"].(map[string]interface{})
			orderID := int(orderData["id"].(float64))

			Convey("更新订单状态", func() {
				updateReqBody := map[string]interface{}{
					"status":        2, // types.OrderStatusIntPaid
					"reason":        "客户完成支付",
					"operator_type": "customer",
				}

				body, _ := json.Marshal(updateReqBody)
				req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/orders/%d/status", testServer.URL, orderID), bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)

				updatedOrder := result["data"].(map[string]interface{})
				So(updatedOrder["status"], ShouldEqual, "paid")
			})

			Convey("获取订单状态历史", func() {
				req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/orders/%d/status-history", testServer.URL, orderID), nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)

				historyList := result["data"].([]interface{})
				So(len(historyList), ShouldBeGreaterThan, 0)

				// 验证最新的状态历史记录
				latestHistory := historyList[0].(map[string]interface{})
				So(latestHistory["to_status"], ShouldEqual, 2) // OrderStatusIntPaid
				So(latestHistory["reason"], ShouldEqual, "客户完成支付")
				So(latestHistory["operator_type"], ShouldEqual, "customer")
			})

			Convey("无效状态流转", func() {
				// 尝试从已支付直接跳转到已完成（跳过处理中）
				updateReqBody := map[string]interface{}{
					"status":        4, // types.OrderStatusIntCompleted
					"reason":        "直接完成订单",
					"operator_type": "merchant",
				}

				body, _ := json.Marshal(updateReqBody)
				req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/orders/%d/status", testServer.URL, orderID), bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 400)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldNotEqual, 0)
				So(result["message"], ShouldContainSubstring, "无效的状态流转")
			})
		})

		Convey("批量订单状态更新API测试", func() {
			token := generateTestJWT(1, 1)

			// 创建多个测试订单
			var orderIDs []int
			for i := 0; i < 3; i++ {
				orderReqBody := map[string]interface{}{
					"merchant_id": 1,
					"items": []map[string]interface{}{
						{
							"product_id": 1001 + i,
							"quantity":   1,
						},
					},
				}

				body, _ := json.Marshal(orderReqBody)
				orderReq, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders", bytes.NewBuffer(body))
				orderReq.Header.Set("Authorization", "Bearer "+token)
				orderReq.Header.Set("X-Tenant-ID", "1")
				orderReq.Header.Set("Content-Type", "application/json")

				orderResp, err := http.DefaultClient.Do(orderReq)
				So(err, ShouldBeNil)

				var orderResult map[string]interface{}
				json.NewDecoder(orderResp.Body).Decode(&orderResult)
				orderData := orderResult["data"].(map[string]interface{})
				orderID := int(orderData["id"].(float64))
				orderIDs = append(orderIDs, orderID)
			}

			Convey("成功批量更新状态", func() {
				batchReqBody := map[string]interface{}{
					"order_ids":     orderIDs,
					"status":        5, // types.OrderStatusIntCancelled
					"reason":        "批量取消订单",
					"operator_type": "merchant",
				}

				body, _ := json.Marshal(batchReqBody)
				req, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders/batch-update-status", bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)

				batchResult := result["data"].(map[string]interface{})
				So(batchResult["success_count"], ShouldEqual, 3)
				So(batchResult["failure_count"], ShouldEqual, 0)
			})
		})

		Convey("订单高级查询API测试", func() {
			token := generateTestJWT(1, 1)

			Convey("订单列表查询", func() {
				// 测试基础查询
				req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/orders/query?page=1&page_size=10", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)

				data := result["data"].(map[string]interface{})
				So(data["items"], ShouldNotBeNil)
				So(data["total"], ShouldBeGreaterThanOrEqualTo, 0)
				So(data["page"], ShouldEqual, 1)
				So(data["page_size"], ShouldEqual, 10)
			})

			Convey("订单状态筛选查询", func() {
				req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/orders/query?status=pending,paid&page=1&page_size=10", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
			})

			Convey("订单搜索", func() {
				req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/orders/search?keyword=ORD&page=1&page_size=10", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
			})

			Convey("订单统计信息", func() {
				req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/orders/stats", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)

				stats := result["data"].(map[string]interface{})
				So(stats["total_count"], ShouldBeGreaterThanOrEqualTo, 0)
				So(stats["pending_count"], ShouldBeGreaterThanOrEqualTo, 0)
				So(stats["paid_count"], ShouldBeGreaterThanOrEqualTo, 0)
			})
		})

		Convey("订单超时管理API测试", func() {
			token := generateTestJWT(1, 1)

			Convey("获取超时统计", func() {
				req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/orders/timeout/statistics", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)

				stats := result["data"].(map[string]interface{})
				So(stats["payment_timeout_count"], ShouldBeGreaterThanOrEqualTo, 0)
				So(stats["processing_timeout_count"], ShouldBeGreaterThanOrEqualTo, 0)
			})

			Convey("手动处理超时订单", func() {
				req, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders/timeout/process", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)
			})

			Convey("启动超时监控", func() {
				req, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders/timeout/start-monitoring", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)

				// 停止监控以清理资源
				stopReq, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders/timeout/stop-monitoring", nil)
				stopReq.Header.Set("Authorization", "Bearer "+token)
				stopReq.Header.Set("X-Tenant-ID", "1")
				http.DefaultClient.Do(stopReq)
			})
		})

		Convey("超时配置管理API测试", func() {
			token := generateTestJWT(1, 1)

			Convey("创建超时配置", func() {
				configReqBody := map[string]interface{}{
					"merchant_id":              1,
					"payment_timeout_minutes":  45,
					"processing_timeout_hours": 48,
					"auto_complete_enabled":    true,
				}

				body, _ := json.Marshal(configReqBody)
				req, _ := http.NewRequest("POST", testServer.URL+"/api/v1/timeout-configs", bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)

				configData := result["data"].(map[string]interface{})
				So(configData["payment_timeout_minutes"], ShouldEqual, 45)
				So(configData["processing_timeout_hours"], ShouldEqual, 48)
				So(configData["auto_complete_enabled"], ShouldBeTrue)
			})

			Convey("获取超时配置列表", func() {
				req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/timeout-configs?page=1&page_size=10", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
				So(result["data"], ShouldNotBeNil)
			})
		})

		Convey("多租户隔离API测试", func() {
			tenant1Token := generateTestJWT(1, 1)
			tenant2Token := generateTestJWT(2, 2)

			// 租户1创建订单
			orderReqBody := map[string]interface{}{
				"merchant_id": 1,
				"items": []map[string]interface{}{
					{
						"product_id": 1001,
						"quantity":   1,
					},
				},
			}

			body, _ := json.Marshal(orderReqBody)
			orderReq, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders", bytes.NewBuffer(body))
			orderReq.Header.Set("Authorization", "Bearer "+tenant1Token)
			orderReq.Header.Set("X-Tenant-ID", "1")
			orderReq.Header.Set("Content-Type", "application/json")

			orderResp, err := http.DefaultClient.Do(orderReq)
			So(err, ShouldBeNil)

			var orderResult map[string]interface{}
			json.NewDecoder(orderResp.Body).Decode(&orderResult)
			orderData := orderResult["data"].(map[string]interface{})
			orderID := int(orderData["id"].(float64))

			Convey("跨租户访问订单状态历史", func() {
				// 租户2尝试访问租户1的订单状态历史
				req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/orders/%d/status-history", testServer.URL, orderID), nil)
				req.Header.Set("Authorization", "Bearer "+tenant2Token)
				req.Header.Set("X-Tenant-ID", "2")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 404) // 应该返回不存在

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldNotEqual, 0)
			})

			Convey("跨租户更新订单状态", func() {
				// 租户2尝试更新租户1的订单状态
				updateReqBody := map[string]interface{}{
					"status":        5, // types.OrderStatusIntCancelled
					"reason":        "跨租户操作尝试",
					"operator_type": "system",
				}

				body, _ := json.Marshal(updateReqBody)
				req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/orders/%d/status", testServer.URL, orderID), bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+tenant2Token)
				req.Header.Set("X-Tenant-ID", "2")
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 404) // 应该返回不存在

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldNotEqual, 0)
			})
		})

		Convey("错误处理API测试", func() {
			token := generateTestJWT(1, 1)

			Convey("不存在的订单状态更新", func() {
				updateReqBody := map[string]interface{}{
					"status":        5,
					"reason":        "测试不存在订单",
					"operator_type": "system",
				}

				body, _ := json.Marshal(updateReqBody)
				req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/orders/999999/status", testServer.URL), bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 404)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldNotEqual, 0)
			})

			Convey("无效的状态值", func() {
				// 先创建订单
				orderReqBody := map[string]interface{}{
					"merchant_id": 1,
					"items": []map[string]interface{}{
						{
							"product_id": 1001,
							"quantity":   1,
						},
					},
				}

				body, _ := json.Marshal(orderReqBody)
				orderReq, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders", bytes.NewBuffer(body))
				orderReq.Header.Set("Authorization", "Bearer "+token)
				orderReq.Header.Set("X-Tenant-ID", "1")
				orderReq.Header.Set("Content-Type", "application/json")

				orderResp, err := http.DefaultClient.Do(orderReq)
				So(err, ShouldBeNil)

				var orderResult map[string]interface{}
				json.NewDecoder(orderResp.Body).Decode(&orderResult)
				orderData := orderResult["data"].(map[string]interface{})
				orderID := int(orderData["id"].(float64))

				// 尝试设置无效状态
				updateReqBody := map[string]interface{}{
					"status":        99, // 无效状态值
					"reason":        "测试无效状态",
					"operator_type": "system",
				}

				body, _ = json.Marshal(updateReqBody)
				req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/orders/%d/status", testServer.URL, orderID), bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 400)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldNotEqual, 0)
			})
		})
	})
}
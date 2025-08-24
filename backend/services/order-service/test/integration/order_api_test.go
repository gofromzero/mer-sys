package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOrderAPI(t *testing.T) {
	Convey("订单API集成测试", t, func() {
		// 初始化HTTP服务器
		s := g.Server()

		// 注册中间件
		authMiddleware := middleware.NewAuthMiddleware()
		s.Group("/api/v1", func(group *ghttp.RouterGroup) {
			group.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)

			// 订单路由
			orderController := controller.NewOrderController()
			group.Group("/orders", func(orderGroup *ghttp.RouterGroup) {
				orderGroup.POST("/", orderController.CreateOrder)
				orderGroup.GET("/{id}", orderController.GetOrder)
				orderGroup.GET("/", orderController.ListOrders)
				orderGroup.PUT("/{id}/cancel", orderController.CancelOrder)
				orderGroup.POST("/confirmation", orderController.GetOrderConfirmation)
			})

			// 购物车路由
			cartController := controller.NewCartController()
			group.Group("/cart", func(cartGroup *ghttp.RouterGroup) {
				cartGroup.GET("/", cartController.GetCart)
				cartGroup.POST("/items", cartController.AddItem)
				cartGroup.PUT("/items/{id}", cartController.UpdateItem)
				cartGroup.DELETE("/items/{id}", cartController.RemoveItem)
				cartGroup.DELETE("/", cartController.ClearCart)
			})

			// 支付路由
			paymentController := controller.NewPaymentController()
			group.Group("/payments", func(paymentGroup *ghttp.RouterGroup) {
				paymentGroup.POST("/orders/{id}/pay", paymentController.InitiatePayment)
				paymentGroup.GET("/orders/{id}/status", paymentController.GetPaymentStatus)
				paymentGroup.POST("/orders/{id}/retry", paymentController.RetryPayment)
			})
		})

		// 支付回调（无需认证）
		s.BindHandler("POST:/api/v1/payments/callback/alipay", controller.NewPaymentController().HandleAlipayCallback)

		testServer := httptest.NewServer(s.Handler())
		defer testServer.Close()

		Convey("购物车API测试", func() {
			token := generateTestJWT(1, 1) // customer_id=1, tenant_id=1

			Convey("获取购物车", func() {
				req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/cart", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
			})

			Convey("添加商品到购物车", func() {
				reqBody := map[string]interface{}{
					"product_id": 1001,
					"quantity":   2,
				}

				body, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest("POST", testServer.URL+"/api/v1/cart/items", bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
			})
		})

		Convey("订单API测试", func() {
			token := generateTestJWT(1, 1)

			Convey("获取订单确认信息", func() {
				reqBody := map[string]interface{}{
					"merchant_id": 1,
					"items": []map[string]interface{}{
						{
							"product_id": 1001,
							"quantity":   2,
						},
					},
				}

				body, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders/confirmation", bytes.NewBuffer(body))
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
			})

			Convey("创建订单", func() {
				reqBody := map[string]interface{}{
					"merchant_id": 1,
					"items": []map[string]interface{}{
						{
							"product_id": 1001,
							"quantity":   1,
						},
					},
				}

				body, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders", bytes.NewBuffer(body))
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

				orderData := result["data"].(map[string]interface{})
				So(orderData["customer_id"], ShouldEqual, 1)
				So(orderData["status"], ShouldEqual, "pending")
			})

			Convey("获取订单列表", func() {
				req, _ := http.NewRequest("GET", testServer.URL+"/api/v1/orders", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
			})
		})

		Convey("支付API测试", func() {
			token := generateTestJWT(1, 1)

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
			So(orderResp.StatusCode, ShouldEqual, 200)

			var orderResult map[string]interface{}
			json.NewDecoder(orderResp.Body).Decode(&orderResult)
			orderData := orderResult["data"].(map[string]interface{})
			orderID := int(orderData["id"].(float64))

			Convey("发起支付", func() {
				payReqBody := map[string]interface{}{
					"payment_method": "alipay",
					"return_url":     "http://localhost:3000/payment/success",
				}

				body, _ := json.Marshal(payReqBody)
				req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/payments/orders/%d/pay", testServer.URL, orderID), bytes.NewBuffer(body))
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
			})

			Convey("查询支付状态", func() {
				req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/payments/orders/%d/status", testServer.URL, orderID), nil)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("X-Tenant-ID", "1")

				resp, err := http.DefaultClient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)

				var result map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&result)
				So(result["code"], ShouldEqual, 0)
			})
		})

		Convey("多租户隔离测试", func() {
			tenant1Token := generateTestJWT(1, 1) // customer_id=1, tenant_id=1
			tenant2Token := generateTestJWT(2, 2) // customer_id=2, tenant_id=2

			// 租户1创建订单
			reqBody := map[string]interface{}{
				"merchant_id": 1,
				"items": []map[string]interface{}{
					{
						"product_id": 1001,
						"quantity":   1,
					},
				},
			}

			body, _ := json.Marshal(reqBody)
			req1, _ := http.NewRequest("POST", testServer.URL+"/api/v1/orders", bytes.NewBuffer(body))
			req1.Header.Set("Authorization", "Bearer "+tenant1Token)
			req1.Header.Set("X-Tenant-ID", "1")
			req1.Header.Set("Content-Type", "application/json")

			resp1, err := http.DefaultClient.Do(req1)
			So(err, ShouldBeNil)
			So(resp1.StatusCode, ShouldEqual, 200)

			// 租户2尝试获取租户1的订单列表（应该为空）
			req2, _ := http.NewRequest("GET", testServer.URL+"/api/v1/orders", nil)
			req2.Header.Set("Authorization", "Bearer "+tenant2Token)
			req2.Header.Set("X-Tenant-ID", "2")

			resp2, err := http.DefaultClient.Do(req2)
			So(err, ShouldBeNil)
			So(resp2.StatusCode, ShouldEqual, 200)

			var result map[string]interface{}
			json.NewDecoder(resp2.Body).Decode(&result)
			So(result["code"], ShouldEqual, 0)

			// 租户2应该看不到租户1的订单
			data := result["data"].(map[string]interface{})
			orders := data["orders"].([]interface{})
			So(len(orders), ShouldEqual, 0)
		})
	})
}

// generateTestJWT 生成测试用的JWT token
func generateTestJWT(customerID, tenantID uint64) string {
	// 这里应该使用实际的JWT生成逻辑
	// 为了简化测试，返回一个Mock token
	return fmt.Sprintf("test_jwt_token_customer_%d_tenant_%d", customerID, tenantID)
}

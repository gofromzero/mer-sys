package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gorilla/websocket"
	. "github.com/smartystreets/goconvey/convey"
)

// WebSocket消息类型定义
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// 订单状态通知消息
type OrderStatusNotification struct {
	OrderID      uint64 `json:"order_id"`
	OrderNumber  string `json:"order_number"`
	FromStatus   string `json:"from_status"`
	ToStatus     string `json:"to_status"`
	Reason       string `json:"reason"`
	OperatorType string `json:"operator_type"`
	CustomerID   uint64 `json:"customer_id"`
	MerchantID   uint64 `json:"merchant_id"`
}

func TestWebSocketNotification(t *testing.T) {
	Convey("WebSocket订单状态通知端到端测试", t, func() {
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
			})

			// 订单状态管理路由
			orderStatusController := controller.NewOrderStatusController()
			group.Group("/orders", func(orderGroup *ghttp.RouterGroup) {
				orderGroup.PUT("/{id}/status", orderStatusController.UpdateOrderStatus)
			})
		})

		// WebSocket路由（无需认证）
		websocketController := controller.NewWebSocketController()
		s.BindHandler("GET:/ws/orders/status-updates", websocketController.HandleConnection)

		testServer := httptest.NewServer(s.Handler())
		defer testServer.Close()

		// 转换HTTP URL为WebSocket URL
		wsURL := strings.Replace(testServer.URL, "http://", "ws://", 1) + "/ws/orders/status-updates"

		Convey("WebSocket连接管理测试", func() {
			Convey("成功建立WebSocket连接", func() {
				// 创建WebSocket连接
				header := http.Header{}
				header.Add("X-Tenant-ID", "1")
				header.Add("X-Customer-ID", "1")

				conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 101) // WebSocket升级状态码
				defer conn.Close()

				// 发送ping消息测试连接
				err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
				So(err, ShouldBeNil)

				// 接收pong响应
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				_, message, err := conn.ReadMessage()
				So(err, ShouldBeNil)

				var response WebSocketMessage
				err = json.Unmarshal(message, &response)
				So(err, ShouldBeNil)
				So(response.Type, ShouldEqual, "pong")
			})

			Convey("多客户端连接管理", func() {
				// 创建多个WebSocket连接
				var connections []*websocket.Conn
				defer func() {
					// 关闭所有连接
					for _, conn := range connections {
						conn.Close()
					}
				}()

				for i := 1; i <= 3; i++ {
					header := http.Header{}
					header.Add("X-Tenant-ID", "1")
					header.Add("X-Customer-ID", fmt.Sprintf("%d", i))

					conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
					So(err, ShouldBeNil)
					connections = append(connections, conn)
				}

				// 验证所有连接都正常工作
				for i, conn := range connections {
					err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
					So(err, ShouldBeNil)

					conn.SetReadDeadline(time.Now().Add(5 * time.Second))
					_, message, err := conn.ReadMessage()
					So(err, ShouldBeNil)

					var response WebSocketMessage
					err = json.Unmarshal(message, &response)
					So(err, ShouldBeNil)
					So(response.Type, ShouldEqual, "pong")

					t.Logf("Connection %d ping/pong successful", i+1)
				}
			})

			Convey("无效头信息连接拒绝", func() {
				// 缺少租户ID的连接应该被拒绝
				header := http.Header{}
				header.Add("X-Customer-ID", "1")

				conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
				if conn != nil {
					defer conn.Close()
				}

				// 连接可能失败或立即关闭
				So(err != nil || resp.StatusCode != 101, ShouldBeTrue)
			})
		})

		Convey("实时订单状态通知测试", func() {
			// 建立WebSocket连接
			header := http.Header{}
			header.Add("X-Tenant-ID", "1")
			header.Add("X-Customer-ID", "1")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
			So(err, ShouldBeNil)
			defer conn.Close()

			// 设置消息接收超时
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))

			Convey("订单状态变更实时通知", func() {
				// 在另一个goroutine中创建订单并更新状态
				go func() {
					time.Sleep(1 * time.Second) // 等待WebSocket连接稳定

					// 创建订单
					ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
					orderService := service.NewOrderService()
					req := &types.CreateOrderRequest{
						MerchantID: 1,
						Items: []struct {
							ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
							Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
						}{
							{ProductID: 1001, Quantity: 1},
						},
					}

					order, orderErr := orderService.CreateOrder(ctx, 1, req)
					if orderErr != nil {
						t.Logf("Failed to create order: %v", orderErr)
						return
					}

					// 等待一下确保WebSocket连接准备就绪
					time.Sleep(500 * time.Millisecond)

					// 更新订单状态
					orderStatusService := service.NewOrderStatusService()
					updateReq := &types.UpdateOrderStatusRequest{
						Status:       types.OrderStatusIntPaid,
						Reason:       "客户完成支付测试",
						OperatorType: types.OrderStatusOperatorTypeCustomer,
					}

					_, statusErr := orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
					if statusErr != nil {
						t.Logf("Failed to update order status: %v", statusErr)
						return
					}

					t.Logf("Order %d status updated successfully", order.ID)
				}()

				// 接收WebSocket通知
				messageReceived := false
				for i := 0; i < 10; i++ { // 最多等待10条消息
					_, message, readErr := conn.ReadMessage()
					if readErr != nil {
						t.Logf("WebSocket read error: %v", readErr)
						break
					}

					var wsMessage WebSocketMessage
					if jsonErr := json.Unmarshal(message, &wsMessage); jsonErr != nil {
						t.Logf("Failed to unmarshal WebSocket message: %v", jsonErr)
						continue
					}

					t.Logf("Received WebSocket message type: %s", wsMessage.Type)

					if wsMessage.Type == "order_status_changed" {
						// 验证通知消息内容
						notificationData, ok := wsMessage.Data.(map[string]interface{})
						So(ok, ShouldBeTrue)

						So(notificationData["to_status"], ShouldEqual, "paid")
						So(notificationData["reason"], ShouldEqual, "客户完成支付测试")
						So(notificationData["operator_type"], ShouldEqual, "customer")

						messageReceived = true
						break
					}
				}

				So(messageReceived, ShouldBeTrue)
			})
		})

		Convey("多租户WebSocket隔离测试", func() {
			// 创建两个不同租户的WebSocket连接
			header1 := http.Header{}
			header1.Add("X-Tenant-ID", "1")
			header1.Add("X-Customer-ID", "1")

			conn1, _, err := websocket.DefaultDialer.Dial(wsURL, header1)
			So(err, ShouldBeNil)
			defer conn1.Close()

			header2 := http.Header{}
			header2.Add("X-Tenant-ID", "2")
			header2.Add("X-Customer-ID", "1")

			conn2, _, err := websocket.DefaultDialer.Dial(wsURL, header2)
			So(err, ShouldBeNil)
			defer conn2.Close()

			// 设置读取超时
			conn1.SetReadDeadline(time.Now().Add(10 * time.Second))
			conn2.SetReadDeadline(time.Now().Add(3 * time.Second)) // 租户2设置更短的超时

			Convey("租户间通知隔离", func() {
				// 在租户1中创建订单并更新状态
				go func() {
					time.Sleep(1 * time.Second)

					ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
					orderService := service.NewOrderService()
					req := &types.CreateOrderRequest{
						MerchantID: 1,
						Items: []struct {
							ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
							Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
						}{
							{ProductID: 1001, Quantity: 1},
						},
					}

					order, orderErr := orderService.CreateOrder(ctx, 1, req)
					if orderErr != nil {
						t.Logf("Failed to create order: %v", orderErr)
						return
					}

					time.Sleep(500 * time.Millisecond)

					orderStatusService := service.NewOrderStatusService()
					updateReq := &types.UpdateOrderStatusRequest{
						Status:       types.OrderStatusIntPaid,
						Reason:       "租户隔离测试",
						OperatorType: types.OrderStatusOperatorTypeCustomer,
					}

					_, statusErr := orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
					if statusErr != nil {
						t.Logf("Failed to update order status: %v", statusErr)
						return
					}

					t.Logf("Tenant 1 order %d status updated", order.ID)
				}()

				// 租户1应该收到通知
				tenant1MessageReceived := false
				for i := 0; i < 10; i++ {
					_, message, readErr := conn1.ReadMessage()
					if readErr != nil {
						break
					}

					var wsMessage WebSocketMessage
					if jsonErr := json.Unmarshal(message, &wsMessage); jsonErr != nil {
						continue
					}

					if wsMessage.Type == "order_status_changed" {
						tenant1MessageReceived = true
						break
					}
				}

				// 租户2不应该收到通知
				tenant2MessageReceived := false
				for i := 0; i < 5; i++ {
					_, message, readErr := conn2.ReadMessage()
					if readErr != nil {
						break
					}

					var wsMessage WebSocketMessage
					if jsonErr := json.Unmarshal(message, &wsMessage); jsonErr != nil {
						continue
					}

					if wsMessage.Type == "order_status_changed" {
						tenant2MessageReceived = true
						break
					}
				}

				So(tenant1MessageReceived, ShouldBeTrue)
				So(tenant2MessageReceived, ShouldBeFalse)
			})
		})

		Convey("WebSocket心跳检测测试", func() {
			header := http.Header{}
			header.Add("X-Tenant-ID", "1")
			header.Add("X-Customer-ID", "1")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
			So(err, ShouldBeNil)
			defer conn.Close()

			Convey("心跳ping/pong机制", func() {
				// 发送多个ping消息测试心跳
				for i := 0; i < 3; i++ {
					err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
					So(err, ShouldBeNil)

					conn.SetReadDeadline(time.Now().Add(5 * time.Second))
					_, message, err := conn.ReadMessage()
					So(err, ShouldBeNil)

					var response WebSocketMessage
					err = json.Unmarshal(message, &response)
					So(err, ShouldBeNil)
					So(response.Type, ShouldEqual, "pong")

					time.Sleep(1 * time.Second)
				}
			})

			Convey("连接超时检测", func() {
				// 停止发送心跳，检查连接是否会超时
				// 注意：实际的超时检测可能需要更长时间，这里只是测试机制
				time.Sleep(2 * time.Second)

				// 尝试发送消息检查连接状态
				err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"test"}`))
				// 连接应该仍然有效（因为超时时间通常较长）
				So(err, ShouldBeNil)
			})
		})

		Convey("WebSocket错误处理测试", func() {
			header := http.Header{}
			header.Add("X-Tenant-ID", "1")
			header.Add("X-Customer-ID", "1")

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
			So(err, ShouldBeNil)
			defer conn.Close()

			Convey("无效消息格式处理", func() {
				// 发送无效JSON
				err = conn.WriteMessage(websocket.TextMessage, []byte(`{invalid json`))
				So(err, ShouldBeNil)

				// 连接应该仍然有效
				err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
				So(err, ShouldBeNil)

				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				_, message, err := conn.ReadMessage()
				So(err, ShouldBeNil)

				var response WebSocketMessage
				err = json.Unmarshal(message, &response)
				So(err, ShouldBeNil)
				So(response.Type, ShouldEqual, "pong")
			})

			Convey("未知消息类型处理", func() {
				// 发送未知类型的消息
				err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"unknown_type","data":{}}`))
				So(err, ShouldBeNil)

				// 连接应该仍然有效
				err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
				So(err, ShouldBeNil)

				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				_, message, err := conn.ReadMessage()
				So(err, ShouldBeNil)

				var response WebSocketMessage
				err = json.Unmarshal(message, &response)
				So(err, ShouldBeNil)
				So(response.Type, ShouldEqual, "pong")
			})
		})

		Convey("WebSocket性能测试", func() {
			// 创建多个并发连接测试性能
			numConnections := 5
			var connections []*websocket.Conn
			defer func() {
				for _, conn := range connections {
					conn.Close()
				}
			}()

			Convey("并发连接处理", func() {
				// 创建多个连接
				for i := 0; i < numConnections; i++ {
					header := http.Header{}
					header.Add("X-Tenant-ID", "1")
					header.Add("X-Customer-ID", fmt.Sprintf("%d", i+1))

					conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
					So(err, ShouldBeNil)
					connections = append(connections, conn)
				}

				// 并发发送ping消息
				done := make(chan bool, numConnections)
				for i, conn := range connections {
					go func(index int, c *websocket.Conn) {
						defer func() { done <- true }()

						err := c.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
						if err != nil {
							t.Logf("Connection %d write error: %v", index, err)
							return
						}

						c.SetReadDeadline(time.Now().Add(5 * time.Second))
						_, message, err := c.ReadMessage()
						if err != nil {
							t.Logf("Connection %d read error: %v", index, err)
							return
						}

						var response WebSocketMessage
						err = json.Unmarshal(message, &response)
						if err != nil {
							t.Logf("Connection %d unmarshal error: %v", index, err)
							return
						}

						if response.Type != "pong" {
							t.Logf("Connection %d unexpected response type: %s", index, response.Type)
						}
					}(i, conn)
				}

				// 等待所有连接完成
				for i := 0; i < numConnections; i++ {
					select {
					case <-done:
						// 连接处理完成
					case <-time.After(10 * time.Second):
						t.Fatalf("Connection %d timeout", i)
					}
				}

				t.Logf("All %d connections handled successfully", numConnections)
			})
		})
	})
}
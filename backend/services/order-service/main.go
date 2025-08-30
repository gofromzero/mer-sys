package main

import (
	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"

	// 导入MySQL驱动
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func main() {
	auth.NewJWTManager()
	ctx := gctx.GetInitCtx()

	s := g.Server()

	// 创建中间件实例
	authMiddleware := middleware.NewAuthMiddleware()

	// 创建WebSocket控制器（作为通知器使用）
	webSocketController := controller.NewWebSocketController()
	
	// 创建其他控制器
	orderController := controller.NewOrderController()
	cartController := controller.NewCartController()
	paymentController := controller.NewPaymentController()
	orderStatusController := controller.NewOrderStatusController()
	
	// 创建超时相关控制器
	orderStatusService := service.NewOrderStatusService()
	notificationService := service.NewNotificationService()
	orderTimeoutController := controller.NewOrderTimeoutController(orderStatusService, notificationService)
	orderTimeoutConfigController := controller.NewOrderTimeoutConfigController()
	
	// 为了简化实现，我们暂时注释掉WebSocket集成
	// 在生产环境中，应该通过依赖注入或服务发现来设置
	g.Log().Info(ctx, "订单服务控制器初始化完成")

	// 注册路由
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		// 购物车路由（需要认证）
		group.Group("/cart", func(cartGroup *ghttp.RouterGroup) {
			cartGroup.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)

			cartGroup.GET("/", cartController.GetCart)
			cartGroup.POST("/items", cartController.AddItem)
			cartGroup.PUT("/items/:item_id", cartController.UpdateItem)
			cartGroup.DELETE("/items/:item_id", cartController.RemoveItem)
			cartGroup.DELETE("/", cartController.ClearCart)
		})

		// 订单路由（需要认证）
		group.Group("/orders", func(orderGroup *ghttp.RouterGroup) {
			orderGroup.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)

			orderGroup.POST("/", orderController.CreateOrder)
			orderGroup.GET("/", orderController.ListOrders)
			orderGroup.GET("/:order_id", orderController.GetOrder)
			orderGroup.PUT("/:order_id/cancel", orderController.CancelOrder)

			// 高级查询功能
			orderGroup.GET("/query", orderController.QueryOrders)
			orderGroup.GET("/:order_id/detail", orderController.GetOrderWithHistory)
			orderGroup.GET("/search", orderController.SearchOrders)
			orderGroup.GET("/stats", orderController.GetOrderStats)

			// 订单状态管理路由
			orderGroup.PUT("/:order_id/status", orderStatusController.UpdateOrderStatus)
			orderGroup.GET("/:order_id/status-history", orderStatusController.GetOrderStatusHistory)
			orderGroup.GET("/:order_id/validate-status-transition", orderStatusController.ValidateStatusTransition)
			orderGroup.POST("/batch-update-status", orderStatusController.BatchUpdateOrderStatus)
			
			// 订单超时管理路由
			orderGroup.POST("/timeout/start", orderTimeoutController.StartTimeoutMonitor)
			orderGroup.POST("/timeout/stop", orderTimeoutController.StopTimeoutMonitor)
			orderGroup.GET("/timeout/statistics", orderTimeoutController.GetTimeoutStatistics)
			orderGroup.POST("/timeout/process", orderTimeoutController.ProcessTimeoutOrdersManually)
			
			// 订单超时配置路由
			orderGroup.POST("/timeout-configs", orderTimeoutConfigController.CreateTimeoutConfig)
			orderGroup.GET("/timeout-configs", orderTimeoutConfigController.ListTimeoutConfigs)
			orderGroup.GET("/timeout-configs/default", orderTimeoutConfigController.GetDefaultTimeoutConfig)
			orderGroup.GET("/timeout-configs/merchant/:merchant_id", orderTimeoutConfigController.GetTimeoutConfig)
			orderGroup.GET("/timeout-configs/effective/:merchant_id", orderTimeoutConfigController.GetEffectiveTimeoutConfig)
			orderGroup.PUT("/timeout-configs", orderTimeoutConfigController.UpdateTimeoutConfig)
			orderGroup.DELETE("/timeout-configs/:id", orderTimeoutConfigController.DeleteTimeoutConfig)

			// 支付相关路由
			orderGroup.POST("/:order_id/pay", paymentController.InitiatePayment)
			orderGroup.GET("/:order_id/payment-status", paymentController.GetPaymentStatus)
			orderGroup.POST("/:order_id/retry-payment", paymentController.RetryPayment)
		})

		// 支付回调路由（无需认证，但需要验证签名）
		group.Group("/payments", func(paymentGroup *ghttp.RouterGroup) {
			paymentGroup.POST("/callback/alipay", paymentController.AlipayCallback)
		})
		
		// WebSocket路由（需要认证）
		group.Group("/ws", func(wsGroup *ghttp.RouterGroup) {
			wsGroup.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)
			wsGroup.GET("/orders/status-updates", webSocketController.HandleOrderStatusUpdates)
		})
	})

	// 健康检查端点
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJsonExit(g.Map{
			"status":  "healthy",
			"service": "order-service",
		})
	})

	// 启动服务器
	g.Log().Info(ctx, "订单服务启动中...")
	s.SetPort(8084) // 订单服务端口
	s.Run()
}

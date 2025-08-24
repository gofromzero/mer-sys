package main

import (
	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/controller"
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

	// 创建控制器
	orderController := controller.NewOrderController()
	cartController := controller.NewCartController()
	paymentController := controller.NewPaymentController()

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

			// 支付相关路由
			orderGroup.POST("/:order_id/pay", paymentController.InitiatePayment)
			orderGroup.GET("/:order_id/payment-status", paymentController.GetPaymentStatus)
			orderGroup.POST("/:order_id/retry-payment", paymentController.RetryPayment)
		})

		// 支付回调路由（无需认证，但需要验证签名）
		group.Group("/payments", func(paymentGroup *ghttp.RouterGroup) {
			paymentGroup.POST("/callback/alipay", paymentController.AlipayCallback)
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

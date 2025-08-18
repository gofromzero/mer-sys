package main

import (
	"github.com/gofromzero/mer-sys/backend/services/user-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/shared/auth"
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

	// 创建认证控制器
	authController := controller.NewAuthController()

	// 注册路由
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		// 认证路由（公开，不需要认证）
		group.Group("/auth", func(authGroup *ghttp.RouterGroup) {
			authGroup.POST("/login", authController.Login)
			authGroup.POST("/logout", authController.Logout)
			authGroup.POST("/refresh", authController.RefreshToken)
		})

		// 需要认证的路由
		group.Group("/user", func(userGroup *ghttp.RouterGroup) {
			// TODO: 添加认证中间件
			// userGroup.Middleware(middleware.Auth)
			userGroup.GET("/info", authController.GetUserInfo)
		})
	})

	// 健康检查端点
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJsonExit(g.Map{
			"status":  "healthy",
			"service": "user-service",
		})
	})

	// 启动服务器
	g.Log().Info(ctx, "用户服务启动中...")
	s.SetPort(8081) // 用户服务端口
	s.Run()
}

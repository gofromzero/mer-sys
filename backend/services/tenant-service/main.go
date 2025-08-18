package main

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gofromzero/mer-sys/backend/services/tenant-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func main() {
	ctx := gctx.GetInitCtx()

	// 创建HTTP服务器
	s := g.Server()

	// 设置端口
	s.SetPort(8081)

	// 注册中间件
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		// 公开路由（不需要认证）
		group.Middleware(middleware.CORS())

		// 需要认证的路由
		group.Group("/", func(authGroup *ghttp.RouterGroup) {
			// TODO: 添加认证中间件
			// authGroup.Middleware(middleware.NewAuthMiddleware().RequirePermissions())

			// 租户相关路由
			tenantController := controller.NewTenantController()
			authGroup.POST("/tenants", tenantController.Create)
			authGroup.GET("/tenants", tenantController.List)
			authGroup.GET("/tenants/:id", tenantController.GetByID)
			authGroup.PUT("/tenants/:id", tenantController.Update)
			authGroup.PUT("/tenants/:id/status", tenantController.UpdateStatus)
			authGroup.GET("/tenants/:id/config", tenantController.GetConfig)
			authGroup.PUT("/tenants/:id/config", tenantController.UpdateConfig)
		})
	})

	// 健康检查路由
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status": "ok",
			"service": "tenant-service",
		})
	})

	// 启动服务器
	g.Log().Info(ctx, "Tenant service starting on port 8081...")
	s.Run()
}

func init() {
	ctx := context.Background()
	
	// 数据库配置
	g.Cfg().MustGet(ctx, "database")
}
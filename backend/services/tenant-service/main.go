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
			// 基础认证中间件
			authGroup.Middleware(middleware.NewAuthMiddleware())

			// 租户相关路由
			tenantController := controller.NewTenantController()
			
			// 创建租户 - 需要创建权限
			authGroup.POST("/tenants", 
				middleware.NewAuthMiddleware().RequirePermissions("tenant:create"),
				tenantController.Create)
			
			// 查看租户列表 - 需要查看权限
			authGroup.GET("/tenants", 
				middleware.NewAuthMiddleware().RequirePermissions("tenant:view"),
				tenantController.List)
			
			// 查看租户详情 - 需要查看权限
			authGroup.GET("/tenants/:id", 
				middleware.NewAuthMiddleware().RequirePermissions("tenant:view"),
				tenantController.GetByID)
			
			// 更新租户信息 - 需要更新权限
			authGroup.PUT("/tenants/:id", 
				middleware.NewAuthMiddleware().RequirePermissions("tenant:update"),
				tenantController.Update)
			
			// 更新租户状态 - 需要管理权限（敏感操作）
			authGroup.PUT("/tenants/:id/status", 
				middleware.NewAuthMiddleware().RequirePermissions("tenant:manage"),
				tenantController.UpdateStatus)
			
			// 查看租户配置 - 需要查看权限
			authGroup.GET("/tenants/:id/config", 
				middleware.NewAuthMiddleware().RequirePermissions("tenant:view"),
				tenantController.GetConfig)
			
			// 更新租户配置 - 需要管理权限（敏感操作）
			authGroup.PUT("/tenants/:id/config", 
				middleware.NewAuthMiddleware().RequirePermissions("tenant:manage"),
				tenantController.UpdateConfig)
			
			// 获取配置变更通知 - 需要查看权限
			authGroup.GET("/tenants/:id/config/notifications", 
				middleware.NewAuthMiddleware().RequirePermissions("tenant:view"),
				tenantController.GetConfigNotification)
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
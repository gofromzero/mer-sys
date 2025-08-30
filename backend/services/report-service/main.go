package main

import (
	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
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
	reportController := controller.NewReportController()
	templateController := controller.NewTemplateController()
	scheduledTaskController := controller.NewScheduledTaskController()

	g.Log().Info(ctx, "报表服务控制器初始化完成")

	// 启动调度服务
	schedulerService := service.NewSchedulerService()
	if err := schedulerService.Start(ctx); err != nil {
		g.Log().Error(ctx, "启动调度服务失败", "error", err)
	}

	// 注册路由
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		// 报表管理路由（需要认证）
		group.Group("/reports", func(reportGroup *ghttp.RouterGroup) {
			reportGroup.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)

			// 报表生成和管理
			reportGroup.POST("/generate", reportController.GenerateReport)
			reportGroup.GET("/", reportController.ListReports)
			reportGroup.GET("/:id", reportController.GetReport)
			reportGroup.DELETE("/:id", reportController.DeleteReport)
			reportGroup.GET("/:uuid/download", reportController.DownloadReport)
		})

		// 报表模板路由（需要认证）
		group.Group("/report-templates", func(templateGroup *ghttp.RouterGroup) {
			templateGroup.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)
			
			// 模板管理
			templateGroup.POST("/", templateController.CreateTemplate)
			templateGroup.GET("/", templateController.ListTemplates)
			templateGroup.GET("/:id", templateController.GetTemplate)
			templateGroup.PUT("/:id", templateController.UpdateTemplate)
			templateGroup.DELETE("/:id", templateController.DeleteTemplate)
			
			// 调度报表
			templateGroup.POST("/schedule", templateController.ScheduleReport)
		})

		// 数据分析路由（需要认证）
		group.Group("/analytics", func(analyticsGroup *ghttp.RouterGroup) {
			analyticsGroup.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)

			// 基础分析数据
			analyticsGroup.GET("/financial", reportController.GetFinancialAnalytics)
			analyticsGroup.GET("/merchants", reportController.GetMerchantAnalytics)
			analyticsGroup.GET("/customers", reportController.GetCustomerAnalytics)
			
			// 自定义查询和趋势数据
			analyticsGroup.POST("/custom", reportController.CustomQuery)
			analyticsGroup.GET("/trends/:metric", reportController.GetTrendData)
			
			// 缓存管理
			analyticsGroup.POST("/cache/clear", reportController.ClearCache)
		})

		// 定时任务路由
		scheduledTaskController.RegisterRoutes(group)
	})

	// 健康检查端点
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJsonExit(g.Map{
			"status":  "healthy",
			"service": "report-service",
		})
	})

	// 启动服务器
	g.Log().Info(ctx, "报表服务启动中...")
	s.SetPort(8085) // 报表服务端口
	s.Run()
}
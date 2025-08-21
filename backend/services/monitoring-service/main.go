package main

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"mer-demo/services/monitoring-service/internal/controller"
	"mer-demo/services/monitoring-service/internal/scheduler"
)

func main() {
	var (
		ctx = context.Background()
		cmd = &gcmd.Command{
			Name:  "monitoring-service",
			Usage: "monitoring-service",
			Brief: "权益监控微服务",
			Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
				s := g.Server()

				// 设置服务端口，默认为8085
				s.SetPort(g.Cfg().MustGet(ctx, "server.port", 8085).Int())

				// 注册路由
				registerRoutes(s)

				// 启动监控调度器
				monitoringScheduler := scheduler.NewScheduler(ctx)
				monitoringScheduler.Start()

				g.Log().Info(ctx, "监控服务启动完成", g.Map{
					"port":      s.GetListenedPort(),
					"scheduler": "started",
				})

				// 启动服务
				s.Run()
				return nil
			},
		}
	)

	cmd.Run(ctx)
}

// registerRoutes 注册路由
func registerRoutes(s *ghttp.Server) {
	// 创建控制器实例
	monitoringController := controller.NewMonitoringController()

	// 权益监控相关路由
	s.Group("/api/v1/monitoring", func(group *ghttp.RouterGroup) {
		// 统计和趋势
		group.GET("/rights/stats", monitoringController.GetRightsStats)
		group.GET("/rights/trends", monitoringController.GetRightsTrends)

		// 预警管理
		group.POST("/alerts/configure", monitoringController.ConfigureAlerts)
		group.GET("/alerts", monitoringController.ListAlerts)
		group.POST("/alerts/:id/resolve", monitoringController.ResolveAlert)

		// 仪表板和报告
		group.GET("/dashboard", monitoringController.GetDashboardData)
		group.POST("/reports/generate", monitoringController.GenerateReport)
	})

	// 健康检查
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"service": "monitoring-service",
			"port":    r.GetHost(),
		})
	})
}
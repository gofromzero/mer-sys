package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"mer-demo/services/fund-service/internal/controller"
	"mer-demo/shared/middleware"
)

func main() {
	// 创建HTTP服务器
	server := g.Server()
	
	// 配置CORS
	server.Use(middleware.CORS())
	
	// 配置认证中间件（如果存在）
	// server.Use(middleware.Auth)
	
	// 配置租户中间件（如果存在）
	// server.Use(middleware.Tenant)
	
	// 注册路由
	registerRoutes(server)
	
	// 启动服务
	server.SetPort(8084)
	server.Run()
}

// registerRoutes 注册路由
func registerRoutes(server *ghttp.Server) {
	// 创建控制器实例
	fundController := controller.NewFundController()
	
	// API v1 路由组
	v1 := server.Group("/api/v1")
	{
		// 资金管理路由
		funds := v1.Group("/funds")
		{
			// 充值相关 (需要充值权限)
			funds.POST("/deposit", middleware.RequireFundDeposit, fundController.Deposit)
			funds.POST("/batch-deposit", middleware.RequireFundDeposit, fundController.BatchDeposit)
			
			// 权益分配 (需要分配权限)
			funds.POST("/allocate", middleware.RequireFundAllocate, fundController.Allocate)
			
			// 余额查询 (需要查看权限)
			funds.GET("/balance/:merchant_id", middleware.RequireFundView, fundController.GetBalance)
			
			// 资金流转历史 (需要查看权限)
			funds.GET("/transactions", middleware.RequireFundView, fundController.ListTransactions)
			
			// 资金概览统计 (需要查看权限)
			funds.GET("/summary", middleware.RequireFundView, fundController.GetSummary)
			
			// 冻结/解冻权益 (需要冻结权限)
			funds.PUT("/freeze/:merchant_id", middleware.RequireFundFreeze, fundController.FreezeBalance)
		}
		
		// 健康检查
		v1.GET("/health", func(r *ghttp.Request) {
			r.Response.WriteJson(g.Map{
				"status":  "healthy",
				"service": "fund-service",
				"time":    r.Request.Header.Get("X-Request-Time"),
			})
		})
	}
}
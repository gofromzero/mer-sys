package main

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gofromzero/mer-sys/backend/services/merchant-service/internal/controller"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"
	"github.com/gofromzero/mer-sys/backend/shared/types"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func main() {
	ctx := gctx.GetInitCtx()

	// 创建HTTP服务器
	s := g.Server()

	// 设置端口
	s.SetPort(8082)

	// 注册中间件
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		// 公开路由（不需要认证）
		group.Middleware(middleware.CORS())

		// 需要认证的路由
		group.Group("/", func(authGroup *ghttp.RouterGroup) {
			// 基础认证中间件
			authGroup.Middleware(middleware.NewAuthMiddleware().JWTAuth)

			// 商户相关路由
			merchantController := controller.NewMerchantController()
			
			// 商户注册申请 - 需要创建权限
			authGroup.POST("/merchants", 
				middleware.NewAuthMiddleware().RequirePermissions(types.PermissionMerchantCreate),
				merchantController.Create)
			
			// 查看商户列表 - 需要查看权限
			authGroup.GET("/merchants", 
				middleware.NewAuthMiddleware().RequirePermissions(types.PermissionMerchantView),
				merchantController.List)
			
			// 查看商户详情 - 需要查看权限
			authGroup.GET("/merchants/:id", 
				middleware.NewAuthMiddleware().RequirePermissions(types.PermissionMerchantView),
				merchantController.GetByID)
			
			// 更新商户信息 - 需要更新权限
			authGroup.PUT("/merchants/:id", 
				middleware.NewAuthMiddleware().RequirePermissions(types.PermissionMerchantUpdate),
				merchantController.Update)
			
			// 更新商户状态 - 需要管理权限（敏感操作）
			authGroup.PUT("/merchants/:id/status", 
				middleware.NewAuthMiddleware().RequirePermissions(types.PermissionMerchantManage),
				merchantController.UpdateStatus)
			
			// 审批商户申请 - 需要管理权限（敏感操作）
			authGroup.POST("/merchants/:id/approve", 
				middleware.NewAuthMiddleware().RequirePermissions(types.PermissionMerchantManage),
				merchantController.Approve)
			
			// 拒绝商户申请 - 需要管理权限（敏感操作）
			authGroup.POST("/merchants/:id/reject", 
				middleware.NewAuthMiddleware().RequirePermissions(types.PermissionMerchantManage),
				merchantController.Reject)
			
			// 获取商户操作历史 - 需要查看权限
			authGroup.GET("/merchants/:id/audit-log", 
				middleware.NewAuthMiddleware().RequirePermissions(types.PermissionMerchantView),
				merchantController.GetAuditLog)
		})
	})

	// 健康检查路由
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status": "ok",
			"service": "merchant-service",
		})
	})

	// 启动服务器
	g.Log().Info(ctx, "Merchant service starting on port 8082...")
	s.Run()
}

func init() {
	ctx := context.Background()
	
	// 数据库配置
	g.Cfg().MustGet(ctx, "database")
}
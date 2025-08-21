package main

import (
	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/controller"
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
	productController := controller.NewProductController()
	categoryController := controller.NewCategoryController()

	// 注册路由
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		// 商品路由（需要认证和商户权限）
		group.Group("/products", func(productGroup *ghttp.RouterGroup) {
			// 添加认证和商户权限中间件
			productGroup.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)
			
			productGroup.POST("/", productController.CreateProduct)
			productGroup.GET("/", productController.ListProducts)
			productGroup.GET("/:id", productController.GetProduct)
			productGroup.PUT("/:id", productController.UpdateProduct)
			productGroup.DELETE("/:id", productController.DeleteProduct)
			productGroup.PATCH("/:id/status", productController.UpdateProductStatus)
			productGroup.POST("/:id/images", productController.UploadImage)
			productGroup.GET("/:id/history", productController.GetProductHistory)
			productGroup.POST("/batch", productController.BatchOperation)
		})

		// 分类路由（需要认证和商户权限）
		group.Group("/categories", func(categoryGroup *ghttp.RouterGroup) {
			// 添加认证和商户权限中间件
			categoryGroup.Middleware(authMiddleware.JWTAuth, authMiddleware.TenantIsolation)
			
			categoryGroup.POST("/", categoryController.CreateCategory)
			categoryGroup.GET("/", categoryController.GetCategoryList)
			categoryGroup.GET("/tree", categoryController.GetCategoryTree)
			categoryGroup.GET("/:id", categoryController.GetCategory)
			categoryGroup.PUT("/:id", categoryController.UpdateCategory)
			categoryGroup.DELETE("/:id", categoryController.DeleteCategory)
			categoryGroup.GET("/:id/children", categoryController.GetCategoryChildren)
			categoryGroup.GET("/:id/path", categoryController.GetCategoryPath)
		})
	})

	// 健康检查端点
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJsonExit(g.Map{
			"status":  "healthy",
			"service": "product-service",
		})
	})

	// 启动服务器
	g.Log().Info(ctx, "商品服务启动中...")
	s.SetPort(8083) // 商品服务端口
	s.Run()
}
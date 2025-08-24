package controller

import (
	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CartController 购物车控制器
type CartController struct {
	cartService service.ICartService
}

// NewCartController 创建购物车控制器实例
func NewCartController() *CartController {
	return &CartController{
		cartService: service.NewCartService(),
	}
}

// GetCart 获取购物车
func (c *CartController) GetCart(r *ghttp.Request) {
	customerID := r.GetCtxVar("user_id").Uint64()

	cart, err := c.cartService.GetCart(r.Context(), customerID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取购物车失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": cart,
	})
}

// AddItem 添加商品到购物车
func (c *CartController) AddItem(r *ghttp.Request) {
	var req types.AddCartItemRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	customerID := r.GetCtxVar("user_id").Uint64()

	err := c.cartService.AddItem(r.Context(), customerID, req.ProductID, req.Quantity)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "添加商品失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "添加成功",
	})
}

// UpdateItem 更新购物车商品数量
func (c *CartController) UpdateItem(r *ghttp.Request) {
	var req types.UpdateCartItemRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	itemID := r.Get("item_id").Uint64()
	if itemID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "购物车项ID不能为空",
		})
		return
	}

	err := c.cartService.UpdateItemQuantity(r.Context(), itemID, req.Quantity)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新商品数量失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "更新成功",
	})
}

// RemoveItem 从购物车中移除商品
func (c *CartController) RemoveItem(r *ghttp.Request) {
	itemID := r.Get("item_id").Uint64()
	if itemID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "购物车项ID不能为空",
		})
		return
	}

	err := c.cartService.RemoveItem(r.Context(), itemID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "删除商品失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "删除成功",
	})
}

// ClearCart 清空购物车
func (c *CartController) ClearCart(r *ghttp.Request) {
	customerID := r.GetCtxVar("user_id").Uint64()

	err := c.cartService.ClearCart(r.Context(), customerID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "清空购物车失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "清空成功",
	})
}

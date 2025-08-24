package controller

import (
	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

// OrderController 订单控制器
type OrderController struct {
	orderService service.IOrderService
}

// NewOrderController 创建订单控制器实例
func NewOrderController() *OrderController {
	return &OrderController{
		orderService: service.NewOrderService(),
	}
}

// CreateOrder 创建订单
func (c *OrderController) CreateOrder(r *ghttp.Request) {
	var req types.CreateOrderRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取当前用户ID作为客户ID
	customerID := r.GetCtxVar("user_id").Uint64()

	order, err := c.orderService.CreateOrder(r.Context(), customerID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建订单失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "订单创建成功",
		"data":    order,
	})
}

// GetOrder 获取订单详情
func (c *OrderController) GetOrder(r *ghttp.Request) {
	orderID := r.Get("order_id").Uint64()
	if orderID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	order, err := c.orderService.GetOrder(r.Context(), orderID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    404,
			"message": "订单不存在",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": order,
	})
}

// ListOrders 获取订单列表
func (c *OrderController) ListOrders(r *ghttp.Request) {
	customerID := r.GetCtxVar("user_id").Uint64()

	// 解析查询参数
	status := gconv.Int(r.Get("status"))
	page := gconv.Int(r.Get("page", 1))
	limit := gconv.Int(r.Get("limit", 20))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	orders, total, err := c.orderService.ListOrders(r.Context(), customerID, types.OrderStatus(status), page, limit)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取订单列表失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": g.Map{
			"items": orders,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// CancelOrder 取消订单
func (c *OrderController) CancelOrder(r *ghttp.Request) {
	orderID := r.Get("order_id").Uint64()
	if orderID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	err := c.orderService.CancelOrder(r.Context(), orderID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "取消订单失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "订单取消成功",
	})
}

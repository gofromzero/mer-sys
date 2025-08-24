package controller

import (
	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// PaymentController 支付控制器
type PaymentController struct {
	paymentService service.IPaymentService
}

// NewPaymentController 创建支付控制器实例
func NewPaymentController() *PaymentController {
	return &PaymentController{
		paymentService: service.NewPaymentService(),
	}
}

// InitiatePayment 发起支付
func (c *PaymentController) InitiatePayment(r *ghttp.Request) {
	orderID := r.Get("order_id").Uint64()
	if orderID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	var req types.InitiatePaymentRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	payment, err := c.paymentService.InitiatePayment(r.Context(), orderID, req.PaymentMethod, req.ReturnURL)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "发起支付失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "发起支付成功",
		"data":    payment,
	})
}

// GetPaymentStatus 查询支付状态
func (c *PaymentController) GetPaymentStatus(r *ghttp.Request) {
	orderID := r.Get("order_id").Uint64()
	if orderID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	status, err := c.paymentService.GetPaymentStatus(r.Context(), orderID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "查询支付状态失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": g.Map{
			"payment_status": status,
		},
	})
}

// RetryPayment 重新支付
func (c *PaymentController) RetryPayment(r *ghttp.Request) {
	orderID := r.Get("order_id").Uint64()
	if orderID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	var req types.InitiatePaymentRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	payment, err := c.paymentService.RetryPayment(r.Context(), orderID, req.PaymentMethod, req.ReturnURL)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "重新支付失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "重新支付发起成功",
		"data":    payment,
	})
}

// AlipayCallback 支付宝支付回调
func (c *PaymentController) AlipayCallback(r *ghttp.Request) {
	// 获取回调数据
	callbackData := make(map[string]interface{})
	for k, v := range r.GetRequestMap() {
		callbackData[k] = v
	}

	err := c.paymentService.HandleAlipayCallback(r.Context(), callbackData)
	if err != nil {
		g.Log().Error(r.Context(), "支付宝回调处理失败", "error", err.Error(), "data", callbackData)
		r.Response.WriteExit("failure")
		return
	}

	r.Response.WriteExit("success")
}

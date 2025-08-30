package controller

import (
	"fmt"
	"strconv"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/frame/g"
)

// OrderStatusController 订单状态管理控制器
type OrderStatusController struct {
	orderStatusService service.IOrderStatusService
}

// NewOrderStatusController 创建订单状态管理控制器实例
func NewOrderStatusController() *OrderStatusController {
	return &OrderStatusController{
		orderStatusService: service.NewOrderStatusService(),
	}
}

// UpdateOrderStatus 更新订单状态
// @Summary 更新订单状态
// @Description 更新指定订单的状态并记录变更历史
// @Tags 订单状态管理
// @Accept json
// @Produce json
// @Param order_id path int true "订单ID"
// @Param body body types.UpdateOrderStatusRequest true "更新订单状态请求"
// @Success 200 {object} utils.Response "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 403 {object} utils.Response "权限不足"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/orders/{order_id}/status [put]
func (c *OrderStatusController) UpdateOrderStatus(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 获取订单ID
	orderIDStr := r.Get("order_id").String()
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "订单ID格式错误")
		return
	}
	
	// 解析请求参数
	var req types.UpdateOrderStatusRequest
	if err := r.Parse(&req); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败: " + err.Error())
		return
	}
	
	// 验证请求参数
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		utils.ErrorResponse(r, 400, "参数验证失败: " + err.Error())
		return
	}
	
	// 更新订单状态
	if err := c.orderStatusService.UpdateOrderStatus(ctx, orderID, &req); err != nil {
		g.Log().Errorf(ctx, "更新订单状态失败: %v", err)
		utils.ErrorResponse(r, 500, err.Error())
		return
	}
	
	utils.SuccessResponse(r, nil)
}

// GetOrderStatusHistory 获取订单状态历史
// @Summary 获取订单状态历史
// @Description 获取指定订单的状态变更历史记录
// @Tags 订单状态管理
// @Accept json
// @Produce json
// @Param order_id path int true "订单ID"
// @Success 200 {object} utils.Response{data=[]types.OrderStatusHistory} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 403 {object} utils.Response "权限不足"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/orders/{order_id}/status-history [get]
func (c *OrderStatusController) GetOrderStatusHistory(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 获取订单ID
	orderIDStr := r.Get("order_id").String()
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "订单ID格式错误")
		return
	}
	
	// 获取状态历史
	history, err := c.orderStatusService.GetOrderStatusHistory(ctx, orderID)
	if err != nil {
		g.Log().Errorf(ctx, "获取订单状态历史失败: %v", err)
		utils.ErrorResponse(r, 500, err.Error())
		return
	}
	
	utils.SuccessResponse(r, history)
}

// BatchUpdateOrderStatus 批量更新订单状态
// @Summary 批量更新订单状态
// @Description 批量更新多个订单的状态，支持最多100个订单
// @Tags 订单状态管理
// @Accept json
// @Produce json
// @Param body body types.BatchUpdateOrderStatusRequest true "批量更新订单状态请求"
// @Success 200 {object} utils.Response{data=types.BatchUpdateOrderStatusResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 403 {object} utils.Response "权限不足"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/orders/batch-update-status [post]
func (c *OrderStatusController) BatchUpdateOrderStatus(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 解析请求参数
	var req types.BatchUpdateOrderStatusRequest
	if err := r.Parse(&req); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败: " + err.Error())
		return
	}
	
	// 验证请求参数
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		utils.ErrorResponse(r, 400, "参数验证失败: " + err.Error())
		return
	}
	
	// 批量更新订单状态
	response, err := c.orderStatusService.BatchUpdateOrderStatus(ctx, &req)
	if err != nil {
		g.Log().Errorf(ctx, "批量更新订单状态失败: %v", err)
		utils.ErrorResponse(r, 500, err.Error())
		return
	}
	
	// 构造响应消息
	message := fmt.Sprintf("批量更新完成: 成功 %d 个, 失败 %d 个", response.SuccessCount, response.FailCount)
	
	responseData := map[string]interface{}{
		"message": message,
		"result":  response,
	}
	utils.SuccessResponse(r, responseData)
}

// ValidateStatusTransition 验证订单状态转换
// @Summary 验证订单状态转换
// @Description 验证指定订单是否可以转换到目标状态
// @Tags 订单状态管理
// @Accept json
// @Produce json
// @Param order_id path int true "订单ID"
// @Param to_status query int true "目标状态"
// @Success 200 {object} utils.Response{data=bool} "验证结果"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 403 {object} utils.Response "权限不足"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/orders/{order_id}/validate-status-transition [get]
func (c *OrderStatusController) ValidateStatusTransition(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 获取订单ID
	orderIDStr := r.Get("order_id").String()
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "订单ID格式错误")
		return
	}
	
	// 获取目标状态
	toStatusStr := r.Get("to_status").String()
	toStatusInt, err := strconv.Atoi(toStatusStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "目标状态格式错误")
		return
	}
	toStatus := types.OrderStatusInt(toStatusInt)
	
	// 验证状态转换
	err = c.orderStatusService.ValidateStatusTransition(ctx, orderID, toStatus)
	if err != nil {
		// 返回验证失败的结果，而不是错误
		utils.SuccessResponse(r, map[string]interface{}{
			"valid":  false,
			"reason": err.Error(),
		})
		return
	}
	
	utils.SuccessResponse(r, map[string]interface{}{
		"valid":  true,
		"reason": "",
	})
}
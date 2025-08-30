package controller

import (
	"strconv"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// OrderTimeoutController 订单超时管理控制器
type OrderTimeoutController struct {
	timeoutService *service.OrderTimeoutService
}

// NewOrderTimeoutController 创建订单超时管理控制器实例
func NewOrderTimeoutController(orderStatusService service.IOrderStatusService, notificationService service.NotificationService) *OrderTimeoutController {
	return &OrderTimeoutController{
		timeoutService: service.NewOrderTimeoutService(orderStatusService, notificationService),
	}
}

// StartTimeoutMonitor 启动超时监控
// @Summary 启动超时监控
// @Description 启动订单超时监控定时任务
// @Tags 订单超时管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout/start [post]
func (c *OrderTimeoutController) StartTimeoutMonitor(r *ghttp.Request) {
	ctx := r.GetCtx()

	c.timeoutService.StartTimeoutMonitor(ctx)

	utils.SuccessResponse(r, g.Map{
		"message": "超时监控已启动",
	})
}

// StopTimeoutMonitor 停止超时监控
// @Summary 停止超时监控
// @Description 停止订单超时监控定时任务
// @Tags 订单超时管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout/stop [post]
func (c *OrderTimeoutController) StopTimeoutMonitor(r *ghttp.Request) {
	ctx := r.GetCtx()

	c.timeoutService.StopTimeoutMonitor(ctx)

	utils.SuccessResponse(r, g.Map{
		"message": "超时监控已停止",
	})
}

// GetTimeoutStatistics 获取超时统计信息
// @Summary 获取超时统计信息
// @Description 获取订单超时处理的统计信息
// @Tags 订单超时管理
// @Accept json
// @Produce json
// @Param merchant_id query uint64 false "商户ID"
// @Success 200 {object} utils.Response{data=types.OrderTimeoutStatistics}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout/statistics [get]
func (c *OrderTimeoutController) GetTimeoutStatistics(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 获取可选的商户ID参数
	var merchantID *uint64
	if merchantIDStr := r.Get("merchant_id").String(); merchantIDStr != "" {
		id, err := strconv.ParseUint(merchantIDStr, 10, 64)
		if err != nil {
			utils.ErrorResponse(r, 400, "无效的商户ID")
			return
		}
		merchantID = &id
	}

	// 获取统计信息
	statistics, err := c.timeoutService.GetTimeoutStatistics(ctx, merchantID)
	if err != nil {
		g.Log().Error(ctx, "获取超时统计信息失败", "error", err)
		utils.ErrorResponse(r, 500, "获取超时统计信息失败")
		return
	}

	utils.SuccessResponse(r, statistics)
}

// ProcessTimeoutOrdersManually 手动处理超时订单
// @Summary 手动处理超时订单
// @Description 手动触发超时订单处理（用于测试或紧急情况）
// @Tags 订单超时管理
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout/process [post]
func (c *OrderTimeoutController) ProcessTimeoutOrdersManually(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 创建一个临时的超时服务实例来处理
	orderStatusService := service.NewOrderStatusService()
	notificationService := service.NewNotificationService()
	tempTimeoutService := service.NewOrderTimeoutService(orderStatusService, notificationService)

	// 执行一次超时订单处理
	err := tempTimeoutService.ProcessTimeoutOrders(ctx)
	if err != nil {
		g.Log().Error(ctx, "手动处理超时订单失败", "error", err)
		utils.ErrorResponse(r, 500, "处理超时订单失败")
		return
	}

	utils.SuccessResponse(r, g.Map{
		"message": "超时订单处理完成",
	})
}
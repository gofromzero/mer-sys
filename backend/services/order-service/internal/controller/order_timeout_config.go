package controller

import (
	"fmt"
	"strconv"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// OrderTimeoutConfigController 订单超时配置控制器
type OrderTimeoutConfigController struct {
	timeoutConfigRepo *repository.OrderTimeoutConfigRepository
}

// NewOrderTimeoutConfigController 创建订单超时配置控制器实例
func NewOrderTimeoutConfigController() *OrderTimeoutConfigController {
	return &OrderTimeoutConfigController{
		timeoutConfigRepo: repository.NewOrderTimeoutConfigRepository(),
	}
}

// CreateTimeoutConfig 创建超时配置
// @Summary 创建超时配置
// @Description 创建订单超时配置
// @Tags 订单超时配置
// @Accept json
// @Produce json
// @Param config body types.OrderTimeoutConfig true "超时配置信息"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout-configs [post]
func (c *OrderTimeoutConfigController) CreateTimeoutConfig(r *ghttp.Request) {
	ctx := r.GetCtx()

	var config types.OrderTimeoutConfig
	if err := r.Parse(&config); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败")
		return
	}

	// 验证参数
	if err := c.validateTimeoutConfig(&config); err != nil {
		utils.ErrorResponse(r, 400, "参数验证失败")
		return
	}

	// 创建配置
	err := c.timeoutConfigRepo.Create(ctx, &config)
	if err != nil {
		g.Log().Error(ctx, "创建超时配置失败", "error", err)
		utils.ErrorResponse(r, 500, "创建超时配置失败")
		return
	}

	utils.SuccessResponse(r, config)
}

// GetTimeoutConfig 获取超时配置
// @Summary 获取超时配置
// @Description 根据商户ID获取超时配置
// @Tags 订单超时配置
// @Accept json
// @Produce json
// @Param merchant_id path uint64 true "商户ID"
// @Success 200 {object} utils.Response{data=types.OrderTimeoutConfig}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout-configs/merchant/{merchant_id} [get]
func (c *OrderTimeoutConfigController) GetTimeoutConfig(r *ghttp.Request) {
	ctx := r.GetCtx()

	merchantIDStr := r.Get("merchant_id").String()
	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "无效的商户ID")
		return
	}

	config, err := c.timeoutConfigRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		g.Log().Error(ctx, "获取超时配置失败", "error", err)
		utils.ErrorResponse(r, 500, "获取超时配置失败")
		return
	}

	if config == nil {
		utils.ErrorResponse(r, 404, "超时配置不存在")
		return
	}

	utils.SuccessResponse(r, config)
}

// GetEffectiveTimeoutConfig 获取有效的超时配置
// @Summary 获取有效的超时配置
// @Description 获取商户的有效超时配置（优先级：商户配置 > 默认配置）
// @Tags 订单超时配置
// @Accept json
// @Produce json
// @Param merchant_id path uint64 true "商户ID"
// @Success 200 {object} utils.Response{data=types.OrderTimeoutConfig}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout-configs/effective/{merchant_id} [get]
func (c *OrderTimeoutConfigController) GetEffectiveTimeoutConfig(r *ghttp.Request) {
	ctx := r.GetCtx()

	merchantIDStr := r.Get("merchant_id").String()
	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "无效的商户ID")
		return
	}

	config, err := c.timeoutConfigRepo.GetEffectiveConfig(ctx, merchantID)
	if err != nil {
		g.Log().Error(ctx, "获取有效超时配置失败", "error", err)
		utils.ErrorResponse(r, 500, "获取有效超时配置失败")
		return
	}

	utils.SuccessResponse(r, config)
}

// GetDefaultTimeoutConfig 获取默认超时配置
// @Summary 获取默认超时配置
// @Description 获取租户的默认超时配置
// @Tags 订单超时配置
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=types.OrderTimeoutConfig}
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout-configs/default [get]
func (c *OrderTimeoutConfigController) GetDefaultTimeoutConfig(r *ghttp.Request) {
	ctx := r.GetCtx()

	config, err := c.timeoutConfigRepo.GetDefaultConfig(ctx)
	if err != nil {
		g.Log().Error(ctx, "获取默认超时配置失败", "error", err)
		utils.ErrorResponse(r, 500, "获取默认超时配置失败")
		return
	}

	utils.SuccessResponse(r, config)
}

// UpdateTimeoutConfig 更新超时配置
// @Summary 更新超时配置
// @Description 更新订单超时配置
// @Tags 订单超时配置
// @Accept json
// @Produce json
// @Param config body types.OrderTimeoutConfig true "超时配置信息"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout-configs [put]
func (c *OrderTimeoutConfigController) UpdateTimeoutConfig(r *ghttp.Request) {
	ctx := r.GetCtx()

	var config types.OrderTimeoutConfig
	if err := r.Parse(&config); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败")
		return
	}

	// 验证参数
	if err := c.validateTimeoutConfig(&config); err != nil {
		utils.ErrorResponse(r, 400, "参数验证失败")
		return
	}

	// 更新配置
	err := c.timeoutConfigRepo.Update(ctx, &config)
	if err != nil {
		g.Log().Error(ctx, "更新超时配置失败", "error", err)
		utils.ErrorResponse(r, 500, "更新超时配置失败")
		return
	}

	utils.SuccessResponse(r, config)
}

// DeleteTimeoutConfig 删除超时配置
// @Summary 删除超时配置
// @Description 删除订单超时配置
// @Tags 订单超时配置
// @Accept json
// @Produce json
// @Param id path uint64 true "配置ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout-configs/{id} [delete]
func (c *OrderTimeoutConfigController) DeleteTimeoutConfig(r *ghttp.Request) {
	ctx := r.GetCtx()

	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "无效的配置ID")
		return
	}

	// 删除配置
	err = c.timeoutConfigRepo.Delete(ctx, id)
	if err != nil {
		g.Log().Error(ctx, "删除超时配置失败", "error", err)
		utils.ErrorResponse(r, 500, "删除超时配置失败")
		return
	}

	utils.SuccessResponse(r, g.Map{
		"message": "配置删除成功",
	})
}

// ListTimeoutConfigs 获取超时配置列表
// @Summary 获取超时配置列表
// @Description 获取租户的所有超时配置列表
// @Tags 订单超时配置
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response{data=[]types.OrderTimeoutConfig}
// @Failure 500 {object} utils.Response
// @Router /api/v1/orders/timeout-configs [get]
func (c *OrderTimeoutConfigController) ListTimeoutConfigs(r *ghttp.Request) {
	ctx := r.GetCtx()

	configs, err := c.timeoutConfigRepo.ListByTenant(ctx)
	if err != nil {
		g.Log().Error(ctx, "获取超时配置列表失败", "error", err)
		utils.ErrorResponse(r, 500, "获取超时配置列表失败")
		return
	}

	utils.SuccessResponse(r, configs)
}

// validateTimeoutConfig 验证超时配置参数
func (c *OrderTimeoutConfigController) validateTimeoutConfig(config *types.OrderTimeoutConfig) error {
	if config.PaymentTimeoutMinutes <= 0 {
		return fmt.Errorf("支付超时时间必须大于0分钟")
	}

	if config.PaymentTimeoutMinutes > 1440 { // 24小时
		return fmt.Errorf("支付超时时间不能超过1440分钟（24小时）")
	}

	if config.ProcessingTimeoutHours <= 0 {
		return fmt.Errorf("处理超时时间必须大于0小时")
	}

	if config.ProcessingTimeoutHours > 720 { // 30天
		return fmt.Errorf("处理超时时间不能超过720小时（30天）")
	}

	return nil
}
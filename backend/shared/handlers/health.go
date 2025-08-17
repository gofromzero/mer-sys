package handlers

import (
	"net/http"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/health"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	checker *health.HealthChecker
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(serviceName, version string) *HealthHandler {
	return &HealthHandler{
		checker: health.NewHealthChecker(serviceName, version),
	}
}

// Health 健康检查端点
// @Summary 系统健康检查
// @Description 检查系统整体健康状态，包括数据库、Redis等依赖项
// @Tags Health
// @Produce json
// @Success 200 {object} health.SystemHealth "系统健康"
// @Success 503 {object} health.SystemHealth "系统不健康"
// @Router /api/v1/health [get]
func (h *HealthHandler) Health(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 执行健康检查
	healthStatus := h.checker.CheckHealth(ctx)

	// 根据健康状态设置HTTP状态码
	statusCode := http.StatusOK
	if healthStatus.Status == health.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if healthStatus.Status == health.HealthStatusDegraded {
		statusCode = http.StatusPartialContent
	}

	// 记录健康检查日志
	if healthStatus.Status != health.HealthStatusHealthy {
		g.Log().Warningf(ctx, "系统健康检查异常: %s", healthStatus.Status)
		for name, component := range healthStatus.Components {
			if component.Status != health.HealthStatusHealthy {
				g.Log().Warningf(ctx, "组件 %s 状态异常: %s - %s", name, component.Status, component.Message)
			}
		}
	}

	r.Response.WriteStatus(statusCode)
	utils.SuccessResponse(r, healthStatus)
}

// Readiness 就绪检查端点
// @Summary 服务就绪检查
// @Description 检查服务是否准备好接收流量，要求所有依赖项都健康
// @Tags Health
// @Produce json
// @Success 200 {object} health.SystemHealth "服务就绪"
// @Success 503 {object} health.SystemHealth "服务未就绪"
// @Router /api/v1/health/ready [get]
func (h *HealthHandler) Readiness(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 执行就绪检查
	readiness := h.checker.GetReadiness(ctx)

	// 就绪检查更严格，只有所有组件都健康才返回200
	statusCode := http.StatusOK
	if readiness.Status != health.HealthStatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	r.Response.WriteStatus(statusCode)
	utils.SuccessResponse(r, readiness)
}

// Liveness 存活检查端点
// @Summary 服务存活检查
// @Description 检查服务是否还活着，只做基础检查
// @Tags Health
// @Produce json
// @Success 200 {object} health.SystemHealth "服务存活"
// @Success 503 {object} health.SystemHealth "服务死亡"
// @Router /api/v1/health/live [get]
func (h *HealthHandler) Liveness(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 执行存活检查
	liveness := h.checker.GetLiveness(ctx)

	statusCode := http.StatusOK
	if liveness.Status != health.HealthStatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	r.Response.WriteStatus(statusCode)
	utils.SuccessResponse(r, liveness)
}

// CheckComponent 检查特定组件
// @Summary 检查特定组件
// @Description 检查指定组件的健康状态
// @Tags Health
// @Param component path string true "组件名称" Enums(database,redis,system)
// @Produce json
// @Success 200 {object} health.ComponentHealth "组件健康"
// @Success 503 {object} health.ComponentHealth "组件不健康"
// @Router /api/v1/health/component/{component} [get]
func (h *HealthHandler) CheckComponent(r *ghttp.Request) {
	ctx := r.GetCtx()
	component := r.Get("component").String()

	if component == "" {
		utils.ErrorResponse(r, http.StatusBadRequest, "组件名称不能为空")
		return
	}

	// 检查特定组件
	componentHealth := h.checker.CheckDependency(ctx, component)

	statusCode := http.StatusOK
	if componentHealth.Status != health.HealthStatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	r.Response.WriteStatus(statusCode)
	utils.SuccessResponse(r, componentHealth)
}

// SimpleHealth 简单健康检查（快速响应）
// @Summary 简单健康检查
// @Description 快速健康检查，只返回基本状态
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{} "健康"
// @Router /api/v1/health/simple [get]
func (h *HealthHandler) SimpleHealth(r *ghttp.Request) {
	ctx := r.GetCtx()

	// 只检查服务是否能响应
	isHealthy := h.checker.IsHealthy(ctx)

	response := map[string]interface{}{
		"status":    "ok",
		"healthy":   isHealthy,
		"timestamp": time.Now(),
		"service":   h.checker.CheckHealth(ctx).Service,
	}

	statusCode := http.StatusOK
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
		response["status"] = "error"
	}

	r.Response.WriteStatus(statusCode)
	utils.SuccessResponse(r, response)
}

// RegisterRoutes 注册健康检查路由
func (h *HealthHandler) RegisterRoutes(group *ghttp.RouterGroup) {
	healthGroup := group.Group("/health")
	{
		healthGroup.GET("/", h.Health)
		healthGroup.GET("/ready", h.Readiness)
		healthGroup.GET("/live", h.Liveness)
		healthGroup.GET("/simple", h.SimpleHealth)
		healthGroup.GET("/component/:component", h.CheckComponent)
	}
}

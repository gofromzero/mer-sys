package controller

import (
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/service"
)

// InventoryAlertController 库存预警控制器
type InventoryAlertController struct {
	alertService     service.IInventoryAlertService
	inventoryService service.IInventoryService
}

// NewInventoryAlertController 创建库存预警控制器实例
func NewInventoryAlertController() *InventoryAlertController {
	inventoryService := service.NewInventoryService()
	return &InventoryAlertController{
		alertService:     service.NewInventoryAlertService(inventoryService),
		inventoryService: inventoryService,
	}
}

// CreateAlert 创建库存预警规则
// POST /api/v1/products/{id}/inventory/alerts
func (c *InventoryAlertController) CreateAlert(r *ghttp.Request) {
	// 权限检查暂时省略

	// 获取商品ID
	productIDStr := r.Get("id").String()
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 解析请求参数
	var req types.InventoryAlertRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 设置商品ID
	req.ProductID = productID

	// 创建预警规则
	alert, err := c.alertService.CreateAlert(r.GetCtx(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建预警规则失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "预警规则创建成功",
		"data":    alert,
	})
}

// GetProductAlerts 获取商品的预警规则
// GET /api/v1/products/{id}/inventory/alerts
func (c *InventoryAlertController) GetProductAlerts(r *ghttp.Request) {
	// 权限检查暂时省略

	// 获取商品ID
	productIDStr := r.Get("id").String()
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取预警规则
	alerts, err := c.alertService.GetAlertsByProduct(r.GetCtx(), productID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取预警规则失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取预警规则成功",
		"data": g.Map{
			"alerts": alerts,
			"total":  len(alerts),
		},
	})
}

// GetActiveAlerts 获取所有活跃的预警规则
// GET /api/v1/inventory/alerts/active
func (c *InventoryAlertController) GetActiveAlerts(r *ghttp.Request) {
	// 权限检查暂时省略

	// 获取所有活跃预警
	alerts, err := c.alertService.GetActiveAlerts(r.GetCtx())
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取活跃预警失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取活跃预警成功",
		"data": g.Map{
			"alerts": alerts,
			"total":  len(alerts),
		},
	})
}

// UpdateAlert 更新预警规则
// PUT /api/v1/inventory/alerts/{alert_id}
func (c *InventoryAlertController) UpdateAlert(r *ghttp.Request) {
	// 权限检查暂时省略

	// 获取预警规则ID
	alertIDStr := r.Get("alert_id").String()
	alertID, err := strconv.ParseUint(alertIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "预警规则ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 解析请求参数
	var req types.InventoryAlertRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 更新预警规则
	err = c.alertService.UpdateAlert(r.GetCtx(), alertID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新预警规则失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "预警规则更新成功",
	})
}

// DeleteAlert 删除预警规则
// DELETE /api/v1/inventory/alerts/{alert_id}
func (c *InventoryAlertController) DeleteAlert(r *ghttp.Request) {
	// 权限检查暂时省略

	// 获取预警规则ID
	alertIDStr := r.Get("alert_id").String()
	alertID, err := strconv.ParseUint(alertIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "预警规则ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 删除预警规则
	err = c.alertService.DeleteAlert(r.GetCtx(), alertID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "删除预警规则失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "预警规则删除成功",
	})
}

// ToggleAlert 切换预警规则状态
// POST /api/v1/inventory/alerts/{alert_id}/toggle
func (c *InventoryAlertController) ToggleAlert(r *ghttp.Request) {
	// 权限检查暂时省略

	// 获取预警规则ID
	alertIDStr := r.Get("alert_id").String()
	alertID, err := strconv.ParseUint(alertIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "预警规则ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 解析请求参数
	var toggleReq struct {
		IsActive bool `json:"is_active"`
	}
	if err := r.Parse(&toggleReq); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 切换预警规则状态
	err = c.alertService.ToggleAlert(r.GetCtx(), alertID, toggleReq.IsActive)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "切换预警状态失败",
			"error":   err.Error(),
		})
		return
	}

	statusText := "禁用"
	if toggleReq.IsActive {
		statusText = "启用"
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "预警规则" + statusText + "成功",
	})
}

// CheckProductAlerts 手动触发商品预警检查
// POST /api/v1/products/{id}/inventory/alerts/check
func (c *InventoryAlertController) CheckProductAlerts(r *ghttp.Request) {
	// 权限检查暂时省略

	// 获取商品ID
	productIDStr := r.Get("id").String()
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 检查预警
	err = c.alertService.CheckProductAlerts(r.GetCtx(), productID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "预警检查失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "预警检查完成",
	})
}

// CheckAllLowStockAlerts 检查所有低库存预警
// POST /api/v1/inventory/alerts/check-low-stock
func (c *InventoryAlertController) CheckAllLowStockAlerts(r *ghttp.Request) {
	// 权限检查暂时省略

	// 检查所有低库存预警
	err := c.alertService.CheckAllLowStockAlerts(r.GetCtx())
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "低库存预警检查失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "低库存预警检查完成",
	})
}

// GetInventoryMonitoring 获取库存监控数据
// GET /api/v1/inventory/monitoring
func (c *InventoryAlertController) GetInventoryMonitoring(r *ghttp.Request) {
	// 权限检查暂时省略

	// 获取活跃预警
	activeAlerts, err := c.alertService.GetActiveAlerts(r.GetCtx())
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取监控数据失败",
			"error":   err.Error(),
		})
		return
	}

	// 统计各种类型的预警数量
	alertStats := map[string]int{
		"total":      len(activeAlerts),
		"low_stock":  0,
		"out_stock":  0,
		"overstock":  0,
	}

	for _, alert := range activeAlerts {
		switch alert.AlertType {
		case types.InventoryAlertTypeLowStock:
			alertStats["low_stock"]++
		case types.InventoryAlertTypeOutOfStock:
			alertStats["out_stock"]++
		case types.InventoryAlertTypeOverstock:
			alertStats["overstock"]++
		}
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取监控数据成功",
		"data": g.Map{
			"alert_stats":    alertStats,
			"active_alerts":  activeAlerts,
			"last_updated":   time.Now(),
		},
	})
}
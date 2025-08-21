package controller

import (
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"mer-demo/services/monitoring-service/internal/service"
	"mer-demo/shared/middleware"
	"mer-demo/shared/types"
)

// MonitoringController 监控控制器
type MonitoringController struct {
	monitoringService service.MonitoringService
}

// NewMonitoringController 创建监控控制器实例
func NewMonitoringController() *MonitoringController {
	return &MonitoringController{
		monitoringService: service.NewMonitoringService(),
	}
}

// GetRightsStats 获取权益使用统计
func (c *MonitoringController) GetRightsStats(r *ghttp.Request) {
	// 权限检查
	if !middleware.CheckPermission(r.Context(), "rights:monitor") {
		r.Response.WriteStatusExit(403, "权限不足")
		return
	}

	var query types.RightsStatsQuery
	if err := r.Parse(&query); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	stats, err := c.monitoringService.GetRightsStats(r.Context(), &query)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取统计数据失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": stats,
	})
}

// GetRightsTrends 获取权益使用趋势
func (c *MonitoringController) GetRightsTrends(r *ghttp.Request) {
	// 权限检查
	if !middleware.CheckPermission(r.Context(), "rights:monitor") {
		r.Response.WriteStatusExit(403, "权限不足")
		return
	}

	var query types.RightsTrendsQuery
	if err := r.Parse(&query); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	trends, err := c.monitoringService.GetRightsTrends(r.Context(), &query)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取趋势数据失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": trends,
	})
}

// ConfigureAlerts 配置预警
func (c *MonitoringController) ConfigureAlerts(r *ghttp.Request) {
	// 权限检查
	if !middleware.CheckPermission(r.Context(), "admin:config") {
		r.Response.WriteStatusExit(403, "权限不足")
		return
	}

	var req types.AlertConfigureRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	if err := c.monitoringService.ConfigureAlerts(r.Context(), &req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "配置预警失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "预警配置成功",
	})
}

// ListAlerts 获取预警列表
func (c *MonitoringController) ListAlerts(r *ghttp.Request) {
	// 权限检查
	if !middleware.CheckPermission(r.Context(), "rights:alert") {
		r.Response.WriteStatusExit(403, "权限不足")
		return
	}

	var query types.AlertListQuery
	if err := r.Parse(&query); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	alerts, total, err := c.monitoringService.ListAlerts(r.Context(), &query)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取预警列表失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": g.Map{
			"list":      alerts,
			"total":     total,
			"page":      query.Page,
			"page_size": query.PageSize,
		},
	})
}

// ResolveAlert 解决预警
func (c *MonitoringController) ResolveAlert(r *ghttp.Request) {
	// 权限检查
	if !middleware.CheckPermission(r.Context(), "rights:alert") {
		r.Response.WriteStatusExit(403, "权限不足")
		return
	}

	alertIDStr := r.Get("id").String()
	alertID, err := strconv.ParseUint(alertIDStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的预警ID",
		})
		return
	}

	var req types.AlertResolveRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	if err := c.monitoringService.ResolveAlert(r.Context(), alertID, req.Resolution); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "解决预警失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "预警已解决",
	})
}

// GetDashboardData 获取监控仪表板数据
func (c *MonitoringController) GetDashboardData(r *ghttp.Request) {
	// 权限检查
	if !middleware.CheckPermission(r.Context(), "rights:monitor") {
		r.Response.WriteStatusExit(403, "权限不足")
		return
	}

	var merchantID *uint64
	merchantIDStr := r.Get("merchant_id").String()
	if merchantIDStr != "" {
		id, err := strconv.ParseUint(merchantIDStr, 10, 64)
		if err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    400,
				"message": "无效的商户ID",
			})
			return
		}
		merchantID = &id
	}

	data, err := c.monitoringService.GetDashboardData(r.Context(), merchantID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取仪表板数据失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": data,
	})
}

// GenerateReport 生成权益使用报告
func (c *MonitoringController) GenerateReport(r *ghttp.Request) {
	// 权限检查
	if !middleware.CheckPermission(r.Context(), "rights:report") {
		r.Response.WriteStatusExit(403, "权限不足")
		return
	}

	var req types.ReportGenerateRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	filename, err := c.monitoringService.GenerateReport(r.Context(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "生成报告失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": g.Map{
			"filename":    filename,
			"download_url": "/api/v1/reports/download/" + filename,
		},
		"message": "报告生成成功",
	})
}
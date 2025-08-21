package controller

import (
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/services/merchant-service/internal/service"
)

// DashboardController 仪表板控制器
type DashboardController struct {
	dashboardService service.DashboardService
}

// NewDashboardController 创建仪表板控制器实例
func NewDashboardController() *DashboardController {
	return &DashboardController{
		dashboardService: service.NewDashboardService(),
	}
}

// GetMerchantDashboard 获取商户仪表板核心数据
// GET /api/v1/merchant/dashboard
func (c *DashboardController) GetMerchantDashboard(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 获取仪表板数据
	dashboardData, err := c.dashboardService.GetMerchantDashboard(ctx, tenantID, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "获取商户仪表板数据失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取仪表板数据失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    dashboardData,
	})
}

// GetMerchantStats 获取指定时间段业务统计
// GET /api/v1/merchant/dashboard/stats/{period}
func (c *DashboardController) GetMerchantStats(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 获取时间周期参数
	periodStr := r.Get("period").String()
	var period types.TimePeriod
	switch periodStr {
	case "daily":
		period = types.TimePeriodDaily
	case "weekly":
		period = types.TimePeriodWeekly
	case "monthly":
		period = types.TimePeriodMonthly
	default:
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的时间周期参数，支持: daily, weekly, monthly",
		})
		return
	}
	
	// 获取统计数据
	stats, err := c.dashboardService.GetMerchantStats(ctx, tenantID, merchantID, period)
	if err != nil {
		g.Log().Errorf(ctx, "获取商户统计数据失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取统计数据失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// GetRightsUsageTrend 获取权益使用趋势数据
// GET /api/v1/merchant/dashboard/rights-trend?days=30
func (c *DashboardController) GetRightsUsageTrend(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 获取天数参数，默认30天
	days := r.Get("days", 30).Int()
	if days < 1 || days > 365 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "天数范围必须在 1-365 之间",
		})
		return
	}
	
	// 获取趋势数据
	trends, err := c.dashboardService.GetRightsUsageTrend(ctx, tenantID, merchantID, days)
	if err != nil {
		g.Log().Errorf(ctx, "获取权益趋势数据失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取权益趋势失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    trends,
	})
}

// GetPendingTasks 获取待处理事项汇总
// GET /api/v1/merchant/dashboard/pending-tasks
func (c *DashboardController) GetPendingTasks(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 获取待处理事项
	tasks, err := c.dashboardService.GetPendingTasks(ctx, tenantID, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "获取待处理事项失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取待处理事项失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    tasks,
	})
}

// GetNotifications 获取系统通知和公告
// GET /api/v1/merchant/dashboard/notifications
func (c *DashboardController) GetNotifications(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 获取通知和公告
	notifications, err := c.dashboardService.GetNotifications(ctx, tenantID, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "获取通知公告失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取通知公告失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    notifications,
	})
}

// GetDashboardConfig 获取仪表板个性化配置
// GET /api/v1/merchant/dashboard/config
func (c *DashboardController) GetDashboardConfig(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 获取仪表板配置
	config, err := c.dashboardService.GetDashboardConfig(ctx, tenantID, merchantID)
	if err != nil {
		g.Log().Errorf(ctx, "获取仪表板配置失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取仪表板配置失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取成功",
		"data":    config,
	})
}

// SaveDashboardConfig 保存仪表板个性化配置
// POST /api/v1/merchant/dashboard/config
func (c *DashboardController) SaveDashboardConfig(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 解析请求参数
	var request service.DashboardConfigRequest
	if err := r.Parse(&request); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"error":   err.Error(),
		})
		return
	}
	
	// 保存配置
	if err := c.dashboardService.SaveDashboardConfig(ctx, tenantID, merchantID, &request); err != nil {
		g.Log().Errorf(ctx, "保存仪表板配置失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "保存仪表板配置失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "保存成功",
	})
}

// UpdateDashboardConfig 更新仪表板布局配置
// PUT /api/v1/merchant/dashboard/config
func (c *DashboardController) UpdateDashboardConfig(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 解析请求参数
	var request service.DashboardConfigRequest
	if err := r.Parse(&request); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"error":   err.Error(),
		})
		return
	}
	
	// 更新配置
	if err := c.dashboardService.UpdateDashboardConfig(ctx, tenantID, merchantID, &request); err != nil {
		g.Log().Errorf(ctx, "更新仪表板配置失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新仪表板配置失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "更新成功",
	})
}

// MarkAnnouncementAsRead 标记公告为已读
// POST /api/v1/merchant/dashboard/announcements/{id}/read
func (c *DashboardController) MarkAnnouncementAsRead(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 从JWT中获取租户和商户信息
	tenantID, merchantID, err := c.extractMerchantInfo(r)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "身份验证失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 获取公告ID
	announcementIDStr := r.Get("id").String()
	announcementID, err := strconv.ParseUint(announcementIDStr, 10, 64)
	if err != nil || announcementID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的公告ID",
		})
		return
	}
	
	// 标记为已读
	if err := c.dashboardService.MarkAnnouncementAsRead(ctx, tenantID, merchantID, announcementID); err != nil {
		g.Log().Errorf(ctx, "标记公告已读失败: %v", err)
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "标记公告已读失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "标记成功",
	})
}

// 辅助方法

// extractMerchantInfo 从JWT中提取租户和商户信息
func (c *DashboardController) extractMerchantInfo(r *ghttp.Request) (tenantID, merchantID uint64, err error) {
	// 从JWT Claims中获取租户ID
	tenantIDValue := r.Get("tenant_id")
	if tenantIDValue.IsNil() {
		return 0, 0, gerror.New("未找到租户ID")
	}
	
	tenantID = tenantIDValue.Uint64()
	if tenantID == 0 {
		return 0, 0, gerror.New("无效的租户ID")
	}
	
	// 从JWT Claims中获取商户ID
	merchantIDValue := r.Get("merchant_id")
	if merchantIDValue.IsNil() {
		return 0, 0, gerror.New("未找到商户ID")
	}
	
	merchantID = merchantIDValue.Uint64()
	if merchantID == 0 {
		return 0, 0, gerror.New("无效的商户ID")
	}
	
	return tenantID, merchantID, nil
}

// validatePermission 验证仪表板访问权限
func (c *DashboardController) validatePermission(r *ghttp.Request, permission string) error {
	// 从JWT Claims中获取权限列表
	permissions := r.Get("permissions")
	if permissions.IsNil() {
		return gerror.New("无权限信息")
	}
	
	permissionList := permissions.Strings()
	
	// 检查是否有仪表板访问权限
	hasPermission := false
	for _, perm := range permissionList {
		if perm == permission || perm == "merchant:dashboard" || perm == "merchant:view" {
			hasPermission = true
			break
		}
	}
	
	if !hasPermission {
		return gerror.Newf("缺少权限: %s", permission)
	}
	
	return nil
}

// RegisterDashboardRoutes 注册仪表板路由
func RegisterDashboardRoutes(group *ghttp.RouterGroup) {
	controller := NewDashboardController()
	
	// 商户仪表板路由
	dashboardGroup := group.Group("/merchant/dashboard")
	
	// 需要merchant:dashboard权限的路由
	dashboardGroup.Middleware(func(r *ghttp.Request) {
		if err := controller.validatePermission(r, "merchant:dashboard"); err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    403,
				"message": "权限不足",
				"error":   err.Error(),
			})
			return
		}
		r.Middleware.Next()
	})
	
	// 注册路由
	dashboardGroup.GET("/", controller.GetMerchantDashboard)              // 获取仪表板核心数据
	dashboardGroup.GET("/stats/{period}", controller.GetMerchantStats)    // 获取统计数据
	dashboardGroup.GET("/rights-trend", controller.GetRightsUsageTrend)   // 获取权益趋势
	dashboardGroup.GET("/pending-tasks", controller.GetPendingTasks)      // 获取待处理事项
	dashboardGroup.GET("/notifications", controller.GetNotifications)     // 获取通知公告
	
	// 配置管理路由 (需要额外的配置权限)
	configGroup := dashboardGroup.Group("/config")
	configGroup.Middleware(func(r *ghttp.Request) {
		if err := controller.validatePermission(r, "merchant:config"); err != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    403,
				"message": "权限不足",
				"error":   err.Error(),
			})
			return
		}
		r.Middleware.Next()
	})
	
	configGroup.GET("/", controller.GetDashboardConfig)     // 获取配置
	configGroup.POST("/", controller.SaveDashboardConfig)  // 保存配置
	configGroup.PUT("/", controller.UpdateDashboardConfig) // 更新配置
	
	// 公告操作路由
	dashboardGroup.POST("/announcements/{id}/read", controller.MarkAnnouncementAsRead) // 标记公告已读
}
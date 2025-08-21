package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// DashboardRepository 仪表板数据访问接口
type DashboardRepository interface {
	// 获取商户仪表板数据
	GetMerchantDashboardData(ctx context.Context, tenantID, merchantID uint64, period types.TimePeriod) (*types.MerchantDashboardData, error)
	
	// 获取商户业务统计
	GetMerchantBusinessStats(ctx context.Context, tenantID, merchantID uint64, period types.TimePeriod) (*MerchantBusinessStats, error)
	
	// 获取权益使用历史趋势
	GetRightsUsageTrend(ctx context.Context, tenantID, merchantID uint64, days int) ([]types.RightsUsagePoint, error)
	
	// 获取待处理事项
	GetPendingTasks(ctx context.Context, tenantID, merchantID uint64) ([]types.PendingTask, error)
	
	// 获取系统通知和公告
	GetMerchantNotifications(ctx context.Context, tenantID, merchantID uint64, limit int) ([]types.Notification, []types.Announcement, error)
	
	// 仪表板配置管理
	GetDashboardConfig(ctx context.Context, tenantID, merchantID uint64) (*types.DashboardConfig, error)
	SaveDashboardConfig(ctx context.Context, tenantID, merchantID uint64, config *types.DashboardConfig) error
	UpdateDashboardConfig(ctx context.Context, tenantID, merchantID uint64, config *types.DashboardConfig) error
	
	// 公告阅读记录
	MarkAnnouncementAsRead(ctx context.Context, tenantID, merchantID, announcementID uint64) error
}

// dashboardRepositoryImpl 仪表板数据访问实现
type dashboardRepositoryImpl struct {
	*BaseRepository
	db gdb.DB
}

// NewDashboardRepository 创建仪表板数据访问实例
func NewDashboardRepository() DashboardRepository {
	return &dashboardRepositoryImpl{
		BaseRepository: NewBaseRepository("merchants"),
		db:             g.DB(),
	}
}

// MerchantBusinessStats 商户业务统计 (用于内部计算)
type MerchantBusinessStats struct {
	TotalSales     float64 `json:"total_sales"`
	TotalOrders    int     `json:"total_orders"`
	TotalCustomers int     `json:"total_customers"`
}

// GetMerchantDashboardData 获取商户仪表板数据
func (r *dashboardRepositoryImpl) GetMerchantDashboardData(ctx context.Context, tenantID, merchantID uint64, period types.TimePeriod) (*types.MerchantDashboardData, error) {
	// 获取商户权益余额
	rightsBalance, err := r.getMerchantRightsBalance(ctx, tenantID, merchantID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取权益余额失败")
	}

	// 获取业务统计数据
	businessStats, err := r.GetMerchantBusinessStats(ctx, tenantID, merchantID, period)
	if err != nil {
		return nil, gerror.Wrap(err, "获取业务统计失败")
	}

	// 获取权益使用趋势
	usageTrend, err := r.GetRightsUsageTrend(ctx, tenantID, merchantID, 30)
	if err != nil {
		return nil, gerror.Wrap(err, "获取权益趋势失败")
	}

	// 获取权益预警
	rightsAlerts, err := r.getRightsAlerts(ctx, tenantID, merchantID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取权益预警失败")
	}

	// 获取待处理事项
	pendingTasks, err := r.GetPendingTasks(ctx, tenantID, merchantID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取待处理事项失败")
	}

	// 获取通知和公告
	notifications, announcements, err := r.GetMerchantNotifications(ctx, tenantID, merchantID, 10)
	if err != nil {
		return nil, gerror.Wrap(err, "获取通知公告失败")
	}

	// 计算预测耗尽天数
	var predictedDepletionDays *int
	if rightsBalance != nil && rightsBalance.AvailableBalance > 0 {
		if avgDailyUsage := r.calculateAvgDailyUsage(usageTrend); avgDailyUsage > 0 {
			days := int(rightsBalance.AvailableBalance / avgDailyUsage)
			predictedDepletionDays = &days
		}
	}

	return &types.MerchantDashboardData{
		MerchantID:             merchantID,
		TenantID:               tenantID,
		Period:                 period,
		TotalSales:             businessStats.TotalSales,
		TotalOrders:            businessStats.TotalOrders,
		TotalCustomers:         businessStats.TotalCustomers,
		RightsBalance:          rightsBalance,
		RightsUsageTrend:       usageTrend,
		RightsAlerts:           rightsAlerts,
		PredictedDepletionDays: predictedDepletionDays,
		PendingOrders:          r.countPendingOrders(ctx, tenantID, merchantID),
		PendingVerifications:   r.countPendingVerifications(ctx, tenantID, merchantID),
		PendingTasks:           pendingTasks,
		Announcements:          announcements,
		Notifications:          notifications,
		LastUpdated:            time.Now(),
	}, nil
}

// GetMerchantBusinessStats 获取商户业务统计
func (r *dashboardRepositoryImpl) GetMerchantBusinessStats(ctx context.Context, tenantID, merchantID uint64, period types.TimePeriod) (*MerchantBusinessStats, error) {
	var startTime time.Time
	now := time.Now()

	switch period {
	case types.TimePeriodDaily:
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case types.TimePeriodWeekly:
		startTime = now.AddDate(0, 0, -7)
	case types.TimePeriodMonthly:
		startTime = now.AddDate(0, -1, 0)
	default:
		startTime = now.AddDate(0, 0, -7) // 默认一周
	}

	// 查询订单统计 - 自动注入租户和商户过滤
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_sales,
			COUNT(*) as total_orders,
			COUNT(DISTINCT customer_id) as total_customers
		FROM orders 
		WHERE tenant_id = ? AND merchant_id = ? 
		AND status IN ('paid', 'processing', 'completed')
		AND created_at >= ?
	`

	stats := &MerchantBusinessStats{}
	err := r.db.Ctx(ctx).Raw(query, tenantID, merchantID, startTime).Scan(stats)
	if err != nil {
		return nil, gerror.Wrap(err, "查询业务统计失败")
	}
	// Scan 已经自动填充了 stats 结构体

	return stats, nil
}

// GetRightsUsageTrend 获取权益使用历史趋势
func (r *dashboardRepositoryImpl) GetRightsUsageTrend(ctx context.Context, tenantID, merchantID uint64, days int) ([]types.RightsUsagePoint, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	
	// 查询权益使用统计 - 自动注入租户和商户过滤
	query := `
		SELECT 
			stat_date as date,
			total_allocated as balance,
			total_consumed as usage,
			usage_trend as trend
		FROM rights_usage_stats 
		WHERE tenant_id = ? AND merchant_id = ? 
		AND stat_date >= ?
		ORDER BY stat_date ASC
	`

	results, err := r.db.GetAll(ctx, query, tenantID, merchantID, startDate)
	if err != nil {
		return nil, gerror.Wrap(err, "查询权益趋势失败")
	}

	var trends []types.RightsUsagePoint
	for _, row := range results {
		trend := types.RightsUsagePoint{
			Date:    row["date"].Time(),
			Balance: row["balance"].Float64(),
			Usage:   row["usage"].Float64(),
			Trend:   types.TrendDirection(row["trend"].Int()),
		}
		trends = append(trends, trend)
	}

	return trends, nil
}

// GetPendingTasks 获取待处理事项
func (r *dashboardRepositoryImpl) GetPendingTasks(ctx context.Context, tenantID, merchantID uint64) ([]types.PendingTask, error) {
	var tasks []types.PendingTask

	// 待处理订单
	pendingOrders := r.countPendingOrders(ctx, tenantID, merchantID)
	if pendingOrders > 0 {
		tasks = append(tasks, types.PendingTask{
			ID:          "pending_orders",
			Type:        types.TaskTypeOrderProcessing,
			Description: fmt.Sprintf("有 %d 个订单待处理", pendingOrders),
			Priority:    types.PriorityNormal,
			Count:       pendingOrders,
		})
	}

	// 待核销订单
	pendingVerifications := r.countPendingVerifications(ctx, tenantID, merchantID)
	if pendingVerifications > 0 {
		tasks = append(tasks, types.PendingTask{
			ID:          "pending_verifications",
			Type:        types.TaskTypeVerificationPending,
			Description: fmt.Sprintf("有 %d 个订单待核销", pendingVerifications),
			Priority:    types.PriorityHigh,
			Count:       pendingVerifications,
		})
	}

	// 权益余额预警
	rightsBalance, _ := r.getMerchantRightsBalance(ctx, tenantID, merchantID)
	if rightsBalance != nil && rightsBalance.WarningThreshold != nil && 
	   rightsBalance.AvailableBalance < *rightsBalance.WarningThreshold {
		priority := types.PriorityHigh
		if rightsBalance.CriticalThreshold != nil && 
		   rightsBalance.AvailableBalance < *rightsBalance.CriticalThreshold {
			priority = types.PriorityUrgent
		}
		
		tasks = append(tasks, types.PendingTask{
			ID:          "low_balance_warning",
			Type:        types.TaskTypeLowBalanceWarning,
			Description: fmt.Sprintf("权益余额不足，当前可用: %.2f", rightsBalance.AvailableBalance),
			Priority:    priority,
			Count:       1,
		})
	}

	return tasks, nil
}

// GetMerchantNotifications 获取系统通知和公告
func (r *dashboardRepositoryImpl) GetMerchantNotifications(ctx context.Context, tenantID, merchantID uint64, limit int) ([]types.Notification, []types.Announcement, error) {
	// 获取通知 (这里假设有notifications表)
	var notifications []types.Notification
	
	// 获取公告
	announcementQuery := `
		SELECT 
			a.id, a.title, a.content, a.priority, a.publish_date, a.expire_date,
			CASE WHEN ar.read_at IS NOT NULL THEN true ELSE false END as read_status
		FROM system_announcements a
		LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id AND ar.merchant_id = ?
		WHERE a.tenant_id = ?
		AND (a.target_merchants IS NULL OR JSON_CONTAINS(a.target_merchants, ?))
		AND (a.expire_date IS NULL OR a.expire_date > NOW())
		AND a.publish_date <= NOW()
		ORDER BY a.priority DESC, a.publish_date DESC
		LIMIT ?
	`

	results, err := r.db.GetAll(ctx, announcementQuery, merchantID, tenantID, fmt.Sprintf(`"%d"`, merchantID), limit)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "查询公告失败")
	}

	var announcements []types.Announcement
	for _, row := range results {
		announcement := types.Announcement{
			ID:          row["id"].Uint64(),
			Title:       row["title"].String(),
			Content:     row["content"].String(),
			Priority:    types.Priority(row["priority"].String()),
			PublishDate: row["publish_date"].Time(),
			ReadStatus:  row["read_status"].Bool(),
		}
		
		if !row["expire_date"].IsNil() {
			expireDate := row["expire_date"].Time()
			announcement.ExpireDate = &expireDate
		}
		
		announcements = append(announcements, announcement)
	}

	return notifications, announcements, nil
}

// GetDashboardConfig 获取仪表板配置
func (r *dashboardRepositoryImpl) GetDashboardConfig(ctx context.Context, tenantID, merchantID uint64) (*types.DashboardConfig, error) {
	query := `
		SELECT layout_config, widget_preferences, refresh_interval, mobile_layout
		FROM merchant_dashboard_configs 
		WHERE tenant_id = ? AND merchant_id = ?
	`

	result, err := r.db.GetOne(ctx, query, tenantID, merchantID)
	if err != nil {
		if err == sql.ErrNoRows {
			// 返回默认配置
			return r.getDefaultDashboardConfig(merchantID), nil
		}
		return nil, gerror.Wrap(err, "查询仪表板配置失败")
	}

	config := &types.DashboardConfig{
		MerchantID:      merchantID,
		RefreshInterval: result["refresh_interval"].Int(),
	}

	// 解析 JSON 字段
	if !result["layout_config"].IsNil() {
		var layoutConfig types.LayoutConfig
		if err := json.Unmarshal([]byte(result["layout_config"].String()), &layoutConfig); err == nil {
			config.LayoutConfig = &layoutConfig
		}
	}

	if !result["widget_preferences"].IsNil() {
		var preferences []types.WidgetPreference
		if err := json.Unmarshal([]byte(result["widget_preferences"].String()), &preferences); err == nil {
			config.WidgetPreferences = preferences
		}
	}

	if !result["mobile_layout"].IsNil() {
		var mobileLayout types.MobileLayoutConfig
		if err := json.Unmarshal([]byte(result["mobile_layout"].String()), &mobileLayout); err == nil {
			config.MobileLayout = &mobileLayout
		}
	}

	return config, nil
}

// SaveDashboardConfig 保存仪表板配置
func (r *dashboardRepositoryImpl) SaveDashboardConfig(ctx context.Context, tenantID, merchantID uint64, config *types.DashboardConfig) error {
	layoutConfigJSON, _ := json.Marshal(config.LayoutConfig)
	preferencesJSON, _ := json.Marshal(config.WidgetPreferences)
	mobileLayoutJSON, _ := json.Marshal(config.MobileLayout)

	query := `
		INSERT INTO merchant_dashboard_configs 
		(tenant_id, merchant_id, layout_config, widget_preferences, refresh_interval, mobile_layout)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		layout_config = VALUES(layout_config),
		widget_preferences = VALUES(widget_preferences),
		refresh_interval = VALUES(refresh_interval),
		mobile_layout = VALUES(mobile_layout),
		updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, tenantID, merchantID, string(layoutConfigJSON), 
		string(preferencesJSON), config.RefreshInterval, string(mobileLayoutJSON))
	if err != nil {
		return gerror.Wrap(err, "保存仪表板配置失败")
	}

	return nil
}

// UpdateDashboardConfig 更新仪表板配置
func (r *dashboardRepositoryImpl) UpdateDashboardConfig(ctx context.Context, tenantID, merchantID uint64, config *types.DashboardConfig) error {
	return r.SaveDashboardConfig(ctx, tenantID, merchantID, config)
}

// MarkAnnouncementAsRead 标记公告为已读
func (r *dashboardRepositoryImpl) MarkAnnouncementAsRead(ctx context.Context, tenantID, merchantID, announcementID uint64) error {
	query := `
		INSERT INTO announcement_reads (tenant_id, announcement_id, merchant_id)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE read_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, tenantID, announcementID, merchantID)
	if err != nil {
		return gerror.Wrap(err, "标记公告已读失败")
	}

	return nil
}

// 辅助方法

// getMerchantRightsBalance 获取商户权益余额
func (r *dashboardRepositoryImpl) getMerchantRightsBalance(ctx context.Context, tenantID, merchantID uint64) (*types.RightsBalance, error) {
	query := `
		SELECT total_balance, used_balance, frozen_balance, available_balance,
			   last_updated, warning_threshold, critical_threshold, trend_coefficient
		FROM merchants 
		WHERE tenant_id = ? AND id = ?
	`

	result, err := r.db.GetOne(ctx, query, tenantID, merchantID)
	if err != nil {
		return nil, gerror.Wrap(err, "查询权益余额失败")
	}

	balance := &types.RightsBalance{
		TotalBalance:      result["total_balance"].Float64(),
		UsedBalance:       result["used_balance"].Float64(),
		FrozenBalance:     result["frozen_balance"].Float64(),
		AvailableBalance:  result["available_balance"].Float64(),
		LastUpdated:       result["last_updated"].Time(),
	}

	if !result["warning_threshold"].IsNil() {
		threshold := result["warning_threshold"].Float64()
		balance.WarningThreshold = &threshold
	}
	if !result["critical_threshold"].IsNil() {
		threshold := result["critical_threshold"].Float64()
		balance.CriticalThreshold = &threshold
	}
	if !result["trend_coefficient"].IsNil() {
		coefficient := result["trend_coefficient"].Float64()
		balance.TrendCoefficient = &coefficient
	}

	return balance, nil
}

// getRightsAlerts 获取权益预警
func (r *dashboardRepositoryImpl) getRightsAlerts(ctx context.Context, tenantID, merchantID uint64) ([]types.RightsAlert, error) {
	query := `
		SELECT id, alert_type, threshold_value, current_value, severity, status,
			   message, triggered_at, resolved_at, notified_channels
		FROM rights_alerts 
		WHERE tenant_id = ? AND merchant_id = ? AND status = ?
		ORDER BY severity DESC, triggered_at DESC
		LIMIT 5
	`

	results, err := r.db.GetAll(ctx, query, tenantID, merchantID, types.AlertStatusActive)
	if err != nil {
		return nil, gerror.Wrap(err, "查询权益预警失败")
	}

	var alerts []types.RightsAlert
	for _, row := range results {
		alert := types.RightsAlert{
			ID:             row["id"].Uint64(),
			TenantID:       tenantID,
			MerchantID:     merchantID,
			AlertType:      types.AlertType(row["alert_type"].Int()),
			ThresholdValue: row["threshold_value"].Float64(),
			CurrentValue:   row["current_value"].Float64(),
			Severity:       types.AlertSeverity(row["severity"].Int()),
			Status:         types.AlertStatus(row["status"].Int()),
			Message:        row["message"].String(),
			TriggeredAt:    row["triggered_at"].Time(),
		}

		if !row["resolved_at"].IsNil() {
			resolvedAt := row["resolved_at"].Time()
			alert.ResolvedAt = &resolvedAt
		}

		if !row["notified_channels"].IsNil() {
			var channels []string
			json.Unmarshal([]byte(row["notified_channels"].String()), &channels)
			alert.NotifiedChannels = channels
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// countPendingOrders 统计待处理订单数量
func (r *dashboardRepositoryImpl) countPendingOrders(ctx context.Context, tenantID, merchantID uint64) int {
	query := `
		SELECT COUNT(*) as count
		FROM orders 
		WHERE tenant_id = ? AND merchant_id = ? 
		AND status IN ('paid', 'processing')
	`

	result, err := r.db.GetValue(ctx, query, tenantID, merchantID)
	if err != nil {
		return 0
	}

	return result.Int()
}

// countPendingVerifications 统计待核销订单数量
func (r *dashboardRepositoryImpl) countPendingVerifications(ctx context.Context, tenantID, merchantID uint64) int {
	query := `
		SELECT COUNT(*) as count
		FROM orders 
		WHERE tenant_id = ? AND merchant_id = ? 
		AND status = 'paid' 
		AND (verification_info IS NULL OR JSON_EXTRACT(verification_info, '$.verified_at') IS NULL)
	`

	result, err := r.db.GetValue(ctx, query, tenantID, merchantID)
	if err != nil {
		return 0
	}

	return result.Int()
}

// calculateAvgDailyUsage 计算平均日使用量
func (r *dashboardRepositoryImpl) calculateAvgDailyUsage(trends []types.RightsUsagePoint) float64 {
	if len(trends) == 0 {
		return 0
	}

	totalUsage := 0.0
	for _, trend := range trends {
		totalUsage += trend.Usage
	}

	return totalUsage / float64(len(trends))
}

// getDefaultDashboardConfig 获取默认仪表板配置
func (r *dashboardRepositoryImpl) getDefaultDashboardConfig(merchantID uint64) *types.DashboardConfig {
	return &types.DashboardConfig{
		MerchantID: merchantID,
		LayoutConfig: &types.LayoutConfig{
			Columns: 4,
			Widgets: []types.DashboardWidget{
				{ID: "sales_overview", Type: types.WidgetTypeSalesOverview, Position: types.Position{X: 0, Y: 0}, Size: types.Size{Width: 2, Height: 1}, Visible: true},
				{ID: "rights_balance", Type: types.WidgetTypeRightsBalance, Position: types.Position{X: 2, Y: 0}, Size: types.Size{Width: 2, Height: 1}, Visible: true},
				{ID: "rights_trend", Type: types.WidgetTypeRightsTrend, Position: types.Position{X: 0, Y: 1}, Size: types.Size{Width: 4, Height: 2}, Visible: true},
				{ID: "pending_tasks", Type: types.WidgetTypePendingTasks, Position: types.Position{X: 0, Y: 3}, Size: types.Size{Width: 2, Height: 2}, Visible: true},
				{ID: "announcements", Type: types.WidgetTypeAnnouncements, Position: types.Position{X: 2, Y: 3}, Size: types.Size{Width: 2, Height: 2}, Visible: true},
			},
		},
		WidgetPreferences: []types.WidgetPreference{
			{WidgetType: types.WidgetTypeSalesOverview, Enabled: true},
			{WidgetType: types.WidgetTypeRightsBalance, Enabled: true},
			{WidgetType: types.WidgetTypeRightsTrend, Enabled: true},
			{WidgetType: types.WidgetTypePendingTasks, Enabled: true},
			{WidgetType: types.WidgetTypeAnnouncements, Enabled: true},
		},
		RefreshInterval: 300, // 5分钟
		MobileLayout: &types.MobileLayoutConfig{
			Columns: 1,
			Widgets: []types.DashboardWidget{
				{ID: "sales_overview", Type: types.WidgetTypeSalesOverview, Position: types.Position{X: 0, Y: 0}, Size: types.Size{Width: 1, Height: 1}, Visible: true},
				{ID: "rights_balance", Type: types.WidgetTypeRightsBalance, Position: types.Position{X: 0, Y: 1}, Size: types.Size{Width: 1, Height: 1}, Visible: true},
				{ID: "pending_tasks", Type: types.WidgetTypePendingTasks, Position: types.Position{X: 0, Y: 2}, Size: types.Size{Width: 1, Height: 2}, Visible: true},
			},
		},
	}
}
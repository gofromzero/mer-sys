package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"mer-demo/shared/types"
)

// MonitoringRepository 监控数据访问接口
type MonitoringRepository interface {
	// Alert operations
	CreateAlert(ctx context.Context, alert *types.RightsAlert) error
	GetAlert(ctx context.Context, id uint64) (*types.RightsAlert, error)
	ListAlerts(ctx context.Context, query *types.AlertListQuery) ([]*types.RightsAlert, int, error)
	UpdateAlert(ctx context.Context, alert *types.RightsAlert) error
	ResolveAlert(ctx context.Context, id uint64, resolution string) error
	GetActiveAlertsCount(ctx context.Context, merchantID *uint64) (int, error)
	GetCriticalAlertsCount(ctx context.Context, merchantID *uint64) (int, error)

	// Usage stats operations
	CreateUsageStats(ctx context.Context, stats *types.RightsUsageStats) error
	GetUsageStats(ctx context.Context, query *types.RightsStatsQuery) ([]*types.RightsUsageStats, error)
	GetUsageTrends(ctx context.Context, query *types.RightsTrendsQuery) ([]*types.RightsUsageStats, error)
	GetMerchantUsageInfo(ctx context.Context, limit int) ([]*types.MerchantUsageInfo, error)
	GetDailyUsageTrends(ctx context.Context, days int, merchantID *uint64) ([]*types.DailyUsageTrend, error)

	// Dashboard data
	GetDashboardData(ctx context.Context, merchantID *uint64) (*types.MonitoringDashboardData, error)

	// Configuration operations
	UpdateMerchantThresholds(ctx context.Context, merchantID uint64, warningThreshold, criticalThreshold *float64) error
	GetMerchantThresholds(ctx context.Context, merchantID uint64) (warning, critical *float64, err error)
}

// monitoringRepository 监控数据访问实现
type monitoringRepository struct {
	*BaseRepository
}

// NewMonitoringRepository 创建监控数据访问实例
func NewMonitoringRepository() MonitoringRepository {
	return &monitoringRepository{
		BaseRepository: NewBaseRepository("rights_alerts"),
	}
}

// CreateAlert 创建预警记录
func (r *monitoringRepository) CreateAlert(ctx context.Context, alert *types.RightsAlert) error {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return fmt.Errorf("获取租户ID失败: %w", err)
	}
	alert.TenantID = tenantID
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	_, err := g.DB().Model("rights_alerts").Ctx(ctx).Insert(alert)
	return err
}

// GetAlert 获取预警记录
func (r *monitoringRepository) GetAlert(ctx context.Context, id uint64) (*types.RightsAlert, error) {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取租户ID失败: %w", err)
	}

	var alert types.RightsAlert
	err := g.DB().Model("rights_alerts").
		Ctx(ctx).
		Where("id", id).
		Where("tenant_id", tenantID).
		Scan(&alert)

	if err != nil {
		return nil, err
	}
	if alert.ID == 0 {
		return nil, types.ErrAlertNotFound
	}
	return &alert, nil
}

// ListAlerts 获取预警记录列表
func (r *monitoringRepository) ListAlerts(ctx context.Context, query *types.AlertListQuery) ([]*types.RightsAlert, int, error) {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("获取租户ID失败: %w", err)
	}

	model := g.DB().Model("rights_alerts").
		Ctx(ctx).
		Where("tenant_id", tenantID).
		OrderDesc("triggered_at")

	// 应用查询条件
	if query.MerchantID != nil {
		model = model.Where("merchant_id", *query.MerchantID)
	}
	if query.AlertType != nil {
		model = model.Where("alert_type", *query.AlertType)
	}
	if query.Severity != nil {
		model = model.Where("severity", *query.Severity)
	}
	if query.Status != nil {
		model = model.Where("status", *query.Status)
	}
	if query.StartTime != nil {
		model = model.Where("triggered_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		model = model.Where("triggered_at <= ?", *query.EndTime)
	}

	// 获取总数
	total, err := model.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var alerts []*types.RightsAlert
	offset := (query.Page - 1) * query.PageSize
	err = model.Limit(query.PageSize).Offset(offset).Scan(&alerts)
	if err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// UpdateAlert 更新预警记录
func (r *monitoringRepository) UpdateAlert(ctx context.Context, alert *types.RightsAlert) error {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return fmt.Errorf("获取租户ID失败: %w", err)
	}

	alert.UpdatedAt = time.Now()
	_, err := g.DB().Model("rights_alerts").
		Ctx(ctx).
		Where("id", alert.ID).
		Where("tenant_id", tenantID).
		Update(alert)

	return err
}

// ResolveAlert 解决预警
func (r *monitoringRepository) ResolveAlert(ctx context.Context, id uint64, resolution string) error {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return fmt.Errorf("获取租户ID失败: %w", err)
	}

	now := time.Now()
	_, err := g.DB().Model("rights_alerts").
		Ctx(ctx).
		Where("id", id).
		Where("tenant_id", tenantID).
		Where("status", types.AlertStatusActive).
		Update(g.Map{
			"status":      types.AlertStatusResolved,
			"resolved_at": now,
			"message":     gdb.Raw("CONCAT(message, '\n解决方案: " + resolution + "')"),
			"updated_at":  now,
		})

	return err
}

// GetActiveAlertsCount 获取活跃预警数量
func (r *monitoringRepository) GetActiveAlertsCount(ctx context.Context, merchantID *uint64) (int, error) {
	tenantID, err := r.GetTenantID(ctx)
	if err != nil {
		return 0, fmt.Errorf("获取租户ID失败: %w", err)
	}

	model := g.DB().Model("rights_alerts").
		Ctx(ctx).
		Where("tenant_id", tenantID).
		Where("status", types.AlertStatusActive)

	if merchantID != nil {
		model = model.Where("merchant_id", *merchantID)
	}

	count, err := model.Count()
	return count, err
}

// GetCriticalAlertsCount 获取严重预警数量
func (r *monitoringRepository) GetCriticalAlertsCount(ctx context.Context, merchantID *uint64) (int, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return 0, ErrTenantRequired
	}

	model := g.DB().Model("rights_alerts").
		Ctx(ctx).
		Where("tenant_id", tenantID).
		Where("status", types.AlertStatusActive).
		Where("severity", types.AlertSeverityCritical)

	if merchantID != nil {
		model = model.Where("merchant_id", *merchantID)
	}

	count, err := model.Count()
	return count, err
}

// CreateUsageStats 创建使用统计记录
func (r *monitoringRepository) CreateUsageStats(ctx context.Context, stats *types.RightsUsageStats) error {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return ErrTenantRequired
	}

	stats.TenantID = tenantID
	stats.CreatedAt = time.Now()

	_, err := g.DB().Model("rights_usage_stats").Ctx(ctx).Insert(stats)
	return err
}

// GetUsageStats 获取使用统计
func (r *monitoringRepository) GetUsageStats(ctx context.Context, query *types.RightsStatsQuery) ([]*types.RightsUsageStats, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	model := g.DB().Model("rights_usage_stats").
		Ctx(ctx).
		Where("tenant_id", tenantID).
		OrderDesc("stat_date")

	if query.MerchantID != nil {
		model = model.Where("merchant_id", *query.MerchantID)
	}
	if query.Period != nil {
		model = model.Where("period_type", *query.Period)
	}
	if query.StartDate != nil {
		model = model.Where("stat_date >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		model = model.Where("stat_date <= ?", *query.EndDate)
	}

	var stats []*types.RightsUsageStats
	err := model.Scan(&stats)
	return stats, err
}

// GetUsageTrends 获取使用趋势
func (r *monitoringRepository) GetUsageTrends(ctx context.Context, query *types.RightsTrendsQuery) ([]*types.RightsUsageStats, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	days := 30
	if query.Days != nil {
		days = *query.Days
	}

	model := g.DB().Model("rights_usage_stats").
		Ctx(ctx).
		Where("tenant_id", tenantID).
		Where("stat_date >= ?", gtime.Now().AddDate(0, 0, -days)).
		OrderAsc("stat_date")

	if query.MerchantID != nil {
		model = model.Where("merchant_id", *query.MerchantID)
	}
	if query.Period != nil {
		model = model.Where("period_type", *query.Period)
	}

	var trends []*types.RightsUsageStats
	err := model.Scan(&trends)
	return trends, err
}

// GetMerchantUsageInfo 获取商户使用信息
func (r *monitoringRepository) GetMerchantUsageInfo(ctx context.Context, limit int) ([]*types.MerchantUsageInfo, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	sql := `
		SELECT 
			m.id as merchant_id,
			m.name as merchant_name,
			COALESCE(AVG(rus.average_daily_usage), 0) as daily_usage,
			COALESCE(JSON_EXTRACT(m.rights_balance, '$.total_balance'), 0) as total_balance
		FROM merchants m
		LEFT JOIN rights_usage_stats rus ON m.id = rus.merchant_id 
			AND rus.stat_date >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
		WHERE m.tenant_id = ?
		GROUP BY m.id, m.name
		ORDER BY daily_usage DESC
		LIMIT ?`

	var results []*types.MerchantUsageInfo
	err := g.DB().Ctx(ctx).GetScan(&results, sql, tenantID, limit)
	return results, err
}

// GetDailyUsageTrends 获取日使用趋势
func (r *monitoringRepository) GetDailyUsageTrends(ctx context.Context, days int, merchantID *uint64) ([]*types.DailyUsageTrend, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	sql := `
		SELECT 
			stat_date as date,
			SUM(total_consumed) as total_usage
		FROM rights_usage_stats
		WHERE tenant_id = ? 
			AND stat_date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
			AND period_type = ?`

	params := []interface{}{tenantID, days, types.TimePeriodDaily}

	if merchantID != nil {
		sql += " AND merchant_id = ?"
		params = append(params, *merchantID)
	}

	sql += " GROUP BY stat_date ORDER BY stat_date ASC"

	var trends []*types.DailyUsageTrend
	err := g.DB().Ctx(ctx).GetScan(&trends, sql, params...)
	return trends, err
}

// GetDashboardData 获取仪表板数据
func (r *monitoringRepository) GetDashboardData(ctx context.Context, merchantID *uint64) (*types.MonitoringDashboardData, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	data := &types.MonitoringDashboardData{}

	// 获取商户总数
	merchantCount, err := g.DB().Model("merchants").
		Ctx(ctx).
		Where("tenant_id", tenantID).
		Where("status", "active").
		Count()
	if err != nil {
		return nil, err
	}
	data.TotalMerchants = merchantCount

	// 获取活跃预警数量
	activeAlerts, err := r.GetActiveAlertsCount(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	data.ActiveAlerts = activeAlerts

	// 获取严重预警数量
	criticalAlerts, err := r.GetCriticalAlertsCount(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	data.CriticalAlerts = criticalAlerts

	// 获取总权益余额
	var totalBalance float64
	balanceSQL := `
		SELECT COALESCE(SUM(JSON_EXTRACT(rights_balance, '$.total_balance')), 0)
		FROM merchants 
		WHERE tenant_id = ? AND status = 'active'`
	if merchantID != nil {
		balanceSQL += " AND id = ?"
		err = g.DB().Ctx(ctx).GetVar(balanceSQL, tenantID, *merchantID).Scan(&totalBalance)
	} else {
		err = g.DB().Ctx(ctx).GetVar(balanceSQL, tenantID).Scan(&totalBalance)
	}
	if err != nil {
		return nil, err
	}
	data.TotalRightsBalance = totalBalance

	// 获取平均日使用量
	avgUsageSQL := `
		SELECT COALESCE(AVG(average_daily_usage), 0)
		FROM rights_usage_stats 
		WHERE tenant_id = ? 
			AND stat_date >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
			AND period_type = ?`
	
	if merchantID != nil {
		avgUsageSQL += " AND merchant_id = ?"
		err = g.DB().Ctx(ctx).GetVar(avgUsageSQL, tenantID, types.TimePeriodDaily, *merchantID).Scan(&data.AvgDailyUsage)
	} else {
		err = g.DB().Ctx(ctx).GetVar(avgUsageSQL, tenantID, types.TimePeriodDaily).Scan(&data.AvgDailyUsage)
	}
	if err != nil {
		return nil, err
	}

	// 获取使用量排名前5的商户
	data.TopMerchantsByUsage, err = r.GetMerchantUsageInfo(ctx, 5)
	if err != nil {
		return nil, err
	}

	// 获取最近5条预警
	recentAlertsQuery := &types.AlertListQuery{
		Page:       1,
		PageSize:   5,
		MerchantID: merchantID,
	}
	recentAlerts, _, err := r.ListAlerts(ctx, recentAlertsQuery)
	if err != nil {
		return nil, err
	}
	
	// Convert to slice of values instead of pointers
	data.RecentAlerts = make([]types.RightsAlert, len(recentAlerts))
	for i, alert := range recentAlerts {
		data.RecentAlerts[i] = *alert
	}

	// 获取最近30天使用趋势
	data.UsageTrends, err = r.GetDailyUsageTrends(ctx, 30, merchantID)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// UpdateMerchantThresholds 更新商户预警阈值
func (r *monitoringRepository) UpdateMerchantThresholds(ctx context.Context, merchantID uint64, warningThreshold, criticalThreshold *float64) error {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return ErrTenantRequired
	}

	// 构建更新的rights_balance JSON
	updateData := g.Map{
		"updated_at": time.Now(),
	}

	if warningThreshold != nil || criticalThreshold != nil {
		// 先获取当前的rights_balance
		var currentBalance string
		err := g.DB().Model("merchants").
			Ctx(ctx).
			Where("id", merchantID).
			Where("tenant_id", tenantID).
			Fields("rights_balance").
			Scan(&currentBalance)
		if err != nil {
			return err
		}

		// 解析当前JSON并更新阈值字段
		var balanceMap map[string]interface{}
		if currentBalance != "" {
			if err := g.JSON().Unmarshal([]byte(currentBalance), &balanceMap); err != nil {
				balanceMap = make(map[string]interface{})
			}
		} else {
			balanceMap = make(map[string]interface{})
		}

		if warningThreshold != nil {
			balanceMap["warning_threshold"] = *warningThreshold
		}
		if criticalThreshold != nil {
			balanceMap["critical_threshold"] = *criticalThreshold
		}

		newBalanceJSON, err := g.JSON().Marshal(balanceMap)
		if err != nil {
			return err
		}
		updateData["rights_balance"] = string(newBalanceJSON)
	}

	_, err := g.DB().Model("merchants").
		Ctx(ctx).
		Where("id", merchantID).
		Where("tenant_id", tenantID).
		Update(updateData)

	return err
}

// GetMerchantThresholds 获取商户预警阈值
func (r *monitoringRepository) GetMerchantThresholds(ctx context.Context, merchantID uint64) (warning, critical *float64, err error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, nil, ErrTenantRequired
	}

	var balanceJSON string
	err = g.DB().Model("merchants").
		Ctx(ctx).
		Where("id", merchantID).
		Where("tenant_id", tenantID).
		Fields("rights_balance").
		Scan(&balanceJSON)

	if err != nil {
		return nil, nil, err
	}

	if balanceJSON == "" {
		return nil, nil, nil
	}

	var balanceMap map[string]interface{}
	if err := g.JSON().Unmarshal([]byte(balanceJSON), &balanceMap); err != nil {
		return nil, nil, err
	}

	if warningVal, ok := balanceMap["warning_threshold"]; ok && warningVal != nil {
		if w, ok := warningVal.(float64); ok {
			warning = &w
		}
	}

	if criticalVal, ok := balanceMap["critical_threshold"]; ok && criticalVal != nil {
		if c, ok := criticalVal.(float64); ok {
			critical = &c
		}
	}

	return warning, critical, nil
}
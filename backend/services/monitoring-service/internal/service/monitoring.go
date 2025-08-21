package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"mer-demo/shared/notification"
	"mer-demo/shared/repository"
	"mer-demo/shared/types"
)

// MonitoringService 监控服务接口
type MonitoringService interface {
	// 统计和趋势分析
	GetRightsStats(ctx context.Context, query *types.RightsStatsQuery) ([]*types.RightsUsageStats, error)
	GetRightsTrends(ctx context.Context, query *types.RightsTrendsQuery) ([]*types.RightsUsageStats, error)
	CalculateUsageTrend(ctx context.Context, merchantID uint64, days int) (*types.TrendDirection, error)
	PredictDepletionDate(ctx context.Context, merchantID uint64) (*time.Time, error)

	// 预警管理
	ConfigureAlerts(ctx context.Context, req *types.AlertConfigureRequest) error
	ListAlerts(ctx context.Context, query *types.AlertListQuery) ([]*types.RightsAlert, int, error)
	ResolveAlert(ctx context.Context, alertID uint64, resolution string) error
	TriggerAlert(ctx context.Context, alert *types.RightsAlert) error
	CheckAlertConditions(ctx context.Context, merchantID uint64) error

	// 仪表板和报告
	GetDashboardData(ctx context.Context, merchantID *uint64) (*types.MonitoringDashboardData, error)
	GenerateReport(ctx context.Context, req *types.ReportGenerateRequest) (string, error)

	// 定时任务
	RunPeriodicChecks(ctx context.Context) error
	CollectUsageData(ctx context.Context) error
}

// monitoringService 监控服务实现
type monitoringService struct {
	monitoringRepo    repository.MonitoringRepository
	fundRepo          repository.FundRepository
	merchantRepo      repository.MerchantRepository
	notificationSvc   notification.NotificationService
}

// NewMonitoringService 创建监控服务实例
func NewMonitoringService() MonitoringService {
	return &monitoringService{
		monitoringRepo:  repository.NewMonitoringRepository(),
		fundRepo:        repository.NewFundRepository(),
		merchantRepo:    repository.NewMerchantRepository(),
		notificationSvc: notification.NewNotificationService(),
	}
}

// GetRightsStats 获取权益使用统计
func (s *monitoringService) GetRightsStats(ctx context.Context, query *types.RightsStatsQuery) ([]*types.RightsUsageStats, error) {
	return s.monitoringRepo.GetUsageStats(ctx, query)
}

// GetRightsTrends 获取权益使用趋势
func (s *monitoringService) GetRightsTrends(ctx context.Context, query *types.RightsTrendsQuery) ([]*types.RightsUsageStats, error) {
	return s.monitoringRepo.GetUsageTrends(ctx, query)
}

// CalculateUsageTrend 计算使用趋势
func (s *monitoringService) CalculateUsageTrend(ctx context.Context, merchantID uint64, days int) (*types.TrendDirection, error) {
	trends, err := s.monitoringRepo.GetUsageTrends(ctx, &types.RightsTrendsQuery{
		MerchantID: &merchantID,
		Days:       &days,
	})
	if err != nil {
		return nil, err
	}

	if len(trends) < 2 {
		direction := types.TrendDirectionStable
		return &direction, nil
	}

	// 计算线性回归斜率来确定趋势
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(trends))

	for i, trend := range trends {
		x := float64(i)
		y := trend.AverageDailyUsage
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算斜率
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	var direction types.TrendDirection
	if slope > 0.1 {
		direction = types.TrendDirectionIncreasing
	} else if slope < -0.1 {
		direction = types.TrendDirectionDecreasing
	} else {
		direction = types.TrendDirectionStable
	}

	return &direction, nil
}

// PredictDepletionDate 预测权益耗尽日期
func (s *monitoringService) PredictDepletionDate(ctx context.Context, merchantID uint64) (*time.Time, error) {
	// 获取商户当前余额
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	if merchant.RightsBalance == nil || merchant.RightsBalance.AvailableBalance <= 0 {
		return nil, nil
	}

	// 获取最近30天的平均使用量
	trends, err := s.monitoringRepo.GetUsageTrends(ctx, &types.RightsTrendsQuery{
		MerchantID: &merchantID,
		Days:       ptrOf(30),
	})
	if err != nil {
		return nil, err
	}

	if len(trends) == 0 {
		return nil, nil
	}

	// 计算平均日使用量
	var totalUsage float64
	for _, trend := range trends {
		totalUsage += trend.AverageDailyUsage
	}
	avgDailyUsage := totalUsage / float64(len(trends))

	if avgDailyUsage <= 0 {
		return nil, nil
	}

	// 计算预计耗尽天数
	daysToDepletion := merchant.RightsBalance.AvailableBalance / avgDailyUsage
	depletionDate := time.Now().AddDate(0, 0, int(math.Ceil(daysToDepletion)))

	return &depletionDate, nil
}

// ConfigureAlerts 配置预警
func (s *monitoringService) ConfigureAlerts(ctx context.Context, req *types.AlertConfigureRequest) error {
	// 验证请求
	if err := req.Validate(); err != nil {
		return err
	}

	// 更新商户的预警阈值
	return s.monitoringRepo.UpdateMerchantThresholds(ctx, req.MerchantID, req.WarningThreshold, req.CriticalThreshold)
}

// ListAlerts 获取预警列表
func (s *monitoringService) ListAlerts(ctx context.Context, query *types.AlertListQuery) ([]*types.RightsAlert, int, error) {
	return s.monitoringRepo.ListAlerts(ctx, query)
}

// ResolveAlert 解决预警
func (s *monitoringService) ResolveAlert(ctx context.Context, alertID uint64, resolution string) error {
	return s.monitoringRepo.ResolveAlert(ctx, alertID, resolution)
}

// TriggerAlert 触发预警
func (s *monitoringService) TriggerAlert(ctx context.Context, alert *types.RightsAlert) error {
	// 检查是否已存在相同类型的活跃预警
	existingAlerts, _, err := s.monitoringRepo.ListAlerts(ctx, &types.AlertListQuery{
		Page:       1,
		PageSize:   1,
		MerchantID: &alert.MerchantID,
		AlertType:  &alert.AlertType,
		Status:     ptrOf(types.AlertStatusActive),
	})
	if err != nil {
		return err
	}

	// 如果已存在活跃预警，更新而不是创建新的
	if len(existingAlerts) > 0 {
		existingAlert := existingAlerts[0]
		existingAlert.CurrentValue = alert.CurrentValue
		existingAlert.Message = alert.Message
		existingAlert.UpdatedAt = time.Now()
		return s.monitoringRepo.UpdateAlert(ctx, existingAlert)
	}

	// 创建新预警
	alert.Status = types.AlertStatusActive
	alert.TriggeredAt = time.Now()
	
	if err := s.monitoringRepo.CreateAlert(ctx, alert); err != nil {
		return err
	}

	// 发送通知
	channels := []notification.NotificationChannel{
		notification.ChannelSystem,
		notification.ChannelEmail,
	}

	// 对于严重预警，也发送短信
	if alert.Severity == types.AlertSeverityCritical {
		channels = append(channels, notification.ChannelSMS)
	}

	if err := s.notificationSvc.SendAlert(ctx, alert, channels); err != nil {
		g.Log().Error(ctx, "Failed to send alert notifications", g.Map{
			"alert_id": alert.ID,
			"error":    err,
		})
		// 通知发送失败不影响预警创建
	}

	return nil
}

// CheckAlertConditions 检查预警条件
func (s *monitoringService) CheckAlertConditions(ctx context.Context, merchantID uint64) error {
	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		return err
	}

	if merchant.RightsBalance == nil {
		return nil
	}

	balance := merchant.RightsBalance
	availableBalance := balance.GetAvailableBalance()

	// 检查余额预警
	if balance.CriticalThreshold != nil && availableBalance <= *balance.CriticalThreshold {
		alert := &types.RightsAlert{
			MerchantID:     merchantID,
			AlertType:      types.AlertTypeBalanceCritical,
			ThresholdValue: *balance.CriticalThreshold,
			CurrentValue:   availableBalance,
			Severity:       types.AlertSeverityCritical,
			Message:        fmt.Sprintf("商户 %s 权益余额已低于紧急阈值 %.2f，当前余额：%.2f", merchant.Name, *balance.CriticalThreshold, availableBalance),
		}
		if err := s.TriggerAlert(ctx, alert); err != nil {
			g.Log().Error(ctx, "Failed to trigger critical balance alert", err)
		}
	} else if balance.WarningThreshold != nil && availableBalance <= *balance.WarningThreshold {
		alert := &types.RightsAlert{
			MerchantID:     merchantID,
			AlertType:      types.AlertTypeBalanceLow,
			ThresholdValue: *balance.WarningThreshold,
			CurrentValue:   availableBalance,
			Severity:       types.AlertSeverityWarning,
			Message:        fmt.Sprintf("商户 %s 权益余额接近预警阈值 %.2f，当前余额：%.2f", merchant.Name, *balance.WarningThreshold, availableBalance),
		}
		if err := s.TriggerAlert(ctx, alert); err != nil {
			g.Log().Error(ctx, "Failed to trigger warning balance alert", err)
		}
	}

	// 检查使用量激增
	trend, err := s.CalculateUsageTrend(ctx, merchantID, 7)
	if err == nil && *trend == types.TrendDirectionIncreasing {
		// 获取最近7天的平均使用量
		trends, err := s.monitoringRepo.GetUsageTrends(ctx, &types.RightsTrendsQuery{
			MerchantID: &merchantID,
			Days:       ptrOf(7),
		})
		if err == nil && len(trends) > 0 {
			var recentAvg float64
			for _, t := range trends {
				recentAvg += t.AverageDailyUsage
			}
			recentAvg /= float64(len(trends))

			// 获取之前的平均使用量作为基准
			olderTrends, err := s.monitoringRepo.GetUsageTrends(ctx, &types.RightsTrendsQuery{
				MerchantID: &merchantID,
				Days:       ptrOf(30),
			})
			if err == nil && len(olderTrends) > 7 {
				var olderAvg float64
				for i := 0; i < len(olderTrends)-7; i++ {
					olderAvg += olderTrends[i].AverageDailyUsage
				}
				olderAvg /= float64(len(olderTrends) - 7)

				// 如果最近使用量比之前增长超过50%，触发使用激增预警
				if recentAvg > olderAvg*1.5 {
					alert := &types.RightsAlert{
						MerchantID:     merchantID,
						AlertType:      types.AlertTypeUsageSpike,
						ThresholdValue: olderAvg * 1.5,
						CurrentValue:   recentAvg,
						Severity:       types.AlertSeverityWarning,
						Message:        fmt.Sprintf("商户 %s 最近使用量激增，当前日均：%.2f，历史日均：%.2f", merchant.Name, recentAvg, olderAvg),
					}
					if err := s.TriggerAlert(ctx, alert); err != nil {
						g.Log().Error(ctx, "Failed to trigger usage spike alert", err)
					}
				}
			}
		}
	}

	// 检查预计耗尽时间
	depletionDate, err := s.PredictDepletionDate(ctx, merchantID)
	if err == nil && depletionDate != nil {
		daysUntilDepletion := int(depletionDate.Sub(time.Now()).Hours() / 24)
		if daysUntilDepletion <= 7 && daysUntilDepletion > 0 {
			alert := &types.RightsAlert{
				MerchantID:     merchantID,
				AlertType:      types.AlertTypePredictedDepletion,
				ThresholdValue: 7,
				CurrentValue:   float64(daysUntilDepletion),
				Severity:       types.AlertSeverityWarning,
				Message:        fmt.Sprintf("商户 %s 预计在 %d 天后权益余额耗尽，预计日期：%s", merchant.Name, daysUntilDepletion, depletionDate.Format("2006-01-02")),
			}
			if err := s.TriggerAlert(ctx, alert); err != nil {
				g.Log().Error(ctx, "Failed to trigger depletion prediction alert", err)
			}
		}
	}

	return nil
}

// GetDashboardData 获取仪表板数据
func (s *monitoringService) GetDashboardData(ctx context.Context, merchantID *uint64) (*types.MonitoringDashboardData, error) {
	return s.monitoringRepo.GetDashboardData(ctx, merchantID)
}

// GenerateReport 生成报告
func (s *monitoringService) GenerateReport(ctx context.Context, req *types.ReportGenerateRequest) (string, error) {
	// 验证请求
	if err := req.Validate(); err != nil {
		return "", err
	}

	// 获取报告数据
	query := &types.RightsStatsQuery{
		Period:    &req.Period,
		StartDate: &req.StartDate,
		EndDate:   &req.EndDate,
	}

	var allStats []*types.RightsUsageStats
	if len(req.MerchantIDs) > 0 {
		// 为每个商户获取数据
		for _, merchantID := range req.MerchantIDs {
			query.MerchantID = &merchantID
			stats, err := s.monitoringRepo.GetUsageStats(ctx, query)
			if err != nil {
				return "", err
			}
			allStats = append(allStats, stats...)
		}
	} else {
		// 获取所有商户数据
		stats, err := s.monitoringRepo.GetUsageStats(ctx, query)
		if err != nil {
			return "", err
		}
		allStats = stats
	}

	// 生成报告文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("rights_usage_report_%s.%s", timestamp, req.Format)

	// 这里应该根据格式生成相应的报告文件
	// 为简化实现，这里返回文件名
	// 实际项目中需要实现PDF、Excel、CSV的生成逻辑

	g.Log().Info(ctx, "Generated report", g.Map{
		"filename":    filename,
		"format":      req.Format,
		"data_points": len(allStats),
		"period":      req.Period,
	})

	return filename, nil
}

// RunPeriodicChecks 运行定期检查
func (s *monitoringService) RunPeriodicChecks(ctx context.Context) error {
	g.Log().Info(ctx, "Starting periodic monitoring checks")

	// 获取所有活跃商户
	merchants, _, err := s.merchantRepo.List(ctx, &types.MerchantListQuery{
		Page:     1,
		PageSize: 1000,
		Status:   types.MerchantStatusActive,
	})
	if err != nil {
		return err
	}

	// 为每个商户检查预警条件
	for _, merchant := range merchants {
		if err := s.CheckAlertConditions(ctx, merchant.ID); err != nil {
			g.Log().Error(ctx, "Failed to check alert conditions for merchant", g.Map{
				"merchant_id": merchant.ID,
				"error":       err,
			})
		}
	}

	g.Log().Info(ctx, "Completed periodic monitoring checks", g.Map{
		"merchants_checked": len(merchants),
	})

	return nil
}

// CollectUsageData 收集使用数据
func (s *monitoringService) CollectUsageData(ctx context.Context) error {
	g.Log().Info(ctx, "Starting usage data collection")

	// 获取所有活跃商户
	merchants, _, err := s.merchantRepo.List(ctx, &types.MerchantListQuery{
		Page:     1,
		PageSize: 1000,
		Status:   types.MerchantStatusActive,
	})
	if err != nil {
		return err
	}

	today := gtime.Now().Format("Y-m-d")
	
	// 为每个商户收集今日使用数据
	for _, merchant := range merchants {
		// 计算今日消费总量
		fundQuery := &types.FundTransactionQuery{
			Page:         1,
			PageSize:     1000,
			MerchantID:   &merchant.ID,
			FundType:     ptrOf(types.FundTypeConsumption),
			StartTime:    ptrOf(gtime.Now().StartOfDay().Time),
			EndTime:      ptrOf(gtime.Now().EndOfDay().Time),
		}

		transactions, _, err := s.fundRepo.ListTransactions(ctx, fundQuery)
		if err != nil {
			g.Log().Error(ctx, "Failed to get fund transactions for merchant", g.Map{
				"merchant_id": merchant.ID,
				"error":       err,
			})
			continue
		}

		var totalConsumed float64
		for _, tx := range transactions {
			totalConsumed += tx.Amount
		}

		// 创建使用统计记录
		stats := &types.RightsUsageStats{
			MerchantID:        &merchant.ID,
			StatDate:          gtime.Now().StartOfDay().Time,
			Period:            types.TimePeriodDaily,
			TotalConsumed:     totalConsumed,
			AverageDailyUsage: totalConsumed,
		}

		// 计算总分配量（从商户余额中获取）
		if merchant.RightsBalance != nil {
			stats.TotalAllocated = merchant.RightsBalance.TotalBalance
		}

		// 计算趋势
		trend, err := s.CalculateUsageTrend(ctx, merchant.ID, 7)
		if err == nil {
			stats.UsageTrend = *trend
		}

		// 预测耗尽日期
		depletionDate, err := s.PredictDepletionDate(ctx, merchant.ID)
		if err == nil && depletionDate != nil {
			stats.PredictedDepletionDate = depletionDate
		}

		// 保存统计数据
		if err := s.monitoringRepo.CreateUsageStats(ctx, stats); err != nil {
			g.Log().Error(ctx, "Failed to create usage stats for merchant", g.Map{
				"merchant_id": merchant.ID,
				"error":       err,
			})
		}
	}

	g.Log().Info(ctx, "Completed usage data collection", g.Map{
		"merchants_processed": len(merchants),
		"date":               today,
	})

	return nil
}

// ptrOf helper function for creating pointers
func ptrOf[T any](v T) *T {
	return &v
}
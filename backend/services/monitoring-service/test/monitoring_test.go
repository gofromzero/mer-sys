package test

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/test/gtest"

	"mer-demo/services/monitoring-service/internal/service"
	"mer-demo/shared/types"
)

// TestMonitoringService_GetRightsStats 测试权益统计获取
func TestMonitoringService_GetRightsStats(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 准备测试查询
		query := &types.RightsStatsQuery{
			Page:     1,
			PageSize: 10,
			Period:   ptrOf(types.TimePeriodDaily),
		}

		// 执行查询
		stats, err := monitoringService.GetRightsStats(ctx, query)

		// 验证结果
		t.AssertNil(err)
		t.AssertNotNil(stats)
		t.Assert(len(stats), ">=", 0) // 允许空结果
	})
}

// TestMonitoringService_CalculateUsageTrend 测试使用趋势计算
func TestMonitoringService_CalculateUsageTrend(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 使用测试商户ID
		merchantID := uint64(1)
		days := 7

		// 计算趋势
		trend, err := monitoringService.CalculateUsageTrend(ctx, merchantID, days)

		// 验证结果
		t.AssertNil(err)
		t.AssertNotNil(trend)
		t.AssertIN(*trend, []types.TrendDirection{
			types.TrendDirectionIncreasing,
			types.TrendDirectionDecreasing,
			types.TrendDirectionStable,
		})
	})
}

// TestMonitoringService_PredictDepletionDate 测试权益耗尽预测
func TestMonitoringService_PredictDepletionDate(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 使用测试商户ID
		merchantID := uint64(1)

		// 预测耗尽日期
		depletionDate, err := monitoringService.PredictDepletionDate(ctx, merchantID)

		// 验证结果（可能返回nil如果无法预测）
		t.AssertNil(err)
		if depletionDate != nil {
			t.Assert(depletionDate.After(time.Now()), true)
		}
	})
}

// TestMonitoringService_ConfigureAlerts 测试预警配置
func TestMonitoringService_ConfigureAlerts(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 准备配置请求
		req := &types.AlertConfigureRequest{
			MerchantID:        1,
			WarningThreshold:  ptrOf(1000.0),
			CriticalThreshold: ptrOf(500.0),
		}

		// 配置预警
		err := monitoringService.ConfigureAlerts(ctx, req)

		// 验证结果
		t.AssertNil(err)
	})
}

// TestMonitoringService_TriggerAlert 测试预警触发
func TestMonitoringService_TriggerAlert(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 准备预警数据
		alert := &types.RightsAlert{
			MerchantID:     1,
			AlertType:      types.AlertTypeBalanceLow,
			ThresholdValue: 1000.0,
			CurrentValue:   800.0,
			Severity:       types.AlertSeverityWarning,
			Message:        "测试预警消息",
		}

		// 触发预警
		err := monitoringService.TriggerAlert(ctx, alert)

		// 验证结果
		t.AssertNil(err)
		t.Assert(alert.Status, types.AlertStatusActive)
		t.Assert(alert.TriggeredAt.IsZero(), false)
	})
}

// TestMonitoringService_CheckAlertConditions 测试预警条件检查
func TestMonitoringService_CheckAlertConditions(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 使用测试商户ID
		merchantID := uint64(1)

		// 检查预警条件
		err := monitoringService.CheckAlertConditions(ctx, merchantID)

		// 验证结果（不应该出错，即使没有触发预警）
		t.AssertNil(err)
	})
}

// TestMonitoringService_RunPeriodicChecks 测试定期检查
func TestMonitoringService_RunPeriodicChecks(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 运行定期检查
		err := monitoringService.RunPeriodicChecks(ctx)

		// 验证结果
		t.AssertNil(err)
	})
}

// TestMonitoringService_CollectUsageData 测试使用数据收集
func TestMonitoringService_CollectUsageData(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 收集使用数据
		err := monitoringService.CollectUsageData(ctx)

		// 验证结果
		t.AssertNil(err)
	})
}

// TestMonitoringService_GetDashboardData 测试仪表板数据获取
func TestMonitoringService_GetDashboardData(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 获取仪表板数据
		data, err := monitoringService.GetDashboardData(ctx, nil)

		// 验证结果
		t.AssertNil(err)
		t.AssertNotNil(data)
		t.Assert(data.TotalMerchants, ">=", 0)
		t.Assert(data.ActiveAlerts, ">=", 0)
		t.AssertNotNil(data.RecentAlerts)
		t.AssertNotNil(data.UsageTrends)
	})
}

// TestMonitoringService_GenerateReport 测试报告生成
func TestMonitoringService_GenerateReport(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		ctx := context.Background()
		monitoringService := service.NewMonitoringService()

		// 准备报告请求
		req := &types.ReportGenerateRequest{
			Period:      types.TimePeriodDaily,
			StartDate:   time.Now().AddDate(0, 0, -7),
			EndDate:     time.Now(),
			MerchantIDs: []uint64{1, 2},
			Format:      "excel",
		}

		// 生成报告
		filename, err := monitoringService.GenerateReport(ctx, req)

		// 验证结果
		t.AssertNil(err)
		t.AssertNE(filename, "")
		t.AssertStrContains(filename, ".excel")
	})
}

// TestAlertTypeString 测试预警类型字符串转换
func TestAlertTypeString(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		testCases := map[types.AlertType]string{
			types.AlertTypeBalanceLow:        "balance_low",
			types.AlertTypeBalanceCritical:   "balance_critical",
			types.AlertTypeUsageSpike:        "usage_spike",
			types.AlertTypePredictedDepletion: "predicted_depletion",
		}

		for alertType, expected := range testCases {
			t.Assert(alertType.String(), expected)
		}
	})
}

// TestAlertSeverityString 测试预警严重程度字符串转换
func TestAlertSeverityString(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		testCases := map[types.AlertSeverity]string{
			types.AlertSeverityInfo:     "info",
			types.AlertSeverityWarning:  "warning",
			types.AlertSeverityCritical: "critical",
		}

		for severity, expected := range testCases {
			t.Assert(severity.String(), expected)
		}
	})
}

// TestTrendDirectionString 测试趋势方向字符串转换
func TestTrendDirectionString(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		testCases := map[types.TrendDirection]string{
			types.TrendDirectionIncreasing: "increasing",
			types.TrendDirectionDecreasing: "decreasing",
			types.TrendDirectionStable:     "stable",
		}

		for trend, expected := range testCases {
			t.Assert(trend.String(), expected)
		}
	})
}

// ptrOf helper function for creating pointers
func ptrOf[T any](v T) *T {
	return &v
}
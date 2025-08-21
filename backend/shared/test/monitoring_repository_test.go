package test

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/test/gtest"

	"mer-demo/shared/repository"
	"mer-demo/shared/types"
)

// TestMonitoringRepository_TenantIsolation 测试监控Repository的租户隔离
func TestMonitoringRepository_TenantIsolation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()

		// 准备两个不同租户的上下文
		tenant1Ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
		tenant2Ctx := context.WithValue(context.Background(), "tenant_id", uint64(2))

		// 为租户1创建预警
		alert1 := &types.RightsAlert{
			MerchantID:     100,
			AlertType:      types.AlertTypeBalanceLow,
			ThresholdValue: 1000.0,
			CurrentValue:   800.0,
			Severity:       types.AlertSeverityWarning,
			Message:        "租户1的预警",
			Status:         types.AlertStatusActive,
			TriggeredAt:    time.Now(),
		}

		err := repo.CreateAlert(tenant1Ctx, alert1)
		t.AssertNil(err)

		// 为租户2创建预警
		alert2 := &types.RightsAlert{
			MerchantID:     200,
			AlertType:      types.AlertTypeBalanceCritical,
			ThresholdValue: 500.0,
			CurrentValue:   300.0,
			Severity:       types.AlertSeverityCritical,
			Message:        "租户2的预警",
			Status:         types.AlertStatusActive,
			TriggeredAt:    time.Now(),
		}

		err = repo.CreateAlert(tenant2Ctx, alert2)
		t.AssertNil(err)

		// 验证租户隔离：租户1只能看到自己的预警
		tenant1Alerts, count1, err := repo.ListAlerts(tenant1Ctx, &types.AlertListQuery{
			Page:     1,
			PageSize: 10,
		})
		t.AssertNil(err)
		t.Assert(count1, ">=", 1)

		// 检查是否只返回租户1的数据
		for _, alert := range tenant1Alerts {
			t.AssertNE(alert.MerchantID, uint64(200)) // 不应该包含租户2的商户
		}

		// 验证租户隔离：租户2只能看到自己的预警
		tenant2Alerts, count2, err := repo.ListAlerts(tenant2Ctx, &types.AlertListQuery{
			Page:     1,
			PageSize: 10,
		})
		t.AssertNil(err)
		t.Assert(count2, ">=", 1)

		// 检查是否只返回租户2的数据
		for _, alert := range tenant2Alerts {
			t.AssertNE(alert.MerchantID, uint64(100)) // 不应该包含租户1的商户
		}
	})
}

// TestMonitoringRepository_CreateAlert 测试预警创建
func TestMonitoringRepository_CreateAlert(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		alert := &types.RightsAlert{
			MerchantID:     1,
			AlertType:      types.AlertTypeBalanceLow,
			ThresholdValue: 1000.0,
			CurrentValue:   800.0,
			Severity:       types.AlertSeverityWarning,
			Message:        "测试预警创建",
			Status:         types.AlertStatusActive,
			TriggeredAt:    time.Now(),
		}

		err := repo.CreateAlert(ctx, alert)
		t.AssertNil(err)
		t.Assert(alert.ID, ">", uint64(0)) // ID应该被设置
	})
}

// TestMonitoringRepository_UpdateAlert 测试预警更新
func TestMonitoringRepository_UpdateAlert(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		// 先创建一个预警
		alert := &types.RightsAlert{
			MerchantID:     1,
			AlertType:      types.AlertTypeBalanceLow,
			ThresholdValue: 1000.0,
			CurrentValue:   800.0,
			Severity:       types.AlertSeverityWarning,
			Message:        "原始预警",
			Status:         types.AlertStatusActive,
			TriggeredAt:    time.Now(),
		}

		err := repo.CreateAlert(ctx, alert)
		t.AssertNil(err)

		// 更新预警
		alert.CurrentValue = 700.0
		alert.Message = "更新后的预警"
		alert.UpdatedAt = time.Now()

		err = repo.UpdateAlert(ctx, alert)
		t.AssertNil(err)
	})
}

// TestMonitoringRepository_ResolveAlert 测试预警解决
func TestMonitoringRepository_ResolveAlert(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		// 先创建一个预警
		alert := &types.RightsAlert{
			MerchantID:     1,
			AlertType:      types.AlertTypeBalanceLow,
			ThresholdValue: 1000.0,
			CurrentValue:   800.0,
			Severity:       types.AlertSeverityWarning,
			Message:        "待解决的预警",
			Status:         types.AlertStatusActive,
			TriggeredAt:    time.Now(),
		}

		err := repo.CreateAlert(ctx, alert)
		t.AssertNil(err)

		// 解决预警
		err = repo.ResolveAlert(ctx, alert.ID, "手动充值解决")
		t.AssertNil(err)
	})
}

// TestMonitoringRepository_CreateUsageStats 测试使用统计创建
func TestMonitoringRepository_CreateUsageStats(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		stats := &types.RightsUsageStats{
			MerchantID:        ptrOf(uint64(1)),
			StatDate:          time.Now(),
			Period:            types.TimePeriodDaily,
			TotalAllocated:    10000.0,
			TotalConsumed:     3000.0,
			AverageDailyUsage: 1000.0,
			UsageTrend:        types.TrendDirectionStable,
		}

		err := repo.CreateUsageStats(ctx, stats)
		t.AssertNil(err)
	})
}

// TestMonitoringRepository_GetUsageStats 测试使用统计查询
func TestMonitoringRepository_GetUsageStats(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		query := &types.RightsStatsQuery{
			Period: ptrOf(types.TimePeriodDaily),
		}

		stats, err := repo.GetUsageStats(ctx, query)
		t.AssertNil(err)
		t.AssertNotNil(stats)
	})
}

// TestMonitoringRepository_GetUsageTrends 测试使用趋势查询
func TestMonitoringRepository_GetUsageTrends(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		query := &types.RightsTrendsQuery{
			MerchantID: ptrOf(uint64(1)),
			Days:       ptrOf(7),
		}

		trends, err := repo.GetUsageTrends(ctx, query)
		t.AssertNil(err)
		t.AssertNotNil(trends)
	})
}

// TestMonitoringRepository_GetDashboardData 测试仪表板数据查询
func TestMonitoringRepository_GetDashboardData(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		data, err := repo.GetDashboardData(ctx, nil)
		t.AssertNil(err)
		t.AssertNotNil(data)
		t.Assert(data.TotalMerchants, ">=", 0)
		t.Assert(data.ActiveAlerts, ">=", 0)
	})
}

// TestMonitoringRepository_UpdateMerchantThresholds 测试商户阈值更新
func TestMonitoringRepository_UpdateMerchantThresholds(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		merchantID := uint64(1)
		warningThreshold := 1000.0
		criticalThreshold := 500.0

		err := repo.UpdateMerchantThresholds(ctx, merchantID, &warningThreshold, &criticalThreshold)
		t.AssertNil(err)
	})
}

// TestMonitoringRepository_ListAlertsWithFilters 测试带筛选条件的预警列表查询
func TestMonitoringRepository_ListAlertsWithFilters(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		repo := repository.NewMonitoringRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		// 测试不同的筛选条件
		testCases := []types.AlertListQuery{
			{
				Page:     1,
				PageSize: 10,
			},
			{
				Page:      1,
				PageSize:  10,
				AlertType: ptrOf(types.AlertTypeBalanceLow),
			},
			{
				Page:     1,
				PageSize: 10,
				Severity: ptrOf(types.AlertSeverityWarning),
			},
			{
				Page:     1,
				PageSize: 10,
				Status:   ptrOf(types.AlertStatusActive),
			},
		}

		for _, query := range testCases {
			alerts, total, err := repo.ListAlerts(ctx, &query)
			t.AssertNil(err)
			t.AssertNotNil(alerts)
			t.Assert(total, ">=", 0)
		}
	})
}

// ptrOf helper function for creating pointers
func ptrOf[T any](v T) *T {
	return &v
}
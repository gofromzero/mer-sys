package test

import (
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
	"mer-demo/shared/types"
)

// TestMonitoringTypes 测试监控相关类型定义
func TestMonitoringTypes(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 测试预警类型枚举
		t.Assert(types.AlertTypeBalanceLow.String(), "balance_low")
		t.Assert(types.AlertTypeBalanceCritical.String(), "balance_critical")
		t.Assert(types.AlertTypeUsageSpike.String(), "usage_spike")
		t.Assert(types.AlertTypePredictedDepletion.String(), "predicted_depletion")

		// 测试预警严重程度枚举
		t.Assert(types.AlertSeverityInfo.String(), "info")
		t.Assert(types.AlertSeverityWarning.String(), "warning")
		t.Assert(types.AlertSeverityCritical.String(), "critical")

		// 测试预警状态枚举
		t.Assert(types.AlertStatusActive.String(), "active")
		t.Assert(types.AlertStatusResolved.String(), "resolved")
		t.Assert(types.AlertStatusAcknowledged.String(), "acknowledged")

		// 测试趋势方向枚举
		t.Assert(types.TrendDirectionIncreasing.String(), "increasing")
		t.Assert(types.TrendDirectionDecreasing.String(), "decreasing")
		t.Assert(types.TrendDirectionStable.String(), "stable")

		// 测试时间周期枚举
		t.Assert(types.TimePeriodDaily.String(), "daily")
		t.Assert(types.TimePeriodWeekly.String(), "weekly")
		t.Assert(types.TimePeriodMonthly.String(), "monthly")
	})
}

// TestRightsBalanceLogic 测试权益余额逻辑
func TestRightsBalanceLogic(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		balance := &types.RightsBalance{
			TotalBalance:  10000.0,
			UsedBalance:   3000.0,
			FrozenBalance: 1000.0,
		}

		// 测试可用余额计算
		available := balance.GetAvailableBalance()
		expected := 6000.0 // 10000 - 3000 - 1000
		t.Assert(available, expected)

		// 测试使用率计算
		rate := balance.GetUsageRate()
		expectedRate := 0.3 // 3000 / 10000
		t.Assert(rate, expectedRate)
	})
}

// TestAlertConfigurationValidation 测试预警配置验证
func TestAlertConfigurationValidation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 测试有效配置
		config := &types.AlertConfigureRequest{
			MerchantID:        1,
			WarningThreshold:  &[]float64{1000.0}[0],
			CriticalThreshold: &[]float64{500.0}[0],
		}

		err := config.Validate()
		t.AssertNil(err)

		// 测试无效商户ID
		invalidConfig := &types.AlertConfigureRequest{
			MerchantID: 0,
		}

		err = invalidConfig.Validate()
		t.AssertNE(err, nil)
	})
}

// TestMonitoringDataStructures 测试监控数据结构
func TestMonitoringDataStructures(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 测试预警结构
		alert := &types.RightsAlert{
			MerchantID:     1,
			AlertType:      types.AlertTypeBalanceLow,
			ThresholdValue: 1000.0,
			CurrentValue:   800.0,
			Severity:       types.AlertSeverityWarning,
			Status:         types.AlertStatusActive,
			Message:        "余额不足预警",
		}

		t.Assert(alert.MerchantID, uint64(1))
		t.Assert(alert.AlertType, types.AlertTypeBalanceLow)
		t.Assert(alert.Severity, types.AlertSeverityWarning)

		// 测试仪表板数据结构
		dashboard := &types.MonitoringDashboardData{
			TotalMerchants:    5,
			ActiveAlerts:      3,
			TotalRightsBalance: 50000.0,
			DailyConsumption:  1500.0,
		}

		t.Assert(dashboard.TotalMerchants, 5)
		t.Assert(dashboard.ActiveAlerts, 3)
		t.Assert(dashboard.TotalRightsBalance, 50000.0)
	})
}
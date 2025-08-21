package test

import (
	"testing"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"
	"mer-demo/shared/types"
)

// TestAlertTypeString 测试预警类型字符串转换
func TestAlertTypeString(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		testCases := map[types.AlertType]string{
			types.AlertTypeBalanceLow:         "balance_low",
			types.AlertTypeBalanceCritical:    "balance_critical",
			types.AlertTypeUsageSpike:         "usage_spike",
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

// TestTimePeriodString 测试时间周期字符串转换
func TestTimePeriodString(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		testCases := map[types.TimePeriod]string{
			types.TimePeriodDaily:   "daily",
			types.TimePeriodWeekly:  "weekly",
			types.TimePeriodMonthly: "monthly",
		}

		for period, expected := range testCases {
			t.Assert(period.String(), expected)
		}
	})
}

// TestAlertStatusString 测试预警状态字符串转换
func TestAlertStatusString(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		testCases := map[types.AlertStatus]string{
			types.AlertStatusActive:       "active",
			types.AlertStatusResolved:     "resolved",
			types.AlertStatusAcknowledged: "acknowledged",
		}

		for status, expected := range testCases {
			t.Assert(status.String(), expected)
		}
	})
}

// TestRightsBalanceCalculation 测试权益余额计算
func TestRightsBalanceCalculation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		balance := &types.RightsBalance{
			TotalBalance:  10000.0,
			UsedBalance:   3000.0,
			FrozenBalance: 1000.0,
		}

		// 测试可用余额计算
		available := balance.GetAvailableBalance()
		expected := 10000.0 - 3000.0 - 1000.0 // 6000.0
		t.Assert(available, expected)
	})
}

// TestAlertValidation 测试预警验证
func TestAlertValidation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 测试有效的预警配置请求
		validRequest := &types.AlertConfigureRequest{
			MerchantID:        1,
			WarningThreshold:  ptrOf(1000.0),
			CriticalThreshold: ptrOf(500.0),
		}
		t.AssertNil(validRequest.Validate())

		// 测试无效的商户ID
		invalidRequest := &types.AlertConfigureRequest{
			MerchantID:        0,
			WarningThreshold:  ptrOf(1000.0),
			CriticalThreshold: ptrOf(500.0),
		}
		t.AssertNE(invalidRequest.Validate(), nil)
	})
}

// TestReportRequestValidation 测试报告请求验证
func TestReportRequestValidation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 测试有效的报告请求
		validRequest := &types.ReportGenerateRequest{
			Period:      types.TimePeriodDaily,
			StartDate:   gtime.Now().AddDate(0, 0, -7).Time,
			EndDate:     gtime.Now().Time,
			MerchantIDs: []uint64{1, 2},
			Format:      "excel",
		}
		t.AssertNil(validRequest.Validate())

		// 测试无效的时间范围
		invalidRequest := &types.ReportGenerateRequest{
			Period:    types.TimePeriodDaily,
			StartDate: gtime.Now().Time,
			EndDate:   gtime.Now().AddDate(0, 0, -7).Time, // 开始时间晚于结束时间
			Format:    "excel",
		}
		t.AssertNE(invalidRequest.Validate(), nil)
	})
}

// ptrOf helper function for creating pointers
func ptrOf[T any](v T) *T {
	return &v
}
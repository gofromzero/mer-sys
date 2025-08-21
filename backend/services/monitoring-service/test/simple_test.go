package test

import (
	"testing"

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
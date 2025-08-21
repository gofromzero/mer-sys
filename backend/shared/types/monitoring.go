package types

import (
	"encoding/json"
	"errors"
	"time"
)

// AlertType 预警类型
type AlertType int

const (
	AlertTypeBalanceLow AlertType = iota + 1
	AlertTypeBalanceCritical
	AlertTypeUsageSpike
	AlertTypePredictedDepletion
)

// String returns the string representation of AlertType
func (at AlertType) String() string {
	switch at {
	case AlertTypeBalanceLow:
		return "balance_low"
	case AlertTypeBalanceCritical:
		return "balance_critical"
	case AlertTypeUsageSpike:
		return "usage_spike"
	case AlertTypePredictedDepletion:
		return "predicted_depletion"
	default:
		return "unknown"
	}
}

// AlertSeverity 预警严重程度
type AlertSeverity int

const (
	AlertSeverityInfo AlertSeverity = iota + 1
	AlertSeverityWarning
	AlertSeverityCritical
)

// String returns the string representation of AlertSeverity
func (as AlertSeverity) String() string {
	switch as {
	case AlertSeverityInfo:
		return "info"
	case AlertSeverityWarning:
		return "warning"
	case AlertSeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// AlertStatus 预警状态
type AlertStatus int

const (
	AlertStatusActive AlertStatus = iota + 1
	AlertStatusResolved
	AlertStatusIgnored
)

// String returns the string representation of AlertStatus
func (as AlertStatus) String() string {
	switch as {
	case AlertStatusActive:
		return "active"
	case AlertStatusResolved:
		return "resolved"
	case AlertStatusIgnored:
		return "ignored"
	default:
		return "unknown"
	}
}

// TimePeriod 时间周期
type TimePeriod int

const (
	TimePeriodDaily TimePeriod = iota + 1
	TimePeriodWeekly
	TimePeriodMonthly
)

// String returns the string representation of TimePeriod
func (tp TimePeriod) String() string {
	switch tp {
	case TimePeriodDaily:
		return "daily"
	case TimePeriodWeekly:
		return "weekly"
	case TimePeriodMonthly:
		return "monthly"
	default:
		return "unknown"
	}
}

// TrendDirection 趋势方向
type TrendDirection int

const (
	TrendDirectionStable TrendDirection = iota
	TrendDirectionIncreasing
	TrendDirectionDecreasing
)

// String returns the string representation of TrendDirection
func (td TrendDirection) String() string {
	switch td {
	case TrendDirectionStable:
		return "stable"
	case TrendDirectionIncreasing:
		return "increasing"
	case TrendDirectionDecreasing:
		return "decreasing"
	default:
		return "unknown"
	}
}

// RightsAlert 权益预警实体
type RightsAlert struct {
	ID               uint64         `json:"id" db:"id"`
	TenantID         uint64         `json:"tenant_id" db:"tenant_id"`
	MerchantID       uint64         `json:"merchant_id" db:"merchant_id"`
	AlertType        AlertType      `json:"alert_type" db:"alert_type"`
	ThresholdValue   float64        `json:"threshold_value" db:"threshold_value"`
	CurrentValue     float64        `json:"current_value" db:"current_value"`
	Severity         AlertSeverity  `json:"severity" db:"severity"`
	Status           AlertStatus    `json:"status" db:"status"`
	Message          string         `json:"message" db:"message"`
	TriggeredAt      time.Time      `json:"triggered_at" db:"triggered_at"`
	ResolvedAt       *time.Time     `json:"resolved_at,omitempty" db:"resolved_at"`
	NotifiedChannels []string       `json:"notified_channels" db:"notified_channels"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
}

// MarshalJSON 自定义JSON序列化，处理NotifiedChannels
func (ra *RightsAlert) MarshalJSON() ([]byte, error) {
	type Alias RightsAlert
	return json.Marshal(&struct {
		*Alias
		NotifiedChannels json.RawMessage `json:"notified_channels"`
	}{
		Alias: (*Alias)(ra),
		NotifiedChannels: func() json.RawMessage {
			if ra.NotifiedChannels != nil {
				data, _ := json.Marshal(ra.NotifiedChannels)
				return data
			}
			return json.RawMessage("[]")
		}(),
	})
}

// RightsUsageStats 权益使用统计
type RightsUsageStats struct {
	ID                    uint64          `json:"id" db:"id"`
	TenantID              uint64          `json:"tenant_id" db:"tenant_id"`
	MerchantID            *uint64         `json:"merchant_id,omitempty" db:"merchant_id"`
	StatDate              time.Time       `json:"stat_date" db:"stat_date"`
	Period                TimePeriod      `json:"period" db:"period_type"`
	TotalAllocated        float64         `json:"total_allocated" db:"total_allocated"`
	TotalConsumed         float64         `json:"total_consumed" db:"total_consumed"`
	AverageDailyUsage     float64         `json:"average_daily_usage" db:"average_daily_usage"`
	PeakUsageDay          *time.Time      `json:"peak_usage_day,omitempty" db:"peak_usage_day"`
	PredictedDepletionDate *time.Time     `json:"predicted_depletion_date,omitempty" db:"predicted_depletion_date"`
	UsageTrend            TrendDirection  `json:"usage_trend" db:"usage_trend"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
}

// AlertConfigureRequest 预警配置请求
type AlertConfigureRequest struct {
	MerchantID        uint64   `json:"merchant_id" binding:"required"`
	WarningThreshold  *float64 `json:"warning_threshold" binding:"omitempty,gte=0"`
	CriticalThreshold *float64 `json:"critical_threshold" binding:"omitempty,gte=0"`
}

// AlertListQuery 预警列表查询参数
type AlertListQuery struct {
	Page       int           `form:"page,default=1" binding:"min=1"`
	PageSize   int           `form:"page_size,default=20" binding:"min=1,max=100"`
	MerchantID *uint64       `form:"merchant_id,omitempty"`
	AlertType  *AlertType    `form:"alert_type,omitempty"`
	Severity   *AlertSeverity `form:"severity,omitempty"`
	Status     *AlertStatus  `form:"status,omitempty"`
	StartTime  *time.Time    `form:"start_time,omitempty"`
	EndTime    *time.Time    `form:"end_time,omitempty"`
}

// AlertResolveRequest 预警解决请求
type AlertResolveRequest struct {
	Resolution string `json:"resolution" binding:"required,max=500"`
}

// RightsStatsQuery 权益统计查询参数
type RightsStatsQuery struct {
	MerchantID *uint64     `form:"merchant_id,omitempty"`
	Period     *TimePeriod `form:"period,omitempty"`
	StartDate  *time.Time  `form:"start_date,omitempty"`
	EndDate    *time.Time  `form:"end_date,omitempty"`
}

// RightsTrendsQuery 权益趋势查询参数
type RightsTrendsQuery struct {
	MerchantID *uint64     `form:"merchant_id,omitempty"`
	Period     *TimePeriod `form:"period,omitempty"`
	Days       *int        `form:"days,default=30" binding:"omitempty,min=1,max=365"`
}

// MonitoringDashboardData 监控仪表板数据
type MonitoringDashboardData struct {
	TotalMerchants     int                   `json:"total_merchants"`
	ActiveAlerts       int                   `json:"active_alerts"`
	CriticalAlerts     int                   `json:"critical_alerts"`
	TotalRightsBalance float64               `json:"total_rights_balance"`
	AvgDailyUsage      float64               `json:"avg_daily_usage"`
	TopMerchantsByUsage []MerchantUsageInfo  `json:"top_merchants_by_usage"`
	RecentAlerts       []RightsAlert         `json:"recent_alerts"`
	UsageTrends        []DailyUsageTrend     `json:"usage_trends"`
}

// MerchantUsageInfo 商户使用信息
type MerchantUsageInfo struct {
	MerchantID   uint64  `json:"merchant_id"`
	MerchantName string  `json:"merchant_name"`
	DailyUsage   float64 `json:"daily_usage"`
	TotalBalance float64 `json:"total_balance"`
}

// DailyUsageTrend 日使用趋势
type DailyUsageTrend struct {
	Date       time.Time `json:"date"`
	TotalUsage float64   `json:"total_usage"`
}

// ReportGenerateRequest 报告生成请求
type ReportGenerateRequest struct {
	MerchantIDs []uint64   `json:"merchant_ids,omitempty"`
	StartDate   time.Time  `json:"start_date" binding:"required"`
	EndDate     time.Time  `json:"end_date" binding:"required"`
	Period      TimePeriod `json:"period" binding:"required"`
	Format      string     `json:"format,default=pdf" binding:"oneof=pdf xlsx csv"`
}

// Validate validates AlertConfigureRequest
func (req *AlertConfigureRequest) Validate() error {
	if req.WarningThreshold != nil && req.CriticalThreshold != nil {
		if *req.CriticalThreshold >= *req.WarningThreshold {
			return errors.New("critical threshold must be less than warning threshold")
		}
	}
	return nil
}

// Validate validates ReportGenerateRequest
func (req *ReportGenerateRequest) Validate() error {
	if req.EndDate.Before(req.StartDate) {
		return errors.New("end date must be after start date")
	}
	maxDays := 365 // Maximum 1 year
	if req.EndDate.Sub(req.StartDate).Hours()/24 > float64(maxDays) {
		return errors.New("date range cannot exceed 1 year")
	}
	return nil
}

// 监控相关错误定义
var (
	ErrAlertNotFound        = errors.New("alert not found")
	ErrAlertAlreadyResolved = errors.New("alert already resolved")
	ErrInvalidThreshold     = errors.New("invalid threshold configuration")
	ErrInvalidDateRange     = errors.New("invalid date range")
)
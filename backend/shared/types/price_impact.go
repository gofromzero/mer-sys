package types

import (
	"time"
)

// PriceImpactAssessment 价格变更影响评估
type PriceImpactAssessment struct {
	ProductID             uint64                    `json:"product_id"`
	OldPrice              Money                     `json:"old_price"`
	NewPrice              Money                     `json:"new_price"`
	ChangePercentage      float64                   `json:"change_percentage"`
	EffectiveDate         time.Time                 `json:"effective_date"`
	ImpactedOrders        PriceImpactedOrders       `json:"impacted_orders"`
	RevenueImpact         RevenueImpactAnalysis     `json:"revenue_impact"`
	InventoryImpact       InventoryImpactAnalysis   `json:"inventory_impact"`
	CustomerImpact        CustomerImpactAnalysis    `json:"customer_impact"`
	RecommendedActions    []string                  `json:"recommended_actions"`
	RiskLevel            ImpactRiskLevel           `json:"risk_level"`
	AssessmentTimestamp  time.Time                 `json:"assessment_timestamp"`
}

// ImpactRiskLevel 影响风险等级
type ImpactRiskLevel string

const (
	ImpactRiskLevelLow    ImpactRiskLevel = "low"    // 低风险
	ImpactRiskLevelMedium ImpactRiskLevel = "medium" // 中等风险
	ImpactRiskLevelHigh   ImpactRiskLevel = "high"   // 高风险
)

// PriceImpactedOrders 受价格变更影响的订单
type PriceImpactedOrders struct {
	PendingOrdersCount     int       `json:"pending_orders_count"`     // 待处理订单数量
	PendingOrdersValue     Money     `json:"pending_orders_value"`     // 待处理订单总价值
	ActiveCartCount        int       `json:"active_cart_count"`        // 活跃购物车数量
	ActiveCartValue        Money     `json:"active_cart_value"`        // 活跃购物车总价值
	RecentOrdersCount      int       `json:"recent_orders_count"`      // 最近订单数量（7天）
	RecentOrdersValue      Money     `json:"recent_orders_value"`      // 最近订单总价值
	AffectedCustomers      int       `json:"affected_customers"`       // 受影响客户数量
	RequiresNotification   bool      `json:"requires_notification"`    // 是否需要通知客户
}

// RevenueImpactAnalysis 收入影响分析
type RevenueImpactAnalysis struct {
	ProjectedDailyRevenue     Money   `json:"projected_daily_revenue"`     // 预计日收入变化
	ProjectedWeeklyRevenue    Money   `json:"projected_weekly_revenue"`    // 预计周收入变化
	ProjectedMonthlyRevenue   Money   `json:"projected_monthly_revenue"`   // 预计月收入变化
	HistoricalSalesVolume     int     `json:"historical_sales_volume"`     // 历史销量（30天）
	PredictedVolumeChange     float64 `json:"predicted_volume_change"`     // 预计销量变化百分比
	BreakEvenPoint           int     `json:"break_even_point"`            // 盈亏平衡点（天数）
	RevenueRiskScore         float64 `json:"revenue_risk_score"`          // 收入风险评分（0-100）
}

// InventoryImpactAnalysis 库存影响分析
type InventoryImpactAnalysis struct {
	CurrentStock              int     `json:"current_stock"`               // 当前库存
	AverageMonthlyTurnover    int     `json:"average_monthly_turnover"`    // 月平均周转
	PredictedTurnoverChange   float64 `json:"predicted_turnover_change"`   // 预计周转变化
	OverstockRisk            float64 `json:"overstock_risk"`              // 滞销风险评分
	UnderstockRisk           float64 `json:"understock_risk"`             // 缺货风险评分
	RecommendedStockLevel     int     `json:"recommended_stock_level"`     // 建议库存水平
	InventoryValue           Money   `json:"inventory_value"`             // 库存价值变化
}

// CustomerImpactAnalysis 客户影响分析
type CustomerImpactAnalysis struct {
	PriceElasticity          float64 `json:"price_elasticity"`           // 价格弹性
	CustomerSegmentImpact    []CustomerSegmentImpact `json:"customer_segment_impact"` // 客户群体影响
	CompetitivePosition      CompetitiveAnalysis     `json:"competitive_position"`    // 竞争地位分析
	ChurnRisk               float64 `json:"churn_risk"`                 // 客户流失风险
	AcquisitionImpact       float64 `json:"acquisition_impact"`         // 获客影响
	SatisfactionImpact      float64 `json:"satisfaction_impact"`        // 满意度影响
}

// CustomerSegmentImpact 客户群体影响
type CustomerSegmentImpact struct {
	SegmentName       string  `json:"segment_name"`        // 群体名称
	CustomerCount     int     `json:"customer_count"`      // 客户数量
	ImpactScore       float64 `json:"impact_score"`        // 影响评分
	ExpectedBehavior  string  `json:"expected_behavior"`   // 预期行为
	MitigationNeeded  bool    `json:"mitigation_needed"`   // 是否需要缓解措施
}

// CompetitiveAnalysis 竞争分析
type CompetitiveAnalysis struct {
	MarketPosition        string  `json:"market_position"`         // 市场地位
	CompetitorPriceGap    float64 `json:"competitor_price_gap"`    // 与竞品价格差距
	CompetitiveAdvantage  bool    `json:"competitive_advantage"`   // 是否具有竞争优势
	MarketShareRisk       float64 `json:"market_share_risk"`       // 市场份额风险
}

// PriceImpactAssessmentRequest 价格影响评估请求
type PriceImpactAssessmentRequest struct {
	ProductID       uint64    `json:"product_id" validate:"required"`
	NewPrice        Money     `json:"new_price" validate:"required"`
	EffectiveDate   time.Time `json:"effective_date" validate:"required"`
	IncludeHistory  bool      `json:"include_history,omitempty"`        // 是否包含历史数据分析
	IncludePredict  bool      `json:"include_predict,omitempty"`        // 是否包含预测分析
	AssessmentPeriod int      `json:"assessment_period,omitempty"`      // 评估周期（天），默认30天
}

// PriceChangeRecommendation 价格变更建议
type PriceChangeRecommendation struct {
	RecommendationType    RecommendationType `json:"recommendation_type"`
	RecommendedPrice      Money              `json:"recommended_price,omitempty"`
	RecommendedTiming     time.Time          `json:"recommended_timing,omitempty"`
	Reasoning            string             `json:"reasoning"`
	ExpectedOutcome      string             `json:"expected_outcome"`
	ImplementationSteps  []string           `json:"implementation_steps"`
	MonitoringMetrics    []string           `json:"monitoring_metrics"`
	RollbackPlan         string             `json:"rollback_plan,omitempty"`
	ConfidenceLevel      float64            `json:"confidence_level"`
}

// RecommendationType 建议类型
type RecommendationType string

const (
	RecommendationTypeApprove    RecommendationType = "approve"     // 批准执行
	RecommendationTypeReject     RecommendationType = "reject"      // 建议拒绝
	RecommendationTypeModify     RecommendationType = "modify"      // 建议修改
	RecommendationTypeDelay      RecommendationType = "delay"       // 建议延迟
	RecommendationTypeGradual    RecommendationType = "gradual"     // 建议分阶段执行
)

// PriceAuditEvent 价格审计事件
type PriceAuditEvent struct {
	ID              uint64                 `json:"id" db:"id"`
	TenantID        uint64                 `json:"tenant_id" db:"tenant_id"`
	EventType       PriceAuditEventType    `json:"event_type" db:"event_type"`
	ProductID       uint64                 `json:"product_id" db:"product_id"`
	UserID          uint64                 `json:"user_id" db:"user_id"`
	EventData       PriceAuditEventData    `json:"event_data" db:"event_data"`
	ImpactLevel     ImpactRiskLevel        `json:"impact_level" db:"impact_level"`
	Description     string                 `json:"description" db:"description"`
	ClientIP        string                 `json:"client_ip,omitempty" db:"client_ip"`
	UserAgent       string                 `json:"user_agent,omitempty" db:"user_agent"`
	SessionID       string                 `json:"session_id,omitempty" db:"session_id"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// PriceAuditEventType 价格审计事件类型
type PriceAuditEventType string

const (
	PriceAuditEventTypeCreate       PriceAuditEventType = "price_rule_create"      // 创建定价规则
	PriceAuditEventTypeUpdate       PriceAuditEventType = "price_rule_update"      // 更新定价规则
	PriceAuditEventTypeDelete       PriceAuditEventType = "price_rule_delete"      // 删除定价规则
	PriceAuditEventTypePriceChange  PriceAuditEventType = "price_change"           // 价格变更
	PriceAuditEventTypePromoCreate  PriceAuditEventType = "promotion_create"       // 创建促销
	PriceAuditEventTypePromoUpdate  PriceAuditEventType = "promotion_update"       // 更新促销
	PriceAuditEventTypePromoDelete  PriceAuditEventType = "promotion_delete"       // 删除促销
	PriceAuditEventTypeRightsCreate PriceAuditEventType = "rights_rule_create"     // 创建权益规则
	PriceAuditEventTypeRightsUpdate PriceAuditEventType = "rights_rule_update"     // 更新权益规则
	PriceAuditEventTypeAccessDenied PriceAuditEventType = "access_denied"          // 访问被拒绝
	PriceAuditEventTypeSuspicious   PriceAuditEventType = "suspicious_activity"    // 可疑活动
)

// PriceAuditEventData 价格审计事件数据
type PriceAuditEventData struct {
	OldValue        interface{} `json:"old_value,omitempty"`
	NewValue        interface{} `json:"new_value,omitempty"`
	ChangeReason    string      `json:"change_reason,omitempty"`
	ApprovalStatus  string      `json:"approval_status,omitempty"`
	ApprovedBy      uint64      `json:"approved_by,omitempty"`
	ImpactAssessment *PriceImpactAssessment `json:"impact_assessment,omitempty"`
	ValidationErrors []string   `json:"validation_errors,omitempty"`
	AdditionalData  map[string]interface{} `json:"additional_data,omitempty"`
}

// TableName 指定表名
func (PriceAuditEvent) TableName() string {
	return "price_audit_events"
}

// CreatePriceAuditEventRequest 创建价格审计事件请求
type CreatePriceAuditEventRequest struct {
	EventType     PriceAuditEventType    `json:"event_type" validate:"required"`
	ProductID     uint64                 `json:"product_id" validate:"required"`
	EventData     PriceAuditEventData    `json:"event_data"`
	ImpactLevel   ImpactRiskLevel        `json:"impact_level"`
	Description   string                 `json:"description" validate:"required,max=500"`
	ClientIP      string                 `json:"client_ip,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
}

// PriceAuditQuery 价格审计查询
type PriceAuditQuery struct {
	StartDate     time.Time               `json:"start_date,omitempty"`
	EndDate       time.Time               `json:"end_date,omitempty"`
	EventTypes    []PriceAuditEventType   `json:"event_types,omitempty"`
	ProductIDs    []uint64                `json:"product_ids,omitempty"`
	UserIDs       []uint64                `json:"user_ids,omitempty"`
	ImpactLevels  []ImpactRiskLevel       `json:"impact_levels,omitempty"`
	Page          int                     `json:"page" validate:"min=1"`
	PageSize      int                     `json:"page_size" validate:"min=1,max=100"`
	OrderBy       string                  `json:"order_by,omitempty"`
	OrderDir      string                  `json:"order_dir,omitempty" validate:"oneof=asc desc"`
}

// PriceAuditReport 价格审计报告
type PriceAuditReport struct {
	ReportID        string                    `json:"report_id"`
	GeneratedAt     time.Time                 `json:"generated_at"`
	ReportPeriod    ReportPeriod              `json:"report_period"`
	Summary         PriceAuditSummary         `json:"summary"`
	EventAnalysis   PriceEventAnalysis        `json:"event_analysis"`
	RiskAnalysis    PriceRiskAnalysis         `json:"risk_analysis"`
	TrendAnalysis   PriceTrendAnalysis        `json:"trend_analysis"`
	Recommendations []PriceChangeRecommendation `json:"recommendations"`
	Anomalies       []PriceAnomalyDetection   `json:"anomalies"`
}

// ReportPeriod 报告周期
type ReportPeriod struct {
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	PeriodType  string    `json:"period_type"` // daily, weekly, monthly, custom
}

// PriceAuditSummary 价格审计摘要
type PriceAuditSummary struct {
	TotalEvents           int                            `json:"total_events"`
	EventsByType          map[PriceAuditEventType]int    `json:"events_by_type"`
	EventsByRiskLevel     map[ImpactRiskLevel]int        `json:"events_by_risk_level"`
	UniqueProductsAffected int                           `json:"unique_products_affected"`
	UniqueUsersInvolved    int                           `json:"unique_users_involved"`
	TotalRevenueImpact     Money                         `json:"total_revenue_impact"`
	ComplianceScore       float64                        `json:"compliance_score"`
}

// PriceEventAnalysis 价格事件分析
type PriceEventAnalysis struct {
	MostFrequentEventType  PriceAuditEventType `json:"most_frequent_event_type"`
	PeakActivityHour       int                 `json:"peak_activity_hour"`
	AverageEventsPerDay    float64             `json:"average_events_per_day"`
	EventDistribution      []EventDistribution `json:"event_distribution"`
	UserActivityPattern    []UserActivity      `json:"user_activity_pattern"`
}

// EventDistribution 事件分布
type EventDistribution struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
	Type  string    `json:"type,omitempty"`
}

// UserActivity 用户活动
type UserActivity struct {
	UserID      uint64 `json:"user_id"`
	EventCount  int    `json:"event_count"`
	RiskScore   float64 `json:"risk_score"`
	LastActivity time.Time `json:"last_activity"`
}

// PriceRiskAnalysis 价格风险分析
type PriceRiskAnalysis struct {
	OverallRiskScore      float64                 `json:"overall_risk_score"`
	RiskFactors          []PriceRiskFactor       `json:"risk_factors"`
	HighRiskProducts     []uint64                `json:"high_risk_products"`
	SuspiciousActivities []SuspiciousActivity    `json:"suspicious_activities"`
	ComplianceIssues     []ComplianceIssue       `json:"compliance_issues"`
}

// PriceRiskFactor 价格风险因子
type PriceRiskFactor struct {
	FactorName    string  `json:"factor_name"`
	RiskLevel     ImpactRiskLevel `json:"risk_level"`
	Impact        float64 `json:"impact"`
	Description   string  `json:"description"`
	Mitigation    string  `json:"mitigation"`
}

// SuspiciousActivity 可疑活动
type SuspiciousActivity struct {
	ActivityType  string    `json:"activity_type"`
	Description   string    `json:"description"`
	RiskScore     float64   `json:"risk_score"`
	DetectedAt    time.Time `json:"detected_at"`
	UserID        uint64    `json:"user_id"`
	ProductID     uint64    `json:"product_id"`
	ActionTaken   string    `json:"action_taken,omitempty"`
}

// ComplianceIssue 合规问题
type ComplianceIssue struct {
	IssueType     string              `json:"issue_type"`
	Severity      ImpactRiskLevel     `json:"severity"`
	Description   string              `json:"description"`
	Recommendation string             `json:"recommendation"`
	DetectedAt    time.Time           `json:"detected_at"`
	Status        string              `json:"status"`
}

// PriceTrendAnalysis 价格趋势分析
type PriceTrendAnalysis struct {
	AveragePriceChange    float64           `json:"average_price_change"`
	PriceVolatility      float64           `json:"price_volatility"`
	TrendDirection       string            `json:"trend_direction"` // up, down, stable
	SeasonalPatterns     []SeasonalPattern `json:"seasonal_patterns"`
	PredictedTrends      []TrendPrediction `json:"predicted_trends"`
}

// SeasonalPattern 季节性模式
type SeasonalPattern struct {
	Pattern     string  `json:"pattern"`
	Strength    float64 `json:"strength"`
	Period      string  `json:"period"`
	Description string  `json:"description"`
}

// TrendPrediction 趋势预测
type TrendPrediction struct {
	Date            time.Time `json:"date"`
	PredictedChange float64   `json:"predicted_change"`
	Confidence      float64   `json:"confidence"`
	Factors         []string  `json:"factors"`
}

// PriceAnomalyDetection 价格异常检测
type PriceAnomalyDetection struct {
	AnomalyID      string              `json:"anomaly_id"`
	DetectedAt     time.Time           `json:"detected_at"`
	AnomalyType    PriceAnomalyType    `json:"anomaly_type"`
	ProductID      uint64              `json:"product_id"`
	AnomalyScore   float64             `json:"anomaly_score"`
	Description    string              `json:"description"`
	CurrentValue   interface{}         `json:"current_value"`
	ExpectedValue  interface{}         `json:"expected_value"`
	Deviation      float64             `json:"deviation"`
	Severity       ImpactRiskLevel     `json:"severity"`
	AutoResolved   bool                `json:"auto_resolved"`
	Resolution     string              `json:"resolution,omitempty"`
}

// PriceAnomalyType 价格异常类型
type PriceAnomalyType string

const (
	PriceAnomalyTypeUnusualIncrease  PriceAnomalyType = "unusual_increase"   // 异常涨价
	PriceAnomalyTypeUnusualDecrease  PriceAnomalyType = "unusual_decrease"   // 异常降价
	PriceAnomalyTypeFrequentChanges  PriceAnomalyType = "frequent_changes"   // 频繁变价
	PriceAnomalyTypeNegativePrice    PriceAnomalyType = "negative_price"     // 负价格
	PriceAnomalyTypeZeroPrice        PriceAnomalyType = "zero_price"         // 零价格
	PriceAnomalyTypeOutlierPrice     PriceAnomalyType = "outlier_price"      // 异常价格
	PriceAnomalyTypeInconsistentRule PriceAnomalyType = "inconsistent_rule"  // 规则不一致
)
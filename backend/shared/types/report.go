package types

import (
	"encoding/json"
	"time"
)

// ReportType 报表类型
type ReportType string

const (
	ReportTypeFinancial         ReportType = "financial"          // 财务报表
	ReportTypeMerchantOperation ReportType = "merchant_operation" // 商户运营报表
	ReportTypeCustomerAnalysis  ReportType = "customer_analysis"  // 客户分析报表
)

// PeriodType 时间周期类型
type PeriodType string

const (
	PeriodTypeDaily     PeriodType = "daily"     // 日报
	PeriodTypeWeekly    PeriodType = "weekly"    // 周报
	PeriodTypeMonthly   PeriodType = "monthly"   // 月报
	PeriodTypeQuarterly PeriodType = "quarterly" // 季报
	PeriodTypeYearly    PeriodType = "yearly"    // 年报
	PeriodTypeCustom    PeriodType = "custom"    // 自定义
)

// ReportStatus 报表状态
type ReportStatus string

const (
	ReportStatusGenerating ReportStatus = "generating" // 生成中
	ReportStatusCompleted  ReportStatus = "completed"  // 已完成
	ReportStatusFailed     ReportStatus = "failed"     // 失败
)

// FileFormat 文件格式
type FileFormat string

const (
	FileFormatExcel FileFormat = "excel" // Excel格式
	FileFormatPDF   FileFormat = "pdf"   // PDF格式
	FileFormatJSON  FileFormat = "json"  // JSON格式
)

// Report 报表实体
type Report struct {
	ID          uint64          `gorm:"primary_key;auto_increment" json:"id"`
	UUID        string          `gorm:"type:char(36);unique_index" json:"uuid"`
	TenantID    uint64          `gorm:"not null;index" json:"tenant_id"`
	ReportType  ReportType      `gorm:"not null;index" json:"report_type"`
	PeriodType  PeriodType      `gorm:"not null" json:"period_type"`
	StartDate   time.Time       `gorm:"not null;index" json:"start_date"`
	EndDate     time.Time       `gorm:"not null;index" json:"end_date"`
	Status      ReportStatus    `gorm:"not null;index;default:'generating'" json:"status"`
	FilePath    string          `gorm:"type:varchar(500)" json:"file_path,omitempty"`
	FileFormat  FileFormat      `gorm:"not null" json:"file_format"`
	GeneratedBy uint64          `gorm:"not null" json:"generated_by"`
	GeneratedAt time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"generated_at"`
	ExpiresAt   *time.Time      `json:"expires_at,omitempty"`
	DataSummary json.RawMessage `gorm:"type:json" json:"data_summary,omitempty"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// ReportTemplate 报表模板
type ReportTemplate struct {
	ID             uint64          `gorm:"primary_key;auto_increment" json:"id"`
	TenantID       uint64          `gorm:"not null;index" json:"tenant_id"`
	Name           string          `gorm:"type:varchar(100);not null" json:"name"`
	ReportType     ReportType      `gorm:"not null;index" json:"report_type"`
	TemplateConfig json.RawMessage `gorm:"type:json;not null" json:"template_config"`
	ScheduleConfig json.RawMessage `gorm:"type:json" json:"schedule_config,omitempty"`
	Recipients     json.RawMessage `gorm:"type:json" json:"recipients,omitempty"`
	FileFormat     FileFormat      `gorm:"not null;default:'excel'" json:"file_format"`
	Enabled        bool            `gorm:"default:true" json:"enabled"`
	CreatedBy      uint64          `gorm:"not null" json:"created_by"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// ReportJob 报表生成任务
type ReportJob struct {
	ID           uint64       `gorm:"primary_key;auto_increment" json:"id"`
	TenantID     uint64       `gorm:"not null;index" json:"tenant_id"`
	TemplateID   *uint64      `json:"template_id,omitempty"`
	ReportID     *uint64      `json:"report_id,omitempty"`
	Status       JobStatus    `gorm:"not null;index;default:'pending'" json:"status"`
	ScheduledAt  time.Time    `gorm:"not null;index" json:"scheduled_at"`
	StartedAt    *time.Time   `json:"started_at,omitempty"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
	ErrorMessage string       `gorm:"type:text" json:"error_message,omitempty"`
	RetryCount   int          `gorm:"default:0" json:"retry_count"`
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"created_at"`
}

// JobStatus 任务状态
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"   // 待处理
	JobStatusRunning   JobStatus = "running"   // 运行中
	JobStatusCompleted JobStatus = "completed" // 已完成
	JobStatusFailed    JobStatus = "failed"    // 失败
)

// AnalyticsCache 数据分析缓存
type AnalyticsCache struct {
	ID         uint64          `gorm:"primary_key;auto_increment" json:"id"`
	TenantID   uint64          `gorm:"not null;index" json:"tenant_id"`
	CacheKey   string          `gorm:"type:varchar(200);not null;uniqueIndex:uk_cache_key" json:"cache_key"`
	MetricType string          `gorm:"type:varchar(50);not null" json:"metric_type"`
	TimePeriod string          `gorm:"type:varchar(20);not null" json:"time_period"`
	Data       json.RawMessage `gorm:"type:json;not null" json:"data"`
	ExpiresAt  time.Time       `gorm:"not null;index" json:"expires_at"`
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`
}

// FinancialReportData 财务报表数据
type FinancialReportData struct {
	TotalRevenue         Money                    `json:"total_revenue"`          // 总收入
	TotalExpenditure     Money                    `json:"total_expenditure"`      // 总支出
	NetProfit            Money                    `json:"net_profit"`             // 净利润
	RightsDistributed    int64                    `json:"rights_distributed"`     // 权益发放总量
	RightsConsumed       int64                    `json:"rights_consumed"`        // 权益消耗总量
	RightsBalance        int64                    `json:"rights_balance"`         // 权益余额
	MerchantCount        int                      `json:"merchant_count"`         // 商户总数
	ActiveMerchantCount  int                      `json:"active_merchant_count"`  // 活跃商户数
	CustomerCount        int                      `json:"customer_count"`         // 客户总数
	ActiveCustomerCount  int                      `json:"active_customer_count"`  // 活跃客户数
	OrderCount           int                      `json:"order_count"`            // 订单总数
	OrderAmount          Money                    `json:"order_amount"`           // 订单总金额
	Breakdown            *FinancialBreakdown      `json:"breakdown,omitempty"`    // 详细分解数据
}

// FinancialBreakdown 财务分解数据
type FinancialBreakdown struct {
	RevenueByMerchant    []MerchantRevenue    `json:"revenue_by_merchant"`    // 按商户收入
	RevenueByCategory    []CategoryRevenue    `json:"revenue_by_category"`    // 按类别收入
	ExpenditureByType    []ExpenditureItem    `json:"expenditure_by_type"`    // 按类型支出
	RightsByCategory     []RightsUsage        `json:"rights_by_category"`     // 按类别权益使用
	MonthlyTrend         []MonthlyFinancial   `json:"monthly_trend"`          // 月度趋势
}

// MerchantRevenue 商户收入数据
type MerchantRevenue struct {
	MerchantID   uint64  `json:"merchant_id"`
	MerchantName string  `json:"merchant_name"`
	Revenue      Money   `json:"revenue"`
	OrderCount   int     `json:"order_count"`
	Percentage   float64 `json:"percentage"` // 占总收入百分比
}

// CategoryRevenue 类别收入数据
type CategoryRevenue struct {
	CategoryID   uint64  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Revenue      Money   `json:"revenue"`
	OrderCount   int     `json:"order_count"`
	Percentage   float64 `json:"percentage"`
}

// ExpenditureItem 支出项目
type ExpenditureItem struct {
	Type       string  `json:"type"`        // 支出类型
	Amount     Money   `json:"amount"`      // 支出金额
	Percentage float64 `json:"percentage"`  // 占总支出百分比
	Description string `json:"description"` // 描述
}

// RightsUsage 权益使用数据
type RightsUsage struct {
	CategoryID      uint64  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	Distributed     int64   `json:"distributed"`     // 发放数量
	Consumed        int64   `json:"consumed"`        // 消耗数量
	Balance         int64   `json:"balance"`         // 余额
	UtilizationRate float64 `json:"utilization_rate"` // 使用率
}

// MonthlyFinancial 月度财务数据
type MonthlyFinancial struct {
	Month             string `json:"month"`              // 月份 (YYYY-MM)
	Revenue           Money  `json:"revenue"`            // 收入
	Expenditure       Money  `json:"expenditure"`        // 支出
	NetProfit         Money  `json:"net_profit"`         // 净利润
	OrderCount        int    `json:"order_count"`        // 订单数量
	RightsDistributed int64  `json:"rights_distributed"` // 权益发放
	RightsConsumed    int64  `json:"rights_consumed"`    // 权益消耗
}

// MerchantOperationReport 商户运营报表数据
type MerchantOperationReport struct {
	MerchantRankings   []MerchantRanking    `json:"merchant_rankings"`   // 商户排名
	PerformanceTrends  []MerchantTrend      `json:"performance_trends"`  // 业绩趋势
	CategoryAnalysis   []CategoryAnalysis   `json:"category_analysis"`   // 类别分析
	GrowthMetrics      *GrowthMetrics       `json:"growth_metrics"`      // 增长指标
}

// MerchantRanking 商户排名数据
type MerchantRanking struct {
	Rank             int     `json:"rank"`              // 排名
	MerchantID       uint64  `json:"merchant_id"`       // 商户ID
	MerchantName     string  `json:"merchant_name"`     // 商户名称
	TotalRevenue     Money   `json:"total_revenue"`     // 总收入
	OrderCount       int     `json:"order_count"`       // 订单数量
	CustomerCount    int     `json:"customer_count"`    // 客户数量
	AverageOrderValue Money  `json:"average_order_value"` // 客单价
	GrowthRate       float64 `json:"growth_rate"`       // 增长率
}

// MerchantTrend 商户趋势数据
type MerchantTrend struct {
	MerchantID   uint64             `json:"merchant_id"`
	MerchantName string             `json:"merchant_name"`
	TrendData    []MonthlyTrendData `json:"trend_data"`
}

// MonthlyTrendData 月度趋势数据
type MonthlyTrendData struct {
	Month        string `json:"month"`         // 月份
	Revenue      Money  `json:"revenue"`       // 收入
	OrderCount   int    `json:"order_count"`   // 订单数量
	CustomerCount int   `json:"customer_count"` // 客户数量
}

// CategoryAnalysis 类别分析数据
type CategoryAnalysis struct {
	CategoryID     uint64  `json:"category_id"`
	CategoryName   string  `json:"category_name"`
	Revenue        Money   `json:"revenue"`
	OrderCount     int     `json:"order_count"`
	MerchantCount  int     `json:"merchant_count"`  // 该类别商户数量
	MarketShare    float64 `json:"market_share"`    // 市场份额
	GrowthRate     float64 `json:"growth_rate"`     // 增长率
}

// GrowthMetrics 增长指标
type GrowthMetrics struct {
	RevenueGrowthRate        float64 `json:"revenue_growth_rate"`         // 收入增长率
	OrderGrowthRate          float64 `json:"order_growth_rate"`           // 订单增长率
	MerchantGrowthRate       float64 `json:"merchant_growth_rate"`        // 商户增长率
	CustomerGrowthRate       float64 `json:"customer_growth_rate"`        // 客户增长率
	AverageOrderValueGrowth  float64 `json:"average_order_value_growth"`  // 客单价增长率
}

// CustomerAnalysisReport 客户分析报表数据
type CustomerAnalysisReport struct {
	UserGrowth            []UserGrowthData       `json:"user_growth"`             // 用户增长
	ActivityMetrics       *ActivityMetrics       `json:"activity_metrics"`        // 活跃度指标
	ConsumptionBehavior   *ConsumptionBehavior   `json:"consumption_behavior"`    // 消费行为
	RetentionAnalysis     *RetentionAnalysis     `json:"retention_analysis"`      // 留存分析
	ChurnAnalysis         *ChurnAnalysis         `json:"churn_analysis"`          // 流失分析
}

// UserGrowthData 用户增长数据
type UserGrowthData struct {
	Month            string `json:"month"`             // 月份
	NewUsers         int    `json:"new_users"`         // 新增用户
	ActiveUsers      int    `json:"active_users"`      // 活跃用户
	CumulativeUsers  int    `json:"cumulative_users"`  // 累计用户
	RetentionRate    float64 `json:"retention_rate"`   // 留存率
}

// ActivityMetrics 活跃度指标
type ActivityMetrics struct {
	DAU              int     `json:"dau"`               // 日活跃用户
	WAU              int     `json:"wau"`               // 周活跃用户
	MAU              int     `json:"mau"`               // 月活跃用户
	AverageSessionTime float64 `json:"average_session_time"` // 平均会话时长(分钟)
	AverageOrderFreq float64 `json:"average_order_freq"`   // 平均下单频次
}

// ConsumptionBehavior 消费行为
type ConsumptionBehavior struct {
	AverageOrderValue  Money   `json:"average_order_value"`   // 平均客单价
	RepurchaseRate     float64 `json:"repurchase_rate"`       // 复购率
	AverageOrderCount  float64 `json:"average_order_count"`   // 平均下单次数
	PreferredCategories []CategoryPreference `json:"preferred_categories"` // 偏好类别
	PaymentMethods     []PaymentMethodStats `json:"payment_methods"`      // 支付方式统计
}

// CategoryPreference 类别偏好
type CategoryPreference struct {
	CategoryID   uint64  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	OrderCount   int     `json:"order_count"`
	Revenue      Money   `json:"revenue"`
	Percentage   float64 `json:"percentage"` // 占该用户总消费百分比
}

// PaymentMethodStats 支付方式统计
type PaymentMethodStats struct {
	Method     string  `json:"method"`     // 支付方式
	Count      int     `json:"count"`      // 使用次数
	Amount     Money   `json:"amount"`     // 支付金额
	Percentage float64 `json:"percentage"` // 使用率
}

// RetentionAnalysis 留存分析
type RetentionAnalysis struct {
	Day1Retention   float64                `json:"day1_retention"`   // 1日留存率
	Day7Retention   float64                `json:"day7_retention"`   // 7日留存率
	Day30Retention  float64                `json:"day30_retention"`  // 30日留存率
	CohortAnalysis  []CohortData          `json:"cohort_analysis"`  // 同期群分析
}

// CohortData 同期群数据
type CohortData struct {
	Cohort          string    `json:"cohort"`           // 同期群标识(月份)
	Users           int       `json:"users"`            // 用户数量
	RetentionRates  []float64 `json:"retention_rates"`  // 各期留存率
}

// ChurnAnalysis 流失分析
type ChurnAnalysis struct {
	ChurnRate         float64           `json:"churn_rate"`          // 流失率
	ChurnReasons      []ChurnReason     `json:"churn_reasons"`       // 流失原因
	RiskUserCount     int               `json:"risk_user_count"`     // 流失风险用户数
	ChurnPrediction   []ChurnPrediction `json:"churn_prediction"`    // 流失预测
}

// ChurnReason 流失原因
type ChurnReason struct {
	Reason     string  `json:"reason"`     // 原因
	Count      int     `json:"count"`      // 数量
	Percentage float64 `json:"percentage"` // 占比
}

// ChurnPrediction 流失预测
type ChurnPrediction struct {
	UserID         uint64  `json:"user_id"`
	Username       string  `json:"username"`
	ChurnRisk      float64 `json:"churn_risk"`       // 流失风险评分 (0-1)
	LastActiveDate time.Time `json:"last_active_date"` // 最后活跃时间
	Recommendation string  `json:"recommendation"`   // 挽回建议
}

// ReportCreateRequest 报表生成请求
type ReportCreateRequest struct {
	ReportType ReportType `json:"report_type" binding:"required"`
	PeriodType PeriodType `json:"period_type" binding:"required"`
	StartDate  time.Time  `json:"start_date" binding:"required"`
	EndDate    time.Time  `json:"end_date" binding:"required"`
	FileFormat FileFormat `json:"file_format" binding:"required"`
	MerchantID *uint64    `json:"merchant_id,omitempty"` // 可选，指定商户
	Config     map[string]interface{} `json:"config,omitempty"` // 自定义配置
}

// ReportListRequest 报表列表请求
type ReportListRequest struct {
	ReportType *ReportType  `json:"report_type,omitempty"`
	Status     *ReportStatus `json:"status,omitempty"`
	StartDate  *time.Time   `json:"start_date,omitempty"`
	EndDate    *time.Time   `json:"end_date,omitempty"`
	Page       int          `json:"page" binding:"min=1"`
	PageSize   int          `json:"page_size" binding:"min=1,max=100"`
}

// ReportScheduleRequest 定时报表请求
type ReportScheduleRequest struct {
	Name           string     `json:"name" binding:"required"`
	ReportType     ReportType `json:"report_type" binding:"required"`
	TemplateConfig map[string]interface{} `json:"template_config" binding:"required"`
	Schedule       ScheduleConfig `json:"schedule" binding:"required"`
	Recipients     []string   `json:"recipients" binding:"required"`
	FileFormat     FileFormat `json:"file_format" binding:"required"`
	Enabled        bool       `json:"enabled"`
}

// ScheduleConfig 调度配置
type ScheduleConfig struct {
	Frequency string `json:"frequency" binding:"required,oneof=daily weekly monthly"` // 频率
	Time      string `json:"time" binding:"required"`     // 时间 (HH:mm)
	Timezone  string `json:"timezone" binding:"required"` // 时区
}

// AnalyticsQueryRequest 数据分析查询请求
type AnalyticsQueryRequest struct {
	MetricType string            `json:"metric_type" binding:"required"` // 指标类型
	StartDate  time.Time         `json:"start_date" binding:"required"`
	EndDate    time.Time         `json:"end_date" binding:"required"`
	GroupBy    string            `json:"group_by,omitempty"`    // 分组字段
	Filters    map[string]interface{} `json:"filters,omitempty"` // 过滤条件
	MerchantID *uint64           `json:"merchant_id,omitempty"` // 可选商户ID
}
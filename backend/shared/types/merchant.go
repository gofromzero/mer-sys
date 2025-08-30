package types

import (
	"encoding/json"
	"errors"
	"time"
)

// MerchantStatus 商户状态
type MerchantStatus string

const (
	MerchantStatusPending     MerchantStatus = "pending"
	MerchantStatusActive      MerchantStatus = "active"
	MerchantStatusSuspended   MerchantStatus = "suspended"
	MerchantStatusDeactivated MerchantStatus = "deactivated"
)

// BusinessInfo 商户业务信息
type BusinessInfo struct {
	Type         string `json:"type"`          // 商户类型: retail, wholesale, service
	Category     string `json:"category"`      // 业务分类
	License      string `json:"license"`       // 营业执照号
	LegalName    string `json:"legal_name"`    // 法人名称
	ContactName  string `json:"contact_name"`  // 联系人姓名
	ContactPhone string `json:"contact_phone"` // 联系电话
	ContactEmail string `json:"contact_email"` // 联系邮箱
	Address      string `json:"address"`       // 经营地址
	Scope        string `json:"scope"`         // 经营范围
	Description  string `json:"description"`   // 商户描述
}

// RightsBalance 权益余额信息
type RightsBalance struct {
	TotalBalance      float64   `json:"total_balance"`      // 总余额
	UsedBalance       float64   `json:"used_balance"`       // 已使用余额
	FrozenBalance     float64   `json:"frozen_balance"`     // 冻结余额
	AvailableBalance  float64   `json:"available_balance"`  // 可用余额
	LastUpdated       time.Time `json:"last_updated"`       // 最后更新时间
	WarningThreshold  *float64  `json:"warning_threshold"`  // 预警阈值
	CriticalThreshold *float64  `json:"critical_threshold"` // 紧急阈值
	TrendCoefficient  *float64  `json:"trend_coefficient"`  // 趋势系数
}

// UpdateAvailableBalance 更新可用余额
func (rb *RightsBalance) UpdateAvailableBalance() {
	rb.AvailableBalance = rb.TotalBalance - rb.UsedBalance - rb.FrozenBalance
	rb.LastUpdated = time.Now()
}

// GetAvailableBalance 获取计算后的可用余额
func (rb *RightsBalance) GetAvailableBalance() float64 {
	return rb.TotalBalance - rb.UsedBalance - rb.FrozenBalance
}

// Merchant 商户实体
type Merchant struct {
	ID             uint64         `json:"id" db:"id"`
	TenantID       uint64         `json:"tenant_id" db:"tenant_id"`
	Name           string         `json:"name" db:"name"`
	Code           string         `json:"code" db:"code"`
	Status         MerchantStatus `json:"status" db:"status"`
	BusinessInfo   *BusinessInfo  `json:"business_info" db:"business_info"`
	RightsBalance  *RightsBalance `json:"rights_balance" db:"rights_balance"`
	RegistrationTime *time.Time   `json:"registration_time" db:"registration_time"` // 注册申请时间
	ApprovalTime     *time.Time   `json:"approval_time" db:"approval_time"`         // 审批时间
	ApprovedBy       *uint64      `json:"approved_by" db:"approved_by"`             // 审批人ID
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// MarshalJSON 自定义JSON序列化
func (m *Merchant) MarshalJSON() ([]byte, error) {
	type Alias Merchant
	return json.Marshal(&struct {
		*Alias
		BusinessInfo  json.RawMessage `json:"business_info,omitempty"`
		RightsBalance json.RawMessage `json:"rights_balance,omitempty"`
	}{
		Alias: (*Alias)(m),
		BusinessInfo: func() json.RawMessage {
			if m.BusinessInfo != nil {
				data, _ := json.Marshal(m.BusinessInfo)
				return data
			}
			return nil
		}(),
		RightsBalance: func() json.RawMessage {
			if m.RightsBalance != nil {
				data, _ := json.Marshal(m.RightsBalance)
				return data
			}
			return nil
		}(),
	})
}

// ProductStatus 商品状态
type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

// Money 金额信息
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// InventoryInfo 库存信息
type InventoryInfo struct {
	StockQuantity    int  `json:"stock_quantity"`
	ReservedQuantity int  `json:"reserved_quantity"`
	TrackInventory   bool `json:"track_inventory"`
}

// AvailableStock 计算可用库存
func (ii *InventoryInfo) AvailableStock() int {
	return ii.StockQuantity - ii.ReservedQuantity
}

// Product 商品实体
type Product struct {
	ID           uint64         `json:"id" db:"id"`
	TenantID     uint64         `json:"tenant_id" db:"tenant_id"`
	MerchantID   uint64         `json:"merchant_id" db:"merchant_id"`
	Name         string         `json:"name" db:"name"`
	Description  string         `json:"description" db:"description"`
	CategoryID   *uint64        `json:"category_id,omitempty" db:"category_id"`
	Tags         StringArray    `json:"tags" db:"tags"`
	PriceAmount  float64        `json:"price_amount" db:"price_amount"`
	PriceCurrency string        `json:"price_currency" db:"price_currency"`
	RightsCost   float64        `json:"rights_cost" db:"rights_cost"`
	InventoryInfo *InventoryInfo `json:"inventory_info" db:"inventory_info"`
	Images       ProductImages  `json:"images" db:"images"`
	Status       ProductStatus  `json:"status" db:"status"`
	Version      int           `json:"version" db:"version"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// GetPrice 获取价格信息
func (p *Product) GetPrice() Money {
	return Money{
		Amount:   p.PriceAmount,
		Currency: p.PriceCurrency,
	}
}

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// OrderItem 订单项目
type OrderItem struct {
	ProductID uint64  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	RightsCost float64 `json:"rights_cost"`
}

// PaymentInfo 支付信息
type PaymentInfo struct {
	Method        string    `json:"method"`         // 支付方式: wechat, alipay, cash
	TransactionID string    `json:"transaction_id"` // 交易ID
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	Amount        float64   `json:"amount"`
}

// VerificationInfo 核销信息
type VerificationInfo struct {
	VerificationCode string     `json:"verification_code"`
	QRCodeURL        string     `json:"qr_code_url"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	VerifiedBy       string     `json:"verified_by,omitempty"`
}

// OrderStatusOperatorType 操作员类型枚举
type OrderStatusOperatorType string

const (
	OrderStatusOperatorTypeCustomer OrderStatusOperatorType = "customer"
	OrderStatusOperatorTypeMerchant OrderStatusOperatorType = "merchant"
	OrderStatusOperatorTypeSystem   OrderStatusOperatorType = "system"
	OrderStatusOperatorTypeAdmin    OrderStatusOperatorType = "admin"
)

// String 返回操作员类型的字符串表示
func (ot OrderStatusOperatorType) String() string {
	return string(ot)
}

// IsValid 验证操作员类型是否有效
func (ot OrderStatusOperatorType) IsValid() bool {
	switch ot {
	case OrderStatusOperatorTypeCustomer, OrderStatusOperatorTypeMerchant, 
		 OrderStatusOperatorTypeSystem, OrderStatusOperatorTypeAdmin:
		return true
	default:
		return false
	}
}

// OrderStatusHistory 订单状态变更历史记录
type OrderStatusHistory struct {
	ID           uint64                  `json:"id" db:"id"`
	TenantID     uint64                  `json:"tenant_id" db:"tenant_id"`
	OrderID      uint64                  `json:"order_id" db:"order_id"`
	FromStatus   OrderStatusInt          `json:"from_status" db:"from_status"`
	ToStatus     OrderStatusInt          `json:"to_status" db:"to_status"`
	Reason       string                  `json:"reason" db:"reason"`
	OperatorID   *uint64                 `json:"operator_id,omitempty" db:"operator_id"`
	OperatorType OrderStatusOperatorType `json:"operator_type" db:"operator_type"`
	Metadata     interface{}             `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time               `json:"created_at" db:"created_at"`
}

// OrderTimeoutConfig 订单超时配置
type OrderTimeoutConfig struct {
	ID                      uint64 `json:"id" db:"id"`
	TenantID                uint64 `json:"tenant_id" db:"tenant_id"`
	MerchantID              *uint64 `json:"merchant_id,omitempty" db:"merchant_id"`
	PaymentTimeoutMinutes   int    `json:"payment_timeout_minutes" db:"payment_timeout_minutes"`
	ProcessingTimeoutHours  int    `json:"processing_timeout_hours" db:"processing_timeout_hours"`
	AutoCompleteEnabled     bool   `json:"auto_complete_enabled" db:"auto_complete_enabled"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

// OrderTimeoutStatistics 订单超时统计信息
type OrderTimeoutStatistics struct {
	PendingTimeoutCount       int     `json:"pending_timeout_count"`       // 待支付超时订单数量
	PendingTimeoutAmount      float64 `json:"pending_timeout_amount"`      // 待支付超时订单金额
	ProcessingTimeoutCount    int     `json:"processing_timeout_count"`    // 处理中超时订单数量
	ProcessingTimeoutAmount   float64 `json:"processing_timeout_amount"`   // 处理中超时订单金额
	TodayAutoCancelledCount   int     `json:"today_auto_cancelled_count"`  // 今日自动取消订单数量
	TodayAutoCancelledAmount  float64 `json:"today_auto_cancelled_amount"` // 今日自动取消订单金额
}

// OrderQueryRequest 订单查询请求
type OrderQueryRequest struct {
	TenantID      uint64           `json:"tenant_id"`
	MerchantID    *uint64          `json:"merchant_id,omitempty"`
	CustomerID    *uint64          `json:"customer_id,omitempty"`
	Status        []OrderStatusInt `json:"status,omitempty"`
	StartDate     *time.Time       `json:"start_date,omitempty"`
	EndDate       *time.Time       `json:"end_date,omitempty"`
	SearchKeyword *string          `json:"search_keyword,omitempty"`
	Page          int              `json:"page" v:"required|min:1"`
	PageSize      int              `json:"page_size" v:"required|min:1|max:100"`
	SortBy        string           `json:"sort_by" v:"in:created_at,updated_at,total_amount"`
	SortOrder     string           `json:"sort_order" v:"in:asc,desc"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Items    []OrderSummary `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	HasNext  bool          `json:"has_next"`
}

// OrderSummary 订单摘要信息
type OrderSummary struct {
	ID                  uint64               `json:"id"`
	OrderNumber         string               `json:"order_number"`
	Status              OrderStatusInt       `json:"status"`
	CustomerName        string               `json:"customer_name"`
	MerchantName        string               `json:"merchant_name"`
	TotalAmount         float64              `json:"total_amount"`
	ItemCount           int                  `json:"item_count"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
	LatestStatusChange  *OrderStatusHistory  `json:"latest_status_change,omitempty"`
}

// UpdateOrderStatusRequest 更新订单状态请求
type UpdateOrderStatusRequest struct {
	Status       OrderStatusInt          `json:"status" v:"required#新状态不能为空"`
	Reason       string                  `json:"reason" v:"required|min:1|max:255#变更原因不能为空且不能超过255字符"`
	OperatorType OrderStatusOperatorType `json:"operator_type" v:"required#操作员类型不能为空"`
	Metadata     interface{}             `json:"metadata,omitempty"`
}

// BatchUpdateOrderStatusRequest 批量更新订单状态请求
type BatchUpdateOrderStatusRequest struct {
	OrderIDs     []uint64                `json:"order_ids" v:"required|length:1,100#订单ID列表不能为空且不能超过100个"`
	Status       OrderStatusInt          `json:"status" v:"required#新状态不能为空"`
	Reason       string                  `json:"reason" v:"required|min:1|max:255#变更原因不能为空且不能超过255字符"`
	OperatorType OrderStatusOperatorType `json:"operator_type" v:"required#操作员类型不能为空"`
	Metadata     interface{}             `json:"metadata,omitempty"`
}

// OrderStatusValidationError 订单状态验证错误
type OrderStatusValidationError struct {
	OrderID    uint64         `json:"order_id"`
	FromStatus OrderStatusInt `json:"from_status"`
	ToStatus   OrderStatusInt `json:"to_status"`
	Message    string         `json:"message"`
}

// BatchUpdateOrderStatusResponse 批量更新订单状态响应
type BatchUpdateOrderStatusResponse struct {
	SuccessCount int                          `json:"success_count"`
	FailCount    int                          `json:"fail_count"`
	Errors       []OrderStatusValidationError `json:"errors,omitempty"`
}

// Order 订单实体（扩展支持状态历史）
type Order struct {
	ID               uint64            `json:"id" db:"id"`
	TenantID         uint64            `json:"tenant_id" db:"tenant_id"`
	MerchantID       uint64            `json:"merchant_id" db:"merchant_id"`
	CustomerID       uint64            `json:"customer_id" db:"customer_id"`
	OrderNumber      string            `json:"order_number" db:"order_number"`
	Status           OrderStatus       `json:"status" db:"status"`
	Items            []OrderItem       `json:"items" db:"items"`
	PaymentInfo      *PaymentInfo      `json:"payment_info" db:"payment_info"`
	VerificationInfo *VerificationInfo `json:"verification_info" db:"verification_info"`
	TotalAmount      float64           `json:"total_amount" db:"total_amount"`
	TotalRightsCost  float64           `json:"total_rights_cost" db:"total_rights_cost"`
	StatusUpdatedAt  time.Time         `json:"status_updated_at" db:"status_updated_at"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at" db:"updated_at"`
	// 状态历史（查询时可选填充）
	StatusHistory    []OrderStatusHistory `json:"status_history,omitempty" db:"-"`
}

// MerchantRegistrationRequest 商户注册请求
type MerchantRegistrationRequest struct {
	Name         string        `json:"name" binding:"required,min=1,max=100"`
	Code         string        `json:"code" binding:"required,min=1,max=50,alphanum"`
	BusinessInfo *BusinessInfo `json:"business_info" binding:"required"`
}

// MerchantApprovalRequest 商户审批请求
type MerchantApprovalRequest struct {
	Action  string `json:"action" binding:"required,oneof=approve reject"` // approve 或 reject
	Comment string `json:"comment,omitempty"`                              // 审批意见
}

// MerchantUpdateRequest 商户信息更新请求
type MerchantUpdateRequest struct {
	Name         *string       `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	BusinessInfo *BusinessInfo `json:"business_info,omitempty"`
}

// MerchantStatusUpdateRequest 商户状态更新请求
type MerchantStatusUpdateRequest struct {
	Status  MerchantStatus `json:"status" binding:"required,oneof=active suspended deactivated"`
	Comment string         `json:"comment,omitempty"` // 状态变更原因
}

// MerchantListQuery 商户列表查询参数
type MerchantListQuery struct {
	Page     int            `form:"page,default=1" binding:"min=1"`
	PageSize int            `form:"page_size,default=20" binding:"min=1,max=100"`
	Status   MerchantStatus `form:"status,omitempty"`
	Name     string         `form:"name,omitempty"`
	Search   string         `form:"search,omitempty"` // 全文搜索
}

// 权益使用趋势点 (复用 monitoring.go 中的 TimePeriod 和 TrendDirection)
type RightsUsagePoint struct {
	Date    time.Time      `json:"date"`
	Balance float64        `json:"balance"`
	Usage   float64        `json:"usage"`
	Trend   TrendDirection `json:"trend"`
}

// 任务类型
type TaskType string

const (
	TaskTypeOrderProcessing       TaskType = "order_processing"
	TaskTypeVerificationPending  TaskType = "verification_pending"
	TaskTypeLowBalanceWarning     TaskType = "low_balance_warning"
	TaskTypeProductUpdateNeeded   TaskType = "product_update_needed"
)

// 优先级
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// 待处理任务
type PendingTask struct {
	ID          string     `json:"id"`
	Type        TaskType   `json:"type"`
	Description string     `json:"description"`
	Priority    Priority   `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Count       int        `json:"count"`
}

// 公告信息
type Announcement struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Priority    Priority   `json:"priority"`
	PublishDate time.Time  `json:"publish_date"`
	ExpireDate  *time.Time `json:"expire_date,omitempty"`
	ReadStatus  bool       `json:"read_status"`
}

// 通知信息
type Notification struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Priority  Priority  `json:"priority"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// 商户仪表板数据
type MerchantDashboardData struct {
	MerchantID   uint64   `json:"merchant_id"`
	TenantID     uint64   `json:"tenant_id"`
	Period       TimePeriod `json:"period"`
	
	// 核心业务指标 (AC: 1)
	TotalSales     float64        `json:"total_sales"`      // 总销售额
	TotalOrders    int            `json:"total_orders"`     // 总订单数
	TotalCustomers int            `json:"total_customers"`  // 总客户数
	RightsBalance  *RightsBalance `json:"rights_balance"`   // 权益余额
	
	// 权益使用情况 (AC: 2) 
	RightsUsageTrend       []RightsUsagePoint `json:"rights_usage_trend"`
	RightsAlerts          []RightsAlert      `json:"rights_alerts"`
	PredictedDepletionDays *int               `json:"predicted_depletion_days,omitempty"`
	
	// 待处理事项 (AC: 3)
	PendingOrders         int            `json:"pending_orders"`         // 待处理订单
	PendingVerifications  int            `json:"pending_verifications"`  // 待核销订单
	PendingTasks         []PendingTask   `json:"pending_tasks"`
	
	// 系统通知 (AC: 4)
	Announcements []Announcement `json:"announcements"`
	Notifications []Notification `json:"notifications"`
	
	LastUpdated time.Time `json:"last_updated"`
}

// 组件位置
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// 组件大小
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// 组件类型
type WidgetType string

const (
	WidgetTypeSalesOverview  WidgetType = "sales_overview"
	WidgetTypeRightsBalance  WidgetType = "rights_balance"
	WidgetTypeRightsTrend    WidgetType = "rights_trend"
	WidgetTypePendingTasks   WidgetType = "pending_tasks"
	WidgetTypeRecentOrders   WidgetType = "recent_orders"
	WidgetTypeAnnouncements  WidgetType = "announcements"
	WidgetTypeQuickActions   WidgetType = "quick_actions"
)

// 仪表板组件
type DashboardWidget struct {
	ID       string                 `json:"id"`
	Type     WidgetType             `json:"type"`
	Position Position               `json:"position"`
	Size     Size                   `json:"size"`
	Config   map[string]interface{} `json:"config"`
	Visible  bool                   `json:"visible"`
}

// 组件偏好设置
type WidgetPreference struct {
	WidgetType WidgetType             `json:"widget_type"`
	Enabled    bool                   `json:"enabled"`
	Config     map[string]interface{} `json:"config"`
}

// 布局配置
type LayoutConfig struct {
	Columns int               `json:"columns"`
	Widgets []DashboardWidget `json:"widgets"`
}

// 移动端布局配置
type MobileLayoutConfig struct {
	Columns int               `json:"columns"`
	Widgets []DashboardWidget `json:"widgets"`
}

// 仪表板配置
type DashboardConfig struct {
	MerchantID       uint64               `json:"merchant_id" db:"merchant_id"`
	LayoutConfig     *LayoutConfig        `json:"layout_config" db:"layout_config"`
	WidgetPreferences []WidgetPreference  `json:"widget_preferences" db:"widget_preferences"`
	RefreshInterval  int                  `json:"refresh_interval" db:"refresh_interval"` // 秒
	MobileLayout     *MobileLayoutConfig  `json:"mobile_layout" db:"mobile_layout"`
}

// 跨租户访问错误
var (
	ErrCrossTenantAccess = errors.New("cross-tenant access denied")
	ErrTenantMismatch    = errors.New("tenant ID mismatch")
	ErrInvalidTenantID   = errors.New("invalid tenant ID")
)
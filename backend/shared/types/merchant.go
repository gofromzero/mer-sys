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
	PriceAmount  float64        `json:"price_amount" db:"price_amount"`
	PriceCurrency string        `json:"price_currency" db:"price_currency"`
	RightsCost   float64        `json:"rights_cost" db:"rights_cost"`
	InventoryInfo *InventoryInfo `json:"inventory_info" db:"inventory_info"`
	Status       ProductStatus  `json:"status" db:"status"`
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

// Order 订单实体
type Order struct {
	ID              uint64            `json:"id" db:"id"`
	TenantID        uint64            `json:"tenant_id" db:"tenant_id"`
	MerchantID      uint64            `json:"merchant_id" db:"merchant_id"`
	CustomerID      uint64            `json:"customer_id" db:"customer_id"`
	OrderNumber     string            `json:"order_number" db:"order_number"`
	Status          OrderStatus       `json:"status" db:"status"`
	Items           []OrderItem       `json:"items" db:"items"`
	PaymentInfo     *PaymentInfo      `json:"payment_info" db:"payment_info"`
	VerificationInfo *VerificationInfo `json:"verification_info" db:"verification_info"`
	TotalAmount     float64           `json:"total_amount" db:"total_amount"`
	TotalRightsCost float64           `json:"total_rights_cost" db:"total_rights_cost"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
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

// 跨租户访问错误
var (
	ErrCrossTenantAccess = errors.New("cross-tenant access denied")
	ErrTenantMismatch    = errors.New("tenant ID mismatch")
	ErrInvalidTenantID   = errors.New("invalid tenant ID")
)
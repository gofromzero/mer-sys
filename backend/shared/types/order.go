package types

import (
	"time"
)

// OrderStatusInt 订单状态枚举（数字形式）
type OrderStatusInt int

const (
	OrderStatusIntPending    OrderStatusInt = 1 // 待支付
	OrderStatusIntPaid       OrderStatusInt = 2 // 已支付
	OrderStatusIntProcessing OrderStatusInt = 3 // 处理中
	OrderStatusIntCompleted  OrderStatusInt = 4 // 已完成
	OrderStatusIntCancelled  OrderStatusInt = 5 // 已取消
)

// ToOrderStatus 转换为字符串类型的OrderStatus
func (os OrderStatusInt) ToOrderStatus() OrderStatus {
	switch os {
	case OrderStatusIntPending:
		return OrderStatusPending
	case OrderStatusIntPaid:
		return OrderStatusPaid
	case OrderStatusIntProcessing:
		return OrderStatusProcessing
	case OrderStatusIntCompleted:
		return OrderStatusCompleted
	case OrderStatusIntCancelled:
		return OrderStatusCancelled
	default:
		return OrderStatusPending
	}
}

// String 返回订单状态的字符串表示
func (os OrderStatusInt) String() string {
	switch os {
	case OrderStatusIntPending:
		return "pending"
	case OrderStatusIntPaid:
		return "paid"
	case OrderStatusIntProcessing:
		return "processing"
	case OrderStatusIntCompleted:
		return "completed"
	case OrderStatusIntCancelled:
		return "cancelled"
	default:
		return "pending"
	}
}

// IsValidTransition 检查状态流转是否合法
func (os OrderStatusInt) IsValidTransition(toStatus OrderStatusInt) bool {
	// 定义合法的状态流转规则
	validTransitions := map[OrderStatusInt][]OrderStatusInt{
		OrderStatusIntPending:    {OrderStatusIntPaid, OrderStatusIntCancelled},
		OrderStatusIntPaid:       {OrderStatusIntProcessing, OrderStatusIntCancelled},
		OrderStatusIntProcessing: {OrderStatusIntCompleted, OrderStatusIntCancelled},
		OrderStatusIntCompleted:  {}, // 已完成状态不能再变更
		OrderStatusIntCancelled:  {}, // 已取消状态不能再变更
	}
	
	allowedStatuses, exists := validTransitions[os]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedStatuses {
		if allowed == toStatus {
			return true
		}
	}
	return false
}

// ExtendedOrderItem 扩展的订单项（包含更多字段）
type ExtendedOrderItem struct {
	ID                 uint64  `json:"id"`
	OrderID            uint64  `json:"order_id"`
	ProductID          uint64  `json:"product_id"`
	ProductName        string  `json:"product_name"`
	Quantity           int     `json:"quantity"`
	UnitPrice          float64 `json:"unit_price"`
	UnitRightsCost     float64 `json:"unit_rights_cost"`
	SubtotalAmount     float64 `json:"subtotal_amount"`
	SubtotalRightsCost float64 `json:"subtotal_rights_cost"`
}

// ExtendedPaymentInfo 扩展的支付信息
type ExtendedPaymentInfo struct {
	PaymentMethod PaymentMethod `json:"payment_method"`
	PaymentID     string        `json:"payment_id"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	PaidAmount    float64       `json:"paid_amount"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	PaymentURL    string        `json:"payment_url,omitempty"`
	CallbackData  interface{}   `json:"callback_data,omitempty"`
}

// PaymentMethod 支付方式枚举
type PaymentMethod string

const (
	PaymentMethodAlipay  PaymentMethod = "alipay"
	PaymentMethodWechat  PaymentMethod = "wechat"
	PaymentMethodBalance PaymentMethod = "balance"
)

// PaymentStatus 支付状态枚举
type PaymentStatus int

const (
	PaymentStatusUnpaid   PaymentStatus = 1 // 未支付
	PaymentStatusPaying   PaymentStatus = 2 // 支付中
	PaymentStatusPaid     PaymentStatus = 3 // 已支付
	PaymentStatusFailed   PaymentStatus = 4 // 支付失败
	PaymentStatusRefunded PaymentStatus = 5 // 已退款
)

// String 返回支付状态的字符串表示
func (ps PaymentStatus) String() string {
	switch ps {
	case PaymentStatusUnpaid:
		return "unpaid"
	case PaymentStatusPaying:
		return "paying"
	case PaymentStatusPaid:
		return "paid"
	case PaymentStatusFailed:
		return "failed"
	case PaymentStatusRefunded:
		return "refunded"
	default:
		return "unknown"
	}
}


// Cart 购物车实体
type Cart struct {
	ID         uint64     `json:"id" db:"id"`
	TenantID   uint64     `json:"tenant_id" db:"tenant_id"`
	CustomerID uint64     `json:"customer_id" db:"customer_id"`
	Items      []CartItem `json:"items"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`
}

// CartItem 购物车项
type CartItem struct {
	ID        uint64    `json:"id" db:"id"`
	TenantID  uint64    `json:"tenant_id" db:"tenant_id"`
	CartID    uint64    `json:"cart_id" db:"cart_id"`
	ProductID uint64    `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	AddedAt   time.Time `json:"added_at" db:"added_at"`
}

// PaymentRecord 支付记录
type PaymentRecord struct {
	ID            uint64        `json:"id" db:"id"`
	TenantID      uint64        `json:"tenant_id" db:"tenant_id"`
	OrderID       uint64        `json:"order_id" db:"order_id"`
	PaymentMethod PaymentMethod `json:"payment_method" db:"payment_method"`
	PaymentID     string        `json:"payment_id" db:"payment_id"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	Amount        float64       `json:"amount" db:"amount"`
	Currency      string        `json:"currency" db:"currency"`
	PaymentURL    string        `json:"payment_url" db:"payment_url"`
	CallbackData  interface{}   `json:"callback_data" db:"callback_data"`
	PaidAt        *time.Time    `json:"paid_at" db:"paid_at"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	MerchantID uint64 `json:"merchant_id" v:"required#商户ID不能为空"`
	Items      []struct {
		ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
		Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
	} `json:"items" v:"required|length:1,50#订单项不能为空|订单项不能超过50个"`
}

// AddCartItemRequest 添加购物车项请求
type AddCartItemRequest struct {
	ProductID uint64 `json:"product_id" v:"required#商品ID不能为空"`
	Quantity  int    `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
}

// UpdateCartItemRequest 更新购物车项请求
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" v:"required|min:1#数量不能为空|数量必须大于0"`
}

// InitiatePaymentRequest 发起支付请求
type InitiatePaymentRequest struct {
	PaymentMethod PaymentMethod `json:"payment_method" v:"required#支付方式不能为空"`
	ReturnURL     string        `json:"return_url,omitempty"`
}

// OrderConfirmation 订单确认信息
type OrderConfirmation struct {
	Items           []OrderConfirmationItem `json:"items"`
	TotalAmount     float64                 `json:"total_amount"`
	TotalRightsCost float64                 `json:"total_rights_cost"`
	AvailableRights float64                 `json:"available_rights"`
	CanCreate       bool                    `json:"can_create"`
	ErrorMessage    string                  `json:"error_message,omitempty"`
}

// OrderConfirmationItem 订单确认项
type OrderConfirmationItem struct {
	ProductID          uint64  `json:"product_id"`
	ProductName        string  `json:"product_name"`
	Quantity           int     `json:"quantity"`
	UnitPrice          float64 `json:"unit_price"`
	UnitRightsCost     float64 `json:"unit_rights_cost"`
	SubtotalAmount     float64 `json:"subtotal_amount"`
	SubtotalRightsCost float64 `json:"subtotal_rights_cost"`
	StockAvailable     int     `json:"stock_available"`
	StockSufficient    bool    `json:"stock_sufficient"`
}
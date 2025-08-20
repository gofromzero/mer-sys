package types

import (
	"fmt"
	"time"
)

// FundType 资金类型枚举
type FundType int

const (
	FundTypeDeposit     FundType = 1 // 充值
	FundTypeAllocation  FundType = 2 // 分配
	FundTypeConsumption FundType = 3 // 消费
	FundTypeRefund      FundType = 4 // 退款
)

func (f FundType) String() string {
	switch f {
	case FundTypeDeposit:
		return "deposit"
	case FundTypeAllocation:
		return "allocation"
	case FundTypeConsumption:
		return "consumption"
	case FundTypeRefund:
		return "refund"
	default:
		return "unknown"
	}
}

// FundStatus 资金状态枚举
type FundStatus int

const (
	FundStatusPending   FundStatus = 1 // 待处理
	FundStatusConfirmed FundStatus = 2 // 已确认
	FundStatusFailed    FundStatus = 3 // 失败
	FundStatusCancelled FundStatus = 4 // 已取消
)

func (f FundStatus) String() string {
	switch f {
	case FundStatusPending:
		return "pending"
	case FundStatusConfirmed:
		return "confirmed"
	case FundStatusFailed:
		return "failed"
	case FundStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// TransactionType 交易类型枚举
type TransactionType int

const (
	TransactionTypeCredit TransactionType = 1 // 入账
	TransactionTypeDebit  TransactionType = 2 // 出账
)

func (t TransactionType) String() string {
	switch t {
	case TransactionTypeCredit:
		return "credit"
	case TransactionTypeDebit:
		return "debit"
	default:
		return "unknown"
	}
}

// Fund 资金记录实体
type Fund struct {
	ID         uint64    `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	TenantID   uint64    `json:"tenant_id" gorm:"not null;index:idx_tenant_merchant" db:"tenant_id"`
	MerchantID uint64    `json:"merchant_id" gorm:"not null;index:idx_tenant_merchant" db:"merchant_id"`
	FundType   FundType  `json:"fund_type" gorm:"not null;index:idx_fund_type" db:"fund_type"`
	Amount     float64   `json:"amount" gorm:"type:decimal(15,2);not null" db:"amount"`
	Currency   string    `json:"currency" gorm:"size:3;default:CNY" db:"currency"`
	Status     FundStatus `json:"status" gorm:"not null;default:1;index:idx_status" db:"status"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

// TableName 指定表名
func (Fund) TableName() string {
	return "funds"
}

// FundTransaction 资金流转记录实体
type FundTransaction struct {
	ID              uint64          `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	TenantID        uint64          `json:"tenant_id" gorm:"not null;index:idx_tenant_merchant" db:"tenant_id"`
	MerchantID      uint64          `json:"merchant_id" gorm:"not null;index:idx_tenant_merchant" db:"merchant_id"`
	FundID          uint64          `json:"fund_id" gorm:"not null;index:idx_fund_id" db:"fund_id"`
	TransactionType TransactionType `json:"transaction_type" gorm:"not null" db:"transaction_type"`
	Amount          float64         `json:"amount" gorm:"type:decimal(15,2);not null" db:"amount"`
	BalanceBefore   float64         `json:"balance_before" gorm:"type:decimal(15,2);not null" db:"balance_before"`
	BalanceAfter    float64         `json:"balance_after" gorm:"type:decimal(15,2);not null" db:"balance_after"`
	OperatorID      uint64          `json:"operator_id" gorm:"not null;index:idx_operator" db:"operator_id"`
	Description     string          `json:"description,omitempty" gorm:"type:text" db:"description"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
}

// TableName 指定表名
func (FundTransaction) TableName() string {
	return "fund_transactions"
}


// DepositRequest 单笔充值请求
type DepositRequest struct {
	MerchantID  uint64  `json:"merchant_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Currency    string  `json:"currency" binding:"required,len=3"`
	Description string  `json:"description,omitempty"`
}

// BatchDepositRequest 批量充值请求
type BatchDepositRequest struct {
	Deposits []DepositRequest `json:"deposits" binding:"required,min=1,max=100"`
}

// AllocateRequest 权益分配请求
type AllocateRequest struct {
	MerchantID  uint64  `json:"merchant_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description,omitempty"`
}

// FundTransactionQuery 资金流转查询参数
type FundTransactionQuery struct {
	TenantID        uint64          `json:"tenant_id,omitempty" form:"tenant_id"`
	MerchantID      uint64          `json:"merchant_id,omitempty" form:"merchant_id"`
	FundID          uint64          `json:"fund_id,omitempty" form:"fund_id"`
	TransactionType TransactionType `json:"transaction_type,omitempty" form:"transaction_type"`
	OperatorID      uint64          `json:"operator_id,omitempty" form:"operator_id"`
	StartTime       *time.Time      `json:"start_time,omitempty" form:"start_time"`
	EndTime         *time.Time      `json:"end_time,omitempty" form:"end_time"`
	Page            int             `json:"page" form:"page" binding:"min=1"`
	PageSize        int             `json:"page_size" form:"page_size" binding:"min=1,max=100"`
}

// FundSummary 资金概览统计
type FundSummary struct {
	TotalDeposits    float64 `json:"total_deposits"`
	TotalAllocations float64 `json:"total_allocations"`
	TotalConsumption float64 `json:"total_consumption"`
	TotalRefunds     float64 `json:"total_refunds"`
	AvailableBalance float64 `json:"available_balance"`
}

// Validate 验证资金记录
func (f *Fund) Validate() error {
	if f.TenantID == 0 {
		return fmt.Errorf("租户ID不能为空")
	}
	if f.MerchantID == 0 {
		return fmt.Errorf("商户ID不能为空")
	}
	if f.Amount <= 0 {
		return fmt.Errorf("金额必须大于0")
	}
	if len(f.Currency) != 3 {
		return fmt.Errorf("货币代码必须为3位")
	}
	if f.FundType < FundTypeDeposit || f.FundType > FundTypeRefund {
		return fmt.Errorf("无效的资金类型")
	}
	if f.Status < FundStatusPending || f.Status > FundStatusCancelled {
		return fmt.Errorf("无效的资金状态")
	}
	return nil
}

// Validate 验证资金流转记录
func (ft *FundTransaction) Validate() error {
	if ft.TenantID == 0 {
		return fmt.Errorf("租户ID不能为空")
	}
	if ft.MerchantID == 0 {
		return fmt.Errorf("商户ID不能为空")
	}
	if ft.FundID == 0 {
		return fmt.Errorf("资金记录ID不能为空")
	}
	if ft.OperatorID == 0 {
		return fmt.Errorf("操作人ID不能为空")
	}
	if ft.Amount <= 0 {
		return fmt.Errorf("交易金额必须大于0")
	}
	if ft.BalanceBefore < 0 {
		return fmt.Errorf("交易前余额不能为负数")
	}
	if ft.BalanceAfter < 0 {
		return fmt.Errorf("交易后余额不能为负数")
	}
	if ft.TransactionType != TransactionTypeCredit && ft.TransactionType != TransactionTypeDebit {
		return fmt.Errorf("无效的交易类型")
	}
	
	// 验证余额计算逻辑
	if ft.TransactionType == TransactionTypeCredit {
		if ft.BalanceAfter != ft.BalanceBefore + ft.Amount {
			return fmt.Errorf("入账余额计算错误")
		}
	} else {
		if ft.BalanceAfter != ft.BalanceBefore - ft.Amount {
			return fmt.Errorf("出账余额计算错误")
		}
	}
	
	return nil
}

// Validate 验证单笔充值请求
func (dr *DepositRequest) Validate() error {
	if dr.MerchantID == 0 {
		return fmt.Errorf("商户ID不能为空")
	}
	if dr.Amount <= 0 {
		return fmt.Errorf("充值金额必须大于0")
	}
	if dr.Amount > 1000000 { // 单笔最大100万
		return fmt.Errorf("单笔充值金额不能超过1,000,000")
	}
	if len(dr.Currency) != 3 {
		return fmt.Errorf("货币代码必须为3位")
	}
	return nil
}

// Validate 验证批量充值请求
func (bdr *BatchDepositRequest) Validate() error {
	if len(bdr.Deposits) == 0 {
		return fmt.Errorf("批量充值列表不能为空")
	}
	if len(bdr.Deposits) > 100 {
		return fmt.Errorf("单次批量充值不能超过100笔")
	}
	
	totalAmount := 0.0
	for i, deposit := range bdr.Deposits {
		if err := deposit.Validate(); err != nil {
			return fmt.Errorf("第%d笔充值验证失败: %v", i+1, err)
		}
		totalAmount += deposit.Amount
	}
	
	if totalAmount > 10000000 { // 批量总额最大1000万
		return fmt.Errorf("批量充值总金额不能超过10,000,000")
	}
	
	return nil
}

// Validate 验证权益分配请求
func (ar *AllocateRequest) Validate() error {
	if ar.MerchantID == 0 {
		return fmt.Errorf("商户ID不能为空")
	}
	if ar.Amount <= 0 {
		return fmt.Errorf("分配金额必须大于0")
	}
	if ar.Amount > 1000000 { // 单次分配最大100万
		return fmt.Errorf("单次分配金额不能超过1,000,000")
	}
	return nil
}
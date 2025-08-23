package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// PricingRuleType 定价规则类型
type PricingRuleType string

const (
	PricingRuleTypeBasePrice        PricingRuleType = "base_price"        // 基础价格
	PricingRuleTypeVolumeDiscount   PricingRuleType = "volume_discount"   // 阶梯价格
	PricingRuleTypeMemberDiscount   PricingRuleType = "member_discount"   // 会员价格
	PricingRuleTypeTimeBasedDiscount PricingRuleType = "time_based_discount" // 时段优惠
)

// RightsRuleType 权益消耗规则类型
type RightsRuleType string

const (
	RightsRuleTypeFixedRate  RightsRuleType = "fixed_rate"  // 固定费率
	RightsRuleTypePercentage RightsRuleType = "percentage"  // 百分比扣减
	RightsRuleTypeTiered     RightsRuleType = "tiered"      // 阶梯消耗
)

// InsufficientRightsAction 权益不足处理策略
type InsufficientRightsAction string

const (
	InsufficientRightsActionBlockPurchase  InsufficientRightsAction = "block_purchase"  // 阻止购买
	InsufficientRightsActionPartialPayment InsufficientRightsAction = "partial_payment" // 部分支付
	InsufficientRightsActionCashPayment    InsufficientRightsAction = "cash_payment"    // 现金补足
)

// PricingConfig 定价配置接口
type PricingConfig interface {
	GetType() PricingRuleType
	Validate() error
}

// BasePriceConfig 基础价格配置
type BasePriceConfig struct {
	Amount   int64  `json:"amount"`   // 价格金额（分）
	Currency string `json:"currency"` // 货币代码
}

func (c BasePriceConfig) GetType() PricingRuleType {
	return PricingRuleTypeBasePrice
}

func (c BasePriceConfig) Validate() error {
	if c.Amount < 0 {
		return fmt.Errorf("价格金额不能为负数")
	}
	if c.Currency == "" {
		return fmt.Errorf("货币代码不能为空")
	}
	return nil
}

// VolumeDiscountTier 阶梯定价层级
type VolumeDiscountTier struct {
	MinQuantity int   `json:"min_quantity"` // 最小数量
	MaxQuantity int   `json:"max_quantity"` // 最大数量（0表示无限制）
	Price       Money `json:"price"`        // 该层级价格
}

// VolumeDiscountConfig 阶梯定价配置
type VolumeDiscountConfig struct {
	Tiers []VolumeDiscountTier `json:"tiers"`
}

func (c VolumeDiscountConfig) GetType() PricingRuleType {
	return PricingRuleTypeVolumeDiscount
}

func (c VolumeDiscountConfig) Validate() error {
	if len(c.Tiers) == 0 {
		return fmt.Errorf("阶梯定价至少需要一个层级")
	}
	for i, tier := range c.Tiers {
		if tier.MinQuantity < 0 {
			return fmt.Errorf("第%d层级最小数量不能为负数", i+1)
		}
		if tier.MaxQuantity > 0 && tier.MaxQuantity < tier.MinQuantity {
			return fmt.Errorf("第%d层级最大数量不能小于最小数量", i+1)
		}
	}
	return nil
}

// MemberDiscountConfig 会员定价配置
type MemberDiscountConfig struct {
	MemberLevels map[string]Money `json:"member_levels"` // 会员等级对应价格
	DefaultPrice Money            `json:"default_price"`  // 默认价格
}

func (c MemberDiscountConfig) GetType() PricingRuleType {
	return PricingRuleTypeMemberDiscount
}

func (c MemberDiscountConfig) Validate() error {
	if len(c.MemberLevels) == 0 {
		return fmt.Errorf("会员定价至少需要一个会员等级")
	}
	return nil
}

// TimeBasedDiscountConfig 时段优惠配置
type TimeBasedDiscountConfig struct {
	TimeSlots []TimeSlot `json:"time_slots"`
}

// TimeSlot 时间段
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Price     Money     `json:"price"`
	WeekDays  []int     `json:"week_days"` // 0-6，0为周日
}

func (c TimeBasedDiscountConfig) GetType() PricingRuleType {
	return PricingRuleTypeTimeBasedDiscount
}

func (c TimeBasedDiscountConfig) Validate() error {
	if len(c.TimeSlots) == 0 {
		return fmt.Errorf("时段优惠至少需要一个时间段")
	}
	for i, slot := range c.TimeSlots {
		if slot.EndTime.Before(slot.StartTime) {
			return fmt.Errorf("第%d时段结束时间不能早于开始时间", i+1)
		}
	}
	return nil
}

// PricingRuleConfig 定价规则配置包装器，用于JSON存储
type PricingRuleConfig struct {
	Type   PricingRuleType `json:"type"`
	Config json.RawMessage `json:"config"`
}

// Value 实现 driver.Valuer 接口
func (p PricingRuleConfig) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口
func (p *PricingRuleConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	}
	return fmt.Errorf("无法扫描 %T 到 PricingRuleConfig", value)
}

// GetConfig 获取具体的配置对象
func (p *PricingRuleConfig) GetConfig() (PricingConfig, error) {
	switch p.Type {
	case PricingRuleTypeBasePrice:
		var config BasePriceConfig
		err := json.Unmarshal(p.Config, &config)
		return config, err
	case PricingRuleTypeVolumeDiscount:
		var config VolumeDiscountConfig
		err := json.Unmarshal(p.Config, &config)
		return config, err
	case PricingRuleTypeMemberDiscount:
		var config MemberDiscountConfig
		err := json.Unmarshal(p.Config, &config)
		return config, err
	case PricingRuleTypeTimeBasedDiscount:
		var config TimeBasedDiscountConfig
		err := json.Unmarshal(p.Config, &config)
		return config, err
	default:
		return nil, fmt.Errorf("未知的定价规则类型: %s", p.Type)
	}
}

// SetConfig 设置配置对象
func (p *PricingRuleConfig) SetConfig(config PricingConfig) error {
	p.Type = config.GetType()
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}
	p.Config = configJSON
	return nil
}

// ProductPricingRule 商品定价规则实体
type ProductPricingRule struct {
	ID         uint64            `json:"id" db:"id"`
	TenantID   uint64            `json:"tenant_id" db:"tenant_id"`
	ProductID  uint64            `json:"product_id" db:"product_id"`
	RuleType   PricingRuleType   `json:"rule_type" db:"rule_type"`
	RuleConfig PricingRuleConfig `json:"rule_config" db:"rule_config"`
	Priority   int               `json:"priority" db:"priority"`
	IsActive   bool              `json:"is_active" db:"is_active"`
	ValidFrom  time.Time         `json:"valid_from" db:"valid_from"`
	ValidUntil *time.Time        `json:"valid_until,omitempty" db:"valid_until"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" db:"updated_at"`
}

// TableName 指定表名
func (ProductPricingRule) TableName() string {
	return "product_pricing_rules"
}

// ProductRightsRule 商品权益消耗规则实体
type ProductRightsRule struct {
	ID                       uint64                   `json:"id" db:"id"`
	TenantID                 uint64                   `json:"tenant_id" db:"tenant_id"`
	ProductID                uint64                   `json:"product_id" db:"product_id"`
	RuleType                 RightsRuleType           `json:"rule_type" db:"rule_type"`
	ConsumptionRate          float64                  `json:"consumption_rate" db:"consumption_rate"`
	MinRightsRequired        float64                  `json:"min_rights_required" db:"min_rights_required"`
	InsufficientRightsAction InsufficientRightsAction `json:"insufficient_rights_action" db:"insufficient_rights_action"`
	IsActive                 bool                     `json:"is_active" db:"is_active"`
	CreatedAt                time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time                `json:"updated_at" db:"updated_at"`
}

// TableName 指定表名
func (ProductRightsRule) TableName() string {
	return "product_rights_rules"
}

// PromotionCondition 促销条件
type PromotionCondition struct {
	Type  string      `json:"type"`  // 条件类型：quantity, member_level, time_range等
	Value interface{} `json:"value"` // 条件值
}

// PromotionConditions 促销条件数组
type PromotionConditions []PromotionCondition

// Value 实现 driver.Valuer 接口
func (p PromotionConditions) Value() (driver.Value, error) {
	if len(p) == 0 {
		return json.Marshal([]PromotionCondition{})
	}
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口
func (p *PromotionConditions) Scan(value interface{}) error {
	if value == nil {
		*p = PromotionConditions{}
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	}
	return nil
}

// PromotionalPrice 促销价格实体
type PromotionalPrice struct {
	ID                 uint64               `json:"id" db:"id"`
	TenantID           uint64               `json:"tenant_id" db:"tenant_id"`
	ProductID          uint64               `json:"product_id" db:"product_id"`
	PromotionalPrice   Money                `json:"promotional_price" db:"promotional_price"`
	DiscountPercentage *float64             `json:"discount_percentage,omitempty" db:"discount_percentage"`
	ValidFrom          time.Time            `json:"valid_from" db:"valid_from"`
	ValidUntil         time.Time            `json:"valid_until" db:"valid_until"`
	Conditions         PromotionConditions  `json:"conditions" db:"conditions"`
	IsActive           bool                 `json:"is_active" db:"is_active"`
	CreatedAt          time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at" db:"updated_at"`
}

// TableName 指定表名
func (PromotionalPrice) TableName() string {
	return "promotional_prices"
}

// PriceHistory 价格变更历史实体
type PriceHistory struct {
	ID            uint64    `json:"id" db:"id"`
	TenantID      uint64    `json:"tenant_id" db:"tenant_id"`
	ProductID     uint64    `json:"product_id" db:"product_id"`
	OldPrice      Money     `json:"old_price" db:"old_price"`
	NewPrice      Money     `json:"new_price" db:"new_price"`
	ChangeReason  string    `json:"change_reason" db:"change_reason"`
	ChangedBy     uint64    `json:"changed_by" db:"changed_by"`
	EffectiveDate time.Time `json:"effective_date" db:"effective_date"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// TableName 指定表名
func (PriceHistory) TableName() string {
	return "price_histories"
}

// CreatePricingRuleRequest 创建定价规则请求
type CreatePricingRuleRequest struct {
	RuleType   PricingRuleType   `json:"rule_type" validate:"required"`
	RuleConfig PricingRuleConfig `json:"rule_config" validate:"required"`
	Priority   int               `json:"priority" validate:"min=0"`
	ValidFrom  time.Time         `json:"valid_from" validate:"required"`
	ValidUntil *time.Time        `json:"valid_until,omitempty"`
}

// UpdatePricingRuleRequest 更新定价规则请求
type UpdatePricingRuleRequest struct {
	RuleConfig *PricingRuleConfig `json:"rule_config,omitempty"`
	Priority   *int               `json:"priority,omitempty" validate:"min=0"`
	IsActive   *bool              `json:"is_active,omitempty"`
	ValidFrom  *time.Time         `json:"valid_from,omitempty"`
	ValidUntil *time.Time         `json:"valid_until,omitempty"`
}

// CreateRightsRuleRequest 创建权益规则请求
type CreateRightsRuleRequest struct {
	RuleType                 RightsRuleType           `json:"rule_type" validate:"required"`
	ConsumptionRate          float64                  `json:"consumption_rate" validate:"min=0"`
	MinRightsRequired        float64                  `json:"min_rights_required" validate:"min=0"`
	InsufficientRightsAction InsufficientRightsAction `json:"insufficient_rights_action" validate:"required"`
}

// UpdateRightsRuleRequest 更新权益规则请求
type UpdateRightsRuleRequest struct {
	ConsumptionRate          *float64                  `json:"consumption_rate,omitempty" validate:"min=0"`
	MinRightsRequired        *float64                  `json:"min_rights_required,omitempty" validate:"min=0"`
	InsufficientRightsAction *InsufficientRightsAction `json:"insufficient_rights_action,omitempty"`
	IsActive                 *bool                     `json:"is_active,omitempty"`
}

// CreatePromotionalPriceRequest 创建促销价格请求
type CreatePromotionalPriceRequest struct {
	PromotionalPrice   Money               `json:"promotional_price" validate:"required"`
	DiscountPercentage *float64            `json:"discount_percentage,omitempty" validate:"min=0,max=100"`
	ValidFrom          time.Time           `json:"valid_from" validate:"required"`
	ValidUntil         time.Time           `json:"valid_until" validate:"required"`
	Conditions         PromotionConditions `json:"conditions,omitempty"`
}

// UpdatePromotionalPriceRequest 更新促销价格请求
type UpdatePromotionalPriceRequest struct {
	PromotionalPrice   *Money              `json:"promotional_price,omitempty"`
	DiscountPercentage *float64            `json:"discount_percentage,omitempty" validate:"min=0,max=100"`
	ValidFrom          *time.Time          `json:"valid_from,omitempty"`
	ValidUntil         *time.Time          `json:"valid_until,omitempty"`
	Conditions         *PromotionConditions `json:"conditions,omitempty"`
	IsActive           *bool               `json:"is_active,omitempty"`
}

// PriceChangeRequest 价格变更请求
type PriceChangeRequest struct {
	NewPrice      Money  `json:"new_price" validate:"required"`
	ChangeReason  string `json:"change_reason" validate:"required,max=255"`
	EffectiveDate time.Time `json:"effective_date" validate:"required"`
}

// ValidateRightsRequest 验证权益请求
type ValidateRightsRequest struct {
	UserID    uint64 `json:"user_id" validate:"required"`
	ProductID uint64 `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"min=1"`
	TotalAmount Money `json:"total_amount" validate:"required"`
}

// ValidateRightsResponse 验证权益响应
type ValidateRightsResponse struct {
	IsValid             bool                     `json:"is_valid"`
	RequiredRights      float64                  `json:"required_rights"`
	AvailableRights     float64                  `json:"available_rights"`
	InsufficientAmount  float64                  `json:"insufficient_amount,omitempty"`
	SuggestedAction     InsufficientRightsAction `json:"suggested_action,omitempty"`
	CashPaymentRequired Money                    `json:"cash_payment_required,omitempty"`
}

// CalculateEffectivePriceRequest 计算有效价格请求
type CalculateEffectivePriceRequest struct {
	UserID      *uint64   `json:"user_id,omitempty"`
	Quantity    int       `json:"quantity" validate:"min=1"`
	MemberLevel *string   `json:"member_level,omitempty"`
	RequestTime time.Time `json:"request_time"`
}

// CalculateEffectivePriceResponse 计算有效价格响应
type CalculateEffectivePriceResponse struct {
	EffectivePrice      Money               `json:"effective_price"`
	BasePrice           Money               `json:"base_price"`
	AppliedRules        []string            `json:"applied_rules"`
	DiscountAmount      Money               `json:"discount_amount,omitempty"`
	PromotionalPrice    *Money              `json:"promotional_price,omitempty"`
	RightsConsumption   float64             `json:"rights_consumption"`
	IsPromotionActive   bool                `json:"is_promotion_active"`
}

// ProcessRightsRequest 权益处理请求
type ProcessRightsRequest struct {
	UserID      uint64 `json:"user_id" validate:"required"`
	ProductID   uint64 `json:"product_id" validate:"required"`
	Quantity    uint32 `json:"quantity" validate:"min=1"`
	TotalAmount Money  `json:"total_amount" validate:"required"`
}

// ProcessRightsResponse 权益处理响应
type ProcessRightsResponse struct {
	Success          bool   `json:"success"`
	ConsumedRights   float64 `json:"consumed_rights"`
	RemainingRights  float64 `json:"remaining_rights"`
	CashPayment      Money  `json:"cash_payment"`
	ProcessingAction string `json:"processing_action"`
}

// RightsConsumptionStats 权益消耗统计
type RightsConsumptionStats struct {
	ProductID        uint64                  `json:"product_id"`
	TotalConsumption float64                 `json:"total_consumption"`
	DailyAverage     float64                 `json:"daily_average"`
	MonthlyTrend     []ConsumptionTrendData `json:"monthly_trend"`
	UserSegmentation []UserSegmentData      `json:"user_segmentation"`
}

// ConsumptionTrendData 消耗趋势数据
type ConsumptionTrendData struct {
	Date        string  `json:"date"`
	Consumption float64 `json:"consumption"`
}

// UserSegmentData 用户分群数据
type UserSegmentData struct {
	Level       string  `json:"level"`
	Consumption float64 `json:"consumption"`
	Count       int     `json:"count"`
}
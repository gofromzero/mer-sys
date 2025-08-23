package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// OrderPricingService 订单定价服务
type OrderPricingService struct {
	pricingService      *PricingService
	rightsValidator     *RightsValidatorService
	pricingValidator    *PricingValidatorService
	productRepo         *repository.ProductRepository
	pricingRuleRepo     *repository.PricingRuleRepository
	rightsRuleRepo      *repository.RightsRuleRepository
	promotionalPriceRepo *repository.PromotionalPriceRepository
}

// NewOrderPricingService 创建订单定价服务
func NewOrderPricingService() *OrderPricingService {
	return &OrderPricingService{
		pricingService:       NewPricingService(),
		rightsValidator:      NewRightsValidatorService(),
		pricingValidator:     NewPricingValidatorService(),
		productRepo:          repository.NewProductRepository(),
		pricingRuleRepo:      repository.NewPricingRuleRepository(),
		rightsRuleRepo:       repository.NewRightsRuleRepository(),
		promotionalPriceRepo: repository.NewPromotionalPriceRepository(),
	}
}

// OrderItemPricingRequest 订单商品定价请求
type OrderItemPricingRequest struct {
	ProductID   uint64  `json:"product_id"`
	Quantity    uint32  `json:"quantity"`
	UserID      uint64  `json:"user_id"`
	MemberLevel string  `json:"member_level"`
	OrderTime   time.Time `json:"order_time"`
}

// OrderPricingRequest 订单定价请求
type OrderPricingRequest struct {
	UserID      uint64                        `json:"user_id"`
	Items       []OrderItemPricingRequest    `json:"items"`
	OrderTime   time.Time                    `json:"order_time"`
	MemberLevel string                       `json:"member_level"`
}

// OrderItemPricingResult 订单商品定价结果
type OrderItemPricingResult struct {
	ProductID           uint64  `json:"product_id"`
	Quantity            uint32  `json:"quantity"`
	OriginalPrice       types.Money `json:"original_price"`       // 原价
	EffectivePrice      types.Money `json:"effective_price"`      // 有效价格
	TotalPrice          types.Money `json:"total_price"`          // 总价
	DiscountAmount      types.Money `json:"discount_amount"`      // 折扣金额
	AppliedRules        []string `json:"applied_rules"`        // 应用的规则
	IsPromotionApplied  bool     `json:"is_promotion_applied"`  // 是否应用促销
	PromotionalPrice    *types.Money `json:"promotional_price,omitempty"` // 促销价格
	RightsConsumption   float64  `json:"rights_consumption"`    // 权益消耗
	CashPayment         types.Money `json:"cash_payment"`         // 现金支付金额
	RightsPayment       float64  `json:"rights_payment"`        // 权益支付金额
	ProcessingAction    string   `json:"processing_action"`     // 处理动作
}

// OrderPricingResult 订单定价结果
type OrderPricingResult struct {
	UserID              uint64                      `json:"user_id"`
	Items               []OrderItemPricingResult   `json:"items"`
	TotalOriginalAmount types.Money                `json:"total_original_amount"` // 原始总金额
	TotalEffectiveAmount types.Money               `json:"total_effective_amount"` // 有效总金额
	TotalDiscountAmount types.Money                `json:"total_discount_amount"` // 总折扣金额
	TotalCashPayment    types.Money                `json:"total_cash_payment"`    // 总现金支付
	TotalRightsPayment  float64                     `json:"total_rights_payment"`  // 总权益支付
	CanProceed          bool                        `json:"can_proceed"`           // 是否可以继续
	BlockedReasons      []string                    `json:"blocked_reasons,omitempty"` // 阻止原因
	Warnings            []string                    `json:"warnings,omitempty"`    // 警告信息
}

// CalculateOrderPricing 计算订单定价
func (s *OrderPricingService) CalculateOrderPricing(ctx context.Context, req *OrderPricingRequest) (*OrderPricingResult, error) {
	result := &OrderPricingResult{
		UserID:              req.UserID,
		Items:               make([]OrderItemPricingResult, 0, len(req.Items)),
		TotalOriginalAmount: types.Money{Amount: 0, Currency: "CNY"},
		TotalEffectiveAmount: types.Money{Amount: 0, Currency: "CNY"},
		TotalDiscountAmount: types.Money{Amount: 0, Currency: "CNY"},
		TotalCashPayment:    types.Money{Amount: 0, Currency: "CNY"},
		TotalRightsPayment:  0,
		CanProceed:          true,
		BlockedReasons:      make([]string, 0),
		Warnings:           make([]string, 0),
	}

	// 遍历每个商品项计算定价
	for _, item := range req.Items {
		itemResult, err := s.calculateItemPricing(ctx, &item)
		if err != nil {
			return nil, gerror.Wrapf(err, "计算商品%d定价失败", item.ProductID)
		}

		result.Items = append(result.Items, *itemResult)

		// 累计总金额
		result.TotalOriginalAmount.Amount += itemResult.OriginalPrice.Amount * float64(itemResult.Quantity)
		result.TotalEffectiveAmount.Amount += itemResult.TotalPrice.Amount
		result.TotalDiscountAmount.Amount += itemResult.DiscountAmount.Amount * float64(itemResult.Quantity)
		result.TotalCashPayment.Amount += itemResult.CashPayment.Amount
		result.TotalRightsPayment += itemResult.RightsPayment

		// 检查是否被阻止
		if itemResult.ProcessingAction == "purchase_blocked" {
			result.CanProceed = false
			result.BlockedReasons = append(result.BlockedReasons, 
				fmt.Sprintf("商品%d权益不足", item.ProductID))
		}

		// 添加警告信息
		if itemResult.ProcessingAction == "partial_payment" {
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("商品%d权益不足，需要部分现金支付", item.ProductID))
		}
	}

	// 应用订单级别的优惠（如果有的话）
	if err := s.applyOrderLevelDiscounts(ctx, result); err != nil {
		g.Log().Warningf(ctx, "应用订单级别优惠失败: %v", err)
		result.Warnings = append(result.Warnings, "订单级别优惠应用失败")
	}

	return result, nil
}

// calculateItemPricing 计算单个商品的定价
func (s *OrderPricingService) calculateItemPricing(ctx context.Context, req *OrderItemPricingRequest) (*OrderItemPricingResult, error) {
	// 1. 获取商品基础信息
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取商品信息失败")
	}

	if product == nil {
		return nil, gerror.New("商品不存在")
	}

	// 检查库存
	if product.InventoryInfo != nil && product.InventoryInfo.TrackInventory && 
		uint32(product.InventoryInfo.StockQuantity) < req.Quantity {
		return nil, gerror.Newf("商品库存不足，可用:%d，需要:%d", 
			product.InventoryInfo.StockQuantity, req.Quantity)
	}

	result := &OrderItemPricingResult{
		ProductID:      req.ProductID,
		Quantity:       req.Quantity,
		OriginalPrice:  product.GetPrice(),
		CashPayment:    types.Money{Amount: 0, Currency: product.PriceCurrency},
		RightsPayment:  0,
	}

	// 2. 计算有效价格
	priceReq := &types.CalculateEffectivePriceRequest{
		UserID:      &req.UserID,
		Quantity:    int(req.Quantity),
		MemberLevel: &req.MemberLevel,
		RequestTime: req.OrderTime,
	}

	priceResp, err := s.pricingService.CalculateEffectivePrice(ctx, req.ProductID, priceReq)
	if err != nil {
		return nil, gerror.Wrap(err, "计算有效价格失败")
	}

	result.EffectivePrice = priceResp.EffectivePrice
	result.TotalPrice = types.Money{
		Amount:   priceResp.EffectivePrice.Amount * float64(req.Quantity),
		Currency: priceResp.EffectivePrice.Currency,
	}
	result.DiscountAmount = priceResp.DiscountAmount
	result.AppliedRules = priceResp.AppliedRules
	result.IsPromotionApplied = priceResp.IsPromotionActive
	result.PromotionalPrice = priceResp.PromotionalPrice

	// 3. 处理权益消耗
	rightsReq := &types.ProcessRightsRequest{
		UserID:      req.UserID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		TotalAmount: result.TotalPrice,
	}

	rightsResp, err := s.rightsValidator.ProcessRightsConsumption(ctx, rightsReq)
	if err != nil {
		g.Log().Warningf(ctx, "处理权益消耗失败: %v", err)
		// 权益处理失败时，默认全现金支付
		result.CashPayment = result.TotalPrice
		result.ProcessingAction = "cash_payment"
	} else {
		result.RightsConsumption = priceResp.RightsConsumption
		result.RightsPayment = rightsResp.ConsumedRights
		result.CashPayment = rightsResp.CashPayment
		result.ProcessingAction = rightsResp.ProcessingAction
	}

	return result, nil
}

// applyOrderLevelDiscounts 应用订单级别的优惠
func (s *OrderPricingService) applyOrderLevelDiscounts(ctx context.Context, result *OrderPricingResult) error {
	// 这里可以实现订单级别的优惠逻辑，比如：
	// - 满额减免
	// - 全场折扣
	// - 会员专属优惠
	// - 首单优惠等

	// 示例：满500减50
	if result.TotalEffectiveAmount.Amount >= 500 {
		discountAmount := 50.0
		result.TotalEffectiveAmount.Amount -= discountAmount
		result.TotalDiscountAmount.Amount += discountAmount
		result.TotalCashPayment.Amount -= discountAmount
		
		result.Warnings = append(result.Warnings, "已应用满500减50优惠")
	}

	return nil
}

// ValidateOrderPricing 验证订单定价结果
func (s *OrderPricingService) ValidateOrderPricing(ctx context.Context, result *OrderPricingResult) error {
	// 1. 验证金额一致性
	calculatedTotal := 0.0
	for _, item := range result.Items {
		calculatedTotal += item.TotalPrice.Amount
	}

	if absFloat(calculatedTotal-result.TotalEffectiveAmount.Amount) > 0.01 {
		return gerror.Newf("订单总金额不一致: 计算值%.2f, 实际值%.2f", 
			calculatedTotal, result.TotalEffectiveAmount.Amount)
	}

	// 2. 验证现金支付金额不能为负数
	if result.TotalCashPayment.Amount < 0 {
		return gerror.New("现金支付金额不能为负数")
	}

	// 3. 验证权益支付金额不能为负数
	if result.TotalRightsPayment < 0 {
		return gerror.New("权益支付金额不能为负数")
	}

	// 4. 验证现金支付 + 权益支付的等价金额 = 总金额
	// 这里需要权益到现金的换算比例
	rightsToMoneyRate := 0.01 // 1权益点 = 0.01元
	totalPayment := result.TotalCashPayment.Amount + (result.TotalRightsPayment * rightsToMoneyRate)
	
	if absFloat(totalPayment-result.TotalEffectiveAmount.Amount) > 0.01 {
		return gerror.Newf("支付金额不匹配: 总支付%.2f, 订单金额%.2f", 
			totalPayment, result.TotalEffectiveAmount.Amount)
	}

	return nil
}

// ProcessOrderPricing 处理订单定价（执行权益扣减等操作）
func (s *OrderPricingService) ProcessOrderPricing(ctx context.Context, result *OrderPricingResult) error {
	// 这个方法在订单确认后调用，执行实际的权益扣减
	// 在计算阶段，我们只是模拟，不实际扣减

	if !result.CanProceed {
		return gerror.New("订单无法继续处理")
	}

	// 1. 验证定价结果
	if err := s.ValidateOrderPricing(ctx, result); err != nil {
		return gerror.Wrap(err, "订单定价验证失败")
	}

	// 2. 创建价格快照（用于订单记录）
	snapshot := s.createPricingSnapshot(result)
	
	// 3. 记录定价日志
	s.logPricingDecision(ctx, result)

	// 存储快照到上下文，供订单服务使用
	ctx = context.WithValue(ctx, "pricing_snapshot", snapshot)

	g.Log().Infof(ctx, "订单定价处理完成，用户%d，总金额%.2f，权益支付%.2f", 
		result.UserID, result.TotalEffectiveAmount.Amount, result.TotalRightsPayment)

	return nil
}

// createPricingSnapshot 创建定价快照
func (s *OrderPricingService) createPricingSnapshot(result *OrderPricingResult) map[string]interface{} {
	return map[string]interface{}{
		"user_id":               result.UserID,
		"total_original_amount": result.TotalOriginalAmount,
		"total_effective_amount": result.TotalEffectiveAmount,
		"total_discount_amount":  result.TotalDiscountAmount,
		"total_cash_payment":     result.TotalCashPayment,
		"total_rights_payment":   result.TotalRightsPayment,
		"items":                  result.Items,
		"calculated_at":          time.Now(),
	}
}

// logPricingDecision 记录定价决策日志
func (s *OrderPricingService) logPricingDecision(ctx context.Context, result *OrderPricingResult) {
	logData := map[string]interface{}{
		"user_id":           result.UserID,
		"item_count":        len(result.Items),
		"total_amount":      result.TotalEffectiveAmount.Amount,
		"discount_amount":   result.TotalDiscountAmount.Amount,
		"rights_payment":    result.TotalRightsPayment,
		"can_proceed":       result.CanProceed,
		"blocked_reasons":   result.BlockedReasons,
		"warnings":          result.Warnings,
	}

	g.Log().Infof(ctx, "订单定价决策: %+v", logData)
}

// absFloat 返回浮点数的绝对值
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
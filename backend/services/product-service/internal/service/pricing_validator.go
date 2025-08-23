package service

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// PricingValidatorService 定价规则验证服务
type PricingValidatorService struct {
	pricingRuleRepo *repository.PricingRuleRepository
	productRepo     *repository.ProductRepository
}

// NewPricingValidatorService 创建定价规则验证服务
func NewPricingValidatorService() *PricingValidatorService {
	return &PricingValidatorService{
		pricingRuleRepo: repository.NewPricingRuleRepository(),
		productRepo:     repository.NewProductRepository(),
	}
}

// ValidateRuleConfig 验证定价规则配置
func (s *PricingValidatorService) ValidateRuleConfig(ctx context.Context, rule *types.ProductPricingRule) error {
	switch rule.RuleType {
	case types.PricingRuleTypeBasePrice:
		return s.validateBasePriceConfig(rule.RuleConfig)
	case types.PricingRuleTypeVolumeDiscount:
		return s.validateVolumeDiscountConfig(rule.RuleConfig)
	case types.PricingRuleTypeMemberDiscount:
		return s.validateMemberDiscountConfig(rule.RuleConfig)
	case types.PricingRuleTypeTimeBasedDiscount:
		return s.validateTimeBasedDiscountConfig(rule.RuleConfig)
	default:
		return gerror.Newf("不支持的定价规则类型: %s", rule.RuleType)
	}
}

// ValidateRuleConflicts 验证规则冲突
func (s *PricingValidatorService) ValidateRuleConflicts(ctx context.Context, productID uint64, newRule *types.ProductPricingRule) error {
	// 获取产品的所有有效规则
	existingRules, err := s.pricingRuleRepo.GetActivePricingRules(ctx, productID, time.Now())
	if err != nil {
		return gerror.Wrap(err, "获取现有定价规则失败")
	}

	// 检查基础价格规则唯一性
	if newRule.RuleType == types.PricingRuleTypeBasePrice {
		for _, rule := range existingRules {
			if rule.RuleType == types.PricingRuleTypeBasePrice && rule.ID != newRule.ID {
				return gerror.New("产品只能有一个基础价格规则")
			}
		}
	}

	// 检查时间重叠冲突
	return s.validateTimeOverlapConflicts(newRule, existingRules)
}

// ValidateBusinessRules 验证业务规则
func (s *PricingValidatorService) ValidateBusinessRules(ctx context.Context, productID uint64, rule *types.ProductPricingRule) error {
	// 获取产品信息
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return gerror.Wrap(err, "获取产品信息失败")
	}

	if product == nil {
		return gerror.New("产品不存在")
	}

	// 验证价格合理性
	if err := s.validatePriceReasonableness(rule, product); err != nil {
		return err
	}

	// 验证库存要求
	if err := s.validateInventoryRequirements(rule, product); err != nil {
		return err
	}

	return nil
}

// validateBasePriceConfig 验证基础价格配置
func (s *PricingValidatorService) validateBasePriceConfig(config types.PricingRuleConfig) error {
	configObj, err := config.GetConfig()
	if err != nil {
		return gerror.Wrap(err, "解析配置失败")
	}
	
	baseConfig, ok := configObj.(types.BasePriceConfig)
	if !ok {
		return gerror.New("基础价格配置格式错误")
	}

	if baseConfig.Amount <= 0 {
		return gerror.New("基础价格必须大于0")
	}

	if baseConfig.Currency == "" {
		return gerror.New("货币类型不能为空")
	}

	// 验证货币类型
	validCurrencies := map[string]bool{"CNY": true, "USD": true}
	if !validCurrencies[baseConfig.Currency] {
		return gerror.Newf("不支持的货币类型: %s", baseConfig.Currency)
	}

	return nil
}

// validateVolumeDiscountConfig 验证阶梯价格配置
func (s *PricingValidatorService) validateVolumeDiscountConfig(config types.PricingRuleConfig) error {
	configObj, err := config.GetConfig()
	if err != nil {
		return gerror.Wrap(err, "解析配置失败")
	}
	
	volumeConfig, ok := configObj.(types.VolumeDiscountConfig)
	if !ok {
		return gerror.New("阶梯价格配置格式错误")
	}

	if len(volumeConfig.Tiers) == 0 {
		return gerror.New("阶梯价格至少需要一个层级")
	}

	// 简化验证：只检查基本条件
	for i, tier := range volumeConfig.Tiers {
		if tier.MinQuantity <= 0 {
			return gerror.Newf("第%d个层级的最小数量必须大于0", i+1)
		}
		if tier.Price.Amount <= 0 {
			return gerror.Newf("第%d个层级的价格必须大于0", i+1)
		}
	}

	return nil
}

// validateMemberDiscountConfig 验证会员价格配置
func (s *PricingValidatorService) validateMemberDiscountConfig(config types.PricingRuleConfig) error {
	configObj, err := config.GetConfig()
	if err != nil {
		return gerror.Wrap(err, "解析配置失败")
	}
	
	memberConfig, ok := configObj.(types.MemberDiscountConfig)
	if !ok {
		return gerror.New("会员价格配置格式错误")
	}

	if len(memberConfig.MemberLevels) == 0 {
		return gerror.New("会员价格至少需要一个等级配置")
	}

	if memberConfig.DefaultPrice.Amount <= 0 {
		return gerror.New("默认价格必须大于0")
	}

	return nil
}

// validateTimeBasedDiscountConfig 验证时段优惠配置
func (s *PricingValidatorService) validateTimeBasedDiscountConfig(config types.PricingRuleConfig) error {
	configObj, err := config.GetConfig()
	if err != nil {
		return gerror.Wrap(err, "解析配置失败")
	}
	
	timeConfig, ok := configObj.(types.TimeBasedDiscountConfig)
	if !ok {
		return gerror.New("时段优惠配置格式错误")
	}

	if len(timeConfig.TimeSlots) == 0 {
		return gerror.New("时段优惠至少需要一个时段配置")
	}

	// 验证时段配置
	for i, slot := range timeConfig.TimeSlots {
		if slot.Price.Amount <= 0 {
			return gerror.Newf("第%d个时段的价格必须大于0", i+1)
		}

		if len(slot.WeekDays) == 0 {
			return gerror.Newf("第%d个时段必须指定适用的星期", i+1)
		}
	}

	return nil
}

// validateTimeOverlapConflicts 验证时间重叠冲突
func (s *PricingValidatorService) validateTimeOverlapConflicts(newRule *types.ProductPricingRule, existingRules []*types.ProductPricingRule) error {
	for _, existingRule := range existingRules {
		if existingRule.ID == newRule.ID {
			continue
		}

		// 检查时间重叠
		if s.hasTimeOverlap(newRule, existingRule) {
			// 同类型规则不能时间重叠
			if newRule.RuleType == existingRule.RuleType {
				return gerror.Newf("与现有规则(ID: %d)时间重叠", existingRule.ID)
			}

			// 检查优先级
			if newRule.Priority == existingRule.Priority {
				return gerror.Newf("与现有规则(ID: %d)优先级相同且时间重叠", existingRule.ID)
			}
		}
	}

	return nil
}

// hasTimeOverlap 检查两个规则是否有时间重叠
func (s *PricingValidatorService) hasTimeOverlap(rule1, rule2 *types.ProductPricingRule) bool {
	// 如果其中一个规则没有结束时间，认为永久有效
	rule1End := rule1.ValidUntil
	if rule1End == nil {
		futureTime := time.Now().AddDate(100, 0, 0)
		rule1End = &futureTime
	}

	rule2End := rule2.ValidUntil
	if rule2End == nil {
		futureTime := time.Now().AddDate(100, 0, 0)
		rule2End = &futureTime
	}

	// 检查时间区间重叠
	return rule1.ValidFrom.Before(*rule2End) && rule2.ValidFrom.Before(*rule1End)
}

// validatePriceReasonableness 验证价格合理性
func (s *PricingValidatorService) validatePriceReasonableness(rule *types.ProductPricingRule, product *types.Product) error {
	// 简化价格合理性验证
	currentPrice := product.PriceAmount

	if currentPrice <= 0 {
		return gerror.New("产品当前价格无效")
	}

	return nil
}

// validateInventoryRequirements 验证库存要求
func (s *PricingValidatorService) validateInventoryRequirements(rule *types.ProductPricingRule, product *types.Product) error {
	// 简化库存验证
	if product.InventoryInfo != nil && product.InventoryInfo.StockQuantity <= 0 {
		return gerror.New("商品库存不足")
	}

	return nil
}

// getMinPriceFromRule 从规则中获取最低价格
func (s *PricingValidatorService) getMinPriceFromRule(rule *types.ProductPricingRule) (float64, error) {
	// 简化处理，返回基础价格
	return 100.0, nil
}

// getMaxPriceFromRule 从规则中获取最高价格
func (s *PricingValidatorService) getMaxPriceFromRule(rule *types.ProductPricingRule) (float64, error) {
	// 简化处理，返回基础价格
	return 100.0, nil
}

// ValidateRuleUpdate 验证规则更新
func (s *PricingValidatorService) ValidateRuleUpdate(ctx context.Context, ruleID uint64, updates map[string]interface{}) error {
	// 获取现有规则
	existingRule, err := s.pricingRuleRepo.GetPricingRuleByID(ctx, ruleID)
	if err != nil {
		return gerror.Wrap(err, "获取现有规则失败")
	}

	if existingRule == nil {
		return gerror.New("规则不存在")
	}

	// 检查是否可以修改
	if existingRule.IsActive && time.Now().After(existingRule.ValidFrom) {
		// 已生效的规则限制修改范围
		restrictedFields := []string{"rule_type", "valid_from"}
		for _, field := range restrictedFields {
			if _, exists := updates[field]; exists {
				return gerror.Newf("已生效的规则不能修改字段: %s", field)
			}
		}
	}

	return nil
}

// ValidateRuleDeletion 验证规则删除
func (s *PricingValidatorService) ValidateRuleDeletion(ctx context.Context, ruleID uint64) error {
	// 获取规则信息
	rule, err := s.pricingRuleRepo.GetPricingRuleByID(ctx, ruleID)
	if err != nil {
		return gerror.Wrap(err, "获取规则失败")
	}

	if rule == nil {
		return gerror.New("规则不存在")
	}

	// 检查是否为基础价格规则
	if rule.RuleType == types.PricingRuleTypeBasePrice {
		// 检查是否还有其他定价规则
		otherRules, err := s.pricingRuleRepo.GetPricingRulesByProductID(ctx, rule.ProductID)
		if err != nil {
			return gerror.Wrap(err, "获取其他规则失败")
		}

		activeOtherRules := 0
		for _, otherRule := range otherRules {
			if otherRule.ID != ruleID && otherRule.IsActive {
				activeOtherRules++
			}
		}

		if activeOtherRules == 0 {
			return gerror.New("不能删除唯一的基础价格规则")
		}
	}

	// 检查是否有关联的促销活动或订单
	// 这里可以添加更复杂的依赖检查逻辑

	return nil
}
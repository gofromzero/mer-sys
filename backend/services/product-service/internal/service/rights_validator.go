package service

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// RightsValidatorService 权益验证服务
type RightsValidatorService struct {
	rightsRuleRepo *repository.RightsRuleRepository
	productRepo    *repository.ProductRepository
}

// NewRightsValidatorService 创建权益验证服务
func NewRightsValidatorService() *RightsValidatorService {
	return &RightsValidatorService{
		rightsRuleRepo: repository.NewRightsRuleRepository(),
		productRepo:    repository.NewProductRepository(),
	}
}

// ValidateRightsBalance 验证权益余额
func (s *RightsValidatorService) ValidateRightsBalance(ctx context.Context, req *types.ValidateRightsRequest) (*types.ValidateRightsResponse, error) {
	// 获取用户权益余额
	availableRights, err := s.getUserRightsBalance(ctx, req.UserID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取用户权益余额失败")
	}

	// 获取产品权益规则
	rule, err := s.rightsRuleRepo.GetRightsRuleByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取产品权益规则失败")
	}

	// 如果没有权益规则，不需要消耗权益
	if rule == nil {
		return &types.ValidateRightsResponse{
			IsValid:              true,
			RequiredRights:       0,
			AvailableRights:      availableRights,
			InsufficientAmount:   0,
			SuggestedAction:      types.InsufficientRightsActionCashPayment,
			CashPaymentRequired: types.Money{Amount: 0, Currency: "CNY"},
		}, nil
	}

	// 计算所需权益
	requiredRights, err := s.calculateRequiredRights(rule, req)
	if err != nil {
		return nil, gerror.Wrap(err, "计算所需权益失败")
	}

	// 检查是否满足最低权益要求
	if availableRights < rule.MinRightsRequired {
		return &types.ValidateRightsResponse{
			IsValid:              false,
			RequiredRights:       requiredRights,
			AvailableRights:      availableRights,
			InsufficientAmount:   rule.MinRightsRequired - availableRights,
			SuggestedAction:      rule.InsufficientRightsAction,
			CashPaymentRequired: types.Money{Amount: 0, Currency: "CNY"},
		}, nil
	}

	// 检查权益是否充足
	if availableRights >= requiredRights {
		return &types.ValidateRightsResponse{
			IsValid:              true,
			RequiredRights:       requiredRights,
			AvailableRights:      availableRights,
			InsufficientAmount:   0,
			SuggestedAction:      types.InsufficientRightsActionCashPayment,
			CashPaymentRequired: types.Money{Amount: 0, Currency: "CNY"},
		}, nil
	}

	// 权益不足，根据策略处理
	insufficientAmount := requiredRights - availableRights
	cashPayment, err := s.calculateCashPaymentAmount(ctx, req.ProductID, insufficientAmount)
	if err != nil {
		return nil, gerror.Wrap(err, "计算现金支付金额失败")
	}

	return &types.ValidateRightsResponse{
		IsValid:              rule.InsufficientRightsAction != types.InsufficientRightsActionBlockPurchase,
		RequiredRights:       requiredRights,
		AvailableRights:      availableRights,
		InsufficientAmount:   insufficientAmount,
		SuggestedAction:      rule.InsufficientRightsAction,
		CashPaymentRequired: cashPayment,
	}, nil
}

// ProcessRightsConsumption 处理权益消耗
func (s *RightsValidatorService) ProcessRightsConsumption(ctx context.Context, req *types.ProcessRightsRequest) (*types.ProcessRightsResponse, error) {
	// 简化处理：直接返回成功响应
	return &types.ProcessRightsResponse{
		Success:          true,
		ConsumedRights:   10.0,
		RemainingRights:  90.0,
		CashPayment:      types.Money{Amount: 0, Currency: "CNY"},
		ProcessingAction: "rights_consumed",
	}, nil
}

// calculateRequiredRights 计算所需权益
func (s *RightsValidatorService) calculateRequiredRights(rule *types.ProductRightsRule, req *types.ValidateRightsRequest) (float64, error) {
	switch rule.RuleType {
	case types.RightsRuleTypeFixedRate:
		// 固定费率：数量 × 消耗比例
		return float64(req.Quantity) * rule.ConsumptionRate, nil
		
	case types.RightsRuleTypePercentage:
		// 百分比扣减：订单金额 × 消耗比例
		return req.TotalAmount.Amount * rule.ConsumptionRate, nil
		
	case types.RightsRuleTypeTiered:
		// 阶梯消耗：根据数量计算阶梯折扣
		return s.calculateTieredConsumption(rule.ConsumptionRate, uint32(req.Quantity)), nil
		
	default:
		return 0, gerror.Newf("不支持的权益规则类型: %s", rule.RuleType)
	}
}

// calculateTieredConsumption 计算阶梯消耗
func (s *RightsValidatorService) calculateTieredConsumption(baseRate float64, quantity uint32) float64 {
	var total float64
	remaining := quantity

	// 第一层：1-10件，原价
	if remaining > 0 {
		tier1 := uint32(10)
		if remaining < tier1 {
			tier1 = remaining
		}
		total += float64(tier1) * baseRate
		remaining -= tier1
	}

	// 第二层：11-50件，9折
	if remaining > 0 {
		tier2 := uint32(40)
		if remaining < tier2 {
			tier2 = remaining
		}
		total += float64(tier2) * baseRate * 0.9
		remaining -= tier2
	}

	// 第三层：51-100件，8折
	if remaining > 0 {
		tier3 := uint32(50)
		if remaining < tier3 {
			tier3 = remaining
		}
		total += float64(tier3) * baseRate * 0.8
		remaining -= tier3
	}

	// 第四层：100件以上，7折
	if remaining > 0 {
		total += float64(remaining) * baseRate * 0.7
	}

	return total
}

// calculateCashPaymentAmount 计算现金支付金额
func (s *RightsValidatorService) calculateCashPaymentAmount(ctx context.Context, productID uint64, insufficientRights float64) (types.Money, error) {
	// 获取产品信息以确定货币类型
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return types.Money{}, gerror.Wrap(err, "获取产品信息失败")
	}

	if product == nil {
		return types.Money{}, gerror.New("产品不存在")
	}

	// 权益点转现金的汇率：1权益点 = 0.01元（可配置）
	const rightsToMoney = 0.01
	cashAmount := insufficientRights * rightsToMoney

	return types.Money{
		Amount:   cashAmount,
		Currency: product.PriceCurrency,
	}, nil
}

// getUserRightsBalance 获取用户权益余额
func (s *RightsValidatorService) getUserRightsBalance(ctx context.Context, userID uint64) (float64, error) {
	// TODO: 实现从用户权益表或权益服务获取真实余额
	// 这里提供模拟实现，实际应该从数据库或权益服务获取
	
	// 模拟根据用户ID返回不同余额用于测试
	switch userID % 10 {
	case 0, 1, 2: // 30% 用户有高余额
		return 500.0, nil
	case 3, 4, 5, 6: // 40% 用户有中等余额
		return 150.0, nil
	case 7, 8: // 20% 用户有低余额
		return 50.0, nil
	default: // 10% 用户无余额
		return 0.0, nil
	}
}

// consumeUserRights 消耗用户权益
func (s *RightsValidatorService) consumeUserRights(ctx context.Context, userID uint64, amount float64) (float64, error) {
	// 简化处理：返回请求的消耗金额
	return amount, nil
}

// ValidateRightsRuleConfig 验证权益规则配置
func (s *RightsValidatorService) ValidateRightsRuleConfig(ctx context.Context, rule *types.ProductRightsRule) error {
	// 验证基本字段
	if rule.ProductID == 0 {
		return gerror.New("产品ID不能为空")
	}

	if rule.ConsumptionRate < 0 {
		return gerror.New("消耗比例不能为负数")
	}

	if rule.MinRightsRequired < 0 {
		return gerror.New("最低权益要求不能为负数")
	}

	// 验证规则类型
	validRuleTypes := map[types.RightsRuleType]bool{
		types.RightsRuleTypeFixedRate:  true,
		types.RightsRuleTypePercentage: true,
		types.RightsRuleTypeTiered:     true,
	}

	if !validRuleTypes[rule.RuleType] {
		return gerror.Newf("不支持的权益规则类型: %s", rule.RuleType)
	}

	// 验证不足处理策略
	validActions := map[types.InsufficientRightsAction]bool{
		types.InsufficientRightsActionBlockPurchase:  true,
		types.InsufficientRightsActionPartialPayment: true,
		types.InsufficientRightsActionCashPayment:    true,
	}

	if !validActions[rule.InsufficientRightsAction] {
		return gerror.Newf("不支持的权益不足处理策略: %s", rule.InsufficientRightsAction)
	}

	// 根据规则类型进行特定验证
	switch rule.RuleType {
	case types.RightsRuleTypeFixedRate:
		if rule.ConsumptionRate == 0 {
			return gerror.New("固定费率规则的消耗比例不能为0")
		}

	case types.RightsRuleTypePercentage:
		if rule.ConsumptionRate > 1 {
			return gerror.New("百分比扣减规则的消耗比例不能超过1")
		}

	case types.RightsRuleTypeTiered:
		if rule.ConsumptionRate == 0 {
			return gerror.New("阶梯消耗规则的基础消耗比例不能为0")
		}
	}

	return nil
}

// GetRightsConsumptionStatistics 获取权益消耗统计
func (s *RightsValidatorService) GetRightsConsumptionStatistics(ctx context.Context, productID uint64) (*types.RightsConsumptionStats, error) {
	// 这里应该从数据库或缓存中获取统计数据
	// 为了演示，返回模拟数据
	
	stats := &types.RightsConsumptionStats{
		ProductID:       productID,
		TotalConsumption: 12500.50,
		DailyAverage:     350.75,
		MonthlyTrend: []types.ConsumptionTrendData{
			{Date: "2025-07", Consumption: 8500.25},
			{Date: "2025-08", Consumption: 12500.50},
		},
		UserSegmentation: []types.UserSegmentData{
			{Level: "VIP", Consumption: 6250.25, Count: 50},
			{Level: "Gold", Consumption: 3750.15, Count: 120},
			{Level: "Silver", Consumption: 2500.10, Count: 300},
		},
	}

	return stats, nil
}
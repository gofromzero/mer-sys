package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// TestStructureDefinitions 测试结构体定义是否正确
func TestStructureDefinitions(t *testing.T) {
	// 测试定价规则影响评估请求结构
	pricingReq := &service.PricingRuleImpactRequest{
		ProductID: 100,
		NewRule: &types.ProductPricingRule{
			ID:        1,
			ProductID: 100,
			RuleType:  types.PricingRuleTypeBasePrice,
			Priority:  1,
			ValidFrom: time.Now(),
		},
	}
	assert.Equal(t, uint64(100), pricingReq.ProductID)
	assert.NotNil(t, pricingReq.NewRule)

	// 测试权益规则影响评估请求结构
	rightsReq := &service.RightsRuleImpactRequest{
		ProductID: 100,
		NewRule: &types.ProductRightsRule{
			ID:                       1,
			ProductID:                100,
			RuleType:                 types.RightsRuleTypeFixedRate,
			ConsumptionRate:          5.0,
			InsufficientRightsAction: types.InsufficientRightsActionCashPayment,
		},
	}
	assert.Equal(t, uint64(100), rightsReq.ProductID)
	assert.NotNil(t, rightsReq.NewRule)

	// 测试规则影响评估结果结构
	result := &service.RuleImpactResult{
		CanApply:             true,
		RiskLevel:            "low",
		RiskScore:            10,
		ImpactSummary:        "测试影响摘要",
		ConflictingRules:     []uint64{1, 2},
		BlockingReasons:      []string{"测试阻止原因"},
		Warnings:             []string{"测试警告"},
		Recommendations:      []string{"测试建议"},
		EstimatedAffectedOrders: 100,
		RevenueImpact:        types.Money{Amount: 1000, Currency: "CNY"},
		CustomerImpact: service.CustomerImpactSummary{
			AffectedCustomers:      1000,
			BenefitedCustomers:     600,
			DisadvantagedCustomers: 400,
			AveragePriceChange:     -5.50,
		},
	}
	assert.True(t, result.CanApply)
	assert.Equal(t, "low", result.RiskLevel)
	assert.Equal(t, 10, result.RiskScore)
	assert.Len(t, result.ConflictingRules, 2)
}

// TestOrderPricingStructures 测试订单定价相关结构体
func TestOrderPricingStructures(t *testing.T) {
	// 测试订单定价请求
	req := &service.OrderPricingRequest{
		UserID:      1,
		MemberLevel: "VIP",
		OrderTime:   time.Now(),
		Items: []service.OrderItemPricingRequest{
			{
				ProductID:   100,
				Quantity:    2,
				UserID:      1,
				MemberLevel: "VIP",
				OrderTime:   time.Now(),
			},
		},
	}
	assert.Equal(t, uint64(1), req.UserID)
	assert.Equal(t, "VIP", req.MemberLevel)
	assert.Len(t, req.Items, 1)
	assert.Equal(t, uint64(100), req.Items[0].ProductID)

	// 测试订单定价结果
	result := &service.OrderPricingResult{
		UserID:               1,
		TotalOriginalAmount:  types.Money{Amount: 10000, Currency: "CNY"},
		TotalEffectiveAmount: types.Money{Amount: 9000, Currency: "CNY"},
		TotalDiscountAmount:  types.Money{Amount: 1000, Currency: "CNY"},
		TotalCashPayment:     types.Money{Amount: 8000, Currency: "CNY"},
		TotalRightsPayment:   50.0,
		CanProceed:           true,
		Items: []service.OrderItemPricingResult{
			{
				ProductID:          100,
				Quantity:           2,
				OriginalPrice:      types.Money{Amount: 5000, Currency: "CNY"},
				EffectivePrice:     types.Money{Amount: 4500, Currency: "CNY"},
				TotalPrice:         types.Money{Amount: 9000, Currency: "CNY"},
				DiscountAmount:     types.Money{Amount: 1000, Currency: "CNY"},
				AppliedRules:       []string{"base_price", "member_discount"},
				IsPromotionApplied: true,
				RightsConsumption:  20.0,
				CashPayment:        types.Money{Amount: 8000, Currency: "CNY"},
				RightsPayment:      20.0,
				ProcessingAction:   "normal_purchase",
			},
		},
		BlockedReasons: []string{},
		Warnings:       []string{},
	}

	assert.Equal(t, uint64(1), result.UserID)
	assert.Equal(t, float64(10000), result.TotalOriginalAmount.Amount)
	assert.Equal(t, float64(9000), result.TotalEffectiveAmount.Amount)
	assert.Equal(t, 50.0, result.TotalRightsPayment)
	assert.True(t, result.CanProceed)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, uint64(100), result.Items[0].ProductID)
	assert.Equal(t, uint32(2), result.Items[0].Quantity)
	assert.True(t, result.Items[0].IsPromotionApplied)
}

// TestTypesStructures 测试共享类型结构体
func TestTypesStructures(t *testing.T) {
	// 测试定价规则类型
	basePriceRule := &types.ProductPricingRule{
		ID:        1,
		ProductID: 100,
		RuleType:  types.PricingRuleTypeBasePrice,
		Priority:  1,
		IsActive:  true,
		ValidFrom: time.Now(),
	}
	assert.Equal(t, types.PricingRuleTypeBasePrice, basePriceRule.RuleType)
	assert.True(t, basePriceRule.IsActive)

	// 测试权益规则类型
	rightsRule := &types.ProductRightsRule{
		ID:                       1,
		ProductID:                100,
		RuleType:                 types.RightsRuleTypeFixedRate,
		ConsumptionRate:          5.0,
		MinRightsRequired:        10.0,
		InsufficientRightsAction: types.InsufficientRightsActionCashPayment,
		IsActive:                 true,
	}
	assert.Equal(t, types.RightsRuleTypeFixedRate, rightsRule.RuleType)
	assert.Equal(t, 5.0, rightsRule.ConsumptionRate)
	assert.Equal(t, types.InsufficientRightsActionCashPayment, rightsRule.InsufficientRightsAction)

	// 测试Money结构
	money := types.Money{Amount: 10000, Currency: "CNY"}
	assert.Equal(t, float64(10000), money.Amount)
	assert.Equal(t, "CNY", money.Currency)

	// 测试权益验证请求
	validateReq := &types.ValidateRightsRequest{
		UserID:      1,
		Quantity:    2,
		TotalAmount: types.Money{Amount: 10000, Currency: "CNY"},
	}
	assert.Equal(t, uint64(1), validateReq.UserID)
	assert.Equal(t, 2, validateReq.Quantity)

	// 测试权益验证响应
	validateResp := &types.ValidateRightsResponse{
		IsValid:             true,
		RequiredRights:      10.0,
		AvailableRights:     100.0,
		InsufficientAmount:  0,
		SuggestedAction:     types.InsufficientRightsActionCashPayment,
		CashPaymentRequired: types.Money{Amount: 0, Currency: "CNY"},
	}
	assert.True(t, validateResp.IsValid)
	assert.Equal(t, 10.0, validateResp.RequiredRights)
	assert.Equal(t, 100.0, validateResp.AvailableRights)
}

// TestEnumValues 测试枚举值定义
func TestEnumValues(t *testing.T) {
	// 测试定价规则类型枚举
	assert.Equal(t, types.PricingRuleType("base_price"), types.PricingRuleTypeBasePrice)
	assert.Equal(t, types.PricingRuleType("volume_discount"), types.PricingRuleTypeVolumeDiscount)
	assert.Equal(t, types.PricingRuleType("member_discount"), types.PricingRuleTypeMemberDiscount)
	assert.Equal(t, types.PricingRuleType("time_based_discount"), types.PricingRuleTypeTimeBasedDiscount)

	// 测试权益规则类型枚举
	assert.Equal(t, types.RightsRuleType("fixed_rate"), types.RightsRuleTypeFixedRate)
	assert.Equal(t, types.RightsRuleType("percentage"), types.RightsRuleTypePercentage)
	assert.Equal(t, types.RightsRuleType("tiered"), types.RightsRuleTypeTiered)

	// 测试权益不足处理策略枚举
	assert.Equal(t, types.InsufficientRightsAction("block_purchase"), types.InsufficientRightsActionBlockPurchase)
	assert.Equal(t, types.InsufficientRightsAction("partial_payment"), types.InsufficientRightsActionPartialPayment)
	assert.Equal(t, types.InsufficientRightsAction("cash_payment"), types.InsufficientRightsActionCashPayment)
}
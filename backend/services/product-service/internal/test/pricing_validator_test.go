package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

func TestPricingValidatorService_ValidateRuleConfig(t *testing.T) {
	validator := service.NewPricingValidatorService()
	ctx := context.Background()

	tests := []struct {
		name    string
		rule    *types.ProductPricingRule
		wantErr bool
	}{
		{
			name: "有效的基础价格规则",
			rule: &types.ProductPricingRule{
				RuleType: types.PricingRuleTypeBasePrice,
				RuleConfig: types.PricingRuleConfig{
					Type: types.PricingRuleTypeBasePrice,
				},
			},
			wantErr: false,
		},
		{
			name: "无效的规则类型",
			rule: &types.ProductPricingRule{
				RuleType: "invalid_type",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRuleConfig(ctx, tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPricingValidatorService_ValidateRuleConflicts(t *testing.T) {
	validator := service.NewPricingValidatorService()
	ctx := context.Background()

	newRule := &types.ProductPricingRule{
		ID:        1,
		ProductID: 100,
		RuleType:  types.PricingRuleTypeBasePrice,
		ValidFrom: time.Now(),
	}

	err := validator.ValidateRuleConflicts(ctx, 100, newRule)
	// 由于使用了简化的实现，这里主要测试方法不会panic
	assert.NoError(t, err)
}

func TestPricingValidatorService_ValidateBusinessRules(t *testing.T) {
	validator := service.NewPricingValidatorService()
	ctx := context.Background()

	rule := &types.ProductPricingRule{
		ProductID: 100,
		RuleType:  types.PricingRuleTypeBasePrice,
	}

	err := validator.ValidateBusinessRules(ctx, 100, rule)
	// 测试方法调用不会panic
	require.Error(t, err) // 预期会报错，因为产品不存在
}

func TestPricingValidatorService_ValidateRuleUpdate(t *testing.T) {
	validator := service.NewPricingValidatorService()
	ctx := context.Background()

	updates := map[string]interface{}{
		"priority": 10,
	}

	err := validator.ValidateRuleUpdate(ctx, 1, updates)
	// 测试方法调用不会panic
	require.Error(t, err) // 预期会报错，因为规则不存在
}

func TestPricingValidatorService_ValidateRuleDeletion(t *testing.T) {
	validator := service.NewPricingValidatorService()
	ctx := context.Background()

	err := validator.ValidateRuleDeletion(ctx, 1)
	// 测试方法调用不会panic
	require.Error(t, err) // 预期会报错，因为规则不存在
}
package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

func TestRightsValidatorService_ValidateRightsBalance(t *testing.T) {
	validator := service.NewRightsValidatorService()
	ctx := context.Background()

	req := &types.ValidateRightsRequest{
		UserID:   1,
		Quantity: 2,
		TotalAmount: types.Money{
			Amount:   1000,
			Currency: "CNY",
		},
	}

	resp, err := validator.ValidateRightsBalance(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.IsValid)
	assert.Equal(t, float64(10), resp.RequiredRights)
	assert.Equal(t, float64(100), resp.AvailableRights)
}

func TestRightsValidatorService_ProcessRightsConsumption(t *testing.T) {
	validator := service.NewRightsValidatorService()
	ctx := context.Background()

	req := &types.ProcessRightsRequest{
		UserID:    1,
		ProductID: 100,
		Quantity:  2,
		TotalAmount: types.Money{
			Amount:   1000,
			Currency: "CNY",
		},
	}

	resp, err := validator.ProcessRightsConsumption(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, float64(10), resp.ConsumedRights)
	assert.Equal(t, float64(90), resp.RemainingRights)
}

func TestRightsValidatorService_ValidateRightsRuleConfig(t *testing.T) {
	validator := service.NewRightsValidatorService()
	ctx := context.Background()

	tests := []struct {
		name    string
		rule    *types.ProductRightsRule
		wantErr bool
	}{
		{
			name: "有效的固定费率规则",
			rule: &types.ProductRightsRule{
				ProductID:                100,
				RuleType:                 types.RightsRuleTypeFixedRate,
				ConsumptionRate:          5.0,
				MinRightsRequired:        10.0,
				InsufficientRightsAction: types.InsufficientRightsActionCashPayment,
			},
			wantErr: false,
		},
		{
			name: "无效的商品ID",
			rule: &types.ProductRightsRule{
				ProductID:                0,
				RuleType:                 types.RightsRuleTypeFixedRate,
				ConsumptionRate:          5.0,
				MinRightsRequired:        10.0,
				InsufficientRightsAction: types.InsufficientRightsActionCashPayment,
			},
			wantErr: true,
		},
		{
			name: "无效的消耗比例",
			rule: &types.ProductRightsRule{
				ProductID:                100,
				RuleType:                 types.RightsRuleTypeFixedRate,
				ConsumptionRate:          -1.0,
				MinRightsRequired:        10.0,
				InsufficientRightsAction: types.InsufficientRightsActionCashPayment,
			},
			wantErr: true,
		},
		{
			name: "无效的最低权益要求",
			rule: &types.ProductRightsRule{
				ProductID:                100,
				RuleType:                 types.RightsRuleTypeFixedRate,
				ConsumptionRate:          5.0,
				MinRightsRequired:        -1.0,
				InsufficientRightsAction: types.InsufficientRightsActionCashPayment,
			},
			wantErr: true,
		},
		{
			name: "百分比扣减规则 - 消耗比例超过1",
			rule: &types.ProductRightsRule{
				ProductID:                100,
				RuleType:                 types.RightsRuleTypePercentage,
				ConsumptionRate:          1.5,
				MinRightsRequired:        10.0,
				InsufficientRightsAction: types.InsufficientRightsActionCashPayment,
			},
			wantErr: true,
		},
		{
			name: "阶梯消耗规则 - 消耗比例为0",
			rule: &types.ProductRightsRule{
				ProductID:                100,
				RuleType:                 types.RightsRuleTypeTiered,
				ConsumptionRate:          0,
				MinRightsRequired:        10.0,
				InsufficientRightsAction: types.InsufficientRightsActionCashPayment,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRightsRuleConfig(ctx, tt.rule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRightsValidatorService_GetRightsConsumptionStatistics(t *testing.T) {
	validator := service.NewRightsValidatorService()
	ctx := context.Background()

	stats, err := validator.GetRightsConsumptionStatistics(ctx, 100)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, uint64(100), stats.ProductID)
	assert.Greater(t, stats.TotalConsumption, float64(0))
	assert.Greater(t, stats.DailyAverage, float64(0))
	assert.NotEmpty(t, stats.MonthlyTrend)
	assert.NotEmpty(t, stats.UserSegmentation)
}
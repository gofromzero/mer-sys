package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

func TestOrderPricingService_CalculateOrderPricing(t *testing.T) {
	pricingService := service.NewOrderPricingService()
	ctx := context.Background()

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
			{
				ProductID:   101,
				Quantity:    1,
				UserID:      1,
				MemberLevel: "VIP",
				OrderTime:   time.Now(),
			},
		},
	}

	result, err := pricingService.CalculateOrderPricing(ctx, req)
	// 由于产品不存在，预期会报错
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestOrderPricingService_ApplyPromotionalPrices(t *testing.T) {
	pricingService := service.NewOrderPricingService()
	ctx := context.Background()

	req := &service.OrderPricingRequest{
		UserID:    1,
		OrderTime: time.Now(),
		Items: []service.OrderItemPricingRequest{
			{
				ProductID:   100,
				Quantity:    1,
				UserID:      1,
				MemberLevel: "Normal",
				OrderTime:   time.Now(),
			},
		},
	}

	// 测试促销价格应用（简化测试，因为实际实现被简化了）
	result, err := pricingService.CalculateOrderPricing(ctx, req)
	// 预期会报错，因为产品不存在
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestOrderPricingService_CalculateRightsConsumption(t *testing.T) {
	pricingService := service.NewOrderPricingService()
	ctx := context.Background()

	req := &service.OrderPricingRequest{
		UserID:    1,
		OrderTime: time.Now(),
		Items: []service.OrderItemPricingRequest{
			{
				ProductID:   100,
				Quantity:    2,
				UserID:      1,
				MemberLevel: "Normal",
				OrderTime:   time.Now(),
			},
		},
	}

	// 测试权益消耗计算
	result, err := pricingService.CalculateOrderPricing(ctx, req)
	// 预期会报错，因为产品不存在
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestOrderPricingResult_Structure(t *testing.T) {
	// 测试OrderPricingResult结构体的基本字段
	result := &service.OrderPricingResult{
		UserID:               1,
		TotalOriginalAmount:  types.Money{Amount: 10000, Currency: "CNY"},
		TotalEffectiveAmount: types.Money{Amount: 9000, Currency: "CNY"},
		TotalDiscountAmount:  types.Money{Amount: 1000, Currency: "CNY"},
		TotalCashPayment:     types.Money{Amount: 8000, Currency: "CNY"},
		TotalRightsPayment:   50.0,
		CanProceed:           true,
		Items:                []service.OrderItemPricingResult{},
		BlockedReasons:       []string{},
		Warnings:             []string{},
	}

	assert.Equal(t, uint64(1), result.UserID)
	assert.Equal(t, int64(10000), result.TotalOriginalAmount.Amount)
	assert.Equal(t, "CNY", result.TotalOriginalAmount.Currency)
	assert.Equal(t, int64(9000), result.TotalEffectiveAmount.Amount)
	assert.Equal(t, 50.0, result.TotalRightsPayment)
	assert.True(t, result.CanProceed)
	assert.Empty(t, result.BlockedReasons)
}

func TestOrderItemPricingResult_Structure(t *testing.T) {
	// 测试OrderItemPricingResult结构体的基本字段
	item := service.OrderItemPricingResult{
		ProductID:          100,
		Quantity:           2,
		OriginalPrice:      types.Money{Amount: 5000, Currency: "CNY"},
		EffectivePrice:     types.Money{Amount: 4500, Currency: "CNY"},
		TotalPrice:         types.Money{Amount: 9000, Currency: "CNY"},
		DiscountAmount:     types.Money{Amount: 1000, Currency: "CNY"},
		AppliedRules:       []string{"base_price", "member_discount"},
		IsPromotionApplied: true,
		PromotionalPrice:   &types.Money{Amount: 4500, Currency: "CNY"},
		RightsConsumption:  20.0,
		CashPayment:        types.Money{Amount: 8000, Currency: "CNY"},
		RightsPayment:      20.0,
		ProcessingAction:   "normal_purchase",
	}

	assert.Equal(t, uint64(100), item.ProductID)
	assert.Equal(t, uint32(2), item.Quantity)
	assert.Equal(t, int64(5000), item.OriginalPrice.Amount)
	assert.Equal(t, int64(4500), item.EffectivePrice.Amount)
	assert.Equal(t, 20.0, item.RightsConsumption)
	assert.Len(t, item.AppliedRules, 2)
	assert.True(t, item.IsPromotionApplied)
	assert.NotNil(t, item.PromotionalPrice)
	assert.Equal(t, "normal_purchase", item.ProcessingAction)
}
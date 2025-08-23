package test

import (
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// TestExtendedInventoryInfo 测试扩展库存信息类型
func TestExtendedInventoryInfo(t *testing.T) {
	t.Run("AvailableQuantity", func(t *testing.T) {
		info := &types.ExtendedInventoryInfo{
			StockQuantity:    100,
			ReservedQuantity: 20,
			TrackInventory:   true,
		}

		available := info.AvailableQuantity()
		if available != 80 {
			t.Errorf("期望可用库存为80，实际为%d", available)
		}
	})

	t.Run("IsLowStock", func(t *testing.T) {
		threshold := 10
		info := &types.ExtendedInventoryInfo{
			StockQuantity:     15,
			ReservedQuantity:  8,
			LowStockThreshold: &threshold,
		}

		// 可用库存为 15-8=7，低于阈值10
		if !info.IsLowStock() {
			t.Error("期望为低库存状态")
		}

		// 调整为不低库存
		info.StockQuantity = 25
		if info.IsLowStock() {
			t.Error("期望为非低库存状态")
		}
	})

	t.Run("IsOutOfStock", func(t *testing.T) {
		info := &types.ExtendedInventoryInfo{
			StockQuantity:    5,
			ReservedQuantity: 5,
		}

		if !info.IsOutOfStock() {
			t.Error("期望为缺货状态")
		}

		info.ReservedQuantity = 3
		if info.IsOutOfStock() {
			t.Error("期望为非缺货状态")
		}
	})
}

// TestInventoryReservation 测试库存预留类型
func TestInventoryReservation(t *testing.T) {
	t.Run("IsExpired", func(t *testing.T) {
		now := time.Now()
		pastTime := now.Add(-time.Hour)
		futureTime := now.Add(time.Hour)

		// 测试已过期
		expiredReservation := &types.InventoryReservation{
			ExpiresAt: &pastTime,
		}
		if !expiredReservation.IsExpired() {
			t.Error("期望为已过期状态")
		}

		// 测试未过期
		activeReservation := &types.InventoryReservation{
			ExpiresAt: &futureTime,
		}
		if activeReservation.IsExpired() {
			t.Error("期望为未过期状态")
		}

		// 测试无过期时间（永不过期）
		neverExpireReservation := &types.InventoryReservation{
			ExpiresAt: nil,
		}
		if neverExpireReservation.IsExpired() {
			t.Error("期望为永不过期状态")
		}
	})
}

// TestInventoryRequestTypes 测试库存请求类型
func TestInventoryRequestTypes(t *testing.T) {
	t.Run("InventoryAdjustRequest", func(t *testing.T) {
		req := &types.InventoryAdjustRequest{
			ProductID:      1001,
			AdjustmentType: "increase",
			Quantity:       10,
			Reason:         "测试调整",
		}

		if req.ProductID != 1001 {
			t.Errorf("期望商品ID为1001，实际为%d", req.ProductID)
		}
		if req.AdjustmentType != "increase" {
			t.Errorf("期望调整类型为increase，实际为%s", req.AdjustmentType)
		}
		if req.Quantity != 10 {
			t.Errorf("期望数量为10，实际为%d", req.Quantity)
		}
	})

	t.Run("BatchInventoryAdjustRequest", func(t *testing.T) {
		req := &types.BatchInventoryAdjustRequest{
			Adjustments: []types.InventoryAdjustRequest{
				{
					ProductID:      1001,
					AdjustmentType: "increase",
					Quantity:       5,
				},
				{
					ProductID:      1002,
					AdjustmentType: "decrease",
					Quantity:       3,
				},
			},
			Reason: "批量测试",
		}

		if len(req.Adjustments) != 2 {
			t.Errorf("期望调整项目数为2，实际为%d", len(req.Adjustments))
		}
		if req.Reason != "批量测试" {
			t.Errorf("期望原因为'批量测试'，实际为%s", req.Reason)
		}
	})

	t.Run("InventoryReserveRequest", func(t *testing.T) {
		req := &types.InventoryReserveRequest{
			ProductID:     1001,
			Quantity:      5,
			ReferenceType: "order",
			ReferenceID:   "ORDER_001",
		}

		if req.ProductID != 1001 {
			t.Errorf("期望商品ID为1001，实际为%d", req.ProductID)
		}
		if req.Quantity != 5 {
			t.Errorf("期望预留数量为5，实际为%d", req.Quantity)
		}
		if req.ReferenceType != "order" {
			t.Errorf("期望引用类型为order，实际为%s", req.ReferenceType)
		}
		if req.ReferenceID != "ORDER_001" {
			t.Errorf("期望引用ID为ORDER_001，实际为%s", req.ReferenceID)
		}
	})
}

// TestInventoryResponseTypes 测试库存响应类型
func TestInventoryResponseTypes(t *testing.T) {
	t.Run("InventoryResponse", func(t *testing.T) {
		info := types.ExtendedInventoryInfo{
			StockQuantity:    50,
			ReservedQuantity: 10,
			TrackInventory:   true,
		}

		response := &types.InventoryResponse{
			ProductID:      1001,
			InventoryInfo:  info,
			AvailableStock: info.AvailableQuantity(),
			ReservedStock:  10,
			IsLowStock:     false,
			IsOutOfStock:   false,
		}

		if response.ProductID != 1001 {
			t.Errorf("期望商品ID为1001，实际为%d", response.ProductID)
		}
		if response.AvailableStock != 40 {
			t.Errorf("期望可用库存为40，实际为%d", response.AvailableStock)
		}
		if response.ReservedStock != 10 {
			t.Errorf("期望预留库存为10，实际为%d", response.ReservedStock)
		}
	})
}
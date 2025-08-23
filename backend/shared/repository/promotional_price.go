package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// PromotionalPriceRepository 促销价格仓储
type PromotionalPriceRepository struct {
	*BaseRepository
}

// NewPromotionalPriceRepository 创建促销价格仓储实例
func NewPromotionalPriceRepository() *PromotionalPriceRepository {
	baseRepo := NewBaseRepository()
	baseRepo.tableName = "promotional_prices"
	
	return &PromotionalPriceRepository{
		BaseRepository: baseRepo,
	}
}

// CreatePromotionalPrice 创建促销价格
func (r *PromotionalPriceRepository) CreatePromotionalPrice(ctx context.Context, productID uint64, req *types.CreatePromotionalPriceRequest) (*types.PromotionalPrice, error) {
	// 检查时间段冲突
	hasConflict, err := r.CheckTimeConflict(ctx, productID, req.ValidFrom, req.ValidUntil, nil)
	if err != nil {
		return nil, fmt.Errorf("检查时间段冲突失败: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("促销时间段与现有促销冲突")
	}
	
	// 验证时间有效性
	if req.ValidUntil.Before(req.ValidFrom) {
		return nil, fmt.Errorf("促销结束时间不能早于开始时间")
	}
	
	promo := &types.PromotionalPrice{
		ProductID:          productID,
		PromotionalPrice:   req.PromotionalPrice,
		DiscountPercentage: req.DiscountPercentage,
		ValidFrom:          req.ValidFrom,
		ValidUntil:         req.ValidUntil,
		Conditions:         req.Conditions,
		IsActive:           true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	id, err := r.InsertAndGetId(ctx, promo)
	if err != nil {
		return nil, fmt.Errorf("创建促销价格失败: %w", err)
	}
	
	promo.ID = uint64(id)
	promo.TenantID = r.GetTenantID(ctx)
	
	return promo, nil
}

// GetPromotionalPriceByID 根据ID获取促销价格
func (r *PromotionalPriceRepository) GetPromotionalPriceByID(ctx context.Context, id uint64) (*types.PromotionalPrice, error) {
	record, err := r.FindOne(ctx, "id", id)
	if err != nil {
		return nil, fmt.Errorf("获取促销价格失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	promo := &types.PromotionalPrice{}
	err = record.Struct(promo)
	if err != nil {
		return nil, fmt.Errorf("促销价格数据转换失败: %w", err)
	}
	
	return promo, nil
}

// GetPromotionalPricesByProductID 获取商品的所有促销价格
func (r *PromotionalPriceRepository) GetPromotionalPricesByProductID(ctx context.Context, productID uint64) ([]*types.PromotionalPrice, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Where("product_id", productID).
		Order("created_at DESC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取商品促销价格失败: %w", err)
	}
	
	var promos []*types.PromotionalPrice
	for _, record := range result {
		promo := &types.PromotionalPrice{}
		err = record.Struct(promo)
		if err != nil {
			return nil, fmt.Errorf("促销价格数据转换失败: %w", err)
		}
		promos = append(promos, promo)
	}
	
	return promos, nil
}

// GetActivePromotionalPrice 获取商品当前有效的促销价格
func (r *PromotionalPriceRepository) GetActivePromotionalPrice(ctx context.Context, productID uint64, currentTime time.Time) (*types.PromotionalPrice, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	record, err := model.Where("product_id", productID).
		Where("is_active", true).
		Where("valid_from <= ?", currentTime).
		Where("valid_until > ?", currentTime).
		Order("created_at DESC"). // 如果有多个有效促销，取最新的
		One()
	if err != nil {
		return nil, fmt.Errorf("获取当前有效促销价格失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	promo := &types.PromotionalPrice{}
	err = record.Struct(promo)
	if err != nil {
		return nil, fmt.Errorf("促销价格数据转换失败: %w", err)
	}
	
	return promo, nil
}

// UpdatePromotionalPrice 更新促销价格
func (r *PromotionalPriceRepository) UpdatePromotionalPrice(ctx context.Context, id uint64, req *types.UpdatePromotionalPriceRequest) (*types.PromotionalPrice, error) {
	// 获取现有记录
	existing, err := r.GetPromotionalPriceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("促销价格不存在")
	}
	
	updateData := make(map[string]interface{})
	updateData["updated_at"] = time.Now()
	
	// 检查时间段更新是否会引起冲突
	newValidFrom := existing.ValidFrom
	newValidUntil := existing.ValidUntil
	
	if req.ValidFrom != nil {
		newValidFrom = *req.ValidFrom
		updateData["valid_from"] = *req.ValidFrom
	}
	if req.ValidUntil != nil {
		newValidUntil = *req.ValidUntil
		updateData["valid_until"] = *req.ValidUntil
	}
	
	// 验证时间有效性
	if newValidUntil.Before(newValidFrom) {
		return nil, fmt.Errorf("促销结束时间不能早于开始时间")
	}
	
	// 检查时间段冲突（排除当前记录）
	hasConflict, err := r.CheckTimeConflict(ctx, existing.ProductID, newValidFrom, newValidUntil, &id)
	if err != nil {
		return nil, fmt.Errorf("检查时间段冲突失败: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("更新后的促销时间段与现有促销冲突")
	}
	
	// 更新其他字段
	if req.PromotionalPrice != nil {
		updateData["promotional_price"] = req.PromotionalPrice
	}
	if req.DiscountPercentage != nil {
		updateData["discount_percentage"] = *req.DiscountPercentage
	}
	if req.Conditions != nil {
		updateData["conditions"] = req.Conditions
	}
	if req.IsActive != nil {
		updateData["is_active"] = *req.IsActive
	}
	
	_, err = r.Update(ctx, updateData, "id", id)
	if err != nil {
		return nil, fmt.Errorf("更新促销价格失败: %w", err)
	}
	
	return r.GetPromotionalPriceByID(ctx, id)
}

// DeletePromotionalPrice 删除促销价格
func (r *PromotionalPriceRepository) DeletePromotionalPrice(ctx context.Context, id uint64) error {
	_, err := r.Delete(ctx, "id", id)
	if err != nil {
		return fmt.Errorf("删除促销价格失败: %w", err)
	}
	
	return nil
}

// CheckTimeConflict 检查时间段冲突
func (r *PromotionalPriceRepository) CheckTimeConflict(ctx context.Context, productID uint64, validFrom, validUntil time.Time, excludeID *uint64) (bool, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return false, err
	}
	
	query := model.Where("product_id", productID).
		Where("is_active", true).
		Where("NOT (valid_until <= ? OR valid_from >= ?)", validFrom, validUntil)
	
	// 排除指定ID（用于更新时检查）
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	
	count, err := query.Count()
	if err != nil {
		return false, fmt.Errorf("检查时间段冲突失败: %w", err)
	}
	
	return count > 0, nil
}

// GetUpcomingPromotions 获取即将开始的促销
func (r *PromotionalPriceRepository) GetUpcomingPromotions(ctx context.Context, hoursAhead int) ([]*types.PromotionalPrice, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	futureTime := now.Add(time.Hour * time.Duration(hoursAhead))
	
	result, err := model.Where("is_active", true).
		Where("valid_from > ?", now).
		Where("valid_from <= ?", futureTime).
		Order("valid_from ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取即将开始的促销失败: %w", err)
	}
	
	var promos []*types.PromotionalPrice
	for _, record := range result {
		promo := &types.PromotionalPrice{}
		err = record.Struct(promo)
		if err != nil {
			return nil, fmt.Errorf("促销价格数据转换失败: %w", err)
		}
		promos = append(promos, promo)
	}
	
	return promos, nil
}

// GetExpiringPromotions 获取即将过期的促销
func (r *PromotionalPriceRepository) GetExpiringPromotions(ctx context.Context, hoursAhead int) ([]*types.PromotionalPrice, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	futureTime := now.Add(time.Hour * time.Duration(hoursAhead))
	
	result, err := model.Where("is_active", true).
		Where("valid_from <= ?", now).
		Where("valid_until > ?", now).
		Where("valid_until <= ?", futureTime).
		Order("valid_until ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取即将过期的促销失败: %w", err)
	}
	
	var promos []*types.PromotionalPrice
	for _, record := range result {
		promo := &types.PromotionalPrice{}
		err = record.Struct(promo)
		if err != nil {
			return nil, fmt.Errorf("促销价格数据转换失败: %w", err)
		}
		promos = append(promos, promo)
	}
	
	return promos, nil
}

// ActivatePromotions 激活到期的促销
func (r *PromotionalPriceRepository) ActivatePromotions(ctx context.Context, currentTime time.Time) (int, error) {
	updateData := map[string]interface{}{
		"is_active":  true,
		"updated_at": time.Now(),
	}
	
	model, err := r.Model(ctx)
	if err != nil {
		return 0, err
	}
	
	result, err := model.Where("valid_from <= ?", currentTime).
		Where("valid_until > ?", currentTime).
		Where("is_active", false).
		Update(updateData)
	if err != nil {
		return 0, fmt.Errorf("激活促销失败: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// DeactivateExpiredPromotions 停用过期的促销
func (r *PromotionalPriceRepository) DeactivateExpiredPromotions(ctx context.Context, currentTime time.Time) (int, error) {
	updateData := map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}
	
	model, err := r.Model(ctx)
	if err != nil {
		return 0, err
	}
	
	result, err := model.Where("valid_until <= ?", currentTime).
		Where("is_active", true).
		Update(updateData)
	if err != nil {
		return 0, fmt.Errorf("停用过期促销失败: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// GetPromotionStatistics 获取促销统计信息
func (r *PromotionalPriceRepository) GetPromotionStatistics(ctx context.Context) (map[string]int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]int)
	now := time.Now()
	
	// 总促销数
	total, err := r.Count(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("获取总促销数失败: %w", err)
	}
	stats["total"] = total
	
	// 活跃促销数
	active, err := model.Clone().Where("is_active", true).
		Where("valid_from <= ?", now).
		Where("valid_until > ?", now).
		Count()
	if err != nil {
		return nil, fmt.Errorf("获取活跃促销数失败: %w", err)
	}
	stats["active"] = active
	
	// 即将开始的促销数（未来24小时）
	upcoming, err := model.Clone().Where("is_active", true).
		Where("valid_from > ?", now).
		Where("valid_from <= ?", now.Add(24*time.Hour)).
		Count()
	if err != nil {
		return nil, fmt.Errorf("获取即将开始促销数失败: %w", err)
	}
	stats["upcoming"] = upcoming
	
	// 过期促销数
	expired, err := model.Clone().Where("valid_until <= ?", now).Count()
	if err != nil {
		return nil, fmt.Errorf("获取过期促销数失败: %w", err)
	}
	stats["expired"] = expired
	
	return stats, nil
}

// GetPromotionsByDateRange 按日期范围获取促销
func (r *PromotionalPriceRepository) GetPromotionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*types.PromotionalPrice, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Where("NOT (valid_until <= ? OR valid_from >= ?)", startDate, endDate).
		Order("valid_from ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("按日期范围获取促销失败: %w", err)
	}
	
	var promos []*types.PromotionalPrice
	for _, record := range result {
		promo := &types.PromotionalPrice{}
		err = record.Struct(promo)
		if err != nil {
			return nil, fmt.Errorf("促销价格数据转换失败: %w", err)
		}
		promos = append(promos, promo)
	}
	
	return promos, nil
}

// BatchUpdatePromotionStatus 批量更新促销状态
func (r *PromotionalPriceRepository) BatchUpdatePromotionStatus(ctx context.Context, promotionIDs []uint64, isActive bool) error {
	if len(promotionIDs) == 0 {
		return nil
	}
	
	updateData := map[string]interface{}{
		"is_active":  isActive,
		"updated_at": time.Now(),
	}
	
	model, err := r.Model(ctx)
	if err != nil {
		return err
	}
	
	_, err = model.WhereIn("id", promotionIDs).Update(updateData)
	if err != nil {
		return fmt.Errorf("批量更新促销状态失败: %w", err)
	}
	
	return nil
}

// ValidatePromotionData 验证促销数据完整性
func (r *PromotionalPriceRepository) ValidatePromotionData(ctx context.Context, promo *types.PromotionalPrice) error {
	// 检查商品是否存在
	productRepo := NewProductRepository()
	exists, err := productRepo.Exists(ctx, "id", promo.ProductID)
	if err != nil {
		return fmt.Errorf("验证商品存在性失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("商品不存在: %d", promo.ProductID)
	}
	
	// 验证价格
	if promo.PromotionalPrice.Amount <= 0 {
		return fmt.Errorf("促销价格必须大于0")
	}
	
	// 验证折扣百分比
	if promo.DiscountPercentage != nil {
		if *promo.DiscountPercentage < 0 || *promo.DiscountPercentage > 100 {
			return fmt.Errorf("折扣百分比必须在0-100之间")
		}
	}
	
	// 验证时间
	if promo.ValidUntil.Before(promo.ValidFrom) {
		return fmt.Errorf("促销结束时间不能早于开始时间")
	}
	
	// 验证促销时长不能超过1年
	if promo.ValidUntil.Sub(promo.ValidFrom) > 365*24*time.Hour {
		return fmt.Errorf("促销时长不能超过1年")
	}
	
	return nil
}
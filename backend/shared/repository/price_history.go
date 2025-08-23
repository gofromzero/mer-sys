package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// PriceHistoryRepository 价格历史仓储
type PriceHistoryRepository struct {
	*BaseRepository
}

// NewPriceHistoryRepository 创建价格历史仓储实例
func NewPriceHistoryRepository() *PriceHistoryRepository {
	baseRepo := NewBaseRepository()
	baseRepo.tableName = "price_histories"
	
	return &PriceHistoryRepository{
		BaseRepository: baseRepo,
	}
}

// CreatePriceHistory 创建价格历史记录
func (r *PriceHistoryRepository) CreatePriceHistory(ctx context.Context, req *types.PriceChangeRequest, productID uint64, oldPrice types.Money) (*types.PriceHistory, error) {
	history := &types.PriceHistory{
		ProductID:     productID,
		OldPrice:      oldPrice,
		NewPrice:      req.NewPrice,
		ChangeReason:  req.ChangeReason,
		ChangedBy:     r.GetUserID(ctx),
		EffectiveDate: req.EffectiveDate,
		CreatedAt:     time.Now(),
	}
	
	id, err := r.InsertAndGetId(ctx, history)
	if err != nil {
		return nil, fmt.Errorf("创建价格历史记录失败: %w", err)
	}
	
	history.ID = uint64(id)
	history.TenantID = r.GetTenantID(ctx)
	
	return history, nil
}

// GetPriceHistoryByID 根据ID获取价格历史记录
func (r *PriceHistoryRepository) GetPriceHistoryByID(ctx context.Context, id uint64) (*types.PriceHistory, error) {
	record, err := r.FindOne(ctx, "id", id)
	if err != nil {
		return nil, fmt.Errorf("获取价格历史记录失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	history := &types.PriceHistory{}
	err = record.Struct(history)
	if err != nil {
		return nil, fmt.Errorf("价格历史数据转换失败: %w", err)
	}
	
	return history, nil
}

// GetPriceHistoryByProductID 获取商品的价格历史记录
func (r *PriceHistoryRepository) GetPriceHistoryByProductID(ctx context.Context, productID uint64, limit int) ([]*types.PriceHistory, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	query := model.Where("product_id", productID).
		Order("effective_date DESC", "created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	result, err := query.All()
	if err != nil {
		return nil, fmt.Errorf("获取商品价格历史失败: %w", err)
	}
	
	var histories []*types.PriceHistory
	for _, record := range result {
		history := &types.PriceHistory{}
		err = record.Struct(history)
		if err != nil {
			return nil, fmt.Errorf("价格历史数据转换失败: %w", err)
		}
		histories = append(histories, history)
	}
	
	return histories, nil
}

// GetPriceHistoryByDateRange 按日期范围获取价格历史
func (r *PriceHistoryRepository) GetPriceHistoryByDateRange(ctx context.Context, productID uint64, startDate, endDate time.Time) ([]*types.PriceHistory, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Where("product_id", productID).
		Where("effective_date >= ?", startDate).
		Where("effective_date <= ?", endDate).
		Order("effective_date DESC").
		All()
	if err != nil {
		return nil, fmt.Errorf("按日期范围获取价格历史失败: %w", err)
	}
	
	var histories []*types.PriceHistory
	for _, record := range result {
		history := &types.PriceHistory{}
		err = record.Struct(history)
		if err != nil {
			return nil, fmt.Errorf("价格历史数据转换失败: %w", err)
		}
		histories = append(histories, history)
	}
	
	return histories, nil
}

// GetLatestPriceHistory 获取商品的最新价格历史记录
func (r *PriceHistoryRepository) GetLatestPriceHistory(ctx context.Context, productID uint64) (*types.PriceHistory, error) {
	record, err := r.FindOne(ctx, "product_id = ?", productID)
	if err != nil {
		return nil, fmt.Errorf("获取最新价格历史失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	history := &types.PriceHistory{}
	err = record.Struct(history)
	if err != nil {
		return nil, fmt.Errorf("价格历史数据转换失败: %w", err)
	}
	
	return history, nil
}

// GetPriceHistoryByUser 获取用户的价格变更历史
func (r *PriceHistoryRepository) GetPriceHistoryByUser(ctx context.Context, userID uint64, limit int) ([]*types.PriceHistory, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	query := model.Where("changed_by", userID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	result, err := query.All()
	if err != nil {
		return nil, fmt.Errorf("获取用户价格变更历史失败: %w", err)
	}
	
	var histories []*types.PriceHistory
	for _, record := range result {
		history := &types.PriceHistory{}
		err = record.Struct(history)
		if err != nil {
			return nil, fmt.Errorf("价格历史数据转换失败: %w", err)
		}
		histories = append(histories, history)
	}
	
	return histories, nil
}

// GetPriceChangeStatistics 获取价格变更统计信息
func (r *PriceHistoryRepository) GetPriceChangeStatistics(ctx context.Context, days int) (map[string]interface{}, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]interface{})
	startDate := time.Now().AddDate(0, 0, -days)
	
	// 总变更次数
	totalChanges, err := model.Clone().Where("created_at >= ?", startDate).Count()
	if err != nil {
		return nil, fmt.Errorf("获取总变更次数失败: %w", err)
	}
	stats["total_changes"] = totalChanges
	
	// 按天统计变更次数
	dailyResult, err := model.Clone().
		Fields("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("DATE(created_at)").
		Order("date").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取每日变更统计失败: %w", err)
	}
	
	dailyStats := make(map[string]int)
	for _, record := range dailyResult {
		date := record["date"].String()
		count := record["count"].Int()
		dailyStats[date] = count
	}
	stats["daily_changes"] = dailyStats
	
	// 按用户统计变更次数
	userResult, err := model.Clone().
		Fields("changed_by, COUNT(*) as count").
		Where("created_at >= ?", startDate).
		Group("changed_by").
		Order("count DESC").
		Limit(10).
		All()
	if err != nil {
		return nil, fmt.Errorf("获取用户变更统计失败: %w", err)
	}
	
	userStats := make(map[uint64]int)
	for _, record := range userResult {
		userID := record["changed_by"].Uint64()
		count := record["count"].Int()
		userStats[userID] = count
	}
	stats["user_changes"] = userStats
	
	return stats, nil
}

// GetPriceImpactAnalysis 获取价格变更影响分析
func (r *PriceHistoryRepository) GetPriceImpactAnalysis(ctx context.Context, productID uint64) (map[string]interface{}, error) {
	histories, err := r.GetPriceHistoryByProductID(ctx, productID, 0)
	if err != nil {
		return nil, err
	}
	
	if len(histories) == 0 {
		return map[string]interface{}{
			"total_changes":    0,
			"avg_change_rate":  0.0,
			"max_increase":     0.0,
			"max_decrease":     0.0,
			"volatility_score": 0.0,
		}, nil
	}
	
	analysis := make(map[string]interface{})
	analysis["total_changes"] = len(histories)
	
	// 计算价格变化统计
	var totalChangeRate float64
	var maxIncrease float64
	var maxDecrease float64
	var changes []float64
	
	for _, history := range histories {
		oldAmount := float64(history.OldPrice.Amount)
		newAmount := float64(history.NewPrice.Amount)
		
		if oldAmount > 0 {
			changeRate := (newAmount - oldAmount) / oldAmount * 100
			totalChangeRate += changeRate
			changes = append(changes, changeRate)
			
			if changeRate > maxIncrease {
				maxIncrease = changeRate
			}
			if changeRate < maxDecrease {
				maxDecrease = changeRate
			}
		}
	}
	
	if len(changes) > 0 {
		analysis["avg_change_rate"] = totalChangeRate / float64(len(changes))
		analysis["max_increase"] = maxIncrease
		analysis["max_decrease"] = maxDecrease
		
		// 计算波动性分数（标准差）
		avgChange := totalChangeRate / float64(len(changes))
		var variance float64
		for _, change := range changes {
			variance += (change - avgChange) * (change - avgChange)
		}
		volatility := variance / float64(len(changes))
		analysis["volatility_score"] = volatility
	}
	
	return analysis, nil
}

// DeleteOldPriceHistory 删除旧的价格历史记录
func (r *PriceHistoryRepository) DeleteOldPriceHistory(ctx context.Context, beforeDate time.Time) (int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return 0, err
	}
	
	result, err := model.Where("created_at < ?", beforeDate).Delete()
	if err != nil {
		return 0, fmt.Errorf("删除旧价格历史记录失败: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// GetPriceHistoryPage 分页获取价格历史记录
func (r *PriceHistoryRepository) GetPriceHistoryPage(ctx context.Context, productID uint64, page, pageSize int) ([]*types.PriceHistory, int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	query := model.Where("product_id", productID)
	
	// 获取总数
	total, err := query.Clone().Count()
	if err != nil {
		return nil, 0, fmt.Errorf("获取价格历史总数失败: %w", err)
	}
	
	// 获取分页数据
	result, err := query.Order("effective_date DESC", "created_at DESC").
		Page(page, pageSize).
		All()
	if err != nil {
		return nil, 0, fmt.Errorf("获取价格历史分页数据失败: %w", err)
	}
	
	var histories []*types.PriceHistory
	for _, record := range result {
		history := &types.PriceHistory{}
		err = record.Struct(history)
		if err != nil {
			return nil, 0, fmt.Errorf("价格历史数据转换失败: %w", err)
		}
		histories = append(histories, history)
	}
	
	return histories, total, nil
}

// GetPriceHistoryByReason 按变更原因获取价格历史
func (r *PriceHistoryRepository) GetPriceHistoryByReason(ctx context.Context, reason string, limit int) ([]*types.PriceHistory, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	query := model.Where("change_reason LIKE ?", "%"+reason+"%").
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	result, err := query.All()
	if err != nil {
		return nil, fmt.Errorf("按原因获取价格历史失败: %w", err)
	}
	
	var histories []*types.PriceHistory
	for _, record := range result {
		history := &types.PriceHistory{}
		err = record.Struct(history)
		if err != nil {
			return nil, fmt.Errorf("价格历史数据转换失败: %w", err)
		}
		histories = append(histories, history)
	}
	
	return histories, nil
}

// ValidatePriceChange 验证价格变更数据
func (r *PriceHistoryRepository) ValidatePriceChange(ctx context.Context, productID uint64, req *types.PriceChangeRequest) error {
	// 检查商品是否存在
	productRepo := NewProductRepository()
	exists, err := productRepo.Exists(ctx, "id", productID)
	if err != nil {
		return fmt.Errorf("验证商品存在性失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("商品不存在: %d", productID)
	}
	
	// 验证价格
	if req.NewPrice.Amount <= 0 {
		return fmt.Errorf("新价格必须大于0")
	}
	
	// 验证生效时间
	if req.EffectiveDate.Before(time.Now().Add(-24 * time.Hour)) {
		return fmt.Errorf("生效时间不能早于24小时前")
	}
	
	// 验证变更原因
	if len(req.ChangeReason) == 0 {
		return fmt.Errorf("变更原因不能为空")
	}
	if len(req.ChangeReason) > 255 {
		return fmt.Errorf("变更原因长度不能超过255个字符")
	}
	
	return nil
}

// GetRecentPriceChanges 获取最近的价格变更
func (r *PriceHistoryRepository) GetRecentPriceChanges(ctx context.Context, hours int) ([]*types.PriceHistory, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	since := time.Now().Add(time.Duration(-hours) * time.Hour)
	
	result, err := model.Where("created_at >= ?", since).
		Order("created_at DESC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取最近价格变更失败: %w", err)
	}
	
	var histories []*types.PriceHistory
	for _, record := range result {
		history := &types.PriceHistory{}
		err = record.Struct(history)
		if err != nil {
			return nil, fmt.Errorf("价格历史数据转换失败: %w", err)
		}
		histories = append(histories, history)
	}
	
	return histories, nil
}
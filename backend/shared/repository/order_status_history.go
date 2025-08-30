package repository

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// OrderStatusHistoryRepository 订单状态历史数据访问层
type OrderStatusHistoryRepository struct {
	*BaseRepository
}

// NewOrderStatusHistoryRepository 创建订单状态历史仓库实例
func NewOrderStatusHistoryRepository() *OrderStatusHistoryRepository {
	return &OrderStatusHistoryRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// Create 创建订单状态历史记录
func (r *OrderStatusHistoryRepository) Create(ctx context.Context, history *types.OrderStatusHistory) error {
	tenantID := r.GetTenantID(ctx)
	
	history.TenantID = tenantID
	
	_, err := g.DB().Model("order_status_history").Ctx(ctx).Data(history).Insert()
	if err != nil {
		return fmt.Errorf("创建订单状态历史记录失败: %v", err)
	}
	
	return nil
}

// GetByOrderID 根据订单ID获取状态历史记录
func (r *OrderStatusHistoryRepository) GetByOrderID(ctx context.Context, orderID uint64) ([]types.OrderStatusHistory, error) {
	tenantID := r.GetTenantID(ctx)
	
	var histories []types.OrderStatusHistory
	err := g.DB().Model("order_status_history").
		Ctx(ctx).
		Where("tenant_id = ? AND order_id = ?", tenantID, orderID).
		OrderAsc("created_at").
		Scan(&histories)
	if err != nil {
		return nil, fmt.Errorf("获取订单状态历史失败: %v", err)
	}
	
	return histories, nil
}

// GetLatestByOrderID 获取订单的最新状态历史记录
func (r *OrderStatusHistoryRepository) GetLatestByOrderID(ctx context.Context, orderID uint64) (*types.OrderStatusHistory, error) {
	tenantID := r.GetTenantID(ctx)
	
	var history types.OrderStatusHistory
	err := g.DB().Model("order_status_history").
		Ctx(ctx).
		Where("tenant_id = ? AND order_id = ?", tenantID, orderID).
		OrderDesc("created_at").
		Limit(1).
		Scan(&history)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("获取订单最新状态历史失败: %v", err)
	}
	
	return &history, nil
}

// GetByOrderIDs 批量获取订单的最新状态历史记录
func (r *OrderStatusHistoryRepository) GetByOrderIDs(ctx context.Context, orderIDs []uint64) (map[uint64]*types.OrderStatusHistory, error) {
	if len(orderIDs) == 0 {
		return make(map[uint64]*types.OrderStatusHistory), nil
	}
	
	tenantID := r.GetTenantID(ctx)
	
	// 使用窗口函数获取每个订单的最新状态历史记录
	query := `
		SELECT h.*
		FROM (
			SELECT *, 
				   ROW_NUMBER() OVER (PARTITION BY order_id ORDER BY created_at DESC) as rn
			FROM order_status_history 
			WHERE tenant_id = ? AND order_id IN (?)
		) h
		WHERE h.rn = 1
	`
	
	var histories []types.OrderStatusHistory
	err := g.DB().Raw(query, tenantID, orderIDs).Scan(&histories)
	if err != nil {
		return nil, fmt.Errorf("批量获取订单状态历史失败: %v", err)
	}
	
	// 转换为map
	result := make(map[uint64]*types.OrderStatusHistory)
	for i := range histories {
		result[histories[i].OrderID] = &histories[i]
	}
	
	return result, nil
}

// DeleteByOrderID 删除订单的所有状态历史记录（用于订单删除时的级联操作）
func (r *OrderStatusHistoryRepository) DeleteByOrderID(ctx context.Context, orderID uint64) error {
	tenantID := r.GetTenantID(ctx)
	
	_, err := g.DB().Model("order_status_history").
		Ctx(ctx).
		Where("tenant_id = ? AND order_id = ?", tenantID, orderID).
		Delete()
	if err != nil {
		return fmt.Errorf("删除订单状态历史记录失败: %v", err)
	}
	
	return nil
}

// CountByStatus 按状态统计历史记录数量
func (r *OrderStatusHistoryRepository) CountByStatus(ctx context.Context, status types.OrderStatusInt, startDate, endDate *string) (int64, error) {
	tenantID := r.GetTenantID(ctx)
	
	query := g.DB().Model("order_status_history").
		Ctx(ctx).
		Where("tenant_id = ? AND to_status = ?", tenantID, status)
	
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}
	
	count, err := query.Count()
	if err != nil {
		return 0, fmt.Errorf("按状态统计历史记录失败: %v", err)
	}
	
	return int64(count), nil
}
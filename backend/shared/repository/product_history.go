package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// ProductHistoryRepository 商品历史数据访问层
type ProductHistoryRepository struct {
	*BaseRepository
}

// NewProductHistoryRepository 创建商品历史仓库实例
func NewProductHistoryRepository() *ProductHistoryRepository {
	return &ProductHistoryRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// RecordChange 记录商品变更
func (r *ProductHistoryRepository) RecordChange(ctx context.Context, productID uint64, operation types.ChangeOperation, changes map[string]interface{}) error {
	tenantID := r.GetTenantID(ctx)
	userID := r.GetUserID(ctx)
	if tenantID == 0 || userID == 0 {
		return fmt.Errorf("missing tenant_id or user_id in context")
	}
	
	// 获取当前商品版本
	var version int
	err := g.DB().Model("products").
		Ctx(ctx).
		Where("id = ? AND tenant_id = ?", productID, tenantID).
		Fields("version").
		Scan(&version)
	if err != nil {
		return err
	}
	
	// 序列化变更数据
	changesJSON, err := json.Marshal(changes)
	if err != nil {
		return err
	}
	
	history := &types.ProductHistory{
		TenantID:  tenantID,
		ProductID: productID,
		Version:   version,
		FieldName: "bulk_change",
		OldValue:  "",
		NewValue:  string(changesJSON),
		Operation: operation,
		ChangedBy: userID,
	}
	
	_, err = g.DB().Model("product_histories").Ctx(ctx).Insert(history)
	return err
}

// RecordFieldChange 记录单个字段变更
func (r *ProductHistoryRepository) RecordFieldChange(ctx context.Context, productID uint64, fieldName string, change interface{}) error {
	tenantID := r.GetTenantID(ctx)
	userID := r.GetUserID(ctx)
	if tenantID == 0 || userID == 0 {
		return fmt.Errorf("missing tenant_id or user_id in context")
	}
	
	// 获取当前商品版本
	var version int
	err := g.DB().Model("products").
		Ctx(ctx).
		Where("id = ? AND tenant_id = ?", productID, tenantID).
		Fields("version").
		Scan(&version)
	if err != nil {
		return err
	}
	
	var oldValue, newValue string
	
	// 解析变更数据
	if changeMap, ok := change.(map[string]interface{}); ok {
		if old, exists := changeMap["old"]; exists {
			if oldBytes, err := json.Marshal(old); err == nil {
				oldValue = string(oldBytes)
			}
		}
		if new, exists := changeMap["new"]; exists {
			if newBytes, err := json.Marshal(new); err == nil {
				newValue = string(newBytes)
			}
		}
	}
	
	history := &types.ProductHistory{
		TenantID:  tenantID,
		ProductID: productID,
		Version:   version,
		FieldName: fieldName,
		OldValue:  oldValue,
		NewValue:  newValue,
		Operation: types.ChangeOperationUpdate,
		ChangedBy: userID,
	}
	
	_, err = g.DB().Model("product_histories").Ctx(ctx).Insert(history)
	return err
}

// RecordStatusChange 记录状态变更
func (r *ProductHistoryRepository) RecordStatusChange(ctx context.Context, productID uint64, oldStatus, newStatus types.ProductStatus) error {
	return r.RecordFieldChange(ctx, productID, "status", map[string]interface{}{
		"old": oldStatus,
		"new": newStatus,
	})
}

// GetProductHistory 获取商品变更历史
func (r *ProductHistoryRepository) GetProductHistory(ctx context.Context, productID uint64) ([]types.ProductHistory, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	var histories []types.ProductHistory
	err := g.DB().Model("product_histories ph").
		LeftJoin("users u", "ph.changed_by = u.id").
		Fields("ph.*, u.username as changed_by_name").
		Where("ph.product_id = ? AND ph.tenant_id = ?", productID, tenantID).
		Order("ph.changed_at DESC").
		Ctx(ctx).
		Scan(&histories)
	
	if err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetVersionHistory 获取指定版本的历史记录
func (r *ProductHistoryRepository) GetVersionHistory(ctx context.Context, productID uint64, version int) ([]types.ProductHistory, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	var histories []types.ProductHistory
	err := g.DB().Model("product_histories").
		Ctx(ctx).
		Where("product_id = ? AND tenant_id = ? AND version = ?", productID, tenantID, version).
		Order("changed_at DESC").
		Scan(&histories)
	
	if err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetUserChanges 获取用户的变更记录
func (r *ProductHistoryRepository) GetUserChanges(ctx context.Context, userID uint64, limit int) ([]types.ProductHistory, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}
	
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	
	var histories []types.ProductHistory
	err := g.DB().Model("product_histories ph").
		LeftJoin("products p", "ph.product_id = p.id").
		Fields("ph.*, p.name as product_name").
		Where("ph.changed_by = ? AND ph.tenant_id = ?", userID, tenantID).
		Order("ph.changed_at DESC").
		Limit(limit).
		Ctx(ctx).
		Scan(&histories)
	
	if err != nil {
		return nil, err
	}
	
	return histories, nil
}

// CleanupOldHistories 清理旧的历史记录
func (r *ProductHistoryRepository) CleanupOldHistories(ctx context.Context, keepDays int) error {
	if keepDays <= 0 {
		keepDays = 90 // 默认保留90天
	}
	
	result, err := g.DB().Exec(ctx, 
		"DELETE FROM product_histories WHERE changed_at < DATE_SUB(NOW(), INTERVAL ? DAY)",
		keepDays,
	)
	
	if err != nil {
		return err
	}
	
	affected, _ := result.RowsAffected()
	g.Log().Infof(ctx, "Cleaned up %d old product history records", affected)
	
	return nil
}
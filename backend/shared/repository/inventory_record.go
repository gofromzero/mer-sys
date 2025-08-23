package repository

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IInventoryRecordRepository 库存记录Repository接口
type IInventoryRecordRepository interface {
	Create(ctx context.Context, record *types.InventoryRecord) error
	GetByProductID(ctx context.Context, tenantID, productID uint64, page, pageSize int) ([]types.InventoryRecord, int64, error)
	GetByReferenceID(ctx context.Context, tenantID uint64, referenceID string) ([]types.InventoryRecord, error)
	GetRecentRecords(ctx context.Context, tenantID uint64, limit int) ([]types.InventoryRecord, error)
}

// inventoryRecordRepository 库存记录Repository实现
type inventoryRecordRepository struct {
	*BaseRepository
	tableName string
}

// NewInventoryRecordRepository 创建库存记录Repository
func NewInventoryRecordRepository() IInventoryRecordRepository {
	return &inventoryRecordRepository{
		BaseRepository: NewBaseRepository(),
		tableName:     "inventory_records",
	}
}

// Create 创建库存记录
func (r *inventoryRecordRepository) Create(ctx context.Context, record *types.InventoryRecord) error {
	if record == nil {
		return errors.New("库存记录不能为空")
	}

	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return errors.New("租户ID不能为空")
	}
	record.TenantID = tenantID

	_, err := g.DB().Model(r.tableName).Ctx(ctx).Insert(record)
	if err != nil {
		g.Log().Errorf(ctx, "创建库存记录失败: %v", err)
		return err
	}

	return nil
}

// GetByProductID 根据商品ID获取库存记录
func (r *inventoryRecordRepository) GetByProductID(ctx context.Context, tenantID, productID uint64, page, pageSize int) ([]types.InventoryRecord, int64, error) {
	if productID == 0 {
		return nil, 0, errors.New("商品ID不能为空")
	}

	var records []types.InventoryRecord
	offset := (page - 1) * pageSize

	db := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset)

	err := db.Scan(&records)
	if err != nil {
		g.Log().Errorf(ctx, "查询库存记录失败: %v", err)
		return nil, 0, err
	}

	// 获取总数
	count, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Count()
	if err != nil {
		g.Log().Errorf(ctx, "查询库存记录总数失败: %v", err)
		return records, 0, err
	}

	return records, int64(count), nil
}

// GetByReferenceID 根据关联ID获取库存记录
func (r *inventoryRecordRepository) GetByReferenceID(ctx context.Context, tenantID uint64, referenceID string) ([]types.InventoryRecord, error) {
	if referenceID == "" {
		return nil, errors.New("关联ID不能为空")
	}

	var records []types.InventoryRecord
	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND reference_id = ?", tenantID, referenceID).
		Order("created_at DESC").
		Scan(&records)
	if err != nil {
		g.Log().Errorf(ctx, "根据关联ID查询库存记录失败: %v", err)
		return nil, err
	}

	return records, nil
}

// GetRecentRecords 获取最近的库存记录
func (r *inventoryRecordRepository) GetRecentRecords(ctx context.Context, tenantID uint64, limit int) ([]types.InventoryRecord, error) {
	var records []types.InventoryRecord
	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Scan(&records)
	if err != nil {
		g.Log().Errorf(ctx, "查询最近库存记录失败: %v", err)
		return nil, err
	}

	return records, nil
}
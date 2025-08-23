package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IInventoryReservationRepository 库存预留Repository接口
type IInventoryReservationRepository interface {
	Create(ctx context.Context, reservation *types.InventoryReservation) error
	GetByID(ctx context.Context, tenantID, reservationID uint64) (*types.InventoryReservation, error)
	GetByReference(ctx context.Context, tenantID uint64, referenceType, referenceID string) ([]types.InventoryReservation, error)
	GetActiveReservations(ctx context.Context, tenantID, productID uint64) ([]types.InventoryReservation, error)
	UpdateStatus(ctx context.Context, tenantID, reservationID uint64, status types.ReservationStatus) error
	GetExpiredReservations(ctx context.Context, tenantID uint64) ([]types.InventoryReservation, error)
	GetTotalReservedQuantity(ctx context.Context, tenantID, productID uint64) (int, error)
	Delete(ctx context.Context, tenantID, reservationID uint64) error
}

// inventoryReservationRepository 库存预留Repository实现
type inventoryReservationRepository struct {
	*BaseRepository
	tableName string
}

// NewInventoryReservationRepository 创建库存预留Repository
func NewInventoryReservationRepository() IInventoryReservationRepository {
	return &inventoryReservationRepository{
		BaseRepository: NewBaseRepository(),
		tableName:     "inventory_reservations",
	}
}

// Create 创建库存预留记录
func (r *inventoryReservationRepository) Create(ctx context.Context, reservation *types.InventoryReservation) error {
	if reservation == nil {
		return errors.New("库存预留记录不能为空")
	}

	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return errors.New("租户ID不能为空")
	}
	reservation.TenantID = tenantID
	reservation.CreatedAt = time.Now()
	reservation.UpdatedAt = time.Now()

	_, err := g.DB().Model(r.tableName).Ctx(ctx).Insert(reservation)
	if err != nil {
		g.Log().Errorf(ctx, "创建库存预留记录失败: %v", err)
		return err
	}

	return nil
}

// GetByID 根据ID获取库存预留记录
func (r *inventoryReservationRepository) GetByID(ctx context.Context, tenantID, reservationID uint64) (*types.InventoryReservation, error) {
	if reservationID == 0 {
		return nil, errors.New("预留ID不能为空")
	}

	var reservation types.InventoryReservation
	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, reservationID).
		Scan(&reservation)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("库存预留记录不存在")
		}
		g.Log().Errorf(ctx, "查询库存预留记录失败: %v", err)
		return nil, err
	}

	return &reservation, nil
}

// GetByReference 根据关联信息获取库存预留记录
func (r *inventoryReservationRepository) GetByReference(ctx context.Context, tenantID uint64, referenceType, referenceID string) ([]types.InventoryReservation, error) {
	if referenceType == "" || referenceID == "" {
		return nil, errors.New("关联信息不能为空")
	}

	var reservations []types.InventoryReservation
	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND reference_type = ? AND reference_id = ?", tenantID, referenceType, referenceID).
		Order("created_at DESC").
		Scan(&reservations)
	if err != nil {
		g.Log().Errorf(ctx, "根据关联信息查询库存预留记录失败: %v", err)
		return nil, err
	}

	return reservations, nil
}

// GetActiveReservations 获取商品的活跃预留记录
func (r *inventoryReservationRepository) GetActiveReservations(ctx context.Context, tenantID, productID uint64) ([]types.InventoryReservation, error) {
	if productID == 0 {
		return nil, errors.New("商品ID不能为空")
	}

	var reservations []types.InventoryReservation
	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND product_id = ? AND status = ?", tenantID, productID, types.ReservationStatusActive).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("created_at ASC").
		Scan(&reservations)
	if err != nil {
		g.Log().Errorf(ctx, "查询活跃库存预留记录失败: %v", err)
		return nil, err
	}

	return reservations, nil
}

// UpdateStatus 更新预留状态
func (r *inventoryReservationRepository) UpdateStatus(ctx context.Context, tenantID, reservationID uint64, status types.ReservationStatus) error {
	if reservationID == 0 {
		return errors.New("预留ID不能为空")
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, reservationID).
		Update(g.Map{
			"status":     status,
			"updated_at": time.Now(),
		})
	if err != nil {
		g.Log().Errorf(ctx, "更新库存预留状态失败: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("库存预留记录不存在或无权限")
	}

	return nil
}

// GetExpiredReservations 获取过期的预留记录
func (r *inventoryReservationRepository) GetExpiredReservations(ctx context.Context, tenantID uint64) ([]types.InventoryReservation, error) {
	var reservations []types.InventoryReservation
	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND status = ? AND expires_at <= ?", tenantID, types.ReservationStatusActive, time.Now()).
		Scan(&reservations)
	if err != nil {
		g.Log().Errorf(ctx, "查询过期库存预留记录失败: %v", err)
		return nil, err
	}

	return reservations, nil
}

// GetTotalReservedQuantity 获取商品总预留数量
func (r *inventoryReservationRepository) GetTotalReservedQuantity(ctx context.Context, tenantID, productID uint64) (int, error) {
	if productID == 0 {
		return 0, errors.New("商品ID不能为空")
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND product_id = ? AND status = ?", tenantID, productID, types.ReservationStatusActive).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Sum("reserved_quantity")
	
	total := int(result)
	if err != nil {
		g.Log().Errorf(ctx, "计算总预留数量失败: %v", err)
		return 0, err
	}

	return total, nil
}

// Delete 删除预留记录
func (r *inventoryReservationRepository) Delete(ctx context.Context, tenantID, reservationID uint64) error {
	if reservationID == 0 {
		return errors.New("预留ID不能为空")
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, reservationID).
		Delete()
	if err != nil {
		g.Log().Errorf(ctx, "删除库存预留记录失败: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("库存预留记录不存在或无权限")
	}

	return nil
}
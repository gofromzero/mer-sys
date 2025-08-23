package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IInventoryAlertRepository 库存预警Repository接口
type IInventoryAlertRepository interface {
	Create(ctx context.Context, alert *types.InventoryAlert) error
	GetByID(ctx context.Context, tenantID, alertID uint64) (*types.InventoryAlert, error)
	GetByProductID(ctx context.Context, tenantID, productID uint64) ([]types.InventoryAlert, error)
	GetActiveAlerts(ctx context.Context, tenantID uint64) ([]types.InventoryAlert, error)
	Update(ctx context.Context, tenantID uint64, alert *types.InventoryAlert) error
	UpdateLastTriggered(ctx context.Context, tenantID, alertID uint64) error
	Delete(ctx context.Context, tenantID, alertID uint64) error
	ToggleStatus(ctx context.Context, tenantID, alertID uint64, isActive bool) error
}

// inventoryAlertRepository 库存预警Repository实现
type inventoryAlertRepository struct {
	*BaseRepository
	tableName string
}

// NewInventoryAlertRepository 创建库存预警Repository
func NewInventoryAlertRepository() IInventoryAlertRepository {
	return &inventoryAlertRepository{
		BaseRepository: NewBaseRepository(),
		tableName:     "inventory_alerts",
	}
}

// Create 创建库存预警规则
func (r *inventoryAlertRepository) Create(ctx context.Context, alert *types.InventoryAlert) error {
	if alert == nil {
		return errors.New("库存预警规则不能为空")
	}

	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return errors.New("租户ID不能为空")
	}
	alert.TenantID = tenantID
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()

	// 序列化通知渠道
	channels, err := json.Marshal(alert.NotificationChannels)
	if err != nil {
		g.Log().Errorf(ctx, "序列化通知渠道失败: %v", err)
		return err
	}

	data := g.Map{
		"tenant_id":              alert.TenantID,
		"product_id":             alert.ProductID,
		"alert_type":             string(alert.AlertType),
		"threshold_value":        alert.ThresholdValue,
		"notification_channels":  string(channels),
		"is_active":              alert.IsActive,
		"created_at":             alert.CreatedAt,
		"updated_at":             alert.UpdatedAt,
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).Insert(data)
	if err != nil {
		g.Log().Errorf(ctx, "创建库存预警规则失败: %v", err)
		return err
	}

	// 获取插入的ID
	lastInsertID, err := result.LastInsertId()
	if err == nil {
		alert.ID = uint64(lastInsertID)
	}

	return nil
}

// GetByID 根据ID获取库存预警规则
func (r *inventoryAlertRepository) GetByID(ctx context.Context, tenantID, alertID uint64) (*types.InventoryAlert, error) {
	if alertID == 0 {
		return nil, errors.New("预警ID不能为空")
	}

	var alert types.InventoryAlert
	var channelsJSON string

	err := g.DB().Model(r.tableName).Ctx(ctx).
		Fields("*, notification_channels as channels_json").
		Where("tenant_id = ? AND id = ?", tenantID, alertID).
		Scan(&alert)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("库存预警规则不存在")
		}
		g.Log().Errorf(ctx, "查询库存预警规则失败: %v", err)
		return nil, err
	}

	// 反序列化通知渠道
	if err := json.Unmarshal([]byte(channelsJSON), &alert.NotificationChannels); err != nil {
		g.Log().Errorf(ctx, "反序列化通知渠道失败: %v", err)
	}

	return &alert, nil
}

// GetByProductID 根据商品ID获取库存预警规则
func (r *inventoryAlertRepository) GetByProductID(ctx context.Context, tenantID, productID uint64) ([]types.InventoryAlert, error) {
	if productID == 0 {
		return nil, errors.New("商品ID不能为空")
	}

	var alerts []types.InventoryAlert
	var results []g.Map

	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Order("created_at DESC").
		Scan(&results)
	if err != nil {
		g.Log().Errorf(ctx, "查询商品库存预警规则失败: %v", err)
		return nil, err
	}

	// 转换数据并反序列化通知渠道
	for _, result := range results {
		var alert types.InventoryAlert
		if err := gconv.Struct(result, &alert); err != nil {
			continue
		}

		if channelsStr, ok := result["notification_channels"].(string); ok {
			if err := json.Unmarshal([]byte(channelsStr), &alert.NotificationChannels); err != nil {
				g.Log().Errorf(ctx, "反序列化通知渠道失败: %v", err)
			}
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetActiveAlerts 获取活跃的库存预警规则
func (r *inventoryAlertRepository) GetActiveAlerts(ctx context.Context, tenantID uint64) ([]types.InventoryAlert, error) {
	var alerts []types.InventoryAlert
	var results []g.Map

	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("created_at DESC").
		Scan(&results)
	if err != nil {
		g.Log().Errorf(ctx, "查询活跃库存预警规则失败: %v", err)
		return nil, err
	}

	// 转换数据并反序列化通知渠道
	for _, result := range results {
		var alert types.InventoryAlert
		if err := gconv.Struct(result, &alert); err != nil {
			continue
		}

		if channelsStr, ok := result["notification_channels"].(string); ok {
			if err := json.Unmarshal([]byte(channelsStr), &alert.NotificationChannels); err != nil {
				g.Log().Errorf(ctx, "反序列化通知渠道失败: %v", err)
			}
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// Update 更新库存预警规则
func (r *inventoryAlertRepository) Update(ctx context.Context, tenantID uint64, alert *types.InventoryAlert) error {
	if alert == nil || alert.ID == 0 {
		return errors.New("库存预警规则信息不完整")
	}

	alert.UpdatedAt = time.Now()

	// 序列化通知渠道
	channels, err := json.Marshal(alert.NotificationChannels)
	if err != nil {
		g.Log().Errorf(ctx, "序列化通知渠道失败: %v", err)
		return err
	}

	data := g.Map{
		"alert_type":             string(alert.AlertType),
		"threshold_value":        alert.ThresholdValue,
		"notification_channels":  string(channels),
		"is_active":              alert.IsActive,
		"updated_at":             alert.UpdatedAt,
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, alert.ID).
		Update(data)
	if err != nil {
		g.Log().Errorf(ctx, "更新库存预警规则失败: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("库存预警规则不存在或无权限")
	}

	return nil
}

// UpdateLastTriggered 更新最后触发时间
func (r *inventoryAlertRepository) UpdateLastTriggered(ctx context.Context, tenantID, alertID uint64) error {
	if alertID == 0 {
		return errors.New("预警ID不能为空")
	}

	_, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, alertID).
		Update(g.Map{
			"last_triggered_at": time.Now(),
			"updated_at":        time.Now(),
		})
	if err != nil {
		g.Log().Errorf(ctx, "更新预警触发时间失败: %v", err)
		return err
	}

	return nil
}

// Delete 删除库存预警规则
func (r *inventoryAlertRepository) Delete(ctx context.Context, tenantID, alertID uint64) error {
	if alertID == 0 {
		return errors.New("预警ID不能为空")
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, alertID).
		Delete()
	if err != nil {
		g.Log().Errorf(ctx, "删除库存预警规则失败: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("库存预警规则不存在或无权限")
	}

	return nil
}

// ToggleStatus 切换预警状态
func (r *inventoryAlertRepository) ToggleStatus(ctx context.Context, tenantID, alertID uint64, isActive bool) error {
	if alertID == 0 {
		return errors.New("预警ID不能为空")
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, alertID).
		Update(g.Map{
			"is_active":  isActive,
			"updated_at": time.Now(),
		})
	if err != nil {
		g.Log().Errorf(ctx, "切换预警状态失败: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("库存预警规则不存在或无权限")
	}

	return nil
}
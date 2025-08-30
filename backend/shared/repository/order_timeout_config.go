package repository

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// OrderTimeoutConfigRepository 订单超时配置数据访问层
type OrderTimeoutConfigRepository struct {
	*BaseRepository
}

// NewOrderTimeoutConfigRepository 创建订单超时配置仓库实例
func NewOrderTimeoutConfigRepository() *OrderTimeoutConfigRepository {
	return &OrderTimeoutConfigRepository{
		BaseRepository: NewBaseRepository(),
	}
}

// Create 创建订单超时配置
func (r *OrderTimeoutConfigRepository) Create(ctx context.Context, config *types.OrderTimeoutConfig) error {
	tenantID := r.GetTenantID(ctx)
	
	config.TenantID = tenantID
	
	_, err := g.DB().Model("order_timeout_configs").Ctx(ctx).Data(config).Insert()
	if err != nil {
		return fmt.Errorf("创建订单超时配置失败: %v", err)
	}
	
	return nil
}

// GetByMerchantID 根据商户ID获取超时配置
func (r *OrderTimeoutConfigRepository) GetByMerchantID(ctx context.Context, merchantID uint64) (*types.OrderTimeoutConfig, error) {
	tenantID := r.GetTenantID(ctx)
	
	var config types.OrderTimeoutConfig
	err := g.DB().Model("order_timeout_configs").
		Ctx(ctx).
		Where("tenant_id = ? AND merchant_id = ?", tenantID, merchantID).
		Scan(&config)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("获取商户超时配置失败: %v", err)
	}
	
	return &config, nil
}

// GetDefaultConfig 获取租户默认超时配置
func (r *OrderTimeoutConfigRepository) GetDefaultConfig(ctx context.Context) (*types.OrderTimeoutConfig, error) {
	tenantID := r.GetTenantID(ctx)
	
	var config types.OrderTimeoutConfig
	err := g.DB().Model("order_timeout_configs").
		Ctx(ctx).
		Where("tenant_id = ? AND merchant_id IS NULL", tenantID).
		Scan(&config)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// 返回默认配置值
			return &types.OrderTimeoutConfig{
				TenantID:               tenantID,
				MerchantID:            nil,
				PaymentTimeoutMinutes: 30,
				ProcessingTimeoutHours: 24,
				AutoCompleteEnabled:   false,
			}, nil
		}
		return nil, fmt.Errorf("获取默认超时配置失败: %v", err)
	}
	
	return &config, nil
}

// GetEffectiveConfig 获取有效的超时配置（优先商户级配置，否则使用默认配置）
func (r *OrderTimeoutConfigRepository) GetEffectiveConfig(ctx context.Context, merchantID uint64) (*types.OrderTimeoutConfig, error) {
	// 先尝试获取商户级配置
	config, err := r.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	
	// 如果没有商户级配置，使用默认配置
	if config == nil {
		config, err = r.GetDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	
	return config, nil
}

// Update 更新订单超时配置
func (r *OrderTimeoutConfigRepository) Update(ctx context.Context, config *types.OrderTimeoutConfig) error {
	tenantID := r.GetTenantID(ctx)
	
	whereCondition := g.Map{
		"tenant_id": tenantID,
		"id":       config.ID,
	}
	
	_, err := g.DB().Model("order_timeout_configs").
		Ctx(ctx).
		Where(whereCondition).
		Data(config).
		Update()
	if err != nil {
		return fmt.Errorf("更新订单超时配置失败: %v", err)
	}
	
	return nil
}

// Delete 删除订单超时配置
func (r *OrderTimeoutConfigRepository) Delete(ctx context.Context, id uint64) error {
	tenantID := r.GetTenantID(ctx)
	
	_, err := g.DB().Model("order_timeout_configs").
		Ctx(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete()
	if err != nil {
		return fmt.Errorf("删除订单超时配置失败: %v", err)
	}
	
	return nil
}

// ListByTenant 获取租户的所有超时配置
func (r *OrderTimeoutConfigRepository) ListByTenant(ctx context.Context) ([]types.OrderTimeoutConfig, error) {
	tenantID := r.GetTenantID(ctx)
	
	var configs []types.OrderTimeoutConfig
	err := g.DB().Model("order_timeout_configs").
		Ctx(ctx).
		Where("tenant_id = ?", tenantID).
		OrderAsc("merchant_id").
		Scan(&configs)
	if err != nil {
		return nil, fmt.Errorf("获取租户超时配置列表失败: %v", err)
	}
	
	return configs, nil
}
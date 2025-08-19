package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/cache"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// TenantConfigCache 租户配置缓存管理器
type TenantConfigCache struct {
	cache *cache.Cache
}

// NewTenantConfigCache 创建租户配置缓存管理器
func NewTenantConfigCache() *TenantConfigCache {
	return &TenantConfigCache{
		cache: cache.NewCache("tenant_config"),
	}
}

// NewTestTenantConfigCache 创建测试专用的租户配置缓存管理器
func NewTestTenantConfigCache() *TenantConfigCache {
	return &TenantConfigCache{
		cache: cache.NewMockCache(),
	}
}

// buildConfigKey 构建配置缓存键
func (c *TenantConfigCache) buildConfigKey(tenantID uint64) string {
	return fmt.Sprintf("config:%d", tenantID)
}

// buildNotificationKey 构建通知缓存键
func (c *TenantConfigCache) buildNotificationKey(tenantID uint64) string {
	return fmt.Sprintf("notification:%d", tenantID)
}

// SetConfig 设置租户配置到缓存
func (c *TenantConfigCache) SetConfig(ctx context.Context, tenantID uint64, config *types.TenantConfig) error {
	key := c.buildConfigKey(tenantID)
	// 缓存1小时
	return c.cache.Set(ctx, key, config, time.Hour)
}

// GetConfig 从缓存获取租户配置
func (c *TenantConfigCache) GetConfig(ctx context.Context, tenantID uint64) (*types.TenantConfig, error) {
	key := c.buildConfigKey(tenantID)
	var config types.TenantConfig
	err := c.cache.GetStruct(ctx, key, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// InvalidateConfig 使租户配置缓存失效
func (c *TenantConfigCache) InvalidateConfig(ctx context.Context, tenantID uint64) error {
	key := c.buildConfigKey(tenantID)
	return c.cache.Delete(ctx, key)
}

// SetConfigChangeNotification 设置配置变更通知
func (c *TenantConfigCache) SetConfigChangeNotification(ctx context.Context, tenantID uint64, changeInfo map[string]interface{}) error {
	key := c.buildNotificationKey(tenantID)
	// 通知缓存5分钟
	return c.cache.Set(ctx, key, changeInfo, 5*time.Minute)
}

// GetConfigChangeNotification 获取配置变更通知
func (c *TenantConfigCache) GetConfigChangeNotification(ctx context.Context, tenantID uint64) (map[string]interface{}, error) {
	key := c.buildNotificationKey(tenantID)
	var changeInfo map[string]interface{}
	err := c.cache.GetStruct(ctx, key, &changeInfo)
	if err != nil {
		return nil, err
	}
	return changeInfo, nil
}

// ClearConfigChangeNotification 清除配置变更通知
func (c *TenantConfigCache) ClearConfigChangeNotification(ctx context.Context, tenantID uint64) error {
	key := c.buildNotificationKey(tenantID)
	return c.cache.Delete(ctx, key)
}

// IsConfigCached 检查配置是否已缓存
func (c *TenantConfigCache) IsConfigCached(ctx context.Context, tenantID uint64) (bool, error) {
	key := c.buildConfigKey(tenantID)
	return c.cache.Exists(ctx, key)
}

// PrefetchConfigs 预取多个租户的配置到缓存
func (c *TenantConfigCache) PrefetchConfigs(ctx context.Context, tenantIDs []uint64, configs map[uint64]*types.TenantConfig) error {
	for _, tenantID := range tenantIDs {
		if config, exists := configs[tenantID]; exists {
			if err := c.SetConfig(ctx, tenantID, config); err != nil {
				// 记录错误但继续处理其他配置
				continue
			}
		}
	}
	return nil
}

// GetCacheStats 获取缓存统计信息
func (c *TenantConfigCache) GetCacheStats(ctx context.Context, tenantIDs []uint64) map[string]interface{} {
	stats := map[string]interface{}{
		"total_tenants": len(tenantIDs),
		"cached_count":  0,
		"cache_details": make(map[uint64]bool),
	}

	cachedCount := 0
	details := make(map[uint64]bool)

	for _, tenantID := range tenantIDs {
		cached, err := c.IsConfigCached(ctx, tenantID)
		if err == nil && cached {
			cachedCount++
			details[tenantID] = true
		} else {
			details[tenantID] = false
		}
	}

	stats["cached_count"] = cachedCount
	stats["cache_details"] = details
	stats["cache_hit_rate"] = float64(cachedCount) / float64(len(tenantIDs))

	return stats
}
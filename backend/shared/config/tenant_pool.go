package config

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantPoolConfig 租户连接池配置
type TenantPoolConfig struct {
	TenantID         uint64 `json:"tenant_id"`
	MaxIdleConnCount int    `json:"max_idle_conn_count"`
	MaxOpenConnCount int    `json:"max_open_conn_count"`
	MaxConnLifeTime  int    `json:"max_conn_life_time"` // 秒
	DatabaseName     string `json:"database_name"`
	ReadOnly         bool   `json:"read_only"`
}

// TenantPoolManager 租户连接池管理器
type TenantPoolManager struct {
	pools      map[uint64]gdb.DB
	configs    map[uint64]*TenantPoolConfig
	defaultDB  gdb.DB
	mutex      sync.RWMutex
}

var (
	poolManager *TenantPoolManager
	poolOnce    sync.Once
)

// GetTenantPoolManager 获取租户连接池管理器单例
func GetTenantPoolManager() *TenantPoolManager {
	poolOnce.Do(func() {
		poolManager = &TenantPoolManager{
			pools:     make(map[uint64]gdb.DB),
			configs:   make(map[uint64]*TenantPoolConfig),
			defaultDB: g.DB(),
		}
		
		// 初始化默认租户配置
		poolManager.initDefaultConfigs()
	})
	return poolManager
}

// initDefaultConfigs 初始化默认租户配置
func (m *TenantPoolManager) initDefaultConfigs() {
	// 默认租户配置
	defaultConfigs := []*TenantPoolConfig{
		{
			TenantID:         1,
			MaxIdleConnCount: 5,
			MaxOpenConnCount: 20,
			MaxConnLifeTime:  300, // 5分钟
			DatabaseName:     "mer_system",
			ReadOnly:         false,
		},
		{
			TenantID:         2,
			MaxIdleConnCount: 2,
			MaxOpenConnCount: 10,
			MaxConnLifeTime:  300,
			DatabaseName:     "mer_system",
			ReadOnly:         false,
		},
	}
	
	for _, config := range defaultConfigs {
		m.configs[config.TenantID] = config
	}
}

// GetTenantDB 获取租户特定的数据库连接
func (m *TenantPoolManager) GetTenantDB(ctx context.Context, tenantID uint64) (gdb.DB, error) {
	m.mutex.RLock()
	
	// 检查是否已有该租户的连接池
	if db, exists := m.pools[tenantID]; exists {
		m.mutex.RUnlock()
		return db, nil
	}
	m.mutex.RUnlock()
	
	// 获取租户配置
	config, err := m.getTenantConfig(tenantID)
	if err != nil {
		return m.defaultDB, err
	}
	
	// 创建租户特定的连接池
	return m.createTenantPool(ctx, config)
}

// getTenantConfig 获取租户配置
func (m *TenantPoolManager) getTenantConfig(tenantID uint64) (*TenantPoolConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if config, exists := m.configs[tenantID]; exists {
		return config, nil
	}
	
	// 如果没有特定配置，使用默认配置
	return &TenantPoolConfig{
		TenantID:         tenantID,
		MaxIdleConnCount: 5,
		MaxOpenConnCount: 20,
		MaxConnLifeTime:  300,
		DatabaseName:     "mer_system",
		ReadOnly:         false,
	}, nil
}

// createTenantPool 创建租户特定的连接池
func (m *TenantPoolManager) createTenantPool(ctx context.Context, config *TenantPoolConfig) (gdb.DB, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 双重检查，防止并发创建
	if db, exists := m.pools[config.TenantID]; exists {
		return db, nil
	}
	
	// 获取基础数据库配置
	baseConfig := GetDefaultDatabaseConfig()
	
	// 创建租户特定的配置
	configName := fmt.Sprintf("tenant_%d", config.TenantID)
	gdb.SetConfig(gdb.Config{
		configName: gdb.ConfigGroup{
			gdb.ConfigNode{
				Host:     baseConfig.Host,
				Port:     baseConfig.Port,
				User:     baseConfig.User,
				Pass:     baseConfig.Pass,
				Name:     config.DatabaseName,
				Type:     baseConfig.Type,
				Role:     gdb.Role(baseConfig.Role),
				Debug:    baseConfig.Debug,
				Charset:  baseConfig.Charset,
				Timezone: baseConfig.Timezone,
				// 租户特定的连接池配置
				MaxIdleConnCount: config.MaxIdleConnCount,
				MaxOpenConnCount: config.MaxOpenConnCount,
				MaxConnLifeTime:  time.Second * time.Duration(config.MaxConnLifeTime),
			},
		},
	})
	
	// 创建数据库实例
	db := g.DB(configName)
	
	// 测试连接
	if err := db.PingMaster(); err != nil {
		g.Log().Errorf(ctx, "租户 %d 数据库连接失败: %v", config.TenantID, err)
		return m.defaultDB, err
	}
	
	// 存储连接池
	m.pools[config.TenantID] = db
	
	g.Log().Infof(ctx, "租户 %d 连接池创建成功", config.TenantID)
	return db, nil
}

// UpdateTenantConfig 更新租户配置
func (m *TenantPoolManager) UpdateTenantConfig(ctx context.Context, config *TenantPoolConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 更新配置
	m.configs[config.TenantID] = config
	
	// 如果已有连接池，需要重新创建
	if _, exists := m.pools[config.TenantID]; exists {
		delete(m.pools, config.TenantID)
		g.Log().Infof(ctx, "租户 %d 连接池配置已更新，将在下次使用时重新创建", config.TenantID)
	}
	
	return nil
}

// GetTenantStats 获取租户连接池统计信息
func (m *TenantPoolManager) GetTenantStats(tenantID uint64) map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	
	if db, exists := m.pools[tenantID]; exists {
		// 这里可以获取连接池的统计信息
		// 由于gf框架没有直接暴露这些信息，我们记录基本状态
		stats["tenant_id"] = tenantID
		stats["pool_exists"] = true
		stats["config"] = m.configs[tenantID]
		
		// 尝试ping来检查连接状态
		err := db.PingMaster()
		stats["connection_healthy"] = err == nil
		if err != nil {
			stats["ping_error"] = err.Error()
		}
	} else {
		stats["tenant_id"] = tenantID
		stats["pool_exists"] = false
		stats["config"] = m.configs[tenantID]
	}
	
	return stats
}

// GetAllTenantsStats 获取所有租户连接池统计信息
func (m *TenantPoolManager) GetAllTenantsStats() map[uint64]map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	allStats := make(map[uint64]map[string]interface{})
	
	// 获取所有已配置租户的统计信息
	for tenantID := range m.configs {
		allStats[tenantID] = m.GetTenantStats(tenantID)
	}
	
	// 获取所有已创建连接池但没有配置的租户
	for tenantID := range m.pools {
		if _, exists := allStats[tenantID]; !exists {
			allStats[tenantID] = m.GetTenantStats(tenantID)
		}
	}
	
	return allStats
}

// CloseTenantPool 关闭租户连接池
func (m *TenantPoolManager) CloseTenantPool(ctx context.Context, tenantID uint64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.pools[tenantID]; exists {
		// GoFrame的DB实例会自动管理连接池
		// 我们只需要从管理器中移除即可
		delete(m.pools, tenantID)
		
		g.Log().Infof(ctx, "租户 %d 连接池已关闭", tenantID)
	}
	
	return nil
}

// CloseAllPools 关闭所有租户连接池
func (m *TenantPoolManager) CloseAllPools(ctx context.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	for tenantID := range m.pools {
		g.Log().Infof(ctx, "租户 %d 连接池已关闭", tenantID)
	}
	
	m.pools = make(map[uint64]gdb.DB)
}
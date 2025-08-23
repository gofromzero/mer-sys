package config

import (
	"fmt"
	"os"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TestDatabaseConfig 测试数据库配置
type TestDatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Charset  string
}

// InitTestDatabase 初始化测试数据库连接
func InitTestDatabase() error {
	// 从环境变量或默认值获取测试数据库配置
	config := TestDatabaseConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "3306"),
		User:     getEnvOrDefault("TEST_DB_USER", "root"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", ""),
		Database: getEnvOrDefault("TEST_DB_NAME", "mer_test"),
		Charset:  "utf8mb4",
	}

	// 构建数据源名称 (DSN)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
	)

	// 配置测试数据库连接
	gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			gdb.ConfigNode{
				Type: "mysql",
				Link: dsn,
			},
		},
	})

	// 测试数据库连接
	db := g.DB()
	if db == nil {
		return fmt.Errorf("无法创建数据库连接")
	}

	// 检查连接是否可用
	if err := db.PingMaster(); err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	return nil
}

// SetupTestTables 创建测试需要的数据表
func SetupTestTables() error {
	db := g.DB()
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 创建测试表的SQL（简化版本，用于测试）
	testTables := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			tenant_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(255) NOT NULL,
			price_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
			price_currency VARCHAR(3) NOT NULL DEFAULT 'CNY',
			rights_cost DECIMAL(10,2) DEFAULT 0,
			status VARCHAR(20) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_tenant (tenant_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS product_pricing_rules (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			tenant_id BIGINT UNSIGNED NOT NULL,
			product_id BIGINT UNSIGNED NOT NULL,
			rule_type VARCHAR(50) NOT NULL,
			rule_config JSON NOT NULL,
			priority INT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			valid_from TIMESTAMP NOT NULL,
			valid_until TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_tenant_product (tenant_id, product_id),
			INDEX idx_active_valid (is_active, valid_from, valid_until)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS product_rights_rules (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			tenant_id BIGINT UNSIGNED NOT NULL,
			product_id BIGINT UNSIGNED NOT NULL,
			rule_type VARCHAR(50) NOT NULL,
			consumption_rate DECIMAL(10,4) NOT NULL,
			min_rights_required DECIMAL(10,2) DEFAULT 0,
			insufficient_rights_action VARCHAR(50) NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_tenant_product (tenant_id, product_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS price_histories (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			tenant_id BIGINT UNSIGNED NOT NULL,
			product_id BIGINT UNSIGNED NOT NULL,
			old_price JSON NOT NULL,
			new_price JSON NOT NULL,
			change_reason VARCHAR(255) NOT NULL,
			changed_by BIGINT UNSIGNED NOT NULL,
			effective_date TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_tenant_product (tenant_id, product_id),
			INDEX idx_effective_date (effective_date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	// 执行创建表的SQL
	for _, sql := range testTables {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("创建测试表失败: %w", err)
		}
	}

	return nil
}

// CleanupTestTables 清理测试数据
func CleanupTestTables() error {
	db := g.DB()
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 清理测试数据的表
	tables := []string{
		"price_histories",
		"product_rights_rules", 
		"product_pricing_rules",
		"products",
	}

	for _, table := range tables {
		if _, err := db.Exec("DELETE FROM " + table + " WHERE tenant_id >= 9000"); err != nil {
			// 仅记录警告，不中断清理过程
			g.Log().Warning(nil, fmt.Sprintf("清理表 %s 失败: %v", table, err))
		}
	}

	return nil
}

// InsertTestData 插入测试数据
func InsertTestData() error {
	db := g.DB()
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 插入测试商品
	testProducts := []map[string]interface{}{
		{
			"tenant_id":      9001,
			"name":           "测试商品1",
			"price_amount":   100.00,
			"price_currency": "CNY",
			"rights_cost":    5.0,
			"status":         "active",
		},
		{
			"tenant_id":      9001,
			"name":           "测试商品2", 
			"price_amount":   200.00,
			"price_currency": "CNY",
			"rights_cost":    10.0,
			"status":         "active",
		},
	}

	for _, product := range testProducts {
		if _, err := db.Insert("products", product); err != nil {
			return fmt.Errorf("插入测试商品失败: %w", err)
		}
	}

	return nil
}

// getEnvOrDefault 获取环境变量或返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
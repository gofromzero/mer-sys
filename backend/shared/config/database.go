package config

import (
	"context"
	"time"
	
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"     yaml:"host"`
	Port     string `json:"port"     yaml:"port"`
	User     string `json:"user"     yaml:"user"`
	Pass     string `json:"pass"     yaml:"pass"`
	Name     string `json:"name"     yaml:"name"`
	Type     string `json:"type"     yaml:"type"`
	Role     string `json:"role"     yaml:"role"`
	Debug    bool   `json:"debug"    yaml:"debug"`
	Charset  string `json:"charset"  yaml:"charset"`
	Timezone string `json:"timezone" yaml:"timezone"`
}

// GetDefaultDatabaseConfig 获取默认数据库配置
func GetDefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     g.Cfg().MustGet(context.Background(), "database.default.host", "127.0.0.1").String(),
		Port:     g.Cfg().MustGet(context.Background(), "database.default.port", "3306").String(),
		User:     g.Cfg().MustGet(context.Background(), "database.default.user", "root").String(),
		Pass:     g.Cfg().MustGet(context.Background(), "database.default.pass", "").String(),
		Name:     g.Cfg().MustGet(context.Background(), "database.default.name", "mer_system").String(),
		Type:     g.Cfg().MustGet(context.Background(), "database.default.type", "mysql").String(),
		Role:     g.Cfg().MustGet(context.Background(), "database.default.role", "master").String(),
		Debug:    g.Cfg().MustGet(context.Background(), "database.default.debug", false).Bool(),
		Charset:  g.Cfg().MustGet(context.Background(), "database.default.charset", "utf8mb4").String(),
		Timezone: g.Cfg().MustGet(context.Background(), "database.default.timezone", "Local").String(),
	}
}

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	config := GetDefaultDatabaseConfig()
	
	// 配置数据库连接
	gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			gdb.ConfigNode{
				Host:     config.Host,
				Port:     config.Port,
				User:     config.User,
				Pass:     config.Pass,
				Name:     config.Name,
				Type:     config.Type,
				// Role字段需要转换为gdb.Role类型
				Role:     gdb.Role(config.Role),
				Debug:    config.Debug,
				Charset:  config.Charset,
				Timezone: config.Timezone,
				// 连接池配置
				MaxIdleConnCount: 10,
				MaxOpenConnCount: 100,
				MaxConnLifeTime:  time.Second * 30,
			},
		},
	})
	
	// 测试数据库连接
	ctx := context.Background()
	db := g.DB()
	
	if err := db.PingMaster(); err != nil {
		g.Log().Errorf(ctx, "数据库连接失败: %v", err)
		return err
	}
	
	g.Log().Info(ctx, "数据库连接成功")
	return nil
}

// GetDB 获取数据库实例
func GetDB() gdb.DB {
	return g.DB()
}
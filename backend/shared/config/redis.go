package config

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Address  string `json:"address"  yaml:"address"`
	DB       int    `json:"db"       yaml:"db"`
	Pass     string `json:"pass"     yaml:"pass"`
	PoolSize int    `json:"poolSize" yaml:"poolSize"`
}

// GetDefaultRedisConfig 获取默认Redis配置
func GetDefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Address:  g.Cfg().MustGet(context.Background(), "redis.default.address", "127.0.0.1:6379").String(),
		DB:       g.Cfg().MustGet(context.Background(), "redis.default.db", 0).Int(),
		Pass:     g.Cfg().MustGet(context.Background(), "redis.default.pass", "").String(),
		PoolSize: g.Cfg().MustGet(context.Background(), "redis.default.poolSize", 10).Int(),
	}
}

// InitRedis 初始化Redis连接
func InitRedis() error {
	config := GetDefaultRedisConfig()
	
	// 配置Redis连接
	redisConfig := &gredis.Config{
		Address: config.Address,
		Db:      config.DB,
		Pass:    config.Pass,
		MaxIdle: config.PoolSize,
	}
	
	gredis.SetConfig(redisConfig)
	
	// 测试Redis连接
	ctx := context.Background()
	redis := g.Redis()
	
	// 执行PING命令测试连接
	_, err := redis.Do(ctx, "PING")
	if err != nil {
		g.Log().Errorf(ctx, "Redis连接失败: %v", err)
		return err
	}
	
	g.Log().Info(ctx, "Redis连接成功")
	return nil
}

// GetRedis 获取Redis实例
func GetRedis() *gredis.Redis {
	return g.Redis()
}

// RedisCache Redis缓存接口
type RedisCache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	GetString(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	FlushDB(ctx context.Context) error
}
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/config"
	"github.com/gogf/gf/v2/util/gconv"
)

// Cache 缓存管理器
type Cache struct {
	prefix string
}

// NewCache 创建缓存实例
func NewCache(prefix string) *Cache {
	return &Cache{
		prefix: prefix,
	}
}

// buildKey 构建带前缀的键
func (c *Cache) buildKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// Set 设置缓存值
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	// 序列化值
	var data interface{}
	if value != nil {
		switch v := value.(type) {
		case string, int, int64, float64, bool:
			data = v
		default:
			// 复杂类型使用JSON序列化
			jsonData, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("序列化缓存值失败: %v", err)
			}
			data = string(jsonData)
		}
	}

	if ttl > 0 {
		_, err := redis.Do(ctx, "SETEX", fullKey, int64(ttl.Seconds()), data)
		return err
	}

	_, err := redis.Do(ctx, "SET", fullKey, data)
	return err
}

// Get 获取缓存值
func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	result, err := redis.Do(ctx, "GET", fullKey)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetString 获取字符串缓存值
func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if value == nil {
		return "", nil
	}

	return gconv.String(value), nil
}

// GetInt 获取整数缓存值
func (c *Cache) GetInt(ctx context.Context, key string) (int, error) {
	value, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return 0, nil
	}

	return gconv.Int(value), nil
}

// GetStruct 获取结构体缓存值
func (c *Cache) GetStruct(ctx context.Context, key string, dst interface{}) error {
	value, err := c.GetString(ctx, key)
	if err != nil {
		return err
	}

	if value == "" {
		return fmt.Errorf("缓存值为空")
	}

	return json.Unmarshal([]byte(value), dst)
}

// Delete 删除缓存
func (c *Cache) Delete(ctx context.Context, key string) error {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	_, err := redis.Do(ctx, "DEL", fullKey)
	return err
}

// Exists 检查缓存是否存在
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	result, err := redis.Do(ctx, "EXISTS", fullKey)
	if err != nil {
		return false, err
	}

	return gconv.Int(result) > 0, nil
}

// Expire 设置缓存过期时间
func (c *Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	_, err := redis.Do(ctx, "EXPIRE", fullKey, int64(ttl.Seconds()))
	return err
}

// TTL 获取缓存剩余生存时间
func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	result, err := redis.Do(ctx, "TTL", fullKey)
	if err != nil {
		return 0, err
	}

	seconds := gconv.Int64(result)
	if seconds < 0 {
		return -1, nil // -1表示永不过期，-2表示不存在
	}

	return time.Duration(seconds) * time.Second, nil
}

// Keys 获取匹配的键列表
func (c *Cache) Keys(ctx context.Context, pattern string) ([]string, error) {
	redis := config.GetRedis()
	fullPattern := c.buildKey(pattern)

	result, err := redis.Do(ctx, "KEYS", fullPattern)
	if err != nil {
		return nil, err
	}

	return gconv.Strings(result), nil
}

// FlushDB 清空当前数据库
func (c *Cache) FlushDB(ctx context.Context) error {
	redis := config.GetRedis()
	_, err := redis.Do(ctx, "FLUSHDB")
	return err
}

// Increment 递增计数器
func (c *Cache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	result, err := redis.Do(ctx, "INCRBY", fullKey, delta)
	if err != nil {
		return 0, err
	}

	return gconv.Int64(result), nil
}

// Decrement 递减计数器
func (c *Cache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return c.Increment(ctx, key, -delta)
}

// HSet 设置哈希字段值
func (c *Cache) HSet(ctx context.Context, key, field string, value interface{}) error {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	_, err := redis.Do(ctx, "HSET", fullKey, field, value)
	return err
}

// HGet 获取哈希字段值
func (c *Cache) HGet(ctx context.Context, key, field string) (string, error) {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	result, err := redis.Do(ctx, "HGET", fullKey, field)
	if err != nil {
		return "", err
	}

	return gconv.String(result), nil
}

// HGetAll 获取哈希所有字段
func (c *Cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	result, err := redis.Do(ctx, "HGETALL", fullKey)
	if err != nil {
		return nil, err
	}

	data := gconv.SliceStr(result)
	hashMap := make(map[string]string)

	for i := 0; i < len(data); i += 2 {
		if i+1 < len(data) {
			hashMap[data[i]] = data[i+1]
		}
	}

	return hashMap, nil
}

// HDel 删除哈希字段
func (c *Cache) HDel(ctx context.Context, key string, fields ...string) error {
	redis := config.GetRedis()
	fullKey := c.buildKey(key)

	args := []interface{}{fullKey}
	for _, field := range fields {
		args = append(args, field)
	}

	_, err := redis.Do(ctx, "HDEL", args...)
	return err
}

package test

import (
	"context"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/auth"
	"github.com/gofromzero/mer-sys/backend/shared/cache"
	"github.com/gofromzero/mer-sys/backend/shared/config"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/test/gtest"
)

// TestRedisConnection 测试Redis连接
func TestRedisConnection(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化Redis连接
		err := config.InitRedis()
		t.AssertNil(err)

		// 测试基本连接
		redis := config.GetRedis()
		t.AssertNE(redis, nil)

		ctx := context.Background()

		// 测试PING命令
		result, err := redis.Do(ctx, "PING")
		t.AssertNil(err)
		t.AssertEQ(result, "PONG")
	})
}

// TestCacheOperations 测试缓存基础操作
func TestCacheOperations(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化Redis连接
		err := config.InitRedis()
		t.AssertNil(err)

		cache := cache.NewCache("test")
		ctx := context.Background()

		// 测试字符串操作
		testKey := "test_string"
		testValue := "hello world"

		// 设置值
		err = cache.Set(ctx, testKey, testValue, time.Minute)
		t.AssertNil(err)

		// 获取值
		value, err := cache.GetString(ctx, testKey)
		t.AssertNil(err)
		t.AssertEQ(value, testValue)

		// 检查存在性
		exists, err := cache.Exists(ctx, testKey)
		t.AssertNil(err)
		t.AssertEQ(exists, true)

		// 测试TTL
		ttl, err := cache.TTL(ctx, testKey)
		t.AssertNil(err)
		t.AssertGT(ttl.Seconds(), 0)

		// 删除值
		err = cache.Delete(ctx, testKey)
		t.AssertNil(err)

		// 验证删除
		exists, err = cache.Exists(ctx, testKey)
		t.AssertNil(err)
		t.AssertEQ(exists, false)
	})
}

// TestCacheComplexData 测试复杂数据缓存
func TestCacheComplexData(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化Redis连接
		err := config.InitRedis()
		t.AssertNil(err)

		cache := cache.NewCache("test")
		ctx := context.Background()

		// 测试结构体缓存
		testUser := types.User{
			ID:       123,
			UUID:     "test-uuid-123",
			Username: "testuser",
			Email:    "test@example.com",
			TenantID: 1,
			Status:   types.UserStatusActive,
		}

		testKey := "test_user"

		// 存储结构体
		err = cache.Set(ctx, testKey, testUser, time.Minute)
		t.AssertNil(err)

		// 获取结构体
		var retrievedUser types.User
		err = cache.GetStruct(ctx, testKey, &retrievedUser)
		t.AssertNil(err)

		// 验证数据
		t.AssertEQ(retrievedUser.ID, testUser.ID)
		t.AssertEQ(retrievedUser.UUID, testUser.UUID)
		t.AssertEQ(retrievedUser.Username, testUser.Username)
		t.AssertEQ(retrievedUser.Email, testUser.Email)
		t.AssertEQ(retrievedUser.TenantID, testUser.TenantID)
		t.AssertEQ(retrievedUser.Status, testUser.Status)

		// 清理
		cache.Delete(ctx, testKey)
	})
}

// TestCacheHashOperations 测试哈希操作
func TestCacheHashOperations(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化Redis连接
		err := config.InitRedis()
		t.AssertNil(err)

		cache := cache.NewCache("test")
		ctx := context.Background()

		testKey := "test_hash"

		// 设置哈希字段
		err = cache.HSet(ctx, testKey, "field1", "value1")
		t.AssertNil(err)

		err = cache.HSet(ctx, testKey, "field2", "value2")
		t.AssertNil(err)

		// 获取单个字段
		value, err := cache.HGet(ctx, testKey, "field1")
		t.AssertNil(err)
		t.AssertEQ(value, "value1")

		// 获取所有字段
		allFields, err := cache.HGetAll(ctx, testKey)
		t.AssertNil(err)
		t.AssertEQ(len(allFields), 2)
		t.AssertEQ(allFields["field1"], "value1")
		t.AssertEQ(allFields["field2"], "value2")

		// 删除字段
		err = cache.HDel(ctx, testKey, "field1")
		t.AssertNil(err)

		// 验证删除
		allFields, err = cache.HGetAll(ctx, testKey)
		t.AssertNil(err)
		t.AssertEQ(len(allFields), 1)
		t.AssertEQ(allFields["field2"], "value2")

		// 清理
		cache.Delete(ctx, testKey)
	})
}

// TestCacheCounterOperations 测试计数器操作
func TestCacheCounterOperations(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化Redis连接
		err := config.InitRedis()
		t.AssertNil(err)

		cache := cache.NewCache("test")
		ctx := context.Background()

		testKey := "test_counter"

		// 递增计数器
		count, err := cache.Increment(ctx, testKey, 1)
		t.AssertNil(err)
		t.AssertEQ(count, int64(1))

		count, err = cache.Increment(ctx, testKey, 5)
		t.AssertNil(err)
		t.AssertEQ(count, int64(6))

		// 递减计数器
		count, err = cache.Decrement(ctx, testKey, 2)
		t.AssertNil(err)
		t.AssertEQ(count, int64(4))

		// 清理
		cache.Delete(ctx, testKey)
	})
}

// TestJWTTokenStorage 测试JWT令牌存储
func TestJWTTokenStorage(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化Redis连接
		err := config.InitRedis()
		t.AssertNil(err)

		jwtManager := auth.NewJWTManager()
		ctx := context.Background()

		// 创建测试用户
		testUser := &types.User{
			ID:       123,
			UUID:     "test-uuid-123",
			Username: "testuser",
			Email:    "test@example.com",
			TenantID: 1,
			Status:   types.UserStatusActive,
		}

		// 生成访问令牌
		accessToken, err := jwtManager.GenerateToken(ctx, testUser, "access")
		t.AssertNil(err)
		t.AssertNE(accessToken, "")

		// 生成刷新令牌
		refreshToken, err := jwtManager.GenerateToken(ctx, testUser, "refresh")
		t.AssertNil(err)
		t.AssertNE(refreshToken, "")

		// 验证访问令牌
		claims, err := jwtManager.ValidateToken(ctx, accessToken)
		t.AssertNil(err)
		t.AssertEQ(claims.UserID, testUser.ID)
		t.AssertEQ(claims.TenantID, testUser.TenantID)
		t.AssertEQ(claims.Username, testUser.Username)
		t.AssertEQ(claims.TokenType, "access")

		// 验证刷新令牌
		claims, err = jwtManager.ValidateToken(ctx, refreshToken)
		t.AssertNil(err)
		t.AssertEQ(claims.TokenType, "refresh")

		// 使用刷新令牌生成新的访问令牌
		newAccessToken, err := jwtManager.RefreshToken(ctx, refreshToken)
		t.AssertNil(err)
		t.AssertNE(newAccessToken, "")
		t.AssertNE(newAccessToken, accessToken)

		// 撤销令牌
		err = jwtManager.RevokeToken(ctx, accessToken)
		t.AssertNil(err)

		// 验证撤销后的令牌无效
		_, err = jwtManager.ValidateToken(ctx, accessToken)
		t.AssertNE(err, nil)

		// 撤销用户所有令牌
		err = jwtManager.RevokeUserTokens(ctx, testUser.ID)
		t.AssertNil(err)

		// 验证所有令牌都无效
		_, err = jwtManager.ValidateToken(ctx, refreshToken)
		t.AssertNE(err, nil)

		_, err = jwtManager.ValidateToken(ctx, newAccessToken)
		t.AssertNE(err, nil)
	})
}

// TestRedisPatternMatching 测试Redis模式匹配
func TestRedisPatternMatching(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化Redis连接
		err := config.InitRedis()
		t.AssertNil(err)

		cache := cache.NewCache("test")
		ctx := context.Background()

		// 设置多个测试键
		testKeys := []string{"test:user:1", "test:user:2", "test:order:1", "test:order:2"}
		for _, key := range testKeys {
			err = cache.Set(ctx, key, "value", time.Minute)
			t.AssertNil(err)
		}

		// 测试模式匹配
		userKeys, err := cache.Keys(ctx, "user:*")
		t.AssertNil(err)
		t.AssertEQ(len(userKeys), 2)

		orderKeys, err := cache.Keys(ctx, "order:*")
		t.AssertNil(err)
		t.AssertEQ(len(orderKeys), 2)

		allKeys, err := cache.Keys(ctx, "*")
		t.AssertNil(err)
		t.AssertGE(len(allKeys), 4) // 至少包含我们创建的4个键

		// 清理测试数据
		for _, key := range testKeys {
			cache.Delete(ctx, key)
		}
	})
}

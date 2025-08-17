package test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gofromzero/mer-sys/backend/shared/config"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// TestTenantIsolation 测试租户隔离机制
func TestTenantIsolation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化数据库连接
		err := config.InitDatabase()
		t.AssertNil(err)

		// 创建测试用户仓储
		userRepo := repository.NewUserRepository()

		// 创建两个不同的租户上下文
		ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
		ctx2 := context.WithValue(context.Background(), "tenant_id", uint64(2))

		// 在租户1下创建用户
		user1 := &types.User{
			UUID:     "test-user-1",
			Username: "testuser1",
			Email:    "test1@example.com",
			Status:   types.UserStatusActive,
		}
		err = userRepo.Create(ctx1, user1)
		t.AssertNil(err)

		// 在租户2下创建用户
		user2 := &types.User{
			UUID:     "test-user-2",
			Username: "testuser2",
			Email:    "test2@example.com",
			Status:   types.UserStatusActive,
		}
		err = userRepo.Create(ctx2, user2)
		t.AssertNil(err)

		// 验证租户隔离：租户1只能看到自己的用户
		users1, err := userRepo.FindAllByTenant(ctx1)
		t.AssertNil(err)

		// 检查租户1的用户数据中是否包含testuser1但不包含testuser2
		found1 := false
		found2 := false
		for _, user := range users1 {
			if user.Username == "testuser1" {
				found1 = true
				t.AssertEQ(user.TenantID, uint64(1))
			}
			if user.Username == "testuser2" {
				found2 = true
			}
		}
		t.AssertEQ(found1, true)
		t.AssertEQ(found2, false)

		// 验证租户隔离：租户2只能看到自己的用户
		users2, err := userRepo.FindAllByTenant(ctx2)
		t.AssertNil(err)

		found1 = false
		found2 = false
		for _, user := range users2 {
			if user.Username == "testuser1" {
				found1 = true
			}
			if user.Username == "testuser2" {
				found2 = true
				t.AssertEQ(user.TenantID, uint64(2))
			}
		}
		t.AssertEQ(found1, false)
		t.AssertEQ(found2, true)

		// 验证跨租户查询：租户1无法通过ID查询租户2的用户
		user2Found, err := userRepo.FindByUsername(ctx1, "testuser2")
		t.AssertNE(err, nil)
		t.AssertNil(user2Found)

		// 验证更新操作的租户隔离
		err = userRepo.UpdateStatus(ctx1, user2.ID, types.UserStatusSuspended)
		t.AssertNE(err, nil)

		// 验证删除操作的租户隔离
		err = userRepo.DeleteByID(ctx1, user2.ID)
		t.AssertNE(err, nil)

		// 清理测试数据
		userRepo.DeleteByID(ctx1, user1.ID)
		userRepo.DeleteByID(ctx2, user2.ID)
	})
}

// TestTenantContextMissing 测试缺少租户上下文的情况
func TestTenantContextMissing(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化数据库连接
		err := config.InitDatabase()
		t.AssertNil(err)

		userRepo := repository.NewUserRepository()

		// 在没有租户上下文的情况下尝试操作
		ctx := context.Background()

		_, err = userRepo.FindAllByTenant(ctx)
		t.AssertNE(err, nil)

		user := &types.User{
			UUID:     "test-no-tenant",
			Username: "notenant",
			Email:    "notenant@example.com",
			Status:   types.UserStatusActive,
		}
		err = userRepo.Create(ctx, user)
		t.AssertNE(err, nil)
	})
}

// TestRepositoryTenantValidation 测试仓储层租户验证
func TestRepositoryTenantValidation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化数据库连接
		err := config.InitDatabase()
		t.AssertNil(err)

		baseRepo := repository.NewBaseRepository("users")

		// 测试获取租户ID
		ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
		tenantID, err := baseRepo.GetTenantID(ctx1)
		t.AssertNil(err)
		t.AssertEQ(tenantID, uint64(1))

		// 测试不同类型的租户ID
		ctx2 := context.WithValue(context.Background(), "tenant_id", int(2))
		tenantID, err = baseRepo.GetTenantID(ctx2)
		t.AssertNil(err)
		t.AssertEQ(tenantID, uint64(2))

		ctx3 := context.WithValue(context.Background(), "tenant_id", "3")
		tenantID, err = baseRepo.GetTenantID(ctx3)
		t.AssertNil(err)
		t.AssertEQ(tenantID, uint64(3))

		// 测试无效租户ID类型
		ctx4 := context.WithValue(context.Background(), "tenant_id", []string{"invalid"})
		_, err = baseRepo.GetTenantID(ctx4)
		t.AssertNE(err, nil)

		// 测试缺少租户ID
		ctx5 := context.Background()
		_, err = baseRepo.GetTenantID(ctx5)
		t.AssertNE(err, nil)
	})
}

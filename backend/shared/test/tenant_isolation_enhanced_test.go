package test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gofromzero/mer-sys/backend/shared/config"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// TestEnhancedTenantIsolation 增强的多租户隔离测试
func TestEnhancedTenantIsolation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化数据库连接
		err := config.InitDatabase()
		if err != nil {
			t.Logf("数据库连接失败，跳过测试: %v", err)
			return
		}

		// 创建仓储实例
		userRepo := repository.NewUserRepository()
		merchantRepo := repository.NewMerchantRepository()

		// 创建测试用的租户上下文
		ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
		ctx2 := context.WithValue(context.Background(), "tenant_id", uint64(2))

		// 测试用户租户隔离
		testUserTenantIsolation(t, userRepo, ctx1, ctx2)

		// 测试商户租户隔离
		testMerchantTenantIsolation(t, merchantRepo, ctx1, ctx2)

		// 测试跨租户访问拒绝
		testCrossTenantAccessDenial(t, userRepo, merchantRepo, ctx1, ctx2)
	})
}

// testUserTenantIsolation 测试用户租户隔离
func testUserTenantIsolation(t *gtest.T, userRepo *repository.UserRepository, ctx1, ctx2 context.Context) {
	// 在租户1创建用户
	user1 := &types.User{
		UUID:     "test-tenant1-user",
		Username: "tenant1user",
		Email:    "tenant1@example.com",
		Status:   types.UserStatusActive,
	}
	err := userRepo.Create(ctx1, user1)
	if err != nil {
		t.Logf("创建租户1用户失败: %v", err)
	}

	// 在租户2创建用户
	user2 := &types.User{
		UUID:     "test-tenant2-user",
		Username: "tenant2user",
		Email:    "tenant2@example.com",
		Status:   types.UserStatusActive,
	}
	err = userRepo.Create(ctx2, user2)
	if err != nil {
		t.Logf("创建租户2用户失败: %v", err)
	}

	// 验证租户1只能看到自己的用户
	users1, err := userRepo.FindAllByTenant(ctx1)
	if err != nil {
		t.Logf("查询租户1用户失败: %v", err)
		return
	}

	found1InTenant1 := false
	found2InTenant1 := false
	for _, user := range users1 {
		if user.Username == "tenant1user" {
			found1InTenant1 = true
			t.AssertEQ(user.TenantID, uint64(1))
		}
		if user.Username == "tenant2user" {
			found2InTenant1 = true
		}
	}
	t.AssertEQ(found1InTenant1, true)
	t.AssertEQ(found2InTenant1, false)

	// 验证租户2只能看到自己的用户
	users2, err := userRepo.FindAllByTenant(ctx2)
	if err != nil {
		t.Logf("查询租户2用户失败: %v", err)
		return
	}

	found1InTenant2 := false
	found2InTenant2 := false
	for _, user := range users2 {
		if user.Username == "tenant1user" {
			found1InTenant2 = true
		}
		if user.Username == "tenant2user" {
			found2InTenant2 = true
			t.AssertEQ(user.TenantID, uint64(2))
		}
	}
	t.AssertEQ(found1InTenant2, false)
	t.AssertEQ(found2InTenant2, true)

	// 清理测试数据
	if user1.ID > 0 {
		userRepo.DeleteByID(ctx1, user1.ID)
	}
	if user2.ID > 0 {
		userRepo.DeleteByID(ctx2, user2.ID)
	}
}

// testMerchantTenantIsolation 测试商户租户隔离
func testMerchantTenantIsolation(t *gtest.T, merchantRepo repository.MerchantRepository, ctx1, ctx2 context.Context) {
	// 在租户1创建商户
	merchant1 := &types.Merchant{
		Name:   "租户1测试商户",
		Code:   "tenant1-merchant",
		Status: types.MerchantStatusActive,
		BusinessInfo: &types.BusinessInfo{
			Type:     "retail",
			Category: "general",
		},
		RightsBalance: &types.RightsBalance{
			TotalBalance:  1000.0,
			UsedBalance:   0.0,
			FrozenBalance: 0.0,
		},
	}
	err := merchantRepo.Create(ctx1, merchant1)
	if err != nil {
		t.Logf("创建租户1商户失败: %v", err)
	}

	// 在租户2创建商户
	merchant2 := &types.Merchant{
		Name:   "租户2测试商户",
		Code:   "tenant2-merchant",
		Status: types.MerchantStatusActive,
		BusinessInfo: &types.BusinessInfo{
			Type:     "service",
			Category: "test",
		},
		RightsBalance: &types.RightsBalance{
			TotalBalance:  500.0,
			UsedBalance:   0.0,
			FrozenBalance: 0.0,
		},
	}
	err = merchantRepo.Create(ctx2, merchant2)
	if err != nil {
		t.Logf("创建租户2商户失败: %v", err)
	}

	// 验证租户1只能看到自己的商户
	merchants1, err := merchantRepo.GetByTenantID(ctx1, 1)
	if err != nil {
		t.Logf("查询租户1商户失败: %v", err)
	} else {
		found1InTenant1 := false
		found2InTenant1 := false
		for _, merchant := range merchants1 {
			if merchant.Code == "tenant1-merchant" {
				found1InTenant1 = true
				t.AssertEQ(merchant.TenantID, uint64(1))
			}
			if merchant.Code == "tenant2-merchant" {
				found2InTenant1 = true
			}
		}
		t.AssertEQ(found1InTenant1, true)
		t.AssertEQ(found2InTenant1, false)
	}

	// 验证租户2只能看到自己的商户
	merchants2, err := merchantRepo.GetByTenantID(ctx2, 2)
	if err != nil {
		t.Logf("查询租户2商户失败: %v", err)
	} else {
		found1InTenant2 := false
		found2InTenant2 := false
		for _, merchant := range merchants2 {
			if merchant.Code == "tenant1-merchant" {
				found1InTenant2 = true
			}
			if merchant.Code == "tenant2-merchant" {
				found2InTenant2 = true
				t.AssertEQ(merchant.TenantID, uint64(2))
			}
		}
		t.AssertEQ(found1InTenant2, false)
		t.AssertEQ(found2InTenant2, true)
	}

	// 清理测试数据
	if merchant1.ID > 0 {
		merchantRepo.Delete(ctx1, merchant1.ID)
	}
	if merchant2.ID > 0 {
		merchantRepo.Delete(ctx2, merchant2.ID)
	}
}

// testCrossTenantAccessDenial 测试跨租户访问拒绝
func testCrossTenantAccessDenial(t *gtest.T, userRepo *repository.UserRepository, merchantRepo repository.MerchantRepository, ctx1, ctx2 context.Context) {
	// 测试租户1尝试访问租户2的数据应该被拒绝
	_, err := merchantRepo.GetByTenantID(ctx1, 2)
	t.AssertNE(err, nil)
	t.AssertEQ(err, types.ErrCrossTenantAccess)

	// 测试租户2尝试访问租户1的数据应该被拒绝
	_, err = merchantRepo.GetByTenantID(ctx2, 1)
	t.AssertNE(err, nil)
	t.AssertEQ(err, types.ErrCrossTenantAccess)
}

// TestTenantContextValidation 测试租户上下文验证
func TestTenantContextValidation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化数据库连接
		err := config.InitDatabase()
		if err != nil {
			t.Logf("数据库连接失败，跳过测试: %v", err)
			return
		}

		baseRepo := repository.NewBaseRepository("users")

		// 测试有效的租户ID类型
		testCases := []struct {
			name     string
			value    interface{}
			expected uint64
		}{
			{"uint64类型", uint64(1), 1},
			{"int64类型", int64(2), 2},
			{"int类型", int(3), 3},
			{"string类型", "4", 4},
		}

		for _, tc := range testCases {
			ctx := context.WithValue(context.Background(), "tenant_id", tc.value)
			tenantID, err := baseRepo.GetTenantID(ctx)
			t.AssertNil(err)
			t.AssertEQ(tenantID, tc.expected)
		}

		// 测试无效的租户ID类型
		invalidCases := []struct {
			name  string
			value interface{}
		}{
			{"空值", nil},
			{"空字符串", ""},
			{"数组", []string{"invalid"}},
			{"浮点数", 3.14},
			{"布尔值", true},
		}

		for _, tc := range invalidCases {
			var ctx context.Context
			if tc.value == nil {
				ctx = context.Background()
			} else {
				ctx = context.WithValue(context.Background(), "tenant_id", tc.value)
			}
			_, err := baseRepo.GetTenantID(ctx)
			t.AssertNE(err, nil)
		}
	})
}

// TestRepositoryTenantAutoInjection 测试仓储层租户ID自动注入
func TestRepositoryTenantAutoInjection(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化数据库连接
		err := config.InitDatabase()
		if err != nil {
			t.Logf("数据库连接失败，跳过测试: %v", err)
			return
		}

		baseRepo := repository.NewBaseRepository("users")
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))

		// 测试Model方法自动注入租户ID
		model, err := baseRepo.Model(ctx)
		t.AssertNil(err)
		t.AssertNE(model, nil)

		// 验证查询条件中包含租户ID过滤
		sql, args := model.GetSql()
		t.AssertNE(sql, "")
		t.AssertEQ(len(args), 1)
		t.AssertEQ(args[0], uint64(1))
	})
}
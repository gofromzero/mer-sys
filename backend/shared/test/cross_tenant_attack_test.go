package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gofromzero/mer-sys/backend/shared/config"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// TestCrossTenantAttackSimulation 跨租户访问攻击模拟测试
func TestCrossTenantAttackSimulation(t *testing.T) {
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

		// 攻击场景1: 恶意用户尝试直接修改上下文中的tenant_id
		testDirectTenantIDModification(t, userRepo, merchantRepo)

		// 攻击场景2: 尝试通过不同的租户上下文访问其他租户的数据
		testCrossTenantDataAccess(t, userRepo, merchantRepo)

		// 攻击场景3: 尝试绕过Repository层直接使用数据库
		testRepositoryBypass(t)

		// 攻击场景4: 验证无租户上下文的访问被拒绝
		testNoTenantContextAccess(t, userRepo, merchantRepo)
	})
}

// testDirectTenantIDModification 测试直接修改租户ID的攻击
func testDirectTenantIDModification(t *gtest.T, userRepo *repository.UserRepository, merchantRepo repository.MerchantRepository) {
	// 创建租户1的上下文
	ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
	
	// 尝试创建商户
	merchant := &types.Merchant{
		Name:   "攻击测试商户",
		Code:   "attack-test",
		Status: types.MerchantStatusActive,
		BusinessInfo: &types.BusinessInfo{
			Type:     "attack",
			Category: "test",
		},
		RightsBalance: &types.RightsBalance{
			TotalBalance:  1000.0,
			UsedBalance:   0.0,
			FrozenBalance: 0.0,
		},
	}
	
	// 即使恶意用户尝试手动设置不同的tenant_id，系统也应该使用上下文中的租户ID
	merchant.TenantID = 999 // 恶意设置
	err := merchantRepo.Create(ctx1, merchant)
	
	if err != nil {
		t.Logf("创建商户失败: %v", err)
	} else {
		// 验证创建的商户实际上属于上下文中的租户(1)，而不是恶意设置的租户(999)
		t.AssertEQ(merchant.TenantID, uint64(1))
		
		// 清理测试数据
		merchantRepo.Delete(ctx1, merchant.ID)
	}
}

// testCrossTenantDataAccess 测试跨租户数据访问攻击
func testCrossTenantDataAccess(t *gtest.T, userRepo *repository.UserRepository, merchantRepo repository.MerchantRepository) {
	// 创建两个不同租户的上下文
	ctx1 := context.WithValue(context.Background(), "tenant_id", uint64(1))
	ctx2 := context.WithValue(context.Background(), "tenant_id", uint64(2))
	
	// 在租户1创建测试数据
	user1 := &types.User{
		UUID:     "attack-test-user1",
		Username: "attackuser1",
		Email:    "attack1@example.com",
		Status:   types.UserStatusActive,
	}
	err := userRepo.Create(ctx1, user1)
	if err != nil {
		t.Logf("创建租户1用户失败: %v", err)
		return
	}
	
	// 攻击尝试：租户2的用户尝试访问租户1的数据
	
	// 1. 尝试通过用户名查找租户1的用户
	attackUser, err := userRepo.FindByUsername(ctx2, "attackuser1")
	t.AssertNE(err, nil) // 应该失败
	t.AssertNil(attackUser) // 应该返回空
	
	// 2. 尝试通过ID访问租户1的用户
	if user1.ID > 0 {
		attackUser, err = userRepo.GetByID(ctx2, user1.ID)
		t.AssertNE(err, nil) // 应该失败
		t.AssertNil(attackUser) // 应该返回空
	}
	
	// 3. 尝试修改租户1的用户状态
	if user1.ID > 0 {
		err = userRepo.UpdateStatus(ctx2, user1.ID, types.UserStatusSuspended)
		t.AssertNE(err, nil) // 应该失败
	}
	
	// 4. 尝试删除租户1的用户
	if user1.ID > 0 {
		err = userRepo.DeleteByID(ctx2, user1.ID)
		t.AssertNE(err, nil) // 应该失败
	}
	
	// 5. 尝试直接请求访问租户1的商户数据
	_, err = merchantRepo.GetByTenantID(ctx2, 1)
	t.AssertEQ(err, types.ErrCrossTenantAccess) // 应该返回跨租户访问错误
	
	// 清理测试数据
	if user1.ID > 0 {
		userRepo.DeleteByID(ctx1, user1.ID)
	}
}

// testRepositoryBypass 测试绕过Repository层的攻击
func testRepositoryBypass(t *gtest.T) {
	baseRepo := repository.NewBaseRepository("users")
	
	// 尝试使用不带租户隔离的模型
	model := baseRepo.ModelWithoutTenant()
	t.AssertNE(model, nil)
	
	// 这个方法应该被严格控制使用
	// 在实际应用中，应该通过代码审查和权限控制来防止滥用
	
	// 验证正常的租户隔离模型需要有效的上下文
	ctx := context.Background() // 无租户上下文
	_, err := baseRepo.Model(ctx)
	t.AssertNE(err, nil) // 应该失败
}

// testNoTenantContextAccess 测试无租户上下文的访问
func testNoTenantContextAccess(t *gtest.T, userRepo *repository.UserRepository, merchantRepo repository.MerchantRepository) {
	// 创建无租户上下文
	ctx := context.Background()
	
	// 尝试各种操作，都应该失败
	
	// 1. 尝试查询用户
	_, err := userRepo.FindAllByTenant(ctx)
	t.AssertNE(err, nil)
	
	// 2. 尝试创建用户
	user := &types.User{
		UUID:     "no-tenant-user",
		Username: "notenant",
		Email:    "notenant@example.com",
		Status:   types.UserStatusActive,
	}
	err = userRepo.Create(ctx, user)
	t.AssertNE(err, nil)
	
	// 3. 尝试查询商户
	_, err = merchantRepo.GetByTenantID(ctx, 1)
	t.AssertNE(err, nil)
	
	// 4. 尝试创建商户
	merchant := &types.Merchant{
		Name:   "无租户商户",
		Code:   "no-tenant-merchant",
		Status: types.MerchantStatusActive,
	}
	err = merchantRepo.Create(ctx, merchant)
	t.AssertNE(err, nil)
}

// TestPerformanceImpactEvaluation 性能影响评估测试
func TestPerformanceImpactEvaluation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化数据库连接
		err := config.InitDatabase()
		if err != nil {
			t.Logf("数据库连接失败，跳过测试: %v", err)
			return
		}

		baseRepo := repository.NewBaseRepository("users")
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
		
		// 测试租户隔离查询的性能
		startTime := time.Now()
		
		// 执行100次租户隔离查询
		for i := 0; i < 100; i++ {
			model, err := baseRepo.Model(ctx)
			if err != nil {
				t.Logf("模型创建失败: %v", err)
				continue
			}
			
			// 执行计数查询
			_, err = model.Count()
			if err != nil {
				t.Logf("查询失败: %v", err)
			}
		}
		
		endTime := time.Now()
		duration := endTime.Sub(startTime).Milliseconds()
		
		t.Logf("100次租户隔离查询耗时: %d毫秒", duration)
		
		// 验证性能影响在可接受范围内 (每次查询平均不超过50ms)
		avgTime := duration / 100
		t.AssertLT(avgTime, int64(50))
	})
}

// TestTenantDataIntegrityValidation 租户数据完整性验证测试
func TestTenantDataIntegrityValidation(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// 初始化数据库连接
		err := config.InitDatabase()
		if err != nil {
			t.Logf("数据库连接失败，跳过测试: %v", err)
			return
		}

		userRepo := repository.NewUserRepository()
		
		// 创建多个租户的测试数据
		tenantContexts := []context.Context{
			context.WithValue(context.Background(), "tenant_id", uint64(1)),
			context.WithValue(context.Background(), "tenant_id", uint64(2)),
		}
		
		var createdUsers []uint64
		
		// 在每个租户下创建用户
		for i, ctx := range tenantContexts {
			user := &types.User{
				UUID:     fmt.Sprintf("integrity-test-user-%d", i+1),
				Username: fmt.Sprintf("integrityuser%d", i+1),
				Email:    fmt.Sprintf("integrity%d@example.com", i+1),
				Status:   types.UserStatusActive,
			}
			
			err := userRepo.Create(ctx, user)
			if err != nil {
				t.Logf("创建用户失败: %v", err)
				continue
			}
			
			if user.ID > 0 {
				createdUsers = append(createdUsers, user.ID)
			}
		}
		
		// 验证数据完整性：每个租户只能看到自己的数据
		for i, ctx := range tenantContexts {
			users, err := userRepo.FindAllByTenant(ctx)
			if err != nil {
				t.Logf("查询租户 %d 用户失败: %v", i+1, err)
				continue
			}
			
			// 验证查询到的用户都属于当前租户
			expectedTenantID := uint64(i + 1)
			for _, user := range users {
				t.AssertEQ(user.TenantID, expectedTenantID)
			}
			
			// 验证不会看到其他租户的测试用户
			for _, user := range users {
				otherTenantUsername := fmt.Sprintf("integrityuser%d", 3-i-1) // 另一个租户的用户名
				t.AssertNE(user.Username, otherTenantUsername)
			}
		}
		
		// 清理测试数据
		for i, userID := range createdUsers {
			if i < len(tenantContexts) {
				userRepo.DeleteByID(tenantContexts[i], userID)
			}
		}
	})
}
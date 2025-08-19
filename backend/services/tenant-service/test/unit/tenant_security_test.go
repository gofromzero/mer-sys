package unit

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/services/tenant-service/internal/service"
)

func TestTenantSecurityService(t *testing.T) {
	Convey("租户安全服务测试", t, func() {
		// 模拟上下文和权限
		ctx := context.Background()
		tenantService := service.NewTenantService()
		securityService := service.NewTenantSecurityService(tenantService)

		Convey("资源访问权限验证", func() {
			Convey("同租户访问应该成功", func() {
				// 模拟租户1的上下文
				ctx = context.WithValue(ctx, "tenant_id", uint64(1))
				
				err := securityService.ValidateResourceAccess(ctx, 1, 1, "read")
				So(err, ShouldBeNil)
			})

			Convey("跨租户访问应该被拒绝（无系统权限）", func() {
				// 模拟租户1尝试访问租户2的资源
				ctx = context.WithValue(ctx, "tenant_id", uint64(1))
				
				err := securityService.ValidateResourceAccess(ctx, 1, 2, "read")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "无权访问其他租户的资源")
			})

			Convey("跨租户访问应该成功（有系统权限）", func() {
				// 模拟具有系统权限的用户
				ctx = context.WithValue(ctx, "tenant_id", uint64(1))
				userPerms := types.UserPermissions{
					UserID:      1,
					TenantID:    1,
					Roles:       []types.RoleType{types.RoleTenantAdmin},
					Permissions: []types.Permission{types.PermissionTenantManage},
				}
				ctx = context.WithValue(ctx, "user_permissions", userPerms)
				
				err := securityService.ValidateResourceAccess(ctx, 1, 2, "read")
				So(err, ShouldBeNil)
			})

			Convey("租户ID不匹配应该被拒绝", func() {
				// 上下文中的租户ID与用户声称的不匹配
				ctx = context.WithValue(ctx, "tenant_id", uint64(2))
				
				err := securityService.ValidateResourceAccess(ctx, 1, 1, "read")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "租户ID验证失败")
			})
		})

		Convey("租户操作权限验证", func() {
			ctx = context.WithValue(ctx, "tenant_id", uint64(1))

			Convey("普通操作不需要系统级权限", func() {
				err := securityService.ValidateTenantOperation(ctx, 1, "read", false)
				So(err, ShouldBeNil)
			})

			Convey("系统级操作需要相应权限", func() {
				Convey("无权限时应该被拒绝", func() {
					err := securityService.ValidateTenantOperation(ctx, 1, "manage", true)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "权限不足：需要系统级租户管理权限")
				})

				Convey("有权限时应该成功", func() {
					userPerms := types.UserPermissions{
						UserID:      1,
						TenantID:    1,
						Permissions: []types.Permission{types.PermissionTenantManage},
					}
					ctx = context.WithValue(ctx, "user_permissions", userPerms)
					
					err := securityService.ValidateTenantOperation(ctx, 1, "manage", true)
					So(err, ShouldBeNil)
				})
			})
		})

		Convey("租户状态变更验证", func() {
			ctx = context.WithValue(ctx, "tenant_id", uint64(1))
			userPerms := types.UserPermissions{
				UserID:      1,
				TenantID:    1,
				Permissions: []types.Permission{types.PermissionTenantManage},
			}
			ctx = context.WithValue(ctx, "user_permissions", userPerms)

			Convey("有效状态转换应该成功", func() {
				validTransitions := []struct {
					from types.TenantStatus
					to   types.TenantStatus
				}{
					{types.TenantStatusActive, types.TenantStatusSuspended},
					{types.TenantStatusActive, types.TenantStatusExpired},
					{types.TenantStatusSuspended, types.TenantStatusActive},
					{types.TenantStatusSuspended, types.TenantStatusExpired},
					{types.TenantStatusExpired, types.TenantStatusActive},
					{types.TenantStatusExpired, types.TenantStatusSuspended},
				}

				for _, transition := range validTransitions {
					err := securityService.ValidateTenantStatusChange(ctx, 1, 
						transition.from, transition.to, "测试状态转换")
					So(err, ShouldBeNil)
				}
			})

			Convey("无效状态转换应该被拒绝", func() {
				// 尝试从不存在的状态转换
				err := securityService.ValidateTenantStatusChange(ctx, 1, 
					types.TenantStatus("invalid"), types.TenantStatusActive, "无效转换")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "无效的状态转换")
			})
		})

		Convey("租户配置变更验证", func() {
			ctx = context.WithValue(ctx, "tenant_id", uint64(1))
			userPerms := types.UserPermissions{
				UserID:      1,
				TenantID:    1,
				Permissions: []types.Permission{types.PermissionTenantManage},
			}
			ctx = context.WithValue(ctx, "user_permissions", userPerms)

			oldConfig := &types.TenantConfig{
				MaxUsers:     100,
				MaxMerchants: 50,
				Features:     []string{"basic"},
				Settings:     map[string]string{},
			}

			Convey("合理的配置变更应该成功", func() {
				newConfig := &types.TenantConfig{
					MaxUsers:     200,
					MaxMerchants: 100,
					Features:     []string{"basic", "advanced_report"},
					Settings:     map[string]string{"theme": "dark"},
				}

				err := securityService.ValidateTenantConfigChange(ctx, 1, oldConfig, newConfig)
				So(err, ShouldBeNil)
			})

			Convey("无效配置值应该被拒绝", func() {
				invalidConfigs := []*types.TenantConfig{
					{MaxUsers: 0, MaxMerchants: 50, Features: []string{"basic"}},     // 无效用户数
					{MaxUsers: 100, MaxMerchants: 0, Features: []string{"basic"}},   // 无效商户数
					{MaxUsers: -1, MaxMerchants: 50, Features: []string{"basic"}},   // 负数用户数
				}

				for _, config := range invalidConfigs {
					err := securityService.ValidateTenantConfigChange(ctx, 1, oldConfig, config)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "配置值无效")
				}
			})

			Convey("危险的配置变更应该被拒绝", func() {
				dangerousConfigs := []*types.TenantConfig{
					// 大幅缩减用户数
					{MaxUsers: 40, MaxMerchants: 50, Features: []string{"basic"}},
					// 大幅缩减商户数
					{MaxUsers: 100, MaxMerchants: 20, Features: []string{"basic"}},
					// 移除关键功能
					{MaxUsers: 100, MaxMerchants: 50, Features: []string{}},
				}

				for _, config := range dangerousConfigs {
					err := securityService.ValidateTenantConfigChange(ctx, 1, oldConfig, config)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "拒绝危险的配置变更")
				}
			})
		})

		Convey("敏感操作验证", func() {
			ctx = context.WithValue(ctx, "tenant_id", uint64(1))

			Convey("无权限时应该被拒绝", func() {
				err := securityService.ValidateSensitiveOperation(ctx, "delete_tenant", 
					map[string]interface{}{"tenant_id": 1})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "权限不足：无法执行敏感操作")
			})

			Convey("有权限时应该成功", func() {
				userPerms := types.UserPermissions{
					UserID:      1,
					TenantID:    1,
					Permissions: []types.Permission{types.PermissionTenantManage},
				}
				ctx = context.WithValue(ctx, "user_permissions", userPerms)

				err := securityService.ValidateSensitiveOperation(ctx, "delete_tenant", 
					map[string]interface{}{"tenant_id": 1})
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestTenantSecurityEdgeCases(t *testing.T) {
	Convey("租户安全服务边界情况测试", t, func() {
		tenantService := service.NewTenantService()
		securityService := service.NewTenantSecurityService(tenantService)

		Convey("空上下文处理", func() {
			ctx := context.Background()

			err := securityService.ValidateResourceAccess(ctx, 1, 1, "read")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "租户ID验证失败")
		})

		Convey("无效权限数据", func() {
			ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
			ctx = context.WithValue(ctx, "user_permissions", "invalid_data")

			err := securityService.ValidateTenantOperation(ctx, 1, "manage", true)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "权限不足")
		})

		Convey("状态转换边界情况", func() {
			ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
			userPerms := types.UserPermissions{
				UserID:      1,
				TenantID:    1,
				Permissions: []types.Permission{types.PermissionTenantManage},
			}
			ctx = context.WithValue(ctx, "user_permissions", userPerms)

			Convey("空状态", func() {
				err := securityService.ValidateTenantStatusChange(ctx, 1, 
					types.TenantStatus(""), types.TenantStatusActive, "空状态测试")
				So(err, ShouldNotBeNil)
			})

			Convey("相同状态转换", func() {
				err := securityService.ValidateTenantStatusChange(ctx, 1, 
					types.TenantStatusActive, types.TenantStatusActive, "相同状态")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "无效的状态转换")
			})
		})

		Convey("配置变更边界情况", func() {
			ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
			userPerms := types.UserPermissions{
				UserID:      1,
				TenantID:    1,
				Permissions: []types.Permission{types.PermissionTenantManage},
			}
			ctx = context.WithValue(ctx, "user_permissions", userPerms)

			Convey("空配置", func() {
				err := securityService.ValidateTenantConfigChange(ctx, 1, nil, nil)
				So(err, ShouldNotBeNil)
			})

			Convey("边界值配置", func() {
				oldConfig := &types.TenantConfig{
					MaxUsers:     1,
					MaxMerchants: 1,
					Features:     []string{"basic"},
				}
				newConfig := &types.TenantConfig{
					MaxUsers:     1,
					MaxMerchants: 1,
					Features:     []string{"basic"},
				}

				err := securityService.ValidateTenantConfigChange(ctx, 1, oldConfig, newConfig)
				So(err, ShouldBeNil)
			})
		})
	})
}
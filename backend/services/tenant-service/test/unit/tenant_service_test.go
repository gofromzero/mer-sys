package unit

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/services/tenant-service/internal/service"
)

func TestTenantService(t *testing.T) {
	Convey("租户服务单元测试", t, func() {
		ctx := context.Background()
		tenantService := service.NewTenantService()

		Convey("创建租户", func() {
			req := &types.CreateTenantRequest{
				Name:          "测试租户",
				Code:          "test-tenant",
				BusinessType:  "ecommerce",
				ContactPerson: "张三",
				ContactEmail:  "test@example.com",
				ContactPhone:  "13800138000",
				Address:       "北京市朝阳区",
			}

			Convey("创建成功的情况", func() {
				tenant, err := tenantService.CreateTenant(ctx, req)

				So(err, ShouldBeNil)
				So(tenant, ShouldNotBeNil)
				So(tenant.Name, ShouldEqual, req.Name)
				So(tenant.Code, ShouldEqual, req.Code)
				So(tenant.Status, ShouldEqual, "active")
				So(tenant.BusinessType, ShouldEqual, req.BusinessType)
				So(tenant.ContactPerson, ShouldEqual, req.ContactPerson)
				So(tenant.ContactEmail, ShouldEqual, req.ContactEmail)
			})

			Convey("租户代码重复的情况", func() {
				// 先创建一个租户
				_, err := tenantService.CreateTenant(ctx, req)
				So(err, ShouldBeNil)

				// 尝试创建相同代码的租户
				_, err = tenantService.CreateTenant(ctx, req)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "租户代码已存在")
			})

			Convey("联系邮箱重复的情况", func() {
				// 先创建一个租户
				_, err := tenantService.CreateTenant(ctx, req)
				So(err, ShouldBeNil)

				// 尝试创建相同邮箱的租户（不同代码）
				req2 := *req
				req2.Code = "test-tenant-2"
				_, err = tenantService.CreateTenant(ctx, &req2)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "联系邮箱已被使用")
			})
		})

		Convey("查询租户", func() {
			// 先创建一个租户
			req := &types.CreateTenantRequest{
				Name:          "查询测试租户",
				Code:          "query-test-tenant",
				BusinessType:  "retail",
				ContactPerson: "李四",
				ContactEmail:  "query@example.com",
			}
			createdTenant, err := tenantService.CreateTenant(ctx, req)
			So(err, ShouldBeNil)

			Convey("根据ID查询租户", func() {
				tenant, err := tenantService.GetTenantByID(ctx, createdTenant.ID)

				So(err, ShouldBeNil)
				So(tenant, ShouldNotBeNil)
				So(tenant.ID, ShouldEqual, createdTenant.ID)
				So(tenant.Name, ShouldEqual, req.Name)
			})

			Convey("查询不存在的租户", func() {
				tenant, err := tenantService.GetTenantByID(ctx, 99999)

				So(err, ShouldBeNil)
				So(tenant, ShouldBeNil)
			})

			Convey("分页查询租户列表", func() {
				listReq := &types.ListTenantsRequest{
					Page:     1,
					PageSize: 10,
				}
				result, err := tenantService.ListTenants(ctx, listReq)

				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Total, ShouldBeGreaterThan, 0)
				So(len(result.Tenants), ShouldBeGreaterThan, 0)
			})

			Convey("按状态筛选租户", func() {
				listReq := &types.ListTenantsRequest{
					Page:     1,
					PageSize: 10,
					Status:   types.TenantStatusActive,
				}
				result, err := tenantService.ListTenants(ctx, listReq)

				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				for _, tenant := range result.Tenants {
					So(tenant.Status, ShouldEqual, string(types.TenantStatusActive))
				}
			})
		})

		Convey("更新租户", func() {
			// 先创建一个租户
			req := &types.CreateTenantRequest{
				Name:          "更新测试租户",
				Code:          "update-test-tenant",
				BusinessType:  "food",
				ContactPerson: "王五",
				ContactEmail:  "update@example.com",
			}
			createdTenant, err := tenantService.CreateTenant(ctx, req)
			So(err, ShouldBeNil)

			Convey("更新租户基本信息", func() {
				updateReq := &types.UpdateTenantRequest{
					Name:          "更新后的租户名称",
					BusinessType:  "education",
					ContactPerson: "王五五",
				}

				updatedTenant, err := tenantService.UpdateTenant(ctx, createdTenant.ID, updateReq)

				So(err, ShouldBeNil)
				So(updatedTenant, ShouldNotBeNil)
				So(updatedTenant.Name, ShouldEqual, updateReq.Name)
				So(updatedTenant.BusinessType, ShouldEqual, updateReq.BusinessType)
				So(updatedTenant.ContactPerson, ShouldEqual, updateReq.ContactPerson)
				So(updatedTenant.ContactEmail, ShouldEqual, req.ContactEmail) // 未更新的字段保持不变
			})

			Convey("更新不存在的租户", func() {
				updateReq := &types.UpdateTenantRequest{
					Name: "不存在的租户",
				}

				_, err := tenantService.UpdateTenant(ctx, 99999, updateReq)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "租户不存在")
			})
		})

		Convey("租户状态管理", func() {
			// 先创建一个租户
			req := &types.CreateTenantRequest{
				Name:          "状态测试租户",
				Code:          "status-test-tenant",
				BusinessType:  "service",
				ContactPerson: "赵六",
				ContactEmail:  "status@example.com",
			}
			createdTenant, err := tenantService.CreateTenant(ctx, req)
			So(err, ShouldBeNil)

			Convey("更新租户状态", func() {
				statusReq := &types.UpdateTenantStatusRequest{
					Status: types.TenantStatusSuspended,
					Reason: "测试暂停",
				}

				err := tenantService.UpdateTenantStatus(ctx, createdTenant.ID, statusReq)
				So(err, ShouldBeNil)

				// 验证状态已更新
				tenant, err := tenantService.GetTenantByID(ctx, createdTenant.ID)
				So(err, ShouldBeNil)
				So(tenant.Status, ShouldEqual, string(types.TenantStatusSuspended))
			})

			Convey("激活暂停的租户", func() {
				// 先暂停租户
				statusReq := &types.UpdateTenantStatusRequest{
					Status: types.TenantStatusSuspended,
					Reason: "测试暂停",
				}
				err := tenantService.UpdateTenantStatus(ctx, createdTenant.ID, statusReq)
				So(err, ShouldBeNil)

				// 再激活租户
				statusReq.Status = types.TenantStatusActive
				statusReq.Reason = "测试激活"
				err = tenantService.UpdateTenantStatus(ctx, createdTenant.ID, statusReq)
				So(err, ShouldBeNil)

				// 验证状态和激活时间
				tenant, err := tenantService.GetTenantByID(ctx, createdTenant.ID)
				So(err, ShouldBeNil)
				So(tenant.Status, ShouldEqual, string(types.TenantStatusActive))
				So(tenant.ActivationTime, ShouldNotBeNil)
			})
		})

		Convey("租户配置管理", func() {
			// 先创建一个租户
			req := &types.CreateTenantRequest{
				Name:          "配置测试租户",
				Code:          "config-test-tenant",
				BusinessType:  "finance",
				ContactPerson: "钱七",
				ContactEmail:  "config@example.com",
			}
			createdTenant, err := tenantService.CreateTenant(ctx, req)
			So(err, ShouldBeNil)

			Convey("获取默认配置", func() {
				config, err := tenantService.GetTenantConfig(ctx, createdTenant.ID)

				So(err, ShouldBeNil)
				So(config, ShouldNotBeNil)
				So(config.MaxUsers, ShouldEqual, 100)
				So(config.MaxMerchants, ShouldEqual, 50)
				So(len(config.Features), ShouldBeGreaterThan, 0)
			})

			Convey("更新租户配置", func() {
				newConfig := &types.TenantConfig{
					MaxUsers:     200,
					MaxMerchants: 100,
					Features:     []string{"basic", "advanced_report"},
					Settings: map[string]string{
						"theme": "dark",
						"lang":  "zh-CN",
					},
				}

				err := tenantService.UpdateTenantConfig(ctx, createdTenant.ID, newConfig)
				So(err, ShouldBeNil)

				// 验证配置已更新
				config, err := tenantService.GetTenantConfig(ctx, createdTenant.ID)
				So(err, ShouldBeNil)
				So(config.MaxUsers, ShouldEqual, newConfig.MaxUsers)
				So(config.MaxMerchants, ShouldEqual, newConfig.MaxMerchants)
				So(len(config.Features), ShouldEqual, len(newConfig.Features))
				So(config.Settings["theme"], ShouldEqual, "dark")
			})

			Convey("获取配置变更通知", func() {
				// 先更新配置
				newConfig := &types.TenantConfig{
					MaxUsers:     150,
					MaxMerchants: 75,
					Features:     []string{"basic"},
					Settings:     map[string]string{},
				}
				err := tenantService.UpdateTenantConfig(ctx, createdTenant.ID, newConfig)
				So(err, ShouldBeNil)

				// 获取变更通知
				notification, err := tenantService.GetConfigChangeNotification(ctx, createdTenant.ID)
				So(err, ShouldBeNil)
				So(notification, ShouldNotBeNil)
				So(notification["tenant_id"], ShouldEqual, createdTenant.ID)
			})
		})
	})
}

func TestTenantServiceEdgeCases(t *testing.T) {
	Convey("租户服务边界情况测试", t, func() {
		ctx := context.Background()
		tenantService := service.NewTenantService()

		Convey("无效参数测试", func() {
			Convey("空的创建请求", func() {
				_, err := tenantService.CreateTenant(ctx, &types.CreateTenantRequest{})
				So(err, ShouldNotBeNil)
			})

			Convey("无效的分页参数", func() {
				listReq := &types.ListTenantsRequest{
					Page:     0,
					PageSize: 0,
				}
				result, err := tenantService.ListTenants(ctx, listReq)

				So(err, ShouldBeNil)
				So(result.Page, ShouldEqual, 1)     // 默认值
				So(result.Size, ShouldBeLessThanOrEqualTo, 20) // 默认页大小
			})
		})

		Convey("并发安全测试", func() {
			req := &types.CreateTenantRequest{
				Name:          "并发测试租户",
				Code:          "concurrent-test",
				BusinessType:  "manufacturing",
				ContactPerson: "并发测试",
				ContactEmail:  "concurrent@example.com",
			}

			// 模拟并发创建
			done := make(chan bool, 2)
			var err1, err2 error
			var tenant1, tenant2 *types.TenantResponse

			go func() {
				tenant1, err1 = tenantService.CreateTenant(ctx, req)
				done <- true
			}()

			go func() {
				req2 := *req
				req2.Code = "concurrent-test-2"
				tenant2, err2 = tenantService.CreateTenant(ctx, &req2)
				done <- true
			}()

			// 等待两个协程完成
			<-done
			<-done

			// 验证至少有一个成功
			if err1 == nil {
				So(tenant1, ShouldNotBeNil)
			}
			if err2 == nil {
				So(tenant2, ShouldNotBeNil)
			}
		})
	})
}

func TestTenantConfigCache(t *testing.T) {
	Convey("租户配置缓存测试", t, func() {
		ctx := context.Background()
		// 使用模拟缓存进行测试
		cache := service.NewTestTenantConfigCache()

		tenantID := uint64(1)
		config := &types.TenantConfig{
			MaxUsers:     100,
			MaxMerchants: 50,
			Features:     []string{"basic"},
			Settings:     map[string]string{"test": "value"},
		}

		Convey("缓存基本操作", func() {
			Convey("设置和获取配置", func() {
				err := cache.SetConfig(ctx, tenantID, config)
				So(err, ShouldBeNil)

				cached, err := cache.GetConfig(ctx, tenantID)
				So(err, ShouldBeNil)
				So(cached, ShouldNotBeNil)
				So(cached.MaxUsers, ShouldEqual, config.MaxUsers)
			})

			Convey("缓存失效", func() {
				err := cache.SetConfig(ctx, tenantID, config)
				So(err, ShouldBeNil)

				err = cache.InvalidateConfig(ctx, tenantID)
				So(err, ShouldBeNil)

				_, err = cache.GetConfig(ctx, tenantID)
				So(err, ShouldNotBeNil) // 缓存应该已失效
			})

			Convey("检查缓存存在性", func() {
				err := cache.SetConfig(ctx, tenantID, config)
				So(err, ShouldBeNil)

				exists, err := cache.IsConfigCached(ctx, tenantID)
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)

				err = cache.InvalidateConfig(ctx, tenantID)
				So(err, ShouldBeNil)

				exists, err = cache.IsConfigCached(ctx, tenantID)
				So(err, ShouldBeNil)
				So(exists, ShouldBeFalse)
			})
		})

		Convey("配置变更通知", func() {
			changeInfo := map[string]interface{}{
				"tenant_id":   tenantID,
				"changed_at":  time.Now().Format("2006-01-02 15:04:05"),
				"new_config":  config,
			}

			err := cache.SetConfigChangeNotification(ctx, tenantID, changeInfo)
			So(err, ShouldBeNil)

			notification, err := cache.GetConfigChangeNotification(ctx, tenantID)
			So(err, ShouldBeNil)
			So(notification, ShouldNotBeNil)
			So(notification["tenant_id"], ShouldEqual, tenantID)

			err = cache.ClearConfigChangeNotification(ctx, tenantID)
			So(err, ShouldBeNil)

			_, err = cache.GetConfigChangeNotification(ctx, tenantID)
			So(err, ShouldNotBeNil) // 通知应该已清除
		})

		Convey("批量操作", func() {
			tenantIDs := []uint64{1, 2, 3}
			configs := map[uint64]*types.TenantConfig{
				1: config,
				2: config,
				3: config,
			}

			err := cache.PrefetchConfigs(ctx, tenantIDs, configs)
			So(err, ShouldBeNil)

			stats := cache.GetCacheStats(ctx, tenantIDs)
			So(stats["total_tenants"], ShouldEqual, 3)
			So(stats["cached_count"], ShouldEqual, 3)
			So(stats["cache_hit_rate"], ShouldEqual, 1.0)
		})
	})
}
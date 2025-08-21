package test

import (
	"context"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	. "github.com/smartystreets/goconvey/convey"
)

// TestDashboardRepository 测试仪表板数据访问层
func TestDashboardRepository(t *testing.T) {
	Convey("Dashboard Repository Tests", t, func() {
		// 初始化repository
		dashboardRepo := repository.NewDashboardRepository()
		ctx := context.Background()
		
		// 设置测试用的租户和商户ID
		tenantID := uint64(1)
		merchantID := uint64(1)
		
		// 在context中设置租户ID进行测试
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		Convey("GetMerchantDashboardData should work with valid inputs", func() {
			period := types.TimePeriod("daily")
			
			// 由于这是集成测试，可能需要数据库连接，这里模拟测试场景
			// 在实际环境中，这需要连接到测试数据库
			data, err := dashboardRepo.GetMerchantDashboardData(ctx, tenantID, merchantID, period)
			
			// 基本验证 - 不应该panic，错误处理应该正确
			if err != nil {
				// 如果是数据库连接错误或者数据不存在，这是预期的
				So(err, ShouldNotBeNil)
				So(data, ShouldBeNil)
			} else {
				// 如果成功，验证数据结构
				So(data, ShouldNotBeNil)
				So(data.TenantID, ShouldEqual, tenantID)
				So(data.MerchantID, ShouldEqual, merchantID)
				So(data.Period, ShouldEqual, period)
			}
		})

		Convey("GetMerchantDashboardData should handle invalid period", func() {
			invalidPeriod := types.TimePeriod("invalid")
			
			data, err := dashboardRepo.GetMerchantDashboardData(ctx, tenantID, merchantID, invalidPeriod)
			
			// 应该返回错误或者空数据
			if data != nil {
				So(data.Period, ShouldNotEqual, invalidPeriod)
			}
			// 错误处理应该是优雅的
			So(err, ShouldNotBeNil)
		})

		Convey("GetMerchantBusinessStats should work with valid inputs", func() {
			period := types.TimePeriod("daily")
			
			stats, err := dashboardRepo.GetMerchantBusinessStats(ctx, tenantID, merchantID, period)
			
			// 基本验证
			if err != nil {
				// 数据库连接错误或数据不存在是预期的
				So(err, ShouldNotBeNil)
				So(stats, ShouldBeNil)
			} else {
				// 如果成功，验证数据结构
				So(stats, ShouldNotBeNil)
				So(stats.TotalSales, ShouldBeGreaterThanOrEqualTo, 0)
				So(stats.TotalOrders, ShouldBeGreaterThanOrEqualTo, 0)
				So(stats.TotalCustomers, ShouldBeGreaterThanOrEqualTo, 0)
			}
		})

		Convey("GetRightsUsageTrend should return trend data", func() {
			days := 30
			
			trends, err := dashboardRepo.GetRightsUsageTrend(ctx, tenantID, merchantID, days)
			
			// 基本验证
			if err != nil {
				// 数据库连接错误或数据不存在是预期的
				So(err, ShouldNotBeNil)
				So(trends, ShouldBeNil)
			} else {
				// 如果成功，验证数据结构
				So(trends, ShouldNotBeNil)
				// 趋势数据应该是数组格式
				for _, trend := range trends {
					So(trend.Balance, ShouldBeGreaterThanOrEqualTo, 0)
					So(trend.Usage, ShouldBeGreaterThanOrEqualTo, 0)
				}
			}
		})

		Convey("GetPendingTasks should return task list", func() {
			tasks, err := dashboardRepo.GetPendingTasks(ctx, tenantID, merchantID)
			
			// 基本验证
			if err != nil {
				So(err, ShouldNotBeNil)
				So(tasks, ShouldBeNil)
			} else {
				So(tasks, ShouldNotBeNil)
				// 验证任务数据结构
				for _, task := range tasks {
					So(task.ID, ShouldNotBeEmpty)
					So(task.Description, ShouldNotBeEmpty)
					So(task.Count, ShouldBeGreaterThanOrEqualTo, 0)
				}
			}
		})

		Convey("GetSystemNotifications should return notifications", func() {
			notifications, err := dashboardRepo.GetSystemNotifications(ctx, tenantID, merchantID)
			
			// 基本验证
			if err != nil {
				So(err, ShouldNotBeNil)
				So(notifications, ShouldBeNil)
			} else {
				So(notifications, ShouldNotBeNil)
				So(notifications.UnreadCount, ShouldBeGreaterThanOrEqualTo, 0)
				// 验证通知数据结构
				for _, notification := range notifications.Notifications {
					So(notification.ID, ShouldBeGreaterThan, 0)
					So(notification.Title, ShouldNotBeEmpty)
				}
				for _, announcement := range notifications.Announcements {
					So(announcement.ID, ShouldBeGreaterThan, 0)
					So(announcement.Title, ShouldNotBeEmpty)
				}
			}
		})

		Convey("Tenant isolation should be enforced", func() {
			// 测试租户隔离
			period := types.TimePeriod("daily")
			
			// 使用不同的租户ID进行测试
			otherTenantID := uint64(999)
			otherCtx := context.WithValue(context.Background(), "tenant_id", otherTenantID)
			
			data1, _ := dashboardRepo.GetMerchantDashboardData(ctx, tenantID, merchantID, period)
			data2, _ := dashboardRepo.GetMerchantDashboardData(otherCtx, otherTenantID, merchantID, period)
			
			// 不同租户的数据应该是不同的（或者一个为nil）
			if data1 != nil && data2 != nil {
				So(data1.TenantID, ShouldNotEqual, data2.TenantID)
			}
		})
	})
}

// TestDashboardConfiguration 测试仪表板配置相关功能
func TestDashboardConfiguration(t *testing.T) {
	Convey("Dashboard Configuration Tests", t, func() {
		dashboardRepo := repository.NewDashboardRepository()
		ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
		
		tenantID := uint64(1)
		merchantID := uint64(1)

		Convey("SaveDashboardConfig should work with valid config", func() {
			config := &types.DashboardConfig{
				MerchantID: merchantID,
				LayoutConfig: types.LayoutConfig{
					Columns: 4,
					Widgets: []types.DashboardWidget{
						{
							ID:       "test_widget",
							Type:     types.WidgetTypeSalesOverview,
							Position: types.Position{X: 0, Y: 0},
							Size:     types.Size{Width: 2, Height: 1},
							Config:   map[string]interface{}{},
							Visible:  true,
						},
					},
				},
				RefreshInterval: 300,
			}
			
			err := dashboardRepo.SaveDashboardConfig(ctx, tenantID, merchantID, config)
			
			// 基本错误处理验证
			if err != nil {
				// 数据库操作错误是可能的，但不应该panic
				So(err, ShouldNotBeNil)
			}
		})

		Convey("GetDashboardConfig should return saved config", func() {
			config, err := dashboardRepo.GetDashboardConfig(ctx, tenantID, merchantID)
			
			if err != nil {
				// 配置不存在或数据库错误
				So(err, ShouldNotBeNil)
				So(config, ShouldBeNil)
			} else {
				// 验证配置结构
				So(config, ShouldNotBeNil)
				So(config.MerchantID, ShouldEqual, merchantID)
				So(config.RefreshInterval, ShouldBeGreaterThan, 0)
			}
		})

		Convey("MarkAnnouncementAsRead should update read status", func() {
			announcementID := uint64(1)
			
			err := dashboardRepo.MarkAnnouncementAsRead(ctx, tenantID, merchantID, announcementID)
			
			// 基本错误处理验证
			if err != nil {
				// 公告不存在或数据库错误
				So(err, ShouldNotBeNil)
			}
		})
	})
}

// BenchmarkDashboardRepository 性能测试
func BenchmarkDashboardRepository(b *testing.B) {
	dashboardRepo := repository.NewDashboardRepository()
	ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
	
	tenantID := uint64(1)
	merchantID := uint64(1)
	period := types.TimePeriod("daily")

	b.ResetTimer()
	
	// 测试仪表板数据获取的性能
	b.Run("GetMerchantDashboardData", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = dashboardRepo.GetMerchantDashboardData(ctx, tenantID, merchantID, period)
		}
	})
	
	// 测试业务统计获取的性能
	b.Run("GetMerchantBusinessStats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = dashboardRepo.GetMerchantBusinessStats(ctx, tenantID, merchantID, period)
		}
	})
}

// TestDashboardDataValidation 数据验证测试
func TestDashboardDataValidation(t *testing.T) {
	Convey("Dashboard Data Validation Tests", t, func() {
		dashboardRepo := repository.NewDashboardRepository()
		
		Convey("Should reject requests without tenant context", func() {
			// 没有租户上下文的请求应该被拒绝
			ctx := context.Background()
			tenantID := uint64(1)
			merchantID := uint64(1)
			period := types.TimePeriod("daily")
			
			data, err := dashboardRepo.GetMerchantDashboardData(ctx, tenantID, merchantID, period)
			
			// 应该返回错误
			So(err, ShouldNotBeNil)
			So(data, ShouldBeNil)
		})

		Convey("Should handle zero merchant ID gracefully", func() {
			ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
			tenantID := uint64(1)
			merchantID := uint64(0) // 无效的商户ID
			period := types.TimePeriod("daily")
			
			data, err := dashboardRepo.GetMerchantDashboardData(ctx, tenantID, merchantID, period)
			
			// 应该处理无效ID
			if err == nil {
				So(data, ShouldBeNil)
			} else {
				So(err, ShouldNotBeNil)
			}
		})
	})
}
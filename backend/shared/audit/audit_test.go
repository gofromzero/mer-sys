package audit

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuditEventTypes(t *testing.T) {
	Convey("审计事件类型测试", t, func() {
		
		Convey("基础事件类型应该正确定义", func() {
			So(string(EventCrossTenantAttempt), ShouldEqual, "cross_tenant_attempt")
			So(string(EventTenantAccess), ShouldEqual, "tenant_access")
			So(string(EventDataQuery), ShouldEqual, "data_query")
			So(string(EventSecurityViolation), ShouldEqual, "security_violation")
		})
		
		Convey("商户用户事件类型应该正确定义", func() {
			So(string(EventMerchantUserLogin), ShouldEqual, "merchant_user_login")
			So(string(EventMerchantUserLogout), ShouldEqual, "merchant_user_logout")
			So(string(EventMerchantUserCreate), ShouldEqual, "merchant_user_create")
			So(string(EventMerchantUserUpdate), ShouldEqual, "merchant_user_update")
			So(string(EventMerchantUserDelete), ShouldEqual, "merchant_user_delete")
			So(string(EventMerchantUserDisable), ShouldEqual, "merchant_user_disable")
			So(string(EventMerchantUserEnable), ShouldEqual, "merchant_user_enable")
			So(string(EventMerchantUserPassword), ShouldEqual, "merchant_user_password")
			So(string(EventMerchantOperation), ShouldEqual, "merchant_operation")
		})
	})
}

func TestAuditSeverity(t *testing.T) {
	Convey("审计严重程度测试", t, func() {
		
		Convey("严重程度级别应该正确定义", func() {
			So(string(SeverityInfo), ShouldEqual, "info")
			So(string(SeverityWarning), ShouldEqual, "warning")
			So(string(SeverityError), ShouldEqual, "error")
			So(string(SeverityCritical), ShouldEqual, "critical")
		})
	})
}

func TestAuditLogger(t *testing.T) {
	Convey("审计日志器测试", t, func() {
		logger := NewAuditLogger()
		ctx := context.Background()
		
		Convey("审计日志器应该可以创建", func() {
			So(logger, ShouldNotBeNil)
		})
		
		Convey("商户用户登录审计应该正确记录", func() {
			// 这里只测试方法存在性，避免依赖日志系统
			var logMerchantUserLogin func(*AuditLogger, context.Context, uint64, uint64, uint64, string, string, interface{}) = (*AuditLogger).LogMerchantUserLogin
			So(logMerchantUserLogin, ShouldNotBeNil)
			
			// 测试调用不会panic
			So(func() {
				logger.LogMerchantUserLogin(ctx, 1, 1, 1, "test_user", "192.168.1.1", nil)
			}, ShouldNotPanic)
		})
		
		Convey("商户用户创建审计应该正确记录", func() {
			var logMerchantUserCreate func(*AuditLogger, context.Context, uint64, uint64, uint64, uint64, string, interface{}) = (*AuditLogger).LogMerchantUserCreate
			So(logMerchantUserCreate, ShouldNotBeNil)
			
			So(func() {
				logger.LogMerchantUserCreate(ctx, 1, 1, 1, 2, "new_user", map[string]interface{}{
					"role": "merchant_operator",
				})
			}, ShouldNotPanic)
		})
		
		Convey("商户用户密码重置审计应该正确记录", func() {
			var logMerchantUserPasswordReset func(*AuditLogger, context.Context, uint64, uint64, uint64, uint64, string, string) = (*AuditLogger).LogMerchantUserPasswordReset
			So(logMerchantUserPasswordReset, ShouldNotBeNil)
			
			So(func() {
				logger.LogMerchantUserPasswordReset(ctx, 1, 1, 1, 2, "target_user", "admin_reset")
			}, ShouldNotPanic)
		})
		
		Convey("商户用户状态变更审计应该正确记录", func() {
			var logMerchantUserStatusChange func(*AuditLogger, context.Context, uint64, uint64, uint64, uint64, string, string, string, interface{}) = (*AuditLogger).LogMerchantUserStatusChange
			So(logMerchantUserStatusChange, ShouldNotBeNil)
			
			So(func() {
				logger.LogMerchantUserStatusChange(ctx, 1, 1, 1, 2, "target_user", "active", "inactive", nil)
			}, ShouldNotPanic)
		})
	})
}

func TestAuditEvent(t *testing.T) {
	Convey("审计事件结构测试", t, func() {
		
		Convey("审计事件应该包含必要字段", func() {
			merchantID := uint64(123)
			targetUserID := uint64(456)
			
			event := AuditEvent{
				EventType:    EventMerchantUserCreate,
				Severity:     SeverityInfo,
				TenantID:     1,
				UserID:       2,
				MerchantID:   &merchantID,
				TargetUserID: &targetUserID,
				ResourceType: "merchant_user",
				ResourceID:   "456",
				Action:       "create",
				Message:      "创建商户用户测试",
				Timestamp:    time.Now(),
			}
			
			So(event.EventType, ShouldEqual, EventMerchantUserCreate)
			So(event.Severity, ShouldEqual, SeverityInfo)
			So(event.TenantID, ShouldEqual, 1)
			So(event.UserID, ShouldEqual, 2)
			So(*event.MerchantID, ShouldEqual, 123)
			So(*event.TargetUserID, ShouldEqual, 456)
			So(event.ResourceType, ShouldEqual, "merchant_user")
			So(event.Action, ShouldEqual, "create")
		})
	})
}

func TestGlobalAuditFunctions(t *testing.T) {
	Convey("全局审计函数测试", t, func() {
		ctx := context.Background()
		
		Convey("全局商户用户审计函数应该存在", func() {
			// 测试函数存在性
			var logMerchantUserLogin func(context.Context, uint64, uint64, uint64, string, string, interface{}) = LogMerchantUserLogin
			var logMerchantUserLogout func(context.Context, uint64, uint64, uint64, string, interface{}) = LogMerchantUserLogout
			var logMerchantUserCreate func(context.Context, uint64, uint64, uint64, uint64, string, interface{}) = LogMerchantUserCreate
			var logMerchantUserUpdate func(context.Context, uint64, uint64, uint64, uint64, string, map[string]interface{}) = LogMerchantUserUpdate
			var logMerchantUserStatusChange func(context.Context, uint64, uint64, uint64, uint64, string, string, string, interface{}) = LogMerchantUserStatusChange
			var logMerchantUserPasswordReset func(context.Context, uint64, uint64, uint64, uint64, string, string) = LogMerchantUserPasswordReset
			var logMerchantUserDelete func(context.Context, uint64, uint64, uint64, uint64, string, interface{}) = LogMerchantUserDelete
			var logMerchantOperation func(context.Context, uint64, uint64, uint64, string, string, string, interface{}) = LogMerchantOperation
			
			So(logMerchantUserLogin, ShouldNotBeNil)
			So(logMerchantUserLogout, ShouldNotBeNil)
			So(logMerchantUserCreate, ShouldNotBeNil)
			So(logMerchantUserUpdate, ShouldNotBeNil)
			So(logMerchantUserStatusChange, ShouldNotBeNil)
			So(logMerchantUserPasswordReset, ShouldNotBeNil)
			So(logMerchantUserDelete, ShouldNotBeNil)
			So(logMerchantOperation, ShouldNotBeNil)
		})
		
		Convey("全局审计函数应该可以正常调用", func() {
			So(func() {
				LogMerchantUserLogin(ctx, 1, 1, 1, "test_user", "192.168.1.1", nil)
				LogMerchantUserCreate(ctx, 1, 1, 1, 2, "new_user", nil)
				LogMerchantUserPasswordReset(ctx, 1, 1, 1, 2, "target_user", "admin_reset")
				LogMerchantOperation(ctx, 1, 1, 1, "product", "create", "创建产品", nil)
			}, ShouldNotPanic)
		})
	})
}

func TestGetMerchantUserAuditLogs(t *testing.T) {
	Convey("获取商户用户审计日志测试", t, func() {
		ctx := context.Background()
		
		Convey("应该能够获取商户用户审计日志", func() {
			logs, total, err := GetMerchantUserAuditLogs(ctx, 1, nil, nil, nil, nil, 1, 10)
			
			So(err, ShouldBeNil)
			So(logs, ShouldNotBeNil)
			So(total, ShouldBeGreaterThanOrEqualTo, 0)
			So(len(logs), ShouldBeLessThanOrEqualTo, 10)
		})
		
		Convey("应该支持分页查询", func() {
			// 第一页
			logs1, total1, err1 := GetMerchantUserAuditLogs(ctx, 1, nil, nil, nil, nil, 1, 1)
			So(err1, ShouldBeNil)
			So(total1, ShouldBeGreaterThanOrEqualTo, 0)
			
			// 第二页
			logs2, total2, err2 := GetMerchantUserAuditLogs(ctx, 1, nil, nil, nil, nil, 2, 1)
			So(err2, ShouldBeNil)
			So(total2, ShouldEqual, total1) // 总数应该相同
			
			// 如果有多条记录，页面内容应该不同
			if total1 > 1 {
				So(logs1, ShouldNotResemble, logs2)
			}
		})
		
		Convey("应该支持按用户ID筛选", func() {
			userID := uint64(1)
			logs, total, err := GetMerchantUserAuditLogs(ctx, 1, &userID, nil, nil, nil, 1, 10)
			
			So(err, ShouldBeNil)
			So(logs, ShouldNotBeNil)
			So(total, ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// AuditEventType 审计事件类型
type AuditEventType string

const (
	EventCrossTenantAttempt AuditEventType = "cross_tenant_attempt"
	EventTenantAccess       AuditEventType = "tenant_access"
	EventDataQuery          AuditEventType = "data_query"
	EventSecurityViolation  AuditEventType = "security_violation"
	// 商户用户相关事件
	EventMerchantUserLogin     AuditEventType = "merchant_user_login"
	EventMerchantUserLogout    AuditEventType = "merchant_user_logout"
	EventMerchantUserCreate    AuditEventType = "merchant_user_create"
	EventMerchantUserUpdate    AuditEventType = "merchant_user_update"
	EventMerchantUserDelete    AuditEventType = "merchant_user_delete"
	EventMerchantUserDisable   AuditEventType = "merchant_user_disable"
	EventMerchantUserEnable    AuditEventType = "merchant_user_enable"
	EventMerchantUserPassword  AuditEventType = "merchant_user_password"
	EventMerchantOperation     AuditEventType = "merchant_operation"
	// 资金相关事件
	EventFundDeposit           AuditEventType = "fund_deposit"
	EventFundBatchDeposit      AuditEventType = "fund_batch_deposit"
	EventFundAllocate          AuditEventType = "fund_allocate"
	EventFundFreeze            AuditEventType = "fund_freeze"
	EventFundUnfreeze          AuditEventType = "fund_unfreeze"
	EventFundBalanceQuery      AuditEventType = "fund_balance_query"
	EventFundTransactionQuery  AuditEventType = "fund_transaction_query"
)

// AuditSeverity 审计事件严重程度
type AuditSeverity string

const (
	SeverityInfo     AuditSeverity = "info"
	SeverityWarning  AuditSeverity = "warning"
	SeverityError    AuditSeverity = "error"
	SeverityCritical AuditSeverity = "critical"
)

// AuditEvent 审计事件
type AuditEvent struct {
	EventType       AuditEventType `json:"event_type"`
	Severity        AuditSeverity  `json:"severity"`
	TenantID        uint64         `json:"tenant_id"`
	UserID          uint64         `json:"user_id,omitempty"`
	MerchantID      *uint64        `json:"merchant_id,omitempty"`    // 商户ID
	TargetUserID    *uint64        `json:"target_user_id,omitempty"` // 目标用户ID（如商户用户管理时的被操作用户）
	ResourceType    string         `json:"resource_type"`
	ResourceID      string         `json:"resource_id,omitempty"`
	Action          string         `json:"action"`
	RequestedTenant uint64         `json:"requested_tenant,omitempty"`
	IPAddress       string         `json:"ip_address,omitempty"`
	UserAgent       string         `json:"user_agent,omitempty"`
	Message         string         `json:"message"`
	Details         interface{}    `json:"details,omitempty"`
	Timestamp       time.Time      `json:"timestamp"`
}

// AuditLogger 审计日志器
type AuditLogger struct {
}

// NewAuditLogger 创建审计日志器
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

// LogCrossTenantAttempt 记录跨租户访问尝试
func (l *AuditLogger) LogCrossTenantAttempt(ctx context.Context, userTenantID, requestedTenantID uint64, resourceType, action string, details interface{}) {
	event := AuditEvent{
		EventType:       EventCrossTenantAttempt,
		Severity:        SeverityCritical,
		TenantID:        userTenantID,
		UserID:          l.getUserID(ctx),
		ResourceType:    resourceType,
		Action:          action,
		RequestedTenant: requestedTenantID,
		IPAddress:       l.getIPAddress(ctx),
		UserAgent:       l.getUserAgent(ctx),
		Message:         "跨租户访问尝试被阻止",
		Details:         details,
		Timestamp:       time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogTenantAccess 记录正常租户访问
func (l *AuditLogger) LogTenantAccess(ctx context.Context, tenantID uint64, resourceType, action string, details interface{}) {
	event := AuditEvent{
		EventType:    EventTenantAccess,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       l.getUserID(ctx),
		ResourceType: resourceType,
		Action:       action,
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      "租户数据访问",
		Details:      details,
		Timestamp:    time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogDataQuery 记录数据查询操作
func (l *AuditLogger) LogDataQuery(ctx context.Context, tenantID uint64, resourceType, query string, rowCount int) {
	event := AuditEvent{
		EventType:    EventDataQuery,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       l.getUserID(ctx),
		ResourceType: resourceType,
		Action:       "query",
		Message:      "数据查询操作",
		Details: map[string]interface{}{
			"query":     query,
			"row_count": rowCount,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogSecurityViolation 记录安全违规事件
func (l *AuditLogger) LogSecurityViolation(ctx context.Context, tenantID uint64, violationType, message string, details interface{}) {
	event := AuditEvent{
		EventType:    EventSecurityViolation,
		Severity:     SeverityError,
		TenantID:     tenantID,
		UserID:       l.getUserID(ctx),
		ResourceType: "security",
		Action:       violationType,
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      message,
		Details:      details,
		Timestamp:    time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogMerchantUserLogin 记录商户用户登录
func (l *AuditLogger) LogMerchantUserLogin(ctx context.Context, tenantID, merchantID, userID uint64, username, ipAddress string, details interface{}) {
	event := AuditEvent{
		EventType:    EventMerchantUserLogin,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       userID,
		MerchantID:   &merchantID,
		ResourceType: "merchant_user",
		Action:       "login",
		IPAddress:    ipAddress,
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("商户用户 %s 登录成功", username),
		Details:      details,
		Timestamp:    time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogMerchantUserLogout 记录商户用户登出
func (l *AuditLogger) LogMerchantUserLogout(ctx context.Context, tenantID, merchantID, userID uint64, username string, details interface{}) {
	event := AuditEvent{
		EventType:    EventMerchantUserLogout,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       userID,
		MerchantID:   &merchantID,
		ResourceType: "merchant_user",
		Action:       "logout",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("商户用户 %s 登出", username),
		Details:      details,
		Timestamp:    time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogMerchantUserCreate 记录商户用户创建
func (l *AuditLogger) LogMerchantUserCreate(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername string, details interface{}) {
	event := AuditEvent{
		EventType:    EventMerchantUserCreate,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       operatorUserID,
		MerchantID:   &merchantID,
		TargetUserID: &targetUserID,
		ResourceType: "merchant_user",
		ResourceID:   fmt.Sprintf("%d", targetUserID),
		Action:       "create",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("创建商户用户 %s", targetUsername),
		Details:      details,
		Timestamp:    time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogMerchantUserUpdate 记录商户用户更新
func (l *AuditLogger) LogMerchantUserUpdate(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername string, changes map[string]interface{}) {
	event := AuditEvent{
		EventType:    EventMerchantUserUpdate,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       operatorUserID,
		MerchantID:   &merchantID,
		TargetUserID: &targetUserID,
		ResourceType: "merchant_user",
		ResourceID:   fmt.Sprintf("%d", targetUserID),
		Action:       "update",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("更新商户用户 %s", targetUsername),
		Details: map[string]interface{}{
			"changes": changes,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogMerchantUserStatusChange 记录商户用户状态变更
func (l *AuditLogger) LogMerchantUserStatusChange(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername, oldStatus, newStatus string, details interface{}) {
	var eventType AuditEventType
	var message string
	
	switch newStatus {
	case "active":
		eventType = EventMerchantUserEnable
		message = fmt.Sprintf("启用商户用户 %s", targetUsername)
	case "inactive":
		eventType = EventMerchantUserDisable
		message = fmt.Sprintf("禁用商户用户 %s", targetUsername)
	default:
		eventType = EventMerchantUserUpdate
		message = fmt.Sprintf("变更商户用户 %s 状态: %s -> %s", targetUsername, oldStatus, newStatus)
	}

	event := AuditEvent{
		EventType:    eventType,
		Severity:     SeverityWarning, // 状态变更为警告级别
		TenantID:     tenantID,
		UserID:       operatorUserID,
		MerchantID:   &merchantID,
		TargetUserID: &targetUserID,
		ResourceType: "merchant_user",
		ResourceID:   fmt.Sprintf("%d", targetUserID),
		Action:       "status_change",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      message,
		Details: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": newStatus,
			"details":    details,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogMerchantUserPasswordReset 记录商户用户密码重置
func (l *AuditLogger) LogMerchantUserPasswordReset(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername string, resetMethod string) {
	event := AuditEvent{
		EventType:    EventMerchantUserPassword,
		Severity:     SeverityWarning, // 密码重置为警告级别
		TenantID:     tenantID,
		UserID:       operatorUserID,
		MerchantID:   &merchantID,
		TargetUserID: &targetUserID,
		ResourceType: "merchant_user",
		ResourceID:   fmt.Sprintf("%d", targetUserID),
		Action:       "password_reset",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("重置商户用户 %s 密码", targetUsername),
		Details: map[string]interface{}{
			"reset_method": resetMethod,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogMerchantUserDelete 记录商户用户删除
func (l *AuditLogger) LogMerchantUserDelete(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername string, details interface{}) {
	event := AuditEvent{
		EventType:    EventMerchantUserDelete,
		Severity:     SeverityError, // 删除操作为错误级别
		TenantID:     tenantID,
		UserID:       operatorUserID,
		MerchantID:   &merchantID,
		TargetUserID: &targetUserID,
		ResourceType: "merchant_user",
		ResourceID:   fmt.Sprintf("%d", targetUserID),
		Action:       "delete",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("删除商户用户 %s", targetUsername),
		Details:      details,
		Timestamp:    time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogMerchantOperation 记录商户业务操作
func (l *AuditLogger) LogMerchantOperation(ctx context.Context, tenantID, merchantID, userID uint64, resourceType, action, message string, details interface{}) {
	event := AuditEvent{
		EventType:    EventMerchantOperation,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       userID,
		MerchantID:   &merchantID,
		ResourceType: resourceType,
		Action:       action,
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      message,
		Details:      details,
		Timestamp:    time.Now(),
	}

	l.logEvent(ctx, event)
}

// logEvent 记录审计事件
func (l *AuditLogger) logEvent(ctx context.Context, event AuditEvent) {
	// 序列化事件为JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		g.Log().Errorf(ctx, "审计事件序列化失败: %v", err)
		return
	}

	// 根据事件严重程度选择日志级别
	switch event.Severity {
	case SeverityInfo:
		g.Log().Info(ctx, "AUDIT: ", string(eventJSON))
	case SeverityWarning:
		g.Log().Warning(ctx, "AUDIT: ", string(eventJSON))
	case SeverityError:
		g.Log().Error(ctx, "AUDIT: ", string(eventJSON))
	case SeverityCritical:
		g.Log().Critical(ctx, "AUDIT: ", string(eventJSON))
	default:
		g.Log().Info(ctx, "AUDIT: ", string(eventJSON))
	}

	// 如果是关键事件，也发送到监控系统
	if event.Severity == SeverityCritical {
		l.sendToMonitoring(ctx, event)
	}
}

// getUserID 从上下文获取用户ID
func (l *AuditLogger) getUserID(ctx context.Context) uint64 {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(uint64); ok {
			return id
		}
	}
	return 0
}

// getMerchantID 从上下文获取商户ID
func (l *AuditLogger) getMerchantID(ctx context.Context) *uint64 {
	if merchantID := ctx.Value("merchant_id"); merchantID != nil {
		if id, ok := merchantID.(uint64); ok {
			return &id
		}
	}
	return nil
}

// getIPAddress 从上下文获取IP地址
func (l *AuditLogger) getIPAddress(ctx context.Context) string {
	if ip := ctx.Value("client_ip"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}
	return ""
}

// getUserAgent 从上下文获取User-Agent
func (l *AuditLogger) getUserAgent(ctx context.Context) string {
	if ua := ctx.Value("user_agent"); ua != nil {
		if uaStr, ok := ua.(string); ok {
			return uaStr
		}
	}
	return ""
}

// sendToMonitoring 发送关键事件到监控系统
func (l *AuditLogger) sendToMonitoring(ctx context.Context, event AuditEvent) {
	// 这里可以集成到监控系统，如Prometheus、Grafana、钉钉告警等
	// 目前仅记录日志
	g.Log().Critical(ctx, "CRITICAL_SECURITY_EVENT: %+v", event)
}

// 全局审计日志器实例
var defaultAuditLogger = NewAuditLogger()

// LogCrossTenantAttempt 全局函数：记录跨租户访问尝试
func LogCrossTenantAttempt(ctx context.Context, userTenantID, requestedTenantID uint64, resourceType, action string, details interface{}) {
	defaultAuditLogger.LogCrossTenantAttempt(ctx, userTenantID, requestedTenantID, resourceType, action, details)
}

// LogTenantAccess 全局函数：记录正常租户访问
func LogTenantAccess(ctx context.Context, tenantID uint64, resourceType, action string, details interface{}) {
	defaultAuditLogger.LogTenantAccess(ctx, tenantID, resourceType, action, details)
}

// LogDataQuery 全局函数：记录数据查询操作
func LogDataQuery(ctx context.Context, tenantID uint64, resourceType, query string, rowCount int) {
	defaultAuditLogger.LogDataQuery(ctx, tenantID, resourceType, query, rowCount)
}

// LogSecurityViolation 全局函数：记录安全违规事件
func LogSecurityViolation(ctx context.Context, tenantID uint64, violationType, message string, details interface{}) {
	defaultAuditLogger.LogSecurityViolation(ctx, tenantID, violationType, message, details)
}

// LogOperation 记录一般操作
func LogOperation(ctx context.Context, resourceType, action string, details interface{}) {
	// 从上下文获取租户ID和用户ID
	var tenantID uint64
	var userID uint64
	
	if tid := ctx.Value("tenant_id"); tid != nil {
		if id, ok := tid.(uint64); ok {
			tenantID = id
		}
	}
	
	if uid := ctx.Value("user_id"); uid != nil {
		if id, ok := uid.(uint64); ok {
			userID = id
		}
	}

	event := AuditEvent{
		EventType:    EventTenantAccess,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       userID,
		ResourceType: resourceType,
		Action:       action,
		IPAddress:    defaultAuditLogger.getIPAddress(ctx),
		UserAgent:    defaultAuditLogger.getUserAgent(ctx),
		Message:      fmt.Sprintf("%s %s 操作", resourceType, action),
		Details:      details,
		Timestamp:    time.Now(),
	}

	defaultAuditLogger.logEvent(ctx, event)
}

// GetOperationLogs 获取操作日志（模拟实现）
func GetOperationLogs(ctx context.Context, resourceType, resourceID string) ([]g.Map, error) {
	// 这里应该从数据库或日志系统查询审计日志
	// 现在返回模拟数据用于开发
	return []g.Map{
		{
			"id":            1,
			"resource_type": resourceType,
			"resource_id":   resourceID,
			"action":        "register",
			"user_id":       1,
			"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
			"details":       "商户注册申请",
		},
	}, nil
}

// 商户用户审计全局函数

// LogMerchantUserLogin 全局函数：记录商户用户登录
func LogMerchantUserLogin(ctx context.Context, tenantID, merchantID, userID uint64, username, ipAddress string, details interface{}) {
	defaultAuditLogger.LogMerchantUserLogin(ctx, tenantID, merchantID, userID, username, ipAddress, details)
}

// LogMerchantUserLogout 全局函数：记录商户用户登出
func LogMerchantUserLogout(ctx context.Context, tenantID, merchantID, userID uint64, username string, details interface{}) {
	defaultAuditLogger.LogMerchantUserLogout(ctx, tenantID, merchantID, userID, username, details)
}

// LogMerchantUserCreate 全局函数：记录商户用户创建
func LogMerchantUserCreate(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername string, details interface{}) {
	defaultAuditLogger.LogMerchantUserCreate(ctx, tenantID, merchantID, operatorUserID, targetUserID, targetUsername, details)
}

// LogMerchantUserUpdate 全局函数：记录商户用户更新
func LogMerchantUserUpdate(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername string, changes map[string]interface{}) {
	defaultAuditLogger.LogMerchantUserUpdate(ctx, tenantID, merchantID, operatorUserID, targetUserID, targetUsername, changes)
}

// LogMerchantUserStatusChange 全局函数：记录商户用户状态变更
func LogMerchantUserStatusChange(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername, oldStatus, newStatus string, details interface{}) {
	defaultAuditLogger.LogMerchantUserStatusChange(ctx, tenantID, merchantID, operatorUserID, targetUserID, targetUsername, oldStatus, newStatus, details)
}

// LogMerchantUserPasswordReset 全局函数：记录商户用户密码重置
func LogMerchantUserPasswordReset(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername string, resetMethod string) {
	defaultAuditLogger.LogMerchantUserPasswordReset(ctx, tenantID, merchantID, operatorUserID, targetUserID, targetUsername, resetMethod)
}

// LogMerchantUserDelete 全局函数：记录商户用户删除
func LogMerchantUserDelete(ctx context.Context, tenantID, merchantID, operatorUserID, targetUserID uint64, targetUsername string, details interface{}) {
	defaultAuditLogger.LogMerchantUserDelete(ctx, tenantID, merchantID, operatorUserID, targetUserID, targetUsername, details)
}

// LogMerchantOperation 全局函数：记录商户业务操作
func LogMerchantOperation(ctx context.Context, tenantID, merchantID, userID uint64, resourceType, action, message string, details interface{}) {
	defaultAuditLogger.LogMerchantOperation(ctx, tenantID, merchantID, userID, resourceType, action, message, details)
}

// GetMerchantUserAuditLogs 获取商户用户审计日志（新增）
func GetMerchantUserAuditLogs(ctx context.Context, merchantID uint64, userID *uint64, eventTypes []AuditEventType, startTime, endTime *time.Time, page, pageSize int) ([]g.Map, int, error) {
	// 这里应该从数据库或日志系统查询商户用户审计日志
	// 现在返回模拟数据用于开发
	logs := []g.Map{
		{
			"id":            1,
			"event_type":    "merchant_user_login",
			"severity":      "info",
			"tenant_id":     1,
			"user_id":       userID,
			"merchant_id":   merchantID,
			"resource_type": "merchant_user",
			"action":        "login",
			"message":       "商户用户登录成功",
			"ip_address":    "192.168.1.100",
			"timestamp":     time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05"),
			"details": map[string]interface{}{
				"user_agent": "Mozilla/5.0 Browser",
			},
		},
		{
			"id":            2,
			"event_type":    "merchant_user_create",
			"severity":      "info",
			"tenant_id":     1,
			"user_id":       1,
			"merchant_id":   merchantID,
			"target_user_id": 2,
			"resource_type": "merchant_user",
			"resource_id":   "2",
			"action":        "create",
			"message":       "创建商户用户 test_operator",
			"ip_address":    "192.168.1.100",
			"timestamp":     time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
			"details": map[string]interface{}{
				"role": "merchant_operator",
			},
		},
	}

	// 模拟分页
	total := len(logs)
	if page > 1 {
		start := (page - 1) * pageSize
		if start >= total {
			return []g.Map{}, total, nil
		}
		end := start + pageSize
		if end > total {
			end = total
		}
		logs = logs[start:end]
	} else if pageSize < total {
		logs = logs[:pageSize]
	}

	return logs, total, nil
}

// 资金操作审计方法

// LogFundDeposit 记录资金充值操作
func (l *AuditLogger) LogFundDeposit(ctx context.Context, tenantID, merchantID, operatorID uint64, amount float64, currency string, fundID uint64, details interface{}) {
	event := AuditEvent{
		EventType:    EventFundDeposit,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       operatorID,
		MerchantID:   &merchantID,
		ResourceType: "fund",
		ResourceID:   fmt.Sprintf("%d", fundID),
		Action:       "deposit",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("为商户ID:%d充值%s%.2f", merchantID, currency, amount),
		Details: map[string]interface{}{
			"amount":     amount,
			"currency":   currency,
			"fund_id":    fundID,
			"details":    details,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogFundBatchDeposit 记录批量资金充值操作
func (l *AuditLogger) LogFundBatchDeposit(ctx context.Context, tenantID, operatorID uint64, batchCount int, totalAmount float64, fundIDs []uint64, details interface{}) {
	event := AuditEvent{
		EventType:    EventFundBatchDeposit,
		Severity:     SeverityWarning, // 批量操作使用警告级别
		TenantID:     tenantID,
		UserID:       operatorID,
		ResourceType: "fund",
		Action:       "batch_deposit",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("批量充值操作：%d笔充值，总金额%.2f", batchCount, totalAmount),
		Details: map[string]interface{}{
			"batch_count":  batchCount,
			"total_amount": totalAmount,
			"fund_ids":     fundIDs,
			"details":      details,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogFundAllocate 记录权益分配操作
func (l *AuditLogger) LogFundAllocate(ctx context.Context, tenantID, merchantID, operatorID uint64, amount float64, fundID uint64, details interface{}) {
	event := AuditEvent{
		EventType:    EventFundAllocate,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       operatorID,
		MerchantID:   &merchantID,
		ResourceType: "fund",
		ResourceID:   fmt.Sprintf("%d", fundID),
		Action:       "allocate",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("为商户ID:%d分配权益%.2f", merchantID, amount),
		Details: map[string]interface{}{
			"amount":  amount,
			"fund_id": fundID,
			"details": details,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogFundFreeze 记录资金冻结操作
func (l *AuditLogger) LogFundFreeze(ctx context.Context, tenantID, merchantID, operatorID uint64, amount float64, reason string, details interface{}) {
	event := AuditEvent{
		EventType:    EventFundFreeze,
		Severity:     SeverityWarning, // 冻结操作使用警告级别
		TenantID:     tenantID,
		UserID:       operatorID,
		MerchantID:   &merchantID,
		ResourceType: "fund",
		Action:       "freeze",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("冻结商户ID:%d权益%.2f，原因：%s", merchantID, amount, reason),
		Details: map[string]interface{}{
			"amount":  amount,
			"reason":  reason,
			"details": details,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogFundUnfreeze 记录资金解冻操作
func (l *AuditLogger) LogFundUnfreeze(ctx context.Context, tenantID, merchantID, operatorID uint64, amount float64, reason string, details interface{}) {
	event := AuditEvent{
		EventType:    EventFundUnfreeze,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       operatorID,
		MerchantID:   &merchantID,
		ResourceType: "fund",
		Action:       "unfreeze",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("解冻商户ID:%d权益%.2f，原因：%s", merchantID, amount, reason),
		Details: map[string]interface{}{
			"amount":  amount,
			"reason":  reason,
			"details": details,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogFundBalanceQuery 记录余额查询操作
func (l *AuditLogger) LogFundBalanceQuery(ctx context.Context, tenantID, merchantID, operatorID uint64, balance *map[string]interface{}) {
	event := AuditEvent{
		EventType:    EventFundBalanceQuery,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       operatorID,
		MerchantID:   &merchantID,
		ResourceType: "fund",
		Action:       "balance_query",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("查询商户ID:%d权益余额", merchantID),
		Details: map[string]interface{}{
			"balance": balance,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// LogFundTransactionQuery 记录交易记录查询操作
func (l *AuditLogger) LogFundTransactionQuery(ctx context.Context, tenantID, operatorID uint64, query map[string]interface{}, resultCount int) {
	event := AuditEvent{
		EventType:    EventFundTransactionQuery,
		Severity:     SeverityInfo,
		TenantID:     tenantID,
		UserID:       operatorID,
		ResourceType: "fund",
		Action:       "transaction_query",
		IPAddress:    l.getIPAddress(ctx),
		UserAgent:    l.getUserAgent(ctx),
		Message:      fmt.Sprintf("查询资金交易记录，返回%d条结果", resultCount),
		Details: map[string]interface{}{
			"query":        query,
			"result_count": resultCount,
		},
		Timestamp: time.Now(),
	}

	l.logEvent(ctx, event)
}

// 资金操作审计全局函数

// LogFundDeposit 全局函数：记录资金充值操作
func LogFundDeposit(ctx context.Context, tenantID, merchantID, operatorID uint64, amount float64, currency string, fundID uint64, details interface{}) {
	defaultAuditLogger.LogFundDeposit(ctx, tenantID, merchantID, operatorID, amount, currency, fundID, details)
}

// LogFundBatchDeposit 全局函数：记录批量资金充值操作
func LogFundBatchDeposit(ctx context.Context, tenantID, operatorID uint64, batchCount int, totalAmount float64, fundIDs []uint64, details interface{}) {
	defaultAuditLogger.LogFundBatchDeposit(ctx, tenantID, operatorID, batchCount, totalAmount, fundIDs, details)
}

// LogFundAllocate 全局函数：记录权益分配操作
func LogFundAllocate(ctx context.Context, tenantID, merchantID, operatorID uint64, amount float64, fundID uint64, details interface{}) {
	defaultAuditLogger.LogFundAllocate(ctx, tenantID, merchantID, operatorID, amount, fundID, details)
}

// LogFundFreeze 全局函数：记录资金冻结操作
func LogFundFreeze(ctx context.Context, tenantID, merchantID, operatorID uint64, amount float64, reason string, details interface{}) {
	defaultAuditLogger.LogFundFreeze(ctx, tenantID, merchantID, operatorID, amount, reason, details)
}

// LogFundUnfreeze 全局函数：记录资金解冻操作
func LogFundUnfreeze(ctx context.Context, tenantID, merchantID, operatorID uint64, amount float64, reason string, details interface{}) {
	defaultAuditLogger.LogFundUnfreeze(ctx, tenantID, merchantID, operatorID, amount, reason, details)
}

// LogFundBalanceQuery 全局函数：记录余额查询操作
func LogFundBalanceQuery(ctx context.Context, tenantID, merchantID, operatorID uint64, balance *map[string]interface{}) {
	defaultAuditLogger.LogFundBalanceQuery(ctx, tenantID, merchantID, operatorID, balance)
}

// LogFundTransactionQuery 全局函数：记录交易记录查询操作
func LogFundTransactionQuery(ctx context.Context, tenantID, operatorID uint64, query map[string]interface{}, resultCount int) {
	defaultAuditLogger.LogFundTransactionQuery(ctx, tenantID, operatorID, query, resultCount)
}

// GetFundAuditLogs 获取资金操作审计日志
func GetFundAuditLogs(ctx context.Context, tenantID uint64, merchantID *uint64, eventTypes []AuditEventType, startTime, endTime *time.Time, page, pageSize int) ([]g.Map, int, error) {
	// 这里应该从数据库或日志系统查询资金审计日志
	// 现在返回模拟数据用于开发
	logs := []g.Map{
		{
			"id":            1,
			"event_type":    "fund_deposit",
			"severity":      "info",
			"tenant_id":     tenantID,
			"user_id":       1,
			"merchant_id":   merchantID,
			"resource_type": "fund",
			"resource_id":   "1001",
			"action":        "deposit",
			"message":       "为商户ID:1充值CNY1000.00",
			"ip_address":    "192.168.1.100",
			"timestamp":     time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05"),
			"details": map[string]interface{}{
				"amount":   1000.00,
				"currency": "CNY",
				"fund_id":  1001,
			},
		},
		{
			"id":            2,
			"event_type":    "fund_allocate",
			"severity":      "info",
			"tenant_id":     tenantID,
			"user_id":       1,
			"merchant_id":   merchantID,
			"resource_type": "fund",
			"resource_id":   "1002",
			"action":        "allocate",
			"message":       "为商户ID:1分配权益500.00",
			"ip_address":    "192.168.1.100",
			"timestamp":     time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
			"details": map[string]interface{}{
				"amount":  500.00,
				"fund_id": 1002,
			},
		},
		{
			"id":            3,
			"event_type":    "fund_freeze",
			"severity":      "warning",
			"tenant_id":     tenantID,
			"user_id":       1,
			"merchant_id":   merchantID,
			"resource_type": "fund",
			"action":        "freeze",
			"message":       "冻结商户ID:1权益200.00，原因：风险控制",
			"ip_address":    "192.168.1.100",
			"timestamp":     time.Now().Add(-30 * time.Minute).Format("2006-01-02 15:04:05"),
			"details": map[string]interface{}{
				"amount": 200.00,
				"reason": "风险控制",
			},
		},
	}

	// 模拟筛选
	if merchantID != nil {
		filteredLogs := []g.Map{}
		for _, log := range logs {
			if log["merchant_id"] == merchantID {
				filteredLogs = append(filteredLogs, log)
			}
		}
		logs = filteredLogs
	}

	// 模拟分页
	total := len(logs)
	if page > 1 {
		start := (page - 1) * pageSize
		if start >= total {
			return []g.Map{}, total, nil
		}
		end := start + pageSize
		if end > total {
			end = total
		}
		logs = logs[start:end]
	} else if pageSize < total {
		logs = logs[:pageSize]
	}

	return logs, total, nil
}
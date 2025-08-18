package audit

import (
	"context"
	"encoding/json"
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
	EventType     AuditEventType `json:"event_type"`
	Severity      AuditSeverity  `json:"severity"`
	TenantID      uint64         `json:"tenant_id"`
	UserID        uint64         `json:"user_id,omitempty"`
	ResourceType  string         `json:"resource_type"`
	ResourceID    string         `json:"resource_id,omitempty"`
	Action        string         `json:"action"`
	RequestedTenant uint64       `json:"requested_tenant,omitempty"`
	IPAddress     string         `json:"ip_address,omitempty"`
	UserAgent     string         `json:"user_agent,omitempty"`
	Message       string         `json:"message"`
	Details       interface{}    `json:"details,omitempty"`
	Timestamp     time.Time      `json:"timestamp"`
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
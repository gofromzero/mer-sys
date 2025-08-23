package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AuditEvent 审计事件结构
type AuditEvent struct {
	Timestamp   time.Time   `json:"timestamp"`
	TenantID    interface{} `json:"tenant_id"`
	UserID      interface{} `json:"user_id"`
	Action      string      `json:"action"`
	Resource    string      `json:"resource"`
	ResourceID  string      `json:"resource_id,omitempty"`
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	ClientIP    string      `json:"client_ip"`
	UserAgent   string      `json:"user_agent"`
	StatusCode  int         `json:"status_code,omitempty"`
	RequestBody string      `json:"request_body,omitempty"`
	Changes     interface{} `json:"changes,omitempty"`
	Success     bool        `json:"success"`
	ErrorMsg    string      `json:"error_msg,omitempty"`
}

// AuditLog 审计日志中间件
func AuditLog(resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		startTime := time.Now()
		
		// 创建审计事件
		event := &AuditEvent{
			Timestamp:  startTime,
			TenantID:   r.GetCtx().Value("tenant_id"),
			UserID:     r.GetCtx().Value("user_id"),
			Action:     getActionFromMethod(r.Method),
			Resource:   resource,
			ResourceID: r.Get("id").String(),
			Method:     r.Method,
			Path:       r.URL.Path,
			ClientIP:   r.GetClientIp(),
			UserAgent:  r.Header.Get("User-Agent"),
		}

		// 记录请求体（仅对写操作）
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if bodyBytes := r.GetBody(); len(bodyBytes) > 0 && len(bodyBytes) < 10240 { // 限制10KB
				event.RequestBody = string(bodyBytes)
			}
		}

		// 设置审计事件到上下文，供业务逻辑使用
		r.SetCtx(context.WithValue(r.GetCtx(), "audit_event", event))

		// 执行请求
		r.Middleware.Next()

		// 完成审计记录
		event.StatusCode = r.Response.Status
		event.Success = r.Response.Status >= 200 && r.Response.Status < 400

		// 从上下文获取业务变更信息
		if changes := r.GetCtx().Value("audit_changes"); changes != nil {
			event.Changes = changes
		}

		// 记录错误信息
		if !event.Success {
			event.ErrorMsg = r.Response.BufferString()
		}

		// 异步写入审计日志
		go writeAuditLog(event)
	}
}

// PricingAuditLog 定价操作专用审计日志
func PricingAuditLog() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		// 使用基础审计中间件
		auditMiddleware := AuditLog("pricing")
		auditMiddleware(r)
	}
}

// RecordPriceChange 记录价格变更审计信息
func RecordPriceChange(r *ghttp.Request, oldPrice, newPrice interface{}, reason string) {
	changes := map[string]interface{}{
		"old_price":     oldPrice,
		"new_price":     newPrice,
		"change_reason": reason,
		"change_time":   time.Now(),
	}
	
	// 将变更信息添加到上下文
	r.SetCtx(context.WithValue(r.GetCtx(), "audit_changes", changes))
}

// getActionFromMethod 从HTTP方法推断操作类型
func getActionFromMethod(method string) string {
	switch method {
	case "GET":
		return "READ"
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// writeAuditLog 写入审计日志
func writeAuditLog(event *AuditEvent) {
	// 将审计事件序列化为JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		g.Log().Error(nil, "审计日志序列化失败", err)
		return
	}

	// 写入审计日志（可以是文件、数据库、消息队列等）
	g.Log().Info(nil, "AUDIT", string(eventJSON))

	// TODO: 可以扩展为写入专门的审计数据库表
	// auditRepo := repository.NewAuditRepository()
	// auditRepo.Create(context.Background(), event)
}
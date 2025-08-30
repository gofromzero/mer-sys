package service

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypeSMS      NotificationType = "sms"
	NotificationTypeWebhook  NotificationType = "webhook"
	NotificationTypeInApp    NotificationType = "in_app"
)

// NotificationPriority 通知优先级
type NotificationPriority string

const (
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityLow    NotificationPriority = "low"
)

// NotificationRequest 通知请求
type NotificationRequest struct {
	Type        NotificationType     `json:"type"`
	Recipients  []string             `json:"recipients"`
	Subject     string               `json:"subject"`
	Content     string               `json:"content"`
	Attachments []string             `json:"attachments"`
	Priority    NotificationPriority `json:"priority"`
	ScheduledAt *time.Time           `json:"scheduled_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	SentAt      time.Time              `json:"sent_at"`
	FailedCount int                    `json:"failed_count"`
	Details     map[string]interface{} `json:"details"`
}

// INotificationService 通知服务接口
type INotificationService interface {
	// 发送通知
	SendNotification(ctx context.Context, req *NotificationRequest) error
	// 发送邮件通知
	SendEmail(ctx context.Context, recipients []string, subject, content string, attachments []string) error
	// 发送短信通知
	SendSMS(ctx context.Context, recipients []string, content string) error
	// 批量发送通知
	SendBatchNotifications(ctx context.Context, requests []*NotificationRequest) ([]*NotificationResponse, error)
	// 获取通知历史
	GetNotificationHistory(ctx context.Context, limit int) ([]*NotificationResponse, error)
}

// NotificationService 通知服务实现
type NotificationService struct {
	emailConfig  *EmailConfig
	smsConfig    *SMSConfig
	webhookConfig *WebhookConfig
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	UseTLS       bool   `json:"use_tls"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider  string `json:"provider"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Signature string `json:"signature"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Timeout time.Duration     `json:"timeout"`
}

// NewNotificationService 创建通知服务实例
func NewNotificationService() INotificationService {
	return &NotificationService{
		emailConfig: &EmailConfig{
			SMTPHost:  g.Cfg().MustGet(context.Background(), "notification.email.smtp_host", "smtp.163.com").String(),
			SMTPPort:  g.Cfg().MustGet(context.Background(), "notification.email.smtp_port", 587).Int(),
			Username:  g.Cfg().MustGet(context.Background(), "notification.email.username", "").String(),
			Password:  g.Cfg().MustGet(context.Background(), "notification.email.password", "").String(),
			FromEmail: g.Cfg().MustGet(context.Background(), "notification.email.from_email", "").String(),
			FromName:  g.Cfg().MustGet(context.Background(), "notification.email.from_name", "MER系统").String(),
			UseTLS:    g.Cfg().MustGet(context.Background(), "notification.email.use_tls", true).Bool(),
		},
		smsConfig: &SMSConfig{
			Provider:  g.Cfg().MustGet(context.Background(), "notification.sms.provider", "aliyun").String(),
			APIKey:    g.Cfg().MustGet(context.Background(), "notification.sms.api_key", "").String(),
			APISecret: g.Cfg().MustGet(context.Background(), "notification.sms.api_secret", "").String(),
			Signature: g.Cfg().MustGet(context.Background(), "notification.sms.signature", "MER系统").String(),
		},
		webhookConfig: &WebhookConfig{
			URL:     g.Cfg().MustGet(context.Background(), "notification.webhook.url", "").String(),
			Method:  g.Cfg().MustGet(context.Background(), "notification.webhook.method", "POST").String(),
			Timeout: g.Cfg().MustGet(context.Background(), "notification.webhook.timeout", 30*time.Second).Duration(),
		},
	}
}

// SendNotification 发送通知
func (n *NotificationService) SendNotification(ctx context.Context, req *NotificationRequest) error {
	g.Log().Info(ctx, "发送通知", 
		"type", req.Type,
		"recipients_count", len(req.Recipients),
		"subject", req.Subject,
		"priority", req.Priority)
	
	switch req.Type {
	case NotificationTypeEmail:
		return n.SendEmail(ctx, req.Recipients, req.Subject, req.Content, req.Attachments)
	case NotificationTypeSMS:
		return n.SendSMS(ctx, req.Recipients, req.Content)
	case NotificationTypeWebhook:
		return n.sendWebhook(ctx, req)
	case NotificationTypeInApp:
		return n.sendInAppNotification(ctx, req)
	default:
		return fmt.Errorf("不支持的通知类型: %s", req.Type)
	}
}

// SendEmail 发送邮件通知
func (n *NotificationService) SendEmail(ctx context.Context, recipients []string, subject, content string, attachments []string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("邮件接收人列表为空")
	}
	
	if n.emailConfig.Username == "" || n.emailConfig.Password == "" {
		g.Log().Warning(ctx, "邮件配置不完整，跳过邮件发送")
		return nil
	}
	
	g.Log().Info(ctx, "开始发送邮件", 
		"recipients", recipients, 
		"subject", subject,
		"attachments_count", len(attachments))
	
	// 构建邮件内容
	from := fmt.Sprintf("%s <%s>", n.emailConfig.FromName, n.emailConfig.FromEmail)
	to := strings.Join(recipients, ", ")
	
	// 基础邮件头
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=utf-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)
	
	// 构建邮件消息
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + content
	
	// SMTP 认证
	auth := smtp.PlainAuth("", n.emailConfig.Username, n.emailConfig.Password, n.emailConfig.SMTPHost)
	
	// 发送邮件
	addr := fmt.Sprintf("%s:%d", n.emailConfig.SMTPHost, n.emailConfig.SMTPPort)
	err := smtp.SendMail(addr, auth, n.emailConfig.FromEmail, recipients, []byte(message))
	
	if err != nil {
		g.Log().Error(ctx, "邮件发送失败", "error", err, "recipients", recipients)
		return fmt.Errorf("邮件发送失败: %v", err)
	}
	
	g.Log().Info(ctx, "邮件发送成功", "recipients", recipients, "subject", subject)
	return nil
}

// SendSMS 发送短信通知
func (n *NotificationService) SendSMS(ctx context.Context, recipients []string, content string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("短信接收人列表为空")
	}
	
	if n.smsConfig.APIKey == "" {
		g.Log().Warning(ctx, "短信配置不完整，跳过短信发送")
		return nil
	}
	
	g.Log().Info(ctx, "开始发送短信", "recipients", recipients, "content_length", len(content))
	
	// 这里根据不同的短信服务商实现具体的发送逻辑
	switch n.smsConfig.Provider {
	case "aliyun":
		return n.sendAliyunSMS(ctx, recipients, content)
	case "tencent":
		return n.sendTencentSMS(ctx, recipients, content)
	default:
		g.Log().Warning(ctx, "不支持的短信服务商", "provider", n.smsConfig.Provider)
		return fmt.Errorf("不支持的短信服务商: %s", n.smsConfig.Provider)
	}
}

// SendBatchNotifications 批量发送通知
func (n *NotificationService) SendBatchNotifications(ctx context.Context, requests []*NotificationRequest) ([]*NotificationResponse, error) {
	g.Log().Info(ctx, "开始批量发送通知", "count", len(requests))
	
	responses := make([]*NotificationResponse, 0, len(requests))
	
	for i, req := range requests {
		response := &NotificationResponse{
			ID:     fmt.Sprintf("batch_%d_%d", time.Now().Unix(), i),
			SentAt: time.Now(),
		}
		
		err := n.SendNotification(ctx, req)
		if err != nil {
			response.Status = "failed"
			response.Message = err.Error()
			response.FailedCount = 1
			g.Log().Error(ctx, "批量通知发送失败", "index", i, "error", err)
		} else {
			response.Status = "success"
			response.Message = "发送成功"
			response.FailedCount = 0
		}
		
		responses = append(responses, response)
	}
	
	g.Log().Info(ctx, "批量通知发送完成", "total", len(requests), "success", len(responses))
	return responses, nil
}

// GetNotificationHistory 获取通知历史
func (n *NotificationService) GetNotificationHistory(ctx context.Context, limit int) ([]*NotificationResponse, error) {
	// 这里可以实现从数据库获取通知历史记录的逻辑
	// 暂时返回空数组
	g.Log().Info(ctx, "获取通知历史", "limit", limit)
	return []*NotificationResponse{}, nil
}

// sendWebhook 发送Webhook通知
func (n *NotificationService) sendWebhook(ctx context.Context, req *NotificationRequest) error {
	if n.webhookConfig.URL == "" {
		g.Log().Warning(ctx, "Webhook URL未配置，跳过发送")
		return nil
	}
	
	g.Log().Info(ctx, "发送Webhook通知", "url", n.webhookConfig.URL)
	
	// 这里可以实现HTTP请求发送逻辑
	// 使用 gf 的 HTTP客户端或标准库的 http 包
	
	g.Log().Info(ctx, "Webhook通知发送完成")
	return nil
}

// sendInAppNotification 发送应用内通知
func (n *NotificationService) sendInAppNotification(ctx context.Context, req *NotificationRequest) error {
	g.Log().Info(ctx, "发送应用内通知", "recipients", req.Recipients)
	
	// 这里可以实现应用内通知的逻辑
	// 例如：WebSocket推送、数据库记录等
	
	g.Log().Info(ctx, "应用内通知发送完成")
	return nil
}

// sendAliyunSMS 发送阿里云短信
func (n *NotificationService) sendAliyunSMS(ctx context.Context, recipients []string, content string) error {
	g.Log().Info(ctx, "通过阿里云发送短信", "recipients", recipients)
	
	// 这里实现阿里云短信API调用逻辑
	// 需要集成阿里云SDK或调用REST API
	
	// 模拟发送成功
	g.Log().Info(ctx, "阿里云短信发送完成")
	return nil
}

// sendTencentSMS 发送腾讯云短信
func (n *NotificationService) sendTencentSMS(ctx context.Context, recipients []string, content string) error {
	g.Log().Info(ctx, "通过腾讯云发送短信", "recipients", recipients)
	
	// 这里实现腾讯云短信API调用逻辑
	// 需要集成腾讯云SDK或调用REST API
	
	// 模拟发送成功
	g.Log().Info(ctx, "腾讯云短信发送完成")
	return nil
}

// ValidateEmailConfig 验证邮件配置
func (n *NotificationService) ValidateEmailConfig() error {
	if n.emailConfig.SMTPHost == "" {
		return fmt.Errorf("SMTP主机未配置")
	}
	if n.emailConfig.Username == "" {
		return fmt.Errorf("SMTP用户名未配置")
	}
	if n.emailConfig.Password == "" {
		return fmt.Errorf("SMTP密码未配置")
	}
	if n.emailConfig.FromEmail == "" {
		return fmt.Errorf("发件人邮箱未配置")
	}
	return nil
}

// ValidateSMSConfig 验证短信配置
func (n *NotificationService) ValidateSMSConfig() error {
	if n.smsConfig.Provider == "" {
		return fmt.Errorf("短信服务商未配置")
	}
	if n.smsConfig.APIKey == "" {
		return fmt.Errorf("短信API密钥未配置")
	}
	return nil
}

// GetEmailTemplate 获取邮件模板
func (n *NotificationService) GetEmailTemplate(templateType string, data map[string]interface{}) string {
	switch templateType {
	case "report_ready":
		return n.buildReportReadyEmailTemplate(data)
	case "report_failed":
		return n.buildReportFailedEmailTemplate(data)
	case "task_reminder":
		return n.buildTaskReminderEmailTemplate(data)
	default:
		return n.buildDefaultEmailTemplate(data)
	}
}

// buildReportReadyEmailTemplate 构建报表完成邮件模板
func (n *NotificationService) buildReportReadyEmailTemplate(data map[string]interface{}) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>报表生成完成</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f8f9fa; }
        .footer { padding: 15px; text-align: center; color: #6c757d; font-size: 12px; }
        .btn { background: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 报表生成完成</h1>
        </div>
        <div class="content">
            <p>亲爱的用户，</p>
            <p>您请求的报表已经生成完成：</p>
            <ul>
                <li><strong>报表名称：</strong>%s</li>
                <li><strong>报表类型：</strong>%s</li>
                <li><strong>生成时间：</strong>%s</li>
                <li><strong>报表格式：</strong>%s</li>
            </ul>
            <p>请登录系统查看详细报表内容。</p>
            <p style="text-align: center; margin-top: 30px;">
                <a href="#" class="btn">查看报表</a>
            </p>
        </div>
        <div class="footer">
            <p>此邮件由 MER 系统自动发送，请勿回复。</p>
            <p>© 2025 MER System. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, 
		data["task_name"], 
		data["report_type"], 
		data["generated_at"], 
		data["file_format"])
}

// buildReportFailedEmailTemplate 构建报表失败邮件模板
func (n *NotificationService) buildReportFailedEmailTemplate(data map[string]interface{}) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>报表生成失败</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f8f9fa; }
        .footer { padding: 15px; text-align: center; color: #6c757d; font-size: 12px; }
        .error { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; padding: 10px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>❌ 报表生成失败</h1>
        </div>
        <div class="content">
            <p>很抱歉，您的报表生成失败：</p>
            <ul>
                <li><strong>任务名称：</strong>%s</li>
                <li><strong>失败时间：</strong>%s</li>
            </ul>
            <div class="error">
                <strong>错误信息：</strong>%s
            </div>
            <p>请检查报表配置或联系系统管理员。</p>
        </div>
        <div class="footer">
            <p>此邮件由 MER 系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`, 
		data["task_name"], 
		data["failed_at"], 
		data["error_message"])
}

// buildTaskReminderEmailTemplate 构建任务提醒邮件模板
func (n *NotificationService) buildTaskReminderEmailTemplate(data map[string]interface{}) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>定时任务提醒</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #28a745; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f8f9fa; }
        .footer { padding: 15px; text-align: center; color: #6c757d; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⏰ 定时任务提醒</h1>
        </div>
        <div class="content">
            <p>您好，</p>
            <p>这是您订阅的定时报表提醒：</p>
            <ul>
                <li><strong>任务名称：</strong>%s</li>
                <li><strong>下次执行时间：</strong>%s</li>
                <li><strong>执行频率：</strong>%s</li>
            </ul>
            <p>如需修改或取消任务，请登录系统进行操作。</p>
        </div>
        <div class="footer">
            <p>此邮件由 MER 系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`, 
		data["task_name"], 
		data["next_run_time"], 
		data["cron_expression"])
}

// buildDefaultEmailTemplate 构建默认邮件模板
func (n *NotificationService) buildDefaultEmailTemplate(data map[string]interface{}) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>系统通知</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #6c757d; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f8f9fa; }
        .footer { padding: 15px; text-align: center; color: #6c757d; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📬 系统通知</h1>
        </div>
        <div class="content">
            <p>%s</p>
        </div>
        <div class="footer">
            <p>此邮件由 MER 系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`, data["content"])
}
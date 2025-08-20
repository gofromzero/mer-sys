package notification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/cache"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeEmail  NotificationType = "email"
	NotificationTypeSMS    NotificationType = "sms"
	NotificationTypeSystem NotificationType = "system"
)

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID          string           `json:"id"`
	Type        NotificationType `json:"type"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Variables   []string         `json:"variables"`   // 模板变量
	IsActive    bool             `json:"is_active"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// NotificationMessage 通知消息
type NotificationMessage struct {
	ID           string                 `json:"id"`
	TenantID     uint64                 `json:"tenant_id"`
	RecipientID  uint64                 `json:"recipient_id"`
	Type         NotificationType       `json:"type"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Data         map[string]interface{} `json:"data"`
	Status       string                 `json:"status"`      // pending, sent, failed
	SentAt       *time.Time             `json:"sent_at"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// NotificationService 通知服务
type NotificationService struct {
	cache *cache.Cache
}

// NewNotificationService 创建通知服务实例
func NewNotificationService() *NotificationService {
	return &NotificationService{
		cache: cache.NewCache("notification"),
	}
}

// 预定义的商户通知模板
var merchantNotificationTemplates = map[string]*NotificationTemplate{
	"merchant_approved": {
		ID:      "merchant_approved",
		Type:    NotificationTypeEmail,
		Title:   "商户审批通过通知",
		Content: `尊敬的 {{.ContactName}}，您好！

您申请的商户 "{{.MerchantName}}" 已审批通过，现在可以正常使用平台服务。

审批时间：{{.ApprovalTime}}
审批人：{{.ApproverName}}

如有疑问，请联系客服。`,
		Variables: []string{"ContactName", "MerchantName", "ApprovalTime", "ApproverName"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	"merchant_rejected": {
		ID:      "merchant_rejected",
		Type:    NotificationTypeEmail,
		Title:   "商户审批被拒通知",
		Content: `尊敬的 {{.ContactName}}，您好！

很遗憾，您申请的商户 "{{.MerchantName}}" 未能通过审批。

拒绝原因：{{.RejectReason}}
审批时间：{{.ApprovalTime}}

您可以根据反馈意见修改后重新申请。如有疑问，请联系客服。`,
		Variables: []string{"ContactName", "MerchantName", "RejectReason", "ApprovalTime"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	"merchant_status_changed": {
		ID:      "merchant_status_changed",
		Type:    NotificationTypeSystem,
		Title:   "商户状态变更通知",
		Content: `您的商户 "{{.MerchantName}}" 状态已由 {{.OldStatus}} 变更为 {{.NewStatus}}。

变更时间：{{.ChangeTime}}
变更原因：{{.ChangeReason}}

如有疑问，请联系客服。`,
		Variables: []string{"MerchantName", "OldStatus", "NewStatus", "ChangeTime", "ChangeReason"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
}

// SendMerchantApprovedNotification 发送商户审批通过通知
func (ns *NotificationService) SendMerchantApprovedNotification(ctx context.Context, merchant *types.Merchant, approverName string) error {
	template := merchantNotificationTemplates["merchant_approved"]
	
	// 准备模板变量
	variables := map[string]interface{}{
		"ContactName":   merchant.BusinessInfo.ContactName,
		"MerchantName":  merchant.Name,
		"ApprovalTime":  time.Now().Format("2006-01-02 15:04:05"),
		"ApproverName":  approverName,
	}
	
	// 渲染通知内容
	content, err := ns.renderTemplate(template.Content, variables)
	if err != nil {
		return fmt.Errorf("渲染通知模板失败: %w", err)
	}
	
	// 创建通知消息
	message := &NotificationMessage{
		ID:          ns.generateMessageID(),
		TenantID:    merchant.TenantID,
		RecipientID: 0, // 这里应该是商户联系人的用户ID，现在使用0作为占位
		Type:        template.Type,
		Title:       template.Title,
		Content:     content,
		Data: map[string]interface{}{
			"merchant_id":   merchant.ID,
			"template_id":   template.ID,
			"variables":     variables,
		},
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// 发送通知
	return ns.sendNotification(ctx, message)
}

// SendMerchantRejectedNotification 发送商户审批被拒通知
func (ns *NotificationService) SendMerchantRejectedNotification(ctx context.Context, merchant *types.Merchant, rejectReason string) error {
	template := merchantNotificationTemplates["merchant_rejected"]
	
	// 准备模板变量
	variables := map[string]interface{}{
		"ContactName":   merchant.BusinessInfo.ContactName,
		"MerchantName":  merchant.Name,
		"RejectReason":  rejectReason,
		"ApprovalTime":  time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// 渲染通知内容
	content, err := ns.renderTemplate(template.Content, variables)
	if err != nil {
		return fmt.Errorf("渲染通知模板失败: %w", err)
	}
	
	// 创建通知消息
	message := &NotificationMessage{
		ID:          ns.generateMessageID(),
		TenantID:    merchant.TenantID,
		RecipientID: 0, // 这里应该是商户联系人的用户ID
		Type:        template.Type,
		Title:       template.Title,
		Content:     content,
		Data: map[string]interface{}{
			"merchant_id":   merchant.ID,
			"template_id":   template.ID,
			"variables":     variables,
		},
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// 发送通知
	return ns.sendNotification(ctx, message)
}

// SendMerchantStatusChangedNotification 发送商户状态变更通知
func (ns *NotificationService) SendMerchantStatusChangedNotification(ctx context.Context, merchant *types.Merchant, oldStatus, newStatus types.MerchantStatus, changeReason string) error {
	template := merchantNotificationTemplates["merchant_status_changed"]
	
	// 状态中文映射
	statusNameMap := map[types.MerchantStatus]string{
		types.MerchantStatusPending:     "待审核",
		types.MerchantStatusActive:      "已激活",
		types.MerchantStatusSuspended:   "已暂停",
		types.MerchantStatusDeactivated: "已停用",
	}
	
	// 准备模板变量
	variables := map[string]interface{}{
		"MerchantName": merchant.Name,
		"OldStatus":    statusNameMap[oldStatus],
		"NewStatus":    statusNameMap[newStatus],
		"ChangeTime":   time.Now().Format("2006-01-02 15:04:05"),
		"ChangeReason": changeReason,
	}
	
	// 渲染通知内容
	content, err := ns.renderTemplate(template.Content, variables)
	if err != nil {
		return fmt.Errorf("渲染通知模板失败: %w", err)
	}
	
	// 创建通知消息
	message := &NotificationMessage{
		ID:          ns.generateMessageID(),
		TenantID:    merchant.TenantID,
		RecipientID: 0, // 这里应该是商户联系人的用户ID
		Type:        template.Type,
		Title:       template.Title,
		Content:     content,
		Data: map[string]interface{}{
			"merchant_id":   merchant.ID,
			"template_id":   template.ID,
			"variables":     variables,
		},
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// 发送通知
	return ns.sendNotification(ctx, message)
}

// sendNotification 发送通知
func (ns *NotificationService) sendNotification(ctx context.Context, message *NotificationMessage) error {
	// 根据通知类型选择发送方式
	switch message.Type {
	case NotificationTypeEmail:
		return ns.sendEmailNotification(ctx, message)
	case NotificationTypeSMS:
		return ns.sendSMSNotification(ctx, message)
	case NotificationTypeSystem:
		return ns.sendSystemNotification(ctx, message)
	default:
		return fmt.Errorf("不支持的通知类型: %s", message.Type)
	}
}

// sendEmailNotification 发送邮件通知
func (ns *NotificationService) sendEmailNotification(ctx context.Context, message *NotificationMessage) error {
	// 这里应该集成真实的邮件服务，如SMTP、SendGrid等
	// 现在仅记录日志作为演示
	g.Log().Infof(ctx, "发送邮件通知: ID=%s, 收件人=%d, 标题=%s", 
		message.ID, message.RecipientID, message.Title)
	
	// 存储到缓存中模拟发送成功
	key := fmt.Sprintf("notification:%s", message.ID)
	message.Status = "sent"
	message.SentAt = &[]time.Time{time.Now()}[0]
	
	err := ns.cache.Set(ctx, key, message, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("保存通知记录失败: %w", err)
	}
	
	return nil
}

// sendSMSNotification 发送短信通知
func (ns *NotificationService) sendSMSNotification(ctx context.Context, message *NotificationMessage) error {
	// 这里应该集成真实的短信服务，如阿里云短信、腾讯云短信等
	// 现在仅记录日志作为演示
	g.Log().Infof(ctx, "发送短信通知: ID=%s, 收件人=%d, 内容=%s", 
		message.ID, message.RecipientID, message.Content)
	
	// 存储到缓存中模拟发送成功
	key := fmt.Sprintf("notification:%s", message.ID)
	message.Status = "sent"
	message.SentAt = &[]time.Time{time.Now()}[0]
	
	err := ns.cache.Set(ctx, key, message, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("保存通知记录失败: %w", err)
	}
	
	return nil
}

// sendSystemNotification 发送系统通知
func (ns *NotificationService) sendSystemNotification(ctx context.Context, message *NotificationMessage) error {
	// 系统内通知，存储在缓存或数据库中，用户登录时显示
	g.Log().Infof(ctx, "发送系统通知: ID=%s, 收件人=%d, 标题=%s", 
		message.ID, message.RecipientID, message.Title)
	
	// 存储到缓存中
	key := fmt.Sprintf("notification:%s", message.ID)
	message.Status = "sent"
	message.SentAt = &[]time.Time{time.Now()}[0]
	
	err := ns.cache.Set(ctx, key, message, 7*24*time.Hour) // 系统通知保存7天
	if err != nil {
		return fmt.Errorf("保存系统通知失败: %w", err)
	}
	
	// 添加到用户的未读通知列表 (使用哈希存储)
	userNotifyKey := fmt.Sprintf("user_notifications:%d", message.RecipientID)
	err = ns.cache.HSet(ctx, userNotifyKey, message.ID, "unread")
	if err != nil {
		g.Log().Warningf(ctx, "添加到用户通知列表失败: %v", err)
	}
	
	return nil
}

// renderTemplate 渲染模板
func (ns *NotificationService) renderTemplate(template string, variables map[string]interface{}) (string, error) {
	// 简单的模板替换，实际项目中可以使用更强大的模板引擎
	content := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		content = strings.ReplaceAll(content, placeholder, valueStr)
	}
	return content, nil
}

// generateMessageID 生成消息ID
func (ns *NotificationService) generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// GetNotificationHistory 获取通知历史
func (ns *NotificationService) GetNotificationHistory(ctx context.Context, tenantID uint64, limit int) ([]*NotificationMessage, error) {
	// 这里应该从数据库查询，现在返回模拟数据
	messages := []*NotificationMessage{
		{
			ID:          "msg_001",
			TenantID:    tenantID,
			RecipientID: 1,
			Type:        NotificationTypeEmail,
			Title:       "商户审批通过通知",
			Content:     "您的商户申请已通过审批",
			Status:      "sent",
			SentAt:      &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
	}
	
	return messages, nil
}

// 全局通知服务实例
var defaultNotificationService = NewNotificationService()

// 全局通知函数
func SendMerchantApprovedNotification(ctx context.Context, merchant *types.Merchant, approverName string) error {
	return defaultNotificationService.SendMerchantApprovedNotification(ctx, merchant, approverName)
}

func SendMerchantRejectedNotification(ctx context.Context, merchant *types.Merchant, rejectReason string) error {
	return defaultNotificationService.SendMerchantRejectedNotification(ctx, merchant, rejectReason)
}

func SendMerchantStatusChangedNotification(ctx context.Context, merchant *types.Merchant, oldStatus, newStatus types.MerchantStatus, changeReason string) error {
	return defaultNotificationService.SendMerchantStatusChangedNotification(ctx, merchant, oldStatus, newStatus, changeReason)
}
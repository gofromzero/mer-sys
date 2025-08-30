package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// NotificationTemplate 通知模板结构
type NotificationTemplate struct {
	ID          string                 `json:"id"`
	Type        NotificationMethodType `json:"type"`      // sms, email
	Category    NotificationCategory   `json:"category"`  // customer, merchant
	Event       NotificationEvent      `json:"event"`     // order_created, status_changed, etc.
	Language    string                 `json:"language"`  // zh-CN, en-US
	Subject     string                 `json:"subject"`   // 邮件主题模板
	Content     string                 `json:"content"`   // 内容模板
	Variables   []string               `json:"variables"` // 模板变量列表
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NotificationMethodType 通知方法类型
type NotificationMethodType string

const (
	NotificationMethodTypeSMS   NotificationMethodType = "sms"
	NotificationMethodTypeEmail NotificationMethodType = "email"
	NotificationMethodTypeWebSocket NotificationMethodType = "websocket"
	NotificationMethodTypeSystem NotificationMethodType = "system"
)

// NotificationCategory 通知分类
type NotificationCategory string

const (
	NotificationCategoryCustomer NotificationCategory = "customer"
	NotificationCategoryMerchant NotificationCategory = "merchant"
	NotificationCategoryAdmin    NotificationCategory = "admin"
)

// NotificationEvent 通知事件
type NotificationEvent string

const (
	NotificationEventOrderCreated         NotificationEvent = "order_created"
	NotificationEventPaymentSuccess       NotificationEvent = "payment_success"
	NotificationEventPaymentFailure       NotificationEvent = "payment_failure"
	NotificationEventOrderProcessing      NotificationEvent = "order_processing"
	NotificationEventOrderCompleted       NotificationEvent = "order_completed"
	NotificationEventOrderCancelled       NotificationEvent = "order_cancelled"
	NotificationEventOrderStatusChanged   NotificationEvent = "order_status_changed"
)

// NotificationTemplateManager 通知模板管理器
type NotificationTemplateManager struct {
	templates map[string]*NotificationTemplate
}

// NewNotificationTemplateManager 创建通知模板管理器
func NewNotificationTemplateManager() *NotificationTemplateManager {
	manager := &NotificationTemplateManager{
		templates: make(map[string]*NotificationTemplate),
	}
	
	// 初始化默认模板
	manager.initializeDefaultTemplates()
	
	return manager
}

// initializeDefaultTemplates 初始化默认模板
func (m *NotificationTemplateManager) initializeDefaultTemplates() {
	defaultTemplates := []*NotificationTemplate{
		// 客户端短信模板
		{
			ID:       "customer_sms_order_created_zh_cn",
			Type:     NotificationMethodTypeSMS,
			Category: NotificationCategoryCustomer,
			Event:    NotificationEventOrderCreated,
			Language: "zh-CN",
			Subject:  "",
			Content:  "您的订单 {{.OrderNumber}} 已创建成功，金额 ¥{{.TotalAmount}}，请及时支付。",
			Variables: []string{"OrderNumber", "TotalAmount"},
			Enabled:  true,
		},
		{
			ID:       "customer_sms_payment_success_zh_cn",
			Type:     NotificationMethodTypeSMS,
			Category: NotificationCategoryCustomer,
			Event:    NotificationEventPaymentSuccess,
			Language: "zh-CN",
			Subject:  "",
			Content:  "您的订单 {{.OrderNumber}} 支付成功，金额 ¥{{.TotalAmount}}，我们将尽快处理。",
			Variables: []string{"OrderNumber", "TotalAmount"},
			Enabled:  true,
		},
		{
			ID:       "customer_sms_order_processing_zh_cn",
			Type:     NotificationMethodTypeSMS,
			Category: NotificationCategoryCustomer,
			Event:    NotificationEventOrderProcessing,
			Language: "zh-CN",
			Subject:  "",
			Content:  "您的订单 {{.OrderNumber}} 已开始处理，我们将尽快为您完成。",
			Variables: []string{"OrderNumber"},
			Enabled:  true,
		},
		{
			ID:       "customer_sms_order_completed_zh_cn",
			Type:     NotificationMethodTypeSMS,
			Category: NotificationCategoryCustomer,
			Event:    NotificationEventOrderCompleted,
			Language: "zh-CN",
			Subject:  "",
			Content:  "您的订单 {{.OrderNumber}} 已完成，感谢使用！",
			Variables: []string{"OrderNumber"},
			Enabled:  true,
		},
		{
			ID:       "customer_sms_order_cancelled_zh_cn",
			Type:     NotificationMethodTypeSMS,
			Category: NotificationCategoryCustomer,
			Event:    NotificationEventOrderCancelled,
			Language: "zh-CN",
			Subject:  "",
			Content:  "您的订单 {{.OrderNumber}} 已取消，原因：{{.Reason}}。如有疑问请联系客服。",
			Variables: []string{"OrderNumber", "Reason"},
			Enabled:  true,
		},
		
		// 客户端邮件模板
		{
			ID:       "customer_email_order_created_zh_cn",
			Type:     NotificationMethodTypeEmail,
			Category: NotificationCategoryCustomer,
			Event:    NotificationEventOrderCreated,
			Language: "zh-CN",
			Subject:  "订单确认 - {{.OrderNumber}}",
			Content: `尊敬的客户，

您的订单已创建成功！

订单信息：
- 订单编号：{{.OrderNumber}}
- 订单金额：¥{{.TotalAmount}}
- 权益消耗：{{.TotalRightsCost}}
- 创建时间：{{.CreatedAt}}

请及时支付以完成订单。

此致
商户系统`,
			Variables: []string{"OrderNumber", "TotalAmount", "TotalRightsCost", "CreatedAt"},
			Enabled:  true,
		},
		{
			ID:       "customer_email_payment_success_zh_cn",
			Type:     NotificationMethodTypeEmail,
			Category: NotificationCategoryCustomer,
			Event:    NotificationEventPaymentSuccess,
			Language: "zh-CN",
			Subject:  "支付确认 - {{.OrderNumber}}",
			Content: `尊敬的客户，

您的订单支付成功！

订单信息：
- 订单编号：{{.OrderNumber}}
- 支付金额：¥{{.TotalAmount}}
- 支付时间：{{.PaymentTime}}

我们将尽快为您处理订单。

此致
商户系统`,
			Variables: []string{"OrderNumber", "TotalAmount", "PaymentTime"},
			Enabled:  true,
		},
		{
			ID:       "customer_email_order_processing_zh_cn",
			Type:     NotificationMethodTypeEmail,
			Category: NotificationCategoryCustomer,
			Event:    NotificationEventOrderProcessing,
			Language: "zh-CN",
			Subject:  "订单处理中 - {{.OrderNumber}}",
			Content: `尊敬的客户，

您的订单正在处理中！

订单信息：
- 订单编号：{{.OrderNumber}}
- 订单金额：¥{{.TotalAmount}}
- 处理时间：{{.ProcessingTime}}

我们将尽快为您完成订单，请耐心等待。

此致
商户系统`,
			Variables: []string{"OrderNumber", "TotalAmount", "ProcessingTime"},
			Enabled:  true,
		},
		
		// 商户端短信模板
		{
			ID:       "merchant_sms_order_status_changed_zh_cn",
			Type:     NotificationMethodTypeSMS,
			Category: NotificationCategoryMerchant,
			Event:    NotificationEventOrderStatusChanged,
			Language: "zh-CN",
			Subject:  "",
			Content:  "商户订单状态变更：订单 {{.OrderNumber}} 状态从 {{.FromStatus}} 变更为 {{.ToStatus}}。",
			Variables: []string{"OrderNumber", "FromStatus", "ToStatus"},
			Enabled:  true,
		},
		
		// 商户端邮件模板
		{
			ID:       "merchant_email_order_status_changed_zh_cn",
			Type:     NotificationMethodTypeEmail,
			Category: NotificationCategoryMerchant,
			Event:    NotificationEventOrderStatusChanged,
			Language: "zh-CN",
			Subject:  "商户订单状态变更 - {{.OrderNumber}}",
			Content: `尊敬的商户管理员，

您管理的订单状态已发生变更。

订单信息：
- 订单编号：{{.OrderNumber}}
- 客户ID：{{.CustomerID}}
- 订单金额：¥{{.TotalAmount}}
- 状态变更：{{.FromStatus}} → {{.ToStatus}}
- 变更时间：{{.UpdatedAt}}
- 变更原因：{{.Reason}}
- 操作类型：{{.OperatorType}}

请登录商户管理后台查看详细信息。

此致
商户管理系统`,
			Variables: []string{"OrderNumber", "CustomerID", "TotalAmount", "FromStatus", "ToStatus", "UpdatedAt", "Reason", "OperatorType"},
			Enabled:  true,
		},
	}
	
	// 注册默认模板
	for _, template := range defaultTemplates {
		m.templates[template.ID] = template
	}
	
	g.Log().Info(context.Background(), "通知模板管理器已初始化", "template_count", len(m.templates))
}

// GetTemplate 获取模板
func (m *NotificationTemplateManager) GetTemplate(methodType NotificationMethodType, category NotificationCategory, event NotificationEvent, language string) *NotificationTemplate {
	templateID := m.buildTemplateID(methodType, category, event, language)
	template, exists := m.templates[templateID]
	if !exists {
		g.Log().Warning(context.Background(), "通知模板不存在", "template_id", templateID)
		return nil
	}
	
	return template
}

// buildTemplateID 构建模板ID
func (m *NotificationTemplateManager) buildTemplateID(methodType NotificationMethodType, category NotificationCategory, event NotificationEvent, language string) string {
	return string(category) + "_" + string(methodType) + "_" + string(event) + "_" + strings.ToLower(strings.ReplaceAll(language, "-", "_"))
}

// RenderTemplate 渲染模板
func (m *NotificationTemplateManager) RenderTemplate(template *NotificationTemplate, data map[string]interface{}) (string, string, error) {
	if template == nil {
		return "", "", fmt.Errorf("模板为空")
	}
	
	// 简单的模板变量替换（生产环境建议使用 text/template）
	subject := template.Subject
	content := template.Content
	
	for key, value := range data {
		placeholder := "{{." + key + "}}"
		valueStr := fmt.Sprintf("%v", value)
		subject = strings.ReplaceAll(subject, placeholder, valueStr)
		content = strings.ReplaceAll(content, placeholder, valueStr)
	}
	
	return subject, content, nil
}

// BuildOrderDataMap 构建订单数据映射
func (m *NotificationTemplateManager) BuildOrderDataMap(order *types.Order, statusHistory *types.OrderStatusHistory) map[string]interface{} {
	data := map[string]interface{}{
		"OrderNumber":      order.OrderNumber,
		"TotalAmount":      order.TotalAmount,
		"TotalRightsCost":  order.TotalRightsCost,
		"CreatedAt":        order.CreatedAt.Format("2006-01-02 15:04:05"),
		"PaymentTime":      time.Now().Format("2006-01-02 15:04:05"),
		"ProcessingTime":   time.Now().Format("2006-01-02 15:04:05"),
		"CustomerID":       order.CustomerID,
		"MerchantID":       order.MerchantID,
	}
	
	if statusHistory != nil {
		data["FromStatus"] = m.getOrderStatusDisplayName(m.orderStatusIntToString(statusHistory.FromStatus))
		data["ToStatus"] = m.getOrderStatusDisplayName(m.orderStatusIntToString(statusHistory.ToStatus))
		data["Reason"] = statusHistory.Reason
		data["OperatorType"] = m.getOperatorTypeDisplayName(statusHistory.OperatorType)
		data["UpdatedAt"] = statusHistory.CreatedAt.Format("2006-01-02 15:04:05")
	}
	
	return data
}

// getOrderStatusDisplayName 获取订单状态显示名称
func (m *NotificationTemplateManager) getOrderStatusDisplayName(status types.OrderStatus) string {
	switch status {
	case types.OrderStatusPending:
		return "待支付"
	case types.OrderStatusPaid:
		return "已支付"
	case types.OrderStatusProcessing:
		return "处理中"
	case types.OrderStatusCompleted:
		return "已完成"
	case types.OrderStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

// orderStatusIntToString 将数字状态转换为字符串状态
func (m *NotificationTemplateManager) orderStatusIntToString(status types.OrderStatusInt) types.OrderStatus {
	switch status {
	case types.OrderStatusIntPending:
		return types.OrderStatusPending
	case types.OrderStatusIntPaid:
		return types.OrderStatusPaid
	case types.OrderStatusIntProcessing:
		return types.OrderStatusProcessing
	case types.OrderStatusIntCompleted:
		return types.OrderStatusCompleted
	case types.OrderStatusIntCancelled:
		return types.OrderStatusCancelled
	default:
		return types.OrderStatusPending
	}
}

// getOperatorTypeDisplayName 获取操作类型显示名称
func (m *NotificationTemplateManager) getOperatorTypeDisplayName(operatorType types.OrderStatusOperatorType) string {
	switch operatorType {
	case types.OrderStatusOperatorTypeCustomer:
		return "客户操作"
	case types.OrderStatusOperatorTypeMerchant:
		return "商户操作"
	case types.OrderStatusOperatorTypeSystem:
		return "系统自动"
	case types.OrderStatusOperatorTypeAdmin:
		return "管理员操作"
	default:
		return "未知操作"
	}
}

// ListTemplates 列出所有模板
func (m *NotificationTemplateManager) ListTemplates() []*NotificationTemplate {
	templates := make([]*NotificationTemplate, 0, len(m.templates))
	for _, template := range m.templates {
		templates = append(templates, template)
	}
	return templates
}

// EnableTemplate 启用模板
func (m *NotificationTemplateManager) EnableTemplate(templateID string) error {
	template, exists := m.templates[templateID]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateID)
	}
	
	template.Enabled = true
	template.UpdatedAt = time.Now()
	
	return nil
}

// DisableTemplate 禁用模板
func (m *NotificationTemplateManager) DisableTemplate(templateID string) error {
	template, exists := m.templates[templateID]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateID)
	}
	
	template.Enabled = false
	template.UpdatedAt = time.Now()
	
	return nil
}
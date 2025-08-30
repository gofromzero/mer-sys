package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// WebSocketNotifier WebSocket通知器接口
type WebSocketNotifier interface {
	BroadcastOrderStatusChange(ctx context.Context, order *types.Order, statusHistory *types.OrderStatusHistory)
	SendOrderStatusChangeToUser(ctx context.Context, userID, tenantID uint64, order *types.Order, statusHistory *types.OrderStatusHistory)
}

// NotificationService 通知服务接口
type NotificationService interface {
	SendOrderCreatedNotification(ctx context.Context, order *types.Order) error
	SendPaymentSuccessNotification(ctx context.Context, order *types.Order) error
	SendPaymentFailureNotification(ctx context.Context, order *types.Order) error
	SendOrderCompletedNotification(ctx context.Context, order *types.Order) error
	
	// 新增：订单状态变更通知
	SendOrderStatusChangedNotification(ctx context.Context, order *types.Order, statusHistory *types.OrderStatusHistory) error
	SendOrderProcessingNotification(ctx context.Context, order *types.Order) error
	SendOrderCancelledNotification(ctx context.Context, order *types.Order, reason string) error
	
	// 新增：商户端通知
	SendMerchantOrderNotification(ctx context.Context, order *types.Order, statusHistory *types.OrderStatusHistory) error
	
	// 设置WebSocket通知器
	SetWebSocketNotifier(notifier WebSocketNotifier)
	
	// 内部通用方法
	sendNotificationByTemplate(ctx context.Context, userID uint64, category NotificationCategory, event NotificationEvent, order *types.Order, statusHistory *types.OrderStatusHistory) error
}

// notificationService 通知服务实现
type notificationService struct {
	smsService       SMSService
	emailService     EmailService
	webSocketNotifier WebSocketNotifier
	templateManager  *NotificationTemplateManager
}

// NewNotificationService 创建通知服务实例
func NewNotificationService() NotificationService {
	return &notificationService{
		smsService:       NewSMSService(),
		emailService:     NewEmailService(),
		webSocketNotifier: nil, // 稍后通过SetWebSocketNotifier设置
		templateManager:  NewNotificationTemplateManager(),
	}
}

// SetWebSocketNotifier 设置WebSocket通知器
func (s *notificationService) SetWebSocketNotifier(notifier WebSocketNotifier) {
	s.webSocketNotifier = notifier
}

// SendOrderCreatedNotification 发送订单创建通知
func (s *notificationService) SendOrderCreatedNotification(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "发送订单创建通知", "order_id", order.ID, "order_number", order.OrderNumber)

	// 构建模板数据
	data := s.templateManager.BuildOrderDataMap(order, nil)

	// 发送短信通知
	if smsTemplate := s.templateManager.GetTemplate(NotificationMethodTypeSMS, NotificationCategoryCustomer, NotificationEventOrderCreated, "zh-CN"); smsTemplate != nil && smsTemplate.Enabled {
		_, smsContent, err := s.templateManager.RenderTemplate(smsTemplate, data)
		if err != nil {
			g.Log().Error(ctx, "渲染短信模板失败", "error", err)
		} else {
			if err := s.smsService.SendSMS(ctx, order.CustomerID, "ORDER_CREATED", smsContent); err != nil {
				g.Log().Error(ctx, "发送订单创建短信通知失败", "error", err)
			}
		}
	}

	// 发送邮件通知
	if emailTemplate := s.templateManager.GetTemplate(NotificationMethodTypeEmail, NotificationCategoryCustomer, NotificationEventOrderCreated, "zh-CN"); emailTemplate != nil && emailTemplate.Enabled {
		emailSubject, emailContent, err := s.templateManager.RenderTemplate(emailTemplate, data)
		if err != nil {
			g.Log().Error(ctx, "渲染邮件模板失败", "error", err)
		} else {
			if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
				g.Log().Error(ctx, "发送订单创建邮件通知失败", "error", err)
			}
		}
	}

	return nil
}

// SendPaymentSuccessNotification 发送支付成功通知
func (s *notificationService) SendPaymentSuccessNotification(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "发送支付成功通知", "order_id", order.ID, "order_number", order.OrderNumber)

	// 构建模板数据
	data := s.templateManager.BuildOrderDataMap(order, nil)

	// 发送短信通知
	if smsTemplate := s.templateManager.GetTemplate(NotificationMethodTypeSMS, NotificationCategoryCustomer, NotificationEventPaymentSuccess, "zh-CN"); smsTemplate != nil && smsTemplate.Enabled {
		_, smsContent, err := s.templateManager.RenderTemplate(smsTemplate, data)
		if err != nil {
			g.Log().Error(ctx, "渲染短信模板失败", "error", err)
		} else {
			if err := s.smsService.SendSMS(ctx, order.CustomerID, "PAYMENT_SUCCESS", smsContent); err != nil {
				g.Log().Error(ctx, "发送支付成功短信通知失败", "error", err)
			}
		}
	}

	// 发送邮件通知
	if emailTemplate := s.templateManager.GetTemplate(NotificationMethodTypeEmail, NotificationCategoryCustomer, NotificationEventPaymentSuccess, "zh-CN"); emailTemplate != nil && emailTemplate.Enabled {
		emailSubject, emailContent, err := s.templateManager.RenderTemplate(emailTemplate, data)
		if err != nil {
			g.Log().Error(ctx, "渲染邮件模板失败", "error", err)
		} else {
			if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
				g.Log().Error(ctx, "发送支付成功邮件通知失败", "error", err)
			}
		}
	}

	return nil
}

// SendPaymentFailureNotification 发送支付失败通知
func (s *notificationService) SendPaymentFailureNotification(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "发送支付失败通知", "order_id", order.ID, "order_number", order.OrderNumber)

	// 短信通知
	smsContent := fmt.Sprintf("您的订单 %s 支付失败，请重新支付或联系客服。", order.OrderNumber)

	if err := s.smsService.SendSMS(ctx, order.CustomerID, "PAYMENT_FAILURE", smsContent); err != nil {
		g.Log().Error(ctx, "发送支付失败短信通知失败", "error", err)
	}

	// 邮件通知
	emailSubject := fmt.Sprintf("支付失败 - %s", order.OrderNumber)
	emailContent := s.generatePaymentFailureEmailContent(order)

	if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
		g.Log().Error(ctx, "发送支付失败邮件通知失败", "error", err)
	}

	return nil
}

// SendOrderCompletedNotification 发送订单完成通知
func (s *notificationService) SendOrderCompletedNotification(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "发送订单完成通知", "order_id", order.ID, "order_number", order.OrderNumber)

	// 短信通知
	smsContent := fmt.Sprintf("您的订单 %s 已完成，感谢您的使用！如有问题请联系客服。", order.OrderNumber)

	if err := s.smsService.SendSMS(ctx, order.CustomerID, "ORDER_COMPLETED", smsContent); err != nil {
		g.Log().Error(ctx, "发送订单完成短信通知失败", "error", err)
	}

	// 邮件通知
	emailSubject := fmt.Sprintf("订单完成 - %s", order.OrderNumber)
	emailContent := s.generateOrderCompletedEmailContent(order)

	if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
		g.Log().Error(ctx, "发送订单完成邮件通知失败", "error", err)
	}

	return nil
}

// generateOrderCreatedEmailContent 生成订单创建邮件内容
func (s *notificationService) generateOrderCreatedEmailContent(order *types.Order) string {
	return fmt.Sprintf(`
尊敬的客户，

您的订单已创建成功！

订单信息：
- 订单编号：%s
- 订单金额：¥%.2f
- 权益消耗：%.2f
- 创建时间：%s

请及时支付以完成订单。

此致
商户系统
`,
		order.OrderNumber,
		order.TotalAmount,
		order.TotalRightsCost,
		order.CreatedAt.Format("2006-01-02 15:04:05"))
}

// generatePaymentSuccessEmailContent 生成支付成功邮件内容
func (s *notificationService) generatePaymentSuccessEmailContent(order *types.Order) string {
	return fmt.Sprintf(`
尊敬的客户，

您的订单支付成功！

订单信息：
- 订单编号：%s
- 支付金额：¥%.2f
- 支付时间：%s

我们将尽快为您处理订单。

此致
商户系统
`,
		order.OrderNumber,
		order.TotalAmount,
		time.Now().Format("2006-01-02 15:04:05"))
}

// generatePaymentFailureEmailContent 生成支付失败邮件内容
func (s *notificationService) generatePaymentFailureEmailContent(order *types.Order) string {
	return fmt.Sprintf(`
尊敬的客户，

您的订单支付失败！

订单信息：
- 订单编号：%s
- 订单金额：¥%.2f

请重新尝试支付或联系我们的客服。

此致
商户系统
`,
		order.OrderNumber,
		order.TotalAmount)
}

// generateOrderCompletedEmailContent 生成订单完成邮件内容
func (s *notificationService) generateOrderCompletedEmailContent(order *types.Order) string {
	return fmt.Sprintf(`
尊敬的客户，

您的订单已完成！

订单信息：
- 订单编号：%s
- 订单金额：¥%.2f
- 完成时间：%s

感谢您的使用！如有任何问题，请联系我们的客服。

此致
商户系统
`,
		order.OrderNumber,
		order.TotalAmount,
		time.Now().Format("2006-01-02 15:04:05"))
}

// SendOrderStatusChangedNotification 发送订单状态变更通知
func (s *notificationService) SendOrderStatusChangedNotification(ctx context.Context, order *types.Order, statusHistory *types.OrderStatusHistory) error {
	g.Log().Info(ctx, "发送订单状态变更通知", 
		"order_id", order.ID, 
		"order_number", order.OrderNumber,
		"from_status", statusHistory.FromStatus,
		"to_status", statusHistory.ToStatus,
		"operator_type", statusHistory.OperatorType)

	// 发送WebSocket实时通知
	if s.webSocketNotifier != nil {
		s.webSocketNotifier.BroadcastOrderStatusChange(ctx, order, statusHistory)
		
		// 如果是客户订单，给客户发送WebSocket通知
		if order.CustomerID > 0 {
			s.webSocketNotifier.SendOrderStatusChangeToUser(ctx, order.CustomerID, order.TenantID, order, statusHistory)
		}
	}
	
	// 发送商户端通知
	if err := s.SendMerchantOrderNotification(ctx, order, statusHistory); err != nil {
		g.Log().Error(ctx, "发送商户端订单通知失败", "error", err, "order_id", order.ID)
		// 不返回错误，因为这不应该阻塞主要的客户通知
	}

	// 根据状态变更类型调用具体的通知方法
	switch statusHistory.ToStatus {
	case types.OrderStatusIntProcessing:
		return s.SendOrderProcessingNotification(ctx, order)
	case types.OrderStatusIntCompleted:
		return s.SendOrderCompletedNotification(ctx, order)
	case types.OrderStatusIntCancelled:
		return s.SendOrderCancelledNotification(ctx, order, statusHistory.Reason)
	default:
		// 通用状态变更通知
		return s.sendGenericStatusChangeNotification(ctx, order, statusHistory)
	}
}

// SendOrderProcessingNotification 发送订单处理中通知
func (s *notificationService) SendOrderProcessingNotification(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "发送订单处理中通知", "order_id", order.ID, "order_number", order.OrderNumber)

	// 短信通知
	smsContent := fmt.Sprintf("您的订单 %s 已开始处理，我们将尽快为您完成订单。",
		order.OrderNumber)

	if err := s.smsService.SendSMS(ctx, order.CustomerID, "ORDER_PROCESSING", smsContent); err != nil {
		g.Log().Error(ctx, "发送订单处理中短信通知失败", "error", err)
	}

	// 邮件通知
	emailSubject := fmt.Sprintf("订单处理中 - %s", order.OrderNumber)
	emailContent := s.generateOrderProcessingEmailContent(order)

	if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
		g.Log().Error(ctx, "发送订单处理中邮件通知失败", "error", err)
	}

	return nil
}

// SendOrderCancelledNotification 发送订单取消通知
func (s *notificationService) SendOrderCancelledNotification(ctx context.Context, order *types.Order, reason string) error {
	g.Log().Info(ctx, "发送订单取消通知", 
		"order_id", order.ID, 
		"order_number", order.OrderNumber,
		"reason", reason)

	// 短信通知
	smsContent := fmt.Sprintf("您的订单 %s 已被取消，原因：%s。如有疑问请联系客服。",
		order.OrderNumber, reason)

	if err := s.smsService.SendSMS(ctx, order.CustomerID, "ORDER_CANCELLED", smsContent); err != nil {
		g.Log().Error(ctx, "发送订单取消短信通知失败", "error", err)
	}

	// 邮件通知
	emailSubject := fmt.Sprintf("订单取消 - %s", order.OrderNumber)
	emailContent := s.generateOrderCancelledEmailContent(order, reason)

	if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
		g.Log().Error(ctx, "发送订单取消邮件通知失败", "error", err)
	}

	return nil
}

// sendGenericStatusChangeNotification 发送通用状态变更通知
func (s *notificationService) sendGenericStatusChangeNotification(ctx context.Context, order *types.Order, statusHistory *types.OrderStatusHistory) error {
	// 获取状态中文名称
	fromStatusName := s.getOrderStatusDisplayName(s.orderStatusIntToString(statusHistory.FromStatus))
	toStatusName := s.getOrderStatusDisplayName(s.orderStatusIntToString(statusHistory.ToStatus))

	// 短信通知
	smsContent := fmt.Sprintf("您的订单 %s 状态已从 %s 变更为 %s。",
		order.OrderNumber, fromStatusName, toStatusName)

	if err := s.smsService.SendSMS(ctx, order.CustomerID, "ORDER_STATUS_CHANGED", smsContent); err != nil {
		g.Log().Error(ctx, "发送订单状态变更短信通知失败", "error", err)
	}

	// 邮件通知
	emailSubject := fmt.Sprintf("订单状态变更 - %s", order.OrderNumber)
	emailContent := s.generateOrderStatusChangeEmailContent(order, statusHistory, fromStatusName, toStatusName)

	if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
		g.Log().Error(ctx, "发送订单状态变更邮件通知失败", "error", err)
	}

	return nil
}

// generateOrderProcessingEmailContent 生成订单处理中邮件内容
func (s *notificationService) generateOrderProcessingEmailContent(order *types.Order) string {
	return fmt.Sprintf(`
尊敬的客户，

您的订单正在处理中！

订单信息：
- 订单编号：%s
- 订单金额：¥%.2f
- 处理时间：%s

我们将尽快为您完成订单，请耐心等待。

此致
商户系统
`,
		order.OrderNumber,
		order.TotalAmount,
		time.Now().Format("2006-01-02 15:04:05"))
}

// generateOrderCancelledEmailContent 生成订单取消邮件内容
func (s *notificationService) generateOrderCancelledEmailContent(order *types.Order, reason string) string {
	return fmt.Sprintf(`
尊敬的客户，

很抱歉，您的订单已被取消。

订单信息：
- 订单编号：%s
- 订单金额：¥%.2f
- 取消时间：%s
- 取消原因：%s

如有任何疑问，请联系我们的客服。

此致
商户系统
`,
		order.OrderNumber,
		order.TotalAmount,
		time.Now().Format("2006-01-02 15:04:05"),
		reason)
}

// generateOrderStatusChangeEmailContent 生成订单状态变更邮件内容
func (s *notificationService) generateOrderStatusChangeEmailContent(order *types.Order, statusHistory *types.OrderStatusHistory, fromStatusName, toStatusName string) string {
	return fmt.Sprintf(`
尊敬的客户，

您的订单状态已发生变更。

订单信息：
- 订单编号：%s
- 订单金额：¥%.2f
- 状态变更：%s → %s
- 变更时间：%s
- 变更原因：%s

如有任何疑问，请联系我们的客服。

此致
商户系统
`,
		order.OrderNumber,
		order.TotalAmount,
		fromStatusName,
		toStatusName,
		statusHistory.CreatedAt.Format("2006-01-02 15:04:05"),
		statusHistory.Reason)
}

// getOrderStatusDisplayName 获取订单状态显示名称
func (s *notificationService) getOrderStatusDisplayName(status types.OrderStatus) string {
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

// SendMerchantOrderNotification 发送商户端订单状态变更通知
func (s *notificationService) SendMerchantOrderNotification(ctx context.Context, order *types.Order, statusHistory *types.OrderStatusHistory) error {
	g.Log().Info(ctx, "发送商户端订单通知", 
		"order_id", order.ID, 
		"order_number", order.OrderNumber,
		"merchant_id", order.MerchantID,
		"from_status", statusHistory.FromStatus,
		"to_status", statusHistory.ToStatus)

	// 获取商户管理员用户ID列表
	merchantAdminIDs := s.getMerchantAdminUserIDs(ctx, order.MerchantID)
	if len(merchantAdminIDs) == 0 {
		g.Log().Warning(ctx, "未找到商户管理员，跳过商户端通知", 
			"merchant_id", order.MerchantID,
			"order_id", order.ID)
		return nil
	}

	// 获取状态中文名称
	fromStatusName := s.getOrderStatusDisplayName(s.orderStatusIntToString(statusHistory.FromStatus))
	toStatusName := s.getOrderStatusDisplayName(s.orderStatusIntToString(statusHistory.ToStatus))

	// 为每个商户管理员发送通知
	for _, adminID := range merchantAdminIDs {
		// 发送WebSocket实时通知
		if s.webSocketNotifier != nil {
			s.webSocketNotifier.SendOrderStatusChangeToUser(ctx, adminID, order.TenantID, order, statusHistory)
		}
		
		// 发送短信通知（如果配置了）
		go func(userID uint64) {
			smsContent := fmt.Sprintf("商户订单状态变更：订单 %s 状态从 %s 变更为 %s。",
				order.OrderNumber, fromStatusName, toStatusName)
			
			// 这里需要从用户服务获取手机号
			// phone := s.getUserPhone(ctx, userID)
			// if phone != "" {
			//     s.smsService.SendSMS(ctx, userID, "MERCHANT_ORDER_STATUS_CHANGED", smsContent)
			// }
			
			// 暂时记录日志
			g.Log().Info(ctx, "商户端短信通知（模拟）", 
				"user_id", userID, 
				"content", smsContent)
		}(adminID)
		
		// 发送邮件通知
		go func(userID uint64) {
			emailSubject := fmt.Sprintf("商户订单状态变更 - %s", order.OrderNumber)
			emailContent := s.generateMerchantOrderStatusChangeEmailContent(order, statusHistory, fromStatusName, toStatusName)
			
			if err := s.emailService.SendEmail(ctx, userID, emailSubject, emailContent); err != nil {
				g.Log().Error(ctx, "发送商户端邮件通知失败", "error", err, "user_id", userID)
			}
		}(adminID)
	}

	// 发送系统内通知
	for _, adminID := range merchantAdminIDs {
		title := fmt.Sprintf("订单状态变更 - %s", order.OrderNumber)
		message := fmt.Sprintf("订单 %s 状态从 %s 变更为 %s，请及时处理。原因：%s",
			order.OrderNumber, fromStatusName, toStatusName, statusHistory.Reason)
		
		go func(userID uint64) {
			// 这里应该调用系统通知服务
			g.Log().Info(ctx, "商户端系统通知（模拟）",
				"user_id", userID,
				"title", title,
				"message", message)
		}(adminID)
	}

	return nil
}

// generateMerchantOrderStatusChangeEmailContent 生成商户端订单状态变更邮件内容
func (s *notificationService) generateMerchantOrderStatusChangeEmailContent(order *types.Order, statusHistory *types.OrderStatusHistory, fromStatusName, toStatusName string) string {
	operatorTypeName := s.getOperatorTypeDisplayName(statusHistory.OperatorType)
	
	return fmt.Sprintf(`
尊敬的商户管理员，

您管理的订单状态已发生变更。

订单信息：
- 订单编号：%s
- 客户ID：%d
- 订单金额：¥%.2f
- 状态变更：%s → %s
- 变更时间：%s
- 变更原因：%s
- 操作类型：%s

请登录商户管理后台查看详细信息。

此致
商户管理系统
`,
		order.OrderNumber,
		order.CustomerID,
		order.TotalAmount,
		fromStatusName,
		toStatusName,
		statusHistory.CreatedAt.Format("2006-01-02 15:04:05"),
		statusHistory.Reason,
		operatorTypeName)
}

// orderStatusIntToString 将数字状态转换为字符串状态
func (s *notificationService) orderStatusIntToString(status types.OrderStatusInt) types.OrderStatus {
	switch status {
	case types.OrderStatusIntPending:
		return "pending"
	case types.OrderStatusIntPaid:
		return "paid"
	case types.OrderStatusIntProcessing:
		return "processing"
	case types.OrderStatusIntCompleted:
		return "completed"
	case types.OrderStatusIntCancelled:
		return "cancelled"
	default:
		return "pending"
	}
}

// getOperatorTypeDisplayName 获取操作类型显示名称
func (s *notificationService) getOperatorTypeDisplayName(operatorType types.OrderStatusOperatorType) string {
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

// getMerchantAdminUserIDs 获取商户管理员用户ID列表
// 这里是模拟实现，实际应该从用户服务查询
func (s *notificationService) getMerchantAdminUserIDs(ctx context.Context, merchantID uint64) []uint64 {
	// 模拟数据：假设每个商户有1-2个管理员
	// 实际实现应该调用用户服务API
	adminIDs := []uint64{
		merchantID * 1000,     // 主管理员
		merchantID*1000 + 1,   // 副管理员
	}
	
	g.Log().Info(ctx, "获取商户管理员ID列表（模拟）", 
		"merchant_id", merchantID, 
		"admin_ids", adminIDs)
	
	return adminIDs
}

// sendNotificationByTemplate 使用模板发送通知的通用方法
func (s *notificationService) sendNotificationByTemplate(ctx context.Context, userID uint64, category NotificationCategory, event NotificationEvent, order *types.Order, statusHistory *types.OrderStatusHistory) error {
	// 构建模板数据
	data := s.templateManager.BuildOrderDataMap(order, statusHistory)
	
	// 发送短信通知
	if smsTemplate := s.templateManager.GetTemplate(NotificationMethodTypeSMS, category, event, "zh-CN"); smsTemplate != nil && smsTemplate.Enabled {
		_, smsContent, err := s.templateManager.RenderTemplate(smsTemplate, data)
		if err != nil {
			g.Log().Error(ctx, "渲染短信模板失败", "error", err, "template_id", smsTemplate.ID)
		} else {
			// 构建短信模板代码
			smsTemplateCode := s.buildSMSTemplateCode(event)
			if err := s.smsService.SendSMS(ctx, userID, smsTemplateCode, smsContent); err != nil {
				g.Log().Error(ctx, "发送短信通知失败", "error", err, "user_id", userID, "event", event)
			} else {
				g.Log().Info(ctx, "短信通知发送成功", "user_id", userID, "event", event)
			}
		}
	}
	
	// 发送邮件通知
	if emailTemplate := s.templateManager.GetTemplate(NotificationMethodTypeEmail, category, event, "zh-CN"); emailTemplate != nil && emailTemplate.Enabled {
		emailSubject, emailContent, err := s.templateManager.RenderTemplate(emailTemplate, data)
		if err != nil {
			g.Log().Error(ctx, "渲染邮件模板失败", "error", err, "template_id", emailTemplate.ID)
		} else {
			if err := s.emailService.SendEmail(ctx, userID, emailSubject, emailContent); err != nil {
				g.Log().Error(ctx, "发送邮件通知失败", "error", err, "user_id", userID, "event", event)
			} else {
				g.Log().Info(ctx, "邮件通知发送成功", "user_id", userID, "event", event)
			}
		}
	}
	
	return nil
}

// buildSMSTemplateCode 根据事件构建短信模板代码
func (s *notificationService) buildSMSTemplateCode(event NotificationEvent) string {
	switch event {
	case NotificationEventOrderCreated:
		return "ORDER_CREATED"
	case NotificationEventPaymentSuccess:
		return "PAYMENT_SUCCESS"
	case NotificationEventPaymentFailure:
		return "PAYMENT_FAILURE"
	case NotificationEventOrderProcessing:
		return "ORDER_PROCESSING"
	case NotificationEventOrderCompleted:
		return "ORDER_COMPLETED"
	case NotificationEventOrderCancelled:
		return "ORDER_CANCELLED"
	case NotificationEventOrderStatusChanged:
		return "ORDER_STATUS_CHANGED"
	default:
		return "ORDER_STATUS_CHANGED"
	}
}

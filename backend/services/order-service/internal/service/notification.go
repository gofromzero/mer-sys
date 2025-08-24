package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// NotificationService 通知服务接口
type NotificationService interface {
	SendOrderCreatedNotification(ctx context.Context, order *types.Order) error
	SendPaymentSuccessNotification(ctx context.Context, order *types.Order) error
	SendPaymentFailureNotification(ctx context.Context, order *types.Order) error
	SendOrderCompletedNotification(ctx context.Context, order *types.Order) error
}

// notificationService 通知服务实现
type notificationService struct {
	smsService   SMSService
	emailService EmailService
}

// NewNotificationService 创建通知服务实例
func NewNotificationService() NotificationService {
	return &notificationService{
		smsService:   NewSMSService(),
		emailService: NewEmailService(),
	}
}

// SendOrderCreatedNotification 发送订单创建通知
func (s *notificationService) SendOrderCreatedNotification(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "发送订单创建通知", "order_id", order.ID, "order_number", order.OrderNumber)

	// 短信通知
	smsContent := fmt.Sprintf("您的订单 %s 已创建成功，订单金额 ¥%.2f，请及时支付。",
		order.OrderNumber, order.TotalAmount)

	if err := s.smsService.SendSMS(ctx, order.CustomerID, "ORDER_CREATED", smsContent); err != nil {
		g.Log().Error(ctx, "发送订单创建短信通知失败", "error", err)
	}

	// 邮件通知
	emailSubject := fmt.Sprintf("订单确认 - %s", order.OrderNumber)
	emailContent := s.generateOrderCreatedEmailContent(order)

	if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
		g.Log().Error(ctx, "发送订单创建邮件通知失败", "error", err)
	}

	return nil
}

// SendPaymentSuccessNotification 发送支付成功通知
func (s *notificationService) SendPaymentSuccessNotification(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "发送支付成功通知", "order_id", order.ID, "order_number", order.OrderNumber)

	// 短信通知
	smsContent := fmt.Sprintf("您的订单 %s 支付成功，金额 ¥%.2f，我们将尽快为您处理。",
		order.OrderNumber, order.TotalAmount)

	if err := s.smsService.SendSMS(ctx, order.CustomerID, "PAYMENT_SUCCESS", smsContent); err != nil {
		g.Log().Error(ctx, "发送支付成功短信通知失败", "error", err)
	}

	// 邮件通知
	emailSubject := fmt.Sprintf("支付确认 - %s", order.OrderNumber)
	emailContent := s.generatePaymentSuccessEmailContent(order)

	if err := s.emailService.SendEmail(ctx, order.CustomerID, emailSubject, emailContent); err != nil {
		g.Log().Error(ctx, "发送支付成功邮件通知失败", "error", err)
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

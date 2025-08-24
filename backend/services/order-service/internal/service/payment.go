package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IPaymentService 支付服务接口
type IPaymentService interface {
	InitiatePayment(ctx context.Context, orderID uint64, paymentMethod types.PaymentMethod, returnURL string) (*types.PaymentInfo, error)
	GetPaymentStatus(ctx context.Context, orderID uint64) (string, error)
	RetryPayment(ctx context.Context, orderID uint64, paymentMethod types.PaymentMethod, returnURL string) (*types.PaymentInfo, error)
	HandleAlipayCallback(ctx context.Context, callbackData map[string]interface{}) error
}

// PaymentService 支付服务实现
type PaymentService struct {
	orderRepo           repository.IOrderRepository
	notificationService NotificationService
}

// NewPaymentService 创建支付服务实例
func NewPaymentService() IPaymentService {
	return &PaymentService{
		orderRepo:           repository.NewOrderRepository(),
		notificationService: NewNotificationService(),
	}
}

// InitiatePayment 发起支付
func (s *PaymentService) InitiatePayment(ctx context.Context, orderID uint64, paymentMethod types.PaymentMethod, returnURL string) (*types.PaymentInfo, error) {
	// 获取订单信息
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("订单不存在: %v", err)
	}

	// 检查订单状态
	if order.Status != types.OrderStatusPending {
		return nil, fmt.Errorf("订单状态不正确，当前状态: %s", order.Status)
	}

	// 根据支付方式创建支付
	var paymentInfo *types.PaymentInfo
	switch paymentMethod {
	case types.PaymentMethodAlipay:
		paymentInfo, err = s.createAlipayPayment(ctx, order, returnURL)
	case types.PaymentMethodWechat:
		return nil, fmt.Errorf("暂不支持微信支付")
	case types.PaymentMethodBalance:
		return nil, fmt.Errorf("暂不支持余额支付")
	default:
		return nil, fmt.Errorf("不支持的支付方式: %s", paymentMethod)
	}

	if err != nil {
		return nil, fmt.Errorf("创建支付失败: %v", err)
	}

	// 更新订单支付信息
	order.PaymentInfo = paymentInfo
	err = s.orderRepo.Update(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("更新订单支付信息失败: %v", err)
	}

	return paymentInfo, nil
}

// sendPaymentNotification 发送支付状态变更通知
func (s *PaymentService) sendPaymentNotification(ctx context.Context, order *types.Order, originalStatus types.OrderStatus) {
	// 只有当状态发生变更时才发送通知
	if order.Status == originalStatus {
		return
	}

	switch order.Status {
	case types.OrderStatusPaid:
		// 支付成功通知
		if err := s.notificationService.SendPaymentSuccessNotification(ctx, order); err != nil {
			fmt.Printf("发送支付成功通知失败: %v\n", err)
		}
	case types.OrderStatusCompleted:
		// 订单完成通知
		if err := s.notificationService.SendOrderCompletedNotification(ctx, order); err != nil {
			fmt.Printf("发送订单完成通知失败: %v\n", err)
		}
	}

	// 如果是支付失败，可以在这里处理
	// 注意：这里没有直接的支付失败状态，可能需要根据具体业务逻辑判断
}

// GetPaymentStatus 查询支付状态
func (s *PaymentService) GetPaymentStatus(ctx context.Context, orderID uint64) (string, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return "", fmt.Errorf("订单不存在: %v", err)
	}

	if order.PaymentInfo == nil {
		return "unpaid", nil
	}

	// TODO: 这里应该调用第三方支付接口查询最新状态
	// 暂时返回基于订单状态的支付状态
	switch order.Status {
	case types.OrderStatusPaid:
		return "paid", nil
	case types.OrderStatusPending:
		return "unpaid", nil
	default:
		return "unknown", nil
	}
}

// RetryPayment 重新支付
func (s *PaymentService) RetryPayment(ctx context.Context, orderID uint64, paymentMethod types.PaymentMethod, returnURL string) (*types.PaymentInfo, error) {
	// 获取订单信息
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("订单不存在: %v", err)
	}

	// 检查订单状态，只有待支付的订单可以重新支付
	if order.Status != types.OrderStatusPending {
		return nil, fmt.Errorf("订单状态不允许重新支付，当前状态: %s", order.Status)
	}

	// 创建新的支付
	return s.InitiatePayment(ctx, orderID, paymentMethod, returnURL)
}

// HandleAlipayCallback 处理支付宝支付回调
func (s *PaymentService) HandleAlipayCallback(ctx context.Context, callbackData map[string]interface{}) error {
	// TODO: 验证支付宝回调签名

	// 获取回调中的订单号
	outTradeNo, ok := callbackData["out_trade_no"].(string)
	if !ok {
		return fmt.Errorf("回调数据中缺少订单号")
	}

	// 获取交易状态
	tradeStatus, ok := callbackData["trade_status"].(string)
	if !ok {
		return fmt.Errorf("回调数据中缺少交易状态")
	}

	// 根据订单号查找订单
	order, err := s.orderRepo.GetByOrderNumber(ctx, outTradeNo)
	if err != nil {
		return fmt.Errorf("订单不存在: %v", err)
	}

	// 记录原始状态用于判断是否需要发送通知
	originalStatus := order.Status

	// 根据交易状态更新订单
	switch tradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		// 支付成功
		order.Status = types.OrderStatusPaid
		// 更新支付时间
		now := time.Now()
		if order.PaymentInfo != nil {
			order.PaymentInfo.PaidAt = &now
		}

	case "TRADE_CLOSED":
		// 交易关闭，保持待支付状态，等待重新支付
		// 不修改订单状态
	}

	// 更新订单
	err = s.orderRepo.Update(ctx, order)
	if err != nil {
		return fmt.Errorf("更新订单失败: %v", err)
	}

	// 发送状态变更通知
	go s.sendPaymentNotification(context.Background(), order, originalStatus)

	// TODO: 支付成功后，应该：
	// 1. 扣减库存
	// 2. 扣减权益余额

	return nil
}

// createAlipayPayment 创建支付宝支付
func (s *PaymentService) createAlipayPayment(ctx context.Context, order *types.Order, returnURL string) (*types.PaymentInfo, error) {
	// TODO: 集成支付宝SDK创建支付
	// 这里使用模拟数据

	paymentID := fmt.Sprintf("alipay_%d", order.ID)

	return &types.PaymentInfo{
		Method:        string(types.PaymentMethodAlipay),
		TransactionID: paymentID,
		Amount:        order.TotalAmount,
	}, nil
}

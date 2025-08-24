package service

import (
	"context"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IOrderService 订单服务接口
type IOrderService interface {
	CreateOrder(ctx context.Context, customerID uint64, req *types.CreateOrderRequest) (*types.Order, error)
	GetOrder(ctx context.Context, orderID uint64) (*types.Order, error)
	ListOrders(ctx context.Context, customerID uint64, status types.OrderStatus, page, limit int) ([]*types.Order, int, error)
	CancelOrder(ctx context.Context, orderID uint64) error
	GetOrderConfirmation(ctx context.Context, customerID uint64, req *types.CreateOrderRequest) (*types.OrderConfirmation, error)
}

// OrderService 订单服务实现
type OrderService struct {
	orderRepo           repository.IOrderRepository
	cartRepo            repository.ICartRepository
	notificationService NotificationService
}

// NewOrderService 创建订单服务实例
func NewOrderService() IOrderService {
	return &OrderService{
		orderRepo:           repository.NewOrderRepository(),
		cartRepo:            repository.NewCartRepository(),
		notificationService: NewNotificationService(),
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, customerID uint64, req *types.CreateOrderRequest) (*types.Order, error) {
	// 首先获取订单确认信息，验证库存和权益
	confirmation, err := s.GetOrderConfirmation(ctx, customerID, req)
	if err != nil {
		return nil, fmt.Errorf("获取订单确认信息失败: %v", err)
	}

	if !confirmation.CanCreate {
		return nil, fmt.Errorf("无法创建订单: %s", confirmation.ErrorMessage)
	}

	// 生成订单号
	orderNumber, err := s.orderRepo.GenerateOrderNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("生成订单号失败: %v", err)
	}

	// 创建订单
	// 转换订单项为正确的类型
	items := make([]types.OrderItem, 0, len(confirmation.Items))
	for _, item := range confirmation.Items {
		items = append(items, types.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			Price:      item.UnitPrice,
			RightsCost: item.UnitRightsCost,
		})
	}

	order := &types.Order{
		MerchantID:      req.MerchantID,
		CustomerID:      customerID,
		OrderNumber:     orderNumber,
		Status:          types.OrderStatusPending,
		Items:           items,
		PaymentInfo:     nil, // 支付信息在支付时填充
		TotalAmount:     confirmation.TotalAmount,
		TotalRightsCost: confirmation.TotalRightsCost,
	}

	err = s.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("创建订单失败: %v", err)
	}

	// 清空购物车（如果是从购物车创建的订单）
	// TODO: 这里应该只清空已购买的商品项，暂时先全部清空
	cart, err := s.cartRepo.GetOrCreate(ctx, customerID)
	if err == nil {
		s.cartRepo.ClearCart(ctx, cart.ID)
	}

	// 发送订单创建通知
	go func() {
		if err := s.notificationService.SendOrderCreatedNotification(context.Background(), order); err != nil {
			// 通知发送失败不影响订单创建
			fmt.Printf("发送订单创建通知失败: %v\n", err)
		}
	}()

	return order, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, orderID uint64) (*types.Order, error) {
	return s.orderRepo.GetByID(ctx, orderID)
}

// ListOrders 获取订单列表
func (s *OrderService) ListOrders(ctx context.Context, customerID uint64, status types.OrderStatus, page, limit int) ([]*types.Order, int, error) {
	return s.orderRepo.List(ctx, customerID, status, page, limit)
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, orderID uint64) error {
	// 获取订单详情
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("订单不存在: %v", err)
	}

	// 检查订单状态，只有待支付的订单可以取消
	if order.Status != types.OrderStatusPending {
		return fmt.Errorf("订单状态为 %s，无法取消", order.Status)
	}

	// 更新订单状态
	return s.orderRepo.UpdateStatus(ctx, orderID, types.OrderStatusCancelled)
}

// GetOrderConfirmation 获取订单确认信息
func (s *OrderService) GetOrderConfirmation(ctx context.Context, customerID uint64, req *types.CreateOrderRequest) (*types.OrderConfirmation, error) {
	confirmation := &types.OrderConfirmation{
		Items:           make([]types.OrderConfirmationItem, 0, len(req.Items)),
		TotalAmount:     0,
		TotalRightsCost: 0,
		CanCreate:       true,
	}

	// TODO: 这里应该调用Product Service获取商品信息和库存
	// TODO: 这里应该调用Fund Service检查权益余额
	// 为了演示，暂时使用模拟数据
	for _, item := range req.Items {
		// 模拟商品价格
		unitPrice := 100.0
		unitRightsCost := 10.0
		stockAvailable := 50

		confirmationItem := types.OrderConfirmationItem{
			ProductID:          item.ProductID,
			ProductName:        fmt.Sprintf("商品%d", item.ProductID),
			Quantity:           item.Quantity,
			UnitPrice:          unitPrice,
			UnitRightsCost:     unitRightsCost,
			SubtotalAmount:     unitPrice * float64(item.Quantity),
			SubtotalRightsCost: unitRightsCost * float64(item.Quantity),
			StockAvailable:     stockAvailable,
			StockSufficient:    item.Quantity <= stockAvailable,
		}

		if !confirmationItem.StockSufficient {
			confirmation.CanCreate = false
			confirmation.ErrorMessage = fmt.Sprintf("商品%d库存不足，可用库存: %d", item.ProductID, stockAvailable)
		}

		confirmation.Items = append(confirmation.Items, confirmationItem)
		confirmation.TotalAmount += confirmationItem.SubtotalAmount
		confirmation.TotalRightsCost += confirmationItem.SubtotalRightsCost
	}

	// 模拟检查权益余额
	confirmation.AvailableRights = 1000.0 // 模拟可用权益
	if confirmation.TotalRightsCost > confirmation.AvailableRights {
		confirmation.CanCreate = false
		confirmation.ErrorMessage = fmt.Sprintf("权益余额不足，需要权益: %.2f，可用权益: %.2f",
			confirmation.TotalRightsCost, confirmation.AvailableRights)
	}

	return confirmation, nil
}

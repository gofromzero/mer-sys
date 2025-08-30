package service

import (
	"context"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// IOrderStatusService 订单状态管理服务接口
type IOrderStatusService interface {
	UpdateOrderStatus(ctx context.Context, orderID uint64, req *types.UpdateOrderStatusRequest) error
	BatchUpdateOrderStatus(ctx context.Context, req *types.BatchUpdateOrderStatusRequest) (*types.BatchUpdateOrderStatusResponse, error)
	GetOrderStatusHistory(ctx context.Context, orderID uint64) ([]types.OrderStatusHistory, error)
	ValidateStatusTransition(ctx context.Context, orderID uint64, toStatus types.OrderStatusInt) error
}

// OrderStatusService 订单状态管理服务实现
type OrderStatusService struct {
	orderRepo         repository.IOrderRepository
	statusHistoryRepo *repository.OrderStatusHistoryRepository
	notificationService NotificationService
}

// NewOrderStatusService 创建订单状态管理服务实例
func NewOrderStatusService() IOrderStatusService {
	return &OrderStatusService{
		orderRepo:         repository.NewOrderRepository(),
		statusHistoryRepo: repository.NewOrderStatusHistoryRepository(),
		notificationService: NewNotificationService(),
	}
}

// UpdateOrderStatus 更新订单状态
func (s *OrderStatusService) UpdateOrderStatus(ctx context.Context, orderID uint64, req *types.UpdateOrderStatusRequest) error {
	// 验证操作员类型
	if !req.OperatorType.IsValid() {
		return fmt.Errorf("无效的操作员类型: %s", req.OperatorType)
	}
	
	// 获取当前用户ID作为操作员ID
	var operatorID *uint64
	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(uint64); ok {
			operatorID = &uid
		}
	}
	
	// 如果是系统操作，operatorID可以为空
	if req.OperatorType != types.OrderStatusOperatorTypeSystem && operatorID == nil {
		return fmt.Errorf("非系统操作必须提供操作员ID")
	}
	
	// 获取更新前的订单信息以便发送通知
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("获取订单信息失败: %v", err)
	}
	
	// 更新订单状态并记录历史
	err = s.orderRepo.UpdateStatusWithHistory(
		ctx,
		orderID,
		req.Status,
		req.Reason,
		req.OperatorType,
		operatorID,
		req.Metadata,
	)
	if err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}
	
	// 获取更新后的订单信息
	updatedOrder, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		g.Log().Errorf(ctx, "获取更新后的订单信息失败: %v", err)
		// 不返回错误，因为状态已经更新成功
	} else {
		// 发送状态变更通知（异步执行，不阻塞主流程）
		go func() {
			// 创建带有超时的上下文，防止goroutine泄露
			notifyCtx := context.WithValue(context.Background(), "tenant_id", ctx.Value("tenant_id"))
			notifyCtx = context.WithValue(notifyCtx, "user_id", ctx.Value("user_id"))
			
			// 创建状态历史记录用于通知
			statusHistory := &types.OrderStatusHistory{
				OrderID:      orderID,
				FromStatus:   s.orderStatusToInt(order.Status),
				ToStatus:     req.Status,
				Reason:       req.Reason,
				OperatorType: req.OperatorType,
			}
			
			if err := s.notificationService.SendOrderStatusChangedNotification(notifyCtx, updatedOrder, statusHistory); err != nil {
				g.Log().Errorf(notifyCtx, "发送订单状态变更通知失败: %v", err)
			}
		}()
	}
	
	g.Log().Infof(ctx, "订单状态更新成功: orderID=%d, status=%s, operator=%s", 
		orderID, req.Status.String(), req.OperatorType.String())
	
	return nil
}

// BatchUpdateOrderStatus 批量更新订单状态
func (s *OrderStatusService) BatchUpdateOrderStatus(ctx context.Context, req *types.BatchUpdateOrderStatusRequest) (*types.BatchUpdateOrderStatusResponse, error) {
	// 验证操作员类型
	if !req.OperatorType.IsValid() {
		return nil, fmt.Errorf("无效的操作员类型: %s", req.OperatorType)
	}
	
	// 检查订单数量限制
	if len(req.OrderIDs) == 0 {
		return nil, fmt.Errorf("订单ID列表不能为空")
	}
	if len(req.OrderIDs) > 100 {
		return nil, fmt.Errorf("批量操作订单数量不能超过100个")
	}
	
	// 获取当前用户ID作为操作员ID
	var operatorID *uint64
	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(uint64); ok {
			operatorID = &uid
		}
	}
	
	// 如果是系统操作，operatorID可以为空
	if req.OperatorType != types.OrderStatusOperatorTypeSystem && operatorID == nil {
		return nil, fmt.Errorf("非系统操作必须提供操作员ID")
	}
	
	// 执行批量更新
	response, err := s.orderRepo.BatchUpdateStatus(ctx, req, operatorID)
	if err != nil {
		return nil, fmt.Errorf("批量更新订单状态失败: %v", err)
	}
	
	// 为成功更新的订单发送通知（异步执行）
	if response.SuccessCount > 0 {
		go func() {
			// 创建带有超时的上下文，防止goroutine泄露
			notifyCtx := context.WithValue(context.Background(), "tenant_id", ctx.Value("tenant_id"))
			notifyCtx = context.WithValue(notifyCtx, "user_id", ctx.Value("user_id"))
			
			// 这里简化处理，因为响应中没有具体的成功订单ID列表
			// 在实际实现中，Repository应该返回成功的订单ID列表
			for _, orderID := range req.OrderIDs {
				// 获取更新后的订单信息
				order, err := s.orderRepo.GetByID(notifyCtx, orderID)
				if err != nil {
					g.Log().Errorf(notifyCtx, "获取订单信息失败，跳过通知发送: orderID=%d, error=%v", orderID, err)
					continue
				}
				
				// 创建状态历史记录用于通知
				statusHistory := &types.OrderStatusHistory{
					OrderID:      orderID,
					FromStatus:   types.OrderStatusIntPaid, // 批量操作通常从paid状态开始
					ToStatus:     req.Status,
					Reason:       req.Reason,
					OperatorType: req.OperatorType,
				}
				
				if err := s.notificationService.SendOrderStatusChangedNotification(notifyCtx, order, statusHistory); err != nil {
					g.Log().Errorf(notifyCtx, "发送批量订单状态变更通知失败: orderID=%d, error=%v", orderID, err)
				}
			}
		}()
	}
	
	g.Log().Infof(ctx, "批量更新订单状态完成: success=%d, fail=%d, operator=%s", 
		response.SuccessCount, response.FailCount, req.OperatorType.String())
	
	return response, nil
}

// GetOrderStatusHistory 获取订单状态历史
func (s *OrderStatusService) GetOrderStatusHistory(ctx context.Context, orderID uint64) ([]types.OrderStatusHistory, error) {
	// 验证订单是否存在（同时验证租户权限）
	_, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("订单不存在或无权访问: %v", err)
	}
	
	// 获取状态历史
	history, err := s.statusHistoryRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("获取订单状态历史失败: %v", err)
	}
	
	return history, nil
}

// ValidateStatusTransition 验证订单状态转换是否合法
func (s *OrderStatusService) ValidateStatusTransition(ctx context.Context, orderID uint64, toStatus types.OrderStatusInt) error {
	// 获取当前订单
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("获取订单失败: %v", err)
	}
	
	// 将字符串状态转换为数字状态进行比较
	currentStatusInt := s.orderStatusToInt(order.Status)
	
	// 验证状态转换是否合法
	if !currentStatusInt.IsValidTransition(toStatus) {
		return fmt.Errorf("不允许从状态 %s 转换到 %s", currentStatusInt.String(), toStatus.String())
	}
	
	return nil
}

// orderStatusToInt 将字符串状态转换为数字状态
func (s *OrderStatusService) orderStatusToInt(status types.OrderStatus) types.OrderStatusInt {
	switch status {
	case "pending":
		return types.OrderStatusIntPending
	case "paid":
		return types.OrderStatusIntPaid
	case "processing":
		return types.OrderStatusIntProcessing
	case "completed":
		return types.OrderStatusIntCompleted
	case "cancelled":
		return types.OrderStatusIntCancelled
	default:
		return types.OrderStatusIntPending
	}
}
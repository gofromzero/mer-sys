package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// OrderTimeoutService 订单超时处理服务
type OrderTimeoutService struct {
	orderRepo         repository.IOrderRepository
	timeoutConfigRepo *repository.OrderTimeoutConfigRepository
	orderStatusService IOrderStatusService
	notificationService NotificationService
	stopCh            chan struct{}
	isRunning         bool
}

// NewOrderTimeoutService 创建订单超时处理服务实例
func NewOrderTimeoutService(orderStatusService IOrderStatusService, notificationService NotificationService) *OrderTimeoutService {
	return &OrderTimeoutService{
		orderRepo:           repository.NewOrderRepository(),
		timeoutConfigRepo:   repository.NewOrderTimeoutConfigRepository(),
		orderStatusService:  orderStatusService,
		notificationService: notificationService,
		stopCh:             make(chan struct{}),
		isRunning:          false,
	}
}

// StartTimeoutMonitor 启动超时监控定时任务
func (s *OrderTimeoutService) StartTimeoutMonitor(ctx context.Context) {
	if s.isRunning {
		g.Log().Warning(ctx, "订单超时监控已在运行中")
		return
	}

	s.isRunning = true
	g.Log().Info(ctx, "启动订单超时监控服务")

	// 启动定时检查任务
	go s.runTimeoutCheck(ctx)
}

// StopTimeoutMonitor 停止超时监控
func (s *OrderTimeoutService) StopTimeoutMonitor(ctx context.Context) {
	if !s.isRunning {
		return
	}

	s.isRunning = false
	close(s.stopCh)
	g.Log().Info(ctx, "订单超时监控服务已停止")
}

// runTimeoutCheck 运行超时检查循环
func (s *OrderTimeoutService) runTimeoutCheck(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.processTimeoutOrders(ctx); err != nil {
				g.Log().Error(ctx, "处理超时订单失败", "error", err)
			}
		}
	}
}

// ProcessTimeoutOrders 处理超时订单（公开方法）
func (s *OrderTimeoutService) ProcessTimeoutOrders(ctx context.Context) error {
	return s.processTimeoutOrders(ctx)
}

// processTimeoutOrders 处理超时订单
func (s *OrderTimeoutService) processTimeoutOrders(ctx context.Context) error {
	// 获取需要检查超时的订单状态
	timeoutStatuses := []types.OrderStatusInt{
		types.OrderStatusIntPending,    // 待支付超时
		types.OrderStatusIntProcessing, // 处理中超时
	}

	for _, status := range timeoutStatuses {
		if err := s.processTimeoutOrdersByStatus(ctx, status); err != nil {
			g.Log().Error(ctx, "处理特定状态的超时订单失败", "status", status, "error", err)
		}
	}

	return nil
}

// processTimeoutOrdersByStatus 处理特定状态的超时订单
func (s *OrderTimeoutService) processTimeoutOrdersByStatus(ctx context.Context, status types.OrderStatusInt) error {
	// 获取该状态的超时订单
	orders, err := s.getTimeoutOrders(ctx, status)
	if err != nil {
		return fmt.Errorf("获取超时订单失败: %v", err)
	}

	if len(orders) == 0 {
		return nil
	}

	g.Log().Info(ctx, "发现超时订单", "status", status, "count", len(orders))

	// 批量处理超时订单
	for _, order := range orders {
		if err := s.processTimeoutOrder(ctx, order); err != nil {
			g.Log().Error(ctx, "处理单个超时订单失败", 
				"order_id", order.ID, 
				"order_number", order.OrderNumber,
				"error", err)
			continue
		}
	}

	return nil
}

// getTimeoutOrders 获取超时订单列表
func (s *OrderTimeoutService) getTimeoutOrders(ctx context.Context, status types.OrderStatusInt) ([]*types.Order, error) {
	// 根据状态构建不同的超时查询条件
	var timeCondition string
	var timeoutMinutes int

	switch status {
	case types.OrderStatusIntPending:
		// 待支付订单：检查创建时间 + 支付超时时间
		timeoutMinutes = 30 // 默认30分钟，后续从配置获取
		timeCondition = fmt.Sprintf("created_at < DATE_SUB(NOW(), INTERVAL %d MINUTE)", timeoutMinutes)
	case types.OrderStatusIntProcessing:
		// 处理中订单：检查最后状态更新时间 + 处理超时时间
		timeoutHours := 24 // 默认24小时，后续从配置获取
		timeCondition = fmt.Sprintf("status_updated_at < DATE_SUB(NOW(), INTERVAL %d HOUR)", timeoutHours)
	default:
		return nil, fmt.Errorf("不支持的订单状态超时检查: %s", status)
	}

	// 查询超时订单
	query := fmt.Sprintf(`
		SELECT o.* FROM orders o 
		WHERE o.status = ? 
		AND %s
		LIMIT 100
	`, timeCondition)

	var orders []*types.Order
	err := g.DB().Raw(query, s.orderStatusToString(status)).Scan(&orders)
	if err != nil {
		return nil, fmt.Errorf("查询超时订单失败: %v", err)
	}

	return orders, nil
}

// processTimeoutOrder 处理单个超时订单
func (s *OrderTimeoutService) processTimeoutOrder(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "开始处理超时订单", 
		"order_id", order.ID, 
		"order_number", order.OrderNumber, 
		"status", order.Status)

	currentStatus := s.stringToOrderStatusInt(order.Status)

	switch currentStatus {
	case types.OrderStatusIntPending:
		return s.processPendingTimeoutOrder(ctx, order)
	case types.OrderStatusIntProcessing:
		return s.processProcessingTimeoutOrder(ctx, order)
	default:
		return fmt.Errorf("不支持的超时订单状态处理: %s", order.Status)
	}
}

// processPendingTimeoutOrder 处理待支付超时订单
func (s *OrderTimeoutService) processPendingTimeoutOrder(ctx context.Context, order *types.Order) error {
	// 获取超时配置
	config, err := s.timeoutConfigRepo.GetEffectiveConfig(ctx, order.MerchantID)
	if err != nil {
		return fmt.Errorf("获取超时配置失败: %v", err)
	}

	// 检查是否真的超时了
	timeoutDuration := time.Duration(config.PaymentTimeoutMinutes) * time.Minute
	if time.Since(order.CreatedAt) < timeoutDuration {
		// 还未真正超时，可能是配置更新导致的
		return nil
	}

	// 自动取消订单
	updateReq := &types.UpdateOrderStatusRequest{
		Status:       types.OrderStatusIntCancelled,
		Reason:       "订单支付超时自动取消",
		OperatorType: types.OrderStatusOperatorTypeSystem,
		Metadata:     map[string]interface{}{
			"timeout_type": "payment_timeout",
			"timeout_config": map[string]interface{}{
				"payment_timeout_minutes": config.PaymentTimeoutMinutes,
			},
		},
	}

	err = s.orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
	if err != nil {
		return fmt.Errorf("自动取消超时订单失败: %v", err)
	}

	// 释放库存和权益
	if err := s.releaseOrderResources(ctx, order); err != nil {
		g.Log().Error(ctx, "释放订单资源失败", "order_id", order.ID, "error", err)
		// 不返回错误，因为订单状态已更新
	}

	g.Log().Info(ctx, "待支付超时订单已自动取消", 
		"order_id", order.ID, 
		"order_number", order.OrderNumber,
		"timeout_minutes", config.PaymentTimeoutMinutes)

	return nil
}

// processProcessingTimeoutOrder 处理处理中超时订单
func (s *OrderTimeoutService) processProcessingTimeoutOrder(ctx context.Context, order *types.Order) error {
	// 获取超时配置
	config, err := s.timeoutConfigRepo.GetEffectiveConfig(ctx, order.MerchantID)
	if err != nil {
		return fmt.Errorf("获取超时配置失败: %v", err)
	}

	// 检查是否真的超时了
	timeoutDuration := time.Duration(config.ProcessingTimeoutHours) * time.Hour
	if !order.StatusUpdatedAt.IsZero() && time.Since(order.StatusUpdatedAt) < timeoutDuration {
		// 还未真正超时
		return nil
	}

	// 根据配置决定处理方式
	if config.AutoCompleteEnabled {
		// 自动完成订单
		return s.autoCompleteOrder(ctx, order, config)
	} else {
		// 发送超时提醒通知，不自动处理
		return s.sendProcessingTimeoutNotification(ctx, order, config)
	}
}

// autoCompleteOrder 自动完成订单
func (s *OrderTimeoutService) autoCompleteOrder(ctx context.Context, order *types.Order, config *types.OrderTimeoutConfig) error {
	updateReq := &types.UpdateOrderStatusRequest{
		Status:       types.OrderStatusIntCompleted,
		Reason:       "订单处理超时自动完成",
		OperatorType: types.OrderStatusOperatorTypeSystem,
		Metadata:     map[string]interface{}{
			"timeout_type": "processing_timeout",
			"auto_complete": true,
			"timeout_config": map[string]interface{}{
				"processing_timeout_hours": config.ProcessingTimeoutHours,
			},
		},
	}

	err := s.orderStatusService.UpdateOrderStatus(ctx, order.ID, updateReq)
	if err != nil {
		return fmt.Errorf("自动完成超时订单失败: %v", err)
	}

	g.Log().Info(ctx, "处理中超时订单已自动完成", 
		"order_id", order.ID, 
		"order_number", order.OrderNumber,
		"timeout_hours", config.ProcessingTimeoutHours)

	return nil
}

// sendProcessingTimeoutNotification 发送处理超时通知
func (s *OrderTimeoutService) sendProcessingTimeoutNotification(ctx context.Context, order *types.Order, config *types.OrderTimeoutConfig) error {
	g.Log().Info(ctx, "发送处理超时提醒通知", 
		"order_id", order.ID, 
		"order_number", order.OrderNumber,
		"timeout_hours", config.ProcessingTimeoutHours)

	// 这里应该发送特殊的超时提醒通知
	// 暂时记录日志
	g.Log().Warning(ctx, "订单处理超时提醒", 
		"order_id", order.ID, 
		"order_number", order.OrderNumber,
		"merchant_id", order.MerchantID,
		"timeout_hours", config.ProcessingTimeoutHours)

	return nil
}

// releaseOrderResources 释放订单资源（库存和权益）
func (s *OrderTimeoutService) releaseOrderResources(ctx context.Context, order *types.Order) error {
	// 释放库存
	if err := s.releaseInventory(ctx, order); err != nil {
		return fmt.Errorf("释放库存失败: %v", err)
	}

	// 释放权益
	if err := s.releaseRights(ctx, order); err != nil {
		return fmt.Errorf("释放权益失败: %v", err)
	}

	// 处理支付退款
	if err := s.processRefund(ctx, order); err != nil {
		return fmt.Errorf("处理退款失败: %v", err)
	}

	return nil
}

// releaseInventory 释放库存
func (s *OrderTimeoutService) releaseInventory(ctx context.Context, order *types.Order) error {
	g.Log().Info(ctx, "释放订单库存", "order_id", order.ID)

	// 这里应该调用库存服务释放库存
	// 暂时模拟实现
	for _, item := range order.Items {
		g.Log().Info(ctx, "释放商品库存", 
			"order_id", order.ID,
			"product_id", item.ProductID,
			"quantity", item.Quantity)
	}

	return nil
}

// releaseRights 释放权益
func (s *OrderTimeoutService) releaseRights(ctx context.Context, order *types.Order) error {
	if order.TotalRightsCost <= 0 {
		return nil // 没有使用权益
	}

	g.Log().Info(ctx, "释放订单权益", 
		"order_id", order.ID,
		"rights_cost", order.TotalRightsCost)

	// 这里应该调用权益服务释放权益
	// 暂时模拟实现
	g.Log().Info(ctx, "权益释放完成", 
		"order_id", order.ID,
		"customer_id", order.CustomerID,
		"released_rights", order.TotalRightsCost)

	return nil
}

// processRefund 处理退款
func (s *OrderTimeoutService) processRefund(ctx context.Context, order *types.Order) error {
	// 只有已支付的订单才需要退款（检查订单状态和支付信息）
	if order.PaymentInfo == nil || order.PaymentInfo.PaidAt == nil {
		return nil // 未支付，无需退款
	}

	// 检查订单状态是否需要退款
	currentStatus := s.stringToOrderStatusInt(order.Status)
	if currentStatus == types.OrderStatusIntPending {
		return nil // 待支付状态不需要退款
	}

	g.Log().Info(ctx, "处理订单退款", 
		"order_id", order.ID,
		"refund_amount", order.TotalAmount)

	// 这里应该调用支付服务处理退款
	// 暂时模拟实现
	g.Log().Info(ctx, "退款处理完成", 
		"order_id", order.ID,
		"refund_amount", order.TotalAmount,
		"payment_method", order.PaymentInfo.Method)

	return nil
}

// GetTimeoutStatistics 获取超时统计信息
func (s *OrderTimeoutService) GetTimeoutStatistics(ctx context.Context, merchantID *uint64) (*types.OrderTimeoutStatistics, error) {
	// 构建查询条件
	whereConditions := []string{"tenant_id = ?"}
	args := []interface{}{ctx.Value("tenant_id")}

	if merchantID != nil {
		whereConditions = append(whereConditions, "merchant_id = ?")
		args = append(args, *merchantID)
	}

	// 统计各种超时情况
	stats := &types.OrderTimeoutStatistics{}

	// 统计待支付超时订单
	pendingTimeoutQuery := fmt.Sprintf(`
		SELECT COUNT(*) as count, COALESCE(SUM(total_amount), 0) as amount 
		FROM orders 
		WHERE %s AND status = 'pending' 
		AND created_at < DATE_SUB(NOW(), INTERVAL 30 MINUTE)
	`, "tenant_id = ?")

	var pendingTimeoutResult struct {
		Count  int     `db:"count"`
		Amount float64 `db:"amount"`
	}
	err := g.DB().Raw(pendingTimeoutQuery, args[0]).Scan(&pendingTimeoutResult)
	if err != nil {
		return nil, fmt.Errorf("统计待支付超时订单失败: %v", err)
	}

	stats.PendingTimeoutCount = pendingTimeoutResult.Count
	stats.PendingTimeoutAmount = pendingTimeoutResult.Amount

	// 统计处理中超时订单
	processingTimeoutQuery := fmt.Sprintf(`
		SELECT COUNT(*) as count, COALESCE(SUM(total_amount), 0) as amount 
		FROM orders 
		WHERE %s AND status = 'processing' 
		AND status_updated_at < DATE_SUB(NOW(), INTERVAL 24 HOUR)
	`, "tenant_id = ?")

	var processingTimeoutResult struct {
		Count  int     `db:"count"`
		Amount float64 `db:"amount"`
	}
	err = g.DB().Raw(processingTimeoutQuery, args[0]).Scan(&processingTimeoutResult)
	if err != nil {
		return nil, fmt.Errorf("统计处理中超时订单失败: %v", err)
	}

	stats.ProcessingTimeoutCount = processingTimeoutResult.Count
	stats.ProcessingTimeoutAmount = processingTimeoutResult.Amount

	// 统计今日自动取消的订单
	todayCancelledQuery := fmt.Sprintf(`
		SELECT COUNT(*) as count, COALESCE(SUM(total_amount), 0) as amount 
		FROM orders o
		INNER JOIN order_status_history h ON o.id = h.order_id
		WHERE o.tenant_id = ? AND o.status = 'cancelled' 
		AND h.to_status = 5 AND h.operator_type = 'system'
		AND h.created_at >= CURDATE()
	`)

	var todayCancelledResult struct {
		Count  int     `db:"count"`
		Amount float64 `db:"amount"`
	}
	err = g.DB().Raw(todayCancelledQuery, args[0]).Scan(&todayCancelledResult)
	if err != nil {
		return nil, fmt.Errorf("统计今日自动取消订单失败: %v", err)
	}

	stats.TodayAutoCancelledCount = todayCancelledResult.Count
	stats.TodayAutoCancelledAmount = todayCancelledResult.Amount

	return stats, nil
}

// orderStatusToString 将OrderStatusInt转换为字符串
func (s *OrderTimeoutService) orderStatusToString(status types.OrderStatusInt) string {
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

// stringToOrderStatusInt 将字符串转换为OrderStatusInt
func (s *OrderTimeoutService) stringToOrderStatusInt(status types.OrderStatus) types.OrderStatusInt {
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
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IInventoryService 库存服务接口
type IInventoryService interface {
	// 基础库存操作
	GetInventoryInfo(ctx context.Context, productID uint64) (*types.InventoryResponse, error)
	AdjustInventory(ctx context.Context, req *types.InventoryAdjustRequest) (*types.InventoryResponse, error)
	BatchAdjustInventory(ctx context.Context, req *types.BatchInventoryAdjustRequest) ([]types.InventoryResponse, error)
	
	// 库存预留操作
	ReserveInventory(ctx context.Context, req *types.InventoryReserveRequest) (*types.InventoryReservation, error)
	ReleaseInventory(ctx context.Context, req *types.InventoryReleaseRequest) error
	ConfirmReservation(ctx context.Context, reservationID uint64) error
	
	// 库存监控
	CheckLowStockAlerts(ctx context.Context) error
	ProcessExpiredReservations(ctx context.Context) error
}

// inventoryService 库存服务实现
type inventoryService struct {
	productRepo     *repository.ProductRepository
	recordRepo      repository.IInventoryRecordRepository
	reservationRepo repository.IInventoryReservationRepository
	alertRepo       repository.IInventoryAlertRepository
}

// NewInventoryService 创建库存服务实例
func NewInventoryService() IInventoryService {
	return &inventoryService{
		productRepo:     repository.NewProductRepository(),
		recordRepo:      repository.NewInventoryRecordRepository(),
		reservationRepo: repository.NewInventoryReservationRepository(),
		alertRepo:       repository.NewInventoryAlertRepository(),
	}
}

// GetInventoryInfo 获取库存信息
func (s *inventoryService) GetInventoryInfo(ctx context.Context, productID uint64) (*types.InventoryResponse, error) {
	if productID == 0 {
		return nil, errors.New("商品ID不能为空")
	}

	// 获取商品信息
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("获取商品信息失败: %w", err)
	}

	// 计算当前预留数量
	tenantID := getTenantIDFromContext(ctx)
	totalReserved, err := s.reservationRepo.GetTotalReservedQuantity(ctx, tenantID, productID)
	if err != nil {
		g.Log().Warningf(ctx, "获取预留数量失败: %v", err)
		totalReserved = 0
	}

	// 转换为扩展库存信息并更新预留数量
	extendedInfo := types.ExtendedInventoryInfo{
		StockQuantity:    product.InventoryInfo.StockQuantity,
		ReservedQuantity: totalReserved,
		TrackInventory:   product.InventoryInfo.TrackInventory,
	}

	response := &types.InventoryResponse{
		ProductID:      productID,
		InventoryInfo:  extendedInfo,
		AvailableStock: extendedInfo.AvailableQuantity(),
		ReservedStock:  totalReserved,
		IsLowStock:     extendedInfo.IsLowStock(),
		IsOutOfStock:   extendedInfo.IsOutOfStock(),
	}

	return response, nil
}

// AdjustInventory 调整库存
func (s *inventoryService) AdjustInventory(ctx context.Context, req *types.InventoryAdjustRequest) (*types.InventoryResponse, error) {
	if req == nil || req.ProductID == 0 {
		return nil, errors.New("请求参数不能为空")
	}

	// 计算调整数量
	var adjustment int
	switch req.AdjustmentType {
	case "increase":
		adjustment = req.Quantity
	case "decrease":
		adjustment = -req.Quantity
	case "set":
		// 获取当前商品信息
		product, err := s.productRepo.GetByID(ctx, req.ProductID)
		if err != nil {
			return nil, fmt.Errorf("获取当前库存失败: %w", err)
		}
		adjustment = req.Quantity - product.InventoryInfo.StockQuantity
	default:
		return nil, errors.New("调整类型不支持")
	}

	// 获取用户ID用于记录
	userID := getUserIDFromContext(ctx)
	tenantID := getTenantIDFromContext(ctx)

	// 执行库存调整（原子操作）
	newInventoryInfo, err := s.productRepo.AdjustInventory(ctx, req.ProductID, adjustment, req.Reason)
	if err != nil {
		return nil, fmt.Errorf("调整库存失败: %w", err)
	}

	// 记录库存变更历史
	record := &types.InventoryRecord{
		TenantID:        tenantID,
		ProductID:       req.ProductID,
		ChangeType:      types.InventoryChangeAdjustment,
		QuantityBefore:  newInventoryInfo.StockQuantity - adjustment,
		QuantityAfter:   newInventoryInfo.StockQuantity,
		QuantityChanged: adjustment,
		Reason:          req.Reason,
		ReferenceID:     req.ReferenceID,
		OperatedBy:      userID,
		CreatedAt:       time.Now(),
	}

	if err := s.recordRepo.Create(ctx, record); err != nil {
		g.Log().Errorf(ctx, "记录库存变更历史失败: %v", err)
	}

	// 检查预警
	go func() {
		if err := s.checkProductAlerts(ctx, req.ProductID); err != nil {
			g.Log().Errorf(ctx, "检查库存预警失败: %v", err)
		}
	}()

	// 返回最新库存信息
	return s.GetInventoryInfo(ctx, req.ProductID)
}

// BatchAdjustInventory 批量调整库存
func (s *inventoryService) BatchAdjustInventory(ctx context.Context, req *types.BatchInventoryAdjustRequest) ([]types.InventoryResponse, error) {
	if req == nil || len(req.Adjustments) == 0 {
		return nil, errors.New("批量调整请求不能为空")
	}

	if len(req.Adjustments) > 1000 {
		return nil, errors.New("批量调整数量不能超过1000个")
	}

	var results []types.InventoryResponse
	var errors []string

	// 逐个处理库存调整
	for i, adjustment := range req.Adjustments {
		adjustment.Reason = req.Reason // 使用统一的调整原因
		
		result, err := s.AdjustInventory(ctx, &adjustment)
		if err != nil {
			errors = append(errors, fmt.Sprintf("第%d个商品(ID:%d)调整失败: %v", i+1, adjustment.ProductID, err))
			continue
		}
		
		results = append(results, *result)
	}

	// 如果有错误，返回部分成功的结果和错误信息
	if len(errors) > 0 {
		return results, fmt.Errorf("批量调整部分失败: %v", errors)
	}

	return results, nil
}

// ReserveInventory 预留库存
func (s *inventoryService) ReserveInventory(ctx context.Context, req *types.InventoryReserveRequest) (*types.InventoryReservation, error) {
	if req == nil || req.ProductID == 0 || req.Quantity <= 0 {
		return nil, errors.New("预留请求参数不完整")
	}

	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)

	// 检查可用库存
	inventoryInfo, err := s.GetInventoryInfo(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("获取库存信息失败: %w", err)
	}

	if inventoryInfo.AvailableStock < req.Quantity {
		return nil, fmt.Errorf("可用库存不足: 需要%d，可用%d", req.Quantity, inventoryInfo.AvailableStock)
	}

	// 创建预留记录
	reservation := &types.InventoryReservation{
		TenantID:         tenantID,
		ProductID:        req.ProductID,
		ReservedQuantity: req.Quantity,
		ReferenceType:    req.ReferenceType,
		ReferenceID:      req.ReferenceID,
		Status:           types.ReservationStatusActive,
		ExpiresAt:        req.ExpiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 执行预留操作（原子操作）
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 预留库存
		if err := s.productRepo.ReserveInventory(ctx, req.ProductID, req.Quantity); err != nil {
			return err
		}

		// 创建预留记录
		if err := s.reservationRepo.Create(ctx, reservation); err != nil {
			return err
		}

		// 记录库存变更历史
		record := &types.InventoryRecord{
			TenantID:        tenantID,
			ProductID:       req.ProductID,
			ChangeType:      types.InventoryChangeReservation,
			QuantityBefore:  inventoryInfo.InventoryInfo.ReservedQuantity,
			QuantityAfter:   inventoryInfo.InventoryInfo.ReservedQuantity + req.Quantity,
			QuantityChanged: req.Quantity,
			Reason:          fmt.Sprintf("预留库存: %s-%s", req.ReferenceType, req.ReferenceID),
			ReferenceID:     &req.ReferenceID,
			OperatedBy:      userID,
			CreatedAt:       time.Now(),
		}

		return s.recordRepo.Create(ctx, record)
	})

	if err != nil {
		return nil, fmt.Errorf("库存预留失败: %w", err)
	}

	return reservation, nil
}

// ReleaseInventory 释放预留库存
func (s *inventoryService) ReleaseInventory(ctx context.Context, req *types.InventoryReleaseRequest) error {
	if req == nil || req.ReservationID == 0 {
		return errors.New("释放请求参数不完整")
	}

	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)

	// 获取预留记录
	reservation, err := s.reservationRepo.GetByID(ctx, tenantID, req.ReservationID)
	if err != nil {
		return fmt.Errorf("获取预留记录失败: %w", err)
	}

	if reservation.Status != types.ReservationStatusActive {
		return fmt.Errorf("预留记录状态无效: %s", reservation.Status)
	}

	// 执行释放操作（原子操作）
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 释放库存
		if err := s.productRepo.ReleaseInventory(ctx, reservation.ProductID, reservation.ReservedQuantity); err != nil {
			return err
		}

		// 更新预留状态
		if err := s.reservationRepo.UpdateStatus(ctx, tenantID, req.ReservationID, types.ReservationStatusReleased); err != nil {
			return err
		}

		// 记录库存变更历史
		record := &types.InventoryRecord{
			TenantID:        tenantID,
			ProductID:       reservation.ProductID,
			ChangeType:      types.InventoryChangeRelease,
			QuantityBefore:  0, // 将在事务提交后更新
			QuantityAfter:   0,
			QuantityChanged: -reservation.ReservedQuantity,
			Reason:          fmt.Sprintf("释放预留库存: %s-%s", reservation.ReferenceType, reservation.ReferenceID),
			ReferenceID:     &reservation.ReferenceID,
			OperatedBy:      userID,
			CreatedAt:       time.Now(),
		}

		return s.recordRepo.Create(ctx, record)
	})
}

// ConfirmReservation 确认预留（消费库存）
func (s *inventoryService) ConfirmReservation(ctx context.Context, reservationID uint64) error {
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)

	// 获取预留记录
	reservation, err := s.reservationRepo.GetByID(ctx, tenantID, reservationID)
	if err != nil {
		return fmt.Errorf("获取预留记录失败: %w", err)
	}

	if reservation.Status != types.ReservationStatusActive {
		return fmt.Errorf("预留记录状态无效: %s", reservation.Status)
	}

	// 执行确认操作（原子操作）
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 从库存中扣减数量
		_, err := s.productRepo.AdjustInventory(ctx, reservation.ProductID, -reservation.ReservedQuantity, "确认预留消费")
		if err != nil {
			return err
		}

		// 释放预留
		if err := s.productRepo.ReleaseInventory(ctx, reservation.ProductID, reservation.ReservedQuantity); err != nil {
			return err
		}

		// 更新预留状态
		if err := s.reservationRepo.UpdateStatus(ctx, tenantID, reservationID, types.ReservationStatusConfirmed); err != nil {
			return err
		}

		// 记录库存变更历史
		record := &types.InventoryRecord{
			TenantID:        tenantID,
			ProductID:       reservation.ProductID,
			ChangeType:      types.InventoryChangeSale,
			QuantityBefore:  0, // 将在事务提交后更新
			QuantityAfter:   0,
			QuantityChanged: -reservation.ReservedQuantity,
			Reason:          fmt.Sprintf("确认预留消费: %s-%s", reservation.ReferenceType, reservation.ReferenceID),
			ReferenceID:     &reservation.ReferenceID,
			OperatedBy:      userID,
			CreatedAt:       time.Now(),
		}

		return s.recordRepo.Create(ctx, record)
	})
}

// CheckLowStockAlerts 检查低库存预警
func (s *inventoryService) CheckLowStockAlerts(ctx context.Context) error {
	// 获取低库存商品
	products, err := s.productRepo.GetLowStockProducts(ctx)
	if err != nil {
		return fmt.Errorf("获取低库存商品失败: %w", err)
	}

	// 检查每个商品的预警规则
	for _, product := range products {
		if err := s.checkProductAlerts(ctx, product.ID); err != nil {
			g.Log().Errorf(ctx, "检查商品%d预警失败: %v", product.ID, err)
		}
	}

	return nil
}

// ProcessExpiredReservations 处理过期预留
func (s *inventoryService) ProcessExpiredReservations(ctx context.Context) error {
	tenantID := getTenantIDFromContext(ctx)
	
	// 获取过期预留
	expiredReservations, err := s.reservationRepo.GetExpiredReservations(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("获取过期预留失败: %w", err)
	}

	// 处理每个过期预留
	for _, reservation := range expiredReservations {
		if err := s.processExpiredReservation(ctx, &reservation); err != nil {
			g.Log().Errorf(ctx, "处理过期预留%d失败: %v", reservation.ID, err)
		}
	}

	return nil
}

// checkProductAlerts 检查商品预警
func (s *inventoryService) checkProductAlerts(ctx context.Context, productID uint64) error {
	tenantID := getTenantIDFromContext(ctx)
	
	// 获取商品的预警规则
	alerts, err := s.alertRepo.GetByProductID(ctx, tenantID, productID)
	if err != nil {
		return fmt.Errorf("获取预警规则失败: %w", err)
	}

	// 获取当前库存
	inventoryInfo, err := s.GetInventoryInfo(ctx, productID)
	if err != nil {
		return fmt.Errorf("获取库存信息失败: %w", err)
	}

	// 检查每个预警规则
	for _, alert := range alerts {
		if !alert.IsActive {
			continue
		}

		shouldTrigger := false
		switch alert.AlertType {
		case types.InventoryAlertTypeLowStock:
			shouldTrigger = inventoryInfo.AvailableStock <= alert.ThresholdValue
		case types.InventoryAlertTypeOutOfStock:
			shouldTrigger = inventoryInfo.AvailableStock <= 0
		case types.InventoryAlertTypeOverstock:
			shouldTrigger = inventoryInfo.InventoryInfo.StockQuantity >= alert.ThresholdValue
		}

		if shouldTrigger {
			if err := s.triggerAlert(ctx, &alert, inventoryInfo); err != nil {
				g.Log().Errorf(ctx, "触发预警失败: %v", err)
			}
		}
	}

	return nil
}

// triggerAlert 触发预警
func (s *inventoryService) triggerAlert(ctx context.Context, alert *types.InventoryAlert, inventoryInfo *types.InventoryResponse) error {
	// 检查是否在冷却期内（避免频繁发送）
	if alert.LastTriggeredAt != nil && time.Since(*alert.LastTriggeredAt) < time.Hour {
		return nil // 在冷却期内，跳过
	}

	// 发送通知（这里简化处理，实际应该调用通知服务）
	g.Log().Infof(ctx, "库存预警: 商品%d，类型%s，阈值%d，当前库存%d", 
		alert.ProductID, alert.AlertType, alert.ThresholdValue, inventoryInfo.AvailableStock)

	// 更新最后触发时间
	tenantID := getTenantIDFromContext(ctx)
	if err := s.alertRepo.UpdateLastTriggered(ctx, tenantID, alert.ID); err != nil {
		return fmt.Errorf("更新预警触发时间失败: %w", err)
	}

	return nil
}

// processExpiredReservation 处理单个过期预留
func (s *inventoryService) processExpiredReservation(ctx context.Context, reservation *types.InventoryReservation) error {
	tenantID := getTenantIDFromContext(ctx)
	
	// 执行释放操作
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 释放库存
		if err := s.productRepo.ReleaseInventory(ctx, reservation.ProductID, reservation.ReservedQuantity); err != nil {
			return err
		}

		// 更新预留状态
		if err := s.reservationRepo.UpdateStatus(ctx, tenantID, reservation.ID, types.ReservationStatusExpired); err != nil {
			return err
		}

		// 记录库存变更历史
		record := &types.InventoryRecord{
			TenantID:        tenantID,
			ProductID:       reservation.ProductID,
			ChangeType:      types.InventoryChangeRelease,
			QuantityBefore:  0,
			QuantityAfter:   0,
			QuantityChanged: -reservation.ReservedQuantity,
			Reason:          fmt.Sprintf("过期预留自动释放: %s-%s", reservation.ReferenceType, reservation.ReferenceID),
			ReferenceID:     &reservation.ReferenceID,
			OperatedBy:      0, // 系统操作
			CreatedAt:       time.Now(),
		}

		return s.recordRepo.Create(ctx, record)
	})
}

// 辅助函数
func getTenantIDFromContext(ctx context.Context) uint64 {
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(uint64); ok {
			return id
		}
	}
	return 0
}

func getUserIDFromContext(ctx context.Context) uint64 {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(uint64); ok {
			return id
		}
	}
	return 0
}
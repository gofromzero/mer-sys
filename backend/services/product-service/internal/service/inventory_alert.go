package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IInventoryAlertService 库存预警服务接口
type IInventoryAlertService interface {
	// 预警规则管理
	CreateAlert(ctx context.Context, req *types.InventoryAlertRequest) (*types.InventoryAlert, error)
	GetAlertsByProduct(ctx context.Context, productID uint64) ([]types.InventoryAlert, error)
	GetActiveAlerts(ctx context.Context) ([]types.InventoryAlert, error)
	UpdateAlert(ctx context.Context, alertID uint64, req *types.InventoryAlertRequest) error
	DeleteAlert(ctx context.Context, alertID uint64) error
	ToggleAlert(ctx context.Context, alertID uint64, isActive bool) error
	
	// 预警检查和处理
	CheckProductAlerts(ctx context.Context, productID uint64) error
	CheckAllLowStockAlerts(ctx context.Context) error
	ProcessTriggeredAlerts(ctx context.Context) error
	
	// 预警通知
	SendAlertNotification(ctx context.Context, alert *types.InventoryAlert, currentStock int) error
}

// inventoryAlertService 库存预警服务实现
type inventoryAlertService struct {
	alertRepo     repository.IInventoryAlertRepository
	productRepo   *repository.ProductRepository
	inventoryService IInventoryService
}

// NewInventoryAlertService 创建库存预警服务实例
func NewInventoryAlertService(inventoryService IInventoryService) IInventoryAlertService {
	return &inventoryAlertService{
		alertRepo:        repository.NewInventoryAlertRepository(),
		productRepo:      repository.NewProductRepository(),
		inventoryService: inventoryService,
	}
}

// CreateAlert 创建库存预警规则
func (s *inventoryAlertService) CreateAlert(ctx context.Context, req *types.InventoryAlertRequest) (*types.InventoryAlert, error) {
	if req == nil || req.ProductID == 0 {
		return nil, fmt.Errorf("预警请求参数不完整")
	}

	// 验证商品是否存在
	_, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("商品不存在: %w", err)
	}

	// 验证阈值合理性
	if req.ThresholdValue < 0 {
		return nil, fmt.Errorf("预警阈值不能为负数")
	}

	// 验证通知渠道
	if len(req.NotificationChannels) == 0 {
		return nil, fmt.Errorf("至少需要指定一个通知渠道")
	}

	tenantID := getTenantIDFromContext(ctx)
	alert := &types.InventoryAlert{
		TenantID:             tenantID,
		ProductID:            req.ProductID,
		AlertType:            req.AlertType,
		ThresholdValue:       req.ThresholdValue,
		NotificationChannels: req.NotificationChannels,
		IsActive:             req.IsActive,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = s.alertRepo.Create(ctx, alert)
	if err != nil {
		return nil, fmt.Errorf("创建预警规则失败: %w", err)
	}

	g.Log().Infof(ctx, "创建库存预警规则成功: 商品ID=%d, 类型=%s, 阈值=%d", 
		req.ProductID, req.AlertType, req.ThresholdValue)

	return alert, nil
}

// GetAlertsByProduct 获取商品的预警规则
func (s *inventoryAlertService) GetAlertsByProduct(ctx context.Context, productID uint64) ([]types.InventoryAlert, error) {
	if productID == 0 {
		return nil, fmt.Errorf("商品ID不能为空")
	}

	tenantID := getTenantIDFromContext(ctx)
	return s.alertRepo.GetByProductID(ctx, tenantID, productID)
}

// GetActiveAlerts 获取所有活跃的预警规则
func (s *inventoryAlertService) GetActiveAlerts(ctx context.Context) ([]types.InventoryAlert, error) {
	tenantID := getTenantIDFromContext(ctx)
	return s.alertRepo.GetActiveAlerts(ctx, tenantID)
}

// UpdateAlert 更新预警规则
func (s *inventoryAlertService) UpdateAlert(ctx context.Context, alertID uint64, req *types.InventoryAlertRequest) error {
	if alertID == 0 {
		return fmt.Errorf("预警规则ID不能为空")
	}

	if req == nil {
		return fmt.Errorf("更新请求不能为空")
	}

	tenantID := getTenantIDFromContext(ctx)
	
	// 获取现有的预警规则
	existingAlert, err := s.alertRepo.GetByID(ctx, tenantID, alertID)
	if err != nil {
		return fmt.Errorf("获取预警规则失败: %w", err)
	}

	// 更新字段
	existingAlert.AlertType = req.AlertType
	existingAlert.ThresholdValue = req.ThresholdValue
	existingAlert.NotificationChannels = req.NotificationChannels
	existingAlert.IsActive = req.IsActive
	existingAlert.UpdatedAt = time.Now()

	return s.alertRepo.Update(ctx, tenantID, existingAlert)
}

// DeleteAlert 删除预警规则
func (s *inventoryAlertService) DeleteAlert(ctx context.Context, alertID uint64) error {
	if alertID == 0 {
		return fmt.Errorf("预警规则ID不能为空")
	}

	tenantID := getTenantIDFromContext(ctx)
	return s.alertRepo.Delete(ctx, tenantID, alertID)
}

// ToggleAlert 切换预警规则状态
func (s *inventoryAlertService) ToggleAlert(ctx context.Context, alertID uint64, isActive bool) error {
	if alertID == 0 {
		return fmt.Errorf("预警规则ID不能为空")
	}

	tenantID := getTenantIDFromContext(ctx)
	return s.alertRepo.ToggleStatus(ctx, tenantID, alertID, isActive)
}

// CheckProductAlerts 检查单个商品的预警
func (s *inventoryAlertService) CheckProductAlerts(ctx context.Context, productID uint64) error {
	if productID == 0 {
		return fmt.Errorf("商品ID不能为空")
	}

	// 获取商品的预警规则
	alerts, err := s.GetAlertsByProduct(ctx, productID)
	if err != nil {
		return fmt.Errorf("获取预警规则失败: %w", err)
	}

	if len(alerts) == 0 {
		return nil // 没有预警规则，直接返回
	}

	// 获取当前库存信息
	inventoryInfo, err := s.inventoryService.GetInventoryInfo(ctx, productID)
	if err != nil {
		return fmt.Errorf("获取库存信息失败: %w", err)
	}

	// 检查每个预警规则
	for _, alert := range alerts {
		if !alert.IsActive {
			continue // 跳过未激活的预警
		}

		// 检查是否需要触发预警
		shouldTrigger := s.shouldTriggerAlert(&alert, inventoryInfo)
		if shouldTrigger {
			// 检查冷却期（避免频繁发送）
			if s.isInCooldown(&alert) {
				continue
			}

			// 发送预警通知
			if err := s.SendAlertNotification(ctx, &alert, inventoryInfo.AvailableStock); err != nil {
				g.Log().Errorf(ctx, "发送预警通知失败: %v", err)
			}

			// 更新最后触发时间
			s.alertRepo.UpdateLastTriggered(ctx, alert.TenantID, alert.ID)
		}
	}

	return nil
}

// CheckAllLowStockAlerts 检查所有低库存预警
func (s *inventoryAlertService) CheckAllLowStockAlerts(ctx context.Context) error {
	// 获取所有低库存商品
	products, err := s.productRepo.GetLowStockProducts(ctx)
	if err != nil {
		return fmt.Errorf("获取低库存商品失败: %w", err)
	}

	g.Log().Infof(ctx, "检查到%d个低库存商品", len(products))

	// 为每个低库存商品检查预警
	for _, product := range products {
		if err := s.CheckProductAlerts(ctx, product.ID); err != nil {
			g.Log().Errorf(ctx, "检查商品%d预警失败: %v", product.ID, err)
		}
	}

	return nil
}

// ProcessTriggeredAlerts 处理已触发的预警
func (s *inventoryAlertService) ProcessTriggeredAlerts(ctx context.Context) error {
	// 获取所有活跃的预警规则
	alerts, err := s.GetActiveAlerts(ctx)
	if err != nil {
		return fmt.Errorf("获取活跃预警失败: %w", err)
	}

	processedCount := 0
	for _, alert := range alerts {
		// 检查这个预警对应的商品
		if err := s.CheckProductAlerts(ctx, alert.ProductID); err != nil {
			g.Log().Errorf(ctx, "处理商品%d预警失败: %v", alert.ProductID, err)
		} else {
			processedCount++
		}
	}

	g.Log().Infof(ctx, "处理了%d个预警规则", processedCount)
	return nil
}

// SendAlertNotification 发送预警通知
func (s *inventoryAlertService) SendAlertNotification(ctx context.Context, alert *types.InventoryAlert, currentStock int) error {
	if alert == nil {
		return fmt.Errorf("预警信息不能为空")
	}

	// 构建通知消息
	message := s.buildAlertMessage(alert, currentStock)
	
	// 记录预警日志
	g.Log().Warningf(ctx, "库存预警触发: 商品ID=%d, 类型=%s, 当前库存=%d, 阈值=%d, 消息=%s", 
		alert.ProductID, alert.AlertType, currentStock, alert.ThresholdValue, message)

	// 根据通知渠道发送通知
	for _, channel := range alert.NotificationChannels {
		switch channel {
		case "system":
			// 系统内通知（记录到系统通知表）
			if err := s.sendSystemNotification(ctx, alert, message); err != nil {
				g.Log().Errorf(ctx, "发送系统通知失败: %v", err)
			}
		case "email":
			// 邮件通知（这里简化处理，实际应该调用邮件服务）
			g.Log().Infof(ctx, "邮件通知: %s", message)
		case "sms":
			// 短信通知（这里简化处理，实际应该调用短信服务）
			g.Log().Infof(ctx, "短信通知: %s", message)
		default:
			g.Log().Warningf(ctx, "未知的通知渠道: %s", channel)
		}
	}

	return nil
}

// shouldTriggerAlert 判断是否应该触发预警
func (s *inventoryAlertService) shouldTriggerAlert(alert *types.InventoryAlert, inventoryInfo *types.InventoryResponse) bool {
	switch alert.AlertType {
	case types.InventoryAlertTypeLowStock:
		return inventoryInfo.AvailableStock <= alert.ThresholdValue
	case types.InventoryAlertTypeOutOfStock:
		return inventoryInfo.AvailableStock <= 0
	case types.InventoryAlertTypeOverstock:
		return inventoryInfo.InventoryInfo.StockQuantity >= alert.ThresholdValue
	default:
		return false
	}
}

// isInCooldown 检查是否在冷却期内
func (s *inventoryAlertService) isInCooldown(alert *types.InventoryAlert) bool {
	if alert.LastTriggeredAt == nil {
		return false // 从未触发过，不在冷却期
	}

	// 默认冷却期为1小时
	cooldownPeriod := time.Hour
	return time.Since(*alert.LastTriggeredAt) < cooldownPeriod
}

// buildAlertMessage 构建预警消息
func (s *inventoryAlertService) buildAlertMessage(alert *types.InventoryAlert, currentStock int) string {
	switch alert.AlertType {
	case types.InventoryAlertTypeLowStock:
		return fmt.Sprintf("库存不足预警: 商品ID %d 当前库存 %d，低于预警阈值 %d", 
			alert.ProductID, currentStock, alert.ThresholdValue)
	case types.InventoryAlertTypeOutOfStock:
		return fmt.Sprintf("库存耗尽预警: 商品ID %d 当前库存为 %d，已无可用库存", 
			alert.ProductID, currentStock)
	case types.InventoryAlertTypeOverstock:
		return fmt.Sprintf("库存过多预警: 商品ID %d 当前库存 %d，超过预警阈值 %d", 
			alert.ProductID, currentStock, alert.ThresholdValue)
	default:
		return fmt.Sprintf("库存预警: 商品ID %d 当前库存 %d", alert.ProductID, currentStock)
	}
}

// sendSystemNotification 发送系统内通知
func (s *inventoryAlertService) sendSystemNotification(ctx context.Context, alert *types.InventoryAlert, message string) error {
	// 这里简化处理，实际应该调用通知服务创建系统通知
	// 可以考虑创建一个 Notification 实体并保存到数据库
	g.Log().Infof(ctx, "系统通知已创建: %s", message)
	return nil
}


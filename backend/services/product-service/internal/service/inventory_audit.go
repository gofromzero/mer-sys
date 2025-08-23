package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IInventoryAuditService 库存审计服务接口
type IInventoryAuditService interface {
	// 审计日志管理
	CreateAuditLog(ctx context.Context, req *types.AuditLogCreateRequest) error
	QueryAuditLogs(ctx context.Context, req *types.AuditLogQueryRequest) (*types.AuditLogResponse, error)
	GetResourceAuditLogs(ctx context.Context, resourceType string, resourceID uint64, limit int) ([]types.InventoryAuditLog, error)
	GetRecentAuditLogs(ctx context.Context, limit int) ([]types.InventoryAuditLog, error)
	
	// 便捷审计方法
	LogInventoryChange(ctx context.Context, productID uint64, operationType string, oldValue, newValue interface{}, description string) error
	LogAlertTriggered(ctx context.Context, alertID uint64, productID uint64, description string) error
	LogSystemOperation(ctx context.Context, resourceType string, resourceID uint64, operationType string, description string) error
	LogUserOperation(ctx context.Context, userID uint64, resourceType string, resourceID uint64, operationType string, description string) error
	
	// 统计和监控
	GetAuditStatistics(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error)
	CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error)
}

// inventoryAuditService 库存审计服务实现
type inventoryAuditService struct {
	auditRepo repository.IInventoryAuditRepository
}

// NewInventoryAuditService 创建库存审计服务实例
func NewInventoryAuditService() IInventoryAuditService {
	return &inventoryAuditService{
		auditRepo: repository.NewInventoryAuditRepository(),
	}
}

// CreateAuditLog 创建审计日志
func (s *inventoryAuditService) CreateAuditLog(ctx context.Context, req *types.AuditLogCreateRequest) error {
	if req == nil {
		return fmt.Errorf("审计日志请求不能为空")
	}

	// 序列化元数据
	var metadataJSON string
	if req.Metadata != nil && len(req.Metadata) > 0 {
		data, err := json.Marshal(req.Metadata)
		if err != nil {
			g.Log().Warningf(ctx, "序列化审计元数据失败: %v", err)
		} else {
			metadataJSON = string(data)
		}
	}

	auditLog := &types.InventoryAuditLog{
		AuditType:     req.AuditType,
		Level:         req.Level,
		ResourceType:  req.ResourceType,
		ResourceID:    req.ResourceID,
		OperationType: req.OperationType,
		OperatorID:    req.OperatorID,
		OperatorType:  req.OperatorType,
		Title:         req.Title,
		Description:   req.Description,
		OldValue:      req.OldValue,
		NewValue:      req.NewValue,
		Metadata:      metadataJSON,
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
	}

	return s.auditRepo.Create(ctx, auditLog)
}

// QueryAuditLogs 查询审计日志
func (s *inventoryAuditService) QueryAuditLogs(ctx context.Context, req *types.AuditLogQueryRequest) (*types.AuditLogResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("查询请求不能为空")
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.auditRepo.Query(ctx, req)
}

// GetResourceAuditLogs 获取资源的审计日志
func (s *inventoryAuditService) GetResourceAuditLogs(ctx context.Context, resourceType string, resourceID uint64, limit int) ([]types.InventoryAuditLog, error) {
	if resourceType == "" || resourceID == 0 {
		return nil, fmt.Errorf("资源类型和ID不能为空")
	}

	if limit <= 0 {
		limit = 50
	}

	return s.auditRepo.GetByResourceID(ctx, resourceType, resourceID, limit)
}

// GetRecentAuditLogs 获取最近的审计日志
func (s *inventoryAuditService) GetRecentAuditLogs(ctx context.Context, limit int) ([]types.InventoryAuditLog, error) {
	if limit <= 0 {
		limit = 100
	}

	return s.auditRepo.GetRecentLogs(ctx, limit)
}

// LogInventoryChange 记录库存变更审计日志
func (s *inventoryAuditService) LogInventoryChange(ctx context.Context, productID uint64, operationType string, oldValue, newValue interface{}, description string) error {
	oldJSON, _ := json.Marshal(oldValue)
	newJSON, _ := json.Marshal(newValue)

	req := &types.AuditLogCreateRequest{
		AuditType:     types.AuditTypeInventoryChange,
		Level:         types.AuditLevelInfo,
		ResourceType:  "product",
		ResourceID:    productID,
		OperationType: operationType,
		OperatorID:    getUserIDFromContextPtr(ctx),
		OperatorType:  s.getOperatorType(ctx),
		Title:         fmt.Sprintf("商品库存%s", operationType),
		Description:   description,
		OldValue:      string(oldJSON),
		NewValue:      string(newJSON),
		IPAddress:     s.getClientIP(ctx),
		UserAgent:     s.getUserAgent(ctx),
	}

	return s.CreateAuditLog(ctx, req)
}

// LogAlertTriggered 记录预警触发审计日志
func (s *inventoryAuditService) LogAlertTriggered(ctx context.Context, alertID uint64, productID uint64, description string) error {
	req := &types.AuditLogCreateRequest{
		AuditType:     types.AuditTypeAlertTriggered,
		Level:         types.AuditLevelWarning,
		ResourceType:  "alert",
		ResourceID:    alertID,
		OperationType: "triggered",
		OperatorType:  "system",
		Title:         "库存预警触发",
		Description:   description,
		Metadata: map[string]interface{}{
			"product_id": productID,
		},
		IPAddress: "system",
		UserAgent: "inventory-system",
	}

	return s.CreateAuditLog(ctx, req)
}

// LogSystemOperation 记录系统操作审计日志
func (s *inventoryAuditService) LogSystemOperation(ctx context.Context, resourceType string, resourceID uint64, operationType string, description string) error {
	req := &types.AuditLogCreateRequest{
		AuditType:     types.AuditTypeSystemOperation,
		Level:         types.AuditLevelInfo,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		OperationType: operationType,
		OperatorType:  "system",
		Title:         fmt.Sprintf("系统%s操作", operationType),
		Description:   description,
		IPAddress:     "system",
		UserAgent:     "inventory-system",
	}

	return s.CreateAuditLog(ctx, req)
}

// LogUserOperation 记录用户操作审计日志
func (s *inventoryAuditService) LogUserOperation(ctx context.Context, userID uint64, resourceType string, resourceID uint64, operationType string, description string) error {
	req := &types.AuditLogCreateRequest{
		AuditType:     types.AuditTypeUserOperation,
		Level:         types.AuditLevelInfo,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		OperationType: operationType,
		OperatorID:    &userID,
		OperatorType:  "user",
		Title:         fmt.Sprintf("用户%s操作", operationType),
		Description:   description,
		IPAddress:     s.getClientIP(ctx),
		UserAgent:     s.getUserAgent(ctx),
	}

	return s.CreateAuditLog(ctx, req)
}

// GetAuditStatistics 获取审计统计信息
func (s *inventoryAuditService) GetAuditStatistics(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	if startTime.IsZero() || endTime.IsZero() {
		return nil, fmt.Errorf("开始时间和结束时间不能为空")
	}

	if endTime.Before(startTime) {
		return nil, fmt.Errorf("结束时间不能早于开始时间")
	}

	return s.auditRepo.GetStatsByTimeRange(ctx, startTime, endTime)
}

// CleanupOldLogs 清理旧的审计日志
func (s *inventoryAuditService) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		return 0, fmt.Errorf("保留天数必须大于0")
	}

	beforeDate := time.Now().AddDate(0, 0, -retentionDays)
	deletedCount, err := s.auditRepo.DeleteOldLogs(ctx, beforeDate)
	if err != nil {
		return 0, fmt.Errorf("清理旧审计日志失败: %w", err)
	}

	// 记录清理操作
	if deletedCount > 0 {
		s.LogSystemOperation(ctx, "audit_log", 0, "cleanup", 
			fmt.Sprintf("清理了%d天前的%d条审计日志", retentionDays, deletedCount))
	}

	return deletedCount, nil
}

// getOperatorType 根据上下文判断操作类型
func (s *inventoryAuditService) getOperatorType(ctx context.Context) string {
	if userID := getUserIDFromContext(ctx); userID > 0 {
		return "user"
	}

	// 检查是否是API调用
	if req := s.getRequestFromContext(ctx); req != nil {
		return "api"
	}

	return "system"
}

// getClientIP 获取客户端IP地址
func (s *inventoryAuditService) getClientIP(ctx context.Context) string {
	if req := s.getRequestFromContext(ctx); req != nil {
		// 尝试从 X-Forwarded-For 头获取
		if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
			return forwarded
		}
		
		// 尝试从 X-Real-IP 头获取
		if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
			return realIP
		}
		
		// 从连接中获取
		if host, _, err := net.SplitHostPort(req.Host); err == nil {
			return host
		}
	}

	return "unknown"
}

// getUserAgent 获取用户代理信息
func (s *inventoryAuditService) getUserAgent(ctx context.Context) string {
	if req := s.getRequestFromContext(ctx); req != nil {
		return req.Header.Get("User-Agent")
	}
	return "unknown"
}

// getRequestFromContext 从上下文获取HTTP请求对象
func (s *inventoryAuditService) getRequestFromContext(ctx context.Context) *ghttp.Request {
	if req := ctx.Value("http_request"); req != nil {
		if httpReq, ok := req.(*ghttp.Request); ok {
			return httpReq
		}
	}
	return nil
}

// getUserIDFromContextPtr 从上下文获取用户ID指针
func getUserIDFromContextPtr(ctx context.Context) *uint64 {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(uint64); ok {
			return &id
		}
	}
	return nil
}
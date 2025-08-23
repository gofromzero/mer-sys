package repository

import (
	"context"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IInventoryAuditRepository 库存审计Repository接口
type IInventoryAuditRepository interface {
	Create(ctx context.Context, auditLog *types.InventoryAuditLog) error
	Query(ctx context.Context, req *types.AuditLogQueryRequest) (*types.AuditLogResponse, error)
	GetByResourceID(ctx context.Context, resourceType string, resourceID uint64, limit int) ([]types.InventoryAuditLog, error)
	GetRecentLogs(ctx context.Context, limit int) ([]types.InventoryAuditLog, error)
	GetStatsByTimeRange(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error)
	DeleteOldLogs(ctx context.Context, beforeDate time.Time) (int64, error)
}

// inventoryAuditRepository 库存审计Repository实现
type inventoryAuditRepository struct {
	*BaseRepository
	tableName string
}

// NewInventoryAuditRepository 创建库存审计Repository
func NewInventoryAuditRepository() IInventoryAuditRepository {
	return &inventoryAuditRepository{
		BaseRepository: NewBaseRepository(),
		tableName:     "inventory_audit_logs",
	}
}

// Create 创建审计日志
func (r *inventoryAuditRepository) Create(ctx context.Context, auditLog *types.InventoryAuditLog) error {
	if auditLog == nil {
		return errors.New("审计日志不能为空")
	}

	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return errors.New("租户ID不能为空")
	}

	auditLog.TenantID = tenantID
	auditLog.CreatedAt = time.Now()

	// 处理元数据JSON序列化
	var metadataJSON string
	if auditLog.Metadata != "" {
		metadataJSON = auditLog.Metadata
	}

	data := g.Map{
		"tenant_id":       auditLog.TenantID,
		"audit_type":      string(auditLog.AuditType),
		"level":           string(auditLog.Level),
		"resource_type":   auditLog.ResourceType,
		"resource_id":     auditLog.ResourceID,
		"operation_type":  auditLog.OperationType,
		"operator_id":     auditLog.OperatorID,
		"operator_type":   auditLog.OperatorType,
		"title":           auditLog.Title,
		"description":     auditLog.Description,
		"old_value":       auditLog.OldValue,
		"new_value":       auditLog.NewValue,
		"metadata":        metadataJSON,
		"ip_address":      auditLog.IPAddress,
		"user_agent":      auditLog.UserAgent,
		"created_at":      auditLog.CreatedAt,
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).Insert(data)
	if err != nil {
		g.Log().Errorf(ctx, "创建审计日志失败: %v", err)
		return err
	}

	// 获取插入的ID
	lastInsertID, err := result.LastInsertId()
	if err == nil {
		auditLog.ID = uint64(lastInsertID)
	}

	return nil
}

// Query 查询审计日志
func (r *inventoryAuditRepository) Query(ctx context.Context, req *types.AuditLogQueryRequest) (*types.AuditLogResponse, error) {
	if req == nil {
		return nil, errors.New("查询请求不能为空")
	}

	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, errors.New("租户ID不能为空")
	}

	// 构建查询条件
	db := g.DB().Model(r.tableName).Ctx(ctx).Where("tenant_id = ?", tenantID)

	if req.AuditType != nil {
		db = db.Where("audit_type = ?", string(*req.AuditType))
	}

	if req.Level != nil {
		db = db.Where("level = ?", string(*req.Level))
	}

	if req.ResourceType != nil {
		db = db.Where("resource_type = ?", *req.ResourceType)
	}

	if req.ResourceID != nil {
		db = db.Where("resource_id = ?", *req.ResourceID)
	}

	if req.OperatorID != nil {
		db = db.Where("operator_id = ?", *req.OperatorID)
	}

	if req.StartTime != nil {
		db = db.Where("created_at >= ?", *req.StartTime)
	}

	if req.EndTime != nil {
		db = db.Where("created_at <= ?", *req.EndTime)
	}

	// 计算总数
	total, err := db.Count()
	if err != nil {
		g.Log().Errorf(ctx, "查询审计日志总数失败: %v", err)
		return nil, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var logs []types.InventoryAuditLog

	err = db.Order("created_at DESC").
		Limit(req.PageSize).
		Offset(offset).
		Scan(&logs)

	if err != nil {
		g.Log().Errorf(ctx, "查询审计日志失败: %v", err)
		return nil, err
	}

	return &types.AuditLogResponse{
		Logs:     logs,
		Total:    int64(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetByResourceID 根据资源ID获取审计日志
func (r *inventoryAuditRepository) GetByResourceID(ctx context.Context, resourceType string, resourceID uint64, limit int) ([]types.InventoryAuditLog, error) {
	if resourceID == 0 {
		return nil, errors.New("资源ID不能为空")
	}

	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, errors.New("租户ID不能为空")
	}

	var logs []types.InventoryAuditLog
	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND resource_type = ? AND resource_id = ?", tenantID, resourceType, resourceID).
		Order("created_at DESC").
		Limit(limit).
		Scan(&logs)

	if err != nil {
		g.Log().Errorf(ctx, "查询资源审计日志失败: %v", err)
		return nil, err
	}

	return logs, nil
}

// GetRecentLogs 获取最近的审计日志
func (r *inventoryAuditRepository) GetRecentLogs(ctx context.Context, limit int) ([]types.InventoryAuditLog, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, errors.New("租户ID不能为空")
	}

	var logs []types.InventoryAuditLog
	err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Scan(&logs)

	if err != nil {
		g.Log().Errorf(ctx, "查询最近审计日志失败: %v", err)
		return nil, err
	}

	return logs, nil
}

// GetStatsByTimeRange 获取时间范围内的统计信息
func (r *inventoryAuditRepository) GetStatsByTimeRange(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, errors.New("租户ID不能为空")
	}

	stats := make(map[string]interface{})

	// 总日志数
	totalCount, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startTime, endTime).
		Count()
	if err != nil {
		return nil, err
	}
	stats["total_logs"] = totalCount

	// 按审计类型统计
	var typeStats []g.Map
	err = g.DB().Model(r.tableName).Ctx(ctx).
		Fields("audit_type, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startTime, endTime).
		Group("audit_type").
		Scan(&typeStats)
	if err != nil {
		return nil, err
	}
	stats["by_audit_type"] = typeStats

	// 按级别统计
	var levelStats []g.Map
	err = g.DB().Model(r.tableName).Ctx(ctx).
		Fields("level, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startTime, endTime).
		Group("level").
		Scan(&levelStats)
	if err != nil {
		return nil, err
	}
	stats["by_level"] = levelStats

	// 按操作类型统计
	var operationStats []g.Map
	err = g.DB().Model(r.tableName).Ctx(ctx).
		Fields("operation_type, COUNT(*) as count").
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startTime, endTime).
		Group("operation_type").
		Scan(&operationStats)
	if err != nil {
		return nil, err
	}
	stats["by_operation_type"] = operationStats

	return stats, nil
}

// DeleteOldLogs 删除旧的审计日志
func (r *inventoryAuditRepository) DeleteOldLogs(ctx context.Context, beforeDate time.Time) (int64, error) {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 {
		return 0, errors.New("租户ID不能为空")
	}

	result, err := g.DB().Model(r.tableName).Ctx(ctx).
		Where("tenant_id = ? AND created_at < ?", tenantID, beforeDate).
		Delete()

	if err != nil {
		g.Log().Errorf(ctx, "删除旧审计日志失败: %v", err)
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	g.Log().Infof(ctx, "删除了%d条旧审计日志", rowsAffected)

	return rowsAffected, nil
}
package types

import (
	"time"
)

// ExtendedInventoryInfo 扩展库存信息结构
type ExtendedInventoryInfo struct {
	StockQuantity      int     `json:"stock_quantity"`        // 实际库存数量
	ReservedQuantity   int     `json:"reserved_quantity"`     // 预留/锁定数量
	TrackInventory     bool    `json:"track_inventory"`       // 是否跟踪库存
	LowStockThreshold  *int    `json:"low_stock_threshold"`   // 低库存预警阈值
	ReorderPoint       *int    `json:"reorder_point"`         // 补货点
	ReorderQuantity    *int    `json:"reorder_quantity"`      // 补货数量
	CostPerUnit        *Money  `json:"cost_per_unit"`         // 单位成本
}

// AvailableQuantity 计算可用库存数量
func (ii *ExtendedInventoryInfo) AvailableQuantity() int {
	return ii.StockQuantity - ii.ReservedQuantity
}

// IsLowStock 检查是否低库存
func (ii *ExtendedInventoryInfo) IsLowStock() bool {
	if ii.LowStockThreshold == nil {
		return false
	}
	return ii.AvailableQuantity() <= *ii.LowStockThreshold
}

// IsOutOfStock 检查是否缺货
func (ii *ExtendedInventoryInfo) IsOutOfStock() bool {
	return ii.AvailableQuantity() <= 0
}

// InventoryChangeType 库存变更类型
type InventoryChangeType string

const (
	InventoryChangePurchase    InventoryChangeType = "purchase"    // 采购入库
	InventoryChangeSale        InventoryChangeType = "sale"        // 销售出库
	InventoryChangeAdjustment  InventoryChangeType = "adjustment"  // 盘点调整
	InventoryChangeTransfer    InventoryChangeType = "transfer"    // 调拨
	InventoryChangeDamage      InventoryChangeType = "damage"      // 损耗
	InventoryChangeReservation InventoryChangeType = "reservation" // 预留锁定
	InventoryChangeRelease     InventoryChangeType = "release"     // 释放预留
)

// InventoryRecord 库存变更记录实体
type InventoryRecord struct {
	ID             uint64              `json:"id" gorm:"primaryKey"`
	TenantID       uint64              `json:"tenant_id" gorm:"not null;index:idx_tenant_product"`
	ProductID      uint64              `json:"product_id" gorm:"not null;index:idx_tenant_product"`
	ChangeType     InventoryChangeType `json:"change_type" gorm:"not null;index:idx_change_type"`
	QuantityBefore int                 `json:"quantity_before" gorm:"not null"`
	QuantityAfter  int                 `json:"quantity_after" gorm:"not null"`
	QuantityChanged int                `json:"quantity_changed" gorm:"not null"`
	Reason         string              `json:"reason" gorm:"size:255;not null"`
	ReferenceID    *string             `json:"reference_id" gorm:"size:100;index:idx_reference_id"`
	OperatedBy     uint64              `json:"operated_by" gorm:"not null"`
	CreatedAt      time.Time           `json:"created_at" gorm:"index:idx_created_at"`
}

// TableName 设置表名
func (InventoryRecord) TableName() string {
	return "inventory_records"
}

// ReservationStatus 预留状态
type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "active"    // 活跃状态
	ReservationStatusConfirmed ReservationStatus = "confirmed" // 确认消费
	ReservationStatusReleased  ReservationStatus = "released"  // 释放预留
	ReservationStatusExpired   ReservationStatus = "expired"   // 过期释放
)

// InventoryReservation 库存锁定记录实体
type InventoryReservation struct {
	ID               uint64            `json:"id" gorm:"primaryKey"`
	TenantID         uint64            `json:"tenant_id" gorm:"not null;index:idx_tenant_product"`
	ProductID        uint64            `json:"product_id" gorm:"not null;index:idx_tenant_product"`
	ReservedQuantity int               `json:"reserved_quantity" gorm:"not null"`
	ReferenceType    string            `json:"reference_type" gorm:"size:50;not null;index:idx_reference"`
	ReferenceID      string            `json:"reference_id" gorm:"size:100;not null;index:idx_reference"`
	Status           ReservationStatus `json:"status" gorm:"size:20;not null;default:'active';index:idx_status_expires"`
	ExpiresAt        *time.Time        `json:"expires_at" gorm:"index:idx_status_expires"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// TableName 设置表名
func (InventoryReservation) TableName() string {
	return "inventory_reservations"
}

// IsExpired 检查是否已过期
func (ir *InventoryReservation) IsExpired() bool {
	return ir.ExpiresAt != nil && time.Now().After(*ir.ExpiresAt)
}

// InventoryAlertType 库存预警类型
type InventoryAlertType string

const (
	InventoryAlertTypeLowStock   InventoryAlertType = "low_stock"   // 低库存
	InventoryAlertTypeOutOfStock InventoryAlertType = "out_of_stock" // 缺货
	InventoryAlertTypeOverstock  InventoryAlertType = "overstock"   // 超储
)

// InventoryAlert 库存预警配置实体
type InventoryAlert struct {
	ID                    uint64    `json:"id" gorm:"primaryKey"`
	TenantID              uint64    `json:"tenant_id" gorm:"not null;index:idx_tenant_product"`
	ProductID             uint64    `json:"product_id" gorm:"not null;index:idx_tenant_product"`
	AlertType             InventoryAlertType `json:"alert_type" gorm:"size:50;not null;index:idx_alert_type"`
	ThresholdValue        int       `json:"threshold_value" gorm:"not null"`
	NotificationChannels  []string  `json:"notification_channels" gorm:"type:json;not null"`
	IsActive              bool      `json:"is_active" gorm:"default:true;index:idx_active"`
	LastTriggeredAt       *time.Time `json:"last_triggered_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// TableName 设置表名
func (InventoryAlert) TableName() string {
	return "inventory_alerts"
}

// InventoryAdjustRequest 库存调整请求
type InventoryAdjustRequest struct {
	ProductID       uint64              `json:"product_id" validate:"required"`
	AdjustmentType  string              `json:"adjustment_type" validate:"required,oneof=increase decrease set"`
	Quantity        int                 `json:"quantity" validate:"required"`
	Reason          string              `json:"reason" validate:"required,max=255"`
	ReferenceID     *string             `json:"reference_id"`
}

// InventoryQueryRequest 库存查询请求
type InventoryQueryRequest struct {
	ProductIDs []uint64 `json:"product_ids"`
	TenantID   uint64   `json:"tenant_id" validate:"required"`
}

// InventoryReserveRequest 库存预留请求
type InventoryReserveRequest struct {
	ProductID       uint64     `json:"product_id" validate:"required"`
	Quantity        int        `json:"quantity" validate:"required,min=1"`
	ReferenceType   string     `json:"reference_type" validate:"required"`
	ReferenceID     string     `json:"reference_id" validate:"required"`
	ExpiresAt       *time.Time `json:"expires_at"`
}

// InventoryReleaseRequest 库存释放请求
type InventoryReleaseRequest struct {
	ReservationID uint64 `json:"reservation_id" validate:"required"`
}

// BatchInventoryAdjustRequest 批量库存调整请求
type BatchInventoryAdjustRequest struct {
	Adjustments []InventoryAdjustRequest `json:"adjustments" validate:"required,min=1,max=1000"`
	Reason      string                   `json:"reason" validate:"required,max=255"`
}

// InventoryAlertRequest 库存预警配置请求
type InventoryAlertRequest struct {
	ProductID             uint64   `json:"product_id" validate:"required"`
	AlertType             InventoryAlertType `json:"alert_type" validate:"required"`
	ThresholdValue        int      `json:"threshold_value" validate:"required,min=0"`
	NotificationChannels  []string `json:"notification_channels" validate:"required,min=1"`
	IsActive              bool     `json:"is_active"`
}

// InventoryResponse 库存信息响应
type InventoryResponse struct {
	ProductID        uint64        `json:"product_id"`
	InventoryInfo    ExtendedInventoryInfo `json:"inventory_info"`
	AvailableStock   int           `json:"available_stock"`
	ReservedStock    int           `json:"reserved_stock"`
	IsLowStock       bool          `json:"is_low_stock"`
	IsOutOfStock     bool          `json:"is_out_of_stock"`
}

// InventoryRecordResponse 库存记录响应
type InventoryRecordResponse struct {
	Records    []InventoryRecord `json:"records"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// InventoryAuditLevel 审计级别
type InventoryAuditLevel string

const (
	AuditLevelInfo     InventoryAuditLevel = "info"     // 信息级别
	AuditLevelWarning  InventoryAuditLevel = "warning"  // 警告级别
	AuditLevelError    InventoryAuditLevel = "error"    // 错误级别
	AuditLevelCritical InventoryAuditLevel = "critical" // 严重级别
)

// InventoryAuditType 审计类型
type InventoryAuditType string

const (
	AuditTypeInventoryChange   InventoryAuditType = "inventory_change"   // 库存变更
	AuditTypeAlertTriggered    InventoryAuditType = "alert_triggered"    // 预警触发
	AuditTypeSystemOperation   InventoryAuditType = "system_operation"   // 系统操作
	AuditTypeUserOperation     InventoryAuditType = "user_operation"     // 用户操作
	AuditTypeApiAccess         InventoryAuditType = "api_access"         // API访问
	AuditTypeDataIntegrity     InventoryAuditType = "data_integrity"     // 数据完整性
)

// InventoryAuditLog 库存审计日志实体
type InventoryAuditLog struct {
	ID              uint64              `json:"id" gorm:"primaryKey"`
	TenantID        uint64              `json:"tenant_id" gorm:"not null;index:idx_tenant_time"`
	AuditType       InventoryAuditType  `json:"audit_type" gorm:"size:50;not null;index:idx_audit_type"`
	Level           InventoryAuditLevel `json:"level" gorm:"size:20;not null;index:idx_level"`
	ResourceType    string              `json:"resource_type" gorm:"size:50;not null"`    // product, alert, reservation
	ResourceID      uint64              `json:"resource_id" gorm:"not null;index:idx_resource"`
	OperationType   string              `json:"operation_type" gorm:"size:50;not null"`   // create, update, delete, check
	OperatorID      *uint64             `json:"operator_id" gorm:"index:idx_operator"`    // 操作人ID，系统操作时为null
	OperatorType    string              `json:"operator_type" gorm:"size:20;not null"`    // user, system, api
	Title           string              `json:"title" gorm:"size:200;not null"`           // 审计标题
	Description     string              `json:"description" gorm:"type:text"`             // 详细描述
	OldValue        string              `json:"old_value,omitempty" gorm:"type:text"`     // 变更前的值
	NewValue        string              `json:"new_value,omitempty" gorm:"type:text"`     // 变更后的值
	Metadata        string              `json:"metadata,omitempty" gorm:"type:json"`      // 扩展元数据
	IPAddress       string              `json:"ip_address" gorm:"size:45"`                // 操作IP
	UserAgent       string              `json:"user_agent" gorm:"size:500"`               // 用户代理
	CreatedAt       time.Time           `json:"created_at" gorm:"autoCreateTime;index:idx_tenant_time"`
}

// TableName 设置表名
func (InventoryAuditLog) TableName() string {
	return "inventory_audit_logs"
}

// InventoryStatistics 库存统计信息
type InventoryStatistics struct {
	TenantID             uint64    `json:"tenant_id"`
	TotalProducts        int       `json:"total_products"`        // 总商品数
	LowStockProducts     int       `json:"low_stock_products"`    // 低库存商品数
	OutOfStockProducts   int       `json:"out_of_stock_products"` // 缺货商品数
	TotalInventoryValue  float64   `json:"total_inventory_value"` // 库存总价值
	ActiveAlerts         int       `json:"active_alerts"`         // 活跃预警数
	TodayChanges         int       `json:"today_changes"`         // 今日库存变更次数
	LastUpdated          time.Time `json:"last_updated"`          // 最后更新时间
}

// InventoryTrend 库存趋势数据点
type InventoryTrend struct {
	Date         time.Time `json:"date"`          // 日期
	ProductID    uint64    `json:"product_id"`    // 商品ID
	StockLevel   int       `json:"stock_level"`   // 库存水平
	ChangeAmount int       `json:"change_amount"` // 变更数量
	ChangeType   string    `json:"change_type"`   // 变更类型
}

// InventoryMonitoringData 库存监控数据
type InventoryMonitoringData struct {
	Statistics      InventoryStatistics `json:"statistics"`
	RecentChanges   []InventoryRecord   `json:"recent_changes"`
	ActiveAlerts    []InventoryAlert    `json:"active_alerts"`
	TrendData       []InventoryTrend    `json:"trend_data"`
	LastUpdated     time.Time           `json:"last_updated"`
}

// AuditLogCreateRequest 审计日志创建请求
type AuditLogCreateRequest struct {
	AuditType       InventoryAuditType  `json:"audit_type" validate:"required"`
	Level           InventoryAuditLevel `json:"level" validate:"required"`
	ResourceType    string              `json:"resource_type" validate:"required"`
	ResourceID      uint64              `json:"resource_id" validate:"required"`
	OperationType   string              `json:"operation_type" validate:"required"`
	OperatorID      *uint64             `json:"operator_id"`
	OperatorType    string              `json:"operator_type" validate:"required"`
	Title           string              `json:"title" validate:"required,max=200"`
	Description     string              `json:"description"`
	OldValue        string              `json:"old_value,omitempty"`
	NewValue        string              `json:"new_value,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	IPAddress       string              `json:"ip_address"`
	UserAgent       string              `json:"user_agent"`
}

// AuditLogQueryRequest 审计日志查询请求
type AuditLogQueryRequest struct {
	Page         int                  `json:"page" validate:"min=1"`
	PageSize     int                  `json:"page_size" validate:"min=1,max=100"`
	AuditType    *InventoryAuditType  `json:"audit_type,omitempty"`
	Level        *InventoryAuditLevel `json:"level,omitempty"`
	ResourceType *string              `json:"resource_type,omitempty"`
	ResourceID   *uint64              `json:"resource_id,omitempty"`
	OperatorID   *uint64              `json:"operator_id,omitempty"`
	StartTime    *time.Time           `json:"start_time,omitempty"`
	EndTime      *time.Time           `json:"end_time,omitempty"`
}

// AuditLogResponse 审计日志响应
type AuditLogResponse struct {
	Logs     []InventoryAuditLog `json:"logs"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// 库存盘点相关类型定义
type StocktakingStatus string

const (
	StocktakingStatusPending    StocktakingStatus = "pending"     // 待开始
	StocktakingStatusInProgress StocktakingStatus = "in_progress" // 进行中
	StocktakingStatusCompleted  StocktakingStatus = "completed"   // 已完成
	StocktakingStatusCancelled  StocktakingStatus = "cancelled"   // 已取消
)

// InventoryStocktaking 库存盘点实体
type InventoryStocktaking struct {
	ID          uint64            `json:"id" gorm:"primaryKey"`
	TenantID    uint64            `json:"tenant_id" gorm:"not null;index:idx_tenant_status"`
	Name        string            `json:"name" gorm:"size:255;not null"`
	Description string            `json:"description" gorm:"type:text"`
	Status      StocktakingStatus `json:"status" gorm:"size:20;not null;default:'pending';index:idx_tenant_status"`
	ProductIDs  []uint64          `json:"product_ids" gorm:"type:json"` // 如果为空则盘点所有商品
	StartedBy   uint64            `json:"started_by" gorm:"not null"`
	CompletedBy *uint64           `json:"completed_by"`
	StartedAt   time.Time         `json:"started_at" gorm:"autoCreateTime"`
	CompletedAt *time.Time        `json:"completed_at"`
	Summary     string            `json:"summary" gorm:"type:text"`
	Notes       string            `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 设置表名
func (InventoryStocktaking) TableName() string {
	return "inventory_stocktaking"
}

// InventoryStocktakingRecord 盘点记录实体
type InventoryStocktakingRecord struct {
	ID            uint64    `json:"id" gorm:"primaryKey"`
	TenantID      uint64    `json:"tenant_id" gorm:"not null;index:idx_tenant_stocktaking"`
	StocktakingID uint64    `json:"stocktaking_id" gorm:"not null;index:idx_tenant_stocktaking"`
	ProductID     uint64    `json:"product_id" gorm:"not null;index:idx_product"`
	SystemCount   int       `json:"system_count" gorm:"not null"`    // 系统库存数量
	ActualCount   int       `json:"actual_count" gorm:"not null"`    // 实际盘点数量
	Difference    int       `json:"difference" gorm:"not null"`      // 差异数量
	Reason        string    `json:"reason" gorm:"size:500"`          // 差异原因
	CheckedBy     uint64    `json:"checked_by" gorm:"not null"`      // 盘点人
	CheckedAt     time.Time `json:"checked_at" gorm:"autoCreateTime"` // 盘点时间
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 设置表名
func (InventoryStocktakingRecord) TableName() string {
	return "inventory_stocktaking_records"
}

// 盘点相关请求类型
type StocktakingCreateRequest struct {
	Name        string   `json:"name" validate:"required,max=255"`
	Description string   `json:"description"`
	ProductIDs  []uint64 `json:"product_ids"` // 如果为空则盘点所有商品
	StartTime   *string  `json:"start_time"`  // 盘点开始时间，为空则立即开始
}

type StocktakingRecordRequest struct {
	ProductID   uint64 `json:"product_id" validate:"required"`
	ActualCount int    `json:"actual_count" validate:"required,min=0"`
	SystemCount int    `json:"system_count" validate:"required,min=0"`
	Reason      string `json:"reason" validate:"max=500"`
}

type StocktakingBatchUpdateRequest struct {
	Records []StocktakingRecordRequest `json:"records" validate:"required,min=1"`
}

type StocktakingCompleteRequest struct {
	Summary string `json:"summary" validate:"max=2000"`
	Notes   string `json:"notes" validate:"max=5000"`
}

// 盘点响应类型
type StocktakingResponse struct {
	Stocktaking InventoryStocktaking         `json:"stocktaking"`
	Records     []InventoryStocktakingRecord `json:"records,omitempty"`
	Statistics  *StocktakingStatistics       `json:"statistics,omitempty"`
}

type StocktakingStatistics struct {
	TotalProducts   int `json:"total_products"`   // 总盘点商品数
	CheckedProducts int `json:"checked_products"` // 已盘点商品数
	DifferenceCount int `json:"difference_count"` // 有差异的商品数
	TotalDifference int `json:"total_difference"` // 总差异数量
}

type StocktakingListResponse struct {
	Stocktakings []InventoryStocktaking `json:"stocktakings"`
	Total        int64                  `json:"total"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"page_size"`
}
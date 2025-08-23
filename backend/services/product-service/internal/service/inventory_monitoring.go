package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// IInventoryMonitoringService 库存监控服务接口
type IInventoryMonitoringService interface {
	// 统计信息
	GetInventoryStatistics(ctx context.Context) (*types.InventoryStatistics, error)
	GetInventoryTrends(ctx context.Context, days int) ([]types.InventoryTrend, error)
	GetMonitoringData(ctx context.Context) (*types.InventoryMonitoringData, error)
	
	// 数据分析
	AnalyzeInventoryHealth(ctx context.Context) (map[string]interface{}, error)
	GetLowStockSummary(ctx context.Context) (map[string]interface{}, error)
	GetInventoryValueAnalysis(ctx context.Context) (map[string]interface{}, error)
	
	// 自动化任务
	RunDailyInventoryCheck(ctx context.Context) error
	ProcessExpiredReservations(ctx context.Context) error
	GenerateInventoryReport(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error)
}

// inventoryMonitoringService 库存监控服务实现
type inventoryMonitoringService struct {
	productRepo     *repository.ProductRepository
	recordRepo      repository.IInventoryRecordRepository
	reservationRepo repository.IInventoryReservationRepository
	alertRepo       repository.IInventoryAlertRepository
	auditRepo       repository.IInventoryAuditRepository
	inventoryService IInventoryService
	alertService    IInventoryAlertService
	auditService    IInventoryAuditService
}

// NewInventoryMonitoringService 创建库存监控服务实例
func NewInventoryMonitoringService() IInventoryMonitoringService {
	inventoryService := NewInventoryService()
	return &inventoryMonitoringService{
		productRepo:      repository.NewProductRepository(),
		recordRepo:       repository.NewInventoryRecordRepository(),
		reservationRepo:  repository.NewInventoryReservationRepository(),
		alertRepo:        repository.NewInventoryAlertRepository(),
		auditRepo:        repository.NewInventoryAuditRepository(),
		inventoryService: inventoryService,
		alertService:     NewInventoryAlertService(inventoryService),
		auditService:     NewInventoryAuditService(),
	}
}

// GetInventoryStatistics 获取库存统计信息
func (s *inventoryMonitoringService) GetInventoryStatistics(ctx context.Context) (*types.InventoryStatistics, error) {
	tenantID := getTenantIDFromContext(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("租户ID不能为空")
	}

	stats := &types.InventoryStatistics{
		TenantID:    tenantID,
		LastUpdated: time.Now(),
	}

	// 获取总商品数
	totalProducts, err := s.getTotalProductCount(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "获取总商品数失败: %v", err)
	}
	stats.TotalProducts = totalProducts

	// 获取低库存商品数
	lowStockProducts, err := s.productRepo.GetLowStockProducts(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "获取低库存商品失败: %v", err)
	}
	stats.LowStockProducts = len(lowStockProducts)

	// 计算缺货商品数
	outOfStockCount := 0
	for _, product := range lowStockProducts {
		if product.InventoryInfo != nil && product.InventoryInfo.AvailableStock() <= 0 {
			outOfStockCount++
		}
	}
	stats.OutOfStockProducts = outOfStockCount

	// 获取活跃预警数
	activeAlerts, err := s.alertService.GetActiveAlerts(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "获取活跃预警失败: %v", err)
	}
	stats.ActiveAlerts = len(activeAlerts)

	// 获取今日库存变更次数
	todayChanges, err := s.getTodayInventoryChanges(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "获取今日变更次数失败: %v", err)
	}
	stats.TodayChanges = todayChanges

	// 计算库存总价值（简化计算）
	totalValue, err := s.calculateTotalInventoryValue(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "计算库存总价值失败: %v", err)
	}
	stats.TotalInventoryValue = totalValue

	return stats, nil
}

// GetInventoryTrends 获取库存趋势数据
func (s *inventoryMonitoringService) GetInventoryTrends(ctx context.Context, days int) ([]types.InventoryTrend, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	tenantID := getTenantIDFromContext(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("租户ID不能为空")
	}

	// 从库存记录表获取趋势数据
	records, err := s.recordRepo.GetRecentRecords(ctx, tenantID, days*10) // 获取更多记录用于分析
	if err != nil {
		return nil, fmt.Errorf("获取库存记录失败: %w", err)
	}

	// 按日期聚合数据
	trendMap := make(map[string]*types.InventoryTrend)
	
	for _, record := range records {
		dateKey := record.CreatedAt.Format("2006-01-02")
		
		if trend, exists := trendMap[dateKey]; exists {
			// 更新现有趋势数据
			trend.ChangeAmount += absInt(record.QuantityChanged)
		} else {
			// 创建新的趋势数据点
			trendMap[dateKey] = &types.InventoryTrend{
				Date:         record.CreatedAt,
				ProductID:    record.ProductID,
				StockLevel:   record.QuantityAfter,
				ChangeAmount: absInt(record.QuantityChanged),
				ChangeType:   string(record.ChangeType),
			}
		}
	}

	// 转换为切片并排序
	var trends []types.InventoryTrend
	for _, trend := range trendMap {
		trends = append(trends, *trend)
	}

	return trends, nil
}

// GetMonitoringData 获取完整的监控数据
func (s *inventoryMonitoringService) GetMonitoringData(ctx context.Context) (*types.InventoryMonitoringData, error) {
	data := &types.InventoryMonitoringData{
		LastUpdated: time.Now(),
	}

	// 获取统计信息
	stats, err := s.GetInventoryStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取统计信息失败: %w", err)
	}
	data.Statistics = *stats

	// 获取最近的变更记录
	recentChanges, err := s.auditService.GetRecentAuditLogs(ctx, 50)
	if err != nil {
		g.Log().Warningf(ctx, "获取最近变更记录失败: %v", err)
	}
	
	// 转换审计日志为库存记录格式（简化处理）
	var inventoryRecords []types.InventoryRecord
	for _, auditLog := range recentChanges {
		if auditLog.AuditType == types.AuditTypeInventoryChange {
			inventoryRecords = append(inventoryRecords, types.InventoryRecord{
				ID:              auditLog.ID,
				TenantID:        auditLog.TenantID,
				ProductID:       auditLog.ResourceID,
				ChangeType:      types.InventoryChangeType(auditLog.OperationType),
				Reason:          auditLog.Description,
				OperatedBy:      getUserIDFromContextValue(auditLog.OperatorID),
				CreatedAt:       auditLog.CreatedAt,
			})
		}
	}
	data.RecentChanges = inventoryRecords

	// 获取活跃预警
	activeAlerts, err := s.alertService.GetActiveAlerts(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "获取活跃预警失败: %v", err)
	}
	data.ActiveAlerts = activeAlerts

	// 获取趋势数据
	trendData, err := s.GetInventoryTrends(ctx, 30)
	if err != nil {
		g.Log().Warningf(ctx, "获取趋势数据失败: %v", err)
	}
	data.TrendData = trendData

	return data, nil
}

// AnalyzeInventoryHealth 分析库存健康状况
func (s *inventoryMonitoringService) AnalyzeInventoryHealth(ctx context.Context) (map[string]interface{}, error) {
	analysis := make(map[string]interface{})

	// 获取统计信息
	stats, err := s.GetInventoryStatistics(ctx)
	if err != nil {
		return nil, err
	}

	// 计算健康分数
	healthScore := s.calculateHealthScore(stats)
	analysis["health_score"] = healthScore
	analysis["health_level"] = s.getHealthLevel(healthScore)

	// 库存分布分析
	distribution := map[string]interface{}{
		"normal_stock":      stats.TotalProducts - stats.LowStockProducts,
		"low_stock":         stats.LowStockProducts,
		"out_of_stock":      stats.OutOfStockProducts,
		"low_stock_ratio":   float64(stats.LowStockProducts) / float64(max(stats.TotalProducts, 1)),
		"out_of_stock_ratio": float64(stats.OutOfStockProducts) / float64(max(stats.TotalProducts, 1)),
	}
	analysis["stock_distribution"] = distribution

	// 预警分析
	alertAnalysis := map[string]interface{}{
		"total_alerts":    stats.ActiveAlerts,
		"alert_coverage":  float64(stats.ActiveAlerts) / float64(max(stats.TotalProducts, 1)),
		"need_attention":  stats.LowStockProducts > 0 || stats.OutOfStockProducts > 0,
	}
	analysis["alert_analysis"] = alertAnalysis

	// 操作频率分析
	operationAnalysis := map[string]interface{}{
		"daily_changes":     stats.TodayChanges,
		"change_frequency":  s.getChangeFrequency(stats.TodayChanges),
	}
	analysis["operation_analysis"] = operationAnalysis

	return analysis, nil
}

// GetLowStockSummary 获取低库存汇总
func (s *inventoryMonitoringService) GetLowStockSummary(ctx context.Context) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// 获取低库存商品
	lowStockProducts, err := s.productRepo.GetLowStockProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取低库存商品失败: %w", err)
	}

	var criticalProducts []map[string]interface{}
	var warningProducts []map[string]interface{}

	for _, product := range lowStockProducts {
		productInfo := map[string]interface{}{
			"id":               product.ID,
			"name":            product.Name,
			"current_stock":   product.InventoryInfo.AvailableStock(),
			"reserved_stock":  product.InventoryInfo.ReservedQuantity,
		}

		if product.InventoryInfo.AvailableStock() <= 0 {
			criticalProducts = append(criticalProducts, productInfo)
		} else {
			warningProducts = append(warningProducts, productInfo)
		}
	}

	summary["total_low_stock"] = len(lowStockProducts)
	summary["critical_products"] = criticalProducts
	summary["warning_products"] = warningProducts
	summary["critical_count"] = len(criticalProducts)
	summary["warning_count"] = len(warningProducts)

	return summary, nil
}

// GetInventoryValueAnalysis 获取库存价值分析
func (s *inventoryMonitoringService) GetInventoryValueAnalysis(ctx context.Context) (map[string]interface{}, error) {
	analysis := make(map[string]interface{})

	// 这里简化处理，实际项目中需要根据商品成本价计算
	totalValue, err := s.calculateTotalInventoryValue(ctx)
	if err != nil {
		return nil, err
	}

	analysis["total_inventory_value"] = totalValue
	analysis["currency"] = "CNY"
	analysis["calculation_time"] = time.Now()

	// 可以进一步扩展：按分类统计、按商户统计等
	analysis["note"] = "价值计算基于商品售价，实际库存价值应基于成本价"

	return analysis, nil
}

// RunDailyInventoryCheck 执行每日库存检查
func (s *inventoryMonitoringService) RunDailyInventoryCheck(ctx context.Context) error {
	g.Log().Infof(ctx, "开始执行每日库存检查任务")

	// 检查所有低库存预警
	if err := s.alertService.CheckAllLowStockAlerts(ctx); err != nil {
		g.Log().Errorf(ctx, "检查低库存预警失败: %v", err)
	}

	// 处理过期预留
	if err := s.ProcessExpiredReservations(ctx); err != nil {
		g.Log().Errorf(ctx, "处理过期预留失败: %v", err)
	}

	// 记录检查结果
	stats, _ := s.GetInventoryStatistics(ctx)
	description := fmt.Sprintf("每日库存检查完成：总商品%d个，低库存%d个，缺货%d个，活跃预警%d个",
		stats.TotalProducts, stats.LowStockProducts, stats.OutOfStockProducts, stats.ActiveAlerts)

	s.auditService.LogSystemOperation(ctx, "inventory", 0, "daily_check", description)

	g.Log().Infof(ctx, "每日库存检查任务完成")
	return nil
}

// ProcessExpiredReservations 处理过期预留
func (s *inventoryMonitoringService) ProcessExpiredReservations(ctx context.Context) error {
	return s.inventoryService.ProcessExpiredReservations(ctx)
}

// GenerateInventoryReport 生成库存报告
func (s *inventoryMonitoringService) GenerateInventoryReport(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	report := make(map[string]interface{})
	report["report_period"] = map[string]interface{}{
		"start_date": startDate,
		"end_date":   endDate,
	}
	report["generated_at"] = time.Now()

	// 获取当前统计信息
	stats, err := s.GetInventoryStatistics(ctx)
	if err != nil {
		return nil, err
	}
	report["current_statistics"] = stats

	// 获取健康分析
	healthAnalysis, err := s.AnalyzeInventoryHealth(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "获取健康分析失败: %v", err)
	}
	report["health_analysis"] = healthAnalysis

	// 获取审计统计
	auditStats, err := s.auditService.GetAuditStatistics(ctx, startDate, endDate)
	if err != nil {
		g.Log().Warningf(ctx, "获取审计统计失败: %v", err)
	}
	report["audit_statistics"] = auditStats

	return report, nil
}

// 辅助函数

// getTotalProductCount 获取总商品数
func (s *inventoryMonitoringService) getTotalProductCount(ctx context.Context) (int, error) {
	// 这里简化处理，实际需要查询数据库
	return 100, nil // 模拟数据
}

// getTodayInventoryChanges 获取今日库存变更次数
func (s *inventoryMonitoringService) getTodayInventoryChanges(ctx context.Context) (int, error) {
	// 这里简化处理，实际需要查询审计日志
	return 25, nil // 模拟数据
}

// calculateTotalInventoryValue 计算库存总价值
func (s *inventoryMonitoringService) calculateTotalInventoryValue(ctx context.Context) (float64, error) {
	// 这里简化处理，实际需要根据商品价格和库存数量计算
	return 100000.00, nil // 模拟数据
}

// calculateHealthScore 计算健康分数
func (s *inventoryMonitoringService) calculateHealthScore(stats *types.InventoryStatistics) int {
	if stats.TotalProducts == 0 {
		return 100
	}

	score := 100
	
	// 低库存扣分
	lowStockRatio := float64(stats.LowStockProducts) / float64(stats.TotalProducts)
	score -= int(lowStockRatio * 30)

	// 缺货扣分
	outOfStockRatio := float64(stats.OutOfStockProducts) / float64(stats.TotalProducts)
	score -= int(outOfStockRatio * 50)

	if score < 0 {
		score = 0
	}

	return score
}

// getHealthLevel 根据分数获取健康级别
func (s *inventoryMonitoringService) getHealthLevel(score int) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 70:
		return "good"
	case score >= 50:
		return "fair"
	default:
		return "poor"
	}
}

// getChangeFrequency 获取变更频率描述
func (s *inventoryMonitoringService) getChangeFrequency(changes int) string {
	switch {
	case changes >= 50:
		return "high"
	case changes >= 20:
		return "medium"
	case changes >= 5:
		return "low"
	default:
		return "very_low"
	}
}


func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getUserIDFromContextValue(operatorID *uint64) uint64 {
	if operatorID != nil {
		return *operatorID
	}
	return 0
}
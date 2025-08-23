package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// PriceAuditRepository 价格审计仓储
type PriceAuditRepository struct {
	*BaseRepository
}

// NewPriceAuditRepository 创建价格审计仓储实例
func NewPriceAuditRepository() *PriceAuditRepository {
	baseRepo := NewBaseRepository()
	baseRepo.tableName = "price_audit_events"
	
	return &PriceAuditRepository{
		BaseRepository: baseRepo,
	}
}

// CreateAuditEvent 创建审计事件
func (r *PriceAuditRepository) CreateAuditEvent(ctx context.Context, req *types.CreatePriceAuditEventRequest) (*types.PriceAuditEvent, error) {
	event := &types.PriceAuditEvent{
		EventType:   req.EventType,
		ProductID:   req.ProductID,
		UserID:      r.GetUserID(ctx),
		EventData:   req.EventData,
		ImpactLevel: req.ImpactLevel,
		Description: req.Description,
		ClientIP:    req.ClientIP,
		UserAgent:   req.UserAgent,
		SessionID:   req.SessionID,
		CreatedAt:   time.Now(),
	}
	
	id, err := r.InsertAndGetId(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("创建价格审计事件失败: %w", err)
	}
	
	event.ID = uint64(id)
	event.TenantID = r.GetTenantID(ctx)
	
	return event, nil
}

// GetAuditEventsByQuery 根据查询条件获取审计事件
func (r *PriceAuditRepository) GetAuditEventsByQuery(ctx context.Context, query *types.PriceAuditQuery) ([]*types.PriceAuditEvent, int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	// 应用查询条件
	if !query.StartDate.IsZero() {
		model = model.Where("created_at >= ?", query.StartDate)
	}
	if !query.EndDate.IsZero() {
		model = model.Where("created_at <= ?", query.EndDate)
	}
	if len(query.EventTypes) > 0 {
		model = model.WhereIn("event_type", query.EventTypes)
	}
	if len(query.ProductIDs) > 0 {
		model = model.WhereIn("product_id", query.ProductIDs)
	}
	if len(query.UserIDs) > 0 {
		model = model.WhereIn("user_id", query.UserIDs)
	}
	if len(query.ImpactLevels) > 0 {
		model = model.WhereIn("impact_level", query.ImpactLevels)
	}
	
	// 获取总数
	total, err := model.Clone().Count()
	if err != nil {
		return nil, 0, fmt.Errorf("获取审计事件总数失败: %w", err)
	}
	
	// 排序
	orderBy := "created_at DESC"
	if query.OrderBy != "" {
		orderDir := "ASC"
		if query.OrderDir == "desc" {
			orderDir = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", query.OrderBy, orderDir)
	}
	
	// 分页查询
	result, err := model.Order(orderBy).
		Page(query.Page, query.PageSize).
		All()
	if err != nil {
		return nil, 0, fmt.Errorf("获取审计事件失败: %w", err)
	}
	
	var events []*types.PriceAuditEvent
	for _, record := range result {
		event := &types.PriceAuditEvent{}
		err = record.Struct(event)
		if err != nil {
			return nil, 0, fmt.Errorf("审计事件数据转换失败: %w", err)
		}
		events = append(events, event)
	}
	
	return events, total, nil
}

// GetAuditEventsByProduct 获取商品的审计事件
func (r *PriceAuditRepository) GetAuditEventsByProduct(ctx context.Context, productID uint64, limit int) ([]*types.PriceAuditEvent, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	query := model.Where("product_id", productID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	result, err := query.All()
	if err != nil {
		return nil, fmt.Errorf("获取商品审计事件失败: %w", err)
	}
	
	var events []*types.PriceAuditEvent
	for _, record := range result {
		event := &types.PriceAuditEvent{}
		err = record.Struct(event)
		if err != nil {
			return nil, fmt.Errorf("审计事件数据转换失败: %w", err)
		}
		events = append(events, event)
	}
	
	return events, nil
}

// GetAuditEventsByUser 获取用户的审计事件
func (r *PriceAuditRepository) GetAuditEventsByUser(ctx context.Context, userID uint64, limit int) ([]*types.PriceAuditEvent, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	query := model.Where("user_id", userID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	result, err := query.All()
	if err != nil {
		return nil, fmt.Errorf("获取用户审计事件失败: %w", err)
	}
	
	var events []*types.PriceAuditEvent
	for _, record := range result {
		event := &types.PriceAuditEvent{}
		err = record.Struct(event)
		if err != nil {
			return nil, fmt.Errorf("审计事件数据转换失败: %w", err)
		}
		events = append(events, event)
	}
	
	return events, nil
}

// GetAuditStatistics 获取审计统计信息
func (r *PriceAuditRepository) GetAuditStatistics(ctx context.Context, days int) (*types.PriceAuditSummary, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	since := time.Now().AddDate(0, 0, -days)
	
	// 总事件数
	total, err := model.Clone().Where("created_at >= ?", since).Count()
	if err != nil {
		return nil, fmt.Errorf("获取总事件数失败: %w", err)
	}
	
	// 按事件类型统计
	eventTypeResult, err := model.Clone().
		Fields("event_type, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("event_type").
		All()
	if err != nil {
		return nil, fmt.Errorf("按事件类型统计失败: %w", err)
	}
	
	eventsByType := make(map[types.PriceAuditEventType]int)
	for _, record := range eventTypeResult {
		eventType := types.PriceAuditEventType(record["event_type"].String())
		count := record["count"].Int()
		eventsByType[eventType] = count
	}
	
	// 按风险等级统计
	riskLevelResult, err := model.Clone().
		Fields("impact_level, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("impact_level").
		All()
	if err != nil {
		return nil, fmt.Errorf("按风险等级统计失败: %w", err)
	}
	
	eventsByRiskLevel := make(map[types.ImpactRiskLevel]int)
	for _, record := range riskLevelResult {
		riskLevel := types.ImpactRiskLevel(record["impact_level"].String())
		count := record["count"].Int()
		eventsByRiskLevel[riskLevel] = count
	}
	
	// 受影响的唯一商品数
	uniqueProducts, err := model.Clone().
		Fields("COUNT(DISTINCT product_id) as count").
		Where("created_at >= ?", since).
		Value()
	if err != nil {
		return nil, fmt.Errorf("获取唯一商品数失败: %w", err)
	}
	
	// 涉及的唯一用户数
	uniqueUsers, err := model.Clone().
		Fields("COUNT(DISTINCT user_id) as count").
		Where("created_at >= ?", since).
		Value()
	if err != nil {
		return nil, fmt.Errorf("获取唯一用户数失败: %w", err)
	}
	
	// 计算合规评分（简化版）
	complianceScore := r.calculateComplianceScore(eventsByRiskLevel, total)
	
	summary := &types.PriceAuditSummary{
		TotalEvents:            total,
		EventsByType:           eventsByType,
		EventsByRiskLevel:      eventsByRiskLevel,
		UniqueProductsAffected: uniqueProducts.Int(),
		UniqueUsersInvolved:    uniqueUsers.Int(),
		ComplianceScore:        complianceScore,
		TotalRevenueImpact:     types.Money{Amount: 0, Currency: "CNY"}, // 需要更复杂的计算
	}
	
	return summary, nil
}

// calculateComplianceScore 计算合规评分
func (r *PriceAuditRepository) calculateComplianceScore(eventsByRiskLevel map[types.ImpactRiskLevel]int, totalEvents int) float64 {
	if totalEvents == 0 {
		return 100.0
	}
	
	highRiskCount := eventsByRiskLevel[types.ImpactRiskLevelHigh]
	mediumRiskCount := eventsByRiskLevel[types.ImpactRiskLevelMedium]
	lowRiskCount := eventsByRiskLevel[types.ImpactRiskLevelLow]
	
	// 风险权重：高风险扣分更多
	riskScore := float64(highRiskCount)*3.0 + float64(mediumRiskCount)*1.5 + float64(lowRiskCount)*0.5
	maxScore := float64(totalEvents) * 3.0 // 如果全是高风险的最大扣分
	
	if maxScore == 0 {
		return 100.0
	}
	
	complianceScore := (1.0 - riskScore/maxScore) * 100.0
	if complianceScore < 0 {
		complianceScore = 0
	}
	
	return complianceScore
}

// DetectAnomalies 检测价格异常
func (r *PriceAuditRepository) DetectAnomalies(ctx context.Context, days int) ([]*types.PriceAnomalyDetection, error) {
	// 获取最近的价格变更事件
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	since := time.Now().AddDate(0, 0, -days)
	
	result, err := model.Where("event_type", types.PriceAuditEventTypePriceChange).
		Where("created_at >= ?", since).
		Order("created_at DESC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取价格变更事件失败: %w", err)
	}
	
	var anomalies []*types.PriceAnomalyDetection
	
	// 分析每个价格变更事件
	for _, record := range result {
		event := &types.PriceAuditEvent{}
		err = record.Struct(event)
		if err != nil {
			continue
		}
		
		// 检测异常的逻辑（简化版）
		anomaly := r.analyzePriceChangeAnomaly(event)
		if anomaly != nil {
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

// analyzePriceChangeAnomaly 分析价格变更异常
func (r *PriceAuditRepository) analyzePriceChangeAnomaly(event *types.PriceAuditEvent) *types.PriceAnomalyDetection {
	// 简化的异常检测逻辑
	if event.EventData.OldValue == nil || event.EventData.NewValue == nil {
		return nil
	}
	
	// 假设价格是Money类型
	oldPrice, ok1 := event.EventData.OldValue.(map[string]interface{})
	newPrice, ok2 := event.EventData.NewValue.(map[string]interface{})
	if !ok1 || !ok2 {
		return nil
	}
	
	oldAmount, ok1 := oldPrice["amount"].(float64)
	newAmount, ok2 := newPrice["amount"].(float64)
	if !ok1 || !ok2 || oldAmount == 0 {
		return nil
	}
	
	// 计算变化百分比
	changePercent := (newAmount - oldAmount) / oldAmount * 100
	
	var anomalyType types.PriceAnomalyType
	var severity types.ImpactRiskLevel
	var description string
	
	// 异常判断逻辑
	if changePercent > 50 {
		anomalyType = types.PriceAnomalyTypeUnusualIncrease
		severity = types.ImpactRiskLevelHigh
		description = fmt.Sprintf("价格异常上涨%.2f%%", changePercent)
	} else if changePercent < -50 {
		anomalyType = types.PriceAnomalyTypeUnusualDecrease
		severity = types.ImpactRiskLevelHigh
		description = fmt.Sprintf("价格异常下跌%.2f%%", changePercent)
	} else if newAmount <= 0 {
		if newAmount == 0 {
			anomalyType = types.PriceAnomalyTypeZeroPrice
		} else {
			anomalyType = types.PriceAnomalyTypeNegativePrice
		}
		severity = types.ImpactRiskLevelHigh
		description = "检测到异常价格值"
	} else {
		return nil // 未检测到异常
	}
	
	return &types.PriceAnomalyDetection{
		AnomalyID:     fmt.Sprintf("anomaly_%d_%d", event.ProductID, event.CreatedAt.Unix()),
		DetectedAt:    event.CreatedAt,
		AnomalyType:   anomalyType,
		ProductID:     event.ProductID,
		AnomalyScore:  float64(changePercent),
		Description:   description,
		CurrentValue:  newPrice,
		ExpectedValue: oldPrice,
		Deviation:     changePercent,
		Severity:      severity,
		AutoResolved:  false,
	}
}

// GetHighRiskEvents 获取高风险事件
func (r *PriceAuditRepository) GetHighRiskEvents(ctx context.Context, hours int) ([]*types.PriceAuditEvent, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	since := time.Now().Add(time.Duration(-hours) * time.Hour)
	
	result, err := model.Where("impact_level", types.ImpactRiskLevelHigh).
		Where("created_at >= ?", since).
		Order("created_at DESC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取高风险事件失败: %w", err)
	}
	
	var events []*types.PriceAuditEvent
	for _, record := range result {
		event := &types.PriceAuditEvent{}
		err = record.Struct(event)
		if err != nil {
			return nil, fmt.Errorf("审计事件数据转换失败: %w", err)
		}
		events = append(events, event)
	}
	
	return events, nil
}

// DeleteOldAuditEvents 删除旧的审计事件
func (r *PriceAuditRepository) DeleteOldAuditEvents(ctx context.Context, beforeDate time.Time) (int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return 0, err
	}
	
	result, err := model.Where("created_at < ?", beforeDate).Delete()
	if err != nil {
		return 0, fmt.Errorf("删除旧审计事件失败: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// GetEventTrend 获取事件趋势
func (r *PriceAuditRepository) GetEventTrend(ctx context.Context, days int) ([]types.EventDistribution, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	since := time.Now().AddDate(0, 0, -days)
	
	result, err := model.Fields("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Order("date ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取事件趋势失败: %w", err)
	}
	
	var trend []types.EventDistribution
	for _, record := range result {
		dateStr := record["date"].String()
		count := record["count"].Int()
		
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		
		trend = append(trend, types.EventDistribution{
			Date:  date,
			Count: count,
		})
	}
	
	return trend, nil
}

// GetUserActivityPattern 获取用户活动模式
func (r *PriceAuditRepository) GetUserActivityPattern(ctx context.Context, days int) ([]types.UserActivity, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	since := time.Now().AddDate(0, 0, -days)
	
	result, err := model.Fields("user_id, COUNT(*) as event_count, MAX(created_at) as last_activity").
		Where("created_at >= ?", since).
		Group("user_id").
		Order("event_count DESC").
		Limit(20). // 只取前20个最活跃用户
		All()
	if err != nil {
		return nil, fmt.Errorf("获取用户活动模式失败: %w", err)
	}
	
	var activities []types.UserActivity
	for _, record := range result {
		userID := record["user_id"].Uint64()
		eventCount := record["event_count"].Int()
		lastActivity := record["last_activity"].Time()
		
		// 简化的风险评分计算
		riskScore := r.calculateUserRiskScore(eventCount, days)
		
		activities = append(activities, types.UserActivity{
			UserID:       userID,
			EventCount:   eventCount,
			RiskScore:    riskScore,
			LastActivity: lastActivity,
		})
	}
	
	return activities, nil
}

// calculateUserRiskScore 计算用户风险评分
func (r *PriceAuditRepository) calculateUserRiskScore(eventCount, days int) float64 {
	averageDaily := float64(eventCount) / float64(days)
	
	// 简化的风险评分：基于日均活动频率
	if averageDaily > 10 {
		return 80.0 // 高风险
	} else if averageDaily > 5 {
		return 50.0 // 中风险
	} else if averageDaily > 2 {
		return 30.0 // 低风险
	}
	
	return 10.0 // 很低风险
}
package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// PriceImpactService 价格影响评估服务
type PriceImpactService struct {
	priceHistoryRepo *repository.PriceHistoryRepository
	priceAuditRepo   *repository.PriceAuditRepository
	productRepo      *repository.ProductRepository
	// TODO: 需要添加订单、库存、客户相关的Repository
}

// NewPriceImpactService 创建价格影响评估服务实例
func NewPriceImpactService() *PriceImpactService {
	return &PriceImpactService{
		priceHistoryRepo: repository.NewPriceHistoryRepository(),
		priceAuditRepo:   repository.NewPriceAuditRepository(),
		productRepo:      repository.NewProductRepository(),
	}
}

// AssessPriceImpact 评估价格变更影响
func (s *PriceImpactService) AssessPriceImpact(ctx context.Context, req *types.PriceImpactAssessmentRequest) (*types.PriceImpactAssessment, error) {
	// 获取商品信息
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("获取商品信息失败: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("商品不存在")
	}
	
	currentPrice := product.GetPrice()
	
	// 计算价格变化百分比
	changePercent := s.calculatePriceChangePercent(currentPrice, req.NewPrice)
	
	assessment := &types.PriceImpactAssessment{
		ProductID:           req.ProductID,
		OldPrice:           currentPrice,
		NewPrice:           req.NewPrice,
		ChangePercentage:   changePercent,
		EffectiveDate:      req.EffectiveDate,
		AssessmentTimestamp: time.Now(),
	}
	
	// 评估对订单的影响
	assessment.ImpactedOrders, err = s.assessOrderImpact(ctx, req.ProductID, currentPrice, req.NewPrice)
	if err != nil {
		return nil, fmt.Errorf("评估订单影响失败: %w", err)
	}
	
	// 评估收入影响
	if req.IncludePredict {
		assessment.RevenueImpact, err = s.assessRevenueImpact(ctx, req.ProductID, currentPrice, req.NewPrice, req.AssessmentPeriod)
		if err != nil {
			return nil, fmt.Errorf("评估收入影响失败: %w", err)
		}
	}
	
	// 评估库存影响
	assessment.InventoryImpact, err = s.assessInventoryImpact(ctx, req.ProductID, changePercent)
	if err != nil {
		return nil, fmt.Errorf("评估库存影响失败: %w", err)
	}
	
	// 评估客户影响
	if req.IncludePredict {
		assessment.CustomerImpact, err = s.assessCustomerImpact(ctx, req.ProductID, changePercent)
		if err != nil {
			return nil, fmt.Errorf("评估客户影响失败: %w", err)
		}
	}
	
	// 确定风险等级
	assessment.RiskLevel = s.determineRiskLevel(changePercent, assessment.ImpactedOrders, assessment.RevenueImpact)
	
	// 生成建议行动
	assessment.RecommendedActions = s.generateRecommendedActions(assessment)
	
	return assessment, nil
}

// calculatePriceChangePercent 计算价格变化百分比
func (s *PriceImpactService) calculatePriceChangePercent(oldPrice, newPrice types.Money) float64 {
	if oldPrice.Amount == 0 {
		return 0
	}
	return (newPrice.Amount - oldPrice.Amount) / oldPrice.Amount * 100
}

// assessOrderImpact 评估对订单的影响
func (s *PriceImpactService) assessOrderImpact(ctx context.Context, productID uint64, oldPrice, newPrice types.Money) (types.PriceImpactedOrders, error) {
	impact := types.PriceImpactedOrders{}
	
	// TODO: 这里需要实际的订单数据查询
	// 目前使用模拟数据
	
	// 模拟待处理订单影响
	impact.PendingOrdersCount = 15
	impact.PendingOrdersValue = types.Money{
		Amount:   1500.0,
		Currency: oldPrice.Currency,
	}
	
	// 模拟活跃购物车影响
	impact.ActiveCartCount = 8
	impact.ActiveCartValue = types.Money{
		Amount:   800.0,
		Currency: oldPrice.Currency,
	}
	
	// 模拟最近订单影响（7天）
	impact.RecentOrdersCount = 45
	impact.RecentOrdersValue = types.Money{
		Amount:   4500.0,
		Currency: oldPrice.Currency,
	}
	
	// 估算受影响客户数量
	impact.AffectedCustomers = impact.PendingOrdersCount + impact.ActiveCartCount
	
	// 判断是否需要通知客户
	changePercent := s.calculatePriceChangePercent(oldPrice, newPrice)
	impact.RequiresNotification = math.Abs(changePercent) > 10 // 价格变化超过10%需要通知
	
	return impact, nil
}

// assessRevenueImpact 评估收入影响
func (s *PriceImpactService) assessRevenueImpact(ctx context.Context, productID uint64, oldPrice, newPrice types.Money, assessmentDays int) (types.RevenueImpactAnalysis, error) {
	impact := types.RevenueImpactAnalysis{}
	
	if assessmentDays == 0 {
		assessmentDays = 30
	}
	
	// 获取历史销售数据（模拟）
	historicalVolume := 100 // 模拟30天销量
	impact.HistoricalSalesVolume = historicalVolume
	
	// 计算价格变化率
	changePercent := s.calculatePriceChangePercent(oldPrice, newPrice)
	
	// 简化的需求弹性计算（实际应该基于历史数据分析）
	priceElasticity := -1.2 // 假设价格弹性系数
	volumeChangePercent := priceElasticity * changePercent / 100
	impact.PredictedVolumeChange = volumeChangePercent
	
	// 预计销量
	predictedVolume := float64(historicalVolume) * (1 + volumeChangePercent)
	
	// 计算收入影响
	oldRevenue := oldPrice.Amount * float64(historicalVolume)
	newRevenue := newPrice.Amount * predictedVolume
	
	dailyRevenueChange := (newRevenue - oldRevenue) / float64(assessmentDays)
	weeklyRevenueChange := dailyRevenueChange * 7
	monthlyRevenueChange := dailyRevenueChange * 30
	
	impact.ProjectedDailyRevenue = types.Money{
		Amount:   dailyRevenueChange,
		Currency: oldPrice.Currency,
	}
	impact.ProjectedWeeklyRevenue = types.Money{
		Amount:   weeklyRevenueChange,
		Currency: oldPrice.Currency,
	}
	impact.ProjectedMonthlyRevenue = types.Money{
		Amount:   monthlyRevenueChange,
		Currency: oldPrice.Currency,
	}
	
	// 计算盈亏平衡点
	if dailyRevenueChange < 0 {
		impact.BreakEvenPoint = int(math.Abs(newRevenue-oldRevenue) / math.Abs(dailyRevenueChange))
	} else {
		impact.BreakEvenPoint = 0
	}
	
	// 计算收入风险评分
	impact.RevenueRiskScore = s.calculateRevenueRiskScore(changePercent, volumeChangePercent)
	
	return impact, nil
}

// assessInventoryImpact 评估库存影响
func (s *PriceImpactService) assessInventoryImpact(ctx context.Context, productID uint64, changePercent float64) (types.InventoryImpactAnalysis, error) {
	impact := types.InventoryImpactAnalysis{}
	
	// TODO: 获取实际库存数据
	// 目前使用模拟数据
	impact.CurrentStock = 200
	impact.AverageMonthlyTurnover = 150
	
	// 预测周转变化
	elasticity := -0.8 // 库存周转对价格的弹性
	impact.PredictedTurnoverChange = elasticity * changePercent / 100
	
	// 评估库存风险
	if changePercent > 0 { // 涨价
		impact.OverstockRisk = math.Min(100, math.Abs(changePercent)*2) // 涨价导致滞销风险
		impact.UnderstockRisk = 10
	} else { // 降价
		impact.UnderstockRisk = math.Min(100, math.Abs(changePercent)*1.5) // 降价导致缺货风险
		impact.OverstockRisk = 10
	}
	
	// 建议库存水平
	adjustmentFactor := 1 + impact.PredictedTurnoverChange
	impact.RecommendedStockLevel = int(float64(impact.AverageMonthlyTurnover) * adjustmentFactor * 1.5) // 1.5倍安全库存
	
	// 库存价值变化（按新价格计算）
	// TODO: 需要实际的成本价格数据
	impact.InventoryValue = types.Money{
		Amount:   float64(impact.CurrentStock) * changePercent * 0.01, // 简化计算
		Currency: "CNY",
	}
	
	return impact, nil
}

// assessCustomerImpact 评估客户影响
func (s *PriceImpactService) assessCustomerImpact(ctx context.Context, productID uint64, changePercent float64) (types.CustomerImpactAnalysis, error) {
	impact := types.CustomerImpactAnalysis{}
	
	// 价格弹性（简化模型）
	impact.PriceElasticity = -1.2
	
	// 模拟客户群体影响
	impact.CustomerSegmentImpact = []types.CustomerSegmentImpact{
		{
			SegmentName:      "价格敏感型客户",
			CustomerCount:    50,
			ImpactScore:      math.Abs(changePercent) * 0.8,
			ExpectedBehavior: "可能流失或增加购买",
			MitigationNeeded: math.Abs(changePercent) > 15,
		},
		{
			SegmentName:      "品牌忠诚客户",
			CustomerCount:    30,
			ImpactScore:      math.Abs(changePercent) * 0.3,
			ExpectedBehavior: "影响较小",
			MitigationNeeded: math.Abs(changePercent) > 25,
		},
	}
	
	// 竞争地位分析
	impact.CompetitivePosition = types.CompetitiveAnalysis{
		MarketPosition:       "中等",
		CompetitorPriceGap:   changePercent, // 简化：假设变价后与竞品的差距
		CompetitiveAdvantage: changePercent < 0, // 降价具有优势
		MarketShareRisk:      math.Min(100, math.Abs(changePercent)*1.5),
	}
	
	// 流失风险评估
	if changePercent > 0 {
		impact.ChurnRisk = math.Min(100, changePercent*2)
	} else {
		impact.ChurnRisk = math.Max(0, impact.ChurnRisk-math.Abs(changePercent))
	}
	
	// 获客影响
	if changePercent < 0 {
		impact.AcquisitionImpact = math.Abs(changePercent) * 1.5 // 降价有利于获客
	} else {
		impact.AcquisitionImpact = -changePercent * 0.8 // 涨价不利于获客
	}
	
	// 满意度影响
	impact.SatisfactionImpact = -changePercent * 0.5 // 涨价降低满意度
	
	return impact, nil
}

// determineRiskLevel 确定风险等级
func (s *PriceImpactService) determineRiskLevel(changePercent float64, orderImpact types.PriceImpactedOrders, revenueImpact types.RevenueImpactAnalysis) types.ImpactRiskLevel {
	absChange := math.Abs(changePercent)
	
	// 基于价格变化幅度的初始风险等级
	var riskScore float64
	if absChange > 30 {
		riskScore = 80
	} else if absChange > 15 {
		riskScore = 60
	} else if absChange > 5 {
		riskScore = 40
	} else {
		riskScore = 20
	}
	
	// 基于受影响订单数量调整风险
	if orderImpact.PendingOrdersCount > 20 {
		riskScore += 15
	} else if orderImpact.PendingOrdersCount > 10 {
		riskScore += 10
	}
	
	// 基于收入影响调整风险
	if revenueImpact.RevenueRiskScore > 70 {
		riskScore += 20
	} else if revenueImpact.RevenueRiskScore > 50 {
		riskScore += 10
	}
	
	// 确定最终风险等级
	if riskScore > 70 {
		return types.ImpactRiskLevelHigh
	} else if riskScore > 40 {
		return types.ImpactRiskLevelMedium
	}
	return types.ImpactRiskLevelLow
}

// calculateRevenueRiskScore 计算收入风险评分
func (s *PriceImpactService) calculateRevenueRiskScore(priceChangePercent, volumeChangePercent float64) float64 {
	// 基于价格和销量变化的组合风险评分
	priceRisk := math.Abs(priceChangePercent) / 2
	volumeRisk := math.Abs(volumeChangePercent) * 100
	
	combinedRisk := (priceRisk + volumeRisk) / 2
	return math.Min(100, combinedRisk)
}

// generateRecommendedActions 生成建议行动
func (s *PriceImpactService) generateRecommendedActions(assessment *types.PriceImpactAssessment) []string {
	var actions []string
	
	switch assessment.RiskLevel {
	case types.ImpactRiskLevelHigh:
		actions = append(actions, "建议分阶段实施价格变更")
		actions = append(actions, "提前通知受影响客户")
		actions = append(actions, "准备客户保留策略")
		actions = append(actions, "密切监控竞争对手反应")
	case types.ImpactRiskLevelMedium:
		actions = append(actions, "建议在非高峰期实施变更")
		actions = append(actions, "通知重要客户")
		actions = append(actions, "调整库存计划")
	case types.ImpactRiskLevelLow:
		actions = append(actions, "可以正常实施价格变更")
		actions = append(actions, "标准通知流程")
	}
	
	// 基于具体影响添加行动建议
	if assessment.ImpactedOrders.RequiresNotification {
		actions = append(actions, "发送价格变更通知邮件")
	}
	
	if assessment.InventoryImpact.OverstockRisk > 50 {
		actions = append(actions, "考虑促销活动清理库存")
	}
	
	if assessment.InventoryImpact.UnderstockRisk > 50 {
		actions = append(actions, "增加库存采购计划")
	}
	
	return actions
}

// GeneratePriceChangeRecommendation 生成价格变更建议
func (s *PriceImpactService) GeneratePriceChangeRecommendation(ctx context.Context, assessment *types.PriceImpactAssessment) (*types.PriceChangeRecommendation, error) {
	recommendation := &types.PriceChangeRecommendation{}
	
	// 基于评估结果确定建议类型
	switch assessment.RiskLevel {
	case types.ImpactRiskLevelHigh:
		if math.Abs(assessment.ChangePercentage) > 25 {
			recommendation.RecommendationType = types.RecommendationTypeReject
			recommendation.Reasoning = "价格变化幅度过大，风险极高"
		} else {
			recommendation.RecommendationType = types.RecommendationTypeGradual
			recommendation.Reasoning = "建议分阶段实施以降低风险"
		}
	case types.ImpactRiskLevelMedium:
		recommendation.RecommendationType = types.RecommendationTypeModify
		recommendation.Reasoning = "建议调整实施策略以降低风险"
	case types.ImpactRiskLevelLow:
		recommendation.RecommendationType = types.RecommendationTypeApprove
		recommendation.Reasoning = "风险可控，可以实施"
	}
	
	// 设置建议价格（如果需要修改）
	if recommendation.RecommendationType == types.RecommendationTypeModify {
		// 建议一个更温和的价格
		moderateChangePercent := assessment.ChangePercentage * 0.7 // 减少30%的变化
		recommendedAmount := assessment.OldPrice.Amount * (1 + moderateChangePercent/100)
		recommendation.RecommendedPrice = types.Money{
			Amount:   recommendedAmount,
			Currency: assessment.OldPrice.Currency,
		}
	}
	
	// 设置建议时机
	if recommendation.RecommendationType == types.RecommendationTypeDelay {
		recommendation.RecommendedTiming = time.Now().Add(7 * 24 * time.Hour) // 延迟1周
	} else {
		recommendation.RecommendedTiming = assessment.EffectiveDate
	}
	
	// 预期结果
	recommendation.ExpectedOutcome = s.generateExpectedOutcome(assessment, recommendation.RecommendationType)
	
	// 实施步骤
	recommendation.ImplementationSteps = s.generateImplementationSteps(recommendation.RecommendationType)
	
	// 监控指标
	recommendation.MonitoringMetrics = []string{
		"销售量变化",
		"客户投诉率",
		"竞争对手反应",
		"库存周转率",
		"客户满意度",
	}
	
	// 回滚计划
	recommendation.RollbackPlan = "如果销量下降超过15%或客户投诉激增，立即回滚到原价格"
	
	// 置信度评估
	recommendation.ConfidenceLevel = s.calculateConfidenceLevel(assessment)
	
	return recommendation, nil
}

// generateExpectedOutcome 生成预期结果
func (s *PriceImpactService) generateExpectedOutcome(assessment *types.PriceImpactAssessment, recType types.RecommendationType) string {
	switch recType {
	case types.RecommendationTypeApprove:
		return fmt.Sprintf("预计销量变化%.1f%%，收入影响%.0f元", 
			assessment.RevenueImpact.PredictedVolumeChange*100, 
			assessment.RevenueImpact.ProjectedMonthlyRevenue.Amount)
	case types.RecommendationTypeModify:
		return "通过调整价格幅度，预计可降低50%的风险"
	case types.RecommendationTypeGradual:
		return "分阶段实施可将客户流失风险降低60%"
	case types.RecommendationTypeReject:
		return "避免高风险的负面影响"
	case types.RecommendationTypeDelay:
		return "延迟实施可获得更多市场观察时间"
	default:
		return "待评估"
	}
}

// generateImplementationSteps 生成实施步骤
func (s *PriceImpactService) generateImplementationSteps(recType types.RecommendationType) []string {
	switch recType {
	case types.RecommendationTypeApprove:
		return []string{
			"1. 更新商品价格",
			"2. 发送客户通知",
			"3. 更新营销材料",
			"4. 监控市场反应",
		}
	case types.RecommendationTypeGradual:
		return []string{
			"1. 第一阶段：调整50%的价格变化",
			"2. 观察2周市场反应",
			"3. 第二阶段：完成剩余价格调整",
			"4. 持续监控和调优",
		}
	case types.RecommendationTypeModify:
		return []string{
			"1. 调整价格幅度至建议值",
			"2. 重新评估影响",
			"3. 实施调整后的价格",
			"4. 加强客户沟通",
		}
	default:
		return []string{"需要进一步分析"}
	}
}

// calculateConfidenceLevel 计算置信度
func (s *PriceImpactService) calculateConfidenceLevel(assessment *types.PriceImpactAssessment) float64 {
	// 简化的置信度计算
	baseConfidence := 70.0
	
	// 根据历史数据质量调整
	if assessment.RevenueImpact.HistoricalSalesVolume > 0 {
		baseConfidence += 10
	}
	
	// 根据风险等级调整
	switch assessment.RiskLevel {
	case types.ImpactRiskLevelLow:
		baseConfidence += 15
	case types.ImpactRiskLevelMedium:
		baseConfidence += 5
	case types.ImpactRiskLevelHigh:
		baseConfidence -= 10
	}
	
	return math.Min(100, math.Max(0, baseConfidence))
}

// CreateAuditEvent 创建审计事件
func (s *PriceImpactService) CreateAuditEvent(ctx context.Context, req *types.CreatePriceAuditEventRequest) (*types.PriceAuditEvent, error) {
	return s.priceAuditRepo.CreateAuditEvent(ctx, req)
}

// GetAuditStatistics 获取审计统计
func (s *PriceImpactService) GetAuditStatistics(ctx context.Context, days int) (*types.PriceAuditSummary, error) {
	return s.priceAuditRepo.GetAuditStatistics(ctx, days)
}

// DetectPriceAnomalies 检测价格异常
func (s *PriceImpactService) DetectPriceAnomalies(ctx context.Context, days int) ([]*types.PriceAnomalyDetection, error) {
	return s.priceAuditRepo.DetectAnomalies(ctx, days)
}
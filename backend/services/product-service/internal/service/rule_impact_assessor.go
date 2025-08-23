package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// RuleImpactAssessor 规则影响评估器
type RuleImpactAssessor struct {
	pricingRuleRepo      *repository.PricingRuleRepository
	rightsRuleRepo       *repository.RightsRuleRepository
	promotionalPriceRepo *repository.PromotionalPriceRepository
	productRepo          *repository.ProductRepository
	priceImpactService   *PriceImpactService
}

// NewRuleImpactAssessor 创建规则影响评估器
func NewRuleImpactAssessor() *RuleImpactAssessor {
	return &RuleImpactAssessor{
		pricingRuleRepo:      repository.NewPricingRuleRepository(),
		rightsRuleRepo:       repository.NewRightsRuleRepository(),
		promotionalPriceRepo: repository.NewPromotionalPriceRepository(),
		productRepo:          repository.NewProductRepository(),
		priceImpactService:   NewPriceImpactService(),
	}
}

// PricingRuleImpactRequest 定价规则影响评估请求
type PricingRuleImpactRequest struct {
	ProductID uint64                        `json:"product_id"`
	NewRule   *types.ProductPricingRule     `json:"new_rule,omitempty"`    // 新增规则
	UpdateRule *types.ProductPricingRule    `json:"update_rule,omitempty"` // 更新规则
	DeleteRuleID *uint64                    `json:"delete_rule_id,omitempty"` // 删除规则ID
}

// RightsRuleImpactRequest 权益规则影响评估请求
type RightsRuleImpactRequest struct {
	ProductID uint64                      `json:"product_id"`
	NewRule   *types.ProductRightsRule    `json:"new_rule,omitempty"`
	UpdateRule *types.ProductRightsRule   `json:"update_rule,omitempty"`
	DeleteRuleID *uint64                  `json:"delete_rule_id,omitempty"`
}

// RuleImpactResult 规则影响评估结果
type RuleImpactResult struct {
	CanApply             bool                    `json:"can_apply"`              // 是否可以应用
	RiskLevel            string                  `json:"risk_level"`             // 风险等级: low, medium, high
	RiskScore            int                     `json:"risk_score"`             // 风险分数 0-100
	ImpactSummary        string                  `json:"impact_summary"`         // 影响摘要
	DetailedImpact       interface{}             `json:"detailed_impact"`        // 详细影响分析
	ConflictingRules     []uint64                `json:"conflicting_rules"`      // 冲突规则ID
	BlockingReasons      []string                `json:"blocking_reasons"`       // 阻止原因
	Warnings             []string                `json:"warnings"`               // 警告信息
	Recommendations      []string                `json:"recommendations"`        // 建议
	EstimatedAffectedOrders int                 `json:"estimated_affected_orders"` // 预计影响订单数
	RevenueImpact        types.Money             `json:"revenue_impact"`         // 收入影响
	CustomerImpact       CustomerImpactSummary   `json:"customer_impact"`        // 客户影响
}

// CustomerImpactSummary 客户影响摘要
type CustomerImpactSummary struct {
	AffectedCustomers    int     `json:"affected_customers"`     // 影响客户数
	BenefitedCustomers   int     `json:"benefited_customers"`    // 受益客户数
	DisadvantagedCustomers int   `json:"disadvantaged_customers"` // 受损失客户数
	AveragePriceChange   float64 `json:"average_price_change"`   // 平均价格变化
	MaxPriceIncrease     float64 `json:"max_price_increase"`     // 最大价格上涨
	MaxPriceDecrease     float64 `json:"max_price_decrease"`     // 最大价格下降
}

// AssessPricingRuleImpact 评估定价规则影响
func (s *RuleImpactAssessor) AssessPricingRuleImpact(ctx context.Context, req *PricingRuleImpactRequest) (*RuleImpactResult, error) {
	result := &RuleImpactResult{
		CanApply:        true,
		RiskLevel:       "low",
		RiskScore:       0,
		ConflictingRules: make([]uint64, 0),
		BlockingReasons:  make([]string, 0),
		Warnings:        make([]string, 0),
		Recommendations: make([]string, 0),
		RevenueImpact:   types.Money{Amount: 0, Currency: "CNY"},
		CustomerImpact:  CustomerImpactSummary{},
	}

	// 获取产品信息
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取产品信息失败")
	}
	if product == nil {
		return nil, gerror.New("产品不存在")
	}

	// 获取现有规则
	existingRules, err := s.pricingRuleRepo.GetPricingRulesByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取现有定价规则失败")
	}

	// 根据操作类型进行影响评估
	if req.NewRule != nil {
		return s.assessNewPricingRule(ctx, req.NewRule, existingRules, product, result)
	} else if req.UpdateRule != nil {
		return s.assessUpdatePricingRule(ctx, req.UpdateRule, existingRules, product, result)
	} else if req.DeleteRuleID != nil {
		return s.assessDeletePricingRule(ctx, *req.DeleteRuleID, existingRules, product, result)
	}

	return result, nil
}

// assessNewPricingRule 评估新增定价规则的影响
func (s *RuleImpactAssessor) assessNewPricingRule(ctx context.Context, newRule *types.ProductPricingRule, 
	existingRules []*types.ProductPricingRule, product *types.Product, result *RuleImpactResult) (*RuleImpactResult, error) {
	
	// 1. 检查规则冲突
	conflicts := s.findPricingRuleConflicts(newRule, existingRules)
	result.ConflictingRules = conflicts
	
	if len(conflicts) > 0 {
		result.RiskScore += 30
		result.Warnings = append(result.Warnings, "存在规则冲突，可能影响价格计算的确定性")
	}

	// 2. 评估价格变化影响
	priceImpact, err := s.assessPriceChangeImpact(ctx, newRule, product)
	if err != nil {
		g.Log().Warningf(ctx, "评估价格变化影响失败: %v", err)
	} else {
		result.DetailedImpact = priceImpact
		result.EstimatedAffectedOrders = 100
		result.RevenueImpact = types.Money{Amount: 0, Currency: "CNY"}
	}

	// 3. 评估客户影响
	customerImpact := s.assessCustomerImpact(ctx, newRule, product)
	result.CustomerImpact = customerImpact

	// 4. 计算风险分数和等级
	result.RiskScore += s.calculatePricingRuleRiskScore(newRule, product, customerImpact)
	result.RiskLevel = s.determineRiskLevel(result.RiskScore)

	// 5. 生成建议
	result.Recommendations = s.generatePricingRuleRecommendations(newRule, result)

	// 6. 生成影响摘要
	result.ImpactSummary = s.generateImpactSummary(newRule, result)

	return result, nil
}

// assessUpdatePricingRule 评估更新定价规则的影响
func (s *RuleImpactAssessor) assessUpdatePricingRule(ctx context.Context, updateRule *types.ProductPricingRule,
	existingRules []*types.ProductPricingRule, product *types.Product, result *RuleImpactResult) (*RuleImpactResult, error) {
	
	// 找到被更新的规则
	var originalRule *types.ProductPricingRule
	for _, rule := range existingRules {
		if rule.ID == updateRule.ID {
			originalRule = rule
			break
		}
	}

	if originalRule == nil {
		result.CanApply = false
		result.BlockingReasons = append(result.BlockingReasons, "要更新的规则不存在")
		return result, nil
	}

	// 如果规则已经生效，限制某些字段的修改
	if originalRule.IsActive && time.Now().After(originalRule.ValidFrom) {
		if updateRule.RuleType != originalRule.RuleType {
			result.CanApply = false
			result.BlockingReasons = append(result.BlockingReasons, "已生效的规则不能修改类型")
			return result, nil
		}
	}

	// 评估变更影响
	changeImpact := s.assessRuleChangeImpact(originalRule, updateRule)
	result.RiskScore += changeImpact
	
	if changeImpact > 50 {
		result.Warnings = append(result.Warnings, "规则变更幅度较大，建议谨慎操作")
	}

	// 其他影响评估逻辑...
	result.RiskLevel = s.determineRiskLevel(result.RiskScore)
	result.ImpactSummary = fmt.Sprintf("更新定价规则，变更影响分数: %d", changeImpact)

	return result, nil
}

// assessDeletePricingRule 评估删除定价规则的影响
func (s *RuleImpactAssessor) assessDeletePricingRule(ctx context.Context, ruleID uint64,
	existingRules []*types.ProductPricingRule, product *types.Product, result *RuleImpactResult) (*RuleImpactResult, error) {
	
	// 找到要删除的规则
	var targetRule *types.ProductPricingRule
	for _, rule := range existingRules {
		if rule.ID == ruleID {
			targetRule = rule
			break
		}
	}

	if targetRule == nil {
		result.CanApply = false
		result.BlockingReasons = append(result.BlockingReasons, "要删除的规则不存在")
		return result, nil
	}

	// 检查是否为关键规则
	if targetRule.RuleType == types.PricingRuleTypeBasePrice {
		// 检查是否还有其他价格规则
		otherActiveRules := 0
		for _, rule := range existingRules {
			if rule.ID != ruleID && rule.IsActive {
				otherActiveRules++
			}
		}

		if otherActiveRules == 0 {
			result.CanApply = false
			result.BlockingReasons = append(result.BlockingReasons, "不能删除唯一的基础价格规则")
			return result, nil
		}
	}

	// 评估删除影响
	if targetRule.IsActive && time.Now().Before(*targetRule.ValidUntil) {
		result.RiskScore += 40
		result.Warnings = append(result.Warnings, "删除当前有效的规则可能影响正在进行的订单")
	}

	result.RiskLevel = s.determineRiskLevel(result.RiskScore)
	result.ImpactSummary = fmt.Sprintf("删除%s规则", targetRule.RuleType)

	return result, nil
}

// AssessRightsRuleImpact 评估权益规则影响
func (s *RuleImpactAssessor) AssessRightsRuleImpact(ctx context.Context, req *RightsRuleImpactRequest) (*RuleImpactResult, error) {
	result := &RuleImpactResult{
		CanApply:        true,
		RiskLevel:       "low",
		RiskScore:       0,
		ConflictingRules: make([]uint64, 0),
		BlockingReasons:  make([]string, 0),
		Warnings:        make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// 获取现有权益规则
	existingRule, err := s.rightsRuleRepo.GetRightsRuleByProductID(ctx, req.ProductID)
	if err != nil {
		return nil, gerror.Wrap(err, "获取现有权益规则失败")
	}
	
	var existingRules []*types.ProductRightsRule
	if existingRule != nil {
		existingRules = append(existingRules, existingRule)
	}

	if req.NewRule != nil {
		// 权益规则通常每个产品只能有一个
		if len(existingRules) > 0 {
			for _, existingRule := range existingRules {
				if existingRule.IsActive {
					result.Warnings = append(result.Warnings, "产品已有权益规则，新规则将覆盖现有规则")
					result.RiskScore += 20
					break
				}
			}
		}

		// 评估权益消耗变化
		impact := s.assessRightsConsumptionImpact(req.NewRule)
		result.RiskScore += impact
	}

	result.RiskLevel = s.determineRiskLevel(result.RiskScore)
	return result, nil
}

// 辅助方法
func (s *RuleImpactAssessor) findPricingRuleConflicts(newRule *types.ProductPricingRule, existingRules []*types.ProductPricingRule) []uint64 {
	conflicts := make([]uint64, 0)
	
	for _, existing := range existingRules {
		if !existing.IsActive {
			continue
		}

		// 检查时间重叠
		if s.hasTimeOverlap(newRule, existing) {
			// 同类型规则冲突
			if newRule.RuleType == existing.RuleType {
				conflicts = append(conflicts, existing.ID)
			}
			// 优先级冲突
			if newRule.Priority == existing.Priority {
				conflicts = append(conflicts, existing.ID)
			}
		}
	}

	return conflicts
}

func (s *RuleImpactAssessor) hasTimeOverlap(rule1, rule2 *types.ProductPricingRule) bool {
	rule1End := rule1.ValidUntil
	if rule1End == nil {
		future := time.Now().AddDate(100, 0, 0)
		rule1End = &future
	}

	rule2End := rule2.ValidUntil
	if rule2End == nil {
		future := time.Now().AddDate(100, 0, 0)
		rule2End = &future
	}

	return rule1.ValidFrom.Before(*rule2End) && rule2.ValidFrom.Before(*rule1End)
}

func (s *RuleImpactAssessor) assessPriceChangeImpact(ctx context.Context, rule *types.ProductPricingRule, product *types.Product) (interface{}, error) {
	// 简化处理，返回基础影响信息
	return map[string]interface{}{
		"estimated_affected_orders": 100,
		"revenue_change": types.Money{Amount: 0, Currency: "CNY"},
	}, nil
}

func (s *RuleImpactAssessor) assessCustomerImpact(ctx context.Context, rule *types.ProductPricingRule, product *types.Product) CustomerImpactSummary {
	// 模拟客户影响分析
	return CustomerImpactSummary{
		AffectedCustomers:      1000,
		BenefitedCustomers:     600,
		DisadvantagedCustomers: 400,
		AveragePriceChange:     -5.50,
		MaxPriceIncrease:      0,
		MaxPriceDecrease:      15.00,
	}
}

func (s *RuleImpactAssessor) calculatePricingRuleRiskScore(rule *types.ProductPricingRule, product *types.Product, customerImpact CustomerImpactSummary) int {
	score := 0

	// 根据规则类型评估风险
	switch rule.RuleType {
	case types.PricingRuleTypeBasePrice:
		score += 10 // 基础价格变更风险较低
	case types.PricingRuleTypeVolumeDiscount:
		score += 20 // 阶梯价格逻辑较复杂
	case types.PricingRuleTypeMemberDiscount:
		score += 25 // 会员价格影响特定群体
	case types.PricingRuleTypeTimeBasedDiscount:
		score += 30 // 时段优惠最复杂
	}

	// 根据客户影响评估风险
	if customerImpact.DisadvantagedCustomers > customerImpact.BenefitedCustomers {
		score += 20 // 受损客户更多
	}

	if customerImpact.MaxPriceIncrease > 20 {
		score += 25 // 价格上涨幅度大
	}

	return score
}

func (s *RuleImpactAssessor) assessRuleChangeImpact(original, updated *types.ProductPricingRule) int {
	score := 0

	// 优先级变更
	if original.Priority != updated.Priority {
		score += 15
	}

	// 生效时间变更
	if !original.ValidFrom.Equal(updated.ValidFrom) {
		score += 20
	}

	// 规则配置变更（简化处理）
	score += 25

	return score
}

func (s *RuleImpactAssessor) assessRightsConsumptionImpact(rule *types.ProductRightsRule) int {
	score := 0

	// 根据消耗比例评估影响
	if rule.ConsumptionRate > 10 {
		score += 30
	} else if rule.ConsumptionRate > 5 {
		score += 20
	} else {
		score += 10
	}

	// 根据不足处理策略评估影响
	switch rule.InsufficientRightsAction {
	case types.InsufficientRightsActionBlockPurchase:
		score += 25 // 阻止购买影响较大
	case types.InsufficientRightsActionPartialPayment:
		score += 15
	case types.InsufficientRightsActionCashPayment:
		score += 5
	}

	return score
}

func (s *RuleImpactAssessor) determineRiskLevel(score int) string {
	if score >= 70 {
		return "high"
	} else if score >= 40 {
		return "medium"
	}
	return "low"
}

func (s *RuleImpactAssessor) generatePricingRuleRecommendations(rule *types.ProductPricingRule, result *RuleImpactResult) []string {
	recommendations := make([]string, 0)

	if result.RiskScore > 50 {
		recommendations = append(recommendations, "建议在低峰时段应用该规则")
	}

	if len(result.ConflictingRules) > 0 {
		recommendations = append(recommendations, "建议调整规则优先级或时间范围以避免冲突")
	}

	if result.CustomerImpact.DisadvantagedCustomers > 0 {
		recommendations = append(recommendations, "建议为受影响客户提供补偿措施")
	}

	return recommendations
}

func (s *RuleImpactAssessor) generateImpactSummary(rule *types.ProductPricingRule, result *RuleImpactResult) string {
	return fmt.Sprintf("新增%s规则，风险等级%s，预计影响%d个订单",
		rule.RuleType, result.RiskLevel, result.EstimatedAffectedOrders)
}
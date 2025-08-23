package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// PricingRuleRepository 定价规则仓储
type PricingRuleRepository struct {
	*BaseRepository
}

// NewPricingRuleRepository 创建定价规则仓储实例
func NewPricingRuleRepository() *PricingRuleRepository {
	baseRepo := NewBaseRepository()
	baseRepo.tableName = "product_pricing_rules"
	
	return &PricingRuleRepository{
		BaseRepository: baseRepo,
	}
}

// CreatePricingRule 创建定价规则
func (r *PricingRuleRepository) CreatePricingRule(ctx context.Context, productID uint64, req *types.CreatePricingRuleRequest) (*types.ProductPricingRule, error) {
	rule := &types.ProductPricingRule{
		ProductID:  productID,
		RuleType:   req.RuleType,
		RuleConfig: req.RuleConfig,
		Priority:   req.Priority,
		IsActive:   true,
		ValidFrom:  req.ValidFrom,
		ValidUntil: req.ValidUntil,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	id, err := r.InsertAndGetId(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("创建定价规则失败: %w", err)
	}
	
	rule.ID = uint64(id)
	rule.TenantID = r.GetTenantID(ctx)
	
	return rule, nil
}

// GetPricingRuleByID 根据ID获取定价规则
func (r *PricingRuleRepository) GetPricingRuleByID(ctx context.Context, id uint64) (*types.ProductPricingRule, error) {
	record, err := r.FindOne(ctx, "id", id)
	if err != nil {
		return nil, fmt.Errorf("获取定价规则失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	rule := &types.ProductPricingRule{}
	err = record.Struct(rule)
	if err != nil {
		return nil, fmt.Errorf("定价规则数据转换失败: %w", err)
	}
	
	return rule, nil
}

// GetPricingRulesByProductID 获取商品的所有定价规则
func (r *PricingRuleRepository) GetPricingRulesByProductID(ctx context.Context, productID uint64) ([]*types.ProductPricingRule, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Where("product_id", productID).
		Order("priority DESC, created_at ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取商品定价规则失败: %w", err)
	}
	
	var rules []*types.ProductPricingRule
	for _, record := range result {
		rule := &types.ProductPricingRule{}
		err = record.Struct(rule)
		if err != nil {
			return nil, fmt.Errorf("定价规则数据转换失败: %w", err)
		}
		rules = append(rules, rule)
	}
	
	return rules, nil
}

// GetActivePricingRules 获取商品的有效定价规则
func (r *PricingRuleRepository) GetActivePricingRules(ctx context.Context, productID uint64, currentTime time.Time) ([]*types.ProductPricingRule, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Where("product_id", productID).
		Where("is_active", true).
		Where("valid_from <= ?", currentTime).
		WhereOrNull("valid_until").WhereOr("valid_until > ?", currentTime).
		Order("priority DESC, created_at ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取有效定价规则失败: %w", err)
	}
	
	var rules []*types.ProductPricingRule
	for _, record := range result {
		rule := &types.ProductPricingRule{}
		err = record.Struct(rule)
		if err != nil {
			return nil, fmt.Errorf("定价规则数据转换失败: %w", err)
		}
		rules = append(rules, rule)
	}
	
	return rules, nil
}

// UpdatePricingRule 更新定价规则
func (r *PricingRuleRepository) UpdatePricingRule(ctx context.Context, id uint64, req *types.UpdatePricingRuleRequest) (*types.ProductPricingRule, error) {
	updateData := make(map[string]interface{})
	updateData["updated_at"] = time.Now()
	
	if req.RuleConfig != nil {
		updateData["rule_config"] = req.RuleConfig
	}
	if req.Priority != nil {
		updateData["priority"] = *req.Priority
	}
	if req.IsActive != nil {
		updateData["is_active"] = *req.IsActive
	}
	if req.ValidFrom != nil {
		updateData["valid_from"] = *req.ValidFrom
	}
	if req.ValidUntil != nil {
		updateData["valid_until"] = *req.ValidUntil
	}
	
	_, err := r.Update(ctx, updateData, "id", id)
	if err != nil {
		return nil, fmt.Errorf("更新定价规则失败: %w", err)
	}
	
	return r.GetPricingRuleByID(ctx, id)
}

// DeletePricingRule 删除定价规则
func (r *PricingRuleRepository) DeletePricingRule(ctx context.Context, id uint64) error {
	_, err := r.Delete(ctx, "id", id)
	if err != nil {
		return fmt.Errorf("删除定价规则失败: %w", err)
	}
	
	return nil
}

// GetBasePriceRule 获取商品的基础价格规则
func (r *PricingRuleRepository) GetBasePriceRule(ctx context.Context, productID uint64) (*types.ProductPricingRule, error) {
	record, err := r.FindOne(ctx, "product_id = ? AND rule_type = ?", productID, types.PricingRuleTypeBasePrice)
	if err != nil {
		return nil, fmt.Errorf("获取基础价格规则失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	rule := &types.ProductPricingRule{}
	err = record.Struct(rule)
	if err != nil {
		return nil, fmt.Errorf("基础价格规则数据转换失败: %w", err)
	}
	
	return rule, nil
}

// CheckRuleConflicts 检查规则冲突
func (r *PricingRuleRepository) CheckRuleConflicts(ctx context.Context, productID uint64, ruleType types.PricingRuleType, validFrom time.Time, validUntil *time.Time, excludeID *uint64) (bool, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return false, err
	}
	
	query := model.Where("product_id", productID).
		Where("rule_type", ruleType).
		Where("is_active", true)
	
	// 排除指定ID（用于更新时检查）
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	
	// 检查时间段重叠
	if validUntil != nil {
		query = query.Where("(valid_from <= ? AND (valid_until IS NULL OR valid_until > ?))", *validUntil, validFrom)
	} else {
		query = query.Where("valid_from <= ?", validFrom)
	}
	
	count, err := query.Count()
	if err != nil {
		return false, fmt.Errorf("检查规则冲突失败: %w", err)
	}
	
	return count > 0, nil
}

// GetRulesByType 按类型获取定价规则
func (r *PricingRuleRepository) GetRulesByType(ctx context.Context, productID uint64, ruleType types.PricingRuleType) ([]*types.ProductPricingRule, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Where("product_id", productID).
		Where("rule_type", ruleType).
		Where("is_active", true).
		Order("priority DESC, created_at ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("按类型获取定价规则失败: %w", err)
	}
	
	var rules []*types.ProductPricingRule
	for _, record := range result {
		rule := &types.ProductPricingRule{}
		err = record.Struct(rule)
		if err != nil {
			return nil, fmt.Errorf("定价规则数据转换失败: %w", err)
		}
		rules = append(rules, rule)
	}
	
	return rules, nil
}

// BatchUpdateRulesStatus 批量更新规则状态
func (r *PricingRuleRepository) BatchUpdateRulesStatus(ctx context.Context, ruleIDs []uint64, isActive bool) error {
	if len(ruleIDs) == 0 {
		return nil
	}
	
	updateData := map[string]interface{}{
		"is_active":  isActive,
		"updated_at": time.Now(),
	}
	
	model, err := r.Model(ctx)
	if err != nil {
		return err
	}
	
	_, err = model.WhereIn("id", ruleIDs).Update(updateData)
	if err != nil {
		return fmt.Errorf("批量更新规则状态失败: %w", err)
	}
	
	return nil
}

// GetRulesCount 获取商品定价规则数量统计
func (r *PricingRuleRepository) GetRulesCount(ctx context.Context, productID uint64) (map[types.PricingRuleType]int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Where("product_id", productID).
		Fields("rule_type, COUNT(*) as count").
		Group("rule_type").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取规则数量统计失败: %w", err)
	}
	
	counts := make(map[types.PricingRuleType]int)
	for _, record := range result {
		ruleType := types.PricingRuleType(record["rule_type"].String())
		count := record["count"].Int()
		counts[ruleType] = count
	}
	
	return counts, nil
}

// ValidateRuleData 验证规则数据完整性
func (r *PricingRuleRepository) ValidateRuleData(ctx context.Context, rule *types.ProductPricingRule) error {
	// 检查商品是否存在
	productRepo := NewProductRepository()
	exists, err := productRepo.Exists(ctx, "id", rule.ProductID)
	if err != nil {
		return fmt.Errorf("验证商品存在性失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("商品不存在: %d", rule.ProductID)
	}
	
	// 验证规则配置
	config, err := rule.RuleConfig.GetConfig()
	if err != nil {
		return fmt.Errorf("规则配置无效: %w", err)
	}
	
	err = config.Validate()
	if err != nil {
		return fmt.Errorf("规则配置验证失败: %w", err)
	}
	
	// 检查时间有效性
	if rule.ValidUntil != nil && rule.ValidUntil.Before(rule.ValidFrom) {
		return fmt.Errorf("规则失效时间不能早于生效时间")
	}
	
	return nil
}

// ArchiveExpiredRules 归档过期规则
func (r *PricingRuleRepository) ArchiveExpiredRules(ctx context.Context, beforeTime time.Time) (int, error) {
	updateData := map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}
	
	model, err := r.Model(ctx)
	if err != nil {
		return 0, err
	}
	
	result, err := model.Where("valid_until IS NOT NULL").
		Where("valid_until < ?", beforeTime).
		Where("is_active", true).
		Update(updateData)
	if err != nil {
		return 0, fmt.Errorf("归档过期规则失败: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}
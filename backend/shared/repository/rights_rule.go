package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// RightsRuleRepository 权益规则仓储
type RightsRuleRepository struct {
	*BaseRepository
}

// NewRightsRuleRepository 创建权益规则仓储实例
func NewRightsRuleRepository() *RightsRuleRepository {
	baseRepo := NewBaseRepository()
	baseRepo.tableName = "product_rights_rules"
	
	return &RightsRuleRepository{
		BaseRepository: baseRepo,
	}
}

// CreateRightsRule 创建权益规则
func (r *RightsRuleRepository) CreateRightsRule(ctx context.Context, productID uint64, req *types.CreateRightsRuleRequest) (*types.ProductRightsRule, error) {
	// 检查是否已存在权益规则
	exists, err := r.Exists(ctx, "product_id", productID)
	if err != nil {
		return nil, fmt.Errorf("检查权益规则存在性失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("商品已存在权益规则，每个商品只能有一个权益规则")
	}
	
	rule := &types.ProductRightsRule{
		ProductID:                productID,
		RuleType:                 req.RuleType,
		ConsumptionRate:          req.ConsumptionRate,
		MinRightsRequired:        req.MinRightsRequired,
		InsufficientRightsAction: req.InsufficientRightsAction,
		IsActive:                 true,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}
	
	id, err := r.InsertAndGetId(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("创建权益规则失败: %w", err)
	}
	
	rule.ID = uint64(id)
	rule.TenantID = r.GetTenantID(ctx)
	
	return rule, nil
}

// GetRightsRuleByID 根据ID获取权益规则
func (r *RightsRuleRepository) GetRightsRuleByID(ctx context.Context, id uint64) (*types.ProductRightsRule, error) {
	record, err := r.FindOne(ctx, "id", id)
	if err != nil {
		return nil, fmt.Errorf("获取权益规则失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	rule := &types.ProductRightsRule{}
	err = record.Struct(rule)
	if err != nil {
		return nil, fmt.Errorf("权益规则数据转换失败: %w", err)
	}
	
	return rule, nil
}

// GetRightsRuleByProductID 根据商品ID获取权益规则
func (r *RightsRuleRepository) GetRightsRuleByProductID(ctx context.Context, productID uint64) (*types.ProductRightsRule, error) {
	record, err := r.FindOne(ctx, "product_id", productID)
	if err != nil {
		return nil, fmt.Errorf("获取商品权益规则失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	rule := &types.ProductRightsRule{}
	err = record.Struct(rule)
	if err != nil {
		return nil, fmt.Errorf("权益规则数据转换失败: %w", err)
	}
	
	return rule, nil
}

// GetActiveRightsRuleByProductID 获取商品的有效权益规则
func (r *RightsRuleRepository) GetActiveRightsRuleByProductID(ctx context.Context, productID uint64) (*types.ProductRightsRule, error) {
	record, err := r.FindOne(ctx, "product_id = ? AND is_active = ?", productID, true)
	if err != nil {
		return nil, fmt.Errorf("获取有效权益规则失败: %w", err)
	}
	
	if record.IsEmpty() {
		return nil, nil
	}
	
	rule := &types.ProductRightsRule{}
	err = record.Struct(rule)
	if err != nil {
		return nil, fmt.Errorf("权益规则数据转换失败: %w", err)
	}
	
	return rule, nil
}

// UpdateRightsRule 更新权益规则
func (r *RightsRuleRepository) UpdateRightsRule(ctx context.Context, id uint64, req *types.UpdateRightsRuleRequest) (*types.ProductRightsRule, error) {
	updateData := make(map[string]interface{})
	updateData["updated_at"] = time.Now()
	
	if req.ConsumptionRate != nil {
		updateData["consumption_rate"] = *req.ConsumptionRate
	}
	if req.MinRightsRequired != nil {
		updateData["min_rights_required"] = *req.MinRightsRequired
	}
	if req.InsufficientRightsAction != nil {
		updateData["insufficient_rights_action"] = *req.InsufficientRightsAction
	}
	if req.IsActive != nil {
		updateData["is_active"] = *req.IsActive
	}
	
	_, err := r.Update(ctx, updateData, "id", id)
	if err != nil {
		return nil, fmt.Errorf("更新权益规则失败: %w", err)
	}
	
	return r.GetRightsRuleByID(ctx, id)
}

// DeleteRightsRule 删除权益规则
func (r *RightsRuleRepository) DeleteRightsRule(ctx context.Context, id uint64) error {
	_, err := r.Delete(ctx, "id", id)
	if err != nil {
		return fmt.Errorf("删除权益规则失败: %w", err)
	}
	
	return nil
}

// GetRightsRulesByType 根据规则类型获取权益规则
func (r *RightsRuleRepository) GetRightsRulesByType(ctx context.Context, ruleType types.RightsRuleType) ([]*types.ProductRightsRule, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Where("rule_type", ruleType).
		Where("is_active", true).
		Order("created_at ASC").
		All()
	if err != nil {
		return nil, fmt.Errorf("根据类型获取权益规则失败: %w", err)
	}
	
	var rules []*types.ProductRightsRule
	for _, record := range result {
		rule := &types.ProductRightsRule{}
		err = record.Struct(rule)
		if err != nil {
			return nil, fmt.Errorf("权益规则数据转换失败: %w", err)
		}
		rules = append(rules, rule)
	}
	
	return rules, nil
}

// CalculateRightsConsumption 计算权益消耗
func (r *RightsRuleRepository) CalculateRightsConsumption(ctx context.Context, productID uint64, quantity int, totalAmount types.Money) (float64, error) {
	rule, err := r.GetActiveRightsRuleByProductID(ctx, productID)
	if err != nil {
		return 0, err
	}
	
	if rule == nil {
		return 0, nil // 没有权益规则，不消耗权益
	}
	
	switch rule.RuleType {
	case types.RightsRuleTypeFixedRate:
		// 固定费率：消耗比例 * 数量
		return rule.ConsumptionRate * float64(quantity), nil
		
	case types.RightsRuleTypePercentage:
		// 百分比扣减：总金额 * 消耗比例
		totalAmountFloat := float64(totalAmount.Amount) / 100 // 转换为元
		return totalAmountFloat * rule.ConsumptionRate, nil
		
	case types.RightsRuleTypeTiered:
		// 阶梯消耗：根据数量计算阶梯消耗
		return r.calculateTieredConsumption(rule.ConsumptionRate, quantity), nil
		
	default:
		return 0, fmt.Errorf("未知的权益规则类型: %s", rule.RuleType)
	}
}

// calculateTieredConsumption 计算阶梯消耗
func (r *RightsRuleRepository) calculateTieredConsumption(baseRate float64, quantity int) float64 {
	// 简化的阶梯算法，可以根据业务需求调整
	// 1-10: baseRate
	// 11-50: baseRate * 0.9
	// 51+: baseRate * 0.8
	
	var totalConsumption float64
	remaining := quantity
	
	// 第一层：1-10
	if remaining > 0 {
		tier1 := remaining
		if tier1 > 10 {
			tier1 = 10
		}
		totalConsumption += float64(tier1) * baseRate
		remaining -= tier1
	}
	
	// 第二层：11-50
	if remaining > 0 {
		tier2 := remaining
		if tier2 > 40 {
			tier2 = 40
		}
		totalConsumption += float64(tier2) * baseRate * 0.9
		remaining -= tier2
	}
	
	// 第三层：51+
	if remaining > 0 {
		totalConsumption += float64(remaining) * baseRate * 0.8
	}
	
	return totalConsumption
}

// ValidateRightsBalance 验证权益余额
func (r *RightsRuleRepository) ValidateRightsBalance(ctx context.Context, productID, userID uint64, quantity int, totalAmount types.Money) (*types.ValidateRightsResponse, error) {
	rule, err := r.GetActiveRightsRuleByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}
	
	response := &types.ValidateRightsResponse{
		IsValid: true,
	}
	
	if rule == nil {
		// 没有权益规则，无需消耗权益
		return response, nil
	}
	
	// 计算所需权益
	requiredRights, err := r.CalculateRightsConsumption(ctx, productID, quantity, totalAmount)
	if err != nil {
		return nil, err
	}
	
	response.RequiredRights = requiredRights
	
	// 获取用户权益余额（这里需要调用用户权益服务）
	// TODO: 集成用户权益查询服务
	availableRights := r.getUserRightsBalance(ctx, userID)
	response.AvailableRights = availableRights
	
	// 检查余额是否充足
	if availableRights < requiredRights {
		response.IsValid = false
		response.InsufficientAmount = requiredRights - availableRights
		response.SuggestedAction = rule.InsufficientRightsAction
		
		// 如果是现金补足模式，计算需要的现金
		if rule.InsufficientRightsAction == types.InsufficientRightsActionCashPayment {
			// 简化计算：不足权益按1:1比例转换为现金（实际可能有汇率）
			response.CashPaymentRequired = types.Money{
				Amount:   response.InsufficientAmount, // 直接使用float64
				Currency: totalAmount.Currency,
			}
		}
	}
	
	return response, nil
}

// getUserRightsBalance 获取用户权益余额（临时实现）
func (r *RightsRuleRepository) getUserRightsBalance(ctx context.Context, userID uint64) float64 {
	// TODO: 实际实现中应该调用用户权益服务
	// 这里返回一个模拟值
	return 1000.0
}

// GetRightsRulesStatistics 获取权益规则统计信息
func (r *RightsRuleRepository) GetRightsRulesStatistics(ctx context.Context) (map[string]int, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	// 按规则类型统计
	result, err := model.Fields("rule_type, COUNT(*) as count").
		Where("is_active", true).
		Group("rule_type").
		All()
	if err != nil {
		return nil, fmt.Errorf("获取权益规则统计失败: %w", err)
	}
	
	stats := make(map[string]int)
	for _, record := range result {
		ruleType := record["rule_type"].String()
		count := record["count"].Int()
		stats[ruleType] = count
	}
	
	// 获取总数
	total, err := r.Count(ctx, "is_active", true)
	if err != nil {
		return nil, fmt.Errorf("获取权益规则总数失败: %w", err)
	}
	stats["total"] = total
	
	return stats, nil
}

// BatchUpdateRightsRulesStatus 批量更新权益规则状态
func (r *RightsRuleRepository) BatchUpdateRightsRulesStatus(ctx context.Context, productIDs []uint64, isActive bool) error {
	if len(productIDs) == 0 {
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
	
	_, err = model.WhereIn("product_id", productIDs).Update(updateData)
	if err != nil {
		return fmt.Errorf("批量更新权益规则状态失败: %w", err)
	}
	
	return nil
}

// ValidateRuleData 验证权益规则数据完整性
func (r *RightsRuleRepository) ValidateRuleData(ctx context.Context, rule *types.ProductRightsRule) error {
	// 检查商品是否存在
	productRepo := NewProductRepository()
	exists, err := productRepo.Exists(ctx, "id", rule.ProductID)
	if err != nil {
		return fmt.Errorf("验证商品存在性失败: %w", err)
	}
	if !exists {
		return fmt.Errorf("商品不存在: %d", rule.ProductID)
	}
	
	// 验证消耗比例
	if rule.ConsumptionRate < 0 {
		return fmt.Errorf("消耗比例不能为负数")
	}
	
	// 验证最低权益要求
	if rule.MinRightsRequired < 0 {
		return fmt.Errorf("最低权益要求不能为负数")
	}
	
	// 验证权益不足处理策略
	validActions := []types.InsufficientRightsAction{
		types.InsufficientRightsActionBlockPurchase,
		types.InsufficientRightsActionPartialPayment,
		types.InsufficientRightsActionCashPayment,
	}
	
	isValidAction := false
	for _, validAction := range validActions {
		if rule.InsufficientRightsAction == validAction {
			isValidAction = true
			break
		}
	}
	
	if !isValidAction {
		return fmt.Errorf("无效的权益不足处理策略: %s", rule.InsufficientRightsAction)
	}
	
	return nil
}

// GetProductsWithRightsRules 获取有权益规则的商品ID列表
func (r *RightsRuleRepository) GetProductsWithRightsRules(ctx context.Context, isActive bool) ([]uint64, error) {
	model, err := r.Model(ctx)
	if err != nil {
		return nil, err
	}
	
	result, err := model.Fields("product_id").
		Where("is_active", isActive).
		Distinct().
		All()
	if err != nil {
		return nil, fmt.Errorf("获取有权益规则的商品列表失败: %w", err)
	}
	
	var productIDs []uint64
	for _, record := range result {
		productIDs = append(productIDs, record["product_id"].Uint64())
	}
	
	return productIDs, nil
}
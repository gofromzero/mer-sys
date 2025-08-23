package service

import (
	"context"
	"fmt"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// PricingService 定价服务
type PricingService struct {
	pricingRuleRepo       *repository.PricingRuleRepository
	rightsRuleRepo        *repository.RightsRuleRepository
	promotionalPriceRepo  *repository.PromotionalPriceRepository
	priceHistoryRepo      *repository.PriceHistoryRepository
	productRepo           *repository.ProductRepository
}

// NewPricingService 创建定价服务实例
func NewPricingService() *PricingService {
	return &PricingService{
		pricingRuleRepo:       repository.NewPricingRuleRepository(),
		rightsRuleRepo:        repository.NewRightsRuleRepository(),
		promotionalPriceRepo:  repository.NewPromotionalPriceRepository(),
		priceHistoryRepo:      repository.NewPriceHistoryRepository(),
		productRepo:           repository.NewProductRepository(),
	}
}

// CreatePricingRule 创建定价规则
func (s *PricingService) CreatePricingRule(ctx context.Context, productID uint64, req *types.CreatePricingRuleRequest) (*types.ProductPricingRule, error) {
	// 验证商品存在
	exists, err := s.productRepo.Exists(ctx, "id", productID)
	if err != nil {
		return nil, fmt.Errorf("验证商品存在性失败: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("商品不存在")
	}
	
	// 验证规则配置
	config, err := req.RuleConfig.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("规则配置无效: %w", err)
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("规则配置验证失败: %w", err)
	}
	
	// 检查规则冲突
	hasConflict, err := s.pricingRuleRepo.CheckRuleConflicts(ctx, productID, req.RuleType, req.ValidFrom, req.ValidUntil, nil)
	if err != nil {
		return nil, fmt.Errorf("检查规则冲突失败: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("定价规则时间段冲突")
	}
	
	// 基础价格规则特殊处理：每个商品只能有一个
	if req.RuleType == types.PricingRuleTypeBasePrice {
		existing, err := s.pricingRuleRepo.GetBasePriceRule(ctx, productID)
		if err != nil {
			return nil, fmt.Errorf("检查基础价格规则失败: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("商品已存在基础价格规则")
		}
	}
	
	return s.pricingRuleRepo.CreatePricingRule(ctx, productID, req)
}

// GetPricingRulesByProductID 获取商品定价规则
func (s *PricingService) GetPricingRulesByProductID(ctx context.Context, productID uint64) ([]*types.ProductPricingRule, error) {
	return s.pricingRuleRepo.GetPricingRulesByProductID(ctx, productID)
}

// UpdatePricingRule 更新定价规则
func (s *PricingService) UpdatePricingRule(ctx context.Context, ruleID uint64, req *types.UpdatePricingRuleRequest) (*types.ProductPricingRule, error) {
	// 获取现有规则
	existing, err := s.pricingRuleRepo.GetPricingRuleByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("定价规则不存在")
	}
	
	// 验证规则配置（如果有更新）
	if req.RuleConfig != nil {
		config, err := req.RuleConfig.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("规则配置无效: %w", err)
		}
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("规则配置验证失败: %w", err)
		}
	}
	
	return s.pricingRuleRepo.UpdatePricingRule(ctx, ruleID, req)
}

// DeletePricingRule 删除定价规则
func (s *PricingService) DeletePricingRule(ctx context.Context, ruleID uint64) error {
	return s.pricingRuleRepo.DeletePricingRule(ctx, ruleID)
}

// CalculateEffectivePrice 计算有效价格
func (s *PricingService) CalculateEffectivePrice(ctx context.Context, productID uint64, req *types.CalculateEffectivePriceRequest) (*types.CalculateEffectivePriceResponse, error) {
	// 获取商品信息
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("获取商品信息失败: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("商品不存在")
	}
	
	response := &types.CalculateEffectivePriceResponse{
		AppliedRules: []string{},
	}
	
	// 获取基础价格
	baseRule, err := s.pricingRuleRepo.GetBasePriceRule(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("获取基础价格规则失败: %w", err)
	}
	
	if baseRule != nil {
		config, err := baseRule.RuleConfig.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("解析基础价格配置失败: %w", err)
		}
		
		if baseConfig, ok := config.(types.BasePriceConfig); ok {
			response.BasePrice = types.Money{
				Amount:   float64(baseConfig.Amount) / 100, // 转换分为元
				Currency: baseConfig.Currency,
			}
			response.EffectivePrice = response.BasePrice
			response.AppliedRules = append(response.AppliedRules, "基础价格")
		}
	} else {
		// 使用商品原始价格作为基础价格
		priceMoney := product.GetPrice()
		response.BasePrice = priceMoney
		response.EffectivePrice = response.BasePrice
	}
	
	// 检查促销价格
	promo, err := s.promotionalPriceRepo.GetActivePromotionalPrice(ctx, productID, req.RequestTime)
	if err != nil {
		return nil, fmt.Errorf("检查促销价格失败: %w", err)
	}
	
	if promo != nil {
		response.PromotionalPrice = &promo.PromotionalPrice
		response.EffectivePrice = promo.PromotionalPrice
		response.IsPromotionActive = true
		response.AppliedRules = append(response.AppliedRules, "促销价格")
		
		// 计算折扣金额
		if response.BasePrice.Amount > promo.PromotionalPrice.Amount {
			response.DiscountAmount = types.Money{
				Amount:   response.BasePrice.Amount - promo.PromotionalPrice.Amount,
				Currency: response.BasePrice.Currency,
			}
		}
	}
	
	// 应用其他定价规则（会员价格、阶梯价格等）
	if !response.IsPromotionActive {
		err = s.applyPricingRules(ctx, productID, req, response)
		if err != nil {
			return nil, fmt.Errorf("应用定价规则失败: %w", err)
		}
	}
	
	// 计算权益消耗
	rightsConsumption, err := s.rightsRuleRepo.CalculateRightsConsumption(ctx, productID, req.Quantity, response.EffectivePrice)
	if err != nil {
		return nil, fmt.Errorf("计算权益消耗失败: %w", err)
	}
	response.RightsConsumption = rightsConsumption
	
	return response, nil
}

// applyPricingRules 应用定价规则
func (s *PricingService) applyPricingRules(ctx context.Context, productID uint64, req *types.CalculateEffectivePriceRequest, response *types.CalculateEffectivePriceResponse) error {
	// 获取有效的定价规则
	rules, err := s.pricingRuleRepo.GetActivePricingRules(ctx, productID, req.RequestTime)
	if err != nil {
		return err
	}
	
	for _, rule := range rules {
		if rule.RuleType == types.PricingRuleTypeBasePrice {
			continue // 基础价格已处理
		}
		
		config, err := rule.RuleConfig.GetConfig()
		if err != nil {
			continue // 跳过无效配置
		}
		
		switch rule.RuleType {
		case types.PricingRuleTypeMemberDiscount:
			if req.MemberLevel != nil {
				err = s.applyMemberDiscount(config, req, response)
				if err == nil {
					response.AppliedRules = append(response.AppliedRules, "会员折扣")
				}
			}
			
		case types.PricingRuleTypeVolumeDiscount:
			err = s.applyVolumeDiscount(config, req, response)
			if err == nil {
				response.AppliedRules = append(response.AppliedRules, "阶梯折扣")
			}
			
		case types.PricingRuleTypeTimeBasedDiscount:
			err = s.applyTimeBasedDiscount(config, req, response)
			if err == nil {
				response.AppliedRules = append(response.AppliedRules, "时段折扣")
			}
		}
	}
	
	return nil
}

// applyMemberDiscount 应用会员折扣
func (s *PricingService) applyMemberDiscount(config types.PricingConfig, req *types.CalculateEffectivePriceRequest, response *types.CalculateEffectivePriceResponse) error {
	memberConfig, ok := config.(types.MemberDiscountConfig)
	if !ok {
		return fmt.Errorf("会员折扣配置类型错误")
	}
	
	if req.MemberLevel != nil {
		if price, exists := memberConfig.MemberLevels[*req.MemberLevel]; exists {
			oldPrice := response.EffectivePrice
			response.EffectivePrice = price
			
			// 计算折扣金额
			if oldPrice.Amount > price.Amount {
				response.DiscountAmount = types.Money{
					Amount:   oldPrice.Amount - price.Amount,
					Currency: oldPrice.Currency,
				}
			}
		}
	}
	
	return nil
}

// applyVolumeDiscount 应用阶梯折扣
func (s *PricingService) applyVolumeDiscount(config types.PricingConfig, req *types.CalculateEffectivePriceRequest, response *types.CalculateEffectivePriceResponse) error {
	volumeConfig, ok := config.(types.VolumeDiscountConfig)
	if !ok {
		return fmt.Errorf("阶梯折扣配置类型错误")
	}
	
	for _, tier := range volumeConfig.Tiers {
		if req.Quantity >= tier.MinQuantity && (tier.MaxQuantity == 0 || req.Quantity <= tier.MaxQuantity) {
			oldPrice := response.EffectivePrice
			response.EffectivePrice = tier.Price
			
			// 计算折扣金额
			if oldPrice.Amount > tier.Price.Amount {
				response.DiscountAmount = types.Money{
					Amount:   oldPrice.Amount - tier.Price.Amount,
					Currency: oldPrice.Currency,
				}
			}
			break
		}
	}
	
	return nil
}

// applyTimeBasedDiscount 应用时段折扣
func (s *PricingService) applyTimeBasedDiscount(config types.PricingConfig, req *types.CalculateEffectivePriceRequest, response *types.CalculateEffectivePriceResponse) error {
	timeConfig, ok := config.(types.TimeBasedDiscountConfig)
	if !ok {
		return fmt.Errorf("时段折扣配置类型错误")
	}
	
	currentWeekday := int(req.RequestTime.Weekday())
	currentTime := req.RequestTime
	
	for _, slot := range timeConfig.TimeSlots {
		// 检查星期
		validWeekday := false
		for _, day := range slot.WeekDays {
			if day == currentWeekday {
				validWeekday = true
				break
			}
		}
		
		if !validWeekday {
			continue
		}
		
		// 检查时间段（简化版，实际可能需要更复杂的逻辑）
		if currentTime.After(slot.StartTime) && currentTime.Before(slot.EndTime) {
			oldPrice := response.EffectivePrice
			response.EffectivePrice = slot.Price
			
			// 计算折扣金额
			if oldPrice.Amount > slot.Price.Amount {
				response.DiscountAmount = types.Money{
					Amount:   oldPrice.Amount - slot.Price.Amount,
					Currency: oldPrice.Currency,
				}
			}
			break
		}
	}
	
	return nil
}

// CreateRightsRule 创建权益规则
func (s *PricingService) CreateRightsRule(ctx context.Context, productID uint64, req *types.CreateRightsRuleRequest) (*types.ProductRightsRule, error) {
	// 验证商品存在
	exists, err := s.productRepo.Exists(ctx, "id", productID)
	if err != nil {
		return nil, fmt.Errorf("验证商品存在性失败: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("商品不存在")
	}
	
	return s.rightsRuleRepo.CreateRightsRule(ctx, productID, req)
}

// GetRightsRuleByProductID 获取商品权益规则
func (s *PricingService) GetRightsRuleByProductID(ctx context.Context, productID uint64) (*types.ProductRightsRule, error) {
	return s.rightsRuleRepo.GetRightsRuleByProductID(ctx, productID)
}

// UpdateRightsRule 更新权益规则
func (s *PricingService) UpdateRightsRule(ctx context.Context, ruleID uint64, req *types.UpdateRightsRuleRequest) (*types.ProductRightsRule, error) {
	return s.rightsRuleRepo.UpdateRightsRule(ctx, ruleID, req)
}

// ValidateRights 验证权益余额
func (s *PricingService) ValidateRights(ctx context.Context, productID uint64, req *types.ValidateRightsRequest) (*types.ValidateRightsResponse, error) {
	return s.rightsRuleRepo.ValidateRightsBalance(ctx, productID, req.UserID, req.Quantity, req.TotalAmount)
}

// CreatePromotionalPrice 创建促销价格
func (s *PricingService) CreatePromotionalPrice(ctx context.Context, productID uint64, req *types.CreatePromotionalPriceRequest) (*types.PromotionalPrice, error) {
	// 验证商品存在
	exists, err := s.productRepo.Exists(ctx, "id", productID)
	if err != nil {
		return nil, fmt.Errorf("验证商品存在性失败: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("商品不存在")
	}
	
	return s.promotionalPriceRepo.CreatePromotionalPrice(ctx, productID, req)
}

// GetPromotionalPricesByProductID 获取商品促销价格
func (s *PricingService) GetPromotionalPricesByProductID(ctx context.Context, productID uint64) ([]*types.PromotionalPrice, error) {
	return s.promotionalPriceRepo.GetPromotionalPricesByProductID(ctx, productID)
}

// UpdatePromotionalPrice 更新促销价格
func (s *PricingService) UpdatePromotionalPrice(ctx context.Context, promoID uint64, req *types.UpdatePromotionalPriceRequest) (*types.PromotionalPrice, error) {
	return s.promotionalPriceRepo.UpdatePromotionalPrice(ctx, promoID, req)
}

// DeletePromotionalPrice 删除促销价格
func (s *PricingService) DeletePromotionalPrice(ctx context.Context, promoID uint64) error {
	return s.promotionalPriceRepo.DeletePromotionalPrice(ctx, promoID)
}

// ChangePriceWithHistory 执行价格变更并记录历史
func (s *PricingService) ChangePriceWithHistory(ctx context.Context, productID uint64, req *types.PriceChangeRequest) error {
	// 获取当前价格
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("获取商品信息失败: %w", err)
	}
	if product == nil {
		return fmt.Errorf("商品不存在")
	}
	
	currentPrice := product.GetPrice()
	
	// 验证价格变更请求
	err = s.priceHistoryRepo.ValidatePriceChange(ctx, productID, req)
	if err != nil {
		return err
	}
	
	// 记录价格变更历史
	_, err = s.priceHistoryRepo.CreatePriceHistory(ctx, req, productID, currentPrice)
	if err != nil {
		return fmt.Errorf("记录价格历史失败: %w", err)
	}
	
	// 更新商品价格（这里可能需要调用ProductService的更新方法）
	updateReq := &types.UpdateProductRequest{
		Price: &req.NewPrice,
	}
	
	productService := NewProductService()
	_, err = productService.UpdateProduct(ctx, productID, updateReq)
	if err != nil {
		return fmt.Errorf("更新商品价格失败: %w", err)
	}
	
	return nil
}

// GetPriceHistoryPage 分页获取价格历史
func (s *PricingService) GetPriceHistoryPage(ctx context.Context, productID uint64, page, pageSize int) ([]*types.PriceHistory, int, error) {
	return s.priceHistoryRepo.GetPriceHistoryPage(ctx, productID, page, pageSize)
}
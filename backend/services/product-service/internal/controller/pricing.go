package controller

import (
	"time"

	"github.com/gofromzero/mer-sys/backend/services/product-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/middleware"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

// PricingController 定价控制器
type PricingController struct {
	pricingService *service.PricingService
}

// NewPricingController 创建定价控制器实例
func NewPricingController() *PricingController {
	return &PricingController{
		pricingService: service.NewPricingService(),
	}
}

// RegisterRoutes 注册带权限控制的路由
func (c *PricingController) RegisterRoutes(group *ghttp.RouterGroup) {
	// 定价规则管理路由（需要相应权限）
	group.POST("/pricing-rules", 
		middleware.RequirePricingPermission(middleware.PricingPermissionCreate),
		middleware.LogPricingOperation("create_pricing_rule"),
		c.CreatePricingRule)
	
	group.GET("/pricing-rules", 
		middleware.RequirePricingPermission(middleware.PricingPermissionRead),
		c.GetPricingRules)
	
	group.PUT("/pricing-rules/:rule_id", 
		middleware.RequirePricingPermission(middleware.PricingPermissionUpdate),
		middleware.LogPricingOperation("update_pricing_rule"),
		c.UpdatePricingRule)
	
	group.DELETE("/pricing-rules/:rule_id", 
		middleware.RequirePricingPermission(middleware.PricingPermissionDelete),
		middleware.LogPricingOperation("delete_pricing_rule"),
		c.DeletePricingRule)
	
	// 权益规则管理路由（需要权益管理权限）
	group.POST("/rights-rules", 
		middleware.RequirePricingPermission(middleware.RightsPermissionManage),
		middleware.LogPricingOperation("create_rights_rule"),
		c.CreateRightsRule)
	
	group.PUT("/rights-rules/:rule_id", 
		middleware.RequirePricingPermission(middleware.RightsPermissionManage),
		middleware.LogPricingOperation("update_rights_rule"),
		c.UpdateRightsRule)
	
	// 价格变更操作（需要价格变更权限）
	group.POST("/price-change", 
		middleware.RequirePricingPermission(middleware.PriceChangePermission),
		middleware.AuditLog("pricing"),
		middleware.LogPricingOperation("change_price"),
		c.ChangePriceWithHistory)
}

// CreatePricingRule 创建定价规则
func (c *PricingController) CreatePricingRule(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	var req types.CreatePricingRuleRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	rule, err := c.pricingService.CreatePricingRule(r.GetCtx(), productID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建定价规则失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "定价规则创建成功",
		"data":    rule,
	})
}

// GetPricingRules 获取商品定价规则
func (c *PricingController) GetPricingRules(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	rules, err := c.pricingService.GetPricingRulesByProductID(r.GetCtx(), productID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取定价规则失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取定价规则成功",
		"data":    rules,
	})
}

// UpdatePricingRule 更新定价规则
func (c *PricingController) UpdatePricingRule(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	ruleID := gconv.Uint64(r.Get("rule_id"))
	if productID == 0 || ruleID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID或规则ID无效",
		})
		return
	}
	
	var req types.UpdatePricingRuleRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	rule, err := c.pricingService.UpdatePricingRule(r.GetCtx(), ruleID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新定价规则失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "定价规则更新成功",
		"data":    rule,
	})
}

// DeletePricingRule 删除定价规则
func (c *PricingController) DeletePricingRule(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	ruleID := gconv.Uint64(r.Get("rule_id"))
	if productID == 0 || ruleID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID或规则ID无效",
		})
		return
	}
	
	err := c.pricingService.DeletePricingRule(r.GetCtx(), ruleID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "删除定价规则失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "定价规则删除成功",
	})
}

// GetEffectivePrice 计算商品有效价格
func (c *PricingController) GetEffectivePrice(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	var req types.CalculateEffectivePriceRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 设置默认值
	if req.Quantity == 0 {
		req.Quantity = 1
	}
	if req.RequestTime.IsZero() {
		req.RequestTime = time.Now()
	}
	
	response, err := c.pricingService.CalculateEffectivePrice(r.GetCtx(), productID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "计算有效价格失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "计算有效价格成功",
		"data":    response,
	})
}

// CreateRightsRule 创建权益规则
func (c *PricingController) CreateRightsRule(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	var req types.CreateRightsRuleRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	rule, err := c.pricingService.CreateRightsRule(r.GetCtx(), productID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建权益规则失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "权益规则创建成功",
		"data":    rule,
	})
}

// GetRightsRules 获取商品权益规则
func (c *PricingController) GetRightsRules(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	rule, err := c.pricingService.GetRightsRuleByProductID(r.GetCtx(), productID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取权益规则失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取权益规则成功",
		"data":    rule,
	})
}

// UpdateRightsRule 更新权益规则
func (c *PricingController) UpdateRightsRule(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	ruleID := gconv.Uint64(r.Get("rule_id"))
	if productID == 0 || ruleID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID或规则ID无效",
		})
		return
	}
	
	var req types.UpdateRightsRuleRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	rule, err := c.pricingService.UpdateRightsRule(r.GetCtx(), ruleID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新权益规则失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "权益规则更新成功",
		"data":    rule,
	})
}

// ValidateRights 验证权益余额
func (c *PricingController) ValidateRights(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	var req types.ValidateRightsRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	response, err := c.pricingService.ValidateRights(r.GetCtx(), productID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "验证权益失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "权益验证成功",
		"data":    response,
	})
}

// GetPriceHistory 获取价格变更历史
func (c *PricingController) GetPriceHistory(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	// 分页参数
	page := r.GetQuery("page", "1").Int()
	pageSize := r.GetQuery("page_size", "20").Int()
	if pageSize > 100 {
		pageSize = 100
	}
	
	histories, total, err := c.pricingService.GetPriceHistoryPage(r.GetCtx(), productID, page, pageSize)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取价格历史失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取价格历史成功",
		"data": g.Map{
			"histories":  histories,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

// ChangePriceWithHistory 执行价格变更
func (c *PricingController) ChangePriceWithHistory(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	var req types.PriceChangeRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	// 设置默认生效时间
	if req.EffectiveDate.IsZero() {
		req.EffectiveDate = time.Now()
	}

	// 获取当前价格用于审计 - 这里可以从产品服务获取
	// TODO: 从产品服务获取当前价格，暂时使用简化处理
	var currentProduct *types.Product

	// 执行价格变更
	err := c.pricingService.ChangePriceWithHistory(r.GetCtx(), productID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "价格变更失败",
			"error":   err.Error(),
		})
		return
	}

	// 记录价格变更审计信息
	if currentProduct != nil {
		oldPrice := types.Money{
			Amount:   currentProduct.PriceAmount,
			Currency: currentProduct.PriceCurrency,
		}
		middleware.RecordPriceChange(r, oldPrice, req.NewPrice, req.ChangeReason)
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "价格变更成功",
	})
}

// CreatePromotionalPrice 创建促销价格
func (c *PricingController) CreatePromotionalPrice(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	var req types.CreatePromotionalPriceRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	promo, err := c.pricingService.CreatePromotionalPrice(r.GetCtx(), productID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建促销价格失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "促销价格创建成功",
		"data":    promo,
	})
}

// GetPromotionalPrices 获取商品促销价格
func (c *PricingController) GetPromotionalPrices(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	if productID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID无效",
		})
		return
	}
	
	promos, err := c.pricingService.GetPromotionalPricesByProductID(r.GetCtx(), productID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取促销价格失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "获取促销价格成功",
		"data":    promos,
	})
}

// UpdatePromotionalPrice 更新促销价格
func (c *PricingController) UpdatePromotionalPrice(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	promoID := gconv.Uint64(r.Get("promo_id"))
	if productID == 0 || promoID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID或促销ID无效",
		})
		return
	}
	
	var req types.UpdatePromotionalPriceRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "参数解析失败",
			"error":   err.Error(),
		})
		return
	}
	
	promo, err := c.pricingService.UpdatePromotionalPrice(r.GetCtx(), promoID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新促销价格失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "促销价格更新成功",
		"data":    promo,
	})
}

// DeletePromotionalPrice 删除促销价格
func (c *PricingController) DeletePromotionalPrice(r *ghttp.Request) {
	productID := gconv.Uint64(r.Get("product_id"))
	promoID := gconv.Uint64(r.Get("promo_id"))
	if productID == 0 || promoID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商品ID或促销ID无效",
		})
		return
	}
	
	err := c.pricingService.DeletePromotionalPrice(r.GetCtx(), promoID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "删除促销价格失败",
			"error":   err.Error(),
		})
		return
	}
	
	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "促销价格删除成功",
	})
}
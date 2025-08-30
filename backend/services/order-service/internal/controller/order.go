package controller

import (
	"strconv"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/order-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

// OrderController 订单控制器
type OrderController struct {
	orderService service.IOrderService
	orderRepo    repository.IOrderRepository
}

// NewOrderController 创建订单控制器实例
func NewOrderController() *OrderController {
	return &OrderController{
		orderService: service.NewOrderService(),
		orderRepo:    repository.NewOrderRepository(),
	}
}

// CreateOrder 创建订单
func (c *OrderController) CreateOrder(r *ghttp.Request) {
	var req types.CreateOrderRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 获取当前用户ID作为客户ID
	customerID := r.GetCtxVar("user_id").Uint64()

	order, err := c.orderService.CreateOrder(r.Context(), customerID, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建订单失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "订单创建成功",
		"data":    order,
	})
}

// GetOrder 获取订单详情
func (c *OrderController) GetOrder(r *ghttp.Request) {
	orderID := r.Get("order_id").Uint64()
	if orderID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	order, err := c.orderService.GetOrder(r.Context(), orderID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    404,
			"message": "订单不存在",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": order,
	})
}

// ListOrders 获取订单列表
func (c *OrderController) ListOrders(r *ghttp.Request) {
	customerID := r.GetCtxVar("user_id").Uint64()

	// 解析查询参数
	status := gconv.Int(r.Get("status"))
	page := gconv.Int(r.Get("page", 1))
	limit := gconv.Int(r.Get("limit", 20))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 将状态码转换为字符串
	var orderStatus types.OrderStatus
	switch status {
	case 1:
		orderStatus = "pending"
	case 2:
		orderStatus = "paid"
	case 3:
		orderStatus = "processing"
	case 4:
		orderStatus = "completed"
	case 5:
		orderStatus = "cancelled"
	default:
		orderStatus = ""
	}
	
	orders, total, err := c.orderService.ListOrders(r.Context(), customerID, orderStatus, page, limit)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取订单列表失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": g.Map{
			"items": orders,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// CancelOrder 取消订单
func (c *OrderController) CancelOrder(r *ghttp.Request) {
	orderID := r.Get("order_id").Uint64()
	if orderID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "订单ID不能为空",
		})
		return
	}

	err := c.orderService.CancelOrder(r.Context(), orderID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "取消订单失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "订单取消成功",
	})
}

// QueryOrders 高级订单查询
// @Summary 高级订单查询
// @Description 支持多维度筛选和排序的订单查询接口
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param merchant_id query int false "商户ID"
// @Param customer_id query int false "客户ID"
// @Param status query []int false "订单状态列表"
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Param search_keyword query string false "搜索关键词（订单号或商品名称）"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param sort_by query string false "排序字段" Enums(created_at,updated_at,total_amount)
// @Param sort_order query string false "排序方式" Enums(asc,desc)
// @Success 200 {object} utils.Response{data=types.OrderListResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/orders/query [get]
func (c *OrderController) QueryOrders(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 构建查询请求
	req := &types.OrderQueryRequest{
		Page:      r.Get("page", 1).Int(),
		PageSize:  r.Get("page_size", 10).Int(),
		SortBy:    r.Get("sort_by", "created_at").String(),
		SortOrder: r.Get("sort_order", "desc").String(),
	}
	
	// 可选参数
	if merchantID := r.Get("merchant_id").Int(); merchantID > 0 {
		mid := uint64(merchantID)
		req.MerchantID = &mid
	}
	
	if customerID := r.Get("customer_id").Int(); customerID > 0 {
		cid := uint64(customerID)
		req.CustomerID = &cid
	}
	
	// 处理状态列表
	if statusStr := r.Get("status").String(); statusStr != "" {
		statusList := r.Get("status").Ints()
		req.Status = make([]types.OrderStatusInt, len(statusList))
		for i, s := range statusList {
			req.Status[i] = types.OrderStatusInt(s)
		}
	}
	
	// 处理日期范围
	if startDateStr := r.Get("start_date").String(); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		}
	}
	
	if endDateStr := r.Get("end_date").String(); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		}
	}
	
	// 搜索关键词
	if keyword := r.Get("search_keyword").String(); keyword != "" {
		req.SearchKeyword = &keyword
	}
	
	// 验证请求参数
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		utils.ErrorResponse(r, 400, "参数验证失败: " + err.Error())
		return
	}
	
	// 执行查询
	response, err := c.orderRepo.QueryList(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "订单查询失败: %v", err)
		utils.ErrorResponse(r, 500, "订单查询失败: " + err.Error())
		return
	}
	
	utils.SuccessResponse(r, response)
}

// GetOrderWithHistory 获取订单详情（包含状态历史）
// @Summary 获取订单详情（包含状态历史）
// @Description 获取指定订单的详细信息，包含完整的状态变更历史
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param order_id path int true "订单ID"
// @Success 200 {object} utils.Response{data=types.Order} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "订单不存在"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/orders/{order_id}/detail [get]
func (c *OrderController) GetOrderWithHistory(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	// 获取订单ID
	orderIDStr := r.Get("order_id").String()
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "订单ID格式错误")
		return
	}
	
	// 获取订单详情（包含历史）
	order, err := c.orderRepo.GetByIDWithHistory(ctx, orderID)
	if err != nil {
		g.Log().Errorf(ctx, "获取订单详情失败: %v", err)
		utils.ErrorResponse(r, 500, "获取订单详情失败: " + err.Error())
		return
	}
	
	if order == nil {
		utils.ErrorResponse(r, 404, "订单不存在")
		return
	}
	
	utils.SuccessResponse(r, order)
}

// SearchOrders 订单搜索
// @Summary 订单搜索
// @Description 根据关键词搜索订单（支持订单号、商品名称等）
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} utils.Response{data=types.OrderListResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/orders/search [get]
func (c *OrderController) SearchOrders(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	keyword := r.Get("q").String()
	if keyword == "" {
		utils.ErrorResponse(r, 400, "搜索关键词不能为空")
		return
	}
	
	// 构建搜索请求
	req := &types.OrderQueryRequest{
		SearchKeyword: &keyword,
		Page:          r.Get("page", 1).Int(),
		PageSize:      r.Get("page_size", 10).Int(),
		SortBy:        "created_at",
		SortOrder:     "desc",
	}
	
	// 执行搜索
	response, err := c.orderRepo.QueryList(ctx, req)
	if err != nil {
		g.Log().Errorf(ctx, "订单搜索失败: %v", err)
		utils.ErrorResponse(r, 500, "订单搜索失败: " + err.Error())
		return
	}
	
	utils.SuccessResponse(r, response)
}

// GetOrderStats 获取订单统计信息
// @Summary 获取订单统计信息
// @Description 获取订单状态分布统计信息
// @Tags 订单管理
// @Accept json
// @Produce json
// @Param merchant_id query int false "商户ID（可选）"
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} utils.Response{data=map[string]interface{}} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/orders/stats [get]
func (c *OrderController) GetOrderStats(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	var merchantID *uint64
	if mid := r.Get("merchant_id").Int(); mid > 0 {
		uid := uint64(mid)
		merchantID = &uid
	}
	
	var startDate, endDate *time.Time
	if startDateStr := r.Get("start_date").String(); startDateStr != "" {
		if sd, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &sd
		}
	}
	
	if endDateStr := r.Get("end_date").String(); endDateStr != "" {
		if ed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &ed
		}
	}
	
	// 统计各状态订单数量
	stats := map[string]interface{}{
		"total": 0,
		"by_status": map[string]int64{
			"pending":    0,
			"paid":       0,
			"processing": 0,
			"completed":  0,
			"cancelled":  0,
		},
	}
	
	// 构建统计查询条件
	for statusInt := types.OrderStatusIntPending; statusInt <= types.OrderStatusIntCancelled; statusInt++ {
		req := &types.OrderQueryRequest{
			MerchantID: merchantID,
			Status:     []types.OrderStatusInt{statusInt},
			StartDate:  startDate,
			EndDate:    endDate,
			Page:       1,
			PageSize:   1, // 只需要总数
			SortBy:     "created_at",
			SortOrder:  "desc",
		}
		
		response, err := c.orderRepo.QueryList(ctx, req)
		if err != nil {
			g.Log().Errorf(ctx, "统计订单失败: %v", err)
			continue
		}
		
		statusName := statusInt.String()
		stats["by_status"].(map[string]int64)[statusName] = response.Total
		stats["total"] = stats["total"].(int64) + response.Total
	}
	
	utils.SuccessResponse(r, stats)
}

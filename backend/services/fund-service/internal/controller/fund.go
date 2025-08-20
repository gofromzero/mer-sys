package controller

import (
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"mer-demo/services/fund-service/internal/service"
	"mer-demo/shared/types"
)

// FundController 资金管理控制器
type FundController struct {
	fundService service.FundService
}

// NewFundController 创建资金管理控制器
func NewFundController() *FundController {
	return &FundController{
		fundService: service.NewFundService(),
	}
}

// Deposit 单笔资金充值
func (c *FundController) Deposit(r *ghttp.Request) {
	var req types.DepositRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "数据验证失败: " + err.Error(),
		})
		return
	}

	// 获取操作人ID
	operatorID := getUserIDFromRequest(r)
	if operatorID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "无效的用户上下文",
		})
		return
	}

	// 执行充值
	fund, err := c.fundService.Deposit(r.Context(), &req, operatorID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "充值失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "充值成功",
		"data":    fund,
	})
}

// BatchDeposit 批量资金充值
func (c *FundController) BatchDeposit(r *ghttp.Request) {
	var req types.BatchDepositRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "数据验证失败: " + err.Error(),
		})
		return
	}

	// 获取操作人ID
	operatorID := getUserIDFromRequest(r)
	if operatorID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "无效的用户上下文",
		})
		return
	}

	// 执行批量充值
	results, err := c.fundService.BatchDeposit(r.Context(), &req, operatorID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "批量充值失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "批量充值完成",
		"data":    results,
	})
}

// Allocate 权益分配
func (c *FundController) Allocate(r *ghttp.Request) {
	var req types.AllocateRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证请求数据
	if err := req.Validate(); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "数据验证失败: " + err.Error(),
		})
		return
	}

	// 获取操作人ID
	operatorID := getUserIDFromRequest(r)
	if operatorID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "无效的用户上下文",
		})
		return
	}

	// 执行权益分配
	fund, err := c.fundService.Allocate(r.Context(), &req, operatorID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "权益分配失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "权益分配成功",
		"data":    fund,
	})
}

// GetBalance 查询商户权益余额
func (c *FundController) GetBalance(r *ghttp.Request) {
	merchantIDStr := r.Get("merchant_id").String()
	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil || merchantID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的商户ID",
		})
		return
	}

	// 获取权益余额
	balance, err := c.fundService.GetMerchantBalance(r.Context(), merchantID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "查询余额失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": balance,
	})
}

// ListTransactions 查询资金流转历史
func (c *FundController) ListTransactions(r *ghttp.Request) {
	var query types.FundTransactionQuery
	if err := r.Parse(&query); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认分页参数
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	// 查询资金流转记录
	transactions, total, err := c.fundService.ListTransactions(r.Context(), &query)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": g.Map{
			"list":      transactions,
			"total":     total,
			"page":      query.Page,
			"page_size": query.PageSize,
		},
	})
}

// GetSummary 获取资金概览统计
func (c *FundController) GetSummary(r *ghttp.Request) {
	var merchantID *uint64
	if merchantIDStr := r.Get("merchant_id").String(); merchantIDStr != "" {
		if id, err := strconv.ParseUint(merchantIDStr, 10, 64); err == nil && id > 0 {
			merchantID = &id
		}
	}

	// 获取资金概览统计
	summary, err := c.fundService.GetFundSummary(r.Context(), merchantID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "查询统计失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"data": summary,
	})
}

// FreezeBalance 冻结/解冻商户权益
func (c *FundController) FreezeBalance(r *ghttp.Request) {
	merchantIDStr := r.Get("merchant_id").String()
	merchantID, err := strconv.ParseUint(merchantIDStr, 10, 64)
	if err != nil || merchantID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "无效的商户ID",
		})
		return
	}

	var req struct {
		Action string  `json:"action" binding:"required,oneof=freeze unfreeze"`
		Amount float64 `json:"amount" binding:"required,gt=0"`
		Reason string  `json:"reason,omitempty"`
	}

	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取操作人ID
	operatorID := getUserIDFromRequest(r)
	if operatorID == 0 {
		r.Response.WriteJsonExit(g.Map{
			"code":    401,
			"message": "无效的用户上下文",
		})
		return
	}

	// 执行冻结/解冻操作
	err = c.fundService.FreezeMerchantBalance(r.Context(), merchantID, req.Action, req.Amount, req.Reason, operatorID)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "操作失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "操作成功",
	})
}

// getUserIDFromRequest 从请求中获取用户ID
func getUserIDFromRequest(r *ghttp.Request) uint64 {
	if userID := r.GetCtxVar("user_id"); userID != nil {
		return gconv.Uint64(userID)
	}
	return 0
}
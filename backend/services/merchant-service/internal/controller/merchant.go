package controller

import (
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/services/merchant-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
)

// MerchantController 商户控制器
type MerchantController struct {
	service *service.MerchantService
}

// NewMerchantController 创建商户控制器实例
func NewMerchantController() *MerchantController {
	return &MerchantController{
		service: service.NewMerchantService(),
	}
}

// Create 创建商户（注册申请）
func (c *MerchantController) Create(r *ghttp.Request) {
	var req types.MerchantRegistrationRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	merchant, err := c.service.RegisterMerchant(r.GetCtx(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "商户注册失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "商户注册申请已提交，等待审核",
		"data":    merchant,
	})
}

// List 获取商户列表
func (c *MerchantController) List(r *ghttp.Request) {
	var query types.MerchantListQuery
	if err := r.Parse(&query); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	merchants, total, err := c.service.GetMerchantList(r.GetCtx(), &query)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取商户列表失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "success",
		"data": g.Map{
			"items":     merchants,
			"total":     total,
			"page":      query.Page,
			"page_size": query.PageSize,
		},
	})
}

// GetByID 获取商户详情
func (c *MerchantController) GetByID(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商户ID格式错误",
		})
		return
	}

	merchant, err := c.service.GetMerchantByID(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取商户信息失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "success",
		"data":    merchant,
	})
}

// Update 更新商户信息
func (c *MerchantController) Update(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商户ID格式错误",
		})
		return
	}

	var req types.MerchantUpdateRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	merchant, err := c.service.UpdateMerchant(r.GetCtx(), id, &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新商户信息失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "商户信息更新成功",
		"data":    merchant,
	})
}

// UpdateStatus 更新商户状态
func (c *MerchantController) UpdateStatus(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商户ID格式错误",
		})
		return
	}

	var req types.MerchantStatusUpdateRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	err = c.service.UpdateMerchantStatus(r.GetCtx(), id, req.Status, req.Comment)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新商户状态失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "商户状态更新成功",
	})
}

// Approve 审批商户申请
func (c *MerchantController) Approve(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商户ID格式错误",
		})
		return
	}

	var req types.MerchantApprovalRequest
	req.Action = "approve"
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	err = c.service.ApproveMerchant(r.GetCtx(), id, req.Comment)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "商户审批失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "商户审批成功",
	})
}

// Reject 拒绝商户申请
func (c *MerchantController) Reject(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商户ID格式错误",
		})
		return
	}

	var req types.MerchantApprovalRequest
	req.Action = "reject"
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	err = c.service.RejectMerchant(r.GetCtx(), id, req.Comment)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "商户拒绝失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "商户申请已拒绝",
	})
}

// GetAuditLog 获取商户操作历史
func (c *MerchantController) GetAuditLog(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "商户ID格式错误",
		})
		return
	}

	logs, err := c.service.GetMerchantAuditLog(r.GetCtx(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取审计日志失败",
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    200,
		"message": "success",
		"data":    logs,
	})
}
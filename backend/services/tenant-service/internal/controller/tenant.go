package controller

import (
	"strconv"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/services/tenant-service/internal/service"
)

type TenantController struct {
	tenantService service.ITenantService
}

func NewTenantController() *TenantController {
	return &TenantController{
		tenantService: service.NewTenantService(),
	}
}

// Create handles POST /api/v1/tenants - 租户注册
func (c *TenantController) Create(r *ghttp.Request) {
	var req *types.CreateTenantRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	tenant, err := c.tenantService.CreateTenant(r.Context(), req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "创建租户失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "租户创建成功",
		"data":    tenant,
	})
}

// List handles GET /api/v1/tenants - 租户列表查询
func (c *TenantController) List(r *ghttp.Request) {
	var req types.ListTenantsRequest
	
	// 解析查询参数
	req.Page = r.GetQuery("page", 1).Int()
	req.PageSize = r.GetQuery("page_size", 20).Int()
	req.Status = types.TenantStatus(r.GetQuery("status").String())
	req.BusinessType = r.GetQuery("business_type").String()
	req.Search = r.GetQuery("search").String()

	result, err := c.tenantService.ListTenants(r.Context(), &req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取租户列表失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "获取租户列表成功",
		"data":    result,
	})
}

// GetByID handles GET /api/v1/tenants/{id} - 获取特定租户信息
func (c *TenantController) GetByID(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "租户ID格式错误",
			"data":    nil,
		})
		return
	}

	tenant, err := c.tenantService.GetTenantByID(r.Context(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取租户信息失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	if tenant == nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    404,
			"message": "租户不存在",
			"data":    nil,
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "获取租户信息成功",
		"data":    tenant,
	})
}

// Update handles PUT /api/v1/tenants/{id} - 更新租户信息
func (c *TenantController) Update(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "租户ID格式错误",
			"data":    nil,
		})
		return
	}

	var req *types.UpdateTenantRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	tenant, err := c.tenantService.UpdateTenant(r.Context(), id, req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新租户信息失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "更新租户信息成功",
		"data":    tenant,
	})
}

// UpdateStatus handles PUT /api/v1/tenants/{id}/status - 变更租户状态
func (c *TenantController) UpdateStatus(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "租户ID格式错误",
			"data":    nil,
		})
		return
	}

	var req *types.UpdateTenantStatusRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	err = c.tenantService.UpdateTenantStatus(r.Context(), id, req)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新租户状态失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "更新租户状态成功",
		"data":    nil,
	})
}

// GetConfig handles GET /api/v1/tenants/{id}/config - 获取租户配置
func (c *TenantController) GetConfig(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "租户ID格式错误",
			"data":    nil,
		})
		return
	}

	config, err := c.tenantService.GetTenantConfig(r.Context(), id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "获取租户配置失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "获取租户配置成功",
		"data":    config,
	})
}

// UpdateConfig handles PUT /api/v1/tenants/{id}/config - 更新租户配置
func (c *TenantController) UpdateConfig(r *ghttp.Request) {
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "租户ID格式错误",
			"data":    nil,
		})
		return
	}

	var config *types.TenantConfig
	if err := r.Parse(&config); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    400,
			"message": "请求参数格式错误",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	err = c.tenantService.UpdateTenantConfig(r.Context(), id, config)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code":    500,
			"message": "更新租户配置失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"code":    0,
		"message": "更新租户配置成功",
		"data":    nil,
	})
}
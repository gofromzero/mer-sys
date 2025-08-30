package controller

import (
	"strconv"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TemplateController 报表模板控制器
type TemplateController struct {
	templateService service.ITemplateService
}

// NewTemplateController 创建报表模板控制器实例
func NewTemplateController() *TemplateController {
	return &TemplateController{
		templateService: service.NewTemplateService(),
	}
}

// CreateTemplate 创建报表模板
// @Summary 创建报表模板
// @Description 创建新的报表模板配置
// @Tags 报表模板
// @Accept json
// @Produce json
// @Param request body types.ReportTemplate true "报表模板配置"
// @Success 200 {object} utils.Response{data=types.ReportTemplate}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/report-templates [post]
func (c *TemplateController) CreateTemplate(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	var template types.ReportTemplate
	if err := r.Parse(&template); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败")
		return
	}
	
	createdTemplate, err := c.templateService.CreateTemplate(ctx, &template)
	if err != nil {
		g.Log().Error(ctx, "创建报表模板失败", "error", err)
		utils.ErrorResponse(r, 500, "创建报表模板失败")
		return
	}
	
	utils.SuccessResponse(r, createdTemplate)
}

// GetTemplate 获取报表模板
// @Summary 获取报表模板
// @Description 根据模板ID获取报表模板详情
// @Tags 报表模板
// @Accept json
// @Produce json
// @Param id path uint64 true "模板ID"
// @Success 200 {object} utils.Response{data=types.ReportTemplate}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/report-templates/{id} [get]
func (c *TemplateController) GetTemplate(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "无效的模板ID")
		return
	}
	
	template, err := c.templateService.GetTemplate(ctx, id)
	if err != nil {
		g.Log().Error(ctx, "获取报表模板失败", "template_id", id, "error", err)
		utils.ErrorResponse(r, 404, "报表模板不存在")
		return
	}
	
	utils.SuccessResponse(r, template)
}

// ListTemplates 获取报表模板列表
// @Summary 获取报表模板列表
// @Description 获取报表模板列表，支持按类型筛选
// @Tags 报表模板
// @Accept json
// @Produce json
// @Param report_type query string false "报表类型"
// @Success 200 {object} utils.Response{data=[]types.ReportTemplate}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/report-templates [get]
func (c *TemplateController) ListTemplates(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	reportTypeStr := r.Get("report_type").String()
	var reportType *types.ReportType
	if reportTypeStr != "" {
		rt := types.ReportType(reportTypeStr)
		reportType = &rt
	}
	
	templates, err := c.templateService.ListTemplates(ctx, reportType)
	if err != nil {
		g.Log().Error(ctx, "获取报表模板列表失败", "error", err)
		utils.ErrorResponse(r, 500, "获取报表模板列表失败")
		return
	}
	
	utils.SuccessResponse(r, templates)
}

// UpdateTemplate 更新报表模板
// @Summary 更新报表模板
// @Description 更新指定的报表模板配置
// @Tags 报表模板
// @Accept json
// @Produce json
// @Param id path uint64 true "模板ID"
// @Param request body types.ReportTemplate true "报表模板配置"
// @Success 200 {object} utils.Response{data=types.ReportTemplate}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/report-templates/{id} [put]
func (c *TemplateController) UpdateTemplate(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "无效的模板ID")
		return
	}
	
	var template types.ReportTemplate
	if err := r.Parse(&template); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败")
		return
	}
	
	template.ID = id
	updatedTemplate, err := c.templateService.UpdateTemplate(ctx, &template)
	if err != nil {
		g.Log().Error(ctx, "更新报表模板失败", "template_id", id, "error", err)
		utils.ErrorResponse(r, 500, "更新报表模板失败")
		return
	}
	
	utils.SuccessResponse(r, updatedTemplate)
}

// DeleteTemplate 删除报表模板
// @Summary 删除报表模板
// @Description 删除指定的报表模板
// @Tags 报表模板
// @Accept json
// @Produce json
// @Param id path uint64 true "模板ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/report-templates/{id} [delete]
func (c *TemplateController) DeleteTemplate(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "无效的模板ID")
		return
	}
	
	err = c.templateService.DeleteTemplate(ctx, id)
	if err != nil {
		g.Log().Error(ctx, "删除报表模板失败", "template_id", id, "error", err)
		utils.ErrorResponse(r, 500, "删除报表模板失败")
		return
	}
	
	utils.SuccessResponse(r, g.Map{
		"message": "报表模板删除成功",
	})
}

// ScheduleReport 调度报表生成
// @Summary 调度报表生成
// @Description 基于模板创建定时报表任务
// @Tags 报表模板
// @Accept json
// @Produce json
// @Param request body types.ReportScheduleRequest true "报表调度参数"
// @Success 200 {object} utils.Response{data=types.ReportTemplate}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/report-templates/schedule [post]
func (c *TemplateController) ScheduleReport(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	var req types.ReportScheduleRequest
	if err := r.Parse(&req); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败")
		return
	}
	
	template, err := c.templateService.ScheduleReport(ctx, &req)
	if err != nil {
		g.Log().Error(ctx, "创建定时报表失败", "error", err)
		utils.ErrorResponse(r, 500, "创建定时报表失败")
		return
	}
	
	utils.SuccessResponse(r, template)
}
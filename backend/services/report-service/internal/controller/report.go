package controller

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gofromzero/mer-sys/backend/shared/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ReportController 报表控制器
type ReportController struct {
	generatorService service.IReportGeneratorService
	analyticsService service.IAnalyticsService
}

// NewReportController 创建报表控制器实例
func NewReportController() *ReportController {
	return &ReportController{
		generatorService: service.NewReportGeneratorService(),
		analyticsService: service.NewAnalyticsService(),
	}
}

// GenerateReport 生成报表
// @Summary 生成报表
// @Description 根据参数生成指定类型的报表
// @Tags 报表管理
// @Accept json
// @Produce json
// @Param request body types.ReportCreateRequest true "报表生成参数"
// @Success 200 {object} utils.Response{data=types.Report}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/reports/generate [post]
func (c *ReportController) GenerateReport(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	var req types.ReportCreateRequest
	if err := r.Parse(&req); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败")
		return
	}
	
	report, err := c.generatorService.GenerateReport(ctx, &req)
	if err != nil {
		g.Log().Error(ctx, "生成报表失败", "error", err)
		utils.ErrorResponse(r, 500, "生成报表失败")
		return
	}
	
	utils.SuccessResponse(r, report)
}

// GetReport 获取报表信息
// @Summary 获取报表信息
// @Description 根据报表ID获取报表详细信息
// @Tags 报表管理
// @Accept json
// @Produce json
// @Param id path uint64 true "报表ID"
// @Success 200 {object} utils.Response{data=types.Report}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/reports/{id} [get]
func (c *ReportController) GetReport(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "无效的报表ID")
		return
	}
	
	report, err := c.generatorService.GetReport(ctx, id)
	if err != nil {
		g.Log().Error(ctx, "获取报表失败", "report_id", id, "error", err)
		utils.ErrorResponse(r, 404, "报表不存在")
		return
	}
	
	utils.SuccessResponse(r, report)
}

// ListReports 获取报表列表
// @Summary 获取报表列表
// @Description 获取用户的报表列表，支持筛选和分页
// @Tags 报表管理
// @Accept json
// @Produce json
// @Param report_type query string false "报表类型"
// @Param status query string false "报表状态"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param page query int true "页码" default(1)
// @Param page_size query int true "每页大小" default(20)
// @Success 200 {object} utils.Response{data=object}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/reports [get]
func (c *ReportController) ListReports(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	var req types.ReportListRequest
	if err := r.Parse(&req); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败")
		return
	}
	
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	
	reports, total, err := c.generatorService.ListReports(ctx, &req)
	if err != nil {
		g.Log().Error(ctx, "获取报表列表失败", "error", err)
		utils.ErrorResponse(r, 500, "获取报表列表失败")
		return
	}
	
	utils.SuccessResponse(r, g.Map{
		"items":     reports,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"has_next":  (req.Page * req.PageSize) < total,
	})
}

// DeleteReport 删除报表
// @Summary 删除报表
// @Description 删除指定的报表文件和记录
// @Tags 报表管理
// @Accept json
// @Produce json
// @Param id path uint64 true "报表ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/reports/{id} [delete]
func (c *ReportController) DeleteReport(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	idStr := r.Get("id").String()
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(r, 400, "无效的报表ID")
		return
	}
	
	err = c.generatorService.DeleteReport(ctx, id)
	if err != nil {
		g.Log().Error(ctx, "删除报表失败", "report_id", id, "error", err)
		utils.ErrorResponse(r, 500, "删除报表失败")
		return
	}
	
	utils.SuccessResponse(r, g.Map{
		"message": "报表删除成功",
	})
}

// DownloadReport 下载报表
// @Summary 下载报表
// @Description 下载指定的报表文件
// @Tags 报表管理
// @Accept json
// @Produce application/octet-stream
// @Param uuid path string true "报表UUID"
// @Success 200 {file} file
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/reports/{uuid}/download [get]
func (c *ReportController) DownloadReport(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	uuid := r.Get("uuid").String()
	if uuid == "" {
		utils.ErrorResponse(r, 400, "报表UUID不能为空")
		return
	}
	
	filePath, err := c.generatorService.DownloadReport(ctx, uuid)
	if err != nil {
		g.Log().Error(ctx, "下载报表失败", "uuid", uuid, "error", err)
		utils.ErrorResponse(r, 404, "报表文件不存在或未生成完成")
		return
	}
	
	// 返回文件
	r.Response.ServeFileDownload(filePath)
}

// GetFinancialAnalytics 获取财务分析数据
// @Summary 获取财务分析数据
// @Description 获取指定时间范围的财务分析数据
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期" format(date)
// @Param end_date query string true "结束日期" format(date)
// @Param merchant_id query uint64 false "商户ID"
// @Success 200 {object} utils.Response{data=types.FinancialReportData}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/analytics/financial [get]
func (c *ReportController) GetFinancialAnalytics(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	startDateStr := r.Get("start_date").String()
	endDateStr := r.Get("end_date").String()
	merchantIDStr := r.Get("merchant_id").String()
	
	startDate, err := parseDate(startDateStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "开始日期格式无效")
		return
	}
	
	endDate, err := parseDate(endDateStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "结束日期格式无效")
		return
	}
	
	var merchantID *uint64
	if merchantIDStr != "" {
		id, err := strconv.ParseUint(merchantIDStr, 10, 64)
		if err != nil {
			utils.ErrorResponse(r, 400, "商户ID格式无效")
			return
		}
		merchantID = &id
	}
	
	data, err := c.analyticsService.GetFinancialData(ctx, startDate, endDate, merchantID)
	if err != nil {
		g.Log().Error(ctx, "获取财务分析数据失败", "error", err)
		utils.ErrorResponse(r, 500, "获取财务分析数据失败")
		return
	}
	
	utils.SuccessResponse(r, data)
}

// GetMerchantAnalytics 获取商户运营分析数据
// @Summary 获取商户运营分析数据
// @Description 获取指定时间范围的商户运营分析数据
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期" format(date)
// @Param end_date query string true "结束日期" format(date)
// @Success 200 {object} utils.Response{data=types.MerchantOperationReport}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/analytics/merchants [get]
func (c *ReportController) GetMerchantAnalytics(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	startDateStr := r.Get("start_date").String()
	endDateStr := r.Get("end_date").String()
	
	startDate, err := parseDate(startDateStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "开始日期格式无效")
		return
	}
	
	endDate, err := parseDate(endDateStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "结束日期格式无效")
		return
	}
	
	data, err := c.analyticsService.GetMerchantOperationData(ctx, startDate, endDate)
	if err != nil {
		g.Log().Error(ctx, "获取商户运营分析数据失败", "error", err)
		utils.ErrorResponse(r, 500, "获取商户运营分析数据失败")
		return
	}
	
	utils.SuccessResponse(r, data)
}

// GetCustomerAnalytics 获取客户分析数据
// @Summary 获取客户分析数据
// @Description 获取指定时间范围的客户分析数据
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期" format(date)
// @Param end_date query string true "结束日期" format(date)
// @Success 200 {object} utils.Response{data=types.CustomerAnalysisReport}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/analytics/customers [get]
func (c *ReportController) GetCustomerAnalytics(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	startDateStr := r.Get("start_date").String()
	endDateStr := r.Get("end_date").String()
	
	startDate, err := parseDate(startDateStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "开始日期格式无效")
		return
	}
	
	endDate, err := parseDate(endDateStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "结束日期格式无效")
		return
	}
	
	data, err := c.analyticsService.GetCustomerAnalysisData(ctx, startDate, endDate)
	if err != nil {
		g.Log().Error(ctx, "获取客户分析数据失败", "error", err)
		utils.ErrorResponse(r, 500, "获取客户分析数据失败")
		return
	}
	
	utils.SuccessResponse(r, data)
}

// CustomQuery 自定义数据查询
// @Summary 自定义数据查询
// @Description 执行自定义的数据分析查询
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param request body types.AnalyticsQueryRequest true "查询参数"
// @Success 200 {object} utils.Response{data=object}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/analytics/custom [post]
func (c *ReportController) CustomQuery(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	var req types.AnalyticsQueryRequest
	if err := r.Parse(&req); err != nil {
		utils.ErrorResponse(r, 400, "请求参数解析失败")
		return
	}
	
	data, err := c.analyticsService.CustomQuery(ctx, &req)
	if err != nil {
		g.Log().Error(ctx, "自定义查询失败", "metric_type", req.MetricType, "error", err)
		utils.ErrorResponse(r, 500, "自定义查询失败")
		return
	}
	
	utils.SuccessResponse(r, data)
}

// GetTrendData 获取趋势数据
// @Summary 获取趋势数据
// @Description 获取指定指标的趋势数据
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param metric path string true "指标类型"
// @Param start_date query string true "开始日期" format(date)
// @Param end_date query string true "结束日期" format(date)
// @Param group_by query string false "分组方式" Enums(day, week, month)
// @Param merchant_id query uint64 false "商户ID"
// @Success 200 {object} utils.Response{data=object}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/analytics/trends/{metric} [get]
func (c *ReportController) GetTrendData(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	metric := r.Get("metric").String()
	startDateStr := r.Get("start_date").String()
	endDateStr := r.Get("end_date").String()
	groupBy := r.Get("group_by", "day").String()
	merchantIDStr := r.Get("merchant_id").String()
	
	if metric == "" {
		utils.ErrorResponse(r, 400, "指标类型不能为空")
		return
	}
	
	startDate, err := parseDate(startDateStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "开始日期格式无效")
		return
	}
	
	endDate, err := parseDate(endDateStr)
	if err != nil {
		utils.ErrorResponse(r, 400, "结束日期格式无效")
		return
	}
	
	var merchantID *uint64
	if merchantIDStr != "" {
		id, err := strconv.ParseUint(merchantIDStr, 10, 64)
		if err != nil {
			utils.ErrorResponse(r, 400, "商户ID格式无效")
			return
		}
		merchantID = &id
	}
	
	// 构建自定义查询请求
	req := &types.AnalyticsQueryRequest{
		MetricType: metric,
		StartDate:  startDate,
		EndDate:    endDate,
		GroupBy:    groupBy,
		MerchantID: merchantID,
	}
	
	data, err := c.analyticsService.CustomQuery(ctx, req)
	if err != nil {
		g.Log().Error(ctx, "获取趋势数据失败", "metric", metric, "error", err)
		utils.ErrorResponse(r, 500, "获取趋势数据失败")
		return
	}
	
	utils.SuccessResponse(r, data)
}

// ClearCache 清理分析数据缓存
// @Summary 清理分析数据缓存
// @Description 清理分析数据缓存，可指定缓存模式
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param pattern query string false "缓存模式"
// @Success 200 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/analytics/cache/clear [post]
func (c *ReportController) ClearCache(r *ghttp.Request) {
	ctx := r.GetCtx()
	
	pattern := r.Get("pattern", "*").String()
	
	err := c.analyticsService.ClearCache(ctx, pattern)
	if err != nil {
		g.Log().Error(ctx, "清理缓存失败", "pattern", pattern, "error", err)
		utils.ErrorResponse(r, 500, "清理缓存失败")
		return
	}
	
	utils.SuccessResponse(r, g.Map{
		"message": "缓存清理成功",
	})
}

// parseDate 解析日期字符串
func parseDate(dateStr string) (time.Time, error) {
	// 支持多种日期格式
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"01/02/2006",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("无法解析日期: %s", dateStr)
}
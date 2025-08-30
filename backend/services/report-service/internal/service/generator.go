package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/xuri/excelize/v2"
)

// IReportGeneratorService 报表生成服务接口
type IReportGeneratorService interface {
	GenerateReport(ctx context.Context, req *types.ReportCreateRequest) (*types.Report, error)
	GetReport(ctx context.Context, reportID uint64) (*types.Report, error)
	GetReportByUUID(ctx context.Context, uuid string) (*types.Report, error)
	ListReports(ctx context.Context, req *types.ReportListRequest) ([]*types.Report, int, error)
	DeleteReport(ctx context.Context, reportID uint64) error
	DownloadReport(ctx context.Context, reportUUID string) (string, error)
	CleanupCache(ctx context.Context) error
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)
	WarmupCache(ctx context.Context, reportType types.ReportType) error
}

// ReportGeneratorService 报表生成服务实现
type ReportGeneratorService struct {
	reportRepo      repository.IReportRepository
	analyticsService IAnalyticsService
	templateEngine   ITemplateEngine
	pdfGenerator     IPDFGenerator
	cacheManager     ICacheManager
}

// NewReportGeneratorService 创建报表生成服务实例
func NewReportGeneratorService() IReportGeneratorService {
	return &ReportGeneratorService{
		reportRepo:      repository.NewReportRepository(),
		analyticsService: NewAnalyticsService(),
		templateEngine:   NewTemplateEngine(),
		pdfGenerator:     NewPDFGenerator(),
		cacheManager:     NewCacheManager(),
	}
}

// GenerateReport 生成报表
func (s *ReportGeneratorService) GenerateReport(ctx context.Context, req *types.ReportCreateRequest) (*types.Report, error) {
	tenantID := ctx.Value("tenant_id").(uint64)
	userID := ctx.Value("user_id").(uint64)
	
	g.Log().Info(ctx, "开始生成报表", 
		"report_type", req.ReportType,
		"period_type", req.PeriodType,
		"start_date", req.StartDate,
		"end_date", req.EndDate,
		"file_format", req.FileFormat)
	
	// 验证请求参数
	if err := s.validateGenerateRequest(req); err != nil {
		return nil, fmt.Errorf("参数验证失败: %v", err)
	}
	
	// 检查缓存中是否已有相同的报表
	if s.cacheManager.ShouldUseCache(req) {
		if cachedReport, err := s.cacheManager.GetReportFromCache(ctx, req); err == nil {
			g.Log().Info(ctx, "从缓存返回报表", "report_id", cachedReport.ID)
			return cachedReport, nil
		} else {
			g.Log().Debug(ctx, "缓存中未找到报表，将生成新报表", "error", err)
		}
	}
	
	// 创建报表记录
	report := &types.Report{
		UUID:        generateUUID(),
		TenantID:    tenantID,
		ReportType:  req.ReportType,
		PeriodType:  req.PeriodType,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      types.ReportStatusGenerating,
		FileFormat:  req.FileFormat,
		GeneratedBy: userID,
		ExpiresAt:   timePtr(time.Now().Add(30 * 24 * time.Hour)), // 30天后过期
	}
	
	err := s.reportRepo.CreateReport(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("创建报表记录失败: %v", err)
	}
	
	// 异步生成报表
	go s.generateReportAsync(context.Background(), report, req)
	
	return report, nil
}

// GetReport 获取报表信息
func (s *ReportGeneratorService) GetReport(ctx context.Context, reportID uint64) (*types.Report, error) {
	return s.reportRepo.GetReportByID(ctx, reportID)
}

// GetReportByUUID 根据UUID获取报表
func (s *ReportGeneratorService) GetReportByUUID(ctx context.Context, uuid string) (*types.Report, error) {
	return s.reportRepo.GetReportByUUID(ctx, uuid)
}

// ListReports 获取报表列表
func (s *ReportGeneratorService) ListReports(ctx context.Context, req *types.ReportListRequest) ([]*types.Report, int, error) {
	return s.reportRepo.ListReports(ctx, req)
}

// DeleteReport 删除报表
func (s *ReportGeneratorService) DeleteReport(ctx context.Context, reportID uint64) error {
	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("报表不存在: %v", err)
	}
	
	// 删除文件
	if report.FilePath != "" {
		if err := os.Remove(report.FilePath); err != nil {
			g.Log().Warning(ctx, "删除报表文件失败", "file_path", report.FilePath, "error", err)
		}
	}
	
	// 删除数据库记录
	return s.reportRepo.DeleteReport(ctx, reportID)
}

// DownloadReport 下载报表文件
func (s *ReportGeneratorService) DownloadReport(ctx context.Context, reportUUID string) (string, error) {
	report, err := s.reportRepo.GetReportByUUID(ctx, reportUUID)
	if err != nil {
		return "", fmt.Errorf("报表不存在: %v", err)
	}
	
	if report.Status != types.ReportStatusCompleted {
		return "", fmt.Errorf("报表尚未生成完成，当前状态: %s", report.Status)
	}
	
	if report.FilePath == "" {
		return "", fmt.Errorf("报表文件路径为空")
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(report.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("报表文件不存在")
	}
	
	return report.FilePath, nil
}

// validateGenerateRequest 验证报表生成请求
func (s *ReportGeneratorService) validateGenerateRequest(req *types.ReportCreateRequest) error {
	if req.StartDate.After(req.EndDate) {
		return fmt.Errorf("开始日期不能晚于结束日期")
	}
	
	if req.EndDate.After(time.Now()) {
		return fmt.Errorf("结束日期不能晚于当前时间")
	}
	
	// 检查时间范围是否合理
	duration := req.EndDate.Sub(req.StartDate)
	maxDuration := 365 * 24 * time.Hour // 最大1年
	
	if duration > maxDuration {
		return fmt.Errorf("时间范围不能超过1年")
	}
	
	return nil
}

// generateReportAsync 异步生成报表
func (s *ReportGeneratorService) generateReportAsync(ctx context.Context, report *types.Report, req *types.ReportCreateRequest) {
	ctx = context.WithValue(ctx, "tenant_id", report.TenantID)
	
	g.Log().Info(ctx, "开始异步生成报表", "report_id", report.ID, "report_uuid", report.UUID)
	
	// 更新状态为生成中
	report.Status = types.ReportStatusGenerating
	s.reportRepo.UpdateReport(ctx, report)
	
	var err error
	var filePath string
	var dataSummary json.RawMessage
	
	defer func() {
		if err != nil {
			g.Log().Error(ctx, "报表生成失败", "report_id", report.ID, "error", err)
			report.Status = types.ReportStatusFailed
		} else {
			g.Log().Info(ctx, "报表生成成功", "report_id", report.ID, "file_path", filePath)
			report.Status = types.ReportStatusCompleted
			report.FilePath = filePath
			report.DataSummary = dataSummary
			
			// 缓存生成的报表
			if s.cacheManager.ShouldUseCache(req) {
				if cacheErr := s.cacheManager.CacheReport(ctx, req, report); cacheErr != nil {
					g.Log().Warning(ctx, "缓存报表失败", "report_id", report.ID, "error", cacheErr)
				} else {
					g.Log().Debug(ctx, "报表缓存成功", "report_id", report.ID)
				}
			}
		}
		
		s.reportRepo.UpdateReport(ctx, report)
	}()
	
	// 获取数据
	var data interface{}
	switch req.ReportType {
	case types.ReportTypeFinancial:
		data, err = s.analyticsService.GetFinancialData(ctx, req.StartDate, req.EndDate, req.MerchantID)
		if err != nil {
			return
		}
		
	case types.ReportTypeMerchantOperation:
		data, err = s.analyticsService.GetMerchantOperationData(ctx, req.StartDate, req.EndDate)
		if err != nil {
			return
		}
		
	case types.ReportTypeCustomerAnalysis:
		data, err = s.analyticsService.GetCustomerAnalysisData(ctx, req.StartDate, req.EndDate)
		if err != nil {
			return
		}
		
	default:
		err = fmt.Errorf("不支持的报表类型: %s", req.ReportType)
		return
	}
	
	// 生成数据摘要
	if dataSummary, err = json.Marshal(s.generateDataSummary(data)); err != nil {
		g.Log().Warning(ctx, "生成数据摘要失败", "error", err)
	}
	
	// 根据文件格式生成文件
	switch req.FileFormat {
	case types.FileFormatExcel:
		filePath, err = s.generateExcelReport(ctx, report, data)
	case types.FileFormatPDF:
		filePath, err = s.generatePDFReport(ctx, report, data)
	case types.FileFormatJSON:
		filePath, err = s.generateJSONReport(ctx, report, data)
	default:
		err = fmt.Errorf("不支持的文件格式: %s", req.FileFormat)
	}
}

// generateExcelReport 生成Excel报表
func (s *ReportGeneratorService) generateExcelReport(ctx context.Context, report *types.Report, data interface{}) (string, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			g.Log().Warning(ctx, "关闭Excel文件失败", "error", err)
		}
	}()
	
	g.Log().Info(ctx, "开始生成Excel报表", 
		"report_type", report.ReportType,
		"report_uuid", report.UUID)
	
	// 删除默认的Sheet1工作表
	f.DeleteSheet("Sheet1")
	
	// 根据报表类型创建不同的工作表结构
	switch report.ReportType {
	case types.ReportTypeFinancial:
		err := s.createFinancialExcelSheets(f, data.(*types.FinancialReportData))
		if err != nil {
			return "", fmt.Errorf("创建财务报表Excel失败: %v", err)
		}
	case types.ReportTypeMerchantOperation:
		err := s.createMerchantOperationExcelSheets(f, data.(*types.MerchantOperationReport))
		if err != nil {
			return "", fmt.Errorf("创建商户运营报表Excel失败: %v", err)
		}
	case types.ReportTypeCustomerAnalysis:
		err := s.createCustomerAnalysisExcelSheets(f, data.(*types.CustomerAnalysisReport))
		if err != nil {
			return "", fmt.Errorf("创建客户分析报表Excel失败: %v", err)
		}
	}
	
	// 应用Excel样式
	if err := s.applyExcelStyles(f); err != nil {
		g.Log().Warning(ctx, "应用Excel样式失败", "error", err)
	}
	
	// 确保报表目录存在
	reportDir := s.getReportDir()
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", fmt.Errorf("创建报表目录失败: %v", err)
	}
	
	// 生成文件路径
	filename := fmt.Sprintf("%s_%s_%s.xlsx", 
		report.ReportType, 
		report.UUID,
		time.Now().Format("20060102_150405"))
	filePath := filepath.Join(reportDir, filename)
	
	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return "", fmt.Errorf("保存Excel文件失败: %v", err)
	}
	
	g.Log().Info(ctx, "Excel报表生成成功", "file_path", filePath)
	return filePath, nil
}

// generatePDFReport 生成PDF报表
func (s *ReportGeneratorService) generatePDFReport(ctx context.Context, report *types.Report, data interface{}) (string, error) {
	return s.pdfGenerator.GeneratePDFReport(ctx, report, data)
}

// generateJSONReport 生成JSON报表
func (s *ReportGeneratorService) generateJSONReport(ctx context.Context, report *types.Report, data interface{}) (string, error) {
	// 确保报表目录存在
	reportDir := s.getReportDir()
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", fmt.Errorf("创建报表目录失败: %v", err)
	}
	
	// 生成文件路径
	filename := fmt.Sprintf("%s_%s_%s.json", 
		report.ReportType, 
		report.UUID,
		time.Now().Format("20060102_150405"))
	filePath := filepath.Join(reportDir, filename)
	
	// 序列化数据
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化JSON数据失败: %v", err)
	}
	
	// 写入文件
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("写入JSON文件失败: %v", err)
	}
	
	return filePath, nil
}

// fillFinancialDataToExcel 填充财务数据到Excel
func (s *ReportGeneratorService) fillFinancialDataToExcel(f *excelize.File, sheetName string, data *types.FinancialReportData) error {
	// 设置标题
	f.SetCellValue(sheetName, "A1", "财务报表")
	f.SetCellValue(sheetName, "A2", "")
	
	// 基础数据
	f.SetCellValue(sheetName, "A3", "总收入")
	f.SetCellValue(sheetName, "B3", data.TotalRevenue.Amount)
	f.SetCellValue(sheetName, "A4", "总支出")
	f.SetCellValue(sheetName, "B4", data.TotalExpenditure.Amount)
	f.SetCellValue(sheetName, "A5", "净利润")
	f.SetCellValue(sheetName, "B5", data.NetProfit.Amount)
	f.SetCellValue(sheetName, "A6", "订单总数")
	f.SetCellValue(sheetName, "B6", data.OrderCount)
	f.SetCellValue(sheetName, "A7", "商户总数")
	f.SetCellValue(sheetName, "B7", data.MerchantCount)
	f.SetCellValue(sheetName, "A8", "客户总数")
	f.SetCellValue(sheetName, "B8", data.CustomerCount)
	
	// 如果有分解数据，添加更多工作表
	if data.Breakdown != nil {
		// 商户收入排行
		if len(data.Breakdown.RevenueByMerchant) > 0 {
			merchantSheet := "商户收入排行"
			f.NewSheet(merchantSheet)
			
			f.SetCellValue(merchantSheet, "A1", "商户ID")
			f.SetCellValue(merchantSheet, "B1", "商户名称")
			f.SetCellValue(merchantSheet, "C1", "收入金额")
			f.SetCellValue(merchantSheet, "D1", "订单数量")
			f.SetCellValue(merchantSheet, "E1", "占比(%)")
			
			for i, merchant := range data.Breakdown.RevenueByMerchant {
				row := i + 2
				f.SetCellValue(merchantSheet, fmt.Sprintf("A%d", row), merchant.MerchantID)
				f.SetCellValue(merchantSheet, fmt.Sprintf("B%d", row), merchant.MerchantName)
				f.SetCellValue(merchantSheet, fmt.Sprintf("C%d", row), merchant.Revenue.Amount)
				f.SetCellValue(merchantSheet, fmt.Sprintf("D%d", row), merchant.OrderCount)
				f.SetCellValue(merchantSheet, fmt.Sprintf("E%d", row), merchant.Percentage)
			}
		}
		
		// 月度趋势
		if len(data.Breakdown.MonthlyTrend) > 0 {
			trendSheet := "月度趋势"
			f.NewSheet(trendSheet)
			
			f.SetCellValue(trendSheet, "A1", "月份")
			f.SetCellValue(trendSheet, "B1", "收入")
			f.SetCellValue(trendSheet, "C1", "支出")
			f.SetCellValue(trendSheet, "D1", "净利润")
			f.SetCellValue(trendSheet, "E1", "订单数量")
			
			for i, trend := range data.Breakdown.MonthlyTrend {
				row := i + 2
				f.SetCellValue(trendSheet, fmt.Sprintf("A%d", row), trend.Month)
				f.SetCellValue(trendSheet, fmt.Sprintf("B%d", row), trend.Revenue.Amount)
				f.SetCellValue(trendSheet, fmt.Sprintf("C%d", row), trend.Expenditure.Amount)
				f.SetCellValue(trendSheet, fmt.Sprintf("D%d", row), trend.NetProfit.Amount)
				f.SetCellValue(trendSheet, fmt.Sprintf("E%d", row), trend.OrderCount)
			}
		}
	}
	
	return nil
}

// fillMerchantOperationDataToExcel 填充商户运营数据到Excel
func (s *ReportGeneratorService) fillMerchantOperationDataToExcel(f *excelize.File, sheetName string, data *types.MerchantOperationReport) error {
	// 设置标题
	f.SetCellValue(sheetName, "A1", "商户运营报表")
	f.SetCellValue(sheetName, "A2", "")
	
	// 商户排名
	if len(data.MerchantRankings) > 0 {
		f.SetCellValue(sheetName, "A3", "排名")
		f.SetCellValue(sheetName, "B3", "商户名称")
		f.SetCellValue(sheetName, "C3", "总收入")
		f.SetCellValue(sheetName, "D3", "订单数量")
		f.SetCellValue(sheetName, "E3", "客户数量")
		f.SetCellValue(sheetName, "F3", "客单价")
		f.SetCellValue(sheetName, "G3", "增长率(%)")
		
		for i, ranking := range data.MerchantRankings {
			row := i + 4
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), ranking.Rank)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), ranking.MerchantName)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), ranking.TotalRevenue.Amount)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), ranking.OrderCount)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), ranking.CustomerCount)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), ranking.AverageOrderValue.Amount)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), ranking.GrowthRate)
		}
	}
	
	return nil
}

// fillCustomerAnalysisDataToExcel 填充客户分析数据到Excel
func (s *ReportGeneratorService) fillCustomerAnalysisDataToExcel(f *excelize.File, sheetName string, data *types.CustomerAnalysisReport) error {
	// 设置标题
	f.SetCellValue(sheetName, "A1", "客户分析报表")
	f.SetCellValue(sheetName, "A2", "")
	
	// 用户增长数据
	if len(data.UserGrowth) > 0 {
		f.SetCellValue(sheetName, "A3", "月份")
		f.SetCellValue(sheetName, "B3", "新增用户")
		f.SetCellValue(sheetName, "C3", "活跃用户")
		f.SetCellValue(sheetName, "D3", "累计用户")
		f.SetCellValue(sheetName, "E3", "留存率(%)")
		
		for i, growth := range data.UserGrowth {
			row := i + 4
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), growth.Month)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), growth.NewUsers)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), growth.ActiveUsers)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), growth.CumulativeUsers)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), growth.RetentionRate)
		}
	}
	
	// 活跃度指标
	if data.ActivityMetrics != nil {
		startRow := len(data.UserGrowth) + 6
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", startRow), "活跃度指标")
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", startRow+1), "日活跃用户(DAU)")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", startRow+1), data.ActivityMetrics.DAU)
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", startRow+2), "周活跃用户(WAU)")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", startRow+2), data.ActivityMetrics.WAU)
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", startRow+3), "月活跃用户(MAU)")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", startRow+3), data.ActivityMetrics.MAU)
	}
	
	return nil
}

// generateDataSummary 生成数据摘要
func (s *ReportGeneratorService) generateDataSummary(data interface{}) map[string]interface{} {
	summary := map[string]interface{}{}
	
	switch d := data.(type) {
	case *types.FinancialReportData:
		summary["type"] = "financial"
		summary["total_revenue"] = d.TotalRevenue.Amount
		summary["order_count"] = d.OrderCount
		summary["merchant_count"] = d.MerchantCount
		summary["customer_count"] = d.CustomerCount
		
	case *types.MerchantOperationReport:
		summary["type"] = "merchant_operation"
		summary["merchant_count"] = len(d.MerchantRankings)
		summary["category_count"] = len(d.CategoryAnalysis)
		
	case *types.CustomerAnalysisReport:
		summary["type"] = "customer_analysis"
		summary["growth_periods"] = len(d.UserGrowth)
		if d.ActivityMetrics != nil {
			summary["mau"] = d.ActivityMetrics.MAU
		}
	}
	
	summary["generated_at"] = time.Now()
	return summary
}

// getReportDir 获取报表存储目录
func (s *ReportGeneratorService) getReportDir() string {
	baseDir := g.Cfg().MustGet(context.Background(), "report.storage_dir", "/tmp/reports").String()
	return baseDir
}

// timePtr 返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}

// generateUUID 生成UUID（简化实现）
func generateUUID() string {
	return fmt.Sprintf("rpt_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}

// createFinancialExcelSheets 创建财务报表Excel工作表
func (s *ReportGeneratorService) createFinancialExcelSheets(f *excelize.File, data *types.FinancialReportData) error {
	// 创建概览工作表
	overviewSheet := "财务概览"
	index, err := f.NewSheet(overviewSheet)
	if err != nil {
		return fmt.Errorf("创建概览工作表失败: %v", err)
	}
	f.SetActiveSheet(index)
	
	// 设置概览数据
	f.SetCellValue(overviewSheet, "A1", "财务报表概览")
	f.SetCellValue(overviewSheet, "A3", "指标")
	f.SetCellValue(overviewSheet, "B3", "数值")
	f.SetCellValue(overviewSheet, "C3", "单位")
	
	// 填充财务指标
	f.SetCellValue(overviewSheet, "A4", "总收入")
	f.SetCellValue(overviewSheet, "B4", data.TotalRevenue.Amount)
	f.SetCellValue(overviewSheet, "C4", "元")
	
	f.SetCellValue(overviewSheet, "A5", "净利润")
	f.SetCellValue(overviewSheet, "B5", data.NetProfit.Amount)
	f.SetCellValue(overviewSheet, "C5", "元")
	
	f.SetCellValue(overviewSheet, "A6", "订单总数")
	f.SetCellValue(overviewSheet, "B6", data.OrderCount)
	f.SetCellValue(overviewSheet, "C6", "笔")
	
	f.SetCellValue(overviewSheet, "A7", "商户总数")
	f.SetCellValue(overviewSheet, "B7", data.MerchantCount)
	f.SetCellValue(overviewSheet, "C7", "个")
	
	f.SetCellValue(overviewSheet, "A8", "客户总数")
	f.SetCellValue(overviewSheet, "B8", data.CustomerCount)
	f.SetCellValue(overviewSheet, "C8", "个")
	
	f.SetCellValue(overviewSheet, "A9", "权益消耗")
	f.SetCellValue(overviewSheet, "B9", data.RightsConsumed)
	f.SetCellValue(overviewSheet, "C9", "份")
	
	// 创建商户收入排行工作表
	if data.Breakdown != nil && len(data.Breakdown.RevenueByMerchant) > 0 {
		merchantSheet := "商户收入排行"
		f.NewSheet(merchantSheet)
		
		f.SetCellValue(merchantSheet, "A1", "商户收入排行榜")
		f.SetCellValue(merchantSheet, "A3", "排名")
		f.SetCellValue(merchantSheet, "B3", "商户ID")
		f.SetCellValue(merchantSheet, "C3", "商户名称")
		f.SetCellValue(merchantSheet, "D3", "收入金额")
		f.SetCellValue(merchantSheet, "E3", "订单数量")
		f.SetCellValue(merchantSheet, "F3", "占比(%)")
		
		for i, merchant := range data.Breakdown.RevenueByMerchant {
			row := i + 4
			f.SetCellValue(merchantSheet, fmt.Sprintf("A%d", row), i+1)
			f.SetCellValue(merchantSheet, fmt.Sprintf("B%d", row), merchant.MerchantID)
			f.SetCellValue(merchantSheet, fmt.Sprintf("C%d", row), merchant.MerchantName)
			f.SetCellValue(merchantSheet, fmt.Sprintf("D%d", row), merchant.Revenue.Amount)
			f.SetCellValue(merchantSheet, fmt.Sprintf("E%d", row), merchant.OrderCount)
			f.SetCellValue(merchantSheet, fmt.Sprintf("F%d", row), merchant.Percentage)
		}
	}
	
	// 创建月度趋势工作表
	if data.Breakdown != nil && len(data.Breakdown.MonthlyTrend) > 0 {
		trendSheet := "月度趋势"
		f.NewSheet(trendSheet)
		
		f.SetCellValue(trendSheet, "A1", "月度财务趋势")
		f.SetCellValue(trendSheet, "A3", "月份")
		f.SetCellValue(trendSheet, "B3", "收入")
		f.SetCellValue(trendSheet, "C3", "支出")
		f.SetCellValue(trendSheet, "D3", "净利润")
		f.SetCellValue(trendSheet, "E3", "订单数量")
		f.SetCellValue(trendSheet, "F3", "权益消耗")
		
		for i, trend := range data.Breakdown.MonthlyTrend {
			row := i + 4
			f.SetCellValue(trendSheet, fmt.Sprintf("A%d", row), trend.Month)
			f.SetCellValue(trendSheet, fmt.Sprintf("B%d", row), trend.Revenue.Amount)
			f.SetCellValue(trendSheet, fmt.Sprintf("C%d", row), trend.Expenditure.Amount)
			f.SetCellValue(trendSheet, fmt.Sprintf("D%d", row), trend.NetProfit.Amount)
			f.SetCellValue(trendSheet, fmt.Sprintf("E%d", row), trend.OrderCount)
			f.SetCellValue(trendSheet, fmt.Sprintf("F%d", row), trend.RightsConsumed)
		}
	}
	
	return nil
}

// createMerchantOperationExcelSheets 创建商户运营报表Excel工作表
func (s *ReportGeneratorService) createMerchantOperationExcelSheets(f *excelize.File, data *types.MerchantOperationReport) error {
	// 商户排行榜工作表
	if len(data.MerchantRankings) > 0 {
		rankingSheet := "商户排行榜"
		index, err := f.NewSheet(rankingSheet)
		if err != nil {
			return fmt.Errorf("创建商户排行榜工作表失败: %v", err)
		}
		f.SetActiveSheet(index)
		
		f.SetCellValue(rankingSheet, "A1", "商户业绩排行榜")
		f.SetCellValue(rankingSheet, "A3", "排名")
		f.SetCellValue(rankingSheet, "B3", "商户名称")
		f.SetCellValue(rankingSheet, "C3", "总收入")
		f.SetCellValue(rankingSheet, "D3", "订单数量")
		f.SetCellValue(rankingSheet, "E3", "客户数量")
		f.SetCellValue(rankingSheet, "F3", "客单价")
		f.SetCellValue(rankingSheet, "G3", "增长率(%)")
		
		for i, ranking := range data.MerchantRankings {
			row := i + 4
			f.SetCellValue(rankingSheet, fmt.Sprintf("A%d", row), ranking.Rank)
			f.SetCellValue(rankingSheet, fmt.Sprintf("B%d", row), ranking.MerchantName)
			f.SetCellValue(rankingSheet, fmt.Sprintf("C%d", row), ranking.TotalRevenue.Amount)
			f.SetCellValue(rankingSheet, fmt.Sprintf("D%d", row), ranking.OrderCount)
			f.SetCellValue(rankingSheet, fmt.Sprintf("E%d", row), ranking.CustomerCount)
			f.SetCellValue(rankingSheet, fmt.Sprintf("F%d", row), ranking.AverageOrderValue.Amount)
			f.SetCellValue(rankingSheet, fmt.Sprintf("G%d", row), ranking.GrowthRate)
		}
	}
	
	// 类别分析工作表
	if len(data.CategoryAnalysis) > 0 {
		categorySheet := "类别分析"
		f.NewSheet(categorySheet)
		
		f.SetCellValue(categorySheet, "A1", "商品类别分析")
		f.SetCellValue(categorySheet, "A3", "类别名称")
		f.SetCellValue(categorySheet, "B3", "收入")
		f.SetCellValue(categorySheet, "C3", "订单数量")
		f.SetCellValue(categorySheet, "D3", "商户数量")
		f.SetCellValue(categorySheet, "E3", "市场份额(%)")
		f.SetCellValue(categorySheet, "F3", "增长率(%)")
		
		for i, category := range data.CategoryAnalysis {
			row := i + 4
			f.SetCellValue(categorySheet, fmt.Sprintf("A%d", row), category.CategoryName)
			f.SetCellValue(categorySheet, fmt.Sprintf("B%d", row), category.Revenue.Amount)
			f.SetCellValue(categorySheet, fmt.Sprintf("C%d", row), category.OrderCount)
			f.SetCellValue(categorySheet, fmt.Sprintf("D%d", row), category.MerchantCount)
			f.SetCellValue(categorySheet, fmt.Sprintf("E%d", row), category.MarketShare)
			f.SetCellValue(categorySheet, fmt.Sprintf("F%d", row), category.GrowthRate)
		}
	}
	
	return nil
}

// createCustomerAnalysisExcelSheets 创建客户分析报表Excel工作表
func (s *ReportGeneratorService) createCustomerAnalysisExcelSheets(f *excelize.File, data *types.CustomerAnalysisReport) error {
	// 用户增长工作表
	if len(data.UserGrowth) > 0 {
		growthSheet := "用户增长"
		index, err := f.NewSheet(growthSheet)
		if err != nil {
			return fmt.Errorf("创建用户增长工作表失败: %v", err)
		}
		f.SetActiveSheet(index)
		
		f.SetCellValue(growthSheet, "A1", "用户增长分析")
		f.SetCellValue(growthSheet, "A3", "月份")
		f.SetCellValue(growthSheet, "B3", "新增用户")
		f.SetCellValue(growthSheet, "C3", "活跃用户")
		f.SetCellValue(growthSheet, "D3", "累计用户")
		f.SetCellValue(growthSheet, "E3", "留存率(%)")
		
		for i, growth := range data.UserGrowth {
			row := i + 4
			f.SetCellValue(growthSheet, fmt.Sprintf("A%d", row), growth.Month)
			f.SetCellValue(growthSheet, fmt.Sprintf("B%d", row), growth.NewUsers)
			f.SetCellValue(growthSheet, fmt.Sprintf("C%d", row), growth.ActiveUsers)
			f.SetCellValue(growthSheet, fmt.Sprintf("D%d", row), growth.CumulativeUsers)
			f.SetCellValue(growthSheet, fmt.Sprintf("E%d", row), growth.RetentionRate)
		}
	}
	
	// 活跃度指标工作表
	if data.ActivityMetrics != nil {
		activitySheet := "活跃度指标"
		f.NewSheet(activitySheet)
		
		f.SetCellValue(activitySheet, "A1", "用户活跃度指标")
		f.SetCellValue(activitySheet, "A3", "指标")
		f.SetCellValue(activitySheet, "B3", "数值")
		
		f.SetCellValue(activitySheet, "A4", "日活跃用户(DAU)")
		f.SetCellValue(activitySheet, "B4", data.ActivityMetrics.DAU)
		
		f.SetCellValue(activitySheet, "A5", "周活跃用户(WAU)")
		f.SetCellValue(activitySheet, "B5", data.ActivityMetrics.WAU)
		
		f.SetCellValue(activitySheet, "A6", "月活跃用户(MAU)")
		f.SetCellValue(activitySheet, "B6", data.ActivityMetrics.MAU)
		
		f.SetCellValue(activitySheet, "A7", "平均会话时长(分钟)")
		f.SetCellValue(activitySheet, "B7", data.ActivityMetrics.AverageSessionTime)
		
		f.SetCellValue(activitySheet, "A8", "平均下单频次")
		f.SetCellValue(activitySheet, "B8", data.ActivityMetrics.AverageOrderFreq)
	}
	
	return nil
}

// applyExcelStyles 应用Excel样式
func (s *ReportGeneratorService) applyExcelStyles(f *excelize.File) error {
	// 创建标题样式
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return fmt.Errorf("创建标题样式失败: %v", err)
	}
	
	// 创建表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("创建表头样式失败: %v", err)
	}
	
	// 创建数据样式
	dataStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("创建数据样式失败: %v", err)
	}
	
	// 应用样式到所有工作表
	sheets := f.GetSheetList()
	for _, sheetName := range sheets {
		// 应用标题样式
		f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
		
		// 合并标题单元格
		f.MergeCell(sheetName, "A1", "G1")
		
		// 应用表头样式
		f.SetCellStyle(sheetName, "A3", "G3", headerStyle)
		
		// 获取数据范围并应用数据样式
		rows, _ := f.GetRows(sheetName)
		if len(rows) > 3 {
			endRow := len(rows)
			f.SetCellStyle(sheetName, "A4", fmt.Sprintf("G%d", endRow), dataStyle)
		}
		
		// 自动调整列宽
		f.SetColWidth(sheetName, "A", "G", 15)
	}
	
	return nil
}

// CleanupCache 清理过期缓存
func (s *ReportGeneratorService) CleanupCache(ctx context.Context) error {
	g.Log().Info(ctx, "开始清理报表缓存")
	
	err := s.cacheManager.CleanupExpiredCache(ctx)
	if err != nil {
		g.Log().Error(ctx, "清理缓存失败", "error", err)
		return fmt.Errorf("清理缓存失败: %v", err)
	}
	
	g.Log().Info(ctx, "缓存清理完成")
	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *ReportGeneratorService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := s.cacheManager.GetCacheStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取缓存统计失败: %v", err)
	}
	
	// 添加报表相关的统计信息
	reportStats, err := s.getReportStats(ctx)
	if err != nil {
		g.Log().Warning(ctx, "获取报表统计失败", "error", err)
	} else {
		stats["report_stats"] = reportStats
	}
	
	return stats, nil
}

// WarmupCache 预热缓存
func (s *ReportGeneratorService) WarmupCache(ctx context.Context, reportType types.ReportType) error {
	g.Log().Info(ctx, "开始预热报表缓存", "report_type", reportType)
	
	err := s.cacheManager.WarmupCache(ctx, reportType)
	if err != nil {
		g.Log().Error(ctx, "缓存预热失败", "report_type", reportType, "error", err)
		return fmt.Errorf("缓存预热失败: %v", err)
	}
	
	g.Log().Info(ctx, "缓存预热完成", "report_type", reportType)
	return nil
}

// getReportStats 获取报表统计信息（内部方法）
func (s *ReportGeneratorService) getReportStats(ctx context.Context) (map[string]interface{}, error) {
	// 查询最近7天的报表生成统计
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	
	// 这里可以扩展更详细的统计查询
	// 目前返回基本的统计信息
	stats := map[string]interface{}{
		"period": "last_7_days",
		"start_date": sevenDaysAgo.Format("2006-01-02"),
		"end_date": now.Format("2006-01-02"),
		// 可以添加更多统计信息，如生成成功率、平均生成时间等
	}
	
	return stats, nil
}
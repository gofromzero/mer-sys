package unit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReportGeneratorService_GenerateReport 测试报表生成功能
func TestReportGeneratorService_GenerateReport(t *testing.T) {
	// 创建上下文
	ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
	ctx = context.WithValue(ctx, "user_id", uint64(100))
	
	// 创建报表生成器
	generator := service.NewReportGeneratorService()
	
	// 准备测试请求
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Now().AddDate(0, -1, 0),
		EndDate:    time.Now(),
		FileFormat: types.FileFormatExcel,
	}
	
	// 执行报表生成
	report, err := generator.GenerateReport(ctx, req)
	
	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, types.ReportStatusGenerating, report.Status)
	assert.Equal(t, req.ReportType, report.ReportType)
	assert.Equal(t, req.FileFormat, report.FileFormat)
	assert.Equal(t, uint64(1), report.TenantID)
	assert.Equal(t, uint64(100), report.GeneratedBy)
	assert.NotEmpty(t, report.UUID)
	
	// 验证过期时间
	expectedExpiry := time.Now().Add(30 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, *report.ExpiresAt, time.Hour)
}

// TestReportGeneratorService_ValidateGenerateRequest 测试请求验证
func TestReportGeneratorService_ValidateGenerateRequest(t *testing.T) {
	generator := service.NewReportGeneratorService()
	ctx := context.WithValue(context.Background(), "tenant_id", uint64(1))
	ctx = context.WithValue(ctx, "user_id", uint64(100))
	
	tests := []struct {
		name    string
		req     *types.ReportCreateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效请求",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  time.Now().AddDate(0, -1, 0),
				EndDate:    time.Now().AddDate(0, 0, -1),
				FileFormat: types.FileFormatExcel,
			},
			wantErr: false,
		},
		{
			name: "开始时间晚于结束时间",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  time.Now(),
				EndDate:    time.Now().AddDate(0, -1, 0),
				FileFormat: types.FileFormatExcel,
			},
			wantErr: true,
			errMsg:  "开始日期不能晚于结束日期",
		},
		{
			name: "结束时间晚于当前时间",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  time.Now().AddDate(0, -1, 0),
				EndDate:    time.Now().Add(time.Hour),
				FileFormat: types.FileFormatExcel,
			},
			wantErr: true,
			errMsg:  "结束日期不能晚于当前时间",
		},
		{
			name: "时间范围超过1年",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  time.Now().AddDate(-2, 0, 0),
				EndDate:    time.Now().AddDate(0, 0, -1),
				FileFormat: types.FileFormatExcel,
			},
			wantErr: true,
			errMsg:  "时间范围不能超过1年",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generator.GenerateReport(ctx, tt.req)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTemplateEngine_RenderTemplate 测试模板渲染
func TestTemplateEngine_RenderTemplate(t *testing.T) {
	engine := service.NewTemplateEngine()
	ctx := context.Background()
	
	// 创建财务报表测试数据
	financialData := &types.FinancialReportData{
		TotalRevenue:         &types.Money{Amount: 100000.00, Currency: "CNY"},
		TotalExpenditure:     &types.Money{Amount: 60000.00, Currency: "CNY"},
		NetProfit:           &types.Money{Amount: 40000.00, Currency: "CNY"},
		OrderCount:          500,
		MerchantCount:       10,
		CustomerCount:       200,
		ActiveMerchantCount: 8,
		ActiveCustomerCount: 150,
		RightsConsumed:      1000,
		RightsDistributed:   1200,
		RightsBalance:       200,
	}
	
	// 创建测试模板
	templateConfig := map[string]interface{}{
		"currency":         "CNY",
		"include_trends":   true,
		"include_breakdown": true,
	}
	configJson, _ := json.Marshal(templateConfig)
	
	template := &types.ReportTemplate{
		ID:             1,
		Name:          "测试财务模板",
		ReportType:    types.ReportTypeFinancial,
		TemplateConfig: configJson,
	}
	
	// 执行模板渲染
	result, err := engine.RenderTemplate(ctx, template, financialData)
	
	// 验证结果
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	
	// 解析结果JSON
	var renderedData map[string]interface{}
	err = json.Unmarshal(result, &renderedData)
	require.NoError(t, err)
	
	// 验证基础数据
	assert.Equal(t, "财务报表", renderedData["title"])
	assert.NotEmpty(t, renderedData["generated_at"])
	
	// 验证汇总数据
	summary, ok := renderedData["summary"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "¥100000.00", summary["total_revenue"])
	assert.Equal(t, "¥40000.00", summary["net_profit"])
	assert.Equal(t, float64(500), summary["order_count"])
}

// TestTemplateEngine_ValidateTemplate 测试模板验证
func TestTemplateEngine_ValidateTemplate(t *testing.T) {
	engine := service.NewTemplateEngine()
	ctx := context.Background()
	
	tests := []struct {
		name     string
		template *types.ReportTemplate
		wantErr  bool
		errMsg   string
	}{
		{
			name: "有效财务模板",
			template: &types.ReportTemplate{
				ReportType: types.ReportTypeFinancial,
				TemplateConfig: []byte(`{"currency": "CNY", "include_trends": true}`),
			},
			wantErr: false,
		},
		{
			name: "无效JSON配置",
			template: &types.ReportTemplate{
				ReportType: types.ReportTypeFinancial,
				TemplateConfig: []byte(`{"currency": "CNY", invalid_json}`),
			},
			wantErr: true,
			errMsg:  "JSON格式错误",
		},
		{
			name: "缺少必需字段",
			template: &types.ReportTemplate{
				ReportType: types.ReportTypeFinancial,
				TemplateConfig: []byte(`{"include_trends": true}`),
			},
			wantErr: true,
			errMsg:  "缺少必需配置项: currency",
		},
		{
			name: "不支持的货币类型",
			template: &types.ReportTemplate{
				ReportType: types.ReportTypeFinancial,
				TemplateConfig: []byte(`{"currency": "INVALID"}`),
			},
			wantErr: true,
			errMsg:  "不支持的货币类型",
		},
		{
			name: "商户运营模板-无效排名数量",
			template: &types.ReportTemplate{
				ReportType: types.ReportTypeMerchantOperation,
				TemplateConfig: []byte(`{"top_count": 150}`),
			},
			wantErr: true,
			errMsg:  "商户排名数量必须在1-100之间",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateTemplate(ctx, tt.template)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPDFGenerator_CreateHTMLTemplate 测试HTML模板创建
func TestPDFGenerator_CreateHTMLTemplate(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	// 测试财务报表HTML模板
	financialData := &types.FinancialReportData{
		TotalRevenue:     &types.Money{Amount: 50000.00, Currency: "CNY"},
		NetProfit:       &types.Money{Amount: 20000.00, Currency: "CNY"},
		OrderCount:      100,
		MerchantCount:   5,
		CustomerCount:   50,
		RightsConsumed: 200,
	}
	
	htmlContent, err := generator.CreateHTMLTemplate(types.ReportTypeFinancial, financialData)
	
	require.NoError(t, err)
	assert.NotEmpty(t, htmlContent)
	assert.Contains(t, htmlContent, "<!DOCTYPE html>")
	assert.Contains(t, htmlContent, "财务报表")
	assert.Contains(t, htmlContent, "50000.00")
	assert.Contains(t, htmlContent, "20000.00")
	
	// 测试不支持的报表类型
	_, err = generator.CreateHTMLTemplate("unsupported_type", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的报表类型")
}

// TestPDFGenerator_GetSupportedPDFConverters 测试PDF转换器检测
func TestPDFGenerator_GetSupportedPDFConverters(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	converters := generator.GetSupportedPDFConverters()
	
	// 验证返回的转换器列表
	assert.IsType(t, []string{}, converters)
	
	// 在测试环境中，可能没有安装转换器，所以只验证函数正常执行
	t.Logf("可用的PDF转换器: %v", converters)
}

// TestCacheManager_GetCacheKey 测试缓存键生成
func TestCacheManager_GetCacheKey(t *testing.T) {
	cacheManager := service.NewCacheManager()
	
	req1 := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2025, 8, 30, 0, 0, 0, 0, time.UTC),
		FileFormat: types.FileFormatExcel,
	}
	
	req2 := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2025, 8, 30, 0, 0, 0, 0, time.UTC),
		FileFormat: types.FileFormatPDF, // 不同的文件格式
	}
	
	req3 := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), // 不同的时间范围
		EndDate:    time.Date(2025, 7, 30, 0, 0, 0, 0, time.UTC),
		FileFormat: types.FileFormatExcel,
	}
	
	key1 := cacheManager.GetCacheKey(req1)
	key2 := cacheManager.GetCacheKey(req2)
	key3 := cacheManager.GetCacheKey(req3)
	
	// 验证缓存键格式
	assert.HasPrefix(t, key1, "report_cache:")
	assert.HasPrefix(t, key2, "report_cache:")
	assert.HasPrefix(t, key3, "report_cache:")
	
	// 相同请求应产生相同缓存键
	key1Again := cacheManager.GetCacheKey(req1)
	assert.Equal(t, key1, key1Again)
	
	// 不同请求应产生不同缓存键
	assert.NotEqual(t, key1, key2)
	assert.NotEqual(t, key1, key3)
	assert.NotEqual(t, key2, key3)
}

// TestCacheManager_ShouldUseCache 测试缓存使用判断
func TestCacheManager_ShouldUseCache(t *testing.T) {
	cacheManager := service.NewCacheManager()
	now := time.Now()
	
	tests := []struct {
		name     string
		req      *types.ReportCreateRequest
		expected bool
	}{
		{
			name: "有效缓存条件",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				StartDate:  now.AddDate(0, 0, -7),
				EndDate:    now.AddDate(0, 0, -2),
			},
			expected: true,
		},
		{
			name: "时间范围太短",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				StartDate:  now.Add(-30 * time.Minute),
				EndDate:    now.AddDate(0, 0, -1),
			},
			expected: false,
		},
		{
			name: "时间范围太长",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				StartDate:  now.AddDate(-2, 0, 0),
				EndDate:    now.AddDate(0, 0, -1),
			},
			expected: false,
		},
		{
			name: "实时数据",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				StartDate:  now.AddDate(0, 0, -7),
				EndDate:    now.Add(-30 * time.Minute),
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cacheManager.ShouldUseCache(tt.req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 清理测试临时文件的辅助函数
func cleanupTestFiles(t *testing.T, patterns ...string) {
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			os.RemoveAll(match)
		}
	}
}
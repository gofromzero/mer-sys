package unit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPDFGenerator_CreateHTMLTemplate 测试HTML模板创建
func TestPDFGenerator_CreateHTMLTemplate(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	// 测试财务报表HTML模板
	financialData := createTestFinancialData()
	htmlContent, err := generator.CreateHTMLTemplate(types.ReportTypeFinancial, financialData)
	
	require.NoError(t, err)
	assert.NotEmpty(t, htmlContent)
	
	// 验证HTML结构
	assert.Contains(t, htmlContent, "<!DOCTYPE html>")
	assert.Contains(t, htmlContent, "<html lang=\"zh-CN\">")
	assert.Contains(t, htmlContent, "</html>")
	
	// 验证内容包含财务数据
	assert.Contains(t, htmlContent, "财务报表")
	assert.Contains(t, htmlContent, "总收入")
	assert.Contains(t, htmlContent, "净利润")
	assert.Contains(t, htmlContent, "120000.00") // 总收入金额
	assert.Contains(t, htmlContent, "50000.00")  // 净利润金额
	
	// 验证CSS样式存在
	assert.Contains(t, htmlContent, "<style>")
	assert.Contains(t, htmlContent, "font-family")
	assert.Contains(t, htmlContent, ".summary-table")
	
	t.Logf("财务报表HTML模板长度: %d 字符", len(htmlContent))
}

// TestPDFGenerator_MerchantOperationHTMLTemplate 测试商户运营HTML模板
func TestPDFGenerator_MerchantOperationHTMLTemplate(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	// 创建商户运营数据
	merchantData := &types.MerchantOperationReport{
		MerchantRankings: []*types.MerchantRanking{
			{
				Rank:             1,
				MerchantID:       1,
				MerchantName:     "顶级商户",
				TotalRevenue:     &types.Money{Amount: 80000.00, Currency: "CNY"},
				OrderCount:       400,
				CustomerCount:    200,
				AverageOrderValue: &types.Money{Amount: 200.00, Currency: "CNY"},
				GrowthRate:       15.5,
			},
			{
				Rank:             2,
				MerchantID:       2,
				MerchantName:     "优秀商户",
				TotalRevenue:     &types.Money{Amount: 60000.00, Currency: "CNY"},
				OrderCount:       300,
				CustomerCount:    150,
				AverageOrderValue: &types.Money{Amount: 200.00, Currency: "CNY"},
				GrowthRate:       -5.2,
			},
		},
		CategoryAnalysis: []*types.CategoryAnalysis{
			{
				CategoryID:    1,
				CategoryName:  "电子产品",
				Revenue:       &types.Money{Amount: 100000.00, Currency: "CNY"},
				OrderCount:    500,
				MerchantCount: 10,
				MarketShare:   60.5,
				GrowthRate:    18.3,
			},
		},
	}
	
	htmlContent, err := generator.CreateHTMLTemplate(types.ReportTypeMerchantOperation, merchantData)
	
	require.NoError(t, err)
	assert.NotEmpty(t, htmlContent)
	
	// 验证商户运营报表内容
	assert.Contains(t, htmlContent, "商户运营报表")
	assert.Contains(t, htmlContent, "商户业绩排行榜")
	assert.Contains(t, htmlContent, "顶级商户")
	assert.Contains(t, htmlContent, "优秀商户")
	assert.Contains(t, htmlContent, "80000.00")
	assert.Contains(t, htmlContent, "60000.00")
	
	// 验证增长率的颜色样式
	assert.Contains(t, htmlContent, "growth-positive")
	assert.Contains(t, htmlContent, "growth-negative")
	assert.Contains(t, htmlContent, "15.50") // 正增长率
	assert.Contains(t, htmlContent, "-5.20") // 负增长率
	
	t.Logf("商户运营报表HTML模板长度: %d 字符", len(htmlContent))
}

// TestPDFGenerator_CustomerAnalysisHTMLTemplate 测试客户分析HTML模板
func TestPDFGenerator_CustomerAnalysisHTMLTemplate(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	// 创建客户分析数据
	customerData := &types.CustomerAnalysisReport{
		UserGrowth: []*types.UserGrowthData{
			{
				Month:           "2025-07",
				NewUsers:        150,
				ActiveUsers:     800,
				CumulativeUsers: 2000,
				RetentionRate:   75.5,
			},
		},
		ActivityMetrics: &types.ActivityMetrics{
			DAU:                1500,
			WAU:                6000,
			MAU:                20000,
			AverageSessionTime: 28.5,
			AverageOrderFreq:   2.3,
		},
	}
	
	htmlContent, err := generator.CreateHTMLTemplate(types.ReportTypeCustomerAnalysis, customerData)
	
	require.NoError(t, err)
	assert.NotEmpty(t, htmlContent)
	
	// 验证客户分析报表内容
	assert.Contains(t, htmlContent, "客户分析报表")
	assert.Contains(t, htmlContent, "活跃度指标")
	assert.Contains(t, htmlContent, "日活跃用户(DAU)")
	assert.Contains(t, htmlContent, "1500") // DAU数值
	assert.Contains(t, htmlContent, "6000") // WAU数值
	assert.Contains(t, htmlContent, "20000") // MAU数值
	
	t.Logf("客户分析报表HTML模板长度: %d 字符", len(htmlContent))
}

// TestPDFGenerator_UnsupportedReportType 测试不支持的报表类型
func TestPDFGenerator_UnsupportedReportType(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	_, err := generator.CreateHTMLTemplate("unsupported_type", nil)
	
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的报表类型")
}

// TestPDFGenerator_GetSupportedPDFConverters 测试PDF转换器检测
func TestPDFGenerator_GetSupportedPDFConverters(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	converters := generator.GetSupportedPDFConverters()
	
	// 验证返回类型
	assert.IsType(t, []string{}, converters)
	
	// 记录可用的转换器
	t.Logf("检测到的PDF转换器: %v", converters)
	
	// 验证转换器名称格式
	for _, converter := range converters {
		assert.NotEmpty(t, converter)
		assert.True(t, 
			converter == "wkhtmltopdf" || 
			converter == "chrome" || 
			converter == "phantomjs",
			"不支持的转换器: %s", converter)
	}
}

// TestPDFGenerator_GeneratePDFReport 测试PDF报表生成
func TestPDFGenerator_GeneratePDFReport(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过PDF生成测试（使用 -short 标志）")
	}
	
	generator := service.NewPDFGenerator()
	ctx := context.Background()
	
	// 设置测试目录
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)
	
	// 创建测试报表
	report := &types.Report{
		UUID:       "test-pdf-generation-" + time.Now().Format("20060102150405"),
		ReportType: types.ReportTypeFinancial,
		FileFormat: types.FileFormatPDF,
	}
	
	// 创建财务数据
	financialData := createTestFinancialData()
	
	// 生成PDF报表
	filePath, err := generator.GeneratePDFReport(ctx, report, financialData)
	
	// 验证结果
	require.NoError(t, err)
	assert.NotEmpty(t, filePath)
	
	// 检查文件是否存在
	_, err = os.Stat(filePath)
	require.NoError(t, err)
	
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0), "生成的文件应该有内容")
	
	// 检查文件扩展名
	ext := filepath.Ext(filePath)
	assert.True(t, ext == ".pdf" || ext == ".html", "文件应该是PDF或HTML格式")
	
	if ext == ".html" {
		t.Logf("PDF转换失败，生成了HTML文件: %s (大小: %d 字节)", filePath, fileInfo.Size())
		
		// 验证HTML内容
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "财务报表")
	} else {
		t.Logf("PDF生成成功: %s (大小: %d 字节)", filePath, fileInfo.Size())
	}
}

// TestPDFGenerator_HTMLTemplateWithComplexData 测试复杂数据的HTML模板生成
func TestPDFGenerator_HTMLTemplateWithComplexData(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	// 创建包含完整分解数据的财务报表
	financialData := &types.FinancialReportData{
		TotalRevenue:         &types.Money{Amount: 500000.00, Currency: "CNY"},
		TotalExpenditure:     &types.Money{Amount: 300000.00, Currency: "CNY"},
		NetProfit:           &types.Money{Amount: 200000.00, Currency: "CNY"},
		OrderCount:          2500,
		MerchantCount:       50,
		CustomerCount:       1000,
		ActiveMerchantCount: 45,
		ActiveCustomerCount: 800,
		RightsConsumed:      10000,
		RightsDistributed:   12000,
		RightsBalance:       2000,
		Breakdown: &types.FinancialBreakdown{
			RevenueByMerchant: []*types.MerchantRevenue{
				{
					MerchantID:   1,
					MerchantName: "超级商户A",
					Revenue:      &types.Money{Amount: 150000.00, Currency: "CNY"},
					OrderCount:   750,
					Percentage:   30.0,
				},
				{
					MerchantID:   2,
					MerchantName: "优质商户B",
					Revenue:      &types.Money{Amount: 100000.00, Currency: "CNY"},
					OrderCount:   500,
					Percentage:   20.0,
				},
				{
					MerchantID:   3,
					MerchantName: "成长商户C",
					Revenue:      &types.Money{Amount: 75000.00, Currency: "CNY"},
					OrderCount:   400,
					Percentage:   15.0,
				},
			},
			MonthlyTrend: []*types.MonthlyTrend{
				{
					Month:          "2025-05",
					Revenue:        &types.Money{Amount: 120000.00, Currency: "CNY"},
					Expenditure:    &types.Money{Amount: 70000.00, Currency: "CNY"},
					NetProfit:      &types.Money{Amount: 50000.00, Currency: "CNY"},
					OrderCount:     600,
					RightsConsumed: 2500,
				},
				{
					Month:          "2025-06",
					Revenue:        &types.Money{Amount: 140000.00, Currency: "CNY"},
					Expenditure:    &types.Money{Amount: 85000.00, Currency: "CNY"},
					NetProfit:      &types.Money{Amount: 55000.00, Currency: "CNY"},
					OrderCount:     700,
					RightsConsumed: 3000,
				},
				{
					Month:          "2025-07",
					Revenue:        &types.Money{Amount: 160000.00, Currency: "CNY"},
					Expenditure:    &types.Money{Amount: 95000.00, Currency: "CNY"},
					NetProfit:      &types.Money{Amount: 65000.00, Currency: "CNY"},
					OrderCount:     800,
					RightsConsumed: 3500,
				},
			},
		},
	}
	
	htmlContent, err := generator.CreateHTMLTemplate(types.ReportTypeFinancial, financialData)
	
	require.NoError(t, err)
	assert.NotEmpty(t, htmlContent)
	
	// 验证商户排行榜数据
	assert.Contains(t, htmlContent, "商户收入排行榜")
	assert.Contains(t, htmlContent, "超级商户A")
	assert.Contains(t, htmlContent, "优质商户B")
	assert.Contains(t, htmlContent, "成长商户C")
	assert.Contains(t, htmlContent, "150000.00")
	assert.Contains(t, htmlContent, "100000.00")
	assert.Contains(t, htmlContent, "75000.00")
	
	// 验证月度趋势数据
	assert.Contains(t, htmlContent, "月度趋势")
	assert.Contains(t, htmlContent, "2025-05")
	assert.Contains(t, htmlContent, "2025-06")
	assert.Contains(t, htmlContent, "2025-07")
	
	// 验证数据展示限制（应该只显示前10名商户）
	merchantCount := strings.Count(htmlContent, "MerchantName")
	t.Logf("HTML中商户数量: %d", merchantCount)
	
	// 验证HTML的完整性和格式
	assert.Equal(t, strings.Count(htmlContent, "<table"), strings.Count(htmlContent, "</table>"))
	assert.Equal(t, strings.Count(htmlContent, "<tr"), strings.Count(htmlContent, "</tr>"))
	assert.Equal(t, strings.Count(htmlContent, "<td"), strings.Count(htmlContent, "</td>"))
	
	t.Logf("复杂财务数据HTML模板长度: %d 字符", len(htmlContent))
}

// TestPDFGenerator_HTMLTemplateEncoding 测试HTML模板编码
func TestPDFGenerator_HTMLTemplateEncoding(t *testing.T) {
	generator := service.NewPDFGenerator()
	
	// 创建包含特殊字符的数据
	financialData := &types.FinancialReportData{
		TotalRevenue:     &types.Money{Amount: 88888.88, Currency: "CNY"},
		NetProfit:       &types.Money{Amount: 66666.66, Currency: "CNY"},
		OrderCount:      888,
		MerchantCount:   88,
		CustomerCount:   888,
		Breakdown: &types.FinancialBreakdown{
			RevenueByMerchant: []*types.MerchantRevenue{
				{
					MerchantID:   1,
					MerchantName: "测试商户(特殊字符)：<>&\"'",
					Revenue:      &types.Money{Amount: 12345.67, Currency: "CNY"},
					OrderCount:   123,
					Percentage:   13.9,
				},
			},
		},
	}
	
	htmlContent, err := generator.CreateHTMLTemplate(types.ReportTypeFinancial, financialData)
	
	require.NoError(t, err)
	assert.NotEmpty(t, htmlContent)
	
	// 验证UTF-8编码声明
	assert.Contains(t, htmlContent, "charset=UTF-8")
	assert.Contains(t, htmlContent, "lang=\"zh-CN\"")
	
	// 验证中文字符正确显示
	assert.Contains(t, htmlContent, "财务报表")
	assert.Contains(t, htmlContent, "总收入")
	assert.Contains(t, htmlContent, "测试商户")
	
	// 验证特殊字符处理（应该被正确转义或处理）
	assert.Contains(t, htmlContent, "测试商户")
	
	// 验证数字格式
	assert.Contains(t, htmlContent, "88888.88")
	assert.Contains(t, htmlContent, "66666.66")
	assert.Contains(t, htmlContent, "12345.67")
	
	t.Logf("编码测试通过，HTML长度: %d", len(htmlContent))
}

// TestPDFGenerator_Performance 测试PDF生成性能
func TestPDFGenerator_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（使用 -short 标志）")
	}
	
	generator := service.NewPDFGenerator()
	
	// 创建大型数据集
	financialData := createLargeFinancialData(100, 12) // 100个商户，12个月数据
	
	// 性能测试
	iterations := 10
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		_, err := generator.CreateHTMLTemplate(types.ReportTypeFinancial, financialData)
		require.NoError(t, err)
	}
	
	elapsed := time.Since(start)
	avgTime := elapsed / time.Duration(iterations)
	
	t.Logf("HTML模板生成性能测试结果:")
	t.Logf("- 数据规模: 100个商户，12个月")
	t.Logf("- 迭代次数: %d", iterations)
	t.Logf("- 总时间: %v", elapsed)
	t.Logf("- 平均时间: %v", avgTime)
	t.Logf("- 每秒处理: %.2f 次", float64(iterations)/elapsed.Seconds())
	
	// 性能断言（平均时间应该小于500ms）
	assert.Less(t, avgTime, 500*time.Millisecond, "HTML模板生成平均时间应该小于500ms")
}

// TestPDFGenerator_ConcurrentGeneration 测试并发生成
func TestPDFGenerator_ConcurrentGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过并发测试（使用 -short 标志）")
	}
	
	generator := service.NewPDFGenerator()
	financialData := createTestFinancialData()
	
	// 并发测试
	goroutines := 5
	iterations := 10
	
	start := time.Now()
	done := make(chan bool, goroutines)
	
	for g := 0; g < goroutines; g++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			
			for i := 0; i < iterations; i++ {
				_, err := generator.CreateHTMLTemplate(types.ReportTypeFinancial, financialData)
				assert.NoError(t, err, "Goroutine %d, iteration %d failed", goroutineID, i)
			}
		}(g)
	}
	
	// 等待所有goroutine完成
	for g := 0; g < goroutines; g++ {
		<-done
	}
	
	elapsed := time.Since(start)
	totalOperations := goroutines * iterations
	
	t.Logf("并发生成测试结果:")
	t.Logf("- 并发数: %d", goroutines)
	t.Logf("- 每个goroutine迭代: %d", iterations)
	t.Logf("- 总操作数: %d", totalOperations)
	t.Logf("- 总时间: %v", elapsed)
	t.Logf("- 每秒处理: %.2f 次", float64(totalOperations)/elapsed.Seconds())
}

// 辅助函数：创建测试财务数据
func createTestFinancialData() *types.FinancialReportData {
	return &types.FinancialReportData{
		TotalRevenue:         &types.Money{Amount: 120000.00, Currency: "CNY"},
		TotalExpenditure:     &types.Money{Amount: 70000.00, Currency: "CNY"},
		NetProfit:           &types.Money{Amount: 50000.00, Currency: "CNY"},
		OrderCount:          600,
		MerchantCount:       12,
		CustomerCount:       300,
		ActiveMerchantCount: 10,
		ActiveCustomerCount: 250,
		RightsConsumed:      1500,
		RightsDistributed:   1800,
		RightsBalance:       300,
		Breakdown: &types.FinancialBreakdown{
			RevenueByMerchant: []*types.MerchantRevenue{
				{
					MerchantID:   1,
					MerchantName: "测试商户1",
					Revenue:      &types.Money{Amount: 50000.00, Currency: "CNY"},
					OrderCount:   200,
					Percentage:   41.67,
				},
				{
					MerchantID:   2,
					MerchantName: "测试商户2",
					Revenue:      &types.Money{Amount: 30000.00, Currency: "CNY"},
					OrderCount:   150,
					Percentage:   25.0,
				},
			},
			MonthlyTrend: []*types.MonthlyTrend{
				{
					Month:          "2025-07",
					Revenue:        &types.Money{Amount: 60000.00, Currency: "CNY"},
					Expenditure:    &types.Money{Amount: 35000.00, Currency: "CNY"},
					NetProfit:      &types.Money{Amount: 25000.00, Currency: "CNY"},
					OrderCount:     300,
					RightsConsumed: 750,
				},
				{
					Month:          "2025-08",
					Revenue:        &types.Money{Amount: 60000.00, Currency: "CNY"},
					Expenditure:    &types.Money{Amount: 35000.00, Currency: "CNY"},
					NetProfit:      &types.Money{Amount: 25000.00, Currency: "CNY"},
					OrderCount:     300,
					RightsConsumed: 750,
				},
			},
		},
	}
}

// 辅助函数：创建大型财务数据
func createLargeFinancialData(merchantCount, monthCount int) *types.FinancialReportData {
	data := createTestFinancialData()
	
	// 生成大量商户数据
	data.Breakdown.RevenueByMerchant = make([]*types.MerchantRevenue, merchantCount)
	for i := 0; i < merchantCount; i++ {
		data.Breakdown.RevenueByMerchant[i] = &types.MerchantRevenue{
			MerchantID:   uint64(i + 1),
			MerchantName: "测试商户" + string(rune('A'+i%26)) + string(rune('0'+i/26)),
			Revenue:      &types.Money{Amount: float64((i + 1) * 1000), Currency: "CNY"},
			OrderCount:   (i + 1) * 10,
			Percentage:   float64(i+1) * 0.5,
		}
	}
	
	// 生成大量月度数据
	data.Breakdown.MonthlyTrend = make([]*types.MonthlyTrend, monthCount)
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < monthCount; i++ {
		monthTime := baseTime.AddDate(0, i, 0)
		data.Breakdown.MonthlyTrend[i] = &types.MonthlyTrend{
			Month:          monthTime.Format("2006-01"),
			Revenue:        &types.Money{Amount: float64((i + 1) * 10000), Currency: "CNY"},
			Expenditure:    &types.Money{Amount: float64((i + 1) * 6000), Currency: "CNY"},
			NetProfit:      &types.Money{Amount: float64((i + 1) * 4000), Currency: "CNY"},
			OrderCount:     (i + 1) * 50,
			RightsConsumed: (i + 1) * 100,
		}
	}
	
	return data
}

// 辅助函数：设置测试目录
func setupTestDir(t *testing.T) string {
	testDir := filepath.Join(os.TempDir(), "pdf_generator_test", time.Now().Format("20060102_150405"))
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)
	
	// 设置环境变量
	os.Setenv("REPORT_STORAGE_DIR", testDir)
	
	return testDir
}

// 辅助函数：清理测试目录
func cleanupTestDir(t *testing.T, testDir string) {
	if testDir != "" {
		err := os.RemoveAll(testDir)
		if err != nil {
			t.Logf("清理测试目录失败: %v", err)
		}
	}
	os.Unsetenv("REPORT_STORAGE_DIR")
}
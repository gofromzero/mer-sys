package unit

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplateEngine_FormattersRegistration 测试格式化器注册
func TestTemplateEngine_FormattersRegistration(t *testing.T) {
	engine := service.NewTemplateEngine()
	
	// 测试获取支持的变量
	variables := engine.GetSupportedVariables()
	
	// 验证基本变量存在
	expectedVariables := []string{
		"{{.TotalRevenue}}",
		"{{.OrderCount}}",
		"{{.MerchantCount}}",
		"{{.CustomerCount}}",
		"{{.NetProfit}}",
	}
	
	for _, expected := range expectedVariables {
		_, exists := variables[expected]
		assert.True(t, exists, "变量 %s 应该存在", expected)
	}
	
	t.Logf("支持的变量数量: %d", len(variables))
}

// TestTemplateEngine_FinancialTemplateRendering 测试财务报表模板渲染
func TestTemplateEngine_FinancialTemplateRendering(t *testing.T) {
	engine := service.NewTemplateEngine()
	ctx := context.Background()
	
	// 创建复杂的财务数据
	financialData := &types.FinancialReportData{
		TotalRevenue:         &types.Money{Amount: 250000.50, Currency: "CNY"},
		TotalExpenditure:     &types.Money{Amount: 150000.25, Currency: "CNY"},
		NetProfit:           &types.Money{Amount: 100000.25, Currency: "CNY"},
		OrderCount:          1250,
		MerchantCount:       25,
		CustomerCount:       500,
		ActiveMerchantCount: 20,
		ActiveCustomerCount: 400,
		RightsConsumed:      5000,
		RightsDistributed:   6000,
		RightsBalance:       1000,
		OrderAmount:         &types.Money{Amount: 250000.50, Currency: "CNY"},
		Breakdown: &types.FinancialBreakdown{
			RevenueByMerchant: []*types.MerchantRevenue{
				{
					MerchantID:   1,
					MerchantName: "顶级商户A",
					Revenue:      &types.Money{Amount: 100000.00, Currency: "CNY"},
					OrderCount:   400,
					Percentage:   40.0,
				},
				{
					MerchantID:   2,
					MerchantName: "优质商户B",
					Revenue:      &types.Money{Amount: 80000.00, Currency: "CNY"},
					OrderCount:   350,
					Percentage:   32.0,
				},
			},
			MonthlyTrend: []*types.MonthlyTrend{
				{
					Month:          "2025-06",
					Revenue:        &types.Money{Amount: 80000.00, Currency: "CNY"},
					Expenditure:    &types.Money{Amount: 50000.00, Currency: "CNY"},
					NetProfit:      &types.Money{Amount: 30000.00, Currency: "CNY"},
					OrderCount:     400,
					RightsConsumed: 2000,
				},
				{
					Month:          "2025-07",
					Revenue:        &types.Money{Amount: 90000.00, Currency: "CNY"},
					Expenditure:    &types.Money{Amount: 55000.00, Currency: "CNY"},
					NetProfit:      &types.Money{Amount: 35000.00, Currency: "CNY"},
					OrderCount:     450,
					RightsConsumed: 2200,
				},
			},
		},
	}
	
	// 测试不同的模板配置
	testConfigs := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "基础配置",
			config: map[string]interface{}{
				"currency": "CNY",
			},
		},
		{
			name: "包含趋势",
			config: map[string]interface{}{
				"currency":         "CNY",
				"include_trends":   true,
			},
		},
		{
			name: "包含分解数据",
			config: map[string]interface{}{
				"currency":          "CNY",
				"include_breakdown": true,
			},
		},
		{
			name: "完整配置",
			config: map[string]interface{}{
				"currency":          "CNY",
				"include_trends":    true,
				"include_breakdown": true,
			},
		},
	}
	
	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {
			// 创建模板
			configJson, _ := json.Marshal(tc.config)
			template := &types.ReportTemplate{
				ID:             1,
				Name:          tc.name,
				ReportType:    types.ReportTypeFinancial,
				TemplateConfig: configJson,
			}
			
			// 渲染模板
			result, err := engine.RenderTemplate(ctx, template, financialData)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			// 解析结果
			var renderedData map[string]interface{}
			err = json.Unmarshal(result, &renderedData)
			require.NoError(t, err)
			
			// 验证基础数据
			assert.Equal(t, "财务报表", renderedData["title"])
			assert.NotEmpty(t, renderedData["generated_at"])
			
			// 验证汇总数据
			summary, ok := renderedData["summary"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "¥250000.50", summary["total_revenue"])
			assert.Equal(t, "¥100000.25", summary["net_profit"])
			assert.Equal(t, float64(1250), summary["order_count"])
			
			// 验证条件性数据
			if tc.config["include_trends"] == true {
				_, hasTrends := renderedData["trends"]
				assert.True(t, hasTrends, "应该包含趋势数据")
			}
			
			if tc.config["include_breakdown"] == true {
				_, hasMerchantRankings := renderedData["merchant_rankings"]
				assert.True(t, hasMerchantRankings, "应该包含商户排名数据")
			}
			
			t.Logf("配置 %s 渲染成功，输出长度: %d", tc.name, len(result))
		})
	}
}

// TestTemplateEngine_MerchantOperationRendering 测试商户运营模板渲染
func TestTemplateEngine_MerchantOperationRendering(t *testing.T) {
	engine := service.NewTemplateEngine()
	ctx := context.Background()
	
	// 创建商户运营数据
	merchantData := &types.MerchantOperationReport{
		MerchantRankings: []*types.MerchantRanking{
			{
				Rank:             1,
				MerchantID:       1,
				MerchantName:     "明星商户",
				TotalRevenue:     &types.Money{Amount: 120000.00, Currency: "CNY"},
				OrderCount:       600,
				CustomerCount:    200,
				AverageOrderValue: &types.Money{Amount: 200.00, Currency: "CNY"},
				GrowthRate:       25.5,
			},
			{
				Rank:             2,
				MerchantID:       2,
				MerchantName:     "优秀商户",
				TotalRevenue:     &types.Money{Amount: 80000.00, Currency: "CNY"},
				OrderCount:       400,
				CustomerCount:    150,
				AverageOrderValue: &types.Money{Amount: 200.00, Currency: "CNY"},
				GrowthRate:       15.2,
			},
		},
		CategoryAnalysis: []*types.CategoryAnalysis{
			{
				CategoryID:    1,
				CategoryName:  "数码产品",
				Revenue:       &types.Money{Amount: 150000.00, Currency: "CNY"},
				OrderCount:    750,
				MerchantCount: 12,
				MarketShare:   60.0,
				GrowthRate:    18.5,
			},
		},
		GrowthMetrics: &types.GrowthMetrics{
			RevenueGrowthRate:      15.8,
			OrderCountGrowthRate:   12.3,
			CustomerGrowthRate:     20.1,
			MerchantGrowthRate:     8.5,
		},
	}
	
	// 测试不同的排名数量配置
	testConfigs := []struct {
		name     string
		topCount int
	}{
		{"默认排名", 0}, // 不设置，使用默认值
		{"前3名", 3},
		{"前5名", 5},
		{"前10名", 10},
		{"超出范围", 50}, // 超出实际数据量
	}
	
	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {
			// 创建配置
			config := map[string]interface{}{
				"include_growth": true,
			}
			if tc.topCount > 0 {
				config["top_count"] = float64(tc.topCount)
			}
			
			configJson, _ := json.Marshal(config)
			template := &types.ReportTemplate{
				ID:             1,
				Name:          tc.name,
				ReportType:    types.ReportTypeMerchantOperation,
				TemplateConfig: configJson,
			}
			
			// 渲染模板
			result, err := engine.RenderTemplate(ctx, template, merchantData)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			// 解析结果
			var renderedData map[string]interface{}
			err = json.Unmarshal(result, &renderedData)
			require.NoError(t, err)
			
			// 验证基础信息
			assert.Equal(t, "商户运营报表", renderedData["title"])
			assert.NotEmpty(t, renderedData["generated_at"])
			
			// 验证商户排名数据
			merchantRankings, ok := renderedData["merchant_rankings"].([]interface{})
			if ok {
				expectedCount := tc.topCount
				if tc.topCount == 0 {
					expectedCount = 10 // 默认值
				}
				if expectedCount > len(merchantData.MerchantRankings) {
					expectedCount = len(merchantData.MerchantRankings)
				}
				assert.LessOrEqual(t, len(merchantRankings), expectedCount)
			}
			
			// 验证增长指标
			_, hasGrowthMetrics := renderedData["growth_metrics"]
			assert.True(t, hasGrowthMetrics, "应该包含增长指标")
			
			t.Logf("配置 %s 渲染成功", tc.name)
		})
	}
}

// TestTemplateEngine_CustomerAnalysisRendering 测试客户分析模板渲染
func TestTemplateEngine_CustomerAnalysisRendering(t *testing.T) {
	engine := service.NewTemplateEngine()
	ctx := context.Background()
	
	// 创建客户分析数据
	customerData := &types.CustomerAnalysisReport{
		UserGrowth: []*types.UserGrowthData{
			{
				Month:           "2025-05",
				NewUsers:        100,
				ActiveUsers:     800,
				CumulativeUsers: 2000,
				RetentionRate:   72.5,
			},
			{
				Month:           "2025-06",
				NewUsers:        120,
				ActiveUsers:     900,
				CumulativeUsers: 2120,
				RetentionRate:   75.2,
			},
			{
				Month:           "2025-07",
				NewUsers:        150,
				ActiveUsers:     1000,
				CumulativeUsers: 2270,
				RetentionRate:   78.8,
			},
		},
		ActivityMetrics: &types.ActivityMetrics{
			DAU:                2500,
			WAU:                8000,
			MAU:                25000,
			AverageSessionTime: 32.5,
			AverageOrderFreq:   2.8,
		},
		ConsumptionBehavior: &types.ConsumptionBehavior{
			AverageOrderValue:     &types.Money{Amount: 180.50, Currency: "CNY"},
			OrderFrequency:        2.3,
			PreferredCategories:   []string{"数码", "服装", "食品"},
			PurchaseTimeDistribution: map[string]float64{
				"morning":   15.5,
				"afternoon": 35.2,
				"evening":   49.3,
			},
		},
		RetentionAnalysis: &types.RetentionAnalysis{
			Day1Retention:  85.2,
			Day7Retention:  68.8,
			Day30Retention: 42.5,
			CohortAnalysis: map[string]float64{
				"2025-05": 78.2,
				"2025-06": 82.1,
				"2025-07": 85.5,
			},
		},
		ChurnAnalysis: &types.ChurnAnalysis{
			ChurnRate:      12.5,
			ChurnReasons:   []string{"价格因素", "产品质量", "服务体验"},
			RiskSegments:   []string{"低活跃用户", "长期未购买用户"},
		},
	}
	
	// 测试不同的分析模块配置
	testConfigs := []struct {
		name           string
		includeRetention bool
		includeChurn     bool
	}{
		{"基础分析", false, false},
		{"包含留存分析", true, false},
		{"包含流失分析", false, true},
		{"完整分析", true, true},
	}
	
	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {
			// 创建配置
			config := map[string]interface{}{
				"include_retention": tc.includeRetention,
				"include_churn":     tc.includeChurn,
			}
			
			configJson, _ := json.Marshal(config)
			template := &types.ReportTemplate{
				ID:             1,
				Name:          tc.name,
				ReportType:    types.ReportTypeCustomerAnalysis,
				TemplateConfig: configJson,
			}
			
			// 渲染模板
			result, err := engine.RenderTemplate(ctx, template, customerData)
			require.NoError(t, err)
			assert.NotEmpty(t, result)
			
			// 解析结果
			var renderedData map[string]interface{}
			err = json.Unmarshal(result, &renderedData)
			require.NoError(t, err)
			
			// 验证基础信息
			assert.Equal(t, "客户分析报表", renderedData["title"])
			assert.NotEmpty(t, renderedData["generated_at"])
			
			// 验证用户增长数据
			userGrowth, ok := renderedData["user_growth"].([]interface{})
			assert.True(t, ok)
			assert.Len(t, userGrowth, 3)
			
			// 验证活跃度指标
			activityMetrics, ok := renderedData["activity_metrics"]
			assert.True(t, ok)
			assert.NotNil(t, activityMetrics)
			
			// 验证条件性数据
			if tc.includeRetention {
				_, hasRetention := renderedData["retention_analysis"]
				assert.True(t, hasRetention, "应该包含留存分析")
			}
			
			if tc.includeChurn {
				_, hasChurn := renderedData["churn_analysis"]
				assert.True(t, hasChurn, "应该包含流失分析")
			}
			
			t.Logf("配置 %s 渲染成功", tc.name)
		})
	}
}

// TestTemplateEngine_VariableProcessing 测试变量处理功能
func TestTemplateEngine_VariableProcessing(t *testing.T) {
	engine := service.NewTemplateEngine()
	
	// 测试数据
	variables := map[string]interface{}{
		"total_revenue":    125000.75,
		"order_count":      500,
		"growth_rate":      15.25,
		"merchant_name":    "测试商户",
		"report_date":      time.Date(2025, 8, 30, 10, 30, 0, 0, time.UTC),
		"conversion_rate":  3.85,
	}
	
	// 测试模板字符串
	templateStr := `
	报表标题：财务分析报告
	总收入：{{.TotalRevenue}}
	订单数量：{{.OrderCount}} 笔
	增长率：{{.GrowthRate}}
	商户名称：{{.MerchantName}}
	报表日期：{{.ReportDate}}
	转换率：{{.ConversionRate}}
	`
	
	// 处理变量替换
	result := engine.ProcessTemplateVariables(templateStr, variables)
	
	// 验证替换结果
	assert.Contains(t, result, "¥125000.75") // 金额格式化
	assert.Contains(t, result, "500") // 数字格式化
	assert.Contains(t, result, "15.25%") // 百分比格式化
	assert.Contains(t, result, "测试商户") // 字符串保持原样
	assert.Contains(t, result, "2025-08-30") // 日期格式化
	
	t.Logf("变量处理结果:\n%s", result)
}

// TestTemplateEngine_ErrorHandling 测试错误处理
func TestTemplateEngine_ErrorHandling(t *testing.T) {
	engine := service.NewTemplateEngine()
	ctx := context.Background()
	
	// 测试无效的JSON配置
	invalidTemplate := &types.ReportTemplate{
		ID:             1,
		Name:          "无效模板",
		ReportType:    types.ReportTypeFinancial,
		TemplateConfig: []byte(`{"invalid": json}`),
	}
	
	_, err := engine.RenderTemplate(ctx, invalidTemplate, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "解析模板配置失败")
	
	// 测试不支持的报表类型
	unsupportedTemplate := &types.ReportTemplate{
		ID:             1,
		Name:          "不支持的模板",
		ReportType:    "unsupported_type",
		TemplateConfig: []byte(`{}`),
	}
	
	_, err = engine.RenderTemplate(ctx, unsupportedTemplate, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的报表类型")
	
	// 测试类型不匹配的数据
	template := &types.ReportTemplate{
		ID:             1,
		Name:          "类型不匹配",
		ReportType:    types.ReportTypeFinancial,
		TemplateConfig: []byte(`{"currency": "CNY"}`),
	}
	
	// 传入错误的数据类型
	wrongData := &types.MerchantOperationReport{}
	_, err = engine.RenderTemplate(ctx, template, wrongData)
	assert.Error(t, err) // 应该在类型转换时出错
}

// TestTemplateEngine_EdgeCases 测试边界情况
func TestTemplateEngine_EdgeCases(t *testing.T) {
	engine := service.NewTemplateEngine()
	ctx := context.Background()
	
	// 测试空数据
	emptyFinancialData := &types.FinancialReportData{
		TotalRevenue:     &types.Money{Amount: 0, Currency: "CNY"},
		TotalExpenditure: &types.Money{Amount: 0, Currency: "CNY"},
		NetProfit:       &types.Money{Amount: 0, Currency: "CNY"},
		OrderCount:      0,
		MerchantCount:   0,
		CustomerCount:   0,
	}
	
	template := &types.ReportTemplate{
		ID:             1,
		Name:          "空数据测试",
		ReportType:    types.ReportTypeFinancial,
		TemplateConfig: []byte(`{"currency": "CNY"}`),
	}
	
	result, err := engine.RenderTemplate(ctx, template, emptyFinancialData)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	
	// 解析结果验证零值处理
	var renderedData map[string]interface{}
	err = json.Unmarshal(result, &renderedData)
	require.NoError(t, err)
	
	summary, ok := renderedData["summary"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "¥0.00", summary["total_revenue"])
	assert.Equal(t, float64(0), summary["order_count"])
	
	// 测试大数值
	largeFinancialData := &types.FinancialReportData{
		TotalRevenue:     &types.Money{Amount: 9999999999.99, Currency: "CNY"},
		NetProfit:       &types.Money{Amount: 1234567890.12, Currency: "CNY"},
		OrderCount:      1000000,
		MerchantCount:   50000,
		CustomerCount:   2000000,
	}
	
	result, err = engine.RenderTemplate(ctx, template, largeFinancialData)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	
	// 验证大数值的正确格式化
	assert.Contains(t, string(result), "9999999999.99")
	assert.Contains(t, string(result), "1234567890.12")
	
	t.Logf("边界情况测试通过")
}

// TestTemplateEngine_Performance 测试性能
func TestTemplateEngine_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（使用 -short 标志）")
	}
	
	engine := service.NewTemplateEngine()
	ctx := context.Background()
	
	// 创建大型数据集
	financialData := &types.FinancialReportData{
		TotalRevenue:     &types.Money{Amount: 1000000.00, Currency: "CNY"},
		NetProfit:       &types.Money{Amount: 500000.00, Currency: "CNY"},
		OrderCount:      10000,
		MerchantCount:   100,
		CustomerCount:   5000,
		Breakdown: &types.FinancialBreakdown{
			RevenueByMerchant: make([]*types.MerchantRevenue, 100),
			MonthlyTrend:     make([]*types.MonthlyTrend, 12),
		},
	}
	
	// 填充大量数据
	for i := 0; i < 100; i++ {
		financialData.Breakdown.RevenueByMerchant[i] = &types.MerchantRevenue{
			MerchantID:   uint64(i + 1),
			MerchantName: "商户" + string(rune(i+65)),
			Revenue:      &types.Money{Amount: float64(i * 1000), Currency: "CNY"},
			OrderCount:   i * 10,
			Percentage:   float64(i),
		}
	}
	
	for i := 0; i < 12; i++ {
		financialData.Breakdown.MonthlyTrend[i] = &types.MonthlyTrend{
			Month:       time.Date(2025, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC).Format("2006-01"),
			Revenue:     &types.Money{Amount: float64(i * 50000), Currency: "CNY"},
			OrderCount:  i * 500,
		}
	}
	
	template := &types.ReportTemplate{
		ID:             1,
		Name:          "性能测试模板",
		ReportType:    types.ReportTypeFinancial,
		TemplateConfig: []byte(`{"currency": "CNY", "include_trends": true, "include_breakdown": true}`),
	}
	
	// 性能测试
	iterations := 100
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		_, err := engine.RenderTemplate(ctx, template, financialData)
		require.NoError(t, err)
	}
	
	elapsed := time.Since(start)
	avgTime := elapsed / time.Duration(iterations)
	
	t.Logf("性能测试结果:")
	t.Logf("- 总时间: %v", elapsed)
	t.Logf("- 平均时间: %v", avgTime)
	t.Logf("- 每秒处理: %.2f 次", float64(iterations)/elapsed.Seconds())
	
	// 性能断言（平均时间应该小于100ms）
	assert.Less(t, avgTime, 100*time.Millisecond, "模板渲染平均时间应该小于100ms")
}
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// ITemplateEngine 报表模板引擎接口
type ITemplateEngine interface {
	RenderTemplate(ctx context.Context, template *types.ReportTemplate, data interface{}) ([]byte, error)
	ValidateTemplate(ctx context.Context, template *types.ReportTemplate) error
	GetSupportedVariables() map[string]string
}

// TemplateEngine 报表模板引擎实现
type TemplateEngine struct {
	formatters map[string]FormatterFunc
	validators map[string]ValidatorFunc
}

// FormatterFunc 格式化函数类型
type FormatterFunc func(value interface{}) string

// ValidatorFunc 验证函数类型
type ValidatorFunc func(template *types.ReportTemplate) error

// NewTemplateEngine 创建模板引擎实例
func NewTemplateEngine() ITemplateEngine {
	engine := &TemplateEngine{
		formatters: make(map[string]FormatterFunc),
		validators: make(map[string]ValidatorFunc),
	}
	
	// 注册默认格式化器
	engine.registerDefaultFormatters()
	
	// 注册默认验证器
	engine.registerDefaultValidators()
	
	return engine
}

// RenderTemplate 渲染报表模板
func (e *TemplateEngine) RenderTemplate(ctx context.Context, template *types.ReportTemplate, data interface{}) ([]byte, error) {
	g.Log().Info(ctx, "开始渲染报表模板", 
		"template_id", template.ID,
		"template_name", template.Name,
		"report_type", template.ReportType)
	
	// 解析模板配置
	var templateConfig map[string]interface{}
	if err := json.Unmarshal(template.TemplateConfig, &templateConfig); err != nil {
		return nil, fmt.Errorf("解析模板配置失败: %v", err)
	}
	
	// 根据报表类型选择渲染器
	switch template.ReportType {
	case types.ReportTypeFinancial:
		return e.renderFinancialTemplate(ctx, templateConfig, data.(*types.FinancialReportData))
	case types.ReportTypeMerchantOperation:
		return e.renderMerchantOperationTemplate(ctx, templateConfig, data.(*types.MerchantOperationReport))
	case types.ReportTypeCustomerAnalysis:
		return e.renderCustomerAnalysisTemplate(ctx, templateConfig, data.(*types.CustomerAnalysisReport))
	default:
		return nil, fmt.Errorf("不支持的报表类型: %s", template.ReportType)
	}
}

// ValidateTemplate 验证模板配置
func (e *TemplateEngine) ValidateTemplate(ctx context.Context, template *types.ReportTemplate) error {
	if validator, exists := e.validators[string(template.ReportType)]; exists {
		return validator(template)
	}
	return nil
}

// GetSupportedVariables 获取支持的变量列表
func (e *TemplateEngine) GetSupportedVariables() map[string]string {
	return map[string]string{
		"{{.TotalRevenue}}":         "总收入",
		"{{.OrderCount}}":           "订单数量",
		"{{.MerchantCount}}":        "商户数量",
		"{{.CustomerCount}}":        "客户数量",
		"{{.ActiveMerchantCount}}":  "活跃商户数",
		"{{.ActiveCustomerCount}}":  "活跃客户数",
		"{{.RightsConsumed}}":       "权益消耗",
		"{{.NetProfit}}":            "净利润",
		"{{.ReportDate}}":           "报表日期",
		"{{.ReportPeriod}}":         "报表周期",
	}
}

// renderFinancialTemplate 渲染财务报表模板
func (e *TemplateEngine) renderFinancialTemplate(ctx context.Context, config map[string]interface{}, data *types.FinancialReportData) ([]byte, error) {
	// 构建财务报表结构化数据
	reportData := map[string]interface{}{
		"title":         "财务报表",
		"generated_at":  time.Now().Format("2006-01-02 15:04:05"),
		"summary": map[string]interface{}{
			"total_revenue":         e.formatMoney(data.TotalRevenue.Amount),
			"total_expenditure":     e.formatMoney(data.TotalExpenditure.Amount),
			"net_profit":           e.formatMoney(data.NetProfit.Amount),
			"order_count":          data.OrderCount,
			"merchant_count":       data.MerchantCount,
			"customer_count":       data.CustomerCount,
			"active_merchant_count": data.ActiveMerchantCount,
			"active_customer_count": data.ActiveCustomerCount,
			"rights_consumed":      data.RightsConsumed,
			"rights_distributed":   data.RightsDistributed,
			"rights_balance":       data.RightsBalance,
		},
		"breakdown": data.Breakdown,
	}
	
	// 应用模板配置
	if includeTrends, ok := config["include_trends"].(bool); ok && includeTrends {
		if data.Breakdown != nil {
			reportData["trends"] = data.Breakdown.MonthlyTrend
		}
	}
	
	if includeBreakdown, ok := config["include_breakdown"].(bool); ok && includeBreakdown {
		if data.Breakdown != nil {
			reportData["merchant_rankings"] = data.Breakdown.RevenueByMerchant
			reportData["category_breakdown"] = data.Breakdown.RevenueByCategory
		}
	}
	
	// 序列化为JSON格式
	return json.MarshalIndent(reportData, "", "  ")
}

// renderMerchantOperationTemplate 渲染商户运营报表模板
func (e *TemplateEngine) renderMerchantOperationTemplate(ctx context.Context, config map[string]interface{}, data *types.MerchantOperationReport) ([]byte, error) {
	reportData := map[string]interface{}{
		"title":        "商户运营报表",
		"generated_at": time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// 商户排名
	if len(data.MerchantRankings) > 0 {
		topCount := 10 // 默认显示前10名
		if count, ok := config["top_count"].(float64); ok {
			topCount = int(count)
		}
		
		if len(data.MerchantRankings) > topCount {
			reportData["merchant_rankings"] = data.MerchantRankings[:topCount]
		} else {
			reportData["merchant_rankings"] = data.MerchantRankings
		}
	}
	
	// 类别分析
	if len(data.CategoryAnalysis) > 0 {
		reportData["category_analysis"] = data.CategoryAnalysis
	}
	
	// 增长指标
	if data.GrowthMetrics != nil {
		reportData["growth_metrics"] = data.GrowthMetrics
	}
	
	// 业绩趋势
	if includeGrowth, ok := config["include_growth"].(bool); ok && includeGrowth {
		if len(data.PerformanceTrends) > 0 {
			reportData["performance_trends"] = data.PerformanceTrends
		}
	}
	
	return json.MarshalIndent(reportData, "", "  ")
}

// renderCustomerAnalysisTemplate 渲染客户分析报表模板
func (e *TemplateEngine) renderCustomerAnalysisTemplate(ctx context.Context, config map[string]interface{}, data *types.CustomerAnalysisReport) ([]byte, error) {
	reportData := map[string]interface{}{
		"title":        "客户分析报表",
		"generated_at": time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// 用户增长数据
	if len(data.UserGrowth) > 0 {
		reportData["user_growth"] = data.UserGrowth
	}
	
	// 活跃度指标
	if data.ActivityMetrics != nil {
		reportData["activity_metrics"] = data.ActivityMetrics
	}
	
	// 消费行为
	if data.ConsumptionBehavior != nil {
		reportData["consumption_behavior"] = data.ConsumptionBehavior
	}
	
	// 留存分析
	if includeRetention, ok := config["include_retention"].(bool); ok && includeRetention {
		if data.RetentionAnalysis != nil {
			reportData["retention_analysis"] = data.RetentionAnalysis
		}
	}
	
	// 流失分析
	if includeChurn, ok := config["include_churn"].(bool); ok && includeChurn {
		if data.ChurnAnalysis != nil {
			reportData["churn_analysis"] = data.ChurnAnalysis
		}
	}
	
	return json.MarshalIndent(reportData, "", "  ")
}

// registerDefaultFormatters 注册默认格式化器
func (e *TemplateEngine) registerDefaultFormatters() {
	e.formatters["money"] = func(value interface{}) string {
		if amount, ok := value.(float64); ok {
			return fmt.Sprintf("¥%.2f", amount)
		}
		return "¥0.00"
	}
	
	e.formatters["number"] = func(value interface{}) string {
		if num, ok := value.(int); ok {
			return fmt.Sprintf("%,d", num)
		}
		if num, ok := value.(float64); ok {
			return fmt.Sprintf("%.2f", num)
		}
		return "0"
	}
	
	e.formatters["percentage"] = func(value interface{}) string {
		if pct, ok := value.(float64); ok {
			return fmt.Sprintf("%.2f%%", pct)
		}
		return "0.00%"
	}
	
	e.formatters["date"] = func(value interface{}) string {
		if t, ok := value.(time.Time); ok {
			return t.Format("2006-01-02")
		}
		return ""
	}
	
	e.formatters["datetime"] = func(value interface{}) string {
		if t, ok := value.(time.Time); ok {
			return t.Format("2006-01-02 15:04:05")
		}
		return ""
	}
}

// registerDefaultValidators 注册默认验证器
func (e *TemplateEngine) registerDefaultValidators() {
	e.validators["financial"] = func(template *types.ReportTemplate) error {
		var config map[string]interface{}
		if err := json.Unmarshal(template.TemplateConfig, &config); err != nil {
			return fmt.Errorf("财务报表模板配置JSON格式错误: %v", err)
		}
		
		// 验证必需的配置项
		requiredFields := []string{"currency"}
		for _, field := range requiredFields {
			if _, exists := config[field]; !exists {
				return fmt.Errorf("财务报表模板缺少必需配置项: %s", field)
			}
		}
		
		// 验证货币类型
		if currency, ok := config["currency"].(string); ok {
			validCurrencies := []string{"CNY", "USD", "EUR", "JPY"}
			valid := false
			for _, validCurrency := range validCurrencies {
				if currency == validCurrency {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("不支持的货币类型: %s", currency)
			}
		}
		
		return nil
	}
	
	e.validators["merchant_operation"] = func(template *types.ReportTemplate) error {
		var config map[string]interface{}
		if err := json.Unmarshal(template.TemplateConfig, &config); err != nil {
			return fmt.Errorf("商户运营报表模板配置JSON格式错误: %v", err)
		}
		
		// 验证top_count参数
		if topCount, exists := config["top_count"]; exists {
			if count, ok := topCount.(float64); ok {
				if count <= 0 || count > 100 {
					return fmt.Errorf("商户排名数量必须在1-100之间")
				}
			} else {
				return fmt.Errorf("top_count必须为数字类型")
			}
		}
		
		return nil
	}
	
	e.validators["customer_analysis"] = func(template *types.ReportTemplate) error {
		var config map[string]interface{}
		if err := json.Unmarshal(template.TemplateConfig, &config); err != nil {
			return fmt.Errorf("客户分析报表模板配置JSON格式错误: %v", err)
		}
		
		// 验证分析模块配置
		analysisModules := []string{"include_retention", "include_churn"}
		for _, module := range analysisModules {
			if value, exists := config[module]; exists {
				if _, ok := value.(bool); !ok {
					return fmt.Errorf("%s配置项必须为布尔类型", module)
				}
			}
		}
		
		return nil
	}
}

// formatMoney 格式化金额
func (e *TemplateEngine) formatMoney(amount float64) string {
	if formatter, exists := e.formatters["money"]; exists {
		return formatter(amount)
	}
	return fmt.Sprintf("¥%.2f", amount)
}

// formatNumber 格式化数字
func (e *TemplateEngine) formatNumber(num interface{}) string {
	if formatter, exists := e.formatters["number"]; exists {
		return formatter(num)
	}
	return fmt.Sprintf("%v", num)
}

// formatPercentage 格式化百分比
func (e *TemplateEngine) formatPercentage(pct float64) string {
	if formatter, exists := e.formatters["percentage"]; exists {
		return formatter(pct)
	}
	return fmt.Sprintf("%.2f%%", pct)
}

// AddFormatter 添加自定义格式化器
func (e *TemplateEngine) AddFormatter(name string, formatter FormatterFunc) {
	e.formatters[name] = formatter
}

// AddValidator 添加自定义验证器
func (e *TemplateEngine) AddValidator(reportType string, validator ValidatorFunc) {
	e.validators[reportType] = validator
}

// ProcessTemplateVariables 处理模板变量替换
func (e *TemplateEngine) ProcessTemplateVariables(template string, variables map[string]interface{}) string {
	result := template
	
	for key, value := range variables {
		placeholder := "{{." + strings.Title(key) + "}}"
		var replacement string
		
		switch v := value.(type) {
		case float64:
			if strings.Contains(key, "amount") || strings.Contains(key, "revenue") || strings.Contains(key, "profit") {
				replacement = e.formatMoney(v)
			} else if strings.Contains(key, "rate") || strings.Contains(key, "percentage") {
				replacement = e.formatPercentage(v)
			} else {
				replacement = e.formatNumber(v)
			}
		case int:
			replacement = e.formatNumber(v)
		case string:
			replacement = v
		case time.Time:
			replacement = v.Format("2006-01-02 15:04:05")
		default:
			replacement = fmt.Sprintf("%v", v)
		}
		
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	
	return result
}
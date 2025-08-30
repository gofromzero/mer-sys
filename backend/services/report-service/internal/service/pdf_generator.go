package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// IPDFGenerator PDF生成器接口
type IPDFGenerator interface {
	GeneratePDFReport(ctx context.Context, report *types.Report, data interface{}) (string, error)
	CreateHTMLTemplate(reportType types.ReportType, data interface{}) (string, error)
	GetSupportedPDFConverters() []string
}

// PDFGenerator PDF生成器实现
type PDFGenerator struct {
	templateEngine ITemplateEngine
}

// NewPDFGenerator 创建PDF生成器实例
func NewPDFGenerator() IPDFGenerator {
	return &PDFGenerator{
		templateEngine: NewTemplateEngine(),
	}
}

// GeneratePDFReport 生成PDF报表
func (p *PDFGenerator) GeneratePDFReport(ctx context.Context, report *types.Report, data interface{}) (string, error) {
	g.Log().Info(ctx, "开始生成PDF报表", 
		"report_type", report.ReportType,
		"report_uuid", report.UUID)
	
	// 创建HTML内容
	htmlContent, err := p.CreateHTMLTemplate(report.ReportType, data)
	if err != nil {
		return "", fmt.Errorf("创建HTML模板失败: %v", err)
	}
	
	// 确保报表目录存在
	reportDir := p.getReportDir()
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return "", fmt.Errorf("创建报表目录失败: %v", err)
	}
	
	// 生成文件路径
	htmlFilename := fmt.Sprintf("temp_%s_%s_%s.html", 
		report.ReportType, 
		report.UUID,
		time.Now().Format("20060102_150405"))
	htmlPath := filepath.Join(reportDir, htmlFilename)
	
	pdfFilename := fmt.Sprintf("%s_%s_%s.pdf", 
		report.ReportType, 
		report.UUID,
		time.Now().Format("20060102_150405"))
	pdfPath := filepath.Join(reportDir, pdfFilename)
	
	// 保存HTML文件
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return "", fmt.Errorf("保存HTML文件失败: %v", err)
	}
	
	// 尝试转换为PDF
	pdfOutputPath, err := p.convertHTMLToPDF(ctx, htmlPath, pdfPath)
	if err != nil {
		g.Log().Warning(ctx, "PDF转换失败，返回HTML文件", "error", err)
		return htmlPath, nil // 转换失败时返回HTML文件
	}
	
	// 清理临时HTML文件
	os.Remove(htmlPath)
	
	g.Log().Info(ctx, "PDF报表生成成功", "file_path", pdfOutputPath)
	return pdfOutputPath, nil
}

// CreateHTMLTemplate 创建HTML模板
func (p *PDFGenerator) CreateHTMLTemplate(reportType types.ReportType, data interface{}) (string, error) {
	switch reportType {
	case types.ReportTypeFinancial:
		return p.createFinancialHTMLTemplate(data.(*types.FinancialReportData))
	case types.ReportTypeMerchantOperation:
		return p.createMerchantOperationHTMLTemplate(data.(*types.MerchantOperationReport))
	case types.ReportTypeCustomerAnalysis:
		return p.createCustomerAnalysisHTMLTemplate(data.(*types.CustomerAnalysisReport))
	default:
		return "", fmt.Errorf("不支持的报表类型: %s", reportType)
	}
}

// createFinancialHTMLTemplate 创建财务报表HTML模板
func (p *PDFGenerator) createFinancialHTMLTemplate(data *types.FinancialReportData) (string, error) {
	tmplStr := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>财务报表</title>
    <style>
        body { 
            font-family: 'SimHei', sans-serif; 
            margin: 20px; 
            line-height: 1.6; 
        }
        .header { 
            text-align: center; 
            margin-bottom: 30px; 
            border-bottom: 2px solid #333; 
            padding-bottom: 10px; 
        }
        .header h1 { 
            color: #333; 
            margin: 0; 
        }
        .header .date { 
            color: #666; 
            font-size: 14px; 
        }
        .summary-table { 
            width: 100%; 
            border-collapse: collapse; 
            margin-bottom: 30px; 
        }
        .summary-table th, .summary-table td { 
            border: 1px solid #ddd; 
            padding: 12px; 
            text-align: left; 
        }
        .summary-table th { 
            background-color: #f5f5f5; 
            font-weight: bold; 
        }
        .summary-table tr:nth-child(even) { 
            background-color: #f9f9f9; 
        }
        .amount { 
            color: #e74c3c; 
            font-weight: bold; 
        }
        .positive { 
            color: #27ae60; 
        }
        .section-title { 
            background-color: #3498db; 
            color: white; 
            padding: 10px; 
            margin: 20px 0 10px 0; 
        }
        .breakdown-table { 
            width: 100%; 
            border-collapse: collapse; 
            margin-bottom: 20px; 
        }
        .breakdown-table th, .breakdown-table td { 
            border: 1px solid #ddd; 
            padding: 8px; 
            text-align: center; 
        }
        .breakdown-table th { 
            background-color: #ecf0f1; 
        }
        .footer { 
            margin-top: 40px; 
            padding-top: 20px; 
            border-top: 1px solid #ddd; 
            text-align: center; 
            color: #666; 
            font-size: 12px; 
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>财务报表</h1>
        <div class="date">生成时间: {{.GeneratedAt}}</div>
    </div>
    
    <table class="summary-table">
        <tr>
            <th>财务指标</th>
            <th>数值</th>
            <th>说明</th>
        </tr>
        <tr>
            <td>总收入</td>
            <td class="amount positive">¥{{.TotalRevenue}}</td>
            <td>已支付订单总金额</td>
        </tr>
        <tr>
            <td>净利润</td>
            <td class="amount positive">¥{{.NetProfit}}</td>
            <td>总收入减去总支出</td>
        </tr>
        <tr>
            <td>订单总数</td>
            <td>{{.OrderCount}} 笔</td>
            <td>已支付订单数量</td>
        </tr>
        <tr>
            <td>商户总数</td>
            <td>{{.MerchantCount}} 个</td>
            <td>有交易的商户数量</td>
        </tr>
        <tr>
            <td>客户总数</td>
            <td>{{.CustomerCount}} 个</td>
            <td>有交易的客户数量</td>
        </tr>
        <tr>
            <td>活跃商户数</td>
            <td>{{.ActiveMerchantCount}} 个</td>
            <td>近30天有交易的商户</td>
        </tr>
        <tr>
            <td>权益消耗</td>
            <td>{{.RightsConsumed}} 份</td>
            <td>用户使用的权益总量</td>
        </tr>
    </table>
    
    {{if .MerchantRankings}}
    <div class="section-title">商户收入排行榜 (前10名)</div>
    <table class="breakdown-table">
        <tr>
            <th>排名</th>
            <th>商户名称</th>
            <th>收入金额</th>
            <th>订单数量</th>
            <th>占比</th>
        </tr>
        {{range $index, $merchant := .MerchantRankings}}
        <tr>
            <td>{{add $index 1}}</td>
            <td>{{$merchant.MerchantName}}</td>
            <td class="amount">¥{{$merchant.Revenue}}</td>
            <td>{{$merchant.OrderCount}}</td>
            <td>{{$merchant.Percentage}}%</td>
        </tr>
        {{end}}
    </table>
    {{end}}
    
    {{if .MonthlyTrend}}
    <div class="section-title">月度趋势</div>
    <table class="breakdown-table">
        <tr>
            <th>月份</th>
            <th>收入</th>
            <th>订单数量</th>
            <th>权益消耗</th>
        </tr>
        {{range .MonthlyTrend}}
        <tr>
            <td>{{.Month}}</td>
            <td class="amount">¥{{.Revenue}}</td>
            <td>{{.OrderCount}}</td>
            <td>{{.RightsConsumed}}</td>
        </tr>
        {{end}}
    </table>
    {{end}}
    
    <div class="footer">
        <p>本报表由MER系统自动生成 | 数据仅供内部使用</p>
    </div>
</body>
</html>
`
	
	// 创建模板
	tmpl, err := template.New("financial").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("解析HTML模板失败: %v", err)
	}
	
	// 准备模板数据
	templateData := map[string]interface{}{
		"GeneratedAt":            time.Now().Format("2006-01-02 15:04:05"),
		"TotalRevenue":           fmt.Sprintf("%.2f", data.TotalRevenue.Amount),
		"NetProfit":              fmt.Sprintf("%.2f", data.NetProfit.Amount),
		"OrderCount":             data.OrderCount,
		"MerchantCount":          data.MerchantCount,
		"CustomerCount":          data.CustomerCount,
		"ActiveMerchantCount":    data.ActiveMerchantCount,
		"RightsConsumed":         data.RightsConsumed,
	}
	
	// 添加商户排行数据
	if data.Breakdown != nil && len(data.Breakdown.RevenueByMerchant) > 0 {
		merchants := data.Breakdown.RevenueByMerchant
		if len(merchants) > 10 {
			merchants = merchants[:10] // 只显示前10名
		}
		
		var merchantData []map[string]interface{}
		for _, merchant := range merchants {
			merchantData = append(merchantData, map[string]interface{}{
				"MerchantName": merchant.MerchantName,
				"Revenue":      fmt.Sprintf("%.2f", merchant.Revenue.Amount),
				"OrderCount":   merchant.OrderCount,
				"Percentage":   fmt.Sprintf("%.2f", merchant.Percentage),
			})
		}
		templateData["MerchantRankings"] = merchantData
	}
	
	// 添加月度趋势数据
	if data.Breakdown != nil && len(data.Breakdown.MonthlyTrend) > 0 {
		var trendData []map[string]interface{}
		for _, trend := range data.Breakdown.MonthlyTrend {
			trendData = append(trendData, map[string]interface{}{
				"Month":          trend.Month,
				"Revenue":        fmt.Sprintf("%.2f", trend.Revenue.Amount),
				"OrderCount":     trend.OrderCount,
				"RightsConsumed": trend.RightsConsumed,
			})
		}
		templateData["MonthlyTrend"] = trendData
	}
	
	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("渲染HTML模板失败: %v", err)
	}
	
	return buf.String(), nil
}

// createMerchantOperationHTMLTemplate 创建商户运营报表HTML模板
func (p *PDFGenerator) createMerchantOperationHTMLTemplate(data *types.MerchantOperationReport) (string, error) {
	tmplStr := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>商户运营报表</title>
    <style>
        body { 
            font-family: 'SimHei', sans-serif; 
            margin: 20px; 
            line-height: 1.6; 
        }
        .header { 
            text-align: center; 
            margin-bottom: 30px; 
            border-bottom: 2px solid #333; 
            padding-bottom: 10px; 
        }
        .section-title { 
            background-color: #e74c3c; 
            color: white; 
            padding: 10px; 
            margin: 20px 0 10px 0; 
        }
        .ranking-table { 
            width: 100%; 
            border-collapse: collapse; 
            margin-bottom: 20px; 
        }
        .ranking-table th, .ranking-table td { 
            border: 1px solid #ddd; 
            padding: 8px; 
            text-align: center; 
        }
        .ranking-table th { 
            background-color: #f8f9fa; 
        }
        .amount { 
            color: #e74c3c; 
            font-weight: bold; 
        }
        .growth-positive { 
            color: #27ae60; 
        }
        .growth-negative { 
            color: #e74c3c; 
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>商户运营报表</h1>
        <div class="date">生成时间: {{.GeneratedAt}}</div>
    </div>
    
    {{if .MerchantRankings}}
    <div class="section-title">商户业绩排行榜</div>
    <table class="ranking-table">
        <tr>
            <th>排名</th>
            <th>商户名称</th>
            <th>总收入</th>
            <th>订单数量</th>
            <th>客户数量</th>
            <th>客单价</th>
            <th>增长率</th>
        </tr>
        {{range .MerchantRankings}}
        <tr>
            <td>{{.Rank}}</td>
            <td>{{.MerchantName}}</td>
            <td class="amount">¥{{.Revenue}}</td>
            <td>{{.OrderCount}}</td>
            <td>{{.CustomerCount}}</td>
            <td class="amount">¥{{.AvgOrderValue}}</td>
            <td class="{{if gt .GrowthRate 0}}growth-positive{{else}}growth-negative{{end}}">{{.GrowthRate}}%</td>
        </tr>
        {{end}}
    </table>
    {{end}}
    
    <div class="footer">
        <p>本报表由MER系统自动生成</p>
    </div>
</body>
</html>
`
	
	// 类似财务报表的模板处理逻辑...
	tmpl, err := template.New("merchant").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("解析商户运营HTML模板失败: %v", err)
	}
	
	templateData := map[string]interface{}{
		"GeneratedAt": time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// 处理商户排行数据
	if len(data.MerchantRankings) > 0 {
		var rankingData []map[string]interface{}
		for _, ranking := range data.MerchantRankings {
			rankingData = append(rankingData, map[string]interface{}{
				"Rank":         ranking.Rank,
				"MerchantName": ranking.MerchantName,
				"Revenue":      fmt.Sprintf("%.2f", ranking.TotalRevenue.Amount),
				"OrderCount":   ranking.OrderCount,
				"CustomerCount": ranking.CustomerCount,
				"AvgOrderValue": fmt.Sprintf("%.2f", ranking.AverageOrderValue.Amount),
				"GrowthRate":   fmt.Sprintf("%.2f", ranking.GrowthRate),
			})
		}
		templateData["MerchantRankings"] = rankingData
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("渲染商户运营HTML模板失败: %v", err)
	}
	
	return buf.String(), nil
}

// createCustomerAnalysisHTMLTemplate 创建客户分析报表HTML模板
func (p *PDFGenerator) createCustomerAnalysisHTMLTemplate(data *types.CustomerAnalysisReport) (string, error) {
	// 简化的客户分析HTML模板
	tmplStr := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>客户分析报表</title>
    <style>
        body { 
            font-family: 'SimHei', sans-serif; 
            margin: 20px; 
            line-height: 1.6; 
        }
        .header { 
            text-align: center; 
            margin-bottom: 30px; 
            border-bottom: 2px solid #333; 
            padding-bottom: 10px; 
        }
        .metrics-table { 
            width: 100%; 
            border-collapse: collapse; 
            margin-bottom: 20px; 
        }
        .metrics-table th, .metrics-table td { 
            border: 1px solid #ddd; 
            padding: 12px; 
            text-align: left; 
        }
        .metrics-table th { 
            background-color: #f5f5f5; 
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>客户分析报表</h1>
        <div class="date">生成时间: {{.GeneratedAt}}</div>
    </div>
    
    {{if .ActivityMetrics}}
    <h2>活跃度指标</h2>
    <table class="metrics-table">
        <tr>
            <th>指标</th>
            <th>数值</th>
        </tr>
        <tr>
            <td>日活跃用户(DAU)</td>
            <td>{{.ActivityMetrics.DAU}}</td>
        </tr>
        <tr>
            <td>周活跃用户(WAU)</td>
            <td>{{.ActivityMetrics.WAU}}</td>
        </tr>
        <tr>
            <td>月活跃用户(MAU)</td>
            <td>{{.ActivityMetrics.MAU}}</td>
        </tr>
    </table>
    {{end}}
    
    <div class="footer">
        <p>本报表由MER系统自动生成</p>
    </div>
</body>
</html>
`
	
	tmpl, err := template.New("customer").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("解析客户分析HTML模板失败: %v", err)
	}
	
	templateData := map[string]interface{}{
		"GeneratedAt":     time.Now().Format("2006-01-02 15:04:05"),
		"ActivityMetrics": data.ActivityMetrics,
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("渲染客户分析HTML模板失败: %v", err)
	}
	
	return buf.String(), nil
}

// getReportDir 获取报表存储目录
func (p *PDFGenerator) getReportDir() string {
	baseDir := g.Cfg().MustGet(context.Background(), "report.storage_dir", "/tmp/reports").String()
	return baseDir
}

// convertHTMLToPDF 将HTML转换为PDF
func (p *PDFGenerator) convertHTMLToPDF(ctx context.Context, htmlPath, pdfPath string) (string, error) {
	g.Log().Debug(ctx, "尝试将HTML转换为PDF", 
		"html_path", htmlPath, 
		"pdf_path", pdfPath)
	
	// 尝试多种PDF转换方案
	converters := []struct {
		name string
		convert func(string, string) error
	}{
		{"wkhtmltopdf", p.convertWithWkhtml},
		{"chrome/chromium", p.convertWithChrome},
		{"phantomjs", p.convertWithPhantom},
	}
	
	for _, converter := range converters {
		g.Log().Debug(ctx, "尝试PDF转换器", "converter", converter.name)
		if err := converter.convert(htmlPath, pdfPath); err != nil {
			g.Log().Warning(ctx, "PDF转换器失败", 
				"converter", converter.name, 
				"error", err)
			continue
		}
		
		// 验证PDF文件是否生成成功
		if _, err := os.Stat(pdfPath); err == nil {
			g.Log().Info(ctx, "PDF转换成功", 
				"converter", converter.name,
				"pdf_path", pdfPath)
			return pdfPath, nil
		}
	}
	
	return "", fmt.Errorf("所有PDF转换器均失败")
}

// convertWithWkhtml 使用wkhtmltopdf转换
func (p *PDFGenerator) convertWithWkhtml(htmlPath, pdfPath string) error {
	// 检查wkhtmltopdf是否可用
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		return fmt.Errorf("wkhtmltopdf未安装: %v", err)
	}
	
	// 执行转换命令
	cmd := exec.Command("wkhtmltopdf", 
		"--page-size", "A4",
		"--encoding", "UTF-8",
		"--margin-top", "10mm",
		"--margin-right", "10mm", 
		"--margin-bottom", "10mm",
		"--margin-left", "10mm",
		htmlPath, pdfPath)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wkhtmltopdf执行失败: %v, stderr: %s", err, stderr.String())
	}
	
	return nil
}

// convertWithChrome 使用Chrome/Chromium转换
func (p *PDFGenerator) convertWithChrome(htmlPath, pdfPath string) error {
	// 尝试常见的Chrome路径
	chromePaths := []string{
		"google-chrome",
		"chromium-browser", 
		"chromium",
		"chrome",
		"/usr/bin/google-chrome",
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
	}
	
	var chromeCmd string
	for _, path := range chromePaths {
		if _, err := exec.LookPath(path); err == nil {
			chromeCmd = path
			break
		}
	}
	
	if chromeCmd == "" {
		return fmt.Errorf("未找到Chrome/Chromium浏览器")
	}
	
	// 执行Chrome转换命令
	cmd := exec.Command(chromeCmd,
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--print-to-pdf="+pdfPath,
		"file://"+htmlPath)
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Chrome转换失败: %v, stderr: %s", err, stderr.String())
	}
	
	return nil
}

// convertWithPhantom 使用PhantomJS转换
func (p *PDFGenerator) convertWithPhantom(htmlPath, pdfPath string) error {
	// 检查phantomjs是否可用
	if _, err := exec.LookPath("phantomjs"); err != nil {
		return fmt.Errorf("phantomjs未安装: %v", err)
	}
	
	// 创建临时的JS脚本
	jsScript := fmt.Sprintf(`
		var page = require('webpage').create();
		page.paperSize = {
			format: 'A4',
			margin: {
				top: '10mm',
				left: '10mm',
				right: '10mm',
				bottom: '10mm'
			}
		};
		
		page.open('%s', function() {
			setTimeout(function() {
				page.render('%s');
				phantom.exit();
			}, 1000);
		});
	`, "file://"+htmlPath, pdfPath)
	
	scriptPath := filepath.Join(filepath.Dir(pdfPath), "convert_script.js")
	if err := os.WriteFile(scriptPath, []byte(jsScript), 0644); err != nil {
		return fmt.Errorf("创建转换脚本失败: %v", err)
	}
	defer os.Remove(scriptPath)
	
	// 执行PhantomJS转换
	cmd := exec.Command("phantomjs", scriptPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PhantomJS转换失败: %v, stderr: %s", err, stderr.String())
	}
	
	return nil
}

// GetSupportedPDFConverters 获取支持的PDF转换器列表
func (p *PDFGenerator) GetSupportedPDFConverters() []string {
	var supported []string
	
	converters := []struct {
		name string
		check func() bool
	}{
		{"wkhtmltopdf", func() bool {
			_, err := exec.LookPath("wkhtmltopdf")
			return err == nil
		}},
		{"chrome", func() bool {
			chromePaths := []string{"google-chrome", "chromium-browser", "chromium", "chrome"}
			for _, path := range chromePaths {
				if _, err := exec.LookPath(path); err == nil {
					return true
				}
			}
			return false
		}},
		{"phantomjs", func() bool {
			_, err := exec.LookPath("phantomjs")
			return err == nil
		}},
	}
	
	for _, converter := range converters {
		if converter.check() {
			supported = append(supported, converter.name)
		}
	}
	
	return supported
}
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ReportSystemIntegrationTestSuite 报表系统集成测试套件
type ReportSystemIntegrationTestSuite struct {
	suite.Suite
	ctx                   context.Context
	generator            service.IReportGeneratorService
	templateEngine       service.ITemplateEngine
	cacheManager         service.ICacheManager
	performanceOptimizer service.IPerformanceOptimizer
	advancedCache        service.IAdvancedCacheStrategy
	notificationService  service.INotificationService
	historyService       service.IHistoryService
	schedulerService     service.ISchedulerService
}

func TestReportSystemIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}
	
	// 检查是否有必要的环境变量或配置
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("需要设置 INTEGRATION_TEST 环境变量来运行集成测试")
	}

	suite.Run(t, new(ReportSystemIntegrationTestSuite))
}

func (s *ReportSystemIntegrationTestSuite) SetupSuite() {
	s.ctx = context.WithValue(context.Background(), "tenant_id", uint64(1))
	
	// 初始化所有服务
	s.generator = service.NewReportGeneratorService()
	s.templateEngine = service.NewTemplateEngine()
	s.cacheManager = service.NewCacheManager()
	s.performanceOptimizer = service.NewPerformanceOptimizer()
	s.advancedCache = service.NewAdvancedCacheStrategy(s.cacheManager)
	s.notificationService = service.NewNotificationService()
	s.historyService = service.NewHistoryService()
	s.schedulerService = service.NewSchedulerService()
}

func (s *ReportSystemIntegrationTestSuite) TearDownSuite() {
	// 清理资源
	if s.schedulerService != nil {
		_ = s.schedulerService.Stop(s.ctx)
	}
}

func (s *ReportSystemIntegrationTestSuite) TestCompleteReportGenerationFlow() {
	s.T().Log("测试完整的报表生成流程")

	// 1. 创建报表请求
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Now().AddDate(0, -1, 0),
		EndDate:    time.Now(),
		FileFormat: types.FileFormatExcel,
		TenantID:   1,
	}

	// 2. 性能优化查询
	optimizedQuery, err := s.performanceOptimizer.OptimizeDataQuery(s.ctx, req)
	s.Require().NoError(err, "查询优化不应失败")
	s.Require().NotNil(optimizedQuery, "优化查询不能为空")
	s.Assert().Greater(optimizedQuery.ChunkSize, 0, "分块大小应大于0")

	// 3. 检查缓存
	cachedReport, err := s.cacheManager.GetReportFromCache(s.ctx, req)
	if err == nil && cachedReport != nil {
		s.T().Log("从缓存中获取到报表，跳过生成")
		return
	}

	// 4. 生成报表
	report, err := s.generator.GenerateReport(s.ctx, req)
	s.Require().NoError(err, "报表生成不应失败")
	s.Require().NotNil(report, "生成的报表不能为空")
	s.Assert().Equal(types.ReportStatusCompleted, report.Status, "报表状态应为已完成")

	// 5. 缓存报表
	if s.cacheManager.ShouldUseCache(req) {
		err = s.cacheManager.CacheReport(s.ctx, req, report)
		s.Assert().NoError(err, "报表缓存不应失败")
	}

	// 6. 记录历史
	err = s.historyService.RecordReportGeneration(s.ctx, report, 5000, report.FilePath, 1024*1024)
	s.Assert().NoError(err, "记录历史不应失败")

	// 7. 发送通知（如果配置了）
	notificationReq := &service.NotificationRequest{
		Type:       service.NotificationTypeEmail,
		Recipients: []string{"test@example.com"},
		Subject:    "报表生成完成",
		Content:    s.notificationService.GetEmailTemplate("report_ready", map[string]interface{}{
			"task_name":     "财务报表",
			"report_type":   "财务报表",
			"generated_at":  time.Now().Format("2006-01-02 15:04:05"),
			"file_format":   "Excel",
		}),
		Priority: service.NotificationPriorityNormal,
	}

	err = s.notificationService.SendNotification(s.ctx, notificationReq)
	s.Assert().NoError(err, "发送通知不应失败")

	s.T().Logf("报表生成完成: ID=%d, 文件路径=%s", report.ID, report.FilePath)
}

func (s *ReportSystemIntegrationTestSuite) TestTemplateBasedReportGeneration() {
	s.T().Log("测试基于模板的报表生成")

	// 1. 创建报表模板
	template := &types.ReportTemplate{
		TenantID:     1,
		Name:         "月度财务报表模板",
		ReportType:   types.ReportTypeFinancial,
		FileFormat:   types.FileFormatPDF,
		Enabled:      true,
	}

	// 模拟模板配置
	templateConfig := map[string]interface{}{
		"include_charts": true,
		"currency":       "CNY",
		"precision":      2,
	}

	configBytes, err := service.JSONMarshal(templateConfig)
	s.Require().NoError(err)
	template.TemplateConfig = configBytes

	// 2. 验证模板
	err = s.templateEngine.ValidateTemplate(s.ctx, template)
	s.Assert().NoError(err, "模板验证不应失败")

	// 3. 使用模板渲染报表
	testData := map[string]interface{}{
		"tenant_name":    "测试租户",
		"report_period":  "2024年1月",
		"total_revenue":  1000000.50,
		"total_expense":  750000.25,
		"profit_margin":  25.5,
		"transaction_count": 1250,
	}

	renderedContent, err := s.templateEngine.RenderTemplate(s.ctx, template, testData)
	s.Require().NoError(err, "模板渲染不应失败")
	s.Assert().NotEmpty(renderedContent, "渲染内容不能为空")

	s.T().Logf("模板渲染完成，内容长度: %d 字节", len(renderedContent))
}

func (s *ReportSystemIntegrationTestSuite) TestAdvancedCacheIntegration() {
	s.T().Log("测试高级缓存策略集成")

	// 1. 监控缓存性能
	initialMetrics, err := s.advancedCache.MonitorCachePerformance(s.ctx)
	s.Require().NoError(err, "缓存性能监控不应失败")
	s.Require().NotNil(initialMetrics, "初始性能指标不能为空")

	// 2. 执行一些缓存操作来生成数据
	testData := make([]byte, 5*1024) // 5KB测试数据
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// 压缩测试
	compressed, err := s.advancedCache.CompressCache(s.ctx, testData)
	s.Assert().NoError(err, "缓存压缩不应失败")
	s.Assert().Less(t, len(compressed), len(testData), "压缩后大小应减少")

	// 解压测试
	decompressed, err := s.advancedCache.DecompressCache(s.ctx, compressed)
	s.Assert().NoError(err, "缓存解压不应失败")
	s.Assert().GreaterOrEqual(t, len(decompressed), len(compressed), "解压后大小应恢复")

	// 3. 分层缓存管理
	err = s.advancedCache.ManageHierarchicalCache(s.ctx)
	s.Assert().NoError(err, "分层缓存管理不应失败")

	// 4. 智能预加载
	err = s.advancedCache.PredictivePreload(s.ctx, 12345)
	s.Assert().NoError(err, "智能预加载不应失败")

	// 5. 最终性能监控
	finalMetrics, err := s.advancedCache.MonitorCachePerformance(s.ctx)
	s.Require().NoError(err, "最终性能监控不应失败")
	s.Assert().GreaterOrEqual(t, finalMetrics.CompressionRatio, 0.0, "压缩率应有效")

	// 6. 动态调整缓存策略
	err = s.advancedCache.AdjustCacheStrategy(s.ctx, finalMetrics)
	s.Assert().NoError(err, "缓存策略调整不应失败")

	s.T().Logf("缓存性能: 命中率=%.2f%%, 压缩率=%.2f", 
		finalMetrics.HitRate, finalMetrics.CompressionRatio*100)
}

func (s *ReportSystemIntegrationTestSuite) TestSchedulerIntegration() {
	s.T().Log("测试调度器集成")

	// 1. 启动调度器
	err := s.schedulerService.Start(s.ctx)
	s.Require().NoError(err, "调度器启动不应失败")

	// 确保在测试结束时停止调度器
	defer func() {
		err := s.schedulerService.Stop(s.ctx)
		s.Assert().NoError(err, "调度器停止不应失败")
	}()

	// 2. 处理待执行任务（这里主要测试方法调用不出错）
	err = s.schedulerService.ProcessPendingJobs(s.ctx)
	s.Assert().NoError(err, "处理待执行任务不应失败")

	// 3. 模板任务调度（这里主要测试方法调用不出错）
	err = s.schedulerService.ScheduleTemplateJobs(s.ctx)
	s.Assert().NoError(err, "模板任务调度不应失败")

	s.T().Log("调度器集成测试完成")
}

func (s *ReportSystemIntegrationTestSuite) TestNotificationIntegration() {
	s.T().Log("测试通知服务集成")

	// 测试不同类型的通知
	notifications := []struct {
		name string
		req  *service.NotificationRequest
	}{
		{
			name: "邮件通知",
			req: &service.NotificationRequest{
				Type:       service.NotificationTypeEmail,
				Recipients: []string{"test@example.com"},
				Subject:    "测试邮件",
				Content:    "这是一个集成测试邮件",
				Priority:   service.NotificationPriorityNormal,
			},
		},
		{
			name: "短信通知",
			req: &service.NotificationRequest{
				Type:       service.NotificationTypeSMS,
				Recipients: []string{"13800138000"},
				Content:    "测试短信内容",
				Priority:   service.NotificationPriorityHigh,
			},
		},
		{
			name: "应用内通知",
			req: &service.NotificationRequest{
				Type:       service.NotificationTypeInApp,
				Recipients: []string{"user123"},
				Subject:    "应用内通知",
				Content:    "这是一个应用内通知",
				Priority:   service.NotificationPriorityLow,
			},
		},
	}

	for _, notification := range notifications {
		s.T().Run(notification.name, func(t *testing.T) {
			err := s.notificationService.SendNotification(s.ctx, notification.req)
			assert.NoError(t, err, "发送通知不应失败")
		})
	}

	// 测试批量通知
	batchReqs := make([]*service.NotificationRequest, 0)
	for _, notification := range notifications {
		batchReqs = append(batchReqs, notification.req)
	}

	responses, err := s.notificationService.SendBatchNotifications(s.ctx, batchReqs)
	s.Assert().NoError(err, "批量发送通知不应失败")
	s.Assert().Equal(len(batchReqs), len(responses), "响应数量应与请求数量一致")

	s.T().Logf("批量通知发送完成，处理了 %d 个通知", len(responses))
}

func (s *ReportSystemIntegrationTestSuite) TestHistoryServiceIntegration() {
	s.T().Log("测试历史服务集成")

	// 1. 创建测试报表
	report := &types.Report{
		ID:          12345,
		TenantID:    1,
		ReportType:  types.ReportTypeFinancial,
		PeriodType:  types.PeriodTypeMonthly,
		StartDate:   time.Now().AddDate(0, -1, 0),
		EndDate:     time.Now(),
		Status:      types.ReportStatusCompleted,
		FileFormat:  types.FileFormatExcel,
		FilePath:    "/tmp/reports/financial_202401.xlsx",
		GeneratedBy: 1,
		GeneratedAt: time.Now(),
	}

	// 2. 记录报表生成历史
	err := s.historyService.RecordReportGeneration(s.ctx, report, 8500, report.FilePath, 2*1024*1024)
	s.Assert().NoError(err, "记录报表历史不应失败")

	// 3. 获取历史列表
	historyReq := &service.ReportHistoryRequest{
		ReportType: types.ReportTypeFinancial,
		Page:       1,
		PageSize:   10,
	}

	histories, total, err := s.historyService.GetReportHistory(s.ctx, historyReq)
	s.Assert().NoError(err, "获取报表历史不应失败")
	s.Assert().NotNil(histories, "历史记录不能为空")
	s.Assert().GreaterOrEqual(t, total, 0, "总数应大于等于0")

	// 4. 获取统计信息
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	stats, err := s.historyService.GetReportStats(s.ctx, startDate, endDate)
	s.Assert().NoError(err, "获取统计信息不应失败")
	s.Assert().NotNil(stats, "统计信息不能为空")
	s.Assert().GreaterOrEqual(t, stats.TotalReports, int64(0), "报表总数应大于等于0")

	// 5. 获取用户报表历史
	userHistories, err := s.historyService.GetUserReportHistory(s.ctx, 1, 5)
	s.Assert().NoError(err, "获取用户历史不应失败")
	s.Assert().NotNil(userHistories, "用户历史不能为空")

	s.T().Logf("历史服务测试完成，总报表数: %d, 成功率: %.2f%%", 
		stats.TotalReports, stats.SuccessRate)
}

func (s *ReportSystemIntegrationTestSuite) TestEndToEndWorkflow() {
	s.T().Log("测试端到端工作流")

	startTime := time.Now()

	// 1. 性能优化器准备
	err := s.performanceOptimizer.EnableParallelProcessing(s.ctx, 4)
	s.Require().NoError(err)

	// 2. 创建并生成报表
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeMerchantOperation,
		PeriodType: types.PeriodTypeWeekly,
		StartDate:  time.Now().AddDate(0, 0, -7),
		EndDate:    time.Now(),
		FileFormat: types.FileFormatJSON,
		TenantID:   1,
	}

	// 3. 优化查询
	optimizedQuery, err := s.performanceOptimizer.OptimizeDataQuery(s.ctx, req)
	s.Require().NoError(err)

	// 4. 检查缓存
	cachedReport, err := s.cacheManager.GetReportFromCache(s.ctx, req)
	if err != nil && cachedReport == nil {
		// 5. 生成新报表
		report, err := s.generator.GenerateReport(s.ctx, req)
		s.Require().NoError(err)
		s.Require().NotNil(report)

		// 6. 缓存报表
		if s.cacheManager.ShouldUseCache(req) {
			err = s.cacheManager.CacheReport(s.ctx, req, report)
			s.Assert().NoError(err)
		}

		// 7. 记录历史
		executionTime := time.Since(startTime).Milliseconds()
		err = s.historyService.RecordReportGeneration(s.ctx, report, executionTime, report.FilePath, 512*1024)
		s.Assert().NoError(err)

		// 8. 发送完成通知
		notificationReq := &service.NotificationRequest{
			Type:       service.NotificationTypeEmail,
			Recipients: []string{"admin@example.com"},
			Subject:    "报表生成完成",
			Content:    s.notificationService.GetEmailTemplate("report_ready", map[string]interface{}{
				"task_name":     "商户运营报表",
				"report_type":   "商户运营",
				"generated_at":  report.GeneratedAt.Format("2006-01-02 15:04:05"),
				"file_format":   string(report.FileFormat),
			}),
			Priority: service.NotificationPriorityNormal,
		}

		err = s.notificationService.SendNotification(s.ctx, notificationReq)
		s.Assert().NoError(err)

		s.T().Logf("端到端工作流完成，报表ID: %d, 总耗时: %v", report.ID, time.Since(startTime))
	} else {
		s.T().Log("使用缓存报表，工作流完成")
	}

	// 9. 性能监控
	metrics := s.performanceOptimizer.GetPerformanceMetrics(s.ctx)
	s.Assert().NotNil(metrics)

	// 10. 缓存性能监控
	cacheMetrics, err := s.advancedCache.MonitorCachePerformance(s.ctx)
	s.Assert().NoError(err)
	s.Assert().NotNil(cacheMetrics)

	s.T().Logf("端到端测试完成，内存使用: %d MB, 缓存命中率: %.2f%%", 
		metrics.MemoryUsage/1024/1024, cacheMetrics.HitRate)
}

// 辅助函数

func (s *ReportSystemIntegrationTestSuite) createTestRepository() repository.IReportRepository {
	// 这里应该返回一个测试用的Repository实例
	// 在真实的集成测试中，这应该连接到测试数据库
	return repository.NewReportRepository()
}

func (s *ReportSystemIntegrationTestSuite) setupTestData() error {
	// 设置测试数据
	// 在真实的集成测试中，这里应该创建测试用的数据
	return nil
}

func (s *ReportSystemIntegrationTestSuite) cleanupTestData() error {
	// 清理测试数据
	return nil
}

// 独立的集成测试函数

func TestReportSystemHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	ctx := context.Background()

	// 测试各个服务的基本功能
	services := map[string]func() error{
		"ReportGenerator": func() error {
			generator := service.NewReportGeneratorService()
			return generator.HealthCheck(ctx)
		},
		"CacheManager": func() error {
			cacheManager := service.NewCacheManager()
			_, err := cacheManager.GetCacheStats(ctx)
			return err
		},
		"PerformanceOptimizer": func() error {
			optimizer := service.NewPerformanceOptimizer()
			metrics := optimizer.GetPerformanceMetrics(ctx)
			if metrics == nil {
				return assert.AnError
			}
			return nil
		},
		"NotificationService": func() error {
			notification := service.NewNotificationService()
			return notification.ValidateEmailConfig()
		},
	}

	for serviceName, healthCheck := range services {
		t.Run(serviceName, func(t *testing.T) {
			err := healthCheck()
			assert.NoError(t, err, fmt.Sprintf("服务%s健康检查失败", serviceName))
		})
	}
}
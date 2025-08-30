package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
)

// TestReportCacheIntegration 测试报表生成服务的缓存集成
func TestReportCacheIntegration(t *testing.T) {
	// 跳过集成测试，除非设置了环境变量
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", uint64(1))
	ctx = context.WithValue(ctx, "user_id", uint64(1))

	// 创建报表生成服务
	reportService := service.NewReportGeneratorService()

	t.Run("缓存键生成测试", func(t *testing.T) {
		cacheManager := service.NewCacheManager()
		
		req := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Now().AddDate(0, -1, 0),
			EndDate:    time.Now().AddDate(0, 0, -1),
			FileFormat: types.FileFormatJSON,
		}

		// 测试缓存键生成
		cacheKey1 := cacheManager.GetCacheKey(req)
		cacheKey2 := cacheManager.GetCacheKey(req)
		
		assert.NotEmpty(t, cacheKey1, "缓存键不能为空")
		assert.Equal(t, cacheKey1, cacheKey2, "相同请求应该生成相同的缓存键")
		assert.Contains(t, cacheKey1, "report_cache:", "缓存键应该包含前缀")
	})

	t.Run("缓存使用条件测试", func(t *testing.T) {
		cacheManager := service.NewCacheManager()
		
		// 测试正常的缓存条件
		validReq := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Now().AddDate(0, -1, 0),
			EndDate:    time.Now().AddDate(0, 0, -2), // 2天前，避免实时数据
			FileFormat: types.FileFormatJSON,
		}
		assert.True(t, cacheManager.ShouldUseCache(validReq), "历史数据应该使用缓存")

		// 测试实时数据不缓存
		realtimeReq := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Now().AddDate(0, 0, -1),
			EndDate:    time.Now(), // 当前时间，实时数据
			FileFormat: types.FileFormatJSON,
		}
		assert.False(t, cacheManager.ShouldUseCache(realtimeReq), "实时数据不应该使用缓存")

		// 测试时间范围太长不缓存
		longRangeReq := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeYearly,
			StartDate:  time.Now().AddDate(-2, 0, 0),
			EndDate:    time.Now().AddDate(0, 0, -1),
			FileFormat: types.FileFormatJSON,
		}
		assert.False(t, cacheManager.ShouldUseCache(longRangeReq), "时间范围太长不应该使用缓存")
	})

	t.Run("缓存统计信息测试", func(t *testing.T) {
		stats, err := reportService.GetCacheStats(ctx)
		
		assert.NoError(t, err, "获取缓存统计不应该出错")
		assert.NotNil(t, stats, "缓存统计不应该为空")
		assert.Contains(t, stats, "cache_enabled", "缓存统计应该包含启用状态")
		assert.Contains(t, stats, "ttl_settings", "缓存统计应该包含TTL设置")
	})

	t.Run("缓存清理测试", func(t *testing.T) {
		err := reportService.CleanupCache(ctx)
		assert.NoError(t, err, "缓存清理不应该出错")
	})

	t.Run("缓存预热测试", func(t *testing.T) {
		err := reportService.WarmupCache(ctx, types.ReportTypeFinancial)
		assert.NoError(t, err, "财务报表缓存预热不应该出错")

		err = reportService.WarmupCache(ctx, types.ReportTypeMerchantOperation)
		assert.NoError(t, err, "商户运营报表缓存预热不应该出错")

		err = reportService.WarmupCache(ctx, types.ReportTypeCustomerAnalysis)
		assert.NoError(t, err, "客户分析报表缓存预热不应该出错")
	})
}

// TestReportCacheLifecycle 测试报表缓存生命周期
func TestReportCacheLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", uint64(1))
	ctx = context.WithValue(ctx, "user_id", uint64(1))

	cacheManager := service.NewCacheManager()

	t.Run("缓存存储和检索测试", func(t *testing.T) {
		req := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Now().AddDate(0, -1, 0),
			EndDate:    time.Now().AddDate(0, 0, -2),
			FileFormat: types.FileFormatJSON,
		}

		// 创建测试报表
		testReport := &types.Report{
			ID:         1,
			UUID:       "test-uuid-123",
			TenantID:   1,
			ReportType: req.ReportType,
			Status:     types.ReportStatusCompleted,
			FilePath:   "/tmp/test-report.json",
		}

		// 尝试从缓存获取（应该失败）
		_, err := cacheManager.GetReportFromCache(ctx, req)
		assert.Error(t, err, "缓存中应该不存在报表")

		// 缓存报表
		err = cacheManager.CacheReport(ctx, req, testReport)
		assert.NoError(t, err, "缓存报表不应该出错")

		// 从缓存获取（应该成功）
		cachedReport, err := cacheManager.GetReportFromCache(ctx, req)
		assert.NoError(t, err, "应该能从缓存获取报表")
		assert.Equal(t, testReport.UUID, cachedReport.UUID, "缓存的报表UUID应该匹配")
		assert.Equal(t, testReport.FilePath, cachedReport.FilePath, "缓存的报表路径应该匹配")
	})

	t.Run("缓存失效测试", func(t *testing.T) {
		// 失效所有缓存
		err := cacheManager.InvalidateCache(ctx, "*")
		assert.NoError(t, err, "缓存失效不应该出错")
	})
}

// BenchmarkCacheKeyGeneration 缓存键生成性能测试
func BenchmarkCacheKeyGeneration(b *testing.B) {
	cacheManager := service.NewCacheManager()
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Now().AddDate(0, -1, 0),
		EndDate:    time.Now().AddDate(0, 0, -1),
		FileFormat: types.FileFormatJSON,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cacheManager.GetCacheKey(req)
	}
}
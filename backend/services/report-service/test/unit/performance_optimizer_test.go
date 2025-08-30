package test

import (
	"context"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformanceOptimizer_OptimizeDataQuery(t *testing.T) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	tests := []struct {
		name        string
		request     *types.ReportCreateRequest
		expectError bool
		checkResult func(t *testing.T, result *service.OptimizedQuery)
	}{
		{
			name: "财务报表查询优化",
			request: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  time.Now().AddDate(0, -1, 0),
				EndDate:    time.Now(),
				TenantID:   1,
			},
			expectError: false,
			checkResult: func(t *testing.T, result *service.OptimizedQuery) {
				assert.NotEmpty(t, result.Query, "查询语句不能为空")
				assert.Contains(t, result.Query, "tenant_id", "应包含租户过滤条件")
				assert.Greater(t, result.ChunkSize, 0, "分块大小应大于0")
				assert.True(t, result.CacheEnabled, "应启用缓存")
				assert.NotEmpty(t, result.Indexes, "应有索引建议")
			},
		},
		{
			name: "商户运营报表查询优化",
			request: &types.ReportCreateRequest{
				ReportType: types.ReportTypeMerchantOperation,
				PeriodType: types.PeriodTypeWeekly,
				StartDate:  time.Now().AddDate(0, 0, -7),
				EndDate:    time.Now(),
				TenantID:   1,
			},
			expectError: false,
			checkResult: func(t *testing.T, result *service.OptimizedQuery) {
				assert.Contains(t, result.Query, "merchants", "应包含merchants表")
				assert.Contains(t, result.Query, "orders", "应包含orders表")
				assert.Greater(t, len(result.Indexes), 0, "应有索引建议")
			},
		},
		{
			name: "客户分析报表查询优化",
			request: &types.ReportCreateRequest{
				ReportType: types.ReportTypeCustomerAnalysis,
				PeriodType: types.PeriodTypeDaily,
				StartDate:  time.Now().AddDate(0, 0, -1),
				EndDate:    time.Now(),
				TenantID:   1,
			},
			expectError: false,
			checkResult: func(t *testing.T, result *service.OptimizedQuery) {
				assert.Contains(t, result.Query, "customers", "应包含customers表")
				assert.Greater(t, result.ChunkSize, 0, "分块大小应大于0")
			},
		},
		{
			name: "大数据量查询优化（超过30天）",
			request: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  time.Now().AddDate(0, -2, 0), // 2个月前
				EndDate:    time.Now(),
				TenantID:   1,
			},
			expectError: false,
			checkResult: func(t *testing.T, result *service.OptimizedQuery) {
				assert.True(t, result.UseStreaming, "大数据量应启用流式处理")
				assert.Equal(t, 5000, result.ChunkSize, "大数据量分块大小应为5000")
				assert.Equal(t, 4*time.Hour, result.CacheDuration, "大数据量缓存时间应为4小时")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := optimizer.OptimizeDataQuery(ctx, tt.request)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestPerformanceOptimizer_ProcessLargeDataset(t *testing.T) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	// 测试数据处理器
	processedChunks := 0
	processor := func(ctx context.Context, chunk []interface{}) error {
		processedChunks++
		assert.NotEmpty(t, chunk, "数据块不能为空")
		return nil
	}

	tests := []struct {
		name        string
		chunkSize   int
		expectError bool
	}{
		{
			name:        "正常分块大小",
			chunkSize:   10000,
			expectError: false,
		},
		{
			name:        "小分块大小",
			chunkSize:   1000,
			expectError: false,
		},
		{
			name:        "大分块大小",
			chunkSize:   50000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processedChunks = 0
			err := optimizer.ProcessLargeDataset(ctx, processor, tt.chunkSize)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.Greater(t, processedChunks, 0, "应该处理了一些数据块")
		})
	}
}

func TestPerformanceOptimizer_OptimizeMemoryUsage(t *testing.T) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	err := optimizer.OptimizeMemoryUsage(ctx)
	assert.NoError(t, err, "内存优化不应产生错误")
}

func TestPerformanceOptimizer_GetPerformanceMetrics(t *testing.T) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	// 先执行一些操作以生成指标
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeDaily,
		StartDate:  time.Now().AddDate(0, 0, -1),
		EndDate:    time.Now(),
		TenantID:   1,
	}
	_, _ = optimizer.OptimizeDataQuery(ctx, req)
	_ = optimizer.OptimizeMemoryUsage(ctx)

	metrics := optimizer.GetPerformanceMetrics(ctx)
	require.NotNil(t, metrics, "性能指标不能为空")
	
	assert.GreaterOrEqual(t, metrics.MemoryUsage, int64(0), "内存使用量应大于等于0")
	assert.GreaterOrEqual(t, metrics.CPUUsage, float64(0), "CPU使用率应大于等于0")
	assert.GreaterOrEqual(t, metrics.ActiveConnections, 0, "活跃连接数应大于等于0")
	assert.GreaterOrEqual(t, metrics.CacheHitRate, float64(0), "缓存命中率应大于等于0")
	assert.GreaterOrEqual(t, metrics.ThroughputPerSecond, int64(0), "吞吐量应大于等于0")
	assert.GreaterOrEqual(t, metrics.ErrorRate, float64(0), "错误率应大于等于0")
}

func TestPerformanceOptimizer_EnableParallelProcessing(t *testing.T) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	tests := []struct {
		name       string
		maxWorkers int
		expectOK   bool
	}{
		{
			name:       "正常工作线程数",
			maxWorkers: 4,
			expectOK:   true,
		},
		{
			name:       "零工作线程数（应使用CPU核数）",
			maxWorkers: 0,
			expectOK:   true,
		},
		{
			name:       "负数工作线程数（应使用CPU核数）",
			maxWorkers: -1,
			expectOK:   true,
		},
		{
			name:       "大工作线程数",
			maxWorkers: 100,
			expectOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := optimizer.EnableParallelProcessing(ctx, tt.maxWorkers)
			
			if tt.expectOK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestPerformanceOptimizer_CreateOptimizationReport(t *testing.T) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	// 先执行一些操作
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Now().AddDate(0, -1, 0),
		EndDate:    time.Now(),
		TenantID:   1,
	}
	_, _ = optimizer.OptimizeDataQuery(ctx, req)
	_ = optimizer.OptimizeMemoryUsage(ctx)

	report, err := optimizer.CreateOptimizationReport(ctx)
	
	assert.NoError(t, err, "创建优化报告不应产生错误")
	assert.NotNil(t, report, "优化报告不能为空")
	
	// 检查报告结构
	assert.Contains(t, report, "timestamp", "报告应包含时间戳")
	assert.Contains(t, report, "metrics", "报告应包含性能指标")
	assert.Contains(t, report, "optimization_suggestions", "报告应包含优化建议")
	assert.Contains(t, report, "resource_usage", "报告应包含资源使用情况")
	
	// 检查优化建议
	suggestions, ok := report["optimization_suggestions"].([]string)
	assert.True(t, ok, "优化建议应为字符串数组")
	assert.NotEmpty(t, suggestions, "应有优化建议")
	
	// 检查资源使用情况
	resourceUsage, ok := report["resource_usage"].(map[string]interface{})
	assert.True(t, ok, "资源使用情况应为映射")
	assert.Contains(t, resourceUsage, "memory_mb", "应包含内存使用量")
	assert.Contains(t, resourceUsage, "cpu_percent", "应包含CPU使用率")
	assert.Contains(t, resourceUsage, "workers", "应包含工作线程数")
	assert.Contains(t, resourceUsage, "chunk_size", "应包含分块大小")
}

func TestPerformanceOptimizer_DataProcessorError(t *testing.T) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	// 创建一个会出错的处理器
	processor := func(ctx context.Context, chunk []interface{}) error {
		return assert.AnError // 返回错误
	}

	err := optimizer.ProcessLargeDataset(ctx, processor, 1000)
	assert.Error(t, err, "处理器出错时应返回错误")
	assert.Contains(t, err.Error(), "个错误", "错误信息应包含错误计数")
}

func TestPerformanceOptimizer_OptimizationSuggestions(t *testing.T) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	// 执行一些操作以生成数据
	_ = optimizer.OptimizeMemoryUsage(ctx)
	
	report, err := optimizer.CreateOptimizationReport(ctx)
	require.NoError(t, err)
	require.NotNil(t, report)

	suggestions, ok := report["optimization_suggestions"].([]string)
	require.True(t, ok, "优化建议应为字符串数组")
	require.NotEmpty(t, suggestions, "应有优化建议")

	// 验证建议内容合理性
	for _, suggestion := range suggestions {
		assert.NotEmpty(t, suggestion, "优化建议不能为空")
		assert.True(t, len(suggestion) > 10, "优化建议应该有足够的描述")
	}
}

// 基准测试

func BenchmarkPerformanceOptimizer_OptimizeDataQuery(b *testing.B) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()
	
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Now().AddDate(0, -1, 0),
		EndDate:    time.Now(),
		TenantID:   1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := optimizer.OptimizeDataQuery(ctx, req)
		if err != nil {
			b.Fatalf("查询优化失败: %v", err)
		}
	}
}

func BenchmarkPerformanceOptimizer_GetPerformanceMetrics(b *testing.B) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := optimizer.GetPerformanceMetrics(ctx)
		if metrics == nil {
			b.Fatal("性能指标不能为空")
		}
	}
}

func BenchmarkPerformanceOptimizer_OptimizeMemoryUsage(b *testing.B) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := optimizer.OptimizeMemoryUsage(ctx)
		if err != nil {
			b.Fatalf("内存优化失败: %v", err)
		}
	}
}
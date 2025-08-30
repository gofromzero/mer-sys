package performance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReportGenerationPerformance 报表生成性能测试
func TestReportGenerationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	optimizer := service.NewPerformanceOptimizer()
	cacheManager := service.NewCacheManager()
	ctx := context.Background()

	tests := []struct {
		name           string
		reportType     types.ReportType
		dataRange      time.Duration
		expectedMaxTime time.Duration
		concurrency    int
	}{
		{
			name:           "小型财务报表生成",
			reportType:     types.ReportTypeFinancial,
			dataRange:      24 * time.Hour,
			expectedMaxTime: 30 * time.Second,
			concurrency:    1,
		},
		{
			name:           "中型商户运营报表生成",
			reportType:     types.ReportTypeMerchantOperation,
			dataRange:      7 * 24 * time.Hour,
			expectedMaxTime: 2 * time.Minute,
			concurrency:    1,
		},
		{
			name:           "大型客户分析报表生成",
			reportType:     types.ReportTypeCustomerAnalysis,
			dataRange:      30 * 24 * time.Hour,
			expectedMaxTime: 5 * time.Minute,
			concurrency:    1,
		},
		{
			name:           "并发财务报表生成",
			reportType:     types.ReportTypeFinancial,
			dataRange:      7 * 24 * time.Hour,
			expectedMaxTime: 1 * time.Minute,
			concurrency:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now()
			
			// 创建报表请求
			req := &types.ReportCreateRequest{
				ReportType: tt.reportType,
				PeriodType: types.PeriodTypeDaily,
				StartDate:  time.Now().Add(-tt.dataRange),
				EndDate:    time.Now(),
				FileFormat: types.FileFormatJSON,
				TenantID:   1,
			}

			if tt.concurrency == 1 {
				// 单个报表生成
				err := performSingleReportGeneration(ctx, optimizer, cacheManager, req)
				assert.NoError(t, err, "报表生成不应失败")
			} else {
				// 并发报表生成
				err := performConcurrentReportGeneration(ctx, optimizer, cacheManager, req, tt.concurrency)
				assert.NoError(t, err, "并发报表生成不应失败")
			}

			executionTime := time.Since(startTime)
			t.Logf("报表生成耗时: %v (期望最大: %v)", executionTime, tt.expectedMaxTime)
			
			assert.Less(t, executionTime, tt.expectedMaxTime, 
				fmt.Sprintf("报表生成时间应小于%v，实际: %v", tt.expectedMaxTime, executionTime))
		})
	}
}

// TestCachePerformance 缓存性能测试
func TestCachePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	cacheManager := service.NewCacheManager()
	advancedStrategy := service.NewAdvancedCacheStrategy(cacheManager)
	ctx := context.Background()

	tests := []struct {
		name             string
		operationCount   int
		expectedAvgTime  time.Duration
		operation        string
	}{
		{
			name:            "缓存读取性能",
			operationCount:  1000,
			expectedAvgTime: 10 * time.Millisecond,
			operation:       "read",
		},
		{
			name:            "缓存写入性能",
			operationCount:  1000,
			expectedAvgTime: 20 * time.Millisecond,
			operation:       "write",
		},
		{
			name:            "缓存压缩性能",
			operationCount:  100,
			expectedAvgTime: 50 * time.Millisecond,
			operation:       "compress",
		},
		{
			name:            "缓存解压性能",
			operationCount:  100,
			expectedAvgTime: 30 * time.Millisecond,
			operation:       "decompress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalTime := time.Duration(0)

			for i := 0; i < tt.operationCount; i++ {
				var err error
				startTime := time.Now()

				switch tt.operation {
				case "read":
					req := createTestReportRequest(i)
					_, err = cacheManager.GetReportFromCache(ctx, req)
					// 预期会失败（缓存未命中），这是正常的
				case "write":
					req := createTestReportRequest(i)
					report := createTestReport(i)
					err = cacheManager.CacheReport(ctx, req, report)
				case "compress":
					testData := createTestData(1024 * 10) // 10KB数据
					_, err = advancedStrategy.CompressCache(ctx, testData)
				case "decompress":
					testData := createTestData(1024 * 6) // 6KB压缩数据
					_, err = advancedStrategy.DecompressCache(ctx, testData)
				}

				executionTime := time.Since(startTime)
				totalTime += executionTime

				// 对于某些操作，错误是预期的
				if tt.operation != "read" {
					assert.NoError(t, err, fmt.Sprintf("操作%s在第%d次时失败", tt.operation, i))
				}
			}

			avgTime := totalTime / time.Duration(tt.operationCount)
			t.Logf("平均执行时间: %v (期望最大: %v)", avgTime, tt.expectedAvgTime)

			assert.Less(t, avgTime, tt.expectedAvgTime,
				fmt.Sprintf("平均执行时间应小于%v，实际: %v", tt.expectedAvgTime, avgTime))
		})
	}
}

// TestMemoryUsagePerformance 内存使用性能测试
func TestMemoryUsagePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	tests := []struct {
		name            string
		dataSize        int
		chunkSize       int
		expectedMaxMem  int64 // MB
	}{
		{
			name:           "小数据集处理",
			dataSize:       10000,
			chunkSize:      1000,
			expectedMaxMem: 50, // 50MB
		},
		{
			name:           "中等数据集处理",
			dataSize:       100000,
			chunkSize:      10000,
			expectedMaxMem: 200, // 200MB
		},
		{
			name:           "大数据集处理",
			dataSize:       1000000,
			chunkSize:      50000,
			expectedMaxMem: 500, // 500MB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 记录初始内存使用
			initialMetrics := optimizer.GetPerformanceMetrics(ctx)
			initialMem := initialMetrics.MemoryUsage

			// 模拟大数据处理
			processor := func(ctx context.Context, chunk []interface{}) error {
				// 模拟数据处理，分配一些内存
				tempData := make([]byte, len(chunk)*100) // 每个条目100字节
				_ = tempData
				return nil
			}

			err := optimizer.ProcessLargeDataset(ctx, processor, tt.chunkSize)
			require.NoError(t, err)

			// 记录最终内存使用
			finalMetrics := optimizer.GetPerformanceMetrics(ctx)
			finalMem := finalMetrics.MemoryUsage

			memUsedMB := (finalMem - initialMem) / 1024 / 1024
			t.Logf("内存使用增长: %d MB (期望最大: %d MB)", memUsedMB, tt.expectedMaxMem)

			assert.LessOrEqual(t, memUsedMB, tt.expectedMaxMem,
				fmt.Sprintf("内存使用应小于等于%d MB，实际: %d MB", tt.expectedMaxMem, memUsedMB))

			// 清理内存
			err = optimizer.OptimizeMemoryUsage(ctx)
			assert.NoError(t, err)
		})
	}
}

// TestConcurrentReportGeneration 并发报表生成测试
func TestConcurrentReportGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	concurrencyLevels := []int{1, 5, 10, 20}
	
	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("并发度_%d", concurrency), func(t *testing.T) {
			startTime := time.Now()
			
			var wg sync.WaitGroup
			errors := make(chan error, concurrency)
			
			// 启用并行处理
			err := optimizer.EnableParallelProcessing(ctx, concurrency)
			require.NoError(t, err)

			// 启动并发任务
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					
					req := &types.ReportCreateRequest{
						ReportType: types.ReportTypeFinancial,
						PeriodType: types.PeriodTypeDaily,
						StartDate:  time.Now().AddDate(0, 0, -1),
						EndDate:    time.Now(),
						FileFormat: types.FileFormatJSON,
						TenantID:   uint64(id + 1),
					}

					if err := simulateReportGeneration(ctx, optimizer, req); err != nil {
						errors <- fmt.Errorf("任务%d失败: %v", id, err)
						return
					}
				}(i)
			}

			// 等待所有任务完成
			wg.Wait()
			close(errors)

			// 检查错误
			for err := range errors {
				t.Error(err)
			}

			executionTime := time.Since(startTime)
			
			// 计算期望的最大时间（基于并发度）
			expectedMaxTime := time.Duration(float64(60*time.Second) / float64(concurrency) * 1.5)
			
			t.Logf("并发度%d的执行时间: %v (期望最大: %v)", concurrency, executionTime, expectedMaxTime)
			
			assert.Less(t, executionTime, expectedMaxTime,
				fmt.Sprintf("并发度%d的执行时间应小于%v，实际: %v", concurrency, expectedMaxTime, executionTime))
		})
	}
}

// TestLargeDatasetProcessing 大数据集处理测试
func TestLargeDatasetProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	datasetSizes := []int{100000, 500000, 1000000} // 10万、50万、100万条记录
	
	for _, size := range datasetSizes {
		t.Run(fmt.Sprintf("数据集大小_%d", size), func(t *testing.T) {
			startTime := time.Now()
			
			processedCount := 0
			processor := func(ctx context.Context, chunk []interface{}) error {
				processedCount += len(chunk)
				
				// 模拟实际的数据处理工作
				for _, item := range chunk {
					data := item.(map[string]interface{})
					_ = fmt.Sprintf("processing_%v", data["id"])
				}
				
				return nil
			}

			// 动态调整块大小
			chunkSize := 10000
			if size > 500000 {
				chunkSize = 50000
			}

			err := optimizer.ProcessLargeDataset(ctx, processor, chunkSize)
			require.NoError(t, err)

			executionTime := time.Since(startTime)
			throughput := float64(processedCount) / executionTime.Seconds()
			
			// 期望的最小吞吐量（记录/秒）
			expectedMinThroughput := 10000.0 // 每秒至少处理1万条记录
			
			t.Logf("数据集大小: %d, 处理时间: %v, 吞吐量: %.0f 记录/秒", 
				size, executionTime, throughput)
			
			assert.GreaterOrEqual(t, throughput, expectedMinThroughput,
				fmt.Sprintf("吞吐量应大于等于%.0f记录/秒，实际: %.0f记录/秒", expectedMinThroughput, throughput))
			
			// 验证所有数据都被处理了
			assert.Equal(t, size, processedCount, "所有数据都应该被处理")
		})
	}
}

// 辅助函数

func performSingleReportGeneration(ctx context.Context, optimizer service.IPerformanceOptimizer, cacheManager service.ICacheManager, req *types.ReportCreateRequest) error {
	// 优化查询
	optimizedQuery, err := optimizer.OptimizeDataQuery(ctx, req)
	if err != nil {
		return fmt.Errorf("查询优化失败: %v", err)
	}

	// 模拟报表生成
	time.Sleep(time.Duration(optimizedQuery.ChunkSize/1000) * time.Millisecond)

	// 模拟缓存操作
	if cacheManager.ShouldUseCache(req) {
		report := createTestReport(1)
		_ = cacheManager.CacheReport(ctx, req, report)
	}

	return nil
}

func performConcurrentReportGeneration(ctx context.Context, optimizer service.IPerformanceOptimizer, cacheManager service.ICacheManager, req *types.ReportCreateRequest, concurrency int) error {
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// 为每个并发任务创建不同的请求
			reqCopy := *req
			reqCopy.TenantID = uint64(id + 1)
			
			if err := performSingleReportGeneration(ctx, optimizer, cacheManager, &reqCopy); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 收集错误
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func simulateReportGeneration(ctx context.Context, optimizer service.IPerformanceOptimizer, req *types.ReportCreateRequest) error {
	// 模拟报表生成过程
	optimizedQuery, err := optimizer.OptimizeDataQuery(ctx, req)
	if err != nil {
		return err
	}

	// 模拟数据处理
	processor := func(ctx context.Context, chunk []interface{}) error {
		// 模拟一些计算
		time.Sleep(time.Microsecond * 100)
		return nil
	}

	return optimizer.ProcessLargeDataset(ctx, processor, optimizedQuery.ChunkSize)
}

func createTestReportRequest(id int) *types.ReportCreateRequest {
	return &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeDaily,
		StartDate:  time.Now().AddDate(0, 0, -1),
		EndDate:    time.Now(),
		FileFormat: types.FileFormatJSON,
		TenantID:   uint64(id + 1),
	}
}

func createTestReport(id int) *types.Report {
	return &types.Report{
		ID:          int64(id),
		TenantID:    uint64(id + 1),
		ReportType:  types.ReportTypeFinancial,
		PeriodType:  types.PeriodTypeDaily,
		StartDate:   time.Now().AddDate(0, 0, -1),
		EndDate:     time.Now(),
		Status:      types.ReportStatusCompleted,
		FileFormat:  types.FileFormatJSON,
		GeneratedBy: 1,
		GeneratedAt: time.Now(),
	}
}

func createTestData(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// 基准测试

func BenchmarkReportGeneration(b *testing.B) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeDaily,
		StartDate:  time.Now().AddDate(0, 0, -1),
		EndDate:    time.Now(),
		FileFormat: types.FileFormatJSON,
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

func BenchmarkCacheOperations(b *testing.B) {
	cacheManager := service.NewCacheManager()
	ctx := context.Background()

	req := createTestReportRequest(1)
	report := createTestReport(1)

	b.Run("CacheWrite", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = cacheManager.CacheReport(ctx, req, report)
		}
	})

	b.Run("CacheRead", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = cacheManager.GetReportFromCache(ctx, req)
		}
	})
}

func BenchmarkDataProcessing(b *testing.B) {
	optimizer := service.NewPerformanceOptimizer()
	ctx := context.Background()

	processor := func(ctx context.Context, chunk []interface{}) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = optimizer.ProcessLargeDataset(ctx, processor, 1000)
	}
}
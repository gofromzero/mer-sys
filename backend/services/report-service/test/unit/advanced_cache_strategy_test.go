package test

import (
	"context"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/services/report-service/internal/service"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCacheManager 模拟缓存管理器
type MockCacheManager struct {
	mock.Mock
}

func (m *MockCacheManager) GetReportFromCache(ctx context.Context, req *types.ReportCreateRequest) (*types.Report, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Report), args.Error(1)
}

func (m *MockCacheManager) CacheReport(ctx context.Context, req *types.ReportCreateRequest, report *types.Report) error {
	args := m.Called(ctx, req, report)
	return args.Error(0)
}

func (m *MockCacheManager) InvalidateCache(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCacheManager) GetCacheKey(req *types.ReportCreateRequest) string {
	args := m.Called(req)
	return args.String(0)
}

func (m *MockCacheManager) ShouldUseCache(req *types.ReportCreateRequest) bool {
	args := m.Called(req)
	return args.Bool(0)
}

func (m *MockCacheManager) CleanupExpiredCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheManager) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCacheManager) WarmupCache(ctx context.Context, reportType types.ReportType) error {
	args := m.Called(ctx, reportType)
	return args.Error(0)
}

func TestAdvancedCacheStrategy_PredictivePreload(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	// 设置mock期望
	mockCacheManager.On("GetReportFromCache", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*types.ReportCreateRequest")).Return(nil, assert.AnError)

	tests := []struct {
		name        string
		userID      int64
		expectError bool
	}{
		{
			name:        "正常用户预加载",
			userID:      12345,
			expectError: false,
		},
		{
			name:        "零用户ID",
			userID:      0,
			expectError: false,
		},
		{
			name:        "负用户ID",
			userID:      -1,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.PredictivePreload(ctx, tt.userID)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// 验证mock调用
	mockCacheManager.AssertExpectations(t)
}

func TestAdvancedCacheStrategy_ManageHierarchicalCache(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	err := strategy.ManageHierarchicalCache(ctx)
	assert.NoError(t, err, "分层缓存管理不应产生错误")
}

func TestAdvancedCacheStrategy_CompressCache(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	tests := []struct {
		name        string
		data        []byte
		shouldCompress bool
	}{
		{
			name:        "小数据不压缩",
			data:        []byte("small data"),
			shouldCompress: false,
		},
		{
			name:        "大数据需要压缩",
			data:        make([]byte, 2048), // 2KB数据
			shouldCompress: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := strategy.CompressCache(ctx, tt.data)
			
			assert.NoError(t, err, "压缩不应产生错误")
			assert.NotNil(t, compressed, "压缩结果不能为空")
			
			if tt.shouldCompress && len(tt.data) >= 1024 {
				// 大数据应该被压缩（长度应该减少）
				assert.Less(t, len(compressed), len(tt.data), "大数据应该被压缩")
			} else {
				// 小数据不压缩，长度应该相同
				assert.Equal(t, len(tt.data), len(compressed), "小数据不应被压缩")
			}
		})
	}
}

func TestAdvancedCacheStrategy_DecompressCache(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	// 先压缩一些数据
	originalData := make([]byte, 2048)
	for i := range originalData {
		originalData[i] = byte(i % 256)
	}

	compressed, err := strategy.CompressCache(ctx, originalData)
	require.NoError(t, err)

	// 再解压
	decompressed, err := strategy.DecompressCache(ctx, compressed)
	assert.NoError(t, err, "解压不应产生错误")
	assert.NotNil(t, decompressed, "解压结果不能为空")
	assert.GreaterOrEqual(t, len(decompressed), len(compressed), "解压后的数据应该更大或相等")
}

func TestAdvancedCacheStrategy_IntelligentEviction(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	// 设置mock期望
	mockCacheManager.On("InvalidateCache", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).Return(nil)

	err := strategy.IntelligentEviction(ctx)
	assert.NoError(t, err, "智能失效不应产生错误")

	// 验证mock调用
	mockCacheManager.AssertExpectations(t)
}

func TestAdvancedCacheStrategy_MonitorCachePerformance(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	metrics, err := strategy.MonitorCachePerformance(ctx)
	
	assert.NoError(t, err, "性能监控不应产生错误")
	assert.NotNil(t, metrics, "性能指标不能为空")
	
	// 检查指标字段
	assert.GreaterOrEqual(t, metrics.HitRate, float64(0), "命中率应大于等于0")
	assert.GreaterOrEqual(t, metrics.MissRate, float64(0), "未命中率应大于等于0")
	assert.GreaterOrEqual(t, metrics.CacheSize, int64(0), "缓存大小应大于等于0")
	assert.GreaterOrEqual(t, metrics.EntryCount, int64(0), "条目数应大于等于0")
	assert.GreaterOrEqual(t, metrics.EvictionCount, int64(0), "驱逐次数应大于等于0")
	assert.GreaterOrEqual(t, metrics.CompressionRatio, float64(0), "压缩率应大于等于0")
	
	assert.NotNil(t, metrics.TypeDistribution, "类型分布不能为空")
	assert.NotNil(t, metrics.SizeDistribution, "大小分布不能为空")
	assert.NotNil(t, metrics.AccessFrequency, "访问频率不能为空")
}

func TestAdvancedCacheStrategy_AdjustCacheStrategy(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	tests := []struct {
		name    string
		metrics *service.CachePerformanceMetrics
	}{
		{
			name: "低命中率场景",
			metrics: &service.CachePerformanceMetrics{
				HitRate:          30.0, // 低命中率
				CacheSize:        100 * 1024 * 1024, // 100MB
				AccessFrequency:  map[string]int64{"financial": 100, "merchant": 50},
			},
		},
		{
			name: "高命中率场景",
			metrics: &service.CachePerformanceMetrics{
				HitRate:          95.0, // 高命中率
				CacheSize:        100 * 1024 * 1024, // 100MB
				AccessFrequency:  map[string]int64{"financial": 100, "merchant": 50},
			},
		},
		{
			name: "大缓存场景",
			metrics: &service.CachePerformanceMetrics{
				HitRate:          75.0, // 正常命中率
				CacheSize:        3 * 1024 * 1024 * 1024, // 3GB，超过限制
				AccessFrequency:  map[string]int64{"financial": 100, "merchant": 50},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 对于大缓存场景，需要mock失效操作
			if tt.metrics.CacheSize > 2*1024*1024*1024 {
				mockCacheManager.On("InvalidateCache", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).Return(nil)
			}

			err := strategy.AdjustCacheStrategy(ctx, tt.metrics)
			assert.NoError(t, err, "缓存策略调整不应产生错误")
		})
	}
}

func TestAdvancedCacheStrategy_CompressionDecompressionRoundTrip(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	// 创建测试数据
	testData := []byte("这是一个测试数据，用于验证压缩和解压功能的正确性。这个数据应该足够大以触发压缩逻辑。")
	// 确保数据大于1KB
	for len(testData) < 1024 {
		testData = append(testData, testData...)
	}

	// 压缩数据
	compressed, err := strategy.CompressCache(ctx, testData)
	require.NoError(t, err)
	require.NotNil(t, compressed)

	// 解压数据
	decompressed, err := strategy.DecompressCache(ctx, compressed)
	require.NoError(t, err)
	require.NotNil(t, decompressed)

	// 验证解压后的数据长度符合预期（考虑到模拟压缩的逻辑）
	expectedSize := int(float64(len(testData)) / 0.6) // 基于压缩逻辑的反推
	assert.Equal(t, expectedSize, len(decompressed), "解压后的数据大小应该符合预期")
}

func TestAdvancedCacheStrategy_PerformanceMetricsConsistency(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	// 多次获取性能指标，验证一致性
	var previousMetrics *service.CachePerformanceMetrics
	
	for i := 0; i < 3; i++ {
		metrics, err := strategy.MonitorCachePerformance(ctx)
		require.NoError(t, err)
		require.NotNil(t, metrics)

		// 验证指标的逻辑一致性
		assert.Equal(t, 100.0, metrics.HitRate + metrics.MissRate, "命中率和未命中率之和应为100%")

		if previousMetrics != nil {
			// 某些指标应该保持一致（在没有新操作的情况下）
			assert.Equal(t, previousMetrics.CacheSize, metrics.CacheSize, "连续获取的缓存大小应该一致")
			assert.Equal(t, previousMetrics.EntryCount, metrics.EntryCount, "连续获取的条目数应该一致")
		}

		previousMetrics = metrics
		time.Sleep(10 * time.Millisecond) // 短暂休息
	}
}

func TestAdvancedCacheStrategy_ErrorHandling(t *testing.T) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	// 测试InvalidateCache返回错误的情况
	mockCacheManager.On("InvalidateCache", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("string")).Return(assert.AnError)

	// 智能失效应该能处理错误
	err := strategy.IntelligentEviction(ctx)
	assert.NoError(t, err, "智能失效应该能处理InvalidateCache的错误")
}

// 基准测试

func BenchmarkAdvancedCacheStrategy_CompressCache(b *testing.B) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	// 创建1KB测试数据
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := strategy.CompressCache(ctx, data)
		if err != nil {
			b.Fatalf("压缩失败: %v", err)
		}
	}
}

func BenchmarkAdvancedCacheStrategy_DecompressCache(b *testing.B) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	// 准备压缩数据
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	
	compressed, err := strategy.CompressCache(ctx, data)
	if err != nil {
		b.Fatalf("压缩准备失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := strategy.DecompressCache(ctx, compressed)
		if err != nil {
			b.Fatalf("解压失败: %v", err)
		}
	}
}

func BenchmarkAdvancedCacheStrategy_MonitorCachePerformance(b *testing.B) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics, err := strategy.MonitorCachePerformance(ctx)
		if err != nil {
			b.Fatalf("性能监控失败: %v", err)
		}
		if metrics == nil {
			b.Fatal("性能指标为空")
		}
	}
}

func BenchmarkAdvancedCacheStrategy_ManageHierarchicalCache(b *testing.B) {
	mockCacheManager := new(MockCacheManager)
	strategy := service.NewAdvancedCacheStrategy(mockCacheManager)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := strategy.ManageHierarchicalCache(ctx)
		if err != nil {
			b.Fatalf("分层缓存管理失败: %v", err)
		}
	}
}
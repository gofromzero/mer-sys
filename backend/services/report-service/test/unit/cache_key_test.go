package unit

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/stretchr/testify/assert"
)

// MockCacheKeyGenerator 模拟缓存键生成器，不依赖数据库
type MockCacheKeyGenerator struct{}

// GetCacheKey 生成缓存键（复制cache_manager.go中的逻辑）
func (m *MockCacheKeyGenerator) GetCacheKey(req *types.ReportCreateRequest) string {
	// 构建缓存键的原始字符串
	keyData := fmt.Sprintf("report:%s:%s:%s:%s", 
		req.ReportType,
		req.PeriodType,
		req.StartDate.Format("2006-01-02"),
		req.EndDate.Format("2006-01-02"))
	
	// 如果指定了商户ID，加入缓存键
	if req.MerchantID != nil {
		keyData += fmt.Sprintf(":merchant:%d", *req.MerchantID)
	}
	
	// 如果有自定义配置，加入缓存键
	if req.Config != nil && len(req.Config) > 0 {
		configJson, _ := json.Marshal(req.Config)
		keyData += fmt.Sprintf(":config:%s", string(configJson))
	}
	
	// 生成MD5哈希作为缓存键
	hash := md5.Sum([]byte(keyData))
	return fmt.Sprintf("report_cache:%x", hash)
}

// ShouldUseCache 判断是否应该使用缓存（复制cache_manager.go中的逻辑）
func (m *MockCacheKeyGenerator) ShouldUseCache(req *types.ReportCreateRequest) bool {
	// 定义支持缓存的报表类型
	supportedTypes := map[types.ReportType]bool{
		types.ReportTypeFinancial:         true,
		types.ReportTypeMerchantOperation: true,
		types.ReportTypeCustomerAnalysis:  true,
	}
	
	// 检查报表类型是否支持缓存
	if !supportedTypes[req.ReportType] {
		return false
	}
	
	// 检查时间范围是否适合缓存
	duration := req.EndDate.Sub(req.StartDate)
	
	// 时间范围太短（小于1小时）或太长（大于1年）不缓存
	if duration < time.Hour || duration > 365*24*time.Hour {
		return false
	}
	
	// 实时数据（结束时间在1小时内）不缓存
	if req.EndDate.After(time.Now().Add(-time.Hour)) {
		return false
	}
	
	return true
}

func TestMockCacheKeyGenerator_GetCacheKey(t *testing.T) {
	generator := &MockCacheKeyGenerator{}

	t.Run("相同请求生成相同缓存键", func(t *testing.T) {
		req := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			FileFormat: types.FileFormatJSON,
		}

		key1 := generator.GetCacheKey(req)
		key2 := generator.GetCacheKey(req)

		assert.Equal(t, key1, key2, "相同请求应生成相同的缓存键")
		assert.Contains(t, key1, "report_cache:", "缓存键应包含前缀")
		assert.Len(t, key1, len("report_cache:")+32, "缓存键长度应为前缀+MD5长度")
	})

	t.Run("不同请求生成不同缓存键", func(t *testing.T) {
		req1 := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			FileFormat: types.FileFormatJSON,
		}

		req2 := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2025, 2, 28, 23, 59, 59, 0, time.UTC),
			FileFormat: types.FileFormatJSON,
		}

		key1 := generator.GetCacheKey(req1)
		key2 := generator.GetCacheKey(req2)

		assert.NotEqual(t, key1, key2, "不同请求应生成不同的缓存键")
	})

	t.Run("包含商户ID的请求生成不同缓存键", func(t *testing.T) {
		baseReq := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			FileFormat: types.FileFormatJSON,
		}

		merchantID := uint64(123)
		merchantReq := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			FileFormat: types.FileFormatJSON,
			MerchantID: &merchantID,
		}

		baseKey := generator.GetCacheKey(baseReq)
		merchantKey := generator.GetCacheKey(merchantReq)

		assert.NotEqual(t, baseKey, merchantKey, "包含商户ID的请求应生成不同的缓存键")
	})

	t.Run("包含自定义配置的请求生成不同缓存键", func(t *testing.T) {
		baseReq := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			FileFormat: types.FileFormatJSON,
		}

		configReq := &types.ReportCreateRequest{
			ReportType: types.ReportTypeFinancial,
			PeriodType: types.PeriodTypeMonthly,
			StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
			FileFormat: types.FileFormatJSON,
			Config:     map[string]interface{}{"include_breakdown": true},
		}

		baseKey := generator.GetCacheKey(baseReq)
		configKey := generator.GetCacheKey(configReq)

		assert.NotEqual(t, baseKey, configKey, "包含自定义配置的请求应生成不同的缓存键")
	})
}

func TestMockCacheKeyGenerator_ShouldUseCache(t *testing.T) {
	generator := &MockCacheKeyGenerator{}
	now := time.Now()

	testCases := []struct {
		name     string
		req      *types.ReportCreateRequest
		expected bool
		reason   string
	}{
		{
			name: "财务报表-历史数据-应该缓存",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  now.AddDate(0, -2, 0),
				EndDate:    now.Add(-2 * time.Hour), // 2小时前
				FileFormat: types.FileFormatJSON,
			},
			expected: true,
			reason:   "历史财务数据应该缓存",
		},
		{
			name: "商户运营报表-历史数据-应该缓存",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeMerchantOperation,
				PeriodType: types.PeriodTypeWeekly,
				StartDate:  now.AddDate(0, 0, -7),
				EndDate:    now.Add(-2 * time.Hour),
				FileFormat: types.FileFormatExcel,
			},
			expected: true,
			reason:   "历史商户运营数据应该缓存",
		},
		{
			name: "客户分析报表-历史数据-应该缓存",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeCustomerAnalysis,
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  now.AddDate(0, -1, 0),
				EndDate:    now.Add(-3 * time.Hour),
				FileFormat: types.FileFormatPDF,
			},
			expected: true,
			reason:   "历史客户分析数据应该缓存",
		},
		{
			name: "实时数据-不应该缓存",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeDaily,
				StartDate:  now.Add(-30 * time.Minute),
				EndDate:    now,
				FileFormat: types.FileFormatJSON,
			},
			expected: false,
			reason:   "实时数据不应该缓存",
		},
		{
			name: "时间范围太短-不应该缓存",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeDaily,
				StartDate:  now.Add(-30 * time.Minute),
				EndDate:    now.Add(-29 * time.Minute),
				FileFormat: types.FileFormatJSON,
			},
			expected: false,
			reason:   "时间范围小于1小时不应该缓存",
		},
		{
			name: "时间范围太长-不应该缓存",
			req: &types.ReportCreateRequest{
				ReportType: types.ReportTypeFinancial,
				PeriodType: types.PeriodTypeYearly,
				StartDate:  now.AddDate(-2, 0, 0),
				EndDate:    now.Add(-2 * time.Hour),
				FileFormat: types.FileFormatJSON,
			},
			expected: false,
			reason:   "时间范围超过1年不应该缓存",
		},
		{
			name: "不支持的报表类型-不应该缓存",
			req: &types.ReportCreateRequest{
				ReportType: "unsupported_type",
				PeriodType: types.PeriodTypeMonthly,
				StartDate:  now.AddDate(0, -1, 0),
				EndDate:    now.Add(-2 * time.Hour),
				FileFormat: types.FileFormatJSON,
			},
			expected: false,
			reason:   "不支持的报表类型不应该缓存",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generator.ShouldUseCache(tc.req)
			assert.Equal(t, tc.expected, result, tc.reason)
		})
	}
}

// BenchmarkCacheKeyGeneration 缓存键生成性能基准测试
func BenchmarkMockCacheKeyGeneration(b *testing.B) {
	generator := &MockCacheKeyGenerator{}
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Now().AddDate(0, -1, 0),
		EndDate:    time.Now().Add(-2 * time.Hour),
		FileFormat: types.FileFormatJSON,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.GetCacheKey(req)
	}
}

// BenchmarkShouldUseCache 缓存条件判断性能基准测试
func BenchmarkMockShouldUseCache(b *testing.B) {
	generator := &MockCacheKeyGenerator{}
	req := &types.ReportCreateRequest{
		ReportType: types.ReportTypeFinancial,
		PeriodType: types.PeriodTypeMonthly,
		StartDate:  time.Now().AddDate(0, -1, 0),
		EndDate:    time.Now().Add(-2 * time.Hour),
		FileFormat: types.FileFormatJSON,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.ShouldUseCache(req)
	}
}
package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofromzero/mer-sys/backend/shared/repository"
	"github.com/gofromzero/mer-sys/backend/shared/types"
	"github.com/gogf/gf/v2/frame/g"
)

// ICacheManager 缓存管理器接口
type ICacheManager interface {
	GetReportFromCache(ctx context.Context, req *types.ReportCreateRequest) (*types.Report, error)
	CacheReport(ctx context.Context, req *types.ReportCreateRequest, report *types.Report) error
	InvalidateCache(ctx context.Context, pattern string) error
	GetCacheKey(req *types.ReportCreateRequest) string
	ShouldUseCache(req *types.ReportCreateRequest) bool
	CleanupExpiredCache(ctx context.Context) error
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)
	WarmupCache(ctx context.Context, reportType types.ReportType) error
}

// CacheManager 缓存管理器实现
type CacheManager struct {
	reportRepo repository.IReportRepository
	// 缓存TTL配置
	cacheTTL map[types.ReportType]time.Duration
}

// NewCacheManager 创建缓存管理器实例
func NewCacheManager() ICacheManager {
	return &CacheManager{
		reportRepo: repository.NewReportRepository(),
		cacheTTL: map[types.ReportType]time.Duration{
			types.ReportTypeFinancial:         2 * time.Hour,  // 财务报表缓存2小时
			types.ReportTypeMerchantOperation: 1 * time.Hour,  // 商户运营报表缓存1小时
			types.ReportTypeCustomerAnalysis:  30 * time.Minute, // 客户分析报表缓存30分钟
		},
	}
}

// GetReportFromCache 从缓存获取报表
func (c *CacheManager) GetReportFromCache(ctx context.Context, req *types.ReportCreateRequest) (*types.Report, error) {
	if !c.ShouldUseCache(req) {
		return nil, fmt.Errorf("该报表类型不支持缓存")
	}
	
	cacheKey := c.GetCacheKey(req)
	
	g.Log().Debug(ctx, "尝试从缓存获取报表", 
		"cache_key", cacheKey,
		"report_type", req.ReportType)
	
	// 查询缓存中的报表
	cachedData, err := c.reportRepo.GetAnalyticsCache(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("查询缓存失败: %v", err)
	}
	
	// 反序列化报表数据
	var cachedReport types.Report
	if err := json.Unmarshal(cachedData.Data, &cachedReport); err != nil {
		return nil, fmt.Errorf("反序列化缓存报表失败: %v", err)
	}
	
	g.Log().Info(ctx, "成功从缓存获取报表", 
		"report_id", cachedReport.ID,
		"cache_key", cacheKey)
	
	return &cachedReport, nil
}

// CacheReport 缓存报表
func (c *CacheManager) CacheReport(ctx context.Context, req *types.ReportCreateRequest, report *types.Report) error {
	if !c.ShouldUseCache(req) {
		return nil // 不需要缓存
	}
	
	cacheKey := c.GetCacheKey(req)
	ttl := c.cacheTTL[req.ReportType]
	
	g.Log().Debug(ctx, "开始缓存报表", 
		"cache_key", cacheKey,
		"report_id", report.ID,
		"ttl", ttl)
	
	// 序列化报表数据
	reportData, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("序列化报表数据失败: %v", err)
	}
	
	// 创建缓存记录
	cache := &types.AnalyticsCache{
		TenantID:   ctx.Value("tenant_id").(uint64),
		CacheKey:   cacheKey,
		MetricType: string(req.ReportType),
		TimePeriod: c.formatTimePeriod(req.StartDate, req.EndDate),
		Data:       reportData,
		ExpiresAt:  time.Now().Add(ttl),
	}
	
	// 保存到缓存
	if err := c.reportRepo.SetAnalyticsCache(ctx, cache); err != nil {
		return fmt.Errorf("保存报表缓存失败: %v", err)
	}
	
	g.Log().Info(ctx, "报表缓存成功", 
		"cache_key", cacheKey,
		"expires_at", cache.ExpiresAt)
	
	return nil
}

// InvalidateCache 失效缓存
func (c *CacheManager) InvalidateCache(ctx context.Context, pattern string) error {
	g.Log().Info(ctx, "开始失效缓存", "pattern", pattern)
	
	// 这里简化实现，删除所有过期缓存
	// 在真实环境中，可以根据pattern进行更精细的缓存失效
	return c.reportRepo.DeleteExpiredCache(ctx)
}

// GetCacheKey 获取缓存键
func (c *CacheManager) GetCacheKey(req *types.ReportCreateRequest) string {
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

// ShouldUseCache 判断是否应该使用缓存
func (c *CacheManager) ShouldUseCache(req *types.ReportCreateRequest) bool {
	// 检查报表类型是否支持缓存
	if _, exists := c.cacheTTL[req.ReportType]; !exists {
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

// CleanupExpiredCache 清理过期缓存
func (c *CacheManager) CleanupExpiredCache(ctx context.Context) error {
	g.Log().Debug(ctx, "开始清理过期缓存")
	
	err := c.reportRepo.DeleteExpiredCache(ctx)
	if err != nil {
		g.Log().Error(ctx, "清理过期缓存失败", "error", err)
		return err
	}
	
	g.Log().Info(ctx, "过期缓存清理完成")
	return nil
}

// formatTimePeriod 格式化时间周期
func (c *CacheManager) formatTimePeriod(startDate, endDate time.Time) string {
	duration := endDate.Sub(startDate)
	
	if duration < 24*time.Hour {
		return "daily"
	} else if duration < 7*24*time.Hour {
		return "weekly"
	} else if duration < 31*24*time.Hour {
		return "monthly"
	} else if duration < 92*24*time.Hour {
		return "quarterly"
	} else {
		return "yearly"
	}
}

// GetCacheStats 获取缓存统计信息
func (c *CacheManager) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// 这里可以实现缓存统计功能
	// 包括缓存命中率、缓存大小、过期时间分布等
	stats := map[string]interface{}{
		"cache_enabled": true,
		"ttl_settings": c.cacheTTL,
		"last_cleanup": time.Now().Format("2006-01-02 15:04:05"),
	}
	
	return stats, nil
}

// WarmupCache 预热缓存
func (c *CacheManager) WarmupCache(ctx context.Context, reportType types.ReportType) error {
	g.Log().Info(ctx, "开始预热缓存", "report_type", reportType)
	
	// 这里可以实现缓存预热逻辑
	// 例如：生成常用的报表并缓存
	
	// 示例：为最近7天、30天生成常用报表
	now := time.Now()
	commonPeriods := []struct {
		startDate time.Time
		endDate   time.Time
		period    types.PeriodType
	}{
		{now.AddDate(0, 0, -7), now.AddDate(0, 0, -1), types.PeriodTypeWeekly},
		{now.AddDate(0, -1, 0), now.AddDate(0, 0, -1), types.PeriodTypeMonthly},
	}
	
	for _, period := range commonPeriods {
		req := &types.ReportCreateRequest{
			ReportType: reportType,
			PeriodType: period.period,
			StartDate:  period.startDate,
			EndDate:    period.endDate,
			FileFormat: types.FileFormatJSON, // 预热时使用JSON格式
		}
		
		// 检查是否已经缓存
		if _, err := c.GetReportFromCache(ctx, req); err == nil {
			continue // 已经有缓存，跳过
		}
		
		g.Log().Debug(ctx, "预热缓存报表", 
			"report_type", reportType,
			"start_date", req.StartDate,
			"end_date", req.EndDate)
	}
	
	g.Log().Info(ctx, "缓存预热完成", "report_type", reportType)
	return nil
}

// SetCacheTTL 设置缓存TTL
func (c *CacheManager) SetCacheTTL(reportType types.ReportType, ttl time.Duration) {
	c.cacheTTL[reportType] = ttl
	g.Log().Info(context.Background(), "更新缓存TTL", 
		"report_type", reportType, 
		"ttl", ttl)
}
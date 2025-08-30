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

// IAnalyticsService 数据分析服务接口
type IAnalyticsService interface {
	GetFinancialData(ctx context.Context, startDate, endDate time.Time, merchantID *uint64) (*types.FinancialReportData, error)
	GetMerchantOperationData(ctx context.Context, startDate, endDate time.Time) (*types.MerchantOperationReport, error)
	GetCustomerAnalysisData(ctx context.Context, startDate, endDate time.Time) (*types.CustomerAnalysisReport, error)
	CustomQuery(ctx context.Context, req *types.AnalyticsQueryRequest) (interface{}, error)
	ClearCache(ctx context.Context, pattern string) error
}

// AnalyticsService 数据分析服务实现
type AnalyticsService struct {
	reportRepo repository.IReportRepository
}

// NewAnalyticsService 创建数据分析服务实例
func NewAnalyticsService() IAnalyticsService {
	return &AnalyticsService{
		reportRepo: repository.NewReportRepository(),
	}
}

// GetFinancialData 获取财务数据
func (s *AnalyticsService) GetFinancialData(ctx context.Context, startDate, endDate time.Time, merchantID *uint64) (*types.FinancialReportData, error) {
	tenantID := ctx.Value("tenant_id").(uint64)
	
	// 构建缓存键
	cacheKey := s.buildCacheKey("financial", tenantID, startDate, endDate, merchantID)
	
	// 尝试从缓存获取
	cachedData, err := s.getFromCache(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var data types.FinancialReportData
		if err := json.Unmarshal(cachedData.Data, &data); err == nil {
			g.Log().Debug(ctx, "财务数据从缓存获取", "cache_key", cacheKey)
			return &data, nil
		}
	}
	
	// 从数据库查询
	g.Log().Info(ctx, "开始查询财务数据", "tenant_id", tenantID, "start_date", startDate, "end_date", endDate)
	
	data, err := s.reportRepo.GetFinancialData(ctx, tenantID, startDate, endDate, merchantID)
	if err != nil {
		return nil, fmt.Errorf("获取财务数据失败: %v", err)
	}
	
	// 缓存数据
	s.setToCache(ctx, cacheKey, "financial", data, time.Hour)
	
	g.Log().Info(ctx, "财务数据查询完成", 
		"total_revenue", data.TotalRevenue.Amount,
		"order_count", data.OrderCount,
		"merchant_count", data.MerchantCount)
	
	return data, nil
}

// GetMerchantOperationData 获取商户运营数据
func (s *AnalyticsService) GetMerchantOperationData(ctx context.Context, startDate, endDate time.Time) (*types.MerchantOperationReport, error) {
	tenantID := ctx.Value("tenant_id").(uint64)
	
	// 构建缓存键
	cacheKey := s.buildCacheKey("merchant_operation", tenantID, startDate, endDate, nil)
	
	// 尝试从缓存获取
	cachedData, err := s.getFromCache(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var data types.MerchantOperationReport
		if err := json.Unmarshal(cachedData.Data, &data); err == nil {
			g.Log().Debug(ctx, "商户运营数据从缓存获取", "cache_key", cacheKey)
			return &data, nil
		}
	}
	
	// 从数据库查询
	g.Log().Info(ctx, "开始查询商户运营数据", "tenant_id", tenantID)
	
	data, err := s.reportRepo.GetMerchantOperationData(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取商户运营数据失败: %v", err)
	}
	
	// 缓存数据
	s.setToCache(ctx, cacheKey, "merchant_operation", data, 30*time.Minute)
	
	g.Log().Info(ctx, "商户运营数据查询完成", 
		"merchant_count", len(data.MerchantRankings),
		"category_count", len(data.CategoryAnalysis))
	
	return data, nil
}

// GetCustomerAnalysisData 获取客户分析数据
func (s *AnalyticsService) GetCustomerAnalysisData(ctx context.Context, startDate, endDate time.Time) (*types.CustomerAnalysisReport, error) {
	tenantID := ctx.Value("tenant_id").(uint64)
	
	// 构建缓存键
	cacheKey := s.buildCacheKey("customer_analysis", tenantID, startDate, endDate, nil)
	
	// 尝试从缓存获取
	cachedData, err := s.getFromCache(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var data types.CustomerAnalysisReport
		if err := json.Unmarshal(cachedData.Data, &data); err == nil {
			g.Log().Debug(ctx, "客户分析数据从缓存获取", "cache_key", cacheKey)
			return &data, nil
		}
	}
	
	// 从数据库查询
	g.Log().Info(ctx, "开始查询客户分析数据", "tenant_id", tenantID)
	
	data, err := s.reportRepo.GetCustomerAnalysisData(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("获取客户分析数据失败: %v", err)
	}
	
	// 缓存数据
	s.setToCache(ctx, cacheKey, "customer_analysis", data, 30*time.Minute)
	
	dauValue := 0
	if data.ActivityMetrics != nil {
		dauValue = data.ActivityMetrics.DAU
	}
	
	g.Log().Info(ctx, "客户分析数据查询完成", 
		"user_growth_periods", len(data.UserGrowth),
		"dau", dauValue)
	
	return data, nil
}

// CustomQuery 自定义数据查询
func (s *AnalyticsService) CustomQuery(ctx context.Context, req *types.AnalyticsQueryRequest) (interface{}, error) {
	tenantID := ctx.Value("tenant_id").(uint64)
	
	g.Log().Info(ctx, "执行自定义数据查询", 
		"metric_type", req.MetricType,
		"tenant_id", tenantID,
		"start_date", req.StartDate,
		"end_date", req.EndDate)
	
	switch req.MetricType {
	case "revenue_trend":
		return s.getRevenueTrend(ctx, tenantID, req.StartDate, req.EndDate, req.GroupBy, req.MerchantID)
		
	case "order_stats":
		return s.getOrderStats(ctx, tenantID, req.StartDate, req.EndDate, req.Filters, req.MerchantID)
		
	case "merchant_comparison":
		return s.getMerchantComparison(ctx, tenantID, req.StartDate, req.EndDate, req.Filters)
		
	case "customer_segments":
		return s.getCustomerSegments(ctx, tenantID, req.StartDate, req.EndDate, req.Filters)
		
	case "rights_usage":
		return s.getRightsUsage(ctx, tenantID, req.StartDate, req.EndDate, req.MerchantID)
		
	case "product_performance":
		return s.getProductPerformance(ctx, tenantID, req.StartDate, req.EndDate, req.Filters, req.MerchantID)
		
	case "payment_methods":
		return s.getPaymentMethodStats(ctx, tenantID, req.StartDate, req.EndDate, req.MerchantID)
		
	case "geographic_analysis":
		return s.getGeographicAnalysis(ctx, tenantID, req.StartDate, req.EndDate, req.Filters)
		
	case "customer_retention":
		return s.getCustomerRetentionAnalysis(ctx, tenantID, req.StartDate, req.EndDate, req.Filters)
		
	case "sales_funnel":
		return s.getSalesFunnelAnalysis(ctx, tenantID, req.StartDate, req.EndDate, req.MerchantID)
		
	default:
		return nil, fmt.Errorf("不支持的指标类型: %s", req.MetricType)
	}
}

// ClearCache 清理缓存
func (s *AnalyticsService) ClearCache(ctx context.Context, pattern string) error {
	g.Log().Info(ctx, "清理分析数据缓存", "pattern", pattern)
	
	// 这里可以实现更精细的缓存清理逻辑
	// 暂时删除所有过期缓存
	err := s.reportRepo.DeleteExpiredCache(ctx)
	if err != nil {
		return fmt.Errorf("清理缓存失败: %v", err)
	}
	
	return nil
}

// buildCacheKey 构建缓存键
func (s *AnalyticsService) buildCacheKey(metricType string, tenantID uint64, startDate, endDate time.Time, merchantID *uint64) string {
	key := fmt.Sprintf("%s:%d:%s:%s", 
		metricType, tenantID,
		startDate.Format("2006-01-02"), 
		endDate.Format("2006-01-02"))
		
	if merchantID != nil {
		key += fmt.Sprintf(":%d", *merchantID)
	}
	
	// 使用MD5生成固定长度的键
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("analytics:%x", hash)
}

// getFromCache 从缓存获取数据
func (s *AnalyticsService) getFromCache(ctx context.Context, cacheKey string) (*types.AnalyticsCache, error) {
	return s.reportRepo.GetAnalyticsCache(ctx, cacheKey)
}

// setToCache 设置缓存数据
func (s *AnalyticsService) setToCache(ctx context.Context, cacheKey, metricType string, data interface{}, ttl time.Duration) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		g.Log().Warning(ctx, "序列化缓存数据失败", "error", err)
		return
	}
	
	cache := &types.AnalyticsCache{
		TenantID:   ctx.Value("tenant_id").(uint64),
		CacheKey:   cacheKey,
		MetricType: metricType,
		TimePeriod: fmt.Sprintf("%v", ttl),
		Data:       jsonData,
		ExpiresAt:  time.Now().Add(ttl),
	}
	
	err = s.reportRepo.SetAnalyticsCache(ctx, cache)
	if err != nil {
		g.Log().Warning(ctx, "设置缓存失败", "error", err)
	}
}

// getRevenueTrend 获取收入趋势数据
func (s *AnalyticsService) getRevenueTrend(ctx context.Context, tenantID uint64, startDate, endDate time.Time, groupBy string, merchantID *uint64) (interface{}, error) {
	// 根据groupBy参数确定分组方式
	var dateFormat string
	switch groupBy {
	case "day":
		dateFormat = "%Y-%m-%d"
	case "week":
		dateFormat = "%Y-%u"  // 年-周
	case "month":
		dateFormat = "%Y-%m"
	case "quarter":
		dateFormat = "%Y-Q%q"
	default:
		dateFormat = "%Y-%m-%d"  // 默认按天
	}
	
	query := `
		SELECT 
			DATE_FORMAT(created_at, ?) as period,
			COUNT(*) as order_count,
			SUM(total_amount) as revenue,
			AVG(total_amount) as avg_order_value
		FROM orders 
		WHERE tenant_id = ? 
			AND created_at BETWEEN ? AND ?
			AND status IN ('completed', 'paid')
	`
	
	args := []interface{}{dateFormat, tenantID, startDate, endDate}
	
	if merchantID != nil {
		query += " AND merchant_id = ?"
		args = append(args, *merchantID)
	}
	
	query += " GROUP BY DATE_FORMAT(created_at, ?) ORDER BY period ASC"
	args = append(args, dateFormat)
	
	var trends []map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&trends)
	
	return trends, err
}

// getOrderStats 获取订单统计数据
func (s *AnalyticsService) getOrderStats(ctx context.Context, tenantID uint64, startDate, endDate time.Time, filters map[string]interface{}, merchantID *uint64) (interface{}, error) {
	query := `
		SELECT 
			status,
			COUNT(*) as count,
			SUM(total_amount) as total_amount,
			AVG(total_amount) as avg_amount
		FROM orders 
		WHERE tenant_id = ? 
			AND created_at BETWEEN ? AND ?
	`
	
	args := []interface{}{tenantID, startDate, endDate}
	
	if merchantID != nil {
		query += " AND merchant_id = ?"
		args = append(args, *merchantID)
	}
	
	// 应用过滤条件
	if filters != nil {
		if minAmount, ok := filters["min_amount"]; ok {
			query += " AND total_amount >= ?"
			args = append(args, minAmount)
		}
		if maxAmount, ok := filters["max_amount"]; ok {
			query += " AND total_amount <= ?"
			args = append(args, maxAmount)
		}
	}
	
	query += " GROUP BY status ORDER BY count DESC"
	
	var stats []map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&stats)
	
	return stats, err
}

// getMerchantComparison 获取商户对比数据
func (s *AnalyticsService) getMerchantComparison(ctx context.Context, tenantID uint64, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	query := `
		SELECT 
			m.id as merchant_id,
			m.name as merchant_name,
			COUNT(o.id) as order_count,
			COALESCE(SUM(o.total_amount), 0) as total_revenue,
			COALESCE(AVG(o.total_amount), 0) as avg_order_value,
			COUNT(DISTINCT o.customer_id) as customer_count
		FROM merchants m
		LEFT JOIN orders o ON m.id = o.merchant_id 
			AND o.created_at BETWEEN ? AND ?
			AND o.status IN ('completed', 'paid')
		WHERE m.tenant_id = ?
	`
	
	args := []interface{}{startDate, endDate, tenantID}
	
	// 应用过滤条件
	if filters != nil {
		if categoryID, ok := filters["category_id"]; ok {
			query += " AND m.category_id = ?"
			args = append(args, categoryID)
		}
	}
	
	query += " GROUP BY m.id, m.name ORDER BY total_revenue DESC LIMIT 20"
	
	var comparison []map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&comparison)
	
	return comparison, err
}

// getCustomerSegments 获取客户分群数据
func (s *AnalyticsService) getCustomerSegments(ctx context.Context, tenantID uint64, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	query := `
		SELECT 
			CASE 
				WHEN order_count = 1 THEN '新客户'
				WHEN order_count BETWEEN 2 AND 5 THEN '普通客户'  
				WHEN order_count BETWEEN 6 AND 10 THEN '忠诚客户'
				ELSE 'VIP客户'
			END as segment,
			COUNT(*) as customer_count,
			AVG(total_spent) as avg_spent,
			AVG(order_count) as avg_orders
		FROM (
			SELECT 
				o.customer_id,
				COUNT(o.id) as order_count,
				SUM(o.total_amount) as total_spent
			FROM orders o
			WHERE o.tenant_id = ? 
				AND o.created_at BETWEEN ? AND ?
				AND o.status IN ('completed', 'paid')
			GROUP BY o.customer_id
		) customer_stats
		GROUP BY 
			CASE 
				WHEN order_count = 1 THEN '新客户'
				WHEN order_count BETWEEN 2 AND 5 THEN '普通客户'  
				WHEN order_count BETWEEN 6 AND 10 THEN '忠诚客户'
				ELSE 'VIP客户'
			END
		ORDER BY customer_count DESC
	`
	
	args := []interface{}{tenantID, startDate, endDate}
	
	var segments []map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&segments)
	
	return segments, err
}

// getRightsUsage 获取权益使用数据
func (s *AnalyticsService) getRightsUsage(ctx context.Context, tenantID uint64, startDate, endDate time.Time, merchantID *uint64) (interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			COUNT(CASE WHEN total_rights_cost > 0 THEN 1 END) as orders_with_rights,
			SUM(total_rights_cost) as total_rights_used,
			AVG(total_rights_cost) as avg_rights_per_order
		FROM orders o
		WHERE o.tenant_id = ? 
			AND o.created_at BETWEEN ? AND ?
			AND o.status IN ('completed', 'paid')
	`
	
	args := []interface{}{tenantID, startDate, endDate}
	
	if merchantID != nil {
		query += " AND o.merchant_id = ?"
		args = append(args, *merchantID)
	}
	
	var usage map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&usage)
	
	return usage, err
}

// getProductPerformance 获取商品销售表现数据
func (s *AnalyticsService) getProductPerformance(ctx context.Context, tenantID uint64, startDate, endDate time.Time, filters map[string]interface{}, merchantID *uint64) (interface{}, error) {
	query := `
		SELECT 
			p.id as product_id,
			p.name as product_name,
			c.name as category_name,
			COUNT(oi.id) as order_count,
			SUM(oi.quantity) as total_quantity,
			SUM(oi.price * oi.quantity) as total_revenue,
			AVG(oi.price) as avg_price,
			COUNT(DISTINCT o.customer_id) as customer_count
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.tenant_id = ? 
			AND o.created_at BETWEEN ? AND ?
			AND o.status IN ('completed', 'paid')
	`
	
	args := []interface{}{tenantID, startDate, endDate}
	
	if merchantID != nil {
		query += " AND p.merchant_id = ?"
		args = append(args, *merchantID)
	}
	
	// 应用过滤条件
	if filters != nil {
		if categoryID, ok := filters["category_id"]; ok {
			query += " AND p.category_id = ?"
			args = append(args, categoryID)
		}
		if minPrice, ok := filters["min_price"]; ok {
			query += " AND oi.price >= ?"
			args = append(args, minPrice)
		}
		if maxPrice, ok := filters["max_price"]; ok {
			query += " AND oi.price <= ?"
			args = append(args, maxPrice)
		}
	}
	
	query += " GROUP BY p.id, p.name, c.name ORDER BY total_revenue DESC LIMIT 50"
	
	var products []map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&products)
	
	return products, err
}

// getPaymentMethodStats 获取支付方式统计
func (s *AnalyticsService) getPaymentMethodStats(ctx context.Context, tenantID uint64, startDate, endDate time.Time, merchantID *uint64) (interface{}, error) {
	query := `
		SELECT 
			payment_method,
			COUNT(*) as transaction_count,
			SUM(total_amount) as total_amount,
			AVG(total_amount) as avg_amount,
			COUNT(*) * 100.0 / (SELECT COUNT(*) FROM orders 
								WHERE tenant_id = ? 
								AND created_at BETWEEN ? AND ?
								AND status IN ('completed', 'paid')) as percentage
		FROM orders o
		WHERE o.tenant_id = ? 
			AND o.created_at BETWEEN ? AND ?
			AND o.status IN ('completed', 'paid')
	`
	
	args := []interface{}{tenantID, startDate, endDate, tenantID, startDate, endDate}
	
	if merchantID != nil {
		query += " AND o.merchant_id = ?"
		args = append(args, *merchantID)
	}
	
	query += " GROUP BY payment_method ORDER BY transaction_count DESC"
	
	var paymentStats []map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&paymentStats)
	
	return paymentStats, err
}

// getGeographicAnalysis 获取地理分析数据
func (s *AnalyticsService) getGeographicAnalysis(ctx context.Context, tenantID uint64, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	query := `
		SELECT 
			COALESCE(u.province, '未知') as province,
			COALESCE(u.city, '未知') as city,
			COUNT(DISTINCT o.customer_id) as customer_count,
			COUNT(o.id) as order_count,
			SUM(o.total_amount) as total_revenue,
			AVG(o.total_amount) as avg_order_value
		FROM orders o
		LEFT JOIN users u ON o.customer_id = u.id
		WHERE o.tenant_id = ? 
			AND o.created_at BETWEEN ? AND ?
			AND o.status IN ('completed', 'paid')
	`
	
	args := []interface{}{tenantID, startDate, endDate}
	
	// 应用过滤条件
	if filters != nil {
		if province, ok := filters["province"]; ok {
			query += " AND u.province = ?"
			args = append(args, province)
		}
	}
	
	query += " GROUP BY u.province, u.city ORDER BY total_revenue DESC LIMIT 100"
	
	var geoStats []map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&geoStats)
	
	return geoStats, err
}

// getCustomerRetentionAnalysis 获取客户留存分析
func (s *AnalyticsService) getCustomerRetentionAnalysis(ctx context.Context, tenantID uint64, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	// 计算同期群留存率
	query := `
		SELECT 
			first_order_month,
			period_offset,
			initial_customers,
			retained_customers,
			retention_rate
		FROM (
			SELECT 
				first_order_month,
				period_offset,
				initial_customers,
				COUNT(*) as retained_customers,
				(COUNT(*) * 100.0 / initial_customers) as retention_rate
			FROM (
				SELECT DISTINCT
					c.customer_id,
					c.first_order_month,
					PERIOD_DIFF(DATE_FORMAT(o.created_at, '%Y%m'), 
								DATE_FORMAT(STR_TO_DATE(c.first_order_month, '%Y-%m'), '%Y%m')) as period_offset,
					c.initial_customers
				FROM (
					SELECT 
						customer_id,
						DATE_FORMAT(MIN(created_at), '%Y-%m') as first_order_month,
						COUNT(*) OVER (PARTITION BY DATE_FORMAT(MIN(created_at), '%Y-%m')) as initial_customers
					FROM orders
					WHERE tenant_id = ? 
						AND status IN ('completed', 'paid')
						AND created_at >= ?
					GROUP BY customer_id
				) c
				JOIN orders o ON c.customer_id = o.customer_id
				WHERE o.tenant_id = ? 
					AND o.status IN ('completed', 'paid')
					AND o.created_at <= ?
			) cohort_data
			GROUP BY first_order_month, period_offset, initial_customers
		) retention_analysis
		ORDER BY first_order_month, period_offset
		LIMIT 200
	`
	
	args := []interface{}{tenantID, startDate, tenantID, endDate}
	
	var retentionStats []map[string]interface{}
	err := g.DB().Raw(query, args...).Scan(&retentionStats)
	
	return retentionStats, err
}

// getSalesFunnelAnalysis 获取销售漏斗分析
func (s *AnalyticsService) getSalesFunnelAnalysis(ctx context.Context, tenantID uint64, startDate, endDate time.Time, merchantID *uint64) (interface{}, error) {
	// 模拟销售漏斗数据
	// 在真实场景中，这些数据可能来自不同的事件追踪表
	
	baseQuery := `
		SELECT 
			'orders_created' as stage,
			COUNT(*) as count,
			1.0 as conversion_rate
		FROM orders 
		WHERE tenant_id = ? 
			AND created_at BETWEEN ? AND ?
	`
	
	args := []interface{}{tenantID, startDate, endDate}
	
	if merchantID != nil {
		baseQuery += " AND merchant_id = ?"
		args = append(args, *merchantID)
	}
	
	funnelQuery := baseQuery + `
		UNION ALL
		SELECT 
			'orders_paid' as stage,
			COUNT(*) as count,
			COUNT(*) * 1.0 / (SELECT COUNT(*) FROM orders WHERE tenant_id = ? AND created_at BETWEEN ? AND ?) as conversion_rate
		FROM orders 
		WHERE tenant_id = ? 
			AND created_at BETWEEN ? AND ?
			AND status IN ('paid', 'processing', 'completed')
	`
	
	// 为支付订单添加参数
	paymentArgs := append(args, tenantID, startDate, endDate)
	paymentArgs = append(paymentArgs, args...)
	if merchantID != nil {
		funnelQuery += " AND merchant_id = ?"
		paymentArgs = append(paymentArgs, *merchantID)
	}
	
	funnelQuery += `
		UNION ALL
		SELECT 
			'orders_completed' as stage,
			COUNT(*) as count,
			COUNT(*) * 1.0 / (SELECT COUNT(*) FROM orders WHERE tenant_id = ? AND created_at BETWEEN ? AND ?) as conversion_rate
		FROM orders 
		WHERE tenant_id = ? 
			AND created_at BETWEEN ? AND ?
			AND status = 'completed'
	`
	
	// 为完成订单添加参数
	completedArgs := append(paymentArgs, tenantID, startDate, endDate, tenantID, startDate, endDate)
	if merchantID != nil {
		funnelQuery += " AND merchant_id = ?"
		completedArgs = append(completedArgs, *merchantID)
	}
	
	funnelQuery += " ORDER BY FIELD(stage, 'orders_created', 'orders_paid', 'orders_completed')"
	
	var funnelStats []map[string]interface{}
	err := g.DB().Raw(funnelQuery, completedArgs...).Scan(&funnelStats)
	
	return funnelStats, err
}
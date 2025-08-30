import { apiClient } from './api';

// 报表类型
export type ReportType = 'financial' | 'merchant_operation' | 'customer_analysis';
export type PeriodType = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly' | 'custom';
export type FileFormat = 'excel' | 'pdf' | 'json';
export type ReportStatus = 'generating' | 'completed' | 'failed';

// 报表生成请求
export interface ReportCreateRequest {
  report_type: ReportType;
  period_type: PeriodType;
  start_date: string;
  end_date: string;
  file_format: FileFormat;
  merchant_id?: number;
  config?: Record<string, any>;
}

// 报表信息
export interface Report {
  id: number;
  uuid: string;
  tenant_id: number;
  report_type: ReportType;
  period_type: PeriodType;
  start_date: string;
  end_date: string;
  status: ReportStatus;
  file_path?: string;
  file_format: FileFormat;
  generated_by: number;
  generated_at: string;
  expires_at?: string;
  data_summary?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// 报表列表请求
export interface ReportListRequest {
  report_type?: ReportType;
  status?: ReportStatus;
  start_date?: string;
  end_date?: string;
  page: number;
  page_size: number;
}

// 报表列表响应
export interface ReportListResponse {
  items: Report[];
  total: number;
  page: number;
  page_size: number;
  has_next: boolean;
}

// 财务报表数据
export interface FinancialReportData {
  total_revenue: { amount: number };
  total_expenditure: { amount: number };
  net_profit: { amount: number };
  rights_distributed: number;
  rights_consumed: number;
  rights_balance: number;
  merchant_count: number;
  active_merchant_count: number;
  customer_count: number;
  active_customer_count: number;
  order_count: number;
  order_amount: { amount: number };
  breakdown?: FinancialBreakdown;
}

// 财务分解数据
export interface FinancialBreakdown {
  revenue_by_merchant: MerchantRevenue[];
  revenue_by_category: CategoryRevenue[];
  expenditure_by_type: ExpenditureItem[];
  rights_by_category: RightsUsage[];
  monthly_trend: MonthlyFinancial[];
}

// 商户收入数据
export interface MerchantRevenue {
  merchant_id: number;
  merchant_name: string;
  revenue: { amount: number };
  order_count: number;
  percentage: number;
}

// 类别收入数据
export interface CategoryRevenue {
  category_id: number;
  category_name: string;
  revenue: { amount: number };
  order_count: number;
  percentage: number;
}

// 支出项目
export interface ExpenditureItem {
  type: string;
  amount: { amount: number };
  percentage: number;
  description: string;
}

// 权益使用数据
export interface RightsUsage {
  category_id: number;
  category_name: string;
  distributed: number;
  consumed: number;
  balance: number;
  utilization_rate: number;
}

// 月度财务数据
export interface MonthlyFinancial {
  month: string;
  revenue: { amount: number };
  expenditure: { amount: number };
  net_profit: { amount: number };
  order_count: number;
  rights_distributed: number;
  rights_consumed: number;
}

// 商户运营报表数据
export interface MerchantOperationReport {
  merchant_rankings: MerchantRanking[];
  performance_trends: MerchantTrend[];
  category_analysis: CategoryAnalysis[];
  growth_metrics?: GrowthMetrics;
}

// 商户排名数据
export interface MerchantRanking {
  rank: number;
  merchant_id: number;
  merchant_name: string;
  total_revenue: { amount: number };
  order_count: number;
  customer_count: number;
  average_order_value: { amount: number };
  growth_rate: number;
}

// 商户趋势数据
export interface MerchantTrend {
  merchant_id: number;
  merchant_name: string;
  trend_data: MonthlyTrendData[];
}

// 月度趋势数据
export interface MonthlyTrendData {
  month: string;
  revenue: { amount: number };
  order_count: number;
  customer_count: number;
}

// 类别分析数据
export interface CategoryAnalysis {
  category_id: number;
  category_name: string;
  revenue: { amount: number };
  order_count: number;
  merchant_count: number;
  market_share: number;
  growth_rate: number;
}

// 增长指标
export interface GrowthMetrics {
  revenue_growth_rate: number;
  order_growth_rate: number;
  merchant_growth_rate: number;
  customer_growth_rate: number;
  average_order_value_growth: number;
}

// 客户分析报表数据
export interface CustomerAnalysisReport {
  user_growth: UserGrowthData[];
  activity_metrics?: ActivityMetrics;
  consumption_behavior?: ConsumptionBehavior;
  retention_analysis?: RetentionAnalysis;
  churn_analysis?: ChurnAnalysis;
}

// 用户增长数据
export interface UserGrowthData {
  month: string;
  new_users: number;
  active_users: number;
  cumulative_users: number;
  retention_rate: number;
}

// 活跃度指标
export interface ActivityMetrics {
  dau: number; // 日活跃用户
  wau: number; // 周活跃用户
  mau: number; // 月活跃用户
  average_session_time: number; // 平均会话时长(分钟)
  average_order_freq: number; // 平均下单频次
}

// 消费行为
export interface ConsumptionBehavior {
  average_order_value: { amount: number }; // 平均客单价
  repurchase_rate: number; // 复购率
  average_order_count: number; // 平均下单次数
  preferred_categories: CategoryPreference[]; // 偏好类别
  payment_methods: PaymentMethodStats[]; // 支付方式统计
}

// 类别偏好
export interface CategoryPreference {
  category_id: number;
  category_name: string;
  order_count: number;
  revenue: { amount: number };
  percentage: number;
}

// 支付方式统计
export interface PaymentMethodStats {
  method: string;
  count: number;
  amount: { amount: number };
  percentage: number;
}

// 留存分析
export interface RetentionAnalysis {
  day1_retention: number;
  day7_retention: number;
  day30_retention: number;
  cohort_analysis: CohortData[];
}

// 同期群数据
export interface CohortData {
  cohort: string;
  users: number;
  retention_rates: number[];
}

// 流失分析
export interface ChurnAnalysis {
  churn_rate: number;
  churn_reasons: ChurnReason[];
  risk_user_count: number;
  churn_prediction: ChurnPrediction[];
}

// 流失原因
export interface ChurnReason {
  reason: string;
  count: number;
  percentage: number;
}

// 流失预测
export interface ChurnPrediction {
  user_id: number;
  username: string;
  churn_risk: number;
  last_active_date: string;
  recommendation: string;
}

// 自定义查询请求
export interface AnalyticsQueryRequest {
  metric_type: string;
  start_date: string;
  end_date: string;
  group_by?: string;
  filters?: Record<string, any>;
  merchant_id?: number;
}

class ReportService {
  // 生成报表
  async generateReport(request: ReportCreateRequest): Promise<Report> {
    const response = await apiClient.post('/reports/generate', request);
    return response.data;
  }

  // 获取报表信息
  async getReport(id: number): Promise<Report> {
    const response = await apiClient.get(`/reports/${id}`);
    return response.data;
  }

  // 获取报表列表
  async listReports(request: ReportListRequest): Promise<ReportListResponse> {
    const params = new URLSearchParams();
    if (request.report_type) params.set('report_type', request.report_type);
    if (request.status) params.set('status', request.status);
    if (request.start_date) params.set('start_date', request.start_date);
    if (request.end_date) params.set('end_date', request.end_date);
    params.set('page', request.page.toString());
    params.set('page_size', request.page_size.toString());

    const response = await apiClient.get(`/reports?${params.toString()}`);
    return response.data;
  }

  // 删除报表
  async deleteReport(id: number): Promise<void> {
    await apiClient.delete(`/reports/${id}`);
  }

  // 下载报表
  async downloadReport(uuid: string): Promise<Blob> {
    const response = await apiClient.get(`/reports/${uuid}/download`, {
      responseType: 'blob',
    });
    return response.data;
  }

  // 获取财务分析数据
  async getFinancialAnalytics(
    startDate: string,
    endDate: string,
    merchantId?: number
  ): Promise<FinancialReportData> {
    const params = new URLSearchParams();
    params.set('start_date', startDate);
    params.set('end_date', endDate);
    if (merchantId) params.set('merchant_id', merchantId.toString());

    const response = await apiClient.get(`/analytics/financial?${params.toString()}`);
    return response.data;
  }

  // 获取商户运营分析数据
  async getMerchantAnalytics(
    startDate: string,
    endDate: string
  ): Promise<MerchantOperationReport> {
    const params = new URLSearchParams();
    params.set('start_date', startDate);
    params.set('end_date', endDate);

    const response = await apiClient.get(`/analytics/merchants?${params.toString()}`);
    return response.data;
  }

  // 获取客户分析数据
  async getCustomerAnalytics(
    startDate: string,
    endDate: string
  ): Promise<CustomerAnalysisReport> {
    const params = new URLSearchParams();
    params.set('start_date', startDate);
    params.set('end_date', endDate);

    const response = await apiClient.get(`/analytics/customers?${params.toString()}`);
    return response.data;
  }

  // 自定义数据查询
  async customQuery(request: AnalyticsQueryRequest): Promise<any> {
    const response = await apiClient.post('/analytics/custom', request);
    return response.data;
  }

  // 获取趋势数据
  async getTrendData(
    metric: string,
    startDate: string,
    endDate: string,
    groupBy?: string,
    merchantId?: number
  ): Promise<any[]> {
    const params = new URLSearchParams();
    params.set('start_date', startDate);
    params.set('end_date', endDate);
    if (groupBy) params.set('group_by', groupBy);
    if (merchantId) params.set('merchant_id', merchantId.toString());

    const response = await apiClient.get(`/analytics/trends/${metric}?${params.toString()}`);
    return response.data;
  }

  // 清理缓存
  async clearCache(pattern?: string): Promise<void> {
    const params = new URLSearchParams();
    if (pattern) params.set('pattern', pattern);

    await apiClient.post(`/analytics/cache/clear?${params.toString()}`);
  }
}

export const reportService = new ReportService();
import { apiService } from './api';
import type {
  RightsUsageStats,
  RightsAlert,
  MonitoringDashboardData,
  RightsStatsQuery,
  RightsTrendsQuery,
  AlertListQuery,
  AlertConfigureRequest,
  AlertResolveRequest,
  ReportGenerateRequest,
  ApiResponse,
  PaginatedResponse
} from '../types/monitoring';

class MonitoringService {
  private readonly baseUrl = '/api/v1/monitoring';

  /**
   * 获取权益使用统计
   */
  async getRightsStats(query: RightsStatsQuery): Promise<RightsUsageStats[]> {
    const response = await apiService.get<ApiResponse<RightsUsageStats[]>>(
      `${this.baseUrl}/rights/stats`,
      { params: query }
    );
    return response.data?.data || [];
  }

  /**
   * 获取权益使用趋势
   */
  async getRightsTrends(query: RightsTrendsQuery): Promise<RightsUsageStats[]> {
    const response = await apiService.get<ApiResponse<RightsUsageStats[]>>(
      `${this.baseUrl}/rights/trends`,
      { params: query }
    );
    return response.data?.data || [];
  }

  /**
   * 配置预警阈值
   */
  async configureAlerts(request: AlertConfigureRequest): Promise<void> {
    await apiService.post<ApiResponse<void>>(
      `${this.baseUrl}/alerts/configure`,
      request
    );
  }

  /**
   * 获取预警列表
   */
  async listAlerts(query: AlertListQuery): Promise<PaginatedResponse<RightsAlert>> {
    const response = await apiService.get<ApiResponse<PaginatedResponse<RightsAlert>>>(
      `${this.baseUrl}/alerts`,
      { params: query }
    );
    return response.data?.data || { list: [], total: 0, page: 1, page_size: 10 };
  }

  /**
   * 解决预警
   */
  async resolveAlert(alertId: number, request: AlertResolveRequest): Promise<void> {
    await apiService.post<ApiResponse<void>>(
      `${this.baseUrl}/alerts/${alertId}/resolve`,
      request
    );
  }

  /**
   * 获取监控仪表板数据
   */
  async getDashboardData(merchantId?: number): Promise<MonitoringDashboardData> {
    const params = merchantId ? { merchant_id: merchantId } : {};
    const response = await apiService.get<ApiResponse<MonitoringDashboardData>>(
      `${this.baseUrl}/dashboard`,
      { params }
    );
    return response.data?.data || {
      total_merchants: 0,
      active_alerts: 0,
      total_rights_balance: 0,
      daily_consumption: 0,
      recent_alerts: [],
      usage_trends: [],
      consumption_chart_data: [],
      balance_distribution: []
    };
  }

  /**
   * 生成权益使用报告
   */
  async generateReport(request: ReportGenerateRequest): Promise<{ filename: string; download_url: string }> {
    const response = await apiService.post<ApiResponse<{ filename: string; download_url: string }>>(
      `${this.baseUrl}/reports/generate`,
      request
    );
    return response.data?.data || { filename: '', download_url: '' };
  }

  /**
   * 下载报告文件
   */
  async downloadReport(filename: string): Promise<Blob> {
    const response = await apiService.get(
      `/api/v1/reports/download/${filename}`,
      { responseType: 'blob' }
    );
    return response.data;
  }
}

export const monitoringService = new MonitoringService();
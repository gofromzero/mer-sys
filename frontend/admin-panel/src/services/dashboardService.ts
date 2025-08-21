import {
  MerchantDashboardData,
  RightsUsagePoint,
  PendingTask,
  NotificationsResponse,
  DashboardConfig,
  DashboardConfigRequest,
  ApiResponse,
  TimePeriod,
  RightsTrendParams,
  DashboardParams
} from '@/types/dashboard';
import { apiClient } from '@/lib/api';

/**
 * 商户仪表板API服务
 * 
 * 提供完整的仪表板数据获取和配置管理功能
 * 所有API调用都会自动添加JWT认证和租户信息
 */
class DashboardService {
  private baseUrl = '/api/v1/merchant/dashboard';

  /**
   * 获取商户仪表板核心数据
   * @returns 仪表板数据
   */
  async getMerchantDashboard(params?: DashboardParams): Promise<MerchantDashboardData> {
    try {
      const response = await apiClient.get<ApiResponse<MerchantDashboardData>>(
        this.baseUrl,
        { params }
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '获取仪表板数据失败');
      }
      
      return response.data.data!;
    } catch (error) {
      console.error('获取仪表板数据失败:', error);
      throw error;
    }
  }

  /**
   * 获取指定时间段业务统计
   * @param period 时间周期
   * @returns 业务统计数据
   */
  async getMerchantStats(period: TimePeriod): Promise<MerchantDashboardData> {
    try {
      const response = await apiClient.get<ApiResponse<MerchantDashboardData>>(
        `${this.baseUrl}/stats/${period}`
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '获取统计数据失败');
      }
      
      return response.data.data!;
    } catch (error) {
      console.error('获取统计数据失败:', error);
      throw error;
    }
  }

  /**
   * 获取权益使用趋势数据
   * @param params 查询参数
   * @returns 权益趋势数据
   */
  async getRightsUsageTrend(params?: RightsTrendParams): Promise<RightsUsagePoint[]> {
    try {
      const queryParams = {
        days: params?.days || 30
      };
      
      const response = await apiClient.get<ApiResponse<RightsUsagePoint[]>>(
        `${this.baseUrl}/rights-trend`,
        { params: queryParams }
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '获取权益趋势失败');
      }
      
      return response.data.data!;
    } catch (error) {
      console.error('获取权益趋势失败:', error);
      throw error;
    }
  }

  /**
   * 获取待处理事项汇总
   * @returns 待处理任务列表
   */
  async getPendingTasks(): Promise<PendingTask[]> {
    try {
      const response = await apiClient.get<ApiResponse<PendingTask[]>>(
        `${this.baseUrl}/pending-tasks`
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '获取待处理事项失败');
      }
      
      return response.data.data!;
    } catch (error) {
      console.error('获取待处理事项失败:', error);
      throw error;
    }
  }

  /**
   * 获取系统通知和公告
   * @returns 通知和公告数据
   */
  async getNotifications(): Promise<NotificationsResponse> {
    try {
      const response = await apiClient.get<ApiResponse<NotificationsResponse>>(
        `${this.baseUrl}/notifications`
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '获取通知公告失败');
      }
      
      return response.data.data!;
    } catch (error) {
      console.error('获取通知公告失败:', error);
      throw error;
    }
  }

  /**
   * 获取仪表板个性化配置
   * @returns 仪表板配置
   */
  async getDashboardConfig(): Promise<DashboardConfig> {
    try {
      const response = await apiClient.get<ApiResponse<DashboardConfig>>(
        `${this.baseUrl}/config`
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '获取仪表板配置失败');
      }
      
      return response.data.data!;
    } catch (error) {
      console.error('获取仪表板配置失败:', error);
      throw error;
    }
  }

  /**
   * 保存仪表板个性化配置
   * @param config 配置数据
   */
  async saveDashboardConfig(config: DashboardConfigRequest): Promise<void> {
    try {
      const response = await apiClient.post<ApiResponse>(
        `${this.baseUrl}/config`,
        config
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '保存仪表板配置失败');
      }
    } catch (error) {
      console.error('保存仪表板配置失败:', error);
      throw error;
    }
  }

  /**
   * 更新仪表板布局配置
   * @param config 配置数据
   */
  async updateDashboardConfig(config: DashboardConfigRequest): Promise<void> {
    try {
      const response = await apiClient.put<ApiResponse>(
        `${this.baseUrl}/config`,
        config
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '更新仪表板配置失败');
      }
    } catch (error) {
      console.error('更新仪表板配置失败:', error);
      throw error;
    }
  }

  /**
   * 标记公告为已读
   * @param announcementId 公告ID
   */
  async markAnnouncementAsRead(announcementId: number): Promise<void> {
    try {
      const response = await apiClient.post<ApiResponse>(
        `${this.baseUrl}/announcements/${announcementId}/read`
      );
      
      if (response.data.code !== 200) {
        throw new Error(response.data.message || '标记公告已读失败');
      }
    } catch (error) {
      console.error('标记公告已读失败:', error);
      throw error;
    }
  }

  /**
   * 批量获取仪表板所有数据
   * @param period 时间周期
   * @returns 包含所有仪表板数据的对象
   */
  async getDashboardAllData(period: TimePeriod = TimePeriod.DAILY) {
    try {
      const [
        dashboardData,
        rightsUsageTrend,
        pendingTasks,
        notifications,
        config
      ] = await Promise.all([
        this.getMerchantStats(period),
        this.getRightsUsageTrend({ days: 30 }),
        this.getPendingTasks(),
        this.getNotifications(),
        this.getDashboardConfig()
      ]);

      return {
        dashboard: dashboardData,
        trends: rightsUsageTrend,
        tasks: pendingTasks,
        notifications,
        config
      };
    } catch (error) {
      console.error('批量获取仪表板数据失败:', error);
      throw error;
    }
  }

  /**
   * 刷新仪表板数据 (清除缓存)
   */
  async refreshDashboard(): Promise<MerchantDashboardData> {
    return this.getMerchantDashboard({ refresh: true });
  }
}

// 创建单例实例
export const dashboardService = new DashboardService();

// 导出默认实例
export default dashboardService;
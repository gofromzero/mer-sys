import { renderHook, act } from '@testing-library/react';
import { useMonitoringStore } from '../monitoringStore';
import { monitoringService } from '../../services/monitoringService';
import { AlertType, AlertSeverity, AlertStatus, TimePeriod, TrendDirection } from '../../types/monitoring';

// Mock monitoring service
jest.mock('../../services/monitoringService', () => ({
  monitoringService: {
    getDashboardData: jest.fn(),
    listAlerts: jest.fn(),
    configureAlerts: jest.fn(),
    resolveAlert: jest.fn(),
    getRightsStats: jest.fn(),
    getRightsTrends: jest.fn(),
    generateReport: jest.fn(),
  },
}));

const mockMonitoringService = monitoringService as jest.Mocked<typeof monitoringService>;

const mockDashboardData = {
  total_merchants: 5,
  active_alerts: 3,
  total_rights_balance: 50000.0,
  daily_consumption: 1500.0,
  recent_alerts: [],
  usage_trends: [],
  consumption_chart_data: [],
  balance_distribution: [],
};

const mockAlerts = [
  {
    id: 1,
    tenant_id: 1,
    merchant_id: 1,
    alert_type: AlertType.BALANCE_LOW,
    threshold_value: 1000.0,
    current_value: 800.0,
    severity: AlertSeverity.WARNING,
    status: AlertStatus.ACTIVE,
    message: '商户余额不足',
    triggered_at: '2025-08-21T10:00:00Z',
    notified_channels: ['email'],
    created_at: '2025-08-21T10:00:00Z',
    updated_at: '2025-08-21T10:00:00Z',
  },
];

const mockUsageStats = [
  {
    tenant_id: 1,
    merchant_id: 1,
    stat_date: '2025-08-21',
    period: TimePeriod.DAILY,
    total_allocated: 10000.0,
    total_consumed: 3000.0,
    average_daily_usage: 1000.0,
    usage_trend: TrendDirection.STABLE,
    created_at: '2025-08-21T10:00:00Z',
  },
];

describe('useMonitoringStore', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Reset store state
    useMonitoringStore.setState({
      dashboardData: null,
      alerts: [],
      alertsTotal: 0,
      usageStats: [],
      trendData: [],
      loading: {
        dashboard: false,
        alerts: false,
        stats: false,
        trends: false,
        configuring: false,
        resolving: false,
        generating: false,
      },
      error: null,
      pagination: {
        page: 1,
        pageSize: 10,
      },
      filters: {},
    });
  });

  describe('fetchDashboardData', () => {
    it('应该成功获取仪表板数据', async () => {
      mockMonitoringService.getDashboardData.mockResolvedValue(mockDashboardData);

      const { result } = renderHook(() => useMonitoringStore());

      await act(async () => {
        await result.current.fetchDashboardData();
      });

      expect(mockMonitoringService.getDashboardData).toHaveBeenCalledWith(undefined);
      expect(result.current.dashboardData).toEqual(mockDashboardData);
      expect(result.current.loading.dashboard).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('应该正确处理获取仪表板数据时的错误', async () => {
      const errorMessage = '获取仪表板数据失败';
      mockMonitoringService.getDashboardData.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMonitoringStore());

      await act(async () => {
        await result.current.fetchDashboardData();
      });

      expect(result.current.dashboardData).toBeNull();
      expect(result.current.error).toBe(errorMessage);
      expect(result.current.loading.dashboard).toBe(false);
    });

    it('应该正确传递商户ID参数', async () => {
      mockMonitoringService.getDashboardData.mockResolvedValue(mockDashboardData);

      const { result } = renderHook(() => useMonitoringStore());

      await act(async () => {
        await result.current.fetchDashboardData(123);
      });

      expect(mockMonitoringService.getDashboardData).toHaveBeenCalledWith(123);
    });
  });

  describe('fetchAlerts', () => {
    it('应该成功获取预警列表', async () => {
      const mockResponse = {
        list: mockAlerts,
        total: 1,
        page: 1,
        page_size: 10,
      };
      mockMonitoringService.listAlerts.mockResolvedValue(mockResponse);

      const { result } = renderHook(() => useMonitoringStore());

      await act(async () => {
        await result.current.fetchAlerts();
      });

      expect(mockMonitoringService.listAlerts).toHaveBeenCalled();
      expect(result.current.alerts).toEqual(mockAlerts);
      expect(result.current.alertsTotal).toBe(1);
      expect(result.current.loading.alerts).toBe(false);
    });

    it('应该正确处理获取预警列表时的错误', async () => {
      const errorMessage = '获取预警列表失败';
      mockMonitoringService.listAlerts.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMonitoringStore());

      await act(async () => {
        await result.current.fetchAlerts();
      });

      expect(result.current.error).toBe(errorMessage);
      expect(result.current.loading.alerts).toBe(false);
    });
  });

  describe('configureAlerts', () => {
    it('应该成功配置预警', async () => {
      mockMonitoringService.configureAlerts.mockResolvedValue(undefined);
      mockMonitoringService.getDashboardData.mockResolvedValue(mockDashboardData);

      const { result } = renderHook(() => useMonitoringStore());

      const configRequest = {
        merchant_id: 1,
        warning_threshold: 1000.0,
        critical_threshold: 500.0,
      };

      await act(async () => {
        await result.current.configureAlerts(configRequest);
      });

      expect(mockMonitoringService.configureAlerts).toHaveBeenCalledWith(configRequest);
      expect(mockMonitoringService.getDashboardData).toHaveBeenCalled(); // 应该刷新仪表板
      expect(result.current.loading.configuring).toBe(false);
    });

    it('应该正确处理配置预警时的错误', async () => {
      const errorMessage = '配置预警失败';
      mockMonitoringService.configureAlerts.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMonitoringStore());

      const configRequest = {
        merchant_id: 1,
        warning_threshold: 1000.0,
      };

      await act(async () => {
        try {
          await result.current.configureAlerts(configRequest);
        } catch (error) {
          // 预期会抛出错误
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(result.current.loading.configuring).toBe(false);
    });
  });

  describe('resolveAlert', () => {
    it('应该成功解决预警', async () => {
      mockMonitoringService.resolveAlert.mockResolvedValue(undefined);
      mockMonitoringService.getDashboardData.mockResolvedValue(mockDashboardData);

      const { result } = renderHook(() => useMonitoringStore());
      
      // 先设置一些预警数据
      act(() => {
        useMonitoringStore.setState({ alerts: mockAlerts });
      });

      await act(async () => {
        await result.current.resolveAlert(1, '手动充值解决');
      });

      expect(mockMonitoringService.resolveAlert).toHaveBeenCalledWith(1, { resolution: '手动充值解决' });
      expect(mockMonitoringService.getDashboardData).toHaveBeenCalled(); // 应该刷新仪表板
      
      // 检查本地状态是否更新
      const updatedAlert = result.current.alerts.find(alert => alert.id === 1);
      expect(updatedAlert?.status).toBe('resolved');
      expect(result.current.loading.resolving).toBe(false);
    });

    it('应该正确处理解决预警时的错误', async () => {
      const errorMessage = '解决预警失败';
      mockMonitoringService.resolveAlert.mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useMonitoringStore());

      await act(async () => {
        try {
          await result.current.resolveAlert(1, '解决方案');
        } catch (error) {
          // 预期会抛出错误
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(result.current.loading.resolving).toBe(false);
    });
  });

  describe('fetchUsageStats', () => {
    it('应该成功获取使用统计', async () => {
      mockMonitoringService.getRightsStats.mockResolvedValue(mockUsageStats);

      const { result } = renderHook(() => useMonitoringStore());

      const query = {
        period: TimePeriod.DAILY,
        page: 1,
        page_size: 10,
      };

      await act(async () => {
        await result.current.fetchUsageStats(query);
      });

      expect(mockMonitoringService.getRightsStats).toHaveBeenCalledWith(query);
      expect(result.current.usageStats).toEqual(mockUsageStats);
      expect(result.current.loading.stats).toBe(false);
    });
  });

  describe('generateReport', () => {
    it('应该成功生成报告', async () => {
      const mockReportResult = {
        filename: 'report_20250821.xlsx',
        download_url: '/downloads/report_20250821.xlsx',
      };
      mockMonitoringService.generateReport.mockResolvedValue(mockReportResult);

      const { result } = renderHook(() => useMonitoringStore());

      const reportRequest = {
        period: TimePeriod.MONTHLY,
        start_date: '2025-08-01',
        end_date: '2025-08-31',
        merchant_ids: [1, 2],
        format: 'excel' as const,
      };

      let reportResult;
      await act(async () => {
        reportResult = await result.current.generateReport(reportRequest);
      });

      expect(mockMonitoringService.generateReport).toHaveBeenCalledWith(reportRequest);
      expect(reportResult).toEqual(mockReportResult);
      expect(result.current.loading.generating).toBe(false);
    });
  });

  describe('state management', () => {
    it('应该正确设置筛选条件', () => {
      const { result } = renderHook(() => useMonitoringStore());

      act(() => {
        result.current.setFilters({
          alert_type: AlertType.BALANCE_LOW,
          severity: AlertSeverity.WARNING,
        });
      });

      expect(result.current.filters.alert_type).toBe(AlertType.BALANCE_LOW);
      expect(result.current.filters.severity).toBe(AlertSeverity.WARNING);
      expect(result.current.pagination.page).toBe(1); // 应该重置到第一页
    });

    it('应该正确设置分页信息', () => {
      const { result } = renderHook(() => useMonitoringStore());

      act(() => {
        result.current.setPagination(2, 20);
      });

      expect(result.current.pagination.page).toBe(2);
      expect(result.current.pagination.pageSize).toBe(20);
    });

    it('应该正确清除错误', () => {
      const { result } = renderHook(() => useMonitoringStore());

      // 先设置一个错误
      act(() => {
        useMonitoringStore.setState({ error: '测试错误' });
      });

      expect(result.current.error).toBe('测试错误');

      act(() => {
        result.current.clearError();
      });

      expect(result.current.error).toBeNull();
    });

    it('应该正确重置状态', () => {
      const { result } = renderHook(() => useMonitoringStore());

      // 先设置一些状态
      act(() => {
        useMonitoringStore.setState({
          dashboardData: mockDashboardData,
          alerts: mockAlerts,
          error: '测试错误',
          filters: { alert_type: AlertType.BALANCE_LOW },
        });
      });

      act(() => {
        result.current.resetState();
      });

      expect(result.current.dashboardData).toBeNull();
      expect(result.current.alerts).toEqual([]);
      expect(result.current.error).toBeNull();
      expect(result.current.filters).toEqual({});
    });
  });
});
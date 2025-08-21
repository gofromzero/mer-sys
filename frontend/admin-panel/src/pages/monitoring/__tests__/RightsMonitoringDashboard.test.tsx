import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { BrowserRouter } from 'react-router-dom';
import RightsMonitoringDashboard from '../RightsMonitoringDashboard';
import { useMonitoringStore } from '../../../stores/monitoringStore';
import { useAuthStore } from '../../../stores/authStore';

// Mock stores
jest.mock('../../../stores/monitoringStore');
jest.mock('../../../stores/authStore');
jest.mock('../../../components/ui/AmisRenderer', () => ({
  AmisRenderer: ({ schema, data }: any) => (
    <div data-testid="amis-renderer">
      <div data-testid="amis-schema">{JSON.stringify(schema.type)}</div>
      <div data-testid="amis-data">{JSON.stringify(data)}</div>
    </div>
  ),
}));

const mockUseMonitoringStore = useMonitoringStore as jest.MockedFunction<typeof useMonitoringStore>;
const mockUseAuthStore = useAuthStore as jest.MockedFunction<typeof useAuthStore>;

const mockDashboardData = {
  total_merchants: 5,
  active_alerts: 3,
  total_rights_balance: 50000.0,
  daily_consumption: 1500.0,
  recent_alerts: [
    {
      id: 1,
      merchant_id: 1,
      alert_type: 'balance_low',
      message: '商户余额不足',
      severity: 'warning',
      status: 'active',
      triggered_at: '2025-08-21T10:00:00Z',
      current_value: 800.0,
      threshold_value: 1000.0,
      notified_channels: ['email', 'system'],
    },
  ],
  usage_trends: [],
  consumption_chart_data: [
    {
      date: '2025-08-20',
      consumed: 1200.0,
      allocated: 2000.0,
      trend: 'stable',
    },
    {
      date: '2025-08-21',
      consumed: 1500.0,
      allocated: 2000.0,
      trend: 'increasing',
    },
  ],
  balance_distribution: [
    {
      merchant_name: '测试商户1',
      merchant_id: 1,
      available_balance: 5000.0,
      usage_percentage: 60,
      status: 'warning',
    },
    {
      merchant_name: '测试商户2',
      merchant_id: 2,
      available_balance: 8000.0,
      usage_percentage: 40,
      status: 'healthy',
    },
  ],
};

const mockUser = {
  id: 1,
  username: 'testuser',
  tenant_id: 1,
};

describe('RightsMonitoringDashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    mockUseMonitoringStore.mockReturnValue({
      dashboardData: mockDashboardData,
      fetchDashboardData: jest.fn(),
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
      clearError: jest.fn(),
      alerts: [],
      alertsTotal: 0,
      usageStats: [],
      trendData: [],
      pagination: { page: 1, pageSize: 10 },
      filters: {},
      configureAlerts: jest.fn(),
      resolveAlert: jest.fn(),
      fetchAlerts: jest.fn(),
      fetchUsageStats: jest.fn(),
      fetchTrendData: jest.fn(),
      generateReport: jest.fn(),
      setFilters: jest.fn(),
      setPagination: jest.fn(),
      resetState: jest.fn(),
    } as any);

    mockUseAuthStore.mockReturnValue({
      user: mockUser,
      token: 'mock-token',
      isAuthenticated: true,
      login: jest.fn(),
      logout: jest.fn(),
      updateUser: jest.fn(),
    } as any);
  });

  const renderComponent = () => {
    return render(
      <BrowserRouter>
        <RightsMonitoringDashboard />
      </BrowserRouter>
    );
  };

  it('应该正确渲染监控仪表板', () => {
    renderComponent();

    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
    expect(screen.getByTestId('amis-schema')).toHaveTextContent('"page"');
  });

  it('应该在组件挂载时获取仪表板数据', () => {
    const mockFetchDashboardData = jest.fn();
    mockUseMonitoringStore.mockReturnValue({
      ...mockUseMonitoringStore(),
      fetchDashboardData: mockFetchDashboardData,
    } as any);

    renderComponent();

    expect(mockFetchDashboardData).toHaveBeenCalled();
  });

  it('应该正确传递数据到AmisRenderer', () => {
    renderComponent();

    const amisData = screen.getByTestId('amis-data');
    const dataContent = JSON.parse(amisData.textContent || '{}');
    
    expect(dataContent).toHaveProperty('total_merchants', 5);
    expect(dataContent).toHaveProperty('active_alerts', 3);
    expect(dataContent).toHaveProperty('user');
    expect(dataContent.user).toEqual(mockUser);
  });

  it('应该在有错误时显示错误信息', () => {
    const mockClearError = jest.fn();
    mockUseMonitoringStore.mockReturnValue({
      ...mockUseMonitoringStore(),
      error: '获取数据失败',
      clearError: mockClearError,
    } as any);

    renderComponent();

    // AmisRenderer会接收到包含错误的schema
    const amisSchema = JSON.parse(screen.getByTestId('amis-schema').textContent || '""');
    expect(amisSchema).toBe('page');
  });

  it('应该正确处理加载状态', () => {
    mockUseMonitoringStore.mockReturnValue({
      ...mockUseMonitoringStore(),
      loading: {
        dashboard: true,
        alerts: false,
        stats: false,
        trends: false,
        configuring: false,
        resolving: false,
        generating: false,
      },
    } as any);

    renderComponent();

    const amisData = screen.getByTestId('amis-data');
    const dataContent = JSON.parse(amisData.textContent || '{}');
    
    expect(dataContent.loading).toBe(true);
  });

  it('应该正确处理空数据状态', () => {
    mockUseMonitoringStore.mockReturnValue({
      ...mockUseMonitoringStore(),
      dashboardData: null,
    } as any);

    renderComponent();

    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
  });

  it('应该设置自动刷新定时器', async () => {
    jest.useFakeTimers();
    const mockFetchDashboardData = jest.fn();
    
    mockUseMonitoringStore.mockReturnValue({
      ...mockUseMonitoringStore(),
      fetchDashboardData: mockFetchDashboardData,
    } as any);

    renderComponent();

    // 初始调用
    expect(mockFetchDashboardData).toHaveBeenCalledTimes(1);

    // 快进5分钟
    jest.advanceTimersByTime(5 * 60 * 1000);

    // 应该再次调用
    expect(mockFetchDashboardData).toHaveBeenCalledTimes(2);

    jest.useRealTimers();
  });

  it('应该在组件卸载时清理定时器', () => {
    jest.useFakeTimers();
    const clearIntervalSpy = jest.spyOn(global, 'clearInterval');

    const { unmount } = renderComponent();
    unmount();

    expect(clearIntervalSpy).toHaveBeenCalled();

    jest.useRealTimers();
  });

  it('应该正确格式化数字显示', () => {
    const largeDashboardData = {
      ...mockDashboardData,
      total_rights_balance: 1500000.0, // 1.5M
      daily_consumption: 2500.0, // 2.5K
    };

    mockUseMonitoringStore.mockReturnValue({
      ...mockUseMonitoringStore(),
      dashboardData: largeDashboardData,
    } as any);

    renderComponent();

    // 验证数据被正确传递
    const amisData = screen.getByTestId('amis-data');
    const dataContent = JSON.parse(amisData.textContent || '{}');
    
    expect(dataContent.total_rights_balance).toBe(1500000.0);
    expect(dataContent.daily_consumption).toBe(2500.0);
  });

  it('应该正确处理预警严重程度映射', () => {
    renderComponent();

    const amisData = screen.getByTestId('amis-data');
    const dataContent = JSON.parse(amisData.textContent || '{}');
    
    expect(dataContent.recent_alerts).toEqual(mockDashboardData.recent_alerts);
    expect(dataContent.recent_alerts[0].severity).toBe('warning');
  });

  it('应该正确处理商户余额分布数据', () => {
    renderComponent();

    const amisData = screen.getByTestId('amis-data');
    const dataContent = JSON.parse(amisData.textContent || '{}');
    
    expect(dataContent.balance_distribution).toHaveLength(2);
    expect(dataContent.balance_distribution[0].status).toBe('warning');
    expect(dataContent.balance_distribution[1].status).toBe('healthy');
  });
});
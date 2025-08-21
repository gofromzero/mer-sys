import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { 
  RightsAlert, 
  RightsUsageStats, 
  MonitoringDashboardData,
  AlertListQuery,
  RightsStatsQuery,
  RightsTrendsQuery,
  AlertConfigureRequest,
  AlertResolveRequest,
  ReportGenerateRequest,
  AlertStatus
} from '../types/monitoring';
import { monitoringService } from '../services/monitoringService';

interface MonitoringState {
  // 数据状态
  dashboardData: MonitoringDashboardData | null;
  alerts: RightsAlert[];
  alertsTotal: number;
  usageStats: RightsUsageStats[];
  trendData: RightsUsageStats[];
  
  // 加载状态
  loading: {
    dashboard: boolean;
    alerts: boolean;
    stats: boolean;
    trends: boolean;
    configuring: boolean;
    resolving: boolean;
    generating: boolean;
  };
  
  // 错误状态
  error: string | null;
  
  // 分页状态
  pagination: {
    page: number;
    pageSize: number;
  };
  
  // 筛选状态
  filters: Partial<AlertListQuery>;
}

interface MonitoringActions {
  // 仪表板相关
  fetchDashboardData: (merchantId?: number) => Promise<void>;
  
  // 预警相关
  fetchAlerts: (query?: Partial<AlertListQuery>) => Promise<void>;
  configureAlerts: (request: AlertConfigureRequest) => Promise<void>;
  resolveAlert: (alertId: number, resolution: string) => Promise<void>;
  
  // 统计相关
  fetchUsageStats: (query: RightsStatsQuery) => Promise<void>;
  fetchTrendData: (query: RightsTrendsQuery) => Promise<void>;
  
  // 报告相关
  generateReport: (request: ReportGenerateRequest) => Promise<{ filename: string; download_url: string }>;
  
  // 状态管理
  setFilters: (filters: Partial<AlertListQuery>) => void;
  setPagination: (page: number, pageSize: number) => void;
  clearError: () => void;
  resetState: () => void;
}

const initialState: MonitoringState = {
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
};

export const useMonitoringStore = create<MonitoringState & MonitoringActions>()(
  devtools(
    (set, get) => ({
      ...initialState,

      fetchDashboardData: async (merchantId?: number) => {
        set((state) => ({ 
          loading: { ...state.loading, dashboard: true },
          error: null 
        }));

        try {
          const dashboardData = await monitoringService.getDashboardData(merchantId);
          set({ dashboardData });
        } catch (error) {
          set({ error: error instanceof Error ? error.message : '获取仪表板数据失败' });
        } finally {
          set((state) => ({ 
            loading: { ...state.loading, dashboard: false }
          }));
        }
      },

      fetchAlerts: async (query?: Partial<AlertListQuery>) => {
        set((state) => ({ 
          loading: { ...state.loading, alerts: true },
          error: null 
        }));

        try {
          const { pagination, filters } = get();
          const fullQuery: AlertListQuery = {
            page: pagination.page,
            page_size: pagination.pageSize,
            ...filters,
            ...query,
          };

          const result = await monitoringService.listAlerts(fullQuery);
          set({ 
            alerts: result.list,
            alertsTotal: result.total,
          });
        } catch (error) {
          set({ error: error instanceof Error ? error.message : '获取预警列表失败' });
        } finally {
          set((state) => ({ 
            loading: { ...state.loading, alerts: false }
          }));
        }
      },

      configureAlerts: async (request: AlertConfigureRequest) => {
        set((state) => ({ 
          loading: { ...state.loading, configuring: true },
          error: null 
        }));

        try {
          await monitoringService.configureAlerts(request);
          // 重新获取仪表板数据以刷新状态
          await get().fetchDashboardData();
        } catch (error) {
          set({ error: error instanceof Error ? error.message : '配置预警失败' });
          throw error;
        } finally {
          set((state) => ({ 
            loading: { ...state.loading, configuring: false }
          }));
        }
      },

      resolveAlert: async (alertId: number, resolution: string) => {
        set((state) => ({ 
          loading: { ...state.loading, resolving: true },
          error: null 
        }));

        try {
          await monitoringService.resolveAlert(alertId, { resolution });
          
          // 更新本地状态
          set((state) => ({
            alerts: state.alerts.map(alert => 
              alert.id === alertId 
                ? { ...alert, status: AlertStatus.RESOLVED, resolved_at: new Date().toISOString() }
                : alert
            )
          }));

          // 重新获取仪表板数据
          await get().fetchDashboardData();
        } catch (error) {
          set({ error: error instanceof Error ? error.message : '解决预警失败' });
          throw error;
        } finally {
          set((state) => ({ 
            loading: { ...state.loading, resolving: false }
          }));
        }
      },

      fetchUsageStats: async (query: RightsStatsQuery) => {
        set((state) => ({ 
          loading: { ...state.loading, stats: true },
          error: null 
        }));

        try {
          const usageStats = await monitoringService.getRightsStats(query);
          set({ usageStats });
        } catch (error) {
          set({ error: error instanceof Error ? error.message : '获取使用统计失败' });
        } finally {
          set((state) => ({ 
            loading: { ...state.loading, stats: false }
          }));
        }
      },

      fetchTrendData: async (query: RightsTrendsQuery) => {
        set((state) => ({ 
          loading: { ...state.loading, trends: true },
          error: null 
        }));

        try {
          const trendData = await monitoringService.getRightsTrends(query);
          set({ trendData });
        } catch (error) {
          set({ error: error instanceof Error ? error.message : '获取趋势数据失败' });
        } finally {
          set((state) => ({ 
            loading: { ...state.loading, trends: false }
          }));
        }
      },

      generateReport: async (request: ReportGenerateRequest) => {
        set((state) => ({ 
          loading: { ...state.loading, generating: true },
          error: null 
        }));

        try {
          const result = await monitoringService.generateReport(request);
          return result;
        } catch (error) {
          set({ error: error instanceof Error ? error.message : '生成报告失败' });
          throw error;
        } finally {
          set((state) => ({ 
            loading: { ...state.loading, generating: false }
          }));
        }
      },

      setFilters: (filters: Partial<AlertListQuery>) => {
        set((state) => ({
          filters: { ...state.filters, ...filters },
          pagination: { ...state.pagination, page: 1 }, // 重置到第一页
        }));
      },

      setPagination: (page: number, pageSize: number) => {
        set((state) => ({
          pagination: { ...state.pagination, page, pageSize },
        }));
      },

      clearError: () => {
        set({ error: null });
      },

      resetState: () => {
        set(initialState);
      },
    }),
    {
      name: 'monitoring-store',
    }
  )
);
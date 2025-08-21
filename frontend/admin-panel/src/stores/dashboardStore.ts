import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import {
  MerchantDashboardData,
  RightsUsagePoint,
  PendingTask,
  NotificationsResponse,
  DashboardConfig,
  DashboardConfigRequest,
  TimePeriod,
  LoadingState,
  RefreshConfig,
  DashboardError,
  DashboardThemeType
} from '../types/dashboard';
import { dashboardService } from '../services/dashboardService';

/**
 * 仪表板状态接口
 */
interface DashboardState {
  // 数据状态
  dashboardData: MerchantDashboardData | null;
  rightsUsageTrend: RightsUsagePoint[];
  pendingTasks: PendingTask[];
  notifications: NotificationsResponse | null;
  config: DashboardConfig | null;
  
  // UI状态
  loading: LoadingState;
  error: DashboardError | null;
  currentPeriod: TimePeriod;
  theme: DashboardThemeType;
  
  // 刷新配置
  refreshConfig: RefreshConfig;
  refreshTimer: NodeJS.Timeout | null;
  
  // 操作方法
  setLoading: (key: keyof LoadingState, loading: boolean) => void;
  setError: (error: DashboardError | null) => void;
  setPeriod: (period: TimePeriod) => void;
  setTheme: (theme: DashboardThemeType) => void;
  
  // 数据加载方法
  loadDashboardData: (period?: TimePeriod) => Promise<void>;
  loadRightsUsageTrend: (days?: number) => Promise<void>;
  loadPendingTasks: () => Promise<void>;
  loadNotifications: () => Promise<void>;
  loadConfig: () => Promise<void>;
  loadAllData: (period?: TimePeriod) => Promise<void>;
  
  // 配置管理方法
  updateConfig: (config: DashboardConfigRequest) => Promise<void>;
  saveConfig: (config: DashboardConfigRequest) => Promise<void>;
  
  // 通知操作方法
  markAnnouncementAsRead: (announcementId: number) => Promise<void>;
  
  // 刷新控制方法
  startAutoRefresh: () => void;
  stopAutoRefresh: () => void;
  refreshDashboard: () => Promise<void>;
  
  // 重置方法
  reset: () => void;
}

/**
 * 初始状态
 */
const initialState = {
  // 数据状态
  dashboardData: null,
  rightsUsageTrend: [],
  pendingTasks: [],
  notifications: null,
  config: null,
  
  // UI状态
  loading: {
    dashboard: false,
    stats: false,
    trends: false,
    tasks: false,
    notifications: false,
    config: false
  } as LoadingState,
  error: null,
  currentPeriod: TimePeriod.DAILY,
  theme: 'light' as DashboardThemeType,
  
  // 刷新配置
  refreshConfig: {
    interval: 300000, // 5分钟
    autoRefresh: true,
    lastRefresh: 0
  } as RefreshConfig,
  refreshTimer: null
};

/**
 * 创建仪表板Store
 */
export const useDashboardStore = create<DashboardState>()(
  devtools(
    (set, get) => ({
      ...initialState,

      // UI状态更新方法
      setLoading: (key: keyof LoadingState, loading: boolean) => {
        set(state => ({
          loading: { ...state.loading, [key]: loading }
        }), false, `setLoading:${String(key)}:${loading}`);
      },

      setError: (error: DashboardError | null) => {
        set({ error }, false, 'setError');
      },

      setPeriod: (period: TimePeriod) => {
        set({ currentPeriod: period }, false, 'setPeriod');
        // 切换周期时重新加载数据
        get().loadDashboardData(period);
      },

      setTheme: (theme: DashboardThemeType) => {
        set({ theme }, false, 'setTheme');
        // 这里可以添加主题切换的副作用，如更新CSS变量
      },

      // 数据加载方法
      loadDashboardData: async (period?: TimePeriod) => {
        const { setLoading, setError, currentPeriod } = get();
        const targetPeriod = period || currentPeriod;
        
        try {
          setLoading('dashboard', true);
          setError(null);
          
          const data = await dashboardService.getMerchantStats(targetPeriod);
          
          set({
            dashboardData: data,
            currentPeriod: targetPeriod,
            refreshConfig: {
              ...get().refreshConfig,
              lastRefresh: Date.now()
            }
          }, false, 'loadDashboardData:success');
          
        } catch (error: any) {
          console.error('加载仪表板数据失败:', error);
          setError({
            code: 'DASHBOARD_LOAD_ERROR',
            message: error.message || '加载仪表板数据失败',
            details: error
          });
        } finally {
          setLoading('dashboard', false);
        }
      },

      loadRightsUsageTrend: async (days = 30) => {
        const { setLoading, setError } = get();
        
        try {
          setLoading('trends', true);
          
          const trends = await dashboardService.getRightsUsageTrend({ days });
          
          set({ rightsUsageTrend: trends }, false, 'loadRightsUsageTrend:success');
          
        } catch (error: any) {
          console.error('加载权益趋势数据失败:', error);
          setError({
            code: 'TRENDS_LOAD_ERROR',
            message: error.message || '加载权益趋势数据失败',
            details: error
          });
        } finally {
          setLoading('trends', false);
        }
      },

      loadPendingTasks: async () => {
        const { setLoading, setError } = get();
        
        try {
          setLoading('tasks', true);
          
          const tasks = await dashboardService.getPendingTasks();
          
          set({ pendingTasks: tasks }, false, 'loadPendingTasks:success');
          
        } catch (error: any) {
          console.error('加载待处理事项失败:', error);
          setError({
            code: 'TASKS_LOAD_ERROR',
            message: error.message || '加载待处理事项失败',
            details: error
          });
        } finally {
          setLoading('tasks', false);
        }
      },

      loadNotifications: async () => {
        const { setLoading, setError } = get();
        
        try {
          setLoading('notifications', true);
          
          const notifications = await dashboardService.getNotifications();
          
          set({ notifications }, false, 'loadNotifications:success');
          
        } catch (error: any) {
          console.error('加载通知公告失败:', error);
          setError({
            code: 'NOTIFICATIONS_LOAD_ERROR',
            message: error.message || '加载通知公告失败',
            details: error
          });
        } finally {
          setLoading('notifications', false);
        }
      },

      loadConfig: async () => {
        const { setLoading, setError } = get();
        
        try {
          setLoading('config', true);
          
          const config = await dashboardService.getDashboardConfig();
          
          set({
            config,
            refreshConfig: {
              ...get().refreshConfig,
              interval: config.refresh_interval * 1000 // 转换为毫秒
            }
          }, false, 'loadConfig:success');
          
        } catch (error: any) {
          console.error('加载仪表板配置失败:', error);
          setError({
            code: 'CONFIG_LOAD_ERROR',
            message: error.message || '加载仪表板配置失败',
            details: error
          });
        } finally {
          setLoading('config', false);
        }
      },

      loadAllData: async (period?: TimePeriod) => {
        const { setError } = get();
        
        try {
          setError(null);
          
          // 并行加载所有数据
          await Promise.all([
            get().loadDashboardData(period),
            get().loadRightsUsageTrend(30),
            get().loadPendingTasks(),
            get().loadNotifications(),
            get().loadConfig()
          ]);
          
        } catch (error: any) {
          console.error('批量加载仪表板数据失败:', error);
          setError({
            code: 'BATCH_LOAD_ERROR',
            message: error.message || '加载仪表板数据失败',
            details: error
          });
        }
      },

      // 配置管理方法
      updateConfig: async (configRequest: DashboardConfigRequest) => {
        const { setLoading, setError } = get();
        
        try {
          setLoading('config', true);
          
          await dashboardService.updateDashboardConfig(configRequest);
          
          // 重新加载配置
          await get().loadConfig();
          
        } catch (error: any) {
          console.error('更新仪表板配置失败:', error);
          setError({
            code: 'CONFIG_UPDATE_ERROR',
            message: error.message || '更新仪表板配置失败',
            details: error
          });
          throw error; // 重新抛出错误，让UI处理
        } finally {
          setLoading('config', false);
        }
      },

      saveConfig: async (configRequest: DashboardConfigRequest) => {
        const { setLoading, setError } = get();
        
        try {
          setLoading('config', true);
          
          await dashboardService.saveDashboardConfig(configRequest);
          
          // 重新加载配置
          await get().loadConfig();
          
        } catch (error: any) {
          console.error('保存仪表板配置失败:', error);
          setError({
            code: 'CONFIG_SAVE_ERROR',
            message: error.message || '保存仪表板配置失败',
            details: error
          });
          throw error; // 重新抛出错误，让UI处理
        } finally {
          setLoading('config', false);
        }
      },

      // 通知操作方法
      markAnnouncementAsRead: async (announcementId: number) => {
        const { setError, notifications } = get();
        
        try {
          await dashboardService.markAnnouncementAsRead(announcementId);
          
          // 更新本地状态
          if (notifications) {
            const updatedAnnouncements = notifications.announcements.map(
              announcement => 
                announcement.id === announcementId
                  ? { ...announcement, read_status: true }
                  : announcement
            );
            
            set({
              notifications: {
                ...notifications,
                announcements: updatedAnnouncements,
                unread_count: Math.max(0, notifications.unread_count - 1)
              }
            }, false, 'markAnnouncementAsRead:success');
          }
          
        } catch (error: any) {
          console.error('标记公告已读失败:', error);
          setError({
            code: 'MARK_READ_ERROR',
            message: error.message || '标记公告已读失败',
            details: error
          });
        }
      },

      // 刷新控制方法
      startAutoRefresh: () => {
        const { refreshConfig, refreshTimer, loadDashboardData } = get();
        
        // 清除现有定时器
        if (refreshTimer) {
          clearInterval(refreshTimer);
        }
        
        if (refreshConfig.autoRefresh) {
          const timer = setInterval(() => {
            loadDashboardData();
          }, refreshConfig.interval);
          
          set({ refreshTimer: timer }, false, 'startAutoRefresh');
        }
      },

      stopAutoRefresh: () => {
        const { refreshTimer } = get();
        
        if (refreshTimer) {
          clearInterval(refreshTimer);
          set({ refreshTimer: null }, false, 'stopAutoRefresh');
        }
      },

      refreshDashboard: async () => {
        const { currentPeriod } = get();
        await get().loadAllData(currentPeriod);
      },

      // 重置方法
      reset: () => {
        const { refreshTimer } = get();
        
        // 清除定时器
        if (refreshTimer) {
          clearInterval(refreshTimer);
        }
        
        set({
          ...initialState,
          refreshTimer: null
        }, false, 'reset');
      }
    }),
    {
      name: 'dashboard-store',
      // 在开发环境启用store devtools
      enabled: process.env.NODE_ENV === 'development'
    }
  )
);

// 导出选择器hooks，优化性能
export const useDashboardData = () => useDashboardStore(state => state.dashboardData);
export const useDashboardLoading = () => useDashboardStore(state => state.loading);
export const useDashboardError = () => useDashboardStore(state => state.error);
export const useRightsUsageTrend = () => useDashboardStore(state => state.rightsUsageTrend);
export const usePendingTasks = () => useDashboardStore(state => state.pendingTasks);
export const useNotifications = () => useDashboardStore(state => state.notifications);
export const useDashboardConfig = () => useDashboardStore(state => state.config);
export const useCurrentPeriod = () => useDashboardStore(state => state.currentPeriod);

// 导出默认store
export default useDashboardStore;
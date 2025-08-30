import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';

interface OrderStatusNotification {
  orderId: number;
  newStatus: number;
  oldStatus: number;
  reason: string;
  timestamp: Date;
  read: boolean;
}

interface OrderStatusStore {
  // 实时更新状态
  notifications: OrderStatusNotification[];
  lastUpdate: Date;
  isRealTimeEnabled: boolean;
  connectionStatus: 'connected' | 'disconnected' | 'connecting';
  
  // 筛选和搜索状态
  filters: {
    status: number[];
    dateRange: [string, string] | null;
    searchKeyword: string;
    merchantId?: number;
  };
  
  // 统计信息
  stats: {
    total: number;
    byStatus: Record<number, number>;
    todayOrders: number;
    pendingPayment: number;
  };
  
  // 操作方法
  addNotification: (notification: Omit<OrderStatusNotification, 'timestamp' | 'read'>) => void;
  markNotificationRead: (index: number) => void;
  clearNotifications: () => void;
  clearAllNotifications: () => void;
  
  // 连接管理
  setRealTimeEnabled: (enabled: boolean) => void;
  setConnectionStatus: (status: 'connected' | 'disconnected' | 'connecting') => void;
  
  // 筛选和搜索
  setFilters: (filters: Partial<OrderStatusStore['filters']>) => void;
  resetFilters: () => void;
  
  // 统计
  updateStats: (stats: Partial<OrderStatusStore['stats']>) => void;
  
  // 获取未读通知数量
  getUnreadCount: () => number;
}

export const useOrderStatusStore = create<OrderStatusStore>()(
  subscribeWithSelector((set, get) => ({
    // 初始状态
    notifications: [],
    lastUpdate: new Date(),
    isRealTimeEnabled: false,
    connectionStatus: 'disconnected',
    
    filters: {
      status: [],
      dateRange: null,
      searchKeyword: '',
    },
    
    stats: {
      total: 0,
      byStatus: {},
      todayOrders: 0,
      pendingPayment: 0,
    },
    
    // 通知管理
    addNotification: (notification) => {
      set(state => ({
        notifications: [
          {
            ...notification,
            timestamp: new Date(),
            read: false,
          },
          ...state.notifications.slice(0, 19), // 保持最多20条通知
        ],
        lastUpdate: new Date(),
      }));
    },
    
    markNotificationRead: (index) => {
      set(state => ({
        notifications: state.notifications.map((notification, i) =>
          i === index ? { ...notification, read: true } : notification
        ),
      }));
    },
    
    clearNotifications: () => {
      set(state => ({
        notifications: state.notifications.filter(n => !n.read),
      }));
    },
    
    clearAllNotifications: () => {
      set({ notifications: [] });
    },
    
    // 连接管理
    setRealTimeEnabled: (enabled) => {
      set({ 
        isRealTimeEnabled: enabled,
        connectionStatus: enabled ? 'connecting' : 'disconnected',
      });
    },
    
    setConnectionStatus: (status) => {
      set({ connectionStatus: status });
    },
    
    // 筛选和搜索
    setFilters: (newFilters) => {
      set(state => ({
        filters: { ...state.filters, ...newFilters },
      }));
    },
    
    resetFilters: () => {
      set({
        filters: {
          status: [],
          dateRange: null,
          searchKeyword: '',
        },
      });
    },
    
    // 统计
    updateStats: (newStats) => {
      set(state => ({
        stats: { ...state.stats, ...newStats },
      }));
    },
    
    // 获取未读通知数量
    getUnreadCount: () => {
      const { notifications } = get();
      return notifications.filter(n => !n.read).length;
    },
  }))
);
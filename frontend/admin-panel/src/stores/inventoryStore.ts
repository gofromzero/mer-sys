// 库存管理状态管理
import { create } from 'zustand';
import { 
  InventoryResponse, 
  InventoryAlert, 
  InventoryMonitoringData,
  InventoryRecord,
  InventoryAdjustRequest 
} from '../types/product';
import { InventoryService } from '../services/inventoryService';

interface InventoryState {
  // 状态数据
  inventory: Record<number, InventoryResponse>;
  activeAlerts: InventoryAlert[];
  monitoringData: InventoryMonitoringData | null;
  recentRecords: InventoryRecord[];
  
  // 加载状态
  loading: {
    inventory: boolean;
    alerts: boolean;
    monitoring: boolean;
    records: boolean;
    adjusting: boolean;
  };
  
  // 错误状态
  error: string | null;
  
  // Actions
  actions: {
    // 库存查询
    fetchInventory: (productId: number) => Promise<void>;
    fetchInventoryBatch: (productIds: number[]) => Promise<void>;
    
    // 库存调整
    adjustInventory: (request: InventoryAdjustRequest) => Promise<void>;
    
    // 预警管理
    fetchActiveAlerts: () => Promise<void>;
    toggleAlert: (alertId: number, isActive: boolean) => Promise<void>;
    checkAllLowStockAlerts: () => Promise<void>;
    
    // 监控数据
    fetchMonitoringData: () => Promise<void>;
    
    // 库存记录
    fetchRecentRecords: (limit?: number) => Promise<void>;
    
    // 错误处理
    clearError: () => void;
    setError: (error: string) => void;
  };
}

export const useInventoryStore = create<InventoryState>((set, get) => ({
  // 初始状态
  inventory: {},
  activeAlerts: [],
  monitoringData: null,
  recentRecords: [],
  
  loading: {
    inventory: false,
    alerts: false,
    monitoring: false,
    records: false,
    adjusting: false,
  },
  
  error: null,
  
  actions: {
    // 获取单个商品库存
    fetchInventory: async (productId: number) => {
      set((state) => ({
        loading: { ...state.loading, inventory: true },
        error: null
      }));
      
      try {
        const inventoryData = await InventoryService.getInventory(productId);
        set((state) => ({
          inventory: {
            ...state.inventory,
            [productId]: inventoryData
          },
          loading: { ...state.loading, inventory: false }
        }));
      } catch (error) {
        set((state) => ({
          error: error instanceof Error ? error.message : '获取库存信息失败',
          loading: { ...state.loading, inventory: false }
        }));
      }
    },

    // 批量获取库存
    fetchInventoryBatch: async (productIds: number[]) => {
      set((state) => ({
        loading: { ...state.loading, inventory: true },
        error: null
      }));
      
      try {
        const inventoryList = await InventoryService.getInventoryBatch(productIds);
        const inventoryMap = inventoryList.reduce((acc, item) => {
          acc[item.product_id] = item;
          return acc;
        }, {} as Record<number, InventoryResponse>);
        
        set((state) => ({
          inventory: { ...state.inventory, ...inventoryMap },
          loading: { ...state.loading, inventory: false }
        }));
      } catch (error) {
        set((state) => ({
          error: error instanceof Error ? error.message : '批量获取库存信息失败',
          loading: { ...state.loading, inventory: false }
        }));
      }
    },

    // 调整库存
    adjustInventory: async (request: InventoryAdjustRequest) => {
      set((state) => ({
        loading: { ...state.loading, adjusting: true },
        error: null
      }));
      
      try {
        const updatedInventory = await InventoryService.adjustInventory(request);
        set((state) => ({
          inventory: {
            ...state.inventory,
            [request.product_id]: updatedInventory
          },
          loading: { ...state.loading, adjusting: false }
        }));
      } catch (error) {
        set((state) => ({
          error: error instanceof Error ? error.message : '库存调整失败',
          loading: { ...state.loading, adjusting: false }
        }));
        throw error; // 重新抛出错误，让调用方处理
      }
    },

    // 获取活跃预警
    fetchActiveAlerts: async () => {
      set((state) => ({
        loading: { ...state.loading, alerts: true },
        error: null
      }));
      
      try {
        const alerts = await InventoryService.getActiveAlerts();
        set((state) => ({
          activeAlerts: alerts,
          loading: { ...state.loading, alerts: false }
        }));
      } catch (error) {
        set((state) => ({
          error: error instanceof Error ? error.message : '获取预警信息失败',
          loading: { ...state.loading, alerts: false }
        }));
      }
    },

    // 切换预警状态
    toggleAlert: async (alertId: number, isActive: boolean) => {
      try {
        await InventoryService.toggleAlert(alertId, isActive);
        set((state) => ({
          activeAlerts: state.activeAlerts.map(alert =>
            alert.id === alertId ? { ...alert, is_active: isActive } : alert
          )
        }));
      } catch (error) {
        set((state) => ({
          error: error instanceof Error ? error.message : '切换预警状态失败'
        }));
        throw error;
      }
    },

    // 检查所有低库存预警
    checkAllLowStockAlerts: async () => {
      set((state) => ({
        loading: { ...state.loading, alerts: true },
        error: null
      }));
      
      try {
        await InventoryService.checkAllLowStockAlerts();
        // 重新获取预警数据
        await get().actions.fetchActiveAlerts();
      } catch (error) {
        set((state) => ({
          error: error instanceof Error ? error.message : '预警检查失败',
          loading: { ...state.loading, alerts: false }
        }));
      }
    },

    // 获取监控数据
    fetchMonitoringData: async () => {
      set((state) => ({
        loading: { ...state.loading, monitoring: true },
        error: null
      }));
      
      try {
        const monitoringData = await InventoryService.getMonitoringData();
        set((state) => ({
          monitoringData,
          loading: { ...state.loading, monitoring: false }
        }));
      } catch (error) {
        set((state) => ({
          error: error instanceof Error ? error.message : '获取监控数据失败',
          loading: { ...state.loading, monitoring: false }
        }));
      }
    },

    // 获取最近记录
    fetchRecentRecords: async (limit = 50) => {
      set((state) => ({
        loading: { ...state.loading, records: true },
        error: null
      }));
      
      try {
        const response = await InventoryService.getInventoryRecords({
          page: 1,
          page_size: limit
        });
        set((state) => ({
          recentRecords: response.records,
          loading: { ...state.loading, records: false }
        }));
      } catch (error) {
        set((state) => ({
          error: error instanceof Error ? error.message : '获取库存记录失败',
          loading: { ...state.loading, records: false }
        }));
      }
    },

    // 错误处理
    clearError: () => {
      set({ error: null });
    },

    setError: (error: string) => {
      set({ error });
    },
  },
}));

// 导出便捷hooks
export const useInventoryActions = () => useInventoryStore(state => state.actions);
export const useInventoryData = () => useInventoryStore(state => ({
  inventory: state.inventory,
  activeAlerts: state.activeAlerts,
  monitoringData: state.monitoringData,
  recentRecords: state.recentRecords,
}));
export const useInventoryLoading = () => useInventoryStore(state => state.loading);
export const useInventoryError = () => useInventoryStore(state => state.error);
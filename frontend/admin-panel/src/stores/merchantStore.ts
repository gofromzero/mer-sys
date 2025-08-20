import { create } from 'zustand';
import type { Merchant, MerchantListQuery, MerchantListResponse } from '../types/merchant';
import { MerchantService } from '../services/merchantService';

interface MerchantState {
  // 商户列表相关状态
  merchants: Merchant[];
  total: number;
  loading: boolean;
  error: string | null;
  
  // 当前查询参数
  query: MerchantListQuery;
  
  // 选中的商户
  selectedMerchant: Merchant | null;
}

interface MerchantActions {
  // 获取商户列表
  fetchMerchants: (query?: MerchantListQuery) => Promise<void>;
  
  // 获取商户详情
  fetchMerchantById: (id: number) => Promise<void>;
  
  // 更新查询参数
  updateQuery: (query: Partial<MerchantListQuery>) => void;
  
  // 清除错误
  clearError: () => void;
  
  // 重置状态
  reset: () => void;
  
  // 审批商户
  approveMerchant: (id: number, comment?: string) => Promise<void>;
  
  // 拒绝商户
  rejectMerchant: (id: number, comment?: string) => Promise<void>;
  
  // 更新商户状态
  updateMerchantStatus: (id: number, status: string, comment?: string) => Promise<void>;
}

type MerchantStore = MerchantState & MerchantActions;

const initialState: MerchantState = {
  merchants: [],
  total: 0,
  loading: false,
  error: null,
  query: {
    page: 1,
    page_size: 20
  },
  selectedMerchant: null
};

export const useMerchantStore = create<MerchantStore>((set, get) => ({
  ...initialState,

  fetchMerchants: async (query?: MerchantListQuery) => {
    set({ loading: true, error: null });
    
    try {
      const finalQuery = { ...get().query, ...query };
      const response: MerchantListResponse = await MerchantService.getMerchantList(finalQuery);
      
      set({
        merchants: response.items,
        total: response.total,
        query: finalQuery,
        loading: false
      });
    } catch (error) {
      console.error('获取商户列表失败:', error);
      set({
        error: error instanceof Error ? error.message : '获取商户列表失败',
        loading: false
      });
    }
  },

  fetchMerchantById: async (id: number) => {
    set({ loading: true, error: null });
    
    try {
      const merchant = await MerchantService.getMerchantById(id);
      set({
        selectedMerchant: merchant,
        loading: false
      });
    } catch (error) {
      console.error('获取商户详情失败:', error);
      set({
        error: error instanceof Error ? error.message : '获取商户详情失败',
        loading: false
      });
    }
  },

  updateQuery: (query: Partial<MerchantListQuery>) => {
    set(state => ({
      query: { ...state.query, ...query }
    }));
  },

  clearError: () => {
    set({ error: null });
  },

  reset: () => {
    set(initialState);
  },

  approveMerchant: async (id: number, comment?: string) => {
    set({ loading: true, error: null });
    
    try {
      await MerchantService.approveMerchant(id, comment);
      
      // 更新本地状态中的商户状态
      set(state => ({
        merchants: state.merchants.map(merchant =>
          merchant.id === id
            ? { ...merchant, status: 'active' as any }
            : merchant
        ),
        loading: false
      }));
      
      // 重新获取列表以确保数据一致性
      await get().fetchMerchants();
    } catch (error) {
      console.error('审批商户失败:', error);
      set({
        error: error instanceof Error ? error.message : '审批商户失败',
        loading: false
      });
      throw error;
    }
  },

  rejectMerchant: async (id: number, comment?: string) => {
    set({ loading: true, error: null });
    
    try {
      await MerchantService.rejectMerchant(id, comment);
      
      // 更新本地状态中的商户状态
      set(state => ({
        merchants: state.merchants.map(merchant =>
          merchant.id === id
            ? { ...merchant, status: 'deactivated' as any }
            : merchant
        ),
        loading: false
      }));
      
      // 重新获取列表以确保数据一致性
      await get().fetchMerchants();
    } catch (error) {
      console.error('拒绝商户失败:', error);
      set({
        error: error instanceof Error ? error.message : '拒绝商户失败',
        loading: false
      });
      throw error;
    }
  },

  updateMerchantStatus: async (id: number, status: string, comment?: string) => {
    set({ loading: true, error: null });
    
    try {
      await MerchantService.updateMerchantStatus(id, { status: status as any, comment });
      
      // 更新本地状态中的商户状态
      set(state => ({
        merchants: state.merchants.map(merchant =>
          merchant.id === id
            ? { ...merchant, status: status as any }
            : merchant
        ),
        loading: false
      }));
      
      // 重新获取列表以确保数据一致性
      await get().fetchMerchants();
    } catch (error) {
      console.error('更新商户状态失败:', error);
      set({
        error: error instanceof Error ? error.message : '更新商户状态失败',
        loading: false
      });
      throw error;
    }
  }
}));
import { create } from 'zustand';
import type { MerchantUser, MerchantUserQueryParams } from '../types/merchantUser';
import { MerchantUserService } from '../services/merchantUserService';

interface MerchantUserState {
  // 数据状态
  merchantUsers: MerchantUser[];
  currentMerchantUser: MerchantUser | null;
  loading: boolean;
  error: string | null;
  
  // 分页状态
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  
  // 查询参数
  queryParams: MerchantUserQueryParams;
  
  // Actions
  setMerchantUsers: (users: MerchantUser[]) => void;
  setCurrentMerchantUser: (user: MerchantUser | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setPagination: (pagination: MerchantUserState['pagination']) => void;
  setQueryParams: (params: MerchantUserQueryParams) => void;
  
  // API Actions
  fetchMerchantUsers: (params?: MerchantUserQueryParams) => Promise<void>;
  fetchMerchantUser: (id: number) => Promise<void>;
  createMerchantUser: (data: any) => Promise<MerchantUser>;
  updateMerchantUser: (id: number, data: any) => Promise<MerchantUser>;
  updateMerchantUserStatus: (id: number, status: string, comment?: string) => Promise<void>;
  resetMerchantUserPassword: (id: number, data: any) => Promise<void>;
  deleteMerchantUser: (id: number) => Promise<void>;
  
  // 清理状态
  clearError: () => void;
  reset: () => void;
}

const initialState = {
  merchantUsers: [],
  currentMerchantUser: null,
  loading: false,
  error: null,
  pagination: {
    page: 1,
    page_size: 20,
    total: 0,
    total_pages: 0
  },
  queryParams: {
    page: 1,
    page_size: 20
  }
};

export const useMerchantUserStore = create<MerchantUserState>((set, get) => ({
  ...initialState,
  
  // Basic setters
  setMerchantUsers: (merchantUsers) => set({ merchantUsers }),
  setCurrentMerchantUser: (currentMerchantUser) => set({ currentMerchantUser }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  setPagination: (pagination) => set({ pagination }),
  setQueryParams: (queryParams) => set({ queryParams }),
  
  // 获取商户用户列表
  fetchMerchantUsers: async (params?: MerchantUserQueryParams) => {
    set({ loading: true, error: null });
    try {
      const finalParams = { ...get().queryParams, ...params };
      const response = await MerchantUserService.getMerchantUsers(finalParams);
      
      set({
        merchantUsers: response.list,
        pagination: response.pagination,
        queryParams: finalParams,
        loading: false
      });
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '获取商户用户列表失败',
        loading: false 
      });
    }
  },
  
  // 获取单个商户用户
  fetchMerchantUser: async (id: number) => {
    set({ loading: true, error: null });
    try {
      const user = await MerchantUserService.getMerchantUser(id);
      set({ currentMerchantUser: user, loading: false });
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '获取商户用户详情失败',
        loading: false 
      });
    }
  },
  
  // 创建商户用户
  createMerchantUser: async (data) => {
    set({ loading: true, error: null });
    try {
      const newUser = await MerchantUserService.createMerchantUser(data);
      const currentUsers = get().merchantUsers;
      set({ 
        merchantUsers: [newUser, ...currentUsers],
        loading: false 
      });
      return newUser;
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '创建商户用户失败',
        loading: false 
      });
      throw error;
    }
  },
  
  // 更新商户用户
  updateMerchantUser: async (id: number, data) => {
    set({ loading: true, error: null });
    try {
      const updatedUser = await MerchantUserService.updateMerchantUser(id, data);
      const currentUsers = get().merchantUsers;
      const updatedUsers = currentUsers.map(user => 
        user.id === id ? updatedUser : user
      );
      
      set({ 
        merchantUsers: updatedUsers,
        currentMerchantUser: get().currentMerchantUser?.id === id ? updatedUser : get().currentMerchantUser,
        loading: false 
      });
      return updatedUser;
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '更新商户用户失败',
        loading: false 
      });
      throw error;
    }
  },
  
  // 更新商户用户状态
  updateMerchantUserStatus: async (id: number, status: string, comment?: string) => {
    set({ loading: true, error: null });
    try {
      const updatedUser = await MerchantUserService.updateMerchantUserStatus(id, { 
        status: status as any, 
        comment 
      });
      
      const currentUsers = get().merchantUsers;
      const updatedUsers = currentUsers.map(user => 
        user.id === id ? updatedUser : user
      );
      
      set({ 
        merchantUsers: updatedUsers,
        currentMerchantUser: get().currentMerchantUser?.id === id ? updatedUser : get().currentMerchantUser,
        loading: false 
      });
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '更新商户用户状态失败',
        loading: false 
      });
      throw error;
    }
  },
  
  // 重置密码
  resetMerchantUserPassword: async (id: number, data) => {
    set({ loading: true, error: null });
    try {
      await MerchantUserService.resetMerchantUserPassword(id, data);
      set({ loading: false });
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '重置密码失败',
        loading: false 
      });
      throw error;
    }
  },
  
  // 删除商户用户
  deleteMerchantUser: async (id: number) => {
    set({ loading: true, error: null });
    try {
      await MerchantUserService.deleteMerchantUser(id);
      const currentUsers = get().merchantUsers;
      const updatedUsers = currentUsers.filter(user => user.id !== id);
      
      set({ 
        merchantUsers: updatedUsers,
        currentMerchantUser: get().currentMerchantUser?.id === id ? null : get().currentMerchantUser,
        loading: false 
      });
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '删除商户用户失败',
        loading: false 
      });
      throw error;
    }
  },
  
  // 清理错误
  clearError: () => set({ error: null }),
  
  // 重置状态
  reset: () => set(initialState)
}));
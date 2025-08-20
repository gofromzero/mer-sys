// 资金管理状态管理 - Zustand Store

import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';
import fundService from '../services/fundService';
import type {
  Fund,
  FundTransaction,
  RightsBalance,
  FundSummary,
  DepositRequest,
  BatchDepositRequest,
  AllocateRequest,
  FundTransactionQuery,
  FreezeRequest,
  PaginationResponse
} from '../types/fund';

// Store状态接口
interface FundState {
  // 数据状态
  fundList: Fund[];
  transactionList: FundTransaction[];
  balanceMap: Record<number, RightsBalance>; // 商户ID -> 余额信息
  summaryMap: Record<number, FundSummary>;   // 商户ID -> 统计信息
  merchantList: Array<{ id: number; name: string; code: string }>;
  
  // 加载状态
  loading: {
    fundList: boolean;
    transactionList: boolean;
    balance: Record<number, boolean>; // 商户ID -> 加载状态
    summary: Record<number, boolean>; // 商户ID -> 加载状态
    deposit: boolean;
    allocate: boolean;
    freeze: boolean;
    merchantList: boolean;
  };
  
  // 分页信息
  pagination: {
    transactions: {
      page: number;
      pageSize: number;
      total: number;
    };
  };
  
  // 错误状态
  error: string | null;
  
  // Actions
  actions: {
    // 充值相关
    deposit: (request: DepositRequest) => Promise<Fund>;
    batchDeposit: (request: BatchDepositRequest) => Promise<Fund[]>;
    
    // 权益分配
    allocate: (request: AllocateRequest) => Promise<Fund>;
    
    // 余额查询
    getMerchantBalance: (merchantId: number, forceRefresh?: boolean) => Promise<RightsBalance>;
    refreshMerchantBalance: (merchantId: number) => Promise<void>;
    
    // 交易记录查询
    listTransactions: (query: FundTransactionQuery) => Promise<PaginationResponse<FundTransaction>>;
    refreshTransactionList: (query?: Partial<FundTransactionQuery>) => Promise<void>;
    
    // 统计查询
    getFundSummary: (merchantId?: number, forceRefresh?: boolean) => Promise<FundSummary>;
    refreshFundSummary: (merchantId?: number) => Promise<void>;
    
    // 冻结/解冻
    freezeMerchantBalance: (merchantId: number, request: FreezeRequest) => Promise<void>;
    
    // 商户列表
    loadMerchantList: (forceRefresh?: boolean) => Promise<void>;
    
    // 工具方法
    clearError: () => void;
    resetStore: () => void;
  };
}

// 初始状态
const initialState = {
  fundList: [],
  transactionList: [],
  balanceMap: {},
  summaryMap: {},
  merchantList: [],
  loading: {
    fundList: false,
    transactionList: false,
    balance: {},
    summary: {},
    deposit: false,
    allocate: false,
    freeze: false,
    merchantList: false,
  },
  pagination: {
    transactions: {
      page: 1,
      pageSize: 20,
      total: 0,
    },
  },
  error: null,
};

export const useFundStore = create<FundState>()(
  subscribeWithSelector((set, get) => ({
    ...initialState,
    
    actions: {
      // 单笔充值
      deposit: async (request: DepositRequest) => {
        set((state) => ({
          loading: { ...state.loading, deposit: true },
          error: null,
        }));
        
        try {
          const fund = await fundService.deposit(request);
          
          // 更新fund列表
          set((state) => ({
            fundList: [fund, ...state.fundList],
            loading: { ...state.loading, deposit: false },
          }));
          
          // 刷新相关商户的余额
          get().actions.refreshMerchantBalance(request.merchant_id);
          
          return fund;
        } catch (error) {
          set((state) => ({
            loading: { ...state.loading, deposit: false },
            error: error instanceof Error ? error.message : '充值失败',
          }));
          throw error;
        }
      },
      
      // 批量充值
      batchDeposit: async (request: BatchDepositRequest) => {
        set((state) => ({
          loading: { ...state.loading, deposit: true },
          error: null,
        }));
        
        try {
          const funds = await fundService.batchDeposit(request);
          
          // 更新fund列表
          set((state) => ({
            fundList: [...funds, ...state.fundList],
            loading: { ...state.loading, deposit: false },
          }));
          
          // 刷新相关商户的余额
          const merchantIds = [...new Set(request.deposits.map(d => d.merchant_id))];
          merchantIds.forEach(merchantId => {
            get().actions.refreshMerchantBalance(merchantId);
          });
          
          return funds;
        } catch (error) {
          set((state) => ({
            loading: { ...state.loading, deposit: false },
            error: error instanceof Error ? error.message : '批量充值失败',
          }));
          throw error;
        }
      },
      
      // 权益分配
      allocate: async (request: AllocateRequest) => {
        set((state) => ({
          loading: { ...state.loading, allocate: true },
          error: null,
        }));
        
        try {
          const fund = await fundService.allocate(request);
          
          // 更新fund列表
          set((state) => ({
            fundList: [fund, ...state.fundList],
            loading: { ...state.loading, allocate: false },
          }));
          
          // 刷新相关商户的余额
          get().actions.refreshMerchantBalance(request.merchant_id);
          
          return fund;
        } catch (error) {
          set((state) => ({
            loading: { ...state.loading, allocate: false },
            error: error instanceof Error ? error.message : '权益分配失败',
          }));
          throw error;
        }
      },
      
      // 获取商户余额
      getMerchantBalance: async (merchantId: number, forceRefresh = false) => {
        const state = get();
        
        // 如果已有数据且不强制刷新，直接返回
        if (!forceRefresh && state.balanceMap[merchantId] && !state.loading.balance[merchantId]) {
          return state.balanceMap[merchantId];
        }
        
        set((state) => ({
          loading: {
            ...state.loading,
            balance: { ...state.loading.balance, [merchantId]: true },
          },
          error: null,
        }));
        
        try {
          const balance = await fundService.getMerchantBalance(merchantId);
          
          set((state) => ({
            balanceMap: { ...state.balanceMap, [merchantId]: balance },
            loading: {
              ...state.loading,
              balance: { ...state.loading.balance, [merchantId]: false },
            },
          }));
          
          return balance;
        } catch (error) {
          set((state) => ({
            loading: {
              ...state.loading,
              balance: { ...state.loading.balance, [merchantId]: false },
            },
            error: error instanceof Error ? error.message : '查询余额失败',
          }));
          throw error;
        }
      },
      
      // 刷新商户余额
      refreshMerchantBalance: async (merchantId: number) => {
        await get().actions.getMerchantBalance(merchantId, true);
      },
      
      // 查询交易记录
      listTransactions: async (query: FundTransactionQuery) => {
        set((state) => ({
          loading: { ...state.loading, transactionList: true },
          error: null,
        }));
        
        try {
          const result = await fundService.listTransactions(query);
          
          set((state) => ({
            transactionList: result.list,
            pagination: {
              ...state.pagination,
              transactions: {
                page: result.page,
                pageSize: result.page_size,
                total: result.total,
              },
            },
            loading: { ...state.loading, transactionList: false },
          }));
          
          return result;
        } catch (error) {
          set((state) => ({
            loading: { ...state.loading, transactionList: false },
            error: error instanceof Error ? error.message : '查询交易记录失败',
          }));
          throw error;
        }
      },
      
      // 刷新交易记录
      refreshTransactionList: async (query?: Partial<FundTransactionQuery>) => {
        const currentQuery: FundTransactionQuery = {
          page: get().pagination.transactions.page,
          page_size: get().pagination.transactions.pageSize,
          ...query,
        };
        await get().actions.listTransactions(currentQuery);
      },
      
      // 获取统计信息
      getFundSummary: async (merchantId?: number, forceRefresh = false) => {
        const summaryKey = merchantId || 0;
        const state = get();
        
        // 如果已有数据且不强制刷新，直接返回
        if (!forceRefresh && state.summaryMap[summaryKey] && !state.loading.summary[summaryKey]) {
          return state.summaryMap[summaryKey];
        }
        
        set((state) => ({
          loading: {
            ...state.loading,
            summary: { ...state.loading.summary, [summaryKey]: true },
          },
          error: null,
        }));
        
        try {
          const summary = await fundService.getFundSummary(merchantId);
          
          set((state) => ({
            summaryMap: { ...state.summaryMap, [summaryKey]: summary },
            loading: {
              ...state.loading,
              summary: { ...state.loading.summary, [summaryKey]: false },
            },
          }));
          
          return summary;
        } catch (error) {
          set((state) => ({
            loading: {
              ...state.loading,
              summary: { ...state.loading.summary, [summaryKey]: false },
            },
            error: error instanceof Error ? error.message : '获取统计失败',
          }));
          throw error;
        }
      },
      
      // 刷新统计信息
      refreshFundSummary: async (merchantId?: number) => {
        await get().actions.getFundSummary(merchantId, true);
      },
      
      // 冻结/解冻权益
      freezeMerchantBalance: async (merchantId: number, request: FreezeRequest) => {
        set((state) => ({
          loading: { ...state.loading, freeze: true },
          error: null,
        }));
        
        try {
          await fundService.freezeMerchantBalance(merchantId, request);
          
          set((state) => ({
            loading: { ...state.loading, freeze: false },
          }));
          
          // 刷新相关商户的余额
          get().actions.refreshMerchantBalance(merchantId);
        } catch (error) {
          set((state) => ({
            loading: { ...state.loading, freeze: false },
            error: error instanceof Error ? error.message : '操作失败',
          }));
          throw error;
        }
      },
      
      // 加载商户列表
      loadMerchantList: async (forceRefresh = false) => {
        const state = get();
        
        // 如果已有数据且不强制刷新，直接返回
        if (!forceRefresh && state.merchantList.length > 0 && !state.loading.merchantList) {
          return;
        }
        
        set((state) => ({
          loading: { ...state.loading, merchantList: true },
          error: null,
        }));
        
        try {
          const merchantList = await fundService.getMerchantList();
          
          set((state) => ({
            merchantList,
            loading: { ...state.loading, merchantList: false },
          }));
        } catch (error) {
          set((state) => ({
            loading: { ...state.loading, merchantList: false },
            error: error instanceof Error ? error.message : '获取商户列表失败',
          }));
        }
      },
      
      // 清除错误
      clearError: () => {
        set({ error: null });
      },
      
      // 重置Store
      resetStore: () => {
        set(initialState);
      },
    },
  }))
);

// 导出hooks
export const useFundActions = () => useFundStore((state) => state.actions);
export const useFundData = () => useFundStore((state) => ({
  fundList: state.fundList,
  transactionList: state.transactionList,
  balanceMap: state.balanceMap,
  summaryMap: state.summaryMap,
  merchantList: state.merchantList,
}));
export const useFundLoading = () => useFundStore((state) => state.loading);
export const useFundPagination = () => useFundStore((state) => state.pagination);
export const useFundError = () => useFundStore((state) => state.error);

export default useFundStore;
// 资金管理Store测试

import { renderHook, act } from '@testing-library/react';
import { useFundStore } from '../fundStore';
import fundService from '../../services/fundService';
import type { Fund, RightsBalance, FundSummary } from '../../types/fund';

// Mock fundService
jest.mock('../../services/fundService');
const mockFundService = fundService as jest.Mocked<typeof fundService>;

describe('FundStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useFundStore.getState().actions.resetStore();
    jest.clearAllMocks();
  });

  describe('Initial State', () => {
    it('should have correct initial state', () => {
      const { result } = renderHook(() => useFundStore());
      
      expect(result.current.fundList).toEqual([]);
      expect(result.current.transactionList).toEqual([]);
      expect(result.current.balanceMap).toEqual({});
      expect(result.current.summaryMap).toEqual({});
      expect(result.current.merchantList).toEqual([]);
      expect(result.current.error).toBe(null);
    });
  });

  describe('Actions', () => {
    describe('deposit', () => {
      it('should handle successful deposit', async () => {
        const mockFund: Fund = {
          id: 1,
          tenant_id: 1,
          merchant_id: 1,
          fund_type: 1,
          amount: 1000,
          currency: 'CNY',
          status: 2,
          created_at: '2023-01-01T00:00:00Z',
          updated_at: '2023-01-01T00:00:00Z'
        };

        mockFundService.deposit.mockResolvedValue(mockFund);

        const { result } = renderHook(() => useFundStore());

        await act(async () => {
          const fund = await result.current.actions.deposit({
            merchant_id: 1,
            amount: 1000,
            currency: 'CNY',
            description: 'Test deposit'
          });

          expect(fund).toEqual(mockFund);
        });

        expect(result.current.fundList).toContain(mockFund);
        expect(result.current.loading.deposit).toBe(false);
        expect(result.current.error).toBe(null);
      });

      it('should handle deposit failure', async () => {
        const errorMessage = 'Deposit failed';
        mockFundService.deposit.mockRejectedValue(new Error(errorMessage));

        const { result } = renderHook(() => useFundStore());

        await act(async () => {
          try {
            await result.current.actions.deposit({
              merchant_id: 1,
              amount: 1000,
              currency: 'CNY'
            });
          } catch (error) {
            expect((error as Error).message).toBe(errorMessage);
          }
        });

        expect(result.current.loading.deposit).toBe(false);
        expect(result.current.error).toBe(errorMessage);
        expect(result.current.fundList).toEqual([]);
      });

      it('should set loading state during deposit', async () => {
        let resolveDeposit: (value: Fund) => void;
        const depositPromise = new Promise<Fund>((resolve) => {
          resolveDeposit = resolve;
        });

        mockFundService.deposit.mockReturnValue(depositPromise);

        const { result } = renderHook(() => useFundStore());

        // Start deposit
        act(() => {
          result.current.actions.deposit({
            merchant_id: 1,
            amount: 1000,
            currency: 'CNY'
          });
        });

        // Check loading state
        expect(result.current.loading.deposit).toBe(true);

        // Resolve deposit
        await act(async () => {
          resolveDeposit!({
            id: 1,
            tenant_id: 1,
            merchant_id: 1,
            fund_type: 1,
            amount: 1000,
            currency: 'CNY',
            status: 2,
            created_at: '2023-01-01T00:00:00Z',
            updated_at: '2023-01-01T00:00:00Z'
          });
        });

        expect(result.current.loading.deposit).toBe(false);
      });
    });

    describe('getMerchantBalance', () => {
      it('should fetch and cache merchant balance', async () => {
        const mockBalance: RightsBalance = {
          total_balance: 10000,
          used_balance: 2000,
          frozen_balance: 500,
          available_balance: 7500,
          last_updated: '2023-01-01T00:00:00Z'
        };

        mockFundService.getMerchantBalance.mockResolvedValue(mockBalance);

        const { result } = renderHook(() => useFundStore());

        let balance: RightsBalance;
        await act(async () => {
          balance = await result.current.actions.getMerchantBalance(1);
        });

        expect(balance!).toEqual(mockBalance);
        expect(result.current.balanceMap[1]).toEqual(mockBalance);
        expect(result.current.loading.balance[1]).toBe(false);
        expect(mockFundService.getMerchantBalance).toHaveBeenCalledWith(1);
      });

      it('should return cached balance when not forcing refresh', async () => {
        const mockBalance: RightsBalance = {
          total_balance: 10000,
          used_balance: 2000,
          frozen_balance: 500,
          available_balance: 7500,
          last_updated: '2023-01-01T00:00:00Z'
        };

        // Set initial balance in store
        const { result } = renderHook(() => useFundStore());
        act(() => {
          useFundStore.setState({
            balanceMap: { 1: mockBalance }
          });
        });

        let balance: RightsBalance;
        await act(async () => {
          balance = await result.current.actions.getMerchantBalance(1, false);
        });

        expect(balance!).toEqual(mockBalance);
        expect(mockFundService.getMerchantBalance).not.toHaveBeenCalled();
      });

      it('should force refresh when requested', async () => {
        const cachedBalance: RightsBalance = {
          total_balance: 10000,
          used_balance: 2000,
          frozen_balance: 500,
          available_balance: 7500,
          last_updated: '2023-01-01T00:00:00Z'
        };

        const freshBalance: RightsBalance = {
          total_balance: 15000,
          used_balance: 3000,
          frozen_balance: 1000,
          available_balance: 11000,
          last_updated: '2023-01-02T00:00:00Z'
        };

        mockFundService.getMerchantBalance.mockResolvedValue(freshBalance);

        // Set initial balance in store
        const { result } = renderHook(() => useFundStore());
        act(() => {
          useFundStore.setState({
            balanceMap: { 1: cachedBalance }
          });
        });

        let balance: RightsBalance;
        await act(async () => {
          balance = await result.current.actions.getMerchantBalance(1, true);
        });

        expect(balance!).toEqual(freshBalance);
        expect(result.current.balanceMap[1]).toEqual(freshBalance);
        expect(mockFundService.getMerchantBalance).toHaveBeenCalledWith(1);
      });
    });

    describe('getFundSummary', () => {
      it('should fetch overall summary', async () => {
        const mockSummary: FundSummary = {
          total_deposits: 100000,
          total_allocations: 80000,
          total_consumption: 30000,
          total_refunds: 5000,
          available_balance: 145000
        };

        mockFundService.getFundSummary.mockResolvedValue(mockSummary);

        const { result } = renderHook(() => useFundStore());

        let summary: FundSummary;
        await act(async () => {
          summary = await result.current.actions.getFundSummary();
        });

        expect(summary!).toEqual(mockSummary);
        expect(result.current.summaryMap[0]).toEqual(mockSummary);
        expect(mockFundService.getFundSummary).toHaveBeenCalledWith(undefined);
      });

      it('should fetch merchant-specific summary', async () => {
        const mockSummary: FundSummary = {
          total_deposits: 50000,
          total_allocations: 40000,
          total_consumption: 15000,
          total_refunds: 2000,
          available_balance: 73000
        };

        mockFundService.getFundSummary.mockResolvedValue(mockSummary);

        const { result } = renderHook(() => useFundStore());

        let summary: FundSummary;
        await act(async () => {
          summary = await result.current.actions.getFundSummary(1);
        });

        expect(summary!).toEqual(mockSummary);
        expect(result.current.summaryMap[1]).toEqual(mockSummary);
        expect(mockFundService.getFundSummary).toHaveBeenCalledWith(1);
      });
    });

    describe('loadMerchantList', () => {
      it('should load merchant list', async () => {
        const mockMerchants = [
          { id: 1, name: 'Merchant 1', code: 'M001' },
          { id: 2, name: 'Merchant 2', code: 'M002' }
        ];

        mockFundService.getMerchantList.mockResolvedValue(mockMerchants);

        const { result } = renderHook(() => useFundStore());

        await act(async () => {
          await result.current.actions.loadMerchantList();
        });

        expect(result.current.merchantList).toEqual(mockMerchants);
        expect(result.current.loading.merchantList).toBe(false);
        expect(mockFundService.getMerchantList).toHaveBeenCalled();
      });

      it('should not reload merchant list if already loaded and not forced', async () => {
        const mockMerchants = [
          { id: 1, name: 'Merchant 1', code: 'M001' }
        ];

        // Set initial merchant list
        const { result } = renderHook(() => useFundStore());
        act(() => {
          useFundStore.setState({
            merchantList: mockMerchants
          });
        });

        await act(async () => {
          await result.current.actions.loadMerchantList(false);
        });

        expect(mockFundService.getMerchantList).not.toHaveBeenCalled();
      });
    });

    describe('clearError', () => {
      it('should clear error state', () => {
        const { result } = renderHook(() => useFundStore());

        // Set error
        act(() => {
          useFundStore.setState({ error: 'Test error' });
        });

        expect(result.current.error).toBe('Test error');

        // Clear error
        act(() => {
          result.current.actions.clearError();
        });

        expect(result.current.error).toBe(null);
      });
    });

    describe('resetStore', () => {
      it('should reset store to initial state', () => {
        const { result } = renderHook(() => useFundStore());

        // Modify store state
        act(() => {
          useFundStore.setState({
            fundList: [{ id: 1 } as Fund],
            error: 'Test error',
            merchantList: [{ id: 1, name: 'Test', code: 'T001' }]
          });
        });

        expect(result.current.fundList).toHaveLength(1);
        expect(result.current.error).toBe('Test error');
        expect(result.current.merchantList).toHaveLength(1);

        // Reset store
        act(() => {
          result.current.actions.resetStore();
        });

        expect(result.current.fundList).toEqual([]);
        expect(result.current.error).toBe(null);
        expect(result.current.merchantList).toEqual([]);
        expect(result.current.balanceMap).toEqual({});
        expect(result.current.summaryMap).toEqual({});
      });
    });
  });

  describe('Hooks', () => {
    it('useFundActions should return actions', () => {
      const { result } = renderHook(() => useFundStore.getState().actions);
      
      expect(typeof result.current.deposit).toBe('function');
      expect(typeof result.current.allocate).toBe('function');
      expect(typeof result.current.getMerchantBalance).toBe('function');
      expect(typeof result.current.loadMerchantList).toBe('function');
      expect(typeof result.current.clearError).toBe('function');
      expect(typeof result.current.resetStore).toBe('function');
    });

    it('useFundData should return data state', () => {
      const { result } = renderHook(() => {
        const state = useFundStore.getState();
        return {
          fundList: state.fundList,
          transactionList: state.transactionList,
          balanceMap: state.balanceMap,
          summaryMap: state.summaryMap,
          merchantList: state.merchantList,
        };
      });
      
      expect(Array.isArray(result.current.fundList)).toBe(true);
      expect(Array.isArray(result.current.transactionList)).toBe(true);
      expect(typeof result.current.balanceMap).toBe('object');
      expect(typeof result.current.summaryMap).toBe('object');
      expect(Array.isArray(result.current.merchantList)).toBe(true);
    });
  });
});
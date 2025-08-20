// 资金服务测试

import { fundService } from '../fundService';
import type { DepositRequest, AllocateRequest, BatchDepositRequest, FreezeRequest } from '../../types/fund';

describe('FundService', () => {
  describe('Validation', () => {
    describe('validateDepositRequest', () => {
      it('should pass for valid deposit request', () => {
        const request: DepositRequest = {
          merchant_id: 1,
          amount: 1000,
          currency: 'CNY',
          description: 'Test deposit'
        };
        
        const errors = fundService.validateDepositRequest(request);
        expect(errors).toEqual([]);
      });
      
      it('should fail for invalid merchant_id', () => {
        const request: DepositRequest = {
          merchant_id: 0,
          amount: 1000,
          currency: 'CNY',
          description: 'Test deposit'
        };
        
        const errors = fundService.validateDepositRequest(request);
        expect(errors).toContain('请选择商户');
      });
      
      it('should fail for invalid amount', () => {
        const request: DepositRequest = {
          merchant_id: 1,
          amount: 0,
          currency: 'CNY',
          description: 'Test deposit'
        };
        
        const errors = fundService.validateDepositRequest(request);
        expect(errors).toContain('充值金额必须大于0');
      });
      
      it('should fail for amount too large', () => {
        const request: DepositRequest = {
          merchant_id: 1,
          amount: 2000000,
          currency: 'CNY',
          description: 'Test deposit'
        };
        
        const errors = fundService.validateDepositRequest(request);
        expect(errors).toContain('单笔充值金额不能超过1,000,000');
      });
      
      it('should fail for invalid currency', () => {
        const request: DepositRequest = {
          merchant_id: 1,
          amount: 1000,
          currency: 'XX',
          description: 'Test deposit'
        };
        
        const errors = fundService.validateDepositRequest(request);
        expect(errors).toContain('货币代码必须为3位');
      });
    });
    
    describe('validateAllocateRequest', () => {
      it('should pass for valid allocate request', () => {
        const request: AllocateRequest = {
          merchant_id: 1,
          amount: 1000,
          description: 'Test allocation'
        };
        
        const errors = fundService.validateAllocateRequest(request);
        expect(errors).toEqual([]);
      });
      
      it('should fail for invalid merchant_id', () => {
        const request: AllocateRequest = {
          merchant_id: 0,
          amount: 1000,
          description: 'Test allocation'
        };
        
        const errors = fundService.validateAllocateRequest(request);
        expect(errors).toContain('请选择商户');
      });
      
      it('should fail for invalid amount', () => {
        const request: AllocateRequest = {
          merchant_id: 1,
          amount: -100,
          description: 'Test allocation'
        };
        
        const errors = fundService.validateAllocateRequest(request);
        expect(errors).toContain('分配金额必须大于0');
      });
      
      it('should fail for amount too large', () => {
        const request: AllocateRequest = {
          merchant_id: 1,
          amount: 2000000,
          description: 'Test allocation'
        };
        
        const errors = fundService.validateAllocateRequest(request);
        expect(errors).toContain('单次分配金额不能超过1,000,000');
      });
    });
    
    describe('validateBatchDepositRequest', () => {
      it('should pass for valid batch request', () => {
        const request: BatchDepositRequest = {
          deposits: [
            { merchant_id: 1, amount: 1000, currency: 'CNY' },
            { merchant_id: 2, amount: 2000, currency: 'CNY' }
          ]
        };
        
        const errors = fundService.validateBatchDepositRequest(request);
        expect(errors).toEqual([]);
      });
      
      it('should fail for empty deposits', () => {
        const request: BatchDepositRequest = {
          deposits: []
        };
        
        const errors = fundService.validateBatchDepositRequest(request);
        expect(errors).toContain('批量充值列表不能为空');
      });
      
      it('should fail for too many deposits', () => {
        const deposits = Array(101).fill(null).map((_, index) => ({
          merchant_id: index + 1,
          amount: 1000,
          currency: 'CNY'
        }));
        
        const request: BatchDepositRequest = { deposits };
        const errors = fundService.validateBatchDepositRequest(request);
        expect(errors).toContain('单次批量充值不能超过100笔');
      });
      
      it('should fail for total amount too large', () => {
        const deposits = Array(10).fill(null).map((_, index) => ({
          merchant_id: index + 1,
          amount: 2000000,
          currency: 'CNY'
        }));
        
        const request: BatchDepositRequest = { deposits };
        const errors = fundService.validateBatchDepositRequest(request);
        expect(errors).toContain('批量充值总金额不能超过10,000,000');
      });
      
      it('should validate individual deposits', () => {
        const request: BatchDepositRequest = {
          deposits: [
            { merchant_id: 0, amount: 1000, currency: 'CNY' }, // Invalid merchant_id
            { merchant_id: 2, amount: 0, currency: 'CNY' }     // Invalid amount
          ]
        };
        
        const errors = fundService.validateBatchDepositRequest(request);
        expect(errors).toContain('第1笔充值: 请选择商户');
        expect(errors).toContain('第2笔充值: 充值金额必须大于0');
      });
    });
    
    describe('validateFreezeRequest', () => {
      it('should pass for valid freeze request', () => {
        const request: FreezeRequest = {
          action: 'freeze',
          amount: 1000,
          reason: 'Risk control'
        };
        
        const errors = fundService.validateFreezeRequest(request);
        expect(errors).toEqual([]);
      });
      
      it('should pass for valid unfreeze request', () => {
        const request: FreezeRequest = {
          action: 'unfreeze',
          amount: 500,
          reason: 'Risk cleared'
        };
        
        const errors = fundService.validateFreezeRequest(request);
        expect(errors).toEqual([]);
      });
      
      it('should fail for invalid action', () => {
        const request: FreezeRequest = {
          action: 'invalid' as any,
          amount: 1000,
          reason: 'Test'
        };
        
        const errors = fundService.validateFreezeRequest(request);
        expect(errors).toContain('操作类型无效');
      });
      
      it('should fail for invalid amount', () => {
        const request: FreezeRequest = {
          action: 'freeze',
          amount: 0,
          reason: 'Test'
        };
        
        const errors = fundService.validateFreezeRequest(request);
        expect(errors).toContain('操作金额必须大于0');
      });
    });
  });
  
  describe('Utility Functions', () => {
    describe('formatAmount', () => {
      it('should format CNY amounts correctly', () => {
        expect(fundService.formatAmount(1000.50, 'CNY')).toBe('¥1,000.50');
        expect(fundService.formatAmount(1234567.89, 'CNY')).toBe('¥1,234,567.89');
      });
      
      it('should format USD amounts correctly', () => {
        expect(fundService.formatAmount(1000.50, 'USD')).toBe('USD1,000.50');
      });
      
      it('should default to CNY', () => {
        expect(fundService.formatAmount(1000.50)).toBe('¥1,000.50');
      });
      
      it('should handle zero amounts', () => {
        expect(fundService.formatAmount(0)).toBe('¥0.00');
      });
      
      it('should handle negative amounts', () => {
        expect(fundService.formatAmount(-500.25)).toBe('¥-500.25');
      });
    });
  });
});
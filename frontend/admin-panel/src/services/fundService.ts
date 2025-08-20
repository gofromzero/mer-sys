// 资金管理API服务

import { api } from './api';
import type {
  Fund,
  FundTransaction,
  RightsBalance,
  DepositRequest,
  BatchDepositRequest,
  AllocateRequest,
  FundTransactionQuery,
  FundSummary,
  FreezeRequest,
  ApiResponse,
  PaginationResponse
} from '../types/fund';

export class FundService {
  // 单笔资金充值
  async deposit(request: DepositRequest): Promise<Fund> {
    const response = await api.post<ApiResponse<Fund>>('/api/v1/funds/deposit', request);
    if (response.data.code !== 0) {
      throw new Error(response.data.message || '充值失败');
    }
    return response.data.data!;
  }

  // 批量资金充值
  async batchDeposit(request: BatchDepositRequest): Promise<Fund[]> {
    const response = await api.post<ApiResponse<Fund[]>>('/api/v1/funds/batch-deposit', request);
    if (response.data.code !== 0) {
      throw new Error(response.data.message || '批量充值失败');
    }
    return response.data.data!;
  }

  // 权益分配
  async allocate(request: AllocateRequest): Promise<Fund> {
    const response = await api.post<ApiResponse<Fund>>('/api/v1/funds/allocate', request);
    if (response.data.code !== 0) {
      throw new Error(response.data.message || '权益分配失败');
    }
    return response.data.data!;
  }

  // 查询商户权益余额
  async getMerchantBalance(merchantId: number): Promise<RightsBalance> {
    const response = await api.get<ApiResponse<RightsBalance>>(`/api/v1/funds/balance/${merchantId}`);
    if (response.data.code !== 0) {
      throw new Error(response.data.message || '查询余额失败');
    }
    return response.data.data!;
  }

  // 查询资金流转历史
  async listTransactions(query: FundTransactionQuery): Promise<PaginationResponse<FundTransaction>> {
    const response = await api.get<ApiResponse<PaginationResponse<FundTransaction>>>('/api/v1/funds/transactions', {
      params: query
    });
    if (response.data.code !== 0) {
      throw new Error(response.data.message || '查询交易记录失败');
    }
    return response.data.data!;
  }

  // 获取资金概览统计
  async getFundSummary(merchantId?: number): Promise<FundSummary> {
    const params = merchantId ? { merchant_id: merchantId } : {};
    const response = await api.get<ApiResponse<FundSummary>>('/api/v1/funds/summary', { params });
    if (response.data.code !== 0) {
      throw new Error(response.data.message || '获取统计失败');
    }
    return response.data.data!;
  }

  // 冻结/解冻商户权益
  async freezeMerchantBalance(merchantId: number, request: FreezeRequest): Promise<void> {
    const response = await api.put<ApiResponse>(`/api/v1/funds/freeze/${merchantId}`, request);
    if (response.data.code !== 0) {
      throw new Error(response.data.message || '操作失败');
    }
  }

  // 获取所有商户列表（用于下拉选择）
  async getMerchantList(): Promise<Array<{ id: number; name: string; code: string }>> {
    try {
      const response = await api.get<ApiResponse<PaginationResponse<any>>>('/api/v1/merchants', {
        params: { page: 1, page_size: 1000 }
      });
      if (response.data.code !== 0) {
        throw new Error(response.data.message || '获取商户列表失败');
      }
      return response.data.data!.list.map(merchant => ({
        id: merchant.id,
        name: merchant.name,
        code: merchant.code
      }));
    } catch (error) {
      console.error('获取商户列表失败:', error);
      return [];
    }
  }

  // 格式化金额显示
  formatAmount(amount: number, currency: string = 'CNY'): string {
    const currencySymbol = currency === 'CNY' ? '¥' : currency;
    return `${currencySymbol}${amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  }

  // 验证充值请求
  validateDepositRequest(request: DepositRequest): string[] {
    const errors: string[] = [];

    if (!request.merchant_id || request.merchant_id <= 0) {
      errors.push('请选择商户');
    }

    if (!request.amount || request.amount <= 0) {
      errors.push('充值金额必须大于0');
    }

    if (request.amount > 1000000) {
      errors.push('单笔充值金额不能超过1,000,000');
    }

    if (!request.currency || request.currency.length !== 3) {
      errors.push('货币代码必须为3位');
    }

    return errors;
  }

  // 验证权益分配请求
  validateAllocateRequest(request: AllocateRequest): string[] {
    const errors: string[] = [];

    if (!request.merchant_id || request.merchant_id <= 0) {
      errors.push('请选择商户');
    }

    if (!request.amount || request.amount <= 0) {
      errors.push('分配金额必须大于0');
    }

    if (request.amount > 1000000) {
      errors.push('单次分配金额不能超过1,000,000');
    }

    return errors;
  }

  // 验证批量充值请求
  validateBatchDepositRequest(request: BatchDepositRequest): string[] {
    const errors: string[] = [];

    if (!request.deposits || request.deposits.length === 0) {
      errors.push('批量充值列表不能为空');
    }

    if (request.deposits && request.deposits.length > 100) {
      errors.push('单次批量充值不能超过100笔');
    }

    let totalAmount = 0;
    request.deposits?.forEach((deposit, index) => {
      const depositErrors = this.validateDepositRequest(deposit);
      depositErrors.forEach(error => {
        errors.push(`第${index + 1}笔充值: ${error}`);
      });
      totalAmount += deposit.amount || 0;
    });

    if (totalAmount > 10000000) {
      errors.push('批量充值总金额不能超过10,000,000');
    }

    return errors;
  }

  // 验证冻结/解冻请求
  validateFreezeRequest(request: FreezeRequest): string[] {
    const errors: string[] = [];

    if (!request.action || !['freeze', 'unfreeze'].includes(request.action)) {
      errors.push('操作类型无效');
    }

    if (!request.amount || request.amount <= 0) {
      errors.push('操作金额必须大于0');
    }

    return errors;
  }
}

// 导出单例实例
export const fundService = new FundService();
export default fundService;
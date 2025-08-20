import { apiClient } from './api';
import type {
  Merchant,
  MerchantRegistrationRequest,
  MerchantApprovalRequest,
  MerchantUpdateRequest,
  MerchantStatusUpdateRequest,
  MerchantListQuery,
  MerchantListResponse
} from '../types/merchant';

export class MerchantService {
  private static readonly BASE_URL = '/api/v1/merchants';

  /**
   * 获取商户列表
   */
  static async getMerchantList(query: MerchantListQuery = {}): Promise<MerchantListResponse> {
    const response = await apiClient.get(this.BASE_URL, { params: query });
    return response.data.data;
  }

  /**
   * 获取商户详情
   */
  static async getMerchantById(id: number): Promise<Merchant> {
    const response = await apiClient.get(`${this.BASE_URL}/${id}`);
    return response.data.data;
  }

  /**
   * 注册商户申请
   */
  static async registerMerchant(data: MerchantRegistrationRequest): Promise<Merchant> {
    const response = await apiClient.post(this.BASE_URL, data);
    return response.data.data;
  }

  /**
   * 更新商户信息
   */
  static async updateMerchant(id: number, data: MerchantUpdateRequest): Promise<Merchant> {
    const response = await apiClient.put(`${this.BASE_URL}/${id}`, data);
    return response.data.data;
  }

  /**
   * 更新商户状态
   */
  static async updateMerchantStatus(id: number, data: MerchantStatusUpdateRequest): Promise<void> {
    await apiClient.put(`${this.BASE_URL}/${id}/status`, data);
  }

  /**
   * 审批商户申请
   */
  static async approveMerchant(id: number, comment?: string): Promise<void> {
    const data: MerchantApprovalRequest = { action: 'approve', comment };
    await apiClient.post(`${this.BASE_URL}/${id}/approve`, data);
  }

  /**
   * 拒绝商户申请
   */
  static async rejectMerchant(id: number, comment?: string): Promise<void> {
    const data: MerchantApprovalRequest = { action: 'reject', comment };
    await apiClient.post(`${this.BASE_URL}/${id}/reject`, data);
  }

  /**
   * 获取商户操作历史
   */
  static async getMerchantAuditLog(id: number): Promise<any[]> {
    const response = await apiClient.get(`${this.BASE_URL}/${id}/audit-log`);
    return response.data.data;
  }
}
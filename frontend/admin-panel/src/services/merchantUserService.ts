import { api } from './api';
import type {
  MerchantUser,
  CreateMerchantUserRequest,
  UpdateMerchantUserRequest,
  MerchantUserStatusRequest,
  MerchantUserPasswordResetRequest,
  MerchantUserAuditLog,
  MerchantUserQueryParams
} from '../types/merchantUser';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedResponse<T> {
  list: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export const MerchantUserService = {
  /**
   * 获取商户用户列表
   */
  async getMerchantUsers(params?: MerchantUserQueryParams): Promise<PaginatedResponse<MerchantUser>> {
    const response = await api.get<ApiResponse<PaginatedResponse<MerchantUser>>>(
      '/api/v1/merchant-users', 
      { params }
    );
    return response.data.data;
  },

  /**
   * 获取单个商户用户详情
   */
  async getMerchantUser(id: number): Promise<MerchantUser> {
    const response = await api.get<ApiResponse<MerchantUser>>(`/api/v1/merchant-users/${id}`);
    return response.data.data;
  },

  /**
   * 创建商户用户
   */
  async createMerchantUser(data: CreateMerchantUserRequest): Promise<MerchantUser> {
    const response = await api.post<ApiResponse<MerchantUser>>('/api/v1/merchant-users', data);
    return response.data.data;
  },

  /**
   * 更新商户用户信息
   */
  async updateMerchantUser(id: number, data: UpdateMerchantUserRequest): Promise<MerchantUser> {
    const response = await api.put<ApiResponse<MerchantUser>>(`/api/v1/merchant-users/${id}`, data);
    return response.data.data;
  },

  /**
   * 更新商户用户状态
   */
  async updateMerchantUserStatus(id: number, data: MerchantUserStatusRequest): Promise<MerchantUser> {
    const response = await api.put<ApiResponse<MerchantUser>>(
      `/api/v1/merchant-users/${id}/status`, 
      data
    );
    return response.data.data;
  },

  /**
   * 重置商户用户密码
   */
  async resetMerchantUserPassword(id: number, data: MerchantUserPasswordResetRequest): Promise<void> {
    await api.post(`/api/v1/merchant-users/${id}/reset-password`, data);
  },

  /**
   * 获取商户用户操作日志
   */
  async getMerchantUserAuditLog(
    id: number,
    params?: { page?: number; page_size?: number }
  ): Promise<PaginatedResponse<MerchantUserAuditLog>> {
    const response = await api.get<ApiResponse<PaginatedResponse<MerchantUserAuditLog>>>(
      `/api/v1/merchant-users/${id}/audit-log`,
      { params }
    );
    return response.data.data;
  },

  /**
   * 批量创建商户用户
   */
  async createMerchantUsersBatch(data: CreateMerchantUserRequest[]): Promise<MerchantUser[]> {
    const response = await api.post<ApiResponse<MerchantUser[]>>('/api/v1/merchant-users/batch', {
      users: data
    });
    return response.data.data;
  },

  /**
   * 删除商户用户（软删除）
   */
  async deleteMerchantUser(id: number): Promise<void> {
    await api.delete(`/api/v1/merchant-users/${id}`);
  }
};
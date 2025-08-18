import { apiClient } from './api';

export interface Tenant {
  id: number;
  name: string;
  code: string;
  status: string;
  business_type: string;
  contact_person: string;
  contact_email: string;
  contact_phone: string;
  address: string;
  registration_time?: string;
  activation_time?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTenantRequest {
  name: string;
  code: string;
  business_type: string;
  contact_person: string;
  contact_email: string;
  contact_phone?: string;
  address?: string;
}

export interface UpdateTenantRequest {
  name?: string;
  business_type?: string;
  contact_person?: string;
  contact_email?: string;
  contact_phone?: string;
  address?: string;
}

export interface UpdateTenantStatusRequest {
  status: 'active' | 'suspended' | 'expired';
  reason: string;
}

export interface ListTenantsRequest {
  page?: number;
  page_size?: number;
  status?: string;
  business_type?: string;
  search?: string;
}

export interface ListTenantsResponse {
  total: number;
  page: number;
  size: number;
  tenants: Tenant[];
}

export interface TenantConfig {
  max_users: number;
  max_merchants: number;
  features: string[];
  settings: Record<string, string>;
}

export const tenantService = {
  /**
   * 创建租户
   */
  async createTenant(data: CreateTenantRequest): Promise<Tenant> {
    const response = await apiClient.post('/tenants', data);
    return response.data;
  },

  /**
   * 获取租户列表
   */
  async listTenants(params: ListTenantsRequest = {}): Promise<ListTenantsResponse> {
    const response = await apiClient.get('/tenants', { params });
    return response.data;
  },

  /**
   * 获取租户详情
   */
  async getTenant(id: number): Promise<Tenant> {
    const response = await apiClient.get(`/tenants/${id}`);
    return response.data;
  },

  /**
   * 更新租户信息
   */
  async updateTenant(id: number, data: UpdateTenantRequest): Promise<Tenant> {
    const response = await apiClient.put(`/tenants/${id}`, data);
    return response.data;
  },

  /**
   * 更新租户状态
   */
  async updateTenantStatus(id: number, data: UpdateTenantStatusRequest): Promise<void> {
    await apiClient.put(`/tenants/${id}/status`, data);
  },

  /**
   * 获取租户配置
   */
  async getTenantConfig(id: number): Promise<TenantConfig> {
    const response = await apiClient.get(`/tenants/${id}/config`);
    return response.data;
  },

  /**
   * 更新租户配置
   */
  async updateTenantConfig(id: number, config: TenantConfig): Promise<void> {
    await apiClient.put(`/tenants/${id}/config`, config);
  },
};
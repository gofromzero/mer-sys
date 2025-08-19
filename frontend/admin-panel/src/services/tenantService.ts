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

  /**
   * 获取配置变更通知
   */
  async getConfigChangeNotification(id: number): Promise<unknown> {
    const response = await apiClient.get(`/tenants/${id}/config/notifications`);
    return response.data;
  },

  /**
   * 批量获取租户信息
   */
  async batchGetTenants(ids: number[]): Promise<Tenant[]> {
    const promises = ids.map(id => this.getTenant(id));
    return Promise.all(promises);
  },

  /**
   * 搜索租户
   */
  async searchTenants(keyword: string, filters?: {
    status?: string;
    business_type?: string;
  }): Promise<ListTenantsResponse> {
    return this.listTenants({
      search: keyword,
      ...filters,
      page: 1,
      page_size: 50
    });
  },

  /**
   * 获取租户统计信息
   */
  async getTenantStats(): Promise<{
    total: number;
    active: number;
    suspended: number;
    expired: number;
    by_business_type: Record<string, number>;
  }> {
    // 获取所有租户进行统计
    const allTenants = await this.listTenants({ page_size: 1000 });
    
    const stats = {
      total: allTenants.total,
      active: 0,
      suspended: 0,
      expired: 0,
      by_business_type: {} as Record<string, number>
    };

    allTenants.tenants.forEach(tenant => {
      // 统计状态
      switch (tenant.status) {
        case 'active':
          stats.active++;
          break;
        case 'suspended':
          stats.suspended++;
          break;
        case 'expired':
          stats.expired++;
          break;
      }

      // 统计业务类型
      if (tenant.business_type) {
        stats.by_business_type[tenant.business_type] = 
          (stats.by_business_type[tenant.business_type] || 0) + 1;
      }
    });

    return stats;
  },

  /**
   * 验证租户代码是否可用
   */
  async checkTenantCodeAvailability(code: string): Promise<boolean> {
    try {
      // 通过代码搜索租户
      const result = await this.searchTenants(code);
      // 如果找到匹配的租户，说明代码已被使用
      return result.tenants.length === 0;
    } catch {
      // 如果搜索失败，说明代码可用
      return true;
    }
  },

  /**
   * 验证联系邮箱是否可用
   */
  async checkEmailAvailability(email: string): Promise<boolean> {
    try {
      // 通过邮箱搜索租户
      const result = await this.searchTenants(email);
      // 如果找到匹配的租户，说明邮箱已被使用
      return result.tenants.length === 0;
    } catch {
      // 如果搜索失败，说明邮箱可用
      return true;
    }
  },
};
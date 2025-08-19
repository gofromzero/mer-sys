import { apiClient } from './api';

export interface AuditLog {
  id: number;
  tenant_id: number;
  user_id: number;
  action: string;
  resource_type: string;
  resource_id: string;
  old_values?: Record<string, unknown>;
  new_values?: Record<string, unknown>;
  ip_address: string;
  user_agent: string;
  created_at: string;
  user_name?: string;
  tenant_name?: string;
}

export interface AuditLogQuery {
  tenant_id?: number;
  user_id?: number;
  action?: string;
  resource_type?: string;
  resource_id?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  page_size?: number;
}

export interface AuditLogResponse {
  total: number;
  page: number;
  size: number;
  logs: AuditLog[];
}

/**
 * 审计日志服务
 */
export const auditService = {
  /**
   * 获取审计日志列表
   */
  async getAuditLogs(query: AuditLogQuery = {}): Promise<AuditLogResponse> {
    const response = await apiClient.get('/audit/logs', { params: query });
    return response.data;
  },

  /**
   * 获取特定租户的审计日志
   */
  async getTenantAuditLogs(tenantId: number, query: Omit<AuditLogQuery, 'tenant_id'> = {}): Promise<AuditLogResponse> {
    const response = await apiClient.get(`/tenants/${tenantId}/audit-logs`, { params: query });
    return response.data;
  },

  /**
   * 获取审计日志详情
   */
  async getAuditLogDetail(id: number): Promise<AuditLog> {
    const response = await apiClient.get(`/audit/logs/${id}`);
    return response.data;
  },

  /**
   * 记录租户操作日志
   */
  async logTenantOperation(data: {
    action: string;
    tenant_id: number;
    old_values?: Record<string, unknown>;
    new_values?: Record<string, unknown>;
    notes?: string;
  }): Promise<void> {
    await apiClient.post('/audit/tenant-operation', data);
  },

  /**
   * 获取操作统计
   */
  async getOperationStats(query: {
    tenant_id?: number;
    start_date?: string;
    end_date?: string;
  } = {}): Promise<{
    total_operations: number;
    operations_by_action: Record<string, number>;
    operations_by_user: Record<string, number>;
    operations_by_date: Record<string, number>;
  }> {
    const response = await apiClient.get('/audit/stats', { params: query });
    return response.data;
  },

  /**
   * 导出审计日志
   */
  async exportAuditLogs(query: AuditLogQuery = {}, format: 'csv' | 'excel' = 'csv'): Promise<Blob> {
    const response = await apiClient.get('/audit/export', {
      params: { ...query, format },
      responseType: 'blob'
    });
    return response.data;
  }
};

/**
 * 审计日志操作类型
 */
export const AUDIT_ACTIONS = {
  TENANT_CREATE: 'tenant:create',
  TENANT_UPDATE: 'tenant:update',
  TENANT_DELETE: 'tenant:delete',
  TENANT_STATUS_CHANGE: 'tenant:status_change',
  TENANT_CONFIG_UPDATE: 'tenant:config_update',
  TENANT_VIEW: 'tenant:view',
  TENANT_EXPORT: 'tenant:export'
} as const;

/**
 * 资源类型
 */
export const RESOURCE_TYPES = {
  TENANT: 'tenant',
  TENANT_CONFIG: 'tenant_config',
  USER: 'user',
  ROLE: 'role',
  PERMISSION: 'permission'
} as const;
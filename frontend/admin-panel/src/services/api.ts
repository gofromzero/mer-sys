import axios from 'axios';
import type { APIResponse } from '../types';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor for adding auth token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  
  const tenantId = localStorage.getItem('tenant_id');
  if (tenantId) {
    config.headers['X-Tenant-ID'] = tenantId;
  }
  
  return config;
});

// Response interceptor for handling errors
apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('tenant_id');
      window.location.href = '/auth/login';
    }
    return Promise.reject(error);
  }
);

export { apiClient };
export const api = apiClient; // 为了兼容性添加 api 导出
export const apiService = apiClient; // 为了兼容性添加 apiService 导出

// Health check service
export const healthService = {
  async check(): Promise<APIResponse> {
    return apiClient.get('/health');
  },
};
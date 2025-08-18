import { apiClient } from './api';
import type { APIResponse, User } from '../types';

// 登录请求参数
export interface LoginRequest {
  username: string;
  password: string;
  tenant_id?: number;
  remember_me?: boolean;
}

// 登录响应数据
export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
  permissions: string[];
}

// 刷新令牌请求参数
export interface RefreshTokenRequest {
  refresh_token: string;
}

// 刷新令牌响应数据
export interface RefreshTokenResponse {
  access_token: string;
  refresh_token: string;
}

// 认证API服务
export const authService = {
  /**
   * 用户登录
   * @param credentials 登录凭证
   * @returns 登录响应数据
   */
  async login(credentials: LoginRequest): Promise<APIResponse<LoginResponse>> {
    try {
      const response = await apiClient.post<APIResponse<LoginResponse>>('/auth/login', credentials);
      return response.data;
    } catch (error: unknown) {
      // 统一错误处理
      const err = error as { response?: { data?: { message?: string } } };
      throw new Error(err.response?.data?.message || '登录失败，请检查用户名和密码');
    }
  },

  /**
   * 用户登出
   * @returns 登出响应
   */
  async logout(): Promise<APIResponse> {
    try {
      const response = await apiClient.post<APIResponse>('/auth/logout');
      return response.data;
    } catch (error: unknown) {
      // 即使登出失败也要清除本地存储
      const err = error as { response?: { data?: { message?: string } } };
      console.warn('登出请求失败:', err.response?.data?.message);
      throw new Error(err.response?.data?.message || '登出失败');
    }
  },

  /**
   * 刷新访问令牌
   * @param refreshToken 刷新令牌
   * @returns 新的令牌对
   */
  async refreshToken(refreshToken: string): Promise<APIResponse<RefreshTokenResponse>> {
    try {
      const response = await apiClient.post<APIResponse<RefreshTokenResponse>>('/auth/refresh', {
        refresh_token: refreshToken
      });
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } } };
      throw new Error(err.response?.data?.message || '令牌刷新失败');
    }
  },

  /**
   * 验证当前令牌是否有效
   * @returns 用户信息
   */
  async validateToken(): Promise<APIResponse<{ user: User; permissions: string[] }>> {
    try {
      const response = await apiClient.get<APIResponse<{ user: User; permissions: string[] }>>('/auth/me');
      return response.data;
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } } };
      throw new Error(err.response?.data?.message || '令牌验证失败');
    }
  }
};
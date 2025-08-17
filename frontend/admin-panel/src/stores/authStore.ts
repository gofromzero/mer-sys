import { create } from 'zustand';
import type { AuthState, User } from '../types';
import { authService, type LoginRequest } from '../services/authService';

interface AuthActions {
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  setRefreshToken: (refreshToken: string | null) => void;
  setPermissions: (permissions: string[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  refreshAccessToken: () => Promise<boolean>;
  validateToken: () => Promise<boolean>;
  clearError: () => void;
}

interface ExtendedAuthState extends AuthState {
  refreshToken: string | null;
  error: string | null;
}

type AuthStore = ExtendedAuthState & AuthActions;

export const useAuthStore = create<AuthStore>((set, get) => ({
  // State
  user: null,
  token: localStorage.getItem('token'),
  refreshToken: localStorage.getItem('refresh_token'),
  permissions: [],
  isLoading: false,
  error: null,

  // Actions
  setUser: (user) => set({ user }),
  
  setToken: (token) => {
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
    set({ token });
  },
  
  setRefreshToken: (refreshToken) => {
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken);
    } else {
      localStorage.removeItem('refresh_token');
    }
    set({ refreshToken });
  },
  
  setPermissions: (permissions) => set({ permissions }),
  setLoading: (isLoading) => set({ isLoading }),
  setError: (error) => set({ error }),
  clearError: () => set({ error: null }),
  
  /**
   * 用户登录
   */
  login: async (credentials) => {
    set({ isLoading: true, error: null });
    
    try {
      const response = await authService.login(credentials);
      
      if (response.data) {
        const { access_token, refresh_token, user, permissions } = response.data;
        
        // 存储令牌
        localStorage.setItem('token', access_token);
        localStorage.setItem('refresh_token', refresh_token);
        localStorage.setItem('tenant_id', user.tenant_id.toString());
        
        // 如果选择了记住我，设置更长的过期时间
        if (credentials.remember_me) {
          localStorage.setItem('remember_me', 'true');
        }
        
        set({ 
          user, 
          token: access_token, 
          refreshToken: refresh_token,
          permissions,
          isLoading: false,
          error: null
        });
      }
    } catch (error: any) {
      set({ 
        isLoading: false, 
        error: error.message || '登录失败' 
      });
      throw error;
    }
  },
  
  /**
   * 用户登出
   */
  logout: async () => {
    set({ isLoading: true });
    
    try {
      await authService.logout();
    } catch (error) {
      console.warn('登出API调用失败:', error);
    } finally {
      // 无论API调用是否成功，都要清除本地状态
      localStorage.removeItem('token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('tenant_id');
      localStorage.removeItem('remember_me');
      
      set({ 
        user: null, 
        token: null, 
        refreshToken: null,
        permissions: [], 
        isLoading: false,
        error: null
      });
    }
  },
  
  /**
   * 刷新访问令牌
   */
  refreshAccessToken: async () => {
    const { refreshToken } = get();
    
    if (!refreshToken) {
      return false;
    }
    
    try {
      const response = await authService.refreshToken(refreshToken);
      
      if (response.data) {
        const { access_token, refresh_token } = response.data;
        
        localStorage.setItem('token', access_token);
        localStorage.setItem('refresh_token', refresh_token);
        
        set({ 
          token: access_token, 
          refreshToken: refresh_token,
          error: null
        });
        
        return true;
      }
    } catch (error) {
      console.error('令牌刷新失败:', error);
      // 刷新失败，清除所有认证信息
      get().logout();
    }
    
    return false;
  },
  
  /**
   * 验证当前令牌
   */
  validateToken: async () => {
    const { token } = get();
    
    if (!token) {
      return false;
    }
    
    try {
      const response = await authService.validateToken();
      
      if (response.data) {
        const { user, permissions } = response.data;
        set({ user, permissions });
        return true;
      }
    } catch (error) {
      console.error('令牌验证失败:', error);
      // 验证失败，尝试刷新令牌
      return await get().refreshAccessToken();
    }
    
    return false;
  },
}));
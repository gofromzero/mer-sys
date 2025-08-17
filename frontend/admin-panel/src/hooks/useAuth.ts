import { useEffect } from 'react';
import { useAuthStore } from '../stores/authStore';
import type { LoginRequest } from '../services/authService';

/**
 * 认证Hook - 提供认证相关的状态和方法
 */
export const useAuth = () => {
  const {
    user,
    token,
    refreshToken,
    permissions,
    isLoading,
    error,
    login,
    logout,
    refreshAccessToken,
    validateToken,
    clearError,
  } = useAuthStore();

  // 检查用户是否已登录
  const isAuthenticated = !!token && !!user;

  // 检查用户是否有特定权限
  const hasPermission = (permission: string): boolean => {
    return permissions.includes(permission);
  };

  // 检查用户是否有任一权限
  const hasAnyPermission = (permissionList: string[]): boolean => {
    return permissionList.some(permission => permissions.includes(permission));
  };

  // 检查用户是否有所有权限
  const hasAllPermissions = (permissionList: string[]): boolean => {
    return permissionList.every(permission => permissions.includes(permission));
  };

  // 获取用户角色（从权限中推断）
  const getUserRoles = (): string[] => {
    const roles: string[] = [];
    
    if (permissions.includes('tenant_admin')) {
      roles.push('tenant_admin');
    }
    if (permissions.includes('merchant_admin')) {
      roles.push('merchant');
    }
    if (permissions.includes('customer_view')) {
      roles.push('customer');
    }
    
    return roles;
  };

  // 检查用户是否有特定角色
  const hasRole = (role: string): boolean => {
    const userRoles = getUserRoles();
    return userRoles.includes(role);
  };

  // 登录方法的包装
  const handleLogin = async (credentials: LoginRequest): Promise<void> => {
    try {
      await login(credentials);
    } catch (error) {
      // 错误已经在store中处理，这里可以添加额外的错误处理逻辑
      throw error;
    }
  };

  // 登出方法的包装
  const handleLogout = async (): Promise<void> => {
    try {
      await logout();
    } catch (error) {
      console.error('登出失败:', error);
      // 即使登出失败也要清除本地状态（store中已处理）
    }
  };

  // 自动刷新令牌
  const autoRefreshToken = async (): Promise<boolean> => {
    if (!refreshToken) {
      return false;
    }
    
    try {
      return await refreshAccessToken();
    } catch (error) {
      console.error('自动刷新令牌失败:', error);
      return false;
    }
  };

  // 初始化认证状态
  const initializeAuth = async (): Promise<void> => {
    if (token) {
      // 验证现有令牌
      const isValid = await validateToken();
      if (!isValid) {
        // 令牌无效，尝试刷新
        const refreshed = await autoRefreshToken();
        if (!refreshed) {
          // 刷新失败，清除认证状态
          await handleLogout();
        }
      }
    }
  };

  // 组件挂载时初始化认证状态
  useEffect(() => {
    initializeAuth();
  }, []);

  // 定期检查令牌有效性（可选）
  useEffect(() => {
    if (!isAuthenticated) {
      return;
    }

    // 每30分钟检查一次令牌有效性
    const interval = setInterval(async () => {
      const isValid = await validateToken();
      if (!isValid) {
        await autoRefreshToken();
      }
    }, 30 * 60 * 1000); // 30分钟

    return () => clearInterval(interval);
  }, [isAuthenticated]);

  return {
    // 状态
    user,
    token,
    refreshToken,
    permissions,
    isLoading,
    error,
    isAuthenticated,
    
    // 方法
    login: handleLogin,
    logout: handleLogout,
    refreshAccessToken: autoRefreshToken,
    validateToken,
    clearError,
    
    // 权限检查方法
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    hasRole,
    getUserRoles,
    
    // 工具方法
    initializeAuth,
  };
};

/**
 * 权限守卫Hook - 用于保护需要特定权限的组件
 */
export const usePermissionGuard = (requiredPermissions: string | string[]) => {
  const { hasPermission, hasAnyPermission, isAuthenticated } = useAuth();
  
  const permissions = Array.isArray(requiredPermissions) ? requiredPermissions : [requiredPermissions];
  
  const hasAccess = isAuthenticated && (
    permissions.length === 1 
      ? hasPermission(permissions[0])
      : hasAnyPermission(permissions)
  );
  
  return {
    hasAccess,
    isAuthenticated,
  };
};

/**
 * 角色守卫Hook - 用于保护需要特定角色的组件
 */
export const useRoleGuard = (requiredRoles: string | string[]) => {
  const { hasRole, isAuthenticated } = useAuth();
  
  const roles = Array.isArray(requiredRoles) ? requiredRoles : [requiredRoles];
  
  const hasAccess = isAuthenticated && roles.some(role => hasRole(role));
  
  return {
    hasAccess,
    isAuthenticated,
  };
};
import { useAuthStore } from '../stores/authStore';
import { useMemo } from 'react';

/**
 * 租户管理权限Hook
 * 提供租户相关操作的权限检查功能
 * 
 * @description 此Hook提供对租户管理功能的细粒度权限控制，
 * 包括查看、创建、编辑、删除等操作的权限检查。
 * 所有权限检查都经过性能优化，使用memoization避免不必要的重计算。
 * 
 * @example
 * ```tsx
 * const { canViewTenants, canCreateTenant } = useTenantPermissions();
 * 
 * if (!canViewTenants) {
 *   return <NoPermissionMessage />;
 * }
 * ```
 */
export const useTenantPermissions = () => {
  const permissions = useAuthStore((state) => state.permissions);
  
  // 使用 useMemo 优化权限检查计算，避免每次渲染都重新计算
  const permissionCheckers = useMemo(() => {
    /**
     * 检查是否有指定权限
     * @param permission 权限字符串
     * @returns 是否有权限
     */
    const hasPermission = (permission: string): boolean => {
      return permissions.includes(permission) || permissions.includes('tenant:*');
    };

    /**
     * 检查是否有多个权限中的任意一个
     * @param permissionList 权限列表
     * @returns 是否有任意一个权限
     */
    const hasAnyPermission = (permissionList: string[]): boolean => {
      return permissionList.some(permission => hasPermission(permission));
    };

    /**
     * 检查是否有所有指定权限
     * @param permissionList 权限列表
     * @returns 是否有所有权限
     */
    const hasAllPermissions = (permissionList: string[]): boolean => {
      return permissionList.every(permission => hasPermission(permission));
    };

    return { hasPermission, hasAnyPermission, hasAllPermissions };
  }, [permissions]);

  // 使用 useMemo 缓存具体的权限检查结果
  const specificPermissions = useMemo(() => {
    const { hasPermission, hasAnyPermission } = permissionCheckers;
    
    return {
      // 租户管理具体权限
      canViewTenants: hasPermission('tenant:view'),
      canCreateTenant: hasPermission('tenant:create'),
      canEditTenant: hasPermission('tenant:edit'),
      canDeleteTenant: hasPermission('tenant:delete'),
      canManageTenantStatus: hasPermission('tenant:manage_status'),
      canManageTenantConfig: hasPermission('tenant:manage_config'),
      canViewTenantAuditLog: hasPermission('tenant:view_audit'),
      
      // 敏感操作权限（需要二次确认）
      canPerformSensitiveOperations: hasAnyPermission([
        'tenant:delete',
        'tenant:manage_status',
        'tenant:manage_config'
      ])
    };
  }, [permissionCheckers]);

  // 创建优化的辅助函数，使用缓存的权限检查结果
  const helperFunctions = useMemo(() => {
    const {
      canViewTenants,
      canCreateTenant,
      canEditTenant,
      canDeleteTenant,
      canManageTenantStatus,
      canManageTenantConfig,
      canPerformSensitiveOperations
    } = specificPermissions;

    return {
      /**
       * 权限检查助手函数
       * @param operation 操作类型
       * @returns 是否有该操作的权限
       */
      checkTenantOperation: (operation: 'view' | 'create' | 'edit' | 'delete' | 'manage_status' | 'manage_config'): boolean => {
        switch (operation) {
          case 'view':
            return canViewTenants;
          case 'create':
            return canCreateTenant;
          case 'edit':
            return canEditTenant;
          case 'delete':
            return canDeleteTenant;
          case 'manage_status':
            return canManageTenantStatus;
          case 'manage_config':
            return canManageTenantConfig;
          default:
            return false;
        }
      },

      /**
       * 检查是否需要二次确认
       * @param operation 敏感操作类型
       * @returns 是否需要二次确认
       */
      requiresConfirmation: (operation: 'delete' | 'manage_status' | 'manage_config'): boolean => {
        const sensitiveOperations = ['delete', 'manage_status', 'manage_config'] as const;
        return sensitiveOperations.includes(operation) && canPerformSensitiveOperations;
      }
    };
  }, [permissionCheckers, specificPermissions]);

  return {
    // 基础权限检查函数
    ...permissionCheckers,
    
    // 租户管理具体权限
    ...specificPermissions,

    // 权限检查助手函数
    ...helperFunctions
  };
};

/**
 * 租户权限常量
 */
export const TENANT_PERMISSIONS = {
  VIEW: 'tenant:view',
  CREATE: 'tenant:create',
  EDIT: 'tenant:edit',
  DELETE: 'tenant:delete',
  MANAGE_STATUS: 'tenant:manage_status',
  MANAGE_CONFIG: 'tenant:manage_config',
  VIEW_AUDIT: 'tenant:view_audit',
  ALL: 'tenant:*'
} as const;

/**
 * 权限级别定义
 */
export const PERMISSION_LEVELS = {
  READONLY: ['tenant:view'],
  OPERATOR: ['tenant:view', 'tenant:create', 'tenant:edit'],
  MANAGER: ['tenant:view', 'tenant:create', 'tenant:edit', 'tenant:manage_status'],
  ADMIN: ['tenant:*']
} as const;
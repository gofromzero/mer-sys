import { useMemo } from 'react';
import { useAuth } from './useAuth';
import type { MerchantRoleType } from '../types/merchantUser';
import { MERCHANT_PERMISSIONS } from '../types/merchantUser';

export interface MerchantPermissionContext {
  merchant_id?: number;
  role_type?: MerchantRoleType;
  permissions: string[];
}

/**
 * 商户权限管理Hook
 * 提供商户级别的权限检查功能
 */
export const useMerchantPermissions = () => {
  const { user, isAuthenticated } = useAuth();

  // 获取当前用户的商户权限上下文
  const merchantContext = useMemo<MerchantPermissionContext | null>(() => {
    if (!isAuthenticated || !user?.merchant_id) {
      return null;
    }

    // 从用户角色中提取商户权限
    const merchantRole = user.roles?.find(role => 
      role.role_type === 'merchant_admin' || role.role_type === 'merchant_operator'
    );

    if (!merchantRole) {
      return null;
    }

    return {
      merchant_id: user.merchant_id,
      role_type: merchantRole.role_type as MerchantRoleType,
      permissions: merchantRole.permissions || []
    };
  }, [user, isAuthenticated]);

  /**
   * 检查是否具有特定商户权限
   */
  const hasMerchantPermission = (permission: string): boolean => {
    if (!merchantContext) return false;
    return merchantContext.permissions.includes(permission);
  };

  /**
   * 检查是否具有任意一个商户权限
   */
  const hasAnyMerchantPermission = (permissions: string[]): boolean => {
    if (!merchantContext) return false;
    return permissions.some(permission => merchantContext.permissions.includes(permission));
  };

  /**
   * 检查是否具有所有指定的商户权限
   */
  const hasAllMerchantPermissions = (permissions: string[]): boolean => {
    if (!merchantContext) return false;
    return permissions.every(permission => merchantContext.permissions.includes(permission));
  };

  /**
   * 检查是否具有特定商户角色
   */
  const hasMerchantRole = (roleType: MerchantRoleType): boolean => {
    if (!merchantContext) return false;
    return merchantContext.role_type === roleType;
  };

  /**
   * 检查是否为商户管理员
   */
  const isMerchantAdmin = (): boolean => {
    return hasMerchantRole('merchant_admin');
  };

  /**
   * 检查是否为商户操作员
   */
  const isMerchantOperator = (): boolean => {
    return hasMerchantRole('merchant_operator');
  };

  /**
   * 获取当前商户ID
   */
  const getCurrentMerchantId = (): number | null => {
    return merchantContext?.merchant_id || null;
  };

  /**
   * 检查是否可以访问指定商户的资源
   */
  const canAccessMerchant = (merchantId: number): boolean => {
    if (!merchantContext) return false;
    return merchantContext.merchant_id === merchantId;
  };

  /**
   * 基于权限的操作权限检查
   */
  const canManageProducts = (): boolean => {
    return hasAnyMerchantPermission([
      MERCHANT_PERMISSIONS.PRODUCT_CREATE,
      MERCHANT_PERMISSIONS.PRODUCT_EDIT,
      MERCHANT_PERMISSIONS.PRODUCT_DELETE
    ]);
  };

  const canViewProducts = (): boolean => {
    return hasMerchantPermission(MERCHANT_PERMISSIONS.PRODUCT_VIEW);
  };

  const canManageOrders = (): boolean => {
    return hasAnyMerchantPermission([
      MERCHANT_PERMISSIONS.ORDER_PROCESS,
      MERCHANT_PERMISSIONS.ORDER_CANCEL
    ]);
  };

  const canViewOrders = (): boolean => {
    return hasMerchantPermission(MERCHANT_PERMISSIONS.ORDER_VIEW);
  };

  const canManageUsers = (): boolean => {
    return hasMerchantPermission(MERCHANT_PERMISSIONS.USER_MANAGE);
  };

  const canViewUsers = (): boolean => {
    return hasMerchantPermission(MERCHANT_PERMISSIONS.USER_VIEW);
  };

  const canViewReports = (): boolean => {
    return hasMerchantPermission(MERCHANT_PERMISSIONS.REPORT_VIEW);
  };

  const canExportReports = (): boolean => {
    return hasMerchantPermission(MERCHANT_PERMISSIONS.REPORT_EXPORT);
  };

  return {
    // 权限上下文
    merchantContext,
    
    // 基础权限检查
    hasMerchantPermission,
    hasAnyMerchantPermission,
    hasAllMerchantPermissions,
    hasMerchantRole,
    
    // 角色检查
    isMerchantAdmin,
    isMerchantOperator,
    
    // 访问控制
    getCurrentMerchantId,
    canAccessMerchant,
    
    // 业务权限检查
    canManageProducts,
    canViewProducts,
    canManageOrders,
    canViewOrders,
    canManageUsers,
    canViewUsers,
    canViewReports,
    canExportReports,
    
    // 状态
    isLoggedIn: isAuthenticated,
    isMerchantUser: !!merchantContext
  };
};
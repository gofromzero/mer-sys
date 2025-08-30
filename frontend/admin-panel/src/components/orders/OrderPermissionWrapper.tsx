import React, { ReactNode } from 'react';
import { useAuthStore } from '../../stores/authStore';

interface OrderPermissionWrapperProps {
  /** 子组件 */
  children: ReactNode;
  /** 所需权限列表 */
  permissions: string[];
  /** 权限检查模式：all（需要全部权限）、any（需要任一权限） */
  mode?: 'all' | 'any';
  /** 无权限时的提示信息 */
  fallback?: ReactNode;
  /** 是否显示权限警告 */
  showPermissionWarning?: boolean;
}

const OrderPermissionWrapper: React.FC<OrderPermissionWrapperProps> = ({
  children,
  permissions,
  mode = 'all',
  fallback = null,
  showPermissionWarning = true,
}) => {
  const { user, hasPermission } = useAuthStore();

  // 检查用户是否有所需权限
  const checkPermissions = (): boolean => {
    if (!user) return false;

    if (mode === 'all') {
      return permissions.every(permission => hasPermission(permission));
    } else {
      return permissions.some(permission => hasPermission(permission));
    }
  };

  const hasRequiredPermissions = checkPermissions();

  // 如果没有权限，显示降级内容
  if (!hasRequiredPermissions) {
    if (fallback) {
      return <>{fallback}</>;
    }

    if (showPermissionWarning) {
      return (
        <div className="permission-denied bg-red-50 border border-red-200 rounded-md p-4 m-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-red-400"
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">权限不足</h3>
              <div className="mt-2 text-sm text-red-700">
                <p>您没有访问此功能的权限。</p>
                <p className="mt-1">
                  所需权限: {permissions.map(p => `"${p}"`).join(mode === 'all' ? ' 和 ' : ' 或 ')}
                </p>
                <p className="mt-1 text-xs">
                  如需访问，请联系管理员为您分配相应权限。
                </p>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return null;
  }

  // 有权限，渲染子组件
  return <>{children}</>;
};

// 订单管理相关权限常量
export const ORDER_PERMISSIONS = {
  // 基本权限
  VIEW: 'order:read',          // 查看订单
  CREATE: 'order:create',      // 创建订单
  UPDATE: 'order:update',      // 更新订单信息
  DELETE: 'order:delete',      // 删除订单
  
  // 状态管理权限
  STATUS_UPDATE: 'order:status:update',    // 更新订单状态
  STATUS_HISTORY: 'order:status:history',  // 查看状态历史
  
  // 批量操作权限
  BATCH_PROCESS: 'order:batch:process',    // 批量处理
  BATCH_CANCEL: 'order:batch:cancel',      // 批量取消
  BATCH_UPDATE: 'order:batch:update',      // 批量更新
  
  // 报表和统计权限
  STATS_VIEW: 'order:stats:view',          // 查看统计
  EXPORT: 'order:export',                  // 导出数据
  REPORT_GENERATE: 'order:report:generate', // 生成报表
  
  // 商户特定权限
  MERCHANT_MANAGE: 'order:merchant:manage', // 商户订单管理
  MERCHANT_VIEW_ALL: 'order:merchant:view_all', // 查看所有商户订单
  
  // 客户特定权限
  CUSTOMER_VIEW_OWN: 'order:customer:view_own', // 客户查看自己的订单
  
  // 支付相关权限
  PAYMENT_INITIATE: 'order:payment:initiate', // 发起支付
  PAYMENT_REFUND: 'order:payment:refund',     // 退款
  PAYMENT_VIEW: 'order:payment:view',         // 查看支付信息
  
  // 通知权限
  NOTIFICATION_SEND: 'order:notification:send',  // 发送通知
  NOTIFICATION_VIEW: 'order:notification:view',  // 查看通知
  
  // 高级权限
  ADMIN_MANAGE: 'order:admin:manage',        // 管理员全权限
  SYSTEM_CONFIG: 'order:system:config',     // 系统配置
} as const;

// 权限组合常量（常用权限组合）
export const ORDER_PERMISSION_GROUPS = {
  // 商户基础权限
  MERCHANT_BASIC: [
    ORDER_PERMISSIONS.VIEW,
    ORDER_PERMISSIONS.STATUS_UPDATE,
    ORDER_PERMISSIONS.STATUS_HISTORY,
    ORDER_PERMISSIONS.MERCHANT_MANAGE,
  ],
  
  // 商户高级权限
  MERCHANT_ADVANCED: [
    ORDER_PERMISSIONS.VIEW,
    ORDER_PERMISSIONS.STATUS_UPDATE,
    ORDER_PERMISSIONS.STATUS_HISTORY,
    ORDER_PERMISSIONS.BATCH_PROCESS,
    ORDER_PERMISSIONS.BATCH_CANCEL,
    ORDER_PERMISSIONS.MERCHANT_MANAGE,
    ORDER_PERMISSIONS.STATS_VIEW,
    ORDER_PERMISSIONS.EXPORT,
  ],
  
  // 客户权限
  CUSTOMER_BASIC: [
    ORDER_PERMISSIONS.VIEW,
    ORDER_PERMISSIONS.CREATE,
    ORDER_PERMISSIONS.CUSTOMER_VIEW_OWN,
    ORDER_PERMISSIONS.PAYMENT_INITIATE,
    ORDER_PERMISSIONS.PAYMENT_VIEW,
  ],
  
  // 管理员权限
  ADMIN_FULL: [
    ORDER_PERMISSIONS.VIEW,
    ORDER_PERMISSIONS.CREATE,
    ORDER_PERMISSIONS.UPDATE,
    ORDER_PERMISSIONS.DELETE,
    ORDER_PERMISSIONS.STATUS_UPDATE,
    ORDER_PERMISSIONS.STATUS_HISTORY,
    ORDER_PERMISSIONS.BATCH_PROCESS,
    ORDER_PERMISSIONS.BATCH_CANCEL,
    ORDER_PERMISSIONS.BATCH_UPDATE,
    ORDER_PERMISSIONS.STATS_VIEW,
    ORDER_PERMISSIONS.EXPORT,
    ORDER_PERMISSIONS.REPORT_GENERATE,
    ORDER_PERMISSIONS.MERCHANT_MANAGE,
    ORDER_PERMISSIONS.MERCHANT_VIEW_ALL,
    ORDER_PERMISSIONS.PAYMENT_REFUND,
    ORDER_PERMISSIONS.NOTIFICATION_SEND,
    ORDER_PERMISSIONS.ADMIN_MANAGE,
    ORDER_PERMISSIONS.SYSTEM_CONFIG,
  ],
} as const;

// 多租户数据隔离检查组件
interface TenantIsolationWrapperProps {
  /** 子组件 */
  children: ReactNode;
  /** 资源所属租户ID */
  resourceTenantId?: number;
  /** 无权限访问时的回退组件 */
  fallback?: ReactNode;
}

export const TenantIsolationWrapper: React.FC<TenantIsolationWrapperProps> = ({
  children,
  resourceTenantId,
  fallback = null,
}) => {
  const { user } = useAuthStore();

  // 检查租户隔离权限
  const checkTenantAccess = (): boolean => {
    if (!user || !resourceTenantId) return false;
    
    // 系统管理员可以访问所有租户数据
    if (user.role === 'system_admin') return true;
    
    // 租户管理员和用户只能访问自己租户的数据
    return user.tenant_id === resourceTenantId;
  };

  const hasTenantAccess = checkTenantAccess();

  if (!hasTenantAccess) {
    if (fallback) {
      return <>{fallback}</>;
    }

    return (
      <div className="tenant-isolation-denied bg-yellow-50 border border-yellow-200 rounded-md p-4 m-4">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg
              className="h-5 w-5 text-yellow-400"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fillRule="evenodd"
                d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                clipRule="evenodd"
              />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-yellow-800">数据访问受限</h3>
            <div className="mt-2 text-sm text-yellow-700">
              <p>您无法访问其他租户的数据。</p>
              <p className="mt-1 text-xs">
                多租户数据隔离机制已生效。
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};

export default OrderPermissionWrapper;
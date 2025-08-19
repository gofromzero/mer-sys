import React, { type ReactNode } from 'react';
import { useTenantPermissions } from '../../hooks/useTenantPermissions';

interface PermissionGuardProps {
  /** 需要的权限 */
  permission?: string;
  /** 需要的权限列表（任意一个即可） */
  anyPermissions?: string[];
  /** 需要的权限列表（必须全部具备） */
  allPermissions?: string[];
  /** 没有权限时显示的内容 */
  fallback?: ReactNode;
  /** 子组件 */
  children: ReactNode;
  /** 是否隐藏无权限内容（默认显示fallback） */
  hideWhenNoPermission?: boolean;
}

/**
 * 权限守护组件
 * 根据用户权限决定是否显示子组件
 * 
 * @description 此组件提供声明式的权限控制，支持多种权限检查模式。
 * 具备完整的可访问性支持和错误边界处理。
 * 
 * @example
 * ```tsx
 * <PermissionGuard permission="tenant:create">
 *   <CreateTenantButton />
 * </PermissionGuard>
 * 
 * <PermissionGuard 
 *   anyPermissions={['tenant:edit', 'tenant:manage_status']}
 *   fallback={<div>权限不足</div>}
 * >
 *   <TenantActions />
 * </PermissionGuard>
 * ```
 */
export const PermissionGuard: React.FC<PermissionGuardProps> = ({
  permission,
  anyPermissions,
  allPermissions,
  fallback = null,
  children,
  hideWhenNoPermission = false
}) => {
  const { hasPermission, hasAnyPermission, hasAllPermissions } = useTenantPermissions();

  // 使用 useMemo 优化权限检查计算
  const hasAccess = React.useMemo(() => {
    try {
      if (permission) {
        return hasPermission(permission);
      } else if (anyPermissions && anyPermissions.length > 0) {
        return hasAnyPermission(anyPermissions);
      } else if (allPermissions && allPermissions.length > 0) {
        return hasAllPermissions(allPermissions);
      } else {
        // 如果没有指定任何权限要求，默认允许访问
        return true;
      }
    } catch (error) {
      // 权限检查出错时，为安全起见默认拒绝访问
      console.warn('PermissionGuard: Error checking permissions', error);
      return false;
    }
  }, [permission, anyPermissions, allPermissions, hasPermission, hasAnyPermission, hasAllPermissions]);

  if (hasAccess) {
    return <>{children}</>;
  }

  if (hideWhenNoPermission) {
    return null;
  }

  // 为无权限状态添加适当的可访问性属性
  if (fallback) {
    return (
      <div 
        role="alert"
        aria-live="polite"
        aria-label="权限不足提示"
      >
        {fallback}
      </div>
    );
  }

  return null;
};

interface ConfirmationDialogProps {
  /** 是否显示对话框 */
  isOpen: boolean;
  /** 操作标题 */
  title: string;
  /** 操作描述 */
  message: string;
  /** 确认按钮文本 */
  confirmText?: string;
  /** 取消按钮文本 */
  cancelText?: string;
  /** 操作类型（影响样式） */
  type?: 'warning' | 'danger' | 'info';
  /** 确认回调 */
  onConfirm: () => void;
  /** 取消回调 */
  onCancel: () => void;
}

/**
 * 二次确认对话框组件
 * 用于敏感操作的确认
 * 
 * @description 提供可访问的模态对话框，支持键盘导航和屏幕阅读器。
 * 包含焦点管理和ESC键关闭功能。
 * 
 * @example
 * ```tsx
 * <ConfirmationDialog
 *   isOpen={showDialog}
 *   title="删除确认"
 *   message="此操作不可逆，确定要删除吗？"
 *   type="danger"
 *   onConfirm={handleDelete}
 *   onCancel={() => setShowDialog(false)}
 * />
 * ```
 */
export const ConfirmationDialog: React.FC<ConfirmationDialogProps> = ({
  isOpen,
  title,
  message,
  confirmText = '确认',
  cancelText = '取消',
  type = 'warning',
  onConfirm,
  onCancel
}) => {
  // 处理 ESC 键关闭对话框
  React.useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && isOpen) {
        onCancel();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      // 阻止背景滚动
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onCancel]);

  if (!isOpen) return null;

  const typeStyles = {
    warning: {
      bgColor: 'bg-yellow-50',
      iconColor: 'text-yellow-600',
      buttonColor: 'bg-yellow-600 hover:bg-yellow-700'
    },
    danger: {
      bgColor: 'bg-red-50',
      iconColor: 'text-red-600',
      buttonColor: 'bg-red-600 hover:bg-red-700'
    },
    info: {
      bgColor: 'bg-blue-50',
      iconColor: 'text-blue-600',
      buttonColor: 'bg-blue-600 hover:bg-blue-700'
    }
  };

  const styles = typeStyles[type];

  // 焦点管理
  const confirmButtonRef = React.useRef<HTMLButtonElement>(null);
  const cancelButtonRef = React.useRef<HTMLButtonElement>(null);

  // 对话框打开时，将焦点设置到确认按钮
  React.useEffect(() => {
    if (isOpen && confirmButtonRef.current) {
      confirmButtonRef.current.focus();
    }
  }, [isOpen]);

  // 处理 Tab 键导航
  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Tab') {
      const isShiftPressed = event.shiftKey;
      const activeElement = document.activeElement;

      if (isShiftPressed) {
        // Shift + Tab - 向后导航
        if (activeElement === confirmButtonRef.current) {
          event.preventDefault();
          cancelButtonRef.current?.focus();
        }
      } else {
        // Tab - 向前导航
        if (activeElement === cancelButtonRef.current) {
          event.preventDefault();
          confirmButtonRef.current?.focus();
        }
      }
    }
  };

  return (
    <div 
      className="fixed inset-0 z-50 overflow-y-auto"
      role="dialog"
      aria-modal="true"
      aria-labelledby="confirmation-dialog-title"
      aria-describedby="confirmation-dialog-description"
      onKeyDown={handleKeyDown}
    >
      <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        {/* 背景遮罩 */}
        <div 
          className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
          onClick={onCancel}
          aria-hidden="true"
        />

        {/* 对话框 */}
        <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div className={`${styles.bgColor} px-4 pt-5 pb-4 sm:p-6 sm:pb-4`}>
            <div className="sm:flex sm:items-start">
              <div className={`mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full ${styles.bgColor} sm:mx-0 sm:h-10 sm:w-10`}>
                {type === 'warning' && (
                  <svg className={`h-6 w-6 ${styles.iconColor}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
                  </svg>
                )}
                {type === 'danger' && (
                  <svg className={`h-6 w-6 ${styles.iconColor}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                )}
                {type === 'info' && (
                  <svg className={`h-6 w-6 ${styles.iconColor}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                )}
              </div>
              <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                <h3 
                  id="confirmation-dialog-title"
                  className="text-lg leading-6 font-medium text-gray-900"
                >
                  {title}
                </h3>
                <div className="mt-2">
                  <p 
                    id="confirmation-dialog-description"
                    className="text-sm text-gray-500"
                  >
                    {message}
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <button
              ref={confirmButtonRef}
              type="button"
              className={`w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 ${styles.buttonColor} text-base font-medium text-white focus:outline-none focus:ring-2 focus:ring-offset-2 sm:ml-3 sm:w-auto sm:text-sm`}
              onClick={onConfirm}
              aria-describedby="confirmation-dialog-description"
            >
              {confirmText}
            </button>
            <button
              ref={cancelButtonRef}
              type="button"
              className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
              onClick={onCancel}
              aria-describedby="confirmation-dialog-description"
            >
              {cancelText}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
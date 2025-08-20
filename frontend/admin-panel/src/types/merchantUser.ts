import type { UserStatus } from './user';

export const MerchantRoleType = {
  MERCHANT_ADMIN: 'merchant_admin',
  MERCHANT_OPERATOR: 'merchant_operator'
} as const;

export type MerchantRoleType = typeof MerchantRoleType[keyof typeof MerchantRoleType];

export const MERCHANT_PERMISSIONS = {
  // 商品管理权限
  PRODUCT_VIEW: 'merchant:product:view',
  PRODUCT_CREATE: 'merchant:product:create',
  PRODUCT_EDIT: 'merchant:product:edit',
  PRODUCT_DELETE: 'merchant:product:delete',
  
  // 订单管理权限
  ORDER_VIEW: 'merchant:order:view',
  ORDER_PROCESS: 'merchant:order:process',
  ORDER_CANCEL: 'merchant:order:cancel',
  
  // 用户管理权限
  USER_VIEW: 'merchant:user:view',
  USER_MANAGE: 'merchant:user:manage',
  
  // 报表权限
  REPORT_VIEW: 'merchant:report:view',
  REPORT_EXPORT: 'merchant:report:export'
} as const;

export interface MerchantUser {
  id: number;
  uuid: string;
  username: string;
  email: string;
  phone?: string;
  tenant_id: number;
  merchant_id: number;
  status: UserStatus;
  role_type: MerchantRoleType;
  permissions: string[];
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface CreateMerchantUserRequest {
  username: string;
  email: string;
  phone?: string;
  merchant_id: number;
  role_type: MerchantRoleType;
  permissions: string[];
  initial_password?: string;
}

export interface UpdateMerchantUserRequest {
  username?: string;
  email?: string;
  phone?: string;
  role_type?: MerchantRoleType;
  permissions?: string[];
}

export interface MerchantUserStatusRequest {
  status: UserStatus;
  comment?: string;
}

export interface MerchantUserPasswordResetRequest {
  new_password?: string;
  send_email?: boolean;
}

export interface MerchantUserAuditLog {
  id: number;
  user_id: number;
  action: string;
  resource: string;
  details: Record<string, any>;
  ip_address: string;
  user_agent: string;
  timestamp: string;
}

export interface MerchantUserQueryParams {
  page?: number;
  page_size?: number;
  merchant_id?: number;
  status?: UserStatus;
  role_type?: MerchantRoleType;
  search?: string;
  username?: string;
  email?: string;
}
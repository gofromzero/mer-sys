export const UserStatus = {
  PENDING: 'pending',
  ACTIVE: 'active',
  SUSPENDED: 'suspended',
  DEACTIVATED: 'deactivated'
} as const;

export type UserStatus = typeof UserStatus[keyof typeof UserStatus];

export interface User {
  id: number;
  uuid: string;
  username: string;
  email: string;
  phone?: string;
  tenant_id: number;
  merchant_id?: number;  // 新增：支持商户用户
  status: UserStatus;
  roles?: UserRole[];    // 新增：用户角色数组
  created_at: string;
  updated_at: string;
  last_login_at?: string;
}

export interface UserRole {
  id: number;
  user_id: number;
  role_type: string;
  resource_id?: number;
  permissions: string[];
}
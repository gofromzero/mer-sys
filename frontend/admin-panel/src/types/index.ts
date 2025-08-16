export * from './user';
export * from './tenant';

import type { User } from './user';

export interface APIResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  permissions: string[];
  isLoading: boolean;
}

export interface UIState {
  sidebarOpen: boolean;
  theme: 'light' | 'dark';
  loading: boolean;
}

// Re-export from user and tenant files
export type { User, UserRole } from './user';
export type { Tenant, TenantConfig } from './tenant';
export { UserStatus } from './user';
export { TenantStatus } from './tenant';
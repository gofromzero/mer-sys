export const TenantStatus = {
  ACTIVE: 'active',
  SUSPENDED: 'suspended',
  EXPIRED: 'expired'
} as const;

export type TenantStatus = typeof TenantStatus[keyof typeof TenantStatus];

export interface Tenant {
  id: number;
  name: string;
  code: string;
  status: TenantStatus;
  config: string; // JSON string - will be parsed on frontend
  created_at: string;
  updated_at: string;
}

export interface TenantConfig {
  max_users: number;
  max_merchants: number;
  features: string[];
  settings: Record<string, string>;
}
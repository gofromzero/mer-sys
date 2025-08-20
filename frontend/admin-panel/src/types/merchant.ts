// 商户状态枚举
export const MerchantStatus = {
  PENDING: 'pending',      // 待审核
  ACTIVE: 'active',        // 已激活
  SUSPENDED: 'suspended',  // 已暂停
  DEACTIVATED: 'deactivated' // 已停用
} as const;

export type MerchantStatus = typeof MerchantStatus[keyof typeof MerchantStatus];

// 商户业务信息
export interface BusinessInfo {
  type: string;          // 商户类型
  category: string;      // 业务分类
  license: string;       // 营业执照号
  legal_name: string;    // 法人名称
  contact_name: string;  // 联系人姓名
  contact_phone: string; // 联系电话
  contact_email: string; // 联系邮箱
  address: string;       // 经营地址
  scope: string;         // 经营范围
  description: string;   // 商户描述
}

// 权益余额信息
export interface RightsBalance {
  total_balance: number;  // 总余额
  used_balance: number;   // 已使用余额
  frozen_balance: number; // 冻结余额
}

// 商户实体
export interface Merchant {
  id: number;
  tenant_id: number;
  name: string;
  code: string;
  status: MerchantStatus;
  business_info: BusinessInfo;
  rights_balance: RightsBalance;
  registration_time?: string; // 注册申请时间
  approval_time?: string;     // 审批时间
  approved_by?: number;       // 审批人ID
  created_at: string;
  updated_at: string;
}

// 商户注册请求
export interface MerchantRegistrationRequest {
  name: string;
  code: string;
  business_info: BusinessInfo;
}

// 商户审批请求
export interface MerchantApprovalRequest {
  action: 'approve' | 'reject';
  comment?: string;
}

// 商户更新请求
export interface MerchantUpdateRequest {
  name?: string;
  business_info?: BusinessInfo;
}

// 商户状态更新请求
export interface MerchantStatusUpdateRequest {
  status: MerchantStatus;
  comment?: string;
}

// 商户列表查询参数
export interface MerchantListQuery {
  page?: number;
  page_size?: number;
  status?: MerchantStatus;
  name?: string;
  search?: string;
}

// 商户列表响应
export interface MerchantListResponse {
  items: Merchant[];
  total: number;
  page: number;
  page_size: number;
}
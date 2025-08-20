// 资金相关TypeScript类型定义

// 资金类型枚举
export const FundType = {
  DEPOSIT: 1,     // 充值
  ALLOCATION: 2,  // 分配
  CONSUMPTION: 3, // 消费
  REFUND: 4       // 退款
} as const;
export type FundType = typeof FundType[keyof typeof FundType];

// 资金状态枚举
export const FundStatus = {
  PENDING: 1,     // 待处理
  CONFIRMED: 2,   // 已确认
  FAILED: 3,      // 失败
  CANCELLED: 4    // 已取消
} as const;
export type FundStatus = typeof FundStatus[keyof typeof FundStatus];

// 交易类型枚举
export const TransactionType = {
  CREDIT: 1,  // 入账
  DEBIT: 2    // 出账
} as const;
export type TransactionType = typeof TransactionType[keyof typeof TransactionType];

// 资金记录接口
export interface Fund {
  id: number;
  tenant_id: number;
  merchant_id: number;
  fund_type: FundType;
  amount: number;
  currency: string;
  status: FundStatus;
  created_at: string;
  updated_at: string;
}

// 资金流转记录接口
export interface FundTransaction {
  id: number;
  tenant_id: number;
  merchant_id: number;
  fund_id: number;
  transaction_type: TransactionType;
  amount: number;
  balance_before: number;
  balance_after: number;
  operator_id: number;
  description?: string;
  created_at: string;
}

// 权益余额接口
export interface RightsBalance {
  total_balance: number;      // 总余额
  used_balance: number;       // 已使用余额
  frozen_balance: number;     // 冻结余额
  available_balance: number;  // 可用余额
  last_updated: string;       // 最后更新时间
}

// 单笔充值请求
export interface DepositRequest {
  merchant_id: number;
  amount: number;
  currency: string;
  description?: string;
}

// 批量充值请求
export interface BatchDepositRequest {
  deposits: DepositRequest[];
}

// 权益分配请求
export interface AllocateRequest {
  merchant_id: number;
  amount: number;
  description?: string;
}

// 资金流转查询参数
export interface FundTransactionQuery {
  tenant_id?: number;
  merchant_id?: number;
  fund_id?: number;
  transaction_type?: TransactionType;
  operator_id?: number;
  start_time?: string;
  end_time?: string;
  page: number;
  page_size: number;
}

// 资金概览统计
export interface FundSummary {
  total_deposits: number;    // 充值总额
  total_allocations: number; // 分配总额
  total_consumption: number; // 消费总额
  total_refunds: number;     // 退款总额
  available_balance: number; // 可用余额
}

// 冻结/解冻请求
export interface FreezeRequest {
  action: 'freeze' | 'unfreeze';
  amount: number;
  reason?: string;
}

// API响应通用结构
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

// 分页响应结构
export interface PaginationResponse<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

// 资金类型显示映射
export const fundTypeLabels: Record<FundType, string> = {
  [FundType.DEPOSIT]: '充值',
  [FundType.ALLOCATION]: '分配',
  [FundType.CONSUMPTION]: '消费',
  [FundType.REFUND]: '退款'
};

// 资金状态显示映射
export const fundStatusLabels: Record<FundStatus, string> = {
  [FundStatus.PENDING]: '待处理',
  [FundStatus.CONFIRMED]: '已确认',
  [FundStatus.FAILED]: '失败',
  [FundStatus.CANCELLED]: '已取消'
};

// 交易类型显示映射
export const transactionTypeLabels: Record<TransactionType, string> = {
  [TransactionType.CREDIT]: '入账',
  [TransactionType.DEBIT]: '出账'
};

// 资金状态颜色映射
export const fundStatusColors: Record<FundStatus, string> = {
  [FundStatus.PENDING]: 'orange',
  [FundStatus.CONFIRMED]: 'green',
  [FundStatus.FAILED]: 'red',
  [FundStatus.CANCELLED]: 'gray'
};

// 交易类型颜色映射
export const transactionTypeColors: Record<TransactionType, string> = {
  [TransactionType.CREDIT]: 'green',
  [TransactionType.DEBIT]: 'red'
};
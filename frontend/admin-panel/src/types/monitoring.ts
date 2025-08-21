// 监控相关类型定义
export enum AlertType {
  BALANCE_LOW = 'balance_low',
  BALANCE_CRITICAL = 'balance_critical',
  USAGE_SPIKE = 'usage_spike',
  PREDICTED_DEPLETION = 'predicted_depletion'
}

export enum AlertSeverity {
  INFO = 'info',
  WARNING = 'warning',
  CRITICAL = 'critical'
}

export enum AlertStatus {
  ACTIVE = 'active',
  RESOLVED = 'resolved',
  ACKNOWLEDGED = 'acknowledged'
}

export enum TrendDirection {
  INCREASING = 'increasing',
  DECREASING = 'decreasing',
  STABLE = 'stable'
}

export enum TimePeriod {
  DAILY = 'daily',
  WEEKLY = 'weekly',
  MONTHLY = 'monthly'
}

export interface RightsAlert {
  id: number;
  tenant_id: number;
  merchant_id: number;
  alert_type: AlertType;
  threshold_value: number;
  current_value: number;
  severity: AlertSeverity;
  status: AlertStatus;
  message: string;
  triggered_at: string;
  resolved_at?: string;
  notified_channels: string[];
  created_at: string;
  updated_at: string;
}

export interface RightsUsageStats {
  tenant_id?: number;
  merchant_id?: number;
  stat_date: string;
  period: TimePeriod;
  total_allocated: number;
  total_consumed: number;
  average_daily_usage: number;
  peak_usage_day?: string;
  predicted_depletion_date?: string;
  usage_trend: TrendDirection;
  created_at: string;
}

export interface MonitoringDashboardData {
  total_merchants: number;
  active_alerts: number;
  total_rights_balance: number;
  daily_consumption: number;
  recent_alerts: RightsAlert[];
  usage_trends: RightsUsageStats[];
  consumption_chart_data: ChartDataPoint[];
  balance_distribution: BalanceDistribution[];
}

export interface ChartDataPoint {
  date: string;
  consumed: number;
  allocated: number;
  trend: TrendDirection;
}

export interface BalanceDistribution {
  merchant_name: string;
  merchant_id: number;
  available_balance: number;
  usage_percentage: number;
  status: 'healthy' | 'warning' | 'critical';
}

// API 查询参数类型
export interface RightsStatsQuery {
  merchant_id?: number;
  period?: TimePeriod;
  start_date?: string;
  end_date?: string;
  page?: number;
  page_size?: number;
}

export interface RightsTrendsQuery {
  merchant_id?: number;
  days?: number;
  period?: TimePeriod;
}

export interface AlertListQuery {
  page: number;
  page_size: number;
  merchant_id?: number;
  alert_type?: AlertType;
  severity?: AlertSeverity;
  status?: AlertStatus;
  start_date?: string;
  end_date?: string;
}

export interface AlertConfigureRequest {
  merchant_id: number;
  warning_threshold?: number;
  critical_threshold?: number;
}

export interface AlertResolveRequest {
  resolution: string;
}

export interface ReportGenerateRequest {
  period: TimePeriod;
  start_date: string;
  end_date: string;
  merchant_ids: number[];
  format: 'pdf' | 'excel' | 'csv';
}

// API 响应类型
export interface ApiResponse<T> {
  code: number;
  message?: string;
  data?: T;
}

export interface PaginatedResponse<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}
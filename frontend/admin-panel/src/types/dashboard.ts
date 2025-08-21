// 商户仪表板相关类型定义

// 时间周期
export enum TimePeriod {
  DAILY = 'daily',
  WEEKLY = 'weekly',  
  MONTHLY = 'monthly'
}

// 趋势方向
export enum TrendDirection {
  UP = 'up',
  DOWN = 'down',
  FLAT = 'flat'
}

// 优先级
export enum Priority {
  LOW = 'low',
  NORMAL = 'normal',
  HIGH = 'high',
  URGENT = 'urgent'
}

// 任务类型
export enum TaskType {
  ORDER_PROCESSING = 'order_processing',
  VERIFICATION_PENDING = 'verification_pending',
  LOW_BALANCE_WARNING = 'low_balance_warning',
  PRODUCT_UPDATE_NEEDED = 'product_update_needed'
}

// 预警类型
export enum AlertType {
  BALANCE_LOW = 'balance_low',
  BALANCE_CRITICAL = 'balance_critical', 
  USAGE_SPIKE = 'usage_spike',
  PREDICTED_DEPLETION = 'predicted_depletion'
}

// 预警严重程度
export enum AlertSeverity {
  INFO = 'info',
  WARNING = 'warning',
  CRITICAL = 'critical'
}

// 预警状态
export enum AlertStatus {
  ACTIVE = 'active',
  RESOLVED = 'resolved',
  IGNORED = 'ignored'
}

// 组件类型
export enum WidgetType {
  SALES_OVERVIEW = 'sales_overview',
  RIGHTS_BALANCE = 'rights_balance',
  RIGHTS_TREND = 'rights_trend', 
  PENDING_TASKS = 'pending_tasks',
  RECENT_ORDERS = 'recent_orders',
  ANNOUNCEMENTS = 'announcements',
  QUICK_ACTIONS = 'quick_actions'
}

// 权益余额信息
export interface RightsBalance {
  total_balance: number;
  used_balance: number;
  frozen_balance: number;
  available_balance: number;
  last_updated: string;
  warning_threshold?: number;
  critical_threshold?: number;
  trend_coefficient?: number;
}

// 权益使用趋势点
export interface RightsUsagePoint {
  date: string;
  balance: number;
  usage: number;
  trend: TrendDirection;
}

// 权益预警
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
}

// 待处理任务
export interface PendingTask {
  id: string;
  type: TaskType;
  description: string;
  priority: Priority;
  due_date?: string;
  count: number;
}

// 公告信息
export interface Announcement {
  id: number;
  title: string;
  content: string;
  priority: Priority;
  publish_date: string;
  expire_date?: string;
  read_status: boolean;
}

// 通知信息
export interface Notification {
  id: number;
  title: string;
  content: string;
  type: string;
  priority: Priority;
  read_at?: string;
  created_at: string;
}

// 商户仪表板数据
export interface MerchantDashboardData {
  merchant_id: number;
  tenant_id: number;
  period: TimePeriod;
  
  // 核心业务指标
  total_sales: number;
  total_orders: number; 
  total_customers: number;
  rights_balance: RightsBalance;
  
  // 权益使用情况
  rights_usage_trend: RightsUsagePoint[];
  rights_alerts: RightsAlert[];
  predicted_depletion_days?: number;
  
  // 待处理事项
  pending_orders: number;
  pending_verifications: number;
  pending_tasks: PendingTask[];
  
  // 系统通知
  announcements: Announcement[];
  notifications: Notification[];
  
  last_updated: string;
}

// 组件位置
export interface Position {
  x: number;
  y: number;
}

// 组件大小
export interface Size {
  width: number;
  height: number;
}

// 仪表板组件
export interface DashboardWidget {
  id: string;
  type: WidgetType;
  position: Position;
  size: Size;
  config: Record<string, any>;
  visible: boolean;
}

// 组件偏好设置
export interface WidgetPreference {
  widget_type: WidgetType;
  enabled: boolean;
  config: Record<string, any>;
}

// 布局配置
export interface LayoutConfig {
  columns: number;
  widgets: DashboardWidget[];
}

// 移动端布局配置
export interface MobileLayoutConfig {
  columns: number;
  widgets: DashboardWidget[];
}

// 仪表板配置
export interface DashboardConfig {
  merchant_id: number;
  layout_config: LayoutConfig;
  widget_preferences: WidgetPreference[];
  refresh_interval: number;
  mobile_layout: MobileLayoutConfig;
}

// 仪表板配置请求
export interface DashboardConfigRequest {
  layout_config: LayoutConfig;
  widget_preferences?: WidgetPreference[];
  refresh_interval: number;
  mobile_layout?: MobileLayoutConfig;
}

// 通知响应
export interface NotificationsResponse {
  notifications: Notification[];
  announcements: Announcement[];
  unread_count: number;
}

// API响应接口
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
  error?: string;
}

// 仪表板API接口
export interface DashboardStats {
  period: TimePeriod;
  total_sales: number;
  total_orders: number;
  total_customers: number;
  growth_rate?: number;
}

// 权益趋势查询参数
export interface RightsTrendParams {
  days?: number;
}

// 仪表板查询参数
export interface DashboardParams {
  period?: TimePeriod;
  refresh?: boolean;
}

// 组件配置选项
export interface WidgetConfig {
  title?: string;
  showHeader?: boolean;
  refreshInterval?: number;
  chartType?: 'line' | 'bar' | 'pie' | 'area';
  timeRange?: number;
  displayMode?: 'card' | 'table' | 'chart';
  colors?: string[];
  showLegend?: boolean;
  animate?: boolean;
}

// 仪表板主题配置
export interface DashboardTheme {
  primaryColor: string;
  backgroundColor: string;
  cardBackground: string;
  textColor: string;
  borderColor: string;
  chartColors: string[];
}

// 错误类型
export interface DashboardError {
  code: string;
  message: string;
  details?: any;
}

// 加载状态
export interface LoadingState {
  dashboard: boolean;
  stats: boolean;
  trends: boolean;
  tasks: boolean;
  notifications: boolean;
  config: boolean;
}

// 刷新配置
export interface RefreshConfig {
  interval: number;
  autoRefresh: boolean;
  lastRefresh: number;
}

// 导出类型别名
export type DashboardWidgetId = string;
export type DashboardPeriod = TimePeriod;
export type DashboardThemeType = 'light' | 'dark' | 'auto';
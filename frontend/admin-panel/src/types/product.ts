// 商品相关TypeScript类型定义

export type ProductStatus = 'draft' | 'active' | 'inactive' | 'deleted';

export type CategoryStatus = 1 | 0; // 1: 启用, 0: 禁用

export type ChangeOperation = 'create' | 'update' | 'delete' | 'status_change';

export interface Money {
  amount: number; // 以分为单位
  currency: string;
}

export interface InventoryInfo {
  stock_quantity: number;
  reserved_quantity: number;
  track_inventory: boolean;
  low_stock_threshold?: number;
  reorder_point?: number;
  reorder_quantity?: number;
  cost_per_unit?: Money;
}

export interface ProductImage {
  id: string;
  url: string;
  alt_text?: string;
  sort_order: number;
  is_primary: boolean;
}

export interface Product {
  id: number;
  tenant_id: number;
  merchant_id: number;
  name: string;
  description?: string;
  category_id?: number;
  category_path?: string;
  tags: string[];
  price: Money;
  rights_cost: number;
  inventory: InventoryInfo;
  status: ProductStatus;
  images: ProductImage[];
  version: number;
  created_at: string;
  updated_at: string;
}

export interface ProductCategory {
  id: number;
  tenant_id: number;
  name: string;
  parent_id?: number;
  level: number;
  path: string;
  sort_order: number;
  status: CategoryStatus;
  created_at: string;
  updated_at: string;
}

export interface ProductHistory {
  id: number;
  tenant_id: number;
  product_id: number;
  version: number;
  field_name: string;
  old_value?: string;
  new_value?: string;
  operation: ChangeOperation;
  changed_by: number;
  changed_at: string;
}

// API请求类型
export interface CreateProductRequest {
  name: string;
  description?: string;
  category_id?: number;
  tags?: string[];
  price: Money;
  rights_cost: number;
  inventory: InventoryInfo;
}

export interface UpdateProductRequest {
  name?: string;
  description?: string;
  category_id?: number;
  tags?: string[];
  price?: Money;
  rights_cost?: number;
  inventory?: InventoryInfo;
}

export interface UpdateProductStatusRequest {
  status: ProductStatus;
}

export interface ProductListRequest {
  page: number;
  page_size: number;
  category_id?: number;
  status?: ProductStatus;
  keyword?: string;
  sort_by?: 'created_at' | 'updated_at' | 'name' | 'price';
  sort_order?: 'asc' | 'desc';
}

export interface ProductBatchOperationRequest {
  product_ids: number[];
  operation: 'activate' | 'deactivate' | 'delete';
}

export interface CreateCategoryRequest {
  name: string;
  parent_id?: number;
  sort_order?: number;
}

export interface UpdateCategoryRequest {
  name?: string;
  parent_id?: number;
  sort_order?: number;
  status?: CategoryStatus;
}

export interface UploadImageRequest {
  alt_text?: string;
  sort_order?: number;
  is_primary?: boolean;
}

// API响应类型
export interface ProductResponse {
  product: Product;
  category?: ProductCategory;
}

export interface ProductListResponse {
  products: ProductResponse[];
  total: number;
  page: number;
  page_size: number;
}

export interface CategoryTreeResponse extends ProductCategory {
  children?: CategoryTreeResponse[];
}

// Amis Schema相关类型
export interface AmisSchema {
  type: string;
  [key: string]: any;
}

export interface ProductListSchema extends AmisSchema {
  type: 'crud';
  api: string;
  columns: Array<{
    name: string;
    label: string;
    type?: string;
    [key: string]: any;
  }>;
}

export interface ProductFormSchema extends AmisSchema {
  type: 'form';
  api: string;
  controls: Array<{
    type: string;
    name: string;
    label: string;
    [key: string]: any;
  }>;
}

// 库存管理相关类型定义
export type InventoryChangeType = 'purchase' | 'sale' | 'adjustment' | 'transfer' | 'damage' | 'reservation' | 'release';
export type InventoryAlertType = 'low_stock' | 'out_of_stock' | 'overstock';
export type ReservationStatus = 'active' | 'confirmed' | 'released' | 'expired';

export interface InventoryRecord {
  id: number;
  tenant_id: number;
  product_id: number;
  change_type: InventoryChangeType;
  quantity_before: number;
  quantity_after: number;
  quantity_changed: number;
  reason: string;
  reference_id?: string;
  operated_by: number;
  created_at: string;
}

export interface InventoryReservation {
  id: number;
  tenant_id: number;
  product_id: number;
  reserved_quantity: number;
  reference_type: string;
  reference_id: string;
  status: ReservationStatus;
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

export interface InventoryAlert {
  id: number;
  tenant_id: number;
  product_id: number;
  alert_type: InventoryAlertType;
  threshold_value: number;
  notification_channels: string[];
  is_active: boolean;
  last_triggered_at?: string;
  created_at: string;
  updated_at: string;
}

export interface InventoryStatistics {
  tenant_id: number;
  total_products: number;
  low_stock_products: number;
  out_of_stock_products: number;
  total_inventory_value: number;
  active_alerts: number;
  today_changes: number;
  last_updated: string;
}

// 库存操作请求类型
export interface InventoryAdjustRequest {
  product_id: number;
  adjustment_type: 'increase' | 'decrease' | 'set';
  quantity: number;
  reason: string;
  reference_id?: string;
}

export interface BatchInventoryAdjustRequest {
  adjustments: InventoryAdjustRequest[];
  reason: string;
}

export interface InventoryReserveRequest {
  product_id: number;
  quantity: number;
  reference_type: string;
  reference_id: string;
  expires_at?: string;
}

export interface InventoryAlertRequest {
  product_id: number;
  alert_type: InventoryAlertType;
  threshold_value: number;
  notification_channels: string[];
  is_active: boolean;
}

// 库存响应类型
export interface InventoryResponse {
  product_id: number;
  inventory_info: InventoryInfo;
  available_stock: number;
  reserved_stock: number;
  is_low_stock: boolean;
  is_out_of_stock: boolean;
}

export interface InventoryRecordResponse {
  records: InventoryRecord[];
  total: number;
  page: number;
  page_size: number;
}

export interface InventoryMonitoringData {
  statistics: InventoryStatistics;
  recent_changes: InventoryRecord[];
  active_alerts: InventoryAlert[];
  last_updated: string;
}

// 库存盘点相关类型定义
export type StocktakingStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled';

export interface InventoryStocktaking {
  id: number;
  tenant_id: number;
  name: string;
  description: string;
  status: StocktakingStatus;
  product_ids: number[]; // 如果为空则盘点所有商品
  started_by: number;
  completed_by?: number;
  started_at: string;
  completed_at?: string;
  summary: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface InventoryStocktakingRecord {
  id: number;
  tenant_id: number;
  stocktaking_id: number;
  product_id: number;
  system_count: number;    // 系统库存数量
  actual_count: number;    // 实际盘点数量
  difference: number;      // 差异数量
  reason: string;          // 差异原因
  checked_by: number;      // 盘点人
  checked_at: string;      // 盘点时间
  created_at: string;
  updated_at: string;
}

// 盘点相关请求类型
export interface StocktakingCreateRequest {
  name: string;
  description?: string;
  product_ids?: number[]; // 如果为空则盘点所有商品
  start_time?: string;    // 盘点开始时间，为空则立即开始
}

export interface StocktakingRecordRequest {
  product_id: number;
  actual_count: number;
  system_count: number;
  reason?: string;
}

export interface StocktakingBatchUpdateRequest {
  records: StocktakingRecordRequest[];
}

export interface StocktakingCompleteRequest {
  summary?: string;
  notes?: string;
}

// 盘点响应类型
export interface StocktakingResponse {
  stocktaking: InventoryStocktaking;
  records?: InventoryStocktakingRecord[];
  statistics?: StocktakingStatistics;
}

export interface StocktakingStatistics {
  total_products: number;   // 总盘点商品数
  checked_products: number; // 已盘点商品数
  difference_count: number; // 有差异的商品数
  total_difference: number; // 总差异数量
}

export interface StocktakingListResponse {
  stocktakings: InventoryStocktaking[];
  total: number;
  page: number;
  page_size: number;
}
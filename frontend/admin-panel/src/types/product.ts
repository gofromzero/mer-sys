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
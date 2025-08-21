// 商品API服务

import {
  Product,
  ProductCategory,
  ProductHistory,
  CreateProductRequest,
  UpdateProductRequest,
  UpdateProductStatusRequest,
  ProductListRequest,
  ProductListResponse,
  ProductBatchOperationRequest,
  CreateCategoryRequest,
  UpdateCategoryRequest,
  CategoryTreeResponse,
  UploadImageRequest,
} from '../types/product';

// API基础配置
const API_BASE_URL = '/api/v1';
const PRODUCT_API = `${API_BASE_URL}/products`;
const CATEGORY_API = `${API_BASE_URL}/categories`;

// HTTP客户端封装
class ApiClient {
  private async request<T>(
    url: string,
    options: RequestInit = {}
  ): Promise<{ code: number; message: string; data: T; error?: string }> {
    const defaultHeaders = {
      'Content-Type': 'application/json',
      'X-Tenant-ID': localStorage.getItem('tenant_id') || '',
      Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
    };

    const response = await fetch(url, {
      headers: { ...defaultHeaders, ...options.headers },
      ...options,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }

  async get<T>(url: string): Promise<{ code: number; message: string; data: T; error?: string }> {
    return this.request<T>(url, { method: 'GET' });
  }

  async post<T>(
    url: string,
    data?: any
  ): Promise<{ code: number; message: string; data: T; error?: string }> {
    return this.request<T>(url, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async put<T>(
    url: string,
    data?: any
  ): Promise<{ code: number; message: string; data: T; error?: string }> {
    return this.request<T>(url, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async patch<T>(
    url: string,
    data?: any
  ): Promise<{ code: number; message: string; data: T; error?: string }> {
    return this.request<T>(url, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async delete<T>(url: string): Promise<{ code: number; message: string; data: T; error?: string }> {
    return this.request<T>(url, { method: 'DELETE' });
  }

  async upload<T>(
    url: string,
    formData: FormData
  ): Promise<{ code: number; message: string; data: T; error?: string }> {
    const defaultHeaders = {
      'X-Tenant-ID': localStorage.getItem('tenant_id') || '',
      Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
    };

    const response = await fetch(url, {
      method: 'POST',
      headers: defaultHeaders,
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }
}

const apiClient = new ApiClient();

// 商品服务
export class ProductService {
  // 创建商品
  static async createProduct(data: CreateProductRequest): Promise<Product> {
    const response = await apiClient.post<Product>(PRODUCT_API, data);
    if (response.code !== 200) {
      throw new Error(response.message || '创建商品失败');
    }
    return response.data;
  }

  // 获取商品详情
  static async getProduct(id: number): Promise<Product> {
    const response = await apiClient.get<Product>(`${PRODUCT_API}/${id}`);
    if (response.code !== 200) {
      throw new Error(response.message || '获取商品失败');
    }
    return response.data;
  }

  // 更新商品
  static async updateProduct(id: number, data: UpdateProductRequest): Promise<Product> {
    const response = await apiClient.put<Product>(`${PRODUCT_API}/${id}`, data);
    if (response.code !== 200) {
      throw new Error(response.message || '更新商品失败');
    }
    return response.data;
  }

  // 更新商品状态
  static async updateProductStatus(id: number, data: UpdateProductStatusRequest): Promise<void> {
    const response = await apiClient.patch<null>(`${PRODUCT_API}/${id}/status`, data);
    if (response.code !== 200) {
      throw new Error(response.message || '更新商品状态失败');
    }
  }

  // 删除商品
  static async deleteProduct(id: number): Promise<void> {
    const response = await apiClient.delete<null>(`${PRODUCT_API}/${id}`);
    if (response.code !== 200) {
      throw new Error(response.message || '删除商品失败');
    }
  }

  // 获取商品列表
  static async listProducts(params: ProductListRequest): Promise<ProductListResponse> {
    const query = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        query.append(key, String(value));
      }
    });

    const response = await apiClient.get<ProductListResponse>(`${PRODUCT_API}?${query}`);
    if (response.code !== 200) {
      throw new Error(response.message || '获取商品列表失败');
    }
    return response.data;
  }

  // 批量操作商品
  static async batchOperation(data: ProductBatchOperationRequest): Promise<void> {
    const response = await apiClient.post<null>(`${PRODUCT_API}/batch`, data);
    if (response.code !== 200) {
      throw new Error(response.message || '批量操作失败');
    }
  }

  // 上传商品图片
  static async uploadImage(
    id: number,
    file: File,
    metadata: UploadImageRequest = {}
  ): Promise<{ url: string }> {
    const formData = new FormData();
    formData.append('image', file);
    
    if (metadata.alt_text) formData.append('alt_text', metadata.alt_text);
    if (metadata.sort_order !== undefined) formData.append('sort_order', String(metadata.sort_order));
    if (metadata.is_primary !== undefined) formData.append('is_primary', String(metadata.is_primary));

    const response = await apiClient.upload<{ url: string }>(`${PRODUCT_API}/${id}/images`, formData);
    if (response.code !== 200) {
      throw new Error(response.message || '上传图片失败');
    }
    return response.data;
  }

  // 获取商品变更历史
  static async getProductHistory(id: number): Promise<ProductHistory[]> {
    const response = await apiClient.get<ProductHistory[]>(`${PRODUCT_API}/${id}/history`);
    if (response.code !== 200) {
      throw new Error(response.message || '获取变更历史失败');
    }
    return response.data;
  }
}

// 分类服务
export class CategoryService {
  // 创建分类
  static async createCategory(data: CreateCategoryRequest): Promise<ProductCategory> {
    const response = await apiClient.post<ProductCategory>(CATEGORY_API, data);
    if (response.code !== 200) {
      throw new Error(response.message || '创建分类失败');
    }
    return response.data;
  }

  // 获取分类详情
  static async getCategory(id: number): Promise<ProductCategory> {
    const response = await apiClient.get<ProductCategory>(`${CATEGORY_API}/${id}`);
    if (response.code !== 200) {
      throw new Error(response.message || '获取分类失败');
    }
    return response.data;
  }

  // 更新分类
  static async updateCategory(id: number, data: UpdateCategoryRequest): Promise<ProductCategory> {
    const response = await apiClient.put<ProductCategory>(`${CATEGORY_API}/${id}`, data);
    if (response.code !== 200) {
      throw new Error(response.message || '更新分类失败');
    }
    return response.data;
  }

  // 删除分类
  static async deleteCategory(id: number): Promise<void> {
    const response = await apiClient.delete<null>(`${CATEGORY_API}/${id}`);
    if (response.code !== 200) {
      throw new Error(response.message || '删除分类失败');
    }
  }

  // 获取分类树
  static async getCategoryTree(): Promise<CategoryTreeResponse[]> {
    const response = await apiClient.get<CategoryTreeResponse[]>(`${CATEGORY_API}/tree`);
    if (response.code !== 200) {
      throw new Error(response.message || '获取分类树失败');
    }
    return response.data;
  }

  // 获取分类列表
  static async getCategoryList(): Promise<ProductCategory[]> {
    const response = await apiClient.get<ProductCategory[]>(CATEGORY_API);
    if (response.code !== 200) {
      throw new Error(response.message || '获取分类列表失败');
    }
    return response.data;
  }

  // 获取子分类
  static async getCategoryChildren(id: number): Promise<ProductCategory[]> {
    const response = await apiClient.get<ProductCategory[]>(`${CATEGORY_API}/${id}/children`);
    if (response.code !== 200) {
      throw new Error(response.message || '获取子分类失败');
    }
    return response.data;
  }

  // 获取分类路径
  static async getCategoryPath(id: number): Promise<ProductCategory[]> {
    const response = await apiClient.get<ProductCategory[]>(`${CATEGORY_API}/${id}/path`);
    if (response.code !== 200) {
      throw new Error(response.message || '获取分类路径失败');
    }
    return response.data;
  }
}
// 商品状态管理
import { create } from 'zustand';
import type {
  Product,
  ProductCategory,
  ProductListResponse,
  CreateProductRequest,
  UpdateProductRequest,
  ProductListRequest,
  CategoryTreeResponse
} from '../types/product';
import { ProductService, CategoryService } from '../services/productService';

interface ProductState {
  // 商品相关状态
  products: Product[];
  currentProduct: Product | null;
  productListData: ProductListResponse | null;
  loading: boolean;
  error: string | null;

  // 分类相关状态
  categories: ProductCategory[];
  categoryTree: CategoryTreeResponse[];
  currentCategory: ProductCategory | null;
  categoryLoading: boolean;
  categoryError: string | null;

  // 筛选和分页状态
  filters: Partial<ProductListRequest>;
  currentPage: number;
  pageSize: number;
  sortBy: string;
  sortOrder: 'asc' | 'desc';
}

interface ProductActions {
  // 商品操作
  setProducts: (products: Product[]) => void;
  setCurrentProduct: (product: Product | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;

  // 分类操作
  setCategories: (categories: ProductCategory[]) => void;
  setCategoryTree: (tree: CategoryTreeResponse[]) => void;
  setCurrentCategory: (category: ProductCategory | null) => void;
  setCategoryLoading: (loading: boolean) => void;
  setCategoryError: (error: string | null) => void;
  clearCategoryError: () => void;

  // 筛选和分页
  setFilters: (filters: Partial<ProductListRequest>) => void;
  setCurrentPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setSorting: (field: string, order: 'asc' | 'desc') => void;

  // API 调用
  fetchProducts: (params?: Partial<ProductListRequest>) => Promise<void>;
  fetchProduct: (id: number) => Promise<void>;
  createProduct: (data: CreateProductRequest) => Promise<Product>;
  updateProduct: (id: number, data: UpdateProductRequest) => Promise<Product>;
  deleteProduct: (id: number) => Promise<void>;
  updateProductStatus: (id: number, status: string) => Promise<void>;
  batchOperation: (productIds: number[], operation: string) => Promise<void>;

  // 分类 API 调用
  fetchCategories: () => Promise<void>;
  fetchCategoryTree: () => Promise<void>;
  fetchCategory: (id: number) => Promise<void>;
  createCategory: (data: any) => Promise<ProductCategory>;
  updateCategory: (id: number, data: any) => Promise<ProductCategory>;
  deleteCategory: (id: number) => Promise<void>;

  // 图片上传
  uploadImage: (productId: number, file: File, metadata?: any) => Promise<{ url: string }>;
  
  // 变更历史
  fetchProductHistory: (productId: number) => Promise<any[]>;
}

type ProductStore = ProductState & ProductActions;

export const useProductStore = create<ProductStore>((set, get) => ({
  // 初始状态
  products: [],
  currentProduct: null,
  productListData: null,
  loading: false,
  error: null,

  categories: [],
  categoryTree: [],
  currentCategory: null,
  categoryLoading: false,
  categoryError: null,

  filters: {},
  currentPage: 1,
  pageSize: 20,
  sortBy: 'created_at',
  sortOrder: 'desc',

  // 基础状态操作
  setProducts: (products) => set({ products }),
  setCurrentProduct: (product) => set({ currentProduct: product }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error }),
  clearError: () => set({ error: null }),

  setCategories: (categories) => set({ categories }),
  setCategoryTree: (tree) => set({ categoryTree: tree }),
  setCurrentCategory: (category) => set({ currentCategory: category }),
  setCategoryLoading: (loading) => set({ categoryLoading: loading }),
  setCategoryError: (error) => set({ categoryError: error }),
  clearCategoryError: () => set({ categoryError: null }),

  setFilters: (filters) => set({ filters: { ...get().filters, ...filters } }),
  setCurrentPage: (page) => set({ currentPage: page }),
  setPageSize: (size) => set({ pageSize: size }),
  setSorting: (field, order) => set({ sortBy: field, sortOrder: order }),

  // 商品 API 操作
  fetchProducts: async (params) => {
    const { currentPage, pageSize, sortBy, sortOrder, filters } = get();
    set({ loading: true, error: null });

    try {
      const requestParams: ProductListRequest = {
        page: currentPage,
        page_size: pageSize,
        sort_by: sortBy as any,
        sort_order: sortOrder,
        ...filters,
        ...params
      };

      const data = await ProductService.listProducts(requestParams);
      set({ 
        productListData: data,
        products: data.products.map(p => p.product),
        loading: false 
      });
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '获取商品列表失败',
        loading: false 
      });
    }
  },

  fetchProduct: async (id) => {
    set({ loading: true, error: null });
    try {
      const product = await ProductService.getProduct(id);
      set({ currentProduct: product, loading: false });
    } catch (error) {
      set({ 
        error: error instanceof Error ? error.message : '获取商品详情失败',
        loading: false 
      });
    }
  },

  createProduct: async (data) => {
    set({ loading: true, error: null });
    try {
      const product = await ProductService.createProduct(data);
      const { products } = get();
      set({ 
        products: [product, ...products],
        loading: false 
      });
      return product;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '创建商品失败';
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  updateProduct: async (id, data) => {
    set({ loading: true, error: null });
    try {
      const product = await ProductService.updateProduct(id, data);
      const { products } = get();
      const updatedProducts = products.map(p => p.id === id ? product : p);
      set({ 
        products: updatedProducts,
        currentProduct: product,
        loading: false 
      });
      return product;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '更新商品失败';
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  deleteProduct: async (id) => {
    set({ loading: true, error: null });
    try {
      await ProductService.deleteProduct(id);
      const { products } = get();
      const filteredProducts = products.filter(p => p.id !== id);
      set({ 
        products: filteredProducts,
        loading: false 
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '删除商品失败';
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  updateProductStatus: async (id, status) => {
    set({ loading: true, error: null });
    try {
      await ProductService.updateProductStatus(id, { status: status as any });
      const { products } = get();
      const updatedProducts = products.map(p => 
        p.id === id ? { ...p, status: status as any } : p
      );
      set({ 
        products: updatedProducts,
        loading: false 
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '更新商品状态失败';
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  batchOperation: async (productIds, operation) => {
    set({ loading: true, error: null });
    try {
      await ProductService.batchOperation({ 
        product_ids: productIds,
        operation: operation as any
      });
      
      // 重新获取商品列表
      await get().fetchProducts();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '批量操作失败';
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  // 分类 API 操作
  fetchCategories: async () => {
    set({ categoryLoading: true, categoryError: null });
    try {
      const categories = await CategoryService.getCategoryList();
      set({ categories, categoryLoading: false });
    } catch (error) {
      set({
        categoryError: error instanceof Error ? error.message : '获取分类列表失败',
        categoryLoading: false
      });
    }
  },

  fetchCategoryTree: async () => {
    set({ categoryLoading: true, categoryError: null });
    try {
      const tree = await CategoryService.getCategoryTree();
      set({ categoryTree: tree, categoryLoading: false });
    } catch (error) {
      set({
        categoryError: error instanceof Error ? error.message : '获取分类树失败',
        categoryLoading: false
      });
    }
  },

  fetchCategory: async (id) => {
    set({ categoryLoading: true, categoryError: null });
    try {
      const category = await CategoryService.getCategory(id);
      set({ currentCategory: category, categoryLoading: false });
    } catch (error) {
      set({
        categoryError: error instanceof Error ? error.message : '获取分类详情失败',
        categoryLoading: false
      });
    }
  },

  createCategory: async (data) => {
    set({ categoryLoading: true, categoryError: null });
    try {
      const category = await CategoryService.createCategory(data);
      const { categories } = get();
      set({
        categories: [category, ...categories],
        categoryLoading: false
      });
      
      // 重新获取分类树
      await get().fetchCategoryTree();
      return category;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '创建分类失败';
      set({ categoryError: errorMessage, categoryLoading: false });
      throw error;
    }
  },

  updateCategory: async (id, data) => {
    set({ categoryLoading: true, categoryError: null });
    try {
      const category = await CategoryService.updateCategory(id, data);
      const { categories } = get();
      const updatedCategories = categories.map(c => c.id === id ? category : c);
      set({
        categories: updatedCategories,
        currentCategory: category,
        categoryLoading: false
      });
      
      // 重新获取分类树
      await get().fetchCategoryTree();
      return category;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '更新分类失败';
      set({ categoryError: errorMessage, categoryLoading: false });
      throw error;
    }
  },

  deleteCategory: async (id) => {
    set({ categoryLoading: true, categoryError: null });
    try {
      await CategoryService.deleteCategory(id);
      const { categories } = get();
      const filteredCategories = categories.filter(c => c.id !== id);
      set({
        categories: filteredCategories,
        categoryLoading: false
      });
      
      // 重新获取分类树
      await get().fetchCategoryTree();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '删除分类失败';
      set({ categoryError: errorMessage, categoryLoading: false });
      throw error;
    }
  },

  // 图片上传
  uploadImage: async (productId, file, metadata = {}) => {
    set({ loading: true, error: null });
    try {
      const result = await ProductService.uploadImage(productId, file, metadata);
      set({ loading: false });
      return result;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '图片上传失败';
      set({ error: errorMessage, loading: false });
      throw error;
    }
  },

  // 变更历史
  fetchProductHistory: async (productId) => {
    set({ loading: true, error: null });
    try {
      const history = await ProductService.getProductHistory(productId);
      set({ loading: false });
      return history;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '获取变更历史失败';
      set({ error: errorMessage, loading: false });
      throw error;
    }
  }
}));
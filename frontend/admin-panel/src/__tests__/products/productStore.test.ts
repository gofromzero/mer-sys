// 商品状态管理测试
import { act, renderHook } from '@testing-library/react';
import { useProductStore } from '../../stores/productStore';

// Mock ProductService and CategoryService
jest.mock('../../services/productService', () => ({
  ProductService: {
    listProducts: jest.fn().mockResolvedValue({
      products: [
        {
          product: {
            id: 1,
            name: '测试商品',
            status: 'active',
            price: { amount: 9999, currency: 'CNY' },
            inventory: { stock_quantity: 100, reserved_quantity: 0, track_inventory: true }
          }
        }
      ],
      total: 1,
      page: 1,
      page_size: 20
    }),
    createProduct: jest.fn().mockResolvedValue({
      id: 2,
      name: '新商品',
      status: 'draft',
      price: { amount: 5000, currency: 'CNY' },
      inventory: { stock_quantity: 50, reserved_quantity: 0, track_inventory: true }
    }),
    updateProduct: jest.fn().mockResolvedValue({
      id: 1,
      name: '更新后的商品',
      status: 'active',
      price: { amount: 8888, currency: 'CNY' },
      inventory: { stock_quantity: 80, reserved_quantity: 5, track_inventory: true }
    }),
    deleteProduct: jest.fn().mockResolvedValue(undefined),
    updateProductStatus: jest.fn().mockResolvedValue(undefined),
    batchOperation: jest.fn().mockResolvedValue(undefined),
    getProduct: jest.fn().mockResolvedValue({
      id: 1,
      name: '测试商品详情',
      status: 'active',
      price: { amount: 9999, currency: 'CNY' },
      inventory: { stock_quantity: 100, reserved_quantity: 0, track_inventory: true }
    }),
    uploadImage: jest.fn().mockResolvedValue({ url: 'http://example.com/image.jpg' }),
    getProductHistory: jest.fn().mockResolvedValue([
      {
        id: 1,
        field_name: 'name',
        old_value: '旧名称',
        new_value: '新名称',
        operation: 'update',
        changed_at: '2023-01-01T00:00:00Z'
      }
    ])
  },
  CategoryService: {
    getCategoryList: jest.fn().mockResolvedValue([
      { id: 1, name: '电子产品', level: 1, path: '电子产品', status: 1 },
      { id: 2, name: '手机', level: 2, path: '电子产品/手机', status: 1 }
    ]),
    getCategoryTree: jest.fn().mockResolvedValue([
      {
        id: 1,
        name: '电子产品',
        level: 1,
        path: '电子产品',
        status: 1,
        children: [
          { id: 2, name: '手机', level: 2, path: '电子产品/手机', status: 1 }
        ]
      }
    ]),
    createCategory: jest.fn().mockResolvedValue({
      id: 3,
      name: '新分类',
      level: 1,
      path: '新分类',
      status: 1
    }),
    updateCategory: jest.fn().mockResolvedValue({
      id: 1,
      name: '更新分类',
      level: 1,
      path: '更新分类',
      status: 1
    }),
    deleteCategory: jest.fn().mockResolvedValue(undefined),
    getCategory: jest.fn().mockResolvedValue({
      id: 1,
      name: '电子产品',
      level: 1,
      path: '电子产品',
      status: 1
    })
  }
}));

describe('useProductStore', () => {
  beforeEach(() => {
    const { result } = renderHook(() => useProductStore());
    act(() => {
      result.current.setProducts([]);
      result.current.setCategories([]);
      result.current.setError(null);
      result.current.setCategoryError(null);
    });
  });

  describe('基础状态管理', () => {
    it('应该正确设置初始状态', () => {
      const { result } = renderHook(() => useProductStore());
      
      expect(result.current.products).toEqual([]);
      expect(result.current.currentProduct).toBeNull();
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBeNull();
      expect(result.current.categories).toEqual([]);
      expect(result.current.categoryTree).toEqual([]);
      expect(result.current.currentPage).toBe(1);
      expect(result.current.pageSize).toBe(20);
    });

    it('应该正确更新商品列表', () => {
      const { result } = renderHook(() => useProductStore());
      const testProducts = [
        { id: 1, name: '测试商品1', status: 'active' as const, price: { amount: 100, currency: 'CNY' }, inventory: { stock_quantity: 10, reserved_quantity: 0, track_inventory: true } },
        { id: 2, name: '测试商品2', status: 'draft' as const, price: { amount: 200, currency: 'CNY' }, inventory: { stock_quantity: 20, reserved_quantity: 0, track_inventory: true } }
      ] as any[];
      
      act(() => {
        result.current.setProducts(testProducts);
      });
      
      expect(result.current.products).toEqual(testProducts);
    });

    it('应该正确设置错误状态', () => {
      const { result } = renderHook(() => useProductStore());
      const errorMessage = '获取商品失败';
      
      act(() => {
        result.current.setError(errorMessage);
      });
      
      expect(result.current.error).toBe(errorMessage);
      
      act(() => {
        result.current.clearError();
      });
      
      expect(result.current.error).toBeNull();
    });
  });

  describe('商品API操作', () => {
    it('应该正确获取商品列表', async () => {
      const { result } = renderHook(() => useProductStore());
      
      await act(async () => {
        await result.current.fetchProducts();
      });
      
      expect(result.current.products).toHaveLength(1);
      expect(result.current.products[0].name).toBe('测试商品');
      expect(result.current.loading).toBe(false);
    });

    it('应该正确创建商品', async () => {
      const { result } = renderHook(() => useProductStore());
      const newProductData = {
        name: '新商品',
        price: { amount: 5000, currency: 'CNY' },
        rights_cost: 0,
        inventory: { stock_quantity: 50, reserved_quantity: 0, track_inventory: true }
      };
      
      let createdProduct;
      await act(async () => {
        createdProduct = await result.current.createProduct(newProductData);
      });
      
      expect(createdProduct).toBeDefined();
      expect(createdProduct?.name).toBe('新商品');
    });

    it('应该正确处理API错误', async () => {
      const { ProductService } = require('../../services/productService');
      ProductService.listProducts.mockRejectedValueOnce(new Error('网络错误'));
      
      const { result } = renderHook(() => useProductStore());
      
      await act(async () => {
        await result.current.fetchProducts();
      });
      
      expect(result.current.error).toBe('网络错误');
      expect(result.current.loading).toBe(false);
    });
  });

  describe('分类操作', () => {
    it('应该正确获取分类树', async () => {
      const { result } = renderHook(() => useProductStore());
      
      await act(async () => {
        await result.current.fetchCategoryTree();
      });
      
      expect(result.current.categoryTree).toHaveLength(1);
      expect(result.current.categoryTree[0].name).toBe('电子产品');
      expect(result.current.categoryTree[0].children).toHaveLength(1);
    });

    it('应该正确创建分类', async () => {
      const { result } = renderHook(() => useProductStore());
      const newCategoryData = {
        name: '新分类',
        sort_order: 0
      };
      
      let createdCategory;
      await act(async () => {
        createdCategory = await result.current.createCategory(newCategoryData);
      });
      
      expect(createdCategory).toBeDefined();
      expect(createdCategory?.name).toBe('新分类');
    });
  });

  describe('筛选和分页', () => {
    it('应该正确设置筛选条件', () => {
      const { result } = renderHook(() => useProductStore());
      const filters = { keyword: '测试', status: 'active' as const };
      
      act(() => {
        result.current.setFilters(filters);
      });
      
      expect(result.current.filters).toEqual(filters);
    });

    it('应该正确设置分页参数', () => {
      const { result } = renderHook(() => useProductStore());
      
      act(() => {
        result.current.setCurrentPage(2);
        result.current.setPageSize(10);
      });
      
      expect(result.current.currentPage).toBe(2);
      expect(result.current.pageSize).toBe(10);
    });

    it('应该正确设置排序', () => {
      const { result } = renderHook(() => useProductStore());
      
      act(() => {
        result.current.setSorting('name', 'asc');
      });
      
      expect(result.current.sortBy).toBe('name');
      expect(result.current.sortOrder).toBe('asc');
    });
  });
});
import { create } from 'zustand';
import type { Cart } from '../types/order';
import { cartService } from '../services/cartService';

interface CartState {
  // 状态
  cart: Cart | null;
  isLoading: boolean;
  error: string | null;

  // 操作
  getCart: () => Promise<void>;
  addItem: (productId: number, quantity: number) => Promise<void>;
  updateItem: (itemId: number, quantity: number) => Promise<void>;
  removeItem: (itemId: number) => Promise<void>;
  clearCart: () => Promise<void>;
  
  // 重置状态
  clearError: () => void;
}

export const useCartStore = create<CartState>((set, get) => ({
  // 初始状态
  cart: null,
  isLoading: false,
  error: null,

  // 获取购物车
  getCart: async () => {
    set({ isLoading: true, error: null });
    try {
      const cart = await cartService.getCart();
      set({ cart, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取购物车失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 添加商品到购物车
  addItem: async (productId: number, quantity: number) => {
    set({ isLoading: true, error: null });
    try {
      await cartService.addItem({ product_id: productId, quantity });
      
      // 重新获取购物车数据
      await get().getCart();
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '添加商品失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 更新购物车商品数量
  updateItem: async (itemId: number, quantity: number) => {
    set({ isLoading: true, error: null });
    try {
      await cartService.updateItem(itemId, { quantity });
      
      // 重新获取购物车数据
      await get().getCart();
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '更新商品数量失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 从购物车移除商品
  removeItem: async (itemId: number) => {
    set({ isLoading: true, error: null });
    try {
      await cartService.removeItem(itemId);
      
      // 重新获取购物车数据
      await get().getCart();
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '删除商品失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 清空购物车
  clearCart: async () => {
    set({ isLoading: true, error: null });
    try {
      await cartService.clearCart();
      set({ cart: null, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '清空购物车失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 重置状态方法
  clearError: () => set({ error: null }),
}));
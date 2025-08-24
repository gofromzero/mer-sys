import { apiClient } from './api';
import type {
  Cart,
  AddCartItemRequest,
  UpdateCartItemRequest,
} from '../types/order';

export const cartService = {
  // 获取购物车
  async getCart(): Promise<Cart> {
    const response = await apiClient.get('/cart');
    return response.data;
  },

  // 添加商品到购物车
  async addItem(data: AddCartItemRequest): Promise<void> {
    await apiClient.post('/cart/items', data);
  },

  // 更新购物车商品数量
  async updateItem(itemId: number, data: UpdateCartItemRequest): Promise<void> {
    await apiClient.put(`/cart/items/${itemId}`, data);
  },

  // 从购物车移除商品
  async removeItem(itemId: number): Promise<void> {
    await apiClient.delete(`/cart/items/${itemId}`);
  },

  // 清空购物车
  async clearCart(): Promise<void> {
    await apiClient.delete('/cart');
  },
};
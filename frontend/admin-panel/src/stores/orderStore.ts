import { create } from 'zustand';
import type { Order, OrderConfirmation } from '../types/order';
import { orderService } from '../services/orderService';

interface OrderState {
  // 状态
  orders: Order[];
  currentOrder: Order | null;
  orderConfirmation: OrderConfirmation | null;
  paymentStatus: string | null;
  isLoading: boolean;
  error: string | null;

  // 操作
  createOrder: (merchantId: number, items: { product_id: number; quantity: number }[]) => Promise<void>;
  getOrder: (orderId: number) => Promise<void>;
  getOrders: (params?: { status?: string; page?: number; limit?: number }) => Promise<void>;
  cancelOrder: (orderId: number) => Promise<void>;
  initiatePayment: (orderId: number, paymentMethod: string, returnUrl?: string) => Promise<void>;
  checkPaymentStatus: (orderId: number) => Promise<void>;
  retryPayment: (orderId: number, paymentMethod: string, returnUrl?: string) => Promise<void>;
  getOrderConfirmation: (merchantId: number, items: { product_id: number; quantity: number }[]) => Promise<void>;
  
  // 重置状态
  clearError: () => void;
  clearCurrentOrder: () => void;
  clearOrderConfirmation: () => void;
}

export const useOrderStore = create<OrderState>((set, get) => ({
  // 初始状态
  orders: [],
  currentOrder: null,
  orderConfirmation: null,
  paymentStatus: null,
  isLoading: false,
  error: null,

  // 创建订单
  createOrder: async (merchantId: number, items: { product_id: number; quantity: number }[]) => {
    set({ isLoading: true, error: null });
    try {
      const order = await orderService.createOrder({
        merchant_id: merchantId,
        items,
      });
      set({ currentOrder: order, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '创建订单失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 获取订单详情
  getOrder: async (orderId: number) => {
    set({ isLoading: true, error: null });
    try {
      const order = await orderService.getOrder(orderId);
      set({ currentOrder: order, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取订单失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 获取订单列表
  getOrders: async (params = {}) => {
    set({ isLoading: true, error: null });
    try {
      const result = await orderService.getOrders(params);
      set({ orders: result.items, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取订单列表失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 取消订单
  cancelOrder: async (orderId: number) => {
    set({ isLoading: true, error: null });
    try {
      await orderService.cancelOrder(orderId);
      
      // 更新当前订单状态
      const { currentOrder } = get();
      if (currentOrder && currentOrder.id === orderId) {
        set({ currentOrder: { ...currentOrder, status: 'cancelled' } });
      }
      
      // 更新订单列表中的状态
      set(state => ({
        orders: state.orders.map(order =>
          order.id === orderId ? { ...order, status: 'cancelled' as const } : order
        ),
        isLoading: false,
      }));
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '取消订单失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 发起支付
  initiatePayment: async (orderId: number, paymentMethod: string, returnUrl?: string) => {
    set({ isLoading: true, error: null });
    try {
      const result = await orderService.initiatePayment(orderId, {
        payment_method: paymentMethod as any,
        return_url: returnUrl,
      });
      
      // 如果返回了支付URL，可以直接跳转
      if (result.payment_url) {
        window.location.href = result.payment_url;
      }
      
      set({ isLoading: false });
      return result;
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '发起支付失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 查询支付状态
  checkPaymentStatus: async (orderId: number) => {
    set({ isLoading: true, error: null });
    try {
      const result = await orderService.getPaymentStatus(orderId);
      set({ paymentStatus: result.payment_status, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '查询支付状态失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 重新支付
  retryPayment: async (orderId: number, paymentMethod: string, returnUrl?: string) => {
    set({ isLoading: true, error: null });
    try {
      const result = await orderService.retryPayment(orderId, {
        payment_method: paymentMethod as any,
        return_url: returnUrl,
      });
      
      // 如果返回了支付URL，可以直接跳转
      if (result.payment_url) {
        window.location.href = result.payment_url;
      }
      
      set({ isLoading: false });
      return result;
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '重新支付失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 获取订单确认信息
  getOrderConfirmation: async (merchantId: number, items: { product_id: number; quantity: number }[]) => {
    set({ isLoading: true, error: null });
    try {
      const confirmation = await orderService.getOrderConfirmation({
        merchant_id: merchantId,
        items,
      });
      set({ orderConfirmation: confirmation, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : '获取订单确认信息失败',
        isLoading: false,
      });
      throw error;
    }
  },

  // 重置状态方法
  clearError: () => set({ error: null }),
  clearCurrentOrder: () => set({ currentOrder: null }),
  clearOrderConfirmation: () => set({ orderConfirmation: null }),
}));
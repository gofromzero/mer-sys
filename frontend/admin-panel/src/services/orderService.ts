import { apiClient } from './api';
import type {
  Order,
  CreateOrderRequest,
  InitiatePaymentRequest,
  OrderConfirmation,
} from '../types/order';

export const orderService = {
  // 创建订单
  async createOrder(data: CreateOrderRequest): Promise<Order> {
    const response = await apiClient.post('/orders', data);
    return response.data;
  },

  // 获取订单详情
  async getOrder(orderId: number): Promise<Order> {
    const response = await apiClient.get(`/orders/${orderId}`);
    return response.data;
  },

  // 获取订单列表
  async getOrders(params: {
    status?: string;
    page?: number;
    limit?: number;
  }): Promise<{
    items: Order[];
    total: number;
    page: number;
    limit: number;
  }> {
    const response = await apiClient.get('/orders', { params });
    return response.data;
  },

  // 取消订单
  async cancelOrder(orderId: number): Promise<void> {
    await apiClient.put(`/orders/${orderId}/cancel`);
  },

  // 发起支付
  async initiatePayment(
    orderId: number,
    data: InitiatePaymentRequest
  ): Promise<any> {
    const response = await apiClient.post(`/orders/${orderId}/pay`, data);
    return response.data;
  },

  // 查询支付状态
  async getPaymentStatus(orderId: number): Promise<{ payment_status: string }> {
    const response = await apiClient.get(`/orders/${orderId}/payment-status`);
    return response.data;
  },

  // 重新支付
  async retryPayment(
    orderId: number,
    data: InitiatePaymentRequest
  ): Promise<any> {
    const response = await apiClient.post(`/orders/${orderId}/retry-payment`, data);
    return response.data;
  },

  // 获取订单确认信息（实际上应该是预创建订单的接口）
  async getOrderConfirmation(data: CreateOrderRequest): Promise<OrderConfirmation> {
    // 注意：这里应该调用一个预览订单的API，暂时使用创建订单的逻辑
    // 在实际项目中，可能需要添加一个 /orders/preview 端点
    const response = await apiClient.post('/orders/preview', data);
    return response.data;
  },
};
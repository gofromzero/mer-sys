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

  // 获取订单详情（包含状态历史）
  async getOrderWithHistory(orderId: number): Promise<Order> {
    const response = await apiClient.get(`/orders/${orderId}/detail`);
    return response.data;
  },

  // 获取订单状态历史
  async getOrderStatusHistory(orderId: number, params?: {
    limit?: number;
    offset?: number;
  }): Promise<{
    items: Array<{
      id: number;
      order_id: number;
      from_status: number;
      to_status: number;
      reason: string;
      operator_type: string;
      operator_id?: number;
      created_at: string;
    }>;
    total: number;
  }> {
    const response = await apiClient.get(`/orders/${orderId}/status-history`, { params });
    return response.data;
  },

  // 高级订单查询
  async queryOrders(params: {
    merchant_id?: number;
    customer_id?: number;
    status?: number[];
    start_date?: string;
    end_date?: string;
    search_keyword?: string;
    page?: number;
    page_size?: number;
    sort_by?: string;
    sort_order?: string;
  }): Promise<{
    items: Order[];
    total: number;
    page: number;
    page_size: number;
    has_next: boolean;
  }> {
    const response = await apiClient.get('/orders/query', { params });
    return response.data;
  },

  // 搜索订单
  async searchOrders(keyword: string, params?: {
    page?: number;
    page_size?: number;
  }): Promise<{
    items: Order[];
    total: number;
    page: number;
    page_size: number;
  }> {
    const response = await apiClient.get('/orders/search', { 
      params: { q: keyword, ...params } 
    });
    return response.data;
  },

  // 获取订单统计信息
  async getOrderStats(params?: {
    merchant_id?: number;
    start_date?: string;
    end_date?: string;
  }): Promise<{
    total: number;
    by_status: Record<string, number>;
  }> {
    const response = await apiClient.get('/orders/stats', { params });
    return response.data;
  },
};
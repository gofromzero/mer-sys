// 订单相关类型定义

export type OrderStatus = 'pending' | 'paid' | 'processing' | 'completed' | 'cancelled';

export type PaymentMethod = 'alipay' | 'wechat' | 'balance';

export interface Order {
  id: number;
  tenant_id: number;
  merchant_id: number;
  customer_id: number;
  order_number: string;
  status: OrderStatus;
  items: OrderItem[];
  payment_info?: PaymentInfo;
  verification_info?: VerificationInfo;
  total_amount: number;
  total_rights_cost: number;
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  product_id: number;
  quantity: number;
  price: number;
  rights_cost: number;
}

export interface PaymentInfo {
  method: string;
  transaction_id: string;
  paid_at?: string;
  amount: number;
}

export interface VerificationInfo {
  verification_code: string;
  qr_code_url: string;
  verified_at?: string;
  verified_by?: string;
}

export interface Cart {
  id: number;
  tenant_id: number;
  customer_id: number;
  items: CartItem[];
  created_at: string;
  updated_at: string;
  expires_at: string;
}

export interface CartItem {
  id: number;
  tenant_id: number;
  cart_id: number;
  product_id: number;
  quantity: number;
  added_at: string;
}

export interface CreateOrderRequest {
  merchant_id: number;
  items: {
    product_id: number;
    quantity: number;
  }[];
}

export interface AddCartItemRequest {
  product_id: number;
  quantity: number;
}

export interface UpdateCartItemRequest {
  quantity: number;
}

export interface InitiatePaymentRequest {
  payment_method: PaymentMethod;
  return_url?: string;
}

export interface OrderConfirmation {
  items: OrderConfirmationItem[];
  total_amount: number;
  total_rights_cost: number;
  available_rights: number;
  can_create: boolean;
  error_message?: string;
}

export interface OrderConfirmationItem {
  product_id: number;
  product_name: string;
  quantity: number;
  unit_price: number;
  unit_rights_cost: number;
  subtotal_amount: number;
  subtotal_rights_cost: number;
  stock_available: number;
  stock_sufficient: boolean;
}
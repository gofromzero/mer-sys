import React from 'react';
import { useNavigate } from 'react-router-dom';
import { OrderForm } from '../../../components/orders/OrderForm';
import { Order } from '../../../types/order';

const OrderFormPage: React.FC = () => {
  const navigate = useNavigate();

  const handleOrderCreated = (order: Order) => {
    // 订单创建成功后跳转到订单详情页面
    navigate(`/orders/${order.id}`);
  };

  const handleCancel = () => {
    // 取消订单创建，返回购物车页面
    navigate('/cart');
  };

  return (
    <div className="container mx-auto px-4 py-6">
      <OrderForm
        onOrderCreated={handleOrderCreated}
        onCancel={handleCancel}
      />
    </div>
  );
};

export default OrderFormPage;
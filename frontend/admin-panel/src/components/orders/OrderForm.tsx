import React, { useState, useEffect } from 'react';
import { AmisRenderer } from 'amis-react';
import { useOrderStore } from '../../stores/orderStore';
import { useCartStore } from '../../stores/cartStore';
import { orderService } from '../../services/orderService';
import { cartService } from '../../services/cartService';
import { Order, OrderConfirmation } from '../../types/order';

interface OrderFormProps {
  onOrderCreated?: (order: Order) => void;
  onCancel?: () => void;
}

export const OrderForm: React.FC<OrderFormProps> = ({
  onOrderCreated,
  onCancel
}) => {
  const { createOrder } = useOrderStore();
  const { cart, clearCart } = useCartStore();
  const [loading, setLoading] = useState(false);
  const [confirmation, setConfirmation] = useState<OrderConfirmation | null>(null);

  // 获取订单确认信息
  useEffect(() => {
    const loadConfirmation = async () => {
      if (!cart || cart.items.length === 0) return;
      
      try {
        const confirmationData = await orderService.getOrderConfirmation(cart.id);
        setConfirmation(confirmationData);
      } catch (error) {
        console.error('获取订单确认信息失败:', error);
      }
    };

    loadConfirmation();
  }, [cart]);

  // 处理订单创建
  const handleCreateOrder = async (formData: any) => {
    if (!cart || !confirmation) return;

    setLoading(true);
    try {
      const orderData = {
        cart_id: cart.id,
        payment_method: formData.payment_method || 'alipay',
        return_url: formData.return_url || window.location.origin + '/orders'
      };

      const newOrder = await createOrder(orderData);
      await clearCart(); // 订单创建成功后清空购物车
      
      onOrderCreated?.(newOrder);
    } catch (error) {
      console.error('创建订单失败:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  // Amis Schema 配置
  const schema = {
    type: 'page',
    title: '订单确认',
    body: [
      // 商品信息展示
      {
        type: 'card',
        header: {
          title: '订单商品',
          className: 'text-lg font-semibold'
        },
        body: {
          type: 'table',
          source: confirmation?.items || [],
          columns: [
            {
              name: 'product_name',
              label: '商品名称',
              type: 'text'
            },
            {
              name: 'quantity',
              label: '数量',
              type: 'text',
              tpl: '${quantity}'
            },
            {
              name: 'unit_price',
              label: '单价',
              type: 'text',
              tpl: '¥${unit_price.amount}'
            },
            {
              name: 'unit_rights_cost',
              label: '权益消耗',
              type: 'text',
              tpl: '${unit_rights_cost} 权益'
            },
            {
              name: 'subtotal_amount',
              label: '小计',
              type: 'text',
              tpl: '¥${subtotal_amount.amount}',
              className: 'font-medium'
            }
          ]
        }
      },

      // 订单汇总
      {
        type: 'card',
        header: {
          title: '订单汇总',
          className: 'text-lg font-semibold'
        },
        body: [
          {
            type: 'grid',
            columns: [
              {
                md: 6,
                body: [
                  {
                    type: 'static',
                    label: '商品总价',
                    value: confirmation ? `¥${confirmation.total_amount.amount}` : '¥0.00'
                  },
                  {
                    type: 'static',
                    label: '权益消耗',
                    value: confirmation ? `${confirmation.total_rights_cost} 权益` : '0 权益'
                  }
                ]
              },
              {
                md: 6,
                body: [
                  {
                    type: 'static',
                    label: '应付金额',
                    value: confirmation ? `¥${confirmation.total_amount.amount}` : '¥0.00',
                    className: 'text-xl font-bold text-red-600'
                  }
                ]
              }
            ]
          }
        ]
      },

      // 支付方式选择和订单创建表单
      {
        type: 'form',
        api: {
          method: 'post',
          url: '/api/orders/create',
          requestAdaptor: (api: any) => {
            // 通过自定义处理函数来处理表单提交
            return api;
          }
        },
        body: [
          {
            type: 'radios',
            name: 'payment_method',
            label: '支付方式',
            value: 'alipay',
            required: true,
            options: [
              {
                label: '支付宝',
                value: 'alipay'
              },
              {
                label: '微信支付',
                value: 'wechat',
                disabled: true,
                description: '暂未开通'
              }
            ]
          },
          {
            type: 'input-url',
            name: 'return_url',
            label: '支付完成返回地址',
            placeholder: '可选，留空则返回订单列表',
            description: '支付完成后将跳转到此地址'
          },
          {
            type: 'divider'
          },
          {
            type: 'grid',
            columns: [
              {
                md: 6,
                body: {
                  type: 'button',
                  label: '取消',
                  level: 'default',
                  size: 'lg',
                  className: 'w-full',
                  onClick: () => onCancel?.()
                }
              },
              {
                md: 6,
                body: {
                  type: 'submit',
                  label: loading ? '创建中...' : '确认下单',
                  level: 'primary',
                  size: 'lg',
                  className: 'w-full',
                  disabled: loading || !confirmation || (confirmation?.items?.length || 0) === 0
                }
              }
            ]
          }
        ],
        onFinished: (values: any, response: any) => {
          // 实际的订单创建逻辑由 handleCreateOrder 处理
          return handleCreateOrder(values);
        }
      }
    ]
  };

  // 如果没有购物车或购物车为空，显示空状态
  if (!cart || cart.items.length === 0) {
    return (
      <AmisRenderer
        schema={{
          type: 'page',
          title: '订单确认',
          body: {
            type: 'alert',
            level: 'info',
            showCloseButton: false,
            body: '购物车为空，请先添加商品到购物车。',
            actions: [
              {
                type: 'button',
                label: '返回购物车',
                level: 'primary',
                onClick: () => onCancel?.()
              }
            ]
          }
        }}
      />
    );
  }

  return (
    <AmisRenderer
      schema={schema}
      data={{
        confirmation,
        cart,
        loading
      }}
    />
  );
};

export default OrderForm;
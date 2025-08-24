import React, { useEffect } from 'react';
import { AmisRenderer } from '../../../components/ui/AmisRenderer';
import { useCartStore } from '../../../stores/cartStore';
import { useOrderStore } from '../../../stores/orderStore';
import type { SchemaNode } from 'amis';

const CartPage: React.FC = () => {
  const { cart, isLoading, error, getCart, updateItem, removeItem, clearCart } = useCartStore();
  const { createOrder } = useOrderStore();

  useEffect(() => {
    getCart();
  }, [getCart]);

  const schema: SchemaNode = {
    type: 'page',
    title: '购物车',
    body: [
      {
        type: 'alert',
        body: '这里是您的购物车，确认商品信息后可以提交订单',
        level: 'info',
        showIcon: true,
      },
      {
        type: 'service',
        api: {
          method: 'get',
          url: '/api/v1/cart',
          adaptor: (payload: any) => {
            return {
              ...payload,
              data: {
                items: cart?.items || [],
                total_items: cart?.items.length || 0,
              },
            };
          },
        },
        body: [
          {
            type: 'table',
            source: '${items}',
            columns: [
              {
                name: 'product_id',
                label: '商品ID',
                type: 'text',
              },
              {
                name: 'quantity',
                label: '数量',
                type: 'input-number',
                min: 1,
                onEvent: {
                  change: {
                    actions: [
                      {
                        actionType: 'ajax',
                        api: {
                          method: 'put',
                          url: '/api/v1/cart/items/${id}',
                          data: {
                            quantity: '${event.data.value}',
                          },
                        },
                      },
                    ],
                  },
                },
              },
              {
                name: 'added_at',
                label: '添加时间',
                type: 'datetime',
                format: 'YYYY-MM-DD HH:mm:ss',
              },
              {
                type: 'operation',
                label: '操作',
                buttons: [
                  {
                    type: 'button',
                    label: '删除',
                    level: 'danger',
                    confirmText: '确认删除这个商品吗？',
                    onEvent: {
                      click: {
                        actions: [
                          {
                            actionType: 'ajax',
                            api: {
                              method: 'delete',
                              url: '/api/v1/cart/items/${id}',
                            },
                          },
                        ],
                      },
                    },
                  },
                ],
              },
            ],
          },
          {
            type: 'divider',
          },
          {
            type: 'flex',
            justify: 'space-between',
            items: [
              {
                type: 'button',
                label: '清空购物车',
                level: 'warning',
                confirmText: '确认清空购物车吗？',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'ajax',
                        api: {
                          method: 'delete',
                          url: '/api/v1/cart',
                        },
                      },
                    ],
                  },
                },
              },
              {
                type: 'button',
                label: '去下单',
                level: 'primary',
                size: 'lg',
                disabled: '${!items || items.length === 0}',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'url',
                        url: '/orders/create',
                        blank: false
                      }
                    ]
                  }
                }
              },
            ],
          },
        ],
      },
    ],
  };

  if (error) {
    return (
      <div className="p-4 text-center">
        <div className="text-red-500 mb-4">加载失败: {error}</div>
        <button
          onClick={() => getCart()}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          重试
        </button>
      </div>
    );
  }

  return <AmisRenderer schema={schema} />;
};

export default CartPage;
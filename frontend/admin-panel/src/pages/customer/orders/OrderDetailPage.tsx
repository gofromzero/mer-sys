import React, { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { AmisRenderer } from '../../../components/ui/AmisRenderer';
import { useOrderStore } from '../../../stores/orderStore';
import type { SchemaNode } from 'amis';

const OrderDetailPage: React.FC = () => {
  const { orderId } = useParams<{ orderId: string }>();
  const { currentOrder, isLoading, error, getOrder, initiatePayment, cancelOrder } = useOrderStore();

  useEffect(() => {
    if (orderId) {
      getOrder(parseInt(orderId));
    }
  }, [orderId, getOrder]);

  const schema: SchemaNode = {
    type: 'page',
    title: '订单详情',
    body: [
      {
        type: 'service',
        api: {
          method: 'get',
          url: `/api/v1/orders/${orderId}/detail`,
        },
        body: [
          {
            type: 'panel',
            title: '订单信息',
            body: [
              {
                type: 'descriptions',
                column: 2,
                items: [
                  {
                    label: '订单号',
                    content: '${order_number}',
                  },
                  {
                    label: '订单状态',
                    content: {
                      type: 'mapping',
                      source: '${status}',
                      map: {
                        1: '<span class="label label-warning">待支付</span>',
                        2: '<span class="label label-info">已支付</span>',
                        3: '<span class="label label-primary">处理中</span>',
                        4: '<span class="label label-success">已完成</span>',
                        5: '<span class="label label-default">已取消</span>',
                      },
                    },
                  },
                  {
                    label: '订单金额',
                    content: '¥${total_amount}',
                  },
                  {
                    label: '权益成本',
                    content: '${total_rights_cost} 积分',
                  },
                  {
                    label: '创建时间',
                    content: '${created_at | date:YYYY-MM-DD HH:mm:ss}',
                  },
                  {
                    label: '更新时间',
                    content: '${updated_at | date:YYYY-MM-DD HH:mm:ss}',
                  },
                ],
              },
            ],
          },
          {
            type: 'panel',
            title: '商品明细',
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
                    type: 'text',
                  },
                  {
                    name: 'price',
                    label: '单价',
                    type: 'text',
                    tpl: '¥${price}',
                  },
                  {
                    name: 'rights_cost',
                    label: '权益成本',
                    type: 'text',
                    tpl: '${rights_cost} 积分',
                  },
                  {
                    label: '小计金额',
                    type: 'text',
                    tpl: '¥${price * quantity}',
                  },
                  {
                    label: '小计权益',
                    type: 'text',
                    tpl: '${rights_cost * quantity} 积分',
                  },
                ],
              },
            ],
          },
          {
            type: 'panel',
            title: '支付信息',
            visibleOn: '${payment_info}',
            body: [
              {
                type: 'descriptions',
                column: 2,
                items: [
                  {
                    label: '支付方式',
                    content: '${payment_info.method}',
                  },
                  {
                    label: '交易号',
                    content: '${payment_info.transaction_id}',
                  },
                  {
                    label: '支付金额',
                    content: '¥${payment_info.amount}',
                  },
                  {
                    label: '支付时间',
                    content: '${payment_info.paid_at | date:YYYY-MM-DD HH:mm:ss}',
                    visibleOn: '${payment_info.paid_at}',
                  },
                ],
              },
            ],
          },
          {
            type: 'panel',
            title: '核销信息',
            visibleOn: '${verification_info}',
            body: [
              {
                type: 'descriptions',
                column: 2,
                items: [
                  {
                    label: '核销码',
                    content: '${verification_info.verification_code}',
                  },
                  {
                    label: '二维码',
                    content: {
                      type: 'qrcode',
                      value: '${verification_info.verification_code}',
                    },
                  },
                  {
                    label: '核销时间',
                    content: '${verification_info.verified_at | date:YYYY-MM-DD HH:mm:ss}',
                    visibleOn: '${verification_info.verified_at}',
                  },
                  {
                    label: '核销人员',
                    content: '${verification_info.verified_by}',
                    visibleOn: '${verification_info.verified_by}',
                  },
                ],
              },
            ],
          },
          {
            type: 'panel',
            title: '订单状态历史',
            body: [
              {
                type: 'table',
                source: '${status_history}',
                placeholder: '暂无状态变更记录',
                columns: [
                  {
                    name: 'from_status',
                    label: '原状态',
                    type: 'mapping',
                    map: {
                      1: '待支付',
                      2: '已支付',
                      3: '处理中',
                      4: '已完成',
                      5: '已取消',
                    },
                  },
                  {
                    name: 'to_status',
                    label: '新状态',
                    type: 'mapping',
                    map: {
                      1: '<span class="label label-warning">待支付</span>',
                      2: '<span class="label label-info">已支付</span>',
                      3: '<span class="label label-primary">处理中</span>',
                      4: '<span class="label label-success">已完成</span>',
                      5: '<span class="label label-default">已取消</span>',
                    },
                  },
                  {
                    name: 'reason',
                    label: '变更原因',
                    type: 'text',
                  },
                  {
                    name: 'operator_type',
                    label: '操作者',
                    type: 'mapping',
                    map: {
                      customer: '<span class="label label-info">客户</span>',
                      merchant: '<span class="label label-success">商户</span>',
                      system: '<span class="label label-default">系统</span>',
                      admin: '<span class="label label-primary">管理员</span>',
                    },
                  },
                  {
                    name: 'created_at',
                    label: '变更时间',
                    type: 'datetime',
                    format: 'YYYY-MM-DD HH:mm:ss',
                  },
                ],
              },
            ],
          },
          {
            type: 'divider',
          },
          {
            type: 'button-group',
            buttons: [
              {
                type: 'button',
                label: '支付订单',
                level: 'primary',
                size: 'lg',
                visibleOn: '${status == 1}',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'dialog',
                        dialog: {
                          title: '选择支付方式',
                          body: {
                            type: 'form',
                            body: [
                              {
                                type: 'radios',
                                name: 'payment_method',
                                label: '支付方式',
                                required: true,
                                value: 'alipay',
                                options: [
                                  { label: '支付宝', value: 'alipay' },
                                  { label: '微信支付', value: 'wechat' },
                                ],
                              },
                              {
                                type: 'input-url',
                                name: 'return_url',
                                label: '支付成功返回页面',
                                placeholder: '请输入支付成功后的返回地址',
                              },
                            ],
                            actions: [
                              {
                                type: 'button',
                                label: '取消',
                                actionType: 'cancel',
                              },
                              {
                                type: 'submit',
                                label: '确认支付',
                                level: 'primary',
                                api: {
                                  method: 'post',
                                  url: `/api/v1/orders/${orderId}/pay`,
                                  data: {
                                    payment_method: '${payment_method}',
                                    return_url: '${return_url}',
                                  },
                                },
                              },
                            ],
                          },
                        },
                      },
                    ],
                  },
                },
              },
              {
                type: 'button',
                label: '取消订单',
                level: 'danger',
                visibleOn: '${status == 1}',
                confirmText: '确认取消这个订单吗？',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'ajax',
                        api: {
                          method: 'put',
                          url: `/api/v1/orders/${orderId}/cancel`,
                        },
                      },
                    ],
                  },
                },
              },
              {
                type: 'button',
                label: '查询支付状态',
                level: 'info',
                visibleOn: '${status == 1 || status == 2}',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'ajax',
                        api: {
                          method: 'get',
                          url: `/api/v1/orders/${orderId}/payment-status`,
                        },
                      },
                    ],
                  },
                },
              },
              {
                type: 'button',
                label: '返回订单列表',
                level: 'default',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'url',
                        url: '/customer/orders',
                      },
                    ],
                  },
                },
              },
            ],
          },
        ],
      },
    ],
  };

  if (isLoading) {
    return <div className="p-4 text-center">加载中...</div>;
  }

  if (error) {
    return (
      <div className="p-4 text-center">
        <div className="text-red-500 mb-4">加载失败: {error}</div>
        <button
          onClick={() => orderId && getOrder(parseInt(orderId))}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          重试
        </button>
      </div>
    );
  }

  return <AmisRenderer schema={schema} />;
};

export default OrderDetailPage;
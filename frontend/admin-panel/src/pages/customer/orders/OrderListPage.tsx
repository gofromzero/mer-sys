import React from 'react';
import { AmisRenderer } from '../../../components/ui/AmisRenderer';
import type { SchemaNode } from 'amis';

const OrderListPage: React.FC = () => {
  const schema: SchemaNode = {
    type: 'page',
    title: '我的订单',
    body: [
      {
        type: 'crud',
        api: {
          method: 'get',
          url: '/api/v1/orders',
          data: {
            page: '${page}',
            limit: '${perPage}',
            status: '${status}',
          },
        },
        filter: {
          title: '条件搜索',
          body: [
            {
              type: 'select',
              name: 'status',
              label: '订单状态',
              placeholder: '请选择订单状态',
              clearable: true,
              options: [
                { label: '待支付', value: 'pending' },
                { label: '已支付', value: 'paid' },
                { label: '处理中', value: 'processing' },
                { label: '已完成', value: 'completed' },
                { label: '已取消', value: 'cancelled' },
              ],
            },
          ],
        },
        columns: [
          {
            name: 'order_number',
            label: '订单号',
            type: 'text',
            copyable: true,
          },
          {
            name: 'status',
            label: '状态',
            type: 'mapping',
            map: {
              pending: '<span class="label label-warning">待支付</span>',
              paid: '<span class="label label-info">已支付</span>',
              processing: '<span class="label label-primary">处理中</span>',
              completed: '<span class="label label-success">已完成</span>',
              cancelled: '<span class="label label-default">已取消</span>',
            },
          },
          {
            name: 'total_amount',
            label: '订单金额',
            type: 'text',
            tpl: '¥${total_amount}',
          },
          {
            name: 'total_rights_cost',
            label: '权益成本',
            type: 'text',
            tpl: '${total_rights_cost} 积分',
          },
          {
            name: 'created_at',
            label: '创建时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm:ss',
          },
          {
            type: 'operation',
            label: '操作',
            buttons: [
              {
                type: 'button',
                label: '查看详情',
                level: 'link',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'drawer',
                        drawer: {
                          title: '订单详情',
                          size: 'lg',
                          body: {
                            type: 'service',
                            api: '/api/v1/orders/${id}',
                            body: [
                              {
                                type: 'descriptions',
                                title: '基本信息',
                                column: 2,
                                items: [
                                  {
                                    label: '订单号',
                                    content: '${order_number}',
                                  },
                                  {
                                    label: '状态',
                                    content: '${status}',
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
                                    content: '${created_at}',
                                  },
                                  {
                                    label: '更新时间',
                                    content: '${updated_at}',
                                  },
                                ],
                              },
                              {
                                type: 'divider',
                              },
                              {
                                type: 'table',
                                title: '商品明细',
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
                                ],
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
                label: '支付',
                level: 'primary',
                visibleOn: '${status === "pending"}',
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
                                  url: '/api/v1/orders/${id}/pay',
                                  data: {
                                    payment_method: '${payment_method}',
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
                visibleOn: '${status === "pending"}',
                confirmText: '确认取消这个订单吗？',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'ajax',
                        api: {
                          method: 'put',
                          url: '/api/v1/orders/${id}/cancel',
                        },
                      },
                    ],
                  },
                },
              },
            ],
          },
        ],
        pagination: {
          enable: true,
          layout: ['pager', 'perpage', 'total'],
        },
      },
    ],
  };

  return <AmisRenderer schema={schema} />;
};

export default OrderListPage;
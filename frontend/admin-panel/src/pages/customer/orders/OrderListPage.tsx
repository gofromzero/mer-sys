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
          url: '/api/v1/orders/query',
          data: {
            page: '${page}',
            page_size: '${perPage}',
            status: '${status}',
            start_date: '${start_date}',
            end_date: '${end_date}',
            search_keyword: '${search_keyword}',
            sort_by: '${orderBy}',
            sort_order: '${orderDir}',
          },
        },
        filter: {
          title: '订单筛选',
          body: [
            {
              type: 'input-text',
              name: 'search_keyword',
              label: '搜索',
              placeholder: '请输入订单号或商品名称搜索',
              clearable: true,
            },
            {
              type: 'select',
              name: 'status',
              label: '订单状态',
              placeholder: '请选择订单状态',
              clearable: true,
              multiple: true,
              options: [
                { label: '待支付', value: 1 },
                { label: '已支付', value: 2 },
                { label: '处理中', value: 3 },
                { label: '已完成', value: 4 },
                { label: '已取消', value: 5 },
              ],
            },
            {
              type: 'input-date-range',
              name: 'date_range',
              label: '创建时间',
              format: 'YYYY-MM-DD',
              placeholder: '请选择创建时间范围',
              clearable: true,
              onEvent: {
                change: {
                  actions: [
                    {
                      actionType: 'setValue',
                      componentId: 'start_date',
                      args: {
                        value: '${event.data.value ? event.data.value.split(",")[0] : ""}',
                      },
                    },
                    {
                      actionType: 'setValue', 
                      componentId: 'end_date',
                      args: {
                        value: '${event.data.value ? event.data.value.split(",")[1] : ""}',
                      },
                    },
                  ],
                },
              },
            },
            {
              type: 'hidden',
              name: 'start_date',
              id: 'start_date',
            },
            {
              type: 'hidden',
              name: 'end_date',
              id: 'end_date',
            },
          ],
        },
        headerToolbar: [
          'filter-toggler',
          {
            type: 'reload',
            label: '刷新',
          },
          {
            type: 'export-excel',
            label: '导出Excel',
            api: '/api/v1/orders/export',
          },
        ],
        sortable: true,
        columns: [
          {
            name: 'order_number',
            label: '订单号',
            type: 'text',
            copyable: true,
            sortable: false,
          },
          {
            name: 'status',
            label: '当前状态',
            type: 'mapping',
            sortable: false,
            map: {
              1: '<span class="label label-warning">待支付</span>',
              2: '<span class="label label-info">已支付</span>',
              3: '<span class="label label-primary">处理中</span>',
              4: '<span class="label label-success">已完成</span>',
              5: '<span class="label label-default">已取消</span>',
            },
          },
          {
            name: 'latest_status_change',
            label: '最新状态变更',
            type: 'container',
            sortable: false,
            body: [
              {
                type: 'tpl',
                tpl: '<div class="text-xs text-gray-500">${latest_status_change.reason || "系统自动"}</div><div class="text-xs">${latest_status_change.created_at | date:MM-DD HH:mm}</div>',
              },
            ],
          },
          {
            name: 'merchant_name',
            label: '商户',
            type: 'text',
            sortable: false,
          },
          {
            name: 'item_count',
            label: '商品数量',
            type: 'text',
            sortable: false,
            tpl: '${item_count} 件',
          },
          {
            name: 'total_amount',
            label: '订单金额',
            type: 'text',
            sortable: true,
            tpl: '¥${total_amount}',
          },
          {
            name: 'total_rights_cost',
            label: '权益成本',
            type: 'text',
            sortable: false,
            tpl: '${total_rights_cost} 积分',
          },
          {
            name: 'created_at',
            label: '创建时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm',
            sortable: true,
          },
          {
            name: 'updated_at',
            label: '最后更新',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm',
            sortable: true,
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
                          size: 'xl',
                          body: {
                            type: 'service',
                            api: '/api/v1/orders/${id}/detail',
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
                                    label: '当前状态',
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
                                    label: '商户名称',
                                    content: '${merchant_name || "未知商户"}',
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
                                    label: '最后更新',
                                    content: '${updated_at | date:YYYY-MM-DD HH:mm:ss}',
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
                                  {
                                    label: '小计金额',
                                    type: 'text',
                                    tpl: '¥${price * quantity}',
                                  },
                                ],
                              },
                              {
                                type: 'divider',
                              },
                              {
                                type: 'panel',
                                title: '状态历史记录',
                                body: {
                                  type: 'table',
                                  source: '${status_history}',
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
                                      label: '操作者类型',
                                      type: 'mapping',
                                      map: {
                                        customer: '客户',
                                        merchant: '商户',
                                        system: '系统',
                                        admin: '管理员',
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
                size: 'sm',
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
                                label: '支付成功返回页面（可选）',
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
                                  url: '/api/v1/orders/${id}/pay',
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
                size: 'sm',
                visibleOn: '${status == 1}',
                confirmText: '确认取消这个订单吗？取消后将无法恢复。',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'ajax',
                        api: {
                          method: 'put',
                          url: '/api/v1/orders/${id}/cancel',
                        },
                        messages: {
                          success: '订单取消成功',
                          failed: '订单取消失败：${msg}',
                        },
                      },
                      {
                        actionType: 'reload',
                        target: 'crud',
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
        interval: 30000, // 每30秒自动刷新一次
        silentPolling: true, // 静默轮询，不显示加载状态
        stopAutoRefreshWhen: 'this.total === 0', // 没有数据时停止刷新
      },
    ],
    initApi: false, // 禁用初始加载，让crud组件自己控制
  };

  return <AmisRenderer schema={schema} />;
};

export default OrderListPage;
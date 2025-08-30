import React, { useState, useEffect } from 'react';
import { AmisRenderer } from '../../../components/ui/AmisRenderer';
import { useOrderStatusStore } from '../../../stores/orderStatusStore';
import { useAuthStore } from '../../../stores/authStore';
import OrderPermissionWrapper, { ORDER_PERMISSION_GROUPS, TenantIsolationWrapper } from '../../../components/orders/OrderPermissionWrapper';
import type { SchemaNode } from 'amis';

const OrderManagePage: React.FC = () => {
  const { updateStats } = useOrderStatusStore();
  const { user } = useAuthStore();
  const [selectedRows, setSelectedRows] = useState<any[]>([]);
  
  // 从认证上下文获取商户ID
  const merchantId = user?.merchant_id || user?.id;
  const tenantId = user?.tenant_id;

  const schema: SchemaNode = {
    type: 'page',
    title: '订单管理',
    subTitle: '管理您店铺的所有订单，支持批量操作和状态变更',
    body: [
      // 统计面板
      {
        type: 'grid',
        columns: [
          {
            md: 3,
            body: {
              type: 'service',
              api: `/api/v1/orders/stats?merchant_id=${merchantId}&tenant_id=${tenantId}`,
              body: {
                type: 'card',
                className: 'bg-blue-50 border-blue-200',
                header: {
                  title: '订单总数',
                  subTitle: '全部订单',
                },
                body: [
                  {
                    type: 'tpl',
                    tpl: '<div class="text-3xl font-bold text-blue-600">${total}</div>',
                  },
                ],
              },
            },
          },
          {
            md: 3,
            body: {
              type: 'service',
              api: `/api/v1/orders/stats?merchant_id=${merchantId}&tenant_id=${tenantId}`,
              body: {
                type: 'card',
                className: 'bg-orange-50 border-orange-200',
                header: {
                  title: '待处理',
                  subTitle: '待支付+已支付',
                },
                body: [
                  {
                    type: 'tpl',
                    tpl: '<div class="text-3xl font-bold text-orange-600">${by_status.pending + by_status.paid}</div>',
                  },
                ],
              },
            },
          },
          {
            md: 3,
            body: {
              type: 'service',
              api: `/api/v1/orders/stats?merchant_id=${merchantId}&tenant_id=${tenantId}`,
              body: {
                type: 'card',
                className: 'bg-purple-50 border-purple-200',
                header: {
                  title: '处理中',
                  subTitle: '正在处理',
                },
                body: [
                  {
                    type: 'tpl',
                    tpl: '<div class="text-3xl font-bold text-purple-600">${by_status.processing}</div>',
                  },
                ],
              },
            },
          },
          {
            md: 3,
            body: {
              type: 'service',
              api: `/api/v1/orders/stats?merchant_id=${merchantId}&tenant_id=${tenantId}`,
              body: {
                type: 'card',
                className: 'bg-green-50 border-green-200',
                header: {
                  title: '已完成',
                  subTitle: '已完成订单',
                },
                body: [
                  {
                    type: 'tpl',
                    tpl: '<div class="text-3xl font-bold text-green-600">${by_status.completed}</div>',
                  },
                ],
              },
            },
          },
        ],
      },
      {
        type: 'divider',
      },
      // 订单列表
      {
        type: 'crud',
        name: 'ordersCrud',
        api: {
          method: 'get',
          url: '/api/v1/orders/query',
          data: {
            merchant_id: merchantId,
            tenant_id: tenantId,
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
        // 筛选条件
        filter: {
          title: '订单筛选',
          body: [
            {
              type: 'input-text',
              name: 'search_keyword',
              label: '搜索',
              placeholder: '请输入订单号或客户信息搜索',
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
        // 工具栏
        headerToolbar: [
          'filter-toggler',
          {
            type: 'reload',
            label: '刷新',
          },
          {
            type: 'export-excel',
            label: '导出Excel',
            api: `/api/v1/orders/export?merchant_id=${merchantId}&tenant_id=${tenantId}`,
          },
          {
            type: 'button',
            label: '批量处理',
            level: 'primary',
            visibleOn: '${selectedItems.length > 0}',
            onEvent: {
              click: {
                actions: [
                  {
                    actionType: 'dialog',
                    dialog: {
                      title: '批量处理订单',
                      body: {
                        type: 'form',
                        body: [
                          {
                            type: 'alert',
                            level: 'info',
                            body: '已选择 ${selectedItems.length} 个订单进行批量处理',
                          },
                          {
                            type: 'select',
                            name: 'target_status',
                            label: '目标状态',
                            required: true,
                            options: [
                              { label: '标记为处理中', value: 3 },
                              { label: '标记为已完成', value: 4 },
                              { label: '取消订单', value: 5 },
                            ],
                          },
                          {
                            type: 'textarea',
                            name: 'reason',
                            label: '处理原因',
                            placeholder: '请输入批量处理的原因（可选）',
                            maxLength: 200,
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
                            label: '确认处理',
                            level: 'primary',
                            api: {
                              method: 'post',
                              url: '/api/v1/orders/batch-update-status',
                              data: {
                                order_ids: '${selectedItems|pick:id}',
                                target_status: '${target_status}',
                                reason: '${reason}',
                              },
                            },
                            confirmText: '确认要批量处理这些订单吗？',
                          },
                        ],
                      },
                    },
                  },
                ],
              },
            },
          },
        ],
        // 表格列配置
        sortable: true,
        selectable: true,
        multiple: true,
        columns: [
          {
            name: 'order_number',
            label: '订单号',
            type: 'text',
            copyable: true,
            sortable: false,
            width: 150,
          },
          {
            name: 'customer_name',
            label: '客户',
            type: 'text',
            sortable: false,
            width: 100,
          },
          {
            name: 'status',
            label: '状态',
            type: 'mapping',
            sortable: false,
            width: 100,
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
            label: '最新变更',
            type: 'container',
            sortable: false,
            width: 120,
            body: [
              {
                type: 'tpl',
                tpl: '<div class="text-xs text-gray-500">${latest_status_change.reason || "系统自动"}</div><div class="text-xs">${latest_status_change.created_at | date:MM-DD HH:mm}</div>',
              },
            ],
          },
          {
            name: 'item_count',
            label: '商品数量',
            type: 'text',
            sortable: false,
            width: 80,
            tpl: '${item_count} 件',
          },
          {
            name: 'total_amount',
            label: '订单金额',
            type: 'text',
            sortable: true,
            width: 100,
            tpl: '¥${total_amount}',
          },
          {
            name: 'total_rights_cost',
            label: '权益成本',
            type: 'text',
            sortable: false,
            width: 80,
            tpl: '${total_rights_cost} 积分',
          },
          {
            name: 'created_at',
            label: '创建时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm',
            sortable: true,
            width: 130,
          },
          {
            type: 'operation',
            label: '操作',
            width: 200,
            buttons: [
              {
                type: 'button',
                label: '查看详情',
                level: 'link',
                size: 'sm',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'drawer',
                        drawer: {
                          title: '订单详情 - ${order_number}',
                          size: 'xl',
                          body: {
                            type: 'service',
                            api: '/api/v1/orders/${id}/detail',
                            body: [
                              // 基本信息
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
                                    label: '客户名称',
                                    content: '${customer_name}',
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
                              // 商品明细
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
                              // 状态历史
                              {
                                type: 'panel',
                                title: '状态历史',
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
                                      label: '操作者',
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
                label: '开始处理',
                level: 'primary',
                size: 'sm',
                visibleOn: '${status == 2}',
                confirmText: '确认开始处理这个订单吗？',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'ajax',
                        api: {
                          method: 'put',
                          url: '/api/v1/orders/${id}/status',
                          data: {
                            status: 3,
                            reason: '商户开始处理订单',
                          },
                        },
                        messages: {
                          success: '订单状态更新成功',
                          failed: '状态更新失败：${msg}',
                        },
                      },
                      {
                        actionType: 'reload',
                        target: 'ordersCrud',
                      },
                    ],
                  },
                },
              },
              {
                type: 'button',
                label: '标记完成',
                level: 'success',
                size: 'sm',
                visibleOn: '${status == 3}',
                confirmText: '确认标记这个订单为已完成吗？',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'ajax',
                        api: {
                          method: 'put',
                          url: '/api/v1/orders/${id}/status',
                          data: {
                            status: 4,
                            reason: '商户确认订单完成',
                          },
                        },
                        messages: {
                          success: '订单已标记为完成',
                          failed: '状态更新失败：${msg}',
                        },
                      },
                      {
                        actionType: 'reload',
                        target: 'ordersCrud',
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
                visibleOn: '${status == 1 || status == 2}',
                onEvent: {
                  click: {
                    actions: [
                      {
                        actionType: 'dialog',
                        dialog: {
                          title: '取消订单',
                          body: {
                            type: 'form',
                            body: [
                              {
                                type: 'alert',
                                level: 'warning',
                                body: '请注意：取消订单后将无法恢复，已支付的款项需要手动退款处理。',
                              },
                              {
                                type: 'textarea',
                                name: 'reason',
                                label: '取消原因',
                                required: true,
                                placeholder: '请输入取消订单的原因',
                                maxLength: 200,
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
                                label: '确认取消订单',
                                level: 'danger',
                                api: {
                                  method: 'put',
                                  url: '/api/v1/orders/${id}/status',
                                  data: {
                                    status: 5,
                                    reason: '${reason}',
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
            ],
          },
        ],
        // 分页配置
        pagination: {
          enable: true,
          layout: ['pager', 'perpage', 'total'],
        },
        // 实时刷新
        interval: 30000,
        silentPolling: true,
        stopAutoRefreshWhen: 'this.total === 0',
      },
    ],
    initApi: false,
  };

  return (
    <OrderPermissionWrapper 
      permissions={ORDER_PERMISSION_GROUPS.MERCHANT_ADVANCED}
      fallback={
        <div className="text-center py-8">
          <h3 className="text-lg font-medium text-gray-900 mb-2">访问受限</h3>
          <p className="text-gray-600">您需要商户订单管理权限才能访问此页面。</p>
        </div>
      }
    >
      <TenantIsolationWrapper 
        resourceTenantId={tenantId}
        fallback={
          <div className="text-center py-8">
            <h3 className="text-lg font-medium text-gray-900 mb-2">数据隔离</h3>
            <p className="text-gray-600">多租户数据隔离已生效，无法访问其他租户数据。</p>
          </div>
        }
      >
        <AmisRenderer schema={schema} />
      </TenantIsolationWrapper>
    </OrderPermissionWrapper>
  );
};

export default OrderManagePage;
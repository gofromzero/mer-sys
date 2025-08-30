import React, { useState } from 'react';
import { AmisRenderer } from '../ui/AmisRenderer';
import type { SchemaNode } from 'amis';

interface OrderBatchActionsProps {
  /** 选中的订单列表 */
  selectedOrders: Array<{
    id: number;
    order_number: string;
    status: number;
    total_amount: number;
    customer_name: string;
  }>;
  /** 批量操作完成回调 */
  onBatchComplete?: (result: { success: number; failed: number }) => void;
  /** 重新加载订单列表回调 */
  onReload?: () => void;
}

const OrderBatchActions: React.FC<OrderBatchActionsProps> = ({
  selectedOrders,
  onBatchComplete,
  onReload,
}) => {
  const [isProcessing, setIsProcessing] = useState(false);

  if (selectedOrders.length === 0) {
    return null;
  }

  // 统计各状态订单数量
  const statusStats = selectedOrders.reduce((acc, order) => {
    acc[order.status] = (acc[order.status] || 0) + 1;
    return acc;
  }, {} as Record<number, number>);

  // 获取状态名称
  const getStatusName = (status: number): string => {
    const statusMap: Record<number, string> = {
      1: '待支付',
      2: '已支付',
      3: '处理中',
      4: '已完成',
      5: '已取消',
    };
    return statusMap[status] || '未知状态';
  };

  // 检查是否可以批量处理
  const canBatchProcess = selectedOrders.some(order => order.status === 2); // 已支付
  const canBatchComplete = selectedOrders.some(order => order.status === 3); // 处理中
  const canBatchCancel = selectedOrders.some(order => order.status === 1 || order.status === 2); // 待支付或已支付

  const schema: SchemaNode = {
    type: 'form',
    title: '批量操作订单',
    className: 'bg-white rounded-lg shadow-sm border p-4 mb-4',
    body: [
      // 选中订单统计
      {
        type: 'alert',
        level: 'info',
        body: `已选择 ${selectedOrders.length} 个订单：${Object.entries(statusStats).map(([status, count]) => 
          `${getStatusName(parseInt(status))} ${count} 个`
        ).join('，')}`,
      },
      // 批量操作按钮组
      {
        type: 'button-group',
        buttons: [
          // 批量开始处理
          {
            type: 'button',
            label: `批量开始处理 (${statusStats[2] || 0} 个)`,
            level: 'primary',
            disabled: !canBatchProcess,
            disabledTip: '没有可以开始处理的订单（需要已支付状态）',
            confirmText: `确认要将 ${statusStats[2] || 0} 个已支付订单标记为处理中吗？`,
            onEvent: {
              click: {
                actions: [
                  {
                    actionType: 'ajax',
                    api: {
                      method: 'post',
                      url: '/api/v1/orders/batch-update-status',
                      data: {
                        order_ids: selectedOrders.filter(o => o.status === 2).map(o => o.id),
                        target_status: 3,
                        reason: '商户批量开始处理订单',
                      },
                    },
                    messages: {
                      success: '批量处理操作成功',
                      failed: '批量处理操作失败：${msg}',
                    },
                  },
                  {
                    actionType: 'custom',
                    script: `
                      if (event.data && onReload) {
                        onReload();
                      }
                    `,
                  },
                ],
              },
            },
          },
          // 批量标记完成
          {
            type: 'button',
            label: `批量标记完成 (${statusStats[3] || 0} 个)`,
            level: 'success',
            disabled: !canBatchComplete,
            disabledTip: '没有可以标记完成的订单（需要处理中状态）',
            confirmText: `确认要将 ${statusStats[3] || 0} 个处理中订单标记为已完成吗？`,
            onEvent: {
              click: {
                actions: [
                  {
                    actionType: 'ajax',
                    api: {
                      method: 'post',
                      url: '/api/v1/orders/batch-update-status',
                      data: {
                        order_ids: selectedOrders.filter(o => o.status === 3).map(o => o.id),
                        target_status: 4,
                        reason: '商户批量确认订单完成',
                      },
                    },
                    messages: {
                      success: '批量完成操作成功',
                      failed: '批量完成操作失败：${msg}',
                    },
                  },
                  {
                    actionType: 'custom',
                    script: `
                      if (event.data && onReload) {
                        onReload();
                      }
                    `,
                  },
                ],
              },
            },
          },
          // 批量取消
          {
            type: 'button',
            label: `批量取消订单 (${(statusStats[1] || 0) + (statusStats[2] || 0)} 个)`,
            level: 'danger',
            disabled: !canBatchCancel,
            disabledTip: '没有可以取消的订单（需要待支付或已支付状态）',
            onEvent: {
              click: {
                actions: [
                  {
                    actionType: 'dialog',
                    dialog: {
                      title: '批量取消订单',
                      size: 'md',
                      body: {
                        type: 'form',
                        body: [
                          {
                            type: 'alert',
                            level: 'warning',
                            body: `即将取消 ${(statusStats[1] || 0) + (statusStats[2] || 0)} 个订单，请注意：
                            <ul class="mt-2 list-disc list-inside">
                              <li>取消后订单状态将无法恢复</li>
                              <li>已支付的订单需要手动处理退款</li>
                              <li>库存将自动释放</li>
                            </ul>`,
                          },
                          {
                            type: 'textarea',
                            name: 'cancel_reason',
                            label: '取消原因',
                            required: true,
                            placeholder: '请输入批量取消订单的原因',
                            maxLength: 200,
                            showCounter: true,
                          },
                          {
                            type: 'switch',
                            name: 'auto_refund',
                            label: '自动退款',
                            option: '对已支付订单启用自动退款（如果支持）',
                            value: false,
                          },
                        ],
                        actions: [
                          {
                            type: 'button',
                            label: '取消操作',
                            actionType: 'cancel',
                          },
                          {
                            type: 'submit',
                            label: `确认取消 ${(statusStats[1] || 0) + (statusStats[2] || 0)} 个订单`,
                            level: 'danger',
                            api: {
                              method: 'post',
                              url: '/api/v1/orders/batch-update-status',
                              data: {
                                order_ids: selectedOrders.filter(o => o.status === 1 || o.status === 2).map(o => o.id),
                                target_status: 5,
                                reason: '${cancel_reason}',
                                metadata: {
                                  auto_refund: '${auto_refund}',
                                },
                              },
                            },
                            messages: {
                              success: '批量取消操作成功',
                              failed: '批量取消操作失败：${msg}',
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
      // 选中订单列表
      {
        type: 'collapse',
        header: '查看选中的订单明细',
        body: [
          {
            type: 'table',
            source: selectedOrders,
            columns: [
              {
                name: 'order_number',
                label: '订单号',
                type: 'text',
                width: 150,
              },
              {
                name: 'customer_name',
                label: '客户',
                type: 'text',
                width: 100,
              },
              {
                name: 'status',
                label: '当前状态',
                type: 'mapping',
                width: 80,
                map: {
                  1: '<span class="label label-warning">待支付</span>',
                  2: '<span class="label label-info">已支付</span>',
                  3: '<span class="label label-primary">处理中</span>',
                  4: '<span class="label label-success">已完成</span>',
                  5: '<span class="label label-default">已取消</span>',
                },
              },
              {
                name: 'total_amount',
                label: '金额',
                type: 'text',
                width: 80,
                tpl: '¥${total_amount}',
              },
            ],
            affixHeader: false,
            resizable: false,
          },
        ],
      },
    ],
  };

  return (
    <div className="order-batch-actions">
      <AmisRenderer 
        schema={schema}
        props={{
          onReload,
          onBatchComplete,
        }}
      />
    </div>
  );
};

export default OrderBatchActions;
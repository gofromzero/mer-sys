import React from 'react';
import { AmisRenderer } from '../ui/AmisRenderer';
import type { SchemaNode } from 'amis';

interface OrderStatusHistoryProps {
  /** 订单ID */
  orderId: number;
  /** 是否显示为时间轴形式 */
  timeline?: boolean;
  /** 是否显示操作者信息 */
  showOperator?: boolean;
  /** 最大显示条数 */
  maxItems?: number;
}

const OrderStatusHistory: React.FC<OrderStatusHistoryProps> = ({
  orderId,
  timeline = false,
  showOperator = true,
  maxItems,
}) => {
  const schema: SchemaNode = timeline ? {
    // 时间轴展示
    type: 'service',
    api: `/api/v1/orders/${orderId}/status-history`,
    body: {
      type: 'timeline',
      source: '${items}',
      itemTitleTpl: '${from_status_name || "创建"} → ${to_status_name}',
      itemTpl: `
        <div class="order-status-timeline-item">
          <div class="status-change">
            <span class="badge badge-\${to_status === 1 ? 'warning' : to_status === 2 ? 'info' : to_status === 3 ? 'primary' : to_status === 4 ? 'success' : 'default'}">
              \${to_status_name}
            </span>
          </div>
          <div class="change-reason text-muted">\${reason || '系统自动'}</div>
          ${showOperator ? '<div class="operator text-sm text-gray-500">操作者：${operator_type_name}</div>' : ''}
          <div class="change-time text-sm text-gray-400">\${created_at | date:MM-DD HH:mm}</div>
        </div>
      `,
      timeFormat: 'MM-DD HH:mm',
      reverse: true,
    },
  } : {
    // 表格展示
    type: 'service',
    api: `/api/v1/orders/${orderId}/status-history${maxItems ? `?limit=${maxItems}` : ''}`,
    body: {
      type: 'table',
      source: '${items}',
      placeholder: '暂无状态变更记录',
      showHeader: true,
      columns: [
        {
          name: 'from_status',
          label: '原状态',
          type: 'mapping',
          width: 80,
          map: {
            0: '新建',
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
          name: 'reason',
          label: '变更原因',
          type: 'text',
          placeholder: '系统自动',
        },
        ...(showOperator ? [{
          name: 'operator_type',
          label: '操作者',
          type: 'mapping',
          width: 80,
          map: {
            customer: '<span class="label label-info">客户</span>',
            merchant: '<span class="label label-success">商户</span>',
            system: '<span class="label label-default">系统</span>',
            admin: '<span class="label label-primary">管理员</span>',
          },
        }] : []),
        {
          name: 'created_at',
          label: '变更时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss',
          width: 140,
        },
      ],
    },
  };

  return (
    <div className="order-status-history">
      <AmisRenderer schema={schema} />
      
      <style jsx>{`
        .order-status-history {
          margin-top: 16px;
        }
        
        .order-status-timeline-item {
          padding: 8px 0;
        }
        
        .status-change {
          margin-bottom: 4px;
        }
        
        .change-reason {
          margin-bottom: 2px;
          font-size: 13px;
        }
        
        .operator {
          margin-bottom: 2px;
        }
        
        .change-time {
          font-size: 12px;
        }
        
        .badge {
          display: inline-block;
          padding: 2px 6px;
          font-size: 11px;
          font-weight: 500;
          line-height: 1;
          color: #fff;
          text-align: center;
          white-space: nowrap;
          vertical-align: baseline;
          border-radius: 3px;
        }
        
        .badge-warning { background-color: #f0ad4e; }
        .badge-info { background-color: #5bc0de; }
        .badge-primary { background-color: #337ab7; }
        .badge-success { background-color: #5cb85c; }
        .badge-default { background-color: #777; }
      `}</style>
    </div>
  );
};

export default OrderStatusHistory;
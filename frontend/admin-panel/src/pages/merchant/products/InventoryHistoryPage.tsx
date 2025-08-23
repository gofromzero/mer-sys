// 库存变更历史页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const InventoryHistoryPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '库存变更历史',
    body: [
      {
        type: 'form',
        title: '筛选条件',
        mode: 'inline',
        wrapWithPanel: false,
        className: 'mb-4',
        controls: [
          {
            type: 'text',
            name: 'product_name',
            label: '商品名称',
            placeholder: '请输入商品名称',
            className: 'w-48'
          },
          {
            type: 'select',
            name: 'change_type',
            label: '变更类型',
            placeholder: '请选择变更类型',
            className: 'w-40',
            options: [
              { label: '全部', value: '' },
              { label: '采购入库', value: 'purchase' },
              { label: '销售出库', value: 'sale' },
              { label: '盘点调整', value: 'adjustment' },
              { label: '调拨', value: 'transfer' },
              { label: '损耗', value: 'damage' },
              { label: '预留锁定', value: 'reservation' },
              { label: '释放预留', value: 'release' }
            ]
          },
          {
            type: 'input-date-range',
            name: 'date_range',
            label: '时间范围',
            className: 'w-64',
            format: 'YYYY-MM-DD'
          },
          {
            type: 'submit',
            label: '查询',
            level: 'primary'
          },
          {
            type: 'reset',
            label: '重置'
          }
        ],
        actions: []
      },
      {
        type: 'crud',
        api: {
          method: 'get',
          url: '/api/v1/inventory/records',
          data: {
            '&': '$$'
          }
        },
        defaultParams: {
          page: 1,
          page_size: 20
        },
        syncLocation: false,
        headerToolbar: [
          'statistics'
        ],
        footerToolbar: [
          'switch-per-page',
          'pagination'
        ],
        columns: [
          {
            name: 'product_name',
            label: '商品名称',
            type: 'text',
            searchable: false
          },
          {
            name: 'change_type',
            label: '变更类型',
            type: 'mapping',
            map: {
              'purchase': '<span class="label label-success">采购入库</span>',
              'sale': '<span class="label label-info">销售出库</span>',
              'adjustment': '<span class="label label-warning">盘点调整</span>',
              'transfer': '<span class="label label-primary">调拨</span>',
              'damage': '<span class="label label-danger">损耗</span>',
              'reservation': '<span class="label label-secondary">预留锁定</span>',
              'release': '<span class="label label-default">释放预留</span>'
            }
          },
          {
            name: 'quantity_before',
            label: '变更前',
            type: 'text',
            className: 'text-right'
          },
          {
            name: 'quantity_changed',
            label: '变更数量',
            type: 'tpl',
            tpl: '<%= data.quantity_changed > 0 ? "+" + data.quantity_changed : data.quantity_changed %>',
            className: 'text-right',
            classNameExpr: '<%= data.quantity_changed > 0 ? "text-success" : "text-danger" %>'
          },
          {
            name: 'quantity_after',
            label: '变更后',
            type: 'text',
            className: 'text-right'
          },
          {
            name: 'reason',
            label: '变更原因',
            type: 'text'
          },
          {
            name: 'reference_id',
            label: '关联单号',
            type: 'text'
          },
          {
            name: 'operated_by_name',
            label: '操作人',
            type: 'text'
          },
          {
            name: 'created_at',
            label: '变更时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm:ss'
          }
        ]
      }
    ]
  };

  return <AmisRenderer schema={schema} />;
};

export default InventoryHistoryPage;
// 批量库存操作页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const InventoryBatchPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '批量库存操作',
    body: [
      {
        type: 'tabs',
        tabs: [
          {
            title: '批量调整',
            body: [
              // 商品选择器
              {
                type: 'form',
                title: '选择商品',
                mode: 'normal',
                wrapWithPanel: false,
                className: 'mb-4',
                controls: [
                  {
                    type: 'picker',
                    name: 'selected_products',
                    label: '选择商品',
                    required: true,
                    multiple: true,
                    modalMode: true,
                    pickerSchema: {
                      type: 'crud',
                      api: '/api/v1/products',
                      columns: [
                        {
                          name: 'id',
                          label: 'ID',
                          type: 'text'
                        },
                        {
                          name: 'name',
                          label: '商品名称',
                          type: 'text'
                        },
                        {
                          name: 'inventory.stock_quantity',
                          label: '当前库存',
                          type: 'text'
                        }
                      ]
                    },
                    labelField: 'name',
                    valueField: 'id',
                    validationErrors: {
                      required: '请至少选择一个商品'
                    }
                  }
                ],
                actions: []
              },

              // 批量操作表单
              {
                type: 'form',
                title: '批量操作设置',
                api: {
                  method: 'post',
                  url: '/api/v1/products/inventory/batch-adjust',
                  messages: {
                    success: '批量调整成功',
                    failed: '批量调整失败'
                  }
                },
                controls: [
                  {
                    type: 'select',
                    name: 'adjustment_type',
                    label: '调整类型',
                    required: true,
                    options: [
                      { label: '增加库存', value: 'increase' },
                      { label: '减少库存', value: 'decrease' },
                      { label: '设置库存', value: 'set' }
                    ],
                    validationErrors: {
                      required: '请选择调整类型'
                    }
                  },
                  {
                    type: 'number',
                    name: 'quantity',
                    label: '数量',
                    required: true,
                    min: 1,
                    validationErrors: {
                      required: '请输入数量',
                      min: '数量必须大于0'
                    }
                  },
                  {
                    type: 'textarea',
                    name: 'reason',
                    label: '调整原因',
                    required: true,
                    rows: 3,
                    placeholder: '请输入批量调整的原因...',
                    validationErrors: {
                      required: '请输入调整原因'
                    }
                  },
                  {
                    type: 'text',
                    name: 'reference_id',
                    label: '参考单号',
                    placeholder: '相关的订单号、入库单号等（可选）'
                  }
                ],
                data: {
                  adjustments: '${ARRAYMAP(selected_products, item => ({product_id: item.id, adjustment_type: adjustment_type, quantity: quantity, reason: reason, reference_id: reference_id}))}'
                }
              }
            ]
          },
          {
            title: '批量查询',
            body: [
              {
                type: 'form',
                title: '批量库存查询',
                mode: 'normal',
                wrapWithPanel: false,
                controls: [
                  {
                    type: 'textarea',
                    name: 'product_ids_text',
                    label: '商品ID列表',
                    placeholder: '请输入商品ID，每行一个或用逗号分隔',
                    rows: 5,
                    required: true
                  },
                  {
                    type: 'submit',
                    label: '查询库存',
                    level: 'primary'
                  }
                ],
                api: {
                  method: 'post',
                  url: '/api/v1/inventory/batch-query',
                  data: {
                    product_ids: '${SPLIT(REPLACE(product_ids_text, /[\\r\\n]+/g, ","), ",") | ARRAYMAP(TRIM) | ARRAYMAP(INT)}'
                  }
                },
                actions: []
              },

              // 查询结果显示
              {
                type: 'service',
                api: '/api/v1/inventory/batch-query',
                body: {
                  type: 'table',
                  source: '${data}',
                  columns: [
                    {
                      name: 'product_id',
                      label: '商品ID',
                      type: 'text'
                    },
                    {
                      name: 'product_name',
                      label: '商品名称',
                      type: 'text'
                    },
                    {
                      name: 'available_stock',
                      label: '可用库存',
                      type: 'text',
                      className: 'text-right'
                    },
                    {
                      name: 'reserved_stock',
                      label: '预留库存',
                      type: 'text',
                      className: 'text-right'
                    },
                    {
                      name: 'inventory_info.stock_quantity',
                      label: '总库存',
                      type: 'text',
                      className: 'text-right'
                    },
                    {
                      name: 'is_low_stock',
                      label: '库存状态',
                      type: 'mapping',
                      map: {
                        true: '<span class="label label-warning">库存不足</span>',
                        false: '<span class="label label-success">库存正常</span>'
                      }
                    },
                    {
                      type: 'operation',
                      label: '操作',
                      buttons: [
                        {
                          type: 'button',
                          actionType: 'dialog',
                          label: '调整',
                          level: 'primary',
                          size: 'sm',
                          dialog: {
                            title: '调整库存 - ${product_name}',
                            size: 'md',
                            body: {
                              type: 'form',
                              api: {
                                method: 'post',
                                url: '/api/v1/products/${product_id}/inventory/adjust',
                                messages: {
                                  success: '库存调整成功',
                                  failed: '库存调整失败'
                                }
                              },
                              controls: [
                                {
                                  type: 'static',
                                  name: 'product_name',
                                  label: '商品名称'
                                },
                                {
                                  type: 'static',
                                  name: 'available_stock',
                                  label: '当前可用库存'
                                },
                                {
                                  type: 'select',
                                  name: 'adjustment_type',
                                  label: '调整类型',
                                  required: true,
                                  options: [
                                    { label: '增加库存', value: 'increase' },
                                    { label: '减少库存', value: 'decrease' },
                                    { label: '设置库存', value: 'set' }
                                  ]
                                },
                                {
                                  type: 'number',
                                  name: 'quantity',
                                  label: '数量',
                                  required: true,
                                  min: 1
                                },
                                {
                                  type: 'textarea',
                                  name: 'reason',
                                  label: '调整原因',
                                  required: true,
                                  rows: 3
                                }
                              ]
                            }
                          }
                        }
                      ]
                    }
                  ]
                }
              }
            ]
          },
          {
            title: '导入导出',
            body: [
              {
                type: 'form',
                title: '批量导入库存',
                mode: 'normal',
                wrapWithPanel: false,
                className: 'mb-4',
                controls: [
                  {
                    type: 'input-file',
                    name: 'inventory_file',
                    label: '选择文件',
                    accept: '.xlsx,.xls,.csv',
                    receiver: '/api/v1/inventory/import',
                    autoUpload: false,
                    required: true,
                    description: '支持Excel和CSV格式，请下载模板文件'
                  },
                  {
                    type: 'select',
                    name: 'import_mode',
                    label: '导入模式',
                    value: 'update',
                    options: [
                      { label: '更新现有库存', value: 'update' },
                      { label: '覆盖现有库存', value: 'overwrite' }
                    ]
                  },
                  {
                    type: 'switch',
                    name: 'validate_only',
                    label: '仅验证不导入',
                    description: '开启后只检查文件格式，不实际导入数据'
                  },
                  {
                    type: 'submit',
                    label: '开始导入',
                    level: 'primary'
                  }
                ],
                actions: []
              },

              // 下载模板和导出功能
              {
                type: 'panel',
                title: '模板下载与数据导出',
                body: [
                  {
                    type: 'button',
                    label: '下载导入模板',
                    level: 'info',
                    actionType: 'download',
                    api: '/api/v1/inventory/template/download'
                  },
                  {
                    type: 'divider'
                  },
                  {
                    type: 'form',
                    title: '导出库存数据',
                    mode: 'inline',
                    wrapWithPanel: false,
                    controls: [
                      {
                        type: 'select',
                        name: 'export_format',
                        label: '导出格式',
                        value: 'xlsx',
                        options: [
                          { label: 'Excel文件 (.xlsx)', value: 'xlsx' },
                          { label: 'CSV文件 (.csv)', value: 'csv' }
                        ]
                      },
                      {
                        type: 'select',
                        name: 'export_scope',
                        label: '导出范围',
                        value: 'all',
                        options: [
                          { label: '全部商品', value: 'all' },
                          { label: '有库存商品', value: 'in_stock' },
                          { label: '低库存商品', value: 'low_stock' },
                          { label: '缺货商品', value: 'out_of_stock' }
                        ]
                      },
                      {
                        type: 'submit',
                        label: '导出数据',
                        level: 'success',
                        actionType: 'download',
                        api: {
                          method: 'post',
                          url: '/api/v1/inventory/export'
                        }
                      }
                    ],
                    actions: []
                  }
                ]
              }
            ]
          }
        ]
      }
    ]
  };

  return <AmisRenderer schema={schema} />;
};

export default InventoryBatchPage;
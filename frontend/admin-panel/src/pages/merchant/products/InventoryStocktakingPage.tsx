// 库存盘点管理页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const InventoryStocktakingPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '库存盘点管理',
    body: [
      {
        type: 'crud',
        api: {
          method: 'get',
          url: '/api/v1/inventory/stocktaking'
        },
        syncLocation: false,
        headerToolbar: [
          {
            type: 'button',
            actionType: 'dialog',
            label: '新建盘点任务',
            level: 'primary',
            dialog: {
              title: '创建库存盘点任务',
              size: 'lg',
              body: {
                type: 'form',
                api: {
                  method: 'post',
                  url: '/api/v1/inventory/stocktaking/start',
                  messages: {
                    success: '盘点任务创建成功',
                    failed: '盘点任务创建失败'
                  }
                },
                controls: [
                  {
                    type: 'text',
                    name: 'name',
                    label: '盘点名称',
                    required: true,
                    placeholder: '请输入盘点任务名称',
                    validationErrors: {
                      required: '请输入盘点名称'
                    }
                  },
                  {
                    type: 'textarea',
                    name: 'description',
                    label: '盘点说明',
                    placeholder: '请输入盘点任务的详细说明',
                    rows: 3
                  },
                  {
                    type: 'switch',
                    name: 'full_stocktaking',
                    label: '全量盘点',
                    value: true,
                    description: '开启后盘点所有商品，关闭后可选择特定商品'
                  },
                  {
                    type: 'picker',
                    name: 'product_ids',
                    label: '选择商品',
                    multiple: true,
                    modalMode: true,
                    hiddenOn: '${full_stocktaking}',
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
                    valueField: 'id'
                  },
                  {
                    type: 'input-datetime',
                    name: 'start_time',
                    label: '开始时间',
                    format: 'YYYY-MM-DD HH:mm:ss',
                    description: '留空则立即开始盘点'
                  }
                ]
              }
            }
          }
        ],
        footerToolbar: [
          'switch-per-page',
          'pagination'
        ],
        columns: [
          {
            name: 'name',
            label: '盘点名称',
            type: 'text'
          },
          {
            name: 'description',
            label: '说明',
            type: 'text'
          },
          {
            name: 'status',
            label: '状态',
            type: 'mapping',
            map: {
              'pending': '<span class="label label-warning">待开始</span>',
              'in_progress': '<span class="label label-info">进行中</span>',
              'completed': '<span class="label label-success">已完成</span>',
              'cancelled': '<span class="label label-danger">已取消</span>'
            }
          },
          {
            name: 'started_at',
            label: '开始时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm:ss'
          },
          {
            name: 'completed_at',
            label: '完成时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm:ss'
          },
          {
            type: 'operation',
            label: '操作',
            buttons: [
              {
                type: 'button',
                actionType: 'dialog',
                label: '查看详情',
                level: 'info',
                size: 'sm',
                dialog: {
                  title: '盘点任务详情 - ${name}',
                  size: 'xl',
                  body: {
                    type: 'service',
                    api: '/api/v1/inventory/stocktaking/${id}',
                    body: [
                      {
                        type: 'panel',
                        title: '基本信息',
                        body: {
                          type: 'form',
                          mode: 'horizontal',
                          static: true,
                          controls: [
                            {
                              type: 'static',
                              name: 'name',
                              label: '盘点名称'
                            },
                            {
                              type: 'static',
                              name: 'description',
                              label: '说明'
                            },
                            {
                              type: 'static',
                              name: 'status',
                              label: '状态'
                            },
                            {
                              type: 'static',
                              name: 'started_at',
                              label: '开始时间'
                            },
                            {
                              type: 'static',
                              name: 'completed_at',
                              label: '完成时间'
                            }
                          ]
                        }
                      },
                      {
                        type: 'divider'
                      },
                      {
                        type: 'table',
                        title: '盘点记录',
                        source: '${records}',
                        columns: [
                          {
                            name: 'product_name',
                            label: '商品名称'
                          },
                          {
                            name: 'system_count',
                            label: '系统库存',
                            className: 'text-right'
                          },
                          {
                            name: 'actual_count',
                            label: '实际库存',
                            className: 'text-right'
                          },
                          {
                            name: 'difference',
                            label: '差异',
                            type: 'tpl',
                            tpl: '<%= data.difference > 0 ? "+" + data.difference : data.difference %>',
                            className: 'text-right',
                            classNameExpr: '<%= data.difference > 0 ? "text-success" : data.difference < 0 ? "text-danger" : "" %>'
                          },
                          {
                            name: 'reason',
                            label: '差异原因'
                          },
                          {
                            name: 'checked_at',
                            label: '盘点时间',
                            type: 'datetime',
                            format: 'MM-DD HH:mm'
                          }
                        ]
                      }
                    ]
                  }
                }
              },
              {
                type: 'button',
                actionType: 'dialog',
                label: '执行盘点',
                level: 'primary',
                size: 'sm',
                visibleOn: '${status === "pending" || status === "in_progress"}',
                dialog: {
                  title: '执行库存盘点 - ${name}',
                  size: 'xl',
                  body: [
                    {
                      type: 'service',
                      api: '/api/v1/inventory/stocktaking/${id}/products',
                      body: {
                        type: 'crud',
                        api: false,
                        source: '${products}',
                        headerToolbar: [
                          {
                            type: 'tpl',
                            tpl: '盘点进度: ${checked_count}/${total_count} (${Math.round(checked_count/total_count*100)}%)'
                          }
                        ],
                        footerToolbar: [],
                        columns: [
                          {
                            name: 'product_name',
                            label: '商品名称',
                            type: 'text'
                          },
                          {
                            name: 'system_count',
                            label: '系统库存',
                            type: 'text',
                            className: 'text-right'
                          },
                          {
                            name: 'actual_count',
                            label: '实际库存',
                            type: 'input-number',
                            quickEdit: {
                              mode: 'inline',
                              saveImmediately: true
                            },
                            className: 'text-right'
                          },
                          {
                            name: 'difference',
                            label: '差异',
                            type: 'tpl',
                            tpl: '<%= (data.actual_count || 0) - data.system_count %>',
                            className: 'text-right'
                          },
                          {
                            name: 'reason',
                            label: '差异原因',
                            type: 'input-text',
                            quickEdit: {
                              mode: 'inline',
                              saveImmediately: true,
                              placeholder: '如有差异请说明原因'
                            }
                          },
                          {
                            name: 'status',
                            label: '状态',
                            type: 'mapping',
                            map: {
                              'pending': '<span class="label label-default">待盘点</span>',
                              'checked': '<span class="label label-success">已盘点</span>'
                            }
                          }
                        ]
                      }
                    },
                    {
                      type: 'divider'
                    },
                    {
                      type: 'form',
                      title: '提交盘点记录',
                      api: {
                        method: 'put',
                        url: '/api/v1/inventory/stocktaking/${id}/records',
                        data: {
                          records: '${products | ARRAYFILTER(item => item.actual_count !== undefined) | ARRAYMAP(item => ({product_id: item.product_id, actual_count: item.actual_count, system_count: item.system_count, reason: item.reason || ""}))}'
                        },
                        messages: {
                          success: '盘点记录提交成功',
                          failed: '盘点记录提交失败'
                        }
                      },
                      controls: [
                        {
                          type: 'static',
                          label: '提示',
                          value: '点击提交将保存当前的盘点记录并调整相应的库存数量'
                        }
                      ]
                    }
                  ]
                }
              },
              {
                type: 'button',
                actionType: 'dialog',
                label: '完成盘点',
                level: 'success',
                size: 'sm',
                visibleOn: '${status === "in_progress"}',
                dialog: {
                  title: '完成库存盘点',
                  size: 'md',
                  body: {
                    type: 'form',
                    api: {
                      method: 'post',
                      url: '/api/v1/inventory/stocktaking/${id}/complete',
                      messages: {
                        success: '盘点任务已完成',
                        failed: '完成盘点失败'
                      }
                    },
                    controls: [
                      {
                        type: 'static',
                        name: 'name',
                        label: '盘点名称'
                      },
                      {
                        type: 'textarea',
                        name: 'summary',
                        label: '盘点总结',
                        rows: 4,
                        placeholder: '请输入本次盘点的总结...'
                      },
                      {
                        type: 'textarea',
                        name: 'notes',
                        label: '备注',
                        rows: 3,
                        placeholder: '其他需要记录的信息...'
                      }
                    ]
                  }
                }
              },
              {
                type: 'button',
                actionType: 'ajax',
                label: '取消盘点',
                level: 'danger',
                size: 'sm',
                visibleOn: '${status === "pending" || status === "in_progress"}',
                confirmText: '确认要取消这个盘点任务吗？',
                api: {
                  method: 'post',
                  url: '/api/v1/inventory/stocktaking/${id}/cancel',
                  messages: {
                    success: '盘点任务已取消',
                    failed: '取消盘点失败'
                  }
                }
              }
            ]
          }
        ]
      }
    ]
  };

  return <AmisRenderer schema={schema} />;
};

export default InventoryStocktakingPage;
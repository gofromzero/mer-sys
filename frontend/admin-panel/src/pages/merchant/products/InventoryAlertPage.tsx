// 库存预警配置页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const InventoryAlertPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '库存预警配置',
    body: [
      {
        type: 'crud',
        api: {
          method: 'get',
          url: '/api/v1/inventory/alerts/active'
        },
        syncLocation: false,
        headerToolbar: [
          {
            type: 'button',
            actionType: 'dialog',
            label: '新增预警规则',
            level: 'primary',
            dialog: {
              title: '新增库存预警规则',
              size: 'md',
              body: {
                type: 'form',
                api: {
                  method: 'post',
                  url: '/api/v1/products/${product_id}/inventory/alerts',
                  messages: {
                    success: '预警规则创建成功',
                    failed: '预警规则创建失败'
                  }
                },
                controls: [
                  {
                    type: 'select',
                    name: 'product_id',
                    label: '商品',
                    required: true,
                    source: {
                      method: 'get',
                      url: '/api/v1/products',
                      data: {
                        page_size: 1000
                      }
                    },
                    labelField: 'name',
                    valueField: 'id',
                    searchable: true,
                    validationErrors: {
                      required: '请选择商品'
                    }
                  },
                  {
                    type: 'select',
                    name: 'alert_type',
                    label: '预警类型',
                    required: true,
                    options: [
                      { label: '低库存预警', value: 'low_stock' },
                      { label: '缺货预警', value: 'out_of_stock' },
                      { label: '超储预警', value: 'overstock' }
                    ],
                    validationErrors: {
                      required: '请选择预警类型'
                    }
                  },
                  {
                    type: 'number',
                    name: 'threshold_value',
                    label: '预警阈值',
                    required: true,
                    min: 0,
                    validationErrors: {
                      required: '请输入预警阈值',
                      min: '阈值不能小于0'
                    }
                  },
                  {
                    type: 'checkboxes',
                    name: 'notification_channels',
                    label: '通知方式',
                    required: true,
                    options: [
                      { label: '系统通知', value: 'system' },
                      { label: '邮件通知', value: 'email' },
                      { label: '短信通知', value: 'sms' }
                    ],
                    validationErrors: {
                      required: '请选择至少一种通知方式'
                    }
                  },
                  {
                    type: 'switch',
                    name: 'is_active',
                    label: '启用状态',
                    value: true
                  }
                ]
              }
            }
          },
          {
            type: 'button',
            actionType: 'ajax',
            label: '批量检查预警',
            level: 'info',
            api: {
              method: 'post',
              url: '/api/v1/inventory/alerts/check-low-stock',
              messages: {
                success: '预警检查完成',
                failed: '预警检查失败'
              }
            },
            confirmText: '确认要执行批量预警检查吗？'
          }
        ],
        footerToolbar: [
          'switch-per-page',
          'pagination'
        ],
        columns: [
          {
            name: 'product_name',
            label: '商品名称',
            type: 'text'
          },
          {
            name: 'alert_type',
            label: '预警类型',
            type: 'mapping',
            map: {
              'low_stock': '<span class="label label-warning">低库存</span>',
              'out_of_stock': '<span class="label label-danger">缺货</span>',
              'overstock': '<span class="label label-info">超储</span>'
            }
          },
          {
            name: 'threshold_value',
            label: '预警阈值',
            type: 'text'
          },
          {
            name: 'notification_channels',
            label: '通知方式',
            type: 'each',
            items: {
              type: 'mapping',
              map: {
                'system': '<span class="label label-default">系统</span>',
                'email': '<span class="label label-primary">邮件</span>',
                'sms': '<span class="label label-success">短信</span>'
              }
            }
          },
          {
            name: 'is_active',
            label: '状态',
            type: 'status',
            map: {
              true: 1,
              false: 0
            }
          },
          {
            name: 'last_triggered_at',
            label: '最后触发时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm:ss'
          },
          {
            type: 'operation',
            label: '操作',
            buttons: [
              {
                type: 'button',
                actionType: 'ajax',
                label: '手动检查',
                level: 'info',
                size: 'sm',
                api: {
                  method: 'post',
                  url: '/api/v1/products/${product_id}/inventory/alerts/check',
                  messages: {
                    success: '预警检查完成',
                    failed: '预警检查失败'
                  }
                }
              },
              {
                type: 'button',
                actionType: 'ajax',
                label: '切换状态',
                level: 'warning',
                size: 'sm',
                api: {
                  method: 'post',
                  url: '/api/v1/inventory/alerts/${id}/toggle',
                  data: {
                    is_active: '${!is_active}'
                  },
                  messages: {
                    success: '状态切换成功',
                    failed: '状态切换失败'
                  }
                }
              },
              {
                type: 'button',
                actionType: 'dialog',
                label: '编辑',
                level: 'primary',
                size: 'sm',
                dialog: {
                  title: '编辑预警规则',
                  size: 'md',
                  body: {
                    type: 'form',
                    api: {
                      method: 'put',
                      url: '/api/v1/inventory/alerts/${id}',
                      messages: {
                        success: '预警规则更新成功',
                        failed: '预警规则更新失败'
                      }
                    },
                    initApi: {
                      method: 'get',
                      url: '/api/v1/inventory/alerts/${id}'
                    },
                    controls: [
                      {
                        type: 'static',
                        name: 'product_name',
                        label: '商品'
                      },
                      {
                        type: 'select',
                        name: 'alert_type',
                        label: '预警类型',
                        required: true,
                        options: [
                          { label: '低库存预警', value: 'low_stock' },
                          { label: '缺货预警', value: 'out_of_stock' },
                          { label: '超储预警', value: 'overstock' }
                        ]
                      },
                      {
                        type: 'number',
                        name: 'threshold_value',
                        label: '预警阈值',
                        required: true,
                        min: 0
                      },
                      {
                        type: 'checkboxes',
                        name: 'notification_channels',
                        label: '通知方式',
                        required: true,
                        options: [
                          { label: '系统通知', value: 'system' },
                          { label: '邮件通知', value: 'email' },
                          { label: '短信通知', value: 'sms' }
                        ]
                      },
                      {
                        type: 'switch',
                        name: 'is_active',
                        label: '启用状态'
                      }
                    ]
                  }
                }
              },
              {
                type: 'button',
                actionType: 'ajax',
                label: '删除',
                level: 'danger',
                size: 'sm',
                confirmText: '确认要删除这个预警规则吗？',
                api: {
                  method: 'delete',
                  url: '/api/v1/inventory/alerts/${id}',
                  messages: {
                    success: '预警规则删除成功',
                    failed: '预警规则删除失败'
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

export default InventoryAlertPage;
// 促销价格配置页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const PromotionalPricePage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '促销价格管理',
    toolbar: [
      {
        type: 'button',
        actionType: 'dialog',
        label: '创建促销活动',
        level: 'primary',
        dialog: {
          title: '创建促销价格',
          size: 'lg',
          body: {
            type: 'form',
            api: {
              method: 'post',
              url: '/api/v1/products/${productId}/promotional-prices',
              messages: {
                success: '促销价格创建成功',
                failed: '促销价格创建失败'
              }
            },
            controls: [
              {
                type: 'alert',
                level: 'info',
                body: '促销价格在指定时间段内自动生效，系统会自动检查时间冲突。'
              },
              {
                type: 'divider',
                title: '基本信息'
              },
              {
                type: 'group',
                body: [
                  {
                    type: 'input-number',
                    name: 'promotional_price.amount',
                    label: '促销价格',
                    required: true,
                    precision: 2,
                    min: 0,
                    suffix: '元',
                    validationErrors: {
                      required: '请输入促销价格'
                    }
                  },
                  {
                    type: 'select',
                    name: 'promotional_price.currency',
                    label: '货币',
                    value: 'CNY',
                    options: [
                      { label: '人民币', value: 'CNY' },
                      { label: '美元', value: 'USD' }
                    ]
                  }
                ]
              },
              {
                type: 'input-number',
                name: 'discount_percentage',
                label: '折扣百分比',
                precision: 2,
                min: 0,
                max: 100,
                suffix: '%',
                description: '可选，用于展示折扣信息'
              },
              {
                type: 'divider',
                title: '时间设置'
              },
              {
                type: 'input-datetime-range',
                name: 'promotion_period',
                label: '促销时段',
                required: true,
                format: 'YYYY-MM-DD HH:mm:ss',
                inputFormat: 'YYYY-MM-DD HH:mm:ss',
                validationErrors: {
                  required: '请选择促销时段'
                }
              },
              {
                type: 'divider',
                title: '促销条件'
              },
              {
                type: 'switch',
                name: 'enable_conditions',
                label: '启用促销条件',
                description: '可以设置特定条件下才能享受促销价格'
              },
              {
                type: 'container',
                visibleOn: '${enable_conditions}',
                body: [
                  {
                    type: 'combo',
                    name: 'conditions',
                    label: '促销条件',
                    multiple: true,
                    items: [
                      {
                        type: 'select',
                        name: 'type',
                        label: '条件类型',
                        options: [
                          { label: '最小购买数量', value: 'min_quantity' },
                          { label: '会员等级要求', value: 'member_level' },
                          { label: '最小订单金额', value: 'min_order_amount' },
                          { label: '首次购买', value: 'first_purchase' },
                          { label: '特定时间段', value: 'time_range' }
                        ]
                      },
                      {
                        type: 'input-text',
                        name: 'value',
                        label: '条件值',
                        description: '根据条件类型填写对应的值'
                      }
                    ]
                  }
                ]
              },
              {
                type: 'divider'
              },
              {
                type: 'alert',
                level: 'warning',
                body: '促销价格创建后会立即按时间计划生效，请确认时间设置无误。'
              }
            ]
          }
        }
      },
      {
        type: 'button',
        actionType: 'dialog',
        label: '促销预览',
        level: 'info',
        dialog: {
          title: '促销价格预览',
          size: 'md',
          body: {
            type: 'service',
            api: '/api/v1/products/${productId}/current-promotion',
            body: [
              {
                type: 'alert',
                level: '${is_promotion_active ? "success" : "info"}',
                body: '${is_promotion_active ? "🎉 当前有促销活动进行中" : "ℹ️ 当前没有促销活动"}'
              },
              {
                type: 'container',
                visibleOn: '${is_promotion_active}',
                body: [
                  {
                    type: 'property',
                    title: '当前促销信息',
                    items: [
                      { label: '原价', content: '¥${original_price}' },
                      { label: '促销价', content: '¥${promotional_price}' },
                      { label: '节省', content: '¥${discount_amount}' },
                      { label: '折扣', content: '${discount_percentage}%' },
                      { label: '开始时间', content: '${valid_from}' },
                      { label: '结束时间', content: '${valid_until}' },
                      { label: '剩余时间', content: '${remaining_time}' }
                    ]
                  }
                ]
              },
              {
                type: 'container',
                visibleOn: '${upcoming_promotions.length > 0}',
                body: [
                  {
                    type: 'divider',
                    title: '即将开始的促销'
                  },
                  {
                    type: 'table',
                    source: '${upcoming_promotions}',
                    columns: [
                      { name: 'promotional_price', label: '促销价', type: 'text' },
                      { name: 'valid_from', label: '开始时间', type: 'datetime' },
                      { name: 'valid_until', label: '结束时间', type: 'datetime' }
                    ]
                  }
                ]
              }
            ]
          }
        }
      }
    ],
    body: {
      type: 'crud',
      api: '/api/v1/products/${productId}/promotional-prices',
      interval: 10000, // 每10秒刷新促销状态
      headerToolbar: [
        {
          type: 'filter-toggler',
          align: 'left'
        },
        {
          type: 'bulk-actions',
          align: 'left'
        },
        {
          type: 'pagination',
          align: 'right'
        }
      ],
      filter: {
        title: '筛选促销活动',
        controls: [
          {
            type: 'select',
            name: 'status',
            label: '促销状态',
            placeholder: '全部状态',
            options: [
              { label: '未开始', value: 'pending' },
              { label: '进行中', value: 'active' },
              { label: '已结束', value: 'expired' },
              { label: '已禁用', value: 'disabled' }
            ]
          },
          {
            type: 'input-datetime-range',
            name: 'date_range',
            label: '时间范围'
          }
        ]
      },
      bulkActions: [
        {
          label: '批量启用',
          actionType: 'ajax',
          api: {
            method: 'put',
            url: '/api/v1/promotional-prices/batch-status',
            data: {
              promotion_ids: '${ids}',
              is_active: true
            }
          },
          confirmText: '确认启用选中的促销活动？'
        },
        {
          label: '批量禁用',
          actionType: 'ajax',
          api: {
            method: 'put',
            url: '/api/v1/promotional-prices/batch-status',
            data: {
              promotion_ids: '${ids}',
              is_active: false
            }
          },
          confirmText: '确认禁用选中的促销活动？'
        }
      ],
      columns: [
        {
          name: 'id',
          label: 'ID',
          type: 'text',
          width: 80
        },
        {
          name: 'promotional_price.amount',
          label: '促销价格',
          type: 'text',
          tpl: '¥${promotional_price.amount | number:2}'
        },
        {
          name: 'discount_percentage',
          label: '折扣',
          type: 'text',
          tpl: '${discount_percentage}%',
          placeholder: '-'
        },
        {
          name: 'valid_from',
          label: '开始时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss'
        },
        {
          name: 'valid_until',
          label: '结束时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss'
        },
        {
          name: 'status',
          label: '状态',
          type: 'mapping',
          map: {
            'pending': '<span class="label label-default">未开始</span>',
            'active': '<span class="label label-success">进行中</span>',
            'expired': '<span class="label label-secondary">已结束</span>',
            'disabled': '<span class="label label-danger">已禁用</span>'
          }
        },
        {
          name: 'is_active',
          label: '启用状态',
          type: 'status',
          map: {
            true: 'success',
            false: 'danger'
          }
        },
        {
          name: 'conditions',
          label: '促销条件',
          type: 'text',
          tpl: '${conditions.length > 0 ? conditions.length + "个条件" : "无条件"}',
          popOver: {
            title: '促销条件详情',
            body: {
              type: 'table',
              source: '${conditions}',
              columns: [
                { name: 'type', label: '条件类型' },
                { name: 'value', label: '条件值' }
              ]
            }
          }
        },
        {
          name: 'created_at',
          label: '创建时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss'
        },
        {
          type: 'operation',
          label: '操作',
          width: 250,
          buttons: [
            {
              label: '编辑',
              type: 'button',
              actionType: 'dialog',
              dialog: {
                title: '编辑促销价格',
                size: 'lg',
                body: {
                  type: 'form',
                  api: {
                    method: 'put',
                    url: '/api/v1/products/${productId}/promotional-prices/${id}',
                    messages: {
                      success: '促销价格更新成功',
                      failed: '促销价格更新失败'
                    }
                  },
                  initApi: '/api/v1/products/${productId}/promotional-prices/${id}',
                  controls: [
                    {
                      type: 'group',
                      body: [
                        {
                          type: 'input-number',
                          name: 'promotional_price.amount',
                          label: '促销价格',
                          precision: 2,
                          min: 0
                        },
                        {
                          type: 'input-number',
                          name: 'discount_percentage',
                          label: '折扣百分比',
                          precision: 2,
                          min: 0,
                          max: 100
                        }
                      ]
                    },
                    {
                      type: 'input-datetime-range',
                      name: 'promotion_period',
                      label: '促销时段',
                      format: 'YYYY-MM-DD HH:mm:ss'
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
              label: '效果分析',
              type: 'button',
              actionType: 'dialog',
              dialog: {
                title: '促销效果分析',
                size: 'lg',
                body: {
                  type: 'service',
                  api: '/api/v1/products/${productId}/promotional-prices/${id}/analytics',
                  body: [
                    {
                      type: 'divider',
                      title: '促销效果概览'
                    },
                    {
                      type: 'cards',
                      source: '${summary_stats}',
                      card: {
                        body: [
                          {
                            type: 'tpl',
                            tpl: '<div class="text-center"><h3 class="text-info">${value}</h3><p class="text-muted">${label}</p></div>'
                          }
                        ]
                      }
                    },
                    {
                      type: 'divider',
                      title: '销售趋势'
                    },
                    {
                      type: 'chart',
                      api: '/api/v1/products/${productId}/promotional-prices/${id}/sales-trend',
                      config: {
                        type: 'line',
                        title: {
                          text: '促销期间销售趋势'
                        },
                        xAxis: {
                          type: 'category',
                          data: '${dates}'
                        },
                        yAxis: [
                          {
                            type: 'value',
                            name: '销量',
                            position: 'left'
                          },
                          {
                            type: 'value',
                            name: '收入',
                            position: 'right'
                          }
                        ],
                        series: [
                          {
                            name: '销量',
                            type: 'line',
                            yAxisIndex: 0,
                            data: '${sales_volume}'
                          },
                          {
                            name: '收入',
                            type: 'line',
                            yAxisIndex: 1,
                            data: '${revenue}'
                          }
                        ]
                      }
                    },
                    {
                      type: 'divider',
                      title: '客户分析'
                    },
                    {
                      type: 'chart',
                      config: {
                        type: 'pie',
                        title: {
                          text: '客户类型分布'
                        },
                        series: [{
                          type: 'pie',
                          data: '${customer_segments}',
                          radius: '50%'
                        }]
                      }
                    }
                  ]
                }
              }
            },
            {
              label: '复制',
              type: 'button',
              actionType: 'dialog',
              dialog: {
                title: '复制促销活动',
                body: {
                  type: 'form',
                  api: {
                    method: 'post',
                    url: '/api/v1/products/${productId}/promotional-prices/${id}/duplicate'
                  },
                  controls: [
                    {
                      type: 'input-datetime-range',
                      name: 'new_period',
                      label: '新的促销时段',
                      required: true
                    },
                    {
                      type: 'switch',
                      name: 'adjust_price',
                      label: '调整价格'
                    },
                    {
                      type: 'input-number',
                      name: 'new_price',
                      label: '新促销价格',
                      visibleOn: '${adjust_price}',
                      precision: 2,
                      min: 0
                    }
                  ]
                }
              }
            },
            {
              label: '删除',
              type: 'button',
              actionType: 'ajax',
              level: 'danger',
              confirmText: '确认删除此促销活动？正在进行的促销将立即停止。',
              api: {
                method: 'delete',
                url: '/api/v1/products/${productId}/promotional-prices/${id}',
                messages: {
                  success: '促销价格删除成功',
                  failed: '促销价格删除失败'
                }
              }
            }
          ]
        }
      ]
    }
  };

  return <AmisRenderer schema={schema} />;
};

export default PromotionalPricePage;
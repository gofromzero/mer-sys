// 商品价格变更历史页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const PriceHistoryPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '价格变更历史',
    toolbar: [
      {
        type: 'button',
        actionType: 'dialog',
        label: '价格变更',
        level: 'primary',
        dialog: {
          title: '执行价格变更',
          size: 'lg',
          body: {
            type: 'form',
            api: {
              method: 'post',
              url: '/api/v1/products/${productId}/price-change',
              messages: {
                success: '价格变更成功',
                failed: '价格变更失败'
              }
            },
            controls: [
              {
                type: 'alert',
                level: 'info',
                body: '价格变更将会影响所有未来的订单，请谨慎操作。'
              },
              {
                type: 'divider',
                title: '新价格设置'
              },
              {
                type: 'group',
                body: [
                  {
                    type: 'input-number',
                    name: 'new_price.amount',
                    label: '新价格',
                    required: true,
                    precision: 2,
                    min: 0,
                    suffix: '元',
                    validationErrors: {
                      required: '请输入新价格',
                      min: '价格必须大于0'
                    }
                  },
                  {
                    type: 'select',
                    name: 'new_price.currency',
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
                type: 'textarea',
                name: 'change_reason',
                label: '变更原因',
                required: true,
                placeholder: '请详细说明价格变更的原因...',
                minRows: 3,
                validationErrors: {
                  required: '请填写变更原因'
                }
              },
              {
                type: 'datetime',
                name: 'effective_date',
                label: '生效时间',
                value: 'now',
                format: 'YYYY-MM-DD HH:mm:ss',
                description: '价格变更的生效时间，默认为立即生效'
              },
              {
                type: 'divider',
                title: '影响评估'
              },
              {
                type: 'service',
                api: {
                  method: 'post',
                  url: '/api/v1/products/${productId}/price-impact-assessment',
                  data: {
                    new_price: '${new_price}',
                    effective_date: '${effective_date}'
                  },
                  adaptor: 'return { ...payload, data: payload.data || {} };'
                },
                body: [
                  {
                    type: 'alert',
                    level: '${risk_level === "high" ? "danger" : (risk_level === "medium" ? "warning" : "success")}',
                    body: '风险等级：${risk_level_text} ${risk_score ? "(" + risk_score + "分)" : ""}'
                  },
                  {
                    type: 'container',
                    body: [
                      {
                        type: 'property',
                        title: '预计影响',
                        items: [
                          { label: '影响订单数', content: '${estimated_affected_orders || 0}' },
                          { label: '收入变化', content: '${revenue_change > 0 ? "+" : ""}${revenue_change || 0}元' },
                          { label: '需求变化', content: '${demand_change > 0 ? "+" : ""}${demand_change || 0}%' },
                          { label: '价格敏感度', content: '${price_elasticity || "暂无数据"}' }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'container',
                    visibleOn: '${recommendations && recommendations.length > 0}',
                    body: [
                      {
                        type: 'divider',
                        title: '建议'
                      },
                      {
                        type: 'list',
                        source: '${recommendations}',
                        listItem: {
                          body: [
                            {
                              type: 'tpl',
                              tpl: '• ${this}'
                            }
                          ]
                        }
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
                body: '价格变更后将立即生效，并记录在价格历史中。'
              }
            ]
          }
        }
      },
      {
        type: 'button',
        actionType: 'dialog',
        label: '批量分析',
        level: 'info',
        dialog: {
          title: '价格趋势分析',
          size: 'xl',
          body: {
            type: 'service',
            api: '/api/v1/products/${productId}/price-analysis',
            body: [
              {
                type: 'divider',
                title: '价格趋势'
              },
              {
                type: 'chart',
                config: {
                  type: 'line',
                  title: {
                    text: '价格变化趋势'
                  },
                  tooltip: {
                    trigger: 'axis',
                    formatter: function(params) {
                      return params[0].name + '<br/>' + 
                             '价格: ¥' + params[0].value + '<br/>' +
                             '变更原因: ' + (params[0].data.reason || '无');
                    }
                  },
                  xAxis: {
                    type: 'category',
                    data: '${trend_dates}',
                    axisLabel: {
                      rotate: 45
                    }
                  },
                  yAxis: {
                    type: 'value',
                    name: '价格 (元)',
                    axisLabel: {
                      formatter: '¥{value}'
                    }
                  },
                  series: [{
                    data: '${trend_prices}',
                    type: 'line',
                    smooth: true,
                    symbol: 'circle',
                    symbolSize: 8,
                    lineStyle: {
                      width: 3
                    },
                    itemStyle: {
                      color: '#1890ff'
                    }
                  }]
                }
              },
              {
                type: 'divider',
                title: '统计信息'
              },
              {
                type: 'cards',
                source: '${statistics}',
                card: {
                  body: [
                    {
                      type: 'tpl',
                      tpl: '<div class="text-center"><h3 class="text-info">${value}</h3><p class="text-muted">${title}</p><small>${description}</small></div>'
                    }
                  ]
                }
              },
              {
                type: 'divider',
                title: '变更频率分析'
              },
              {
                type: 'chart',
                config: {
                  type: 'bar',
                  title: {
                    text: '月度价格变更频率'
                  },
                  xAxis: {
                    type: 'category',
                    data: '${monthly_labels}'
                  },
                  yAxis: {
                    type: 'value',
                    name: '变更次数'
                  },
                  series: [{
                    data: '${monthly_changes}',
                    type: 'bar',
                    itemStyle: {
                      color: '#52c41a'
                    }
                  }]
                }
              }
            ]
          }
        }
      }
    ],
    body: {
      type: 'crud',
      api: {
        method: 'get',
        url: '/api/v1/products/${productId}/price-history',
        data: {
          page: '${page}',
          page_size: '${perPage}',
          start_date: '${start_date}',
          end_date: '${end_date}',
          changed_by: '${changed_by}',
          sort_by: '${orderBy || "effective_date"}',
          sort_order: '${orderDir || "desc"}'
        }
      },
      interval: 30000, // 每30秒刷新一次
      defaultParams: {
        page: 1,
        perPage: 20
      },
      headerToolbar: [
        {
          type: 'filter-toggler',
          align: 'left'
        },
        {
          type: 'reload',
          align: 'left'
        },
        {
          type: 'export-excel',
          align: 'right',
          label: '导出Excel'
        },
        {
          type: 'pagination',
          align: 'right'
        }
      ],
      filter: {
        title: '筛选条件',
        controls: [
          {
            type: 'input-datetime-range',
            name: 'date_range',
            label: '时间范围',
            format: 'YYYY-MM-DD',
            inputFormat: 'YYYY-MM-DD'
          },
          {
            type: 'input-text',
            name: 'changed_by',
            label: '操作人',
            placeholder: '请输入操作人姓名或ID'
          },
          {
            type: 'input-number-range',
            name: 'price_range',
            label: '价格范围',
            suffix: '元'
          }
        ]
      },
      columns: [
        {
          name: 'id',
          label: 'ID',
          type: 'text',
          width: 80
        },
        {
          name: 'old_price',
          label: '原价格',
          type: 'tpl',
          tpl: '¥${old_price.amount | number:2}',
          width: 100
        },
        {
          name: 'new_price',
          label: '新价格',
          type: 'tpl',
          tpl: '¥${new_price.amount | number:2}',
          width: 100
        },
        {
          name: 'price_change',
          label: '变化',
          type: 'tpl',
          tpl: '<span class="${new_price.amount > old_price.amount ? \'text-success\' : \'text-danger\'}">${new_price.amount > old_price.amount ? \'+\' : \'\'}${(new_price.amount - old_price.amount) | number:2}元 (${((new_price.amount - old_price.amount) / old_price.amount * 100) | number:1}%)</span>',
          width: 150
        },
        {
          name: 'change_reason',
          label: '变更原因',
          type: 'text',
          width: 200,
          popOver: {
            title: '变更原因详情',
            body: '${change_reason}'
          }
        },
        {
          name: 'changed_by_name',
          label: '操作人',
          type: 'text',
          width: 100
        },
        {
          name: 'effective_date',
          label: '生效时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss',
          sortable: true,
          width: 150
        },
        {
          name: 'created_at',
          label: '记录时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss',
          sortable: true,
          width: 150
        },
        {
          type: 'operation',
          label: '操作',
          width: 200,
          buttons: [
            {
              label: '影响分析',
              type: 'button',
              actionType: 'dialog',
              dialog: {
                title: '价格变更影响分析',
                size: 'lg',
                body: {
                  type: 'service',
                  api: '/api/v1/products/${productId}/price-history/${id}/impact',
                  body: [
                    {
                      type: 'divider',
                      title: '变更详情'
                    },
                    {
                      type: 'property',
                      title: '基本信息',
                      items: [
                        { label: '变更时间', content: '${effective_date}' },
                        { label: '原价格', content: '¥${old_price.amount}' },
                        { label: '新价格', content: '¥${new_price.amount}' },
                        { label: '变更幅度', content: '${price_change_percentage}%' },
                        { label: '操作人', content: '${changed_by_name}' }
                      ]
                    },
                    {
                      type: 'divider',
                      title: '实际影响'
                    },
                    {
                      type: 'cards',
                      source: '${impact_metrics}',
                      card: {
                        body: [
                          {
                            type: 'tpl',
                            tpl: '<div class="text-center"><h4>${metric_name}</h4><h2 class="${value_trend === \'up\' ? \'text-success\' : (value_trend === \'down\' ? \'text-danger\' : \'text-info\')}">${metric_value}</h2><p class="text-muted">${metric_description}</p></div>'
                          }
                        ]
                      }
                    },
                    {
                      type: 'container',
                      visibleOn: '${sales_data && sales_data.length > 0}',
                      body: [
                        {
                          type: 'divider',
                          title: '销售数据对比'
                        },
                        {
                          type: 'chart',
                          config: {
                            type: 'line',
                            title: {
                              text: '价格变更前后销售对比'
                            },
                            legend: {
                              data: ['变更前', '变更后']
                            },
                            xAxis: {
                              type: 'category',
                              data: '${comparison_dates}'
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
                                name: '销量-变更前',
                                type: 'line',
                                yAxisIndex: 0,
                                data: '${sales_before}',
                                itemStyle: { color: '#ff4d4f' }
                              },
                              {
                                name: '销量-变更后',
                                type: 'line',
                                yAxisIndex: 0,
                                data: '${sales_after}',
                                itemStyle: { color: '#52c41a' }
                              },
                              {
                                name: '收入-变更前',
                                type: 'line',
                                yAxisIndex: 1,
                                data: '${revenue_before}',
                                itemStyle: { color: '#1890ff' },
                                lineStyle: { type: 'dashed' }
                              },
                              {
                                name: '收入-变更后',
                                type: 'line',
                                yAxisIndex: 1,
                                data: '${revenue_after}',
                                itemStyle: { color: '#722ed1' },
                                lineStyle: { type: 'dashed' }
                              }
                            ]
                          }
                        }
                      ]
                    },
                    {
                      type: 'container',
                      visibleOn: '${customer_feedback && customer_feedback.length > 0}',
                      body: [
                        {
                          type: 'divider',
                          title: '客户反馈'
                        },
                        {
                          type: 'table',
                          source: '${customer_feedback}',
                          columns: [
                            { name: 'feedback_date', label: '时间', type: 'datetime', format: 'MM-DD HH:mm' },
                            { name: 'customer_level', label: '客户等级', type: 'text' },
                            { name: 'feedback_type', label: '反馈类型', type: 'text' },
                            { name: 'feedback_content', label: '内容', type: 'text' }
                          ]
                        }
                      ]
                    }
                  ]
                }
              }
            },
            {
              label: '回滚',
              type: 'button',
              actionType: 'ajax',
              level: 'warning',
              visibleOn: '${can_rollback}',
              confirmText: '确认要回滚到此价格吗？这将创建一个新的价格变更记录。',
              api: {
                method: 'post',
                url: '/api/v1/products/${productId}/price-rollback',
                data: {
                  target_history_id: '${id}',
                  rollback_reason: '手动回滚到历史价格'
                },
                messages: {
                  success: '价格回滚成功',
                  failed: '价格回滚失败'
                }
              }
            },
            {
              label: '导出报告',
              type: 'button',
              actionType: 'download',
              api: '/api/v1/products/${productId}/price-history/${id}/report.pdf',
              target: '_blank'
            }
          ]
        }
      ],
      placeholder: '暂无价格变更历史记录'
    }
  };

  return <AmisRenderer schema={schema} />;
};

export default PriceHistoryPage;
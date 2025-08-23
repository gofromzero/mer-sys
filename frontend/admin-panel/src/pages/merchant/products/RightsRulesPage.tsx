// 权益规则配置页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const RightsRulesPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '权益消耗规则配置',
    toolbar: [
      {
        type: 'button',
        actionType: 'dialog',
        label: '配置权益规则',
        level: 'primary',
        dialog: {
          title: '权益消耗规则配置',
          size: 'lg',
          body: {
            type: 'form',
            api: {
              method: 'post',
              url: '/api/v1/products/${productId}/rights-rules',
              messages: {
                success: '权益规则配置成功',
                failed: '权益规则配置失败'
              }
            },
            controls: [
              {
                type: 'alert',
                level: 'info',
                body: '每个商品只能配置一个权益消耗规则，新配置将覆盖原有规则。'
              },
              {
                type: 'select',
                name: 'rule_type',
                label: '消耗规则类型',
                required: true,
                options: [
                  { 
                    label: '固定费率', 
                    value: 'fixed_rate',
                    description: '每单位商品消耗固定权益点数'
                  },
                  { 
                    label: '百分比扣减', 
                    value: 'percentage',
                    description: '按订单总金额的百分比消耗权益'
                  },
                  { 
                    label: '阶梯消耗', 
                    value: 'tiered',
                    description: '根据购买数量采用阶梯式消耗'
                  }
                ],
                validationErrors: {
                  required: '请选择消耗规则类型'
                }
              },
              {
                type: 'container',
                visibleOn: "${rule_type === 'fixed_rate'}",
                body: [
                  {
                    type: 'static',
                    label: '固定费率说明',
                    tpl: '每购买1件商品固定消耗指定权益点数'
                  },
                  {
                    type: 'input-number',
                    name: 'consumption_rate',
                    label: '消耗比例',
                    required: true,
                    precision: 2,
                    min: 0,
                    suffix: '点/件',
                    description: '每件商品消耗的权益点数',
                    validationErrors: {
                      required: '请输入消耗比例'
                    }
                  }
                ]
              },
              {
                type: 'container',
                visibleOn: "${rule_type === 'percentage'}",
                body: [
                  {
                    type: 'static',
                    label: '百分比扣减说明',
                    tpl: '根据订单总金额的百分比消耗权益点数'
                  },
                  {
                    type: 'input-number',
                    name: 'consumption_rate',
                    label: '消耗比例',
                    required: true,
                    precision: 4,
                    min: 0,
                    max: 1,
                    suffix: '%',
                    step: 0.001,
                    description: '订单金额的百分比（0.1 = 10%）',
                    validationErrors: {
                      required: '请输入消耗比例',
                      max: '消耗比例不能超过100%'
                    }
                  }
                ]
              },
              {
                type: 'container',
                visibleOn: "${rule_type === 'tiered'}",
                body: [
                  {
                    type: 'static',
                    label: '阶梯消耗说明',
                    tpl: '根据购买数量采用不同的权益消耗比例'
                  },
                  {
                    type: 'input-number',
                    name: 'consumption_rate',
                    label: '基础消耗比例',
                    required: true,
                    precision: 2,
                    min: 0,
                    suffix: '点/件',
                    description: '基础消耗比例，系统会根据数量自动计算阶梯优惠',
                    validationErrors: {
                      required: '请输入基础消耗比例'
                    }
                  }
                ]
              },
              {
                type: 'input-number',
                name: 'min_rights_required',
                label: '最低权益要求',
                value: 0,
                precision: 2,
                min: 0,
                suffix: '点',
                description: '用户需要拥有的最低权益点数才能购买'
              },
              {
                type: 'select',
                name: 'insufficient_rights_action',
                label: '权益不足处理策略',
                required: true,
                options: [
                  { 
                    label: '阻止购买', 
                    value: 'block_purchase',
                    description: '权益不足时禁止购买'
                  },
                  { 
                    label: '部分支付', 
                    value: 'partial_payment',
                    description: '使用现有权益抵扣，剩余部分现金支付'
                  },
                  { 
                    label: '现金补足', 
                    value: 'cash_payment',
                    description: '权益不足部分自动转为现金支付'
                  }
                ],
                validationErrors: {
                  required: '请选择权益不足处理策略'
                }
              },
              {
                type: 'divider'
              },
              {
                type: 'alert',
                level: 'warning',
                body: '权益规则配置后将立即生效，请谨慎操作。'
              }
            ]
          }
        }
      },
      {
        type: 'button',
        actionType: 'dialog',
        label: '验证权益余额',
        level: 'info',
        dialog: {
          title: '权益余额验证',
          size: 'md',
          body: {
            type: 'form',
            api: {
              method: 'post',
              url: '/api/v1/products/${productId}/validate-rights',
              adaptor: 'return {\n  ...payload,\n  data: payload.data || {}\n};'
            },
            controls: [
              {
                type: 'input-number',
                name: 'user_id',
                label: '用户ID',
                required: true,
                description: '要验证的用户ID'
              },
              {
                type: 'input-number',
                name: 'quantity',
                label: '购买数量',
                value: 1,
                min: 1,
                required: true
              },
              {
                type: 'group',
                body: [
                  {
                    type: 'input-number',
                    name: 'total_amount.amount',
                    label: '订单金额',
                    precision: 2,
                    min: 0,
                    required: true
                  },
                  {
                    type: 'select',
                    name: 'total_amount.currency',
                    label: '货币',
                    value: 'CNY',
                    options: [
                      { label: '人民币', value: 'CNY' }
                    ]
                  }
                ]
              },
              {
                type: 'divider'
              },
              {
                type: 'static',
                label: '验证结果',
                tpl: '${is_valid ? "✅ 权益充足" : "❌ 权益不足"}',
                className: 'text-lg'
              },
              {
                type: 'static',
                label: '所需权益',
                tpl: '${required_rights}点'
              },
              {
                type: 'static',
                label: '可用权益',
                tpl: '${available_rights}点'
              },
              {
                type: 'static',
                label: '不足金额',
                tpl: '${insufficient_amount}点',
                visibleOn: '${insufficient_amount > 0}'
              },
              {
                type: 'static',
                label: '建议处理',
                tpl: '${suggested_action | mapping:{"block_purchase":"阻止购买","partial_payment":"部分支付","cash_payment":"现金补足"}}',
                visibleOn: '${suggested_action}'
              },
              {
                type: 'static',
                label: '需要现金',
                tpl: '¥${cash_payment_required.amount | number:2}',
                visibleOn: '${cash_payment_required.amount > 0}'
              }
            ]
          }
        }
      }
    ],
    body: {
      type: 'crud',
      api: '/api/v1/products/${productId}/rights-rules',
      interval: 5000, // 每5秒刷新一次
      columns: [
        {
          name: 'id',
          label: 'ID',
          type: 'text',
          width: 80
        },
        {
          name: 'rule_type',
          label: '规则类型',
          type: 'mapping',
          map: {
            'fixed_rate': '<span class="label label-info">固定费率</span>',
            'percentage': '<span class="label label-success">百分比扣减</span>',
            'tiered': '<span class="label label-warning">阶梯消耗</span>'
          }
        },
        {
          name: 'consumption_rate',
          label: '消耗比例',
          type: 'text',
          tpl: '${consumption_rate}${rule_type === "percentage" ? "%" : "点"}'
        },
        {
          name: 'min_rights_required',
          label: '最低权益要求',
          type: 'text',
          tpl: '${min_rights_required}点'
        },
        {
          name: 'insufficient_rights_action',
          label: '权益不足策略',
          type: 'mapping',
          map: {
            'block_purchase': '<span class="label label-danger">阻止购买</span>',
            'partial_payment': '<span class="label label-warning">部分支付</span>',
            'cash_payment': '<span class="label label-success">现金补足</span>'
          }
        },
        {
          name: 'is_active',
          label: '状态',
          type: 'status',
          map: {
            true: 'success',
            false: 'danger'
          }
        },
        {
          name: 'created_at',
          label: '创建时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss'
        },
        {
          name: 'updated_at',
          label: '更新时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss'
        },
        {
          type: 'operation',
          label: '操作',
          width: 200,
          buttons: [
            {
              label: '编辑',
              type: 'button',
              actionType: 'dialog',
              dialog: {
                title: '编辑权益规则',
                size: 'lg',
                body: {
                  type: 'form',
                  api: {
                    method: 'put',
                    url: '/api/v1/products/${productId}/rights-rules/${id}',
                    messages: {
                      success: '权益规则更新成功',
                      failed: '权益规则更新失败'
                    }
                  },
                  initApi: '/api/v1/products/${productId}/rights-rules/${id}',
                  controls: [
                    {
                      type: 'static',
                      label: '规则类型',
                      tpl: '${rule_type | mapping:{"fixed_rate":"固定费率","percentage":"百分比扣减","tiered":"阶梯消耗"}}'
                    },
                    {
                      type: 'input-number',
                      name: 'consumption_rate',
                      label: '消耗比例',
                      precision: 4,
                      min: 0
                    },
                    {
                      type: 'input-number',
                      name: 'min_rights_required',
                      label: '最低权益要求',
                      precision: 2,
                      min: 0
                    },
                    {
                      type: 'select',
                      name: 'insufficient_rights_action',
                      label: '权益不足处理策略',
                      options: [
                        { label: '阻止购买', value: 'block_purchase' },
                        { label: '部分支付', value: 'partial_payment' },
                        { label: '现金补足', value: 'cash_payment' }
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
              label: '权益统计',
              type: 'button',
              actionType: 'dialog',
              dialog: {
                title: '权益消耗统计',
                size: 'lg',
                body: {
                  type: 'service',
                  api: '/api/v1/products/${productId}/rights-statistics',
                  body: [
                    {
                      type: 'divider',
                      title: '消耗统计'
                    },
                    {
                      type: 'cards',
                      source: '${statistics}',
                      card: {
                        body: [
                          {
                            type: 'tpl',
                            tpl: '<h4>${title}</h4><p class="text-muted">${description}</p><h2 class="text-info">${value}</h2>'
                          }
                        ]
                      }
                    },
                    {
                      type: 'divider',
                      title: '趋势图表'
                    },
                    {
                      type: 'chart',
                      api: '/api/v1/products/${productId}/rights-consumption-trend',
                      config: {
                        type: 'line',
                        title: {
                          text: '权益消耗趋势'
                        },
                        xAxis: {
                          type: 'category',
                          data: '${dates}'
                        },
                        yAxis: {
                          type: 'value'
                        },
                        series: [{
                          data: '${consumption_data}',
                          type: 'line',
                          smooth: true
                        }]
                      }
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
              confirmText: '确认删除此权益规则？删除后商品将不再消耗权益。',
              api: {
                method: 'delete',
                url: '/api/v1/products/${productId}/rights-rules/${id}',
                messages: {
                  success: '权益规则删除成功',
                  failed: '权益规则删除失败'
                }
              }
            }
          ]
        }
      ],
      placeholder: '该商品暂未配置权益消耗规则'
    }
  };

  return <AmisRenderer schema={schema} />;
};

export default RightsRulesPage;
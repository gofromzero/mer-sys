// 商品定价配置页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const PricingConfigPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '定价规则配置',
    toolbar: [
      {
        type: 'button',
        actionType: 'dialog',
        label: '新增定价规则',
        level: 'primary',
        dialog: {
          title: '新增定价规则',
          size: 'lg',
          body: {
            type: 'form',
            api: {
              method: 'post',
              url: '/api/v1/products/${productId}/pricing-rules',
              messages: {
                success: '定价规则创建成功',
                failed: '定价规则创建失败'
              }
            },
            controls: [
              {
                type: 'select',
                name: 'rule_type',
                label: '规则类型',
                required: true,
                options: [
                  { label: '基础价格', value: 'base_price' },
                  { label: '阶梯价格', value: 'volume_discount' },
                  { label: '会员价格', value: 'member_discount' },
                  { label: '时段优惠', value: 'time_based_discount' }
                ],
                validationErrors: {
                  required: '请选择规则类型'
                }
              },
              {
                type: 'input-number',
                name: 'priority',
                label: '优先级',
                value: 0,
                description: '数值越大优先级越高',
                min: 0,
                max: 100
              },
              {
                type: 'datetime',
                name: 'valid_from',
                label: '生效时间',
                required: true,
                format: 'YYYY-MM-DD HH:mm:ss',
                validationErrors: {
                  required: '请选择生效时间'
                }
              },
              {
                type: 'datetime',
                name: 'valid_until',
                label: '失效时间',
                format: 'YYYY-MM-DD HH:mm:ss',
                description: '留空表示永久有效'
              },
              {
                type: 'divider'
              },
              {
                type: 'container',
                visibleOn: "${rule_type === 'base_price'}",
                body: [
                  {
                    type: 'static',
                    label: '基础价格配置',
                    tpl: '设置商品的基础售价'
                  },
                  {
                    type: 'input-number',
                    name: 'base_amount',
                    label: '价格金额',
                    required: true,
                    precision: 2,
                    min: 0,
                    suffix: '元'
                  },
                  {
                    type: 'select',
                    name: 'base_currency',
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
                type: 'container',
                visibleOn: "${rule_type === 'volume_discount'}",
                body: [
                  {
                    type: 'static',
                    label: '阶梯价格配置',
                    tpl: '根据购买数量设置不同价格'
                  },
                  {
                    type: 'input-array',
                    name: 'volume_tiers',
                    label: '阶梯配置',
                    items: {
                      type: 'container',
                      body: [
                        {
                          type: 'input-number',
                          name: 'min_quantity',
                          label: '最小数量',
                          required: true,
                          min: 1
                        },
                        {
                          type: 'input-number',
                          name: 'max_quantity',
                          label: '最大数量',
                          description: '0表示无限制'
                        },
                        {
                          type: 'input-number',
                          name: 'price_amount',
                          label: '价格',
                          required: true,
                          precision: 2,
                          min: 0
                        }
                      ]
                    }
                  }
                ]
              },
              {
                type: 'container',
                visibleOn: "${rule_type === 'member_discount'}",
                body: [
                  {
                    type: 'static',
                    label: '会员价格配置',
                    tpl: '为不同会员等级设置专属价格'
                  },
                  {
                    type: 'combo',
                    name: 'member_levels',
                    label: '会员等级价格',
                    multiple: true,
                    items: [
                      {
                        type: 'select',
                        name: 'level',
                        label: '会员等级',
                        options: [
                          { label: '普通会员', value: 'regular' },
                          { label: '银牌会员', value: 'silver' },
                          { label: '金牌会员', value: 'gold' },
                          { label: 'VIP会员', value: 'vip' }
                        ]
                      },
                      {
                        type: 'input-number',
                        name: 'price_amount',
                        label: '价格',
                        precision: 2,
                        min: 0
                      }
                    ]
                  },
                  {
                    type: 'input-number',
                    name: 'default_price_amount',
                    label: '默认价格',
                    description: '非会员价格',
                    precision: 2,
                    min: 0
                  }
                ]
              },
              {
                type: 'container',
                visibleOn: "${rule_type === 'time_based_discount'}",
                body: [
                  {
                    type: 'static',
                    label: '时段优惠配置',
                    tpl: '在特定时间段内提供优惠价格'
                  },
                  {
                    type: 'input-array',
                    name: 'time_slots',
                    label: '时段配置',
                    items: {
                      type: 'container',
                      body: [
                        {
                          type: 'input-time',
                          name: 'start_time',
                          label: '开始时间',
                          format: 'HH:mm'
                        },
                        {
                          type: 'input-time',
                          name: 'end_time',
                          label: '结束时间',
                          format: 'HH:mm'
                        },
                        {
                          type: 'checkboxes',
                          name: 'week_days',
                          label: '适用星期',
                          options: [
                            { label: '周日', value: 0 },
                            { label: '周一', value: 1 },
                            { label: '周二', value: 2 },
                            { label: '周三', value: 3 },
                            { label: '周四', value: 4 },
                            { label: '周五', value: 5 },
                            { label: '周六', value: 6 }
                          ]
                        },
                        {
                          type: 'input-number',
                          name: 'price_amount',
                          label: '优惠价格',
                          precision: 2,
                          min: 0
                        }
                      ]
                    }
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
      api: '/api/v1/products/${productId}/pricing-rules',
      headerToolbar: [
        {
          type: 'bulk-actions',
          align: 'left'
        },
        {
          type: 'pagination',
          align: 'right'
        }
      ],
      bulkActions: [
        {
          label: '批量启用',
          actionType: 'ajax',
          api: {
            method: 'put',
            url: '/api/v1/pricing-rules/batch-status',
            data: {
              rule_ids: '${ids}',
              is_active: true
            }
          },
          confirmText: '确认启用选中的定价规则？'
        },
        {
          label: '批量禁用',
          actionType: 'ajax',
          api: {
            method: 'put',
            url: '/api/v1/pricing-rules/batch-status',
            data: {
              rule_ids: '${ids}',
              is_active: false
            }
          },
          confirmText: '确认禁用选中的定价规则？'
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
          name: 'rule_type',
          label: '规则类型',
          type: 'mapping',
          map: {
            'base_price': '<span class="label label-info">基础价格</span>',
            'volume_discount': '<span class="label label-success">阶梯价格</span>',
            'member_discount': '<span class="label label-warning">会员价格</span>',
            'time_based_discount': '<span class="label label-primary">时段优惠</span>'
          }
        },
        {
          name: 'priority',
          label: '优先级',
          type: 'text',
          width: 80
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
          name: 'valid_from',
          label: '生效时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss'
        },
        {
          name: 'valid_until',
          label: '失效时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss',
          placeholder: '永久有效'
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
          width: 200,
          buttons: [
            {
              label: '编辑',
              type: 'button',
              actionType: 'dialog',
              dialog: {
                title: '编辑定价规则',
                size: 'lg',
                body: {
                  type: 'form',
                  api: {
                    method: 'put',
                    url: '/api/v1/products/${productId}/pricing-rules/${id}',
                    messages: {
                      success: '定价规则更新成功',
                      failed: '定价规则更新失败'
                    }
                  },
                  initApi: '/api/v1/products/${productId}/pricing-rules/${id}',
                  controls: [
                    {
                      type: 'input-number',
                      name: 'priority',
                      label: '优先级',
                      min: 0,
                      max: 100
                    },
                    {
                      type: 'switch',
                      name: 'is_active',
                      label: '启用状态'
                    },
                    {
                      type: 'datetime',
                      name: 'valid_from',
                      label: '生效时间',
                      format: 'YYYY-MM-DD HH:mm:ss'
                    },
                    {
                      type: 'datetime',
                      name: 'valid_until',
                      label: '失效时间',
                      format: 'YYYY-MM-DD HH:mm:ss'
                    }
                  ]
                }
              }
            },
            {
              label: '计算价格',
              type: 'button',
              actionType: 'dialog',
              dialog: {
                title: '价格计算预览',
                size: 'md',
                body: {
                  type: 'form',
                  api: {
                    method: 'get',
                    url: '/api/v1/products/${productId}/effective-price',
                    adaptor: 'return {\n  ...payload,\n  data: payload.data || {}\n};'
                  },
                  controls: [
                    {
                      type: 'input-number',
                      name: 'quantity',
                      label: '购买数量',
                      value: 1,
                      min: 1
                    },
                    {
                      type: 'select',
                      name: 'member_level',
                      label: '会员等级',
                      options: [
                        { label: '普通会员', value: 'regular' },
                        { label: '银牌会员', value: 'silver' },
                        { label: '金牌会员', value: 'gold' },
                        { label: 'VIP会员', value: 'vip' }
                      ]
                    },
                    {
                      type: 'divider'
                    },
                    {
                      type: 'static',
                      label: '基础价格',
                      tpl: '¥${base_price.amount | number:2}'
                    },
                    {
                      type: 'static',
                      label: '有效价格',
                      tpl: '¥${effective_price.amount | number:2}'
                    },
                    {
                      type: 'static',
                      label: '折扣金额',
                      tpl: '¥${discount_amount.amount | number:2}',
                      visibleOn: '${discount_amount.amount > 0}'
                    },
                    {
                      type: 'static',
                      label: '应用规则',
                      tpl: '${applied_rules | join:","}'
                    },
                    {
                      type: 'static',
                      label: '权益消耗',
                      tpl: '${rights_consumption}积分',
                      visibleOn: '${rights_consumption > 0}'
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
              confirmText: '确认删除此定价规则？',
              api: {
                method: 'delete',
                url: '/api/v1/products/${productId}/pricing-rules/${id}',
                messages: {
                  success: '定价规则删除成功',
                  failed: '定价规则删除失败'
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

export default PricingConfigPage;
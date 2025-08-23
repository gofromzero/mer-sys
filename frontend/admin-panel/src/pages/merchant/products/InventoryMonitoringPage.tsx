// 库存监控仪表板页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const InventoryMonitoringPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '库存监控',
    body: [
      // 统计卡片
      {
        type: 'grid',
        className: 'mb-4',
        columns: [
          {
            type: 'card',
            className: 'col-sm-6 col-md-3',
            body: {
              type: 'service',
              api: '/api/v1/inventory/monitoring',
              body: {
                type: 'tpl',
                tpl: `
                  <div class="text-center">
                    <h2 class="text-info mb-1">\${statistics.total_products}</h2>
                    <p class="text-muted mb-0">总商品数</p>
                  </div>
                `
              }
            }
          },
          {
            type: 'card',
            className: 'col-sm-6 col-md-3',
            body: {
              type: 'service',
              api: '/api/v1/inventory/monitoring',
              body: {
                type: 'tpl',
                tpl: `
                  <div class="text-center">
                    <h2 class="text-warning mb-1">\${statistics.low_stock_products}</h2>
                    <p class="text-muted mb-0">低库存商品</p>
                  </div>
                `
              }
            }
          },
          {
            type: 'card',
            className: 'col-sm-6 col-md-3',
            body: {
              type: 'service',
              api: '/api/v1/inventory/monitoring',
              body: {
                type: 'tpl',
                tpl: `
                  <div class="text-center">
                    <h2 class="text-danger mb-1">\${statistics.out_of_stock_products}</h2>
                    <p class="text-muted mb-0">缺货商品</p>
                  </div>
                `
              }
            }
          },
          {
            type: 'card',
            className: 'col-sm-6 col-md-3',
            body: {
              type: 'service',
              api: '/api/v1/inventory/monitoring',
              body: {
                type: 'tpl',
                tpl: `
                  <div class="text-center">
                    <h2 class="text-success mb-1">\${statistics.active_alerts}</h2>
                    <p class="text-muted mb-0">活跃预警</p>
                  </div>
                `
              }
            }
          }
        ]
      },

      // 库存健康分析
      {
        type: 'grid',
        className: 'mb-4',
        columns: [
          {
            type: 'panel',
            title: '库存健康分析',
            className: 'col-md-8',
            body: {
              type: 'service',
              api: '/api/v1/inventory/health-analysis',
              body: [
                {
                  type: 'tpl',
                  tpl: `
                    <div class="row">
                      <div class="col-md-6">
                        <h4>健康评分: <span class="text-\${health_level === 'excellent' ? 'success' : health_level === 'good' ? 'info' : health_level === 'fair' ? 'warning' : 'danger'}">\${health_score}/100</span></h4>
                        <p>健康等级: \${health_level === 'excellent' ? '优秀' : health_level === 'good' ? '良好' : health_level === 'fair' ? '一般' : '较差'}</p>
                      </div>
                      <div class="col-md-6">
                        <p>低库存比例: \${(stock_distribution.low_stock_ratio * 100).toFixed(2)}%</p>
                        <p>缺货比例: \${(stock_distribution.out_of_stock_ratio * 100).toFixed(2)}%</p>
                      </div>
                    </div>
                  `
                },
                {
                  type: 'chart',
                  api: '/api/v1/inventory/health-analysis',
                  config: {
                    type: 'pie',
                    data: {
                      source: '${stock_distribution}',
                      transform: {
                        type: 'map',
                        callback: (data: any) => [
                          { name: '正常库存', value: data.normal_stock },
                          { name: '低库存', value: data.low_stock },
                          { name: '缺货', value: data.out_of_stock }
                        ]
                      }
                    },
                    encode: {
                      itemName: 'name',
                      value: 'value'
                    }
                  }
                }
              ]
            }
          },
          {
            type: 'panel',
            title: '今日操作统计',
            className: 'col-md-4',
            body: {
              type: 'service',
              api: '/api/v1/inventory/monitoring',
              body: {
                type: 'tpl',
                tpl: `
                  <div class="text-center">
                    <h3 class="text-primary">\${statistics.today_changes}</h3>
                    <p class="text-muted">次库存变更</p>
                    <hr>
                    <p>库存总价值</p>
                    <h4 class="text-success">¥\${(statistics.total_inventory_value / 100).toFixed(2)}</h4>
                  </div>
                `
              }
            }
          }
        ]
      },

      // 活跃预警和最近变更
      {
        type: 'grid',
        columns: [
          {
            type: 'panel',
            title: '活跃预警',
            className: 'col-md-6',
            body: {
              type: 'service',
              api: '/api/v1/inventory/alerts/active',
              body: {
                type: 'table',
                source: '${data.alerts}',
                columns: [
                  {
                    name: 'product_name',
                    label: '商品'
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
                    label: '阈值'
                  },
                  {
                    name: 'last_triggered_at',
                    label: '触发时间',
                    type: 'datetime',
                    format: 'MM-DD HH:mm'
                  }
                ]
              }
            }
          },
          {
            type: 'panel',
            title: '最近变更',
            className: 'col-md-6',
            body: {
              type: 'service',
              api: '/api/v1/inventory/monitoring',
              body: {
                type: 'table',
                source: '${recent_changes}',
                columns: [
                  {
                    name: 'product_name',
                    label: '商品'
                  },
                  {
                    name: 'change_type',
                    label: '类型',
                    type: 'mapping',
                    map: {
                      'purchase': '<span class="text-success">入库</span>',
                      'sale': '<span class="text-info">出库</span>',
                      'adjustment': '<span class="text-warning">调整</span>'
                    }
                  },
                  {
                    name: 'quantity_changed',
                    label: '数量',
                    type: 'tpl',
                    tpl: '<%= data.quantity_changed > 0 ? "+" + data.quantity_changed : data.quantity_changed %>',
                    classNameExpr: '<%= data.quantity_changed > 0 ? "text-success" : "text-danger" %>'
                  },
                  {
                    name: 'created_at',
                    label: '时间',
                    type: 'datetime',
                    format: 'MM-DD HH:mm'
                  }
                ]
              }
            }
          }
        ]
      }
    ]
  };

  return <AmisRenderer schema={schema} />;
};

export default InventoryMonitoringPage;
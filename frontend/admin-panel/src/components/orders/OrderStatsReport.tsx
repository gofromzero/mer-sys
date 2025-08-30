import React, { useState, useEffect } from 'react';
import { AmisRenderer } from '../ui/AmisRenderer';
import type { SchemaNode } from 'amis';

interface OrderStatsReportProps {
  /** 商户ID */
  merchantId?: number;
  /** 统计时间范围类型 */
  timeRange?: 'today' | 'week' | 'month' | 'quarter' | 'year' | 'custom';
  /** 自定义时间范围 */
  customRange?: {
    start: string;
    end: string;
  };
  /** 是否显示详细报表 */
  showDetailedReport?: boolean;
}

const OrderStatsReport: React.FC<OrderStatsReportProps> = ({
  merchantId,
  timeRange = 'month',
  customRange,
  showDetailedReport = true,
}) => {
  const [activeTab, setActiveTab] = useState('overview');

  // 构建API查询参数
  const buildApiParams = () => {
    const params: any = {};
    
    if (merchantId) {
      params.merchant_id = merchantId;
    }

    // 根据时间范围设置日期参数
    const now = new Date();
    let startDate: Date;
    let endDate: Date = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59);

    switch (timeRange) {
      case 'today':
        startDate = new Date(now.getFullYear(), now.getMonth(), now.getDate());
        break;
      case 'week':
        startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case 'month':
        startDate = new Date(now.getFullYear(), now.getMonth(), 1);
        break;
      case 'quarter':
        const quarterStart = Math.floor(now.getMonth() / 3) * 3;
        startDate = new Date(now.getFullYear(), quarterStart, 1);
        break;
      case 'year':
        startDate = new Date(now.getFullYear(), 0, 1);
        break;
      case 'custom':
        if (customRange) {
          startDate = new Date(customRange.start);
          endDate = new Date(customRange.end);
        } else {
          startDate = new Date(now.getFullYear(), now.getMonth(), 1);
        }
        break;
      default:
        startDate = new Date(now.getFullYear(), now.getMonth(), 1);
    }

    params.start_date = startDate.toISOString().split('T')[0];
    params.end_date = endDate.toISOString().split('T')[0];

    return params;
  };

  const apiParams = buildApiParams();
  const apiUrl = `/api/v1/orders/stats?${new URLSearchParams(apiParams).toString()}`;

  const schema: SchemaNode = {
    type: 'page',
    className: 'order-stats-report',
    body: [
      // 时间范围选择器
      {
        type: 'form',
        className: 'bg-white rounded-lg shadow-sm border p-4 mb-4',
        body: [
          {
            type: 'button-group',
            name: 'time_range',
            label: '统计时间范围',
            value: timeRange,
            options: [
              { label: '今日', value: 'today' },
              { label: '近7天', value: 'week' },
              { label: '本月', value: 'month' },
              { label: '本季度', value: 'quarter' },
              { label: '本年', value: 'year' },
              { label: '自定义', value: 'custom' },
            ],
          },
          {
            type: 'input-date-range',
            name: 'custom_range',
            label: '自定义时间范围',
            visibleOn: '${time_range === "custom"}',
            format: 'YYYY-MM-DD',
            required: true,
          },
          {
            type: 'submit',
            label: '更新统计',
            level: 'primary',
          },
        ],
      },
      // 统计概览卡片
      {
        type: 'service',
        api: apiUrl,
        body: {
          type: 'grid',
          columns: [
            {
              md: 3,
              body: {
                type: 'card',
                className: 'stats-card bg-gradient-to-r from-blue-50 to-blue-100 border-blue-200',
                header: {
                  title: '订单总数',
                  subTitle: '统计期间内的总订单数',
                },
                body: [
                  {
                    type: 'tpl',
                    className: 'text-center',
                    tpl: '<div class="text-4xl font-bold text-blue-600 mb-2">${total}</div>',
                  },
                  {
                    type: 'progress',
                    value: '${(total / (total + 1)) * 100}',
                    strokeWidth: 6,
                    showLabel: false,
                  },
                ],
              },
            },
            {
              md: 3,
              body: {
                type: 'card',
                className: 'stats-card bg-gradient-to-r from-green-50 to-green-100 border-green-200',
                header: {
                  title: '销售金额',
                  subTitle: '已完成订单的总金额',
                },
                body: [
                  {
                    type: 'service',
                    api: `/api/v1/orders/query?${new URLSearchParams({...apiParams, status: '4'}).toString()}`,
                    body: {
                      type: 'tpl',
                      className: 'text-center',
                      tpl: '<div class="text-4xl font-bold text-green-600 mb-2">¥${items|sum:total_amount}</div>',
                    },
                  },
                  {
                    type: 'progress',
                    value: 85,
                    strokeWidth: 6,
                    showLabel: false,
                  },
                ],
              },
            },
            {
              md: 3,
              body: {
                type: 'card',
                className: 'stats-card bg-gradient-to-r from-orange-50 to-orange-100 border-orange-200',
                header: {
                  title: '待处理订单',
                  subTitle: '待支付+已支付订单数',
                },
                body: [
                  {
                    type: 'tpl',
                    className: 'text-center',
                    tpl: '<div class="text-4xl font-bold text-orange-600 mb-2">${by_status.pending + by_status.paid}</div>',
                  },
                  {
                    type: 'progress',
                    value: '${((by_status.pending + by_status.paid) / total) * 100}',
                    strokeWidth: 6,
                    showLabel: false,
                  },
                ],
              },
            },
            {
              md: 3,
              body: {
                type: 'card',
                className: 'stats-card bg-gradient-to-r from-purple-50 to-purple-100 border-purple-200',
                header: {
                  title: '完成率',
                  subTitle: '订单完成率统计',
                },
                body: [
                  {
                    type: 'tpl',
                    className: 'text-center',
                    tpl: '<div class="text-4xl font-bold text-purple-600 mb-2">${Math.round((by_status.completed / total) * 100)}%</div>',
                  },
                  {
                    type: 'progress',
                    value: '${(by_status.completed / total) * 100}',
                    strokeWidth: 6,
                    showLabel: false,
                  },
                ],
              },
            },
          ],
        },
      },
      // 详细报表标签页
      showDetailedReport && {
        type: 'tabs',
        className: 'bg-white rounded-lg shadow-sm border',
        activeKey: activeTab,
        tabs: [
          {
            title: '状态分析',
            key: 'overview',
            body: {
              type: 'service',
              api: apiUrl,
              body: [
                {
                  type: 'grid',
                  columns: [
                    {
                      md: 6,
                      body: {
                        type: 'chart',
                        api: apiUrl,
                        config: {
                          type: 'doughnut',
                          data: {
                            labels: ['待支付', '已支付', '处理中', '已完成', '已取消'],
                            datasets: [{
                              data: [
                                '${by_status.pending}',
                                '${by_status.paid}',
                                '${by_status.processing}',
                                '${by_status.completed}',
                                '${by_status.cancelled}',
                              ],
                              backgroundColor: [
                                '#f59e0b', // 待支付 - 橙色
                                '#3b82f6', // 已支付 - 蓝色
                                '#8b5cf6', // 处理中 - 紫色
                                '#10b981', // 已完成 - 绿色
                                '#6b7280', // 已取消 - 灰色
                              ],
                            }],
                          },
                          options: {
                            responsive: true,
                            plugins: {
                              title: {
                                display: true,
                                text: '订单状态分布',
                              },
                              legend: {
                                position: 'bottom',
                              },
                            },
                          },
                        },
                      },
                    },
                    {
                      md: 6,
                      body: {
                        type: 'table',
                        source: [
                          { status: '待支付', count: '${by_status.pending}', color: 'warning' },
                          { status: '已支付', count: '${by_status.paid}', color: 'info' },
                          { status: '处理中', count: '${by_status.processing}', color: 'primary' },
                          { status: '已完成', count: '${by_status.completed}', color: 'success' },
                          { status: '已取消', count: '${by_status.cancelled}', color: 'secondary' },
                        ],
                        columns: [
                          {
                            name: 'status',
                            label: '订单状态',
                            type: 'text',
                          },
                          {
                            name: 'count',
                            label: '数量',
                            type: 'text',
                          },
                          {
                            name: 'percentage',
                            label: '占比',
                            type: 'tpl',
                            tpl: '${Math.round((count / total) * 100)}%',
                          },
                        ],
                        title: '状态统计详情',
                      },
                    },
                  ],
                },
              ],
            },
          },
          {
            title: '趋势分析',
            key: 'trends',
            body: {
              type: 'service',
              api: `/api/v1/orders/trends?${new URLSearchParams(apiParams).toString()}`,
              body: {
                type: 'chart',
                config: {
                  type: 'line',
                  data: {
                    labels: '${dates}',
                    datasets: [
                      {
                        label: '订单数量',
                        data: '${order_counts}',
                        borderColor: '#3b82f6',
                        backgroundColor: 'rgba(59, 130, 246, 0.1)',
                        tension: 0.1,
                      },
                      {
                        label: '销售金额',
                        data: '${sales_amounts}',
                        borderColor: '#10b981',
                        backgroundColor: 'rgba(16, 185, 129, 0.1)',
                        yAxisID: 'y1',
                        tension: 0.1,
                      },
                    ],
                  },
                  options: {
                    responsive: true,
                    interaction: {
                      mode: 'index',
                      intersect: false,
                    },
                    plugins: {
                      title: {
                        display: true,
                        text: '订单趋势分析',
                      },
                    },
                    scales: {
                      x: {
                        display: true,
                        title: {
                          display: true,
                          text: '日期',
                        },
                      },
                      y: {
                        type: 'linear',
                        display: true,
                        position: 'left',
                        title: {
                          display: true,
                          text: '订单数量',
                        },
                      },
                      y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        title: {
                          display: true,
                          text: '销售金额 (¥)',
                        },
                        grid: {
                          drawOnChartArea: false,
                        },
                      },
                    },
                  },
                },
              },
            },
          },
          {
            title: '客户分析',
            key: 'customers',
            body: {
              type: 'service',
              api: `/api/v1/orders/customer-stats?${new URLSearchParams(apiParams).toString()}`,
              body: [
                {
                  type: 'grid',
                  columns: [
                    {
                      md: 6,
                      body: {
                        type: 'table',
                        title: '客户订单排行榜（前10名）',
                        source: '${top_customers}',
                        columns: [
                          {
                            name: 'rank',
                            label: '排名',
                            type: 'text',
                            width: 60,
                          },
                          {
                            name: 'customer_name',
                            label: '客户姓名',
                            type: 'text',
                          },
                          {
                            name: 'order_count',
                            label: '订单数',
                            type: 'text',
                            width: 80,
                          },
                          {
                            name: 'total_amount',
                            label: '消费金额',
                            type: 'text',
                            width: 100,
                            tpl: '¥${total_amount}',
                          },
                        ],
                      },
                    },
                    {
                      md: 6,
                      body: {
                        type: 'chart',
                        config: {
                          type: 'bar',
                          data: {
                            labels: '${top_customers|pick:customer_name}',
                            datasets: [{
                              label: '订单数量',
                              data: '${top_customers|pick:order_count}',
                              backgroundColor: 'rgba(59, 130, 246, 0.8)',
                            }],
                          },
                          options: {
                            responsive: true,
                            plugins: {
                              title: {
                                display: true,
                                text: '客户订单数量分布',
                              },
                            },
                            scales: {
                              y: {
                                beginAtZero: true,
                              },
                            },
                          },
                        },
                      },
                    },
                  ],
                },
              ],
            },
          },
          {
            title: '导出报告',
            key: 'export',
            body: [
              {
                type: 'form',
                title: '导出统计报告',
                body: [
                  {
                    type: 'select',
                    name: 'export_format',
                    label: '导出格式',
                    value: 'excel',
                    options: [
                      { label: 'Excel表格', value: 'excel' },
                      { label: 'PDF报告', value: 'pdf' },
                      { label: 'CSV数据', value: 'csv' },
                    ],
                  },
                  {
                    type: 'checkboxes',
                    name: 'export_content',
                    label: '导出内容',
                    options: [
                      { label: '基础统计数据', value: 'basic_stats' },
                      { label: '状态分析', value: 'status_analysis' },
                      { label: '趋势数据', value: 'trend_data' },
                      { label: '客户分析', value: 'customer_analysis' },
                      { label: '详细订单列表', value: 'order_details' },
                    ],
                    value: ['basic_stats', 'status_analysis'],
                  },
                  {
                    type: 'submit',
                    label: '生成并下载报告',
                    level: 'primary',
                    api: {
                      method: 'post',
                      url: '/api/v1/orders/export-report',
                      data: {
                        ...apiParams,
                        format: '${export_format}',
                        content: '${export_content}',
                      },
                      responseType: 'blob',
                      downloadFileName: `订单统计报告_${new Date().toISOString().split('T')[0]}.${export_format === 'excel' ? 'xlsx' : export_format}`,
                    },
                  },
                ],
              },
            ],
          },
        ],
      },
    ].filter(Boolean), // 过滤掉 false 值
  };

  return <AmisRenderer schema={schema} />;
};

export default OrderStatsReport;
import React, { useEffect, useState } from 'react';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { useMonitoringStore } from '../../stores/monitoringStore';
import { useAuthStore } from '../../stores/authStore';
import { AlertSeverity, AlertType, TrendDirection } from '../../types/monitoring';

const RightsMonitoringDashboard: React.FC = () => {
  const { user } = useAuthStore();
  const { 
    dashboardData, 
    fetchDashboardData, 
    loading, 
    error,
    clearError 
  } = useMonitoringStore();

  const [refreshInterval, setRefreshInterval] = useState<NodeJS.Timeout | null>(null);

  // 初始化和自动刷新
  useEffect(() => {
    fetchDashboardData();

    // 每5分钟自动刷新一次
    const interval = setInterval(() => {
      fetchDashboardData();
    }, 5 * 60 * 1000);

    setRefreshInterval(interval);

    return () => {
      if (interval) {
        clearInterval(interval);
      }
    };
  }, [fetchDashboardData]);

  // 错误处理
  useEffect(() => {
    if (error) {
      console.error('监控仪表板错误:', error);
    }
  }, [error]);

  // 格式化数字显示
  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toFixed(2);
  };

  // 获取预警严重程度对应的颜色
  const getSeverityColor = (severity: AlertSeverity): string => {
    switch (severity) {
      case AlertSeverity.CRITICAL:
        return 'danger';
      case AlertSeverity.WARNING:
        return 'warning';
      case AlertSeverity.INFO:
        return 'info';
      default:
        return 'secondary';
    }
  };

  // 获取趋势方向对应的图标
  const getTrendIcon = (trend: TrendDirection): string => {
    switch (trend) {
      case TrendDirection.INCREASING:
        return 'fa fa-arrow-up text-danger';
      case TrendDirection.DECREASING:
        return 'fa fa-arrow-down text-success';
      case TrendDirection.STABLE:
        return 'fa fa-minus text-muted';
      default:
        return 'fa fa-question text-secondary';
    }
  };

  // 获取预警类型的中文名称
  const getAlertTypeName = (type: AlertType): string => {
    switch (type) {
      case AlertType.BALANCE_LOW:
        return '余额不足';
      case AlertType.BALANCE_CRITICAL:
        return '余额紧急';
      case AlertType.USAGE_SPIKE:
        return '使用激增';
      case AlertType.PREDICTED_DEPLETION:
        return '预计耗尽';
      default:
        return '未知类型';
    }
  };

  const dashboardSchema = {
    type: 'page',
    title: '权益监控仪表板',
    subTitle: '实时监控权益使用情况和预警信息',
    body: [
      // 错误提示
      error && {
        type: 'alert',
        level: 'danger',
        body: error,
        showCloseButton: true,
        onClose: clearError,
      },

      // 顶部统计卡片
      {
        type: 'grid',
        columns: [
          {
            type: 'panel',
            title: '商户总数',
            body: {
              type: 'tpl',
              tpl: `<div class="text-center">
                      <div class="text-3xl font-bold text-blue-600">${dashboardData?.total_merchants || 0}</div>
                      <div class="text-sm text-gray-500 mt-1">活跃商户</div>
                    </div>`,
            },
            className: 'bg-gradient-to-r from-blue-50 to-blue-100',
          },
          {
            type: 'panel',
            title: '活跃预警',
            body: {
              type: 'tpl',
              tpl: `<div class="text-center">
                      <div class="text-3xl font-bold text-red-600">${dashboardData?.active_alerts || 0}</div>
                      <div class="text-sm text-gray-500 mt-1">需要处理</div>
                    </div>`,
            },
            className: 'bg-gradient-to-r from-red-50 to-red-100',
          },
          {
            type: 'panel',
            title: '总权益余额',
            body: {
              type: 'tpl',
              tpl: `<div class="text-center">
                      <div class="text-3xl font-bold text-green-600">${formatNumber(dashboardData?.total_rights_balance || 0)}</div>
                      <div class="text-sm text-gray-500 mt-1">元</div>
                    </div>`,
            },
            className: 'bg-gradient-to-r from-green-50 to-green-100',
          },
          {
            type: 'panel',
            title: '今日消费',
            body: {
              type: 'tpl',
              tpl: `<div class="text-center">
                      <div class="text-3xl font-bold text-purple-600">${formatNumber(dashboardData?.daily_consumption || 0)}</div>
                      <div class="text-sm text-gray-500 mt-1">元</div>
                    </div>`,
            },
            className: 'bg-gradient-to-r from-purple-50 to-purple-100',
          },
        ],
      },

      // 主要内容区域
      {
        type: 'grid',
        columns: [
          // 左侧：消费趋势图表
          {
            md: 8,
            body: {
              type: 'panel',
              title: '权益消费趋势',
              body: {
                type: 'chart',
                config: {
                  title: {
                    text: '最近30天权益消费趋势',
                  },
                  tooltip: {
                    trigger: 'axis',
                  },
                  legend: {
                    data: ['消费金额', '分配金额'],
                  },
                  grid: {
                    left: '3%',
                    right: '4%',
                    bottom: '3%',
                    containLabel: true,
                  },
                  xAxis: {
                    type: 'category',
                    boundaryGap: false,
                    data: dashboardData?.consumption_chart_data?.map(item => item.date) || [],
                  },
                  yAxis: {
                    type: 'value',
                    axisLabel: {
                      formatter: (value: number) => formatNumber(value),
                    },
                  },
                  series: [
                    {
                      name: '消费金额',
                      type: 'line',
                      smooth: true,
                      itemStyle: { color: '#f56565' },
                      data: dashboardData?.consumption_chart_data?.map(item => item.consumed) || [],
                    },
                    {
                      name: '分配金额',
                      type: 'line',
                      smooth: true,
                      itemStyle: { color: '#48bb78' },
                      data: dashboardData?.consumption_chart_data?.map(item => item.allocated) || [],
                    },
                  ],
                },
                height: 400,
              },
            },
          },

          // 右侧：余额分布
          {
            md: 4,
            body: {
              type: 'panel',
              title: '商户余额分布',
              body: {
                type: 'each',
                items: dashboardData?.balance_distribution || [],
                name: 'item',
                body: {
                  type: 'card',
                  body: [
                    {
                      type: 'tpl',
                      tpl: `<div class="flex justify-between items-center">
                              <div>
                                <div class="font-semibold">\${item.merchant_name}</div>
                                <div class="text-sm text-gray-500">余额: ¥\${item.available_balance}</div>
                              </div>
                              <div class="text-right">
                                <div class="text-sm">使用率</div>
                                <div class="text-lg font-bold \${item.status === 'critical' ? 'text-red-600' : item.status === 'warning' ? 'text-yellow-600' : 'text-green-600'}">
                                  \${item.usage_percentage}%
                                </div>
                              </div>
                            </div>`,
                    },
                    {
                      type: 'progress',
                      value: '${item.usage_percentage}',
                      mode: 'line',
                      strokeWidth: 6,
                      map: {
                        '${item.status === "critical"}': 'danger',
                        '${item.status === "warning"}': 'warning',
                        '1': 'success',
                      },
                    },
                  ],
                  className: 'mb-3',
                },
              },
            },
          },
        ],
      },

      // 最近预警
      {
        type: 'panel',
        title: '最近预警',
        body: {
          type: 'table',
          source: '${recent_alerts}',
          columns: [
            {
              name: 'alert_type',
              label: '预警类型',
              type: 'mapping',
              map: {
                'balance_low': `<span class="badge badge-warning">余额不足</span>`,
                'balance_critical': `<span class="badge badge-danger">余额紧急</span>`,
                'usage_spike': `<span class="badge badge-info">使用激增</span>`,
                'predicted_depletion': `<span class="badge badge-warning">预计耗尽</span>`,
              },
            },
            {
              name: 'message',
              label: '预警信息',
              type: 'text',
            },
            {
              name: 'severity',
              label: '严重程度',
              type: 'mapping',
              map: {
                'info': `<span class="badge badge-info">信息</span>`,
                'warning': `<span class="badge badge-warning">警告</span>`,
                'critical': `<span class="badge badge-danger">紧急</span>`,
              },
            },
            {
              name: 'triggered_at',
              label: '触发时间',
              type: 'datetime',
              format: 'YYYY-MM-DD HH:mm:ss',
            },
            {
              name: 'status',
              label: '状态',
              type: 'mapping',
              map: {
                'active': `<span class="badge badge-danger">活跃</span>`,
                'resolved': `<span class="badge badge-success">已解决</span>`,
                'acknowledged': `<span class="badge badge-info">已确认</span>`,
              },
            },
          ],
          actions: [
            {
              type: 'button',
              label: '查看详情',
              level: 'link',
              actionType: 'dialog',
              dialog: {
                title: '预警详情',
                body: {
                  type: 'form',
                  body: [
                    {
                      type: 'static',
                      name: 'message',
                      label: '预警信息',
                    },
                    {
                      type: 'static',
                      name: 'current_value',
                      label: '当前值',
                    },
                    {
                      type: 'static',
                      name: 'threshold_value',
                      label: '阈值',
                    },
                    {
                      type: 'static',
                      name: 'triggered_at',
                      label: '触发时间',
                    },
                  ],
                },
              },
            },
          ],
        },
        data: {
          recent_alerts: dashboardData?.recent_alerts || [],
        },
      },

      // 操作按钮
      {
        type: 'flex',
        justify: 'center',
        className: 'mt-6',
        items: [
          {
            type: 'button',
            label: '预警配置',
            level: 'primary',
            actionType: 'link',
            link: '/monitoring/alerts/config',
            icon: 'fa fa-cog',
          },
          {
            type: 'button',
            label: '预警列表',
            level: 'default',
            actionType: 'link',
            link: '/monitoring/alerts',
            icon: 'fa fa-list',
          },
          {
            type: 'button',
            label: '使用报告',
            level: 'info',
            actionType: 'link',
            link: '/monitoring/reports',
            icon: 'fa fa-chart-line',
          },
          {
            type: 'button',
            label: '手动刷新',
            level: 'secondary',
            actionType: 'ajax',
            api: {
              method: 'get',
              url: 'javascript:void(0)',
            },
            onClick: () => {
              fetchDashboardData();
            },
            icon: 'fa fa-refresh',
            className: loading.dashboard ? 'is-loading' : '',
          },
        ],
      },
    ].filter(Boolean), // 过滤掉 error 为 null 的情况
  };

  return (
    <div className="h-full">
      <AmisRenderer 
        schema={dashboardSchema} 
        data={{ 
          ...dashboardData,
          user,
          loading: loading.dashboard,
        }} 
      />
    </div>
  );
};

export default RightsMonitoringDashboard;
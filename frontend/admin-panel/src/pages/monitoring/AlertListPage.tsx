import React, { useEffect } from 'react';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { useMonitoringStore } from '../../stores/monitoringStore';
import { useAuthStore } from '../../stores/authStore';
import { AlertType, AlertSeverity, AlertStatus } from '../../types/monitoring';

const AlertListPage: React.FC = () => {
  const { user } = useAuthStore();
  const { 
    alerts, 
    alertsTotal, 
    fetchAlerts, 
    resolveAlert, 
    loading, 
    error, 
    clearError,
    pagination,
    filters,
    setFilters,
    setPagination,
  } = useMonitoringStore();

  // 初始化数据
  useEffect(() => {
    fetchAlerts();
  }, [fetchAlerts, pagination, filters]);

  // 处理解决预警
  const handleResolveAlert = async (alertId: number, resolution: string) => {
    try {
      await resolveAlert(alertId, resolution);
      // 刷新列表
      await fetchAlerts();
    } catch (error) {
      console.error('解决预警失败:', error);
    }
  };

  const alertListSchema = {
    type: 'page',
    title: '预警管理',
    subTitle: '查看和处理权益预警事件',
    toolbar: [
      {
        type: 'button',
        label: '返回仪表板',
        level: 'info',
        actionType: 'link',
        link: '/monitoring/dashboard',
        icon: 'fa fa-arrow-left',
      },
      {
        type: 'button',
        label: '预警配置',
        level: 'primary',
        actionType: 'link',
        link: '/monitoring/alerts/config',
        icon: 'fa fa-cog',
      },
    ],
    body: [
      // 错误提示
      error && {
        type: 'alert',
        level: 'danger',
        body: error,
        showCloseButton: true,
        onClose: clearError,
      },

      // 筛选条件
      {
        type: 'form',
        mode: 'horizontal',
        wrapWithPanel: false,
        className: 'mb-4 bg-gray-50 p-4 rounded',
        body: [
          {
            type: 'grid',
            columns: [
              {
                md: 3,
                body: {
                  type: 'select',
                  name: 'alert_type',
                  label: '预警类型',
                  placeholder: '全部类型',
                  clearable: true,
                  options: [
                    { label: '余额不足', value: AlertType.BALANCE_LOW },
                    { label: '余额紧急', value: AlertType.BALANCE_CRITICAL },
                    { label: '使用激增', value: AlertType.USAGE_SPIKE },
                    { label: '预计耗尽', value: AlertType.PREDICTED_DEPLETION },
                  ],
                  value: filters.alert_type || '',
                  onChange: (value: AlertType) => {
                    setFilters({ alert_type: value || undefined });
                  },
                },
              },
              {
                md: 3,
                body: {
                  type: 'select',
                  name: 'severity',
                  label: '严重程度',
                  placeholder: '全部程度',
                  clearable: true,
                  options: [
                    { label: '信息', value: AlertSeverity.INFO },
                    { label: '警告', value: AlertSeverity.WARNING },
                    { label: '紧急', value: AlertSeverity.CRITICAL },
                  ],
                  value: filters.severity || '',
                  onChange: (value: AlertSeverity) => {
                    setFilters({ severity: value || undefined });
                  },
                },
              },
              {
                md: 3,
                body: {
                  type: 'select',
                  name: 'status',
                  label: '状态',
                  placeholder: '全部状态',
                  clearable: true,
                  options: [
                    { label: '活跃', value: AlertStatus.ACTIVE },
                    { label: '已解决', value: AlertStatus.RESOLVED },
                    { label: '已确认', value: AlertStatus.ACKNOWLEDGED },
                  ],
                  value: filters.status || '',
                  onChange: (value: AlertStatus) => {
                    setFilters({ status: value || undefined });
                  },
                },
              },
              {
                md: 3,
                body: {
                  type: 'select',
                  name: 'merchant_id',
                  label: '商户',
                  placeholder: '全部商户',
                  clearable: true,
                  searchable: true,
                  source: {
                    method: 'get',
                    url: '/api/v1/merchants',
                    adaptor: (payload: any) => {
                      return {
                        ...payload,
                        options: payload.data?.list?.map((merchant: any) => ({
                          label: merchant.name,
                          value: merchant.id,
                        })) || [],
                      };
                    },
                  },
                  value: filters.merchant_id || '',
                  onChange: (value: number) => {
                    setFilters({ merchant_id: value || undefined });
                  },
                },
              },
            ],
          },
          {
            type: 'grid',
            columns: [
              {
                md: 4,
                body: {
                  type: 'input-date-range',
                  name: 'dateRange',
                  label: '时间范围',
                  format: 'YYYY-MM-DD',
                  clearable: true,
                  onChange: (value: string[]) => {
                    if (value && value.length === 2) {
                      setFilters({ 
                        start_date: value[0], 
                        end_date: value[1] 
                      });
                    } else {
                      setFilters({ 
                        start_date: undefined, 
                        end_date: undefined 
                      });
                    }
                  },
                },
              },
              {
                md: 8,
                body: {
                  type: 'flex',
                  justify: 'end',
                  items: [
                    {
                      type: 'button',
                      label: '重置筛选',
                      level: 'default',
                      actionType: 'reset',
                      onClick: () => {
                        setFilters({});
                      },
                    },
                    {
                      type: 'button',
                      label: '刷新数据',
                      level: 'info',
                      actionType: 'ajax',
                      onClick: () => {
                        fetchAlerts();
                      },
                      className: loading.alerts ? 'is-loading' : '',
                      icon: 'fa fa-refresh',
                    },
                  ],
                },
              },
            ],
          },
        ],
      },

      // 预警列表
      {
        type: 'crud',
        api: {
          method: 'get',
          url: '/api/v1/monitoring/alerts',
          adaptor: (payload: any) => {
            return {
              ...payload,
              data: {
                items: alerts,
                total: alertsTotal,
              },
            };
          },
        },
        filter: false, // 禁用默认筛选，使用自定义筛选
        headerToolbar: [
          'statistics',
          'reload',
          {
            type: 'export-excel',
            label: '导出Excel',
            api: '/api/v1/monitoring/alerts/export',
          },
        ],
        footerToolbar: [
          'statistics',
          {
            type: 'pagination',
            layout: ['total', 'perPage', 'pager'],
            perPageAvailable: [10, 20, 50, 100],
            page: pagination.page,
            perPage: pagination.pageSize,
            total: alertsTotal,
            onPageChange: (page: number, perPage: number) => {
              setPagination(page, perPage);
            },
          },
        ],
        columns: [
          {
            name: 'id',
            label: 'ID',
            type: 'text',
            width: 80,
          },
          {
            name: 'alert_type',
            label: '预警类型',
            type: 'mapping',
            map: {
              [AlertType.BALANCE_LOW]: `<span class="badge badge-warning">余额不足</span>`,
              [AlertType.BALANCE_CRITICAL]: `<span class="badge badge-danger">余额紧急</span>`,
              [AlertType.USAGE_SPIKE]: `<span class="badge badge-info">使用激增</span>`,
              [AlertType.PREDICTED_DEPLETION]: `<span class="badge badge-warning">预计耗尽</span>`,
            },
            width: 120,
          },
          {
            name: 'merchant_name',
            label: '商户',
            type: 'text',
            source: '/api/v1/merchants/${merchant_id}',
            width: 150,
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
              [AlertSeverity.INFO]: `<span class="badge badge-info">信息</span>`,
              [AlertSeverity.WARNING]: `<span class="badge badge-warning">警告</span>`,
              [AlertSeverity.CRITICAL]: `<span class="badge badge-danger">紧急</span>`,
            },
            width: 100,
          },
          {
            name: 'current_value',
            label: '当前值',
            type: 'number',
            precision: 2,
            width: 100,
          },
          {
            name: 'threshold_value',
            label: '阈值',
            type: 'number',
            precision: 2,
            width: 100,
          },
          {
            name: 'status',
            label: '状态',
            type: 'mapping',
            map: {
              [AlertStatus.ACTIVE]: `<span class="badge badge-danger">活跃</span>`,
              [AlertStatus.RESOLVED]: `<span class="badge badge-success">已解决</span>`,
              [AlertStatus.ACKNOWLEDGED]: `<span class="badge badge-info">已确认</span>`,
            },
            width: 100,
          },
          {
            name: 'triggered_at',
            label: '触发时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm:ss',
            width: 160,
          },
          {
            name: 'notified_channels',
            label: '通知渠道',
            type: 'each',
            items: {
              type: 'tag',
              label: '${item}',
              color: 'processing',
            },
            width: 120,
          },
        ],
        rowSelection: {
          type: 'checkbox',
          keyField: 'id',
        },
        actions: [
          {
            type: 'button',
            label: '查看详情',
            level: 'info',
            size: 'sm',
            actionType: 'dialog',
            dialog: {
              title: '预警详情 - ${alert_type}',
              size: 'lg',
              body: {
                type: 'form',
                mode: 'horizontal',
                body: [
                  {
                    type: 'static',
                    name: 'id',
                    label: '预警ID',
                  },
                  {
                    type: 'static',
                    name: 'merchant_name',
                    label: '商户名称',
                  },
                  {
                    type: 'static',
                    name: 'alert_type',
                    label: '预警类型',
                  },
                  {
                    type: 'static',
                    name: 'message',
                    label: '预警信息',
                  },
                  {
                    type: 'static',
                    name: 'severity',
                    label: '严重程度',
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
                  {
                    type: 'static',
                    name: 'resolved_at',
                    label: '解决时间',
                    visibleOn: 'this.resolved_at',
                  },
                  {
                    type: 'static',
                    name: 'notified_channels',
                    label: '通知渠道',
                  },
                ],
              },
            },
          },
          {
            type: 'button',
            label: '解决',
            level: 'success',
            size: 'sm',
            actionType: 'dialog',
            visibleOn: 'this.status === "active"',
            dialog: {
              title: '解决预警',
              body: {
                type: 'form',
                body: [
                  {
                    type: 'static',
                    name: 'message',
                    label: '预警信息',
                  },
                  {
                    type: 'textarea',
                    name: 'resolution',
                    label: '解决方案',
                    placeholder: '请描述如何解决了这个预警...',
                    required: true,
                    minRows: 3,
                    maxRows: 6,
                  },
                ],
                actions: [
                  {
                    type: 'button',
                    label: '取消',
                    level: 'default',
                    actionType: 'cancel',
                  },
                  {
                    type: 'submit',
                    label: '确认解决',
                    level: 'primary',
                    className: loading.resolving ? 'is-loading' : '',
                  },
                ],
                api: {
                  method: 'post',
                  url: '/api/v1/monitoring/alerts/${id}/resolve',
                  adaptor: (payload: any, response: any, api: any) => {
                    const { id, resolution } = payload;
                    handleResolveAlert(parseInt(id), resolution);
                    return response;
                  },
                },
                messages: {
                  saveSuccess: '预警已成功解决！',
                  saveFailed: '解决预警失败，请重试。',
                },
              },
            },
          },
        ],
        source: alerts,
        loadDataOnce: true,
      },
    ].filter(Boolean),
  };

  return (
    <div className="h-full">
      <AmisRenderer 
        schema={alertListSchema} 
        data={{ 
          user,
          alerts,
          alertsTotal,
          loading: loading.alerts,
          pagination,
          filters,
        }} 
      />
    </div>
  );
};

export default AlertListPage;
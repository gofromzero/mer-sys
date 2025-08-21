import React, { useState } from 'react';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { useMonitoringStore } from '../../stores/monitoringStore';
import { useAuthStore } from '../../stores/authStore';
import { TimePeriod } from '../../types/monitoring';

const UsageReportPage: React.FC = () => {
  const { user } = useAuthStore();
  const { 
    generateReport, 
    fetchUsageStats,
    usageStats,
    loading, 
    error, 
    clearError 
  } = useMonitoringStore();

  const [reportHistory, setReportHistory] = useState<any[]>([]);

  // 处理报告生成
  const handleGenerateReport = async (formData: any) => {
    try {
      const result = await generateReport(formData);
      
      // 添加到报告历史
      const newReport = {
        id: Date.now(),
        filename: result.filename,
        download_url: result.download_url,
        generated_at: new Date().toISOString(),
        format: formData.format,
        period: formData.period,
        start_date: formData.start_date,
        end_date: formData.end_date,
        merchant_count: formData.merchant_ids?.length || 0,
      };
      
      setReportHistory(prev => [newReport, ...prev]);
      
      return { status: 0, msg: '报告生成成功！', data: result };
    } catch (error) {
      return { 
        status: 500, 
        msg: error instanceof Error ? error.message : '报告生成失败' 
      };
    }
  };

  const reportSchema = {
    type: 'page',
    title: '权益使用报告',
    subTitle: '生成和下载权益使用统计报告',
    toolbar: [
      {
        type: 'button',
        label: '返回仪表板',
        level: 'info',
        actionType: 'link',
        link: '/monitoring/dashboard',
        icon: 'fa fa-arrow-left',
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

      {
        type: 'grid',
        columns: [
          // 左侧：报告生成表单
          {
            md: 6,
            body: {
              type: 'panel',
              title: '生成新报告',
              body: {
                type: 'form',
                mode: 'horizontal',
                api: {
                  method: 'post',
                  url: 'javascript:void(0)',
                  adaptor: handleGenerateReport,
                },
                body: [
                  {
                    type: 'divider',
                    title: '报告设置',
                  },
                  {
                    type: 'radios',
                    name: 'period',
                    label: '统计周期',
                    required: true,
                    value: TimePeriod.MONTHLY,
                    options: [
                      { label: '日报', value: TimePeriod.DAILY },
                      { label: '周报', value: TimePeriod.WEEKLY },
                      { label: '月报', value: TimePeriod.MONTHLY },
                    ],
                    description: '选择报告的统计周期',
                  },
                  {
                    type: 'input-date-range',
                    name: 'date_range',
                    label: '时间范围',
                    required: true,
                    format: 'YYYY-MM-DD',
                    maxDate: '${TODAY()}',
                    description: '选择要统计的时间范围',
                    onChange: (value: string[]) => {
                      if (value && value.length === 2) {
                        return {
                          start_date: value[0],
                          end_date: value[1],
                        };
                      }
                      return {};
                    },
                  },

                  {
                    type: 'divider',
                    title: '商户选择',
                  },
                  {
                    type: 'radios',
                    name: 'merchant_scope',
                    label: '商户范围',
                    value: 'all',
                    options: [
                      { label: '全部商户', value: 'all' },
                      { label: '指定商户', value: 'specific' },
                    ],
                  },
                  {
                    type: 'select',
                    name: 'merchant_ids',
                    label: '选择商户',
                    multiple: true,
                    visibleOn: 'this.merchant_scope === "specific"',
                    required: true,
                    requiredOn: 'this.merchant_scope === "specific"',
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
                    searchable: true,
                    checkAll: true,
                    description: '可选择多个商户进行对比分析',
                  },

                  {
                    type: 'divider',
                    title: '报告格式',
                  },
                  {
                    type: 'radios',
                    name: 'format',
                    label: '导出格式',
                    required: true,
                    value: 'excel',
                    options: [
                      { label: 'Excel表格', value: 'excel' },
                      { label: 'PDF文档', value: 'pdf' },
                      { label: 'CSV数据', value: 'csv' },
                    ],
                  },

                  {
                    type: 'checkboxes',
                    name: 'report_sections',
                    label: '报告内容',
                    value: ['summary', 'trends', 'alerts'],
                    options: [
                      { label: '汇总统计', value: 'summary' },
                      { label: '趋势分析', value: 'trends' },
                      { label: '预警记录', value: 'alerts' },
                      { label: '商户对比', value: 'comparison' },
                      { label: '详细明细', value: 'details' },
                    ],
                    description: '选择要包含在报告中的内容模块',
                  },
                ],
                actions: [
                  {
                    type: 'reset',
                    label: '重置',
                    level: 'default',
                  },
                  {
                    type: 'submit',
                    label: '生成报告',
                    level: 'primary',
                    className: loading.generating ? 'is-loading' : '',
                    icon: 'fa fa-file-export',
                  },
                ],
                messages: {
                  saveSuccess: '报告生成成功！',
                  saveFailed: '报告生成失败，请检查参数后重试。',
                },
              },
            },
          },

          // 右侧：实时预览
          {
            md: 6,
            body: {
              type: 'panel',
              title: '数据预览',
              body: {
                type: 'service',
                api: {
                  method: 'get',
                  url: '/api/v1/monitoring/rights/stats',
                  adaptor: (payload: any) => {
                    fetchUsageStats({
                      period: TimePeriod.DAILY,
                      page: 1,
                      page_size: 10,
                    });
                    return { ...payload, preview_data: usageStats };
                  },
                },
                body: {
                  type: 'table',
                  source: '${preview_data}',
                  columns: [
                    {
                      name: 'stat_date',
                      label: '日期',
                      type: 'date',
                      format: 'YYYY-MM-DD',
                    },
                    {
                      name: 'total_consumed',
                      label: '消费金额',
                      type: 'number',
                      precision: 2,
                    },
                    {
                      name: 'average_daily_usage',
                      label: '日均使用',
                      type: 'number',
                      precision: 2,
                    },
                    {
                      name: 'usage_trend',
                      label: '趋势',
                      type: 'mapping',
                      map: {
                        'increasing': `<span class="text-red-500">↗ 上升</span>`,
                        'decreasing': `<span class="text-green-500">↘ 下降</span>`,
                        'stable': `<span class="text-gray-500">→ 平稳</span>`,
                      },
                    },
                  ],
                  placeholder: '暂无预览数据',
                },
              },
            },
          },
        ],
      },

      // 报告历史
      {
        type: 'panel',
        title: '报告历史',
        className: 'mt-6',
        body: {
          type: 'table',
          source: reportHistory,
          columns: [
            {
              name: 'filename',
              label: '报告文件',
              type: 'text',
            },
            {
              name: 'format',
              label: '格式',
              type: 'mapping',
              map: {
                'excel': `<span class="badge badge-success">Excel</span>`,
                'pdf': `<span class="badge badge-info">PDF</span>`,
                'csv': `<span class="badge badge-secondary">CSV</span>`,
              },
            },
            {
              name: 'period',
              label: '周期',
              type: 'mapping',
              map: {
                'daily': '日报',
                'weekly': '周报',
                'monthly': '月报',
              },
            },
            {
              name: 'start_date',
              label: '开始日期',
              type: 'date',
              format: 'YYYY-MM-DD',
            },
            {
              name: 'end_date',
              label: '结束日期',
              type: 'date',
              format: 'YYYY-MM-DD',
            },
            {
              name: 'merchant_count',
              label: '商户数量',
              type: 'number',
            },
            {
              name: 'generated_at',
              label: '生成时间',
              type: 'datetime',
              format: 'YYYY-MM-DD HH:mm:ss',
            },
          ],
          actions: [
            {
              type: 'button',
              label: '下载',
              level: 'primary',
              size: 'sm',
              actionType: 'url',
              url: '${download_url}',
              blank: true,
              icon: 'fa fa-download',
            },
            {
              type: 'button',
              label: '删除',
              level: 'danger',
              size: 'sm',
              actionType: 'ajax',
              confirmText: '确认要删除这个报告记录吗？',
              api: {
                method: 'delete',
                url: 'javascript:void(0)',
                adaptor: (payload: any, response: any, api: any) => {
                  const itemIndex = api.data.__index;
                  setReportHistory(prev => prev.filter((_, index) => index !== itemIndex));
                  return { status: 0, msg: '删除成功' };
                },
              },
              icon: 'fa fa-trash',
            },
          ],
          placeholder: '暂无报告历史',
          footerToolbar: [
            {
              type: 'statistics',
              text: '共 ${COUNT(items)} 个报告',
            },
          ],
        },
      },

      // 快速操作
      {
        type: 'panel',
        title: '快速操作',
        className: 'mt-6',
        body: {
          type: 'flex',
          justify: 'center',
          items: [
            {
              type: 'button',
              label: '生成本月报告',
              level: 'primary',
              actionType: 'ajax',
              api: {
                method: 'post',
                url: 'javascript:void(0)',
                data: {
                  period: TimePeriod.MONTHLY,
                  start_date: '${DATETOSTR(STARTOF(TODAY(), "month"), "YYYY-MM-DD")}',
                  end_date: '${DATETOSTR(TODAY(), "YYYY-MM-DD")}',
                  merchant_scope: 'all',
                  format: 'excel',
                  report_sections: ['summary', 'trends', 'alerts'],
                },
                adaptor: handleGenerateReport,
              },
              className: loading.generating ? 'is-loading' : '',
              icon: 'fa fa-calendar-alt',
            },
            {
              type: 'button',
              label: '生成昨日报告',
              level: 'info',
              actionType: 'ajax',
              api: {
                method: 'post',
                url: 'javascript:void(0)',
                data: {
                  period: TimePeriod.DAILY,
                  start_date: '${DATETOSTR(DATEMODIFY(TODAY(), "-1day"), "YYYY-MM-DD")}',
                  end_date: '${DATETOSTR(DATEMODIFY(TODAY(), "-1day"), "YYYY-MM-DD")}',
                  merchant_scope: 'all',
                  format: 'excel',
                  report_sections: ['summary', 'details'],
                },
                adaptor: handleGenerateReport,
              },
              className: loading.generating ? 'is-loading' : '',
              icon: 'fa fa-file-alt',
            },
            {
              type: 'button',
              label: '生成上周报告',
              level: 'secondary',
              actionType: 'ajax',
              api: {
                method: 'post',
                url: 'javascript:void(0)',
                data: {
                  period: TimePeriod.WEEKLY,
                  start_date: '${DATETOSTR(DATEMODIFY(STARTOF(TODAY(), "week"), "-1week"), "YYYY-MM-DD")}',
                  end_date: '${DATETOSTR(DATEMODIFY(ENDOF(TODAY(), "week"), "-1week"), "YYYY-MM-DD")}',
                  merchant_scope: 'all',
                  format: 'pdf',
                  report_sections: ['summary', 'trends', 'comparison'],
                },
                adaptor: handleGenerateReport,
              },
              className: loading.generating ? 'is-loading' : '',
              icon: 'fa fa-chart-line',
            },
          ],
        },
      },
    ].filter(Boolean),
  };

  return (
    <div className="h-full">
      <AmisRenderer 
        schema={reportSchema} 
        data={{ 
          user,
          usageStats,
          reportHistory,
          loading: loading.generating,
        }} 
      />
    </div>
  );
};

export default UsageReportPage;
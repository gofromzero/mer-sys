import React from 'react';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { useMonitoringStore } from '../../stores/monitoringStore';
import { useAuthStore } from '../../stores/authStore';

const AlertConfigurationPage: React.FC = () => {
  const { user } = useAuthStore();
  const { configureAlerts, loading, error, clearError } = useMonitoringStore();

  const configSchema = {
    type: 'page',
    title: '预警配置',
    subTitle: '设置权益余额预警阈值和通知方式',
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
        type: 'form',
        api: {
          method: 'post',
          url: '/api/v1/monitoring/alerts/configure',
          adaptor: (payload: any) => {
            // 通过 store 方法处理请求
            configureAlerts(payload).then(() => {
              // 成功处理
              window.location.hash = '#/monitoring/dashboard';
            }).catch(() => {
              // 错误已在 store 中处理
            });
            return payload;
          },
        },
        body: [
          {
            type: 'divider',
            title: '商户选择',
          },
          {
            type: 'select',
            name: 'merchant_id',
            label: '选择商户',
            required: true,
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
            placeholder: '请选择要配置预警的商户',
            clearable: false,
            searchable: true,
          },

          {
            type: 'divider',
            title: '预警阈值设置',
          },
          {
            type: 'grid',
            columns: [
              {
                md: 6,
                body: {
                  type: 'input-number',
                  name: 'warning_threshold',
                  label: '预警阈值',
                  placeholder: '输入预警阈值金额',
                  min: 0,
                  step: 100,
                  precision: 2,
                  description: '当可用余额低于此值时触发预警通知',
                  validations: {
                    minimum: 0,
                  },
                  validationErrors: {
                    minimum: '预警阈值不能小于0',
                  },
                },
              },
              {
                md: 6,
                body: {
                  type: 'input-number',
                  name: 'critical_threshold',
                  label: '紧急阈值',
                  placeholder: '输入紧急阈值金额',
                  min: 0,
                  step: 100,
                  precision: 2,
                  description: '当可用余额低于此值时触发紧急通知',
                  validations: {
                    minimum: 0,
                  },
                  validationErrors: {
                    minimum: '紧急阈值不能小于0',
                  },
                },
              },
            ],
          },

          {
            type: 'alert',
            level: 'info',
            body: [
              '阈值设置说明：',
              '• 预警阈值：用于日常监控，当余额接近不足时发送预警',
              '• 紧急阈值：用于紧急情况，当余额严重不足时发送紧急通知',
              '• 紧急阈值应小于预警阈值',
              '• 设置为空表示不启用该级别的预警',
            ].join('<br/>'),
            className: 'mb-4',
          },

          {
            type: 'divider',
            title: '通知方式配置',
          },
          {
            type: 'checkboxes',
            name: 'notification_channels',
            label: '通知渠道',
            options: [
              {
                label: '系统内通知',
                value: 'system',
              },
              {
                label: '邮件通知',
                value: 'email',
              },
              {
                label: '短信通知（仅紧急预警）',
                value: 'sms',
              },
            ],
            value: ['system', 'email'], // 默认选中
            description: '选择预警触发时的通知方式',
          },

          {
            type: 'divider',
            title: '高级设置',
          },
          {
            type: 'grid',
            columns: [
              {
                md: 6,
                body: {
                  type: 'switch',
                  name: 'enable_trend_analysis',
                  label: '启用趋势分析',
                  value: true,
                  description: '根据使用趋势预测权益耗尽时间',
                },
              },
              {
                md: 6,
                body: {
                  type: 'switch',
                  name: 'enable_usage_spike_detection',
                  label: '启用使用激增检测',
                  value: true,
                  description: '检测异常的权益使用激增情况',
                },
              },
            ],
          },

          {
            type: 'input-number',
            name: 'notification_frequency',
            label: '通知频率限制（小时）',
            value: 1,
            min: 0.5,
            max: 24,
            step: 0.5,
            description: '同一类型预警的最小通知间隔，防止通知轰炸',
          },
        ],
        actions: [
          {
            type: 'button',
            label: '取消',
            level: 'default',
            actionType: 'link',
            link: '/monitoring/dashboard',
          },
          {
            type: 'submit',
            label: '保存配置',
            level: 'primary',
            className: loading.configuring ? 'is-loading' : '',
          },
        ],
        redirect: '/monitoring/dashboard',
        messages: {
          saveSuccess: '预警配置保存成功！',
          saveFailed: '预警配置保存失败，请检查输入并重试。',
        },
      },

      // 当前配置预览
      {
        type: 'panel',
        title: '配置预览',
        className: 'mt-6',
        body: {
          type: 'service',
          api: {
            method: 'get',
            url: '/api/v1/monitoring/alerts/config/${merchant_id}',
            sendOn: 'this.merchant_id',
          },
          body: {
            type: 'table',
            source: '${config_list}',
            columns: [
              {
                name: 'alert_type',
                label: '预警类型',
                type: 'mapping',
                map: {
                  'balance_warning': '余额预警',
                  'balance_critical': '余额紧急',
                  'usage_spike': '使用激增',
                  'predicted_depletion': '预计耗尽',
                },
              },
              {
                name: 'threshold_value',
                label: '阈值',
                type: 'text',
              },
              {
                name: 'enabled',
                label: '启用状态',
                type: 'status',
                map: {
                  1: 'success',
                  0: 'fail',
                },
              },
              {
                name: 'notification_channels',
                label: '通知渠道',
                type: 'each',
                items: {
                  type: 'tag',
                  label: '${item}',
                  color: 'processing',
                },
              },
            ],
          },
        },
      },
    ].filter(Boolean),
  };

  return (
    <div className="h-full">
      <AmisRenderer 
        schema={configSchema} 
        data={{ 
          user,
          loading: loading.configuring,
        }} 
      />
    </div>
  );
};

export default AlertConfigurationPage;
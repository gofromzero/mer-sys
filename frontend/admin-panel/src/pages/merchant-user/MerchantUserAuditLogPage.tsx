import React, { useState } from 'react';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { PermissionGuard } from '../../components/ui/PermissionGuard';
import { useMerchantPermissions } from '../../hooks/useMerchantPermissions';

/**
 * 商户用户操作日志页面
 * 支持查看和筛选所有商户用户的操作历史
 */
export const MerchantUserAuditLogPage: React.FC = () => {
  const [refreshKey, setRefreshKey] = useState(0);
  const { getCurrentMerchantId, canManageUsers } = useMerchantPermissions();

  // 导出日志
  const handleExportLogs = async (filters: any) => {
    try {
      // 构建导出参数
      const exportParams = {
        ...filters,
        merchant_id: getCurrentMerchantId(),
        format: 'excel'
      };
      
      // 创建导出请求
      const queryString = new URLSearchParams(exportParams).toString();
      const exportUrl = `/api/v1/merchant-users/audit-log/export?${queryString}`;
      
      // 触发下载
      const link = document.createElement('a');
      link.href = exportUrl;
      link.download = `merchant_user_audit_log_${new Date().toISOString().split('T')[0]}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (error) {
      console.error('导出失败:', error);
    }
  };

  // Amis页面配置
  const pageSchema = {
    type: 'page',
    title: '商户用户操作日志',
    body: {
      type: 'crud',
      api: {
        method: 'get',
        url: '/api/v1/merchant-users/audit-log',
        data: {
          '&': '$$',
          merchant_id: getCurrentMerchantId() // 自动注入当前商户ID
        }
      },
      key: refreshKey,
      syncLocation: false,
      draggable: false,
      columns: [
        {
          name: 'timestamp',
          label: '操作时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm:ss',
          width: 150,
          sortable: true
        },
        {
          name: 'user_info.username',
          label: '操作用户',
          type: 'text',
          width: 120,
          searchable: true
        },
        {
          name: 'action',
          label: '操作类型',
          type: 'mapping',
          width: 120,
          map: {
            'create_user': '<span class="badge badge-success">创建用户</span>',
            'update_user': '<span class="badge badge-info">更新用户</span>',
            'delete_user': '<span class="badge badge-danger">删除用户</span>',
            'status_change': '<span class="badge badge-warning">状态变更</span>',
            'password_reset': '<span class="badge badge-primary">密码重置</span>',
            'permission_change': '<span class="badge badge-secondary">权限变更</span>',
            'login': '<span class="badge badge-success">登录</span>',
            'logout': '<span class="badge badge-light">登出</span>',
            'login_failed': '<span class="badge badge-danger">登录失败</span>'
          }
        },
        {
          name: 'resource',
          label: '操作对象',
          type: 'text',
          width: 120
        },
        {
          name: 'target_user.username',
          label: '目标用户',
          type: 'text',
          width: 120,
          searchable: true
        },
        {
          name: 'ip_address',
          label: 'IP地址',
          type: 'text',
          width: 120
        },
        {
          name: 'result',
          label: '操作结果',
          type: 'mapping',
          width: 100,
          map: {
            'success': '<span class="badge badge-success">成功</span>',
            'failed': '<span class="badge badge-danger">失败</span>',
            'partial': '<span class="badge badge-warning">部分成功</span>'
          }
        },
        {
          type: 'operation',
          label: '操作',
          width: 120,
          buttons: [
            {
              type: 'button',
              label: '详情',
              level: 'link',
              actionType: 'dialog',
              dialog: {
                title: '操作详情',
                size: 'lg',
                body: [
                  {
                    type: 'form',
                    mode: 'horizontal',
                    disabled: true,
                    body: [
                      {
                        type: 'input-text',
                        name: 'timestamp',
                        label: '操作时间'
                      },
                      {
                        type: 'input-text',
                        name: 'user_info.username',
                        label: '操作用户'
                      },
                      {
                        type: 'input-text',
                        name: 'action',
                        label: '操作类型'
                      },
                      {
                        type: 'input-text',
                        name: 'resource',
                        label: '操作对象'
                      },
                      {
                        type: 'input-text',
                        name: 'target_user.username',
                        label: '目标用户'
                      },
                      {
                        type: 'input-text',
                        name: 'ip_address',
                        label: 'IP地址'
                      },
                      {
                        type: 'input-text',
                        name: 'user_agent',
                        label: '客户端'
                      },
                      {
                        type: 'input-text',
                        name: 'result',
                        label: '操作结果'
                      },
                      {
                        type: 'json',
                        name: 'details',
                        label: '详细信息'
                      },
                      {
                        type: 'textarea',
                        name: 'comment',
                        label: '备注'
                      }
                    ]
                  }
                ]
              }
            }
          ]
        }
      ],
      headerToolbar: [
        'filter-toggler',
        {
          type: 'button',
          label: '导出日志',
          level: 'primary',
          icon: 'fa fa-download',
          visibleOn: canManageUsers(),
          onClick: (_e: any, props: any) => {
            handleExportLogs(props.query || {});
          }
        },
        {
          type: 'button',
          label: '刷新',
          level: 'default',
          icon: 'fa fa-refresh',
          onClick: () => {
            setRefreshKey(prev => prev + 1);
          }
        }
      ],
      filter: {
        title: '日志筛选',
        body: [
          {
            type: 'input-text',
            name: 'username',
            label: '操作用户',
            placeholder: '请输入操作用户名',
            clearable: true
          },
          {
            type: 'input-text',
            name: 'target_username',
            label: '目标用户',
            placeholder: '请输入目标用户名',
            clearable: true
          },
          {
            type: 'select',
            name: 'action',
            label: '操作类型',
            placeholder: '请选择操作类型',
            clearable: true,
            options: [
              { label: '创建用户', value: 'create_user' },
              { label: '更新用户', value: 'update_user' },
              { label: '删除用户', value: 'delete_user' },
              { label: '状态变更', value: 'status_change' },
              { label: '密码重置', value: 'password_reset' },
              { label: '权限变更', value: 'permission_change' },
              { label: '登录', value: 'login' },
              { label: '登出', value: 'logout' },
              { label: '登录失败', value: 'login_failed' }
            ]
          },
          {
            type: 'select',
            name: 'result',
            label: '操作结果',
            placeholder: '请选择操作结果',
            clearable: true,
            options: [
              { label: '成功', value: 'success' },
              { label: '失败', value: 'failed' },
              { label: '部分成功', value: 'partial' }
            ]
          },
          {
            type: 'input-datetime-range',
            name: 'time_range',
            label: '时间范围',
            placeholder: '请选择时间范围',
            clearable: true,
            format: 'YYYY-MM-DD HH:mm:ss'
          },
          {
            type: 'input-text',
            name: 'ip_address',
            label: 'IP地址',
            placeholder: '请输入IP地址',
            clearable: true
          },
          {
            type: 'input-text',
            name: 'search',
            label: '全文搜索',
            placeholder: '搜索用户名、IP、操作详情等',
            clearable: true
          }
        ]
      },
      perPageAvailable: [20, 50, 100, 200],
      defaultParams: {
        page: 1,
        page_size: 20,
        order_by: 'timestamp',
        order_direction: 'desc'
      }
    }
  };

  return (
    <PermissionGuard 
      anyPermissions={['merchant:user:view', 'merchant:user:manage']}
      fallback={<div className="text-center p-4">您没有权限访问操作日志</div>}
    >
      <div className="merchant-user-audit-log-page">
        <AmisRenderer schema={pageSchema} />
      </div>
    </PermissionGuard>
  );
};
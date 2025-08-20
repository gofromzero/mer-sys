import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { PermissionGuard } from '../../components/ui/PermissionGuard';
import { useMerchantPermissions } from '../../hooks/useMerchantPermissions';
import { MerchantUserService } from '../../services/merchantUserService';
import type { MerchantUser } from '../../types/merchantUser';

/**
 * 商户用户状态管理页面
 * 提供详细的用户状态管理和历史记录
 */
export const MerchantUserStatusManagePage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { canManageUsers } = useMerchantPermissions();
  const [user, setUser] = useState<MerchantUser | null>(null);
  const [auditLogs, setAuditLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  // 加载用户信息和审计日志
  useEffect(() => {
    if (id) {
      loadUserData();
    }
  }, [id]);

  const loadUserData = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const [userData, auditData] = await Promise.all([
        MerchantUserService.getMerchantUser(Number(id)),
        MerchantUserService.getMerchantUserAuditLog(Number(id), { page: 1, page_size: 20 })
      ]);
      
      setUser(userData);
      setAuditLogs(auditData.list);
    } catch (error) {
      console.error('加载用户数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 处理状态更新
  const handleStatusUpdate = async (status: string, comment?: string) => {
    if (!id) return;
    
    try {
      await MerchantUserService.updateMerchantUserStatus(Number(id), { 
        status: status as any, 
        comment 
      });
      await loadUserData(); // 重新加载数据
    } catch (error) {
      console.error('更新状态失败:', error);
      throw error;
    }
  };

  // 处理密码重置
  const handlePasswordReset = async (data: any) => {
    if (!id) return;
    
    try {
      await MerchantUserService.resetMerchantUserPassword(Number(id), data);
      await loadUserData(); // 重新加载数据
    } catch (error) {
      console.error('重置密码失败:', error);
      throw error;
    }
  };

  if (loading) {
    return <div className="text-center p-4">加载中...</div>;
  }

  if (!user) {
    return <div className="text-center p-4">用户不存在</div>;
  }

  // 页面配置
  const pageSchema = {
    type: 'page',
    title: `用户状态管理 - ${user.username}`,
    body: [
      {
        type: 'grid',
        columns: [
          {
            type: 'panel',
            title: '用户基本信息',
            body: [
              {
                type: 'property',
                items: [
                  { label: '用户名', content: user.username },
                  { label: '邮箱', content: user.email },
                  { label: '手机号', content: user.phone || '未设置' },
                  { 
                    label: '角色', 
                    content: user.role_type === 'merchant_admin' ? '商户管理员' : '商户操作员' 
                  },
                  {
                    label: '当前状态',
                    content: {
                      'pending': '<span class="badge badge-warning">待激活</span>',
                      'active': '<span class="badge badge-success">已激活</span>',
                      'suspended': '<span class="badge badge-secondary">已暂停</span>',
                      'deactivated': '<span class="badge badge-danger">已停用</span>'
                    }[user.status]
                  },
                  { label: '创建时间', content: user.created_at },
                  { label: '最后登录', content: user.last_login_at || '从未登录' }
                ]
              }
            ]
          },
          {
            type: 'panel',
            title: '状态管理操作',
            body: [
              {
                type: 'form',
                title: '状态变更',
                body: [
                  {
                    type: 'alert',
                    level: 'info',
                    body: '选择新的用户状态并填写变更原因。状态变更将立即生效并记录审计日志。'
                  },
                  {
                    type: 'select',
                    name: 'new_status',
                    label: '新状态',
                    required: true,
                    value: user.status,
                    options: [
                      { label: '激活', value: 'active', disabled: user.status === 'active' },
                      { label: '暂停', value: 'suspended', disabled: user.status === 'suspended' },
                      { label: '停用', value: 'deactivated', disabled: user.status === 'deactivated' }
                    ]
                  },
                  {
                    type: 'textarea',
                    name: 'comment',
                    label: '变更原因',
                    required: true,
                    placeholder: '请详细说明状态变更的原因',
                    maxLength: 500
                  }
                ],
                actions: [
                  {
                    type: 'button',
                    label: '确认变更',
                    level: 'primary',
                    actionType: 'submit',
                    disabled: !canManageUsers(),
                    onClick: async (_e: any, props: any) => {
                      const { new_status, comment } = props.scope;
                      try {
                        await handleStatusUpdate(new_status, comment);
                        props.setData('new_status', '');
                        props.setData('comment', '');
                      } catch (error) {
                        // 错误处理
                      }
                    }
                  }
                ]
              },
              {
                type: 'divider'
              },
              {
                type: 'form',
                title: '密码管理',
                body: [
                  {
                    type: 'alert',
                    level: 'warning',
                    body: '重置用户密码将强制用户下次登录时更改密码。'
                  },
                  {
                    type: 'switch',
                    name: 'send_email',
                    label: '发送邮件通知',
                    value: true
                  },
                  {
                    type: 'switch',
                    name: 'force_change',
                    label: '强制下次登录修改',
                    value: true
                  }
                ],
                actions: [
                  {
                    type: 'button',
                    label: '重置密码',
                    level: 'warning',
                    actionType: 'submit',
                    disabled: !canManageUsers(),
                    onClick: async (_e: any, props: any) => {
                      const { send_email, force_change } = props.scope;
                      try {
                        await handlePasswordReset({ 
                          send_email, 
                          force_change 
                        });
                      } catch (error) {
                        // 错误处理
                      }
                    }
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        type: 'divider'
      },
      {
        type: 'panel',
        title: '操作历史记录',
        body: [
          {
            type: 'table',
            source: auditLogs,
            columns: [
              {
                name: 'timestamp',
                label: '时间',
                type: 'datetime',
                format: 'YYYY-MM-DD HH:mm:ss',
                width: 150
              },
              {
                name: 'action',
                label: '操作类型',
                type: 'mapping',
                width: 120,
                map: {
                  'status_change': '<span class="badge badge-info">状态变更</span>',
                  'password_reset': '<span class="badge badge-warning">密码重置</span>',
                  'login': '<span class="badge badge-success">登录</span>',
                  'logout': '<span class="badge badge-secondary">登出</span>',
                  'permission_change': '<span class="badge badge-primary">权限变更</span>'
                }
              },
              {
                name: 'resource',
                label: '操作对象',
                type: 'text',
                width: 100
              },
              {
                name: 'details',
                label: '详细信息',
                type: 'json',
                width: 200
              },
              {
                name: 'ip_address',
                label: 'IP地址',
                type: 'text',
                width: 120
              },
              {
                name: 'user_agent',
                label: '客户端',
                type: 'text',
                width: 150,
                breakpoint: '*'
              }
            ],
            footerToolbar: [
              'pagination'
            ]
          }
        ]
      },
      {
        type: 'divider'
      },
      {
        type: 'button-group',
        buttons: [
          {
            type: 'button',
            label: '返回用户列表',
            level: 'default',
            onClick: () => {
              navigate('/merchant-user');
            }
          },
          {
            type: 'button',
            label: '编辑用户信息',
            level: 'primary',
            disabled: !canManageUsers(),
            onClick: () => {
              navigate(`/merchant-user/edit/${id}`);
            }
          }
        ]
      }
    ]
  };

  return (
    <PermissionGuard 
      anyPermissions={['merchant:user:view', 'merchant:user:manage']}
      fallback={<div className="text-center p-4">您没有权限访问此页面</div>}
    >
      <div className="merchant-user-status-manage-page">
        <AmisRenderer schema={pageSchema} />
      </div>
    </PermissionGuard>
  );
};
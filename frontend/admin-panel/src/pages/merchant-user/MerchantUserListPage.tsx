import React, { useState } from 'react';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { PermissionGuard } from '../../components/ui/PermissionGuard';
import { MerchantUserService } from '../../services/merchantUserService';
import { useMerchantPermissions } from '../../hooks/useMerchantPermissions';
import type { MerchantRoleType } from '../../types/merchantUser';
import type { UserStatus } from '../../types/user';

/**
 * 商户用户列表页面
 * 支持搜索、筛选、分页和用户管理
 */
export const MerchantUserListPage: React.FC = () => {
  const [refreshKey, setRefreshKey] = useState(0);
  const { canManageUsers, getCurrentMerchantId } = useMerchantPermissions();

  // 处理用户状态更新
  const handleStatusUpdate = async (id: number, status: UserStatus, comment?: string) => {
    try {
      await MerchantUserService.updateMerchantUserStatus(id, { status, comment });
      setRefreshKey(prev => prev + 1); // 刷新列表
    } catch (error) {
      console.error('更新商户用户状态失败:', error);
      throw error;
    }
  };

  // 处理密码重置
  const handlePasswordReset = async (id: number, sendEmail = true) => {
    try {
      await MerchantUserService.resetMerchantUserPassword(id, { send_email: sendEmail });
      setRefreshKey(prev => prev + 1); // 刷新列表
    } catch (error) {
      console.error('重置商户用户密码失败:', error);
      throw error;
    }
  };

  // 处理用户删除
  const handleDeleteUser = async (id: number) => {
    try {
      await MerchantUserService.deleteMerchantUser(id);
      setRefreshKey(prev => prev + 1); // 刷新列表
    } catch (error) {
      console.error('删除商户用户失败:', error);
      throw error;
    }
  };

  // Amis页面配置
  const pageSchema = {
    type: 'page',
    title: '商户用户管理',
    body: {
      type: 'crud',
      api: {
        method: 'get',
        url: '/api/v1/merchant-users',
        data: {
          '&': '$$',
          merchant_id: getCurrentMerchantId() // 自动注入当前商户ID
        }
      },
      key: refreshKey, // 用于强制刷新
      syncLocation: false,
      draggable: false,
      columns: [
        {
          name: 'username',
          label: '用户名',
          searchable: true,
          type: 'text',
          width: 120
        },
        {
          name: 'email',
          label: '邮箱',
          searchable: true,
          type: 'text',
          width: 180
        },
        {
          name: 'phone',
          label: '手机号',
          type: 'text',
          width: 120
        },
        {
          name: 'role_type',
          label: '角色',
          type: 'mapping',
          width: 100,
          map: {
            'merchant_admin': '<span class="badge badge-primary">商户管理员</span>',
            'merchant_operator': '<span class="badge badge-info">商户操作员</span>'
          }
        },
        {
          name: 'status',
          label: '状态',
          type: 'mapping',
          width: 100,
          map: {
            'pending': '<span class="badge badge-warning">待激活</span>',
            'active': '<span class="badge badge-success">已激活</span>',
            'suspended': '<span class="badge badge-secondary">已暂停</span>',
            'deactivated': '<span class="badge badge-danger">已停用</span>'
          }
        },
        {
          name: 'last_login_at',
          label: '最后登录',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm',
          width: 150
        },
        {
          name: 'created_at',
          label: '创建时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm',
          width: 150
        },
        {
          type: 'operation',
          label: '操作',
          width: 250,
          buttons: [
            {
              type: 'button',
              label: '查看',
              level: 'link',
              actionType: 'dialog',
              dialog: {
                title: '用户详情',
                size: 'lg',
                body: {
                  type: 'service',
                  api: '/api/v1/merchant-users/${id}',
                  body: [
                    {
                      type: 'form',
                      mode: 'horizontal',
                      disabled: true,
                      body: [
                        {
                          type: 'input-text',
                          name: 'username',
                          label: '用户名'
                        },
                        {
                          type: 'input-text',
                          name: 'email',
                          label: '邮箱'
                        },
                        {
                          type: 'input-text',
                          name: 'phone',
                          label: '手机号'
                        },
                        {
                          type: 'select',
                          name: 'role_type',
                          label: '角色类型',
                          options: [
                            { label: '商户管理员', value: 'merchant_admin' },
                            { label: '商户操作员', value: 'merchant_operator' }
                          ]
                        },
                        {
                          type: 'input-text',
                          name: 'status',
                          label: '状态'
                        },
                        {
                          type: 'datetime',
                          name: 'created_at',
                          label: '创建时间',
                          format: 'YYYY-MM-DD HH:mm:ss'
                        },
                        {
                          type: 'datetime',
                          name: 'last_login_at',
                          label: '最后登录',
                          format: 'YYYY-MM-DD HH:mm:ss'
                        }
                      ]
                    }
                  ]
                }
              }
            },
            {
              type: 'button',
              label: '编辑',
              level: 'primary',
              visibleOn: canManageUsers(),
              actionType: 'dialog',
              dialog: {
                title: '编辑用户',
                body: {
                  type: 'form',
                  api: {
                    method: 'put',
                    url: '/api/v1/merchant-users/${id}'
                  },
                  body: [
                    {
                      type: 'input-text',
                      name: 'username',
                      label: '用户名',
                      required: true
                    },
                    {
                      type: 'input-email',
                      name: 'email',
                      label: '邮箱',
                      required: true
                    },
                    {
                      type: 'input-text',
                      name: 'phone',
                      label: '手机号'
                    },
                    {
                      type: 'select',
                      name: 'role_type',
                      label: '角色类型',
                      required: true,
                      options: [
                        { label: '商户管理员', value: 'merchant_admin' },
                        { label: '商户操作员', value: 'merchant_operator' }
                      ]
                    }
                  ]
                },
                actions: [
                  {
                    type: 'button',
                    actionType: 'cancel',
                    label: '取消'
                  },
                  {
                    type: 'button',
                    actionType: 'submit',
                    label: '保存',
                    level: 'primary',
                    onClick: (_e: any, props: any) => {
                      setRefreshKey(prev => prev + 1);
                      props.onClose();
                    }
                  }
                ]
              }
            },
            {
              type: 'button',
              label: '状态管理',
              level: 'info',
              visibleOn: canManageUsers(),
              actionType: 'dialog',
              dialog: {
                title: '用户状态管理',
                body: {
                  type: 'form',
                  body: [
                    {
                      type: 'static',
                      name: 'username',
                      label: '用户名'
                    },
                    {
                      type: 'select',
                      name: 'status',
                      label: '新状态',
                      required: true,
                      options: [
                        { label: '激活', value: 'active' },
                        { label: '暂停', value: 'suspended' },
                        { label: '停用', value: 'deactivated' }
                      ]
                    },
                    {
                      type: 'textarea',
                      name: 'comment',
                      label: '变更原因',
                      placeholder: '请输入状态变更原因'
                    }
                  ]
                },
                actions: [
                  {
                    type: 'button',
                    actionType: 'cancel',
                    label: '取消'
                  },
                  {
                    type: 'button',
                    actionType: 'submit',
                    label: '确认',
                    level: 'primary',
                    onClick: async (_e: any, props: any) => {
                      const { status, comment } = props.scope;
                      const id = props.scope.id;
                      
                      try {
                        await handleStatusUpdate(id, status, comment);
                        props.onClose();
                      } catch (error) {
                        // 错误处理已在handleStatusUpdate中完成
                      }
                    }
                  }
                ]
              }
            },
            {
              type: 'button',
              label: '重置密码',
              level: 'warning',
              visibleOn: canManageUsers(),
              actionType: 'dialog',
              dialog: {
                title: '重置密码',
                body: {
                  type: 'form',
                  body: [
                    {
                      type: 'static',
                      name: 'username',
                      label: '用户名'
                    },
                    {
                      type: 'switch',
                      name: 'send_email',
                      label: '发送邮件通知',
                      value: true
                    },
                    {
                      type: 'alert',
                      level: 'info',
                      body: '系统将生成随机密码并通过邮件发送给用户。'
                    }
                  ]
                },
                actions: [
                  {
                    type: 'button',
                    actionType: 'cancel',
                    label: '取消'
                  },
                  {
                    type: 'button',
                    actionType: 'submit',
                    label: '重置',
                    level: 'warning',
                    onClick: async (_e: any, props: any) => {
                      const { send_email } = props.scope;
                      const id = props.scope.id;
                      
                      try {
                        await handlePasswordReset(id, send_email);
                        props.onClose();
                      } catch (error) {
                        // 错误处理已在handlePasswordReset中完成
                      }
                    }
                  }
                ]
              }
            },
            {
              type: 'button',
              label: '删除',
              level: 'danger',
              visibleOn: canManageUsers() + " && ${status !== 'active'}",
              actionType: 'dialog',
              dialog: {
                title: '删除用户',
                body: [
                  {
                    type: 'alert',
                    level: 'danger',
                    body: '此操作将永久删除该用户账号，确定要继续吗？'
                  },
                  {
                    type: 'static',
                    name: 'username',
                    label: '用户名'
                  }
                ],
                actions: [
                  {
                    type: 'button',
                    actionType: 'cancel',
                    label: '取消'
                  },
                  {
                    type: 'button',
                    actionType: 'submit',
                    label: '确认删除',
                    level: 'danger',
                    onClick: async (_e: any, props: any) => {
                      const id = props.scope.id;
                      
                      try {
                        await handleDeleteUser(id);
                        props.onClose();
                      } catch (error) {
                        // 错误处理已在handleDeleteUser中完成
                      }
                    }
                  }
                ]
              }
            }
          ]
        }
      ],
      headerToolbar: [
        'filter-toggler',
        'bulkActions',
        {
          type: 'button',
          label: '新增用户',
          level: 'primary',
          visibleOn: canManageUsers(),
          actionType: 'dialog',
          dialog: {
            title: '新增商户用户',
            size: 'lg',
            body: {
              type: 'form',
              api: {
                method: 'post',
                url: '/api/v1/merchant-users'
              },
              body: [
                {
                  type: 'input-text',
                  name: 'username',
                  label: '用户名',
                  required: true,
                  placeholder: '请输入用户名'
                },
                {
                  type: 'input-email',
                  name: 'email',
                  label: '邮箱',
                  required: true,
                  placeholder: '请输入邮箱地址'
                },
                {
                  type: 'input-text',
                  name: 'phone',
                  label: '手机号',
                  placeholder: '请输入手机号'
                },
                {
                  type: 'select',
                  name: 'role_type',
                  label: '角色类型',
                  required: true,
                  options: [
                    { label: '商户管理员', value: 'merchant_admin' },
                    { label: '商户操作员', value: 'merchant_operator' }
                  ]
                },
                {
                  type: 'checkboxes',
                  name: 'permissions',
                  label: '权限设置',
                  options: [
                    { label: '商品查看', value: 'merchant:product:view' },
                    { label: '商品管理', value: 'merchant:product:create' },
                    { label: '订单查看', value: 'merchant:order:view' },
                    { label: '订单处理', value: 'merchant:order:process' },
                    { label: '用户查看', value: 'merchant:user:view' },
                    { label: '用户管理', value: 'merchant:user:manage' },
                    { label: '报表查看', value: 'merchant:report:view' },
                    { label: '报表导出', value: 'merchant:report:export' }
                  ]
                },
                {
                  type: 'switch',
                  name: 'send_welcome_email',
                  label: '发送欢迎邮件',
                  value: true
                }
              ]
            },
            actions: [
              {
                type: 'button',
                actionType: 'cancel',
                label: '取消'
              },
              {
                type: 'button',
                actionType: 'submit',
                label: '创建',
                level: 'primary',
                onClick: (_e: any, props: any) => {
                  setRefreshKey(prev => prev + 1);
                  props.onClose();
                }
              }
            ]
          }
        }
      ],
      filter: {
        title: '条件搜索',
        body: [
          {
            type: 'input-text',
            name: 'username',
            label: '用户名',
            placeholder: '请输入用户名',
            clearable: true
          },
          {
            type: 'input-text',
            name: 'email',
            label: '邮箱',
            placeholder: '请输入邮箱',
            clearable: true
          },
          {
            type: 'select',
            name: 'role_type',
            label: '角色类型',
            placeholder: '请选择角色',
            clearable: true,
            options: [
              { label: '商户管理员', value: 'merchant_admin' },
              { label: '商户操作员', value: 'merchant_operator' }
            ]
          },
          {
            type: 'select',
            name: 'status',
            label: '状态',
            placeholder: '请选择状态',
            clearable: true,
            options: [
              { label: '待激活', value: 'pending' },
              { label: '已激活', value: 'active' },
              { label: '已暂停', value: 'suspended' },
              { label: '已停用', value: 'deactivated' }
            ]
          },
          {
            type: 'input-text',
            name: 'search',
            label: '全文搜索',
            placeholder: '搜索用户名、邮箱或手机号',
            clearable: true
          }
        ]
      },
      perPageAvailable: [10, 20, 50, 100],
      defaultParams: {
        page: 1,
        page_size: 20
      }
    }
  };

  return (
    <PermissionGuard 
      anyPermissions={['merchant:user:view']}
      fallback={<div className="text-center p-4">您没有权限访问商户用户管理</div>}
    >
      <div className="merchant-user-list-page">
        <AmisRenderer schema={pageSchema} />
      </div>
    </PermissionGuard>
  );
};
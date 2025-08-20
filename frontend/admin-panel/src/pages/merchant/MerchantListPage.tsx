import React, { useState } from 'react';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { PermissionGuard } from '../../components/ui/PermissionGuard';
import { MerchantService } from '../../services/merchantService';
import type { MerchantStatus } from '../../types/merchant';

/**
 * 商户列表页面
 * 支持搜索、筛选、分页和状态管理
 */
export const MerchantListPage: React.FC = () => {
  const [refreshKey, setRefreshKey] = useState(0);

  // 处理审批操作
  const handleApprove = async (id: number, comment?: string) => {
    try {
      await MerchantService.approveMerchant(id, comment);
      setRefreshKey(prev => prev + 1); // 刷新列表
    } catch (error) {
      console.error('审批商户失败:', error);
      throw error;
    }
  };

  // 处理拒绝操作
  const handleReject = async (id: number, comment?: string) => {
    try {
      await MerchantService.rejectMerchant(id, comment);
      setRefreshKey(prev => prev + 1); // 刷新列表
    } catch (error) {
      console.error('拒绝商户失败:', error);
      throw error;
    }
  };

  // 处理状态更新
  const handleStatusUpdate = async (id: number, status: MerchantStatus, comment?: string) => {
    try {
      await MerchantService.updateMerchantStatus(id, { status, comment });
      setRefreshKey(prev => prev + 1); // 刷新列表
    } catch (error) {
      console.error('更新商户状态失败:', error);
      throw error;
    }
  };

  // Amis页面配置
  const pageSchema = {
    type: 'page',
    title: '商户管理',
    body: {
      type: 'crud',
      api: {
        method: 'get',
        url: '/api/v1/merchants',
        data: {
          '&': '$$'
        }
      },
      key: refreshKey, // 用于强制刷新
      syncLocation: false,
      draggable: false,
      columns: [
        {
          name: 'name',
          label: '商户名称',
          searchable: true,
          type: 'text',
          width: 150
        },
        {
          name: 'code',
          label: '商户代码',
          searchable: true,
          type: 'text',
          width: 120
        },
        {
          name: 'status',
          label: '状态',
          type: 'mapping',
          width: 100,
          map: {
            'pending': '<span class="badge badge-warning">待审核</span>',
            'active': '<span class="badge badge-success">已激活</span>',
            'suspended': '<span class="badge badge-secondary">已暂停</span>',
            'deactivated': '<span class="badge badge-danger">已停用</span>'
          }
        },
        {
          name: 'business_info.contact_name',
          label: '联系人',
          type: 'text',
          width: 100
        },
        {
          name: 'business_info.contact_phone',
          label: '联系电话',
          type: 'text',
          width: 120
        },
        {
          name: 'registration_time',
          label: '申请时间',
          type: 'datetime',
          format: 'YYYY-MM-DD HH:mm',
          width: 150
        },
        {
          type: 'operation',
          label: '操作',
          width: 200,
          buttons: [
            {
              type: 'button',
              label: '查看',
              level: 'link',
              actionType: 'dialog',
              dialog: {
                title: '商户详情',
                size: 'lg',
                body: {
                  type: 'service',
                  api: '/api/v1/merchants/${id}',
                  body: [
                    {
                      type: 'form',
                      mode: 'horizontal',
                      disabled: true,
                      body: [
                        {
                          type: 'input-text',
                          name: 'name',
                          label: '商户名称'
                        },
                        {
                          type: 'input-text',
                          name: 'code',
                          label: '商户代码'
                        },
                        {
                          type: 'input-text',
                          name: 'business_info.contact_name',
                          label: '联系人'
                        },
                        {
                          type: 'input-text',
                          name: 'business_info.contact_phone',
                          label: '联系电话'
                        },
                        {
                          type: 'input-text',
                          name: 'business_info.contact_email',
                          label: '联系邮箱'
                        },
                        {
                          type: 'textarea',
                          name: 'business_info.address',
                          label: '经营地址'
                        },
                        {
                          type: 'textarea',
                          name: 'business_info.scope',
                          label: '经营范围'
                        }
                      ]
                    }
                  ]
                }
              }
            },
            {
              type: 'button',
              label: '审批',
              level: 'primary',
              visibleOn: "${status === 'pending'}",
              actionType: 'dialog',
              dialog: {
                title: '商户审批',
                body: {
                  type: 'form',
                  body: [
                    {
                      type: 'static',
                      name: 'name',
                      label: '商户名称'
                    },
                    {
                      type: 'radios',
                      name: 'action',
                      label: '审批决定',
                      required: true,
                      options: [
                        { label: '批准', value: 'approve' },
                        { label: '拒绝', value: 'reject' }
                      ]
                    },
                    {
                      type: 'textarea',
                      name: 'comment',
                      label: '审批意见',
                      placeholder: '请输入审批意见'
                    }
                  ]
                },
                confirm: {
                  title: '确认审批',
                  text: '确定要执行此审批操作吗？'
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
                      const { action, comment } = props.scope;
                      const id = props.scope.id;
                      
                      try {
                        if (action === 'approve') {
                          await handleApprove(id, comment);
                        } else {
                          await handleReject(id, comment);
                        }
                        props.onClose();
                      } catch (error) {
                        // 错误处理已在handleApprove/handleReject中完成
                      }
                    }
                  }
                ]
              }
            },
            {
              type: 'button',
              label: '状态管理',
              level: 'info',
              visibleOn: "${status === 'active' || status === 'suspended'}",
              actionType: 'dialog',
              dialog: {
                title: '状态管理',
                body: {
                  type: 'form',
                  body: [
                    {
                      type: 'static',
                      name: 'name',
                      label: '商户名称'
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
            }
          ]
        }
      ],
      headerToolbar: [
        'filter-toggler',
        'bulkActions',
        {
          type: 'button',
          label: '新增商户',
          level: 'primary',
          actionType: 'link',
          link: '/merchant/register'
        }
      ],
      filter: {
        title: '条件搜索',
        body: [
          {
            type: 'input-text',
            name: 'name',
            label: '商户名称',
            placeholder: '请输入商户名称',
            clearable: true
          },
          {
            type: 'select',
            name: 'status',
            label: '状态',
            placeholder: '请选择状态',
            clearable: true,
            options: [
              { label: '待审核', value: 'pending' },
              { label: '已激活', value: 'active' },
              { label: '已暂停', value: 'suspended' },
              { label: '已停用', value: 'deactivated' }
            ]
          },
          {
            type: 'input-text',
            name: 'search',
            label: '全文搜索',
            placeholder: '搜索商户名称、代码或联系人',
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
      anyPermissions={['merchant:view']}
      fallback={<div className="text-center p-4">您没有权限访问商户管理</div>}
    >
      <div className="merchant-list-page">
        <AmisRenderer schema={pageSchema} />
      </div>
    </PermissionGuard>
  );
};
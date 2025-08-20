import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { PermissionGuard } from '../../components/ui/PermissionGuard';
import { useMerchantPermissions } from '../../hooks/useMerchantPermissions';
import { MerchantUserService } from '../../services/merchantUserService';
import { MERCHANT_PERMISSIONS } from '../../types/merchantUser';

/**
 * 商户用户批量创建页面
 * 支持批量导入和创建多个用户
 */
export const MerchantUserBatchCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const { getCurrentMerchantId } = useMerchantPermissions();
  const [uploadResult, setUploadResult] = useState<any>(null);

  // 处理批量创建
  const handleBatchCreate = async (users: any[]) => {
    try {
      const usersWithMerchantId = users.map(user => ({
        ...user,
        merchant_id: getCurrentMerchantId()
      }));
      
      const result = await MerchantUserService.createMerchantUsersBatch(usersWithMerchantId);
      setUploadResult({
        success: true,
        total: users.length,
        created: result.length,
        message: `成功创建 ${result.length} 个用户`
      });
      return result;
    } catch (error) {
      setUploadResult({
        success: false,
        message: error instanceof Error ? error.message : '批量创建失败'
      });
      throw error;
    }
  };

  // 表单配置
  const pageSchema = {
    type: 'page',
    title: '批量创建商户用户',
    body: [
      {
        type: 'alert',
        level: 'info',
        body: '您可以通过Excel文件批量导入用户，或手动添加多个用户。建议先下载模板文件。'
      },
      {
        type: 'tabs',
        tabs: [
          {
            title: 'Excel导入',
            body: [
              {
                type: 'form',
                title: '文件上传',
                body: [
                  {
                    type: 'button',
                    label: '下载Excel模板',
                    level: 'link',
                    icon: 'fa fa-download',
                    onClick: () => {
                      // 创建模板下载逻辑
                      const template = [
                        ['用户名*', '邮箱*', '手机号', '角色类型*', '权限'],
                        ['user001', 'user001@example.com', '13800138001', 'merchant_operator', 'merchant:product:view,merchant:order:view'],
                        ['user002', 'user002@example.com', '13800138002', 'merchant_admin', 'merchant:product:view,merchant:product:create,merchant:order:view,merchant:order:process']
                      ];
                      
                      // 简单的CSV下载实现
                      const csvContent = template.map(row => row.join(',')).join('\n');
                      const blob = new Blob([csvContent], { type: 'text/csv' });
                      const url = window.URL.createObjectURL(blob);
                      const a = document.createElement('a');
                      a.href = url;
                      a.download = '商户用户导入模板.csv';
                      a.click();
                      window.URL.revokeObjectURL(url);
                    }
                  },
                  {
                    type: 'divider'
                  },
                  {
                    type: 'input-file',
                    name: 'userFile',
                    label: '选择Excel文件',
                    accept: '.xlsx,.xls,.csv',
                    autoUpload: false,
                    description: '支持Excel (.xlsx, .xls) 和CSV (.csv) 格式'
                  },
                  {
                    type: 'alert',
                    level: 'warning',
                    body: '文件格式要求：\n- 第一行为标题行\n- 用户名和邮箱必填\n- 角色类型：merchant_admin 或 merchant_operator\n- 权限用逗号分隔',
                    visibleOn: 'this.userFile'
                  }
                ],
                actions: [
                  {
                    type: 'button',
                    label: '解析文件',
                    level: 'primary',
                    actionType: 'submit',
                    onClick: async (_e: any, props: any) => {
                      // 这里应该实现文件解析逻辑
                      console.log('解析文件:', props.scope.userFile);
                      // 模拟解析结果
                      props.setData('parsedUsers', [
                        {
                          username: 'user001',
                          email: 'user001@example.com',
                          phone: '13800138001',
                          role_type: 'merchant_operator',
                          permissions: ['merchant:product:view', 'merchant:order:view']
                        }
                      ]);
                    }
                  }
                ]
              },
              {
                type: 'form',
                title: '确认用户信息',
                visibleOn: 'this.parsedUsers && this.parsedUsers.length > 0',
                body: [
                  {
                    type: 'table',
                    name: 'parsedUsers',
                    columns: [
                      { name: 'username', label: '用户名' },
                      { name: 'email', label: '邮箱' },
                      { name: 'phone', label: '手机号' },
                      { name: 'role_type', label: '角色类型' },
                      { name: 'permissions', label: '权限', type: 'json' }
                    ]
                  },
                  {
                    type: 'switch',
                    name: 'send_welcome_emails',
                    label: '发送欢迎邮件',
                    value: true
                  }
                ],
                actions: [
                  {
                    type: 'button',
                    label: '批量创建',
                    level: 'primary',
                    onClick: async (_e: any, props: any) => {
                      const users = props.scope.parsedUsers;
                      try {
                        await handleBatchCreate(users);
                      } catch (error) {
                        console.error('批量创建失败:', error);
                      }
                    }
                  }
                ]
              }
            ]
          },
          {
            title: '手动添加',
            body: [
              {
                type: 'form',
                title: '批量用户信息',
                body: [
                  {
                    type: 'input-array',
                    name: 'users',
                    label: '用户列表',
                    items: {
                      type: 'object',
                      properties: {
                        username: {
                          type: 'input-text',
                          label: '用户名',
                          required: true,
                          placeholder: '请输入用户名'
                        },
                        email: {
                          type: 'input-email',
                          label: '邮箱',
                          required: true,
                          placeholder: '请输入邮箱'
                        },
                        phone: {
                          type: 'input-text',
                          label: '手机号',
                          placeholder: '请输入手机号'
                        },
                        role_type: {
                          type: 'select',
                          label: '角色类型',
                          required: true,
                          options: [
                            { label: '商户管理员', value: 'merchant_admin' },
                            { label: '商户操作员', value: 'merchant_operator' }
                          ]
                        },
                        permissions: {
                          type: 'checkboxes',
                          label: '权限',
                          options: [
                            { label: '商品查看', value: MERCHANT_PERMISSIONS.PRODUCT_VIEW },
                            { label: '商品管理', value: MERCHANT_PERMISSIONS.PRODUCT_CREATE },
                            { label: '订单查看', value: MERCHANT_PERMISSIONS.ORDER_VIEW },
                            { label: '订单处理', value: MERCHANT_PERMISSIONS.ORDER_PROCESS },
                            { label: '用户查看', value: MERCHANT_PERMISSIONS.USER_VIEW },
                            { label: '用户管理', value: MERCHANT_PERMISSIONS.USER_MANAGE },
                            { label: '报表查看', value: MERCHANT_PERMISSIONS.REPORT_VIEW },
                            { label: '报表导出', value: MERCHANT_PERMISSIONS.REPORT_EXPORT }
                          ]
                        }
                      }
                    },
                    minItems: 1,
                    maxItems: 50,
                    addButtonText: '添加用户',
                    value: [
                      {
                        username: '',
                        email: '',
                        phone: '',
                        role_type: 'merchant_operator',
                        permissions: [MERCHANT_PERMISSIONS.PRODUCT_VIEW, MERCHANT_PERMISSIONS.ORDER_VIEW]
                      }
                    ]
                  },
                  {
                    type: 'switch',
                    name: 'send_welcome_emails',
                    label: '发送欢迎邮件',
                    value: true
                  }
                ],
                actions: [
                  {
                    type: 'button',
                    actionType: 'cancel',
                    label: '取消',
                    onClick: () => {
                      navigate('/merchant-user');
                    }
                  },
                  {
                    type: 'button',
                    label: '批量创建',
                    level: 'primary',
                    onClick: async (_e: any, props: any) => {
                      const users = props.scope.users;
                      try {
                        await handleBatchCreate(users);
                      } catch (error) {
                        console.error('批量创建失败:', error);
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
        type: 'alert',
        level: uploadResult?.success ? 'success' : 'danger',
        body: uploadResult?.message,
        visibleOn: '!!uploadResult',
        showCloseButton: true
      },
      {
        type: 'button',
        label: '返回用户列表',
        level: 'link',
        visibleOn: 'uploadResult && uploadResult.success',
        onClick: () => {
          navigate('/merchant-user');
        }
      }
    ]
  };

  return (
    <PermissionGuard 
      anyPermissions={['merchant:user:manage']}
      fallback={<div className="text-center p-4">您没有权限管理商户用户</div>}
    >
      <div className="merchant-user-batch-create-page">
        <AmisRenderer schema={pageSchema} />
      </div>
    </PermissionGuard>
  );
};
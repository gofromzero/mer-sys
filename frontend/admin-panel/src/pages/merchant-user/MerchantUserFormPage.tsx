import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { PermissionGuard } from '../../components/ui/PermissionGuard';
import { useMerchantPermissions } from '../../hooks/useMerchantPermissions';
import { MERCHANT_PERMISSIONS } from '../../types/merchantUser';

/**
 * 商户用户创建/编辑表单页面
 * 支持新建和编辑模式
 */
export const MerchantUserFormPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id?: string }>();
  const { canManageUsers, getCurrentMerchantId } = useMerchantPermissions();
  
  const isEditMode = !!id;
  const pageTitle = isEditMode ? '编辑商户用户' : '新增商户用户';

  // 表单配置
  const formSchema = {
    type: 'page',
    title: pageTitle,
    body: {
      type: 'form',
      api: {
        method: isEditMode ? 'put' : 'post',
        url: isEditMode ? `/api/v1/merchant-users/${id}` : '/api/v1/merchant-users',
        data: {
          '&': '$$',
          merchant_id: getCurrentMerchantId() // 自动注入当前商户ID
        }
      },
      initApi: isEditMode ? `/api/v1/merchant-users/${id}` : undefined,
      mode: 'horizontal',
      horizontal: {
        left: 3,
        right: 9
      },
      body: [
        {
          type: 'divider',
          title: '基本信息'
        },
        {
          type: 'input-text',
          name: 'username',
          label: '用户名',
          required: true,
          placeholder: '请输入用户名',
          description: '用户名必须唯一，建议使用字母和数字组合',
          validations: {
            minLength: 3,
            maxLength: 20,
            matchRegexp: '^[a-zA-Z0-9_]+$'
          },
          validationErrors: {
            minLength: '用户名至少3个字符',
            maxLength: '用户名最多20个字符',
            matchRegexp: '用户名只能包含字母、数字和下划线'
          }
        },
        {
          type: 'input-email',
          name: 'email',
          label: '邮箱地址',
          required: true,
          placeholder: '请输入邮箱地址',
          description: '邮箱将用于接收系统通知和密码重置'
        },
        {
          type: 'input-text',
          name: 'phone',
          label: '手机号码',
          placeholder: '请输入手机号码',
          validations: {
            matchRegexp: '^1[3-9]\\d{9}$'
          },
          validationErrors: {
            matchRegexp: '请输入正确的手机号码'
          }
        },
        {
          type: 'divider',
          title: '角色权限'
        },
        {
          type: 'select',
          name: 'role_type',
          label: '角色类型',
          required: true,
          options: [
            { 
              label: '商户管理员', 
              value: 'merchant_admin',
              description: '拥有商户内所有权限，可以管理其他用户'
            },
            { 
              label: '商户操作员', 
              value: 'merchant_operator',
              description: '拥有基本操作权限，不能管理其他用户'
            }
          ],
          description: '角色决定用户的基本权限范围'
        },
        {
          type: 'checkboxes',
          name: 'permissions',
          label: '详细权限',
          required: true,
          options: [
            {
              label: '商品管理',
              children: [
                { label: '商品查看', value: MERCHANT_PERMISSIONS.PRODUCT_VIEW },
                { label: '商品创建', value: MERCHANT_PERMISSIONS.PRODUCT_CREATE },
                { label: '商品编辑', value: MERCHANT_PERMISSIONS.PRODUCT_EDIT },
                { label: '商品删除', value: MERCHANT_PERMISSIONS.PRODUCT_DELETE }
              ]
            },
            {
              label: '订单管理',
              children: [
                { label: '订单查看', value: MERCHANT_PERMISSIONS.ORDER_VIEW },
                { label: '订单处理', value: MERCHANT_PERMISSIONS.ORDER_PROCESS },
                { label: '订单取消', value: MERCHANT_PERMISSIONS.ORDER_CANCEL }
              ]
            },
            {
              label: '用户管理',
              children: [
                { label: '用户查看', value: MERCHANT_PERMISSIONS.USER_VIEW },
                { label: '用户管理', value: MERCHANT_PERMISSIONS.USER_MANAGE }
              ]
            },
            {
              label: '报表分析',
              children: [
                { label: '报表查看', value: MERCHANT_PERMISSIONS.REPORT_VIEW },
                { label: '报表导出', value: MERCHANT_PERMISSIONS.REPORT_EXPORT }
              ]
            }
          ],
          description: '选择用户具体的功能权限',
          checkAll: true,
          defaultCheckAll: false,
          joinValues: false,
          extractValue: true
        },
        {
          type: 'divider',
          title: '账号设置',
          visibleOn: '!this.id' // 只在新建时显示
        },
        {
          type: 'switch',
          name: 'send_welcome_email',
          label: '发送欢迎邮件',
          value: true,
          description: '系统将向用户发送包含初始密码的欢迎邮件',
          visibleOn: '!this.id' // 只在新建时显示
        },
        {
          type: 'input-password',
          name: 'initial_password',
          label: '初始密码',
          placeholder: '留空将自动生成随机密码',
          description: '如果留空，系统将生成8位随机密码',
          visibleOn: '!this.id && !this.send_welcome_email', // 新建且不发邮件时显示
          validations: {
            minLength: 6,
            maxLength: 20
          },
          validationErrors: {
            minLength: '密码至少6个字符',
            maxLength: '密码最多20个字符'
          }
        },
        {
          type: 'alert',
          level: 'info',
          body: '提示：新建用户后，系统将自动为其分配指定的权限。用户首次登录时需要修改密码。',
          visibleOn: '!this.id'
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
          type: 'submit',
          label: isEditMode ? '保存修改' : '创建用户',
          level: 'primary'
        }
      ],
      redirect: '/merchant-user?_t=${timestamp}', // 提交成功后跳转
      onSubmit: () => {
        // 提交成功后的处理
        const message = isEditMode ? '用户信息更新成功' : '用户创建成功';
        // 这里可以添加成功提示逻辑
        console.log(message);
      }
    }
  };

  return (
    <PermissionGuard 
      anyPermissions={['merchant:user:manage']}
      fallback={<div className="text-center p-4">您没有权限管理商户用户</div>}
    >
      <div className="merchant-user-form-page">
        <AmisRenderer schema={formSchema} />
      </div>
    </PermissionGuard>
  );
};
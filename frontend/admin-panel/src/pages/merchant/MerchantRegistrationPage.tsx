import React from 'react';
import { AmisRenderer } from '../../components/ui/AmisRenderer';
import { PermissionGuard } from '../../components/ui/PermissionGuard';

/**
 * 商户注册页面
 * 支持完整的商户信息录入和验证
 */
export const MerchantRegistrationPage: React.FC = () => {
  // 处理表单提交（如果需要自定义处理逻辑）
  // const handleSubmit = async (data: MerchantRegistrationRequest) => {
  //   try {
  //     const result = await MerchantService.registerMerchant(data);
  //     console.log('商户注册成功:', result);
  //     // 注册成功后跳转到商户列表
  //     window.location.href = '/merchant/list';
  //   } catch (error) {
  //     console.error('商户注册失败:', error);
  //     throw error;
  //   }
  // };

  // Amis页面配置
  const pageSchema = {
    type: 'page',
    title: '商户注册',
    body: {
      type: 'form',
      api: {
        method: 'post',
        url: '/api/v1/merchants',
        adaptor: (payload: any) => {
          // 可以在这里对提交数据进行预处理
          return payload;
        }
      },
      redirect: '/merchant/list',
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
          name: 'name',
          label: '商户名称',
          required: true,
          placeholder: '请输入商户名称',
          description: '商户的对外显示名称',
          validations: {
            minLength: 2,
            maxLength: 100
          },
          validationErrors: {
            minLength: '商户名称至少2个字符',
            maxLength: '商户名称不能超过100个字符'
          }
        },
        {
          type: 'input-text',
          name: 'code',
          label: '商户代码',
          required: true,
          placeholder: '请输入商户代码',
          description: '商户的唯一标识符，只能包含字母和数字',
          validations: {
            isAlphanumeric: true,
            minLength: 3,
            maxLength: 50
          },
          validationErrors: {
            isAlphanumeric: '商户代码只能包含字母和数字',
            minLength: '商户代码至少3个字符',
            maxLength: '商户代码不能超过50个字符'
          }
        },
        {
          type: 'divider',
          title: '业务信息'
        },
        {
          type: 'select',
          name: 'business_info.type',
          label: '商户类型',
          required: true,
          placeholder: '请选择商户类型',
          options: [
            { label: '零售商户', value: 'retail' },
            { label: '批发商户', value: 'wholesale' },
            { label: '服务商户', value: 'service' }
          ]
        },
        {
          type: 'input-text',
          name: 'business_info.category',
          label: '业务分类',
          required: true,
          placeholder: '请输入业务分类',
          description: '如：餐饮、零售、服务等'
        },
        {
          type: 'input-text',
          name: 'business_info.license',
          label: '营业执照号',
          required: true,
          placeholder: '请输入营业执照号',
          validations: {
            minLength: 10,
            maxLength: 30
          },
          validationErrors: {
            minLength: '营业执照号至少10位',
            maxLength: '营业执照号不能超过30位'
          }
        },
        {
          type: 'input-text',
          name: 'business_info.legal_name',
          label: '法人姓名',
          required: true,
          placeholder: '请输入法人姓名'
        },
        {
          type: 'divider',
          title: '联系信息'
        },
        {
          type: 'input-text',
          name: 'business_info.contact_name',
          label: '联系人',
          required: true,
          placeholder: '请输入联系人姓名'
        },
        {
          type: 'input-text',
          name: 'business_info.contact_phone',
          label: '联系电话',
          required: true,
          placeholder: '请输入联系电话',
          validations: {
            isNumeric: true,
            minLength: 10,
            maxLength: 20
          },
          validationErrors: {
            isNumeric: '请输入正确的电话号码',
            minLength: '电话号码至少10位',
            maxLength: '电话号码不能超过20位'
          }
        },
        {
          type: 'input-email',
          name: 'business_info.contact_email',
          label: '联系邮箱',
          required: true,
          placeholder: '请输入联系邮箱',
          description: '用于接收重要通知'
        },
        {
          type: 'textarea',
          name: 'business_info.address',
          label: '经营地址',
          required: true,
          placeholder: '请输入详细的经营地址',
          minRows: 2,
          maxRows: 4
        },
        {
          type: 'textarea',
          name: 'business_info.scope',
          label: '经营范围',
          required: true,
          placeholder: '请描述主要经营范围',
          description: '详细描述商户的主要业务范围',
          minRows: 3,
          maxRows: 5
        },
        {
          type: 'textarea',
          name: 'business_info.description',
          label: '商户描述',
          placeholder: '请输入商户简介（可选）',
          description: '对商户的简要介绍',
          minRows: 2,
          maxRows: 4
        }
      ],
      actions: [
        {
          type: 'button',
          actionType: 'reset',
          label: '重置'
        },
        {
          type: 'button',
          actionType: 'cancel',
          label: '取消',
          onClick: () => {
            window.location.href = '/merchant/list';
          }
        },
        {
          type: 'submit',
          actionType: 'submit',
          label: '提交申请',
          level: 'primary'
        }
      ],
      submitText: '提交申请',
      resetAfterSubmit: true,
      messages: {
        saveSuccess: '商户注册申请已提交，请等待审核',
        saveFailed: '注册申请提交失败，请检查信息后重试'
      }
    }
  };

  return (
    <PermissionGuard 
      anyPermissions={['merchant:create']}
      fallback={<div className="text-center p-4">您没有权限注册商户</div>}
    >
      <div className="merchant-registration-page">
        <AmisRenderer schema={pageSchema} />
      </div>
    </PermissionGuard>
  );
};
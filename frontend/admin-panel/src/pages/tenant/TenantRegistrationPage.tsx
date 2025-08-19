import AmisRenderer from '../../components/ui/AmisRenderer'
import { useTenantPermissions } from '../../hooks/useTenantPermissions'

const TenantRegistrationPage = () => {
  const { canCreateTenant } = useTenantPermissions();

  // 如果没有创建权限，显示无权限提示
  if (!canCreateTenant) {
    return (
      <div className="flex items-center justify-center min-h-64">
        <div className="text-center">
          <div className="text-gray-400 text-6xl mb-4">🔒</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">无创建权限</h3>
          <p className="text-gray-500">您没有创建新租户的权限，请联系管理员。</p>
        </div>
      </div>
    );
  }
  const schema = {
    type: 'page',
    title: '租户注册',
    body: [
      {
        type: 'panel',
        title: '租户注册申请',
        body: [
          {
            type: 'form',
            api: 'post:/api/v1/tenants',
            redirect: '/tenant/list',
            mode: 'horizontal',
            horizontal: {
              left: 'col-sm-2',
              right: 'col-sm-10'
            },
            body: [
              {
                type: 'fieldset',
                title: '基本信息',
                body: [
                  {
                    type: 'grid',
                    columns: [
                      {
                        md: 6,
                        body: [
                          {
                            type: 'input-text',
                            name: 'name',
                            label: '租户名称',
                            required: true,
                            validations: {
                              minLength: 2,
                              maxLength: 100
                            },
                            description: '请输入租户的完整名称',
                            placeholder: '例如：某某科技有限公司'
                          },
                          {
                            type: 'input-text',
                            name: 'code',
                            label: '租户代码',
                            required: true,
                            validations: {
                              minLength: 2,
                              maxLength: 50,
                              pattern: '^[a-zA-Z0-9_-]+$'
                            },
                            description: '只能包含字母、数字、下划线和连字符，用于系统识别',
                            placeholder: '例如：company-tech'
                          }
                        ]
                      },
                      {
                        md: 6,
                        body: [
                          {
                            type: 'select',
                            name: 'business_type',
                            label: '业务类型',
                            required: true,
                            options: [
                              { label: '电子商务', value: 'ecommerce' },
                              { label: '零售连锁', value: 'retail' },
                              { label: '餐饮服务', value: 'food' },
                              { label: '教育培训', value: 'education' },
                              { label: '医疗健康', value: 'healthcare' },
                              { label: '金融服务', value: 'finance' },
                              { label: '物流运输', value: 'logistics' },
                              { label: '制造业', value: 'manufacturing' },
                              { label: '服务业', value: 'service' },
                              { label: '其他', value: 'other' }
                            ],
                            description: '请选择与您业务最匹配的类型'
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                type: 'fieldset',
                title: '联系信息',
                body: [
                  {
                    type: 'grid',
                    columns: [
                      {
                        md: 6,
                        body: [
                          {
                            type: 'input-text',
                            name: 'contact_person',
                            label: '联系人',
                            required: true,
                            description: '主要联系人姓名',
                            placeholder: '请输入联系人姓名'
                          },
                          {
                            type: 'input-email',
                            name: 'contact_email',
                            label: '联系邮箱',
                            required: true,
                            description: '用于接收重要通知和系统消息',
                            placeholder: 'example@company.com'
                          }
                        ]
                      },
                      {
                        md: 6,
                        body: [
                          {
                            type: 'input-text',
                            name: 'contact_phone',
                            label: '联系电话',
                            validations: {
                              isPhoneNumber: true
                            },
                            description: '建议填写手机号码',
                            placeholder: '例如：13800138000'
                          }
                        ]
                      }
                    ]
                  },
                  {
                    type: 'textarea',
                    name: 'address',
                    label: '详细地址',
                    rows: 3,
                    description: '请填写完整的办公地址',
                    placeholder: '请输入详细地址，包括省市区街道门牌号'
                  }
                ]
              },
              {
                type: 'fieldset',
                title: '服务协议',
                body: [
                  {
                    type: 'checkboxes',
                    name: 'agreements',
                    required: true,
                    options: [
                      {
                        label: '我已阅读并同意《用户服务协议》',
                        value: 'service_agreement'
                      },
                      {
                        label: '我已阅读并同意《隐私政策》',
                        value: 'privacy_policy'
                      },
                      {
                        label: '同意接收产品更新和服务通知',
                        value: 'notifications'
                      }
                    ],
                    validations: {
                      minLength: 2
                    },
                    validationErrors: {
                      minLength: '请至少同意服务协议和隐私政策'
                    }
                  }
                ]
              },
              {
                type: 'divider'
              },
              {
                type: 'group',
                body: [
                  {
                    type: 'submit',
                    label: '提交注册申请',
                    level: 'primary',
                    size: 'lg'
                  },
                  {
                    type: 'reset',
                    label: '重置表单',
                    level: 'default',
                    size: 'lg'
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        type: 'panel',
        title: '注册须知',
        className: 'mt-4',
        body: [
          {
            type: 'alert',
            level: 'info',
            body: [
              {
                type: 'tpl',
                tpl: `
                  <h5>租户注册流程说明：</h5>
                  <ul>
                    <li>1. 填写完整的注册信息并提交申请</li>
                    <li>2. 系统将自动创建租户账户并分配初始配置</li>
                    <li>3. 租户状态默认为"激活"，可立即使用系统功能</li>
                    <li>4. 如需修改配置或状态，请联系系统管理员</li>
                  </ul>
                  <h5>默认配置：</h5>
                  <ul>
                    <li>最大用户数：100</li>
                    <li>最大商户数：50</li>
                    <li>功能特性：基础功能</li>
                  </ul>
                `
              }
            ]
          }
        ]
      }
    ]
  }

  return <AmisRenderer schema={schema} />
}

export default TenantRegistrationPage
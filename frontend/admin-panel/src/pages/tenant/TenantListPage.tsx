import AmisRenderer from '../../components/ui/AmisRenderer'
import { useTenantPermissions } from '../../hooks/useTenantPermissions'

const TenantListPage = () => {
  const {
    canViewTenants,
    canCreateTenant,
    canEditTenant,
    canManageTenantStatus,
    canManageTenantConfig
  } = useTenantPermissions();

  // 确认对话框状态 (暂时注释，未来用于敏感操作确认)
  // const [confirmDialog, setConfirmDialog] = useState<{
  //   isOpen: boolean;
  //   title: string;
  //   message: string;
  //   type: 'warning' | 'danger' | 'info';
  //   onConfirm: () => void;
  // }>({
  //   isOpen: false,
  //   title: '',
  //   message: '',
  //   type: 'warning',
  //   onConfirm: () => {}
  // });

  // 显示确认对话框
  // const showConfirmDialog = (config: Omit<typeof confirmDialog, 'isOpen'>) => {
  //   setConfirmDialog({ ...config, isOpen: true });
  // };

  // 隐藏确认对话框
  // const hideConfirmDialog = () => {
  //   setConfirmDialog(prev => ({ ...prev, isOpen: false }));
  // };

  // 处理敏感操作的确认 (保留以备后续使用)
  // const handleSensitiveOperation = (operation: string, callback: () => void) => {
  //   if (requiresConfirmation(operation as any)) {
  //     const operationLabels = {
  //       'delete': { title: '删除租户', message: '此操作不可逆，确定要删除该租户吗？', type: 'danger' as const },
  //       'manage_status': { title: '变更租户状态', message: '变更租户状态可能影响其正常使用，确定要继续吗？', type: 'warning' as const },
  //       'manage_config': { title: '修改租户配置', message: '修改配置可能影响租户的功能和限制，确定要继续吗？', type: 'warning' as const }
  //     };

  //     const config = operationLabels[operation as keyof typeof operationLabels];
  //     if (config) {
  //       showConfirmDialog({
  //         ...config,
  //         onConfirm: () => {
  //           callback();
  //           hideConfirmDialog();
  //         }
  //       });
  //     } else {
  //       callback();
  //     }
  //   } else {
  //     callback();
  //   }
  // };

  // 如果没有查看权限，显示无权限提示
  if (!canViewTenants) {
    return (
      <div className="flex items-center justify-center min-h-64">
        <div className="text-center">
          <div className="text-gray-400 text-6xl mb-4">🔒</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">无访问权限</h3>
          <p className="text-gray-500">您没有查看租户信息的权限，请联系管理员。</p>
        </div>
      </div>
    );
  }
  const schema = {
    type: 'page',
    title: '租户管理',
    body: [
      {
        type: 'crud',
        api: '/api/v1/tenants',
        searchable: true,
        filter: {
          body: [
            {
              type: 'input-text',
              name: 'search',
              label: '搜索',
              placeholder: '搜索租户名称、代码、联系人或邮箱',
              clearable: true
            },
            {
              type: 'select',
              name: 'status',
              label: '状态',
              placeholder: '全部状态',
              clearable: true,
              options: [
                { label: '激活', value: 'active' },
                { label: '暂停', value: 'suspended' },
                { label: '过期', value: 'expired' }
              ]
            },
            {
              type: 'input-text',
              name: 'business_type',
              label: '业务类型',
              placeholder: '业务类型',
              clearable: true
            }
          ]
        },
        columns: [
          {
            name: 'id',
            label: 'ID',
            type: 'text',
            width: 80
          },
          {
            name: 'name',
            label: '租户名称',
            type: 'text',
            searchable: true
          },
          {
            name: 'code',
            label: '租户代码',
            type: 'text',
            searchable: true
          },
          {
            name: 'business_type',
            label: '业务类型',
            type: 'text'
          },
          {
            name: 'contact_person',
            label: '联系人',
            type: 'text'
          },
          {
            name: 'contact_email',
            label: '联系邮箱',
            type: 'text'
          },
          {
            name: 'status',
            label: '状态',
            type: 'status',
            map: {
              'active': {
                value: 'active',
                label: '激活',
                level: 'success'
              },
              'suspended': {
                value: 'suspended',
                label: '暂停',
                level: 'warning'
              },
              'expired': {
                value: 'expired',
                label: '过期',
                level: 'danger'
              }
            }
          },
          {
            name: 'created_at',
            label: '创建时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm:ss'
          },
          {
            type: 'operation',
            label: '操作',
            width: 200,
            buttons: [
              {
                type: 'button',
                label: '详情',
                level: 'link',
                actionType: 'dialog',
                visibleOn: canViewTenants,
                dialog: {
                  title: '租户详情',
                  size: 'lg',
                  body: {
                    type: 'service',
                    api: '/api/v1/tenants/${id}',
                    body: [
                      {
                        type: 'form',
                        mode: 'horizontal',
                        disabled: true,
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
                                    label: '租户名称'
                                  },
                                  {
                                    type: 'input-text',
                                    name: 'code',
                                    label: '租户代码'
                                  },
                                  {
                                    type: 'input-text',
                                    name: 'business_type',
                                    label: '业务类型'
                                  },
                                  {
                                    type: 'status',
                                    name: 'status',
                                    label: '状态'
                                  }
                                ]
                              },
                              {
                                md: 6,
                                body: [
                                  {
                                    type: 'input-text',
                                    name: 'contact_person',
                                    label: '联系人'
                                  },
                                  {
                                    type: 'input-text',
                                    name: 'contact_email',
                                    label: '联系邮箱'
                                  },
                                  {
                                    type: 'input-text',
                                    name: 'contact_phone',
                                    label: '联系电话'
                                  },
                                  {
                                    type: 'textarea',
                                    name: 'address',
                                    label: '地址'
                                  }
                                ]
                              }
                            ]
                          },
                          {
                            type: 'grid',
                            columns: [
                              {
                                md: 6,
                                body: [
                                  {
                                    type: 'input-datetime',
                                    name: 'registration_time',
                                    label: '注册时间',
                                    format: 'YYYY-MM-DD HH:mm:ss'
                                  },
                                  {
                                    type: 'input-datetime',
                                    name: 'created_at',
                                    label: '创建时间',
                                    format: 'YYYY-MM-DD HH:mm:ss'
                                  }
                                ]
                              },
                              {
                                md: 6,
                                body: [
                                  {
                                    type: 'input-datetime',
                                    name: 'activation_time',
                                    label: '激活时间',
                                    format: 'YYYY-MM-DD HH:mm:ss'
                                  },
                                  {
                                    type: 'input-datetime',
                                    name: 'updated_at',
                                    label: '更新时间',
                                    format: 'YYYY-MM-DD HH:mm:ss'
                                  }
                                ]
                              }
                            ]
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
                level: 'link',
                actionType: 'dialog',
                visibleOn: canEditTenant,
                dialog: {
                  title: '编辑租户',
                  body: {
                    type: 'form',
                    api: 'put:/api/v1/tenants/${id}',
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
                                }
                              },
                              {
                                type: 'input-text',
                                name: 'business_type',
                                label: '业务类型',
                                required: true
                              },
                              {
                                type: 'input-text',
                                name: 'contact_person',
                                label: '联系人',
                                required: true
                              }
                            ]
                          },
                          {
                            md: 6,
                            body: [
                              {
                                type: 'input-email',
                                name: 'contact_email',
                                label: '联系邮箱',
                                required: true
                              },
                              {
                                type: 'input-text',
                                name: 'contact_phone',
                                label: '联系电话',
                                validations: {
                                  isPhoneNumber: true
                                }
                              },
                              {
                                type: 'textarea',
                                name: 'address',
                                label: '地址'
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
                }
              },
              {
                type: 'button',
                label: '状态',
                level: 'link',
                actionType: 'dialog',
                visibleOn: canManageTenantStatus,
                dialog: {
                  title: '变更租户状态',
                  body: {
                    type: 'form',
                    api: 'put:/api/v1/tenants/${id}/status',
                    body: [
                      {
                        type: 'select',
                        name: 'status',
                        label: '新状态',
                        required: true,
                        options: [
                          { label: '激活', value: 'active' },
                          { label: '暂停', value: 'suspended' },
                          { label: '过期', value: 'expired' }
                        ]
                      },
                      {
                        type: 'textarea',
                        name: 'reason',
                        label: '变更原因',
                        required: true,
                        placeholder: '请输入状态变更的原因'
                      }
                    ]
                  }
                }
              },
              {
                type: 'button',
                label: '配置',
                level: 'link',
                actionType: 'dialog',
                visibleOn: canManageTenantConfig,
                dialog: {
                  title: '租户配置管理',
                  size: 'lg',
                  body: {
                    type: 'service',
                    api: '/api/v1/tenants/${id}/config',
                    body: [
                      {
                        type: 'form',
                        api: 'put:/api/v1/tenants/${id}/config',
                        body: [
                          {
                            type: 'input-number',
                            name: 'max_users',
                            label: '最大用户数',
                            min: 1,
                            max: 10000,
                            value: 100
                          },
                          {
                            type: 'input-number',
                            name: 'max_merchants',
                            label: '最大商户数',
                            min: 1,
                            max: 1000,
                            value: 50
                          },
                          {
                            type: 'checkboxes',
                            name: 'features',
                            label: '功能特性',
                            options: [
                              { label: '基础功能', value: 'basic' },
                              { label: '高级报表', value: 'advanced_report' },
                              { label: '批量操作', value: 'batch_operation' },
                              { label: 'API接口', value: 'api_access' },
                              { label: '演示模式', value: 'demo' }
                            ]
                          },
                          {
                            type: 'combo',
                            name: 'settings',
                            label: '自定义设置',
                            multiple: true,
                            multiLine: true,
                            items: [
                              {
                                type: 'input-text',
                                name: 'key',
                                label: '设置键',
                                required: true
                              },
                              {
                                type: 'input-text',
                                name: 'value',
                                label: '设置值',
                                required: true
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
                }
              }
            ]
          }
        ],
        headerToolbar: [
          ...(canCreateTenant ? [{
            type: 'button',
            label: '新增租户',
            level: 'primary',
            actionType: 'dialog',
            dialog: {
              title: '新增租户',
              body: {
                type: 'form',
                api: 'post:/api/v1/tenants',
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
                            }
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
                            description: '只能包含字母、数字、下划线和连字符'
                          },
                          {
                            type: 'input-text',
                            name: 'business_type',
                            label: '业务类型',
                            required: true
                          },
                          {
                            type: 'input-text',
                            name: 'contact_person',
                            label: '联系人',
                            required: true
                          }
                        ]
                      },
                      {
                        md: 6,
                        body: [
                          {
                            type: 'input-email',
                            name: 'contact_email',
                            label: '联系邮箱',
                            required: true
                          },
                          {
                            type: 'input-text',
                            name: 'contact_phone',
                            label: '联系电话',
                            validations: {
                              isPhoneNumber: true
                            }
                          },
                          {
                            type: 'textarea',
                            name: 'address',
                            label: '地址'
                          }
                        ]
                      }
                    ]
                  }
                ]
              }
            }
          }] : []),
          'bulkActions',
          'pagination'
        ]
      }
    ]
  }

  return <AmisRenderer schema={schema} />
  
  // 暂时注释确认对话框，未来用于敏感操作
  // return (
  //   <>
  //     <AmisRenderer schema={schema} />
  //     <ConfirmationDialog
  //       isOpen={confirmDialog.isOpen}
  //       title={confirmDialog.title}
  //       message={confirmDialog.message}
  //       type={confirmDialog.type}
  //       onConfirm={confirmDialog.onConfirm}
  //       onCancel={hideConfirmDialog}
  //     />
  //   </>
  // )
}

export default TenantListPage
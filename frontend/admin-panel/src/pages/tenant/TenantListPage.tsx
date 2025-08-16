import AmisRenderer from '../../components/ui/AmisRenderer'

const TenantListPage = () => {
  const schema = {
    type: 'page',
    title: '租户管理',
    body: [
      {
        type: 'crud',
        api: '/api/v1/tenants',
        columns: [
          {
            name: 'id',
            label: 'ID',
            type: 'text'
          },
          {
            name: 'name',
            label: '租户名称',
            type: 'text'
          },
          {
            name: 'code',
            label: '租户代码',
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
              'inactive': {
                value: 'inactive',
                label: '停用',
                level: 'danger'
              }
            }
          },
          {
            name: 'created_at',
            label: '创建时间',
            type: 'datetime'
          },
          {
            type: 'operation',
            label: '操作',
            buttons: [
              {
                type: 'button',
                label: '编辑',
                level: 'link',
                actionType: 'dialog',
                dialog: {
                  title: '编辑租户',
                  body: {
                    type: 'form',
                    api: 'put:/api/v1/tenants/${id}',
                    body: [
                      {
                        type: 'input-text',
                        name: 'name',
                        label: '租户名称',
                        required: true
                      },
                      {
                        type: 'input-text',
                        name: 'code',
                        label: '租户代码',
                        required: true
                      },
                      {
                        type: 'select',
                        name: 'status',
                        label: '状态',
                        options: [
                          { label: '激活', value: 'active' },
                          { label: '停用', value: 'inactive' }
                        ]
                      }
                    ]
                  }
                }
              },
              {
                type: 'button',
                label: '删除',
                level: 'link',
                className: 'text-red-600',
                actionType: 'ajax',
                api: 'delete:/api/v1/tenants/${id}',
                confirmText: '确定要删除此租户吗？'
              }
            ]
          }
        ],
        headerToolbar: [
          {
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
                    type: 'input-text',
                    name: 'name',
                    label: '租户名称',
                    required: true
                  },
                  {
                    type: 'input-text',
                    name: 'code',
                    label: '租户代码',
                    required: true
                  },
                  {
                    type: 'select',
                    name: 'status',
                    label: '状态',
                    value: 'active',
                    options: [
                      { label: '激活', value: 'active' },
                      { label: '停用', value: 'inactive' }
                    ]
                  }
                ]
              }
            }
          }
        ]
      }
    ]
  }

  return <AmisRenderer schema={schema} />
}

export default TenantListPage
// 商品分类管理页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const CategoryManagePage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '分类管理',
    toolbar: [
      {
        type: 'button',
        actionType: 'dialog',
        label: '新增分类',
        level: 'primary',
        dialog: {
          title: '新增分类',
          body: {
            type: 'form',
            api: {
              method: 'post',
              url: '/api/v1/categories',
              messages: {
                success: '分类创建成功',
                failed: '分类创建失败'
              }
            },
            controls: [
              {
                type: 'text',
                name: 'name',
                label: '分类名称',
                required: true,
                validations: {
                  maxLength: 100
                },
                validationErrors: {
                  required: '请输入分类名称',
                  maxLength: '分类名称不能超过100个字符'
                }
              },
              {
                type: 'tree-select',
                name: 'parent_id',
                label: '上级分类',
                placeholder: '留空表示创建顶级分类',
                source: {
                  method: 'get',
                  url: '/api/v1/categories/tree'
                },
                labelField: 'name',
                valueField: 'id'
              },
              {
                type: 'input-number',
                name: 'sort_order',
                label: '排序',
                value: 0,
                description: '数字越小排序越靠前'
              }
            ]
          }
        }
      }
    ],
    body: [
      {
        type: 'crud',
        api: {
          method: 'get',
          url: '/api/v1/categories/tree'
        },
        mode: 'table',
        expandable: true,
        autoGenerateFilter: false,
        headerToolbar: [
          'bulk-actions',
          {
            type: 'columns-toggler',
            align: 'right'
          }
        ],
        bulkActions: [
          {
            label: '批量启用',
            actionType: 'ajax',
            api: {
              method: 'put',
              url: '/api/v1/categories/batch',
              data: {
                ids: '${ids}',
                status: 1
              }
            },
            confirmText: '确定要启用选中的分类吗？'
          },
          {
            label: '批量禁用',
            actionType: 'ajax',
            api: {
              method: 'put',
              url: '/api/v1/categories/batch',
              data: {
                ids: '${ids}',
                status: 0
              }
            },
            confirmText: '确定要禁用选中的分类吗？'
          },
          {
            label: '批量删除',
            actionType: 'ajax',
            api: {
              method: 'delete',
              url: '/api/v1/categories/batch',
              data: {
                ids: '${ids}'
              }
            },
            confirmText: '确定要删除选中的分类吗？此操作不可恢复，且会影响关联的商品！'
          }
        ],
        columns: [
          {
            name: 'id',
            label: 'ID',
            type: 'text',
            width: 60
          },
          {
            name: 'name',
            label: '分类名称',
            type: 'text',
            searchable: true
          },
          {
            name: 'level',
            label: '层级',
            type: 'text',
            width: 60
          },
          {
            name: 'path',
            label: '完整路径',
            type: 'text'
          },
          {
            name: 'sort_order',
            label: '排序',
            type: 'text',
            width: 60
          },
          {
            name: 'status',
            label: '状态',
            type: 'status',
            width: 80,
            map: {
              1: {
                label: '启用',
                level: 'success'
              },
              0: {
                label: '禁用',
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
                label: '编辑',
                level: 'link',
                actionType: 'dialog',
                dialog: {
                  title: '编辑分类',
                  body: {
                    type: 'form',
                    api: {
                      method: 'put',
                      url: '/api/v1/categories/${id}',
                      messages: {
                        success: '分类更新成功',
                        failed: '分类更新失败'
                      }
                    },
                    initApi: {
                      method: 'get',
                      url: '/api/v1/categories/${id}'
                    },
                    controls: [
                      {
                        type: 'text',
                        name: 'name',
                        label: '分类名称',
                        required: true,
                        validations: {
                          maxLength: 100
                        }
                      },
                      {
                        type: 'tree-select',
                        name: 'parent_id',
                        label: '上级分类',
                        source: {
                          method: 'get',
                          url: '/api/v1/categories/tree'
                        },
                        labelField: 'name',
                        valueField: 'id',
                        disabledOn: 'this.children && this.children.length > 0'
                      },
                      {
                        type: 'input-number',
                        name: 'sort_order',
                        label: '排序'
                      },
                      {
                        type: 'switch',
                        name: 'status',
                        label: '启用状态',
                        trueValue: 1,
                        falseValue: 0
                      }
                    ]
                  }
                }
              },
              {
                type: 'button',
                label: '子分类',
                level: 'link',
                actionType: 'dialog',
                dialog: {
                  title: '子分类管理',
                  size: 'lg',
                  body: {
                    type: 'crud',
                    api: {
                      method: 'get',
                      url: '/api/v1/categories/${id}/children'
                    },
                    headerToolbar: [
                      {
                        type: 'button',
                        actionType: 'dialog',
                        label: '新增子分类',
                        level: 'primary',
                        dialog: {
                          title: '新增子分类',
                          body: {
                            type: 'form',
                            api: {
                              method: 'post',
                              url: '/api/v1/categories',
                              data: {
                                parent_id: '${id}'
                              }
                            },
                            controls: [
                              {
                                type: 'hidden',
                                name: 'parent_id',
                                value: '${id}'
                              },
                              {
                                type: 'text',
                                name: 'name',
                                label: '分类名称',
                                required: true,
                                validations: {
                                  maxLength: 100
                                }
                              },
                              {
                                type: 'input-number',
                                name: 'sort_order',
                                label: '排序',
                                value: 0
                              }
                            ]
                          }
                        }
                      }
                    ],
                    columns: [
                      {
                        name: 'name',
                        label: '分类名称',
                        type: 'text'
                      },
                      {
                        name: 'sort_order',
                        label: '排序',
                        type: 'text'
                      },
                      {
                        name: 'status',
                        label: '状态',
                        type: 'status',
                        map: {
                          1: { label: '启用', level: 'success' },
                          0: { label: '禁用', level: 'danger' }
                        }
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
                              title: '编辑子分类',
                              body: {
                                type: 'form',
                                api: {
                                  method: 'put',
                                  url: '/api/v1/categories/${id}'
                                },
                                initApi: {
                                  method: 'get',
                                  url: '/api/v1/categories/${id}'
                                },
                                controls: [
                                  {
                                    type: 'text',
                                    name: 'name',
                                    label: '分类名称',
                                    required: true
                                  },
                                  {
                                    type: 'input-number',
                                    name: 'sort_order',
                                    label: '排序'
                                  },
                                  {
                                    type: 'switch',
                                    name: 'status',
                                    label: '启用状态',
                                    trueValue: 1,
                                    falseValue: 0
                                  }
                                ]
                              }
                            }
                          },
                          {
                            type: 'button',
                            label: '删除',
                            level: 'link',
                            className: 'text-danger',
                            actionType: 'ajax',
                            api: {
                              method: 'delete',
                              url: '/api/v1/categories/${id}'
                            },
                            confirmText: '确定要删除这个分类吗？此操作不可恢复！'
                          }
                        ]
                      }
                    ]
                  }
                }
              },
              {
                type: 'dropdown-button',
                label: '更多',
                level: 'link',
                trigger: 'hover',
                buttons: [
                  {
                    type: 'button',
                    label: '分类路径',
                    actionType: 'dialog',
                    dialog: {
                      title: '分类路径',
                      body: {
                        type: 'crud',
                        api: {
                          method: 'get',
                          url: '/api/v1/categories/${id}/path'
                        },
                        columns: [
                          {
                            name: 'level',
                            label: '层级',
                            type: 'text'
                          },
                          {
                            name: 'name',
                            label: '分类名称',
                            type: 'text'
                          }
                        ]
                      }
                    }
                  },
                  {
                    type: 'button',
                    label: '启用',
                    visibleOn: 'this.status === 0',
                    actionType: 'ajax',
                    api: {
                      method: 'put',
                      url: '/api/v1/categories/${id}',
                      data: {
                        status: 1
                      }
                    },
                    confirmText: '确定要启用这个分类吗？'
                  },
                  {
                    type: 'button',
                    label: '禁用',
                    visibleOn: 'this.status === 1',
                    actionType: 'ajax',
                    api: {
                      method: 'put',
                      url: '/api/v1/categories/${id}',
                      data: {
                        status: 0
                      }
                    },
                    confirmText: '确定要禁用这个分类吗？'
                  },
                  {
                    type: 'button',
                    label: '删除',
                    level: 'danger',
                    actionType: 'ajax',
                    api: {
                      method: 'delete',
                      url: '/api/v1/categories/${id}'
                    },
                    confirmText: '确定要删除这个分类吗？此操作不可恢复，且会影响关联的商品！',
                    disabledOn: 'this.children && this.children.length > 0'
                  }
                ]
              }
            ]
          }
        ]
      }
    ]
  };

  return <AmisRenderer schema={schema} />;
};

export default CategoryManagePage;
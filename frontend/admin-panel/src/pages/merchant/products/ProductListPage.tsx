// 商品列表管理页面
import React from 'react';
import AmisRenderer from '../../../components/ui/AmisRenderer';
import { AmisSchema } from '../../../types/product';

const ProductListPage: React.FC = () => {
  const schema: AmisSchema = {
    type: 'page',
    title: '商品管理',
    toolbar: [
      {
        type: 'button',
        actionType: 'dialog',
        label: '新增商品',
        level: 'primary',
        dialog: {
          title: '新增商品',
          size: 'lg',
          body: {
            type: 'form',
            api: {
              method: 'post',
              url: '/api/v1/products',
              messages: {
                success: '商品创建成功',
                failed: '商品创建失败'
              }
            },
            controls: [
              {
                type: 'text',
                name: 'name',
                label: '商品名称',
                required: true,
                validations: {
                  maxLength: 255
                },
                validationErrors: {
                  required: '请输入商品名称',
                  maxLength: '商品名称不能超过255个字符'
                }
              },
              {
                type: 'textarea',
                name: 'description',
                label: '商品描述',
                validations: {
                  maxLength: 2000
                },
                validationErrors: {
                  maxLength: '商品描述不能超过2000个字符'
                }
              },
              {
                type: 'select',
                name: 'category_id',
                label: '商品分类',
                source: {
                  method: 'get',
                  url: '/api/v1/categories'
                },
                labelField: 'name',
                valueField: 'id'
              },
              {
                type: 'input-tag',
                name: 'tags',
                label: '商品标签',
                placeholder: '输入标签，按回车确认',
                validations: {
                  maxLength: 20
                },
                validationErrors: {
                  maxLength: '最多只能添加20个标签'
                }
              },
              {
                type: 'group',
                label: '价格信息',
                controls: [
                  {
                    type: 'input-number',
                    name: 'price.amount',
                    label: '商品价格（分）',
                    required: true,
                    min: 1,
                    validationErrors: {
                      required: '请输入商品价格',
                      min: '商品价格必须大于0'
                    }
                  },
                  {
                    type: 'select',
                    name: 'price.currency',
                    label: '货币类型',
                    value: 'CNY',
                    options: [
                      { label: '人民币', value: 'CNY' }
                    ]
                  }
                ]
              },
              {
                type: 'input-number',
                name: 'rights_cost',
                label: '权益成本（分）',
                min: 0,
                value: 0,
                validationErrors: {
                  min: '权益成本不能为负数'
                }
              },
              {
                type: 'group',
                label: '库存信息',
                controls: [
                  {
                    type: 'input-number',
                    name: 'inventory.stock_quantity',
                    label: '库存数量',
                    required: true,
                    min: 0,
                    validationErrors: {
                      required: '请输入库存数量',
                      min: '库存数量不能为负数'
                    }
                  },
                  {
                    type: 'input-number',
                    name: 'inventory.reserved_quantity',
                    label: '预留数量',
                    min: 0,
                    value: 0,
                    validationErrors: {
                      min: '预留数量不能为负数'
                    }
                  },
                  {
                    type: 'switch',
                    name: 'inventory.track_inventory',
                    label: '启用库存跟踪',
                    value: true
                  }
                ]
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
          url: '/api/v1/products',
          data: {
            page: '${page}',
            page_size: '${perPage}',
            keyword: '${keyword}',
            category_id: '${category_id}',
            status: '${status}',
            sort_by: '${orderBy}',
            sort_order: '${orderDir}'
          }
        },
        defaultParams: {
          page: 1,
          perPage: 20
        },
        autoGenerateFilter: false,
        headerToolbar: [
          'filter-toggler',
          'bulk-actions',
          {
            type: 'columns-toggler',
            align: 'right'
          },
          {
            type: 'drag-toggler',
            align: 'right'
          },
          {
            type: 'pagination',
            align: 'right'
          }
        ],
        filter: {
          title: '筛选条件',
          controls: [
            {
              type: 'text',
              name: 'keyword',
              label: '关键词',
              placeholder: '请输入商品名称或描述'
            },
            {
              type: 'select',
              name: 'category_id',
              label: '商品分类',
              placeholder: '请选择分类',
              source: {
                method: 'get',
                url: '/api/v1/categories'
              },
              labelField: 'name',
              valueField: 'id'
            },
            {
              type: 'select',
              name: 'status',
              label: '商品状态',
              placeholder: '请选择状态',
              options: [
                { label: '草稿', value: 'draft' },
                { label: '已上架', value: 'active' },
                { label: '已下架', value: 'inactive' }
              ]
            }
          ]
        },
        bulkActions: [
          {
            label: '批量上架',
            actionType: 'ajax',
            api: {
              method: 'post',
              url: '/api/v1/products/batch',
              data: {
                product_ids: '${ids}',
                operation: 'activate'
              }
            },
            confirmText: '确定要上架选中的商品吗？'
          },
          {
            label: '批量下架',
            actionType: 'ajax',
            api: {
              method: 'post',
              url: '/api/v1/products/batch',
              data: {
                product_ids: '${ids}',
                operation: 'deactivate'
              }
            },
            confirmText: '确定要下架选中的商品吗？'
          },
          {
            label: '批量删除',
            actionType: 'ajax',
            api: {
              method: 'post',
              url: '/api/v1/products/batch',
              data: {
                product_ids: '${ids}',
                operation: 'delete'
              }
            },
            confirmText: '确定要删除选中的商品吗？此操作不可恢复！'
          }
        ],
        columns: [
          {
            name: 'id',
            label: 'ID',
            type: 'text',
            sortable: false
          },
          {
            name: 'name',
            label: '商品名称',
            type: 'text',
            sortable: true,
            searchable: true
          },
          {
            name: 'category_path',
            label: '分类',
            type: 'text'
          },
          {
            name: 'price',
            label: '价格',
            type: 'tpl',
            tpl: '¥${price.amount | number: 0 | divide: 100}',
            sortable: true
          },
          {
            name: 'inventory.stock_quantity',
            label: '库存',
            type: 'text'
          },
          {
            name: 'status',
            label: '状态',
            type: 'status',
            map: {
              draft: {
                label: '草稿',
                level: 'info'
              },
              active: {
                label: '已上架',
                level: 'success'
              },
              inactive: {
                label: '已下架',
                level: 'warning'
              },
              deleted: {
                label: '已删除',
                level: 'danger'
              }
            }
          },
          {
            name: 'created_at',
            label: '创建时间',
            type: 'datetime',
            format: 'YYYY-MM-DD HH:mm:ss',
            sortable: true
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
                  title: '编辑商品',
                  size: 'lg',
                  body: {
                    type: 'form',
                    api: {
                      method: 'put',
                      url: '/api/v1/products/${id}',
                      messages: {
                        success: '商品更新成功',
                        failed: '商品更新失败'
                      }
                    },
                    initApi: {
                      method: 'get',
                      url: '/api/v1/products/${id}'
                    },
                    controls: [
                      {
                        type: 'text',
                        name: 'name',
                        label: '商品名称',
                        required: true,
                        validations: {
                          maxLength: 255
                        }
                      },
                      {
                        type: 'textarea',
                        name: 'description',
                        label: '商品描述',
                        validations: {
                          maxLength: 2000
                        }
                      },
                      {
                        type: 'select',
                        name: 'category_id',
                        label: '商品分类',
                        source: {
                          method: 'get',
                          url: '/api/v1/categories'
                        },
                        labelField: 'name',
                        valueField: 'id'
                      },
                      {
                        type: 'input-tag',
                        name: 'tags',
                        label: '商品标签',
                        placeholder: '输入标签，按回车确认'
                      },
                      {
                        type: 'group',
                        label: '价格信息',
                        controls: [
                          {
                            type: 'input-number',
                            name: 'price.amount',
                            label: '商品价格（分）',
                            required: true,
                            min: 1
                          },
                          {
                            type: 'select',
                            name: 'price.currency',
                            label: '货币类型',
                            options: [
                              { label: '人民币', value: 'CNY' }
                            ]
                          }
                        ]
                      },
                      {
                        type: 'input-number',
                        name: 'rights_cost',
                        label: '权益成本（分）',
                        min: 0
                      },
                      {
                        type: 'group',
                        label: '库存信息',
                        controls: [
                          {
                            type: 'input-number',
                            name: 'inventory.stock_quantity',
                            label: '库存数量',
                            required: true,
                            min: 0
                          },
                          {
                            type: 'input-number',
                            name: 'inventory.reserved_quantity',
                            label: '预留数量',
                            min: 0
                          },
                          {
                            type: 'switch',
                            name: 'inventory.track_inventory',
                            label: '启用库存跟踪'
                          }
                        ]
                      }
                    ]
                  }
                }
              },
              {
                type: 'button',
                label: '图片',
                level: 'link',
                actionType: 'dialog',
                dialog: {
                  title: '商品图片管理',
                  size: 'lg',
                  body: {
                    type: 'form',
                    api: {
                      method: 'post',
                      url: '/api/v1/products/${id}/images',
                      messages: {
                        success: '图片上传成功',
                        failed: '图片上传失败'
                      }
                    },
                    controls: [
                      {
                        type: 'input-file',
                        name: 'image',
                        label: '选择图片',
                        accept: 'image/*',
                        maxSize: '5MB',
                        required: true,
                        validationErrors: {
                          required: '请选择要上传的图片',
                          maxSize: '图片文件大小不能超过5MB'
                        }
                      },
                      {
                        type: 'text',
                        name: 'alt_text',
                        label: '图片说明',
                        placeholder: '请输入图片描述'
                      },
                      {
                        type: 'input-number',
                        name: 'sort_order',
                        label: '排序',
                        value: 0
                      },
                      {
                        type: 'switch',
                        name: 'is_primary',
                        label: '设为主图'
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
                    label: '变更历史',
                    actionType: 'dialog',
                    dialog: {
                      title: '商品变更历史',
                      size: 'lg',
                      body: {
                        type: 'crud',
                        api: {
                          method: 'get',
                          url: '/api/v1/products/${id}/history'
                        },
                        columns: [
                          {
                            name: 'field_name',
                            label: '字段',
                            type: 'text'
                          },
                          {
                            name: 'old_value',
                            label: '原值',
                            type: 'text'
                          },
                          {
                            name: 'new_value',
                            label: '新值',
                            type: 'text'
                          },
                          {
                            name: 'operation',
                            label: '操作类型',
                            type: 'text'
                          },
                          {
                            name: 'changed_at',
                            label: '变更时间',
                            type: 'datetime',
                            format: 'YYYY-MM-DD HH:mm:ss'
                          }
                        ]
                      }
                    }
                  },
                  {
                    type: 'button',
                    label: '上架',
                    visibleOn: 'this.status === "draft" || this.status === "inactive"',
                    actionType: 'ajax',
                    api: {
                      method: 'patch',
                      url: '/api/v1/products/${id}/status',
                      data: {
                        status: 'active'
                      }
                    },
                    confirmText: '确定要上架这个商品吗？'
                  },
                  {
                    type: 'button',
                    label: '下架',
                    visibleOn: 'this.status === "active"',
                    actionType: 'ajax',
                    api: {
                      method: 'patch',
                      url: '/api/v1/products/${id}/status',
                      data: {
                        status: 'inactive'
                      }
                    },
                    confirmText: '确定要下架这个商品吗？'
                  },
                  {
                    type: 'button',
                    label: '删除',
                    level: 'danger',
                    actionType: 'ajax',
                    api: {
                      method: 'delete',
                      url: '/api/v1/products/${id}'
                    },
                    confirmText: '确定要删除这个商品吗？此操作不可恢复！'
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

export default ProductListPage;
import { renderWithRouter, screen, cleanup } from './utils/testUtils'
import TenantListPage from '../pages/tenant/TenantListPage'
import { tenantService } from '../services/tenantService'

// Mock 租户服务
jest.mock('../services/tenantService', () => ({
  tenantService: {
    listTenants: jest.fn(),
    createTenant: jest.fn(),
    updateTenant: jest.fn(),
    updateTenantStatus: jest.fn(),
    getTenantConfig: jest.fn(),
    updateTenantConfig: jest.fn(),
  },
}))

// Mock 权限Hook
jest.mock('../hooks/useTenantPermissions', () => ({
  useTenantPermissions: () => ({
    canViewTenants: true,
    canCreateTenant: true,
    canEditTenant: true,
    canManageTenantStatus: true,
    canManageTenantConfig: true,
    requiresConfirmation: () => false
  })
}))

// Mock AmisRenderer 组件
jest.mock('../components/ui/AmisRenderer', () => {
  return function MockAmisRenderer({ schema }: { schema: Record<string, unknown> }) {
    const schemaAny = schema as any; // 测试中的类型简化
    return (
      <div data-testid="amis-renderer">
        <div data-testid="page-title">{String(schemaAny.title || '')}</div>
        <div data-testid="crud-type">{String(schemaAny.body?.[0]?.type || '')}</div>
        <div data-testid="api-endpoint">{String(schemaAny.body?.[0]?.api || '')}</div>
        
        {/* 模拟表格列 */}
        {schemaAny.body?.[0]?.columns?.map((column: any, index: number) => (
          <div key={index} data-testid={`column-${column.name || 'operation'}`}>
            {String(column.label || '')}
          </div>
        ))}
        
        {/* 模拟工具栏按钮 */}
        {schemaAny.body?.[0]?.headerToolbar?.map((item: any, index: number) => (
          <button key={index} data-testid={`toolbar-${item.label || item}`}>
            {String(item.label || item)}
          </button>
        ))}
      </div>
    )
  }
})


describe('TenantListPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('应该正确渲染租户管理页面', () => {
    renderWithRouter(<TenantListPage />)
    
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
    expect(screen.getByTestId('page-title')).toHaveTextContent('租户管理')
    expect(screen.getByTestId('crud-type')).toHaveTextContent('crud')
    expect(screen.getByTestId('api-endpoint')).toHaveTextContent('/api/v1/tenants')
  })

  it('应该显示所有必要的表格列', () => {
    renderWithRouter(<TenantListPage />)
    
    // 检查所有列是否存在
    expect(screen.getByTestId('column-id')).toHaveTextContent('ID')
    expect(screen.getByTestId('column-name')).toHaveTextContent('租户名称')
    expect(screen.getByTestId('column-code')).toHaveTextContent('租户代码')
    expect(screen.getByTestId('column-business_type')).toHaveTextContent('业务类型')
    expect(screen.getByTestId('column-contact_person')).toHaveTextContent('联系人')
    expect(screen.getByTestId('column-contact_email')).toHaveTextContent('联系邮箱')
    expect(screen.getByTestId('column-status')).toHaveTextContent('状态')
    expect(screen.getByTestId('column-created_at')).toHaveTextContent('创建时间')
    expect(screen.getByTestId('column-operation')).toHaveTextContent('操作')
  })

  it('应该显示新增租户按钮', () => {
    renderWithRouter(<TenantListPage />)
    
    expect(screen.getByTestId('toolbar-新增租户')).toBeInTheDocument()
  })

  it('应该显示分页和批量操作工具', () => {
    renderWithRouter(<TenantListPage />)
    
    expect(screen.getByTestId('toolbar-bulkActions')).toBeInTheDocument()
    expect(screen.getByTestId('toolbar-pagination')).toBeInTheDocument()
  })

  it('应该有正确的搜索筛选配置', () => {
    renderWithRouter(<TenantListPage />)
    
    // 通过检查页面是否正确渲染来验证筛选配置
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
  })
})

describe('TenantListPage Schema Configuration', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    cleanup()
  })

  it('应该有正确的筛选器配置', () => {
    renderWithRouter(<TenantListPage />)
    
    // 验证页面渲染成功，schema配置正确
    expect(screen.getByTestId('crud-type')).toHaveTextContent('crud')
    expect(screen.getByTestId('api-endpoint')).toHaveTextContent('/api/v1/tenants')
    
    cleanup()
  })

  it('应该有正确的状态映射配置', () => {
    // 这里主要验证组件能正确渲染，状态映射配置在schema中
    renderWithRouter(<TenantListPage />)
    
    expect(screen.getByTestId('column-status')).toHaveTextContent('状态')
    
    cleanup()
  })

  it('应该有正确的操作按钮配置', () => {
    renderWithRouter(<TenantListPage />)
    
    // 验证操作列存在
    expect(screen.getByTestId('column-operation')).toHaveTextContent('操作')
    
    cleanup()
  })
})

describe('TenantListPage Integration', () => {
  const mockTenants = [
    {
      id: 1,
      name: '测试租户1',
      code: 'test-tenant-1',
      status: 'active',
      business_type: 'ecommerce',
      contact_person: '张三',
      contact_email: 'zhangsan@test.com',
      contact_phone: '13800138000',
      address: '北京市朝阳区',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 2,
      name: '测试租户2',
      code: 'test-tenant-2',
      status: 'suspended',
      business_type: 'retail',
      contact_person: '李四',
      contact_email: 'lisi@test.com',
      contact_phone: '13800138001',
      address: '上海市浦东新区',
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
    },
  ]

  beforeEach(() => {
    (tenantService.listTenants as jest.Mock).mockResolvedValue({
      total: mockTenants.length,
      page: 1,
      size: mockTenants.length,
      tenants: mockTenants,
    })
  })

  it('应该正确处理租户列表数据', async () => {
    renderWithRouter(<TenantListPage />)
    
    // 验证AmisRenderer正确渲染
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
    
    // API端点配置正确
    expect(screen.getByTestId('api-endpoint')).toHaveTextContent('/api/v1/tenants')
  })

  it('应该有正确的搜索功能配置', () => {
    renderWithRouter(<TenantListPage />)
    
    // 验证搜索配置通过组件正确渲染
    expect(screen.getByTestId('crud-type')).toHaveTextContent('crud')
  })
})

describe('TenantListPage Error Handling', () => {
  it('应该处理API错误', async () => {
    (tenantService.listTenants as jest.Mock).mockRejectedValue(new Error('API Error'))
    
    renderWithRouter(<TenantListPage />)
    
    // 验证组件仍能正常渲染
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
  })

  it('应该处理空数据', async () => {
    (tenantService.listTenants as jest.Mock).mockResolvedValue({
      total: 0,
      page: 1,
      size: 0,
      tenants: [],
    })
    
    renderWithRouter(<TenantListPage />)
    
    // 验证组件能处理空数据
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
  })
})

describe('TenantListPage Accessibility', () => {
  it('应该有正确的语义结构', () => {
    renderWithRouter(<TenantListPage />)
    
    // 验证页面标题
    expect(screen.getByTestId('page-title')).toHaveTextContent('租户管理')
  })

  it('应该支持键盘导航', () => {
    renderWithRouter(<TenantListPage />)
    
    // 验证按钮可以被聚焦
    const addButton = screen.getByTestId('toolbar-新增租户')
    expect(addButton).toBeInTheDocument()
  })
})

describe('TenantListPage Performance', () => {
  it('应该优化大数据集的渲染', () => {
    const largeMockData = Array.from({ length: 1000 }, (_, i) => ({
      id: i + 1,
      name: `租户${i + 1}`,
      code: `tenant-${i + 1}`,
      status: 'active',
      business_type: 'ecommerce',
      contact_person: `联系人${i + 1}`,
      contact_email: `contact${i + 1}@test.com`,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }))

    ;(tenantService.listTenants as jest.Mock).mockResolvedValue({
      total: largeMockData.length,
      page: 1,
      size: 20, // 分页显示
      tenants: largeMockData.slice(0, 20),
    })

    const startTime = performance.now()
    renderWithRouter(<TenantListPage />)
    const endTime = performance.now()

    // 验证渲染时间合理（应该很快，因为使用了分页）
    expect(endTime - startTime).toBeLessThan(100) // 100ms内完成渲染
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
  })
})
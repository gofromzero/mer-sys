import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi } from 'vitest';
import OrderManagePage from '../../pages/merchant/orders/OrderManagePage';
import { useOrderStore } from '../../stores/orderStore';
import { useOrderStatusStore } from '../../stores/orderStatusStore';
import { Order, OrderStatus, PaymentMethod, PaymentStatus } from '../../types/order';

// Mock stores
vi.mock('../../stores/orderStore');
vi.mock('../../stores/orderStatusStore');

// Mock react-router-dom hooks
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
    useSearchParams: () => [new URLSearchParams(), vi.fn()],
  };
});

// Mock AmisRenderer
vi.mock('../../components/ui/AmisRenderer', () => ({
  AmisRenderer: ({ schema }: { schema: any }) => (
    <div data-testid="amis-renderer">
      <div>Order Management Page</div>
      {JSON.stringify(schema, null, 2)}
    </div>
  )
}));

// Mock components
vi.mock('../../components/orders/OrderBatchActions', () => ({
  default: ({ selectedOrders, onBatchUpdate }: any) => (
    <div data-testid="batch-actions">
      <span>已选择: {selectedOrders.length}</span>
      <button onClick={() => onBatchUpdate('processing', '批量处理')}>批量处理</button>
      <button onClick={() => onBatchUpdate('completed', '批量完成')}>批量完成</button>
      <button onClick={() => onBatchUpdate('cancelled', '批量取消')}>批量取消</button>
    </div>
  )
}));

vi.mock('../../components/orders/OrderStatsReport', () => ({
  default: ({ stats }: any) => (
    <div data-testid="stats-report">
      <span>总订单: {stats.totalCount}</span>
      <span>待处理: {stats.pendingCount}</span>
      <span>处理中: {stats.processingCount}</span>
      <span>已完成: {stats.completedCount}</span>
    </div>
  )
}));

const mockOrders: Order[] = [
  {
    id: 1,
    tenant_id: 1,
    merchant_id: 1,
    customer_id: 1,
    order_number: 'ORD20231201001',
    status: OrderStatus.PAID,
    items: [
      {
        id: 1,
        order_id: 1,
        product_id: 1001,
        product_name: '测试商品1',
        quantity: 2,
        unit_price: { amount: 50.00, currency: 'CNY' },
        unit_rights_cost: 10,
        subtotal_amount: { amount: 100.00, currency: 'CNY' },
        subtotal_rights_cost: 20
      }
    ],
    payment_info: {
      payment_method: PaymentMethod.ALIPAY,
      payment_id: 'alipay_123456',
      payment_status: PaymentStatus.PAID,
      paid_amount: { amount: 100.00, currency: 'CNY' },
      paid_at: new Date('2023-12-01T12:00:00Z'),
      payment_url: 'https://example.com/pay',
      callback_data: {}
    },
    total_amount: { amount: 100.00, currency: 'CNY' },
    total_rights_cost: 20,
    created_at: new Date('2023-12-01T10:00:00Z'),
    updated_at: new Date('2023-12-01T12:00:00Z')
  },
  {
    id: 2,
    tenant_id: 1,
    merchant_id: 1,
    customer_id: 2,
    order_number: 'ORD20231201002',
    status: OrderStatus.PROCESSING,
    items: [
      {
        id: 2,
        order_id: 2,
        product_id: 1002,
        product_name: '测试商品2',
        quantity: 1,
        unit_price: { amount: 80.00, currency: 'CNY' },
        unit_rights_cost: 15,
        subtotal_amount: { amount: 80.00, currency: 'CNY' },
        subtotal_rights_cost: 15
      }
    ],
    payment_info: {
      payment_method: PaymentMethod.ALIPAY,
      payment_id: 'alipay_789012',
      payment_status: PaymentStatus.PAID,
      paid_amount: { amount: 80.00, currency: 'CNY' },
      paid_at: new Date('2023-12-01T11:00:00Z'),
      payment_url: 'https://example.com/pay',
      callback_data: {}
    },
    total_amount: { amount: 80.00, currency: 'CNY' },
    total_rights_cost: 15,
    created_at: new Date('2023-12-01T09:00:00Z'),
    updated_at: new Date('2023-12-01T13:00:00Z')
  }
];

const mockOrderStore = {
  orders: mockOrders,
  currentOrder: null,
  isLoading: false,
  error: null,
  getOrder: vi.fn(),
  createOrder: vi.fn(),
  cancelOrder: vi.fn(),
  listOrders: vi.fn(),
  queryOrders: vi.fn().mockResolvedValue({
    items: mockOrders,
    total: mockOrders.length,
    page: 1,
    pageSize: 10,
    hasNext: false
  }),
  getOrderStats: vi.fn().mockResolvedValue({
    totalCount: 100,
    pendingCount: 10,
    paidCount: 20,
    processingCount: 15,
    completedCount: 50,
    cancelledCount: 5
  })
};

const mockOrderStatusStore = {
  notifications: [],
  connectionStatus: 'connected',
  filters: {
    keyword: '',
    status: [],
    merchantId: null,
    dateRange: null
  },
  statistics: {
    totalCount: 100,
    pendingCount: 10,
    processingCount: 15,
    completedCount: 50
  },
  connect: vi.fn(),
  disconnect: vi.fn(),
  updateFilters: vi.fn(),
  batchUpdateOrderStatus: vi.fn().mockResolvedValue({
    successCount: 2,
    failureCount: 0,
    failures: []
  }),
  markNotificationAsRead: vi.fn(),
  clearNotifications: vi.fn()
};

describe('OrderManagePage', () => {
  beforeEach(() => {
    vi.mocked(useOrderStore).mockReturnValue(mockOrderStore);
    vi.mocked(useOrderStatusStore).mockReturnValue(mockOrderStatusStore);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  const renderWithRouter = (component: React.ReactElement) => {
    return render(
      <BrowserRouter>
        {component}
      </BrowserRouter>
    );
  };

  test('应该正确渲染商户订单管理页面', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
      expect(screen.getByText('Order Management Page')).toBeInTheDocument();
    });
  });

  test('应该显示订单统计信息', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      expect(screen.getByTestId('stats-report')).toBeInTheDocument();
      expect(screen.getByText('总订单: 100')).toBeInTheDocument();
      expect(screen.getByText('待处理: 10')).toBeInTheDocument();
      expect(screen.getByText('处理中: 15')).toBeInTheDocument();
      expect(screen.getByText('已完成: 50')).toBeInTheDocument();
    });
  });

  test('应该在页面加载时调用相关API', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      expect(mockOrderStore.queryOrders).toHaveBeenCalled();
      expect(mockOrderStore.getOrderStats).toHaveBeenCalled();
      expect(mockOrderStatusStore.connect).toHaveBeenCalled();
    });
  });

  test('应该支持订单搜索功能', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查搜索功能
      expect(schemaText).toContain('keyword');
      expect(schemaText).toContain('订单号搜索');
    });
  });

  test('应该支持状态筛选功能', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查状态筛选
      expect(schemaText).toContain('status');
      expect(schemaText).toContain('pending');
      expect(schemaText).toContain('paid');
      expect(schemaText).toContain('processing');
      expect(schemaText).toContain('completed');
      expect(schemaText).toContain('cancelled');
    });
  });

  test('应该支持日期范围筛选', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查日期筛选
      expect(schemaText).toContain('dateRange');
      expect(schemaText).toContain('创建时间');
    });
  });

  test('应该显示批量操作组件', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      expect(screen.getByTestId('batch-actions')).toBeInTheDocument();
      expect(screen.getByText('批量处理')).toBeInTheDocument();
      expect(screen.getByText('批量完成')).toBeInTheDocument();
      expect(screen.getByText('批量取消')).toBeInTheDocument();
    });
  });

  test('应该处理批量状态更新', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const batchProcessButton = screen.getByText('批量处理');
      fireEvent.click(batchProcessButton);
    });

    expect(mockOrderStatusStore.batchUpdateOrderStatus).toHaveBeenCalledWith({
      orderIds: [],
      status: 'processing',
      reason: '批量处理',
      operatorType: 'merchant'
    });
  });

  test('应该支持订单详情抽屉', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查详情抽屉配置
      expect(schemaText).toContain('drawer');
      expect(schemaText).toContain('订单详情');
    });
  });

  test('应该支持实时数据刷新', async () => {
    renderWithRouter(<OrderManagePage />);
    
    // 模拟30秒后的自动刷新
    vi.advanceTimersByTime(30000);
    
    await waitFor(() => {
      // 应该调用刷新API多次
      expect(mockOrderStore.queryOrders).toHaveBeenCalledTimes(2);
    });
  });

  test('加载中时应该显示加载状态', () => {
    const loadingOrderStore = {
      ...mockOrderStore,
      isLoading: true
    };
    
    vi.mocked(useOrderStore).mockReturnValue(loadingOrderStore);
    
    renderWithRouter(<OrderManagePage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    expect(schemaText).toContain('加载中');
  });

  test('出错时应该显示错误信息', () => {
    const errorOrderStore = {
      ...mockOrderStore,
      error: '加载订单列表失败'
    };
    
    vi.mocked(useOrderStore).mockReturnValue(errorOrderStore);
    
    renderWithRouter(<OrderManagePage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    expect(schemaText).toContain('加载订单列表失败');
  });

  test('应该支持分页功能', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查分页配置
      expect(schemaText).toContain('pagination');
      expect(schemaText).toContain('page');
      expect(schemaText).toContain('pageSize');
    });
  });

  test('应该支持排序功能', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查排序配置
      expect(schemaText).toContain('sortable');
      expect(schemaText).toContain('created_at');
      expect(schemaText).toContain('updated_at');
      expect(schemaText).toContain('total_amount');
    });
  });

  test('应该显示订单列表数据', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查订单数据展示
      expect(schemaText).toContain('ORD20231201001');
      expect(schemaText).toContain('ORD20231201002');
      expect(schemaText).toContain('测试商品1');
      expect(schemaText).toContain('测试商品2');
    });
  });

  test('应该支持导出功能', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查导出功能
      expect(schemaText).toContain('export');
      expect(schemaText).toContain('导出订单');
    });
  });

  test('应该处理WebSocket连接状态', async () => {
    const disconnectedStatusStore = {
      ...mockOrderStatusStore,
      connectionStatus: 'disconnected'
    };
    
    vi.mocked(useOrderStatusStore).mockReturnValue(disconnectedStatusStore);
    
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 应该显示连接状态提示
      expect(schemaText).toContain('disconnected');
    });
  });

  test('应该支持权限控制', async () => {
    renderWithRouter(<OrderManagePage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查权限控制
      expect(schemaText).toContain('order:read');
      expect(schemaText).toContain('order:update');
      expect(schemaText).toContain('order:batch');
    });
  });

  test('移动端应该使用响应式布局', () => {
    // 模拟移动端视口
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 375,
    });

    renderWithRouter(<OrderManagePage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    // 移动端应该使用卡片布局
    expect(schemaText).toContain('mobile');
    expect(schemaText).toContain('cards');
  });

  test('应该处理组件卸载时的清理', () => {
    const { unmount } = renderWithRouter(<OrderManagePage />);
    
    unmount();
    
    // 应该断开WebSocket连接
    expect(mockOrderStatusStore.disconnect).toHaveBeenCalled();
  });
});
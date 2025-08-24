import { render, screen, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi } from 'vitest';
import OrderDetailPage from '../../pages/customer/orders/OrderDetailPage';
import { useOrderStore } from '../../stores/orderStore';
import { Order, OrderStatus, PaymentMethod, PaymentStatus } from '../../types/order';

// Mock stores
vi.mock('../../stores/orderStore');

// Mock react-router-dom hooks
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: () => ({ id: '1' }),
    useNavigate: () => vi.fn(),
  };
});

// Mock AmisRenderer
vi.mock('../../components/ui/AmisRenderer', () => ({
  AmisRenderer: ({ schema }: { schema: any }) => (
    <div data-testid="amis-renderer">
      <div>Order Detail Page</div>
      {JSON.stringify(schema, null, 2)}
    </div>
  )
}));

const mockOrder: Order = {
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
};

const mockOrderStore = {
  orders: [mockOrder],
  currentOrder: mockOrder,
  isLoading: false,
  error: null,
  getOrder: vi.fn().mockResolvedValue(mockOrder),
  createOrder: vi.fn(),
  cancelOrder: vi.fn(),
  listOrders: vi.fn()
};

describe('OrderDetailPage', () => {
  beforeEach(() => {
    vi.mocked(useOrderStore).mockReturnValue(mockOrderStore);
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

  test('应该正确渲染订单详情页面', async () => {
    renderWithRouter(<OrderDetailPage />);
    
    await waitFor(() => {
      expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
      expect(screen.getByText('Order Detail Page')).toBeInTheDocument();
    });
  });

  test('应该在组件挂载时调用getOrder', async () => {
    renderWithRouter(<OrderDetailPage />);
    
    await waitFor(() => {
      expect(mockOrderStore.getOrder).toHaveBeenCalledWith('1');
    });
  });

  test('加载中时应该显示加载状态', () => {
    const loadingOrderStore = {
      ...mockOrderStore,
      isLoading: true,
      currentOrder: null
    };
    
    vi.mocked(useOrderStore).mockReturnValue(loadingOrderStore);
    
    renderWithRouter(<OrderDetailPage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    expect(schemaText).toContain('加载中');
  });

  test('当订单加载出错时应该显示错误信息', () => {
    const errorOrderStore = {
      ...mockOrderStore,
      error: '订单加载失败',
      currentOrder: null
    };
    
    vi.mocked(useOrderStore).mockReturnValue(errorOrderStore);
    
    renderWithRouter(<OrderDetailPage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    expect(schemaText).toContain('订单加载失败');
  });

  test('订单详情schema应该包含正确的订单信息', async () => {
    renderWithRouter(<OrderDetailPage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查基本订单信息
      expect(schemaText).toContain('订单详情');
      expect(schemaText).toContain('ORD20231201001');
      expect(schemaText).toContain('paid');
      
      // 检查商品信息
      expect(schemaText).toContain('商品清单');
      expect(schemaText).toContain('测试商品1');
      
      // 检查支付信息
      expect(schemaText).toContain('支付信息');
      expect(schemaText).toContain('alipay');
    });
  });

  test('已支付订单不应该显示取消按钮', async () => {
    renderWithRouter(<OrderDetailPage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 已支付订单不应该有取消按钮
      expect(schemaText).toContain('paid');
      // 检查条件渲染逻辑
      expect(schemaText).toContain('status === "pending"');
    });
  });

  test('待支付订单应该显示支付和取消按钮', async () => {
    const pendingOrder = {
      ...mockOrder,
      status: OrderStatus.PENDING,
      payment_info: {
        ...mockOrder.payment_info,
        payment_status: PaymentStatus.UNPAID,
        paid_at: null
      }
    };
    
    const pendingOrderStore = {
      ...mockOrderStore,
      currentOrder: pendingOrder
    };
    
    vi.mocked(useOrderStore).mockReturnValue(pendingOrderStore);
    
    renderWithRouter(<OrderDetailPage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      expect(schemaText).toContain('pending');
      expect(schemaText).toContain('立即支付');
      expect(schemaText).toContain('取消订单');
    });
  });

  test('订单不存在时应该显示404页面', () => {
    const notFoundOrderStore = {
      ...mockOrderStore,
      currentOrder: null,
      error: null,
      isLoading: false
    };
    
    vi.mocked(useOrderStore).mockReturnValue(notFoundOrderStore);
    
    renderWithRouter(<OrderDetailPage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    expect(schemaText).toContain('订单不存在');
    expect(schemaText).toContain('404');
  });

  test('schema应该包含正确的订单状态样式', async () => {
    renderWithRouter(<OrderDetailPage />);
    
    await waitFor(() => {
      const amisRenderer = screen.getByTestId('amis-renderer');
      const schemaText = amisRenderer.textContent;
      
      // 检查状态映射
      expect(schemaText).toContain('pending');
      expect(schemaText).toContain('paid');
      expect(schemaText).toContain('processing');
      expect(schemaText).toContain('completed');
      expect(schemaText).toContain('cancelled');
    });
  });
});
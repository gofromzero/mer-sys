import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { vi } from 'vitest';
import CartPage from '../../pages/customer/cart/CartPage';
import { useCartStore } from '../../stores/cartStore';
import { useOrderStore } from '../../stores/orderStore';

// Mock stores
vi.mock('../../stores/cartStore');
vi.mock('../../stores/orderStore');

// Mock AmisRenderer
vi.mock('../../components/ui/AmisRenderer', () => ({
  AmisRenderer: ({ schema }: { schema: any }) => (
    <div data-testid="amis-renderer">
      <div>Cart Page</div>
      {JSON.stringify(schema, null, 2)}
    </div>
  )
}));

const mockCartStore = {
  cart: {
    id: 1,
    items: [
      {
        id: 1,
        product_id: 1001,
        quantity: 2,
        added_at: '2023-12-01T10:00:00Z'
      },
      {
        id: 2,
        product_id: 1002,
        quantity: 1,
        added_at: '2023-12-01T11:00:00Z'
      }
    ]
  },
  isLoading: false,
  error: null,
  getCart: vi.fn(),
  updateItem: vi.fn(),
  removeItem: vi.fn(),
  clearCart: vi.fn()
};

const mockOrderStore = {
  createOrder: vi.fn()
};

describe('CartPage', () => {
  beforeEach(() => {
    vi.mocked(useCartStore).mockReturnValue(mockCartStore);
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

  test('应该正确渲染购物车页面', () => {
    renderWithRouter(<CartPage />);
    
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument();
    expect(screen.getByText('Cart Page')).toBeInTheDocument();
  });

  test('应该在组件挂载时调用getCart', () => {
    renderWithRouter(<CartPage />);
    
    expect(mockCartStore.getCart).toHaveBeenCalled();
  });

  test('当购物车加载出错时应该显示错误信息', () => {
    const errorCartStore = {
      ...mockCartStore,
      error: '加载购物车失败'
    };
    
    vi.mocked(useCartStore).mockReturnValue(errorCartStore);
    
    renderWithRouter(<CartPage />);
    
    expect(screen.getByText(/加载失败: 加载购物车失败/)).toBeInTheDocument();
    expect(screen.getByText('重试')).toBeInTheDocument();
  });

  test('点击重试按钮应该重新加载购物车', () => {
    const errorCartStore = {
      ...mockCartStore,
      error: '网络错误'
    };
    
    vi.mocked(useCartStore).mockReturnValue(errorCartStore);
    
    renderWithRouter(<CartPage />);
    
    const retryButton = screen.getByText('重试');
    fireEvent.click(retryButton);
    
    expect(mockCartStore.getCart).toHaveBeenCalledTimes(2); // 一次初始化，一次重试
  });

  test('Amis schema应该包含正确的购物车数据', () => {
    renderWithRouter(<CartPage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    // 检查schema是否包含预期的结构
    expect(schemaText).toContain('购物车');
    expect(schemaText).toContain('table');
    expect(schemaText).toContain('items');
  });

  test('购物车为空时应该禁用提交订单按钮', () => {
    const emptyCartStore = {
      ...mockCartStore,
      cart: {
        ...mockCartStore.cart,
        items: []
      }
    };
    
    vi.mocked(useCartStore).mockReturnValue(emptyCartStore);
    
    renderWithRouter(<CartPage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    // 检查按钮是否被禁用
    expect(schemaText).toContain('!items || items.length === 0');
  });

  test('schema应该包含去下单按钮配置', () => {
    renderWithRouter(<CartPage />);
    
    const amisRenderer = screen.getByTestId('amis-renderer');
    const schemaText = amisRenderer.textContent;
    
    expect(schemaText).toContain('去下单');
    expect(schemaText).toContain('/orders/create');
  });
});
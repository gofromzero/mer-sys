import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import OrderStatusHistory from '../../components/orders/OrderStatusHistory';
import { OrderStatusHistory as OrderStatusHistoryType, OrderStatusOperatorType } from '../../types/order';

// Mock orderService
vi.mock('../../services/orderService', () => ({
  default: {
    getOrderStatusHistory: vi.fn()
  }
}));

const mockStatusHistory: OrderStatusHistoryType[] = [
  {
    id: 1,
    order_id: 1,
    from_status: 1, // pending
    to_status: 2, // paid
    reason: '客户完成支付',
    operator_id: 100,
    operator_type: OrderStatusOperatorType.CUSTOMER,
    created_at: new Date('2023-12-01T12:00:00Z'),
    metadata: {
      payment_method: 'alipay',
      payment_id: 'alipay_123456'
    }
  },
  {
    id: 2,
    order_id: 1,
    from_status: 2, // paid
    to_status: 3, // processing
    reason: '商户开始处理订单',
    operator_id: 200,
    operator_type: OrderStatusOperatorType.MERCHANT,
    created_at: new Date('2023-12-01T13:00:00Z'),
    metadata: {
      merchant_note: '订单已分配给处理团队'
    }
  },
  {
    id: 3,
    order_id: 1,
    from_status: 3, // processing
    to_status: 4, // completed
    reason: '订单处理完成',
    operator_id: 200,
    operator_type: OrderStatusOperatorType.MERCHANT,
    created_at: new Date('2023-12-01T15:30:00Z'),
    metadata: {
      completion_note: '所有商品已发货'
    }
  }
];

describe('OrderStatusHistory', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  test('应该正确渲染状态历史列表', () => {
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
      />
    );

    // 检查表格模式渲染
    expect(screen.getByText('状态历史')).toBeInTheDocument();
    expect(screen.getByText('客户完成支付')).toBeInTheDocument();
    expect(screen.getByText('商户开始处理订单')).toBeInTheDocument();
    expect(screen.getByText('订单处理完成')).toBeInTheDocument();
  });

  test('应该正确显示状态标签', () => {
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
      />
    );

    // 检查状态标签
    expect(screen.getByText('待支付')).toBeInTheDocument();
    expect(screen.getByText('已支付')).toBeInTheDocument();
    expect(screen.getByText('处理中')).toBeInTheDocument();
    expect(screen.getByText('已完成')).toBeInTheDocument();
  });

  test('应该正确显示操作者类型', () => {
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
      />
    );

    // 检查操作者类型
    expect(screen.getByText('客户')).toBeInTheDocument();
    expect(screen.getByText('商户')).toBeInTheDocument();
  });

  test('应该支持时间轴模式显示', () => {
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
        displayMode="timeline"
      />
    );

    // 检查时间轴元素
    expect(screen.getByTestId('status-timeline')).toBeInTheDocument();
    
    // 时间轴中应该包含所有状态变更
    expect(screen.getByText('客户完成支付')).toBeInTheDocument();
    expect(screen.getByText('商户开始处理订单')).toBeInTheDocument();
    expect(screen.getByText('订单处理完成')).toBeInTheDocument();
  });

  test('应该支持展开查看元数据', async () => {
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
        showMetadata={true}
      />
    );

    // 查找展开按钮
    const expandButton = screen.getAllByText('详情')[0];
    fireEvent.click(expandButton);

    await waitFor(() => {
      // 检查元数据显示
      expect(screen.getByText('alipay')).toBeInTheDocument();
      expect(screen.getByText('alipay_123456')).toBeInTheDocument();
    });
  });

  test('空历史记录时应该显示空状态', () => {
    render(
      <OrderStatusHistory 
        history={[]}
        loading={false}
      />
    );

    expect(screen.getByText('暂无状态变更记录')).toBeInTheDocument();
    expect(screen.getByTestId('empty-history')).toBeInTheDocument();
  });

  test('加载中时应该显示骨架屏', () => {
    render(
      <OrderStatusHistory 
        history={[]}
        loading={true}
      />
    );

    expect(screen.getByTestId('history-skeleton')).toBeInTheDocument();
  });

  test('应该正确格式化时间显示', () => {
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
      />
    );

    // 检查时间格式
    expect(screen.getByText('2023-12-01 12:00:00')).toBeInTheDocument();
    expect(screen.getByText('2023-12-01 13:00:00')).toBeInTheDocument();
    expect(screen.getByText('2023-12-01 15:30:00')).toBeInTheDocument();
  });

  test('应该支持筛选功能', async () => {
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
        showFilters={true}
      />
    );

    // 检查筛选器
    expect(screen.getByTestId('status-filter')).toBeInTheDocument();
    expect(screen.getByTestId('operator-filter')).toBeInTheDocument();

    // 筛选操作者类型
    const operatorFilter = screen.getByTestId('operator-filter');
    fireEvent.change(operatorFilter, { target: { value: 'customer' } });

    await waitFor(() => {
      // 应该只显示客户操作的记录
      expect(screen.getByText('客户完成支付')).toBeInTheDocument();
      expect(screen.queryByText('商户开始处理订单')).not.toBeInTheDocument();
    });
  });

  test('应该支持状态图标显示', () => {
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
        showStatusIcons={true}
      />
    );

    // 检查状态图标
    expect(screen.getByTestId('status-icon-pending')).toBeInTheDocument();
    expect(screen.getByTestId('status-icon-paid')).toBeInTheDocument();
    expect(screen.getByTestId('status-icon-processing')).toBeInTheDocument();
    expect(screen.getByTestId('status-icon-completed')).toBeInTheDocument();
  });

  test('应该支持导出功能', async () => {
    const mockExport = vi.fn();
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
        showExport={true}
        onExport={mockExport}
      />
    );

    const exportButton = screen.getByText('导出历史');
    fireEvent.click(exportButton);

    await waitFor(() => {
      expect(mockExport).toHaveBeenCalledWith(mockStatusHistory);
    });
  });

  test('应该正确处理长文本截断', () => {
    const longReasonHistory = [{
      ...mockStatusHistory[0],
      reason: '这是一个非常长的状态变更原因，应该会被截断显示，以保持界面的整洁性和可读性，确保用户体验'
    }];

    render(
      <OrderStatusHistory 
        history={longReasonHistory}
        loading={false}
        maxReasonLength={20}
      />
    );

    // 检查文本截断
    expect(screen.getByText(/这是一个非常长的状态变更原因.../)).toBeInTheDocument();
    expect(screen.getByTitle(/这是一个非常长的状态变更原因/)).toBeInTheDocument();
  });

  test('应该支持刷新功能', async () => {
    const mockRefresh = vi.fn();
    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
        onRefresh={mockRefresh}
      />
    );

    const refreshButton = screen.getByTestId('refresh-history');
    fireEvent.click(refreshButton);

    await waitFor(() => {
      expect(mockRefresh).toHaveBeenCalled();
    });
  });

  test('应该在不同屏幕尺寸下响应式显示', () => {
    // 模拟移动端视口
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 375,
    });

    render(
      <OrderStatusHistory 
        history={mockStatusHistory}
        loading={false}
        responsive={true}
      />
    );

    // 在移动端应该使用卡片模式
    expect(screen.getByTestId('mobile-history-cards')).toBeInTheDocument();
  });

  test('应该处理错误状态', () => {
    const error = new Error('加载状态历史失败');
    
    render(
      <OrderStatusHistory 
        history={[]}
        loading={false}
        error={error.message}
      />
    );

    expect(screen.getByText('加载状态历史失败')).toBeInTheDocument();
    expect(screen.getByTestId('error-state')).toBeInTheDocument();
  });
});
// 商品列表页面组件测试
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import ProductListPage from '../../pages/merchant/products/ProductListPage';

// Mock AmisRenderer
jest.mock('../../components/ui/AmisRenderer', () => {
  return function MockAmisRenderer({ schema }: { schema: any }) {
    return (
      <div data-testid="amis-renderer">
        <div data-testid="page-title">{schema.title}</div>
        <div data-testid="page-type">{schema.type}</div>
      </div>
    );
  };
});

const renderWithRouter = (component: React.ReactElement) => {
  return render(
    <BrowserRouter>
      {component}
    </BrowserRouter>
  );
};

describe('ProductListPage', () => {
  it('应该正确渲染商品管理页面', () => {
    renderWithRouter(<ProductListPage />);
    
    const renderer = screen.getByTestId('amis-renderer');
    const title = screen.getByTestId('page-title');
    const type = screen.getByTestId('page-type');
    
    expect(renderer).toBeDefined();
    expect(title.textContent).toBe('商品管理');
    expect(type.textContent).toBe('page');
  });

  it('应该包含正确的页面配置', () => {
    renderWithRouter(<ProductListPage />);
    
    const renderer = screen.getByTestId('amis-renderer');
    expect(renderer).toBeDefined();
  });

  it('组件应该成功挂载和卸载', () => {
    const { unmount } = renderWithRouter(<ProductListPage />);
    
    const renderer = screen.getByTestId('amis-renderer');
    expect(renderer).toBeDefined();
    
    unmount();
    
    const unmountedRenderer = screen.queryByTestId('amis-renderer');
    expect(unmountedRenderer).toBeNull();
  });
});
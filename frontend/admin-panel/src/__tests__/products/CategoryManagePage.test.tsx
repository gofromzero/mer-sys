// 商品分类管理页面组件测试
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import CategoryManagePage from '../../pages/merchant/products/CategoryManagePage';

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

describe('CategoryManagePage', () => {
  it('应该正确渲染分类管理页面', () => {
    renderWithRouter(<CategoryManagePage />);
    
    const renderer = screen.getByTestId('amis-renderer');
    const title = screen.getByTestId('page-title');
    const type = screen.getByTestId('page-type');
    
    expect(renderer).toBeDefined();
    expect(title.textContent).toBe('分类管理');
    expect(type.textContent).toBe('page');
  });

  it('应该包含正确的页面配置', () => {
    renderWithRouter(<CategoryManagePage />);
    
    const renderer = screen.getByTestId('amis-renderer');
    expect(renderer).toBeDefined();
  });

  it('组件应该成功挂载和卸载', () => {
    const { unmount } = renderWithRouter(<CategoryManagePage />);
    
    const renderer = screen.getByTestId('amis-renderer');
    expect(renderer).toBeDefined();
    
    unmount();
    
    const unmountedRenderer = screen.queryByTestId('amis-renderer');
    expect(unmountedRenderer).toBeNull();
  });
});
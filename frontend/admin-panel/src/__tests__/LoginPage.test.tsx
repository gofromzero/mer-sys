import * as React from 'react';
import { render } from '@testing-library/react';
import { screen, fireEvent, waitFor } from '@testing-library/dom';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { LoginPage } from '../pages/auth/LoginPage';
import { useAuth } from '../hooks/useAuth';
import '@testing-library/jest-dom';

// Mock useAuth hook
jest.mock('../hooks/useAuth');
const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

// Mock react-router-dom
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
  useLocation: () => ({ state: null }),
}));

// 测试组件包装器
const TestWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return <BrowserRouter>{children}</BrowserRouter>;
};

describe('LoginPage', () => {
  const mockLogin = jest.fn();
  const mockClearError = jest.fn();
  
  const defaultAuthState = {
    login: mockLogin,
    isLoading: false,
    error: null,
    clearError: mockClearError,
    isAuthenticated: false,
  };

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue(defaultAuthState as any);
  });

  describe('渲染测试', () => {
    it('应该正确渲染登录表单', () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      expect(screen.getByText('登录到管理后台')).toBeInTheDocument();
      expect(screen.getByText('Mer Demo 多租户管理系统')).toBeInTheDocument();
      expect(screen.getByLabelText('用户名')).toBeInTheDocument();
      expect(screen.getByLabelText('密码')).toBeInTheDocument();
      expect(screen.getByLabelText('记住我')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '登录' })).toBeInTheDocument();
    });

    it('应该显示占位符文本', () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      expect(screen.getByPlaceholderText('请输入用户名')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('请输入密码')).toBeInTheDocument();
    });

    it('在开发环境下应该显示开发提示', () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'development';

      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      expect(screen.getByText(/开发环境提示/)).toBeInTheDocument();
      expect(screen.getByText(/请确保后端服务运行在 localhost:8080/)).toBeInTheDocument();

      process.env.NODE_ENV = originalEnv;
    });
  });

  describe('表单验证测试', () => {
    it('应该在用户名为空时显示错误信息', async () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const submitButton = screen.getByRole('button', { name: '登录' });
      const passwordInput = screen.getByLabelText('密码');
      
      // 只填写密码，不填用户名
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('请输入用户名')).toBeInTheDocument();
      });
    });

    it('应该在密码为空时显示错误信息', async () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const submitButton = screen.getByRole('button', { name: '登录' });
      const usernameInput = screen.getByLabelText('用户名');
      
      // 只填写用户名，不填密码
      fireEvent.change(usernameInput, { target: { value: 'testuser' } });
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('请输入密码')).toBeInTheDocument();
      });
    });

    it('应该在用户名长度不足时显示错误信息', async () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const submitButton = screen.getByRole('button', { name: '登录' });
      const usernameInput = screen.getByLabelText('用户名');
      const passwordInput = screen.getByLabelText('密码');
      
      fireEvent.change(usernameInput, { target: { value: 'ab' } }); // 少于3个字符
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('用户名至少需要3个字符')).toBeInTheDocument();
      });
    });

    it('应该在密码长度不足时显示错误信息', async () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const submitButton = screen.getByRole('button', { name: '登录' });
      const usernameInput = screen.getByLabelText('用户名');
      const passwordInput = screen.getByLabelText('密码');
      
      fireEvent.change(usernameInput, { target: { value: 'testuser' } });
      fireEvent.change(passwordInput, { target: { value: '12345' } }); // 少于6个字符
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('密码至少需要6个字符')).toBeInTheDocument();
      });
    });

    it('应该在输入有效内容后清除错误信息', async () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const submitButton = screen.getByRole('button', { name: '登录' });
      const usernameInput = screen.getByLabelText('用户名');
      
      // 先触发错误
      fireEvent.click(submitButton);
      await waitFor(() => {
        expect(screen.getByText('请输入用户名')).toBeInTheDocument();
      });

      // 输入有效用户名
      fireEvent.change(usernameInput, { target: { value: 'testuser' } });
      
      // 错误信息应该消失
      await waitFor(() => {
        expect(screen.queryByText('请输入用户名')).not.toBeInTheDocument();
      });
    });
  });

  describe('登录功能测试', () => {
    it('应该在表单有效时调用登录函数', async () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const usernameInput = screen.getByLabelText('用户名');
      const passwordInput = screen.getByLabelText('密码');
      const rememberMeCheckbox = screen.getByLabelText('记住我');
      const submitButton = screen.getByRole('button', { name: '登录' });

      // 填写表单
      fireEvent.change(usernameInput, { target: { value: 'testuser' } });
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      fireEvent.click(rememberMeCheckbox);
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith({
          username: 'testuser',
          password: 'password123',
          remember_me: true,
        });
      });
    });

    it('应该在登录成功后导航到仪表板', async () => {
      // 模拟登录成功
      mockLogin.mockResolvedValueOnce(undefined);

      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const usernameInput = screen.getByLabelText('用户名');
      const passwordInput = screen.getByLabelText('密码');
      const submitButton = screen.getByRole('button', { name: '登录' });

      fireEvent.change(usernameInput, { target: { value: 'testuser' } });
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true });
      });
    });

    it('应该在登录失败时不导航', async () => {
      // 模拟登录失败
      mockLogin.mockRejectedValueOnce(new Error('登录失败'));

      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const usernameInput = screen.getByLabelText('用户名');
      const passwordInput = screen.getByLabelText('密码');
      const submitButton = screen.getByRole('button', { name: '登录' });

      fireEvent.change(usernameInput, { target: { value: 'testuser' } });
      fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
      fireEvent.click(submitButton);

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalled();
      });

      // 不应该导航
      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });

  describe('加载状态测试', () => {
    it('应该在加载时显示加载状态', () => {
      mockUseAuth.mockReturnValue({
        ...defaultAuthState,
        isLoading: true,
      } as any);

      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      expect(screen.getByText('登录中...')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '登录中...' })).toBeDisabled();
    });

    it('应该在加载时禁用表单输入', () => {
      mockUseAuth.mockReturnValue({
        ...defaultAuthState,
        isLoading: true,
      } as any);

      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      expect(screen.getByLabelText('用户名')).toBeDisabled();
      expect(screen.getByLabelText('密码')).toBeDisabled();
      expect(screen.getByLabelText('记住我')).toBeDisabled();
    });
  });

  describe('错误处理测试', () => {
    it('应该显示错误信息', () => {
      const errorMessage = '用户名或密码错误';
      mockUseAuth.mockReturnValue({
        ...defaultAuthState,
        error: errorMessage,
      } as any);

      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      expect(screen.getByText('登录失败')).toBeInTheDocument();
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });

    it('应该能够关闭错误信息', () => {
      const errorMessage = '用户名或密码错误';
      mockUseAuth.mockReturnValue({
        ...defaultAuthState,
        error: errorMessage,
      } as any);

      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const closeButton = screen.getByRole('button', { name: '关闭' });
      fireEvent.click(closeButton);

      expect(mockClearError).toHaveBeenCalled();
    });
  });

  describe('已认证用户重定向测试', () => {
    it('应该在用户已登录时重定向到仪表板', () => {
      mockUseAuth.mockReturnValue({
        ...defaultAuthState,
        isAuthenticated: true,
      } as any);

      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true });
    });
  });

  describe('记住我功能测试', () => {
    it('应该正确处理记住我选项', () => {
      render(
        <TestWrapper>
          <LoginPage />
        </TestWrapper>
      );

      const rememberMeCheckbox = screen.getByLabelText('记住我') as HTMLInputElement;
      
      // 初始状态应该是未选中
      expect(rememberMeCheckbox.checked).toBe(false);
      
      // 点击后应该选中
      fireEvent.click(rememberMeCheckbox);
      expect(rememberMeCheckbox.checked).toBe(true);
      
      // 再次点击应该取消选中
      fireEvent.click(rememberMeCheckbox);
      expect(rememberMeCheckbox.checked).toBe(false);
    });
  });
});
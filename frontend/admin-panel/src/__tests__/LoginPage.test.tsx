// React import not needed with new JSX transform
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import '@testing-library/jest-dom';
import { LoginPage } from '../pages/auth/LoginPage';

// Mock the useAuth hook
const mockLogin = jest.fn();
const mockClearError = jest.fn();
const mockNavigate = jest.fn();

jest.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    login: mockLogin,
    isLoading: false,
    error: null,
    clearError: mockClearError,
    isAuthenticated: false,
  }),
}));

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
  useLocation: () => ({ state: null }),
}));

const renderLoginPage = () => {
  return render(
    <BrowserRouter>
      <LoginPage />
    </BrowserRouter>
  );
};

describe('LoginPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders login form correctly', () => {
    renderLoginPage();
    
    expect(screen.getByText('登录到管理后台')).toBeInTheDocument();
    expect(screen.getByText('Mer Demo 多租户管理系统')).toBeInTheDocument();
    expect(screen.getByLabelText('用户名')).toBeInTheDocument();
    expect(screen.getByLabelText('密码')).toBeInTheDocument();
    expect(screen.getByLabelText('记住我')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '登录' })).toBeInTheDocument();
  });

  test('shows validation errors for empty fields', async () => {
    renderLoginPage();
    
    const submitButton = screen.getByRole('button', { name: '登录' });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('请输入用户名')).toBeInTheDocument();
      expect(screen.getByText('请输入密码')).toBeInTheDocument();
    });
  });

  test('shows validation errors for short inputs', async () => {
    renderLoginPage();
    
    const usernameInput = screen.getByLabelText('用户名');
    const passwordInput = screen.getByLabelText('密码');
    const submitButton = screen.getByRole('button', { name: '登录' });

    fireEvent.change(usernameInput, { target: { value: 'ab' } });
    fireEvent.change(passwordInput, { target: { value: '123' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('用户名至少需要3个字符')).toBeInTheDocument();
      expect(screen.getByText('密码至少需要6个字符')).toBeInTheDocument();
    });
  });

  test('calls login function with correct credentials', async () => {
    renderLoginPage();
    
    const usernameInput = screen.getByLabelText('用户名');
    const passwordInput = screen.getByLabelText('密码');
    const rememberMeCheckbox = screen.getByLabelText('记住我');
    const submitButton = screen.getByRole('button', { name: '登录' });

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

  test('clears field errors when user starts typing', async () => {
    renderLoginPage();
    
    const usernameInput = screen.getByLabelText('用户名');
    const submitButton = screen.getByRole('button', { name: '登录' });

    // First trigger validation error
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('请输入用户名')).toBeInTheDocument();
    });

    // Then start typing to clear the error
    fireEvent.change(usernameInput, { target: { value: 'test' } });
    
    await waitFor(() => {
      expect(screen.queryByText('请输入用户名')).not.toBeInTheDocument();
    });
  });

  test('disables submit button when form is invalid', () => {
    renderLoginPage();
    
    const submitButton = screen.getByRole('button', { name: '登录' });
    expect(submitButton).toBeDisabled();
  });

  test('enables submit button when form is valid', async () => {
    renderLoginPage();
    
    const usernameInput = screen.getByLabelText('用户名');
    const passwordInput = screen.getByLabelText('密码');
    const submitButton = screen.getByRole('button', { name: '登录' });

    fireEvent.change(usernameInput, { target: { value: 'testuser' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });

    await waitFor(() => {
      expect(submitButton).not.toBeDisabled();
    });
  });
});
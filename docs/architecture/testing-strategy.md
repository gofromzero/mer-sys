# Testing Strategy

## Testing Pyramid

```
            E2E Tests (10%)
           /              \
      Integration Tests (20%)
     /                        \
Frontend Unit (35%)    Backend Unit (35%)
```

## Test Organization

### Frontend Tests

```
src/
├── __tests__/              # 测试文件
│   ├── components/         # 组件测试
│   ├── hooks/              # Hook测试
│   ├── services/           # 服务测试
│   └── utils/              # 工具函数测试
├── __mocks__/              # Mock文件
└── test-utils/             # 测试工具
```

### Backend Tests

```
services/user-service/
├── test/
│   ├── unit/               # 单元测试
│   ├── integration/        # 集成测试
│   └── fixtures/           # 测试数据
└── internal/
    └── *_test.go           # 测试文件
```

### E2E Tests

```
e2e/
├── specs/                  # 测试用例
├── fixtures/               # 测试数据
├── pages/                  # 页面对象
└── utils/                  # 测试工具
```

## Test Examples

### Frontend Component Test

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { UserCard } from '@/components/UserCard';

describe('UserCard', () => {
  const mockUser = {
    id: '1',
    username: 'test-user',
    email: 'test@example.com'
  };

  it('should render user information', () => {
    render(<UserCard user={mockUser} />);
    
    expect(screen.getByText('test-user')).toBeInTheDocument();
    expect(screen.getByText('test@example.com')).toBeInTheDocument();
  });

  it('should call onEdit when edit button clicked', () => {
    const onEdit = jest.fn();
    render(<UserCard user={mockUser} onEdit={onEdit} />);
    
    fireEvent.click(screen.getByText('编辑'));
    expect(onEdit).toHaveBeenCalledWith(mockUser);
  });
});
```

### Backend API Test

```go
func TestUserController_Create(t *testing.T) {
    // 设置测试环境
    app := gtest.NewApp()
    defer app.Stop()
    
    // 创建测试请求
    req := &model.CreateUserRequest{
        Username: "test-user",
        Email:    "test@example.com",
        Password: "password123",
        TenantID: 1,
    }
    
    // 发送请求
    resp := app.POST("/api/v1/users").JSON(req).Exec()
    
    // 验证响应
    resp.AssertStatus(201)
    resp.AssertJsonPath("$.data.username", "test-user")
}
```

### E2E Test

```typescript
import { test, expect } from '@playwright/test';

test('user login flow', async ({ page }) => {
  // 导航到登录页面
  await page.goto('/auth/login');
  
  // 填写登录表单
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'password');
  
  // 点击登录按钮
  await page.click('button[type="submit"]');
  
  // 验证登录成功
  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('text=欢迎回来')).toBeVisible();
});
```

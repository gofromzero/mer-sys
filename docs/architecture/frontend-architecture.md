# Frontend Architecture

## Component Architecture

### Component Organization

```
src/
├── components/           # 通用组件
│   ├── ui/              # 基础UI组件
│   ├── forms/           # 表单组件
│   └── layouts/         # 布局组件
├── pages/               # 页面组件
│   ├── tenant/          # 租户管理页面
│   ├── merchant/        # 商户管理页面
│   ├── customer/        # 客户页面
│   └── auth/            # 认证页面
├── hooks/               # 自定义Hooks
├── services/            # API服务层
├── stores/              # 状态管理
├── utils/               # 工具函数
└── types/               # TypeScript类型定义
```

### Component Template

```typescript
import React from 'react';
import { Card } from '@/components/ui';
import { useAuth } from '@/hooks';
import type { User } from '@/types';

interface UserCardProps {
  user: User;
  onEdit?: (user: User) => void;
}

export const UserCard: React.FC<UserCardProps> = ({ user, onEdit }) => {
  const { hasPermission } = useAuth();

  return (
    <Card>
      <div className="p-4">
        <h3>{user.username}</h3>
        <p>{user.email}</p>
        {hasPermission('user:edit') && (
          <button onClick={() => onEdit?.(user)}>编辑</button>
        )}
      </div>
    </Card>
  );
};
```

## State Management Architecture

### State Structure

```typescript
interface AppState {
  auth: AuthState;
  tenant: TenantState;
  merchants: MerchantState;
  products: ProductState;
  orders: OrderState;
  ui: UIState;
}

interface AuthState {
  user: User | null;
  token: string | null;
  permissions: string[];
  isLoading: boolean;
}

interface MerchantState {
  current: Merchant | null;
  list: Merchant[];
  isLoading: boolean;
  filters: MerchantFilters;
}
```

### State Management Patterns

- 使用Zustand进行全局状态管理
- 按功能模块分离store
- 使用immer进行不可变更新
- 实现乐观更新提升用户体验

## Routing Architecture

### Route Organization

```
/
├── /auth/login          # 登录页面
├── /auth/register       # 注册页面
├── /dashboard           # 仪表板
├── /tenant/             # 租户管理
│   ├── /settings        # 租户设置
│   └── /merchants       # 商户管理
├── /merchant/           # 商户门户
│   ├── /products        # 商品管理
│   ├── /orders          # 订单管理
│   └── /reports         # 报表分析
└── /customer/           # 客户应用
    ├── /products        # 商品浏览
    ├── /orders          # 订单中心
    └── /profile         # 个人中心
```

### Protected Route Pattern

```typescript
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/hooks';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: string;
  requiredPermission?: string;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requiredRole,
  requiredPermission
}) => {
  const { user, hasRole, hasPermission } = useAuth();
  const location = useLocation();

  if (!user) {
    return <Navigate to="/auth/login" state={{ from: location }} />;
  }

  if (requiredRole && !hasRole(requiredRole)) {
    return <Navigate to="/unauthorized" />;
  }

  if (requiredPermission && !hasPermission(requiredPermission)) {
    return <Navigate to="/forbidden" />;
  }

  return <>{children}</>;
};
```

## Frontend Services Layer

### API Client Setup

```typescript
import axios from 'axios';
import { useAuthStore } from '@/stores';

const apiClient = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
});

// 请求拦截器
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout();
    }
    return Promise.reject(error);
  }
);
```

### Service Example

```typescript
import { apiClient } from '@/lib/api';
import type { User, CreateUserRequest } from '@/types';

export const userService = {
  async getUsers(): Promise<User[]> {
    const response = await apiClient.get('/users');
    return response.data;
  },

  async createUser(user: CreateUserRequest): Promise<User> {
    const response = await apiClient.post('/users', user);
    return response.data;
  },

  async updateUser(id: string, user: Partial<User>): Promise<User> {
    const response = await apiClient.put(`/users/${id}`, user);
    return response.data;
  },

  async deleteUser(id: string): Promise<void> {
    await apiClient.delete(`/users/${id}`);
  },
};
```

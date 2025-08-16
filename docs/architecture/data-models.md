# Data Models

## User (用户)

**Purpose:** 统一的用户实体，支持三层B2B2C架构中的所有用户角色

**Key Attributes:**
- id: string - 全局唯一用户标识
- username: string - 用户名，租户内唯一
- email: string - 邮箱地址，全局唯一
- tenant_id: string - 所属租户ID
- roles: UserRole[] - 用户角色列表

### TypeScript Interface

```typescript
interface User {
  id: string;
  uuid: string;
  username: string;
  email: string;
  phone?: string;
  status: UserStatus;
  tenant_id: string;
  roles: UserRole[];
  profile?: UserProfile;
  created_at: Date;
  updated_at: Date;
  last_login_at?: Date;
}

enum UserStatus {
  PENDING = 'pending',
  ACTIVE = 'active', 
  SUSPENDED = 'suspended',
  DEACTIVATED = 'deactivated'
}

interface UserRole {
  id: string;
  user_id: string;
  role_type: RoleType;
  resource_id?: string;
  permissions: string[];
}
```

### Relationships

- User belongs to Tenant (多对一)
- User has many UserRoles (一对多)
- User can be associated with Merchant through UserRole (多对多)

## Tenant (租户)

**Purpose:** 多租户架构的核心实体，提供数据隔离和配置管理

### TypeScript Interface

```typescript
interface Tenant {
  id: string;
  name: string;
  code: string;
  status: TenantStatus;
  config: TenantConfig;
  created_at: Date;
  updated_at: Date;
}

enum TenantStatus {
  ACTIVE = 'active',
  SUSPENDED = 'suspended',
  EXPIRED = 'expired'
}
```

### Relationships

- Tenant has many Users (一对多)
- Tenant has many Merchants (一对多)

## Merchant (商户)

**Purpose:** 商户实体，B2B2C架构中的运营层核心

### TypeScript Interface

```typescript
interface Merchant {
  id: string;
  tenant_id: string;
  name: string;
  code: string;
  status: MerchantStatus;
  business_info: BusinessInfo;
  rights_balance: RightsBalance;
  created_at: Date;
  updated_at: Date;
}

interface RightsBalance {
  total_balance: number;
  used_balance: number;
  frozen_balance: number;
}
```

### Relationships

- Merchant belongs to Tenant (多对一)
- Merchant has many Products (一对多)
- Merchant has many Orders (一对多)

## Product (商品)

**Purpose:** 商品实体，支持商户的商品目录管理和客户浏览购买

### TypeScript Interface

```typescript
interface Product {
  id: string;
  tenant_id: string;
  merchant_id: string;
  name: string;
  description?: string;
  price: Money;
  rights_cost: number;
  inventory: InventoryInfo;
  status: ProductStatus;
  created_at: Date;
  updated_at: Date;
}

interface Money {
  amount: number;
  currency: string;
}

interface InventoryInfo {
  stock_quantity: number;
  reserved_quantity: number;
  track_inventory: boolean;
}
```

### Relationships

- Product belongs to Tenant (多对一)
- Product belongs to Merchant (多对一)

## Order (订单)

**Purpose:** 订单实体，完整的交易生命周期管理和核销流程

### TypeScript Interface

```typescript
interface Order {
  id: string;
  tenant_id: string;
  merchant_id: string;
  customer_id: string;
  order_number: string;
  status: OrderStatus;
  items: OrderItem[];
  payment_info: PaymentInfo;
  verification?: VerificationInfo;
  total_amount: Money;
  total_rights_cost: number;
  created_at: Date;
  updated_at: Date;
}

enum OrderStatus {
  PENDING = 'pending',
  PAID = 'paid', 
  PROCESSING = 'processing',
  COMPLETED = 'completed',
  CANCELLED = 'cancelled'
}

interface VerificationInfo {
  verification_code: string;
  qr_code_url: string;
  verified_at?: Date;
  verified_by?: string;
}
```

### Relationships

- Order belongs to Tenant, Merchant, Customer (多对一)
- Order has many OrderItems (一对多)

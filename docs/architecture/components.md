# Components

## API Gateway

**Responsibility:** 统一API入口，处理认证、路由、限流、监控

**Key Interfaces:**
- HTTP请求路由和转发
- JWT token验证
- 请求限流和熔断
- 访问日志和监控

**Dependencies:** 所有后端微服务

**Technology Stack:** Nginx + Lua脚本或Kong API Gateway

## User Service

**Responsibility:** 统一用户管理，支持三层B2B2C用户体系

**Key Interfaces:**
- 用户注册、登录、注销API
- 用户信息CRUD操作
- 角色权限管理
- JWT token生成和验证

**Dependencies:** MySQL, Redis, 邮件服务

**Technology Stack:** Go + GoFrame + GORM

## Tenant Service

**Responsibility:** 多租户管理和配置

**Key Interfaces:**
- 租户注册和配置API
- 租户状态管理
- 多租户隔离策略

**Dependencies:** MySQL, User Service

**Technology Stack:** Go + GoFrame

## Merchant Service

**Responsibility:** 商户生命周期管理

**Key Interfaces:**
- 商户注册审批API
- 商户信息管理
- 商户状态控制

**Dependencies:** MySQL, Tenant Service, User Service

**Technology Stack:** Go + GoFrame

## Product Service

**Responsibility:** 商品目录和库存管理

**Key Interfaces:**
- 商品CRUD操作API
- 库存管理API
- 商品搜索和分类

**Dependencies:** MySQL, OSS, Merchant Service

**Technology Stack:** Go + GoFrame + ElasticSearch

## Order Service

**Responsibility:** 订单处理和核销管理

**Key Interfaces:**
- 订单创建和管理API
- 支付集成接口
- 核销码生成和验证

**Dependencies:** MySQL, Product Service, Fund Service, 支付服务

**Technology Stack:** Go + GoFrame + 第三方支付SDK

## Fund & Rights Service

**Responsibility:** 资金和权益管理

**Key Interfaces:**
- 充值和分配API
- 权益使用和监控
- 财务报表生成

**Dependencies:** MySQL, Merchant Service

**Technology Stack:** Go + GoFrame

## Report Service

**Responsibility:** 数据分析和报表生成

**Key Interfaces:**
- 数据聚合API
- 报表生成和导出
- 实时数据查询

**Dependencies:** MySQL (只读), Redis

**Technology Stack:** Go + GoFrame + ClickHouse

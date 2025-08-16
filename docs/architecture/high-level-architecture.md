# High Level Architecture

## Technical Summary

多租户商户管理SaaS系统采用微服务架构，使用Go+GoFrame构建领域驱动的后端服务，React+Amis构建多层级管理前端。系统通过RESTful API实现前后端分离，使用MySQL存储业务数据，Redis提供缓存和会话管理。基于Docker容器化部署到云平台，支持水平扩展。该架构实现了三层B2B2C用户体系的完整业务流程，从租户管理到商户运营再到客户服务，确保数据隔离、权限控制和高性能访问。

## Platform and Infrastructure Choice

**Platform**: 阿里云  
**Key Services**: ECS集群、RDS MySQL、Redis、OSS对象存储、SLB负载均衡、CDN加速  
**Deployment Host and Regions**: 华东2（上海）主区域，华北2（北京）灾备区域

## Repository Structure

**Structure**: Monorepo  
**Monorepo Tool**: Go workspace (后端) + npm workspaces (前端)  
**Package Organization**: 按领域服务分离，共享代码独立包管理

## High Level Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        WEB[Web Browser]
        MOBILE[Mobile Browser]
    end
    
    subgraph "CDN & Edge"
        CDN[Alibaba Cloud CDN]
    end
    
    subgraph "Frontend Layer"
        LB1[SLB Load Balancer]
        ADMIN[Admin Panel<br/>租户管理后台]
        MERCHANT[Merchant Portal<br/>商户运营门户]
        CUSTOMER[Customer App<br/>客户服务应用]
    end
    
    subgraph "API Gateway"
        GATEWAY[API Gateway<br/>统一入口]
    end
    
    subgraph "Backend Services"
        LB2[SLB Load Balancer]
        
        subgraph "Core Services"
            USER_SVC[User Service<br/>用户服务]
            TENANT_SVC[Tenant Service<br/>租户服务]
            MERCHANT_SVC[Merchant Service<br/>商户服务]
            PRODUCT_SVC[Product Service<br/>商品服务]
            ORDER_SVC[Order Service<br/>订单服务]
            FUND_SVC[Fund Service<br/>资金权益服务]
            REPORT_SVC[Report Service<br/>报表服务]
        end
    end
    
    subgraph "Data Layer"
        MYSQL[(MySQL 8.0<br/>主数据库)]
        REDIS[(Redis<br/>缓存/会话)]
        OSS[OSS对象存储<br/>文件存储]
    end
    
    subgraph "External Services"
        SMS[短信服务]
        EMAIL[邮件服务]
        PAYMENT[支付接口]
    end
    
    WEB --> CDN
    MOBILE --> CDN
    CDN --> LB1
    
    LB1 --> ADMIN
    LB1 --> MERCHANT  
    LB1 --> CUSTOMER
    
    ADMIN --> GATEWAY
    MERCHANT --> GATEWAY
    CUSTOMER --> GATEWAY
    
    GATEWAY --> LB2
    LB2 --> USER_SVC
    LB2 --> TENANT_SVC
    LB2 --> MERCHANT_SVC
    LB2 --> PRODUCT_SVC
    LB2 --> ORDER_SVC
    LB2 --> FUND_SVC
    LB2 --> REPORT_SVC
    
    USER_SVC --> MYSQL
    USER_SVC --> REDIS
    TENANT_SVC --> MYSQL
    MERCHANT_SVC --> MYSQL
    PRODUCT_SVC --> MYSQL
    PRODUCT_SVC --> OSS
    ORDER_SVC --> MYSQL
    FUND_SVC --> MYSQL
    REPORT_SVC --> MYSQL
    
    ORDER_SVC --> SMS
    USER_SVC --> EMAIL
    ORDER_SVC --> PAYMENT
```

## Architectural Patterns

- **DDD (Domain-Driven Design)**: 按业务领域组织代码结构，清晰的聚合根和实体边界 - _理由:_ 复杂业务逻辑需要清晰的领域建模
- **微服务架构**: 按业务能力拆分独立服务，支持独立部署和扩展 - _理由:_ 三层用户架构需要不同的扩展策略和开发节奏
- **CQRS模式**: 读写分离，优化查询性能 - _理由:_ 报表查询和业务操作有不同的性能要求
- **事件驱动架构**: 服务间通过领域事件松耦合通信 - _理由:_ 减少服务间直接依赖，提高系统弹性
- **API网关模式**: 统一API入口，处理认证、限流和路由 - _理由:_ 简化前端调用，统一安全策略
- **Repository模式**: 抽象数据访问层，支持测试和数据源切换 - _理由:_ 提高代码可测试性和可维护性
- **多租户单库模式**: 通过tenant_id实现数据隔离 - _理由:_ 平衡数据隔离和运维复杂度

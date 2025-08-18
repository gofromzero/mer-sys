# MER System - 多租户商户管理平台

<div align="center">

![MER System](https://img.shields.io/badge/MER%20System-v1.0.0-blue?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.25.0-00ADD8?style=for-the-badge&logo=go)
![React](https://img.shields.io/badge/React-19.1.1-61DAFB?style=for-the-badge&logo=react)
![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=for-the-badge&logo=mysql)
![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=for-the-badge&logo=redis)

**基于多租户架构的企业级商户资金权益管理SaaS平台**

</div>

## 📋 目录

- [项目概述](#-项目概述)
- [核心特性](#-核心特性)
- [技术架构](#-技术架构)
- [快速开始](#-快速开始)
- [项目结构](#-项目结构)
- [开发指南](#-开发指南)
- [API 文档](#-api-文档)
- [部署指南](#-部署指南)
- [测试](#-测试)
- [贡献指南](#-贡献指南)
- [许可证](#-许可证)

## 🚀 项目概述

MER System 是一个专为多租户场景设计的商户管理SaaS平台，致力于解决多商户环境下充值资金管理和服务权益分配的复杂性问题。该系统采用领域驱动设计(DDD)架构，提供统一的资金流转管理、透明的权益追踪机制和完整的多租户数据隔离。

### 核心价值

- 🏢 **多租户隔离**: 基于 `tenant_id` 的严格数据隔离架构
- 💰 **资金管理**: 统一管理多商户充值资金和权益分配
- 🔒 **安全可靠**: JWT认证 + RBAC权限控制 + 数据加密
- 📊 **智能分析**: 实时监控和多维度财务分析报告
- 🎯 **高可扩展**: 微服务架构支持弹性扩展

## ✨ 核心特性

### 🏢 多租户管理
- **租户级数据隔离**: 每个租户的数据完全隔离，确保安全性
- **灵活权限配置**: 支持租户级别的个性化配置和权限管理
- **弹性扩展**: 支持租户数量的无限扩展

### 💼 商户运营
- **商户注册与管理**: 完整的商户生命周期管理
- **商品管理**: 商品上架、定价、库存和权益消耗规则配置
- **订单处理**: 端到端的订单处理和核销流程
- **权益监控**: 实时权益余额查询和使用统计

### 💰 资金权益
- **充值管理**: 集中化的充值资金管理和分配
- **权益池管理**: 灵活的权益定义、分配和使用策略
- **资金流转追踪**: 完整的资金流转记录和审计追踪
- **智能预警**: 自动化的余额预警和权益到期提醒

### 📊 数据分析
- **实时报表**: 租户、商户、客户三层次的数据报表
- **财务分析**: 多维度的财务数据分析和趋势预测
- **运营洞察**: 商户经营状况和客户行为分析

## 🏗️ 技术架构

### 技术栈

**后端技术栈**
- **语言**: Go 1.25.0
- **框架**: GoFrame v2.9.0 (企业级Web框架)
- **架构**: DDD (领域驱动设计) + 微服务
- **数据库**: MySQL 8.0 (主数据库) + Redis 7.0 (缓存)
- **认证**: JWT + RBAC
- **工作空间**: Go Workspace 管理多服务

**前端技术栈**
- **框架**: React 19.1.1 + TypeScript 5.8.3
- **UI框架**: Amis 6.13.0 (低代码可视化框架)
- **状态管理**: Zustand 4.5.7
- **样式**: Tailwind CSS 3.4.17
- **构建工具**: Vite 7.1.2

**基础设施**
- **容器化**: Docker + Docker Compose
- **网络**: 自定义网络隔离
- **数据卷**: 持久化数据存储
- **健康检查**: 多层次服务健康监控

### 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                        前端层                                │
├─────────────────────────────────────────────────────────────┤
│  Admin Panel (React + Amis)  │  Tenant Portal (Future)     │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      API网关层                              │
├─────────────────────────────────────────────────────────────┤
│  Gateway Service (认证、路由、限流、监控)                    │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      微服务层                               │
├─────────────────────────────────────────────────────────────┤
│ User Service │ Tenant Service │ Merchant Service │ ...      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      共享层                                 │
├─────────────────────────────────────────────────────────────┤
│  Repository │ Auth │ Cache │ Config │ Types │ Utils         │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      数据层                                 │
├─────────────────────────────────────────────────────────────┤
│          MySQL 8.0          │        Redis 7.0             │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 环境要求

- **Go**: 1.25.0+
- **Node.js**: 18.0+
- **Docker**: 20.0+
- **Docker Compose**: 2.0+
- **Git**: 2.30+

### 一键启动

```bash
# 1. 克隆项目
git clone <repository-url>
cd mer-sys

# 2. 启动完整开发环境
./scripts/docker-dev.sh start

# 3. 查看服务状态
./scripts/docker-dev.sh status

# 4. 访问服务
# 前端管理后台: http://localhost:5173
# API网关: http://localhost:8080
# 用户服务: http://localhost:8081
# 租户服务: http://localhost:8082
```

### 环境配置

复制并配置环境变量：

```bash
# 复制环境配置文件
cp .env.example .env

# 根据需要修改配置
vim .env
```

主要配置项：
- 数据库连接信息
- Redis 连接信息
- JWT 密钥配置
- 服务端口配置

## 📁 项目结构

```
mer-sys/
├── backend/                     # Go workspace 根目录
│   ├── go.work                 # Go workspace 配置
│   ├── shared/                 # 共享包（核心业务逻辑）
│   │   ├── repository/         # 数据访问层（多租户隔离）
│   │   ├── auth/              # JWT 认证和授权
│   │   ├── cache/             # Redis 缓存封装
│   │   ├── config/            # 数据库和 Redis 配置
│   │   ├── health/            # 健康检查系统
│   │   ├── middleware/        # HTTP 中间件
│   │   ├── types/             # 共享数据类型
│   │   └── database/          # 数据库迁移文件
│   ├── gateway/               # API 网关服务
│   └── services/              # 微服务集合
│       ├── user-service/      # 用户管理服务
│       ├── tenant-service/    # 租户管理服务
│       ├── merchant-service/  # 商户管理服务（规划中）
│       ├── product-service/   # 商品管理服务（规划中）
│       ├── order-service/     # 订单管理服务（规划中）
│       └── report-service/    # 报表分析服务（规划中）
├── frontend/
│   └── admin-panel/          # 管理后台 (React + Amis)
│       ├── src/
│       │   ├── components/   # 共享组件
│       │   ├── pages/        # 页面组件
│       │   ├── services/     # API 服务
│       │   ├── stores/       # 状态管理
│       │   └── types/        # TypeScript 类型定义
│       └── package.json
├── docker/                   # Docker 配置文件
│   ├── mysql/               # MySQL 配置
│   └── redis/               # Redis 配置
├── docs/                    # 项目文档
├── scripts/                 # 开发脚本
├── docker-compose.yml       # 开发环境编排
└── CLAUDE.md               # Claude Code 开发指南
```

### 关键目录说明

- **`backend/shared/`**: 包含所有微服务共享的核心业务逻辑
- **`backend/services/`**: 各个独立的微服务实现
- **`frontend/admin-panel/`**: 基于 React + Amis 的管理后台
- **`docker/`**: 数据库和缓存的配置文件
- **`docs/`**: 完整的项目文档和架构说明

## 💻 开发指南

### 后端开发

```bash
cd backend

# 同步 workspace 依赖
go work sync

# 运行特定服务（开发模式）
go run ./gateway
go run ./services/user-service

# 构建所有服务
go build ./...

# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./shared/repository -v

# 清理模块依赖
go mod tidy -C ./shared
```

### 前端开发

```bash
cd frontend/admin-panel

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 代码检查
npm run lint

# 运行测试
npm run test
```

### 多租户开发规范

**⚠️ 重要：多租户数据访问规范**

所有数据访问必须通过 Repository 层，确保租户隔离：

```go
// ✅ 正确 - 自动注入 tenant_id
userRepo := repository.NewUserRepository()
users, err := userRepo.FindAllByTenant(ctx)

// ❌ 错误 - 绕过租户隔离
db := g.DB()
users, err := db.Model("users").All()
```

租户上下文通过 HTTP 中间件自动注入：

```go
// HTTP 中间件自动从 X-Tenant-ID 头注入租户信息
func SomeHandler(r *ghttp.Request) {
    ctx := r.GetCtx() // 包含 tenant_id
    // 使用 ctx 调用 Repository 方法
}
```

## 📖 API 文档

### 健康检查端点

系统提供多层次健康检查：

- `GET /api/v1/health` - 完整健康检查（数据库、Redis、系统资源）
- `GET /api/v1/health/ready` - 就绪检查（所有依赖必须健康）
- `GET /api/v1/health/live` - 存活检查（基础系统检查）
- `GET /api/v1/health/simple` - 快速健康检查
- `GET /api/v1/health/component/{name}` - 单个组件检查

### 认证端点

- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/logout` - 用户登出
- `POST /api/v1/auth/refresh` - 刷新令牌
- `GET /api/v1/auth/profile` - 获取用户信息

### API 规范

- **认证**: Bearer Token (JWT)
- **租户标识**: HTTP Header `X-Tenant-ID`
- **内容类型**: `application/json`
- **版本控制**: URL路径版本控制 (`/api/v1/`)

## 🚀 部署指南

### 开发环境部署

```bash
# 启动完整开发环境
./scripts/docker-dev.sh start

# 查看服务状态
./scripts/docker-dev.sh status

# 查看服务日志
./scripts/docker-dev.sh logs [service-name]

# 停止所有服务
./scripts/docker-dev.sh stop
```

### 生产环境部署

1. **环境准备**
   ```bash
   # 设置生产环境变量
   export GO_ENV=production
   export DB_HOST=your-mysql-host
   export REDIS_HOST=your-redis-host
   export JWT_SECRET=your-secure-jwt-secret
   ```

2. **构建镜像**
   ```bash
   # 构建后端服务
   docker build -f backend/gateway/Dockerfile backend/ -t mer-gateway:latest
   docker build -f backend/services/user-service/Dockerfile backend/ -t mer-user-service:latest
   
   # 构建前端应用
   docker build frontend/admin-panel/ -t mer-admin-panel:latest
   ```

3. **部署服务**
   ```bash
   # 使用生产环境配置启动
   docker-compose -f docker-compose.prod.yml up -d
   ```

### 数据库迁移

```bash
# 数据库迁移文件位于 backend/shared/database/migrations/
# 系统启动时会自动执行未执行的迁移
```

## 🧪 测试

### 运行测试

```bash
# 后端测试
cd backend
go test ./...                              # 运行所有测试
go test ./shared/repository -v              # 运行特定包测试
go test ./shared/test -run TestTenantIsolation -v  # 运行特定测试

# 前端测试
cd frontend/admin-panel
npm run test                               # 运行单元测试
npm run test:coverage                      # 运行测试覆盖率
npm run test:watch                         # 监听模式运行测试
```

### 测试分类

- **单元测试**: 测试单个函数和方法
- **集成测试**: 测试服务间交互和数据库操作
- **E2E测试**: 测试完整的用户流程
- **多租户隔离测试**: 确保租户数据隔离的专项测试

### 测试覆盖率要求

- 核心业务逻辑: ≥ 80%
- Repository 层: ≥ 90%
- 多租户隔离: 100%

## 🔧 故障排查

### 常见问题

1. **Go workspace 问题**
   ```bash
   go work sync  # 同步依赖
   ```

2. **Docker 容器启动失败**
   ```bash
   docker-compose logs -f [service]  # 查看容器日志
   ./scripts/docker-dev.sh status   # 检查服务状态
   ```

3. **前端构建失败**
   ```bash
   cd frontend/admin-panel
   rm -rf node_modules package-lock.json
   npm install
   ```

4. **数据库连接失败**
   ```bash
   # 确认 MySQL 容器状态
   docker-compose ps mysql
   
   # 检查数据库配置
   cat .env | grep MYSQL
   ```

### 调试工具

- **服务日志**: `docker-compose logs -f [service]`
- **健康检查**: `curl http://localhost:8080/api/v1/health`
- **数据库查询**: 通过 GoFrame 的 debug 模式查看 SQL
- **Redis 监控**: `docker exec -it mer_redis redis-cli monitor`

## 🤝 贡献指南

### 开发流程

1. **Fork 项目**
   ```bash
   git clone <your-fork-url>
   cd mer-sys
   ```

2. **创建特性分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **开发和测试**
   ```bash
   # 开发完成后运行测试
   cd backend && go test ./...
   cd frontend/admin-panel && npm run test
   ```

4. **提交变更**
   ```bash
   git add .
   git commit -m "feat: 添加新功能描述"
   git push origin feature/your-feature-name
   ```

5. **创建 Pull Request**

### 代码规范

**Go 代码规范**
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 所有公共函数和类型必须有注释
- 使用中文注释和文档

**前端代码规范**
- 使用 ESLint 进行代码检查
- 遵循 React Hooks 最佳实践
- TypeScript 严格模式
- 组件和函数使用中文注释

**Git 提交规范**
- `feat:` 新功能
- `fix:` 修复问题
- `docs:` 文档更新
- `style:` 代码格式调整
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建和辅助工具

### 重要的开发约定

1. **多租户隔离**: 永远不要绕过 Repository 层的租户过滤
2. **错误处理**: 使用 GoFrame 的错误处理机制
3. **配置管理**: 通过 `g.Cfg()` 访问配置
4. **日志记录**: 使用 `g.Log()` 进行结构化日志
5. **测试先行**: 新功能必须包含相应的测试用例

## 📜 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 📞 联系方式

- **项目维护者**: MER System Team
- **技术支持**: 请提交 [Issue](issues)
- **功能建议**: 请提交 [Feature Request](issues/new?template=feature_request.md)

## 🙏 致谢

感谢以下开源项目为本项目提供的支持：

- [GoFrame](https://goframe.org) - 企业级 Go Web 开发框架
- [Amis](https://aisuda.bce.baidu.com/amis) - 低代码前端框架
- [React](https://reactjs.org) - 用户界面库
- [MySQL](https://www.mysql.com) - 关系型数据库
- [Redis](https://redis.io) - 内存数据结构存储

---

<div align="center">

**🌟 如果这个项目对您有帮助，请给我们一个 Star！**

Made with ❤️ by MER System Team

</div>
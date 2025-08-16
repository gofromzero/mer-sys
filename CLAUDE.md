# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# MER System - 多租户商户管理平台

这是一个全栈项目，采用 Go workspace 管理后端微服务，frontend 目录存放多个前端项目。请在所有交互中默认使用中文进行沟通。

## 语言偏好

- **默认语言**: 中文（简体）
- **备用语言**: 英文（仅在必要时）

## 核心架构

### 多租户隔离架构
项目采用基于 `tenant_id` 的多租户数据隔离模式：
- 所有数据表都包含 `tenant_id` 字段
- `BaseRepository` 自动在所有查询中注入租户过滤条件
- 租户上下文通过中间件自动从 HTTP 头 `X-Tenant-ID` 注入到 `context.Context`
- 严禁绕过 Repository 层直接进行数据库查询

### 技术栈
- **后端**: Go 1.21+ + GoFrame v2.6+ (使用 Go workspace)
- **前端**: React 19+ + TypeScript 5+ + Amis 6+ + Zustand 4+ + Tailwind CSS 3+
- **数据库**: MySQL 8.0 + Redis 7.0+
- **容器化**: Docker + Docker Compose

### 项目结构
```
mer-demo/
├── backend/                 # Go workspace 根目录
│   ├── go.work             # Go workspace 配置文件
│   ├── shared/             # 共享包（核心业务逻辑）
│   │   ├── repository/     # 数据访问层（多租户隔离）
│   │   ├── auth/          # JWT 认证和授权
│   │   ├── cache/         # Redis 缓存封装
│   │   ├── config/        # 数据库和 Redis 配置
│   │   ├── health/        # 健康检查系统
│   │   ├── middleware/    # HTTP 中间件
│   │   └── types/         # 共享数据类型
│   ├── gateway/           # API 网关服务
│   └── services/          # 微服务集合
│       ├── user-service/
│       ├── tenant-service/
│       └── ...
├── frontend/
│   └── admin-panel/       # 管理后台 (React + Amis)
├── docker-compose.yml     # 开发环境容器编排
└── scripts/
    └── docker-dev.sh      # Docker 开发管理脚本
```

## 常用开发命令

### 启动开发环境
```bash
# 启动完整开发环境（推荐）
./scripts/docker-dev.sh start

# 查看服务状态
./scripts/docker-dev.sh status

# 查看服务日志
./scripts/docker-dev.sh logs [service-name]

# 停止所有服务
./scripts/docker-dev.sh stop
```

### 后端开发 (Go workspace)
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

# 运行单个测试函数
go test ./shared/test -run TestTenantIsolation -v

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

# 类型检查
npm run build  # TypeScript 检查包含在构建中
```

### 数据库操作
```bash
# 进入 MySQL 容器
docker exec -it mer_mysql mysql -u mer_user -p mer_system

# 查看迁移状态（需要运行相应的 Go 代码）
# 迁移文件位于 backend/shared/database/migrations/
```

## 核心开发模式

### 多租户数据访问
所有数据访问必须通过 Repository 层：
```go
// ✅ 正确 - 自动注入 tenant_id
userRepo := repository.NewUserRepository()
users, err := userRepo.FindAllByTenant(ctx)

// ❌ 错误 - 绕过租户隔离
db := g.DB()
users, err := db.Model("users").All()
```

### 上下文传递
租户信息通过 context 传递：
```go
// HTTP 中间件自动注入 tenant_id
func SomeHandler(r *ghttp.Request) {
    ctx := r.GetCtx() // 包含 tenant_id
    // 使用 ctx 调用 Repository 方法
}
```

### 前端 API 调用
前端统一通过 AmisRenderer 组件和 API 适配器调用后端：
```typescript
// 使用 Amis schema 定义页面
const schema = {
    type: 'crud',
    api: '/api/v1/users',  // 自动添加租户头
    // ...
}
```

## 健康检查端点

系统提供多层次健康检查：
- `/api/v1/health` - 完整健康检查（数据库、Redis、系统资源）
- `/api/v1/health/ready` - 就绪检查（所有依赖必须健康）
- `/api/v1/health/live` - 存活检查（基础系统检查）
- `/api/v1/health/simple` - 快速健康检查
- `/api/v1/health/component/{name}` - 单个组件检查

## 重要的开发约定

### 代码规范
1. **中文注释和文档**：所有注释和文档默认使用中文
2. **错误处理**：使用 GoFrame 的错误处理机制
3. **配置管理**：通过 `g.Cfg()` 访问配置，禁止直接使用环境变量
4. **日志记录**：使用 `g.Log()` 进行结构化日志记录

### 数据安全
1. **租户隔离**：永远不要绕过 Repository 层的租户过滤
2. **输入验证**：所有用户输入必须验证和清理
3. **敏感信息**：JWT secret、数据库密码等通过环境变量管理

### 测试要求
1. **多租户测试**：每个 Repository 方法都要有租户隔离测试
2. **集成测试**：健康检查、数据库连接、Redis 连接
3. **单元测试**：业务逻辑和工具函数

## 环境配置

复制 `.env.example` 为 `.env` 并根据需要修改配置。主要配置项：
- 数据库连接信息
- Redis 连接信息  
- JWT 密钥
- 服务端口配置

## 故障排查

### 常见问题
1. **Go workspace 问题**：运行 `go work sync` 同步依赖
2. **Docker 容器启动失败**：检查端口占用，查看容器日志
3. **前端构建失败**：清理 node_modules 重新安装依赖
4. **数据库连接失败**：确认 MySQL 容器正常运行，检查配置文件

### 调试工具
- Docker 容器日志：`docker-compose logs -f [service]`
- Go 应用日志：通过健康检查端点查看系统状态
- 数据库查询：通过 GoFrame 的 debug 模式查看 SQL

---

**重要提醒**：在此项目中请始终使用中文进行交流，遵循多租户隔离的开发模式，确保数据安全。
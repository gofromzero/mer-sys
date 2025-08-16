# Unified Project Structure

```
mer-demo/
├── .github/                    # CI/CD workflows (预留)
├── backend/                    # Go workspace根目录
│   ├── go.work                # Go workspace配置
│   ├── shared/                # 共享包
│   │   ├── types/             # 共享类型定义
│   │   ├── constants/         # 常量定义
│   │   ├── utils/             # 工具函数
│   │   └── middleware/        # 共享中间件
│   ├── services/              # 微服务
│   │   ├── user-service/      # 用户服务
│   │   ├── tenant-service/    # 租户服务
│   │   ├── merchant-service/  # 商户服务
│   │   ├── product-service/   # 商品服务
│   │   ├── order-service/     # 订单服务
│   │   ├── fund-service/      # 资金权益服务
│   │   └── report-service/    # 报表服务
│   ├── gateway/               # API网关
│   └── scripts/               # 构建部署脚本
├── frontend/                  # 前端项目集合
│   ├── admin-panel/           # 租户管理后台
│   │   ├── src/
│   │   │   ├── components/    # 组件
│   │   │   ├── pages/         # 页面
│   │   │   ├── hooks/         # Hooks
│   │   │   ├── services/      # API服务
│   │   │   ├── stores/        # 状态管理
│   │   │   └── utils/         # 工具函数
│   │   ├── public/            # 静态资源
│   │   └── package.json
│   ├── merchant-portal/       # 商户运营门户
│   │   └── src/               # 类似结构
│   ├── customer-app/          # 客户服务应用
│   │   └── src/               # 类似结构
│   └── shared/                # 前端共享包
│       ├── components/        # 共享组件
│       ├── types/             # TypeScript类型
│       ├── utils/             # 工具函数
│       └── hooks/             # 共享Hooks
├── infrastructure/            # 基础设施配置
│   ├── terraform/             # Terraform配置
│   ├── docker/                # Docker配置
│   └── k8s/                   # Kubernetes配置
├── docs/                      # 项目文档
│   ├── prd.md                 # 产品需求文档
│   ├── user-service-design.md # 用户服务设计
│   └── architecture.md        # 本架构文档
├── scripts/                   # 项目脚本
│   ├── dev.sh                 # 开发环境启动
│   ├── build.sh               # 构建脚本
│   └── deploy.sh              # 部署脚本
├── .env.example               # 环境变量模板
├── docker-compose.yml         # 开发环境Docker编排
├── package.json               # 根package.json (npm workspaces)
└── README.md                  # 项目说明
```

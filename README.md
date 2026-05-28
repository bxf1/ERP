# ERP System

新一代 AI 原生 ERP 系统，混合架构：元数据引擎驱动 + 业务模块固定开发 + AI 能力贯穿全栈。

## 技术栈

| 层 | 技术 |
|---|------|
| 前端 | React 18 + TypeScript + Vite + Ant Design + Formily + ECharts |
| 后端 | Go 1.22 + Gin + GORM + PostgreSQL + Zap |
| AI | NL2SQL + Builder Agent + RAG 知识库 + 语义层 |
| 基础设施 | 多租户 + RBAC 权限 + 工作流引擎 |

## 项目结构

```
ERP/
├── backend/                    # Go 后端
│   ├── cmd/                    # 入口（server / migrate / vectorize）
│   ├── config/                 # 配置管理
│   ├── internal/               # 内部模块
│   │   ├── database/           # 数据库连接 / 多租户 / 迁移
│   │   ├── middleware/         # CORS / Logger / Tenant / Recovery
│   │   ├── model/              # 元数据模型定义
│   │   ├── mcp/                # AI MCP 协议服务
│   │   ├── gateway/            # AI 安全网关
│   │   └── semantic/           # 语义层
│   ├── pkg/                    # 可复用包
│   │   ├── builder/            # AI Builder Agent
│   │   ├── datadict/           # 数据字典
│   │   ├── database/           # 数据库工具
│   │   ├── embedding/          # 向量嵌入（OpenAI / Mock）
│   │   ├── errors/             # 错误处理
│   │   ├── nl2sql/             # NL2SQL 引擎
│   │   ├── permission/         # RBAC 权限管理
│   │   ├── rag/                # RAG 知识库
│   │   ├── response/           # 统一响应格式
│   │   └── semantic/           # 语义层服务
│   ├── handler/                # HTTP 处理器
│   ├── service/                # 业务服务层
│   ├── repository/             # 数据访问层
│   ├── router/                 # 路由注册
│   └── migrations/             # SQL 迁移脚本
├── frontend/                   # React 前端
│   └── src/
│       ├── components/         # 通用组件
│       │   ├── DynamicForm/    # 动态表单引擎
│       │   ├── NL2SQL/         # NL2SQL 聊天界面
│       │   ├── BuilderAgent/   # AI 建模块助手
│       │   └── Charts/         # 可视化图表
│       ├── pages/              # 业务页面
│       │   ├── User/           # 用户管理
│       │   ├── NL2SQL/         # 自然语言查询
│       │   ├── customer/       # 客户管理
│       │   ├── supplier/       # 供应商管理
│       │   ├── purchase/       # 采购管理
│       │   ├── sales/          # 销售管理
│       │   ├── inventory/      # 库存管理
│       │   ├── stocktaking/    # 库存盘点
│       │   └── dashboard/      # 经营仪表盘
│       ├── layouts/            # 布局组件
│       ├── services/           # API 封装
│       ├── hooks/              # 自定义 Hooks
│       ├── models/             # 状态模型
│       └── utils/              # 工具函数
└── .github/workflows/          # CI/CD
```

## 开发阶段

| 阶段 | 内容 | 状态 |
|------|------|------|
| P0 | 项目初始化与多租户基础设施 | 代码已整合 |
| P1 | 动态表单引擎 / 权限引擎 / 工作流引擎 / 菜单导航 | 代码已整合 |
| P2 | NL2SQL / Builder Agent / MCP 接口 / RAG / 语义层 | 代码已整合 |
| P3 | 进销存业务模块（采购 / 销售 / 库存） | 代码已整合 |

## 已知问题

- **权限前端**：session 76dd0d44 的前端代码使用 Vue 3，需重写为 React 以保持一致
- **Go import 路径**：各模块原 module path 不同，整合后的 import 路径需统一修正
- **前端路由合并**：App.tsx 和 routes.tsx 需要合并所有页面的路由配置

## 开发工作流

1. 从 `main` 创建 feature 分支：`git checkout -b feature/<描述>`
2. 开发完成后 push 分支：`git push origin feature/<描述>`
3. 创建 Pull Request（PR 描述中引用 Multica issue）
4. Code review 通过后合并到 `main`

所有改动必须经过 PR 流程，不直接 push 到 main。

## 快速开始

```bash
# 后端
cd backend
go mod tidy
go run main.go

# 前端
cd frontend
npm install
npm run dev
```

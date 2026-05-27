# ERP System

新一代 ERP 系统，采用混合架构：底层元数据引擎 + 上层业务模块固定开发，AI 能力贯穿全系统。

## 技术栈

- **前端**: React 18 + Ant Design Pro + Formily
- **后端**: Go + Gin
- **数据库**: PostgreSQL
- **工作流**: Temporal

## 项目结构

```
erp/
├── frontend/          # React 前端
├── backend/           # Go 后端
├── docs/              # 文档
└── docker/            # Docker 配置
```

## 开发阶段

- **Phase 0**: 项目初始化与基础设施
- **Phase 1**: 引擎核心（动态表单、权限、工作流）
- **Phase 2**: AI 能力基座（NL2SQL、Builder Agent、RAG）
- **Phase 3**: 业务模块（进销存）

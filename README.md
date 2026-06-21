# 云值守管理系统

前后端分离架构：Go + Vue3 + MySQL

## 项目结构

```
go-yzs/
├── backend/          # Go 后端
│   ├── main.go
│   ├── config/       # 配置
│   ├── database/     # 数据库连接和迁移
│   ├── models/       # 数据模型
│   ├── handlers/     # HTTP 处理器
│   ├── middleware/   # 中间件（JWT 鉴权）
│   ├── routes/       # 路由
│   └── scheduler/    # 定时同步任务
└── frontend/         # Vue3 前端
    └── src/
        ├── api/      # API 请求封装
        ├── stores/   # Pinia 状态管理
        ├── router/   # 路由
        └── views/    # 页面组件
```

## 快速启动

### 1. 创建数据库

```sql
CREATE DATABASE go_yzs CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 启动后端

```bash
cd backend
cp .env.example .env
# 修改 .env 中的数据库密码
export $(cat .env | xargs)
go run main.go
```

### 3. 启动前端

```bash
cd frontend
npm run dev
```

访问 http://localhost:5173

默认账号：admin / admin123

## 功能说明

- **登录页**：Chrome 浏览器检测、单点登录（SSO）、自动鉴权跳转
- **用户管理**：CRUD，角色：管理员/统计员/操作员（仅管理员可见）
- **异常订单**：展示从外部 API 同步的数据，支持关键词/状态/类型/时间筛选
- **数据同步**：每 30 分钟自动同步，重复数据忽略，新数据追加

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/login | 登录 |
| POST | /api/logout | 退出 |
| GET  | /api/me | 当前用户信息 |
| GET  | /api/trades | 异常订单列表 |
| GET  | /api/users | 用户列表（管理员） |
| POST | /api/users | 创建用户（管理员） |
| PUT  | /api/users/:id | 更新用户（管理员） |
| DELETE | /api/users/:id | 删除用户（管理员） |

# Tanzanite 管理后台

基于 Vue 3 + Element Plus 的现代化管理后台系统。

## 技术栈

- **Vue 3** - 渐进式 JavaScript 框架
- **Vite** - 下一代前端构建工具
- **Element Plus** - 基于 Vue 3 的组件库
- **Vue Router** - 官方路由管理器
- **Pinia** - Vue 状态管理
- **Axios** - HTTP 客户端

## 功能特性

- ✅ 用户认证和授权
- ✅ 基于角色的权限控制 (RBAC)
- ✅ 响应式布局
- ✅ 仪表板统计
- 🚧 商品管理
- 🚧 订单管理
- 🚧 用户管理
- 🚧 内容管理
- 🚧 FAQ 管理
- 🚧 图库管理
- 🚧 订阅管理
- 🚧 工单管理
- 🚧 营销管理
- 🚧 系统设置

## 开发指南

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 项目结构

```
web/admin/
├── src/
│   ├── assets/          # 静态资源
│   ├── components/      # 公共组件
│   ├── layouts/         # 布局组件
│   ├── router/          # 路由配置
│   ├── stores/          # 状态管理
│   ├── utils/           # 工具函数
│   ├── views/           # 页面组件
│   ├── App.vue          # 根组件
│   └── main.js          # 入口文件
├── public/              # 公共文件
├── index.html           # HTML 模板
├── vite.config.js       # Vite 配置
└── package.json         # 项目配置
```

## 权限系统

### 角色

- **admin** - 超级管理员（所有权限）
- **manager** - 经理（大部分权限）
- **editor** - 编辑（内容管理权限）
- **support** - 客服（工单和客户权限）
- **viewer** - 查看者（只读权限）

### 权限列表

- `product:view/create/edit/delete` - 商品管理
- `order:view/edit/refund/delete` - 订单管理
- `user:view/create/edit/delete` - 用户管理
- `content:view/create/edit/delete` - 内容管理
- `faq:view/create/edit/delete` - FAQ 管理
- `gallery:view/create/edit/delete` - 图库管理
- `subscription:view/edit/delete/export` - 订阅管理
- `ticket:view/create/edit/assign/close` - 工单管理
- `marketing:view/create/edit/delete` - 营销管理
- `settings:view/edit` - 设置管理
- `logs:view` - 审计日志
- `system:manage` - 系统管理

## API 接口

后端 API 地址：`http://localhost:8080/api/admin`

### 认证接口

- `POST /auth/login` - 登录
- `GET /auth/profile` - 获取用户信息
- `POST /auth/refresh` - 刷新令牌
- `POST /auth/logout` - 登出
- `GET /auth/permissions` - 获取权限列表

### 仪表板接口

- `GET /dashboard/stats` - 获取统计数据
- `GET /dashboard/recent-orders` - 最近订单
- `GET /dashboard/recent-users` - 最近用户
- `GET /dashboard/recent-tickets` - 最近工单
- `GET /dashboard/sales-chart` - 销售图表

## 默认测试账号

需要在数据库中创建管理员账号：

```sql
-- 创建管理员账号（密码需要使用 bcrypt 加密）
INSERT INTO users (email, username, password, role, status) 
VALUES ('admin@example.com', 'admin', '$2a$10$...', 'admin', 'active');
```

## 开发注意事项

1. **路由守卫**：所有需要认证的页面都会自动检查登录状态
2. **权限控制**：使用 `hasPermission()` 方法检查权限
3. **API 调用**：统一使用 `@/utils/axios` 进行 HTTP 请求
4. **状态管理**：使用 Pinia 管理全局状态

## 待实现功能

- [ ] 商品管理完整功能
- [ ] 订单管理完整功能
- [ ] 用户管理完整功能
- [ ] 内容编辑器（富文本）
- [ ] 图片上传组件
- [ ] 数据导出功能
- [ ] 实时通知
- [ ] 多语言支持

## License

MIT

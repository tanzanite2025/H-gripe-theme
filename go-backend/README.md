# Tanzanite Go Backend

WordPress 到 Go 后端迁移项目 - 高性能、可扩展的 RESTful API 服务

## 📋 项目概述

这是一个将 WordPress 后端迁移到 Go 的完整实现，提供：
- 🚀 高性能 RESTful API
- 🌍 34种语言国际化支持
- 🔐 JWT 认证系统
- 📦 Redis 缓存
- 🗄️ PostgreSQL/MySQL 数据库
- 🐳 Docker 容器化部署

## 🏗️ 项目结构

```
tanzanite-go-backend/
├── cmd/
│   └── server/              # 应用入口 (main.go)
├── internal/
│   ├── api/
│   │   ├── middleware/      # 中间件 (认证、CORS、日志等)
│   │   └── v1/              # API v1 路由和处理器
│   │       ├── auth/        # 认证相关
│   │       ├── cart/        # 购物车
│   │       ├── content/     # 内容管理 (文章、FAQ)
│   │       ├── product/     # 产品管理
│   │       └── settings/    # 设置管理
│   ├── domain/              # 领域模型
│   │   ├── user/            # 用户模型
│   │   ├── post/            # 文章模型
│   │   ├── product/         # 产品模型
│   │   ├── faq/             # FAQ模型
│   │   └── ...
│   ├── repository/          # 数据访问层
│   ├── service/             # 业务逻辑层
│   └── pkg/                 # 内部工具包
│       ├── cache/           # Redis缓存
│       ├── config/          # 配置管理
│       ├── database/        # 数据库连接
│       ├── i18n/            # 国际化
│       └── logger/          # 日志系统
├── migrations/              # 数据库迁移文件
├── config/                  # 配置文件
├── scripts/                 # 工具脚本
│   ├── wp-data-export.php   # WordPress数据导出
│   ├── import-data.go       # 数据导入工具
│   ├── setup-dev.sh         # 开发环境设置
│   └── start.sh/bat         # 启动脚本
├── docker-compose.yml       # Docker编排
├── Dockerfile               # Docker镜像
├── Makefile                 # 构建脚本
└── API.md                   # API文档
```

## 🚀 快速开始

### 前置要求

- **Go 1.21+** - [安装指南](https://go.dev/doc/install)
- **PostgreSQL 15+** 或 **MySQL 8+**
- **Redis 7+**
- **Docker** (可选，推荐)

### 方式一：使用 Docker (推荐)

```bash
# 1. 克隆项目
cd go-backend

# 2. 创建配置文件
cp .env.example .env
cp config/config.example.yaml config/config.yaml

# 3. 启动所有服务
docker-compose up -d

# 4. 查看日志
docker-compose logs -f api

# 5. 访问健康检查
curl http://localhost:9000/health
```

### 方式二：本地开发

#### Windows

```cmd
# 1. 安装依赖
go mod download

# 2. 创建配置
copy config\config.example.yaml config\config.yaml
copy .env.example .env

# 3. 启动 PostgreSQL 和 Redis (使用 Docker)
docker-compose up -d postgres redis

# 4. 运行应用
scripts\start.bat
```

#### Linux/Mac

```bash
# 1. 运行设置脚本
chmod +x scripts/setup-dev.sh
./scripts/setup-dev.sh

# 2. 启动服务
make run

# 或使用热重载开发模式
make dev
```

## 📝 配置说明

### config/config.yaml

主要配置项：

```yaml
server:
  port: ":9000"
  mode: "debug"  # debug, release, test

database:
  driver: "postgres"  # postgres 或 mysql
  host: "localhost"
  port: 5432
  username: "tanzanite"
  password: "your_password"
  database: "tanzanite"

redis:
  host: "localhost"
  port: 6379

jwt:
  secret: "your-secret-key"
  expire_hours: 24

i18n:
  default_locale: "en"
  supported_locales: [en, zh, fr, de, es, ...]
```

### .env

环境变量（敏感信息）：

```env
DB_PASSWORD=your_secure_password
JWT_SECRET=your_jwt_secret_key
```

## 🔧 开发命令

```bash
# 构建应用
make build

# 运行应用
make run

# 热重载开发模式 (需要安装 air)
make dev

# 运行测试
make test

# 测试覆盖率
make test-coverage

# 代码格式化
make fmt

# 代码检查
make lint

# Docker 构建
make docker-build

# Docker 启动
make docker-up

# Docker 停止
make docker-down

# 查看所有命令
make help
```

## 📊 数据迁移

### 从 WordPress 导出数据

```bash
# 在 WordPress 根目录运行
php go-backend/scripts/wp-data-export.php
```

这将在 `scripts/export/` 目录生成以下文件：
- `users.json` - 用户数据
- `posts.json` - 文章数据
- `products.json` - 产品数据
- `settings.json` - 设置数据
- `faqs.json` - FAQ数据

### 导入到 Go 后端

```bash
# 确保导出的文件在 scripts/export/ 目录
go run scripts/import-data.go
```

## 🌐 API 文档

详细的 API 文档请查看 [API.md](./API.md)

### 主要端点

```
# 认证
POST   /api/v1/auth/register      # 用户注册
POST   /api/v1/auth/login         # 用户登录
GET    /api/v1/auth/profile       # 获取用户信息

# 内容
GET    /api/v1/content/posts      # 文章列表
GET    /api/v1/content/posts/:id  # 文章详情
GET    /api/v1/content/faqs       # FAQ列表

# 产品
GET    /api/v1/products           # 产品列表
GET    /api/v1/products/:id       # 产品详情

# 购物车
GET    /api/v1/cart/summary       # 购物车摘要
POST   /api/v1/cart/add           # 添加到购物车

# 设置
GET    /api/v1/settings/site      # 站点设置
GET    /api/v1/settings/quick-buy # 快速购买设置

# 健康检查
GET    /health                    # 服务健康状态
```

### 国际化支持

所有端点支持多语言，通过以下方式指定：

```bash
# 1. URL 路径
GET /fr/api/v1/products

# 2. Header
GET /api/v1/products
Accept-Language: fr

# 3. Cookie
Cookie: locale=fr
```

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service/...

# 带覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🐳 Docker 部署

### 开发环境

```bash
docker-compose up -d
```

### 生产环境

```bash
# 构建镜像
docker build -t tanzanite-api:latest .

# 运行容器
docker run -d \
  -p 9000:9000 \
  -e DATABASE_HOST=your-db-host \
  -e REDIS_HOST=your-redis-host \
  --name tanzanite-api \
  tanzanite-api:latest
```

## 📈 性能优化

- ✅ Redis 多层缓存
- ✅ 数据库连接池
- ✅ GORM 预加载优化
- ✅ 速率限制 (100 req/min)
- ✅ 并发处理
- ✅ 静态资源缓存

## 🔒 安全特性

- ✅ JWT 认证
- ✅ 密码 bcrypt 加密
- ✅ CORS 配置
- ✅ 速率限制
- ✅ SQL 注入防护 (GORM)
- ✅ XSS 防护

## 📚 技术栈

- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL / MySQL
- **缓存**: Redis
- **认证**: JWT
- **日志**: Zap
- **配置**: Viper
- **容器**: Docker

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证

## 📞 支持

- 📧 Email: support@tanzanite.site
- 📖 文档: [API.md](./API.md)
- 🐛 问题: [GitHub Issues](https://github.com/tanzanite/go-backend/issues)

## 🗺️ 路线图

- [x] 基础架构搭建
- [x] 认证系统
- [x] 内容管理 API
- [x] 产品管理 API
- [x] 购物车功能
- [x] 多语言支持
- [ ] 订单系统
- [ ] 支付集成
- [ ] 邮件通知
- [ ] 管理后台 API
- [ ] GraphQL 支持
- [ ] WebSocket 实时通信

---

**版本**: v1.0.0  
**最后更新**: 2026-05-25

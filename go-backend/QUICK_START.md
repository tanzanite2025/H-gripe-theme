# 🚀 快速开始指南

5分钟内启动 Tanzanite Go Backend！

## 📋 前置检查

确保已安装：
- ✅ Go 1.21+ (`go version`)
- ✅ Docker & Docker Compose (`docker --version`)
- ✅ Git (`git --version`)

## 🎯 三步启动

### 步骤 1: 克隆并配置

```bash
# 进入项目目录
cd go-backend

# 复制配置文件
cp config/config.example.yaml config/config.yaml
cp .env.example .env

# (可选) 编辑配置
# nano config/config.yaml
```

### 步骤 2: 启动服务

```bash
# 使用 Docker 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f app
```

### 步骤 3: 验证

```bash
# 健康检查
curl http://localhost:8080/health

# 预期响应:
# {"status":"ok","version":"1.0.0"}
```

## ✅ 成功！

Docker 模式下 API 运行在: **http://localhost:8080**

本地 `go run` 模式默认读取 `config/config.example.yaml` 中的 `:9000`，可通过 `SERVER_PORT` 覆盖。

---

## 🧪 测试 API

### 1. 注册用户

```bash
curl -X POST http://localhost:9000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123"
  }'
```

### 2. 登录

```bash
curl -X POST http://localhost:9000/api/v1/auth/login \
  -c cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "email_or_username": "test@example.com",
    "password": "password123"
  }'
```

登录成功后会写入 `auth_token`、`refresh_token`、`csrf_token` Cookie。

### 3. 获取用户信息

```bash
curl http://localhost:9000/api/v1/auth/profile \
  -b cookies.txt
```

### 4. 获取产品列表

```bash
curl http://localhost:9000/api/v1/products?page=1&page_size=10
```

### 5. 获取站点设置

```bash
curl http://localhost:9000/api/v1/settings/site \
  -H "Accept-Language: en"
```

---

## 🛠️ 本地开发（不使用 Docker）

### Windows

```cmd
REM 1. 启动数据库和Redis (使用Docker)
docker-compose up -d postgres redis

REM 2. 安装依赖
go mod download

REM 3. 运行应用
go run cmd\server\main.go
```

### Linux/Mac

```bash
# 1. 运行设置脚本
chmod +x scripts/setup-dev.sh
./scripts/setup-dev.sh

# 2. 启动应用
make run

# 或使用热重载
make dev
```

---

## 📊 离线导入数据

将需要导入的 JSON 文件放到 `data/import/`，然后按模块运行 Go 导入命令：

```bash
go run ./cmd/import/settings -input data/import/settings.json
go run ./cmd/import/faqs -input data/import/faqs.json
```

---

## 🐛 常见问题

### 端口已被占用

```bash
# 修改 config/config.yaml 中的端口
server:
  port: ":9001"  # 改为其他端口
```

### 数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
docker-compose ps

# 重启数据库
docker-compose restart postgres
```

### Redis 连接失败

```bash
# 检查 Redis 是否运行
docker-compose ps

# 重启 Redis
docker-compose restart redis
```

### 查看详细日志

```bash
# Docker 日志
docker-compose logs -f app

# 应用日志
tail -f /var/log/tanzanite/app.log
```

---

## 📚 下一步

- 📖 阅读 [API 文档](./API.md)
- 🧭 阅读 [PHP → Go 迁移工作流](../docs/PHP_TO_GO_MIGRATION_WORKFLOW.md)
- 🚀 查看 [部署指南](./DEPLOYMENT.md)
- 🔧 了解 [配置选项](./config/config.example.yaml)
- 🧪 运行测试: `make test`

---

## 🆘 需要帮助？

- 📧 Email: support@tanzanite.site
- 📖 完整文档: [README.md](./README.md)
- 🐛 报告问题: GitHub Issues

---

**祝你使用愉快！** 🎉

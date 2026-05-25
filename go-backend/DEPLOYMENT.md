# Tanzanite Go Backend 部署指南

## 📋 目录

1. [环境准备](#环境准备)
2. [本地开发部署](#本地开发部署)
3. [Docker 部署](#docker-部署)
4. [生产环境部署](#生产环境部署)
5. [Nginx 配置](#nginx-配置)
6. [监控和日志](#监控和日志)
7. [故障排查](#故障排查)

---

## 环境准备

### 系统要求

- **操作系统**: Linux (Ubuntu 20.04+), macOS, Windows Server
- **CPU**: 2核心+
- **内存**: 4GB+
- **磁盘**: 20GB+

### 软件依赖

- Go 1.21+
- PostgreSQL 15+ 或 MySQL 8+
- Redis 7+
- Nginx (生产环境)
- Docker & Docker Compose (可选)

---

## 本地开发部署

### 1. 安装 Go

```bash
# Linux
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# macOS
brew install go

# Windows
# 下载安装包: https://go.dev/dl/
```

### 2. 安装数据库

#### PostgreSQL

```bash
# Ubuntu
sudo apt update
sudo apt install postgresql postgresql-contrib

# macOS
brew install postgresql
brew services start postgresql

# 创建数据库
sudo -u postgres psql
CREATE DATABASE tanzanite;
CREATE USER tanzanite WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE tanzanite TO tanzanite;
```

#### Redis

```bash
# Ubuntu
sudo apt install redis-server
sudo systemctl start redis

# macOS
brew install redis
brew services start redis
```

### 3. 配置应用

```bash
# 克隆项目
cd go-backend

# 复制配置文件
cp config/config.example.yaml config/config.yaml
cp .env.example .env

# 编辑配置
nano config/config.yaml
nano .env
```

### 4. 运行应用

```bash
# 下载依赖
go mod download

# 运行
go run cmd/server/main.go

# 或使用热重载
air
```

---

## Docker 部署

### 1. 安装 Docker

```bash
# Ubuntu
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. 配置环境变量

```bash
# 创建 .env 文件
cat > .env << EOF
DB_PASSWORD=your_secure_password_here
JWT_SECRET=your_jwt_secret_key_change_this
SERVER_MODE=release
EOF
```

### 3. 启动服务

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f api

# 停止服务
docker-compose down

# 重启服务
docker-compose restart api
```

### 4. 数据持久化

Docker volumes 自动管理数据持久化：
- `postgres_data`: PostgreSQL 数据
- `redis_data`: Redis 数据

备份数据：
```bash
# 备份 PostgreSQL
docker exec tanzanite-postgres pg_dump -U tanzanite tanzanite > backup.sql

# 恢复 PostgreSQL
docker exec -i tanzanite-postgres psql -U tanzanite tanzanite < backup.sql
```

---

## 生产环境部署

### 1. 服务器准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装必要工具
sudo apt install -y git curl wget vim

# 创建应用用户
sudo useradd -m -s /bin/bash tanzanite
sudo su - tanzanite
```

### 2. 部署应用

```bash
# 克隆代码
git clone https://github.com/your-org/tanzanite-go-backend.git
cd tanzanite-go-backend

# 配置
cp config/config.example.yaml config/config.yaml
cp .env.example .env

# 编辑配置（生产环境）
nano config/config.yaml
# 修改:
# - server.mode: "release"
# - database: 生产数据库配置
# - redis: 生产Redis配置
# - log.level: "info"
# - log.output: "/var/log/tanzanite/app.log"

# 构建
go build -o tanzanite-api ./cmd/server
```

### 3. 创建 Systemd 服务

```bash
sudo nano /etc/systemd/system/tanzanite-api.service
```

内容：
```ini
[Unit]
Description=Tanzanite Go Backend API
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=tanzanite
WorkingDirectory=/home/tanzanite/tanzanite-go-backend
ExecStart=/home/tanzanite/tanzanite-go-backend/tanzanite-api
Restart=on-failure
RestartSec=5s

# 环境变量
Environment="GIN_MODE=release"

# 日志
StandardOutput=append:/var/log/tanzanite/app.log
StandardError=append:/var/log/tanzanite/error.log

# 安全设置
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
# 创建日志目录
sudo mkdir -p /var/log/tanzanite
sudo chown tanzanite:tanzanite /var/log/tanzanite

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable tanzanite-api
sudo systemctl start tanzanite-api

# 查看状态
sudo systemctl status tanzanite-api

# 查看日志
sudo journalctl -u tanzanite-api -f
```

---

## Nginx 配置

### 1. 安装 Nginx

```bash
sudo apt install nginx
```

### 2. 配置反向代理

```bash
sudo nano /etc/nginx/sites-available/tanzanite
```

内容：
```nginx
upstream go_backend {
    server 127.0.0.1:9000;
    keepalive 32;
}

# HTTP -> HTTPS 重定向
server {
    listen 80;
    server_name tanzanite.site www.tanzanite.site;
    return 301 https://$server_name$request_uri;
}

# HTTPS 配置
server {
    listen 443 ssl http2;
    server_name tanzanite.site www.tanzanite.site;

    # SSL 证书
    ssl_certificate /etc/letsencrypt/live/tanzanite.site/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/tanzanite.site/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 日志
    access_log /var/log/nginx/tanzanite-access.log;
    error_log /var/log/nginx/tanzanite-error.log;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # API 路由
    location /api/ {
        proxy_pass http://go_backend;
        proxy_http_version 1.1;
        
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 健康检查
    location /health {
        proxy_pass http://go_backend;
        access_log off;
    }

    # 静态文件（如果有）
    location /static/ {
        alias /var/www/tanzanite/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

启用配置：
```bash
sudo ln -s /etc/nginx/sites-available/tanzanite /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 3. SSL 证书（Let's Encrypt）

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d tanzanite.site -d www.tanzanite.site

# 自动续期
sudo certbot renew --dry-run
```

---

## 监控和日志

### 1. 应用日志

```bash
# 实时查看日志
sudo journalctl -u tanzanite-api -f

# 查看最近100行
sudo journalctl -u tanzanite-api -n 100

# 查看特定时间段
sudo journalctl -u tanzanite-api --since "2026-05-25 10:00:00"
```

### 2. Nginx 日志

```bash
# 访问日志
sudo tail -f /var/log/nginx/tanzanite-access.log

# 错误日志
sudo tail -f /var/log/nginx/tanzanite-error.log
```

### 3. 数据库监控

```bash
# PostgreSQL 连接数
sudo -u postgres psql -c "SELECT count(*) FROM pg_stat_activity;"

# Redis 信息
redis-cli info
```

### 4. 系统资源监控

```bash
# CPU 和内存
htop

# 磁盘使用
df -h

# 网络连接
netstat -tulpn | grep :9000
```

---

## 故障排查

### 应用无法启动

```bash
# 检查配置文件
cat config/config.yaml

# 检查端口占用
sudo lsof -i :9000

# 检查数据库连接
psql -h localhost -U tanzanite -d tanzanite

# 检查 Redis 连接
redis-cli ping
```

### 性能问题

```bash
# 查看 Go 应用性能
curl http://localhost:9000/debug/pprof/

# 数据库慢查询
sudo -u postgres psql -c "SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"

# Redis 慢日志
redis-cli slowlog get 10
```

### 内存泄漏

```bash
# 查看内存使用
ps aux | grep tanzanite-api

# Go 内存分析
curl http://localhost:9000/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

---

## 备份策略

### 数据库备份

```bash
# 创建备份脚本
cat > /home/tanzanite/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/home/tanzanite/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份 PostgreSQL
pg_dump -U tanzanite tanzanite | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# 删除7天前的备份
find $BACKUP_DIR -name "db_*.sql.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
EOF

chmod +x /home/tanzanite/backup.sh

# 添加到 crontab (每天凌晨2点)
crontab -e
# 添加: 0 2 * * * /home/tanzanite/backup.sh
```

---

## 更新部署

### 零停机更新

```bash
# 1. 拉取最新代码
git pull origin main

# 2. 构建新版本
go build -o tanzanite-api-new ./cmd/server

# 3. 测试新版本
./tanzanite-api-new &
NEW_PID=$!
sleep 5
curl http://localhost:9001/health
kill $NEW_PID

# 4. 替换旧版本
mv tanzanite-api tanzanite-api-old
mv tanzanite-api-new tanzanite-api

# 5. 重启服务
sudo systemctl restart tanzanite-api

# 6. 验证
curl http://localhost:9000/health

# 7. 如果成功，删除旧版本
rm tanzanite-api-old
```

---

## 安全建议

1. **防火墙配置**
```bash
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

2. **定期更新**
```bash
sudo apt update && sudo apt upgrade -y
```

3. **限制数据库访问**
```bash
# PostgreSQL: 编辑 pg_hba.conf
# 只允许本地连接
```

4. **使用强密码**
- 数据库密码
- JWT Secret
- Redis 密码（如果启用）

5. **定期备份**
- 数据库
- 配置文件
- 应用日志

---

**文档版本**: v1.0.0  
**最后更新**: 2026-05-25

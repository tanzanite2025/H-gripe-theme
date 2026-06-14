# 订阅系统快速参考

## 快速开始

### 1. 运行数据库迁移

```bash
mysql -u root -p tanzanite_db < migrations/008_enhance_subscription_system.sql
```

### 2. 启动服务器

```bash
go run cmd/server/main.go
```

### 3. 测试 API

```powershell
.\test-subscription-api.ps1
```

---

## API 端点速查

### 公开端点

#### 订阅
```bash
POST /api/v1/subscriptions
Content-Type: application/json

{
  "email": "user@example.com",
  "source": "website",
  "locale": "zh",
  "tags": ["newsletter", "promotions"]
}
```

#### 通过令牌退订
```bash
GET /api/v1/subscriptions/unsubscribe/:token
```

#### 通过邮箱退订
```bash
POST /api/v1/subscriptions/unsubscribe
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### 重新订阅
```bash
POST /api/v1/subscriptions/resubscribe
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### 获取订阅状态
```bash
GET /api/v1/subscriptions/status/:email
```

---

### 管理员端点（需要认证）

#### 获取所有订阅
```bash
GET /api/v1/admin/subscriptions?page=1&page_size=20&status=active
Authorization: Bearer <token>
```

#### 根据标签获取订阅
```bash
GET /api/v1/admin/subscriptions/tags?tags=newsletter,promotions&page=1&page_size=20
Authorization: Bearer <token>
```

#### 获取统计信息
```bash
GET /api/v1/admin/subscriptions/stats
Authorization: Bearer <token>
```

#### 删除订阅
```bash
DELETE /api/v1/admin/subscriptions/:email
Authorization: Bearer <token>
```

#### 导出邮箱列表
```bash
# 导出所有活跃邮箱
GET /api/v1/admin/subscriptions/export
Authorization: Bearer <token>

# 按标签导出
GET /api/v1/admin/subscriptions/export?tags=newsletter
Authorization: Bearer <token>
```

---

## 数据迁移

### 从 WordPress 导出

1. 编辑 `scripts/wordpress-export/export-subscriptions.php`，配置数据库连接
2. 运行导出脚本：

```bash
php scripts/wordpress-export/export-subscriptions.php
```

3. 数据将导出到 `data/subscriptions.json`

### 导入到 Go 后端

```bash
go run ./cmd/import/subscriptions data/subscriptions.json
```

---

## 常用场景

### 场景 1: 用户订阅新闻通讯

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "source": "newsletter_form",
    "locale": "zh",
    "tags": ["newsletter"]
  }'
```

### 场景 2: 用户点击退订链接

```bash
# 退订链接格式: https://example.com/unsubscribe?token=abc123...
curl http://localhost:8080/api/v1/subscriptions/unsubscribe/abc123def456...
```

### 场景 3: 管理员导出营销邮箱列表

```bash
curl http://localhost:8080/api/v1/admin/subscriptions/export?tags=promotions \
  -H "Authorization: Bearer <admin_token>"
```

### 场景 4: 查看订阅统计

```bash
curl http://localhost:8080/api/v1/admin/subscriptions/stats \
  -H "Authorization: Bearer <admin_token>"
```

---

## 响应示例

### 订阅成功响应

```json
{
  "message": "Subscribed successfully",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "status": "active",
    "locale": "zh",
    "source": "website",
    "tags": "newsletter,promotions",
    "unsub_token": "abc123def456...",
    "subscribed_at": "2024-01-15T10:30:00Z",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### 统计信息响应

```json
{
  "total": 1500,
  "active": 1200,
  "unsubscribed": 250,
  "bounced": 50,
  "by_source": {
    "website": 800,
    "popup": 400,
    "api": 300
  },
  "by_locale": {
    "zh": 900,
    "en": 600
  }
}
```

### 邮箱导出响应

```json
{
  "emails": [
    "user1@example.com",
    "user2@example.com",
    "user3@example.com"
  ],
  "count": 3
}
```

---

## 错误处理

### 重复订阅

```json
{
  "error": "Email already subscribed"
}
```
HTTP 状态码: 409 Conflict

### 无效的退订令牌

```json
{
  "error": "invalid unsubscribe token"
}
```
HTTP 状态码: 400 Bad Request

### 订阅不存在

```json
{
  "error": "Subscription not found"
}
```
HTTP 状态码: 404 Not Found

---

## 标签使用建议

### 推荐标签

- `newsletter` - 新闻通讯
- `promotions` - 促销信息
- `product_updates` - 产品更新
- `blog` - 博客文章
- `events` - 活动通知
- `vip` - VIP 用户
- `locale_zh` - 中文用户
- `locale_en` - 英文用户

### 标签命名规范

- 使用小写字母
- 使用下划线分隔单词
- 保持简短和描述性
- 避免特殊字符

---

## 性能优化

### 数据库索引

系统已创建以下索引：
- `idx_subscriptions_status` - 状态查询
- `idx_subscriptions_tags` - 标签查询
- `idx_subscriptions_unsub_token` - 退订令牌查询
- `idx_subscriptions_subscribed_at` - 时间范围查询

### 分页建议

- 默认每页 20 条记录
- 最大每页 100 条记录
- 大量数据导出使用 export 端点

---

## 安全注意事项

1. **退订令牌**: 使用加密安全的随机数生成器（crypto/rand）
2. **邮箱验证**: 所有邮箱输入都经过格式验证
3. **管理员端点**: 需要认证，建议添加角色检查
4. **速率限制**: 建议为订阅端点添加速率限制
5. **HTTPS**: 生产环境必须使用 HTTPS

---

## 故障排查

### 问题: 编译错误 "undefined: subscription"

**解决方案**: 确保在 `router.go` 中导入了 subscription 包：
```go
import (
    "tanzanite/internal/api/v1/subscription"
    // ...
)
```

### 问题: 数据库迁移失败

**解决方案**: 
1. 检查数据库连接
2. 确保有足够的权限
3. 检查表是否已存在

### 问题: 测试脚本无法运行

**解决方案**:
1. 确保服务器正在运行
2. 检查端口是否正确（默认 8080）
3. 对于管理员测试，需要先设置 `$adminToken` 变量

---

## 相关文档

- [订阅系统完成报告](SUBSCRIPTION_SYSTEM_COMPLETE.md)
- [Week 4 完成总结](WEEK4_DAY1-3_COMPLETE_SUMMARY.md)
- [API 文档](API.md)
- [部署指南](DEPLOYMENT.md)

---

## 支持

如有问题，请查看：
1. 完整文档: `SUBSCRIPTION_SYSTEM_COMPLETE.md`
2. 测试脚本: `test-subscription-api.ps1`
3. 代码示例: `internal/api/v1/subscription/handler.go`

---

*最后更新: 2024-01-15*

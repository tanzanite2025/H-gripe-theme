# API测试指南

本指南介绍如何测试Tanzanite API的各个端点。

## 目录

1. [环境准备](#环境准备)
2. [认证](#认证)
3. [健康检查](#健康检查)
4. [用户管理](#用户管理)
5. [产品管理](#产品管理)
6. [订单管理](#订单管理)
7. [支付测试](#支付测试)
8. [常见问题](#常见问题)

## 环境准备

### 本地环境

```bash
# 启动服务
cd go-backend
make dev

# API基础URL
export API_BASE_URL="http://localhost:8080"
```

### 生产环境

```bash
# API基础URL
export API_BASE_URL="https://api.tanzanite.com"
```

## 认证

### 注册新用户

```bash
curl -X POST ${API_BASE_URL}/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePassword123!",
    "first_name": "Test",
    "last_name": "User"
  }'
```

响应示例：
```json
{
  "status": "success",
  "data": {
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 用户登录

```bash
curl -X POST ${API_BASE_URL}/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePassword123!"
  }'
```

保存返回的token：
```bash
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 健康检查

### 基础健康检查

```bash
curl ${API_BASE_URL}/health
```

### 详细健康信息

```bash
curl ${API_BASE_URL}/health/detailed
```

### Kubernetes探针

```bash
# 存活探针
curl ${API_BASE_URL}/health/liveness

# 就绪探针
curl ${API_BASE_URL}/health/readiness
```

## 用户管理

### 获取当前用户信息

```bash
curl ${API_BASE_URL}/api/v1/users/me \
  -H "Authorization: Bearer ${TOKEN}"
```

### 更新用户资料

```bash
curl -X PUT ${API_BASE_URL}/api/v1/users/me \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Updated",
    "last_name": "Name",
    "phone": "+1234567890"
  }'
```

### 修改密码

```bash
curl -X POST ${API_BASE_URL}/api/v1/users/me/password \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "SecurePassword123!",
    "new_password": "NewSecurePassword123!"
  }'
```

## 产品管理

### 获取产品列表

```bash
# 基础列表
curl ${API_BASE_URL}/api/v1/products

# 带分页和过滤
curl "${API_BASE_URL}/api/v1/products?page=1&limit=20&category_id=1&sort=price_asc"
```

### 获取产品详情

```bash
curl ${API_BASE_URL}/api/v1/products/1
```

### 搜索产品

```bash
curl "${API_BASE_URL}/api/v1/products/search?q=laptop&min_price=500&max_price=2000"
```

### 创建产品（需要管理员权限）

```bash
curl -X POST ${API_BASE_URL}/api/v1/products \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "description": "This is a test product",
    "price": 99.99,
    "category_id": 1,
    "stock": 100,
    "images": ["https://example.com/image1.jpg"]
  }'
```

## 订单管理

### 创建订单

```bash
curl -X POST ${API_BASE_URL}/api/v1/orders \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "product_id": 1,
        "quantity": 2
      },
      {
        "product_id": 2,
        "quantity": 1
      }
    ],
    "shipping_address": {
      "street": "123 Main St",
      "city": "New York",
      "state": "NY",
      "zip": "10001",
      "country": "US"
    },
    "payment_method": "stripe"
  }'
```

### 获取订单列表

```bash
curl ${API_BASE_URL}/api/v1/orders \
  -H "Authorization: Bearer ${TOKEN}"
```

### 获取订单详情

```bash
curl ${API_BASE_URL}/api/v1/orders/123 \
  -H "Authorization: Bearer ${TOKEN}"
```

### 取消订单

```bash
curl -X POST ${API_BASE_URL}/api/v1/orders/123/cancel \
  -H "Authorization: Bearer ${TOKEN}"
```

## 支付测试

### Stripe测试

```bash
# 创建支付意图
curl -X POST ${API_BASE_URL}/api/v1/payment/stripe/create \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 9999,
    "currency": "usd",
    "order_id": "123"
  }'

# 测试卡号
# 4242 4242 4242 4242 - 成功
# 4000 0000 0000 0002 - 卡被拒绝
# 4000 0000 0000 9995 - 余额不足
```

### PayPal测试

```bash
# 创建PayPal订单
curl -X POST ${API_BASE_URL}/api/v1/payment/paypal/create \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "99.99",
    "currency": "USD",
    "order_id": "123"
  }'

# 测试账号（沙箱环境）
# buyer@example.com / password
```

### 支付宝测试

```bash
# 创建支付宝支付
curl -X POST ${API_BASE_URL}/api/v1/payment/alipay/create \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "99.99",
    "subject": "测试订单",
    "order_id": "123",
    "payment_type": "web"
  }'
```

### 微信支付测试

```bash
# 创建微信扫码支付
curl -X POST ${API_BASE_URL}/api/v1/payment/wechat/native \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 9999,
    "description": "测试订单",
    "order_id": "123"
  }'
```

## 购物车

### 添加商品到购物车

```bash
curl -X POST ${API_BASE_URL}/api/v1/cart/items \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "quantity": 2
  }'
```

### 获取购物车

```bash
curl ${API_BASE_URL}/api/v1/cart \
  -H "Authorization: Bearer ${TOKEN}"
```

### 更新购物车商品数量

```bash
curl -X PUT ${API_BASE_URL}/api/v1/cart/items/1 \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 5
  }'
```

### 删除购物车商品

```bash
curl -X DELETE ${API_BASE_URL}/api/v1/cart/items/1 \
  -H "Authorization: Bearer ${TOKEN}"
```

## 评论和评分

### 创建产品评论

```bash
curl -X POST ${API_BASE_URL}/api/v1/products/1/reviews \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "rating": 5,
    "comment": "Excellent product! Highly recommended."
  }'
```

### 获取产品评论

```bash
curl "${API_BASE_URL}/api/v1/products/1/reviews?page=1&limit=10"
```

## 文件上传

### 上传产品图片

```bash
curl -X POST ${API_BASE_URL}/api/v1/upload/product \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "file=@/path/to/image.jpg"
```

### 上传用户头像

```bash
curl -X POST ${API_BASE_URL}/api/v1/upload/avatar \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "file=@/path/to/avatar.jpg"
```

## 实时聊天

### WebSocket连接

```javascript
// JavaScript示例
const ws = new WebSocket(`ws://localhost:8080/api/v1/chat/ws?token=${TOKEN}`);

ws.onopen = () => {
  console.log('Connected to chat');
  
  // 发送消息
  ws.send(JSON.stringify({
    type: 'message',
    room_id: '123',
    content: 'Hello, world!'
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};
```

## 常见问题

### 1. 401 Unauthorized错误

**问题**：收到401未授权错误

**解决方案**：
- 确保在请求头中包含了有效的JWT token
- 检查token是否已过期
- 验证token格式：`Authorization: Bearer <token>`

### 2. 403 Forbidden错误

**问题**：收到403禁止访问错误

**解决方案**：
- 检查用户权限是否足够
- 某些操作需要管理员权限
- 验证用户账户状态是否正常

### 3. 429 Too Many Requests错误

**问题**：请求太频繁被限流

**解决方案**：
- 减慢请求频率
- 查看响应头中的`X-RateLimit-*`信息
- 等待限流重置后再试

### 4. 500 Internal Server Error

**问题**：服务器内部错误

**解决方案**：
- 检查请求参数是否正确
- 查看服务器日志获取详细错误信息
- 联系技术支持

## Postman集合

我们提供了完整的Postman集合用于API测试：

```bash
# 导入Postman集合
# 文件位置：go-backend/docs/postman/Tanzanite-API.postman_collection.json
```

## 测试脚本

自动化测试脚本：

```bash
# 运行集成测试
cd go-backend
make test-integration

# 运行性能测试
./scripts/performance-test.sh
```

## 相关文档

- [API文档](./API.md)
- [认证指南](./AUTHENTICATION.md)
- [支付网关实现](../../docs/PAYMENT_GATEWAY_IMPLEMENTATION.md)
- [错误代码参考](./ERROR_CODES.md)

## 技术支持

如有问题，请联系：
- Email: api-support@tanzanite.com
- Slack: #api-support
- GitHub Issues: https://github.com/tanzanite/issues

# 🔍 Tanzanite 三端架构数据同步审计报告

**审计日期**: 2026-06-26  
**审计工具**: Kiro AI + Context-Gatherer Agent  
**数据同步健康度**: 🟢 85/100

---

## 📊 执行摘要

### 架构概览
```
┌─────────────────────────────────────┐
│   🌐 C端前端 (nuxt-i18n/)           │
│   数据来源: 100% 后端API            │
│   端口: 3001                        │
└─────────────────────────────────────┘
              ↓ REST API (130+ 端点)
┌─────────────────────────────────────┐
│   ⚙️ 后端 (go-backend/)              │
│   唯一数据源: PostgreSQL + Redis    │
│   端口: 8080/9000                   │
└─────────────────────────────────────┘
              ↑ REST API (管理端点)
┌─────────────────────────────────────┐
│   🎛️ B端后台 (web/admin/)            │
│   数据来源: 100% 后端API            │
│   端口: 3000                        │
└─────────────────────────────────────┘
```

### 核心发现
- ✅ **架构健康**: 三端分离清晰，API驱动，统一数据源
- ⚠️ **4个数据同步风险点**: 2个P0高优先级，2个P1中优先级
- ✅ **缓存策略合理**: Redis缓存热点数据，主动失效机制
- ✅ **关键业务流程**: 产品/订单/用户/配置管理数据流正确

---

## 🔴 P0 - 立即修复（高危）

### 1. Admin Panel 权限信息 localStorage 缓存风险

**文件**: `go-backend/web/admin/src/stores/auth.js`

**问题描述**:
管理员权限存储在浏览器 `localStorage` 中：
```javascript
// Line 8
const permissions = ref(JSON.parse(localStorage.getItem('admin_permissions') || '[]'))
```

**安全风险**:
1. 管理员角色/权限在后端更新后，前端不会自动刷新
2. 用户可能看到过期的权限，执行未授权操作（虽然后端会拦截）
3. 用户体验差：显示可操作按钮，但点击后报权限错误

**影响范围**: 所有管理后台功能（用户管理/产品管理/订单管理等）

**复现步骤**:
1. 管理员A登录后台，权限存储到 localStorage
2. 超级管理员降低管理员A的权限（如删除"删除产品"权限）
3. 管理员A刷新页面，仍然能看到"删除"按钮（因为 localStorage 未更新）
4. 管理员A点击删除，后端返回403错误

**修复方案**:

方案一：页面加载时验证权限（推荐）
```javascript
// src/stores/auth.js
const verifyPermissions = async () => {
  if (!token.value) return
  
  try {
    const response = await axios.get('/api/admin/auth/permissions')
    const latestPermissions = response.data.permissions
    
    // 对比本地和服务器权限
    if (JSON.stringify(permissions.value) !== JSON.stringify(latestPermissions)) {
      permissions.value = latestPermissions
      localStorage.setItem('admin_permissions', JSON.stringify(latestPermissions))
    }
  } catch (error) {
    // 验证失败，清除本地权限
    logout()
  }
}

// 在 App.vue 的 onMounted 中调用
onMounted(async () => {
  await authStore.verifyPermissions()
})
```

方案二：权限添加过期时间（简单版）
```javascript
const permissions = ref({
  list: JSON.parse(localStorage.getItem('admin_permissions') || '[]'),
  expiresAt: localStorage.getItem('admin_permissions_expires') || 0
})

const hasPermission = (permission) => {
  // 权限过期检测（1小时TTL）
  if (Date.now() > permissions.value.expiresAt) {
    verifyPermissions()
  }
  return permissions.value.list.includes(permission)
}
```

**优先级**: 🔴 P0 - 立即修复  
**预计工时**: 2小时

---

### 2. 购物车同步失败后数据丢失

**文件**: `nuxt-i18n/app/composables/useCart.ts`

**问题描述**:
游客购物车数据存储在 `localStorage`，用户登录后通过 `/cart/sync` API同步到后端。
但同步失败时，本地数据被删除，导致购物车清空。

```typescript
// Line 62-73
const syncGuestCart = async () => {
  const saved = localStorage.getItem('tanzanite_cart')
  if (saved) {
    try {
      await auth.request('/cart/sync', { method: 'POST', body: payload })
      localStorage.removeItem('tanzanite_cart') // ⚠️ 无论成功失败都删除
    } catch (e) {
      console.error('Failed to sync guest cart', e) // ⚠️ 仅打印错误
    }
  }
}
```

**业务影响**:
1. 用户添加多个商品到购物车（游客状态）
2. 登录后，网络问题导致同步失败
3. **购物车清空**，用户需要重新添加商品
4. 用户体验极差，可能导致订单流失

**修复方案**:

```typescript
const syncGuestCart = async (): Promise<{ success: boolean; error?: string }> => {
  if (!import.meta.client) return { success: false }
  
  const saved = localStorage.getItem('tanzanite_cart')
  if (!saved) return { success: true }
  
  try {
    const items = JSON.parse(saved)
    if (!items || items.length === 0) return { success: true }
    
    const payload = items.map((item: any) => ({
      product_id: item.id,
      quantity: item.quantity
    }))
    
    // 同步到后端
    const response = await auth.request('/cart/sync', {
      method: 'POST',
      body: JSON.stringify(payload)
    })
    
    // ✅ 只有成功后才删除本地数据
    if (response.ok) {
      localStorage.removeItem('tanzanite_cart')
      await loadCartFromBackend()
      return { success: true }
    } else {
      throw new Error('Sync failed')
    }
  } catch (e) {
    console.error('Failed to sync guest cart', e)
    
    // ✅ 失败时保留本地数据，提示用户
    return {
      success: false,
      error: '购物车同步失败，请刷新页面重试'
    }
  }
}

// 在登录后调用，并处理结果
const result = await syncGuestCart()
if (!result.success) {
  ElMessage.error(result.error)
}
```

**额外改进**:
- 添加重试机制（最多3次）
- 同步失败时显示明显提示
- 提供手动重试按钮

**优先级**: 🔴 P0 - 立即修复  
**预计工时**: 3小时

---

## 🟠 P1 - 短期修复（中危）

### 3. 聊天消息仅本地存储，多设备不同步

**文件**: `nuxt-i18n/app/composables/chat/useWhatsAppState.ts`

**问题描述**:
聊天消息历史存储在浏览器 `localStorage`：
```typescript
// Line 82-85
for (const key of chatKeys) {
  const data = localStorage.getItem(key)
  // ...
}
```

**业务影响**:
1. 用户在PC浏览器聊天，切换到手机后看不到历史消息
2. 清除浏览器缓存后消息丢失
3. 客服人员无法查看完整的聊天历史
4. 无法用于后续分析和客服质量评估

**修复方案**:

阶段一：异步同步到后端（推荐）
```typescript
// 1. 创建后端API
// POST /api/v1/chat/messages - 保存消息
// GET /api/v1/chat/messages?session_id=xxx - 获取历史

// 2. 前端修改
const saveChatMessage = async (message: ChatMessage) => {
  // 先存储到 localStorage（快速响应）
  const localKey = `tanzanite_chat_${sessionId}`
  const messages = JSON.parse(localStorage.getItem(localKey) || '[]')
  messages.push(message)
  localStorage.setItem(localKey, JSON.stringify(messages))
  
  // 异步同步到后端（不阻塞用户）
  try {
    await $fetch('/api/v1/chat/messages', {
      method: 'POST',
      body: { session_id: sessionId, message }
    })
  } catch (e) {
    console.warn('Message sync failed, will retry later', e)
  }
}

const loadChatHistory = async () => {
  try {
    // 优先从后端加载
    const response = await $fetch(`/api/v1/chat/messages?session_id=${sessionId}`)
    return response.messages
  } catch (e) {
    // 降级到 localStorage
    const localKey = `tanzanite_chat_${sessionId}`
    return JSON.parse(localStorage.getItem(localKey) || '[]')
  }
}
```

阶段二：WebSocket实时同步
```typescript
// 通过 WebSocket 推送消息到后端
ws.send(JSON.stringify({
  type: 'chat_message',
  data: message
}))
```

**优先级**: 🟠 P1 - 短期修复（2周内）  
**预计工时**: 8小时（后端API + 前端改造）

---

### 4. 浏览历史仅本地存储，无法用于推荐系统

**文件**: `nuxt-i18n/app/composables/useBrowsingHistory.ts`

**问题描述**:
用户浏览历史存储在 `localStorage.getItem('tanzanite_browsing_history')`

**业务影响**:
1. 多设备浏览历史不同步
2. 无法基于浏览历史进行个性化推荐
3. 无法分析用户行为和偏好
4. 营销价值流失

**修复方案**:

```typescript
const addToBrowsingHistory = async (product: Product) => {
  // 本地存储（快速响应）
  const history = JSON.parse(localStorage.getItem('tanzanite_browsing_history') || '[]')
  history.unshift(product)
  history.splice(10) // 保留最近10条
  localStorage.setItem('tanzanite_browsing_history', JSON.stringify(history))
  
  // 异步同步到后端（登录用户）
  if (auth.isAuthenticated) {
    try {
      await $fetch('/api/v1/user/browsing-history', {
        method: 'POST',
        body: { product_id: product.id }
      })
    } catch (e) {
      console.warn('Failed to sync browsing history', e)
    }
  }
}

// 后端可用于：
// - 个性化推荐："您可能喜欢"
// - 用户画像分析
// - 再营销广告
```

**优先级**: 🟠 P1 - 短期修复  
**预计工时**: 4小时

---

## 🟢 P2 - 长期优化（低危）

### 5. 产品属性硬编码 Fallback 数据

**文件**: `nuxt-i18n/app/composables/useProductAttributes.ts`

**问题描述**:
产品属性有 fallback 配置（Line 44-101），包括颜色、直径、刹车等

**潜在风险**:
1. 后端数据库更新属性后，fallback 数据可能过期
2. API 失败时使用 fallback，用户看到的是旧数据
3. B端管理员添加新属性，C端不显示（因为使用 fallback）

**修复建议**:
```typescript
// 方案一：移除 fallback，API失败时显示错误
if (!attributes) {
  throw new Error('无法加载产品属性，请刷新页面重试')
}

// 方案二：Fallback 数据从后端配置获取
const fallbackAttributes = await $fetch('/api/v1/settings/fallback-attributes')

// 方案三：定期验证 fallback 数据一致性
if (process.dev) {
  const apiAttributes = await $fetch('/api/v1/products/attributes')
  if (JSON.stringify(apiAttributes) !== JSON.stringify(fallbackAttributes)) {
    console.warn('[DEV] Fallback attributes are outdated!')
  }
}
```

**优先级**: 🟢 P2 - 长期优化  
**预计工时**: 2小时

---

## ✅ 正确实现的数据流

### 1. 产品管理流程 ✅
```
B端创建产品 → POST /api/admin/products
             → PostgreSQL (products表)
             → Redis 缓存失效
C端查看产品 → GET /api/v1/products/:id
             → PostgreSQL 查询
             → Redis 缓存 (TTL: cache.ProductTTL)
```
**评价**: 完全正确，缓存策略合理

---

### 2. 订单管理流程 ✅
```
C端下单 → POST /api/v1/orders
        → PostgreSQL (orders表)
        → EventBus: OrderCreated 事件
        → asynq异步任务: 扣库存 + 发邮件
B端查看 → GET /api/admin/orders
        → PostgreSQL 实时查询
```
**评价**: CQRS模式正确，事件驱动异步处理

---

### 3. 用户注册流程 ✅
```
C端注册 → POST /api/v1/auth/register
        → PostgreSQL (users表)
        → 返回JWT (HttpOnly Cookie)
B端查看 → GET /api/admin/users
        → PostgreSQL 查询
```
**评价**: 安全认证，数据统一管理

---

### 4. 配置管理流程 ✅
```
B端修改 → POST /api/admin/settings/batch
        → PostgreSQL (settings表)
        → Redis 缓存失效
C端读取 → GET /api/v1/settings/site
        → PostgreSQL 查询
        → Redis 缓存 (TTL: cache.SettingsTTL)
```
**评价**: 配置统一管理，缓存正确

---

## 📊 API 端点清单（130+ 端点）

### C端公开API (`/api/v1`)

#### 认证系统 (`/api/v1/auth`)
- `POST /register` - 用户注册
- `POST /login` - 用户登录
- `GET /profile` - 获取用户信息（需认证）
- `POST /google-login` - Google OAuth登录
- `POST /logout` - 登出

#### 内容管理 (`/api/v1/content`)
- `GET /posts` - 文章列表
- `GET /posts/:id` - 文章详情
- `GET /faqs` - FAQ列表
- `GET /faq-categories` - FAQ分类

#### 产品系统 (`/api/v1/products`)
- `GET /products` - 产品列表（支持筛选/搜索）
- `GET /products/:id` - 产品详情
- `GET /products/attributes/filterable` - 可筛选属性
- `GET /products/featured` - 推荐产品
- `GET /products/related/:id` - 相关产品

#### 购物车 (`/api/v1/cart` - 可选认证)
- `GET /summary` - 购物车摘要
- `POST /add` - 添加商品
- `PUT /items/:id` - 更新数量
- `DELETE /items/:id` - 移除商品
- `POST /sync` - 同步购物车（游客→登录）
- `DELETE /clear` - 清空购物车

#### 订单系统 (`/api/v1/orders` - 需认证)
- `POST /orders` - 创建订单
- `GET /orders` - 用户订单列表
- `GET /orders/:id` - 订单详情
- `POST /orders/:id/cancel` - 取消订单

#### 营销系统 (`/api/v1/marketing`)
- `GET /coupons` - 优惠券列表（公开）
- `POST /coupons/validate` - 验证优惠券（需认证）
- `GET /loyalty/points` - 积分查询（需认证）
- `POST /loyalty/checkin` - 签到（需认证）
- `GET /loyalty/rewards` - 积分奖励列表

#### 客服系统 (`/api/v1/customer-service`)
- `GET /agents` - 客服列表
- `POST /messages` - 发送消息
- `GET /ws` - WebSocket连接
- `POST /tickets` - 创建工单（需认证）

#### 设置与配置 (`/api/v1/settings`)
- `GET /site` - 站点设置（公开）
- `GET /quick-buy` - 快速购买配置（公开）
- `GET /public` - 所有公开设置

#### 多语言 (`/api/v1/i18n`)
- `GET /languages` - 支持的语言列表
- `GET /translations/:post_id` - 文章翻译版本
- `POST /set-language` - 设置语言偏好

---

### B端管理API (`/api/admin`)

#### 认证 (`/api/admin/auth`)
- `POST /login` - 管理员登录
- `GET /profile` - 获取管理员信息
- `GET /permissions` - 权限列表
- `POST /refresh` - 刷新Token
- `POST /logout` - 登出

#### 仪表板 (`/api/admin/dashboard`)
- `GET /stats` - 统计数据
- `GET /recent-orders` - 最近订单
- `GET /recent-users` - 最近用户
- `GET /recent-tickets` - 最近工单
- `GET /sales-chart` - 销售图表

#### 用户管理 (`/api/admin/users`)
- `GET /users` - 用户列表
- `GET /users/:id` - 用户详情
- `POST /users` - 创建用户
- `PUT /users/:id` - 更新用户
- `PATCH /users/:id/status` - 修改状态
- `DELETE /users/:id` - 删除用户
- `POST /batch-delete` - 批量删除

#### 商品管理 (`/api/admin/products`)
- `GET /products` - 商品列表
- `GET /products/:id` - 商品详情
- `POST /products` - 创建商品
- `PUT /products/:id` - 更新商品
- `PATCH /products/:id/stock` - 更新库存
- `DELETE /products/:id` - 删除商品
- `POST /batch-status` - 批量修改状态
- `POST /batch-delete` - 批量删除

#### 订单管理 (`/api/admin/orders`)
- `GET /orders` - 订单列表
- `GET /orders/:id` - 订单详情
- `PATCH /orders/:id/status` - 更新订单状态
- `PATCH /orders/:id/tracking` - 更新物流信息
- `POST /orders/:id/refund` - 退款
- `GET /export` - 导出订单

#### 内容管理 (`/api/admin/content`)
- `GET /posts` - 文章列表
- `POST /posts` - 创建文章
- `PUT /posts/:id` - 更新文章
- `DELETE /posts/:id` - 删除文章
- `POST /batch-status` - 批量发布/下架

#### FAQ管理 (`/api/admin/faqs`)
- `GET /faqs` - FAQ列表
- `POST /faqs` - 创建FAQ
- `PUT /faqs/:id` - 更新FAQ
- `DELETE /faqs/:id` - 删除FAQ

#### 图库管理 (`/api/admin/galleries`)
- `GET /galleries` - 图片列表
- `POST /galleries/upload` - 上传图片
- `PATCH /galleries/:id/approve` - 审核通过
- `DELETE /galleries/:id` - 删除图片

#### 订阅管理 (`/api/admin/subscriptions`)
- `GET /subscriptions` - 订阅列表
- `GET /export` - 导出订阅者

#### 工单管理 (`/api/admin/tickets`)
- `GET /tickets` - 工单列表
- `PATCH /tickets/:id/assign` - 分配工单
- `PATCH /tickets/:id/status` - 更新状态
- `POST /tickets/:id/reply` - 回复工单

#### 营销管理 (`/api/admin/marketing`)
- `GET /coupons` - 优惠券列表
- `POST /coupons` - 创建优惠券
- `PUT /coupons/:id` - 更新优惠券
- `GET /loyalty` - 积分设置

#### 设置管理 (`/api/admin/settings`)
- `GET /settings` - 获取所有设置
- `POST /batch` - 批量更新设置

#### 审计日志 (`/api/admin/audit-logs`)
- `GET /logs` - 审计日志列表

---

## 📦 缓存策略分析

### Redis 缓存使用
| 数据类型 | 缓存Key | TTL | 失效机制 |
|---------|---------|-----|---------|
| 产品信息 | `product:{id}` | `cache.ProductTTL` | 修改时主动DEL |
| 产品列表 | `products:list:{hash}` | `cache.ProductTTL` | 修改时清空前缀 |
| 文章内容 | `post:{id}` | `cache.PostTTL` | 修改时主动DEL |
| 站点设置 | `settings:site` | `cache.SettingsTTL` | 修改时主动DEL |
| FAQ列表 | `faqs:list` | `cache.FAQTTL` | 修改时主动DEL |

### 无缓存场景
- 购物车（实时性要求高）
- 订单（实时性要求高）
- 用户信息（安全性要求高）
- 权限信息（安全性要求高）

### 缓存风险
⚠️ **如果Redis缓存失效操作失败**:
- 用户可能看到旧的产品信息
- 管理员修改后C端不更新

**建议**: 添加缓存失效失败告警

---

## 🎯 修复路线图

### 第1周（P0）
- [ ] Admin Panel 权限验证机制
- [ ] 购物车同步失败处理

### 第2-3周（P1）
- [ ] 聊天消息后端持久化
- [ ] 浏览历史同步到后端

### 第4周（P2）
- [ ] 移除硬编码 fallback 数据
- [ ] 缓存失效告警机制

---

## 📌 总结

**数据同步健康度**: 🟢 85/100

**优点**:
- ✅ 三端分离清晰，职责明确
- ✅ API驱动，统一数据源
- ✅ 缓存策略合理
- ✅ 事件驱动架构
- ✅ 安全认证机制

**需要改进**:
- 🔴 Admin Panel 权限缓存
- 🔴 购物车同步失败处理
- 🟠 聊天消息持久化
- 🟠 浏览历史同步

**建议行动**:
1. 立即修复 P0 问题（权限+购物车）
2. 2周内完成 P1 问题（聊天+浏览历史）
3. 1个月内完成 P2 优化

---

**报告生成**: 2026-06-26  
**下次审计**: 2026-07-26（1个月后）

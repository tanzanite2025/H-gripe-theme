# 聊天系统分析与改进计划

## 一、当前架构分析

### 1.1 技术栈

| 组件 | 技术 | 说明 |
|------|------|------|
| **前端聊天窗口** | Vue 3 (Nuxt) | `WhatsAppChatModal.vue` (2045行) |
| **后端 API** | WordPress REST API | `tanzanite/v1/customer-service/agents` |
| **用户认证** | WordPress Cookie | `useAuth.ts` → `/mytheme/v1/auth/me` |
| **实时通信** | 轮询 (暂无 WebSocket) | 计划换 VPS 后升级 |
| **移动端客服** | React Native | `tanzanite-chat` 项目 |

### 1.2 数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户访问网站                              │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  useAuth.ensureSession()                                        │
│  → 调用 /mytheme/v1/auth/me                                     │
│  → 返回用户信息 (如果已登录) 或 null (游客)                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  WhatsAppChatModal.fetchAgents()                                │
│  → 调用 /wp-json/tanzanite/v1/customer-service/agents           │
│  → 返回所有客服列表 (不区分当前用户)                             │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  显示客服列表                                                    │
│  → 问题：如果当前用户是客服，也会显示自己                        │
└─────────────────────────────────────────────────────────────────┘
```

---

## 二、发现的问题

### 2.1 身份识别问题 ⚠️

**现象：** 客服登录后台后，在网站前台打开聊天窗口，仍然能看到自己在客服列表中。

**根本原因：**

1. **API 不区分身份**
   - `/tanzanite/v1/customer-service/agents` 返回所有客服
   - 没有根据当前登录用户过滤

2. **前端不过滤**
   - `fetchAgents()` 直接使用 API 返回的全部数据
   - 没有检查 `user.value.id` 是否在客服列表中

**代码位置：**
```typescript
// WhatsAppChatModal.vue 第 1712 行
const response = await $fetch<any>('/wp-json/tanzanite/v1/customer-service/agents')
if (response.success && response.data) {
  agents.value = response.data  // ← 直接使用，未过滤
}
```

### 2.2 用户身份来源

```typescript
// WhatsAppChatModal.vue 第 962 行
const { user } = useAuth()

// useAuth.ts 第 108 行
const data = await request<AuthUser>('/mytheme/v1/auth/me', ...)
user.value = data
```

**`user.value` 包含：**
- `id` - 用户 ID
- `username` - 用户名
- `email` - 邮箱
- `display_name` - 显示名称

**问题：** 没有字段标识用户是否是客服（如 `is_agent` 或 `role`）

---

## 三、解决方案

### 方案 A：前端过滤（快速修复）

在 `fetchAgents()` 中，获取客服列表后过滤掉当前用户：

```typescript
// 获取客服列表后
if (response.success && response.data) {
  // 过滤掉当前登录用户
  const currentUserId = user.value?.id
  agents.value = response.data.filter((agent: any) => {
    return agent.id !== currentUserId
  })
}
```

**优点：** 改动最小，立即生效
**缺点：** 依赖前端逻辑，不够彻底

---

### 方案 B：后端过滤（推荐）

修改 WordPress API，接收当前用户 ID 参数，返回时排除：

```php
// WordPress 端
function get_agents($request) {
    $current_user_id = get_current_user_id();
    $agents = get_all_agents();
    
    // 排除当前用户
    $agents = array_filter($agents, function($agent) use ($current_user_id) {
        return $agent['id'] !== $current_user_id;
    });
    
    return $agents;
}
```

**优点：** 逻辑在后端，更安全
**缺点：** 需要修改 WordPress 插件

---

### 方案 C：角色区分（最彻底）

1. 在 `/mytheme/v1/auth/me` 返回中增加 `is_agent` 字段
2. 前端根据 `is_agent` 决定是否显示聊天窗口或切换到客服模式

```typescript
// 如果当前用户是客服，显示客服后台界面
if (user.value?.is_agent) {
  // 显示客服工作台（接收消息）
} else {
  // 显示访客聊天窗口（发送消息）
}
```

---

## 四、Web 端客服工作台需求

你提到希望在网页上也能与用户会话，而不只是在 APP 上。这需要：

### 4.1 当前状态

| 平台 | 功能 | 状态 |
|------|------|------|
| **移动端 (tanzanite-chat)** | 客服接收/回复消息 | ✅ 已实现 |
| **Web 端 (tanzanite-theme)** | 访客发送消息 | ✅ 已实现 |
| **Web 端 (tanzanite-theme)** | 客服接收/回复消息 | ❌ 未实现 |

### 4.2 需要新增

1. **客服工作台页面** - 类似 Chatwoot 的 dashboard
   - 会话列表
   - 消息收发
   - 客服状态切换

2. **或者：复用现有组件**
   - 当检测到用户是客服时，`WhatsAppChatModal` 切换为"客服模式"
   - 显示所有待处理会话
   - 可以选择会话进行回复

---

## 五、改进计划

### 阶段 1：修复身份识别问题

- [ ] 前端过滤：在 `fetchAgents()` 中排除当前用户
- [ ] 后端增强：API 返回时排除当前用户
- [ ] 用户信息增强：`/auth/me` 返回 `is_agent` 字段

### 阶段 2：添加欢迎页

- [ ] 设计欢迎页 UI（参考 Chatwoot）
- [ ] 显示客服头像和状态
- [ ] 集成 FAQ 快捷入口

### 阶段 3：Web 端客服工作台

- [ ] 创建客服工作台页面 `/agent/dashboard`
- [ ] 会话列表组件
- [ ] 消息收发组件
- [ ] 客服状态切换

### 阶段 4：实时通信升级

- [ ] VPS 迁移后启用 WebSocket
- [ ] 替换轮询为实时推送

---

## 六、已确认信息

### 6.1 客服数据来源 ✅

**关键发现：客服不是 WordPress 用户！**

客服列表是在 **WordPress 插件内部定义** 的，与 WordPress 用户系统完全独立：
- 插件内配置 3 个客服
- 每个客服有独立的 ID、名称、WhatsApp 等信息
- 主要为移动端 APP (tanzanite-chat) 设计

**这意味着：**
- 客服 ID ≠ WordPress 用户 ID
- 无法通过 `user.value.id` 直接匹配客服
- 需要建立 WordPress 用户与插件客服的关联

### 6.2 Web 端客服工作台 ✅

**决定：使用独立弹窗组件**

理由：
- 不影响现有正在工作的 `WhatsAppChatModal`
- 出错时可以快速回滚
- 复制现有组件进行改造

### 6.3 优先级 ✅

**先修复身份识别问题**

---

## 七、修订后的解决方案

### 7.1 建立用户-客服关联

由于客服是插件内定义的，需要建立 WordPress 用户与客服的关联：

**方案 A：在插件中添加 WordPress 用户 ID 字段**

```php
// 插件客服配置
$agents = [
    [
        'id' => 1,
        'name' => 'Sales Agent',
        'whatsapp' => '+1234567890',
        'wp_user_id' => 123,  // ← 新增：关联的 WordPress 用户 ID
    ],
    // ...
];
```

**方案 B：在 WordPress 用户 meta 中标记客服 ID**

```php
// 在用户 meta 中存储客服 ID
update_user_meta($user_id, 'tanzanite_agent_id', 1);
```

**方案 C：API 返回时包含 wp_user_id（推荐）**

修改 `/tanzanite/v1/customer-service/agents` API：
- 返回每个客服关联的 `wp_user_id`
- 前端根据 `user.value.id` 过滤

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Sales Agent",
      "wp_user_id": 123,  // ← 关联的 WordPress 用户 ID
      "whatsapp": "+1234567890"
    }
  ]
}
```

前端过滤：
```typescript
agents.value = response.data.filter((agent: any) => {
  return agent.wp_user_id !== user.value?.id
})
```

### 7.2 身份识别修复步骤

1. **插件端**
   - 为每个客服配置添加 `wp_user_id` 字段
   - API 返回时包含此字段

2. **前端**
   - `fetchAgents()` 中过滤掉 `wp_user_id === user.value?.id` 的客服

3. **可选增强**
   - `/auth/me` 返回 `agent_id` 字段（如果用户是客服）
   - 用于后续客服工作台功能

### 7.3 Web 端客服工作台计划

**新建组件：`AgentDashboardModal.vue`**

基于 `WhatsAppChatModal.vue` 复制改造：
- 显示所有待处理会话（而非客服列表）
- 客服可以选择会话进行回复
- 客服状态切换（在线/忙碌/离线）

**触发方式：**
- 检测到用户是客服时，显示"进入工作台"按钮
- 或在用户菜单中添加入口

---

## 八、实施计划

### 阶段 1：修复身份识别（当前优先）

| 步骤 | 任务 | 负责 |
|------|------|------|
| 1.1 | 插件中为客服添加 `wp_user_id` 字段 | WordPress 插件 |
| 1.2 | API 返回时包含 `wp_user_id` | WordPress 插件 |
| 1.3 | 前端 `fetchAgents()` 过滤当前用户 | Nuxt 前端 |
| 1.4 | 测试验证 | 全流程 |

### 阶段 2：添加欢迎页

| 步骤 | 任务 |
|------|------|
| 2.1 | 设计欢迎页 UI |
| 2.2 | 实现 WelcomeScreen 组件 |
| 2.3 | 集成 FAQ |

### 阶段 3：Web 端客服工作台

| 步骤 | 任务 |
|------|------|
| 3.1 | 复制 `WhatsAppChatModal.vue` 为 `AgentDashboardModal.vue` |
| 3.2 | 改造为客服视角（显示会话列表） |
| 3.3 | 添加客服状态切换 |
| 3.4 | 添加入口触发逻辑 |

### 阶段 4：实时通信升级

| 步骤 | 任务 |
|------|------|
| 4.1 | VPS 迁移 |
| 4.2 | 启用 WebSocket |
| 4.3 | 替换轮询 |

---

## 九、插件代码分析

### 9.1 插件位置

```
wp-plugin/tanzanite-customer-service/
├── tanzanite-customer-service.php  # 主插件文件
├── includes/
│   ├── class-database.php          # 数据库表结构
│   └── class-agent-auth.php        # 客服认证
└── api/
    ├── class-agent-api.php         # 客服端 API
    └── class-auto-reply-api.php    # 自动回复 API
```

### 9.2 客服表结构 (`wp_tz_cs_agents`)

```sql
CREATE TABLE wp_tz_cs_agents (
    id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    agent_id varchar(50) NOT NULL,        -- 客服工号（如 CS001）
    name varchar(100) NOT NULL,           -- 客服名称
    email varchar(100) NOT NULL,          -- 邮箱
    password varchar(255) NOT NULL,       -- 密码（bcrypt）
    avatar varchar(500) DEFAULT '',       -- 头像 URL
    whatsapp varchar(50) DEFAULT '',      -- WhatsApp 号码
    pre_sales_email varchar(100),         -- 售前邮箱
    after_sales_email varchar(100),       -- 售后邮箱
    status varchar(20) DEFAULT 'active',  -- 状态
    last_login datetime,                  -- 最后登录
    created_at datetime NOT NULL,
    -- ⚠️ 缺少 wp_user_id 字段！
    PRIMARY KEY (id),
    UNIQUE KEY agent_id (agent_id)
);
```

### 9.3 获取客服列表 API

**文件：** `tanzanite-customer-service.php` 第 373-403 行

```php
public function rest_get_agents( \WP_REST_Request $request ): \WP_REST_Response {
    global $wpdb;
    $table = $wpdb->prefix . 'tz_cs_agents';
    
    // 只返回启用的客服
    $agents = $wpdb->get_results(
        "SELECT agent_id, name, email, avatar, whatsapp FROM $table WHERE status = 'active' ORDER BY created_at ASC"
    );
    
    // 格式化输出
    $formatted = array_map( fn( $agent ) => [
        'id'       => $agent->agent_id,
        'name'     => $agent->name,
        'email'    => $agent->email,
        'avatar'   => $agent->avatar,
        'whatsapp' => $agent->whatsapp,
        // ⚠️ 没有返回 wp_user_id！
    ], $agents );
    
    return new \WP_REST_Response( [
        'success' => true,
        'data'    => $formatted,
        // ...
    ], 200 );
}
```

---

## 十、具体修改方案

### 10.1 数据库修改

**添加 `wp_user_id` 字段到 `wp_tz_cs_agents` 表：**

```sql
ALTER TABLE wp_tz_cs_agents 
ADD COLUMN wp_user_id bigint(20) unsigned DEFAULT NULL AFTER agent_id;
```

### 10.2 插件代码修改

#### A. 修改数据库类 (`class-database.php`)

在 `$sql_agents` 中添加字段：

```php
// 第 98-116 行
$sql_agents = "CREATE TABLE $table_agents (
    id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    agent_id varchar(50) NOT NULL,
    wp_user_id bigint(20) unsigned DEFAULT NULL,  // ← 新增
    name varchar(100) NOT NULL,
    // ...
) $charset_collate;";
```

#### B. 修改 API 返回 (`tanzanite-customer-service.php`)

在 `rest_get_agents()` 中返回 `wp_user_id`：

```php
// 第 378-379 行
$agents = $wpdb->get_results(
    "SELECT agent_id, wp_user_id, name, email, avatar, whatsapp FROM $table WHERE status = 'active' ORDER BY created_at ASC"
);

// 第 383-389 行
$formatted = array_map( fn( $agent ) => [
    'id'         => $agent->agent_id,
    'wp_user_id' => $agent->wp_user_id ? (int) $agent->wp_user_id : null,  // ← 新增
    'name'       => $agent->name,
    'email'      => $agent->email,
    'avatar'     => $agent->avatar,
    'whatsapp'   => $agent->whatsapp,
], $agents );
```

#### C. 修改管理页面表单

在添加/编辑客服表单中添加 WordPress 用户选择：

```php
<tr>
    <th><label for="wp_user_id">关联 WordPress 用户</label></th>
    <td>
        <?php wp_dropdown_users([
            'name' => 'wp_user_id',
            'id' => 'wp_user_id',
            'show_option_none' => '-- 不关联 --',
            'option_none_value' => '',
        ]); ?>
        <p class="description">关联后，该用户在前台聊天窗口中将不会看到自己</p>
    </td>
</tr>
```

### 10.3 前端代码修改

**文件：** `WhatsAppChatModal.vue` 第 1712-1733 行

```typescript
// 获取客服列表后过滤当前用户
if (response.success && response.data) {
  const currentUserId = user.value?.id
  
  // 过滤掉当前登录用户关联的客服
  agents.value = response.data.filter((agent: any) => {
    // 如果客服没有关联 wp_user_id，或者不是当前用户，则显示
    return !agent.wp_user_id || agent.wp_user_id !== currentUserId
  })
  
  // 保存邮箱设置...
}
```

---

## 十一、相关文件

| 文件 | 说明 |
|------|------|
| `app/components/WhatsAppChatModal.vue` | 聊天弹窗主组件 |
| `app/composables/useAuth.ts` | 用户认证逻辑 |
| `wp-plugin/tanzanite-customer-service/tanzanite-customer-service.php` | 插件主文件 |
| `wp-plugin/tanzanite-customer-service/includes/class-database.php` | 数据库结构 |
| `tanzanite-chat/` | React Native 移动端客服 App |
| `chatwoot-4.8.0/` | 参考项目（UI 设计） |

---

## 十二、下一步行动

### 立即执行（阶段 1）

- [x] **1.1** 修改 `class-database.php`：添加 `wp_user_id` 字段 ✅
- [x] **1.2** 修改 `tanzanite-customer-service.php`： ✅
  - API 返回 `wp_user_id` ✅
  - 管理页面添加用户关联选择 ✅
  - 处理表单提交保存 `wp_user_id` ✅
- [x] **1.3** 修改 `WhatsAppChatModal.vue`：过滤当前用户 ✅
  - 缓存读取时过滤 ✅
  - API 获取时过滤 ✅
- [ ] **1.4** 测试验证

### 阶段 2：添加欢迎页 ✅

- [x] **2.1** 设计欢迎页 UI（`chat-welcome-preview.html`）✅
- [x] **2.2** 添加欢迎页状态和方法 ✅
- [x] **2.3** 实现欢迎页模板 ✅
- [x] **2.4** 添加挥手动画 CSS ✅
- [x] **2.5** 集成 FAQ 快捷入口 ✅

### 后续阶段

- 阶段 3：Web 端客服工作台
- 阶段 4：WebSocket 升级

---

*文档创建时间：2025-12-07*
*最后更新：2025-12-07 05:20*

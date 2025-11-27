# User Feedback / Comment System Design

## 1. Overall Goals

- 提供一个可复用的「用户留言组件」：
  - 上方展示留言列表。
  - 下方是留言表单：`name + email + message`。
- 支持 **多页面复用**：
  - 每个页面有独立的留言「线程」(thread)。
- 留言对所有访客公开可见。
- **必须登录** 才能提交留言（过滤恶意刷屏，记录清楚是哪个账号）。
- 留言采用 **审核制**：
  - 用户提交后先进入 `pending`。
  - 只有后台管理员审核通过的留言才对前台显示。
- 前端组件带有 **搜索框**，可以按内容搜索当前页面的留言。
- 后端实现整合进现有 `tanzanite-setting` 插件，并与其他模块解耦。

---

## 2. Data Model

### 2.1 Database Table: `wp_tanz_feedback`

Using `$wpdb->prefix . 'tanz_feedback'`.

Fields (draft):

- `id` (bigint, PK, auto increment)
- `thread_key` (varchar):
  - 用来区分页面 / 话题，例如：
    - `"support-spoke-calculator"`
    - `"support-warranty"`
    - `"product-12345"`
- `user_id` (bigint, not null):
  - WP 用户 ID，提交人。
- `name` (varchar, nullable):
  - 展示用姓名（可以来自用户 profile 或前端表单）。
- `email` (varchar, nullable):
  - 仅用于后台参考，不在前台公开。
- `content` (text):
  - 留言正文。
- `status` (varchar, enum-like):
  - `pending` | `approved` | `rejected` | `hidden`
- `locale` (varchar, optional):
  - 例如 `"en"`, `"zh-Hans"`，用于多语言区分（可选）。
- `created_at` (datetime)
- `updated_at` (datetime)

---

## 3. Backend: REST API Design (inside tanzanite-setting plugin)

### 3.1 New Controller

- Class: `Tanzanite_REST_Feedback_Controller extends Tanzanite_REST_Controller`
- REST namespace: already `tanzanite/v1` from base class.
- `rest_base = 'feedback'`.

### 3.2 Routes

#### 3.2.1 List Feedback (public)

- **Endpoint**: `GET /wp-json/tanzanite/v1/feedback`
- **Query params**:
  - `thread` (string, required): `thread_key`
  - `page` (int, optional, default 1)
  - `per_page` (int, optional, default 20, max 100)
  - `search` (string, optional): 按内容模糊搜索
  - `status` (string, optional, default `"approved"`):
    - 前台通常固定为 `"approved"`。
    - 后台管理页可请求 `"pending"` / `"rejected"` / `"all"`。
- **Permission**:
  - `permission_callback` = `__return_true` （公开读取）。
- **Response (example)**:
  ```json
  {
    "data": [
      {
        "id": 123,
        "thread_key": "support-spoke-calculator",
        "user_id": 45,
        "name": "Alice",
        "content": "Great tool, but please add ...",
        "status": "approved",
        "created_at": "2025-11-27T06:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 5,
      "total_pages": 1
    }
  }
  ```

#### 3.2.2 Create Feedback (login required, audit mode)

- **Endpoint**: `POST /wp-json/tanzanite/v1/feedback`
- **Headers**:
  - `X-WP-Nonce`（由 WP 注入，Nuxt 通过 SSR/bridge 获取）
- **Body** (JSON):
  ```json
  {
    "thread": "support-spoke-calculator",
    "content": "text...",
    "name": "Optional, override display name",
    "email": "Optional, if needed"
  }
  ```
- **Permission**:
  - `permission_callback` = `'is_user_logged_in'`
    => **必须登录** 才能提交。
- **Server logic**:
  - `$user_id = get_current_user_id()`，如果 0 则返回 401。
  - 基于 body 写入 `tanz_feedback` 表：
    - `status = 'pending'` 默认。
    - `name` / `email` 可以优先用传入值；为空时可退回到 `user_display_name` / `user_email`。
  - 审核逻辑在后台进行，此处不自动 `approved`。
- **Response (example)**:
  ```json
  {
    "message": "Feedback submitted, pending review.",
    "comment": {
      "id": 456,
      "status": "pending"
    }
  }
  ```

#### 3.2.3 Admin: Update Status / Delete

- **Endpoint**: `POST /wp-json/tanzanite/v1/feedback/{id}/status`
- **Body**:
  ```json
  { "status": "approved" }
  ```
- **Permission**:
  - `permission_callback` = `$this->permission_callback('manage_options', true)`
    => 仅管理员 / 有相应 capability 的用户可操作。
- 类似方式提供批量删除或单条删除（不一定立刻需要）。

#### 3.2.4 Optional: Eligibility Check Endpoint

- **Endpoint**: `GET /wp-json/tanzanite/v1/feedback/eligibility`
- **Query**:
  - `thread` (string) – 可选，仅做日志或未来限制用途。
- **Permission**:
  - 公开：
    - 如果已登录：返回 `can_post: true`。
    - 未登录：返回 `can_post: false`，并说明原因。
- **Response example**:
  ```json
  {
    "can_post": true,
    "reason": null,
    "logged_in": true
  }
  ```

前端可以用这个接口，提前决定是否展示表单。

---

## 4. Frontend: Reusable Vue Component

### 4.1 Main Component: `UserFeedbackThread.vue`

**Props**:

- `threadKey: string` (**required**)
- UI 文案：
  - `title?: string`
  - `subtitle?: string`
- 行为：
  - `showSearch?: boolean` (default `true`)

**Internal State**:

- `comments`：当前线程的留言列表
- `loadingList` / `loadingSubmit` / `error`
- `searchQuery`
- `eligibility`：
  - `{ canPost: boolean; loggedIn: boolean; reason?: string }`

**Lifecycle**:

1. `onMounted`:
   - `GET /feedback?thread=threadKey&status=approved`
   - `GET /feedback/eligibility?thread=threadKey`
2. 显示 UI：
   - 顶部标题 & 介绍文案。
   - 搜索框（绑定 `searchQuery`）。
   - 留言列表区域（过滤或服务端搜索）。
   - 底部留言表单：
     - 若 `eligibility.canPost = false` 显示提示文案，如：
       - `Please log in to leave a comment.`
     - 若 `canPost = true` 显示表单。

**Form Submission**:

- `POST /feedback` with `{ thread, content, name, email }`.
- 成功后：
  - 显示 `已提交，等待审核` 提示。
  - 不必立即把该留言加入 `comments`（避免用户误以为已公开）；也可以加一个「半透明 pending 状态」视图，具体可视化后再定。

---

## 5. Search Behaviour

### 5.1 Phase 1: Client-side Search

- 接口只按 `thread` 拉取当前页的 `approved` 留言。
- 组件内部用 `computed`：
  - `filteredComments = comments.filter(c => c.content.includes(searchQuery))`
- 实现简单，适合留言数量不大的阶段。

### 5.2 Phase 2: Server-side Search (Optional)

- 若将来留言很多，扩展为：
  - 搜索时调用 `GET /feedback?thread=...&search=...`。
  - 加 `debounce`，减轻服务器压力。

---

## 6. Authentication & Security Notes

- **必须登录** 才能 `POST /feedback`：
  - 基于 WP 的 `is_user_logged_in()` + `X-WP-Nonce`。
- 全部权限判断在后端执行，前端只做 UI 层面的引导。
- 不接入 loyalty / 积分系统：
  - 留言系统和 Loyalty 解耦，减少相互影响。
- 将来若需要“只允许 Bronze 以上会员”：
  - 可以在 `POST /feedback` 内部附加一段 tier 判断逻辑，不改变整体架构。

---

## 7. Frontend Placement Examples

Example usage in different pages:

```vue
<!-- Support: Spoke Calculator page -->
<UserFeedbackThread
  threadKey="support-spoke-calculator"
  title="Share your feedback about the Spoke Calculator"
  :showSearch="true"
/>

<!-- Support: Warranty page -->
<UserFeedbackThread
  threadKey="support-warranty"
  title="Warranty & after-sales feedback"
/>

<!-- Product detail page (dynamic thread) -->
<UserFeedbackThread
  :threadKey="`product-${productId}`"
  :title="`Feedback for ${productName}`"
/>
```

---

## 8. Implementation Steps (For Future Work)

1. **DB Migration**：
   - 在插件激活时创建 `tanz_feedback` 表。
2. **REST Controller**：
   - 新建 `Tanzanite_REST_Feedback_Controller`，实现上述路由。
   - 在 `class-plugin.php` 中注册此控制器。
3. **Nuxt API client / composable**：
   - `useFeedback(threadKey)`：封装 `GET/POST` 调用。
4. **前端组件**：
   - 实现 `UserFeedbackThread.vue` + `FeedbackList.vue` + `FeedbackForm.vue`。
5. **Admin UI (future)**：
   - 在 WP 后台添加 Feedback 管理页面（列表 + 审核）。

---

## 9. Spoke Calculator 专用搜索（草案想法）

> 这一节是扩展设计，**和用户留言系统解耦**，只是记录目前的产品想法，方便后续一起评估。

### 9.1 目标

- 在 Spoke Calculator 页面内增加一个**只针对辐条历史数据的搜索框**：
  - 你维护一张「历史轮组 / 辐条长度」数据表。
  - 搜索框可以按轮径、花鼓型号、辐条长度等关键词快速检索你过往经验数据。
- 使用场景：
  - 设计新轮组时，先查「类似配置以前用过多少 mm」，再用现在的计算器做微调。

### 9.2 数据表（建议草案）

后端可以在 `tanzanite-setting` 插件里新建一张表，例如：

- 表名：`{$wpdb->prefix}tanz_spoke_history`

示例字段（可精简）：

- **基础信息**：
  - `id` (bigint, PK)
    - 仅作为数据库内部主键使用，**不参与计算器侧的匹配逻辑**；
    - 实际匹配会依赖 `hub_brand` / `hub_model` 及相关几何参数.
  - `wheel_type` (varchar)：`front` / `rear` / `pair`
  - `source_type` (varchar)：`measured` / `calculated` / `vendor`（手工测量、计算器结果、厂商数据）
- **关键技术参数（方便搜索 & 回填）**：
  - `rim_brand`, `rim_model`
  - `hub_brand`, `hub_model`
  - `erd_mm`
  - `flange_left_mm`, `flange_right_mm`
  - `pcd_left_mm`, `pcd_right_mm`
  - `center_to_flange_left_mm`, `center_to_flange_right_mm`
  - `spoke_count`
  - `lacing_pattern`（如 `3x`, `2x`, `radial`）
- **结果数据**：
  - `left_length_mm`, `right_length_mm`
  - 可选：`spoke_brand`, `spoke_model`, `nipple_type`
- **元数据**：
  - `created_at`, `updated_at`（可选，用于按时间排序或后续统计分析；不存放客户信息）

索引可以重点放在：`hub_model`（主）、`hub_brand`、`rim_model` 等字段，以支撑「按花鼓型号」为主的搜索方式。

### 9.3 后端 API 草案

保持和 Feedback 系统一样的风格，在 `tanzanite/v1` 下增加一个只读接口，例如：

- `GET /tanzanite/v1/spoke-history`

Query 参数示例：

- `search`：自由文本，主要按 `hub_model` / `hub_brand` / `rim_model` 等字段做模糊匹配.
  - Phase 1 不提供额外过滤条件（如 `wheel_type` / `spoke_count` / 长度区间），保持接口和前端交互尽量简单，全部交给模糊搜索处理.

权限：

- Phase 1 可以先设计为**公开只读**，因为数据本身不敏感；
- 写入新记录（如果以后做「一键保存到历史」）再要求登录。

### 9.4 Nuxt 端集成（草案）

可以新建一个 composable，例如：

- `useSpokeHistory()`：
  - `items`, `loading`, `error`
  - `searchText`
  - `fetchHistory({ search })`

在 `SpokeCalculatorCore.vue` 中：

- 在计算器上方或结果区域旁边增加一个小卡片：
  - 标题：`Search past wheel builds`。
  - 文本输入框：`searchText`。
  - 下方列表展示匹配的历史记录（比如最近 10 条）。
- 点击某一条记录：
  - Phase 1：只展开详细信息（方便你肉眼对照）。
  - 后续可以升级为：**一键回填**当前计算器的输入（ERD、花鼓参数、辐条数、交叉方式等）。

### 9.5 当前共识 / 已回答问题

- **[Q1] 最常用的搜索关键词是什么？**

  - 结论：**主要搜索维度是「花鼓型号」**，其次才是轮径、辐条长度等信息。
  - 启发：在搜索实现和索引设计上，必须确保 `hub_model`（以及 `hub_brand`）是**一等公民**，提供前缀匹配 / 模糊匹配体验。

- **[Q2] 每条历史记录「最低限度必须存」的字段？**

  - 现在 Spoke Calculator 已经使用并依赖的一组字段：
    - `spoke_count`
    - `lacing_pattern`
    - `nipple_type`
    - `erd_mm`
    - `left_flange_distance_mm`
    - `right_flange_distance_mm`
  - 此外，历史记录需要补充：
    - `hub_brand`
    - `hub_model`
  - 系统中已经存在一个 WP 后台，用来「针对性添加商品字段」。建议：
    - 在现有后台表/结构上**追加** `hub_brand`、`hub_model` 字段；
    - Spoke Length History 功能尽量**共用这批字段定义**，保证：
      - 一条历史记录可以**直接映射**到 Spoke Calculator 的输入参数；
      - 未来可以实现「从历史记录一键导入到计算器」.

- **[Q3] 写入方式偏好？（谁来维护这批历史数据）**

  - 期望：
    - **Phase 1 不做日常后台维护界面**，减少重复维护成本；
    - 由你一次性根据约定字段结构，准备完整的历史数据（例如 CSV / SQL），
      然后由我们导入到数据库中作为初始数据集.
  - 后续如果需要频繁更新，再考虑：
    - 在 WP 后台为历史数据做一个轻量的管理 UI；
    - 或者复用现有商品字段管理界面，增加一个「导出到 Spoke Length History」的动作.

- **[Q4] 权限和可见性？**

  - 你的期望是：
    - 这些历史数据不仅仅服务内部/后台，而是**对登录会员也开放**；
    - 也就是说，Spoke Calculator 页面上的「历史搜索」组件，在权限上至少需要：
      - 登录会员可以使用；
      - 是否对未登录访客开放，可以以后再根据产品策略决定（例如仅展示概要或隐藏模块）。

> 备注：上面这部分已经是「设计共识」，后续如果有变动，可以在本节继续补充修订历史.
为了便于复用，建议将「历史长度搜索输入框 + 结果列表」封装为一个独立组件：

在 Spoke Calculator 页面中，作为会员自助查询历史数据的入口；
在客服工作界面（例如客服列表或会话详情）中，也可以复用该组件，用于快速调取历史记录.
当客服或用户在组件中找到合适的历史记录后，如需继续进行具体的长度计算，再统一引导跳转到 Spoke Calculator 主页面进行操作。


#### 9.x 组件化与 WhatsApp 客服弹窗集成

- 系统中已存在一个 WhatsApp 弹窗组件：
  - 在每个客服的 `SHARE PRODUCTS` 分页下，已有一个 `History` 按钮；
  - 当前点击 `History` 后，会弹出一个 HISTORY 弹窗（已有完整的弹窗布局和样式）。

- 为了降低改动风险，本设计约定：
  - **不改动** `SHARE PRODUCTS` 分页的整体结构；
  - **不改动** `History` 按钮本身的行为入口（依然由它打开 HISTORY 弹窗）；
  - **不改动** HISTORY 弹窗的外层容器、位置、遮罩和样式框架.
  - 仅在实现时，将 HISTORY 弹窗的「内部内容区域」替换为本节定义的「Spoke Length History 搜索 + 结果」组件（暂定命名为一个独立前端组件，实际命名可在实现阶段确定）：
  - 这样可复用现有弹窗的 UI 和交互，不需要重新设计样式；
  - 也避免对其它 WhatsApp 功能产生连带影响，将改动范围限制在弹窗内容层.

- 后续如果在客服场景中需要新增能力（例如：根据当前会话的轮组信息预填搜索关键字、记录客服使用历史等），可以在保持该组件公共能力不变的前提下，通过弹窗外层向组件传入参数或配置来扩展，而无需更改整体架构。

> 命名约定（避免混淆）

  - 当前 WhatsApp 弹窗中的 `History` 按钮和 HISTORY 弹窗名称，仅为现有实现中的旧命名.
  - 在接入本节定义的「Spoke Length History 搜索 + 结果」组件时，建议同时统一调整命名，例如：
  - 按钮文案：`Spoke Length History`（或中文「辐条长度记录」/「轮组长度记录」，视语言环境确定）；
  - 弹窗标题：与按钮文案保持一致（例如「Spoke Length History」）。
  - 通过更明确的命名，将该入口与其它历史类功能（如订单历史、聊天记录等）区分开，避免用户和客服在使用时产生混淆.
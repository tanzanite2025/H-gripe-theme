# 事务性邮件与通知系统架构设计规范 (Transactional Email & Notification Architecture)

> **文档状态**：实现基线（SMTP provider 配置、运行时选择、连通性测试和模板版本管理已实现；外部订阅与正式生产放量验收仍属后续阶段）。  
> **最后审计**：2026-09-22  
> **完成度清单**：[`transactional-email-notification-status.md`](./transactional-email-notification-status.md)  
> **适用范围**：后端 Go 邮件与 Outbox 服务、邮件模板引擎、售后（RMA）流转自动化通知、订单事务性通知、管理中台发件配置与模板管理。  
> **实施边界**：本设计约束后续实现方向；状态/事件契约以 [`order-status-event-template-contract.md`](./order-status-event-template-contract.md) 为准。在代码实施前，严禁新增业务代码采用“在业务逻辑内直接调用 SMTP”的做法。

---

## 1. 架构目标与核心原则

### 1.1 现状痛点与重构动因
在当前的工程基线中，邮件通知与发件能力存在以下深层问题：
1. **职责强耦合**：发件通道配置（SMTP 参数）、邮件内容（内嵌 HTML 模板）与业务逻辑（订单完成）紧密混杂在 `internal/pkg/email/email.go` 中，缺乏独立抽象。
2. **缺乏多语言支持（i18n）**：重构前的订单确认与发货邮件模板是硬编码的英文内嵌文件，无法根据海外买家的语言偏好（如法语、德语、中文、西班牙语等）动态匹配；现行链路使用数据库模板编码与 locale fallback。
3. **售后（After-Sales / RMA）通知盲区（历史问题）**：重构前 `after_sales_cases` 的申请、审核、寄回、签收、完结、驳回流转没有自动化邮件通知。本项目已通过 canonical Outbox、模板规则和异步投递链路补齐代码能力；生产 SMTP 凭据和模板内容验收仍未完成。
4. **安全与运维风险**：SMTP 密码目前散落在环境变量或未加密的旧配置中，缺少统一的加密存储、脱敏展示与连通性自检手段。

### 1.2 三层解耦核心架构原则
系统必须彻底分化为相互独立、单向依赖的三层架构，绝不能把“怎么发”、“发什么”与“什么时候发”混为一谈：

```text
+-----------------------------------------------------------------------------------+
| 3. 事件绑定层 (Event Binding & Outbox Layer)                                      |
|    负责【什么时候发】：业务状态机事务内写入 Outbox -> 异步 Worker 调度 -> 幂等去重与重试    |
+------------------------------------------+----------------------------------------+
                                           |
                                           v
+-----------------------------------------------------------------------------------+
| 2. 邮件模板层 (Email Template & i18n Layer)                                       |
|    负责【发什么】：模板编码 + 语言判定 -> 受控白名单变量渲染 (禁止任意代码执行) -> 标题/HTML  |
+------------------------------------------+----------------------------------------+
                                           |
                                           v
+-----------------------------------------------------------------------------------+
| 1. 发件配置层 (Email Provider Config Layer)                                       |
|    负责【怎么发】：SMTP 连接池 / API 网关 -> 凭据加密(AES-256-GCM) / 脱敏 -> 连通性测试     |
+-----------------------------------------------------------------------------------+
```

---

## 2. 第一层：发件通道配置层 (Email Provider Config)

负责抽象“底层发信通道”，屏蔽不同 SMTP 提供商（如 SendGrid、AWS SES、Postmark、Gmail Workspace 或自建 SMTP）的通信差异。

### 2.1 配置属性与安全规范
系统引入专用的 `email_provider_configs` 存储发件凭据，字段规约如下：

| 属性 | 类型 | 说明与安全约束 |
| :--- | :--- | :--- |
| `id` | `BIGINT` | 主键 ID |
| `code` | `VARCHAR(64)` | 通道编码（如 `primary_smtp`, `backup_smtp`） |
| `name` | `VARCHAR(128)` | 显示名称（如 `AWS SES Production SMTP`） |
| `driver` | `VARCHAR(32)` | 驱动类型：`smtp`（后续可扩展 `sendgrid_api`, `ses_api` 等） |
| `host` | `VARCHAR(255)` | SMTP 主机地址（如 `email-smtp.us-east-1.amazonaws.com`） |
| `port` | `INT` | 端口：`587` (STARTTLS) 或 `465` (SSL/TLS) |
| `username` | `VARCHAR(255)` | SMTP 认证用户名 |
| `password_encrypted`| `TEXT` | **必须经 AES-256-GCM 主密钥加密存储**，数据库严禁存储明文密码 |
| `from_name` | `VARCHAR(128)` | 发件人名称（如 `H-GRIPE Customer Care`） |
| `from_email` | `VARCHAR(255)` | 发件人邮箱（如 `support@learn.gripe`） |
| `reply_to` | `VARCHAR(255)` | 回复邮箱（如 `service@learn.gripe`） |
| `encryption_type` | `VARCHAR(16)` | 连接加密协议：`starttls` / `tls` / `none` |
| `is_active` | `BOOLEAN` | 启用/停用状态 |
| `is_default` | `BOOLEAN` | 是否为系统默认主通道 |
| `last_tested_at` | `TIMESTAMPTZ` | 最近一次连通性测试时间 |
| `last_test_status` | `VARCHAR(32)` | 最近一次测试结果：`healthy` / `failed` |

### 2.2 凭据加密与优先级规则
1. **加密机制**：密码加密复用平台级 `PAYMENT_CONFIG_MASTER_KEY` 或专用的 `SECURITY_MASTER_KEY`，使用 `AES-256-GCM` 算法派生子秘钥加密。
2. **API 脱敏原则**：管理端 GET 接口下发的密码必须脱敏为 `******` 或布尔标志位 `has_password: true`，严禁在网络报文中回显真实解密密码。
3. **分层覆盖优先级**：
   $$\text{生效配置} = \text{数据库中启用的默认配置 (DB Config)} \gg \text{系统环境变量 (ENV Fallback)} \gg \text{无发信能力 (Mock/Disabled)}$$
   - 生产环境中若未在管理后台录入启用且默认的真实配置，则平滑降级至 `SMTP_HOST`、`SMTP_PORT`、`SMTP_USERNAME`、`SMTP_PASSWORD` 等系统环境变量；
   - 环境变量模式下不允许通过后台界面修改，保证基础设施只读部署的稳定性。

   - 数据库 provider 一旦被选为启用且默认配置，即优先于环境变量；provider 密码无法解密或配置无效时必须失败闭环，不得静默回退到环境变量。

### 2.3 通道自检与连通性测试
管理端提供独立操作接口：
- `POST /api/admin/email/providers/:id/test`
  - 携带测试目标邮箱（`target_email`）；
  - 核心逻辑：发起真实的 SMTP 建立连接、TLS 握手、AUTH 认证及发送一封标准的验证测试邮件；
  - 返回脱敏 provider 视图，并将 `last_tested_at`、`last_test_status`、`last_test_error` 持久化；测试目标邮箱必须显式传入，认证失败不得返回密码。

---

## 3. 第二层：邮件模板与多语言引擎 (Template & i18n Layer)

负责将动态业务上下文与标准化视觉表现进行拼装，定义“发什么内容”。

### 3.1 模板编码规范 (Template Taxonomy)
邮件模板以**受控模板编码**作为系统调用的唯一锚点，覆盖订单与售后中经过审查的客户可见通知节点；完整的状态/事件边界见 [`order-status-event-template-contract.md`](./order-status-event-template-contract.md)：

| 模板编码 (`code`) | 业务场景 | 触发时机 |
| :--- | :--- | :--- |
| `order_confirmation` | 订单支付确认 | 订单交易支付成功并完成入账 |
| `order_payment_expired` | 支付超时 | 未付款订单被系统过期任务成功认领 |
| `order_cancelled` | 订单取消 | 未付款取消及库存回滚事务提交 |
| `order_shipping_notification` | 订单发货通知 | 订单生成物流单号并进入已发货状态 |
| `order_delivered` | 订单送达 | 所有相关包裹均确认送达 |
| `order_completed` | 订单完成 | 订单完成事实提交 |
| `order_refunded` | 退款完成 | 网关退款验证完成并提交本地退款状态 |
| `after_sales_requested` | 售后申请已受理 | 客户提交售后/退换货申请 |
| `after_sales_approved` | 售后申请已批准 | 客服或管理员审核同意该售后诉求 |
| `after_sales_awaiting_return` | 寄回指引与地址告知 | 售后类型需要买家寄回商品，告知仓库收件地址与物流填报指引 |
| `after_sales_return_in_transit` | 退货在途 | 承运商和追踪号与售后状态同事务落库 |
| `after_sales_received` | 售后商品签收确认 | 海外收货仓/售后中心确认收到退件包裹 |
| `after_sales_resolving` | 售后进入处理 | 客服开始核验或执行解决方案（支持 `notification_after_sales_resolving_enabled` 运营开关，缺失时默认发送） |
| `after_sales_completed` | 售后完结结案 | 退款原路入账、换货已发出或补偿已落实 |
| `after_sales_rejected` | 售后申请已驳回 | 审核不通过或超时未寄回驳回并说明原因 |

### 3.2 受控变量白名单机制 (Controlled Variables)
为彻底杜绝任意代码执行（RCE）与模板注入（SSTI）安全隐患，邮件模板引擎**严禁支持任意代码表达式或动态反射执行**。只允许使用类似 Mustache 风格的确定性受控变量替换：

```text
{{variable_name}}
```

系统预设并强制白名单校验的变量集：
- **公共全局变量**：
  - `{{site_name}}`：品牌/站点名称（如 `H-GRIPE`）
  - `{{site_url}}`：官网主站链接
  - `{{customer_name}}`：客户称谓/用户名/姓名
  - `{{support_email}}`：官方客服邮箱
- **订单相关变量**：
  - `{{order_number}}`：业务订单号（如 `ORD-202609-8832`）
  - `{{order_amount}}`：订单支付总金额（格式化展示，如 `$1,299.00 USD`）
  - `{{paid_at}}`：支付成功时间（本地化格式）
  - `{{tracking_number}}`：承运商快递单号
  - `{{tracking_url}}`：物流查询跟踪直达链接
  - `{{carrier_name}}`：承运商名称（如 `DHL Express`, `FedEx`）
  - `{{shipping_address}}`：脱敏后的收件地址摘要
- **售后专属变量**：
  - `{{after_sales_case_number}}`：售后服务单号（如 `RMA-2026-0041`）
  - `{{after_sales_type}}`：售后类型名称（如 `退货退款`, `换货`）
  - `{{after_sales_status}}`：当前流转状态标签
  - `{{return_warehouse_name}}`：退货接收仓名称
  - `{{return_warehouse_address}}`：退货指定寄送完整地址及收件人
  - `{{return_label_url}}`：已生成的退货面单下载/打印地址（没有面单时为空）
  - `{{return_deadline}}`：寄回截止期限
  - `{{refund_amount}}`：核定退款金额（如适用）
  - `{{rejection_reason}}`：驳回/拒绝具体审核原因说明

### 3.3 多语言匹配与内容降级 (i18n Fallback)
模板数据由 `(template_code, locale)` 联合唯一确定。语言匹配规则：
1. **精确匹配**：按订单/售后单所绑定的客户语言代码（如 `fr`、`de`、`zh_cn`）获取对应启用的模板；
2. **全局降级**：若客户指定的语言模板未配置或未启用，强制自动平滑回退至英文主模板（`en`）；
3. **硬核兜底**：若数据库中全量模板均缺失，系统内置一份极简的只读静态纯文本代码兜底，确保事件不丢单。

### 3.4 模板版本追踪与防误触 (Audit & Versioning)
每次管理员在后台编辑并保存模板内容（标题、HTML 正文、纯文本回退）时，系统自动将当前内容快照归档至 `email_template_versions` 表中，记录操作人、修改时间与 Diff，支持一键版本回滚。

---

## 4. 第三层：事件驱动与 Outbox 事务绑定 (Event Binding & Outbox)

负责明确“什么时候发”，确保高并发与复杂状态流转下的**强一致性与可靠投递**。

### 4.1 绝对禁止同步调用
> 🛑 **架构禁令**：
> 严禁在订单支付 Webhook、发货处理、售后状态变更的 HTTP 业务事务中同步创建 SMTP 客户端并发送邮件。
> SMTP 属于不可靠网络 I/O，同步调用会导致：
> 1. 请求挂起超时，拖垮数据库长事务；
> 2. 外部 SMTP 瞬时抖动直接引发客户核心业务回滚（如退款成功但因邮件发送失败导致事务中断）。

### 4.2 事务级 Outbox 写入契约
所有邮件通知触发，必须作为业务变更的伴随副作用，与核心实体（`orders`、`after_sales_cases`）在**同一个本地数据库事务**内持久化到 Outbox 消息队列中：

```text
[HTTP 业务请求]
       │
       ▼
[开始本地事务 (Begin Tx)]
  ├── 1. 更新订单 / 售后业务状态 (Update orders / after_sales_cases)
  ├── 2. 写入状态变更审计日志 (Insert audit_logs)
  └── 3. 构造并写入 Outbox 事实 (Insert outbox_events: event_type=<canonical domain event>)
[提交本地事务 (Commit Tx)]
       │ (事务原子提交成功)
       ▼
[异步 Outbox Worker 轮询/监听]
       ├── 读取待处理事件 (Lock row: SKIP LOCKED)
       ├── 匹配发件通道 (Email Provider)
       ├── 匹配多语言模板 (Template Engine)
       ├── 组装变量并渲染 HTML/Text
       ├── 调用 SMTP 投递发信
       └── 标记 Outbox 状态为 PROCESSED (或失败重试 RETRY)
```

### 4.3 状态映射矩阵与幂等键规则

| 业务事件来源 | 触发状态转换 | 目标模板编码 (`template_code`) | 幂等键生成规则 (`idempotency_key`) |
| :--- | :--- | :--- | :--- |
| **订单支付成功** | `payment_status` 变为 `paid` 且网关交易已验证 | `order_confirmation` | `payment_transaction:{transaction_id}` |
| **支付过期** | 主订单变为 `payment_expired` 且支付状态变为 `expired` | `order_payment_expired` | `order_expiration:{order_id}`（当前单次过期事实） |
| **订单取消** | 未付款订单取消并完成库存回滚 | `order_cancelled` | `order_cancellation:{order_id}`（当前单次取消事实） |
| **订单拆单发货** | 新增包裹并取得追踪号 | `order_shipping_notification` | `shipment:{shipment_id}` |
| **订单送达** | 所有相关包裹确认送达 | `order_delivered` | `delivery_fact:{order_id}:{tracking_number}`（当前拆单兼容键） |
| **订单完成** | 主订单完成事实提交 | `order_completed` | `order_completion:{order_id}`（当前完成事实） |
| **退款完成** | 网关退款验证完成并写入本地退款记录 | `order_refunded` | `refund:{refund_id}` |
| **售后单创建** | 状态初始化为 `requested` | `after_sales_requested` | `after_sales_case:{case_id}:transition:{transition_id}` |
| **售后单已批准** | 状态流转为 `approved` | `after_sales_approved` | `after_sales_case:{case_id}:transition:{transition_id}` |
| **指示买家寄回** | 状态流转为 `awaiting_return` | `after_sales_awaiting_return` | `after_sales_case:{case_id}:transition:{transition_id}` |
| **退货在途** | 状态流转为 `return_in_transit` 且包裹同事务保存 | `after_sales_return_in_transit` | `after_sales_case:{case_id}:transition:{transition_id}` |
| **仓库确认签收** | 状态流转为 `received` 且收货人已记录 | `after_sales_received` | `after_sales_case:{case_id}:transition:{transition_id}` |
| **售后进入处理** | 状态流转为 `resolving` | `after_sales_resolving`（支持运营开关，缺失时默认发送） | `after_sales_case:{case_id}:transition:{transition_id}` |
| **售后已完结** | 状态流转为 `completed` 且解决方案事实提交 | `after_sales_completed` | `after_sales_case:{case_id}:transition:{transition_id}` |
| **售后已驳回** | 状态流转为 `rejected` | `after_sales_rejected` | `after_sales_case:{case_id}:transition:{transition_id}` |

**幂等性保障**：幂等键必须绑定支付交易、包裹、退款或状态转换事实 ID，并映射到现有 `outbox_events.event_key` 唯一索引；不得仅使用当前时间戳。即使由于网络重试导致状态机重复执行，重复的 Outbox 事件插入也会被安全忽略。

### 4.4 失败重试、指数退避与死信机制
- **重试策略**：初次发送失败后，状态置为 `failed`，重试间隔采用指数退避（$2^n \times 30\text{s}$，即 30s、60s、120s、240s、480s）；
- **最大重试上限**：最多重试 5 次；
- **死信状态 (`dead_letter`)**：超过 5 次仍失败（如目标邮箱不存在或 SMTP 服务商持续拒绝），事件进入死信状态并触发调度/监控统计。后台已提供基于现有 `outbox_events` 的失败/死信列表、详情、人工重试和忽略页面；不新增第二套邮件队列。

---

## 5. 数据模型与数据库迁移规划 (Database Schema)

模板与版本表已使用 migration `335_notification_templates` 创建，英文默认模板由 `336_seed_transactional_notification_templates` 写入；发件通道表使用 migration `341_email_provider_configs`，不得复用已有 migration：

### 5.1 发件配置表 (`email_provider_configs`)
```sql
CREATE TABLE IF NOT EXISTS email_provider_configs (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    driver VARCHAR(32) NOT NULL DEFAULT 'smtp',
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL DEFAULT 587,
    username VARCHAR(255) NOT NULL DEFAULT '',
    password_encrypted TEXT NOT NULL DEFAULT '',
    from_name VARCHAR(128) NOT NULL,
    from_email VARCHAR(255) NOT NULL,
    reply_to VARCHAR(255) NOT NULL DEFAULT '',
    encryption_type VARCHAR(16) NOT NULL DEFAULT 'starttls',
    is_active BOOLEAN NOT NULL DEFAULT false,
    is_default BOOLEAN NOT NULL DEFAULT false,
    last_tested_at TIMESTAMPTZ NULL,
    last_test_status VARCHAR(32) NOT NULL DEFAULT '',
    last_test_error VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_provider_configs_active_default 
    ON email_provider_configs(is_active, is_default);
```

### 5.2 邮件模板表 (`email_templates`)
```sql
CREATE TABLE IF NOT EXISTS email_templates (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    locale VARCHAR(16) NOT NULL DEFAULT 'en',
    category VARCHAR(32) NOT NULL DEFAULT 'order', -- 'order' | 'after_sales' | 'system'
    name VARCHAR(128) NOT NULL,
    subject_template VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL DEFAULT '',
    body_text TEXT NOT NULL DEFAULT '',
    allowed_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    required_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_email_template_code_locale UNIQUE (code, locale)
);

CREATE INDEX IF NOT EXISTS idx_email_templates_lookup 
    ON email_templates(code, locale, is_enabled);
```

### 5.3 模板版本历史表 (`email_template_versions`)
```sql
CREATE TABLE IF NOT EXISTS email_template_versions (
    id BIGSERIAL PRIMARY KEY,
    template_id BIGINT NOT NULL REFERENCES email_templates(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL,
    locale VARCHAR(16) NOT NULL,
    version INT NOT NULL,
    name VARCHAR(128) NOT NULL,
    subject_template VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL DEFAULT '',
    body_text TEXT NOT NULL DEFAULT '',
    allowed_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    required_variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    changed_by_user_id BIGINT NULL,
    change_reason VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_template_versions_template_id 
    ON email_template_versions(template_id, version DESC);
```

---

## 6. 管理中台交互规范 (Admin Console UDS 1.0)

新功能应作为独立的功能子域接入 `go-backend/web/admin`，页面严格遵守 ERP UDS 1.0 工业级规范：

### 6.1 页面规划与菜单挂载
- **发件通道管理**：挂载在 **系统设置 (Settings) $\to$ 邮件通道 (Email Providers)**。
  - 提供通道列表、默认主通道切换、编辑抽屉；
  - 核心操作：**测试发送弹窗 (Send Test Modal)**，支持指定收件人并在 3 秒内反馈连接测试结果；
  - 密码输入框支持“修改密码”切换，平时严格展示脱敏掩码。
- **邮件模板管理**：挂载在 **内容/营销 (Content/Marketing) $\to$ 事务邮件模板 (Email Templates)**。
  - 左侧/顶栏提供模板分类 Filter：`全部`、`订单通知 (Order)`、`售后通知 (After-sales)`；
  - 列表展示：模板名称、编码、当前语言版本标签（如 `EN (默认)`, `ZH`, `FR`）、状态 Badge（`HEALTHY` / `ALERT`）；
  - 编辑器抽屉（Drawer）：
    - 采用标准两栏/选项卡切换：**HTML 源码编辑** / **富文本预览 (Live Preview)** / **纯文本回退**；
    - 右侧提供“受控变量白名单标签板”，点击即可复制或插入 `{{order_number}}` 等占位符；
    - 提供“发送测试邮件”入口，自动注入 Mock 变量数据，直接发送到操作人员邮箱用于真机渲染验证。

### 6.2 视觉原子标准 (Strict UDS Alignment)
1. **模块页眉**：统一采用 `rounded-[32px] border border-dashed border-border/80 bg-muted/5 p-6`，主标题必须使用 `text-lg font-black tracking-tighter italic uppercase`。
2. **状态指示器**：
   - 通道就绪：`Emerald-500/10 背景 + Emerald-600 文字`（`HEALTHY`）。
   - 凭证未测试/警告：`Amber-500/10 背景 + Amber-600 文字`（`ALERT`）。
   - 连通性报错：`Rose-500/10 背景 + Rose-600 文字 + animate-pulse`（`CRITICAL`）。
3. **按钮与交互**：主操作采用胶囊按钮 `rounded-full h-11 font-black text-[10px] uppercase tracking-widest`，具备乐观更新与错误防误触。

---

## 7. 现有文档冲突清理与对齐结论

本章节作为对现有工程文档的审计结论，明确哪些旧描述已被本架构规范正式取代：

| 现有文档 | 冲突/陈旧描述 | 权威对齐修正结论 |
| :--- | :--- | :--- |
| **`docs/ops/production-readiness-status.md`** | 生产状态表区分了当前 email challenge 与尚未完成生产验收的事务邮件。 | **已对齐**：订单/售后 canonical Outbox、模板渲染、provider 后台配置、运行时 provider 选择和 SMTP 测试已落地；正式生产放量仍须按运行手册完成真实凭据验收。 |
| **`docs/ops/hostinger-vps-docker-runbook.md`** | 当前运行手册只描述已接通的 newsletter/warranty 出站 SMTP。 | **保持现状**：订单确认与售后通知只有在本契约及验收通过后才纳入生产流量；容器网络需允许标准 587/465 连接。 |
| **`docs/design/payment-state-machine-architecture.md`** | 仅在图示中标注 `Outbox (email, fulfilment, conversion, audit)`，未细化事件结构。 | **继承细化**：支付域确认生成 `order_confirmation` 模板对应的 Outbox 事件，严格依循本文档第 4.3 节的幂等键与 payload 规范。 |
| **`go-backend/internal/pkg/email/email.go`** | 订单确认、发货曾使用内嵌英文模板和专用发送方法，与 canonical 领域事件/数据库模板形成双轨。 | **已清理**：订单内嵌模板及专用方法已删除；订单和售后事务通知以 `email_templates` 和 canonical Outbox 为准。密码重置、欢迎邮件等非本次范围的账户邮件仍可使用既有内嵌模板。 |

---

## 8. 分期实施路线图 (Implementation Sequence)

以下阶段与 [`order-status-event-template-contract.md`](./order-status-event-template-contract.md)
保持同一顺序。阶段名称描述交付边界，不表示 SMTP、模板和业务事实可以互相替代。

- **阶段 0：状态/事件契约冻结**
  - 先完成 [`order-status-event-template-contract.md`](./order-status-event-template-contract.md) 的评审和 payload schema 定稿；
  - 在契约冻结前，不实现 SMTP 配置页面或“状态下拉框绑定模板”。
- **阶段 1：领域事件与事务 Outbox（已完成）**
  - 为支付成功、支付过期、取消、发货、送达、退款完成、订单完成和售后状态变更提供版本化事件；
  - 业务状态、审计 transition 和 Outbox 必须在同一本地事务内提交；
  - 冻结事件到模板编码、模板分类、必填/允许变量和 locale fallback 契约；Outbox 校验幂等 envelope 与事件最小事实字段，读取启用模板并完成正文渲染及异步投递；
  - `order.paid` 继续作为业务集成事件；确认/发货通知统一由 canonical 领域事实驱动，不保留旧邮件命令分支。
  - **阶段 2：发件通道与模板持久化（已完成）**
  - 模板与版本表已由 migration `335_notification_templates` 创建，英文默认种子由 `336_seed_transactional_notification_templates` 写入；发件 provider 表由 migration `341_email_provider_configs` 创建；
  - 模板保存、版本快照、受控变量渲染和 locale fallback 已完成；
  - `EmailProviderService`、AES-256-GCM 密码加密/解密、provider 脱敏 CRUD、运行时默认 provider 选择和 SMTP 测试发送已完成；每次发送动态读取数据库默认 provider，无默认 provider 时回退环境变量，解密失败时 fail-closed。
  - `en` 模板种子、后台模板编辑、HTML/纯文本预览、版本历史和回滚为新版本已完成；未完成事件绑定的模板必须标记为未启用。
- **阶段 3：管理端与通知投递**
  - 当前代码已提供 provider-neutral 的 `TransactionalNotificationDeliveryPlan`：异步 worker 将领域事件解析为模板编码、recipient、locale、幂等键和变量快照，并完成模板渲染与邮件发送；发送层优先读取数据库默认 provider，无默认 provider 时才复用环境变量 SMTP；
  - 模板版本管理和预览已交付；后续只补充真实网关/物流/售后回放验收和外部订阅；
  - 复用阶段 1 已冻结的领域事件→模板编码规则，由异步 worker 根据投递计划完成 locale fallback、受控变量渲染和发送，不在业务请求内同步调用 SMTP。
- **阶段 4：外部订阅与全量验收**
  - 在真实网关/物流/售后回放测试通过后，再接入 Zendesk、邮件网关或站内通知；
  - 旧邮件命令已废弃；后续外部订阅或邮件网关只能消费 canonical 领域事实，不能恢复旧事件或按旧文件名分流。

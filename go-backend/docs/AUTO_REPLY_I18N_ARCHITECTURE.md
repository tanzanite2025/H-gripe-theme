# 自动回复多语言架构与落地方案

最后审阅日期：2026-08-09

本文定义 Tanzanite 客服自动回复的多语言规则、FAQ 引用、翻译 API
和迁移边界。本文是 `AUTO_REPLY_SYSTEM_DESIGN.md` 的多语言补充决策。

当两份文档在“自动回复是否允许全局语言兜底”上存在差异时，以本文为准：

- 短期：一条自动回复只能绑定一个语言。
- 当前语言没有录入内容时，不自动回复。
- 不把一条中文或英文文本复制成“全部语言”。
- 长期“全部语言”只表示翻译管理视图或翻译任务范围，不表示运行时使用
  `*` 匹配所有语言。

## 1. 核心决策

### 1.1 语言是内容属性，不只是匹配条件

自动回复包含两类事实：

1. **规则事实**
   - 欢迎语或关键词；
   - 关键词匹配方式；
   - 客服或客服组范围；
   - 优先级；
   - 冷却时间；
   - 启用状态。
2. **语言内容事实**
   - 该语言下的回复文本；
   - 消息类型；
   - 链接、商品、订单、图片或 FAQ metadata；
   - 翻译状态和审核状态。

短期当前表结构把这两类事实放在同一行中，因此一条数据库记录只能表达：

```text
一个规则 + 一个语言 + 一份回复内容
```

不能把 `locale = *` 理解成“一条记录包含 20 种 storefront locale 的内容”。
它实际上只有一份 `reply_message`，无法安全代表 20 种语言。

### 1.2 短期的运行时原则

假设访客当前选择的是法语：

```text
请求语言：fr
查找：locale = fr 的规则
存在：执行匹配
不存在：不自动回复
```

禁止以下行为：

- `fr` 没有内容时降级到 `en`；
- `fr` 没有内容时使用其他语言的规则；
- 新建一条 `*` 规则并把它当作所有语言；
- 把浏览器语言当成高于用户当前站点语言的事实；
- FAQ 没有当前语言版本时展示其他语言 FAQ。

`en-US`、`en-GB` 等区域语言可以规范化为系统支持的 `en`。这属于语言
代码规范化，不属于“找不到语言后的 fallback”。

`zh-CN`、`zh`、`zh_CN` 应规范化为系统标准代码 `zh_cn`。

### 1.3 长期的运行时原则

长期拆分规则与语言内容后，运行时仍然只做精确语言查找：

```text
规则：Payment security
请求语言：de
查找：该规则的 de 已发布内容
存在：执行匹配
不存在：不自动回复
```

“翻译全部语言”只是后台批量生成内容的操作，不会改变运行时的精确匹配
规则。

## 2. 当前实现审计

### 2.1 已存在的语言事实源

跨前后端的 storefront locale registry 是：

- `shared/storefront-locales.json`

它定义当前固定的 20 个 storefront locale，包括 canonical `code`、Nuxt `iso`
和 `file`、后端/Admin 使用的英文 `name` 与 `native_name`。新增、删除或改名
必须先改这个 registry，并运行 `nuxt-i18n/scripts/check-locales.ts`，确保
Nuxt manifest、Go backend locales 和 Admin fallback 三处完全对齐。

后端运行时语言清单：

- `internal/pkg/locales/locales.go`
  - `SupportedLanguages`
  - `ResolveSupported`
  - `SupportedLocaleCodes`
  - `EnabledLocaleCodes`
- `internal/api/v1/i18n/handler.go`
  - `GET /api/v1/i18n/languages`
- `internal/service/locale.go`
  - `requireSupportedLocale`

后台已有复用语言清单的通用逻辑：

- `web/admin/src/composables/useSupportedLanguages.ts`
- `web/admin/src/lib/languages.ts`
- `web/admin/src/api/i18n.ts`

Nuxt 前台还有自己的 locale manifest：

- `nuxt-i18n/app/i18n/locales.manifest.ts`

当前 20 个 storefront locale 代码必须完全一致，中文的系统标准代码是 `zh_cn`。
自动回复不应再创建第三套语言数组，而应复用后端语言 API 和后台已有的
语言 composable；本地 fallback 也必须与 `shared/storefront-locales.json` 对齐。

### 2.2 当前自动回复规则结构

当前模型是：

- `internal/domain/ticket/auto_reply_rule.go`
- 表：`ticket_auto_replies`
- 关键字段：
  - `type`
  - `trigger_keyword`
  - `reply_message`
  - `agent_id`
  - `group_id`
  - `locale`
  - `message_type`
  - `metadata`
  - `attachments`
  - `is_active`
  - `priority`
  - `match_type`
  - `cooldown_seconds`

当前结构天然更适合“每个语言一条规则”。它暂时不适合一条规则保存
20 份文本，因为没有独立的语言内容表。

### 2.3 当前存在的职责问题

#### 禁止后台语言自由输入

后台不能为 storefront-facing 内容提供自由文本 locale 输入。历史上如果允许输入：

```text
* / en / zh-cn
```

这会造成：

- 拼写错误；
- 空格和大小写不一致；
- 保存未知语言；
- FAQ 选择器无法确定应该加载哪个语言；
- 规则已经保存，但永远无法匹配。

当前和后续 Admin 实现都必须使用受控 storefront locale 选择器，选项来源于
`useSupportedLanguages()` / `StorefrontLocaleSelect`，而不是页面本地数组或自由输入。

#### 自动回复 locale 解析职责

旧实现曾在 `internal/service/ticket_auto_reply_service.go` 中单独处理
自动回复 locale，容易和 FAQ、产品、文章的语言格式不一致。

当前实现已将自动回复 locale 解析独立到
`internal/service/auto_reply_locale.go`，并统一调用
`locales.ResolveSupported()`。数据库和运行时只使用 canonical locale，
例如 `zh_cn`，不会再维护第二套自动回复语言规范。

#### Repository 也参与语言策略

`internal/repository/ticket_repository.go` 当前会在 SQL 条件中做：

- locale 大小写处理；
- `_` 和 `-` 转换；
- 基础语言匹配；
- `*` 和空值兜底。

Repository 应该只负责查询已经确定的语言。是否允许基础语言匹配、是否允许
全局兜底，属于 service 业务政策，不应该分散在 SQL 和 service 两处。

#### Nuxt 没有明确传递站点语言

`nuxt-i18n/app/composables/chat/useCustomerServiceChatSync.ts` 当前创建会话
和发送消息时没有明确传递 `useI18n().locale.value`。

后端因此可能依赖 `Accept-Language`。例如：

```text
站点当前语言：en
浏览器 Accept-Language：zh-CN,zh;q=0.9
```

如果只看请求头，可能会匹配中文自动回复，和用户正在浏览的页面语言
不一致。

## 3. 术语和边界

### 3.1 Supported Locale

系统已经支持并允许使用的语言代码，例如：

```text
en
fr
de
zh_cn
```

来源必须是 `internal/pkg/locales`，而不是自动回复页面自己维护的数组。

### 3.2 Request Locale

本次客服请求实际使用的站点语言。

优先级应为：

1. Nuxt 当前 `locale.value`；
2. 显式 API 参数或请求体中的 `locale`；
3. 已保存的语言 cookie；
4. `Accept-Language`；
5. 无法解析时标记为未确定。

对于自动回复而言，未确定语言时不应该猜测，也不应该自动使用英文。

### 3.3 Content Locale

自动回复内容实际编写或翻译成的语言。

短期它直接等于规则行的 `locale`。长期它是语言内容表中的
`auto_reply_rule_contents.locale`。

### 3.4 Translation Set

同一条语言无关规则下的所有语言内容集合。

例如：

```text
规则：Payment security
内容：
  en -> English text
  zh_cn -> Chinese text
  fr -> French text
  de -> German text
```

Translation Set 是后台管理概念，不是运行时的 `*` wildcard。

## 4. 短期方案：一条规则一个语言

### 4.1 后台表单

自动回复编辑页面应改为：

```text
语言
[选择一个语言]
```

下拉选项来自：

```text
GET /api/v1/i18n/languages
```

只显示 `enabled = true` 的语言。

短期不显示：

- “全部语言”；
- `*`；
- 自由文本输入；
- 区域语言自由输入，例如 `en-US`、`fr-CA`。

后台可以展示语言名称和代码，例如：

```text
English · en
Français · fr
简体中文 · zh_cn
```

### 4.2 后端保存规则

保存时必须调用统一的：

```go
locales.ResolveSupported(input.Locale)
```

保存结果只能是：

- 一个受支持的 canonical locale；
- 或返回 `unsupported locale`。

空值、`*`、未知代码均不能创建新的本地化自动回复规则。

请求中的 `en-US` 等别名可以接受，但只能在进入 service 后解析成
`en`，数据库中不保存别名。

### 4.3 运行时匹配

运行时处理顺序：

1. 取得当前请求语言；
2. 调用统一 locale resolver；
3. 如果无法解析，直接跳过自动回复；
4. 查询规则时只使用 canonical locale；
5. 按客服、客服组、优先级和关键词继续匹配；
6. 找不到该语言的规则，返回“无自动回复”；
7. 不执行跨语言 fallback。

推荐的 service 语义：

```go
resolvedLocale, ok := ResolveCustomerServiceLocale(requestLocale)
if !ok {
    return NoAutoReply
}

rules := repository.FindActiveAutoReplyRulesByExactLocale(
    ruleType,
    resolvedLocale,
    agentID,
    groupIDs,
)
```

Repository 不再自行决定：

- `*` 是否匹配；
- 基础语言是否匹配；
- 英文是否为默认语言；
- 是否允许其他 locale 兜底。

### 4.4 FAQ 自动回复

当前 FAQ 已经按语言拆开管理：

- `faq_pages.locale`
- `faq_categories.locale`
- `faqs.locale`
- `faqs.parent_id` 可用于翻译关联

自动回复 FAQ 选择器当前通过 `form.locale` 查询对应语言的已发布 FAQ，
这个方向是正确的，但必须保证 `form.locale` 一定是具体语言，不能是 `*`。

短期 FAQ 自动回复保存时至少保留：

```json
{
  "faq_id": 123,
  "page_id": "support-payment",
  "category": "payment-security",
  "locale": "en",
  "question": "Is my payment secure?",
  "url": "/support/faqs?page=support-payment&faq=123"
}
```

后端保存前应再次校验：

1. FAQ 存在；
2. FAQ 的 locale 等于规则 locale；
3. FAQ 状态为 published；
4. FAQ 所属页面和分类仍然有效。

如果 `fr` 规则引用了 `en` FAQ，应该拒绝保存，而不是让前端之后猜测。

## 5. 现有 `*` 数据的处理

### 5.1 为什么不能自动复制

一条现有 `*` 规则只有一份文本：

```text
Welcome to Tanzanite.
```

系统不能假设它已经包含：

- 法语；
- 德语；
- 中文；
- 日语；
- 其他 storefront locale。

自动复制只会制造“看起来有 20 种语言，实际内容都是英文”的假翻译状态。

### 5.2 迁移建议

在短期迁移阶段：

1. 统计 `ticket_auto_replies.locale = '*'` 的规则；
2. 在后台标记为“未拆分语言规则”；
3. 暂停这些规则参与本地化自动回复匹配；
4. 管理员选择一个真实语言后重新保存；
5. 需要其他语言时，再分别创建或复制到对应语言；
6. 不自动把 `*` 转为 20 条规则。

如果业务明确确认某一条规则是完全语言无关的，例如只返回一个不含自然
语言的订单链接，未来可以单独增加 `language_neutral` 类型。但它不应和
普通文本自动回复共用 `*` 语义。

### 5.3 旧接口兼容

旧的 welcome/match HTTP 接口可以继续存在一段时间，但必须使用同一套
locale resolver 和精确匹配规则。不能因为旧接口没有传 locale，就恢复
`*` 全局兜底。

如果旧客户端完全没有语言信息，最安全的行为是：

```text
不触发本地化自动回复
```

而不是盲目发英文或中文。

## 6. 长期方案：规则与语言内容拆表

### 6.1 目标结构

长期建议把当前单表结构演进为：

```text
ticket_auto_replies
  规则本体，不保存某一个语言的文本

ticket_auto_reply_contents
  同一规则的多语言内容
```

### 6.2 规则本体

`ticket_auto_replies` 长期保留：

| 字段 | 说明 |
| --- | --- |
| `id` | 规则主键 |
| `type` | `welcome` 或 `keyword` |
| `trigger_keyword` | 关键词 |
| `match_type` | `exact` 或 `contains` |
| `agent_id` | 指定客服 |
| `group_id` | 指定客服组 |
| `is_active` | 规则是否启用 |
| `priority` | 规则优先级 |
| `cooldown_seconds` | 冷却时间 |
| `source_locale` | 翻译源语言 |
| `created_at` / `updated_at` | 审计时间 |

长期不应再把 `reply_message`、`message_type`、`metadata` 和
`attachments` 作为规则本体字段。

### 6.3 语言内容表

建议新增：

```text
ticket_auto_reply_contents
```

建议字段：

| 字段 | 说明 |
| --- | --- |
| `id` | 内容主键 |
| `rule_id` | 所属规则 |
| `locale` | canonical locale |
| `reply_message` | 当前语言文本 |
| `message_type` | `text`、`image`、`link`、`product`、`order`、`faq` |
| `metadata` | 结构化内容 |
| `attachments` | 图片或附件 JSON |
| `status` | `draft`、`published`、`translation_pending`、`translation_failed` |
| `source_locale` | 本条内容翻译来源 |
| `translation_provider` | 翻译服务名称 |
| `translation_version` | 翻译版本 |
| `source_hash` | 源文本 hash |
| `reviewed_at` | 人工审核时间 |
| `published_at` | 发布日期 |
| `created_at` / `updated_at` | 审计时间 |

建议约束：

```text
UNIQUE(rule_id, locale)
```

同一规则同一语言只能有一份当前内容。

### 6.4 长期运行时查询

运行时不查询所有语言内容再由前端选择，而是直接查询：

```text
rule_id = matched_rule_id
locale = resolved_request_locale
status = published
```

如果内容不存在或不是 `published`：

```text
不生成自动回复消息
```

这可以避免以下错误：

- 翻译草稿被访客看到；
- 英文文本误用于德语用户；
- 规则存在但语言内容缺失时仍然发出错误消息；
- 前端收到 20 份内容后自行判断显示哪一份。

## 7. 翻译 API 设计

### 7.1 翻译 API 的边界

翻译 API 只负责生成候选内容，不负责：

- 创建客服规则；
- 决定规则是否匹配；
- 自动发布未经审核内容；
- 修改 FAQ 的页面结构；
- 将一个 FAQ ID 强行复制到其他语言。

后台系统仍然是内容发布事实源。

### 7.2 推荐工作流

```text
管理员选择规则
  -> 选择源语言
  -> 填写源语言内容
  -> 点击自动翻译
  -> 后端读取目标语言列表
  -> 翻译 API 生成目标语言草稿
  -> 后端保存 translation_pending / draft
  -> 管理员审核
  -> 发布通过的语言内容
  -> 只有 published 内容参与运行时匹配
```

### 7.3 翻译接口建议

长期可以设计为：

```text
GET  /api/admin/customer-service/auto-reply/rules/:id/contents
PUT  /api/admin/customer-service/auto-reply/rules/:id/contents/:locale
POST /api/admin/customer-service/auto-reply/rules/:id/translate/preview
POST /api/admin/customer-service/auto-reply/rules/:id/translate/apply
POST /api/admin/customer-service/auto-reply/rules/:id/contents/:locale/publish
```

翻译预览请求示例：

```json
{
  "source_locale": "en",
  "target_locales": ["fr", "de", "zh_cn"],
  "reply_message": "How can we help you?",
  "message_type": "text",
  "metadata": {},
  "attachments": []
}
```

翻译应用必须支持：

- 只翻译缺失语言；
- 覆盖已有草稿；
- 不覆盖已发布内容，除非明确确认；
- 单个目标语言失败时保留其他成功结果；
- 记录 provider、请求时间、错误信息和源文本 hash；
- 防止重复点击造成重复任务；
- 允许管理员逐语言重新翻译。

### 7.4 结构化消息翻译

翻译 API 只翻译自然语言字段：

- `reply_message`
- `title`
- `description`
- FAQ 的 `question` 和 `answer`，如果业务允许自动翻译 FAQ

以下字段不能被翻译 API 改写：

- URL；
- 商品 ID；
- 订单 ID；
- FAQ ID；
- 资产 URL；
- SKU；
- JSON 字段名；
- 内部 route path。

翻译 API 返回结果后，后端必须再次通过现有 metadata 校验器验证，
不能因为内容来自第三方 API 就跳过安全校验。

## 8. FAQ 与翻译 API 的长期关系

### 8.1 FAQ 不应只依赖数字 ID

目前自动回复 FAQ metadata 使用 `faq_id`。但 FAQ 的 20 个 storefront locale
内容分别存储，
不同语言的 FAQ 可能拥有不同的数据库 ID。

长期应保存稳定的翻译组身份，例如：

```json
{
  "faq_group_key": "payment-security-card-data",
  "page_id": "support-payment",
  "category": "payment-security",
  "locale": "en"
}
```

后端根据：

```text
faq_group_key + requested locale
```

找到对应语言的已发布 FAQ。

当前 `faqs.parent_id` 已经表达翻译关联意图，可以优先评估是否能稳定承担
该职责。如果现有导入数据的 `parent_id` 不完整，应增加明确的 FAQ 翻译组
标识，而不是让客服自动回复自己猜测关联关系。

### 8.2 FAQ 运行时规则

对于 `message_type = faq`：

1. 自动回复规则必须有具体语言；
2. 对应语言的 FAQ 必须存在；
3. FAQ 必须是 published；
4. FAQ 页面和分类必须存在；
5. FAQ 的文本、图片和链接全部来自该语言版本；
6. 任意一步失败，就不发送 FAQ 自动回复。

不能出现：

```text
德语自动回复规则
  -> 英文 FAQ
```

如果德语 FAQ 尚未录入，正确结果是没有 FAQ 自动回复，而不是显示英文
内容。

## 9. 前后端职责划分

### 9.1 `internal/pkg/locales`

唯一负责：

- 支持语言清单；
- canonical locale；
- alias 解析；
- 是否为启用语言。

它的输入必须与 `shared/storefront-locales.json` 保持一致；不负责自动回复
优先级和规则匹配。

### 9.2 自动回复 service

建议新增独立的 locale helper，例如：

```text
internal/service/ticket_auto_reply_locale.go
```

负责：

- 解析客服请求语言；
- 将 `en-US` 解析为 `en`；
- 将 `zh-CN` 解析为 `zh_cn`；
- 处理缺少语言的情况；
- 生成 exact-match 查询参数；
- 计算语言内容是否可用。

不负责保存页面 FAQ 结构。

### 9.3 Repository

负责：

- 按已经解析好的 locale 查询；
- 查询 active rule；
- 查询 published localized content；
- 处理数据库事务和索引。

不负责：

- 判断 `*` 的业务意义；
- 自动选择英文；
- 解释 Accept-Language；
- 处理翻译 API。

### 9.4 Admin frontend

短期负责：

- 使用受控语言下拉框；
- 显示一条规则对应一个语言；
- 显示 FAQ 当前语言；
- 阻止选择语言不一致的 FAQ。

长期负责：

- 规则编辑；
- 20 个 storefront locale 内容状态矩阵；
- 翻译预览和确认；
- 逐语言审核、发布和撤回。

### 9.5 Nuxt storefront

负责：

- 传递当前 `locale.value`；
- 展示后端已经生成的消息；
- 根据 `message_type` 渲染文本、FAQ、图片、链接和卡片；
- 订阅 SSE 或读取客服历史。

不负责：

- 选择哪条自动回复规则；
- 跨语言 fallback；
- 判断翻译是否完成；
- 从多份语言内容中自行挑选。

## 10. 数据迁移路线

### Phase 0：数据盘点

统计：

- 当前自动回复总数；
- 各 locale 数量；
- `*` 规则数量；
- 未知 locale 数量；
- `zh-cn`、`zh_CN` 等非 canonical 数据；
- FAQ 类型规则是否引用了错误语言 FAQ；
- 是否有重复关键词和重复客服范围。

此阶段只读，不改变线上行为。

### Phase 1：收紧短期规则

目标：

- 后台语言改为下拉框；
- 后端拒绝未知 locale；
- 新规则禁止 `*`；
- 自动回复只做精确语言匹配；
- Nuxt 显式传递站点语言；
- 没有语言内容就不回复。

已有 `*` 规则暂时标记为 legacy/unscoped，并从本地化运行时匹配中移除。

### Phase 2：FAQ 一致性

目标：

- FAQ 选择器只能加载当前规则语言；
- 保存时校验 FAQ locale；
- FAQ 只引用 published 内容；
- 对缺失语言返回明确后台状态；
- 不从英文 FAQ 自动降级。

### Phase 3：规则与内容拆表

目标：

- 新增 `ticket_auto_reply_contents`；
- 将现有每语言规则迁移为：
  - 一条规则本体；
  - 一条对应语言内容；
- 保留旧 API 响应兼容层；
- 运行时改为读取 published localized content。

迁移时不能按关键词简单合并规则。只有以下字段全部一致且业务范围
一致时，才允许合并：

- type；
- trigger_keyword；
- match_type；
- agent_id；
- group_id；
- priority；
- cooldown_seconds；

否则必须保留为不同规则。

### Phase 4：翻译 API

目标：

- 后台显示 Translation Set；
- 支持源语言内容编辑；
- 支持目标语言批量翻译；
- 翻译结果默认 draft；
- 逐语言审核发布；
- 记录翻译版本和源文本 hash。

## 11. API 和错误语义

建议统一以下错误：

| 错误 | 含义 |
| --- | --- |
| `unsupported locale` | 语言不在系统支持列表 |
| `locale is required` | 本地化自动回复缺少语言 |
| `localized reply content is missing` | 规则存在但该语言没有内容 |
| `localized faq not found` | 当前语言没有对应 FAQ |
| `localized faq is not published` | FAQ 存在但未发布 |
| `translation source is missing` | 翻译源语言没有有效内容 |
| `translation target is invalid` | 目标语言不受支持 |
| `translation review required` | 翻译结果尚未审核 |

对访客请求而言，“没有内容”通常不是 HTTP 错误，而是：

```text
请求成功，但没有自动回复
```

后台管理和翻译 API 可以返回明确错误，帮助管理员修复配置。

## 12. 测试矩阵

### Locale 解析

| 输入 | 期望 |
| --- | --- |
| `en` | `en` |
| `en-US` | `en` |
| `EN_us` | `en` |
| `zh` | `zh_cn` |
| `zh-CN` | `zh_cn` |
| `zh_cn` | `zh_cn` |
| `fr-FR;q=0.8` | `fr` |
| `zz-ZZ` | 不支持 |
| 空字符串 | 未确定，不触发本地化回复 |
| `*` | 新规则拒绝 |

### 规则匹配

| 场景 | 期望 |
| --- | --- |
| `en` 有规则 | 发送英文规则 |
| `fr` 无规则 | 不回复 |
| `fr` 有规则、`en` 有规则 | 只发送法语规则 |
| 只有 `*` 旧规则 | 短期不作为本地化规则匹配 |
| `zh-CN` 请求、`zh_cn` 规则 | 匹配 |
| 无法解析请求语言 | 不回复 |
| 语言正确但客服组不匹配 | 不回复 |
| 语言正确且客服组匹配 | 按优先级匹配 |

### FAQ

| 场景 | 期望 |
| --- | --- |
| `en` 规则引用 `en` FAQ | 允许 |
| `fr` 规则引用 `en` FAQ | 拒绝 |
| FAQ 为 draft | 拒绝 |
| 当前语言 FAQ 缺失 | 不发送 FAQ |
| FAQ 图片存在但 URL 非法 | 拒绝 |
| FAQ 翻译组存在但目标语言缺失 | 不回退到其他语言 |

### 翻译 API

| 场景 | 期望 |
| --- | --- |
| 只翻译缺失语言 | 已发布内容不被覆盖 |
| 翻译 API 部分失败 | 成功语言保留，失败语言可重试 |
| 重复点击翻译 | 不产生重复任务 |
| 翻译结果未审核 | 不参与运行时 |
| 源文本变化 | 旧翻译标记为过期或需要复核 |
| metadata 中的 URL 被翻译 | 后端拒绝或丢弃非法修改 |

## 13. 验收标准

短期完成标准：

- 自动回复语言不再是自由文本；
- 每条新自动回复只能绑定一个受支持语言；
- 新规则不能使用 `*`；
- `en-US`、`zh-CN` 等请求能正确映射 canonical locale；
- 当前语言没有规则时完全不回复；
- Nuxt 当前站点语言优先于浏览器 `Accept-Language`；
- FAQ 自动回复只能引用相同语言的已发布 FAQ；
- Repository 不再自行决定跨语言 fallback；
- 旧 `*` 规则有清晰的迁移状态；
- 现有英文、中文和其他已录入语言规则不会互相串语言。

长期完成标准：

- 规则和语言内容已经拆分；
- 一条规则可以维护 20 个独立语言内容；
- 后台可以显示每种语言的 draft/published/missing 状态；
- 翻译 API 只生成草稿，不直接替代人工发布；
- 缺失语言内容不会触发自动回复；
- FAQ 引用可以按翻译组解析当前语言版本；
- 每次翻译都有来源、版本、审核和发布记录；
- Nuxt 不再自行执行多语言规则选择。

## 14. 最终业务结论

本系统的正确模型不是：

```text
一条规则 + 全部语言
```

而是：

```text
一条语言无关规则
  + 多条独立语言内容
  + 每种语言独立审核和发布
```

短期在不改变数据库结构的情况下，使用：

```text
一条规则 = 一个语言 = 一份自动回复内容
```

长期再升级为：

```text
一条规则 = 触发与业务范围
一对多语言内容 = 翻译与展示内容
```

最重要的运行时原则始终不变：

> 用户当前语言没有已经录入并发布的自动回复内容，就不要自动回复。

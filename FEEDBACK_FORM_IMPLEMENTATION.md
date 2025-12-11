# 产品 / 服务反馈模块实施方案

> 范围：在 Tanzanite Setting 插件（WordPress）内实现一套结构化的「产品&服务建议」流程，并在 Nuxt 前端（如 `/company/feedback`）呈现表单。以下计划拆解得尽量细，方便团队逐项落地与跟踪。

---

## 1. 目标与成功标准

- 访客可提交包含联系方式、订单信息、产品明细与建议内容的结构化反馈。
- 数据保存在 WordPress（Tanzanite Setting）内，便于长期查询、审核和导出。
- 附件上传仅限达到指定会员等级的用户，普通登录用户仍可提交文字反馈。
- 提交内容进入后台审核后才会展示或导出，避免恶意评论直出。
- 可选通知（Email / Slack）在收到反馈时触发。
- 前端延续 Support / Company 的视觉风格，支持多语言与良好的校验体验。
- API 具备防垃圾机制（nonce/token、限流、验证码等）。

---

## 2. 数据模型与字段

| 字段                       | 类型         | 说明 |
|---------------------------|-------------|------|
| id                        | bigint (PK) | 自增主键 |
| created_at / updated_at   | datetime    | 服务器生成 |
| status                    | enum        | `new`/`in_review`/`resolved`/`archived`（默认 `new`）|
| full_name                 | varchar     | 必填 |
| email                     | varchar     | 必填并校验格式 |
| country                   | varchar     | 选填 |
| order_number              | varchar     | 选填，便于查单 |
| product_category          | enum/text   | 如 `mtb`/`road`/`gravel`/`service`/`other` |
| request_type              | enum        | `product`/`service`/`logistics`/`warranty`/`other` |
| message                   | text        | 必填，建议 200–1000 字 |
| attachments               | json        | 预留附件（CDN 地址等） |
| meta                      | json        | UA、语言、版本等扩展字段 |
| consent                   | boolean     | 同意条款记录 |

实现方式：
- **自定义数据表**（推荐）`wp_tanzanite_feedback`，字段同上。
- 或 **自定义 Post Type** `tanz_feedback` + meta，易复用 WP 后台但开销更大。

---

## 3. 后端（WordPress / Tanzanite Setting）

### 3.1 文件结构参考

```
wp-content/plugins/tanzanite-setting/
  ├─ includes/
  │   ├─ api/
  │   │   └─ class-feedback-controller.php
  │   ├─ models/
  │   │   └─ class-feedback-model.php
  │   └─ admin/
  │       └─ class-feedback-admin.php
  ├─ assets/admin/js/feedback.js
  ├─ assets/admin/css/feedback.css
  └─ tanzanite-setting.php
```

### 3.2 实施步骤

1. **数据库迁移**
   - 激活/升级钩子中创建 `wp_tanzanite_feedback`。
   - 在 `options('tz_feedback_schema_version')` 记录版本，便于后续同步。

2. **数据模型层**
   - `Feedback_Model` 负责 CRUD、状态切换、基础校验。
   - 公共的 sanitize 方法供 REST 控制器复用。

3. **REST API**
   - Namespace：`tanzanite/v1`。
   - Endpoint：`POST /feedback`（访客）、`GET /feedback`、`PATCH /feedback/<id>`、`DELETE /feedback/<id>`（需权限）。
   - Public `POST` 流程：
     - 确认用户已登录，检查会员等级：达到配置阈值（如银牌）才允许上传附件；未达到等级则忽略附件字段。
     - 校验 nonce / captcha（见安全章节）。
     - 可按 IP 设置 transient 限流。
     - 清洗数据 → 插入 → 返回 `201` 和 `id`。
   - 所有新记录默认 `status = new`，等待后台审核；管理端需 `manage_options` 权限才能查看/修改。
   - 2025-12-11：`class-rest-suggestion-feedback-controller.php` 已就绪，提供 `GET/POST /suggestion-feedback`、`PATCH /suggestion-feedback/<id>/status` 与 `GET /suggestion-feedback/eligibility`，并写入独立的 `wp_tanz_feedback_suggestions` 表；同日确认 `class-plugin.php` 在 `run()/activate()` 时加载 `Tanzanite_Suggestion_Feedback_Database`，新增 `render_suggestion_feedback()` 作为后台菜单回调，避免回调缺失的 fatal error。

4. **后台 UI**
   - 菜单：`Tanzanite → Feedback`。
   - 列表默认展示待审核项（status=new），支持筛选（状态/类型/时间），提供批量导出/更新。
   - 详情弹窗或单页展示完整记录、订单链接、联系方式。
   - 可复用现有 `WP_List_Table` 封装。
   - 已创建 `class-suggestion-feedback-admin.php`，在 `tanzanite-settings-suggestion-feedback` 页面加载 `assets/js/suggestion-feedback.js`，可查看、筛选并切换状态。

5. **通知（可选但推荐）**
   - 插入成功后触发：
     - 邮件至 `support@tanzanite.site`。
     - 可选 Slack/Teams Webhook（在设置页开关）。

6. **配置项**
   - 后台页面配置通知开关、验证码 Key、限流阈值等。
   - 保存于 `tanzanite_feedback_settings`。

### 3.3 命名与模块区隔

- 现有插件中已有“留言板”实现（REST base 为 `feedback`、数据表 `tanz_feedback`、thread 概念）。为了避免混淆：
  1. **新功能命名**：推荐使用单独的 REST base（如 `product-suggestions`）或至少使用独立的 `thread_key`（例如 `product_service_suggestion`），在代码文件名、类名上加前缀（`Suggestion_`）便于检索。
  2. **数据库**：可重用 `tanz_feedback` 表，通过 thread 区分；若需要完全隔离，可新建 `tanz_feedback_suggestions` 表。
  3. **后台菜单**：新增 `Tanzanite → Feedback Suggestions` 子菜单，与原留言列表并列，方便运营区分。
- 在 Tanzanite Setting 目录结构上，建议新增 `includes/api/class-rest-suggestion-controller.php`、`includes/admin/class-suggestion-admin.php` 等独立文件，避免和旧留言逻辑耦合。
- **故障隔离**：后台页面尽量采用独立入口（单独的 submenu + 独立渲染函数），所有异常需被 try/catch 包裹并记录到专用日志，确保即使新模块出错也不会阻塞 `tanzanite-setting` 其他功能；常量、hook、全局变量命名统一带 `suggestion_` 前缀，定位问题更快。
- **数据库升级与 REST**：2025-12-11 新增 `wp_tanz_feedback_suggestions`（`class-suggestion-feedback-database.php` 负责创建/升级，字段含会员等级、附件 JSON、审核信息）。`class-plugin.php` 在 `run()`/`activate()` 时加载 `Tanzanite_Suggestion_Feedback_Database`，注册独立控制器 `class-rest-suggestion-feedback-controller.php` 与后台菜单 `tanzanite-settings-suggestion-feedback`，并在 autoloader 未覆盖的情况下提供 shim 文件 `includes/class-suggestion-feedback-database.php`，确保线上环境也能正确加载。

---

## 4. 前端（Nuxt）

### 4.1 页面与路由

- 新建 `app/pages/company/feedback.vue`（或嵌入 Company 布局）。
- `definePageMeta({ layout: 'products' })` 复用 Company 顶部导航/背景。
- Hero 样式参考 `/support/index.vue`，保持横向菜单和标题一致。

### 4.2 组件结构

1. **Hero**：标题 + 简介。
2. **表单卡片**：
   - 字段：姓名、邮箱、国家（下拉）、订单号、产品类别（下拉）、反馈类型（按钮组）、详情（textarea）、同意勾选。
   - 附件上传区域仅在满足会员等级条件时显示，并提示仅支持指定格式/尺寸。
   - 提交中显示 loading / disabled 状态，成功后提示“已进入审核”。
3. **校验**：
   - 前端校验必填、格式、长度。
   - Tailwind/ErrorText 显示错误提示。

### 4.2.1 可复用组件

- 将表单封装为 `components/feedback/SuggestionForm.vue`（或类似命名），对外暴露 `threadKey`、`successMessage`、`showAttachments` 等 prop，供 `/company/feedback` 页面、聊天弹窗等场景复用。
- 组件内部处理：字段渲染、附件上传（仅对合资格会员显示）、表单提交、状态提示；外部只需提供 thread key、标题文案即可。
- 未来若其他页面需要嵌入反馈入口，只需引用该组件并传入相应配置即可保持一致体验。

### 4.3 API 集成

- 新建 `useFeedback()` composable（如 `app/composables/useFeedback.ts`）。
- `submitFeedback(payload)` → `$fetch('/wp-json/tanzanite/v1/feedback', { method: 'POST', body, headers: { 'X-WP-Nonce': token } })`。
- nonce 来源：SSR 注入、专用 endpoint 或 captcha token。
- 错误处理：toast/banner 呈现 WP 返回信息。

### 4.4 i18n

- 在 `i18n/locales/en.json`（及未来语言包）新增所有标签、占位、校验提示、成功/失败文案。

---

## 5. 安全与防垃圾

1. **Nonce / Token**
   - 登录用户：沿用 WP nonce。
   - 访客：集成 hCaptcha / reCAPTCHA v3，并在服务器端校验。

2. **限流**
   - 如 `transient('tz_feedback_ip_'.$hash, count, 1 hour)`，限制同 IP 高频提交。

3. **输入清洗**
   - 插入前统一 `sanitize_text_field`、`sanitize_email`、`wp_kses_post` 等。

4. **附件处理**
   - 仅当用户满足会员等级要求时才生成签名 URL。
   - 上传前在前端压缩/限制尺寸，服务器端再次确认 MIME/尺寸再写入。

---

## 6. 测试清单

- [ ] REST `POST` 在有效负载下返回 201。
- [ ] 必填/格式校验能正确阻止非法输入。
- [ ] 防垃圾机制对缺失/无效验证码给予拒绝。
- [ ] 后台可筛选/导出/修改反馈。
- [ ] 邮件/Slack 通知正常触发（若开启）。
- [ ] 前端表单支持 EN/中文，并正确切换文案。
- [ ] Lighthouse 可访问性评分通过（标签、错误提示、焦点管理等）。

---

## 7. 上线步骤

1. 将插件改动部署到 Staging，执行 DB migration。
2. Nuxt 端已完成 `SuggestionForm` 骨架并挂载 `/support/product-feedback`，后续在该组件上接入真实 API 与会员校验。
3. 使用真实数据 QA，验证后台流程与通知。
4. 上线前启用验证码密钥与限流策略。
5. 通知市场/客服团队新的反馈入口与流程。

---

## 8. 后续扩展

- 支持上传订单相关文档/照片。
- 提交反馈时自动创建 Chatwoot 工单或 CRM 记录。
- 在 WP 后台做反馈统计看板（分类/状态等）。
- 根据聚合结果生成「高频建议」公开 FAQ。

---

_实施中请持续更新本文档，记录完成情况或范围变更。最后更新：2025-12-11。_

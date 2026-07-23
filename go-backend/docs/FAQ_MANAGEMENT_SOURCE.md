# FAQ management source of truth

FAQ 页面结构与分类现在由 Go 后端数据库管理，Nuxt 前端的 `app/data/faq/pages/*` 只作为接口不可用或后台尚未录入问答内容时的静态兜底。

## Current ownership

- `faq_pages`：每个 Nuxt `PageFaq pageId` 对应一条页面元信息，包含 `route_path`、`domain`、`locale`、`title`、`subtitle`、`sort_order`、`status`。
- `faq_categories`：每个页面下的 FAQ 分类，包含 `page_id`、`category_key`、`name`、`icon`、`locale`、`sort_order`、`status`。
- `faqs`：FAQ 问答内容。`page_id` 指向页面，`category` 存储 `faq_categories.category_key`。
- `cmd/import/faqs`：批量导入也必须提供 `page_id`，并走同一套页面/分类校验，不能绕过结构事实源。
- Admin `/faqs`：负责编辑页面结构、分类结构和 FAQ 内容。
- Nuxt `PageFaq`：优先读取 `/api/v1/content/faq-pages/:page_id`；当后端没有可展示 FAQ 内容时回退静态文件。
- Nuxt `PageFaqSlot`：页面/layout 固定 FAQ 容器。它根据当前路由匹配 `faq_pages.route_path`，自动把对应 FAQ 插入页面底部；等待接口时显示 `LOAD` 状态。
- Locale：后台 FAQ 结构目前以 `en` / `zh` 管理；公开接口会把 `zh_cn`、`zh-CN` 等中文 locale 归一到 `zh`。

## Admin code responsibilities

后台 FAQ 页面按“编排、流程、接口、展示”分层，避免再次形成单个超大页面文件：

- `web/admin/src/views/FAQs.vue`：仅组装页面区块和事件，不放请求、表单校验或表格细节。
- `web/admin/src/composables/useFaqAdmin.js`：FAQ 后台状态、表单校验、权限判断、刷新流程和用户操作编排。
- `web/admin/src/api/faq.js`：FAQ 后台 HTTP 协议唯一入口；URL、方法和响应解包只能放在这里。
- `web/admin/src/components/admin/faq/FAQStructurePanel.vue`：前端页面和分类结构展示。
- `web/admin/src/components/admin/faq/FAQFilterPanel.vue`：筛选输入。
- `web/admin/src/components/admin/faq/FAQTable.vue`：FAQ 列表、批量选择和分页展示。
- `web/admin/src/components/admin/faq/FAQEditorDialog.vue`：FAQ 内容编辑。
- `web/admin/src/components/admin/faq/FAQPageEditorDialog.vue`：FAQ 页面元信息编辑。
- `web/admin/src/components/admin/faq/FAQCategoryEditorDialog.vue`：FAQ 分类编辑。

以后新增 FAQ 后台操作时，先判断它属于接口协议、流程协调还是展示组件；不要把请求和业务状态重新写进 `FAQs.vue`。

## Go service responsibilities

Go 后端 FAQ service 同样按职责拆分；这些文件都在 `internal/service` 包内，外部调用方式不变：

- `faq_service.go`：FAQService 构造、缓存失效、上传入口、FAQ 基础 CRUD、排序、搜索和批量删除。
- `faq_service_types.go`：后台输入 DTO、后台结构视图 DTO、公开 FAQ 页面 DTO。
- `faq_admin_structure.go`：后台页面结构、分类结构、页面/分类保存删除、FAQ 归属校验。
- `faq_public.go`：公开 FAQ 页面读取、按 route path 解析、公开数据组装、公开答案清洗。
- `faq_content.go`：答案 HTML 清洗、FAQ 图片边界校验、locale/status/route/category key 规范化。

新增 FAQ service 逻辑时按上面的职责进入对应文件；不要把公开展示组装、后台结构维护和内容清洗重新堆回 `faq_service.go`。

## Go handler and repository responsibilities

- `internal/api/admin/faq_handler.go`：FAQ handler 构造与共享错误分类。
- `faq_requests.go`：后台 HTTP request DTO。
- `faq_items_handler.go`：FAQ 条目 CRUD、列表、排序和批量删除。
- `faq_structure_handler.go`：FAQ 页面与分类结构接口。
- `faq_media_handler.go`：FAQ 答案图片上传和文件边界校验。
- `internal/repository/faq_repository.go`：FAQRepository 构造。
- `faq_items_repository.go`：FAQ 条目、搜索、排序、浏览数和公开查询 SQL。
- `faq_structure_repository.go`：FAQ 页面、分类、分类同步与统计 SQL。

所有文件仍在原有 Go package 内，路由、HTTP 请求/响应字段、service/repository 方法名均保持不变。后续新增 SQL 或 handler 时，按“条目、结构、媒体”归属进入对应文件。

## Nuxt storefront responsibilities

Nuxt FAQ 用户端按“自动插入、通用 FAQ 容器、分类折叠 UI、答案渲染、数据来源”分层：

- `app/components/PageFaqSlot.vue`：layout 内的自动插入容器，只负责根据当前 route 查找 FAQ 页面和显示 `LOAD`。
- `app/components/PageFaq.vue`：通用 FAQ 区块容器，只负责标题、空状态、查看全部入口和分类列表装配。
- `app/composables/usePageFaq.ts`：PageFaq 的取数、静态 fallback、展开状态、`maxItems` 截断和 `hasMoreItems` 计算。
- `app/components/faq/FaqCategoryAccordion.vue`：单个 FAQ 分类卡片和问题折叠 UI。
- `app/components/FaqAnswerContent.vue`：答案轻量 HTML 与单张 FAQ 图片渲染。
- `app/data/faq/registry.ts`：静态 FAQ fallback 注册表、pageId 与 route path 映射、聚合静态条目。
- `app/data/faq/routing.ts`：route path 归一化和 locale 前缀剥离。
- `app/data/faq/backend.ts`：Go 后端 FAQ 请求、公开结构接口优先、旧内容接口兜底转换。
- `app/data/faq/index.ts`：FAQ 数据层统一出口，不承载业务逻辑。

新增 Nuxt FAQ 能力时，先判断是 layout 插入、容器展示、分类折叠、答案渲染、静态 fallback 还是后端请求；不要把请求和数据转换重新写进 `PageFaq.vue`。

## Storefront insertion model

Nuxt 页面不再长期手写 `<PageFaq page-id="..." />`。页面只负责自己的主体内容，FAQ 由 layout 内的固定容器统一插入。

Flow:

1. `PageFaqSlot` 读取当前 route path，并依据 Nuxt 的 locale manifest 移除 locale 前缀，例如 `/zh_cn/support/payment` -> `/support/payment`；新增语言不需要再手写一份 FAQ 路由正则。
2. `PageFaqSlot` 请求 Go backend 的 route resolver，通过 `faq_pages.route_path` 找到对应 `page_id`。
3. 如果后台返回可展示 FAQ 内容，slot 直接渲染 `PageFaq`。
4. 如果后台结构存在但还没有 FAQ 内容，Nuxt 可回退本地静态 FAQ 文件；后台仍然能看到该页面和分类结构。
5. 如果后台和静态 fallback 都没有内容，用户端不显示空 FAQ 区块。

Loading rule:

- 初始 SSR 能拿到内容时直接渲染 FAQ。
- 客户端路由切换或接口等待时显示 `LOAD` 占位。
- `LOAD` 的最终视觉样式以后按设计稿替换，组件入口保持不变。

## Answer content boundary

FAQ 答案不是通用文章 CMS。长期只支持轻量答案内容，避免后台录入把 Nuxt 页面排版撑乱。

- `faqs.answer`：唯一的答案文本事实源，保存后端清洗后的轻量 HTML。允许段落、换行、加粗、斜体和链接。
- `faqs.answer_image_url`：可选的一张 FAQ 专用图片，不允许把图片写进 `answer` HTML。
- `faqs.answer_image_alt`：图片替代文本。
- 每条 FAQ 最多一张答案图片。
- FAQ 图片必须是 `image/webp`，并且必须是固定 `800 x 800` 像素、严格 `1:1`。
- 后台上传前给出明确提示；前端可以做即时校验，但 Go 后端必须再次读取真实图片内容并拒绝不符合规则的文件。
- 不是 WebP、不是 800 x 800、不是 1:1 的图片不能进入 FAQ 图片事实源。
- Nuxt FAQ 容器必须由内容自动撑高，不能再用固定 `max-height` 截断展开内容。
- Nuxt 桌面端固定为图片列 + 文本链接列；移动端图片在上、答案在下。

Allowed answer HTML:

- `p`
- `br`
- `strong`
- `em`
- `a`

Allowed link protocols:

- `https:`
- `http:`
- `mailto:`
- site-relative paths such as `/support/shipping`

Forbidden in `answer`:

- `img`
- `script`
- `style`
- `iframe`
- event handlers such as `onclick`
- custom classes, inline styles, arbitrary font sizes, arbitrary colors

## Implementation status

当前阶段先收口结构，不做页面级零散补丁：

- 已新增 backend FAQ 页面/分类事实源，后台可以看到 Nuxt 现有 FAQ 页面和分类，即使该页面暂时没有 FAQ 问答内容。
- 已新增公开 route resolver，Nuxt 可以按当前路由查询 FAQ 页面结构，不再依赖页面内硬编码 `pageId`。
- 已新增 Nuxt `PageFaqSlot`，作为固定插入容器。`products` 和 `support` layout 先接入该容器，覆盖当前已有 FAQ 页面所在的 storefront 区域。
- 正在清理 Nuxt 页面内旧的手写 `<PageFaq />`，避免同一页面被 layout slot 和页面自身重复渲染 FAQ。
- 暂不把 `PageFaqSlot` 放到 `default` / `spokecalc` layout；这些 layout 目前没有稳定的 FAQ 插入需求，等页面域确认后再扩展。
- 已新增 FAQ 轻量答案边界：后台答案编辑支持轻量 HTML，Go 后端保存和公开读取时统一清洗，图片从答案 HTML 中拆出为 FAQ 专用单图字段。
- 已新增 FAQ 图片上传边界：后台和 Go 后端都校验 `image/webp`、`800 x 800`、最多一张；Nuxt 用固定答案内容组件渲染桌面两列/移动单列，并取消 FAQ 展开内容的固定 `max-height` 截断。

下一次涉及 FAQ 后台、FAQ 页面结构、Nuxt 页面 FAQ 自动插入、`LOAD` 样式时，必须同步更新本文件，记录完成范围和仍未完成的部分。

## Initial seed

Migration `028_faq_page_category_source.up.sql` seeds the existing Nuxt FAQ page/category structure for `en` and `zh`, so a fresh admin panel is not empty even before FAQ items are added.

## Update rule

Whenever a new page uses `PageFaq pageId="..."` on the Nuxt storefront, update these places together:

1. Add or edit the matching `faq_pages` / `faq_categories` records through Admin, or add a migration seed if it must exist in every environment.
2. Keep Nuxt static fallback data in `nuxt-i18n/app/data/faq/pages/*` only as fallback content.
3. If the backend public response shape changes, update `nuxt-i18n/app/data/faq/index.ts` at the same time.

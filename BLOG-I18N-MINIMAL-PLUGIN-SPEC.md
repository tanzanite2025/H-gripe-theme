# 自研最小多语言 Blog 插件（WP）设计清单

> 目标：在 **不依赖付费插件**、不引入臃肿功能的前提下，让 Nuxt3 的 `/[locale]/blog` 可以按当前语言读取 WordPress 的文章内容，并支持“同一篇文章不同语言 slug 不同”。
>
> 约束：文章量少（每月 2-3 篇），图文简单；重点是“翻译关联 + 干净 API”。

---

## 1. 关键结论（你要的核心能力）

- **UI i18n（Nuxt）** 与 **内容翻译（Blog Posts）** 是两件事。
- Nuxt i18n 只能决定 `locale`，文章内容必须在 WP 侧按语言存在。
- **允许不同语言 slug 不同** 时，不能用 slug 做跨语言的唯一标识，必须引入“翻译组（translation group）”来关联多语言版本。

## 1.1 已确认的固定规则（本方案不再讨论分支）

- **插件代码目录**：`C:\Users\P16V\Desktop\Wordpress\tanzanite-theme\wp-plugin`
- **语言 code 列表来源**：完全复用 Nuxt i18n 的 locale code（以 Nuxt locales 为准）
- **频道路径**（所有语言固定，不翻译）：
  - `/:locale/blog/news`
  - `/:locale/blog/wheelsbuild`
- **分类 slug**（固定，不翻译）：`news` / `wheelsbuild`
- **文章范围过滤**：必须同时满足
  - 有 `tz_lang`
  - 且属于分类 `news` 或 `wheelsbuild`
- **分类创建**：插件自动创建 `news` / `wheelsbuild`（若不存在）
- **创建翻译时复制内容**：标题、正文、摘要、特色图、分类、tags
- **详情正文返回格式**：`contentHtml`（接受 WP 渲染后的 HTML 输出）
- **缓存策略**：启用 transient 缓存

## 1.2 当前实现进度（已落地代码）

- 插件目录已创建：`tanzanite-theme/wp-plugin/tanzanite-blog-i18n/`
- 已实现：
  - `tz_lang` taxonomy（文章语言）
  - `tz_translation_group` meta（翻译组 UUID）
  - 后台 metabox：Create translation / Link existing translation
  - REST API：`/posts`、`/post`、`/translations`
  - 自动创建分类：`news` / `wheelsbuild`（若不存在）
  - 自动创建语言 terms：与 Nuxt locales 一致
  - transient 缓存 + 版本号失效（发布/更新文章后自动失效）

- Nuxt 前端（已落地：先用 mock 数据，后续替换为 WP REST）
  - 路由与页面：
    - `/:locale/blog`（All 列表）
    - `/:locale/blog/news`（News 列表）
    - `/:locale/blog/wheelsbuild`（Wheelsbuild 列表）
    - `/:locale/blog/news/:slug`（News 详情，slug 可自定义）
    - `/:locale/blog/wheelsbuild/:slug`（Wheelsbuild 详情，slug 可自定义）
    - `/:locale/blog/:slug`（All 详情入口，保留；建议最终 canonical 走分类详情页）
  - 列表页交互：
    - 默认展示 5 条
    - `Read more` 每次追加 5 条
    - 不使用 `page2` / `?page=` 形式分页 URL
  - 二级 Tab（All / News / Wheelsbuild）
    - 进入详情页时 Tab 不动，保持正确高亮
  - SEO / hreflang
    - 详情页根据 `translations` 映射生成每个语言对应 slug 的 `hreflang`（解决不同语言 slug 不同的问题）
  - i18n
    - `en.json` 已补齐 Blog 页面标题/简介/按钮/空状态文案
  - 相关文件（Nuxt）
    - `tanzanite-theme/nuxt-i18n/app/utils/blogMock.ts`（mock 数据与查询函数）
    - `tanzanite-theme/nuxt-i18n/app/pages/blog/*`（列表与详情页）
    - `tanzanite-theme/nuxt-i18n/app/components/ProductsTopNav.vue`（Tab 高亮逻辑）
    - `tanzanite-theme/nuxt-i18n/app/layouts/default.vue`（`hreflang` override 支持）
    - `tanzanite-theme/nuxt-i18n/i18n/locales/en.json`（Blog 文案）

---

## 1.3 部署方式（你的工作流）

- 本仓库的 `tanzanite-theme/wp-plugin` 用于管理插件源码。
- 上线时把 `tanzanite-blog-i18n/` 整个目录复制到服务器：
  - `wp-content/plugins/tanzanite-blog-i18n/`
  - 然后在 WP 后台启用。

---

## 2. 插件边界（MVP）

### 2.1 MVP 支持

- **语言管理（最小）**：维护 34 个语言 code（与 Nuxt locale 对齐）。
- **文章语言标记**：每篇文章属于一个语言。
- **翻译关联**：同一篇文章的多语言版本属于同一个翻译组。
- **自定义 REST API**：提供给 Nuxt 读取：列表、详情、翻译映射。
- **只读发布内容**：默认只输出 `publish` 状态文章（不暴露草稿）。

### 2.2 不做（明确不做，保持“瘦”）

- 不做自动机器翻译。
- 不做 WooCommerce、多语言菜单、多语言小工具。
- 不做分类/标签翻译（MVP 先用统一分类 slug，比如 `news`/`wheelsbuild`）。
- 不做媒体库多语言（特色图可直接用当前文章的 featured image）。
- 不做复杂 SEO 面板（Yoast/RankMath 之类）。

---

## 3. 数据结构（最小可行）

建议使用 **WP 原生 Post**（`post`），不新建 CPT，减少复杂度。

### 3.1 语言字段

- 自定义 taxonomy `tz_lang`
  - Term slug：**完全复用 Nuxt i18n 的 locale code**（例如 `en`, `fr`, `de`, ...）
  - 每篇文章必须选 1 个语言

### 3.2 翻译组字段

- post meta：`tz_translation_group`（string）
  - 同一篇文章的各语言版本共享同一个 group id
  - group id 生成策略：UUID（v4）

### 3.3 文章基本字段

仍使用 WP 原生：
- `post_title`, `post_name`（slug）, `post_excerpt`, `post_content`, `post_date`
- `featured image`（`get_post_thumbnail_id()`）
- 分类（Categories）：
  - MVP 建议仅两类：`news` / `wheelsbuild`（固定 slug）

---

## 4. 后台编辑体验（最少但够用）

### 4.1 发布前必填

- 语言（`tz_lang`）：必须选。
- 分类：必须选 `news` 或 `wheelsbuild`（slug 固定不翻译）。

### 4.2 翻译关联 Meta Box（核心）

在文章编辑页右侧增加一个 meta box：

- 显示当前文章：语言 + group id
- 显示该 group 下已存在的翻译版本列表（按语言 code）：
  - `en → Post #123 / slug`
  - `fr → Post #456 / slug`
- 操作按钮：
  - **Create translation**（选择目标语言）
    - 行为：复制当前文章生成新草稿
    - 自动写入：同一个 `tz_translation_group`
    - 自动设置：目标语言 term
    - 复制内容：标题、正文、摘要、特色图、分类、tags
  - **Link existing translation**（选择已有文章并指定语言）
    - 用于你已经手工创建文章后，再把它归组

> 备注：
> - 文章 slug 允许不同：翻译版本不需要共享 slug。
> - 你只要确保同组多语言文章的 `tz_translation_group` 相同。

---

## 5. REST API 设计（给 Nuxt 用，字段尽量少）

统一 namespace：`/wp-json/tanzanite/v1`

### 5.1 列表：按语言 + 分类

`GET /wp-json/tanzanite/v1/posts`

Query：
- `lang`（必填）：例如 `fr`
- `category`（可选）：分类 slug，仅允许 `news` / `wheelsbuild`（不传表示不按分类过滤）
- `page`（可选，默认 1）
- `per_page`（可选，默认 10，最大 50）

返回（建议）：
```json
{
  "page": 1,
  "per_page": 10,
  "total": 123,
  "items": [
    {
      "id": 456,
      "lang": "fr",
      "group": "a1b2c3",
      "slug": "comment-choisir-des-jantes",
      "title": "...",
      "excerpt": "...",
      "date": "2025-12-01T12:00:00Z",
      "featuredImage": {
        "url": "https://tanzanite.site/wp-content/uploads/...jpg",
        "width": 1200,
        "height": 630,
        "alt": "..."
      },
      "categories": ["news"],
      "translations": {
        "en": {"id": 123, "slug": "how-to-choose-rims"},
        "fr": {"id": 456, "slug": "comment-choisir-des-jantes"}
      }
    }
  ]
}
```

### 5.2 详情：按语言 + slug

`GET /wp-json/tanzanite/v1/post`

Query：
- `lang`（必填）
- `slug`（必填）

返回：
- 同列表字段
- 增加：
  - `contentHtml`（WP 渲染后的 HTML，Nuxt 直接渲染）
  - `canonicalUrl`：当前语言文章的 canonical

### 5.3 翻译映射：按 group

`GET /wp-json/tanzanite/v1/translations`

Query：
- `group`（必填）

返回：
```json
{
  "group": "a1b2c3",
  "translations": {
    "en": {"id": 123, "slug": "how-to-choose-rims"},
    "fr": {"id": 456, "slug": "comment-choisir-des-jantes"}
  }
}
```

---

## 6. Nuxt 侧的对接方式（不写代码，先定契约）

### 6.1 路由建议

- `/:locale/blog`：All 列表页
- `/:locale/blog/news`：频道列表（**所有语言固定使用 `news`，不翻译**）
- `/:locale/blog/wheelsbuild`：频道列表（**所有语言固定使用 `wheelsbuild`，不翻译**）
- `/:locale/blog/news/:slug`：文章详情（News）
- `/:locale/blog/wheelsbuild/:slug`：文章详情（Wheelsbuild）
- `/:locale/blog/:slug`：可选的详情入口（All），建议最终 canonical 走分类详情页

### 6.2 获取策略（与你“写完就不改”匹配）

- 插件侧只提供 REST API（内容源），不绑定 Nuxt 的渲染模式。
- 前端推荐先用 **Runtime 拉取**：新增文章不需要改 Nuxt 代码。
- 如果未来要更强 SEO/速度，再做 SSG/预渲染即可（属于 Nuxt 构建策略，不影响插件契约）。

### 6.3 语言切换（slug 不同）

- 详情页 API 返回 `translations` 映射
- 切换语言时：
  - 找到目标语言的 `{ slug }`
  - 跳转到 `/:targetLocale/blog/:targetSlug`

---

## 7. 安全与权限（最小）

- REST API 只返回：
  - `post_status=publish`
  - `post_type=post`
- 过滤参数：
  - `lang` 必须在允许列表中
  - `per_page` 限制最大值
- 避免输出敏感字段：作者邮箱、内部 meta 等

---

## 8. 性能与缓存（确定启用）

- 列表和详情可以做轻量缓存：
  - 缓存：`transient`
  - TTL：默认 900 秒（15 分钟）
  - key：内部会带一个全局版本号前缀（文章保存/删除/回收站时自动 bump 版本号，实现整体失效）

---

## 9. 插件目录结构（已实现）

```
/wp-content/plugins/tanzanite-blog-i18n/
  tanzanite-blog-i18n.php
  /includes/
    class-blog-admin.php
    class-blog-cache.php
    class-blog-languages.php
    class-blog-rest.php
    class-blog-setup.php
```

---

## 10. 开发清单（进度）

### 已完成

- [x] 注册 `tz_lang` taxonomy
- [x] `tz_translation_group` 写入与读取（UUID v4）
- [x] Admin meta box：展示 group 与 translations
- [x] Create translation：复制文章生成草稿，设置目标语言 + 同组（复制分类 + tags）
- [x] Link existing translation：把已有文章加入翻译组并指定语言
- [x] REST：`/posts`、`/post`、`/translations`
- [x] 自动创建分类：`news` / `wheelsbuild`
- [x] 自动创建语言 terms：与 Nuxt locales 一致
- [x] transient 缓存 + 版本号失效

- [x] Nuxt 前端：All/News/Wheelsbuild 列表页
- [x] Nuxt 前端：详情页路由（`/blog/news/:slug`、`/blog/wheelsbuild/:slug`、`/blog/:slug`）
- [x] Nuxt 前端：列表页 Load more（每次 5 条，不做分页 URL）
- [x] Nuxt 前端：二级 Tab 高亮在详情页保持正确
- [x] Nuxt 前端：详情页 `hreflang` 按 `translations` 生成（slug 可不同）
- [x] Nuxt 前端：补齐 `en.json` Blog 文案 key

### 待做（如需要再加，暂不影响 Nuxt 承接页）

- [ ] 保存文章时强制阻断：未选择 `tz_lang` 或未选择分类时不允许发布

---

## 12. 最小验收标准（上线前必测）

- [ ] `/wp-json/tanzanite/v1/posts?lang=en&category=news` 返回正常
- [ ] `/wp-json/tanzanite/v1/post?lang=fr&slug=xxx` 返回正常
- [ ] 同一 group 的 translations 映射正确，且 slug 可不同
- [ ] Nuxt 切换语言时能跳到对应语言文章
- [ ] 无草稿/私密内容泄露

---

## 13. 下一步（先做 Nuxt 承接页，再统一联调）

- Nuxt 承接页（已完成：目前使用 mock 数据，待替换为 WP 插件 API）：
  - `/:locale/blog`
  - `/:locale/blog/news`
  - `/:locale/blog/wheelsbuild`
  - `/:locale/blog/news/:slug`
  - `/:locale/blog/wheelsbuild/:slug`
  - `/:locale/blog/:slug`

- 下一步：替换 mock 为真实 WP REST + 联调验收：
  - 列表：`/wp-json/tanzanite/v1/posts?lang=en&category=news`
  - 详情：`/wp-json/tanzanite/v1/post?lang=en&slug=xxx`
  - 翻译映射：`/wp-json/tanzanite/v1/translations?group=xxx`

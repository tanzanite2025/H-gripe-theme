# 首页框架（模板 A）+ SEO 落地指南

## 目标

- 一次性把首页的「板块顺序 + 语义结构 + 组件化方式」定下来
- 后续维护尽量变成：
  - 替换图片（上传到 `public/` 或你们既定的资源目录）
  - 改文案（走 i18n 资源文件）
  - 改链接（内部路由）
- 保证 SEO 友好：
  - 重要内容在服务端输出（SSR/SSG 皆可）
  - 清晰的标题层级（H1/H2/H3）
  - 可抓取的内部链接（`href` / `NuxtLink`）
  - 图片 `alt`、性能（LCP/CLS）与基础结构化数据

---

## 当前首页现状（`nuxt-i18n/app/pages/index.vue`）

目前首页主要是这些块：

- Hero：`HomeHero` ✅
- Shop with Confidence：`HomeShopWithConfidence` ✅
- Why Choose Us：`HomeWhyChooseUs` ✅
- Features：`HomeFeatures` ✅
- Factory Stories：`HomeFactoryStories` ✅
- Featured Products：`HomeFeaturedProducts` ✅
- Innovation & R&D：`HomeInnovationRd` ✅
- FAQ 预览：`HomeFaqPreview`
- 联系/地图：`ContactLocationMap`
- Final CTA：`HomeFinalCta` ✅

### SEO/H1 校验结果（2025-12-15）

- 首页仅保留一个主 `h1`：`HomeHero`（作为页面唯一 H1）
- `SiteHeader` 的站点 Logo 不再使用 `h1`（改为普通容器元素），避免全站每个页面额外出现一个 H1
- 首页各大板块标题使用 `h2`；卡片标题使用 `h3`（层级递进清晰）
- `title/description` 由首页 `useHead` 提供；canonical/alternate 等全局链接与 `titleTemplate` 由 `layouts/default.vue` 的 `useHead` 统一管理

当前仍待补齐的块（按模板 A）：

（首页骨架已补齐，后续进入逐块调样式与填充真实内容/链接）

已处理的关键实现点（与布局相关）：

- Hero 顶部避让：通过 `SiteHeader` 自动计算高度并写入 `--site-header-offset`，`HomeHero` 使用该变量作为顶部 padding，避免遮挡
- Hero 右侧媒体：桌面端两张图占位；移动端媒体上移到 CTA 之前
- CTA：移动端 3 按钮同一行（Shop / About / Gallery）；桌面端保持主/次按钮 + 第三入口（Gallery 文本链接）
- 视觉：按钮/图片去掉白色边框，使用右下偏移纯黑色硬阴影（范围已收紧）
- Features / Factory Stories：已接入首页作为占位块（文案走 `en.json`，链接后续补齐）

---

## 你选定的首页模板 A（标准顺序）

从上到下建议顺序：

1. **Hero 首屏**（H1 + 核心卖点 + CTA）
2. **Trust / Social Proof**（信任背书：支付/物流/质保/口碑/媒体等）
3. **Features / Why Choose Us**（核心卖点要点化）
4. **Factory Stories（你的 4 卡片块）**
5. **Featured Products / Categories**（主转化入口：精选产品/分类）
6. **Innovation & R&D**（研发与方向：与 Why Choose Us 不重复）
7. **FAQ Preview**（减少疑虑）
8. **Contact / Location**（门店/联系/地图）
9. **Final CTA**（最后一次行动召唤）

> 实际落地可按阶段推进：先把 1/2/4/7/8 做对（结构与 SEO），再补 3/5/6/9。

---

## SEO 总体原则（适用于所有板块）

### 渲染策略：SSR vs SSG

- **SSG**：更“保险”，稳定、速度快、CDN 友好，几乎不受运行时波动影响
- **SSR**：只要确保核心内容服务端输出（不 `ClientOnly`，不依赖客户端 JS 才出现），SEO 同样友好

**结论**：对 SEO 来说关键不是“组件化/页面内写死”，而是**最终 HTML 是否包含核心内容**。

### 标题层级（强约束）

- 首页必须有且仅有一个核心 `h1`（Hero）
- 每个大板块标题用 `h2`
- 板块里的卡片标题用 `h3`
- 全局 `SiteHeader` 的 Logo 不使用 `h1`（已落实），避免每个页面出现额外的 H1

### 链接策略

- 首页的每个卡片/入口尽量链接到站内的实体页面（`NuxtLink`）
- 避免 `href="#"`（除非确实是锚点），否则对 SEO 与可用性都不友好

### 图片策略（性能 + 可访问性）

- 首屏大图（LCP）避免懒加载；其余图片 `loading="lazy" decoding="async"`
- 图片必须有 `alt`（装饰图可 `alt=""` 并 `aria-hidden="true"`）
- 尽量避免 CLS：明确图片尺寸/容器比例（aspect ratio）

### i18n（内容可维护）

- 文案不要硬编码，落到 `en.json`（前端）
- 这样后续你只改 i18n 文案即可，不需要动组件结构

### 结构化数据（可选，后期加分项）

- Organization / WebSite（基础）
- 如有产品强转化：首页可以逐步引入 Product / AggregateRating（谨慎，需真实数据）

---

## 每个板块怎么做更利于 SEO（落地细则）

### 1) Hero（首屏）

**目的**：用最少字解释你是谁 + 你提供什么 + 你与竞品的关键差异，并提供主 CTA。

- **语义结构**
  - `h1`：品牌/核心主张
  - `p`：补充一句解释（包含核心关键词，但不要堆砌）
  - CTA：3 个入口（主按钮 + 次按钮 + 第三入口）
- **SEO/可用性要点**
  - `h1` 里自然包含类目词（例如 wheelset / carbon / etc）
  - CTA 用站内链接（`NuxtLink`）
  - 如果有首屏图：保证不会导致 CLS

#### Hero 推荐布局（方案 B：左文案 + 右侧大图）

- **桌面端（lg+）**
  - 左侧：`h1` + 描述 + CTA
  - 右侧：两张图占位（后续替换为真实图片），并与左侧文案区域等高
- **移动端（<lg）**
  - 文案优先显示（单列），图片位于 CTA 之前（两张图一行两列）
  - 不要分裂成两套文案；如需要差异化，仅对“媒体区”做分端布局

#### Hero 文案（已定稿）

- **H1**
  - `Factory-Direct Carbon Rims & Wheelsets`
- **Subhead**
  - `In-house carbon layup and QC—engineered for speed, stiffness, and reliability.`

#### Hero CTA（已确认路径与层级）

- **主 CTA（Primary）**
  - 目标：`/shop`
  - 建议样式：实心按钮（最高对比度）
- **次 CTA（Secondary）**
  - 目标：`/company/about`
  - 建议样式：描边/幽灵按钮
- **第三入口（Tertiary）**
  - 目标：`/picture-warehouse`
  - 文案：`Gallery`
  - 建议样式：桌面端文本链接；移动端可与主/次同排显示（按钮更紧凑）

### 2) Trust / Social Proof（Shop with Confidence）

- **语义结构**
  - `section` + `h2`
  - 4-6 个要点卡：包邮/质保/退换/评价等
- **SEO要点**
  - 这些点建议用纯文本（爬虫可读）
  - 如果每个 trust item 有详情页，可链接到 `/support/...` 页面

### 3) Features / Why Choose Us（要点化）

- **语义结构**
  - `section` + `h2` + 3-6 个 feature
  - 每个 feature：`h3` + 1-2 句描述
- **SEO要点**
  - 用“用户语言”描述（解决什么痛点），不是纯参数堆叠

#### 第一版推荐（性能 + 制造能力，兼顾 B2C/B2B）

- **布局**
  - 桌面端：4 张等高卡片（2x2 或 4 列）
  - 移动端：单列堆叠（4 张）
- **链接策略（第一版）**
  - 暂不为卡片添加内链，待对应内容页完成后再补

##### 卡片文案（已确认 3/4）

1. **Consistent QC, Reliable Results**
   - 强调质量控制与稳定交付（可靠性 / 质量控制）
2. **Road / Gravel / MTB Options**
   - 强调多场景覆盖与规格选择（场景覆盖）
3. **OEM/ODM & Custom Layup**
   - 强调工厂能力与定制化（B2B 重点）
4. **Lightweight Without Compromise**
   - 轻量化与强度平衡（偏 B2C 性能）

### 4) Factory Stories（你的 4 卡片块）

- **语义结构**
  - `section` + `h2`（Factory Stories）
  - 卡片：`NuxtLink`（整卡可点击）
  - 卡片内：`h3` + `p` + “Read more”
- **SEO要点**
  - 链接必须指向真实内容页（例如 `/blog`、`/blog/news`、`/blog/wheelsbuild` 或具体文章）
  - 卡片上的图如果是装饰背景：不必强行 `<img>`，但如果是真实内容图，建议用 `<img alt>`

### 5) Featured Products / Categories（强转化）

- **语义结构**
  - `section` + `h2`
  - 产品卡：名称（`h3`）、价格、关键卖点、链接到详情页
- **SEO要点**
  - 内链到产品详情是 SEO 的核心资产
  - 价格/库存这种动态内容：SSR/SSG 需权衡（后期可做 ISR/SSR）

### 6) Innovation & R&D（替代 Testimonials）

- **语义结构**
  - `section` + `h2`
  - 3 个研发卡：每张卡 `h3` + `p` + bullets
- **SEO要点**
  - 该区块更适合“方向与能力展示”，避免与 Why Choose Us 重复
  - 若后续要做详情页，建议把卡片底部的 “Read more” 改为 `NuxtLink` 指向真实页面（纯 `button` 对 SEO 内链没有帮助）

### 7) FAQ Preview（你已存在 HomeFaqPreview）

- **语义结构**
  - `section` + `h2` + FAQ 列表
- **SEO要点**
  - FAQ 的内容最好是纯文本（你现在是 `v-html`，需要确保内容安全与可控）
  - 若后期要做 FAQ schema，可以基于固定数据源生成 JSON-LD

### 8) Contact / Location（你已存在 ContactLocationMap）

- **SEO要点**
  - 地图 iframe 使用 `ClientOnly` 是合理的（地图不是 SEO 核心）
  - 关键是：地址、营业信息、联系方式要用文本输出（你已经有）

### 9) Final CTA

- **目的**：给还没下决心的用户一个最后的明确行动。
- **SEO要点**
  - CTA 不直接提升 SEO，但提升转化与停留时间

---

## 推荐的工程落地方式（一次性搭好骨架，后续只换图/文案）

### 组件拆分（建议）

建议新增一组首页专用组件（名称仅建议，可调整）：

- `app/components/home/HomeHero.vue`
- `app/components/home/HomeShopWithConfidence.vue`
- `app/components/home/HomeWhyChooseUs.vue`
- `app/components/home/HomeFeatures.vue`
- `app/components/home/HomeFactoryStories.vue`
- `app/components/home/HomeFeaturedProducts.vue`
- `app/components/home/HomeInnovationRd.vue`
- `app/components/home/HomeFaq.vue`（或复用 `HomeFaqPreview`）
- `app/components/home/HomeContact.vue`（复用 `ContactLocationMap`）
- `app/components/home/HomeFinalCta.vue`

首页 `pages/index.vue` 只负责按顺序“拼装”。

### 内容与图片的维护方式（让后续操作尽量简单）

推荐两条路，二选一：

1. **配置文件驱动（推荐）**
   - 新建 `app/data/homepage.ts`（或你们约定的 data 目录）
   - 里面导出结构化数据：标题、描述、卡片数组、图片路径、链接
   - 组件只负责渲染，不写死内容

2. **i18n + 少量常量**
   - 文案走 `en.json`
   - 图片路径/链接在组件里用常量数组（后续改动也很小）

### Meta（`useHead`）建议

- 首页 `title` / `description` 用专门的 i18n key
- OG 也同理
- 不要复用 `t('welcome')` 作为 description（太泛）

---

## 下一步：我们如何按这份 MD 一次性完善首页

建议按阶段推进：

1. **先搭骨架与语义**（不追求最终图片）
   - 补 Hero（H1）
   - 插入 Shop with Confidence（4 卡片，占位文案可后改）
   - 插入 Factory Stories（4 卡片块）
   - 保留 FAQ 与 Contact
   - 清理 Demo 内容（已移除 TWCarousel）

2. **再补齐模板 A 的其余块**
   - Features / Featured Products / Innovation & R&D / Final CTA

3. **最后做内容运营化**
   - 把内容抽到 data/i18n
   - 你就只需要上传图片 + 改文案

---

## 需要你确认的信息（我们才能把首页一次性按模板 A 落地）

1. Hero 的主标题（H1）你希望写什么？（中文/英文均可）
2. Hero 的主图素材来自哪里？（本地图片路径/图片库里的哪几张）
3. 4 个 Case Studies 卡片最终链接到哪些页面？（`/blog` / 分类页 / 文章页）

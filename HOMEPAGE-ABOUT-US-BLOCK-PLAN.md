# 首页 About Us 区块备忘（待 About 页完善后实施）

本文档用于**防止遗忘**：等 `/company/about`（About Us 分页）内容完善后，把其中的核心内容抽出来，作为**首页（Home）一个独立区块**展示，并确保后续维护只改一处。

---

## 目标

- 首页新增一个 **About Us 内容块**（用于讲清楚我们是谁/工厂与能力/质量与流程等）
- 首页的 About 内容尽量**复用 About 分页**（避免写两套文案/两套图片）
- 内容后续可扩展，但应保持：
  - **单一数据源**（文案/图片路径维护一份）
  - 首页与 About 页共享 UI/数据（通过组件/props 变体复用）

## 非目标

- 现在立刻实现该区块（About 页内容未最终确定）
- 做复杂动效/长滚动叙事（优先 MVP，可先做“预览版 + CTA 跳转 About 页”）

---

## 仓库现状（入口位置）

- About 分页：
  - `nuxt-i18n/app/pages/company/about.vue`
- 首页：
  - `nuxt-i18n/app/pages/index.vue`

首页当前已有多个 Section：`TWCarousel`、`TrustCards`、`HomeFaqPreview`、`ContactLocationMap`。

---

## 推荐实现原则（后续正式做时按这个走）

### 1) 单一数据源（必须）

不要把 About 文案/图片在首页再写一份。推荐把 About 的“结构化内容”放到一个 util 文件里（或拆成多文件），由 About 页与首页同时引用。

建议路径（二选一即可）：

- 方案 A（推荐）：
  - `nuxt-i18n/app/utils/aboutUsContent.ts`
  - 导出一个对象/数组，包含：标题、简介、tab 列表、每个 tab 的段落与图片列表等。
- 方案 B（更 i18n 化）：
  - 文案放 `en.json`（frontend i18n），图片与结构放 `aboutUsContent.ts`

> 注：当前 `company/about.vue` 里有大量英文硬编码内容；等你要正式上首页时，再一起整理成 i18n key，避免后面反复替换。

### 2) 抽可复用组件（必须）

把 About 区块抽成组件，并提供 `variant`（或 `compact`）来适配首页的“预览”展示。

推荐组件：

- `nuxt-i18n/app/components/HomeAboutUsSection.vue`
  - `variant: 'home' | 'page'`（或 `compact: boolean`）
  - 读取 `aboutUsContent.ts`（或由父组件传入）
  - 统一渲染标题/简介/图片卡片/CTA

然后：

- About 页 `company/about.vue`：渲染 `variant="page"`
- 首页 `pages/index.vue`：渲染 `variant="home"`（只展示精简版 + CTA）

### 3) 首页展示策略（建议从 MVP 开始）

首页不要把 About 分页完整内容都塞进来，建议：

- 展示：
  - 1 个标题（About Us）
  - 1 段简介（2-3 行）
  - 3 张图片/卡片（例如 Factory / Quality / Manufacture 三个亮点）
  - 1 个 CTA：`Learn more` 跳转 `/company/about`

后续如果你觉得首页需要更丰富，再逐步加内容（保持复用架构不变）。

---

## 实施步骤（等 About 页内容确认后再做）

### Step 1：确认 About 页最终内容结构

在 `nuxt-i18n/app/pages/company/about.vue` 中确认：

- 最终要保留哪些 tab（目前有：Factory / Appearance / Facility / Manufacture / Quality control）
- 哪些段落/图片要复用到首页

### Step 2：落地单一数据源

新建：

- `nuxt-i18n/app/utils/aboutUsContent.ts`

内容建议包含：

- `title`
- `description`
- `highlights`（用于首页）
- `tabs`（用于 About 页）

### Step 3：抽组件并让 About 页先用起来

新建：

- `nuxt-i18n/app/components/HomeAboutUsSection.vue`

先把 `company/about.vue` 的主要渲染迁移到组件里，保证**About 页不变**。

### Step 4：首页接入（插入一个 Section）

在 `nuxt-i18n/app/pages/index.vue` 中插入一个 section（建议位置：TrustCards 后、FAQ 前 或 FAQ 后），结构参考已有：

- `max-w-5xl mx-auto px-4 py-*`

使用：

- `<HomeAboutUsSection variant="home" />`

### Step 5：i18n & 文案收口（建议与 Step 2/3 一起做）

- 把首页用到的文案（标题、简介、CTA）放到 `en.json`
- About 页长文案可分阶段迁移（先迁移首页会显示的部分即可）

---

## 验收清单

- 首页出现 About Us 区块，样式与首页现有 section 视觉一致
- CTA 能跳转到 `/company/about`
- About 页与首页的文案/图片**不重复维护**（单一数据源/组件复用已落地）
- 后续改 About 内容时：修改一处即可同步影响首页

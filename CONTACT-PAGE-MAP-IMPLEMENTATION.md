# Contact 页地图实施方案（Nuxt）

本文档用于指导在 Nuxt 前端中，为以下页面加入 **性能友好、SEO 友好、移动端体验好** 的地图模块：

- `nuxt-i18n/app/pages/company/contact.vue`

你选择的方案是：

- **先展示静态地图预览（图片）**
- **用户点击后，再在页面内懒加载交互地图（inline embed）**

并且有一个非常重要的实现原则：

- **抽成可复用的新组件 + 单一数据源**
  - `/company/contact` 与首页（Home）都复用同一个组件
  - 地址/坐标/链接等只维护一份，避免后续改地址要改两处

---

## 目标

- `/company/contact` 首屏保持 **快**（移动端 LCP/INP 友好）
- 用户能清楚地进行 **导航/路线规划（Get directions）**（移动端体验优先）
- 在用户点击之前 **不加载第三方地图资源**（脚本/iframe）
- SEO 重点放在 **可爬取的地址文本 + 结构化数据**（地图本身不是 SEO 的核心）

## 非目标

- 多门店地图/门店筛选/复杂地图交互

---

## 推荐交互（UX）

### 默认状态（不加载任何第三方地图）

- 页面展示：
  - 可复制的地址文本
  - 电话 / 邮箱
  - 营业时间（如有）
  - 一张 **静态地图预览图**（可点击）
  - CTA 按钮：
    - **打开 Google Maps**
    - **打开 Apple Maps**（建议保留，照顾 iOS）
    - **加载交互地图**（也可以直接点击预览图触发）

### 用户点击后（方案 B：页面内加载交互地图）

- 用交互地图 iframe 替换静态预览图（inline 展示）
- 地图下方保留 “Open in Maps” 按钮，便于用户直接打开系统地图 App 导航

---

## 关键架构：新建可复用组件 + 单一数据源

### 为什么要拆组件

- `contact.vue` 与首页都可能需要展示：
  - 地址
  - 地图预览
  - 跳转导航按钮
- 如果直接在页面里写两份，后续：
  - 改地址/改坐标/改 embed 链接 -> **容易漏改**

### 推荐文件结构（建议）

- 单一数据源（只维护一份地点信息）：
  - `nuxt-i18n/app/utils/contactLocation.ts`
- 可复用组件（地图 + 地址 + CTA + 懒加载逻辑）：
  - `nuxt-i18n/app/components/ContactLocationMap.vue`
- 页面使用：
  - `nuxt-i18n/app/pages/company/contact.vue` 引用组件
  - 首页（对应的 page 或组件）也引用同一个组件

### 数据源内容建议

`contactLocation.ts` 导出一个对象（或按需导出多个字段），包含：

- `name`：公司名
- `addressText`：完整地址（用于展示 + SEO）
- `lat` / `lng`：经纬度（用于导航链接）
- `googleEmbedUrl`：Google Maps iframe embed URL
- `openGoogleMapsUrl`：打开 Google Maps 的 URL
- `openAppleMapsUrl`：打开 Apple Maps 的 URL（可选）
- `previewImageSrc`：静态预览图路径（建议走 `public/`）

---

## 性能（Performance）建议

### 为什么这个方案更快

- 首屏不渲染 iframe，不触发地图请求
- 静态图片可缓存、可压缩（WebP），成本极低

### 静态预览图方案

- 方案 A（推荐）：提交一张本地图片到 `public/`
  - 示例：`nuxt-i18n/public/company/contact/map-preview.webp`
  - 优点：
    - 不需要 API key
    - 速度最好
    - 缓存最好
- 方案 B：使用第三方 Static Map API（Google / Mapbox）
  - 缺点：需要 API key、额度/成本、key 管理

### 交互地图（点击后加载）

- 推荐：Google Maps embed iframe
  - 在 Google Maps → Share → Embed a map 获取 iframe URL
  - 只在用户点击后渲染 iframe

---

## SEO 建议

### SEO 真正看重什么

- 页面上可爬取的文本信息：
  - 公司名
  - 完整地址
  - 电话（`tel:`）
  - 邮箱（`mailto:`）
  - 营业时间（如有）
- JSON-LD 结构化数据：`LocalBusiness`（或 `Organization`）

### 结构化数据放哪里更合适

- 推荐：只在 `contact.vue` 里通过 `useHead()` 输出 JSON-LD
  - 组件用于展示/交互
  - SEO 脚本放页面，避免首页也输出同一份 JSON-LD（可能造成重复语义）

澄清（避免误解）：

- 这里的意思不是“不要做组件”。
- **组件仍然建议抽出来复用**（首页/Contact 共用一份展示与交互逻辑）。
- 需要“单独放在页面层”的是 **JSON-LD 这段 SEO 数据**：
  - Contact 页面输出（更符合页面语义）
  - 首页一般不需要重复输出同一份 `LocalBusiness` JSON-LD，避免重复/混淆

需要确认字段：

- `name`
- `address`
- `telephone`
- `email`
- `geo`（lat/lng）
- `openingHours`（可选）

---

## 移动端与响应式（Mobile）建议

- 静态预览容器：
  - 宽度 100%
  - 圆角与现有卡片一致
  - 高度建议：
    - Mobile：240–320px
    - Desktop：360–420px
- 避免地图占太高导致联系信息被挤下去
- 保留 “Open in Maps” CTA，移动端直接跳转地图 App 体验最好

---

## 实施步骤（对照执行）

### 第 1 步：准备地点信息（填 TODO）

- 地址文本：
  - `Building 6, No. 639 Tongji South Road, Industrial Concentration Zone, Tong'an District, Xiamen, China (Xiandu Composite Materials Technology)`
- 经纬度（建议）：
  - （当前未填）
- Google iframe embed URL：
  - 当前使用（可用但可能不是精准 pin）：
    - `https://www.google.com/maps?q=TODO_URL_ENCODED_ADDRESS&output=embed`
  - 推荐替换为精准方式（见下方“当前实现进度 / 待完成”里的操作步骤）：
    - Google Maps → Share → Embed a map → 复制 iframe 的 src
- 打开 Google Maps URL（二选一）：
  - 按经纬度：
    - `https://www.google.com/maps/search/?api=1&query=TODO_LAT,TODO_LNG`
  - 按地址：
    - `https://www.google.com/maps/search/?api=1&query=TODO_URL_ENCODED_ADDRESS`
- Apple Maps URL（可选）：
  - `https://maps.apple.com/?q=TODO_URL_ENCODED_ADDRESS`

### 第 2 步：新增单一数据源

- 新建：`nuxt-i18n/app/utils/contactLocation.ts`
- 把第 1 步的地点信息统一放进去导出

建议直接按下面的字段结构写（可复制粘贴，再按需调整）：

```ts
export const contactLocation = {
  name: '纤镀复材科技',
  addressText: "Building 6, No. 639 Tongji South Road, Industrial Concentration Zone, Tong'an District, Xiamen, China (Xiandu Composite Materials Technology)",

  // 推荐：后续你可以用更精确的经纬度替换（用于导航链接更稳定）
  lat: '',
  lng: '',

  // Google Maps：点击后用于页面内显示交互地图（不需要 API Key）
  // 获取方式：Google Maps → Share → Embed a map → 复制 iframe 的 src
  // 当前实现：先用 q + output=embed 可跑通，但可能不是精准定位 pin
  googleEmbedUrl: 'https://www.google.com/maps?q=TODO_URL_ENCODED_ADDRESS&output=embed',

  // 打开 Google Maps（搜索/定位）
  openGoogleMapsUrl: 'https://www.google.com/maps/search/?api=1&query=%E5%8E%A6%E9%97%A8%E5%B8%82%E5%90%8C%E5%AE%89%E5%8C%BA%E5%B7%A5%E4%B8%9A%E9%9B%86%E4%B8%AD%E5%8C%BA%E5%90%8C%E9%9B%86%E5%8D%97%E8%B7%AF639%E5%8F%B76%E5%8F%B7%E6%A5%BC%E7%BA%A4%E9%95%80%E5%A4%8D%E6%9D%90%E7%A7%91%E6%8A%80',

  // Google Maps：直接进入导航（目的地为该地址）
  openGoogleDirectionsUrl: 'https://www.google.com/maps/dir/?api=1&destination=%E5%8E%A6%E9%97%A8%E5%B8%82%E5%90%8C%E5%AE%89%E5%8C%BA%E5%B7%A5%E4%B8%9A%E9%9B%86%E4%B8%AD%E5%8C%BA%E5%90%8C%E9%9B%86%E5%8D%97%E8%B7%AF639%E5%8F%B76%E5%8F%B7%E6%A5%BC%E7%BA%A4%E9%95%80%E5%A4%8D%E6%9D%90%E7%A7%91%E6%8A%80',

  // Apple Maps（iOS 体验更好，建议提供）
  openAppleMapsUrl: 'https://maps.apple.com/?q=%E5%8E%A6%E9%97%A8%E5%B8%82%E5%90%8C%E5%AE%89%E5%8C%BA%E5%B7%A5%E4%B8%9A%E9%9B%86%E4%B8%AD%E5%8C%BA%E5%90%8C%E9%9B%86%E5%8D%97%E8%B7%AF639%E5%8F%B76%E5%8F%B7%E6%A5%BC%E7%BA%A4%E9%95%80%E5%A4%8D%E6%9D%90%E7%A7%91%E6%8A%80',

  // 静态预览图（首屏用，建议放 public/，不需要 API Key）
  // 当前实现：未提供静态预览图时，组件会自动显示渐变占位
  previewImageSrc: '',
} as const
```

### 第 3 步：准备静态预览图

- 新增文件：
  - `nuxt-i18n/public/company/contact/map-preview.webp`
- 建议：宽度 ~1200px，压缩成 WebP
 - 当前实现：该文件暂未添加（可选项）。如添加后，把 `contactLocation.previewImageSrc` 改为 `/company/contact/map-preview.webp` 即可启用。

### 第 4 步：新建可复用组件

- 新建：`nuxt-i18n/app/components/ContactLocationMap.vue`
- 组件职责：
  - 读取 `contactLocation.ts`
  - 展示地址/CTA
  - 默认展示预览图
  - 点击后切换显示 iframe

组件边界（重要）：

- 组件内建议 **不要** 写 `useHead()` 去注入 JSON-LD
- 组件只负责 UI/交互，SEO 结构化数据交给页面层处理（下一步）

实现要点：

- `const showInteractiveMap = ref(false)`
- 点击后 `showInteractiveMap.value = true`
- iframe 建议包一层 `ClientOnly`（避免 SSR/水合差异风险）
- iframe 需要：
  - `title`
  - `loading="lazy"`
  - `allowfullscreen`

### 第 5 步：在 contact 页面引用组件

- 在 `nuxt-i18n/app/pages/company/contact.vue` 中引用 `ContactLocationMap`
- 页面层通过 `useHead()` 输出 JSON-LD（结构化数据）

### 第 6 步：在首页引用同一个组件

- 在首页需要展示地图/地址的位置，直接复用 `ContactLocationMap`
- 需要差异（例如首页更紧凑）时：
  - 给组件加 `variant`/`compact` prop
  - 但地点数据仍然只从 `contactLocation.ts` 来

---

## 当前实现进度（已落地）

### 已完成

- 单一数据源：已新增 `nuxt-i18n/app/utils/contactLocation.ts`
  - `addressText` 已改为英文（用于 UI 展示 + JSON-LD）
  - `openGoogleMapsUrl` / `openGoogleDirectionsUrl` / `openAppleMapsUrl` 已配置
  - `googleEmbedUrl` 已提供可用版本（q + output=embed）
- 可复用组件：已新增 `nuxt-i18n/app/components/ContactLocationMap.vue`
  - 首屏不加载 iframe
  - 点击后才加载交互地图（iframe）
  - 第一个按钮（Get directions）已改为白色背景
- Contact 页面：`/company/contact`
  - 已接入组件
  - 已在页面层通过 `useHead()` 注入 JSON-LD（LocalBusiness）
- 首页：首页已接入组件
  - 位置：FAQ 下方
  - 使用 `compact` 变体
- i18n：已补充 `en.json` 的 `contactLocation.*` 文案

### 未完成 / 待优化（推荐）

#### 1) Google Maps 精准定位（推荐必须做）

当前 `googleEmbedUrl` 使用的是 `q=地址&output=embed`：
- 优点：不需要 API key，能跑通。
- 风险：可能展示的是“搜索结果/不稳定位置”，不是你想要的“精准 pin”。

推荐替换成“Google Maps 官方分享出来的 embed 链接（iframe src）”，做法：
- 打开 Google Maps
- 搜索公司名或地址，确认地图上 pin 指向正确的位置
- 点击左侧地点卡片里的：Share（分享）
- 切换到：Embed a map
- 复制 iframe 代码中的 `src`（只要 src 的 URL）
- 粘贴到 `nuxt-i18n/app/utils/contactLocation.ts` 的 `googleEmbedUrl`
- 刷新页面，点击地图模块确认展示的位置就是精准 pin

#### 2) 经纬度（可选，但建议）

为了让“导航”更稳定（尤其是地址文本有歧义时），建议补齐经纬度：
- 在 Google Maps 打开精准 pin
- 右键该点 → “What\'s here?”（这是什么地方？）
- 复制坐标（lat,lng）
- 填入 `contactLocation.lat` / `contactLocation.lng`
- 后续可以把导航链接改为坐标形式（更稳定）：
  - `https://www.google.com/maps/dir/?api=1&destination=LAT,LNG`

#### 3) 静态预览图（可选）

当前没有提供 `map-preview.webp`，组件会显示渐变占位。
如要更“像地图”的首屏体验：
- 增加 `nuxt-i18n/public/company/contact/map-preview.webp`
- 设置 `contactLocation.previewImageSrc = '/company/contact/map-preview.webp'`

---

## QA 检查清单

- 首屏：
  - DOM 中没有 iframe
  - Network 中没有地图请求
- 点击后：
  - iframe 才出现并加载
  - 移动端可用，且不会把页面滚动体验搞坏
- 导航按钮：
  - Google Maps 打开正常
  - Apple Maps 打开正常（如配置）
- SEO：
  - 地址文本存在
  - JSON-LD 可通过校验（Google Rich Results Test）

---

## 常见坑

- 把 API key 写进前端 bundle（避免）
- 首屏直接加载 iframe（性能明显变差）
- 地图高度过高，挤占移动端首屏（体验差）
- 只靠地图、不提供可爬取的地址文本与 JSON-LD（SEO 收益很小）

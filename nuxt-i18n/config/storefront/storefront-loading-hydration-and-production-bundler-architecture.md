# Nuxt 用户端加载、延迟激活与生产构建架构

复审日期：2026-10-26。本文维护当前代码的加载边界、框架相关实现和升级验收要求；历史体积、浏览器记录及测量条件见[构建分析报告](../../../frontend-bundle-tree-shaking-analysis-and-safe-optimization-guide.md)。本次复审核对源码、锁文件和实际依赖解析，没有重新测量线上性能。

本文放在 `config/storefront/`，使加载契约与代码一起保存。仓库的 `docs/` 目录默认被 Git 忽略，本地架构索引和 `output/` 报告不能承担唯一的长期维护入口。

## 1. Nuxt、Vite 与打包引擎的层次

Nuxt 是应用框架，管理 SSR、路由、自动导入、payload、head 和资源清单；Vite 是 Nuxt builder 使用的开发与构建工具。Rollup / Rolldown 是具体打包引擎，Nitro 还负责服务端产物与运行时。后台使用 Vue + Vite 的客户端应用，不经过 Nuxt 的 SSR 和 hydration 链路。

2026-10-26 在当前工作区核对 `package-lock.json` 和实际安装依赖得到：

| 项目 | 当前链路 | 配置位置 |
| --- | --- | --- |
| 用户端浏览器产物 | Nuxt 4.4.8 → `@nuxt/vite-builder` → Vite 7.3.6 → Rollup 4.62.2；Vue 3.5.43 | [`nuxt.config.ts`](../../nuxt.config.ts) 的 `vite.build.rollupOptions` |
| 用户端服务端产物 | Nuxt SSR + Nitro 2.13.4 | `nuxt.config.ts` 的 `nitro`；不能用客户端 treemap 代表服务端体积 |
| 后台浏览器产物 | Vue → Vite 8.1.5 → Rolldown 1.1.5 | [`vite.config.ts`](../../../go-backend/web/admin/vite.config.ts) 的 `build.rollupOptions`，当前 Vite 提供兼容映射 |

这些是核对时的版本，不是永久固定的架构条件。`package.json` 的版本范围、编辑器解析到的类型和顶层 `node_modules/vite` 都不能单独证明实际构建引擎：用户端顶层 Vite 是 **8.1.5**，而 builder 实际解析到其自身目录下的 **7.3.6**。插件的类型检查通过，也不能替代实际 builder 的运行验证。

在 `nuxt-i18n/` 中可检查 builder 的解析位置和版本：

```powershell
node -e 'const { createRequire } = require("node:module"); const nuxtBuilderRequire = createRequire(require.resolve("@nuxt/vite-builder")); console.log(nuxtBuilderRequire.resolve("vite/package.json")); console.log(nuxtBuilderRequire("vite/package.json").version);'
```

后台在 `go-backend/web/admin/` 核对：

```powershell
node -p 'require("vite/package.json").version'
node -p 'require("rolldown/package.json").version'
```

Nuxt 的顶层 `build` 还包含 `analyze` 等框架选项；不能把普通 Vite 示例的 `build.rollupOptions` 直接贴到 Nuxt 顶层。开发服务器的 ESM、依赖预构建和 HMR 也不能证明生产树摇、chunk 拆分和 SSR 预加载正确。

## 2. 优化必须区分的六个边界

| 边界 | 解决的问题 | 当前实现与限制 |
| --- | --- | --- |
| Tree shaking | 删除未使用且可证明无副作用的代码 | 已使用的国家、省州及富文本安全解析数据仍需保留 |
| Chunk 拆分 | 把代码组织到不同文件 | `manualChunks` 只分组；静态依赖或共享 vendor 仍可把代码带回入口 |
| 按需下载 | 用户需要时再发起模块请求 | 动态 `import()` 必须位于真正延后的执行路径；请求时机由生产浏览器验证 |
| 延迟 hydration | 推迟 Vue 接管已有 SSR DOM 和事件 | SSR 正文仍存在；仅添加 hydration strategy 不保证 loader 延后执行 |
| 资源提示 | 控制浏览器提前获取资源 | SSR `prefetch`、`modulepreload` 与客户端 NuxtLink 预取是不同路径 |
| SSR 与 payload | 保留抓取内容并复用服务端数据 | 推迟 UI 下载不会自动减少 HTML、FAQ 数据或 payload 大小 |

### 默认树摇及副作用参数

两端都保留默认树摇与依赖包的副作用声明。需要改变某个包时，先定位引用链，再验证行为与最终产物。

| 参数 | 不能照搬的原因 |
| --- | --- |
| `moduleSideEffects: 'no-external'` | 假设 external 模块没有导入副作用，并不表示严格分析所有 npm 包；已打包依赖也不会因此自动全部变小 |
| `propertyReadSideEffects: false` | 可以删掉结果未使用但会触发 getter 或异常的属性读取，可能改变运行行为 |
| `tryCatchDeoptimization: false` | Rollup 支持，但会影响依赖异常进行能力探测的代码；当前 Rolldown 1.1.5 的公开树摇选项不包含它 |

不要为整个应用声明 `moduleSideEffects: false`，或未经验证移除 CSS、polyfill、全局注册。后台保留 `rollupOptions` 兼容入口不代表两个引擎的配置语义和插件行为完全相同。

`country-region-data` 在用户端已按支持国家命名导入，位于账户/结账加载链路。支持范围及省州代码关联地址、运费、税率和 Stripe，不按地名冷热随意裁剪。后台税率页需要全球目录，目录延迟到相关页签再加载；这是推迟下载，不能算作全部产物体积消失。

## 3. 当前用户端加载责任

| 责任 | 实现入口 |
| --- | --- |
| SSR 资源清单、交互预取、vendor 拆分及分析器 | [`nuxt.config.ts`](../../nuxt.config.ts) |
| 全局入口静态依赖约束 | [`public-client-entry-javascript-dependency-boundary-plugin.ts`](public-client-entry-javascript-dependency-boundary-plugin.ts) |
| 加载分层与可见范围参数 | [`storefrontLoadingPolicy.ts`](../../app/utils/storefrontLoadingPolicy.ts) |
| 首页 SSR 保留、下载延迟及 CSR 占位 | [`HomeDeferredSection.vue`](../../app/components/home/HomeDeferredSection.vue) 与 [`index.vue`](../../app/pages/index.vue) |
| FAQ 数据、立即管理的 head 与延迟 UI | [`PageFaqSlot.vue`](../../app/components/PageFaqSlot.vue) 与 [`PageFaq.vue`](../../app/components/PageFaq.vue) |
| 未激活页脚的语言切换 | [`StorefrontFooterWithLocaleSafeDeferredHydration.vue`](../../app/components/StorefrontFooterWithLocaleSafeDeferredHydration.vue)，供四种公共布局使用 |
| 账户首次打开、页签首次加载与实例保留 | [`SidePanel.vue`](../../app/components/SidePanel.vue) 与 [`AccountSidebarPanel.vue`](../../app/components/account/AccountSidebarPanel.vue) |
| 页头当前页面翻译、分组菜单翻译及路由计算复用 | [`SiteHeader.vue`](../../app/components/SiteHeader.vue) |
| HTML 与资源缓存 | [`html-cache-policies.ts`](html-cache-policies.ts)、[`route-rules.ts`](route-rules.ts) 和[缓存维护说明](README.md) |

### 异步 loader 与 hydration 的先后顺序

Vue 3.5.43 的 `__asyncHydrate` 先解析异步组件 loader，再运行 hydration strategy。`defineAsyncComponent({ loader: () => import(...), hydrate: ... })` 会先请求业务 JS，即使 hydration 被推迟。

`HomeDeferredSection` 使用两层异步组件：外层 loader 只返回本地组件定义，内层才导入业务模块。接近视口、鼠标进入或键盘聚焦后激活外层，才有机会执行内层 loader。这与当前 Nuxt 原生 lazy runtime 的结构一致；仍需同时排除静态导入和 SSR 提示造成的提前请求。

SSR 首次访问与 CSR 导航必须分别验收：

- 首次 SSR 与初始接管保留完整首页正文，不能用 `ClientOnly` 或客户端 `v-if` 删除可抓取内容。
- 从其他页面导航回首页没有现成 SSR DOM；`HomeDeferredSection` 使用近视口再挂载的占位路径。最小高度保留到真实内容 mounted，锚点在占位与内容之间交接，不能重复 ID 或让主按钮找不到目标。
- Nuxt 的 `hydrate-on-visible` 面向已有 SSR DOM。FAQ 与页脚在普通 CSR 挂载时可能直接下载，不能把它描述为所有导航路径的可见范围下载门槛。
- 延迟组件的输入在接管前发生变化也可能促使 Vue 提前接管。必须检查快速切页、切语言和异步数据更新，而非只检查静止的首屏。

### SSR 资源提示：依赖内部调用顺序的兼容点

`build:manifest` 关闭 SSR 自动 `prefetch`，并关闭指定首页区块的 `preload`。NuxtLink 使用交互预取，保留实际 HTML 链接和当前页面必需的 `modulepreload`。首次点击未预取页面可能需要现场下载，应在减少无操作下载与导航等待之间验收。

当前 Nuxt renderer 的顺序为：

```text
完成 SSR HTML
  → 收集 SFC 内联样式
  → app:rendered
  → getRequestDependencies 收集并缓存样式和脚本依赖
  → 按 ~lazyHydratedModules 删除模块登记
  → 生成 preload / prefetch
```

仅依赖最后的原生 lazy 排除可能已晚于依赖缓存。首页包装与 FAQ slot 在 `app:rendered` 中提前移除特定 `ssrContext.modules` 登记；首页同时登记到 `~lazyHydratedModules`。FAQ 的排除范围是 `PageFaq.vue`、`FaqAnswerContent.vue`、`faq/DesktopFaqMasterDetail.vue`、`SafeRichText.vue` 的组件模块 ID。

这是**当前版本的兼容实现**：`~lazyHydratedModules`、模块 ID 和 renderer 顺序都依赖框架内部机制。此时内联样式已收集，不删除 SSR HTML。共享富文本正文仍须通过自身页面静态依赖获得资源；同一模块的其他使用者必须作为对照验证，不能把 FAQ 的排除逻辑扩大成全站删除共享依赖。

### SEO、富文本安全与状态保留

- FAQ 所有折叠答案继续由 SSR 输出，折叠使用 `v-show`；接口失败继续提供占位与重试。数据请求与 UI 延迟分开，不能把网络错误解释为无内容。
- 自动插入 FAQ 的 JSON-LD 由立即运行的 `PageFaqSlot` 管理，使用稳定 key `storefront-page-faq-structured-data`。内部 `PageFaq` 关闭重复输出；独立使用 `PageFaq` 默认仍输出结构化数据。切页、无 FAQ 页面和切语言后都只保留当前内容的 schema。schema 存在不代表搜索引擎承诺富结果展示。
- FAQ 内容身份还可由 `route.meta.faqPageId` 指定，不能假设每条新 URL 都必须创建新 FAQ 数据。canonical、语言和内容身份分别按现有规则解析。
- `FaqAnswerContent → SafeRichText → utils/security/safeRichText.ts → parse5` 承担 HTML 清洗、链接和媒体来源限制以及 SSR/客户端一致性。优化下载时机必须保留这条安全路径，不能用原始 `v-html` 替换解析器。需要富文本正文的页面仍可能首屏下载 parse5。
- 页脚尚未激活时切语言，旧 SSR 文本可能与新 locale 不一致。公共页脚包装只替换尚未激活的实例；已激活实例保留，让翻译正常更新且订阅邮箱草稿不丢失。
- 账户区首次打开后保留实例，页签以“首次访问才挂载 + 后续 `v-show`”保留地址等草稿。改成只按当前页签 `v-if` 会反复销毁表单。
- 当前页面导航词条需在 SSR 就绪；其他分组词条在打开菜单时请求。不能为了减少翻译下载，让首屏导航退化为原始翻译 key。

## 4. 最终产物与缓存的限制

依赖边界插件在 `writeBundle` 遍历入口 chunk 及其静态 `imports`，使 Nuxt 在 `generateBundle` 中追加的依赖也进入检查。账户组件、CheckoutModal、country-region-data、Zod、focus-trap、qrcode、Stripe SDK 进入此静态图会使构建失败。

这个检查不遍历动态 imports，不覆盖额外的路由/语言块、SSR head、CSS、图片、字体，也没有宣称固定体积预算。构建通过仍需用 SSR 检查与生产浏览器确认未操作时的实际请求。升级后应再次独立解析 `.output/public/` 的真实文件，确认插件记录与最终 import 图一致。

所有语言首页（`/`、`/zh_cn`、`/fr` 等）当前都是 `no-store`；hashed `/_nuxt/**` 资源仍长期缓存。曾经遇到浏览器复用旧构建 HTML、请求已删除哈希文件的情况，不能把旧页面请求失败判成当前构建的拆包错误。

其他公开内容页仍可使用 Nitro HTML cache。首页 `no-store` 不能替代全站部署时的 HTML/资源版本一致性：发布时需要清理旧 HTML cache 或保留其引用的旧资源，CDN 与浏览器缓存也要按各自层次处理。开发环境关闭 HTML cache，不足以验证生产缓存行为。

## 5. 修改与框架升级后的验收

升级 Nuxt、Vue、Vite builder、Nitro 或打包引擎，或移动延迟组件时，先核对实际解析版本，再检查以下当前源码位置；框架升级后路径也可能变化：

| 依赖源码（相对于 `nuxt-i18n/node_modules/`） | 要核对的行为 |
| --- | --- |
| `@vue/runtime-core/dist/runtime-core.esm-bundler.js` | `__asyncHydrate` 的 loader 与 strategy 顺序、更新时的提前接管 |
| `nuxt/dist/components/runtime/lazy-hydrated-component.js` | 两层包装、hydrated 事件和服务端模块登记 |
| `@nuxt/nitro-server/dist/runtime/handlers/renderer.mjs` | 内联样式、`app:rendered`、依赖缓存和 lazy 排除的先后顺序 |
| `@nuxt/vite-builder/dist/index.mjs` | 产物依赖后处理、manifest 与分析器集成；不能只按插件类型推断 |

在用户端目录执行对应检查；修改运行实现或升级依赖时再运行生产构建：

```powershell
npm run scripts:typecheck
npm run test:checkout-state-province-catalog
npm run test:footer-independent-translation-loading
npm run build:analyze
npm run check:html-cache
```

`npm run build:analyze` 使用 Nuxt 生产构建，输出 `output/bundle-analysis/storefront-client-treemap.html`；每次普通构建也输出 `storefront-global-entry-dependency-summary.json`。后台同名命令输出其自己的 treemap，保持生产 mode 和字体检查，不能改用 `--mode analyze` 而遗漏 `.env.production`。

运行独立生产预览后，再让只读检查指向它与相应后端；以下地址仅为示例，应使用当前验证环境的真实端口：

```powershell
# 一个终端运行生产预览
npm run preview -- --host 127.0.0.1 --port 10340
```

```powershell
# 另一个终端运行检查
$env:HOME_SSR_URL = 'http://127.0.0.1:10340/'
npm run check:home-ssr
$env:COMPANY_ABOUT_SSR_ORIGIN = 'http://127.0.0.1:10340'
$env:COMPANY_ABOUT_FAQ_BACKEND_ORIGIN = 'http://127.0.0.1:10241'
npm run test:backend-configured-page-faq-ssr
```

生产浏览器验收至少覆盖：

1. 禁用缓存的新会话首次打开桌面和窄屏首页、胎压页、胎宽兼容页及 Schwalbe 页。先不滚动、不悬停链接，记录初始请求，再滚动激活。进入预加载边距内的区块允许立即下载，不能要求所有 FAQ 都等到滚动。
2. 检查完整 SSR 问答、标题、canonical、hreflang 与 JSON-LD；从有 FAQ 页面切到其他页面及切语言后，不遗留旧 schema。以需要富文本的 `/website/me-this-website` 和 `/website/why-this-name` 为对照，确认共享安全依赖仍能及时加载。
3. 检查 CSR 返回首页、主按钮锚点、占位高度，以及激活后的 FAQ 展开与键盘交互。
4. 四种布局分别检查未激活页脚前切语言、已激活页脚填邮箱草稿后切语言；账户页签与侧栏开关后地址草稿保留。记录 hydration 警告、控制台错误与网络失败。
5. 请求各语言首页，核对 `no-store` 与无 Nitro HTML cache 配置；构建哈希变化后验证 HTML 与资源版本一致。现有 `check:html-cache` 不能代替全部语言首页及浏览器缓存层的实测。

## 6. 如何解释性能证据

treemap 用于定位模块与引用关系。模块级 gzip 不能直接相加；全局入口静态图的 gzip 是涉及文件分别压缩后的和，既不是整个页面下载量，也不是服务器真实传输字节数。

按需下载会改变首次加载成本，用户打开功能后仍可能下载相同总代码。窄 viewport 不等于手机 CPU/网络仿真；本地采样和体积下降也不能直接换算为线上 LCP、INP、CLS、TTFB 或搜索排名提升。

本次复审没有新增浏览器或线上指标结论。2026-10-24 的历史测量、对照路由与限制保留在[分析报告](../../../frontend-bundle-tree-shaking-analysis-and-safe-optimization-guide.md)，避免把历史数字当作持续有效的性能承诺。

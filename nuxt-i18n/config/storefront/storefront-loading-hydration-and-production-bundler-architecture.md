# Nuxt 用户端加载、延迟激活与生产构建架构

2026-11-19 字体资源更新：保留内置字体及 `font-display: block`，将原 6.4 MB CJK／扩展合并文件细分为按字符命中的 WOFF2；完整覆盖和字形一致性由构建检查约束。字体的生成、Stripe 同步、发布门禁及验收口径见 [CJK 字符分片契约](./storefront-cjk-font-character-sharding-generation-validation-and-loading-contract.md)。

复审日期：2026-10-26。本文维护当前代码的加载边界、框架相关实现和升级验收要求；历史体积、浏览器记录及测量条件见[构建分析报告](../../../frontend-bundle-tree-shaking-analysis-and-safe-optimization-guide.md)。本次复审核对源码、锁文件和实际依赖解析，没有重新测量线上性能。

2026-11-17 更新主导航加载契约：七个顶部菜单统一为立即可显示的外壳和导航链接，菜单文案与商店专用组件在首屏稳定后预热。以下主导航规则替代此前“所有菜单资源等点击后加载”的约定；2026-10 的体积与性能数字仍是历史数据。

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
| 页头翻译、导航意图预热及路由计算复用 | [`SiteHeader.vue`](../../app/components/SiteHeader.vue) |
| 七个菜单的立即响应外壳、链接与商店加载反馈 | [`HeaderMegaMenu.vue`](../../app/components/HeaderMegaMenu.vue) |
| 首屏稳定后的菜单资源预热与 10 秒启动兜底 | [`scheduleStorefrontPrimaryNavigationResourceWarmup.ts`](../../app/utils/scheduleStorefrontPrimaryNavigationResourceWarmup.ts) |
| 单项翻译慢请求的等待上限与过期语言任务停止 | [`warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait.ts`](../../app/utils/warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait.ts) |
| 商店菜单组件下载复用与失败重试 | [`headerPrimaryNavigationShopComponentPreloading.ts`](../../app/utils/headerPrimaryNavigationShopComponentPreloading.ts) |
| 底部箭头入口、首次操作下载与移动端展开／收起 | [`GradientDockMenuShell.vue`](../../app/components/GradientDockMenuShell.vue)、[`GradientDockMenuDeferred.vue`](../../app/components/GradientDockMenuDeferred.vue) 与 [`GradientDockMenu.vue`](../../app/components/GradientDockMenu.vue) |
| HTML 与资源缓存 | [`html-cache-policies.ts`](html-cache-policies.ts)、[`route-rules.ts`](route-rules.ts) 和[缓存维护说明](README.md) |

### 顶部主导航：立即响应、稳定后预热、页面按需下载

商城、技术、资源、支持与服务、关于我们、工具、指南共用 `HeaderMegaMenu`，桌面和手机执行相同规则。它们独立于 `StorefrontClientOverlaysDeferred` 中的购物车、结账和搜索等全局弹层。改变全局容器的延迟不会预热顶部菜单，也不代表其内部弹层已下载。

此前菜单仅由 `LazyHeaderMegaMenu v-if` 在点击后首次下载，既无空闲预热，也无下载期间的可见反馈。2026-11-17 本机访问线上中文首页，停留约 255 秒后点击技术，才发起菜单 JS/CSS 请求，本次资源响应等待约 8.9 秒；加载完成后再次打开约 53 毫秒且无新增资源请求。这是单次、本机网络条件下的排查证据，不能当作所有用户的耗时或 PageSpeed 分数。

当前契约：

| 阶段 | 行为与边界 |
| --- | --- |
| 页头首次加载 | `HeaderMegaMenu` 外壳、菜单样式及普通导航链接逻辑静态随页头就绪；点击后立即渲染，无需等待另一个菜单 chunk。页头内联关键 CSS 包含菜单基础布局，完整 entry CSS 延迟时仍可使用。关闭态不输出全部弹层 DOM；首屏顶部导航文案继续 SSR 输出。内页导航所需翻译由 `onServerPrefetch` 保证 SSR；只序列化导航词条供客户端接管重放，客户端完整 namespace 下载不阻塞页头 setup。 |
| 页头初始化时刻 | 页头 DOM 的函数 ref 确认其已接管后初始化关闭监听、尺寸测量和预热调度；不等待正文组件的异步 setup，也不依赖可能被 Nuxt 根 Suspense 延后的 mounted 回调。组件卸载前后都检查取消状态。 |
| 首屏稳定后 | 从页头 DOM 接管后的初始化时刻起至少等待 2,500 ms，且最近一次观察到的 LCP 候选后安静 800 ms，再在 `requestIdleCallback` 中启动预热。不支持相关 API 时按相同最短延迟使用定时器。此规则近似判断渲染稳定，不承诺最终 LCP 已结束。 |
| 启动兜底 | 从同一初始化时刻起 10,000 ms 独立启动预热，不等 `window.load`，也不再叠加 idle 超时。后台标签等待重新可见；浏览器节流、主线程阻塞仍可能推迟执行。10 秒约束启动调度，不保证资源已下载完成。 |
| 明确导航意图 | 桌面、手机导航按钮的 pointerenter、键盘 focus 和点击立即请求对应菜单文案；商城同时请求商店专用组件，直接越过后台等待。失败不阻断外壳及已有回退链接。 |
| 后台预热内容 | 仅下载商店组件定义及当前语言导航使用的翻译 namespace；翻译逐项请求，单项最多等待 1,500 ms 后继续下一项，避免一个挂起请求挡住所有后续菜单。原动态 import 无法中止，超时后仍可完成合并，不能把等待上限理解为网络取消或失败缓存。英文回退和当前语言由现有翻译加载器合并。namespace 资源仍包含其原有文本，尚未细拆为逐条菜单词条。切语言使用递增任务代次，A→B→A 也不会让旧任务恢复启动后续下载；卸载同样停止启动新项。 |
| 商城组件和数据 | 商城组件与普通菜单分离，不因点击技术等菜单而加载。预热只下载组件定义，不挂载、不提前请求组件数据或图片。商城点击后立即出现可关闭外壳及商城入口；组件未就绪时显示加载状态，下载失败可重试。主力产品异步 setup 有独立 Suspense 占位，不阻塞整个弹层。弹层主力产品使用独立锚点与紧凑布局，不能重复首页 `home-main-product-categories` ID 或覆盖首页自身间距。 |
| 具体跳转页面 | 不在定时菜单预热中加载目标页面组件、路由 payload、页面图片或全语言资源。页面仍使用 NuxtLink 的悬停／聚焦预取和点击导航。 |

参数唯一入口是 `STOREFRONT_PRIMARY_NAVIGATION_RESOURCE_WARMUP`。调度取消与页头生命周期绑定；商店组件下载 Promise 供预热和点击复用，失败清除缓存以允许重试。原有 7.5 秒商品分类数据预热、15 秒全局容器／其他后台任务属于不同责任，不作为菜单准备完成的信号。

PageSpeed / Lighthouse 与搜索引擎抓取没有统一的“10 秒结束”时刻。后台下载或脚本执行仍可能竞争带宽、CPU，并影响仍在发生的最大内容绘制；加载耗时和 INP 也须独立验收。禁止根据爬虫、测速 User-Agent 切换菜单行为。验收目标是页面已接管后点击立即出现有内容或加载状态的弹层，不能以单一首屏分数接受数秒无反馈。

2026-11-17 第二轮复查确认并修复了此前首页常规点击测试遗漏的边界：内页客户端 setup 仍等待完整导航 namespace；正文 async setup 会经 Nuxt 根 Suspense 延后 mounted 回调，连带延后关闭监听和预热计时；一个挂起 namespace 会阻塞全部后续预热；商城主力产品与首页重复 ID，且紧凑样式把同一个根节点误当成后代，导致桌面上下仍各有 56 px 留白。相应修复见上述初始化、导航词条快照、单项等待上限与独立锚点契约。

最新本地生产预览共通过 12 个浏览器场景：桌面与 390 × 844 手机 UA／触摸仿真中，主动阻断商店组件与技术菜单翻译请求，七个菜单点击到可见约 40～117 ms；直接进入轮圈测试内页并持续挂起页面翻译时，桌面／手机约 54／50 ms，Esc 仍能关闭，中文面包屑在完整词条到达前后相同。商城下载失败重试、接口慢请求下的加载反馈／失败重试、关闭后不自动重开、快速连续切菜单后一次浏览器返回关闭、独立锚点／零紧凑内边距、完整 entry CSS 阻断下的外壳均通过。挂起第一个翻译请求时，后续 namespace 仍在约导航后 4.6 秒发起；中文→英文→中文后链接使用当前语言路由。

独立内容页无操作时约导航后 2.81 秒开始空闲预热，未请求主力商品展示数据；禁用所有 idle 回调并持续挂起 `window.load` 时，约导航后 10.29 秒开始预热。另一个场景持续挂起内页翻译以保持根 Suspense 未完成，并禁用 idle 回调，仍通过 10 秒预热启动与 Esc 关闭检查。时间均含页头接管前耗时，10 秒约束从页头自身初始化时刻计算。各场景没有未捕获浏览器异常或 hydration 不一致；本地媒体缺失通过只读图片夹具隔离。以上不是手机 CPU／网速仿真，也不是线上 PageSpeed 复测。

同次验证通过调度与挂起请求／语言任务代次回归、面包屑与页脚翻译回归、脚本与 Vue 类型检查、生产构建的账户／支付静态依赖边界、关键 CSS、HTML cache 和首页 SSR 检查；首页仍有 9 个非空延迟区块，17 项延迟资源不进入 modulepreload／prefetch。浏览器脚本与记录保存在被 Git 忽略的 `nuxt-i18n/output/playwright/`：`verify-production-primary-navigation-immediate-shell-loading-failure-retry-and-warmup-deadline.js`、`verify-production-header-inner-page-hydration-stalled-translations-shop-anchors-and-slow-api-recovery.js`、`verify-production-header-warmup-deadline-while-page-translation-keeps-root-suspense-pending.js`，对应最新记录为 `primary-navigation-production-browser-verification-after-hidden-issues-fix.log`、`primary-navigation-hidden-issues-production-browser-verification.log`、`primary-navigation-root-suspense-independent-warmup-deadline-verification.log`。确定性回归通过 `npm run test:primary-navigation-resource-warmup` 执行。

### 移动端底部菜单：上沿居中箭头、完整图标行、首次操作才下载

2026-11-17 调整底部入口与展开／收起布局，加载责任仍由 `GradientDockMenuDeferred` 管理，独立于上述顶部菜单的后台预热。不能因为调整开关或尺寸而把业务组件改成静态导入、定时下载或提前挂载。

| 状态与操作 | 当前契约 |
| --- | --- |
| 首屏收起态 | 手机底部中央显示 96 × 32 px 的向上双箭头按钮，使用入口组件自身的内联 SVG 和 CSS；底部为 8 px 加安全区。业务菜单和快购组件仍不因入口可见而下载。 |
| 第一次点击 | 复用原有动态 import Promise，下载并挂载业务菜单。箭头移动至操作栏上沿中央并朝下，提供收起入口；展开态呈现与底栏连续的一体凸起，使用相同底色，底边无分隔线、无独立胶囊阴影，悬停也不变成另一块绿色背景。2026-11-18 将凸起两侧与底栏的连接处改为 16 px 反向圆角，用内联 SVG 绘制连续轮廓；装饰左右各延伸 16 px，不挤占按钮、图标或金额区域，不产生额外图片请求。不添加遮罩。全局 `quickbuy:open-entry` 仍通过原有激活入口转交快购请求。 |
| 移动端展开布局 | 小于 768 px 时底栏为视口宽度减 4 px，左右各留 2 px，解除桌面 500 px 上限并贴底。图标行高 72 px 加底部安全区；箭头独占行上方的 32 px，不占用图标列或金额宽度。用户中心、聊天和快购保持纯图标，只保留购物车金额及数量等业务文本。 |
| 金额区比例 | 恢复原金额白色胶囊，图标入口与金额列使用 `1fr : 1.85fr`，金额列最小 146.4 px；保留原货币图标 20 px、金额字号 16.8 px、数量徽标 30 px，不能为容纳开关擅自缩小。360 px 及以下恢复原 `1fr : 1.9fr`、金额列最小 138.4 px，使用 18.4 px 货币图标、16 px 金额和 28 px 徽标；长金额仍按原规则省略。 |
| 可访问名称 | 账户图标使用“用户中心”，全部 20 种语言同步对应词条。图标无需增加可见说明文字；箭头同步 `aria-expanded`，展开时关联操作栏。 |
| 继续操作与收起 | 点击正文和其他入口后保留横栏，用户可以继续操作页面；点击向下箭头或按 Esc 收起。上层可见弹层优先处理自己的 Esc。收起只隐藏横栏、保留已加载实例，再次展开复用已下载模块。 |
| 桌面与快购边界 | 桌面保留原居中箭头入口及展开后的底栏。快购组件、弹层与数据的加载时机保持原链路，手机按钮也保持原样式；用户中心、购物车和聊天继续执行原操作。 |

这些约束保护现有首屏下载边界，不表示实测 LCP 或线上 PageSpeed 分数完全不变；不能用“爬虫已结束”作为提前下载底部业务模块的依据。

2026-11-17 的本地生产构建通过 5 个浏览器场景，覆盖 390、320、767 px 手机视口、收起后切到桌面、快购弹层与用户中心操作、英文桌面布局。390 px 下金额区约 146 px，另外三个入口各约 73 px；320 px 下金额区约 138 px，其他入口各约 54 px，均无横向溢出。展开箭头与底栏底色相同、无独立阴影和底部分隔边；正文点击后横栏保留，快购先处理自己的 Esc，反复收起／展开不重复下载业务模块。首页无操作停留 16.5 秒未请求或挂载业务底栏，首次点击后仅请求一次对应模块；各场景无未捕获异常或 hydration 警告。

2026-11-18 的反向圆角调整另外通过开发浏览器的 320、390、767 px 显示、悬停、收起／展开与切至桌面检查：图标行仍高 72 px，金额区宽度及字号不变，左右装饰共延伸 32 px，未产生横向溢出，桌面不显示该装饰。本次仅修改入口的内联 SVG 与样式，没有重新运行上述完整生产验证或线上性能测量。

同次验证通过 Vue 类型、生产构建、关键 CSS、HTML cache 与首页 SSR 检查；entry 与页头关键 CSS 仍合计 6.55 KiB gzip，首页仍为 9 个非空延迟区块、17 项延迟资源不进入 modulepreload／prefetch。浏览器使用手机 UA／触摸仿真，本地缺失媒体通过图片夹具隔离，未重测线上 PageSpeed。脚本和最终记录位于被忽略的 `output/playwright/verify-production-mobile-bottom-dock-centered-arrow-full-width-icon-row-cart-sizing-and-unchanged-intent-loading.js` 与 `output/playwright/mobile-bottom-dock-integrated-center-arrow-production-browser-verification.log`。

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
- 当前页面导航词条需在 SSR 就绪，并通过少量导航词条快照保持接管一致；客户端不能再次 await 整个页面翻译 namespace 才让页头接管。其他分组词条在导航意图或首屏稳定后的后台预热时请求。词条缺失或失败时使用已有文字回退，不能阻塞弹层或让首屏导航退化为原始翻译 key。

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
npm run test:primary-navigation-resource-warmup
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
6. 冷缓存、桌面和手机各自首次点击七个菜单，确保无需菜单资源请求就能出现外壳和普通链接；将商店组件请求延迟或阻断，验证加载占位、失败重试、Escape、关闭按钮、点击外部及快速换菜单。不允许下载完成后自动重开已关闭的菜单。阻断完整 entry CSS 时，内联关键 CSS 仍须提供可见、可关闭的菜单。
7. 无操作时核对最早启动、LCP 候选变化、浏览器一直不空闲、`window.load` 未完成、后台标签恢复及组件卸载；普通可见页 10 秒兜底必须独立于 load 和 idle。预热不应请求目的页面、支付与账户模块；当前语言和英文回退之外的语言不应预取。切语言、CSR 导航后重复检查。

## 6. 如何解释性能证据

treemap 用于定位模块与引用关系。模块级 gzip 不能直接相加；全局入口静态图的 gzip 是涉及文件分别压缩后的和，既不是整个页面下载量，也不是服务器真实传输字节数。

按需下载会改变首次加载成本，用户打开功能后仍可能下载相同总代码。窄 viewport 不等于手机 CPU/网络仿真；本地采样和体积下降也不能直接换算为线上 LCP、INP、CLS、TTFB 或搜索排名提升。

本次复审没有新增浏览器或线上指标结论。2026-10-24 的历史测量、对照路由与限制保留在[分析报告](../../../frontend-bundle-tree-shaking-analysis-and-safe-optimization-guide.md)，避免把历史数字当作持续有效的性能承诺。

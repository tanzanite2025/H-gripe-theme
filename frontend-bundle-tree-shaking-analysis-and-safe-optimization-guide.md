# 前后台构建体积分析与树摇优化

本文保留 2026-10-24 各轮生产构建与浏览器验证的历史证据，数字不会自动随代码更新。当前长期约束、Nuxt 与后台 Vite 的分层及版本差异、已遇到的坑与升级验收见 [Nuxt 用户端加载、延迟激活与生产构建架构](nuxt-i18n/config/storefront/storefront-loading-hydration-and-production-bundler-architecture.md)。2026-10-26 的复审核对源码、锁文件与实际解析依赖，没有重新测量线上指标。

## 第一轮：Nuxt 用户端首屏与抓取

本轮优先处理所有访客都会下载的 Nuxt 全局入口，以及 SSR HTML 发出的资源提示。正文、导航链接和搜索元数据继续由服务端输出；账户、地址和支付功能在用户操作时加载。

2026-10-24，同一工作区生产构建的最终 JavaScript 文件对比：

| 全局入口及其静态依赖 | 优化前 | 优化后 | 减少 |
| --- | ---: | ---: | ---: |
| 原始大小 | 1,030,681 bytes | 825,170 bytes | 205,511 bytes（19.94%） |
| 各 chunk 分别 gzip 后求和 | 257,861 bytes | 187,637 bytes | 70,224 bytes（27.23%） |

该口径包含 Nuxt 注入后的最终依赖映射，不包含额外的页面、语言、CSS、字体或图片资源，也不代表首页全部下载量。统计遍历全局入口的静态 `import`；动态导入不计入全局必需依赖。通过独立解析 `.output/public/` 的实际文件复核，数字与构建报告一致。

具体改动：

- 账户侧栏在首次打开时才挂载；六个账户页签各自在首次访问时加载。已打开的页签继续保留实例，切换页签、关闭再打开侧栏时保留未保存的表单状态。
- 拆开原来的 `vendor-utils`，分别生成 VueUse core/shared、Zod 和 focus-trap 的依赖块。`@vueuse/integrations` 按实际引用加载，避免其 focus-trap 集成把弹窗代码重新拉回全局入口。
- 关闭构建清单中自动生成的 SSR `prefetch` 提示。此前仅优化账户加载后，首页仍输出 197 条预取提示，覆盖其他页面、语言和部分账户代码；这些提示使浏览器在用户没有操作时就发起下载。
- `NuxtLink` 改为鼠标悬停或键盘聚焦后预取，避免滚动到可见链接时自动加载其他页面。当前页面需要的 `modulepreload` 和实际 HTML 链接保留。
- 每次生产构建检查全局入口依赖：账户组件、CheckoutModal、地区字典、Zod、focus-trap、二维码及 Stripe SDK 一旦重新进入静态依赖，构建直接失败。统计报告为 `nuxt-i18n/output/bundle-analysis/storefront-global-entry-dependency-summary.json`。

生产构建配合本地只读后端夹具验证资源加载边界与 SSR 输出。首页 SSR 检查覆盖 9 个有内容的延迟区域、17 项延迟资源的预加载边界，并要求 HTML 不再包含构建 JS/CSS 的自动预取。英文首页、中文首页、轮组购买指南和商店页面检查 HTTP 状态、H1、canonical 和 21 条语言 alternate 链接；首页和商店还包含 JSON-LD。浏览器检查确认未操作时不下载账户与支付依赖，六个账户页签均能显示内容，地址草稿在切换页签及关闭重开侧栏后保留，购买指南资源在悬停链接后才预取；浏览器没有错误或警告。

另用相同生产产物连接已有的本地后端，通过 `test:shop-and-wheelset-guide-seo-ssr` 验证英文、中文、法文共 27 条商店和购买指南路由的服务端标题、可见 H1、搜索及社交元数据、canonical、语言链接，以及商店 CollectionPage 结构化数据。没有测量线上 LCP、TTFB 或搜索排名，也没有验证生产环境的真实商品配置。

## 第二轮：页头重复计算与未打开菜单的翻译下载

继续使用生产预览和本地后端排查。Chromium CPU 采样发现，`SiteHeader.vue` 在生成面包屑标签、路由分类和同级菜单时，反复重建 20 种语言的列表与集合，并多次遍历、归一化完整路由目录。这些计算发生在全局页头，会影响页面初始化和后续导航。

现在语言列表和小写代码集合通过 `computed` 复用；路由候选目录在当前导航中只生成一次，静态候选从同一目录筛选。路径或语言配置变化时重新计算，保留原来的路由排序、去重、动态路由与虚拟子页面发现逻辑。

同一台机器、相同测量脚本，对中文辐条长度计算器执行三次完整页面加载，CPU 采样中位数如下：

| 本地 Chromium CPU 采样口径 | 优化前 | 优化后 |
| --- | ---: | ---: |
| 除 idle 外的采样时间 | 1,766.582 ms | 297.918 ms |
| 调用帧 URL 属于 `/_nuxt/` 的采样时间 | 1,429.445 ms | 89.286 ms |

这些数值是加载期间 V8 CPU profiler 的采样时间，包含框架与页面脚本。各轮存在波动，未清空 HTTP 缓存，也未模拟移动设备或慢速网络；它们用于确认重复计算已经减少，不能当作线上 LCP、TTFB 或整页加载时间的改善幅度。

还修复了页头提前加载整个当前导航分组翻译的问题。首屏只加载当前页面导航需要的文案，用户打开 mega menu 或面包屑下拉时，再加载该分组的其他导航文案。实测未操作的轮组指南页面：

| 页面 | 延后下载的翻译 | 减少的首屏请求 | 延后下载的原始 JS |
| --- | --- | ---: | ---: |
| 英文轮组指南 | 英文轮胎指南导航 | 1 | 3,529 bytes |
| 中文轮组指南 | 英文和中文轮胎指南导航 | 2 | 7,226 bytes |

中文与英文辐条计算器没有下载无关页面翻译，前后结果一致。以上只统计被延后的翻译文件，不代表所有页面 JS 请求或总构建产物减少。第二轮完成时，全局入口及静态依赖为 825,242 bytes / 187,636 gzip bytes，与第一轮完成时基本相同；本轮收益主要来自执行时间和请求时机。

生产验证结果：

- 构建及产物检查、面包屑契约、脚本 TypeScript 检查和页脚独立翻译检查通过。
- 浏览器确认未打开菜单时没有轮胎指南翻译请求；打开 Guides mega menu 或面包屑下拉后加载两个翻译文件，中文子菜单正常显示。
- 从轮组指南经面包屑跳转到轮胎指南及其子页面，子菜单目标与当前项高亮正确。切换为英文后面包屑同步更新；在 390 × 844 视口通过 Tools 菜单跳转到辐条计算器，同级菜单正确更新。浏览器无错误或警告。
- 六条代表性路由的 SSR 状态、H1、canonical、hreflang 和面包屑文本与改动前一致。首页检查通过；英文、中文、法文共 27 条商店和轮组指南 SEO SSR 检查再次通过。

可复核数据位于 `nuxt-i18n/output/bundle-analysis/storefront-header-navigation-browser-cpu-and-translation-download-comparison.json`，优化前后 SSR 对比位于同目录的 `storefront-header-navigation-before-optimization-ssr-evidence.json` 与 `storefront-header-navigation-after-optimization-ssr-evidence.json`。浏览器测量脚本与原始输出保存在 `nuxt-i18n/output/playwright/`。

继续检查大模块时，已确认 `parse5` 用于富文本安全清洗，当前独立 chunk 约 163.8 KB 原始大小、48.4 KB gzip；不能为了体积直接删除。多语言路由也承担实际导航和抓取入口，不能仅凭未压缩大小删除语言。后续仍应以真实首屏依赖、执行采样和 SSR 输出为依据选择优化点。

## 第三轮：首页区块的 JS 下载真正按可见范围推迟

浏览器记录发现，首页此前虽然通过 `HomeDeferredSection.vue` 延迟区块激活，并移除了 SSR 预加载提示，但未滚动时仍然下载全部 9 个下方区块的 JS。部分区块距离首屏超过 5,000 像素。原因是 Vue 在执行异步组件的 hydration 策略前，会先调用该组件的 loader；直接把业务模块导入放在这个 loader 上，只能推迟激活，不能推迟下载。

现在参考 Nuxt 自身的延迟组件实现：外层异步包装只返回本地组件定义，实际模块导入放入内层组件，接近视口、鼠标进入或键盘聚焦时再激活。首次 SSR 的 9 个区块继续输出完整正文，客户端也保留这些 HTML；从其他页面返回首页时，区块则在接近视口后挂载。购买路径和精选商品的占位区块保留锚点，高度一直保留到内容挂载，保证首页两个主按钮仍能找到对应区域。

2026-10-24，本地生产预览、同一后端配置，页面完成初始加载并等待 500 ms，尚未滚动时的对比：

| Chromium 视口 | 下方区块 JS 已下载数量 | 独立 JS 文件数 | 文件分别 gzip 后合计 | 减少 |
| --- | ---: | ---: | ---: | ---: |
| 1440 × 900 | 9 → 2 | 45 → 25 | 317,468 → 240,746 bytes | 76,722 bytes（24.17%） |
| 390 × 844 | 9 → 1 | 45 → 24 | 317,468 → 238,384 bytes | 79,084 bytes（24.91%） |

对应原始 JS 分别从 1,235,792 bytes 降到 1,000,708 / 994,370 bytes。这是浏览器初始请求涉及的独立 JS 文件大小，包含当前页面、语言和共用依赖，与前两轮“全局入口及静态依赖”的统计范围不同。gzip 数字通过分别压缩这些最终文件得到，不是线上实际传输字节；不包含 CSS、图片、字体、HTML，也不表示总构建产物减少。窄屏测量使用桌面 Chromium 调整视口，未模拟手机 CPU、网络或移动端 User-Agent。线上后台内容和视口不同，实际触发的区块数量也会变化。

生产验证结果：

- 最终生产构建、原有产物与关键 CSS 检查通过；首页 SSR 检查仍确认 9 个有内容的区块、17 项延迟资源没有预加载、没有自动 JS/CSS 预取。
- 中英文首页逐项比较，9 个区块的全部文本、38 个 H1–H3 标题、canonical 与语言链接、搜索和社交元数据、JSON-LD 均与改动前一致。
- 依次滚动验证全部 9 个区块均下载并激活，FAQ 在延迟激活后可以展开完整答案。首屏测量没有加载 FAQ 使用的 `parse5` chunk；富文本清洗代码保留原有实现。
- 桌面和窄屏分别从指南通过客户端导航返回首页；未加载的购买路径和精选商品锚点各保留一个实例。点击两个首页主按钮后，仍在首页显示对应区块标题，内容加载后锚点没有重复。
- 客户端导航后再次逐一滚动，9 个区块全部下载并挂载；最终这组浏览器验证无错误或警告。本轮没有验证 QuickBuy 会话创建或支付提交。

下载对比数据保存在 `nuxt-i18n/output/bundle-analysis/storefront-homepage-true-section-download-deferral-comparison.json`，服务端正文与搜索数据对比保存在同目录的 `storefront-homepage-section-download-deferral-server-rendered-content-comparison.json`，测量脚本与交互记录位于 `nuxt-i18n/output/playwright/`。这轮证明的是首页实际初始下载边界改善，未测量线上 Core Web Vitals。

## 第四轮：工具页 FAQ 下载边界、路由结构化数据与页脚语言切换

继续检查用户端的真实请求发现，`PageFaqSlot.vue` 直接引入 FAQ 组件，导致页面底部 FAQ 的交互、答案渲染和富文本解析依赖在首屏提前下载。代表页面的 FAQ 距离页面顶部约 1,500～4,600 像素。现在使用 Nuxt 的可见范围延迟组件，在进入视口前 300 像素开始激活；SSR 仍输出完整问题和所有折叠答案。

同时处理了服务端资源提示：当前 Nuxt 渲染器会先缓存包含样式和 JS 的依赖集合，再处理延迟组件的模块排除。外层容器在 `app:rendered` 中提前移除 FAQ 子树的组件登记，避免这些登记单独触发 JS 预加载。SSR 样式已在此前收集，完整 HTML 和样式保留；其他页面通过自身静态导入需要的共享依赖仍正常预加载。实测胎压工具页不再预加载 `parse5`，而 `/website/me-this-website` 和 `/website/why-this-name` 的正文需要富文本，仍预加载安全渲染与解析器。没有修改富文本清洗或删除解析器。

2026-10-24，最终生产构建连接同一本地后端，页面初始加载完成后再等待 500 ms，尚未滚动时：

| 页面 | 初始 JS 文件数 | 各文件 gzip 合计：优化前 → 优化后 | 减少 |
| --- | ---: | ---: | ---: |
| 英文胎压工具 | 32 → 30 | 310,953 → 258,366 bytes | 52,587 bytes（16.91%） |
| 英文胎宽兼容工具 | 39 → 37 | 332,931 → 280,341 bytes | 52,590 bytes（15.80%） |
| 英文 Schwalbe 轮胎选择器 | 32 → 30 | 314,084 → 261,492 bytes | 52,592 bytes（16.74%） |
| 中文胎压工具 | 34 → 32 | 321,267 → 268,680 bytes | 52,587 bytes（16.37%） |
| 英文支付说明 | 28 → 26 | 291,221 → 238,623 bytes | 52,598 bytes（18.06%） |
| 英文联系页面 | 29 → 27 | 294,817 → 242,225 bytes | 52,592 bytes（17.84%） |

这些页面在 1440 × 900 与 390 × 844 两种视口中得到相同的初始文件集合和上述体积变化，原始 JS 减少约 174.5 KB。轮组指南的 FAQ 已位于首屏或提前加载范围内，仍及时激活；该控制页面没有节省，因组件边界增加约 1,405 gzip bytes（0.44%）。不能将本轮收益推断到所有页面。

统计只包含浏览器请求涉及的独立 `/_nuxt/` JS 文件，按最终文件分别 gzip 后求和，不是实际网络传输量，也不包含 CSS、字体、图片、HTML 或根目录脚本。最终复测禁用了浏览器缓存，避免复用旧构建的 HTML；这里比较文件集合和大小，不比较加载耗时。窄屏仍使用桌面 Chromium，未模拟手机 CPU、网络和 User-Agent，未测量线上 Core Web Vitals。

延迟 UI 的 SEO 生命周期也一并处理：FAQPage JSON-LD 由立即运行的 `PageFaqSlot` 管理，使用稳定的 head 标识随当前路由和语言更新；内部 FAQ UI 关闭自身的重复输出，独立使用 `PageFaq` 时仍默认生成结构化数据。这样在 FAQ 尚未激活就离开页面时，旧 SSR 结构化数据也能正确移除。浏览器验证从胎压页切换到胎宽兼容页，再切换中文后，均只保留一份当前页面的 FAQPage，问题数量与页面完全一致。

语言切换验证还定位到延迟页脚接管旧语言 HTML 时的 hydration 错误，具体节点为 `footer-site-overview-title`。新增 `StorefrontFooterWithLocaleSafeDeferredHydration.vue`，由四种公共布局共用：语言切换只替换尚未激活的页脚，已经激活的页脚保留实例和表单状态。

最终验证：

- 生产分析构建、原有产物与关键 CSS 检查、脚本 TypeScript 检查、页脚和反馈独立翻译检查通过。
- 7 条代表路由的 SSR 全部 FAQ 文本、完整折叠答案、H1–H3、搜索与社交元数据、canonical、hreflang 和 JSON-LD 与改动前逐项相同。
- 中英文 14 条后台 FAQ 路由，以及 4 条包含查询参数的塔基兼容 FAQ 路由，完整 SSR 答案与 JSON-LD 回归通过；27 条商店和轮组指南 SEO 路由及首页 SSR 检查通过。
- 桌面与窄屏共 12 次滚动激活及首次问题展开检查通过，完整答案和唯一 FAQPage 保留。切换页面和语言时，当前问题与结构化数据保持一致。
- 四种布局分别验证“未激活页脚前切中文，再滚动”和“已激活页脚填写订阅邮箱草稿，再切英文”；草稿全部保留，浏览器没有错误或警告。只填写表单，没有提交订阅。

最终下载与 SSR 对比数据为 `nuxt-i18n/output/bundle-analysis/storefront-page-faq-visible-download-deferral-comparison.json`。解析器预加载控制数据、最终生产构建日志位于同目录；导航、语言切换、折叠答案和四种布局的浏览器脚本与结果位于 `nuxt-i18n/output/playwright/`。

## 生成可视化报告

在 `nuxt-i18n` 或 `go-backend/web/admin` 目录执行：

```powershell
npm run build:analyze
```

命令沿用各自的生产构建与检查流程，并额外启用 `rollup-plugin-visualizer`：

| 项目 | 报告路径（相对于对应项目目录） |
| --- | --- |
| Nuxt 前台 | `output/bundle-analysis/storefront-client-treemap.html` |
| Vue 后台 | `output/bundle-analysis/admin-client-treemap.html` |

用浏览器打开 HTML，可以搜索依赖、查看引用关系，比较 rendered、gzip 和 Brotli 体积。模块统计与最终 chunk 的压缩体积口径不同，不能把各模块的 gzip 数字直接相加当作网络下载量。

普通 `npm run build` 不启用分析器。报告保存在 Git 已忽略的 `output/`，不会写入后台 `dist/` 或前台 `.output/public/` 发布目录。后台分析命令使用生产 mode，保持 `.env.production` 的加载行为。

## 国家和省州数据

2026-10-24 的生产构建检查结果：

- 前台 `app/data/checkoutStateProvinceCatalog.ts` 已经按支持的国家使用命名导入。报告中 `country-region-data` 模块的 rendered 体积为 53,599 bytes，gzip 为 22,202 bytes；并非全量引入。原有国家范围和省州代码用于地址、Stripe Express Checkout 与税率匹配，不应为了体积随意删除。
- 后台税率页支持全球运费范围，确实需要完整的国家和省州目录。该页面原本已按路由加载，现在进一步把字典延迟到“覆盖检查”和“结算规则”页签。查看配置和快照时只加载税率页代码。
- 字典请求在页面实例内复用；加载失败会清除请求缓存，通过现有刷新操作重试。目录未就绪时不能创建或编辑规则，避免把加载失败误认为国家没有省州代码。目录使用 `shallowRef`，避免把整份字典转换为深层响应式对象。

同一工作区前后两次后台构建对比：

| 后台资源 | 原始大小 | gzip 大小 |
| --- | ---: | ---: |
| 优化前税率页 chunk | 142,655 bytes | 约 53.60 KB |
| 优化后税率页 chunk | 48,014 bytes | 13,567 bytes |
| 新增按需地区字典 chunk | 95,143 bytes | 39,301 bytes |

这是推迟非必要下载带来的改善。进入覆盖检查或结算规则时仍然需要加载完整目录，不能把它描述为总产物减少约 95 KB。chunk 哈希和具体体积可能随依赖、业务代码和构建环境变化。

## 为什么没有照搬全局激进参数

以实际构建日志为准：前台 Nuxt 4.4.8 的 builder 使用 Vite 7.3.6 / Rollup；后台使用 Vite 8.1.5 / Rolldown 1.1.5。前台顶层 `node_modules/vite` 的版本不能代表 Nuxt builder 实际使用的版本。

| 建议参数 | 实际含义与处理 |
| --- | --- |
| `moduleSideEffects: 'no-external'` | 假设被 external 的模块没有导入副作用；不会自动对已打包的所有 npm 依赖进行更深入检查。保留默认行为和依赖声明的副作用元数据。 |
| `propertyReadSideEffects: false` | 可以删除结果未使用的属性读取，包括有副作用的 getter 和会抛错的读取。没有针对应用与所有依赖完成验证前，不全局启用。 |
| `tryCatchDeoptimization: false` | Rollup 支持，但可能改变依靠异常进行能力探测的代码。后台当前 Rolldown 的公开选项不支持该参数，不能向两个构建器同时复制。 |

分析还确认后台图表只保留了折线图及其依赖；ECharts 的多个 vendor chunk 并不代表所有图表都被全量打包。继续优化时应先在报告里检查具体模块和引用路径，再调整导入或加载时机。

## 已执行验证

- 两端 `npm run build:analyze`，包含原有构建产物检查。
- 后台 `npm run typecheck`。
- 后台税率适用范围注释与历史快照兼容测试。
- 前台 `npm run test:checkout-state-province-catalog`。
- 前台 `npm run scripts:typecheck`、`npm run test:footer-independent-translation-loading`。
- 前台 `HOME_SSR_URL` 指向生产预览时的 `npm run check:home-ssr`；新增的预取检查已能复现并拦截优化前的 197 条自动预取提示。
- 前台 `npm run test:shop-and-wheelset-guide-seo-ssr`：将 `SHOP_AND_WHEELSET_GUIDE_SSR_ORIGIN` 指向生产预览，`SHOP_AND_WHEELSET_GUIDE_BACKEND_ORIGIN` 指向本地后端，27 条路由全部通过。
- 真实浏览器验证账户六个页签、地址表单状态保留，以及链接在用户悬停后才预取。
- 检查生产税率页的静态依赖不包含地区字典；按需 chunk 的完整国家、省州数据与 npm 包原始导出一致。

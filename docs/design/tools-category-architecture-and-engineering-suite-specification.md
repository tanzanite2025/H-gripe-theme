# 🛠️ 独立工程工具体系 (/tools) 全景架构与迁移实施规范
## Architecture & Migration Specification for Engineering Tools Suite

> **文档版本**: `v1.10-Engineering`
> **文档位置**: `docs/design/tools-category-architecture-and-engineering-suite-specification.md`
> **适用范围**: 针对现有代码库中已实现的所有计算器、适配器、选型矩阵及仿真看板，判定其应**保留完整混合页面还是整体迁移为工具页面**，并规范后续导航重组。
>
> 🛑 **【最高核心原则：算法零改动，纯架构重组】**
> **迁移复用现有算法、数据和交互契约，不推定全部工程模型已验证正确。已知标准冲突和算法复核应另立任务，不混入路由迁移。**
> **本文档严禁编写、推导或重新设计任何数学公式与物理算法，绝不增加任何额外计算需求，杜绝技术跑偏！**
> **本文档唯一目标是：全面盘点现有分散在 `/guides` 中的各块工具资产，先依据搜索意图、页面主任务、内容独立性与重复风险判定应保留完整混合页面还是整体迁移为工具页面，再理清路由、菜单、面包屑与迁移映射，确保平稳落地、不落下一块。工具页必须是可以独立完成任务的完整页面；如果混合页面以后迁移到 `/tools`，说明、交互、结果解释、数据来源和使用边界一并迁移，不拆成一个地方放说明、另一个地方放计算器。**

> **路由核对口径**：下文“现有访问路径”以 `nuxt-i18n/public/storefront-route-manifest.json` 与页面 `definePageMeta` 的实际结果为准。`?tab=...` 只是兼容性查询参数，不作为独立规范路由。`/tools` 与 `/guides` 是信息架构标签，但 `/tools` 对用户有明确承诺：打开后应能直接交互、查表或得到结果，并能在同一页面理解输入、输出和限制。迁移前必须确认页面的主任务、内容完整性和重复风险；混合页面只有在整体内容准备好后才能迁移；按用户最新决定，完成迁移后旧路径退役为 404，不建立别名、兼容页面或迁移重定向。

---

## 本批实施与后续顺序（2026-10-13）

已完成工具大厅、Schwalbe、品牌轮组辐条规格、辐条长度计算器、轮组编法拓扑、动态胎压、不锈钢辐条应力释放力学、内胎气门嘴匹配、塔基与飞轮兼容助手和胎圈宽度兼容引擎的整页迁移。工具大厅展示九项已有完整工具，不把大厅自身再列为同级工具卡片，不预建空分类，不展示规划工具卡片。SPOKEGUIDES 已清空，分类路由 `/guides/spokeguides` 与其唯一子页旧路径均退役为 404；后续新增辐条指南必须先有独立完整内容，不得恢复一个空分类路由。

- 新路由：`/tools`（大厅）以及九个具体工具 `/tools/schwalbe-tire-selector`、`/tools/brand-wheelset-spoke-specs`、`/tools/spoke-length-calculator`、`/tools/spoke-lacing-topology-simulator`、`/tools/dynamic-tire-pressure-calculator`、`/tools/spoke-stress-relief-mechanics`、`/tools/valve-length-fitment-calculator`、`/tools/freehub-cassette-fitment-engine`、`/tools/tire-rim-width-compatibility-engine`；已迁移旧路径及其全部语言、查询参数变体退役为 404，不设置 301、路由别名或兼容页面。
- Schwalbe 介绍、搜索、筛选、结果卡片、原生分页、来源日期和限制继续复用同一个组件。组件目录和翻译 namespace 不因公开路由迁移重命名。回归发现并修复筛选抽屉 `canApply` 默认值导致按钮禁用，以及可见条件不变时旧查询参数未清理的两个既有问题；目录查询与工程算法不变。
- FAQ manifest key 保留 `guides-tireguides-schwalbe-tire-selector`，page_id 保留 `guides-schwalbe-tire-selector`。后台启动自动同步路由台账与 FAQ；已有问答、标题和编辑分组保留。上线需确认启动同步成功，失败可使用已有 URL 台账同步或 FAQ routes/sync。
- canonical/hreflang 使用新路径；仅合法 `page=N>1` 保留可收录分页，筛选与搜索为 `noindex,follow`。Sitemap 排除退役旧路径和已有重定向源路径，防止未同步 FAQ 再写入旧 URL。
- 大厅采用公开内容缓存；Schwalbe、品牌规格目录页、辐条长度计算器与轮组拓扑明确 `no-store`。退役地址显式 `cache:false` 与 `no-store`，防止继承 `/guides/**` 缓存并继续返回旧页面。未来计算器逐页选择策略，不为整个 `/tools/**` 统一缓存。
- 品牌规格整页保留品牌筛选、型号/辐条搜索、分页、左右侧长度、头型、编法、条帽规格、修补包商品区和 Shimano/SAPIM 说明；商品详情型号链接保留 `model` 查询及行定位 hash。公开 API 和静态备用目录继续复用，FAQ key、page_id、反馈键均保留 `guides-spokeguides-brand-wheelset-spoke-specs`。
- 品牌规格迁移前页面实际公开可收录，但台账仍为 `is_indexable:false`；本批对齐为中英文主页面可收录，合法分页维持原规则，型号/搜索/非法查询 `noindex,follow`。仅中英文正文已经发布，另外 18 种语言保留访问但 `noindex,follow`，hreflang、台账 `sitemap_locales` 和最终 sitemap XML 只声明中英文及 x-default，不把翻译回退当作完整译文。
- 辐条长度已完成整页迁移：五步输入向导与第六步计算结果、参数说明、测量图示、数据目录和反馈整体保留。FAQ key、page_id、反馈键均保留 `guides-spokeguides-spoke-length-calculator`；正文只发布中英文，其他 18 种语言可访问但 `noindex,follow`，hreflang 与最终 sitemap 只声明中英文及 x-default。
- 轮组编法拓扑已整页迁移：头型、孔位、交叉数、2:1/G3、双法兰参数、视图、图层、SVG/剖面及工程说明一起保留。FAQ manifest key 仍为 `guides-wheelset-buyers-spoke-lacing-topology`，真实 page_id 仍为 `resources-wheelset-lacing-topology`，后台编辑分组保留 `guides`；仅中英文正文已发布，其他 18 种语言继续 `noindex,follow`，hreflang 与 sitemap 只声明中英文及 x-default。保留原 TechArticle/FAQPage Schema 和几何预览边界，不新增物理安全承诺。
- 动态胎压已完成整页迁移：原页面的 `calculator` 与 `details` Tab、外胎型号目录、速度/载荷/倾角/姿势输入、动态受力 API、最低胎压比较、湿地压力演示、模型说明、数据来源、结果解释、风险提示和反馈线程整体保留。FAQ manifest key 与 page_id 继续使用 `guides-tireguides-tire-pressure`，后台分组仍为 `guides`；20 种语言保留可访问页面和对称 canonical/hreflang，旧路径及全部查询变体 404，无 Location。后端算法、参考数据和计算 composable 未修改，新页明确 `no-store`。
- 胎压迁移前已保存 SSR、head/canonical/hreflang/Schema、FAQ、reference-data、动态 API 与 10 组浏览器交互基线；迁移验收必须逐项对照这些记录，并将迁移差异单独提交。不锈钢位错页面已按完整页面边界迁移，说明、材料目录、计算器、结果解释和 FAQ 共用一个工具 URL。
- 内胎页面已完成整页迁移：ETRTO/尺寸、气嘴孔径与类型、丁基/乳胶/TPU 材质、安装边界、框高与气嘴长度计算器、结果解释、参考矩阵、数据来源、限制和 FAQ 均保留在 `/tools/valve-length-fitment-calculator`。旧 `/guides/tireguides/choose-inner-tube` 及其语言、Tab、分享参数变体返回 404；中英文进入 sitemap，其他语言保留访问但 `noindex,follow`，不拆回指南页。
- 胎圈宽度已完成整页迁移：计算、模拟、DT Swiss、Schwalbe 四个页签、输入说明、结果解释、来源图表、限制、FAQ、Schema 和反馈共用 `/tools/tire-rim-width-compatibility-engine`；现有 v11 记录的 16 组最佳算法与正常范围冲突仍原样保留，本批不协调算法，不把它改造成静态指南，也不拆成“指南说明页 + 计算器页”。旧路径及全部语言、查询参数变体 404，FAQ 身份保留。
- 塔基与飞轮兼容助手已完成整页迁移：完整选型说明、三步流程、品牌/飞轮输入、兼容结果、实物图片、标准矩阵、动态兼容矩阵、HowTo、FAQ、Schema、结果限制和反馈均保留在 `/tools/freehub-cassette-fitment-engine`；复用 Go 兼容性接口，不修改算法。旧 `/guides/wheelset-buyers/choose-freehub` 及 20 种语言、`tab`、品牌/规格查询变体均返回 404，无 `Location`、别名或兼容页；FAQ 身份保留，动态矩阵接口不可用时页面仍使用静态备用数据并明确验证边界。

### 第一批验收记录（工具大厅与 Schwalbe）

- 路由 manifest、面包屑、页脚契约、脚本 TypeScript 和 Nuxt 类型检查通过。
- 导航浏览器回归覆盖 320、390、768、1024、1279、1440px 宽度及 390×300 矮窗口：扳手位于商品图标右侧、主导航不重复工具文字、工具和商品菜单共用页头底部起点、外框贴合内容；切换、再次点击、Escape、收起按钮、外部点击关闭和页脚工具链接均通过。工具弹层在 390×844 手机窗口实测为 158px 高，内容未超过可视高度时不填满屏幕。
- 实际 HTTP 覆盖 20 种语言 × 无查询、分页、搜索筛选，共 60 条旧路径 404，并断言不存在迁移 Location；中英文 SSR 验证完整介绍与搜索面板、canonical/hreflang、合法分页及筛选 noindex；动态 sitemap 包含新路径并排除旧路径。
- Go 回归使用当前 manifest 验证地址迁移和原 FAQ 问答保留；本地后台启动同步已完成，manifest_version 为 2026-10-12。
- Schwalbe 19 项 Chromium 回归全部通过，覆盖宽屏、手机、介绍页签、轮径、筛选、排序、废弃 URL 条件清理及分页前进后退。
- 本次验证在本地开发服务完成，Nuxt 开发响应会将 Cache-Control 替换为 no-transform；新页 no-store 已通过路由配置断言。HTTP 验证脚本默认要求生产 no-store，本地验证显式开启 STOREFRONT_ALLOW_DEVELOPMENT_CACHE_HEADERS=true。尚未执行生产构建或部署。
- 全量 i18n 聚合一致性检查仍报告既有 checkout namespace 不同步；基线与现状均存在，本批仅新增 toolsNavigation，所有语言的源文件与聚合内容一致。

### 第二批验收记录（品牌轮组辐条规格）

- 整页模板、规格读取与备用目录、筛选/搜索/分页逻辑、规格格式化及样式与迁移前逐项比对一致；仅公开路径、导航元数据和未发布语言索引规则变化。
- 保存中英文默认页、第二页和型号直达页共 6 组 SSR 基线；迁移后全部规格行、介绍及 Shimano/SAPIM 说明与基线一致，62 个公开 API 型号的全部字段未变。
- 20 种语言 × 默认、分页、型号查询共 60 条旧 URL 返回 404，无 Location。新页保留合法分页 canonical、型号查询 noindex 和行定位；其他 18 种语言可访问但不收录。
- 动态 sitemap 与 20 份最终 XML 均验证新旧地址和语言范围；最终 XML 钩子排除退役工具地址，品牌规格只声明中英文和 x-default，防止自动静态路由或 hreflang 再加入旧地址及回退译文。
- Go 回归验证全部语言原 FAQ 内容、标题、分组和 page_id 保留，中英文公开地址更新为工具路径，未发布语言 FAQ 台账标记 stale，旧公开地址不再解析；本地后台启动同步成功。
- 浏览器验证品牌筛选、型号搜索、无结果提示、分页前进后退、越界页码归一化、型号 hash 直达、修补包锚点、两项工具大厅和页脚标题/独立箭头。导航覆盖 320、390、768、1024、1279、1440px 和 390×300 矮窗口，沿用完整页头底部起点和实际内容高度；390px 手机和 768px 平板的三项工具菜单外框实测约 206px。
- 后续回溯确认新增工具导航使绝对居中的六项导航与 Logo 相撞，不能以按钮右侧点击作为验收。现已为 Logo、导航和右侧操作分配独立网格空间，导航间距自适应；1280、1366、1440、1920px 均验证无重叠，商品按钮中心正常点击，工具/商品弹层切换、再次点击关闭和 Escape 正常。
- 页脚、面包屑、route manifest、脚本 TypeScript、Nuxt 类型检查及 FAQ 回归通过。仍为本地开发验收，no-store 由路由配置断言，开发 HTTP 显式允许 no-transform；未执行生产部署。20 种语言只增补 toolsNavigation，已有 FAQ 等未提交修改未混入迁移提交。

### 开始第三批前的回溯验收

- Schwalbe 与品牌规格的 HTTP、SEO、FAQ、最终 sitemap 及内容/数据基线重新通过，合计 120 条旧地址变体返回 404。
- 浏览器 99 项检查通过，覆盖筛选、搜索、分页、型号直达、修补包、大厅、页脚及 320–1920px 的导航。弹层起点仍为完整页头底部减 1px，实际内容决定高度，390×300 矮窗口才按可视上限滚动；未将弹层固定为屏幕高度。
- 回溯修复验收通过后才进入辐条长度整页迁移，算法和计算 API 不变。

### 第三次整页迁移验收记录（辐条长度）

- 完整模板、样式及原有工作区 SEO 改动逐字比对保留；所有向导组件、状态管理、计算请求、后端算法和目录读取均未改动。原有未提交改动随整页移动，提交仅包含迁移差异。
- 保存中英文完整 SSR 内容和图像，以及 J 弯头对称轮组、直拉非对称轮组两种输入；迁移后 4 条完整流程、24 个步骤内容、8 次前后轮 API 请求/响应和可见结果与基线精确一致。这是迁移回归，不代表另行证明工程算法正确。
- 20 种语言 × 默认、分页、计算参数查询共 60 条旧 URL 返回 404，无 Location；新页 canonical 不携带未使用查询参数。动态 sitemap 与 20 份最终 XML 只发布中英文，通用工具 sitemap 钩子按台账语言范围过滤 URL 与 alternatives，并排除全部退役工具路径。
- FAQ 回归验证原有各语言标题、问答、分组和 page_id 保留，中英文公开路径更新，未发布语言标记 stale，旧公开地址不再解析。仍复用页面原有反馈区，未添加重复反馈组件。
- 大厅和页脚列出三项具体工具，页脚“工具”标题单独链接大厅；指南目录和导航不再出现旧计算器入口。93 项浏览器检查覆盖 320–1920px 及 390×300 矮窗口；响应式扳手和桌面菜单继续共用弹层，起点为完整页头底部减 1px，高度按内容。三项工具加入后手机/平板外框约 256px，1280–1440px 桌面约 212px；矮窗口按上限 156px 滚动，未固定全屏高度。
- 补查俄语、德语、阿拉伯语 RTL 和中文的 1280px 导航，均无 Logo/操作区重叠，按钮中心可点击。客户端 query/hash 跳转发现公共 SEO 状态会清除语言覆盖，已由本页在导航完成后恢复已发布语言；浏览器确认跳转后仍为中英文及 x-default，canonical 与 hreflang 均不携带未使用查询或 hash。
- 页脚、面包屑、route manifest、脚本 TypeScript、Nuxt 类型检查和三项迁移的 FAQ 回归通过。仍为本地开发验收，no-store 通过路由配置验证，HTTP 显式允许开发 no-transform；尚未执行生产部署。本地计算验收启动参数需将测试来源端口加入 CORS_ORIGINS，不能因此关闭生产 CSRF 保护。

### 第四次整页迁移验收记录（轮组编法拓扑）

- 完整模板、样式、绘图和交互实现与迁移前逐字一致；仅公开路径、导航元数据和 SEO 路由处理变化。删除 `nuxt.config.ts` 中把平铺页面挂到轮组指南下的特殊注册，新工具文件显式声明唯一公开路径。
- 保存中英文各 23 组、合计 46 组完整交互基线；迁移后全部可见正文、输入值、控件状态、轮组 SVG、双法兰剖面 SVG 和 API 请求/完整响应一致。覆盖 J 弯头/直拉、16–36H、0–4X、2:1、21H G3、几何参数、轮侧视图、图层和非法 G3 间距保留上一幅有效绘图。这是迁移回归，不是新的工程正确性或物理安全证明。
- 20 种语言 × 三种旧地址 × 默认、分页和拓扑查询，共 180 条退役 URL 均为 404，无 Location。完整中英文 SSR 正文、原 TechArticle/FAQPage Schema、真实后台 FAQ 问答和默认几何响应与基线一致；FAQ page_id `resources-wheelset-lacing-topology` 和编辑分组保留，旧公开 FAQ 地址不再解析。
- 新页 `no-store`，全部退役工具路径显式 `cache:false`，大厅继续独立公开缓存。动态 sitemap 和 20 份最终 XML 排除退役地址，轮组拓扑只发布中英文；18 种回退语言继续可访问但不收录。
- 123 项导航检查覆盖 320–1920px 与 390×300 矮窗口：大厅四项完整工具，Header 五个目的地（大厅加四工具），页脚标题链接大厅且下面无重复大厅项；扳手仍在商品图标右侧，切换、再次点击、Escape、收起和外部点击正常。弹层起点为完整页头底部减 1px；手机/平板外框约 305px，桌面约 212px，矮窗口才按 156px 上限滚动。
- 另验证俄语、德语、阿拉伯语 RTL 和中文的桌面按钮中心点击及 Logo/操作区无重叠；客户端 query/hash 跳转和中英文/德语切换后 canonical、hreflang 与收录规则保持正确，不把参数或 hash 带入规范 URL。
- 页脚、面包屑、route manifest、拓扑选择契约、脚本 TypeScript、Nuxt 类型检查、Go 拓扑领域/HTTP 测试和四项迁移 FAQ 回归通过；辐条长度的 HTTP/SEO/sitemap 回归也通过。20 种语言只增补工具导航翻译，126 项原有未提交工作保留，提交仅包含迁移差异。仍为本地开发验收，HTTP 显式允许 `no-transform`，生产 `no-store` 由配置断言；未执行生产构建或部署。

### 第五次整页迁移验收记录（动态胎压）

- 迁移前基线目录为 `tire_pressure_whole_page_migration_baseline_20261012_225221`，保存了 Git/dirty 文件快照、中文与英文 SSR 正文和 head、canonical/hreflang/Schema、FAQ API、产品胎压 reference-data API、动态受力 API 请求基线及 10 组浏览器状态。迁移提交只允许包含本轮路由、导航、文案、契约和验收差异，原有用户修改必须保留。
- 胎压页面作为完整工具工作台迁移到 `/tools/dynamic-tire-pressure-calculator`：`calculator` 与 `details` Tab、外胎型号搜索与选择、速度/骑手重量/车辆重量/倾角/姿势、动态受力请求、最低胎压对比、湿地压力演示、模型信息、数据来源、限制、结果解释、风险提示和 `guides-tire-pressure` 反馈线程全部保留；不修改胎压算法、计算 composable、后端路由或 reference-data 契约。
- 新页面 footer 与 Tools breadcrumb 使用 `toolsNavigation.dynamicTirePressureCalculatorLabel`；大厅、电脑 Header、手机/平板扳手菜单和页脚列出五个具体工具，页脚标题只链接 `/tools` 大厅。指南 Tab、指南页内容、面包屑和 Tubeless CTA 不再生成旧胎压入口，CTA 直接指向新工具 URL。
- 旧 `/guides/tireguides/tire-pressure` 及 20 种语言前缀、查询参数和 hash 变体必须 HTTP 404、无 `Location`、无别名、无兼容页面；旧路径显式 `cache:false` 与 `no-store`。新页 canonical/hreflang、robots、Schema、动态 sitemap 和 20 份最终 XML 必须只声明新地址；FAQ page_id、route key 和问答内容保留，旧公开 FAQ 地址不再解析。
- SSR 正文、两个 Tab、目录和 SEO head 与迁移前基线逐项对照；浏览器 10 组输入与 Tab 状态、reference-data/catalog 和动态 API 结果逐项对照。工具弹层继续从完整页头底部减 1px 展开，内容决定高度，手机/平板扳手与电脑入口共用同一弹层。验收在本地服务完成，不部署生产。

### 第七次整页迁移验收记录（内胎气门嘴匹配）

- 迁移前基线目录为 `C:\Users\P16V\AppData\Local\Temp\inner_tube_valve_fitment_whole_page_migration_baseline_20261013`，保存当前脏工作区快照、内胎组件、计算 composable、SEO/Schema、路由、导航和测试文件；上一提交及无关用户修改没有混入本轮迁移。
- 内胎完整页面迁移到 `/tools/valve-length-fitment-calculator`：ETRTO 与宽度范围、Presta/Schrader/Dunlop 气嘴、丁基/乳胶/TPU 材质、安装边界、框高与气嘴长度计算器、自动/手动试配、结果解释、参考矩阵、HowTo、数据来源、限制和 FAQ 均保留在同一个页面。没有留下旧指南正文，也没有抽出只有输入面板的半页工具。
- 工具大厅、桌面 Header、移动端/平板扳手菜单和页脚新增具体工具；页脚“工具”标题仍只链接 `/tools`，不会把大厅作为同级子链接。旧 `/guides/tireguides/choose-inner-tube` 及语言、`tab`、分享参数和 hash 变体返回 404，无 `Location`、别名或兼容页面。
- route manifest、面包屑、页脚、i18n、脚本 TypeScript、Nuxt 类型检查、内胎分享状态/Tab 导航/发布语言策略、Go FAQ 路由迁移回归均通过。SSR 已验证英文、中文及回退语言的正文、双 Tab、canonical/hreflang、`noindex,follow`、Article、WebApplication、HowTo、Dataset 和社交预览图；本地 Go 后端未启动时动态 sitemap 接口为空，因此该项需在后台启动后再执行最终 sitemap 验收。
- FAQ page_id、route key 和问答内容保持 `guides-tireguides-choose-inner-tube`；中英文进入 sitemap，其他语言继续可访问但由发布策略标记 `noindex,follow`。后续任何混合页面迁移仍须先保存整页基线，不得只迁移单独交互组件。
- 塔基迁移后 FAQ page_id、route key 和问答内容继续保持 `guides-wheelset-buyers-choose-freehub`；工具页与轮组指南均不再保留内嵌的小计算器，避免说明与交互割裂。工具页使用 `no-store`，大厅、Header、移动端/平板扳手菜单和页脚均只列具体工具，页脚“工具”标题单独链接大厅。

### 第八次整页迁移验收记录（塔基与飞轮兼容助手）

- 迁移前基线目录为 `C:\Users\P16V\AppData\Local\Temp\freehub_cassette_fitment_whole_page_migration_baseline_20261013`，保存 Git/dirty 文件快照、旧路由中英文 SSR、组件与翻译、FAQ、路由、导航、Schema 和 HTTP 代表性变体；本轮只暂存迁移差异，工作区其他修改保持不动。
- `/tools/freehub-cassette-fitment-engine` 保留完整页面闭环：页面标题与介绍、三步选型说明、品牌与飞轮规格选择、动态兼容结果、静态备用结果、实物图片、标准矩阵、兼容性速查、技术参考、HowTo、FAQ、Schema、结果限制和反馈均在同一 URL；不拆成指南说明页与单独计算器页，也不删除 `WheelsetChooseFreehubSection` 与 `FreehubGroupsetHelper` 的领域交互。
- 旧 `/guides/wheelset-buyers/choose-freehub` 的 20 种语言、无查询、`tab`、品牌和飞轮规格查询变体共 60 条真实 HTTP 请求均为 404，无 `Location`；旧路由显式 `cache:false`/`no-store`。新页 20 种语言 SSR 200，canonical/hreflang 指向对应工具路径，路由 manifest、工具大厅、桌面 Header、手机/平板扳手菜单和页脚均已更新。
- Go FAQ 路由迁移回归、Nuxt 脚本 TypeScript、页脚/面包屑合同和 41 条路由清单通过；整页回归脚本通过 60 个退役变体和 20 个工具页面。当前本地后台未启动，动态矩阵 HTTP 返回 502，因此本轮仅以静态 SSR/备用矩阵完成页面回归，Go 接口自身测试已通过；生产部署前仍需在后台可用时复跑动态矩阵和最终 sitemap。

### 第九次整页迁移验收记录（胎圈宽度兼容引擎）

- 迁移前基线目录为 `C:\Users\P16V\AppData\Local\Temp\tire_rim_width_compatibility_whole_page_migration_baseline_20261013`，保存旧路由正文长度、Git/dirty 文件状态、组件、页面、导航和本架构文档；只允许提交本轮迁移差异，工作区其他修改继续保留。
- 胎圈宽度页面整体迁移到 `/tools/tire-rim-width-compatibility-engine`：计算、模拟、DT Swiss 参考、Schwalbe 参考四个页签，输入和参数说明、分层结果、截面图、来源图表、限制、FAQ、TechArticle/Dataset Schema 和反馈都保留在同一 URL。四个页签不拆成指南正文与独立计算器，也没有改变原有计算 composable、参考数据或后端契约。
- 旧 `/guides/tireguides/tire-rim-width-compatibility` 的 20 种语言、无查询、旧 Tab、参考 Tab 和代表性尺寸查询共 80 条 HTTP 变体均返回 404，无 `Location`；旧路径显式 `cache:false`/`no-store`。新工具 20 种语言 SSR 200，英文和简体中文进入 sitemap，其他语言可访问但 `noindex,follow`，canonical/hreflang 使用新工具路径。
- 工具大厅、桌面 Header、手机/平板扳手菜单、页脚和面包屑均指向九个具体工具；页脚“工具”标题仍只链接 `/tools` 大厅，不重复列大厅卡片。根语言包同步补齐塔基与胎圈宽度工具标签，避免 Header/Footer 运行时缺失翻译。
- 胎圈宽度整页回归脚本通过 80 个退役 URL 变体、20 个本地化工具页、四个页签、canonical/robots/Schema、工具大厅、导航、页脚和 route manifest；旧路由没有兼容别名。动态工程 API 未作为迁移条件，页面保留既有静态参考和结果边界；生产部署前仍需在后台可用环境复跑动态接口和最终 sitemap。

## 目录
1. [一、 为什么要做这次架构重组（纯信息架构与体验升级）](#一-为什么要做这次架构重组纯信息架构与体验升级)
2. [二、 “工具 (Tools) + 指南 (Guides)”双飞轮架构](#二-工具-tools--指南-guides双飞轮架构)
3. [三、 现有工程工具资产全盘清点与迁移映射表（完整梳理不遗漏）](#三-现有工程工具资产全盘清点与迁移映射表完整梳理不遗漏)
   - [3.1 板块 A：轮胎、车圈与气门系统](#31-板块-a轮胎车圈与气门系统)
   - [3.2 板块 B：轮组、辐条与力学工程](#32-板块-b轮组辐条与力学工程)
   - [3.3 板块 C：传动系统与整车兼容性](#33-板块-c传动系统与整车兼容性)
4. [四、 导航体系与菜单布局改造方案](#四-导航体系与菜单布局改造方案)
5. [五、 整页迁移与旧路径退役策略](#五-整页迁移与旧路径退役策略)
6. [六、 独立工具页的 SEO/GEO 外壳与 Schema 规范](#六-独立工具页的-seogeo-外壳与-schema-规范)
7. [七、 研发落地分批拆解检查清单](#七-研发落地分批拆解检查清单)

---

## 一、 为什么要做这次架构重组（纯信息架构与体验升级）

项目已有真实计算器、查表工具和工程交互页面。本次迁移复用现有行为，不代表现有算法和标准已经通过完整正确性审计。然而在**信息架构（Information Architecture）**上，这些工具大多被塞在 `/guides`（指南）的深层 Tab、子目录或折叠面板中：

1. **用户找工具路径过长**：用户想算气门嘴或辐条长度，必须先打开指南文章，再在一堆科普图文中找“计算器”Tab 切换，操作链路繁琐；
2. **SEO/GEO 意图表达不清**：搜索引擎和检索系统不会仅凭 `/tools` 或 `/guides` 路径给页面固定加权，主要还是根据可抓取内容、查询意图匹配、页面质量、内部链接和用户任务完成度判断相关性。把一个本可独立完成任务的工具深藏在文章 Tab 中，可能让用户和系统都更难识别其主任务；但如果图文和计算器本来服务同一意图，机械拆分反而会制造近重复页面；
3. **页面复杂度过高**：一个指南页面既要加载长图文，又要加载复杂的计算组件，导致单个页面体积臃肿、维护困难。

**解决方向**：当页面的主要任务是计算、查表、匹配或仿真时，应把它视为工具候选，因为“工具”向用户明确承诺可以直接操作并得到结果。工具页不应只留下一个输入面板：原理简介、参数说明、结果解释、数据来源和使用边界都应与交互放在同一个完整页面中。对于当前同时承担指南和交互的页面，先保留现状；如果后续确认工具任务是主任务，则整体迁移页面到 `/tools`，而不是把说明留在 `/guides`、把计算器单独抽走。只有确实存在两份独立内容、且各自服务不同任务时，才另建指南 URL。这样既能缩短工具任务的到达路径，也不会因机械拆分造成近重复页面、上下文断裂或权重分散。

### 1. 工具定位先于路由迁移

“工具”是一个面向人的产品承诺：用户打开页面后可以直接输入、筛选、查表、模拟或计算，并在同一页面得到可解释的结果。对于已有真实交互且以完成任务为主的资产，工具应作为主要产品定位；页面包含较多图文说明，不会自动使它失去工具属性。评估时应先决定工具任务是否为主任务，再规划迁移路径：

| 判断维度 | 需要回答的问题 | 对路由的影响 |
| :--- | :--- | :--- |
| 搜索意图 | 用户是在寻找一个数值/兼容结果，还是在学习原理、测量方法或选型背景？ | 两种意图清晰分离，才有建立两个 URL 的理由；意图相同则保留一个主 URL。 |
| 页面主任务 | 首屏和主要操作是否引导用户输入、筛选、查表、模拟或计算并得到结果？ | 是则优先定位为工具；指南可以是辅助内容，但不应遮挡或替代主要交互。 |
| 工具完整性 | 同一页面是否包含必要的输入说明、结果解释、数据来源和使用边界？ | 如果信息散落在多个 Tab 或区块，整体迁移时一并保留；不能拆成说明页和计算器页。 |
| 交互真实性 | 控件是否真实可用，并产生与用户输入相关的结果？ | 只有静态表格、图片或规划中的契约时，仍是指南/参考页，不标记成工具。 |
| 重复风险 | 迁移是否会同时留下内容高度相同的旧、新 URL？ | 整页迁移后只保留一个可索引主 URL；旧路径在整页验收通过后退役为 404，不保留重复页或迁移重定向。 |
| 独立指南价值 | 是否另有一份不依赖工具也能完成的原理、测量或选型指南？ | 只有确有独立内容和不同用户任务时才保留/建立单独指南 URL；不得把工具页面的必要说明剥离出去。 |

据此采用以下顺序：

1. **已有真实交互，且交互是主任务**：产品定位为工具。先保持当前页面完整可用；迁移时把页面整体搬到 `/tools`，保留说明、交互、结果解释、数据来源和限制。
2. **已有交互，但主要任务仍是阅读或参考**：暂时保留完整页面和当前 URL；如后续决定让工具成为主任务，也整体迁移，不抽走计算器。
3. **没有真实交互**：保留指南/参考定位，不得只靠名称、Schema 或规划卡片标成工具。
4. **确实另有独立指南**：工具页和指南页可以并存，但两者必须各自完整、各自解决任务；禁止将一页的说明和另一页的计算器拼成一个任务闭环。

如果当前页面没有真实可用的交互或查表能力，只能先作为指南/参考页，不得为了填充 `/tools` 目录创建一个名义上的工具页。反过来，如果交互已经是页面主任务，也不要为了保留“指南”标签而把工具降格为只有导流卡片的页面。

### 2. 本轮评估的决策基线

以下结论是本规范后续实施的默认边界，除非新的页面证据、搜索数据或验收结果明确改变它们，否则后续任务不得把“完整工具迁移”改成“只迁移计算器”：

| 资产 | 当前决策 | 允许的下一步 |
| :--- | :--- | :--- |
| 内胎气门嘴匹配 | 已完成整体迁移 | 新路径 `/tools/valve-length-fitment-calculator`；尺寸、气嘴、材质、安装说明、计算器、结果、参考矩阵、限制和 FAQ 整体保留；旧路径及查询变体 404，后续不得拆回指南。 |
| 胎压 | 已完成整体迁移 | 新路径 `/tools/dynamic-tire-pressure-calculator`；`calculator` 与 `details`、目录、动态受力、比较演示、说明和反馈整体保留，旧路径全部语言与查询变体 404；后续不得拆回指南。 |
| Schwalbe 选型器 | 本批整页迁移 | 新路径 `/tools/schwalbe-tire-selector`；全部介绍和交互保留，旧路径退役为 404。 |
| 车架间隙 | 保留指南/参考页 | 当前不创建不存在的计算器页面；若以后形成真实交互工具，整体迁移完整页面。 |
| 辐条长度 | 第三次整页迁移已完成 | 新路径 `/tools/spoke-length-calculator`；五步输入、第六步结果、测量说明和目录整体保留，旧路径 404，不再拆分或改回指南分类。 |
| 品牌轮组辐条规格 | 第二批已完成整页迁移 | 新路径 `/tools/brand-wheelset-spoke-specs`；筛选、搜索、规格矩阵、修补包和全部说明保留，旧路径 404；下一批不得再抽取说明或改回指南分类。 |
| 轮组编法拓扑 | 第四次整页迁移已完成 | 新路径 `/tools/spoke-lacing-topology-simulator`；复杂 SVG/几何工作台、工程说明和结果解释整体保留；旧主路径、旧平铺文件路径与历史 resources 路径全部 404，不再拆分或改回指南分类。 |
| 不锈钢辐条位错 | 第六次整页迁移已完成 | 新路径 `/tools/spoke-stress-relief-mechanics`；材料说明、计算器、后台数据目录、结果解释和 FAQ 整体保留，旧路径 404，不拆回指南。 |
| 塔基与飞轮兼容助手 | 第八次整页迁移已完成 | 新路径 `/tools/freehub-cassette-fitment-engine`；三步选型、兼容结果、图片、标准矩阵、动态矩阵、限制、FAQ 和 Schema 整体保留；旧路径及查询变体 404，FAQ 身份保留，不拆回轮组指南。 |
| 胎圈宽度兼容引擎 | 第九次整页迁移已完成 | 新路径 `/tools/tire-rim-width-compatibility-engine`；计算、模拟、DT Swiss 与 Schwalbe 四页签、参数说明、结果、来源、限制、FAQ 和 Schema 整体保留；旧路径及查询变体 404，FAQ 身份保留，不拆成指南页与计算器页。 |

该表是迁移清单的约束来源。任何新增 `/tools/*` 路由、工具大厅卡片或旧路径退役规则，都必须先在本表或对应资产表中更新决策，再进入开发任务。对标记“整体迁移”的资产，任务拆分可以按组件和验收步骤分批，但线上页面不能出现说明留在旧 URL、交互单独迁走的中间产品状态。

---

## 二、 “工具 (Tools) + 指南 (Guides)”双飞轮架构

工具页与指南页不是按组件机械切分的两个容器。工具页承载完整的任务闭环；指南页承载独立的原理、测量或选型内容。对当前的混合页面，先保留一个完整 URL；如果后续判断工具是主任务，就把整页迁移到 `/tools`，而不是把原理留在 `/guides`、把交互抽到另一个 URL。只有真正存在两份独立内容时，才形成“双飞轮”分工：

```
                    ┌────────────────────────────────────────────────────────┐
                    │               站点信息架构双轮驱动体系                  │
                    └───────────────────────────┬────────────────────────────┘
                                                │
             ┌──────────────────────────────────┴──────────────────────────────────┐
             ▼                                                                     ▼
   【 /tools (独立工程工具中心) 】                                       【 /guides (技术指南与白皮书) 】
   • 页面形态：完整的工具交互、参数输入、结果、解释和使用边界       • 页面形态：独立的原理解析、测量步骤、结构科普和图文指南
   • 核心心智：解决“具体参数怎么选？数值是多少？”                        • 核心心智：解决“原理是什么？为什么这样选？”
   • 内部代码：直接复用成熟计算组件与逻辑，并携带完成任务所需说明   • 内部代码：只承载独立指南内容；不为制造工具页而拆走混合页的必要说明
             │                                                                     │
             └──────────────────────► [ 双向上下文锚点高频引流 ] ◄─────────────────┘
               • 工具页面底部（存在独立指南时）："深入了解此方案的工程原理解析 ➔ 查看对应指南"
               • 指南文章内部（已完成整体迁移时）："直接使用工具进行参数匹配 ➔ 启动完整工具"
```

---

## 三、 现有工程工具资产全盘清点与迁移映射表（完整梳理不遗漏）

在此对代码库中所有**已存在或已规划的工具模块**进行地毯式梳理，明确每个工具的**现有代码位置**、**目标独立路由**与**迁移重构策略**。**（不修改内部已有计算逻辑，只做页面解耦与路由迁移）**：

### 3.1 板块 A：轮胎、车圈与气门系统

| 序号 | 工具名称 | 代码库现有文件位置 | 现有访问路径 | 目标独立路由 | 迁移方案说明 |
| :---: | :--- | :--- | :--- | :--- | :--- |
| **1** | **车圈框高与气门嘴/延长嘴穿透匹配器** | `app/components/tireguides/InnerTubeGuide.vue`<br>`app/components/tireguides/InnerTubeValveLengthAndExtenderFitmentGuide.vue`<br>`app/composables/useInnerTubeValveFitmentCalculator.ts` | `/guides/tireguides/choose-inner-tube`（已退役；历史兼容 `/guides/tireguides?tab=choose-inner-tube` 同样 404） | `/tools/valve-length-fitment-calculator` | **第七次整页迁移已完成**：ETRTO/尺寸、气嘴、材质、安装边界、参数说明、匹配交互、结果解释、参考矩阵、数据来源、限制、FAQ 和结构化数据共用一个工具 URL；不保留精简指南页或独立半页计算器。 |
| **2** | **动态前后轮智能胎压与滚阻计算器** | `app/pages/tools/dynamic-tire-pressure-calculator.vue`<br>`app/components/tireguides/tirepressure/*` | `/tools/dynamic-tire-pressure-calculator`（旧 `/guides/tireguides/tire-pressure` 已退役） | `/tools/dynamic-tire-pressure-calculator` | **第五次整页迁移已完成**：`calculator` 与 `details` Tab、计算、图文说明、目录、比较演示和结果共同保留；20 种语言旧路径 404，FAQ 身份和后端算法不变，不拆正文与计算器。 |
| **3** | **Schwalbe 官方外胎选型器** | `app/pages/tools/schwalbe-tire-selector.vue`<br>`app/components/tireguides/schwalbe/*` | `/guides/tireguides/schwalbe-tire-selector` | `/tools/schwalbe-tire-selector` | **本批整页迁移**：介绍、搜索、筛选和目录结果全部保留。现行路径为 `/tools/schwalbe-tire-selector`，表内旧访问路径仅用于退役 404 核对。 |
| **4** | **外胎实测膨胀率与车架安全间隙校验器** | `app/components/tireguides/TireFrameClearanceGuide.vue` | `/guides/tireguides/tire-frame-clearance` | **暂不新建**（候选：`/tools/tire-frame-clearance-checker`） | **先保留参考指南**：当前主要是测量说明、参考表和图片，未确认存在独立交互计算器。不得为了匹配工具目录创建不存在的计算页；若以后补齐真实计算能力，再按工具定位和整页迁移规则评估。 |
| **4.1** | **胎圈宽度计算与模拟工作台** | `app/pages/tools/tire-rim-width-compatibility-engine.vue`<br>`app/components/tireguides/HowToChooseGuide.vue` 与 Go v11 胎圈工程引擎 | `/guides/tireguides/tire-rim-width-compatibility`（已退役） | `/tools/tire-rim-width-compatibility-engine` | **第九次整页迁移已完成**：计算、模拟、DT Swiss、Schwalbe 四页签、输入说明、分层结果、来源、限制、FAQ、Schema 和反馈整体保留；正常范围与最佳算法的 16 组冲突原样保留，旧路径及查询变体 404，不拆成指南页与计算器页。 |
| **5** | **Hookless 车圈与轮胎物理安全校验器** | 关联已有 Hookless 适配契约与数据模型 | 规划中（当前仓库未发现独立页面） | `/tools/hookless-compatibility-checker` | **保留规划项**：在出现可迁移页面和验收用例前，不加入可点击工具入口、不加入 sitemap，也不建立旧路径迁移规则。 |

---

### 3.2 板块 B：轮组、辐条与力学工程

| 序号 | 工具名称 | 代码库现有文件位置 | 现有访问路径 | 目标独立路由 | 迁移方案说明 |
| :---: | :--- | :--- | :--- | :--- | :--- |
| **6** | **多步向导式辐条长度精密计算器** | `app/pages/tools/spoke-length-calculator.vue`<br>`app/components/Spoke*.vue` (五步输入 + 第六步结果)<br>`app/composables/useSpokeCalculator*.ts` | `/tools/spoke-length-calculator`（旧 `/guides/spokeguides/spoke-length-calculator` 已退役） | `/tools/spoke-length-calculator` | **已完成整页迁移**：参数输入、测量说明、目录和结果构成完整任务；算法、交互和数据复用，旧路径全部语言与查询变体为 404。 |
| **7** | **高端大牌轮组官方出厂辐条规格速查** | `app/pages/tools/brand-wheelset-spoke-specs.vue`<br>`app/data/brand-wheelset-spoke-specs/*` | `/tools/brand-wheelset-spoke-specs` | `/tools/brand-wheelset-spoke-specs` | **第二批已完成整页迁移**：品牌筛选、搜索、分页、完整规格矩阵、修补包商品和说明均保留；公开访问方式不变。旧 `/guides/spokeguides/brand-wheelset-spoke-specs` 已退役为 404，不另造重复指南。 |
| **8** | **轮组编法拓扑与双法兰几何预览器** | `app/pages/tools/spoke-lacing-topology-simulator.vue` | `/tools/spoke-lacing-topology-simulator`（旧指南路径已退役） | `/tools/spoke-lacing-topology-simulator` | **已完成整页迁移**：复杂 SVG/几何工作台是页面核心交互。工程说明、全部仿真控件、结果解释和几何限制整体保留；移除旧指南特殊路由注册和分类入口。仅用于原有拓扑/几何预览，不以迁移或“仿真器”名称新增法兰物理干涉、安全或长度计算保证。 |
| **9** | **不锈钢辐条微观位错力学评估器** | `app/pages/tools/spoke-stress-relief-mechanics.vue`<br>`app/composables/useStainlessSteelSpokeDislocationMechanicsCalculation.ts` | `/tools/spoke-stress-relief-mechanics`（旧 `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics` 已退役） | `/tools/spoke-stress-relief-mechanics` | **第六次整页迁移已完成**：材料力学说明、计算器、后台数据目录、结果解释、FAQ 和结构化数据共同保留；SPOKEGUIDES 分类页删除，不另建说明页和计算器页两个近重复 URL。 |
| **10** | **车圈气压径向压缩与张力暴跌预测器** | 关联现有力学看板规范与工程数据 | 规划中（当前仓库未发现独立页面） | `/tools/rim-compression-tension-drop-predictor` | **保留规划项**：未有页面、数据契约和回归用例前，不加入可点击工具入口、不加入 sitemap。 |

---

### 3.3 板块 C：传动系统与整车兼容性

| 序号 | 工具名称 | 代码库现有文件位置 | 现有访问路径 | 目标独立路由 | 迁移方案说明 |
| :---: | :--- | :--- | :--- | :--- | :--- |
| **11** | **塔基与飞轮兼容助手** | `app/components/WheelsetChooseFreehubSection.vue`、`app/components/FreehubGroupsetHelper.vue`，复用 Go 兼容性接口 | `/guides/wheelset-buyers/choose-freehub`（已退役） | `/tools/freehub-cassette-fitment-engine` | **第八次整页迁移已完成**：三步选型说明、输入、兼容结果、图片、标准矩阵、动态矩阵、HowTo、FAQ、Schema、限制和反馈共用一个工具 URL；轮组指南移除内嵌小计算器，旧路径及 20 种语言/查询变体 404，不拆成说明页与计算器页。 |

---

## 四、 导航体系与菜单布局改造方案

为了让用户能快速找到已经确认适合独立使用的工程工具，前台导航可以进行轻量而明确的调整。导航分类用于发现和组织内容；进入 `/tools` 的页面必须是完整可用的任务页面，不能只是从指南页抽出的计算器面板。尚未整体迁移的混合页继续留在原指南导航中，不提前占用工具入口。

### 1. 顶部 Header 导航栏与手机、平板工具入口
* **主导航栏新增顶级项**：
  在 `SHOP`、`GUIDES`、`ABOUT` 之间，新增 **`TOOLS`（工程工具）** 顶级导航项：
  ```
  [ LOGO ]    SHOP    TOOLS (新)    GUIDES    ABOUT    [ 搜索 / 语言 / 购物车 ]
  ```
* **下拉菜单（Mega Menu / Dropdown）三列布局**：
  点击 `TOOLS` 展开现有 HeaderMegaMenu；未来工具数量足够后再按以下分类组织，当前保留九个具体工具卡片（大厅是标题入口，不作为同级工具卡片）：Schwalbe、品牌规格、辐条长度、轮组拓扑、动态胎压、内胎气门嘴匹配、塔基与飞轮兼容、不锈钢应力释放力学、胎圈宽度兼容引擎：
  * **第一列：轮胎与气门系统**（列出已完成整体迁移的胎压、Schwalbe、内胎气门嘴匹配和胎圈宽度兼容工具；车架间隙继续留在指南导航）
  * **第二列：轮组与辐条工程**（列出已完成整体迁移的辐条长度、品牌规格、轮组拓扑与不锈钢应力释放力学工具）
  * **第三列：传动与通用系统**（后续完成整页迁移的塔基助手；规划中的 Hookless 工具不展示卡片）

**本批入口与弹层约束（后续开发必须沿用）**：

* 手机和平板在顶部商品图标右侧放置扳手图标按钮，使用工具分组的本地化名称作为可访问标签；主导航不再重复占用一个工具文字入口。
* 扳手、商品及其他导航都调用同一套 HeaderMegaMenu 开关状态，保持点击展开、再次点击关闭、切换分组、Escape 和外部点击关闭的现有行为；不创建另一套工具抽屉。
* 电脑、手机和平板均以完整页头底部为弹层起点，沿用电脑端 1px 边框重叠规则；不能单独锚定商品图标、顶部第一行或移动主导航，造成各入口位置不同。
* 弹层外框和内容面板都按实际内容决定高度，不设固定高度、视口填满高度或 bottom 锚点；仅超过现有可视高度上限时在内容区滚动，保留收起按钮和安全区。
* 页脚顶部“工具”分组标题链接 `/tools` 大厅，下方只列具体已迁移工具，不把大厅再列成与工具同级的子链接；手机端标题仍链接大厅，右侧独立箭头控制子工具列表折叠，平板和电脑保留现有列表布局。
### 2. `/tools` 大厅入口页面 (Tools Hub / Dashboard)
* **路由**: `/tools`
* **页面形态**: 工业仪表盘风格的工具导航大厅；
* **内容构成**:
  * 顶部：清晰的工具集介绍；当前九项具体工具，暂不增加大厅搜索框；
  * 主体：只展示已整体迁移且可用的工具；当前为 Schwalbe、品牌规格、辐条长度、轮组拓扑、动态胎压、内胎气门嘴匹配、塔基与飞轮兼容、不锈钢应力释放力学和胎圈宽度兼容引擎，后续数量足够再按三类任务分组。大厅作为工具分组标题入口，不在卡片列表中重复出现。暂未迁移的页面继续从原导航访问，规划项不展示卡片、不提供启动链接或可索引工具页；
  * 工具大厅以导航和任务选择为主，可保留简短说明；不得为了保持目录数量而复制指南正文、制造第二份计算器页面或把半个页面标记为工具。

### 3. 面包屑对齐
所有独立工具页面的面包屑统一规范为：
`首页 ➔ 工程工具 (Tools) ➔ [具体工具名称]`

---

## 五、 整页迁移与旧路径退役策略

用户最新决定优先于此前的 301 方案：本轮及后续工具迁移默认不做旧地址兼容。完成整页迁移后，新工具 URL 是唯一公开主路径；旧路径返回 404，不保留别名、重复页面、跳转页面或迁移重定向。尚未完成整页迁移的页面继续使用现有 URL。此规则仅针对本工具迁移，不撤销网站其他既有重定向。

### 1. 完整页面迁移与退役映射

以下是整页迁移目标，不是重定向配置；只有完成内容范围核对、输入输出回归、canonical/hreflang 和 sitemap 验收后，才退役对应旧路径：
* `/guides/tireguides/schwalbe-tire-selector` ➔ `/tools/schwalbe-tire-selector`（本批完成，旧路径 404）
* `/guides/spokeguides/spoke-length-calculator` ➔ `/tools/spoke-length-calculator`（第三次完成，旧路径 404）
* `/guides/spokeguides/brand-wheelset-spoke-specs` ➔ `/tools/brand-wheelset-spoke-specs`（第二批完成，旧路径 404）
* `/guides/tireguides/tire-pressure` ➔ `/tools/dynamic-tire-pressure-calculator`（第五次整页迁移完成；旧路径全部语言与查询变体 404）
* `/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference` ➔ `/tools/spoke-lacing-topology-simulator`（第四次完成；该主路径、旧平铺 `/guides/wheelset-spoke-lacing-topology-and-geometry-reference`、历史 `/resources/wheelset-spoke-lacing-topology-and-geometry-reference` 均退役为 404）
* `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics` ➔ `/tools/spoke-stress-relief-mechanics`
* `/guides/spokeguides`：SPOKEGUIDES 已无剩余子页，分类路由随本次迁移删除并返回 404，不设置空分类页、别名或重定向。
* `/guides/tireguides/choose-inner-tube` ➔ `/tools/valve-length-fitment-calculator`（已完成；旧路径及查询变体 404）
* `/guides/wheelset-buyers/choose-freehub` ➔ `/tools/freehub-cassette-fitment-engine`（第八次整页迁移完成；旧路径及 20 种语言、`tab`、品牌/规格查询变体 404）

退役路径的各语言前缀和查询参数变体均应返回 404，且没有迁移 Location。退役路径必须显式关闭继承的 HTML cache 并声明 `no-store`，避免删除路由后仍命中旧缓存。同步更新内部链接、页脚、导航、canonical、hreflang、路由台账及 FAQ 公开路径；动态 sitemap 即使收到后台未同步的旧路径，也必须排除。FAQ key、page_id 等内部内容身份可以继续保留，保留问答数据不等于提供旧 URL 兼容。

SEO 影响：404 不会通过永久重定向承接旧 URL 的外链信号，旧收藏和外部链接将失效；新路径依靠完整内容、内部链接和 sitemap 重新被发现与评估。不得因为通用 SEO 迁移建议而自动恢复本任务已取消的 301。

### 2. 图文混合页面与内嵌组件的整页策略

以下页面在迁移完成前保留完整内容，不能把其中的说明和交互拆成两个半页；迁移时沿用整页策略：
* `/tools/dynamic-tire-pressure-calculator`（原 `/guides/tireguides/tire-pressure`）：已保留现有 `calculator` 与 `details` Tab 的组合页面；说明、交互、结果解释、来源、限制和反馈在同一工具 URL 内。
* `/guides/tireguides/choose-inner-tube`：已退役。内胎尺寸、气嘴、材质、安装说明、交互、结果解释、参考矩阵、限制和 FAQ 已整体放入 `/tools/valve-length-fitment-calculator`，没有留下说明页与计算器页的割裂组合。
* `/tools/spoke-lacing-topology-simulator`：已将原轮组指南拓扑页完整迁入；工程说明和仿真工作台共同保留，不再创建重复指南页或恢复旧路径。
* `/tools/spoke-stress-relief-mechanics`（原 `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics`）：材料力学正文、计算器、数据目录、结果解释和 FAQ 已整体保留在同一个工具 URL。
* `/tools/freehub-cassette-fitment-engine`（原 `/guides/wheelset-buyers/choose-freehub`）：塔基与飞轮的三步选型说明、输入、动态/静态兼容结果、实物图片、标准矩阵、动态矩阵、HowTo、FAQ、Schema、限制和反馈整体保留在同一个工具 URL；轮组指南只保留其余独立指南内容，不恢复内嵌小计算器。
* `/guides/tireguides/tire-frame-clearance`：保留测量说明和参考表；当前未发现独立交互计算器，不创建名义上的工具页。

对于原先内嵌在 `/guides/tireguides/choose-inner-tube`（以及兼容的 `/guides/tireguides?tab=choose-inner-tube`）内部的气门嘴计算器：
* **当前策略**：已整体迁移到 `/tools/valve-length-fitment-calculator`，说明、参数解释、结果上下文、参考矩阵、数据来源和限制与交互共用一个 URL。
* **后续约束**：不得把说明留回 `/guides`、只保留一个计算器面板在 `/tools`，也不得为了恢复“指南”标签重新建立旧分类或兼容路由。
* **迁移后的旧 URL**：完整页面迁移、回归验收及 canonical/hreflang 完成后，旧路径退役为 404；不能因退役旧路径而省略必要的说明和结果上下文。

---

## 六、 独立工具页的 SEO/GEO 外壳与 Schema 规范

本项改造仅针对页面的**外壳元数据与结构化标签**进行对齐，不涉及内部计算逻辑：

### 1. Schema.org 结构化数据
按真实页面内容选择结构化数据。工具大厅不自动声明 WebApplication；迁移不要求所有工具统一改为应用 Schema。已有 Article、Dataset、FAQPage 或 BreadcrumbList 必须与可见内容一致。只有确实能准确描述可用交互应用时，才按需添加 WebApplication，例如：
```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "WebApplication",
  "@id": "/tools/valve-length-fitment-calculator#app",
  "name": "公路车内胎车圈框高与法嘴/延长嘴穿透力学装配匹配引擎",
  "applicationCategory": "EngineeringApplication",
  "operatingSystem": "All",
  "browserRequirements": "Requires JavaScript for interactive calculation"
}
</script>
```

### 2. 页面标题与多语言规范
* 页面标题准确表达实际工具任务并本地化；本批保留 Schwalbe 既有标题，不通过统一追加“工程工具中心”改写全部页面 SEO。
* 自动继承已有的 `useLocalePath` 与 `@nuxtjs/i18n` 多语言配置，生成标准的对称 hreflang。

### 3. SEO 拆分与索引规则

* 搜索引擎主要依据查询意图、页面内容和任务完成度判断相关性，不会因为 URL 位于 `/tools` 或 `/guides` 就自动给予不同排名。目录名称用于组织和导航，不能替代页面价值。
* `/tools` 对人和检索系统都应表达一个清晰承诺：页面打开后可以直接交互、查表或得到结果。工具页的简介、参数说明、结果解释、数据来源和限制说明应与交互放在同一 URL；“工具”标签本身不是排名保证，但比把可操作任务藏在指南 Tab 中更准确地表达任务意图。
* 当前混合页如果以后迁移，优先执行整页迁移：保留完整说明和交互，只改变主路由和页面归类。不要为了得到一个 `/tools` URL 而把同一页面拆成“指南正文 URL + 计算器 URL”。
* 两个 URL 只有在确实存在两份独立内容、用户任务明显不同、正文与说明可以各自成立、标题和摘要不重复，并且各自能通过独立验收时，才同时允许索引。相同计算器可以复用同一个 composable、API 或领域组件，但不能在两个 URL 复制完整正文和完整交互。
* 近重复页面只保留一个主 URL 和 canonical；`noindex`、canonical、合并页面或保留混合页的选择要基于实际内容，而不是为了填满工具目录。
* 本任务默认采用整页迁移后旧路径 404 的策略，不建立迁移 301、路由别名或兼容页面；后续实施不得把通用迁移建议当作恢复兼容的授权。未完成整页迁移的页面继续使用当前 URL。
* `WebApplication` 等 Schema 只描述真实存在、可用且可访问的页面，不是拆分理由，也不保证排名。规划中的工具不得伪装成已上线应用，不得进入 sitemap。

---

## 七、 研发落地分批拆解检查清单

为保证开发工作稳健有序、不遗漏任何细节，按以下四个批次逐步推进。每一批都必须遵守“算法零改动、先验证路由再迁移”的边界：

### 迁移前置核对（必须先完成）
- [ ] 0. 对照 `public/storefront-route-manifest.json`、`definePageMeta` 和实际 HTTP 响应，登记每个旧路径的 canonical、语言前缀、sitemap 状态和是否含指南正文。
- [ ] 0.1 将工具标记为“已实现可迁移”“已实现但与指南混合”或“规划中”；规划项不得生成可点击的工具页链接。
- [ ] 0.2 为每个已实现工具保存迁移前的代表性输入、输出和交互快照；迁移提交不得修改计算 composable、后端服务、领域模型或状态机。

### 第一批：搭建 `/tools` 骨架与大厅页 (Foundation)
- [x] 1. 新建 `app/pages/tools/interactive-tool-catalog-navigation-overview.vue`，公开路径 `/tools`；当前展示已迁移的 Schwalbe、品牌规格、辐条长度和轮组拓扑工具，后续按实际数量再分组；
- [x] 2. 电脑 Header 保留 `TOOLS` 顶级入口；手机、平板在商品图标右侧提供扳手按钮，复用同一点击弹层；页脚标题链接大厅，下面只列已迁移的具体工具，不重复放同级大厅项；
- [x] 3. 验证 `/tools` 大厅页面的 i18n 多语言翻译文本与路由跳转畅通。

### 第二批：迁移已有独立页面工具 (Page Migration)
- [x] 4. 迁移 Schwalbe 外胎选型器至 `app/pages/tools/schwalbe-tire-selector.vue`，保留完成选型所需的介绍和目录说明；
- [x] 5. 迁移大牌辐条速查矩阵至 `app/pages/tools/brand-wheelset-spoke-specs.vue`，保留完整规格、修补包、说明和公开访问方式；
- [x] 6. 迁移辐条长度向导至 `app/pages/tools/spoke-length-calculator.vue`，保持五步输入、第六步结果、测量说明和目录；
- [x] 7. 以上页面均按完整页面迁移，不删减完成任务所需的介绍、数据来源或限制说明；
- [x] 8. 整页验收通过后退役对应旧路径，验证全部语言与查询参数变体为 404、没有迁移 Location、没有旧别名或兼容页；Schwalbe、品牌规格和辐条长度均已完成。

### 第三批：整体迁移工具优先的混合页面 (Whole-page Migration)
- [x] 9. 对需要迁移的混合页面先确定完整页面边界，列出标题、说明、参数解释、结果解释、数据来源和限制说明；不先抽走单独计算器；胎压与胎圈宽度迁移前基线均已保存。
- [x] 10. 将胎压页面整体迁移至 `app/pages/tools/dynamic-tire-pressure-calculator.vue`，保留 `calculator` 和 `details` 全部内容；旧路径退役为 404，FAQ 身份和后端算法保持不变。
- [x] 11. 将内胎气门嘴匹配页面整体迁移至 `app/pages/tools/valve-length-fitment-calculator.vue`，保留完整上下文、说明和交互，不制作精简指南页 + 独立半页工具；
- [x] 12. 将轮组编法拓扑页面整体迁移至 `app/pages/tools/spoke-lacing-topology-simulator.vue`，保留工程说明、仿真控件和结果解释；
- [x] 13. 将不锈钢辐条位错页面整体迁移至 `app/pages/tools/spoke-stress-relief-mechanics.vue`，保留材料说明、计算器、数据目录和结果解释；
- [x] 14. 整体迁移验收通过后，再退役对应旧 URL，返回 404；轮组拓扑、胎压、不锈钢位错、内胎气门嘴、塔基与飞轮兼容助手和胎圈宽度兼容引擎均已完成，旧路径及查询变体已逐项验证。

### 第四批：双向飞轮与收尾验证 (Verification)
- [ ] 15. 在每个独立工具页确认页面自身已包含完成任务所需的说明、结果解释和限制；只有存在独立指南内容时才添加“前往阅读相关指南”链接；
- [ ] 16. 检查动态 Sitemap 只包含已整体迁移、可访问且已确认需要独立 URL 的 `/tools/*` 路由；规划项和仍在旧指南 URL 的页面不得进入工具 sitemap；
- [ ] 17. 逐个整体迁移候选页面进行功能与内容回归，确认**算法、计算数值、参数联动、主任务、页面说明、结果解释和限制均与迁移前一致**；同时验证已迁移旧 URL 的 404 和未迁移页面的完整保留策略；
- [x] 18. 本轮已更新“本轮评估的决策基线”和资产迁移表，记录塔基与胎圈宽度工具的整页范围、重复风险、canonical、sitemap、FAQ 身份和旧路径退役决策；后续新增工具仍必须先补齐同类记录，防止把完整迁移改成只抽走计算器。

---
*文档编制完成。本规范为纯信息架构重组与路由迁移实施指南，所有开发工作严禁篡改现有算法逻辑。*

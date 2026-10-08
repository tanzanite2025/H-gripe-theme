# 🛠️ 独立工程工具体系 (/tools) 全景架构与迁移实施规范
## Architecture & Migration Specification for Engineering Tools Suite

> **文档版本**: `v1.2-Engineering`
> **文档位置**: `docs/design/tools-category-architecture-and-engineering-suite-specification.md`
> **适用范围**: 针对现有代码库中已实现的所有计算器、适配器、选型矩阵及仿真看板，判定其应**保留完整混合页面还是整体迁移为工具页面**，并规范后续导航重组。
>
> 🛑 **【最高核心原则：算法零改动，纯架构重组】**
> **本项目所有核心力学模型、材料物性参数、物理计算引擎与状态机逻辑已在代码库中 100% 完整实现且实机验证正确。**
> **本文档严禁编写、推导或重新设计任何数学公式与物理算法，绝不增加任何额外计算需求，杜绝技术跑偏！**
> **本文档唯一目标是：全面盘点现有分散在 `/guides` 中的各块工具资产，先依据搜索意图、页面主任务、内容独立性与重复风险判定应保留完整混合页面还是整体迁移为工具页面，再理清路由、菜单、面包屑与迁移映射，确保平稳落地、不落下一块。工具页必须是可以独立完成任务的完整页面；如果混合页面以后迁移到 `/tools`，说明、交互、结果解释、数据来源和使用边界一并迁移，不拆成一个地方放说明、另一个地方放计算器。**

> **路由核对口径**：下文“现有访问路径”以 `nuxt-i18n/public/storefront-route-manifest.json` 与页面 `definePageMeta` 的实际结果为准。`?tab=...` 只是兼容性查询参数，不作为独立规范路由。`/tools` 与 `/guides` 是信息架构标签，但 `/tools` 对用户有明确承诺：打开后应能直接交互、查表或得到结果，并能在同一页面理解输入、输出和限制。迁移前必须确认页面的主任务、内容完整性和重复风险；混合页面只有在整体内容准备好后才能迁移和 301。

---

## 目录
1. [一、 为什么要做这次架构重组（纯信息架构与体验升级）](#一-为什么要做这次架构重组纯信息架构与体验升级)
2. [二、 “工具 (Tools) + 指南 (Guides)”双飞轮架构](#二-工具-tools--指南-guides双飞轮架构)
3. [三、 现有工程工具资产全盘清点与迁移映射表（完整梳理不遗漏）](#三-现有工程工具资产全盘清点与迁移映射表完整梳理不遗漏)
   - [3.1 板块 A：轮胎、车圈与气门系统](#31-板块-a轮胎车圈与气门系统)
   - [3.2 板块 B：轮组、辐条与力学工程](#32-板块-b轮组辐条与力学工程)
   - [3.3 板块 C：传动系统与整车兼容性](#33-板块-c传动系统与整车兼容性)
4. [四、 导航体系与菜单布局改造方案](#四-导航体系与菜单布局改造方案)
5. [五、 原 /guides 兼容与 301 重定向策略（无损过渡）](#五-原-guides-兼容与-301-重定向策略无损过渡)
6. [六、 独立工具页的 SEO/GEO 外壳与 Schema 规范](#六-独立工具页的-seogeo-外壳与-schema-规范)
7. [七、 研发落地分批拆解检查清单](#七-研发落地分批拆解检查清单)

---

## 一、 为什么要做这次架构重组（纯信息架构与体验升级）

目前项目里的各个计算器和工程工具**在算法与功能层面已经做得非常扎实、完全正确**。然而在**信息架构（Information Architecture）**上，这些工具大多被塞在 `/guides`（指南）的深层 Tab、子目录或折叠面板中：

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
| 重复风险 | 迁移是否会同时留下内容高度相同的旧、新 URL？ | 整页迁移后只保留一个可索引主 URL；旧页面只有在整体等价时才 301。 |
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
| 内胎气门嘴匹配 | 工具优先；整体页面迁移候选 | 当前先保持完整页面；迁移时将选型上下文、说明、交互、结果和限制整体放入 `/tools`，不另留一张精简指南卡片承接原内容。 |
| 胎压 | 工具优先；整体页面迁移候选 | `calculator` 与 `details` 共同构成可操作的胎压任务；整体迁移全部 Tab 内容和说明，不拆正文和计算器。 |
| Schwalbe 选型器 | 整体迁移为工具页 | 保留必要介绍、搜索筛选、目录结果和数据说明；新旧页等价验收后再 301。 |
| 车架间隙 | 保留指南/参考页 | 当前不创建不存在的计算器页面；若以后形成真实交互工具，整体迁移完整页面。 |
| 辐条长度 | 独立工具优先 | 迁移完整五步向导、输入说明和结果面板，并做输入/输出回归。 |
| 品牌轮组辐条规格 | 整体迁移为查表工具 | 保留筛选、搜索、规格矩阵、数据来源和使用说明；等价验收后再 301。 |
| 轮组编法拓扑 | 工具优先；整体页面迁移候选 | 复杂 SVG/几何仿真是页面的核心交互；迁移时连同工程说明和结果解释整体迁移，不抽走单独的仿真器半页。 |
| 不锈钢辐条位错 | 工具优先；整体页面迁移候选 | 保留材料说明、计算器、后台数据目录和结果解释；整体迁移，不把原理留在指南页。 |

该表是迁移清单的约束来源。任何新增 `/tools/*` 路由、工具大厅卡片或 301 规则，都必须先在本表或对应资产表中更新决策，再进入开发任务。对标记“整体迁移”的资产，任务拆分可以按组件和验收步骤分批，但线上页面不能出现说明留在旧 URL、交互单独迁走的中间产品状态。

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
| **1** | **车圈框高与气门嘴/延长嘴穿透匹配器** | `app/components/tireguides/InnerTubeValveLengthAndExtenderFitmentGuide.vue`<br>`app/composables/useInnerTubeValveFitmentCalculator.ts` | `/guides/tireguides/choose-inner-tube`（兼容 `/guides/tireguides?tab=choose-inner-tube`） | `/tools/valve-length-fitment-calculator` | **完整工具页迁移候选**：将内胎选型上下文、参数说明、匹配交互、结果解释和使用限制整体迁移到工具页；若当前路由包含更广的内胎指南内容，先核实边界，不能只搬计算器。 |
| **2** | **动态前后轮智能胎压与滚阻计算器** | `app/pages/guides/tire-pressure.vue`<br>`app/components/tireguides/tirepressure/*` | `/guides/tireguides/tire-pressure` | `/tools/dynamic-tire-pressure-calculator` | **整体迁移为工具页**：当前页面有 `calculator` 与 `details` Tab，计算、图文说明和结果共同完成胎压任务。迁移时带上全部 Tab 和说明，不拆正文与计算器；新旧页面整体等价验收后再 301。 |
| **3** | **Schwalbe 官方外胎选型与周长适配器** | `app/pages/guides/schwalbe-tire-selector.vue`<br>`app/components/tireguides/schwalbe/*` | `/guides/tireguides/schwalbe-tire-selector` | `/tools/schwalbe-tire-selector` | **整体迁移为工具页**：页面已有介绍、搜索、筛选和目录结果，保留完成选型所需的介绍与数据说明。新旧页面内容和意图验收等价后，才配置 301。 |
| **4** | **外胎实测膨胀率与车架安全间隙校验器** | `app/components/tireguides/TireFrameClearanceGuide.vue` | `/guides/tireguides/tire-frame-clearance` | **暂不新建**（候选：`/tools/tire-frame-clearance-checker`） | **先保留参考指南**：当前主要是测量说明、参考表和图片，未确认存在独立交互计算器。不得为了匹配工具目录创建不存在的计算页；若以后补齐真实计算能力，再按工具定位和整页迁移规则评估。 |
| **5** | **Hookless 车圈与轮胎物理安全校验器** | 关联已有 Hookless 适配契约与数据模型 | 规划中（当前仓库未发现独立页面） | `/tools/hookless-compatibility-checker` | **保留规划项**：在出现可迁移页面和验收用例前，不加入可点击工具入口、不加入 sitemap，也不配置 301。 |

---

### 3.2 板块 B：轮组、辐条与力学工程

| 序号 | 工具名称 | 代码库现有文件位置 | 现有访问路径 | 目标独立路由 | 迁移方案说明 |
| :---: | :--- | :--- | :--- | :--- | :--- |
| **6** | **多步向导式辐条长度精密计算器** | `app/pages/guides/spokeguides/spoke-length-calculator.vue`<br>`app/components/Spoke*.vue` (向导 5 步)<br>`app/composables/useSpokeCalculator*.ts` | `/guides/spokeguides/spoke-length-calculator` | `/tools/spoke-length-calculator` | **整体迁移为工具页**：五步向导、参数输入和结果面板构成完整计算任务。平移至 `pages/tools/` 时保持向导、使用说明和计算代码原样复用；新页验收通过后，旧页才可按等价性配置 301。 |
| **7** | **高端大牌轮组官方出厂辐条规格速查** | `app/pages/guides/spokeguides/brand-wheelset-spoke-specs.vue`<br>`app/data/brand-wheelset-spoke-specs/*` | `/guides/spokeguides/brand-wheelset-spoke-specs` | `/tools/brand-wheelset-spoke-specs` | **整体迁移为查表工具**：品牌筛选、搜索、分页和规格矩阵本身构成独立查表任务；保留必要的数据来源与使用说明，不另造重复指南。新旧页面等价后再配置 301。 |
| **8** | **轮组编法拓扑与法兰干涉仿真器** | `app/pages/guides/wheelset-spoke-lacing-topology-and-geometry-reference.vue` | `/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference` | `/tools/spoke-lacing-topology-simulator` | **整体迁移为工具页**：复杂 SVG/几何工作台是页面核心交互。整体迁移工程说明、仿真控件、结果解释和几何限制，不抽走单独的仿真器半页。 |
| **9** | **不锈钢辐条微观位错力学评估器** | `app/pages/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics.vue`<br>`app/composables/useStainlessSteelSpokeDislocationMechanicsCalculation.ts` | `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics` | `/tools/spoke-stress-relief-mechanics` | **整体迁移为工具页**：材料力学说明、计算器、后台数据目录和结果解释共同让工程评估可用；整体迁移，不另建说明页和计算器页两个近重复 URL。 |
| **10** | **车圈气压径向压缩与张力暴跌预测器** | 关联现有力学看板规范与工程数据 | 规划中（当前仓库未发现独立页面） | `/tools/rim-compression-tension-drop-predictor` | **保留规划项**：未有页面、数据契约和回归用例前，不加入可点击工具入口、不加入 sitemap。 |

---

### 3.3 板块 C：传动系统与整车兼容性

| 序号 | 工具名称 | 代码库现有文件位置 | 现有访问路径 | 目标独立路由 | 迁移方案说明 |
| :---: | :--- | :--- | :--- | :--- | :--- |
| **11** | **塔基花键与飞轮变速兼容性求解器** | 关联现有传动系统适配规范 | 规划中（当前仓库未发现独立页面） | `/tools/freehub-cassette-fitment-engine` | **保留规划项**：未有页面、数据契约和回归用例前，不加入可点击工具入口、不加入 sitemap。 |

---

## 四、 导航体系与菜单布局改造方案

为了让用户能快速找到已经确认适合独立使用的工程工具，前台导航可以进行轻量而明确的调整。导航分类用于发现和组织内容；进入 `/tools` 的页面必须是完整可用的任务页面，不能只是从指南页抽出的计算器面板。尚未整体迁移的混合页继续留在原指南导航中，不提前占用工具入口。

### 1. 顶部 Header 导航栏与移动端 Drawer 抽屉
* **主导航栏新增顶级项**：
  在 `SHOP`、`GUIDES`、`ABOUT` 之间，新增 **`TOOLS`（工程工具）** 顶级导航项：
  ```
  [ LOGO ]    SHOP    TOOLS (新)    GUIDES    ABOUT    [ 搜索 / 语言 / 购物车 ]
  ```
* **下拉菜单（Mega Menu / Dropdown）三列布局**：
  点击/悬浮 `TOOLS` 展开清晰的分类菜单：
  * **第一列：轮胎与气门系统**（只列出已完成整体迁移的气门嘴匹配、胎压和 Schwalbe 工具；车架间隙参考留在指南导航）
  * **第二列：轮组与辐条工程**（只列出已完成整体迁移的辐条长度和品牌规格工具；拓扑与微观位错在整页迁移前留在指南导航）
  * **第三列：传动与通用系统**（塔基飞轮兼容求解器、Hookless 安全校验器；规划项显示为置灰状态，不提供启动链接）

### 2. `/tools` 大厅入口页面 (Tools Hub / Dashboard)
* **路由**: `/tools`
* **页面形态**: 工业仪表盘风格的工具导航大厅；
* **内容构成**:
  * 顶部：清晰的工具集介绍与快速搜索框；
  * 主体：按 3 大板块陈列已整体迁移且可用的工具；暂未迁移的混合页或参考页不进入工具卡片，继续从指南导航访问；另外 3 个规划项可以用明确的“规划中”状态卡展示，但不得提供【启动工具 ➔】按钮、可索引工具页或虚假计算结果；
  * 工具大厅以导航和任务选择为主，可保留简短说明；不得为了保持目录数量而复制指南正文、制造第二份计算器页面或把半个页面标记为工具。

### 3. 面包屑对齐
所有独立工具页面的面包屑统一规范为：
`首页 ➔ 工程工具 (Tools) ➔ [具体工具名称]`

---

## 五、 原 /guides 兼容与 301 重定向策略（无损过渡）

为了保障现有已被收录的 URL 权重不丢失，并兼顾老用户的使用习惯，采取以下过渡策略：

### 1. 整体迁移工具页面的服务端 301 永久重定向映射表

只对**完整页面整体迁移**后，目标工具页已经承接原页面全部用户任务和必要内容，且新旧页面在搜索意图和主要内容上基本等价的页面配置整页 301。以下是**完成整页迁移并验收后可执行的候选映射**，不是立即生效的重定向清单：
* `/guides/tireguides/schwalbe-tire-selector` ➔ `301` ➔ `/tools/schwalbe-tire-selector`
* `/guides/spokeguides/spoke-length-calculator` ➔ `301` ➔ `/tools/spoke-length-calculator`
* `/guides/spokeguides/brand-wheelset-spoke-specs` ➔ `301` ➔ `/tools/brand-wheelset-spoke-specs`
* `/guides/tireguides/tire-pressure` ➔ `301` ➔ `/tools/dynamic-tire-pressure-calculator`
* `/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference` ➔ `301` ➔ `/tools/spoke-lacing-topology-simulator`
* `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics` ➔ `301` ➔ `/tools/spoke-stress-relief-mechanics`
* `/guides/tireguides/choose-inner-tube` ➔ `301` ➔ `/tools/valve-length-fitment-calculator`（须先核实旧路径承载范围与工具页范围完全等价）

上述候选不代表现在立即执行 301。每条映射都必须先完成完整页面迁移、内容范围核对、输入输出回归、canonical/hreflang 和 sitemap 验收；不能先抽走计算器，再用 301 掩盖旧页面内容缺失。

重定向必须在服务端返回 `301`，保留语言前缀和必要的查询参数，且目标页的 canonical、hreflang 和 sitemap 只指向新的 `/tools/*` 路径。旧路径上线前要用真实路由清单和 HTTP 集成测试逐条核对。

### 2. 图文混合页面与内嵌组件的兼容策略

以下页面在迁移完成前保留完整内容，不能把其中的说明和交互拆成两个半页；迁移时沿用整页策略：
* `/guides/tireguides/tire-pressure`：保留现有 `calculator` 与 `details` Tab 的组合页面；迁移时将整个页面放入 `/tools/dynamic-tire-pressure-calculator`。
* `/guides/tireguides/choose-inner-tube`：保留内胎选型上下文和气门嘴匹配交互；迁移时将页面范围内的说明、交互和结果解释整体放入 `/tools/valve-length-fitment-calculator`，前提是两边范围完全等价。
* `/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference`：保留工程说明和仿真工作台；迁移时将完整页面放入 `/tools/spoke-lacing-topology-simulator`。
* `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics`：保留材料力学正文、计算器和结果解释；迁移时将完整页面放入 `/tools/spoke-stress-relief-mechanics`。
* `/guides/tireguides/tire-frame-clearance`：保留测量说明和参考表；当前未发现独立交互计算器，不创建名义上的工具页。

对于原先内嵌在 `/guides/tireguides/choose-inner-tube`（以及兼容的 `/guides/tireguides?tab=choose-inner-tube`）内部的气门嘴计算器：
* **当前策略**：保留完整内胎选型页面，不把说明、参数解释和结果上下文拆到另一个 URL。
* **未来整体迁移条件**：只有在确认“直接完成气门嘴匹配”是页面主任务，并准备好完整的工具标题、输入说明、结果解释、数据来源和限制说明后，才将页面范围内的内容整体迁移到 `/tools/valve-length-fitment-calculator`。
* **迁移后的旧 URL**：新旧页面内容和任务等价、回归验收及 canonical/hreflang 完成后，才评估对旧页面整页 301；不能用一个只含计算器的工具页承接旧指南 URL。

---

## 六、 独立工具页的 SEO/GEO 外壳与 Schema 规范

本项改造仅针对页面的**外壳元数据与结构化标签**进行对齐，不涉及内部计算逻辑：

### 1. Schema.org 结构化数据
独立后的 `/tools/*` 页面，统一定义为 WebApplication 实体：
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
* 页面 `<title>` 统一命名规范：`[工具名称] - 工程工具中心 | 品牌名`；
* 自动继承已有的 `useLocalePath` 与 `@nuxtjs/i18n` 多语言配置，生成标准的对称 hreflang。

### 3. SEO 拆分与索引规则

* 搜索引擎主要依据查询意图、页面内容和任务完成度判断相关性，不会因为 URL 位于 `/tools` 或 `/guides` 就自动给予不同排名。目录名称用于组织和导航，不能替代页面价值。
* `/tools` 对人和检索系统都应表达一个清晰承诺：页面打开后可以直接交互、查表或得到结果。工具页的简介、参数说明、结果解释、数据来源和限制说明应与交互放在同一 URL；“工具”标签本身不是排名保证，但比把可操作任务藏在指南 Tab 中更准确地表达任务意图。
* 当前混合页如果以后迁移，优先执行整页迁移：保留完整说明和交互，只改变主路由和页面归类。不要为了得到一个 `/tools` URL 而把同一页面拆成“指南正文 URL + 计算器 URL”。
* 两个 URL 只有在确实存在两份独立内容、用户任务明显不同、正文与说明可以各自成立、标题和摘要不重复，并且各自能通过独立验收时，才同时允许索引。相同计算器可以复用同一个 composable、API 或领域组件，但不能在两个 URL 复制完整正文和完整交互。
* 近重复页面只保留一个主 URL 和 canonical；`noindex`、canonical、合并页面或保留混合页的选择要基于实际内容，而不是为了填满工具目录。
* 301 只用于旧页和新页的搜索意图、主要内容与用户任务基本等价的情况。图文混合旧页不能因为计划建立工具页就自动整页 301；整体迁移完成并验证等价后，才允许把旧 URL 301 到完整工具页。
* `WebApplication` 等 Schema 只描述真实存在、可用且可访问的页面，不是拆分理由，也不保证排名。规划中的工具不得伪装成已上线应用，不得进入 sitemap。

---

## 七、 研发落地分批拆解检查清单

为保证开发工作稳健有序、不遗漏任何细节，按以下四个批次逐步推进。每一批都必须遵守“算法零改动、先验证路由再迁移”的边界：

### 迁移前置核对（必须先完成）
- [ ] 0. 对照 `public/storefront-route-manifest.json`、`definePageMeta` 和实际 HTTP 响应，登记每个旧路径的 canonical、语言前缀、sitemap 状态和是否含指南正文。
- [ ] 0.1 将工具标记为“已实现可迁移”“已实现但与指南混合”或“规划中”；规划项不得生成可点击的工具页链接。
- [ ] 0.2 为每个已实现工具保存迁移前的代表性输入、输出和交互快照；迁移提交不得修改计算 composable、后端服务、领域模型或状态机。

### 第一批：搭建 `/tools` 骨架与大厅页 (Foundation)
- [ ] 1. 新建 `app/pages/tools/index.vue`（工具大厅入口页），实现 3 大板块卡片式网格布局；
- [ ] 2. 在导航配置（Header 菜单及移动端 Drawer）中增加 `TOOLS` 顶级入口；
- [ ] 3. 验证 `/tools` 大厅页面的 i18n 多语言翻译文本与路由跳转畅通。

### 第二批：迁移已有独立页面工具 (Page Migration)
- [ ] 4. 迁移 Schwalbe 外胎选型器至 `app/pages/tools/schwalbe-tire-selector.vue`，保留完成选型所需的介绍和目录说明；
- [ ] 5. 迁移大牌辐条速查矩阵至 `app/pages/tools/brand-wheelset-spoke-specs.vue`，保留数据来源和使用限制；
- [ ] 6. 迁移辐条长度向导至 `app/pages/tools/spoke-length-calculator.vue`，保持完整向导、输入说明和结果面板；
- [ ] 7. 以上页面均按完整页面迁移，不删减完成任务所需的介绍、数据来源或限制说明；
- [ ] 8. 仅为已通过整页等价性验收的页面配置 301：候选为 `/guides/tireguides/schwalbe-tire-selector`、`/guides/spokeguides/spoke-length-calculator` 和 `/guides/spokeguides/brand-wheelset-spoke-specs`。

### 第三批：整体迁移工具优先的混合页面 (Whole-page Migration)
- [ ] 9. 对需要迁移的混合页面先确定完整页面边界，列出标题、说明、参数解释、结果解释、数据来源和限制说明；不先抽走单独计算器；
- [ ] 10. 将胎压页面整体迁移至 `app/pages/tools/dynamic-tire-pressure-calculator.vue`，保留 `calculator` 和 `details` 全部内容；
- [ ] 11. 将内胎气门嘴匹配页面整体迁移至 `app/pages/tools/valve-length-fitment-calculator.vue`，保留完整上下文、说明和交互，不制作精简指南页 + 独立半页工具；
- [ ] 12. 将轮组编法拓扑页面整体迁移至 `app/pages/tools/spoke-lacing-topology-simulator.vue`，保留工程说明、仿真控件和结果解释；
- [ ] 13. 将不锈钢辐条位错页面整体迁移至 `app/pages/tools/spoke-stress-relief-mechanics.vue`，保留材料说明、计算器、数据目录和结果解释；
- [ ] 14. 整体迁移验收通过后，再为胎压、内胎气门嘴、轮组拓扑和不锈钢位错配置对应旧 URL 的 301；任何未完成整页迁移的页面继续保留原 URL。

### 第四批：双向飞轮与收尾验证 (Verification)
- [ ] 15. 在每个独立工具页确认页面自身已包含完成任务所需的说明、结果解释和限制；只有存在独立指南内容时才添加“前往阅读相关指南”链接；
- [ ] 16. 检查动态 Sitemap 只包含已整体迁移、可访问且已确认需要独立 URL 的 `/tools/*` 路由；规划项和仍在旧指南 URL 的页面不得进入工具 sitemap；
- [ ] 17. 逐个整体迁移候选页面进行功能与内容回归，确认**算法、计算数值、参数联动、主任务、页面说明、结果解释和限制均与迁移前一致**；同时验证旧 URL 的 301 或完整页面保留策略；
- [ ] 18. 每次新增工具页前更新“本轮评估的决策基线”和资产迁移表，记录整页范围、重复风险、canonical、sitemap 和 301 决策，防止后续任务把完整迁移改成只抽走计算器。

---
*文档编制完成。本规范为纯信息架构重组与路由迁移实施指南，所有开发工作严禁篡改现有算法逻辑。*

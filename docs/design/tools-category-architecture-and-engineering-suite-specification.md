# 🛠️ 独立工程工具体系 (/tools) 全景架构与迁移实施规范
## Architecture & Migration Specification for Engineering Tools Suite

> **文档版本**: `v1.0-Engineering`  
> **文档位置**: `docs/design/tools-category-architecture-and-engineering-suite-specification.md`  
> **适用范围**: 针对现有代码库中已实现的所有计算器、适配器、选型矩阵及仿真看板的**目录拆解、路由独立化与导航重组**。  
> 
> 🛑 **【最高核心原则：算法零改动，纯架构重组】**  
> **本项目所有核心力学模型、材料物性参数、物理计算引擎与状态机逻辑已在代码库中 100% 完整实现且实机验证正确。**  
> **本文档严禁编写、推导或重新设计任何数学公式与物理算法，绝不增加任何额外计算需求，杜绝技术跑偏！**  
> **本文档唯一目标是：全面盘点现有分散在 `/guides` 中的各块工具资产，规范如何将其解耦平移至独立的顶级分类 `/tools`，理清路由、菜单、面包屑与迁移映射，确保平稳落地、不落下一块。**

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
2. **SEO/GEO 意图错配**：AI 爬虫（ChatGPT, Perplexity）与搜索引擎对“工具（Tool/Calculator）”和“文章（Guide/Article）”分配不同的权重。把成熟工具作为子 Tab 折叠隐藏在文章里，导致工具在搜索结果中无法作为独立的 Web 应用被直接推荐；
3. **页面复杂度过高**：一个指南页面既要加载长图文，又要加载复杂的计算组件，导致单个页面体积臃肿、维护困难。

**解决方案**：将既有的工具组件抽离为独立的顶级 `/tools` 页面，原指南页面仅保留清晰的卡片导流，既让工具简单纯粹、一触即达，又让指南保持清晰轻量。

---

## 二、 “工具 (Tools) + 指南 (Guides)”双飞轮架构

将工具独立后，工具与指南不是非此即彼，而是形成清晰的“双飞轮”分工：

```
                    ┌────────────────────────────────────────────────────────┐
                    │               站点信息架构双轮驱动体系                  │
                    └───────────────────────────┬────────────────────────────┘
                                                │
             ┌──────────────────────────────────┴──────────────────────────────────┐
             ▼                                                                     ▼
   【 /tools (独立工程工具中心) 】                                       【 /guides (技术指南与白皮书) 】
   • 页面形态：纯粹的工具交互、参数输入、结果与工单导出                  • 页面形态：原理深度解析、结构科普、图文指南
   • 核心心智：解决“具体参数怎么选？数值是多少？”                        • 核心心智：解决“原理是什么？为什么这样选？”
   • 内部代码：直接复用现有的成熟计算组件与逻辑                          • 内部代码：纯图文静态内容，移除内嵌的重型计算组件
             │                                                                     │
             └──────────────────────► [ 双向上下文锚点高频引流 ] ◄─────────────────┘
               • 工具页面底部："深入了解此方案的工程原理解析 ➔ 查看对应指南"
               • 指南文章内部："直接使用计算器进行参数匹配 ➔ 启动独立工具"
```

---

## 三、 现有工程工具资产全盘清点与迁移映射表（完整梳理不遗漏）

在此对代码库中所有**已存在或已规划的工具模块**进行地毯式梳理，明确每个工具的**现有代码位置**、**目标独立路由**与**迁移重构策略**。**（不修改内部已有计算逻辑，只做页面解耦与路由迁移）**：

### 3.1 板块 A：轮胎、车圈与气门系统

| 序号 | 工具名称 | 代码库现有文件位置 | 现有访问路径 | 目标独立路由 | 迁移方案说明 |
| :---: | :--- | :--- | :--- | :--- | :--- |
| **1** | **车圈框高与气门嘴/延长嘴穿透匹配器** | `app/components/tireguides/InnerTubeValveLengthAndExtenderFitmentGuide.vue`<br>`app/composables/useInnerTubeValveFitmentCalculator.ts` | `/guides/tireguides` (内嵌在 `choose-inner-tube` Tab 的二级 Tab) | `/tools/valve-length-fitment-calculator` | **重点抽离**：将其从 `InnerTubeGuide.vue` 中拆出，包装为独立的工具页面；原内胎指南中用大卡片引导跳转。现有算法与后端接口 100% 直接复用。 |
| **2** | **动态前后轮智能胎压与滚阻计算器** | `app/pages/guides/tire-pressure.vue`<br>`app/components/tireguides/tirepressure/*` | `/guides/tire-pressure` | `/tools/dynamic-tire-pressure-calculator` | **平移归类**：已是独立页面，直接平移至 `pages/tools/` 目录；原路径配置 301 重定向。 |
| **3** | **Schwalbe 官方外胎选型与周长适配器** | `app/pages/guides/schwalbe-tire-selector.vue`<br>`app/components/tireguides/schwalbe/*` | `/guides/schwalbe-tire-selector` | `/tools/schwalbe-tire-selector` | **平移归类**：已是成熟独立页面，平移至 `pages/tools/` 目录；原路径配置 301 重定向。 |
| **4** | **外胎实测膨胀率与车架安全间隙校验器** | `app/components/tireguides/TireFrameClearanceGuide.vue` | `/guides/tireguides` (内嵌 Tab) | `/tools/tire-frame-clearance-checker` | **独立化**：现有组件封装成独立页面，原 Guides 保留科普说明与跳转链接。 |
| **5** | **Hookless 车圈与轮胎物理安全校验器** | 关联已有 Hookless 适配契约与数据模型 | 看板资料/规划中 | `/tools/hookless-compatibility-checker` | **标准化落地**：按工具统一规范挂载至 Tools 体系。 |

---

### 3.2 板块 B：轮组、辐条与力学工程

| 序号 | 工具名称 | 代码库现有文件位置 | 现有访问路径 | 目标独立路由 | 迁移方案说明 |
| :---: | :--- | :--- | :--- | :--- | :--- |
| **6** | **多步向导式辐条长度精密计算器** | `app/pages/guides/spokeguides/spoke-length-calculator.vue`<br>`app/components/Spoke*.vue` (向导 5 步)<br>`app/composables/useSpokeCalculator*.ts` | `/guides/spokeguides/spoke-length-calculator` | `/tools/spoke-length-calculator` | **核心平移**：现有 5 步向导（头型、车圈几何、PCD、物理修正、条帽）与计算结果已完全实现且验证正确。平移至 `pages/tools/`，不修改内部向导与计算代码，仅对齐顶部导航与布局。 |
| **7** | **高端大牌轮组官方出厂辐条规格速查** | `app/pages/guides/spokeguides/brand-wheelset-spoke-specs.vue`<br>`app/data/brand-wheelset-spoke-specs/*` | `/guides/spokeguides/brand-wheelset-spoke-specs` | `/tools/brand-wheelset-spoke-specs` | **平移归类**：已包含 DT Swiss、Enve、Shimano、Zipp 等原厂数据，直接平移至 `pages/tools/`；原路径配置 301 重定向。 |
| **8** | **轮组编法拓扑与法兰干涉仿真器** | `app/pages/guides/wheelset-spoke-lacing-topology-and-geometry-reference.vue` | `/guides/wheelset-spoke-lacing-topology-and-geometry-reference` | `/tools/spoke-lacing-topology-simulator` | **路由精简平移**：原页面 URL 较长，将其平移至简洁的 Tools 路由，并保留 301 重定向。 |
| **9** | **不锈钢辐条微观位错力学评估器** | `app/pages/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics.vue`<br>`app/composables/useStainlessSteelSpokeDislocationMechanicsCalculation.ts` | `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics` | `/tools/spoke-stress-relief-mechanics` | **平移归类**：平移至 `pages/tools/`，现有微观位错计算与图表保持 100% 原样。 |
| **10** | **车圈气压径向压缩与张力暴跌预测器** | 关联现有力学看板规范与工程数据 | 看板资料/规划中 | `/tools/rim-compression-tension-drop-predictor` | **标准化挂载**：作为轮组力学重要工具挂载至 Tools 体系。 |

---

### 3.3 板块 C：传动系统与整车兼容性

| 序号 | 工具名称 | 代码库现有文件位置 | 现有访问路径 | 目标独立路由 | 迁移方案说明 |
| :---: | :--- | :--- | :--- | :--- | :--- |
| **11** | **塔基花键与飞轮变速兼容性求解器** | 关联现有传动系统适配规范 | 规范文档/规划中 | `/tools/freehub-cassette-fitment-engine` | **标准化挂载**：挂载至 Tools 传动板块。 |

---

## 四、 导航体系与菜单布局改造方案

为了让用户和搜索引擎能清晰感知到全新的工程工具体系，前台导航必须进行轻量而明确的调整：

### 1. 顶部 Header 导航栏与移动端 Drawer 抽屉
* **主导航栏新增顶级项**：
  在 `SHOP`、`GUIDES`、`ABOUT` 之间，新增 **`TOOLS`（工程工具）** 顶级导航项：
  ```
  [ LOGO ]    SHOP    TOOLS (新)    GUIDES    ABOUT    [ 搜索 / 语言 / 购物车 ]
  ```
* **下拉菜单（Mega Menu / Dropdown）三列布局**：
  点击/悬浮 `TOOLS` 展开清晰的分类菜单：
  * **第一列：轮胎与气门系统**（气门嘴/延长嘴匹配、动态胎压计算、Schwalbe 外胎选型、外胎间隙校验）
  * **第二列：轮组与辐条工程**（辐条长度计算器、大牌出厂辐条速查、轮组编法拓扑仿真、微观位错力学）
  * **第三列：传动与通用系统**（塔基飞轮兼容求解器、Hookless 安全校验器）

### 2. `/tools` 大厅入口页面 (Tools Hub / Dashboard)
* **路由**: `/tools`
* **页面形态**: 工业仪表盘风格的工具导航大厅；
* **内容构成**:
  * 顶部：清晰的工具集介绍与快速搜索框；
  * 主体：按 3 大板块以卡片网格形式陈列上述 11 个工具，每张卡片展示工具名称、简短用途说明、核心功能标签与【启动工具 ➔】按钮；
  * 无多余长篇图文，保持纯粹的工具导航索引属性。

### 3. 面包屑对齐
所有独立工具页面的面包屑统一规范为：
`首页 ➔ 工程工具 (Tools) ➔ [具体工具名称]`

---

## 五、 原 /guides 兼容与 301 重定向策略（无损过渡）

为了保障现有已被收录的 URL 权重不丢失，并兼顾老用户的使用习惯，采取以下过渡策略：

### 1. 服务端 301 永久重定向映射表
对于已经作为独立页面存在的工具路由，在 `server/routes` 或中间件中配置 301 重定向：
* `/guides/tire-pressure` ➔ `301` ➔ `/tools/dynamic-tire-pressure-calculator`
* `/guides/schwalbe-tire-selector` ➔ `301` ➔ `/tools/schwalbe-tire-selector`
* `/guides/spokeguides/spoke-length-calculator` ➔ `301` ➔ `/tools/spoke-length-calculator`
* `/guides/spokeguides/brand-wheelset-spoke-specs` ➔ `301` ➔ `/tools/brand-wheelset-spoke-specs`
* `/guides/wheelset-spoke-lacing-topology-and-geometry-reference` ➔ `301` ➔ `/tools/spoke-lacing-topology-simulator`
* `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics` ➔ `301` ➔ `/tools/spoke-stress-relief-mechanics`

### 2. 原内嵌 Tab 改造（例如内胎指南）
对于原先内嵌在 `/guides/tireguides?tab=choose-inner-tube` 内部的气门嘴计算器：
* **改造前**：内胎指南里直接挂载庞大的 `InnerTubeValveLengthAndExtenderFitmentGuide.vue` 组件，导致指南页面代码量巨大。
* **改造后**：
  * 计算器组件整体迁移至 `/tools/valve-length-fitment-calculator` 独立运行；
  * 内胎指南的 Tab 中保留精简优雅的**“选型工具导流卡片”**（展示核心预览图 + 说明），提供高亮按钮：【🚀 启动全功能气门嘴穿透力学匹配器 (跳转至 /tools)】；
  * 既彻底净化了指南页面的体积，又让独立工具拥有专有 URL。

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

---

## 七、 研发落地分批拆解检查清单

为保证开发工作稳健有序、不遗漏任何细节，按以下四个批次逐步推进：

### 第一批：搭建 `/tools` 骨架与大厅页 (Foundation)
- [ ] 1. 新建 `app/pages/tools/index.vue`（工具大厅入口页），实现 3 大板块卡片式网格布局；
- [ ] 2. 在导航配置（Header 菜单及移动端 Drawer）中增加 `TOOLS` 顶级入口；
- [ ] 3. 验证 `/tools` 大厅页面的 i18n 多语言翻译文本与路由跳转畅通。

### 第二批：迁移已有独立页面工具 (Page Migration)
- [ ] 4. 迁移胎压计算器至 `app/pages/tools/dynamic-tire-pressure-calculator.vue`；
- [ ] 5. 迁移 Schwalbe 外胎选型器至 `app/pages/tools/schwalbe-tire-selector.vue`；
- [ ] 6. 迁移大牌辐条速查矩阵至 `app/pages/tools/brand-wheelset-spoke-specs.vue`；
- [ ] 7. 迁移轮组编法拓扑仿真器至 `app/pages/tools/spoke-lacing-topology-simulator.vue`；
- [ ] 8. 迁移微观位错力学评估器至 `app/pages/tools/spoke-stress-relief-mechanics.vue`；
- [ ] 9. 为上述原 `/guides/...` 路径配置 301 重定向，验证旧链接无缝跳转。

### 第三批：抽离内嵌组件并独立化 (Component Extraction)
- [ ] 10. 将气门嘴穿透力学组件封装为独立页面 `app/pages/tools/valve-length-fitment-calculator.vue`；
- [ ] 11. 在原 `app/components/tireguides/InnerTubeGuide.vue` 中将内嵌组件替换为跳转导流卡片；
- [ ] 12. 将辐条长度向导计算器平移至 `app/pages/tools/spoke-length-calculator.vue`，保持向导 5 步逻辑 100% 原样复用。

### 第四批：双向飞轮与收尾验证 (Verification)
- [ ] 13. 在各个独立工具页面底部，确认均已添加“前往阅读相关指南”的卡片链接；
- [ ] 14. 检查动态 Sitemap 自动包含所有 `/tools/*` 路由；
- [ ] 15. 逐个工具页面进行功能回归测试，确认**算法、计算数值、参数联动完全与迁移前 100% 保持一致**。

---
*文档编制完成。本规范为纯信息架构重组与路由迁移实施指南，所有开发工作严禁篡改现有算法逻辑。*

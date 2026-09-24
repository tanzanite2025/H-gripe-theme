# ⚙️ 传动系统与塔基飞轮适配引擎实施方案
## Drivetrain & Freehub Body Fitment Engineering Specification

> **文档版本**: `v1.0-Engineering`  
> **文档位置**: `docs/design/drivetrain-freehub-fitment-engine-specification.md`  
> **关联架构基线**: 
> - [《技术算法与工程仿真页面体系化配套实施规范》](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/docs/design/technical-algorithm-engineering-standards.md)
> - [《车型前叉花鼓适配资料库设计》](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/docs/design/fitment-catalog-domain-architecture.md)
> - [《电商 URL 与 SEO 黄金架构》](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/docs/seo/ECOMMERCE_URL_ARCHITECTURE.md)
> - [《轮组适配问卷规格》](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/docs/design/wheelset-fit-questionnaire-specification.md)
> 
> **适用范围**: 前台轮组购买指南塔基计算器 (`/guides/wheelset-buyers`)、商品详情页技术资料弹窗、前台传动与塔基兼容性速查页面。  
> **核心目标**: 彻底颠覆前端纯静态数组打补丁模式，建立以 **Go 后端权威物理干涉与垫圈规则库** 为中枢、**SEO/GEO 服务端首屏直出与 Schema.org 权威索引** 为流量抓手、**复用项目现有高清实物摄影图库 (`/public/wheelsetbuyersguide/choose freehub/*.webp`)** 进行直观实拍对比的纯粹、高精度技术计算器。

---

## 目录
1. [背景与问题定性（为什么禁止前端打补丁）](#一-背景与问题定性为什么禁止前端打补丁)
2. [官方权威机械工学数据与物理干涉模型](#二-官方权威机械工学数据与物理干涉模型)
   - [2.1 链条节距与 10T vs 11T / 9T 物理干涉数学证明](#21-链条节距与-10t-vs-11t--9t-物理干涉数学证明)
   - [2.2 SRAM 官方 XD / XDR / T-Type 驱动体标准](#22-sram-官方-xd--xdr--t-type-驱动体标准)
   - [2.3 Shimano 官方 HG / HG-EV / HG L2 / Micro Spline 标准](#23-shimano-官方-hg--hg-ev--hg-l2--micro-spline-标准)
   - [2.4 Campagnolo 官方 N3W 驱动体与适配套筒标准](#24-campagnolo-官方-n3w-驱动体与适配套筒标准)
   - [2.5 国产与高频第三方传动生态标准（L-TWOO, SENSAH, microSHIFT, SUNSHINE, ZTTO）](#25-国产与高频第三方传动生态标准l-twoo-sensah-microshift-sunshine-ztto)
   - [2.6 全球传动系统塔基适配与垫圈权威全景矩阵 (Spacer Guide)](#26-全球传动系统塔基适配与垫圈权威全景矩阵-spacer-guide)
3. [系统领域架构与前后端契约 (Domain Engine)](#三-系统领域架构与前后端契约-domain-engine)
   - [3.1 Go 后端领域模型设计](#31-go-后端领域模型设计)
   - [3.2 判定算法状态机与 Fail Loudly 物理阻断](#32-判定算法状态机与-fail-loudly-物理阻断)
   - [3.3 权威 API 接口契约](#33-权威-api-接口契约)
4. [SEO 与 GEO（生成式 AI 优化）实战落地](#四-seo-与-geo生成式-ai-优化实战落地)
   - [4.1 SSR 服务端首屏直出知识图谱](#41-ssr-服务端首屏直出知识图谱)
   - [4.2 Schema.org 结构化数据注入（TechArticle + HowTo + Dataset）](#42-schemaorg-结构化数据注入techarticle--howto--dataset)
   - [4.3 国产出海长尾搜索与全球车友流量护城河 (The Global Drivetrain MOAT)](#43-国产出海长尾搜索与全球车友流量护城河-the-global-drivetrain-moat)
   - [4.4 多语言与动态 Sitemap 自动收录](#44-多语言与动态-sitemap-自动收录)
5. [前端交互与实物图库复用规范 (Real Photo Showcase)](#五-前端交互与实物图库复用规范-real-photo-showcase)
   - [5.1 联动复用项目现有高清实拍图库](#51-联动复用项目现有高清实拍图库)
   - [5.2 现有两级联动极简交互无缝升级 (2-Step Direct Fitment)](#52-现有两级联动极简交互无缝升级-2-step-direct-fitment)
6. [严格明确业务边界与 Non-Goals（绝不侵入交易与订单）](#六-严格明确业务边界与-non-goals绝不侵入交易与订单)
   - [6.1 纯只读技术资料库定位](#61-纯只读技术资料库定位)
   - [6.2 零商品侵入与零订单耦合原则](#62-零商品侵入与零订单耦合原则)
7. [分阶段实施蓝图与研发核对清单](#七-分阶段实施蓝图与研发核对清单)

---

## 一、 背景与问题定性（为什么禁止前端打补丁）

当前前端组件 [`FreehubGroupsetHelper.vue`](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/nuxt-i18n/app/components/FreehubGroupsetHelper.vue) 采用纯前端硬编码数组 `FREEHUB_OPTIONS`，仅提供品牌与套件型号的简单映射。

### 存在的三大根本缺陷：
1. **机械逻辑脱离现实**：
   - 无法处理飞轮齿数（10T vs 11T）的分水岭差异；
   - 遗漏了 NX Eagle（HG 塔基）与 GX Eagle（XD 塔基）的同速别分裂；
   - 缺失所有关键垫圈（1.85mm / 1.0mm / 4.4mm 适配套筒）的装配指引；
   - 无法提示 Shimano 12速公路飞轮对 11速 HG 塔基的向下兼容。
2. **SEO 与 GEO 价值归零**：
   - 客户端纯 JS 渲染导致搜索引擎与 AI 爬虫（Google Gemini, ChatGPT, Perplexity）首屏抓取为空白下拉列表，无法建立全球长尾技术问答权威。
3. **多语言维护地狱与逻辑碎片化**：
   - 前端硬编码导致必须在 20 种语言包中重复维护文案，每增加一种齿数或垫圈情况，多语言漏配率极高；计算逻辑散落在 Vue 组件内，无法统一复用。

**架构原则**：禁止任何前端打补丁，全面升级为 **后端权威引擎 + SSR/SEO/GEO 直出 + 纯粹只读计算工具**。

---

## 二、 官方权威机械工学数据与物理干涉模型

本节所有数据均依据 **SRAM 官方服务手册**、**Shimano SI 官方技术文档**、**Campagnolo 官方技术手册** 及 ISO 自行车标准严格核实。

### 2.1 链条节距与 10T vs 11T / 9T 物理干涉数学证明

自行车链条国际标准节距（Pitch）恒定为：
$$P = \frac{1}{2}\text{ 英寸} = 12.7\text{ mm}$$

对于齿数为 $N$ 的飞轮齿片，其理论节圆直径（Pitch Diameter $D_p$）公式为：
$$D_p = \frac{P}{\sin(180^\circ / N)} = \frac{12.7}{\sin(180^\circ / N)}\text{ mm}$$

齿根槽底圆直径（Root Diameter $D_{\text{root}}$）公式为：
$$D_{\text{root}} \approx D_p - 7.94\text{ mm}$$

#### 物理干涉对比表：
| 齿数 ($N$) | 节圆直径 $D_p$ | 齿根直径 $D_{\text{root}}$ | Shimano 传统 HG 塔基外径 | 物理装配结论与力学机理 |
| :---: | :---: | :---: | :---: | :--- |
| **11T** | **$44.88\text{ mm}$** | **$36.94\text{ mm}$** | **$34.80\text{ mm}$** | **【物理兼容】** $D_{\text{root}} > 34.8\text{mm}$，齿根下方留有约 $1.07\text{mm}$ 的金属壁厚，可承受高扭矩冲刺。 |
| **10T** | **$41.10\text{ mm}$** | **$33.16\text{ mm}$** | **$34.80\text{ mm}$** | **【严重物理干涉】** 齿根比塔基外径小 $1.64\text{mm}$！若硬套 HG 塔基，齿根直接被塔基贯穿，**壁厚为负数，物理上绝无可能安装**。 |
| **9T** | **$37.35\text{ mm}$** | **$29.41\text{ mm}$** | **$34.80\text{ mm}$** | **【严重物理干涉】** 齿根比 HG 外径小 $5.39\text{mm}$，必须采用大幅收窄前端外径的专用塔基（如 Campy N3W 或 E*thirteen）。 |

**结论**：10T 及更小齿片必须使用阶梯式收窄前端外径的专用塔基（XD/XDR/Micro Spline/N3W）。

---

### 2.2 SRAM 官方 XD / XDR / T-Type 驱动体标准

依据 SRAM Driver Body Service Manual：

1. **XD vs XDR 轴向尺寸差**：
   - XDR（XD Road）花键主体轴向长度为 **$38.00\text{ mm}$**；
   - XD（山地标准）花键主体轴向长度为 **$36.15\text{ mm}$**；
   - **XDR 比 XD 整整长了 $1.85\text{ mm}$**（为公路车 130/135/142mm 开档下维持法兰距与辐条张力平衡而设计）。
2. **垫圈硬性安装规范**：
   - **XD 飞轮装在 XDR 塔基上**：**必须使用 $1.85\text{ mm}$ 垫圈**，且该垫圈必须安装在塔基最内侧台阶处（behind the cassette）。漏装将导致飞轮锁死后依然产生 $1.85\text{mm}$ 轴向旷量，引起严重跳齿与后拨撞击辐条！
   - **XDR 飞轮装在 XDR 塔基上**：**严禁使用任何垫圈**。若误加垫圈，飞轮锁环螺纹咬合深度不足，受力时直接滑丝损坏塔基！
   - **XDR 飞轮装在 XD 塔基上**：**物理不兼容**。飞轮内螺纹无法接触到塔基螺纹，绝对无法安装。
3. **SRAM Eagle 12速内部断层**：
   - **SX Eagle / NX Eagle (PG-1210 / PG-1230)**：最小齿为 **11T**（11-50T），采用传统分体齿片与标准花键内槽，**必须安装在 Shimano HG-10/11 塔基上**！
   - **GX / X01 / XX1 Eagle (XG-1275 / XG-1295 / XG-1299)**：最小齿为 **10T**（10-50T / 10-52T），一体切削工艺，**必须安装在 XD / XDR（需垫圈）塔基上**。
4. **SRAM Eagle Transmission (T-Type)**：
   - 全系飞轮（XS-1270 / XS-1295 / XS-1297 / XS-1299）必须安装在 **XD 塔基**上（或在 XDR 塔基上配合 $1.85\text{mm}$ 垫圈）。

---

### 2.3 Shimano 官方 HG / HG-EV / HG L2 / Micro Spline 标准

依据 Shimano SI (Service Instructions) 官方规范：

1. **Shimano 12速公路飞轮向下兼容性**：
   - 12速公路飞轮（Dura-Ace CS-R9200、Ultegra CS-R8100、105 CS-R7100）全系以 **11T 起跳**（11-30T / 11-34T）；
   - **飞轮内花键采用向下兼容设计，可以直接安装在标准 11速公路 HG (HG-EV) 塔基上**；
   - **安装时无需使用任何垫圈**！直接涂抹防咬膏按 40 N·m 扭矩锁紧。
2. **Shimano HG L2 塔基的单向兼容性**：
   - HG L2 是 Shimano 在新一代顶级轮组（如 WH-R9270）上推出的全铝轻量化 12速公路塔基（18齿花键）；
   - **单向兼容性**：12速公路飞轮兼容普通 HG 塔基和 HG L2 塔基；但 **HG L2 塔基绝对不支持老款 11速公路飞轮**（11速飞轮花键槽深度与齿数无法插进 HG L2）。
3. **Shimano 11速山地飞轮装在 11速公路塔基**：
   - Shimano 11速山地飞轮（CS-M8000 / CS-M7000 / CS-M9000，如 11-40T / 11-42T）最大齿片背面采用内凹悬臂蜘蛛爪设计，其底座花键厚度与 10速公路飞轮完全相同（$34.9\text{mm}$）；
   - 装在 $36.75\text{mm}$ 的 11速公路 HG 塔基上时，**必须加装 $1.85\text{ mm}$ 垫圈**。
4. **Shimano 10速公路飞轮的 2.85mm 复合垫圈细节**：
   - Shimano 10速高阶公路飞轮（CS-7900 / CS-6700 / CS-5700）背面设计有 $1.0\text{mm}$ 内凹容差，飞轮出厂原装自带一片 $1.0\text{mm}$ 垫圈；
   - 当安装在 11速公路 HG 塔基上时，**必须同时安装 $1.85\text{ mm}$ 塔基转换垫圈 + 飞轮自带的 $1.0\text{ mm}$ 垫圈 = 共计 $2.85\text{ mm}$ 垫圈**！
   - 例外：Tiagra CS-4600 / CS-HG500-10 因背面无凹槽设计，仅需 $1.85\text{ mm}$ 垫圈。
5. **Shimano Micro Spline (MS) 塔基边界**：
   - 23花键高精密微型花键，专为 12速山地/碎石飞轮（10T 起跳，如 10-45T / 10-51T）设计；
   - **与公路 12速 HG 花键完全物理互斥**，公路 12速飞轮不可装在 Micro Spline 塔基上。

---

### 2.4 Campagnolo 官方 N3W 驱动体与适配套筒标准

依据 Campagnolo N3W 官方白皮书与技术公报：

1. **N3W 几何尺寸**：
   - 花键剖面形状与经典 Campagnolo 9/10/11/12速花键完全一致；
   - **花键有效轴向长度缩短了整整 $4.40\text{ mm}$**（由 $34.5\text{mm}$ 缩短至 $30.1\text{mm}$），以便在前端悬臂处收窄容纳 9T / 10T 最小齿片。
2. **适配套件 AC21-N3W (Kit Adapter)**：
   - 配备包含：**$+4.4\text{ mm}$ 延长花键阳极氧化铝套筒** + **专用加长深螺纹锁环 (Extended Lockring)**；
   - **使用场景**：当车友使用配备 N3W 塔基的新轮组，安装传统的 Campagnolo 10速、11速或 12速飞轮（11T 起跳）时，**必须加装该 $+4.4\text{mm}$ 延长套筒并使用加长锁环**；
   - **原生场景**：安装 Campagnolo Ekar 13速（9-42T / 10-44T）或原生 N3W 飞轮时，**严禁加装套筒，直装锁紧**。

---

### 2.5 国产与高频第三方传动生态标准（L-TWOO, SENSAH, microSHIFT, SUNSHINE, ZTTO）

在海外 AliExpress、Amazon 以及国内平民骑行圈中，**蓝图（L-TWOO）、顺泰（SENSAH）、微转（microSHIFT）、日晖（SUNSHINE）、ZTTO** 等品牌占有极高的市场存量。欧美大厂（DT Swiss / Hunt / Zipp）的计算器傲慢地将它们完全排除，导致车友在选购轮组时存在巨大的塔基适配焦虑。

收录国产与第三方品牌的机械工学本质在于：**绝大多数国产品牌为了最大化降低车友换轮门槛，全系采用了以 11T 为最小齿的设计，从而 100% 拥抱存量最大的标准 Shimano HG 塔基**：

1. **蓝图 L-TWOO**：
   - **公路电子/机械 12速 (eRX / RX / R9 12S)**：原厂飞轮全系采用 **11-32T / 11-34T**，在物理上**直接兼容标准 Shimano HG-11 公路塔基（无需更换任何特殊塔基，无垫圈）**；
   - **山地 12速 (A12)**：采用 11-50T / 11-52T 分体齿片，**兼容 Shimano HG 塔基**（装在公路轮组上需垫 1.85mm 垫圈）；
   - **Gravel 11速/12速 (GR9 / GRT)**：采用 11-42T 齿比，兼容 Shimano HG 塔基。
2. **顺泰 SENSAH**：
   - **公路 12速 (Empire Pro 帝国 12S)**：搭配 11-32T / 11-34T 飞轮，**完全运行于标准 Shimano HG-11 塔基**；
   - **Gravel 11/12速 (SRX Pro 单盘)**：搭配 11-42T / 11-46T / 11-50T 飞轮，**全系兼容 Shimano HG 塔基**。
3. **微转 microSHIFT**：
   - **Gravel 10速单盘 (Sword 剑)**：原厂 11-48T 宽齿比，**专为传统 Shimano HG 塔基优化**，无需 XD 或 MS 塔基；
   - **山地 10速/9速 (Advent X 11-48T / Advent 11-42T)**：全系兼容 **Shimano HG 塔基**。
4. **日晖 SUNSHINE / VG Sports（全球出海主力飞轮）**：
   - **12速公路改装飞轮 (11-30T / 11-32T / 11-34T)**：专门为了让车友在老款 **Shimano HG 塔基轮组** 上运行 12速变速套件（如 SRAM AXS 或 Shimano 12S），无需更换昂贵的 XDR 塔基；
   - **12速山地大飞轮 (11-50T / 11-52T)**：让车友在普通 HG 塔基轮组上直接运行 SRAM Eagle 或 Shimano 12速山地后拨。
5. **ZTTO（外贸超轻一体切削飞轮）**：
   - **ZTTO HG 12速超轻 (11-32T / 11-34T / 11-50T)**：走 **Shimano HG 塔基**；
   - **ZTTO XD 12速超轻 (9-50T / 10-50T / 10-52T)**：走 **SRAM XD 塔基**；
   - **ZTTO MS 12速超轻 (10-51T / 10-52T)**：走 **Shimano Micro Spline 塔基**。

---

### 2.6 全球传动系统塔基适配与垫圈权威全景矩阵 (Spacer Guide)

下表为本系统 Go 领域引擎与 SSR 首屏直出的核心事实矩阵（涵盖三大洋品牌与国产主流品牌）：

| 变速套件 / 飞轮品牌与型号 | 速别 | 飞轮最小齿 | 权威推荐塔基 | 必需垫圈/适配套件 | 关键机械工学原因 |
| :--- | :---: | :---: | :--- | :--- | :--- |
| **SRAM Red / Force / Rival AXS** | 12S | **10T** | **SRAM XDR** | **无垫圈 (0 mm)** | 原厂 10T 一体飞轮，直接锁紧在 XDR 塔基。 |
| **SRAM AXS (搭配日晖/第三方飞轮)** | 12S | **11T** | **Shimano HG-11** | **无垫圈 (0 mm)** | 第三方 11T 起跳 12速公路飞轮，兼容老款 HG-11 轮组。 |
| **SRAM XX1/X01/GX Eagle** | 12S | **10T** | **SRAM XD** (山地)<br>或 **SRAM XDR** | XDR 必须加 **$1.85\text{ mm}$ 垫圈**；<br>XD **无垫圈** | XDR 比 XD 塔基长 1.85mm，不加垫圈飞轮轴向旷动。 |
| **SRAM NX / SX Eagle** | 12S | **11T** | **Shimano HG-11** | **$1.85\text{ mm}$ 垫圈** (若装公路HG-11) | NX 为 11-50T 传统花键，基座为山地宽，公路塔基需垫圈。 |
| **SRAM Transmission (T-Type)** | 12S | **10T** | **SRAM XD** | XDR 必须加 **$1.85\text{ mm}$ 垫圈** | 全系 XS-12xx 飞轮采用 XD 接口规范。 |
| **Shimano Dura-Ace R9200 / Ultegra R8100 / 105 R7100** | 12S | **11T** | **Shimano HG-11**<br>或 **HG L2** | **无垫圈 (0 mm)** | 12速公路飞轮花键向下兼容 11速 HG 塔基，严禁装 Micro Spline。 |
| **Shimano XTR M9100 / XT M8100 / SLX M7100 / Deore M6100** | 12S | **10T** | **Micro Spline (MS)** | **无垫圈 (0 mm)** | 10T 齿根干涉 HG，必须使用 Shimano 微型 23 花键塔基。 |
| **Shimano Dura-Ace R9100 / Ultegra R8000 / 105 R7000** | 11S | **11T** | **Shimano HG-11** | **无垫圈 (0 mm)** | 行业标准 11速公路长塔基（$36.75\text{mm}$）。 |
| **Shimano XT M8000 / SLX M7000 (山地 11速)** | 11S | **11T** | **Shimano HG-11** | **$1.85\text{ mm}$ 垫圈** | 山地 11速大齿背部悬臂，底座为 10速宽度。 |
| **Shimano 10速公路 (CS-6700 / CS-5700)** | 10S | **11T/12T** | **Shimano HG-11** | **$1.85\text{ mm} + 1.0\text{ mm} = 2.85\text{ mm}$** | 塔基转换差 1.85mm + 飞轮背部凹槽 1.0mm 组合垫圈。 |
| **Shimano Tiagra 10速 (CS-4600 / HG500-10)** | 10S | **11T/12T** | **Shimano HG-11** | **$1.85\text{ mm}$ 垫圈** | 平底飞轮设计，无需额外的 1.0mm 垫圈。 |
| **Campagnolo Ekar** | 13S | **9T/10T** | **Campagnolo N3W** | **无套筒 (0 mm)** | 原生 N3W 紧凑塔基设计。 |
| **Campagnolo Super Record / Record / Chorus** | 11S/12S | **11T** | **Campagnolo N3W** | **$+4.4\text{ mm}$ 套筒 + 延长锁环**<br>(AC21-N3W) | N3W 较经典塔基短 4.4mm，装 11/12S 必须加原厂适配套筒。 |
| **Campagnolo 经典轮组** | 9S-12S | **11T** | **Campagnolo Classic** | **无套筒 (0 mm)** | 传统标准长塔基（$34.5\text{mm}$）。 |
| **蓝图 L-TWOO eRX / RX / R9 (公路 12速)** | 12S | **11T** (11-32T/34T) | **Shimano HG-11** | **无垫圈 (0 mm)** | 国产 12速公路采用 11T 起跳花键，向下直插 HG-11 塔基。 |
| **蓝图 L-TWOO A12 (山地 12速 11-50T)** | 12S | **11T** (11-50T/52T) | **Shimano HG-11** | **$1.85\text{ mm}$ 垫圈** (若装公路轮) | 传统 HG 花键山地底座，公路长塔基需垫圈，山地塔基直装。 |
| **顺泰 SENSAH Empire Pro (公路 12速)** | 12S | **11T** (11-32T/34T) | **Shimano HG-11** | **无垫圈 (0 mm)** | 专为 Shimano HG 塔基设计，车友升级 12速无需换花鼓塔基。 |
| **顺泰 SENSAH SRX Pro (Gravel 11S/12S)** | 11S/12S | **11T** (11-42T/50T) | **Shimano HG-11** | **$1.85\text{ mm}$ 垫圈** (若装公路轮) | 采用标准山地 HG 飞轮接口。 |
| **微转 microSHIFT Sword (Gravel 10速)** | 10S | **11T** (11-48T) | **Shimano HG-11** | **$1.85\text{ mm}$ 垫圈** (若装公路轮) | 宽齿比 10速单盘系统，彻底拥抱通用性最强的 HG 塔基。 |
| **微转 microSHIFT Advent X (山地 10速)** | 10S | **11T** (11-48T) | **Shimano HG-11** | **$1.85\text{ mm}$ 垫圈** (若装公路轮) | 平民高性价比 10速宽齿比，全系支持标准 HG 塔基。 |
| **日晖 SUNSHINE 12速公路飞轮** | 12S | **11T** (11-30T~34T) | **Shimano HG-11** | **无垫圈 (0 mm)** | 全球 AliExpress 出海爆款，解决 12速套件免换塔基痛点。 |
| **日晖 SUNSHINE 12速山地大飞轮** | 12S | **11T** (11-50T/52T) | **Shimano HG-11** | **$1.85\text{ mm}$ 垫圈** (若装公路轮) | 让普通 HG 轮组直接运行 Eagle / Shimano 12速山地后拨。 |
| **ZTTO SLR 超轻 12速 HG 飞轮** | 12S | **11T** (11-32T/34T) | **Shimano HG-11** | **无垫圈 (0 mm)** | 超轻切削公路飞轮，走标准 HG-11 塔基。 |
| **ZTTO SLR 超轻 12速 XD 飞轮** | 12S | **9T/10T** (9-50T/10-52T) | **SRAM XD** | XDR 塔基需 **$1.85\text{ mm}$ 垫圈** | 一体切削 9T/10T 起跳，专为 SRAM XD 塔基设计。 |
| **轮峰 Wheeltop EDS TX (电子 11S/12S)** | 11S/12S | **11T** (11-32T/34T) | **Shimano HG-11** | **无垫圈 (0 mm)** | 国产无线电变系统，标配飞轮完美走 Shimano HG 塔基。 |

---

## 三、 系统领域架构与前后端契约 (Domain Engine)

根据 [`docs/design/fitment-catalog-domain-architecture.md`](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/docs/design/fitment-catalog-domain-architecture.md) 确立的领域划分原则，传动与塔基适配隶属于独立的 `fitmentcatalog` 领域，由 Go 后端作为单一权威事实源（SSOT）。

### 3.1 Go 后端领域模型设计

在 `internal/domain/fitmentcatalog/drivetrain_rules.go` 中建立规则模型：

```go
package fitmentcatalog

// FreehubStandard 塔基标准枚举
type FreehubStandard string

const (
    FreehubStandardHG11        FreehubStandard = "HG-11"
    FreehubStandardHGL2        FreehubStandard = "HG-L2"
    FreehubStandardXDR         FreehubStandard = "XDR"
    FreehubStandardXD          FreehubStandard = "XD"
    FreehubStandardMicroSpline FreehubStandard = "MICRO_SPLINE"
    FreehubStandardN3W         FreehubStandard = "N3W"
    FreehubStandardCampyClassic FreehubStandard = "CAMPY_CLASSIC"
)

// SpacerRequirement 垫圈与适配套件要求
type SpacerRequirement struct {
    Required    bool    `json:"required"`
    ThicknessMM float64 `json:"thickness_mm"`
    PartCode    string  `json:"part_code,omitempty"`    // 如 AC21-N3W
    Position    string  `json:"position,omitempty"`     // "inner_base" | "adapter_sleeve"
    Description string  `json:"description"`
}

// CassetteFitmentRule 飞轮与塔基适配规则
type CassetteFitmentRule struct {
    RuleID             string          `json:"rule_id"`
    Brand              string          `json:"brand"`               // Shimano, SRAM, Campagnolo
    CassetteSpec       string          `json:"cassette_spec"`       // 机器唯一标识，如 "sram_12s_10_52t"
    DisplayName        string          `json:"display_name"`        // "12速 10-52T / 10-50T"
    HintGroupsets      string          `json:"hint_groupsets"`      // "如 GX / X01 / XX1 Eagle / Transmission"
    Speed              int             `json:"speed"`               // 12
    MinCogTeeth        int             `json:"min_cog_teeth"`       // 10
    MaxCogTeeth        int             `json:"max_cog_teeth"`       // 52
    RecommendedFreehub FreehubStandard `json:"recommended_freehub"` // FreehubStandardXD / FreehubStandardXDR 等
    Spacer             SpacerRequirement `json:"spacer"`
    ImageSrc           string          `json:"image_src"`           // 项目已有实拍图路径
    MechanicalNotes    string          `json:"mechanical_notes"`
}
```

### 3.2 判定算法状态机与 Fail Loudly 物理阻断

核心算法依据飞轮规格（齿数）进行绝对物理判定，严禁任何隐式兜底：

```go
func (e *DrivetrainFitmentEngine) CalculateByCassette(brand string, cassetteSpec string) (*CassetteFitmentRule, error) {
    rule, exists := e.rulesIndex[fmt.Sprintf("%s:%s", strings.ToUpper(brand), cassetteSpec)]
    if !exists {
        return nil, fmt.Errorf("[CRITICAL] Unknown cassette spec '%s' for brand '%s'", cassetteSpec, brand)
    }

    // 齿根干涉断言 (10T / 9T 严禁装配 HG 塔基)
    if rule.RecommendedFreehub == FreehubStandardHG11 && rule.MinCogTeeth <= 10 {
        return nil, fmt.Errorf(
            "[CRITICAL] Physical Interference: %dT cog root diameter (%.2fmm) < HG-11 outer diameter (34.80mm)",
            rule.MinCogTeeth, calculateRootDiameter(rule.MinCogTeeth),
        )
    }

    return rule, nil
}
```

### 3.3 权威 API 接口契约

#### 1. 动态选型试算接口
- **Endpoint**: `POST /api/v1/fitment/drivetrain/calculate`
- **Request Payload**:
  ```json
  {
    "brand": "SRAM",
    "cassette_spec": "sram_12s_10_52t"
  }
  ```
- **Response Payload**:
  ```json
  {
    "code": 0,
    "data": {
      "brand": "SRAM",
      "cassette_spec": "sram_12s_10_52t",
      "display_name": "12速 10-52T / 10-50T (如 GX/X01/XX1 Eagle)",
      "recommended_freehub": "SRAM XD (山地) / SRAM XDR (公路/Gravel)",
      "spacer": {
        "required": true,
        "thickness_mm": 1.85,
        "position": "inner_base",
        "description": "若轮组为 XDR 塔基，必须在塔基底部加装 1.85mm 垫圈；若为 XD 塔基，直接安装无垫圈。"
      },
      "image_src": "/public/wheelsetbuyersguide/choose%20freehub/sram-xdr-road-11-12-speed-freehub.webp",
      "mechanical_fact": "10T 最小齿齿根内径 (33.16mm) 小于 HG 塔基外径 (34.80mm)，必须使用阶梯收窄的 XD/XDR 塔基。",
      "rule_version": "v1.0"
    }
  }
  ```

#### 2. 全景知识矩阵接口（供 SSR 首屏直出）
- **Endpoint**: `GET /api/v1/fitment/drivetrain/matrix`
- **Cache**: Cloudflare CDN Edge Cache 24 小时（带 ETag 校验）。

---

## 四、 SEO 与 GEO（生成式 AI 优化）实战落地

### 4.1 SSR 服务端首屏直出知识图谱
Nuxt 3 服务端渲染（SSR）在生命周期钩子 `async setup()` 中预先拉取 `/api/v1/fitment/drivetrain/matrix`：
- 首屏直出原生 HTML `<table>`，展示全量套件、齿数、塔基及垫圈规则；
- 爬虫（Googlebot, GPTBot, PerplexityBot, ClaudeBot）在 **禁用 JavaScript** 的情况下，依然能 100% 读取完整技术事实，实现精准长尾检索词收割。

### 4.2 Schema.org 结构化数据注入
在页面 `<head>` 中动态注入结构化数据：

```html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "TechArticle",
      "headline": "全球自行车塔基与套件齿数（10T/11T）适配及垫圈安装工程白皮书",
      "description": "依据 ISO 链条节距力学公式，深度剖析 SRAM XD/XDR、Shimano HG/Micro Spline、Campagnolo N3W 塔基兼容性与 1.85mm 垫圈规范。",
      "proficiencyLevel": "Expert",
      "inLanguage": "zh-CN"
    },
    {
      "@type": "HowTo",
      "name": "如何在 SRAM XDR 塔基上正确安装 SRAM XD 飞轮",
      "step": [
        {
          "@type": "HowToStep",
          "name": "放置 1.85mm 专用垫圈",
          "text": "将 1.85mm 垫圈套入 XDR 塔基的最内侧根部台阶处。"
        },
        {
          "@type": "HowToStep",
          "name": "旋入 XD 飞轮并锁紧",
          "text": "使用标准花键工具将飞轮以 40 N·m 扭矩锁紧，确保无轴向旷动。"
        }
      ]
    },
    {
      "@type": "Dataset",
      "name": "Bicycle Freehub & Cassette Mechanical Compatibility Matrix",
      "description": "结构化塔基与齿根物理干涉数据集。"
    }
  ]
}
</script>
```

### 4.3 国产出海长尾搜索与全球车友流量护城河 (The Global Drivetrain MOAT)
“只有方便查得到，车友才会自发收藏和推荐”。在真实海外市场（尤其是 AliExpress、Amazon 以及 Reddit 预算车友圈 r/bikewrench）：
- 欧美大厂（DT Swiss / Hunt / Zipp）的计算器 100% 傲慢地排除了中国品牌；
- 当全球买家在 Google / Perplexity / ChatGPT 检索：
  - *“What freehub body for LTWOO eRX 12-speed?”*
  - *“Sensah Empire Pro 12s freehub compatibility”*
  - *“Do I need a spacer for Sunshine 12 speed cassette on Shimano hub?”*
  - *“Microshift Sword 10 speed cassette freehub standard”*
- **本项目是全网唯一能直接提供精准机械解答并直出实物图解的平台**！
- 爬虫在 SSR 首屏中抓取到上述问答后，AI 搜索引擎在回答车友疑问时将优先作为权威信源引用本站。车友在论坛发帖询问时，其他老车友会直接把本站链接贴上去——**“去这个网站查，连蓝图顺泰日晖都有，两秒就出结果”**，形成真正的病毒式技术口碑传播！

---

## 五、 前端交互与实物图库复用规范 (Real Photo Showcase)

车友选配自行车塔基，最核心的诉求是**与自己手头的真实零部件外观进行比对**。系统全面复用项目中已经拍摄并处理完成的高清真实产品实拍图库，坚决不搞脱离实际的抽象矢量图纸。

### 5.1 联动复用项目现有高清实拍图库
前台计算器输出结果时，直接联动渲染 [`WheelsetChooseFreehubSection.vue`](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/nuxt-i18n/app/components/WheelsetChooseFreehubSection.vue) 已经配备好的高清 WebP 实物照片，并通过 `<GuideImage :zoomOnClick="true" />` 支持买家点击无损放大查看花键槽、锁死螺纹与金属材质细节：

| 匹配塔基标准 | 项目现有官方实拍资源路径 | 实物特征指引 |
| :--- | :--- | :--- |
| **SRAM XDR** | `/public/wheelsetbuyersguide/choose freehub/sram-xdr-road-11-12-speed-freehub.webp` | 长款公路驱动体，外螺纹收窄，端面光滑。 |
| **SRAM XD** | `/public/wheelsetbuyersguide/choose freehub/sram-xd-11-12-speed-mountain-freehub.webp` | 短款山地驱动体，比 XDR 短 1.85mm。 |
| **Shimano HG-11 Road** | `/public/wheelsetbuyersguide/choose freehub/shimano-8-9-10-11-speed-road-hyper-freehub.webp` | 经典 9/10/11速公路长花键，宽槽单键防呆。 |
| **Shimano HG Mountain** | `/public/wheelsetbuyersguide/choose freehub/shimano-8-9-10-11-speed-mountain-hyper-freehub.webp` | 经典山地短花键（$34.9\text{mm}$）。 |
| **Shimano Micro Spline** | `/public/wheelsetbuyersguide/choose freehub/shimano-micro-spline-11-12-speed-mountain-freehub.webp` | 23 齿极密微型花键，前端阶梯收窄。 |
| **Campagnolo N3W** | `/public/wheelsetbuyersguide/choose freehub/Campagnolo-8-9-10-11-N3W-freehub.webp` | 缩短 4.4mm 的深槽轻量化紧凑塔基。 |
| **Campagnolo Classic** | `/public/wheelsetbuyersguide/choose freehub/Campagnolo-8-9-10-11-spd-freehub.webp` | 经典传统长深齿花键塔基。 |

### 5.2 现有两级联动极简交互无缝升级 (2-Step Direct Fitment)

现有组件 [`FreehubGroupsetHelper.vue`](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/nuxt-i18n/app/components/FreehubGroupsetHelper.vue) 本身就是极其高效的两级联动结构（并排两列选择框），升级时**完全无需破坏现有的 UI 布局与 CSS 结构**。

核心升级点在于：**在品牌列表中正式加入车友高频选购的国产与出海主流品牌，第二级直接以“飞轮齿数规格（辅以套件别名提示）”联动展示，瞬间出结果**：

#### 第一级：选择传动品牌 (Brand)
提供主流三大品牌与高频出海国产大厂（下拉框或标签分类）：
- **国际三大厂**：`Shimano` | `SRAM` | `Campagnolo`
- **国产知名大厂**：`L-TWOO (蓝图)` | `SENSAH (顺泰)` | `microSHIFT (微转)` | `SUNSHINE / ZTTO (日晖/副厂改装)`

#### 第二级：选择飞轮齿数规格 (Cassette Spec)
根据所选品牌，第二级自动过滤出该品牌下的常见飞轮齿数规格（齿数为主，括号内标注套件别名方便用户对照认领）：
- **SRAM 联动选项**：
  - `12速 10-52T / 10-50T (如 GX / X01 / XX1 Eagle / Transmission)` ➔ **SRAM XD (山地) / XDR (公路需 1.85mm 垫圈)**
  - `12速 11-50T (如 NX / SX Eagle 及第三方山地12速)` ➔ **Shimano HG 传统塔基**
  - `12速 10-28T / 10-33T / 10-36T (如 RED / Force / Rival AXS 公路)` ➔ **SRAM XDR 塔基 (无垫圈)**
  - `11速 10-42T (SRAM 11速山地 XD 飞轮)` ➔ **SRAM XD 塔基**
  - `11速 11-28T ~ 11-36T (SRAM 11速公路 PG 飞轮)` ➔ **Shimano HG 传统塔基**
- **Shimano 联动选项**：
  - `12速 10-45T / 10-51T (如 XTR / XT / SLX / Deore 山地)` ➔ **Micro Spline (微型花键)**
  - `12速 11-30T / 11-34T (如 Dura-Ace R9200 / Ultegra R8100 / 105 R7100 公路)` ➔ **Shimano HG-11 塔基 (向下兼容，无垫圈)**
  - `11速 11-25T ~ 11-34T (公路 R9100 / R8000 / R7000)` ➔ **Shimano HG-11 塔基 (无垫圈)**
  - `11速 11-40T / 11-42T / 11-46T (山地 M8000 / M7000)` ➔ **Shimano HG 塔基 (公路轮组需 1.85mm 垫圈)**
  - `10速 11/12-28T / 11-30T (高阶 CS-6700 / CS-5700)` ➔ **Shimano HG-11 塔基 (需 2.85mm 复合垫圈)**
- **Campagnolo 联动选项**：
  - `13速 9-42T / 10-44T (Ekar 碎石/公路)` ➔ **Campagnolo N3W 塔基 (无套筒)**
  - `11/12速 11-29T ~ 11-34T (Super Record / Record / Chorus)` ➔ **Campy 经典塔基，或在 N3W 上加装 AC21-N3W 延长套筒**
- **国产主流品牌联动选项**：
  - **蓝图 L-TWOO**：
    - `12速 公路 11-32T / 11-34T (eRX / RX / R9 12S)` ➔ **Shimano HG-11 塔基 (直接锁紧，无需换塔基)**
    - `12速 山地 11-50T / 11-52T (A12 12S)` ➔ **Shimano HG 塔基 (公路长塔基需 1.85mm 垫圈)**
    - `11速 公路/Gravel 11-32T / 11-42T (eR9 / GR9)` ➔ **Shimano HG-11 塔基 (直接安装)**
  - **顺泰 SENSAH**：
    - `12速 公路 11-32T / 11-34T (Empire Pro 帝国 12S)` ➔ **Shimano HG-11 塔基 (直接锁紧，无垫圈)**
    - `11/12速 Gravel 11-42T / 11-50T (SRX Pro 单盘)` ➔ **Shimano HG 塔基 (公路长塔基需 1.85mm 垫圈)**
  - **微转 microSHIFT**：
    - `10速 Gravel 宽齿比 11-48T (Sword 剑)` ➔ **Shimano HG 塔基 (拥抱通用 HG，无需 MS/XD)**
    - `10速 山地 11-48T (Advent X) / 9速 11-42T (Advent)` ➔ **Shimano HG 塔基**
  - **日晖 SUNSHINE / ZTTO 改装飞轮**：
    - `12速 公路改装 11-30T / 11-32T / 11-34T (如配 AXS / 12S 套件)` ➔ **Shimano HG-11 塔基 (免换 XDR 塔基)**
    - `12速 山地改装 11-50T / 11-52T (如配 Eagle 12S 后拨)` ➔ **Shimano HG 塔基 (免换 XD 塔基)**
    - `12速 超轻切削 9-50T / 10-52T (ZTTO SLR XD 规格)` ➔ **SRAM XD 塔基 (XDR需 1.85mm 垫圈)**

#### 第三区：结果即时呈现 (Zero-Latency Diagnostic Panel)
买家选中第二级飞轮规格的瞬间，下方即刻刷出：
1. **实物实拍照**：无缝调用项目中对应的 WebP 高清照片，买家可点击放大查看真实金属花键；
2. **标准代号**：大字显示权威塔基类型名称（如 `Shimano HG-11 (推荐)` 或 `SRAM XDR`）；
3. **垫圈指引**：显式提示是否需要垫圈（如 `直接安装，无需任何垫圈` 或 `⚠️ 需在塔基根部安装 1.85 mm 垫圈` 并附带说明）。

---

## 六、 严格明确业务边界与 Non-Goals（绝不侵入交易与订单）

必须严格对齐 [`docs/design/fitment-catalog-domain-architecture.md`](file:///c:/Users/P16V/Desktop/Github/tanzanite-theme/docs/design/fitment-catalog-domain-architecture.md) 中确立的资料库纯只读原则，**坚决杜绝任何过度设计**：

### 6.1 纯只读技术资料库与科普定位
- **定位于技术指南与选型辅助**：该计算器是一个纯粹面向买家的“传动与塔基适配速查工具”。车友在选购轮组前，通过该工具弄明白自己的套件到底匹配哪款塔基、飞轮是否需要垫圈。
- **用户自决权**：买家查完计算器后，去商品页该买哪个轮组变体（HG-11 / XDR / Micro Spline）完全由买家自己点击选择，计算器不剥夺用户的自主选择权。

### 6.2 绝对禁止的 Non-Goals（零商品、零订单、零车间耦合）
1. **禁止与商品 SKU 或变体强绑定**：计算器不读取特定商品的库存，也不自动修改商品页的选中变体，不把适配结果写入商品库。
2. **禁止侵入订单与结算链路**：计算器不生成任何“订单快照”，也不修改购物车行或结算请求；订单中买家买了什么塔基变体，完全依据买家在商品页的实际勾选。
3. **禁止臆造车间履约逻辑**：本项目为电商零售系统，不涉及出库扫码枪硬件系统、车间排产打印工单等制造执行系统（MES）范畴。计算器职责只到“给出正确答案并指导车友”为止。

---

## 七、 分阶段实施蓝图与研发核对清单

### 7.1 分阶段实施路线
- **阶段 1（后端权威知识库）**：创建 `internal/domain/fitmentcatalog/drivetrain_rules.go` 与 `drivetrain_engine.go`，收敛官方全景兼容矩阵与物理干涉校验（纯函数/领域服务）；
- **阶段 2（接口与 SSR 首屏直出）**：发布公开只读接口 `/api/v1/fitment/drivetrain/*`，Nuxt 页面在 SSR 阶段拉取并在 HTML 中直出结构化表格与 Schema.org JSON-LD；
- **阶段 3（前端实物对照组件改造）**：废弃 `FreehubGroupsetHelper.vue` 的静态硬编码数组，重构成支持齿数（10T/11T）选型、装配垫圈指引并联动现有实拍图库（`GuideImage`）的纯工具组件。

### 7.2 研发检查清单 (Engineering Checklist)
- [ ] 1. **物理阻断**：10T 飞轮强配 HG 塔基是否在后端 100% 触发 `[CRITICAL]` 物理干涉错误拦截？
- [ ] 2. **公路12速兼容**：Shimano 12速公路飞轮装在 11速 HG 塔基上是否被正确判定为“兼容且无需垫圈”？
- [ ] 3. **XDR/XD 垫圈**：XDR 塔基搭配 XD 飞轮是否明确指导“在塔基底座必须安装 1.85mm 垫圈”？
- [ ] 4. **Campy 套筒**：Campagnolo N3W 塔基搭配 11/12速飞轮是否正确提示 `AC21-N3W` 适配套筒与长锁环？
- [ ] 5. **SSR 可读性**：禁用浏览器 JavaScript 时刷新页面，能否完整看到套件与塔基兼容性全景表格？
- [ ] 6. **结构化数据**：页面 `<head>` 中是否包含有效的 `TechArticle` 与 `HowTo` 结构化 JSON-LD？
- [ ] 7. **实物对照**：选型结果是否正确联动展示项目既有的对应塔基高清实拍图（WebP），并支持点击放大核对真实花键？
- [ ] 8. **边界清爽**：计算器是否完全保持为独立只读工具，无任何商品、订单或非必要履约耦合？

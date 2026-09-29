# Phase 2：Schwalbe 商品规格查询页实施指南

> **状态**：Phase 2 第一批目录选型切片、Telemetry Guide 和页面端 SSR 分页已实现；销售商品附加层与结构化 Product/Offer 仍待后续切片。
> **页面**：`/guides/tireguides/schwalbe-tire-selector`  
> **Phase 1 数据边界**：[Phase 1 商品模板实施指南](./phase1-schwalbe-tire-system-template-implementation-guide.md)  
> **字段矩阵**：[Schwalbe 商品模板字段矩阵](./schwalbe-master-catalog-specification-matrix.md)  
> **数据基线**：迁移 359 已导入 2026-09-28 官方 sitemap 快照（773 条 live 外胎记录，15 个 404 旧 URL 排除）；迁移 360 已建立销售商品 Article No. 唯一索引；迁移 361 已为 Schwalbe 选型页建立后台可配置 FAQ 路由并预置 `en`/`zh_cn` 的胎圈问答；迁移 362 已导入官网 05/2024 胎宽—车圈内宽可能组合矩阵 13 条规则，并通过只读 API 提供。当前快照中的文本枚举基准见字段矩阵第 8 节。

> **第一批实现位置**：Nuxt 页面为 `app/pages/guides/tireguides/schwalbe-tire-selector.vue`，目录适配器为 `app/data/tireguides/schwalbeCatalog.ts`，独立筛选模型为 `app/data/tireguides/schwalbeTireCatalogFilterModel.ts`，车圈内宽规则适配器为 `app/data/tireguides/schwalbeTireRimWidthCombinationRules.ts`，查询状态在 `app/composables/useSchwalbeTireSelector.ts`，卡片和数据边界说明位于 `app/components/tireguides/schwalbe/`。该切片只读取 Phase 1 的只读目录接口，不载入旧原型数组。

## 1. 数据边界

- 全谱系型号候选保存在独立目录 `schwalbe_tire_specifications`；实际在售商品继续作为普通 Product 管理，商品规格保存到该 Product 的 `product_spec_values`。
- 每个已上架的独立 Article No. 对应一个商品和一个默认销售 SKU；不同 Article No. 或尺寸不合并为同一商品下的 SKU 变体。匹配候选不要求已上架。
- 搜索/匹配结果来自候选目录，不代表商城销售商品。默认“全谱系总览”展示目录中的全部候选；每条结果必须显示是否存在对应销售商品。
- 以 Article No. 检查候选是否对应在售商品。命中时，商品名称、19 个销售字段、链接及销售状态从实际 Product、模板值和 SKU 读取；未命中时仍展示目录候选，但不显示虚假商品购买信息。
- 文本商品字段按官网英文原文存储。商品界面翻译只通过明确的 i18n 映射处理，不修改数据或把不同语言值混成多个筛选项。
- `e_bike_rating` 保持文本语义：快照取值为 `E-25`、`E-50` 或空值；空值表示普通自行车或官网未提供评级。前端不得把它转换成 boolean，也不得由评级值推导轮圈或安全兼容结论。
- 匹配候选数据与销售 Product 是不同数据职责。Phase 1 的 19 个字段仅定义当前外胎销售商品模板；`discipline`、`series`、`construction`、`hooklessApproved`、推荐轮圈宽度等旧草案字段不能因此冒充已核实商品事实或安全结论。
- 产品页显示的 Bar/PSI 是该商品页面给出的压力值，不代表轮圈适配或 Hookless 认证。迁移 362 的胎宽—车圈内宽表只提供官方“可能组合”指导，不认证具体型号兼容性，也不替代车架间隙、TLE/TLR 或车圈厂商要求。

原型文件 `preview-schwalbe-tire-selector.html` 中的布局、“全谱系总览”入口和筛选交互可作为设计依据；其中 mock 型号值、价格、库存和 Hookless 结果不能当作生产事实。生产候选来自独立目录，并通过有官方来源的数据库导入/同步维护；不得把原型静态数组冒充已核实数据。

## 2. Phase 2 范围

1. 实现 SSR 页面 `/guides/tireguides/schwalbe-tire-selector`，以候选目录呈现全谱系总览、搜索和匹配结果。
2. 默认筛选状态为 `ALL`，显示目录里的全部型号候选，不受是否上架影响；用户可按目录中来源可核验的字段搜索和筛选。
3. 每个候选按 Article No. 查询销售 Product，并在结果中标出存在状态。匹配候选不因参与搜索或匹配而成为 Product。
4. 命中真实商品时，商品标题和 19 项销售规格来自 Product 与模板值；价格、库存及可购买状态来自现有 SKU/库存查询结果。
5. 对候选提供搜索、排序和匹配字段筛选；底层筛选模型支持文本多选、ETRTO 派生胎宽/胎圈座直径，以及由迁移 362 规则驱动的车圈内宽可能组合过滤。暂不展示未经核实的“官方兼容”“Hookless 认证”“黄金搭配”或推导出的安全压力。
6. 页面视觉沿用站点字体和组件规范，接入指南导航、FAQ 与合适的结构化数据。
7. `SchwalbeTelemetryGuide.vue` 属于 Phase 2 的独立技术说明组件，不扩充 Phase 1 的 19 个商品字段；它在候选卡片前首屏输出，使用静态 i18n 技术卡片说明 Radial 胎体、Green Marathon 材料、Schwalbe Protection Level 1–7（含 6+ Super Defense）和 ADDIX 胶料，不接收 `catalogItems` 也不读取搜索、筛选、分页或目录记录数，因此可以直接嵌入商品页或弹窗复用。

不包含通过计算器候选自动创建或上架商品、商品审核、静态原型数据导入、购物车改造和 Hookless 兼容性引擎。

## 3. 商品查询与页面数据契约

候选事实从 `schwalbe_tire_specifications` 查询，销售状态再通过 Article No. 对现有 `schwalbe_tire` Product 做只读查询。两类数据不能互相创建或替代。目录查询和商品查询可由现有 API 扩展或只读接口承载，不增加商品提交或审核 API。

商品数据附加层可采用以下结构。`facts` 的键只对应 Phase 1 的 19 个模板 slug；其值直接来自已命中的商品的规格值。`product_exists` 表示是否找到对应的销售 Product；价格和库存仅在找到商品时由其 SKU 查询。商品标题 `products.name` 不等于官方 `model_name`：后台型号选择器应建议生成可编辑标题 `Schwalbe {model_name} {etrto} ({inch_designation}) - {article_no}`，并在 `inch_designation` 为空时省略括号部分；手工修改过的标题不能被后续回填覆盖：

```ts
export interface SchwalbeOfficialFacts {
  article_no?: string
  ean?: string
  model_name?: string
  etrto?: string
  inch_designation?: string
  weight_g?: number
  version_label?: string
  compound?: string
  color?: string
  bead?: string
  e_bike_rating?: string
  epi?: number
  load_kg?: number
  seal?: string
  tread?: string
  min_pressure_bar?: number
  max_pressure_bar?: number
  min_pressure_psi?: number
  max_pressure_psi?: number
}

export interface SchwalbeSalesProductData {
  id: number
  name: string
  slug: string
  facts: SchwalbeOfficialFacts
  product_url: string
  price?: { amount_minor: number; currency: string }
  availability: 'in_stock' | 'out_of_stock' | 'unavailable'
}

export interface SchwalbeMatchResult<TCandidate> {
  article_no: string
  candidate: TCandidate
  product_exists: boolean
  sales_product?: SchwalbeSalesProductData
}
```

`candidate` 来自独立目录行，至少包含 19 项官方字段和来源信息；计算得出的匹配值与官方字段分开标识。销售商品字段不得靠标题解析或手工静态数组补齐；是否存在商品按 Article No. 查询。可选字段缺失时保持空值；`availability`、`price` 必须复用系统商品/SKU 查询结果，不得硬编码。`product_exists` 为 false 时不得输出虚构的 `sales_product`。

## 4. 选型页逻辑

搜索集合来自 `schwalbe_tire_specifications`，默认“全谱系总览”覆盖整张候选目录，不使用 `SCHWALBE_TIRE_CATALOG` 或 `rawTires` 原型常量。每条候选按 Article No. 独立检查销售商品存在状态；没有销售商品的候选不会被过滤掉，也不会获得购买链接、商品价格或库存。

搜索接口（`GET /api/v1/products/schwalbe-tire-catalog?search=<term>`）与后台型号选择器共用以下契约：`search` 是单个完整搜索词，不做多词分词或跨字段联合。服务端把同一个词分别与 Article No.、`model_name`、ETRTO、Inch 做不区分大小写的包含匹配，四个字段之间是 OR；例如 `Pro One 28-622` 不会拆成两个词去匹配名称和尺寸，不能保证命中预期记录。省略或传空 `search` 时返回全部 773 条当前候选（按 `model_name`、ETRTO、Article No. 升序）；前端应在提交搜索前提示使用一个型号、编号或尺寸词，并按真实接口结果处理空结果。

### 分页与 URL 状态

目录卡片固定每页 20 条。`page` 是从 1 开始的 URL 查询参数，缺省值为 1；`search` 仍是服务端搜索参数。提交或清空搜索、切换型号筛选或排序时都回到 `page=1`，翻页保留当前搜索词。当前 `GET /api/v1/products/schwalbe-tire-catalog` 仍按既有契约返回该搜索词命中的全量数组，以兼容后台型号选择器；Nuxt 页面在 SSR 和客户端渲染层按 URL `page` 对过滤排序后的全量结果切片，HTML 每页只输出当前 20 条卡片。分页使用原生可抓取的 `NuxtLink`，支持直链、刷新、浏览器前进后退和搜索引擎跟踪，不使用点击展开或无限滚动替代分页。后续若新增服务端分页 API，必须另行定义 `{ items, total, page, page_size, total_pages }` 契约并同步 Admin。

实际商品详情由 `schwalbe_tire` 模板和现有 Product/SKU 读取。模板 `is_filterable` 当前仅将 ETRTO、Inch、Version 标记为可筛选；该标记约束商品规格筛选，不代表计算器候选数据的筛选字段集合。Phase 2 候选筛选字段按匹配数据契约确定，不能把两套元数据混为一谈。

ETRTO 派生尺寸只用于目录筛选，不单独证明轮圈兼容性。车圈内宽筛选必须使用迁移 362 的官方可能组合规则和对应 API；同一维度内多选使用 OR，跨维度使用 AND。该范围不能推出具体型号兼容认证、Hookless 批准、车架间隙或压力上限；`seal`、`version_label`、型号名称和市场经验也不能替代这些规则。`version_label` 混合防刺结构、胎体结构和其他官方标签，当前不从它猜测数字防刺等级。

页面至少处理以下状态：

- 加载中：显示站点统一的加载状态。
- 查询失败：说明候选目录暂不可用，并允许重试。
- 候选目录为空：显示目录空状态，不回退到原型 mock；这与当前销售商品数量无关。
- 候选有数据但没有对应销售 Product：仍显示候选，并标记“未上架”。
- 候选命中真实销售商品：附加实际 Product 链接及真实 SKU 信息。
- 文本枚举本地化：优先使用字段矩阵第 8 节基于 773 条快照固化的 `compound`、`version_label`、`bead`、`e_bike_rating`、`seal` 和 `color` 值；未知新值保留英文原文并记录待翻译项。`tread` 是官网花纹编号，应按原文显示，不当作系列枚举翻译。

## 5. 页面与组件结构

```text
nuxt-i18n/app/
├── data/tireguides/schwalbeCatalog.ts       # 候选目录与销售商品响应适配器，不内嵌型号常量
├── data/tireguides/schwalbeTireCatalogFilterModel.ts # 独立多选/尺寸筛选模型
├── data/tireguides/schwalbeTireRimWidthCombinationRules.ts # 迁移 362 规则 API 适配器
├── composables/useSchwalbeTireSelector.ts   # 查询状态、搜索、排序和官方字段筛选
├── components/tireguides/schwalbe/
│   ├── SchwalbeTireSelector.vue             # 页面查询与筛选容器
│   ├── SchwalbeTireCard.vue                 # 商品字段和真实售卖信息
│   └── SchwalbeTelemetryGuide.vue            # 有官方来源支撑的技术说明
└── pages/guides/tireguides/
    └── schwalbe-tire-selector.vue            # SSR 页面与 SEO
```

保留原型“全谱系总览”入口、卡片布局和搜索交互。目录列表必须能显示所有已导入候选，即使没有销售 Product。筛选条件只使用目录中实际存储且来源可核验的字段；未经来源核对的 `discipline`、`series`、`hooklessApproved`、`minRimWidthMm`、`optimalRimWidthMm` 等原型字段不得作为官方事实或安全结论。界面使用 Tanzanite 本地字体与现有基础组件，不引入外部字体。

`SchwalbeTelemetryGuide.vue` 使用四个独立主题标签（Radial 胎体、Green Marathon、防刺等级 1–7、ADDIX 胶料）和展开/收起交互；移除重复的“技术总览”标签，默认只展示 Radial 主题，避免把四个主题重复堆叠在首屏。Radial 卡片依据 [Schwalbe Radial MTB 技术页](https://www.schwalbe.com/en/radialtires-mtb) 说明约 45° 交叉帘线与更钝、接近 90° 的排布、选择性变形、相同胎压下约 30% 的接地面积变化及滚阻取舍；不写入统一胎压、兼容性或型号性能保证。防刺卡片依据 [Schwalbe 防刺技术页](https://www.schwalbe.com/en/technology-faq/puncture-protection/) 展示 Level 7 SmartGuard / Smart DualGuard、Level 6+ Super Defense、Level 6 Double Defense、Level 5 V-Guard / GreenGuard / RaceGuard、Level 4 RaceGuard、Level 3 K-Guard、Level 2 67 EPI 和 Level 1 50 EPI；PunctureGuard 作为官方单独命名的 3 mm 入门结构说明，不强行归入数字等级。ADDIX 专题保留七条静态命名线：ADDIX Race、ADDIX 4-Season、ADDIX Speed、ADDIX Mid（原 SpeedGrip）、ADDIX Soft、ADDIX Ultra Soft 和 Endurance Compound；其中官方 ADDIX 页给出的 Speed、Mid、Soft、Ultra Soft 颜色用于对应色条，Race、4-Season 和 Endurance 的颜色仅作本指南视觉图例，不代表额外认证或统一性能等级，具体型号仍以对应官方产品页为准。Green Marathon 专题只展示面向用户的来源说明（ADDIX Eco 含回收工业炭黑和 100% 天然橡胶；GreenGuard 为 3 mm 柔性印度橡胶防刺层，部分材料来自回收来源；官方页面标注 Fair Rubber，并说明材料中 80% 为回收或可再生来源）。组件不展示快照记录数、来源字符串分组、标签计数或数据治理用语，也不把等级名称解释成跨品牌统一耐刺评分；原型中的奖项、兼容性、安全压力和绝对化宣传必须有逐条官方来源后才能加入。

## 6. SSR、SEO 与内容

- 在 SSR 阶段查询候选/匹配结果，并为每条结果附加 Article No. 对应的销售商品存在状态。
- SSR 从 URL 的 `search`、`page` 状态读取结果；首屏只输出当前页 20 条候选卡片，并输出可抓取的分页链接。分页页码不改变候选事实，也不把目录候选变成 Product。
- Telemetry Guide 在候选列表之前作为首屏语义内容输出；折叠按钮只改变视觉展开状态，不通过点击后再请求技术内容。可按实际文案使用 `TechArticle`，但不得为未核实的技术结论生成结构化数据。
- 候选/匹配结果本身不是 Product；只有实际存在并展示的销售商品才输出 Product 结构化数据，有真实报价时才输出 Offer。不要输出虚构价格、库存、认证或兼容结论。
- 通用技术说明可以声明为 `TechArticle`。目录确实完整导入并公开后，才可按实际数据描述覆盖范围；不得为 GEO 虚构数量或完整性。
- FAQ 使用现有后台 FAQ 查询与 `PageFaqSlot`，不在页面代码中复制后台 FAQ 内容。Schwalbe 选型页按 `route_path=/guides/tireguides/schwalbe-tire-selector` 命中 `page_id=guides-schwalbe-tire-selector`；只有 `faq_pages.status=active` 且对应条目 `status=published` 时，已发布问答才会在 SSR 页面显示。
- 页面标题、描述和空状态描述候选目录的实际数据状态；不得把“无销售商品”误写成“无型号候选”，也不得声称未经核验的覆盖数量。

## 7. 导航与验收

路由加入指南域导航和面包屑。验收至少覆盖：

### 7.1 第一批实现基线

已落地的目录选型切片包括：

- SSR 首屏读取 `GET /api/v1/products/schwalbe-tire-catalog`；省略搜索词时读取完整 773 条候选，提交搜索时只发送一个 `search` 参数。
- 卡片分页固定为每页 20 条；`?page=N` 的直链在 SSR 中只输出该页卡片，搜索、型号筛选和排序会重置页码，分页链接保留搜索状态并可被爬取。
- 首屏包含 Telemetry Guide 的四个主题标签、展开/收起按钮和静态来源说明；默认只输出一个主题，切换标签时替换主题内容；ADDIX 标签下保留七条彩色命名线；技术卡片不读取或显示目录计数，不复制旧 HTML 的 mock 型号或未经核实的性能结论。
- 页面提供单词搜索、型号筛选、型号名/ETRTO/Article No. 排序，并分别处理加载、接口失败、无结果和候选目录为空状态。
- 每张卡片展示目录中已保存的官方字段、来源链接、核验日期和 `product_exists` 状态；未上架候选不会被隐藏，也不会显示价格、库存或购买按钮。
- 页面提示四字段 OR 包含匹配和“不拆分多词”的接口限制，避免把 `Pro One 28-622` 当作跨字段联合查询。
- 底层筛选模型已通过独立测试验证：同一维度多选为 OR、跨维度为 AND；严格解析 ETRTO 的胎宽/胎圈座直径；`E-25`、`E-50` 与官方空值保持可区分。迁移 362 规则 API 只读返回 13 条来源带版本的可能组合范围，尚未把多选控件接入页面。
- `e_bike_rating` 继续按 `E-25`、`E-50` 或空文本显示；页面不生成 Hookless、车圈兼容或安全压力结论。
- FAQ 路由命中且当前 locale 有已发布条目时，SSR 输出后台已发布的 Schwalbe 问答（包括 WIRED、Folding 和 bead 解释）；未配置页面、未发布条目或缺少该 locale 时不输出伪造内容，选型页主体仍正常渲染。

本批接口响应目前只提供目录候选和 `product_exists`。`SchwalbeSalesProductData` 中的商品标题、链接、19 项销售快照、价格、币种和可售状态，必须在下一批后端附加层一次性按 Article No. 批量读取后再接入页面；在该附加层完成前，不能通过前端逐条请求商品、解析标题或使用静态价格补齐。

### 7.2 下一批实现顺序

1. 在后端为候选结果增加按 Article No. 批量绑定的公开销售 Product 投影，保留候选与商品两层数据职责。
2. 用现有 Product/SKU 响应生成 `sales_product`，只在真实公开商品命中时提供标题、链接、模板字段、价格和可售状态。
3. 为真实商品结果加入 Product/Offer 结构化数据，并补充 API、SSR 和无商品候选的回归测试。
4. 在附加层上线后，再把卡片的“商城商品已存在”状态升级为真实商品链接和购买信息；目录候选排序、搜索和空状态契约保持不变。

- 匹配候选不要求对应销售 Product 才能出现在结果中；
- 每条结果都显示 Article No. 是否存在于销售商品中；
- 命中商品时，19 个字段从该 Product 自己的模板规格值读取，未定义字段不会被伪造；
- 商品与 SKU 的价格、币种、库存只在真实商品命中时来自现有查询逻辑；
- 默认全谱系状态展示整个候选目录；目录为空时显示空状态，不载入 `rawTires` 或其他 mock；
- 即使所有候选都没有对应销售商品，候选仍显示并逐条标记商品不存在；
- SSR 与客户端状态一致，查询失败可重试；
- 结构化 Product/Offer 只包含已命中的真实销售商品和真实报价，匹配候选本身不伪装为 Product；
- 搜索与筛选只使用实际字段，不输出 Hookless/轮圈兼容安全结论。

## 8. 后续独立工作

迁移 362 和只读 API 已提供官方“可能组合”范围，下一步可在不复制规则的前提下把多选控件接入 SSR 页面或产品弹窗。若要升级为具体型号、Hookless、TLE/TLR 或轮圈认证判断，仍需另行核验逐型号事实、完整官方规则、轮圈制造商限制和车架间隙，再制定独立接口和测试。不能把这类结论写进 Phase 1 的 19 个官方商品字段，也不能以本文件中的旧原型示例作为依据。


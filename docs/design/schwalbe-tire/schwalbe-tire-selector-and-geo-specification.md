# Schwalbe 选型页 SEO/GEO 实施边界

> **职责**：本文只维护搜索引擎/生成式引擎收录、canonical、结构化数据和公开内容声明。字段契约由[字段矩阵](./schwalbe-master-catalog-specification-matrix.md)维护，页面行为由[Phase 2 实施指南](./phase2-standalone-page-implementation-guide.md)维护。完整文件职责见[文档索引](./README.md)。
> **页面**：`/guides/tireguides/schwalbe-tire-selector`
> **FAQ 内容**：[FAQ 内容指南](./schwalbe-faq-knowledge-base-input-guide.md)

## 1. 数据事实与公开边界

- 页面使用迁移导入的候选目录；候选存在不代表已创建销售 Product、SKU、有库存或可购买。
- 商品字段、销售价格和可售状态必须分别来自字段矩阵所述的目录或真实 Product/SKU 数据。不得用旧 HTML mock、名称推断或静态数据补齐。
- `source_url` 保存在候选表用于内部来源追溯，不出现在公开目录 API、页面卡片或 Nuxt 水合数据中。卡片可显示 `source_checked_at` 核验日期。
- 胎体、Radial、Bead、Seal、E-Bike、Color、Compound 等筛选只表示目录值匹配，不构成官方兼容认证、性能排序或安全推荐。

## 2. 抓取与 URL 收录

SSR、搜索、筛选、分页 URL 参数以及每页 20 条卡片的实现契约只在 [Phase 2 指南](./phase2-standalone-page-implementation-guide.md) 维护。本节不重复筛选字段清单。

- 无 query 的本地化选型页可索引，使用该语言 URL 的 self-canonical，并输出各语言页面和 `x-default` 的 hreflang。
- 只含 `?page=N` 且 `N >= 2` 的分页页可索引；canonical 保留规范化页码，各语言 hreflang 也保留同一页码。`?page=1` canonical 到不带 query 的本地化选型页。
- 搜索、筛选、排序、未知或格式错误的 query，以及它们与页码的组合，均使用 `noindex,follow`，canonical 指向当前语言不带 query 的选型页；对应 hreflang 指向各语言不带 query 的选型页。原生分页链接继续保留 query 状态供用户使用。
- SSR 只查询并输出当前页候选；服务端分页、水合 payload 与响应体测量以 [Phase 2 指南](./phase2-standalone-page-implementation-guide.md) 的当前实现基线为准。DOM 卡片数不能代替 payload 大小指标。

## 3. 结构化数据

- 目录候选不是 Product。只有页面真实展示了对应在售 Product 时才输出 Product 结构化数据；只有真实报价存在且已展示时才输出 Offer。
- Product 名称、Article No.、规格、价格和可售状态必须与可见页面内容一致。不得为未上架候选合成商品链接、价格、库存或 SKU。
- Telemetry Guide 可按实际技术文章内容使用 `TechArticle`；不要把静态技术主题、枚举数量或候选数量转换成 Product、Offer 或全谱系 Dataset。
- 当前不生成 FAQPage JSON-LD。以后如需增加，必须逐项对应当前 locale 下后台已发布的 FAQ 内容；FAQ 的编辑与发布流程见 [FAQ 内容指南](./schwalbe-faq-knowledge-base-input-guide.md)。

## 4. 官方技术内容与生成式摘要

- 技术说明只陈述来源能够支持的内容、条件和范围。Radial、保护结构、ADDIX 与 Green Marathon 的核验来源及页面组件内容由 [Phase 2 技术说明部分](./phase2-standalone-page-implementation-guide.md)维护。
- 生成式摘要不得把 `version_label` 直接解释为数字防刺等级，不得将 E-Bike 空值推成车型适用结论，也不得从 Bead、Seal、ETRTO 或型号名称推导轮圈兼容性、Hookless 认证或安全胎压。
- 公开选型内容不使用“覆盖完整”“官方适配”“最佳搭配”等超出实际数据和来源范围的断言。种子记录数属于数据治理状态，不作为技术卖点或产品卡片内容。
- 页面不以 Schwalbe 官方链接作为购买引导。选型页来源核验日期可以展示；官方来源链接只保存在内部来源追溯数据中。

## 5. SEO/GEO 验收

- SSR HTML 对当前状态只输出当前页的真实候选卡片，并提供可抓取分页链接；实际 API/Nuxt payload 体积另行监控。
- 无 query 与纯分页 URL 按可索引规则输出 self-canonical 和对应 hreflang；搜索、筛选、排序及无效 query 输出 `noindex,follow` 并 canonical 到本地化无 query URL。
- 未上架候选不带 Product/Offer 标记、价格、库存或购买链接；上架商品的结构化数据只引用真实 Product/SKU 字段。
- 页面技术说明、FAQ 与生成式摘要不会把目录筛选结果升级成兼容认证或普遍安全结论。

# Schwalbe 销售商品模板字段矩阵

> 本文件定义两层数据的共同字段契约：全谱系候选型号保存在独立目录 `schwalbe_tire_specifications`；实际在售型号仍是普通 Product，选择 `Schwalbe Tire` 模板后从目录选择型号并自动回填这些字段。候选目录不是销售商品库；选型结果需显示该 Article No. 是否对应真实销售 Product。
>
> 对应实施指南：[Phase 1：Schwalbe 外胎商品模板实施指南](./phase1-schwalbe-tire-system-template-implementation-guide.md)  
> 官方字段核验日期：2026-09-27。迁移 359 已将 2026-09-28 官方 sitemap 快照导入候选目录（773 条 live 外胎记录，15 个 404 旧 URL 排除）；迁移 360 已建立销售商品 Article No. 的数据库唯一边界；迁移 362 已导入官网 05/2024 胎宽—车圈内宽可能组合矩阵（13 条规则）。当前 Docker 开发库中仍没有真实销售 Schwalbe Product，候选目录和独立规则表只供选型与商品表单回填。

## 1. 官方来源

| 来源 | 用途 |
| --- | --- |
| [Green Marathon 产品页](https://www.schwalbe.com/en/Green-Marathon-11159397) | 核对产品页字段标签和展示方式 |
| [Schwalbe Radial MTB 技术页](https://www.schwalbe.com/en/radialtires-mtb) | 核对 Telemetry Guide 的 Radial 胎体结构说明 |
| [Schwalbe 防刺技术页](https://www.schwalbe.com/en/technology-faq/puncture-protection/) | 核对 Protection Level 1–7、6+ 和独立防刺结构说明 |
| [Schwalbe ETRTO 尺寸说明](https://www.schwalbe.com/en/technology-faq/tire-sizes/) | ETRTO 和英制尺寸说明 |
| [轮胎与轮圈尺寸匹配](https://www.schwalbe.com/en/technology-faq/tire-dimensions/) | 尺寸匹配说明 |
| [Hookless/TLE/TLR 要求](https://www.schwalbe.com/en/tubeless-racing-bike-requirement/) | 后续兼容功能的规则来源 |
| [胎压说明](https://www.schwalbe.com/en/technology-faq/tire-pressure/) | 产品压力范围说明 |

产品页字段是 19 项官方产品事实。FAQ 页面用于解释字段和规则，不能代替具体产品页核对真实型号值。录入商品时以该型号对应的 Schwalbe 官方产品页为准；未显示的可选值留空。

## 2. 模板字段矩阵

必填项只有 `article_no`、`model_name`、`etrto`。候选目录以 `article_no` 为唯一键；商品模板也有同名字段。商品编辑页选择目录中的具体型号后，自动回填模板字段。保存商品时，字段值再随该 Product 保存到 `product_spec_values`。

| # | 模板 slug | 商品参数类型 | 单位 | 官网标签 | 必填 | `is_filterable` | 录入规则 |
| ---: | --- | --- | --- | --- | --- | --- | --- |
| 1 | `article_no` | text | — | Article No. | 是 | 否 | 保留官网编号格式；去除首尾空格和零宽不可见字符，不补位、不改写 |
| 2 | `ean` | text | — | EAN | 否 | 否 | 不假设固定长度 |
| 3 | `model_name` | text | — | Product name | 是 | 否 | 使用官方产品页名称；它不与商城标题 `products.name` 双向自动同步，后台可按 Phase 1 规则提供一次性可编辑默认标题 |
| 4 | `etrto` | text | — | ETRTO | 是 | 是 | 使用 ASCII 连字符；保存时常见 Unicode 连字符会规范为 ASCII 连字符 |
| 5 | `inch_designation` | text | — | Inch | 否 | 是 | 保留官方原文 |
| 6 | `weight_g` | number | g | Weight | 否 | 否 | 仅填官网明确数值 |
| 7 | `version_label` | text | — | Version | 否 | 是 | 原样填写 |
| 8 | `compound` | text | — | Compound | 否 | 否 | 原样填写 |
| 9 | `color` | text | — | Color | 否 | 否 | 原样填写 |
| 10 | `bead` | text | — | Bead | 否 | 否 | 不根据型号推断 |
| 11 | `e_bike_rating` | text | — | E-Bike | 否 | 否 | 官网当前快照取值为 `E-25`、`E-50` 或空；保留评级文字，不转布尔值 |
| 12 | `epi` | number | EPI | Epi | 否 | 否 | 页面明确数字时填写；`2x67` 等双层写法取每层 EPI `67`，不把层数当作 EPI |
| 13 | `load_kg` | number | kg | Load (kg) | 否 | 否 | 只录官网数值 |
| 14 | `seal` | text | — | Seal | 否 | 否 | 原样填写 |
| 15 | `tread` | text | — | Tread | 否 | 否 | 原样填写 |
| 16 | `min_pressure_bar` | number | bar | min. Bar | 否 | 否 | 官网下限 |
| 17 | `max_pressure_bar` | number | bar | max. Bar | 否 | 否 | 官网上限 |
| 18 | `min_pressure_psi` | number | psi | min. PSI | 否 | 否 | 官网下限 |
| 19 | `max_pressure_psi` | number | psi | max. PSI | 否 | 否 | 官网上限 |

所有字段都是 `attribute`，不是变体选项。每个独立 Article No. 对应一个 Product，并配置一个默认销售 SKU；不同 Article No. 或尺寸必须使用不同商品，不能作为同一商品的 SKU 变体，否则商品级规格会被错误共用。价格、库存、SKU、销售状态和商品图片不属于这 19 个官网字段，继续填写商品表单中的对应信息。

商城标题建议由商品编辑页按 `Schwalbe {model_name} {etrto} ({inch_designation}) - {article_no}` 提供可编辑默认值；`inch_designation` 为空时省略括号部分。该建议值不是 `model_name` 与 `products.name` 的持续同步，客服手工修改后应保留手工标题。

## 3. 值口径与商品边界

具体新增和编辑步骤见 [Phase 1 实施指南](./phase1-schwalbe-tire-system-template-implementation-guide.md)。字段值按 Schwalbe 英文产品页原文保存；规格值没有 locale 列，前台本地化只能通过明确的 i18n 映射完成。

候选目录描述 Schwalbe 全谱系型号；`product_spec_values` 只描述真实 Product。两者可以包含同一型号的官网字段值，但职责不同：目录提供匹配候选和表单自动回填，商品规格值提供实际销售商品快照。匹配结果按 Article No. 显示是否存在对应销售 Product；未命中时保留候选结果，但不生成商品价格、库存或购买信息。

Phase 2 的 `SchwalbeTelemetryGuide` 不属于这 19 个字段，也不改变模板 `is_filterable` 契约。它是可在选型页、商品页或弹窗中复用的独立静态技术说明组件；Radial、Protection Level 1–7（含 6+）、ADDIX 七条命名线（Race、4-Season、Speed、Mid（原 SpeedGrip）、Soft、Ultra Soft、Endurance）和 Green Marathon 材料说明均来自已核验的官方资料或官方产品命名。ADDIX 卡片颜色只作本组件的视觉图例（Speed、Mid、Soft、Ultra Soft 的四种颜色沿用官方 ADDIX 资料；其余三种为本页视觉标记），不作为额外认证或统一性能等级，不读取目录、搜索结果或分页状态。公域不展示快照记录数、来源字符串分组或标签计数，也不把等级名称解释成跨品牌统一耐刺评分、认证、兼容性、性能或销售事实。

目录只读查询的 `search` 使用单个完整词，在 Article No.、`model_name`、ETRTO、Inch 四个字段上分别做不区分大小写的包含匹配，四个字段之间为 OR；不做多词分词或跨字段联合（例如 `Pro One 28-622` 不会拆分查询）。

选型筛选不读取商品模板的 `is_filterable`。独立模型 `schwalbeTireCatalogFilterModel.ts` 严格从 ETRTO 派生胎宽和胎圈座直径，保留 `bead`、`seal`、`e_bike_rating`、`color`、`compound` 和 `version_label` 的官方原文；同一维度多选为 OR，跨维度为 AND。页面已接入 Color 与 Compound 精确多选；Color 按完整官方字符串匹配，不拆分颜色组合或推断色系，Compound 按完整胶料原文匹配，不翻译、拆分或归并大小写。`e_bike_rating` 的空值作为“官网未标注评级”选项保留，不据此推断车型类别或适用性。车圈内宽筛选通过迁移 362 的 `schwalbe_tire_rim_width_combination_rules` 和只读 API 提供可能组合范围；范围不构成具体型号兼容认证，也不替代车架间隙、Hookless/TLE/TLR 或车圈厂商要求。`version_label` 是复合官方文本，不能直接推导数字防刺等级。

后续官方快照 upsert 只更新候选目录，不静默覆盖已保存的 `product_spec_values`。销售商品保留保存时的规格快照；重新导入后如候选字段发生变化，应按 Article No. 生成差异报告并提醒管理员复核，再由管理员明确更新商品。这样目录刷新不会在没有人工判断的情况下改变在售商品页面事实。

## 4. 字段边界

以下不是已核验的统一产品页字段，不加入本模板：`discipline`、`series`、`wheel_diameter`、`nominal_width_mm`、`construction`、`casing`、`tpi`、轮圈宽度推导值、`hookless_approved`、`max_pressure_hookless_bar`。

兼容性结论需要轮胎结构、轮圈制造商范围和官网规则共同判断，属于后续独立兼容性功能。不能由这 19 个产品参数推断“统一适用”或通用无钩圈上限。

## 5. 存储与迁移

- `product_specification_templates` 中系统模板的 slug 为 `schwalbe_tire`。
- `product_spec_definitions` 保存固定的 19 个字段定义。
- 每个已填写的商品参数值单独保存在通用表 `product_spec_values`，通过 `product_id` 关联 Product、通过 `spec_definition_id` 关联字段定义。Schwalbe 没有当前运行中的专用规格值表。
- 独立表 `schwalbe_tire_specifications` 保存候选目录中的全谱系型号事实，Article No. 为主键；同时保存 `source_url` 和 `source_checked_at` 供来源追溯，不含审批人、复核状态或审核流字段。
- `product_spec_values` 只保存真实 Product 的商品数据，不代表全谱系候选表，也不决定匹配结果是否出现。
- 迁移 358 前向重建候选目录，以保留 354–357 历史顺序；迁移 359 导入官方快照，迁移 360 为销售商品 Article No. 建立数据库唯一边界，迁移 362 建立并导入独立的官方胎宽—车圈内宽可能组合规则；不要执行 down 或回滚迁移。
- 不将原型 `rawTires`、经销商页面或未经核验的型号数值直接写入生产商品。

## 6. 模板元数据的实际配置

迁移 355 为这 19 个定义设置了以下共同值：

| 列 | 配置值 | 说明 |
| --- | --- | --- |
| `group` | `Official product facts` | 后台字段分组 |
| `presentation` | `text` | 展示方式；输入类型仍由 `field_type` 决定 |
| `is_visible` | `TRUE` | 字段可见 |
| `role` | `attribute` | 商品级参数，不是 SKU 变体或定制选项 |
| `selection_mode` | `single` | 单值字段 |
| `min_selections` | `0` | 单值属性的选择下限元数据 |
| `max_selections` | `NULL` | 不配置多选数量上限 |
| `validation` | 空字符串 | 本模板未用该列保存业务校验规则 |
| `option_items` | 无 | 当前 19 个定义不配置枚举候选值 |

`is_filterable` 仅 ETRTO、Inch、Version 为 `TRUE`；其它 16 项为 `FALSE`。该标志属于销售商品模板定义。计算器候选数据的搜索字段和筛选规则由匹配数据契约决定，不能从这组商品模板元数据推断。

## 7. 当前数据库结构与约束

| 表 | 当前相关列 | 关联、唯一约束和用途 |
| --- | --- | --- |
| `product_specification_templates` | `id BIGSERIAL PK`、`name VARCHAR(120)`、`slug VARCHAR(120) UNIQUE NOT NULL`、`description TEXT`、`sort_order INTEGER`、`is_enabled BOOLEAN`、`is_system_managed BOOLEAN`、`revision INTEGER`、`created_at/updated_at TIMESTAMP` | `schwalbe_tire` 是系统模板 slug；每个模板拥有字段定义 |
| `products` | `product_specification_template_id BIGINT NULL` | 外键指向模板；`products.name` 是商城标题，不等同于官方 `model_name` |
| `product_spec_definitions` | `id BIGSERIAL PK`、`product_specification_template_id BIGINT NOT NULL`、`group VARCHAR(80)`、`name/slug VARCHAR(120)`、`field_type/presentation VARCHAR(32)`、`unit VARCHAR(32)`、`is_required/is_filterable/is_visible BOOLEAN`、`role VARCHAR(24)`、`selection_mode VARCHAR(16)`、`min_selections INTEGER`、`max_selections INTEGER NULL`、`sort_order INTEGER`、`validation TEXT` | 外键指向模板且删除模板时级联；`(product_specification_template_id, slug)` 唯一；role、selection mode 和选择范围有 CHECK 约束 |
| `product_spec_values` | `id BIGSERIAL PK`、`product_id BIGINT NOT NULL`、`spec_definition_id BIGINT NOT NULL`、`value TEXT NOT NULL`、`created_at/updated_at TIMESTAMP` | 两个外键分别指向 Product 和定义，删除父记录时级联；`(product_id, spec_definition_id)` 唯一，一商品一字段最多一行；迁移 360 对 Schwalbe `article_no` 以 `lower(btrim(value))` 建立跨商品唯一索引，关闭并发重复写入窗口 |
| `schwalbe_tire_specifications` | `article_no VARCHAR(32) PK`、19 项官网事实列、`source_url TEXT NOT NULL`、`source_checked_at DATE NOT NULL`、时间戳 | 独立候选目录；Article No. 全局唯一；压力 min/max 与正数测量值有 CHECK；索引覆盖 ETRTO、Inch 和 Version；不关联 Product、不表达销售状态、不含审批列 |
| `schwalbe_tire_rim_width_combination_rules` | `id BIGSERIAL PK`、`tire_width_min_mm/max_mm INTEGER`、`inner_rim_width_min_mm/max_mm INTEGER`、`source_basis TEXT`、`source_version VARCHAR(64)`、`source_url TEXT`、`source_checked_at DATE`、时间戳 | 迁移 362 的官方可能组合指导；四个范围端点为正数且 min 不得大于 max，四端点组合唯一；不关联 Product，不认证具体型号兼容性，不替代车架间隙或车圈厂商要求 |

候选目录 19 个事实列的实际 SQL 类型如下；它们与上方模板 `slug` 一一对应：

| slug | SQL 列类型 | NULL 规则 |
| --- | --- | --- |
| `article_no` | `VARCHAR(32)` | 主键，不可空 |
| `ean` | `VARCHAR(32)` | 可空 |
| `model_name` | `VARCHAR(180)` | 不可空 |
| `etrto` | `VARCHAR(32)` | 不可空 |
| `inch_designation` | `VARCHAR(32)` | 可空 |
| `weight_g` | `NUMERIC(7,2)` | 可空，填写时必须大于 0 |
| `version_label` | `VARCHAR(120)` | 可空 |
| `compound` | `VARCHAR(120)` | 可空 |
| `color` | `VARCHAR(120)` | 可空 |
| `bead` | `VARCHAR(64)` | 可空 |
| `e_bike_rating` | `VARCHAR(64)` | 可空 |
| `epi` | `INTEGER` | 可空，填写时必须大于 0 |
| `load_kg` | `NUMERIC(7,2)` | 可空，填写时必须大于 0 |
| `seal` | `VARCHAR(120)` | 可空 |
| `tread` | `VARCHAR(120)` | 可空 |
| `min_pressure_bar`, `max_pressure_bar` | `NUMERIC(5,2)` | 可空，填写时必须大于 0；min 不得大于 max |
| `min_pressure_psi`, `max_pressure_psi` | `NUMERIC(6,2)` | 可空，填写时必须大于 0；min 不得大于 max |

目录附加列：`source_url TEXT NOT NULL`、`source_checked_at DATE NOT NULL`、`created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`、`updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`。目录不设置 `verification_status`、`reviewed_by_id`、`source_submitted_by_id` 等审批字段。

值关联关系：

```text
products.product_specification_template_id
  -> product_specification_templates.id (slug = 'schwalbe_tire')

product_spec_values.product_id -> products.id
product_spec_values.spec_definition_id -> product_spec_definitions.id
product_spec_definitions.product_specification_template_id -> product_specification_templates.id

schwalbe_tire_specifications.article_no -> standalone catalog key (no Product foreign key)
```

数据库对 `product_spec_values.value` 只施加 `TEXT NOT NULL`，所以 number 字段以数字文本保存；迁移 360 已为 Schwalbe `article_no` 增加 `lower(btrim(value))` 跨 Product 唯一索引，并在建索引前拒绝历史重复值。目录表自身以 Article No. 主键防止目录重号。应用层预检查用于返回可读错误，数据库唯一索引负责并发安全。必填、数字和商品级跨字段规则由商品服务校验。`product_spec_definitions.validation` 当前为空，不要误认为校验规则已写入表元数据。

## 8. 2026-09-28 快照枚举基准

以下集合是迁移 359 导入的 773 条 live 记录中按原文去重得到的值，供 Phase 2 本地化字典和筛选展示使用。它们是这次快照的已观测枚举基准，不代表 Schwalbe 未来新增型号时不会出现新值；导入新快照后应重新计算并审阅差异。除 `e_bike_rating` 外，本批记录在这些字段上均有值；空值仍需保留为空。

| 字段 | 快照中的有效值 |
| --- | --- |
| `compound` | `ADDIX`、`ADDIX 365`、`ADDIX 4-Season`、`ADDIX E`、`ADDIX Eco`、`ADDIX Green`、`ADDIX Race`、`ADDIX Soft`、`ADDIX Speed`、`ADDIX SpeedGrip`、`ADDIX Ultra Soft`、`Black'n'Roll`、`Endurance`、`GRC`、`Green Compound`、`MID`、`SBC`、`SOFT`、`SPEED`、`Silica`、`ULTRA SOFT`、`WheelStar`、`Winter` |
| `version_label` | `DD, GreenGuard`、`DD, GreenGuard, Radial`、`DD, RaceGuard`、`DD, RaceGuard, Radial`、`Evolution`、`GRAVITY`、`GRAVITY PRO`、`GRAVITY PRO, Radial`、`GreenGuard`、`K-Guard`、`PRO, DD, RaceGuard`、`PRO, DD, V-Guard`、`PRO, V-Guard`、`Performance`、`PunctureGuard`、`RACE PRO`、`RACE PRO, RaceGuard`、`RACE PRO, V-Guard`、`RACE, RaceGuard`、`RaceGuard`、`Reinforced`、`Reinforced, RaceGuard`、`SCHWALBE`、`Smart DualGuard`、`SmartGuard`、`Super Defense`、`Super Downhill`、`Super Ground`、`Super Race`、`Super Race, RaceGuard`、`Super Race, V-Guard`、`Super Trail`、`TRAIL`、`TRAIL PRO`、`TRAIL PRO, Radial`、`V-Guard`、`XC PRO` |
| `bead` | `Folding`、`WIRED` |
| `e_bike_rating` | `E-25`、`E-50`；另有官网未标注评级的 NULL 值，不据此推断车型类别 |
| `seal` | `TLE`、`TLR`、`Tube` |
| `color` | `Black`、`Black+BlackReflex`、`Black+Reflex`、`Black/Coffee+Reflex`、`Blue Stripes`、`Bronze`、`Bronze Sidewall`、`Bronze+Reflex`、`Brown+Reflex`、`Brown/Whitewall+Reflex`、`Classic`、`Creme+Reflex`、`Grey Stripes`、`Grey/Black`、`Gumwall`、`Red Stripes`、`Transparent Sidewall`、`White Stripes`、`White/Bordeaux`、`Whitewall`、`Whitewall+Reflex` |
| `tread` | `HS342`、`HS371`、`HS371A`、`HS375`、`HS379`、`HS385`、`HS387`、`HS396`、`HS417`、`HS425`、`HS429`、`HS431`、`HS438`、`HS439`、`HS440`、`HS442`、`HS447`、`HS447B`、`HS451`、`HS462`、`HS462A`、`HS462B`、`HS463`、`HS464`、`HS466`、`HS468`、`HS471`、`HS472`、`HS473`、`HS475`、`HS483`、`HS484`、`HS489`、`HS490`、`HS492`、`HS493`、`HS493A`、`HS493D`、`HS497`、`HS498`、`HS499`、`HS600`、`HS602`、`HS604`、`HS605`、`HS608`、`HS609`、`HS610`、`HS611`、`HS612`、`HS613`、`HS614`、`HS617`、`HS618`、`HS619`、`HS620`、`HS621`、`HS622`、`HS624`、`HS625`、`HS626`、`HS630`、`HS632`、`HS634`、`HS635`、`HS636`、`HS637`、`HS638`、`HS639`、`HS641`、`HS642`、`HS643`、`HS646`、`HS647`、`HS648`、`HS651` |

`tread` 是官网的花纹编号（本快照 76 个 `HS...` 值），不应当当作可翻译的产品系列；前端应按完整原文显示或建立逐值字典。所有枚举本地化只改变显示文本，查询和存储仍使用上述英文原值。

这批 773 条快照为选型卡片提供当前枚举基线。页面保留每条记录的 `Radial`、保护技术 Version 和 `compound` 原文，并按每页 20 条输出；这些原文不再转换为公域计数或统一防刺等级。Telemetry Guide 使用独立静态技术卡片，不从这批记录计算内容。分页不会改变字段矩阵。

# Phase 1：Schwalbe 外胎商品模板实施指南

> **数据边界**：全谱系 Schwalbe 型号先保存在独立候选目录 `schwalbe_tire_specifications`，供选型器检索，也供商品编辑页自动回填。它不代表销售商品，不通过后台人工逐条录入或审批。实际销售仍是普通 Product：选择 `Schwalbe Tire` 模板，从目录选择具体型号并自动填充 19 项官网字段，再维护销售 SKU 并保存到现有商品表。  
> **官方字段核验日期**：2026-09-27  
> **文档更新日期**：2026-09-29  
> **文档索引**：[Schwalbe 文档职责与权威范围](./README.md)
> **迁移与部署操作**：Schwalbe 迁移从 354 起追加；执行顺序、前置检查和各迁移职责见[第 6 节](#6-迁移规则)，不回滚既有迁移。
> **字段矩阵**：[Schwalbe 商品模板字段矩阵](./schwalbe-master-catalog-specification-matrix.md)  
> **Phase 2 实施指南**：[Phase 2：Schwalbe 商品规格查询页实施指南](./phase2-standalone-page-implementation-guide.md)
> **Phase 2 页面技术说明**：Telemetry Guide 与页面端 SSR 分页契约见 [Phase 2 实施指南](./phase2-standalone-page-implementation-guide.md)；技术说明不是 19 个商品字段或销售事实。
> **种子快照基线**：导入快照范围、记录条数和枚举统一见[字段矩阵第 8 节](./schwalbe-master-catalog-specification-matrix.md#8-2026-09-28-快照枚举基准)。

## 1. 设计结论

1. `schwalbe_tire_specifications` 是全谱系候选目录，按 Article No. 唯一保存官网型号事实；它既不是 Product，也不表示该型号在售。
2. `schwalbe_tire` 商品模板定义固定的 19 个字段。选择模板后，后台从候选目录加载可选型号；选择 Article No./型号后自动填充模板字段，不要求客服重复抄录官网资料。
3. 保存实际销售商品时，商品服务仍使用常规 `specs` 路径，将所选型号的 19 项值保存到该 Product 的 `product_spec_values`。商品自己的价格、库存、SKU、图片和销售状态由现有 Product/SKU 字段维护。
4. 选型器始终可以检索候选目录中的型号，并为每个候选按 Article No. 查询是否存在真实销售 Product。目录候选本身不会自动变成 Product。
5. 候选目录通过有官方来源的数据库导入/同步维护，不新增人工录入台、Admin 提交流程、审批或二人复核。
6. 不修改或回滚迁移 1–353。Schwalbe 迁移的执行顺序、用途和发布前检查统一见[第 6 节](#6-迁移规则)；迁移不通过 down 操作删除既有目录数据。

## 2. 官网字段边界

核对来源：

| 来源 | 用途 |
| --- | --- |
| [Green Marathon 产品页](https://www.schwalbe.com/en/Green-Marathon-11159397) | 核对产品页字段名称及展示方式 |
| [Schwalbe ETRTO 尺寸说明](https://www.schwalbe.com/en/technology-faq/tire-sizes/) | 理解 ETRTO 和英制尺寸 |
| [轮胎与轮圈尺寸匹配](https://www.schwalbe.com/en/technology-faq/tire-dimensions/) | 轮胎与轮圈尺寸规则 |
| [Hookless/TLE/TLR 要求](https://www.schwalbe.com/en/tubeless-racing-bike-requirement/) | 无钩圈、TLE/TLR 与厂商上限规则 |
| [胎压说明](https://www.schwalbe.com/en/technology-faq/tire-pressure/) | 产品压力值的含义 |

产品页展示字段为 Article No.、EAN、Product name（产品名称/页面标题）、ETRTO、Inch、Weight、Version、Compound、Color、Bead、E-Bike、Epi、Load (kg)、Seal、Tread、min. Bar、max. Bar、min. PSI、max. PSI。模板统一将 `Psi` 的大小写规范为 `PSI`，不改变字段含义。

以下不是已确认的统一产品页字段，不属于这个模板：discipline、series、wheel_diameter、nominal_width_mm、construction、casing、tpi、protection_level、轮圈宽度范围、`hookless_approved` 和 `max_pressure_hookless_bar`。不要用推测值补齐这些字段，也不要将它们藏入 metadata。

## 3. 字段录入口径

完整的 19 字段契约（slug、类型、单位、官网标签、必填项和 `is_filterable`）只在[字段矩阵第 2 节](./schwalbe-master-catalog-specification-matrix.md#2-模板字段矩阵)维护。本指南负责录入和保存流程；必填字段只有 Article No.、Product name 和 ETRTO。官网没有展示或无法从具体官方产品页确认的可选字段留空，不用猜测值、零或空字符串伪造数据。

录入口径：

- Article No. 是该具体售卖规格的标识。每个独立 Article No. 建一个独立商品，并为该商品配置一个默认销售 SKU；不同 Article No. 或不同尺寸必须建成不同商品，不能作为同一 Schwalbe 商品下的多 SKU 变体。型号事实由候选目录提供，选择具体 Article No. 后由商品表单自动回填。
- Article No. 去除首尾空格及复制带入的零宽不可见字符；保留官网的数字、大小写和标点，不补零、不改写编号。
- ETRTO 使用标准 ASCII 连字符（半角减号）。保存时会把常见 Unicode 连字符规范为 ASCII 减号，方便搜索和后续解析。
- 文本规格按 Schwalbe 英文产品页原文入库，包括产品名称、Version、Compound、Color、Bead、E-Bike、Seal 和 Tread。不要录入端中英混写或自行翻译。规格值没有单独的语言列；前台如需本地化展示，应通过明确的 i18n 字典映射，未知值保留英文原文。
- `e_bike_rating` 是官网 E-Bike 文本评级，当前快照出现 `E-25`、`E-50` 和空值。保留官网文本，禁止改为 boolean 开关；空值表示源产品页没有提供评级，不应被转换成 `false`，也不用于推断车型类别或兼容结论。
- 数值字段留空表示官网没有提供该值；不要用 `0` 代替未知值。已填写的 Weight、EPI、Load 和胎压值必须是有限的正数（必须大于 0）；EPI 还必须为整数。官网把部分双层胎体写成 `2x67` 时，目录数值保存每层 EPI `67`，不能把层数 `2` 当作 EPI。若同时填写同一单位的 min/max pressure，min 不得大于 max。候选表约束和商品保存服务都按 `<= 0` 拒绝这些数值。
- Article No. 在创建或编辑 Schwalbe 商品时由商品服务预检查，拒绝与其他商品重复的规范化编号；迁移 360 已在 `product_spec_values` 上建立大小写/首尾空白不敏感的数据库唯一索引，作为并发写入的最终边界。若并发请求撞上该索引，保存会失败并回滚，不能产生多个商品共用一个 Article No.

## 4. 商品录入和保存

**商品与 SKU 粒度：1 个 Article No. 对应 1 个 Product，并配置 1 个默认销售 SKU。** 19 个官网规格是商品级属性；不能把 Pro One TLE 等系列名称当作一个多尺寸商品，再让不同尺寸 SKU 共用同一份 Article No.、ETRTO、Weight 等规格。每遇到不同 Article No. 或尺寸，分别新增商品并填写对应规格。

1. 打开后台商品新增页，在“商品规格模板”中选择 `Schwalbe Tire`。
2. 商品表单读取候选目录并显示可选型号；选择具体 Article No. 后，系统自动填充 19 个官网字段。
3. 检查回填的型号事实。建议型号选择后为 `products.name` 自动提供默认标题：`Schwalbe {model_name} {etrto} ({inch_designation}) - {article_no}`，例如 `Schwalbe Kojak 35-559 (26x1.35) - 11100062.02`；`inch_designation` 为空时省略括号部分。默认标题应可编辑，且后续更换型号时只更新仍由系统生成的标题，不覆盖客服手工修改过的标题。商城标题与官网 `model_name` 分开维护，不重复录入其余规格字段。
4. 填写该商品自己的默认 SKU、价格、库存、图片等销售信息。
5. 保存商品。模板 ID 与回填的 19 个参数随商品请求提交；后端通过常规商品创建/编辑服务校验并保存到商品自己的 `product_spec_values`。

编辑已有商品时，页面从该商品的 `product_spec_values` 载入参数；修改后仍保存回这件商品。Article No.、EAN 等字段值可随商品直接编辑，不需要先改另一份物料记录。

商品保存沿用现有接口：`POST /api/admin/products` 创建，`PUT /api/admin/products/:id` 编辑。`product_specification_template_id` 选择模板，`specs` 携带自动回填后的商品字段。后台型号选择器使用只读接口 `GET /api/v1/products/schwalbe-tire-catalog`，支持可选的 `search` 参数；搜索字段、匹配方式和返回顺序与 Phase 2 选型页共用，具体契约见 [Phase 2 指南](./phase2-standalone-page-implementation-guide.md)。官方可能组合规则另由 `GET /api/v1/products/schwalbe-tire-rim-width-combination-rules` 只读提供；不新增目录人工提交、审核或复核接口。

商品模板本身没有 `source_url`、导入批次等目录来源字段，因为它只承载商品的 19 项事实。来源 URL 和核对日期保存在候选目录行；`source_url` 仅供内部来源追溯，不进入公开目录 API 或选型页水合数据。销售文案、物流、供应商等信息继续使用各自现有商品模块。

## 5. 系统模板约束

模板 slug 为 `schwalbe_tire`，由迁移 355 创建并设置 `is_system_managed = TRUE`。模板字段集合与第 3 节一致，字段结构受系统模板保护，避免删除、改名或改变字段类型后造成已有商品参数失配。模板不能删除；模板字段值仍然属于各自商品。

模板字段结构仍由 `product_spec_definitions` 定义；具体型号与字段值来自 `schwalbe_tire_specifications`。后台的型号选择器按 Article No. 读取目录行，并将 19 个字段回填至现有表单。目录型号不是 `product_spec_definitions.option_items`，也不是 SKU 变体；不要恢复已移除的 `options` 或 `is_variant_option` 旧列。

模板元数据 `is_filterable` 仅将 `etrto`、`inch_designation` 和 `version_label` 标记为 `TRUE`，其余 16 项标记为 `FALSE`。该标志属于销售商品模板定义，只约束商品规格筛选，不约束候选目录的选型查询；Phase 2 的搜索和筛选按候选目录数据契约执行。

后台 number 输入按字段精度设置步长：EPI（`epi`）必须严格为整数，使用 `step=1`；`weight_g`、`load_kg` 支持最多两位小数，可使用 `step=0.01` 或 `step="any"`；四个胎压字段也必须支持小数，bar 默认使用 `step=0.1`，PSI 可使用 `step=0.01` 或 `step="any"`。这些字段不能用 `step=1` 限制。服务端仍按 number 类型校验并保留输入的小数值。后台 EPI 输入框提示“双层胎体如官网 `2x67` 只填 `67`”；表单不接受把 `2x67` 原文粘贴进 number 字段。

## 6. 迁移规则

- 迁移 1–353 是既有业务基线，不修改、不重排、不执行 down。
- 迁移 354 创建独立候选目录的初版结构；356 是不再采用的审核方案，不构成运行时审核流程。
- 迁移 355 创建系统模板与 19 个字段定义。
- 迁移 357 是已存在的历史前向变更：它只在旧目录为空时删除该表；若表非空会主动失败，不会静默删数据。
- 迁移 358 在 357 之后重建无审批字段的候选目录，字段包含 19 项官网事实、Article No. 主键和来源 URL/核对日期。这样保留既有迁移记录，不回滚，也不放弃全谱系数据职责。
- 迁移 359 使用 [Schwalbe 目录导入脚本](../../../scripts/import-schwalbe-catalog.mjs) 生成并幂等 upsert 官方英文 sitemap 快照；它只写经核验的 live 产品页，不创建 Product、SKU、价格或库存。重新抓取时应生成新的带核对日期的 seed；当前 seed 的来源、范围和记录数见[字段矩阵第 8 节](./schwalbe-master-catalog-specification-matrix.md#8-2026-09-28-快照枚举基准)。
- 迁移 360 为 `product_spec_values` 上的 Schwalbe `article_no` 建立大小写/首尾空白不敏感的数据库唯一索引，并在建索引前拒绝已有重复值；应用层预检查只负责更早返回可读错误，不能替代该并发安全边界。
- 迁移 361 为 `guides-schwalbe-tire-selector` 建立精确路由 `/guides/tireguides/schwalbe-tire-selector` 的 FAQ 页面，页面元信息覆盖当前支持的 locale，初始 `en`/`zh_cn` 问答解释 `WIRED`、`Folding` 和 bead；其他语言由后台维护翻译，FAQ 不回填商品字段。
- 迁移 362 建立 `schwalbe_tire_rim_width_combination_rules`，导入官网 05/2024 矩阵的 13 条可能组合范围；它不认证具体型号兼容性，不替代车架间隙判断，Hookless/straight-side 仍须满足 TLE/TLR 与车圈厂商要求。规则通过 `GET /api/v1/products/schwalbe-tire-rim-width-combination-rules` 只读提供。
- 发布前按 [`go-backend/DEPLOYMENT.md`](../../../go-backend/DEPLOYMENT.md#schwalbe-migrations-357-362-preflight) 检查每个环境的旧表行数；非空时先检查并保留数据，不能清空后继续。

## 7. 当前实现验收与后续工作

Phase 1 路径已实现：商品模板选择、19 个模板字段、商品新增/编辑的 `specs` 保存、目录查询、后台型号选择、选中型号后的字段回填，以及目录候选的 Article No. 商品存在标记。导入脚本负责校验必填字段、正数测量值、EPI 整数约束和 min/max 顺序，并可重复生成新的 JSON 与 SQL seed。目标环境的迁移版本、行数和约束状态应在部署时逐环境核验；本机 Docker 的结果不能代表其他环境。当前种子快照的数据范围和记录数以[字段矩阵第 8 节](./schwalbe-master-catalog-specification-matrix.md#8-2026-09-28-快照枚举基准)为准，页面选型功能状态以 [Phase 2 指南](./phase2-standalone-page-implementation-guide.md)为准。

型号选择后的商城标题建议生成规则尚未接入后台表单；实现时按第 4 节生成可编辑默认值，并避免覆盖手工改过的标题。

候选目录数据与在售 Product 是两件事。目录数据通过来源明确的导入/同步任务写库；实际销售 Product 仍由后台按第 4 节正常保存。导入目录不创建商品、SKU、价格或库存，也不需要额外 Admin 提交或第二位管理员复核。后续 seed 的 upsert 只更新候选目录，不自动覆盖已保存的 `product_spec_values`；销售商品保留保存时的规格快照。官网字段发生变化时，应按 Article No. 对比候选目录与销售商品并提醒管理员复核，再由管理员明确提交商品更新，避免静默改写在售资料。

后续刷新 seed 前，先生成官方页 JSON 快照，再对目标数据库执行只读审查。以下命令从仓库根目录运行：

```powershell
New-Item -ItemType Directory -Force .tmp | Out-Null
node scripts/import-schwalbe-catalog.mjs `
  --output-json .tmp/schwalbe-catalog-next.json
Push-Location go-backend
go run ./cmd/adminctl audit-schwalbe-catalog-refresh `
  -snapshot ../.tmp/schwalbe-catalog-next.json `
  -format json | Set-Content -Encoding utf8 ../.tmp/schwalbe-catalog-review.json
Pop-Location
```

审查命令使用 `-config` 指定的配置文件，或数据库环境变量与开发默认值；它会开启 PostgreSQL `READ ONLY`、`REPEATABLE READ` 事务，不执行迁移、不写入目录、不修改销售商品。报告列出新增 Article No.、当前目录中未出现在新快照里的编号、规格字段变化、来源 URL/核对日期变化，以及在售商品相对当前目录和新快照的字段差异。商品范围是未软删除、状态为 `active` 且至少有一个启用 SKU 的 Schwalbe 商品；库存为零仍纳入。商品差异会标明是否需要人工复核。

`not_in_incoming_snapshot` 只表示当前目录 Article No. 不在本次成功抓取的 JSON 中；可能是官网 sitemap 移除，也可能是产品页返回 404，不能据此自动删除候选记录。先人工核对官网及报告，再生成后续迁移 seed；来源 URL 和核对日期变化单独列出，不算商品规格变化。应用目录更新后，销售商品仍保留原快照，任何在售规格修订都由管理员明确确认并保存。

迁移 362 的胎宽—车圈内宽表只提供 Schwalbe/ETRTO 的可能组合指导；它不是 Hookless/TLE/TLR 兼容性引擎，也不代表具体型号认证。产品页压力字段不代表轮圈适配结论，当前不得推断统一 hookless 标记或通用压力上限。

## 8. Phase 2 数据来源约定

本阶段只定义商品模板和销售 Product 的数据边界。选型页如何读取候选目录、管理筛选、URL、SSR 分页及展示独立 Telemetry Guide，统一见 [Phase 2 实施指南](./phase2-standalone-page-implementation-guide.md)。字段及模板元数据以[字段矩阵](./schwalbe-master-catalog-specification-matrix.md)为准；不要在本节复制页面行为清单。


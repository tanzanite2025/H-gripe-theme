# 车型前叉花鼓适配资料库设计

最后更新：2026-08-26

## 1. 决策

车型、前叉和花鼓规格组成一个完全独立的适配资料库域。

这个域不是商品域的附加字段，也不是商品适配配置。它维护一套可被多个页面查询的全局技术资料：

```text
车型前叉花鼓适配资料库
  ├── 车架 / 车型资料
  ├── 前叉资料
  └── 花鼓规格资料
```

首版的核心关系是：

```text
车架 / 车型记录  N:M  花鼓规格
前叉记录        N:M  花鼓规格
```

车架和前叉之间首版不建立直接配对关系。用户可以分别查询车架或前叉，并查看各自适配的花鼓规格。

这个域不与任何商品、SKU、商品变体或商品翻译行建立关系。商品选择页面中的问号只是打开资料查询入口，不会保存查询结果，也不会修改商品。

推荐内部域名：

```text
fitmentcatalog
```

不要使用 `productcompatibility`、`shippingfitment` 或其他会暗示商品关系、物流关系的名称。

## 2. 目标

本功能解决以下问题：

- 在后台集中录入车架/车型的品牌、型号、年份和适配花鼓规格；
- 在后台集中录入前叉的品牌、型号、年份和适配花鼓规格；
- 用一个独立 TAB 维护可重复使用的花鼓规格；
- 避免不同记录中重复手写、拼写不一致的花鼓规格；
- 让商品选择、帮助按钮或其他用户入口可以只读搜索这套资料；
- 让资料库的生命周期、权限、接口和数据校验独立于商品域。

## 3. 明确边界

### 3.1 适配资料库负责

适配资料库负责：

- 车架/车型资料；
- 前叉资料；
- 花鼓规格主数据；
- 车架与花鼓规格的关联；
- 前叉与花鼓规格的关联；
- 后台维护、启用、停用和查询；
- 面向用户的只读搜索结果。

### 3.2 禁止进入商品域的内容

本功能不得增加或修改：

- `products` 表；
- `product_variants` 表；
- 商品规格模板；
- 商品信息模板；
- 商品保存请求；
- 商品保存事务；
- 商品 SKU；
- 商品详情数据；
- 商品包装配置；
- Shipping/物流域；
- 商品缓存、商品事件或商品搜索索引。

适配表不得包含：

- `product_id`；
- `product_variant_id`；
- `sku`；
- 指向商品表的外键；
- 因商品删除、改名或翻译而自动变化的逻辑。

即使查询入口出现在花鼓商品选择页面，请求也不应携带商品 ID，也不应在数据库保存“该商品查询过哪些适配资料”。

### 3.3 本阶段不负责

本阶段不做：

- 商品与适配资料绑定；
- 商品自动推荐；
- 根据商品规格自动计算适配结果；
- 车架与前叉的直接配对规则；
- 完整机械兼容性计算；
- 轮组适配问卷；
- 物流、包装或运输规则；
- 供应商、库存或采购资料；
- 产品品牌主数据复用；
- 商品多语言翻译模型。

首版的“适配”含义是：后台明确维护的车架/前叉记录与花鼓规格之间的资料关联，不根据未录入的技术参数自动推断。

## 4. 后台信息架构

后台提供一个独立入口，建议显示名称为：

```text
车型前叉花鼓适配库
```

页面内部包含三个独立 TAB：

```text
车型前叉花鼓适配库
├── 车架 / 车型
├── 前叉
└── 花鼓规格
```

三个 TAB 共享页面外壳，但每个 TAB 必须拥有自己的：

- 列表状态；
- 搜索和筛选状态；
- 表格列；
- 新增/编辑表单；
- 校验逻辑；
- API 请求；
- 类型和响应结构；
- 启用/停用逻辑；
- 删除逻辑。

不做一个同时承载三种记录的“大适配表单”，也不做一个使用 `type = frame/fork/hub` 的混合编辑器。

后台 Admin 当前只维护中英文两套界面文案：

```text
zh-CN
en-US
```

前台 Nuxt 的 20 语言配置属于 storefront 用户侧展示层，不属于本后台资料录入阶段。后台新增 TAB、表单、提示和校验文案不得复制到 Nuxt 20 语言目录，也不得把 Admin 扩成 20 语言。

### 4.1 车架 / 车型 TAB

列表至少显示：

- 车架品牌；
- 车型/型号；
- 系列或代际；
- 年份；
- 适配花鼓规格数量；
- 启用状态；
- 更新时间；
- 操作。

表单录入：

- 车架品牌；
- 车型名称或型号；
- 系列；
- 代际；
- 年份模式；
- 年份；
- 市场/地区；
- 适配花鼓规格，多选；
- 备注；
- 是否启用；
- 排序。

适配花鼓规格必须从“花鼓规格 TAB”已启用的资料中选择，不允许在车架表单中自由填写规格名称。

按自行车常规规则，车架默认只能选择 `rear` 后花鼓规格。若业务存在特殊车型，再单独扩展规则，不在首版静默放开。

### 4.2 前叉 TAB

列表至少显示：

- 前叉品牌；
- 前叉型号；
- 系列或代际；
- 年份；
- 适配花鼓规格数量；
- 启用状态；
- 更新时间；
- 操作。

表单录入：

- 前叉品牌；
- 前叉型号；
- 系列；
- 代际；
- 年份模式；
- 年份；
- 市场/地区；
- 适配花鼓规格，多选；
- 备注；
- 是否启用；
- 排序。

适配花鼓规格同样只能从“花鼓规格 TAB”选择。

按自行车常规规则，前叉默认只能选择 `front` 前花鼓规格。

### 4.3 花鼓规格 TAB

花鼓规格是整个资料库的可复用主数据。

列表至少显示：

- 规格名称；
- 规格编码；
- 前花鼓/后花鼓；
- 轴类型；
- 轴开档；
- 被车架记录引用数量；
- 被前叉记录引用数量；
- 启用状态；
- 更新时间；
- 操作。

首版基础字段固定为：

```text
position
axle_type
axle_spacing_mm
```

字段含义：

- `position`：`front` 前花鼓或 `rear` 后花鼓；
- `axle_type`：轴类型；
- `axle_spacing_mm`：轴开档，单位为毫米。

建议同时保留：

- `spec_code`：稳定的机器编码；
- `display_name`：后台和用户看到的名称；
- `notes`：补充说明；
- `is_enabled`：是否允许被选择和被用户查询；
- `sort_order`：展示顺序。

首版不把轴径、轴长、塔基、碟刹安装方式等字段设为必填。后续如需扩大规格模型，应通过新的设计评审增加字段，不能在备注中长期堆放可检索的结构化数据。

## 5. 数据模型

### 5.1 `fitment_frame_entries`

一条记录代表一个车架/车型在某个年份或年份范围内的适配资料。

```text
id
brand_name
model_name
series_name
generation_name
year_mode
year_from
year_to
market_code
notes
is_enabled
sort_order
created_by
updated_by
created_at
updated_at
deleted_at
```

字段规则：

- `brand_name` 必填；
- `model_name` 必填；
- `series_name` 和 `generation_name` 可选；
- `year_mode` 必填；
- `market_code` 可选；
- `is_enabled` 新建时默认关闭，完成资料和花鼓规格关联后由后台明确启用；
- `deleted_at` 用于软删除。

### 5.2 `fitment_fork_entries`

一条记录代表一个前叉型号在某个年份或年份范围内的适配资料。

```text
id
brand_name
model_name
series_name
generation_name
year_mode
year_from
year_to
market_code
notes
is_enabled
sort_order
created_by
updated_by
created_at
updated_at
deleted_at
```

车架和前叉使用独立实体，即使两者当前字段相似，也不合并成同一张表。

### 5.3 `fitment_hub_specifications`

花鼓规格是被车架和前叉复用的主数据。

```text
id
spec_code
display_name
position
axle_type
axle_spacing_mm
notes
is_enabled
sort_order
created_by
updated_by
created_at
updated_at
deleted_at
```

字段规则：

- `spec_code` 必须唯一；
- `display_name` 必填；
- `position` 必填，只允许 `front` 或 `rear`；
- `axle_type` 必填，使用受控值；
- `axle_spacing_mm` 必填，必须为正整数；
- `is_enabled` 新建时默认关闭，完成校验后由后台明确启用；
- `is_enabled = false` 的规格不能被新记录选择，也不出现在用户查询结果中。

轴类型首版使用受控值。初始值可以包括：

```text
quick_release
thru_axle
bolt_on
other
```

具体中文显示名称由后台界面统一提供，数据库保存稳定值。

### 5.4 `fitment_frame_hub_specifications`

车架与花鼓规格的多对多关联表：

```text
frame_entry_id
hub_specification_id
created_at
```

唯一约束：

```text
unique(frame_entry_id, hub_specification_id)
```

### 5.5 `fitment_fork_hub_specifications`

前叉与花鼓规格的多对多关联表：

```text
fork_entry_id
hub_specification_id
created_at
```

唯一约束：

```text
unique(fork_entry_id, hub_specification_id)
```

两张关联表只引用适配资料库内部的表，不引用商品表。

## 6. 年份规则

年份不能作为任意文本保存。首版使用以下模式：

```text
year_mode = single
year_mode = range
year_mode = all
year_mode = unknown
```

规则：

- `single`：只填写 `year_from`；
- `range`：必须填写 `year_from` 和 `year_to`，且开始年份不大于结束年份；
- `all`：表示明确适用于所有年份；
- `unknown`：表示资料来源没有提供年份；
- `all` 和 `unknown` 都不填写年份，但含义必须严格区分；
- 年份使用四位数字；
- 搜索指定年份时，返回单年匹配、范围覆盖该年份和 `all` 的记录；
- `unknown` 记录默认不作为指定年份的精确匹配结果，但可以在用户选择“未标明年份”时查看。

如果同一车型不同年份对应不同花鼓规格，必须拆成不同的车型记录，不允许用备注覆盖差异。

## 7. 关联与校验规则

### 7.1 花鼓规格选择

- 车架只能选择 `position = rear` 的花鼓规格；
- 前叉只能选择 `position = front` 的花鼓规格；
- 一个车架或前叉可以关联多个花鼓规格；
- 一个花鼓规格可以被多个车架和前叉复用；
- 停用的花鼓规格不能被新记录选择；
- 编辑已有记录时，如果关联规格已停用，应显示停用警告。

### 7.2 记录启用

建议允许后台先保存不完整的停用记录，但启用时必须满足：

- 品牌和型号已填写；
- 年份模式合法；
- 至少关联一个启用的花鼓规格；
- 所有关联规格的前后位置正确。

这样可以支持后台分阶段录入，同时避免不完整资料被用户查询到。

### 7.3 重复记录

同一 TAB 内，以下信息相同的记录应视为重复：

```text
品牌
型号
系列
代际
年份模式和年份范围
市场/地区
```

同一记录可以关联多个花鼓规格，但不能通过重复创建相同车型记录来表达多个规格。

品牌、型号和系列在校验与搜索时应清理首尾空格，并进行大小写不敏感处理。首版不复用商品品牌表，也不把商品品牌作为外键。

### 7.4 停用与删除

- 已被引用的花鼓规格允许停用，不允许硬删除；
- 停用规格仍保留在后台关联记录中，并显示警告；
- 用户查询只返回启用的车型/前叉记录及启用的花鼓规格；
- 已被引用的车架/前叉记录建议软删除；
- 软删除记录不出现在用户搜索结果中；
- 删除或停用操作应记录后台操作人和时间，优先复用现有审计能力。

## 8. 后台接口边界

后台接口按三个资源拆分，不提供跨 TAB 的“大一键保存”接口。

建议资源路径：

```text
/api/admin/fitment-catalog/frame-entries
/api/admin/fitment-catalog/fork-entries
/api/admin/fitment-catalog/hub-specifications
```

基础操作：

```text
GET    /api/admin/fitment-catalog/frame-entries
POST   /api/admin/fitment-catalog/frame-entries
GET    /api/admin/fitment-catalog/frame-entries/:id
PUT    /api/admin/fitment-catalog/frame-entries/:id
DELETE /api/admin/fitment-catalog/frame-entries/:id

GET    /api/admin/fitment-catalog/fork-entries
POST   /api/admin/fitment-catalog/fork-entries
GET    /api/admin/fitment-catalog/fork-entries/:id
PUT    /api/admin/fitment-catalog/fork-entries/:id
DELETE /api/admin/fitment-catalog/fork-entries/:id

GET    /api/admin/fitment-catalog/hub-specifications
POST   /api/admin/fitment-catalog/hub-specifications
GET    /api/admin/fitment-catalog/hub-specifications/:id
PUT    /api/admin/fitment-catalog/hub-specifications/:id
DELETE /api/admin/fitment-catalog/hub-specifications/:id
```

停用/启用可以使用独立的 `PATCH` 操作，也可以作为编辑请求中的字段，但必须保持三个资源的一致接口风格。

车架或前叉保存时，实体和关联花鼓规格应在适配资料库自己的数据库事务中一起提交。关联失败时，不能只保存半条关联数据。

## 9. 用户只读查询

用户侧只读接口只返回启用资料，不需要商品上下文。

建议提供三个资源查询：

```text
GET /api/v1/fitment-catalog/frame-entries
GET /api/v1/fitment-catalog/fork-entries
GET /api/v1/fitment-catalog/hub-specifications
```

支持的查询条件可以包括：

- 关键词；
- 品牌；
- 型号；
- 年份；
- 花鼓规格；
- 前/后位置；
- 分页；
- 排序。

车架查询结果至少包含：

```text
车架品牌
车型/型号
年份
适配花鼓规格列表
```

前叉查询结果至少包含：

```text
前叉品牌
前叉型号
年份
适配花鼓规格列表
```

花鼓规格查询可以反向展示：

```text
使用该规格的车架/车型
使用该规格的前叉
```

查询入口可以出现在花鼓选择、帮助按钮或其他页面，但接口不得因为入口来自某个商品而建立商品关系。

当前公开查询实现使用 `code/data` 响应结构，并且只暴露用户侧需要的只读字段：

- 不返回 `is_enabled`、后台排序字段或后台更新时间；
- 只查询 `is_enabled = true` 的车架、前叉和花鼓规格；
- 车架和前叉结果中的 `hub_specifications` 会过滤掉已停用的花鼓规格；
- 车架和前叉查询支持 `search`、`year`、`page`、`page_size`；
- 花鼓规格查询支持 `search`、`position`、`axle_type`、`page`、`page_size`；
- 请求参数不包含商品、SKU、商品变体、包装或 Shipping 上下文。

## 10. 后台权限

适配资料库使用独立权限，不复用商品、物流或商品规格权限：

```text
fitment_catalog:view
fitment_catalog:create
fitment_catalog:edit
fitment_catalog:delete
```

用户只读查询属于公开读取能力，不授予后台写权限。后台接口必须根据现有鉴权机制校验独立权限。

## 11. 推荐代码边界

Go 域建议使用业务事实命名，避免通用 `model.go`：

```text
go-backend/internal/domain/fitmentcatalog/
  frame_fitment_entry.go
  fork_fitment_entry.go
  hub_specification.go
  frame_hub_specification.go
  fork_hub_specification.go
  fitment_catalog_contract.go
  errors.go
```

后台实现按资源拆分文件，不把三个 TAB 写进一个混合 handler：

```text
go-backend/internal/repository/
  fitment_catalog_repository.go
  fitment_fork_entry_repository.go
  fitment_frame_hub_specification_repository.go
  fitment_fork_hub_specification_repository.go
  fitment_hub_specification_repository.go

go-backend/internal/service/
  fitment_catalog_service.go
  fitment_fork_entry_service.go
  fitment_hub_specification_service.go

go-backend/internal/api/admin/
  fitment_catalog_http.go
  fitment_frame_entry_handler.go
  fitment_fork_entry_handler.go
  fitment_hub_specification_handler.go
```

Admin 前端也按资源拆分：

```text
go-backend/web/admin/src/api/fitmentCatalog/
  types.ts
  readers.ts
  frameEntries.ts
  forkEntries.ts
  hubSpecifications.ts

go-backend/web/admin/src/views/
  FrameFitmentEntries.vue
  ForkFitmentEntries.vue
  HubSpecificationEntries.vue

go-backend/web/admin/src/lib/fitmentCatalogTabs.ts
```

如果公开查询需要独立鉴权或响应结构，可以增加单独的 read handler，但仍然归属于 `fitmentcatalog` 域。

## 11.1 当前实施状态

截至 2026-08-26，后台首阶段已实现：

- `fitment_frame_entries` 车架/车型资料；
- `fitment_fork_entries` 前叉资料；
- `fitment_hub_specifications` 花鼓规格资料；
- `fitment_frame_hub_specifications` 车架到花鼓规格关联；
- `fitment_fork_hub_specifications` 前叉到花鼓规格关联；
- Admin 三个独立 TAB：车架/车型、前叉、花鼓规格；
- Admin 中英文文案；
- 后台独立权限 `fitment_catalog:view/create/edit/delete`；
- 车架只能选择后花鼓规格；
- 前叉只能选择前花鼓规格；
- 花鼓规格被车架或前叉引用时不能删除，也不能修改前/后位置；
- 面向用户的公开只读查询接口：
  - `GET /api/v1/fitment-catalog/frame-entries`
  - `GET /api/v1/fitment-catalog/frame-entries/:id`
  - `GET /api/v1/fitment-catalog/fork-entries`
  - `GET /api/v1/fitment-catalog/fork-entries/:id`
  - `GET /api/v1/fitment-catalog/hub-specifications`
  - `GET /api/v1/fitment-catalog/hub-specifications/:id`

尚未进入本阶段的内容：

- Nuxt storefront 20 语言文案；
- 商品页问号查询入口；
- 车架与前叉的直接组合规则；
- 商品、SKU、包装或 Shipping 关联。

## 12. 故障隔离

适配资料库的保存只影响适配资料库自己的事务：

- 车架保存失败，不影响前叉或花鼓规格；
- 前叉保存失败，不影响商品；
- 花鼓规格保存失败，不影响车架或前叉已有记录；
- 商品保存失败或商品删除，不影响适配资料；
- 适配资料查询失败，不应阻止商品浏览、商品保存或购物流程。

适配资料库的缓存、搜索索引和失效逻辑也只能由本域管理。

## 13. 首版验收标准

首版完成后应满足：

1. 后台存在独立的车型前叉花鼓适配库入口；
2. 页面有车架/车型、前叉、花鼓规格三个独立 TAB；
3. 可以独立新增、编辑、停用和查询三类资料；
4. 车架和前叉可以多选花鼓规格；
5. 车架只能选择后花鼓规格，前叉只能选择前花鼓规格；
6. 花鼓规格包含前/后、轴类型、轴开档三个基础字段；
7. 用户可以只读搜索启用的资料；
8. 搜索结果能够展示关联花鼓规格；
9. 适配资料库的任何表都不含商品 ID、SKU 或商品外键；
10. 商品保存、商品编辑、商品删除和 Shipping/包装流程不因本域改动；
11. 适配资料域可以独立配置权限和审计；
12. 不因商品选择页面打开查询入口而产生数据写入。

## 14. 后续扩展

以下能力可以在首版稳定后单独评审：

- 轴径、轴长、塔基和刹车安装方式；
- 车架与前叉的直接组合规则；
- 品牌、车型和前叉的独立主数据；
- 多语言显示名称；
- 批量导入和导出；
- 资料来源、审核和版本历史；
- 更复杂的技术兼容性计算。

这些扩展不能通过给商品表添加字段或把适配数据塞进商品规格模板来实现。

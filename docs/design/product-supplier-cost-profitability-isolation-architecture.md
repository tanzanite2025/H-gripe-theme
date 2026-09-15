# 商品成本与利润附加域隔离设计

最后更新：2026-09-08

## 0. 范围声明

当前系统没有采购业务模块。本设计只维护 SKU 级单位成本价、供应商参考资料和利润计算输入，不实现供应商下单、收货、库存入库、供应商付款、质检、跟单或任何供应商侧执行流程。

文中保留的 `product_procurement_records`、`/api/admin/procurement/...`、`procurement:*` 和 `purchase_price` 都是历史兼容标识：它们保留是为了不破坏既有数据库、路由、权限和旧客户端，不代表系统存在采购业务。

### 初始化与测试边界

应用初始化、SQL 迁移和隔离测试仍必须创建或保留
`product_procurement_records`，因为当前成本仓储通过历史物理表名读写
SKU 成本资料，利润快照也依赖这条成本资料链路。迁移 `239` 用幂等的
`CREATE TABLE IF NOT EXISTS` 修复“迁移账本已前进但历史表缺失”的数据库异常；
它修复的是成本资料表的 schema，不是开启采购模块。

同理，测试中的 `AutoMigrate(ProductSupplierCostRecord{})` 产生的也是这张
历史兼容表。测试可以用 `suppliercostdomain`、`ProductSupplierCostRecord`
和 `unit_cost` 作为业务命名，但不能因为业务不包含采购流程就跳过这张表的
初始化，否则会把真实的成本资料写入路径误判为缺表。

## 1. 目的

在后台商品编辑器中录入每个 SKU 的单位成本价和供应商参考资料，并在录入过程中实时看到预计毛利。

这个功能是商品目录的独立附加域：

```text
商品目录域
  products
  product_variants
  商品创建/更新接口
  商品缓存、商品事件和前台响应

商品成本与供应商资料附加域
  product_procurement_records  # 历史物理表名，只存 SKU 成本资料
  product_profit_calculations
  单位成本价、供应商资料、到货周期、起订量、附加成本和利润快照
```

成本写模型和利润计算只保存 SKU 文本、名称快照和成本资料，不通过数据库外键绑定商品。
后台选择真实商品时，商品成本域可以使用一个独立的只读目录投影查询；这个查询只读取选择器所需的最小字段，不能进入商品写入事务，也不能反向修改商品域。

## 2. 最终范围

### 2.1 保留字段

每个 SKU 的成本资料保留：

- 单位成本价；
- 成本币种；
- 供应商名称/成本来源；
- 供应商联系人；
- 供应商电话；
- 供应商邮箱；
- 到货周期；
- 起订量；
- 入库运费/件；
- 包装成本/件；
- 其他附加成本/件。

目标客户地区关税不属于本域。前端结算会根据目标客户地区动态计算并收取，不进入产品成本或 SKU 毛利快照。

起订量是供应商的交易条件，不是实际下单数量。它只表示“供应商至少接受多少件”，不代表系统要记录进货或入库了多少件。

### 2.2 明确不做

本阶段不做：

- 实际进货或供应商下单数量；
- 供应商订单或 PO；
- 库存补货、收货或入库流程；
- 供应商货号；
- 供应商地址；
- 供应商下单备注；
- 成本记录启用/停用状态；
- 付款、质检、跟单流程；
- 供应商历史报价；
- 物流订单或库存变动；
- 订单级实际成本；
- 税务、平台佣金、支付手续费和净利润。

“预计毛利”不是财务净利润。

## 3. 不可违反的商品域边界

以下对象不能因为本功能而修改：

- `products` 表；
- `product_variants` 表；
- 商品领域结构体；
- 商品创建请求和更新请求；
- `ProductService`；
- `ProductRepository`；
- 商品事务管理器；
- 商品缓存失效逻辑；
- 商品事件或 Merchant Center 事件；
- 前台商品、购物车、订单、SEO 和 Merchant Center 响应。

成本写模型和利润计算不能直接依赖商品写模型；成本新增时如需校验 SKU，只能通过独立的只读目录投影适配器完成。具体不能：

- 增加 `product_id`；
- 增加指向 `products` 或 `product_variants` 的外键；
- 在成本写入或利润计算中调用完整商品仓储、加入商品写事务或执行商品写操作；
- 在后端商品保存事务内保存成本；
- 因商品删除自动删除成本记录；
- 因 SKU 改名自动删除旧成本记录；
- 把客户端传来的利润金额当成事实；
- 把缺失单位成本价静默当成零成本；
- 自动进行币种换算。

## 4. 保存顺序与故障隔离

商品编辑器保存时按以下顺序执行：

```text
1. 浏览器根据当前商品表单计算本地预览；
2. 使用现有商品请求保存商品；
3. 商品 API 成功返回；
4. 浏览器使用返回结果中的 SKU 和名称快照组装成本域请求；
5. 调用独立的成本/利润批量接口；
6. 分别显示商品保存结果和成本保存结果。
```

两个操作不共享数据库事务。

### 商品保存失败

- 不调用成本接口；
- 不写成本表；
- 保持现有商品错误处理。

### 商品保存成功、成本保存失败

- 商品保持已保存状态；
- 成本域标记为待重试；
- 页面显示独立重试按钮；
- 重试只调用成本接口，不重复创建商品。

### 成本批量请求失败

成本域自己的批量事务整体回滚，不允许只保存半批 SKU。

## 5. 数据模型

### 5.1 历史物理表 `product_procurement_records`

这是历史物理表名。当前业务语义是 SKU 级商品成本记录，不是供应商订单、收货记录、付款记录、库存记录或供应商交易。

保留字段：

```text
id
product_code
product_name
purchase_price  # 历史列名；业务语义是 unit_cost
currency
supplier_name
supplier_contact_name
supplier_phone
supplier_email
lead_time_days
minimum_order_quantity
created_at
updated_at
```

唯一键是 `product_code`。它是文本集成键，不是商品表关系。

`minimum_order_quantity` 只表示起订量：

- 允许为空输入时使用默认值 1；
- 必须大于等于 1；
- 不表示库存；
- 不表示本次下单或入库数量；
- 不触发供应商订单或库存逻辑。

### 5.2 `product_profit_calculations`

这是每个 SKU 的当前利润快照。

保留字段：

```text
id
product_code
product_name
currency
list_price
sale_price
effective_selling_price
purchase_price  # 历史列名；业务语义是 unit_cost
inbound_shipping_unit_cost
packaging_unit_cost
other_unit_cost
landed_cost
gross_profit
gross_margin_bps
calculation_status
formula_version
warnings
calculated_at
created_at
updated_at
```

`calculation_status` 是利润计算结果状态，不是成本记录启用/停用状态。供应商下单状态字段已从成本资料中移除。

两个表都不包含：

- `product_id`；
- `product_variant_id`；
- `REFERENCES products`；
- `REFERENCES product_variants`。

### 5.3 迁移

历史迁移 `210` 和 `211` 不修改。

迁移 `214_reduce_product_procurement_metadata` 只从成本来源表删除不再需要的元数据：

- `supplier_address`；
- `supplier_product_code`；
- `notes`；
- `is_enabled`。

迁移 `214` 不删除：

- `purchase_price`；
- 供应商名称和联系方式；
- `lead_time_days`；
- `minimum_order_quantity`；
- 入库运费、包装和其他附加成本字段；
- 利润快照表。

迁移 `221_remove_destination_tax_from_product_costs` 会从历史 `product_procurement_records` 表和利润快照表删除历史 `customs_unit_cost` 列，并按新的不含目标地区关税公式重算已有的可用利润快照。

迁移上下行都只操作成本附加域，不触碰商品表。

## 6. 利润计算契约

### 6.1 售价

```text
effective_selling_price =
  sale_price 有值时使用 sale_price
  否则使用 list_price
```

促销价缺失时使用常规售价，并返回 `sale_price_missing` 警告。

### 6.2 成本

```text
landed_cost =
  unit_cost
  + inbound_shipping_unit_cost
  + packaging_unit_cost
  + other_unit_cost
```

### 6.3 毛利

```text
gross_profit =
  effective_selling_price - landed_cost

gross_margin =
  gross_profit / effective_selling_price
```

示例：

```text
实际售价       90.00 USD
单位成本价     55.00 USD
入库运费        2.00 USD
包装            0.50 USD
其他成本        0.00 USD

含附加成本     57.50 USD
预计单位毛利   32.50 USD
预计毛利率     36.11%
```

公式版本使用：

```text
gross-margin-v3-no-customs
```

### 6.4 缺失单位成本价

必须区分：

```text
unit_cost 未填写
  -> 成本未知
  -> 不生成可用利润快照
  -> 清除该 SKU 的旧利润快照和旧成本来源记录

unit_cost = 0
  -> 明确输入的零成本
  -> 允许计算，但应由管理员确认业务含义
```

旧请求字段 `purchase_price` 仍被后端接受，旧响应字段也仍会返回；新客户端和新文档应使用 `unit_cost` 表达业务含义。

### 6.5 币种

成本域保存成本币种，但利润计算要求成本币种与商品主币种一致：

- 一致时计算；
- 不一致时返回 `currency_mismatch`；
- 不调用汇率接口；
- 不静默换算；
- 不把换算结果伪装成原始成本。

前端商品编辑器默认使用商品主币种；后端仍然执行严格校验。

### 6.6 金额精度

后端是最终计算权威：

1. 标准化币种代码；
2. 转换为币种最小单位；
3. 使用整数最小单位计算；
4. 按币种精度输出成本、毛利；
5. 用基点保存毛利率；
6. 返回规范化结果给前端。

前端预览只用于即时反馈，不能作为持久化事实。

## 7. 后端结构

成本域的依赖方向：

```text
admin handler
  -> supplier-cost/profit service
  -> supplier-cost/profit repository
  -> independent tables
```

主要文件：

```text
go-backend/internal/domain/productsuppliercost/
go-backend/internal/repository/product_supplier_cost_record_repository.go
go-backend/internal/repository/product_profit_calculation_repository.go
go-backend/internal/service/product_supplier_cost_record_service.go
go-backend/internal/service/product_profitability_service.go
go-backend/internal/api/admin/product_supplier_cost_record_handler.go
go-backend/internal/api/admin/product_profitability_handler.go
```

纯计算器不能导入商品领域，也不能访问数据库。

### 7.1 真实商品绑定与只读目录投影

独立商品成本页不能继续让用户手工输入产品编码和产品名称作为默认入口，否则会产生无法确认来源的孤立 SKU 和名称错误。选择器的绑定单位是一个真实的商品变体 SKU：

```text
product_variants.sku
```

产品编辑器和独立商品成本页都以一个 `product_variants` 变体行的 SKU 为成本资料的业务键。商品成本域不把 `product_id` 或 `variant_id` 写入历史 `product_procurement_records` 表，也不把它们当成跨域外键。

选择器当前使用历史兼容路由下的专用只读接口：

```text
GET /api/admin/procurement/product-options
```

接口必须分页，并支持按 SKU、商品名称和变体标题搜索。默认只返回未删除、商品为 active、变体为 active 的选项。编辑历史成本记录时，如果原 SKU 已停用、删除或 SKU 已被改名，接口可以通过精确 SKU 返回历史状态；如果已经不存在，则页面保留成本记录自身的名称和 SKU 快照，并明确显示“原商品不可用”，不能静默替换成另一个 SKU。

选择器允许返回的最小字段：

```text
product_name     商品名称展示值和成本名称快照来源
variant_title    变体展示值
sku              稳定的跨域业务键
available        当前选项是否仍可用；仅用于编辑历史记录时提示
```

选择器不暴露 `product_id` 或 `variant_id`。SKU 在商品域中已经是全局唯一的业务键，内部数据库 ID 对当前绑定没有额外价值，也不会被带入前端或成本请求。

选择器禁止返回或透传以下无关数据：售价、币种、库存、媒体、描述、分类、品牌、规格详情、物流模板、售后模板、清关字段、浏览量和前台扩展状态。独立成本页的成本币种由成本资料自身维护；利润预览需要售价时，走已有利润接口或商品编辑器自己的 SKU 数据，不扩充绑定选择器。它不复用完整的 `/api/admin/products` 响应，也不调用完整商品详情服务。

这里的“只读目录投影”是成本域的查询适配器，不是商品域的写入依赖。它可以读取商品表的明确列，但不能加入商品事务、商品仓储写方法、商品缓存失效、商品事件或商品删除钩子。若未来拆分服务，再将这份投影替换为商品域提供的同等最小字段只读接口，成本表契约无需改变。

## 8. API 契约

下面的 `/api/admin/procurement/...` 是后端保留的历史兼容路径。它们当前承载的是
商品供应商成本资料和利润计算接口，不提供供应商下单、收货、入库、付款或采购单能力；
新的代码、页面和函数必须使用 `product_supplier_cost` / `supplier_cost` 语义理解这些接口。

### 成本来源接口

```text
GET    /api/admin/procurement/records
GET    /api/admin/procurement/records/by-codes
GET    /api/admin/procurement/records/:id
POST   /api/admin/procurement/records
PUT    /api/admin/procurement/records/:id
DELETE /api/admin/procurement/records/:id
GET    /api/admin/procurement/product-options
```

新增成本记录请求只提交 SKU 和成本资料：

```json
{
  "sku": "SKU-001",
  "unit_cost": 55,
  "currency": "USD",
  "supplier_name": "Example supplier",
  "supplier_contact_name": "Lina",
  "supplier_phone": "+86-123",
  "supplier_email": "lina@example.com",
  "lead_time_days": 14,
  "minimum_order_quantity": 20
}
```

服务端通过商品成本域专用的只读目录投影按精确 SKU 查询商品。只有商品未删除、商品为 active 且变体为 active 时，才能新建成本记录；`product_name` 由服务端生成并保存为名称快照，客户端不能伪造。

编辑成本记录时不再提交 `sku`、`product_code` 或 `product_name`。服务端根据记录 ID 读取原有 SKU 和名称快照，只更新单位成本价、供应商资料、到货周期、起订量和附加成本。即使原商品后来不可用，历史成本记录仍可维护，不能静默改绑其他 SKU。

独立成本接口不接受旧的手工 `product_code/product_name` 写入契约，也不接受供应商下单数量、供应商下单状态、地址、供应商货号或备注；这些词只出现在“不做”范围内。请求体使用严格字段校验，未知字段直接返回 400。

### 利润接口

```text
GET  /api/admin/procurement/profitability/by-codes
POST /api/admin/procurement/profitability/preview
POST /api/admin/procurement/profitability/bulk-upsert
```

利润批量请求包含售价、单位成本价和附加成本，并可携带供应商资料：

这里的 `product_code/product_name` 属于商品编辑器保存成功后提交的利润快照入口，不是独立成本页的商品选择入口。它保留用于商品编辑器把商品保存结果投影到独立附加域；不能作为成本页手工创建任意 SKU 的兼容接口。

```json
{
  "items": [
    {
      "product_code": "SKU-001",
      "product_name": "Example item",
      "currency": "USD",
      "list_price": 100,
      "sale_price": 90,
      "unit_cost": 55,
      "unit_cost_known": true,
      "inbound_shipping_unit_cost": 2,
      "packaging_unit_cost": 0.5,
      "other_unit_cost": 0,
      "supplier_cost_details": {
        "supplier_name": "Example supplier",
        "supplier_contact_name": "Lina",
        "supplier_phone": "+86-123",
        "supplier_email": "lina@example.com",
        "lead_time_days": 14,
        "minimum_order_quantity": 20
      }
    }
  ]
}
```

后端会重新计算并覆盖客户端传入的利润值。

## 9. 前端行为

### 9.1 商品编辑器

商品编辑器新增独立区域：

```text
SKU
单位成本价、成本币种
供应商、联系人、电话、邮箱
到货周期、起订量
入库运费、包装成本、其他成本
实际售价、含附加成本、预计单位毛利、预计毛利率
```

成本草稿按 SKU 存储在独立状态中，不加入商品表单类型，不进入 `buildProductPayload()`。

### 9.2 侧边栏

侧边栏名称为“商品成本”，页面只维护独立成本资料和上述供应商资料，不显示供应商下单状态。新增记录时通过真实 SKU 选择器绑定商品；产品编码和产品名称不再作为普通自由输入框。选择结果只读回填产品名称、变体标题和 SKU；单位成本价、成本币种、供应商、周期、起订量和附加成本仍由本页维护。

后台页面使用 `/product-supplier-cost/records`。历史 `/procurement/records` 页面地址只保留为重定向，避免旧书签失效；它不是新的业务命名。

### 9.3 权限

继续使用独立的成本域权限：

```text
procurement:view
procurement:create
procurement:edit
procurement:delete
```

无查看权限的用户不能读取单位成本价、供应商资料和利润。

## 10. SKU 变更和商品删除

### SKU 改名

- 旧 SKU 记录不自动删除；
- 新 SKU 只有在明确保存成本资料后才创建记录；
- 成本域不推断旧 SKU 是否已经废弃；
- 后续如需退休 SKU，应增加独立的显式操作。

### 商品删除

- 商品域按原有逻辑删除；
- 成本和利润记录不级联删除；
- 不增加商品删除钩子；
- 清理成本记录必须是成本域独立操作。

## 11. 测试与验收

### 后端

- 单位成本价缺失与显式零值区分；
- 销售价优先使用促销价；
- 运费、包装和其他成本进入含附加成本；
- 起订量保存、默认值和边界校验；
- 到货周期边界校验；
- 币种不一致不计算；
- 负毛利返回警告；
- 批量请求失败不部分写入；
- 同一 SKU 重复提交保持单条记录；
- 清空单位成本价会清理独立旧快照；
- 预览接口不写数据库；
- 无商品表也能运行成本仓储测试；
- 迁移不引用商品表或外键。

### 前端

- 商品请求不包含成本或供应商字段；
- 成本请求不包含 `product_id`；
- 改售价、单位成本价和非地区税附加成本时预览即时变化；
- 单位成本价为空时不显示零成本利润；
- 起订量和到货周期可编辑；
- 商品成功但成本失败时可独立重试；
- 重试不重复提交商品；
- 页面不显示供应商下单状态字段；
- 独立成本页通过分页搜索选择真实商品变体 SKU；
- 选择器不请求完整商品对象，也不把库存、媒体、描述、分类或规格带入成本请求；
- 选择结果只读回填名称、变体标题和 SKU；
- 编辑已删除或停用的 SKU 时保留历史快照，不能静默改绑到其他 SKU；
- 选择器权限只要求 `procurement:view`，不强制依赖 `product:view`；
- 选择器响应字段保持最小集合，不包含无关商品数据；
- `npm run typecheck` 通过；
- `npm run build` 通过。

### 静态安全检查

合并前必须检查：

```text
1. products/product_variants 没有 schema diff；
2. 商品后端服务、仓储和事务管理器没有成本写入依赖；
3. 成本域没有 product_id；
4. 成本新增只通过只读目录投影校验 SKU 和生成名称快照，不调用完整商品仓储、不加入商品写事务；利润计算本身不查询商品表；
5. 迁移 214 只操作成本表；
6. 商品请求快照前后不包含成本或供应商字段；
7. Go 测试、vet、前端类型检查和构建全部通过。
```

## 12. 回滚与发布顺序

发布顺序：

```text
1. 应用独立成本域迁移；
2. 发布后端成本/利润接口；
3. 验证成本域权限和契约测试；
4. 发布后台侧边栏和商品编辑器；
5. 观察保存失败和独立重试；
6. 再扩大权限范围。
```

出现问题时：

- 隐藏商品成本区域；
- 停止前端调用利润批量接口；
- 保持商品创建和编辑继续可用；
- 不回滚或修改商品表；
- 不修改商品服务；
- 成本域数据是否保留由独立数据保留策略决定。

## 13. 完成定义

只有同时满足以下条件才算完成：

- 单位成本价、供应商联系方式、到货周期和起订量可保存；
- 运费、包装和其他附加成本参与利润计算；
- 供应商下单状态和实际进货数量没有进入新契约；
- 毛利按 SKU 计算；
- 缺失单位成本价不会被当作零；
- 成本域没有商品外键；
- 商品表和商品后端数据链路没有修改；
- 商品保存与成本保存故障隔离；
- 成本失败可以独立重试；
- 设计文档、迁移、测试和前端实现保持一致。

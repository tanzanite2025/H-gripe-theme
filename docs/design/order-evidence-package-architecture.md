# 订单履约证据包架构

Last updated: 2026-09-09

## 设计结论

订单履约证据不是 Wheels 专属能力，而是所有已成交订单的通用订单域能力。标准商品、定制商品、组件、套装、轮组和高价值单件商品都应建立订单级证据链。

高价判断只有一个触发源：订单创建时锁定的最终订单总额。订单总额达到或超过 750 USD 等值时，进入高价完整证据策略；不再使用商品名称、商品分类、单个商品价格、支付争议金额或 QUICKBUY 作为第二套高价判断。

轮组只比其他产品多一项条件证据：编轮质检张力表（Spoke QC tension record）。是否需要张力表由后台维护的产品/Variant 履约要求明确配置，普通辐条、轮圈、花鼓或其他单独零件不会因为名称相近而自动触发。

前台入口不参与证据判定，也不需要为证据包增加专用字段、服务或分支；系统只读取订单总额和订单行快照。

订单管理下已经接入独立“履约证据”入口，产品中心下也已经接入独立“产品履约要求”入口。两个入口分别维护订单证据事实和产品/Variant 规则，不把逻辑重新塞回现有大型订单页面、订单 handler 或支付争议服务。

## 财务订单留存与后台隐藏边界

订单、订单行、支付交易、退款、拒付和订单证据包共同构成财务对账与拒付抗辩链。只要订单发生过已支付或退款等资金事实，就必须持续保留订单主记录及其可关联的快照、流水和证据；不能为了从后台默认列表移除而物理删除。

- `POST /api/admin/orders/:id/hide-unpaid-terminal` 是当前唯一推荐的后台隐藏入口，只允许 `status=cancelled,payment_status=unpaid` 或 `status=payment_expired,payment_status=expired`。`payment_status=paid`、`payment_status=refunded`、`status=refunded` 以及其他不匹配状态都必须拒绝。
- 当前实现使用 GORM 软删除让符合条件的未支付终态订单退出默认查询；软删除不是财务归档，也不是物理清除。当前证据组装、对账和拒付工作流不能把隐藏订单当成可恢复的 Archived 订单。
- 旧 `DELETE /api/admin/orders/:id` 仅为兼容旧客户端，实际调用同一保护逻辑且不会物理删除；新代码、前端和文档不得继续把它称为删除订单。
- 当前订单域没有完整的 `Archived` 状态、归档查询、恢复契约或归档审计流程。因此，不能把本次“从默认列表隐藏”描述成“已归档”；若未来需要可运营归档，应另行设计状态、查询和审计契约。

## 证据适用范围

### 所有订单的通用基础证据

每一笔已成交订单都应围绕订单和订单行形成以下基础证据：

1. 订单产品或配置确认单。
2. 产品或组件身份记录。
3. 出库称重与包装照片。
4. 物流签收或交付证明（POD）。

“订单产品或配置确认单”不是事后从当前商品页面重新拼出来的说明。它必须使用下单时锁定的产品、variant、数量、价格、重量、规格、客户确认内容和版本信息生成订单快照。订单创建后商品目录发生变化，不能改变历史确认单的含义。

身份记录也不等于所有商品都必须有物理 serial number：

- 有物理序列号的产品，记录 serial number。
- 没有序列号的产品，记录 SKU、variant、批次、生产批号或订单行唯一标识。
- 系统不得为了满足清单而虚构 serial number。

### 轮组条件追加证据

当订单行命中后台已启用的产品/Variant 履约要求 `spoke_tension_qc_required = true` 时，追加编轮质检张力表。

产品分类、商品名称、前台入口和价格都不是轮组识别条件。后台配置必须绑定到稳定的 `product_id`，必要时下沉到 `variant_id`；分类只能用于页面筛选，不能作为最终判定。

张力表应绑定到订单、订单行或装配组，并在发货时由人工上传照片或扫描件。若同时录入测量值、允许范围、测量工具、测量时间、操作人或人工结论，这些内容都只作为人工提交的原始记录保存。系统不识别照片内容、不比较测量值、不推导“合格/不合格”，也不把张力表变成所有产品的统一必填字段。

### 高价值订单

订单创建时的最终订单总额达到或超过 750 USD 等值时，无论产品类别，都进入高价值证据策略：

- 生成完整通用基础证据包。
- 将高价标记和 USD 等值金额固化在订单快照和证据包元数据中，供后台筛选和后续拒付适配使用。
- 使用下单时锁定的订单汇率快照进行等值判断。
- 如果某个订单行同时命中产品/Variant 的张力表要求，再追加该订单行的编轮质检张力表。

高价本身不会新增一个重复的 `high_value` 证据项，也不会改变基础证据项的类型。高价订单仍然必须完成同一套通用基础证据；只有订单行命中已快照化的产品/Variant 履约要求时，才追加张力表证据。

因此，750 USD 以上的辐条订单需要高价完整证据包，但不需要编轮张力表；是否需要张力表只看该订单行在下单时命中的产品履约要求。

### 混合订单

一笔订单包含多个产品或多个装配角色时：

- 订单级公共证据只保存一份。
- 产品身份和配置证据按订单行或装配组保存。
- 高价条件是订单级，只计算一次。
- 张力表条件是订单行级，只追加命中的订单行。
- 不能因为其中一行是普通商品，就覆盖另一行的张力表要求。

## 配置确认单的定义

“定制配置确认单”应定义为订单创建时生成的不可变订单产品快照，而不是事后从当前商品页面重新拼出来的说明。它适用于所有订单，不区分订单来自哪个前台入口。

最小契约：

```text
schema_version
confirmed_at
currency
order_total
order_total_usd
items[]
  order_item_id
  product_id
  variant_id
  quantity
  sku
  product_name
  selected_specs_json
  unit_price
  weight_g
  product_requirement_snapshot_json
snapshot_sha256
```

其中 `product_requirement_snapshot_json` 至少要固化 `spoke_tension_qc_required`、命中的规则 ID、规则版本和判定原因。订单创建后，商品规则可以变化，但历史订单不能重新读取当前规则。

## 当前代码审计

当前订单代码可以继续承载订单核心生命周期，但不适合直接吸收订单证据包业务：

| 现有文件 | 当前职责 | 处理原则 |
| --- | --- | --- |
| `go-backend/internal/domain/order/order.go` | 订单状态、履约模式、金额、地址、生产和签收要求 | 只增加稳定的订单快照引用或最小关联字段，不放证据项和附件逻辑 |
| `go-backend/internal/domain/order/order_item.go` | 订单行的商品、价格、税费、海关和属性快照 | 只保留订单行事实；配置确认和身份记录拆到订单证据域 |
| `go-backend/internal/service/order_create_service.go` | 报价、幂等、创建订单、支付结算、政策披露、归因和库存提交 | 文件已经承担多个创建后动作；先抽出订单快照构建边界，再接入证据快照，不直接加入上传或 TAB 逻辑 |
| `go-backend/internal/service/order_admin_service.go` | 后台订单查询、状态、履约、备注、导出和订单管理动作 | 不增加证据包聚合、附件存储或拒付提交逻辑 |
| `go-backend/internal/api/admin/order_handler.go` | 订单管理 HTTP handler、状态、物流、生产、备注和海关接口 | 生产完成请求只处理确认和状态变更；不承载证据包查询、附件存储或证据规则 |
| `go-backend/internal/service/order_production_service.go` | 生产开始/完成状态变更 | 只处理生产状态、支付和生产模式；不读取证据包、不上传或写入张力表、不判定证据质量 |
| `go-backend/web/admin/src/views/Orders.vue` | 订单列表、详情、拒付、物流、售后、导出和批量动作 | 维持订单编排职责；证据 TAB 使用独立视图/组件和 API 模块接入，不把证据规则写进页面脚本 |
| `go-backend/internal/domain/product/product.go` 与商品编辑服务 | 商品目录、规格、Variant、价格、库存和分类 | 不通过分类或名称推断轮组；产品履约要求使用独立规则存储和独立后台入口 |
| `go-backend/web/admin/src/views/Products.vue` 与 `ProductEditorDialog.vue` | 商品目录维护和商品内容编辑 | 不承载张力表规则的批量维护；最多展示当前规则状态的只读入口 |
| `go-backend/internal/service/payment_dispute_evidence_service.go` | 按 Stripe dispute 生成和提交拒付证据 | 后续只读取订单证据域，不作为订单证据录入或唯一存储 |
| `go-backend/internal/service/payment_paypal_dispute_evidence_service.go` | PayPal dispute 证据生成和提交 | 同上，保留支付渠道格式转换职责 |
| `go-backend/internal/service/media_evidence.go` | 媒体版权证据导出和完整性校验 | 不直接复用为订单附件域；可复用哈希和存储适配思路 |

结论：受影响文件并非不能维护，但已经存在明显的职责密度。当前阶段不应在这些文件中新增证据包聚合、上传、订单检索和拒付导出的一体化逻辑。

## 新的独立职责边界

独立订单证据域已经按业务事实拆分。产品履约要求是证据域读取的上游配置，不塞进订单或支付文件：

```text
go-backend/internal/domain/productrequirement/
  product_quality_requirement.go

go-backend/internal/domain/orderevidence/
  order_evidence_package.go
  order_evidence_item.go
  order_evidence_attachment.go
  order_evidence_lifecycle.go
  order_evidence_outbound_weight_packaging.go
  order_evidence_signed_pod.go
  order_evidence_snapshot.go
  order_evidence_submission_snapshot.go
  order_evidence_export_snapshot.go

go-backend/internal/repository/
  product_quality_requirement_repository.go
  order_evidence_repository.go
  order_evidence_snapshot_repository.go
  order_evidence_submission_snapshot_repository.go
  order_evidence_export_snapshot_repository.go

go-backend/internal/service/
  product_quality_requirement_service.go
  order_evidence_service.go
  order_evidence_admin_service.go
  order_evidence_snapshot_service.go
  order_evidence_attachment_service.go
  order_evidence_package_assembler.go
  order_evidence_export_snapshot_service.go

go-backend/internal/api/admin/
  product_quality_requirement_handler.go
  product_quality_requirement_contract.go
  product_quality_requirement_routes.go
  order_evidence_handler.go
  order_evidence_contract.go
  order_evidence_routes.go

go-backend/migrations/
  243_product_quality_requirement_rules.up.sql
  244_order_evidence_snapshots.up.sql
  245_order_evidence_package_items.up.sql
  247_order_evidence_submission_snapshots.up.sql
  248_order_evidence_export_snapshots.up.sql
```

职责分工：

- `order_evidence_snapshot_service`：在订单创建边界接收已经确认的产品、配置、价格、重量和追溯快照，生成不可变订单证据快照。
- `order_evidence_service`：按订单总额计算高价证据策略，按订单行的产品履约要求追加张力表，维护证据状态和版本；不判定张力表或实物是否合格。
- `order_evidence_attachment_service`：处理附件的 MIME、大小、对象存储、SHA-256、访问权限、所有权和删除/替换审计。
- `order_evidence_package_assembler`：只读取订单证据和当前物流，生成后台详情与支付适配器使用的实时组装视图；不上传、不修改源记录、不转换支付渠道字段。
- `order_evidence_export_snapshot_service`：只在证据包锁定后创建或复用不可变 JSON 导出快照；冻结当时的订单、证据和匹配物流，不包含对象存储 Key。
- `order_evidence_handler`：提供订单证据查询、保存结构化记录、附件引用和锁定清单导出接口，不直接操作文件系统。
- `product_quality_requirement_service`：维护产品/Variant 的张力表要求，并在订单创建时解析出规则快照；不负责订单证据附件和拒付提交。
- 支付争议服务：读取订单证据域，按 Stripe 或 PayPal 的字段限制转换为提交材料。

产品履约要求独立维护：

```text
product_quality_requirement_rules
  product_id
  variant_id                 -- nullable; null means all variants
  requirement_type           -- spoke_tension_qc
  spoke_tension_qc_required
  status                     -- active / inactive
  rule_version
  reason
  created_by
  created_at
  updated_at
```

它由独立的“产品履约要求”后台页面维护。页面先从商品目录分页读取真实商品，打开单个商品后读取当前 Variant 和该商品的显式规则；商品级规则与 Variant 级覆盖规则分别保存。当前版本不提供批量启停，也不通过分类、名称或订单数量反推规则。停用只影响新订单判定，不改写历史订单快照；已经被历史订单引用的规则不能物理删除。

规则解析优先级固定为：精确 `variant_id` 规则 > `product_id` 默认规则 > 未命中。订单创建时只解析一次，并把命中的规则 ID、版本和结果写入订单证据快照；后续商品编辑或规则停用不回溯历史订单。首版只实现 `spoke_tension_qc` 一个规则类型，不提前做通用规则引擎。

## 建议的数据对象

### `order_evidence_packages`

```text
id
order_id
package_version
status                      -- incomplete / ready / locked / superseded
order_total_usd_snapshot
is_high_value
has_spoke_tension_qc
schema_version
created_by
locked_at
created_at
updated_at
```

### `order_evidence_items`

```text
id
package_id
order_id
order_item_id               -- nullable for order-level evidence
item_type                   -- configuration_confirmation
                            -- product_identity
                            -- outbound_weight_packaging
                            -- signed_pod
                            -- spoke_qc_tension
status                      -- missing / draft / complete / waived
required_reason             -- base / high_value / spoke_tension_qc
data_json
captured_at
captured_by
content_sha256
created_at
updated_at
```

其中 `spoke_qc_tension` 只能由订单创建时保存的产品履约要求产生，不能由前端把它作为所有订单的固定必填项提交。它的完成只表示人工附件和可选备注已经保存，不表示系统确认张力合格。

### `order_evidence_attachments`

```text
id
evidence_item_id
storage_key
original_filename
mime_type
size_bytes
sha256
uploaded_by
created_at
```

附件只保存对象存储引用和完整性信息。订单证据域不复制文件系统实现，也不把外部 URL 当成没有哈希的永久事实。

### 证据包生命周期

证据包创建后先处于 `incomplete`。证据项由独立证据服务更新，服务每次保存后重新计算完整度：

- `complete` 和明确记录原因的 `waived` 都算已满足。
- 仍有 `missing` 或 `draft` 时保持 `incomplete`。
- 产品身份、出库称重/包装、张力表、人工签收/POD 等照片或扫描件证据，在标记 `complete` 前必须至少登记一个附件。
- 包级 `waived` 只是后台完整度和审计状态；它不能替代发货门禁要求。产品身份、出库称重/包装和必要的张力表在发货时必须是 `complete` 并且有附件。
- 附件存在只证明证据已上传，不证明照片中的实物、数值或人工结论合格。
- 物流供应商的 tracking、交付事件和 POD URL 只作为证据详情的只读上下文展示；它们不会自动完成或替代人工 `signed_pod` 证据。
- 全部满足时自动变为 `ready`。
- 只有 `ready` 才能锁定；锁定时写入 `locked_at`。

配置确认项是订单创建时的快照投影，不能被后台改写。产品身份、张力表、出库包装和 POD 属于可补录的履约记录；包锁定后这些记录也不能再覆盖。附件引用同样不能在锁定包上继续追加。

锁定包需要修订时，创建新的 `package_version`，复制结构化证据和已登记的附件引用，旧版本保持 `locked` 不变。附件对象键按证据项唯一，而不是全局唯一，因此同一个已校验对象可以被新版本安全复用；对象本身仍不可更新或删除。

## 后台入口现状与设计顺序

当前后台有两个独立入口，但用途不同：

- 商品中心 / 产品履约要求：`/catalog/fulfillment-requirements`，维护商品级和 Variant 级 `spoke_tension_qc` 规则。
- 订单管理 / 履约证据：`/orders/evidence`，按订单读取证据包、检索完整度并补录后续证据。
- 订单管理 / 发货动作：订单列表或订单详情中的“发货”按钮打开履约取证弹窗；该弹窗是发货前的主要取证入口，不要求操作人员先在证据 TAB 中手工添加订单号。

发货前还必须完成订单行级报关申报确认：

- 每个 `order_items` 行都必须有大于 `0` 的 `declared_value`，并且 `declared_value_confirmed=true`。这两个字段是订单行最终申报值的确认状态，不是可选备注。
- 商品或 Variant 的 HS/CN 编码、原产国、报关描述和分类资料，只提供报关基础资料，不会自动替代订单行最终申报价值确认。后台通过 `PATCH /api/admin/orders/:id/items/:item_id/customs` 写入或清除该确认。
- `POST /api/admin/orders/:id/fulfillment` 在事务内锁定订单并逐行校验申报值；任一订单行缺失、未确认或小于等于 `0` 时，发货、物流写入和发货通知均不发生。错误码为 `order_customs_declared_value_incomplete`。
- `GET /api/admin/orders/:id/customs-export` 使用同一校验门槛；申报不完整时返回错误，不生成含空 `Declared Value` 的 CSV。

发货和物流纠正是两个不同的订单动作：

- `POST /api/admin/orders/:id/fulfillment` 只用于首次确认发货。已发货订单重复提交相同物流时幂等成功，提交不同物流时返回冲突；这个冲突只防止重复发货，不阻止后续纠正错填的物流号。
- `PATCH /api/admin/orders/:id/tracking` 是明确的物流纠正入口。它更新订单当前物流和当前唯一 tracking shipment，不重新发货，也不改写原始 `shipped_at`。
- 物流来源变化时，先前 shipment 的 tracking events 不再进入当前订单证据组装结果；当前实现会清理当前事件集，避免错误物流轨迹继续进入证据包，新物流号需要重新同步后才产生新的当前轨迹。
- 物流纠正必须记录 before/after 审计。审计只保存订单、物流标识、状态和操作来源等元数据，不保存供应商密钥、Webhook Secret 或证据文件内容。
- 若已经生成 Stripe/PayPal 渠道提交快照，物流纠正不能覆盖该快照。当前系统尚未把“创建订单证据包修订版、创建新的渠道快照并再次调用支付渠道”串成可承诺的端到端流程；不要把物流纠正或证据包修订接口当作成功提交后的拒付重提入口。

### 物流纠正审计契约

物流纠正由 `go-backend/internal/api/admin/order_handler.go` 的
`PATCH /api/admin/orders/:id/tracking` 处理，审计投影位于
`go-backend/internal/api/admin/order_tracking_audit.go`。它与首次发货的
`order_fulfillment` 审计分开，避免把“纠正物流”误记成“再次发货”。

固定审计契约如下：

```text
resource: order_tracking
resource_id: order_id
action: update
status: success | failed
```

成功或业务失败时，`changes` 只允许包含：

```text
tracking_number
tracking_provider_id
carrier_id
carrier_service_id
operation = tracking_correction
```

`old_value` 和成功时的 `new_value` 只允许包含当前订单物流元数据：

```text
order_number
status
shipping_status
tracking_number
tracking_provider_id
carrier_id
carrier_service_id
provider_carrier_code
shipped_at_present
```

请求 JSON 无效时，审计使用 `changes.request_valid = false`，并只记录可读取到的
旧订单状态。审计不会记录 API Key、Webhook Secret、证据原文、附件内容、对象存储
Key 或先前物流轨迹正文。先前 tracking events 被清理后，审计中的 `old_value` 只证明
纠正前的物流元数据，不等同于保存事件历史。审计写入失败由审计基础设施记录，
不能回滚或阻断已经成功完成的物流纠正。

订单证据入口依赖以下前置条件：

1. 订单创建时已经保存统一的产品和配置快照。
2. 普通商品、组件、轮组和混合订单都有稳定的订单行关联。
3. 高价规则只使用订单最终总额和 USD 等值快照。
4. 张力表规则只使用产品/Variant 履约要求快照。
5. 附件上传已有独立的存储、哈希、权限和所有权校验。

履约证据 TAB 交互建议：

- 订单证据 TAB 面向所有订单，采用分页和筛选，不一次读取全部订单。
- 通过订单号、客户、状态、金额、是否高价、是否需要张力表和证据完整度检索订单。
- 打开订单后由独立履约证据视图读取该订单证据包，用于查看、复核和补录。
- 发货前的产品照片、出库称重/包装照片和必要的张力表，应从订单的“发货”按钮进入履约取证弹窗完成。
- “补录证据”使用独立弹窗，按证据项类型显示结构化字段和附件上传；它不替代发货按钮触发的发货前取证流程。
- 弹窗不允许创建没有 `order_id` 的孤立证据包。
- 锁定后的证据包以新版本修订，不能静默覆盖已用于拒付提交的版本。

## 与拒付工作台的集成

订单证据包是订单域的事实源，Stripe/PayPal 拒付工作台是支付渠道的提交适配层：

```text
order / order_item snapshots
  + order evidence items / attachments
  + current shipping shipment / tracking events
  -> read-time order evidence package assembler
  -> evidence readiness check
  -> Stripe or PayPal field mapper
  -> human review or provider submission
  -> submission audit
```

拒付工作台可以显示缺失项和完整度，但不应在提交时临时创造配置确认单、唯一编码或张力表。客户是否拒付不决定订单是否应该留存基础履约证据。

### 统一组装规则

订单证据、附件和物流仍然分域保存，查看时才按订单实时组装。当前统一组装器是
`go-backend/internal/service/order_evidence_package_assembler.go`，它只负责读取和拼接，不上传文件、不修改源记录，也不负责 Stripe/PayPal 字段转换。

统一组装器的固定顺序和边界：

1. 读取订单和订单行快照。
2. 读取订单最新的非废弃证据包，并带出证据项、附件引用和 SHA-256。
3. 读取订单当前唯一 shipment。
4. 只有 tracking number 与当前 shipment 一致的 tracking events 才进入当前物流上下文；没有当前 shipment 时，历史 tracking events 一律不作为当前妥投事实。
5. 输出来源 ID、证据完整度和缺失警告；物流供应商配置、API Key、Webhook Secret 等原始敏感对象不进入对外结果。

后台订单证据详情和 Stripe/PayPal 拒付包都使用同一组装结果。支付服务只把组装结果转换成渠道字段，不能再单独查询 shipping repository，也不能把拒付包当作证据的唯一存储。

实时查看允许看到最新同步到的物流事实。首次拒付提交前，支付适配器先把统一组装结果、渠道字段、来源 ID、附件 SHA-256 和物流上下文写入 `order_evidence_submission_snapshots`，计算 SHA-256 后锁定，再调用 Stripe/PayPal。外部提交失败后的重试必须先读取同一渠道、同一 dispute 的锁定快照，不能重新读取订单、证据、物流、客服沟通，也不能重新生成或上传 PayPal 商业发票；本次请求传入的新备注、文件 ID 或其他证据字段不覆盖锁定版本。提交审计和返回结果记录快照 ID、版本和 SHA-256。这个已实现的重试复用只覆盖同一锁定快照，不代表成功提交后可以自动创建修订版并再次提交；后者当前尚未形成受支持的端到端流程。

普通订单证据导出使用独立的 `order_evidence_export_snapshots`，不与 Stripe/PayPal dispute 提交快照共用记录。后台入口为 `GET /api/admin/orders/:id/evidence/export`，仅允许当前证据包为 `locked` 时首次生成；首次生成会冻结订单、订单行、证据包、证据项、附件元数据与哈希，以及当时匹配的 shipment/tracking events。后续重复导出复用同一订单证据包版本的锁定 JSON 和 SHA-256，不重新读取实时证据或物流。创建新的证据包修订版后，才允许产生新的导出快照。导出清单只包含附件 ID、元数据和哈希，不包含对象存储 Key；附件仍通过现有认证接口访问。

## 分阶段实施与当前状态

以下步骤记录实际落地顺序和验收状态，不再表示“先完成快照、再开始 TAB”的当前待办顺序。
订单快照、独立证据域、履约节点、后台入口、拒付复用和导出已经在同一套订单证据域上
协同运行；后续只补边界测试和维护性拆分，不重新规划一套平行功能。

### 第 0 步：基础文件体检

- [x] 保持现有 `Orders.vue`、订单 handler 和支付拒付服务的现有职责。
- [x] 为订单金额、订单行和产品履约要求读取补充边界测试。
- [x] 把现有 `SignatureRequired` 使用的 750 USD 判定抽象成唯一的高价订单策略；证据包和发货签名都读取同一个结果，不各自复制阈值计算。
- 订单创建服务中与快照构建无关的后置动作仍保留为后续维护性拆分，不作为证据 TAB 或拒付复用的前置阻断。
- [x] 不把规则、附件上传和订单证据编排继续堆进旧的订单页面；后续能力继续沿独立入口扩展。

### 第 1 步：订单快照契约

- [x] 保存订单级和订单行级产品、规格、价格、重量和产品履约要求快照。
- [x] 固化订单最终总额、USD 等值金额和金额规则版本。
- [x] 为快照增加 schema version 和 SHA-256。
- [x] 不改 `QUICKBUY` 领域；所有订单入口统一进入同一套订单快照契约。

### 第 2 步：独立订单证据域

- [x] 新增独立 domain、repository、service、handler 和迁移文件。
- [x] 由后端根据订单事实计算基础、高价值和轮组条件证据。
- [x] 支持结构化记录和附件引用；后台履约证据页面已接入，复杂拒付适配仍保持独立。

### 第 3 步：履约节点接入

- [x] 生产开始/完成接口只更新生产状态，不读取证据包，不接收张力表字段，也不因证据缺失或人工结论阻断生产。
- [x] 点击订单“发货”后先打开履约取证弹窗；所有订单行都必须上传产品身份照片，必要时同时上传张力表和出库称重/包装照片，证据保存完成后才执行发货。
- [x] 张力表可以绑定订单行或装配组；人工录入的测量值、范围、工具、时间和结论原样保存，不做自动质量判断。
- [x] 照片/扫描件类证据只有在附件已登记后才能标记为 `complete`；这只是证据完整性校验，不是质量验收。
- [x] 出库时记录重量、包装和照片。结构化记录至少包含 `gross_weight_g`、`package_count`、`packaging_method`；照片或扫描件仍通过独立附件引用保存。
- [x] 发货时保存 tracking 上下文；运输签收/POD 无法在发货时取得，必须在后续签收后补录。人工 POD 至少记录 `tracking_number` 和 `delivered_at`；当前 shipment 的 tracking 只用于显示 `matched`、`mismatch`、`not_recorded` 或 `shipment_unavailable` 上下文，不自动完成 POD，也不阻塞发货。
- [x] 为出库、生产、发货和签收节点补充统一的节点审计记录；审计只保存订单、物流、证据项状态和操作来源等元数据，不保存证据原文、文件内容、对象存储 Key、API Key 或 Webhook Secret。
- [x] 物流号填错时通过独立 `PATCH /api/admin/orders/:id/tracking` 纠正；首次发货的冲突保护不能阻止纠正，纠正后先前 shipment 的事件不得进入当前证据视图。

### 第 4 步：后台入口和弹窗

- [x] 已新建独立 API 客户端、类型、订单证据页面/弹窗，以及产品履约要求页面/弹窗。
- [x] 订单号选择读取真实订单并检查权限。
- [x] 订单“发货”按钮打开的弹窗统一录入身份、包装和必要的张力表证据；上传经过证据附件服务，不能由前端直接拼接存储 URL。
- [x] 独立履约证据 TAB 用于跨订单检索、查看和补录，包括发货后的人工 POD。

### 第 5 步：拒付复用和导出

- [x] Stripe/PayPal 工作台通过统一组装器读取订单证据包、附件哈希和当前物流上下文。
- [x] 按支付渠道限制生成提交字段和附件。
- [x] Stripe/PayPal 首次拒付提交前创建并锁定不可变提交快照，快照 envelope 与渠道 request 一起参与 SHA-256。
- [x] Stripe/PayPal 失败重试复用同一锁定快照；PayPal 商业发票 URL 只在首次创建快照时生成并上传一次。
- [x] 提交结果和本地提交审计记录快照 ID、版本、SHA-256、package version、来源 ID、附件 SHA-256 和物流事实时间点。
- [x] 证据导出入口已接入独立的订单证据导出快照服务；导出只允许锁定包，首次生成后复用同一包版本的不可变 JSON 和 SHA-256。
- [ ] 成功提交后创建证据包修订版、生成新的渠道提交快照并再次提交：当前尚未实现，不能写成现有运营能力。

## 回归测试覆盖

物流纠正和首次发货保护已通过服务层、组装器和后台 handler 测试固定：

- `go-backend/internal/service/order_fulfillment_idempotency_test.go`
  - `TestOrderServiceFulfillOrderRetriesSameShippedTrackingWithoutChangingShipment`
    固定相同物流重复发货可以幂等成功，并且不会创建第二个 shipment。
  - `TestOrderServiceFulfillOrderRejectsDifferentTrackingAfterShipment`
    固定首次发货后通过 `POST /fulfillment` 提交不同物流必须返回冲突，原订单物流不变。
  - `TestOrderServiceCorrectsTrackingAfterShipmentAndKeepsFulfillmentIdempotency`
    固定先发货、再通过 `UpdateTrackingInfo` 纠正物流后：订单和当前 shipment 使用新物流，
    `shipped_at` 不变，先前 shipment 的事件不再保留为当前事件，新物流事件可以进入组装器，
    使用新物流重试发货幂等成功，使用旧物流重试仍然冲突。
- `go-backend/internal/service/shipping_tracking_service_test.go`
  - `TestUpsertTrackingShipmentResetsOperationalStateWhenSourceChanges`
    固定物流来源变化时 shipment 的同步状态、事件计数和旧事件集被重置。
- `go-backend/internal/service/order_evidence_package_assembler_test.go`
  - `TestOrderEvidencePackageAssemblerIncludesAttachmentHashAndCurrentTracking`
    固定组装器只使用当前 shipment 对应的物流事件。
  - `TestOrderEvidencePackageAssemblerDoesNotUseTrackingEventsWithoutShipment`
    固定没有当前 shipment 时，历史 tracking events 不得作为当前妥投事实。
- `go-backend/internal/api/admin/order_handler_fulfillment_audit_test.go`
  - `TestUpdateTrackingInfoRecordsTrackingCorrectionAudit`
    固定物流纠正写入 `order_tracking`、`update`、成功状态，以及纠正前后的物流号；
    同时验证审计不包含 `api_key` 和 `webhook_secret`。

本阶段已验证：

```text
go test ./internal/service -count=1
go test ./internal/api/admin -count=1
go test ./internal/api/v1/payment -count=1
npm run typecheck
git diff --check
```

## 明确禁止的做法

- 不把证据包写成 Wheels 专用能力。
- 不把张力表设置为所有订单统一必填项。
- 不因为订单金额超过 750 USD 就推断客户购买的是轮组。
- 不从当前商品标题、商品分类或当前产品规则反推历史订单配置。
- 不把 `SignatureRequired` 直接当成证据包状态；它们可以共享高价金额判定，但业务含义和后续写入位置要分开。
- 不在 `Orders.vue` 中直接实现证据规则、上传和支付提交。
- 不在 `order_handler.go` 中直接操作对象存储。
- 不让支付拒付服务成为订单证据的唯一数据库。
- 不创建没有订单关联的独立证据记录。
- 不让订单证据 TAB 与基础快照契约形成两套事实源；TAB 只读取订单证据域的订单快照和证据包。
- 不通过直接修改 `status` 绕过证据完整度计算或锁定检查。
- 不在生产完成、发货或证据保存流程中自动判定张力表“合格/不合格”。
- 不把人工录入的 `pass`、`fail` 或其他结论改写成系统结论。
- 不让后台详情、Stripe 和 PayPal 各自读取一套物流事件。
- 不在没有当前 shipment 时把历史 tracking event 当作当前订单的妥投证明。

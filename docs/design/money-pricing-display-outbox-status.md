# 金额、算价、展示与 Outbox 架构整改状态

更新时间：2026-09-23

本文记录最初提出的四项架构整改的实际落地状态。状态以代码和测试为准，不以历史规划文档中的“完成”描述为准。

## 1. Money 值对象

状态：核心完成，交易与后台金额输入已统一。

已落地：

- `go-backend/internal/domain/money` 提供不可变金额对象。
- 金额运算使用 minor integer，支持加法、整数乘法、比例乘法、精确汇率转换和 `Allocate` 分摊。
- 商品、购物车、Wishlist、订单、支付契约使用 `*_minor`。
- 商品、购物车、Wishlist 公开价格使用 `price_decimal` / `sale_price_decimal` 字符串。
- PayPal 发票预览在 API 边界解析 decimal 字符串，内部使用 Money。

边界说明：

- 核心交易与金额展示快照已不再暴露 `float64` 金额字段；展示快照唯一金额字段为按币种校验的 `amount_decimal` 字符串。
- 仍保留的 `float64` 仅限汇率、比例、评分、重量/尺寸、性能指标，以及支付/发票等外部 SDK 的 major-unit 边界，禁止回流交易域。

## 2. Pricing Pipeline 与分摊矩阵

状态：已完成。

已落地：

- `internal/domain/pricing` 提供行级快照、折扣分配和管道校验。
- Checkout 已按商品行生成基础价、折扣、税费、净额和总额。
- 会员、优惠券、积分折扣会记录为行级分摊。
- 订单行保存不可变 `pricing_snapshot`。
- 退款定价已读取订单快照，并有一致性测试。
- 退款积分逆向结算支持封顶与债务追踪（Loyalty Points Debt）：当退款需扣回的已消费积分折算现金超出单次退款实付额时，现金扣减以本次退款金额严格封顶（实退扣减至 0 元），超出未追缴积分自动沉降为用户账户积分债务（`loyalty_points_debt`），由后续赚取积分自动抵扣清偿，彻底消除硬抛错导致的合法退款物理死锁（已由 `TestPaymentServiceRefundExecutionCapsLoyaltyCashRecoveryAndTracksDebt` 定向测试覆盖）。

剩余：

- 业务代码已完成；缺失快照拒绝、行级退款上限、重复/超额退款均由现有退款测试覆盖。
- 已提供只读历史审计命令 `adminctl audit-pricing-snapshots`：按批读取订单与订单行，校验快照 JSON、schema、货币、行级金额和订单汇总；命令只输出文本/JSON 报告，不执行补数、修复或写入。
- 运行验收仍需在生产副本或受控只读连接上执行该命令，确认历史订单数据迁移后所有可退款行均有不可变快照；任何补数或修复必须另行评审，不由审计命令自动完成。

## 3. 展示态与结算态分离

状态：基础完成，后台编辑金额输入已完成清理。

已落地：

- 商品和变体交易价格使用源币种 minor 字段。
- 展示价格使用独立快照模型和刷新服务。
- 展示 API 与结算 API 已区分 `display_price` 和 `price_decimal`。
- 展示刷新不应更新核心商品交易价格。

剩余：

- 清理旧的浮点展示 DTO已完成；展示快照、SEO、后台报表、Merchant 和客服上下文均使用 decimal 字符串。
- 后台商品编辑表单金额现在使用 decimal 字符串，提交边界统一转换为 `price_minor` / `sale_price_minor`，不再使用 `v-model.number` 或金额 `number` 类型。
- SEO 商品结构化数据预览的 `offers.price` 已改为 decimal 字符串；客服上下文金额此前已统一为 decimal 字符串。
- 商品领域 `DisplayPrices` 与公开商品/推荐投影已统一返回 decimal 字符串，移除了该边界的 `float64` 金额模型。
- 展示价格刷新已修复快照读模型初始化与 decimal 回填问题；汇率转换保持 Money/big.Rat 精确计算，最终仅在 JSON 兼容边界写入字符串。
- Storefront SEO 输入、结构化数据和 SEO 输出测试已统一使用 decimal 字符串价格。
- Google Merchant 价格解析、汇率转换和 micros 生成已改为 Money/big.Rat；营销促销风险报表的金额字段已改为 decimal 字符串，避免统计展示再暴露金额 `float64`。
- 商品搜索价格筛选已改为 minor-unit 整数查询参数（`price_min_minor` / `price_max_minor`），移除搜索服务边界的金额 `float64`。
- 商品变体 `EffectivePrice` 已改为 decimal 字符串；搜索前端同步提交 minor-unit 范围参数。
- 商品搜索筛选现在同时携带 `price_currency`，后端仅在同币种下直接比较 `price_minor`，不再做 major 除法或固定除以 100。
- Google Merchant 管理端覆盖价编辑已改为 decimal 字符串，提交时唯一转换为 `price_override_minor` / `sale_price_override_minor`。
- SEO 管理端结构化数据价格与营销风险金额 DTO 已与后端 decimal 字符串契约对齐；优惠券和 PayPal 发票预览金额输入不再使用 `v-model.number`。
- PayPal 发票预览的行小计边界改用十进制字符串处理，避免编辑态金额直接进入二进制浮点乘法。
- 展示缓存刷新与交易表写入隔离已由 CAS 和定向测试验证：源价格变化时旧快照不会覆盖，刷新不会修改 `price_minor`。
- 运费报价 `ShippingQuote`、`ShippingQuotePlan`、`ShippingQuoteItem`、`ShippingQuoteLeg` 的展示金额已统一为固定格式 decimal 字符串；交易计算和分摊继续只使用 Money/minor 字段。
- 已删除运费报价旧的 major-unit `float64` 包装入口及 `roundMoney` 测试辅助，避免金额在业务层发生浮点往返。
- 商业发票 `LineItem` 与 `CommercialInvoice` 金额快照已统一为 decimal 字符串；订单构建和 PDF 渲染直接使用 `Money.FormatMajor()`，不再经过 `MajorFloat()`。
- PayPal 商业发票预览 API 的行项目和汇总金额只接受 decimal 字符串，内部通过 `Money.ParseMajor` 校验并计算，不保留 JSON 数字兼容层。
- 订单 CSV 导出和 quick-buy 产品/变体快照的展示金额已改为 decimal 字符串，避免历史展示投影重新暴露金额 `float64`。
- 推荐商品价格直接消费商品领域的 decimal 展示值，删除旧的浮点价格格式化辅助。
- 支付网关适配器（PayPal、Stripe、支付宝、微信及 mock）响应金额统一由 `Money.FormatMajor()` 生成 decimal 字符串；删除 `Money -> float64 -> string` 的内部往返和 `paymentMajorFloatFromMinor`。
- PayPal 风险 webhook 的争议金额只接受供应商 decimal 字符串并通过 `Money.ParseMajor` 校验；不再把 JSON 数字金额转成 `float64`。

## 4. Transactional Outbox 与 Saga

状态：基础设施和主要业务路径已完成，代码级重试与幂等验证完成，剩余为运行验收审计。

已落地：

- SQL Outbox、Repository、Dispatcher、重试、锁和死信状态已存在。
- 订单状态、支付退款、邮件、发货、商品缓存和客服事件已接入主要 Outbox handler。
- 生产配置包含 Outbox dispatcher 的开关、批量和锁超时校验。
- 幂等中间件、支付操作幂等、Outbox 重试/认领、HTTP 瞬态错误重试及邮件 Outbox 失败重试测试均已通过。

剩余：

- 业务代码已完成：订单退款通过本地 execution intent + Outbox dispatcher 执行，含重试、幂等和失败状态；邮件、订单状态、客服实时事件、发货注册和商品缓存均通过 Outbox handler 投递。
- 仅剩部署验收：真实支付供应商 3DS 回调、供应商超时重试、重复 webhook 幂等、退款失败恢复和邮件 Outbox 投递。

## 当前优先级

1. 部署验收：验证真实支付供应商的 3DS 回调、超时重试和幂等行为。
2. 历史数据验收：确认迁移后的历史订单可退款行均具备不可变 `pricing_snapshot`。
3. 展示投影：金额相关展示和 Google Merchant 适配已统一为 Money/decimal 字符串；非金额比例、评分、性能指标保留 `float64`，不属于 Money 范围。支付供应商 SDK 要求的 major-unit 响应/请求边界仍保持适配函数隔离，不回流交易模型。

## 当前验证边界

- `go build ./internal/service` 已通过，说明本轮运费报价生产代码可编译。
- `go test ./internal/domain/currency`、`go test ./internal/api/v1/product` 为定向通过范围。
- `go test ./internal/service -count=1` 已通过，包含运费报价快照清理、退款和订单金额相关测试。
- 发票与后台支付管理定向测试：`go test ./internal/pkg/invoice ./internal/api/admin -run 'CommercialInvoice|Invoice|OrderExport' -count=1` 已通过。
- `go test ./internal/pkg/payment -count=1`、`go test ./internal/api/v1/payment -count=1` 已通过。
- `go test ./... -count=1` 已通过；测试中的数据库 `record not found` 日志来自预期的缺失数据分支，不是失败。

### 历史快照只读审计

```text
cd go-backend
go run ./cmd/adminctl audit-pricing-snapshots -config path/to/config.yaml -format json > pricing-snapshot-audit.json
```

默认按 200 个订单一批读取，最多保留 1000 条问题明细；可用 `-max-orders`、`-from-order-id`、`-to-order-id` 和 `-max-issues` 做分段审计。`missing`、`invalid` 和 `mismatch` 仅表示发现，需要人工确认后再制定迁移方案。

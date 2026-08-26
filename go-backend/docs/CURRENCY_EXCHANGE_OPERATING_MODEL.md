# 币种与汇率运行规则

这份文档是后台币种、市场展示币种和汇率 API 的唯一说明。后续排查问题时，先按这里的定义判断，不要把“后台录入币种”和“前端展示币种”当成同一个配置。

## 一、四个概念必须分开

### 1. 后台录入币种

- 设置项：`currency_primary_currency`
- 位置：后台“币种与汇率”页面
- 用途：管理员录入商品、SKU、运费以及其他商业金额时使用的币种
- 示例：设置为 `CNY`，就表示后台新录入的商品金额按人民币填写
- 这不是前端用户看到的地区展示币种

### 2. 商品源金额

商品或 SKU 保存的 `currency`、`price`、`sale_price` 是商品目录的真实金额。

运费模板、运费规则和承运商线路也保存自己的源币种。新建记录未填写源币种时，服务层读取当前 `currency_primary_currency`；编辑历史记录时如果请求未提交币种，则保留原币种。

汇率同步不得修改：

- 商品 `currency`、`price`、`sale_price`
- SKU/Variant 的源币种和源金额
- 订单金额和订单币种
- 支付、退款金额和币种

### 3. 前端展示币种

- 位置：后台“市场与本地化语种”TAB
- 配置字段：市场的 `default_currency` 和 `display_currencies`
- 用途：决定不同国家/地区的 Nuxt 前端默认显示币种，以及用户可以切换的币种
- 示例：
  - US 市场：`USD`
  - EU 市场：`EUR`、`USD`
  - UK 市场：`GBP`

前端展示币种不在全局“后台录入币种”页面新增。

### 4. 汇率缓存

- 存储表：`currency_exchange_rates`
- API base：当前后台录入币种
- API quote targets：所有启用市场的 `default_currency` 和 `display_currencies` 去重后的集合
- 作用：为前端展示、运费展示和后台预览提供缓存汇率

## 二、当前采用的切换策略

系统采用：

> **保留原有商品金额 + 检测币种不一致 + 人工确认修正**

当后台录入币种从 `CNY` 改为 `USD` 时：

1. 系统只把全局录入币种改为 `USD`。
2. 历史商品仍保留原来的 `currency` 和金额，例如 `CNY 699`。
3. 系统不会把数字 `699` 静默理解成 `USD 699`。
4. 系统不会自动批量换算商品、SKU、订单或支付金额。
5. 后台自动检测商品和 SKU 的保存币种是否与当前后台录入币种一致。
6. 管理员确认实际金额后，再人工修改不一致的商品/SKU。
7. 汇率同步随后以新的后台录入币种 `USD` 为 base，重新更新前端展示所需的汇率缓存。

这样做是为了避免把历史金额错误解释成新币种。自动批量转换以后如果要做，必须单独做成明确的迁移工具，不能和日常汇率同步混在一起。

## 三、实际例子

初始配置：

- 后台录入币种：`CNY`
- 商品价格：`CNY 699`
- 市场展示币种：`USD`、`EUR`、`GBP`

每天同步时：

```text
API base = CNY
API targets = USD, EUR, GBP
商品源金额仍然是 CNY 699
前端根据缓存汇率展示对应地区的 USD/EUR/GBP
```

后来把后台录入币种改成 `USD`：

```text
当前后台录入币种 = USD
历史商品仍然是 CNY 699
后台检测结果 = 该商品币种与当前录入币种不一致
```

此时不能把 `699` 直接当作 `USD 699`。管理员应先确认商品真实价格，再手动改商品/SKU 的币种和金额。确认后再同步：

```text
API base = USD
API targets = 启用市场中的展示币种
```

## 四、每日同步流程

1. 后台读取 `currency_primary_currency`，作为 API base。
2. 后台读取启用市场的 `default_currency` 和 `display_currencies`，作为 quote targets。
3. 后端调用 ExchangeRate-API。
4. 把汇率写入 `currency_exchange_rates`。
5. 使用最新汇率重建商品和 SKU 的 `display_prices` 展示快照。
6. 只有商品/SKU 源币种与当前后台录入币种一致时，才重建对应快照。
7. 币种不一致的商品/SKU 保留原有源金额和原有展示快照，并继续出现在检测结果中，等待人工修正。
8. 前端商品接口读取已保存的展示快照，不在请求时直接调用第三方汇率 API。
9. 调度器启动时执行一次，之后默认每 `86400` 秒执行一次。
10. 管理员手动点击“同步汇率”时，调用同一个 `ExchangeRateService.Sync()` 函数。
11. 运费模板/规则的源币种必须和报价币种一致；发现历史运费币种不一致时，报价明确失败，不进行静默换算。

定时同步与手动同步共用同一条服务路径，避免两套逻辑产生不同结果。

## 五、币种不一致检测

检测接口：

```http
GET /api/admin/settings/currency-policy/audit
```

检测内容：

- `products.currency`
- `product_variants.currency`

返回内容包含：

- 当前期望币种
- 商品不一致数量
- SKU 不一致数量
- 不一致总数
- 一组商品/SKU 示例，方便管理员定位

保存新商品和新 SKU 时，系统会按当前后台录入币种校验；历史数据不会因为修改全局币种而被自动重写。

## 六、函数职责边界

- `CurrencyPolicyService.BackendEntryCurrency()`
  - 读取后台唯一录入币种。
- `StorefrontMarketService.ListStorefrontDisplayCurrencies(true)`
  - 读取启用市场的前端展示币种。
- `ExchangeRateService.GetConfig()`
  - 组合 API base 和 quote targets。
- `ExchangeRateService.Sync()`
  - 调用第三方 API，写入汇率缓存，并触发展示价格快照刷新。
- `ProductService.RefreshDisplayPriceSnapshots()`
  - 根据当前源币种和缓存汇率刷新商品/SKU 的 `display_prices`，不修改源币种或源金额。
- `ProductRepository.UpdateDisplayPriceSnapshots()`
  - 只更新展示快照 JSON；通过显式标志避免误清空币种不一致商品的旧快照。
- `ProductService.AuditBackendEntryCurrencyConsistency()`
  - 检测商品/SKU 源币种与后台录入币种是否一致。
- `ShippingService`
  - 为新增运费金额补齐后台录入币种，保留历史记录的显式源币种，并拒绝模板/规则之间的币种分裂。
- `ExchangeRateSyncScheduler`
  - 启动时和按间隔调用 `ExchangeRateService.Sync()`。

## 七、明确禁止的做法

以下逻辑不能放进每日汇率同步：

- 修改商品或 SKU 的 `price`
- 修改商品或 SKU 的 `currency`
- 把币种不一致的商品/SKU 快照按新 base 强行重算
- 修改订单、支付或退款金额
- 把前端展示币种保存回商品源金额
- 把市场 TAB 的币种复制到全局后台录入币种
- 在 Nuxt 前端直接调用第三方汇率 API
- 把 API Key 暴露给 Nuxt 前端

如果未来要实现“把整个商品目录从 CNY 迁移到 USD”，必须另建迁移流程，至少包含：

- 明确的迁移确认
- 固定汇率时间点
- 舍入规则
- 预览结果
- 操作日志
- 失败处理和回滚方案

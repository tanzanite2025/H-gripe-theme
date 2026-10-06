# 物流服务集合与运费模板数据契约

## 职责边界

- 4PX、燕文域负责各自官网接口、环境配置和官方服务目录。
- 物流管理负责承运商、运费模板、模板规则及线路计费参数。
- 服务集合是物流管理读取官方线路身份和发布状态的边界，不负责维护运费模板的价格规则。
- 新报价和前台公开线路读取当前已发布集合；已生成的报价快照保留生成时的线路与价格数据。

## 当前数据链

```text
4PX 环境配置
  -> 对应环境的官网服务同步
  -> shipping_fpx_channels(environment, service_code)
  -> 生产环境已启用集合
  -> 物流管理运费模板选择（shipping_carrier_services.fpx_channel_id）
  -> 新报价与前台公开线路投影
```

4PX 测试与生产目录按 `environment + service_code` 隔离。4PX 域可查看和审批两个环境的记录；模板校验、概览统计、新报价和公开线路仅读取生产环境。测试环境同步不会更新生产环境同代码渠道的名称、地区或启用状态。

```text
燕文环境配置
  -> 对应环境的官网产品目录
  -> 燕文精选服务集合(environment, product_code)
  -> 生产环境已启用集合
  -> 物流管理运费模板选择（shipping_carrier_services.yanwen_published_channel_id）
  -> 新报价与前台公开线路投影
```

燕文精选渠道以 `environment + product_code` 唯一标识来源。精选集合页面按 FAT/生产环境分别读取官方产品目录和渠道；FAT 精选渠道只供 FAT 环境的燕文真实运单使用。物流管理模板校验、新报价及前台线路投影仅读取生产环境已启用渠道。迁移前的精选渠道统一回填为生产环境，保持现有运费模板引用继续有效。

燕文真实建单同时校验所选精选渠道、产品目录、国家目录和交货仓均属于请求指定的环境，避免生产凭据使用 FAT 渠道或反向混用。
运单幂等键使用 `environment + order_id + product_code`，允许同一订单和产品在 FAT/生产分别测试，不会跨环境返回已有运单。

燕文建单读取订单时使用订单仓库提供的 `FindYanwenOrderFactsByID` 窄投影，只带入订单编号、支付/履约状态、币种、收件地址、付款日期和报关商品字段；燕文服务不接收完整订单聚合，也不读取通用订单备注、账单地址或其他订单关系。

燕文精选渠道还可以保存三个仅供燕文建单使用的渠道级规则：`require_receiver_tax_number`、`require_ioss` 和 `require_eori`。物流管理只读取渠道身份、地区和发布状态，不解释这些关务规则；燕文创建真实运单时由燕文域按所选渠道重新校验缺失字段。未配置规则时三个字段保持可选，且规则不写入通用关务表、不接入 4PX。

目录数据支持燕文专属定时同步：启用 `worker.yanwen_catalog_sync_enabled` 后，调度器按 `worker.yanwen_catalog_sync_interval_seconds` 分别刷新 FAT/生产的产品、国家和交货仓目录；失败只记录燕文域错误，不回退到其他物流域数据。

## 稳定关联规则

`shipping_carrier_services` 使用来源明确的可空字段关联服务集合：

- `fpx_channel_id` 只允许指向生产环境的 `shipping_fpx_channels.id`。
- `yanwen_published_channel_id` 只允许指向生产环境的 `shipping_yanwen_published_channels.id`。
- 同一线路不能同时填写两个字段；泛用承运商两个字段都为空。

物流管理运费模板选择器和独立线路编辑器保存集合记录 ID。后端保存时按 ID 校验承运商来源、生产环境、存在性和启用状态，再从集合回填线路代码、名称和配送地区。报价及前台公开线路投影也按 ID 读取最新集合数据，因此集合名称、代码或地区更新后会自动反映到新链路；集合停用、删除或旧线路未完成 ID 回填时，线路会关闭，不会按相同代码重新绑定到另一条集合记录。

物流管理的“线路服务” TAB 继续保留给通用线路能力和本地计费参数（首续重、体积重、附加费、时效、启停和追踪映射）。当承运商为 4PX 或燕文时，它只负责把已发布集合记录挂载到模板：线路代码、线路名称和配送地区均为集合投影，只读显示；需要更换身份或地区必须回到对应承运商域切换精选服务。通用承运商仍可在该 TAB 维护自己的线路代码、名称和国家范围。

迁移 `392_add_stable_published_collection_ids_to_shipping_carrier_services` 会依据承运商代码和原线路代码回填生产集合 ID。无法唯一回填的旧记录保留原线路数据，但保存、报价和前台投影会明确要求重新选择集合，不会静默删除。

## 维护规则

- 同步发现的 4PX 服务默认停用；人工审批状态只影响该环境下的记录。
- 只有生产环境启用的 4PX 服务可以进入物流管理运费模板。
- 燕文渠道审批状态只影响该环境；只有生产环境已启用渠道可进入物流管理运费模板和前台公开线路。
- 4PX、燕文集合中的地区和名称是新报价与公开线路的当前读模型；模板线路表中保留的地区、名称属于编辑/历史数据，不作为这两类承运商的新报价权威值。
- 泛用承运商继续使用物流管理中的线路配置，不从 4PX 或燕文集合读取数据。

## 跨域读取 DTO 与权限

物流管理读取燕文生产集合的接口为 `/api/admin/logistics/yanwen/collection`，响应是最小引用 DTO。管理端燕文页面的接口与类型位于 `src/api/yanwenLogisticsAdminApi.ts`，4PX 页面位于 `src/api/fpxLogisticsAdminApi.ts`，物流管理的最小集合投影位于 `src/api/shippingServiceCollectionReferenceApi.ts`，通用物流页面继续使用 `src/api/shipping.ts`；这些边界不通过同一个前端 API 门面互相调用：

```json
{
  "id": 12,
  "product_code": "481",
  "display_name": "燕文专线追踪",
  "countries": "[\"US\",\"CA\"]",
  "enabled": true
}
```

`package_type`、`max_weight_grams`、`volumetric_divisor`、`require_receiver_tax_number`、`require_ioss`、`require_eori` 和 `notes` 只通过燕文域的完整渠道接口读取，不进入物流管理的模板选择链。该集合读取允许 `shipping:view` 或 `logistics:yanwen:view`；燕文集合增删改仍需要 `logistics:yanwen:manage`。

燕文集合管理接口的完整列表、创建和更新响应使用独立的管理响应 DTO；该 DTO 只服务 `/api/admin/logistics/yanwen/channels` 页面，不直接复用持久化实体。跨域的 `/collection` 响应继续使用上面的最小引用 DTO，因此关务规则、包裹限制、备注和时间字段不会进入物流管理模板选择链。集合服务会在读取入口统一规范 `fat`/`production` 环境，空环境按生产处理，非法环境直接报错，避免内部调用跨环境读取。

通用物流服务通过 `YanwenPublishedCollectionReferenceReader` 只读接口消费生产集合引用；它没有燕文集合管理服务的 CRUD、目录、凭据、轨迹或运单方法，避免主运费域反向拥有燕文内部能力。

精选集合保存前必须在相同 `environment` 的官方产品目录中找到 `product_code`。目录同步后，官方响应中已经不存在的产品会自动停用对应精选渠道；停用渠道不会进入运费模板校验、报价或前台线路投影。

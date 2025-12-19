# Shipping Template Options API（后端已实现，供未来 UI 对接）

本文档用于记录：

- **单个 Shipping Template 返回多条可选物流线路（options）**的后端接口
- UI（商品详情页/结账页）未来应如何调用与展示
- 推荐的“用户选择结果”落库/传递方式（不做强绑定，便于后续迭代）

> 适用范围：`tanzanite-setting` 插件（WordPress）提供的 REST API。

---

## 1. 背景与目标

目前产品（Product）只保存一个 `shipping_template_id`（模板），模板内部包含多条 `rules`。

为了支持**用户可选物流方式**（例如同一个国家/重量范围下出现多条线路：Standard / Express 等），后端提供一个专门的接口：

- 输入：目的国家、邮编、重量、数量、金额等
- 输出：该模板在当前输入下**匹配到的多条线路 options**，按优先级排序

这样 UI 只需要渲染 options 列表让用户选择即可。

---

## 2. REST API：获取模板可选线路

### 2.1 Endpoint

`GET /wp-json/tanzanite/v1/shipping-templates/{id}/options`

示例：

```
/wp-json/tanzanite/v1/shipping-templates/12/options?country=US&zip=94016&weight=2.3&subtotal=120
```

### 2.2 Query 参数

- `country`：**必填**（国家/地区代码）。后端会 `strtoupper`，例如 `US`、`CA`。
- `zip`：可选（邮编）。
  - 如果某条 rule 配置了 `zip_ranges`，但请求未提供 `zip`，该 rule 会被跳过。
- `weight`：可选（重量）。当 rule.type 为 `weight` 或 `volume` 时使用。
- `quantity`：可选（购买数量）。当 rule.type 为 `quantity` 时使用。
- `items`：可选（件数）。当 rule.type 为 `items` 时使用；若未传，后端会 fallback 到 `quantity`。
- `amount`：可选（金额）。当 rule.type 为 `amount` 时使用。
- `subtotal`：可选（小计）。当 rule.type 为 `amount` 时优先使用 `subtotal`，fallback 到 `amount`。

> 注意：**只有当对应 rule.type 的数值存在**时，该 rule 才可能匹配。

---

## 3. 返回结构（Response）

成功时返回：

```json
{
  "ok": true,
  "data": {
    "template": {
      "id": 12,
      "template_name": "USA Shipping",
      "description": "...",
      "is_active": true,
      "meta": {
        "carrier": "..."
      }
    },
    "input": {
      "country": "US",
      "zip": "94016",
      "weight": 2.3,
      "quantity": null,
      "items": null,
      "amount": null,
      "subtotal": 120
    },
    "options": [
      {
        "key": "express",
        "service": "express",
        "service_label": "Express",
        "fee": 25,
        "priority": 100,
        "eta_min_days": 3,
        "eta_max_days": 5,
        "type": "weight",
        "min": 0,
        "max": 5,
        "free_over": null,
        "carrier": "...",
        "rule_index": 0
      },
      {
        "key": "standard",
        "service": "standard",
        "service_label": "Standard",
        "fee": 10,
        "priority": 50,
        "eta_min_days": 7,
        "eta_max_days": 12,
        "type": "weight",
        "min": 0,
        "max": 5,
        "free_over": 200,
        "carrier": "...",
        "rule_index": 1
      }
    ]
  }
}
```

### 3.1 options 字段解释

- `key`
  - 线路唯一 key。
  - 优先使用 rule 的 `service` 作为 key。
  - 若 rule 未配置 `service`，后端会使用 `rule_{index}` 作为 key。
- `service` / `service_label`
  - `service`：可机器识别的服务编码（例如 `standard` / `express`）
  - `service_label`：UI 显示名（例如 `Standard` / `Express`）
- `fee`
  - 运费（float）。
  - 若命中 `free_over` 且传入了 `subtotal` 且 `subtotal >= free_over`，则 `fee` 置为 `0`。
- `priority`
  - 线路优先级（int）。用于排序。
- `eta_min_days` / `eta_max_days`
  - 预计时效（可选）。
- `type` / `min` / `max`
  - 命中的 rule 匹配条件，用于 debug。
- `rule_index`
  - 命中的 rule 在模板 rules 数组中的索引，便于排查。

### 3.2 排序与去重规则（后端实现）

- **去重**：同一个 `service`（即同一个 `key`）出现多个命中时：
  - 优先保留 `priority` 更高的
  - `priority` 相同时，保留 `fee` 更低的
- **排序**（options 输出顺序）：
  1. `priority` DESC
  2. `fee` ASC
  3. `eta_min_days` ASC（未设置会按极大值处理）

---

## 4. zip_ranges 匹配规则

模板 rule 可配置 `zip_ranges` 数组，元素示例：

- 精确匹配：`"94016"`
- 区间匹配：`"94000-94999"`

后端匹配策略：

- 数字邮编：若 zip/start/end 都是纯数字，按数字区间比较
- 非纯数字：按字符串比较（大写 + 去空格后）

---

## 5. UI（未来商品详情/结账）建议调用方式

### 5.1 获取 options 的输入来源建议

- `country`
  - 来自收货地址国家
- `zip`
  - 来自收货地址邮编（若未填写，可能导致某些含 zip_ranges 的线路不出现）
- `weight`
  - 推荐：购物车 SKU/商品重量合计
- `quantity/items`
  - 推荐：购物车总购买数量/总件数
- `subtotal`
  - 推荐：商品金额或订单小计（不含税/含税按你的业务定义）

> 当前后端规则对 `amount/subtotal` 的处理是：如果传了 `subtotal`，优先用它。

### 5.2 UI 展示结构建议

- 将 `options` 渲染为 radio list 或 select：
  - title：`service_label`
  - subtitle：`eta_min_days`~`eta_max_days`（若存在）
  - price：`fee`（0 显示 Free Shipping）

### 5.3 “用户选择结果”建议传递/落库（不强制）

为了后续可追溯，建议订单/草稿里存：

- `shipping_template_id`（模板）
- `shipping_option_key`（= options[i].key）
- `shipping_fee`（= options[i].fee）
- 可选：`shipping_service_label`（= options[i].service_label）

> 说明：当前插件订单结构是否已经有对应字段，需要你后续做 UI 时再决定放在哪里。

---

## 6. 常见问题

### 6.1 为什么必须传 country？

规则目前按 `regions`（国家列表）做第一层过滤。

### 6.2 为什么有时 options 为空？

常见原因：

- 没传 `country`
- 模板 `is_active=false`
- rules 没有覆盖该国家
- rule 配了 `zip_ranges` 但你没传 `zip`
- rule.type 对应的值没传（例如 rule.type=weight 但请求没传 weight）

---

## 7. 关联实现位置（后端代码）

- 路由与匹配实现：
  - `wp-plugin/tanzanite-setting/includes/rest-api/class-rest-shippingtemplates-controller.php`
  - 路由：`/shipping-templates/{id}/options`

（本文件为对接说明文档，方便未来 UI 开发快速定位调用方式。）

# Shipping Templates 配送模板对接说明

面向前端（Nuxt / 其它消费者）的 Shipping Templates 接口与数据结构说明。

---

## 1. REST 路由

基于 `Tanzanite_REST_ShippingTemplates_Controller`：

- 列表：`GET /wp-json/tanzanite/v1/shipping-templates`
  - 返回：
    - `items`: Shipping Template 列表
    - `meta.rule_types`: 支持的规则类型（`weight`, `quantity`, `volume`, `amount`, `items`）
- 单条：`GET /wp-json/tanzanite/v1/shipping-templates/{id}`
- 创建：`POST /wp-json/tanzanite/v1/shipping-templates`
- 更新：`PUT /wp-json/tanzanite/v1/shipping-templates/{id}`
- 删除：`DELETE /wp-json/tanzanite/v1/shipping-templates/{id}`

> 认证：需携带 WP REST Nonce（后台已有 `TzShippingConfig.nonce`，前端可参考现有 admin JS 的 `apiRequest` 封装）。

---

## 2. Shipping Template 数据结构

接口返回的单个模板结构（`GET` / 创建 / 更新响应中）示例：

```jsonc
{
  "id": 1,
  "template_name": "默认模板",
  "description": "日本 / 美国默认运费规则",
  "is_active": true,
  "updated_at": "2025-01-01 12:00:00",
  "rules": [
    // 见下方「规则结构」
  ],
  "meta": {
    "carrier": "sf_express",   // 模板绑定的物流公司编码
    "currency": "CNY"          // 可选，最多 3 位大写字母
  }
}
```

### 2.1 meta 说明

- `carrier`（必选，推荐）
  - 类型：`string`
  - 存储为 `sanitize_key()` 之后的值，例如：`sf_express`, `dhl`, `ems`。
  - 由后台 Shipping Templates 表单中的「物流公司编码 (Carrier)」输入框写入：
    - DOM：`#tz-shipping-carrier`
    - JS：`payload.meta.carrier`

- `currency`（可选）
  - 类型：`string`
  - 会被转换为大写，仅保留 A-Z，长度最多 3（例如：`CNY`, `USD`）。

前端在消费模板时，通常需要：

- 根据 `meta.carrier` 选择对应的物流公司 / 渠道；
- 如需多币种显示，可以利用 `meta.currency` 决定展示单位。

---

## 3. 规则 `rules[]` 结构

每条规则在后端经过 `sanitize_shipping_rules()` 清洗，返回结构大致如下：

```jsonc
{
  "type": "weight",          // 规则类型：weight/amount/quantity/volume/items
  "min": 0,                   // 匹配区间下限（可为 null）
  "max": 10,                  // 匹配区间上限（可为 null）
  "fee": 50,                  // 运费金额（float）
  "priority": 0,              // 预留优先级，当前默认 0
  "free_over": 500,           // 满额包邮阈值（float，可为 null）
  "regions": ["JP", "US"]   // 适用国家代码数组（ISO，字符串）
}
```

> 注意：后端会对 `regions` 数组中的每一项做 `sanitize_text_field()`，空数组表示「所有国家」通用规则。

当前 admin UI 已支持的前端字段（在 `assets/js/shipping-templates.js` 中）：

- 规则表单 DOM：
  - `#tz-shipping-rule-type`      → `type`
  - `#tz-shipping-rule-service`   → 预留 `service`（尚未进入后端规则存储）
  - `#tz-shipping-rule-service-label` → 预留 `service_label`（尚未进入后端规则存储）
  - `#tz-shipping-rule-regions`   → `regions`（输入格式：`JP` 或 `JP,US`）
  - `#tz-shipping-rule-min`       → `min`
  - `#tz-shipping-rule-max`       → `max`
  - `#tz-shipping-rule-fee`       → `fee`
  - `#tz-shipping-rule-free-over` → `free_over`
  - `#tz-shipping-rule-eta-min`   → 预留 `eta_min_days`（当前后端未持久化）
  - `#tz-shipping-rule-eta-max`   → 预留 `eta_max_days`（当前后端未持久化）

前端保存逻辑（简化版）：

```js
const rule = {
  type: type,                // weight / amount / ...
  min: min,                  // number | null
  max: max,                  // number | null
  fee: fee,                  // number
  priority: 0,
  free_over: freeOver,       // number | null
  service: service || undefined,
  service_label: serviceLabel || undefined,
  regions: regions,          // ["JP", "US"]
  eta_min_days: etaMin,      // number | null（当前后端未保存）
  eta_max_days: etaMax       // number | null（当前后端未保存）
};
```

发送到后端时，`rules` 是上述对象数组，后端会过滤掉当前未支持的字段，仅持久化已知字段。

---

## 4. 前端运费计算建议流程

> 实际计算可以放在 Nuxt 前端，也可以放在 WordPress PHP 侧，这里给出一个纯前端的建议算法，方便实现。

假设你在前端已经有：

- `shippingTemplate`: 单个模板对象（含 `rules` 和 `meta.carrier`）；
- 请求参数：
  - `carrierCode`: 当前选用的物流公司编码；
  - `countryCode`: 收货国家代码（ISO，例如 `JP`）；
  - `weight`, `amount`, `quantity`, `volume`, `items`：根据不同类型传入实际值。

伪代码：

```ts
interface ShippingRule {
  type: 'weight' | 'amount' | 'quantity' | 'volume' | 'items';
  min: number | null;
  max: number | null;
  fee: number;
  free_over?: number | null;
  regions?: string[]; // ISO codes
}

interface ShippingTemplate {
  id: number;
  template_name: string;
  meta: { carrier?: string; currency?: string };
  rules: ShippingRule[];
}

function calcShippingFee(
  tpl: ShippingTemplate,
  params: {
    carrierCode: string;
    countryCode: string;
    weight?: number;
    amount?: number;
    quantity?: number;
    volume?: number;
    items?: number;
  }
): { fee: number; free: boolean; rule?: ShippingRule } {
  // 1. carrier 匹配（可选：如果你在前端按模板已经筛过 carrier，这一步可以省略）
  if (tpl.meta && tpl.meta.carrier && tpl.meta.carrier !== params.carrierCode) {
    return { fee: 0, free: true }; // 或者视为不适用
  }

  const valueMap: Record<string, number | undefined> = {
    weight: params.weight,
    amount: params.amount,
    quantity: params.quantity,
    volume: params.volume,
    items: params.items
  };

  const normCountry = params.countryCode.toUpperCase();

  // 2. 在 rules 中找到第一条「同时满足国家 + 区间」的规则
  for (const rule of tpl.rules || []) {
    const v = valueMap[rule.type];
    if (typeof v !== 'number') continue;

    // 国家匹配：regions 为空表示所有国家
    if (Array.isArray(rule.regions) && rule.regions.length > 0) {
      const regionMatch = rule.regions
        .map(c => String(c).toUpperCase())
        .includes(normCountry);
      if (!regionMatch) continue;
    }

    // 区间匹配
    if (rule.min != null && v < rule.min) continue;
    if (rule.max != null && v > rule.max) continue;

    // 命中规则
    const free = rule.free_over != null && params.amount != null && params.amount >= rule.free_over;

    return {
      fee: free ? 0 : rule.fee,
      free,
      rule
    };
  }

  // 3. 没有匹配规则的默认行为
  return { fee: 0, free: true };
}
```

> 注意：上面只是前端示例算法，具体业务可以按需调整（例如：多条规则取最低价 / 最高优先级等）。

---

## 5. 与后台 UI 的对应关系

- 后台「配送模板列表」与「编辑 / 新增」页面由 `includes/legacy-pages.php` 中的 `render_shipping_templates()` 输出：
  - 模板字段：`#tz-shipping-name`, `#tz-shipping-description`, `#tz-shipping-active`, `#tz-shipping-carrier`；
  - 规则编辑区域：`#tz-shipping-rules-list` + `#tz-shipping-rule-editor`。
- Admin JS：`assets/js/shipping-templates.js`
  - 负责：
    - 从 REST API 加载列表 / 单条模板；
    - 填充表单与规则编辑器；
    - 组装 `payload = { template_name, description, is_active, rules, meta }`；
    - 调用 `/shipping-templates` 的 `POST/PUT/DELETE` 接口。

前端（Nuxt / 其它）在对接时，**只要遵守上述数据结构和 REST 路由，即可无缝使用现有后台配置出的 Shipping Templates 数据来计算运费。**

---

## 6. 下单接口：前端建议传参

当前订单创建接口由 `Tanzanite_REST_Orders_Controller` 提供：

- `POST /wp-json/tanzanite/v1/orders`
  - 重要参数（见 `register_routes()`）：
    - `channel` (`string`)
    - `payment_method` (`string`)
    - `total` (`number`)
    - `subtotal` (`number`)
    - `discount_total` (`number`)
    - `shipping_total` (`number`)
    - `status` (`string`, 默认 `pending`)
    - `tracking_provider` (`string`)
    - `tracking_number` (`string`)
    - `items` (`array`, **必填**)

> 说明：当前版本的订单接口 **不会自动计算运费**，`shipping_total` 需要前端 / 业务侧根据 Shipping Templates 算好后再传入。

### 6.1 建议的前端下单 payload 结构

一个典型的下单请求体（示意）：

```jsonc
{
  "channel": "web",                // 订单来源：web / miniapp / app ...
  "payment_method": "stripe",      // 支付方式
  "subtotal": 1000,                 // 商品小计（不含折扣 & 运费）
  "discount_total": 100,            // 各种优惠金额合计
  "shipping_total": 50,             // 使用 Shipping Templates 算出来的运费
  "total": 950,                     // subtotal - discount_total + shipping_total
  "status": "pending",            // 可省略，后端默认 pending
  "tracking_provider": "",        // 下单阶段通常留空
  "tracking_number": "",          // 下单阶段通常留空
  "items": [
    {
      "product_id": 123,
      "sku_id": 456,
      "product_title": "商品标题（可选）",
      "sku_code": "SKU-001",
      "quantity": 2,
      "price": 500,
      "total": 1000,
      "meta": {
        // 这里可以按需扩展，与前端业务协定：
        "weight": 1.2,
        "shipping_template_id": 1,
        "shipping_rule_snapshot": {
          "type": "weight",
          "regions": ["JP"],
          "fee": 50
        }
      }
    }
  ]
}
```

后端会：

- 将 `shipping_total` 直接写入订单表 `tanz_orders.shipping_total`；
- `items[].meta` 会被 `wp_json_encode()` 成字符串存入 `tanz_order_items.meta`，**当前版本不会解析其中字段，仅原样保存**。

### 6.2 与 Shipping Templates 的对接约定

**推荐流程：**

1. 前端在结算页根据用户地址 / 选中的物流方式，先确定：
   - `shipping_template_id`
   - `carrierCode`（通常等于模板的 `meta.carrier`）
   - `countryCode`（收货国家 ISO 代码）
   - 用于匹配的值：`weight` / `amount` / `quantity` / `volume` / `items` 等。
2. 调用第 1 节/第 2 节描述的 Shipping Templates 接口，拿到对应模板数据。
3. 使用第 4 节的 `calcShippingFee` 思路，在前端算出：
   - 命中的 `rule`
   - 运费金额 `fee`
   - 是否包邮 `free`。
4. 在调用 `POST /orders` 时：
   - 把 `fee` 作为 `shipping_total` 传给后端；
   - 把 `subtotal`、`discount_total` 一并算好，`total = subtotal - discount_total + shipping_total`；
   - 如需记录调价依据，可把 `shipping_template_id`、`countryCode`、命中 `rule` 的关键信息，写进 `items[].meta` 中。

> 当前版本的订单 REST 接口 **没有暴露 order-level meta 字段**，也没有专门的 `shipping_template_id` 列。后续如果要在后端更智能地重算 / 回溯运费，可以扩展订单 `meta` 结构或增加专门字段；此处先通过 `items[].meta` 做前端自定义快照是最简单的做法。


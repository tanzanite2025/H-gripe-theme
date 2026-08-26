# 电商 URL 与 SEO 黄金架构

## 1. 文档状态

本文档记录 Tanzanite storefront 的正式 URL 与 SEO 架构。

- 状态：正式架构与实现契约。
- 分类和商品已经使用不同的路由前缀；实现、sitemap、FAQ lookup、推荐链接和缓存规则必须遵守本契约。
- `/shop/:slug` 只表示单级分类候选路径；未匹配到分类时直接返回 404，不尝试查询商品。
- 站点尚未运营，因此不预埋商品旧 URL、商品 alias 或自动 301。站点正式运营后产生的真实 URL 迁移，才由 URL Management domain 手工维护 301 规则。

## 2. 核心决策

采用“分类走层级，产品走扁平”的路由模型：

```text
分类聚合页：/shop/road-wheels
分类子页：  /shop/road-wheels/climbing
商品详情页：/products/hyper-45-disc-carbon-wheelset
```

分类 URL 负责表达目录层级和品类关键词。商品 URL 不携带当前父级分类，
只表达商品自身的稳定 slug。

页面上的面包屑层级不由商品 URL 的斜杠数量决定，而由页面运行时的分类
上下文和 `BreadcrumbList` JSON-LD 表达。

## 3. 路由模型

### 3.1 分类聚合页

分类允许多级嵌套：

| 层级 | 示例 | 责任 |
| --- | --- | --- |
| 一级分类 | `/shop/road-wheels` | 覆盖主品类关键词 |
| 二级分类 | `/shop/road-wheels/climbing` | 覆盖细分场景或用途 |
| 三级分类 | `/shop/road-wheels/climbing/aero` | 仅在有真实内容和商品规模时启用 |

分类页应当具备独立的页面标题、描述、主标题、商品集合和内部链接。
不能只创建一个空路径来承接关键词。

### 3.2 商品详情页

商品详情页统一使用：

```text
/products/:slug
```

示例：

```text
/products/hyper-45-disc-carbon-wheelset
```

商品 URL 规则：

1. 不包含父级分类 slug。
2. 不包含筛选条件、搜索词或排序参数。
3. 不因商品被移动到其他分类而改变。
4. 站点正式运营并产生外部 URL 后，商品 slug 变更才通过 URL Management domain 配置 301；运营前直接使用新的正式 URL，不创建旧 alias。
5. 同一商品在同一语言下只保留一个可索引的主详情 URL。
6. 语言版本可以有不同 slug，但每个语言版本都必须有稳定、真实存在的
   本地化 URL。

在当前 `prefix_except_default` locale 策略下，目标形态为：

```text
默认语言：/products/hyper-45-disc-carbon-wheelset
其他语言：/zh_cn/products/hyper-45-disc-carbon-wheelset
```

具体语言前缀必须来自共享 locale registry，不得在页面或 SEO 管理后台
手写一套语言列表。

### 3.3 查询参数和变体

商品主 permalink 是 `/products/:slug`。用于筛选、排序、弹窗、推荐来源
或临时 UI 状态的 query 参数不进入 canonical。

如果未来需要可分享的变体落地页，必须同时满足：

- URL 能稳定恢复同一个变体；
- 页面正文、价格、库存、媒体和结构化数据与该变体一致；
- 变体 URL 有明确的 canonical 策略；
- 该策略单独记录并通过 URL 管理流程验证。

在没有完成上述约束前，变体参数只能作为商品详情页的 UI 状态，不得扩展
为新的商品 permalink。

## 4. 为什么采用“分类深层 + 商品扁平”

### 4.1 商品 URL 具备长期稳定性

商品可以被重新归类、增加标签、参加促销或改变推荐位置，但商品本身的
详情地址不应随目录变化。

这样可以把外链、收藏夹、分享链接、历史索引和 canonical 信号集中在同一
个商品 URL 上，避免因为目录调整而批量制造重复页面或失效链接。

### 4.2 分类页可以自由表达目录层级

分类页承担品类词、用途词和场景词的 SEO 责任。分类可以继续向下扩展，
但扩展分类不应把商品 URL 绑定到某一条目录路径。

### 4.3 目录归属与商品身份解耦

一个商品可以拥有：

- 一个主分类；
- 多个辅助分类；
- 多个营销集合；
- 多个筛选标签；
- 多个推荐场景。

这些变化不应造成多个商品详情 URL。商品身份由商品 slug 和产品数据
决定，目录归属由分类关系决定。

## 5. 面包屑与 BreadcrumbList

### 5.1 页面展示层

商品页面可以在页面顶部展示面包屑，例如：

```text
Home / Shop / Road Wheels / Climbing / Hyper 45
```

展示层面包屑服务于用户导航，不要求与商品 URL 的路径结构一一对应。

### 5.2 JSON-LD 结构化数据

商品详情页必须在 SSR 输出有效的 `BreadcrumbList`。目标结构：

```text
Home
└── Shop
    └── Road Wheels
        └── Climbing
            └── Hyper 45
```

在 Nuxt 3 中，目标实现方式为：

```ts
useSchemaOrg([
  defineBreadcrumb([
    { name: 'Home', item: '/' },
    { name: 'Shop', item: '/shop' },
    { name: 'Road Wheels', item: '/shop/road-wheels' },
    { name: 'Climbing', item: '/shop/road-wheels/climbing' },
    { name: 'Hyper 45', item: '/products/hyper-45-disc-carbon-wheelset' },
  ]),
])
```

实际代码必须使用 locale-aware URL helper 和当前资源的真实分类数据，
上面的代码只表示结构，不是要求在页面中硬编码英文文案或路径。

### 5.3 面包屑数据规则

1. 商品最终项的 `item` 必须是商品扁平 URL。
2. 分类项必须指向真实存在的分类聚合页。
3. 多个分类都能归属商品时，选择一个稳定的 primary category 作为 SEO
   面包屑，其余分类仍可通过页面链接和产品关系表达。
4. 不得从 `/products/:slug` 反推分类层级。
5. 不得因为商品 URL 是扁平路径，就省略真实的分类层级。
6. JSON-LD 必须在初始 HTML 中输出，不能只在 hydration 后生成。
7. 分类移动后，面包屑可以改变；商品 canonical URL 不改变。

## 6. 数据与所有权

| 数据 | 事实源 | SEO 输出 |
| --- | --- | --- |
| 商品 slug | Product domain | 商品 canonical URL |
| 商品名称 | Product domain | H1、Product.name、面包屑末项 |
| 分类 slug 与 parent | Category/Product domain | 分类 URL、面包屑分类项 |
| 主分类关系 | Product/Category domain | primary breadcrumb |
| Meta Title/Description | SEO domain 或产品 SEO 字段 | HTML head |
| 商品价格与库存 | Product/variant/inventory domain | 可见价格、Offer |
| 面包屑 JSON-LD | Storefront SEO renderer | SSR JSON-LD |
| 301 重定向 | URL Management domain | 旧 URL 到新 URL 的迁移 |

SEO 管理后台可以编辑 Meta Title 和 Meta Description，但不得复制商品
slug、分类关系、价格、库存或 JSON-LD 作为第二份事实源。

## 7. Canonical、hreflang 与重定向

### 7.1 Canonical

- 分类页 canonical 指向当前语言的分类 URL。
- 商品页 canonical 指向当前语言的 `/products/:slug`。
- 商品不因为从某个分类进入，就 canonical 到带分类的 URL。
- query 参数默认不改变 canonical。
- 人工 canonical override 只能用于迁移或明确的异常场景。

### 7.2 Hreflang

- 每个语言版本只链接到真实存在的同一商品翻译。
- 不得通过替换 locale 前缀来猜测翻译 slug。
- 缺少翻译时不输出虚假的 hreflang。
- `x-default` 只指向真实存在并且已约定的默认市场 URL。

### 7.3 上线后的 URL 迁移

未运营阶段不需要兼容旧商品 URL：

1. 商品详情只使用 `/products/:slug`。
2. `/shop/...` 只用于分类层级；`/shop/:slug` 不是商品入口。
3. `/shop/:slug` 未匹配到真实分类时返回 404，不查询商品，也不自动 301 到
   `/products/:slug`。
4. 站点正式运营后，如果确实存在已经对外传播的旧 URL，再由 URL Management
   domain 配置单跳 301。该规则是人工迁移数据，不是商品服务、sitemap 或路由目录
   自动生成的 alias。
5. URL Management 规则发布后，验证不存在循环、链式跳转和把分类路径误判为商品
   路径。

## 8. 分类设计边界

分类层级应该由真实的商品目录和用户搜索意图驱动：

- 一级分类表达稳定的产品大类。
- 二级分类表达用途、骑行场景或明确的技术子类。
- 三级分类只有在拥有独立内容和足够商品覆盖时才创建。
- 空分类、只有一个商品的临时营销分类不应默认进入可索引层级。
- 分类排序、启用状态和本地化名称由分类事实源管理。

分类页可以使用筛选参数增强用户体验，但筛选结果页不应自动生成无限
可索引 URL。需要 SEO 的细分组合应升级为正式分类页，并拥有独立内容。

## 9. 文档与实现边界

代码和文档都以正式路由契约为准：

- `nuxt-i18n/app/pages/products/[slug].vue` 负责商品详情。
- `nuxt-i18n/app/pages/shop/[slug].vue` 只负责单级分类，并在非分类路径返回 404。
- `nuxt-i18n/app/pages/shop/[...category].vue` 负责多级分类。
- 商品 canonical、sitemap、FAQ lookup、推荐链接、聊天附件和 HTML cache 使用
  `/products/**`。
- 面包屑中的分类项使用 `/shop/...`，商品末项使用 `/products/...`。

SEO 后台只编辑 Meta Title / Meta Description，并展示只读的分类、面包屑和
结构化数据诊断；它不复制商品 slug 或分类关系。

## 10. 验收清单

至少需要验证：

- `/shop/road-wheels` 和 `/shop/road-wheels/climbing` 是可访问的分类页。
- `/products/hyper-45-disc-carbon-wheelset` 是唯一正式商品详情 URL。
- `/shop/non-category-product` 返回 404，不返回商品详情，也不产生自动 301。
- 商品从一个分类移动到另一个分类时，canonical 不变。
- 商品页初始 HTML 含有效 `Product` 和 `BreadcrumbList` JSON-LD。
- 面包屑分类项都指向真实分类 URL。
- 商品主 URL 不包含父级分类 slug。
- 商品路由目录不生成 `/shop/:slug` alias。
- hreflang 只引用真实存在的商品翻译。
- sitemap、FAQ lookup、推荐链接、聊天附件链接和 HTML cache 规则使用同一
  套最终路由契约。

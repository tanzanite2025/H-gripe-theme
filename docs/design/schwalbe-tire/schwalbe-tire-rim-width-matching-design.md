# Schwalbe 车圈内宽参考设计

> **状态：已实施。**
>
> 本文定义车圈内宽参考的用户语义、数据链路、文件职责和维护边界。当前选型页不把用户输入的车圈内宽作为筛选条件，而是在每条外胎卡片上展示按 ETRTO 公称胎宽得到的官方可能组合参考范围。

## 1. 目标

让用户先按熟悉的轮径 + BSD、型号和其他目录条件找到外胎，再在卡片上直接看到可供核对的车圈内宽范围，例如：

```text
车圈内宽参考  17–27 mm
```

这解决的是“这条外胎对应的官方胎宽—车圈内宽参考范围是多少”。它不把用户输入框和轮径、BSD及其他筛选条件叠加，避免隐藏条件造成结果为 0。

## 2. 数据语义

### 2.1 规则来源

迁移 362 的 `schwalbe_tire_rim_width_combination_rules` 表保存 Schwalbe/ETRTO 可能组合矩阵。每条规则包含：

- ETRTO 公称胎宽区间；
- 车圈内宽区间；
- 来源依据、版本和核验日期。

规则表示可能组合参考，不是某个具体型号、胎压、车圈系统的独立兼容认证。卡片和页面文案使用“参考”“可能组合范围”，不使用“绝对适应”“认证兼容”或“最佳内宽”。

### 2.2 卡片计算

后端从目录记录解析 ETRTO 的公称胎宽。对当前页每条记录，返回所有满足下列条件的规则：

```text
规则胎宽最小值 <= 目录 ETRTO 公称胎宽 <= 规则胎宽最大值
```

规则命中后，API 将范围放入该条记录的 `rim_width_guidance`。如果 ETRTO 缺失、格式异常或没有规则覆盖，记录仍然可以出现在目录中，但不返回该字段。

车圈内宽范围不会根据轮径、BSD、型号名称或销售 Product 状态重新推导；轮径 + BSD 仍然负责尺寸筛选，规则表只负责胎宽与内宽的参考关系。

### 2.3 不能从规则表推导的结论

以下内容需要逐型号事实、轮圈制造商要求、胎压规则或车架测量，不能由当前矩阵生成：

- 具体型号的完整轮圈兼容认证；
- Hookless、TLE/TLR 或胎圈系统认证；
- 车架、前叉和挡泥板间隙；
- 最大安全胎压；
- “黄金搭配”、最佳截面或跨型号性能排名。

## 3. 用户流程

### 3.1 筛选弹窗

页面结果区上方显示轮径 + BSD 导航，继续用复合值区分 `28-622`、`29-622`、`26-559` 和 `26-590` 等尺寸。筛选弹窗只保留胎体方向、胎圈结构、密封结构和 E-BIKE 等次级条件；车圈内宽数字输入、清除按钮、单一轮径校验和相关冲突提示均已移除。

这样用户可以：

1. 在页面导航中选择轮径 + BSD；
2. 使用弹窗中的胎体方向、胎圈结构、密封结构、E-BIKE 等仍保留的目录条件缩小结果；
3. 打开卡片查看该外胎的车圈内宽参考范围。

旧链接可能仍含 `inner_rim_width_mm`。选择器会在路由状态解析时将其视为无效的隐藏条件，不发送给新的页面请求；用户下一次提交弹窗或修改筛选时，路由合并逻辑会删除该参数。后端继续读取该参数，以兼容尚未升级的旧客户端。

### 3.2 结果表达

页面摘要在当前页确实有参考数据时显示统一说明：

```text
车圈内宽范围按外胎 ETRTO 公称宽度与官方可能组合参考表提供；实际安装请核对车圈、胎压和车架间隙。
```

卡片显示：

- `车圈内宽参考` 或 `Rim inner-width reference`；
- 一个或多个 `min–max mm` 范围；
- 现有 ETRTO、英制尺寸、重量、结构和商品存在状态。

不显示用户输入的内宽、不显示“命中候选数量”，也不把没有规则范围的记录标成不兼容。

## 4. API 契约

### 4.1 Selector 请求

`GET /products/schwalbe-tire-catalog/selector` 的页面请求增加：

```text
include_rim_width_guidance=1
```

这是“返回卡片参考数据”的显式开关，不是筛选条件。开启后，服务端读取迁移 362 的规则，并在分页后的每条记录上附加 `rim_width_guidance`。

`inner_rim_width_mm` 仍保留在 API 查询契约中，旧客户端继续可以使用原来的服务端筛选语义：它与轮径 + BSD 组合后参与结果过滤，并返回活动匹配上下文。Nuxt 选型页不再发送该参数。

### 4.2 Selector 响应

卡片记录的公开字段为：

```json
{
  "article_no": "…",
  "etrto": "40-622",
  "rim_width_guidance": [
    {
      "tire_width_min_mm": 35,
      "tire_width_max_mm": 46,
      "inner_rim_width_min_mm": 17,
      "inner_rim_width_max_mm": 27
    }
  ]
}
```

`source_url` 不进入 storefront 响应。来源版本和核验日期仍由规则只读接口提供；卡片使用已审定的参考文案，不在前端复制规则表。

`rim_width_guidance` 只包含当前页记录的参考范围。它不改变 `total`、`total_pages`、排序或分页；开启 `include_rim_width_guidance` 不会让结果变少。

### 4.3 服务端顺序

选型服务保持以下顺序：

1. 读取文本搜索命中集；
2. 派生严格 ETRTO 胎宽和轮径 + BSD；
3. 应用 URL 中仍有效的目录条件；
4. 仅在旧客户端传入 `inner_rim_width_mm` 时应用内宽过滤；
5. 按重量排序；
6. 计算总数并分页；
7. 根据 `include_rim_width_guidance` 为当前页记录附加参考范围。

参考范围计算位于 service 的纯函数中，handler 只解析查询参数，repository 只读取规则和来源元数据。

## 5. 文件职责审计

| 文件或模块 | 车圈内宽参考职责 | 边界 |
| --- | --- | --- |
| `go-backend/migrations/362_add_schwalbe_tire_rim_width_combination_rules.up.sql` | 保存并维护官方可能组合矩阵 | 不存具体 Product 兼容认证 |
| `go-backend/internal/repository/schwalbe_tire_rim_width_combination_rules_repository.go` | 读取规则和来源元数据 | 不解析 URL、不排序分页 |
| `go-backend/internal/service/schwalbe_tire_rim_width_matching.go` | 按 ETRTO 胎宽生成卡片参考范围；保留旧输入匹配函数 | 不生成 Hookless、胎压或车架结论 |
| `go-backend/internal/service/schwalbe_tire_catalog_selector.go` | 读取规则、筛选旧输入、为当前页建立 `RimWidthGuidanceByArticle` | 不把规则写入目录字段 |
| `go-backend/internal/api/v1/product/handler.go` | 解析 `include_rim_width_guidance` 和旧 `inner_rim_width_mm` | 不复制匹配算法 |
| `go-backend/internal/api/v1/product/schwalbe_tire_catalog_response.go` | 将内部 guidance map 投影为条目字段 | 不暴露 `source_url` |
| `nuxt-i18n/app/data/tireguides/schwalbeCatalog.ts` | 发送 include 开关并解析条目范围 | 不硬编码矩阵 |
| `nuxt-i18n/app/data/tireguides/schwalbeTireCatalogFilterQuery.ts` | 保留共享旧 URL/API 类型 | 不让新页面产生内宽筛选状态 |
| `nuxt-i18n/app/composables/useSchwalbeTireSelector.ts` | 忽略旧 URL 内宽条件、清理旧参数、管理轮径等页面筛选 | 不实现规则计算 |
| `SchwalbeTireCatalogFilterPanel.vue` | 展示 Radial、胎圈、密封结构和 E-BIKE 次级筛选 | 不再展示轮径 + BSD 或车圈内宽输入 |
| `SchwalbeTireSelector.vue` | 在摘要显示参考说明并承载卡片 | 不判断兼容性 |
| `SchwalbeTireCard.vue` | 显示 `rim_width_guidance` 范围 | 不把范围改写成认证或承诺 |
| `app/i18n/page-messages/guidesSchwalbeTireSelector/*.json` | 提供参考标签和边界提示 | 不把内部字段名当成用户必须理解的术语 |

## 6. 维护规则

1. 官方矩阵有变化时，只更新迁移/数据同步层及其来源元数据；不要在 Vue 或 TypeScript 中复制数字区间。
2. 新增或修改规则必须保留胎宽端点、内宽端点、来源依据、版本、URL 和核验日期，并更新迁移契约测试。
3. 如果未来引入其他品牌或其他规则体系，应增加独立的数据源和 service 纯函数，不把不同来源范围合并成一个“兼容范围”。
4. 规则表未迁移或读取失败时，selector 请求应报错并进入页面错误状态；部署必须先执行迁移 362，不能静默使用前端常量。
5. `inner_rim_width_mm` 的旧 API 筛选可以继续维护，直到所有外部客户端完成迁移；它不应重新出现在当前页面的筛选弹窗中。

## 7. 边界状态

- 当前页没有任何记录：不显示参考说明；保持既有空结果状态。
- 记录缺少可解析 ETRTO：记录可以正常展示，但不显示内宽参考。
- ETRTO 胎宽不在规则表：记录可以正常展示，但不显示内宽参考。
- 规则表为空：不会在前端伪造范围；部署和监控应发现规则数据缺失。
- 旧 URL 带内宽：页面忽略该隐藏筛选；提交其它筛选后由 URL 合并逻辑删除它。
- 轮径 + BSD 多选：只影响尺寸筛选，不改变卡片的胎宽参考范围。

## 8. 测试和验收

1. service 纯函数：端点包含、区间外排除、缺少 ETRTO 排除、结果排序和重复规则去重。
2. 后端 selector：`include_rim_width_guidance=1` 返回参考范围但不减少结果；旧 `inner_rim_width_mm` 测试继续验证兼容过滤和上下文。
3. API projection：范围进入条目字段，`source_url` 不进入公开响应。
4. 前端 URL：旧内宽参数不会限制新页面结果；提交弹窗后参数被移除；轮径 + BSD 继续按复合键工作。
5. 前端显示：卡片显示 `车圈内宽参考` 和范围，弹窗不再出现数字输入或清除内宽按钮。
6. 移动端：弹窗保持可滚动，页面级轮径导航横向滚动且不产生页面溢出，次级筛选不产生横向溢出。

## 9. 暂不处理的扩展

- 逐型号 Hookless/TLE/TLR 认证；
- 轮圈制造商型号白名单；
- 车架和前叉间隙计算；
- 依据具体型号官方数据生成最佳内宽或适配分数；
- 从内宽自动推导胎压上限；
- 将内宽参考写入 Schwalbe 销售商品模板字段。

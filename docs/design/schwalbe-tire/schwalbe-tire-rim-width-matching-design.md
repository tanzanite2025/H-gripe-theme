# Schwalbe 有钩 / 无钩与车圈内宽参考设计

> **状态：已实施。**
>
> Schwalbe 选型页不把 ETRTO/Schwalbe 的宽泛组合矩阵放进产品卡片，也不把车圈内宽作为选型页接口筛选条件。卡片先表达车圈系统兼容层，再给出 DT Swiss 的具体内宽参考。

## 1. 用户看到的结果

每张卡片始终有一层：

```text
有钩车圈    23–25 mm
```

只有独立数据表明确记录该型号支持 Hookless 时，才增加第二层：

```text
无钩车圈    26–27 mm
```

两层中的数字来自 DT Swiss TSS/TC 推荐表，按目录 ETRTO 公称胎宽读取。数字是安装前的推荐参考，不是 Schwalbe 对某个轮圈型号的认证；图表范围内的空缺宽度按相邻 DT Swiss 数据点计算，图表范围外才显示暂无数据。

## 2. 数据边界

### 2.1 车圈系统事实

迁移 369 建立 `schwalbe_tire_hookless_compatibility`。记录可以按 `article_no` 精确到单条 Article No.，也可以按完整 `model_name` 作为模型回退。解析顺序是：

1. Article No. 记录；
2. 模型记录；
3. 没有记录则 `unknown`。

只有 `status = supported` 才生成无钩层。`not_supported`、`unknown` 和没有记录都不生成无钩层；有钩层不依赖该表，始终存在。

不能从 `TLE/TLR`、`seal`、胎体结构、山地分类、Radial 或型号名称推断无钩兼容。官方 MTB 筛选同时出现在 Yes 和 No 的型号（例如 Jumbo Jim、Rocket Ron、Wicked Will）保持 unknown，避免把型号家族推断成错误认证。

### 2.2 车圈内宽参考

卡片使用 Go 后端 `tirerim.ResolveTireRimWidthReference`：

- 精确命中 DT Swiss recommended 行时直接显示该行；
- 图表内未列出的整数宽度按相邻数据点逐段线性计算，不因为区间跨度大就拒绝；
- 只有 DT Swiss 的 recommended 行作为推荐结果；某个精确宽度只有 possible 行时标记为 `possible_reference`，不提升为推荐；
- 图表覆盖范围外才返回暂无公开数据；不使用 Schwalbe/ETRTO 的宽泛“可能组合”矩阵；
- Hookless 参考只在 Schwalbe 独立事实层支持时计算和显示。

## 3. API 契约

`GET /api/v1/products/schwalbe-tire-catalog/selector` 的条目包含：

```json
{
  "article_no": "…",
  "etrto": "50-622",
  "wheel_size": {
    "value": "29-622",
    "wheel_diameter_in": "29",
    "bsd_mm": 622
  },
  "rim_compatibility": [
    { "rim_system": "hooked", "status": "supported" },
    { "rim_system": "hookless", "status": "supported" }
  ],
  "rim_width_guidance": [
    {
      "rim_system": "hooked",
      "result_kind": "exact_recommended",
      "rim_width_ranges": [{ "min": 23, "max": 25 }],
      "model_version": "dt-swiss-tss-tc-v1"
    }
  ]
}
```

`rim_compatibility` 只投影 `rim_system` 和 `status`；`rim_width_guidance` 是 Go 后端 DT Swiss 参考引擎的结果，包含结果类型、来源行、插值信息、模型版本和限制说明。来源 URL、Hookless 事实版本和核验日期只留在维护表中，不进入 storefront 卡片。

选型页请求不携带车圈内宽参数，也不再调用旧组合规则接口。

## 4. 文件职责

| 文件或模块 | 职责 |
| --- | --- |
| `go-backend/migrations/369_add_schwalbe_tire_hookless_compatibility.up.sql` | 建立可追溯 Hookless 事实表并种植官方 Yes/No 快照 |
| `go-backend/internal/repository/schwalbe_tire_hookless_compatibility_repository.go` | 读取 article/model 记录；表未迁移时返回空事实集 |
| `go-backend/internal/service/schwalbe_tire_hookless_compatibility.go` | Article 优先、模型回退和 unknown 保守解析；始终生成 hooked 层 |
| `go-backend/internal/service/schwalbe_tire_catalog_selector.go` | 为当前页建立 `RimCompatibilityByArticle` |
| `go-backend/internal/service/schwalbe_tire_catalog_rim_width_reference.go` | 按 ETRTO 公称胎宽调用 Go 后端 DT Swiss 参考引擎 |
| `go-backend/internal/api/v1/product/schwalbe_tire_catalog_response.go` | 将兼容层投影为 `rim_compatibility` |
| `nuxt-i18n/app/data/tireguides/schwalbeCatalog.ts` | 严格解析 `hooked` / `hookless` 和 `supported` |
| `nuxt-i18n/app/components/tireguides/schwalbe/SchwalbeTireCard.vue` | 展示后端返回的两层兼容性和 DT Swiss 参考结果 |
| `go-backend/internal/domain/tirerim/tire_rim_width_reference_engine.go` | DT Swiss TSS/TC 矩阵、精确行、possible 标记和分段插值的唯一计算源 |
| `go-backend/internal/api/v1/tirerim/tire_rim_width_reference_handler.go` | 提供 `/api/v1/engineering/tire-rim/matrix` 和 `/solve` |
| `nuxt-i18n/app/composables/useTireRimWidthReferenceRecommendation.ts` | 发送输入并展示后端结果，不复制矩阵或公式 |

## 5. 维护规则

1. 官方 Hookless 筛选变化时新增修订迁移或同步任务，保留来源 URL、筛选值、版本和核验日期；不要在 Vue 中写型号名单。
2. 新增 Article 级记录时，它优先于同型号回退记录；`not_supported` 和 `unknown` 必须保留，不能删除后让系统误显示无钩层。
3. 不得因为一个系列有 TLE/TLR 就批量标记 Hookless；必须有独立官方依据。
4. 更新 DT Swiss 图表时同步更新图表测试和来源说明；不得把 Schwalbe/ETRTO 的宽泛组合矩阵复制到卡片或常量中。
5. 新增或修订结果类型时同步更新 Go domain 测试、API 测试和前端类型；不得在 Vue 中恢复 DT Swiss 数值或线性公式。
6. 具体轮圈制造商的型号、胎压上限、无钩限制和车架间隙仍需用户核对对应厂商资料。

## 6. 验收

- 没有迁移 369 或没有型号事实：卡片只有有钩层；不报错、不猜无钩。
- 有明确 supported 的模型：卡片有有钩层和无钩层。
- Article 级 not_supported 覆盖模型级 supported：只显示有钩层。
- Jumbo Jim、Rocket Ron、Wicked Will 等 Yes/No 冲突家族：不显示无钩层。
- 选型响应不含已退休的宽泛车圈内宽规则字段。
- 页面网络请求不包含已退休的车圈内宽输入。
- `31 mm Hookless` 等图表断点之间的整数宽度返回 `interpolated`，不因跨度超过任意阈值而拒绝。
- `30 mm Hookless` 等只有 possible 行的精确宽度返回 `possible_reference`，不伪装成推荐。
- 没有 DT Swiss 图表覆盖时不显示伪造的内宽范围。

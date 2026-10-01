# Schwalbe 轮径页面导航设计

> **状态：已实施。** 本文记录已落地的页面级轮径 + BSD 导航和卡片尺寸投影。轮径 + BSD 已从筛选弹窗移到页面结果区上方；筛选弹窗只保留次级条件。
>
> **页面：** `/guides/tireguides/schwalbe-tire-selector`
>
> **页面交互与当前实现：** [Phase 2 实施指南](./phase2-standalone-page-implementation-guide.md)
> **字段来源：** [Schwalbe 字段矩阵](./schwalbe-master-catalog-specification-matrix.md)

## 1. 目标

轮径 + BSD 是用户进入轮胎目录时最先会辨认的尺寸入口。把它留在筛选弹窗里，用户必须先打开弹窗、展开尺寸项才能看到可选轮径。页面应直接展示目录中实际存在的尺寸选项，让用户一眼找到想查的规格；弹窗则只保留胎体方向、胎圈结构、密封结构和 E-BIKE 等次级筛选。

每张卡片也要直接显示轮径和 BSD，用户从尺寸入口进入结果后仍能核对单条型号，避免只看 ETRTO 或英寸标记产生误判。

## 2. 页面交互

### 2.1 轮径导航

在搜索/型号控件之后、结果摘要和卡片列表之前，放置一组页面级尺寸导航：

```text
全部  20″ · BSD 406  26″ · BSD 559  26″ · BSD 590  27.5″ · BSD 584  28″ · BSD 622  29″ · BSD 622 …
```

- 首项为“全部”，表示当前不按轮径 + BSD 限制结果。
- 其余项按目录中实际存在的轮径 + BSD 复合尺寸生成；不维护前端尺寸常量，也不从 BSD 单独猜轮径。
- `28″ · BSD 622` 与 `29″ · BSD 622` 必须是两个入口；`26″ · BSD 559` 与 `26″ · BSD 590` 也必须分开。
- 这些控件在视觉上使用 Tab 样式，但它们切换的是 URL 筛选状态和服务端结果，不是同一页面内的 `tabpanel` 内容。因此优先使用按钮/链接组并提供清楚的选中状态，不为外观强行套用不匹配的 ARIA `tablist`/`tabpanel` 语义。
- 桌面端允许按钮自然换行，常规宽度下形成两行；空间更窄时继续完整换行，避免尾部按钮被截断。移动端使用原生下拉框，只占一行并完整显示当前轮径 + BSD；下拉选项来自同一份动态尺寸数据。
- 移动端默认显示“全部轮径”；有已选尺寸时，下拉框显示当前完整尺寸。旧链接带多个 `wheel_size` 时显示“已选择多个轮径”，用户打开下拉后可切换为单个尺寸或全部。
- 当前“型号”下拉仍表示产品型号维度，与轮径导航不同；除非另行设计，不因本方案删除或改造型号选择。

### 2.2 与其它筛选条件的关系

- 点选一个尺寸入口时，页面通过现有 `wheel_size` 查询状态请求结果；查询值仍使用完整复合键，例如 `28-622`。
- “全部”只清除 `wheel_size`，保留搜索词、型号、重量排序及其它弹窗条件。
- 其它弹窗条件与轮径尺寸按现有跨维度 AND 规则组合；弹窗提交仍重置页码。页面尺寸入口的变化也重置到第一页。
- 文本搜索提交后保留已选尺寸。可用尺寸入口来自当前搜索命中集；如果直链或旧状态中的已选尺寸不在当前命中集的尺寸列表里，仍须把该已选值以可辨认的标签显示出来，用户可以切回“全部”解除它，不能留下看不见的隐藏条件。
- 当前 `wheel_size` 可接受多个重复值。新入口推荐一次突出一个尺寸；已有多值 URL 必须继续可读，结果仍按这些值匹配，并将相应入口显示为已选。用户点选某一个尺寸后收敛为该尺寸；点“全部”清除所有尺寸值。后端对旧 API 客户端的多值兼容保持不变。
- 这些尺寸项是筛选入口，不是产品型号页面，也不改变目录候选与在售 Product 的数据边界。

## 3. 数据与 API 链路

### 3.1 复用现有派生规则

后端 selector 已根据严格 ETRTO 和 `inch_designation` 派生：

- 英寸轮径：取官方 `inch_designation` 的首个英寸尺寸数值；
- BSD：只从严格 `ETRTO` 的第二个数值解析；
- 稳定筛选键：`轮径-BSD`，例如 `28-622`。

页面尺寸控件继续读取 selector 响应的 `filter_options.wheel_sizes`（`value`、`wheel_diameter_in`、`bsd_mm`），桌面按钮和移动下拉都由该数据生成。选项目前基于当前文本搜索命中集，搜索改变时随响应更新；不得把同一尺寸表另抄一份到 Nuxt 页面或翻译文件。

### 3.2 卡片尺寸投影

卡片需要显示一份由同一后端派生规则产生的尺寸投影，例如：

```json
{
  "wheel_size": {
    "value": "29-622",
    "wheel_diameter_in": "29",
    "bsd_mm": 622
  }
}
```

- 将尺寸投影作为 selector 条目响应的派生展示数据，复用 `deriveSchwalbeTireCatalogSelectorDimensions` 的结果；不新增数据库事实列或迁移。
- 同一 Article No. 的筛选键、Tab 标签和卡片展示必须来自相同尺寸派生结果，避免前端再次解析官方字符串造成边界不一致。
- 若英寸或 ETRTO 数据无法严格解析，不伪造完整轮径/BSD 组合。候选仍可展示，并继续显示目录原有的 ETRTO/英制字段；该记录不产生尺寸 Tab 项，也不附加不完整的 `wheel_size` 展示对象。
- `wheel_size` 只描述目录尺寸，不代表车圈兼容、轮胎间隙或安全认证。

## 4. 卡片表现

把轮径 + BSD 作为卡片主尺寸信息之一，例如：

```text
29″ · BSD 622 mm · ETRTO 57-622
```

单位和标签使用页面 i18n；Article No.、型号名和重量等其它卡片事实维持既有职责。若卡片已通过尺寸投影显示 ETRTO，避免重复显示造成拥挤。移动端应保持一行可读；空间不足时可自然换行，不截断 BSD。

## 5. URL 与状态恢复

- 继续使用当前 URL 契约 `wheel_size=轮径-BSD`，不新增平行参数。
- 直接打开、刷新、前进/后退和分页均恢复同一尺寸状态；翻页链接保留当前尺寸。
- 选择尺寸入口更新 URL、回到第一页并获取对应结果；不需要打开筛选弹窗。
- 进入弹窗后，尺寸入口已是页面上的主要尺寸控制，弹窗移除“轮径（BSD）”手风琴以及对应的草稿状态、尺寸复选控件和重复的 active-filter 计数；弹窗清除操作仍清除其它筛选条件。
- “全部”仅清除尺寸条件，不能顺带清空其它筛选。
- 路由中由旧界面或手工输入带来的多个 `wheel_size` 值继续按当前 OR 逻辑返回结果；页面上对应尺寸应明确可见为选中项，避免用户误以为当前是全部尺寸。

## 6. 文件职责与实施边界

| 文件/模块 | 计划职责 |
| --- | --- |
| `go-backend/internal/service/schwalbe_tire_catalog_selector.go` | 为当前页记录提供与 facet 相同来源的 `wheel_size` 派生投影；继续生成搜索命中集 `filter_options.wheel_sizes` |
| `go-backend/internal/api/v1/product/schwalbe_tire_catalog_response.go` | 将尺寸投影加入 selector 条目公开响应，不暴露来源 URL |
| `nuxt-i18n/app/data/tireguides/schwalbeCatalog.ts` | 校验和读取 `wheel_size` 投影及尺寸导航选项 |
| `nuxt-i18n/app/components/tireguides/schwalbe/SchwalbeTireSelector.vue` | 展示页面尺寸导航、同步 URL 选择、承载卡片 |
| `nuxt-i18n/app/components/tireguides/schwalbe/SchwalbeTireCatalogFilterPanel.vue` | 移除轮径复选手风琴和相应受控属性；继续管理剩余弹窗字段 |
| `nuxt-i18n/app/composables/useSchwalbeTireSelector.ts` | 继续以 `wheel_size` 作为唯一 URL/API 状态来源；提供尺寸入口选择所需动作 |
| `nuxt-i18n/app/components/tireguides/schwalbe/SchwalbeTireCard.vue` | 用服务端派生投影明确展示轮径、BSD 与 ETRTO |
| `nuxt-i18n/app/i18n/page-messages/guidesSchwalbeTireSelector/*.json` | 提供“全部”、BSD 标签及单位文案 |
| `nuxt-i18n/tests/schwalbe-selector-url-state.spec.ts` 与 Go selector/API 测试 | 验证真实 Tab/卡片数据链路、URL 状态和边界尺寸 |

本设计不改变迁移 362 的车圈内宽参考；卡片上现有车圈内宽参考应独立保留。也不改搜索语义、每页数量、排序、目录源、FAQ、SEO canonical/indexing 策略或商品发布流程。

## 7. 验收要求

1. 页面在筛选弹窗外直接显示全部可用轮径 + BSD 入口；弹窗不再出现轮径手风琴。
2. 页面尺寸标签由服务端搜索命中集的 `wheel_sizes` 动态生成，没有前端硬编码规格清单。
3. `28-622` 与 `29-622`、`26-559` 与 `26-590` 都能分别筛选，URL 中不发生交叉匹配。
4. 选择“全部”只删除 `wheel_size`，保留其它已提交条件；每次尺寸变化回到第一页。
5. 搜索、筛选弹窗、分页、刷新和浏览器历史导航保留或恢复同一尺寸状态；旧多值 URL 仍能理解且不会隐藏约束。
6. 卡片从 selector API 的同一派生投影显示 `轮径 · BSD · ETRTO`，不在前端从原始字段重复解析尺寸。
7. 缺失或异常尺寸不会被伪造成 Tab 或 BSD 值，相关候选仍正常显示。
8. 桌面尺寸按钮允许两行显示且不出现截断；390px 移动端使用一行原生下拉、无横向页面溢出，当前选项可辨认；卡片不截断 BSD。
9. 筛选弹窗现有剩余条件可以独立清除和提交；删除尺寸复选项后没有残留隐藏的尺寸草稿状态。
10. 相关 Go 测试、`vue-tsc`、selector URL 状态 E2E 均通过。

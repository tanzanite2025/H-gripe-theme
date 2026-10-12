# 不锈钢辐条本体轴向伸长页面实施规范

> 文档状态：当前生产实现说明（2026-10-03）  
> 正式路由：`/tools/spoke-stress-relief-mechanics`  
> 早期 HTML 白皮书只是演示存档，不是正式页面、接口契约或算法来源。

## 1. 产品范围

页面回答一个明确问题：**指定材料、截面积、有效受力长度和辐条张力下，单根辐条本体拉长多少？**

页面不计算整轮张力、车圈变形、辐条角度、孔座凹陷、接触沉降、轮组总张力衰减或工厂测试载荷。应力释放和日常负载只能作为说明和计算余量背景，不能被伪装成单根辐条轴向结果。

页面中的两个数据层必须分开：

1. **实时计算器**：使用用户输入的实际有效长度和张力，由后端即时计算。
2. **型号参考表**：统一使用 270 mm、180 kgf（约 1765 N）计算余量基准，用于给每个截面一个可直接比较的伸长结果。

180 kgf 的解释必须写在页面上：DT Swiss 官网部分轮组的辐条张力约 110 kgf，这是对应型号的官方参考；140 kgf 只是常见工作张力的通用比较值，不得覆盖具体厂商和型号规格。180 kgf 是给应力释放、日常负载和单根轴向模型未覆盖因素预留余量，不是轮组编制目标。

## 2. 页面与后端文件职责

```text
nuxt-i18n/app/pages/tools/
└── spoke-stress-relief-mechanics.vue
    页面结构、图示、输入控件和结果呈现

nuxt-i18n/app/composables/
└── useStainlessSteelSpokeDislocationMechanicsCalculation.ts
    metadata/calculate 请求、契约校验、SSR 结果和请求状态

nuxt-i18n/app/i18n/page-messages/guidesSpokeDislocationMechanics/
├── zh_cn.json
└── en.json
    页面文案、范围说明和错误状态

go-backend/internal/domain/spokedislocationmechanics/
├── stainless_steel_spoke_dislocation_mechanics_engine.go
├── stainless_steel_spoke_material_reference_catalog.go
├── stainless_steel_spoke_material_reference_catalog.json
├── stainless_steel_spoke_model_reference_catalog.go
└── stainless_steel_spoke_model_reference_catalog.json
    材料曲线、型号截面、公式、范围、舍入和安全分类

go-backend/internal/api/v1/spokedislocationmechanics/
├── stainless_steel_spoke_dislocation_mechanics_http_handler.go
└── stainless_steel_spoke_dislocation_mechanics_http_handler_test.go
    HTTP 请求解析、错误映射和接口契约测试
```

前端不能复制后端公式、材料数字或安全门槛，也不能在请求失败时回退到本地估算。后端暂时不可用时必须显示“计算服务不可用”；输入不合法时必须显示“输入无效”。

## 3. HTTP 契约

生产接口为：

```text
GET  /api/v1/engineering/spoke-dislocation-mechanics/metadata
POST /api/v1/engineering/spoke-dislocation-mechanics/calculate
```

请求字段：

```json
{
  "model_id": "cx-ray",
  "custom_effective_area_mm2": null,
  "effective_spoke_length_mm": 290,
  "nominal_working_tension_n": 1200,
  "overload_ratio_percent": 135
}
```

目录型号自动解析有效截面积；只有 `model_id=custom` 时才接受自定义截面积。选中目录型号后，页面的有效承力截面积必须锁定为后端值。

后端输入范围：

| 字段 | 范围 |
| --- | --- |
| `nominal_working_tension_n` | `0 < F ≤ 3000 N` |
| `custom_effective_area_mm2` | `0.2 < A ≤ 10 mm²` |
| `effective_spoke_length_mm` | `1 < L ≤ 1000 mm` |
| `overload_ratio_percent` | `100% ≤ η ≤ 200%` |

所有非有限值、未知型号、缺少自定义截面积和越界请求都必须返回结构化错误。工作或过载应力高于材料曲线 1346 MPa 时返回曲线越界错误，不外推断裂力。

## 4. 材料和算法契约

正式材料档案是 AISI 304 / X5CrNi18-10 冷拔线材参考状态：

| 字段 | 值 |
| --- | ---: |
| `elastic_modulus_mpa` | 193000 |
| `yield_strength_mpa` | 1137 |
| `ultimate_tensile_strength_mpa` | 1346 |
| `total_elongation_to_failure_percent` | 5 |

计算顺序：

```text
σ = F / A
εtotal = interpolate(stored engineering stress–strain curve, σ)
εelastic = σ / E
εplastic = max(0, εtotal − εelastic)
ΔLtotal = εtotal × L
ΔLelastic = εelastic × L
ΔLpermanent = εplastic × L
```

曲线屈服后的上升段已经包含该冷拔状态的冷作硬化。实现不得重新引入 950 MPa 独立位错门槛、135%“黄金”门槛、位错全部钉死、绝对零蠕变或人为硬化系数。

## 5. 型号参考表契约

`metadata.reference_calculation` 固定返回：

```text
effective_spoke_length_mm = 270
reference_calculation_tension_kgf = 180
reference_calculation_tension_n = 1765.197
```

每个型号使用自己的有效承力截面积计算 `σ=F/A`，再从同一材料曲线得到总应变。该表的参考伸长是：

| 型号 | 面积 mm² | 伸长 mm |
| --- | ---: | ---: |
| Sapim CX-Ray | 1.56 | 1.583 |
| Pillar Wing 20 | 1.62 | 1.524 |
| Sapim Laser / DT Swiss Aerolite / DT Swiss Revolution | 1.77 | 1.395 |
| DT Swiss Revolite | 1.94 | 1.273 |
| Sapim D-Light | 2.14 | 1.154 |
| Sapim Race / DT Swiss Aero Comp / DT Swiss Competition | 2.54 | 0.972 |
| Sapim Leader / DT Swiss Champion | 3.14 | 0.786 |

同一表格还可以按 `F = σ × A` 显示曲线屈服边界力和曲线终点边界力。边界力不是轮组官方张力，也不是车圈认证极限。

## 6. 断裂和车圈说明

材料档案的 5% 只用于按当前 `L` 换算拉伸试验总断裂伸长参考。公开资料没有颈缩后的真实应力—应变、断面收缩率和断裂位置，所以接口不得输出完整真实断裂力，也不得拿普通 304、其他冷拔状态或品牌数据补齐。

碳圈孔座 1850 N 是页面自设比较基准；1650 N 是页面内高应力提示基准。两者都不是工厂测试标准、车圈认证值或普适安全极限。页面说明实际测试方向和轮组辐条角度存在差异，本页面不求解角度因素。

## 7. 图示和材料背景

晶格图、材料学图和三张大卡片属于页面解释层，必须保留。图示只说明冷拔材料和曲线状态的背景，不得把示意图文字升级为可计算参数。特别是：

- “可动位错”只能作为材料学背景，不能变成独立输入；
- “曲线屈服后状态”表示使用曲线上升段，不表示所有位错已被钉死；
- 不能从图示推导骑行后张力衰减率、疲劳寿命或车圈承载能力。

## 8. 验收清单

- [x] 正式路由是 `/tools/spoke-stress-relief-mechanics`，旧 `/guides/spokeguides/stainless-steel-microstructural-dislocation-mechanics` 已退役为 404，不建立兼容页或重定向。
- [ ] 目录型号的有效截面积由后端锁定；自定义型号才允许编辑。
- [ ] 180 kgf / 270 mm 只出现在型号参考表，并标注为计算余量基准。
- [ ] 页面区分 DT Swiss 官网部分型号约 110 kgf 的厂商参考、约 140 kgf 的通用比较值，以及 180 kgf 的表格计算余量；180 kgf 不是编制目标。
- [ ] 计算结果包含总伸长、弹性伸长和理想卸载后的永久伸长。
- [ ] 材料曲线越界时后端拒绝结果；页面不外推断裂力。
- [ ] 后端不可用时不显示本地估算。
- [ ] 页面保留晶格图、材料学图和三张大卡片。
- [ ] 不出现 950 MPa 独立门槛、135% 黄金窗口、官方疲劳次数或未经证实的品牌断裂力。
- [ ] Go 测试、`vue-tsc --noEmit` 和 `npm run i18n:check` 通过。

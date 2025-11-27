# Spoke Calculator Design

> 辐条计算器的数据结构与架构说明，包含当前实现进度（WordPress 插件 + Nuxt 页面）和后续 TODO。

---

## 0. 当前实现进度（后端 + 前端）

- **后台插件：tanzanite-setting**
  - 已在 `tanz_product` 编辑页新增「Spoke Geometry / 轮组几何」meta box（类：`Tanzanite_Product_Geometry_Admin`），作为备用编辑入口。
  - 新增独立后台页面「Spoke Geometry / 轮组几何」子菜单（slug: `tanzanite-spoke-geometry`，类：`Tanzanite_Spoke_Geometry_Admin`），用于：
    - 按标题 / 分类（`rim` / `hub` / `nipple`）筛选 `tanz_product`；
    - 查看每个商品 Rim / Hub 几何是否完整；
    - 为单个商品补充 / 编辑 Rim + 前/后花鼓几何字段（使用与 REST `spoke-products` 相同的 meta key）。
  - Rim 几何字段已按本文件的 `RimGeometry` 落地为 meta key（见 1.3）。
  - Hub 采用「一条商品 = 一套前后花鼓」的成对方案（front + rear 几何字段各 4 个，见 1.2 / 1.3）。
- **REST API：WordPress**
  - 新增 `GET /wp-json/tanzanite/v1/spoke-products`（类：`Tanzanite_REST_Spoke_Products_Controller`）。
  - 返回结构：`{ rims: RimGeometryLite[], hubs: HubGeometryLite[], nipples: NippleOption[] }`，仅包含几何字段齐全的商品。
- **Nuxt 前端**
  - `/spokecalculator` 页面与 `useSpokeCalculator()` composable 已实现 UI 与计算 API 调用骨架。
  - 目前 `/api/spoke-products` 仍使用本地 mock 数据，下一步将改为调用 WordPress 的 `/tanzanite/v1/spoke-products`。

## 1. 商品几何字段 Schema 设计

整体思路：

- 在 WordPress 后台的 **tanzanite-setting 插件** 中，在「新建商品」页面统一增加这些几何字段输入框（对应下面的 RimGeometry / HubGeometry 字段以及 meta key），适用于所有商品类型：
  - 对于需要参与辐条计算的商品（轮圈 / 花鼓 / 辐条帽等），后台填写这些几何字段，前端 / API 即可按 meta 直接读取；
  - 对于与计算无关的商品，这些字段可以留空，API 在读取时会自动过滤掉这些商品，不参与计算。
- 新建或编辑商品时，由你在后台手动填写这些值；
- 前端 Nuxt / 后端 API 不再自己“猜测”，而是**直接按约定的 meta key 读取这些字段**，如果某个字段为空，就按以下策略处理：
  - 对于必须字段（比如 `erd`, `spokeHoles`, 法兰到中心的距离），如果缺失则直接提示“几何信息不完整”；
  - 对于可选字段（如内宽、外宽、辐条帽类型等），缺失时仅影响展示或过滤，不影响基础计算。

前端 UI 方面的规划：

- 将辐条计算页面中的 Rim / Hub 输入框从「纯文本」改为**下拉选择框**，选项来源于商店中指定分类（例如 `rim`、`hub`、`nipple`）下的商品列表；
- 在 Nipple Type、Rim、Hub 一组配置附近增加一个「加入购物车」按钮：
  - 计算完成后，可一键将当前选中的 Rim + Hub + Nipple 商品组合加入购物车；
  - 这个能力是增值功能，后续落实现实时价格、库存等逻辑时再细化；
- 为了方便统一管理和筛选，建议在商品分类中预先建立 `rim`、`hub`、`nipple` 等分类名称，供前端下拉组件按分类过滤商品候选项。

### 1.1 Rim（轮圈）几何字段

用 TypeScript 接口描述（方便前后端共用类型）：

```ts
export interface RimGeometry {
  id: string            // 唯一 ID，可以用 SKU、slug、数据库主键
  sku?: string
  brand?: string
  model?: string

  // 几何参数 —— 计算辐条长最关键
  erd: number           // Effective Rim Diameter，有效轮径，单位 mm
  spokeHoles: number    // 孔数（28 / 32 / 36 等）
  offsetMm?: number     // ERD 中心相对轮圈几何中心的偏移（有就填）

  // 选填：与计算无关，仅用于展示 / 过滤
  diameterLabel?: string      // 轮径显示用标签，例如 "700C"、"29"（可选）
  internalWidthMm?: number    // 轮圈内宽，单位 mm（可选）
  externalWidthMm?: number    // 轮圈外宽，单位 mm（可选）
  nippleSeatType?: string     // 辐条帽 / 孔位类型，例如 standard / eyelet / washer（可选）
  holeType?: 'eyelet' | 'non-eyelet' | string
  material?: 'alloy' | 'carbon' | string
}
```

**最关键字段（不可少）：**

- `id`
- `erd`
- `spokeHoles`

其余字段主要用于：

- 前端筛选 / 展示（直观显示轮径、材质等）；
- 后台维护时更好识别产品。

---

### 1.2 Hub（花鼓）几何字段

左右几何参数通常不同，因此需要拆开 left/right：

```ts
export interface HubGeometry {
  id: string          // 唯一 ID
  sku?: string
  brand?: string
  model?: string

  spokeHoles: number  // 孔数：24 / 28 / 32 / 36 等
  type?: 'front' | 'rear' | 'front-rear-compatible'
  brakeType?: 'disc' | 'rim' | 'centerlock' | string
  axleWidthMm?: number  // OLD，花鼓总宽，比如 100 / 142 / 148

  // 几何参数 —— 计算辐条长使用
  leftFlangePcdMm: number          // 左侧法兰直径（pitch circle diameter）
  rightFlangePcdMm: number         // 右侧法兰直径
  leftFlangeToCenterMm: number     // 轮轴中心到左法兰中心的距离
  rightFlangeToCenterMm: number    // 轮轴中心到右法兰中心的距离
}
```

如果以后需要更详细的信息（比如仅用于展示或者用于更精细的计算），可以在 `HubGeometry` 中再扩展：

- `flangeToFlangeMm?: number` —— 左右法兰之间的距离；
- `spokeHoleDiameterMm?: number` —— 辐条孔直径。

当前版本的计算不依赖这两个字段，所以暂时不在接口中强制要求。

**最关键字段（不可少）：**

- `spokeHoles`
- `leftFlangePcdMm` / `rightFlangePcdMm`
- `leftFlangeToCenterMm` / `rightFlangeToCenterMm`

这些基本可以支持常规的辐条长度计算公式（基于三角函数）。

---

### 1.3 在商品系统中的落地建议

自建系统，在工作区的 tanzanite-setting 插件中，这些几何字段已经作为 `tanz_product` 的 meta 存储。
当前实现：通过独立后台页面「Spoke Geometry / 轮组几何」为**已有商品**补充 / 编辑这些 meta（不改动现有「Add New Product」SPA）。
如果以后希望在新建商品时同步填写几何参数，可以在 `Add New Product` 页面中接入同一批 meta 字段（沿用下面的 key），保持与 REST / 计算逻辑兼容。

- **Rim 商品 meta 字段示例**：
  - `_tanz_erd`
  - `_tanz_spoke_holes`
  - `_tanz_diameter`            // 对应 `diameterLabel`，仅展示用标签（可选）
  - `_tanz_internal_width_mm`   // 轮圈内宽 mm（可选）
  - `_tanz_external_width_mm`   // 轮圈外宽 mm（可选）
  - `_tanz_nipple_seat_type`    // 辐条帽 / 孔位类型（可选）
  - `_tanz_material`

- **Hub 商品 meta 字段示例（成对前/后花鼓，方案 B）**：
  - `_tanz_spoke_holes_hub`          // 整套前/后花鼓的孔数
  - `_tanz_axle_width_mm`            // OLD 轴宽（整套）
  - `_tanz_front_left_flange_pcd_mm`
  - `_tanz_front_right_flange_pcd_mm`
  - `_tanz_front_left_flange_to_center_mm`
  - `_tanz_front_right_flange_to_center_mm`
  - `_tanz_rear_left_flange_pcd_mm`
  - `_tanz_rear_right_flange_pcd_mm`
  - `_tanz_rear_left_flange_to_center_mm`
  - `_tanz_rear_right_flange_to_center_mm`

服务端 / API 负责从这些 meta 中读取值，拼成 `RimGeometry` / `HubGeometry` 对象，再交给计算公式使用。当前实现中，Hub 在 REST 返回中使用 `position: 'front-rear-compatible'` 表示一条记录包含前/后两侧几何。

---

## 2. 计算架构：前端 composable + 后端 API

总体推荐结构：

- **服务端（Nuxt server API 或独立后端）**
  - 持有完整的几何数据（可以从 Excel 导入）；
  - 封装真实的辐条长度计算公式；
  - 提供一个 HTTP 接口，例如 `POST /api/spoke-calc`。

- **前端（Nuxt 页面）**
  - 只负责 UI 和表单；
  - 通过 `useSpokeCalculator()` composable 调用上面的 API；
  - 不直接携带 3000+ 行数据，也不暴露公式细节。

好处：

- 前端 JS bundle 体积小，首屏加载更快；
- 3000 行数据可以继续增长而不影响页面速度；
- 公式和数据藏在服务端，更易维护、也更难被直接复制；
- 将来可以给其它前端（小程序、App、后台工具）复用同一个 API。

---

## 3. Nuxt Server API 示例：`server/api/spoke-calc.post.ts`

> 这里只是结构示例，公式和数据源都是占位的，方便后续按实际 Excel 和商品系统改写。

### 3.1 请求 / 响应类型

```ts
// 请求体
interface SpokeCalcRequest {
  rimId: string
  hubId: string
  wheelPosition: 'front' | 'rear'
  spokeCount: number
  crossing: number
}

// 响应体
interface SpokeCalcResponse {
  leftLengthMm: number
  rightLengthMm: number
  debug?: {
    rim: any
    hub: any
    formulaVersion: string
  }
}
```

### 3.2 API 结构草案

关于计算结果的策略，可以采用：

- **有记录就直接取记录**：如果数据库 / JSON 里已经保存了某个组合（特定 rim + hub + crossing 等）的精确结果，直接返回，速度最快；
- **无记录则用公式计算**：当查不到现成记录时，退回到几何公式计算；如有需要，可以在计算完成后把结果写回数据库，作为下次的缓存。

```ts
// server/api/spoke-calc.post.ts
import { defineEventHandler, readBody } from 'h3'

export default defineEventHandler(async (event) => {
  const body = await readBody<SpokeCalcRequest>(event)

  // 1) 根据 ID 获取轮圈 / 花鼓几何参数
  const rim = await getRimGeometry(body.rimId)
  const hub = await getHubGeometry(body.hubId, body.wheelPosition)

  if (!rim || !hub) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Unknown rim or hub ID'
    })
  }

  // 2) 调用真实计算公式
  const { left, right } = computeSpokeLengths({
    rim,
    hub,
    crossing: body.crossing,
    spokeCount: body.spokeCount
  })

  return {
    leftLengthMm: left,
    rightLengthMm: right,
    debug: {
      rim,
      hub,
      formulaVersion: 'v0.1-prototype'
    }
  } satisfies SpokeCalcResponse
})
```

上面的几个辅助函数需要根据你的数据源实现：

- `getRimGeometry(id: string)`
- `getHubGeometry(id: string, position: 'front' | 'rear')`
- `computeSpokeLengths({ rim, hub, crossing, spokeCount })`

其中 `computeSpokeLengths` 应该使用你 Excel 中那套已经验证的公式。

### 3.3 从 Excel / 数据表读取几何数据的思路

你现在有一张约 **3000 行的 Excel 表**（后期也可以根据需要精简到约 500 行）。可以先采用下面这种典型落地方式：

1. **一次性导入数据库 / JSON 文件**
   - 建一张 `rims` 表、一张 `hubs` 表，字段对应上面的 schema；
   - 或者导成两个 JSON 文件，服务端在启动时加载进内存；
   - `getRimGeometry` / `getHubGeometry` 就是查数据库 / 查内存表。

> TODO：在本节下面，可以补充你最终选用的存储方案（MySQL / tanzanite-setting meta / JSON 文件等）。

---

## 4. 前端 composable：`useSpokeCalculator()`

前端页面不直接调用 `useFetch('/api/spoke-calc')`，而是通过一个 composable 统一封装请求逻辑和状态：

```ts
// composables/useSpokeCalculator.ts
interface SpokeCalcInput {
  rimId: string
  hubId: string
  wheelPosition: 'front' | 'rear'
  spokeCount: number
  crossing: number
}

interface SpokeCalcResult {
  leftLengthMm: number
  rightLengthMm: number
}

export function useSpokeCalculator() {
  const loading = ref(false)
  const error = ref<string | null>(null)
  const result = ref<SpokeCalcResult | null>(null)

  const calculate = async (input: SpokeCalcInput) => {
    loading.value = true
    error.value = null

    try {
      const { data, error: fetchError } = await useFetch<SpokeCalcResult>('/api/spoke-calc', {
        method: 'POST',
        body: input
      })

      if (fetchError.value) {
        throw fetchError.value
      }

      result.value = data.value ?? null
    } catch (e: any) {
      console.error('Spoke calc failed', e)
      error.value = e?.message || 'Failed to calculate spoke lengths'
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    result,
    calculate
  }
}
```

在页面（例如 `spoke-calculator.vue`）中的使用方式：

```ts
const { result, loading, error, calculate } = useSpokeCalculator()

const onCalculate = () => {
  calculate({
    rimId: selectedRimId.value,
    hubId: selectedHubId.value,
    wheelPosition: wheelPosition.value,
    spokeCount: spokeCount.value,
    crossing: crossing.value
  })
}
```

这样可以保证：

- 页面代码更干净，只关心“有哪些输入”和“拿到什么结果”；
- 将来如果 API 地址或鉴权方式改了，只需要改 composable 一处。

---

### 5.2 推荐方案：服务端 API + 前端 composable（已按此路线实现骨架）

目前仓库中已经存在：

- `server/api/spoke-calc.post.ts`：Nuxt server API 骨架；
- `app/composables/useSpokeCalculator.ts`：封装调用的 composable；
- `app/pages/spokecalculator.vue`：使用上述 composable 的页面 UI。

后续只需要将计算公式替换为 Excel 中的真实逻辑，并把 `/api/spoke-products` 切换为调用 WordPress 的 `/tanzanite/v1/spoke-products`，即可完成端到端串联。

## 6. 下一步：Nuxt 接入 WordPress spoke-products（TODO 清单）

- [ ] 在 Nuxt 项目中，让 `/api/spoke-products` 通过服务端请求 WordPress 的 `GET /tanzanite/v1/spoke-products`，替换当前的本地 mock 数据。
- [ ] 将 WordPress 返回的 rims / hubs / nipples 映射为前端使用的 `RimGeometryLite[]` / `HubGeometryLite[]` / `NippleOption[]` 结构。
- [ ] 在 `spokecalculator` 页面中，用真实的 rims / hubs / nipples 数据填充下拉框，并处理加载中 / 无数据等状态。
- [ ] 处理请求失败时的错误提示和重试逻辑，避免影响辐条计算表单的基本使用。
- [ ] 选取一组在 WordPress 后台「Spoke Geometry」页面中填好的 Rim + Hub + Nipple 商品，验证前端下拉可见且计算接口可以正常使用这些数据。

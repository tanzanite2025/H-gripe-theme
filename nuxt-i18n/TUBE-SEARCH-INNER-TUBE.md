# TUBE-SEARCH（Inner tube）实施方案（方案 B）

## 1. 背景与目标

- 页面位置：`/guides/tireguides` 的 **Inner tube** Tab。
- 业务目标：
  - 只会有大约 **20 多款**内胎 SKU。
  - 在后台为内胎维护**少量关键参数**，让「现有的高级搜索」能准确读取这些字段，从而在 `/shop` / 高级搜索弹窗里正确筛选出对应商品。
  - **不再增加**现有「Add new product」后台表单的复杂度，避免把 tanzanite-setting 插件继续膨胀。
  - 前端 Inner tube Tab **不做额外搜索逻辑**：
    - 只需要一段文案 + 一个按钮；
    - 按钮点击后**直接打开现有的高级搜索弹窗**（并预设好「内胎」相关筛选条件）。

因此，本方案 B 的核心是：

> **「新增一层‘内胎规格配置’（配置表 + 后台页面），前端复用现有高级搜索弹窗，不再额外造一个 tube search 流程。」**

---

## 2. 整体架构概览

### 2.1 架构要点

- **商品仍由现有 Add new product 页面创建**：
  - 只需保证内胎商品归类到特定分类（如 `tube` 或 `inner-tube`）。
- **新增一张「内胎规格配置表」作为中间层**：
  - 单独的 admin 页面：只给内胎用；
  - 存放「这个内胎适用于哪些尺寸 / 气嘴类型 / 系列」等信息。
- **同步机制**（必须保证）：
  - 在 REST API 层按商品 ID 关联 `tube_specs`，将关键规格字段合并进商品 JSON，让现有产品搜索只依赖一套统一字段进行过滤和展示。
- **前端**：
  - `/guides/tireguides` → Inner tube Tab 内：
    - 只展示介绍文字 +「查找内胎」按钮；
    - 点击按钮：调用已有的高级搜索弹窗（并传入「内胎分类」为默认筛选）。

### 2.2 与现有系统的关系

- **不修改**：
  - 现有 Add new product 大表单结构；
  - 现有高级搜索弹窗的基础逻辑（只是在其内部增加/识别一些 tube 相关字段/分类）。
- **新增**：
  - 后台：`Tanzanite Setting → Tube Specs`（名字示例）。
  - 数据：`tube_specs` 配置表（独立于商品主表）。
- **轻微扩展**：
  - 在 REST API 或 WP JSON 输出中，通过 join `tube_specs`，为内胎商品附带统一的 tube 规格字段。

---

## 3. 数据模型设计（Tube Specs 配置）

### 3.1 存储结构（概念）

可以采用单独一张 DB 表，例如：`wp_tanzanite_tube_specs`（名称仅示例）：

- `id`：自增 ID；
- `product_id`：关联内胎商品 ID（WP Product / 自定义 post ID）；
- `size_label`：人类可读的尺寸标签（如 `700x25-32C`，方便在后台/前台展示）；
- `etrto_range`：ETRTO 范围（如 `622x25-32`；也可以拆成 `etrto_min` / `etrto_max`）；
- `valve_type`：气嘴类型（枚举）：
  - `presta` / `schrader` / `dunlop` 等；
- `segment`：应用场景（可选）：
  - `road` / `gravel` / `mtb` / `city` / `kids` 等；
- `execution`：系列 / 版本（可选）：
  - 如 `standard` / `plus` / `light` / `extra_light`；
- `notes`：备注（可选）。

> **控制复杂度**：  
> 鉴于只需要大约 20 多款内胎，建议只保留：
> - `product_id`
> - `size_label` 或简单的 `etrto_range`
> - `valve_type`
> - （可选）一两个简单的自定义枚举字段（例如 `segment`），其他先不做。

### 3.1.1 系列（execution）与阀门规格建模

内胎系列（execution）可以直接**写死为固定枚举**，便于前后端统一：

- `STANDARD`
- `AIR_PLUS`
- `EXTRALIGHT`
- `XXLIGHT`
- `FREERIDE`
- `DOWNHILL`

实际实现时，推荐：

- 在 `tube_specs.execution` 字段中使用上述枚举值之一；
- 前端与文案层通过映射表将其转成展示文案（如 “Standard” / “Air Plus” 等）。

阀门类型与长度的组合同样是**有限集合**，可以作为可搜索字段固定下来：

- `AV`：仅 40 mm 直阀；
- `AV_45`：45° 角度阀；
- `DV`：32 mm、40 mm 两种长度；
- `SV`：40 mm、50 mm、60 mm、80 mm 四种长度。

结合上述约束，每一条 tube 规格行可以按以下方式建模：

- `valve_family`：`AV` / `DV` / `SV`（物理阀门类型，不包含角度信息）；
- `valve_angle_deg`：
  - 对 AV 可以是 `0` 或 `45`（直阀 / 45° 阀）；
  - 对 DV / SV 一律为 `0`；
- `valve_length_mm`：
  - 对 `AV` 固定为 40 mm 长度（无论 `valve_angle_deg` 是 0° 还是 45°，长度都写 40）；
  - 对 `DV` 只允许 32 或 40 mm 长度；
  - 对 `SV` 只允许 40 / 50 / 60 / 80 mm 长度；
- `execution`：上述 6 个系列之一。

搜索层面：

- `execution`、`valve_family`、`valve_angle_deg`、`valve_length_mm` 都可以作为 REST API 的过滤字段；
- 例如查询参数可以设计为：`tube_execution=STANDARD&tube_valve_family=SV&tube_valve_angle=0&tube_valve_length=60`；
- 在 `tube_specs` 中一行就代表「某个具体型号（product_id）在某个系列 + 某种阀门家族 + 某个角度 + 某个长度」的组合；
- 对于同一型号在同一系列下支持多种阀门/长度/角度的情况，可以用多行记录同一个 `product_id`，但 `valve_family` / `valve_angle_deg` / `valve_length_mm` 不同；
- 特别是 AV 45° 的记录，应写成：`valve_family=AV`、`valve_angle_deg=45`、`valve_length_mm=40`，从字段层面避免将 45° 误写成 45 mm 长度。

### 3.2 与高级搜索的字段对接

为了让「现有高级搜索」能够读取这些规格字段，采用**单一的统一方式**：

- **在 REST API 中按商品 ID 读取并关联 `tube_specs` 表**（左连接 join）：
  - 不改商品 meta 结构；
  - 在 API 层将 tube_specs 字段合并进商品 JSON（例如挂在 `tubeSpecs` 字段下）；
  - 高级搜索里的过滤逻辑只基于这一套字段进行过滤和展示。

这样可以保证：

- 同一商品的 tube 规格只有一份真源（tube_specs 表）；
- 前端和高级搜索只需要理解一套字段，不需要区分「meta vs specs」两种来源。

---

## 4. 后台 UI：Tube Specs 管理页

### 4.1 入口位置

- WordPress 后台菜单：
  - `Tanzanite Setting`
    - `General`
    - `Products`
    - ...
    - **`Tube Specs`**（新加）。

### 4.2 页面功能需求

- **列表视图**：
  - 每一行对应一条 inner tube 规格配置；
  - 列表字段：
    - 关联商品（商品名称 + 链接）；
    - `size_label`；
    - `valve_type`；
    - `segment`（如有）；
    - `execution`（如有）；
  - 提供简单筛选 / 搜索（按商品名或 size_label）。

- **新增 / 编辑表单**：
  - 选择商品：
    - 下拉框或自动完成，仅列出「分类 = 内胎」的商品；
  - 填写字段：
    - 尺寸 / ETRTO（用简单文本或少量预设选项即可）；
    - 气嘴类型（单选下拉：Presta / Schrader / Dunlop）；
    - 可选字段（如 segment / execution）。
  - 保存时：
    - 写入 `tube_specs` 表；
    - 不直接改动商品表单结构，仅通过 REST API join `tube_specs` 做统一字段输出。

- **删除 / 禁用**：
  - 可以将配置行 `删除` 或 `禁用`，前端搜索不再用此行。

> 由于只有 20 多个 SKU，这个页面的工作量很有限，可以完全手工维护，风险也可控。

---

## 5. 前端集成：Inner tube Tab

### 5.1 Inner tube Tab 的 UX 约定

- 页面：`/guides/tireguides` → Inner tube Tab；
- 不再做 Step 1/2/3 的复杂交互，改为：
  - 文本说明：  
    简要介绍：
    - 如何看轮胎上的尺寸标记；
    - 选择内胎时需要注意哪些点（尺寸 / 气嘴 / 使用场景）。
  - **一个主按钮**：
    - 文案示例：`Find inner tubes` / `Search inner tubes`；
    - 样式使用当前 guides CTA 统一按钮样式；
    - 点击后：

      ```ts
      // 示例逻辑（Nuxt 前端）
      openShopSearchSheet({
        presetCategory: 'tube',        // 内胎分类
        // 如需要，还可以带上少量预设 filter key
      })
      ```

  - 按钮只做一件事：**打开现有「商品高级搜索」弹窗，并预设为内胎维度**。

### 5.2 与高级搜索弹窗的衔接

- 在现有高级搜索弹窗（已经用在 `/shop` 和 SiteHeader）中：
  - 支持根据分类/属性过滤内胎商品：
    - 例如有一个 Category = Tube；
    - 或者在 Attribute 里有 Tube / Valve / Size 等字段；
  - 在 Inner tube Tab 的按钮点击时：
    - 传一个简单的「预设参数」（如 `pendingSearch`）告诉弹窗：
      - 默认选中内胎分类；
      - 也可以默认展开某个「Tube」属性区域。

- 高级搜索的实际查询逻辑：
  - 仍由现有 `/wp-json/tanzanite/v1/products` 接口负责；
  - 接口内部如果能读取 `tube_specs` 字段（见第 3 节），就可以根据这些字段提供更精确的过滤。

> 关键点：  
> Inner tube Tab **不负责搜索逻辑**，只负责 CTA。  
> 搜索逻辑集中在已有的「高级搜索弹窗 + REST API + tube_specs」。

---

## 6. 实施步骤（开发视角）

### Step 1：数据表与模型

- 在 tanzanite-setting 插件中：
  - 创建 `tube_specs` 表（使用 dbDelta 或现有迁移机制）；
  - 定义基础字段：`id, product_id, size_label, etrto_range, valve_type, segment?, execution?`。

### Step 2：Admin Tube Specs 页面

- 在 WP 后台注册一个新菜单页 `Tube Specs`；
- 实现：
  - 列表视图（分页 + 简单搜索）；
  - 新增 / 编辑表单：
    - 加载内胎商品列表（根据分类 / tag 过滤）；
  - 保存逻辑写入 `tube_specs` 表；

### Step 3：扩展 REST API / WP JSON 输出

- 在现有 `tanzanite/v1/products` 接口中：
  - ✅ 已实现：在 `Tanzanite_REST_Products_Controller::prepare_item_for_response()` 中，按 `product_id` 从 `{$wpdb->prefix}tanz_tube_specs` 读取 tube 规格；
  - ✅ 已实现：将读取结果挂载到返回 JSON 的 `tubeSpecs` 字段下，结构大致为：
    ```jsonc
    {
      "id": 123,
      "title": "Inner Tube ...",
      // ... 现有字段 ...
      "tubeSpecs": [
        {
          "size_label": "700x25-32C",
          "etrto_range": "622x25-32",
          "valve_family": "SV",
          "valve_angle_deg": 0,
          "valve_length_mm": 60,
          "execution": "STANDARD",
          "segment": "road",
          "notes": "..."
        }
      ]
    }
    ```
  - ✅ 已实现：在 `get_items()` 中解析 tube 相关查询参数（`tube_execution`、`tube_valve_family`、`tube_valve_angle`、`tube_valve_length`），调用 `find_product_ids_by_tube_specs()` 查询 `tanz_tube_specs`，并通过 `post__in` 收窄商品结果集；

### Step 4：前端高级搜索弹窗识别 tube 规格（可选增强）

- 在 Nuxt 的高级搜索组件中：
  - 当当前分类是「Tube」或某个「Tube 模式」时：
    - 可以在 UI 上显示 1–2 个 tube 专用 filter（例如 Valve Type）；
  - 过滤时将参数传回 `tanzanite/v1/products` 接口。

- **热门 TAB / chip ↔ 过滤条件 映射建议**：
  - 将现有热门 TAB（`Carbon rim`、`Sapim` 等）与 `Inner tube` 统一为一组配置化的 chip：
    - 每个 chip 至少包含：`id`、`label`、可选的 `categorySlug`、可选的 `tube_*` 预设；
  - 在 `ProductSearchPanel` 中：
    - 选中 `Inner tube` chip 时：
      - 前端追加一个「入口预设分类」字段（推荐使用 `category_slug: 'inner-tube'` 或在前端通过 slug → categoryId 映射得到 `categoryId`）；
      - 将来如需按 `tube_execution` / `tube_valve_*` 过滤时，也从 chip 配置表中读取并合并到 filters；
    - 取消 `Inner tube` chip 时：
      - 同时清除它所附带的 `category_slug` / `categoryId` 与所有 tube_* 预设，避免影响轮圈等其他搜索场景；
  - 其它 chip（如 `Carbon rim`）也走同样的映射路径，只是挂接到不同的分类 / 属性。

> 如果短期内只需要「按分类 = Tube」筛选出全部内胎，甚至可以先**不做** tube 规格 filter，只用分类 + 关键词检索，这样 REST 扩展可以留到后续。

### Step 5：Inner tube Tab 集成 CTA 按钮

- 在 `sizecharts.vue` 的 Inner tube section 中：
  - ✅ 已实现：在 `Inner tube` 标题下方添加了一个居中的 CTA 按钮；
  - ✅ 已实现：按钮点击时调用 `useShopSearchSheet().open({ presetCategorySlug: 'inner-tube', presetKeywords: ['Inner tube'] })` 打开现有的高级搜索 Bottom Sheet：
    - 在全局 state 中记录入口预设分类 slug；
    - 在高级搜索面板的热门搜索 TAB 区域中，自动勾选 `Inner tube` chip（与 `Carbon rim` / `Sapim` 等热门入口合并为同一组）；
  - TODO（可选增强）：在 `/shop` 页面或搜索面板内消费 `presetCategorySlug`，将 `inner-tube` slug 映射为具体分类 ID，并自动应用到搜索过滤条件中；同时在 ProductSearchPanel 中根据不同 chip 的配置，将 `Inner tube` 等入口 chip 转换为更精确的分类 / tube_* 过滤条件。

> 关于 slug / 分类稳定性的约定：
>
> - 前端长期应以 **slug 作为入口标识**（例如 `inner-tube`），而不是直接写死 term_id；
> - 如果后台删除旧分类并新建一个**slug 不同**的新分类（例如从 `inner-tube` 改为 `inner-tube-v2`）：
>   - 在「前端仅依赖 slug」或「REST API 提供 `category_slug` 参数」的设计下，旧 slug 将不再匹配任何分类，行为会自动退化为「不按内胎分类过滤」或返回空结果；
>   - 要恢复预设行为，需要同时在 **WordPress 分类** 和 **前端配置 / 文档** 中将 slug 更新为同一个值；
> - 因此推荐在上线前固定好一组长期稳定的 slug（如 `inner-tube`），后续尽量避免随意更改；如确需更名，应同步更新：
>   - WP 分类的 slug；
>   - REST API 中基于 slug 的查询逻辑（如有）；
>   - 前端中所有 `presetCategorySlug` / chip 配置使用的 slug 文本。

---

## 7. 总结

- **不扩展 Add new product 大表单**，避免核心插件继续变复杂；
- **新增 Tube Specs 配置页**，集中管理最多 20+ 款内胎的关键参数；
- **REST API 读取 Tube Specs**，为高级搜索提供更精确的 tube 筛选能力；
- **前端 `/guides/tireguides` Inner tube Tab**：
  - 仅作为“教育 + 入口”页面；
  - 使用一个 CTA 按钮打开已有高级搜索弹窗，并通过 `presetKeywords` 预选 `Inner tube` 热门 TAB；
- **/shop 页面 slug → id 映射已落地**：
  - 通过 `useShopSearchSheet` 提供的 `presetCategorySlug` 与 `useShopCategories()` 的分类列表，在 `/shop.vue` 中将 `inner-tube` 这类 slug 映射为具体 `ShopCategory`，写入 `selectedCategory`；
  - 这样从 Inner tube CTA 进入高级搜索并提交后，落地到 `/shop` 的首个搜索会真正按「内胎分类」收窄结果集；
  - slug 找不到匹配分类时，安全退化为「不预设分类」，避免出现“入口挂了就永远 0 结果”的情况。

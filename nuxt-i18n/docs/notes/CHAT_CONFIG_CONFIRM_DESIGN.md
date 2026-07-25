# Chat "Config Confirm" Feature Design

Last updated: 2026-07-25

> 一键「和客服确认配置」—— WhatsAppChatModal / Products Tab 内嵌，不打断聊天

## 1. 背景 & 目标

- 用户在 **聊天窗口** 内浏览 / 搜索商品，希望：
  - 选好款式、尺寸、金属、刻字等配置后，**一键把当前配置发给客服确认**。
  - 不需要跳转新页面或离开聊天上下文。
- 背景：
  - Products（`share`） Tab 已支持搜索，并通过 `WhatsAppProductSearchResultDrawer` 弹窗展示结果。
  - 旧版曾在关闭弹窗后继续于 Products Tab 下方渲染一份结果列表；该重复视图已清理，结果唯一展示源为抽屉。

本设计文档约定：

- **搜索结果的唯一展示视图** 为 `WhatsAppProductSearchResultDrawer`。
- Products Tab 下方区域将主要用于工具按钮 & 功能说明，而不是冗余的结果列表。


## 2. 当前架构概览（2026-07-25 audited state）

### 2.1 Products Tab（share）区域

- 移动端和桌面端：
  - 搜索输入框 + `Search` 按钮。
  - 一行 `History / Cart / Wishlist` 按钮。
  - Products Tab 不再渲染第二份结果列表；结果统一进入 `WhatsAppProductSearchResultDrawer`。

### 2.2 搜索结果弹窗：`WhatsAppProductSearchResultDrawer`

- 全局挂载在 `WhatsAppChatModal.vue` 末尾：

  ```vue
  <WhatsAppProductSearchResultDrawer
    v-model="productDrawerVisible"
    :loading="isSearching"
    :results="searchResults"
    :error="productDrawerError"
    :agent="selectedAgent"
    :query="productDrawerQuery"
    @close="handleProductDrawerClose"
    @select="shareProductToChat"
  />
  ```

- `searchProducts()` 在被触发时会：
  - 设置 `productDrawerVisible = true`。
  - 更新 `searchResults` / `productDrawerQuery` / loading 状态等。

> 结论：**用户搜索后第一视图一定是结果弹窗**，而不是 Tab 内部列表。


## 3. 功能目标（5. 一键「和客服确认配置」）

- 在 **搜索结果弹窗内部** 为每个商品提供入口：
  - `和客服确认配置` 按钮。
- 点击后：
  - 不离开聊天窗口；
  - 在同一个弹窗里切换到 **配置确认视图**；
  - 显示该商品的基础信息，以及未来的“配置表单”和“发送给客服”按钮。
- 聊天窗口里最终会收到一条结构化的“配置确认”卡片消息（第二阶段实现）。


## 4. 分阶段实现规划

### Phase 1：只搭「配置确认」视图的壳（当前优先）

**不依赖真实商品配置数据，仅需最少字段即可验证交互。**

#### 4.1 入口位置

- 放在 `WhatsAppProductSearchResultDrawer` 内部的商品卡片上：
  - 每个 product 卡片上，在原有“分享到聊天”的入口旁边/下方新增：
    - 按钮：`和客服确认配置`
  - 目前点击行为：
    - 只切换到配置确认视图，不调用 API，不发送消息。


#### 4.2 视图切换模式（推荐方案，已确认）

- 不再新增第二层弹窗，而是在 **同一个 `WhatsAppProductSearchResultDrawer` 内「切 view」**：

  - 新增内部状态（示意）：

    ```ts
    const viewMode = ref<'list' | 'configConfirm'>('list')
    const selectedConfigProduct = ref<Product | null>(null)
    ```

  - 列表视图（默认）：

    ```vue
    <template v-if="viewMode === 'list'">
      <!-- 现有搜索结果卡片列表 -->
    </template>
    ```

  - 配置确认视图：

    ```vue
    <template v-else-if="viewMode === 'configConfirm'">
      <!-- 新的配置确认页面壳 -->
    </template>
    ```

- 打开配置确认视图时：

  ```ts
  const openConfigConfirm = (product: Product) => {
    selectedConfigProduct.value = product
    viewMode.value = 'configConfirm'
  }
  ```

- 返回列表视图：

  ```ts
  const backToList = () => {
    viewMode.value = 'list'
    // 视情况是否清空 selectedConfigProduct
  }
  ```


#### 4.3 配置确认视图（壳结构）

**头部：**

- 标题：`和客服确认配置`。
- 左侧返回按钮（`<` 或 `Back`）：调用 `backToList()`。
- 右上角关闭按钮继续使用抽屉的全局关闭逻辑。

**主体内容（占位，可用于后期验证数据是否正确传入）：**

- 产品基础信息：
  - 缩略图：`selectedConfigProduct.thumbnail`（如有）。
  - 标题：`selectedConfigProduct.title`。
  - 价格：`selectedConfigProduct.price`（如有）。
- 一块说明性占位文本，例如：

  > 这里将展示此商品的可配置选项（尺寸、金属、刻字等），以及发送给客服的确认按钮。

**底部操作区（占位）：**

- 主按钮：`发送配置给客服`（占位实现）：
  - 第一阶段可仅做：
    - `console.log('config confirm clicked', selectedConfigProduct)` 或
    - 弹出一个 Toast 提醒：“配置确认功能开发中”。


#### 4.4 抽屉关闭行为

- 关闭 `WhatsAppProductSearchResultDrawer` 时，除现有逻辑外，还需重置：

  ```ts
  const handleProductDrawerClose = () => {
    productDrawerVisible.value = false
    productDrawerError.value = null
    productDrawerQuery.value = ''
    searchQuery.value = ''

    // 新增：重置配置确认相关状态
    viewMode.value = 'list'
    selectedConfigProduct.value = null
  }
  ```

> 第一阶段完成后：即使当前没有真实商品数据，也可以通过 mock / 假数据验证：
> - 搜索 → 打开结果抽屉；
> - 点击某条商品的「和客服确认配置」→ 切到配置确认视图，展示该商品基本信息；
> - 点击返回 → 回到列表视图；
> - 关闭抽屉 → 所有状态重置。

#### 4.5 进度快照（2026-07-25）

- ✅ `WhatsAppProductSearchResultDrawer` 内已经加入 `viewMode` / `selectedConfigProduct` 状态，并在关闭时统一重置。
- ✅ 商品卡片新增「分享至聊天」与「和客服确认配置」双按钮，后者会记录所选商品并切换到配置确认视图。
- ✅ 配置确认视图完成第一版壳结构：
  - 头部包含返回按钮 + 标题。
  - 左侧展示已选商品缩略图 / 标题 / 价格，右侧提供占位区说明后续会放配置字段与禁用的 CTA（“发送配置给客服”）。
- ✅ 交互流程已打通：搜索 → 查看结果 → 进入配置确认 → 返回列表 → 关闭抽屉，全程保持暗色玻璃风格一致，不再触发模板报错。
- ✅ `发送配置给客服` 已接入现有 Public Chat HTTP + SSE 链路，生成并持久化 `message_type = config_confirm` 的结构化消息。
- ✅ `ticket_messages` 增加 `message_type` 与 `metadata` 字段，Nuxt 用户端和 Admin 客服端都从同一条持久化消息渲染配置确认卡片。
- ✅ 公共商品搜索响应已经暴露现有事实源里的 `product_type.spec_definitions`、`variants.option_values`、`variants.weight_grams`。
- ✅ Nuxt 已通过 `app/composables/chat/useProductConfigConfirmPayload.ts` 统一生成配置确认 metadata：
  - `metadata.product`：商品和选中 SKU 的展示信息；
  - `metadata.selections`：选中 variant、SKU、option rows、库存、重量、价格；
  - 用户端聊天卡片和 Admin 客服卡片都只读渲染这份持久化 metadata。
- ✅ Orders Tab 已新增显式 `和客服确认订单` 入口，订单消息沿用 `message_type = order`，通过持久化 metadata 在 Nuxt 用户端和 Admin 客服端渲染订单确认卡片。
- ⚠️ 待办：如果后续要做“基于历史订单再次购买 / 修改配置”，只能从订单 items 的 SKU/attributes 回填到独立配置确认流程；后续后台产品模板字段继续扩展时，只更新商品/SKU事实源与 payload 构造器，不在聊天模块新增第二套配置字段。


### Phase 2：接入真实配置字段 & 发送

配置卡片

> 注：结构化消息发送与双端卡片渲染已经完成；当前可渲染 SKU variant 的真实 option / 重量 / 价格事实。下一步扩展只应继续读取产品模板与 variant 数据，不能在聊天侧重建一套配置模板。

#### 4.5 配置字段（示例）

具体字段待和产品 / 后端再确认，这里给出一个参考结构：

```ts
interface ProductConfigSelection {
  size?: string          // 戒指尺码等
  metal?: string         // 金属材质，如 18K White Gold
  centerStone?: {
    carat?: number
    color?: string
    clarity?: string
  }
  engraving?: string     // 刻字内容
  notes?: string         // 用户附加备注
}
```

配置确认视图将：

- 读取商品自身可选项（如 SKU 选项，或从单独接口获取）。
- 使用表单控件（下拉、多选、文本输入）让用户完成选择。


#### 4.6 发送到聊天的消息结构（建议）

当用户点击「发送配置给客服」时：

- 通过现有的聊天发送接口发出一条特殊类型消息，例如：

```jsonc
{
  "type": "config_confirm",
  "product": {
    "id": 123,
    "title": "...",
    "thumbnail": "...",
    "price": "..."
  },
  "config": {
    "size": "13",
    "metal": "18K White Gold",
    "centerStone": {
      "carat": 1.0,
      "color": "D",
      "clarity": "VVS1"
    },
    "engraving": "Forever Love",
    "notes": "希望日常佩戴舒适，适合细手指"
  }
}
```

- 聊天窗口中渲染为一张“配置确认卡片”，类似：

  > **Configuration confirmation request**  \
  > Product: xxx  \
  > Size: 13  \
  > Metal: 18K White Gold  \
  > Center stone: 1.0ct / D / VVS1  \
  > Engraving: “Forever Love”  \
  >  \
  > _"Please confirm if this configuration is suitable for me."_

- 客服端也以相同卡片样式展示，方便快速核对。

当前实现约定：

- `message` 字段存可读摘要，例如 `Configuration confirmation request: {product title}`。
- `message_type` 固定为 `config_confirm`。
- `metadata.product` 存产品基础信息：`id`、`variant_id`、`title`、`slug`、`sku`、`url`、`thumbnail`、`price`、`price_value`。
- `metadata.selections` 存选中 SKU 事实：`variant_id`、`variant_title`、`sku`、`options[]`、`stock`、`weight_grams`、`price`、`price_value`。
- `metadata.note` 存客服可读说明。


### Phase 3（可选）：Orders Tab 集成

在 Orders Tab 的每个订单卡片增加：

- 按钮：`再次和客服确认配置` / `Confirm configuration again`。
- 行为：
  - 打开同一个 `WhatsAppProductSearchResultDrawer` 或单独的配置确认视图，
  - 预填该订单的商品和配置（如历史订单中已有尺寸 / 金属等），
  - 允许用户再次发送配置确认卡片给客服。


## 5. Products Tab 自身的简化（已完成）

> 这部分是对现有 UI 的整理，不属于配置确认核心功能，但与信息架构有关。

- 已去除 Products Tab 内部的重复搜索结果列表，搜索结果唯一展示源为 `WhatsAppProductSearchResultDrawer`。
- Products Tab 下方区域现在专注于：
  - 搜索输入 + 工具按钮（History / Cart / Wishlist / 未来的 Member 等）。
  - 产品相关说明 / 功能入口（例如：提示用户可以在结果弹窗中使用“和客服确认配置”）。

该整理已完成；后续新增商品搜索交互必须继续复用抽屉，不要在 Products Tab 内重新渲染结果卡片。


## 6. 待办清单（实现用 Checklist）

### Phase 1：壳

- [x] 在 `WhatsAppProductSearchResultDrawer` 中新增内部状态：`viewMode`、`selectedConfigProduct`。
- [x] 为商品卡片增加 `和客服确认配置` 按钮，并实现 `openConfigConfirm(product)`。
- [x] 在弹窗内添加 `viewMode === 'configConfirm'` 视图：
  - [x] 头部：标题 + 返回按钮。
  - [x] 主体：商品基础信息 + 占位说明文案。
  - [x] 底部：占位主按钮（目前禁用显示“即将上线”提示）。
- [x] 关闭弹窗时重置 `viewMode` 和 `selectedConfigProduct`。
- [x] 删除 Products Tab 内重复搜索结果视图，保留抽屉作为唯一结果展示源。

### Phase 2：内容 & 发送

- [x] 明确当前商品可配置字段来源：产品模板定义 + SKU variant `option_values` / `weight_grams`。
- [ ] 在配置确认视图中渲染未来非 SKU 扩展表单：备注 / 客制化文字 / 其他后台定义的可编辑字段。
- [x] 点击「发送配置给客服」时构造 `config_confirm` 类型消息，并通过现有发送接口发出。
- [x] 在聊天前端实现 `config_confirm` 卡片消息样式；客服端也同步实现同样逻辑。
- [x] 将 `message_type` / `metadata` 持久化到 `ticket_messages`，避免刷新后退化为普通文本。

### Phase 3：Orders 集成（可选）

- [x] 在 Orders Tab 订单卡片中新增显式 `和客服确认订单` 按钮，避免整卡误触发送。
- [x] 订单消息沿用 `message_type = order`，通过 `app/composables/chat/useOrderChatPayload.ts` 统一生成订单 metadata。
- [x] Nuxt 用户端与 Admin 客服端都按结构化订单卡片渲染同一条持久化消息。
- [ ] 后续如果要“基于历史订单再次购买 / 修改配置”，应从订单 items 的 SKU/attributes 回填到独立配置确认流程，不在订单卡片里直接编辑。


## 7. 备注

- 本文档仅描述前端交互与数据结构建议，不包含具体 API 实现细节。
- 当前 DEV 环境下尚未上架真实商品数据，Phase 1 可以通过 mock 数据验证流转；
  Phase 2/3 需要在商品与会员体系完整后再推进。

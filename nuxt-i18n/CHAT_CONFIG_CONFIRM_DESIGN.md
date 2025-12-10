# Chat "Config Confirm" Feature Design

> 一键「和客服确认配置」—— WhatsAppChatModal / Products Tab 内嵌，不打断聊天

## 1. 背景 & 目标

- 用户在 **聊天窗口** 内浏览 / 搜索商品，希望：
  - 选好款式、尺寸、金属、刻字等配置后，**一键把当前配置发给客服确认**。
  - 不需要跳转新页面或离开聊天上下文。
- 现状：
  - Products（`share`） Tab 已支持搜索，并通过 `WhatsAppProductSearchResultDrawer` 弹窗展示结果。
  - 关闭弹窗后，Products Tab 下方仍会渲染一份结果列表（被认为是多余设计）。

本设计文档约定：

- **搜索结果的唯一展示视图** 为 `WhatsAppProductSearchResultDrawer`。
- Products Tab 下方区域将主要用于工具按钮 & 功能说明，而不是冗余的结果列表。


## 2. 当前架构概览（2025-12 DEV 状态）

### 2.1 Products Tab（share）区域

- 移动端：
  - 搜索输入框 + `Search` 按钮。
  - 一行 `History / Cart / Wishlist` 按钮。
  - `v-if="!productDrawerVisible"` 时，在 Tab 内部渲染一个按 `searchResults` 的列表（计划后续移除）。

- 桌面端：
  - 搜索输入框 + `Search` 按钮。
  - 一行 `History / Cart / Wishlist` 按钮。
  - `v-if="searchResults.length > 0 && !productDrawerVisible"` 时，渲染结果网格（同样计划后续移除）。

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


### Phase 2：接入真实配置字段 & 发送配置卡片

> 注：本阶段在商品和配置数据准备好之后再实现，这里仅做设计占位。

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


### Phase 3（可选）：Orders Tab 集成

在 Orders Tab 的每个订单卡片增加：

- 按钮：`再次和客服确认配置` / `Confirm configuration again`。
- 行为：
  - 打开同一个 `WhatsAppProductSearchResultDrawer` 或单独的配置确认视图，
  - 预填该订单的商品和配置（如历史订单中已有尺寸 / 金属等），
  - 允许用户再次发送配置确认卡片给客服。


## 5. Products Tab 自身的简化（后续小优化）

> 这部分是对现有 UI 的整理，不属于 5 的核心功能，但与信息架构有关。

- 去除 / 精简 Products Tab 内部 `!productDrawerVisible` 时展示的搜索结果列表：
  - 移动端：`v-if="!productDrawerVisible"` 的列表块。
  - 桌面端：`searchResults.length > 0 && !productDrawerVisible` 的网格块。
- 这样可以让 Products Tab 下方区域专注于：
  - 搜索输入 + 工具按钮（History / Cart / Wishlist / 未来的 Member 等）。
  - 产品相关说明 / 功能入口（例如：提示用户可以在结果弹窗中使用“和客服确认配置”）。

该整理并非必须在 Phase 1 同步完成，但设计上默认后续会朝“只保留弹窗视图”的方向收敛。


## 6. 待办清单（实现用 Checklist）

### Phase 1：壳

- [ ] 在 `WhatsAppProductSearchResultDrawer` 中新增内部状态：`viewMode`、`selectedConfigProduct`。
- [ ] 为商品卡片增加 `和客服确认配置` 按钮，并实现 `openConfigConfirm(product)`。
- [ ] 在弹窗内添加 `viewMode === 'configConfirm'` 视图：
  - [ ] 头部：标题 + 返回按钮。
  - [ ] 主体：商品基础信息 + 占位说明文案。
  - [ ] 底部：占位主按钮，点击仅做 log / toast。
- [ ] 关闭弹窗时重置 `viewMode` 和 `selectedConfigProduct`。

### Phase 2：内容 & 发送

- [ ] 明确商品可配置字段来源（SKU 数据结构 / 单独接口）。
- [ ] 在配置确认视图中渲染真实配置表单：尺寸 / 金属 / 刻字等。
- [ ] 点击「发送配置给客服」时构造 `config_confirm` 类型消息，并通过现有发送接口发出。
- [ ] 在聊天前端实现 `config_confirm` 卡片消息样式；客服端也同步实现或复用同样逻辑。

### Phase 3：Orders 集成（可选）

- [ ] 在 Orders Tab 订单卡片中新增 `再次和客服确认配置` 按钮。
- [ ] 打开配置确认视图时，预填订单中的历史配置。


## 7. 备注

- 本文档仅描述前端交互与数据结构建议，不包含具体 API 实现细节。
- 当前 DEV 环境下尚未上架真实商品数据，Phase 1 可以通过 mock 数据验证流转；
  Phase 2/3 需要在商品与会员体系完整后再推进。

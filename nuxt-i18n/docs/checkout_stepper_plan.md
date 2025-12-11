# Checkout Stepper 改造计划

## 当前状态
- `CheckoutModal.vue` 恢复为原始结构（仍使用旧的 Tabs + 内容块）。
- `CheckoutStepper.vue` 已搭建完成 Step A 骨架：包含 Step 1/Step 2 的完整布局、示例数据驱动的 `currentStep` / `activeMethod` 状态与事件；Step 2 目前使用占位内容等待正式迁移。

## 目标
1. 在 `CheckoutStepper.vue` 中实现完整的两步流程：
   - **Step 1**：支付方式长条卡片（点击展开详细说明）。
   - **Step 2**：优惠券、积分、收货地址、订单摘要等所有现有块。
   - 包含返回 Step 1 的入口以及移动端/桌面端的说明。
2. 所有说明/表单/CTA 都放在 `CheckoutStepper.vue`，便于独立调试。
3. 组件完成后，再在 `CheckoutModal.vue` 中引入 `<CheckoutStepper />`，并清理旧结构。

## 执行步骤
### Step A：在 CheckoutStepper.vue 中搭建骨架
1. 复制 Mockup HTML，生成 Step 1/Step 2 的基本结构。注意：当前 HTML 中多出一个整体的 “credit / debit cards” 说明块，后续必须把这些说明移入各卡片的下拉内容中，确保用户点击具体卡片才看到对应详情。
2. 加入 `currentStep`、`activeMethod` 状态（ref + emits），先用示例数据跑通切换。
3. 确保 Step 1 卡片展开、Step 2 框架与返回按钮都能正常切换。

### Step B：迁移现有内容到 Stepper
1. 将 “Pay with …” 的说明复制到 Step 1 卡片的展开区域。
2. Step 2 中按模块复制（优惠券 → 积分 → Notes → 地址 → 订单摘要 → CTA）。
   - 每迁移一个模块，先运行确认正常，再从 `CheckoutModal.vue` 中移除。
3. 根据移动端/桌面端差异，必要时拆分 slot 或加入 `v-if` 控制。

### Step C：组件接入
1. 在 `CheckoutModal.vue` 的安全提示下方挂入 `<CheckoutStepper />`，传递需要的 props/事件。
2. 将旧的 Tabs / 内容块删除，并避免重复逻辑。
3. 处理 lint / TS 报错（例如 `EventTarget` 类型、未使用的变量等）。

### Step D：验证
1. 桌面端核对：Step 1/Step 2 切换、滚动、CTA 触发。
2. 移动端核对：卡片展开、底部 CTA。
3. 运行构建检查（`pnpm dev`/`pnpm build`）确保无语法/模板错误。

## 备注
- 在 `CheckoutStepper.vue` 完整之前，不要改动 `CheckoutModal.vue`。
- 每次迁移模块前先备份（例如 `CheckoutModal0-xx.vue.bak`），便于回滚。
- 文案/样式与原文件保持一致，后续若需新增说明，优先在 Stepper 中完成。

## 最新进度
- 2025-12-12：完成 Step A，`CheckoutStepper.vue` 具备双步骤 UI、示例交互、占位 CTA，后续可逐块迁移真实内容。
- 2025-12-12：`CheckoutModal.vue` 中已在 SSL 安全提示下方挂入 `<CheckoutStepper />`（使用 `initial-step` / `initial-method` 示例参数），暂未移除旧结构，方便并行对比。
- 2025-12-12：完成 Coupon 模块对接。`CheckoutStepper.vue` 的优惠券卡片已与 `CheckoutModal.vue` 的真实状态/事件串联，按钮禁用、loading、成功提示全部基于 `couponCode` / `isApplyingCoupon` / `calculation.appliedCoupon`。
- 2025-12-12：完成 Points 模块对接。积分卡片已放在 Coupon 下方，可根据 `calculation.userPoints`、`usePointsDiscount`、`pointsToUse` 等实时状态进行开关与输入同步，并在无积分可用时显示禁用提示。
- 2025-12-12：Step 2 区域已在 `CheckoutStepper.vue` 内部加入 `max-height + overflow-y-auto`，不修改 `CheckoutModal.vue` 的情况下即可滚动查看全部卡片。
- 2025-12-12：完成 Order Summary 模块迁移。`CheckoutStepper.vue` 直接读取 `priceBreakdown` / `cartItems` 生成商品列表及费用明细，含运费状态、积分/优惠券折扣展示，结构保持与旧卡片一致。
- 2025-12-12：完成 Shipping Address 表单迁移。`CheckoutStepper.vue` 接收 `form`/`countrySearch`/`shippingValidation`/`estimatedDelivery` 等 props 并抛出 `update-shipping-field` 等事件，所有字段、提示、不可配送处理与原 `CheckoutModal.vue` 等效。

完成 Step A/B/C 后，再更新本文件记录结果。***

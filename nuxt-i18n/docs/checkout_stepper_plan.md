# Checkout Stepper 改造计划

## 当前状态
- `CheckoutModal.vue` 模板层仅保留弹窗外壳 + `<CheckoutStepper />`，移动端内容拉伸至全宽且弹窗内部整体支持滚动。
- `CheckoutStepper.vue` 已完成三步流（Step 1：支付方式、Step 2：Shipping Address、Step 3：Coupon/Points/Notes/Order Summary/CTA），并针对移动端/桌面端做了独立布局优化（如隐藏步骤文案、卡片居中、按钮上移等）。
- DEV 环境仍使用本地 Tanzanite Setting（尚未上传插件配置），因此部分配送规则 / 费用模板尚未完善。为了继续调试后续步骤，`validateShipping` 在 DEV 环境会 fallback 为可配送，待上生产前移除/恢复真实校验。

## 目标
1. 保持 `CheckoutModal.vue` 与 Stepper 的松耦合：父组件仅负责状态/校验/事件中转，UI 都在 Stepper 内部。
2. Stepper 维持 **三步流程**，并根据后续需求可扩展更多步骤（如支付 SDK 接入等）。
3. Tanzanite Setting 插件同步最新的运费模板后，验证 Step 2 的 shipping 提示、运费标签、Estimated delivery 等逻辑。
4. 完成 QA（桌面+移动）与 lint/TS 清理，为上线做准备。
5. 上线前撤掉 DEV-only 的 shipping fallback，并确认 Tanzanite 插件模板已更新。

## 执行步骤
### Step A：在 CheckoutStepper.vue 中搭建骨架（✅ 完成）
1. 复制 Mockup HTML，生成 Step 1/Step 2 的基本结构。
2. 加入 `currentStep`、`activeMethod` 状态（ref + emits），先用示例数据跑通切换。
3. 确保 Step 1 卡片展开、Step 2 框架与返回按钮都能正常切换。

### Step B：迁移现有内容到 Stepper（✅ 完成）
1. 将 “Pay with …” 的说明复制到 Step 1 卡片的展开区域。
2. Step 2 中按模块复制（优惠券 → 积分 → Notes → 地址 → 订单摘要 → CTA）。
3. 根据移动端/桌面端差异，必要时拆分 slot 或加入 `v-if` 控制。

### Step C：组件接入（✅ 完成）
1. 在 `CheckoutModal.vue` 的安全提示下方挂入 `<CheckoutStepper />`，传递需要的 props/事件。
2. 将旧的 Tabs / 内容块删除，并避免重复逻辑。
3. 处理 lint / TS 报告（例如 `EventTarget` 类型、未使用的变量等）。

### Step D：验证（进行中）
1. 桌面端核对：Step 1/Step 2/Step 3 切换、滚动、CTA 触发。
2. 移动端核对：卡片展开、底部 CTA。
3. 运行构建检查（`pnpm dev`/`pnpm build`）确保无语法/模板错误。
4. Tanzanite Setting 插件运费模板更新后，再次验证不可配送提示/Estimated delivery/Shipping label。

## 备注
- 旧的 Tabs + 内容块已删除，若需要回顾旧代码，可读取 `CheckoutModal*.vue.bak`。
- DEV 环境暂未接入最新运费模板，`shippingValidation` 结果以占位数据为准；上线前需确认插件数据。
- 文案/样式与原文件保持一致，后续若需新增说明，优先在 Stepper 中完成。

## 最新进度
- 2025-12-12：完成 Step A，`CheckoutStepper.vue` 具备基础 UI、示例交互。
- 2025-12-12：完成 Coupon / Points / Order Summary / Shipping Address 模块迁移。
- 2025-12-12：Stepper 改造为三步流（Step 1/2/3），按钮/状态流转全部更新。
- 2025-12-12：`CheckoutModal.vue` 传递真实 props（coupon、points、shipping、CTA 等）并移除旧模板结构；Step 状态同步，移动端 padding/高度调整完毕。
- 2025-12-12：移动端 Step 1/Step 2/Step 3 均支持独立滚动，按钮位置与卡片居中问题已解决；桌面 Coupon/Order Notes 卡片细节完成微调。
- 2025-12-12：`validateShipping` 在 DEV（`import.meta.dev`）模式下新增 fallback，允许在模板为空时继续流程；上线前需移除或确保模板已加载。

后续更新：待 Tanzanite 插件完成运费模板配置并验证 Step 2 提示，再记录上线前 QA 结论；同时计划统一处理 lint/TS 报错。***

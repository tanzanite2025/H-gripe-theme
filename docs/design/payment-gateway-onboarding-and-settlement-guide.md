# 💳 Stripe & PayPal 生产网关准入、多币种原币结算与管理端指引规范

> **文档定位**：系统设计与运维准入规范（Design & Ops Guide）。  
> **核心目标**：指导在管理后台（Admin Console）的 Stripe 与 PayPal 域内规划专属的**配置指引 TAB 页面**；为未来新上线、更换商户账号、或多主体入驻提供**不可遗漏的网关后台配置全家桶清单**，彻底杜绝多币种汇率侵蚀、Webhook 漏配、拒付无保护等严重资金与运维风险。  
> **实施说明**：**本文档仅提供架构规范与自写代码指引，不直接修改工程业务代码**。

---

## 📌 一、为什么必须在管理端内置“指引 TAB 页面”？

在海外 DTC 自行车独立站的生命周期中，支付网关的接入与维护通常面临以下痛点：
1. **记忆衰退与信息孤岛**：初期配置完成后，可能隔半年、一年才会更换商户主体、升级账号或处理异常；当时的踩坑经验早已被遗忘；
2. **多币种隐形扣费**：若未在 Stripe/PayPal 开启原币种多币种账户，每次跨国结算都会被网关强行换汇两次，并遭遇汇率波动亏空；
3. **Webhook 漏配事故**：漏勾选一个 `charge.refunded` 或 `PAYMENT.CAPTURE.REFUNDED`，系统退款与订单状态就会彻底失联；
4. **内置 SOP 的价值**：将外部网关复杂的后台配置要点**产品化、资产化**，直接做成管理后台对应渠道下的一个只读/指引 TAB 页面。任何运维人员打开后台，照着勾选即可 100% 准确上线。

---

## 🎨 二、管理端指引 TAB 页面设计规范 (UDS 1.0)

新页面应挂载在管理后台的 **风控与支付策略模块**（例如 `RiskStrategy.vue` 或 `PaymentMethodsSettingsPanel.vue`）中，分别作为 **Stripe 域** 与 **PayPal 域** 下的一个独立子 TAB。

### 1. 页面排布与视觉参数 (严格遵循 ERP UDS 1.0)
- **根容器布局**：`flex flex-col gap-8 animate-in fade-in duration-700`（禁止在根容器使用 ad-hoc `p-6`，保证页眉高度对齐）。
- **模块页眉 (Module Header)**：
  - 样式：`rounded-[32px] border border-dashed border-border/80 bg-muted/5 p-6`。
  - 主标题：`text-lg font-black tracking-tighter italic uppercase text-foreground`（如 `STRIPE ONBOARDING & SETTLEMENT GUIDE`）。
  - 微缩说明：`text-[9px] font-black uppercase tracking-widest opacity-60`。
- **检查卡片 (Checklist Card)**：
  - 容器：`rounded-[24px] border border-dashed border-border/80 bg-card p-6 shadow-sm`。
  - 卡片标题：`text-sm font-black tracking-tighter italic text-foreground`。
  - 状态 Badge：`rounded-full h-5 text-[8px] font-mono`，标注 `CRITICAL (必开)`、`RECOMMENDED (推荐)`、`ADVANCED (进阶)`。
- **快捷工具栏**：
  - 提供一键复制当前站点的 Webhook URL（如 `https://your-domain.com/api/v1/payments/stripe/webhook`）；
  - 提供外链直接跳转至 Stripe Dashboard / PayPal Developer Console 对应设置页的胶囊按钮 (`rounded-full h-9 text-[10px] font-black uppercase`)。

---

## ⚡ 三、Stripe 域：必须开启的配置全家桶清单 (Must-Have Checklist)

在 Stripe Dashboard 接入我们系统时，**必须且只能**按照以下清单完成核对：

### 1. 核心项 A：开启多币种原生结算账户 (Multi-Currency Settlement)
- **风险背景**：
  如果 Stripe 只绑了 USD 银行卡，欧洲买家付 1,000 EUR 时，Stripe 自动折算为 USD（收取约 2% 换汇费）；退款时 Stripe 再次按退款日即时汇率折算为 EUR。两次换汇加上 30 天汇率波动，商家承受无谓的汇差磨损。
- **Stripe 后台操作路径**：
  `Settings (设置) -> Bank accounts and scheduling (银行账户与调度) -> Manage currencies / Add bank account`。
- **必须配置**：
  - 添加 **EUR 本地银行账户**（可使用 Wise、Airwallex、Payoneer 签发的欧洲虚拟 IBAN）；
  - 添加 **GBP 本地银行账户**（英国 Sort Code + Account Number）；
  - 开启 **Like-for-like settlement (原币结算)**。
- **业务效果**：EUR 付款直接入账 EUR 余额，退款直接从 EUR 余额扣划。**0 汇率转换费，0 汇率波动风险**。

### 2. 核心项 B：Webhooks 必开事件清单 (Endpoints Checklist)
- **Webhook 接收地址**：`https://[域名]/api/v1/payments/stripe/webhook`
- **必须勾选监听的 14 个核心事件（漏配将导致订单/资金失联）**：
  1. `payment_intent.succeeded` —— 支付成功入账（核心扣款确认）
  2. `payment_intent.payment_failed` —— 支付失败（记录失败原因与日志）
  3. `payment_intent.requires_action` —— 3DS 强认证发起（触发买家挑战验证）
  4. `payment_intent.processing` —— 异步支付中（如某些欧洲本地银行转账）
  5. `charge.refunded` —— 退款成功完成（网关资金划扣落地）
  6. `refund.created` —— 退款意向建立
  7. `refund.updated` —— 退款状态变更（含部分退款）
  8. `charge.dispute.created` —— 买家发起信用卡拒付（立即触发订单发货冻结 `FulfillmentHold=true`）
  9. `charge.dispute.updated` —— 拒付证据状态更新
  10. `charge.dispute.funds_withdrawn` —— 争议资金暂扣
  11. `charge.dispute.funds_reinstated` —— 争议胜诉资金返还
  12. `charge.dispute.closed` —— 争议结案
  13. `radar.early_fraud_warning.created` —— 早期欺诈预警（EFW，提前 24~48 小时预警盗刷，可主动退款避免 Dispute 罚金）
  14. `review.opened` / `review.closed` —— 人工风控审核开启/关闭
- **重要技术约束**：
  - 拷贝生成的 `whsec_...` 密钥填入后端环境变量 `STRIPE_WEBHOOK_SECRET`；
  - 确保服务器系统时钟与标准 NTP 偏差不超过 300 秒（5 分钟），否则触发重放防护阻断。

### 3. 核心项 C：Radar 风控与 3DS 责任转移 (Liability Shift)
- **Stripe 后台操作路径**：`Radar -> Rules (规则引擎)`
- **必须配置的规则**：
  - **规则 1 (高危拦截)**：`Block if :risk_level: = 'highest'`；
  - **规则 2 (高客单强制 3DS)**：对于自行车高端轮组（如金额 > 500 USD / EUR），配置规则：  
    `Request 3D Secure if :amount_in_usd: > 500`；
  - **规则 3 (跨国 AVS 不符审核)**：`Review if :address_line1_check: = 'fail' or :address_zip_check: = 'fail'`。
- **业务价值**：一旦 3DS 挑战通过，发生欺诈拒付时由 Visa/Mastercard 发卡行承担损失（责任转移 Liability Shifted），商家无需自掏腰包赔付。

### 4. 核心项 D：结账账单地址强制收集 (Full Billing Address Collection)
- **Stripe 后台操作路径**：`Settings -> Checkout and Payment Links`
- **必须配置**：
  - 勾选 `Require customers to provide a billing address`；
  - 确保收集买家完整的街道地址与邮编，否则无法与发卡行系统做 AVS（Address Verification System）匹配。

---

## ⚡ 四、PayPal 域：必须开启的配置全家桶清单 (Must-Have Checklist)

PayPal 的商户体系与 Stripe 存在较大差异，尤其在多币种处理上存在**“默认强制换汇”**的著名陷阱，必须严格按以下步骤设置：

### 1. 核心项 A：开启非本币付款“自动接收原币余额”（致命易漏点）
- **陷阱背景**：
  PayPal 企业的默认设置是：如果账户主货币是 USD，当欧洲买家用 EUR 付款时，PayPal 会发出邮件提示“您收到一笔外币”，若 30 天未手动处理会退回；或者默认强制使用 **PayPal 极高的零售汇率（包含 3%~4% 汇差点差）自动转换为 USD**！
- **PayPal 后台操作路径**：
  `Account Settings (账户设置) -> Payment preferences (付款首选项) -> Block payments (阻止付款 / 付款接收偏好)`
- **必须配置**：
  - 找到设置项：`Allow payments sent to me in a currency I do not hold (允许以我不持有的货币向我发送付款)`；
  - 勾选：👉 **`Yes, accept and create a balance in that currency (是，接受并以该货币创建余额)`**。
  - **绝对不要勾选**：“Ask me” 或 “No, deny payments”。
- **业务效果**：买家付 EUR 就保留在 EUR 余额，付 GBP 就保留在 GBP 余额，后续退款直接从对应原币余额划扣，避免高达 4% 的换汇损耗。

### 2. 核心项 B：PayPal Developer REST Webhook 配置
- **Webhook 接收地址**：`https://[域名]/api/v1/payments/paypal/webhook`
- **必须勾选监听的 8 个核心事件**：
  1. `CHECKOUT.ORDER.APPROVED` —— 买家在 PayPal 弹窗完成授权
  2. `PAYMENT.CAPTURE.COMPLETED` —— 扣款成功到账
  3. `PAYMENT.CAPTURE.DENIED` —— 扣款被拒
  4. `PAYMENT.CAPTURE.REFUNDED` —— 退款成功完成
  5. `CUSTOMER.DISPUTE.CREATED` —— 买家发起 PayPal 纠纷/补偿申请（立即冻结发货）
  6. `CUSTOMER.DISPUTE.RESOLVED` —— 纠纷结案（释放冻结或确认损失）
  7. `CUSTOMER.DISPUTE.UPDATED` —— 纠纷补充材料流转
  8. `RISK.DISPUTE.CREATED` —— PayPal 内部风控争议
- **重要技术约束**：
  - 将生成的 `Webhook ID` 填入后端配置 `PAYPAL_WEBHOOK_ID`；
  - 系统验签接口需通过调用 PayPal API 验证 transmission headers，需确保服务器对外访问 `api.paypal.com` 网络畅通。

### 3. 核心项 C：卖家保障计划（Seller Protection）达标条件
- **发货规则**：
  - 发货地址必须**严格等于** PayPal Capture 返回的 `shipping_address`；如果客户下单后邮件要求改送其他地址，**严禁私下改发**！必须全额退款后让买家在 PayPal 页面修改地址重新下单；
  - 发货后必须提供具有在线可查妥投记录（POD / Signature）的国际物流单号（DHL/FedEx），并在管理后台触发 PayPal Tracking API 同步。

---

## 💻 五、管理端 Vue 3 组件实现结构指引 (参考代码骨架)

供您后续在管理端自写代码时对照实现（建议放在 `web/admin/src/views/risk-strategy/` 或对应设置组件中）：

```vue
<!-- PaymentGatewayOnboardingGuideTab.vue (实现骨架参考) -->
<template>
  <div class="flex flex-col gap-8 animate-in fade-in duration-700">
    <!-- 模块页眉 -->
    <header class="rounded-[32px] border border-dashed border-border/80 bg-muted/5 p-6">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">
            Payment Gateway Setup SOP
          </span>
          <h1 class="text-lg font-black tracking-tighter italic uppercase text-foreground">
            {{ activeProvider === 'paypal' ? 'PayPal' : 'Stripe' }} 生产网关准入与多币种配置指南
          </h1>
          <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/60 mt-1">
            新开商户号、更换主体或日常巡检时，必须核对以下关键配置，防止汇率亏空与状态脱节。
          </p>
        </div>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" class="rounded-full h-9 text-[10px] font-black uppercase" @click="copyWebhookUrl">
            复制 Webhook 接收地址
          </Button>
          <Button size="sm" class="rounded-full h-9 text-[10px] font-black uppercase" @click="openExternalDashboard">
            跳转官方网关后台 ↗
          </Button>
        </div>
      </div>
    </header>

    <!-- 核心排查卡片矩阵 (Grid) -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- 卡片 1: 多币种原币结算 -->
      <section class="rounded-[24px] border border-dashed border-border/80 bg-card p-6 flex flex-col justify-between">
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-black tracking-tighter italic uppercase text-foreground">
              1. 多币种原币结算配置
            </h2>
            <span class="rounded-full bg-rose-500/10 text-rose-600 px-2 py-0.5 text-[8px] font-mono font-bold">
              CRITICAL 必开
            </span>
          </div>
          <p class="text-xs text-muted-foreground leading-relaxed">
            {{ activeProvider === 'paypal' 
                ? '必须在 PayPal 设置「允许以我不持有的货币向我发送付款 -> 是，接受并创建余额」，否则触发 4% 强制换汇损耗。'
                : '必须在 Stripe 绑定 EUR 本地 IBAN（Wise/Airwallex）并开启 Like-for-like 原币结算，彻底避免二次换汇和汇率时间差亏空。' }}
          </p>
          <div class="rounded-xl bg-muted/40 p-3 text-[11px] font-mono space-y-1">
            <div class="text-muted-foreground">路径: {{ activeProvider === 'paypal' ? 'Account Settings -> Payment Preferences' : 'Settings -> Bank accounts and scheduling' }}</div>
            <div class="text-emerald-600 font-bold">✓ 核心目标：同币种入账，同币种退款，0 汇率风险</div>
          </div>
        </div>
      </section>

      <!-- 卡片 2: Webhook 必须监听事件 -->
      <section class="rounded-[24px] border border-dashed border-border/80 bg-card p-6 flex flex-col justify-between">
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-black tracking-tighter italic uppercase text-foreground">
              2. Webhook 必开事件清单
            </h2>
            <span class="rounded-full bg-rose-500/10 text-rose-600 px-2 py-0.5 text-[8px] font-mono font-bold">
              CRITICAL 必开
            </span>
          </div>
          <p class="text-xs text-muted-foreground leading-relaxed">
            必须确保扣款、退款、争议这三大生命周期事件 100% 勾选。
          </p>
          <ul class="text-[11px] font-mono space-y-1 text-muted-foreground">
            <li v-for="evt in currentEvents" :key="evt">• {{ evt }}</li>
          </ul>
        </div>
      </section>
      
      <!-- 更多卡片: 3DS/Radar 规则、账单地址、发货地址保障等 -->
    </div>
  </div>
</template>
```

---

## 📌 六、总结核对口诀 (Quick Summary)

| 渠道 | 核心致命项 | 漏配后果 | 达标标志 |
| :--- | :--- | :--- | :--- |
| **Stripe** | **EUR/GBP 原币结算账户** | 退款遭遇二次换汇损失与汇率时间差亏空 | 欧洲付款直接进入 EUR 余额，退款直接划扣 EUR，0 换汇 |
| **Stripe** | **14 项 Webhook 事件全家桶** | 订单支付成功后不流转、退款丢失凭证、漏接拒付 | Webhook 接收端无 400 签名超时报错，事件实时推送 |
| **Stripe** | **Radar 高客单强制 3DS 规则** | 发生无卡盗刷拒付时商家全额兜底赔偿 | 高于 $500 交易强制触发 3DS，争议享受发卡行责任转移 |
| **PayPal** | **非本位币“接受并创建余额”** | 欧元付款被拦截或被 PayPal 4% 极差汇率强行吃掉 | 收到外币付款时自动建立对应币种余额 |
| **PayPal** | **REST Capture & Dispute Webhook** | 买家发起纠纷后台毫无感知，包裹照常发出致钱货两空 | 收到纠纷时系统秒级冻结发货（`FulfillmentHold`） |

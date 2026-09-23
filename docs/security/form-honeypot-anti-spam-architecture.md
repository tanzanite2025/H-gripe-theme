# 独立站无感防刷与 Honeypot 蜜罐防御架构设计规范

> **文档路径**: `docs/security/form-honeypot-anti-spam-architecture.md`  
> **适用场景**: 邮件订阅 (Newsletter)、客户留言 (Contact/Inquiry)、用户注册、质保初审等公开表单接口  
> **设计目标**: **“0 用户摩擦、尽量少的页面延迟、最小化数据收集、0 SEO 干扰”**。Honeypot 是低成本的第一层信号，不承诺拦截所有机器人，也不替代限流、邮件验证、账户保护或支付风控。

> **当前项目基线（2026-09）**：Newsletter 和质保邮件验证已经接入 Cloudflare Turnstile、Redis/进程内限流及邮件挑战；反馈接口已经接入认证和 Redis 限流。Honeypot 将以叠加方式接入，第一阶段不删除现有防线。文档中的“0 第三方请求”和“天然符合 GDPR”不是默认事实，必须结合部署配置、数据处理协议和安全/合规评审确认。

---

## 目录

1. [为什么采用 Honeypot 作为第一层防线？](#1-为什么采用-honeypot-作为第一层防线)
2. [Honeypot 蜜罐的核心工作原理](#2-honeypot-蜜罐的核心工作原理)
3. [对搜索引擎爬虫 (Googlebot) 的低干扰边界](#3-对搜索引擎爬虫-googlebot-的低干扰边界)
4. [前端工业级诱饵埋设规范 (Vue 3 / Nuxt)](#4-前端工业级诱饵埋设规范-vue-3--nuxt)
5. [后端 Go 校验机制与“静默丢弃 (Silent Drop)”准则](#5-后端-go-校验机制与静默丢弃-silent-drop-准则)
6. [进阶防线：时间差陷阱 (Timestamp Trap)](#6-进阶防线时间差陷阱-timestamp-trap)
7. [全站安全分级联动矩阵](#7-全站安全分级联动矩阵)
8. [落地方案与代码改造检查清单](#8-落地方案与代码改造检查清单)

---

## 1. 为什么采用 Honeypot 作为第一层防线？

在高端自行车零配件独立站（高客单价 $500 ~ $3,000+）中，买家每前进一步都至关重要：

| 对比维度 | 交互式验证码 (reCAPTCHA / hCaptcha / Turnstile) | Honeypot 蜜罐方案 |
| :--- | :--- | :--- |
| **买家体验** | 可能出现弹窗、找斑马线/红绿灯，增加跳出风险 | 正常用户无需额外交互；仍需通过自动填充和无障碍回归验证 |
| **页面性能** | 可能加载第三方 JS 并产生额外网络往返；是否加载取决于站点策略 | Honeypot 本身不需要额外脚本；若同时保留 Turnstile，仍会产生其网络开销 |
| **隐私与合规** | 需要评估供应商、传输、Cookie/设备信号和法律依据 | 只提交一个空/非空字段，但日志、IP、邮箱仍需按站点隐私政策处理；不能据此自动宣称 GDPR 合规 |
| **维护成本** | 依赖外部服务可用性、密钥和供应商策略 | 无第三方依赖，但容易被高级机器人识别，必须和限流、验证、监控联动 |

Honeypot 的正确定位是“便宜且透明的第一道筛选”。对于会发送邮件、创建质保索赔、修改账户状态或涉及支付的接口，不能仅依赖 Honeypot。

---

## 2. Honeypot 蜜罐的核心工作原理

网络上的自动化垃圾脚本（Spam Bots）绝大多数是通过无头浏览器或正则表达式扫描页面的 HTML DOM 结构。它们的目标是在全网所有表单中群发广告、注入博彩/钓鱼链接、或者刷爆网站的邮件发信额度。

```
                       [ 正常人类买家 ]                      [ 恶意群发垃圾脚本 ]
                              │                                      │
                              ▼                                      ▼
                      浏览器渲染 CSS 样式                     直接扫描 HTML DOM 结构
                              │                                      │
                      ┌───────┴───────┐                      ┌───────┴───────┐
                      │ 看到正常表单: │                      │ 看到所有字段: │
                      │ • 邮箱: [填写]│                      │ • 邮箱: [填写]│
                      │ • 诱饵: [隐形]│                      │ • 诱饵: [无脑填满]
                      └───────┬───────┘                      └───────┬───────┘
                              │ 提交: 诱饵为空                       │ 提交: 诱饵有值
                              ▼                                      ▼
                     ┌────────────────────────────────────────────────────────┐
                     │                 后端 Go API 蜜罐网关校验                │
                     └────────────────────────────────────────────────────────┘
                                     │                                      │
                          [诱饵为空: 判定为真人]                 [诱饵有值: 判定为机器人]
                                     │                                      │
                                     ▼                                      ▼
                              正常入库并发送邮件                    【静默丢弃 (Silent Drop)】
                                                               假装成功返回 200，不入库不发信
```

---

## 3. 对搜索引擎爬虫 (Googlebot) 的低干扰边界

许多开发者担心：*“加了隐藏字段，会不会被 Google 认为是作弊甚至被降权惩罚？”*

**预期不会影响正常抓取，但不能写成绝对保证。**
1. 常规搜索引擎抓取主要通过 HTTP `GET` 解析页面和链接，不会因为一个不参与导航的隐藏输入框而改变页面正文或链接结构。
2. Honeypot 不能通过给搜索引擎和用户展示不同内容来实现；字段必须是页面真实 DOM 的一部分，且不影响可见文案、结构化数据、链接和渲染结果。
3. 应在发布前用 Lighthouse、axe/屏幕阅读器和真实 Googlebot/搜索控制台检查，确认没有布局偏移、键盘焦点、无障碍或抓取异常。
4. 拦截 POST 只发生在 API 业务入口，不应对商品页面、sitemap、robots.txt 或公开 GET 资源加 Honeypot 逻辑。

---

## 4. 前端工业级诱饵埋设规范 (Vue 3 / Nuxt)

黑客脚本现在也会扫描诸如 `name="honeypot"` 或带有 `style="display:none"` 的明显标记。因此，埋设诱饵必须具备**“高度伪装性”**与**“无障碍兼容性”**：

### 4.1 字段命名规范（以假乱真）
- ❌ **严禁使用低幼命名**：`honeypot`, `bot_trap`, `is_bot`, `hidden_field`
- ✅ **推荐使用看似合法且按表单区分的 decoy 命名**：
  - `corporate_website`（企业网址）
  - `fax_number`（传真号码）
  - `company_tax_id`（公司税号）
  - `secondary_phone`（备用联系电话）

### 4.2 前端代码实现样板 (Nuxt 组件)

```vue
<template>
  <form @submit.prevent="handleSubmit" class="space-y-4">
    <!-- 真实输入框 -->
    <div>
      <label for="email" class="block text-xs font-bold uppercase">Your Email</label>
      <input id="email" v-model="form.email" type="email" required class="h-10 w-full rounded border px-3" />
    </div>

    <!-- 工业级 Honeypot 诱饵：真实 DOM 中存在，但不参与视觉布局和键盘导航 -->
    <div
      class="sr-only"
      aria-hidden="true"
      style="position: absolute; left: -9999px; top: -9999px; opacity: 0; pointer-events: none; width: 0; height: 0; overflow: hidden;"
    >
      <label for="corporate_tax_number">Corporate tax number</label>
      <input
        id="corporate_tax_number"
        v-model="form.corporate_tax_number"
        type="text"
        name="corporate_tax_number"
        tabindex="-1"
        autocomplete="off"
      />
    </div>

    <button type="submit" class="rounded-full bg-primary px-6 py-2 text-xs font-black uppercase text-white">
      Subscribe
    </button>
  </form>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

const form = reactive({
  email: '',
  corporate_tax_number: '' // 诱饵字段默认必须为空字符串
})

const handleSubmit = async () => {
  // 正常回传给后端
  await $fetch('/api/v1/marketing/subscribe', {
    method: 'POST',
    body: form
  })
}
</script>
```

项目中的 `SubscriptionOptIn.vue` 当前还会在提交时按配置执行 Turnstile。第一阶段只增加 Honeypot，不删除 Turnstile；是否在后续把 Turnstile 改为高风险兜底，需要根据拦截指标、误判率和安全评审决定。

#### 关键防死角参数：
- `tabindex="-1"`：防止正常人类用 `Tab` 键切换光标时不小心移入该框。
- `aria-hidden="true"`：让读屏软件忽略诱饵区域；由于其中包含输入控件，必须同时用 `tabindex="-1"`，并通过 axe/实际读屏软件验证，避免出现“可聚焦元素位于 aria-hidden 区域”的无障碍问题。必要时补充 `inert`。
- `autocomplete="off"`：防止浏览器密码管理器或自动填充插件（如 1Password、Chrome 自动保存）误填此框。

不要把字段标签写成“Do not fill this field”并暴露给读屏软件；应使用中性的业务字段名和隐藏区域，并确认密码管理器、浏览器自动填充不会误填。字段默认值必须是空字符串，提交失败或重试时也不能残留上一次的诱饵值。

---

## 5. 后端 Go 校验机制与“静默丢弃 (Silent Drop)”准则

在后端 API 设计中，对捕获的机器人处理有一个至高准则：**“永远不要向机器人大声报警！”**

### ❌ 错误做法：返回 400 或 403 报错
- 如果返回 `{"error": "Bot detected! Honeypot filled."}`，黑客只要看一眼报错，立刻就会修改爬虫脚本排除这个字段。

### ✅ 正确做法：静默假成功 (Silent Drop / Tarpit)
- 命中后立即结束业务逻辑，**不入数据库、不发通知邮件、不上传文件、不扣减发信或验证配额**。
- 返回该接口正常成功时使用的状态码和尽可能相同的通用响应。不要强制所有接口都返回 `200`：例如当前订阅和质保验证使用 `202`，反馈和质保索赔使用 `201`。
- 在 Prometheus 和结构化日志中记录 `honeypot_blocked_total`；不得记录原始邮箱、IP 或完整请求体。

### 实现边界（必须遵守）

不要把下面的逻辑直接做成一个无差别的全局 Middleware：

1. `c.PostForm()` 只适用于表单/multipart，不能读取 JSON body。
2. Middleware 读取 JSON body 后必须恢复 body，否则后续 Handler 无法再次绑定。
3. multipart 质保索赔必须在对象存储上传之前检查诱饵字段。

推荐在每个 Handler 完成 body 解析后，用共享 helper 判断字段，再调用该接口的成功响应函数：

```go
import (
	"strings"

	"commerce-platform/internal/pkg/honeypot"
)

type SubscribeRequest struct {
	Email              string `json:"email"`
	Source             string `json:"source"`
	Locale             string `json:"locale"`
	CorporateTaxNumber string `json:"corporate_tax_number"`
}

func (req SubscribeRequest) IsDecoyFilled() bool {
	return strings.TrimSpace(req.CorporateTaxNumber) != ""
}

func (h *Handler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 真实字段仍按接口契约返回校验错误。
		return
	}
	if req.IsDecoyFilled() {
		honeypot.RecordBlocked("newsletter", "corporate_tax_number", c.Request.URL.Path)
		acceptedSubscriptionResponse(c) // 当前接口为 202；不触发服务、邮件或配额逻辑
		return
	}
	// 继续真实的验证、限流、入库和邮件流程。
}
```

对需要绑定校验的接口，应先解析不带 `binding:"required"` 的外层请求，优先识别诱饵，再对真实字段执行校验；否则机器人可能从 400 响应中获得额外特征。具体取舍需按接口风险和现有测试确定。

---

## 6. 进阶防线：时间差陷阱 (Timestamp Trap)

时间差只能作为辅助信号，不能直接相信浏览器提交的 `_ts`：客户端可以任意修改时间戳，自动填充、重试和无障碍工具也可能让真人提交很快。

### 分阶段规则

1. **观测阶段（当前实现）**：Newsletter 页面挂载后调用 `GET /api/v1/subscriptions/timing-token`。后端返回短期、不可缓存的服务端签名 token；提交时只回传 token，后端验证后记录耗时，不拦截请求。没有 token、token 过期或签名错误都不能阻塞真实用户。
2. token 载荷为 `version + form + issued_at(Unix ms) + nonce`，签名为 HMAC-SHA256。优先使用服务端专用的 `HONEYPOT_TIMING_SECRET`；未配置时从 `JWT_SECRET` 派生域隔离密钥，浏览器永远不会拿到签名密钥。默认有效期为 900 秒，可由 `HONEYPOT_TIMING_TTL_SECONDS` 调整。
3. 当前记录 `commerce_platform_honeypot_timing_evaluations_total{form,result}`（`valid` / `missing` / `malformed` / `invalid` / `expired`）、`commerce_platform_honeypot_timing_seconds{form}` 直方图，以及 Redis shadow 检查的 `commerce_platform_honeypot_timing_replays_total{form,result}`（`first_seen` / `reused` / `error`）。Redis 只保存 nonce 的哈希和短 TTL，重复 token 仍然放行；进入 enforce 前再评估是否把 `reused` 变成拒绝条件。Redis 故障对表单请求 fail-open，不记录邮箱、IP、原始 nonce 或请求体。
4. `1.5 秒`只能作为初始观测阈值，必须按真实流量校准；不能单独用于账户锁定、质保拒绝或支付拒绝。未来若启用硬拦截，必须保留 `off / shadow / enforce` 回滚开关，并先在 Newsletter 小流量灰度。

### 6.1 观测发布门槛

- 先以 `shadow` 运行至少 7 天，并确认 Newsletter 有足够的真实样本（建议不少于 1,000 次有效 token 提交）。
- 用 `commerce_platform_honeypot_timing_seconds` 的 P50/P95/P99 和 `valid/missing/expired` 比例建立正常基线；不要把单个浏览器、单个地区或单次极短提交当成阈值依据。
- 检查 `honeypot_timing_replays_total{result="reused"}` 是否主要来自用户重试、多标签页或网络重放；Redis `error` 不得被误判为机器人。
- 只有在误判率、Redis 可用性和正常订阅副作用回归均达标后，才允许小流量启用 enforce；任何异常都通过 `HONEYPOT_MODE=shadow` 或 `off` 回滚。

---

## 7. 全站安全分级联动矩阵

将各种安全防线合理落位，各司其职，坚决不搞“一刀切”：

| 业务场景 / 接口 | 第一防线 | 第二防线 | 应急后备防线 | 买家端感知度 |
| :--- | :--- | :--- | :--- | :--- |
| **商品浏览 / 列表 / 算法计算器** | Cloudflare CDN 边缘缓存 | Redis 令牌桶高频限流 | 无 | **0 摩擦（极速秒开）** |
| **邮件订阅 / 质保留言 / 建议反馈** | **Honeypot 蜜罐诱饵** | 现有限流、邮件挑战或认证 | Redis 单 IP/目标配额（按需增加） | **0 摩擦（完全透明）** |
| **买家登录 / 注册** | Honeypot 蜜罐诱饵 | 密码重试次数冻结 (5次) | 账号锁定 15 分钟 | **0 摩擦** |
| **找回密码 / 发送验证邮件** | Redis 验证码 60秒冷却 | Honeypot 蜜罐 | Cloudflare Turnstile (仅在连续发送失败时) | **0 摩擦** (除非恶意刷接口) |
| **订单结算 / 信用卡支付创建** | 3DS 强责任转移 | Stripe 欺诈评分 Radar | 连续支付失败 3 次触发 Turnstile 阻断试卡 | **0 摩擦** (正常买家绝不弹窗) |

---

## 8. 落地方案与代码改造检查清单

后续进行表单加固时，按以下阶段执行，不要一次性全站启用：

- [x] **阶段 0：基线与开关**：确认现有 Turnstile、认证、邮件挑战和限流行为；增加 `anti_abuse.honeypot_mode`（`off` / `shadow` / `enforce`），支持观测和回滚。
- [x] **阶段 1：共享能力**：增加前端 `HoneypotField`、Go helper、`honeypot_blocked_total` 指标和 `[HONEYPOT_BLOCKED]` 结构化日志。
- [x] **阶段 2：Newsletter**：在 `SubscriptionOptIn.vue` 增加 `corporate_tax_number`，后端订阅 Handler 在邮件验证和入库前静默丢弃。
- [x] **阶段 3：反馈**：页面反馈已增加 `fax_number`；建议反馈后端入口已增加 `company_tax_id`，待确认实际前端页面后再补埋设。
- [x] **阶段 4：质保**：在验证和索赔 multipart 流程增加 `secondary_phone`，必须在对象存储上传前检查；保留邮件验证和 Turnstile。
- [x] **阶段 5：认证**：最后在登录/注册增加 `corporate_website`，先观察误判，再决定是否拦截；Google 登录不套用普通表单字段。
- [x] **时间差陷阱（shadow 实现）**：Newsletter 获取服务端签名 token，提交时验证并记录耗时；Redis 记录 nonce 的首次/重复状态；不信任裸客户端时间戳，不因 token 缺失/失效拦截。
- [ ] **生产 enforce 签核**：用真实分布校准阈值，决定是否将 Redis shadow 记录升级为 nonce 一次性拒绝，并按 `off / shadow / enforce` 灰度。该项是上线策略验收，不是代码实现阻塞条件。
- [ ] **测试与发布**：覆盖正常请求、副作用、静默响应、重复 token、附件上传、键盘导航、axe/读屏、构建和 SEO/GET 回归。

第一批发布的验收条件：Newsletter 正常请求仍创建 pending 订阅并发送一次确认邮件；诱饵非空时返回原接口的成功状态，但数据库、邮件发送器和验证配额均无变化。

本次已完成阶段 0 至阶段 5 的后端/已确认前端范围，并完成 Newsletter 时间差陷阱的 shadow 实现；suggestion-feedback 目前没有确认中的前端页面，因此只保护了后端创建入口。Newsletter、页面反馈、质保和登录/注册通过 `anti_abuse.honeypot_mode` 控制，默认 `enforce`，可切换到 `shadow` 观测或 `off` 回滚。时间 token 当前只观测，不参与业务拒绝；后端单测、路由回归、Redis shadow replay 检查、监控面板和前端类型检查已完成。剩余的 7 天/1,000 次有效提交是生产 enforce 的数据验收条件，真实浏览器的键盘/读屏/axe 回归仍需在发布环境执行。

---

> **结语**：Honeypot 是低摩擦、低成本的第一方拦截信号。只有和现有限流、邮件挑战、认证、存储保护、监控及必要的第三方风控组合使用，才能形成可审计、可回滚、对正常买家透明的防刷防线。任何 GDPR、SEO 或拦截率结论都应以实际部署配置、测试证据和合规评审为准。

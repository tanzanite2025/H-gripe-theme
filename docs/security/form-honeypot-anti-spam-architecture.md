# 独立站无感防刷与 Honeypot 蜜罐防御架构设计规范

> **文档路径**: `docs/security/form-honeypot-anti-spam-architecture.md`  
> **适用场景**: 邮件订阅 (Newsletter)、客户留言 (Contact/Inquiry)、用户注册、质保初审等公开表单接口  
> **设计哲学**: **“0 用户摩擦、0 页面延迟、0 隐私泄露 (GDPR 免申报)、0 SEO 干扰”**。用最轻量的工程手段，阻断 95% 以上的无脑自动化爬虫与垃圾邮件脚本。

---

## 目录

1. [为什么放弃传统验证码，选择 Honeypot？](#1-为什么放弃传统验证码选择-honeypot)
2. [Honeypot 蜜罐的核心工作原理](#2-honeypot-蜜罐的核心工作原理)
3. [对搜索引擎爬虫 (Googlebot) 的 0 干扰证明](#3-对搜索引擎爬虫-googlebot-的-0-干扰证明)
4. [前端工业级诱饵埋设规范 (Vue 3 / Nuxt)](#4-前端工业级诱饵埋设规范-vue-3--nuxt)
5. [后端 Go 校验机制与“静默丢弃 (Silent Drop)”准则](#5-后端-go-校验机制与静默丢弃-silent-drop-准则)
6. [进阶防线：时间差陷阱 (Timestamp Trap)](#6-进阶防线时间差陷阱-timestamp-trap)
7. [全站安全分级联动矩阵](#7-全站安全分级联动矩阵)
8. [落地方案与代码改造检查清单](#8-落地方案与代码改造检查清单)

---

## 1. 为什么放弃传统验证码，选择 Honeypot？

在高端自行车零配件独立站（高客单价 $500 ~ $3,000+）中，买家每前进一步都至关重要：

| 对比维度 | 传统验证码 (reCAPTCHA / hCaptcha) | Honeypot 蜜罐方案 |
| :--- | :--- | :--- |
| **买家体验** | 弹窗、找斑马线/红绿灯，挫败感极强，**导致 10%~20% 跳出** | **完全隐形，买家 100% 毫无察觉，0 摩擦感** |
| **页面性能** | 加载 200KB+ 第三方 JS，多次网络往返，Lighthouse 跑分大跌 | **0 KB 额外体积，0 网络开销，页面秒开** |
| **欧洲 GDPR 合规** | 回传用户设备指纹至境外，未授权弹窗前加载涉嫌违规 | **纯第一方内生逻辑，不收集任何隐私，天然符合 GDPR** |
| **维护成本** | 依赖外部 API 密钥，第三方服务宕机时表单卡死 | **完全自主掌控，无任何外部网络依赖** |

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

## 3. 对搜索引擎爬虫 (Googlebot) 的 0 干扰证明

许多开发者担心：*“加了隐藏字段，会不会被 Google 认为是作弊甚至被降权惩罚？”*

**答案：绝对不会，0 影响。** 理由如下：
1. **Googlebot 只抓取，不填表**：Google 官方文档明确指出，Google 抓取蜘蛛只发起 HTTP `GET` 请求来解析超链接和文本内容，**Google 绝不会自动填写表单并发送 `POST` 请求**。
2. **Google 明确认可 Honeypot 模式**：Google 官方多次确认，用于阻止垃圾邮件的屏幕阅读器可访问隐藏字段（Screen-reader accessible forms for spam prevention）属于行业标准做法，不属于 Cloaking（障眼法黑帽作弊）。
3. **更利好 SEO**：由于垃圾脚本的大量无效请求在入口处被瞬时清空，保护了服务器 CPU 和数据库连接池，使得 Google 爬虫在抓取商品主页时响应速度更快。

---

## 4. 前端工业级诱饵埋设规范 (Vue 3 / Nuxt)

黑客脚本现在也会扫描诸如 `name="honeypot"` 或带有 `style="display:none"` 的明显标记。因此，埋设诱饵必须具备**“高度伪装性”**与**“无障碍兼容性”**：

### 4.1 字段命名规范（以假乱真）
- ❌ **严禁使用低幼命名**：`honeypot`, `bot_trap`, `is_bot`, `hidden_field`
- ✅ **推荐使用看似合法的 decoy 命名**：
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

    <!-- 工业级 Honeypot 诱饵 (人类肉眼不可见，无障碍屏幕阅读器忽略，脚本无脑填写) -->
    <div
      class="sr-only"
      aria-hidden="true"
      style="position: absolute; left: -9999px; top: -9999px; opacity: 0; pointer-events: none; width: 0; height: 0; overflow: hidden;"
    >
      <label for="corporate_tax_number">Corporate Tax Number (Do not fill this field)</label>
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

#### 关键防死角参数：
- `tabindex="-1"`：防止正常人类用 `Tab` 键切换光标时不小心移入该框。
- `aria-hidden="true"`：告诉盲人读屏软件“忽略此元素”，不影响无障碍合规。
- `autocomplete="off"`：防止浏览器密码管理器或自动填充插件（如 1Password、Chrome 自动保存）误填此框。

---

## 5. 后端 Go 校验机制与“静默丢弃 (Silent Drop)”准则

在后端 API 设计中，对捕获的机器人处理有一个至高准则：**“永远不要向机器人大声报警！”**

### ❌ 错误做法：返回 400 或 403 报错
- 如果返回 `{"error": "Bot detected! Honeypot filled."}`，黑客只要看一眼报错，立刻就会修改爬虫脚本排除这个字段。

### ✅ 正确做法：静默假成功 (Silent Drop / Tarpit)
- **直接返回 `HTTP 200 OK: {"code": 0, "message": "Success"}`**。
- **但在程序内部**：直接退出逻辑，**不入数据库、不发任何通知邮件、不扣减发信配额**。
- **同时记录审计日志与指标监控**：在 Prometheus / 日志中记录 `honeypot_dropped_total + 1`。

### Go 核心实现示例：

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HoneypotGuard 校验诱饵字段。如果包含任何内容，假装成功并静默截断。
func HoneypotGuard(decoyFieldName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅针对表单 POST/PUT 请求
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
			c.Next()
			return
		}

		// 获取请求体中的诱饵值 (适用于 JSON 或 Form)
		decoyValue := strings.TrimSpace(c.PostForm(decoyFieldName))
		
		// 若为 JSON 请求，由具体 Handler 或结构体校验
		if decoyValue != "" {
			// [CRITICAL] 识别出机器人：静默假成功并阻断后续执行
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "Subscription received successfully",
			})
			return
		}

		c.Next()
	}
}
```

在请求体 DTO 中定义：
```go
type SubscribeRequest struct {
    Email              string `json:"email" binding:"required,email"`
    CorporateTaxNumber string `json:"corporate_tax_number"` // 诱饵字段
}

func (req *SubscribeRequest) IsBot() bool {
    return strings.TrimSpace(req.CorporateTaxNumber) != ""
}
```

---

## 6. 进阶防线：时间差陷阱 (Timestamp Trap)

为了防范稍微高级一点的爬虫，可以与“时间差检测”结合使用：

- **基本事实**：一个正常人类用户打开网页、看懂文案、输入邮箱并点击提交，耗时**至少需要 3~5 秒以上**。
- **机器人的特征**：脚本请求网页后，在 50~200 毫秒内即完成全自动 POST 提交。

### 规则：
- 前端页面加载表单时，写入一个当前时间戳 `_ts`（可做简单的 base64 或签名混淆）；
- 提交给后端时，后端计算：`time_taken = NOW() - _ts`；
- 如果 `time_taken < 1.5 秒`，判定为自动化脚本瞬间提交，同样采取**静默丢弃**策略。

---

## 7. 全站安全分级联动矩阵

将各种安全防线合理落位，各司其职，坚决不搞“一刀切”：

| 业务场景 / 接口 | 第一防线 | 第二防线 | 应急后备防线 | 买家端感知度 |
| :--- | :--- | :--- | :--- | :--- |
| **商品浏览 / 列表 / 算法计算器** | Cloudflare CDN 边缘缓存 | Redis 令牌桶高频限流 | 无 | **0 摩擦（极速秒开）** |
| **邮件订阅 / 质保留言 / 建议反馈** | **Honeypot 蜜罐诱饵** | 1.5 秒时间差防秒刷 | Redis 单 IP 日配额 | **0 摩擦（完全透明）** |
| **买家登录 / 注册** | Honeypot 蜜罐诱饵 | 密码重试次数冻结 (5次) | 账号锁定 15 分钟 | **0 摩擦** |
| **找回密码 / 发送验证邮件** | Redis 验证码 60秒冷却 | Honeypot 蜜罐 | Cloudflare Turnstile (仅在连续发送失败时) | **0 摩擦** (除非恶意刷接口) |
| **订单结算 / 信用卡支付创建** | 3DS 强责任转移 | Stripe 欺诈评分 Radar | 连续支付失败 3 次触发 Turnstile 阻断试卡 | **0 摩擦** (正常买家绝不弹窗) |

---

## 8. 落地方案与代码改造检查清单

后续进行表单加固时，可对照以下清单逐项勾选实施：

- [ ] **表单字段埋设**：在 `SubscriptionOptIn.vue`（邮件订阅）中增加隐形 `corporate_tax_number` 诱饵字段；
- [ ] **留言与反馈表单**：在意见反馈和客服离线工单表单中增加隐形 `fax_number` 诱饵字段；
- [ ] **DTO 绑定与静默丢弃**：Go 后端在对应的 Request 结构体中解析诱饵值，若非空直接返回 `200 OK` 并打断逻辑；
- [ ] **审计度量记录**：新增日志输出 `[HONEYPOT_BLOCKED]`，便于后续在控制台观察拦截成效；
- [ ] **无障碍校验**：检查所有隐藏字段均具备 `tabindex="-1"` 与 `aria-hidden="true"`，保证残障人士读屏软件通过无误。

---

> **结语**：通过 Honeypot 蜜罐机制，我们用几行干净、优雅的第一方代码，在没有引入任何第三方臃肿脚本、没有触犯任何欧洲 GDPR 隐私条款、没有惊扰 Google 爬虫的前提下，筑起了一道坚固的静默防线。

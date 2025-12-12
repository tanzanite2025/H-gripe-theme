# Tanzanite（Nuxt 前端 + WordPress 后端）鉴权 / Cookie / CSRF 方案整理

## 0. 背景与当前现状（基于本项目代码）

### 0.1 项目架构

- 前端：Nuxt 3（SSR 开启，`nitro.preset = static`）
- 后端：WordPress（自写商店系统 + 自写 REST API，路径形如 `/wp-json/tanzanite/v1/...`）
- 站点：同域（浏览器请求可携带同域 Cookie）

### 0.2 当前 Nuxt → WP REST 的调用模式（现有实现）

在 Nuxt 代码中多处出现：

- `credentials: 'include'`：让浏览器自动携带 Cookie（同域会话）
- `X-WP-Nonce`：从 `runtimeConfig.public.wpNonce` 取值并注入 Header

可见这套体系属于：

- **Cookie 会话鉴权（WordPress 登录态）**
- **WP Nonce（用于 CSRF 防护/REST 请求校验的一部分）**

### 0.3 当前 WP 插件侧的登录与鉴权（现有实现）

在 `wp-plugin/tanzanite-setting/includes/rest-api/class-rest-chat-controller.php` 中：

- 登录接口 `/tanzanite/v1/chat/login`：
  - `wp_authenticate()` 验证用户
  - `wp_set_current_user()`
  - `wp_set_auth_cookie($user->ID, true)` **写入 WP 登录 Cookie**
- 多数业务接口：`permission_callback => 'is_user_logged_in'` **依赖 WP 会话**

因此：

- 当前项目核心鉴权是 **WP Cookie 会话**（不是 JWT）
- CSRF/REST 校验目前通过 **`X-WP-Nonce`** 来配合

> 注意：如果 `wpNonce` 被静态化写死（例如打包后固定），会引入“过期/不一致/不可刷新”的稳定性问题。

---

## 1. 你现在用 Cookie 的四类用途与“现代化方向”

### 1.1 登录（敏感）

- **推荐继续使用 Cookie**，但要使用“更安全、更现代”的配置：
  - `HttpOnly`（避免 JS 读取，降低 XSS 盗取风险）【优先级：高】
  - `Secure`（HTTPS 必须开启）【你已开始】
  - `SameSite=Lax`（同站点常规场景；如遇跨站支付回跳再评估）

#### 1.1.1 本项目（WP 会话）如何实现/确认 `HttpOnly`、`Secure`、`SameSite`

> 结论先写在前面：对于“登录 Cookie”，**HttpOnly 应该做且优先级高**；但它通常是后端（WP）设置的，不是前端 Nuxt 控制。

- `HttpOnly`
  - WordPress 的登录态 Cookie（由 `wp_set_auth_cookie()` 写入）在主流版本中默认就是 **HttpOnly**。
  - 你可以在浏览器 DevTools → Application → Cookies 查看对应 Cookie 的 **HttpOnly** 列是否为 true。
  - 如果不是（或你有自定义 cookie），应确保后端 `setcookie` 时启用 HttpOnly。

- `Secure`
  - 线上 HTTPS 场景下应该为 true。
  - 你可以在 DevTools 里确认 Cookie 的 **Secure** 列。
  - 若发现未开启，通常需要从 WordPress/服务器层面确保站点正确识别 HTTPS（例如反向代理/Cloudflare 场景的 HTTPS 头），并检查相关配置。

- `SameSite`
  - 同域 SPA/SSR 的常规交互，默认建议 `Lax`。
  - 如果你有“第三方站点跳转回来后继续保持登录/支付回跳”等跨站点场景，可能需要评估 `None; Secure`。
  - WordPress 对 SameSite 的控制方式与版本/服务器环境相关：
    - 若你计划“严格掌控 SameSite”，建议把它作为 **后端增强项**（需要在 WP 层统一设置 cookie options），而不是在 Nuxt 前端处理。

### 1.2 语言（非敏感偏好）

- 更现代做法：
  - `localStorage` 持久化（推荐）【你目前使用 nuxt-i18n，不确定是否用 localStorage，需要核对配置】
  - 或 URL 前缀（`/en/`、`/zh/`，SEO/分享更友好）【你说你“好像是这样设置的”，也需要核对配置】
- 只有在需要 SSR 首屏“无闪烁”读取偏好时，才考虑用非 HttpOnly Cookie 保存语言

#### 1.2.1 语言与 `HttpOnly Cookie` 的关系（重要）

- **语言偏好不建议用 `HttpOnly Cookie`**。
  - 原因：`HttpOnly` 的设计目的就是“不允许 JS 读取”。
  - 语言选择通常需要前端（Nuxt）读取并立即切换 UI，所以更适合：
    - `localStorage`（前端可读）
    - URL 前缀（无需存储也可恢复）
    - 或普通 cookie（非 HttpOnly，仅用于偏好）

#### 1.2.2 如何确认你当前 nuxt-i18n 属于哪一种（本项目核对方法）

- 方式一：检查 `nuxt-i18n` 配置
  - 在 `nuxt-i18n/nuxt.config.ts` 或相关 i18n 配置文件里，通常会有：
    - `strategy: 'prefix' | 'prefix_except_default' | 'no_prefix'`（决定是否使用 URL 前缀）
    - `detectBrowserLanguage`（决定是否检测并持久化）
      - 可能包含 `useCookie`、`cookieKey`、`alwaysRedirect` 等（决定是否用 cookie）

- 方式二：运行时验证（最快）
  - 切换语言后：
    - 看 URL 是否自动变为 `/en/...`、`/zh/...`（说明用了 URL 前缀策略）
    - 看浏览器 Storage：
      - Application → Local Storage 是否出现 `locale` 或类似 key
      - Application → Cookies 是否出现 `i18n_redirected`（nuxt-i18n 常见 cookie key）

- 你希望“更现代”的语言方案建议
  - 如果你已经使用 URL 前缀：优先保持（对 SEO 与分享友好）。【我已经使用URL前缀，因为我首要的也是对SEO与分享友好】
  - 如果你不想 URL 前缀：建议用 localStorage（或 nuxt-i18n 的默认持久化），不要用 HttpOnly。

### 1.3 购物车（你已是后端实现）

- 你现在购物车在后端插件中维护，这是正确方向。
- 前端建议：
  - 最多保存一个 `cartId`（如果需要匿名购物车）
  - 不建议把完整购物车 JSON 存 Cookie（4KB 限制 + 每次请求都携带 + 一致性风险）

### 1.4 统计（GA/Meta 等）

- 更现代做法：
  - **Consent（同意）后再加载第三方脚本**
  - 未同意不落第三方 Cookie
  - 同意状态持久化（cookie/localStorage 均可，但要分类“必要/统计/营销”）

---

## 2. 三条主要方案路线（A 升级版 / B2 / B1）

### 2.1 方案 A（升级版，推荐优先）：继续使用 WP Nonce，但改为“动态获取”

#### A 的目标

- 仍然使用 WP 的 nonce 生态（与 WP REST 兼容）
- 解决 `runtimeConfig.public.wpNonce` 静态化、过期、不一致的问题

#### A 的关键做法

- 后端新增一个接口（示例）：
  - `GET /wp-json/tanzanite/v1/auth/nonce`
  - 返回：`{ nonce: "..." }`
  - 生成：`wp_create_nonce('wp_rest')`（或你自定义的 action）
- 前端在以下时机拉取 nonce 并保存在内存（Pinia/useState）：
  - 应用启动时（可选）
  - 登录成功后（推荐）
  - nonce 校验失败时自动刷新（可选）

#### A 的优点

- 改动最小，保持 WP 体系一致
- nonce 与会话绑定逻辑更清晰
- 兼容 `permission_callback => is_user_logged_in` 的既有路由

#### A 的风险/注意点

- nonce 本身不是万能：
  - 能防 CSRF，但不能防 XSS
- 需要规范化前端请求封装：
  - 在统一 request 层注入 `X-WP-Nonce`

---

### 2.2 方案 B2（更彻底但风险可控）：保留 WP Cookie 会话，但自建 CSRF（替换 nonce）

> 这是“更彻底”的 CSRF 方案，但不破坏 WP 会话体系。

#### B2 的目标

- 保留 `wp_set_auth_cookie` / `is_user_logged_in` 的 WP 会话
- 放弃 WP nonce，改为你自己的 CSRF token 方案（可控、易理解）

#### B2 推荐实现：Double Submit Cookie

- 后端发放 CSRF token：
  - `GET /wp-json/tanzanite/v1/auth/csrf`
  - 服务器生成随机 token（例如 `wp_generate_password(32, false)` 或更安全的随机源）
  - `Set-Cookie: csrf_token=<token>; Secure; SameSite=Lax; Path=/`
  - 同时响应 body 也返回 token（可选）

- 前端发送写请求（POST/PUT/DELETE）时：
  - 从 cookie 读出 `csrf_token`
  - 加 header：`X-CSRF-Token: <token>`

- 后端校验：
  - 比较 `$_COOKIE['csrf_token']` 与 `$request->get_header('X-CSRF-Token')`
  - 使用 `hash_equals()` 防止时序攻击

#### B2 的优点

- CSRF 防护强度与 nonce 类似（关键在正确实现）
- 不依赖 nonce 生命周期/静态配置
- 与同域 cookie 会话非常匹配

#### B2 的风险/注意点

- CSRF cookie 必须可被 JS 读取（不能 HttpOnly）
- 如果站点发生 XSS：攻击者仍可以读取 CSRF token（因此 XSS 仍是关键风险点）
- 要注意 CORS/跨域策略（你当前同域，难度较低）

---

### 2.3 方案 B1（最彻底）：彻底脱离 WP Cookie 会话，改 JWT/自建 Session

#### B1 的目标

- 前端不再依赖 WP 登录 cookie
- 改为：
  - `Authorization: Bearer <access_token>`
  - 配合 refresh token（可放 HttpOnly cookie）或纯服务端 session

#### B1 在当前项目中的实际成本

由于你大量路由使用：

- `permission_callback => is_user_logged_in`
- 以及 WP 的 `get_current_user_id()` / `wp_get_current_user()`

若改为 JWT，你还需要在 WP 侧实现：

- 解析 `Authorization` Header
- 把 token 映射成 WP user（设置 current user），否则 `is_user_logged_in` 都会失败
- refresh/吊销/过期/黑名单、跨设备登出、异常检测等全套机制

#### B1 的优点

- API 更通用（未来可给 App/第三方使用）
- 不依赖浏览器 cookie 行为（某些跨域/嵌入场景更灵活）

#### B1 的风险/注意点

- 改动大、风险高、测试成本高
- 安全性是否更高取决于实现质量（不是“换 JWT 就自动更安全”）

---

## 3. 安全对比（按攻击面）

### 3.1 CSRF

- A（动态 nonce）：强 ✅
- B2（double submit CSRF）：强 ✅
- B1（JWT）：默认不受 CSRF 影响（因为不靠浏览器自动携带 Cookie），但如果 refresh 放 cookie 仍需 CSRF/额外策略

### 3.2 XSS

- 三种方案都无法“靠 CSRF 机制”解决 XSS
- XSS 防护重点：
  - CSP（Content-Security-Policy）
  - 严格转义/避免注入
  - 依赖审计
  - 不把敏感 token 存 localStorage

### 3.3 会话劫持/重放

关键不在 nonce/CSRF token，而在：

- HTTPS + `Secure`
- 合理过期策略
- 可选：绑定设备/刷新策略/异常 IP 检测

### 3.4 暴力破解

建议对登录接口增加：

- IP + username 限流
- 登录失败惩罚策略
- 必要时验证码

### 3.5 DoS/滥用

你在保修查询中已经做了限流（`transient`），建议把同类策略扩展到：

- 登录
- 下单/优惠码
- 聊天发送

> 限流（DoS/滥用防护）可以 **单独实施**，与“动态 nonce”或 CSRF 方案无依赖关系。你可以按业务风险优先级，选择先做限流或先做动态 nonce，两者互不阻塞。
---

## 4. 推荐路线（最少改动 + 最安全收益）

### 推荐顺序

1. **优先做 A 的升级版（动态 nonce）**
   - 最少改动
   - 解决 nonce 静态化/过期的根源问题
   - 与限流、统计等其他安全措施互不冲突，可以并行推进
2. 如果你仍然想完全掌控 CSRF：再做 **B2（double submit）**
3. 除非你要长期“多端开放 API / 脱离 WP 会话”，否则不建议直接上 **B1（JWT）**

---

## 5. 实施清单（你后续可在本文档内勾选/修改）

### 5.1 A（动态 nonce）实施清单

- [ ] 后端新增 `GET /tanzanite/v1/auth/nonce`
- [ ] 前端新增 `useWpNonce()`（或集成到 `useAuth` / 统一 request）
- [ ] 前端在登录成功后刷新 nonce
- [ ] 前端请求层统一注入 `X-WP-Nonce`
- [ ] nonce 失效时（403/401）自动刷新与重试策略（可选）

### 5.2 B2（double submit CSRF）实施清单

- [ ] 后端新增 `GET /tanzanite/v1/auth/csrf`（发 cookie）
- [ ] 前端统一 request：写请求自动加 `X-CSRF-Token`
- [ ] 后端统一校验中间件/基类方法：所有写接口强制校验 CSRF
- [ ] 处理登录/登出时 token 轮换（可选）

### 5.3 B1（JWT/自建 session）实施清单

- [ ] 定义 token 策略（access/refresh、过期、吊销）
- [ ] WP 端实现 Authorization 解析并映射为 WP user
- [ ] 将所有 `permission_callback` 从 `is_user_logged_in` 迁移为 token 校验
- [ ] 前端封装 token 刷新
- [ ] 全面回归测试

---

## 6. 你可以在这里记录你最终的选择

- 计划采用：
  - [ ] A（动态 nonce）
  - [ ] B2（double submit CSRF）
  - [ ] B1（JWT/自建 session）

- 优先级：
  - [ ] 先确保“nonce 动态化”
  - [ ] 再做统计 consent
  - [ ] 再优化安全策略（限流/CSP/登录防爆破）

---

## 7. 备注：与本项目相关的文件/线索（供你回看）

- Nuxt：
  - `nuxt-i18n/app/composables/useAuth.ts`（`credentials: include`、`X-WP-Nonce` 注入）
  - `nuxt-i18n/app/composables/useWarrantyCheck.ts`（调用 `/wp-json/tanzanite/v1/warranty/...`）
- WP 插件：
  - `wp-plugin/tanzanite-setting/includes/rest-api/class-rest-chat-controller.php`（`wp_set_auth_cookie`）
  - `wp-plugin/tanzanite-product-registry/includes/rest-api/class-rest-warranty-controller.php`（`permission_callback => is_user_logged_in` + 限流示例）

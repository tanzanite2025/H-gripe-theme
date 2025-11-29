# Tanzanite Subscription & Notification System Design

> 目标：在不依赖第三方邮件服务的前提下，完全用 WordPress（含自建商城 tanzanite-setting）+ Nuxt 前端，实现「用户留下邮箱 → 新博客/新品上架自动邮件通知」的闭环。

---

## 1. 总体架构概览

- **前端（Nuxt）**
  - 页面上提供一个订阅区域（类似 `OPT-IN`）：一个邮箱输入框 + `SUBSCRIBE` 按钮。
  - 建议封装为一个可复用的 Vue 组件（例如 `SubscriptionOptIn.vue`），在首页、博客详情页、商品页等需要的地方直接引用。
  - 组件内部通过 `fetch` / `axios` 调用 WordPress 暴露的 REST API，将邮箱发送给后台。

- **后端（WordPress + 独立订阅插件）**
  - 使用一个独立插件（推荐命名为 `tanzanite-subscription`）承载所有订阅逻辑，避免和 `tanzanite-setting` 商城逻辑混在一起，便于维护和单独启停。
  - 在数据库中维护一张「订阅者表」。
  - 【当前状态：独立订阅插件及其数据表、REST API、通知和后台设置均已按本设计实现。】
  - 暴露 REST 路由：
    - `POST /wp-json/tanz/v1/subscribe`：处理用户提交邮箱。
    - `GET  /wp-json/tanz/v1/subscribe/confirm`：用户点击邮件中的确认链接。
    - `GET  /wp-json/tanz/v1/unsubscribe`：用户点击退订链接。

- **触发通知**
  - 新博客发布：监听 `publish_post` 钩子，向所有已确认且未退订的邮箱发送「新文章」通知。
  - 新商品上架：
    - 如果商品是自定义 Post Type：监听 `publish_{post_type}`。
    - 如果商品是自定义表 + 业务逻辑：在商品「上架」的业务代码里触发自定义 Action（`do_action('tanz_new_product_published', $product_id)`），订阅插件监听这个 Action 并发送「新品」邮件。

---

## 2. 数据结构设计：订阅者表（已实现）

> 在插件激活时创建一张独立的订阅表，避免乱写入用户表/评论表。

### 2.1 表名建议

- `wp_tanz_subscribers`（前缀部分以实际安装为准，通常用 `$wpdb->prefix . 'tanz_subscribers'`）。

### 2.2 字段设计

- `id` (bigint, primary key, auto_increment)
- `email` (varchar(255), unique, NOT NULL)
- `created_at` (datetime, NOT NULL)
- `confirmed` (tinyint(1), 默认 0)
  - 0 = 未确认（只提交了邮箱但没点邮件里的确认链接）
  - 1 = 已确认
- `unsubscribed` (tinyint(1), 默认 0)
  - 0 = 正常订阅状态
  - 1 = 已退订
- `confirm_token` (varchar(64), 可空)
  - 发送确认邮件时生成的唯一 token，用于链接 `.../subscribe/confirm?token=xxx`
- `unsubscribe_token` (varchar(64), 可空)
  - 每个订阅者一份，用于退订链接 `.../unsubscribe?token=yyy`

> 备注：也可以额外存 `last_notified_post_id` / `last_notified_product_id` 等更细粒度的字段，但**第一版可以先不做**，直接「新内容一律通知所有订阅者」。后续如果需要减少打扰，可以再扩展表结构和逻辑。

### 2.3 表创建时机

在订阅插件主文件中：

- 使用 `register_activation_hook()` 在插件激活时运行一个函数：
  - 检查表是否存在；
  - 不存在则用 `dbDelta()` 创建。

> 这一部分将来可以写成：`tanz_subscription_install()` 之类的函数，单独维护。

---

## 3. REST API：订阅、确认、退订（已实现）

### 3.1 命名空间和路由

统一使用命名空间：`tanz/v1`。

- `POST /wp-json/tanz/v1/subscribe`
  - Body 参数：`email`（必填）
- `GET  /wp-json/tanz/v1/subscribe/confirm`
  - Query 参数：`token`（必填）
- `GET  /wp-json/tanz/v1/unsubscribe`
  - Query 参数：`token`（必填）

在插件中通过 `register_rest_route()` 注册以上路由，权限一般为 `permission_callback => '__return_true'`，但内部要做严格的参数校验和速率限制。

### 3.2 订阅接口：POST /subscribe

**基本流程**：

1. 从请求中读取 `email`，用 `sanitize_email()` 清洗，再用 `is_email()` 校验格式。
2. 如果邮箱无效：返回 400 + 错误消息。
3. 在 `wp_tanz_subscribers` 中查找是否已存在该 email：
   - 如果存在并且 `unsubscribed = 0`：
     - 可以直接返回「已订阅」提示（不要报错），也可以视为「重新发送确认邮件」（视你策略而定）。
   - 如果存在并且 `unsubscribed = 1`：
     - 认为是「重新订阅」，更新 `unsubscribed = 0`。
4. 生成 `confirm_token` 和 `unsubscribe_token`：
   - 可以用 `wp_generate_password(32, false)` 或 `bin2hex(random_bytes(16))`。
5. 插入或更新记录：
   - `confirmed = 0`，`created_at = current_time('mysql')`。
6. 发送一封确认邮件给该邮箱：
   - 标题：`Please confirm your Tanzanite subscription`
   - 内容包含一个确认链接，例如：
     - `https://your-site.com/wp-json/tanz/v1/subscribe/confirm?token=CONFIRM_TOKEN`
7. 返回 JSON：
   - 成功：`{ success: true, message: 'Please check your email to confirm subscription.' }`
   - 失败：`{ success: false, message: '...' }`

> 考虑到隐私和反垃圾，**推荐保留双重确认**。

### 3.3 确认接口：GET /subscribe/confirm

**流程**：

1. 从 `$_GET['token']` 中获取 token，并清洗。
2. 在表中查找 `confirm_token = ?` 且 `unsubscribed = 0` 的记录：
   - 找不到：
     - 返回「链接无效或已使用」的 JSON 或简单 HTML。
   - 找到：
     - 设置 `confirmed = 1`；
     - 可以同时清空 `confirm_token`（可选）。
3. 响应方式：
   - 方式 A：直接输出一段简单 HTML：「订阅成功」+ 返回到你 Nuxt 网站的按钮；
   - 方式 B：返回 JSON 并带 302 重定向到某个前端页面，如 `/subscription/confirmed`。

### 3.4 退订接口：GET /unsubscribe

**流程**：

1. 从 `$_GET['token']` 获取 token。
2. 在表中查找 `unsubscribe_token = ?` 的记录：
   - 找不到：提示「退订链接无效」。
   - 找到：
     - 设置 `unsubscribed = 1`，可选地把 `confirmed` 也置为 0。
3. 输出结果：
   - 简单 HTML 提示「你已成功退订」，或跳转到一个 Nuxt 的「退订成功」页面。

**邮件里的退订链接格式**：

- `https://your-site.com/wp-json/tanz/v1/unsubscribe?token=UNSUB_TOKEN`

---

## 4. 内容更新 → 自动发送邮件（已实现）

### 4.1 发信渠道与第一版策略

- **第一版不强制依赖任何 SMTP 插件或第三方发信服务**：
  - 直接使用 WordPress 的 `wp_mail()` + 服务器默认的 mail 机制；
  - 发件人信息统一由订阅插件设置页的 `From Email / From Name` 控制。
- **发送时机完全由你掌控**：
  - 是否在「新博客发布」「新商品上架」时自动发邮件，由后台开关 `tanz_sub_auto_notify_posts` / `tanz_sub_auto_notify_products` 决定；
  - 也可以关闭所有自动通知，只在「Subscription Broadcasts」页面手动选择某篇文章/商品进行群发，实现「想发就发，不想发就不发」。
- **后续可选优化（非当前需求）**：
  - 如果未来发现送达率不足或需要更细的统计，再考虑接入 SMTP 插件或外部邮件 API（SendGrid、Amazon SES 等），但订阅表结构和通知逻辑本身可以保持不变，仅替换底层发信实现。

> 设计决策：当前版本**不计划**接入任何 SMTP 插件或外部邮件服务，仅在未来实际出现明显送达问题或新需求时，再单独评估是否引入。

### 4.2 新博客发布通知（实际实现）

- 在订阅插件中由 `TSUB_Notifications::handle_new_post()` 通过 `publish_post` 钩子触发。
- 逻辑要点：
  1. 仅处理标准文章 `post` 类型。
  2. 先检查后台开关 `tanz_sub_auto_notify_posts`：
     - 为 `1` 时才发送通知；
     - 为 `0` 时直接返回，方便你暂时关闭自动通知、只用手动群发。
  3. 读取所有订阅者：
     - 查询 `tanz_subscribers` 表中 `confirmed = 1 AND unsubscribed = 0` 的记录。
  4. 构造邮件内容：
     - 标题：`New blog post on Tanzanite: {post_title}`；
     - 正文：文章内容经 `wp_strip_all_tags()` 清洗后，用 `wp_trim_words()` 截取简介；
     - 附上文章链接 `get_permalink( $post_id )` 和退订链接。
  5. 逐个发送：
     - 调用统一的 `tanz_sub_send_mail()`，自动带上发件人信息和尾部备注。

> 如后期担心一次性发送量过大，可以按 3.x 版设计延伸：用 `wp_schedule_single_event()` 将真正发送逻辑放到 WP Cron 里异步执行。

### 4.3 新商品发布通知（与 tanzanite-setting 的实际集成）

当前商城商品使用自定义文章类型 `tanz_product`，由 `tanzanite-setting` 插件管理。为了与订阅系统解耦，实际实现分两层：

1. **在 tanzanite-setting 里发出「商品已上架」事件**
   - 新增 `includes/subscription-hooks.php`，监听商品状态变化：
     - 使用 `transition_post_status` 钩子，捕获 `tanz_product` 从非 `publish` 变为 `publish` 的时刻；
     - 在该时刻调用：
       - `do_action( 'tanz_new_product_published', $product_id );`
   - 这样无论是单个商品发布还是批量改状态，只要进入 `publish`，订阅插件都会收到统一事件。

2. **在订阅插件中根据事件发送「新品」通知**
   - `TSUB_Notifications::handle_new_product( $product_id )` 监听 `tanz_new_product_published`：
     1. 先检查后台开关 `tanz_sub_auto_notify_products`：
        - 为 `1` 时才发送通知；
        - 为 `0` 时直接返回。
     2. 通过 `apply_filters( 'tanz_sub_product_email_data', array(), $product_id )` 向 `tanzanite-setting` 请求邮件所需数据：
        - `title`：商品标题；
        - `url`：商品详情页链接；
        - `excerpt`：商品简介（优先摘要，其次从内容截取）。
     3. 构造标题：`New product on Tanzanite: {title}`，正文包含简介、详情页链接和退订说明。
     4. 调用 `broadcast_to_all_subscribers()` 对所有已确认且未退订的邮箱群发，内部统一附带退订链接和邮件尾部备注。

> 这种「tanzanite-setting 只负责发出事件 + 提供商品数据，tanzanite-subscription 负责监听与发信」的方式，保证了商城核心逻辑与订阅逻辑彻底解耦。

---

## 5. Nuxt 前端订阅表单集成（已实现）

> 目标：前端只关注样式和调用 API，不参与邮件发送和数据存储。

### 5.1 表单布局示例

- 结构：
  - 一个 `input[type=email]`。
  - 一个按钮 `SUBSCRIBE`。
  - 一个小区域显示成功/失败提示。

- 行为：
  1. 用户输入邮箱后点击按钮；
  2. 前端简单校验（非空 + 基本邮箱正则）；
  3. 通过 `fetch` / `axios` 调用 `POST /wp-json/tanz/v1/subscribe`；
  4. 根据返回的 JSON 里的 `success` / `message`，在页面展示提示信息；
  5. 不刷新整页，不暴露后台 URL 细节。

### 5.2 错误处理 & UX 建议

- 提交期间禁用按钮，避免重复点击；
- 成功后清空输入框，显示「请前往邮箱点击确认链接」；
- 如果接口返回邮箱已存在且已订阅：
  - 可以统一提示为「你已经订阅过了，如未收到邮件请检查垃圾箱」；
- 接口返回 500 或网络错误时，提示用户稍后重试。

> 实现情况：`SubscriptionOptIn.vue` 组件已按上述行为实现，并通过 `AppFooter.vue` 嵌入到所有使用全局页脚的 Nuxt 页面中。

---

## 6. 安全与合规注意事项

1. **防止滥用**
   - 对订阅接口增加简单的频率限制（例如短时间内重复提交同一 IP / 同一邮箱时直接返回成功但不真正重新发送邮件）。
   - 可选方案：在前端表单增加一个简单的随机验证码（例如 4 位数字或简单加减法），在 REST API 中一并提交并校验，用来拦截脚本批量请求。可以在系统上线初期先不开启，只依赖频率限制和双重确认；一旦订阅量增多或出现恶意脚本，再开启验证码。
   - 也可以考虑使用第三方的基础人机验证（如 reCAPTCHA 等），由前端完成验证并在后端做 token 校验，适合访问量更大、风控要求更高的阶段。
   - 【当前状态：目前版本尚未实现接口频率限制/验证码，仅依赖双重确认和一键退订；如后续出现滥用再按本节方案扩展。】

2. **隐私与法律**
   - 邮件里必须提供清晰可见的退订链接，且一键退订生效。
   - 在订阅区域旁边简单标注将用于接收新品/博客更新，不会乱用邮箱。

3. **发信信誉**
   - 建议使用独立的发信邮箱（如 `no-reply@your-domain.com`），并在 DNS 中配置 SPF/DKIM。
   - 如有需要，可以后期考虑把「发送邮件」部分切换到专门的邮件 API（SendGrid、Amazon SES 等），但逻辑上仍复用这套订阅表和钩子。

---

## 7. 订阅插件后台设置页（完全自建，不依赖 SMTP 插件）【已实现】

> 目标：在不安装 SMTP 插件的前提下，为订阅系统提供一个简单、独立的后台设置界面，用来配置发件邮箱和发件人名称，并在所有订阅相关邮件中统一使用。

### 7.1 菜单位置与页面结构

- 在 `tanzanite-subscription` 插件中注册一个 options page，例如：
  - 顶层菜单「Tanzanite」，子菜单「Subscription Settings」，或
  - 直接挂在「设置 > Tanzanite Subscription」。
- 页面上只做最小必要字段，保持简洁：
  - 发件人邮箱（From Email）
  - 发件人名称（From Name）

### 7.2 存储方式（wp_options）

- 使用 `get_option()` / `update_option()` 存取配置，例如：
  - `tanz_sub_from_email`
  - `tanz_sub_from_name`
- 订阅插件所有发送邮件的地方，统一从上述 option 中读取配置：
  - 如果用户未在设置页中填写，则可以回退到 WordPress 默认 `admin_email` 或服务器默认发件配置。

### 7.3 表单与校验

- 后台设置表单提交时：
  - 对发件邮箱做 `sanitize_email()` + `is_email()` 校验；
  - 发件人名称用 `sanitize_text_field()` 清洗；
  - 保存前使用 `check_admin_referer()` 防止 CSRF。
- UI 侧给出简单说明：
  - 「发件邮箱建议使用与你域名一致的邮箱，如 `no-reply@your-domain.com`，并确保已在服务器/DNS 中完成 SPF/DKIM 配置。」

### 7.4 在 wp_mail() 中应用设置

- 订阅插件内部统一通过 `tanz_sub_send_mail( $to, $subject, $message )` 发送邮件：
  - 先通过 `get_option()` 读取 `tanz_sub_from_email`、`tanz_sub_from_name`；
  - 构造 `From: {From Name} <{From Email}>` 头部并传给 `wp_mail()`；
  - 这样即使没有 SMTP 插件，仍然能通过这套设置页面控制所有订阅相关邮件的发件人信息；未来如果迁移到 SMTP，只需在服务器或外部服务里配置同一个发件邮箱即可。

### 7.5 邮件尾部备注

- 设置项：`tanz_sub_footer_note`，在后台表单中以多行文本框的形式出现。
- 用途：
  - 在所有通过 `tanz_sub_send_mail()` 发送的邮件正文末尾自动追加一段备注；
  - 适合放品牌签名、联系方式、版权或合规性说明。
- 生效范围：
  - 订阅确认邮件；
  - 新博客 / 新商品自动通知邮件；
  - 手动群发（Subscription Broadcasts 页面）发送的广播邮件。

----

## 8. 通知策略与推送控制（手动群发 + 开关自动发）【已实现】

> 目标：避免每次内容更新都「强制推送」打扰用户，同时又保留「需要时可以一键群发」的能力，让站点运营可以灵活控制通知节奏。

### 8.1 自动通知开关

- 在订阅插件的后台设置页中增加几个布尔开关：
  - 「新博客发布时自动发送通知邮件」
  - 「新商品上架时自动发送通知邮件」
- 存储到 `wp_options` 中，例如：
  - `tanz_sub_auto_notify_posts`（0/1）
  - `tanz_sub_auto_notify_products`（0/1）
- 在对应的钩子处理函数中（`tanz_notify_new_post()` / `tanz_notify_new_product_by_id()`）：
  - 先检查相关开关是否为开启状态；
  - 关闭时直接 return，不发送任何邮件；
  - 这样运营可以在某个阶段暂停自动通知，改为只用手动群发。

### 8.2 手动群发界面（选文章 / 商品 → 发送）

- 在订阅插件中增加一个「手动群发」管理页，例如菜单：
  - 「Tanzanite > Subscription Broadcasts」。
- 页面功能（第一版已实现部分）：
  - 选择群发类型：
    - 根据文章 ID / 文章列表选择一篇博客。
    - 或根据商品 ID / 商品列表选择一个商品。
  - 自动预填部分内容：
    - 默认标题可以基于文章/商品标题生成，运营可以手动修改。
    - 默认正文可以包含摘要 + 链接 + 退订说明。
  - 在表单上方显示一个「订阅者总览」信息块：
    - 当前活跃订阅数（confirmed = 1 且 unsubscribed = 0，将会收到本次群发）；
    - 待确认数量（confirmed = 0 且 unsubscribed = 0）；
    - 已退订数量（unsubscribed = 1）；
    - 总共收集到的邮箱数量（所有状态）。
  - 在同一页面展示「订阅者列表」表格：
    - 列包含：邮箱地址、订阅时间、状态（已确认/待确认/已退订）。
    - 顶部提供状态筛选按钮：
      - 全部；
      - 仅活跃订阅者（confirmed = 1 且 unsubscribed = 0）；
      - 仅待确认（confirmed = 0 且 unsubscribed = 0）；
      - 仅已退订（unsubscribed = 1）。
    - 每一行提供复选框，可勾选一批邮箱作为本次目标人群。
  - 点击「发送」按钮后：
    - 如页面中勾选了一个或多个邮箱：仅对这些邮箱发送邮件（内部仍会二次过滤，只给当前依然处于活跃状态的订阅者发信）。
    - 如未勾选任何邮箱：对所有活跃订阅者（confirmed = 1 且 unsubscribed = 0）发送邮件。
  - 导出 CSV：
    - 在列表上方提供「导出勾选邮箱为 CSV」按钮；
    - 如勾选了邮箱：仅导出这些邮箱及其订阅时间、状态；
    - 如未勾选任何邮箱：导出当前筛选条件下的所有订阅者（例如仅活跃、仅待确认等）。
- 可选扩展（当前未实现，仅作为后续优化方向）：
  - 记录一次「广播日志」，包括时间、类型（博客/商品）、目标 ID、已发送数量等。

### 8.3 与自动通知的配合方式

- 推荐策略：
  - 日常产品/博客更新时，可以关闭自动通知，按需选择重要内容进行手动群发，避免用户频繁收到邮件产生反感。
  - 对特别重要的更新（如大版本发布、新品首发），可以临时开启对应的自动通知开关，或使用手动群发页面专门发一封策划好的邮件。
- 通过同时提供「自动开关 + 手动群发」两种机制，你可以完全掌控：
  - 哪些内容要推送；
  - 什么时候推送；
  - 频率多高，避免对订阅者造成打扰。

----

## 9. 后续可能的扩展点（可选，不影响第一版）

- 按兴趣标签分组订阅（只收新品、不收博客，或反之）。
- 针对特定商品系列或博客分类的订阅列表。
- 记录「最后一次发送的文章/商品 ID」，避免重复推送相同内容。
- 后台管理页面：
  - 订阅者列表 + 状态筛选 + 勾选导出/群发（部分已实现：已在 Subscription Broadcasts 页面内集成，只读且无单条编辑功能）；
  - 手动导出 CSV（部分已实现：可通过 Subscription Broadcasts 页面导出当前筛选结果或仅导出勾选邮箱）；
  - 手动添加/删除订阅者（尚未实现，如未来需要更细粒度运营可单独扩展独立管理页）。

----

## 10. 实施顺序建议

1. 在一个独立插件或 `tanzanite-setting` 里：（已完成，见第 2 节）
   - 写好订阅表的安装代码（activation hook）。
2. 实现 REST API：`subscribe / confirm / unsubscribe`，用 Postman 或浏览器先测通。（已完成，见第 3 节）
3. 在 Nuxt 里接入订阅表单 UI，连上 `POST /subscribe`，验证前后端联通。（已完成，见第 5 节）
4. 加上博客发布通知 `publish_post` 的钩子，先用测试邮箱观察是否正常发出。（已完成，见第 4.2 节）
5. 在 tanzanite-setting 里确定商品上架的关键点，加上 `do_action('tanz_new_product_published', $product_id)`，并实现监听和邮件发送。（已完成，见第 4.3 节）
6. 视需要逐步优化邮件模板、频率控制和后台管理界面。（部分已完成：已增加邮件尾部备注和手动群发页面；频率控制等仍作为后续优化）

# PHP API → Go API 迁移矩阵

## 范围

本文件只做 D1b 盘点和下一步拆分，不实现迁移。

更新：根目录 WordPress 主题壳文件和主题元数据（`index.php`、`header.php`、`footer.php`、`page.php`、`single.php`、`functions.php`、`style.css`）已确认属于早期虚拟机限制下的历史文件并已删除。下表中来自 `removed root functions.php` 的 endpoint 只表示历史参考或 Nuxt 仍可能存在的旧调用目标，不能作为可恢复的 PHP 实现。

目标：

- 找出仍在 PHP/WordPress 暴露或 Nuxt 调用的旧 API。
- 标出可直接切到 Go 的接口、需要补 Go 能力的接口、以及只应作为历史参考的接口。
- 给后续每个单模块 PR 明确边界。

## 状态标记

- `Go 已有`：Go 端已有相近路由，后续模块 PR 主要做响应契约核对和 Nuxt 切流。
- `Go 部分已有`：Go 端有领域模型或部分路由，但旧 PHP 行为/响应结构尚未完全覆盖。
- `Go 缺口`：Go 端没有对应业务 API，必须先补 Go 模块，不能直接删 PHP。
- `历史后台`：WordPress admin/admin-post/ajax 能力，只用于旧后台参考；目标是迁到 `go-backend/web/admin`。

## Nuxt 仍在使用的旧入口

| 旧入口 | 主要调用文件 | 当前判断 | 建议模块 |
| --- | --- | --- | --- |
| `runtimeConfig.public.wpApiBase` 默认 `/wp-json` | `nuxt-i18n/nuxt.config.ts` | 全局 WordPress 兼容入口仍存在 | 每个业务模块切完后逐步移除 |
| `/wp-json/tanzanite/v1/settings`、`/settings/quick-buy` | `app/layouts/default.vue`, `app/composables/useSiteTitle.ts`, `app/composables/useSocialLinks.ts` | 旧根目录 PHP 来源已删除；Go 已有 `/api/v1/settings/site`、`/api/v1/settings/quick-buy`，但旧 `/settings` 与 Go `/settings/site` 名称不一致 | M1.1 settings/quick-buy |
| `/wp-json/tanzanite/v1/products`、`/product-categories`、`/attributes/filterable` | `app/pages/shop.vue`, `app/pages/shop/[slug].vue`, `app/components/CategoryProductsStrip.vue`, `app/composables/useShopCategories.ts`, `app/composables/useProductAttributes.ts` | 产品列表 Go 已有；分类/属性存在缺口或契约需核对 | M3 product catalog |
| `/wp-json/tanzanite/v1/seo/product/*` | `app/pages/shop/[slug].vue` | Go 只有通用 settings SEO；产品 SEO/schema 缺口 | M3 product SEO |
| `/wp-json/tanzanite/v1/wishlist` | `app/composables/useWishlist.ts` | M2 已切到 Go `/api/v1/wishlist` | M2 wishlist |
| `/wp-json/tanzanite/v1/warranty/*` | `app/components/warranty/SubmitClaimTab.vue`, `app/composables/useWarrantyCheck.ts` | M2 已切到 Go `/api/v1/registrations/warranty/*`；底层 registrations/warranty-claims 可继续复用 | M2 warranty/registration |
| `/wp-json/tanzanite/v1/suggestion-feedback`、`/feedback` | `app/composables/useSuggestionFeedback.ts`, `app/composables/useFeedback.ts` | M2 已切到 Go `/api/v1/suggestion-feedback`、`/api/v1/feedback`；后台审核接口另迁 | M2 feedback |
| `/wp-json/tanzanite/v1/spoke-history`、`/spoke-db-export` | `app/composables/useSpokeHistory.ts`, `scripts/sync-spoke-data.mjs` | Go 缺口 | M2 spoke |
| `/wp-json/tanzanite/v1/shipping-templates`、`/tax-rates`、`/packaging-rules` | `app/composables/useCartCalculation.ts`, `app/composables/useShippingValidation.ts`, `app/composables/usePackagingCalculation.ts` | Shipping/tax Go 已有但路径不同； packaging 缺口 | M3 cart calculation |
| `/wp-json/tanzanite/v1/coupons/validate`、`/loyalty/points` | `app/composables/useCartCalculation.ts` | Go 已有 marketing routes，路径和 auth 需切换 | M3 coupon/loyalty |
| `/wp-json/tanzanite/v1/customer-service/**`、`/auto-reply/**`、`/agent/**` | `app/components/WhatsAppChatModal.vue`, `app/composables/useChat.ts` | Go ticket 模块不能等同实时客服；customer-service/auto-reply/agent 是 Go 缺口 | M2/M3 customer service |
| `/wp-json/mytheme-vue/v1/my-orders` | `app/components/WhatsAppChatModal.vue` | Go 有 `/api/v1/orders`，但旧 WooCommerce 响应需映射 | M3 orders |
| `/tanz/v1/subscribe` | `app/components/SubscriptionOptIn.vue` | Go 已有 `/api/v1/subscriptions` | M1.2 subscription |

## PHP endpoint 分组矩阵

### 1. 认证/session

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `POST /wp-json/custom/v1/login` | removed root `functions.php` | `POST /api/v1/auth/login` | Go 已有 | 前端 `useAuth` 已直连 Go；清理错误文案和旧 alias |
| `POST /wp-json/custom/v1/logout` | removed root `functions.php` | `POST /api/v1/auth/logout` | Go 已有 | 后续删除 PHP auth route 前确认 cookie/JWT 不混用 |
| `POST /wp-json/custom/v1/register` | removed root `functions.php` | `POST /api/v1/auth/register` | Go 已有 | 与用户数据迁移一起收尾 |
| `POST /wp-json/tanzanite/v1/auth/*`、`/chat/*` auth alias | `wp-plugin/tanzanite-setting/includes/rest-api/class-rest-auth-controller.php` | `/api/v1/auth/*` | Go 已有 | 保持兼容直到所有 Nuxt 调用移除 |

### 2. Settings / SEO / 站点公共配置

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `GET /wp-json/tanzanite/v1/settings/site` | removed root `functions.php` | `GET /api/v1/settings/site` | Go 已有 | M1.1 切 Nuxt 到 Go |
| `GET /wp-json/tanzanite/v1/settings/quick-buy` | removed root `functions.php` | `GET /api/v1/settings/quick-buy` | Go 已有 | M1.1 同步 QuickBuy 配置字段 |
| `GET/POST /wp-json/tanzanite/v1/seo/settings` | `wp-plugin/tanzanite-setting/includes/class-mytheme-seo.php` | `GET /api/v1/settings/seo`, admin `/api/admin/settings/seo` | Go 部分已有 | 分离 public SEO 与后台 SEO 管理 |
| `GET/POST /wp-json/tanzanite/v1/seo/homepage` | `class-mytheme-seo.php` | `GET /api/v1/settings/seo` 或新增 content SEO | Go 部分已有 | 与产品 SEO 分开做 |
| `GET /wp-json/tanzanite/v1/seo/product/*`、`/seo/schema/product/*` | `class-mytheme-seo.php` | 待定：`/api/v1/products/:id/seo` 或 `/api/v1/content/seo/product/:id` | Go 缺口 | 放入 Product SEO 单独 PR |
| `GET/POST /wp-json/tanzanite/v1/seo/languages` | `class-mytheme-seo.php` | `GET /api/v1/i18n/languages` | Go 部分已有 | 只读语言先切；导入功能走 admin |

### 3. Content / Blog / FAQ / Gallery

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `GET /wp-json/tanzanite/v1/posts`、`/post`、`/translations` | `wp-plugin/tanzanite-blog-i18n/includes/class-blog-rest.php` | Go WP 兼容层 `/wp-json/tanzanite/v1/*`，最终 `/api/v1/content/posts` | Go 部分已有 | M1.3 已固定 Nuxt blog 走 Go WP 兼容层；直连 `/api/v1/content/posts` 等 Go content 契约补齐后再切 |
| FAQ admin-post actions | `wp-plugin/tanzanite-faq-content/includes/class-faq-editor.php` | `/api/v1/content/faqs`、`/api/admin/content/faqs` | Go 已有 | Nuxt FAQ 当前已尝试 Go；后台迁到 `web/admin` |
| `tanz-photo/v1/**` | `wp-plugin/tanzanite-photo-gallery/includes/class-tpg-rest.php` | `/api/v1/gallery`、`/api/admin/content/galleries` | Go 已有 | M1 gallery：核对图片字段和批量删除 |
| `GET /wp-json/mytheme-vue/v1/menu/:location` | removed root `functions.php` | 待定：`/api/v1/settings/public` 或新增 `/api/v1/content/menus/:location` | Go 缺口 | 与站点导航/菜单单独 PR |
| `GET /wp-json/mytheme-vue/v1/search` | removed root `functions.php` | 待定：跨 content/product search | Go 缺口 | 不并入 blog；另建 search 模块 |

M1.3 固定的 Blog 读取契约：

- Nuxt `useBlogApi` 使用 Go WP 兼容层，而不是 WordPress PHP 或直连 `/api/v1/content/posts`。
- `GET /wp-json/tanzanite/v1/posts?lang=&category=&page=&per_page=` 返回 `{ page, per_page, total, items }`，`items[]` 使用 `id/lang/group/slug/title/excerpt/date/featuredImage/categories/translations`。
- `GET /wp-json/tanzanite/v1/post?lang=&slug=` 返回列表字段加 `contentHtml`、`canonicalUrl`。
- `GET /wp-json/tanzanite/v1/translations?group=` 返回 `{ group, translations }`，`translations` 是语言码到 `{ id, slug }` 的映射。
- 暂不切到 `/api/v1/content/posts`：当前直接 Go content 响应缺少 Nuxt blog 需要的 `categories`、`contentHtml`、`translations` 兼容形状。

### 4. Product catalog / 商品后台

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `GET/POST/PUT /wp-json/tanzanite/v1/products` | `class-rest-products-controller.php` | public `GET /api/v1/products`; admin `/api/admin/products` | Go 部分已有 | M3 product catalog：先只读列表/详情 |
| `GET/PUT/DELETE /wp-json/tanzanite/v1/products/:id` | `class-rest-products-controller.php` | `GET /api/v1/products/:id`; admin `/api/admin/products/:id` | Go 部分已有 | 后台编辑与前台读取分开 PR |
| `GET /wp-json/tanzanite/v1/product-categories` | Nuxt 预留调用 | 待定：`/api/v1/products/categories` | Go 缺口 | Product category 子模块 |
| `/wp-json/tanzanite/v1/attributes/**` | `class-rest-attributes-controller.php` | 待定：`/api/v1/products/attributes` | Go 缺口 | Product attributes 子模块 |
| `wp_ajax_tanzanite_pr_* product/type/import/export` | `tanzanite-product-registry/includes/admin/**` | `/api/admin/registrations` 或 `/api/admin/products` | 历史后台 | 只迁 `web/admin` 需要的能力 |

### 5. Cart / Order / Checkout calculation

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `GET /wp-json/mytheme-vue/v1/cart-summary` | removed root `functions.php` | `GET /api/v1/cart/summary` | Go 已有 | M3 cart：响应字段核对 |
| Woo Store API via `storeApiBase` | `QuickBuy.vue` | `/api/v1/cart/add`、`/api/v1/orders` | Go 部分已有 | QuickBuy 不能和 settings PR 混做 |
| `GET/POST/PUT /wp-json/tanzanite/v1/orders` | `class-rest-orders-controller.php` | `/api/v1/orders`; admin `/api/admin/orders` | Go 已有 | M3 orders：订单字段/状态映射 |
| `GET/PUT/DELETE /wp-json/tanzanite/v1/orders/:id` | `class-rest-orders-controller.php` | `/api/v1/orders/:id`; admin `/api/admin/orders/:id` | Go 已有 | 同 orders 模块 |
| `POST /wp-json/tanzanite/v1/orders/:id/tracking` | `class-rest-orders-controller.php` | `PATCH /api/admin/orders/:id/tracking` 或 `/api/v1/shipping/orders/:id/tracking` | Go 部分已有 | 与 shipping tracking 一起做 |
| `GET /wp-json/mytheme-vue/v1/my-orders` | removed root `functions.php` | `GET /api/v1/orders` | Go 已有 | 旧 WooCommerce order shape 需适配 |

### 6. Marketing / Loyalty / Gift card

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `/wp-json/tanzanite/v1/coupons`、`/coupons/:id` | `class-rest-coupons-controller.php` | public `/api/v1/marketing/coupons`; admin `/api/admin/marketing/coupons` | Go 已有 | M3 coupon admin/read |
| `POST /wp-json/tanzanite/v1/coupons/validate` | `class-rest-coupons-controller.php` | `POST /api/v1/marketing/coupons/validate` | Go 已有 | Cart calculation 子模块 |
| `POST /wp-json/tanzanite/v1/coupons/apply` | `class-rest-coupons-controller.php` | 建议并入 order creation 或 coupon validation result | Go 部分已有 | 不单独保留旧 apply 语义 |
| `GET /wp-json/tanzanite/v1/coupons/my` | `class-rest-coupons-controller.php` | 待定：`/api/v1/marketing/coupons/my` | Go 缺口 | 用户资产模块 |
| `/wp-json/tanzanite/v1/giftcards/**` | `class-rest-giftcards-controller.php` | `/api/v1/marketing/gift-cards*` | Go 部分已有 | 礼品卡单独 PR |
| `/wp-json/tanzanite/v1/loyalty/checkin|points|referral|config|assets` | `class-rest-loyalty-controller.php` | `/api/v1/marketing/loyalty/*` | Go 部分已有 | 用户积分资产 PR |
| `GET /wp-json/tanzanite/v1/redeem/config`、`POST /redeem/apply` | `class-rest-redeem-controller.php` | 待定：marketing redeem config/use | Go 缺口 | 不并入 coupon validate |

### 7. Payment / Shipping / Tax / Packaging

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `/wp-json/tanzanite/v1/payment-methods` | `class-rest-payments-controller.php` | `/api/v1/payment/methods`; admin payment methods | Go 已有 | M3 payment settings |
| `/wp-json/tanzanite/v1/tax-rates` | `class-rest-taxrates-controller.php` | `/api/v1/payment/tax-rates`; admin tax-rates | Go 已有 | Cart calculation 子模块 |
| `/wp-json/tanzanite/v1/shipping-templates` | `class-rest-shippingtemplates-controller.php` | `/api/v1/shipping/templates` | Go 已有 | 路径切换和字段映射 |
| `/wp-json/tanzanite/v1/carriers` | `class-rest-carriers-controller.php` | `/api/v1/shipping/carriers` | Go 已有 | Shipping 子模块 |
| `/wp-json/tanzanite/v1/packaging-rules` | `class-rest-packaging-controller.php` | 待定：`/api/v1/shipping/packaging-rules` | Go 缺口 | Packaging 独立 PR |

### 8. User interaction / support

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `/wp-json/tanzanite/v1/wishlist`、`/wishlist/:id` | `class-rest-wishlist-controller.php` | `/api/v1/wishlist`、`/api/v1/wishlist/:id` | Go 已有 | M2 wishlist 已补 Go API 并切 Nuxt；删除 PHP 前需验证数据迁移/代理 |
| `/wp-json/tanzanite/v1/reviews`、`/reviews/:id` | `class-rest-reviews-controller.php` | `/api/v1/reviews`、`/api/v1/reviews/:id`、`/api/v1/reviews/my`、`/api/v1/reviews/summary/:product_id` | Go 已有 | M2 review 已固定 Go 契约；Nuxt 当前无旧 review REST 调用 |
| `/wp-json/tanzanite/v1/feedback`、`/feedback/eligibility` | `class-rest-feedback-controller.php` | `/api/v1/feedback`、`/api/v1/feedback/eligibility` | Go 已有 | M2 feedback 已补 Go API 并切 Nuxt；后台 status 审核另行迁移 |
| `/wp-json/tanzanite/v1/suggestion-feedback/**` | `class-rest-suggestion-feedback-controller.php` | `/api/v1/suggestion-feedback`、`/api/v1/suggestion-feedback/eligibility` | Go 已有 | M2 suggestion feedback 已补前台提交/资格检查并切 Nuxt；后台列表/status 审核另行迁移 |
| `/wp-json/tanzanite/v1/customer-service/**` | `tanzanite-customer-service/**` | 不能直接等同 `/api/v1/tickets`; 需新增 customer-service 或重设计为 tickets | Go 缺口 | M2/M3 customer service |
| `/wp-json/tanzanite/v1/auto-reply/**` | `class-auto-reply-api.php` | 待定：customer-service auto-reply | Go 缺口 | 与 customer-service 同模块 |
| `/wp-json/tanzanite/v1/agent/**` | `class-agent-api.php` | 待定：`/api/admin/customer-service/agent` | Go 缺口 | `web/admin` 客服工作台模块 |
| removed root `functions.php` legacy `/chat/**` | removed root `functions.php` | 待定：customer-service 或 tickets | Go 缺口 | 不恢复旧 alias |

M2 review 固定的 Go 契约：

- Nuxt 当前没有调用旧 `/wp-json/tanzanite/v1/reviews*`；后续新增评价 UI 时直接走 Go `/api/v1/reviews*`。
- 公开读取：`GET /api/v1/reviews?product_id=&page=&page_size=` 返回 `{ data, pagination }`，默认只返回 approved 评价；`GET /api/v1/reviews/summary/:product_id` 返回 rating 聚合。
- 登录用户：`POST /api/v1/reviews` 使用 JWT，body 为 `{ product_id, rating, title, content, images }`，新评价进入 pending；`GET /api/v1/reviews/my` 读取当前用户评价；`DELETE /api/v1/reviews/:id` 只允许作者删除。
- Admin 审核仍走 `/api/v1/admin/reviews/*`，不混入前台评价提交 PR。

### 9. Product registration / Warranty / Spoke

| PHP endpoint | 来源 | Go 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `/wp-json/tanzanite/v1/warranty/:code` | `tanzanite-product-registry/includes/rest-api/class-rest-warranty-controller.php` | `GET /api/v1/registrations/warranty/:code` | Go 已有 | Nuxt warranty check 已切流 |
| `/wp-json/tanzanite/v1/warranty/verify-order` | Nuxt warranty form | `POST /api/v1/registrations/warranty/verify-order` | Go 已有 | 通过 Go orders 邮箱校验 |
| `/wp-json/tanzanite/v1/warranty/claim` | `class-rest-warranty-claims-controller.php` | `POST /api/v1/registrations/warranty/claim` | Go 已有 | 前台 multipart claim 已切流；后台 claims 审核沿用 registrations 模块 |
| `/wp-json/tanzanite/v1/spoke-db-export` | `class-rest-spoke-export-controller.php` | 待定：static generated data or `/api/v1/spoke/export` | Go 缺口 | Spoke data PR |
| `/wp-json/tanzanite/v1/spoke-history` | Nuxt spoke history | 待定：`/api/v1/spoke/history` | Go 缺口 | Spoke history PR |

### 10. Legacy admin/ajax/admin-post

| Legacy action | 来源 | Go/Web Admin 目标 | 状态 | 下一步 |
| --- | --- | --- | --- | --- |
| `admin_post_tsub_send_broadcast`、`admin_post_tsub_export_subscribers` | `tanzanite-subscription` | `/api/admin/subscriptions` export/broadcast | 历史后台 | 仅在 admin subscriptions PR 迁移 |
| `admin_post_tanz_save_tracking_settings`、`admin_post_tanz_test_tracking` | `tanzanite-setting/includes/class-plugin.php` | `/api/admin/settings` 或 shipping tracking test | 历史后台 | 不进入前台迁移 |
| `wp_ajax_tanzanite_pr_*` | `tanzanite-product-registry/includes/admin/**` | `/api/admin/registrations` | 历史后台 | Product registry admin PR |
| FAQ `admin_post_*` | `tanzanite-faq-content/includes/class-faq-editor.php` | `/api/admin/content/faqs` | 历史后台 | FAQ admin PR |

## 推荐下一批单模块 PR

不要一次性做完以下内容；每项单独 PR。

1. **M1.1 settings/quick-buy 切流**
   - 只改 Nuxt 对 `settings/site`、`settings/quick-buy`、site title/social links 的调用。
   - 目标 Go route：`/api/v1/settings/site`、`/api/v1/settings/quick-buy`。
   - 不碰产品、购物车、订单。
2. **M1.2 subscription 切流**
   - `SubscriptionOptIn.vue` 从 `/tanz/v1/subscribe` 切到 `/api/v1/subscriptions`。
   - 不处理后台导出/广播。
3. **M1.3 blog/content 切流**
   - 从 `/wp-json/tanzanite/v1/posts|post|translations` 转到 Go content 或确认是否继续短期使用 Go 的 WP 兼容层。
   - 不处理产品 SEO。
4. **M2 wishlist**
   - 先补 Go wishlist API，再切 `useWishlist.ts`。
   - 不混 review/feedback。
5. **M3 product catalog**
   - 先只做商品列表/详情，不做 category/attributes/SEO/order。

## 删除 PHP 的前置条件

某个 PHP endpoint 只能在对应模块满足以下条件后删除：

1. Nuxt 搜索不到对应 `/wp-json/**` 调用。
2. Go route 覆盖旧 endpoint 的成功/错误响应和鉴权语义。
3. 旧数据迁移或读取路径已验证。
4. 对应模块 PR 已合入并通过验证。
5. 线上代理不再把该 endpoint 转发到 WordPress。

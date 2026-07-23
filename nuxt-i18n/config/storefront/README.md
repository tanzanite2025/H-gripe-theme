# Storefront route policy and SSR HTML cache

这个目录是用户端 Nuxt storefront 的路由策略中心，包含页面级 SSR HTML 缓存策略、API 代理规则、静态资源缓存头、i18n 本地化路由规则和旧路由重定向规则。

长期约束：不要把新的页面缓存规则继续堆回 `nuxt.config.ts`。所有用户端公开页面的缓存策略，必须先在这里声明，再由 `route-rules.ts` 转成 Nuxt/Nitro `routeRules`。

## 为什么要做这层结构

Nuxt SSR 页面在高流量时会消耗 Node.js CPU。热门商品详情页、指南页、政策页如果每次请求都重新渲染 HTML，会让 Node 渲染进程先于 Go API 成为瓶颈。

长期方案是：

1. 公开、非个性化、可短暂过期的页面使用 Nitro route cache 缓存 SSR HTML。
2. 生产环境使用 Redis 作为共享缓存后端，避免多个 Node 进程或容器各自维护一份内存缓存。
3. 账号、购物车、结账、查询筛选、用户专属状态页面保持 `no-store`，除非后续为它们设计明确的 cache key。
4. 缓存策略、routeRules 生成、运行时存储挂载分层放置，避免后续维护变成临时补丁。

## 文件职责

- `html-cache-policies.ts`
  - 定义哪些公开页面可以缓存。
  - 定义每类页面的 `maxAge` 和 `staleMaxAge`。
  - 定义必须 `no-store` 的页面。

- `route-rules.ts`
  - 把 `html-cache-policies.ts` 中的策略转换为 Nuxt/Nitro `routeRules`。
  - 统一处理 i18n 语言前缀路由。
  - 统一处理 `/api/**` 代理、静态资源长缓存、旧路由重定向。

- `../../server/plugins/html-cache-storage.server.ts`
  - 运行时读取 `NUXT_HTML_CACHE_*` 和兼容的 `REDIS_*` 环境变量。
  - 当 `NUXT_HTML_CACHE_DRIVER=redis` 时，把 Nitro `cache` storage 挂载到 Redis。
  - 启动时先 ping Redis，成功才挂载；失败则 warning 并回退 Nitro 默认内存缓存。

- `../../server/routes/_internal/html-cache/purge.post.ts`
  - 内部 HTML cache purge 端点。
  - 只接受带 `x-html-cache-purge-token` 或 `Authorization: Bearer ...` 的请求。
  - 按 `/cache/html` 前缀列出并删除 Nitro route cache key，适用于 Redis 和默认内存缓存。
  - 生产公网入口必须拦截 `/_internal/**`，只允许 Docker 内部服务直连 storefront。

- `../../nuxt.config.ts`
  - 只调用 `buildStorefrontRouteRules(...)`。
  - 不直接维护具体页面缓存规则。

- `../../../go-backend/internal/service/storefront_html_cache_invalidator.go`
  - Go 后端内部 invalidator。
  - 商品、产品类型、文章、FAQ 等会影响公开 HTML 的写操作成功后，异步请求 Nuxt `/_internal/html-cache/purge`。
  - purge 失败只记录日志，不阻断后台保存，因为数据库仍然是事实源，HTML cache 只是加速层。

## 目前已经完成

- 已拆出 storefront route policy 目录，避免 `nuxt.config.ts` 继续膨胀。
- 已新增页面级 SSR HTML cache 策略。
- 已新增 Redis 运行时缓存挂载插件。
- 已在生产 compose 中为 storefront 增加 Redis 环境变量和 `data` network。
- 已在生产 compose 中为 Go API 和 storefront 增加内部 purge token。
- 已在本地 `docker-compose.yml` 中为 frontend 启用 Redis HTML cache，避免本地和生产缓存后端不一致。
- 已在 `deploy.sh` 中加入 Redis HTML cache 生产环境变量校验，避免 driver、token、TTL、SCAN_COUNT 配置错误时继续部署。
- 已在 Nginx 入口对公网 `/_internal/**` 返回 404，避免内部 purge 端点外露。
- 已在 `deployment/production.env.example` 中补充 HTML cache 相关环境变量示例。
- 已显式加入 `unstorage` 和 `ioredis` 依赖，避免依赖 Nitro 的传递依赖。
- 已新增 Go 后端商品/产品类型写操作后的自动 HTML cache purge。
- 已新增 Go 后端文章和 FAQ 写操作后的自动 HTML cache purge。
- 已新增 Go 后端 HTML cache purge 合并窗口，短时间内多次写操作会合并为一次 purge。
- 已验证 `npm run build` 通过。
- 已验证构建产物中包含 `/shop/**`、语言前缀路由、`no-store` 路由和 Redis cache 插件。

## 当前缓存策略

### 商品详情页

- 路由：`/shop/**` 和对应语言前缀，例如 `/fr/shop/**`、`/zh_cn/shop/**`
- fresh TTL：300 秒
- stale TTL：3600 秒

说明：商品详情页可能包含价格和库存快照，所以 fresh TTL 不能太长。购物车、结账、库存确认仍然必须以 API 为事实源。

### 语言首页

- 路由：语言前缀首页，例如 `/fr`、`/zh_cn`
- fresh TTL：300 秒
- stale TTL：3600 秒

说明：根路径 `/` 不缓存，因为它可能根据 i18n cookie 进行语言跳转，缓存后会污染不同用户的语言入口。

### 内容页

- 路由：`/blog/**`、`/guides/**`、`/picture-warehouse`、`/faq`
- fresh TTL：3600 秒
- stale TTL：86400 秒

说明：内容页变化频率低于商品库存/价格，适合更长缓存。

### 稳定政策和公司页面

- 路由：`/company/**`、`/policies/**`、`/support/faqs`、`/support/payment`、`/support/shipping`、`/support/warranty`
- fresh TTL：86400 秒
- stale TTL：604800 秒

说明：这类页面更新频率低，适合长缓存。如果后续支持页接入用户状态或实时查询，必须把对应页面移出这个策略。

## 当前明确 no-store 的页面

- `/`
- `/shop`
- `/membershipandpoints`
- `/spoke-calculator`
- `/support/product-feedback`
- `/support/test-report`
- `/support/warranty-check`
- `/api/**`

说明：

- `/shop` 是列表页，包含搜索、筛选和 query 状态；暂不缓存 HTML，避免缓存污染。
- `/api/**` 保持 `no-store`，API 数据缓存应由后端或明确的数据缓存层处理，不能混进页面 HTML 缓存。
- 会员、反馈、查询、计算器等页面可能包含用户输入或实时结果，默认不进 HTML cache。

## Redis 长期方案

生产环境推荐：

```env
NUXT_HTML_CACHE_DRIVER=redis
NUXT_HTML_CACHE_PREFIX=tanzanite:storefront:html-cache
NUXT_HTML_CACHE_REDIS_DB=1
NUXT_HTML_CACHE_REDIS_TTL_SECONDS=604800
NUXT_HTML_CACHE_PURGE_TOKEN=CHANGE_ME_HTML_CACHE_PURGE_TOKEN_AT_LEAST_32_CHARS
```

可选变量：

```env
NUXT_HTML_CACHE_REDIS_URL=
NUXT_HTML_CACHE_REDIS_HOST=redis
NUXT_HTML_CACHE_REDIS_PORT=6379
NUXT_HTML_CACHE_REDIS_PASSWORD=
NUXT_HTML_CACHE_REDIS_CONNECT_TIMEOUT_MS=1000
NUXT_HTML_CACHE_REDIS_MAX_RETRIES=1
NUXT_HTML_CACHE_REDIS_SCAN_COUNT=100
NUXT_HTML_CACHE_LOG=silent
STOREFRONT_HTML_CACHE_PURGE_URL=http://storefront:3000/_internal/html-cache/purge
STOREFRONT_HTML_CACHE_PURGE_TOKEN=${NUXT_HTML_CACHE_PURGE_TOKEN}
STOREFRONT_HTML_CACHE_PURGE_DEBOUNCE_MS=500
```

规则：

- 多 Node 进程或多容器部署时必须使用 Redis，否则每个进程只会有自己的内存缓存。
- 本地 Docker compose 也启用 Redis HTML cache，便于在上线前复现生产缓存行为。
- Redis 只缓存公开页面 HTML，不作为商品、库存、订单、用户状态的事实源。
- 商品价格、库存、SKU 重量、购物车、结账必须继续走 API 实时确认。
- `NUXT_HTML_CACHE_PREFIX` 必须稳定，避免多个项目或环境共用同一个 key 空间。
- 变更缓存 key、prefix、TTL 时，需要评估是否要清理 Redis 旧 key。
- `NUXT_HTML_CACHE_PURGE_TOKEN` 和 `STOREFRONT_HTML_CACHE_PURGE_TOKEN` 必须保持一致。
- Go API 只通过 Docker 内部地址调用 `STOREFRONT_HTML_CACHE_PURGE_URL`，不要从公网域名绕一圈。
- `NUXT_HTML_CACHE_REDIS_SCAN_COUNT` 用于控制 purge 时 Redis `SCAN` 每轮建议扫描数量，缓存 key 变多后可按实际延迟调大。
- `STOREFRONT_HTML_CACHE_PURGE_DEBOUNCE_MS` 用于合并短时间内的重复 purge 请求，默认 500ms；后台批量保存或连续编辑时通常只会触发一次实际 purge。

## 缓存失效链路

当前自动失效范围：

- 后台创建商品。
- 后台更新商品基础信息、规格、SKU、媒体、价格、库存、状态。
- 后台删除商品。
- 后台批量更新商品状态或批量删除商品。
- 后台创建、更新、删除文章。
- 后台批量更新或删除文章。
- 后台创建、更新、删除 FAQ。
- 后台更新 FAQ 排序或批量删除 FAQ。
- 后台创建、更新、删除产品类型和规格模板。

失效方式：

1. Go 后端先完成数据库写入。
2. Go 后端清理对应的数据缓存，例如商品缓存或文章缓存。
3. Go 后端把 HTML purge 请求放入合并窗口，短时间内多次写操作合并成一次实际请求。
4. Nuxt 校验 purge token。
5. Nuxt 按 `/cache/html` 前缀删除对应 Nitro route cache key，并返回本次删除数量。

这个版本先采用“清理全部 storefront HTML cache”的策略。原因是 Nitro route cache 的 key 内部包含 URL hash 和 varies header，精确删除某一个商品详情页需要复刻 Nitro key 生成逻辑，维护风险更高。全量清理的成本可控：当前缓存范围只包含公开页面 HTML，不包含 API 数据、订单、购物车或用户状态。

## 缓存页面的数据源边界

| 页面范围 | SSR HTML 中的数据源 | 当前失效责任 |
| --- | --- | --- |
| `/shop/**` | Go 商品详情、商品媒体、SKU 等公开数据 | 商品和产品类型写操作触发 HTML purge |
| `/blog/**` | Go 文章列表和文章详情，失败时回退本地 mock | 文章写操作触发 HTML purge |
| `/support/faqs` 和各页面 `PageFaq` | Go FAQ，失败时回退本地 FAQ 文件 | FAQ 写操作、排序操作触发 HTML purge |
| `/guides/**` | 静态指南内容 + `PageFaq`；商品搜索抽屉是用户点击后才请求 API | FAQ 写操作触发 HTML purge；点击后搜索结果不进 HTML cache |
| `/company/**` | 静态公司内容 + `PageFaq` | FAQ 写操作触发 HTML purge；公司静态文案变更需要重新构建部署 |
| `/picture-warehouse` | SSR 只输出页面壳；图片库列表、评论和上传在客户端挂载或交互后请求 API | showcase 数据不进 SSR HTML cache；无需触发 HTML purge |
| `/faq` | 本地 FAQ/i18n 内容 | 代码或文案变更需要重新构建部署 |

判断规则：

- 如果页面在 SSR 阶段通过 `useAsyncData`、`useFetch` 或直接 `$fetch` 读取 Go API，并且该页面进入 HTML cache，那么对应 Go 写操作必须接入 `StorefrontHTMLCacheInvalidator`。
- 如果页面只在 `onMounted`、用户点击、弹窗打开后请求 API，那么 HTML cache 只缓存页面壳；这类 API 数据不需要触发 HTML purge。
- 如果页面包含用户态、价格实时计算、运费计算、购物车、结账、会员权益或查询结果，默认先放进 `storefrontNoStorePagePaths`，等有明确 cache key 和失效策略后再缓存。

如果后续商品量和内容量大到全量 purge 成本明显升高，再升级为：

- 维护业务侧 cache tag，例如 `product:{id}`、`product-type:{id}`。
- 写入时记录 route path 到 Redis set。
- purge 时按 tag 删除相关 key。
- 或增加 HTML cache version，让变更后的页面自然写入新 namespace，旧 namespace 由 Redis TTL 回收。

## 后续修改必须同步更新这里

任何人后续做以下改动时，必须同步更新本 README：

- 新增、删除或调整 HTML cache 页面范围。
- 调整 `maxAge`、`staleMaxAge` 或 Redis TTL。
- 把某个页面从 `no-store` 改为可缓存，或从可缓存改为 `no-store`。
- 新增涉及用户状态、价格、库存、运费、购物车、结账、会员权益的页面。
- 修改 Redis 环境变量、Docker network、compose 连接方式。
- 修改 `server/plugins/html-cache-storage.server.ts` 的 Redis 挂载逻辑。
- 修改 `server/routes/_internal/html-cache/purge.post.ts` 的 purge 逻辑。
- 修改 Go 后端 `storefront_html_cache_invalidator.go` 或任何自动触发 HTML cache purge 的位置。

如果代码改了但文档没改，默认视为未完成。

## 验证建议

每次调整这套策略后至少运行：

```bash
npm run build
npm run check:html-cache
npm run smoke:html-cache
```

并检查：

- 构建是否通过。
- `check:html-cache` 是否通过。
- Nitro 产物中是否包含目标 `routeRules`。
- `/shop`、`/api/**`、`/_internal/**` 仍然是 `no-store`。
- `/shop/**` 和各语言前缀商品详情页仍然使用 `/cache/html`。
- purge 端点产物仍然按 `/cache/html` 列 key、删 key，并返回 `purgedKeys`。
- `smoke:html-cache` 是否能启动本地 preview，并实际缓存一个页面后通过 purge 删除至少 1 个 HTML key。
- Redis 不可用时 Nuxt 仍能启动并 warning 回退内存缓存。

部署前还应在接近生产的环境里请求同一个热门商品详情页两次，确认第二次命中共享缓存，且 Redis 中出现对应 `NUXT_HTML_CACHE_PREFIX` key。

生产部署脚本 `deploy.sh` 会在拉取镜像前校验：

- `NUXT_HTML_CACHE_DRIVER=redis`。
- `NUXT_HTML_CACHE_PREFIX` 非空。
- `NUXT_HTML_CACHE_REDIS_DB` 是非负整数。
- `NUXT_HTML_CACHE_REDIS_TTL_SECONDS` 和 `NUXT_HTML_CACHE_REDIS_SCAN_COUNT` 是正整数。
- `NUXT_HTML_CACHE_PURGE_TOKEN` 至少 32 位，且不能复用 `REDIS_PASSWORD` 或 `JWT_SECRET`。
- `STOREFRONT_HTML_CACHE_PURGE_DEBOUNCE_MS` 是正整数。

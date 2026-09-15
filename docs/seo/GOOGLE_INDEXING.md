# Google URL Notification

商品详情页不使用 Google Indexing API。Google 官方将该 API 限定为
JobPosting 或包含 BroadcastEvent 的 VideoObject 页面；普通电商商品
URL（例如 /products/:slug）不在允许范围内。

本项目已在服务层和管理后台禁用商品 URL 的 URL_UPDATED 推送。即使遗留
配置开启、凭据完整，商品推送也会在读取商品、申请 URL 冷却、获取 token
或发出 HTTP 请求之前被服务层拒绝；后台入口同样直接返回策略错误，避免因
违规提交导致 Search Console 处置或全站降权。

## 受支持的收录方式

- 为商品和分类页生成并维护站点地图。
- 在 Google Search Console 中提交站点地图，让 Google 按正常抓取流程发现
  商品 URL。
- 使用结构化数据、规范链接和稳定的内部链接帮助 Google 抓取商品页。

商品发布和保存流程不会调用 Google Indexing API，也不存在商品批量推送、
删除通知或抓取状态查询。

## 遗留配置

GOOGLE_INDEXING_ENABLED、服务账号凭据和 google_indexing YAML 配置仅为
兼容旧部署而保留。不要为商品页启用或配置该 API；后台状态会明确显示
“商品页不适用”。如部署中仍有旧凭据，建议从 Secret Manager、环境变量和
Google Cloud 项目中撤销或删除。

## 管理后台与 API

SEO 商品页不再显示“通知 Google”按钮，也不会请求 Google Indexing 状态。
遗留的 POST /api/admin/seo/products/:id/indexing 端点保留为兼容层，但对
所有商品返回 422 和 google_indexing_product_unsupported，不会产生外部
副作用，也不再配置幂等或推送限流中间件。站点地图工作流仍是商品收录入口。

## 官方参考

- [Google Indexing API 文档](https://developers.google.com/search/apis/indexing-api/v3/quickstart)
- [Google Search Central：Indexing API 使用范围](https://developers.google.com/search/apis/indexing-api)

# 运维中心网页登录授权

运营中心的连接器支持从后台一键启动网页登录授权。授权完成后，后端会保存加密凭据，并自动把当前选中环境中已有的台账绑定到远端资源。

## 配置

后端必须配置：

```text
OPS_CONNECTOR_MASTER_KEY=<稳定的后端密钥>
OPS_CONNECTOR_OAUTH_REDIRECT_URL=https://<后台域名>/api/admin/ops/connectors/oauth/callback
OPS_CONNECTOR_OAUTH_POST_CONNECT_URL=https://<管理后台域名>/ops/connectors
```

Cloudflare 还需要在 Cloudflare OAuth Clients 中创建 OAuth client，并配置：

```text
OPS_CLOUDFLARE_OAUTH_CLIENT_ID=<Cloudflare client id>
OPS_CLOUDFLARE_OAUTH_CLIENT_SECRET=<private client secret；public PKCE client 可留空>
OPS_CLOUDFLARE_OAUTH_SCOPES=<可选：与 OAuth client 中已授权 scope 一致>
```

Cloudflare 的 redirect URL 必须与 OAuth client 中的值完全一致。`OPS_CLOUDFLARE_OAUTH_SCOPES` 留空时使用 OAuth client 自己配置的默认 scope，避免把 Cloudflare 权限名称硬编码在应用里。Hostinger 使用官方 OAuth client 动态注册，不需要把 Hostinger 登录密码或 API Token 放进前端。
开发环境如果管理后台和 API 使用不同端口，`OPS_CONNECTOR_OAUTH_POST_CONNECT_URL` 必须指向管理后台端口。
Hostinger 这里由本后端直接调用 OAuth issuer 完成授权码回调，不依赖 MCP HTTP transport。正式域名部署后必须用真实 Hostinger 账号完成一次网页登录，确认该回调域名已被授权服务器接受。

## 使用

1. 打开后台的“运营中心 / 连接器中心”，先选择要操作的环境。
2. 需要同时接入两个平台时，点击“一键连接全部”：先登录 Hostinger，回调成功后自动继续 Cloudflare，两个步骤都使用当前环境。
3. 只接入一个平台时，点击对应的“连接 Cloudflare”或“连接 Hostinger”。列表中的钥匙图标可对已有同环境连接器重新授权。
4. 在提供商页面完成登录和授权。
5. 回到后台后检查连接器、服务器、项目和域名状态。
6. 打开部署中心重新执行 Preflight。

Hostinger 会优先匹配当前环境已登记的 VPS ID `provider_resource_id`，然后匹配该 VPS 上已登记的 Compose 项目名。Cloudflare 会按当前环境已有域名的 zone 或域名后缀匹配 Zone，并写入对应连接器。未知远端资源不会被静默写入其他环境。

如果授权、token 保存和连接器测试成功，但部分 VPS、项目、域名绑定或同步失败，后台会显示“已连接但需要复核”，并继续返回已成功绑定的数量。此状态不代表所有台账都已完成；请按提示回到对应 TAB 检查 `provider_resource_id`、Compose 项目名、Zone 和观察状态。

如果 Hostinger 账号下有多个 VPS，但当前环境台账没有对应的 `provider_resource_id`，系统只展示候选资源，不会把未知服务器静默绑定到项目。此时先在 VPS 台账中补一次资源选择，再重新连接。

## 凭据边界

- access token 和 refresh token 只在后端加密存储，接口和前端不回显原值。
- OAuth `state`、PKCE verifier 和回调会话保存于 `ops_connector_oauth_sessions`，回调只能使用一次且十分钟后失效。
- 连接测试、Zone/VPS/项目发现和同步均为只读操作；部署发布仍然需要从部署中心显式发起。
- 连接器 access token 过期后，后端使用 refresh token 自动换新并重新加密保存。

## 官方端点

- Cloudflare OAuth 文档：`https://developers.cloudflare.com/fundamentals/oauth/`
- Cloudflare OAuth 集成端点：`https://dash.cloudflare.com/oauth2/auth`、`https://dash.cloudflare.com/oauth2/token`
- Hostinger 官方 OAuth issuer：`https://auth.hostinger.com`
- Hostinger 官方 API：`https://developers.hostinger.com`

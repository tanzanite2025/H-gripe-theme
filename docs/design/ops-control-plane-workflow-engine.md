# 运维控制台与工作流引擎设计

状态：实施中（运维总览、域名中心、连接器中心、VPS/项目中心、Hostinger 只读同步、部署 Preflight、dry-run、受控 Hostinger 生产工作流和按域名 Cloudflare 缓存清理第一版已完成）
文档类型：后台产品与系统架构设计  
适用范围：当前 `tanzanite-theme` 项目的 Hostinger、Cloudflare、域名、VPS 和发布运维

## 1. 文档目的

本文件把此前对话中提出的“运维控制台 + 工作流引擎”正式独立出来，作为后续后台开发的设计基线。

目标不是把一堆 SSH、Docker、DNS 命令搬到后台，而是让后台成为运维配置的源数据入口，由系统根据已确认的配置生成部署、网关、DNS 和检查步骤，并保留完整的安全控制与操作证据。

核心原则：

- 连接器负责连接外部平台，不把 Hostinger 或 Cloudflare 的细节散落到业务页面。
- 域名中心负责维护域名关系、解析和跳转，不把域名写死在多个配置文件中。
- 部署中心负责把发布动作编排成可检查、可确认、可回滚的工作流。
- 后台保存声明式配置，系统生成 Compose、Caddy、Nginx、DNS 检查和部署参数。
- 所有有副作用的动作都必须有权限、预览、确认、审计和失败处理。
- 密钥只保存加密引用或安全凭据，不在页面和日志中展示原文。

## 2. 来源与当前记录

这份规划来源于此前关于后台运维配置的对话记录，核心结构为：

- 连接器层：Cloudflare、Hostinger 各自独立连接，支持 OAuth、API Token 或手动密钥。
- 域名中心：主域名、别名、admin 域、跳转域、验证域统一维护，可新增、停用和切换。
- 部署中心：把“检查环境 -> 同步配置 -> 发布 -> 健康检查 -> 清缓存 -> 回滚点”编排成工作流。
- 安全层：密钥加密存储、角色权限、操作审计、确认弹窗、dry-run。
- 生成层：后台是源数据，系统生成 Compose、Caddy、Nginx 和 DNS 配置，不允许人工直接改生成结果作为长期状态。

现有仓库记录中，`docs/ops/brand-rename-continuity-plan.md` 曾明确把 Cloudflare、Hostinger 和部署编排后台列为后续阶段；该阶段计划已完成并归档。本文件记录后台控制台第一版及其后续边界。

## 3. 当前基线

以下能力已经存在，可以作为控制台的第一批适配对象：

| 能力 | 当前来源 | 当前状态 |
| --- | --- | --- |
| Hostinger VPS 目标和网络边界 | `docs/ops/hostinger-vps-docker-runbook.md` | 已有执行手册 |
| Hostinger Docker Compose 发布 | `compose.prod.yml`、`deploy.sh` | 已有发布路径 |
| GHCR 镜像发布 | `.github/workflows/publish-images.yml` | 已有发布工作流 |
| VPS 边界检查 | `deployment/verify-vps-release-boundary.sh` | 已有静态和运行时检查 |
| Cloudflare/Caddy 网关边界 | `deployment/EDGE_SECURITY_RUNBOOK.md`、`deployment/edge/` | 已有配置和安全手册 |
| 当前生产域名 | `learn.gripe`、`www.learn.gripe`、`admin.learn.gripe` | 已有运行配置 |
| 后台权限和审计 | `go-backend/internal/domain/auth/`、`audit/` | 已有通用能力 |
| 后台设置和密钥加密边界 | `go-backend/internal/service/`、`web/admin/src/views/Settings.vue` | 已有业务设置能力 |

这些内容不是运维控制台已经完成的证明。它们只是第一阶段应复用的执行基础，不能继续依赖人工在多个文件之间查找和复制。

## 3.1 当前实施状态

已完成第一版闭环：

- 后台一级域：`运维中心`。
- `域名中心`：维护域名角色、环境、提供商、Zone、目标、代理、TLS、跳转、启停和审计。
- `连接器中心`：维护 Cloudflare、Hostinger、GitHub/GHCR 等连接登记，支持凭据加密存储、字段脱敏、启停、只读连接测试和审计；列表默认生产环境，可切换或查看全部环境。
- 基础独立权限：`ops:view`、`ops:domain:view`、`ops:domain:edit`、`ops:domain:sync`、`ops:connector:view`、`ops:connector:edit`。
- `VPS 中心`：维护 Hostinger VPS ID、主机名、IPv4、系统、期望状态和观察状态；列表默认生产环境，可切换或查看全部环境。
- `项目中心`：维护 VPS 绑定、Hostinger 连接器、Compose 来源、服务、网络、卷、网关别名、镜像标签、Commit SHA、健康状态、部署记录和备份恢复说明；列表默认生产环境，可切换或查看全部环境。
- `运维总览`：按环境聚合 VPS、项目、域名和连接器台账，默认生产环境，展示拓扑、待处理项和运维审计记录。
- 运维总览的环境选择会写入路由上下文；从总览进入 VPS、项目、域名或连接器中心时会保留相同环境，避免概览与台账页面查询不同环境的数据。
- `后台账号`入口保留在运维导航以便发布值守时访问，但它属于身份管理而非基础设施台账：账号目录和创建/重置操作均只允许 `admin` 角色，前端路由使用管理员专属的 `system:manage` 权限。审计日志的 `user_id` 和 `username` 表示操作者，`resource=user` 与 `resource_id` 表示被创建或重置的目标账号；无登录主体的 `adminctl` bootstrap 会保留空操作者 ID 和 `adminctl` 标识。
- 部署中心沿用相同的环境路由上下文：项目台账和 Preflight 总览都会向后端传递具体环境，切换环境时清除跨环境的项目、报告和工作流选择；`all` 仅用于明确的全量视图，不能作为后端环境参数。预检总览接口会拒绝非法环境值，避免界面显示筛选结果而请求仍然读取全量项目。
- 域名中心列表默认生产环境，支持按生产、预发布、测试、本地或全部环境筛选；非法环境值由后端拒绝，避免前端筛选退化为静默全量查询。
- 域名中心表单、列表、Cloudflare 同步、差异明细和 DNS/Caddy/Nginx 预览职责独立；编辑时项目和 Cloudflare 连接器只展示同环境选项，环境或提供商切换会清除不兼容绑定。
- 域名中心已增加 `Desired State / Observed State` 展示边界：期望状态继续由后台维护，观察状态默认是 `unknown`，可通过 Cloudflare 手动只读同步填充；单域同步直接局部回写观察字段，不覆盖期望台账或刷新整页。
- 连接器中心已将表单、OAuth 编排和列表观察状态拆分：OAuth 可从当前环境或指定连接器重新授权，一键连接会在同一环境内依次授权 Hostinger 和 Cloudflare；回调发现和绑定只处理同环境 VPS、项目和域名，不会跨环境修改台账。连接测试完成后只局部回写当前连接器状态，不刷新整页。
- Cloudflare 域名只读同步第一版已完成：域名可绑定 Cloudflare 只读连接器，读取 Zone、A/AAAA/CNAME、代理状态和 SSL/TLS 模式，写回 Observed State，并将同步动作写入运维审计。
- Hostinger VPS / 项目只读同步已完成：VPS 读取并持久化远端状态、同步来源、主机名、IPv4、系统信息、计划和数据中心；项目读取并持久化 Docker 项目状态、同步来源、容器总数、运行数和健康数，写回 Observed State，并将同步动作写入运维审计。
- VPS/项目编辑保存已和 Observed State 分离：后台表单只维护声明式台账和期望状态，不再手工覆盖 Hostinger 同步得到的远端主机信息、观察状态、检查时间、错误摘要和容器计数。
- VPS 与连接器必须处于同一提供商和环境；已有项目绑定的 VPS 不允许直接变更提供商或环境，必须先迁移项目绑定，避免跨环境台账关系失真。
- 项目可以显式绑定同环境的 Hostinger 连接器，也可以沿用绑定 VPS 的连接器；保存和 Hostinger 只读同步都会拒绝项目、VPS、连接器之间的跨环境组合，连接器名称和继承来源在项目列表中可见。
- 域名中心配置只读预览和差异明细已完成：可根据域名 Desired State 生成 DNS 记录草稿、Caddy 路由草稿和 Nginx 路由草稿，并查看 Desired/Observed 的目标、代理、TLS 和 Zone 差异；当前不会写入 Cloudflare、Hostinger、Caddy、Nginx 或生产网关。
- 部署中心 dry-run 第一版已完成：可以从项目 Preflight 创建持久化工作流，重新验证、提交人工审批、执行只读步骤并固化步骤证据；该执行不会写入 Hostinger、Cloudflare、Docker、Compose、Caddy、Nginx 或生产网关。
- Hostinger 生产工作流第一版已完成：生产工作流只能绑定 `production` 环境项目，并且必须经过 Preflight、人工审批、项目锁和执行权限校验，按当前生产运行手册调用既有 Compose 项目更新，随后重新同步项目并执行 DNS、HTTP、HTTPS 健康检查；发布后健康检查失败会进入 `rollback_required`，不会未经人工确认自动回滚。
- 发布后健康检查通过后，工作流会按当前项目绑定的 Cloudflare 域名和 Zone 分组清理缓存；没有 Cloudflare 域名时显式跳过，缓存清理失败会保留工作流失败证据，不自动回滚源站。
- 工作流已记录发布前回滚点、幂等键、远端操作 ID、健康检查快照和项目锁；回滚点优先保存可作为 `deploy.sh` 的 `DEPLOY_REF` 使用的 40 位 Commit SHA；锁过期后可恢复，不会因为后台进程中断永久占用项目；失败、暂停或需回滚状态下，可从标记为可重试的失败步骤继续，已成功步骤不会重跑。
- 真实回滚已接入受限 SSH 执行器：只允许生产工作流使用绑定 VPS 的主机、固定工作目录和 `DEPLOY_REF=<40位SHA> ./deploy.sh` 命令，强制使用 `known_hosts` 主机指纹校验并限制输出；回滚后会重新同步 Hostinger、执行 DNS/HTTP/HTTPS 健康检查和 Cloudflare 缓存清理，所有结果写回工作流步骤和审计。
- 独立权限：`ops:view`、`ops:domain:view`、`ops:domain:edit`、`ops:domain:sync`、`ops:connector:view`、`ops:connector:edit`、`ops:vps:view`、`ops:vps:edit`、`ops:vps:sync`、`ops:project:view`、`ops:project:edit`、`ops:project:sync`、`ops:deploy:view`、`ops:deploy:dry_run`、`ops:deploy:execute`、`ops:deploy:rollback`、`ops:workflow:approve`。工作流取消按模式校验：dry-run 需要 `ops:deploy:dry_run`，production 需要 `ops:deploy:execute`；重试/继续沿用执行权限。
- 独立迁移：`116_create_ops_domain_bindings`、`117_create_ops_connectors`、`118_create_ops_vps_bindings`、`119_create_ops_project_bindings`、`120_add_ops_domain_observed_state`、`122_bind_ops_domains_to_connectors`、`124_add_ops_hostinger_observed_state`、`125_add_ops_vps_observed_identity`、`137_create_ops_deployment_workflows`、`139_extend_ops_deployment_workflows`、`160_create_ops_connector_oauth_sessions`。
- 当前生产初始基线已登记：Hostinger VPS `1834903`、`srv1834903.hstgr.cloud`、`2.25.85.201`、`commerce-platform` 项目。

当前明确未完成：

- Hostinger 的完整发布编排和任意 SHA 生产发布；当前正常生产工作流仍限制为既有 Compose 项目的 `IMAGE_TAG=master` 受控更新，任意 SHA 只能通过已登记的受限 SSH 回滚路径执行。
- Cloudflare Zone、DNS 和代理的写入动作；当前域名中心只提供 DNS 与网关配置草稿预览。
- Cloudflare/Hostinger 的统一 DNS、网关真实写入，以及真实回滚后的恢复演练；生产项目回滚执行器已完成，但必须先配置 SSH 私钥、`known_hosts`、绑定 VPS 主机和工作目录。
- 定时漂移检测。

`运维总览` 当前是声明式台账的聚合展示，不等于 Hostinger、Cloudflare 或 VPS 的实时状态查询。总览接口为 `GET /api/admin/ops/overview?environment=production`；`environment` 支持 `production`、`staging`、`test`、`local`，未传时默认生产环境。部署 Preflight 总览接口为 `GET /api/admin/ops/deployments/preflight-overview?environment=production`，遵循同一环境枚举；未传时才表示全量汇总。未同步、未知和待测试状态会保留为待处理项，不会被折算为健康。

域名中心当前已经可以分别显示期望目标、实际目标、期望代理/TLS、实际代理/TLS，以及观察来源和最后观察时间。列表默认只查询生产域名，切换环境后重新查询对应台账；Cloudflare 域名支持单域同步和批量手动同步，单域同步完成后只更新当前行的观察字段。尚未建立定时同步任务时，未同步记录仍保持 `unknown`。

## 4. 产品边界

### 4.1 第一阶段必须解决

1. 记录并展示当前 Hostinger VPS、项目、网络和服务状态。
2. 记录并展示 Cloudflare 连接、Zone、DNS、代理、SSL/TLS 和缓存相关状态。
3. 统一维护主域名、`www`、`admin`、跳转域、验证域和内部服务域的关系。
4. 生成域名解析、网关路由、部署环境和检查清单。
5. 提供部署前 dry-run、配置差异预览、发布确认、健康检查和回滚点记录。
6. 记录每次操作的执行人、目标、输入配置、结果、失败原因和外部平台返回摘要。
7. 支持手动刷新外部状态，并显示“最后同步时间”和“状态来源”。

### 4.2 第一阶段明确不做

- 不在后台直接托管完整 SSH 终端。
- 不允许后台任意执行用户输入的 Shell 命令。
- 不自动创建或销毁生产 VPS，除非后续单独设计资源编排和审批流程。
- 不替代 Hostinger 和 Cloudflare 官方控制台的全部能力。
- 不把 ERP、其他项目或其他仓库的基础设施纳入本模块。
- 不把支付、订单、商品等业务工作流混入运维工作流。

## 5. 控制台模块

后台新增独立的“运维中心”，建议分为以下页面：

### 5.1 运维总览

展示当前环境的整体状态：

- 当前环境：生产、预发布或本地。
- 绑定的 Hostinger VPS 和项目。
- Cloudflare Zone 和代理状态。
- 当前主域名及别名。
- 当前部署版本、镜像标签或 Commit SHA。
- 最近一次部署结果。
- 最近一次健康检查结果。
- 待处理告警、配置漂移和失败工作流。

总览只做状态汇总，不直接承载所有配置编辑。

### 5.2 连接器

连接器是外部平台的统一适配层。

第一批连接器：

- Cloudflare Connector
- Hostinger Connector
- GitHub/GHCR Connector

每个连接器至少记录：

- 连接名称和用途。
- 提供商类型。
- 认证方式：OAuth、API Token、API Key 或手动密钥引用。
- 凭据引用，不保存页面可读的原始密钥。
- 作用域和允许操作。
- 测试连接结果。
- 最后成功同步时间。
- 最后失败原因。
- 启用、停用和轮换状态。

连接器测试只验证权限和读取能力。写入、发布、删除等高风险操作必须由具体工作流显式申请。

### 5.3 域名中心

域名中心是这次建设的核心页面，用于承接当前绑定关系和后续维护。

域名实体建议包含：

| 字段 | 说明 |
| --- | --- |
| Domain | 完整域名 |
| Role | 主域名、别名、`admin` 域、跳转域、验证域、内部域 |
| Environment | 生产、预发布、测试 |
| Provider | Cloudflare、Hostinger 或其他注册商/解析服务 |
| Zone | 所属 DNS Zone |
| Target | 当前目标，例如 VPS IP、Caddy 服务或外部 URL |
| Proxy | DNS only 或 proxied |
| TLS Mode | 当前 SSL/TLS 模式 |
| Redirect | 是否跳转及跳转目标 |
| Status | active、pending、disabled、drifted、error |
| Last Checked | 最近检查时间 |
| Notes | 维护备注和变更原因 |

域名中心必须支持：

- 新增、停用和重新启用域名。
- 设置主域名和别名。
- 维护 `www`、`admin`、跳转域和验证域。
- 查看 DNS 记录差异。
- 查看 Cloudflare Zone 状态。
- 预览将要生成的 Caddy/Nginx 路由。
- 预览将要执行的 DNS 变更。
- 记录切换前后的解析、跳转和 HTTPS 检查。
- 保留旧域名作为迁移期别名，但不能让历史域名继续成为当前源数据。

域名页面不得直接把 DNS 记录当作唯一事实。DNS 记录应当分为：

1. Desired State：后台维护的期望状态。
2. Observed State：从 Cloudflare、Hostinger 或公共 DNS 检查得到的实际状态。
3. Generated State：系统生成并下发的配置。

三者不一致时，页面必须显示配置漂移，而不是静默覆盖。

### 5.4 VPS 与项目

用于记录当前绑定的 Hostinger 资源：

- VPS ID、主机名、IPv4、区域和操作系统。
- Hostinger 项目 ID、项目名和 Compose 来源。
- 当前运行的服务。
- Docker 网络和卷边界。
- 当前镜像标签、Commit SHA 或 digest。
- 服务健康状态。
- 最后一次部署和最后一次检查。
- 备份和恢复演练记录。

第一阶段以读取、同步和发布现有项目为主，不自动创建新项目。当前生产目标以 `docs/ops/hostinger-vps-docker-runbook.md` 中记录的 VPS 和 `commerce-platform` 项目为初始基线，但实际调用前必须再次从 Hostinger 查询并确认，不能只信任文档中的静态值。

### 5.5 部署中心

部署中心把发布动作拆成可观察的步骤：

1. 检查连接器和凭据。
2. 检查 VPS、项目和外部网络。
3. 检查镜像是否存在且版本一致。
4. 生成并预览环境、Compose、Caddy、Nginx 和 DNS 变更。
5. 执行人工确认。
6. 同步配置。
7. 发布指定版本。
8. 等待服务健康。
9. 执行 HTTP、HTTPS、DNS、WebSocket 和边界检查。
10. 按需清理 Cloudflare 缓存。
11. 固化发布证据和回滚点。
12. 失败时暂停并进入回滚或人工处理。

部署中心必须支持：

- `master` 流动标签发布。
- 完整 Commit SHA 或 digest 固定发布。
- dry-run。
- 配置差异预览。
- 单步重试。
- 从失败步骤继续。
- 取消尚未执行的后续步骤。
- 记录发布前版本和可回滚版本。
- 发布后自动生成检查报告。

## 6. 工作流模型

工作流不是一段拼接好的 Shell 字符串，而是结构化步骤和状态转换。

### 6.1 工作流实体

- `WorkflowDefinition`：流程定义和版本。
- `WorkflowRun`：一次实际执行。
- `WorkflowStep`：步骤定义。
- `WorkflowStepRun`：步骤执行记录。
- `ApprovalRequest`：需要人工确认的节点。
- `ExecutionArtifact`：配置快照、差异、检查报告和发布证据。
- `RollbackPoint`：可恢复版本及其依赖信息。

### 6.2 步骤类型

建议第一批支持：

- `check_connector`
- `discover_vps`
- `discover_project`
- `check_image`
- `render_config`
- `diff_config`
- `record_rollback_point`
- `update_project`
- `apply_dns`
- `apply_gateway`
- `update_project`
- `health_check`
- `post_deploy_health_check`
- `purge_cache`
- `record_release`
- `rollback_project`

### 6.3 状态

工作流状态：

`draft -> validated -> awaiting_approval -> running -> succeeded`

异常状态：

`failed`、`cancelled`、`paused`、`rollback_required`、`rolled_back`

每个步骤都必须有：

- 输入快照。
- 输出摘要。
- 开始和结束时间。
- 重试次数。
- 可重试标记。
- 是否产生外部副作用。
- 外部请求 ID 或操作 ID。
- 脱敏后的日志。

### 6.4 幂等和锁

- 同一个环境同一时间只允许一个高风险部署工作流运行。
- DNS、路由和项目更新必须使用幂等操作。
- 重试前必须判断上一次请求是否已经在外部平台成功。
- 工作流必须带环境锁和资源锁，避免两个操作同时修改同一 Zone、VPS 项目或网关。
- 生成配置应带版本号或内容摘要，防止旧工作流覆盖新配置。

## 7. 配置源与生成层

后台维护声明式配置，系统生成执行文件和请求参数。

### 7.1 Desired State

建议按环境维护以下配置对象：

- Environment
- ProviderConnection
- VPSBinding
- ProjectBinding
- DomainBinding
- DNSRecordIntent
- GatewayRouteIntent
- DeploymentPolicy
- HealthCheckSuite
- BackupPolicy

### 7.2 生成结果

生成层第一批输出：

- `compose.prod.yml` 的环境变量和项目参数。
- Caddy 路由片段。
- Nginx 域名和反向代理配置。
- Cloudflare DNS、代理和缓存规则请求。
- Hostinger 项目更新参数。
- 发布前检查清单。
- 发布后健康检查清单。
- 回滚说明和证据索引。

生成结果必须可预览、可下载、可比较，但不能被当作后台的长期源数据。手工修改生成结果后，下一次生成可以检测到差异并要求确认。

## 8. 安全设计

### 8.1 凭据

- 使用独立的加密密钥加密连接器凭据。
- 数据库只保存加密值、密钥版本、权限范围和最后使用时间。
- 列表页只显示提供商、名称、范围和状态。
- 禁止把 Token、Secret、私钥写入普通审计日志、浏览器控制台或工作流输出。
- 支持轮换、撤销和连接测试。

### 8.2 权限

建议增加独立权限域：

- `ops:view`
- `ops:connector:view`
- `ops:connector:edit`
- `ops:domain:view`
- `ops:domain:edit`
- `ops:deploy:dry_run`
- `ops:deploy:execute`
- `ops:deploy:rollback`
- `ops:workflow:approve`
- `ops:audit:view`

查看状态和执行变更必须分开授权。生产发布、DNS 切换、缓存清理和回滚不应共用一个普通编辑权限。

### 8.3 高风险确认

以下动作必须二次确认：

- 修改主域名或 `admin` 域。
- 修改 DNS 委派、A/AAAA/CNAME 或代理状态。
- 修改生产网关路由。
- 更新生产项目。
- 清理生产缓存。
- 回滚生产版本。
- 删除或停用连接器、域名或项目绑定。

确认内容必须显示目标、差异、影响范围、预计步骤和回滚方式。

### 8.4 审计

复用现有 `audit_logs` 能力，并增加运维资源类型和工作流关联信息。

每次操作至少记录：

- 操作人。
- 角色和权限。
- 环境。
- 资源类型和资源 ID。
- 工作流和步骤 ID。
- 操作前摘要。
- 操作后摘要。
- 脱敏后的差异。
- 结果、错误和外部操作 ID。
- IP、User-Agent、开始时间、结束时间和耗时。

## 9. 后端边界

建议新增独立的 `ops` 领域，而不是把运维配置塞进通用站点设置。

建议的后端边界：

```text
internal/domain/ops/
internal/repository/ops_*.go
internal/service/ops_*.go
internal/api/admin/ops_*.go
web/admin/src/views/Ops*.vue
web/admin/src/components/admin/ops/
```

建议 API 分组：

```text
/api/admin/ops/overview
/api/admin/ops/connectors
/api/admin/ops/connectors/:id/test
/api/admin/ops/domains
/api/admin/ops/domains/:id/diff
/api/admin/ops/vps
/api/admin/ops/projects
/api/admin/ops/workflows
/api/admin/ops/workflows/:id
/api/admin/ops/workflows/:id/validate
/api/admin/ops/workflows/:id/approve
/api/admin/ops/workflows/:id/execute
/api/admin/ops/workflows/:id/cancel
/api/admin/ops/audit
```

外部平台调用必须经过后端连接器服务，不能由浏览器直接持有 Hostinger 或 Cloudflare 凭据。

## 10. 实施顺序

### Phase 0：边界和数据模型

- [x] 确认运维中心权限。
- [x] 建立连接器、域名绑定模型。
- [x] 接入连接器凭据加密和审计关联。
- [x] 建立 VPS 绑定和项目绑定模型。
- [x] 建立 VPS 中心和项目中心后台页面，支持声明式台账编辑、启停和审计；VPS/项目列表默认生产环境并支持环境筛选，编辑时保留异步加载中的既有连接器绑定；项目可显式绑定同环境 Hostinger 连接器或沿用 VPS 连接器。
- [x] 登记当前生产 VPS、`commerce-platform` 项目及 Compose、网关、服务、网络、卷和备份边界。
- 明确生产环境初始数据，不自动猜测或覆盖当前资源。

### Phase 1：只读运维总览

- [x] 增加运维总览聚合接口和后台页面，展示 VPS、项目、域名、连接器统计、生产拓扑、待处理项和运维审计记录。
- [x] 为域名台账增加 Desired/Observed 字段和“未同步 / 已匹配 / 漂移 / 检查错误”展示。
- [x] 接入 Hostinger 读取 VPS 和项目，支持后台单项手动同步、Observed State 持久化回写、刷新后展示和审计；该阶段本身仍不执行外部写入。
- [x] 接入 Cloudflare 读取 Zone、DNS、代理、SSL/TLS 状态，并写回域名 Observed State。
- [x] 展示当前部署版本、服务状态、最近检查、远端状态和 Hostinger 容器计数。
- [x] 增加手动同步和状态来源。

### Phase 2：域名中心

- [x] 支持域名、别名、`admin` 域、跳转域和验证域。
- [x] 支持域名期望状态的新增、编辑、启停和审计。
- [x] 实现 Desired/Observed 差异，并保留 `unknown`、`matched`、`drifted`、`error` 边界。
- [x] 生成 DNS、Caddy 和 Nginx 只读预览。
- [x] 提供域名期望 / 实际差异明细接口和后台查看入口。
- [ ] 先提供人工确认后的变更，不做无确认自动切换。

### Phase 3：部署中心与 dry-run

- [x] 展示当前台账版本、配置差异和部署前检查；dry-run 支持填写 SHA/digest 作为发布引用证据，生产工作流当前仍使用 `master` 受控发布策略。
- [x] 持久化部署工作流、步骤状态、人工审批和 dry-run 执行证据。
- [x] 对 dry-run 步骤标记外部副作用边界，并明确不调用生产写入接口。
- [x] 将回滚所需的 `deploy.sh` 固定命令、绑定 VPS 目标和远程操作证据封装为受限 SSH 执行器。
- [ ] 将正常发布所需的 `deploy.sh`、发布镜像检查和 `verify-vps-release-boundary.sh` 封装为可执行的外部步骤。

### Phase 4：生产执行、健康检查和回滚

- [x] 接入 Hostinger 既有 Docker Compose 项目更新；当前只支持 `IMAGE_TAG=master` 策略和后端受控执行。
- [x] 接入发布后 DNS、HTTP、HTTPS 健康检查，失败进入 `rollback_required`。
- [x] 建立发布前回滚点、项目锁、幂等键和失败暂停证据。
- [x] 支持失败、暂停或需回滚工作流从可重试失败步骤继续，保留原回滚点和幂等键，已成功步骤不重跑。
- [x] 接入按项目绑定域名、Cloudflare Zone 和连接器分组的缓存清理；单次请求最多 30 个 host，没有 Cloudflare 域名时显式跳过。
- [x] 建立人工批准后的真实回滚执行器：仅接受完整 Commit SHA，限制绑定 VPS、SSH 用户、私钥文件、`known_hosts` 和固定工作目录，完成后重新同步、健康检查和缓存清理。
- [ ] 完成至少一次恢复演练。

### Phase 5：漂移检测与自动化

- 定时同步实际状态。
- 检测 DNS、网关、项目和镜像漂移。
- 对低风险检查支持定时执行。
- 高风险变更继续保留人工审批。

## 11. 第一阶段验收标准

第一阶段完成前，不得把运维中心描述为“自动部署平台”。至少满足：

- 后台可以查看当前 Hostinger VPS、项目和服务状态。
- 后台可以查看 Cloudflare Zone、域名和关键 DNS 状态。
- 页面能区分期望状态和实际状态。
- 连接器凭据不以明文出现在页面、日志或 API 响应中。
- 可以对当前生产配置执行 Cloudflare 与 Hostinger 只读同步和连接测试。
- 可以生成部署前检查结果和配置差异，但不会未经确认修改生产环境。
- 所有连接测试、同步、dry-run 和失败都进入审计记录。

## 12. 相关文档

- Hostinger 执行手册：`docs/ops/hostinger-vps-docker-runbook.md`
- Cloudflare 与边缘安全：`deployment/EDGE_SECURITY_RUNBOOK.md`
- 生产状态：`docs/ops/production-readiness-status.md`
- VPS 发布边界检查：`deployment/verify-vps-release-boundary.sh`
- 已完成的域名切换记录：`docs/archive/ops/learn-gripe-cutover-completed.md`
- 已完成的品牌连续性记录：`docs/archive/ops/brand-rename-continuity-plan-completed.md`

最后更新：2026-08-16。

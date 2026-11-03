# 业务模块导出、正式恢复可写范围与共享依赖交叉索引

核对日期：2026-11-02；ARCH-07 基线见主架构 4.73 / 阶段 3.66，媒体范围见 4.74–4.75，配置内容及适配目录补缺见 4.76–4.81 / 阶段 3.69–3.74。本文说明当前本地源码，正式操作按[生命周期导航](backup-recovery-lifecycle-supported-entrypoints-internal-rehearsals-and-persisted-task-compatibility.zh-cn.md)和[运行手册](../../../docs/design/ops/local-computer-whole-storefront-backup-upload-validation-and-paired-recovery-runbook.zh-cn.md)。未发布服务器；各模块只按自身专属契约写入。

## 1. 三种范围各自回答什么问题

- **导出包含**：哪些表的哪些完整行进入加密包，以及附带哪些实际文件。来源是[六模块选择器](selected_business_module_backup_scope_registry_and_range_preview.go)、[完整行及引用媒体采集](selected_business_module_complete_row_snapshot_and_referenced_media_capture.go)和商品专属采集。进入包不等于允许覆盖目标。
- **正式恢复可写**：哪个领域契约允许补入或替换哪些行。来源是商品合并、订单逐行资格、客户补缺、聊天四表策略、独立媒体及配置 53 表内容/目录/流程/问卷/市场/汇率/税率/模板/会员等级/物流基础资料/包装/报关补缺契约。必须经过原包认证、目标绑定、领域准入及各自执行流程；不能只查表名。
- **共享保留与依赖核对**：不能随所选模块重建的当前身份、媒体、库存、配置及执行事实。实际依赖来自目标数据库外键、显式业务/JSON 引用和原包摘要；[旧依赖登记](versioned_order_and_customer_service_module_dependency_registry.go)的 `OwnedTables` / `SharedTables` 只描述诊断分类，不是完整的可写/只读授权表。

三者允许交叉：订单导出包含 `coupon_usage` 和 `outbox_events`，旧登记也把它们列为共享；聊天导出的坐席、策略、上传收据及事件在正式聊天事务中保持。不能为了“统一清单”删除这些差别，或给所有导出表赋写权限。

## 2. 六模块能力和责任来源

“直接导出表数”只统计单选时选择器的键；除独立媒体外，还会按引用附带 `media_assets` / `media_asset_derivatives` 及原件/派生文件。它不是最终非空表数、实际行数或可写表数，最终以包清单为准。

| 模块 | 直接导出范围 | 本地已实现的正式写入 | 共享保留及责任来源 |
| --- | --- | --- | --- |
| 商品 `products` | 24 张商品表；另带引用的两类媒体登记、图片和同版 company 文件。单选自动生成 `product_catalog_with_images` | 商品专属候选合并、持续停写切换及正式接收；不能把普通组合包直接送到商品入口 | [商品表契约](product_catalog_module_explicit_table_contract_and_bounded_json_snapshot.go)、[库存保留](product_catalog_current_inventory_preservation_stable_sku_identity_and_shared_inventory_owner_validation.go)；保留当前已有 SKU 库存、浏览量和未选业务，核验共享媒体/静态文件及外部引用 |
| 订单 `orders` | 45 张：旧登记 42 张，加优惠使用、发送记录和限定事件 | [订单缺失记录资格](order_module_missing_order_restore_eligibility_current_execution_scope_and_insert_guard_contract.go)批准的订单/明细/证据；[已完成事实](order_module_completed_financial_facts_processed_events_and_current_note_preservation_contract.go)批准的原交易/退款/幂等/发货/事件/通知收据，以及三种已完成全局支付收据；由[正常约束 writer](order_module_missing_rows_normal_constraint_dependency_ordered_transactional_writer.go)执行 | 可导出售后/争议/风控不代表缺失时能补入；共享客户、商品/组件/主库存、优惠积分及推荐事实须兼容，保留当前库存及符合条件的较新备注/预警状态；不重新支付、退款、发货或发推荐注册奖励 |
| 客户 `customers` | 12 张客户表；`users` 仅 `user` / `test_user`，其他直接选择器为全表；另带实际引用媒体 | [客户补缺契约及 writer](customer_profile_module_missing_account_cohort_financial_fact_and_existing_identity_preservation_contract.go)的同一组 12 张表，有明确依赖顺序；写入仍受缺失身份及完整缺失账号组资格约束 | 当前身份、密码、限制、积分、退订和较新资料保留；订单、聊天、媒体登记及执行事实不能因客户包存在就补写。推荐奖励仅复制已完成原事实，不重新发放 |
| 聊天 `customer_service` | 11 张：旧登记 10 张加限定事件；会话/消息/收件箱/附件按客服类别过滤，上传收据按业务上下文过滤 | [保存策略的完整行 writer](selected_customer_service_module_snapshot_bound_saved_restore_strategy_and_complete_row_candidate_rehearsal.go)只写 `tickets`、`ticket_messages`、`customer_service_inbox_states`、`ticket_message_media_attachments`；两种策略均保留目标独有记录 | [策略及依赖投影](selected_customer_service_module_snapshot_bounded_overwrite_strategies_and_preserved_event_contract_preview.go)和[正式事务](customer_service_module_formal_restore_saved_strategy_atomic_rows_audit_and_recoverable_attachment_installation.go)要求共享身份、上传/媒体登记、原事件和配置兼容；旧持久候选不属于当前正式写入 |
| 图片附件 `media` | 11 张媒体/派生/收据/绑定/标识表及全部实际媒体文件，包含未登记文件 | [媒体资格](media_module_standalone_restore_eligibility_completed_uploads_shared_owners_and_cleanup_hazard_inspection.go)批准的缺失行，由[正常约束 writer](media_module_missing_registration_dependency_ordered_normal_constraint_writer_and_retained_table_verification.go)按依赖顺序插入；[正式执行及中断核验](media_module_formal_restore_atomic_registration_files_audit_and_verified_interruption_recovery.go) | 只支持独立媒体包，保留已有完整行；登记与原包一致时可按 v2 计划修复文件字节并保留原链接至提交核验；不写订单、消息、用户、设置、事件。已绑定上传/完成任务才准入；[原件、派生及私有业务身份](media_module_restored_original_derivative_file_identity_and_private_conversation_binding_validation.go)独立核验，company 须同版已部署 |
| 业务配置 `configuration` | 79 张站点内容、市场税率、交易/邮件/物流、选择流程及适配目录等表；另带引用图片。完整名单见六模块选择器 | [网站内容资格及正常约束 writer](configuration_module_missing_site_content_scope_public_file_prerequisites_and_normal_constraint_writer.go)只补 `categories`、`posts`、`post_categories`、`galleries`、`gallery_images`、`gallery_product_links`、`faq_pages`、`faqs`、`visual_showcase_items`、`storefront_redirect_rules` 十表及[适配/辐条/轮胎目录](configuration_module_fitment_spoke_and_tire_catalog_missing_row_scope_and_business_relationship_validation.go)十二表及[选购流程/问卷范围与业务验证](configuration_module_selection_flow_and_questionnaire_missing_rows_affected_published_versions_and_json_reference_validation.go)十七表及[市场/精确汇率/税率范围与关系保护](configuration_module_market_currency_and_tax_missing_rows_exact_decimal_and_effective_rule_preservation_validation.go)六表及[邮件模板/完整历史与会员等级](configuration_module_email_template_history_and_member_level_missing_rows_renderability_version_chain_and_range_validation.go)三表及[物流基础资料、包装归属与报关校验](configuration_module_shipping_reference_packaging_product_variant_and_customs_missing_rows_validation.go)五表的缺失行；[正式事务/中断](configuration_module_missing_site_content_atomic_restore_audit_and_verified_interruption_recovery.go)提交后保持维护，经[人工核对及释放](configuration_module_site_content_projection_review_bound_service_release_and_http_routes.go)恢复服务 | 其余 26 张配置表及媒体登记只作兼容前置，缺失/变化阻断；已有内容不覆盖，目标新增行保留。包内全部图片及同版 company 文件须已匹配，入口不安装文件，不引入私有/客户/访客/会话素材。完整配置、密钥生效及图片联动仍属 DELIVERY-02 |

整站另走完整库和完整媒体路径，不等于六模块全部勾选。多选导出共用一次快照，仍是 `selected_business_modules`；组合导入和整站包子范围导入属于 DELIVERY-03，不能继承任一单模块资格。

快速购买旧步骤翻译表已由[迁移 123](../../migrations/123_unify_quick_buy_flow_help_text.up.sql)合并至 `quick_buy_flow_translations` 并删除，现行导出包含流程翻译；不能把旧表未列入选择器误认成导出缺口或重新建表。选购流程/问卷已按 4.78 接入补缺；市场/汇率/税率按 4.79 接入，不包括 settings 或同步控制。

## 3. 已核对的集合差异和行过滤

| 比较 | 当前结果 | 必须保留的含义 |
| --- | --- | --- |
| 客户直接导出 vs 客户 writer | 12 张表的集合相同，顺序各有职责 | 集合相同不能免除角色过滤、身份核对、完整账号组和当前余额保护 |
| 聊天导出 vs 四表可替换范围 | 另有 7 张：`media_upload_sessions`、`customer_service_retention_policies`、`ticket_auto_replies`、`customer_service_agent_profiles`、`customer_service_agent_groups`、`customer_service_agent_group_members`、`outbox_events` | 这些完整行随包保存，但当前聊天正式事务不写它们；引用媒体登记另行采集，也不是聊天可写表 |
| 订单导出 vs 旧登记 OwnedTables | 多 3 张：`coupon_usage`、`email_delivery_records`、`outbox_events` | 优惠使用保留；后两者只有原归属和已完成收据契约才可补入，不能恢复为待执行事件 |
| 配置直接导出 vs 配置内容 writer | 79 张直接导出，只有上述 53 表可补缺；其余 26 表只读兼容，引用媒体登记也只读 | 不授予 settings、支付/邮件/物流密文或运行控制写权限；其他随包行缺失/变化仍会阻断，不是保留变化后忽略。媒体文件已存在且同字节才可恢复 |
| 商品/媒体导出交叉 | `product_media` | 全媒体包含商品关联不赋予商品数据覆盖权限 |
| 聊天/媒体导出交叉 | `media_upload_sessions`、`ticket_message_media_attachments` | 单选聊天受客服行条件限制；显式选择全媒体时这两表扩大到全部行 |
| 配置/媒体、配置/订单交叉 | `site_logo_assets`、`site_favicon_assets`；`payment_protection_controls` | 是共享范围，组合导入仍须合并一次并核验原身份，不能按两个模块各自覆盖 |

组合选择保持客户账号角色和客服工单类别过滤；订单+聊天的 Outbox 条件按事件族取并集。最终引用媒体来自实际行/文件引用，并非无条件导出所有客户或其他模块数据。外部依赖只记录身份和值摘要；已随包包含的目标行不重复记为外部依赖。多态 `loyalty_transactions.source_id` 仅在 `source=order` 时视为订单引用，注册推荐和签到不能误归订单。

表级索引也不替代完整外键、业务快照、JSON 引用和文件权限检查；未登记表仍拒绝。`audit_logs`、迁移元数据、浏览量幂等收据和运维平台表不进入这些模块选择器；正式恢复新写的同事务审计收据属于恢复行为，不是从包覆盖旧审计。

## 4. 可重复的差异检查

[范围交叉回归](business_module_export_writable_restore_and_read_only_shared_dependency_scope_cross_checks_test.go)比较实际选择器、客户 writer 集合、聊天逐行替换准入和旧依赖登记，并调用订单行级资格检查共享表补入阻断：

- 六模块全部 63 种非空组合：表集合只合并显式选择，关键账号/会话过滤保持，全媒体扩大范围仅限已登记交叉表，Outbox 条件仍限定事件族。
- 客户导出与 12 表 writer 的集合对齐，媒体/订单/聊天/事件不会获得客户补缺资格。
- 聊天四表准入与 11 表导出/旧归属登记对照，七张只读随包表及共享登记不能获得会话替换资格。
- 订单导出的三项额外表和共享登记交叉明确；除有专属已完成契约的 Outbox 外，其余 22 张共享表即使包中缺失，也被实际订单资格检查拒绝。

本批四项测试、85 个子场景通过；无数据库连接，不能替代数据库事务、真实文件或页面验收。聊天 writer 和依赖投影中的实际四表字面量另作本批静态核对；测试不是新的写入授权来源。生产源码仅补三处职责注释，算法、策略、包/任务格式和已保存收据保持。证据及既有验证限制见[阶段登记 3.66](../../../docs/design/ops/local-computer-backup-roadmap-stage-alignment-and-target-environment-acceptance-register-2026-10-22.zh-cn.md)。

DELIVERY-01 已接通媒体补缺及登记一致文件修复；DELIVERY-02 已接通内容、目录、流程/问卷、市场/汇率/税率及邮件模板/历史和会员等级补缺，仍须其余配置补入/覆盖、图片联动及生效契约；组合按 DELIVERY-03 推进。六条有界路径均待发布与实机验收，媒体复杂冲突单独保留。新增表或能力须同步领域契约、交叉检查和状态，不能仅修改导出表单或把旧候选待办作为统一前置。

市场/汇率/税率新增六表为 storefront_markets、storefront_market_countries、currency_exchange_rates、tax_rates、tax_rate_source_snapshots、tax_rate_source_snapshot_entries。settings、tax_rate_source_configs 仍只读；currency_exchange_sync_leases 是运行租约，不在配置导出选择器中。恢复不延长汇率有效期、不联网同步、不重算旧订单；参考快照不变为结算规则，来源摘要原样保留而非重新认证外部内容。

## 5. DELIVERY-02 剩余 26 张直接配置表

以下是实际 79 表选择器减去当前 53 表 writer 的差集；只读兼容表示随包保存但本入口不补写，不能由此判断是废表。当前目标有缺失或内容变化仍阻断，不会静默跳过。

| 职责组 | 张数 | 尚无正式补缺契约的表 |
| --- | --- | --- |
| 通用配置、客户展示和站点素材登记 | 4 | `settings`、`showcases`、`site_logo_assets`、`site_favicon_assets` |
| 外部税率采集控制 | 1 | `tax_rate_source_configs` |
| 支付及第三方登录 | 3 | `payment_methods`、`payment_protection_controls`、`social_oauth_connections` |
| 邮件发送方配置 | 1 | `email_provider_configs` |
| 推荐、积分及优惠规则 | 3 | `loyalty_program_configs`、`referral_program_configs`、`coupons` |
| 运费模板、线路聚合及外部渠道 | 14 | `shipping_templates`、`shipping_rules`、`shipping_carrier_services`、`shipping_route_groups`、`shipping_route_group_carrier_services`、`shipping_fpx_api_configs`、`shipping_fpx_channels`、`shipping_yanwen_api_configs`、`shipping_yanwen_country_catalog_entries`、`shipping_yanwen_product_catalog_entries`、`shipping_yanwen_published_channels`、`shipping_yanwen_warehouse_catalog_entries`、`shipping_tracking_carrier_mappings`、`shipping_tracking_providers` |

邮件模板两表与会员等级已按 4.80 接入补缺，不恢复邮件提供者、账户积分或原奖励。物流基础资料、包装和报关五表已按 4.81 接入补缺；下一批处理运费模板/计费规则与已发布线路聚合，密文、开关、外部账号、积分/推荐计划生效和物流运行任务仍需各自契约。已有覆盖、图片安装联动与组合导入也不因差集缩小而自动完成。

showcases 实际由客户展示领域及仓储使用，关联用户和订单；不是废表。本批更正原分类文字，导出分组与正式恢复另行收敛，未删表或改变选择器。邮件模板恢复校验完整连续版本链、当前快照及纯渲染；会员等级恢复校验精确折扣和最终区间，不产生邮件、积分或推荐奖励。

物流本批五表为 carriers、shipping_zones、customs_classification_profiles、shipping_packaging_rules、shipping_packaging_rule_applies。包装商品/SKU原指纹只读，缺失/变化/已删除/归属错误阻断；旧包有绑定但缺指纹须重新导出或使用合格整站包，重演练不会补造。原尺寸、重量、启用/停用、报关风险/来源/核验日期保留，不联网同步、不发货/购买面单、不改旧订单。运费模板、服务/线路、外部发布渠道及凭据仍只读；正常写入、审计和核对释放沿原入口。

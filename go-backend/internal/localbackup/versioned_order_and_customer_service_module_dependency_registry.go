package localbackup

import "strings"

// This registry describes table boundaries, not permission to restore a module.
// Row ownership in shared tables, JSON references and file retention still need
// a separate executable contract before any partial replacement is enabled.
const moduleDependencyRegistryVersion = 8
const moduleDependencyRegistryMigrationVersion int64 = 434

// Migration 435 reconciles product descriptions; 436 adds a preserved view flush
// receipt table, outside every module-owned set. Keep the baseline in existing
// plan/journal serialization and explicitly audit each additional version.
func moduleDependencyRegistrySupportsAuditedMigrationVersion(version int64) bool {
	return version == moduleDependencyRegistryMigrationVersion || version == 435 || version == 436
}

type moduleRecoveryDependencyContract struct {
	Name           string
	Label          string
	OwnedTables    []string
	SharedTables   []string
	ReviewMessages []string
}

func registeredOrderAndCustomerServiceRecoveryDependencyContracts() map[string]moduleRecoveryDependencyContract {
	return map[string]moduleRecoveryDependencyContract{
		"orders": {
			Name: "orders", Label: "订单交易",
			OwnedTables: strings.Fields(`
after_sales_case_attachments after_sales_case_events after_sales_case_events_archive
after_sales_case_items after_sales_cases after_sales_refund_reviews after_sales_return_shipments
order_attributions order_evidence_attachments order_evidence_export_snapshots order_evidence_items
order_evidence_packages order_evidence_snapshots order_evidence_submission_snapshots
order_idempotencies order_items order_policy_disclosures orders
payment_operation_idempotencies payment_protection_controls payment_refund_executions
payment_refund_recommendations payment_reviews payment_risk_alert_states
payment_risk_checkout_decisions payment_risk_events payment_risk_snapshots
paypal_disputes refund_idempotencies refund_line_items refunds shipment_records
shipping_quote_snapshots shipping_tracking_shipments shipping_yanwen_tracking_snapshots
shipping_yanwen_waybills stripe_disputes stripe_webhook_events tracking_events transactions
warranty_claims warranty_service_records`),
			SharedTables: strings.Fields(`users visitor_profiles products product_variants
media_assets media_asset_derivatives outbox_events coupons coupon_usage loyalty_transactions
user_loyalty referral_records referral_reward_logs referral_transitions
referrals user_referral_identities payment_methods settings currency_exchange_rates
shipping_templates shipping_rules shipping_zones tax_rates`),
			ReviewMessages: []string{
				"库存、优惠、积分及推荐奖励与现存订单共用事实，须核对所选订单对应的行和增减记录。",
				"支付、退款、物流和通知的外部结果须独立对账；事件状态与幂等键不能随局部恢复重放。",
				"证据快照中的 JSON、文本和媒体引用须登记对象归属，并验证原件与访问权限。",
			},
		},
		"customer_service": {
			Name: "customer_service", Label: "客服对话",
			OwnedTables: strings.Fields(`tickets ticket_messages ticket_message_media_attachments
media_upload_sessions customer_service_inbox_states customer_service_retention_policies
ticket_auto_replies customer_service_agent_profiles customer_service_agent_groups
customer_service_agent_group_members`),
			SharedTables: strings.Fields(`users visitor_profiles media_assets media_asset_derivatives
media_derivative_rebuild_jobs outbox_events settings`),
			ReviewMessages: []string{
				"tickets 同时承载多类工单，须明确客服会话筛选及关联消息、收件箱、上传收据的行边界。",
				"客户、访客哈希与签名密钥须保持身份连续性；软删除及保留窗口内的记录仍属于恢复范围。",
				"历史附件 JSON、私有原件、派生文件、复用素材与清理租约须逐对象核对，不能清空共享媒体表。",
				"客服与订单共用 Outbox，须按事件及聚合身份划分范围并保留其他业务的状态和幂等键。",
			},
		},
	}
}

// Explicit names only. A new table is unregistered even if its name looks like
// a familiar module; empty tables and tables without foreign keys also count.
func registeredStorefrontModuleRecoveryTableNames() map[string]bool {
	names := strings.Fields(`
schema_migrations after_sales_case_attachments after_sales_case_events after_sales_case_events_archive
after_sales_case_items after_sales_cases after_sales_refund_reviews after_sales_return_shipments
audit_logs browsing_history business_draft_media_upload_receipts carriers cart_items carts
categories check_ins comments coupon_usage coupons currency_exchange_rates currency_exchange_sync_leases
customer_service_agent_group_members customer_service_agent_groups customer_service_agent_profiles
customer_service_inbox_states customer_service_retention_policies customs_classification_profiles
email_delivery_records email_provider_configs email_template_versions email_templates email_verification_challenges
faq_pages faqs feedback fitment_fork_entries fitment_fork_hub_specifications fitment_frame_entries
fitment_frame_hub_specifications fitment_hub_specifications galleries gallery_images gallery_product_links
global_ip_block_rules google_merchant_connections google_merchant_offers
loyalty_program_configs loyalty_transactions media media_asset_derivatives
media_assets media_derivative_presets media_derivative_rebuild_jobs media_upload_sessions member_levels
ops_connector_oauth_sessions ops_connectors ops_deployment_workflow_locks ops_deployment_workflow_runs
ops_deployment_workflow_steps ops_domain_bindings ops_network_rules ops_project_bindings ops_vps_bindings
order_attributions order_evidence_attachments order_evidence_export_snapshots order_evidence_items
order_evidence_packages order_evidence_snapshots order_evidence_submission_snapshots order_idempotencies
order_items order_policy_disclosures orders outbox_events payment_methods payment_operation_idempotencies
payment_protection_controls payment_refund_executions payment_refund_recommendations payment_reviews
payment_risk_alert_states payment_risk_checkout_decisions payment_risk_events payment_risk_snapshots
paypal_disputes post_categories posts preflight_content_link_issue_events preflight_content_link_issues
preflight_content_link_runs product_attribute_values product_attributes product_brands product_categories
product_category_translations product_custom_option_policies product_display_price_snapshots
product_information_templates product_media product_option_group_variant_rules product_option_value_relations
product_option_value_variant_rules product_procurement_records product_profit_calculations
product_quality_requirement_rules product_shipping_template_bindings product_spec_definitions
product_spec_option_items product_spec_values product_specification_templates product_spoke_repair_kit_models
product_variant_option_values product_variants product_view_count_flush_receipts products quick_buy_compatibility_rules quick_buy_flow_translations
quick_buy_flow_versions quick_buy_flows quick_buy_session_items quick_buy_sessions quick_buy_step_filters
quick_buy_step_product_categories quick_buy_step_product_specification_templates quick_buy_steps
recommendation_events referral_program_configs referral_records referral_reward_logs referral_transitions
referrals refund_idempotencies refund_line_items refunds review_helpful review_summaries reviews
schwalbe_tire_hookless_compatibility schwalbe_tire_specifications selection_assistant_flow_versions
selection_assistant_flows selection_configuration_keys settings shipment_records shipping_carrier_services
shipping_display_price_snapshots shipping_fpx_api_configs shipping_fpx_channels shipping_packaging_rule_applies
shipping_packaging_rules shipping_quote_snapshots shipping_route_group_carrier_services shipping_route_groups
shipping_rules shipping_templates shipping_tracking_carrier_mappings shipping_tracking_providers
shipping_tracking_shipments shipping_yanwen_api_configs shipping_yanwen_country_catalog_entries
shipping_yanwen_product_catalog_entries shipping_yanwen_published_channels shipping_yanwen_tracking_snapshots
shipping_yanwen_warehouse_catalog_entries shipping_yanwen_waybills shipping_zones showcases site_favicon_assets
site_logo_assets site_quality_evaluations site_quality_finding_events site_quality_findings site_quality_jobs
site_quality_provider_slots site_quality_runs site_quality_runs_archive site_quality_targets
social_oauth_connections social_oauth_sessions spoke_build_presets spoke_histories spoke_hub_brands
spoke_hub_models spoke_rim_brands spoke_rim_models storefront_market_countries storefront_markets
storefront_redirect_rules storefront_route_catalog_entries storefront_route_check_results
storefront_url_issue_events storefront_url_issues storefront_url_search_profiles stripe_disputes
stripe_webhook_events subscriptions suggestion_feedback tax_rate_source_configs tax_rate_source_snapshot_entries
tax_rate_source_snapshots tax_rates ticket_auto_replies ticket_message_media_attachments ticket_messages
tickets tracking_events transactions user_loyalty user_referral_identities users visitor_profiles
visitor_risk_daily_facts visitor_risk_decisions visual_showcase_items warranty_claims warranty_service_records
wheelset_fit_question_option_translations wheelset_fit_question_options wheelset_fit_question_translations
wheelset_fit_questionnaire_versions wheelset_fit_questionnaires wheelset_fit_questions wishlist_items
workbench_feed_entries workbench_feed_media workbench_feed_tagged_products`)
	registered := make(map[string]bool, len(names))
	for _, name := range names {
		registered[name] = true
	}
	return registered
}

type ModuleRecoveryDependencyReference struct {
	SourceTable   string   `json:"source_table"`
	SourceColumns []string `json:"source_columns"`
	TargetTable   string   `json:"target_table"`
	TargetColumns []string `json:"target_columns"`
	Kind          string   `json:"kind"`
	Description   string   `json:"description"`
}

func registeredNonForeignKeyModuleRecoveryReferences() []ModuleRecoveryDependencyReference {
	var references []ModuleRecoveryDependencyReference
	for _, entry := range [][4]string{
		{"visitor_profiles", "user_id", "users", "id"},
		{"tickets", "visitor_session_hash", "visitor_profiles", "customer_service_visitor_hash"},
		{"customer_service_inbox_states", "last_read_message_id", "ticket_messages", "id"},
		{"media_upload_sessions", "upload_actor_account_id", "users", "id"},
		{"media_upload_sessions", "upload_actor_visitor_session_hash", "visitor_profiles", "customer_service_visitor_hash"},
		{"order_items", "product_id", "products", "id"},
		{"order_items", "variant_id", "product_variants", "id"},
		{"order_attributions", "order_id", "orders", "id"},
		{"tracking_events", "order_id", "orders", "id"},
	} {
		references = append(references, ModuleRecoveryDependencyReference{SourceTable: entry[0], SourceColumns: []string{entry[1]}, TargetTable: entry[2], TargetColumns: []string{entry[3]}, Kind: "registered_business_reference", Description: "业务引用；有效值、空值与行归属仍须按对应业务规则核对"})
	}
	return references
}

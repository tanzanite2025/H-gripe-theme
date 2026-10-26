package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const selectedBusinessModulesBackupScope = "selected_business_modules"
const selectedBusinessModulesRowsFilename = "selected-business-module-complete-rows.json"

type BusinessModuleBackupOption struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// These are export contracts. Availability here never authorizes online restore.
func availableBusinessModuleBackupOptions() []BusinessModuleBackupOption {
	return []BusinessModuleBackupOption{
		{"products", "商品与库存", "商品详情、多语言、SKU、规格、分类、价格库存及实际图片"},
		{"media", "图片与附件", "全部三类媒体原件、派生文件、素材登记及同版 company 静态文件；含私有附件"},
		{"orders", "订单与交易", "订单明细、支付退款、发货售后、证据原件及对应事件状态"},
		{"customers", "客户资料", "商城客户、访客、积分与推荐事实；不重发推荐奖励"},
		{"customer_service", "客户聊天记录", "客服会话、全部消息、阅读位置、归属、实际附件及对应事件状态"},
		{"configuration", "商城业务配置", "站点内容、市场税率、交易规则、邮件支付等数据库配置及相关图片"},
	}
}

func normalizeSelectedBusinessBackupModules(modules []string) ([]string, error) {
	if len(modules) == 0 || len(modules) > len(availableBusinessModuleBackupOptions()) {
		return nil, errors.New("请选择至少一个业务模块")
	}
	allowed, selected := map[string]bool{}, map[string]bool{}
	for _, option := range availableBusinessModuleBackupOptions() {
		allowed[option.ID] = true
	}
	for _, module := range modules {
		if !allowed[module] || selected[module] {
			return nil, errors.New("备份模块不受支持或重复，请重新选择")
		}
		selected[module] = true
	}
	result := append([]string(nil), modules...)
	sort.Strings(result)
	return result, nil
}

// Every selector is an internal literal. HTTP clients cannot supply table names,
// SQL, predicates, storage roots, or arbitrary package paths.
func selectedBusinessModuleTableSelectors(modules []string) map[string]string {
	selectors := map[string]string{}
	includeTables := func(names []string) {
		for _, name := range names {
			selectors[name] = "true"
		}
	}
	eventPredicates := []string{}
	for _, module := range modules {
		switch module {
		case "products":
			includeTables(productCatalogOwnedTableNames)
		case "orders":
			includeTables(registeredOrderAndCustomerServiceRecoveryDependencyContracts()["orders"].OwnedTables)
			includeTables(strings.Fields("coupon_usage email_delivery_records"))
			selectors["email_delivery_records"] = `r.reference_type IN ('order','payment','refund','after_sales_case') OR EXISTS(SELECT 1 FROM outbox_events e WHERE e.id=r.outbox_event_id AND e.aggregate_type IN ('order','payment','after_sales_case'))`
			eventPredicates = append(eventPredicates, `r.aggregate_type IN ('order','payment','after_sales_case')`)
		case "customers":
			includeTables(strings.Fields(`users visitor_profiles user_loyalty loyalty_transactions check_ins
user_referral_identities referral_records referral_reward_logs referral_transitions referrals
subscriptions wishlist_items`))
			selectors["users"] = `r.role IN ('user','test_user')`
		case "customer_service":
			includeTables(registeredOrderAndCustomerServiceRecoveryDependencyContracts()["customer_service"].OwnedTables)
			selectors["tickets"] = `r.category='customer_service'`
			selectors["ticket_messages"] = `EXISTS(SELECT 1 FROM tickets t WHERE t.id=r.ticket_id AND t.category='customer_service')`
			selectors["customer_service_inbox_states"] = selectors["ticket_messages"]
			selectors["ticket_message_media_attachments"] = `EXISTS(SELECT 1 FROM ticket_messages m JOIN tickets t ON t.id=m.ticket_id WHERE m.id=r.ticket_message_id AND t.category='customer_service')`
			selectors["media_upload_sessions"] = `r.business_context_type='customer_service'`
			eventPredicates = append(eventPredicates, `r.aggregate_type IN ('customer_service_conversation','customer_service_agent_profile') OR r.event_type IN ('customer_service.realtime','customer_service.avatar_cleanup','customer_service.retention_cleanup_requested')`)
		case "media":
			includeTables(strings.Fields(`media media_assets media_asset_derivatives media_derivative_presets
media_derivative_rebuild_jobs media_upload_sessions business_draft_media_upload_receipts
product_media ticket_message_media_attachments site_logo_assets site_favicon_assets`))
		case "configuration":
			includeTables(strings.Fields(`settings posts post_categories categories galleries gallery_images gallery_product_links
faqs faq_pages showcases visual_showcase_items storefront_markets storefront_market_countries
currency_exchange_rates tax_rates tax_rate_source_configs tax_rate_source_snapshots tax_rate_source_snapshot_entries
payment_methods payment_protection_controls email_provider_configs email_templates email_template_versions social_oauth_connections
loyalty_program_configs member_levels referral_program_configs coupons carriers customs_classification_profiles
shipping_templates shipping_rules shipping_zones shipping_carrier_services shipping_packaging_rules
shipping_packaging_rule_applies shipping_route_groups shipping_route_group_carrier_services shipping_fpx_api_configs
shipping_fpx_channels shipping_yanwen_api_configs shipping_yanwen_country_catalog_entries
shipping_yanwen_product_catalog_entries shipping_yanwen_published_channels shipping_yanwen_warehouse_catalog_entries
shipping_tracking_carrier_mappings shipping_tracking_providers storefront_redirect_rules site_logo_assets site_favicon_assets
selection_assistant_flows selection_assistant_flow_versions selection_configuration_keys quick_buy_flows
quick_buy_flow_versions quick_buy_flow_translations quick_buy_steps quick_buy_step_filters
quick_buy_step_product_categories quick_buy_step_product_specification_templates quick_buy_compatibility_rules
wheelset_fit_questionnaires wheelset_fit_questionnaire_versions wheelset_fit_questions wheelset_fit_question_translations
wheelset_fit_question_options wheelset_fit_question_option_translations spoke_build_presets spoke_hub_brands
spoke_hub_models spoke_rim_brands spoke_rim_models fitment_fork_entries fitment_fork_hub_specifications
fitment_frame_entries fitment_frame_hub_specifications fitment_hub_specifications
schwalbe_tire_hookless_compatibility schwalbe_tire_specifications`))
		}
	}
	// A broader explicit choice wins over a narrower shared-table selector.
	for _, module := range modules {
		if module == "media" {
			selectors["media_upload_sessions"] = "true"
			selectors["ticket_message_media_attachments"] = "true"
		}
	}
	if len(eventPredicates) > 0 {
		selectors["outbox_events"] = "(" + strings.Join(eventPredicates, ") OR (") + ")"
	}
	return selectors
}

type SelectedBusinessModuleBackupPreview struct {
	Modules                  []string         `json:"modules"`
	Tables                   map[string]int64 `json:"tables"`
	IncludedDependencies     []string         `json:"included_dependencies"`
	RecoveryLimitations      []string         `json:"recovery_limitations"`
	OnlineOverwriteAvailable bool             `json:"online_overwrite_available"`
}

func (manager *Manager) PreviewSelectedBusinessModuleBackup(ctx context.Context, modules []string) (*SelectedBusinessModuleBackupPreview, error) {
	modules, err := normalizeSelectedBusinessBackupModules(modules)
	if err != nil {
		return nil, err
	}
	db, err := manager.Configuration.connectDatabase(ctx, manager.Configuration.DatabaseName)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(db.Close)
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	snapshot, err := manager.Configuration.captureSelectedBusinessModuleCompleteSnapshot(ctx, tx, modules)
	if err != nil {
		return nil, err
	}
	preview := &SelectedBusinessModuleBackupPreview{
		Modules: modules, Tables: map[string]int64{},
		IncludedDependencies: []string{"按所选记录附带引用的素材登记、派生图和实际文件；选择图片与附件时保存全部媒体。", "未选择的共享客户、商品等只记录依赖身份及内容摘要，不导出其完整资料。"},
		RecoveryLimitations:  []string{"分模块组合包支持下载、回传及内容完整性校验，正式覆盖导入尚未开放。", "恢复到全新服务器请使用整站包；模块数据依赖兼容版本及共享身份，不能当作整站恢复。", "以下数量为预览时快照；最终数量和附带媒体以生成后的包清单为准。"},
	}
	for table, contents := range snapshot.Tables {
		preview.Tables[table] = int64(len(contents.Rows))
	}
	preview.IncludedDependencies = append(preview.IncludedDependencies, fmt.Sprintf("本次引用实际文件 %d 个；外部共享身份依赖摘要 %d 项。选择全部媒体时还会保存媒体目录中其余文件。", len(snapshot.RequiredFilePaths), len(snapshot.ExternalDependencies)))
	return preview, tx.Commit()
}

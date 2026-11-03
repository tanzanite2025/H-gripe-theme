package localbackup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrationpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func TestPostgreSQLConfigurationModuleBusinessContentRulesAndReferenceDataRoundTripAtomicRecoveryAndReviewedRelease(t *testing.T) {
	connectionURL, migrationsDirectory := os.Getenv("LOCAL_BACKUP_INTEGRATION_DATABASE_URL"), os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY")
	if connectionURL == "" || migrationsDirectory == "" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires disposable PostgreSQL and complete application migrations")
	}
	for _, scenario := range []string{"authenticated_restore", "restore_intent_persisted", "rows_and_sequences_before_commit", "rows_committed_before_acknowledgement", "stale_database_after_preview", "missing_picture", "missing_setting", "different_existing_content", "review_result_changed", "disabled_foreign_key_guard", "wrong_fork_hub_position", "wrong_frame_hub_position", "disabled_associated_hub", "enabled_fork_without_hub", "wrong_preset_rim_brand", "wrong_preset_wheel_position", "invalid_catalog_domain_field", "missing_catalog_product_brand", "catalog_natural_identity_conflict", "changed_existing_catalog", "legacy_policy_requires_rehearsal", "legacy_v2_policy_requires_rehearsal", "partial_questionnaire_source_translation", "partial_quickbuy_rule", "selection_missing_node", "selection_cycle", "selection_wrong_json_type", "selection_unknown_filter", "quickbuy_missing_rule_step", "quickbuy_invalid_selection_bounds", "questionnaire_missing_source_label", "questionnaire_unregistered_key", "questionnaire_disabled_key", "questionnaire_invalid_effects_type", "questionnaire_invalid_filter_values", "questionnaire_no_enabled_questions", "invalid_selection_key_kind", "partial_questionnaire_child_invalid_published_parent", "partial_quickbuy_rule_invalid_published_parent", "quickbuy_disabled_product_template", "selection_disabled_filter_definition", "quickbuy_missing_product_template", "quickbuy_second_published_version", "selection_second_published_version", "questionnaire_second_published_version", "selection_changed_existing_configuration", "partial_market_country_restore", "partial_reference_entry_restore", "legacy_reference_comments_not_collected", "market_invalid_locale", "market_default_currency_not_selectable", "market_country_wrong_case", "market_country_soft_deleted_parent", "exchange_unknown_currency", "exchange_invalid_expiry", "tax_unsupported_compound", "tax_invalid_location", "reference_count_mismatch", "reference_invalid_comment_shape", "reference_invalid_entry_type", "reference_missing_entry", "market_country_current_owner_conflict", "tax_newer_current_scope_conflict", "exchange_current_reverse_pair_conflict", "changed_existing_exchange_rate", "missing_tax_sync_configuration", "partial_reference_target_only_extra_entry", "legacy_v3_policy_requires_rehearsal", "partial_template_history_restore", "partial_email_current_version_restore", "disabled_email_template_restore", "email_unknown_template_code", "email_invalid_locale", "email_category_mismatch", "email_variables_wrong_type", "email_unknown_placeholder", "email_subject_header_injection", "email_history_wrong_identity", "email_history_missing_version", "email_current_snapshot_mismatch", "email_history_future_version", "partial_email_history_invalid_parent", "member_level_invalid_name", "member_level_invalid_range", "member_level_package_overlap", "member_level_current_overlap", "changed_existing_email_template", "partial_template_target_only_future_history", "legacy_v4_policy_requires_rehearsal", "partial_packaging_variant_binding_restore", "carrier_empty_code", "zone_invalid_json", "zone_invalid_country", "packaging_negative_dimension", "packaging_invalid_existing_parent", "packaging_wrong_variant_product", "packaging_missing_source_product", "packaging_missing_source_variant", "packaging_deleted_source_product", "packaging_deleted_source_variant", "customs_invalid_hs_code", "customs_invalid_risk_tag", "customs_repeated_risk_tag", "customs_invalid_status", "carrier_natural_identity_conflict", "customs_natural_identity_conflict", "packaging_current_default_binding_conflict", "packaging_current_variant_binding_conflict", "packaging_changed_read_only_product", "packaging_changed_read_only_variant", "partial_packaging_target_only_wrong_variant", "changed_existing_packaging_rule", "legacy_v5_policy_requires_rehearsal"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			cfg := createIndependentRecoveryTestConfiguration(t)
			cfg.DatabaseURL, cfg.DatabaseName = connectionURL, "local_backup_configuration_"+newTaskID()[:12]
			cfg.ApplicationDatabaseRole = "synthetic_configuration_" + newTaskID()[:12]
			cfg.MaxFiles, cfg.MaxExpandedBytes, cfg.MaxWorkspaceBytes = 10000, 256<<20, 1<<30
			control, err := cfg.connectDatabase(ctx, "postgres")
			require.NoError(t, err)
			_, err = control.ExecContext(ctx, `CREATE ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)+` NOLOGIN`)
			require.NoError(t, err)
			_, err = control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName))
			require.NoError(t, err)
			var manager *Manager
			t.Cleanup(func() {
				if manager != nil {
					require.NoError(t, manager.Close())
				}
				_, err := control.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(cfg.DatabaseName)+` WITH (FORCE)`)
				require.NoError(t, err)
				_, err = control.ExecContext(ctx, `DROP ROLE `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole))
				require.NoError(t, err)
				require.NoError(t, control.Close())
			})
			_, err = control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` SET timezone='Asia/Shanghai'`)
			require.NoError(t, err)
			database, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			driver, err := migrationpostgres.WithInstance(database, &migrationpostgres.Config{})
			require.NoError(t, err)
			migrator, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(migrationsDirectory), "postgres", driver)
			require.NoError(t, err)
			err = migrator.Up()
			if err != migrate.ErrNoChange {
				require.NoError(t, err)
			}
			_, err = migrator.Close()
			require.NoError(t, err)
			database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE NOT dirty`).Scan(&cfg.ExpectedMigrationVersion))
			require.NoError(t, maintenancegate.RegisterApplicationBoundary(cfg.MaintenanceDirectory, cfg.DatabaseName, cfg.ApplicationVersion, cfg.StorageRoots, cfg.RecoverySecrets))
			picture := filepath.Join(cfg.StorageRoots["uploads"], "configuration-site-content-fixture.webp")
			require.NoError(t, os.WriteFile(picture, []byte("synthetic website picture"), 0600))
			_, err = database.ExecContext(ctx, `
INSERT INTO users(id,email,username,password,role) VALUES(900001,'configuration-admin@example.invalid','configuration-admin','synthetic-original-password','admin');
INSERT INTO products(id,name,slug,price_minor,currency) VALUES(900001,'Preserved product','configuration-preserved-product',123,'USD');
INSERT INTO product_variants(id,product_id,sku,price_minor,currency,stock) VALUES(900001,900001,'configuration-preserved-sku',123,'USD',19);
INSERT INTO orders(id,order_number,status,currency,payment_currency,payment_amount_minor,subtotal_amount_minor,total_amount_minor,shipping_fee_minor,tax_amount_minor,discount_amount_minor,points_value_minor) VALUES(900001,'configuration-preserved-order','pending','USD','USD',123,123,123,0,0,0,0);
INSERT INTO tickets(id,ticket_number,conversation_id,user_id,subject,category) VALUES(900001,'configuration-preserved-chat','configuration-preserved-chat',900001,'Preserved chat','customer_service');
INSERT INTO ticket_messages(id,ticket_id,user_id,content,attachments,attachment_reference_format) VALUES(900001,900001,900001,'Current chat must survive configuration restore','[]','media_asset_bindings');
INSERT INTO settings(key,value,"group",locale) VALUES('configuration_preserved_secret','synthetic-secret-do-not-publish','api','global');
INSERT INTO media_assets(id,filename,url,storage_key,status,visibility,uploader_id,uploader_actor_type) VALUES(900050,'configuration-site-content-fixture.webp','/uploads/configuration-site-content-fixture.webp','configuration-site-content-fixture.webp','active','public',900001,'administrator');
INSERT INTO categories(id,name,slug,locale) VALUES(900010,'Parent category','configuration-parent','en');
INSERT INTO categories(id,name,slug,locale,parent_id) VALUES(900011,'Child category','configuration-child','en',900010);
INSERT INTO posts(id,title,slug,content,status,author_id,locale,featured_image) VALUES(900010,'Restored article','configuration-content-parent','<p>Original content</p>','published',900001,'en','/uploads/configuration-site-content-fixture.webp');
INSERT INTO posts(id,title,slug,content,status,author_id,locale,parent_id) VALUES(900011,'Translated article','configuration-content-child','<p>Translation</p>','published',900001,'zh',900010);
INSERT INTO post_categories(id,post_id,category_id) VALUES(900010,900010,900011);
INSERT INTO galleries(id,name,slug,status,cover_image) VALUES(900010,'Restored gallery','configuration-gallery','published','/uploads/configuration-site-content-fixture.webp');
INSERT INTO gallery_images(id,gallery_id,url,media_asset_id) VALUES(900010,900010,'/uploads/configuration-site-content-fixture.webp',900050);
INSERT INTO gallery_product_links(id,gallery_id,product_id) VALUES(900010,900010,900001);
INSERT INTO faq_pages(id,page_id,route_path,locale,title,status) VALUES(900010,'configuration-faq','/support/configuration-fixture','en','Restored FAQ','active');
INSERT INTO faqs(id,question,answer,page_id,locale,status) VALUES(900010,'Restored question?','Restored answer','configuration-faq','en','published');
INSERT INTO visual_showcase_items(id,showcase_key,locale,image_url,storage_key,title,alt_text) VALUES(900010,'configuration-visual','en','/uploads/configuration-site-content-fixture.webp','configuration-site-content-fixture.webp','Restored visual','Example');
INSERT INTO storefront_redirect_rules(id,source_path,target_path,state,created_by_id,published_by_id,published_at) VALUES(900010,'/old-configuration-content','/support/configuration-fixture','published',900001,900001,now());
`)
			require.NoError(t, err)
			createConfigurationFitmentCatalogRestoreIntegrationFixture(t, database, scenario)
			createConfigurationSelectionFlowAndQuestionnaireRestoreIntegrationFixture(t, database, scenario)
			createConfigurationMarketCurrencyAndTaxRestoreIntegrationFixture(t, database, scenario)
			createConfigurationEmailTemplateHistoryAndMemberLevelRestoreIntegrationFixture(t, database, scenario)
			createConfigurationShippingReferencePackagingAndCustomsRestoreIntegrationFixture(t, database, scenario)
			// These invalid source values cannot be persisted under the production
			// schema. The fixture checks the exact constraint and unchanged row;
			// malformed package fields are covered separately by domain tests.
			if scenario == "customs_invalid_hs_code" || scenario == "customs_invalid_status" {
				require.NoError(t, database.Close())
				return
			}
			manager, err = NewManager(cfg)
			require.NoError(t, err)
			backup, err := manager.CreateSelectedBusinessModuleBackup("synthetic-exporter", "BACKUP MODULES "+cfg.SiteID+"/"+cfg.Environment+"/configuration", []string{"configuration"})
			require.NoError(t, err)
			backup = waitForIndependentRecoveryTask(t, manager, backup.ID)
			require.Equal(t, "ready_to_download", backup.State, backup.Error)
			key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
			require.NoError(t, err)
			token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic-configuration-operator", []string{PermissionExport, PermissionRestore})
			require.NoError(t, err)
			ticketRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+backup.ID+"/download-ticket", strings.NewReader(`{}`))
			ticketRequest.Header.Set("Authorization", "Bearer "+token)
			ticketResponse := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(ticketResponse, ticketRequest)
			require.Equal(t, http.StatusOK, ticketResponse.Code, ticketResponse.Body.String())
			var ticket struct {
				DownloadPath string `json:"download_path"`
			}
			require.NoError(t, json.Unmarshal(ticketResponse.Body.Bytes(), &ticket))
			download := httptest.NewRequestWithContext(context.Background(), http.MethodGet, ticket.DownloadPath, nil)
			for _, cookie := range ticketResponse.Result().Cookies() {
				download.AddCookie(cookie)
			}
			downloadResponse := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(downloadResponse, download)
			require.Equal(t, http.StatusOK, downloadResponse.Code)
			task := uploadProductCatalogPackageForIntegrationTest(t, manager, downloadResponse.Body.Bytes())
			require.Equal(t, "module_package_integrity_verified", task.State, task.Error)
			_, err = database.ExecContext(ctx, `
DELETE FROM post_categories WHERE id=900010;DELETE FROM posts WHERE id=900011;DELETE FROM posts WHERE id=900010;
DELETE FROM categories WHERE id=900011;DELETE FROM categories WHERE id=900010;
DELETE FROM gallery_product_links WHERE id=900010;DELETE FROM gallery_images WHERE id=900010;DELETE FROM galleries WHERE id=900010;
DELETE FROM faqs WHERE id=900010;DELETE FROM faq_pages WHERE id=900010;
DELETE FROM visual_showcase_items WHERE id=900010;DELETE FROM storefront_redirect_rules WHERE id=900010;
INSERT INTO posts(id,title,slug,content,status,author_id,locale) VALUES(900020,'Current extra article','configuration-current-extra','Must remain','published',900001,'en');
`)
			require.NoError(t, err)
			deleteConfigurationFitmentCatalogFixtureRowsAndAddPreservedTargetOnlyEntry(t, database, scenario)
			expectedSelectionRows := deleteConfigurationSelectionFlowAndQuestionnaireFixtureRowsAndAddPreservedRuntimeData(t, database, scenario)
			expectedMarketRows := deleteConfigurationMarketCurrencyAndTaxFixtureRowsAndPreserveNewCurrentRules(t, database, scenario)
			expectedEmailAndMemberRows := deleteConfigurationEmailTemplateAndMemberLevelFixtureRowsAndPreserveCurrentRules(t, database, scenario)
			expectedShippingReferenceRows := deleteConfigurationShippingReferencePackagingAndCustomsFixtureRowsAndPreserveCurrentData(t, database, scenario)
			if scenario == "missing_picture" {
				require.NoError(t, os.Remove(picture))
			}
			if scenario == "missing_setting" {
				_, err = database.ExecContext(ctx, `DELETE FROM settings WHERE key='configuration_preserved_secret'`)
				require.NoError(t, err)
			}
			if scenario == "different_existing_content" {
				_, err = database.ExecContext(ctx, `INSERT INTO posts(id,title,slug,content,status,author_id,locale) VALUES(900010,'Changed current article','configuration-content-parent','New content must not be overwritten','published',900001,'en')`)
				require.NoError(t, err)
			}
			if scenario == "disabled_foreign_key_guard" {
				_, err = database.ExecContext(ctx, `ALTER TABLE post_categories DISABLE TRIGGER ALL`)
				require.NoError(t, err)
			}
			_, before, _, err := captureDatabaseInventory(ctx, database)
			require.NoError(t, err)
			require.NoError(t, database.Close())
			_, err = manager.StartSelectedBusinessModuleIsolatedRestorePreview(task.ID, "VERIFY MODULE RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.ID)
			require.NoError(t, err)
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			if configurationShippingReferencePackagingAndCustomsIntegrationScenarioRequiresRejection(scenario) || configurationEmailTemplateHistoryAndMemberLevelIntegrationScenarioRequiresRejection(scenario) || configurationMarketCurrencyAndTaxIntegrationScenarioRequiresRejection(scenario) || configurationSelectionFlowAndQuestionnaireIntegrationScenarioRequiresRejection(scenario) || scenario == "missing_picture" || scenario == "missing_setting" || scenario == "different_existing_content" || scenario == "disabled_foreign_key_guard" || scenario == "wrong_fork_hub_position" || scenario == "wrong_frame_hub_position" || scenario == "disabled_associated_hub" || scenario == "enabled_fork_without_hub" || scenario == "wrong_preset_rim_brand" || scenario == "wrong_preset_wheel_position" || scenario == "invalid_catalog_domain_field" || scenario == "missing_catalog_product_brand" || scenario == "catalog_natural_identity_conflict" || scenario == "changed_existing_catalog" {
				require.Contains(t, []string{"module_restore_conflicts_found", "failed"}, task.State, task.Error)
				require.Error(t, validateConfigurationModuleSiteContentFormalRestorePreview(task, task.SelectedModuleRestorePreview.SHA256))
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				defer database.Close()
				_, actual, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, before, actual)
				return
			}
			require.Equal(t, "module_isolated_restore_verified", task.State, task.Error)
			require.True(t, task.SelectedModuleRestorePreview.ConfigurationSiteContentRestoreEligibility.Eligible)
			require.Equal(t, 25+expectedSelectionRows+expectedMarketRows+expectedEmailAndMemberRows+expectedShippingReferenceRows, task.SelectedModuleRestorePreview.RestoredRows)
			if scenario == "legacy_policy_requires_rehearsal" || scenario == "legacy_v2_policy_requires_rehearsal" || scenario == "legacy_v3_policy_requires_rehearsal" || scenario == "legacy_v4_policy_requires_rehearsal" || scenario == "legacy_v5_policy_requires_rehearsal" {
				task.SelectedModuleRestorePreview.ConfigurationSiteContentRestoreEligibility.Policy = "restore_missing_site_content_preserve_existing_configuration_and_registered_files_v1"
				if scenario == "legacy_v2_policy_requires_rehearsal" {
					task.SelectedModuleRestorePreview.ConfigurationSiteContentRestoreEligibility.Policy = "restore_missing_site_content_and_fitment_catalog_preserve_existing_configuration_and_registered_files_v2"
				}
				if scenario == "legacy_v3_policy_requires_rehearsal" {
					task.SelectedModuleRestorePreview.ConfigurationSiteContentRestoreEligibility.Policy = "restore_missing_content_catalog_selection_flows_and_questionnaires_preserve_existing_configuration_and_registered_files_v3"
				}
				if scenario == "legacy_v4_policy_requires_rehearsal" {
					task.SelectedModuleRestorePreview.ConfigurationSiteContentRestoreEligibility.Policy = "restore_missing_content_catalog_selection_and_market_currency_tax_preserve_existing_configuration_and_registered_files_v4"
				}
				if scenario == "legacy_v5_policy_requires_rehearsal" {
					task.SelectedModuleRestorePreview.ConfigurationSiteContentRestoreEligibility.Policy = "restore_missing_business_configuration_including_email_templates_and_member_levels_preserve_existing_rows_and_registered_files_v5"
				}
				task.SelectedModuleRestorePreview.SHA256 = ""
				task.SelectedModuleRestorePreview.SHA256, err = businessModuleRecoveryCanonicalJSONSHA256(task.SelectedModuleRestorePreview)
				require.NoError(t, err)
				require.NoError(t, manager.saveTask(task))
				_, err = manager.StartConfigurationModuleSiteContentFormalRestore(task.ID, "synthetic-configuration-operator", ConfigurationModuleSiteContentFormalRestoreRequest{Confirmation: "RESTORE CONFIGURATION CONTENT " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, ExpectedPreviewSHA256: task.SelectedModuleRestorePreview.SHA256})
				require.Error(t, err)
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				defer database.Close()
				_, actual, _, err := captureDatabaseInventory(ctx, database)
				require.NoError(t, err)
				require.Equal(t, before, actual)
				return
			}
			invoke := func(action string, body any) *httptest.ResponseRecorder {
				t.Helper()
				data, err := json.Marshal(body)
				require.NoError(t, err)
				request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, ConsolePrefix+"/api/tasks/"+task.ID+"/configuration-site-content-"+action, strings.NewReader(string(data)))
				request.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				manager.HTTPHandler().ServeHTTP(response, request)
				return response
			}
			if scenario == "stale_database_after_preview" {
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				_, err = database.ExecContext(ctx, `UPDATE users SET first_name='Newer current profile' WHERE id=900001`)
				require.NoError(t, err)
				require.NoError(t, database.Close())
			}
			if strings.HasPrefix(scenario, "rows_") || scenario == "restore_intent_persisted" {
				require.NoError(t, manager.startTaskOperation(task, "synthetic_configuration_content_restore", func(ctx context.Context, task *Task) error {
					return manager.restoreConfigurationModuleMissingSiteContentWithAtomicReceipt(ctx, task, "synthetic-configuration-operator", func(point string) error {
						if point == scenario {
							return errors.New("synthetic configuration interruption")
						}
						return nil
					})
				}))
			} else {
				response := invoke("formal-restore", ConfigurationModuleSiteContentFormalRestoreRequest{Confirmation: "RESTORE CONFIGURATION CONTENT " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, ExpectedPreviewSHA256: task.SelectedModuleRestorePreview.SHA256})
				require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			}
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			if scenario == "stale_database_after_preview" {
				require.Equal(t, "failed", task.State)
				require.Nil(t, task.ConfigurationModuleSiteContentRestore)
				return
			}
			if strings.HasPrefix(scenario, "rows_") || scenario == "restore_intent_persisted" {
				require.Equal(t, "failed", task.State, task.Error)
				require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
				task.ExpiresAt = time.Now().Add(-time.Hour)
				require.NoError(t, manager.saveTask(task))
				require.NoError(t, manager.ExpireWorkCopies(ctx))
				require.NoError(t, manager.Close())
				manager, err = NewManager(cfg)
				require.NoError(t, err)
				response := invoke("interrupted-restore-recovery", map[string]string{"confirmation": "RECOVER CONFIGURATION CONTENT " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID})
				require.Equal(t, http.StatusOK, response.Code, response.Body.String())
				task = waitForIndependentRecoveryTask(t, manager, task.ID)
				if scenario != "rows_committed_before_acknowledgement" {
					require.Equal(t, "rolled_back", task.State, task.Error)
					database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
					require.NoError(t, err)
					defer database.Close()
					_, actual, _, err := captureDatabaseInventory(ctx, database)
					require.NoError(t, err)
					require.Equal(t, before, actual)
					return
				}
			}
			require.Equal(t, "configuration_site_content_restored_still_isolated", task.State, task.Error)
			fenced, err := maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.True(t, fenced)
			require.Error(t, manager.ReleaseUnmodifiedMaintenance(task.ID))
			release := ConfigurationContentReviewedServiceReleaseRequest{Confirmation: "RELEASE CONFIGURATION CONTENT " + cfg.SiteID + "/" + cfg.Environment + "/" + task.ID, ExpectedJournalSHA256: task.ConfigurationModuleSiteContentRestore.SHA256}
			response := invoke("reviewed-service-release", release)
			require.Equal(t, http.StatusConflict, response.Code)
			release.ContentAndCacheReviewReference = "synthetic isolated content and cache review 2026-10-30"
			if scenario == "review_result_changed" {
				database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
				require.NoError(t, err)
				_, err = database.ExecContext(ctx, `UPDATE posts SET title='Changed after commit' WHERE id=900010`)
				require.NoError(t, err)
				require.NoError(t, database.Close())
			}
			response = invoke("reviewed-service-release", release)
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			task = waitForIndependentRecoveryTask(t, manager, task.ID)
			if scenario == "review_result_changed" {
				require.Equal(t, "failed", task.State)
				fenced, err = maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
				require.NoError(t, err)
				require.True(t, fenced)
				return
			}
			require.Equal(t, "completed", task.State, task.Error)
			require.NoError(t, validateTransactionalBusinessModuleRestoreProgress(task))
			require.NotNil(t, task.ConfigurationModuleSiteContentRestore.ConfigurationContentProjectionReview)
			fenced, err = maintenancegate.IsApplicationFenced(cfg.MaintenanceDirectory)
			require.NoError(t, err)
			require.False(t, fenced)
			database, err = cfg.connectDatabase(ctx, cfg.DatabaseName)
			require.NoError(t, err)
			defer database.Close()
			var title, chat, secret string
			var audits, stock int
			require.NoError(t, database.QueryRowContext(ctx, `SELECT title FROM posts WHERE slug='configuration-content-parent' AND locale='en' AND status='published'`).Scan(&title))
			require.Equal(t, "Restored article", title)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT content FROM ticket_messages WHERE id=900001`).Scan(&chat))
			require.Equal(t, "Current chat must survive configuration restore", chat)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT stock FROM product_variants WHERE id=900001`).Scan(&stock))
			require.Equal(t, 19, stock)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT value FROM settings WHERE key='configuration_preserved_secret'`).Scan(&secret))
			require.Equal(t, "synthetic-secret-do-not-publish", secret)
			require.NoError(t, database.QueryRowContext(ctx, `SELECT count(*) FROM audit_logs WHERE resource='configuration_module_site_content_restore'`).Scan(&audits))
			require.Equal(t, 1, audits)
			public, err := json.Marshal(task)
			require.NoError(t, err)
			require.NotContains(t, string(public), secret)
			require.NotContains(t, string(public), "Original content")
			verifyRestoredConfigurationFitmentCatalogThroughPublicBusinessHTTPAndIndependentTireFacts(t, database)
			verifyRestoredConfigurationSelectionFlowsAndQuestionnairesThroughPublicBusinessHTTPWithoutPublishingDraftsOrChangingSessions(t, database)
			verifyRestoredConfigurationMarketsPublicTaxAndExactExchangeWithoutChangingOrdersExpiryOrSyncControls(t, database, scenario)
			verifyRestoredEmailTemplatesHistoricalVersionsAndMemberLevelsThroughBusinessServicesWithoutSendingOrRewarding(t, database)
			verifyRestoredShippingReferencePackagingAndCustomsThroughBusinessRepositoriesWithoutExternalCallsOrOrderWrites(t, database)
		})
	}
}

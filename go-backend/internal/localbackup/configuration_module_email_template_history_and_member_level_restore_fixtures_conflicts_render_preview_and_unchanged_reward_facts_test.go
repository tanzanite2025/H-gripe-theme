package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestConfigurationEmailTemplatesAndMemberLevelsKeepOriginalValuesAndRejectInvalidRenderableContracts(t *testing.T) {
	template := `{"code":"order_confirmation","locale":"fr","category":"order","name":"Original template","version":2,"subject_template":"Order {{order_number}}","body_html":"<p>{{customer_name}}</p>","body_text":"Paid {{order_amount}} at {{paid_at}}","allowed_variables":["order_number","order_amount","paid_at","customer_name","currency","items"],"required_variables":["order_number","order_amount","paid_at"]}`
	level := `{"name":"Original tier","min_points":1000000000,"max_points":1499999999,"discount_rate_decimal":7.125000000000001}`
	for _, example := range []struct {
		name, table, row string
		valid            bool
	}{
		{"renderable_template", "email_templates", template, true},
		{"unknown_code", "email_templates", strings.ReplaceAll(template, "order_confirmation", "unknown_template"), false},
		{"locale_alias", "email_templates", strings.ReplaceAll(template, `"fr"`, `"fr-FR"`), false},
		{"wrong_category", "email_templates", strings.ReplaceAll(template, `"category":"order"`, `"category":"system"`), false},
		{"unknown_placeholder", "email_templates", strings.ReplaceAll(template, "{{customer_name}}", "{{server_secret}}"), false},
		{"malformed_placeholder", "email_templates", strings.ReplaceAll(template, "{{customer_name}}", "{{{customer_name}}}"), false},
		{"subject_injection", "email_templates", strings.ReplaceAll(template, "Order {{order_number}}", `Order\nBcc: injected`), false},
		{"variable_wrong_json", "email_templates", strings.ReplaceAll(template, `["order_number","order_amount","paid_at"]`, `{}`), false},
		{"variable_whitespace", "email_templates", strings.Replace(template, `"order_number",`, `" order_number",`, 1), false},
		{"variable_duplicate", "email_templates", strings.ReplaceAll(template, `"currency","items"`, `"currency","currency"`), false},
		{"empty_body", "email_templates", strings.ReplaceAll(strings.ReplaceAll(template, `<p>{{customer_name}}</p>`, ""), `Paid {{order_amount}} at {{paid_at}}`, ""), false},
		{"historical_snapshot", "email_template_versions", strings.ReplaceAll(template, `,"category":"order"`, ""), true},
		{"exact_discount", "member_levels", level, true},
		{"negative_range", "member_levels", strings.ReplaceAll(level, `1000000000`, `-1`), false},
		{"reversed_range", "member_levels", strings.ReplaceAll(level, `1499999999`, `999999999`), false},
		{"out_of_range_discount", "member_levels", strings.ReplaceAll(level, `7.125000000000001`, `101`), false},
		{"excess_discount_precision", "member_levels", strings.ReplaceAll(level, `7.125000000000001`, `7.1250000000000001`), false},
	} {
		t.Run(example.name, func(t *testing.T) {
			row := json.RawMessage(example.row)
			err := validateConfigurationModuleMissingEmailTemplateAndMemberLevelDomainRows(map[string][]json.RawMessage{example.table: {row}})
			require.Equal(t, example.valid, err == nil, "%v", err)
			require.Equal(t, example.row, string(row))
		})
	}
	for _, table := range []string{"email_provider_configs", "loyalty_program_configs", "referral_program_configs", "coupons", "user_loyalty", "loyalty_transactions", "outbox_events", "showcases"} {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, table), table)
	}
}

func createConfigurationEmailTemplateHistoryAndMemberLevelRestoreIntegrationFixture(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO email_templates(id,code,locale,category,name,subject_template,body_html,body_text,allowed_variables,required_variables,is_enabled,version)
VALUES(900110,'order_confirmation','fr','order','Restored French notification','Restored order {{order_number}}','<p>{{customer_name}}</p>','Paid {{order_amount}} at {{paid_at}}','["order_number","order_amount","paid_at","customer_name","currency","items"]','["order_number","order_amount","paid_at"]',true,2),
(900111,'order_confirmation','ja','order','Preserved disabled notification','Disabled order {{order_number}}','','Paid {{order_amount}} at {{paid_at}}','["order_number","order_amount","paid_at","customer_name","currency","items"]','["order_number","order_amount","paid_at"]',false,1);
INSERT INTO email_template_versions(id,template_id,code,locale,version,name,subject_template,body_html,body_text,allowed_variables,required_variables,changed_by_user_id,change_reason)
SELECT CASE WHEN id=900110 THEN 900111 ELSE 900112 END,id,code,locale,version,name,subject_template,body_html,body_text,allowed_variables,required_variables,900001,'Original historical change' FROM email_templates WHERE id IN(900110,900111);
INSERT INTO email_template_versions(id,template_id,code,locale,version,name,subject_template,body_html,body_text,allowed_variables,required_variables,changed_by_user_id,change_reason)
SELECT 900110,id,code,locale,1,'Historical French notification','Historical order {{order_number}}',body_html,body_text,allowed_variables,required_variables,900001,'Original version one' FROM email_templates WHERE id=900110;
INSERT INTO member_levels(id,name,min_points,max_points,discount_rate_decimal,benefits,icon,color,sort_order) VALUES(900110,'Restored exact discount tier',1000000000,1499999999,7.125000000000001,'["Original benefit"]','gem','#b9f2ff',90);
INSERT INTO user_loyalty(id,user_id,total_points,available_points) VALUES(900110,900001,17,17);
INSERT INTO loyalty_transactions(id,user_id,type,points,balance,source,source_id,description) VALUES(900110,900001,'earn',17,17,'referral',0,'Original registration referral reward');
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"email_unknown_template_code":          `UPDATE email_templates SET code='unknown_template' WHERE id=900110;UPDATE email_template_versions SET code='unknown_template' WHERE template_id=900110`,
		"email_invalid_locale":                 `UPDATE email_templates SET locale='xx' WHERE id=900110;UPDATE email_template_versions SET locale='xx' WHERE template_id=900110`,
		"email_category_mismatch":              `UPDATE email_templates SET category='system' WHERE id=900110`,
		"email_variables_wrong_type":           `UPDATE email_templates SET allowed_variables='{}' WHERE id=900110`,
		"email_unknown_placeholder":            `UPDATE email_templates SET body_text='{{server_secret}}' WHERE id=900110`,
		"email_subject_header_injection":       `UPDATE email_templates SET subject_template=E'Order\nBcc: injected' WHERE id=900110`,
		"email_history_wrong_identity":         `UPDATE email_template_versions SET locale='de' WHERE id=900110`,
		"email_history_missing_version":        `DELETE FROM email_template_versions WHERE id=900110`,
		"email_current_snapshot_mismatch":      `UPDATE email_template_versions SET body_text='Changed current snapshot' WHERE id=900111`,
		"email_history_future_version":         `UPDATE email_template_versions SET version=3 WHERE id=900110`,
		"partial_email_history_invalid_parent": `UPDATE email_templates SET category='system' WHERE id=900110`,
		"member_level_invalid_name":            `UPDATE member_levels SET name=' ' WHERE id=900110`,
		"member_level_invalid_range":           `UPDATE member_levels SET max_points=999999999 WHERE id=900110`,
		"member_level_package_overlap":         `INSERT INTO member_levels(id,name,min_points,max_points,discount_rate_decimal) VALUES(900111,'Overlapping package rule',1100000000,1200000000,3)`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationEmailTemplateAndMemberLevelFixtureRowsAndPreserveCurrentRules(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM email_templates WHERE id IN(900110,900111);DELETE FROM member_levels WHERE id IN(900110,900111);`, 6
	switch scenario {
	case "partial_template_history_restore", "partial_email_history_invalid_parent", "partial_template_target_only_future_history":
		statement, count = `DELETE FROM email_template_versions WHERE id=900110;`, 1
	case "partial_email_current_version_restore":
		statement, count = `DELETE FROM email_template_versions WHERE id=900111;`, 1
	case "disabled_email_template_restore":
		statement, count = `DELETE FROM email_templates WHERE id=900111;`, 2
	}
	_, err := database.ExecContext(context.Background(), statement+`INSERT INTO member_levels(id,name,min_points,max_points,discount_rate_decimal) VALUES(900119,'Preserved new current tier',1500000000,1599999999,3);`)
	require.NoError(t, err)
	mutation := map[string]string{
		"member_level_current_overlap":                `INSERT INTO member_levels(id,name,min_points,max_points,discount_rate_decimal) VALUES(900118,'Current overlapping decision',1400000000,1499999999,4)`,
		"changed_existing_email_template":             `INSERT INTO email_templates(id,code,locale,category,name,subject_template,body_text,allowed_variables,required_variables,version) VALUES(900110,'order_confirmation','fr','order','Current changed template','Current order','Current text','[]','[]',3)`,
		"partial_template_target_only_future_history": `INSERT INTO email_template_versions(id,template_id,code,locale,version,name,subject_template,body_html,body_text,allowed_variables,required_variables) SELECT 900119,id,code,locale,3,name,subject_template,body_html,body_text,allowed_variables,required_variables FROM email_templates WHERE id=900110`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationEmailTemplateHistoryAndMemberLevelIntegrationScenarioRequiresRejection(scenario string) bool {
	switch scenario {
	case "email_unknown_template_code", "email_invalid_locale", "email_category_mismatch", "email_variables_wrong_type", "email_unknown_placeholder", "email_subject_header_injection", "email_history_wrong_identity", "email_history_missing_version", "email_current_snapshot_mismatch", "email_history_future_version", "partial_email_history_invalid_parent", "member_level_invalid_name", "member_level_invalid_range", "member_level_package_overlap", "member_level_current_overlap", "changed_existing_email_template", "partial_template_target_only_future_history":
		return true
	}
	return false
}

func verifyRestoredEmailTemplatesHistoricalVersionsAndMemberLevelsThroughBusinessServicesWithoutSendingOrRewarding(t *testing.T, database *sql.DB) {
	t.Helper()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	templateService := service.NewTransactionalNotificationTemplateService(repository.NewNotificationTemplateRepository(gormDatabase))
	loyaltyRepository := repository.NewLoyaltyRepository(gormDatabase)
	template, err := templateService.Get(900110)
	require.NoError(t, err)
	require.Equal(t, "Restored French notification", template.Name)
	versions, err := templateService.Versions(900110)
	require.NoError(t, err)
	require.Len(t, versions, 2)
	require.Equal(t, "Historical French notification", versions[1].Name)
	disabled, err := templateService.Get(900111)
	require.NoError(t, err)
	require.False(t, disabled.IsEnabled)
	preview, err := templateService.Render("order_confirmation", "fr", map[string]string{"order_number": "RESTORE-123", "order_amount": "100.00", "paid_at": "original time", "customer_name": "<Customer>"})
	require.NoError(t, err)
	require.Equal(t, "Restored order RESTORE-123", preview.Subject)
	require.Equal(t, "<p>&lt;Customer&gt;</p>", preview.HTML)
	require.Equal(t, 2, preview.TemplateVersion)
	resolved, err := templateService.Resolve("order_confirmation", "ja")
	require.NoError(t, err)
	require.Equal(t, "en", resolved.Locale)
	level, err := loyaltyRepository.FindMemberLevelByPoints(1200000000)
	require.NoError(t, err)
	require.Equal(t, uint(900110), level.ID)
	require.Equal(t, "7.125000000000001", level.DiscountRateDecimal)
	points, err := loyaltyRepository.GetUserPointsBalance(900001)
	require.NoError(t, err)
	require.Equal(t, 17, points)
	var rewards int
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT count(*) FROM loyalty_transactions WHERE user_id=900001 AND source='referral'`).Scan(&rewards))
	require.Equal(t, 1, rewards)
}

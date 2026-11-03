package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	quickbuyhttp "commerce-platform/internal/api/v1/quickbuy"
	selectionhttp "commerce-platform/internal/api/v1/selectionassistant"
	wheelsethttp "commerce-platform/internal/api/v1/wheelsetfit"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestConfigurationSelectionConfigurationKeysRejectInvalidKindsOrCodesWithoutChangingOriginalBytes(t *testing.T) {
	for _, example := range []struct {
		row   string
		valid bool
	}{
		{`{"kind":"question_key","code":"restored_question"}`, true},
		{`{"kind":"answer_key","code":"restored_answer"}`, true},
		{`{"kind":"other","code":"restored_answer"}`, false},
		{`{"kind":"question_key","code":" Bad KEY "}`, false},
		{`{"kind":"answer_key","code":""}`, false},
		{`{"kind":null,"code":"restored_answer"}`, false},
	} {
		t.Run(example.row, func(t *testing.T) {
			row := json.RawMessage(example.row)
			err := validateConfigurationModuleMissingSelectionConfigurationKeyRows(map[string][]json.RawMessage{"selection_configuration_keys": {row}})
			require.Equal(t, example.valid, err == nil)
			require.Equal(t, example.row, string(row))
		})
	}
}

func createConfigurationSelectionFlowAndQuestionnaireRestoreIntegrationFixture(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO product_specification_templates(id,name,slug,is_enabled) VALUES(900070,'Preserved selection template','configuration-selection-template',true);
INSERT INTO product_spec_definitions(id,product_specification_template_id,name,slug,is_filterable) VALUES(900070,900070,'Preserved filter definition','configuration_restore_depth',true);
INSERT INTO selection_configuration_keys(id,kind,code,is_enabled) VALUES(900070,'question_key','configuration_restored_question',true),(900071,'answer_key','configuration_restored_answer',true);
INSERT INTO selection_assistant_flows(id,slug,name) VALUES(900070,'configuration-restored-selection','Restored selection assistant');
INSERT INTO selection_assistant_flow_versions(id,flow_id,version_number,status,published_at,published_by,config) VALUES
(900070,900070,1,'published',now(),900001,'{"kind":"product_selection_assistant","schema_version":1,"entry_node_key":"choose","base_product_query":{"category_slug":"wheelset"},"nodes":[{"key":"choose","type":"question","prompt":{"en":"Restored selection question"},"options":[{"key":"road","label":{"en":"Restored selection option"},"next_node_key":"finish","query_effects":{"spec_filters":{"configuration_restore_depth":["40"]}}}]},{"key":"finish","type":"terminal"}]}'),
(900071,900070,2,'draft',NULL,NULL,'{}');
INSERT INTO quick_buy_flows(id,slug,name,entry_surface) VALUES(900070,'configuration-restored-quick','Restored quick flow','configuration-restore-fixture');
INSERT INTO quick_buy_flow_versions(id,flow_id,version_number,status,published_at,published_by) VALUES(900070,900070,1,'published',now(),900001),(900071,900070,2,'draft',NULL,NULL);
INSERT INTO quick_buy_flow_translations(id,flow_id,locale,help_text) VALUES(900070,900070,'en','Restored quick help');
INSERT INTO quick_buy_steps(id,flow_version_id,step_key,name,selection_mode,min_select,max_select) VALUES(900070,900070,'choose','Restored quick step','single',1,1);
INSERT INTO quick_buy_step_filters(id,step_id,filter_type,spec_definition_id,operator,value) VALUES(900070,900070,'spec',900070,'eq','"40"');
INSERT INTO quick_buy_step_product_categories(id,step_id,product_category_id,is_primary) SELECT 900070,900070,id,true FROM product_categories WHERE slug='wheelset';
INSERT INTO quick_buy_step_product_specification_templates(id,step_id,product_specification_template_id,is_primary) VALUES(900070,900070,900070,true);
INSERT INTO quick_buy_compatibility_rules(id,flow_version_id,rule_key,rule_type,source_step_key,target_step_key,rule) VALUES(900070,900070,'configuration-rule','match','choose','choose','{"original_rule_fact":true}');
DELETE FROM wheelset_fit_questionnaires;
INSERT INTO wheelset_fit_questionnaires(id,slug) VALUES(900070,'wheelset-fit');
INSERT INTO wheelset_fit_questionnaire_versions(id,questionnaire_id,version_number,status,published_at,published_by) VALUES(900070,900070,1,'published',now(),900001),(900071,900070,2,'draft',NULL,NULL);
INSERT INTO wheelset_fit_questions(id,questionnaire_version_id,question_key,answer_key,sort_order,allow_unknown) VALUES(900070,900070,'configuration_restored_question','configuration_restored_answer',10,false);
INSERT INTO wheelset_fit_question_translations(id,question_id,locale,prompt,translated_revision) VALUES(900070,900070,'zh_cn','Restored questionnaire prompt',1);
INSERT INTO wheelset_fit_question_options(id,question_id,option_key,answer_value,sort_order,product_filter_effects) VALUES(900070,900070,'road','restored-road-answer',10,'{"spec_filters":{"configuration_restore_depth":["40"]}}');
INSERT INTO wheelset_fit_question_option_translations(id,option_id,locale,label,translated_revision) VALUES(900070,900070,'zh_cn','Restored questionnaire option',1);
`)
	require.NoError(t, err)
	mutations := map[string]string{
		"selection_missing_node":                               `UPDATE selection_assistant_flow_versions SET config=jsonb_set(config,'{nodes,0,options,0,next_node_key}','"missing"') WHERE id=900070`,
		"selection_cycle":                                      `UPDATE selection_assistant_flow_versions SET config=jsonb_set(config,'{nodes,0,options,0,next_node_key}','"choose"') WHERE id=900070`,
		"selection_wrong_json_type":                            `UPDATE selection_assistant_flow_versions SET config='[]' WHERE id=900070`,
		"selection_unknown_filter":                             `UPDATE selection_assistant_flow_versions SET config=jsonb_set(config,'{nodes,0,options,0,query_effects,spec_filters}','{"missing_filter":["40"]}') WHERE id=900070`,
		"quickbuy_missing_rule_step":                           `UPDATE quick_buy_compatibility_rules SET target_step_key='missing' WHERE id=900070`,
		"quickbuy_invalid_selection_bounds":                    `UPDATE quick_buy_steps SET min_select=3 WHERE id=900070`,
		"questionnaire_missing_source_label":                   `DELETE FROM wheelset_fit_question_option_translations WHERE id=900070`,
		"questionnaire_unregistered_key":                       `DELETE FROM selection_configuration_keys WHERE id=900070`,
		"questionnaire_disabled_key":                           `UPDATE selection_configuration_keys SET is_enabled=false WHERE id=900070`,
		"questionnaire_invalid_effects_type":                   `UPDATE wheelset_fit_question_options SET product_filter_effects='[]' WHERE id=900070`,
		"questionnaire_invalid_filter_values":                  `UPDATE wheelset_fit_question_options SET product_filter_effects='{"spec_filters":{"configuration_restore_depth":123}}' WHERE id=900070`,
		"questionnaire_no_enabled_questions":                   `UPDATE wheelset_fit_questions SET is_enabled=false WHERE id=900070`,
		"invalid_selection_key_kind":                           `UPDATE selection_configuration_keys SET kind='unknown' WHERE id=900071`,
		"partial_questionnaire_child_invalid_published_parent": `UPDATE wheelset_fit_questions SET allow_unknown=false WHERE id=900070; UPDATE wheelset_fit_question_options SET is_unknown=true WHERE id=900070`,
		"partial_quickbuy_rule_invalid_published_parent":       `UPDATE quick_buy_compatibility_rules SET target_step_key='missing' WHERE id=900070`,
	}[scenario]
	if mutations != "" {
		_, err = database.ExecContext(context.Background(), mutations)
		require.NoError(t, err)
	}
}

func deleteConfigurationSelectionFlowAndQuestionnaireFixtureRowsAndAddPreservedRuntimeData(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement := `DELETE FROM selection_assistant_flows WHERE id=900070; DELETE FROM quick_buy_flows WHERE id=900070; DELETE FROM wheelset_fit_questionnaires WHERE id=900070; DELETE FROM selection_configuration_keys WHERE id IN (900070,900071);`
	restoredRows := 21
	switch scenario {
	case "partial_questionnaire_source_translation", "partial_questionnaire_child_invalid_published_parent":
		statement = `DELETE FROM wheelset_fit_question_option_translations WHERE id=900070;`
		restoredRows = 1
	case "partial_quickbuy_rule", "partial_quickbuy_rule_invalid_published_parent":
		statement = `DELETE FROM quick_buy_compatibility_rules WHERE id=900070;`
		restoredRows = 1
	case "quickbuy_second_published_version", "selection_second_published_version", "questionnaire_second_published_version":
		statement = `DELETE FROM quick_buy_flow_versions WHERE id=900070; DELETE FROM selection_assistant_flow_versions WHERE id=900070; DELETE FROM wheelset_fit_questionnaire_versions WHERE id=900070;`
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO selection_assistant_flows(id,slug,name,is_enabled) VALUES(900080,'configuration-preserved-current-selection','Current selection must survive',false);
INSERT INTO quick_buy_sessions(id,session_token,flow_id,flow_version_id,user_id) SELECT 900080,'configuration-preserved-current-session',flow_id,id,900001 FROM quick_buy_flow_versions WHERE id<900000 ORDER BY id LIMIT 1;
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"quickbuy_disabled_product_template":       `UPDATE product_specification_templates SET is_enabled=false WHERE id=900070`,
		"selection_disabled_filter_definition":     `UPDATE product_spec_definitions SET is_filterable=false WHERE id=900070`,
		"quickbuy_missing_product_template":        `DELETE FROM product_specification_templates WHERE id=900070`,
		"quickbuy_second_published_version":        `INSERT INTO quick_buy_flow_versions(id,flow_id,version_number,status) VALUES(900080,900070,3,'published')`,
		"selection_second_published_version":       `INSERT INTO selection_assistant_flow_versions(id,flow_id,version_number,status) VALUES(900080,900070,3,'published')`,
		"questionnaire_second_published_version":   `INSERT INTO wheelset_fit_questionnaire_versions(id,questionnaire_id,version_number,status) VALUES(900080,900070,3,'published')`,
		"selection_changed_existing_configuration": `INSERT INTO selection_assistant_flows(id,slug,name) VALUES(900070,'configuration-restored-selection','Newer current configuration')`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return restoredRows
}

func configurationSelectionFlowAndQuestionnaireIntegrationScenarioRequiresRejection(scenario string) bool {
	switch scenario {
	case "selection_missing_node", "selection_cycle", "selection_wrong_json_type", "selection_unknown_filter", "quickbuy_missing_rule_step", "quickbuy_invalid_selection_bounds", "questionnaire_missing_source_label", "questionnaire_unregistered_key", "questionnaire_disabled_key", "questionnaire_invalid_effects_type", "questionnaire_invalid_filter_values", "questionnaire_no_enabled_questions", "invalid_selection_key_kind", "partial_questionnaire_child_invalid_published_parent", "partial_quickbuy_rule_invalid_published_parent", "quickbuy_disabled_product_template", "selection_disabled_filter_definition", "quickbuy_missing_product_template", "quickbuy_second_published_version", "selection_second_published_version", "questionnaire_second_published_version", "selection_changed_existing_configuration":
		return true
	}
	return false
}

func verifyRestoredConfigurationSelectionFlowsAndQuestionnairesThroughPublicBusinessHTTPWithoutPublishingDraftsOrChangingSessions(t *testing.T, database *sql.DB) {
	t.Helper()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	router := gin.New()
	router.GET("/selection-assistant/flows/:slug", selectionhttp.NewHandler(service.NewSelectionAssistantService(repository.NewSelectionAssistantRepository(gormDatabase))).GetPublishedFlow)
	router.GET("/quick-buy/flows/current", quickbuyhttp.NewHandler(service.NewQuickBuyService(repository.NewQuickBuyRepository(gormDatabase), repository.NewProductRepository(gormDatabase), repository.NewProductCategoryRepository(gormDatabase))).GetCurrentFlow)
	router.GET("/wheelset-fit/current", wheelsethttp.NewHandler(service.NewWheelsetFitQuestionnaireService(repository.NewWheelsetFitQuestionnaireRepository(gormDatabase), repository.NewSelectionConfigurationKeyRepository(gormDatabase))).GetCurrentFlow)
	for _, example := range []struct {
		path     string
		expected []string
	}{
		{"/selection-assistant/flows/configuration-restored-selection", []string{"Restored selection option", "configuration_restore_depth", `"id":900070`, `"status":"published"`}},
		{"/quick-buy/flows/current?surface=configuration-restore-fixture&locale=en", []string{"Restored quick step", "Restored quick help", `"id":900070`, `"status":"published"`}},
		{"/wheelset-fit/current", []string{"Restored questionnaire option", "Restored questionnaire prompt", "restored-road-answer", "configuration_restore_depth", `"id":900070`, `"status":"published"`}},
	} {
		t.Run("restored_public_selection"+example.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, example.path, nil))
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			for _, expected := range example.expected {
				require.Contains(t, response.Body.String(), expected)
			}
			require.NotContains(t, response.Body.String(), `"id":900071`)
		})
	}
	for _, table := range []string{"selection_assistant_flow_versions", "quick_buy_flow_versions", "wheelset_fit_questionnaire_versions"} {
		var status string
		require.NoError(t, database.QueryRowContext(context.Background(), `SELECT status FROM public.`+table+` WHERE id=900071`).Scan(&status))
		require.Equal(t, "draft", status)
	}
	var count int
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT count(*) FROM quick_buy_sessions WHERE id=900080 AND session_token='configuration-preserved-current-session' AND status='active'`).Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT count(*) FROM selection_assistant_flows WHERE id=900080 AND name='Current selection must survive'`).Scan(&count))
	require.Equal(t, 1, count)
}

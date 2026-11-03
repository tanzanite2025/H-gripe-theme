package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestConfigurationShippingReferencePackagingAndCustomsRejectInvalidOriginalFieldsWithoutRewritingThem(t *testing.T) {
	packaging := `{"rule_name":"Original carton","box_weight":0.125,"box_length":60.25,"box_width":30,"box_height":10,"max_weight":0}`
	customs := `{"name":"Original classification","slug":"original-classification","hs_code":"871499","cn_code":"87149990","country_of_origin":"CN","status":"active","trade_remedy_risk_level":"high","trade_remedy_risk_tags_json":["eu_anti_dumping_attention"]}`
	for _, sample := range []struct {
		name, table, row string
		valid            bool
	}{
		{"original_exact_dimensions", "shipping_packaging_rules", packaging, true},
		{"negative_weight", "shipping_packaging_rules", strings.ReplaceAll(packaging, "0.125", "-0.125"), false},
		{"weight_precision_loss", "shipping_packaging_rules", strings.ReplaceAll(packaging, "0.125", "0.1251"), false},
		{"dimension_precision_loss", "shipping_packaging_rules", strings.ReplaceAll(packaging, "60.25", "60.251"), false},
		{"empty_packaging_name", "shipping_packaging_rules", strings.ReplaceAll(packaging, "Original carton", " "), false},
		{"default_binding", "shipping_packaging_rule_applies", `{"rule_id":1,"product_id":2,"variant_id":null}`, true},
		{"variant_binding", "shipping_packaging_rule_applies", `{"rule_id":1,"product_id":2,"variant_id":3}`, true},
		{"zero_variant", "shipping_packaging_rule_applies", `{"rule_id":1,"product_id":2,"variant_id":0}`, false},
		{"missing_product", "shipping_packaging_rule_applies", `{"rule_id":1,"product_id":0}`, false},
		{"original_carrier", "carriers", `{"name":"Original carrier","code":"original"}`, true},
		{"empty_carrier_code", "carriers", `{"name":"Original carrier","code":" "}`, false},
		{"original_zone", "shipping_zones", `{"name":"Original zone","countries":"[\"US\"]","states":"[]","postal_codes":"[\"100*\"]"}`, true},
		{"zone_wrong_json_type", "shipping_zones", `{"name":"Zone","countries":"{}","states":"[]","postal_codes":"[]"}`, false},
		{"zone_country_wrong_case", "shipping_zones", `{"name":"Zone","countries":"[\"us\"]","states":"[]","postal_codes":"[]"}`, false},
		{"zone_repeated_country", "shipping_zones", `{"name":"Zone","countries":"[\"US\",\"US\"]","states":"[]","postal_codes":"[]"}`, false},
		{"zone_null_state_array", "shipping_zones", `{"name":"Zone","countries":"[]","states":"null","postal_codes":"[]"}`, false},
		{"original_customs", "customs_classification_profiles", customs, true},
		{"customs_wrong_hs_length", "customs_classification_profiles", strings.ReplaceAll(customs, "871499\"", "87149\""), false},
		{"customs_formatted_hs", "customs_classification_profiles", strings.ReplaceAll(customs, "871499\"", "8714.99\""), false},
		{"customs_wrong_cn_length", "customs_classification_profiles", strings.ReplaceAll(customs, "87149990", "8714999"), false},
		{"customs_wrong_origin", "customs_classification_profiles", strings.ReplaceAll(customs, `"CN"`, `"cn"`), false},
		{"customs_unknown_status", "customs_classification_profiles", strings.ReplaceAll(customs, `"active"`, `"unknown"`), false},
		{"customs_unknown_risk", "customs_classification_profiles", strings.ReplaceAll(customs, `"high"`, `"unknown"`), false},
		{"customs_null_risk_tags", "customs_classification_profiles", strings.ReplaceAll(customs, `["eu_anti_dumping_attention"]`, `null`), false},
		{"customs_unknown_risk_tag", "customs_classification_profiles", strings.ReplaceAll(customs, "eu_anti_dumping_attention", "unknown"), false},
		{"customs_repeated_risk_tag", "customs_classification_profiles", strings.ReplaceAll(customs, `["eu_anti_dumping_attention"]`, `["eu_anti_dumping_attention","eu_anti_dumping_attention"]`), false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			row := json.RawMessage(sample.row)
			err := validateConfigurationModuleMissingShippingReferencePackagingAndCustomsDomainRows(map[string][]json.RawMessage{sample.table: {row}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
			require.Equal(t, sample.row, string(row))
		})
	}
	for _, table := range strings.Fields("shipping_templates shipping_rules shipping_carrier_services shipping_route_groups shipping_route_group_carrier_services shipping_fpx_channels shipping_fpx_api_configs shipping_yanwen_published_channels shipping_tracking_providers product_shipping_template_bindings orders shipping_quote_snapshots") {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, table), table)
	}
}

func createConfigurationShippingReferencePackagingAndCustomsRestoreIntegrationFixture(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO carriers(id,name,code,enabled) VALUES(900120,'Restored carrier','CONFIGURATION-RESTORE',true),(900121,'Preserved disabled carrier','CONFIGURATION-DISABLED',false);
INSERT INTO shipping_zones(id,name,countries,states,postal_codes) VALUES(900120,'Restored delivery zone','["US","CA"]','[]','["100*"]');
INSERT INTO customs_classification_profiles(id,name,slug,hs_code,cn_code,country_of_origin,source_url,source_url_us,source_url_eu,source_url_uk,verified_at,review_due_at,trade_remedy_risk_level,trade_remedy_risk_tags_json,trade_remedy_declaration_advice)
VALUES(900120,'Restored customs profile','configuration-restored-customs','871499','87149990','CN','https://example.invalid/original-us','https://example.invalid/original-us','https://example.invalid/original-eu','https://example.invalid/original-uk','2026-01-01','2027-01-01','high','["eu_anti_dumping_attention"]','Keep original declaration advice');
INSERT INTO products(id,name,slug,price_minor,currency) VALUES(900122,'Other preserved product','configuration-other-packaging-product',123,'USD');
INSERT INTO product_variants(id,product_id,sku,price_minor,currency,stock) VALUES(900122,900122,'configuration-other-packaging-sku',123,'USD',7);
INSERT INTO shipping_packaging_rules(id,rule_name,box_weight,box_length,box_width,box_height,max_weight,is_active)
VALUES(900120,'Restored default carton',0.125,50.25,30,10,10,true),(900121,'Restored SKU carton',0.375,60.25,40,20,20,true),(900122,'Preserved inactive carton',0,0,0,0,0,false);
INSERT INTO shipping_packaging_rule_applies(id,rule_id,product_id,variant_id) VALUES(900120,900120,900001,null),(900121,900121,900001,900001);
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"carrier_empty_code":                `UPDATE carriers SET code=' ' WHERE id=900120`,
		"zone_invalid_json":                 `UPDATE shipping_zones SET countries='{}' WHERE id=900120`,
		"zone_invalid_country":              `UPDATE shipping_zones SET countries='["USA"]' WHERE id=900120`,
		"packaging_negative_dimension":      `UPDATE shipping_packaging_rules SET box_length=-1 WHERE id=900121`,
		"packaging_invalid_existing_parent": `UPDATE shipping_packaging_rules SET rule_name=' ' WHERE id=900121`,
		"packaging_wrong_variant_product":   `UPDATE shipping_packaging_rule_applies SET variant_id=900122 WHERE id=900121`,
		"packaging_missing_source_product":  `UPDATE shipping_packaging_rule_applies SET product_id=999999 WHERE id=900121`,
		"packaging_missing_source_variant":  `UPDATE shipping_packaging_rule_applies SET variant_id=999999 WHERE id=900121`,
		"packaging_deleted_source_product":  `UPDATE products SET deleted_at=now() WHERE id=900001`,
		"packaging_deleted_source_variant":  `UPDATE product_variants SET deleted_at=now() WHERE id=900001`,
		"customs_invalid_hs_code":           `UPDATE customs_classification_profiles SET hs_code='87149' WHERE id=900120`,
		"customs_invalid_risk_tag":          `UPDATE customs_classification_profiles SET trade_remedy_risk_tags_json='["unknown"]' WHERE id=900120`,
		"customs_repeated_risk_tag":         `UPDATE customs_classification_profiles SET trade_remedy_risk_tags_json='["eu_anti_dumping_attention","eu_anti_dumping_attention"]' WHERE id=900120`,
		"customs_invalid_status":            `UPDATE customs_classification_profiles SET status='unknown' WHERE id=900120`,
	}[scenario]
	if mutation != "" {
		if scenario == "customs_invalid_hs_code" || scenario == "customs_invalid_status" {
			var before, after string
			require.NoError(t, database.QueryRowContext(context.Background(), `SELECT to_jsonb(p)::text FROM customs_classification_profiles p WHERE id=900120`).Scan(&before))
			_, err = database.ExecContext(context.Background(), mutation)
			var constraintError *pq.Error
			require.ErrorAs(t, err, &constraintError)
			require.Equal(t, pq.ErrorCode("23514"), constraintError.Code)
			expectedConstraint := "ck_customs_classification_hs_code"
			if scenario == "customs_invalid_status" {
				expectedConstraint = "ck_customs_classification_status"
			}
			require.Equal(t, expectedConstraint, constraintError.Constraint)
			require.NoError(t, database.QueryRowContext(context.Background(), `SELECT to_jsonb(p)::text FROM customs_classification_profiles p WHERE id=900120`).Scan(&after))
			require.Equal(t, before, after)
			return
		}
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationShippingReferencePackagingAndCustomsFixtureRowsAndPreserveCurrentData(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM shipping_packaging_rule_applies WHERE id IN(900120,900121);DELETE FROM shipping_packaging_rules WHERE id IN(900120,900121,900122);DELETE FROM customs_classification_profiles WHERE id=900120;DELETE FROM shipping_zones WHERE id=900120;DELETE FROM carriers WHERE id IN(900120,900121);`, 9
	if scenario == "partial_packaging_variant_binding_restore" || scenario == "packaging_invalid_existing_parent" || scenario == "partial_packaging_target_only_wrong_variant" {
		statement, count = `DELETE FROM shipping_packaging_rule_applies WHERE id=900121;`, 1
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO carriers(id,name,code) VALUES(900129,'Preserved new current carrier','CONFIGURATION-CURRENT-EXTRA');
INSERT INTO shipping_packaging_rules(id,rule_name,box_weight) VALUES(900129,'Preserved new current carton',0.625);`)
	require.NoError(t, err)
	mutation := map[string]string{
		"carrier_natural_identity_conflict":           `INSERT INTO carriers(id,name,code) VALUES(900128,'New current carrier identity','CONFIGURATION-RESTORE')`,
		"customs_natural_identity_conflict":           `INSERT INTO customs_classification_profiles(id,name,slug,hs_code) VALUES(900128,'New current customs identity','configuration-restored-customs','871499')`,
		"packaging_current_default_binding_conflict":  `INSERT INTO shipping_packaging_rule_applies(id,rule_id,product_id) VALUES(900128,900129,900001)`,
		"packaging_current_variant_binding_conflict":  `INSERT INTO shipping_packaging_rule_applies(id,rule_id,product_id,variant_id) VALUES(900128,900129,900001,900001)`,
		"packaging_changed_read_only_product":         `UPDATE products SET name='Changed current product identity' WHERE id=900001`,
		"packaging_changed_read_only_variant":         `UPDATE product_variants SET sku='changed-current-packaging-sku' WHERE id=900001`,
		"partial_packaging_target_only_wrong_variant": `INSERT INTO shipping_packaging_rule_applies(id,rule_id,product_id,variant_id) VALUES(900128,900121,900001,900122)`,
		"changed_existing_packaging_rule":             `INSERT INTO shipping_packaging_rules(id,rule_name,box_weight) VALUES(900121,'Changed current carton',0.875)`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationShippingReferencePackagingAndCustomsIntegrationScenarioRequiresRejection(scenario string) bool {
	switch scenario {
	case "carrier_empty_code", "zone_invalid_json", "zone_invalid_country", "packaging_negative_dimension", "packaging_invalid_existing_parent", "packaging_wrong_variant_product", "packaging_missing_source_product", "packaging_missing_source_variant", "packaging_deleted_source_product", "packaging_deleted_source_variant", "customs_invalid_hs_code", "customs_invalid_risk_tag", "customs_repeated_risk_tag", "customs_invalid_status", "carrier_natural_identity_conflict", "customs_natural_identity_conflict", "packaging_current_default_binding_conflict", "packaging_current_variant_binding_conflict", "packaging_changed_read_only_product", "packaging_changed_read_only_variant", "partial_packaging_target_only_wrong_variant", "changed_existing_packaging_rule":
		return true
	}
	return false
}

func verifyRestoredShippingReferencePackagingAndCustomsThroughBusinessRepositoriesWithoutExternalCallsOrOrderWrites(t *testing.T, database *sql.DB) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	shippingRepository := repository.NewShippingRepository(gormDatabase)
	shippingService := service.NewShippingService(shippingRepository)
	carrier, err := shippingService.GetCarrier(900120)
	require.NoError(t, err)
	require.Equal(t, "CONFIGURATION-RESTORE", carrier.Code)
	disabled, err := shippingService.GetCarrier(900121)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)
	zone, err := shippingRepository.FindZoneByID(900120)
	require.NoError(t, err)
	require.Equal(t, `["US","CA"]`, zone.Countries)
	rules, err := shippingRepository.FindActivePackagingRulesByProductIDsAndVariants([]uint{900001})
	require.NoError(t, err)
	require.Equal(t, uint(900120), rules[900001][0].ID)
	require.Equal(t, uint(900121), rules[900001][900001].ID)
	require.Equal(t, 0.125, rules[900001][0].BoxWeight)
	require.Equal(t, 0.375, rules[900001][900001].BoxWeight)
	require.Equal(t, 60.25, rules[900001][900001].BoxLength)
	inactive, err := shippingService.GetPackagingRule(900122)
	require.NoError(t, err)
	require.False(t, inactive.IsActive)
	profile, err := repository.NewCustomsClassificationRepository(gormDatabase).FindByID(900120)
	require.NoError(t, err)
	require.Equal(t, "871499", profile.HSCode)
	require.Equal(t, "87149990", profile.CNCode)
	require.Equal(t, "high", profile.TradeRemedyRiskLevel)
	require.Equal(t, []string{"eu_anti_dumping_attention"}, []string(profile.TradeRemedyRiskTags))
	require.Equal(t, "https://example.invalid/original-eu", profile.SourceURLEU)
	require.Equal(t, "2026-01-01", profile.VerifiedAt.Format("2006-01-02"))
	require.Equal(t, "2027-01-01", profile.ReviewDueAt.Format("2006-01-02"))
	var exactWeight, preservedWeight string
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT box_weight::text FROM shipping_packaging_rules WHERE id=900121`).Scan(&exactWeight))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT box_weight::text FROM shipping_packaging_rules WHERE id=900129`).Scan(&preservedWeight))
	require.Equal(t, "0.375", exactWeight)
	require.Equal(t, "0.625", preservedWeight)
}

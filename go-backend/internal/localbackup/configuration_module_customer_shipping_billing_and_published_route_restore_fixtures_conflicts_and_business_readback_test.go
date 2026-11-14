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

func TestConfigurationCustomerShippingRecoveryRejectsInvalidBillingWithoutRoundingOrNormalizingOriginalRows(t *testing.T) {
	template := `{"id":1,"name":"Customer price","type":"weight","template_kind":"carrier","is_system_managed":false,"route_group_id":null,"currency":"USD","free_shipping":false,"free_threshold_minor":0,"default_fee_minor":9007199254740993,"customer_billing_mode":"greater_of_actual_and_volumetric","customer_first_weight_grams":1000,"customer_additional_weight_grams":500,"customer_min_charge_weight_grams":0,"customer_volumetric_divisor":6000,"customer_fuel_surcharge_percent_decimal":12.123456789012345,"customer_remote_surcharge_minor":123,"customer_remote_postal_codes":"[\"100*\"]","customer_eta_min_days":2,"customer_eta_max_days":5,"free_shipping_countries":"[]"}`
	systemTemplate := strings.NewReplacer(`"type":"weight"`, `"type":"free_shipping"`, `"template_kind":"carrier"`, `"template_kind":"system_free_shipping"`, `"is_system_managed":false`, `"is_system_managed":true`, `"free_shipping":false`, `"free_shipping":true`, `"default_fee_minor":9007199254740993`, `"default_fee_minor":0`).Replace(template)
	rule := `{"id":1,"template_id":1,"currency":"USD","region":"US","min_value":0,"max_value":100.125,"min_value_minor":0,"max_value_minor":0,"fee_minor":9007199254740993,"additional_minor":123}`
	for _, sample := range []struct {
		name, table, original, column, replacement string
		valid                                      bool
	}{
		{"exact_customer_amounts", "shipping_templates", template, "", "", true},
		{"negative_customer_amount", "shipping_templates", template, "default_fee_minor", "-1", false},
		{"fractional_minor_amount", "shipping_templates", template, "default_fee_minor", "1.5", false},
		{"unrecognized_currency", "shipping_templates", template, "currency", `"ZZZ"`, false},
		{"missing_billing_mode", "shipping_templates", template, "customer_billing_mode", `""`, false},
		{"missing_volumetric_divisor", "shipping_templates", template, "customer_volumetric_divisor", "0", false},
		{"negative_fuel_percent", "shipping_templates", template, "customer_fuel_surcharge_percent_decimal", "-1", false},
		{"excessive_fuel_precision", "shipping_templates", template, "customer_fuel_surcharge_percent_decimal", "12.1234567890123456", false},
		{"eta_inverted", "shipping_templates", template, "customer_eta_max_days", "1", false},
		{"null_postal_scope", "shipping_templates", template, "customer_remote_postal_codes", `"null"`, false},
		{"invalid_system_discriminator", "shipping_templates", template, "template_kind", `"system_free_shipping"`, false},
		{"valid_system_free_shipping", "shipping_templates", systemTemplate, "", "", true},
		{"system_template_cannot_bind_route", "shipping_templates", systemTemplate, "route_group_id", "1", false},
		{"system_template_cannot_charge_fee", "shipping_templates", systemTemplate, "default_fee_minor", "1", false},
		{"system_template_cannot_lose_ownership", "shipping_templates", systemTemplate, "is_system_managed", "false", false},
		{"invalid_template_type", "shipping_templates", template, "type", `"unknown"`, false},
		{"lowercase_country", "shipping_templates", template, "free_shipping_countries", `"[\"us\"]"`, false},
		{"wildcard_country", "shipping_templates", template, "free_shipping_countries", `"[\"*\"]"`, true},
		{"exact_rule_amounts_and_thresholds", "shipping_rules", rule, "", "", true},
		{"negative_rule_threshold", "shipping_rules", rule, "min_value", "-0.001", false},
		{"inverted_rule_threshold", "shipping_rules", rule, "min_value", "101", false},
		{"nondecimal_rule_threshold", "shipping_rules", rule, "min_value", `"1/2"`, false},
		{"excessive_threshold_precision", "shipping_rules", rule, "max_value", "100.1251", false},
		{"rule_negative_fee", "shipping_rules", rule, "fee_minor", "-1", false},
		{"valid_aggregate", "shipping_route_groups", `{"id":1,"name":"Published routes","countries":"[\"US\"]"}`, "", "", true},
		{"aggregate_wrong_country_type", "shipping_route_groups", `{"id":1,"name":"Published routes","countries":"{}"}`, "", "", false},
		{"valid_composite_membership", "shipping_route_group_carrier_services", `{"route_group_id":1,"carrier_service_id":2}`, "", "", true},
		{"invalid_service_identity", "shipping_route_group_carrier_services", `{"route_group_id":1,"carrier_service_id":0}`, "", "", false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			row := json.RawMessage(sample.original)
			if sample.column != "" {
				var object map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(row, &object))
				object[sample.column] = json.RawMessage(sample.replacement)
				var err error
				row, err = json.Marshal(object)
				require.NoError(t, err)
			}
			before := string(row)
			err := validateConfigurationModuleMissingCustomerShippingTemplatesRulesAndRouteMembershipDomainRows(map[string][]json.RawMessage{sample.table: {row}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
			require.Equal(t, before, string(row))
		})
	}
	for _, table := range strings.Fields("shipping_tracking_shipments tracking_events product_shipping_template_bindings shipping_quote_snapshots orders") {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, table), table)
	}
}

func createConfigurationCustomerShippingBillingAndPublishedRouteRestoreIntegrationFixture(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO carriers(id,name,code,enabled) VALUES(900140,'Preserved route carrier','CONFIGURATION-ROUTE-CARRIER',true);
INSERT INTO shipping_carrier_services(id,carrier_id,service_code,service_name,countries,currency,remote_surcharge_minor,enabled)
VALUES(900140,900140,'CONFIGURATION-ROUTE','Preserved provider service','["US","CA"]','EUR',999999,true);
INSERT INTO shipping_route_groups(id,name,countries,enabled) VALUES(900140,'Restored route aggregate','["US","CA"]',true);
INSERT INTO shipping_route_group_carrier_services(route_group_id,carrier_service_id,sort_order) VALUES(900140,900140,7);
INSERT INTO shipping_templates(id,name,type,template_kind,route_group_id,currency,default_fee_minor,customer_billing_mode,customer_first_weight_grams,customer_additional_weight_grams,customer_fuel_surcharge_percent_decimal,customer_remote_surcharge_minor,customer_remote_postal_codes,customer_eta_min_days,customer_eta_max_days,enabled)
VALUES(900140,'Restored customer weight price','weight','carrier',900140,'USD',1299,'greater_of_actual_and_volumetric',1000,500,12.123456789012345,123,'["100*"]',2,5,true),
 (900141,'Restored disabled price template','price','carrier',null,'USD',0,'actual_weight',0,0,0,0,'[]',0,0,false);
INSERT INTO shipping_rules(id,template_id,region,currency,min_value,max_value,min_value_minor,max_value_minor,fee_minor,additional_minor)
VALUES(900140,900140,'US','USD',0,100.125,0,0,9007199254740993,123),
 (900141,900141,'CA','USD',0,0,100,10000,199,0);
`)
	require.NoError(t, err)
	if strings.Contains(scenario, "shipping_fpx_") {
		_, err = database.ExecContext(context.Background(), `
INSERT INTO shipping_fpx_channels(id,service_code,display_name,countries,enabled,environment) VALUES(900140,'CONFIGURATION-FPX','Preserved published 4PX channel','["US","CA"]',true,'production');
INSERT INTO carriers(name,code,enabled) VALUES('Published 4PX','4PX',true) ON CONFLICT(code) DO NOTHING;
UPDATE shipping_carrier_services SET carrier_id=(SELECT id FROM carriers WHERE code='4PX'),fpx_channel_id=900140 WHERE id=900140;`)
		require.NoError(t, err)
	}
	if strings.Contains(scenario, "shipping_yanwen_") {
		_, err = database.ExecContext(context.Background(), `
INSERT INTO shipping_yanwen_published_channels(id,product_code,display_name,countries,enabled,environment) VALUES(900140,'CONFIGURATION-YANWEN','Preserved published Yanwen channel','["US","CA"]',true,'production');
INSERT INTO carriers(name,code,enabled) VALUES('Published Yanwen','YANWEN',true) ON CONFLICT(code) DO NOTHING;
UPDATE shipping_carrier_services SET carrier_id=(SELECT id FROM carriers WHERE code='YANWEN'),yanwen_published_channel_id=900140,service_code='YANWEN:CONFIGURATION-YANWEN' WHERE id=900140;`)
		require.NoError(t, err)
	}
	mutation := map[string]string{
		"shipping_invalid_template_type":                  `UPDATE shipping_templates SET type='unknown' WHERE id=900140`,
		"shipping_invalid_fuel_percent":                   `UPDATE shipping_templates SET customer_fuel_surcharge_percent_decimal=101 WHERE id=900140`,
		"shipping_invalid_postal_json":                    `UPDATE shipping_templates SET customer_remote_postal_codes='{}' WHERE id=900140`,
		"shipping_rule_currency_mismatch":                 `UPDATE shipping_rules SET currency='EUR' WHERE id=900140`,
		"shipping_rule_inverted_threshold":                `UPDATE shipping_rules SET min_value=101 WHERE id=900140`,
		"shipping_rule_missing_weight_step":               `UPDATE shipping_templates SET customer_additional_weight_grams=0 WHERE id=900140`,
		"shipping_system_template_has_rules":              `UPDATE shipping_rules SET template_id=(SELECT id FROM shipping_templates WHERE template_kind='system_free_shipping' AND deleted_at IS NULL LIMIT 1) WHERE id=900140`,
		"shipping_partial_rule_invalid_existing_template": `UPDATE shipping_templates SET customer_additional_weight_grams=0 WHERE id=900140`,
		"shipping_partial_member_invalid_existing_group":  `UPDATE shipping_route_groups SET countries='{}' WHERE id=900140`,
		"shipping_empty_route_aggregate":                  `DELETE FROM shipping_route_group_carrier_services WHERE route_group_id=900140`,
		"shipping_deleted_source_service":                 `UPDATE shipping_carrier_services SET deleted_at=now() WHERE id=900140`,
		"shipping_wrong_yanwen_provider_identity":         `UPDATE shipping_carrier_services SET service_code='YANWEN:WRONG-PROVIDER' WHERE id=900140`,
		"shipping_deleted_source_group":                   `UPDATE shipping_route_groups SET deleted_at=now() WHERE id=900140`,
		"shipping_fpx_test_channel":                       `UPDATE shipping_fpx_channels SET environment='test' WHERE id=900140`,
		"shipping_fpx_disabled_channel":                   `UPDATE shipping_fpx_channels SET enabled=false WHERE id=900140`,
		"shipping_yanwen_fat_channel":                     `UPDATE shipping_yanwen_published_channels SET environment='fat' WHERE id=900140`,
		"shipping_yanwen_disabled_channel":                `UPDATE shipping_yanwen_published_channels SET enabled=false WHERE id=900140`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationCustomerShippingBillingAndRouteFixtureRowsAndPreserveProviderData(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM shipping_rules WHERE id IN(900140,900141);DELETE FROM shipping_templates WHERE id IN(900140,900141);DELETE FROM shipping_route_group_carrier_services WHERE route_group_id=900140;DELETE FROM shipping_route_groups WHERE id=900140;`, 6
	if scenario == "shipping_missing_service_restored" {
		count++
	}
	if scenario == "shipping_empty_route_aggregate" {
		count = 5
	}
	if strings.HasPrefix(scenario, "shipping_partial_rule_") {
		statement, count = `DELETE FROM shipping_rules WHERE id=900140;`, 1
	}
	if strings.HasPrefix(scenario, "shipping_partial_member_") {
		statement, count = `DELETE FROM shipping_route_group_carrier_services WHERE route_group_id=900140;`, 1
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO shipping_templates(id,name,type,currency,enabled) VALUES(900149,'Preserved new customer price','quantity','USD',false);`)
	require.NoError(t, err)
	mutation := map[string]string{
		"shipping_changed_existing_template":                  `INSERT INTO shipping_templates(id,name,type,currency) VALUES(900140,'Changed current price','weight','USD')`,
		"shipping_changed_read_only_service":                  `UPDATE shipping_carrier_services SET service_name='Changed current provider identity' WHERE id=900140`,
		"shipping_missing_service_restored":                   `DELETE FROM shipping_carrier_services WHERE id=900140`,
		"shipping_partial_rule_target_only_currency_conflict": `INSERT INTO shipping_rules(id,template_id,currency,fee_minor) VALUES(900149,900140,'EUR',1)`,
		"shipping_partial_member_target_only_deleted_service": `INSERT INTO shipping_carrier_services(id,carrier_id,service_code,service_name,countries,currency,deleted_at) VALUES(900149,900140,'TARGET-ONLY-DELETED','Deleted target service','[]','USD',now()); INSERT INTO shipping_route_group_carrier_services(route_group_id,carrier_service_id) VALUES(900140,900149)`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationCustomerShippingBillingAndPublishedRouteIntegrationScenarioRequiresRejection(scenario string) bool {
	return strings.HasPrefix(scenario, "shipping_") && scenario != "shipping_missing_service_restored" && scenario != "shipping_partial_rule_restore" && scenario != "shipping_partial_member_restore" && scenario != "shipping_fpx_production_channel_restore" && scenario != "shipping_yanwen_production_channel_restore"
}

func verifyRestoredCustomerShippingBillingAndRouteProjectionThroughReadOnlyBusinessServicesWithoutRepricingOrders(t *testing.T, database *sql.DB) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	shippingService := service.NewShippingService(repository.NewShippingRepository(gormDatabase))
	shippingService.ConfigureYanwenPublishedCollectionService(service.NewYanwenPublishedCollectionService(repository.NewYanwenPublishedChannelRepository(gormDatabase), nil))
	template, err := shippingService.GetTemplate(900140)
	require.NoError(t, err)
	require.EqualValues(t, 1299, template.DefaultFeeMinor)
	require.Equal(t, "USD", template.Currency)
	require.Equal(t, "12.123456789012345", template.CustomerFuelSurchargePercentDecimal)
	require.Equal(t, 500, template.CustomerAdditionalWeightGrams)
	require.Equal(t, `["100*"]`, template.CustomerRemotePostalCodes)
	require.Len(t, template.Rules, 1)
	require.EqualValues(t, 9007199254740993, template.Rules[0].FeeMinor)
	require.Equal(t, 100.125, template.Rules[0].MaxValue)
	group, err := shippingService.GetShippingRouteGroup(900140)
	require.NoError(t, err)
	require.Len(t, group.CarrierServices, 1)
	require.JSONEq(t, `["CA","US"]`, group.Countries)
	require.EqualValues(t, 900140, group.CarrierServices[0].ID)
	require.Equal(t, "EUR", group.CarrierServices[0].Currency)
	disabled, err := shippingService.GetTemplate(900141)
	require.NoError(t, err)
	require.False(t, disabled.Enabled)
	var providerCost, sortOrder int64
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT remote_surcharge_minor FROM shipping_carrier_services WHERE id=900140`).Scan(&providerCost))
	require.EqualValues(t, 999999, providerCost)
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT sort_order FROM shipping_route_group_carrier_services WHERE route_group_id=900140 AND carrier_service_id=900140`).Scan(&sortOrder))
	require.EqualValues(t, 7, sortOrder)
}

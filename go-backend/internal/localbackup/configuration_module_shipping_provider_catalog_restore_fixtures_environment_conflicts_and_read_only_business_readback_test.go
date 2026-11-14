package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestConfigurationShippingProviderCatalogRecoveryPreservesOriginalFieldsAndRejectsInvalidEnvironmentIdentityAndBilling(t *testing.T) {
	carrierService := `{"id":1,"carrier_id":2,"fpx_channel_id":null,"yanwen_published_channel_id":null,"service_code":"ORIGINAL","service_name":"Original service","countries":"[\"US\"]","currency":"EUR","billing_mode":"actual_weight","first_weight_grams":1000,"additional_weight_grams":500,"min_charge_weight_grams":0,"volumetric_divisor":6000,"remote_surcharge_minor":9007199254740993,"remote_postal_codes":"[\"100*\"]","eta_min_days":2,"eta_max_days":5,"fuel_surcharge_percent_decimal":7.123456789012345}`
	fpx := `{"id":1,"environment":"production","service_code":"ORIGINAL","display_name":"Original channel","countries":"[\"US\"]"}`
	yanwen := `{"id":1,"environment":"production","product_code":"ORIGINAL","display_name":"Original channel","countries":"[\"US\"]","max_weight_grams":10000,"volumetric_divisor":8000}`
	product := `{"id":1,"environment":"production","product_id":"ORIGINAL","name_en":"Original product","last_synced_at":"2025-01-02T03:04:05Z"}`
	country := `{"id":1,"environment":"production","country_id":"USA","country_code":"US","name_en":"United States","last_synced_at":"2025-01-02T03:04:05Z"}`
	warehouse := `{"id":1,"environment":"production","warehouse_code":"WH-ORIGINAL","name":"Original warehouse","last_synced_at":"2025-01-02T03:04:05Z"}`
	for _, sample := range []struct {
		name, table, row, column, replacement string
		valid                                 bool
	}{
		{"exact_provider_fields", "shipping_carrier_services", carrierService, "", "", true},
		{"negative_provider_fee", "shipping_carrier_services", carrierService, "remote_surcharge_minor", "-1", false},
		{"fractional_provider_fee", "shipping_carrier_services", carrierService, "remote_surcharge_minor", "1.5", false},
		{"invalid_provider_currency", "shipping_carrier_services", carrierService, "currency", `"ZZZ"`, false},
		{"provider_fuel_precision_loss", "shipping_carrier_services", carrierService, "fuel_surcharge_percent_decimal", "7.1234567890123456", false},
		{"invalid_provider_divisor", "shipping_carrier_services", carrierService, "volumetric_divisor", "0", false},
		{"inverted_provider_eta", "shipping_carrier_services", carrierService, "eta_max_days", "1", false},
		{"invalid_provider_postal_json", "shipping_carrier_services", carrierService, "remote_postal_codes", `"{}"`, false},
		{"unbound_carrier", "shipping_carrier_services", carrierService, "carrier_id", "0", false},
		{"original_production_fpx", "shipping_fpx_channels", fpx, "", "", true},
		{"original_test_fpx", "shipping_fpx_channels", fpx, "environment", `"test"`, true},
		{"wrong_fpx_environment", "shipping_fpx_channels", fpx, "environment", `"fat"`, false},
		{"implicit_fpx_environment", "shipping_fpx_channels", fpx, "environment", `""`, false},
		{"fpx_code_requires_normalization", "shipping_fpx_channels", fpx, "service_code", `" ORIGINAL "`, false},
		{"malformed_fpx_countries", "shipping_fpx_channels", fpx, "countries", `"{}"`, false},
		{"original_production_yanwen", "shipping_yanwen_published_channels", yanwen, "", "", true},
		{"original_fat_yanwen", "shipping_yanwen_published_channels", yanwen, "environment", `"fat"`, true},
		{"yanwen_environment_alias", "shipping_yanwen_published_channels", yanwen, "environment", `"test"`, false},
		{"yanwen_negative_weight", "shipping_yanwen_published_channels", yanwen, "max_weight_grams", "-1", false},
		{"yanwen_invalid_divisor", "shipping_yanwen_published_channels", yanwen, "volumetric_divisor", "0", false},
		{"original_product_snapshot", "shipping_yanwen_product_catalog_entries", product, "", "", true},
		{"empty_product_name", "shipping_yanwen_product_catalog_entries", product, "name_en", `""`, false},
		{"missing_product_sync_time", "shipping_yanwen_product_catalog_entries", product, "last_synced_at", `"0001-01-01T00:00:00Z"`, false},
		{"invalid_product_environment", "shipping_yanwen_product_catalog_entries", product, "environment", `"test"`, false},
		{"original_country_snapshot", "shipping_yanwen_country_catalog_entries", country, "", "", true},
		{"country_requires_normalization", "shipping_yanwen_country_catalog_entries", country, "country_code", `"us"`, false},
		{"empty_country_name", "shipping_yanwen_country_catalog_entries", country, "name_en", `""`, false},
		{"original_warehouse_snapshot", "shipping_yanwen_warehouse_catalog_entries", warehouse, "", "", true},
		{"empty_warehouse_name", "shipping_yanwen_warehouse_catalog_entries", warehouse, "name", `""`, false},
	} {
		t.Run(sample.name, func(t *testing.T) {
			row := json.RawMessage(sample.row)
			if sample.column != "" {
				var object map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(row, &object))
				object[sample.column] = json.RawMessage(sample.replacement)
				var err error
				row, err = json.Marshal(object)
				require.NoError(t, err)
			}
			before := string(row)
			err := validateConfigurationModuleMissingShippingCarrierServicesPublishedChannelsAndYanwenReferenceCatalogDomainRows(map[string][]json.RawMessage{sample.table: {row}})
			require.Equal(t, sample.valid, err == nil, "%v", err)
			require.Equal(t, before, string(row))
		})
	}
	for _, table := range strings.Fields("shipping_tracking_shipments tracking_events orders shipping_quote_snapshots") {
		require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, table))
	}
}

func configurationShippingProviderCatalogRestoreIntegrationScenarios() []string {
	return []string{"provider_catalog_partial_yanwen_product_restore", "provider_catalog_partial_fpx_test_channel_restore", "provider_catalog_partial_fpx_service_and_channel_restore", "provider_catalog_partial_yanwen_service_and_channel_restore", "provider_catalog_partial_country_and_warehouse_restore", "provider_catalog_invalid_fpx_environment", "provider_catalog_invalid_product_name", "provider_catalog_invalid_country_code", "provider_catalog_invalid_warehouse_name", "provider_catalog_missing_original_sync_time", "provider_catalog_negative_published_weight", "provider_catalog_invalid_service_postal_json", "provider_catalog_invalid_service_currency", "provider_catalog_service_wrong_provider", "provider_catalog_service_missing_stable_channel", "provider_catalog_missing_official_product", "provider_catalog_wrong_official_environment", "provider_catalog_partial_channel_invalid_existing_product", "provider_catalog_current_product_case_alias", "provider_catalog_changed_existing_channel", "provider_catalog_current_service_natural_identity_conflict", "provider_catalog_current_channel_natural_identity_conflict", "provider_catalog_partial_product_target_only_deleted_member", "provider_catalog_partial_product_target_only_rule_currency_conflict", "provider_catalog_disabled_production_channel", "provider_catalog_deleted_carrier", "legacy_v7_policy_requires_rehearsal"}
}

func createConfigurationShippingProviderCatalogRestoreIntegrationFixture(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO carriers(id,name,code) VALUES(900160,'Restored generic provider carrier','PROVIDER-CATALOG-GENERIC');
INSERT INTO carriers(name,code) VALUES('Restored 4PX provider','4PX'),('Restored Yanwen provider','YANWEN') ON CONFLICT(code) DO NOTHING;
INSERT INTO shipping_yanwen_product_catalog_entries(id,environment,product_id,name_en,last_synced_at) VALUES
 (900160,'production','PROVIDER-CATALOG-YANWEN','Original production product','2025-01-02T03:04:05Z'),
 (900161,'fat','PROVIDER-CATALOG-FAT','Original FAT product','2025-01-02T03:04:05Z');
INSERT INTO shipping_yanwen_country_catalog_entries(id,environment,country_id,country_code,name_en,last_synced_at) VALUES(900160,'production','USA','US','Original US reference','2025-01-02T03:04:05Z');
INSERT INTO shipping_yanwen_warehouse_catalog_entries(id,environment,warehouse_code,name,area,last_synced_at) VALUES(900160,'production','PROVIDER-WH','Original warehouse','Original area','2025-01-02T03:04:05Z');
INSERT INTO shipping_fpx_channels(id,environment,service_code,display_name,countries,enabled) VALUES
 (900160,'production','PROVIDER-CATALOG-FPX','Original 4PX channel','["US","CA"]',true),
 (900161,'test','PROVIDER-CATALOG-FPX-TEST','Preserved disabled test channel','["US"]',false);
INSERT INTO shipping_yanwen_published_channels(id,environment,product_code,display_name,countries,enabled,max_weight_grams,require_receiver_tax_number,require_ioss,require_eori,notes) VALUES
 (900160,'production','PROVIDER-CATALOG-YANWEN','Original Yanwen channel','["US","CA"]',true,10000,true,true,true,'Original declaration requirements'),
 (900161,'fat','PROVIDER-CATALOG-FAT','Preserved disabled FAT channel','["US"]',false,5000,false,false,false,'Original FAT note');
INSERT INTO shipping_carrier_services(id,carrier_id,service_code,service_name,countries,currency,billing_mode,first_weight_grams,additional_weight_grams,fuel_surcharge_percent_decimal,remote_surcharge_minor,remote_postal_codes,enabled)
VALUES(900160,900160,'PROVIDER-CATALOG-GENERIC','Original generic provider','["US","CA"]','EUR','actual_weight',1000,500,7.123456789012345,9007199254740993,'["100*"]',true),
 (900163,900160,'PROVIDER-CATALOG-DISABLED','Original disabled provider','["US"]','EUR','actual_weight',0,0,0,0,'[]',false);
INSERT INTO shipping_carrier_services(id,carrier_id,service_code,service_name,countries,currency,fpx_channel_id,yanwen_published_channel_id)
VALUES(900161,(SELECT id FROM carriers WHERE code='4PX'),'PROVIDER-CATALOG-FPX','Original 4PX service','["US","CA"]','EUR',900160,null),
 (900162,(SELECT id FROM carriers WHERE code='YANWEN'),'YANWEN:PROVIDER-CATALOG-YANWEN','Original Yanwen service','["US","CA"]','EUR',null,900160);
INSERT INTO shipping_route_groups(id,name,countries) VALUES(900160,'Provider restoration affected aggregate','["US","CA"]');
INSERT INTO shipping_route_group_carrier_services(route_group_id,carrier_service_id) VALUES(900160,900160),(900160,900161),(900160,900162);
INSERT INTO shipping_templates(id,name,type,currency,route_group_id,default_fee_minor) VALUES(900160,'Provider restoration customer price','quantity','USD',900160,1299);
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"provider_catalog_invalid_fpx_environment":                  `UPDATE shipping_fpx_channels SET environment='fat' WHERE id=900160`,
		"provider_catalog_invalid_product_name":                     `UPDATE shipping_yanwen_product_catalog_entries SET name_en='' WHERE id=900160`,
		"provider_catalog_invalid_country_code":                     `UPDATE shipping_yanwen_country_catalog_entries SET country_code='us' WHERE id=900160`,
		"provider_catalog_invalid_warehouse_name":                   `UPDATE shipping_yanwen_warehouse_catalog_entries SET name=' ' WHERE id=900160`,
		"provider_catalog_missing_original_sync_time":               `UPDATE shipping_yanwen_product_catalog_entries SET last_synced_at='0001-01-01T00:00:00Z' WHERE id=900160`,
		"provider_catalog_negative_published_weight":                `UPDATE shipping_yanwen_published_channels SET max_weight_grams=-1 WHERE id=900160`,
		"provider_catalog_invalid_service_postal_json":              `UPDATE shipping_carrier_services SET remote_postal_codes='{}' WHERE id=900160`,
		"provider_catalog_invalid_service_currency":                 `UPDATE shipping_carrier_services SET currency='ZZZ' WHERE id=900160`,
		"provider_catalog_service_wrong_provider":                   `UPDATE shipping_carrier_services SET carrier_id=900160 WHERE id=900161`,
		"provider_catalog_service_missing_stable_channel":           `UPDATE shipping_carrier_services SET fpx_channel_id=null WHERE id=900161`,
		"provider_catalog_missing_official_product":                 `DELETE FROM shipping_yanwen_product_catalog_entries WHERE id=900160`,
		"provider_catalog_wrong_official_environment":               `UPDATE shipping_yanwen_product_catalog_entries SET environment='fat' WHERE id=900160`,
		"provider_catalog_partial_channel_invalid_existing_product": `UPDATE shipping_yanwen_product_catalog_entries SET name_en='' WHERE id=900160`,
		"provider_catalog_disabled_production_channel":              `UPDATE shipping_fpx_channels SET enabled=false WHERE id=900160`,
		"provider_catalog_deleted_carrier":                          `UPDATE carriers SET deleted_at=now() WHERE id=900160`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationShippingProviderCatalogFixtureRowsAndPreserveCurrentProviderControls(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement, count := `DELETE FROM shipping_templates WHERE id=900160;DELETE FROM shipping_route_groups WHERE id=900160;DELETE FROM shipping_carrier_services WHERE id IN(900160,900161,900162,900163);DELETE FROM shipping_fpx_channels WHERE id IN(900160,900161);DELETE FROM shipping_yanwen_published_channels WHERE id IN(900160,900161);DELETE FROM shipping_yanwen_product_catalog_entries WHERE id IN(900160,900161);DELETE FROM shipping_yanwen_country_catalog_entries WHERE id=900160;DELETE FROM shipping_yanwen_warehouse_catalog_entries WHERE id=900160;`, 17
	if scenario == "provider_catalog_missing_official_product" {
		count--
	}
	switch scenario {
	case "provider_catalog_partial_yanwen_product_restore", "provider_catalog_partial_product_target_only_deleted_member", "provider_catalog_partial_product_target_only_rule_currency_conflict":
		statement, count = `DELETE FROM shipping_yanwen_product_catalog_entries WHERE id=900160;`, 1
	case "provider_catalog_partial_fpx_test_channel_restore":
		statement, count = `DELETE FROM shipping_fpx_channels WHERE id=900161;`, 1
	case "provider_catalog_partial_fpx_service_and_channel_restore":
		statement, count = `DELETE FROM shipping_carrier_services WHERE id=900161;DELETE FROM shipping_fpx_channels WHERE id=900160;`, 3
	case "provider_catalog_partial_yanwen_service_and_channel_restore", "provider_catalog_partial_channel_invalid_existing_product":
		statement, count = `DELETE FROM shipping_carrier_services WHERE id=900162;DELETE FROM shipping_yanwen_published_channels WHERE id=900160;`, 3
	case "provider_catalog_partial_country_and_warehouse_restore":
		statement, count = `DELETE FROM shipping_yanwen_country_catalog_entries WHERE id=900160;DELETE FROM shipping_yanwen_warehouse_catalog_entries WHERE id=900160;`, 2
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO shipping_yanwen_product_catalog_entries(id,environment,product_id,name_en,last_synced_at) VALUES(900169,'production','PRESERVED-CURRENT-PRODUCT','Preserved newer product','2026-01-02T03:04:05Z');`)
	require.NoError(t, err)
	mutation := map[string]string{
		"provider_catalog_current_product_case_alias":                         `INSERT INTO shipping_yanwen_product_catalog_entries(environment,product_id,name_en,last_synced_at) VALUES('production','provider-catalog-yanwen','Current case alias','2026-01-02T03:04:05Z')`,
		"provider_catalog_changed_existing_channel":                           `INSERT INTO shipping_fpx_channels(id,environment,service_code,display_name) VALUES(900160,'production','PROVIDER-CATALOG-FPX','Changed current channel')`,
		"provider_catalog_current_service_natural_identity_conflict":          `INSERT INTO shipping_carrier_services(id,carrier_id,service_code,service_name,currency) VALUES(900169,900160,'PROVIDER-CATALOG-GENERIC','Current conflicting identity','EUR')`,
		"provider_catalog_current_channel_natural_identity_conflict":          `INSERT INTO shipping_fpx_channels(id,environment,service_code,display_name) VALUES(900169,'production','PROVIDER-CATALOG-FPX','Current conflicting channel')`,
		"provider_catalog_partial_product_target_only_deleted_member":         `INSERT INTO shipping_carrier_services(id,carrier_id,service_code,service_name,currency,deleted_at) VALUES(900169,900160,'CURRENT-DELETED-MEMBER','Current deleted service','EUR',now());INSERT INTO shipping_route_group_carrier_services(route_group_id,carrier_service_id) VALUES(900160,900169)`,
		"provider_catalog_partial_product_target_only_rule_currency_conflict": `INSERT INTO shipping_rules(id,template_id,currency,fee_minor) VALUES(900169,900160,'EUR',1)`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationShippingProviderCatalogIntegrationScenarioRequiresRejection(scenario string) bool {
	return strings.HasPrefix(scenario, "provider_catalog_") && scenario != "provider_catalog_partial_yanwen_product_restore" && scenario != "provider_catalog_partial_fpx_test_channel_restore" && scenario != "provider_catalog_partial_fpx_service_and_channel_restore" && scenario != "provider_catalog_partial_yanwen_service_and_channel_restore" && scenario != "provider_catalog_partial_country_and_warehouse_restore"
}

func verifyRestoredShippingProviderCatalogsThroughReadOnlyBusinessServicesWithoutSyncPublishingOrCustomerPriceWrites(t *testing.T, database *sql.DB) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	require.NoError(t, err)
	defer tx.Rollback()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	shippingRepository := repository.NewShippingRepository(gormDatabase)
	shippingService := service.NewShippingService(shippingRepository)
	collection := service.NewYanwenPublishedCollectionService(repository.NewYanwenPublishedChannelRepository(gormDatabase), repository.NewYanwenProductCatalogRepository(gormDatabase))
	shippingService.ConfigureYanwenPublishedCollectionService(collection)
	group, err := shippingService.GetShippingRouteGroup(900160)
	require.NoError(t, err)
	require.Len(t, group.CarrierServices, 3)
	require.JSONEq(t, `["CA","US"]`, group.Countries)
	template, err := shippingService.GetTemplate(900160)
	require.NoError(t, err)
	require.Equal(t, "USD", template.Currency)
	require.EqualValues(t, 1299, template.DefaultFeeMinor)
	channel, err := collection.GetYanwenPublishedChannel(900160)
	require.NoError(t, err)
	require.True(t, channel.RequireIOSS && channel.RequireEORI && channel.RequireReceiverTaxNumber)
	require.Equal(t, "Original declaration requirements", channel.Notes)
	products, err := repository.NewYanwenProductCatalogRepository(gormDatabase).FindYanwenProductCatalogEntriesByEnvironment("production")
	require.NoError(t, err)
	var originalSyncTime time.Time
	for _, entry := range products {
		if entry.ID == 900160 {
			originalSyncTime = entry.LastSyncedAt
		}
	}
	require.True(t, originalSyncTime.Equal(time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)))
	countries, err := repository.NewYanwenCountryCatalogRepository(gormDatabase).FindYanwenCountryCatalogEntriesByEnvironment("production")
	require.NoError(t, err)
	require.Contains(t, countries[0].CountryCode, "US")
	warehouses, err := repository.NewYanwenWarehouseCatalogRepository(gormDatabase).FindYanwenWarehouseCatalogEntriesByEnvironment("production")
	require.NoError(t, err)
	require.Equal(t, "PROVIDER-WH", warehouses[0].WarehouseCode)
	var currency, fuel string
	var providerFee int64
	var disabledService, disabledTestChannel, disabledFatChannel bool
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT currency,fuel_surcharge_percent_decimal::text,remote_surcharge_minor FROM shipping_carrier_services WHERE id=900160`).Scan(&currency, &fuel, &providerFee))
	require.Equal(t, "EUR", currency)
	require.Equal(t, "7.123456789012345", fuel)
	require.EqualValues(t, 9007199254740993, providerFee)
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT enabled FROM shipping_carrier_services WHERE id=900163`).Scan(&disabledService))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT enabled FROM shipping_fpx_channels WHERE id=900161 AND environment='test'`).Scan(&disabledTestChannel))
	require.NoError(t, tx.QueryRowContext(context.Background(), `SELECT enabled FROM shipping_yanwen_published_channels WHERE id=900161 AND environment='fat'`).Scan(&disabledFatChannel))
	require.False(t, disabledService || disabledTestChannel || disabledFatChannel)
}

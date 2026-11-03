package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	paymenthttp "commerce-platform/internal/api/v1/payment"
	storefronthttp "commerce-platform/internal/api/v1/storefront"
	"commerce-platform/internal/domain/money"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestConfigurationMarketCurrencyAndTaxDomainGuardsKeepExactBackupValuesAndRefuseSilentFallbacks(t *testing.T) {
	market := `{"code":"RESTORE_NZ","name":"Restored market","default_locale":"en","supported_locales":["en","zh_cn"],"default_currency":"NZD","display_currencies":["NZD","USD"]}`
	rate := `{"base_currency":"USD","quote_currency":"NZD","rate_decimal":1.234567891234567,"fetched_at":"2020-01-01T00:00:00Z","expires_at":"2021-01-01T00:00:00Z"}`
	tax := `{"name":"Restored tax","country":"NZ","state":"","postal_code":"","rate_decimal":7.500000000000000,"compound":false}`
	for _, test := range []struct {
		name, table, row string
		valid            bool
	}{
		{"market", "storefront_markets", market, true},
		{"unsupported_locale", "storefront_markets", strings.ReplaceAll(market, `"default_locale":"en"`, `"default_locale":"xx"`), false},
		{"missing_default_currency", "storefront_markets", strings.ReplaceAll(market, `["NZD","USD"]`, `["USD"]`), false},
		{"invalid_locale_json", "storefront_markets", strings.ReplaceAll(market, `["en","zh_cn"]`, `{}`), false},
		{"exact_expired_rate", "currency_exchange_rates", rate, true},
		{"unknown_currency", "currency_exchange_rates", strings.ReplaceAll(rate, `"NZD"`, `"XYZ"`), false},
		{"zero_rate", "currency_exchange_rates", strings.ReplaceAll(rate, `1.234567891234567`, `0`), false},
		{"bad_expiry", "currency_exchange_rates", strings.ReplaceAll(rate, `2021-01-01`, `2019-01-01`), false},
		{"self_rate_ten", "currency_exchange_rates", strings.ReplaceAll(strings.ReplaceAll(rate, `"NZD"`, `"USD"`), `1.234567891234567`, `10`), false},
		{"exact_tax", "tax_rates", tax, true},
		{"invalid_tax", "tax_rates", strings.ReplaceAll(tax, `7.500000000000000`, `101`), false},
		{"compound_tax", "tax_rates", strings.ReplaceAll(tax, `false`, `true`), false},
		{"lowercase_country", "storefront_market_countries", `{"code":"nz"}`, false},
		{"sync_control", "tax_rate_source_configs", `{}`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			row := json.RawMessage(test.row)
			err := validateConfigurationModuleMissingMarketCurrencyAndTaxDomainRows(map[string][]json.RawMessage{test.table: {row}})
			require.Equal(t, test.valid, err == nil)
			require.Equal(t, test.row, string(row))
			if test.table == "tax_rate_source_configs" {
				require.False(t, configurationModuleHasMissingSiteContentWriteContract(selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}, test.table))
			}
		})
	}
}

func createConfigurationMarketCurrencyAndTaxRestoreIntegrationFixture(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO storefront_markets(id,code,name,default_locale,supported_locales,default_currency,display_currencies) VALUES(900090,'RESTORE_NZ','Restored New Zealand market','en','["en","zh_cn"]','NZD','["NZD","USD"]');
INSERT INTO storefront_market_countries(id,market_id,code) VALUES(900090,900090,'NZ');
INSERT INTO currency_exchange_rates(id,base_currency,quote_currency,rate_decimal,source,fetched_at,expires_at) VALUES(900090,'USD','NZD',1.234567891234567,'original-synthetic-rate','2026-01-01','2040-01-01'),(900091,'USD','JPY',151.123456789123456,'expired-synthetic-rate','2019-01-01','2020-01-01');
INSERT INTO tax_rates(id,name,country,state,postal_code,rate_decimal,enabled) VALUES(900090,'Restored NZ checkout tax','NZ','','',7.500000000000000,true);
INSERT INTO tax_rate_source_snapshots(id,provider_code,source_endpoint,version,content_sha256,captured_at,country_count,rate_count,country_rate_comments) VALUES(900090,'vatcomply','https://synthetic.invalid/reference-only','configuration-tax-reference',repeat('a',64),'2026-01-01',1,2,'{"GR":{"24":["Preserved original rate note"]}}');
INSERT INTO tax_rate_source_snapshot_entries(id,snapshot_id,country_code,source_country_code,country_name,currency,rate_type,rate_decimal) VALUES(900090,900090,'GR','EL','Greece','EUR','standard',24),(900091,900090,'GR','EL','Greece','EUR','reduced',13);
INSERT INTO currency_exchange_sync_leases(lease_key,owner_id,lease_expires_at) VALUES('configuration-preserved-rate-lease','original-owner','2040-01-01');
`)
	require.NoError(t, err)
	if scenario == "legacy_reference_comments_not_collected" {
		_, err = database.ExecContext(context.Background(), `UPDATE tax_rate_source_snapshots SET country_rate_comments=NULL WHERE id=900090`)
		require.NoError(t, err)
	}
	mutation := map[string]string{
		"market_invalid_locale":                  `UPDATE storefront_markets SET default_locale='xx' WHERE id=900090`,
		"market_default_currency_not_selectable": `UPDATE storefront_markets SET display_currencies='["USD"]' WHERE id=900090`,
		"market_country_wrong_case":              `UPDATE storefront_market_countries SET code='nz' WHERE id=900090`,
		"market_country_soft_deleted_parent":     `UPDATE storefront_markets SET deleted_at=now() WHERE id=900090`,
		"exchange_unknown_currency":              `UPDATE currency_exchange_rates SET quote_currency='XYZ' WHERE id=900090`,
		"exchange_invalid_expiry":                `UPDATE currency_exchange_rates SET expires_at='2025-01-01' WHERE id=900090`,
		"tax_unsupported_compound":               `UPDATE tax_rates SET compound=true WHERE id=900090`,
		"tax_invalid_location":                   `UPDATE tax_rates SET country='nz' WHERE id=900090`,
		"reference_count_mismatch":               `UPDATE tax_rate_source_snapshots SET rate_count=3 WHERE id=900090`,
		"reference_invalid_comment_shape":        `UPDATE tax_rate_source_snapshots SET country_rate_comments='{"GR":{"24":[123]}}' WHERE id=900090`,
		"reference_invalid_entry_type":           `UPDATE tax_rate_source_snapshot_entries SET rate_type='unknown' WHERE id=900090`,
		"reference_missing_entry":                `DELETE FROM tax_rate_source_snapshot_entries WHERE id=900091`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
}

func deleteConfigurationMarketCurrencyAndTaxFixtureRowsAndPreserveNewCurrentRules(t *testing.T, database *sql.DB, scenario string) int {
	t.Helper()
	statement := `DELETE FROM storefront_markets WHERE id=900090;DELETE FROM currency_exchange_rates WHERE id IN(900090,900091);DELETE FROM tax_rates WHERE id=900090;DELETE FROM tax_rate_source_snapshots WHERE id=900090;`
	count := 8
	switch scenario {
	case "partial_market_country_restore":
		statement = `DELETE FROM storefront_market_countries WHERE id=900090;`
		count = 1
	case "partial_reference_entry_restore", "partial_reference_target_only_extra_entry":
		statement = `DELETE FROM tax_rate_source_snapshot_entries WHERE id=900091;`
		count = 1
	case "legacy_reference_comments_not_collected":
		// Separate fixture mutation is applied before export below.
	}
	_, err := database.ExecContext(context.Background(), statement+`
INSERT INTO tax_rates(id,name,country,state,postal_code,rate_decimal,enabled) VALUES(900099,'Preserved current unrelated tax','AU','','',10,true);
INSERT INTO currency_exchange_rates(id,base_currency,quote_currency,rate_decimal,source,fetched_at,expires_at) VALUES(900099,'USD','CAD',1.4,'preserved-current-rate','2026-01-01','2040-01-01');
`)
	require.NoError(t, err)
	mutation := map[string]string{
		"market_country_current_owner_conflict":     `INSERT INTO storefront_market_countries(id,market_id,code) SELECT 900099,id,'NZ' FROM storefront_markets WHERE code='US' AND deleted_at IS NULL`,
		"tax_newer_current_scope_conflict":          `INSERT INTO tax_rates(id,name,country,state,postal_code,rate_decimal,priority,enabled) VALUES(900098,'Current NZ decision','NZ','','',12,100,true)`,
		"exchange_current_reverse_pair_conflict":    `INSERT INTO currency_exchange_rates(id,base_currency,quote_currency,rate_decimal,source,fetched_at) VALUES(900098,'NZD','USD',0.8,'new-current-reverse-rate','2026-01-01')`,
		"changed_existing_exchange_rate":            `INSERT INTO currency_exchange_rates(id,base_currency,quote_currency,rate_decimal,source,fetched_at) VALUES(900090,'USD','NZD',1.7,'newer-current-rate','2026-01-01')`,
		"missing_tax_sync_configuration":            `DELETE FROM tax_rate_source_configs WHERE id=1`,
		"partial_reference_target_only_extra_entry": `INSERT INTO tax_rate_source_snapshot_entries(id,snapshot_id,country_code,source_country_code,country_name,currency,rate_type,rate_decimal) VALUES(900098,900090,'GR','EL','Greece','EUR','parking',6)`,
	}[scenario]
	if mutation != "" {
		_, err = database.ExecContext(context.Background(), mutation)
		require.NoError(t, err)
	}
	return count
}

func configurationMarketCurrencyAndTaxIntegrationScenarioRequiresRejection(scenario string) bool {
	switch scenario {
	case "market_invalid_locale", "market_default_currency_not_selectable", "market_country_wrong_case", "market_country_soft_deleted_parent", "exchange_unknown_currency", "exchange_invalid_expiry", "tax_unsupported_compound", "tax_invalid_location", "reference_count_mismatch", "reference_invalid_comment_shape", "reference_invalid_entry_type", "reference_missing_entry", "market_country_current_owner_conflict", "tax_newer_current_scope_conflict", "exchange_current_reverse_pair_conflict", "changed_existing_exchange_rate", "missing_tax_sync_configuration", "partial_reference_target_only_extra_entry":
		return true
	}
	return false
}

func verifyRestoredConfigurationMarketsPublicTaxAndExactExchangeWithoutChangingOrdersExpiryOrSyncControls(t *testing.T, database *sql.DB, scenario string) {
	t.Helper()
	gormDatabase, err := gorm.Open(postgres.New(postgres.Config{Conn: database}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	settings := repository.NewSettingRepository(gormDatabase)
	marketService := service.NewStorefrontMarketService(repository.NewStorefrontMarketRepository(gormDatabase))
	contextService := service.NewStorefrontContextServiceWithMarkets(service.NewCurrencyPolicyService(settings), marketService)
	taxService := service.NewTaxRateService(repository.NewTaxRateRepository(gormDatabase))
	paymentHandler := paymenthttp.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	paymentHandler.ConfigureTaxRateService(taxService)
	router := gin.New()
	router.GET("/storefront/context", storefronthttp.NewContextHandler(contextService).GetContext)
	router.GET("/payment/tax-rates/:id", paymentHandler.GetTaxRate)
	for _, example := range []struct{ path, expected string }{
		{"/storefront/context?country=NZ", `"code":"RESTORE_NZ"`},
		{"/storefront/context?country=NZ&locale=zh_cn&currency=USD", `"resolved":"zh_cn"`},
		{"/payment/tax-rates/900090", "Restored NZ checkout tax"},
	} {
		t.Run("restored_market_and_tax_http"+example.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, example.path, nil))
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.Contains(t, response.Body.String(), example.expected)
		})
	}
	exchange := service.NewExchangeRateService(repository.NewExchangeRateRepository(gormDatabase), settings)
	converted, err := exchange.ConvertMoneyStrict(money.MustNew(10000, "USD"), "NZD")
	require.NoError(t, err)
	require.Equal(t, int64(12346), converted.AmountMinor())
	_, err = exchange.ConvertMoneyStrict(money.MustNew(10000, "USD"), "JPY")
	require.ErrorIs(t, err, service.ErrExchangeRateMissing)
	decimal, calculatedTax, err := taxService.CalculateTaxMoney(money.MustNew(10000, "NZD"), "NZ", "")
	require.NoError(t, err)
	require.Equal(t, "7.500000000000000", decimal)
	require.Equal(t, int64(750), calculatedTax.AmountMinor())
	reference, err := service.NewTaxRateSourceSnapshotService(repository.NewTaxRateSourceSnapshotRepository(gormDatabase)).GetCurrentSnapshot()
	require.NoError(t, err)
	require.NotNil(t, reference.Snapshot)
	require.Equal(t, uint(900090), reference.Snapshot.ID)
	require.Len(t, reference.Entries, 2)
	if scenario == "legacy_reference_comments_not_collected" {
		require.Nil(t, reference.Snapshot.CountryRateComments)
	} else {
		require.Equal(t, []string{"Preserved original rate note"}, reference.Snapshot.CountryRateComments["GR"]["24"])
	}
	var exactRate, expiry, owner string
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT rate_decimal::text FROM currency_exchange_rates WHERE id=900090`).Scan(&exactRate))
	require.Equal(t, "1.234567891234567", exactRate)
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT expires_at::date::text FROM currency_exchange_rates WHERE id=900091`).Scan(&expiry))
	require.Equal(t, "2020-01-01", expiry)
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT owner_id FROM currency_exchange_sync_leases WHERE lease_key='configuration-preserved-rate-lease'`).Scan(&owner))
	require.Equal(t, "original-owner", owner)
	var enabled bool
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT enabled FROM tax_rate_source_configs WHERE id=1`).Scan(&enabled))
	require.False(t, enabled)
	var originalTotal int64
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT total_amount_minor FROM orders WHERE id=900001`).Scan(&originalTotal))
	require.Equal(t, int64(123), originalTotal)
}

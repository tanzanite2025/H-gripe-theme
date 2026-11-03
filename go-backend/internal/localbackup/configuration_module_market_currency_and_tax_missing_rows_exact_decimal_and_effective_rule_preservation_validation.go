package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/pkg/locales"
)

// Source sync controls, credentials, leases and primary pricing currency stay
// read-only. Provider snapshots never become checkout rules through restore.
func configurationModuleMissingMarketCurrencyAndTaxWritableTableNames() []string {
	return []string{"storefront_markets", "storefront_market_countries", "currency_exchange_rates", "tax_rates", "tax_rate_source_snapshots", "tax_rate_source_snapshot_entries"}
}

var (
	configurationModuleRestoredMarketCodePattern      = regexp.MustCompile(`^[A-Z0-9_-]{1,32}$`)
	configurationModuleRestoredCountryCodePattern     = regexp.MustCompile(`^[A-Z]{2}$`)
	configurationModuleRestoredCurrencyCodePattern    = regexp.MustCompile(`^[A-Z]{3}$`)
	configurationModuleRestoredPositiveDecimalPattern = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?$`)
	configurationModuleRestoredSnapshotDigestPattern  = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

func validateConfigurationModuleMissingMarketCurrencyAndTaxDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingMarketCurrencyAndTaxWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleRestoredMarketCurrencyOrTaxCompleteRow(table, row); err != nil {
				return fmt.Errorf("市场、汇率或税率备份字段不符合恢复规则：%s", table)
			}
		}
	}
	return nil
}

// Decode JSON numbers as their original decimal text, never through float64 or
// a normalizing save hook. Full-row SQL comparison still verifies exact values.
func validateConfigurationModuleRestoredMarketCurrencyOrTaxCompleteRow(table string, row json.RawMessage) error {
	field := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	switch table {
	case "storefront_markets":
		var record struct {
			Code              string   `json:"code"`
			Name              string   `json:"name"`
			DefaultLocale     string   `json:"default_locale"`
			SupportedLocales  []string `json:"supported_locales"`
			DefaultCurrency   string   `json:"default_currency"`
			DisplayCurrencies []string `json:"display_currencies"`
		}
		if err := json.Unmarshal(row, &record); err != nil {
			return err
		}
		if !configurationModuleRestoredMarketCodePattern.MatchString(record.Code) || strings.TrimSpace(record.Name) == "" {
			return errors.New("invalid market identity")
		}
		if !locales.IsSupportedCode(record.DefaultLocale) || !configurationModuleRestoredStringListContainsCanonicalValues(record.SupportedLocales, record.DefaultLocale, locales.IsSupportedCode) {
			return errors.New("invalid market locales")
		}
		if currency.NormalizeCode(record.DefaultCurrency) != record.DefaultCurrency || !configurationModuleRestoredStringListContainsCanonicalValues(record.DisplayCurrencies, record.DefaultCurrency, func(value string) bool {
			return currency.NormalizeCode(value) == value && currency.IsCatalogCode(value)
		}) {
			return errors.New("invalid market currencies")
		}
	case "storefront_market_countries":
		if !configurationModuleRestoredCountryCodePattern.MatchString(field("code")) || field("code") == "ZZ" {
			return errors.New("invalid country")
		}
	case "currency_exchange_rates":
		base, quote, decimal := field("base_currency"), field("quote_currency"), field("rate_decimal")
		if !currency.IsCatalogCode(base) || !currency.IsCatalogCode(quote) || currency.NormalizeCode(base) != base || currency.NormalizeCode(quote) != quote || !configurationModuleRestoredPositiveDecimalPattern.MatchString(decimal) {
			return errors.New("invalid rate identity or decimal")
		}
		restoredExchangeRate := currency.ExchangeRate{RateDecimal: decimal}
		if err := restoredExchangeRate.NormalizeRateDecimal(); err != nil {
			return err
		}
		if value, ok := new(big.Rat).SetString(decimal); base == quote && (!ok || value.Cmp(big.NewRat(1, 1)) != 0) {
			return errors.New("self rate must be one")
		}
		fetched, err := time.Parse(time.RFC3339Nano, field("fetched_at"))
		if err != nil {
			return err
		}
		if expiry := field("expires_at"); expiry != "" {
			expires, err := time.Parse(time.RFC3339Nano, expiry)
			if err != nil || !expires.After(fetched) {
				return errors.New("invalid rate time interval")
			}
		}
	case "tax_rates":
		if strings.TrimSpace(field("name")) == "" || !configurationModuleRestoredCountryCodePattern.MatchString(field("country")) || field("country") == "ZZ" {
			return errors.New("invalid tax rule identity")
		}
		if err := taxrate.ValidateTaxRatePercentage(field("rate_decimal")); err != nil {
			return err
		}
		// Current checkout matches country/state/postcode and applies one rule.
		// Refuse legacy fields that would falsely imply city or compound support.
		if field("compound") == "true" || field("city") != "" {
			return errors.New("unsupported compound or city tax matching")
		}
		for _, column := range []string{"state", "postal_code"} {
			if value := field(column); value != strings.ToUpper(strings.TrimSpace(value)) {
				return errors.New("noncanonical tax location")
			}
		}
	case "tax_rate_source_snapshots":
		if field("provider_code") != taxrate.DefaultTaxRateSourceProviderCode || !configurationModuleRestoredSnapshotDigestPattern.MatchString(field("content_sha256")) || field("version") == "" || field("source_endpoint") == "" {
			return errors.New("invalid reference snapshot identity")
		}
		var record struct {
			Comments taxrate.TaxRateSourceCountryRateComments `json:"country_rate_comments"`
		}
		if err := json.Unmarshal(row, &record); err != nil {
			return err
		}
		for country, rates := range record.Comments {
			if !configurationModuleRestoredCountryCodePattern.MatchString(country) {
				return errors.New("invalid comment country")
			}
			for decimal := range rates {
				if err := taxrate.ValidateTaxRatePercentage(decimal); err != nil {
					return err
				}
			}
		}
	case "tax_rate_source_snapshot_entries":
		if !configurationModuleRestoredCountryCodePattern.MatchString(field("country_code")) || !configurationModuleRestoredCountryCodePattern.MatchString(field("source_country_code")) || strings.TrimSpace(field("country_name")) == "" {
			return errors.New("invalid reference country")
		}
		if code := field("currency"); code != "" && !configurationModuleRestoredCurrencyCodePattern.MatchString(code) {
			return errors.New("invalid reference currency")
		}
		if err := taxrate.ValidateTaxRatePercentage(field("rate_decimal")); err != nil {
			return err
		}
		switch field("rate_type") {
		case "standard", "reduced", "super_reduced", "parking":
			if field("rate_category") != "" {
				return errors.New("unexpected reference category")
			}
		case "product_category":
			if strings.TrimSpace(field("rate_category")) == "" {
				return errors.New("missing reference category")
			}
		default:
			return errors.New("unknown reference rate type")
		}
	}
	return nil
}

func configurationModuleRestoredStringListContainsCanonicalValues(values []string, required string, valid func(string) bool) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if !valid(value) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return required != "" && seen[required]
}

// Validate only affected parents, including existing parents of restored child
// rows. Counts cover the complete final graph, not just backup-selected entries.
func validateConfigurationModuleRestoredMarketCurrencyAndTaxRelationships(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, missing map[string][]json.RawMessage) error {
	if err := validateConfigurationModuleMissingMarketCurrencyAndTaxDomainRows(missing); err != nil {
		return err
	}
	for _, group := range []struct{ parent, child, childParentColumn string }{
		{"storefront_markets", "storefront_market_countries", "market_id"},
		{"tax_rate_source_snapshots", "tax_rate_source_snapshot_entries", "snapshot_id"},
	} {
		ids := []string{}
		for _, row := range missing[group.parent] {
			id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
			ids = append(ids, id)
		}
		for _, row := range missing[group.child] {
			id, _ := readBusinessModuleSnapshotRowColumnText(row, group.childParentColumn)
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			continue
		}
		parents, err := readProductCatalogRows(ctx, tx, group.parent, `r.id::text=ANY($1)`, pq.Array(ids))
		if err != nil {
			return err
		}
		for _, row := range parents {
			if err := validateConfigurationModuleRestoredMarketCurrencyOrTaxCompleteRow(group.parent, row); err != nil {
				return errors.New("恢复子记录所属的市场或税率参考快照无效")
			}
		}
		children, err := readProductCatalogRows(ctx, tx, group.child, `r.`+pq.QuoteIdentifier(group.childParentColumn)+`::text=ANY($1)`, pq.Array(ids))
		if err != nil {
			return err
		}
		for _, row := range children {
			if err := validateConfigurationModuleRestoredMarketCurrencyOrTaxCompleteRow(group.child, row); err != nil {
				return errors.New("恢复后的市场国家或税率参考条目字段无效")
			}
		}
		var invalid bool
		if group.parent == "storefront_markets" {
			err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.storefront_markets m JOIN public.storefront_market_countries c ON c.market_id=m.id WHERE m.id::text=ANY($1) AND c.deleted_at IS NULL AND m.deleted_at IS NOT NULL)`, pq.Array(ids)).Scan(&invalid)
		} else {
			err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.tax_rate_source_snapshots s LEFT JOIN LATERAL (SELECT count(*) rates,count(DISTINCT country_code) countries FROM public.tax_rate_source_snapshot_entries e WHERE e.snapshot_id=s.id) a ON true WHERE s.id::text=ANY($1) AND (s.rate_count<>a.rates OR s.country_count<>a.countries))`, pq.Array(ids)).Scan(&invalid)
		}
		if err != nil {
			return err
		}
		if invalid {
			return errors.New("市场国家关联或税率参考快照的完整国家/税率数量不一致")
		}
	}
	// Tax scopes and direct/reverse exchange pairs lack a cross-row SQL guard.
	// Do not let an old missing row shadow a target-only current decision.
	for _, table := range []string{"tax_rates", "currency_exchange_rates"} {
		if len(missing[table]) == 0 {
			continue
		}
		packageIDs := []string{}
		for _, row := range snapshot.Tables[table].Rows {
			id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
			packageIDs = append(packageIDs, id)
		}
		for _, row := range missing[table] {
			field := func(column string) string { v, _ := readBusinessModuleSnapshotRowColumnText(row, column); return v }
			if field("deleted_at") != "" || table == "tax_rates" && field("enabled") != "true" {
				continue
			}
			var conflict bool
			var err error
			if table == "tax_rates" {
				err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.tax_rates t WHERE t.deleted_at IS NULL AND t.enabled AND NOT(t.id::text=ANY($1)) AND t.country=$2 AND COALESCE(t.state,'')=$3 AND COALESCE(t.postal_code,'')=$4)`, pq.Array(packageIDs), field("country"), field("state"), field("postal_code")).Scan(&conflict)
			} else {
				err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM public.currency_exchange_rates r WHERE r.deleted_at IS NULL AND NOT(r.id::text=ANY($1)) AND ((r.base_currency=$2 AND r.quote_currency=$3) OR (r.base_currency=$3 AND r.quote_currency=$2)))`, pq.Array(packageIDs), field("base_currency"), field("quote_currency")).Scan(&conflict)
			}
			if err != nil {
				return err
			}
			if conflict {
				return errors.New("补回旧税率或汇率会改变当前新增规则的匹配结果，请先核对该范围")
			}
		}
	}
	return nil
}

package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/shipping"
)

// Customer prices, published route collections, provider credentials and tracking
// execution controls have separate contracts. Exporting them grants no writes.
func configurationModuleMissingShippingReferencePackagingAndCustomsWritableTableNames() []string {
	return []string{"carriers", "shipping_zones", "customs_classification_profiles", "shipping_packaging_rules", "shipping_packaging_rule_applies"}
}

// These two columns have no database FKs. Capture the original complete product
// and SKU fingerprints as read-only dependencies; an existing numeric ID alone
// cannot prove that a restored packaging binding still belongs to that product.
func configurationModulePackagingReadOnlyBusinessDependencyReferences() []ModuleRecoveryDependencyReference {
	return []ModuleRecoveryDependencyReference{
		{SourceTable: "shipping_packaging_rule_applies", SourceColumns: []string{"product_id"}, TargetTable: "products", TargetColumns: []string{"id"}},
		{SourceTable: "shipping_packaging_rule_applies", SourceColumns: []string{"variant_id"}, TargetTable: "product_variants", TargetColumns: []string{"id"}},
	}
}

func validateConfigurationModuleMissingShippingReferencePackagingAndCustomsDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingShippingReferencePackagingAndCustomsWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleRestoredShippingReferencePackagingOrCustomsCompleteRow(table, row); err != nil {
				return fmt.Errorf("物流基础资料、包装或报关备份字段无效：%s", table)
			}
		}
	}
	return nil
}

var configurationModuleRestoredCustomsSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func validateConfigurationModuleRestoredShippingReferencePackagingOrCustomsCompleteRow(table string, row json.RawMessage) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(row, &object); err != nil || object == nil {
		return errors.New("expected a complete configuration object")
	}
	field := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	positiveID := func(column string, optional bool) bool {
		value := field(column)
		if optional && (object[column] == nil || string(object[column]) == "null") {
			return true
		}
		id, err := strconv.ParseInt(value, 10, 64)
		return err == nil && id > 0
	}
	switch table {
	case "carriers":
		if strings.TrimSpace(field("name")) == "" || strings.TrimSpace(field("code")) == "" {
			return errors.New("carrier name and code are required")
		}
	case "shipping_zones":
		if strings.TrimSpace(field("name")) == "" {
			return errors.New("shipping zone name is required")
		}
		for _, column := range []string{"countries", "states", "postal_codes"} {
			if err := validateConfigurationModuleRestoredShippingStringArray(field(column), column == "countries"); err != nil {
				return err
			}
		}
	case "shipping_packaging_rules":
		if strings.TrimSpace(field("rule_name")) == "" {
			return errors.New("packaging rule name is required")
		}
		// Keep NUMERIC values exact. Never rewrite original dimensions or round
		// a malformed scale while validating recovery data.
		for _, column := range []string{"box_weight", "box_length", "box_width", "box_height", "max_weight"} {
			value := field(column)
			scale := int64(100)
			if column == "box_weight" || column == "max_weight" {
				scale = 1000
			}
			decimal, ok := new(big.Rat).SetString(value)
			if !configurationModuleRestoredPositiveDecimalPattern.MatchString(value) || !ok || decimal.Sign() < 0 || !new(big.Rat).Mul(decimal, new(big.Rat).SetInt64(scale)).IsInt() {
				return errors.New("invalid exact packaging dimension or weight")
			}
		}
	case "shipping_packaging_rule_applies":
		if !positiveID("rule_id", false) || !positiveID("product_id", false) || !positiveID("variant_id", true) {
			return errors.New("invalid packaging rule, product or variant identity")
		}
	case "customs_classification_profiles":
		digits := func(value string, length int) bool {
			return len(value) == length && strings.IndexFunc(value, func(r rune) bool { return r < '0' || r > '9' }) == -1
		}
		if strings.TrimSpace(field("name")) == "" || !configurationModuleRestoredCustomsSlugPattern.MatchString(field("slug")) || !digits(field("hs_code"), 6) || (field("cn_code") != "" && !digits(field("cn_code"), 8)) || (field("country_of_origin") != "" && !configurationModuleRestoredCountryCodePattern.MatchString(field("country_of_origin"))) {
			return errors.New("invalid customs name, slug, classification code or origin")
		}
		if !product.IsCustomsClassificationStatus(field("status")) || !product.IsCustomsTradeRemedyRiskLevel(field("trade_remedy_risk_level")) || len(strings.TrimSpace(field("customs_description"))) > 255 || len(strings.TrimSpace(field("trade_remedy_declaration_advice"))) > 4000 {
			return errors.New("invalid customs status, description or trade remedy risk")
		}
		var tags []string
		if raw := object["trade_remedy_risk_tags_json"]; len(raw) == 0 || string(raw) == "null" || json.Unmarshal(raw, &tags) != nil {
			return errors.New("expected customs trade remedy tag array")
		}
		seen := map[string]bool{}
		for _, tag := range tags {
			if !product.IsCustomsTradeRemedyRiskTag(tag) || seen[tag] {
				return errors.New("invalid or repeated customs trade remedy tag")
			}
			seen[tag] = true
		}
	}
	return nil
}

func validateConfigurationModuleRestoredShippingStringArray(value string, countryCodes bool) error {
	var values []string
	if strings.TrimSpace(value) == "null" || json.Unmarshal([]byte(value), &values) != nil || values == nil {
		return errors.New("expected shipping string array")
	}
	seen := map[string]bool{}
	for _, entry := range values {
		if entry == "" || entry != strings.TrimSpace(entry) || seen[entry] {
			return errors.New("empty or repeated shipping scope entry")
		}
		seen[entry] = true
		if countryCodes {
			if entry != strings.ToUpper(entry) {
				return errors.New("noncanonical shipping country")
			}
			encoded, _ := json.Marshal([]string{entry})
			if err := shipping.ValidateShippingCountryCodes(string(encoded)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate the affected final graph, including an existing parent of a restored
// binding and target-only bindings of a restored rule. Product/SKU rows remain
// read-only. Normal INSERT constraints also enforce unique default/SKU bindings.
func validateConfigurationModuleRestoredShippingPackagingProductVariantRelationships(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage) error {
	if err := validateConfigurationModuleMissingShippingReferencePackagingAndCustomsDomainRows(missing); err != nil {
		return err
	}
	ruleIDs := []string{}
	for _, pair := range []struct{ table, column string }{{"shipping_packaging_rules", "id"}, {"shipping_packaging_rule_applies", "rule_id"}} {
		for _, row := range missing[pair.table] {
			id, _ := readBusinessModuleSnapshotRowColumnText(row, pair.column)
			ruleIDs = append(ruleIDs, id)
		}
	}
	if len(ruleIDs) == 0 {
		return nil
	}
	for _, pair := range []struct{ table, predicate string }{{"shipping_packaging_rules", `r.id::text=ANY($1)`}, {"shipping_packaging_rule_applies", `r.rule_id::text=ANY($1)`}} {
		rows, err := readProductCatalogRows(ctx, tx, pair.table, pair.predicate, pq.Array(ruleIDs))
		if err != nil {
			return err
		}
		for _, row := range rows {
			if err := validateConfigurationModuleRestoredShippingReferencePackagingOrCustomsCompleteRow(pair.table, row); err != nil {
				return errors.New("恢复后包装规则的现有父记录或关联记录无效")
			}
		}
	}
	var invalid bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS (
SELECT 1 FROM public.shipping_packaging_rule_applies a
LEFT JOIN public.products p ON p.id=a.product_id
LEFT JOIN public.product_variants v ON v.id=a.variant_id
WHERE a.rule_id::text=ANY($1) AND (p.id IS NULL OR p.deleted_at IS NOT NULL OR
 (a.variant_id IS NOT NULL AND (v.id IS NULL OR v.deleted_at IS NOT NULL OR v.product_id<>a.product_id))))`, pq.Array(ruleIDs)).Scan(&invalid)
	if err != nil {
		return err
	}
	if invalid {
		return errors.New("包装绑定的商品或 SKU 缺失、已删除或所属商品不符；请先恢复原商品")
	}
	return nil
}

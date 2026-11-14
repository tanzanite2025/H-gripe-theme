package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/shipping"
)

// Provider identities and collections have their own shipping-reference contract.
// Ordering follows migration 415, never the removed carrier-service template_id.
// Credentials and tracking execution controls remain outside both contracts.
func configurationModuleMissingCustomerShippingTemplatesRulesAndRouteMembershipWritableTableNames() []string {
	return []string{"shipping_route_groups", "shipping_templates", "shipping_rules", "shipping_route_group_carrier_services"}
}

func validateConfigurationModuleMissingCustomerShippingTemplatesRulesAndRouteMembershipDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingCustomerShippingTemplatesRulesAndRouteMembershipWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleRestoredCustomerShippingTemplateRuleOrRouteCompleteRow(table, row); err != nil {
				return fmt.Errorf("运费模板、计费规则或线路聚合字段无效（%s）：%w", table, err)
			}
		}
	}
	return nil
}

func validateConfigurationModuleRestoredCustomerShippingTemplateRuleOrRouteCompleteRow(table string, row json.RawMessage) error {
	var object map[string]json.RawMessage
	if json.Unmarshal(row, &object) != nil || object == nil {
		return errors.New("expected complete shipping configuration object")
	}
	field := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	positiveID := func(column string, optional bool) bool {
		if optional && string(object[column]) == "null" {
			return true
		}
		value, err := strconv.ParseInt(field(column), 10, 64)
		return err == nil && value > 0
	}
	nonnegativeIntegers := func(columns string) bool {
		for _, column := range strings.Fields(columns) {
			value, err := strconv.ParseInt(field(column), 10, 64)
			if err != nil || value < 0 {
				return false
			}
		}
		return true
	}
	if table == "shipping_route_group_carrier_services" {
		if !positiveID("route_group_id", false) || !positiveID("carrier_service_id", false) {
			return errors.New("invalid route aggregate or carrier service identity")
		}
		return nil
	}
	if !positiveID("id", false) {
		return errors.New("invalid shipping configuration identity")
	}
	if table == "shipping_rules" {
		if !positiveID("template_id", false) || !currency.IsCatalogCode(field("currency")) || field("currency") != currency.NormalizeCode(field("currency")) || !nonnegativeIntegers("min_value_minor max_value_minor fee_minor additional_minor") {
			return errors.New("invalid rule template, currency or exact monetary amount")
		}
		// Read dimension thresholds directly: ShippingRule's JSON contract
		// intentionally omits these columns. Do not round through float64.
		for _, column := range []string{"min_value", "max_value"} {
			value, ok := new(big.Rat).SetString(field(column))
			if !configurationModuleRestoredPositiveDecimalPattern.MatchString(field(column)) || !ok || value.Sign() < 0 || !new(big.Rat).Mul(value, big.NewRat(1000, 1)).IsInt() {
				return errors.New("invalid exact rule dimensional threshold")
			}
		}
		minimum, _ := new(big.Rat).SetString(field("min_value"))
		maximum, _ := new(big.Rat).SetString(field("max_value"))
		minimumMinor, _ := strconv.ParseInt(field("min_value_minor"), 10, 64)
		maximumMinor, _ := strconv.ParseInt(field("max_value_minor"), 10, 64)
		if (maximum.Sign() > 0 && maximum.Cmp(minimum) < 0) || (maximumMinor > 0 && maximumMinor < minimumMinor) {
			return errors.New("rule maximum cannot precede minimum")
		}
		return nil
	}
	if strings.TrimSpace(field("name")) == "" {
		return errors.New("shipping configuration name is required")
	}
	if table == "shipping_route_groups" {
		return validateConfigurationModuleRestoredCustomerShippingCountryScope(field("countries"))
	}
	if !positiveID("route_group_id", true) || !currency.IsCatalogCode(field("currency")) || field("currency") != currency.NormalizeCode(field("currency")) || !nonnegativeIntegers("free_threshold_minor default_fee_minor customer_first_weight_grams customer_additional_weight_grams customer_min_charge_weight_grams customer_volumetric_divisor customer_remote_surcharge_minor customer_eta_min_days customer_eta_max_days") {
		return errors.New("invalid template route, currency or billing amounts")
	}
	switch field("type") {
	case "weight", "quantity", "price", "amount", shipping.ShippingTemplateTypeSystemFreeShipping:
	default:
		return errors.New("unsupported customer shipping template type")
	}
	if field("template_kind") != shipping.ShippingTemplateKindCarrier && field("template_kind") != shipping.ShippingTemplateKindSystemFreeShipping {
		return errors.New("invalid template kind")
	}
	if field("customer_billing_mode") == "" || field("customer_volumetric_divisor") == "0" || !configurationModuleRestoredPositiveDecimalPattern.MatchString(field("customer_fuel_surcharge_percent_decimal")) {
		return errors.New("billing defaults must already be explicit in the original row")
	}
	if err := validateConfigurationModuleRestoredShippingStringArray(field("customer_remote_postal_codes"), false); err != nil {
		return err
	}
	if err := validateConfigurationModuleRestoredCustomerShippingCountryScope(field("free_shipping_countries")); err != nil {
		return err
	}
	// PostgreSQL represents NUMERIC as a JSON number whereas the domain API
	// uses a decimal string. Convert only the validation copy, never the row.
	object["customer_fuel_surcharge_percent_decimal"], _ = json.Marshal(field("customer_fuel_surcharge_percent_decimal"))
	encoded, err := json.Marshal(object)
	if err != nil {
		return err
	}
	var template shipping.ShippingTemplate
	if err := json.Unmarshal(encoded, &template); err != nil {
		return err
	}
	isSystem := template.TemplateKind == shipping.ShippingTemplateKindSystemFreeShipping
	if isSystem != (template.Type == shipping.ShippingTemplateTypeSystemFreeShipping) || (isSystem && (!template.IsSystemManaged || !template.FreeShipping || template.FreeThresholdMinor != 0 || template.DefaultFeeMinor != 0 || template.RouteGroupID != nil)) {
		return errors.New("inconsistent system free-shipping identity or customer pricing")
	}
	return template.ValidateCustomerShippingBillingConfiguration()
}

func validateConfigurationModuleRestoredCustomerShippingCountryScope(value string) error {
	if err := validateConfigurationModuleRestoredShippingStringArray(value, false); err != nil {
		return err
	}
	var codes []string
	_ = json.Unmarshal([]byte(value), &codes)
	for _, code := range codes {
		if code != "*" && code != "ALL" && code != "GLOBAL" && code != "WORLDWIDE" && !configurationModuleRestoredCountryCodePattern.MatchString(code) {
			return errors.New("invalid shipping country scope")
		}
	}
	return nil
}

// Validate the final affected graph, including retained parents and target-only
// children. Checking just inserted rows would miss a partial rule/member restore
// against an invalid existing template or an extra incompatible child.
func validateConfigurationModuleRestoredCustomerShippingBillingAndPublishedRouteRelationships(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage) error {
	templateIDs, groupIDs := []string{}, []string{}
	for _, table := range configurationModuleMissingCustomerShippingTemplatesRulesAndRouteMembershipWritableTableNames() {
		for _, row := range missing[table] {
			column := "id"
			switch table {
			case "shipping_rules":
				column = "template_id"
			case "shipping_route_group_carrier_services":
				column = "route_group_id"
			}
			id, _ := readBusinessModuleSnapshotRowColumnText(row, column)
			if table == "shipping_templates" || table == "shipping_rules" {
				templateIDs = append(templateIDs, id)
			} else {
				groupIDs = append(groupIDs, id)
			}
		}
	}
	if len(templateIDs)+len(groupIDs) == 0 {
		return nil
	}
	templates, err := readProductCatalogRows(ctx, tx, "shipping_templates", `r.id::text=ANY($1) OR r.route_group_id::text=ANY($2)`, pq.Array(templateIDs), pq.Array(groupIDs))
	if err != nil {
		return err
	}
	for _, row := range templates {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		templateIDs = append(templateIDs, id)
		groupID, _ := readBusinessModuleSnapshotRowColumnText(row, "route_group_id")
		if groupID != "" {
			groupIDs = append(groupIDs, groupID)
		}
		if err := validateConfigurationModuleRestoredCustomerShippingTemplateRuleOrRouteCompleteRow("shipping_templates", row); err != nil {
			return fmt.Errorf("恢复后运费模板（含现有父记录）无效：%w", err)
		}
	}
	for _, entry := range []struct {
		table, predicate string
		ids              []string
	}{
		{"shipping_rules", `r.template_id::text=ANY($1)`, templateIDs},
		{"shipping_route_groups", `r.id::text=ANY($1)`, groupIDs},
	} {
		rows, err := readProductCatalogRows(ctx, tx, entry.table, entry.predicate, pq.Array(entry.ids))
		if err != nil {
			return err
		}
		for _, row := range rows {
			if err := validateConfigurationModuleRestoredCustomerShippingTemplateRuleOrRouteCompleteRow(entry.table, row); err != nil {
				return fmt.Errorf("恢复后的计价/线路关系无效（%s）：%w", entry.table, err)
			}
		}
	}
	var invalid bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (
SELECT 1 FROM public.shipping_rules r JOIN public.shipping_templates t ON t.id=r.template_id
WHERE t.id::text=ANY($1) AND (t.template_kind='system_free_shipping' OR r.currency<>t.currency OR
 (t.type='weight' AND r.additional_minor>0 AND (t.customer_first_weight_grams<=0 OR t.customer_additional_weight_grams<=0)))
UNION ALL
SELECT 1 FROM public.shipping_templates t LEFT JOIN public.shipping_route_groups g ON g.id=t.route_group_id
WHERE t.id::text=ANY($1) AND t.deleted_at IS NULL AND t.route_group_id IS NOT NULL AND (g.id IS NULL OR g.deleted_at IS NOT NULL)
UNION ALL
SELECT 1 FROM public.shipping_route_groups g WHERE g.id::text=ANY($2) AND g.deleted_at IS NULL AND NOT EXISTS
 (SELECT 1 FROM public.shipping_route_group_carrier_services m WHERE m.route_group_id=g.id)
UNION ALL
SELECT 1 FROM public.shipping_route_group_carrier_services m
JOIN public.shipping_route_groups g ON g.id=m.route_group_id
LEFT JOIN public.shipping_carrier_services s ON s.id=m.carrier_service_id
LEFT JOIN public.carriers c ON c.id=s.carrier_id
LEFT JOIN public.shipping_fpx_channels f ON f.id=s.fpx_channel_id
LEFT JOIN public.shipping_yanwen_published_channels y ON y.id=s.yanwen_published_channel_id
WHERE g.id::text=ANY($2) AND g.deleted_at IS NULL AND
 (m.route_group_id<=0 OR m.carrier_service_id<=0 OR s.id IS NULL OR s.deleted_at IS NOT NULL OR c.id IS NULL OR c.deleted_at IS NOT NULL OR
 (UPPER(TRIM(s.service_code)) LIKE 'YANWEN:%' AND UPPER(TRIM(c.code))<>'YANWEN') OR
 (s.fpx_channel_id IS NOT NULL AND s.yanwen_published_channel_id IS NOT NULL) OR
 (UPPER(TRIM(c.code)) IN ('4PX','FPX') AND (s.fpx_channel_id IS NULL OR f.id IS NULL OR f.deleted_at IS NOT NULL OR f.environment<>'production' OR NOT f.enabled)) OR
 (UPPER(TRIM(c.code))='YANWEN' AND (s.yanwen_published_channel_id IS NULL OR y.id IS NULL OR y.deleted_at IS NOT NULL OR y.environment<>'production' OR NOT y.enabled)) OR
 (UPPER(TRIM(c.code)) NOT IN ('4PX','FPX','YANWEN') AND (s.fpx_channel_id IS NOT NULL OR s.yanwen_published_channel_id IS NOT NULL))))`, pq.Array(templateIDs), pq.Array(groupIDs)).Scan(&invalid)
	if err != nil {
		return err
	}
	if invalid {
		return errors.New("运费规则币种、系统免邮、续重单位、现有模板或已发布线路依赖不兼容；不会改写渠道或重新报价")
	}
	return nil
}

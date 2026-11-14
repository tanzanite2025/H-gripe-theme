package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/shipping"
)

// Local reference data is restored before published channels, carrier services
// and route memberships. Credentials, sync commands and tracking controls are
// deliberately outside this contract.
func configurationModuleMissingShippingCarrierServicesPublishedChannelsAndYanwenReferenceCatalogWritableTableNames() []string {
	return []string{"shipping_yanwen_product_catalog_entries", "shipping_yanwen_country_catalog_entries", "shipping_yanwen_warehouse_catalog_entries", "shipping_fpx_channels", "shipping_yanwen_published_channels", "shipping_carrier_services"}
}

func validateConfigurationModuleMissingShippingCarrierServicesPublishedChannelsAndYanwenReferenceCatalogDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingShippingCarrierServicesPublishedChannelsAndYanwenReferenceCatalogWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleRestoredShippingCarrierServicePublishedChannelOrYanwenCatalogCompleteRow(table, row); err != nil {
				return fmt.Errorf("承运服务、发布渠道或燕文参考目录字段无效（%s）：%w", table, err)
			}
		}
	}
	return nil
}

func validateConfigurationModuleRestoredShippingCarrierServicePublishedChannelOrYanwenCatalogCompleteRow(table string, row json.RawMessage) error {
	var object map[string]json.RawMessage
	if json.Unmarshal(row, &object) != nil || object == nil {
		return errors.New("expected complete shipping reference object")
	}
	field := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	positiveID := func(column string, optional bool) bool {
		if optional && string(object[column]) == "null" {
			return true
		}
		id, err := strconv.ParseInt(field(column), 10, 64)
		return err == nil && id > 0
	}
	if !positiveID("id", false) {
		return errors.New("invalid shipping reference identity")
	}
	if table != "shipping_carrier_services" {
		environment := field("environment")
		validEnvironment := environment == "production"
		if table == "shipping_fpx_channels" {
			validEnvironment = validEnvironment || environment == "test"
		} else {
			validEnvironment = validEnvironment || environment == "fat"
		}
		if !validEnvironment {
			return errors.New("original gateway environment must be explicit and canonical")
		}
	}
	if table == "shipping_carrier_services" || table == "shipping_fpx_channels" || table == "shipping_yanwen_published_channels" {
		if err := validateConfigurationModuleRestoredCustomerShippingCountryScope(field("countries")); err != nil {
			return err
		}
	}
	if table == "shipping_carrier_services" {
		if !positiveID("carrier_id", false) || !positiveID("fpx_channel_id", true) || !positiveID("yanwen_published_channel_id", true) || (field("fpx_channel_id") != "" && field("yanwen_published_channel_id") != "") || strings.TrimSpace(field("service_code")) == "" || strings.TrimSpace(field("service_name")) == "" || !currency.IsCatalogCode(field("currency")) || field("currency") != currency.NormalizeCode(field("currency")) {
			return errors.New("invalid carrier, stable channel identity, service name or provider currency")
		}
		switch field("billing_mode") {
		case "actual_weight", "volumetric_weight", "greater_of_actual_and_volumetric":
		default:
			return errors.New("invalid provider billing mode")
		}
		values := map[string]int64{}
		for _, column := range strings.Fields("first_weight_grams additional_weight_grams min_charge_weight_grams volumetric_divisor remote_surcharge_minor eta_min_days eta_max_days") {
			value, err := strconv.ParseInt(field(column), 10, 64)
			if err != nil || value < 0 {
				return errors.New("invalid exact provider billing amount or weight")
			}
			values[column] = value
		}
		if values["volumetric_divisor"] == 0 || (values["eta_min_days"] > 0 && values["eta_max_days"] > 0 && values["eta_max_days"] < values["eta_min_days"]) {
			return errors.New("invalid provider divisor or delivery time range")
		}
		if err := validateConfigurationModuleRestoredShippingStringArray(field("remote_postal_codes"), false); err != nil {
			return err
		}
		if !configurationModuleRestoredPositiveDecimalPattern.MatchString(field("fuel_surcharge_percent_decimal")) {
			return errors.New("provider fuel surcharge must be an original exact decimal")
		}
		_, err := (shipping.CarrierService{FuelSurchargePercentDecimal: field("fuel_surcharge_percent_decimal")}).FuelSurchargeRate()
		return err
	}
	var domain interface{ Validate() error }
	var identityColumns []string
	switch table {
	case "shipping_fpx_channels":
		domain, identityColumns = &shipping.FpxChannel{}, []string{"service_code", "display_name"}
	case "shipping_yanwen_published_channels":
		domain, identityColumns = &shipping.YanwenPublishedChannel{}, []string{"product_code", "display_name"}
	case "shipping_yanwen_product_catalog_entries":
		domain, identityColumns = &shipping.YanwenProductCatalogEntry{}, []string{"product_id"}
	case "shipping_yanwen_country_catalog_entries":
		domain, identityColumns = &shipping.YanwenCountryCatalogEntry{}, []string{"country_id", "country_code"}
		if field("country_code") != strings.ToUpper(field("country_code")) {
			return errors.New("country reference code must already be uppercase")
		}
	case "shipping_yanwen_warehouse_catalog_entries":
		domain, identityColumns = &shipping.YanwenWarehouseCatalogEntry{}, []string{"warehouse_code", "name"}
	default:
		return errors.New("unregistered shipping reference table")
	}
	for _, column := range identityColumns {
		if field(column) == "" || field(column) != strings.TrimSpace(field(column)) {
			return errors.New("original shipping reference identity cannot be empty or normalized during recovery")
		}
	}
	// Validators may normalize their receiver. They receive a separate copy;
	// the original full row and its dates/flags remain the INSERT payload.
	if err := json.Unmarshal(row, domain); err != nil {
		return err
	}
	return domain.Validate()
}

// Expand by natural product/environment references as well as stable FKs.
// This checks retained services and target-only memberships affected by a
// partial catalog/channel restore without granting those rows write permission.
func validateConfigurationModuleRestoredShippingServicesPublishedChannelsAndOfficialProductReferences(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage) error {
	serviceIDs, fpxIDs, yanwenIDs, productKeys := []string{}, []string{}, []string{}, []string{}
	for _, table := range configurationModuleMissingShippingCarrierServicesPublishedChannelsAndYanwenReferenceCatalogWritableTableNames() {
		for _, row := range missing[table] {
			id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
			switch table {
			case "shipping_carrier_services":
				serviceIDs = append(serviceIDs, id)
			case "shipping_fpx_channels":
				fpxIDs = append(fpxIDs, id)
			case "shipping_yanwen_published_channels":
				yanwenIDs = append(yanwenIDs, id)
			case "shipping_yanwen_product_catalog_entries":
				environment, _ := readBusinessModuleSnapshotRowColumnText(row, "environment")
				code, _ := readBusinessModuleSnapshotRowColumnText(row, "product_id")
				productKeys = append(productKeys, environment+":"+strings.ToLower(strings.TrimSpace(code)))
			}
		}
	}
	if len(serviceIDs)+len(fpxIDs)+len(yanwenIDs)+len(productKeys) == 0 {
		return nil
	}
	channels, err := readProductCatalogRows(ctx, tx, "shipping_yanwen_published_channels", `r.id::text=ANY($1) OR (r.environment||':'||LOWER(TRIM(r.product_code)))=ANY($2)`, pq.Array(yanwenIDs), pq.Array(productKeys))
	if err != nil {
		return err
	}
	for _, row := range channels {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		yanwenIDs = append(yanwenIDs, id)
		if err := validateConfigurationModuleRestoredShippingCarrierServicePublishedChannelOrYanwenCatalogCompleteRow("shipping_yanwen_published_channels", row); err != nil {
			return fmt.Errorf("受影响的现有燕文发布渠道无效：%w", err)
		}
	}
	services, err := readProductCatalogRows(ctx, tx, "shipping_carrier_services", `r.id::text=ANY($1) OR r.fpx_channel_id::text=ANY($2) OR r.yanwen_published_channel_id::text=ANY($3)`, pq.Array(serviceIDs), pq.Array(fpxIDs), pq.Array(yanwenIDs))
	if err != nil {
		return err
	}
	for _, row := range services {
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		serviceIDs = append(serviceIDs, id)
		fpxID, _ := readBusinessModuleSnapshotRowColumnText(row, "fpx_channel_id")
		yanwenID, _ := readBusinessModuleSnapshotRowColumnText(row, "yanwen_published_channel_id")
		if fpxID != "" {
			fpxIDs = append(fpxIDs, fpxID)
		}
		if yanwenID != "" {
			yanwenIDs = append(yanwenIDs, yanwenID)
		}
		if err := validateConfigurationModuleRestoredShippingCarrierServicePublishedChannelOrYanwenCatalogCompleteRow("shipping_carrier_services", row); err != nil {
			return fmt.Errorf("受影响的现有承运服务无效：%w", err)
		}
	}
	for _, selection := range []struct {
		table string
		ids   []string
	}{{"shipping_fpx_channels", fpxIDs}, {"shipping_yanwen_published_channels", yanwenIDs}} {
		rows, err := readProductCatalogRows(ctx, tx, selection.table, `r.id::text=ANY($1)`, pq.Array(selection.ids))
		if err != nil {
			return err
		}
		for _, row := range rows {
			if err := validateConfigurationModuleRestoredShippingCarrierServicePublishedChannelOrYanwenCatalogCompleteRow(selection.table, row); err != nil {
				return fmt.Errorf("受影响的现有发布渠道字段无效：%w", err)
			}
			if selection.table == "shipping_yanwen_published_channels" {
				environment, _ := readBusinessModuleSnapshotRowColumnText(row, "environment")
				code, _ := readBusinessModuleSnapshotRowColumnText(row, "product_code")
				productKeys = append(productKeys, environment+":"+strings.ToLower(strings.TrimSpace(code)))
			}
		}
	}
	products, err := readProductCatalogRows(ctx, tx, "shipping_yanwen_product_catalog_entries", `(r.environment||':'||LOWER(TRIM(r.product_id)))=ANY($1)`, pq.Array(productKeys))
	if err != nil {
		return err
	}
	for _, row := range products {
		if err := validateConfigurationModuleRestoredShippingCarrierServicePublishedChannelOrYanwenCatalogCompleteRow("shipping_yanwen_product_catalog_entries", row); err != nil {
			return fmt.Errorf("受影响的现有燕文产品参考无效：%w", err)
		}
	}
	var invalid bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (
 SELECT 1 FROM public.shipping_yanwen_published_channels y WHERE y.id::text=ANY($1) AND y.deleted_at IS NULL AND
 (SELECT count(*) FROM public.shipping_yanwen_product_catalog_entries p WHERE p.environment=y.environment AND LOWER(TRIM(p.product_id))=LOWER(TRIM(y.product_code)))<>1
 UNION ALL
 SELECT 1 FROM public.shipping_carrier_services s LEFT JOIN public.carriers c ON c.id=s.carrier_id
 LEFT JOIN public.shipping_fpx_channels f ON f.id=s.fpx_channel_id
 LEFT JOIN public.shipping_yanwen_published_channels y ON y.id=s.yanwen_published_channel_id
 WHERE s.id::text=ANY($2) AND s.deleted_at IS NULL AND (c.id IS NULL OR c.deleted_at IS NOT NULL OR
 (UPPER(TRIM(s.service_code)) LIKE 'YANWEN:%' AND UPPER(TRIM(c.code))<>'YANWEN') OR
 (UPPER(TRIM(c.code)) IN ('4PX','FPX') AND (f.id IS NULL OR f.deleted_at IS NOT NULL OR f.environment<>'production' OR NOT f.enabled OR s.yanwen_published_channel_id IS NOT NULL)) OR
 (UPPER(TRIM(c.code))='YANWEN' AND (y.id IS NULL OR y.deleted_at IS NOT NULL OR y.environment<>'production' OR NOT y.enabled OR s.fpx_channel_id IS NOT NULL OR
 (SELECT count(*) FROM public.shipping_yanwen_product_catalog_entries p WHERE p.environment=y.environment AND LOWER(TRIM(p.product_id))=LOWER(TRIM(y.product_code)))<>1)) OR
 (UPPER(TRIM(c.code)) NOT IN ('4PX','FPX','YANWEN') AND (s.fpx_channel_id IS NOT NULL OR s.yanwen_published_channel_id IS NOT NULL))))`, pq.Array(yanwenIDs), pq.Array(serviceIDs)).Scan(&invalid)
	if err != nil {
		return err
	}
	if invalid {
		return errors.New("承运服务、生产发布渠道或同环境燕文产品参考不兼容；恢复不会自动改绑或联网同步")
	}
	// The established billing validator owns route/template relationships.
	// These IDs expand its read-only affected graph, never the writer payload.
	groups, err := readProductCatalogRows(ctx, tx, "shipping_route_groups", `EXISTS (SELECT 1 FROM public.shipping_route_group_carrier_services m WHERE m.route_group_id=r.id AND m.carrier_service_id::text=ANY($1))`, pq.Array(serviceIDs))
	if err != nil {
		return err
	}
	if len(groups) > 0 {
		return validateConfigurationModuleRestoredCustomerShippingBillingAndPublishedRouteRelationships(ctx, tx, map[string][]json.RawMessage{"shipping_route_groups": groups})
	}
	return nil
}

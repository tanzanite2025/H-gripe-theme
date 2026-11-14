package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Preserve original credentials and switches in the authenticated configuration
// package. This writer does not register shipments, poll providers or collect tax.
func configurationModuleMissingTrackingProviderMappingAndTaxCollectionControlWritableTableNames() []string {
	return []string{"shipping_tracking_providers", "shipping_tracking_carrier_mappings", "tax_rate_source_configs"}
}

func validateConfigurationModuleMissingTrackingProviderMappingAndTaxCollectionControlDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range configurationModuleMissingTrackingProviderMappingAndTaxCollectionControlWritableTableNames() {
		for _, row := range missing[table] {
			if err := validateConfigurationModuleRestoredTrackingProviderMappingOrTaxCollectionControlCompleteRow(table, row); err != nil {
				// Never include a row, endpoint, credential or original sync error.
				return fmt.Errorf("追踪服务、承运映射或税率采集配置字段无效（%s）：%w", table, err)
			}
		}
	}
	return nil
}

func validateConfigurationModuleRestoredTrackingProviderMappingOrTaxCollectionControlCompleteRow(table string, row json.RawMessage) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(row, &fields) != nil || fields == nil {
		return errors.New("original complete row is required")
	}
	readOriginalText := func(column string) string {
		value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
		return value
	}
	readOriginalInteger := func(column string, minimum int64) (int64, error) {
		value, err := strconv.ParseInt(string(fields[column]), 10, 64)
		if err != nil || value < minimum {
			return 0, errors.New("original integer field is invalid: " + column)
		}
		return value, nil
	}
	if _, err := readOriginalInteger("id", 1); err != nil {
		return err
	}
	if string(fields["enabled"]) != "true" && string(fields["enabled"]) != "false" {
		return errors.New("original enabled state is required")
	}
	switch table {
	case "tax_rate_source_configs":
		interval, err := readOriginalInteger("refresh_interval_hours", 24)
		if readOriginalText("id") != "1" || readOriginalText("provider_code") != "vatcomply" || err != nil || interval > 8760 {
			return errors.New("tax collection singleton, provider or interval is invalid")
		}
		for _, column := range []string{"last_checked_at", "last_successful_sync_at"} {
			if original := readOriginalText(column); original != "" {
				if _, err := time.Parse(time.RFC3339Nano, original); err != nil {
					return errors.New("tax collection original timestamp is invalid")
				}
			}
		}
	case "shipping_tracking_providers":
		if strings.TrimSpace(readOriginalText("provider_code")) == "" || strings.TrimSpace(readOriginalText("provider_name")) == "" {
			return errors.New("tracking provider identity is required")
		}
		if environment := readOriginalText("environment"); environment != "production" && environment != "sandbox" {
			return errors.New("tracking environment must be production or sandbox")
		}
		for _, column := range []string{"polling_interval_minutes", "request_timeout_seconds", "sort_order"} {
			if _, err := readOriginalInteger(column, 0); err != nil {
				return err
			}
		}
		for _, column := range []string{"webhook_enabled", "auto_register", "polling_enabled"} {
			if string(fields[column]) != "true" && string(fields[column]) != "false" {
				return errors.New("original tracking switch is required: " + column)
			}
		}
		if readOriginalText("enabled") == "true" && readOriginalText("deleted_at") == "" {
			code := strings.ToLower(strings.TrimSpace(readOriginalText("provider_code")))
			if code != "mock" && code != "17track" && code != "track17" {
				return errors.New("enabled tracking provider is unsupported")
			}
			if code != "mock" {
				endpoint, err := url.Parse(readOriginalText("base_url"))
				if err != nil || endpoint == nil || (endpoint.Scheme != "https" && endpoint.Scheme != "http") || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.Fragment != "" || strings.TrimSpace(readOriginalText("api_key")) == "" {
					return errors.New("enabled tracking provider requires an original HTTP endpoint and credential")
				}
			}
			if readOriginalText("webhook_enabled") == "true" && strings.TrimSpace(readOriginalText("webhook_secret")) == "" {
				return errors.New("enabled tracking webhook requires its original secret")
			}
		}
	case "shipping_tracking_carrier_mappings":
		if _, err := readOriginalInteger("provider_id", 1); err != nil {
			return err
		}
		if _, err := readOriginalInteger("priority", 0); err != nil {
			return err
		}
		if strings.TrimSpace(readOriginalText("provider_carrier_code")) == "" {
			return errors.New("original provider carrier code is required")
		}
		target, other := "carrier_id", "carrier_service_id"
		if scope := readOriginalText("scope"); scope == "carrier_service" {
			target, other = other, target
		} else if scope != "carrier" {
			return errors.New("tracking mapping scope is invalid")
		}
		if _, err := readOriginalInteger(target, 1); err != nil {
			return err
		}
		if readOriginalText(other) != "" {
			return errors.New("tracking mapping must preserve one local target for its scope")
		}
	default:
		return errors.New("unregistered collection or tracking configuration table")
	}
	return nil
}

// A partial mapping restore also checks its retained provider/local target.
// Restoring a carrier/service/provider includes existing and target-only maps.
func validateConfigurationModuleRestoredTrackingProviderMappingAndTaxCollectionControlRelationships(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage) error {
	if err := validateConfigurationModuleMissingTrackingProviderMappingAndTaxCollectionControlDomainRows(missing); err != nil {
		return err
	}
	providerIDs, mappingIDs, carrierIDs, serviceIDs := []string{}, []string{}, []string{}, []string{}
	for _, pair := range []struct {
		table string
		ids   *[]string
	}{{"shipping_tracking_providers", &providerIDs}, {"shipping_tracking_carrier_mappings", &mappingIDs}, {"carriers", &carrierIDs}, {"shipping_carrier_services", &serviceIDs}} {
		for _, row := range missing[pair.table] {
			id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
			*pair.ids = append(*pair.ids, id)
		}
	}
	if len(providerIDs)+len(mappingIDs)+len(carrierIDs)+len(serviceIDs) == 0 {
		return nil
	}
	mappings, err := readProductCatalogRows(ctx, tx, "shipping_tracking_carrier_mappings", `r.id::text=ANY($1) OR r.provider_id::text=ANY($2) OR r.carrier_id::text=ANY($3) OR r.carrier_service_id::text=ANY($4)`, pq.Array(mappingIDs), pq.Array(providerIDs), pq.Array(carrierIDs), pq.Array(serviceIDs))
	if err != nil {
		return err
	}
	for _, row := range mappings {
		if err := validateConfigurationModuleRestoredTrackingProviderMappingOrTaxCollectionControlCompleteRow("shipping_tracking_carrier_mappings", row); err != nil {
			return errors.New("受影响的现有承运追踪映射字段无效")
		}
		id, _ := readBusinessModuleSnapshotRowColumnText(row, "id")
		mappingIDs = append(mappingIDs, id)
		provider, _ := readBusinessModuleSnapshotRowColumnText(row, "provider_id")
		providerIDs = append(providerIDs, provider)
	}
	providers, err := readProductCatalogRows(ctx, tx, "shipping_tracking_providers", `r.id::text=ANY($1)`, pq.Array(providerIDs))
	if err != nil {
		return err
	}
	for _, row := range providers {
		if err := validateConfigurationModuleRestoredTrackingProviderMappingOrTaxCollectionControlCompleteRow("shipping_tracking_providers", row); err != nil {
			return errors.New("受影响的原追踪服务商配置无效；不会替换凭据或开关")
		}
	}
	var invalid bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS (
SELECT 1 FROM public.shipping_tracking_providers p JOIN public.shipping_tracking_providers other
 ON p.id<>other.id AND lower(btrim(p.provider_code))=lower(btrim(other.provider_code))
WHERE p.id::text=ANY($1) AND p.deleted_at IS NULL AND other.deleted_at IS NULL
) OR EXISTS (
SELECT 1 FROM public.shipping_tracking_carrier_mappings m
LEFT JOIN public.shipping_tracking_providers p ON p.id=m.provider_id
LEFT JOIN public.carriers c ON c.id=m.carrier_id
LEFT JOIN public.shipping_carrier_services s ON s.id=m.carrier_service_id
LEFT JOIN public.carriers service_carrier ON service_carrier.id=s.carrier_id
WHERE m.id::text=ANY($2) AND m.enabled AND m.deleted_at IS NULL AND
 (p.id IS NULL OR p.deleted_at IS NOT NULL OR
 (m.scope='carrier' AND (c.id IS NULL OR c.deleted_at IS NOT NULL OR NOT c.enabled)) OR
 (m.scope='carrier_service' AND (s.id IS NULL OR s.deleted_at IS NOT NULL OR NOT s.enabled OR service_carrier.id IS NULL OR service_carrier.deleted_at IS NOT NULL OR NOT service_carrier.enabled)))
)`, pq.Array(providerIDs), pq.Array(mappingIDs)).Scan(&invalid)
	if err != nil {
		return err
	}
	if invalid {
		return errors.New("追踪服务商代码有歧义，或启用的承运映射缺少有效原服务商/承运商/服务")
	}
	return nil
}

package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"commerce-platform/internal/domain/fitmentcatalog"
)

// Parent-first order is part of this configuration write contract. These are
// catalog facts; quick-buy sessions, calculator histories and stock stay read-only.
func configurationModuleMissingFitmentSpokeAndTireCatalogWritableTableNames() []string {
	return []string{
		"fitment_hub_specifications", "spoke_rim_brands", "spoke_hub_brands",
		"spoke_rim_models", "spoke_hub_models", "spoke_build_presets",
		"fitment_frame_entries", "fitment_fork_entries",
		"fitment_frame_hub_specifications", "fitment_fork_hub_specifications",
		"schwalbe_tire_specifications", "schwalbe_tire_hookless_compatibility",
	}
}

// Reuse domain validation on a copy, without normalizing or rewriting backup
// bytes. SQL constraints still enforce the original full row during insertion.
func validateConfigurationModuleMissingFitmentCatalogDomainRows(missing map[string][]json.RawMessage) error {
	for _, table := range []string{"fitment_hub_specifications", "fitment_frame_entries", "fitment_fork_entries"} {
		for _, row := range missing[table] {
			var record interface{ Validate() error }
			switch table {
			case "fitment_hub_specifications":
				record = &fitmentcatalog.HubSpecification{}
			case "fitment_frame_entries":
				record = &fitmentcatalog.FrameFitmentEntry{}
			case "fitment_fork_entries":
				record = &fitmentcatalog.ForkFitmentEntry{}
			}
			if err := json.Unmarshal(row, record); err != nil {
				return fmt.Errorf("适配目录完整行格式无效：%s", table)
			}
			if err := record.Validate(); err != nil {
				return fmt.Errorf("适配目录字段不符合当前领域规则：%s", table)
			}
		}
	}
	return nil
}

// Validate the final SQL graph, including target-only associations on affected
// parents. FK existence alone cannot establish front/rear or model/brand identity.
func validateConfigurationModuleRestoredFitmentCatalogBusinessRelationships(ctx context.Context, tx *sql.Tx, missing map[string][]json.RawMessage) error {
	for _, group := range []struct {
		parentTable, associationTable, parentColumn, position string
	}{
		{"fitment_frame_entries", "fitment_frame_hub_specifications", "frame_entry_id", "rear"},
		{"fitment_fork_entries", "fitment_fork_hub_specifications", "fork_entry_id", "front"},
	} {
		ids := []string{}
		for _, entry := range []struct{ table, column string }{{group.parentTable, "id"}, {group.associationTable, group.parentColumn}} {
			for _, row := range missing[entry.table] {
				id, err := readBusinessModuleSnapshotRowColumnText(row, entry.column)
				if err != nil || id == "" {
					return errors.New("适配目录缺少完整父记录身份")
				}
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			continue
		}
		var invalid bool
		query := `SELECT EXISTS(SELECT 1 FROM public.` + pq.QuoteIdentifier(group.parentTable) + ` p
WHERE p.id::text=ANY($1) AND (
 EXISTS(SELECT 1 FROM public.` + pq.QuoteIdentifier(group.associationTable) + ` a
 LEFT JOIN public.fitment_hub_specifications h ON h.id=a.hub_specification_id
 WHERE a.` + pq.QuoteIdentifier(group.parentColumn) + `=p.id
 AND (h.id IS NULL OR h.position<>$2 OR h.deleted_at IS NOT NULL OR (p.is_enabled AND NOT h.is_enabled)))
 OR (p.deleted_at IS NULL AND p.is_enabled AND NOT EXISTS(
 SELECT 1 FROM public.` + pq.QuoteIdentifier(group.associationTable) + ` a
 JOIN public.fitment_hub_specifications h ON h.id=a.hub_specification_id
 WHERE a.` + pq.QuoteIdentifier(group.parentColumn) + `=p.id AND h.position=$2 AND h.deleted_at IS NULL AND h.is_enabled))))`
		if err := tx.QueryRowContext(ctx, query, pq.Array(ids), group.position).Scan(&invalid); err != nil {
			return err
		}
		if invalid {
			return fmt.Errorf("适配目录前后轮、启用状态或关联规格不一致：%s", group.parentTable)
		}
	}
	ids := []string{}
	for _, row := range missing["spoke_build_presets"] {
		id, err := readBusinessModuleSnapshotRowColumnText(row, "id")
		if err != nil || id == "" {
			return errors.New("辐条计算预设缺少完整身份")
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}
	var invalid bool
	query := `SELECT EXISTS(SELECT 1 FROM public.spoke_build_presets p
JOIN public.spoke_rim_models rim ON rim.id=p.rim_model_id
JOIN public.spoke_hub_models hub ON hub.id=p.hub_model_id
JOIN public.spoke_rim_brands rb ON rb.id=p.rim_brand_id
JOIN public.spoke_hub_brands hb ON hb.id=p.hub_brand_id
LEFT JOIN public.fitment_hub_specifications specification ON specification.id=hub.fitment_hub_specification_id
WHERE p.id::text=ANY($1) AND (
 rim.brand_id<>p.rim_brand_id OR hub.brand_id<>p.hub_brand_id
 OR (p.deleted_at IS NULL AND p.is_enabled AND (
 rim.deleted_at IS NOT NULL OR hub.deleted_at IS NOT NULL OR rb.deleted_at IS NOT NULL OR hb.deleted_at IS NOT NULL
 OR NOT rim.is_enabled OR NOT hub.is_enabled OR NOT rb.is_enabled OR NOT hb.is_enabled
 OR (hub.fitment_hub_specification_id IS NOT NULL AND (
 specification.deleted_at IS NOT NULL OR NOT specification.is_enabled
 OR (p.wheel_position<>'auto' AND p.wheel_position<>specification.position)))))))`
	if err := tx.QueryRowContext(ctx, query, pq.Array(ids)).Scan(&invalid); err != nil {
		return err
	}
	if invalid {
		return errors.New("辐条计算预设的型号品牌、启用状态或前后轮归属不一致")
	}
	return nil
}

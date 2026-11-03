package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"

	"commerce-platform/internal/domain/outbox"
)

const mediaModuleFormalRestorePolicy = "restore_missing_media_and_repair_registered_file_bytes_preserve_current_business_data_v2"

type MediaModuleRestoreEligibility struct {
	Policy                       string   `json:"policy"`
	Eligible                     bool     `json:"eligible"`
	Blockers                     []string `json:"blockers"`
	ExistingFileRepairs          int      `json:"existing_file_repairs,omitempty"`
	ExistingFileRepairPlanSHA256 string   `json:"existing_file_repair_plan_sha256,omitempty"`
}

// Parent-first writes are bounded to this standalone module. Shared products,
// conversations, users, configuration and cleanup events are never written.
func mediaModuleOwnedTableNames() []string {
	return strings.Fields("media media_derivative_presets media_assets media_asset_derivatives media_upload_sessions business_draft_media_upload_receipts product_media ticket_message_media_attachments site_logo_assets site_favicon_assets media_derivative_rebuild_jobs")
}

func mediaModuleHasDedicatedMissingRowContract(snapshot selectedBusinessModuleCompleteSnapshot, table string) bool {
	if !reflect.DeepEqual(snapshot.Modules, []string{"media"}) {
		return false
	}
	for _, name := range mediaModuleOwnedTableNames() {
		if name == table {
			return true
		}
	}
	return false
}

func validateMediaModuleInsertOnlyWriteContract(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, table string, observation *moduleRecoveryDatabaseObservation) error {
	if !mediaModuleHasDedicatedMissingRowContract(snapshot, table) {
		return errors.New("媒体写入表未登记或不是单独媒体包")
	}
	return validateBusinessModuleInsertOnlyTableContractWithUpdateOnlyGuardsPreserved(ctx, tx, table, observation)
}

func readMediaModuleRestoreRowText(row json.RawMessage, column string) string {
	value, _ := readBusinessModuleSnapshotRowColumnText(row, column)
	return value
}

func inspectMediaModuleMissingRowLifecycle(table string, row json.RawMessage) string {
	field := func(column string) string { return readMediaModuleRestoreRowText(row, column) }
	if field("deleted_at") != "" {
		return "已删除的媒体登记不能重新补入，请先处理历史删除状态"
	}
	switch table {
	case "media_assets":
		if field("status") != "active" || field("customer_service_derivative_write_token") != "" || field("customer_service_derivative_write_expires_at") != "" {
			return "媒体不是有效原件或仍有派生写入租约"
		}
	case "media_upload_sessions", "business_draft_media_upload_receipts":
		if field("status") != "bound" {
			return "未绑定的历史上传不能恢复为待清理或待执行状态"
		}
	case "media_derivative_rebuild_jobs":
		if field("status") != "succeeded" || field("finished_at") == "" || field("locked_at") != "" || field("locked_by") != "" || field("lease_expires_at") != "" {
			return "只可补回已完成且没有执行租约的派生任务历史"
		}
	}
	return ""
}

func (cfg Configuration) inspectMediaModuleRestoreEligibility(ctx context.Context, tx *sql.Tx, snapshot selectedBusinessModuleCompleteSnapshot, missing map[string][]json.RawMessage, manifest *Manifest) (*MediaModuleRestoreEligibility, error) {
	result := &MediaModuleRestoreEligibility{Policy: mediaModuleFormalRestorePolicy, Eligible: true, Blockers: []string{}}
	record := func(reason string) {
		for _, previous := range result.Blockers {
			if previous == reason {
				return
			}
		}
		result.Eligible = false
		result.Blockers = append(result.Blockers, reason)
	}
	if !reflect.DeepEqual(snapshot.Modules, []string{"media"}) || manifest == nil || !reflect.DeepEqual(manifest.SelectedModules, snapshot.Modules) {
		record("需要单独图片与附件包")
		return result, nil
	}
	for table, rows := range missing {
		if len(rows) > 0 && !mediaModuleHasDedicatedMissingRowContract(snapshot, table) {
			record("包包含未授权写入表：" + table)
		}
		for _, row := range rows {
			if reason := inspectMediaModuleMissingRowLifecycle(table, row); reason != "" {
				record(reason + "（" + table + "）")
			}
		}
	}
	// The application is stopped, but unfinished work could resume on release.
	// Never reset a lease, mark work done, or replay a cleanup to gain eligibility.
	for _, check := range []struct{ query, reason string }{
		{`SELECT EXISTS(SELECT 1 FROM public.media_derivative_rebuild_jobs WHERE status<>'succeeded' OR locked_at IS NOT NULL OR locked_by<>'' OR lease_expires_at IS NOT NULL)`, "当前仍有未完成或持有租约的派生任务，请先在正常系统中完成处置"},
		{`SELECT EXISTS(SELECT 1 FROM public.media_assets WHERE customer_service_derivative_write_token<>'' OR customer_service_derivative_write_expires_at IS NOT NULL)`, "当前媒体仍有派生写入租约"},
		{`SELECT EXISTS(SELECT 1 FROM public.media_upload_sessions WHERE status<>'bound') OR EXISTS(SELECT 1 FROM public.business_draft_media_upload_receipts WHERE status<>'bound')`, "当前仍有未绑定上传，请先完成上传或由正常清理流程处理"},
	} {
		var unsafe bool
		if err := tx.QueryRowContext(ctx, check.query).Scan(&unsafe); err != nil {
			return nil, err
		}
		if unsafe {
			record(check.reason)
		}
	}
	keys, assetIDs := map[string]bool{}, map[string]bool{}
	for _, entry := range manifest.Files {
		if strings.HasPrefix(entry.Path, "media/") {
			_, key, err := transactionalBusinessModuleRestoreMediaPath(entry.Path)
			if err != nil {
				return nil, err
			}
			keys[key] = true
		}
	}
	parents, derivatives := map[string]productCatalogMediaObjectRegistration{}, map[string]productCatalogMediaObjectRegistration{}
	for _, table := range []string{"media_assets", "media_asset_derivatives"} {
		current, err := readProductCatalogRows(ctx, tx, table, "true")
		if err != nil {
			return nil, err
		}
		if len(current)+len(snapshot.Tables[table].Rows) > businessModuleSnapshotMaximumRows {
			return nil, errors.New("媒体清理引用检查超过容量限制")
		}
		for _, row := range append(current, snapshot.Tables[table].Rows...) {
			field := func(column string) string { return readMediaModuleRestoreRowText(row, column) }
			registration := productCatalogMediaObjectRegistration{ID: field("id"), ParentID: field("media_asset_id"), Key: field("storage_key"), URL: field("url"), DeletedAt: field("deleted_at")}
			if registration.Key == "" {
				_, registration.Key = cfg.resolveLocalMediaReference(registration.URL)
			}
			if table == "media_assets" {
				parents[registration.ID] = registration
				assetIDs[registration.ID] = true
			} else {
				derivatives[registration.ID] = registration
			}
		}
	}
	events, err := readProductCatalogRows(ctx, tx, "outbox_events", "event_type ILIKE '%cleanup%'")
	if err != nil {
		return nil, err
	}
	if len(events) > 20000 {
		return nil, errors.New("媒体清理事件检查超过容量限制")
	}
	for _, row := range events {
		if err := cfg.validateProductCatalogRetainedCleanupReceipt(row, keys, assetIDs, parents, derivatives); err != nil {
			record("当前清理事件可能影响恢复文件，须先核对原始事件（不允许在恢复中重放或重置）")
			continue
		}
		// The product check deliberately excludes private retention references;
		// standalone media owns all roots and must protect these references too.
		if readMediaModuleRestoreRowText(row, "event_type") == outbox.EventTypeCustomerServiceRetentionCleanup {
			var fields map[string]json.RawMessage
			var payload outbox.CustomerServiceRetentionCleanupPayload
			if json.Unmarshal(row, &fields) != nil || decodeStrictJSON(fields["payload"], &payload) != nil {
				record("保留期清理事件无法核对")
				continue
			}
			execution := customerServiceRecoveryCleanupReceiptFutureExecution(outbox.EventTypeCustomerServiceRetentionCleanup, readMediaModuleRestoreRowText(row, "status"), readMediaModuleRestoreRowText(row, "deleted_at") != "", readMediaModuleRestoreRowText(row, "locked_at") != "" || readMediaModuleRestoreRowText(row, "locked_by") != "")
			for _, reference := range payload.AttachmentReferences {
				_, key := cfg.resolveLocalMediaReference(reference)
				if keys[key] && execution != "suppressed_original_receipt" {
					record("原保留期清理事件仍引用包内附件，须先完成事件处置")
				}
			}
		}
	}
	sort.Strings(result.Blockers)
	return result, nil
}

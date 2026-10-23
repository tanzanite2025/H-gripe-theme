package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

func readAndValidateProductCatalogModuleSnapshot(extracted string, manifest Manifest) (productCatalogModuleSnapshot, error) {
	var snapshot productCatalogModuleSnapshot
	info, err := os.Stat(filepath.Join(extracted, productCatalogRowsFilename))
	if err != nil {
		return snapshot, err
	}
	if info.Size() > productCatalogMaximumJSONBytes {
		return snapshot, errors.New("product snapshot exceeds JSON capacity")
	}
	data, err := os.ReadFile(filepath.Join(extracted, productCatalogRowsFilename))
	if err != nil {
		return snapshot, err
	}
	if err := decodeStrictJSON(data, &snapshot); err != nil {
		return snapshot, err
	}
	if snapshot.ContractVersion != 1 || manifest.DatabaseFingerprint != productCatalogBytesSHA256(data) {
		return snapshot, errors.New("product snapshot contract or content fingerprint mismatch")
	}
	allowed := productCatalogAllowedTableNames()
	if len(snapshot.Tables) != len(allowed) || len(manifest.Tables) != len(allowed) {
		return snapshot, errors.New("product snapshot has missing or unregistered tables")
	}
	total := 0
	for name, table := range snapshot.Tables {
		if !allowed[name] || int64(len(table.Rows)) != manifest.Tables[name] {
			return snapshot, errors.New("product table inventory mismatch")
		}
		total += len(table.Rows)
		if total > productCatalogMaximumRows {
			return snapshot, errors.New("product row capacity exceeded")
		}
		ids := map[string]bool{}
		for _, row := range table.Rows {
			id, err := productCatalogRowString(row, "id")
			numericID, parseErr := strconv.ParseInt(id, 10, 64)
			if err != nil || parseErr != nil || numericID <= 0 || strconv.FormatInt(numericID, 10) != id || ids[id] {
				return snapshot, errors.New("invalid or duplicate product row identity")
			}
			ids[id] = true
		}
	}
	return snapshot, nil
}

func captureProductCatalogPreservedRowsFingerprint(ctx context.Context, tx *sql.Tx, snapshot productCatalogModuleSnapshot) (map[string]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY tablename`)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, err
		}
		names = append(names, name)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, name := range names {
		predicate := "true"
		arguments := []any{}
		if table, selected := snapshot.Tables[name]; selected {
			ids := []string{}
			for _, row := range table.Rows {
				id, err := productCatalogRowString(row, "id")
				if err != nil {
					return nil, err
				}
				ids = append(ids, id)
			}
			predicate = "NOT (r.id::text=ANY($1))"
			arguments = append(arguments, pq.Array(ids))
		}
		rows, err := tx.QueryContext(ctx, `SELECT to_jsonb(r)::text FROM public.`+pq.QuoteIdentifier(name)+` r WHERE `+predicate+` ORDER BY to_jsonb(r)::text`, arguments...)
		if err != nil {
			return nil, err
		}
		digest := sha256.New()
		for rows.Next() {
			var raw string
			if err := rows.Scan(&raw); err != nil {
				rows.Close()
				return nil, err
			}
			fmt.Fprintf(digest, "%d:%s\n", len(raw), raw)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		result[name] = fmt.Sprintf("%x", digest.Sum(nil))
	}
	return result, nil
}

func (cfg Configuration) applyProductCatalogRowsOnlyToIsolatedDatabase(ctx context.Context, db *sql.DB, snapshot productCatalogModuleSnapshot) (int, int, *ProductCatalogInventoryPreservationResult, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, 0, nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL timezone='UTC'`); err != nil {
		return 0, 0, nil, err
	}
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path=pg_catalog,public`); err != nil {
		return 0, 0, nil, err
	}
	var databaseName string
	var isolatedOwner bool
	if err := tx.QueryRowContext(ctx, `SELECT current_database(),pg_get_userbyid(datdba)='local_backup_isolated_owner' FROM pg_database WHERE datname=current_database()`).Scan(&databaseName, &isolatedOwner); err != nil {
		return 0, 0, nil, err
	}
	if !strings.HasPrefix(databaseName, "local_backup_stage_") || !safeDatabaseName.MatchString(databaseName) || !isolatedOwner {
		return 0, 0, nil, errors.New("product restore accepts only a maintenance-owned isolated database")
	}
	protectedCategoryIdentities, err := lockProductCatalogWriteTablesAndValidateStandardBusinessContractsBeforeMerge(ctx, tx, snapshot)
	if err != nil {
		return 0, 0, nil, err
	}
	if err := cfg.validateProductCatalogIncomingMediaOwnersAndUnregisteredLegacyURLsBeforeMerge(ctx, tx, snapshot); err != nil {
		return 0, 0, nil, err
	}
	if err := cfg.validateProductCatalogIncomingRichTextLegacyMediaOccurrencesBeforeMerge(ctx, tx, snapshot); err != nil {
		return 0, 0, nil, err
	}
	before, err := captureProductCatalogPreservedRowsFingerprint(ctx, tx, snapshot)
	if err != nil {
		return 0, 0, nil, err
	}
	dependencies, err := captureProductCatalogExternalDependencies(ctx, tx, &snapshot)
	if err != nil {
		return 0, 0, nil, err
	}
	if !reflect.DeepEqual(dependencies, snapshot.ExternalDependencies) {
		return 0, 0, nil, errors.New("shared product dependencies differ; restore requires explicit dependency reconciliation")
	}
	prepared, inventoryPreservation, inventoryFacts, err := prepareProductCatalogSnapshotWithCurrentInventoryAndStableBusinessIdentities(ctx, tx, snapshot)
	if err != nil {
		return 0, 0, nil, err
	}
	snapshot = prepared
	names := []string{}
	for name := range snapshot.Tables {
		names = append(names, name)
	}
	sort.Strings(names)
	// This transaction runs only in a clone. Suppress application triggers and
	// validate every affected FK explicitly before committing; no event is emitted.
	if _, err := tx.ExecContext(ctx, `SET LOCAL session_replication_role='replica'`); err != nil {
		return 0, 0, nil, err
	}
	restoredRows := 0
	for _, name := range names {
		table := snapshot.Tables[name]
		schema, columns, err := inspectProductCatalogTableSchema(ctx, tx, name)
		if err != nil {
			return 0, 0, nil, err
		}
		if schema != table.SchemaSHA256 {
			return 0, 0, nil, fmt.Errorf("product schema mismatch: %s", name)
		}
		quotedColumns := []string{}
		assignments := []string{}
		for _, column := range columns {
			quoted := pq.QuoteIdentifier(column)
			quotedColumns = append(quotedColumns, quoted)
			if column != "id" {
				assignments = append(assignments, quoted+"=EXCLUDED."+quoted)
			}
		}
		for _, row := range table.Rows {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(row, &fields); err != nil {
				return 0, 0, nil, err
			}
			if len(fields) != len(columns) {
				return 0, 0, nil, errors.New("product row field inventory differs from schema")
			}
			for _, column := range columns {
				if _, ok := fields[column]; !ok {
					return 0, 0, nil, errors.New("product row lacks a required persisted field")
				}
			}
			if name == "media_assets" || name == "media_asset_derivatives" {
				id, _ := productCatalogRowString(row, "id")
				var same bool
				err := tx.QueryRowContext(ctx, `SELECT to_jsonb(r)=$1::jsonb FROM public.`+pq.QuoteIdentifier(name)+` r WHERE id::text=$2`, string(row), id).Scan(&same)
				if err != nil && err != sql.ErrNoRows {
					return 0, 0, nil, err
				}
				if err == nil && !same {
					return 0, 0, nil, errors.New("shared media identity content differs; product restore blocked")
				}
			}
			statement := `INSERT INTO public.` + pq.QuoteIdentifier(name) + ` (` + strings.Join(quotedColumns, ",") + `) SELECT ` + strings.Join(quotedColumns, ",") + ` FROM jsonb_populate_record(NULL::public.` + pq.QuoteIdentifier(name) + `,$1::jsonb) ON CONFLICT (id) DO UPDATE SET ` + strings.Join(assignments, ",")
			if _, err := tx.ExecContext(ctx, statement, string(row)); err != nil {
				return 0, 0, nil, fmt.Errorf("isolated product row restore failed in %s; transaction rolled back", name)
			}
			restoredRows++
		}
	}
	if err := validateProductCatalogAffectedForeignKeys(ctx, tx, snapshot); err != nil {
		return 0, 0, nil, err
	}
	if err := validateProductCatalogSelectedAndRetainedSharedInventoryOwners(ctx, tx, snapshot); err != nil {
		return 0, 0, nil, err
	}
	actualCategoryIdentities, err := inspectProductCatalogRequiredSystemCategoriesAndCarbonRimBindings(ctx, tx)
	if err != nil {
		return 0, 0, nil, err
	}
	if actualCategoryIdentities != protectedCategoryIdentities {
		return 0, 0, nil, errors.New("product restore cannot reassign protected system category identities")
	}
	currentInventoryFingerprint, err := captureProductCatalogExistingInventoryAndIdentityFingerprint(ctx, tx, inventoryFacts)
	if err != nil || currentInventoryFingerprint != inventoryFacts.PreservedFingerprint {
		return 0, 0, nil, errors.New("existing inventory, product identity, or current view counts changed during product restore")
	}
	inventoryPreservation.ExistingInventoryAndIdentityFactsVerified = true
	after, err := captureProductCatalogPreservedRowsFingerprint(ctx, tx, snapshot)
	if err != nil {
		return 0, 0, nil, err
	}
	if !reflect.DeepEqual(before, after) {
		return 0, 0, nil, errors.New("unselected or target-only records changed during product restore")
	}
	if _, err := cfg.inspectProductCatalogRequiredImagesAndRetainedCleanupReceipts(ctx, tx); err != nil {
		return 0, 0, nil, err
	}
	// PostgreSQL sequences do not roll back. Advance only after all row and FK
	// checks, in the disposable clone; never decrease a current sequence.
	for _, name := range names {
		var sequence sql.NullString
		if err := tx.QueryRowContext(ctx, `SELECT pg_get_serial_sequence($1,'id')`, "public."+name).Scan(&sequence); err != nil {
			return 0, 0, nil, err
		}
		if sequence.Valid {
			var maximum, current int64
			var called bool
			if err := tx.QueryRowContext(ctx, `SELECT COALESCE(max(id),0) FROM public.`+pq.QuoteIdentifier(name)).Scan(&maximum); err != nil {
				return 0, 0, nil, err
			}
			parts := strings.Split(sequence.String, ".")
			if len(parts) != 2 {
				return 0, 0, nil, errors.New("unexpected product sequence namespace")
			}
			if err := tx.QueryRowContext(ctx, `SELECT last_value,is_called FROM `+pq.QuoteIdentifier(parts[0])+`.`+pq.QuoteIdentifier(parts[1])).Scan(&current, &called); err != nil {
				return 0, 0, nil, err
			}
			if maximum > current || (maximum == current && !called) {
				if _, err := tx.ExecContext(ctx, `SELECT setval($1::regclass,$2,true)`, sequence.String, maximum); err != nil {
					return 0, 0, nil, err
				}
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, nil, err
	}
	return restoredRows, len(before) - len(snapshot.Tables), &inventoryPreservation, nil
}

func validateProductCatalogAffectedForeignKeys(ctx context.Context, tx *sql.Tx, snapshot productCatalogModuleSnapshot) error {
	rows, err := tx.QueryContext(ctx, `SELECT t.relname,c.conname,p.relname,pg_get_constraintdef(c.oid),c.convalidated FROM pg_constraint c JOIN pg_class t ON t.oid=c.conrelid JOIN pg_class p ON p.oid=c.confrelid WHERE c.contype='f' AND t.relnamespace='public'::regnamespace ORDER BY t.relname,c.conname`)
	if err != nil {
		return err
	}
	type constraint struct{ table, name, definition string }
	constraints := []constraint{}
	for rows.Next() {
		var table, name, parent, definition string
		var validated bool
		if err := rows.Scan(&table, &name, &parent, &definition, &validated); err != nil {
			rows.Close()
			return err
		}
		_, childSelected := snapshot.Tables[table]
		_, parentSelected := snapshot.Tables[parent]
		if childSelected || parentSelected {
			if !validated {
				rows.Close()
				return errors.New("unvalidated product dependency constraint")
			}
			constraints = append(constraints, constraint{table, name, definition})
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, constraint := range constraints {
		table := `public.` + pq.QuoteIdentifier(constraint.table)
		name := pq.QuoteIdentifier(constraint.name)
		if _, err := tx.ExecContext(ctx, `ALTER TABLE `+table+` DROP CONSTRAINT `+name+`, ADD CONSTRAINT `+name+` `+constraint.definition+` NOT VALID`); err != nil {
			return errors.New("could not rebuild isolated product dependency validation")
		}
		if _, err := tx.ExecContext(ctx, `ALTER TABLE `+table+` VALIDATE CONSTRAINT `+name); err != nil {
			return errors.New("product restore would break a selected or preserved foreign-key relationship")
		}
	}
	return nil
}

func (manager *Manager) restoreProductCatalogIntoIsolatedTargetClone(ctx context.Context, task *Task, extracted string) (resultErr error) {
	cfg := manager.Configuration
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return err
	}
	snapshot, err := readAndValidateProductCatalogModuleSnapshot(extracted, *task.Manifest)
	if err != nil {
		return err
	}
	if err := validateProductCatalogPackageFileInventory(snapshot, *task.Manifest, cfg); err != nil {
		return err
	}
	requiredWorkspace, err := cfg.estimateBackupWorkspaceBytes(ctx)
	if err != nil {
		return err
	}
	if err := cfg.checkWorkspaceCapacity(requiredWorkspace); err != nil {
		return err
	}
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer lease.Close()
	defer func() { resultErr = errors.Join(resultErr, manager.releaseMaintenance(task)) }()
	return manager.restoreProductCatalogIntoIsolatedTargetCloneWithHeldMaintenanceLease(ctx, task, extracted, snapshot)
}

// Caller retains the exclusive application lease throughout drain and rebuild.
func (manager *Manager) restoreProductCatalogIntoIsolatedTargetCloneWithHeldMaintenanceLease(ctx context.Context, task *Task, extracted string, snapshot productCatalogModuleSnapshot) error {
	cfg := manager.Configuration
	previewBinding, err := manager.captureProductCatalogRecoveryPreviewBaseline(ctx)
	if err != nil {
		return err
	}
	directory, _ := manager.taskDirectory(task.ID)
	candidate := filepath.Join(directory, "product-isolated-target")
	if err := os.RemoveAll(candidate); err != nil {
		return err
	}
	if err := os.Mkdir(candidate, 0700); err != nil {
		return err
	}
	// Target data stays in the private server workspace and never enters the
	// downloadable product package. It is needed to prove preservation.
	defer os.Remove(filepath.Join(candidate, "database.dump"))
	current, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	contractTx, err := current.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		current.Close()
		return err
	}
	if err := requireRegisteredStorefrontTablesForProductCatalogModule(ctx, contractTx); err != nil {
		contractTx.Rollback()
		current.Close()
		return err
	}
	if err := validateProductCatalogStandardTriggersAndRejectUnregisteredRewriteRules(ctx, contractTx); err != nil {
		contractTx.Rollback()
		current.Close()
		return err
	}
	if _, err := inspectProductCatalogRequiredSystemCategoriesAndCarbonRimBindings(ctx, contractTx); err != nil {
		contractTx.Rollback()
		current.Close()
		return err
	}
	for name, table := range snapshot.Tables {
		constraints, err := captureProductCatalogConstraintFingerprint(ctx, contractTx, name)
		if err != nil || constraints != table.ConstraintsSHA256 {
			contractTx.Rollback()
			current.Close()
			return fmt.Errorf("product constraint contract differs in current target: %s", name)
		}
	}
	err = contractTx.Commit()
	current.Close()
	if err != nil {
		return err
	}
	var targetManifest Manifest
	if err := cfg.captureSnapshotAndDump(ctx, candidate, &targetManifest); err != nil {
		return err
	}
	if targetManifest.MigrationVersion != cfg.ExpectedMigrationVersion {
		return errors.New("target migration changed before product validation")
	}
	mediaBefore, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	if task.StageDatabase != "" {
		if task.StageDatabase != "local_backup_stage_"+task.ID[:16] {
			return errors.New("product stage identity mismatch")
		}
		if err := manager.dropProductCatalogStageOnlyWithBoundIdentity(ctx, task); err != nil {
			return err
		}
	}
	task.ProductCatalogPreviewBinding = nil
	task.StageDatabase = "local_backup_stage_" + task.ID[:16]
	task.Phase = "restoring_product_module_in_isolated_target_clone"
	task.ProductCatalogRecovery = nil
	if err := manager.saveTask(task); err != nil {
		return err
	}
	success := false
	defer func() {
		if !success {
			_ = cfg.dropOwnedStageDatabase(context.Background(), task.StageDatabase)
		}
	}()
	if err := cfg.restoreIsolatedDatabase(ctx, candidate, task.StageDatabase, targetManifest); err != nil {
		return err
	}
	// pg_dump/pg_restore compacts dropped-column slots and may reserialize
	// equivalent casts. Bind the tool-created clone before the module writes;
	// independently keep the original target's exact schema for stopped checks.
	candidateCloneBinding, err := cfg.captureProductCatalogRecoveryDatabaseBinding(ctx, task.StageDatabase)
	if err != nil {
		return err
	}
	stage, err := cfg.connectDatabase(ctx, task.StageDatabase)
	if err != nil {
		return err
	}
	defer stage.Close()
	restored, preserved, inventoryPreservation, err := cfg.applyProductCatalogRowsOnlyToIsolatedDatabase(ctx, stage, snapshot)
	if err != nil {
		return err
	}
	// Clone all current media first; selected files can only fill missing paths
	// or match current bytes. Conflicting paths require a separate decision.
	var candidateMedia Manifest
	if err := cfg.copyMediaSnapshot(ctx, candidate, &candidateMedia); err != nil {
		return err
	}
	if err := manager.prepareProductCatalogCandidateCompleteMediaDirectories(task, previewBinding); err != nil {
		return err
	}
	verifiedImages := 0
	for _, entry := range task.Manifest.Files {
		if entry.Path == productCatalogRowsFilename {
			continue
		}
		destination := filepath.Join(candidate, filepath.FromSlash(entry.Path))
		if _, err := os.Stat(destination); err == nil {
			current, err := fileIdentity(destination, entry.Path)
			if err != nil || current != entry {
				return errors.New("product image conflicts with a retained target path")
			}
		} else if errors.Is(err, os.ErrNotExist) {
			if _, err := copyVerifiedRegularFile(filepath.Join(extracted, filepath.FromSlash(entry.Path)), destination, entry.Path); err != nil {
				return err
			}
		} else {
			return err
		}
		verifiedImages++
	}
	live, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	defer live.Close()
	_, fingerprint, _, err := captureDatabaseInventory(ctx, live)
	if err != nil {
		return err
	}
	mediaAfter, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	if fingerprint != targetManifest.DatabaseFingerprint || mediaBefore != mediaAfter {
		return errors.New("target database or media changed during product isolation validation")
	}
	_, task.StageFingerprint, _, err = captureDatabaseInventory(ctx, stage)
	if err != nil {
		return err
	}
	if err := manager.completeProductCatalogRecoveryPreviewBinding(ctx, task, previewBinding, candidateCloneBinding); err != nil {
		return err
	}
	task.PreviewTargetTables = targetManifest.Tables
	task.ProductCatalogRecovery = &ProductCatalogIsolatedRecoveryResult{RestoredRows: restored, VerifiedImageFiles: verifiedImages, PreservedTables: preserved, PreservedTargetOnlyRows: true, OnlineOverwriteAvailable: false, InventoryPreservation: inventoryPreservation}
	task.State = "product_module_isolated_verified"
	task.Phase = "product_rows_images_and_unselected_data_preservation_verified"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	success = true
	return nil
}

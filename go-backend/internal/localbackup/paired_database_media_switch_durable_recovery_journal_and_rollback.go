package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lib/pq"
)

type RecoveryFileMove struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Completed   bool   `json:"completed"`
	Started     bool   `json:"started"`
}

type RecoveryJournal struct {
	OldDatabaseOID              int64                                           `json:"old_database_oid"`
	NewDatabaseOID              int64                                           `json:"new_database_oid"`
	RollbackDatabase            string                                          `json:"rollback_database"`
	Moves                       []RecoveryFileMove                              `json:"moves"`
	DatabaseSwitched            bool                                            `json:"database_switched"`
	ProductCatalog              *productCatalogPairedRehearsalJournal           `json:"product_catalog_rehearsal,omitempty"`
	ProductCatalogStoppedSwitch *ProductCatalogContinuouslyStoppedSwitchJournal `json:"product_catalog_stopped_switch,omitempty"`
}

type ReconciliationEvidence struct {
	ReconciledDatabaseFingerprint        string                                    `json:"reconciled_database_fingerprint"`
	PaymentAndRefundReference            string                                    `json:"payment_and_refund_reference"`
	ShippingReference                    string                                    `json:"shipping_reference"`
	DurableEventsAndJobsReference        string                                    `json:"durable_events_and_jobs_reference"`
	DedicatedRedisRebuildReference       string                                    `json:"dedicated_redis_rebuild_reference"`
	AttachmentAccessAndIdentityReference string                                    `json:"attachment_access_and_identity_reference"`
	Operator                             string                                    `json:"operator"`
	ExternalProjections                  *ExternalProjectionReconciliationEvidence `json:"external_projections"`
}

func (manager *Manager) ApplyVerifiedRecovery(id, confirmation string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	cfg := manager.Configuration
	if task.State != "verified" || task.Manifest == nil || task.Manifest.Scope != "whole_application" || task.Journal != nil {
		return nil, errors.New("a fresh isolated restore and overwrite preview are required")
	}
	if confirmation != "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+task.Manifest.PackageID {
		return nil, errors.New("explicit target, environment and package overwrite confirmation is required")
	}
	if err := manager.startTaskOperation(task, "waiting_for_application_stop", manager.applyPairedRecovery); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func (manager *Manager) applyPairedRecovery(ctx context.Context, task *Task) error {
	if task.Manifest == nil || task.Manifest.Scope != "whole_application" {
		return errors.New("whole-storefront switching requires a whole-application package")
	}
	if err := manager.requireCustomerServicePersistentCandidateDiscarded(task.ID); err != nil {
		return err
	}
	cfg := manager.Configuration
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return err
	}
	if err := cfg.checkPostgreSQLAvailableBytes(minimumPostgreSQLFreeReserveBytes); err != nil {
		return err
	}
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer lease.Close()
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	if err := assertNoUnfencedDatabaseSessions(ctx, control, cfg.DatabaseName); err != nil {
		return err
	}
	fingerprint, _, err := cfg.targetFingerprint(ctx)
	if err != nil {
		return err
	}
	if fingerprint != task.PreviewTargetFingerprint {
		return errors.New("target changed after preview; regenerate the preview before overwrite")
	}
	if err := manager.verifyPreparedRecoveryFiles(task); err != nil {
		return err
	}
	directory, _ := manager.taskDirectory(task.ID)
	if err := cfg.validateWholeApplicationCompanyStaticPackage(*task.Manifest, filepath.Join(directory, "verified")); err != nil {
		return err
	}
	stage, err := cfg.connectDatabase(ctx, task.StageDatabase)
	if err != nil {
		return err
	}
	_, stageFingerprint, _, inventoryErr := captureDatabaseInventory(ctx, stage)
	if inventoryErr != nil || stageFingerprint != task.StageFingerprint {
		stage.Close()
		return errors.New("isolated database changed after validation")
	}
	// Grant the application's configured role ownership before the database is
	// sealed. Nothing is imported into the online database.
	_, err = stage.ExecContext(ctx, `REASSIGN OWNED BY local_backup_isolated_owner TO `+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole))
	stage.Close()
	if err != nil {
		return err
	}
	journal := &RecoveryJournal{RollbackDatabase: "local_backup_old_" + task.ID[:16], Moves: []RecoveryFileMove{}}
	if err := readDatabaseOID(ctx, control, cfg.DatabaseName, &journal.OldDatabaseOID); err != nil {
		return err
	}
	if err := readDatabaseOID(ctx, control, task.StageDatabase, &journal.NewDatabaseOID); err != nil {
		return err
	}
	if journal.NewDatabaseOID == 0 {
		return errors.New("isolated database is missing")
	}
	// Prepare complete replacement trees inside their own mounted filesystems;
	// all later moves are same-filesystem renames, including Docker volumes.
	if err := manager.cleanupUnswitchedPreparedMedia(task); err != nil {
		return err
	}
	for _, storageRoot := range cfg.wholeApplicationRecoveryStorageRoots(wholeApplicationPackageIncludesCompanyStatic(task.Manifest)) {
		rootName, root := storageRoot.Name, storageRoot.Directory
		stageRoot := filepath.Join(root, ".local-backup-stage-"+task.ID)
		oldRoot := filepath.Join(root, ".local-backup-old-"+task.ID)
		if err := os.Mkdir(stageRoot, 0700); err != nil {
			return err
		}
		if err := os.Mkdir(oldRoot, 0700); err != nil {
			return err
		}
		if err := manager.prepareReplacementMediaTree(task, rootName, stageRoot); err != nil {
			return err
		}
		current, err := os.ReadDir(root)
		if err != nil {
			return err
		}
		for _, entry := range current {
			if strings.HasPrefix(entry.Name(), ".local-backup-") {
				continue
			}
			journal.Moves = append(journal.Moves, RecoveryFileMove{Source: filepath.Join(root, entry.Name()), Destination: filepath.Join(oldRoot, entry.Name())})
		}
		replacements, err := os.ReadDir(stageRoot)
		if err != nil {
			return err
		}
		for _, entry := range replacements {
			journal.Moves = append(journal.Moves, RecoveryFileMove{Source: filepath.Join(stageRoot, entry.Name()), Destination: filepath.Join(root, entry.Name())})
		}
	}
	task.Journal = journal
	task.Phase = "paired_switch_journal_persisted"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	// From this point any failure keeps the persistent application fence. The
	// journal identifies each old/new database by OID, independently of names.
	if err := sealDatabase(ctx, control, task.StageDatabase); err != nil {
		return err
	}
	if journal.OldDatabaseOID != 0 {
		if err := sealDatabase(ctx, control, cfg.DatabaseName); err != nil {
			return err
		}
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` RENAME TO `+pq.QuoteIdentifier(journal.RollbackDatabase)); err != nil {
			return err
		}
	}
	task.Phase = "previous_database_retained"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	for index := range journal.Moves {
		journal.Moves[index].Started = true
		if err := manager.saveTask(task); err != nil {
			return err
		}
		if err := performJournaledFileMove(&journal.Moves[index]); err != nil {
			return err
		}
		task.Phase = fmt.Sprintf("switching_media_%d_of_%d", index+1, len(journal.Moves))
		if err := manager.saveTask(task); err != nil {
			return err
		}
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(task.StageDatabase)+` RENAME TO `+pq.QuoteIdentifier(cfg.DatabaseName)); err != nil {
		return err
	}
	journal.DatabaseSwitched = true
	task.Phase = "database_and_media_switched_service_still_isolated"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	// Even database logins remain disabled until explicit reconciliation. A
	// restart cannot accidentally consume restored jobs or process callbacks.
	task.State = "awaiting_reconciliation"
	return manager.saveTask(task)
}

func (manager *Manager) verifyPreparedRecoveryFiles(task *Task) error {
	directory, _ := manager.taskDirectory(task.ID)
	for _, expected := range task.Manifest.Files {
		actual, err := fileIdentity(filepath.Join(directory, "verified", filepath.FromSlash(expected.Path)), expected.Path)
		if err != nil || actual != expected {
			return errors.New("verified recovery files changed; repeat package validation")
		}
	}
	return nil
}

func (manager *Manager) prepareReplacementMediaTree(task *Task, rootName, destination string) error {
	directory, _ := manager.taskDirectory(task.ID)
	prefix := "media/" + rootName + "/"
	if rootName == companyStaticRecoveryRootName {
		if !wholeApplicationPackageIncludesCompanyStatic(task.Manifest) || manager.Configuration.ProductStaticAssetsDirectory == "" {
			return errors.New("company static replacement requires a whole-storefront release-bound package")
		}
		prefix = companyStaticPackagePrefix
	}
	var required int64
	for _, entry := range task.Manifest.Files {
		if strings.HasPrefix(entry.Path, prefix) {
			required += entry.Size
		}
	}
	free, err := availableDiskBytes(destination)
	if err != nil {
		return err
	}
	if uint64(required) > free || free-uint64(required) < 64<<20 {
		return errors.New("insufficient media volume capacity for paired rollback")
	}
	for _, entry := range task.Manifest.Files {
		if !strings.HasPrefix(entry.Path, prefix) {
			continue
		}
		filename := filepath.Join(destination, filepath.FromSlash(strings.TrimPrefix(entry.Path, prefix)))
		actual, err := copyVerifiedRegularFile(filepath.Join(directory, "verified", filepath.FromSlash(entry.Path)), filename, entry.Path)
		if err != nil {
			return err
		}
		if actual != entry {
			return errors.New("replacement media verification failed")
		}
	}
	return filepath.WalkDir(destination, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filename == destination {
			return nil
		}
		mode := os.FileMode(0600)
		if entry.IsDir() {
			mode = 0700
		}
		if rootName != "uploads_private" {
			mode = 0644
			if entry.IsDir() {
				mode = 0755
			}
		}
		if err := os.Chmod(filename, mode); err != nil {
			return err
		}
		if entry.IsDir() {
			return syncDirectory(filename)
		}
		return nil
	})
}

func performJournaledFileMove(move *RecoveryFileMove) error {
	_, sourceErr := os.Lstat(move.Source)
	_, destinationErr := os.Lstat(move.Destination)
	if errors.Is(sourceErr, os.ErrNotExist) && destinationErr == nil {
		move.Completed = true
		return nil
	}
	if sourceErr != nil || !errors.Is(destinationErr, os.ErrNotExist) {
		return errors.New("ambiguous media switch state; maintenance retained")
	}
	if err := os.Rename(move.Source, move.Destination); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(move.Source)); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(move.Destination)); err != nil {
		return err
	}
	move.Completed = true
	return nil
}

func (manager *Manager) RollbackRecovery(id, confirmation string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
		return nil, errors.New("use the dedicated stopped product switch rollback with its bound window confirmation")
	}
	if task.Journal == nil || task.State == "completed" || task.State == "rolled_back" || task.State == "resuming" || task.State == "releasing_rollback" {
		return nil, errors.New("no unfinished paired switch can be rolled back")
	}
	if confirmation != "ROLLBACK "+manager.Configuration.SiteID+"/"+id {
		return nil, errors.New("explicit paired rollback confirmation required")
	}
	if err := manager.startTaskOperation(task, "rolling_back_paired_switch", manager.rollbackPairedRecovery); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func (manager *Manager) rollbackPairedRecovery(ctx context.Context, task *Task) error {
	if taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) {
		return errors.New("stopped product switching requires its dedicated verified paired rollback")
	}
	if taskHasProductCatalogPairedRecoveryJournal(task) {
		return manager.rollbackProductCatalogPairedRehearsal(ctx, task, nil)
	}
	cfg := manager.Configuration
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer lease.Close()
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	journal := task.Journal
	var targetOID, oldOID, stageOID int64
	if err := readDatabaseOID(ctx, control, cfg.DatabaseName, &targetOID); err != nil {
		return err
	}
	if err := readDatabaseOID(ctx, control, journal.RollbackDatabase, &oldOID); err != nil {
		return err
	}
	if err := readDatabaseOID(ctx, control, task.StageDatabase, &stageOID); err != nil {
		return err
	}
	if targetOID != 0 && targetOID != journal.NewDatabaseOID && targetOID != journal.OldDatabaseOID {
		return errors.New("target database identity changed; refusing rollback")
	}
	if oldOID != 0 && oldOID != journal.OldDatabaseOID {
		return errors.New("rollback database identity changed")
	}
	if stageOID != 0 && stageOID != journal.NewDatabaseOID {
		return errors.New("isolated database identity changed")
	}
	if targetOID == journal.NewDatabaseOID {
		if stageOID != 0 {
			return errors.New("isolated rollback name is occupied")
		}
		if err := sealDatabase(ctx, control, cfg.DatabaseName); err != nil {
			return err
		}
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` RENAME TO `+pq.QuoteIdentifier(task.StageDatabase)); err != nil {
			return err
		}
		targetOID = 0
	}
	for index := len(journal.Moves) - 1; index >= 0; index-- {
		move := &journal.Moves[index]
		if !move.Started {
			continue
		}
		_, sourceErr := os.Lstat(move.Source)
		_, destinationErr := os.Lstat(move.Destination)
		if sourceErr == nil && errors.Is(destinationErr, os.ErrNotExist) {
			move.Completed = false
			move.Started = false
			if err := manager.saveTask(task); err != nil {
				return err
			}
			continue
		}
		if !errors.Is(sourceErr, os.ErrNotExist) || destinationErr != nil {
			return errors.New("ambiguous media rollback state")
		}
		reverse := RecoveryFileMove{Source: move.Destination, Destination: move.Source}
		if err := performJournaledFileMove(&reverse); err != nil {
			return err
		}
		move.Completed = false
		move.Started = false
		if err := manager.saveTask(task); err != nil {
			return err
		}
	}
	if journal.OldDatabaseOID != 0 && targetOID == 0 {
		if oldOID != journal.OldDatabaseOID {
			return errors.New("paired previous database is missing")
		}
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.RollbackDatabase)+` RENAME TO `+pq.QuoteIdentifier(cfg.DatabaseName)); err != nil {
			return err
		}
	}
	if journal.OldDatabaseOID != 0 {
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS true`); err != nil {
			return err
		}
	}
	fingerprint, _, err := cfg.targetFingerprint(ctx)
	if err != nil {
		return err
	}
	if fingerprint != task.PreviewTargetFingerprint {
		return errors.New("paired rollback fingerprint mismatch; service remains isolated")
	}
	task.State = "releasing_rollback"
	task.Phase = "previous_database_and_media_verified"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if err := manager.releaseMaintenance(task); err != nil {
		return err
	}
	task.State = "rolled_back"
	return manager.saveTask(task)
}

func (manager *Manager) CompleteReconciledRecovery(id, confirmation string, evidence ReconciliationEvidence) error {
	if !manager.operationMutex.TryLock() {
		return errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return err
	}
	cfg := manager.Configuration
	if task.State != "awaiting_reconciliation" || !taskHasSwitchedWholeApplicationRecoveryJournal(task) || confirmation != "RESUME "+cfg.SiteID+"/"+cfg.Environment+"/"+id {
		return errors.New("paired recovery is not ready or resume confirmation is incorrect")
	}
	for _, reference := range []string{evidence.PaymentAndRefundReference, evidence.ShippingReference, evidence.DurableEventsAndJobsReference, evidence.DedicatedRedisRebuildReference, evidence.AttachmentAccessAndIdentityReference, evidence.Operator} {
		if len(strings.TrimSpace(reference)) < 4 || len(reference) > 2000 {
			return errors.New("payment, shipping, jobs, Redis, access and operator reconciliation evidence is required")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer lease.Close()
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	var targetOID int64
	if err := readDatabaseOID(ctx, control, cfg.DatabaseName, &targetOID); err != nil {
		return err
	}
	if targetOID != task.Journal.NewDatabaseOID {
		return errors.New("switched database identity changed")
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return err
	}
	// Final validation must not reopen administrative logins on return from
	// inspection. Otherwise another SQL session could change the database after
	// its evidence fingerprint was accepted but before the fence is released.
	task.ReconciliationConnectionsEnabled = false
	if err := manager.saveTask(task); err != nil {
		return err
	}
	report, err := manager.inspectReconciliationUnderMaintenance(ctx, task)
	if err != nil {
		return err
	}
	if report.DatabaseFingerprint != evidence.ReconciledDatabaseFingerprint || report.MigrationVersion != cfg.ExpectedMigrationVersion || len(report.Gaps) > 0 {
		return errors.New("reconciliation report changed or has unresolved business/attachment gaps; review the latest report")
	}
	if err := validateTaskBoundExternalProjectionReconciliationEvidence(report.ExternalProjectionReconciliation, evidence.ExternalProjections); err != nil {
		return err
	}
	boundaryBytes, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
	if err != nil {
		return err
	}
	if err := writeDurableFile(filepath.Join(cfg.MaintenanceDirectory, "restored-data-identity-boundary.json"), boundaryBytes); err != nil {
		return err
	}
	// Reconciliation is an explicit operator attestation. It never triggers a
	// refund, shipment, event replay or a shared Redis FLUSHDB.
	task.Reconciliation = &evidence
	task.ExternalProjectionReconciliationContract = report.ExternalProjectionReconciliation
	task.State = "resuming"
	task.Phase = "reconciliation_recorded_ready_to_resume"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS true`); err != nil {
		return err
	}
	if err := manager.releaseMaintenance(task); err != nil {
		return err
	}
	task.State = "completed"
	return manager.saveTask(task)
}

func readDatabaseOID(ctx context.Context, db *sql.DB, name string, oid *int64) error {
	err := db.QueryRowContext(ctx, `SELECT oid::bigint FROM pg_database WHERE datname=$1`, name).Scan(oid)
	if errors.Is(err, sql.ErrNoRows) {
		*oid = 0
		return nil
	}
	return err
}
func sealDatabase(ctx context.Context, db *sql.DB, name string) error {
	if _, err := db.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(name)+` ALLOW_CONNECTIONS false`); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()`, name)
	return err
}

func (cfg Configuration) dropOwnedStageDatabase(ctx context.Context, name string) error {
	if !strings.HasPrefix(name, "local_backup_stage_") || !safeDatabaseName.MatchString(name) {
		return errors.New("refusing to drop a non-isolated database")
	}
	db, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `DROP DATABASE IF EXISTS `+pq.QuoteIdentifier(name)+` WITH (FORCE)`); err != nil {
		return err
	}
	return nil
}

func (manager *Manager) cleanupUnswitchedPreparedMedia(task *Task) error {
	if task.Journal != nil {
		return errors.New("prepared media belongs to an active paired switch")
	}
	for _, storageRoot := range manager.Configuration.wholeApplicationRecoveryStorageRoots(wholeApplicationPackageIncludesCompanyStatic(task.Manifest)) {
		root := storageRoot.Directory
		oldRoot := filepath.Join(root, ".local-backup-old-"+task.ID)
		entries, err := os.ReadDir(oldRoot)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if len(entries) > 0 {
			return errors.New("unjournaled previous media exists; maintenance retained for manual inspection")
		}
		if err := os.RemoveAll(filepath.Join(root, ".local-backup-stage-"+task.ID)); err != nil {
			return err
		}
		if err == nil {
			if err := os.Remove(oldRoot); err != nil {
				return err
			}
		}
	}
	return nil
}

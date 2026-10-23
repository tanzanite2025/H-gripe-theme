package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gofrs/flock"
	"github.com/lib/pq"

	"commerce-platform/internal/pkg/maintenancegate"
)

func (manager *Manager) acquireCustomerServicePairedSwitchExistingMaintenanceLease(ctx context.Context, task *Task) (*flock.Flock, error) {
	if err := manager.Configuration.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	data, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename), 8192)
	var owner struct {
		TaskID string `json:"task_id"`
	}
	if err != nil || json.Unmarshal(data, &owner) != nil || owner.TaskID != task.ID {
		return nil, errors.New("independent paired switch requires its existing maintenance fence; inspection required")
	}
	lease, err := maintenancegate.AcquireExclusiveRecoveryLease(ctx, manager.Configuration.MaintenanceDirectory)
	if err != nil {
		return nil, err
	}
	data, err = readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename), 8192)
	if err != nil || json.Unmarshal(data, &owner) != nil || owner.TaskID != task.ID {
		_ = lease.Close()
		return nil, errors.New("independent paired maintenance fence changed while acquiring the lease")
	}
	return lease, nil
}

func (manager *Manager) loadCustomerServicePairedSwitchSealedCandidateAndInventory(task *Task, journal *customerServiceIndependentPairedSwitchJournal) (*customerServicePersistentCandidateLifecycle, *customerServiceCandidateCompleteMediaUnionInventory, error) {
	record, err := manager.loadCustomerServicePersistentCandidateLifecycle(task.ID)
	if err != nil {
		return nil, nil, err
	}
	if record == nil || record.State != "sealed_not_switchable" || record.MediaUnion == nil || record.MediaUnion.State != "prepared_not_switchable" || record.Validation == nil || task.Journal != nil {
		return nil, nil, errors.New("independent paired rehearsal requires a sealed complete candidate and no whole-storefront journal")
	}
	inventory, err := manager.loadCustomerServicePairedSwitchBoundCompleteMediaInventory(record)
	if err != nil {
		return nil, nil, err
	}
	if journal != nil {
		configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(manager.Configuration)
		if err != nil {
			return nil, nil, err
		}
		taskDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(task)
		if err != nil {
			return nil, nil, err
		}
		if journal.AttemptID != record.AttemptID || journal.CandidateLifecycleSHA256 != record.SHA256 || journal.InventorySHA256 != inventory.SHA256 || journal.SelectionSHA256 != inventory.SelectionSHA256 || journal.PlanSHA256 != record.Request.PlanSHA256 || journal.CandidateDatabaseOID != record.Identity.DatabaseOID || journal.BaselineDatabaseFingerprint != record.Binding.TargetDatabaseFingerprint || journal.CandidateDatabaseFingerprint != record.Validation.CandidateFingerprint || journal.ConfigurationSHA256 != configurationDigest || journal.OriginalTaskSHA256 != taskDigest {
			return nil, nil, errors.New("independent paired switch candidate, task or configuration binding changed")
		}
		if err := reconstructCustomerServicePairedSwitchBaselineAndValidateMoves(journal, inventory); err != nil {
			return nil, nil, err
		}
	}
	return record, inventory, nil
}

func (manager *Manager) verifyCustomerServicePairedSwitchLatestSavedPlanUnderMaintenance(ctx context.Context, task *Task, record *customerServicePersistentCandidateLifecycle) error {
	review, err := manager.loadCustomerServiceRecoveryConflictReviewWhileOperationLocked(task.ID)
	if err != nil {
		return err
	}
	_, err = manager.inspectVerifiedPackageModuleDependenciesWithCustomerServiceSelectionObserverWhileOperationLocked(ctx, task.ID, "customer_service", func(ctx context.Context, report *ModuleRecoveryDependencyReport, source, target *moduleRecoveryRowDependencyGraph) error {
		plan := report.ConversationRows.ConversationRecoveryPlan
		if plan == nil || plan.SHA256 != record.Request.PlanSHA256 {
			return errors.New("independent paired switch saved plan changed")
		}
		_, err := validateCustomerServiceSavedDecisionsAgainstLatestPlan(plan, review, record.Request, time.Now().UTC())
		return err
	})
	return err
}

// Reading a sealed database briefly enables privileged maintenance connections.
// Its prior connection policy is restored even on cancellation; application
// privileges are never granted and the persistent process fence remains held.
func (manager *Manager) verifyCustomerServicePairedSwitchBoundDatabaseContents(ctx context.Context, control *sql.DB, name string, oid int64, owner, fingerprint string, binding *CustomerServiceRecoveryPlanSnapshotBinding) (resultErr error) {
	var actualOID int64
	var actualOwner string
	var allowed bool
	if err := control.QueryRowContext(ctx, `SELECT oid::bigint,pg_get_userbyid(datdba),datallowconn FROM pg_database WHERE datname=$1`, name).Scan(&actualOID, &actualOwner, &allowed); err != nil {
		return err
	}
	if actualOID != oid || actualOwner != owner {
		return errors.New("independent paired database OID, name or owner changed")
	}
	if err := assertNoUnfencedDatabaseSessions(ctx, control, name); err != nil {
		return err
	}
	if !allowed {
		defer func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			var currentOID int64
			if err := readDatabaseOID(cleanupCtx, control, name, &currentOID); err != nil || currentOID != oid {
				resultErr = errors.Join(resultErr, errors.New("cannot reseal a changed paired database identity"))
				return
			}
			resultErr = errors.Join(resultErr, sealDatabase(cleanupCtx, control, name))
		}()
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(name)+` ALLOW_CONNECTIONS true`); err != nil {
			return err
		}
	}
	database, err := manager.Configuration.connectDatabase(ctx, name)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(database.Close)
	transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(transaction.Rollback)
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
	if err != nil {
		return err
	}
	if observation.Fingerprint != fingerprint || observation.Migration != manager.Configuration.ExpectedMigrationVersion {
		return errors.New("independent paired database content or sequence fingerprint changed")
	}
	for _, contract := range []struct {
		value    any
		expected string
	}{{observation.Shapes, binding.TargetColumnConstraintFingerprint}, {observation.Indexes, binding.TargetIndexFingerprint}, {observation.WriteSideEffects, binding.TargetWriteSideEffectFingerprint}} {
		digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(contract.value)
		if err != nil || digest != contract.expected {
			return errors.New("independent paired database schema or write contract changed")
		}
	}
	return assertNoUnfencedDatabaseSessions(ctx, control, name)
}

func (manager *Manager) inspectCustomerServicePairedSwitchDatabaseTopology(ctx context.Context, control *sql.DB, journal *customerServiceIndependentPairedSwitchJournal) (string, string, error) {
	positions := map[string]int64{}
	owners := map[string]string{}
	for _, name := range []string{journal.TargetDatabase, journal.CandidateDatabase, journal.RollbackDatabase} {
		var oid int64
		if err := readDatabaseOID(ctx, control, name, &oid); err != nil {
			return "", "", err
		}
		positions[name] = oid
		if oid != 0 {
			var owner string
			if err := control.QueryRowContext(ctx, `SELECT pg_get_userbyid(datdba) FROM pg_database WHERE oid=$1`, oid).Scan(&owner); err != nil {
				return "", "", err
			}
			owners[name] = owner
			if err := assertNoUnfencedDatabaseSessions(ctx, control, name); err != nil {
				return "", "", err
			}
		}
	}
	originalName, candidateName := "", ""
	switch {
	case positions[journal.TargetDatabase] == journal.OriginalDatabaseOID && positions[journal.CandidateDatabase] == journal.CandidateDatabaseOID && positions[journal.RollbackDatabase] == 0:
		originalName, candidateName = journal.TargetDatabase, journal.CandidateDatabase
	case positions[journal.TargetDatabase] == 0 && positions[journal.CandidateDatabase] == journal.CandidateDatabaseOID && positions[journal.RollbackDatabase] == journal.OriginalDatabaseOID:
		originalName, candidateName = journal.RollbackDatabase, journal.CandidateDatabase
	case positions[journal.TargetDatabase] == journal.CandidateDatabaseOID && positions[journal.CandidateDatabase] == 0 && positions[journal.RollbackDatabase] == journal.OriginalDatabaseOID:
		originalName, candidateName = journal.RollbackDatabase, journal.TargetDatabase
	default:
		return "", "", errors.New("independent paired database topology or OID changed; inspection required")
	}
	if owners[originalName] != journal.OriginalDatabaseOwner || owners[candidateName] != customerServiceDisposableCandidateOwner {
		return "", "", errors.New("independent paired database owner changed")
	}
	var candidateAllowsConnections bool
	if err := control.QueryRowContext(ctx, `SELECT datallowconn FROM pg_database WHERE oid=$1`, journal.CandidateDatabaseOID).Scan(&candidateAllowsConnections); err != nil {
		return "", "", err
	}
	if candidateAllowsConnections {
		return "", "", errors.New("independent paired candidate connection policy changed")
	}
	if err := ensureCustomerServiceDisposableCandidateOwnerAndApplicationIsolation(ctx, control, manager.Configuration.ApplicationDatabaseRole); err != nil {
		return "", "", err
	}
	var applicationCanAccess bool
	if err := control.QueryRowContext(ctx, `SELECT has_database_privilege($1,$2,'CONNECT') OR has_database_privilege($1,$2,'CREATE') OR has_database_privilege($1,$2,'TEMP')`, manager.Configuration.ApplicationDatabaseRole, candidateName).Scan(&applicationCanAccess); err != nil {
		return "", "", err
	}
	if applicationCanAccess {
		return "", "", errors.New("application role gained access to the isolated candidate")
	}
	return originalName, candidateName, nil
}

func (manager *Manager) prepareCustomerServiceIndependentPairedSwitchRehearsalJournal(ctx context.Context, id, expectedCandidateLifecycleSHA256 string) (*customerServiceIndependentPairedSwitchJournal, error) {
	if !manager.operationMutex.TryLock() {
		return nil, errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	if err := manager.requireCustomerServiceIndependentPairedSwitchRolledBack(id); err != nil {
		return nil, err
	}
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	record, inventory, err := manager.loadCustomerServicePairedSwitchSealedCandidateAndInventory(task, nil)
	if err != nil {
		return nil, err
	}
	if record.SHA256 != expectedCandidateLifecycleSHA256 {
		return nil, errors.New("sealed candidate changed before journal preparation")
	}
	if err := manager.verifyCustomerServiceCandidateBoundCleanupDispositionAudit(record); err != nil {
		return nil, err
	}
	lease, err := manager.acquireCustomerServicePairedSwitchExistingMaintenanceLease(ctx, task)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	if err := manager.verifyCustomerServicePairedSwitchLatestSavedPlanUnderMaintenance(ctx, task, record); err != nil {
		return nil, err
	}
	if err := manager.verifyCustomerServiceCompleteCandidateMediaUnion(ctx, record, inventory); err != nil {
		return nil, err
	}
	if err := manager.validateCustomerServicePairedSwitchOwnedMediaVolumes(record); err != nil {
		return nil, err
	}
	baseline, err := manager.Configuration.inventoryCustomerServiceCompleteCurrentMedia(ctx)
	if err != nil {
		return nil, err
	}
	moves, err := buildCustomerServicePairedSwitchExpectedMediaMoves(baseline, inventory)
	if err != nil {
		return nil, err
	}
	journal := &customerServiceIndependentPairedSwitchJournal{Version: 1, State: "prepared_isolated_rehearsal", TaskID: id, AttemptID: record.AttemptID, CreatedAt: time.Now().UTC(), CandidateLifecycleSHA256: record.SHA256, InventorySHA256: inventory.SHA256, PlanSHA256: record.Request.PlanSHA256, SelectionSHA256: record.Validation.SelectionSHA256, TargetDatabase: manager.Configuration.DatabaseName, CandidateDatabase: record.Identity.DatabaseName, RollbackDatabase: "local_backup_cs_old_" + record.AttemptID[:16], CandidateDatabaseOID: record.Identity.DatabaseOID, BaselineDatabaseFingerprint: record.Binding.TargetDatabaseFingerprint, CandidateDatabaseFingerprint: record.Validation.CandidateFingerprint, BaselineDirectories: baseline.Directories, Moves: moves}
	journal.DatabasePhase = "not_started"
	journal.ConfigurationSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(manager.Configuration)
	if err != nil {
		return nil, err
	}
	journal.OriginalTaskSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(task)
	if err != nil {
		return nil, err
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	if err := control.QueryRowContext(ctx, `SELECT oid::bigint,pg_get_userbyid(datdba) FROM pg_database WHERE datname=$1`, journal.TargetDatabase).Scan(&journal.OriginalDatabaseOID, &journal.OriginalDatabaseOwner); err != nil {
		return nil, err
	}
	if _, _, err := manager.inspectCustomerServicePairedSwitchDatabaseTopology(ctx, control, journal); err != nil {
		return nil, err
	}
	if err := manager.verifyCustomerServicePairedSwitchBoundDatabaseContents(ctx, control, journal.TargetDatabase, journal.OriginalDatabaseOID, journal.OriginalDatabaseOwner, journal.BaselineDatabaseFingerprint, record.Binding); err != nil {
		return nil, err
	}
	if err := manager.verifyCustomerServicePairedSwitchBoundDatabaseContents(ctx, control, journal.CandidateDatabase, journal.CandidateDatabaseOID, customerServiceDisposableCandidateOwner, journal.CandidateDatabaseFingerprint, record.Binding); err != nil {
		return nil, err
	}
	if err := reconstructCustomerServicePairedSwitchBaselineAndValidateMoves(journal, inventory); err != nil {
		return nil, err
	}
	layout, err := manager.inspectCustomerServicePairedSwitchCompletePhysicalMediaLayout(ctx, record)
	if err != nil {
		return nil, err
	}
	if err := validateCustomerServicePairedSwitchPhysicalLayoutAgainstDurableProgress(journal, layout); err != nil {
		return nil, err
	}
	previous, err := manager.loadCustomerServiceIndependentPairedSwitchJournal(id)
	if err != nil {
		return nil, err
	}
	if previous != nil {
		directory, _ := manager.taskDirectory(id)
		if err := writeDurableJSON(filepath.Join(directory, "customer-service-rolled-back-paired-switch-attempt-"+previous.AttemptID+".json"), previous); err != nil {
			return nil, err
		}
	}
	return journal, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal)
}

func (manager *Manager) retainCustomerServicePairedSwitchFailureForInspection(journal *customerServiceIndependentPairedSwitchJournal, err error) error {
	journal.State, journal.ErrorCode = "inspection_required", "independent_paired_switch_requires_verified_rollback_or_inspection"
	return errors.Join(err, manager.saveCustomerServiceIndependentPairedSwitchJournal(journal))
}

// Internal rehearsal only. Even after all renames, the candidate keeps its
// NOLOGIN owner, revoked application privileges, disabled logins and fence.
func (manager *Manager) executeCustomerServiceIndependentPairedSwitchRehearsal(ctx context.Context, id, expectedJournalSHA256 string) (resultErr error) {
	if !manager.operationMutex.TryLock() {
		return errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	journal, err := manager.loadCustomerServiceIndependentPairedSwitchJournal(id)
	if err != nil {
		return err
	}
	if journal == nil || journal.State != "prepared_isolated_rehearsal" || journal.SHA256 != expectedJournalSHA256 {
		return errors.New("a fresh bound isolated rehearsal journal is required")
	}
	defer func() {
		if resultErr != nil {
			resultErr = manager.retainCustomerServicePairedSwitchFailureForInspection(journal, resultErr)
		}
	}()
	task, err := manager.GetTask(id)
	if err != nil {
		return err
	}
	record, inventory, err := manager.loadCustomerServicePairedSwitchSealedCandidateAndInventory(task, journal)
	if err != nil {
		return err
	}
	if err := manager.verifyCustomerServiceCandidateBoundCleanupDispositionAudit(record); err != nil {
		return err
	}
	lease, err := manager.acquireCustomerServicePairedSwitchExistingMaintenanceLease(ctx, task)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	if err := manager.verifyCustomerServicePairedSwitchLatestSavedPlanUnderMaintenance(ctx, task, record); err != nil {
		return err
	}
	if err := manager.verifyCustomerServiceCompleteCandidateMediaUnion(ctx, record, inventory); err != nil {
		return err
	}
	layout, err := manager.inspectCustomerServicePairedSwitchCompletePhysicalMediaLayout(ctx, record)
	if err != nil {
		return err
	}
	if err := validateCustomerServicePairedSwitchPhysicalLayoutAgainstDurableProgress(journal, layout); err != nil {
		return err
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	originalName, candidateName, err := manager.inspectCustomerServicePairedSwitchDatabaseTopology(ctx, control, journal)
	if err != nil {
		return err
	}
	if originalName != journal.TargetDatabase || candidateName != journal.CandidateDatabase {
		return errors.New("prepared independent journal already has renamed databases")
	}
	for _, database := range []struct {
		name               string
		oid                int64
		owner, fingerprint string
	}{{originalName, journal.OriginalDatabaseOID, journal.OriginalDatabaseOwner, journal.BaselineDatabaseFingerprint}, {candidateName, journal.CandidateDatabaseOID, customerServiceDisposableCandidateOwner, journal.CandidateDatabaseFingerprint}} {
		if err := manager.verifyCustomerServicePairedSwitchBoundDatabaseContents(ctx, control, database.name, database.oid, database.owner, database.fingerprint, record.Binding); err != nil {
			return err
		}
	}
	if err := manager.Configuration.checkPostgreSQLAvailableBytes(minimumPostgreSQLFreeReserveBytes); err != nil {
		return err
	}
	journal.State, journal.DatabasePhase = "switching_isolated_rehearsal", "retain_original_database_intent"
	if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
		return err
	}
	for _, volume := range record.MediaUnion.Volumes {
		directory, _ := customerServiceCandidateMediaVolumeDirectory(record, volume)
		previous := filepath.Join(directory, "previous")
		if err := os.Mkdir(previous, 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		if err := syncDirectory(directory); err != nil {
			return err
		}
	}
	if err := sealDatabase(ctx, control, journal.TargetDatabase); err != nil {
		return err
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.TargetDatabase)+` RENAME TO `+pq.QuoteIdentifier(journal.RollbackDatabase)); err != nil {
		return err
	}
	journal.DatabasePhase = "original_database_retained"
	if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
		return err
	}
	for index := range journal.Moves {
		journal.Moves[index].Started = true
		if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
			return err
		}
		if err := manager.performCustomerServicePairedSwitchBoundMediaRename(ctx, record, journal.Moves[index], false); err != nil {
			return err
		}
		journal.Moves[index].Completed = true
		if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
			return err
		}
	}
	// Persist a database rename intent even when the media set is empty.
	journal.DatabasePhase = "install_candidate_database_intent"
	if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
		return err
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.CandidateDatabase)+` RENAME TO `+pq.QuoteIdentifier(journal.TargetDatabase)); err != nil {
		return err
	}
	if _, _, err := manager.inspectCustomerServicePairedSwitchDatabaseTopology(ctx, control, journal); err != nil {
		return err
	}
	layout, err = manager.inspectCustomerServicePairedSwitchCompletePhysicalMediaLayout(ctx, record)
	if err != nil {
		return err
	}
	if err := validateCustomerServicePairedSwitchPhysicalLayoutAgainstDurableProgress(journal, layout); err != nil {
		return err
	}
	for _, database := range []struct {
		name               string
		oid                int64
		owner, fingerprint string
	}{
		{journal.RollbackDatabase, journal.OriginalDatabaseOID, journal.OriginalDatabaseOwner, journal.BaselineDatabaseFingerprint},
		{journal.TargetDatabase, journal.CandidateDatabaseOID, customerServiceDisposableCandidateOwner, journal.CandidateDatabaseFingerprint},
	} {
		if err := manager.verifyCustomerServicePairedSwitchBoundDatabaseContents(ctx, control, database.name, database.oid, database.owner, database.fingerprint, record.Binding); err != nil {
			return err
		}
	}
	journal.State, journal.DatabasePhase = "switched_isolated_rollback_required", "candidate_database_installed"
	return manager.saveCustomerServiceIndependentPairedSwitchJournal(journal)
}

func (manager *Manager) rollbackCustomerServiceIndependentPairedSwitchWhileOperationLocked(ctx context.Context, task *Task) (resultErr error) {
	journal, err := manager.loadCustomerServiceIndependentPairedSwitchJournal(task.ID)
	if err != nil || journal == nil || journal.State == "rolled_back_isolated" {
		return err
	}
	defer func() {
		if resultErr != nil {
			resultErr = manager.retainCustomerServicePairedSwitchFailureForInspection(journal, resultErr)
		}
	}()
	record, inventory, err := manager.loadCustomerServicePairedSwitchSealedCandidateAndInventory(task, journal)
	if err != nil {
		return err
	}
	lease, err := manager.acquireCustomerServicePairedSwitchExistingMaintenanceLease(ctx, task)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	layout, err := manager.inspectCustomerServicePairedSwitchCompletePhysicalMediaLayout(ctx, record)
	if err != nil {
		return err
	}
	if err := validateCustomerServicePairedSwitchPhysicalLayoutAgainstDurableProgress(journal, layout); err != nil {
		return err
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	originalName, candidateName, err := manager.inspectCustomerServicePairedSwitchDatabaseTopology(ctx, control, journal)
	if err != nil {
		return err
	}
	for _, database := range []struct {
		name               string
		oid                int64
		owner, fingerprint string
	}{{originalName, journal.OriginalDatabaseOID, journal.OriginalDatabaseOwner, journal.BaselineDatabaseFingerprint}, {candidateName, journal.CandidateDatabaseOID, customerServiceDisposableCandidateOwner, journal.CandidateDatabaseFingerprint}} {
		if err := manager.verifyCustomerServicePairedSwitchBoundDatabaseContents(ctx, control, database.name, database.oid, database.owner, database.fingerprint, record.Binding); err != nil {
			return err
		}
	}
	journal.State, journal.DatabasePhase = "rolling_back_isolated_rehearsal", "return_candidate_database_intent"
	if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
		return err
	}
	if candidateName == journal.TargetDatabase {
		if err := sealDatabase(ctx, control, journal.TargetDatabase); err != nil {
			return err
		}
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.TargetDatabase)+` RENAME TO `+pq.QuoteIdentifier(journal.CandidateDatabase)); err != nil {
			return err
		}
	}
	journal.DatabasePhase = "rolling_back_media"
	if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
		return err
	}
	for index := len(journal.Moves) - 1; index >= 0; index-- {
		move := &journal.Moves[index]
		if !move.Started || move.RolledBack {
			continue
		}
		move.RollbackStarted = true
		if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
			return err
		}
		if err := manager.performCustomerServicePairedSwitchBoundMediaRename(ctx, record, *move, true); err != nil {
			return err
		}
		move.RolledBack = true
		if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
			return err
		}
	}
	if originalName == journal.RollbackDatabase {
		journal.DatabasePhase = "restore_original_database_intent"
		if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
			return err
		}
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.RollbackDatabase)+` RENAME TO `+pq.QuoteIdentifier(journal.TargetDatabase)); err != nil {
			return err
		}
	}
	if err := manager.verifyCustomerServicePairedSwitchBoundDatabaseContents(ctx, control, journal.TargetDatabase, journal.OriginalDatabaseOID, journal.OriginalDatabaseOwner, journal.BaselineDatabaseFingerprint, record.Binding); err != nil {
		return err
	}
	if err := manager.verifyCustomerServiceCompleteCandidateMediaUnion(ctx, record, inventory); err != nil {
		return err
	}
	journal.DatabasePhase = "original_database_and_media_verified"
	if err := manager.saveCustomerServiceIndependentPairedSwitchJournal(journal); err != nil {
		return err
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(journal.TargetDatabase)+` ALLOW_CONNECTIONS true`); err != nil {
		return err
	}
	journal.State, journal.ErrorCode = "rolled_back_isolated", ""
	return manager.saveCustomerServiceIndependentPairedSwitchJournal(journal)
}

func (manager *Manager) rollbackCustomerServiceIndependentPairedSwitch(ctx context.Context, id string) error {
	if !manager.operationMutex.TryLock() {
		return errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return err
	}
	return manager.rollbackCustomerServiceIndependentPairedSwitchWhileOperationLocked(ctx, task)
}

func (manager *Manager) recoverInterruptedCustomerServiceIndependentPairedSwitches(tasks []*Task) {
	for _, task := range tasks {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		err := manager.rollbackCustomerServiceIndependentPairedSwitchWhileOperationLocked(ctx, task)
		cancel()
		if err != nil {
			_, _ = os.Stderr.WriteString("independent customer-service paired recovery requires inspection; candidate disposal and maintenance release remain blocked\n")
		}
	}
}

// Cleanup independently checks that the recorded originals are live again.
// A final journal flag alone never authorizes removing a previous media tree.
func (manager *Manager) verifyCustomerServicePairedRollbackBeforeCandidateDisposal(ctx context.Context, record *customerServicePersistentCandidateLifecycle) error {
	journal, err := manager.loadCustomerServiceIndependentPairedSwitchJournal(record.TaskID)
	if err != nil || journal == nil || journal.AttemptID != record.AttemptID {
		return err
	}
	if journal.State != "rolled_back_isolated" {
		return errors.New("independent paired rollback must be verified before candidate disposal")
	}
	configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(manager.Configuration)
	if err != nil || configurationDigest != journal.ConfigurationSHA256 || record.Binding == nil || record.MediaUnion == nil {
		return errors.New("paired disposal configuration or baseline binding changed")
	}
	task, err := manager.GetTask(record.TaskID)
	if err != nil {
		return err
	}
	lease, err := manager.acquireCustomerServicePairedSwitchExistingMaintenanceLease(ctx, task)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var targetOID, previousOID, candidateOID int64
	for _, position := range []struct {
		name string
		oid  *int64
	}{{journal.TargetDatabase, &targetOID}, {journal.RollbackDatabase, &previousOID}, {journal.CandidateDatabase, &candidateOID}} {
		if err := readDatabaseOID(ctx, control, position.name, position.oid); err != nil {
			return err
		}
	}
	if targetOID != journal.OriginalDatabaseOID || previousOID != 0 || (candidateOID != 0 && candidateOID != journal.CandidateDatabaseOID) {
		return errors.New("independent paired originals are not restored; candidate disposal refused")
	}
	if err := manager.verifyCustomerServicePairedSwitchBoundDatabaseContents(ctx, control, journal.TargetDatabase, journal.OriginalDatabaseOID, journal.OriginalDatabaseOwner, journal.BaselineDatabaseFingerprint, record.Binding); err != nil {
		return err
	}
	current, err := manager.Configuration.inventoryCustomerServiceCompleteCurrentMedia(ctx)
	if err != nil {
		return err
	}
	if fingerprintCustomerServiceCandidateMediaInventory(current) != record.MediaUnion.BaselineMediaFingerprint || !reflect.DeepEqual(current.Directories, journal.BaselineDirectories) {
		return errors.New("independent paired original media are not restored; candidate disposal refused")
	}
	return nil
}

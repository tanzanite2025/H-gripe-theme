package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/lib/pq"
)

const customerServiceDisposableCandidateOwner = "local_backup_customer_service_candidate_owner"

type customerServiceDisposableCandidateDatabaseIdentity struct {
	Version      int    `json:"version"`
	State        string `json:"state"`
	DatabaseName string `json:"database_name"`
	DatabaseOID  int64  `json:"database_oid"`
	PlanSHA256   string `json:"plan_sha256"`
}

type customerServiceDisposableCandidateRehearsalResult struct {
	Version                         int                                          `json:"version"`
	State                           string                                       `json:"state"`
	ExecutionAvailable              bool                                         `json:"execution_available"`
	PlanSHA256                      string                                       `json:"plan_sha256"`
	SelectionSHA256                 string                                       `json:"selection_sha256"`
	CleanupDispositionEntriesSHA256 string                                       `json:"cleanup_disposition_entries_sha256,omitempty"`
	BaselineFingerprint             string                                       `json:"baseline_fingerprint"`
	CandidateFingerprint            string                                       `json:"candidate_fingerprint"`
	SelectedFullRows                int                                          `json:"selected_full_rows"`
	VerifiedRetainedTables          int                                          `json:"verified_retained_tables"`
	VerifiedSequenceStates          int                                          `json:"verified_sequence_states"`
	SequenceAdjustments             []customerServiceCandidateSequenceAdjustment `json:"sequence_adjustments"`
	DependencyBlockers              []ModuleRecoveryDependencyBlocker            `json:"dependency_blockers"`
	CandidateDeleted                bool                                         `json:"candidate_deleted"`
}

// The construction workspace records allocation intent before CREATE DATABASE.
// Its caller owns disposal; successful construction alone never permits a switch.
type customerServiceRecordedCandidateConstructionWorkspace struct {
	Directory                        string
	Identity                         customerServiceDisposableCandidateDatabaseIdentity
	PersistIdentity                  func(customerServiceDisposableCandidateDatabaseIdentity) error
	PrepareCompleteMediaForSelection func(context.Context, *CustomerServiceEffectiveSelectionInspection) error
}

// Internal disposable rehearsal; it neither fences writers nor retains a candidate.
func (cfg Configuration) rehearseCustomerServiceSelectedRowsInDisposableTargetClone(ctx context.Context, source, target *moduleRecoveryRowDependencyGraph, plan *CustomerServiceConversationRecoveryPlan, review *CustomerServiceRecoveryConflictReview, request CustomerServiceEffectiveSelectionInspectionRequest, manifest *Manifest, verifiedDirectory string) (result *customerServiceDisposableCandidateRehearsalResult, resultErr error) {
	directory, err := os.MkdirTemp(cfg.Workspace, "customer-service-candidate-rehearsal-")
	if err != nil {
		return nil, err
	}
	workspace := &customerServiceRecordedCandidateConstructionWorkspace{Directory: directory,
		Identity: customerServiceDisposableCandidateDatabaseIdentity{Version: 1, State: "allocation_not_attempted", DatabaseName: "local_backup_candidate_" + newTaskID()[:16], PlanSHA256: request.PlanSHA256}}
	workspace.PersistIdentity = func(identity customerServiceDisposableCandidateDatabaseIdentity) error {
		return writeDurableJSON(filepath.Join(directory, "customer-service-disposable-candidate-database-identity.json"), identity)
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := cfg.discardCustomerServiceRecordedCandidateConstructionWorkspace(cleanupCtx, workspace); err != nil {
			resultErr = errors.Join(resultErr, err)
			return
		}
		if result != nil {
			result.CandidateDeleted = true
		}
	}()
	if err := workspace.PersistIdentity(workspace.Identity); err != nil {
		return nil, err
	}
	return cfg.constructCustomerServiceSelectedFullRowsInRecordedTargetClone(ctx, source, target, plan, review, request, manifest, verifiedDirectory, workspace)
}

// Both callers keep the same two snapshots open and use the same full-row,
// retention and sequence validation. Persistence does not bypass these checks.
func (cfg Configuration) constructCustomerServiceSelectedFullRowsInRecordedTargetClone(ctx context.Context, source, target *moduleRecoveryRowDependencyGraph, plan *CustomerServiceConversationRecoveryPlan, review *CustomerServiceRecoveryConflictReview, request CustomerServiceEffectiveSelectionInspectionRequest, manifest *Manifest, verifiedDirectory string, workspace *customerServiceRecordedCandidateConstructionWorkspace) (*customerServiceDisposableCandidateRehearsalResult, error) {
	if workspace == nil || workspace.PersistIdentity == nil || workspace.Identity.Version != 1 || workspace.Identity.State != "allocation_not_attempted" || workspace.Identity.DatabaseOID != 0 || !strings.HasPrefix(workspace.Identity.DatabaseName, "local_backup_candidate_") || !safeDatabaseName.MatchString(workspace.Identity.DatabaseName) || workspace.Identity.PlanSHA256 != request.PlanSHA256 {
		return nil, errors.New("candidate construction requires its durable unallocated workspace")
	}
	if source == nil || target == nil || source.Transaction == nil || target.Transaction == nil || source.Transaction == target.Transaction || manifest == nil {
		return nil, errors.New("candidate rehearsal requires independent existing source and target snapshots")
	}
	for _, transaction := range []*sql.Tx{source.Transaction, target.Transaction} {
		var readOnly, isolation string
		if err := transaction.QueryRowContext(ctx, `SELECT current_setting('transaction_read_only'),current_setting('transaction_isolation')`).Scan(&readOnly, &isolation); err != nil {
			return nil, err
		}
		if readOnly != "on" || isolation != "repeatable read" {
			return nil, errors.New("candidate rehearsal requires read-only repeatable-read observations")
		}
	}
	var sourceName, targetName string
	if err := source.Transaction.QueryRowContext(ctx, `SELECT current_database()`).Scan(&sourceName); err != nil {
		return nil, err
	}
	if err := target.Transaction.QueryRowContext(ctx, `SELECT current_database()`).Scan(&targetName); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(sourceName, "local_backup_stage_") || !safeDatabaseName.MatchString(sourceName) || sourceName == targetName || targetName != cfg.DatabaseName {
		return nil, errors.New("candidate rehearsal source or target database identity is invalid")
	}
	if plan == nil || !safeTaskID.MatchString(plan.Binding.TaskID) || sourceName != "local_backup_stage_"+plan.Binding.TaskID[:16] || manifest.FormatVersion != PackageVersion || manifest.Scope != "whole_application" || manifest.SiteID != cfg.SiteID || manifest.Environment != cfg.Environment || manifest.ApplicationVersion != cfg.ApplicationVersion || manifest.MigrationVersion != cfg.ExpectedMigrationVersion || manifest.DatabaseFingerprint != source.Observation.Fingerprint || len(manifest.Gaps) != 0 {
		return nil, errors.New("candidate rehearsal requires a compatible complete package and its bound source")
	}
	mediaFingerprint, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return nil, err
	}
	expectedPlan := *plan
	expectedTask := &Task{ID: plan.Binding.TaskID, Manifest: manifest, CiphertextSHA256: plan.Binding.CiphertextSHA256, StageDatabase: sourceName, PreviewTargetFingerprint: plan.Binding.OriginalWholeStorefrontPreviewFingerprint, ExpiresAt: plan.Binding.WorkCopyExpiresAt}
	expectedReport := &ModuleRecoveryDependencyReport{SourceDatabaseFingerprint: source.Observation.Fingerprint, TargetDatabaseFingerprint: target.Observation.Fingerprint, Blockers: plan.DependencyBlockers}
	if err := cfg.bindCustomerServiceRecoveryPlanToObservedSnapshots(&expectedPlan, expectedTask, expectedReport, mediaFingerprint, source.Observation, target.Observation); err != nil {
		return nil, err
	}
	if err := verifyCustomerServiceRecoveryPlanDraftDigestExpirationAndSnapshotBinding(plan, expectedPlan.Binding, time.Now().UTC()); err != nil {
		return nil, err
	}
	selection, err := cfg.inspectCustomerServiceEffectiveSelectedRows(ctx, source, target, plan, review, request, manifest, verifiedDirectory)
	if err != nil {
		return nil, err
	}
	if !selection.SelectionDependenciesSatisfied {
		return nil, errors.New("candidate rehearsal requires complete selections without original or effective row repair blockers")
	}
	if plan.Binding.SourceDatabaseFingerprint != source.Observation.Fingerprint || plan.Binding.TargetDatabaseFingerprint != target.Observation.Fingerprint || source.Observation.Migration != cfg.ExpectedMigrationVersion || target.Observation.Migration != cfg.ExpectedMigrationVersion {
		return nil, errors.New("candidate rehearsal snapshot binding changed")
	}
	for _, pair := range [][2]any{{source.Observation.Shapes, target.Observation.Shapes}, {source.Observation.Indexes, target.Observation.Indexes}, {source.Observation.WriteSideEffects, target.Observation.WriteSideEffects}} {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			return nil, errors.New("candidate rehearsal requires matching source and target database contracts")
		}
	}
	if err := assertCustomerServiceCandidateTablesHaveSupportedWriteContracts(ctx, target.Transaction, target.Observation); err != nil {
		return nil, err
	}
	baselineDatabase, err := cfg.connectDatabase(ctx, targetName)
	if err != nil {
		return nil, err
	}
	baselineSequenceGaps := &Manifest{}
	err = findOwnedUniqueSequenceCollisionOrExhaustionGaps(ctx, baselineDatabase, baselineSequenceGaps)
	_ = baselineDatabase.Close()
	if err != nil {
		return nil, err
	}
	if len(baselineSequenceGaps.Gaps) != 0 {
		return nil, errors.New("current target baseline has unsafe owned unique sequences; candidate rehearsal cannot repair them")
	}
	rows, selectedKeys, err := readCustomerServiceCandidateSelectedFullRows(ctx, source, selection)
	if err != nil {
		return nil, err
	}
	baselineSequences, err := readCustomerServiceCandidateSequenceStates(ctx, target.Transaction)
	if err != nil {
		return nil, err
	}
	var databaseBytes int64
	if err := target.Transaction.QueryRowContext(ctx, `SELECT pg_database_size(current_database())`).Scan(&databaseBytes); err != nil {
		return nil, err
	}
	// Existing current/stage databases are already allocated. This additional
	// free-space estimate covers a target dump, clone/indexes and WAL reserve;
	// it is not the budget for media union, rollback or a production switch.
	if databaseBytes <= 0 || databaseBytes > (math.MaxInt64-(256<<20))/3 {
		return nil, errors.New("candidate baseline size is invalid")
	}
	if workspace.PrepareCompleteMediaForSelection != nil {
		if err := workspace.PrepareCompleteMediaForSelection(ctx, selection); err != nil {
			return nil, err
		}
	}
	dumpByteLimit := 2*databaseBytes + (16 << 20)
	if err := cfg.checkWorkspaceCapacity(dumpByteLimit); err != nil {
		return nil, err
	}
	if err := cfg.checkPostgreSQLAvailableBytes(uint64(databaseBytes)*3 + minimumPostgreSQLFreeReserveBytes); err != nil {
		return nil, err
	}
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	if err := ensureCustomerServiceDisposableCandidateOwnerAndApplicationIsolation(ctx, control, cfg.ApplicationDatabaseRole); err != nil {
		return nil, err
	}
	identity := &workspace.Identity
	var snapshot string
	if err := target.Transaction.QueryRowContext(ctx, `SELECT pg_export_snapshot()`).Scan(&snapshot); err != nil {
		return nil, err
	}
	dumpPath := filepath.Join(workspace.Directory, "current-target-snapshot.dump")
	if err := cfg.dumpCustomerServiceTargetSnapshotWithAnExplicitByteLimit(ctx, targetName, snapshot, dumpPath, dumpByteLimit); err != nil {
		return nil, err
	}
	if err := cfg.checkWorkspaceCapacity(16 << 20); err != nil {
		return nil, err
	}
	// Once allocation is attempted, an ambiguous/cancelled response must retain
	// its durable intent. Without an OID cleanup refuses to guess by name.
	identity.State = "allocation_intent"
	if err := workspace.PersistIdentity(*identity); err != nil {
		return nil, err
	}
	if _, err := control.ExecContext(ctx, `CREATE DATABASE `+pq.QuoteIdentifier(identity.DatabaseName)+` WITH TEMPLATE template0 OWNER `+pq.QuoteIdentifier(customerServiceDisposableCandidateOwner)+` ALLOW_CONNECTIONS false`); err != nil {
		return nil, errors.New("cannot allocate disposable candidate database")
	}
	if err := control.QueryRowContext(ctx, `SELECT oid::bigint FROM pg_database WHERE datname=$1`, identity.DatabaseName).Scan(&identity.DatabaseOID); err != nil {
		return nil, err
	}
	identity.State = "allocated_not_switchable"
	if err := workspace.PersistIdentity(*identity); err != nil {
		return nil, err
	}
	if _, err := control.ExecContext(ctx, `REVOKE ALL ON DATABASE `+pq.QuoteIdentifier(identity.DatabaseName)+` FROM PUBLIC,`+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)); err != nil {
		return nil, err
	}
	var applicationCanAccessCandidate bool
	if err := control.QueryRowContext(ctx, `SELECT has_database_privilege($1,$2,'CONNECT') OR has_database_privilege($1,$2,'TEMP') OR has_database_privilege($1,$2,'CREATE')`, cfg.ApplicationDatabaseRole, identity.DatabaseName).Scan(&applicationCanAccessCandidate); err != nil {
		return nil, err
	}
	if applicationCanAccessCandidate {
		return nil, errors.New("application role still has access to the disposable candidate")
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(identity.DatabaseName)+` ALLOW_CONNECTIONS true`); err != nil {
		return nil, err
	}
	if err := cfg.runPostgreSQLTool(ctx, "pg_restore", identity.DatabaseName, "--exit-on-error", "--single-transaction", "--no-owner", "--no-privileges", "--no-tablespaces", "--role="+customerServiceDisposableCandidateOwner, dumpPath); err != nil {
		return nil, err
	}
	candidate, err := cfg.connectDatabase(ctx, identity.DatabaseName)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(candidate.Close)
	transaction, err := candidate.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(transaction.Rollback)
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, transaction)
	if err != nil {
		return nil, err
	}
	if observation.Fingerprint != target.Observation.Fingerprint || !reflect.DeepEqual(observation.Shapes, target.Observation.Shapes) || !reflect.DeepEqual(observation.Indexes, target.Observation.Indexes) || !reflect.DeepEqual(observation.WriteSideEffects, target.Observation.WriteSideEffects) {
		return nil, errors.New("disposable clone differs from the observed current target baseline")
	}
	// Writes execute as the unprivileged database owner with the normal trigger,
	// constraint and RLS behavior. No replica role or trigger disabling is used.
	if _, err := transaction.ExecContext(ctx, `SET LOCAL ROLE `+pq.QuoteIdentifier(customerServiceDisposableCandidateOwner)); err != nil {
		return nil, err
	}
	if err := writeCustomerServiceSelectedFullRowsAndValidateRetainedTables(ctx, transaction, observation, selection, rows, selectedKeys); err != nil {
		return nil, err
	}
	adjustments, err := advanceCustomerServiceCandidateOwnedSequencesWithoutRewinding(ctx, transaction, baselineSequences, selectedKeys)
	if err != nil {
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		return nil, errors.New("candidate transaction could not be committed")
	}
	gaps := &Manifest{}
	if err := findOwnedUniqueSequenceCollisionOrExhaustionGaps(ctx, candidate, gaps); err != nil {
		return nil, err
	}
	if len(gaps.Gaps) != 0 {
		return nil, errors.New("candidate has an unsafe owned unique sequence")
	}
	_, fingerprint, _, err := captureDatabaseInventory(ctx, candidate)
	if err != nil {
		return nil, err
	}
	// Sequences are outside MVCC. Compare fresh whole inventories after the dump
	// and copy; concurrent writers invalidate this observational rehearsal.
	for _, baseline := range []struct {
		name        string
		observation *moduleRecoveryDatabaseObservation
	}{{sourceName, source.Observation}, {targetName, target.Observation}} {
		database, err := cfg.connectDatabase(ctx, baseline.name)
		if err != nil {
			return nil, err
		}
		readTransaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
		if err != nil {
			_ = database.Close()
			return nil, err
		}
		actual, err := inspectModuleRecoveryDatabaseObservation(ctx, readTransaction)
		_ = readTransaction.Rollback()
		_ = database.Close()
		if err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(actual, baseline.observation) {
			return nil, errors.New("source or current target changed during candidate rehearsal")
		}
	}
	finalMediaFingerprint, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return nil, err
	}
	if finalMediaFingerprint != mediaFingerprint {
		return nil, errors.New("current target media changed during candidate rehearsal")
	}
	if !time.Now().Before(plan.Binding.WorkCopyExpiresAt) {
		return nil, errors.New("candidate rehearsal work copy expired")
	}
	return &customerServiceDisposableCandidateRehearsalResult{Version: 1, State: "rehearsed_not_switchable", PlanSHA256: plan.SHA256, SelectionSHA256: selection.SHA256, CleanupDispositionEntriesSHA256: selection.DependencyChecks.DurableEvents.CleanupDispositions.EntriesSHA256, BaselineFingerprint: target.Observation.Fingerprint, CandidateFingerprint: fingerprint, SelectedFullRows: len(rows), VerifiedRetainedTables: len(selection.TargetPreservation), VerifiedSequenceStates: len(baselineSequences), SequenceAdjustments: adjustments, DependencyBlockers: selection.DependencyBlockers}, nil
}

type customerServiceCandidateSnapshotDumpByteLimitWriter struct {
	File           *os.File
	RemainingBytes int64
}

func (writer *customerServiceCandidateSnapshotDumpByteLimitWriter) Write(content []byte) (int, error) {
	if int64(len(content)) > writer.RemainingBytes {
		return 0, errors.New("candidate snapshot dump byte limit exceeded")
	}
	written, err := writer.File.Write(content)
	writer.RemainingBytes -= int64(written)
	return written, err
}

// pg_dump streams into a bounded private file. A large or pathological target
// cannot grow this temporary artifact beyond the preflight workspace budget.
func (cfg Configuration) dumpCustomerServiceTargetSnapshotWithAnExplicitByteLimit(ctx context.Context, databaseName, snapshot, filename string, maximumBytes int64) error {
	if maximumBytes <= 0 {
		return errors.New("candidate dump byte limit is invalid")
	}
	connectionURL, err := cfg.databaseConnectionURL(databaseName)
	if err != nil {
		return err
	}
	parsed, _ := url.Parse(connectionURL)
	password, _ := parsed.User.Password()
	parsed.User = url.User(parsed.User.Username())
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600) // #nosec G304 -- The snapshot file is exclusively created inside the manager-owned private workspace.
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(file.Close)
	command := exec.CommandContext(ctx, "pg_dump", "--dbname", parsed.String(), "--format=custom", "--no-owner", "--no-privileges", "--no-tablespaces", "--snapshot="+snapshot) // #nosec G204 -- The executable is fixed pg_dump, with separate argv for the validated database URL and exported snapshot; no shell is used.
	command.Env = append(os.Environ(), "PGPASSWORD="+password, "PGOPTIONS=-c statement_timeout=900000 -c lock_timeout=10000")
	command.Stdout = &customerServiceCandidateSnapshotDumpByteLimitWriter{File: file, RemainingBytes: maximumBytes}
	command.Stderr = nil
	if err := command.Run(); err != nil {
		return errors.New("candidate snapshot dump failed or exceeded its byte limit")
	}
	return file.Sync()
}

func ensureCustomerServiceDisposableCandidateOwnerAndApplicationIsolation(ctx context.Context, control *sql.DB, applicationRole string) error {
	if applicationRole == "" || applicationRole == customerServiceDisposableCandidateOwner {
		return errors.New("candidate application role is invalid")
	}
	var unsafe bool
	if err := control.QueryRowContext(ctx, `SELECT rolsuper OR rolcreatedb OR rolcreaterole OR rolbypassrls FROM pg_roles WHERE rolname=$1`, applicationRole).Scan(&unsafe); err != nil || unsafe {
		return errors.New("candidate rehearsal requires an unprivileged application database role")
	}
	if _, err := control.ExecContext(ctx, `DO $$ BEGIN IF NOT EXISTS(SELECT 1 FROM pg_roles WHERE rolname='local_backup_customer_service_candidate_owner') THEN CREATE ROLE local_backup_customer_service_candidate_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS; END IF; END $$`); err != nil {
		return err
	}
	query := `SELECT rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolcanlogin OR rolbypassrls OR EXISTS(SELECT 1 FROM pg_auth_members WHERE roleid=pg_roles.oid OR member=pg_roles.oid) FROM pg_roles WHERE rolname=$1`
	if err := control.QueryRowContext(ctx, query, customerServiceDisposableCandidateOwner).Scan(&unsafe); err != nil || unsafe {
		return errors.New("candidate owner has unsafe privileges or role membership")
	}
	return nil
}

// A name prefix alone is never authority to drop a database. An interrupted
// allocation without a persisted OID is left for inspection, never guessed.
func cleanupCustomerServiceDisposableCandidateDatabaseByOID(ctx context.Context, control *sql.DB, identity customerServiceDisposableCandidateDatabaseIdentity) error {
	if identity.Version != 1 || identity.DatabaseOID <= 0 || !strings.HasPrefix(identity.DatabaseName, "local_backup_candidate_") || !safeDatabaseName.MatchString(identity.DatabaseName) {
		return errors.New("candidate cleanup requires a valid persisted database identity and OID")
	}
	var name, owner string
	if err := control.QueryRowContext(ctx, `SELECT datname,pg_get_userbyid(datdba) FROM pg_database WHERE oid=$1`, identity.DatabaseOID).Scan(&name, &owner); errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		return err
	}
	if name != identity.DatabaseName || owner != customerServiceDisposableCandidateOwner {
		return errors.New("candidate cleanup database OID, name or owner changed; inspection is required")
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(name)+` ALLOW_CONNECTIONS false`); err != nil {
		return err
	}
	if _, err := control.ExecContext(ctx, `DROP DATABASE `+pq.QuoteIdentifier(name)+` WITH (FORCE)`); err != nil {
		return fmt.Errorf("disposable candidate cleanup failed; its identity record is retained")
	}
	return nil
}

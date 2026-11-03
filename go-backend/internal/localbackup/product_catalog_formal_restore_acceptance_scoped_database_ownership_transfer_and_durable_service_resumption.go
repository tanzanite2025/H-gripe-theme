package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lib/pq"

	"commerce-platform/internal/pkg/maintenancegate"
)

// This receipt belongs only to the stopped product switch. It cannot authorize
// whole-storefront release, another module, or rollback after service resumes.
type ProductCatalogFormalRestoreAcceptance struct {
	Version                  int       `json:"version"`
	Operator                 string    `json:"operator"`
	ApplicationDatabaseRole  string    `json:"application_database_role"`
	ProjectionEvidenceSHA256 string    `json:"projection_evidence_sha256"`
	ProjectionContractSHA256 string    `json:"projection_contract_sha256"`
	AcceptedAt               time.Time `json:"accepted_at"`
	CustodyVerifiedAt        time.Time `json:"custody_verified_at,omitempty"`
	ServiceReleasedAt        time.Time `json:"service_released_at,omitempty"`
}

type ProductCatalogFormalRestoreAcceptanceRequest struct {
	Confirmation           string `json:"confirmation"`
	ExpectedEvidenceSHA256 string `json:"expected_evidence_sha256"`
}

func validateProductCatalogFormalAcceptanceReceipt(task *Task) error {
	journal := task.Journal.ProductCatalogStoppedSwitch
	receipt, evidence := journal.Acceptance, task.ProductCatalogStoppedProjectionEvidence
	if receipt == nil || receipt.Version != 1 || strings.TrimSpace(receipt.Operator) == "" || receipt.ApplicationDatabaseRole == "" || receipt.AcceptedAt.IsZero() || evidence == nil ||
		receipt.ProjectionEvidenceSHA256 != evidence.SHA256 || receipt.ProjectionContractSHA256 != evidence.ContractSHA256 || evidence.TaskID != task.ID || evidence.WindowID != journal.WindowID {
		return errors.New("product acceptance requires its exact operator, role and projection evidence")
	}
	snapshot := *evidence
	snapshot.SHA256 = ""
	digest, err := businessModuleRecoveryCanonicalJSONSHA256(snapshot)
	if err != nil || digest != evidence.SHA256 || evidence.Version != 1 || evidence.Scope != productCatalogBackupScope || evidence.RecordedAt.IsZero() {
		return errors.New("product acceptance projection evidence changed")
	}
	if receipt.AcceptedAt.Before(evidence.RecordedAt) || (!receipt.CustodyVerifiedAt.IsZero() && receipt.CustodyVerifiedAt.Before(receipt.AcceptedAt)) || (!receipt.ServiceReleasedAt.IsZero() && receipt.ServiceReleasedAt.Before(receipt.CustodyVerifiedAt)) {
		return errors.New("product acceptance evidence and release times are out of order")
	}
	if journal.Phase != "accepting_product" && receipt.CustodyVerifiedAt.IsZero() {
		return errors.New("product acceptance has no verified durable custody transfer")
	}
	if journal.Phase == "product_released" && receipt.ServiceReleasedAt.IsZero() {
		return errors.New("product acceptance release acknowledgement is missing")
	}
	return nil
}

func taskHasDurablyReleasedAcceptedProductCatalog(task *Task) bool {
	return taskHasDurablyVerifiedAcceptedProductCatalogReleaseIntent(task) && task.Journal.ProductCatalogStoppedSwitch.Phase == "product_released"
}

func taskHasDurablyVerifiedAcceptedProductCatalogReleaseIntent(task *Task) bool {
	if !taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) || (task.Journal.ProductCatalogStoppedSwitch.Phase != "releasing_product" && task.Journal.ProductCatalogStoppedSwitch.Phase != "product_released") {
		return false
	}
	if !task.Journal.DatabaseSwitched || validateProductCatalogFormalAcceptanceReceipt(task) != nil || !stoppedProductSwitchImmutableIdentityAndCompleteJournalMatch(task) {
		return false
	}
	for _, move := range task.Journal.ProductCatalogStoppedSwitch.Moves {
		if !move.Completed || move.RollbackStarted || move.RolledBack {
			return false
		}
	}
	return true
}

// This accepts the continuously stopped product switch after its own evidence
// checks. It does not promote an internal paired rehearsal or a chat candidate.
func (manager *Manager) StartProductCatalogFormalRestoreAcceptance(id, actor string, request ProductCatalogFormalRestoreAcceptanceRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if request.Confirmation != "ACCEPT PRODUCT RESTORE "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id || strings.TrimSpace(actor) == "" ||
		task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != productCatalogBackupScope || task.State != "product_module_switched_still_isolated" || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) || !taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) || task.Journal.ProductCatalogStoppedSwitch.Phase != "switched_still_isolated" ||
		task.ProductCatalogStoppedProjectionEvidence == nil || request.ExpectedEvidenceSHA256 != task.ProductCatalogStoppedProjectionEvidence.SHA256 {
		return nil, errors.New("explicit product restore acceptance and the current stopped projection evidence are required")
	}
	if err := manager.startTaskOperation(task, "accepting_verified_product_restore", func(ctx context.Context, task *Task) error {
		return manager.acceptStoppedProductCatalogRestoreAndResumeService(ctx, task, actor, nil)
	}); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

// checkpoint is used only by disposable failure/restart tests.
func (manager *Manager) acceptStoppedProductCatalogRestoreAndResumeService(ctx context.Context, task *Task, actor string, checkpoint func(string) error) error {
	// The operation wrapper marks the task running. The existing inspection must
	// still validate precisely the same switched snapshot before accepting it.
	inspectionTask := *task
	inspectionTask.State = "product_module_switched_still_isolated"
	report, err := manager.inspectStoppedProductStorefrontProjectionBaselineWithHeldLease(ctx, &inspectionTask)
	if err != nil {
		return err
	}
	evidence := task.ProductCatalogStoppedProjectionEvidence
	if evidence == nil {
		return errors.New("independent product projection evidence is required")
	}
	request := ProductCatalogStoppedProjectionEvidenceRequest{ContractSHA256: evidence.ContractSHA256, Attestations: evidence.Attestations}
	if err := validateStoppedProductProjectionOperatorAttestations(report, request); err != nil {
		return err
	}
	snapshot := *evidence
	snapshot.SHA256 = ""
	digest, err := businessModuleRecoveryCanonicalJSONSHA256(snapshot)
	if err != nil || digest != evidence.SHA256 || evidence.TaskID != task.ID || evidence.WindowID != report.WindowID || evidence.DetailCacheProgressSHA256 != report.DetailCacheProgressSHA256 || evidence.RecordedAt.IsZero() || evidence.RecordedAt.After(time.Now().UTC()) {
		return errors.New("saved product projection evidence is stale or changed")
	}
	journal := task.Journal.ProductCatalogStoppedSwitch
	journal.Acceptance = &ProductCatalogFormalRestoreAcceptance{Version: 1, Operator: actor, ApplicationDatabaseRole: manager.Configuration.ApplicationDatabaseRole, ProjectionEvidenceSHA256: evidence.SHA256, ProjectionContractSHA256: report.ContractSHA256, AcceptedAt: time.Now().UTC()}
	journal.Phase = "accepting_product"
	if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("acceptance_intent_persisted"); err != nil {
			return err
		}
	}
	if err := manager.transferStoppedProductDatabaseCustodyWithinExactDatabase(ctx, task, "local_backup_isolated_owner", manager.Configuration.ApplicationDatabaseRole); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("database_custody_committed"); err != nil {
			return err
		}
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	acceptedBinding := task.ProductCatalogPreviewBinding.Candidate
	acceptedBinding.Owner = manager.Configuration.ApplicationDatabaseRole
	if err := manager.verifyProductCatalogBoundDatabaseContents(ctx, control, manager.Configuration.DatabaseName, acceptedBinding); err != nil {
		return err
	}
	if err := manager.verifyProductCatalogBoundDatabaseContents(ctx, control, task.Journal.RollbackDatabase, task.ProductCatalogPreviewBinding.Target); err != nil {
		return err
	}
	if err := manager.validateProductCatalogStoppedSwitchPhysicalMediaLayout(ctx, task); err != nil {
		return err
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return err
	}
	if err := manager.verifyStoppedProductHandoffRedisEmptyAndDatabaseClients(ctx, task); err != nil {
		return err
	}
	client, _, err := manager.connectAndVerifyRuntimeOwnedProductViewRedis(ctx)
	if err != nil {
		return err
	}
	keys, cacheErr := inspectExactStoppedProductDetailCacheAndReadLeaseKeys(ctx, client)
	_ = client.Close()
	if cacheErr != nil || len(keys) != 0 {
		return errors.New("product cache or read leases changed before service resumption")
	}
	boundary, err := os.ReadFile(filepath.Join(manager.Configuration.MaintenanceDirectory, "application-boundary.json"))
	if err != nil {
		return err
	}
	if err := writeDurableFile(filepath.Join(manager.Configuration.MaintenanceDirectory, "restored-data-identity-boundary.json"), boundary); err != nil {
		return err
	}
	journal.Acceptance.CustodyVerifiedAt, journal.Phase = time.Now().UTC(), "releasing_product"
	task.Phase = "accepted_product_database_and_media_verified_before_service_release"
	if err := manager.saveProductCatalogStoppedSwitchProgress(task); err != nil {
		return err
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(manager.Configuration.DatabaseName)+` ALLOW_CONNECTIONS true`); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("accepted_database_open_before_fence_removal"); err != nil {
			return err
		}
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename)); err != nil {
		return err
	}
	if err := syncDirectory(manager.Configuration.MaintenanceDirectory); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("accepted_fence_removed_before_acknowledgement"); err != nil {
			return err
		}
	}
	return manager.acknowledgeAcceptedProductCatalogServiceRelease(task)
}

func (manager *Manager) acknowledgeAcceptedProductCatalogServiceRelease(task *Task) error {
	if !taskHasDurablyVerifiedAcceptedProductCatalogReleaseIntent(task) {
		return errors.New("accepted product release receipt or immutable identity changed")
	}
	journal := task.Journal.ProductCatalogStoppedSwitch
	if journal.Phase != "releasing_product" && journal.Phase != "product_released" {
		return errors.New("accepted product service release has no durable verified intent")
	}
	if manager.productCatalogStoppedHandoffLease != nil && manager.productCatalogStoppedHandoffLeaseTaskID != task.ID {
		return errors.New("another maintenance window holds the application lease")
	}
	if err := manager.closeProductCatalogStoppedHandoffLease(); err != nil {
		return err
	}
	journal.Phase = "product_released"
	if journal.Acceptance.ServiceReleasedAt.IsZero() {
		journal.Acceptance.ServiceReleasedAt = time.Now().UTC()
	}
	task.State, task.Phase, task.Error = "completed", "product_restore_accepted_and_service_resumed", ""
	return manager.saveProductCatalogStoppedSwitchProgress(task)
}

// REASSIGN OWNED also changes shared database objects throughout the cluster.
// Enumerate only this bound database's public objects and transfer them in one
// transaction, so another isolated candidate or another application's DB stays
// untouched. A crash leaves either the entire old or entire new ownership.
func (manager *Manager) transferStoppedProductDatabaseCustodyWithinExactDatabase(ctx context.Context, task *Task, fromRole, toRole string) (resultErr error) {
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return err
	}
	cfg := manager.Configuration
	if task.Journal.ProductCatalogStoppedSwitch.Acceptance == nil || task.Journal.ProductCatalogStoppedSwitch.Acceptance.ApplicationDatabaseRole != cfg.ApplicationDatabaseRole ||
		(fromRole != "local_backup_isolated_owner" || toRole != cfg.ApplicationDatabaseRole) && (toRole != "local_backup_isolated_owner" || fromRole != cfg.ApplicationDatabaseRole) {
		return errors.New("product database custody roles do not match the accepted application")
	}
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var unsafeRole bool
	if err := control.QueryRowContext(ctx, `SELECT rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls OR pg_has_role(oid,'local_backup_isolated_owner','MEMBER') FROM pg_roles WHERE rolname=$1`, cfg.ApplicationDatabaseRole).Scan(&unsafeRole); err != nil || unsafeRole {
		return errors.New("product acceptance requires an unprivileged application role independent of the isolated owner")
	}
	expected := task.ProductCatalogPreviewBinding.Candidate
	expected.Owner = fromRole
	if err := manager.verifyProductCatalogBoundDatabaseContents(ctx, control, cfg.DatabaseName, expected); err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		resultErr = errors.Join(resultErr, sealDatabase(cleanup, control, cfg.DatabaseName))
	}()
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` ALLOW_CONNECTIONS true`); err != nil {
		return err
	}
	database, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(database.Close)
	tx, err := database.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(tx.Rollback)
	rows, err := tx.QueryContext(ctx, `SELECT command FROM (
SELECT format('ALTER %s %I.%I OWNER TO ',CASE c.relkind WHEN 'S' THEN 'SEQUENCE' WHEN 'v' THEN 'VIEW' WHEN 'm' THEN 'MATERIALIZED VIEW' WHEN 'f' THEN 'FOREIGN TABLE' ELSE 'TABLE' END,n.nspname,c.relname) command,1 priority
FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relowner=(SELECT oid FROM pg_roles WHERE rolname=$1) AND c.relkind IN ('r','p','v','m','f')
UNION ALL SELECT format('ALTER SEQUENCE %I.%I OWNER TO ',n.nspname,c.relname),2 FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relowner=(SELECT oid FROM pg_roles WHERE rolname=$1) AND c.relkind='S'
UNION ALL SELECT format('ALTER ROUTINE %I.%I(%s) OWNER TO ',n.nspname,p.proname,pg_get_function_identity_arguments(p.oid)),3 FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='public' AND p.proowner=(SELECT oid FROM pg_roles WHERE rolname=$1) AND NOT EXISTS(SELECT 1 FROM pg_depend d WHERE d.classid='pg_proc'::regclass AND d.objid=p.oid AND d.deptype='e')
UNION ALL SELECT format('ALTER %s %I.%I OWNER TO ',CASE t.typtype WHEN 'd' THEN 'DOMAIN' ELSE 'TYPE' END,n.nspname,t.typname),4 FROM pg_type t JOIN pg_namespace n ON n.oid=t.typnamespace WHERE n.nspname='public' AND t.typowner=(SELECT oid FROM pg_roles WHERE rolname=$1) AND t.typtype IN ('e','d','c','r','m') AND (t.typrelid=0 OR EXISTS(SELECT 1 FROM pg_class c WHERE c.oid=t.typrelid AND c.relkind='c')) AND NOT EXISTS(SELECT 1 FROM pg_depend d WHERE d.classid='pg_type'::regclass AND d.objid=t.oid AND d.deptype='e')
UNION ALL SELECT format('ALTER SCHEMA %I OWNER TO ',nspname),4 FROM pg_namespace WHERE nspname='public' AND nspowner=(SELECT oid FROM pg_roles WHERE rolname=$1)
) scoped_objects ORDER BY priority,command LIMIT 100001`, fromRole)
	if err != nil {
		return err
	}
	var commands []string
	for rows.Next() {
		var command string
		if err := rows.Scan(&command); err != nil {
			_ = rows.Close()
			return err
		}
		commands = append(commands, command+pq.QuoteIdentifier(toRole))
	}
	readErr, closeErr := rows.Err(), rows.Close()
	if readErr != nil || closeErr != nil || len(commands) > 100000 {
		return errors.Join(readErr, closeErr, errors.New("product ownership inventory is unavailable or exceeds its bound"))
	}
	for _, command := range commands {
		if _, err := tx.ExecContext(ctx, command); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` OWNER TO `+pq.QuoteIdentifier(toRole)); err != nil {
		return err
	}
	return tx.Commit()
}

func (manager *Manager) restoreAcceptedProductIsolatedCustodyBeforePairedRollback(ctx context.Context, task *Task) error {
	journal := task.Journal.ProductCatalogStoppedSwitch
	if journal.Phase != "accepting_product" && journal.Phase != "releasing_product" {
		return nil
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	var owner string
	err = control.QueryRowContext(ctx, `SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname=$1 AND oid=$2`, manager.Configuration.DatabaseName, task.Journal.NewDatabaseOID).Scan(&owner)
	_ = control.Close()
	if err != nil {
		return err
	}
	if owner == manager.Configuration.ApplicationDatabaseRole {
		if err := manager.transferStoppedProductDatabaseCustodyWithinExactDatabase(ctx, task, owner, "local_backup_isolated_owner"); err != nil {
			return err
		}
	} else if owner != "local_backup_isolated_owner" {
		return errors.New("interrupted product acceptance database owner changed; retain isolation")
	}
	journal.Acceptance, journal.Phase = nil, "switched_still_isolated"
	return manager.saveProductCatalogStoppedSwitchProgress(task)
}

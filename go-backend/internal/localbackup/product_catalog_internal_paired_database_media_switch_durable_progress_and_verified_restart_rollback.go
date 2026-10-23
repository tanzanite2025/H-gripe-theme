package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Internal rehearsal only. No HTTP route or production resume operation accepts
// this journal. The original database and complete media must be restored before
// the persistent application fence can be released.
type productCatalogPairedRehearsalJournal struct {
	Version              int                            `json:"version"`
	TaskID               string                         `json:"task_id"`
	PreviewBindingSHA256 string                         `json:"preview_binding_sha256"`
	PackageBindingSHA256 string                         `json:"package_binding_sha256"`
	Moves                []moduleRecoveryMediaEntryMove `json:"moves"`
	RollbackVerified     bool                           `json:"rollback_verified"`
	SHA256               string                         `json:"sha256"`
}

func taskHasProductCatalogPairedRecoveryJournal(task *Task) bool {
	return task.Journal != nil && !taskHasProductCatalogContinuouslyStoppedSwitchJournal(task) && (task.Journal.ProductCatalog != nil || (task.Manifest != nil && task.Manifest.Scope == productCatalogBackupScope))
}

func (manager *Manager) requireNoUnresolvedProductCatalogPairedRehearsalExceptTask(id string) error {
	tasks, err := manager.ListTasks()
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.ID != id && taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
			return errors.New("an unresolved stopped product handoff preparation retains the maintenance window; resume or cancel that task first")
		}
		if task.ID != id && task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
			return errors.New("an unresolved product view drain retains the maintenance window; resume that task first")
		}
		if task.ID != id && taskHasProductCatalogPairedRecoveryJournal(task) && task.State != "rolled_back" {
			return errors.New("an unresolved product paired journal requires verified rollback before another maintenance operation")
		}
	}
	return nil
}

func productCatalogPairedRehearsalPackageBinding(task *Task) (string, error) {
	return customerServiceRecoveryPlanCanonicalJSONSHA256(struct {
		Manifest                                                             *Manifest
		CiphertextSHA256, StageDatabase, StageFingerprint, TargetFingerprint string
		ExpiresAt                                                            time.Time
	}{task.Manifest, task.CiphertextSHA256, task.StageDatabase, task.StageFingerprint, task.PreviewTargetFingerprint, task.ExpiresAt})
}

func (manager *Manager) saveProductCatalogPairedRehearsalProgress(task *Task) error {
	journal := task.Journal.ProductCatalog
	journal.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(journal)
	if err != nil {
		return err
	}
	journal.SHA256 = digest
	return manager.saveTask(task)
}

func (manager *Manager) validateProductCatalogPairedRehearsalBinding(task *Task) error {
	if task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != productCatalogBackupScope || task.ProductCatalogPreviewBinding == nil || task.ProductCatalogPreviewBinding.TargetMedia == nil || task.ProductCatalogPreviewBinding.CandidateMedia == nil || task.Journal == nil || task.Journal.ProductCatalog == nil {
		return errors.New("product paired rehearsal binding is missing")
	}
	journal, binding := task.Journal.ProductCatalog, task.ProductCatalogPreviewBinding
	snapshotCopy := *journal
	snapshotCopy.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(snapshotCopy)
	if err != nil || digest != journal.SHA256 || journal.Version != 1 || journal.TaskID != task.ID || binding.Version != 1 {
		return errors.New("product paired rehearsal journal identity or digest changed")
	}
	digest, err = customerServiceRecoveryPlanCanonicalJSONSHA256(manager.Configuration)
	if err != nil || digest != binding.ConfigurationSHA256 {
		return errors.New("product paired rehearsal configuration changed")
	}
	digest, err = customerServiceRecoveryPlanCanonicalJSONSHA256(binding)
	if err != nil || digest != journal.PreviewBindingSHA256 {
		return errors.New("product paired preview binding changed")
	}
	digest, err = productCatalogPairedRehearsalPackageBinding(task)
	if err != nil || digest != journal.PackageBindingSHA256 {
		return errors.New("product paired package or stage binding changed")
	}
	if task.StageDatabase != "local_backup_stage_"+task.ID[:16] || task.Journal.RollbackDatabase != "local_backup_product_old_"+task.ID[:16] || task.Journal.OldDatabaseOID != binding.Target.OID || task.Journal.NewDatabaseOID != binding.Candidate.OID || binding.Target.OID <= 0 || binding.Candidate.OID <= 0 || binding.Target.OID == binding.Candidate.OID || len(task.Journal.Moves) != 0 {
		return errors.New("product paired database identity changed")
	}
	expected, err := buildModuleRecoveryPairedMediaMoves(binding.TargetMedia, binding.CandidateMedia)
	if err != nil || len(expected) != len(journal.Moves) {
		return errors.New("product paired complete media move contract changed")
	}
	for index, move := range journal.Moves {
		move.Started, move.Completed, move.RollbackStarted, move.RolledBack = false, false, false, false
		if move != expected[index] {
			return errors.New("product paired media move identity changed")
		}
	}
	return manager.Configuration.validateRuntimeBoundary()
}

func (manager *Manager) productCatalogPairedMediaVolumeDirectory(task *Task, rootName string) string {
	return filepath.Join(manager.Configuration.StorageRoots[rootName], ".local-backup-product-pair-"+task.ID)
}

func (manager *Manager) productCatalogPairedMediaLogicalPath(task *Task, logical string) (string, error) {
	parts := strings.SplitN(logical, "/", 3)
	if len(parts) != 3 || strings.ContainsAny(parts[2], "/\\") {
		return "", errors.New("product paired logical media path is invalid")
	}
	if _, err := customerServiceCandidateMediaPathComponents(parts[1], parts[2]); err != nil {
		return "", err
	}
	if parts[0] == "live" {
		return filepath.Join(manager.Configuration.StorageRoots[parts[1]], parts[2]), nil
	}
	if parts[0] != "candidate" && parts[0] != "previous" {
		return "", errors.New("product paired media location is invalid")
	}
	return filepath.Join(manager.productCatalogPairedMediaVolumeDirectory(task, parts[1]), parts[0], parts[2]), nil
}

func (manager *Manager) inspectProductCatalogPairedPhysicalMediaLayout(ctx context.Context, task *Task, includeLive bool) (map[string]string, error) {
	layout := map[string]string{}
	for _, rootName := range mediaRoots {
		volume := manager.productCatalogPairedMediaVolumeDirectory(task, rootName)
		for _, parent := range []string{manager.Configuration.StorageRoots[rootName], volume, filepath.Join(volume, "candidate"), filepath.Join(volume, "previous")} {
			info, err := os.Lstat(parent)
			if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return nil, errors.New("product paired media volume is missing or unsafe")
			}
		}
		marker, err := readCustomerServicePairedRecoveryBoundedRegularFile(filepath.Join(volume, "product-paired-volume-binding.json"), 4096)
		var identity struct{ TaskID, RootName, PreviewBindingSHA256 string }
		previewDigest, previewErr := customerServiceRecoveryPlanCanonicalJSONSHA256(task.ProductCatalogPreviewBinding)
		if err != nil || previewErr != nil || task.ProductCatalogPreviewBinding == nil || decodeStrictJSON(marker, &identity) != nil || identity.TaskID != task.ID || identity.RootName != rootName || identity.PreviewBindingSHA256 != previewDigest {
			return nil, errors.New("product paired media volume identity changed")
		}
		entries, err := os.ReadDir(volume)
		if err != nil {
			return nil, err
		}
		if len(entries) != 3 {
			return nil, errors.New("product paired media volume has unregistered entries")
		}
		for _, location := range []string{"live", "candidate", "previous"} {
			if location == "live" && !includeLive {
				continue
			}
			parent := manager.Configuration.StorageRoots[rootName]
			if location != "live" {
				parent = filepath.Join(volume, location)
			}
			entries, err := os.ReadDir(parent)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				if location == "live" && strings.HasPrefix(entry.Name(), ".local-backup-") {
					continue
				}
				digest, err := manager.Configuration.inspectModuleRecoveryMediaEntryDigest(ctx, parent, rootName, entry.Name())
				if err != nil {
					return nil, err
				}
				layout[location+"/"+rootName+"/"+entry.Name()] = digest
			}
		}
	}
	return layout, nil
}

func (manager *Manager) prepareProductCatalogPairedCompleteMediaVolumes(ctx context.Context, task *Task, previewDigest string) (resultErr error) {
	cfg := manager.Configuration
	// Account all roots sharing a filesystem together; originals are renamed,
	// not recopied for rollback. Charge directory/allocation overhead as well.
	groups := map[string]struct {
		need int64
		free uint64
	}{}
	for _, rootName := range mediaRoots {
		root := cfg.StorageRoots[rootName]
		filesystem, err := customerServiceCandidateCapacityFilesystemIdentity(root)
		if err != nil {
			return err
		}
		free, err := availableDiskBytes(root)
		if err != nil {
			return err
		}
		group := groups[filesystem]
		if group.free == 0 || free < group.free {
			group.free = free
		}
		group.need += 64 << 20
		for _, file := range task.ProductCatalogPreviewBinding.CandidateMedia.Files {
			if file.RootName == rootName {
				group.need += file.Identity.Size + (64 << 10)
			}
		}
		for _, directory := range task.ProductCatalogPreviewBinding.CandidateMedia.Directories {
			if directory.RootName == rootName {
				group.need += 64 << 10
			}
		}
		groups[filesystem] = group
	}
	for _, group := range groups {
		if group.need < 0 || uint64(group.need) > group.free {
			return errors.New("insufficient combined media filesystem capacity for product paired recovery")
		}
	}
	created := []string{}
	defer func() {
		if resultErr != nil && task.Journal == nil {
			for _, directory := range created {
				resultErr = errors.Join(resultErr, os.RemoveAll(directory))
			}
		}
	}()
	candidateCfg := manager.productCatalogCandidateMediaConfiguration(task)
	for _, rootName := range mediaRoots {
		volume := manager.productCatalogPairedMediaVolumeDirectory(task, rootName)
		if err := os.Mkdir(volume, 0700); err != nil {
			return err
		}
		created = append(created, volume)
		if err := writeDurableJSON(filepath.Join(volume, "product-paired-volume-binding.json"), struct{ TaskID, RootName, PreviewBindingSHA256 string }{task.ID, rootName, previewDigest}); err != nil {
			return err
		}
		for _, location := range []string{"candidate", "previous"} {
			if err := os.Mkdir(filepath.Join(volume, location), 0700); err != nil {
				return err
			}
		}
		for _, directory := range task.ProductCatalogPreviewBinding.CandidateMedia.Directories {
			if directory.RootName != rootName {
				continue
			}
			if err := os.MkdirAll(filepath.Join(volume, "candidate", filepath.FromSlash(directory.Key)), 0700); err != nil {
				return err
			}
		}
		for _, file := range task.ProductCatalogPreviewBinding.CandidateMedia.Files {
			if file.RootName != rootName {
				continue
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			identity, err := copyVerifiedRegularFile(filepath.Join(candidateCfg.StorageRoots[rootName], filepath.FromSlash(file.Key)), filepath.Join(volume, "candidate", filepath.FromSlash(file.Key)), file.Identity.Path)
			if err != nil || identity != file.Identity {
				return errors.New("product paired complete media copy changed")
			}
		}
		fileMode, directoryMode := os.FileMode(0600), os.FileMode(0700)
		if rootName != "uploads_private" {
			fileMode, directoryMode = 0644, 0755
		}
		if err := applyRecoveryMediaTreePermissionsWithinOpenedRoot(ctx, filepath.Join(volume, "candidate"), fileMode, directoryMode, false); err != nil {
			return err
		}
		if err := syncDirectory(volume); err != nil {
			return err
		}
		if err := syncDirectory(cfg.StorageRoots[rootName]); err != nil {
			return err
		}
	}
	return nil
}

func (manager *Manager) verifyProductCatalogBoundDatabaseContents(ctx context.Context, control *sql.DB, name string, expected productCatalogRecoveryDatabaseBinding) (resultErr error) {
	var oid int64
	var owner string
	var allowed bool
	if err := control.QueryRowContext(ctx, `SELECT oid::bigint,pg_get_userbyid(datdba),datallowconn FROM pg_database WHERE datname=$1`, name).Scan(&oid, &owner, &allowed); err != nil {
		return err
	}
	if oid != expected.OID || owner != expected.Owner {
		return errors.New("product paired database OID or owner changed")
	}
	if owner == "local_backup_isolated_owner" {
		var unsafeOwner, applicationCanAccess bool
		if err := control.QueryRowContext(ctx, `SELECT rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolcanlogin OR rolbypassrls FROM pg_roles WHERE rolname=$1`, owner).Scan(&unsafeOwner); err != nil {
			return err
		}
		if err := control.QueryRowContext(ctx, `SELECT has_database_privilege($1,$2,'CONNECT') OR has_database_privilege($1,$2,'CREATE') OR has_database_privilege($1,$2,'TEMP')`, manager.Configuration.ApplicationDatabaseRole, name).Scan(&applicationCanAccess); err != nil {
			return err
		}
		if unsafeOwner || applicationCanAccess {
			return errors.New("product paired candidate owner or application isolation changed")
		}
	}
	if err := assertNoUnfencedDatabaseSessions(ctx, control, name); err != nil {
		return err
	}
	if !allowed {
		defer func() {
			cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			var actualOID int64
			if err := readDatabaseOID(cleanup, control, name, &actualOID); err != nil || actualOID != expected.OID {
				resultErr = errors.Join(resultErr, errors.New("product paired database cannot safely be resealed"))
				return
			}
			resultErr = errors.Join(resultErr, sealDatabase(cleanup, control, name))
		}()
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(name)+` ALLOW_CONNECTIONS true`); err != nil { // #nosec G701 -- Dynamic database, role, table and column identifiers use pq.QuoteIdentifier; data values remain bound parameters.
			return err
		}
	}
	actual, err := manager.Configuration.captureProductCatalogRecoveryDatabaseBinding(ctx, name)
	if err != nil {
		return err
	}
	if *actual != expected {
		return errors.New("product paired database data, sequence or schema changed")
	}
	return assertNoUnfencedDatabaseSessions(ctx, control, name)
}

func (manager *Manager) inspectProductCatalogPairedDatabaseTopology(ctx context.Context, control *sql.DB, task *Task) (string, string, error) {
	cfg, journal := manager.Configuration, task.Journal
	positions := map[string]int64{}
	for _, name := range []string{cfg.DatabaseName, task.StageDatabase, journal.RollbackDatabase} {
		var oid int64
		if err := readDatabaseOID(ctx, control, name, &oid); err != nil {
			return "", "", err
		}
		positions[name] = oid
	}
	original, candidate := "", ""
	switch {
	case positions[cfg.DatabaseName] == journal.OldDatabaseOID && positions[task.StageDatabase] == journal.NewDatabaseOID && positions[journal.RollbackDatabase] == 0:
		original, candidate = cfg.DatabaseName, task.StageDatabase
	case positions[cfg.DatabaseName] == 0 && positions[task.StageDatabase] == journal.NewDatabaseOID && positions[journal.RollbackDatabase] == journal.OldDatabaseOID:
		original, candidate = journal.RollbackDatabase, task.StageDatabase
	case positions[cfg.DatabaseName] == journal.NewDatabaseOID && positions[task.StageDatabase] == 0 && positions[journal.RollbackDatabase] == journal.OldDatabaseOID:
		original, candidate = journal.RollbackDatabase, cfg.DatabaseName
	default:
		return "", "", errors.New("product paired database topology changed; maintenance retained")
	}
	if err := manager.verifyProductCatalogBoundDatabaseContents(ctx, control, original, task.ProductCatalogPreviewBinding.Target); err != nil {
		return "", "", err
	}
	if err := manager.verifyProductCatalogBoundDatabaseContents(ctx, control, candidate, task.ProductCatalogPreviewBinding.Candidate); err != nil {
		return "", "", err
	}
	return original, candidate, nil
}

// checkpoint is an internal fault-injection boundary, never a network input.
// An error retains the journal and fence so restart recovery can verify rollback.
func (manager *Manager) rehearseProductCatalogPairedSwitchAndVerifiedRollback(ctx context.Context, id string, checkpoint func(string) error) error {
	if !manager.operationMutex.TryLock() {
		return errors.New("an operation is running")
	}
	defer manager.operationMutex.Unlock()
	task, err := manager.GetTask(id)
	if err != nil {
		return err
	}
	if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
		return errors.New("stopped product handoff preparation cannot be consumed by an internal rehearsal")
	}
	if task.State != "product_module_isolated_verified" || task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != productCatalogBackupScope || task.Journal != nil || task.ProductCatalogPreviewBinding == nil || task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("a fresh bound product candidate is required for internal rehearsal")
	}
	cfg, binding := manager.Configuration, task.ProductCatalogPreviewBinding
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return err
	}
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	defer func() {
		if task.Journal == nil {
			_ = manager.releaseMaintenance(task)
		}
	}()
	if err := manager.validateProductCatalogCandidateStillMatchesIsolatedPreview(ctx, task); err != nil {
		return err
	}
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(control.Close)
	var unsafeApplicationRole bool
	if err := control.QueryRowContext(ctx, `SELECT rolsuper OR rolcreatedb OR rolcreaterole OR rolbypassrls OR pg_has_role(oid,'local_backup_isolated_owner','MEMBER') FROM pg_roles WHERE rolname=$1`, cfg.ApplicationDatabaseRole).Scan(&unsafeApplicationRole); err != nil {
		return err
	}
	if unsafeApplicationRole {
		return errors.New("product paired rehearsal requires an unprivileged application role independent of the candidate owner")
	}
	if err := assertNoUnfencedDatabaseSessions(ctx, control, cfg.DatabaseName); err != nil {
		return err
	}
	if err := assertNoUnfencedDatabaseSessions(ctx, control, task.StageDatabase); err != nil {
		return err
	}
	previewDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(binding)
	if err != nil {
		return err
	}
	packageDigest, err := productCatalogPairedRehearsalPackageBinding(task)
	if err != nil {
		return err
	}
	moves, err := buildModuleRecoveryPairedMediaMoves(binding.TargetMedia, binding.CandidateMedia)
	if err != nil {
		return err
	}
	rollbackName := "local_backup_product_old_" + task.ID[:16]
	var occupied int64
	if err := readDatabaseOID(ctx, control, rollbackName, &occupied); err != nil {
		return err
	}
	if occupied != 0 {
		return errors.New("product paired rollback database name is occupied")
	}
	if err := manager.prepareProductCatalogPairedCompleteMediaVolumes(ctx, task, previewDigest); err != nil {
		return err
	}
	task.Journal = &RecoveryJournal{OldDatabaseOID: binding.Target.OID, NewDatabaseOID: binding.Candidate.OID, RollbackDatabase: rollbackName, Moves: []RecoveryFileMove{}, ProductCatalog: &productCatalogPairedRehearsalJournal{Version: 1, TaskID: task.ID, PreviewBindingSHA256: previewDigest, PackageBindingSHA256: packageDigest, Moves: moves}}
	task.State, task.Phase = "running", "product_paired_rehearsal_journal_persisted"
	if err := manager.saveProductCatalogPairedRehearsalProgress(task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("journal_persisted"); err != nil {
			return err
		}
	}
	// Persist before changing even connection privileges. Sealed candidates never
	// receive ownership or production access during an internal rehearsal.
	if _, err := control.ExecContext(ctx, `REVOKE ALL ON DATABASE `+pq.QuoteIdentifier(task.StageDatabase)+` FROM PUBLIC,`+pq.QuoteIdentifier(cfg.ApplicationDatabaseRole)); err != nil {
		return err
	}
	for _, name := range []string{cfg.DatabaseName, task.StageDatabase} {
		if err := sealDatabase(ctx, control, name); err != nil {
			return err
		}
	}
	if _, _, err := manager.inspectProductCatalogPairedDatabaseTopology(ctx, control, task); err != nil {
		return err
	}
	if err := manager.validateProductCatalogPairedPhysicalLayout(ctx, task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("databases_sealed"); err != nil {
			return err
		}
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(cfg.DatabaseName)+` RENAME TO `+pq.QuoteIdentifier(rollbackName)); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("original_database_renamed"); err != nil {
			return err
		}
	}
	for index := range task.Journal.ProductCatalog.Moves {
		move := &task.Journal.ProductCatalog.Moves[index]
		move.Started = true
		if err := manager.saveProductCatalogPairedRehearsalProgress(task); err != nil {
			return err
		}
		if checkpoint != nil {
			if err := checkpoint(fmt.Sprintf("media_%d_intent", index)); err != nil {
				return err
			}
		}
		if err := manager.performProductCatalogPairedMediaRename(ctx, task, *move, false); err != nil {
			return err
		}
		if checkpoint != nil {
			if err := checkpoint(fmt.Sprintf("media_%d_renamed_before_acknowledgement", index)); err != nil {
				return err
			}
		}
		move.Completed = true
		if err := manager.saveProductCatalogPairedRehearsalProgress(task); err != nil {
			return err
		}
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(task.StageDatabase)+` RENAME TO `+pq.QuoteIdentifier(cfg.DatabaseName)); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("candidate_database_renamed_before_acknowledgement"); err != nil {
			return err
		}
	}
	task.Journal.DatabaseSwitched = true
	task.State, task.Phase = "product_module_rehearsal_switched_service_isolated", "product_database_and_complete_media_switched_internal_rehearsal_only"
	if err := manager.saveProductCatalogPairedRehearsalProgress(task); err != nil {
		return err
	}
	if _, _, err := manager.inspectProductCatalogPairedDatabaseTopology(ctx, control, task); err != nil {
		return err
	}
	if err := manager.validateProductCatalogPairedPhysicalLayout(ctx, task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("paired_switch_verified"); err != nil {
			return err
		}
	}
	return manager.rollbackProductCatalogPairedRehearsalWithHeldLease(ctx, task, control, checkpoint)
}

func (manager *Manager) validateProductCatalogPairedPhysicalLayout(ctx context.Context, task *Task) error {
	if err := manager.validateProductCatalogPairedRehearsalBinding(task); err != nil {
		return err
	}
	layout, err := manager.inspectProductCatalogPairedPhysicalMediaLayout(ctx, task, true)
	if err != nil {
		return err
	}
	return validateModuleRecoveryPairedMediaLayout(task.Journal.ProductCatalog.Moves, layout)
}

func (manager *Manager) performProductCatalogPairedMediaRename(ctx context.Context, task *Task, move moduleRecoveryMediaEntryMove, reverse bool) error {
	if err := manager.validateProductCatalogPairedPhysicalLayout(ctx, task); err != nil {
		return err
	}
	return manager.renameProductCatalogPairedMediaEntryAfterLayoutValidation(ctx, task, move, reverse)
}

// Both callers validate their own distinct durable journal and full physical
// layout before this shared same-filesystem rename primitive.
func (manager *Manager) renameProductCatalogPairedMediaEntryAfterLayoutValidation(ctx context.Context, task *Task, move moduleRecoveryMediaEntryMove, reverse bool) error {
	sourceLogical, destinationLogical := customerServicePairedSwitchMediaMoveLogicalPaths(move)
	if reverse {
		sourceLogical, destinationLogical = destinationLogical, sourceLogical
	}
	source, err := manager.productCatalogPairedMediaLogicalPath(task, sourceLogical)
	if err != nil {
		return err
	}
	destination, err := manager.productCatalogPairedMediaLogicalPath(task, destinationLogical)
	if err != nil {
		return err
	}
	_, sourceErr := os.Lstat(source)           // #nosec G703 -- productCatalogPairedMediaLogicalPath validates journal paths against the registered root and task-owned volume.
	_, destinationErr := os.Lstat(destination) // #nosec G703 -- productCatalogPairedMediaLogicalPath validates journal paths against the registered root and task-owned volume.
	if reverse && destinationErr == nil {
		digest, err := manager.Configuration.inspectModuleRecoveryMediaEntryDigest(ctx, filepath.Dir(destination), move.RootName, move.EntryName)
		if err == nil && digest == move.EntrySHA256 && (errors.Is(sourceErr, os.ErrNotExist) || !move.Completed) {
			return nil
		}
	}
	if sourceErr != nil || !errors.Is(destinationErr, os.ErrNotExist) {
		return errors.New("product paired media rename endpoints are ambiguous")
	}
	digest, err := manager.Configuration.inspectModuleRecoveryMediaEntryDigest(ctx, filepath.Dir(source), move.RootName, move.EntryName)
	if err != nil || digest != move.EntrySHA256 {
		return errors.New("product paired media rename source content changed")
	}
	if err := os.Rename(source, destination); err != nil { // #nosec G703 -- productCatalogPairedMediaLogicalPath validates journal paths against the registered root and task-owned volume.
		return err
	}
	return errors.Join(syncDirectory(filepath.Dir(source)), syncDirectory(filepath.Dir(destination)))
}

func (manager *Manager) rollbackProductCatalogPairedRehearsal(ctx context.Context, task *Task, checkpoint func(string) error) error {
	if err := manager.validateProductCatalogPairedRehearsalBinding(task); err != nil {
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
	return manager.rollbackProductCatalogPairedRehearsalWithHeldLease(ctx, task, control, checkpoint)
}

func (manager *Manager) rollbackProductCatalogPairedRehearsalWithHeldLease(ctx context.Context, task *Task, control *sql.DB, checkpoint func(string) error) error {
	if err := manager.validateProductCatalogPairedRehearsalBinding(task); err != nil {
		return err
	}
	// A crash immediately after journal persistence can precede privilege
	// isolation. Reapply only to the bound OID, never to an unverified name.
	var candidateOID int64
	if err := readDatabaseOID(ctx, control, task.StageDatabase, &candidateOID); err != nil {
		return err
	}
	if candidateOID == task.Journal.NewDatabaseOID {
		if _, err := control.ExecContext(ctx, `REVOKE ALL ON DATABASE `+pq.QuoteIdentifier(task.StageDatabase)+` FROM PUBLIC,`+pq.QuoteIdentifier(manager.Configuration.ApplicationDatabaseRole)); err != nil { // #nosec G701 -- Dynamic database, role, table and column identifiers use pq.QuoteIdentifier; data values remain bound parameters.
			return err
		}
	}
	original, candidate, err := manager.inspectProductCatalogPairedDatabaseTopology(ctx, control, task)
	if err != nil {
		return err
	}
	if err := manager.validateProductCatalogPairedPhysicalLayout(ctx, task); err != nil {
		return err
	}
	for _, name := range []string{original, candidate} {
		if err := sealDatabase(ctx, control, name); err != nil {
			return err
		}
	}
	if candidate == manager.Configuration.DatabaseName {
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(candidate)+` RENAME TO `+pq.QuoteIdentifier(task.StageDatabase)); err != nil { // #nosec G701 -- Dynamic database, role, table and column identifiers use pq.QuoteIdentifier; data values remain bound parameters.
			return err
		}
		if checkpoint != nil {
			if err := checkpoint("candidate_database_returned_before_media_rollback"); err != nil {
				return err
			}
		}
	}
	for index := len(task.Journal.ProductCatalog.Moves) - 1; index >= 0; index-- {
		move := &task.Journal.ProductCatalog.Moves[index]
		if !move.Started || move.RolledBack {
			continue
		}
		move.RollbackStarted = true
		if err := manager.saveProductCatalogPairedRehearsalProgress(task); err != nil {
			return err
		}
		if err := manager.performProductCatalogPairedMediaRename(ctx, task, *move, true); err != nil {
			return err
		}
		if checkpoint != nil {
			if err := checkpoint(fmt.Sprintf("media_%d_reversed_before_acknowledgement", index)); err != nil {
				return err
			}
		}
		move.RolledBack = true
		if err := manager.saveProductCatalogPairedRehearsalProgress(task); err != nil {
			return err
		}
	}
	if original == task.Journal.RollbackDatabase {
		if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(original)+` RENAME TO `+pq.QuoteIdentifier(manager.Configuration.DatabaseName)); err != nil { // #nosec G701 -- Dynamic database, role, table and column identifiers use pq.QuoteIdentifier; data values remain bound parameters.
			return err
		}
		if checkpoint != nil {
			if err := checkpoint("original_database_restored_before_verification"); err != nil {
				return err
			}
		}
	}
	if _, _, err := manager.inspectProductCatalogPairedDatabaseTopology(ctx, control, task); err != nil {
		return err
	}
	if err := manager.validateProductCatalogPairedPhysicalLayout(ctx, task); err != nil {
		return err
	}
	// The sealed content checks include every row, sequence and schema contract;
	// the full layout check includes all retained files and empty directories.
	task.Journal.ProductCatalog.RollbackVerified = true
	task.State, task.Phase = "releasing_rollback", "product_original_database_and_complete_media_verified"
	if err := manager.saveProductCatalogPairedRehearsalProgress(task); err != nil {
		return err
	}
	if _, err := control.ExecContext(ctx, `ALTER DATABASE `+pq.QuoteIdentifier(manager.Configuration.DatabaseName)+` ALLOW_CONNECTIONS true`); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("rollback_verified_before_maintenance_release"); err != nil {
			return err
		}
	}
	if err := manager.releaseMaintenance(task); err != nil {
		return err
	}
	task.State = "rolled_back"
	task.Error = ""
	return manager.saveProductCatalogPairedRehearsalProgress(task)
}

func (manager *Manager) recoverInterruptedProductCatalogPairedRehearsals(tasks []*Task) {
	for _, task := range tasks {
		if !taskHasProductCatalogPairedRecoveryJournal(task) || task.State == "rolled_back" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		err := manager.rollbackProductCatalogPairedRehearsal(ctx, task, nil)
		cancel()
		if err != nil {
			task.State, task.Phase, task.Error = "interrupted", "product_paired_rehearsal_retained_for_inspection", err.Error()
			_ = manager.saveTask(task)
		}
	}
}

func (manager *Manager) cleanupProductCatalogVerifiedPairedMediaVolumes(ctx context.Context, task *Task) error {
	if task.State != "rolled_back" || task.Journal == nil || task.Journal.ProductCatalog == nil || !task.Journal.ProductCatalog.RollbackVerified {
		return errors.New("product paired media cleanup requires verified rollback")
	}
	if err := manager.validateProductCatalogPairedRehearsalBinding(task); err != nil {
		return err
	}
	actual, err := manager.inspectProductCatalogPairedPhysicalMediaLayout(ctx, task, false)
	if err != nil {
		return err
	}
	expected := map[string]string{}
	for _, move := range task.Journal.ProductCatalog.Moves {
		if move.Started && !move.RolledBack {
			return errors.New("product media cleanup has pending reverse moves")
		}
		if move.Direction == "install_candidate" {
			source, _ := customerServicePairedSwitchMediaMoveLogicalPaths(move)
			expected[source] = move.EntrySHA256
		}
	}
	if !reflect.DeepEqual(actual, expected) {
		return errors.New("product cleanup candidate or previous media changed; retained for inspection")
	}
	for _, rootName := range mediaRoots {
		if err := os.RemoveAll(manager.productCatalogPairedMediaVolumeDirectory(task, rootName)); err != nil {
			return err
		}
		if err := syncDirectory(manager.Configuration.StorageRoots[rootName]); err != nil {
			return err
		}
	}
	return nil
}

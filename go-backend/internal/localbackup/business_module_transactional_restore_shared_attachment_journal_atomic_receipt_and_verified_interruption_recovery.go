package localbackup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
)

type transactionalBusinessModuleRestoreContract struct {
	ModuleName, AttachmentStagePrefix, AuditResource, ActionSuffix      string
	VerifiedPhase, RestoredPhase, CancelledPhase, RecoveryRequiredPhase string
}

func transactionalBusinessModuleRestoreContractForTask(task *Task) transactionalBusinessModuleRestoreContract {
	if countTransactionalBusinessModuleRestoreProgressRecords(task) != 1 {
		return transactionalBusinessModuleRestoreContract{}
	}
	if task.ConfigurationModuleSiteContentRestore != nil {
		return transactionalBusinessModuleRestoreContract{ModuleName: "configuration", AttachmentStagePrefix: ".local-backup-configuration-content-", AuditResource: "configuration_module_site_content_restore", ActionSuffix: "configuration-site-content-formal-restore", VerifiedPhase: "configuration_site_content_result_verified_before_release", RestoredPhase: "configuration_site_content_restored_and_service_resumed", CancelledPhase: "uncommitted_configuration_site_content_cancelled_and_service_resumed", RecoveryRequiredPhase: "configuration_site_content_requires_verified_recovery"}
	}
	if task.MediaModuleRestore != nil {
		return transactionalBusinessModuleRestoreContract{ModuleName: "media", AttachmentStagePrefix: ".local-backup-media-restore-", AuditResource: "media_module_formal_restore", ActionSuffix: "media-module-formal-restore", VerifiedPhase: "media_restore_result_verified_before_service_release", RestoredPhase: "media_restored_and_service_resumed", CancelledPhase: "uncommitted_media_restore_cancelled_and_service_resumed", RecoveryRequiredPhase: "media_restore_requires_verified_recovery"}
	}
	if task.OrderModuleRestore != nil {
		return transactionalBusinessModuleRestoreContract{ModuleName: "orders", AttachmentStagePrefix: ".local-backup-order-history-", AuditResource: "order_module_history_restore", ActionSuffix: "order-module-formal-history-restore", VerifiedPhase: "order_history_restore_result_verified_before_service_release", RestoredPhase: "order_history_restored_and_service_resumed", CancelledPhase: "uncommitted_order_restore_cancelled_and_service_resumed", RecoveryRequiredPhase: "order_history_restore_requires_verified_recovery"}
	}
	if task.CustomerServiceModuleRestore != nil {
		return transactionalBusinessModuleRestoreContract{ModuleName: "customer_service", AttachmentStagePrefix: ".local-backup-customer-service-restore-", AuditResource: "customer_service_module_formal_restore", ActionSuffix: "customer-service-module-formal-restore", VerifiedPhase: "customer_service_module_restore_result_verified_before_service_release", RestoredPhase: "customer_service_module_restored_and_service_resumed", CancelledPhase: "uncommitted_customer_service_restore_cancelled_and_service_resumed", RecoveryRequiredPhase: "customer_service_module_restore_requires_verified_recovery"}
	}
	if task.CustomerProfileModuleRestore != nil {
		return transactionalBusinessModuleRestoreContract{ModuleName: "customers", AttachmentStagePrefix: ".local-backup-customer-profile-restore-", AuditResource: "customer_profile_module_formal_restore", ActionSuffix: "customer-profile-module-formal-restore", VerifiedPhase: "customer_profile_restore_result_verified_before_service_release", RestoredPhase: "customer_profiles_restored_and_service_resumed", CancelledPhase: "uncommitted_customer_profile_restore_cancelled_and_service_resumed", RecoveryRequiredPhase: "customer_profile_restore_requires_verified_recovery"}
	}
	return transactionalBusinessModuleRestoreContract{}
}

func transactionalBusinessModuleRestoreProgressForTask(task *Task) *BusinessModuleTransactionalRestoreProgress {
	if task != nil && countTransactionalBusinessModuleRestoreProgressRecords(task) == 1 && task.ConfigurationModuleSiteContentRestore != nil {
		return task.ConfigurationModuleSiteContentRestore
	}
	if countTransactionalBusinessModuleRestoreProgressRecords(task) != 1 {
		return nil
	}
	if task.MediaModuleRestore != nil {
		return task.MediaModuleRestore
	}
	if task.OrderModuleRestore != nil && task.CustomerServiceModuleRestore == nil {
		return task.OrderModuleRestore
	}
	if task.CustomerServiceModuleRestore != nil && task.OrderModuleRestore == nil {
		return task.CustomerServiceModuleRestore
	}
	if task.CustomerProfileModuleRestore != nil {
		return task.CustomerProfileModuleRestore
	}
	return nil
}

func countTransactionalBusinessModuleRestoreProgressRecords(task *Task) int {
	if task == nil {
		return 0
	}
	count := 0
	for _, progress := range []*BusinessModuleTransactionalRestoreProgress{task.OrderModuleRestore, task.CustomerServiceModuleRestore, task.CustomerProfileModuleRestore, task.MediaModuleRestore, task.ConfigurationModuleSiteContentRestore} {
		if progress != nil {
			count++
		}
	}
	return count
}

func taskHasAnyTransactionalBusinessModuleRestoreProgress(task *Task) bool {
	return countTransactionalBusinessModuleRestoreProgressRecords(task) > 0
}

type BusinessModuleTransactionalRestoreFile struct {
	Identity         FileIdentity  `json:"identity"`
	Started          bool          `json:"started"`
	Installed        bool          `json:"installed"`
	OriginalIdentity *FileIdentity `json:"original_identity,omitempty"`
}

// Only hashes, registered paths and operation identities enter the journal.
// Complete business rows remain solely in the authenticated encrypted package.
// Shared by supported order/chat/customer/media/configuration-content transactions. Exactly one module
// progress record is allowed; no whole-storefront/product switch journal or
// disposable database identity may coexist. Recover by SQL receipt, not by replay.
type BusinessModuleTransactionalRestoreProgress struct {
	ConfigurationContentProjectionReview *ConfigurationContentProjectionReview    `json:"configuration_content_projection_review,omitempty"`
	Version                              int                                      `json:"version"`
	TaskID                               string                                   `json:"task_id"`
	Operator                             string                                   `json:"operator"`
	Phase                                string                                   `json:"phase"`
	PreviewSHA256                        string                                   `json:"preview_sha256"`
	PackageSHA256                        string                                   `json:"package_sha256"`
	ConfigurationSHA256                  string                                   `json:"configuration_sha256"`
	RuntimeBoundarySHA256                string                                   `json:"runtime_boundary_sha256"`
	FenceSHA256                          string                                   `json:"fence_sha256"`
	DatabaseOID                          int64                                    `json:"database_oid"`
	BeforeDatabaseFingerprint            string                                   `json:"before_database_fingerprint"`
	AfterDatabaseFingerprint             string                                   `json:"after_database_fingerprint,omitempty"`
	BeforeMediaFingerprint               string                                   `json:"before_media_fingerprint"`
	AfterMediaFingerprint                string                                   `json:"after_media_fingerprint,omitempty"`
	Files                                []BusinessModuleTransactionalRestoreFile `json:"files"`
	CreatedDirectories                   []string                                 `json:"created_directories"`
	AuditID                              int64                                    `json:"audit_id,omitempty"`
	AuditReceipt                         string                                   `json:"audit_receipt,omitempty"`
	ChoiceSHA256                         string                                   `json:"choice_sha256,omitempty"`
	ReplacedRows                         int                                      `json:"replaced_rows,omitempty"`
	RestoredRows                         int                                      `json:"restored_rows"`
	Outcome                              string                                   `json:"outcome,omitempty"`
	StartedAt                            time.Time                                `json:"started_at"`
	VerifiedAt                           time.Time                                `json:"verified_at,omitempty"`
	ReleasedAt                           time.Time                                `json:"released_at,omitempty"`
	SHA256                               string                                   `json:"sha256"`
}

func transactionalBusinessModuleRestoreBytesSHA256(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func validateTransactionalBusinessModuleRestoreProgress(task *Task) error {
	contract := transactionalBusinessModuleRestoreContractForTask(task)
	if contract.ModuleName == "" || task == nil || transactionalBusinessModuleRestoreProgressForTask(task) == nil || task.Kind != "upload" || task.Manifest == nil || task.Manifest.Scope != selectedBusinessModulesBackupScope || !reflect.DeepEqual(task.Manifest.SelectedModules, []string{contract.ModuleName}) || task.Journal != nil || task.StageDatabase != "" || task.SelectedModuleDisposableStage != nil {
		return errors.New("module restore journal has an invalid or mixed scope")
	}
	p := transactionalBusinessModuleRestoreProgressForTask(task)
	copyProgress := *p
	copyProgress.SHA256 = ""
	digest, err := businessModuleRecoveryCanonicalJSONSHA256(copyProgress)
	if err != nil || digest != p.SHA256 || p.Version != 1 || p.TaskID != task.ID || !safeTaskID.MatchString(task.ID) || p.Operator == "" || p.PackageSHA256 != task.CiphertextSHA256 || p.PreviewSHA256 == "" || p.DatabaseOID <= 0 || p.BeforeDatabaseFingerprint == "" || p.BeforeMediaFingerprint == "" || p.StartedAt.IsZero() {
		return errors.New("module restore immutable identity or digest changed")
	}
	if p.RestoredRows < 0 || p.ReplacedRows < 0 || contract.ModuleName == "customer_service" && !moduleRecoverySideEffectReceiptRequestDigestIsValid(p.ChoiceSHA256) {
		return errors.New("module restore counts or saved choice identity are invalid")
	}
	switch p.Phase {
	case "installing_files", "committing_rows", "rows_committed", "releasing", "released":
	default:
		return errors.New("unsupported module restore journal phase")
	}
	if p.Phase == "committing_rows" || p.Phase == "rows_committed" {
		if p.AuditID <= 0 || p.AuditReceipt == "" || p.AfterDatabaseFingerprint == "" || p.AfterMediaFingerprint == "" {
			return errors.New("module restore commit intent is incomplete")
		}
	}
	if p.Phase == "releasing" || p.Phase == "released" {
		if p.VerifiedAt.IsZero() || (p.Outcome != "restored" && p.Outcome != "cancelled_before_commit") {
			return errors.New("module service release has no verified outcome")
		}
		if p.Outcome == "restored" && (p.AuditID <= 0 || p.AfterDatabaseFingerprint == "" || p.AfterMediaFingerprint == "") {
			return errors.New("module service release has no committed result")
		}
	}
	if p.Phase == "released" && p.ReleasedAt.IsZero() {
		return errors.New("module service release acknowledgement missing")
	}
	if err := validateConfigurationContentProjectionReviewBinding(task); err != nil {
		return err
	}
	manifestFiles := map[string]FileIdentity{}
	for _, entry := range task.Manifest.Files {
		manifestFiles[entry.Path] = entry
	}
	seen := map[string]bool{}
	for _, file := range p.Files {
		if err := validateMediaModuleByteRepairJournalOriginalIdentity(task, file); err != nil {
			return err
		}
		if _, _, err := transactionalBusinessModuleRestoreMediaPath(file.Identity.Path); err != nil {
			return err
		}
		if manifestFiles[file.Identity.Path] != file.Identity || seen[file.Identity.Path] || file.Installed && !file.Started {
			return errors.New("module restore attachment identity is not a unique package member")
		}
		seen[file.Identity.Path] = true
	}
	for _, directory := range p.CreatedDirectories {
		if _, _, err := transactionalBusinessModuleRestoreMediaPath(directory); err != nil {
			return err
		}
		owned := false
		for filename := range seen {
			owned = owned || strings.HasPrefix(filename, directory+"/")
		}
		if !owned {
			return errors.New("module restore directory is not an attachment ancestor")
		}
	}
	return nil
}

func taskHasUnresolvedTransactionalBusinessModuleRestore(task *Task) bool {
	return taskHasAnyTransactionalBusinessModuleRestoreProgress(task) && (validateTransactionalBusinessModuleRestoreProgress(task) != nil || transactionalBusinessModuleRestoreProgressForTask(task).Phase != "released")
}

func (manager *Manager) saveTransactionalBusinessModuleRestoreProgress(task *Task) error {
	p := transactionalBusinessModuleRestoreProgressForTask(task)
	if p == nil {
		return errors.New("module restore journal has an invalid or mixed scope")
	}
	p.SHA256 = ""
	var err error
	p.SHA256, err = businessModuleRecoveryCanonicalJSONSHA256(p)
	if err != nil {
		return err
	}
	return manager.saveTask(task)
}

func transactionalBusinessModuleRestoreMediaPath(packagePath string) (string, string, error) {
	parts := strings.SplitN(packagePath, "/", 3)
	if len(parts) != 3 || parts[0] != "media" {
		return "", "", errors.New("module restore only installs registered media files")
	}
	if _, err := customerServiceCandidateMediaPathComponents(parts[1], parts[2]); err != nil {
		return "", "", err
	}
	return parts[1], parts[2], nil
}

func (manager *Manager) installTransactionalBusinessModuleRestoreFiles(ctx context.Context, task *Task, extracted string, checkpoint func(string) error) error {
	p := transactionalBusinessModuleRestoreProgressForTask(task)
	if err := manager.prepareOwnedTransactionalBusinessModuleAttachmentStaging(task); err != nil {
		return err
	}
	for index := range p.Files {
		file := &p.Files[index]
		rootName, key, _ := transactionalBusinessModuleRestoreMediaPath(file.Identity.Path)
		directory := manager.Configuration.StorageRoots[rootName]
		stageKey := filepath.Join(transactionalBusinessModuleRestoreContractForTask(task).AttachmentStagePrefix+task.ID, "files", filepath.FromSlash(key))
		mode := os.FileMode(0644)
		directoryMode := os.FileMode(0755)
		if rootName == "uploads_private" {
			mode, directoryMode = 0600, 0700
		}
		if err := copyCustomerServiceCandidateBoundedVerifiedMediaFile(ctx, extracted, file.Identity.Path, directory, filepath.ToSlash(stageKey), file.Identity, mode); err != nil {
			return err
		}
		for parent := filepath.Dir(stageKey); parent != "."; parent = filepath.Dir(parent) {
			if err := syncDirectory(filepath.Join(directory, parent)); err != nil {
				return err
			}
		}
		if err := syncDirectory(directory); err != nil {
			return err
		}
		root, err := os.OpenRoot(directory)
		if err != nil {
			return err
		}
		parent := filepath.Dir(filepath.FromSlash(key))
		components := strings.Split(filepath.ToSlash(parent), "/")
		partial := ""
		for _, component := range components {
			if component == "." {
				continue
			}
			partial = filepath.Join(partial, component)
			info, statErr := root.Lstat(partial)
			if errors.Is(statErr, os.ErrNotExist) {
				p.CreatedDirectories = append(p.CreatedDirectories, "media/"+rootName+"/"+filepath.ToSlash(partial))
				if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
					_ = root.Close()
					return err
				}
				if err := root.Mkdir(partial, directoryMode); err != nil {
					_ = root.Close()
					return err
				}
				if err := syncDirectory(filepath.Join(directory, filepath.Dir(partial))); err != nil {
					_ = root.Close()
					return err
				}
			} else if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				_ = root.Close()
				return errors.New("module attachment parent is unsafe")
			}
		}
		file.Started = true
		if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
			_ = root.Close()
			return err
		}
		if file.OriginalIdentity != nil {
			_ = root.Close()
			if err := manager.installMediaModuleRegisteredFileRepairWithOriginalHardLink(task, file, checkpoint); err != nil {
				return err
			}
			file.Installed = true
			if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
				return err
			}
			continue
		}
		if err := root.Link(stageKey, filepath.FromSlash(key)); err != nil {
			_ = root.Close()
			return err
		}
		_ = root.Close()
		if err := syncDirectory(filepath.Join(directory, parent)); err != nil {
			return err
		}
		if checkpoint != nil {
			if err := checkpoint("attachment_linked_before_acknowledgement"); err != nil {
				return err
			}
		}
		file.Installed = true
		if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
			return err
		}
	}
	return nil
}

func (manager *Manager) verifyTransactionalBusinessModuleRestoreFenceAndRuntime(task *Task) error {
	if err := validateTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return err
	}
	p := transactionalBusinessModuleRestoreProgressForTask(task)
	configurationSHA, err := businessModuleRecoveryCanonicalJSONSHA256(manager.Configuration)
	if err != nil || configurationSHA != p.ConfigurationSHA256 {
		return errors.New("module restore configuration changed")
	}
	if err := manager.Configuration.validateRuntimeBoundary(); err != nil {
		return err
	}
	for _, entry := range []struct{ filename, digest string }{{maintenancegate.FenceFilename, p.FenceSHA256}, {"application-boundary.json", p.RuntimeBoundarySHA256}} {
		data, err := os.ReadFile(filepath.Join(manager.Configuration.MaintenanceDirectory, entry.filename))
		if err != nil || transactionalBusinessModuleRestoreBytesSHA256(data) != entry.digest {
			return errors.New("module restore maintenance window or application boundary changed")
		}
	}
	return nil
}

func (manager *Manager) completeOrCancelInterruptedTransactionalBusinessModuleRestoreWithHeldLease(ctx context.Context, task *Task, live *sql.DB, checkpoint func(string) error) error {
	if err := manager.verifyTransactionalBusinessModuleRestoreFenceAndRuntime(task); err != nil {
		return err
	}
	if err := assertNoUnfencedDatabaseSessions(ctx, live, manager.Configuration.DatabaseName); err != nil {
		return err
	}
	p := transactionalBusinessModuleRestoreProgressForTask(task)
	var oid int64
	if err := live.QueryRowContext(ctx, `SELECT oid FROM pg_database WHERE datname=current_database()`).Scan(&oid); err != nil || oid != p.DatabaseOID {
		return errors.New("module restore database identity changed")
	}
	var receipts int
	if p.AuditID > 0 {
		if err := live.QueryRowContext(ctx, `SELECT count(*) FROM public.audit_logs WHERE id=$1 AND username=$2 AND action='restore' AND resource=$6 AND method='POST' AND path=$3 AND status='success' AND new_value=$4 AND created_at=$5`, p.AuditID, p.Operator, "/api/tasks/"+task.ID+"/"+transactionalBusinessModuleRestoreContractForTask(task).ActionSuffix, p.AuditReceipt, p.StartedAt, transactionalBusinessModuleRestoreContractForTask(task).AuditResource).Scan(&receipts); err != nil {
			return err
		}
	}
	_, fingerprint, _, err := captureDatabaseInventory(ctx, live)
	if err != nil {
		return err
	}
	if receipts == 1 {
		if fingerprint != p.AfterDatabaseFingerprint {
			return errors.New("committed module restore differs from its durable result; retain maintenance")
		}
		for _, file := range p.Files {
			if !file.Started || !file.Installed {
				return errors.New("committed module attachment has no complete installation acknowledgement")
			}
			rootName, key, _ := transactionalBusinessModuleRestoreMediaPath(file.Identity.Path)
			identity, err := fileIdentity(filepath.Join(manager.Configuration.StorageRoots[rootName], filepath.FromSlash(key)), file.Identity.Path)
			if err != nil || identity != file.Identity {
				return errors.New("committed module attachment changed")
			}
		}
		media, err := manager.Configuration.mediaFingerprint(ctx)
		if err != nil || media != p.AfterMediaFingerprint {
			return errors.New("committed module media changed; retain maintenance")
		}
		p.Outcome = "restored"
	} else {
		if fingerprint != p.BeforeDatabaseFingerprint {
			return errors.New("module commit receipt missing or original database changed; retain maintenance")
		}
		if err := manager.removeUncommittedTransactionalBusinessModuleRestoreFileAdditions(task); err != nil {
			return err
		}
		media, err := manager.Configuration.mediaFingerprint(ctx)
		if err != nil || media != p.BeforeMediaFingerprint {
			return errors.New("uncommitted module media cannot be verified as restored")
		}
		p.Outcome = "cancelled_before_commit"
		p.RestoredRows, p.ReplacedRows = 0, 0
	}
	if err := manager.removeOwnedTransactionalBusinessModuleAttachmentStaging(task); err != nil {
		return err
	}
	if p.Outcome == "restored" && task.ConfigurationModuleSiteContentRestore != nil && p.ConfigurationContentProjectionReview == nil {
		p.Phase = "rows_committed"
		task.State, task.Phase = "configuration_site_content_restored_still_isolated", "configuration_site_content_requires_projection_review"
		task.Error = ""
		return manager.saveTransactionalBusinessModuleRestoreProgress(task)
	}
	p.VerifiedAt = time.Now().UTC()
	p.Phase = "releasing"
	task.Phase = transactionalBusinessModuleRestoreContractForTask(task).VerifiedPhase
	if err := manager.saveTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("verified_before_fence_removal"); err != nil {
			return err
		}
	}
	if err := manager.verifyTransactionalBusinessModuleRestoreFenceAndRuntime(task); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(manager.Configuration.MaintenanceDirectory, maintenancegate.FenceFilename)); err != nil {
		return err
	}
	if err := syncDirectory(manager.Configuration.MaintenanceDirectory); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint("fence_removed_before_acknowledgement"); err != nil {
			return err
		}
	}
	return manager.acknowledgeTransactionalBusinessModuleRestoreServiceRelease(task)
}

func (manager *Manager) removeUncommittedTransactionalBusinessModuleRestoreFileAdditions(task *Task) error {
	p := transactionalBusinessModuleRestoreProgressForTask(task)
	for index := len(p.Files) - 1; index >= 0; index-- {
		file := p.Files[index]
		if !file.Started {
			continue
		}
		if file.OriginalIdentity != nil {
			if err := manager.restoreOriginalMediaModuleFileAfterUncommittedByteRepair(task, file); err != nil {
				return err
			}
			continue
		}
		rootName, key, _ := transactionalBusinessModuleRestoreMediaPath(file.Identity.Path)
		directory := manager.Configuration.StorageRoots[rootName]
		root, err := os.OpenRoot(directory)
		if err != nil {
			return err
		}
		installed, err := root.Lstat(filepath.FromSlash(key))
		if errors.Is(err, os.ErrNotExist) {
			_ = root.Close()
			continue
		}
		if err != nil {
			_ = root.Close()
			return err
		}
		staged, err := root.Lstat(filepath.Join(transactionalBusinessModuleRestoreContractForTask(task).AttachmentStagePrefix+task.ID, "files", filepath.FromSlash(key)))
		if err != nil || !installed.Mode().IsRegular() || !staged.Mode().IsRegular() || !os.SameFile(installed, staged) {
			_ = root.Close()
			return errors.New("uncommitted module attachment has a changed physical identity")
		}
		if err := root.Remove(filepath.FromSlash(key)); err != nil {
			_ = root.Close()
			return err
		}
		_ = root.Close()
		if err := syncDirectory(filepath.Join(directory, filepath.Dir(filepath.FromSlash(key)))); err != nil {
			return err
		}
	}
	for index := len(p.CreatedDirectories) - 1; index >= 0; index-- {
		rootName, key, err := transactionalBusinessModuleRestoreMediaPath(p.CreatedDirectories[index])
		if err != nil {
			return err
		}
		directory := manager.Configuration.StorageRoots[rootName]
		root, err := os.OpenRoot(directory)
		if err != nil {
			return err
		}
		err = root.Remove(filepath.FromSlash(key))
		_ = root.Close()
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err == nil {
			if err := syncDirectory(filepath.Join(directory, filepath.Dir(filepath.FromSlash(key)))); err != nil {
				return err
			}
		}
	}
	return nil
}

func (manager *Manager) acknowledgeTransactionalBusinessModuleRestoreServiceRelease(task *Task) error {
	if err := validateTransactionalBusinessModuleRestoreProgress(task); err != nil {
		return err
	}
	p := transactionalBusinessModuleRestoreProgressForTask(task)
	if p.Phase != "releasing" && p.Phase != "released" {
		return errors.New("module service release intent missing")
	}
	p.Phase = "released"
	if p.ReleasedAt.IsZero() {
		p.ReleasedAt = time.Now().UTC()
	}
	task.Error = ""
	if p.Outcome == "restored" {
		task.State, task.Phase = "completed", transactionalBusinessModuleRestoreContractForTask(task).RestoredPhase
	} else {
		task.State, task.Phase = "rolled_back", transactionalBusinessModuleRestoreContractForTask(task).CancelledPhase
	}
	return manager.saveTransactionalBusinessModuleRestoreProgress(task)
}

func (manager *Manager) retainInterruptedTransactionalBusinessModuleRestoreAfterRestart(task *Task) error {
	if validateTransactionalBusinessModuleRestoreProgress(task) == nil {
		p := transactionalBusinessModuleRestoreProgressForTask(task)
		if p.Phase == "released" {
			return nil
		}
		if p.Phase == "releasing" {
			fenced, err := maintenancegate.IsApplicationFenced(manager.Configuration.MaintenanceDirectory)
			if err == nil && !fenced {
				return manager.acknowledgeTransactionalBusinessModuleRestoreServiceRelease(task)
			}
		}
	}
	task.State, task.Phase, task.Error = "interrupted", transactionalBusinessModuleRestoreContractForTask(task).RecoveryRequiredPhase, "module restore retains maintenance; recover its atomic commit receipt and attachment installation before service resumption"
	return manager.saveTask(task)
}

func transactionalBusinessModuleAttachmentStagingOwnershipBytes(task *Task, rootName string) []byte {
	data, _ := json.Marshal(struct{ TaskID, PackageSHA256, StorageRoot string }{task.ID, transactionalBusinessModuleRestoreProgressForTask(task).PackageSHA256, rootName})
	return data
}

func (manager *Manager) prepareOwnedTransactionalBusinessModuleAttachmentStaging(task *Task) error {
	roots := map[string]bool{}
	for _, file := range transactionalBusinessModuleRestoreProgressForTask(task).Files {
		rootName, _, err := transactionalBusinessModuleRestoreMediaPath(file.Identity.Path)
		if err != nil {
			return err
		}
		roots[rootName] = true
	}
	for rootName := range roots {
		directory := manager.Configuration.StorageRoots[rootName]
		root, err := os.OpenRoot(directory)
		if err != nil {
			return err
		}
		stage := transactionalBusinessModuleRestoreContractForTask(task).AttachmentStagePrefix + task.ID
		if _, err := root.Lstat(stage); !errors.Is(err, os.ErrNotExist) {
			_ = root.Close()
			return errors.New("module attachment staging path already exists or cannot be inspected")
		}
		marker, err := root.OpenFile(stage+".owner.json", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			_ = root.Close()
			return err
		}
		_, writeErr := marker.Write(transactionalBusinessModuleAttachmentStagingOwnershipBytes(task, rootName))
		err = errors.Join(writeErr, marker.Sync(), marker.Close())
		if err == nil {
			err = syncDirectory(directory)
		}
		if err == nil {
			err = root.Mkdir(stage, 0700)
		}
		if err == nil {
			err = syncDirectory(directory)
		}
		_ = root.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (manager *Manager) removeOwnedTransactionalBusinessModuleAttachmentStaging(task *Task) error {
	registered := map[string]map[string]bool{}
	stage := transactionalBusinessModuleRestoreContractForTask(task).AttachmentStagePrefix + task.ID
	for _, file := range transactionalBusinessModuleRestoreProgressForTask(task).Files {
		rootName, key, err := transactionalBusinessModuleRestoreMediaPath(file.Identity.Path)
		if err != nil {
			return err
		}
		if registered[rootName] == nil {
			registered[rootName] = map[string]bool{}
		}
		filename := stage + "/files/" + key
		registered[rootName][filename] = false
		for _, repairDirectory := range mediaModuleByteRepairAdditionalStagingDirectoryNames(file) {
			repairFilename := stage + "/" + repairDirectory + "/" + key
			registered[rootName][repairFilename] = false
			for parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(repairFilename))); parent != "."; parent = filepath.ToSlash(filepath.Dir(filepath.FromSlash(parent))) {
				registered[rootName][parent] = true
			}
		}
		for parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(filename))); parent != "."; parent = filepath.ToSlash(filepath.Dir(filepath.FromSlash(parent))) {
			registered[rootName][parent] = true
		}
	}
	for rootName, expected := range registered {
		directory := manager.Configuration.StorageRoots[rootName]
		root, err := os.OpenRoot(directory)
		if err != nil {
			return err
		}
		markerName := stage + ".owner.json"
		marker, markerErr := root.ReadFile(markerName)
		_, stageErr := root.Lstat(stage)
		if errors.Is(markerErr, os.ErrNotExist) && errors.Is(stageErr, os.ErrNotExist) {
			_ = root.Close()
			continue
		}
		markerInfo, statErr := root.Lstat(markerName)
		if markerErr != nil || statErr != nil || !markerInfo.Mode().IsRegular() || string(marker) != string(transactionalBusinessModuleAttachmentStagingOwnershipBytes(task, rootName)) {
			_ = root.Close()
			return errors.New("module attachment staging ownership changed")
		}
		var entries []string
		if !errors.Is(stageErr, os.ErrNotExist) {
			err = fs.WalkDir(root.FS(), stage, func(name string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				isDirectory, known := expected[name]
				if !known || entry.Type()&os.ModeSymlink != 0 || isDirectory != entry.IsDir() || !entry.IsDir() && !entry.Type().IsRegular() {
					return errors.New("module attachment staging contains an unregistered entry")
				}
				entries = append(entries, name)
				return nil
			})
			if err != nil {
				_ = root.Close()
				return err
			}
		}
		sort.Slice(entries, func(i, j int) bool { return len(entries[i]) > len(entries[j]) })
		for _, name := range entries {
			if err := root.Remove(filepath.FromSlash(name)); err != nil {
				_ = root.Close()
				return err
			}
			if err := syncDirectory(filepath.Join(directory, filepath.Dir(filepath.FromSlash(name)))); err != nil {
				_ = root.Close()
				return err
			}
		}
		err = root.Remove(markerName)
		_ = root.Close()
		if err != nil {
			return err
		}
		if err := syncDirectory(directory); err != nil {
			return err
		}
	}
	return nil
}

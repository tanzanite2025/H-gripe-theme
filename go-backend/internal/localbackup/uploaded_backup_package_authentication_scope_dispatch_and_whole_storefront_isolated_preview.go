package localbackup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
)

// Authenticate uploads, dispatch existing package scopes and rehearse whole-storefront restoration without expanding module eligibility.
func (manager *Manager) VerifyUploadedPackage(id string) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if taskHasAnyTransactionalBusinessModuleRestoreProgress(task) {
		return nil, errors.New("module restore journal exists; use its dedicated recovery")
	}
	if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
		return nil, errors.New("resume or cancel the stopped product handoff preparation before package verification")
	}
	if task.Kind != "upload" || task.UploadedBytes != task.Size || (task.State != "uploaded" && task.State != "failed" && task.State != "interrupted") {
		return nil, errors.New("package upload is incomplete or not eligible for validation")
	}
	if task.Journal != nil {
		return nil, errors.New("a recovery journal already exists; use rollback or reconciliation")
	}
	if task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
		return nil, errors.New("resume the stopped product view drain and candidate rebuild first")
	}
	if err := manager.startTaskOperation(task, "authenticating_and_extracting", manager.verifyAndPreview); err != nil {
		return nil, err
	}
	return manager.GetTask(task.ID)
}

func (manager *Manager) verifyAndPreview(ctx context.Context, task *Task) error {
	cfg := manager.Configuration
	if err := manager.cleanupSelectedBusinessModuleDisposableStageByPersistedIdentity(ctx, task); err != nil {
		return err
	}
	task.SelectedModuleRestorePreview = nil
	task.SelectedCustomerServiceRestoreChoice = nil
	task.SelectedCustomerServiceCandidateRehearsal = nil
	if err := cfg.checkWorkspaceCapacity(64 << 20); err != nil {
		return err
	}
	directory, _ := manager.taskDirectory(task.ID)
	extracted := filepath.Join(directory, "verified")
	if err := os.RemoveAll(extracted); err != nil {
		return err
	}
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	if err != nil {
		return err
	}
	filename, _ := manager.taskPackagePath(task.ID)
	identity, err := fileIdentity(filename, "")
	if err != nil || identity.Size != task.Size {
		return errors.New("uploaded ciphertext length changed")
	}
	task.CiphertextSHA256 = identity.SHA256
	manifest, err := extractAndVerifyEncryptedPackageWithAllowedScopes(filename, extracted, passphrase, cfg.MaxExpandedBytes, cfg.MaxFiles, true, cfg.checkWorkspaceCapacity)
	if err != nil {
		return err
	}
	if manifest.SiteID != cfg.SiteID || manifest.Environment != cfg.Environment || manifest.ApplicationVersion != cfg.ApplicationVersion || manifest.MigrationVersion != cfg.ExpectedMigrationVersion {
		return errors.New("package site, environment, application or migration version is incompatible")
	}
	task.Manifest = &manifest
	if len(manifest.Gaps) > 0 {
		return errors.New("package contains unresolved attachment or media gaps; overwrite is blocked")
	}
	if manifest.Scope == productCatalogBackupScope {
		return manager.restoreProductCatalogIntoIsolatedTargetClone(ctx, task, extracted)
	}
	if manifest.Scope == selectedBusinessModulesBackupScope {
		return manager.validateUploadedSelectedBusinessModulePackage(task, extracted)
	}
	if err := cfg.validateWholeApplicationCompanyStaticPackage(manifest, extracted); err != nil {
		return err
	}
	materials, err := readRecoveryMaterials(extracted)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(materials, cfg.RecoverySecrets) {
		return errors.New("recovery signing/encryption keys do not match the target; provision the original keys separately")
	}
	task.Phase = "restoring_isolated_database"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if task.StageDatabase != "" {
		if err := cfg.dropOwnedStageDatabase(ctx, task.StageDatabase); err != nil {
			return err
		}
	}
	task.StageDatabase = "local_backup_stage_" + task.ID[:16]
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if err := cfg.restoreIsolatedDatabase(ctx, extracted, task.StageDatabase, manifest); err != nil {
		return err
	}
	staged, err := cfg.connectDatabase(ctx, task.StageDatabase)
	if err != nil {
		return err
	}
	_, task.StageFingerprint, _, err = captureDatabaseInventory(ctx, staged)
	_ = staged.Close()
	if err != nil {
		return err
	}
	task.PreviewTargetFingerprint, task.PreviewTargetTables, err = cfg.targetFingerprint(ctx)
	if err != nil {
		return err
	}
	task.State = "verified"
	task.Phase = "isolated_restore_verified_overwrite_preview_ready"
	return manager.saveTask(task)
}

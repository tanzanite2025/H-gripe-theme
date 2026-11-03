package localbackup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Capture the whole storefront under its existing maintenance window and generate the authenticated local package.
func (manager *Manager) CreateBackup(actor, confirmation string) (*Task, error) {
	cfg := manager.Configuration
	if confirmation != "BACKUP "+cfg.SiteID+"/"+cfg.Environment {
		return nil, errors.New("explicit confirmation of the storefront maintenance window is required")
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	preflight, cancelPreflight := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelPreflight()
	required, err := cfg.estimateBackupWorkspaceBytes(preflight)
	if err != nil {
		return nil, err
	}
	if err := cfg.checkWorkspaceCapacity(required); err != nil {
		return nil, err
	}
	task, err := manager.createTask("backup", actor)
	if err != nil {
		return nil, err
	}
	if err := manager.startTaskOperation(task, "waiting_for_application_stop", manager.generateBackup); err != nil {
		task.State = "failed"
		task.Error = err.Error()
		_ = manager.saveTask(task)
		return nil, err
	}
	return manager.GetTask(task.ID)
}

func (manager *Manager) generateBackup(ctx context.Context, task *Task) (resultErr error) {
	cfg := manager.Configuration
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	// No live data is changed by generation. Successful or failed capture can
	// release maintenance; restore and switch failures deliberately cannot.
	defer func() { resultErr = errors.Join(resultErr, manager.releaseMaintenance(task)) }()
	directory, _ := manager.taskDirectory(task.ID)
	snapshot := filepath.Join(directory, "snapshot")
	if err := os.Mkdir(snapshot, 0700); err != nil {
		return err
	}
	defer func(cleanupPath string) { _ = os.RemoveAll(cleanupPath) }(snapshot)
	manifest := Manifest{FormatVersion: PackageVersion, PackageID: task.ID, SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, CapturedAt: time.Now().UTC(), Scope: "whole_application", Files: []FileIdentity{}, Gaps: []string{}}
	task.Phase = "capturing_consistent_database_snapshot"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if err := cfg.captureSnapshotAndDump(ctx, snapshot, &manifest); err != nil {
		return err
	}
	if manifest.MigrationVersion != cfg.ExpectedMigrationVersion {
		return errors.New("application migration does not match configured backup contract")
	}
	task.Phase = "copying_and_verifying_three_media_roots"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	if err := cfg.copyMediaSnapshot(ctx, snapshot, &manifest); err != nil {
		return err
	}
	if err := cfg.copyWholeApplicationCompanyStaticSnapshot(ctx, snapshot, &manifest); err != nil {
		return err
	}
	db, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	err = cfg.findMediaReferenceGaps(ctx, db, &manifest)
	if err == nil {
		err = findBusinessRelationshipGaps(ctx, db, &manifest)
	}
	if err == nil {
		_, fingerprint, _, inventoryErr := captureDatabaseInventory(ctx, db)
		err = inventoryErr
		if err == nil && fingerprint != manifest.DatabaseFingerprint {
			err = errors.New("database changed despite the maintenance fence; backup rejected")
		}
	}
	_ = db.Close()
	if err != nil {
		return err
	}
	if err := writeDurableJSON(filepath.Join(snapshot, "recovery-secrets.json"), cfg.RecoverySecrets); err != nil {
		return err
	}
	for _, name := range []string{"database.dump", "recovery-secrets.json"} {
		entry, err := fileIdentity(filepath.Join(snapshot, name), name)
		if err != nil {
			return err
		}
		manifest.Files = append(manifest.Files, entry)
	}
	cfg.sortedManifestFiles(&manifest)
	liveMediaFingerprint, err := cfg.mediaFingerprint(ctx)
	if err != nil {
		return err
	}
	if liveMediaFingerprint != fingerprintManifestMedia(manifest) {
		return errors.New("media changed despite maintenance isolation; package rejected")
	}
	if wholeApplicationPackageIncludesCompanyStatic(&manifest) {
		fingerprint, err := cfg.companyStaticReleaseFileFingerprint(ctx)
		if err != nil || fingerprint != fingerprintCompanyStaticManifestFiles(manifest.Files) {
			return errors.New("company static files changed during backup; package rejected")
		}
	}
	// The copied files are now independent of live roots. Encryption can run
	// after capture, but retaining the fence also protects against rogue writers.
	task.Phase = "encrypting_package"
	if err := manager.saveTask(task); err != nil {
		return err
	}
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	if err != nil {
		return err
	}
	destination, _ := manager.taskPackagePath(task.ID)
	if err := WriteEncryptedPackage(destination, snapshot, manifest, passphrase, cfg.MaxPackageBytes); err != nil {
		return err
	}
	identity, err := fileIdentity(destination, "")
	if err != nil {
		return err
	}
	task.Manifest = &manifest
	task.Size = identity.Size
	task.CiphertextSHA256 = identity.SHA256
	task.State = "ready_to_download"
	task.Phase = "package_generated"
	if len(manifest.Gaps) > 0 {
		task.State = "generated_with_gaps"
		task.Phase = "package_generated_restore_blocked"
	}
	return manager.saveTask(task)
}

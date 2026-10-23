package localbackup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModuleDependencyRegistryAcceptsOnlyAuditedDescriptionMigrationWithoutEnablingModuleRecovery(t *testing.T) {
	for _, version := range []int64{433, 434, 435, 436, 437} {
		t.Run(fmt.Sprintf("migration_%d", version), func(t *testing.T) {
			contract := moduleRecoveryDependencyContract{Name: "orders", OwnedTables: []string{"orders"}}
			observation := &moduleRecoveryDatabaseObservation{Migration: version, Tables: map[string]int64{"orders": 0}}
			report := buildModuleRecoveryDependencyReport(contract, &Task{Manifest: &Manifest{}}, observation, observation)
			codes := collectModuleDependencyReportBlockerCodes(report)
			if version == 434 || version == 435 || version == 436 {
				require.NotContains(t, codes, "module_registry_migration_incompatible:")
			} else {
				require.Contains(t, codes, "module_registry_migration_incompatible:")
			}
			require.Equal(t, 8, report.RegistryVersion)
			require.Equal(t, int64(434), report.RegistryMigrationVersion, "legacy plan/journal baseline must remain unchanged")
			require.Contains(t, codes, "module_recovery_not_enabled:")
			require.False(t, report.RestoreAvailable)
		})
	}
}

func TestAuditedMigrationCompatibilityDoesNotPermitCrossVersionPackageUploads(t *testing.T) {
	for _, scope := range []string{"whole_application", productCatalogBackupScope} {
		for _, versions := range []struct{ source, target int64 }{{434, 435}, {434, 436}, {435, 436}} {
			t.Run(fmt.Sprintf("%s/source_%d_target_%d", scope, versions.source, versions.target), func(t *testing.T) {
				cfg := createIndependentRecoveryTestConfiguration(t)
				cfg.ExpectedMigrationVersion = versions.target
				manager, err := NewManager(cfg)
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, manager.Close()) })
				manifest := Manifest{FormatVersion: PackageVersion, PackageID: newTaskID(), Scope: scope, SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: versions.source, Files: []FileIdentity{}}
				filename := filepath.Join(t.TempDir(), "synthetic-migration-434-backup.age")
				sourceDirectory := t.TempDir()
				files := map[string]string{productCatalogRowsFilename: `{}`}
				if scope == "whole_application" {
					files = map[string]string{"database.dump": "synthetic unexecuted dump", "recovery-secrets.json": `{}`}
				}
				for name, contents := range files {
					path := filepath.Join(sourceDirectory, name)
					require.NoError(t, os.WriteFile(path, []byte(contents), 0600))
					identity, err := fileIdentity(path, name)
					require.NoError(t, err)
					manifest.Files = append(manifest.Files, identity)
				}
				passphrase, err := readProtectedCredential(cfg.PassphraseFile)
				require.NoError(t, err)
				require.NoError(t, WriteEncryptedPackage(filename, sourceDirectory, manifest, passphrase, cfg.MaxPackageBytes))
				contents, err := os.ReadFile(filename)
				require.NoError(t, err)
				task, err := manager.BeginUpload("synthetic-operator", int64(len(contents)))
				require.NoError(t, err)
				digest := sha256.Sum256(contents)
				_, err = manager.AppendUploadChunk(task.ID, 0, contents, hex.EncodeToString(digest[:]))
				require.NoError(t, err)
				_, err = manager.CommitUpload(task.ID)
				require.NoError(t, err)
				_, err = manager.VerifyUploadedPackage(task.ID)
				require.NoError(t, err)
				verified := waitForIndependentRecoveryTask(t, manager, task.ID)
				require.Equal(t, "failed", verified.State)
				require.Contains(t, verified.Error, "package site, environment, application or migration version is incompatible")
				require.Empty(t, verified.StageDatabase, "version rejection precedes any isolated database restore")
				require.Nil(t, verified.Journal)
			})
		}
	}
}

func TestProductViewCountReplayReceiptTableIsRegisteredAndPreservedOutsideEveryModuleOwnedSet(t *testing.T) {
	const receiptTable = "product_view_count_flush_receipts"
	require.True(t, registeredStorefrontModuleRecoveryTableNames()[receiptTable])
	require.False(t, productCatalogAllowedTableNames()[receiptTable])
	for _, contract := range registeredOrderAndCustomerServiceRecoveryDependencyContracts() {
		require.NotContains(t, contract.OwnedTables, receiptTable)
	}
	require.Equal(t, 8, moduleDependencyRegistryVersion)
	require.Equal(t, int64(434), moduleDependencyRegistryMigrationVersion)
}

package localbackup

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigurationContentRestoreRejectsMixedScopeCredentialsAndUnreviewedServiceRelease(t *testing.T) {
	for _, scenario := range []string{"content_missing", "credential_missing", "mixed_modules", "mixed_journals", "review_wrong_result", "review_before_commit", "release_without_review", "valid_committed_review", "other_module_with_configuration_review", "configuration_file_installation"} {
		t.Run(scenario, func(t *testing.T) {
			configurationSnapshot := selectedBusinessModuleCompleteSnapshot{Modules: []string{"configuration"}}
			if scenario == "content_missing" {
				require.True(t, configurationModuleHasMissingSiteContentWriteContract(configurationSnapshot, "posts"))
				return
			}
			if scenario == "credential_missing" {
				require.False(t, configurationModuleHasMissingSiteContentWriteContract(configurationSnapshot, "payment_methods"))
				require.False(t, configurationModuleHasMissingSiteContentWriteContract(configurationSnapshot, "settings"))
				return
			}
			if scenario == "mixed_modules" {
				configurationSnapshot.Modules = []string{"configuration", "orders"}
				require.False(t, configurationModuleHasMissingSiteContentWriteContract(configurationSnapshot, "posts"))
				return
			}
			task := &Task{ID: newTaskID(), Kind: "upload", CiphertextSHA256: strings.Repeat("b", 64), Manifest: &Manifest{Scope: selectedBusinessModulesBackupScope, SelectedModules: []string{"configuration"}}}
			progress := &BusinessModuleTransactionalRestoreProgress{Version: 1, TaskID: task.ID, Operator: "synthetic-operator", Phase: "rows_committed", PreviewSHA256: strings.Repeat("a", 64), PackageSHA256: task.CiphertextSHA256, DatabaseOID: 42, BeforeDatabaseFingerprint: "before-db", BeforeMediaFingerprint: "before-media", AfterDatabaseFingerprint: "after-db", AfterMediaFingerprint: "after-media", AuditID: 42, AuditReceipt: "synthetic atomic receipt", StartedAt: time.Now().UTC(), RestoredRows: 12}
			task.ConfigurationModuleSiteContentRestore = progress
			review := &ConfigurationContentProjectionReview{Operator: "synthetic-reviewer", Reference: "isolated-content-and-cache-review", AfterDatabaseFingerprint: progress.AfterDatabaseFingerprint, AfterMediaFingerprint: progress.AfterMediaFingerprint, ReviewedAt: time.Now().UTC()}
			switch scenario {
			case "mixed_journals":
				task.OrderModuleRestore = &BusinessModuleTransactionalRestoreProgress{}
			case "review_wrong_result":
				progress.ConfigurationContentProjectionReview = review
				review.AfterDatabaseFingerprint = "other-result"
			case "review_before_commit":
				progress.ConfigurationContentProjectionReview = review
				progress.Phase = "installing_files"
			case "release_without_review":
				progress.Phase = "releasing"
				progress.Outcome = "restored"
				progress.VerifiedAt = time.Now().UTC()
			case "valid_committed_review":
				progress.ConfigurationContentProjectionReview = review
			case "other_module_with_configuration_review":
				progress.ConfigurationContentProjectionReview = review
				task.ConfigurationModuleSiteContentRestore = nil
				task.MediaModuleRestore = progress
				task.Manifest.SelectedModules = []string{"media"}
			case "configuration_file_installation":
				progress.Files = []BusinessModuleTransactionalRestoreFile{{Identity: FileIdentity{Path: "media/uploads/unauthorized.png"}}}
			}
			progress.SHA256 = ""
			var err error
			progress.SHA256, err = businessModuleRecoveryCanonicalJSONSHA256(progress)
			require.NoError(t, err)
			if scenario == "valid_committed_review" {
				require.NoError(t, validateTransactionalBusinessModuleRestoreProgress(task))
			} else {
				require.Error(t, validateTransactionalBusinessModuleRestoreProgress(task))
			}
		})
	}
}

func TestConfigurationContentOptionalFieldsPreserveLegacyJournalDigestAndModuleDispatch(t *testing.T) {
	legacy := `{"version":1,"task_id":"synthetic","operator":"old-operator","phase":"installing_files","preview_sha256":"preview","package_sha256":"package","configuration_sha256":"config","runtime_boundary_sha256":"runtime","fence_sha256":"fence","database_oid":42,"before_database_fingerprint":"db","before_media_fingerprint":"media","files":[],"created_directories":[],"restored_rows":0,"started_at":"2026-10-01T00:00:00Z","verified_at":"0001-01-01T00:00:00Z","released_at":"0001-01-01T00:00:00Z","sha256":""}`
	var progress BusinessModuleTransactionalRestoreProgress
	require.NoError(t, json.Unmarshal([]byte(legacy), &progress))
	encoded, err := json.Marshal(progress)
	require.NoError(t, err)
	// The historical sequence of existing JSON fields remains digest-stable.
	require.Equal(t, legacy, string(encoded))
	for _, module := range []string{"orders", "customer_service", "customers", "media", "configuration"} {
		t.Run(module, func(t *testing.T) {
			task := &Task{}
			switch module {
			case "orders":
				task.OrderModuleRestore = &progress
			case "customer_service":
				task.CustomerServiceModuleRestore = &progress
			case "customers":
				task.CustomerProfileModuleRestore = &progress
			case "media":
				task.MediaModuleRestore = &progress
			case "configuration":
				task.ConfigurationModuleSiteContentRestore = &progress
			}
			require.Equal(t, 1, countTransactionalBusinessModuleRestoreProgressRecords(task))
			require.Equal(t, module, transactionalBusinessModuleRestoreContractForTask(task).ModuleName)
			require.Same(t, &progress, transactionalBusinessModuleRestoreProgressForTask(task))
		})
	}
}

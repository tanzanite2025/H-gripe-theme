package localbackup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectedBusinessModuleBackupRejectsUnknownEmptyDuplicateAndSQLModuleSelections(t *testing.T) {
	manager := &Manager{}
	for _, modules := range [][]string{nil, {}, {"unknown"}, {"orders", "orders"}, {"orders; DROP TABLE users"}} {
		_, err := manager.PreviewSelectedBusinessModuleBackup(context.Background(), modules)
		require.Error(t, err)
		_, err = manager.CreateSelectedBusinessModuleBackup("synthetic", "", modules)
		require.Error(t, err)
	}
	for _, option := range availableBusinessModuleBackupOptions() {
		selectors := selectedBusinessModuleTableSelectors([]string{option.ID})
		for table := range selectors {
			require.True(t, registeredStorefrontModuleRecoveryTableNames()[table], "%s contains unregistered %s", option.ID, table)
		}
		if option.ID != "customers" {
			require.NotContains(t, selectors, "users")
		}
		if option.ID != "orders" {
			require.NotContains(t, selectors, "orders")
		}
		if option.ID != "customer_service" {
			require.NotContains(t, selectors, "ticket_messages")
		}
		if option.ID != "configuration" {
			require.NotContains(t, selectors, "social_oauth_connections", "store-wide OAuth credentials belong to business configuration")
		}
		require.NotContains(t, selectors, "ops_connectors")
	}
}

func TestSelectedBusinessModulePreviewAndGenerationRequireExportPermission(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	key, err := readProtectedCredential(cfg.AuthenticationKeyFile)
	require.NoError(t, err)
	for _, permission := range []string{PermissionView, PermissionUpload, PermissionRestore} {
		token, err := IssueMaintenanceSession(key, cfg.SiteID, cfg.Environment, "synthetic", []string{permission})
		require.NoError(t, err)
		for _, endpoint := range []string{"/api/module-backup-preview", "/api/module-backups"} {
			request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, strings.NewReader(`{"modules":["orders"]}`))
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			manager.HTTPHandler().ServeHTTP(response, request)
			require.Equal(t, http.StatusForbidden, response.Code)
		}
	}
}

func TestSelectedBusinessModuleAttachmentDiscoveryPreservesPrivateFilesAndRejectsDirtyJSONAndTraversal(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	snapshot := selectedBusinessModuleCompleteSnapshot{Tables: map[string]productCatalogTableSnapshot{
		"ticket_messages":         {Rows: []json.RawMessage{json.RawMessage(`{"id":1,"content":"customer message","attachments":"[\"/uploads-private/conversations/photo.png\"]"}`)}},
		"media_assets":            {Rows: []json.RawMessage{json.RawMessage(`{"id":2,"visibility":"private","storage_key":"conversations/original.png","url":"/api/customer-service/attachments/2"}`)}},
		"media_asset_derivatives": {Rows: []json.RawMessage{json.RawMessage(`{"id":3,"media_asset_id":2,"storage_key":"conversations/small.png","url":"/api/customer-service/attachments/2/small"}`)}},
	}}
	paths, err := cfg.collectSelectedBusinessModuleRequiredFilePaths(snapshot)
	require.NoError(t, err)
	require.Equal(t, map[string]bool{"media/uploads_private/conversations/photo.png": true, "media/uploads_private/conversations/original.png": true, "media/uploads_private/conversations/small.png": true}, paths)
	for _, raw := range []string{`{"attachments":"not-json"}`, `{"attachments":"{}"}`, `{"attachments":"[\"/uploads/../outside\"]"}`, `{"attachments":"[\"https://remote.invalid/file.png\"]"}`, `{"attachments":"[\"unresolved-filename\"]"}`} {
		snapshot.Tables["ticket_messages"] = productCatalogTableSnapshot{Rows: []json.RawMessage{json.RawMessage(raw)}}
		_, err := cfg.collectSelectedBusinessModuleRequiredFilePaths(snapshot)
		require.Error(t, err, raw)
	}
	for _, filename := range []string{"database.dump", "recovery-secrets.json", "media/uploads/../secret", "media/unregistered/file", "product-static/company/../../secret"} {
		require.False(t, isAllowedBackupFileForScope(selectedBusinessModulesBackupScope, filename), filename)
	}
}

func TestSelectedBusinessModuleEncryptedUploadChecksActualBytesWithoutPromotingToWholeRestore(t *testing.T) {
	cfg := createIndependentRecoveryTestConfiguration(t)
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	snapshot := selectedBusinessModuleCompleteSnapshot{ContractVersion: 1, Modules: []string{"orders"}, Tables: map[string]productCatalogTableSnapshot{}, RequiredFilePaths: []string{"media/uploads_private/evidence/receipt.bin"}}
	selectors := selectedBusinessModuleTableSelectors(snapshot.Modules)
	selectors["media_assets"], selectors["media_asset_derivatives"] = "true", "true"
	manifest := Manifest{FormatVersion: PackageVersion, PackageID: newTaskID(), SiteID: cfg.SiteID, Environment: cfg.Environment, ApplicationVersion: cfg.ApplicationVersion, MigrationVersion: cfg.ExpectedMigrationVersion, Scope: selectedBusinessModulesBackupScope, SelectedModules: snapshot.Modules, Tables: map[string]int64{}, Files: []FileIdentity{}, Gaps: []string{}}
	for name := range selectors {
		snapshot.Tables[name] = productCatalogTableSnapshot{SchemaSHA256: strings.Repeat("a", 64), ConstraintsSHA256: strings.Repeat("b", 64), Rows: []json.RawMessage{}}
		manifest.Tables[name] = 0
	}
	evidence := []byte("complete private evidence bytes")
	row, err := json.Marshal(map[string]any{"id": 1, "storage_key": "evidence/receipt.bin", "size_bytes": len(evidence), "sha256": productCatalogBytesSHA256(evidence)})
	require.NoError(t, err)
	contents := snapshot.Tables["order_evidence_attachments"]
	contents.Rows = []json.RawMessage{row}
	snapshot.Tables["order_evidence_attachments"] = contents
	manifest.Tables["order_evidence_attachments"] = 1
	data, err := json.Marshal(snapshot)
	require.NoError(t, err)
	manifest.DatabaseFingerprint = productCatalogBytesSHA256(data)
	source := t.TempDir()
	for path, body := range map[string][]byte{selectedBusinessModulesRowsFilename: data, snapshot.RequiredFilePaths[0]: evidence} {
		filename := filepath.Join(source, filepath.FromSlash(path))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, body, 0600))
		identity, err := fileIdentity(filename, path)
		require.NoError(t, err)
		manifest.Files = append(manifest.Files, identity)
	}
	require.NoError(t, cfg.validateSelectedBusinessModulePackageContents(snapshot, manifest))
	filename := filepath.Join(t.TempDir(), "selected-order-module.age")
	passphrase, err := readProtectedCredential(cfg.PassphraseFile)
	require.NoError(t, err)
	require.NoError(t, WriteEncryptedPackage(filename, source, manifest, passphrase, cfg.MaxPackageBytes))
	_, err = ExtractAndVerifyEncryptedPackage(filename, filepath.Join(t.TempDir(), "whole-restore-must-refuse"), passphrase, cfg.MaxExpandedBytes, cfg.MaxFiles)
	require.Error(t, err)
	encrypted, err := os.ReadFile(filename)
	require.NoError(t, err)
	upload := uploadProductCatalogPackageForIntegrationTest(t, manager, encrypted)
	require.Equal(t, "module_package_integrity_verified", upload.State, upload.Error)
	require.Empty(t, upload.StageDatabase)
	require.Nil(t, upload.Journal)
	require.Equal(t, snapshot.Modules, upload.SelectedBackupModules)
	directory, err := manager.taskDirectory(upload.ID)
	require.NoError(t, err)
	restored, err := os.ReadFile(filepath.Join(directory, "verified", filepath.FromSlash(snapshot.RequiredFilePaths[0])))
	require.NoError(t, err)
	require.Equal(t, evidence, restored)
	_, err = manager.ApplyVerifiedRecovery(upload.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+manifest.PackageID)
	require.Error(t, err)
	// Even a forged generic state cannot open the whole-storefront scope gate.
	upload.State = "verified"
	require.NoError(t, manager.saveTask(upload))
	_, err = manager.ApplyVerifiedRecovery(upload.ID, "RESTORE "+cfg.SiteID+"/"+cfg.Environment+"/"+manifest.PackageID)
	require.Error(t, err)
	manifest.Files = append(manifest.Files, FileIdentity{Path: "media/uploads/unselected-customer.txt", Size: 1, SHA256: strings.Repeat("c", 64)})
	require.ErrorContains(t, cfg.validateSelectedBusinessModulePackageContents(snapshot, manifest), "未选择")
	delete(snapshot.Tables, "order_items")
	require.Error(t, cfg.validateSelectedBusinessModulePackageContents(snapshot, manifest))
}

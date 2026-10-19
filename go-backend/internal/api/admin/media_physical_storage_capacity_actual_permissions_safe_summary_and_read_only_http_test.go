package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyPhysicalCapacityHTTPRequiresMediaConfigurationPermissionBeforeReadingSources(t *testing.T) {
	for _, role := range []string{"editor", "support", "viewer"} {
		t.Run(role, func(t *testing.T) {
			_, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
			router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, role)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/admin/media/storage-capacity-audits", nil))
			require.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
		})
	}
}

func TestReadOnlyPhysicalCapacityHTTPReturnsOnlySafeAggregatesWithoutChangingDatabaseOrFiles(t *testing.T) {
	db, _ := createMediaLibraryPartitionHTTPTestEnvironment(t)
	require.NoError(t, db.AutoMigrate(&media.MediaUploadSession{}, &media.BusinessDraftMediaUploadReceipt{}))
	root, privateRoot := t.TempDir(), t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, PrivateLocalPath: privateRoot, BaseURL: "https://capacity.example.test/uploads"})
	require.NoError(t, err)
	key := "private-customer-name.png"
	require.NoError(t, os.WriteFile(filepath.Join(root, key), []byte("customer-photo"), 0600))
	asset := media.MediaAsset{Filename: key, URL: provider.GetURL(key), StorageKey: key, Size: 14}
	require.NoError(t, db.Create(&asset).Error)
	handler := NewMediaHandler(service.NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://capacity.example.test", 20<<30))
	router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "admin")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/admin/media/storage-capacity-audits", nil))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	for _, sensitive := range []string{root, privateRoot, key, "capacity.example.test"} {
		require.NotContains(t, response.Body.String(), sensitive)
	}
	var body struct {
		Data struct {
			Report service.MediaPhysicalCapacityAuditReport `json:"report"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.True(t, body.Data.Report.PhysicalScanComplete)
	require.Equal(t, service.MediaPhysicalCapacityMetric{Objects: 1, Bytes: 14}, body.Data.Report.Observed)
	var after media.MediaAsset
	require.NoError(t, db.First(&after, asset.ID).Error)
	require.True(t, asset.UpdatedAt.Equal(after.UpdatedAt))
	content, err := os.ReadFile(filepath.Join(root, key))
	require.NoError(t, err)
	require.Equal(t, "customer-photo", string(content))
}

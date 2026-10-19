package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"github.com/stretchr/testify/require"
)

func TestMediaAssociationDiagnosisHTTPUsesActualPermissionsAndRedactsRestrictedReferences(t *testing.T) {
	for _, role := range []string{"user", "editor", "support", "viewer", "manager", "admin"} {
		t.Run(role, func(t *testing.T) {
			db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
			asset := media.MediaAsset{Filename: "diagnose.png", StorageKey: "diagnose.png", URL: "https://shop.example.test/uploads/diagnose.png", LibraryPartition: media.MediaLibraryPartitionShared}
			require.NoError(t, db.Create(&asset).Error)
			require.NoError(t, db.Exec("CREATE TABLE ticket_messages(id INTEGER PRIMARY KEY,ticket_id INTEGER,attachments TEXT)").Error)
			require.NoError(t, db.Exec("INSERT INTO ticket_messages VALUES(876543,765432,?)", `["`+asset.URL+`"]`).Error)
			router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, role)
			result := httptest.NewRecorder()
			router.ServeHTTP(result, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/admin/media/assets/%d/association-diagnosis?allow_all_references=true", asset.ID), nil))
			if role == "user" {
				require.Equal(t, http.StatusForbidden, result.Code)
				return
			}
			require.Equal(t, http.StatusOK, result.Code, result.Body.String())
			require.Equal(t, "no-store", result.Header().Get("Cache-Control"))
			if role != "admin" && role != "manager" {
				require.Contains(t, result.Body.String(), `"has_restricted_references":true`)
				for _, secret := range []string{"876543", "765432", "ticket_message", asset.URL} {
					require.NotContains(t, result.Body.String(), secret)
				}
			}
		})
	}
}

func TestMissingGalleryParentRepairHTTPRejectsPermissionAndUntrustedRequestFields(t *testing.T) {
	for _, role := range []string{"user", "editor", "support", "viewer"} {
		t.Run(role, func(t *testing.T) {
			_, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
			router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, role)
			for _, suffix := range []string{"missing-gallery-parent-repair-previews", "missing-gallery-parent-repairs"} {
				result := httptest.NewRecorder()
				router.ServeHTTP(result, httptest.NewRequest(http.MethodPost, "/api/admin/media/assets/1/"+suffix, strings.NewReader(`{"image_id":"1","destination_gallery_id":"2"}`)))
				require.Equal(t, http.StatusForbidden, result.Code, result.Body.String())
			}
		})
	}
	for _, payload := range []string{`{"image_id":1,"destination_gallery_id":"2"}`, `{"image_id":"01","destination_gallery_id":"2"}`, `{"image_id":"1","destination_gallery_id":"2","user_id":42}`, `{"image_id":"1","destination_gallery_id":"2"} {}`, `{"image_id":"18446744073709551616","destination_gallery_id":"2"}`, `null`, strings.Repeat(" ", 17<<10) + `{}`} {
		t.Run(fmt.Sprintf("invalid payload %d", len(payload)), func(t *testing.T) {
			_, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
			result := httptest.NewRecorder()
			createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "admin").ServeHTTP(result, httptest.NewRequest(http.MethodPost, "/api/admin/media/assets/1/missing-gallery-parent-repair-previews", strings.NewReader(payload)))
			require.Equal(t, http.StatusBadRequest, result.Code, result.Body.String())
		})
	}
}

func TestMissingGalleryParentPreviewHTTPReturnsOnlyReviewedIdentitiesAndStaleHistoryConflict(t *testing.T) {
	db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	require.NoError(t, db.AutoMigrate(&gallery.Gallery{}, &gallery.GalleryImage{}))
	asset := media.MediaAsset{Filename: "review.png", StorageKey: "review.png", URL: "https://shop.example.test/uploads/review.png", MediaType: "image", Status: "active", Visibility: "public", LibraryPartition: media.MediaLibraryPartitionShared}
	require.NoError(t, db.Create(&asset).Error)
	destination := gallery.Gallery{Name: "Selected draft", Slug: "selected-draft", Status: "draft"}
	require.NoError(t, db.Create(&destination).Error)
	image := gallery.GalleryImage{GalleryID: 876543, MediaAssetID: &asset.ID, URL: asset.URL, Title: "Historical image"}
	require.NoError(t, db.Create(&image).Error)
	router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "manager")
	requestBody := fmt.Sprintf(`{"image_id":"%d","destination_gallery_id":"%d"}`, image.ID, destination.ID)
	result := httptest.NewRecorder()
	router.ServeHTTP(result, httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/media/assets/%d/missing-gallery-parent-repair-previews", asset.ID), strings.NewReader(requestBody)))
	require.Equal(t, http.StatusOK, result.Code, result.Body.String())
	require.Contains(t, result.Body.String(), "Selected draft")
	require.Contains(t, result.Body.String(), "fingerprint")
	require.NotContains(t, result.Body.String(), asset.URL)
	require.NoError(t, db.Delete(&image).Error)
	result = httptest.NewRecorder()
	router.ServeHTTP(result, httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/media/assets/%d/missing-gallery-parent-repair-previews", asset.ID), strings.NewReader(requestBody)))
	require.Equal(t, http.StatusConflict, result.Code, result.Body.String())
	require.Contains(t, result.Body.String(), "media_association_repair_conflict")
}

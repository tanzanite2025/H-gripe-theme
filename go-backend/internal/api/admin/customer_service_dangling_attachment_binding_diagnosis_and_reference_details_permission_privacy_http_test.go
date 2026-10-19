package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
)

func TestCustomerServiceDanglingBindingHTTPDiagnosisAndDetailsRequireFullConversationMetadataScope(t *testing.T) {
	for _, role := range []string{"user", "editor", "support", "viewer", "manager", "admin"} {
		t.Run(role, func(t *testing.T) {
			db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
			asset := media.MediaAsset{Filename: "safe-title.png", URL: "https://shop.example.test/uploads/safe-title.png", StorageKey: "safe-title.png", LibraryPartition: media.MediaLibraryPartitionShared}
			require.NoError(t, db.Create(&asset).Error)
			require.NoError(t, db.Exec("CREATE TABLE ticket_message_media_attachments(id INTEGER PRIMARY KEY,ticket_message_id INTEGER,media_asset_id INTEGER)").Error)
			require.NoError(t, db.Exec("INSERT INTO ticket_message_media_attachments VALUES(876543,765432,?)", asset.ID).Error)
			router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, role)
			for _, operation := range []struct{ method, suffix string }{{http.MethodGet, "association-diagnosis"}, {http.MethodGet, "references"}, {http.MethodPost, ""}} {
				path := fmt.Sprintf("/api/admin/media/assets/%d/%s", asset.ID, operation.suffix)
				if operation.method == http.MethodPost {
					path = "/api/admin/media/usage-summaries"
				}
				response := httptest.NewRecorder()
				router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), operation.method, path+"?allow_all_references=true", nil))
				if role == "user" {
					require.Equal(t, http.StatusForbidden, response.Code)
					continue
				}
				require.Equal(t, http.StatusOK, response.Code, response.Body.String())
				require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
				if role == "admin" || role == "manager" {
					require.Contains(t, response.Body.String(), "ticket_message_attachment_binding")
					require.Contains(t, response.Body.String(), "876543")
					require.Contains(t, response.Body.String(), "missing_parent")
				} else {
					for _, secret := range []string{"876543", "765432", "ticket_message_attachment_binding", "missing_parent"} {
						require.NotContains(t, response.Body.String(), secret)
					}
				}
				require.NotContains(t, response.Body.String(), asset.URL)
			}
		})
	}
}

package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/service"
)

func TestReadOnlyMediaUsageSummaryHTTPRequiresMediaViewAndHidesConversationMetadataOutsideFullScope(t *testing.T) {
	for _, role := range []string{"user", "editor", "support", "viewer", "manager", "admin"} {
		t.Run(role, func(t *testing.T) {
			db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
			public := media.MediaAsset{Filename: "public.png", URL: "https://shop.example.test/uploads/public.png", LibraryPartition: media.MediaLibraryPartitionShared}
			private := media.MediaAsset{Filename: "private-customer-identity.png", URL: "https://shop.example.test/uploads/customer-service/attachments/originals/private-customer-identity.png",
				StorageKey: "customer-service/attachments/originals/private-customer-identity.png", Visibility: "private", LibraryPartition: media.MediaLibraryPartitionCustomerService,
				UploadOrigin: media.MediaUploadOriginCustomerServiceConversation, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeConversationAttachment}
			require.NoError(t, db.Create(&public).Error)
			require.NoError(t, db.Create(&private).Error)
			require.NoError(t, db.Exec("CREATE TABLE ticket_messages(id INTEGER PRIMARY KEY,ticket_id INTEGER,attachments TEXT)").Error)
			require.NoError(t, db.Exec("INSERT INTO ticket_messages(id,ticket_id,attachments) VALUES(987654,876543,?)", `["`+public.URL+`"]`).Error)
			router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, role)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/admin/media/usage-summaries?exclude_conversation_attachments=false&allow_all_references=true", nil))
			if role == "user" {
				require.Equal(t, http.StatusForbidden, response.Code)
				return
			}
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
			var body struct {
				Data struct {
					Report service.MediaUsageSummaryReport `json:"report"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
			report := body.Data.Report
			if role != "admin" && role != "manager" {
				require.Len(t, report.Rows, 1)
				require.True(t, report.Rows[0].HasRestrictedReferences)
				require.False(t, report.Rows[0].NoRecognizedReferences)
				require.Zero(t, report.Rows[0].VisibleReferenceTotal)
				for _, secret := range []string{"private-customer-identity", "ticket_message", "987654", "876543", "customer-service/", public.URL, "storage_key"} {
					require.NotContains(t, response.Body.String(), secret)
				}
			} else {
				require.Len(t, report.Rows, 2)
				require.Equal(t, 1, report.Rows[0].VisibleReferenceTotal)
				require.Contains(t, response.Body.String(), "987654")
				require.NotContains(t, response.Body.String(), "customer-service/")
			}
			var after media.MediaAsset
			require.NoError(t, db.First(&after, public.ID).Error)
			require.True(t, public.UpdatedAt.Equal(after.UpdatedAt))
			// The existing detail endpoint must not bypass summary redaction.
			inspection := httptest.NewRecorder()
			router.ServeHTTP(inspection, httptest.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("/api/admin/media/assets/%d/references", public.ID), nil))
			require.Equal(t, http.StatusOK, inspection.Code, inspection.Body.String())
			require.Equal(t, "no-store", inspection.Header().Get("Cache-Control"))
			if role != "admin" && role != "manager" {
				require.Contains(t, inspection.Body.String(), "restricted_reference")
				require.Contains(t, inspection.Body.String(), `"total":1`)
				for _, secret := range []string{"ticket_message", "987654", "876543"} {
					require.NotContains(t, inspection.Body.String(), secret)
				}
			}
			if role == "manager" || role == "admin" {
				reviewResponse := httptest.NewRecorder()
				router.ServeHTTP(reviewResponse, httptest.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("/api/admin/media/assets/%d/library-classification-review", public.ID), nil))
				require.Equal(t, http.StatusOK, reviewResponse.Code, reviewResponse.Body.String())
				require.Contains(t, reviewResponse.Body.String(), `"recognized_reference_total":1`)
			}
		})
	}
}

func TestReadOnlyMediaUsageSummaryRedactionAlsoProtectsDeletionConflictResponses(t *testing.T) {
	_, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "manager")
	// Exercise the exact error mapper used after the authoritative deletion
	// transaction; no client input can replace this server-side reference set.
	router.GET("/test-media-deletion-conflict", func(c *gin.Context) {
		respondMediaError(c, &service.MediaAssetInUseError{References: []media.AssetReference{
			{ResourceType: "warranty_claim", Category: media.ReferenceCategoryCustomer, ResourceID: 987654, ParentResourceID: 876543, Label: "private customer"},
			{ResourceType: "warranty_claim", Category: media.ReferenceCategoryCustomer, ResourceID: 987655, ParentResourceID: 876543, Label: "private customer"},
		}})
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test-media-deletion-conflict", nil))
	require.Equal(t, http.StatusConflict, response.Code)
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	require.Contains(t, response.Body.String(), "media_asset_in_use")
	require.Contains(t, response.Body.String(), "restricted_reference")
	require.Contains(t, response.Body.String(), `"total":1`)
	for _, private := range []string{"987654", "987655", "876543", "private customer", "warranty_claim"} {
		require.NotContains(t, response.Body.String(), private)
	}
}

func TestReadOnlyMediaUsageSummaryOwnershipReviewDoesNotExposePrivateCustomerReferenceCounts(t *testing.T) {
	db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	asset := media.MediaAsset{Filename: "safe.png", URL: "https://shop.example.test/safe.png"}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, db.Exec("CREATE TABLE suggestion_feedback(id INTEGER PRIMARY KEY,attachments TEXT,deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO suggestion_feedback(id,attachments) VALUES(876543,?),(987654,?)", `["`+asset.URL+`"]`, `["`+asset.URL+`"]`).Error)
	response := httptest.NewRecorder()
	createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "manager").ServeHTTP(response,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("/api/admin/media/assets/%d/library-classification-review", asset.ID), nil))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	require.Contains(t, response.Body.String(), `"has_restricted_references":true`)
	require.Contains(t, response.Body.String(), `"recognized_reference_total":0`)
	for _, private := range []string{`"customer"`, "876543", "987654", "suggested_ownership"} {
		require.NotContains(t, response.Body.String(), private)
	}
}

func TestReadOnlyMediaUsageSummaryHTTPRejectsForgedConversationScopeAndInvalidFilters(t *testing.T) {
	_, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "editor")
	for _, query := range []string{"library_partition=unknown", "media_type=script", "status=deleted", "visibility=secret", "customer_service_media_purpose=conversation_attachment"} {
		t.Run(query, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/admin/media/usage-summaries?"+query, nil))
			require.Contains(t, []int{http.StatusBadRequest, http.StatusForbidden}, response.Code, response.Body.String())
			require.NotContains(t, response.Body.String(), `"report"`)
		})
	}
}

func TestReadOnlyMediaUsageSummaryHTTPScannerFailureDoesNotReturnZeroReferencesOrInternalDetails(t *testing.T) {
	db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	require.NoError(t, db.Create(&media.MediaAsset{Filename: "safe.png", URL: "https://shop.example.test/safe.png"}).Error)
	require.NoError(t, db.Exec("CREATE TABLE suggestion_feedback(id INTEGER PRIMARY KEY,attachments TEXT,deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO suggestion_feedback(id,attachments) VALUES(1,'{broken')").Error)
	response := httptest.NewRecorder()
	createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "admin").ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/admin/media/usage-summaries", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.Contains(t, response.Body.String(), "media_usage_summary_unavailable")
	for _, internal := range []string{"suggestion_feedback", "attachments", "broken", `"report"`, `"rows"`} {
		require.NotContains(t, response.Body.String(), internal)
	}
}

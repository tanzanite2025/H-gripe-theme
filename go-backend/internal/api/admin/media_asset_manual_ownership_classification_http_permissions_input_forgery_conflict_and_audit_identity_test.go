package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/media"
)

func sendMediaAssetManualClassificationJSONRequest(router http.Handler, path, payload string) *httptest.ResponseRecorder {
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPatch, path, bytes.NewBufferString(payload))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestMediaAssetManualClassificationHTTPRequiresConfigurationPermissionAndRecordsAuthenticatedReviewer(t *testing.T) {
	for _, role := range []string{"admin", "manager", "editor", "support", "viewer"} {
		t.Run(role, func(t *testing.T) {
			db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
			require.NoError(t, db.AutoMigrate(&audit.AuditLog{}))
			asset := media.MediaAsset{Filename: "history.png", URL: "https://shop.example.test/uploads/history.png"}
			require.NoError(t, db.Create(&asset).Error)
			router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, role)
			path := fmt.Sprintf("/api/admin/media/assets/%d", asset.ID)
			reviewRecorder := httptest.NewRecorder()
			router.ServeHTTP(reviewRecorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path+"/library-classification-review", nil))
			recorder := sendMediaAssetManualClassificationJSONRequest(router, path+"/library-classification", `{"library_partition":"catalog","expected_revision":0,"reason":"人工确认商品素材"}`)
			if role != "admin" && role != "manager" {
				require.Equal(t, http.StatusForbidden, reviewRecorder.Code)
				require.Equal(t, http.StatusForbidden, recorder.Code)
				return
			}
			require.Equal(t, http.StatusOK, reviewRecorder.Code, reviewRecorder.Body.String())
			require.NotContains(t, reviewRecorder.Body.String(), asset.URL)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			var log audit.AuditLog
			require.NoError(t, db.First(&log).Error)
			require.EqualValues(t, 42, log.UserID)
			require.Equal(t, "reclassify", log.Action)
			require.Equal(t, "PATCH", log.Method)
			require.Equal(t, path+"/library-classification", log.Path)
			require.Contains(t, log.OldValue, `"upload_origin":"legacy_unknown"`)
			require.Contains(t, log.NewValue, `"classification_revision":1`)
			recorder = sendMediaAssetManualClassificationJSONRequest(router, path+"/library-classification", `{"library_partition":"content","expected_revision":0,"reason":"此处版本已过期"}`)
			require.Equal(t, http.StatusConflict, recorder.Code)
			require.Contains(t, recorder.Body.String(), "media_library_classification_revision_conflict")
			var logCount int64
			require.NoError(t, db.Model(&audit.AuditLog{}).Count(&logCount).Error)
			require.EqualValues(t, 1, logCount)
		})
	}
}

func TestMediaAssetManualClassificationHTTPRejectsForgedSourceActorAndConversationOwnership(t *testing.T) {
	db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	require.NoError(t, db.AutoMigrate(&audit.AuditLog{}))
	asset := media.MediaAsset{Filename: "history.png", URL: "https://shop.example.test/uploads/history.png"}
	require.NoError(t, db.Create(&asset).Error)
	router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "admin")
	path := fmt.Sprintf("/api/admin/media/assets/%d/library-classification", asset.ID)
	for _, payload := range []string{
		`{"library_partition":"catalog","expected_revision":0,"reason":"核对商品素材","upload_origin":"administrator_catalog_upload"}`,
		`{"library_partition":"catalog","expected_revision":0,"reason":"核对商品素材","user_id":99}`,
		`{"library_partition":"catalog","expected_revision":0,"reason":"核对商品素材","classification_method":"recorded_upload_context"}`,
		`{"library_partition":"customer_service","customer_service_media_purpose":"conversation_attachment","expected_revision":0,"reason":"核对客服附件"}`,
		`{"library_partition":"catalog","reason":"缺少预期版本"}`,
		`{"library_partition":"catalog","expected_revision":0,"reason":"短"}`,
		`{"library_partition":"catalog","expected_revision":0,"reason":"核对商品素材"} {}`,
	} {
		recorder := sendMediaAssetManualClassificationJSONRequest(router, path, payload)
		require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
	stored, err := handler.mediaService.GetAsset(asset.ID)
	require.NoError(t, err)
	require.Equal(t, media.MediaUploadOriginLegacyUnknown, stored.UploadOrigin)
	require.Zero(t, stored.ClassificationRevision)
	conversation := media.MediaAsset{Filename: "conversation.png", URL: "https://shop.example.test/uploads/customer-service/public-conversation-attachments/conversation.png"}
	require.NoError(t, db.Create(&conversation).Error)
	recorder := sendMediaAssetManualClassificationJSONRequest(router, fmt.Sprintf("/api/admin/media/assets/%d/library-classification", conversation.ID), `{"library_partition":"shared","expected_revision":0,"reason":"禁止转为公共素材"}`)
	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.NotContains(t, recorder.Body.String(), conversation.URL)
	var logCount int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&logCount).Error)
	require.Zero(t, logCount)
}

func TestMediaAssetManualClassificationHTTPAuditFailureDoesNotReportSuccessOrPersistOwnership(t *testing.T) {
	db, handler := createMediaLibraryPartitionHTTPTestEnvironment(t)
	asset := media.MediaAsset{Filename: "history.png", URL: "https://shop.example.test/uploads/history.png"}
	require.NoError(t, db.Create(&asset).Error)
	router := createMediaLibraryPartitionHTTPRouterWithActualRegisteredPermissions(handler, "manager")
	recorder := sendMediaAssetManualClassificationJSONRequest(router, fmt.Sprintf("/api/admin/media/assets/%d/library-classification", asset.ID), `{"library_partition":"shared","expected_revision":0,"reason":"核对公共素材"}`)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	stored, err := handler.mediaService.GetAsset(asset.ID)
	require.NoError(t, err)
	before, _ := json.Marshal(media.CreateMediaAssetLibraryClassificationSnapshot(&asset))
	after, _ := json.Marshal(media.CreateMediaAssetLibraryClassificationSnapshot(stored))
	require.JSONEq(t, string(before), string(after))
}

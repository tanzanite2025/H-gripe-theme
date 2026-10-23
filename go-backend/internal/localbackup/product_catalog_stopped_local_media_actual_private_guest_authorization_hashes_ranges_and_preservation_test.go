package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/visitorcookie"
)

const stoppedProductMediaSyntheticVisitorSessionID = "33333333-3333-4333-8333-333333333333"
const stoppedProductMediaSyntheticOriginalKey = "customer-service/attachments/originals/stopped-recovery-private.png"

func seedStoppedProductMediaPrivateGuestConversationBeforeRecoverySnapshots(t *testing.T, database *sql.DB, configuration Configuration) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `DO $$ DECLARE conversation bigint; asset bigint; BEGIN
INSERT INTO tickets(ticket_number,user_id,conversation_id,visitor_session_hash,subject,category)
VALUES('stopped-recovery-guest',(SELECT id FROM users WHERE username='product-backup-customer'),'stopped-recovery-guest','`+visitorcookie.CustomerServiceVisitorHash(stoppedProductMediaSyntheticVisitorSessionID)+`','Guest media recovery fixture','customer_service') RETURNING id INTO conversation;
INSERT INTO media_assets(filename,url,storage_key,size,mime_type,media_type,visibility,status,library_partition,upload_origin,classification_method,customer_service_media_purpose)
VALUES('stopped-recovery-private.png','/uploads/`+stoppedProductMediaSyntheticOriginalKey+`','`+stoppedProductMediaSyntheticOriginalKey+`',26,'image/png','image','private','active','customer_service','customer_service_conversation_upload','recorded_upload_context','conversation_attachment') RETURNING id INTO asset;
INSERT INTO media_asset_derivatives(media_asset_id,preset,preset_version,url,storage_key,mime_type,size)
VALUES(asset,'thumbnail',1,'/uploads/customer-service/attachments/derivatives/'||asset||'/thumbnail/v1/stopped-recovery-private.png','customer-service/attachments/derivatives/'||asset||'/thumbnail/v1/stopped-recovery-private.png','image/png',24);
INSERT INTO ticket_messages(ticket_id,content,attachments,attachment_reference_format)
VALUES(conversation,'Retained guest media message','["/uploads/`+stoppedProductMediaSyntheticOriginalKey+`"]','legacy_urls');
END $$`)
	require.NoError(t, err)
	var derivativeKey string
	require.NoError(t, database.QueryRowContext(context.Background(), `SELECT storage_key FROM media_asset_derivatives WHERE storage_key LIKE 'customer-service/attachments/derivatives/%/thumbnail/v1/stopped-recovery-private.png'`).Scan(&derivativeKey))
	for key, body := range map[string]string{stoppedProductMediaSyntheticOriginalKey: "private-original-fixture-v1", derivativeKey: "private-thumbnail-fixture"} {
		filename := filepath.Join(configuration.StorageRoots["uploads_private"], filepath.FromSlash(key))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, []byte(body), 0600))
	}
}

func verifyStoppedProductMediaActualGuestAuthorizationHashesRangesAndUnchangedRecoveryBoundary(t *testing.T, manager *Manager, taskID string, invokeHTTP func(string, any, int) *httptest.ResponseRecorder) {
	t.Helper()
	ctx := context.Background()
	before, err := manager.GetTask(taskID)
	require.NoError(t, err)
	baseline, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	var messageID string
	require.NoError(t, manager.withBoundStoppedProductReadOnlyTransaction(ctx, manager.Configuration.DatabaseName, before.ProductCatalogPreviewBinding.Candidate, func(tx *sql.Tx) error {
		return tx.QueryRowContext(context.Background(), `SELECT m.id::text FROM ticket_messages m JOIN tickets t ON t.id=m.ticket_id WHERE t.conversation_id='stopped-recovery-guest'`).Scan(&messageID)
	}))
	cookie := visitorcookie.SignCustomerServiceVisitorSessionID(stoppedProductMediaSyntheticVisitorSessionID, []byte(manager.Configuration.RecoverySecrets["JWT_SECRET"]))
	for _, scenario := range []string{"public_original", "public_derivative", "public_range", "private_direct_anonymous", "private_original_anonymous", "private_original_owner", "private_thumbnail_owner", "private_owner_range", "private_owner_invalid_range", "private_foreign_visitor", "private_tampered_cookie", "private_missing_message"} {
		t.Run("media_read_only_"+scenario, func(t *testing.T) {
			query := StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: baseline.ContractSHA256, Operation: "conversation_attachment", ConversationID: "stopped-recovery-guest", MessageID: messageID, VisitorCookie: cookie}
			expectedStatus := http.StatusOK
			expectedBody := "private-original-fixture-v1"
			switch scenario {
			case "public_original", "public_derivative", "public_range":
				query = StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: baseline.ContractSHA256, Operation: "public_upload", ObjectKey: "products/product-original.png"}
				if scenario == "public_derivative" {
					query.ObjectKey = "products/product-small.png"
				}
				body, err := os.ReadFile(filepath.Join(manager.Configuration.StorageRoots["uploads"], filepath.FromSlash(query.ObjectKey)))
				require.NoError(t, err)
				expectedBody = string(body)
				if scenario == "public_range" {
					query.Range = "bytes=0-7"
					expectedStatus = http.StatusPartialContent
					expectedBody = expectedBody[:8]
				}
			case "private_direct_anonymous":
				query = StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: baseline.ContractSHA256, Operation: "public_upload", ObjectKey: stoppedProductMediaSyntheticOriginalKey}
				expectedStatus = http.StatusNotFound
				expectedBody = ""
			case "private_original_anonymous":
				query.VisitorCookie = ""
				expectedStatus = http.StatusNotFound
				expectedBody = ""
			case "private_thumbnail_owner":
				query.Preset = "thumbnail"
				expectedBody = "private-thumbnail-fixture"
			case "private_owner_range":
				query.Range = "bytes=0-7"
				expectedStatus = http.StatusPartialContent
				expectedBody = expectedBody[:8]
			case "private_owner_invalid_range":
				query.Range = "bytes=999999-"
				expectedStatus = http.StatusRequestedRangeNotSatisfiable
				expectedBody = "invalid range: failed to overlap\n"
			case "private_foreign_visitor":
				query.VisitorCookie = visitorcookie.SignCustomerServiceVisitorSessionID("44444444-4444-4444-8444-444444444444", []byte(manager.Configuration.RecoverySecrets["JWT_SECRET"]))
				expectedStatus = http.StatusNotFound
				expectedBody = ""
			case "private_tampered_cookie":
				query.VisitorCookie += "tampered"
				expectedStatus = http.StatusNotFound
				expectedBody = ""
			case "private_missing_message":
				query.MessageID = "999999999"
				expectedStatus = http.StatusNotFound
				expectedBody = ""
			}
			response := invokeHTTP("product-catalog-stopped-read-only-media-http-query", query, http.StatusOK)
			require.NotContains(t, response.Body.String(), cookie)
			require.NotContains(t, response.Body.String(), "private-original-fixture-v1")
			var payload struct {
				Result StoppedProductReadOnlyMediaHTTPResult
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
			result := payload.Result
			require.Equal(t, expectedStatus, result.HTTPStatus)
			require.Equal(t, int64(len(expectedBody)), result.ResponseBytes)
			require.Equal(t, productCatalogBytesSHA256([]byte(expectedBody)), result.ResponseSHA256)
			require.Empty(t, result.Target.VisitorCookie)
			require.Equal(t, query.VisitorCookie != "", result.VisitorCredentialSupplied)
			require.Equal(t, baseline.ContractSHA256, result.ContractSHA256)
			require.True(t, result.SQLReadOnly)
			require.True(t, result.ApplicationRemainsStopped)
			require.False(t, result.OnlineOverwriteAvailable)
			if strings.HasPrefix(scenario, "private_") {
				require.Equal(t, "private, no-store", result.Headers["Cache-Control"])
			}
			if scenario == "public_original" || scenario == "public_derivative" {
				require.NotNil(t, result.MatchesPackageFile)
				require.True(t, *result.MatchesPackageFile)
			}
		})
	}
	stale := StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: strings.Repeat("b", 64), Operation: "public_upload", ObjectKey: "products/product-original.png"}
	invokeHTTP("product-catalog-stopped-read-only-media-http-query", stale, http.StatusConflict)
	after, err := manager.GetTask(taskID)
	require.NoError(t, err)
	require.Equal(t, before, after)
	current, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, baseline.ContractSHA256, current.ContractSHA256)
}

func TestStoppedProductMediaRequestsRejectForeignTargetsCredentialsAndUnboundedRanges(t *testing.T) {
	valid := StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: "public_upload", ObjectKey: "products/image.png"}
	require.NoError(t, validateStoppedProductReadOnlyMediaHTTPRequest(valid))
	for _, key := range []string{"https://foreign.invalid/image.png", "../secret", "products/%2e%2e/secret", "products/image.png?token=x", "products\\image.png", "/products/image.png"} {
		request := valid
		request.ObjectKey = key
		require.Error(t, validateStoppedProductReadOnlyMediaHTTPRequest(request))
	}
	for _, rangeValue := range []string{"bytes=0-1,4-9", "bytes=-1", "bytes=0-9999999999999999999", "bytes=0-1\n"} {
		request := valid
		request.Range = rangeValue
		require.Error(t, validateStoppedProductReadOnlyMediaHTTPRequest(request))
	}
}

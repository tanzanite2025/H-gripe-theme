package localbackup

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/service"
)

func seedStoppedProductMediaMemberAndStaffAccountsBeforeRecoverySnapshots(t *testing.T, database *sql.DB) {
	t.Helper()
	_, err := database.ExecContext(context.Background(), `
INSERT INTO users(email,username,password,role,status) VALUES
('recovery-other@example.invalid','recovery-other-member','synthetic','user','active'),
('recovery-assigned@example.invalid','recovery-assigned-support','synthetic','support','active'),
('recovery-unassigned@example.invalid','recovery-unassigned-support','synthetic','support','active'),
('recovery-manager@example.invalid','recovery-current-manager','synthetic','manager','active'),
('recovery-admin@example.invalid','recovery-current-admin','synthetic','admin','active'),
('recovery-editor@example.invalid','recovery-current-editor','synthetic','editor','active'),
('recovery-inactive@example.invalid','recovery-inactive-support','synthetic','support','inactive'),
('recovery-deleted@example.invalid','recovery-deleted-support','synthetic','support','active');
UPDATE users SET deleted_at=now() WHERE username='recovery-deleted-support';
INSERT INTO tickets(ticket_number,user_id,customer_user_id,conversation_id,subject,category,assigned_to)
SELECT 'stopped-recovery-member',id,id,'stopped-recovery-member','Member authorization recovery fixture','customer_service',
(SELECT id FROM users WHERE username='recovery-assigned-support') FROM users WHERE username='product-backup-customer';
INSERT INTO ticket_messages(ticket_id,content,is_internal,attachments,attachment_reference_format)
SELECT id,'Member attachment',false,'["/uploads/`+stoppedProductMediaSyntheticOriginalKey+`"]','legacy_urls'
FROM tickets WHERE conversation_id='stopped-recovery-member';
INSERT INTO ticket_messages(ticket_id,content,is_internal,attachments,attachment_reference_format)
SELECT id,'Internal attachment',true,'["/uploads/`+stoppedProductMediaSyntheticOriginalKey+`"]','legacy_urls'
FROM tickets WHERE conversation_id='stopped-recovery-member';
INSERT INTO tickets(ticket_number,user_id,customer_user_id,conversation_id,subject,category)
SELECT 'stopped-'||username,id,id,'stopped-'||username,'Unavailable member own conversation','customer_service'
FROM users WHERE username IN ('recovery-inactive-support','recovery-deleted-support');
INSERT INTO ticket_messages(ticket_id,content,attachments,attachment_reference_format)
SELECT id,'Unavailable member own attachment','["/uploads/`+stoppedProductMediaSyntheticOriginalKey+`"]','legacy_urls'
FROM tickets WHERE conversation_id IN ('stopped-recovery-inactive-support','stopped-recovery-deleted-support');`)
	require.NoError(t, err)
}

// Only fixtures sign identities. The maintenance implementation accepts existing
// business cookies and never issues a token or injects user_id/role itself.
func signSyntheticStoppedProductMediaBusinessAccessCookie(t *testing.T, userID uint, tokenRole, tokenUse, secret string, expired bool) string {
	t.Helper()
	expires := time.Now().Add(time.Hour)
	if expired {
		expires = time.Now().Add(-time.Hour)
	}
	claims := service.Claims{UserID: userID, Role: tokenRole, TokenUse: tokenUse,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expires)}}
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	return value
}

func verifyStoppedProductMediaActualMemberAccountAndAgentPermissionsWithoutChangingRecoveryBoundary(t *testing.T, manager *Manager, taskID string, invokeHTTP func(string, any, int) *httptest.ResponseRecorder) {
	t.Helper()
	ctx := context.Background()
	before, err := manager.GetTask(taskID)
	require.NoError(t, err)
	baseline, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	identities := map[string]uint{}
	unavailableMemberMessages := map[string]string{}
	var conversationID, memberMessageID, internalMessageID, guestMessageID string
	require.NoError(t, manager.withBoundStoppedProductReadOnlyTransaction(ctx, manager.Configuration.DatabaseName, before.ProductCatalogPreviewBinding.Candidate, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(context.Background(), `SELECT username,id FROM users WHERE username='product-backup-customer' OR username LIKE 'recovery-%'`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var username string
			var id uint
			if err := rows.Scan(&username, &id); err != nil {
				_ = rows.Close()
				return err
			}
			identities[username] = id
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		for _, username := range []string{"recovery-inactive-support", "recovery-deleted-support"} {
			var messageID string
			if err := tx.QueryRowContext(context.Background(), `SELECT m.id::text FROM ticket_messages m JOIN tickets t ON t.id=m.ticket_id WHERE t.conversation_id=$1`, "stopped-"+username).Scan(&messageID); err != nil {
				return err
			}
			unavailableMemberMessages[username] = messageID
		}
		if err := tx.QueryRowContext(context.Background(), `SELECT id::text FROM tickets WHERE conversation_id='stopped-recovery-member'`).Scan(&conversationID); err != nil {
			return err
		}
		if err := tx.QueryRowContext(context.Background(), `SELECT id::text FROM ticket_messages WHERE content='Member attachment'`).Scan(&memberMessageID); err != nil {
			return err
		}
		if err := tx.QueryRowContext(context.Background(), `SELECT id::text FROM ticket_messages WHERE content='Internal attachment'`).Scan(&internalMessageID); err != nil {
			return err
		}
		return tx.QueryRowContext(context.Background(), `SELECT m.id::text FROM ticket_messages m JOIN tickets t ON t.id=m.ticket_id WHERE t.conversation_id='stopped-recovery-guest'`).Scan(&guestMessageID)
	}))
	secret := manager.Configuration.RecoverySecrets["JWT_SECRET"]
	for _, scenario := range []string{
		"member_owner", "member_thumbnail", "member_range", "member_anonymous", "member_foreign", "member_expired", "member_tampered", "member_wrong_key", "member_refresh", "member_inactive", "member_deleted", "member_internal_refused", "member_foreign_message_refused",
		"agent_assigned_current_role", "agent_thumbnail", "agent_range", "agent_unassigned_stale_admin_claim", "agent_current_manager", "agent_current_admin", "agent_editor_stale_admin_claim", "agent_storefront_account_stale_admin_claim", "agent_inactive", "agent_deleted", "agent_anonymous", "agent_expired", "agent_tampered", "agent_wrong_key", "agent_refresh", "agent_internal_allowed", "agent_foreign_message_refused",
	} {
		t.Run("media_business_auth_"+scenario, func(t *testing.T) {
			query := StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: baseline.ContractSHA256, Operation: "conversation_attachment", ConversationID: "stopped-recovery-member", MessageID: memberMessageID}
			username, claimRole, tokenUse := "product-backup-customer", "user", "access"
			tokenSecret := secret
			expectedStatus, expectedBody := http.StatusOK, "private-original-fixture-v1"
			isAgent := strings.HasPrefix(scenario, "agent_")
			if isAgent {
				query.Operation = "agent_conversation_attachment"
				query.ConversationID = conversationID
				username = "recovery-assigned-support"
				// A stale low privilege claim must use the current support role.
				claimRole = "user"
			}
			switch scenario {
			case "member_thumbnail", "agent_thumbnail":
				query.Preset = "thumbnail"
				expectedBody = "private-thumbnail-fixture"
			case "member_range", "agent_range":
				query.Range = "bytes=0-7"
				expectedStatus = http.StatusPartialContent
				expectedBody = expectedBody[:8]
			case "member_foreign":
				username = "recovery-other-member"
				expectedStatus = http.StatusNotFound
			case "member_anonymous":
				expectedStatus = http.StatusNotFound
			case "agent_anonymous":
				expectedStatus = http.StatusUnauthorized
			case "member_internal_refused":
				query.MessageID = internalMessageID
				expectedStatus = http.StatusNotFound
			case "agent_internal_allowed":
				query.MessageID = internalMessageID
			case "member_foreign_message_refused", "agent_foreign_message_refused":
				query.MessageID = guestMessageID
				expectedStatus = http.StatusNotFound
			case "agent_unassigned_stale_admin_claim":
				username = "recovery-unassigned-support"
				claimRole = "admin"
				expectedStatus = http.StatusNotFound
			case "agent_current_manager":
				username = "recovery-current-manager"
			case "agent_current_admin":
				username = "recovery-current-admin"
			case "agent_editor_stale_admin_claim":
				username = "recovery-current-editor"
				claimRole = "admin"
				expectedStatus = http.StatusForbidden
			case "agent_storefront_account_stale_admin_claim":
				username = "product-backup-customer"
				claimRole = "admin"
				expectedStatus = http.StatusForbidden
			case "member_inactive", "agent_inactive":
				username = "recovery-inactive-support"
				expectedStatus = http.StatusUnauthorized
			case "member_deleted", "agent_deleted":
				username = "recovery-deleted-support"
				expectedStatus = http.StatusUnauthorized
			case "member_refresh", "agent_refresh":
				tokenUse = "refresh"
				expectedStatus = http.StatusUnauthorized
			case "member_wrong_key", "agent_wrong_key":
				tokenSecret = "synthetic-wrong-key-for-recovery-tests-only"
				expectedStatus = http.StatusUnauthorized
			case "member_expired", "agent_expired", "member_tampered", "agent_tampered":
				expectedStatus = http.StatusUnauthorized
			}
			if !isAgent && expectedStatus == http.StatusUnauthorized {
				expectedStatus = http.StatusNotFound
			}
			if !isAgent && (scenario == "member_inactive" || scenario == "member_deleted") {
				// This credential actually owns the requested conversation. A 404
				// therefore cannot pass merely because the fixture chose a foreign owner.
				query.ConversationID = "stopped-" + username
				query.MessageID = unavailableMemberMessages[username]
			}
			if expectedStatus >= 400 {
				expectedBody = ""
			}
			id, exists := identities[username]
			require.True(t, exists)
			credential := signSyntheticStoppedProductMediaBusinessAccessCookie(t, id, claimRole, tokenUse, tokenSecret, strings.HasSuffix(scenario, "_expired"))
			if strings.HasSuffix(scenario, "_tampered") {
				credential += "tampered"
			}
			if strings.HasSuffix(scenario, "_anonymous") {
				credential = ""
			}
			if isAgent {
				query.AdminAuthCookie = credential
			} else {
				query.StorefrontAuthCookie = credential
			}
			response := invokeHTTP("product-catalog-stopped-read-only-media-http-query", query, http.StatusOK)
			if credential != "" {
				require.NotContains(t, response.Body.String(), credential)
			}
			require.NotContains(t, response.Body.String(), "private-original-fixture-v1")
			var payload struct {
				Result StoppedProductReadOnlyMediaHTTPResult
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
			result := payload.Result
			require.Equal(t, 2, result.Version)
			require.Equal(t, expectedStatus, result.HTTPStatus)
			require.Empty(t, result.Target.VisitorCookie)
			require.Empty(t, result.Target.StorefrontAuthCookie)
			require.Empty(t, result.Target.AdminAuthCookie)
			require.Equal(t, query.StorefrontAuthCookie != "", result.StorefrontCredentialSupplied)
			require.Equal(t, query.AdminAuthCookie != "", result.AdminCredentialSupplied)
			require.False(t, result.VisitorCredentialSupplied)
			require.Equal(t, baseline.ContractSHA256, result.ContractSHA256)
			require.True(t, result.SQLReadOnly)
			require.True(t, result.ApplicationRemainsStopped)
			require.False(t, result.OnlineOverwriteAvailable)
			if expectedStatus < 400 || expectedStatus == http.StatusNotFound {
				require.Equal(t, int64(len(expectedBody)), result.ResponseBytes)
				require.Equal(t, productCatalogBytesSHA256([]byte(expectedBody)), result.ResponseSHA256)
				require.Equal(t, "private, no-store", result.Headers["Cache-Control"])
			}
		})
	}
	// Optional authentication must retain the actual guest fallback behavior.
	guest := StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: baseline.ContractSHA256, Operation: "conversation_attachment", ConversationID: "stopped-recovery-guest", MessageID: guestMessageID,
		VisitorCookie: visitorcookie.SignCustomerServiceVisitorSessionID(stoppedProductMediaSyntheticVisitorSessionID, []byte(secret)), StorefrontAuthCookie: "invalid-existing-member-token"}
	var guestPayload struct {
		Result StoppedProductReadOnlyMediaHTTPResult
	}
	require.NoError(t, json.Unmarshal(invokeHTTP("product-catalog-stopped-read-only-media-http-query", guest, http.StatusOK).Body.Bytes(), &guestPayload))
	require.Equal(t, http.StatusOK, guestPayload.Result.HTTPStatus)
	after, err := manager.GetTask(taskID)
	require.NoError(t, err)
	require.Equal(t, before, after)
	current, err := manager.InspectStoppedProductStorefrontProjectionBaseline(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, baseline.ContractSHA256, current.ContractSHA256)
}

func TestStoppedProductMediaBusinessCookieScopesAndBoundsRejectCrossSurfaceCredentials(t *testing.T) {
	valid := StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: strings.Repeat("a", 64), Operation: "conversation_attachment", ConversationID: "member-chat", MessageID: "42"}
	for _, scenario := range []string{"public_storefront_cookie", "public_admin_cookie", "customer_admin_cookie", "agent_visitor_cookie", "agent_storefront_cookie", "agent_non_numeric_id", "oversized_storefront_cookie", "oversized_admin_cookie", "cookie_header_injection", "cookie_multiple_values", "cookie_non_ascii_transformation"} {
		t.Run(scenario, func(t *testing.T) {
			request := valid
			switch scenario {
			case "public_storefront_cookie", "public_admin_cookie":
				request = StoppedProductReadOnlyMediaHTTPRequest{ContractSHA256: valid.ContractSHA256, Operation: "public_upload", ObjectKey: "products/image.png"}
				if scenario == "public_storefront_cookie" {
					request.StorefrontAuthCookie = "synthetic"
				} else {
					request.AdminAuthCookie = "synthetic"
				}
			case "customer_admin_cookie":
				request.AdminAuthCookie = "synthetic"
			case "agent_visitor_cookie", "agent_storefront_cookie", "agent_non_numeric_id":
				request.Operation = "agent_conversation_attachment"
				request.ConversationID = "42"
				if scenario == "agent_visitor_cookie" {
					request.VisitorCookie = "synthetic"
				}
				if scenario == "agent_storefront_cookie" {
					request.StorefrontAuthCookie = "synthetic"
				}
				if scenario == "agent_non_numeric_id" {
					request.ConversationID = "chat-id"
				}
			case "oversized_storefront_cookie":
				request.StorefrontAuthCookie = strings.Repeat("a", 4097)
			case "oversized_admin_cookie":
				request.AdminAuthCookie = strings.Repeat("a", 4097)
			case "cookie_header_injection":
				request.StorefrontAuthCookie = "synthetic\r\nX-Injected: yes"
			case "cookie_multiple_values":
				request.StorefrontAuthCookie = "first; admin_auth_token=other"
			case "cookie_non_ascii_transformation":
				request.StorefrontAuthCookie = "不得在添加 Cookie 时自动删除字符"
			}
			require.Error(t, validateStoppedProductReadOnlyMediaHTTPRequest(request))
		})
	}
}

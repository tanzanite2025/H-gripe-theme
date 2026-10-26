package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOutboxFailureTransactionalAuditHTTPBindsAuthenticatedActorAndReturnsCommittedReceipt(t *testing.T) {
	for _, operation := range []string{"ignore", "retry"} {
		t.Run(operation, func(t *testing.T) {
			router, repo, db := createDashboardOutboxFailureClosedLoopTestRouter(t)
			event := outbox.Event{EventKey: "synthetic-failure-http", EventType: outbox.EventTypeCustomerServiceRealtime,
				AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12",
				Payload: []byte(`{"status":"open","private":"synthetic-private-event-body"}`),
				Status:  outbox.EventStatusDeadLetter, Attempts: 10, LastError: "synthetic-private-original-failure"}
			require.NoError(t, repo.CreateEvent(&event))
			endpoint := fmt.Sprintf("/api/admin/ops/outbox/failures/%d/%s", event.ID, operation)
			body := `{"note":"人工核对原失败事件","user_id":99999,"resource_id":99999,"old_value":"forged","new_value":{"operation":"forged"},"status":"processed"}`
			recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodPost, endpoint, string(auth.RoleAdmin), body)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			var response struct {
				Data outbox.FailureDispositionResult `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			require.Positive(t, response.Data.AuditID)
			var records []audit.AuditLog
			require.NoError(t, db.Find(&records).Error)
			require.Len(t, records, 1)
			require.Equal(t, uint(12), records[0].UserID)
			require.Equal(t, event.ID, records[0].ResourceID)
			require.Equal(t, endpoint, records[0].Path)
			require.Equal(t, "POST", records[0].Method)
			require.Equal(t, "success", records[0].Status)
			require.Equal(t, response.Data.AuditID, records[0].ID)
			var receipt outbox.FailureDispositionAuditReceipt
			require.NoError(t, json.Unmarshal([]byte(records[0].NewValue), &receipt))
			require.Equal(t, operation, receipt.Operation)
			require.Equal(t, "人工核对原失败事件", receipt.Note)
			require.Equal(t, event.EventKey, receipt.Event.EventKey)
			for _, privateValue := range []string{"synthetic-private-event-body", "synthetic-private-original-failure", event.EventKey, "row_sha256"} {
				require.NotContains(t, recorder.Body.String(), privateValue)
			}
			for _, privateValue := range []string{"synthetic-private-event-body", "synthetic-private-original-failure", "forged"} {
				require.NotContains(t, records[0].OldValue+records[0].NewValue, privateValue)
			}
			after, err := repo.FindEventByID(event.ID)
			require.NoError(t, err)
			if operation == "ignore" {
				require.Equal(t, outbox.EventStatusIgnored, after.Status)
				require.Equal(t, event.LastError, after.LastError)
			} else {
				require.Equal(t, outbox.EventStatusPending, after.Status)
			}
		})
	}
}

func TestOutboxFailureTransactionalAuditHTTPDoesNotCommitWhenAuditPersistenceFails(t *testing.T) {
	for _, operation := range []string{"ignore", "retry"} {
		t.Run(operation, func(t *testing.T) {
			router, repo, db := createDashboardOutboxFailureClosedLoopTestRouter(t)
			event := outbox.Event{EventKey: "synthetic-failure-audit-rollback", EventType: outbox.EventTypeCustomerServiceRealtime,
				AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: outbox.EventStatusDeadLetter, Attempts: 10,
				Payload: []byte(`{"status":"open"}`), LastError: "synthetic-original-error"}
			require.NoError(t, repo.CreateEvent(&event))
			before, err := repo.FindEventByID(event.ID)
			require.NoError(t, err)
			require.NoError(t, db.Exec(`CREATE TRIGGER reject_synthetic_audit_insert BEFORE INSERT ON audit_logs BEGIN SELECT RAISE(ABORT, 'synthetic-private-database-failure'); END`).Error)
			endpoint := fmt.Sprintf("/api/admin/ops/outbox/failures/%d/%s", event.ID, operation)
			recorder := requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodPost, endpoint, string(auth.RoleAdmin), `{"note":"reviewed historical failure"}`)
			require.Equal(t, http.StatusServiceUnavailable, recorder.Code, recorder.Body.String())
			require.NotContains(t, recorder.Body.String(), "synthetic-private-database-failure")
			require.NotContains(t, recorder.Body.String(), "audit_id")
			after, err := repo.FindEventByID(event.ID)
			require.NoError(t, err)
			require.Equal(t, before, after)
			var count int64
			require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestOutboxFailureTransactionalAuditHTTPRejectsLeaseConflictInvalidNotesAndMissingActor(t *testing.T) {
	router, repo, db := createDashboardOutboxFailureClosedLoopTestRouter(t)
	event := outbox.Event{EventKey: "synthetic-failure-rejections", EventType: outbox.EventTypeCustomerServiceRealtime,
		AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "12", Status: outbox.EventStatusDeadLetter, LockedBy: "synthetic-worker"}
	require.NoError(t, repo.CreateEvent(&event))
	endpoint := fmt.Sprintf("/api/admin/ops/outbox/failures/%d/ignore", event.ID)
	require.Equal(t, http.StatusConflict, requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodPost, endpoint, string(auth.RoleAdmin), `{"note":"reviewed failure"}`).Code)
	for _, body := range []string{`{"note":"xx"}`, `{"note":"` + strings.Repeat("长", 2001) + `"}`, `{not json}`} {
		require.Equal(t, http.StatusBadRequest, requestDashboardOutboxFailureClosedLoopRoute(router, http.MethodPost, endpoint, string(auth.RoleAdmin), body).Code)
	}
	missingActorRouter := gin.New()
	missingActorRouter.POST("/api/admin/ops/outbox/failures/:id/ignore", NewOutboxReconciliationHandler(service.NewOutboxService(repository.NewOutboxRepository(db))).IgnoreFailure)
	require.Equal(t, http.StatusUnauthorized, requestDashboardOutboxFailureClosedLoopRoute(missingActorRouter, http.MethodPost, endpoint, "", `{"note":"reviewed failure","user_id":99999}`).Code)
	var count int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
	require.Zero(t, count)
}

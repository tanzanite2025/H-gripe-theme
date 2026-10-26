package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createFailedOutboxTransactionalDispositionAuditTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, db.AutoMigrate(&outbox.Event{}, &audit.AuditLog{}))
	return db
}

func createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t *testing.T, db *gorm.DB) outbox.Event {
	t.Helper()
	lastAttempt := time.Now().UTC().Add(-time.Hour)
	event := outbox.Event{
		EventKey: "customer_service.conversation.created:123", EventType: outbox.EventTypeCustomerServiceRealtime,
		AggregateType: outbox.AggregateTypeCustomerServiceConversation, AggregateID: "123",
		Payload: []byte(`{"status":"open","assigned_to":123,"private_body":"synthetic-private-payload"}`),
		Status:  outbox.EventStatusDeadLetter, Attempts: 10, MaxAttempts: 10,
		AvailableAt: lastAttempt, LastAttemptAt: &lastAttempt, LastError: "synthetic-private-original-failure",
	}
	require.NoError(t, db.Create(&event).Error)
	var saved outbox.Event
	require.NoError(t, db.First(&saved, event.ID).Error)
	return saved
}

func TestFailedOutboxTransactionalDispositionAuditCommitsOriginalEvidenceAndExactlyOneReceipt(t *testing.T) {
	for _, operation := range []string{"ignore", "retry"} {
		for _, status := range []string{outbox.EventStatusFailed, outbox.EventStatusDeadLetter} {
			t.Run(operation+"_"+status, func(t *testing.T) {
				db := createFailedOutboxTransactionalDispositionAuditTestDatabase(t)
				event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
				require.NoError(t, db.Model(&event).Update("status", status).Error)
				require.NoError(t, db.First(&event, event.ID).Error)
				changedAt := time.Now().UTC().Add(time.Second)
				actor := &audit.AuditLog{UserID: 42, Username: "synthetic-operator", Method: "POST", Path: "/api/admin/ops/outbox/failures/1/" + operation, Action: "forged", ResourceID: 9000, NewValue: "forged"}
				result, err := NewOutboxRepository(db).ReconcileFailedEventWithTransactionalAudit(event.ID, operation, " reviewed historical failure ", changedAt, actor)
				require.NoError(t, err)
				require.Positive(t, result.AuditID)
				var after outbox.Event
				require.NoError(t, db.First(&after, event.ID).Error)
				expected := event
				expected.Status, expected.UpdatedAt = outbox.EventStatusIgnored, changedAt
				if operation == "retry" {
					expected.Status, expected.Attempts, expected.AvailableAt = outbox.EventStatusPending, 0, changedAt
					expected.LastError = "manual retry: reviewed historical failure"
				}
				require.Equal(t, expected, after)
				require.Equal(t, after.Status, result.Status)
				var records []audit.AuditLog
				require.NoError(t, db.Find(&records).Error)
				require.Len(t, records, 1)
				record := records[0]
				require.Equal(t, result.AuditID, record.ID)
				require.Equal(t, uint(42), record.UserID)
				require.Equal(t, event.ID, record.ResourceID)
				require.Equal(t, "outbox_event", record.Resource)
				require.Equal(t, "reconcile", record.Action)
				require.Equal(t, "success", record.Status)
				var previous outbox.FailureDispositionAuditSnapshot
				var receipt outbox.FailureDispositionAuditReceipt
				require.NoError(t, json.Unmarshal([]byte(record.OldValue), &previous))
				require.NoError(t, json.Unmarshal([]byte(record.NewValue), &receipt))
				require.Equal(t, outbox.FailureDispositionAuditContract, receipt.Contract)
				require.Equal(t, operation, receipt.Operation)
				require.Equal(t, "reviewed historical failure", receipt.Note)
				require.Equal(t, event.EventKey, previous.EventKey)
				require.Equal(t, previous.EventKey, receipt.Event.EventKey)
				require.Equal(t, previous.PayloadSHA256, receipt.Event.PayloadSHA256)
				digest := sha256.Sum256(event.Payload)
				require.Equal(t, hex.EncodeToString(digest[:]), previous.PayloadSHA256)
				require.Len(t, previous.RowSHA256, 64)
				require.NotEqual(t, previous.RowSHA256, receipt.Event.RowSHA256)
				require.Equal(t, "domain_event_json_sha256_v1", previous.RowFingerprintFormat)
				encoded, err := json.Marshal(record)
				require.NoError(t, err)
				for _, privateValue := range []string{"synthetic-private-payload", "synthetic-private-original-failure"} {
					require.NotContains(t, string(encoded), privateValue)
				}
				_, err = NewOutboxRepository(db).ReconcileFailedEventWithTransactionalAudit(event.ID, operation, "second concurrent decision", changedAt.Add(time.Second), actor)
				require.ErrorIs(t, err, ErrOutboxInvalidTransition)
				var count int64
				require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
				require.EqualValues(t, 1, count)
			})
		}
	}
}

func TestFailedOutboxTransactionalDispositionAuditRollsBackEventAndReceiptWhenEitherWriteFails(t *testing.T) {
	for _, operation := range []string{"ignore", "retry"} {
		for _, failure := range []string{"audit_insert", "event_update", "missing_audit_table"} {
			t.Run(operation+"_"+failure, func(t *testing.T) {
				db := createFailedOutboxTransactionalDispositionAuditTestDatabase(t)
				event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
				switch failure {
				case "audit_insert":
					require.NoError(t, db.Exec(`CREATE TRIGGER reject_synthetic_audit BEFORE INSERT ON audit_logs BEGIN SELECT RAISE(ABORT, 'synthetic-private-audit-error'); END`).Error)
				case "event_update":
					require.NoError(t, db.Exec(`CREATE TRIGGER reject_synthetic_event_update BEFORE UPDATE ON outbox_events BEGIN SELECT RAISE(ABORT, 'synthetic event update error'); END`).Error)
				case "missing_audit_table":
					require.NoError(t, db.Migrator().DropTable(&audit.AuditLog{}))
				}
				result, err := NewOutboxRepository(db).ReconcileFailedEventWithTransactionalAudit(event.ID, operation, "reviewed historical failure", time.Now(), &audit.AuditLog{UserID: 42})
				require.Error(t, err)
				require.Zero(t, result.AuditID)
				if failure != "event_update" {
					require.ErrorIs(t, err, ErrOutboxDispositionAuditWrite)
				}
				var after outbox.Event
				require.NoError(t, db.First(&after, event.ID).Error)
				require.Equal(t, event, after)
				if failure != "missing_audit_table" {
					var count int64
					require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
					require.Zero(t, count)
				}
			})
		}
	}
}

func TestFailedOutboxTransactionalDispositionAuditRejectsLeasesUnknownResultsAndStaleDecisions(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		changes map[string]interface{}
	}{
		{"pending", map[string]interface{}{"status": "pending"}},
		{"processing", map[string]interface{}{"status": "processing"}},
		{"processed", map[string]interface{}{"status": "processed"}},
		{"ignored", map[string]interface{}{"status": "ignored"}},
		{"unknown", map[string]interface{}{"status": "unknown"}},
		{"lock_timestamp", map[string]interface{}{"locked_at": time.Now()}},
		{"lock_owner", map[string]interface{}{"locked_by": "synthetic-worker"}},
		{"uncertain_result", map[string]interface{}{"uncertain_at": time.Now()}},
		{"reconciliation_schedule", map[string]interface{}{"reconcile_after": time.Now()}},
		{"previous_processing_receipt", map[string]interface{}{"processed_at": time.Now()}},
		{"updated_after_request_began", map[string]interface{}{"updated_at": time.Now().Add(time.Hour)}},
		{"soft_deleted", map[string]interface{}{"deleted_at": time.Now()}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			db := createFailedOutboxTransactionalDispositionAuditTestDatabase(t)
			event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
			require.NoError(t, db.Model(&event).Updates(scenario.changes).Error)
			var before outbox.Event
			require.NoError(t, db.Unscoped().First(&before, event.ID).Error)
			for _, operation := range []string{"ignore", "retry"} {
				_, err := NewOutboxRepository(db).ReconcileFailedEventWithTransactionalAudit(event.ID, operation, "reviewed historical failure", time.Now(), &audit.AuditLog{UserID: 42})
				if scenario.name == "soft_deleted" {
					require.ErrorIs(t, err, ErrOutboxEventNotFound)
				} else {
					require.ErrorIs(t, err, ErrOutboxInvalidTransition)
				}
			}
			var after outbox.Event
			require.NoError(t, db.Unscoped().First(&after, event.ID).Error)
			require.Equal(t, before, after)
			var count int64
			require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestFailedOutboxTransactionalDispositionAuditRequiresActorBoundedNoteAndLiveRequestContext(t *testing.T) {
	db := createFailedOutboxTransactionalDispositionAuditTestDatabase(t)
	event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
	repo := NewOutboxRepository(db)
	for _, actor := range []*audit.AuditLog{nil, {}} {
		_, err := repo.ReconcileFailedEventWithTransactionalAudit(event.ID, "ignore", "reviewed failure", time.Now(), actor)
		require.ErrorIs(t, err, ErrOutboxDispositionAuditRequired)
	}
	for _, note := range []string{"", "xx", strings.Repeat("长", 2001)} {
		_, err := repo.ReconcileFailedEventWithTransactionalAudit(event.ID, "ignore", note, time.Now(), &audit.AuditLog{UserID: 42})
		require.ErrorIs(t, err, ErrOutboxDispositionInputInvalid)
	}
	_, err := repo.ReconcileFailedEventWithTransactionalAudit(event.ID, "processed", "reviewed failure", time.Now(), &audit.AuditLog{UserID: 42})
	require.ErrorIs(t, err, ErrOutboxDispositionInputInvalid)
	_, err = repo.ReconcileFailedEventWithTransactionalAudit(999999, "ignore", "reviewed failure", time.Now(), &audit.AuditLog{UserID: 42})
	require.ErrorIs(t, err, ErrOutboxEventNotFound)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = repo.WithContext(ctx).ReconcileFailedEventWithTransactionalAudit(event.ID, "ignore", "reviewed failure", time.Now(), &audit.AuditLog{UserID: 42})
	require.ErrorIs(t, err, context.Canceled)
	var after outbox.Event
	require.NoError(t, db.First(&after, event.ID).Error)
	require.Equal(t, event, after)
	var count int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
	require.Zero(t, count)
}

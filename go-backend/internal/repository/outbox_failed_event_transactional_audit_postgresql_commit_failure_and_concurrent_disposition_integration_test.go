package repository

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createExplicitlyDisposablePostgreSQLFailedOutboxTransactionalAuditDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	connection := os.Getenv("OUTBOX_AUDIT_DISPOSABLE_POSTGRESQL_URL")
	if connection == "" || os.Getenv("OUTBOX_AUDIT_DISPOSABLE_TEST_SERVER") != "true" {
		t.Skip("requires the dedicated disposable outbox audit PostgreSQL server")
	}
	parsed, err := url.Parse(connection)
	require.NoError(t, err)
	require.Equal(t, "outbox-audit-postgresql", parsed.Hostname())
	require.Equal(t, "/postgres", parsed.Path)
	control, err := gorm.Open(postgres.Open(connection), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	controlPool, err := control.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, controlPool.Close()) })
	schema := "outbox_disposition_audit_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, control.Exec("CREATE SCHEMA "+pq.QuoteIdentifier(schema)).Error)
	t.Cleanup(func() { require.NoError(t, control.Exec("DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE").Error) })
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	pool.SetMaxOpenConns(12)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })
	require.NoError(t, db.AutoMigrate(&outbox.Event{}, &audit.AuditLog{}))
	return db
}

func assertPostgreSQLFailedOutboxDispositionEventAndAuditCount(t *testing.T, db *gorm.DB, expected outbox.Event, count int64) {
	t.Helper()
	var current outbox.Event
	require.NoError(t, db.First(&current, expected.ID).Error)
	require.Equal(t, expected, current)
	var actualCount int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&actualCount).Error)
	require.Equal(t, count, actualCount)
}

func waitForPostgreSQLFailedOutboxDispositionRowLockWaiter(t *testing.T, db *gorm.DB, blockerPID int) {
	t.Helper()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE ?=ANY(pg_blocking_pids(pid)))", blockerPID).Scan(&waiting).Error
		require.NoError(t, err)
		return waiting
	}, 5*time.Second, 10*time.Millisecond, "the competing disposition must wait on the existing event row lock")
}

func TestPostgreSQLFailedOutboxTransactionalAuditBindsCompletePersistedRowsIncludingUnmodeledColumns(t *testing.T) {
	for _, operation := range []string{"ignore", "retry"} {
		t.Run(operation, func(t *testing.T) {
			db := createExplicitlyDisposablePostgreSQLFailedOutboxTransactionalAuditDatabase(t)
			require.NoError(t, db.Exec(`ALTER TABLE outbox_events ADD COLUMN synthetic_preserved_extension TEXT NOT NULL DEFAULT 'synthetic-private-extension'`).Error)
			event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
			var originalFingerprint string
			require.NoError(t, db.Raw(`SELECT encode(sha256(convert_to(to_jsonb(event)::text,'UTF8')),'hex') FROM outbox_events event WHERE id=?`, event.ID).Scan(&originalFingerprint).Error)
			result, err := NewOutboxRepository(db).ReconcileFailedEventWithTransactionalAudit(event.ID, operation, "reviewed historical failure", time.Now(), &audit.AuditLog{UserID: 42})
			require.NoError(t, err)
			var record audit.AuditLog
			require.NoError(t, db.First(&record, result.AuditID).Error)
			var previous outbox.FailureDispositionAuditSnapshot
			var receipt outbox.FailureDispositionAuditReceipt
			require.NoError(t, json.Unmarshal([]byte(record.OldValue), &previous))
			require.NoError(t, json.Unmarshal([]byte(record.NewValue), &receipt))
			require.Equal(t, "postgresql_complete_row_jsonb_sha256_v1", previous.RowFingerprintFormat)
			require.Equal(t, originalFingerprint, previous.RowSHA256)
			var afterFingerprint string
			require.NoError(t, db.Raw(`SELECT encode(sha256(convert_to(to_jsonb(event)::text,'UTF8')),'hex') FROM outbox_events event WHERE id=?`, event.ID).Scan(&afterFingerprint).Error)
			require.Equal(t, afterFingerprint, receipt.Event.RowSHA256)
			require.NotEqual(t, originalFingerprint, afterFingerprint)
			require.Equal(t, previous.PayloadSHA256, receipt.Event.PayloadSHA256)
			require.NotContains(t, record.OldValue+record.NewValue, "synthetic-private-extension")
			if operation == "ignore" {
				var after outbox.Event
				require.NoError(t, db.First(&after, event.ID).Error)
				expected := event
				expected.Status, expected.UpdatedAt = outbox.EventStatusIgnored, after.UpdatedAt
				assertPostgreSQLFailedOutboxDispositionEventAndAuditCount(t, db, expected, 1)
			}
		})
	}
}

func TestPostgreSQLFailedOutboxTransactionalAuditUsesCanonicalUTCCompleteRowFingerprintsAcrossSessionTimezones(t *testing.T) {
	for _, operation := range []string{"ignore", "retry"} {
		t.Run(operation, func(t *testing.T) {
			db := createExplicitlyDisposablePostgreSQLFailedOutboxTransactionalAuditDatabase(t)
			event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
			require.NoError(t, db.Connection(func(connection *gorm.DB) error {
				connection = connection.Session(&gorm.Session{NewDB: true})
				defer connection.Exec("SET TIME ZONE DEFAULT")
				const completeRowFingerprintQuery = `SELECT encode(sha256(convert_to(to_jsonb(event)::text,'UTF8')),'hex') FROM outbox_events event WHERE event.id=?`
				require.NoError(t, connection.Exec("SET TIME ZONE 'UTC'").Error)
				var originalUTC, originalOtherTimezone string
				require.NoError(t, connection.Raw(completeRowFingerprintQuery, event.ID).Scan(&originalUTC).Error)
				require.NoError(t, connection.Exec("SET TIME ZONE 'Asia/Shanghai'").Error)
				require.NoError(t, connection.Raw(completeRowFingerprintQuery, event.ID).Scan(&originalOtherTimezone).Error)
				require.NotEqual(t, originalUTC, originalOtherTimezone, "The fixture must expose timezone-dependent PostgreSQL JSON rendering")
				result, err := NewOutboxRepository(connection).ReconcileFailedEventWithTransactionalAudit(event.ID, operation, "reviewed timezone-independent failure", time.Now(), &audit.AuditLog{UserID: 42})
				require.NoError(t, err)
				var originalTimezone string
				require.NoError(t, connection.Raw("SELECT current_setting('TimeZone')").Scan(&originalTimezone).Error)
				require.Equal(t, "Asia/Shanghai", originalTimezone, "Canonical hashing must not change the caller's session timezone after commit")
				var record audit.AuditLog
				require.NoError(t, connection.First(&record, result.AuditID).Error)
				var previous outbox.FailureDispositionAuditSnapshot
				var receipt outbox.FailureDispositionAuditReceipt
				require.NoError(t, json.Unmarshal([]byte(record.OldValue), &previous))
				require.NoError(t, json.Unmarshal([]byte(record.NewValue), &receipt))
				require.Equal(t, originalUTC, previous.RowSHA256)
				require.NoError(t, connection.Exec("SET TIME ZONE 'UTC'").Error)
				var finalUTC string
				require.NoError(t, connection.Raw(completeRowFingerprintQuery, event.ID).Scan(&finalUTC).Error)
				require.Equal(t, finalUTC, receipt.Event.RowSHA256)
				return nil
			}))
		})
	}
}

func TestPostgreSQLFailedOutboxTransactionalAuditRollsBackOnAuditInsertAndDeferredCommitFailure(t *testing.T) {
	for _, operation := range []string{"ignore", "retry"} {
		for _, failure := range []string{"insert", "commit"} {
			t.Run(operation+"_"+failure, func(t *testing.T) {
				db := createExplicitlyDisposablePostgreSQLFailedOutboxTransactionalAuditDatabase(t)
				event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
				require.NoError(t, db.Exec(`CREATE FUNCTION reject_synthetic_disposition_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic audit persistence failure'; END $$`).Error)
				trigger := `CREATE TRIGGER reject_synthetic_audit BEFORE INSERT ON audit_logs FOR EACH ROW EXECUTE FUNCTION reject_synthetic_disposition_audit()`
				if failure == "commit" {
					trigger = `CREATE CONSTRAINT TRIGGER reject_synthetic_audit AFTER INSERT ON audit_logs DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_synthetic_disposition_audit()`
				}
				require.NoError(t, db.Exec(trigger).Error)
				result, err := NewOutboxRepository(db).ReconcileFailedEventWithTransactionalAudit(event.ID, operation, "reviewed failure", time.Now(), &audit.AuditLog{UserID: 42})
				require.Error(t, err)
				require.Zero(t, result.AuditID)
				assertPostgreSQLFailedOutboxDispositionEventAndAuditCount(t, db, event, 0)
			})
		}
	}
}

func TestPostgreSQLFailedOutboxTransactionalAuditAllowsOnlyOneConcurrentManualDecision(t *testing.T) {
	db := createExplicitlyDisposablePostgreSQLFailedOutboxTransactionalAuditDatabase(t)
	event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	start := make(chan struct{})
	results := make(chan error, 8)
	for index := range 8 {
		go func() {
			<-start
			operation := "ignore"
			if index%2 == 1 {
				operation = "retry"
			}
			_, err := NewOutboxRepository(db).WithContext(ctx).ReconcileFailedEventWithTransactionalAudit(event.ID, operation, "reviewed concurrent failure", time.Now(), &audit.AuditLog{UserID: uint(index + 1)})
			results <- err
		}()
	}
	close(start)
	successes := 0
	for range 8 {
		select {
		case err := <-results:
			if err == nil {
				successes++
			} else {
				require.ErrorIs(t, err, ErrOutboxInvalidTransition)
			}
		case <-ctx.Done():
			t.Fatal("concurrent manual dispositions did not finish")
		}
	}
	require.Equal(t, 1, successes)
	var count int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestPostgreSQLFailedOutboxTransactionalAuditHoldsWorkerAndManualRetryUntilAtomicCommit(t *testing.T) {
	db := createExplicitlyDisposablePostgreSQLFailedOutboxTransactionalAuditDatabase(t)
	event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
	require.NoError(t, db.Model(&event).Updates(map[string]interface{}{"status": outbox.EventStatusFailed, "attempts": 1}).Error)
	require.NoError(t, db.First(&event, event.ID).Error)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	blockedAtAudit, release := make(chan int, 1), make(chan struct{})
	var once sync.Once
	releaseAudit := func() { once.Do(func() { close(release) }) }
	defer releaseAudit()
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("pause_synthetic_disposition_audit", func(tx *gorm.DB) {
		if tx.Statement.Table != "audit_logs" {
			return
		}
		var pid int
		if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
			tx.AddError(err)
			return
		}
		blockedAtAudit <- pid
		select {
		case <-release:
		case <-ctx.Done():
			tx.AddError(ctx.Err())
		}
	}))
	defer db.Callback().Create().Remove("pause_synthetic_disposition_audit")
	ignoreDone := make(chan error, 1)
	go func() {
		_, err := NewOutboxRepository(db).WithContext(ctx).ReconcileFailedEventWithTransactionalAudit(event.ID, "ignore", "reviewed ignore decision", time.Now(), &audit.AuditLog{UserID: 42})
		ignoreDone <- err
	}()
	var pid int
	select {
	case pid = <-blockedAtAudit:
	case <-ctx.Done():
		t.Fatal("disposition did not reach its in-transaction audit insert")
	}
	assertPostgreSQLFailedOutboxDispositionEventAndAuditCount(t, db, event, 0)
	claimed, err := NewOutboxRepository(db).WithContext(ctx).ClaimReadyEvents(time.Now(), "synthetic-worker", 10, time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed, "worker must skip the event locked by its pending audit transaction")
	retryDone := make(chan error, 1)
	go func() {
		_, err := NewOutboxRepository(db).WithContext(ctx).ReconcileFailedEventWithTransactionalAudit(event.ID, "retry", "reviewed concurrent retry", time.Now(), &audit.AuditLog{UserID: 43})
		retryDone <- err
	}()
	waitForPostgreSQLFailedOutboxDispositionRowLockWaiter(t, db, pid)
	releaseAudit()
	require.NoError(t, <-ignoreDone)
	require.ErrorIs(t, <-retryDone, ErrOutboxInvalidTransition)
	var after outbox.Event
	require.NoError(t, db.First(&after, event.ID).Error)
	expected := event
	expected.Status, expected.UpdatedAt = outbox.EventStatusIgnored, after.UpdatedAt
	assertPostgreSQLFailedOutboxDispositionEventAndAuditCount(t, db, expected, 1)
}

func TestPostgreSQLFailedOutboxTransactionalAuditRechecksConcurrentRetryLeaseAndNewFailureAfterRowLock(t *testing.T) {
	for _, scenario := range []string{"retry", "worker_lease", "new_failure"} {
		t.Run(scenario, func(t *testing.T) {
			db := createExplicitlyDisposablePostgreSQLFailedOutboxTransactionalAuditDatabase(t)
			event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			tx := db.WithContext(ctx).Begin()
			require.NoError(t, tx.Error)
			defer tx.Rollback()
			require.NoError(t, tx.Exec("SELECT id FROM outbox_events WHERE id=? FOR UPDATE", event.ID).Error)
			var pid int
			require.NoError(t, tx.Raw("SELECT pg_backend_pid()").Scan(&pid).Error)
			requestAt := time.Now().UTC()
			done := make(chan error, 1)
			go func() {
				_, err := NewOutboxRepository(db).WithContext(ctx).ReconcileFailedEventWithTransactionalAudit(event.ID, "ignore", "reviewed older failure", requestAt, &audit.AuditLog{UserID: 42})
				done <- err
			}()
			waitForPostgreSQLFailedOutboxDispositionRowLockWaiter(t, db, pid)
			updates := map[string]interface{}{"updated_at": requestAt.Add(time.Millisecond)}
			switch scenario {
			case "retry":
				updates["status"] = outbox.EventStatusPending
			case "worker_lease":
				updates["locked_at"], updates["locked_by"] = requestAt, "synthetic-worker"
			case "new_failure":
				updates["last_error"] = "synthetic newer failure"
			}
			require.NoError(t, tx.Model(&outbox.Event{}).Where("id=?", event.ID).Updates(updates).Error)
			require.NoError(t, tx.Commit().Error)
			var current outbox.Event
			require.NoError(t, db.First(&current, event.ID).Error)
			require.ErrorIs(t, <-done, ErrOutboxInvalidTransition)
			assertPostgreSQLFailedOutboxDispositionEventAndAuditCount(t, db, current, 0)
		})
	}
}

func TestPostgreSQLFailedOutboxTransactionalAuditRollsBackWhenRequestIsCanceledBeforeReceiptInsert(t *testing.T) {
	db := createExplicitlyDisposablePostgreSQLFailedOutboxTransactionalAuditDatabase(t)
	event := createHistoricalFailedOutboxEventForTransactionalDispositionAudit(t, db)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("cancel_synthetic_disposition_before_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_logs" {
			cancel()
			tx.AddError(ctx.Err())
		}
	}))
	defer db.Callback().Create().Remove("cancel_synthetic_disposition_before_audit")
	result, err := NewOutboxRepository(db).WithContext(ctx).ReconcileFailedEventWithTransactionalAudit(event.ID, "ignore", "reviewed failure", time.Now(), &audit.AuditLog{UserID: 42})
	require.True(t, errors.Is(err, context.Canceled))
	require.Zero(t, result.AuditID)
	assertPostgreSQLFailedOutboxDispositionEventAndAuditCount(t, db, event, 0)
}

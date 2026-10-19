package service

import (
	"context"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"
)

func createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t *testing.T) (customerServicePrivateAttachmentAccessTestEnvironment, *CustomerServiceRetentionService, *media.MediaAsset) {
	t.Helper()
	host, username, password := os.Getenv("DB_HOST"), os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD")
	if host == "" || username == "" || password == "" {
		t.Skip("PostgreSQL integration environment is not configured")
	}
	port, databaseName := os.Getenv("DB_PORT"), os.Getenv("DB_NAME")
	if port == "" {
		port = "5432"
	}
	if databaseName == "" {
		databaseName = "postgres"
	}
	connectionURL := &url.URL{Scheme: "postgres", User: url.UserPassword(username, password), Host: net.JoinHostPort(host, port), Path: "/" + databaseName, RawQuery: "sslmode=disable&connect_timeout=5"}
	bootstrap, err := gorm.Open(postgres.Open(connectionURL.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	bootstrapSQL, err := bootstrap.DB()
	require.NoError(t, err)
	schema := "test_customer_service_attachment_retirement_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, bootstrap.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() {
		require.NoError(t, bootstrap.Exec("DROP SCHEMA "+schema+" CASCADE").Error)
		_ = bootstrapSQL.Close()
	})
	query := connectionURL.Query()
	query.Set("search_path", schema)
	connectionURL.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(connectionURL.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(8)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&user.User{}, &ticket.Ticket{}, &ticket.TicketMessage{}, &ticket.CustomerServiceInboxState{},
		&media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}, &audit.AuditLog{}))
	for _, migrationName := range []string{"420_add_customer_service_message_media_asset_bindings_and_pending_upload_sessions.up.sql", "421_add_customer_service_private_derivative_write_lease_for_safe_asset_cleanup.up.sql", "428_add_pending_customer_service_upload_expected_capacity_bytes.up.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", migrationName))
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(migration)).Error)
	}
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	var support user.User
	require.NoError(t, env.db.First(&support, env.conversation.AssignedTo).Error)
	require.NoError(t, db.Create(&support).Error)
	require.NoError(t, db.Omit("Messages", "User").Create(&env.conversation).Error)
	require.NoError(t, db.Omit("Derivatives").Create(&env.asset).Error)
	// Copy the historical fixture without issuing a new provider registration.
	// Uploads below use the configured MediaService and verify actual bytes.
	require.NoError(t, db.Create(&env.asset.Derivatives).Error)
	require.NoError(t, db.Create(&env.message).Error)
	for _, table := range []string{"users", "tickets", "ticket_messages", "media_assets", "media_asset_derivatives"} {
		require.NoError(t, db.Exec("SELECT setval(pg_get_serial_sequence('"+table+"','id'), (SELECT MAX(id) FROM "+table+"), true)").Error)
	}
	env.db = db
	env.mediaService = NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
	env.mediaService.ConfigureCustomerServicePrivateAttachmentUploads(true)
	env.mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	tickets := NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	tickets.ConfigureCustomerServiceRealtimeOutbox(repository.NewOutboxRepository(db))
	env.access = NewCustomerServiceConversationMessageAttachmentAccessService(tickets, env.mediaService)
	uploads := NewCustomerServiceConversationAttachmentUploadService(tickets, env.mediaService)
	asset, err := uploads.UploadVerifiedCustomerServiceConversationAttachment(t.Context(), *env.conversation.ConversationID, env.owner,
		multipartFileHeader(t, "postgres-pending.png", "image/png", testOpaquePNG(t, 64, 32)))
	require.NoError(t, err)
	retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
	// Binding uses the real clock; expiry observes the later scheduled cutoff.
	retention.now = func() time.Time { return time.Now().Add(48 * time.Hour) }
	return env, retention, asset
}

func TestPostgreSQLCustomerServiceMessageBindingAndUploadExpirySerializeWithoutDeletingCommittedAttachments(t *testing.T) {
	for _, firstOperation := range []string{"bind", "expire"} {
		t.Run(firstOperation+" first", func(t *testing.T) {
			env, retention, asset := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			firstLocked := make(chan int, 1)
			releaseFirst := make(chan struct{})
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
			t.Cleanup(release)
			pauseTransaction := func(tx *gorm.DB) {
				var pid int
				if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
					tx.AddError(err)
					return
				}
				firstLocked <- pid
				select {
				case <-releaseFirst:
				case <-ctx.Done():
					tx.AddError(ctx.Err())
				}
			}
			callbackName := "pause_customer_service_attachment_retirement_race_" + uuid.NewString()
			if firstOperation == "bind" {
				require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
					if tx.Statement.Table == "ticket_message_media_attachments" {
						pauseTransaction(tx)
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callbackName) })
			} else {
				require.NoError(t, env.db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
					if tx.Statement.Table == "media_upload_sessions" {
						pauseTransaction(tx)
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callbackName) })
			}
			bindResult := make(chan error, 1)
			expiryResult := make(chan CustomerServicePendingAttachmentUploadExpiryBatchResult, 1)
			expiryError := make(chan error, 1)
			bind := func() {
				_, _, err := env.access.tickets.AddPublicCustomerServiceMessage(*env.conversation.ConversationID, env.owner, "Racing photo", 0, "image", "", "",
					ticket.CustomerServiceMessageMediaAttachmentBindingRequest{UploadSessionID: asset.PendingCustomerServiceUploadSessionID})
				bindResult <- err
			}
			expire := func() {
				result, err := retention.ExpireUnboundCustomerServiceAttachmentUploads(ctx, 10)
				expiryResult <- result
				expiryError <- err
			}
			if firstOperation == "bind" {
				go bind()
			} else {
				go expire()
			}
			var blockerPID int
			select {
			case blockerPID = <-firstLocked:
			case <-ctx.Done():
				t.Fatal("first transaction did not acquire its lock")
			}
			if firstOperation == "bind" {
				go expire()
			} else {
				go bind()
			}
			require.Eventually(t, func() bool {
				var waiting bool
				err := env.db.WithContext(ctx).Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)))", blockerPID).Scan(&waiting).Error
				return err == nil && waiting
			}, 5*time.Second, 20*time.Millisecond, "second operation did not wait for the shared asset lock")
			release()
			var sendError error
			select {
			case sendError = <-bindResult:
			case <-ctx.Done():
				t.Fatal("message send did not finish")
			}
			var result CustomerServicePendingAttachmentUploadExpiryBatchResult
			select {
			case result = <-expiryResult:
			case <-ctx.Done():
				t.Fatal("upload expiry did not finish")
			}
			require.NoError(t, <-expiryError)
			var bindings, cleanupEvents int64
			require.NoError(t, env.db.Model(&ticket.TicketMessageMediaAttachment{}).Count(&bindings).Error)
			require.NoError(t, env.db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).Count(&cleanupEvents).Error)
			if firstOperation == "bind" {
				require.NoError(t, sendError)
				require.EqualValues(t, 1, bindings)
				require.Zero(t, cleanupEvents)
				require.Equal(t, 1, result.Blocked)
				session, err := env.mediaService.repo.FindCustomerServiceMediaUploadSession(ctx, asset.PendingCustomerServiceUploadSessionID)
				require.NoError(t, err)
				require.Equal(t, "bound", session.Status)
			} else {
				require.ErrorIs(t, sendError, repository.ErrCustomerServiceAttachmentBindingUnavailable)
				require.Zero(t, bindings)
				require.EqualValues(t, 1, cleanupEvents)
				require.Equal(t, 1, result.Retired)
			}
		})
	}
}

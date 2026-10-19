package repository

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
)

func createIsolatedPostgreSQLCustomerServiceAttachmentBindingRepository(t *testing.T) *MediaRepository {
	t.Helper()
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	require.NoError(t, repo.db.Exec("CREATE TABLE tickets(id BIGSERIAL PRIMARY KEY); INSERT INTO tickets DEFAULT VALUES").Error)
	require.NoError(t, repo.db.AutoMigrate(&ticket.TicketMessage{}))
	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "420_add_customer_service_message_media_asset_bindings_and_pending_upload_sessions.up.sql"))
	require.NoError(t, err)
	require.NoError(t, repo.db.Exec(string(migration)).Error)
	capacityMigration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "428_add_pending_customer_service_upload_expected_capacity_bytes.up.sql"))
	require.NoError(t, err)
	require.NoError(t, repo.db.Exec(string(capacityMigration)).Error)
	return repo
}

func TestPostgreSQLCustomerServicePrivateUploadReceiptCanBeConsumedByOnlyOneConcurrentMessage(t *testing.T) {
	repo := createIsolatedPostgreSQLCustomerServiceAttachmentBindingRepository(t)
	identity := media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorVisitor, VisitorSessionHash: strings.Repeat("a", 64)}
	key := storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/concurrent.png"
	asset := media.MediaAsset{Filename: "concurrent.png", StorageKey: key, URL: "https://private.example.test/" + key,
		MediaType: "image", MimeType: "image/png", Status: "active", Visibility: "private",
		UploaderActorType: identity.ActorType, UploaderVisitorSessionHash: identity.VisitorSessionHash}
	require.NoError(t, repo.db.Create(&asset).Error)
	session := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
		BusinessContextID: 1, UploadActorType: identity.ActorType, UploadActorVisitorSessionHash: identity.VisitorSessionHash,
		ObjectKey: key, MediaAssetID: &asset.ID, Status: "ready", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, repo.CreateCustomerServicePendingUploadSession(t.Context(), &session))
	start := make(chan struct{})
	results := make(chan error, 8)
	var sends sync.WaitGroup
	for range 8 {
		sends.Go(func() {
			<-start
			results <- repo.db.WithContext(t.Context()).Transaction(func(tx *gorm.DB) error {
				message := ticket.TicketMessage{TicketID: 1, Content: "Photo", MessageType: "image", AttachmentReferenceFormat: ticket.TicketMessageMediaAssetBindings}
				if err := tx.Create(&message).Error; err != nil {
					return err
				}
				return NewTicketRepository(tx).BindCustomerServiceMessageMediaAssetsWithinTransaction(1, message.ID, identity,
					[]ticket.CustomerServiceMessageMediaAttachmentBindingRequest{{UploadSessionID: session.ID}})
			})
		})
	}
	close(start)
	sends.Wait()
	close(results)
	successes, rejections := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrCustomerServiceAttachmentBindingUnavailable) {
			rejections++
		} else {
			t.Fatalf("unexpected concurrent binding error: %v", err)
		}
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 7, rejections)
	for _, model := range []interface{}{&ticket.TicketMessage{}, &ticket.TicketMessageMediaAttachment{}} {
		var count int64
		require.NoError(t, repo.db.Model(model).Count(&count).Error)
		require.EqualValues(t, 1, count)
	}
	storedSession, err := repo.FindCustomerServiceMediaUploadSession(t.Context(), session.ID)
	require.NoError(t, err)
	require.Equal(t, "bound", storedSession.Status)
}

func waitForCustomerServiceAttachmentConcurrentAssetOperationToBlockOnPostgreSQLTransaction(t *testing.T, db *gorm.DB, blockerPID int) {
	t.Helper()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)))", blockerPID).Scan(&waiting).Error
		require.NoError(t, err)
		return waiting
	}, 5*time.Second, 20*time.Millisecond, "concurrent operation did not wait for the locked asset row")
}

func TestPostgreSQLCustomerServiceAssetBindingAndDeletionSerializeUsingTheSameAssetRowLock(t *testing.T) {
	for _, firstOperation := range []string{"bind", "delete"} {
		t.Run(firstOperation+" first", func(t *testing.T) {
			repo := createIsolatedPostgreSQLCustomerServiceAttachmentBindingRepository(t)
			asset := media.MediaAsset{Filename: "public.png", StorageKey: "catalog/public.png", URL: "https://public.example.test/catalog/public.png",
				MediaType: "image", MimeType: "image/png", Status: "active", Visibility: "public"}
			require.NoError(t, repo.db.Create(&asset).Error)
			message := ticket.TicketMessage{TicketID: 1, Content: "Public photo", MessageType: "image", AttachmentReferenceFormat: ticket.TicketMessageMediaAssetBindings}
			require.NoError(t, repo.db.Create(&message).Error)
			identity := media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42}
			bindAsset := func(tx *gorm.DB) error {
				return NewTicketRepository(tx).BindCustomerServiceMessageMediaAssetsWithinTransaction(1, message.ID, identity,
					[]ticket.CustomerServiceMessageMediaAttachmentBindingRequest{{MediaAssetID: asset.ID}})
			}
			assetInUse := errors.New("asset is referenced")
			// Exercise the same repository lock/reference/delete protocol as MediaService.DeleteAsset.
			deleteAsset := func(tx *gorm.DB) error {
				transactionRepository := repo.WithTx(tx)
				current, err := transactionRepository.FindAssetByIDForUpdate(asset.ID)
				if err != nil {
					return err
				}
				references, err := transactionRepository.FindAssetReferences(current)
				if err != nil {
					return err
				}
				if len(references) > 0 {
					return assetInUse
				}
				return transactionRepository.DeleteAsset(asset.ID)
			}
			first := repo.db.WithContext(t.Context()).Begin()
			require.NoError(t, first.Error)
			t.Cleanup(func() { _ = first.Rollback().Error })
			var blockerPID int
			require.NoError(t, first.Raw("SELECT pg_backend_pid()").Scan(&blockerPID).Error)
			secondResult := make(chan error, 1)
			if firstOperation == "bind" {
				require.NoError(t, bindAsset(first))
				go func() { secondResult <- repo.db.WithContext(t.Context()).Transaction(deleteAsset) }()
			} else {
				require.NoError(t, deleteAsset(first))
				go func() { secondResult <- repo.db.WithContext(t.Context()).Transaction(bindAsset) }()
			}
			waitForCustomerServiceAttachmentConcurrentAssetOperationToBlockOnPostgreSQLTransaction(t, repo.db, blockerPID)
			require.NoError(t, first.Commit().Error)
			select {
			case err := <-secondResult:
				if firstOperation == "bind" {
					require.ErrorIs(t, err, assetInUse)
				} else {
					require.ErrorIs(t, err, ErrCustomerServiceAttachmentBindingUnavailable)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("concurrent operation did not finish after asset lock release")
			}
			var count int64
			require.NoError(t, repo.db.Model(&ticket.TicketMessageMediaAttachment{}).Count(&count).Error)
			if firstOperation == "bind" {
				require.EqualValues(t, 1, count)
				current, err := repo.FindAssetByID(asset.ID)
				require.NoError(t, err)
				require.False(t, current.DeletedAt.Valid)
			} else {
				require.Zero(t, count)
				_, err := repo.FindAssetByID(asset.ID)
				require.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}

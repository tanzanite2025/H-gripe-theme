package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

type permanentMediaOriginalProviderWithInjectedReadAndResponseFailures struct {
	*recordingMediaUploadStorage
	responseOverride *string
	readFailure      string
}

func (s *permanentMediaOriginalProviderWithInjectedReadAndResponseFailures) WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	reference, err := s.recordingMediaUploadStorage.WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx, reader, key)
	if s.responseOverride != nil {
		reference = *s.responseOverride
	}
	return reference, err
}

func (s *permanentMediaOriginalProviderWithInjectedReadAndResponseFailures) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	switch s.readFailure {
	case "open-error":
		return nil, errors.New("synthetic-secret-must-not-be-exposed")
	case "nil-object":
		return nil, nil
	case "nil-reader":
		return &storage.StoredObject{}, nil
	case "read-error", "close-error":
		return &storage.StoredObject{ReadCloser: permanentMediaOriginalFailingReadCloser{failRead: s.readFailure == "read-error"}}, nil
	}
	return s.recordingMediaUploadStorage.Open(ctx, key)
}

type permanentMediaOriginalFailingReadCloser struct{ failRead bool }

func (r permanentMediaOriginalFailingReadCloser) Read([]byte) (int, error) {
	if r.failRead {
		return 0, errors.New("injected read failure")
	}
	return 0, io.EOF
}
func (r permanentMediaOriginalFailingReadCloser) Close() error {
	return errors.New("injected close failure")
}

func TestPermanentMediaOriginalUploadIntegrityFailuresKeepExactPendingReceiptWithoutAsset(t *testing.T) {
	for _, failure := range []string{"empty-response", "foreign-host", "missing-bytes", "changed-bytes", "extra-bytes", "open-error", "nil-object", "nil-reader", "read-error", "close-error", "canceled-after-write"} {
		t.Run(failure, func(t *testing.T) {
			service, recording := newMediaAssetUploadTestService(t)
			provider := &permanentMediaOriginalProviderWithInjectedReadAndResponseFailures{recordingMediaUploadStorage: recording, readFailure: failure}
			service.storage = provider
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			recording.onPermanentWrite = func(key string) {
				switch failure {
				case "empty-response":
					value := ""
					provider.responseOverride = &value
				case "foreign-host":
					value := "https://foreign.example.test/uploads/" + key
					provider.responseOverride = &value
				case "missing-bytes":
					delete(recording.objects, key)
				case "changed-bytes":
					recording.objects[key][len(recording.objects[key])-1] ^= 1
				case "extra-bytes":
					recording.objects[key] = append(recording.objects[key], 1)
				case "canceled-after-write":
					cancel()
				}
			}
			asset, err := service.UploadAsset(ctx, MediaUploadInput{File: multipartFileHeader(t, "source.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42})
			require.Error(t, err)
			require.NotContains(t, err.Error(), "synthetic-secret")
			require.Nil(t, asset)
			assets, receipts := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
			require.Zero(t, assets)
			require.EqualValues(t, 1, receipts)
			receipt := permanentMediaOriginalUploadOnlyReceipt(t, service.repo)
			require.Equal(t, outbox.EventStatusPending, receipt.Status)
			require.Zero(t, receipt.Attempts)
			require.Equal(t, 1, recording.uploadCalls)
		})
	}
}

func TestPermanentMediaOriginalUploadAcceptsVideoUsingSameSniffingRulesBeforeAndAfterWrite(t *testing.T) {
	for _, video := range []struct {
		name, mime string
		bytes      []byte
	}{
		{"source.mov", "video/quicktime", []byte{0, 0, 0, 20, 'f', 't', 'y', 'p', 'q', 't', ' ', ' ', 0, 0, 0, 0}},
		{"source.mp4", "video/mp4", []byte{0, 0, 0, 20, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 0, 0}},
		{"source.webm", "video/webm", []byte{0x1a, 0x45, 0xdf, 0xa3, 0, 0, 'w', 'e', 'b', 'm'}},
	} {
		t.Run(video.name, func(t *testing.T) {
			service, _ := newMediaAssetUploadTestService(t)
			asset, err := service.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, video.name, video.mime, video.bytes), MediaType: "video", UploaderID: 42})
			require.NoError(t, err)
			require.Equal(t, video.mime, asset.MimeType)
			assets, receipts := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
			require.EqualValues(t, 1, assets)
			require.Zero(t, receipts)
		})
	}
}

func TestPermanentMediaOriginalUploadReceiptAndAssetPersistenceFailuresAreAtomic(t *testing.T) {
	for _, failure := range []string{"receipt-create", "asset-create", "receipt-delete"} {
		t.Run(failure, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &outbox.Event{})
			provider := &recordingMediaUploadStorage{}
			service := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", 20<<30)
			service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			injected := errors.New("injected atomic exchange failure")
			require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_permanent_original_registration", func(tx *gorm.DB) {
				if (failure == "receipt-create" && tx.Statement.Table == "outbox_events") || (failure == "asset-create" && tx.Statement.Table == "media_assets") {
					tx.AddError(injected)
				}
			}))
			require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register("reject_permanent_original_receipt_exchange", func(tx *gorm.DB) {
				if failure == "receipt-delete" && tx.Statement.Table == "outbox_events" {
					tx.AddError(injected)
				}
			}))
			asset, err := service.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "source.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42})
			require.ErrorIs(t, err, injected)
			require.Nil(t, asset)
			assets, receipts := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
			require.Zero(t, assets)
			if failure == "receipt-create" {
				require.Zero(t, receipts)
				require.Zero(t, provider.uploadCalls)
			} else {
				require.EqualValues(t, 1, receipts)
				require.Equal(t, 1, provider.uploadCalls)
				require.Equal(t, outbox.EventStatusPending, permanentMediaOriginalUploadOnlyReceipt(t, service.repo).Status)
			}
		})
	}
}

func TestPermanentMediaOriginalUploadQuotaRecheckRejectsUsageCommittedAfterPreflightBeforeAnyObjectWrite(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &outbox.Event{})
	provider := &recordingMediaUploadStorage{}
	payload := mediaUploadTestPNG(t)
	service := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", int64(len(payload)))
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	require.NoError(t, db.Callback().Create().After("gorm:create").Register("consume_quota_before_permanent_writer_lock", func(tx *gorm.DB) {
		if tx.Statement.Table != "outbox_events" {
			return
		}
		// This fixture commits the competing usage with the receipt, after preflight
		// and before the writer transaction can recheck quota or call the provider.
		tx.AddError(tx.Session(&gorm.Session{NewDB: true}).Create(&media.MediaAsset{Filename: "competing-original.png", URL: "https://media.example.test/uploads/competing-original.png", Size: int64(len(payload)), UploaderID: 42, UploaderActorType: media.MediaUploadActorLegacyAccount}).Error)
	}))
	asset, err := service.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "source.png", "image/png", payload), MediaType: "image", UploaderID: 42})
	require.ErrorIs(t, err, ErrMediaAccountStorageQuotaExceeded)
	require.Nil(t, asset)
	require.Zero(t, provider.uploadCalls)
	assets, receipts := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
	require.EqualValues(t, 1, assets)
	require.EqualValues(t, 1, receipts)
}

func TestPermanentMediaOriginalFailedUploadHistoricalReferenceRetainsOriginalTaskUntilLastUnbind(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	service := newMediaDeleteTestService(t, db, t.TempDir())
	require.NoError(t, db.Exec("CREATE TABLE warranty_claims (id INTEGER PRIMARY KEY, images TEXT, video_url TEXT, deleted_at TIMESTAMP)").Error)
	provider := &permanentMediaOriginalLocalProviderWithLostResponse{StorageService: service.storage}
	service.storage = provider
	_, err := service.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "source.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42})
	require.Error(t, err)
	receipt := permanentMediaOriginalUploadOnlyReceipt(t, service.repo)
	require.NoError(t, db.Exec("INSERT INTO warranty_claims (id, images, deleted_at) VALUES (1, ?, ?)", encodeLegacyMediaReferenceRetentionTestJSON(t, []string{provider.GetURL(provider.key)}), time.Now()).Error)
	handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), receipt), ErrMediaAssetInUse)
	object, err := provider.Open(t.Context(), provider.key)
	require.NoError(t, err)
	require.NoError(t, object.ReadCloser.Close())
	retained := permanentMediaOriginalUploadOnlyReceipt(t, service.repo)
	require.Equal(t, receipt.ID, retained.ID)
	require.Equal(t, receipt.EventKey, retained.EventKey)
	require.NoError(t, db.Exec("UPDATE warranty_claims SET images = '[]' WHERE id = 1").Error)
	require.NoError(t, handler.Handle(t.Context(), retained))
	_, err = provider.Open(t.Context(), provider.key)
	require.Error(t, err)
	require.Equal(t, receipt.ID, permanentMediaOriginalUploadOnlyReceipt(t, service.repo).ID)
}

type permanentMediaOriginalLocalProviderWithLostResponse struct {
	storage.StorageService
	key string
}

func (s *permanentMediaOriginalLocalProviderWithLostResponse) WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.key = key
	_, err := s.StorageService.(storage.PreallocatedPermanentMediaOriginalObjectWriter).WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx, reader, key)
	if err != nil {
		return "", err
	}
	return "", errors.New("lost provider response")
}
func (s *permanentMediaOriginalLocalProviderWithLostResponse) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func TestPermanentMediaOriginalUploadRejectsNestedTransactionBeforeReceiptOrProviderIO(t *testing.T) {
	service, provider := newMediaAssetUploadTestService(t)
	require.NoError(t, service.repo.WithTransaction(func(repo *repository.MediaRepository, _ *gorm.DB) error {
		transactionService := *service
		transactionService.repo = repo
		_, err := transactionService.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "source.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42})
		require.ErrorContains(t, err, "own transaction lifetime")
		return nil
	}))
	require.Zero(t, provider.uploadCalls)
	assets, receipts := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
	require.Zero(t, assets)
	require.Zero(t, receipts)
}

func TestPermanentMediaOriginalUploadChangedExpiredOrPreviouslyAttemptedReceiptPreventsObjectWrite(t *testing.T) {
	for _, change := range []string{"payload", "event-type", "expired", "attempts", "previous-attempt", "missing"} {
		t.Run(change, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &outbox.Event{})
			provider := &recordingMediaUploadStorage{}
			service := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", 20<<30)
			service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			require.NoError(t, db.Callback().Create().After("gorm:create").Register("change_committed_permanent_original_receipt", func(tx *gorm.DB) {
				event, ok := tx.Statement.Dest.(*outbox.Event)
				if !ok || !strings.HasPrefix(event.EventKey, repository.PermanentMediaOriginalUploadReceiptEventKeyPrefix) {
					return
				}
				query := tx.Session(&gorm.Session{NewDB: true}).Model(&outbox.Event{}).Where("id = ?", event.ID)
				var err error
				switch change {
				case "payload":
					err = query.Update("payload", []byte(`{"resource_type":"changed"}`)).Error
				case "event-type":
					err = query.Update("event_type", "changed").Error
				case "expired":
					err = query.Update("available_at", time.Now().Add(-time.Minute)).Error
				case "attempts":
					err = query.Update("attempts", 1).Error
				case "previous-attempt":
					err = query.Update("last_attempt_at", time.Now()).Error
				case "missing":
					err = query.Unscoped().Delete(&outbox.Event{}).Error
				}
				tx.AddError(err)
			}))
			_, err := service.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "source.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42})
			require.ErrorIs(t, err, repository.ErrPermanentMediaOriginalUploadReceiptNoLongerWritable)
			require.Zero(t, provider.uploadCalls)
			assets, _ := permanentMediaOriginalUploadAssetAndReceiptCounts(t, service.repo)
			require.Zero(t, assets)
		})
	}
}

func TestPermanentMediaOriginalSuccessfulExchangePreservesExistingCleanupTask(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &outbox.Event{})
	provider := &recordingMediaUploadStorage{}
	service := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", 20<<30)
	service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	existing, err := newObjectStorageCleanupEvent(objectCleanupResourceUnregisteredMediaUpload, "historical", "media_asset", "historical", []string{"catalog/historical.png"})
	require.NoError(t, err)
	require.NoError(t, db.Create(existing).Error)
	var before outbox.Event
	require.NoError(t, db.First(&before, existing.ID).Error)
	_, err = service.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "source.png", "image/png", mediaUploadTestPNG(t)), MediaType: "image", UploaderID: 42})
	require.NoError(t, err)
	var remaining []outbox.Event
	require.NoError(t, db.Find(&remaining).Error)
	require.Equal(t, []outbox.Event{before}, remaining)
}

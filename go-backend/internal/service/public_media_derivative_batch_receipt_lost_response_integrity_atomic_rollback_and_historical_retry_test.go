package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

type publicMediaDerivativeReceiptProviderWithInjectedFailures struct {
	storage.StorageService
	mode        string
	keys        []string
	beforeWrite func(string)
	afterWrite  func(string)
}

func (s *publicMediaDerivativeReceiptProviderWithInjectedFailures) WritePublicMediaDerivativeObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if s.beforeWrite != nil {
		s.beforeWrite(key)
	}
	s.keys = append(s.keys, key)
	if s.mode == "fail-before-write" || s.mode == "second-write" && len(s.keys) == 2 {
		return "", errors.New("injected write failure")
	}
	reference, err := s.StorageService.(storage.PreallocatedPublicMediaDerivativeObjectWriter).WritePublicMediaDerivativeObjectAtPreallocatedKey(ctx, reader, key)
	if err != nil {
		return reference, err
	}
	if s.afterWrite != nil {
		s.afterWrite(key)
	}
	switch s.mode {
	case "lost-response":
		return "", errors.New("response lost after bytes were accepted")
	case "empty-response":
		return "", nil
	case "wrong-key":
		return s.GetURL("media-derivatives/42/card/foreign.jpg"), nil
	case "foreign-host":
		return "https://foreign.example.test/uploads/" + key, nil
	}
	return reference, nil
}
func (s *publicMediaDerivativeReceiptProviderWithInjectedFailures) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	if storage.IsPreallocatedPublicMediaDerivativeObjectKey(key) {
		switch s.mode {
		case "open-error":
			return nil, errors.New("synthetic-private-secret")
		case "nil-object":
			return nil, nil
		case "nil-reader":
			return &storage.StoredObject{}, nil
		case "read-error":
			return &storage.StoredObject{ReadCloser: permanentMediaOriginalFailingReadCloser{failRead: true}}, nil
		case "close-error":
			return &storage.StoredObject{ReadCloser: permanentMediaOriginalFailingReadCloser{}}, nil
		}
	}
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func createPublicMediaDerivativeReceiptTestEnvironment(t *testing.T) (*gorm.DB, *MediaService, *publicMediaDerivativeReceiptProviderWithInjectedFailures, *media.MediaAsset, string, []byte) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "public-derivative-receipt-state.sqlite")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.Exec("PRAGMA journal_mode=WAL").Error)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}))
	root := t.TempDir()
	service := newMediaDeleteTestService(t, db, root)
	payload := testOpaquePNG(t, 80, 40)
	parent, err := service.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "parent.png", "image/png", payload), MediaType: "image", UploaderID: 4242})
	require.NoError(t, err)
	provider := &publicMediaDerivativeReceiptProviderWithInjectedFailures{StorageService: service.storage}
	service.storage = provider
	return db, service, provider, parent, root, payload
}
func generatePublicMediaDerivativeReceiptTestBatch(ctx context.Context, service *MediaService, parent *media.MediaAsset, payload []byte) ([]media.MediaAssetDerivative, error) {
	return service.generateAssetDerivativesFromSource(ctx, parent, func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload)), nil }, MediaDerivativePresetDefinitions())
}
func findPublicMediaDerivativeReceiptsForTest(t *testing.T, db *gorm.DB) []outbox.Event {
	t.Helper()
	var events []outbox.Event
	require.NoError(t, db.Where("event_key LIKE ?", repository.PublicMediaDerivativeUploadReceiptEventKeyPrefix+"%").Order("id ASC").Find(&events).Error)
	return events
}
func countPublicMediaDerivativeRowsForTest(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&media.MediaAssetDerivative{}).Count(&count).Error)
	return count
}

func TestPublicMediaDerivativeBatchCommitsEveryReceiptBeforeFirstWriteAndAtomicallyRegistersAllPresets(t *testing.T) {
	db, service, provider, parent, _, payload := createPublicMediaDerivativeReceiptTestEnvironment(t)
	provider.beforeWrite = func(key string) {
		receipts := findPublicMediaDerivativeReceiptsForTest(t, db)
		require.Len(t, receipts, 3)
		found := false
		for _, receipt := range receipts {
			require.Equal(t, outbox.EventStatusPending, receipt.Status)
			require.Zero(t, receipt.Attempts)
			require.True(t, receipt.AvailableAt.After(time.Now()))
			if strings.Contains(string(receipt.Payload), key) {
				found = true
			}
		}
		require.True(t, found)
		require.Zero(t, countPublicMediaDerivativeRowsForTest(t, db))
	}
	generated, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
	require.NoError(t, err)
	require.Len(t, generated, 3)
	require.EqualValues(t, 3, countPublicMediaDerivativeRowsForTest(t, db))
	require.Empty(t, findPublicMediaDerivativeReceiptsForTest(t, db))
	for _, derivative := range generated {
		require.NotZero(t, derivative.ID)
		owner, valid := storage.PreallocatedPublicMediaDerivativeParentAssetID(derivative.StorageKey)
		require.True(t, valid)
		require.Equal(t, parent.ID, owner)
	}
}

func TestPublicMediaDerivativeBatchWriteAndIntegrityFailuresRetainAllOriginalReceiptsWithoutRegistrations(t *testing.T) {
	for _, mode := range []string{"fail-before-write", "second-write", "lost-response", "empty-response", "wrong-key", "foreign-host", "missing-bytes", "changed-bytes", "extra-bytes", "open-error", "nil-object", "nil-reader", "read-error", "close-error", "canceled-after-write"} {
		t.Run(mode, func(t *testing.T) {
			db, service, provider, parent, root, payload := createPublicMediaDerivativeReceiptTestEnvironment(t)
			provider.mode = mode
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			provider.afterWrite = func(key string) {
				filename := filepath.Join(root, filepath.FromSlash(key))
				switch mode {
				case "missing-bytes":
					require.NoError(t, os.Remove(filename))
				case "changed-bytes":
					data, err := os.ReadFile(filename)
					require.NoError(t, err)
					data[len(data)-1] ^= 1
					require.NoError(t, os.WriteFile(filename, data, 0600))
				case "extra-bytes":
					data, err := os.ReadFile(filename)
					require.NoError(t, err)
					require.NoError(t, os.WriteFile(filename, append(data, 1), 0600))
				case "canceled-after-write":
					cancel()
				}
			}
			generated, err := generatePublicMediaDerivativeReceiptTestBatch(ctx, service, parent, payload)
			require.Error(t, err)
			require.NotContains(t, err.Error(), "synthetic-private-secret")
			require.Empty(t, generated)
			require.Zero(t, countPublicMediaDerivativeRowsForTest(t, db))
			receipts := findPublicMediaDerivativeReceiptsForTest(t, db)
			require.Len(t, receipts, 3)
			for _, receipt := range receipts {
				require.Equal(t, outbox.EventStatusPending, receipt.Status)
				require.Zero(t, receipt.Attempts)
			}
			require.FileExists(t, filepath.Join(root, filepath.FromSlash(parent.StorageKey)))
			handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
			for _, receipt := range receipts {
				require.NoError(t, handler.Handle(t.Context(), receipt))
			}
			for _, key := range provider.keys {
				require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(key)))
			}
			require.Equal(t, receipts, findPublicMediaDerivativeReceiptsForTest(t, db), "direct cleanup keeps the same durable tasks for worker acknowledgment")
		})
	}
}

func TestPublicMediaDerivativeBatchPersistenceFailuresRollBackAllMetadataAndKeepCommittedReceipts(t *testing.T) {
	for _, failure := range []string{"second-receipt-create", "derivative-create", "second-receipt-delete"} {
		t.Run(failure, func(t *testing.T) {
			db, service, provider, parent, _, payload := createPublicMediaDerivativeReceiptTestEnvironment(t)
			creates, deletes := 0, 0
			injected := errors.New("injected derivative batch persistence failure")
			require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_public_derivative_batch_create", func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" {
					creates++
					if failure == "second-receipt-create" && creates == 2 {
						tx.AddError(injected)
					}
				}
				if failure == "derivative-create" && tx.Statement.Table == "media_asset_derivatives" {
					tx.AddError(injected)
				}
			}))
			require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register("reject_public_derivative_batch_exchange", func(tx *gorm.DB) {
				if tx.Statement.Table == "outbox_events" {
					deletes++
					if failure == "second-receipt-delete" && deletes == 2 {
						tx.AddError(injected)
					}
				}
			}))
			generated, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
			require.ErrorIs(t, err, injected)
			require.Empty(t, generated)
			require.Zero(t, countPublicMediaDerivativeRowsForTest(t, db))
			if failure == "second-receipt-create" {
				require.Empty(t, provider.keys)
				require.Empty(t, findPublicMediaDerivativeReceiptsForTest(t, db))
			} else {
				require.Len(t, provider.keys, 3)
				require.Len(t, findPublicMediaDerivativeReceiptsForTest(t, db), 3)
			}
		})
	}
}

func TestPublicMediaDerivativeFailedBatchHistoricalReferenceKeepsExactOriginalTaskUntilLastUnbind(t *testing.T) {
	db, service, provider, parent, root, payload := createPublicMediaDerivativeReceiptTestEnvironment(t)
	provider.mode = "lost-response"
	_, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
	require.Error(t, err)
	require.Len(t, provider.keys, 1)
	require.NoError(t, db.Exec("CREATE TABLE warranty_claims (id INTEGER PRIMARY KEY, images TEXT, video_url TEXT, deleted_at TIMESTAMP)").Error)
	require.NoError(t, db.Exec("INSERT INTO warranty_claims (id,images,deleted_at) VALUES (1,?,?)", encodeLegacyMediaReferenceRetentionTestJSON(t, []string{provider.GetURL(provider.keys[0])}), time.Now()).Error)
	var retained outbox.Event
	for _, receipt := range findPublicMediaDerivativeReceiptsForTest(t, db) {
		var document outbox.ObjectStorageCleanupPayload
		require.NoError(t, json.Unmarshal(receipt.Payload, &document))
		if document.ObjectKeys[0] == provider.keys[0] {
			retained = receipt
		}
	}
	require.NotZero(t, retained.ID)
	handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
	require.ErrorIs(t, handler.Handle(t.Context(), retained), ErrMediaAssetInUse)
	require.FileExists(t, filepath.Join(root, filepath.FromSlash(provider.keys[0])))
	require.NoError(t, db.Exec("UPDATE warranty_claims SET images = '[]' WHERE id = 1").Error)
	require.NoError(t, handler.Handle(t.Context(), retained))
	require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(provider.keys[0])))
	var same outbox.Event
	require.NoError(t, db.First(&same, retained.ID).Error)
	require.Equal(t, retained, same)
}

type publicMediaDerivativeProviderWithoutSourceReader struct {
	storage.StorageService
	storage.PreallocatedPublicMediaDerivativeObjectWriter
}

func TestPublicMediaDerivativeBatchRejectsMissingProviderCapabilitiesWithoutFallbackOrReceipts(t *testing.T) {
	for _, missing := range []string{"exact writer", "source reader"} {
		t.Run(missing, func(t *testing.T) {
			db, service, provider, parent, _, payload := createPublicMediaDerivativeReceiptTestEnvironment(t)
			if missing == "exact writer" {
				service.storage = struct{ storage.StorageService }{provider}
			} else {
				service.storage = publicMediaDerivativeProviderWithoutSourceReader{StorageService: provider, PreallocatedPublicMediaDerivativeObjectWriter: provider}
			}
			_, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
			require.ErrorIs(t, err, ErrMediaStorageUnavailable)
			require.Empty(t, provider.keys)
			require.Empty(t, findPublicMediaDerivativeReceiptsForTest(t, db))
		})
	}
}

func TestPublicMediaDerivativeBatchChangedExpiredOrPreviouslyAttemptedReceiptPreventsEveryWrite(t *testing.T) {
	for _, change := range []string{"payload", "event-type", "expired", "attempts", "previous-attempt", "missing", "parent-retired"} {
		t.Run(change, func(t *testing.T) {
			db, service, provider, parent, _, payload := createPublicMediaDerivativeReceiptTestEnvironment(t)
			require.NoError(t, db.Callback().Create().After("gorm:create").Register("alter_public_derivative_receipt_before_writer_lock", func(tx *gorm.DB) {
				event, ok := tx.Statement.Dest.(*outbox.Event)
				if !ok || !strings.HasPrefix(event.EventKey, repository.PublicMediaDerivativeUploadReceiptEventKeyPrefix) {
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
				case "parent-retired":
					err = tx.Session(&gorm.Session{NewDB: true}).Model(&media.MediaAsset{}).Where("id = ?", parent.ID).Update("status", "deleting").Error
				}
				tx.AddError(err)
			}))
			_, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
			require.ErrorIs(t, err, repository.ErrPublicMediaDerivativeUploadReceiptNoLongerWritable)
			require.Empty(t, provider.keys)
			require.Zero(t, countPublicMediaDerivativeRowsForTest(t, db))
		})
	}
}

func TestPublicMediaDerivativeBatchRejectsNestedTransactionBeforeReceiptsOrProviderIO(t *testing.T) {
	db, service, provider, parent, _, payload := createPublicMediaDerivativeReceiptTestEnvironment(t)
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		transactionService := *service
		transactionService.repo = service.repo.WithTx(tx)
		_, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), &transactionService, parent, payload)
		require.ErrorContains(t, err, "own transaction lifetime")
		return nil
	}))
	require.Empty(t, provider.keys)
	require.Empty(t, findPublicMediaDerivativeReceiptsForTest(t, db))
}

func TestPublicMediaDerivativeSuccessfulRetryConsumesOnlyNewReceiptsAndPreservesOldFailedTask(t *testing.T) {
	db, service, provider, parent, _, payload := createPublicMediaDerivativeReceiptTestEnvironment(t)
	provider.mode = "lost-response"
	_, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
	require.Error(t, err)
	previous := findPublicMediaDerivativeReceiptsForTest(t, db)
	require.Len(t, previous, 3)
	provider.mode = ""
	generated, err := generatePublicMediaDerivativeReceiptTestBatch(t.Context(), service, parent, payload)
	require.NoError(t, err)
	require.Len(t, generated, 3)
	require.Equal(t, previous, findPublicMediaDerivativeReceiptsForTest(t, db))
	handler := NewObjectStorageCleanupOutboxHandler(service, nil, nil, nil)
	for _, receipt := range previous {
		require.NoError(t, handler.Handle(t.Context(), receipt))
	}
	for _, derivative := range generated {
		object, err := provider.Open(t.Context(), derivative.StorageKey)
		require.NoError(t, err)
		require.NoError(t, object.ReadCloser.Close())
	}
}

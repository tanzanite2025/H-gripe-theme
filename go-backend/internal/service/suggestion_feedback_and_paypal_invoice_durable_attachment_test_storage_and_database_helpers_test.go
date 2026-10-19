package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func configurePayPalInvoiceDurableStorageForRegression(t *testing.T, db *gorm.DB, svc *PaymentService, provider *fakePayPalDisputeDocumentStorage) {
	t.Helper()
	var err error
	provider.StorageService, err = storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://cdn.example.test"})
	require.NoError(t, err)
	svc.ConfigurePayPalInvoiceDurableAttachmentUploads(NewSuggestionAndPayPalAttachmentUploadService(repository.NewOutboxRepository(db), provider, "https://shop.example.test"))
	svc.ConfigurePayPalDisputeEvidenceDocumentStorage(provider)
}

func (s *fakePayPalDisputeDocumentStorage) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *fakePayPalDisputeDocumentStorage) WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	s.data, s.filename, s.uploads = data, key, s.uploads+1
	s.url, err = s.StorageService.(storage.PreallocatedSuggestionOrPayPalAttachmentWriter).WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(ctx, bytes.NewReader(data), key)
	return s.url, err
}

type suggestionAndPayPalAttachmentControlledStorage struct {
	storage.StorageService
	writtenKeys       []string
	replacementBytes  []byte
	returnedReference *string
	writeError        error
	afterWrite        func()
}

func (s *suggestionAndPayPalAttachmentControlledStorage) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *suggestionAndPayPalAttachmentControlledStorage) WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	s.writtenKeys = append(s.writtenKeys, key)
	if s.replacementBytes != nil {
		reader = bytes.NewReader(s.replacementBytes)
	}
	url, err := s.StorageService.(storage.PreallocatedSuggestionOrPayPalAttachmentWriter).WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(ctx, reader, key)
	if err != nil {
		return url, err
	}
	if s.afterWrite != nil {
		s.afterWrite()
	}
	if s.returnedReference != nil {
		url = *s.returnedReference
	}
	return url, s.writeError
}

func createSuggestionAndPayPalAttachmentRegressionEnvironment(t *testing.T) (*gorm.DB, *SuggestionFeedbackService, *SuggestionAndPayPalAttachmentUploadService, *suggestionAndPayPalAttachmentControlledStorage) {
	t.Helper()
	db, _ := newTestPaymentService(t)
	require.NoError(t, db.AutoMigrate(&suggestionfeedback.SuggestionFeedback{}, &loyalty.MemberLevel{}))
	level := loyalty.MemberLevel{Name: "silver"}
	require.NoError(t, db.Create(&level).Error)
	require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: 71, MemberLevelID: level.ID}).Error)
	require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: 72, MemberLevelID: level.ID}).Error)
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://cdn.example.test"})
	require.NoError(t, err)
	controlled := &suggestionAndPayPalAttachmentControlledStorage{StorageService: provider}
	uploads := NewSuggestionAndPayPalAttachmentUploadService(repository.NewOutboxRepository(db), controlled, "https://shop.example.test")
	svc := NewSuggestionFeedbackService(repository.NewSuggestionFeedbackRepository(db))
	svc.ConfigureSuggestionAttachmentDurableUploads(uploads)
	return db, svc, uploads, controlled
}

func findSuggestionOrPayPalAttachmentReceiptForRegression(t *testing.T, db *gorm.DB, key string) (outbox.Event, outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload) {
	t.Helper()
	_, _, id, valid := storage.SuggestionOrPayPalAttachmentObjectIdentity(key)
	require.True(t, valid)
	var event outbox.Event
	require.NoError(t, db.Where("event_key = ?", outbox.SuggestionAndPayPalAttachmentReceiptEventKeyPrefix+id).First(&event).Error)
	var payload outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	return event, payload
}

func processSuggestionAndPayPalAttachmentCleanupForRegression(t *testing.T, db *gorm.DB, uploads *SuggestionAndPayPalAttachmentUploadService, event outbox.Event) OutboxDispatchResult {
	t.Helper()
	handler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
	handler.ConfigureSuggestionAndPayPalAttachmentUploads(uploads)
	worker := NewOutboxService(repository.NewOutboxRepository(db))
	worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, handler.Handle)
	result, err := worker.ProcessPending(t.Context(), event.AvailableAt.Add(1), 100)
	require.NoError(t, err)
	return result
}

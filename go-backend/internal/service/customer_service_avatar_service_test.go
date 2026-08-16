package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCustomerServiceAvatarUploadReplacesProfileReferenceAndQueuesCleanup(t *testing.T) {
	db := newCustomerServiceAvatarTestDB(t)
	profile := createCustomerServiceAvatarProfile(t, db, "https://assets.example.test/customer-service/avatars/2026/08/16/old.webp")
	store := &recordingCustomerServiceAvatarStorage{
		uploadURL: "https://assets.example.test/customer-service/avatars/2026/08/16/new.webp",
	}
	service := NewCustomerServiceAvatarService(
		repository.NewUserRepository(db),
		store,
		repository.NewOutboxRepository(db),
	)

	avatarURL, err := service.Upload(context.Background(), *profile.UserID, customerServiceAvatarFile(t))
	require.NoError(t, err)
	require.Equal(t, store.uploadURL, avatarURL)
	require.Equal(t, CustomerServiceAvatarCacheControl, store.cacheControl)
	require.Equal(t, []string{profile.Avatar}, store.deletedURLs)

	var saved user.AgentProfile
	require.NoError(t, db.First(&saved, profile.ID).Error)
	require.Equal(t, store.uploadURL, saved.Avatar)

	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeCustomerServiceAvatarCleanup).First(&event).Error)
	require.Equal(t, outbox.AggregateTypeCustomerServiceAgentProfile, event.AggregateType)
	require.Equal(t, fmt.Sprint(profile.ID), event.AggregateID)
	require.Equal(t, customerServiceAvatarCleanupMaxAttempts, event.MaxAttempts)

	var payload outbox.CustomerServiceAvatarCleanupPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	require.Equal(t, profile.Avatar, payload.URL)
}

func TestCustomerServiceAvatarUploadDoesNotDeleteHistoricalExternalURL(t *testing.T) {
	db := newCustomerServiceAvatarTestDB(t)
	profile := createCustomerServiceAvatarProfile(t, db, "https://example.invalid/legacy-avatar.png")
	store := &recordingCustomerServiceAvatarStorage{
		uploadURL: "https://assets.example.test/customer-service/avatars/2026/08/16/new.webp",
	}
	service := NewCustomerServiceAvatarService(
		repository.NewUserRepository(db),
		store,
		repository.NewOutboxRepository(db),
	)

	_, err := service.Upload(context.Background(), *profile.UserID, customerServiceAvatarFile(t))
	require.NoError(t, err)
	require.Empty(t, store.deletedURLs)

	var cleanupCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeCustomerServiceAvatarCleanup).Count(&cleanupCount).Error)
	require.Zero(t, cleanupCount)
}

func TestCustomerServiceAvatarRemoveClearsReferenceAndQueuesCleanup(t *testing.T) {
	db := newCustomerServiceAvatarTestDB(t)
	profile := createCustomerServiceAvatarProfile(t, db, "https://assets.example.test/customer-service/avatars/2026/08/16/old.webp")
	store := &recordingCustomerServiceAvatarStorage{}
	service := NewCustomerServiceAvatarService(
		repository.NewUserRepository(db),
		store,
		repository.NewOutboxRepository(db),
	)

	require.NoError(t, service.Remove(context.Background(), *profile.UserID))
	require.Equal(t, []string{profile.Avatar}, store.deletedURLs)

	var saved user.AgentProfile
	require.NoError(t, db.First(&saved, profile.ID).Error)
	require.Empty(t, saved.Avatar)

	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeCustomerServiceAvatarCleanup).First(&event).Error)
	var payload outbox.CustomerServiceAvatarCleanupPayload
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	require.Equal(t, profile.Avatar, payload.URL)
}

func TestCustomerServiceAvatarUploadRejectsMissingProfileBeforeStorageWrite(t *testing.T) {
	db := newCustomerServiceAvatarTestDB(t)
	store := &recordingCustomerServiceAvatarStorage{
		uploadURL: "https://assets.example.test/customer-service/avatars/2026/08/16/new.webp",
	}
	service := NewCustomerServiceAvatarService(
		repository.NewUserRepository(db),
		store,
		repository.NewOutboxRepository(db),
	)

	_, err := service.Upload(context.Background(), 999, customerServiceAvatarFile(t))
	require.ErrorIs(t, err, ErrCustomerServiceAvatarProfileNotFound)
	require.Zero(t, store.uploadCalls)
}

func TestCustomerServiceAvatarCleanupHandlerOnlyDeletesUnreferencedManagedAvatar(t *testing.T) {
	db := newCustomerServiceAvatarTestDB(t)
	profile := createCustomerServiceAvatarProfile(t, db, "https://assets.example.test/customer-service/avatars/2026/08/16/current.webp")
	store := &recordingCustomerServiceAvatarStorage{}
	handler := NewCustomerServiceAvatarCleanupHandler(repository.NewUserRepository(db), store)

	payload, err := json.Marshal(outbox.CustomerServiceAvatarCleanupPayload{URL: profile.Avatar})
	require.NoError(t, err)
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeCustomerServiceAvatarCleanup,
		Payload:   payload,
	}))
	require.Empty(t, store.deletedURLs)

	staleURL := "https://assets.example.test/customer-service/avatars/2026/08/16/stale.webp"
	payload, err = json.Marshal(outbox.CustomerServiceAvatarCleanupPayload{URL: staleURL})
	require.NoError(t, err)
	require.NoError(t, handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeCustomerServiceAvatarCleanup,
		Payload:   payload,
	}))
	require.Equal(t, []string{staleURL}, store.deletedURLs)
}

func TestCustomerServiceAvatarCleanupHandlerReturnsStorageFailureForRetry(t *testing.T) {
	db := newCustomerServiceAvatarTestDB(t)
	store := &recordingCustomerServiceAvatarStorage{deleteErr: errors.New("object store unavailable")}
	handler := NewCustomerServiceAvatarCleanupHandler(repository.NewUserRepository(db), store)
	staleURL := "https://assets.example.test/customer-service/avatars/2026/08/16/stale.webp"
	payload, err := json.Marshal(outbox.CustomerServiceAvatarCleanupPayload{URL: staleURL})
	require.NoError(t, err)

	err = handler.Handle(context.Background(), outbox.Event{
		EventType: outbox.EventTypeCustomerServiceAvatarCleanup,
		Payload:   payload,
	})
	require.ErrorContains(t, err, "object store unavailable")
}

func TestCustomerServiceAvatarPublicAccessRequiresCurrentReference(t *testing.T) {
	db := newCustomerServiceAvatarTestDB(t)
	profile := createCustomerServiceAvatarProfile(t, db, "https://assets.example.test/customer-service/avatars/2026/08/16/current.webp")
	store := &recordingCustomerServiceAvatarStorage{}
	service := NewCustomerServiceAvatarService(repository.NewUserRepository(db), store, repository.NewOutboxRepository(db))

	allowed, err := service.CanServePublicAvatar(context.Background(), "customer-service/avatars/2026/08/16/current.webp")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = service.CanServePublicAvatar(context.Background(), "customer-service/avatars/2026/08/16/orphan.webp")
	require.NoError(t, err)
	require.False(t, allowed)

	require.NotNil(t, profile)
}

func newCustomerServiceAvatarTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&user.User{}, &user.AgentProfile{}, &outbox.Event{}))
	return db
}

func createCustomerServiceAvatarProfile(t *testing.T, db *gorm.DB, avatarURL string) *user.AgentProfile {
	t.Helper()
	agentUser := user.User{
		Email:    fmt.Sprintf("agent-%d@example.test", len(avatarURL)),
		Username: fmt.Sprintf("agent-%d", len(avatarURL)),
		Password: "test-password",
		Role:     "support",
		Status:   "active",
	}
	require.NoError(t, db.Create(&agentUser).Error)
	profile := &user.AgentProfile{
		AgentID: "agent-" + strings.ReplaceAll(fmt.Sprint(agentUser.ID), " ", ""),
		UserID:  &agentUser.ID,
		Name:    "Support",
		Avatar:  avatarURL,
		Status:  "active",
	}
	require.NoError(t, db.Create(profile).Error)
	return profile
}

func customerServiceAvatarFile(t *testing.T) *multipart.FileHeader {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString("UklGRrIBAABXRUJQVlA4TKUBAAAvSsAYAA8w//M///MfeJAkbXvaSG7m8Q3GfYSBJekwQztm/IcZlgwnmWImn2BK7aFmBtnVir6q//8VOkFE/xm4baTIu8c48ArEo6+B3zFKYln3pqClSCKX0begFTAXFOLXHSyF8cCNcZEG4OywuA4KVVfJCiArU7GAgJI8+lJP/OKMT/fBAjevg1cYB7YVkFuWga2lyPi5I0HFy5YTpWIHg0RZpkniRVW9odHAKOwosWuOGdxIyn2OvaCDvhg/we6TwadPBPbqBV58MsLmMJ8yZnOWk8SRz4N+QoyPL+MnamzMvcE1rHNEr91F9GKZPVUcS9w7PhhH36suB9qPeYb/oLk6cuTiJ0wOK3m5h1cKjW6EVZCYMK7dxcKCBdgP9HkKr9gkAO2P8GKZGWVdIAatQa+1IDpt6qyorVwdy01xdW8Jkfk6xjEXmVQQ+HQdFr6OKhIN34dXWq0+0qr6EJSCeeVLH9+gvGTLyqM65PQ44ihzlTXxQKjKbAvshXgir7Lil9w4L2bvMycmjQcqXaMCO6BlY28i+FOLzbfI1vEqxAhotocAAA==")
	require.NoError(t, err)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "avatar.webp")
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, request.ParseMultipartForm(int64(body.Len()+1024)))
	t.Cleanup(func() { _ = request.MultipartForm.RemoveAll() })
	return request.MultipartForm.File["file"][0]
}

type recordingCustomerServiceAvatarStorage struct {
	uploadURL    string
	uploadCalls  int
	cacheControl string
	deletedURLs  []string
	deleteErr    error
}

func (s *recordingCustomerServiceAvatarStorage) Upload(context.Context, *multipart.FileHeader) (string, error) {
	return s.UploadWithPrefix(context.Background(), nil, "")
}

func (s *recordingCustomerServiceAvatarStorage) UploadWithPrefix(_ context.Context, _ *multipart.FileHeader, prefix string) (string, error) {
	s.uploadCalls++
	if !strings.HasPrefix(prefix, CustomerServiceAvatarStoragePrefix) {
		return "", errors.New("unexpected storage prefix")
	}
	return s.uploadURL, nil
}

func (s *recordingCustomerServiceAvatarStorage) UploadWithPrefixAndCacheControl(
	ctx context.Context,
	file *multipart.FileHeader,
	prefix string,
	cacheControl string,
) (string, error) {
	s.cacheControl = cacheControl
	return s.UploadWithPrefix(ctx, file, prefix)
}

func (s *recordingCustomerServiceAvatarStorage) UploadFromReader(context.Context, io.Reader, string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *recordingCustomerServiceAvatarStorage) UploadFromReaderWithPrefix(context.Context, io.Reader, string, string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *recordingCustomerServiceAvatarStorage) Delete(_ context.Context, url string) error {
	s.deletedURLs = append(s.deletedURLs, url)
	return s.deleteErr
}

func (s *recordingCustomerServiceAvatarStorage) GetURL(filename string) string {
	return "https://assets.example.test/" + strings.TrimLeft(filename, "/")
}

func (s *recordingCustomerServiceAvatarStorage) ObjectKey(reference string) (string, error) {
	prefix := "https://assets.example.test/"
	if !strings.HasPrefix(reference, prefix) {
		return "", errors.New("untrusted object URL")
	}
	key := strings.TrimPrefix(reference, prefix)
	if normalized, ok := storage.NormalizeObjectKey(key); ok {
		return normalized, nil
	}
	return "", errors.New("invalid object key")
}

func (s *recordingCustomerServiceAvatarStorage) CopyObject(context.Context, string, string) error {
	return errors.New("not implemented")
}

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	suggestionapi "commerce-platform/internal/api/v1/suggestionfeedback"
	"commerce-platform/internal/app"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSuggestionAndPayPalAttachmentLocalRouteServesExactOwnedBytesAndPreservesNoStoreAcrossUploadAndBinding(t *testing.T) {
	root := t.TempDir()
	t.Setenv("STORAGE_TYPE", "local")
	t.Setenv("STORAGE_LOCAL_PATH", root)
	t.Setenv("STORAGE_BASE_URL", "https://cdn.example.test")
	t.Setenv("SITE_LOGO_STORAGE_TYPE", "local")
	t.Setenv("SITE_LOGO_STORAGE_LOCAL_PATH", t.TempDir())
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://cdn.example.test"})
	require.NoError(t, err)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&outbox.Event{}, &suggestionfeedback.SuggestionFeedback{}, &loyalty.MemberLevel{}, &loyalty.UserLoyalty{}, &orderevidence.OrderEvidenceSubmissionSnapshot{}))
	level := loyalty.MemberLevel{Name: "silver"}
	require.NoError(t, db.Create(&level).Error)
	require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: 71, MemberLevelID: level.ID}).Error)
	uploads := service.NewSuggestionAndPayPalAttachmentUploadService(repository.NewOutboxRepository(db), provider, "https://shop.example.test")
	suggestions := service.NewSuggestionFeedbackService(repository.NewSuggestionFeedbackRepository(db))
	suggestions.ConfigureSuggestionAttachmentDurableUploads(uploads)
	access := service.NewPublicUploadAccessService(nil, nil)
	access.ConfigureSuggestionAndPayPalAttachmentUploads(uploads)
	router := gin.New()
	registerLocalUploadsRoute(router, &app.Dependencies{Services: app.Services{PublicUploadAccess: access}})
	handler := suggestionapi.NewHandler(suggestions, provider)
	router.POST("/test/feedback/upload", func(c *gin.Context) { c.Set("user_id", uint(71)) }, handler.Upload)
	var source bytes.Buffer
	require.NoError(t, png.Encode(&source, image.NewNRGBA(image.Rect(0, 0, 24, 24))))
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", "feedback.png")
	require.NoError(t, err)
	_, err = part.Write(source.Bytes())
	require.NoError(t, err)
	require.NoError(t, form.Close())
	request := httptest.NewRequest(http.MethodPost, "/test/feedback/upload", &body)
	request.Header.Set("Content-Type", form.FormDataContentType())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var attachment suggestionfeedback.Attachment
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &attachment))
	location, err := url.Parse(attachment.URL)
	require.NoError(t, err)
	readObject := func(path string, status int, expected []byte) {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, status, response.Code, response.Body.String())
		require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
		if status == http.StatusOK {
			require.Equal(t, expected, response.Body.Bytes())
			require.Equal(t, "nosniff", response.Header().Get("X-Content-Type-Options"))
		}
	}
	readObject(location.Path, http.StatusOK, source.Bytes())
	feedback := suggestionfeedback.SuggestionFeedback{UserID: 71, Message: "local route feedback"}
	require.NoError(t, suggestions.CreateWithAuthorizedSuggestionAttachmentReceipts(t.Context(), &feedback, []suggestionfeedback.Attachment{attachment}))
	readObject(location.Path, http.StatusOK, source.Bytes())
	require.NoError(t, db.Delete(&feedback).Error)
	readObject(location.Path, http.StatusNotFound, nil)
	invoiceBytes := []byte("%PDF-local-route-invoice")
	invoiceKey, err := storage.AllocateSuggestionOrPayPalAttachmentObjectKey(storage.PayPalDisputeInvoiceAttachmentNamespace, 81, "invoice.pdf")
	require.NoError(t, err)
	_, err = provider.(storage.PreallocatedSuggestionOrPayPalAttachmentWriter).WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(t.Context(), bytes.NewReader(invoiceBytes), invoiceKey)
	require.NoError(t, err)
	readObject("/uploads/"+invoiceKey, http.StatusNotFound, nil)
	payload, _ := json.Marshal(map[string]any{"request": map[string]any{"documents": []map[string]string{{"type": "commercial_invoice", "url": provider.GetURL(invoiceKey)}}}})
	snapshot := orderevidence.OrderEvidenceSubmissionSnapshot{Provider: "paypal", DisputeID: 81, OrderID: 1, Version: 1, Status: "locked", SchemaVersion: 1, LockedAt: time.Now(), SnapshotData: payload, SnapshotSHA256: fmt.Sprintf("%x", sha256.Sum256(payload))}
	require.NoError(t, db.Create(&snapshot).Error)
	readObject("/uploads/"+invoiceKey, http.StatusOK, invoiceBytes)
}

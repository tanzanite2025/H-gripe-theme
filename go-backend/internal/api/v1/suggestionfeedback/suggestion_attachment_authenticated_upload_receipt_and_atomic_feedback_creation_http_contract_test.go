package suggestionfeedback

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestSuggestionAttachmentHTTPUploadAndCreateRequireEligibleSameUploaderAndConsumeDurableReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:suggestion-attachment-http-"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&outbox.Event{}, &suggestionfeedback.SuggestionFeedback{}, &loyalty.MemberLevel{}, &loyalty.UserLoyalty{}))
	level := loyalty.MemberLevel{Name: "silver"}
	require.NoError(t, db.Create(&level).Error)
	for _, id := range []uint{71, 72} {
		require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: id, MemberLevelID: level.ID}).Error)
	}
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://cdn.example.test"})
	require.NoError(t, err)
	svc := service.NewSuggestionFeedbackService(repository.NewSuggestionFeedbackRepository(db))
	svc.ConfigureSuggestionAttachmentDurableUploads(service.NewSuggestionAndPayPalAttachmentUploadService(repository.NewOutboxRepository(db), provider, "https://shop.example.test"))
	handler := NewHandler(svc, provider)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if id, err := strconv.ParseUint(c.GetHeader("X-Synthetic-Test-User"), 10, 64); err == nil {
			c.Set("user_id", uint(id))
		}
	})
	router.POST("/upload", handler.Upload)
	router.POST("/feedback", handler.Create)
	var imageBytes bytes.Buffer
	require.NoError(t, png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 24, 24))))
	createRequest := func(method, path string, user uint, payload []byte, contentType string) *httptest.ResponseRecorder {
		request := httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(payload))
		request.Header.Set("Content-Type", contentType)
		if user != 0 {
			request.Header.Set("X-Synthetic-Test-User", strconv.FormatUint(uint64(user), 10))
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		return recorder
	}
	uploadRequest := func(user uint) *httptest.ResponseRecorder {
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		part, err := form.CreateFormFile("file", "feedback.png")
		require.NoError(t, err)
		_, err = part.Write(imageBytes.Bytes())
		require.NoError(t, err)
		require.NoError(t, form.Close())
		return createRequest(http.MethodPost, "/upload", user, body.Bytes(), form.FormDataContentType())
	}
	t.Run("anonymous upload denied", func(t *testing.T) { require.Equal(t, http.StatusUnauthorized, uploadRequest(0).Code) })
	t.Run("membership checked before provider write", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, uploadRequest(99).Code)
		var count int64
		require.NoError(t, db.Model(&outbox.Event{}).Count(&count).Error)
		require.Zero(t, count)
	})
	response := uploadRequest(71)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var uploaded suggestionfeedback.Attachment
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &uploaded))
	require.NotEmpty(t, uploaded.URL)
	encoded, _ := json.Marshal(map[string]any{"message": "attachment receipt HTTP contract", "attachments": []suggestionfeedback.Attachment{uploaded}})
	t.Run("different member cannot bind receipt", func(t *testing.T) {
		response := createRequest(http.MethodPost, "/feedback", 72, encoded, "application/json")
		require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	})
	t.Run("same member binds and consumes receipt", func(t *testing.T) {
		response := createRequest(http.MethodPost, "/feedback", 71, encoded, "application/json")
		require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
		var count int64
		require.NoError(t, db.Model(&outbox.Event{}).Count(&count).Error)
		require.Zero(t, count)
		var row suggestionfeedback.SuggestionFeedback
		require.NoError(t, db.First(&row).Error)
		require.JSONEq(t, string(suggestionfeedback.JSONFromAttachments([]suggestionfeedback.Attachment{uploaded})), string(row.Attachments))
	})
	t.Run("consumed receipt cannot create second placement", func(t *testing.T) {
		require.Equal(t, http.StatusBadRequest, createRequest(http.MethodPost, "/feedback", 71, encoded, "application/json").Code)
	})
}

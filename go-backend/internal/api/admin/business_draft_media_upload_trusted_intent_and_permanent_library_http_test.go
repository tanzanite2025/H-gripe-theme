package admin

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBusinessDraftMediaHTTPIntentAndIdentityComeFromAuthorizedHandlerOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, draft := range []bool{false, true} {
		t.Run(map[bool]string{false: "permanent library", true: "unsaved editor"}[draft], func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			sqlDB.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = sqlDB.Close() })
			require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.BusinessDraftMediaUploadReceipt{}, &outbox.Event{}))
			store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://http.example.test"})
			require.NoError(t, err)
			assets := service.NewMediaService(repository.NewMediaRepository(db), store, nil, "https://http.example.test", 20<<30)
			assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			handler := NewMediaHandler(assets)
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			for key, value := range map[string]string{"media_type": "image", "upload_origin": "customer_service_conversation_upload", "uploader_id": "999", "receipt_id": "forged", "business_draft": "true", "expires_at": "2099-01-01"} {
				require.NoError(t, writer.WriteField(key, value))
			}
			file, err := writer.CreateFormFile("file", "editor.png")
			require.NoError(t, err)
			require.NoError(t, png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 2))))
			require.NoError(t, writer.Close())
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/media/uploads", &body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())
			c.Set("user_id", uint(42))
			operation := handler.UploadOperationalMediaAssetForAuthorizedOrigin(media.MediaUploadOriginAdministratorCatalog)
			if draft {
				operation = handler.UploadBusinessDraftMediaAssetForAuthorizedOrigin(media.MediaUploadOriginAdministratorCatalog)
			}
			operation(c)
			require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
			var asset media.MediaAsset
			require.NoError(t, db.First(&asset).Error)
			require.EqualValues(t, 42, asset.UploaderID)
			require.Equal(t, media.MediaUploadOriginAdministratorCatalog, asset.UploadOrigin)
			var receipts []media.BusinessDraftMediaUploadReceipt
			require.NoError(t, db.Find(&receipts).Error)
			if draft {
				require.Len(t, receipts, 1)
				require.Equal(t, "ready", receipts[0].Status)
				require.NotEqual(t, "forged", receipts[0].ID)
				require.Equal(t, asset.ID, *receipts[0].AssetID)
			} else {
				require.Empty(t, receipts)
			}
		})
	}
}

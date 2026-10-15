package admin

import (
	"bytes"
	"image"
	"image/png"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaUploadAdministratorIdentityComesOnlyFromAuthenticatedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, authenticated := range []bool{true, false} {
		t.Run(map[bool]string{true: "authenticated administrator", false: "missing authenticated identity"}[authenticated], func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			sqlDB.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = sqlDB.Close() })
			require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}))
			uploadRoot := t.TempDir()
			storageService, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: uploadRoot, BaseURL: "https://shop.example.test"})
			require.NoError(t, err)
			handler := NewMediaHandler(service.NewMediaService(repository.NewMediaRepository(db), storageService, nil, "https://shop.example.test", 20<<30))
			var payload bytes.Buffer
			writer := multipart.NewWriter(&payload)
			require.NoError(t, writer.WriteField("media_type", "image"))
			require.NoError(t, writer.WriteField("uploader_id", "999"))
			require.NoError(t, writer.WriteField("uploader_actor_type", "visitor"))
			require.NoError(t, writer.WriteField("uploader_visitor_session_hash", strings.Repeat("a", 64)))
			file, err := writer.CreateFormFile("file", "administrator-material.png")
			require.NoError(t, err)
			require.NoError(t, png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 2))))
			require.NoError(t, writer.Close())
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/media", &payload)
			context.Request.Header.Set("Content-Type", writer.FormDataContentType())
			if authenticated {
				context.Set("user_id", uint(42))
			}
			handler.UploadAsset(context)
			if authenticated {
				require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
				var asset media.MediaAsset
				require.NoError(t, db.First(&asset).Error)
				require.Equal(t, media.MediaUploadActorAdministrator, asset.UploaderActorType)
				require.EqualValues(t, 42, asset.UploaderID)
				require.Empty(t, asset.UploaderVisitorSessionHash)
			} else {
				require.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())
				var count int64
				require.NoError(t, db.Model(&media.MediaAsset{}).Count(&count).Error)
				require.Zero(t, count)
				require.NoError(t, filepath.WalkDir(uploadRoot, func(_ string, entry fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					require.True(t, entry.IsDir(), "unauthenticated upload created a physical file")
					return nil
				}))
			}
		})
	}
}

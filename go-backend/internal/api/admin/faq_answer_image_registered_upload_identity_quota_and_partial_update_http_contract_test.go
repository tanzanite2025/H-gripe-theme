package admin

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"commerce-platform/internal/testfixtures"
)

func createFAQManagedMediaHTTPContractTestHandler(t *testing.T, quota int64) (*gorm.DB, *FAQHandler, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&faq.FAQ{}, &faq.FAQPage{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.BusinessDraftMediaUploadReceipt{}, &outbox.Event{}))
	root := t.TempDir()
	files, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	faqs := service.NewFAQService(repository.NewFAQRepository(db))
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), files, nil, "https://shop.example.test", quota)
	mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	faqs.ConfigureMediaService(mediaService)
	require.NoError(t, db.Create(&faq.FAQPage{PageID: "media-test", Locale: "en", RoutePath: "/support/media", Title: "Media", Status: "active"}).Error)
	return db, NewFAQHandler(faqs), root
}

func TestFAQImageUploadRegistersTrustedContentOriginAndRejectsMissingIdentityOrExceededQuota(t *testing.T) {
	for _, scenario := range []string{"authenticated", "missing identity", "quota exceeded"} {
		t.Run(scenario, func(t *testing.T) {
			quota := int64(20 << 30)
			if scenario == "quota exceeded" {
				quota = 1
			}
			db, handler, root := createFAQManagedMediaHTTPContractTestHandler(t, quota)
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			for key, value := range map[string]string{"alt": "FAQ illustration", "uploader_id": "999", "uploader_actor_type": "visitor", "library_partition": "catalog", "upload_origin": "customer_service_conversation"} {
				require.NoError(t, writer.WriteField(key, value))
			}
			file, err := writer.CreateFormFile("file", "answer.webp")
			require.NoError(t, err)
			_, err = file.Write(testfixtures.FAQAnswerImageUploadRegistrationAndBinding800By800LosslessWebPFixtureBytes())
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/faqs/answer-image", &body)
			context.Request.Header.Set("Content-Type", writer.FormDataContentType())
			if scenario != "missing identity" {
				context.Set("user_id", uint(42))
			}
			handler.UploadAnswerImage(context)
			status := map[string]int{"authenticated": http.StatusCreated, "missing identity": http.StatusUnauthorized, "quota exceeded": http.StatusRequestEntityTooLarge}[scenario]
			require.Equal(t, status, recorder.Code, recorder.Body.String())
			if scenario == "authenticated" {
				var asset media.MediaAsset
				require.NoError(t, db.First(&asset).Error)
				require.Equal(t, media.MediaLibraryPartitionContent, asset.LibraryPartition)
				require.Equal(t, media.MediaUploadOriginAdministratorContent, asset.UploadOrigin)
				require.Equal(t, media.MediaUploadActorAdministrator, asset.UploaderActorType)
				require.EqualValues(t, 42, asset.UploaderID)
				require.Empty(t, asset.UploaderVisitorSessionHash)
				var response struct {
					Image struct {
						URL, Alt      string
						Width, Height int
					}
				}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
				require.Equal(t, asset.URL, response.Image.URL)
				require.Equal(t, 800, response.Image.Width)
				require.Equal(t, 800, response.Image.Height)
				require.Equal(t, "FAQ illustration", response.Image.Alt)
			} else {
				var count int64
				require.NoError(t, db.Model(&media.MediaAsset{}).Count(&count).Error)
				require.Zero(t, count)
				require.NoError(t, filepath.WalkDir(root, func(_ string, entry fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					require.True(t, entry.IsDir(), "rejected upload must not leave a file")
					return nil
				}))
			}
		})
	}
}

func TestFAQMediaHTTPValidationAndOmittedImagePreservation(t *testing.T) {
	db, handler, _ := createFAQManagedMediaHTTPContractTestHandler(t, 20<<30)
	legacy := &faq.FAQ{Question: "Legacy", Answer: "<p>Old answer</p>", AnswerImageURL: "/uploads/old/answer.webp", AnswerImageAlt: "Old", AnswerImageWidth: 800, AnswerImageHeight: 800, Locale: "en", PageID: "media-test", Status: "draft"}
	require.NoError(t, db.Create(legacy).Error)
	for _, scenario := range []struct {
		name, body string
		status     int
	}{
		{"omitted image", `{"question":"Ordinary edit"}`, http.StatusOK},
		{"invalid dedicated image", `{"answer_image_url":"https://foreign.example.test/uploads/old/answer.webp","answer_image_alt":"Forged","answer_image_width":800,"answer_image_height":800}`, http.StatusBadRequest},
		{"invalid answer link", `{"answer":"<a href=\"/uploads/missing.webp\">Download</a>"}`, http.StatusBadRequest},
		{"explicit detach", `{"answer_image_url":""}`, http.StatusOK},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Params = gin.Params{{Key: "id", Value: "1"}}
			context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/faqs/1", strings.NewReader(scenario.body))
			context.Request.Header.Set("Content-Type", "application/json")
			handler.UpdateFAQ(context)
			require.Equal(t, scenario.status, recorder.Code, recorder.Body.String())
			var current faq.FAQ
			require.NoError(t, db.First(&current, legacy.ID).Error)
			if scenario.name == "explicit detach" {
				require.Empty(t, current.AnswerImageURL)
			} else {
				require.Equal(t, legacy.AnswerImageURL, current.AnswerImageURL)
			}
		})
	}
}

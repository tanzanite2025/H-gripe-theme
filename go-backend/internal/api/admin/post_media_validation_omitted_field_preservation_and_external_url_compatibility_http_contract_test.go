package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func createPostManagedMediaHTTPContractTestHandler(t *testing.T) (*gorm.DB, *ContentHandler, *service.MediaService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	// Create the explicit relationship schema before Post can infer a join table.
	require.NoError(t, db.AutoMigrate(&post.Category{}, &post.PostCategory{}, &post.Post{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}))
	files, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	categories := repository.NewBlogCategoryRepository(db)
	posts := repository.NewPostRepository(db)
	posts.ConfigureCategoryRepository(categories)
	articles := service.NewPostService(posts, nil, 0, categories)
	assets := service.NewMediaService(repository.NewMediaRepository(db), files, nil, "https://shop.example.test", 20<<30)
	articles.ConfigureMediaService(assets)
	return db, NewContentHandler(articles), assets
}

func performPostMediaHTTPContractRequest(t *testing.T, handler *ContentHandler, id uint, payload map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	method, path := http.MethodPost, "/content/posts"
	if id != 0 {
		method, path = http.MethodPut, fmt.Sprintf("/content/posts/%d", id)
		context.Params = gin.Params{{Key: "id", Value: fmt.Sprint(id)}}
	}
	context.Set("user_id", uint(42))
	context.Request = httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	if id == 0 {
		handler.CreatePost(context)
	} else {
		handler.UpdatePost(context)
	}
	return recorder
}

func TestPostMediaHTTPRejectsInvalidCreateAndUpdateWithoutPartialWrites(t *testing.T) {
	db, handler, _ := createPostManagedMediaHTTPContractTestHandler(t)
	legacy := &post.Post{Title: "Original", Slug: "original", Locale: "en", Content: `<p>Old</p><img src="/uploads/legacy.png">`, FeaturedImg: "/uploads/legacy-cover.png"}
	require.NoError(t, db.Create(legacy).Error)
	for _, field := range []string{"content", "featured_image", "category_ids"} {
		for _, id := range []uint{0, legacy.ID} {
			t.Run(fmt.Sprintf("%s/%d", field, id), func(t *testing.T) {
				payload := map[string]interface{}{"title": "Must roll back"}
				if id == 0 {
					payload["slug"], payload["status"], payload["locale"] = "invalid-new", "draft", "en"
				}
				switch field {
				case "content":
					payload[field] = `<a href="/uploads/missing.pdf">Download</a>`
				case "featured_image":
					payload[field] = "/uploads/missing.png"
				case "category_ids":
					payload[field] = []uint{999}
				}
				response := performPostMediaHTTPContractRequest(t, handler, id, payload)
				require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
				var count int64
				require.NoError(t, db.Unscoped().Model(&post.Post{}).Count(&count).Error)
				require.EqualValues(t, 1, count)
				var stored post.Post
				require.NoError(t, db.First(&stored, legacy.ID).Error)
				require.Equal(t, legacy.Title, stored.Title)
				require.Equal(t, legacy.Content, stored.Content)
				require.Equal(t, legacy.FeaturedImg, stored.FeaturedImg)
			})
		}
	}
	response := performPostMediaHTTPContractRequest(t, handler, legacy.ID, map[string]interface{}{"title": "Ordinary edit"})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var stored post.Post
	require.NoError(t, db.First(&stored, legacy.ID).Error)
	require.Equal(t, legacy.Content, stored.Content)
	require.Equal(t, legacy.FeaturedImg, stored.FeaturedImg)
	response = performPostMediaHTTPContractRequest(t, handler, legacy.ID, map[string]interface{}{"content": "", "featured_image": ""})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.NoError(t, db.First(&stored, legacy.ID).Error)
	require.Empty(t, stored.Content)
	require.Empty(t, stored.FeaturedImg)
}

func TestPostMediaHTTPPreservesExternalURLsWithoutClaimingSamePathManagedAssets(t *testing.T) {
	db, handler, assets := createPostManagedMediaHTTPContractTestHandler(t)
	asset := media.MediaAsset{Filename: "external-path.png", StorageKey: "content/external-path.png", URL: "https://media.example.test/uploads/content/external-path.png", MediaType: "image", Status: "active", Visibility: "public"}
	require.NoError(t, db.Create(&asset).Error)
	foreignURL := "https://external.example.test/uploads/" + asset.StorageKey
	body := `<img src="` + foreignURL + `"><a href="https://external.example.test/manual.pdf#page=2">Download</a>`
	response := performPostMediaHTTPContractRequest(t, handler, 0, map[string]interface{}{"title": "External", "slug": "external", "status": "draft", "locale": "en", "content": body, "featured_image": foreignURL})
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var stored post.Post
	require.NoError(t, db.First(&stored).Error)
	require.Equal(t, foreignURL, stored.FeaturedImg)
	require.Contains(t, stored.Content, foreignURL)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Zero(t, report.Total)
}

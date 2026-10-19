package workbenchfeed

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
	feed "commerce-platform/internal/workbenchfeed"
)

func TestWorkbenchFeedInvalidMediaBindingReturnsBadRequestWithoutPartialEntryOrChildWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	connection, err := db.DB()
	require.NoError(t, err)
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = connection.Close() })
	require.NoError(t, db.AutoMigrate(&feed.FeedEntry{}, &feed.FeedMedia{}, &feed.TaggedProduct{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}))
	service := feed.NewService(db, feed.NewRepository(db), feed.NewCatalog(db), nil)
	handler := NewHandler(service)
	router := gin.New()
	router.POST("/entries", handler.Create)
	router.PUT("/entries/:id", handler.Update)
	entry, err := service.Create(feed.EntryInput{Content: "Existing entry", Status: feed.StatusDraft})
	require.NoError(t, err)
	for _, mutation := range []struct{ method, url string }{{http.MethodPost, "/entries"}, {http.MethodPut, fmt.Sprintf("/entries/%d", entry.ID)}} {
		body, err := json.Marshal(map[string]any{"content": "Must roll back", "status": "published", "media": []map[string]any{{"file_path": "/uploads/workbench-feed/missing.png", "width": 64, "height": 32}}})
		require.NoError(t, err)
		request := httptest.NewRequestWithContext(t.Context(), mutation.method, mutation.url, bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusBadRequest, response.Code)
		var entries, placements int64
		require.NoError(t, db.Unscoped().Model(&feed.FeedEntry{}).Count(&entries).Error)
		require.NoError(t, db.Unscoped().Model(&feed.FeedMedia{}).Count(&placements).Error)
		require.EqualValues(t, 1, entries)
		require.Zero(t, placements)
		stored, err := service.Get(entry.ID)
		require.NoError(t, err)
		require.Equal(t, entry.Content, stored.Content)
		require.Equal(t, entry.Status, stored.Status)
	}
}

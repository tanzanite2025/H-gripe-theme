package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaImageDimensionsHandlerListIncludesPresetDefinitions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&mediadomain.MediaAsset{}, &mediadomain.MediaAssetDerivative{}))

	mediaService := service.NewMediaService(
		repository.NewMediaRepository(db),
		nil,
		nil,
		"https://shop.example.test",
		20<<30,
	)
	handler := NewMediaImageDimensionsHandler(service.NewMediaImageDimensionEngine(mediaService))

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/preflight/image-dimensions", nil)

	handler.List(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			Presets []service.MediaDerivativePresetDefinition `json:"presets"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, service.MediaDerivativePresetDefinitions(), body.Data.Presets)
}

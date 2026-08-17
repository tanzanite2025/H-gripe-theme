package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"commerce-platform/internal/domain/audit"
	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaDerivativePresetHandlerListsAndAuditsChanges(t *testing.T) {
	handler, auditRecorder := newMediaDerivativePresetHandlerForTest(t)

	listRecorder := httptest.NewRecorder()
	listContext, _ := gin.CreateTestContext(listRecorder)
	listContext.Request = httptest.NewRequest(http.MethodGet, "/api/admin/media/derivative-presets", nil)
	handler.ListDerivativePresets(listContext)

	require.Equal(t, http.StatusOK, listRecorder.Code)
	var listBody struct {
		Code int `json:"code"`
		Data struct {
			Presets []mediadomain.MediaDerivativePreset `json:"presets"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(listRecorder.Body.Bytes(), &listBody))
	require.Equal(t, 0, listBody.Code)
	require.Len(t, listBody.Data.Presets, 3)

	createRecorder := httptest.NewRecorder()
	createContext, _ := gin.CreateTestContext(createRecorder)
	createContext.Set("user_id", uint(7))
	createContext.Set("username", "media-admin")
	createContext.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/media/derivative-presets",
		strings.NewReader(`{"code":"article-banner","label":"文章横幅","max_width":1280,"sort_order":45,"enabled":true}`),
	)
	createContext.Request.Header.Set("Content-Type", "application/json")
	handler.CreateDerivativePreset(createContext)

	require.Equal(t, http.StatusCreated, createRecorder.Code)
	var createBody struct {
		Code int `json:"code"`
		Data struct {
			Preset mediadomain.MediaDerivativePreset `json:"preset"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(createRecorder.Body.Bytes(), &createBody))
	require.Equal(t, 0, createBody.Code)
	require.Equal(t, "article-banner", createBody.Data.Preset.Code)
	require.Len(t, auditRecorder.logs, 1)
	require.Equal(t, adminAuditActionCreate, auditRecorder.logs[0].Action)
	require.Equal(t, adminAuditResourceMediaDerivativePreset, auditRecorder.logs[0].Resource)
	require.Equal(t, adminAuditStatusSuccess, auditRecorder.logs[0].Status)

	updateRecorder := httptest.NewRecorder()
	updateContext, _ := gin.CreateTestContext(updateRecorder)
	updateContext.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(createBody.Data.Preset.ID), 10)}}
	updateContext.Set("user_id", uint(7))
	updateContext.Set("username", "media-admin")
	updateContext.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/admin/media/derivative-presets/"+strconv.FormatUint(uint64(createBody.Data.Preset.ID), 10),
		strings.NewReader(`{"label":"文章页横幅","max_width":1440,"sort_order":50,"enabled":true}`),
	)
	updateContext.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateDerivativePreset(updateContext)

	require.Equal(t, http.StatusOK, updateRecorder.Code)
	require.Len(t, auditRecorder.logs, 2)
	log := auditRecorder.logs[1]
	require.Equal(t, adminAuditActionUpdate, log.Action)
	require.Equal(t, adminAuditResourceMediaDerivativePreset, log.Resource)
	require.Equal(t, createBody.Data.Preset.ID, log.ResourceID)
	require.Equal(t, adminAuditStatusSuccess, log.Status)

	var newValue mediadomain.MediaDerivativePreset
	require.NoError(t, json.Unmarshal([]byte(log.NewValue), &newValue))
	require.Equal(t, "文章页横幅", newValue.Label)
	require.Equal(t, 1440, newValue.MaxWidth)
	require.Equal(t, 2, newValue.GenerationVersion)
}

type mediaDerivativePresetAuditRecorder struct {
	logs []*audit.AuditLog
}

func (r *mediaDerivativePresetAuditRecorder) CreateAuditLog(log *audit.AuditLog) error {
	copy := *log
	r.logs = append(r.logs, &copy)
	return nil
}

func newMediaDerivativePresetHandlerForTest(t *testing.T) (*MediaHandler, *mediaDerivativePresetAuditRecorder) {
	t.Helper()
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
	require.NoError(t, db.AutoMigrate(
		&mediadomain.MediaAsset{},
		&mediadomain.MediaAssetDerivative{},
		&mediadomain.MediaDerivativePreset{},
	))

	presetRepo := repository.NewMediaDerivativePresetRepository(db)
	require.NoError(t, service.SeedDefaultMediaDerivativePresets(presetRepo))
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), nil, nil, "https://shop.example.test", 20<<30)
	mediaService.ConfigureDerivativePresetRepository(presetRepo)

	auditRecorder := &mediaDerivativePresetAuditRecorder{}
	handler := NewMediaHandler(mediaService)
	handler.ConfigureAuditService(auditRecorder)
	return handler, auditRecorder
}

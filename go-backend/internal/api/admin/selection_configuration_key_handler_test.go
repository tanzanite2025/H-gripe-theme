package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	selectionconfiguration "commerce-platform/internal/domain/selectionconfiguration"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSelectionConfigurationKeyHandlerListOptionsReturnsEnabledKeysArray(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newSelectionConfigurationKeyHandlerTestDatabase(t)
	repo := repository.NewSelectionConfigurationKeyRepository(db)
	svc := service.NewSelectionConfigurationKeyService(repo)

	_, err := svc.CreateSelectionConfigurationKey(service.SelectionConfigurationKeyInput{
		Kind:         service.SelectionConfigurationKeyKindQuestionKey,
		Code:         "rear_axle",
		DisplayLabel: "后轴规格",
		Description:  "启用项",
		IsEnabled:    true,
		SortOrder:    10,
	})
	require.NoError(t, err)
	_, err = svc.CreateSelectionConfigurationKey(service.SelectionConfigurationKeyInput{
		Kind:         service.SelectionConfigurationKeyKindQuestionKey,
		Code:         "front_axle",
		DisplayLabel: "前轴规格",
		Description:  "停用项",
		IsEnabled:    false,
		SortOrder:    20,
	})
	require.NoError(t, err)

	handler := NewSelectionConfigurationKeyHandler(svc)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/selection-configuration/keys/options?kind=question_key", nil)

	handler.ListOptions(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))

	data, ok := payload["data"].([]any)
	require.True(t, ok, "response data should be an array")
	require.Len(t, data, 1)

	option, ok := data[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "rear_axle", option["code"])
	assert.Equal(t, "后轴规格", option["display_label"])
}

func newSelectionConfigurationKeyHandlerTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

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

	require.NoError(t, db.AutoMigrate(&selectionconfiguration.SelectionConfigurationKey{}))
	return db
}

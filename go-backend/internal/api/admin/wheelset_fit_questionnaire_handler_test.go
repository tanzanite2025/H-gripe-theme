package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWheelsetFitQuestionnaireHandlerGetCurrentVersionReturnsNullWhenMissingVersion(t *testing.T) {
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
		&wheelsetfit.Questionnaire{},
		&wheelsetfit.Version{},
		&wheelsetfit.Question{},
		&wheelsetfit.Option{},
		&wheelsetfit.QuestionTranslation{},
		&wheelsetfit.OptionTranslation{},
	))

	require.NoError(t, db.Create(&wheelsetfit.Questionnaire{
		Slug:                wheelsetfit.QuestionnaireSlug,
		ProductCategorySlug: wheelsetfit.WheelsetProductCategorySlug,
		SourceLocale:        wheelsetfit.SourceLocale,
		IsEnabled:           true,
	}).Error)

	handler := NewWheelsetFitQuestionnaireHandler(
		service.NewWheelsetFitQuestionnaireService(repository.NewWheelsetFitQuestionnaireRepository(db)),
		nil,
	)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/wheelset-fit-questionnaire/current", nil)

	handler.GetCurrentVersion(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, float64(0), payload["code"])

	data, ok := payload["data"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, data, "data")
	require.Nil(t, data["data"])
}

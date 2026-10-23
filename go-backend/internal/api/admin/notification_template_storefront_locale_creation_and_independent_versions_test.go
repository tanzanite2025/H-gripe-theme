package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/notification"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestNotificationTemplateRoutesCreateAllStorefrontLocalesAndPreserveIndependentVersions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&notification.EmailTemplate{}, &notification.EmailTemplateVersion{}))
	templateService := service.NewTransactionalNotificationTemplateService(repository.NewNotificationTemplateRepository(db))
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_role", "admin"); c.Next() })
	registerNotificationTemplateRoutes(router.Group("/api/admin"), NewNotificationTemplateHandler(templateService))

	savedByLocale := make(map[string]notification.EmailTemplate)
	for _, locale := range locales.SupportedLocaleCodes() {
		input := map[string]any{
			"code": service.NotificationTemplateOrderConfirmation, "locale": locale, "category": "order",
			"name": "Order confirmation " + locale, "subject_template": locale + " {{order_number}}",
			"body_html": "<p>" + locale + " {{order_number}}</p>", "body_text": locale + " {{order_number}}",
			"allowed_variables":  []string{"order_number", "order_amount", "paid_at", "customer_name", "currency", "items"},
			"required_variables": []string{"order_number", "order_amount", "paid_at"},
			"is_enabled":         locale == "en", "version": 1,
		}
		body, marshalErr := json.Marshal(input)
		require.NoError(t, marshalErr)
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/admin/email/templates", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		var response struct {
			Data notification.EmailTemplate `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		require.Equal(t, locale, response.Data.Locale)
		require.Equal(t, locale == "en", response.Data.IsEnabled)
		require.Equal(t, 1, response.Data.Version)
		savedByLocale[locale] = response.Data
		var persisted notification.EmailTemplate
		require.NoError(t, db.First(&persisted, response.Data.ID).Error)
		require.Equal(t, locale == "en", persisted.IsEnabled)
	}
	require.Len(t, savedByLocale, 20)

	readOnlyRouter := gin.New()
	readOnlyRouter.Use(func(c *gin.Context) { c.Set("user_role", "viewer"); c.Next() })
	registerNotificationTemplateRoutes(readOnlyRouter.Group("/api/admin"), NewNotificationTemplateHandler(templateService))
	readOnlyRecorder := httptest.NewRecorder()
	readOnlyRouter.ServeHTTP(readOnlyRecorder, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/admin/email/templates", bytes.NewBufferString(`{}`)))
	require.Equal(t, http.StatusForbidden, readOnlyRecorder.Code, readOnlyRecorder.Body.String())
	invalidLocaleRecorder := httptest.NewRecorder()
	invalidLocaleRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/admin/email/templates", bytes.NewBufferString(`{"code":"order_confirmation","locale":"invalid","category":"order"}`))
	invalidLocaleRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(invalidLocaleRecorder, invalidLocaleRequest)
	require.Equal(t, http.StatusBadRequest, invalidLocaleRecorder.Code, invalidLocaleRecorder.Body.String())

	resolved, err := templateService.Resolve(service.NotificationTemplateOrderConfirmation, "fr")
	require.NoError(t, err)
	require.Equal(t, "en", resolved.Locale, "a disabled translated draft must keep using English")
	french := savedByLocale["fr"]
	updateBody, err := json.Marshal(map[string]any{
		"code": french.Code, "locale": french.Locale, "category": french.Category, "name": french.Name,
		"subject_template": "French revision {{order_number}}", "body_html": french.BodyHTML,
		"allowed_variables": json.RawMessage(french.AllowedVariables), "required_variables": json.RawMessage(french.RequiredVariables),
		"is_enabled": true, "version": 2,
	})
	require.NoError(t, err)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPut, fmt.Sprintf("/api/admin/email/templates/%d", french.ID), bytes.NewReader(updateBody))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	resolved, err = templateService.Resolve(service.NotificationTemplateOrderConfirmation, "fr-FR")
	require.NoError(t, err)
	require.Equal(t, "fr", resolved.Locale)
	require.Equal(t, "French revision {{order_number}}", resolved.SubjectTemplate)
	for locale, original := range savedByLocale {
		persisted, findErr := templateService.Get(original.ID)
		require.NoError(t, findErr)
		versions, versionsErr := templateService.Versions(original.ID)
		require.NoError(t, versionsErr)
		if locale == "fr" {
			require.Equal(t, 2, persisted.Version)
			require.Len(t, versions, 2)
		} else {
			require.Equal(t, original.SubjectTemplate, persisted.SubjectTemplate)
			require.Equal(t, 1, persisted.Version)
			require.Len(t, versions, 1)
		}
	}
}

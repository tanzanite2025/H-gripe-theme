package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRejectProductSEORequestFieldsBlocksCatalogSEOFields(t *testing.T) {
	for _, field := range productSEORequestFields {
		requestBody := map[string]json.RawMessage{
			"name": json.RawMessage(`"Product"`),
			field:  json.RawMessage(`"Managed by SEO"`),
		}

		blockedField, blocked := rejectProductSEORequestFields(requestBody)

		require.True(t, blocked)
		require.Equal(t, field, blockedField)
	}
}

func TestRejectProductSEORequestFieldsAllowsCatalogFields(t *testing.T) {
	requestBody := map[string]json.RawMessage{
		"name":              json.RawMessage(`"Product"`),
		"slug":              json.RawMessage(`"product"`),
		"short_description": json.RawMessage(`"Summary"`),
		"description":       json.RawMessage(`"Body"`),
		"status":            json.RawMessage(`"active"`),
		"locale":            json.RawMessage(`"en"`),
		"variants":          json.RawMessage(`[]`),
		"media":             json.RawMessage(`[]`),
	}

	blockedField, blocked := rejectProductSEORequestFields(requestBody)

	require.False(t, blocked)
	require.Empty(t, blockedField)
}

func TestPreviewProductTemplateSyncRejectsInvalidProductID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, value := range []string{"0", "not-a-number"} {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		context.Params = gin.Params{{Key: "id", Value: value}}

		handler := NewProductHandler(nil)
		handler.PreviewProductTemplateSync(context)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Contains(t, recorder.Body.String(), "Invalid product ID")
	}
}

func TestSyncProductTemplateValidatesExpectedRevisionBeforeServiceCall(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{`{}`, `{"expected_revision":0}`, `{"expected_revision":"4"}`} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/admin/products/7/template-sync", strings.NewReader(body))
		context, _ := gin.CreateTestContext(recorder)
		context.Request = request
		context.Params = gin.Params{{Key: "id", Value: "7"}}

		handler := NewProductHandler(nil)
		handler.SyncProductTemplate(context)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
	}
}

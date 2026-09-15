package seo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/audit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type seoAuditRecorderStub struct {
	logs []audit.AuditLog
}

func (r *seoAuditRecorderStub) CreateAuditLog(log *audit.AuditLog) error {
	if log != nil {
		r.logs = append(r.logs, *log)
	}
	return nil
}

func TestProductsHandlerPushIndexingRejectsUnsupportedProductPages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &seoAuditRecorderStub{}
	handler := NewProductsHandler(nil)
	handler.ConfigureAuditService(auditRecorder)

	response, context := newProductsIndexingTestContext("7")
	handler.PushIndexing(context)

	require.Equal(t, http.StatusUnprocessableEntity, response.Code)
	require.Contains(t, response.Body.String(), "google_indexing_product_unsupported")
	require.Len(t, auditRecorder.logs, 1)
	require.Equal(t, seoAuditStatusFailed, auditRecorder.logs[0].Status)
	require.NotContains(t, auditRecorder.logs[0].NewValue, "notification_type")
	require.Contains(t, auditRecorder.logs[0].ErrorMessage, "does not support product pages")
}

func newProductsIndexingTestContext(id string) (*httptest.ResponseRecorder, *gin.Context) {
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Params = gin.Params{{Key: "id", Value: id}}
	context.Set("user_id", uint(11))
	context.Set("username", "ops-admin")
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/seo/products/"+id+"/indexing", nil)
	return response, context
}

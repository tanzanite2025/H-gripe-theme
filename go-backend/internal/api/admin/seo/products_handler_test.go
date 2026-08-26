package seo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type googleIndexingHandlerStub struct {
	result *service.GoogleIndexingPushResult
	err    error
}

func (s googleIndexingHandlerStub) Status() service.GoogleIndexingStatus {
	return service.GoogleIndexingStatus{Enabled: true, Configured: true, Ready: true}
}

func (s googleIndexingHandlerStub) PushProduct(context.Context, uint) (*service.GoogleIndexingPushResult, error) {
	return s.result, s.err
}

type seoAuditRecorderStub struct {
	logs []audit.AuditLog
}

func (r *seoAuditRecorderStub) CreateAuditLog(log *audit.AuditLog) error {
	if log != nil {
		r.logs = append(r.logs, *log)
	}
	return nil
}

func TestProductsHandlerPushIndexingAuditsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &seoAuditRecorderStub{}
	handler := NewProductsHandler(nil)
	handler.ConfigureAuditService(auditRecorder)
	handler.googleIndexing = googleIndexingHandlerStub{
		result: &service.GoogleIndexingPushResult{
			ProductID:        7,
			URL:              "https://store.example.test/products/carbon-wheel",
			NotificationType: "URL_UPDATED",
			Accepted:         true,
			HTTPStatus:       http.StatusOK,
			SubmittedAt:      time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC),
		},
	}

	response, context := newProductsIndexingTestContext("7")
	handler.PushIndexing(context)

	require.Equal(t, http.StatusOK, response.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, seoAuditActionIndexing, log.Action)
	require.Equal(t, seoAuditResourceProduct, log.Resource)
	require.Equal(t, uint(7), log.ResourceID)
	require.Equal(t, seoAuditStatusOK, log.Status)

	var value map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.NewValue), &value))
	require.Equal(t, "https://store.example.test/products/carbon-wheel", value["url"])
	require.Equal(t, "URL_UPDATED", value["notification_type"])
	require.Equal(t, float64(http.StatusOK), value["http_status"])
	require.Equal(t, true, value["accepted"])
}

func TestProductsHandlerPushIndexingAuditsUpstreamFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &seoAuditRecorderStub{}
	handler := NewProductsHandler(nil)
	handler.ConfigureAuditService(auditRecorder)
	handler.googleIndexing = googleIndexingHandlerStub{
		result: &service.GoogleIndexingPushResult{
			ProductID:        7,
			URL:              "https://store.example.test/products/carbon-wheel",
			NotificationType: "URL_UPDATED",
			HTTPStatus:       http.StatusBadGateway,
		},
		err: fmt.Errorf("%w: timeout", service.ErrGoogleIndexingUpstream),
	}

	response, context := newProductsIndexingTestContext("7")
	handler.PushIndexing(context)

	require.Equal(t, http.StatusBadGateway, response.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, seoAuditActionIndexing, log.Action)
	require.Equal(t, seoAuditStatusFailed, log.Status)
	require.Contains(t, log.ErrorMessage, "timeout")
	require.Contains(t, log.NewValue, "carbon-wheel")
}

func TestProductsHandlerPushIndexingReturnsCooldownRetryAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &seoAuditRecorderStub{}
	handler := NewProductsHandler(nil)
	handler.ConfigureAuditService(auditRecorder)
	handler.googleIndexing = googleIndexingHandlerStub{
		result: &service.GoogleIndexingPushResult{
			ProductID:        7,
			URL:              "https://store.example.test/products/carbon-wheel",
			NotificationType: "URL_UPDATED",
		},
		err: &service.GoogleIndexingCooldownError{RetryAfter: 42 * time.Second},
	}

	response, context := newProductsIndexingTestContext("7")
	handler.PushIndexing(context)

	require.Equal(t, http.StatusConflict, response.Code)
	require.Equal(t, "42", response.Header().Get("Retry-After"))
	require.Contains(t, response.Body.String(), "google_indexing_recently_notified")
	require.Len(t, auditRecorder.logs, 1)
	require.Equal(t, seoAuditStatusFailed, auditRecorder.logs[0].Status)
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

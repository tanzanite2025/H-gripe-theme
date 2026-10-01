package wheelsetlacing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func TestListTopologiesReturnsCacheableContract(t *testing.T) {
	router := newTestRouter()
	first := performRequest(router, http.MethodGet, "/api/v1/wheelset-lacing/topologies", "")
	if first.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", first.Code, http.StatusOK, first.Body.String())
	}
	if first.Header().Get("ETag") == "" || first.Header().Get("Cache-Control") != "public, max-age=86400" {
		t.Fatalf("missing cache headers: etag=%q cache=%q", first.Header().Get("ETag"), first.Header().Get("Cache-Control"))
	}
	for _, fragment := range []string{`"topologies"`, `"21h-g3-2to1"`, `"uniform_2to1"`, `"contract_version":"v1.0"`, `"length_unit":"mm"`} {
		if !strings.Contains(first.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, first.Body.String())
		}
	}
	if strings.Contains(first.Body.String(), `"spoke_length"`) || strings.Contains(first.Body.String(), `"tension"`) {
		t.Fatalf("topology contract must not expose physical calculations: %s", first.Body.String())
	}

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/wheelset-lacing/topologies", nil)
	request.Header.Set("If-None-Match", first.Header().Get("ETag"))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotModified {
		t.Fatalf("cached status = %d, want %d", response.Code, http.StatusNotModified)
	}
}

func TestValidateReturnsG3Mapping(t *testing.T) {
	router := newTestRouter()
	response := performRequest(router, http.MethodPost, "/api/v1/wheelset-lacing/validate", `{"topology_id":"21h-g3-2to1","hole_count":21,"cross":2,"distribution":"g3_2to1"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	for _, fragment := range []string{`"valid":true`, `"hole_count":21`, `"hub_holes_a"`, `"hub_holes_b"`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, response.Body.String())
		}
	}
}

func TestValidateFailsLoudlyForUnknownAndMismatchedSelections(t *testing.T) {
	router := newTestRouter()
	unknown := performRequest(router, http.MethodPost, "/api/v1/wheelset-lacing/validate", `{"topology_id":"24h-g3-2to1"}`)
	if unknown.Code != http.StatusNotFound || !strings.Contains(unknown.Body.String(), "UNKNOWN_WHEELSET_LACING_TOPOLOGY") {
		t.Fatalf("unknown response = %d %s", unknown.Code, unknown.Body.String())
	}
	mismatch := performRequest(router, http.MethodPost, "/api/v1/wheelset-lacing/validate", `{"topology_id":"24h-uniform-2to1","cross":3}`)
	if mismatch.Code != http.StatusUnprocessableEntity || !strings.Contains(mismatch.Body.String(), "TOPOLOGY_SELECTION_MISMATCH") {
		t.Fatalf("mismatch response = %d %s", mismatch.Code, mismatch.Body.String())
	}
	unknownField := performRequest(router, http.MethodPost, "/api/v1/wheelset-lacing/validate", `{"topology_id":"21h-g3-2to1","erd":600}`)
	if unknownField.Code != http.StatusBadRequest || !strings.Contains(unknownField.Body.String(), "INVALID_WHEELSET_LACING_REQUEST") {
		t.Fatalf("unknown field response = %d %s", unknownField.Code, unknownField.Body.String())
	}
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(service.NewWheelsetLacingService())
	handler.RegisterRoutes(router.Group("/api/v1/wheelset-lacing"))
	return router
}

func performRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

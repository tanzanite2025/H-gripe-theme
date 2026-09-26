package fitmentdrivetrain

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/fitmentcatalog"

	"github.com/gin-gonic/gin"
)

func TestCalculateReturnsRecommendedOptionAndMatrixData(t *testing.T) {
	router := newTestRouter()
	recorder := performRequest(router, http.MethodPost, "/api/v1/fitment/drivetrain/calculate", `{"brand":"Shimano","cassette_spec":"shimano_12s_road_11_34t"}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	body := recorder.Body.String()
	for _, fragment := range []string{`"recommended_freehub":"HG-11"`, `"thickness_mm":0`, `"rule_version":"v1.0"`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("body = %s, want fragment %s", body, fragment)
		}
	}
	if strings.Contains(body, `source_refs`) {
		t.Fatalf("public calculation response must not expose internal source_refs: %s", body)
	}
	if !strings.Contains(body, `"knowledge_as_of":"2026-09-25"`) {
		t.Fatalf("body = %s, want knowledge_as_of metadata", body)
	}
}

func TestCalculateRejectsTenToothCassetteForcedOntoHG(t *testing.T) {
	router := newTestRouter()
	recorder := performRequest(router, http.MethodPost, "/api/v1/fitment/drivetrain/calculate", `{"brand":"SRAM","cassette_spec":"sram_12s_10_52t","freehub_standard":"HG-11"}`)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusUnprocessableEntity, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "physical interference") {
		t.Fatalf("body = %s, want physical interference message", recorder.Body.String())
	}
}

func TestCalculateRejectsUnknownFreehubStandard(t *testing.T) {
	router := newTestRouter()
	recorder := performRequest(router, http.MethodPost, "/api/v1/fitment/drivetrain/calculate", `{"brand":"SRAM","cassette_spec":"sram_12s_10_52t","freehub_standard":"mystery-freehub"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "UNKNOWN_FREEHUB_STANDARD") {
		t.Fatalf("body = %s, want unknown freehub code", recorder.Body.String())
	}
}

func TestCalculateRejectsUnknownCassette(t *testing.T) {
	router := newTestRouter()
	recorder := performRequest(router, http.MethodPost, "/api/v1/fitment/drivetrain/calculate", `{"brand":"SRAM","cassette_spec":"unknown"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), "UNKNOWN_CASSETTE_SPEC") {
		t.Fatalf("body = %s, want unknown cassette code", recorder.Body.String())
	}
}

func TestMatrixSupportsETagCaching(t *testing.T) {
	router := newTestRouter()
	first := performRequest(router, http.MethodGet, "/api/v1/fitment/drivetrain/matrix", "")
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", first.Code, http.StatusOK)
	}
	etag := first.Header().Get("ETag")
	if etag == "" || !strings.Contains(first.Body.String(), `"rules"`) {
		t.Fatalf("first response missing ETag or rules: etag=%q body=%s", etag, first.Body.String())
	}
	if !strings.Contains(first.Body.String(), `"knowledge_as_of":"2026-09-25"`) {
		t.Fatalf("matrix response missing knowledge_as_of: %s", first.Body.String())
	}
	if strings.Contains(first.Body.String(), `source_refs`) {
		t.Fatalf("public matrix response must not expose internal source_refs: %s", first.Body.String())
	}
	if got := first.Header().Get("Cache-Control"); got != "public, max-age=86400" {
		t.Fatalf("Cache-Control = %q, want public, max-age=86400", got)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/fitment/drivetrain/matrix", nil)
	request.Header.Set("If-None-Match", etag)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotModified {
		t.Fatalf("cached status = %d, want %d", recorder.Code, http.StatusNotModified)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=86400" {
		t.Fatalf("cached Cache-Control = %q, want public, max-age=86400", got)
	}
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(fitmentcatalog.NewDefaultDrivetrainFitmentEngine())
	handler.RegisterRoutes(router.Group("/api/v1/fitment/drivetrain"))
	return router
}

func performRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

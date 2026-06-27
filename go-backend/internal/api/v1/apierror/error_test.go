package apierror

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appLogger "tanzanite/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSendDoesNotLogRawUnhandledError(t *testing.T) {
	var output bytes.Buffer
	previousLogger := appLogger.Log
	appLogger.Log = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&output),
		zapcore.DebugLevel,
	))
	defer func() {
		appLogger.Log = previousLogger
	}()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/boom/:id", func(c *gin.Context) {
		Send(c, errors.New("postgres password=supersecret host=db.internal user=alice@example.com"))
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boom/123?token=secret-token", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	body := recorder.Body.String()
	logs := output.String()
	for _, sensitive := range []string{"supersecret", "alice@example.com", "db.internal", "secret-token", "password="} {
		if strings.Contains(body, sensitive) {
			t.Fatalf("response leaked %q: %s", sensitive, body)
		}
		if strings.Contains(logs, sensitive) {
			t.Fatalf("logs leaked %q: %s", sensitive, logs)
		}
	}
	if !strings.Contains(logs, "unhandled API error") {
		t.Fatalf("expected sanitized log entry, got %s", logs)
	}
}

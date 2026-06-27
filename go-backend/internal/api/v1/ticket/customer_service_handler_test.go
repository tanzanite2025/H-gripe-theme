package ticket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestReadVisitorSessionIDRequiresSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, Options{VisitorSecret: "test-secret"})
	sessionID := uuid.NewString()
	context := testContextWithVisitorCookie(sessionID)

	if _, ok := handler.readVisitorSessionID(context); ok {
		t.Fatal("expected unsigned visitor cookie to be rejected")
	}
}

func TestReadVisitorSessionIDAcceptsSignedCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, Options{VisitorSecret: "test-secret"})
	sessionID := uuid.NewString()
	signedCookie := handler.signVisitorSessionID(sessionID)
	context := testContextWithVisitorCookie(signedCookie)

	got, ok := handler.readVisitorSessionID(context)
	if !ok {
		t.Fatal("expected signed visitor cookie to be accepted")
	}
	if got != sessionID {
		t.Fatalf("expected session %q, got %q", sessionID, got)
	}
}

func TestReadVisitorSessionIDRejectsTamperedSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(nil, Options{VisitorSecret: "test-secret"})
	sessionID := uuid.NewString()
	signedCookie := handler.signVisitorSessionID(sessionID)
	context := testContextWithVisitorCookie(signedCookie + "tampered")

	if _, ok := handler.readVisitorSessionID(context); ok {
		t.Fatal("expected tampered visitor cookie to be rejected")
	}
}

func testContextWithVisitorCookie(value string) *gin.Context {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Cookie", customerServiceVisitorCookie+"="+strings.TrimSpace(value))
	context.Request = request
	return context
}

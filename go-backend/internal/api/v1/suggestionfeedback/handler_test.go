package suggestionfeedback

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateHoneypotSilentlyDropsBeforeServiceWork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(nil, nil)
	router := gin.New()
	router.POST("/suggestion-feedback", func(c *gin.Context) {
		c.Set("user_id", uint(7))
	}, h.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/suggestion-feedback",
		strings.NewReader(`{"message":"spam","company_tax_id":"123-456"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("Create() status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
}

func TestSuggestionUploadQuotaEnforcesCountAndBytesAndReleases(t *testing.T) {
	h := NewHandler(nil, nil)
	for i := 0; i < suggestionFeedbackUploadMaxPerUserPerDay; i++ {
		if !h.reserveSuggestionUpload(7, 1<<20) {
			t.Fatalf("reserve #%d unexpectedly rejected", i+1)
		}
	}
	if h.reserveSuggestionUpload(7, 1) {
		t.Fatal("reserve beyond daily file count unexpectedly succeeded")
	}
	h.releaseSuggestionUpload(7, 1<<20)
	remaining := suggestionFeedbackUploadBytesPerUserDay - 19*(1<<20)
	if !h.reserveSuggestionUpload(7, remaining) {
		t.Fatal("reserve after release should be governed by remaining byte quota")
	}
}

func TestSuggestionUploadQuotaRejectsOversizedSingleFile(t *testing.T) {
	h := NewHandler(nil, nil)
	if h.reserveSuggestionUpload(7, suggestionFeedbackUploadBytesPerUserDay+1) {
		t.Fatal("single file larger than daily quota unexpectedly succeeded")
	}
}

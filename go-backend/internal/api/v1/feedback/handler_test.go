package feedback

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainfeedback "commerce-platform/internal/domain/feedback"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateRejectsInvalidUserIDContextWithoutPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _ := newTestFeedbackHandler(t)
	router := gin.New()
	router.POST("/feedback", func(c *gin.Context) {
		c.Set("user_id", "not-a-uint")
	}, handler.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/feedback",
		strings.NewReader(`{"thread":"support-payment","content":"Works well."}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("Create() status = %d, want %d: %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}
}

func TestCreateRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, _ := newTestFeedbackHandler(t)
	router := gin.New()
	router.POST("/feedback", func(c *gin.Context) {
		c.Set("user_id", uint(7))
	}, handler.Create)

	body := `{"thread":"support-payment","content":"` + strings.Repeat("a", maxFeedbackCreateBodyBytes+1) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/feedback", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("Create() status = %d, want %d: %s", recorder.Code, http.StatusRequestEntityTooLarge, recorder.Body.String())
	}
}

func TestCreateStoresHMACSourceHashInsteadOfRawIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, db := newTestFeedbackHandler(t)
	handler.ConfigureSourceHashSecret("test-source-secret")
	router := gin.New()
	router.POST("/feedback", func(c *gin.Context) {
		c.Set("user_id", uint(7))
	}, handler.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/feedback",
		strings.NewReader(`{"thread":"support-payment","content":"Works well."}`),
	)
	request.RemoteAddr = "203.0.113.64:12345"
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("Create() status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var item domainfeedback.Feedback
	if err := db.First(&item).Error; err != nil {
		t.Fatalf("load feedback: %v", err)
	}
	if item.SourceHash == "" {
		t.Fatal("Create() did not store a source hash")
	}
	if item.SourceHash == "203.0.113.64" {
		t.Fatal("Create() stored the raw client IP")
	}
	if len(item.SourceHash) != 64 {
		t.Fatalf("source hash length = %d, want 64", len(item.SourceHash))
	}
}

func TestListPublicResponseDoesNotExposeUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, db := newTestFeedbackHandler(t)
	if err := db.Create(&domainfeedback.Feedback{
		ThreadKey: "support-payment",
		UserID:    42,
		Content:   "Approved feedback",
		Status:    "approved",
	}).Error; err != nil {
		t.Fatalf("seed feedback: %v", err)
	}

	router := gin.New()
	router.GET("/feedback", handler.List)

	request := httptest.NewRequest(http.MethodGet, "/feedback?thread=support-payment", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("List() status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "user_id") {
		t.Fatalf("public feedback response leaked user_id: %s", recorder.Body.String())
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Data) != 1 {
		t.Fatalf("response data length = %d, want 1", len(response.Data))
	}
}

func newTestFeedbackHandler(t *testing.T) (*Handler, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&domainfeedback.Feedback{}); err != nil {
		t.Fatalf("migrate feedback: %v", err)
	}

	feedbackService := service.NewFeedbackService(repository.NewFeedbackRepository(db))
	return NewHandler(feedbackService), db
}

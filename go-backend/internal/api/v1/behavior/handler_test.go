package behavior

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"commerce-platform/internal/domain/recommendation"
	attributionpkg "commerce-platform/internal/pkg/attribution"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestIngestBatchAcceptsAnonymousEvents(t *testing.T) {
	router, db := newBehaviorTestRouter(t, nil)
	body := map[string]any{
		"events": []map[string]any{
			{
				"event_id":     "event_handler_001",
				"event_type":   "product_view",
				"anonymous_id": "anon_handler",
				"session_id":   "session_handler",
				"locale":       "en-US",
				"path":         "/products/carbon-wheelset",
				"occurred_at":  time.Now().UTC(),
			},
		},
	}

	recorder := performBehaviorRequest(t, router, body)

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"accepted":1`)

	var stored recommendation.Event
	require.NoError(t, db.First(&stored, "event_id = ?", "event_handler_001").Error)
	require.Equal(t, "anon_handler", stored.AnonymousID)
	require.Nil(t, stored.UserID)
}

func TestIngestBatchUsesAuthenticatedUserWhenAnonymousIdentityIsMissing(t *testing.T) {
	userID := uint(17)
	router, db := newBehaviorTestRouter(t, func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	body := map[string]any{
		"events": []map[string]any{
			{
				"event_id":    "event_handler_002",
				"event_type":  "wishlist_add",
				"occurred_at": time.Now().UTC(),
			},
		},
	}

	recorder := performBehaviorRequest(t, router, body)

	require.Equal(t, http.StatusAccepted, recorder.Code)

	var stored recommendation.Event
	require.NoError(t, db.First(&stored, "event_id = ?", "event_handler_002").Error)
	require.NotNil(t, stored.UserID)
	require.Equal(t, userID, *stored.UserID)
}

func TestIngestBatchRejectsInvalidPayload(t *testing.T) {
	router, _ := newBehaviorTestRouter(t, nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/behavior-events/batch", bytes.NewBufferString(`{"events":[`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":"validation_error"`)
}

func TestIngestBatchReportsDuplicateEventIDs(t *testing.T) {
	router, _ := newBehaviorTestRouter(t, nil)
	event := map[string]any{
		"event_id":     "event_handler_003",
		"event_type":   "recommendation_click",
		"anonymous_id": "anon_handler",
		"session_id":   "session_handler",
		"occurred_at":  time.Now().UTC(),
	}
	body := map[string]any{
		"events": []map[string]any{event, event},
	}

	recorder := performBehaviorRequest(t, router, body)

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"accepted":1`)
	require.Contains(t, recorder.Body.String(), `"duplicates":1`)
}

func TestIngestBatchStoresSignedAttributionCookie(t *testing.T) {
	signer, err := attributionpkg.NewSigner("handler-attribution-secret")
	require.NoError(t, err)

	handler := NewHandler(service.NewBehaviorEventService(repository.NewRecommendationEventRepository(openBehaviorTestDB(t))))
	handler.ConfigureAttribution(signer, securecookie.Options{Secure: false, SameSite: http.SameSiteLaxMode})
	router := gin.New()
	router.POST("/behavior-events/batch", handler.IngestBatch)

	recorder := performBehaviorRequest(t, router, map[string]any{
		"events": []map[string]any{{
			"event_id":     "event_handler_attribution",
			"event_type":   "ad_landing",
			"anonymous_id": "anon_handler",
			"metadata": map[string]any{
				"utm_source":   "newsletter",
				"utm_campaign": "summer",
			},
			"occurred_at": time.Now().UTC(),
		}},
	})

	require.Equal(t, http.StatusAccepted, recorder.Code)
	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, attributionpkg.CookieName, cookies[0].Name)
	context, err := signer.Decode(cookies[0].Value)
	require.NoError(t, err)
	require.Equal(t, "newsletter", context.Source)
	require.Equal(t, "summer", context.Campaign)
}

func newBehaviorTestRouter(t *testing.T, middleware gin.HandlerFunc) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))

	eventService := service.NewBehaviorEventService(repository.NewRecommendationEventRepository(db))
	handler := NewHandler(eventService)
	router := gin.New()
	if middleware != nil {
		router.Use(middleware)
	}
	router.POST("/behavior-events/batch", handler.IngestBatch)

	return router, db
}

func openBehaviorTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&recommendation.Event{}))
	return db
}

func performBehaviorRequest(t *testing.T, router *gin.Engine, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/behavior-events/batch", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

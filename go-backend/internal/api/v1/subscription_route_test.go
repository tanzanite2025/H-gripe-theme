package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"tanzanite/internal/app"
	domainsubscription "tanzanite/internal/domain/subscription"
	"tanzanite/internal/domain/verification"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSubscriptionEmailLinksReachPublicRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	subscriptionService, emailSender := newSubscriptionRouteFixture(t)
	router := gin.New()
	RegisterRoutes(router, &app.Dependencies{
		Services: app.Services{
			Subscription: subscriptionService,
		},
	}, &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	})

	subscribeResponse := subscriptionRouteRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/subscriptions",
		`{"email":"rider@example.test","source":"website","locale":"en","tags":["newsletter"]}`,
	)
	require.Equal(t, http.StatusAccepted, subscribeResponse.Code, subscribeResponse.Body.String())

	subscription, err := subscriptionService.GetSubscription("rider@example.test")
	require.NoError(t, err)
	require.Equal(t, "pending", subscription.Status)

	confirmationLink := emailSender.LastLink(t)
	require.Equal(t, "https", confirmationLink.Scheme)
	require.Equal(t, "storefront.example.test", confirmationLink.Host)
	require.True(t, strings.HasPrefix(confirmationLink.Path, "/api/v1/subscriptions/confirm/"))

	confirmationResponse := subscriptionLinkRequest(t, router, confirmationLink)
	require.Equal(t, http.StatusOK, confirmationResponse.Code, confirmationResponse.Body.String())

	subscription, err = subscriptionService.GetSubscription("rider@example.test")
	require.NoError(t, err)
	require.Equal(t, "active", subscription.Status)

	replayResponse := subscriptionLinkRequest(t, router, confirmationLink)
	require.Equal(t, http.StatusBadRequest, replayResponse.Code, replayResponse.Body.String())

	unsubscribeResponse := subscriptionRouteRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/subscriptions/unsubscribe",
		`{"email":"rider@example.test"}`,
	)
	require.Equal(t, http.StatusAccepted, unsubscribeResponse.Code, unsubscribeResponse.Body.String())

	unsubscribeLink := emailSender.LastLink(t)
	require.True(t, strings.HasPrefix(unsubscribeLink.Path, "/api/v1/subscriptions/unsubscribe/"))
	unsubscribeResult := subscriptionLinkRequest(t, router, unsubscribeLink)
	require.Equal(t, http.StatusOK, unsubscribeResult.Code, unsubscribeResult.Body.String())

	subscription, err = subscriptionService.GetSubscription("rider@example.test")
	require.NoError(t, err)
	require.Equal(t, "unsubscribed", subscription.Status)

	resubscribeResponse := subscriptionRouteRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/subscriptions/resubscribe",
		`{"email":"rider@example.test"}`,
	)
	require.Equal(t, http.StatusAccepted, resubscribeResponse.Code, resubscribeResponse.Body.String())

	resubscribeLink := emailSender.LastLink(t)
	require.True(t, strings.HasPrefix(resubscribeLink.Path, "/api/v1/subscriptions/resubscribe/"))
	resubscribeResult := subscriptionLinkRequest(t, router, resubscribeLink)
	require.Equal(t, http.StatusOK, resubscribeResult.Code, resubscribeResult.Body.String())

	statusRequest := subscriptionRouteRequest(
		t,
		router,
		http.MethodGet,
		"/api/v1/subscriptions/status/rider@example.test",
		"",
	)
	require.Equal(t, http.StatusAccepted, statusRequest.Code, statusRequest.Body.String())

	statusLink := emailSender.LastLink(t)
	require.True(t, strings.HasPrefix(statusLink.Path, "/api/v1/subscriptions/status-token/"))
	statusResult := subscriptionLinkRequest(t, router, statusLink)
	require.Equal(t, http.StatusOK, statusResult.Code, statusResult.Body.String())
	require.Contains(t, statusResult.Body.String(), `"status":"active"`)
	require.NotContains(t, statusResult.Body.String(), "token")

	subscription, err = subscriptionService.GetSubscription("rider@example.test")
	require.NoError(t, err)
	require.Equal(t, "active", subscription.Status)

}

type recordingSubscriptionEmailSender struct {
	bodies []string
}

func (s *recordingSubscriptionEmailSender) SendEmail(_ []string, _ string, body string) error {
	s.bodies = append(s.bodies, body)
	return nil
}

func (s *recordingSubscriptionEmailSender) LastLink(t *testing.T) *url.URL {
	t.Helper()
	require.NotEmpty(t, s.bodies)

	body := s.bodies[len(s.bodies)-1]
	for _, line := range strings.Split(body, "\n") {
		link := strings.TrimSpace(line)
		if !strings.HasPrefix(link, "https://") && !strings.HasPrefix(link, "http://") {
			continue
		}

		parsed, err := url.Parse(link)
		require.NoError(t, err)
		return parsed
	}

	t.Fatalf("email body did not contain an action link: %q", body)
	return nil
}

func newSubscriptionRouteFixture(t *testing.T) (*service.SubscriptionService, *recordingSubscriptionEmailSender) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&domainsubscription.Subscription{},
		&verification.EmailChallenge{},
	))

	emailSender := &recordingSubscriptionEmailSender{}
	subscriptionService := service.NewSubscriptionService(repository.NewSubscriptionRepository(db))
	subscriptionService.ConfigureEmailChallenges(
		repository.NewEmailChallengeRepository(db),
		"test-email-secret",
		emailSender,
	)
	subscriptionService.ConfigureEmailBaseURL("https://storefront.example.test")

	return subscriptionService, emailSender
}

func subscriptionRouteRequest(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func subscriptionLinkRequest(t *testing.T, router *gin.Engine, link *url.URL) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, link.RequestURI(), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

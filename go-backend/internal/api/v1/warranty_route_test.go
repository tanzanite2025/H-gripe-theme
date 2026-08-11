package v1

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"commerce-platform/internal/app"
	domainorder "commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/registration"
	"commerce-platform/internal/domain/verification"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWarrantyRoutesCompleteEmailVerificationAndClaimSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, registrationService, emailSender := newWarrantyRouteFixture(t)
	router := gin.New()
	RegisterRoutes(router, &app.Dependencies{
		Services: app.Services{
			Registration: registrationService,
		},
	}, &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-email-secret"},
	})

	order := domainorder.Order{
		OrderNumber:   "TZ-WARRANTY-ROUTE",
		UserID:        7,
		Status:        "paid",
		PaymentStatus: "paid",
		TotalAmount:   199,
		Currency:      "USD",
	}
	order.ShippingAddress.Email = "rider@example.test"
	require.NoError(t, db.Create(&order).Error)

	verifyResponse := warrantyJSONRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/registrations/warranty/verify-order",
		`{"order_number":"TZ-WARRANTY-ROUTE","email":"rider@example.test"}`,
	)
	require.Equal(t, http.StatusAccepted, verifyResponse.Code, verifyResponse.Body.String())
	require.Contains(t, verifyResponse.Body.String(), "If the order can be verified")
	require.Len(t, emailSender.bodies, 1)

	verificationURL := emailSender.LastLink(t)
	require.Equal(t, "https", verificationURL.Scheme)
	require.Equal(t, "storefront.example.test", verificationURL.Host)
	require.Equal(t, "/support/warranty", verificationURL.Path)
	verificationToken := verificationURL.Query().Get("verification_token")
	require.NotEmpty(t, verificationToken)

	verifyTokenResponse := warrantyLinkRequest(
		t,
		router,
		"/api/v1/registrations/warranty/verify/"+url.PathEscape(verificationToken),
	)
	require.Equal(t, http.StatusOK, verifyTokenResponse.Code, verifyTokenResponse.Body.String())
	require.Contains(t, verifyTokenResponse.Body.String(), `"verified":true`)

	claimResponse := warrantyClaimRequest(t, router, map[string]string{
		"order_number":       order.OrderNumber,
		"email":              order.ShippingAddress.Email,
		"verification_token": verificationToken,
		"issue_description":  "rim issue",
		"tire_pressure":      "45 PSI",
		"is_tubeless":        "yes",
	})
	require.Equal(t, http.StatusCreated, claimResponse.Code, claimResponse.Body.String())
	require.Contains(t, claimResponse.Body.String(), `"success":true`)

	var claims []registration.WarrantyClaim
	require.NoError(t, db.Find(&claims).Error)
	require.Len(t, claims, 1)
	require.Equal(t, order.OrderNumber, claims[0].OrderNumber)
	require.Equal(t, order.ShippingAddress.Email, claims[0].Email)
	require.Equal(t, "rim issue", claims[0].Description)
	require.Equal(t, "45 PSI", claims[0].TirePressure)
	require.True(t, claims[0].IsTubeless)
	require.Equal(t, "submitted", claims[0].Status)
	require.Equal(t, order.UserID, claims[0].UserID)

	replayResponse := warrantyClaimRequest(t, router, map[string]string{
		"order_number":       order.OrderNumber,
		"email":              order.ShippingAddress.Email,
		"verification_token": verificationToken,
		"issue_description":  "replay",
	})
	require.Equal(t, http.StatusUnauthorized, replayResponse.Code, replayResponse.Body.String())

	var claimCount int64
	require.NoError(t, db.Model(&registration.WarrantyClaim{}).Count(&claimCount).Error)
	require.Equal(t, int64(1), claimCount)
}

func TestWarrantyOrderVerificationDoesNotEnumerateUnknownOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, registrationService, emailSender := newWarrantyRouteFixture(t)
	router := gin.New()
	RegisterRoutes(router, &app.Dependencies{
		Services: app.Services{
			Registration: registrationService,
		},
	}, &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-email-secret"},
	})

	response := warrantyJSONRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/registrations/warranty/verify-order",
		`{"order_number":"TZ-WARRANTY-MISSING","email":"rider@example.test"}`,
	)
	require.Equal(t, http.StatusAccepted, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "If the order can be verified")
	require.Empty(t, emailSender.bodies)
}

type recordingWarrantyEmailSender struct {
	bodies []string
}

func (s *recordingWarrantyEmailSender) SendEmail(_ []string, _ string, body string) error {
	s.bodies = append(s.bodies, body)
	return nil
}

func (s *recordingWarrantyEmailSender) LastLink(t *testing.T) *url.URL {
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

func newWarrantyRouteFixture(t *testing.T) (*gorm.DB, *service.RegistrationService, *recordingWarrantyEmailSender) {
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
		&domainorder.Order{},
		&domainorder.OrderItem{},
		&registration.WarrantyClaim{},
		&verification.EmailChallenge{},
	))

	emailSender := &recordingWarrantyEmailSender{}
	registrationService := service.NewRegistrationService(
		repository.NewRegistrationRepository(db),
		nil,
		repository.NewOrderRepository(db),
	)
	registrationService.ConfigureEmailChallenges(
		repository.NewEmailChallengeRepository(db),
		"test-email-secret",
		emailSender,
	)
	registrationService.ConfigureEmailBaseURL("https://storefront.example.test")

	return db, registrationService, emailSender
}

func warrantyJSONRequest(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func warrantyLinkRequest(t *testing.T, router *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func warrantyClaimRequest(t *testing.T, router *gin.Engine, fields map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		require.NoError(t, writer.WriteField(key, value))
	}
	require.NoError(t, writer.Close())

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/registrations/warranty/claim", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

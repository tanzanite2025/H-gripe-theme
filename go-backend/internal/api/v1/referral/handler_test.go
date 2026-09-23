package referral

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubReferralService struct {
	createAttributionToken func(code, source string) (string, int, error)
	validateCode           func(code string) (*service.ReferralValidation, error)
	bindFromToken          func(userID uint, token, clientIP string) (*loyalty.ReferralRecord, error)
}

func (s *stubReferralService) CreateAttributionToken(code, source string) (string, int, error) {
	return s.createAttributionToken(code, source)
}

func (s *stubReferralService) ValidateCode(code string) (*service.ReferralValidation, error) {
	return s.validateCode(code)
}

func (s *stubReferralService) Dashboard(userID uint) (*service.ReferralDashboard, error) {
	return nil, errors.New("unexpected Dashboard call")
}

func (s *stubReferralService) History(userID uint, page, pageSize int) (*service.ReferralHistory, error) {
	return nil, errors.New("unexpected History call")
}

func (s *stubReferralService) BindFromToken(userID uint, token, clientIP string) (*loyalty.ReferralRecord, error) {
	if s.bindFromToken != nil {
		return s.bindFromToken(userID, token, clientIP)
	}
	return nil, errors.New("unexpected BindFromToken call")
}

func TestCaptureSetsSignedAttributionCookieAndKeepsLocalRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&stubReferralService{
		createAttributionToken: func(code, source string) (string, int, error) {
			assert.Equal(t, "RACE2345", code)
			assert.Equal(t, "link", source)
			return "signed.payload", 30 * 24 * 60 * 60, nil
		},
		validateCode: func(string) (*service.ReferralValidation, error) { return nil, nil },
	}, securecookie.Options{Secure: true, SameSite: http.SameSiteStrictMode, Domain: "shop.example.test"}, "https://shop.example.test")
	router := gin.New()
	router.GET("/r/:code", handler.Capture)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/r/RACE2345?next=%2Fproducts%3Fsort%3Dnew", nil)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusFound, response.Code)
	assert.Equal(t, "https://shop.example.test/products?sort=new", response.Header().Get("Location"))
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	assert.Equal(t, "storefront_referral", cookie.Name)
	assert.Equal(t, "signed.payload", cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, "shop.example.test", cookie.Domain)
	assert.Equal(t, 30*24*60*60, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
}

func TestCaptureRejectsExternalRedirectTargets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&stubReferralService{
		createAttributionToken: func(code, source string) (string, int, error) {
			return "signed.payload", 60, nil
		},
		validateCode: func(string) (*service.ReferralValidation, error) { return nil, nil },
	}, securecookie.DefaultOptions())
	router := gin.New()
	router.GET("/r/:code", handler.Capture)

	for _, next := range []string{"//evil.example/path", "https://evil.example/path", "javascript:alert(1)"} {
		t.Run(next, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/r/RACE2345?next="+next, nil)
			router.ServeHTTP(response, request)
			assert.Equal(t, http.StatusFound, response.Code)
			assert.Equal(t, "/", response.Header().Get("Location"))
		})
	}
}

func TestValidateMapsUnavailableCodesToNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, serviceErr := range []error{service.ErrReferralProgramDisabled, service.ErrReferralCodeNotFound} {
		handler := NewHandler(&stubReferralService{
			createAttributionToken: func(string, string) (string, int, error) { return "", 0, nil },
			validateCode:           func(string) (*service.ReferralValidation, error) { return nil, serviceErr },
		}, securecookie.DefaultOptions())
		router := gin.New()
		router.POST("/validate", handler.Validate)
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"referral_code":"BADCODE"}`))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusNotFound, response.Code)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
		assert.Equal(t, "referral_code_not_found", payload["code"])
	}
}

func TestCustomerReferralHandlersRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&stubReferralService{
		createAttributionToken: func(string, string) (string, int, error) { return "", 0, nil },
		validateCode:           func(string) (*service.ReferralValidation, error) { return nil, nil },
	}, securecookie.DefaultOptions())
	router := gin.New()
	router.GET("/me", handler.Me)
	router.GET("/history", handler.History)
	router.POST("/bind", handler.Bind)

	for _, route := range []struct {
		method string
		path   string
	}{{http.MethodGet, "/me"}, {http.MethodGet, "/history"}, {http.MethodPost, "/bind"}} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(route.method, route.path, nil)
		router.ServeHTTP(response, request)
		assert.Equal(t, http.StatusUnauthorized, response.Code, route.path)
	}
}

func TestBindClearsTerminalAttributionCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&stubReferralService{
		createAttributionToken: func(string, string) (string, int, error) { return "", 0, nil },
		validateCode:           func(string) (*service.ReferralValidation, error) { return nil, nil },
		bindFromToken: func(userID uint, token, clientIP string) (*loyalty.ReferralRecord, error) {
			assert.Equal(t, uint(42), userID)
			assert.Equal(t, "invalid-signed-token", token)
			return nil, service.ErrRefereeNotEligible
		},
	}, securecookie.DefaultOptions())
	router := gin.New()
	router.POST("/bind", func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Next()
	}, handler.Bind)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/bind", nil)
	request.AddCookie(&http.Cookie{Name: "storefront_referral", Value: "invalid-signed-token"})
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "storefront_referral", cookies[0].Name)
	assert.Less(t, cookies[0].MaxAge, 0)
	assert.True(t, cookies[0].HttpOnly)
}

func TestBindAcceptsExplicitReferralCodeWithoutAttributionCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&stubReferralService{
		createAttributionToken: func(code, source string) (string, int, error) {
			assert.Equal(t, "FRIEND123", code)
			assert.Equal(t, "manual_input", source)
			return "signed.manual", 3600, nil
		},
		validateCode: func(string) (*service.ReferralValidation, error) { return nil, nil },
		bindFromToken: func(userID uint, token, clientIP string) (*loyalty.ReferralRecord, error) {
			assert.Equal(t, uint(42), userID)
			assert.Equal(t, "signed.manual", token)
			return &loyalty.ReferralRecord{ID: 7, Status: loyalty.ReferralStatusPending}, nil
		},
	}, securecookie.DefaultOptions())
	router := gin.New()
	router.POST("/bind", func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Next()
	}, handler.Bind)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(`{"referral_code":"FRIEND123"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), `"referral_id":7`)
}

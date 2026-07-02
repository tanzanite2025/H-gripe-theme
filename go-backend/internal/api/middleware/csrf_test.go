package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"tanzanite/internal/pkg/securecookie"

	"github.com/gin-gonic/gin"
)

func TestCSRFProtectionBlocksCookieWriteWithoutOriginOrToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFProtection([]string{"https://app.example.com"}))
	router.POST("/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/orders", nil)
	request.AddCookie(&http.Cookie{Name: securecookie.AuthTokenCookie, Value: "jwt"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestCSRFProtectionBlocksTrustedOriginWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFProtection([]string{"https://app.example.com"}))
	router.POST("/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/orders", nil)
	request.Host = "api.example.com"
	request.Header.Set("Origin", "https://app.example.com")
	request.AddCookie(&http.Cookie{Name: securecookie.AuthTokenCookie, Value: "jwt"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestCSRFProtectionAllowsTrustedOriginWithDoubleSubmitToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFProtection([]string{"https://app.example.com"}))
	router.POST("/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/orders", nil)
	request.Host = "api.example.com"
	request.Header.Set("Origin", "https://app.example.com")
	request.Header.Set(securecookie.CSRFTokenHeader, "csrf-token")
	request.AddCookie(&http.Cookie{Name: securecookie.AuthTokenCookie, Value: "jwt"})
	request.AddCookie(&http.Cookie{Name: securecookie.CSRFTokenCookie, Value: "csrf-token"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestCSRFProtectionAllowsDoubleSubmitToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFProtection(nil))
	router.POST("/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/orders", nil)
	request.AddCookie(&http.Cookie{Name: securecookie.AuthTokenCookie, Value: "jwt"})
	request.AddCookie(&http.Cookie{Name: securecookie.CSRFTokenCookie, Value: "csrf-token"})
	request.Header.Set(securecookie.CSRFTokenHeader, "csrf-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestCSRFProtectionAllowsUnsafeRequestWithoutAuthCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFProtection(nil))
	router.POST("/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/orders", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestCSRFProtectionBlocksCookieRequestWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFProtection([]string{"https://app.example.com"}))
	router.POST("/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/orders", nil)
	request.Header.Set("Origin", "https://app.example.com")
	request.AddCookie(&http.Cookie{Name: securecookie.AuthTokenCookie, Value: "jwt"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestCSRFProtectionBlocksUntrustedOriginEvenWithToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CSRFProtection([]string{"https://app.example.com"}))
	router.POST("/orders", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/orders", nil)
	request.Header.Set("Origin", "https://evil.example.com")
	request.Header.Set(securecookie.CSRFTokenHeader, "csrf-token")
	request.AddCookie(&http.Cookie{Name: securecookie.AuthTokenCookie, Value: "jwt"})
	request.AddCookie(&http.Cookie{Name: securecookie.CSRFTokenCookie, Value: "csrf-token"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

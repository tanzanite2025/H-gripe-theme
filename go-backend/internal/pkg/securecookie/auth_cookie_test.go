package securecookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetAuthTokenUsesSecureHttpOnlySameSiteLax(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	SetAuthToken(context, "token", 3600)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != AuthTokenCookie {
		t.Fatalf("cookie name = %s, want %s", cookie.Name, AuthTokenCookie)
	}
	if !cookie.Secure {
		t.Fatal("auth cookie must be Secure")
	}
	if !cookie.HttpOnly {
		t.Fatal("auth cookie must be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
}

func TestSetCSRFTokenUsesReadableSameSiteLaxCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	token, err := SetCSRFToken(context, 3600)
	if err != nil {
		t.Fatalf("SetCSRFToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("CSRF token must not be empty")
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != CSRFTokenCookie {
		t.Fatalf("cookie name = %s, want %s", cookie.Name, CSRFTokenCookie)
	}
	if !cookie.Secure {
		t.Fatal("CSRF cookie must be Secure")
	}
	if cookie.HttpOnly {
		t.Fatal("CSRF cookie must be readable by frontend JS for double-submit header")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
	}
}

func TestSetAuthTokenAcceptsConfiguredCookieOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	SetAuthToken(context, "token", 3600, Options{
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Domain:   "example.com",
	})

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Secure {
		t.Fatal("auth cookie should honor configured Secure=false")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("SameSite = %v, want %v", cookie.SameSite, http.SameSiteStrictMode)
	}
	if cookie.Domain != "example.com" {
		t.Fatalf("Domain = %q, want %q", cookie.Domain, "example.com")
	}
}

func TestCookieScopesUseDistinctNamesAndPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storefrontRecorder := httptest.NewRecorder()
	storefrontContext, _ := gin.CreateTestContext(storefrontRecorder)
	storefrontOptions := Options{
		Secure:   false,
		Path:     "/api/v1",
		CSRFPath: "/",
		Names:    StorefrontCookieNames(),
	}
	SetAuthToken(storefrontContext, "storefront", 3600, storefrontOptions)
	if cookie := storefrontRecorder.Result().Cookies()[0]; cookie.Name != StorefrontAuthTokenCookie || cookie.Path != "/api/v1" {
		t.Fatalf("storefront auth cookie = (%q, %q), want (%q, %q)", cookie.Name, cookie.Path, StorefrontAuthTokenCookie, "/api/v1")
	}

	adminRecorder := httptest.NewRecorder()
	adminContext, _ := gin.CreateTestContext(adminRecorder)
	adminOptions := Options{
		Secure:   false,
		Path:     "/api/admin",
		CSRFPath: "/",
		Names:    AdminCookieNames(),
	}
	SetCSRFToken(adminContext, 3600, adminOptions)
	if cookie := adminRecorder.Result().Cookies()[0]; cookie.Name != AdminCSRFTokenCookie || cookie.Path != "/" {
		t.Fatalf("admin CSRF cookie = (%q, %q), want (%q, %q)", cookie.Name, cookie.Path, AdminCSRFTokenCookie, "/")
	}

	if StorefrontAuthTokenCookie == AdminAuthTokenCookie || StorefrontRefreshTokenCookie == AdminRefreshTokenCookie || StorefrontCSRFTokenCookie == AdminCSRFTokenCookie {
		t.Fatal("storefront and admin cookie names must be distinct")
	}
}

func TestScopedDefaults(t *testing.T) {
	storefront := StorefrontOptions()
	if storefront.Path != "/api/v1" || storefront.CSRFPath != "/" || storefront.Names != StorefrontCookieNames() {
		t.Fatalf("unexpected storefront defaults: %+v", storefront)
	}

	admin := AdminOptions()
	if admin.Path != "/api/admin" || admin.CSRFPath != "/" || admin.Names != AdminCookieNames() {
		t.Fatalf("unexpected admin defaults: %+v", admin)
	}
}

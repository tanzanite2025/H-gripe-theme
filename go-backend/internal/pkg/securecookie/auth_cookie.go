package securecookie

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// Storefront cookies are deliberately namespaced so that a storefront
	// login can never overwrite a backoffice session on the same host.
	StorefrontAuthTokenCookie    = "storefront_auth_token"
	StorefrontRefreshTokenCookie = "storefront_refresh_token"
	StorefrontCSRFTokenCookie    = "storefront_csrf_token"

	AdminAuthTokenCookie    = "admin_auth_token"
	AdminRefreshTokenCookie = "admin_refresh_token"
	AdminCSRFTokenCookie    = "admin_csrf_token"

	// These aliases keep the default (storefront) API surface source-compatible.
	AuthTokenCookie    = StorefrontAuthTokenCookie
	RefreshTokenCookie = StorefrontRefreshTokenCookie
	CSRFTokenCookie    = StorefrontCSRFTokenCookie
	CSRFTokenHeader    = "X-CSRF-Token"
)

const csrfTokenBytes = 32

type Options struct {
	Secure   bool
	SameSite http.SameSite
	Domain   string
	// Path scopes HttpOnly auth and refresh cookies to their API namespace.
	Path string
	// CSRFPath is separate because browser JavaScript needs to read the CSRF
	// cookie from storefront and backoffice pages outside the API path.
	CSRFPath string
	Names    CookieNames
}

type CookieNames struct {
	AuthToken    string
	RefreshToken string
	CSRFToken    string
}

func StorefrontCookieNames() CookieNames {
	return CookieNames{
		AuthToken:    StorefrontAuthTokenCookie,
		RefreshToken: StorefrontRefreshTokenCookie,
		CSRFToken:    StorefrontCSRFTokenCookie,
	}
}

func AdminCookieNames() CookieNames {
	return CookieNames{
		AuthToken:    AdminAuthTokenCookie,
		RefreshToken: AdminRefreshTokenCookie,
		CSRFToken:    AdminCSRFTokenCookie,
	}
}

func DefaultOptions() Options {
	return Options{
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		CSRFPath: "/",
		Names:    StorefrontCookieNames(),
	}
}

func StorefrontOptions() Options {
	options := DefaultOptions()
	options.Path = "/api/v1"
	options.CSRFPath = "/"
	options.Names = StorefrontCookieNames()
	return options
}

func AdminOptions() Options {
	options := DefaultOptions()
	options.Path = "/api/admin"
	options.CSRFPath = "/"
	options.Names = AdminCookieNames()
	return options
}

func SetAuthToken(c *gin.Context, token string, maxAge int, options ...Options) {
	resolved := resolveOptions(options)
	setCookie(c, resolved.Names.AuthToken, token, maxAge, true, resolved.Path, resolved)
}

func SetRefreshToken(c *gin.Context, token string, maxAge int, options ...Options) {
	resolved := resolveOptions(options)
	setCookie(c, resolved.Names.RefreshToken, token, maxAge, true, resolved.Path, resolved)
}

func SetCSRFToken(c *gin.Context, maxAge int, options ...Options) (string, error) {
	token, err := NewCSRFToken()
	if err != nil {
		return "", err
	}
	resolved := resolveOptions(options)
	setCookie(c, resolved.Names.CSRFToken, token, maxAge, false, resolved.CSRFPath, resolved)
	return token, nil
}

func ClearAuthToken(c *gin.Context, options ...Options) {
	resolved := resolveOptions(options)
	setCookie(c, resolved.Names.AuthToken, "", -1, true, resolved.Path, resolved)
}

func ClearRefreshToken(c *gin.Context, options ...Options) {
	resolved := resolveOptions(options)
	setCookie(c, resolved.Names.RefreshToken, "", -1, true, resolved.Path, resolved)
}

func ClearCSRFToken(c *gin.Context, options ...Options) {
	resolved := resolveOptions(options)
	setCookie(c, resolved.Names.CSRFToken, "", -1, false, resolved.CSRFPath, resolved)
}

func NewCSRFToken() (string, error) {
	token := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}

func resolveOptions(options []Options) Options {
	if len(options) == 0 {
		return DefaultOptions()
	}

	resolved := options[0]
	if resolved.SameSite == 0 {
		resolved.SameSite = http.SameSiteLaxMode
	}
	if resolved.Path == "" {
		resolved.Path = "/"
	}
	if resolved.CSRFPath == "" {
		resolved.CSRFPath = "/"
	}
	defaults := StorefrontCookieNames()
	if resolved.Names.AuthToken == "" {
		resolved.Names.AuthToken = defaults.AuthToken
	}
	if resolved.Names.RefreshToken == "" {
		resolved.Names.RefreshToken = defaults.RefreshToken
	}
	if resolved.Names.CSRFToken == "" {
		resolved.Names.CSRFToken = defaults.CSRFToken
	}
	return resolved
}

// NormalizeOptions fills security, path, and storefront-name defaults for
// handlers that retain cookie options between requests.
func NormalizeOptions(options Options) Options {
	return resolveOptions([]Options{options})
}

func setCookie(c *gin.Context, name, value string, maxAge int, httpOnly bool, path string, options Options) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   options.Domain,
		MaxAge:   maxAge,
		Secure:   options.Secure,
		HttpOnly: httpOnly,
		SameSite: options.SameSite,
	})
}

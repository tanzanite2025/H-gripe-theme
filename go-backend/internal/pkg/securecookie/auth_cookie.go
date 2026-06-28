package securecookie

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	AuthTokenCookie    = "auth_token"
	RefreshTokenCookie = "refresh_token"
	CSRFTokenCookie    = "csrf_token"
	CSRFTokenHeader    = "X-CSRF-Token"
)

const csrfTokenBytes = 32

type Options struct {
	Secure   bool
	SameSite http.SameSite
	Domain   string
}

func DefaultOptions() Options {
	return Options{
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func SetAuthToken(c *gin.Context, token string, maxAge int, options ...Options) {
	setCookie(c, AuthTokenCookie, token, maxAge, true, resolveOptions(options))
}

func SetRefreshToken(c *gin.Context, token string, maxAge int, options ...Options) {
	setCookie(c, RefreshTokenCookie, token, maxAge, true, resolveOptions(options))
}

func SetCSRFToken(c *gin.Context, maxAge int, options ...Options) (string, error) {
	token, err := NewCSRFToken()
	if err != nil {
		return "", err
	}
	setCookie(c, CSRFTokenCookie, token, maxAge, false, resolveOptions(options))
	return token, nil
}

func ClearAuthToken(c *gin.Context, options ...Options) {
	setCookie(c, AuthTokenCookie, "", -1, true, resolveOptions(options))
}

func ClearRefreshToken(c *gin.Context, options ...Options) {
	setCookie(c, RefreshTokenCookie, "", -1, true, resolveOptions(options))
}

func ClearCSRFToken(c *gin.Context, options ...Options) {
	setCookie(c, CSRFTokenCookie, "", -1, false, resolveOptions(options))
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
	return resolved
}

func setCookie(c *gin.Context, name, value string, maxAge int, httpOnly bool, options Options) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   options.Domain,
		MaxAge:   maxAge,
		Secure:   options.Secure,
		HttpOnly: httpOnly,
		SameSite: options.SameSite,
	})
}

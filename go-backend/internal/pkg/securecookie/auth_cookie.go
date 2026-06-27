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

func SetAuthToken(c *gin.Context, token string, maxAge int) {
	setCookie(c, AuthTokenCookie, token, maxAge, true, true)
}

func SetRefreshToken(c *gin.Context, token string, maxAge int) {
	setCookie(c, RefreshTokenCookie, token, maxAge, true, true)
}

func SetCSRFToken(c *gin.Context, maxAge int) (string, error) {
	token, err := NewCSRFToken()
	if err != nil {
		return "", err
	}
	setCookie(c, CSRFTokenCookie, token, maxAge, true, false)
	return token, nil
}

func ClearAuthToken(c *gin.Context) {
	setCookie(c, AuthTokenCookie, "", -1, true, true)
}

func ClearRefreshToken(c *gin.Context) {
	setCookie(c, RefreshTokenCookie, "", -1, true, true)
}

func ClearCSRFToken(c *gin.Context) {
	setCookie(c, CSRFTokenCookie, "", -1, true, false)
}

func NewCSRFToken() (string, error) {
	token := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}

func setCookie(c *gin.Context, name, value string, maxAge int, secure, httpOnly bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}

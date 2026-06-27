package middleware

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"tanzanite/internal/pkg/securecookie"

	"github.com/gin-gonic/gin"
)

func CSRFProtection(allowedOrigins []string) gin.HandlerFunc {
	trustedOrigins := buildTrustedOriginSet(allowedOrigins)

	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin != "" {
			if !isTrustedOrigin(origin, c.Request, trustedOrigins) {
				abortCSRF(c)
				return
			}
		}

		referer := strings.TrimSpace(c.GetHeader("Referer"))
		if origin == "" {
			if referer != "" {
				if !isTrustedReferer(referer, c.Request, trustedOrigins) {
					abortCSRF(c)
					return
				}
			} else if strings.EqualFold(c.GetHeader("Sec-Fetch-Site"), "cross-site") {
				abortCSRF(c)
				return
			}
		}

		if requestHasAuthCookie(c.Request) && !hasValidCSRFToken(c) {
			abortCSRF(c)
			return
		}

		c.Next()
	}
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func requestHasAuthCookie(r *http.Request) bool {
	for _, name := range []string{securecookie.AuthTokenCookie, securecookie.RefreshTokenCookie} {
		if cookie, err := r.Cookie(name); err == nil && cookie.Value != "" {
			return true
		}
	}
	return false
}

func hasValidCSRFToken(c *gin.Context) bool {
	headerToken := strings.TrimSpace(c.GetHeader(securecookie.CSRFTokenHeader))
	if headerToken == "" {
		return false
	}

	cookie, err := c.Cookie(securecookie.CSRFTokenCookie)
	if err != nil || cookie == "" {
		return false
	}

	if len(headerToken) != len(cookie) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookie)) == 1
}

func buildTrustedOriginSet(origins []string) map[string]struct{} {
	trusted := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" || origin == "*" {
			continue
		}
		if parsed, err := url.Parse(origin); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			trusted[canonicalOrigin(parsed)] = struct{}{}
		}
	}
	return trusted
}

func isTrustedOrigin(origin string, r *http.Request, trustedOrigins map[string]struct{}) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	if sameRequestOrigin(parsed, r) {
		return true
	}
	_, ok := trustedOrigins[canonicalOrigin(parsed)]
	return ok
}

func isTrustedReferer(referer string, r *http.Request, trustedOrigins map[string]struct{}) bool {
	parsed, err := url.Parse(referer)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	if sameRequestOrigin(parsed, r) {
		return true
	}
	_, ok := trustedOrigins[canonicalOrigin(parsed)]
	return ok
}

func sameRequestOrigin(originURL *url.URL, r *http.Request) bool {
	return strings.EqualFold(canonicalOrigin(originURL), requestOrigin(r))
}

func canonicalOrigin(u *url.URL) string {
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
}

func requestOrigin(r *http.Request) string {
	scheme := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")))
	if scheme == "" {
		if strings.EqualFold(r.Header.Get("X-Forwarded-Ssl"), "on") {
			scheme = "https"
		} else if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	if commaIndex := strings.Index(scheme, ","); commaIndex >= 0 {
		scheme = strings.TrimSpace(scheme[:commaIndex])
	}

	return strings.ToLower(scheme) + "://" + strings.ToLower(r.Host)
}

func abortCSRF(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": "csrf validation failed",
		"code":  "CSRF_VALIDATION_FAILED",
	})
}

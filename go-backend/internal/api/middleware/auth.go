package middleware

import (
	"net/http"
	"strings"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *service.AuthService, cookieScopes ...securecookie.CookieNames) gin.HandlerFunc {
	return func(c *gin.Context) {
		names := resolveCookieNames(c, cookieScopes)
		tokenString, err := c.Cookie(names.AuthToken)
		if err != nil || tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication cookie required"})
			c.Abort()
			return
		}

		claims, err := authService.ValidateActiveToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		setAuthClaims(c, claims)
		c.Next()
	}
}

func OptionalAuthMiddleware(authService *service.AuthService, cookieScopes ...securecookie.CookieNames) gin.HandlerFunc {
	return func(c *gin.Context) {
		names := resolveCookieNames(c, cookieScopes)
		tokenString, err := c.Cookie(names.AuthToken)
		if err != nil || tokenString == "" {
			c.Next()
			return
		}

		claims, err := authService.ValidateActiveToken(tokenString)
		if err == nil {
			setAuthClaims(c, claims)
		}

		c.Next()
	}
}

func resolveCookieNames(c *gin.Context, cookieScopes []securecookie.CookieNames) securecookie.CookieNames {
	if len(cookieScopes) > 0 {
		return cookieScopes[0]
	}
	if c != nil && strings.HasPrefix(c.Request.URL.Path, "/api/admin/") {
		return securecookie.AdminCookieNames()
	}
	return securecookie.StorefrontCookieNames()
}

func setAuthClaims(c *gin.Context, claims *service.Claims) {
	c.Set("user_id", claims.UserID)
	c.Set("email", claims.Email)
	c.Set("username", claims.Username)
	c.Set("role", claims.Role)
	c.Set("user_role", claims.Role)
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			c.Abort()
			return
		}

		roleStr := string(auth.NormalizeRole(userRole.(string)))
		allowed := false
		for _, role := range roles {
			if string(auth.NormalizeRole(role)) == roleStr {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

package admin

import (
	"fmt"
	"testing"

	"commerce-platform/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

func TestAdminRouteRegistrationHasNoDuplicateRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	admin := router.Group("/api/admin")
	admin.Use(middleware.CSRFProtection(nil))

	assertNoPanic(t, func() {
		registerPublicAdminRoutes(admin, nil, nil, nil, nil)

		authenticated := admin.Group("")
		authenticated.Use(middleware.AuthMiddleware(nil), middleware.RequireBackofficeAccess())

		storefrontGroup := authenticated.Group("/storefront")
		{
			marketGroup := storefrontGroup.Group("/markets")
			marketGroup.GET("", noopHandler)
			marketGroup.GET("/options", noopHandler)
			marketGroup.GET("/:id", noopHandler)
			marketGroup.POST("", noopHandler)
			marketGroup.PUT("/:id", noopHandler)
			marketGroup.DELETE("/:id", noopHandler)
		}

		registerAuthenticatedCoreRoutes(authenticated, nil, nil, nil, nil)
		registerProductRoutes(authenticated, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		registerIntegrationRoutes(authenticated, nil, nil)
		registerMediaAndPreflightRoutes(authenticated, nil, nil, nil, nil, nil)
		registerCommerceRoutes(authenticated, nil, nil, nil, nil, nil, nil, nil)
		registerContentRoutes(authenticated, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		registerBusinessRoutes(authenticated, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		registerSystemRoutes(authenticated, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		registerOperationsRoutes(authenticated, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	})

	seen := make(map[string]struct{}, len(router.Routes()))
	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate route registered: %s", key)
		}
		seen[key] = struct{}{}
	}
}

func noopHandler(c *gin.Context) {}

func assertNoPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("route registration panicked: %s", fmt.Sprint(recovered))
		}
	}()

	fn()
}

package v1

import (
	"testing"

	"tanzanite/internal/app"
	"tanzanite/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesBuildsCompleteRouteTree(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	cfg := &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	}
	deps := &app.Dependencies{}

	RegisterRoutes(router, deps, cfg)
}

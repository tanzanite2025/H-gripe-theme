package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdminOnlyRequiresAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		role           string
		wantStatus     int
		wantNextCalled bool
	}{
		{name: "admin", role: "admin", wantStatus: http.StatusOK, wantNextCalled: true},
		{name: "manager", role: "manager", wantStatus: http.StatusForbidden},
		{name: "support", role: "support", wantStatus: http.StatusForbidden},
		{name: "missing role", wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			nextCalled := false
			router.Use(func(c *gin.Context) {
				if test.role != "" {
					c.Set("user_role", test.role)
				}
				c.Next()
			})
			router.GET("/", AdminOnly(), func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusOK)
			})

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

			require.Equal(t, test.wantStatus, recorder.Code)
			require.Equal(t, test.wantNextCalled, nextCalled)
		})
	}
}

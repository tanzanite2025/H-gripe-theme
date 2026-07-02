package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/securecookie"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testUserRepository struct {
	users map[uint]*user.User
}

func (r *testUserRepository) Create(u *user.User) error {
	return nil
}

func (r *testUserRepository) FindByEmail(email string) (*user.User, error) {
	return nil, errors.New("not implemented")
}

func (r *testUserRepository) FindByUsername(username string) (*user.User, error) {
	return nil, errors.New("not implemented")
}

func (r *testUserRepository) FindByID(id uint) (*user.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (r *testUserRepository) Update(u *user.User) error {
	return nil
}

func (r *testUserRepository) Delete(id uint) error {
	return nil
}

func (r *testUserRepository) List(offset, limit int) ([]user.User, int64, error) {
	return nil, 0, nil
}

func TestBackofficeDashboardGateUsesCurrentUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		tokenRole      string
		currentRole    string
		currentStatus  string
		expectedStatus int
	}{
		{
			name:           "admin can access dashboard",
			tokenRole:      "admin",
			currentRole:    "admin",
			currentStatus:  "active",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "front office user cannot access dashboard",
			tokenRole:      "user",
			currentRole:    "user",
			currentStatus:  "active",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "demoted staff token cannot access dashboard",
			tokenRole:      "admin",
			currentRole:    "user",
			currentStatus:  "active",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "inactive staff token is rejected",
			tokenRole:      "admin",
			currentRole:    "admin",
			currentStatus:  "suspended",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := service.NewAuthService(
				&testUserRepository{
					users: map[uint]*user.User{
						1: {
							ID:       1,
							Email:    "person@example.com",
							Username: "person",
							Role:     tt.currentRole,
							Status:   tt.currentStatus,
						},
					},
				},
				config.JWTConfig{
					Secret:             "test-secret",
					ExpireHours:        24,
					RefreshExpireHours: 168,
				},
			)
			token, err := authService.GenerateToken(&user.User{
				ID:       1,
				Email:    "person@example.com",
				Username: "person",
				Role:     tt.tokenRole,
				Status:   "active",
			})
			assert.NoError(t, err)

			router := gin.New()
			router.GET(
				"/api/admin/dashboard/stats",
				AuthMiddleware(authService),
				RequireBackofficeAccess(),
				RequireAnyPermission(auth.PermOrderView, auth.PermUserView, auth.PermTicketView, auth.PermSubscriptionView),
				func(c *gin.Context) {
					c.Status(http.StatusOK)
				},
			)

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/admin/dashboard/stats", nil)
			req.AddCookie(&http.Cookie{Name: securecookie.AuthTokenCookie, Value: token})
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAuthMiddlewareRejectsBearerWithoutCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := service.NewAuthService(
		&testUserRepository{
			users: map[uint]*user.User{
				1: {
					ID:       1,
					Email:    "person@example.com",
					Username: "person",
					Role:     "admin",
					Status:   "active",
				},
			},
		},
		config.JWTConfig{
			Secret:             "test-secret",
			ExpireHours:        24,
			RefreshExpireHours: 168,
		},
	)
	token, err := authService.GenerateToken(&user.User{
		ID:       1,
		Email:    "person@example.com",
		Username: "person",
		Role:     "admin",
		Status:   "active",
	})
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/private", AuthMiddleware(authService), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

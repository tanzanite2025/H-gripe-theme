package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/app"
	orderdomain "commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"
	userdomain "commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShippingTrackingRoutesEnforceAuthenticationAndOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, users := newShippingTrackingRouteFixture(t)
	authService := service.NewAuthService(users, config.JWTConfig{
		Secret:             "test-secret",
		ExpireHours:        24,
		RefreshExpireHours: 168,
	})
	router := gin.New()
	RegisterRoutes(router, &app.Dependencies{
		Services: app.Services{
			Auth:     authService,
			Order:    service.NewOrderService(nil, repository.NewOrderRepository(db), nil, nil),
			Shipping: service.NewShippingService(repository.NewShippingRepository(db)),
		},
	}, &config.Config{
		CORS: config.CORSConfig{},
		JWT:  config.JWTConfig{Secret: "test-secret"},
	})

	tokens := map[string]string{
		"owner": ownerTrackingRouteToken(t, authService, users.users[10]),
		"other": ownerTrackingRouteToken(t, authService, users.users[20]),
		"admin": ownerTrackingRouteToken(t, authService, users.users[99]),
	}

	tests := []struct {
		name   string
		path   string
		token  string
		status int
	}{
		{
			name:   "anonymous tracking number lookup is rejected",
			path:   "/api/v1/shipping/track/TRACK-ROUTE-1",
			status: http.StatusUnauthorized,
		},
		{
			name:   "wrong owner tracking number lookup is hidden",
			path:   "/api/v1/shipping/track/TRACK-ROUTE-1",
			token:  tokens["other"],
			status: http.StatusNotFound,
		},
		{
			name:   "missing tracking number uses the same generic not found response",
			path:   "/api/v1/shipping/track/TRACK-ROUTE-MISSING",
			token:  tokens["owner"],
			status: http.StatusNotFound,
		},
		{
			name:   "owner tracking number lookup is allowed",
			path:   "/api/v1/shipping/track/TRACK-ROUTE-1",
			token:  tokens["owner"],
			status: http.StatusOK,
		},
		{
			name:   "admin tracking number lookup is allowed",
			path:   "/api/v1/shipping/track/TRACK-ROUTE-1",
			token:  tokens["admin"],
			status: http.StatusOK,
		},
		{
			name:   "wrong owner order tracking lookup is rejected",
			path:   "/api/v1/shipping/orders/1/tracking",
			token:  tokens["other"],
			status: http.StatusForbidden,
		},
		{
			name:   "owner order tracking lookup is allowed",
			path:   "/api/v1/shipping/orders/1/tracking",
			token:  tokens["owner"],
			status: http.StatusOK,
		},
		{
			name:   "admin order tracking lookup is allowed",
			path:   "/api/v1/shipping/orders/1/tracking",
			token:  tokens["admin"],
			status: http.StatusOK,
		},
	}

	responses := make(map[string]*httptest.ResponseRecorder, len(tests))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.path, nil)
			if tt.token != "" {
				req.AddCookie(&http.Cookie{Name: securecookie.AuthTokenCookie, Value: tt.token})
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			responses[tt.name] = recorder

			require.Equal(t, tt.status, recorder.Code, recorder.Body.String())
			if tt.status == http.StatusOK {
				require.Contains(t, recorder.Body.String(), `"status":"in_transit"`)
			}
		})
	}

	require.Equal(
		t,
		responses["wrong owner tracking number lookup is hidden"].Body.String(),
		responses["missing tracking number uses the same generic not found response"].Body.String(),
	)
}

type shippingTrackingRouteUserRepository struct {
	users map[uint]*userdomain.User
}

func (r *shippingTrackingRouteUserRepository) Create(*userdomain.User) error {
	return nil
}

func (r *shippingTrackingRouteUserRepository) FindByEmail(string) (*userdomain.User, error) {
	return nil, errors.New("user not found")
}

func (r *shippingTrackingRouteUserRepository) FindByUsername(string) (*userdomain.User, error) {
	return nil, errors.New("user not found")
}

func (r *shippingTrackingRouteUserRepository) FindByID(id uint) (*userdomain.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *shippingTrackingRouteUserRepository) Update(*userdomain.User) error {
	return nil
}

func (r *shippingTrackingRouteUserRepository) Delete(uint) error {
	return nil
}

func (r *shippingTrackingRouteUserRepository) List(int, int) ([]userdomain.User, int64, error) {
	return nil, 0, nil
}

func newShippingTrackingRouteFixture(t *testing.T) (*gorm.DB, *shippingTrackingRouteUserRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&orderdomain.Order{},
		&orderdomain.OrderItem{},
		&shippingdomain.TrackingShipment{},
		&shippingdomain.TrackingEvent{},
	))

	order := orderdomain.Order{
		OrderNumber: "ORDER-TRACK-ROUTE",
		UserID:      10,
		TotalAmount: 100,
		Currency:    "USD",
	}
	require.NoError(t, db.Create(&order).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             order.ID,
		TrackingProviderID:  1,
		TrackingNumber:      "TRACK-ROUTE-1",
		ProviderCarrierCode: "carrier",
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID:             order.ID,
		TrackingNumber:      "TRACK-ROUTE-1",
		Status:              "in_transit",
		Description:         "Package is moving",
		ProviderCarrierCode: "carrier",
	}).Error)

	return db, &shippingTrackingRouteUserRepository{
		users: map[uint]*userdomain.User{
			10: {ID: 10, Email: "owner@example.com", Username: "owner", Role: "user", Status: "active"},
			20: {ID: 20, Email: "other@example.com", Username: "other", Role: "user", Status: "active"},
			99: {ID: 99, Email: "admin@example.com", Username: "admin", Role: "admin", Status: "active"},
		},
	}
}

func ownerTrackingRouteToken(t *testing.T, authService *service.AuthService, user *userdomain.User) string {
	t.Helper()

	token, err := authService.GenerateToken(user)
	require.NoError(t, err)
	return token
}

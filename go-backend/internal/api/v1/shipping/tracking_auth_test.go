package shipping

import (
	stdcontext "context"
	"net/http"
	"net/http/httptest"
	"testing"

	orderdomain "tanzanite/internal/domain/order"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAuthorizeOrderTrackingReadRejectsDifferentUser(t *testing.T) {
	db, orderService := newTestOrderTrackingAuthService(t)
	item := seedOrderTrackingAuthOrder(t, db, 10)

	handler := &Handler{orderService: orderService}
	context, recorder := orderTrackingAuthContext()
	context.Set("user_id", uint(20))

	assert.False(t, handler.authorizeOrderTrackingRead(context, item.ID))
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAuthorizeOrderTrackingReadAllowsOwner(t *testing.T) {
	db, orderService := newTestOrderTrackingAuthService(t)
	item := seedOrderTrackingAuthOrder(t, db, 10)

	handler := &Handler{orderService: orderService}
	context, _ := orderTrackingAuthContext()
	context.Set("user_id", uint(10))

	assert.True(t, handler.authorizeOrderTrackingRead(context, item.ID))
}

func TestAuthorizeOrderTrackingReadAllowsOrderViewRole(t *testing.T) {
	handler := &Handler{}
	context, _ := orderTrackingAuthContext()
	context.Set("role", "admin")

	assert.True(t, handler.authorizeOrderTrackingRead(context, 999))
}

func TestAuthorizeTrackingNumberReadRequiresOrderOwnership(t *testing.T) {
	db, orderService := newTestOrderTrackingAuthService(t)
	require.NoError(t, db.AutoMigrate(
		&shippingdomain.TrackingShipment{},
		&shippingdomain.TrackingEvent{},
	))

	item := seedOrderTrackingAuthOrder(t, db, 10)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             item.ID,
		TrackingProviderID:  1,
		TrackingNumber:      "TRACK-OWNER-1",
		ProviderCarrierCode: "carrier",
	}).Error)

	handler := &Handler{
		shippingService: service.NewShippingService(repository.NewShippingRepository(db)),
		orderService:    orderService,
	}
	context, recorder := orderTrackingAuthContext()
	context.Set("user_id", uint(20))

	assert.False(t, handler.authorizeTrackingNumberRead(context, "TRACK-OWNER-1"))
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	ownerContext, _ := orderTrackingAuthContext()
	ownerContext.Set("user_id", uint(10))
	assert.True(t, handler.authorizeTrackingNumberRead(ownerContext, "TRACK-OWNER-1"))
}

func newTestOrderTrackingAuthService(t *testing.T) (*gorm.DB, *service.OrderService) {
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

	require.NoError(t, db.AutoMigrate(&orderdomain.Order{}, &orderdomain.OrderItem{}))
	return db, service.NewOrderService(nil, repository.NewOrderRepository(db), nil)
}

func seedOrderTrackingAuthOrder(t *testing.T, db *gorm.DB, userID uint) orderdomain.Order {
	t.Helper()

	item := orderdomain.Order{
		OrderNumber: "ORDER-TRACK-AUTH",
		UserID:      userID,
		TotalAmount: 100,
	}
	require.NoError(t, db.Create(&item).Error)
	return item
}

func orderTrackingAuthContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequestWithContext(stdcontext.Background(), http.MethodGet, "/api/v1/shipping/orders/1/tracking", nil)
	return context, recorder
}

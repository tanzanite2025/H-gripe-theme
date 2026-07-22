package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestListTrackingEventsReturnsEventsForShipment(t *testing.T) {
	db, shippingService := newAdminShippingTestService(t)

	shipment := shippingdomain.TrackingShipment{
		OrderID:             1001,
		TrackingProviderID:  1,
		TrackingNumber:      "1Z999",
		ProviderCarrierCode: "190271",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		EventCount:          2,
		Enabled:             true,
	}
	if err := db.Create(&shipment).Error; err != nil {
		t.Fatalf("seed tracking shipment: %v", err)
	}

	deliveredAt := time.Date(2026, 7, 22, 12, 30, 0, 0, time.UTC)
	pickedUpAt := deliveredAt.Add(-2 * time.Hour)
	events := []shippingdomain.TrackingEvent{
		{
			OrderID:             shipment.OrderID,
			TrackingNumber:      shipment.TrackingNumber,
			ProviderCarrierCode: shipment.ProviderCarrierCode,
			Status:              "PICKUP",
			Location:            "Origin",
			Description:         "Picked up",
			EventTime:           pickedUpAt,
		},
		{
			OrderID:             shipment.OrderID,
			TrackingNumber:      shipment.TrackingNumber,
			ProviderCarrierCode: shipment.ProviderCarrierCode,
			Status:              "DELIVERED",
			Location:            "Destination",
			Description:         "Delivered",
			EventTime:           deliveredAt,
		},
	}
	if err := db.Create(&events).Error; err != nil {
		t.Fatalf("seed tracking events: %v", err)
	}

	handler := NewShippingHandler(shippingService)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/shipping/tracking-shipments/1001/events", nil)
	c.Params = gin.Params{{Key: "orderID", Value: "1001"}}

	handler.ListTrackingEvents(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Data []shippingdomain.TrackingEvent `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected response code 0, got %d", payload.Code)
	}
	if len(payload.Data.Data) != 2 {
		t.Fatalf("expected 2 tracking events, got %d", len(payload.Data.Data))
	}
	if payload.Data.Data[0].Description != "Delivered" {
		t.Fatalf("expected latest event first, got %#v", payload.Data.Data[0])
	}
}

func newAdminShippingTestService(t *testing.T) (*gorm.DB, *service.ShippingService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(
		&shippingdomain.ShippingTemplate{},
		&shippingdomain.Carrier{},
		&shippingdomain.CarrierService{},
		&shippingdomain.TrackingProviderConfig{},
		&shippingdomain.TrackingCarrierMapping{},
		&shippingdomain.TrackingShipment{},
		&shippingdomain.TrackingEvent{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	shippingRepo := repository.NewShippingRepository(db)
	return db, service.NewShippingService(shippingRepo)
}

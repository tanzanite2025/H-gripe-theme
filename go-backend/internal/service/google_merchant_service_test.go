package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"tanzanite/internal/domain/merchant"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGoogleMerchantUpdateOfferBlocksSyncedRemoteIdentityChange(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	seedGoogleMerchantOffer(t, googleMerchantService, productRecord.ID, variantRecord.ID, "synced")

	_, err := googleMerchantService.UpdateOffer(1, GoogleMerchantOfferInput{
		ProductID:             productRecord.ID,
		VariantID:             variantRecord.ID,
		OfferID:               "tz-wheel-updated",
		Brand:                 "Tanzanite",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		IdentifierExists:      boolPtrForGoogleMerchantTest(false),
		TargetCountry:         "US",
		ContentLanguage:       "en",
		CurrencyCode:          "USD",
		FeedLabel:             "US",
		PublicationStatus:     "ready",
	})
	if !errors.Is(err, ErrGoogleMerchantOfferInvalid) || !strings.Contains(err.Error(), "remove it from Google first") {
		t.Fatalf("UpdateOffer() error = %v, want remote identity guard", err)
	}
}

func TestGoogleMerchantUpdateOfferMarksSyncedPayloadChangeReady(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	seedGoogleMerchantOffer(t, googleMerchantService, productRecord.ID, variantRecord.ID, "synced")

	updated, err := googleMerchantService.UpdateOffer(1, GoogleMerchantOfferInput{
		ProductID:             productRecord.ID,
		VariantID:             variantRecord.ID,
		OfferID:               "tz-wheel",
		Title:                 "Updated Google Title",
		Brand:                 "Tanzanite",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		IdentifierExists:      boolPtrForGoogleMerchantTest(false),
		TargetCountry:         "US",
		ContentLanguage:       "en",
		CurrencyCode:          "USD",
		FeedLabel:             "US",
		PublicationStatus:     "ready",
	})
	if err != nil {
		t.Fatalf("UpdateOffer() error = %v", err)
	}
	if updated.SyncStatus != "ready" {
		t.Fatalf("SyncStatus = %q, want ready", updated.SyncStatus)
	}
	if updated.LastSyncAt == nil {
		t.Fatal("LastSyncAt was cleared, want preserved remote sync marker")
	}
	if !strings.Contains(updated.LastError, "local Google Merchant fields changed") {
		t.Fatalf("LastError = %q, want local change marker", updated.LastError)
	}
}

func TestGoogleMerchantDeleteOfferRequiresRemoteRemovalFirst(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	seedGoogleMerchantOffer(t, googleMerchantService, productRecord.ID, variantRecord.ID, "synced")

	err := googleMerchantService.DeleteOffer(1)
	if !errors.Is(err, ErrGoogleMerchantOfferInvalid) || !strings.Contains(err.Error(), "remove the offer from Google") {
		t.Fatalf("DeleteOffer() error = %v, want remote removal guard", err)
	}
}

func newTestGoogleMerchantService(t *testing.T) (*gorm.DB, *GoogleMerchantService, product.Product, product.ProductVariant) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(
		&product.ProductType{},
		&product.SpecDefinition{},
		&product.Product{},
		&product.ProductMedia{},
		&product.ProductSpecValue{},
		&product.ProductVariant{},
		&merchant.GoogleMerchantConnection{},
		&merchant.GoogleMerchantOffer{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	productRecord := product.Product{
		SKU:         "TZ-WHEEL",
		Name:        "Carbon Wheelset",
		Slug:        "carbon-wheelset",
		Description: "Fast carbon wheelset.",
		Price:       1299,
		Stock:       5,
		Status:      "active",
		Locale:      "en",
	}
	if err := db.Create(&productRecord).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	variantRecord := product.ProductVariant{
		ProductID: productRecord.ID,
		SKU:       "TZ-WHEEL-700C",
		Price:     1299,
		Stock:     5,
		IsActive:  true,
	}
	if err := db.Create(&variantRecord).Error; err != nil {
		t.Fatalf("create variant: %v", err)
	}

	service := NewGoogleMerchantService(
		repository.NewGoogleMerchantRepository(db),
		repository.NewProductRepository(db),
		nil,
		configlessGoogleMerchantForTest(),
		"https://tanzanite.site",
	)
	return db, service, productRecord, variantRecord
}

func seedGoogleMerchantOffer(t *testing.T, service *GoogleMerchantService, productID, variantID uint, syncStatus string) {
	t.Helper()

	identifierExists := false
	offer, err := service.CreateOffer(GoogleMerchantOfferInput{
		ProductID:             productID,
		VariantID:             variantID,
		OfferID:               "tz-wheel",
		Brand:                 "Tanzanite",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		IdentifierExists:      &identifierExists,
		TargetCountry:         "US",
		ContentLanguage:       "en",
		CurrencyCode:          "USD",
		FeedLabel:             "US",
		PublicationStatus:     "ready",
	})
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	if syncStatus == "" {
		return
	}
	now := time.Now().UTC()
	offer.SyncStatus = syncStatus
	offer.LastSyncAt = &now
	offer.LastError = ""
	if err := service.offers.UpdateOffer(offer); err != nil {
		t.Fatalf("mark synced offer: %v", err)
	}
}

func boolPtrForGoogleMerchantTest(value bool) *bool {
	return &value
}

func configlessGoogleMerchantForTest() config.GoogleMerchantConfig {
	return config.GoogleMerchantConfig{}
}

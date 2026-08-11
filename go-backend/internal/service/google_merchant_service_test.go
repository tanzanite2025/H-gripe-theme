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
		Brand:                 "H-GRIPE",
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
		Brand:                 "H-GRIPE",
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

func TestGoogleMerchantUpdateOfferQueuesReadyOfferRevalidationOnPayloadChange(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	offer := seedGoogleMerchantOffer(t, googleMerchantService, productRecord.ID, variantRecord.ID, "")
	publisher := &recordingMerchantPublisher{}
	googleMerchantService.ConfigureMerchantEventPublisher(publisher)

	_, err := googleMerchantService.UpdateOffer(offer.ID, GoogleMerchantOfferInput{
		ProductID:             productRecord.ID,
		VariantID:             variantRecord.ID,
		OfferID:               "tz-wheel",
		Title:                 "Updated Google Title",
		Brand:                 "H-GRIPE",
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
	if publisher.revalidate != 1 || publisher.revalidateOfferID != offer.ID || publisher.revalidateReason != "merchant_fields_changed" {
		t.Fatalf("revalidate event = count %d offer %d reason %q, want one event for offer %d", publisher.revalidate, publisher.revalidateOfferID, publisher.revalidateReason, offer.ID)
	}
}

func TestGoogleMerchantUpdateOfferQueuesRemoteWithdrawalWhenPaused(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	offer := seedGoogleMerchantOffer(t, googleMerchantService, productRecord.ID, variantRecord.ID, "synced")
	publisher := &recordingMerchantPublisher{}
	googleMerchantService.ConfigureMerchantEventPublisher(publisher)

	updated, err := googleMerchantService.UpdateOffer(offer.ID, googleMerchantOfferInputForTest(productRecord.ID, variantRecord.ID, "paused"))
	if err != nil {
		t.Fatalf("UpdateOffer() error = %v", err)
	}
	if updated.SyncStatus != "withdraw_pending" {
		t.Fatalf("SyncStatus = %q, want withdraw_pending", updated.SyncStatus)
	}
	if updated.LastSyncAt == nil {
		t.Fatal("LastSyncAt was cleared, want preserved remote sync marker")
	}
	if !strings.Contains(updated.LastError, "publication status changed") {
		t.Fatalf("LastError = %q, want publication status marker", updated.LastError)
	}
	if publisher.revalidate != 1 || publisher.revalidateOfferID != offer.ID || publisher.revalidateReason != "merchant_fields_changed" {
		t.Fatalf("revalidate event = count %d offer %d reason %q, want one withdrawal event for offer %d", publisher.revalidate, publisher.revalidateOfferID, publisher.revalidateReason, offer.ID)
	}
}

func TestGoogleMerchantUpdateOfferSkipsRevalidationWhenPayloadUnchanged(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	offer := seedGoogleMerchantOffer(t, googleMerchantService, productRecord.ID, variantRecord.ID, "")
	publisher := &recordingMerchantPublisher{}
	googleMerchantService.ConfigureMerchantEventPublisher(publisher)

	_, err := googleMerchantService.UpdateOffer(offer.ID, googleMerchantOfferInputForTest(productRecord.ID, variantRecord.ID, "ready"))
	if err != nil {
		t.Fatalf("UpdateOffer() error = %v", err)
	}
	if publisher.revalidate != 0 {
		t.Fatalf("revalidate events = %d, want 0 for unchanged payload", publisher.revalidate)
	}
}

func TestGoogleMerchantUpdateOfferSkipsDraftRevalidation(t *testing.T) {
	_, googleMerchantService, productRecord, variantRecord := newTestGoogleMerchantService(t)
	identifierExists := false
	offer, err := googleMerchantService.CreateOffer(GoogleMerchantOfferInput{
		ProductID:             productRecord.ID,
		VariantID:             variantRecord.ID,
		OfferID:               "tz-wheel",
		Brand:                 "H-GRIPE",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		IdentifierExists:      &identifierExists,
		TargetCountry:         "US",
		ContentLanguage:       "en",
		CurrencyCode:          "USD",
		FeedLabel:             "US",
		PublicationStatus:     "draft",
	})
	if err != nil {
		t.Fatalf("CreateOffer() error = %v", err)
	}
	publisher := &recordingMerchantPublisher{}
	googleMerchantService.ConfigureMerchantEventPublisher(publisher)

	_, err = googleMerchantService.UpdateOffer(offer.ID, GoogleMerchantOfferInput{
		ProductID:             productRecord.ID,
		VariantID:             variantRecord.ID,
		OfferID:               "tz-wheel",
		Title:                 "Draft Google Title",
		Brand:                 "H-GRIPE",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		IdentifierExists:      boolPtrForGoogleMerchantTest(false),
		TargetCountry:         "US",
		ContentLanguage:       "en",
		CurrencyCode:          "USD",
		FeedLabel:             "US",
		PublicationStatus:     "draft",
	})
	if err != nil {
		t.Fatalf("UpdateOffer() error = %v", err)
	}
	if publisher.revalidate != 0 {
		t.Fatalf("revalidate events = %d, want 0 for draft payload change", publisher.revalidate)
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
		configlessGoogleMerchantForTest(),
		"https://tanzanite.site",
	)
	return db, service, productRecord, variantRecord
}

func seedGoogleMerchantOffer(t *testing.T, service *GoogleMerchantService, productID, variantID uint, syncStatus string) *merchant.GoogleMerchantOffer {
	t.Helper()

	offer, err := service.CreateOffer(googleMerchantOfferInputForTest(productID, variantID, "ready"))
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	if syncStatus == "" {
		return offer
	}
	now := time.Now().UTC()
	offer.SyncStatus = syncStatus
	offer.LastSyncAt = &now
	offer.LastError = ""
	if err := service.offers.UpdateOffer(offer); err != nil {
		t.Fatalf("mark synced offer: %v", err)
	}
	return offer
}

func googleMerchantOfferInputForTest(productID, variantID uint, status string) GoogleMerchantOfferInput {
	identifierExists := false
	return GoogleMerchantOfferInput{
		ProductID:             productID,
		VariantID:             variantID,
		OfferID:               "tz-wheel",
		Brand:                 "H-GRIPE",
		Condition:             "new",
		GoogleProductCategory: "Sporting Goods",
		IdentifierExists:      &identifierExists,
		TargetCountry:         "US",
		ContentLanguage:       "en",
		CurrencyCode:          "USD",
		FeedLabel:             "US",
		PublicationStatus:     status,
	}
}

func boolPtrForGoogleMerchantTest(value bool) *bool {
	return &value
}

func configlessGoogleMerchantForTest() config.GoogleMerchantConfig {
	return config.GoogleMerchantConfig{}
}

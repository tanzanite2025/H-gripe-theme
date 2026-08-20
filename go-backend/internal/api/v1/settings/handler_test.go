package settings

import (
	"testing"

	"commerce-platform/internal/domain/media"
	settingdomain "commerce-platform/internal/domain/setting"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPublicSiteSettingsCanonicalizesKnownMediaURLs(t *testing.T) {
	handler := &Handler{}
	handler.ConfigureMediaService(service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30))

	settings := handler.publicSiteSettings(&settingdomain.SiteSettings{
		SiteLogo:    "http://media.internal:8080/uploads/site/logo.webp",
		SiteFavicon: "http://media.internal:8080/uploads/site/favicon.ico",
	})

	if settings.SiteLogo != "https://shop.example.test/uploads/site/logo.webp" {
		t.Fatalf("unexpected site logo URL: %s", settings.SiteLogo)
	}
	if settings.SiteFavicon != "https://shop.example.test/uploads/site/favicon.ico" {
		t.Fatalf("unexpected site favicon URL: %s", settings.SiteFavicon)
	}
}

func TestPublicWebsiteProfileCanonicalizesKnownMediaURLs(t *testing.T) {
	handler := &Handler{}
	handler.ConfigureMediaService(service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30))

	settings := handler.publicWebsiteProfile(&settingdomain.WebsiteProfileSettings{
		AvatarURL:       "http://media.internal:8080/uploads/profile/avatar.webp",
		FactoryImageURL: "http://media.internal:8080/uploads/profile/factory.webp",
	})

	if settings.AvatarURL != "https://shop.example.test/uploads/profile/avatar.webp" {
		t.Fatalf("unexpected avatar URL: %s", settings.AvatarURL)
	}
	if settings.FactoryImageURL != "https://shop.example.test/uploads/profile/factory.webp" {
		t.Fatalf("unexpected factory image URL: %s", settings.FactoryImageURL)
	}
}

func TestPublicSiteSettingsUsesCurrentDedicatedSiteLogo(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&sitelogodomain.Asset{}); err != nil {
		t.Fatalf("migrate site logo assets: %v", err)
	}
	if err := db.Create(&sitelogodomain.Asset{
		ID:         sitelogodomain.CurrentAssetID,
		Filename:   "current.svg",
		URL:        "http://media.internal:8080/uploads/site-logo/current.svg",
		StorageKey: "site-logo/current.svg",
		MimeType:   "image/svg+xml",
		Width:      48,
		Height:     48,
	}).Error; err != nil {
		t.Fatalf("seed current site logo: %v", err)
	}

	handler := &Handler{}
	handler.ConfigureMediaService(service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30))
	handler.ConfigureSiteLogoService(service.NewSiteLogoService(
		repository.NewSiteLogoRepository(db),
		nil,
		"https://shop.example.test",
	))

	settings := handler.publicSiteSettings(&settingdomain.SiteSettings{
		SiteLogo: "https://shop.example.test/uploads/site-logo/old.svg",
	})

	if settings.SiteLogo != "https://shop.example.test/uploads/site-logo/current.svg" {
		t.Fatalf("unexpected site logo URL: %s", settings.SiteLogo)
	}
	if settings.SiteLogoWidth != 48 || settings.SiteLogoHeight != 48 {
		t.Fatalf("unexpected site logo dimensions: %dx%d", settings.SiteLogoWidth, settings.SiteLogoHeight)
	}
}

func TestPublicSiteSettingsIncludesKnownLogoDimensions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&media.MediaAsset{}); err != nil {
		t.Fatalf("migrate media assets: %v", err)
	}
	if err := db.Create(&media.MediaAsset{
		Filename:   "logo.webp",
		URL:        "http://media.internal:8080/uploads/site/logo.webp",
		StorageKey: "site/logo.webp",
		MimeType:   "image/webp",
		MediaType:  "image",
		Status:     "active",
		Visibility: "public",
		Width:      320,
		Height:     80,
	}).Error; err != nil {
		t.Fatalf("seed logo asset: %v", err)
	}

	handler := &Handler{}
	handler.ConfigureMediaService(service.NewMediaService(
		repository.NewMediaRepository(db),
		nil,
		nil,
		"https://shop.example.test",
		20<<30,
	))

	settings := handler.publicSiteSettings(&settingdomain.SiteSettings{
		SiteLogo: "http://media.internal:8080/uploads/site/logo.webp?cache=1",
	})

	if settings.SiteLogo != "https://shop.example.test/uploads/site/logo.webp?cache=1" {
		t.Fatalf("unexpected site logo URL: %s", settings.SiteLogo)
	}
	if settings.SiteLogoWidth != 320 || settings.SiteLogoHeight != 80 {
		t.Fatalf("unexpected site logo dimensions: %dx%d", settings.SiteLogoWidth, settings.SiteLogoHeight)
	}
}

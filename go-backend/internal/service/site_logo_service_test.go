package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/setting"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSiteLogoUploadCurrentReplacesAndDestroysPrevious(t *testing.T) {
	db := newSiteLogoTestDB(t)
	uploadRoot := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	if err != nil {
		t.Fatalf("create storage service: %v", err)
	}

	logos := NewSiteLogoService(repository.NewSiteLogoRepository(db), storageService, "https://shop.example.test")
	logos.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	logos.ConfigureSettingService(NewSettingService(repository.NewSettingRepository(db), nil, 60))
	first, err := logos.UploadCurrent(context.Background(), createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "first.webp"), 11)
	if err != nil {
		t.Fatalf("upload first logo: %v", err)
	}
	firstPath := filepath.Join(uploadRoot, filepath.FromSlash(first.StorageKey))
	if _, err := os.Stat(firstPath); err != nil {
		t.Fatalf("expected first logo object to exist: %v", err)
	}

	second, err := logos.UploadCurrent(context.Background(), createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "second.webp"), 12)
	if err != nil {
		t.Fatalf("upload replacement logo: %v", err)
	}
	if first.StorageKey == second.StorageKey {
		t.Fatal("expected replacement logo to use a new storage key")
	}
	var cleanupEvent outbox.Event
	if err := db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&cleanupEvent).Error; err != nil {
		t.Fatalf("load logo cleanup event: %v", err)
	}
	if err := NewObjectStorageCleanupOutboxHandler(nil, logos, nil, nil).Handle(context.Background(), cleanupEvent); err != nil {
		t.Fatalf("deliver logo cleanup event: %v", err)
	}
	if _, err := os.Stat(firstPath); !os.IsNotExist(err) {
		t.Fatalf("expected previous logo object to be destroyed by outbox worker, got %v", err)
	}

	var count int64
	if err := db.Model(&sitelogodomain.Asset{}).Count(&count).Error; err != nil {
		t.Fatalf("count site logo rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one current site logo row, got %d", count)
	}

	current, err := repository.NewSiteLogoRepository(db).Current()
	if err != nil {
		t.Fatalf("load current site logo: %v", err)
	}
	if current == nil || current.StorageKey != second.StorageKey || current.ID != sitelogodomain.CurrentAssetID {
		t.Fatalf("unexpected current site logo row: %#v", current)
	}
	if second.URL != "https://shop.example.test/uploads/"+second.StorageKey {
		t.Fatalf("unexpected canonical public URL: %s", second.URL)
	}
}

func TestSiteLogoDeleteCurrentDestroysObjectAndRow(t *testing.T) {
	db := newSiteLogoTestDB(t)
	uploadRoot := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	if err != nil {
		t.Fatalf("create storage service: %v", err)
	}

	logos := NewSiteLogoService(repository.NewSiteLogoRepository(db), storageService, "https://shop.example.test")
	logos.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	logos.ConfigureSettingService(NewSettingService(repository.NewSettingRepository(db), nil, 60))
	current, err := logos.UploadCurrent(context.Background(), createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "current.webp"), 11)
	if err != nil {
		t.Fatalf("upload current logo: %v", err)
	}
	currentPath := filepath.Join(uploadRoot, filepath.FromSlash(current.StorageKey))

	if err := logos.DeleteCurrent(context.Background()); err != nil {
		t.Fatalf("delete current logo: %v", err)
	}
	var cleanupEvent outbox.Event
	if err := db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&cleanupEvent).Error; err != nil {
		t.Fatalf("load logo cleanup event: %v", err)
	}
	if err := NewObjectStorageCleanupOutboxHandler(nil, logos, nil, nil).Handle(context.Background(), cleanupEvent); err != nil {
		t.Fatalf("deliver logo cleanup event: %v", err)
	}
	if _, err := os.Stat(currentPath); !os.IsNotExist(err) {
		t.Fatalf("expected current logo object to be destroyed by outbox worker, got %v", err)
	}

	stored, err := repository.NewSiteLogoRepository(db).Current()
	if err != nil {
		t.Fatalf("load deleted site logo: %v", err)
	}
	if stored != nil {
		t.Fatalf("expected no current site logo row, got %#v", stored)
	}
}

func TestPublicUploadAccessAllowsOnlyCurrentSiteLogo(t *testing.T) {
	db := newSiteLogoTestDB(t)
	repo := repository.NewSiteLogoRepository(db)
	logos := NewSiteLogoService(repo, nil, "https://shop.example.test")
	err := db.Create(&sitelogodomain.Asset{
		ID:         sitelogodomain.CurrentAssetID,
		StorageKey: "site-logo/current.webp",
		URL:        "https://shop.example.test/uploads/site-logo/current.webp",
		MimeType:   "image/webp",
		Width:      512,
		Height:     512,
	}).Error
	if err != nil {
		t.Fatalf("seed current logo: %v", err)
	}

	access := NewPublicUploadAccessService(nil, nil)
	access.ConfigureSiteLogoService(logos)

	allowed, err := access.CanServePublicUpload(context.Background(), "site-logo/current.webp")
	if err != nil || !allowed {
		t.Fatalf("expected current logo to be allowed, allowed=%v err=%v", allowed, err)
	}

	allowed, err = access.CanServePublicUpload(context.Background(), "site-logo/old.webp")
	if err != nil {
		t.Fatalf("old logo access returned error: %v", err)
	}
	if allowed {
		t.Fatal("expected old site logo object to be denied")
	}
}

func newSiteLogoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

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

	if err := db.AutoMigrate(&sitelogodomain.Asset{}, &outbox.Event{}, &setting.Setting{}); err != nil {
		t.Fatalf("migrate site logo assets: %v", err)
	}
	return db
}

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"commerce-platform/internal/domain/outbox"
	settingdomain "commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSiteFaviconLifecycleKeepsSettingAndObjectCleanupConsistent(t *testing.T) {
	db := newSiteFaviconTestDB(t)
	root := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://media.example.test"})
	if err != nil {
		t.Fatalf("create storage service: %v", err)
	}
	settings := NewSettingService(repository.NewSettingRepository(db), nil, 60)
	favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(db), storageService, "https://shop.example.test")
	favicons.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	favicons.ConfigureSettingService(settings)

	first, err := favicons.UploadCurrent(context.Background(), siteLogoTestFileHeader(t, "first.webp"), 11)
	if err != nil {
		t.Fatalf("upload first favicon: %v", err)
	}
	firstPath := filepath.Join(root, filepath.FromSlash(first.StorageKey))
	if _, err := os.Stat(firstPath); err != nil {
		t.Fatalf("expected first favicon object: %v", err)
	}
	stored, err := settings.Get("site_favicon", "en")
	if err != nil || stored.Value != first.URL {
		t.Fatalf("expected setting to point at first favicon, setting=%#v err=%v", stored, err)
	}

	second, err := favicons.UploadCurrent(context.Background(), siteLogoTestFileHeader(t, "second.webp"), 12)
	if err != nil {
		t.Fatalf("upload replacement favicon: %v", err)
	}
	if second.StorageKey == first.StorageKey {
		t.Fatal("expected replacement favicon to use a new key")
	}
	var cleanupEvent outbox.Event
	if err := db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).First(&cleanupEvent).Error; err != nil {
		t.Fatalf("load favicon cleanup event: %v", err)
	}
	cleanupHandler := NewObjectStorageCleanupOutboxHandler(nil, nil, nil, nil)
	cleanupHandler.ConfigureSiteFaviconService(favicons)
	if err := cleanupHandler.Handle(context.Background(), cleanupEvent); err != nil {
		t.Fatalf("deliver favicon cleanup event: %v", err)
	}
	if _, err := os.Stat(firstPath); !os.IsNotExist(err) {
		t.Fatalf("expected old favicon object to be cleaned, got %v", err)
	}
	stored, err = settings.Get("site_favicon", "en")
	if err != nil || stored.Value != second.URL {
		t.Fatalf("expected setting to point at replacement favicon, setting=%#v err=%v", stored, err)
	}
	allowed, err := favicons.CanServePublicFavicon(context.Background(), second.StorageKey)
	if err != nil || !allowed {
		t.Fatalf("expected current favicon to be public, allowed=%v err=%v", allowed, err)
	}
	allowed, err = favicons.CanServePublicFavicon(context.Background(), first.StorageKey)
	if err != nil || allowed {
		t.Fatalf("expected retired favicon to be private, allowed=%v err=%v", allowed, err)
	}

	if err := favicons.DeleteCurrent(context.Background()); err != nil {
		t.Fatalf("delete favicon: %v", err)
	}
	stored, err = settings.Get("site_favicon", "en")
	if err != nil || stored.Value != "" {
		t.Fatalf("expected favicon setting to clear, setting=%#v err=%v", stored, err)
	}
	var cleanupEvents []outbox.Event
	if err := db.Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).Order("id DESC").Find(&cleanupEvents).Error; err != nil {
		t.Fatalf("load delete cleanup event: %v", err)
	}
	if len(cleanupEvents) < 2 {
		t.Fatalf("expected replacement and delete cleanup events, got %d", len(cleanupEvents))
	}
	cleanupEvent = cleanupEvents[0]
	if err := cleanupHandler.Handle(context.Background(), cleanupEvent); err != nil {
		t.Fatalf("deliver delete cleanup event: %v", err)
	}
	secondPath := filepath.Join(root, filepath.FromSlash(second.StorageKey))
	if _, err := os.Stat(secondPath); !os.IsNotExist(err) {
		t.Fatalf("expected deleted favicon object to be cleaned, got %v", err)
	}
	current, err := repository.NewSiteFaviconRepository(db).Current()
	if err != nil {
		t.Fatalf("load deleted favicon row: %v", err)
	}
	if current != nil {
		t.Fatalf("expected no current favicon row, got %#v", current)
	}
}

func TestPublicUploadAccessAllowsOnlyCurrentSiteFavicon(t *testing.T) {
	db := newSiteFaviconTestDB(t)
	repo := repository.NewSiteFaviconRepository(db)
	favicons := NewSiteFaviconService(repo, nil, "https://shop.example.test")
	if err := repo.ReplaceCurrent(&sitefavicondomain.Asset{
		StorageKey: "site-favicon/current.webp",
		URL:        "https://shop.example.test/uploads/site-favicon/current.webp",
		MimeType:   "image/webp",
		Width:      512,
		Height:     512,
	}, nil); err != nil {
		t.Fatalf("seed current favicon: %v", err)
	}

	access := NewPublicUploadAccessService(nil, nil)
	access.ConfigureSiteFaviconService(favicons)
	allowed, err := access.CanServePublicUpload(context.Background(), "site-favicon/current.webp")
	if err != nil || !allowed {
		t.Fatalf("expected current favicon to be allowed, allowed=%v err=%v", allowed, err)
	}
	allowed, err = access.CanServePublicUpload(context.Background(), "site-favicon/old.webp")
	if err != nil || allowed {
		t.Fatalf("expected old favicon to be denied, allowed=%v err=%v", allowed, err)
	}
}

func TestSiteFaviconMissingObjectDoesNotProducePublicURL(t *testing.T) {
	db := newSiteFaviconTestDB(t)
	root := t.TempDir()
	storageService, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://media.example.test"})
	if err != nil {
		t.Fatalf("create storage service: %v", err)
	}
	repo := repository.NewSiteFaviconRepository(db)
	if err := repo.ReplaceCurrent(&sitefavicondomain.Asset{StorageKey: "site-favicon/missing.webp", URL: "https://shop.example.test/uploads/site-favicon/missing.webp"}, nil); err != nil {
		t.Fatalf("seed missing favicon: %v", err)
	}
	favicons := NewSiteFaviconService(repo, storageService, "https://shop.example.test")
	if got := favicons.CurrentPublicURL(); got != "" {
		t.Fatalf("expected missing physical favicon to resolve to empty URL, got %q", got)
	}
	allowed, err := favicons.CanServePublicFavicon(context.Background(), "site-favicon/missing.webp")
	if err != nil || allowed {
		t.Fatalf("expected missing physical favicon to be denied, allowed=%v err=%v", allowed, err)
	}
}

func newSiteFaviconTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&sitefavicondomain.Asset{}, &outbox.Event{}, &settingdomain.Setting{}); err != nil {
		t.Fatalf("migrate favicon test db: %v", err)
	}
	return db
}

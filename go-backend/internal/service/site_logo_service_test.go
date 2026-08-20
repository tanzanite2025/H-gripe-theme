package service

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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
	first, err := logos.UploadCurrent(context.Background(), siteLogoTestFileHeader(t, "first.svg"), 11)
	if err != nil {
		t.Fatalf("upload first logo: %v", err)
	}
	firstPath := filepath.Join(uploadRoot, filepath.FromSlash(first.StorageKey))
	if _, err := os.Stat(firstPath); err != nil {
		t.Fatalf("expected first logo object to exist: %v", err)
	}

	second, err := logos.UploadCurrent(context.Background(), siteLogoTestFileHeader(t, "second.svg"), 12)
	if err != nil {
		t.Fatalf("upload replacement logo: %v", err)
	}
	if first.StorageKey == second.StorageKey {
		t.Fatal("expected replacement logo to use a new storage key")
	}
	if _, err := os.Stat(firstPath); !os.IsNotExist(err) {
		t.Fatalf("expected previous logo object to be destroyed, got %v", err)
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
	current, err := logos.UploadCurrent(context.Background(), siteLogoTestFileHeader(t, "current.svg"), 11)
	if err != nil {
		t.Fatalf("upload current logo: %v", err)
	}
	currentPath := filepath.Join(uploadRoot, filepath.FromSlash(current.StorageKey))

	if err := logos.DeleteCurrent(context.Background()); err != nil {
		t.Fatalf("delete current logo: %v", err)
	}
	if _, err := os.Stat(currentPath); !os.IsNotExist(err) {
		t.Fatalf("expected current logo object to be destroyed, got %v", err)
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
	_, err := repo.ReplaceCurrent(&sitelogodomain.Asset{
		StorageKey: "site-logo/current.svg",
		URL:        "https://shop.example.test/uploads/site-logo/current.svg",
		Width:      48,
		Height:     48,
	})
	if err != nil {
		t.Fatalf("seed current logo: %v", err)
	}

	access := NewPublicUploadAccessService(nil, nil)
	access.ConfigureSiteLogoService(logos)

	allowed, err := access.CanServePublicUpload(context.Background(), "site-logo/current.svg")
	if err != nil || !allowed {
		t.Fatalf("expected current logo to be allowed, allowed=%v err=%v", allowed, err)
	}

	allowed, err = access.CanServePublicUpload(context.Background(), "site-logo/old.svg")
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

	if err := db.AutoMigrate(&sitelogodomain.Asset{}); err != nil {
		t.Fatalf("migrate site logo assets: %v", err)
	}
	return db
}

func siteLogoTestFileHeader(t *testing.T, filename string) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 48 48"><path d="M0 0h48v48H0z"/></svg>`)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(int64(body.Len() + 1024)); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}
	files := request.MultipartForm.File["file"]
	if len(files) != 1 {
		t.Fatalf("expected 1 uploaded file, got %d", len(files))
	}
	return files[0]
}

package storage

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLocalStorageCreatesPublicReadableUploadPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are verified in Linux CI")
	}

	root := filepath.Join(t.TempDir(), "uploads")
	baseURL := "https://example.test"
	service, err := newLocalStorage(&Config{
		Type:      StorageTypeLocal,
		LocalPath: root,
		BaseURL:   baseURL,
	})
	if err != nil {
		t.Fatalf("newLocalStorage() error = %v", err)
	}

	url, err := service.UploadFromReader(context.Background(), strings.NewReader("logo"), "logo.png")
	if err != nil {
		t.Fatalf("UploadFromReader() error = %v", err)
	}

	key := strings.TrimPrefix(url, baseURL+"/uploads/")
	filePath := filepath.Join(root, filepath.FromSlash(key))
	for _, dir := range uploadPathDirectories(root, filePath) {
		assertPerm(t, dir, localDirPerm)
	}
	assertPerm(t, filePath, localFilePerm)
}

func TestLoadSiteLogoConfigFromEnvUsesDedicatedLocalPath(t *testing.T) {
	t.Setenv("SITE_LOGO_STORAGE_TYPE", "")
	t.Setenv("SITE_LOGO_STORAGE_LOCAL_PATH", "")
	t.Setenv("SITE_LOGO_STORAGE_BASE_URL", "")

	cfg := LoadSiteLogoConfigFromEnv(&Config{
		Type:      StorageTypeLocal,
		LocalPath: "/app/uploads",
		BaseURL:   "https://example.test",
	})

	if cfg.Type != StorageTypeLocal {
		t.Fatalf("Type = %q, want %q", cfg.Type, StorageTypeLocal)
	}
	if got, want := filepath.ToSlash(cfg.LocalPath), "/app/site-logo-uploads"; got != want {
		t.Fatalf("LocalPath = %q, want %q", got, want)
	}
	if cfg.BaseURL != "https://example.test" {
		t.Fatalf("BaseURL = %q, want inherited base URL", cfg.BaseURL)
	}
}

func TestLoadSiteLogoConfigFromEnvAllowsDedicatedBucket(t *testing.T) {
	t.Setenv("SITE_LOGO_STORAGE_TYPE", "s3")
	t.Setenv("SITE_LOGO_STORAGE_LOCAL_PATH", "")
	t.Setenv("SITE_LOGO_STORAGE_BASE_URL", "https://logo-cdn.example.test")
	t.Setenv("SITE_LOGO_STORAGE_BUCKET", "site-logo-bucket")
	t.Setenv("SITE_LOGO_STORAGE_REGION", "us-west-2")

	cfg := LoadSiteLogoConfigFromEnv(&Config{
		Type:      StorageTypeLocal,
		LocalPath: "/app/uploads",
		BaseURL:   "https://media.example.test",
		Bucket:    "media-bucket",
		Region:    "us-east-1",
	})

	if cfg.Type != StorageTypeS3 {
		t.Fatalf("Type = %q, want %q", cfg.Type, StorageTypeS3)
	}
	if cfg.Bucket != "site-logo-bucket" || cfg.Region != "us-west-2" {
		t.Fatalf("bucket/region = %q/%q, want dedicated site logo values", cfg.Bucket, cfg.Region)
	}
	if cfg.BaseURL != "https://logo-cdn.example.test" {
		t.Fatalf("BaseURL = %q, want dedicated base URL", cfg.BaseURL)
	}
}

func uploadPathDirectories(root, filePath string) []string {
	dirs := []string{root}
	dir := filepath.Dir(filePath)
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." {
		return dirs
	}
	current := root
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		current = filepath.Join(current, part)
		dirs = append(dirs, current)
	}
	return dirs
}

func assertPerm(t *testing.T, path string, expected os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if mode := info.Mode().Perm(); mode != expected {
		t.Fatalf("%s permissions = %v, want %v", path, mode, expected)
	}
}

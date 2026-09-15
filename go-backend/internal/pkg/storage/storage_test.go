package storage

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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

func TestLocalStorageRejectsSharedPrivatePath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "uploads")
	if _, err := newLocalStorage(&Config{
		Type:             StorageTypeLocal,
		LocalPath:        root,
		PrivateLocalPath: root,
	}); err == nil {
		t.Fatal("newLocalStorage() unexpectedly accepted shared public/private path")
	}
}

func TestLocalPrivateUploadUsesDedicatedPathAndObjectOperations(t *testing.T) {
	publicRoot := filepath.Join(t.TempDir(), "uploads")
	privateRoot := filepath.Join(t.TempDir(), "private-uploads")
	service, err := newLocalStorage(&Config{
		Type:             StorageTypeLocal,
		LocalPath:        publicRoot,
		BaseURL:          "https://public.example.test",
		PrivateLocalPath: privateRoot,
		PrivateBaseURL:   "https://private.example.test",
	})
	if err != nil {
		t.Fatalf("newLocalStorage() error = %v", err)
	}
	uploader, ok := service.(PrivateObjectUploader)
	if !ok {
		t.Fatal("local storage does not implement PrivateObjectUploader")
	}
	file := testStorageMultipartFile(t, "evidence.pdf", "application/pdf", []byte("sensitive"))
	reference, err := uploader.UploadWithPrefixPrivate(context.Background(), file, "warranty")
	if err != nil {
		t.Fatalf("UploadWithPrefixPrivate() error = %v", err)
	}
	key, err := service.ObjectKey(reference)
	if err != nil {
		t.Fatalf("ObjectKey() error = %v", err)
	}
	if !IsPrivateObjectKey(key) {
		t.Fatalf("private upload key %q is not private", key)
	}
	privatePath := filepath.Join(privateRoot, filepath.FromSlash(key))
	if _, err := os.Stat(privatePath); err != nil {
		t.Fatalf("private object missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(publicRoot, filepath.FromSlash(key))); !os.IsNotExist(err) {
		t.Fatalf("private object unexpectedly present in public root (err=%v)", err)
	}
	opener := service.(ObjectOpener)
	object, err := opener.Open(context.Background(), key)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	_ = object.ReadCloser.Close()
	if err := service.Delete(context.Background(), reference); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(privatePath); !os.IsNotExist(err) {
		t.Fatalf("private object still exists after Delete (err=%v)", err)
	}
}

func TestLocalUploadWithPrivatePrefixUsesDedicatedPath(t *testing.T) {
	publicRoot := filepath.Join(t.TempDir(), "uploads")
	privateRoot := filepath.Join(t.TempDir(), "private-uploads")
	service, err := newLocalStorage(&Config{
		Type:             StorageTypeLocal,
		LocalPath:        publicRoot,
		BaseURL:          "https://public.example.test",
		PrivateLocalPath: privateRoot,
		PrivateBaseURL:   "https://private.example.test",
	})
	if err != nil {
		t.Fatalf("newLocalStorage() error = %v", err)
	}
	file := testStorageMultipartFile(t, "evidence.pdf", "application/pdf", []byte("sensitive"))
	reference, err := service.UploadWithPrefix(context.Background(), file, "warranty")
	if err != nil {
		t.Fatalf("UploadWithPrefix() error = %v", err)
	}
	key, err := service.ObjectKey(reference)
	if err != nil {
		t.Fatalf("ObjectKey() error = %v", err)
	}
	if !IsPrivateObjectKey(key) {
		t.Fatalf("private prefix upload key %q is not private", key)
	}
	if _, err := os.Stat(filepath.Join(privateRoot, filepath.FromSlash(key))); err != nil {
		t.Fatalf("private object missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(publicRoot, filepath.FromSlash(key))); !os.IsNotExist(err) {
		t.Fatalf("private object unexpectedly present in public root (err=%v)", err)
	}
}

func testStorageMultipartFile(t *testing.T, filename, mimeType string, contents []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(contents); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err := request.ParseMultipartForm(int64(len(contents) + 1024)); err != nil {
		t.Fatalf("ParseMultipartForm() error = %v", err)
	}
	t.Cleanup(func() { _ = request.MultipartForm.RemoveAll() })
	file := request.MultipartForm.File["file"][0]
	file.Header.Set("Content-Type", mimeType)
	return file
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

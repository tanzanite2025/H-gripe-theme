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

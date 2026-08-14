package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizePublicUploadKeyRejectsTraversal(t *testing.T) {
	if _, ok := sanitizePublicUploadKey("/showcase/pending/../secret.webp"); ok {
		t.Fatal("sanitizePublicUploadKey accepted a traversal segment")
	}
	if key, ok := sanitizePublicUploadKey(`showcase\approved\2026\08\13\image.webp`); !ok || key != "showcase/approved/2026/08/13/image.webp" {
		t.Fatalf("sanitizePublicUploadKey normalized key = %q ok=%v", key, ok)
	}
}

func TestRestrictedUploadDirectoryDoesNotOpenDirectories(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "showcase", "approved"), 0o755); err != nil {
		t.Fatal(err)
	}

	dir := restrictedUploadDirectory{root: root}
	file, err := dir.Open("/showcase/approved")
	if err == nil {
		_ = file.Close()
		t.Fatal("restrictedUploadDirectory opened a directory")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("restrictedUploadDirectory directory error = %v, want os.ErrNotExist", err)
	}
}

func TestRestrictedUploadDirectoryOpensRegularFiles(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "showcase", "approved", "image.webp")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}

	dir := restrictedUploadDirectory{root: root}
	file, err := dir.Open("/showcase/approved/image.webp")
	if err != nil {
		t.Fatalf("restrictedUploadDirectory regular file error = %v", err)
	}
	defer func() { _ = file.Close() }()
}

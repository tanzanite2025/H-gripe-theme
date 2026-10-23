package storage

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// This adapter exposes only local reads and URL identity. Construction never
// creates directories, changes permissions, or constructs a remote client.
type readOnlyLocalRecoveryObjectStorage struct{ identity *localStorage }

func NewReadOnlyLocalRecoveryObjectStorageWithoutDirectoryCreation(config Config) (StorageService, error) {
	if config.Type != StorageTypeLocal || config.LocalPath == "" || config.PrivateLocalPath == "" {
		return nil, errors.New("explicit local public and private recovery roots are required")
	}
	return &readOnlyLocalRecoveryObjectStorage{identity: &localStorage{config: &config}}, nil
}

func (store *readOnlyLocalRecoveryObjectStorage) GetURL(key string) string {
	return store.identity.GetURL(key)
}
func (store *readOnlyLocalRecoveryObjectStorage) ObjectKey(reference string) (string, error) {
	return store.identity.ObjectKey(reference)
}
func (store *readOnlyLocalRecoveryObjectStorage) Open(ctx context.Context, key string) (*StoredObject, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	normalized, valid := NormalizeObjectKey(key)
	if !valid || normalized != key {
		return nil, os.ErrNotExist
	}
	rootPath := store.identity.config.LocalPath
	if IsPrivateObjectKey(key) {
		rootPath = store.identity.config.PrivateLocalPath
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}
	defer func(closeResource func() error) { _ = closeResource() }(root.Close)
	file, err := root.Open(filepath.FromSlash(key))
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, os.ErrNotExist
	}
	return &StoredObject{ReadCloser: file, Name: filepath.Base(key), MimeType: detectContentType(key), Size: info.Size(), ModTime: info.ModTime()}, nil
}

var errReadOnlyRecoveryStorageMutation = errors.New("recovery media verification storage refuses all mutations")

func (*readOnlyLocalRecoveryObjectStorage) Upload(context.Context, *multipart.FileHeader) (string, error) {
	return "", errReadOnlyRecoveryStorageMutation
}
func (*readOnlyLocalRecoveryObjectStorage) UploadWithPrefix(context.Context, *multipart.FileHeader, string) (string, error) {
	return "", errReadOnlyRecoveryStorageMutation
}
func (*readOnlyLocalRecoveryObjectStorage) UploadFromReader(context.Context, io.Reader, string) (string, error) {
	return "", errReadOnlyRecoveryStorageMutation
}
func (*readOnlyLocalRecoveryObjectStorage) UploadFromReaderWithPrefix(context.Context, io.Reader, string, string) (string, error) {
	return "", errReadOnlyRecoveryStorageMutation
}
func (*readOnlyLocalRecoveryObjectStorage) Delete(context.Context, string) error {
	return errReadOnlyRecoveryStorageMutation
}
func (*readOnlyLocalRecoveryObjectStorage) CopyObject(context.Context, string, string) error {
	return errReadOnlyRecoveryStorageMutation
}

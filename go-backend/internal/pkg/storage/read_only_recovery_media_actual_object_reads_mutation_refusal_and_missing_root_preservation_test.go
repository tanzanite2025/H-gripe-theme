package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadOnlyRecoveryMediaStorageReadsPublicPrivateAndRefusesMutationWithoutCreatingRoots(t *testing.T) {
	root := t.TempDir()
	publicRoot := filepath.Join(root, "public")
	privateRoot := filepath.Join(root, "private")
	provider, err := NewReadOnlyLocalRecoveryObjectStorageWithoutDirectoryCreation(Config{Type: StorageTypeLocal, LocalPath: publicRoot, PrivateLocalPath: privateRoot, BaseURL: "https://shop.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	_, err = os.Stat(publicRoot)
	require.True(t, os.IsNotExist(err))
	_, err = provider.Upload(context.Background(), nil)
	require.ErrorIs(t, err, errReadOnlyRecoveryStorageMutation)
	_, err = provider.UploadWithPrefix(context.Background(), nil, "catalog")
	require.ErrorIs(t, err, errReadOnlyRecoveryStorageMutation)
	_, err = provider.UploadFromReader(context.Background(), strings.NewReader("unexpected"), "new.png")
	require.ErrorIs(t, err, errReadOnlyRecoveryStorageMutation)
	_, err = provider.UploadFromReaderWithPrefix(context.Background(), strings.NewReader("unexpected"), "new.png", "catalog")
	require.ErrorIs(t, err, errReadOnlyRecoveryStorageMutation)
	require.ErrorIs(t, provider.Delete(context.Background(), "catalog/original.png"), errReadOnlyRecoveryStorageMutation)
	require.ErrorIs(t, provider.CopyObject(context.Background(), "catalog/original.png", "catalog/copy.png"), errReadOnlyRecoveryStorageMutation)
	_, err = os.Stat(publicRoot)
	require.True(t, os.IsNotExist(err))
	_, err = os.Stat(privateRoot)
	require.True(t, os.IsNotExist(err))
	for key, body := range map[string]string{"catalog/original.png": "public-original", "customer-service/attachments/originals/private.png": "private-original"} {
		selected := publicRoot
		if IsPrivateObjectKey(key) {
			selected = privateRoot
		}
		filename := filepath.Join(selected, filepath.FromSlash(key))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, []byte(body), 0600))
		object, err := provider.(ObjectOpener).Open(context.Background(), key)
		require.NoError(t, err)
		bytes, err := io.ReadAll(object.ReadCloser)
		require.NoError(t, err)
		require.NoError(t, object.ReadCloser.Close())
		require.Equal(t, body, string(bytes))
	}
	for _, key := range []string{"../private", "/catalog/original.png", "catalog"} {
		_, err := provider.(ObjectOpener).Open(context.Background(), key)
		require.Error(t, err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = provider.(ObjectOpener).Open(canceled, "catalog/original.png")
	require.ErrorIs(t, err, context.Canceled)
}

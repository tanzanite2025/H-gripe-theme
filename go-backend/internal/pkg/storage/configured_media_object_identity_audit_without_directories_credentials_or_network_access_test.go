package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfiguredMediaObjectIdentityAuditDoesNotCreateDirectoriesOrRequireCloudClients(t *testing.T) {
	root := filepath.Join(t.TempDir(), "must-not-be-created")
	for _, config := range []*Config{
		{Type: StorageTypeLocal, LocalPath: root, BaseURL: "https://media.example.test"},
		{Type: StorageTypeS3, Bucket: "public-media", Region: "test-1", BaseURL: "https://media.example.test"},
		{Type: StorageTypeOSS, Bucket: "public-media", Endpoint: "https://oss-test.example.test", BaseURL: "https://media.example.test"},
	} {
		key, err := ResolveConfiguredPublicMediaObjectKeyWithoutStorageAccess("https://media.example.test/catalog/image.png?token=ignored", config, "https://shop.example.test")
		require.NoError(t, err)
		require.Equal(t, "catalog/image.png", key)
		key, err = ResolveConfiguredPublicMediaObjectKeyWithoutStorageAccess("https://foreign.example.test/uploads/catalog/image.png", config, "https://shop.example.test")
		require.NoError(t, err)
		require.Empty(t, key)
	}
	_, err := os.Stat(root)
	require.True(t, os.IsNotExist(err))
	config := &Config{Type: StorageTypeS3, Bucket: "public-media", PrivateBucket: "private-media", Region: "test-1"}
	key, err := ResolveConfiguredPublicMediaObjectKeyWithoutStorageAccess("https://public-media.s3.test-1.amazonaws.com/catalog/image.png", config, "")
	require.NoError(t, err)
	require.Equal(t, "catalog/image.png", key)
	_, err = ResolveConfiguredPublicMediaObjectKeyWithoutStorageAccess("https://private-media.s3.test-1.amazonaws.com/catalog/image.png", config, "")
	require.Error(t, err)
}

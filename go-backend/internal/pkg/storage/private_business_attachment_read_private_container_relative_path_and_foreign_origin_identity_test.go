package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrivateBusinessAttachmentReadRejectsForeignOriginsPublicBucketsAndAmbiguousPaths(t *testing.T) {
	key := "warranty/claims/proof.png"
	config := &Config{BaseURL: "https://public.example.test/media", PrivateBaseURL: "https://private.example.test/evidence", Bucket: "public", PrivateBucket: "private", Region: "test-region", Endpoint: "https://objects.example.test"}
	providers := map[string]StorageService{"s3": &s3StorageImpl{config: config}, "oss": &ossStorageImpl{config: config}}
	for name, provider := range providers {
		t.Run(name, func(t *testing.T) {
			for _, reference := range []string{key, "/uploads/" + key, "uploads/" + key, config.PrivateBaseURL + "/" + key} {
				resolved, err := ResolvePrivateBusinessAttachmentObjectKeyForAuthorizedRead(provider, reference)
				require.NoError(t, err, reference)
				require.Equal(t, key, resolved)
			}
			for _, reference := range []string{
				config.BaseURL + "/" + key, "https://foreign.example.test/uploads/" + key,
				"https://objects.example.test/public/" + key, "https://public.s3.test-region.amazonaws.com/" + key,
				"/unexpected/uploads/" + key, "//private.example.test/evidence/" + key,
				"https://name@private.example.test/evidence/" + key, "https://private.example.test/evidence/warranty/claims/../proof.png",
				"/uploads/warranty/claims/%2e%2e/proof.png", "https://private.example.test/evidence/catalog/proof.png",
			} {
				_, err := ResolvePrivateBusinessAttachmentObjectKeyForAuthorizedRead(provider, reference)
				require.Error(t, err, reference)
			}
		})
	}
}

func TestPrivateBusinessAttachmentReadLocalHistoricalLocatorKeepsSeparatePrivateRoot(t *testing.T) {
	provider, err := NewStorageService(&Config{Type: StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), BaseURL: "https://shop.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	for _, reference := range []string{"https://shop.example.test/uploads/after-sales/pending/proof.png", "https://private.example.test/uploads/after-sales/pending/proof.png", "/uploads/after-sales/pending/proof.png"} {
		key, err := ResolvePrivateBusinessAttachmentObjectKeyForAuthorizedRead(provider, reference)
		require.NoError(t, err)
		require.Equal(t, "after-sales/pending/proof.png", key)
	}
	_, err = ResolvePrivateBusinessAttachmentObjectKeyForAuthorizedRead(provider, "https://untrusted.example.test/uploads/after-sales/pending/proof.png")
	require.Error(t, err)
}

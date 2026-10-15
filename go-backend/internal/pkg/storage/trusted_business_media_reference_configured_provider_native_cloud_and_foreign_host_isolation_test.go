package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrustedBusinessMediaReferenceIdentityUsesConfiguredOriginsAndFullKeys(t *testing.T) {
	local := &localStorage{config: &Config{BaseURL: "https://media.example.test/assets"}}
	s3 := &s3StorageImpl{config: &Config{BaseURL: "https://cdn.example.test/media", Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", Endpoint: "https://objects.example.test"}}
	oss := &ossStorageImpl{config: &Config{Bucket: "public-media", Endpoint: "https://oss-cn-hangzhou.aliyuncs.com"}}
	for _, test := range []struct {
		reference string
		provider  StorageService
		key       string
	}{
		{"/uploads/catalog/photo.png?v=2#preview", local, "catalog/photo.png"},
		{"uploads/catalog/photo.png", local, "catalog/photo.png"},
		{"https://media.example.test/assets/uploads/catalog/photo.png", local, "catalog/photo.png"},
		{"https://shop.example.test/store/uploads/catalog/photo.png", local, "catalog/photo.png"},
		{"https://foreign.example.test/uploads/catalog/photo.png", local, ""},
		{"https://media.example.test:9443/assets/uploads/catalog/photo.png", local, ""},
		{"https://shop.example.test/outside/uploads/catalog/photo.png", local, ""},
		{"/images/static/photo.png", local, ""},
		{"/other/uploads/catalog/photo.png", local, ""},
		{"https://cdn.example.test/media/catalog/photo.png?width=40", s3, "catalog/photo.png"},
		{"https://public-media.s3.us-east-1.amazonaws.com/catalog/photo.png?signature=temporary", s3, "catalog/photo.png"},
		{"https://objects.example.test/public-media/catalog/photo.png", s3, "catalog/photo.png"},
		{"https://objects.example.test/another-bucket/catalog/photo.png", s3, ""},
		{"https://public-media.s3.us-east-1.amazonaws.com.attacker.test/catalog/photo.png", s3, ""},
		{"https://public-media.oss-cn-hangzhou.aliyuncs.com/catalog/photo.png", oss, "catalog/photo.png"},
		{"https://foreign.oss-cn-hangzhou.aliyuncs.com/catalog/photo.png", oss, ""},
	} {
		t.Run(test.reference, func(t *testing.T) {
			key, err := TrustedBusinessMediaReferenceObjectKey(test.reference, test.provider, "https://shop.example.test/store")
			require.NoError(t, err)
			require.Equal(t, test.key, key)
		})
	}
}

func TestTrustedBusinessMediaReferenceIdentityRejectsUnsafeOrAmbiguousURLSyntax(t *testing.T) {
	for _, reference := range []string{"javascript:alert(1)", "data:image/png;base64,AAAA", "file:///uploads/file.png", "//media.example.test/uploads/file.png", "https://user@media.example.test/uploads/file.png", "/uploads/../private/file.png", "/uploads/%2e%2e/private/file.png", "/uploads/folder\\file.png", "https://media.example.test", "?file=photo.png"} {
		t.Run(reference, func(t *testing.T) {
			_, err := TrustedBusinessMediaReferenceObjectKey(reference, nil, "https://media.example.test")
			require.Error(t, err)
		})
	}
}

func TestTrustedBusinessMediaReferenceRejectsPrivateCloudBucketEvenWhenPublicBucketContainsSameKey(t *testing.T) {
	s3 := &s3StorageImpl{config: &Config{BaseURL: "https://public.example.test", PrivateBaseURL: "https://private.example.test", Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", Endpoint: "https://objects.example.test"}}
	oss := &ossStorageImpl{config: &Config{Bucket: "public-media", PrivateBucket: "private-media", Endpoint: "https://oss-cn-hangzhou.aliyuncs.com"}}
	local := &localStorage{config: &Config{BaseURL: "https://media.example.test", PrivateBaseURL: "https://private.example.test"}}
	for _, test := range []struct {
		reference string
		provider  StorageService
	}{
		{"https://private.example.test/catalog/photo.png", s3},
		{"https://private-media.s3.us-east-1.amazonaws.com/catalog/photo.png", s3},
		{"http://private-media.s3.us-east-1.amazonaws.com/catalog/photo.png", s3},
		{"https://objects.example.test/private-media/catalog/photo.png", s3},
		{"https://private-media.oss-cn-hangzhou.aliyuncs.com/catalog/photo.png", oss},
		{"https://private.example.test/uploads/catalog/photo.png", local},
	} {
		t.Run(test.reference, func(t *testing.T) {
			_, err := TrustedBusinessMediaReferenceObjectKey(test.reference, test.provider, "")
			require.Error(t, err)
		})
	}
}

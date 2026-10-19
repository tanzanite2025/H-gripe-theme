package storage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyPhysicalCapacityLocalMissingRootsAreNotCreated(t *testing.T) {
	root := filepath.Join(t.TempDir(), "not-created-public")
	provider := &localStorage{config: &Config{Type: StorageTypeLocal, LocalPath: root}}
	locations, err := provider.PhysicalCapacityLocations()
	require.NoError(t, err)
	require.Len(t, locations, 2)
	for _, location := range locations {
		require.NoError(t, provider.WalkPhysicalCapacityObjectMetadata(t.Context(), location, func(PhysicalCapacityObjectMetadata) error { t.Fatal("missing root produced an object"); return nil }))
		_, err := os.Lstat(location.localRoot)
		require.True(t, os.IsNotExist(err))
	}
}

func TestReadOnlyPhysicalCapacityLocalNestedProvidersSharePhysicalObjectIdentity(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "logos")
	require.NoError(t, os.MkdirAll(nested, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "same.png"), []byte("logo"), 0600))
	parent := &localStorage{config: &Config{LocalPath: root}}
	child := &localStorage{config: &Config{LocalPath: nested}}
	parentAddress, err := parent.PhysicalCapacityObjectAddress("logos/same.png")
	require.NoError(t, err)
	childAddress, err := child.PhysicalCapacityObjectAddress("same.png")
	require.NoError(t, err)
	require.Equal(t, parentAddress.ObjectID, childAddress.ObjectID)
	locations, err := parent.PhysicalCapacityLocations()
	require.NoError(t, err)
	var objects []PhysicalCapacityObjectMetadata
	require.NoError(t, parent.WalkPhysicalCapacityObjectMetadata(t.Context(), locations[0], func(object PhysicalCapacityObjectMetadata) error { objects = append(objects, object); return nil }))
	require.Len(t, objects, 1)
	require.Equal(t, parentAddress.ObjectID, objects[0].ObjectID)
	require.Equal(t, int64(4), objects[0].Size)
	_, err = parent.PhysicalCapacityObjectAddress("../outside.png")
	require.Error(t, err)
}

func TestReadOnlyPhysicalCapacityLocalRefusesChildSymlinksAndHonorsCancellation(t *testing.T) {
	root := t.TempDir()
	provider := &localStorage{config: &Config{LocalPath: root}}
	locations, err := provider.PhysicalCapacityLocations()
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	require.ErrorIs(t, provider.WalkPhysicalCapacityObjectMetadata(ctx, locations[0], func(PhysicalCapacityObjectMetadata) error { return nil }), context.Canceled)
	outside := filepath.Join(t.TempDir(), "private.txt")
	require.NoError(t, os.WriteFile(outside, []byte("outside"), 0600))
	if err := os.Symlink(outside, filepath.Join(root, "link.txt")); err != nil {
		t.Skip("OS does not permit test symlinks")
	}
	require.Error(t, provider.WalkPhysicalCapacityObjectMetadata(t.Context(), locations[0], func(PhysicalCapacityObjectMetadata) error { t.Fatal("followed symlink"); return nil }))
}

func TestReadOnlyPhysicalCapacityRemoteUsesEveryPageAndSeparatesPrivateContainerWithoutDownloadingBodies(t *testing.T) {
	for _, kind := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		for _, scenario := range []string{"complete", "repeated_cursor", "provider_error", "cancelled"} {
			t.Run(string(kind)+"/"+scenario, func(t *testing.T) {
				requests := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, "/private-media", strings.TrimSuffix(r.URL.Path, "/"))
					require.Equal(t, "1000", r.URL.Query().Get("max-keys"))
					require.Empty(t, r.URL.Query().Get("prefix"))
					if scenario == "provider_error" {
						w.WriteHeader(403)
						return
					}
					cursor := r.URL.Query().Get("continuation-token")
					if kind == StorageTypeOSS {
						cursor = r.URL.Query().Get("marker")
					}
					truncated, next, key, size := true, "page-two", "same.png", 7
					if cursor != "" {
						key, size = "second.png", 11
						if scenario == "complete" {
							truncated, next = false, ""
						}
					}
					w.Header().Set("Content-Type", "application/xml")
					_, _ = fmt.Fprintf(w, "<ListBucketResult><IsTruncated>%t</IsTruncated><NextContinuationToken>%s</NextContinuationToken><NextMarker>%s</NextMarker><Contents><Key>%s</Key><Size>%d</Size></Contents></ListBucketResult>", truncated, next, next, key, size)
				}))
				defer server.Close()
				cfg := &Config{Type: kind, Endpoint: server.URL, Region: "us-east-1", Bucket: "public-media", PrivateBucket: "private-media", AccessKeyID: "test-capacity-key", SecretAccessKey: "test-capacity-secret"}
				var provider ReadOnlyPhysicalCapacityMetadataProvider
				if kind == StorageTypeS3 {
					store, err := NewS3Storage(cfg)
					require.NoError(t, err)
					provider = store.(ReadOnlyPhysicalCapacityMetadataProvider)
				} else {
					client, err := oss.New(server.URL, cfg.AccessKeyID, cfg.SecretAccessKey, oss.ForcePathStyle(true))
					require.NoError(t, err)
					provider = &ossStorageImpl{config: cfg, client: client}
				}
				locations, err := provider.PhysicalCapacityLocations()
				require.NoError(t, err)
				require.Len(t, locations, 2)
				require.NotEqual(t, physicalCapacityIdentity(locations[0].ID, "same.png"), physicalCapacityIdentity(locations[1].ID, "same.png"))
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				if scenario == "cancelled" {
					cancel()
				}
				var sizes []int64
				err = provider.WalkPhysicalCapacityObjectMetadata(ctx, locations[1], func(object PhysicalCapacityObjectMetadata) error { sizes = append(sizes, object.Size); return nil })
				if scenario == "complete" {
					require.NoError(t, err)
					require.Equal(t, []int64{7, 11}, sizes)
					require.Equal(t, 2, requests)
				} else {
					require.Error(t, err)
				}
				if scenario == "cancelled" {
					require.Zero(t, requests)
				}
				if scenario == "repeated_cursor" {
					require.Equal(t, 2, requests)
				}
			})
		}
	}
}

func TestReadOnlyPhysicalCapacityRemoteIdentityIgnoresCredentialsAndRejectsAmbiguousPrivateStorage(t *testing.T) {
	first := &s3StorageImpl{config: &Config{Type: StorageTypeS3, Endpoint: "https://OBJECTS.example.test/", Bucket: "shared", AccessKeyID: "first"}}
	second := &s3StorageImpl{config: &Config{Type: StorageTypeS3, Endpoint: "https://objects.example.test", Bucket: "shared", AccessKeyID: "second"}}
	firstAddress, err := first.PhysicalCapacityObjectAddress("same.png")
	require.NoError(t, err)
	secondAddress, err := second.PhysicalCapacityObjectAddress("same.png")
	require.NoError(t, err)
	require.Equal(t, firstAddress, secondAddress)
	_, err = first.PhysicalCapacityObjectAddress(CustomerServiceConversationAttachmentOriginalsPrefix + "/private.png")
	require.Error(t, err)
}

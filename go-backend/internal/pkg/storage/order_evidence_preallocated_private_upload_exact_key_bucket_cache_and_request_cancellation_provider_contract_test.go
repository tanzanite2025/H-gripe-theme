package storage

import (
	"bytes"
	"context"

	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/require"
)

func TestOrderEvidencePreallocatedPrivateWriteUsesExactKeyPrivateBucketNoStoreAndCancellation(t *testing.T) {
	requirePreallocatedPrivateObjectProviderExactKeyPrivateBucketNoStoreAndCancellationContract(t, "order-evidence/41/57/owned-upload-11111111-1111-4111-8111-111111111111")
}

func requirePreallocatedPrivateObjectProviderExactKeyPrivateBucketNoStoreAndCancellationContract(t *testing.T, prefix string) {
	t.Helper()
	for _, kind := range []StorageType{StorageTypeLocal, StorageTypeS3, StorageTypeOSS} {
		t.Run(string(kind), func(t *testing.T) {
			type observedObjectRequest struct {
				path   string
				header http.Header
			}
			requests := make(chan observedObjectRequest, 8)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				requests <- observedObjectRequest{r.URL.Path, r.Header.Clone()}
				w.Header().Set("ETag", `"00000000000000000000000000000000"`)
			}))
			defer server.Close()
			cfg := &Config{Type: kind, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", Endpoint: server.URL, AccessKeyID: "synthetic", SecretAccessKey: "synthetic", BaseURL: "https://public.example.test/uploads", PrivateBaseURL: "https://private.example.test"}
			var provider StorageService
			var err error
			if kind == StorageTypeOSS {
				client, err := oss.New(server.URL, cfg.AccessKeyID, cfg.SecretAccessKey, oss.ForcePathStyle(true))
				require.NoError(t, err)
				public, err := client.Bucket(cfg.Bucket)
				require.NoError(t, err)
				private, err := client.Bucket(cfg.PrivateBucket)
				require.NoError(t, err)
				provider = &ossStorageImpl{config: cfg, client: client, bucket: public, privateBucket: private}
			} else {
				provider, err = NewStorageService(cfg)
				require.NoError(t, err)
			}
			key, err := AllocatePrivateObjectKeyForPendingUpload("proof.png", prefix)
			require.NoError(t, err)
			reference, err := provider.(PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("private evidence bytes")), key)
			require.NoError(t, err)
			resolved, err := provider.ObjectKey(reference)
			require.NoError(t, err)
			require.Equal(t, key, resolved)
			if kind == StorageTypeLocal {
				data, err := os.ReadFile(filepath.Join(cfg.PrivateLocalPath, filepath.FromSlash(key)))
				require.NoError(t, err)
				require.Equal(t, "private evidence bytes", string(data))
				_, err = os.Stat(filepath.Join(cfg.LocalPath, filepath.FromSlash(key)))
				require.True(t, os.IsNotExist(err))
			} else {
				request := <-requests
				require.Equal(t, "/private-media/"+key, request.path)
				require.Equal(t, "private, no-store", request.header.Get("Cache-Control"))
				if kind == StorageTypeOSS {
					require.Equal(t, "private", request.header.Get("X-Oss-Object-Acl"))
				}
			}
			canceled, cancel := context.WithCancel(t.Context())
			cancel()
			_, err = provider.(PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(canceled, bytes.NewReader([]byte("canceled")), key+".canceled")
			require.ErrorIs(t, err, context.Canceled)
			require.Empty(t, requests, "a canceled upload must not send a storage request")
		})
	}
}

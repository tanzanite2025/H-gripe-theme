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

func TestCustomerServicePreallocatedPrivateLocalWriterPreservesExactKeyRejectsOverwriteAndUnsafePaths(t *testing.T) {
	publicRoot, privateRoot := t.TempDir(), t.TempDir()
	store, err := NewStorageService(&Config{Type: StorageTypeLocal, LocalPath: publicRoot, PrivateLocalPath: privateRoot,
		BaseURL: "https://public.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	writer := store.(PreallocatedPrivateObjectWriter)
	key, err := AllocatePrivateObjectKeyForPendingUpload("customer.png", CustomerServiceConversationAttachmentOriginalsPrefix)
	require.NoError(t, err)
	reference, err := writer.WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("original")), key)
	require.NoError(t, err)
	resolvedKey, err := store.ObjectKey(reference)
	require.NoError(t, err)
	require.Equal(t, key, resolvedKey)
	payload, err := os.ReadFile(filepath.Join(privateRoot, filepath.FromSlash(key)))
	require.NoError(t, err)
	require.Equal(t, "original", string(payload))
	_, err = os.Stat(filepath.Join(publicRoot, filepath.FromSlash(key)))
	require.True(t, os.IsNotExist(err))
	_, err = writer.WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("overwrite")), key)
	require.Error(t, err)
	payload, err = os.ReadFile(filepath.Join(privateRoot, filepath.FromSlash(key)))
	require.NoError(t, err)
	require.Equal(t, "original", string(payload))
	for _, unsafeKey := range []string{"uploads/public.png", "../outside.png", CustomerServiceConversationAttachmentOriginalsPrefix + "/../foreign.png", "/" + key} {
		_, err = writer.WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("invalid")), unsafeKey)
		require.Error(t, err, unsafeKey)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = writer.WritePrivateObjectAtPreallocatedKey(ctx, bytes.NewReader([]byte("canceled")), CustomerServiceConversationAttachmentOriginalsPrefix+"/canceled.png")
	require.ErrorIs(t, err, context.Canceled)
	_, err = AllocatePrivateObjectKeyForPendingUpload("public.png", "uploads")
	require.Error(t, err)
}

func TestCustomerServicePreallocatedPrivateRemoteWriterUsesExactStoredKeyPrivateBucketAndNoStore(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		t.Run(string(provider), func(t *testing.T) {
			type recordedPrivateUpload struct {
				path, method string
				headers      http.Header
				payload      []byte
			}
			uploads := make(chan recordedPrivateUpload, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				payload, _ := io.ReadAll(r.Body)
				uploads <- recordedPrivateUpload{r.URL.Path, r.Method, r.Header.Clone(), payload}
				w.Header().Set("ETag", "\"00000000000000000000000000000000\"")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			cfg := &Config{Type: provider, Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1",
				Endpoint: server.URL, AccessKeyID: "test-key", SecretAccessKey: "test-secret",
				BaseURL: "https://public.example.test", PrivateBaseURL: "https://private.example.test"}
			var writer PreallocatedPrivateObjectWriter
			if provider == StorageTypeS3 {
				store, err := NewS3Storage(cfg)
				require.NoError(t, err)
				writer = store.(PreallocatedPrivateObjectWriter)
			} else {
				client, err := oss.New(server.URL, cfg.AccessKeyID, cfg.SecretAccessKey, oss.ForcePathStyle(true))
				require.NoError(t, err)
				privateBucket, err := client.Bucket(cfg.PrivateBucket)
				require.NoError(t, err)
				writer = &ossStorageImpl{config: cfg, client: client, privateBucket: privateBucket}
			}
			key := CustomerServiceConversationAttachmentOriginalsPrefix + "/preallocated-customer.png"
			reference, err := writer.WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("private-original")), key)
			require.NoError(t, err)
			require.Equal(t, cfg.PrivateBaseURL+"/"+key, reference)
			require.Len(t, uploads, 1)
			upload := <-uploads
			require.Equal(t, "/private-media/"+key, upload.path)
			require.Equal(t, http.MethodPut, upload.method)
			require.Equal(t, "private-original", string(upload.payload))
			require.Equal(t, "private, no-store", upload.headers.Get("Cache-Control"))
			if provider == StorageTypeOSS {
				require.Equal(t, "private", upload.headers.Get("X-Oss-Object-Acl"))
			} else {
				require.Empty(t, upload.headers.Get("X-Amz-Acl"))
			}
			_, err = writer.WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("public")), "uploads/public.png")
			require.Error(t, err)
			cfg.PrivateBucket = cfg.Bucket
			_, err = writer.WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("invalid")), key)
			require.Error(t, err)
			require.Empty(t, uploads)
		})
	}
}

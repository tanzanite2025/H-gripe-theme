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
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDedicatedImmediateOwnerPreallocatedWriterPreservesExactKeyPublicBucketCachePolicyAndNamespaceBoundary(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeLocal, StorageTypeS3, StorageTypeOSS} {
		t.Run(string(provider), func(t *testing.T) {
			type recordedDedicatedImmediateOwnerWrite struct {
				path    string
				header  http.Header
				payload []byte
			}
			uploads := make(chan recordedDedicatedImmediateOwnerWrite, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				payload, _ := io.ReadAll(r.Body)
				uploads <- recordedDedicatedImmediateOwnerWrite{r.URL.Path, r.Header.Clone(), payload}
				w.Header().Set("ETag", "\"00000000000000000000000000000000\"")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			cfg := &Config{Type: provider, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", Endpoint: server.URL, AccessKeyID: "synthetic", SecretAccessKey: "synthetic", BaseURL: "https://draft.example.test/uploads", PrivateBaseURL: "https://private.example.test"}
			var store StorageService
			var err error
			if provider == StorageTypeOSS {
				client, err := oss.New(server.URL, cfg.AccessKeyID, cfg.SecretAccessKey, oss.ForcePathStyle(true))
				require.NoError(t, err)
				bucket, err := client.Bucket(cfg.Bucket)
				require.NoError(t, err)
				store = &ossStorageImpl{config: cfg, client: client, bucket: bucket}
			} else {
				store, err = NewStorageService(cfg)
				require.NoError(t, err)
			}
			writer := store.(PreallocatedDedicatedImmediateOwnerObjectWriter)
			id := uuid.NewString()
			key, err := AllocateDedicatedImmediateOwnerUploadObjectKey("customer-service/avatars", "face.png")
			require.NoError(t, err)
			reference, err := writer.WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("draft bytes")), key, DedicatedImmediateOwnerUploadCacheControl)
			require.NoError(t, err)
			require.Equal(t, store.GetURL(key), reference)
			resolved, err := store.ObjectKey(reference)
			require.NoError(t, err)
			require.Equal(t, key, resolved)
			if provider == StorageTypeLocal {
				payload, err := os.ReadFile(filepath.Join(cfg.LocalPath, filepath.FromSlash(key)))
				require.NoError(t, err)
				require.Equal(t, "draft bytes", string(payload))
				_, err = os.Stat(filepath.Join(cfg.PrivateLocalPath, filepath.FromSlash(key)))
				require.True(t, os.IsNotExist(err))
				_, err = writer.WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("overwrite")), key, DedicatedImmediateOwnerUploadCacheControl)
				require.Error(t, err)
			} else {
				upload := <-uploads
				require.Equal(t, "/public-media/"+key, upload.path)
				require.Equal(t, "draft bytes", string(upload.payload))
				require.Equal(t, DedicatedImmediateOwnerUploadCacheControl, upload.header.Get("Cache-Control"))
				require.Empty(t, upload.header.Get("X-Oss-Object-Acl"))
				require.Empty(t, upload.header.Get("X-Amz-Acl"))
			}
			for _, unsafe := range []string{"../escape.png", "/" + key, "business-draft-media/not-uuid/file.png", "business-draft-media/" + id + "/../file.png", key + "?signature=untrusted", key + "#fragment", "site-logo/logo.png", "visual-showcase/preallocated-uploads/" + id + ".png", "customer-service/avatars/preallocated-uploads/not-a-uuid.png", CustomerServiceConversationAttachmentOriginalsPrefix + "/attachment.png"} {
				_, err = writer.WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("denied")), unsafe, DedicatedImmediateOwnerUploadCacheControl)
				require.Error(t, err, "%s", unsafe)
			}
			require.Empty(t, uploads)
			canceled, cancel := context.WithCancel(t.Context())
			cancel()
			_, err = writer.WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(canceled, bytes.NewReader([]byte("canceled")), key, DedicatedImmediateOwnerUploadCacheControl)
			require.Error(t, err)
		})
	}
}

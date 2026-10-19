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

func TestPermanentMediaOriginalPreallocatedWriterOnlyWritesExactPublicNamespaceAndRejectsDedicatedOrPrivateKeys(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeLocal, StorageTypeS3, StorageTypeOSS} {
		t.Run(string(provider), func(t *testing.T) {
			type recordedPermanentOriginalWrite struct {
				path    string
				header  http.Header
				payload []byte
			}
			uploads := make(chan recordedPermanentOriginalWrite, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				payload, _ := io.ReadAll(r.Body)
				uploads <- recordedPermanentOriginalWrite{r.URL.Path, r.Header.Clone(), payload}
				w.Header().Set("ETag", "\"00000000000000000000000000000000\"")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			cfg := &Config{Type: provider, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", Endpoint: server.URL, AccessKeyID: "synthetic", SecretAccessKey: "synthetic", BaseURL: "https://permanent.example.test/uploads", PrivateBaseURL: "https://private.example.test"}
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
			writer := store.(PreallocatedPermanentMediaOriginalObjectWriter)
			id := uuid.NewString()
			key, err := AllocatePermanentMediaOriginalObjectKey("catalog/product-images", ".png")
			require.NoError(t, err)
			reference, err := writer.WritePermanentMediaOriginalObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("permanent bytes")), key)
			require.NoError(t, err)
			require.Equal(t, store.GetURL(key), reference)
			resolved, err := store.ObjectKey(reference)
			require.NoError(t, err)
			require.Equal(t, key, resolved)
			if provider == StorageTypeLocal {
				payload, err := os.ReadFile(filepath.Join(cfg.LocalPath, filepath.FromSlash(key)))
				require.NoError(t, err)
				require.Equal(t, "permanent bytes", string(payload))
				_, err = os.Stat(filepath.Join(cfg.PrivateLocalPath, filepath.FromSlash(key)))
				require.True(t, os.IsNotExist(err))
				_, err = writer.WritePermanentMediaOriginalObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("overwrite")), key)
				require.Error(t, err)
			} else {
				upload := <-uploads
				require.Equal(t, "/public-media/"+key, upload.path)
				require.Equal(t, "permanent bytes", string(upload.payload))
				require.Empty(t, upload.header.Get("X-Oss-Object-Acl"))
				require.Empty(t, upload.header.Get("X-Amz-Acl"))
			}
			for _, unsafe := range []string{"../escape.png", "/" + key, "preallocated-permanent-originals/not-uuid.png", "business-draft-media/" + id + "/../file.png", key + "?signature=untrusted", key + "#fragment", "site-logo/preallocated-permanent-originals/" + id + ".png", "business-draft-media/preallocated-permanent-originals/" + id + ".png", "ugc-showcase/preallocated-permanent-originals/" + id + ".png", CustomerServiceConversationAttachmentOriginalsPrefix + "/attachment.png"} {
				_, err = writer.WritePermanentMediaOriginalObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("denied")), unsafe)
				require.Error(t, err, unsafe)
			}
			require.Empty(t, uploads)
			canceled, cancel := context.WithCancel(t.Context())
			cancel()
			_, err = writer.WritePermanentMediaOriginalObjectAtPreallocatedKey(canceled, bytes.NewReader([]byte("canceled")), key)
			require.Error(t, err)
		})
	}
}

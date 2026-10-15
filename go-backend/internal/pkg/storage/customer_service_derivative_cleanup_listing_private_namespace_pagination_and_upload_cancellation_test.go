package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/require"
)

func TestCustomerServicePrivateDerivativeLocalCleanupListingIsBoundedAndDoesNotCrossAssetNamespaces(t *testing.T) {
	privateRoot := t.TempDir()
	store, err := NewStorageService(&Config{Type: StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: privateRoot, BaseURL: "https://public.example.test"})
	require.NoError(t, err)
	prefix := CustomerServiceConversationAttachmentDerivativesPrefix + "/12/thumbnail/v1/"
	directory := filepath.Join(privateRoot, filepath.FromSlash(prefix))
	require.NoError(t, os.MkdirAll(directory, 0755))
	for index := range 1001 {
		require.NoError(t, os.WriteFile(filepath.Join(directory, fmt.Sprintf("%04d.png", index)), []byte("derivative"), 0600))
	}
	foreignKey := CustomerServiceConversationAttachmentDerivativesPrefix + "/123/thumbnail/v1/foreign.png"
	_, err = store.(PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("foreign")), foreignKey)
	require.NoError(t, err)
	lister := store.(CustomerServiceAttachmentDerivativeNamespaceCleanupLister)
	keys, err := lister.ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(t.Context(), 12)
	require.NoError(t, err)
	require.Len(t, keys, 1000)
	for _, key := range keys {
		require.Contains(t, key, prefix)
		require.NoError(t, store.Delete(t.Context(), key))
	}
	keys, err = lister.ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(t.Context(), 12)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.NoError(t, store.Delete(t.Context(), keys[0]))
	keys, err = lister.ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(t.Context(), 12)
	require.NoError(t, err)
	require.Empty(t, keys)
	_, err = os.Stat(filepath.Join(privateRoot, filepath.FromSlash(foreignKey)))
	require.NoError(t, err)
	_, err = lister.ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(t.Context(), 0)
	require.Error(t, err)
}

func TestCustomerServicePrivateDerivativeRemoteCleanupListingUsesPrivateBucketAndExactAssetPrefix(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		t.Run(string(provider), func(t *testing.T) {
			type recordedList struct{ path, prefix, limit string }
			requests := make(chan recordedList, 2)
			key := CustomerServiceConversationAttachmentDerivativesPrefix + "/12/thumbnail/v1/known.png"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- recordedList{r.URL.Path, r.URL.Query().Get("prefix"), r.URL.Query().Get("max-keys")}
				w.Header().Set("Content-Type", "application/xml")
				_, _ = io.WriteString(w, "<ListBucketResult><Name>private-media</Name><IsTruncated>false</IsTruncated><Contents><Key>"+key+"</Key></Contents></ListBucketResult>")
			}))
			defer server.Close()
			cfg := &Config{Type: provider, Endpoint: server.URL, Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1",
				AccessKeyID: "test-key", SecretAccessKey: "test-secret", BaseURL: "https://public.example.test", PrivateBaseURL: "https://private.example.test"}
			var lister CustomerServiceAttachmentDerivativeNamespaceCleanupLister
			if provider == StorageTypeS3 {
				store, err := NewS3Storage(cfg)
				require.NoError(t, err)
				lister = store.(CustomerServiceAttachmentDerivativeNamespaceCleanupLister)
			} else {
				client, err := oss.New(server.URL, cfg.AccessKeyID, cfg.SecretAccessKey, oss.ForcePathStyle(true))
				require.NoError(t, err)
				bucket, err := client.Bucket(cfg.PrivateBucket)
				require.NoError(t, err)
				lister = &ossStorageImpl{config: cfg, client: client, privateBucket: bucket}
			}
			keys, err := lister.ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(t.Context(), 12)
			require.NoError(t, err)
			require.Equal(t, []string{key}, keys)
			request := <-requests
			require.Equal(t, "/private-media", strings.TrimSuffix(request.path, "/"))
			require.Equal(t, CustomerServiceConversationAttachmentDerivativesPrefix+"/12/", request.prefix)
			require.Equal(t, "1000", request.limit)
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			_, err = lister.ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx, 12)
			require.Error(t, err)
			require.Empty(t, requests)
		})
	}
}

type privateOriginalUploadReaderCancelingContextAfterFirstChunk struct {
	cancel context.CancelFunc
	read   bool
}

func (r *privateOriginalUploadReaderCancelingContextAfterFirstChunk) Read(data []byte) (int, error) {
	if r.read {
		return 0, io.EOF
	}
	r.read = true
	r.cancel()
	return copy(data, []byte("partial original")), nil
}

func TestCustomerServicePreallocatedPrivateOriginalStopsLocalWriteWhenRequestContextIsCanceled(t *testing.T) {
	privateRoot := t.TempDir()
	store, err := NewStorageService(&Config{Type: StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: privateRoot, BaseURL: "https://public.example.test"})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	key := CustomerServiceConversationAttachmentOriginalsPrefix + "/canceled-during-write.png"
	_, err = store.(PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(ctx, &privateOriginalUploadReaderCancelingContextAfterFirstChunk{cancel: cancel}, key)
	require.ErrorIs(t, err, context.Canceled)
	payload, err := os.ReadFile(filepath.Join(privateRoot, filepath.FromSlash(key)))
	require.NoError(t, err)
	require.Equal(t, "partial original", string(payload), "persisted receipt key can locate the partial object for cleanup")
}

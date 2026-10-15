package storage

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/require"
)

func TestCustomerServicePrivateReaderUploadsUseSeparateBucketsNoStoreAndExplicitOSSPrivateACL(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		t.Run(string(provider), func(t *testing.T) {
			type recordedObjectUpload struct {
				method, path string
				headers      http.Header
				payload      []byte
			}
			uploads := make(chan recordedObjectUpload, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				payload, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				uploads <- recordedObjectUpload{method: r.Method, path: r.URL.Path, headers: r.Header.Clone(), payload: payload}
				w.Header().Set("ETag", "\"00000000000000000000000000000000\"")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			cfg := &Config{Type: provider, Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1",
				Endpoint: server.URL, AccessKeyID: "test-access-key", SecretAccessKey: "test-secret",
				BaseURL: "https://public.example.test", PrivateBaseURL: "https://private.example.test"}
			var store StorageService
			var err error
			if provider == StorageTypeS3 {
				store, err = NewS3Storage(cfg)
			} else {
				client, clientErr := oss.New(server.URL, cfg.AccessKeyID, cfg.SecretAccessKey, oss.ForcePathStyle(true))
				require.NoError(t, clientErr)
				publicBucket, bucketErr := client.Bucket(cfg.Bucket)
				require.NoError(t, bucketErr)
				privateBucket, bucketErr := client.Bucket(cfg.PrivateBucket)
				require.NoError(t, bucketErr)
				store = &ossStorageImpl{config: cfg, client: client, bucket: publicBucket, privateBucket: privateBucket}
			}
			require.NoError(t, err)
			prefix := CustomerServiceConversationAttachmentDerivativesPrefix + "/4/thumbnail/v1"
			originalPayload := []byte("private-derivative-content")
			reference, err := store.(PrivateCacheControlledReaderUploader).UploadPrivateObjectFromReaderWithPrefixAndCacheControl(
				t.Context(), bytes.NewReader(originalPayload), "thumbnail.png", prefix, "private, no-store")
			require.NoError(t, err)
			require.True(t, strings.HasPrefix(reference, cfg.PrivateBaseURL+"/"))
			require.Len(t, uploads, 1)
			upload := <-uploads
			require.Equal(t, http.MethodPut, upload.method)
			require.True(t, strings.HasPrefix(upload.path, "/private-media/"+prefix+"/"), upload.path)
			require.Equal(t, originalPayload, upload.payload)
			require.Equal(t, "private, no-store", upload.headers.Get("Cache-Control"))
			if provider == StorageTypeOSS {
				require.Equal(t, "private", upload.headers.Get("X-Oss-Object-Acl"))
			} else {
				require.Empty(t, upload.headers.Get("X-Amz-Acl"))
			}

			_, err = store.(CacheControlledReaderUploader).UploadFromReaderWithPrefixAndCacheControl(
				t.Context(), bytes.NewReader([]byte("public")), "public.png", "media-derivatives/5/thumbnail/v1", "public, max-age=31536000, immutable")
			require.NoError(t, err)
			require.Len(t, uploads, 1)
			publicUpload := <-uploads
			require.True(t, strings.HasPrefix(publicUpload.path, "/public-media/"))
			require.Equal(t, "public, max-age=31536000, immutable", publicUpload.headers.Get("Cache-Control"))
			require.Empty(t, publicUpload.headers.Get("X-Oss-Object-Acl"))
		})
	}
}

func TestCustomerServicePrivateReaderUploadsRejectPublicNamespaceAndSharedBucketOrOriginBeforeWriting(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		for _, scenario := range []string{"public namespace", "missing private bucket", "shared bucket", "shared origin"} {
			t.Run(string(provider)+"/"+scenario, func(t *testing.T) {
				cfg := &Config{Bucket: "public-media", PrivateBucket: "private-media", BaseURL: "https://public.example.test",
					PrivateBaseURL: "https://private.example.test"}
				prefix := CustomerServiceConversationAttachmentDerivativesPrefix + "/4/thumbnail/v1"
				switch scenario {
				case "public namespace":
					prefix = "media-derivatives/4/thumbnail/v1"
				case "missing private bucket":
					cfg.PrivateBucket = ""
				case "shared bucket":
					cfg.PrivateBucket = cfg.Bucket
				case "shared origin":
					cfg.PrivateBaseURL = cfg.BaseURL + "/"
				}
				var uploader PrivateCacheControlledReaderUploader = &s3StorageImpl{config: cfg}
				if provider == StorageTypeOSS {
					uploader = &ossStorageImpl{config: cfg}
				}
				// A nil network client makes any attempted write fail the test.
				reference, err := uploader.UploadPrivateObjectFromReaderWithPrefixAndCacheControl(
					t.Context(), bytes.NewReader([]byte("private")), "thumbnail.png", prefix, "private, no-store")
				require.Error(t, err)
				require.Empty(t, reference)
			})
		}
	}
}

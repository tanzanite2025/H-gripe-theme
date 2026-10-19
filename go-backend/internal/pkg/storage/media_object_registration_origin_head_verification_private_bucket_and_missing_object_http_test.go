package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/require"
)

func TestMediaObjectRegistrationExistenceVerificationUsesOriginHeadAndCorrectPublicOrPrivateBucket(t *testing.T) {
	for _, providerType := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		for _, objectState := range []string{"public-existing", "private-existing", "missing"} {
			t.Run(string(providerType)+"/"+objectState, func(t *testing.T) {
				var requests []struct{ method, path string }
				server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
					requests = append(requests, struct{ method, path string }{request.Method, request.URL.Path})
					if objectState == "missing" {
						response.WriteHeader(http.StatusNotFound)
						return
					}
					response.Header().Set("Content-Length", "42")
					response.WriteHeader(http.StatusOK)
				}))
				defer server.Close()
				config := &Config{Type: providerType, Bucket: "public-media-test", PrivateBucket: "private-media-test", Region: "us-east-1",
					AccessKeyID: "synthetic-test-key", SecretAccessKey: "synthetic-test-secret", Endpoint: server.URL, BaseURL: "https://unused-cdn.example.test"}
				var verifier ObjectExistenceVerifier
				if providerType == StorageTypeS3 {
					provider, err := NewStorageService(config)
					require.NoError(t, err)
					verifier = provider.(ObjectExistenceVerifier)
				} else {
					client, err := oss.New(server.URL, config.AccessKeyID, config.SecretAccessKey, oss.ForcePathStyle(true))
					require.NoError(t, err)
					publicBucket, err := client.Bucket(config.Bucket)
					require.NoError(t, err)
					privateBucket, err := client.Bucket(config.PrivateBucket)
					require.NoError(t, err)
					verifier = &ossStorageImpl{config: config, client: client, bucket: publicBucket, privateBucket: privateBucket}
				}
				key, expectedBucket := "catalog/image.png", config.Bucket
				if strings.HasPrefix(objectState, "private-") {
					key, expectedBucket = CustomerServiceConversationAttachmentOriginalsPrefix+"/image.png", config.PrivateBucket
				}
				err := verifier.VerifyStoredObjectExists(t.Context(), key)
				if objectState == "missing" {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				require.Len(t, requests, 1)
				require.Equal(t, http.MethodHead, requests[0].method)
				require.Equal(t, "/"+expectedBucket+"/"+key, requests[0].path)
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				require.Error(t, verifier.VerifyStoredObjectExists(ctx, key))
				require.Len(t, requests, 1, "canceled verification must not make an origin request")
			})
		}
	}
}

func TestMediaObjectRegistrationLocalExistenceVerificationRejectsMissingObjectsDirectoriesAndPublicPrivateRootConfusion(t *testing.T) {
	publicRoot, privateRoot := t.TempDir(), t.TempDir()
	provider, err := NewStorageService(&Config{Type: StorageTypeLocal, LocalPath: publicRoot, PrivateLocalPath: privateRoot})
	require.NoError(t, err)
	verifier := provider.(ObjectExistenceVerifier)
	privateKey := CustomerServiceConversationAttachmentOriginalsPrefix + "/private.png"
	publicCopy := filepath.Join(publicRoot, filepath.FromSlash(privateKey))
	require.NoError(t, os.MkdirAll(filepath.Dir(publicCopy), 0o755))
	require.NoError(t, os.WriteFile(publicCopy, []byte("wrong root"), 0o600))
	require.Error(t, verifier.VerifyStoredObjectExists(t.Context(), privateKey))
	privateCopy := filepath.Join(privateRoot, filepath.FromSlash(privateKey))
	require.NoError(t, os.MkdirAll(filepath.Dir(privateCopy), 0o755))
	require.NoError(t, os.WriteFile(privateCopy, []byte("private origin"), 0o600))
	require.NoError(t, verifier.VerifyStoredObjectExists(t.Context(), privateKey))
	require.Error(t, verifier.VerifyStoredObjectExists(t.Context(), "missing.png"))
	require.Error(t, verifier.VerifyStoredObjectExists(t.Context(), "customer-service"))
	require.Error(t, verifier.VerifyStoredObjectExists(t.Context(), "../outside.png"))
}

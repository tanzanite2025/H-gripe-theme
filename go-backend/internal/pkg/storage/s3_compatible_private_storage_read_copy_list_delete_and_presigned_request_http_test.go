package storage

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestS3CompatibleStoragePrivateReadsCopiesListsDeletesAndPresignsUseConfiguredEndpointAndPrivateBucket(t *testing.T) {
	type recordedStorageRequest struct {
		method, path string
		query        url.Values
		headers      http.Header
	}
	requests := make(chan recordedStorageRequest, 4)
	key := CustomerServiceConversationAttachmentOriginalsPrefix + "/2026/10/15/customer-photo.png"
	payload := "customer-private-photo"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- recordedStorageRequest{method: r.Method, path: r.URL.Path, query: r.URL.Query(), headers: r.Header.Clone()}
		switch {
		case r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = fmt.Fprintf(w, "<ListBucketResult><Name>private-media</Name><KeyCount>1</KeyCount><IsTruncated>false</IsTruncated><Contents><Key>%s</Key><Size>%d</Size></Contents></ListBucketResult>", key, len(payload))
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "private, no-store")
			w.Header().Set("Last-Modified", time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC).Format(http.TimeFormat))
			_, _ = io.WriteString(w, payload)
		case r.Method == http.MethodPut:
			w.Header().Set("Content-Type", "application/xml")
			_, _ = io.WriteString(w, "<CopyObjectResult><ETag>\"00000000000000000000000000000000\"</ETag><LastModified>2026-10-15T00:00:00Z</LastModified></CopyObjectResult>")
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	store, err := NewS3Storage(&Config{Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1",
		Endpoint: server.URL, AccessKeyID: "test-access-key", SecretAccessKey: "test-secret",
		BaseURL: "https://public.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	object, err := store.(ObjectOpener).Open(t.Context(), key)
	require.NoError(t, err)
	content, err := io.ReadAll(object.ReadCloser)
	require.NoError(t, err)
	require.NoError(t, object.ReadCloser.Close())
	require.Equal(t, payload, string(content))
	require.Equal(t, int64(len(payload)), object.Size)
	require.Equal(t, "image/png", object.MimeType)
	require.False(t, object.ModTime.IsZero())
	originalRequest := <-requests
	require.Equal(t, http.MethodGet, originalRequest.method)
	require.Equal(t, "/private-media/"+key, originalRequest.path)
	require.NotEmpty(t, originalRequest.headers.Get("Authorization"))

	copyKey := CustomerServiceConversationAttachmentOriginalsPrefix + "/2026/10/15/copy.png"
	require.NoError(t, store.CopyObject(t.Context(), key, copyKey))
	copyRequest := <-requests
	require.Equal(t, http.MethodPut, copyRequest.method)
	require.Equal(t, "/private-media/"+copyKey, copyRequest.path)
	require.Equal(t, "private-media/"+key, copyRequest.headers.Get("X-Amz-Copy-Source"))

	keys, err := store.(*s3StorageImpl).ListObjects(t.Context(), CustomerServiceConversationAttachmentOriginalsPrefix+"/", 7)
	require.NoError(t, err)
	require.Equal(t, []string{key}, keys)
	listRequest := <-requests
	require.Equal(t, "/private-media", strings.TrimSuffix(listRequest.path, "/"))
	require.Equal(t, "7", listRequest.query.Get("max-keys"))
	require.Equal(t, CustomerServiceConversationAttachmentOriginalsPrefix+"/", listRequest.query.Get("prefix"))

	require.NoError(t, store.Delete(t.Context(), store.GetURL(copyKey)))
	deleteRequest := <-requests
	require.Equal(t, http.MethodDelete, deleteRequest.method)
	require.Equal(t, "/private-media/"+copyKey, deleteRequest.path)

	signedReference, err := store.(PresignedURLProvider).GetPresignedURL(t.Context(), key, time.Minute)
	require.NoError(t, err)
	signedURL, err := url.Parse(signedReference)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(signedReference, server.URL+"/private-media/"))
	require.Equal(t, "/private-media/"+key, signedURL.Path)
	require.NotEmpty(t, signedURL.Query().Get("X-Amz-Signature"))
	require.Equal(t, "60", signedURL.Query().Get("X-Amz-Expires"))
	require.Empty(t, requests)
}

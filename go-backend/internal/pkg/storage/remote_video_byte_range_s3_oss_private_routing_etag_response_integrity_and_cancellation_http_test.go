package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/require"
)

func createRemoteVideoByteRangeProviderForSyntheticHTTPServer(t *testing.T, provider StorageType, endpoint string) StorageService {
	t.Helper()
	config := &Config{Type: provider, Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", Endpoint: endpoint,
		AccessKeyID: "synthetic-range-access", SecretAccessKey: "synthetic-range-secret", BaseURL: "https://public.example.test", PrivateBaseURL: "https://private.example.test"}
	if provider == StorageTypeS3 {
		store, err := NewS3Storage(config)
		require.NoError(t, err)
		return store
	}
	client, err := oss.New(endpoint, config.AccessKeyID, config.SecretAccessKey, oss.ForcePathStyle(true))
	require.NoError(t, err)
	public, err := client.Bucket(config.Bucket)
	require.NoError(t, err)
	private, err := client.Bucket(config.PrivateBucket)
	require.NoError(t, err)
	return &ossStorageImpl{config: config, client: client, bucket: public, privateBucket: private}
}

func TestRemoteVideoByteRangeProvidersRouteExactKeysAndUseMetadataBeforeConditionalBoundedReads(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		for _, key := range []string{CustomerServiceConversationAttachmentOriginalsPrefix + "/sample.mp4", CustomerServiceConversationAttachmentDerivativesPrefix + "/7/sample.mp4", "historical/sample.mp4"} {
			t.Run(string(provider)+"/"+key, func(t *testing.T) {
				type requestRecord struct{ method, path, byteRange, ifMatch, authorization string }
				requests := make(chan requestRecord, 8)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests <- requestRecord{r.Method, r.URL.Path, r.Header.Get("Range"), r.Header.Get("If-Match"), r.Header.Get("Authorization")}
					w.Header().Set("ETag", `"synthetic-video-v1"`)
					w.Header().Set("Last-Modified", "Mon, 19 Oct 2026 00:00:00 GMT")
					w.Header().Set("Content-Type", "video/mp4")
					if r.Method == http.MethodHead {
						w.Header().Set("Content-Length", "10")
						return
					}
					var start, end int
					if _, err := fmt.Sscanf(r.Header.Get("Range"), "bytes=%d-%d", &start, &end); err != nil {
						w.WriteHeader(400)
						return
					}
					w.Header().Set("Content-Length", strconv.Itoa(end-start+1))
					w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/10", start, end))
					w.WriteHeader(http.StatusPartialContent)
					_, _ = io.WriteString(w, "0123456789"[start:end+1])
				}))
				defer server.Close()
				store := createRemoteVideoByteRangeProviderForSyntheticHTTPServer(t, provider, server.URL)
				source, err := store.(ObjectByteRangeSourceOpener).OpenObjectByteRangeSource(t.Context(), key)
				require.NoError(t, err)
				require.Len(t, requests, 1, "opening metadata must not GET any video body")
				head := <-requests
				require.Equal(t, http.MethodHead, head.method)
				bucket := "public-media"
				if IsPrivateObjectKey(key) {
					bucket = "private-media"
				}
				require.Equal(t, "/"+bucket+"/"+key, head.path)
				require.NotEmpty(t, head.authorization)
				require.EqualValues(t, 10, source.Metadata().Size)
				body, err := source.OpenByteRange(t.Context(), 2, 4)
				require.NoError(t, err)
				payload, err := io.ReadAll(body)
				require.NoError(t, err)
				require.NoError(t, body.Close())
				require.Equal(t, "2345", string(payload))
				read := <-requests
				require.Equal(t, "bytes=2-5", read.byteRange)
				require.Equal(t, `"synthetic-video-v1"`, read.ifMatch)
				require.Equal(t, head.path, read.path)
				require.NotEmpty(t, read.authorization)
				all, err := io.ReadAll(source)
				require.NoError(t, err)
				require.Equal(t, "0123456789", string(all))
				require.Equal(t, "bytes=0-9", (<-requests).byteRange)
				require.NoError(t, source.Close())
				require.NoError(t, source.Close())
				_, err = source.OpenByteRange(t.Context(), 0, 1)
				require.Error(t, err)
				require.Empty(t, requests)
			})
		}
	}
}

func TestRemoteVideoByteRangeProvidersRejectUnverifiableMetadataAndIncorrectProviderResponses(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		for _, scenario := range []string{"missing etag", "weak etag", "encoded object", "ignored range", "wrong offset", "wrong total", "wrong length", "changed etag", "precondition failure", "truncated body"} {
			t.Run(string(provider)+"/"+scenario, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("ETag", `"pinned-version"`)
					w.Header().Set("Content-Length", "10")
					if r.Method == http.MethodHead {
						if scenario == "missing etag" {
							w.Header().Del("ETag")
						}
						if scenario == "weak etag" {
							w.Header().Set("ETag", `W/"version"`)
						}
						if scenario == "encoded object" {
							w.Header().Set("Content-Encoding", "gzip")
						}
						return
					}
					w.Header().Set("Content-Length", "4")
					w.Header().Set("Content-Range", "bytes 2-5/10")
					if scenario == "ignored range" {
						w.Header().Del("Content-Range")
						w.WriteHeader(200)
						_, _ = io.WriteString(w, "2345")
						return
					}
					if scenario == "wrong offset" {
						w.Header().Set("Content-Range", "bytes 1-4/10")
					}
					if scenario == "wrong total" {
						w.Header().Set("Content-Range", "bytes 2-5/11")
					}
					if scenario == "wrong length" {
						w.Header().Set("Content-Length", "3")
					}
					if scenario == "changed etag" {
						w.Header().Set("ETag", `"changed-version"`)
					}
					if scenario == "precondition failure" {
						w.Header().Del("Content-Length")
						w.WriteHeader(412)
						return
					}
					w.WriteHeader(206)
					if scenario == "truncated body" {
						_, _ = io.WriteString(w, "23")
						return
					}
					_, _ = io.WriteString(w, "2345")
				}))
				defer server.Close()
				store := createRemoteVideoByteRangeProviderForSyntheticHTTPServer(t, provider, server.URL)
				source, err := store.(ObjectByteRangeSourceOpener).OpenObjectByteRangeSource(t.Context(), CustomerServiceConversationAttachmentOriginalsPrefix+"/protected.mp4")
				if scenario == "missing etag" || scenario == "weak etag" || scenario == "encoded object" {
					require.Error(t, err)
					require.Nil(t, source)
					return
				}
				require.NoError(t, err)
				defer func() { _ = source.Close() }()
				body, err := source.OpenByteRange(t.Context(), 2, 4)
				if scenario == "truncated body" {
					require.NoError(t, err)
					_, err = io.ReadAll(body)
					require.Error(t, err)
					require.NoError(t, body.Close())
				} else {
					require.Error(t, err)
					require.Nil(t, body)
				}
			})
		}
	}
}

func TestRemoteVideoByteRangeProvidersCancelMetadataAndActiveResponseBodies(t *testing.T) {
	for _, provider := range []StorageType{StorageTypeS3, StorageTypeOSS} {
		t.Run(string(provider), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "10")
				w.Header().Set("ETag", `"version"`)
				if r.Method == http.MethodHead {
					return
				}
				w.Header().Set("Content-Range", "bytes 0-9/10")
				w.WriteHeader(206)
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			}))
			defer server.Close()
			store := createRemoteVideoByteRangeProviderForSyntheticHTTPServer(t, provider, server.URL)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			source, err := store.(ObjectByteRangeSourceOpener).OpenObjectByteRangeSource(ctx, CustomerServiceConversationAttachmentOriginalsPrefix+"/cancel.mp4")
			require.NoError(t, err)
			body, err := source.OpenByteRange(ctx, 0, 10)
			require.NoError(t, err)
			cancel()
			_, err = io.ReadAll(body)
			require.Error(t, err)
			require.NoError(t, body.Close())
			require.NoError(t, source.Close())
			_, err = store.(ObjectByteRangeSourceOpener).OpenObjectByteRangeSource(ctx, CustomerServiceConversationAttachmentOriginalsPrefix+"/cancel.mp4")
			require.Error(t, err)
		})
	}
}

func TestRemoteVideoByteRangeMetadataRejectsUnsafeETagsAndBoundsWithoutOpeningBodies(t *testing.T) {
	for _, etag := range []string{"", "unquoted", `W/"weak"`, "\"line\nbreak\"", "\"embedded\"quote\"", "\"" + strings.Repeat("a", 1024) + "\""} {
		require.Error(t, validatePinnedRemoteObjectByteRangeMetadata(ObjectByteRangeMetadata{Size: 10, ETag: etag}, ""))
	}
	source := &remoteObjectByteRangeSource{ctx: t.Context(), metadata: ObjectByteRangeMetadata{Size: 10, ETag: `"version"`}, openRange: func(context.Context, int64, int64) (io.ReadCloser, error) {
		t.Fatal("invalid bounds must not open a provider body")
		return nil, nil
	}}
	for _, bounds := range [][2]int64{{-1, 1}, {0, 0}, {10, 1}, {9, 2}, {1, 9223372036854775807}} {
		_, err := source.OpenByteRange(t.Context(), bounds[0], bounds[1])
		require.Error(t, err)
	}
}

package storage

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/stretchr/testify/require"
)

func TestUGCShowcasePreallocatedPrivateWriteAndApprovedCopyRouteBucketsAndHonorCancellation(t *testing.T) {
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
				if r.Header.Get("X-Amz-Copy-Source") != "" || r.Header.Get("X-Oss-Copy-Source") != "" {
					w.Header().Set("Content-Type", "application/xml")
					_, _ = io.WriteString(w, xml.Header+`<CopyObjectResult><LastModified>2026-10-16T00:00:00.000Z</LastModified><ETag>"00000000000000000000000000000000"</ETag></CopyObjectResult>`)
				}
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
			key, err := AllocatePrivateObjectKeyForPendingUpload("ugc.webp", "showcase/pending")
			require.NoError(t, err)
			reference, err := provider.(PreallocatedPrivateObjectWriter).WritePrivateObjectAtPreallocatedKey(t.Context(), bytes.NewReader([]byte("private ride image")), key)
			require.NoError(t, err)
			resolved, err := provider.ObjectKey(reference)
			require.NoError(t, err)
			require.Equal(t, key, resolved)
			if kind == StorageTypeLocal {
				data, err := os.ReadFile(filepath.Join(cfg.PrivateLocalPath, filepath.FromSlash(key)))
				require.NoError(t, err)
				require.Equal(t, "private ride image", string(data))
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
			destination := "showcase/approved/precommitted-test/ride.webp"
			require.NoError(t, provider.CopyObject(t.Context(), key, destination))
			if kind == StorageTypeLocal {
				data, err := os.ReadFile(filepath.Join(cfg.LocalPath, filepath.FromSlash(destination)))
				require.NoError(t, err)
				require.Equal(t, "private ride image", string(data))
			} else {
				request := <-requests
				require.Equal(t, "/public-media/"+destination, request.path)
				source := request.header.Get("X-Amz-Copy-Source")
				if kind == StorageTypeOSS {
					source = request.header.Get("X-Oss-Copy-Source")
				}
				decodedSource, err := url.PathUnescape(strings.TrimPrefix(source, "/"))
				require.NoError(t, err)
				require.Equal(t, "private-media/"+key, decodedSource)
			}
			canceled, cancel := context.WithCancel(t.Context())
			cancel()
			err = provider.CopyObject(canceled, key, destination+".canceled")
			require.ErrorIs(t, err, context.Canceled)
			require.Empty(t, requests, "a canceled copy must not send a storage request")
		})
	}
}

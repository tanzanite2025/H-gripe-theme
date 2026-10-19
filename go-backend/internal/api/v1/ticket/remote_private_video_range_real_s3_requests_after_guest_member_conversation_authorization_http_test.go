package ticket

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestRemotePrivateVideoRangeHTTPFetchesOnlyRequestedS3BytesAfterRealConversationAuthorization(t *testing.T) {
	for _, scenario := range []string{"visitor", "member", "missing cookie", "foreign member", "internal message", "unsatisfiable", "changed object"} {
		t.Run(scenario, func(t *testing.T) {
			member := scenario == "member" || scenario == "foreign member"
			env := createCustomerServicePrivateAttachmentHTTPTestEnvironment(t, member)
			type recordedRequest struct{ method, path, byteRange, ifMatch, cookie string }
			requests := make(chan recordedRequest, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- recordedRequest{r.Method, r.URL.Path, r.Header.Get("Range"), r.Header.Get("If-Match"), r.Header.Get("Cookie")}
				w.Header().Set("Content-Length", "10")
				w.Header().Set("ETag", `"pinned-video-version"`)
				w.Header().Set("Content-Type", "video/mp4")
				if r.Method == http.MethodHead {
					return
				}
				if scenario == "changed object" {
					w.Header().Del("Content-Length")
					w.WriteHeader(412)
					return
				}
				w.Header().Set("Content-Length", "4")
				w.Header().Set("Content-Range", "bytes 2-5/10")
				w.WriteHeader(206)
				_, _ = io.WriteString(w, "2345")
			}))
			defer server.Close()
			store, err := storage.NewS3Storage(&storage.Config{Type: storage.StorageTypeS3, Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", Endpoint: server.URL,
				AccessKeyID: "synthetic-range-key", SecretAccessKey: "synthetic-range-secret", BaseURL: "https://shop.example.test", PrivateBaseURL: "https://private.example.test"})
			require.NoError(t, err)
			key := storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/synthetic-video.mp4"
			url := store.GetURL(key)
			require.NoError(t, env.db.Model(&env.asset).Updates(map[string]any{"storage_key": key, "url": url, "media_type": "video", "mime_type": "video/mp4", "filename": "video.mp4", "original_filename": "video.mp4", "size": 10}).Error)
			attachments, err := json.Marshal([]string{url})
			require.NoError(t, err)
			require.NoError(t, env.db.Model(&env.message).Updates(map[string]any{"attachments": string(attachments), "message_type": "video", "is_internal": scenario == "internal message"}).Error)
			mediaService := service.NewMediaService(repository.NewMediaRepository(env.db), store, nil, "https://shop.example.test", 20<<30)
			env.handler.conversationAttachmentAccess = service.NewCustomerServiceConversationMessageAttachmentAccessService(env.handler.ticketService, mediaService)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("/api/v1/customer-service/conversations/%s/messages/%d/attachments/0/file?object_key=ignored", *env.conversation.ConversationID, env.message.ID), nil)
			request.Header.Set("Range", "bytes=2-5")
			if scenario == "unsatisfiable" {
				request.Header.Set("Range", "bytes=10-")
			}
			if scenario != "missing cookie" {
				request.AddCookie(env.cookie)
			}
			if member {
				request.Header.Set("X-Test-Member-Identity", "owner")
			}
			if scenario == "foreign member" {
				request.Header.Set("X-Test-Member-Identity", "other")
			}
			response := httptest.NewRecorder()
			env.router.ServeHTTP(response, request)
			expected := http.StatusPartialContent
			if scenario == "missing cookie" || scenario == "foreign member" || scenario == "internal message" {
				expected = http.StatusNotFound
			}
			if scenario == "unsatisfiable" {
				expected = http.StatusRequestedRangeNotSatisfiable
			}
			if scenario == "changed object" {
				expected = http.StatusServiceUnavailable
			}
			require.Equal(t, expected, response.Code, response.Body.String())
			require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
			require.Contains(t, response.Header().Values("Vary"), "Cookie")
			require.Empty(t, response.Header().Get("Location"))
			if expected == http.StatusNotFound {
				require.Empty(t, requests, "unauthorized requests must not even inspect storage metadata")
				require.Empty(t, response.Body.String())
				return
			}
			head := <-requests
			require.Equal(t, http.MethodHead, head.method)
			require.Equal(t, "/private-media/"+key, head.path)
			require.Empty(t, head.cookie)
			if expected == http.StatusRequestedRangeNotSatisfiable {
				require.Empty(t, requests)
				return
			}
			read := <-requests
			require.Equal(t, "bytes=2-5", read.byteRange)
			require.Equal(t, `"pinned-video-version"`, read.ifMatch)
			require.Equal(t, head.path, read.path)
			require.Empty(t, read.cookie)
			require.Empty(t, requests)
			if expected == http.StatusPartialContent {
				require.Equal(t, "2345", response.Body.String())
				require.Equal(t, strconv.Itoa(len(response.Body.Bytes())), response.Header().Get("Content-Length"))
				require.Equal(t, "bytes 2-5/10", response.Header().Get("Content-Range"))
			} else {
				require.Empty(t, response.Body.String())
			}
		})
	}
}

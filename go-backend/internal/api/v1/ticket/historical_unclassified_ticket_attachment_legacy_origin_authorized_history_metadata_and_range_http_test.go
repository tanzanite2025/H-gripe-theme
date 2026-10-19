package ticket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHistoricalUnclassifiedTicketAttachmentLegacyOriginHistoryAndRangeRequireOwnerCookie(t *testing.T) {
	for _, member := range []bool{false, true} {
		t.Run(map[bool]string{false: "visitor", true: "member"}[member], func(t *testing.T) {
			env := createCustomerServicePrivateAttachmentHTTPTestEnvironment(t, member)
			provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
			require.NoError(t, err)
			originalURL, err := provider.UploadFromReaderWithPrefix(t.Context(), bytes.NewReader(env.payload), "historical.png", "2026/08/20")
			require.NoError(t, err)
			key, err := provider.ObjectKey(originalURL)
			require.NoError(t, err)
			oldURL := "http://localhost:9200/uploads/" + key
			require.NoError(t, env.db.Model(&env.asset).Updates(map[string]any{"url": oldURL, "storage_key": key, "visibility": "public"}).Error)
			thumbnailURL, err := provider.UploadFromReaderWithPrefix(t.Context(), bytes.NewReader(env.payload), "thumbnail.png", "media-derivatives/historical")
			require.NoError(t, err)
			thumbnailKey, err := provider.ObjectKey(thumbnailURL)
			require.NoError(t, err)
			oldThumbnailURL := "http://localhost:9200/uploads/" + thumbnailKey
			require.NoError(t, env.db.Model(&media.MediaAssetDerivative{}).Where("media_asset_id = ?", env.asset.ID).Updates(map[string]any{"url": oldThumbnailURL, "storage_key": thumbnailKey}).Error)
			oldStorefrontURL := "http://localhost:9199/uploads/" + key
			attachments, err := json.Marshal([]string{oldStorefrontURL})
			require.NoError(t, err)
			metadata, err := json.Marshal(map[string]any{"url": oldStorefrontURL, "thumbnail": oldThumbnailURL, "nested": []string{"http://localhost:9199/uploads/" + thumbnailKey + "?v=1"}})
			require.NoError(t, err)
			require.NoError(t, env.db.Model(&env.message).Updates(map[string]any{"attachments": string(attachments), "metadata": string(metadata)}).Error)
			assets := service.NewMediaService(repository.NewMediaRepository(env.db), provider, nil, "https://shop.example.test", 20<<30)
			require.NoError(t, assets.ConfigureReviewedHistoricalCustomerServiceAttachmentOrigins("http://localhost:9199"))
			handler := NewHandler(env.handler.ticketService, Options{MediaService: assets, VisitorSecret: "test-secret"})
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if member && c.GetHeader("X-Test-Member-Identity") == "owner" {
					c.Set("user_id", *env.conversation.CustomerUserID)
				}
				c.Next()
			})
			router.GET("/api/v1/customer-service/messages/:conversation_id", handler.GetPublicCustomerServiceMessages)
			router.GET("/api/v1/customer-service/conversations/:conversation_id/messages/:message_id/attachments/:attachment_index/file", handler.ServePublicCustomerServiceMessageAttachmentFile)
			historyRequest := httptest.NewRequest(http.MethodGet, "/api/v1/customer-service/messages/"+*env.conversation.ConversationID, nil)
			historyRequest.AddCookie(env.cookie)
			historyRequest.Header.Set("X-Test-Member-Identity", "owner")
			history := httptest.NewRecorder()
			router.ServeHTTP(history, historyRequest)
			require.Equal(t, http.StatusOK, history.Code, history.Body.String())
			require.NotContains(t, history.Body.String(), "localhost:9200")
			require.NotContains(t, history.Body.String(), "localhost:9199")
			require.NotContains(t, history.Body.String(), key)
			require.NotContains(t, history.Body.String(), thumbnailKey)
			require.Contains(t, history.Body.String(), "/attachments/0/file?preset=thumbnail")
			path := fmt.Sprintf("/api/v1/customer-service/conversations/%s/messages/%d/attachments/0/file", *env.conversation.ConversationID, env.message.ID)
			for _, preset := range []string{"", "?preset=thumbnail"} {
				for _, authorized := range []bool{true, false} {
					request := httptest.NewRequest(http.MethodGet, path+preset, nil)
					request.Header.Set("Range", "bytes=0-9")
					if authorized {
						request.AddCookie(env.cookie)
						request.Header.Set("X-Test-Member-Identity", "owner")
					}
					response := httptest.NewRecorder()
					router.ServeHTTP(response, request)
					require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
					require.Empty(t, response.Header().Get("Location"))
					if authorized {
						require.Equal(t, http.StatusPartialContent, response.Code, response.Body.String())
						require.Equal(t, env.payload[:10], response.Body.Bytes())
						require.Contains(t, strings.Join(response.Header().Values("Vary"), ","), "Cookie")
					} else {
						require.NotEqual(t, http.StatusOK, response.Code)
						require.NotEqual(t, http.StatusPartialContent, response.Code)
					}
				}
			}
		})
	}
}

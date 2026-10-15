package ticket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	ticketdomain "commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

type customerServicePrivateAttachmentHTTPTestEnvironment struct {
	db           *gorm.DB
	handler      *Handler
	router       *gin.Engine
	conversation ticketdomain.Ticket
	message      ticketdomain.TicketMessage
	asset        media.MediaAsset
	cookie       *http.Cookie
	payload      []byte
}

func createCustomerServicePrivateAttachmentHTTPTestEnvironment(t *testing.T, member bool) customerServicePrivateAttachmentHTTPTestEnvironment {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, originalHandler := newPublicCustomerServiceHandlerTestEnv(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}))
	store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(),
		BaseURL: "https://shop.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), store, nil, "https://shop.example.test", 20<<30)
	handler := NewHandler(originalHandler.ticketService, Options{MediaService: mediaService, VisitorSecret: "test-secret"})
	sessionID := uuid.NewString()
	cookie := &http.Cookie{Name: customerServiceVisitorCookie, Value: handler.signVisitorSessionID(sessionID)}
	var customerUserID *uint
	if member {
		customer := seedCustomerServiceAttachmentIdentityMember(t, db)
		customerUserID = &customer.ID
	}
	conversation := seedCustomerServiceAttachmentIdentityConversation(t, db, customerUserID, visitorcookie.CustomerServiceVisitorHash(sessionID))
	var payload bytes.Buffer
	require.NoError(t, png.Encode(&payload, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	originalURL, err := store.(storage.CacheControlledReaderUploader).UploadFromReaderWithPrefixAndCacheControl(t.Context(), bytes.NewReader(payload.Bytes()),
		"customer-photo.png", storage.CustomerServiceConversationAttachmentOriginalsPrefix, "private, no-store")
	require.NoError(t, err)
	key, err := store.ObjectKey(originalURL)
	require.NoError(t, err)
	asset := media.MediaAsset{Filename: filepath.Base(key), OriginalFilename: "customer-photo.png", URL: originalURL, StorageKey: key,
		MimeType: "image/png", MediaType: "image", Status: "active", Visibility: "private", Size: int64(payload.Len())}
	require.NoError(t, db.Create(&asset).Error)
	derivativeURL, err := store.(storage.CacheControlledReaderUploader).UploadFromReaderWithPrefixAndCacheControl(t.Context(), bytes.NewReader(payload.Bytes()),
		"thumbnail.png", fmt.Sprintf("%s/%d/thumbnail/v1", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID), "private, no-store")
	require.NoError(t, err)
	derivativeKey, err := store.ObjectKey(derivativeURL)
	require.NoError(t, err)
	require.NoError(t, db.Create(&media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", PresetVersion: 1,
		URL: derivativeURL, StorageKey: derivativeKey, MimeType: "image/png", Size: int64(payload.Len()), Width: 2, Height: 2}).Error)
	attachments, err := json.Marshal([]string{originalURL})
	require.NoError(t, err)
	message := ticketdomain.TicketMessage{TicketID: conversation.ID, Content: "Photo", MessageType: "image", Attachments: string(attachments)}
	require.NoError(t, db.Create(&message).Error)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Claims represent successful OptionalAuthMiddleware validation. Identity
		// is never taken from a query or from the attachment's asset record.
		if c.GetHeader("X-Test-Member-Identity") == "owner" && customerUserID != nil {
			c.Set("user_id", *customerUserID)
		}
		if c.GetHeader("X-Test-Member-Identity") == "other" {
			c.Set("user_id", uint(999))
		}
		c.Header("Vary", "Origin")
		c.Next()
	})
	router.GET("/api/v1/customer-service/conversations/:conversation_id/messages/:message_id/attachments/:attachment_index/file", handler.ServePublicCustomerServiceMessageAttachmentFile)
	router.GET("/api/v1/customer-service/messages/:conversation_id", handler.GetPublicCustomerServiceMessages)
	return customerServicePrivateAttachmentHTTPTestEnvironment{db: db, handler: handler, router: router, conversation: conversation, message: message, asset: asset, cookie: cookie, payload: payload.Bytes()}
}

func TestPublicCustomerServicePrivateAttachmentHistoryProvidesCookieAuthorizedOriginalAndDerivativeLinks(t *testing.T) {
	for _, member := range []bool{false, true} {
		t.Run(map[bool]string{false: "visitor", true: "member"}[member], func(t *testing.T) {
			env := createCustomerServicePrivateAttachmentHTTPTestEnvironment(t, member)
			historyRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/customer-service/messages/"+*env.conversation.ConversationID, nil)
			historyRequest.AddCookie(env.cookie)
			if member {
				historyRequest.Header.Set("X-Test-Member-Identity", "owner")
			}
			historyRecorder := httptest.NewRecorder()
			env.router.ServeHTTP(historyRecorder, historyRequest)
			require.Equal(t, http.StatusOK, historyRecorder.Code, historyRecorder.Body.String())
			require.Equal(t, "private, no-store", historyRecorder.Header().Get("Cache-Control"))
			require.Contains(t, historyRecorder.Header().Values("Vary"), "Cookie")
			require.NotContains(t, historyRecorder.Body.String(), "private.example.test")
			require.NotContains(t, historyRecorder.Body.String(), env.asset.StorageKey)
			var history struct {
				Data []struct {
					Attachments     []string
					AttachmentMedia []service.CustomerServiceMessageAttachmentAccessDescription `json:"attachment_media"`
				}
			}
			require.NoError(t, json.Unmarshal(historyRecorder.Body.Bytes(), &history))
			require.Len(t, history.Data, 1)
			require.Len(t, history.Data[0].AttachmentMedia, 1)
			links := []string{history.Data[0].Attachments[0], history.Data[0].AttachmentMedia[0].Derivatives[0].AccessURL}
			for _, link := range links {
				request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, link, nil)
				request.AddCookie(env.cookie)
				if member {
					request.Header.Set("X-Test-Member-Identity", "owner")
				}
				recorder := httptest.NewRecorder()
				env.router.ServeHTTP(recorder, request)
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Equal(t, env.payload, recorder.Body.Bytes())
				require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
				require.Contains(t, recorder.Header().Values("Vary"), "Cookie")
				require.Contains(t, recorder.Header().Values("Vary"), "Origin")
				require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
				require.Empty(t, recorder.Header().Get("Location"))
			}
		})
	}
}

func TestPublicCustomerServicePrivateAttachmentFileRejectsMissingTamperedAndForeignCookies(t *testing.T) {
	for _, scenario := range []string{"missing cookie", "tampered cookie", "other visitor", "member without authentication", "other member", "foreign message", "internal message"} {
		t.Run(scenario, func(t *testing.T) {
			member := strings.Contains(scenario, "member")
			env := createCustomerServicePrivateAttachmentHTTPTestEnvironment(t, member)
			messageID := env.message.ID
			if scenario == "foreign message" {
				messageID += 100
			}
			if scenario == "internal message" {
				require.NoError(t, env.db.Model(&env.message).Update("is_internal", true).Error)
			}
			path := fmt.Sprintf("/api/v1/customer-service/conversations/%s/messages/%d/attachments/0/file?preset=thumbnail", *env.conversation.ConversationID, messageID)
			request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
			cookie := *env.cookie
			if scenario == "tampered cookie" {
				cookie.Value += "tampered"
			}
			if scenario == "other visitor" {
				cookie.Value = env.handler.signVisitorSessionID(uuid.NewString())
			}
			if scenario != "missing cookie" {
				request.AddCookie(&cookie)
			}
			if scenario == "other member" {
				request.Header.Set("X-Test-Member-Identity", "other")
			}
			recorder := httptest.NewRecorder()
			env.router.ServeHTTP(recorder, request)
			require.Equal(t, http.StatusNotFound, recorder.Code)
			require.Empty(t, recorder.Body.String())
			require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
		})
	}
}

func TestPublicCustomerServicePrivateAttachmentRangeResponsesRemainPrivateIncludingErrors(t *testing.T) {
	env := createCustomerServicePrivateAttachmentHTTPTestEnvironment(t, false)
	for _, rangeValue := range []string{"bytes=0-7", "bytes=" + strconv.Itoa(len(env.payload)+100) + "-"} {
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("/api/v1/customer-service/conversations/%s/messages/%d/attachments/0/file", *env.conversation.ConversationID, env.message.ID), nil)
		request.Header.Set("Range", rangeValue)
		request.AddCookie(env.cookie)
		recorder := httptest.NewRecorder()
		env.router.ServeHTTP(recorder, request)
		if rangeValue == "bytes=0-7" {
			require.Equal(t, http.StatusPartialContent, recorder.Code)
			require.Equal(t, env.payload[:8], recorder.Body.Bytes())
		} else {
			require.Equal(t, http.StatusRequestedRangeNotSatisfiable, recorder.Code)
		}
		require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
	}
}

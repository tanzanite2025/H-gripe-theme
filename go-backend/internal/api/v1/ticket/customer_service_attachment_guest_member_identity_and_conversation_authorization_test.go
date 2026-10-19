package ticket

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	ticketdomain "commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPublicCustomerServiceGuestAttachmentUploadThenMessagePreservesVerifiedIdentityAndReference(t *testing.T) {
	db, handler, _, router, _ := createCustomerServiceAttachmentIdentityTestEnvironment(t, 0)
	sessionID := uuid.NewString()
	visitorHash := visitorcookie.CustomerServiceVisitorHash(sessionID)
	conversation := seedCustomerServiceAttachmentIdentityConversation(t, db, nil, visitorHash)
	cookie := &http.Cookie{Name: customerServiceVisitorCookie, Value: handler.signVisitorSessionID(sessionID)}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, createCustomerServiceAttachmentMultipartRequest(t, *conversation.ConversationID, cookie))
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	var response struct {
		Asset struct {
			ID     uint
			URL    string
			Source string
		}
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotZero(t, response.Asset.ID)
	require.Equal(t, "camera", response.Asset.Source)
	var asset media.MediaAsset
	require.NoError(t, db.First(&asset, response.Asset.ID).Error)
	require.Equal(t, media.MediaUploadActorVisitor, asset.UploaderActorType)
	require.Equal(t, visitorHash, asset.UploaderVisitorSessionHash)
	require.Zero(t, asset.UploaderID)
	require.NotContains(t, recorder.Body.String(), visitorHash)
	require.NotContains(t, asset.CopyrightClaimJSON, cookie.Value)

	payload, err := json.Marshal(gin.H{
		"conversation_id": *conversation.ConversationID, "message": "Photo of the wheel", "sender_name": "Guest",
		"message_type": "image", "attachments": []string{response.Asset.URL},
	})
	require.NoError(t, err)
	messageRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/messages", bytes.NewReader(payload))
	messageRequest.Header.Set("Content-Type", "application/json")
	messageRequest.AddCookie(cookie)
	messageRecorder := httptest.NewRecorder()
	router.ServeHTTP(messageRecorder, messageRequest)
	require.Equal(t, http.StatusOK, messageRecorder.Code, messageRecorder.Body.String())
	var message ticketdomain.TicketMessage
	require.NoError(t, db.Where("ticket_id = ? AND is_staff = ?", conversation.ID, false).First(&message).Error)
	require.Equal(t, "image", message.MessageType)
	require.Contains(t, message.Attachments, asset.StorageKey)
	references, err := repository.NewMediaRepository(db).FindAssetReferences(&asset)
	require.NoError(t, err)
	require.NotEmpty(t, references)
}

func TestPublicCustomerServiceAttachmentRejectsUnauthorizedOwnersWithoutWritingObjectsOrAssets(t *testing.T) {
	for _, scenario := range []string{"missing cookie", "tampered cookie", "other guest", "guest accessing member", "member accessing guest", "member accessing other member", "deleted conversation", "noncustomer-service ticket"} {
		t.Run(scenario, func(t *testing.T) {
			db, handler, _, _, uploadRoot := createCustomerServiceAttachmentIdentityTestEnvironment(t, 0)
			ownerSessionID := uuid.NewString()
			ownerHash := visitorcookie.CustomerServiceVisitorHash(ownerSessionID)
			var customerUserID *uint
			if scenario == "guest accessing member" || scenario == "member accessing other member" {
				customer := seedCustomerServiceAttachmentIdentityMember(t, db)
				customerUserID = &customer.ID
			}
			conversation := seedCustomerServiceAttachmentIdentityConversation(t, db, customerUserID, ownerHash)
			if scenario == "deleted conversation" {
				require.NoError(t, db.Delete(&conversation).Error)
			}
			if scenario == "noncustomer-service ticket" {
				require.NoError(t, db.Model(&conversation).Update("category", "order").Error)
			}
			cookie := &http.Cookie{Name: customerServiceVisitorCookie, Value: handler.signVisitorSessionID(ownerSessionID)}
			if scenario == "missing cookie" {
				cookie = nil
			}
			if scenario == "tampered cookie" {
				cookie.Value += "tampered"
			}
			if scenario == "other guest" {
				cookie.Value = handler.signVisitorSessionID(uuid.NewString())
			}
			router := gin.New()
			if strings.HasPrefix(scenario, "member accessing") {
				router.Use(func(c *gin.Context) { c.Set("user_id", uint(999)); c.Next() })
			}
			router.POST("/attachments", handler.UploadPublicCustomerServiceAttachment)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, createCustomerServiceAttachmentMultipartRequest(t, *conversation.ConversationID, cookie))
			require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
			_, total, err := repository.NewMediaRepository(db).ListAssets(repository.MediaAssetFilter{})
			require.NoError(t, err)
			require.Zero(t, total)
			require.Empty(t, collectCustomerServiceAttachmentStoredFilePaths(t, uploadRoot))
		})
	}
}

func TestPublicCustomerServiceMemberAttachmentUsesAuthenticatedAccountAndReturnsQuotaError(t *testing.T) {
	db, handler, _, _, _ := createCustomerServiceAttachmentIdentityTestEnvironment(t, 0)
	customer := seedCustomerServiceAttachmentIdentityMember(t, db)
	conversation := seedCustomerServiceAttachmentIdentityConversation(t, db, &customer.ID, "")
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", customer.ID); c.Next() })
	router.POST("/attachments", handler.UploadPublicCustomerServiceAttachment)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, createCustomerServiceAttachmentMultipartRequest(t, *conversation.ConversationID, nil))
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	var asset media.MediaAsset
	require.NoError(t, db.First(&asset).Error)
	require.Equal(t, media.MediaUploadActorCustomer, asset.UploaderActorType)
	require.Equal(t, customer.ID, asset.UploaderID)
	require.Empty(t, asset.UploaderVisitorSessionHash)
	// A service configured with a depleted account budget still authenticates
	// the member and returns the existing account-quota response contract.
	storageService := createCustomerServiceAttachmentLocalStorage(t, t.TempDir())
	limitedMediaService := service.NewMediaService(repository.NewMediaRepository(db), storageService, nil, "https://shop.example.test", asset.Size)
	handler.conversationAttachmentUploads = service.NewCustomerServiceConversationAttachmentUploadService(handler.ticketService, limitedMediaService)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, createCustomerServiceAttachmentMultipartRequest(t, *conversation.ConversationID, nil))
	require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), "account_storage_quota_exceeded")
}

func TestPublicCustomerServiceGuestAttachmentQuotaReturnsSpecificErrorBeforeWritingFile(t *testing.T) {
	db, handler, _, router, uploadRoot := createCustomerServiceAttachmentIdentityTestEnvironment(t, 1)
	sessionID := uuid.NewString()
	conversation := seedCustomerServiceAttachmentIdentityConversation(t, db, nil, visitorcookie.CustomerServiceVisitorHash(sessionID))
	cookie := &http.Cookie{Name: customerServiceVisitorCookie, Value: handler.signVisitorSessionID(sessionID)}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, createCustomerServiceAttachmentMultipartRequest(t, *conversation.ConversationID, cookie))
	require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), "visitor_session_storage_quota_exceeded")
	require.Empty(t, collectCustomerServiceAttachmentStoredFilePaths(t, uploadRoot))
}

func createCustomerServiceAttachmentIdentityTestEnvironment(t *testing.T, visitorQuotaBytes int64) (*gorm.DB, *Handler, *service.MediaService, *gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, originalHandler := newPublicCustomerServiceHandlerTestEnv(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}))
	originalHandler.ticketService.ConfigureCustomerServiceRealtimeOutbox(repository.NewOutboxRepository(db))
	uploadRoot := t.TempDir()
	if visitorQuotaBytes == 0 {
		visitorQuotaBytes = service.DefaultMediaVisitorSessionStorageQuotaBytes
	}
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), createCustomerServiceAttachmentLocalStorage(t, uploadRoot), nil, "https://shop.example.test", 20<<30,
		service.MediaUploadActorQuotaOptions{VisitorSessionStorageQuotaBytes: visitorQuotaBytes})
	mediaService.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	handler := NewHandler(originalHandler.ticketService, Options{MediaService: mediaService, VisitorSecret: "test-secret"})
	router := gin.New()
	router.POST("/attachments", handler.UploadPublicCustomerServiceAttachment)
	router.POST("/messages", handler.SendPublicCustomerServiceMessage)
	return db, handler, mediaService, router, uploadRoot
}

func createCustomerServiceAttachmentLocalStorage(t *testing.T, uploadRoot string) storage.StorageService {
	t.Helper()
	storageService, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: uploadRoot, PrivateLocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	return storageService
}

func seedCustomerServiceAttachmentIdentityMember(t *testing.T, db *gorm.DB) user.User {
	t.Helper()
	customer := user.User{Email: "customer@example.test", Username: "customer", Password: "test-password", Role: "customer", Status: "active"}
	require.NoError(t, db.Create(&customer).Error)
	return customer
}

func seedCustomerServiceAttachmentIdentityConversation(t *testing.T, db *gorm.DB, customerUserID *uint, visitorHash string) ticketdomain.Ticket {
	t.Helper()
	support := seedPublicCustomerServiceSupportUser(t, db)
	conversationID := uuid.NewString()
	conversation := ticketdomain.Ticket{UserID: support.ID, AssignedTo: support.ID, ConversationID: &conversationID,
		CustomerUserID: customerUserID, VisitorSessionHash: visitorHash, Category: "customer_service", Subject: "Attachment ownership", Status: "in_progress"}
	require.NoError(t, db.Create(&conversation).Error)
	return conversation
}

func createCustomerServiceAttachmentMultipartRequest(t *testing.T, conversationID string, cookie *http.Cookie) *http.Request {
	t.Helper()
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	require.NoError(t, writer.WriteField("conversation_id", conversationID))
	require.NoError(t, writer.WriteField("source", "camera"))
	// Forged identity fields must not override the authenticated conversation.
	require.NoError(t, writer.WriteField("uploader_id", "123456"))
	require.NoError(t, writer.WriteField("uploader_actor_type", "administrator"))
	require.NoError(t, writer.WriteField("uploader_visitor_session_hash", strings.Repeat("f", 64)))
	file, err := writer.CreateFormFile("file", "customer-attachment.png")
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	require.NoError(t, writer.Close())
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/attachments", &payload)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if cookie != nil {
		request.AddCookie(cookie)
	}
	return request
}

func collectCustomerServiceAttachmentStoredFilePaths(t *testing.T, uploadRoot string) []string {
	t.Helper()
	var storedFilePaths []string
	require.NoError(t, filepath.WalkDir(uploadRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			storedFilePaths = append(storedFilePaths, path)
		}
		return nil
	}))
	return storedFilePaths
}

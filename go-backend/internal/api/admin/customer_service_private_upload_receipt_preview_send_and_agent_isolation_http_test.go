package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestAdminCustomerServicePrivateUploadPreviewAndSendEnforceUploaderAssignmentAndTicketPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&user.User{}, &ticket.Ticket{}, &ticket.TicketMessage{}, &ticket.CustomerServiceInboxState{},
		&media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.MediaUploadSession{}, &ticket.TicketMessageMediaAttachment{}, &outbox.Event{}))
	agent := user.User{Email: "private-upload-support@example.test", Username: "private-upload-support", Password: "test-password", Role: "support", Status: "active"}
	require.NoError(t, db.Create(&agent).Error)
	conversationID := uuid.NewString()
	conversation := ticket.Ticket{TicketNumber: uuid.NewString(), UserID: agent.ID, AssignedTo: agent.ID, ConversationID: &conversationID,
		Category: "customer_service", Subject: "Private upload", Status: "in_progress"}
	require.NoError(t, db.Create(&conversation).Error)
	store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(),
		BaseURL: "https://shop.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), store, nil, "https://shop.example.test", 20<<30)
	tickets := service.NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	tickets.ConfigureCustomerServiceRealtimeOutbox(repository.NewOutboxRepository(db))
	handler := NewTicketHandler(tickets, nil, nil, nil, mediaService)
	createRouter := func(agentID uint, role string) *gin.Engine {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", agentID)
			c.Set("role", role)
			c.Set("user_role", role)
			c.Next()
		})
		router.POST("/api/admin/customer-service/conversations/:id/attachment-uploads", middleware.RequirePermission(auth.PermTicketEdit), handler.UploadAgentCustomerServiceConversationAttachment)
		router.GET("/api/admin/customer-service/conversations/:id/attachment-uploads/:upload_session_id/file", middleware.RequirePermission(auth.PermTicketView), handler.ServeAgentCustomerServicePendingAttachmentUploadPreview)
		router.POST("/api/admin/customer-service/conversations/:id/messages", middleware.RequirePermission(auth.PermTicketEdit), handler.CreateCustomerServiceConversationMessage)
		router.GET("/api/admin/customer-service/conversations/:id/messages/:message_id/attachments/:attachment_index/file", middleware.RequirePermission(auth.PermTicketView), handler.ServeAgentCustomerServiceMessageAttachmentFile)
		return router
	}
	uploadPath := fmt.Sprintf("/api/admin/customer-service/conversations/%d/attachment-uploads", conversation.ID)
	photo := image.NewRGBA(image.Rect(0, 0, 32, 16))
	draw.Draw(photo, photo.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	var photoBytes, multipartBytes bytes.Buffer
	require.NoError(t, png.Encode(&photoBytes, photo))
	multipartWriter := multipart.NewWriter(&multipartBytes)
	part, err := multipartWriter.CreateFormFile("file", "agent-photo.png")
	require.NoError(t, err)
	_, err = part.Write(photoBytes.Bytes())
	require.NoError(t, err)
	require.NoError(t, multipartWriter.Close())
	uploadRequest := func() *http.Request {
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, uploadPath, bytes.NewReader(multipartBytes.Bytes()))
		request.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		return request
	}
	router := createRouter(agent.ID, "support")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, uploadRequest())
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	var compatiblePublicAsset media.MediaAsset
	require.NoError(t, db.First(&compatiblePublicAsset).Error)
	require.Equal(t, "public", compatiblePublicAsset.Visibility)
	require.Equal(t, media.MediaLibraryPartitionCustomerService, compatiblePublicAsset.LibraryPartition)
	require.Equal(t, media.CustomerServiceMediaPurposeConversationAttachment, compatiblePublicAsset.CustomerServiceMediaPurpose)
	require.Empty(t, compatiblePublicAsset.PendingCustomerServiceUploadSessionID)
	allowedPublicRead, err := service.NewPublicUploadAccessService(mediaService, nil).CanServePublicUpload(t.Context(), compatiblePublicAsset.StorageKey)
	require.NoError(t, err)
	require.True(t, allowedPublicRead)
	for _, privateUploadsEnabled := range []bool{false, true} {
		mediaService.ConfigureCustomerServicePrivateAttachmentUploads(privateUploadsEnabled)
		for _, unauthorized := range []struct {
			agentID uint
			role    string
			status  int
		}{
			{agent.ID, "editor", http.StatusForbidden}, {agent.ID + 100, "support", http.StatusForbidden},
		} {
			recorder = httptest.NewRecorder()
			createRouter(unauthorized.agentID, unauthorized.role).ServeHTTP(recorder, uploadRequest())
			require.Equal(t, unauthorized.status, recorder.Code, recorder.Body.String())
		}
	}
	// The disabled private rollout retains the existing public video flow.
	createAgentConversationVideoMultipartRequest := func() *http.Request {
		var videoMultipartBytes bytes.Buffer
		videoMultipartWriter := multipart.NewWriter(&videoMultipartBytes)
		videoPart, multipartErr := videoMultipartWriter.CreateFormFile("file", "conversation-video.mp4")
		require.NoError(t, multipartErr)
		_, multipartErr = videoPart.Write([]byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 0, 0, 'i', 's', 'o', 'm', 'm', 'p', '4', '2'})
		require.NoError(t, multipartErr)
		require.NoError(t, videoMultipartWriter.WriteField("media_type", "image")) // Actual file content controls type.
		require.NoError(t, videoMultipartWriter.Close())
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, uploadPath, bytes.NewReader(videoMultipartBytes.Bytes()))
		request.Header.Set("Content-Type", videoMultipartWriter.FormDataContentType())
		return request
	}
	mediaService.ConfigureCustomerServicePrivateAttachmentUploads(false)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, createAgentConversationVideoMultipartRequest())
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	var videoResponse struct{ Asset media.MediaAsset }
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &videoResponse))
	require.Equal(t, "video", videoResponse.Asset.MediaType)
	require.Equal(t, media.MediaLibraryPartitionCustomerService, videoResponse.Asset.LibraryPartition)
	require.Equal(t, media.CustomerServiceMediaPurposeConversationAttachment, videoResponse.Asset.CustomerServiceMediaPurpose)
	mediaService.ConfigureCustomerServicePrivateAttachmentUploads(true)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, createAgentConversationVideoMultipartRequest())
	require.Equal(t, http.StatusUnsupportedMediaType, recorder.Code, recorder.Body.String())
	var beforeUploads int64
	require.NoError(t, db.Model(&media.MediaUploadSession{}).Count(&beforeUploads).Error)
	require.Zero(t, beforeUploads)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, uploadRequest())
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
	require.NotContains(t, recorder.Body.String(), "private.example.test")
	require.NotContains(t, recorder.Body.String(), "customer-service/attachments/originals")
	var response struct {
		Asset struct {
			ID              uint
			URL             string
			UploadSessionID string `json:"upload_session_id"`
		}
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotEmpty(t, response.Asset.UploadSessionID)
	for _, preview := range []struct {
		agentID uint
		role    string
		status  int
	}{
		{agent.ID, "support", http.StatusOK}, {agent.ID + 100, "manager", http.StatusNotFound}, {agent.ID, "editor", http.StatusForbidden},
	} {
		recorder = httptest.NewRecorder()
		createRouter(preview.agentID, preview.role).ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, response.Asset.URL, nil))
		require.Equal(t, preview.status, recorder.Code, recorder.Body.String())
		if preview.status == http.StatusOK {
			require.Equal(t, photoBytes.Bytes(), recorder.Body.Bytes())
		}
	}
	// Transfer removes the original uploader's current conversation permission.
	require.NoError(t, db.Model(&conversation).Update("assigned_to", agent.ID+100).Error)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, response.Asset.URL, nil))
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.NoError(t, db.Model(&conversation).Update("assigned_to", agent.ID).Error)
	payload, err := json.Marshal(gin.H{"message": "Agent photo", "message_type": "image", "attachments": []string{response.Asset.URL}})
	require.NoError(t, err)
	messagePath := fmt.Sprintf("/api/admin/customer-service/conversations/%d/messages", conversation.ID)
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, messagePath, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	var message ticket.TicketMessage
	require.NoError(t, db.Where("ticket_id = ?", conversation.ID).First(&message).Error)
	require.Equal(t, ticket.TicketMessageMediaAssetBindings, message.AttachmentReferenceFormat)
	require.Empty(t, message.Attachments)
	var session media.MediaUploadSession
	require.NoError(t, db.First(&session, "id = ?", response.Asset.UploadSessionID).Error)
	require.Equal(t, "bound", session.Status)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, response.Asset.URL, nil))
	require.Equal(t, http.StatusNotFound, recorder.Code)
	recorder = httptest.NewRecorder()
	createRouter(agent.ID+100, "manager").ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/%d/attachments/0/file", messagePath, message.ID), nil))
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, photoBytes.Bytes(), recorder.Body.Bytes())
}

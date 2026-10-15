package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
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
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestAdminCustomerServicePrivateAttachmentFileEnforcesTicketPermissionAndCurrentAssignment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&user.User{}, &ticket.Ticket{}, &ticket.TicketMessage{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}))
	store, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal,
		LocalPath: t.TempDir(), PrivateLocalPath: t.TempDir(), BaseURL: "https://shop.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), store, nil, "https://shop.example.test", 20<<30)
	tickets := service.NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	handler := NewTicketHandler(tickets, nil, nil, nil, mediaService)
	originalURL, err := store.(storage.CacheControlledReaderUploader).UploadFromReaderWithPrefixAndCacheControl(t.Context(), bytes.NewReader([]byte("private-photo")),
		"staff-photo.png", storage.CustomerServiceConversationAttachmentOriginalsPrefix, "private, no-store")
	require.NoError(t, err)
	key, err := store.ObjectKey(originalURL)
	require.NoError(t, err)
	asset := media.MediaAsset{Filename: filepath.Base(key), OriginalFilename: "staff-photo.png", URL: originalURL, StorageKey: key,
		MediaType: "image", MimeType: "image/png", Visibility: "private", Status: "active", Size: 13}
	require.NoError(t, db.Create(&asset).Error)
	conversationID := uuid.NewString()
	conversation := ticket.Ticket{TicketNumber: "test-" + uuid.NewString(), UserID: 42, AssignedTo: 42,
		ConversationID: &conversationID, Category: "customer_service", Subject: "Staff access", Status: "in_progress"}
	require.NoError(t, db.Create(&conversation).Error)
	attachments, err := json.Marshal([]string{originalURL})
	require.NoError(t, err)
	message := ticket.TicketMessage{TicketID: conversation.ID, Content: "Attachment", MessageType: "image", Attachments: string(attachments)}
	require.NoError(t, db.Create(&message).Error)
	for _, scenario := range []struct {
		name, role     string
		agentID        uint
		expectedStatus int
	}{
		{name: "assigned support", role: "support", agentID: 42, expectedStatus: http.StatusOK},
		{name: "unassigned support", role: "support", agentID: 43, expectedStatus: http.StatusNotFound},
		{name: "manager", role: "manager", agentID: 43, expectedStatus: http.StatusOK},
		{name: "administrator", role: "admin", agentID: 43, expectedStatus: http.StatusOK},
		{name: "editor cannot read tickets", role: "editor", agentID: 42, expectedStatus: http.StatusForbidden},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", scenario.agentID)
				c.Set("role", scenario.role)
				c.Set("user_role", scenario.role)
				c.Next()
			})
			router.GET("/api/admin/customer-service/conversations/:id/messages/:message_id/attachments/:attachment_index/file",
				middleware.RequirePermission(auth.PermTicketView), handler.ServeAgentCustomerServiceMessageAttachmentFile)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("/api/admin/customer-service/conversations/%d/messages/%d/attachments/0/file", conversation.ID, message.ID), nil)
			router.ServeHTTP(recorder, request)
			require.Equal(t, scenario.expectedStatus, recorder.Code, recorder.Body.String())
			if recorder.Code == http.StatusOK {
				require.Equal(t, "private-photo", recorder.Body.String())
				require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
				require.Empty(t, recorder.Header().Get("Location"))
			}
		})
	}
	response := adminCustomerServiceMessageResponse(message, handler.conversationAttachmentAccess)
	encodedResponse, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotContains(t, string(encodedResponse), "private.example.test")
	require.True(t, strings.HasPrefix(response["attachment_url"].(string), "/api/admin/customer-service/conversations/"))
}

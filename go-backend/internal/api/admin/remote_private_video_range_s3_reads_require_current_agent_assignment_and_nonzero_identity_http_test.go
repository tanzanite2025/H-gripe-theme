package admin

import (
	"encoding/json"
	"fmt"
	"io"
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
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestRemotePrivateVideoRangeS3ReadsRequireCurrentAgentAssignmentPermissionAndNonzeroIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&user.User{}, &ticket.Ticket{}, &ticket.TicketMessage{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}))
	requests := make(chan *http.Request, 8)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r
		w.Header().Set("ETag", `"staff-video-version"`)
		w.Header().Set("Content-Length", "10")
		if r.Method == http.MethodHead {
			return
		}
		w.Header().Set("Content-Length", "4")
		w.Header().Set("Content-Range", "bytes 2-5/10")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, "2345")
	}))
	t.Cleanup(server.Close)
	store, err := storage.NewS3Storage(&storage.Config{Type: storage.StorageTypeS3, Bucket: "public-media", PrivateBucket: "private-media", Region: "us-east-1", Endpoint: server.URL,
		AccessKeyID: "synthetic-range-key", SecretAccessKey: "synthetic-range-secret", BaseURL: "https://shop.example.test", PrivateBaseURL: "https://private.example.test"})
	require.NoError(t, err)
	key := storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/synthetic-staff-video.mp4"
	originalURL := store.GetURL(key)
	asset := media.MediaAsset{Filename: "video.mp4", OriginalFilename: "video.mp4", URL: originalURL, StorageKey: key,
		MediaType: "video", MimeType: "video/mp4", Visibility: "private", Status: "active", Size: 10}
	require.NoError(t, db.Create(&asset).Error)
	conversationID := uuid.NewString()
	conversation := ticket.Ticket{TicketNumber: "range-" + uuid.NewString(), UserID: 42, AssignedTo: 42,
		ConversationID: &conversationID, Category: "customer_service", Subject: "Remote staff video", Status: "in_progress"}
	require.NoError(t, db.Create(&conversation).Error)
	attachments, err := json.Marshal([]string{originalURL})
	require.NoError(t, err)
	message := ticket.TicketMessage{TicketID: conversation.ID, Content: "Private staff attachment", MessageType: "video", Attachments: string(attachments), IsInternal: true}
	require.NoError(t, db.Create(&message).Error)
	mediaService := service.NewMediaService(repository.NewMediaRepository(db), store, nil, "https://shop.example.test", 20<<30)
	tickets := service.NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	handler := NewTicketHandler(tickets, nil, nil, nil, mediaService)
	for _, scenario := range []struct {
		name, role          string
		agentID, assignedTo uint
		expectedStatus      int
	}{
		{"assigned support", "support", 42, 42, 206},
		{"assignment revoked before next range", "support", 42, 43, 404},
		{"newly assigned support", "support", 43, 43, 206},
		{"unassigned support cannot elevate query scope", "support", 44, 43, 404},
		{"manager", "manager", 44, 43, 206},
		{"administrator", "admin", 44, 43, 206},
		{"zero identity administrator", "admin", 0, 43, 404},
		{"editor permission denied", "editor", 43, 43, 403},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			require.NoError(t, db.Model(&conversation).Update("assigned_to", scenario.assignedTo).Error)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", scenario.agentID)
				c.Set("role", scenario.role)
				c.Set("user_role", scenario.role)
				c.Next()
			})
			router.GET("/api/admin/customer-service/conversations/:id/messages/:message_id/attachments/:attachment_index/file",
				middleware.RequirePermission(auth.PermTicketView), handler.ServeAgentCustomerServiceMessageAttachmentFile)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("/api/admin/customer-service/conversations/%d/messages/%d/attachments/0/file?can_view_all=true&object_key=ignored", conversation.ID, message.ID), nil)
			request.Header.Set("Range", "bytes=2-5")
			if scenario.expectedStatus == http.StatusNotFound {
				request.Header.Set("If-None-Match", `"staff-video-version"`)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, scenario.expectedStatus, response.Code, response.Body.String())
			require.Empty(t, response.Header().Get("Location"))
			if scenario.expectedStatus != http.StatusForbidden {
				require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
				require.Contains(t, response.Header().Values("Vary"), "Cookie")
			}
			if scenario.expectedStatus != http.StatusPartialContent {
				require.Empty(t, requests, "denied requests must not inspect remote metadata or evaluate object validators")
				return
			}
			require.Equal(t, "2345", response.Body.String())
			require.Len(t, requests, 2)
			head, read := <-requests, <-requests
			require.Equal(t, http.MethodHead, head.Method)
			require.Equal(t, http.MethodGet, read.Method)
			require.Equal(t, "/private-media/"+key, read.URL.Path)
			require.Equal(t, head.URL.Path, read.URL.Path)
			require.Equal(t, "bytes=2-5", read.Header.Get("Range"))
			require.Equal(t, `"staff-video-version"`, read.Header.Get("If-Match"))
			require.Empty(t, requests)
		})
	}
}

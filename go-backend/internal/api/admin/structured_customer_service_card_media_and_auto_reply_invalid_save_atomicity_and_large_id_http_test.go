package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func createStructuredAdminCustomerServiceMediaHTTPTestEnvironment(t *testing.T) (*gorm.DB, *gin.Engine, *ticket.AutoReplyRule, ticket.Ticket) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&user.User{}, &ticket.Ticket{}, &ticket.TicketMessage{}, &ticket.AutoReplyRule{}, &ticket.CustomerServiceInboxState{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &outbox.Event{}))
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	assets := service.NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
	tickets := service.NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	tickets.ConfigureMediaService(assets)
	tickets.ConfigureCustomerServiceRealtimeOutbox(repository.NewOutboxRepository(db))
	agent := user.User{Email: "synthetic-structured-agent@example.test", Username: "structured-agent", Password: "synthetic", Role: "support", Status: "active"}
	require.NoError(t, db.Create(&agent).Error)
	conversation := ticket.Ticket{UserID: agent.ID, AssignedTo: agent.ID, Subject: "Synthetic structured media", Category: "customer_service", Status: "in_progress"}
	require.NoError(t, db.Create(&conversation).Error)
	rule, err := tickets.CreateAutoReplyRule(service.AutoReplyRuleInput{Type: "welcome", ReplyMessage: "Previous reply", MessageType: "text", IsActive: true, Locale: "en"})
	require.NoError(t, err)
	handler := NewTicketHandler(tickets, nil, nil, nil, assets)
	rules := NewAutoReplyHandler(tickets, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", agent.ID); c.Set("role", "support"); c.Next() })
	router.POST("/conversations/:id/messages", handler.CreateCustomerServiceConversationMessage)
	router.POST("/rules", rules.CreateRule)
	router.PUT("/rules/:id", rules.UpdateRule)
	return db, router, rule, conversation
}

func TestStructuredAdminCustomerServiceAndAutoReplyInvalidMediaReturnHTTP400AndRollback(t *testing.T) {
	db, router, previous, conversation := createStructuredAdminCustomerServiceMediaHTTPTestEnvironment(t)
	metadata := `{"thumbnail":"/uploads/site-logo/synthetic-sensitive-detail.png","url":"/products/card"}`
	for _, mutation := range []string{"message", "create-rule", "update-rule"} {
		t.Run(mutation, func(t *testing.T) {
			method, path := http.MethodPost, fmt.Sprintf("/conversations/%d/messages", conversation.ID)
			body := map[string]any{"message": "Must rollback", "message_type": "product", "metadata": json.RawMessage(metadata)}
			if mutation != "message" {
				path = "/rules"
				body = map[string]any{"type": "welcome", "reply_message": "Must rollback", "message_type": "product", "metadata": metadata, "is_active": true, "locale": "en"}
				if mutation == "update-rule" {
					method, path = http.MethodPut, fmt.Sprintf("/rules/%d", previous.ID)
				}
			}
			encoded, err := json.Marshal(body)
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(encoded))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			require.NotContains(t, response.Body.String(), "synthetic-sensitive-detail")
		})
	}
	var messages, events, rules int64
	require.NoError(t, db.Model(&ticket.TicketMessage{}).Count(&messages).Error)
	require.NoError(t, db.Model(&outbox.Event{}).Count(&events).Error)
	require.NoError(t, db.Model(&ticket.AutoReplyRule{}).Count(&rules).Error)
	require.Zero(t, messages)
	require.Zero(t, events)
	require.EqualValues(t, 1, rules)
	var stored ticket.AutoReplyRule
	require.NoError(t, db.First(&stored, previous.ID).Error)
	require.Equal(t, previous.ReplyMessage, stored.ReplyMessage)
	require.Equal(t, previous.Metadata, stored.Metadata)
}

func TestStructuredAdminCustomerServiceMediaCanonicalizationPreservesLargeIDsThroughRequestStorageAndResponse(t *testing.T) {
	db, router, _, conversation := createStructuredAdminCustomerServiceMediaHTTPTestEnvironment(t)
	asset := media.MediaAsset{Filename: "card.png", URL: "https://shop.example.test/uploads/content/agent-card.png", StorageKey: "content/agent-card.png", MediaType: "image", MimeType: "image/png", Status: "active", Visibility: "public", UploadOrigin: media.MediaUploadOriginAdministratorShared, ClassificationMethod: media.MediaClassificationRecordedUploadContext, LibraryPartition: media.MediaLibraryPartitionShared}
	require.NoError(t, db.Create(&asset).Error)
	metadata := json.RawMessage(`{"thumbnail":"/uploads/content/agent-card.png","url":"/products/card","product_id":9007199254740997}`)
	encoded, err := json.Marshal(map[string]any{"message": "Valid agent card", "message_type": "product", "metadata": metadata})
	require.NoError(t, err)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, fmt.Sprintf("/conversations/%d/messages", conversation.ID), bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "9007199254740997")
	var message ticket.TicketMessage
	require.NoError(t, db.First(&message, "ticket_id = ?", conversation.ID).Error)
	require.Contains(t, message.Metadata, asset.URL)
	require.Contains(t, message.Metadata, "9007199254740997")
}

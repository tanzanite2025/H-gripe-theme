package ticket

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	ticketdomain "commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestEnsurePublicCustomerServiceConversationTouchesVisitorTimezone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := newPublicCustomerServiceHandlerTestEnv(t)
	supportUser := seedPublicCustomerServiceSupportUser(t, db)

	router := gin.New()
	router.POST("/conversations", handler.EnsurePublicCustomerServiceConversation)

	recorder := httptest.NewRecorder()
	body := []byte(`{"agent_id":"` + itoaUint(supportUser.ID) + `","locale":"fr"}`)
	request := httptest.NewRequest(http.MethodPost, "/conversations", bytes.NewReader(body))
	request.Header.Set("X-Timezone", "America/Los_Angeles")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response struct {
		Success        bool   `json:"success"`
		ConversationID string `json:"conversation_id"`
		Data           struct {
			ConversationID string `json:"conversation_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.True(t, response.Success)
	assert.NotEmpty(t, response.ConversationID)
	assert.NotEmpty(t, response.Data.ConversationID)

	var profile visitor.Profile
	require.NoError(t, db.First(&profile).Error)
	assert.Equal(t, "fr", profile.Locale)
	assert.Equal(t, "request", profile.LocaleSource)
	assert.Equal(t, "America/Los_Angeles", profile.Timezone)
	assert.Equal(t, service.VisitorProfileActionCustomerService, profile.LastMeaningfulAction)
	assert.Equal(t, service.VisitorProfileQualityCustomerService, profile.ProfileQualityScore)
	assert.NotEmpty(t, profile.CustomerServiceVisitorHash)
}

func TestSendPublicCustomerServiceMessageRetryReturnsOriginalResponseWithoutRepeatingEffects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := newPublicCustomerServiceHandlerTestEnv(t)
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	handler.ticketService.ConfigureCustomerServiceRealtimeOutbox(repository.NewOutboxRepository(db))
	supportUser := seedPublicCustomerServiceSupportUser(t, db)

	router := gin.New()
	router.POST("/conversations", handler.EnsurePublicCustomerServiceConversation)
	router.POST("/messages", handler.SendPublicCustomerServiceMessage)
	router.GET("/messages/:conversation_id", handler.GetPublicCustomerServiceMessages)
	createRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/conversations", bytes.NewReader([]byte(`{"agent_id":"`+itoaUint(supportUser.ID)+`","locale":"en"}`)))
	createRequest.Header.Set("Content-Type", "application/json")
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, createRequest)
	require.Equal(t, http.StatusOK, createRecorder.Code, createRecorder.Body.String())
	var conversationResponse struct {
		ConversationID string `json:"conversation_id"`
	}
	require.NoError(t, json.Unmarshal(createRecorder.Body.Bytes(), &conversationResponse))
	require.NotEmpty(t, conversationResponse.ConversationID)
	cookies := createRecorder.Result().Cookies()
	require.NotEmpty(t, cookies)

	clientMessageID := "local-1732450000000-a1b2c3"
	send := func(message string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		body := []byte(`{"conversation_id":"` + conversationResponse.ConversationID + `","client_message_id":"` + clientMessageID + `","message":"` + message + `","sender_name":"Visitor","message_type":"text"}`)
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/messages", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}
		router.ServeHTTP(recorder, request)
		return recorder
	}

	firstRecorder := send("hello")
	require.Equal(t, http.StatusOK, firstRecorder.Code, firstRecorder.Body.String())
	var firstResponse struct {
		MessageID uint `json:"message_id"`
		Data      struct {
			ClientMessageID string `json:"client_message_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(firstRecorder.Body.Bytes(), &firstResponse))
	require.NotZero(t, firstResponse.MessageID)
	assert.Equal(t, clientMessageID, firstResponse.Data.ClientMessageID)
	var outboxCountAfterFirstSend int64
	require.NoError(t, db.Model(&outbox.Event{}).Count(&outboxCountAfterFirstSend).Error)

	replayRecorder := send("hello")
	require.Equal(t, http.StatusOK, replayRecorder.Code, replayRecorder.Body.String())
	var replayResponse struct {
		MessageID uint `json:"message_id"`
	}
	require.NoError(t, json.Unmarshal(replayRecorder.Body.Bytes(), &replayResponse))
	assert.Equal(t, firstResponse.MessageID, replayResponse.MessageID)

	historyRecorder := httptest.NewRecorder()
	historyRequest := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/messages/"+url.PathEscape(conversationResponse.ConversationID), nil)
	for _, cookie := range cookies {
		historyRequest.AddCookie(cookie)
	}
	router.ServeHTTP(historyRecorder, historyRequest)
	require.Equal(t, http.StatusOK, historyRecorder.Code, historyRecorder.Body.String())
	var historyResponse struct {
		Data []struct {
			ClientMessageID string `json:"client_message_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(historyRecorder.Body.Bytes(), &historyResponse))
	require.Len(t, historyResponse.Data, 1)
	assert.Equal(t, clientMessageID, historyResponse.Data[0].ClientMessageID)

	var messageCount int64
	require.NoError(t, db.Model(&ticketdomain.TicketMessage{}).Count(&messageCount).Error)
	assert.EqualValues(t, 1, messageCount)
	var outboxCountAfterReplay int64
	require.NoError(t, db.Model(&outbox.Event{}).Count(&outboxCountAfterReplay).Error)
	assert.Equal(t, outboxCountAfterFirstSend, outboxCountAfterReplay)

	conflictRecorder := send("different message")
	assert.Equal(t, http.StatusConflict, conflictRecorder.Code, conflictRecorder.Body.String())
	var messageCountAfterConflict int64
	require.NoError(t, db.Model(&ticketdomain.TicketMessage{}).Count(&messageCountAfterConflict).Error)
	assert.Equal(t, messageCount, messageCountAfterConflict)
}

func TestHasPublicCustomerServiceConversationBackfillsExistingVisitorTimezone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := newPublicCustomerServiceHandlerTestEnv(t)
	supportUser := seedPublicCustomerServiceSupportUser(t, db)

	router := gin.New()
	router.POST("/conversations", handler.EnsurePublicCustomerServiceConversation)
	router.GET("/has-conversation", handler.HasPublicCustomerServiceConversation)

	createRecorder := httptest.NewRecorder()
	createBody := []byte(`{"agent_id":"` + itoaUint(supportUser.ID) + `","locale":"en"}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/conversations", bytes.NewReader(createBody))
	createRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(createRecorder, createRequest)
	require.Equal(t, http.StatusOK, createRecorder.Code)
	require.NotEmpty(t, createRecorder.Result().Cookies())

	var profile visitor.Profile
	require.NoError(t, db.First(&profile).Error)
	assert.Empty(t, profile.Timezone)
	qualityBefore := profile.ProfileQualityScore
	meaningfulSeenBefore := profile.LastMeaningfulSeenAt

	hasRecorder := httptest.NewRecorder()
	hasRequest := httptest.NewRequest(http.MethodGet, "/has-conversation", nil)
	hasRequest.Header.Set("X-Timezone", "Asia/Shanghai")
	for _, cookie := range createRecorder.Result().Cookies() {
		hasRequest.AddCookie(cookie)
	}
	router.ServeHTTP(hasRecorder, hasRequest)
	require.Equal(t, http.StatusOK, hasRecorder.Code)

	require.NoError(t, db.First(&profile, profile.ID).Error)
	assert.Equal(t, "Asia/Shanghai", profile.Timezone)
	assert.Equal(t, qualityBefore, profile.ProfileQualityScore)
	assert.Equal(t, meaningfulSeenBefore, profile.LastMeaningfulSeenAt)
}

func TestGetPublicCustomerServiceMessagesSupportsLatestAndStableCursorPages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := newPublicCustomerServiceHandlerTestEnv(t)
	supportUser := seedPublicCustomerServiceSupportUser(t, db)

	router := gin.New()
	router.POST("/conversations", handler.EnsurePublicCustomerServiceConversation)
	router.GET("/messages/:conversation_id", handler.GetPublicCustomerServiceMessages)

	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/conversations", bytes.NewReader([]byte(`{"agent_id":"`+itoaUint(supportUser.ID)+`","locale":"en"}`)))
	createRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(createRecorder, createRequest)
	require.Equal(t, http.StatusOK, createRecorder.Code)

	var conversationResponse struct {
		ConversationID string `json:"conversation_id"`
	}
	require.NoError(t, json.Unmarshal(createRecorder.Body.Bytes(), &conversationResponse))
	require.NotEmpty(t, conversationResponse.ConversationID)

	var conversation ticketdomain.Ticket
	require.NoError(t, db.Where("category = ?", "customer_service").First(&conversation).Error)
	createdAt := time.Date(2027, time.September, 11, 10, 0, 0, 0, time.UTC)
	for index, content := range []string{"first", "second", "third", "fourth", "fifth"} {
		require.NoError(t, db.Create(&ticketdomain.TicketMessage{
			ID:        uint(9_007_199_254_740_991 + index),
			TicketID:  conversation.ID,
			Content:   content,
			CreatedAt: createdAt,
		}).Error)
	}

	requestWithVisitorCookies := func(path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
		for _, cookie := range createRecorder.Result().Cookies() {
			request.AddCookie(cookie)
		}
		router.ServeHTTP(recorder, request)
		return recorder
	}

	latestRecorder := requestWithVisitorCookies("/messages/" + url.PathEscape(conversationResponse.ConversationID) + "?limit=2&latest=true")
	require.Equal(t, http.StatusOK, latestRecorder.Code)
	var latestResponse struct {
		Data []struct {
			Message  string `json:"message"`
			CursorID string `json:"cursor_id"`
		} `json:"data"`
		HasOlder   bool `json:"has_older"`
		HasNewer   bool `json:"has_newer"`
		NextBefore struct {
			CreatedAt time.Time `json:"created_at"`
			ID        string    `json:"id"`
		} `json:"next_before"`
	}
	require.NoError(t, json.Unmarshal(latestRecorder.Body.Bytes(), &latestResponse))
	require.Len(t, latestResponse.Data, 2)
	assert.Equal(t, "fourth", latestResponse.Data[0].Message)
	assert.Equal(t, "fifth", latestResponse.Data[1].Message)
	assert.Equal(t, "9007199254740994", latestResponse.Data[0].CursorID)
	assert.Equal(t, "9007199254740994", latestResponse.NextBefore.ID)
	assert.True(t, latestResponse.HasOlder)
	assert.False(t, latestResponse.HasNewer)

	query := url.Values{}
	query.Set("limit", "2")
	query.Set("before_created_at", latestResponse.NextBefore.CreatedAt.Format(time.RFC3339Nano))
	query.Set("before_id", latestResponse.NextBefore.ID)
	olderRecorder := requestWithVisitorCookies("/messages/" + url.PathEscape(conversationResponse.ConversationID) + "?" + query.Encode())
	require.Equal(t, http.StatusOK, olderRecorder.Code)
	var olderResponse struct {
		Data []struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(olderRecorder.Body.Bytes(), &olderResponse))
	require.Len(t, olderResponse.Data, 2)
	assert.Equal(t, "second", olderResponse.Data[0].Message)
	assert.Equal(t, "third", olderResponse.Data[1].Message)
}

func newPublicCustomerServiceHandlerTestEnv(t *testing.T) (*gorm.DB, *Handler) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&user.User{},
		&user.AgentProfile{},
		&visitor.Profile{},
		&ticketdomain.Ticket{},
		&ticketdomain.TicketMessage{},
		&ticketdomain.AutoReplyRule{},
		&ticketdomain.CustomerServiceInboxState{},
	))

	ticketService := service.NewTicketService(repository.NewTicketRepository(db), repository.NewUserRepository(db))
	visitorProfileService := service.NewVisitorProfileService(repository.NewVisitorProfileRepository(db))
	handler := NewHandler(ticketService, Options{
		VisitorSecret:         "test-secret",
		VisitorProfileService: visitorProfileService,
	})
	return db, handler
}

func seedPublicCustomerServiceSupportUser(t *testing.T, db *gorm.DB) user.User {
	t.Helper()

	support := user.User{
		Email:    "support@example.test",
		Username: "support",
		Password: "test-password",
		Role:     "support",
		Status:   "active",
	}
	require.NoError(t, db.Create(&support).Error)
	return support
}

func itoaUint(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

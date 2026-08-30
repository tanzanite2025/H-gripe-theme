package ticket

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

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
	body := []byte(`{"agent_id":"` + itoaUint(supportUser.ID) + `"}`)
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
	assert.Equal(t, "America/Los_Angeles", profile.Timezone)
	assert.Equal(t, service.VisitorProfileActionCustomerService, profile.LastMeaningfulAction)
	assert.Equal(t, service.VisitorProfileQualityCustomerService, profile.ProfileQualityScore)
	assert.NotEmpty(t, profile.CustomerServiceVisitorHash)
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

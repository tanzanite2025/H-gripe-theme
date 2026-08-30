package service

import (
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAdminPublicChatAgentProfileAllowsEmailOnlyWhenActive(t *testing.T) {
	db, service := newTestAdminPublicChatAgentService(t)
	agentUser := createTestAdminPublicChatAgentUser(t, db, "", "agent-email-only", "support")

	agent, created, err := service.UpsertPublicChatAgentProfile(AdminPublicChatAgentUpsertInput{
		UserID:       agentUser.ID,
		Email:        "support@example.test",
		Status:       "active",
		OnlineStatus: "offline",
	})
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, agent)
	assert.Equal(t, "support@example.test", agent.Email)
	assert.Empty(t, agent.WhatsApp)
}

func TestAdminPublicChatAgentProfileAllowsWhatsAppOnlyWhenActive(t *testing.T) {
	db, service := newTestAdminPublicChatAgentService(t)
	agentUser := createTestAdminPublicChatAgentUser(t, db, "", "agent-whatsapp-only", "support")

	agent, created, err := service.UpsertPublicChatAgentProfile(AdminPublicChatAgentUpsertInput{
		UserID:       agentUser.ID,
		WhatsApp:     "+1 555 0100",
		Status:       "active",
		OnlineStatus: "offline",
	})
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, agent)
	assert.Equal(t, "", agent.Email)
	assert.Equal(t, "+1 555 0100", agent.WhatsApp)
}

func TestAdminPublicChatAgentProfileRejectsActiveProfileWithoutContact(t *testing.T) {
	db, service := newTestAdminPublicChatAgentService(t)
	agentUser := createTestAdminPublicChatAgentUser(t, db, "", "agent-no-contact", "support")

	agent, created, err := service.UpsertPublicChatAgentProfile(AdminPublicChatAgentUpsertInput{
		UserID:       agentUser.ID,
		Status:       "active",
		OnlineStatus: "offline",
	})
	require.ErrorIs(t, err, ErrPublicChatAgentContactRequired)
	assert.Nil(t, agent)
	assert.False(t, created)
}

func TestAdminPublicChatAgentProfileAllowsInactiveProfileWithoutContact(t *testing.T) {
	db, service := newTestAdminPublicChatAgentService(t)
	agentUser := createTestAdminPublicChatAgentUser(t, db, "", "agent-inactive", "support")

	agent, created, err := service.UpsertPublicChatAgentProfile(AdminPublicChatAgentUpsertInput{
		UserID:       agentUser.ID,
		Status:       "inactive",
		OnlineStatus: "offline",
	})
	require.NoError(t, err)
	require.True(t, created)
	require.NotNil(t, agent)
	assert.Empty(t, agent.Email)
	assert.Empty(t, agent.WhatsApp)
}

func newTestAdminPublicChatAgentService(t *testing.T) (*gorm.DB, *AdminPublicChatAgentService) {
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
		&user.AgentGroup{},
		&user.AgentGroupMember{},
	))

	return db, NewAdminPublicChatAgentService(repository.NewUserRepository(db))
}

func createTestAdminPublicChatAgentUser(t *testing.T, db *gorm.DB, email, username, role string) user.User {
	t.Helper()

	item := user.User{
		Email:    email,
		Username: username,
		Password: "test-password",
		Role:     role,
		Status:   "active",
	}
	require.NoError(t, db.Create(&item).Error)
	return item
}

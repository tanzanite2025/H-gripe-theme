package service

import (
	"testing"

	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCustomerServiceAgentGroupsRoundTrip(t *testing.T) {
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
		&user.AgentGroup{},
		&user.AgentProfile{},
		&user.AgentGroupMember{},
	))

	agentUser := user.User{
		Email:    "tech@example.com",
		Username: "tech",
		Password: "test-password",
		Role:     "support",
		Status:   "active",
	}
	require.NoError(t, db.Create(&agentUser).Error)

	profile := user.AgentProfile{
		AgentID: "tech-1",
		UserID:  &agentUser.ID,
		Name:    "Technical Support",
		Status:  "active",
	}
	require.NoError(t, db.Create(&profile).Error)

	group := user.AgentGroup{
		Code:      "technical_support",
		Name:      "Technical Support",
		Status:    "active",
		SortOrder: 20,
	}
	require.NoError(t, db.Create(&group).Error)

	repo := repository.NewUserRepository(db)
	require.NoError(t, repo.ReplaceCustomerServiceAgentProfileGroups(profile.ID, []uint{group.ID}))

	profiles, err := repo.FindCustomerServiceAgentProfiles(10)
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	require.Len(t, profiles[0].Groups, 1)
	require.Equal(t, group.ID, profiles[0].Groups[0].ID)
	require.Equal(t, "technical_support", profiles[0].Groups[0].Code)
}

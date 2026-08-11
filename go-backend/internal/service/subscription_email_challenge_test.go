package service

import (
	"path"
	"strings"
	"testing"

	domainsubscription "commerce-platform/internal/domain/subscription"
	"commerce-platform/internal/domain/verification"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSubscriptionRequiresConfirmationAndConsumesTokensOnce(t *testing.T) {
	db := newEmailChallengeTestDB(t)
	subscriptionService := NewSubscriptionService(repository.NewSubscriptionRepository(db))
	emailSender := &recordingEmailSender{}
	subscriptionService.ConfigureEmailChallenges(
		repository.NewEmailChallengeRepository(db),
		"test-email-secret",
		emailSender,
	)
	subscriptionService.ConfigureEmailBaseURL("https://api.example.test")

	sub, err := subscriptionService.Subscribe("Rider@Example.test", "website", "en", nil)
	require.NoError(t, err)
	assert.Equal(t, "pending", sub.Status)

	createdToken, err := subscriptionService.IssueSubscriptionConfirmation(sub.Email)
	require.NoError(t, err)
	require.NotEmpty(t, createdToken)
	require.Len(t, emailSender.bodies, 1)

	pending, err := subscriptionService.GetSubscription(sub.Email)
	require.NoError(t, err)
	assert.Equal(t, "pending", pending.Status)

	require.NoError(t, subscriptionService.ConfirmSubscription(createdToken))
	active, err := subscriptionService.GetSubscription(sub.Email)
	require.NoError(t, err)
	assert.Equal(t, "active", active.Status)

	require.ErrorIs(t, subscriptionService.ConfirmSubscription(createdToken), ErrInvalidSubscriptionToken)
}

func TestSubscriptionEmailActionDoesNotMutateByEmailAlone(t *testing.T) {
	db := newEmailChallengeTestDB(t)
	subscriptionService := NewSubscriptionService(repository.NewSubscriptionRepository(db))
	emailSender := &recordingEmailSender{}
	subscriptionService.ConfigureEmailChallenges(
		repository.NewEmailChallengeRepository(db),
		"test-email-secret",
		emailSender,
	)
	subscriptionService.ConfigureEmailBaseURL("https://api.example.test")

	sub, err := subscriptionService.Subscribe("rider@example.test", "website", "en", nil)
	require.NoError(t, err)
	confirmToken, err := subscriptionService.IssueSubscriptionConfirmation(sub.Email)
	require.NoError(t, err)
	require.NoError(t, subscriptionService.ConfirmSubscription(confirmToken))

	require.NoError(t, subscriptionService.UnsubscribeByEmail(sub.Email))
	stillActive, err := subscriptionService.GetSubscription(sub.Email)
	require.NoError(t, err)
	assert.Equal(t, "active", stillActive.Status)
	require.Len(t, emailSender.bodies, 2)

	unsubscribeURL := strings.TrimSpace(strings.Split(emailSender.bodies[1], "\n\n")[1])
	unsubscribeToken := path.Base(unsubscribeURL)
	require.NoError(t, subscriptionService.Unsubscribe(unsubscribeToken))

	unsubscribed, err := subscriptionService.GetSubscription(sub.Email)
	require.NoError(t, err)
	assert.Equal(t, "unsubscribed", unsubscribed.Status)
}

func newEmailChallengeTestDB(t *testing.T) *gorm.DB {
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
		&domainsubscription.Subscription{},
		&verification.EmailChallenge{},
	))
	return db
}

package service

import (
	"path"
	"strings"
	"testing"

	"commerce-platform/internal/domain/outbox"
	domainsubscription "commerce-platform/internal/domain/subscription"
	"commerce-platform/internal/domain/verification"
	"commerce-platform/internal/pkg/emailtoken"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSubscriptionRequiresConfirmationAndConsumesTokensOnce(t *testing.T) {
	db := newEmailChallengeTestDB(t)
	subscriptionService := NewSubscriptionService(newTestEmailChallengeTxManager(db), repository.NewSubscriptionRepository(db))
	emailSender := &recordingEmailSender{}
	subscriptionService.ConfigureEmailChallenges("test-email-secret")
	subscriptionService.ConfigureEmailBaseURL("https://api.example.test")

	sub, createdToken, err := subscriptionService.Subscribe("Rider@Example.test", "website", "en", nil)
	require.NoError(t, err)
	assert.Equal(t, "pending", sub.Status)
	require.NotEmpty(t, createdToken)
	var deliveryEvent outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeEmailChallengeDelivery).First(&deliveryEvent).Error)
	require.NotContains(t, string(deliveryEvent.Payload), createdToken)
	processEmailChallengeDelivery(t, db, emailSender)
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

func TestSubscriptionRollsBackWhenOutboxWriteFails(t *testing.T) {
	db := newEmailChallengeTestDB(t)
	require.NoError(t, db.Migrator().DropTable(&outbox.Event{}))

	service := NewSubscriptionService(newTestEmailChallengeTxManager(db), repository.NewSubscriptionRepository(db))
	service.ConfigureEmailChallenges("test-email-secret")
	service.ConfigureEmailBaseURL("https://api.example.test")

	_, _, err := service.Subscribe("rollback@example.test", "website", "en", nil)
	require.Error(t, err)

	var subscriptionCount int64
	require.NoError(t, db.Model(&domainsubscription.Subscription{}).Count(&subscriptionCount).Error)
	require.Zero(t, subscriptionCount)
	var challengeCount int64
	require.NoError(t, db.Model(&verification.EmailChallenge{}).Count(&challengeCount).Error)
	require.Zero(t, challengeCount)
}

func TestSubscriptionStatusFailureDoesNotConsumeChallenge(t *testing.T) {
	db := newEmailChallengeTestDB(t)
	service := NewSubscriptionService(newTestEmailChallengeTxManager(db), repository.NewSubscriptionRepository(db))
	service.ConfigureEmailChallenges("test-email-secret")
	service.ConfigureEmailBaseURL("https://api.example.test")

	_, token, err := service.Subscribe("retry@example.test", "website", "en", nil)
	require.NoError(t, err)
	require.NoError(t, db.Migrator().DropTable(&domainsubscription.Subscription{}))
	require.Error(t, service.ConfirmSubscription(token))

	challenge, err := repository.NewEmailChallengeRepository(db).Find(emailtoken.Hash(token), subscriptionConfirmPurpose)
	require.NoError(t, err)
	require.Nil(t, challenge.UsedAt)
}

func TestSubscriptionEmailActionDoesNotMutateByEmailAlone(t *testing.T) {
	db := newEmailChallengeTestDB(t)
	subscriptionService := NewSubscriptionService(newTestEmailChallengeTxManager(db), repository.NewSubscriptionRepository(db))
	emailSender := &recordingEmailSender{}
	subscriptionService.ConfigureEmailChallenges("test-email-secret")
	subscriptionService.ConfigureEmailBaseURL("https://api.example.test")

	sub, confirmToken, err := subscriptionService.Subscribe("rider@example.test", "website", "en", nil)
	require.NoError(t, err)
	processEmailChallengeDelivery(t, db, emailSender)
	require.NoError(t, subscriptionService.ConfirmSubscription(confirmToken))

	require.NoError(t, subscriptionService.UnsubscribeByEmail(sub.Email))
	processEmailChallengeDelivery(t, db, emailSender)
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
		&outbox.Event{},
	))
	return db
}

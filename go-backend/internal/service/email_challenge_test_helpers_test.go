package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type recordingEmailSender struct {
	bodies []string
}

func (s *recordingEmailSender) SendEmail(_ []string, _ string, body string) error {
	s.bodies = append(s.bodies, body)
	return nil
}

func processEmailChallengeDelivery(t *testing.T, db *gorm.DB, sender EmailChallengeSender) {
	t.Helper()
	var event outbox.Event
	require.NoError(t, db.Where("event_type = ?", outbox.EventTypeEmailChallengeDelivery).Order("id DESC").First(&event).Error)
	require.NoError(t, NewEmailChallengeDeliveryOutboxHandler(sender, nil, "test-email-secret").Handle(context.Background(), event))
	_ = repository.NewOutboxRepository(db).MarkProcessed(event.ID, event.CreatedAt)
}

func newTestEmailChallengeTxManager(db *gorm.DB) *repository.EmailChallengeTxManager {
	return repository.NewEmailChallengeTxManager(
		db,
		repository.NewSubscriptionRepository(db),
		repository.NewWarrantyRepository(db),
		repository.NewEmailChallengeRepository(db),
		repository.NewOutboxRepository(db),
	)
}

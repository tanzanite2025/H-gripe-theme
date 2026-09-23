package repository

import (
	"errors"

	"gorm.io/gorm"
)

var ErrEmailChallengeTransactionNotConfigured = errors.New("email challenge transaction is not configured")

type EmailChallengeTxRepositories struct {
	Subscription   *SubscriptionRepository
	Warranty       *WarrantyRepository
	EmailChallenge *EmailChallengeRepository
	Outbox         *OutboxRepository
}

// EmailChallengeTxManager owns the small transaction boundary shared by
// subscription and warranty verification workflows.
type EmailChallengeTxManager struct {
	db                 *gorm.DB
	subscriptionRepo   *SubscriptionRepository
	warrantyRepo       *WarrantyRepository
	emailChallengeRepo *EmailChallengeRepository
	outboxRepo         *OutboxRepository
}

func NewEmailChallengeTxManager(
	db *gorm.DB,
	subscriptionRepo *SubscriptionRepository,
	warrantyRepo *WarrantyRepository,
	emailChallengeRepo *EmailChallengeRepository,
	outboxRepo *OutboxRepository,
) *EmailChallengeTxManager {
	return &EmailChallengeTxManager{
		db:                 db,
		subscriptionRepo:   subscriptionRepo,
		warrantyRepo:       warrantyRepo,
		emailChallengeRepo: emailChallengeRepo,
		outboxRepo:         outboxRepo,
	}
}

func (m *EmailChallengeTxManager) WithinTx(fn func(EmailChallengeTxRepositories) error) error {
	if m == nil || m.db == nil || m.emailChallengeRepo == nil || m.outboxRepo == nil || fn == nil {
		return ErrEmailChallengeTransactionNotConfigured
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var subscriptionRepo *SubscriptionRepository
		if m.subscriptionRepo != nil {
			subscriptionRepo = m.subscriptionRepo.WithTx(tx)
		}
		var warrantyRepo *WarrantyRepository
		if m.warrantyRepo != nil {
			warrantyRepo = m.warrantyRepo.WithTx(tx)
		}
		return fn(EmailChallengeTxRepositories{
			Subscription:   subscriptionRepo,
			Warranty:       warrantyRepo,
			EmailChallenge: m.emailChallengeRepo.WithTx(tx),
			Outbox:         m.outboxRepo.WithTx(tx),
		})
	})
}

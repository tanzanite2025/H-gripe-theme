package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/domain/subscription"
	"commerce-platform/internal/repository"
)

type SubscriptionService struct {
	txManager        *repository.EmailChallengeTxManager
	subscriptionRepo *repository.SubscriptionRepository
	challengeSecret  string
	baseURL          string
}

var (
	ErrInvalidSubscriptionStatus = errors.New("invalid subscription status")
	ErrInvalidSubscriptionToken  = errors.New("invalid subscription token")
)

const (
	subscriptionConfirmPurpose     = "subscription:confirm"
	subscriptionUnsubscribePurpose = "subscription:unsubscribe"
	subscriptionResubscribePurpose = "subscription:resubscribe"
	subscriptionStatusPurpose      = "subscription:status"
)

func NewSubscriptionService(txManager *repository.EmailChallengeTxManager, subscriptionRepo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		txManager:        txManager,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *SubscriptionService) ConfigureEmailChallenges(secret string) {
	s.challengeSecret = secret
}

func (s *SubscriptionService) ConfigureEmailBaseURL(baseURL string) {
	s.baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

// Subscribe creates the pending subscription and confirmation delivery event
// atomically. It becomes active only after the signed token is consumed.
func (s *SubscriptionService) Subscribe(email, source, locale string, tags []string) (*subscription.Subscription, string, error) {
	if s == nil || s.txManager == nil {
		return nil, "", ErrEmailChallengeUnavailable
	}
	email = normalizeSubscriptionEmail(email)
	unsubToken, err := generateUnsubToken()
	if err != nil {
		return nil, "", err
	}
	var result *subscription.Subscription
	var confirmationToken string
	err = s.txManager.WithinTx(func(repos repository.EmailChallengeTxRepositories) error {
		if repos.Subscription == nil {
			return repository.ErrEmailChallengeTransactionNotConfigured
		}
		sub, findErr := repos.Subscription.FindByEmail(email)
		switch {
		case findErr == nil && sub.Status == "active":
			return errors.New("email already subscribed")
		case findErr == nil:
			sub.Status = "pending"
			sub.Locale = locale
			sub.Source = source
			sub.Tags = joinTags(tags)
			sub.UnsubToken = unsubToken
			if err := repos.Subscription.Update(sub); err != nil {
				return err
			}
		case repository.IsRecordNotFound(findErr):
			sub = &subscription.Subscription{
				Email:        email,
				Status:       "pending",
				Locale:       locale,
				Source:       source,
				Tags:         joinTags(tags),
				UnsubToken:   unsubToken,
				SubscribedAt: time.Now().UTC(),
			}
			if err := repos.Subscription.Create(sub); err != nil {
				return err
			}
		default:
			return findErr
		}
		confirmationToken, err = issueEmailChallengeWithDelivery(
			repos.EmailChallenge,
			repos.Outbox,
			s.challengeSecret,
			subscriptionConfirmPurpose,
			email,
			email,
			"Confirm your newsletter subscription",
			func(token string) string {
				return s.subscriptionChallengeBody("confirm", token)
			},
			24*time.Hour,
		)
		if err != nil {
			return err
		}
		result = sub
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	return result, confirmationToken, nil
}

func (s *SubscriptionService) ConfirmSubscription(token string) error {
	return s.consumeChallengeAndSetStatus(token, subscriptionConfirmPurpose, "active")
}

func (s *SubscriptionService) ResubscribeByToken(token string) error {
	return s.consumeChallengeAndSetStatus(token, subscriptionResubscribePurpose, "active")
}

// Unsubscribe consumes a signed, single-use email token.
func (s *SubscriptionService) Unsubscribe(token string) error {
	return s.consumeChallengeAndSetStatus(token, subscriptionUnsubscribePurpose, "unsubscribed")
}

// UnsubscribeByEmail requests a signed email action; it does not mutate by email alone.
func (s *SubscriptionService) UnsubscribeByEmail(email string) error {
	return s.requestSubscriptionAction(
		email,
		subscriptionUnsubscribePurpose,
		"Unsubscribe from newsletter",
		"unsubscribe",
	)
}

func (s *SubscriptionService) UpdateStatus(email, status string) error {
	switch status {
	case "active":
		return s.setStatus(email, "active")
	case "unsubscribed":
		return s.setStatus(email, "unsubscribed")
	default:
		return ErrInvalidSubscriptionStatus
	}
}

// Resubscribe requests a signed email action; it does not mutate by email alone.
func (s *SubscriptionService) Resubscribe(email string) error {
	return s.requestSubscriptionAction(
		email,
		subscriptionResubscribePurpose,
		"Resume your newsletter subscription",
		"resubscribe",
	)
}

func (s *SubscriptionService) GetSubscription(email string) (*subscription.Subscription, error) {
	return s.subscriptionRepo.FindByEmail(normalizeSubscriptionEmail(email))
}

func (s *SubscriptionService) GetSubscriptionByToken(token string) (*subscription.Subscription, error) {
	if s == nil || s.txManager == nil {
		return nil, ErrEmailChallengeUnavailable
	}
	var result *subscription.Subscription
	err := s.txManager.WithinTx(func(repos repository.EmailChallengeTxRepositories) error {
		if repos.Subscription == nil {
			return repository.ErrEmailChallengeTransactionNotConfigured
		}
		claims, err := consumeEmailChallenge(repos.EmailChallenge, s.challengeSecret, token, subscriptionStatusPurpose)
		if err != nil {
			return ErrInvalidSubscriptionToken
		}
		result, err = repos.Subscription.FindByEmail(normalizeSubscriptionEmail(claims.Email))
		if err != nil {
			return ErrInvalidSubscriptionToken
		}
		return nil
	})
	return result, err
}

func (s *SubscriptionService) RequestStatus(email string) error {
	email = normalizeSubscriptionEmail(email)
	return s.requestSubscriptionAction(
		email,
		subscriptionStatusPurpose,
		"View your newsletter subscription status",
		"status",
	)
}

func (s *SubscriptionService) GetAllSubscriptions(page, pageSize int, status string) ([]subscription.Subscription, int64, error) {
	return s.subscriptionRepo.FindAll(page, pageSize, status)
}

func (s *SubscriptionService) GetSubscriptionsByTags(tags []string, page, pageSize int) ([]subscription.Subscription, int64, error) {
	return s.subscriptionRepo.FindByTags(tags, page, pageSize)
}

func (s *SubscriptionService) UpdateSubscription(sub *subscription.Subscription) error {
	return s.subscriptionRepo.Update(sub)
}

func (s *SubscriptionService) DeleteSubscription(email string) error {
	return s.subscriptionRepo.Delete(normalizeSubscriptionEmail(email))
}

func (s *SubscriptionService) BatchDelete(emails []string) (int, error) {
	deleted := 0
	for _, email := range emails {
		if err := s.DeleteSubscription(email); err == nil {
			deleted++
		}
	}
	return deleted, nil
}

func (s *SubscriptionService) GetStats() (map[string]interface{}, error) {
	return s.subscriptionRepo.GetStats()
}

func (s *SubscriptionService) GetActiveEmails() ([]string, error) {
	return s.subscriptionRepo.GetActiveEmails()
}

func (s *SubscriptionService) GetActiveEmailsByTags(tags []string) ([]string, error) {
	subscriptions, _, err := s.subscriptionRepo.FindByTags(tags, 1, 10000)
	if err != nil {
		return nil, err
	}

	emails := make([]string, len(subscriptions))
	for i, sub := range subscriptions {
		emails[i] = sub.Email
	}

	return emails, nil
}

func (s *SubscriptionService) setStatus(email, status string) error {
	return setSubscriptionStatus(s.subscriptionRepo, email, status)
}

func setSubscriptionStatus(repo *repository.SubscriptionRepository, email, status string) error {
	if repo == nil {
		return repository.ErrEmailChallengeTransactionNotConfigured
	}
	sub, err := repo.FindByEmail(normalizeSubscriptionEmail(email))
	if err != nil {
		return err
	}

	sub.Status = status
	switch status {
	case "active":
		sub.SubscribedAt = time.Now().UTC()
		sub.UnsubscribedAt = nil
	case "unsubscribed":
		now := time.Now().UTC()
		sub.UnsubscribedAt = &now
	}
	return repo.Update(sub)
}

func (s *SubscriptionService) consumeChallengeAndSetStatus(token, purpose, status string) error {
	if s == nil || s.txManager == nil {
		return ErrEmailChallengeUnavailable
	}
	return s.txManager.WithinTx(func(repos repository.EmailChallengeTxRepositories) error {
		if repos.Subscription == nil {
			return repository.ErrEmailChallengeTransactionNotConfigured
		}
		claims, err := consumeEmailChallenge(repos.EmailChallenge, s.challengeSecret, token, purpose)
		if err != nil {
			return ErrInvalidSubscriptionToken
		}
		if err := setSubscriptionStatus(repos.Subscription, claims.Email, status); err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrInvalidSubscriptionToken
			}
			return err
		}
		return nil
	})
}

func (s *SubscriptionService) requestSubscriptionAction(email, purpose, subject, action string) error {
	if s == nil || s.txManager == nil {
		return ErrEmailChallengeUnavailable
	}
	email = normalizeSubscriptionEmail(email)
	return s.txManager.WithinTx(func(repos repository.EmailChallengeTxRepositories) error {
		if repos.Subscription == nil {
			return repository.ErrEmailChallengeTransactionNotConfigured
		}
		if _, err := repos.Subscription.FindByEmail(email); err != nil {
			if repository.IsRecordNotFound(err) {
				return nil
			}
			return err
		}
		_, err := issueEmailChallengeWithDelivery(
			repos.EmailChallenge,
			repos.Outbox,
			s.challengeSecret,
			purpose,
			email,
			email,
			subject,
			func(token string) string {
				return s.subscriptionChallengeBody(action, token)
			},
			24*time.Hour,
		)
		return err
	})
}

func (s *SubscriptionService) subscriptionChallengeBody(action, token string) string {
	pathAction := action
	if action == "status" {
		pathAction = "status-token"
	}
	link := fmt.Sprintf("%s/api/v1/subscriptions/%s/%s", s.baseURL, pathAction, url.PathEscape(token))
	return fmt.Sprintf(
		"Please use the following link to complete your newsletter request:\n\n%s\n\nThis link expires in 24 hours and can only be used once.",
		link,
	)
}

func normalizeSubscriptionEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateUnsubToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ","
		}
		result += tag
	}
	return result
}

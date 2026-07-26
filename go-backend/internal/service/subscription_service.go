package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"tanzanite/internal/domain/subscription"
	"tanzanite/internal/repository"
)

type SubscriptionService struct {
	subscriptionRepo *repository.SubscriptionRepository
	challengeRepo    *repository.EmailChallengeRepository
	challengeSecret  string
	emailSender      EmailChallengeSender
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

func NewSubscriptionService(subscriptionRepo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *SubscriptionService) ConfigureEmailChallenges(
	challengeRepo *repository.EmailChallengeRepository,
	secret string,
	senders ...EmailChallengeSender,
) {
	s.challengeRepo = challengeRepo
	s.challengeSecret = secret
	if len(senders) > 0 {
		s.emailSender = senders[0]
	}
}

func (s *SubscriptionService) ConfigureEmailBaseURL(baseURL string) {
	s.baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

// Subscribe creates a pending subscription. It becomes active only after the
// signed confirmation token is consumed.
func (s *SubscriptionService) Subscribe(email, source, locale string, tags []string) (*subscription.Subscription, error) {
	email = normalizeSubscriptionEmail(email)
	exists, err := s.subscriptionRepo.CheckEmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already subscribed")
	}

	token, err := generateUnsubToken()
	if err != nil {
		return nil, err
	}

	existing, findErr := s.subscriptionRepo.FindByEmail(email)
	if findErr == nil {
		existing.Status = "pending"
		existing.Locale = locale
		existing.Source = source
		existing.Tags = joinTags(tags)
		existing.UnsubToken = token
		if err := s.subscriptionRepo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if !repository.IsRecordNotFound(findErr) {
		return nil, findErr
	}

	sub := &subscription.Subscription{
		Email:        email,
		Status:       "pending",
		Locale:       locale,
		Source:       source,
		Tags:         joinTags(tags),
		UnsubToken:   token,
		SubscribedAt: time.Now(),
	}
	if err := s.subscriptionRepo.Create(sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) IssueSubscriptionConfirmation(email string) (string, error) {
	email = normalizeSubscriptionEmail(email)
	sub, err := s.subscriptionRepo.FindByEmail(email)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return "", nil
		}
		return "", err
	}
	if sub.Status == "active" {
		return "", nil
	}

	token, err := issueEmailChallenge(
		s.challengeRepo,
		s.challengeSecret,
		subscriptionConfirmPurpose,
		email,
		email,
		24*time.Hour,
	)
	if err != nil {
		return "", err
	}

	return token, s.sendSubscriptionChallenge(
		email,
		"Confirm your Tanzanite newsletter subscription",
		"confirm",
		token,
	)
}

func (s *SubscriptionService) ConfirmSubscription(token string) error {
	claims, err := consumeEmailChallenge(
		s.challengeRepo,
		s.challengeSecret,
		token,
		subscriptionConfirmPurpose,
	)
	if err != nil {
		return ErrInvalidSubscriptionToken
	}

	sub, err := s.subscriptionRepo.FindByEmail(normalizeSubscriptionEmail(claims.Email))
	if err != nil {
		return ErrInvalidSubscriptionToken
	}
	if sub.Status == "active" {
		return nil
	}

	sub.Status = "active"
	sub.SubscribedAt = time.Now()
	sub.UnsubscribedAt = nil
	return s.subscriptionRepo.Update(sub)
}

func (s *SubscriptionService) ResubscribeByToken(token string) error {
	claims, err := consumeEmailChallenge(
		s.challengeRepo,
		s.challengeSecret,
		token,
		subscriptionResubscribePurpose,
	)
	if err != nil {
		return ErrInvalidSubscriptionToken
	}
	return s.setStatus(claims.Email, "active")
}

// Unsubscribe consumes a signed, single-use email token.
func (s *SubscriptionService) Unsubscribe(token string) error {
	claims, err := consumeEmailChallenge(
		s.challengeRepo,
		s.challengeSecret,
		token,
		subscriptionUnsubscribePurpose,
	)
	if err != nil {
		return ErrInvalidSubscriptionToken
	}

	return s.setStatus(claims.Email, "unsubscribed")
}

// UnsubscribeByEmail requests a signed email action; it does not mutate by email alone.
func (s *SubscriptionService) UnsubscribeByEmail(email string) error {
	return s.requestSubscriptionAction(
		email,
		subscriptionUnsubscribePurpose,
		"Unsubscribe from Tanzanite newsletter",
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
		"Resume your Tanzanite newsletter subscription",
		"resubscribe",
	)
}

func (s *SubscriptionService) GetSubscription(email string) (*subscription.Subscription, error) {
	return s.subscriptionRepo.FindByEmail(normalizeSubscriptionEmail(email))
}

func (s *SubscriptionService) GetSubscriptionByToken(token string) (*subscription.Subscription, error) {
	claims, err := consumeEmailChallenge(
		s.challengeRepo,
		s.challengeSecret,
		token,
		subscriptionStatusPurpose,
	)
	if err != nil {
		return nil, ErrInvalidSubscriptionToken
	}
	return s.GetSubscription(claims.Email)
}

func (s *SubscriptionService) RequestStatus(email string) error {
	email = normalizeSubscriptionEmail(email)
	return s.requestSubscriptionAction(
		email,
		subscriptionStatusPurpose,
		"View your Tanzanite newsletter subscription status",
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
	sub, err := s.subscriptionRepo.FindByEmail(normalizeSubscriptionEmail(email))
	if err != nil {
		return err
	}

	sub.Status = status
	switch status {
	case "active":
		sub.SubscribedAt = time.Now()
		sub.UnsubscribedAt = nil
	case "unsubscribed":
		now := time.Now()
		sub.UnsubscribedAt = &now
	}
	return s.subscriptionRepo.Update(sub)
}

func (s *SubscriptionService) requestSubscriptionAction(email, purpose, subject, action string) error {
	email = normalizeSubscriptionEmail(email)
	if _, err := s.subscriptionRepo.FindByEmail(email); err != nil {
		if repository.IsRecordNotFound(err) {
			return nil
		}
		return err
	}

	token, err := issueEmailChallenge(
		s.challengeRepo,
		s.challengeSecret,
		purpose,
		email,
		email,
		24*time.Hour,
	)
	if err != nil {
		return err
	}
	return s.sendSubscriptionChallenge(email, subject, action, token)
}

func (s *SubscriptionService) sendSubscriptionChallenge(email, subject, action, token string) error {
	if s.emailSender == nil {
		return ErrEmailChallengeUnavailable
	}

	pathAction := action
	if action == "status" {
		pathAction = "status-token"
	}
	link := fmt.Sprintf("%s/api/v1/subscriptions/%s/%s", s.baseURL, pathAction, url.PathEscape(token))
	body := fmt.Sprintf(
		"Please use the following link to complete your Tanzanite newsletter request:\n\n%s\n\nThis link expires in 24 hours and can only be used once.",
		link,
	)
	return s.emailSender.SendEmail([]string{email}, subject, body)
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

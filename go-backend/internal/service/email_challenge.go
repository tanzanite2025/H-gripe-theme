package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/verification"
	"commerce-platform/internal/pkg/emailtoken"
	"commerce-platform/internal/repository"
	"gorm.io/datatypes"
)

var (
	ErrEmailChallengeUnavailable = errors.New("email challenge service is unavailable")
	ErrEmailChallengeInvalid     = errors.New("email challenge is invalid or expired")
)

type EmailChallengeSender interface {
	SendEmail(to []string, subject, body string) error
}

const emailChallengeTokenPlaceholder = "EMAIL_CHALLENGE_TOKEN_PLACEHOLDER_7F3A"

// issueEmailChallengeWithDelivery writes through repositories supplied by the
// caller's transaction. The outbox worker owns the actual SMTP call.
func issueEmailChallengeWithDelivery(
	repo *repository.EmailChallengeRepository,
	outboxRepo *repository.OutboxRepository,
	secret, purpose, email, challengeSubject, deliverySubject string,
	bodyBuilder func(token string) string,
	ttl time.Duration,
) (string, error) {
	if repo == nil || outboxRepo == nil || strings.TrimSpace(secret) == "" {
		return "", ErrEmailChallengeUnavailable
	}
	now := time.Now().UTC()
	token, err := emailtoken.Sign(secret, emailtoken.Claims{
		Purpose:   purpose,
		Email:     email,
		Subject:   challengeSubject,
		ExpiresAt: now.Add(ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	signedClaims, err := emailtoken.Verify(secret, token, purpose, now)
	if err != nil {
		return "", fmt.Errorf("read signed email challenge claims: %w", err)
	}
	if bodyBuilder == nil {
		return "", ErrEmailChallengeUnavailable
	}
	bodyTemplate := bodyBuilder(emailChallengeTokenPlaceholder)
	if !strings.Contains(bodyTemplate, emailChallengeTokenPlaceholder) {
		return "", errors.New("email challenge body template must contain token placeholder")
	}
	payload, err := json.Marshal(outbox.EmailChallengeDeliveryPayload{
		RecipientEmail:   strings.TrimSpace(email),
		DeliverySubject:  strings.TrimSpace(deliverySubject),
		BodyTemplate:     bodyTemplate,
		Purpose:          strings.TrimSpace(purpose),
		ChallengeSubject: challengeSubject,
		Nonce:            signedClaims.Nonce,
		ExpiresAt:        time.Unix(signedClaims.ExpiresAt, 0).UTC(),
		RequestedAt:      now,
	})
	if err != nil {
		return "", fmt.Errorf("encode email challenge delivery event: %w", err)
	}
	tokenHash := emailtoken.Hash(token)
	event := &outbox.Event{
		EventKey:      fmt.Sprintf("%s:%s", outbox.EventTypeEmailChallengeDelivery, tokenHash),
		EventType:     outbox.EventTypeEmailChallengeDelivery,
		AggregateType: outbox.AggregateTypeEmailChallenge,
		AggregateID:   tokenHash,
		Payload:       datatypes.JSON(payload),
		AvailableAt:   now,
	}
	if err := repo.Create(&verification.EmailChallenge{
		Purpose:   purpose,
		Email:     email,
		Subject:   challengeSubject,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(ttl),
	}); err != nil {
		return "", err
	}
	if err := outboxRepo.CreateEvent(event); err != nil {
		return "", err
	}
	return token, nil
}

func consumeEmailChallenge(
	repo *repository.EmailChallengeRepository,
	secret, token, purpose string,
) (emailtoken.Claims, error) {
	if repo == nil || strings.TrimSpace(secret) == "" {
		return emailtoken.Claims{}, ErrEmailChallengeUnavailable
	}

	claims, err := emailtoken.Verify(secret, token, purpose, time.Now())
	if err != nil {
		return emailtoken.Claims{}, ErrEmailChallengeInvalid
	}

	challenge, err := repo.Consume(emailtoken.Hash(token), purpose, time.Now())
	if err != nil {
		return emailtoken.Claims{}, ErrEmailChallengeInvalid
	}
	if !strings.EqualFold(challenge.Email, claims.Email) || challenge.Subject != claims.Subject {
		return emailtoken.Claims{}, ErrEmailChallengeInvalid
	}

	return claims, nil
}

func validateEmailChallenge(
	repo *repository.EmailChallengeRepository,
	secret, token, purpose string,
) (emailtoken.Claims, error) {
	if repo == nil || strings.TrimSpace(secret) == "" {
		return emailtoken.Claims{}, ErrEmailChallengeUnavailable
	}

	claims, err := emailtoken.Verify(secret, token, purpose, time.Now())
	if err != nil {
		return emailtoken.Claims{}, ErrEmailChallengeInvalid
	}

	challenge, err := repo.Find(emailtoken.Hash(token), purpose)
	if err != nil || challenge.UsedAt != nil || !challenge.ExpiresAt.After(time.Now()) {
		return emailtoken.Claims{}, ErrEmailChallengeInvalid
	}
	if !strings.EqualFold(challenge.Email, claims.Email) || challenge.Subject != claims.Subject {
		return emailtoken.Claims{}, ErrEmailChallengeInvalid
	}
	return claims, nil
}

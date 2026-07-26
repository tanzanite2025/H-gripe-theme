package service

import (
	"errors"
	"strings"
	"time"

	"tanzanite/internal/domain/verification"
	"tanzanite/internal/pkg/emailtoken"
	"tanzanite/internal/repository"
)

var (
	ErrEmailChallengeUnavailable = errors.New("email challenge service is unavailable")
	ErrEmailChallengeInvalid     = errors.New("email challenge is invalid or expired")
)

type EmailChallengeSender interface {
	SendEmail(to []string, subject, body string) error
}

func issueEmailChallenge(
	repo *repository.EmailChallengeRepository,
	secret, purpose, email, subject string,
	ttl time.Duration,
) (string, error) {
	if repo == nil || strings.TrimSpace(secret) == "" {
		return "", ErrEmailChallengeUnavailable
	}

	now := time.Now()
	token, err := emailtoken.Sign(secret, emailtoken.Claims{
		Purpose:   purpose,
		Email:     email,
		Subject:   subject,
		ExpiresAt: now.Add(ttl).Unix(),
	})
	if err != nil {
		return "", err
	}

	if err := repo.Create(&verification.EmailChallenge{
		Purpose:   purpose,
		Email:     email,
		Subject:   subject,
		TokenHash: emailtoken.Hash(token),
		ExpiresAt: now.Add(ttl),
	}); err != nil {
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

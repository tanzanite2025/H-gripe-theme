package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/pkg/resilience"
)

const (
	googleTokenInfoEndpoint      = "https://oauth2.googleapis.com/tokeninfo"
	googleOAuthCircuitBreakerKey = "google-oauth-tokeninfo"
)

type googleTokenInfo struct {
	Issuer        string `json:"iss"`
	Subject       string `json:"sub"`
	Audience      string `json:"aud"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

var ErrGoogleBackofficeAccessDenied = errors.New("google account is not allowed to access admin panel")

// ConfigureGoogleOAuthResilience wires the shared outbound policy used by
// Google tokeninfo verification. tokeninfo is a GET endpoint, so retries stay
// limited to safe methods even when the application-wide policy allows
// retries for other integrations.
func (s *AuthService) ConfigureGoogleOAuthResilience(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) {
	if s == nil {
		return
	}
	retry.RetryUnsafeMethods = false
	s.googleOAuthHTTPClient = resilience.NewHTTPClient(
		&http.Client{Timeout: 5 * time.Second},
		retry,
		breaker,
		googleOAuthCircuitBreakerKey,
	)
}

// ConfigureGoogleTokenInfoEndpoint keeps token verification injectable for
// tests without changing the production Google endpoint.
func (s *AuthService) ConfigureGoogleTokenInfoEndpoint(endpoint string) {
	if s == nil {
		return
	}
	if trimmed := strings.TrimSpace(endpoint); trimmed != "" {
		s.googleTokenInfoEndpoint = strings.TrimRight(trimmed, "?&")
	}
}

func (s *AuthService) GoogleClientID() string {
	return strings.TrimSpace(s.oauthCfg.GoogleClientID)
}

func (s *AuthService) LoginWithGoogle(ctx context.Context, idToken string) (string, *user.User, error) {
	tokenInfo, err := s.verifyGoogleLoginToken(ctx, idToken)
	if err != nil {
		return "", nil, err
	}

	existingUser, err := s.userRepo.FindByEmail(tokenInfo.Email)
	if err == nil {
		if existingUser.Status != "active" {
			return "", nil, errors.New("user account is not active")
		}
		token, err := s.GenerateToken(existingUser)
		if err != nil {
			return "", nil, err
		}
		return token, existingUser, nil
	}

	createdUser, err := s.createGoogleUser(tokenInfo)
	if err != nil {
		return "", nil, err
	}
	token, err := s.GenerateToken(createdUser)
	if err != nil {
		return "", nil, err
	}
	return token, createdUser, nil
}

func (s *AuthService) LoginBackofficeWithGoogle(ctx context.Context, idToken string) (string, *user.User, error) {
	tokenInfo, err := s.verifyGoogleLoginToken(ctx, idToken)
	if err != nil {
		return "", nil, err
	}

	existingUser, err := s.userRepo.FindByEmail(tokenInfo.Email)
	if err != nil {
		return "", nil, ErrGoogleBackofficeAccessDenied
	}
	if existingUser.Status != "active" {
		return "", nil, errors.New("user account is not active")
	}
	if !auth.IsBackofficeRole(existingUser.Role) {
		return "", nil, ErrGoogleBackofficeAccessDenied
	}

	token, err := s.GenerateToken(existingUser)
	if err != nil {
		return "", nil, err
	}
	return token, existingUser, nil
}

func (s *AuthService) verifyGoogleLoginToken(ctx context.Context, idToken string) (*googleTokenInfo, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, errors.New("google id token is required")
	}
	if s.GoogleClientID() == "" {
		return nil, errors.New("google login is not configured")
	}

	tokenInfo, err := s.verifyGoogleIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return tokenInfo, nil
}

func (s *AuthService) verifyGoogleIDToken(ctx context.Context, idToken string) (*googleTokenInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	endpoint, err := s.googleTokenInfoURL(idToken)
	if err != nil {
		return nil, err
	}

	resp, err := s.googleOAuthClient().Do(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to verify google id token: %w", err)
	}
	if resp == nil {
		return nil, errors.New("failed to verify google id token: empty response")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid google id token")
	}

	var tokenInfo googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse google token response: %w", err)
	}

	tokenInfo.Email = strings.ToLower(strings.TrimSpace(tokenInfo.Email))
	tokenInfo.Subject = strings.TrimSpace(tokenInfo.Subject)
	if tokenInfo.Audience != s.GoogleClientID() {
		return nil, errors.New("google token audience mismatch")
	}
	if tokenInfo.Issuer != "accounts.google.com" && tokenInfo.Issuer != "https://accounts.google.com" {
		return nil, errors.New("google token issuer mismatch")
	}
	if tokenInfo.Email == "" || !strings.EqualFold(tokenInfo.EmailVerified, "true") {
		return nil, errors.New("google email is not verified")
	}
	if tokenInfo.Subject == "" {
		return nil, errors.New("google subject is missing")
	}
	return &tokenInfo, nil
}

func (s *AuthService) googleTokenInfoURL(idToken string) (string, error) {
	endpoint := googleTokenInfoEndpoint
	if s != nil && strings.TrimSpace(s.googleTokenInfoEndpoint) != "" {
		endpoint = strings.TrimSpace(s.googleTokenInfoEndpoint)
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid google tokeninfo endpoint: %w", err)
	}
	query := parsed.Query()
	query.Set("id_token", idToken)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func (s *AuthService) googleOAuthClient() *resilience.HTTPClient {
	if s != nil && s.googleOAuthHTTPClient != nil {
		return s.googleOAuthHTTPClient
	}
	return resilience.NewHTTPClient(
		&http.Client{Timeout: 5 * time.Second},
		resilience.HTTPRetryPolicy{
			MaxAttempts: 3,
			Backoff: resilience.BackoffPolicy{
				BaseDelay: 250 * time.Millisecond,
				MaxDelay:  2 * time.Second,
				Jitter:    250 * time.Millisecond,
			},
			RetryUnsafeMethods: false,
		},
		resilience.SharedCircuitBreaker(
			googleOAuthCircuitBreakerKey,
			resilience.CircuitBreakerConfig{
				FailureThreshold: 4,
				FailureWindow:    60 * time.Second,
				OpenDuration:     30 * time.Second,
			},
		),
		googleOAuthCircuitBreakerKey,
	)
}

func (s *AuthService) createGoogleUser(tokenInfo *googleTokenInfo) (*user.User, error) {
	password, err := randomPassword()
	if err != nil {
		return nil, err
	}

	createdUser := &user.User{
		Email:     tokenInfo.Email,
		Username:  s.googleUsername(tokenInfo.Email, tokenInfo.Subject),
		FirstName: strings.TrimSpace(tokenInfo.GivenName),
		LastName:  strings.TrimSpace(tokenInfo.FamilyName),
		Role:      string(auth.RoleUser),
		Status:    "active",
	}
	if createdUser.FirstName == "" && createdUser.LastName == "" {
		createdUser.FirstName, createdUser.LastName = splitGoogleName(tokenInfo.Name)
	}

	if err := createdUser.HashPassword(password); err != nil {
		return nil, err
	}
	if err := s.userRepo.Create(createdUser); err != nil {
		return nil, err
	}
	return createdUser, nil
}

func (s *AuthService) googleUsername(email string, subject string) string {
	emailPrefix := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")[0]
	base := sanitizeUsername(emailPrefix)
	if len(base) < 3 {
		base = "google_user"
	}

	shortSubject := subject
	if len(shortSubject) > 12 {
		shortSubject = shortSubject[:12]
	}
	candidates := []string{
		base,
		base + "_" + shortSubject,
		"google_" + shortSubject,
	}
	for _, candidate := range candidates {
		if _, err := s.userRepo.FindByUsername(candidate); err != nil {
			return candidate
		}
	}
	return fmt.Sprintf("google_%s_%d", shortSubject, time.Now().Unix())
}

func sanitizeUsername(value string) string {
	re := regexp.MustCompile(`[^a-z0-9_]+`)
	value = re.ReplaceAllString(strings.ToLower(strings.TrimSpace(value)), "_")
	value = strings.Trim(value, "_")
	if len(value) > 50 {
		value = value[:50]
	}
	return value
}

func splitGoogleName(name string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func randomPassword() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

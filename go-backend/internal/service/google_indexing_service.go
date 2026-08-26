package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"commerce-platform/internal/domain/product"
	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/resilience"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const (
	googleIndexingScope      = "https://www.googleapis.com/auth/indexing"
	googleIndexingTokenURL   = "https://oauth2.googleapis.com/token"
	googleIndexingPublishURL = "https://indexing.googleapis.com/v3/urlNotifications:publish"
	googleIndexingNotifyType = "URL_UPDATED"

	googleIndexingCooldownTTL     = 15 * time.Minute
	googleIndexingRedisOpTimeout  = 2 * time.Second
	googleIndexingCooldownKeyBase = "commerce_platform:google_indexing:cooldown:"
)

var (
	releaseGoogleIndexingCooldownScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) ~= ARGV[1] then
	return 0
end
return redis.call("DEL", KEYS[1])
`)

	ErrGoogleIndexingDisabled         = errors.New("Google Indexing is disabled")
	ErrGoogleIndexingNotConfigured    = errors.New("Google Indexing is not configured")
	ErrGoogleIndexingProductNotFound  = errors.New("product not found for Google Indexing")
	ErrGoogleIndexingProductNotPublic = errors.New("product must be active before Google Indexing notification")
	ErrGoogleIndexingInvalidURL       = errors.New("product URL is invalid for Google Indexing")
	ErrGoogleIndexingUpstream         = errors.New("Google Indexing upstream request failed")
	ErrGoogleIndexingRecentlyNotified = errors.New("Google Indexing notification was submitted recently")
	ErrGoogleIndexingProtection       = errors.New("Google Indexing duplicate protection is unavailable")
)

type GoogleIndexingCooldownError struct {
	RetryAfter time.Duration
}

func (e *GoogleIndexingCooldownError) Error() string {
	return ErrGoogleIndexingRecentlyNotified.Error()
}

func (e *GoogleIndexingCooldownError) Unwrap() error {
	return ErrGoogleIndexingRecentlyNotified
}

type googleIndexingProductReader interface {
	GetAdminProduct(id uint) (*product.Product, error)
}

type googleIndexingCredentials struct {
	ClientEmail  string
	PrivateKey   *rsa.PrivateKey
	PrivateKeyID string
	TokenURI     string
}

type googleIndexingToken struct {
	Value     string
	ExpiresAt time.Time
}

type GoogleIndexingService struct {
	products      googleIndexingProductReader
	config        config.GoogleIndexingConfig
	storefrontURL string
	credentials   googleIndexingCredentials

	httpClient      *resilience.HTTPClient
	oauthHTTPClient *resilience.HTTPClient
	publishURL      string
	redisClient     redis.UniversalClient
	tokenMu         sync.Mutex
	token           googleIndexingToken
	now             func() time.Time
}

type GoogleIndexingStatus struct {
	Enabled    bool   `json:"enabled"`
	Configured bool   `json:"configured"`
	Ready      bool   `json:"ready"`
	Message    string `json:"message"`
}

type GoogleIndexingPushResult struct {
	ProductID        uint                   `json:"product_id"`
	URL              string                 `json:"url"`
	NotificationType string                 `json:"notification_type"`
	Accepted         bool                   `json:"accepted"`
	HTTPStatus       int                    `json:"http_status"`
	SubmittedAt      time.Time              `json:"submitted_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type googleIndexingServiceAccountJSON struct {
	Type         string `json:"type"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	TokenURI     string `json:"token_uri"`
}

func NewGoogleIndexingService(
	products googleIndexingProductReader,
	googleConfig config.GoogleIndexingConfig,
	storefrontURL string,
) (*GoogleIndexingService, error) {
	service := &GoogleIndexingService{
		products:      products,
		config:        googleConfig,
		storefrontURL: strings.TrimRight(strings.TrimSpace(storefrontURL), "/"),
		publishURL:    googleIndexingPublishURL,
		now:           time.Now,
	}
	service.configureDefaultHTTPClients()

	if !googleConfig.Enabled {
		return service, nil
	}

	credentials, err := loadGoogleIndexingCredentials(googleConfig)
	if err != nil {
		return nil, err
	}
	service.credentials = credentials
	return service, nil
}

func (s *GoogleIndexingService) ConfigureOutboundHTTPResilience(
	retry resilience.HTTPRetryPolicy,
	breaker resilience.CircuitController,
) {
	if s == nil {
		return
	}
	timeout := time.Duration(s.config.RequestTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	s.httpClient = resilience.NewHTTPClient(
		&http.Client{Timeout: timeout},
		retry,
		breaker,
		"google-indexing-publish",
	)
	s.oauthHTTPClient = resilience.NewHTTPClient(
		&http.Client{Timeout: timeout},
		retry,
		breaker,
		"google-indexing-oauth",
	)
}

func (s *GoogleIndexingService) ConfigureRedisClient(redisClient redis.UniversalClient) {
	if s == nil {
		return
	}
	s.redisClient = redisClient
}

func (s *GoogleIndexingService) Status() GoogleIndexingStatus {
	if s == nil {
		return GoogleIndexingStatus{
			Enabled:    false,
			Configured: false,
			Ready:      false,
			Message:    "Google Indexing service is unavailable",
		}
	}

	configured := strings.TrimSpace(s.config.ServiceAccountJSON) != "" ||
		strings.TrimSpace(s.config.ServiceAccountFile) != ""
	if !s.config.Enabled {
		return GoogleIndexingStatus{
			Enabled:    false,
			Configured: configured,
			Ready:      false,
			Message:    "Google Indexing is disabled",
		}
	}
	if strings.TrimSpace(s.credentials.ClientEmail) == "" || s.credentials.PrivateKey == nil {
		return GoogleIndexingStatus{
			Enabled:    true,
			Configured: configured,
			Ready:      false,
			Message:    "Google service account credentials are not ready",
		}
	}
	return GoogleIndexingStatus{
		Enabled:    true,
		Configured: true,
		Ready:      true,
		Message:    "Google Indexing is ready",
	}
}

func (s *GoogleIndexingService) PushProduct(ctx context.Context, productID uint) (*GoogleIndexingPushResult, error) {
	if s == nil {
		return nil, ErrGoogleIndexingNotConfigured
	}
	status := s.Status()
	if !status.Enabled {
		return nil, ErrGoogleIndexingDisabled
	}
	if !status.Ready {
		return nil, ErrGoogleIndexingNotConfigured
	}
	if productID == 0 || s.products == nil {
		return nil, ErrGoogleIndexingProductNotFound
	}

	item, err := s.products.GetAdminProduct(productID)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return nil, ErrGoogleIndexingProductNotFound
		}
		return nil, err
	}
	if item == nil {
		return nil, ErrGoogleIndexingProductNotFound
	}
	if strings.TrimSpace(item.Status) != "active" {
		return nil, ErrGoogleIndexingProductNotPublic
	}

	productURL, err := s.productURL(item)
	if err != nil {
		return nil, err
	}

	result := &GoogleIndexingPushResult{
		ProductID:        item.ID,
		URL:              productURL,
		NotificationType: googleIndexingNotifyType,
	}
	if s.redisClient == nil {
		return result, ErrGoogleIndexingProtection
	}

	if ctx == nil {
		ctx = context.Background()
	}
	reservationToken, err := s.reserveCooldown(ctx, productURL)
	if err != nil {
		return result, err
	}

	accessToken, err := s.accessToken(ctx)
	if err != nil {
		s.releaseCooldown(productURL, reservationToken)
		return result, err
	}

	result.SubmittedAt = s.currentTime().UTC()
	httpStatus, metadata, err := s.publish(ctx, accessToken, productURL)
	result.HTTPStatus = httpStatus
	result.Metadata = metadata
	if errors.Is(err, errGoogleIndexingUnauthorized) {
		s.clearAccessToken()
		accessToken, refreshErr := s.accessToken(ctx)
		if refreshErr != nil {
			return result, refreshErr
		}
		httpStatus, metadata, err = s.publish(ctx, accessToken, productURL)
		result.HTTPStatus = httpStatus
		result.Metadata = metadata
	}
	if err != nil {
		return result, err
	}

	result.Accepted = httpStatus >= http.StatusOK && httpStatus < http.StatusMultipleChoices
	return result, nil
}

func (s *GoogleIndexingService) reserveCooldown(ctx context.Context, productURL string) (string, error) {
	if s == nil || s.redisClient == nil {
		return "", ErrGoogleIndexingProtection
	}
	if ctx == nil {
		ctx = context.Background()
	}

	operationCtx, cancel := context.WithTimeout(ctx, googleIndexingRedisOpTimeout)
	defer cancel()

	reservationToken := newGoogleIndexingReservationToken()
	key := googleIndexingCooldownKey(productURL)
	accepted, err := s.redisClient.SetNX(operationCtx, key, reservationToken, googleIndexingCooldownTTL).Result()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGoogleIndexingProtection, err)
	}
	if accepted {
		return reservationToken, nil
	}

	ttl, err := s.redisClient.TTL(operationCtx, key).Result()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGoogleIndexingProtection, err)
	}
	if ttl <= 0 {
		ttl = googleIndexingCooldownTTL
	}
	return "", &GoogleIndexingCooldownError{RetryAfter: ttl}
}

func (s *GoogleIndexingService) releaseCooldown(productURL, reservationToken string) {
	if s == nil || s.redisClient == nil || strings.TrimSpace(reservationToken) == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), googleIndexingRedisOpTimeout)
	defer cancel()
	_, _ = releaseGoogleIndexingCooldownScript.Run(
		ctx,
		s.redisClient,
		[]string{googleIndexingCooldownKey(productURL)},
		reservationToken,
	).Result()
}

func newGoogleIndexingReservationToken() string {
	var token [16]byte
	if _, err := rand.Read(token[:]); err == nil {
		return hex.EncodeToString(token[:])
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(sum[:])
}

func googleIndexingCooldownKey(productURL string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(productURL)))
	return googleIndexingCooldownKeyBase + hex.EncodeToString(sum[:])
}

func (s *GoogleIndexingService) productURL(item *product.Product) (string, error) {
	if item == nil {
		return "", ErrGoogleIndexingProductNotFound
	}
	routePath := seodomain.BuildProductRoute(item.Locale, item.Slug).Path
	if routePath == "" {
		return "", ErrGoogleIndexingInvalidURL
	}
	base, err := url.Parse(s.storefrontURL)
	if err != nil || base.Host == "" ||
		(base.Scheme != "http" && base.Scheme != "https") {
		return "", fmt.Errorf("%w: storefront base URL is invalid", ErrGoogleIndexingInvalidURL)
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/" + strings.TrimLeft(routePath, "/")
	base.RawQuery = ""
	base.Fragment = ""
	return base.String(), nil
}

func (s *GoogleIndexingService) accessToken(ctx context.Context) (string, error) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	now := s.currentTime()
	if s.token.Value != "" && now.Add(60*time.Second).Before(s.token.ExpiresAt) {
		return s.token.Value, nil
	}

	assertion, err := s.serviceAccountAssertion(now)
	if err != nil {
		return "", err
	}
	form := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {assertion},
	}
	requestURL := s.credentials.TokenURI
	if requestURL == "" {
		requestURL = googleIndexingTokenURL
	}

	response, err := s.oauthHTTPClient.Do(ctx, func() (*http.Request, error) {
		request, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return request, nil
	})
	if err != nil {
		return "", fmt.Errorf("%w: token request: %v", ErrGoogleIndexingUpstream, err)
	}
	body, readErr := readGoogleIndexingResponseBody(response)
	if readErr != nil {
		return "", fmt.Errorf("%w: token response: %v", ErrGoogleIndexingUpstream, readErr)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", googleIndexingUpstreamResponseError(response.StatusCode, body)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("%w: invalid token response: %v", ErrGoogleIndexingUpstream, err)
	}
	if strings.TrimSpace(tokenResponse.AccessToken) == "" || tokenResponse.ExpiresIn <= 0 {
		return "", fmt.Errorf("%w: token response is missing access_token or expires_in", ErrGoogleIndexingUpstream)
	}
	s.token = googleIndexingToken{
		Value:     tokenResponse.AccessToken,
		ExpiresAt: now.Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}
	return s.token.Value, nil
}

var errGoogleIndexingUnauthorized = errors.New("Google Indexing authorization rejected")

func (s *GoogleIndexingService) publish(
	ctx context.Context,
	accessToken string,
	productURL string,
) (int, map[string]interface{}, error) {
	payload, err := json.Marshal(map[string]string{
		"url":  productURL,
		"type": googleIndexingNotifyType,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("%w: build notification: %v", ErrGoogleIndexingUpstream, err)
	}

	publishURL := s.publishURL
	if strings.TrimSpace(publishURL) == "" {
		publishURL = googleIndexingPublishURL
	}
	response, err := s.httpClient.Do(ctx, func() (*http.Request, error) {
		request, err := http.NewRequest(http.MethodPost, publishURL, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		request.Header.Set("Authorization", "Bearer "+accessToken)
		request.Header.Set("Content-Type", "application/json")
		return request, nil
	})
	if err != nil {
		return 0, nil, fmt.Errorf("%w: publish request: %v", ErrGoogleIndexingUpstream, err)
	}
	body, readErr := readGoogleIndexingResponseBody(response)
	if readErr != nil {
		return response.StatusCode, nil, fmt.Errorf("%w: publish response: %v", ErrGoogleIndexingUpstream, readErr)
	}
	if response.StatusCode == http.StatusUnauthorized {
		return response.StatusCode, nil, errGoogleIndexingUnauthorized
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response.StatusCode, nil, googleIndexingUpstreamResponseError(response.StatusCode, body)
	}

	var envelope struct {
		Metadata map[string]interface{} `json:"urlNotificationMetadata"`
	}
	if len(bytes.TrimSpace(body)) > 0 {
		if err := json.Unmarshal(body, &envelope); err != nil {
			return response.StatusCode, nil, fmt.Errorf("%w: invalid publish response: %v", ErrGoogleIndexingUpstream, err)
		}
	}
	return response.StatusCode, envelope.Metadata, nil
}

func (s *GoogleIndexingService) serviceAccountAssertion(now time.Time) (string, error) {
	if strings.TrimSpace(s.credentials.ClientEmail) == "" || s.credentials.PrivateKey == nil {
		return "", ErrGoogleIndexingNotConfigured
	}
	claims := jwt.MapClaims{
		"iss":   s.credentials.ClientEmail,
		"scope": googleIndexingScope,
		"aud":   s.credentials.TokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if keyID := strings.TrimSpace(s.credentials.PrivateKeyID); keyID != "" {
		token.Header["kid"] = keyID
	}
	signed, err := token.SignedString(s.credentials.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("%w: sign service account assertion: %v", ErrGoogleIndexingUpstream, err)
	}
	return signed, nil
}

func (s *GoogleIndexingService) clearAccessToken() {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()
	s.token = googleIndexingToken{}
}

func (s *GoogleIndexingService) configureDefaultHTTPClients() {
	timeout := time.Duration(s.config.RequestTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	retry := resilience.HTTPRetryPolicy{
		MaxAttempts:        1,
		RetryUnsafeMethods: false,
	}
	s.httpClient = resilience.NewHTTPClient(&http.Client{Timeout: timeout}, retry, nil, "google-indexing-publish")
	s.oauthHTTPClient = resilience.NewHTTPClient(&http.Client{Timeout: timeout}, retry, nil, "google-indexing-oauth")
}

func (s *GoogleIndexingService) currentTime() time.Time {
	if s != nil && s.now != nil {
		return s.now()
	}
	return time.Now()
}

func loadGoogleIndexingCredentials(cfg config.GoogleIndexingConfig) (googleIndexingCredentials, error) {
	raw := strings.TrimSpace(cfg.ServiceAccountJSON)
	if raw == "" {
		filePath := strings.TrimSpace(cfg.ServiceAccountFile)
		if filePath == "" {
			return googleIndexingCredentials{}, ErrGoogleIndexingNotConfigured
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return googleIndexingCredentials{}, fmt.Errorf("read Google Indexing service account file: %w", err)
		}
		raw = strings.TrimSpace(string(content))
	}
	if raw == "" {
		return googleIndexingCredentials{}, ErrGoogleIndexingNotConfigured
	}

	var account googleIndexingServiceAccountJSON
	if err := json.Unmarshal([]byte(raw), &account); err != nil {
		return googleIndexingCredentials{}, fmt.Errorf("parse Google Indexing service account JSON: %w", err)
	}
	if strings.TrimSpace(account.ClientEmail) == "" || strings.TrimSpace(account.PrivateKey) == "" {
		return googleIndexingCredentials{}, errors.New("Google Indexing service account JSON must contain client_email and private_key")
	}
	privateKey, err := parseGoogleIndexingPrivateKey(account.PrivateKey)
	if err != nil {
		return googleIndexingCredentials{}, fmt.Errorf("parse Google Indexing service account private key: %w", err)
	}
	tokenURI := strings.TrimSpace(account.TokenURI)
	if tokenURI == "" {
		tokenURI = googleIndexingTokenURL
	}
	parsedTokenURI, err := url.Parse(tokenURI)
	if err != nil || parsedTokenURI.Host == "" ||
		(parsedTokenURI.Scheme != "http" && parsedTokenURI.Scheme != "https") {
		return googleIndexingCredentials{}, errors.New("Google Indexing service account token_uri must be an absolute HTTP(S) URL")
	}
	return googleIndexingCredentials{
		ClientEmail:  strings.TrimSpace(account.ClientEmail),
		PrivateKey:   privateKey,
		PrivateKeyID: strings.TrimSpace(account.PrivateKeyID),
		TokenURI:     tokenURI,
	}, nil
}

func parseGoogleIndexingPrivateKey(value string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, errors.New("private_key does not contain PEM data")
	}
	if parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if key, ok := parsed.(*rsa.PrivateKey); ok {
			return key, nil
		}
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, errors.New("private_key is not a supported RSA private key")
}

func readGoogleIndexingResponseBody(response *http.Response) ([]byte, error) {
	if response == nil || response.Body == nil {
		return nil, errors.New("empty upstream response")
	}
	defer response.Body.Close()
	return io.ReadAll(io.LimitReader(response.Body, 1<<20))
}

func googleIndexingUpstreamResponseError(statusCode int, body []byte) error {
	detail := strings.TrimSpace(string(body))
	if len(detail) > 500 {
		detail = detail[:500]
	}
	if detail == "" {
		return fmt.Errorf("%w: Google Indexing API returned HTTP %d", ErrGoogleIndexingUpstream, statusCode)
	}
	return fmt.Errorf("%w: Google Indexing API returned HTTP %d: %s", ErrGoogleIndexingUpstream, statusCode, detail)
}

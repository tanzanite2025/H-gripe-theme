package referral

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/loyalty"
)

const (
	CookieName       = "storefront_referral"
	maximumCookieTTL = 90 * 24 * time.Hour
)

var (
	ErrSignerSecretRequired = errors.New("referral cookie signer secret is required")
	ErrInvalidCookie        = errors.New("invalid referral cookie")
)

type Claims struct {
	Code       string    `json:"code"`
	Source     string    `json:"source"`
	CapturedAt time.Time `json:"captured_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type Signer struct {
	secret []byte
	now    func() time.Time
}

func NewSigner(secret string) (*Signer, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, ErrSignerSecretRequired
	}
	return &Signer{secret: []byte(secret), now: func() time.Time { return time.Now().UTC() }}, nil
}

func (s *Signer) Encode(code, source string, ttl time.Duration) (string, Claims, error) {
	if s == nil || len(s.secret) == 0 {
		return "", Claims{}, ErrSignerSecretRequired
	}
	code = loyalty.NormalizeReferralCode(code)
	if err := loyalty.ValidateReferralCode(code); err != nil {
		return "", Claims{}, err
	}
	source = strings.TrimSpace(source)
	if source != "link" && source != "manual_input" {
		return "", Claims{}, ErrInvalidCookie
	}
	if ttl <= 0 || ttl > maximumCookieTTL {
		return "", Claims{}, ErrInvalidCookie
	}
	now := s.now().UTC()
	claims := Claims{Code: code, Source: source, CapturedAt: now, ExpiresAt: now.Add(ttl)}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", Claims{}, err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	signature := base64.RawURLEncoding.EncodeToString(s.sign(encoded))
	return encoded + "." + signature, claims, nil
}

func (s *Signer) Decode(token string) (Claims, error) {
	if s == nil || len(s.secret) == 0 {
		return Claims{}, ErrSignerSecretRequired
	}
	encoded, encodedSignature, ok := strings.Cut(strings.TrimSpace(token), ".")
	if !ok || encoded == "" || encodedSignature == "" {
		return Claims{}, ErrInvalidCookie
	}
	providedSignature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return Claims{}, ErrInvalidCookie
	}
	expectedSignature := s.sign(encoded)
	if len(providedSignature) != len(expectedSignature) || subtle.ConstantTimeCompare(providedSignature, expectedSignature) != 1 {
		return Claims{}, ErrInvalidCookie
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return Claims{}, ErrInvalidCookie
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrInvalidCookie
	}
	claims.Code = loyalty.NormalizeReferralCode(claims.Code)
	if err := loyalty.ValidateReferralCode(claims.Code); err != nil {
		return Claims{}, ErrInvalidCookie
	}
	if claims.Source != "link" && claims.Source != "manual_input" {
		return Claims{}, ErrInvalidCookie
	}
	if claims.CapturedAt.IsZero() || claims.ExpiresAt.IsZero() || !claims.ExpiresAt.After(s.now().UTC()) ||
		claims.ExpiresAt.Sub(claims.CapturedAt) > maximumCookieTTL {
		return Claims{}, ErrInvalidCookie
	}
	claims.CapturedAt = claims.CapturedAt.UTC()
	claims.ExpiresAt = claims.ExpiresAt.UTC()
	return claims, nil
}

func (s *Signer) sign(value string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte("referral-cookie:v1:" + value))
	return mac.Sum(nil)
}

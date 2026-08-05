package attribution

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	CookieName   = "storefront_attribution"
	CookieMaxAge = 30 * 24 * 60 * 60
)

var (
	ErrSignerSecretRequired = errors.New("attribution signer secret is required")
	ErrInvalidContext       = errors.New("invalid attribution context")
)

// Context contains a bounded first-party advertising source signal. It is not
// proof of a conversion; payment verification remains the conversion authority.
type Context struct {
	Source      string    `json:"source,omitempty"`
	Medium      string    `json:"medium,omitempty"`
	Campaign    string    `json:"campaign,omitempty"`
	Term        string    `json:"term,omitempty"`
	Content     string    `json:"content,omitempty"`
	ClickIDKind string    `json:"click_id_kind,omitempty"`
	ClickID     string    `json:"click_id,omitempty"`
	CapturedAt  time.Time `json:"captured_at"`
	ExpiresAt   time.Time `json:"expires_at"`
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
	return &Signer{
		secret: []byte(secret),
		now:    func() time.Time { return time.Now().UTC() },
	}, nil
}

func (s *Signer) Encode(value Context) (string, error) {
	if s == nil || len(s.secret) == 0 {
		return "", ErrSignerSecretRequired
	}

	normalized, ok := Normalize(value)
	if !ok {
		return "", ErrInvalidContext
	}
	now := s.now().UTC()
	normalized.CapturedAt = now
	normalized.ExpiresAt = now.Add(CookieMaxAge * time.Second)

	payload, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := s.sign(encodedPayload)
	return encodedPayload + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (s *Signer) Decode(token string) (Context, error) {
	if s == nil || len(s.secret) == 0 {
		return Context{}, ErrSignerSecretRequired
	}

	encodedPayload, encodedSignature, ok := strings.Cut(strings.TrimSpace(token), ".")
	if !ok || encodedPayload == "" || encodedSignature == "" {
		return Context{}, ErrInvalidContext
	}
	providedSignature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return Context{}, ErrInvalidContext
	}
	expectedSignature := s.sign(encodedPayload)
	if len(providedSignature) != len(expectedSignature) ||
		subtle.ConstantTimeCompare(providedSignature, expectedSignature) != 1 {
		return Context{}, ErrInvalidContext
	}

	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return Context{}, ErrInvalidContext
	}
	var value Context
	if err := json.Unmarshal(payload, &value); err != nil {
		return Context{}, ErrInvalidContext
	}
	normalized, ok := Normalize(value)
	if !ok || value.CapturedAt.IsZero() || value.ExpiresAt.IsZero() || !value.ExpiresAt.After(s.now().UTC()) {
		return Context{}, ErrInvalidContext
	}
	normalized.CapturedAt = value.CapturedAt.UTC()
	normalized.ExpiresAt = value.ExpiresAt.UTC()
	return normalized, nil
}

func (s *Signer) sign(value string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

// FromMetadata accepts only the allowlisted first-party campaign fields.
func FromMetadata(metadata map[string]any) (Context, bool) {
	if len(metadata) == 0 {
		return Context{}, false
	}
	value := Context{
		Source:   metadataString(metadata, "utm_source"),
		Medium:   metadataString(metadata, "utm_medium"),
		Campaign: metadataString(metadata, "utm_campaign"),
		Term:     metadataString(metadata, "utm_term"),
		Content:  metadataString(metadata, "utm_content"),
	}
	for _, candidate := range []struct {
		key    string
		kind   string
		source string
	}{
		{key: "gclid", kind: "gclid", source: "google"},
		{key: "fbclid", kind: "fbclid", source: "facebook"},
		{key: "msclkid", kind: "msclkid", source: "microsoft"},
		{key: "ttclid", kind: "ttclid", source: "tiktok"},
	} {
		if clickID := metadataString(metadata, candidate.key); clickID != "" {
			value.ClickIDKind = candidate.kind
			value.ClickID = clickID
			if value.Source == "" {
				value.Source = candidate.source
			}
			break
		}
	}
	return Normalize(value)
}

func (c Context) Metadata() map[string]any {
	metadata := map[string]any{}
	if c.Source != "" {
		metadata["utm_source"] = c.Source
	}
	if c.Medium != "" {
		metadata["utm_medium"] = c.Medium
	}
	if c.Campaign != "" {
		metadata["utm_campaign"] = c.Campaign
	}
	if c.Term != "" {
		metadata["utm_term"] = c.Term
	}
	if c.Content != "" {
		metadata["utm_content"] = c.Content
	}
	if c.ClickIDKind != "" && c.ClickID != "" {
		metadata[c.ClickIDKind] = c.ClickID
	}
	return metadata
}

func Normalize(value Context) (Context, bool) {
	value.Source = normalizeValue(value.Source, 96)
	value.Medium = normalizeValue(value.Medium, 96)
	value.Campaign = normalizeValue(value.Campaign, 160)
	value.Term = normalizeValue(value.Term, 160)
	value.Content = normalizeValue(value.Content, 160)
	value.ClickIDKind = normalizeClickIDKind(value.ClickIDKind)
	value.ClickID = normalizeValue(value.ClickID, 256)
	if value.ClickIDKind == "" {
		value.ClickID = ""
	}
	if value.ClickIDKind != "" && value.ClickID == "" {
		value.ClickIDKind = ""
	}
	if value.Source == "" && value.Medium == "" && value.Campaign == "" && value.ClickID == "" {
		return Context{}, false
	}
	return value, true
}

func metadataString(metadata map[string]any, key string) string {
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func normalizeClickIDKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "gclid", "fbclid", "msclkid", "ttclid":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeValue(value string, limit int) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, value)
	if len(value) > limit {
		value = value[:limit]
	}
	return value
}

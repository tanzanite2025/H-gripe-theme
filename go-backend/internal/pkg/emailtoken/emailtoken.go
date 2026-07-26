package emailtoken

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrMissingSecret = errors.New("email token secret is not configured")
	ErrInvalidToken  = errors.New("invalid email token")
	ErrExpiredToken  = errors.New("email token has expired")
)

type Claims struct {
	Purpose   string `json:"purpose"`
	Email     string `json:"email"`
	Subject   string `json:"subject"`
	Nonce     string `json:"nonce"`
	ExpiresAt int64  `json:"exp"`
}

func Sign(secret string, claims Claims) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", ErrMissingSecret
	}
	if strings.TrimSpace(claims.Purpose) == "" || strings.TrimSpace(claims.Subject) == "" || claims.ExpiresAt <= 0 {
		return "", ErrInvalidToken
	}
	if claims.Nonce == "" {
		nonce, err := randomNonce()
		if err != nil {
			return "", err
		}
		claims.Nonce = nonce
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := signPayload(secret, encodedPayload)
	return encodedPayload + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func Verify(secret, token, purpose string, now time.Time) (Claims, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return Claims{}, ErrMissingSecret
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Claims{}, ErrInvalidToken
	}

	expected := signPayload(secret, parts[0])
	actual, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(expected, actual) {
		return Claims{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if claims.Purpose != purpose || claims.Subject == "" || claims.ExpiresAt <= 0 {
		return Claims{}, ErrInvalidToken
	}
	if now.Unix() >= claims.ExpiresAt {
		return Claims{}, ErrExpiredToken
	}

	return claims, nil
}

func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func signPayload(secret, payload string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return mac.Sum(nil)
}

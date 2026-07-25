package visitorcookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CustomerServiceVisitorCookie = "tz_customer_service_visitor"
	CustomerServiceVisitorMaxAge = 86400 * 365
)

func EnsureCustomerServiceVisitorHash(c *gin.Context, secret []byte) (string, bool) {
	sessionID, ok := ReadCustomerServiceVisitorSessionID(c, secret)
	if !ok {
		sessionID = uuid.NewString()
	}
	SetCustomerServiceVisitorCookie(c, sessionID, secret)
	return CustomerServiceVisitorHash(sessionID), true
}

func ExistingCustomerServiceVisitorHash(c *gin.Context, secret []byte) (string, bool) {
	sessionID, ok := ReadCustomerServiceVisitorSessionID(c, secret)
	if !ok {
		return "", false
	}
	return CustomerServiceVisitorHash(sessionID), true
}

func ReadCustomerServiceVisitorSessionID(c *gin.Context, secret []byte) (string, bool) {
	rawCookie, err := c.Cookie(CustomerServiceVisitorCookie)
	if err != nil {
		return "", false
	}
	rawCookie = strings.TrimSpace(rawCookie)
	if rawCookie == "" {
		return "", false
	}

	sessionID, signature, signed := strings.Cut(rawCookie, ".")
	sessionID = strings.TrimSpace(sessionID)
	if _, err := uuid.Parse(sessionID); err != nil {
		return "", false
	}
	if !signed || strings.TrimSpace(signature) == "" {
		return "", false
	}
	if ValidCustomerServiceVisitorSignature(sessionID, signature, secret) {
		return sessionID, true
	}
	return "", false
}

func SetCustomerServiceVisitorCookie(c *gin.Context, sessionID string, secret []byte) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(CustomerServiceVisitorCookie, SignCustomerServiceVisitorSessionID(sessionID, secret), CustomerServiceVisitorMaxAge, "/", "", cookieSecure(c), true)
}

func SignCustomerServiceVisitorSessionID(sessionID string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(sessionID))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return sessionID + "." + signature
}

func ValidCustomerServiceVisitorSignature(sessionID string, signature string, secret []byte) bool {
	expected := SignCustomerServiceVisitorSessionID(sessionID, secret)
	_, expectedSignature, _ := strings.Cut(expected, ".")
	if len(signature) != len(expectedSignature) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) == 1
}

func CustomerServiceVisitorHash(sessionID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(sessionID)))
	return hex.EncodeToString(sum[:])
}

func cookieSecure(c *gin.Context) bool {
	return c.Request != nil && (c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https"))
}

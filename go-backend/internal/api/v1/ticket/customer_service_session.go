package ticket

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	customerServiceVisitorCookie = "tz_customer_service_visitor"
	customerServiceVisitorMaxAge = 86400 * 365
)

func publicCustomerUserID(c *gin.Context) *uint {
	value, exists := c.Get("user_id")
	if !exists {
		return nil
	}
	userID, ok := value.(uint)
	if !ok {
		return nil
	}
	return &userID
}

func (h *Handler) publicCustomerOwner(c *gin.Context) service.CustomerServiceOwner {
	return service.CustomerServiceOwner{
		UserID:             publicCustomerUserID(c),
		VisitorSessionHash: h.ensureVisitorSessionHash(c),
	}
}

func (h *Handler) ensureVisitorSessionHash(c *gin.Context) string {
	hash, _ := h.visitorSessionHash(c, true)
	return hash
}

func (h *Handler) existingVisitorSessionHash(c *gin.Context) (string, bool) {
	return h.visitorSessionHash(c, false)
}

func (h *Handler) visitorSessionHash(c *gin.Context, create bool) (string, bool) {
	sessionID, ok := h.readVisitorSessionID(c)
	if !ok {
		if !create {
			return "", false
		}
		sessionID = uuid.NewString()
	}
	h.setVisitorSessionCookie(c, sessionID)
	sum := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(sum[:]), true
}

func (h *Handler) readVisitorSessionID(c *gin.Context) (string, bool) {
	rawCookie, err := c.Cookie(customerServiceVisitorCookie)
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
	if h.validVisitorSignature(sessionID, signature) {
		return sessionID, true
	}
	return "", false
}

func (h *Handler) setVisitorSessionCookie(c *gin.Context, sessionID string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(customerServiceVisitorCookie, h.signVisitorSessionID(sessionID), customerServiceVisitorMaxAge, "/", "", visitorCookieSecure(c), true)
}

func (h *Handler) signVisitorSessionID(sessionID string) string {
	mac := hmac.New(sha256.New, h.visitorSecret)
	_, _ = mac.Write([]byte(sessionID))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return sessionID + "." + signature
}

func (h *Handler) validVisitorSignature(sessionID string, signature string) bool {
	expected := h.signVisitorSessionID(sessionID)
	_, expectedSignature, _ := strings.Cut(expected, ".")
	if len(signature) != len(expectedSignature) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) == 1
}

func visitorCookieSecure(c *gin.Context) bool {
	return c.Request != nil && (c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https"))
}

func parseCustomerServiceAgentID(value string) uint {
	if strings.TrimSpace(value) == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil {
		return 0
	}
	return uint(parsed)
}

func publicConversationID(item *ticket.Ticket) string {
	if item == nil {
		return ""
	}
	if item.ConversationID != nil && strings.TrimSpace(*item.ConversationID) != "" {
		return strings.TrimSpace(*item.ConversationID)
	}
	if strings.HasPrefix(item.Tags, "conversation_id:") {
		return strings.TrimPrefix(item.Tags, "conversation_id:")
	}
	return ""
}

func writePublicCustomerServiceError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrCustomerServiceConversationAccessDenied) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "conversation access denied"})
		return
	}
	if errors.Is(err, service.ErrCustomerServiceOwnerRequired) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "conversation owner is required"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
}

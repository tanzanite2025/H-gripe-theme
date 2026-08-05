package ticket

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/pkg/visitorcookie"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	customerServiceVisitorCookie = visitorcookie.CustomerServiceVisitorCookie
	customerServiceVisitorMaxAge = visitorcookie.CustomerServiceVisitorMaxAge
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

func (h *Handler) existingPublicCustomerOwner(c *gin.Context) service.CustomerServiceOwner {
	visitorHash, _ := h.existingVisitorSessionHash(c)
	return service.CustomerServiceOwner{
		UserID:             publicCustomerUserID(c),
		VisitorSessionHash: visitorHash,
	}
}

func (h *Handler) ensureVisitorSessionHash(c *gin.Context) string {
	hash, _ := visitorcookie.EnsureCustomerServiceVisitorHash(c, h.visitorSecret)
	return hash
}

func (h *Handler) existingVisitorSessionHash(c *gin.Context) (string, bool) {
	return visitorcookie.ExistingCustomerServiceVisitorHash(c, h.visitorSecret)
}

func (h *Handler) readVisitorSessionID(c *gin.Context) (string, bool) {
	return visitorcookie.ReadCustomerServiceVisitorSessionID(c, h.visitorSecret)
}

func (h *Handler) signVisitorSessionID(sessionID string) string {
	return visitorcookie.SignCustomerServiceVisitorSessionID(sessionID, h.visitorSecret)
}

func (h *Handler) validVisitorSignature(sessionID string, signature string) bool {
	return visitorcookie.ValidCustomerServiceVisitorSignature(sessionID, signature, h.visitorSecret)
}

func (h *Handler) touchCustomerServiceVisitorProfile(c *gin.Context, owner service.CustomerServiceOwner, email string, emailSource string) {
	if h.visitorProfileService == nil {
		return
	}

	meaningfulAction := service.VisitorProfileActionCustomerService
	qualityScore := service.VisitorProfileQualityCustomerService
	if strings.TrimSpace(email) != "" {
		meaningfulAction = service.VisitorProfileActionEmailCapture
		qualityScore = service.VisitorProfileQualityEmailCapture
	}

	input := service.VisitorProfileTouchInput{
		UserID:                     owner.UserID,
		CustomerServiceVisitorHash: owner.VisitorSessionHash,
		CartSessionID:              existingCartSessionID(c),
		Email:                      email,
		EmailSource:                emailSource,
		Locale:                     requestLocale(c),
		LocaleSource:               "accept_language",
		CountryCode:                requestCountryCode(c),
		Region:                     firstNonEmptyHeader(c, "CF-Region", "X-Region"),
		City:                       firstNonEmptyHeader(c, "CF-IPCity", "X-City"),
		Timezone:                   firstNonEmptyHeader(c, "CF-Timezone", "X-Timezone"),
		IPAddress:                  requestIP(c),
		UserAgent:                  c.GetHeader("User-Agent"),
		MeaningfulAction:           meaningfulAction,
		QualityScoreDelta:          qualityScore,
	}
	if _, err := h.visitorProfileService.TouchMeaningfulAction(input); err != nil {
		return
	}
}

func existingCartSessionID(c *gin.Context) string {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(sessionID)
}

func requestLocale(c *gin.Context) string {
	return firstNonEmptyHeader(c, "X-Locale", "Accept-Language")
}

func requestCountryCode(c *gin.Context) string {
	return middleware.TrustedEdgeCountry(c)
}

func requestIP(c *gin.Context) string {
	if c.Request == nil {
		return ""
	}
	return c.ClientIP()
}

func firstNonEmptyHeader(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.GetHeader(key)); value != "" {
			return value
		}
	}
	return ""
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

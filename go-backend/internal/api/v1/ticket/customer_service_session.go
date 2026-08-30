package ticket

import (
	"commerce-platform/internal/api/v1/visitorcapture"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/service"
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	input := visitorcapture.BuildVisitorProfileTouchInput(c, visitorcapture.TouchOptions{
		UserID:                     owner.UserID,
		CustomerServiceVisitorHash: owner.VisitorSessionHash,
		CartSessionID:              visitorcapture.ExistingCartSessionID(c),
		Email:                      email,
		EmailSource:                emailSource,
		MeaningfulAction:           meaningfulAction,
		QualityScoreDelta:          qualityScore,
	})
	if _, err := h.visitorProfileService.TouchMeaningfulAction(input); err != nil {
		return
	}
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

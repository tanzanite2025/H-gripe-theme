package admin

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"commerce-platform/internal/api/restrictedmediaresponse"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *TicketHandler) ServeAgentCustomerServiceMessageAttachmentFile(c *gin.Context) {
	restrictedmediaresponse.SetConversationAttachmentPrivateResponseHeaders(c)
	ticketID, ticketErr := strconv.ParseUint(c.Param("id"), 10, 64)
	messageID, messageErr := strconv.ParseUint(c.Param("message_id"), 10, 64)
	attachmentIndex, indexErr := strconv.Atoi(c.Param("attachment_index"))
	if ticketErr != nil || messageErr != nil || indexErr != nil || ticketID == 0 || messageID == 0 || attachmentIndex < 0 {
		c.Status(http.StatusNotFound)
		return
	}
	agentUserID, canViewAll := adminCustomerServiceScope(c)
	file, err := h.conversationAttachmentAccess.OpenAgentCustomerServiceMessageAttachment(c.Request.Context(),
		uint(ticketID), agentUserID, canViewAll, uint(messageID), attachmentIndex, c.Query("preset"))
	respondWithAgentCustomerServiceAuthorizedAttachmentFile(c, file, err)
}

func respondWithAgentCustomerServiceAuthorizedAttachmentFile(c *gin.Context, file *service.MediaAssetFile, err error) {
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerServiceAgentAccessDenied), errors.Is(err, service.ErrCustomerServiceMessageAttachmentUnavailable),
			errors.Is(err, gorm.ErrRecordNotFound), os.IsNotExist(err):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusServiceUnavailable)
		}
		return
	}
	restrictedmediaresponse.StreamConversationAuthorizedMediaFile(c, file)
}

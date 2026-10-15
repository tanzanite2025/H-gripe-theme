package ticket

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/api/restrictedmediaresponse"
	"commerce-platform/internal/service"
)

func (h *Handler) ServePublicCustomerServiceMessageAttachmentFile(c *gin.Context) {
	restrictedmediaresponse.SetConversationAttachmentPrivateResponseHeaders(c)
	messageID, messageErr := strconv.ParseUint(c.Param("message_id"), 10, 64)
	attachmentIndex, indexErr := strconv.Atoi(c.Param("attachment_index"))
	if messageErr != nil || indexErr != nil || messageID == 0 || attachmentIndex < 0 {
		c.Status(http.StatusNotFound)
		return
	}
	file, err := h.conversationAttachmentAccess.OpenPublicCustomerServiceMessageAttachment(c.Request.Context(),
		c.Param("conversation_id"), h.existingPublicCustomerOwner(c), uint(messageID), attachmentIndex, c.Query("preset"))
	respondWithPublicCustomerServiceAuthorizedAttachmentFile(c, file, err)
}

func respondWithPublicCustomerServiceAuthorizedAttachmentFile(c *gin.Context, file *service.MediaAssetFile, err error) {
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerServiceConversationAccessDenied),
			errors.Is(err, service.ErrCustomerServiceOwnerRequired),
			errors.Is(err, service.ErrCustomerServiceMessageAttachmentUnavailable), os.IsNotExist(err):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusServiceUnavailable)
		}
		return
	}
	restrictedmediaresponse.StreamConversationAuthorizedMediaFile(c, file)
}

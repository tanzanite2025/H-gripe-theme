package ticket

import (
	"github.com/gin-gonic/gin"

	"commerce-platform/internal/api/restrictedmediaresponse"
)

func (h *Handler) ServePublicCustomerServicePendingAttachmentUploadPreview(c *gin.Context) {
	restrictedmediaresponse.SetConversationAttachmentPrivateResponseHeaders(c)
	file, err := h.conversationAttachmentAccess.OpenPublicCustomerServicePendingUploadPreview(c.Request.Context(),
		c.Param("conversation_id"), h.existingPublicCustomerOwner(c), c.Param("upload_session_id"), c.Query("preset"))
	respondWithPublicCustomerServiceAuthorizedAttachmentFile(c, file, err)
}

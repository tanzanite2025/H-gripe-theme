package admin

import (
	"commerce-platform/internal/api/restrictedmediaresponse"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *TicketHandler) ServeAgentCustomerServiceMessageAttachmentFile(c *gin.Context) {
	restrictedmediaresponse.NewAgentCustomerServiceMessageAttachmentAuthorizedHTTPHandler(h.conversationAttachmentAccess)(c)
}

func respondWithAgentCustomerServiceAuthorizedAttachmentFile(c *gin.Context, file *service.MediaAssetFile, err error) {
	restrictedmediaresponse.RespondWithAgentCustomerServiceAuthorizedAttachmentFile(c, file, err)
}

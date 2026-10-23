package restrictedmediaresponse

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/service"
)

// AuthenticatedCustomerServiceAgentScope reads only claims installed by the
// business authentication middleware. Callers must also enforce ticket:view.
func AuthenticatedCustomerServiceAgentScope(c *gin.Context) (uint, bool) {
	if c == nil {
		return 0, false
	}
	identity, _ := c.Get("user_id")
	userID, _ := identity.(uint)
	claim, _ := c.Get("role")
	rawRole, _ := claim.(string)
	role := auth.NormalizeRole(rawRole)
	return userID, role == auth.RoleAdmin || role == auth.RoleManager
}

// NewAgentCustomerServiceMessageAttachmentAuthorizedHTTPHandler is shared by
// the business route and isolated recovery checks, after their auth middleware.
func NewAgentCustomerServiceMessageAttachmentAuthorizedHTTPHandler(access *service.CustomerServiceConversationMessageAttachmentAccessService) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetConversationAttachmentPrivateResponseHeaders(c)
		ticketID, ticketErr := strconv.ParseUint(c.Param("id"), 10, 64)
		messageID, messageErr := strconv.ParseUint(c.Param("message_id"), 10, 64)
		attachmentIndex, indexErr := strconv.Atoi(c.Param("attachment_index"))
		if ticketErr != nil || messageErr != nil || indexErr != nil || ticketID == 0 || messageID == 0 || attachmentIndex < 0 {
			c.Status(http.StatusNotFound)
			return
		}
		agentUserID, canViewAll := AuthenticatedCustomerServiceAgentScope(c)
		file, err := access.OpenAgentCustomerServiceMessageAttachment(c.Request.Context(),
			uint(ticketID), agentUserID, canViewAll, uint(messageID), attachmentIndex, c.Query("preset"))
		RespondWithAgentCustomerServiceAuthorizedAttachmentFile(c, file, err)
	}
}

func RespondWithAgentCustomerServiceAuthorizedAttachmentFile(c *gin.Context, file *service.MediaAssetFile, err error) {
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
	StreamConversationAuthorizedMediaFile(c, file)
}

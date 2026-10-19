package order

import (
	"github.com/gin-gonic/gin"

	"commerce-platform/internal/api/privateattachment"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/service"
)

func (h *Handler) ServeAfterSalesAttachmentForCurrentOrderOwner(c *gin.Context) {
	privateattachment.SetPrivateAttachmentResponseHeaders(c)
	userID, ok := c.Get("user_id")
	ownerID, valid := userID.(uint)
	if !ok || !valid || ownerID == 0 {
		apierror.RespondUnauthorized(c)
		return
	}
	if h.afterSalesService == nil {
		apierror.RespondInternalError(c, service.ErrAfterSalesAttachmentStorageUnavailable)
		return
	}
	file, err := h.afterSalesService.OpenAfterSalesAttachmentForCustomer(c.Request.Context(), c.Param("order_number"), ownerID, c.Param("attachment_locator"))
	if err != nil {
		if service.IsPrivateOrderBusinessAttachmentNotFound(err) {
			apierror.RespondNotFound(c, "Attachment")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	privateattachment.ServeAuthorizedPrivateAttachmentFile(c, &storage.StoredObject{ReadCloser: file.ReadCloser, Name: file.Filename, MimeType: file.MimeType, Size: file.Size})
}

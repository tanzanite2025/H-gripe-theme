package warranty

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/api/privateattachment"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"
)

func (h *Handler) ServeWarrantyClaimAttachmentForAuthorizedViewer(c *gin.Context) {
	privateattachment.SetPrivateAttachmentResponseHeaders(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	index, indexErr := strconv.Atoi(c.Param("attachment_index"))
	if err != nil || id == 0 || indexErr != nil || index < 0 {
		apierror.RespondBadRequest(c, "Invalid attachment selector")
		return
	}
	userID, _ := warrantyUserID(c)
	email, _ := c.Get("email")
	viewerEmail, _ := email.(string)
	// File requests use a header so the long-lived claim credential never
	// appears in a downloaded URL, browser history, or referrer.
	token := strings.TrimSpace(c.GetHeader("X-Warranty-Claim-Token"))
	if userID == 0 && token == "" {
		apierror.RespondUnauthorized(c)
		return
	}
	object, err := h.warrantySvc.OpenWarrantyClaimAttachmentForViewer(c.Request.Context(), uint(id), userID, viewerEmail, token, false, c.Param("attachment_kind"), index)
	if err != nil {
		respondPrivateWarrantyAttachmentReadError(c, err)
		return
	}
	privateattachment.ServeAuthorizedPrivateAttachmentFile(c, object)
}

func (h *Handler) ServeShipmentImageForCurrentOrderOwner(c *gin.Context) {
	privateattachment.SetPrivateAttachmentResponseHeaders(c)
	userID, ok := warrantyUserID(c)
	if !ok {
		apierror.RespondUnauthorized(c)
		return
	}
	index, err := strconv.Atoi(c.Param("attachment_index"))
	if err != nil || index < 0 {
		apierror.RespondBadRequest(c, "Invalid attachment selector")
		return
	}
	if h.shipmentSvc == nil {
		apierror.RespondInternalError(c, service.ErrShipmentRecordUnavailable)
		return
	}
	object, err := h.shipmentSvc.OpenShipmentImageForCustomer(c.Request.Context(), c.Param("order_number"), userID, index)
	if err != nil {
		respondPrivateWarrantyAttachmentReadError(c, err)
		return
	}
	privateattachment.ServeAuthorizedPrivateAttachmentFile(c, object)
}

func respondPrivateWarrantyAttachmentReadError(c *gin.Context, err error) {
	if service.IsPrivateOrderBusinessAttachmentNotFound(err) {
		apierror.RespondNotFound(c, "Attachment")
		return
	}
	respondWarrantyServiceError(c, err)
}

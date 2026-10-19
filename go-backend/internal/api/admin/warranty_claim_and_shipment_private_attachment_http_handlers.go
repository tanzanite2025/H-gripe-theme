package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/api/privateattachment"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"
)

func (h *WarrantyHandler) ServeWarrantyClaimAttachmentForAdministrator(c *gin.Context) {
	privateattachment.SetPrivateAttachmentResponseHeaders(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	index, indexErr := strconv.Atoi(c.Param("attachment_index"))
	if err != nil || id == 0 || indexErr != nil || index < 0 {
		apierror.RespondBadRequest(c, "Invalid attachment selector")
		return
	}
	object, err := h.warrantyService.OpenWarrantyClaimAttachmentForViewer(c.Request.Context(), uint(id), 0, "", "", true, c.Param("attachment_kind"), index)
	if err != nil {
		respondAdministratorPrivateOrderAttachmentReadError(c, err)
		return
	}
	privateattachment.ServeAuthorizedPrivateAttachmentFile(c, object)
}

func (h *ShipmentRecordHandler) ServeShipmentImageForAdministrator(c *gin.Context) {
	privateattachment.SetPrivateAttachmentResponseHeaders(c)
	id, ok := parseShipmentRecordID(c)
	if !ok {
		return
	}
	index, err := strconv.Atoi(c.Param("attachment_index"))
	if err != nil || index < 0 {
		apierror.RespondBadRequest(c, "Invalid attachment selector")
		return
	}
	object, err := h.shipmentRecordService.OpenShipmentImageForAdministrator(c.Request.Context(), id, index)
	if err != nil {
		respondAdministratorPrivateOrderAttachmentReadError(c, err)
		return
	}
	privateattachment.ServeAuthorizedPrivateAttachmentFile(c, object)
}

func respondAdministratorPrivateOrderAttachmentReadError(c *gin.Context, err error) {
	if service.IsPrivateOrderBusinessAttachmentNotFound(err) {
		apierror.RespondNotFound(c, "Attachment")
		return
	}
	apierror.RespondInternalError(c, err)
}

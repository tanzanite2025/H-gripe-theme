package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/api/restrictedmediaresponse"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"
)

func (h *TicketHandler) ServeAgentCustomerServicePendingAttachmentUploadPreview(c *gin.Context) {
	restrictedmediaresponse.SetConversationAttachmentPrivateResponseHeaders(c)
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}
	agentID, canViewAll := adminCustomerServiceScope(c)
	file, err := h.conversationAttachmentAccess.OpenAgentCustomerServicePendingUploadPreview(c.Request.Context(),
		ticketID, agentID, canViewAll, c.Param("upload_session_id"), c.Query("preset"))
	respondWithAgentCustomerServiceAuthorizedAttachmentFile(c, file, err)
}

func (h *TicketHandler) UploadAgentCustomerServiceConversationAttachment(c *gin.Context) {
	restrictedmediaresponse.SetConversationAttachmentPrivateResponseHeaders(c)
	ticketID, ok := parseAdminCustomerServiceConversationID(c)
	if !ok {
		return
	}
	if h.mediaService == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, mediaMaxRequestBytes)
	file, err := c.FormFile("file")
	if err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "attachment request is too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}
	agentID, canViewAll := adminCustomerServiceScope(c)
	uploads := service.NewCustomerServiceConversationAttachmentUploadService(h.ticketService, h.mediaService)
	asset, err := uploads.UploadVerifiedAgentCustomerServiceConversationAttachment(c.Request.Context(), ticketID, agentID, canViewAll, file)
	if err != nil {
		var validationError *upload.ValidationError
		switch {
		case errors.As(err, &validationError):
			c.JSON(upload.HTTPStatus(err), gin.H{"error": "invalid attachment", "code": upload.ErrorCode(err)})
		case errors.Is(err, service.ErrMediaAccountStorageQuotaExceeded):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "storage quota exceeded"})
		case errors.Is(err, service.ErrUnsupportedMediaType):
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "private conversation uploads currently support images only"})
		default:
			respondAdminCustomerServiceError(c, err)
		}
		return
	}
	attachmentURL := service.AgentCustomerServicePendingUploadPreviewURL(ticketID, asset.PendingCustomerServiceUploadSessionID)
	if asset.PendingCustomerServiceUploadSessionID == "" {
		attachmentURL = h.mediaService.CanonicalPublicMediaURL(asset.URL)
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "asset": gin.H{
		"id": asset.ID, "url": attachmentURL, "media_type": asset.MediaType,
		"library_partition": asset.LibraryPartition, "customer_service_media_purpose": asset.CustomerServiceMediaPurpose, "upload_origin": asset.UploadOrigin,
		"upload_session_id": asset.PendingCustomerServiceUploadSessionID, "original_filename": asset.OriginalFilename, "mime_type": asset.MimeType, "size": asset.Size,
	}})
}

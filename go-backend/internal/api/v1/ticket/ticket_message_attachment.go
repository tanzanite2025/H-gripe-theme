package ticket

import (
	"errors"
	"net/http"
	"strings"

	"commerce-platform/internal/pkg/ugc"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ticketMessageJSONMaxBytes              = 128 << 10
	ticketMessageAttachmentMaxCount        = 4
	publicGuestMessageAttachmentMaxCount   = 1
	publicMemberMessageAttachmentMaxCount  = 4
	ticketAttachmentValidationErrorMessage = "invalid attachment reference"
)

func (h *Handler) sanitizeTicketMessageAttachments(values []string, maxCount int) ([]string, error) {
	refs, err := ugc.NormalizeUploadImageAttachmentReferences(values, maxCount)
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return []string{}, nil
	}

	attachments := make([]string, 0, len(refs))
	for _, ref := range refs {
		value := ref.Value
		if h.mediaService != nil {
			canonicalURL, err := h.mediaService.CanonicalPublicImageUploadURL(ref.RawValue)
			if err != nil {
				return nil, err
			}
			value = canonicalURL
		}
		attachments = append(attachments, value)
	}
	return attachments, nil
}

func publicCustomerServiceMessageAttachmentLimit(c *gin.Context) int {
	if publicCustomerUserID(c) == nil {
		return publicGuestMessageAttachmentMaxCount
	}
	return publicMemberMessageAttachmentMaxCount
}

func mergeTicketAttachmentInputs(primary string, values []string) []string {
	merged := make([]string, 0, len(values)+1)
	appendCandidate := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		merged = append(merged, value)
	}

	appendCandidate(primary)
	for _, value := range values {
		appendCandidate(value)
	}
	return merged
}

func respondTicketJSONBindError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		status = http.StatusRequestEntityTooLarge
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func respondPublicCustomerServiceJSONBindError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		status = http.StatusRequestEntityTooLarge
	}
	c.JSON(status, gin.H{"success": false, "message": "[CRITICAL] " + err.Error()})
}

func respondTicketAttachmentError(c *gin.Context, err error) {
	c.JSON(ticketAttachmentHTTPStatus(err), gin.H{
		"error": ticketAttachmentValidationErrorMessage,
		"code":  ticketAttachmentErrorCode(err),
	})
}

func respondPublicCustomerServiceAttachmentError(c *gin.Context, err error) {
	c.JSON(ticketAttachmentHTTPStatus(err), gin.H{
		"success": false,
		"message": ticketAttachmentValidationErrorMessage,
		"code":    ticketAttachmentErrorCode(err),
	})
}

func ticketAttachmentHTTPStatus(err error) int {
	switch {
	case errors.Is(err, ugc.ErrAttachmentTooLong):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, ugc.ErrAttachmentInvalidType), errors.Is(err, service.ErrUnsupportedMediaType):
		return http.StatusUnsupportedMediaType
	default:
		return http.StatusBadRequest
	}
}

func ticketAttachmentErrorCode(err error) string {
	switch {
	case errors.Is(err, ugc.ErrAttachmentTooMany):
		return "too_many_attachments"
	case errors.Is(err, ugc.ErrAttachmentTooLong):
		return "attachment_reference_too_long"
	case errors.Is(err, ugc.ErrAttachmentInvalidType), errors.Is(err, service.ErrUnsupportedMediaType):
		return "invalid_attachment_type"
	case errors.Is(err, service.ErrMediaAssetNotFound):
		return "attachment_not_found"
	case errors.Is(err, service.ErrMediaAssetForbidden):
		return "attachment_unavailable"
	default:
		return "invalid_attachment"
	}
}

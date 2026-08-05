package admin

import (
	"errors"
	"net/http"
	"strings"

	"tanzanite/internal/pkg/ugc"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

const adminCustomerServiceAttachmentMaxCount = 4
const adminCustomerServiceMessageJSONMaxBytes = 128 << 10

func limitAdminCustomerServiceJSONBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, adminCustomerServiceMessageJSONMaxBytes)
}

func (h *TicketHandler) sanitizeAdminCustomerServiceAttachments(primary string, values []string) ([]string, error) {
	refs, err := ugc.NormalizeUploadImageAttachmentReferences(mergeAdminAttachmentInputs(primary, values), adminCustomerServiceAttachmentMaxCount)
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

func mergeAdminAttachmentInputs(primary string, values []string) []string {
	merged := make([]string, 0, len(values)+1)
	if strings.TrimSpace(primary) != "" {
		merged = append(merged, strings.TrimSpace(primary))
	}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			merged = append(merged, strings.TrimSpace(value))
		}
	}
	return merged
}

func respondAdminAttachmentError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, ugc.ErrAttachmentTooLong) {
		status = http.StatusRequestEntityTooLarge
	}
	if errors.Is(err, ugc.ErrAttachmentInvalidType) || errors.Is(err, service.ErrUnsupportedMediaType) {
		status = http.StatusUnsupportedMediaType
	}
	c.JSON(status, gin.H{
		"error": "invalid attachment reference",
		"code":  adminAttachmentErrorCode(err),
	})
}

func respondAdminJSONBindError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		status = http.StatusRequestEntityTooLarge
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func adminAttachmentErrorCode(err error) string {
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

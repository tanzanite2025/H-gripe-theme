package admin

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
)

type mediaAssetManualLibraryClassificationRequest struct {
	LibraryPartition            media.MediaAssetLibraryPartition `json:"library_partition"`
	CustomerServiceMediaPurpose string                           `json:"customer_service_media_purpose"`
	ExpectedRevision            *int64                           `json:"expected_revision"`
	Reason                      string                           `json:"reason"`
}

func canManuallyAssignMediaAssetToBusinessLibraryPartition(role auth.Role, partition media.MediaAssetLibraryPartition) bool {
	if !role.HasPermission(auth.PermMediaConfigure) {
		return false
	}
	switch partition {
	case media.MediaLibraryPartitionCatalog:
		return role.HasPermission(auth.PermProductEdit)
	case media.MediaLibraryPartitionContent:
		return role.HasPermission(auth.PermContentEdit)
	case media.MediaLibraryPartitionCustomerService:
		return role.HasPermission(auth.PermTicketEdit)
	case media.MediaLibraryPartitionShared, media.MediaLibraryPartitionUnclassified:
		return true
	default:
		return false
	}
}

func (h *MediaHandler) ReviewMediaAssetLibraryOwnershipForManualClassification(c *gin.Context) {
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}
	review, err := h.mediaService.ReviewMediaAssetLibraryOwnershipForManualClassification(c.Request.Context(), id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	role := auth.NormalizeRole(c.GetString("role"))
	allowedChoices := make([]media.MediaAssetLibraryOwnershipChoice, 0, len(review.AllowedChoices))
	for _, choice := range review.AllowedChoices {
		if canManuallyAssignMediaAssetToBusinessLibraryPartition(role, choice.LibraryPartition) {
			allowedChoices = append(allowedChoices, choice)
		}
	}
	review.AllowedChoices = allowedChoices
	if review.SuggestedOwnership != nil && !canManuallyAssignMediaAssetToBusinessLibraryPartition(role, review.SuggestedOwnership.LibraryPartition) {
		review.SuggestedOwnership = nil
	}
	response.Success(c, gin.H{"review": review})
}

func (h *MediaHandler) ReclassifyMediaAssetLibraryOwnershipWithTransactionalAudit(c *gin.Context) {
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}
	var request mediaAssetManualLibraryClassificationRequest
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media library classification payload"})
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media library classification payload"})
		return
	}
	if !canManuallyAssignMediaAssetToBusinessLibraryPartition(auth.NormalizeRole(c.GetString("role")), request.LibraryPartition) {
		c.JSON(http.StatusForbidden, gin.H{"error": "media library classification requires target business permission"})
		return
	}
	startedAt := adminAuditStartedAt()
	asset, err := h.mediaService.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(c.Request.Context(), id,
		service.MediaAssetLibraryClassificationChangeInput{LibraryPartition: request.LibraryPartition,
			CustomerServiceMediaPurpose: request.CustomerServiceMediaPurpose, ExpectedRevision: request.ExpectedRevision, Reason: request.Reason},
		func(before, after media.MediaAssetLibraryClassificationSnapshot) (*audit.AuditLog, error) {
			return newAdminAuditLog(c, adminAuditEvent{StartedAt: startedAt, OldValue: before, NewValue: after}), nil
		})
	if err != nil {
		respondMediaError(c, err)
		return
	}
	response.Success(c, gin.H{"asset": asset})
}

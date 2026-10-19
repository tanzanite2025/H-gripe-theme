package admin

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type missingGalleryParentRepairRequest struct {
	ImageID              string `json:"image_id"`
	DestinationGalleryID string `json:"destination_gallery_id"`
	ExpectedFingerprint  string `json:"expected_fingerprint,omitempty"`
	Reason               string `json:"reason,omitempty"`
}

func respondMediaAssociationDiagnosisOrRepairError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrMediaAssociationRepairInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供有效记录、最新预览和 3 至 1000 字审核理由", "code": "media_association_repair_invalid"})
	case errors.Is(err, repository.ErrMediaAssociationRepairConflict), errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusConflict, gin.H{"error": "关联已变化或不满足修复条件，请重新诊断；原引用继续保留", "code": "media_association_repair_conflict"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "本次核查或修复未完成，请重新诊断并核对审计记录", "code": "media_association_operation_unavailable"})
	}
}

func (h *MediaHandler) DiagnoseMediaReferenceAssociations(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}
	asset, err := h.mediaService.GetAsset(id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	if !h.authorizeCustomerServiceMediaAssetMetadataAccess(c, asset) {
		return
	}
	report, err := h.mediaService.DiagnoseMediaReferenceAssociationsWithPermissionScopedPresentation(c.Request.Context(), id, mediaUsageSummaryReferenceScopeForAuthenticatedAdministrator(c))
	if err != nil {
		respondMediaAssociationDiagnosisOrRepairError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"diagnosis": report}})
}

func (h *MediaHandler) reviewOrApplyMissingGalleryParentRepair(c *gin.Context, apply bool) {
	c.Header("Cache-Control", "no-store")
	role := auth.NormalizeRole(c.GetString("role"))
	for _, permission := range []auth.Permission{auth.PermMediaConfigure, auth.PermGalleryView, auth.PermGalleryEdit, auth.PermGalleryCreate} {
		if !role.HasPermission(permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "修复需要媒体配置及图库查看、创建、编辑权限"})
			return
		}
	}
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}
	var request missingGalleryParentRepairRequest
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		respondMediaAssociationDiagnosisOrRepairError(c, service.ErrMediaAssociationRepairInvalidInput)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		respondMediaAssociationDiagnosisOrRepairError(c, service.ErrMediaAssociationRepairInvalidInput)
		return
	}
	imageID, imageErr := strconv.ParseUint(request.ImageID, 10, strconv.IntSize)
	destinationID, destinationErr := strconv.ParseUint(request.DestinationGalleryID, 10, strconv.IntSize)
	if imageErr != nil || destinationErr != nil || imageID == 0 || destinationID == 0 || strconv.FormatUint(imageID, 10) != request.ImageID || strconv.FormatUint(destinationID, 10) != request.DestinationGalleryID || (!apply && (request.ExpectedFingerprint != "" || request.Reason != "")) {
		respondMediaAssociationDiagnosisOrRepairError(c, service.ErrMediaAssociationRepairInvalidInput)
		return
	}
	asset, err := h.mediaService.GetAsset(id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	if !h.authorizeCustomerServiceMediaAssetMetadataAccess(c, asset) {
		return
	}
	scope := mediaUsageSummaryReferenceScopeForAuthenticatedAdministrator(c)
	var result service.MissingGalleryParentRepairPreview
	if apply {
		startedAt := adminAuditStartedAt()
		result, err = h.mediaService.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(c.Request.Context(), id, uint(imageID), uint(destinationID), request.ExpectedFingerprint, request.Reason, scope, func() *audit.AuditLog {
			return newAdminAuditLog(c, adminAuditEvent{StartedAt: startedAt})
		})
	} else {
		result, err = h.mediaService.PreviewMissingGalleryParentReattachment(c.Request.Context(), id, uint(imageID), uint(destinationID), scope)
	}
	if err != nil {
		respondMediaAssociationDiagnosisOrRepairError(c, err)
		return
	}
	if apply {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"repaired": true, "review": result}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"preview": result}})
}

func (h *MediaHandler) PreviewMissingGalleryParentRepair(c *gin.Context) {
	h.reviewOrApplyMissingGalleryParentRepair(c, false)
}
func (h *MediaHandler) ApplyMissingGalleryParentRepairWithRequiredAudit(c *gin.Context) {
	h.reviewOrApplyMissingGalleryParentRepair(c, true)
}

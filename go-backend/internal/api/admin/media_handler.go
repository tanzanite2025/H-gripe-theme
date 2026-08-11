package admin

import (
	"errors"
	"net/http"
	"strconv"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	mediaService *service.MediaService
}

func NewMediaHandler(mediaService *service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

func parseMediaAssetID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid media asset id"})
		return 0, false
	}
	return uint(id), true
}

func respondMediaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrMediaAssetNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrUnsupportedMediaType), errors.Is(err, service.ErrUnsupportedMediaStatus), errors.Is(err, service.ErrUnsupportedVisibility):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrMediaStorageUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrMediaAssetURLUnavailable):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrMediaAssetForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrMediaUploadIdentityRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": "upload_identity_required"})
	case errors.Is(err, service.ErrMediaAccountStorageQuotaExceeded):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error(), "code": "account_storage_quota_exceeded"})
	case errors.Is(err, service.ErrMediaDeleteConfirmationRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "delete_confirmation_required"})
	case errors.Is(err, service.ErrMediaAssetInUse):
		var inUse *service.MediaAssetInUseError
		if errors.As(err, &inUse) {
			c.JSON(http.StatusConflict, gin.H{
				"error":      err.Error(),
				"code":       "media_asset_in_use",
				"references": inUse.References,
				"total":      len(inUse.References),
			})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "code": "media_asset_in_use"})
	case errors.Is(err, service.ErrMediaEvidenceUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrMediaEvidenceIntegrityMismatch):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "media operation failed"})
	}
}

func currentUserID(c *gin.Context) uint {
	value, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	switch typed := value.(type) {
	case uint:
		return typed
	case uint64:
		return uint(typed)
	case int:
		if typed > 0 {
			return uint(typed)
		}
	}
	return 0
}

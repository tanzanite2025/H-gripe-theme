package admin

import (
	"github.com/gin-gonic/gin"

	"commerce-platform/internal/domain/media"
)

func mediaAssetUploadOriginSelectedByAuthorizedHandler(forcedPurpose string, trustedOrigin []media.MediaAssetUploadOrigin) media.MediaAssetUploadOrigin {
	if len(trustedOrigin) > 0 {
		return trustedOrigin[0]
	}
	if forcedPurpose != "" {
		return media.MediaUploadOriginAdministratorContent
	}
	return media.MediaUploadOriginAdministratorGeneral
}

// Registration supplies a fixed origin after the existing domain permissions.
// Multipart library_partition/upload_origin/classification_method are ignored.
func (h *MediaHandler) UploadOperationalMediaAssetForAuthorizedOrigin(origin media.MediaAssetUploadOrigin) gin.HandlerFunc {
	return func(c *gin.Context) { h.uploadAsset(c, "", origin) }
}

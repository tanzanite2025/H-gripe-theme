package admin

import (
	"mime/multipart"
	"net/http"
	"strings"
	"tanzanite/internal/pkg/upload"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	mediaService *service.MediaService
}

func NewMediaHandler(mediaService *service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

func (h *MediaHandler) UploadAsset(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	mediaType, err := validateProductMediaUpload(file, c.PostForm("media_type"))
	if err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{
			"error": err.Error(),
			"code":  upload.ErrorCode(err),
		})
		return
	}

	asset, err := h.mediaService.UploadAsset(c.Request.Context(), service.MediaUploadInput{
		File:       file,
		MediaType:  mediaType,
		Alt:        strings.TrimSpace(c.PostForm("alt")),
		Caption:    strings.TrimSpace(c.PostForm("caption")),
		UploaderID: currentUserID(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload media asset"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Media asset uploaded successfully",
		"asset":   asset,
	})
}

func validateProductMediaUpload(file *multipart.FileHeader, requestedType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(requestedType)) {
	case "image":
		return "image", upload.ValidateFile(file, upload.ProductImageRule)
	case "video":
		return "video", upload.ValidateFile(file, upload.ProductVideoRule)
	default:
		if err := upload.ValidateFile(file, upload.ProductImageRule); err == nil {
			return "image", nil
		}
		if err := upload.ValidateFile(file, upload.ProductVideoRule); err == nil {
			return "video", nil
		} else {
			return "", err
		}
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

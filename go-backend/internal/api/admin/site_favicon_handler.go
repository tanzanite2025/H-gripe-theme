package admin

import (
	"errors"
	"net/http"

	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	siteFaviconUploadMaxRequestBytes = 2 << 20
	siteFaviconUploadMultipartMemory = 1 << 20
)

type SiteFaviconHandler struct {
	service *service.SiteFaviconService
}

func NewSiteFaviconHandler(faviconService *service.SiteFaviconService, _ ...*service.AdminSettingsService) *SiteFaviconHandler {
	return &SiteFaviconHandler{service: faviconService}
}

func (h *SiteFaviconHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, siteFaviconUploadMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(siteFaviconUploadMultipartMemory); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "site favicon upload is too large", "code": "request_too_large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart upload", "code": "invalid_upload"})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required", "code": upload.CodeEmptyFile})
		return
	}
	favicon, err := h.service.UploadCurrent(c.Request.Context(), file, currentUserID(c), c.DefaultQuery("locale", "en"))
	if err != nil {
		respondSiteFaviconError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Response{Code: 0, Message: "Site favicon uploaded successfully", Data: gin.H{"favicon": favicon}})
}

func (h *SiteFaviconHandler) Delete(c *gin.Context) {
	if err := h.service.DeleteCurrent(c.Request.Context(), c.DefaultQuery("locale", "en")); err != nil {
		respondSiteFaviconError(c, err)
		return
	}
	response.NoContent(c)
}

func respondSiteFaviconError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSiteFaviconUploadIdentityRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": "upload_identity_required"})
	case errors.Is(err, service.ErrSiteFaviconUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error(), "code": "site_favicon_unavailable"})
	case errors.Is(err, service.ErrObjectStorageCleanupUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error(), "code": "object_cleanup_unavailable"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": upload.ErrorCode(err)})
	}
}

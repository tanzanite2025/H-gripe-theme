package admin

import (
	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	siteLogoUploadMaxRequestBytes = 2 << 20
	siteLogoUploadMultipartMemory = 1 << 20
)

type SiteLogoHandler struct {
	siteLogoService *service.SiteLogoService
	settingsService *service.AdminSettingsService
}

func NewSiteLogoHandler(siteLogoService *service.SiteLogoService, settingsService *service.AdminSettingsService) *SiteLogoHandler {
	return &SiteLogoHandler{
		siteLogoService: siteLogoService,
		settingsService: settingsService,
	}
}

func (h *SiteLogoHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, siteLogoUploadMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(siteLogoUploadMultipartMemory); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "site logo upload is too large", "code": "request_too_large"})
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

	logo, err := h.siteLogoService.UploadCurrent(c.Request.Context(), file, currentUserID(c))
	if err != nil {
		if code := upload.ErrorCode(err); code != "invalid_upload" {
			c.JSON(upload.HTTPStatus(err), gin.H{"error": err.Error(), "code": code})
			return
		}
		respondSiteLogoError(c, err)
		return
	}

	if h.settingsService != nil {
		_, err = h.settingsService.UpdateSetting(settingdomain.UpdateSettingRequest{
			Key:         "site_logo",
			Value:       logo.URL,
			Type:        "string",
			Group:       "site",
			Locale:      c.DefaultQuery("locale", "en"),
			IsPublic:    true,
			Description: "Site logo URL",
		})
		if err != nil {
			respondSiteLogoError(c, err)
			return
		}
	}

	c.JSON(http.StatusCreated, response.Response{
		Code:    0,
		Message: "Site logo uploaded successfully",
		Data:    gin.H{"logo": logo},
	})
}

func (h *SiteLogoHandler) Delete(c *gin.Context) {
	if h.settingsService != nil {
		_, err := h.settingsService.UpdateSetting(settingdomain.UpdateSettingRequest{
			Key:         "site_logo",
			Value:       "",
			Type:        "string",
			Group:       "site",
			Locale:      c.DefaultQuery("locale", "en"),
			IsPublic:    true,
			Description: "Site logo URL",
		})
		if err != nil {
			respondSiteLogoError(c, err)
			return
		}
	}

	if err := h.siteLogoService.DeleteCurrent(c.Request.Context()); err != nil {
		respondSiteLogoError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Code:    0,
		Message: "Site logo deleted successfully",
	})
}

func respondSiteLogoError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSiteLogoUploadIdentityRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": "upload_identity_required"})
	case errors.Is(err, service.ErrSiteLogoUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error(), "code": "site_logo_unavailable"})
	case errors.Is(err, service.ErrSiteLogoPreviousDestroyFailed):
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "previous_site_logo_destroy_failed"})
	case errors.Is(err, service.ErrSiteLogoCurrentDestroyFailed):
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "code": "current_site_logo_destroy_failed"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload site logo", "code": "site_logo_upload_failed"})
	}
}

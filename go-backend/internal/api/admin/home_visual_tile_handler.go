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
	visualShowcaseUploadMaxRequestBytes = 14 << 20
	visualShowcaseUploadMultipartMemory = 8 << 20
)

type HomeVisualTileHandler struct {
	HomeVisualTileService *service.HomeVisualTileService
}

func NewHomeVisualTileHandler(homeVisualTileService *service.HomeVisualTileService) *HomeVisualTileHandler {
	return &HomeVisualTileHandler{HomeVisualTileService: homeVisualTileService}
}

type visualShowcaseItemRequest struct {
	ImageURL        string `json:"image_url"`
	ThumbnailURL    string `json:"thumbnail_url"`
	StorageKey      string `json:"storage_key"`
	Title           string `json:"title"`
	Caption         string `json:"caption"`
	AltText         string `json:"alt_text"`
	DesktopOrder    int    `json:"desktop_order"`
	MobilePairIndex int    `json:"mobile_pair_index"`
	TargetURL       string `json:"target_url"`
	TargetLabel     string `json:"target_label"`
	LayoutVariant   string `json:"layout_variant"`
	IsPublished     bool   `json:"is_published"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
}

func respondVisualShowcaseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrHomeVisualTileStorageUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrHomeVisualTilePreviousDestroyFailed):
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrHomeVisualTileKeyRequired),
		errors.Is(err, service.ErrHomeVisualTileLocaleRequired),
		errors.Is(err, service.ErrHomeVisualTileItemLimit),
		errors.Is(err, service.ErrHomeVisualTileTitleRequired),
		errors.Is(err, service.ErrHomeVisualTileAltTextRequired),
		errors.Is(err, service.ErrHomeVisualTileImageRequired),
		errors.Is(err, service.ErrHomeVisualTileImageInvalid),
		errors.Is(err, service.ErrHomeVisualTileUploadFileRequired),
		errors.Is(err, service.ErrHomeVisualTileAspectRatioInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "visual showcase operation failed"})
	}
}

func (h *HomeVisualTileHandler) GetItems(c *gin.Context) {
	tileSetKey := c.Param("showcase_key")
	locale := c.DefaultQuery("locale", "en")
	items, err := h.HomeVisualTileService.GetAdminItems(tileSetKey, locale)
	if err != nil {
		respondVisualShowcaseError(c, err)
		return
	}
	response.Success(c, gin.H{
		"showcase_key": tileSetKey,
		"locale":       locale,
		"items":        items,
	})
}

func (h *HomeVisualTileHandler) UploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, visualShowcaseUploadMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(visualShowcaseUploadMultipartMemory); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "visual showcase upload is too large", "code": "request_too_large"})
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

	tileSetKey := c.Param("showcase_key")
	locale := c.DefaultPostForm("locale", c.DefaultQuery("locale", "en"))
	image, err := h.HomeVisualTileService.UploadAdminImage(c.Request.Context(), tileSetKey, locale, file)
	if err != nil {
		if code := upload.ErrorCode(err); code != "invalid_upload" {
			c.JSON(upload.HTTPStatus(err), gin.H{"error": err.Error(), "code": code})
			return
		}
		respondVisualShowcaseError(c, err)
		return
	}

	response.SuccessWithMessage(c, "Visual showcase image uploaded", gin.H{
		"asset": image,
	})
}

func (h *HomeVisualTileHandler) ReplaceItems(c *gin.Context) {
	tileSetKey := c.Param("showcase_key")
	var req struct {
		Locale string                      `json:"locale"`
		Items  []visualShowcaseItemRequest `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inputs := make([]service.HomeVisualTileInput, 0, len(req.Items))
	for _, item := range req.Items {
		inputs = append(inputs, service.HomeVisualTileInput{
			ImageURL:        item.ImageURL,
			ThumbnailURL:    item.ThumbnailURL,
			StorageKey:      item.StorageKey,
			Title:           item.Title,
			Caption:         item.Caption,
			AltText:         item.AltText,
			DesktopOrder:    item.DesktopOrder,
			MobilePairIndex: item.MobilePairIndex,
			TargetURL:       item.TargetURL,
			TargetLabel:     item.TargetLabel,
			LayoutVariant:   item.LayoutVariant,
			IsPublished:     item.IsPublished,
			Width:           item.Width,
			Height:          item.Height,
		})
	}

	items, err := h.HomeVisualTileService.ReplaceAdminItems(c.Request.Context(), tileSetKey, req.Locale, inputs)
	if err != nil {
		respondVisualShowcaseError(c, err)
		return
	}
	response.SuccessWithMessage(c, "Visual showcase saved", gin.H{
		"showcase_key": tileSetKey,
		"locale":       req.Locale,
		"items":        items,
	})
}

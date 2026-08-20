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

type VisualShowcaseHandler struct {
	visualShowcaseService *service.VisualShowcaseService
}

func NewVisualShowcaseHandler(visualShowcaseService *service.VisualShowcaseService) *VisualShowcaseHandler {
	return &VisualShowcaseHandler{visualShowcaseService: visualShowcaseService}
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
	case errors.Is(err, service.ErrVisualShowcaseStorageUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrVisualShowcasePreviousDestroyFailed):
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrVisualShowcaseKeyRequired),
		errors.Is(err, service.ErrVisualShowcaseLocaleRequired),
		errors.Is(err, service.ErrVisualShowcaseItemLimit),
		errors.Is(err, service.ErrVisualShowcaseTitleRequired),
		errors.Is(err, service.ErrVisualShowcaseAltTextRequired),
		errors.Is(err, service.ErrVisualShowcaseImageRequired),
		errors.Is(err, service.ErrVisualShowcaseImageInvalid),
		errors.Is(err, service.ErrVisualShowcaseUploadFileRequired),
		errors.Is(err, service.ErrVisualShowcaseAspectRatioInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "visual showcase operation failed"})
	}
}

func (h *VisualShowcaseHandler) GetItems(c *gin.Context) {
	showcaseKey := c.Param("showcase_key")
	locale := c.DefaultQuery("locale", "en")
	items, err := h.visualShowcaseService.GetAdminItems(showcaseKey, locale)
	if err != nil {
		respondVisualShowcaseError(c, err)
		return
	}
	response.Success(c, gin.H{
		"showcase_key": showcaseKey,
		"locale":       locale,
		"items":        items,
	})
}

func (h *VisualShowcaseHandler) UploadImage(c *gin.Context) {
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

	showcaseKey := c.Param("showcase_key")
	locale := c.DefaultPostForm("locale", c.DefaultQuery("locale", "en"))
	image, err := h.visualShowcaseService.UploadAdminImage(c.Request.Context(), showcaseKey, locale, file)
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

func (h *VisualShowcaseHandler) ReplaceItems(c *gin.Context) {
	showcaseKey := c.Param("showcase_key")
	var req struct {
		Locale string                      `json:"locale"`
		Items  []visualShowcaseItemRequest `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inputs := make([]service.VisualShowcaseItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		inputs = append(inputs, service.VisualShowcaseItemInput{
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

	items, err := h.visualShowcaseService.ReplaceAdminItems(c.Request.Context(), showcaseKey, req.Locale, inputs)
	if err != nil {
		respondVisualShowcaseError(c, err)
		return
	}
	response.SuccessWithMessage(c, "Visual showcase saved", gin.H{
		"showcase_key": showcaseKey,
		"locale":       req.Locale,
		"items":        items,
	})
}

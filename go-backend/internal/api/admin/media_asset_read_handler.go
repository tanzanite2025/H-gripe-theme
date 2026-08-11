package admin

import (
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *MediaHandler) ListAssets(c *gin.Context) {
	params := pagination.ParsePagination(c)
	assets, total, err := h.mediaService.ListAssets(service.MediaAssetListInput{
		Page:       params.Page,
		PageSize:   params.PageSize,
		Search:     c.Query("search"),
		MediaType:  c.Query("media_type"),
		Status:     c.Query("status"),
		Visibility: c.Query("visibility"),
	})
	if err != nil {
		respondMediaError(c, err)
		return
	}

	response.Paged(c, gin.H{"assets": assets}, params.Page, params.PageSize, total)
}

func (h *MediaHandler) GetAsset(c *gin.Context) {
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}

	asset, err := h.mediaService.GetAsset(id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	response.Success(c, gin.H{"asset": asset})
}

func (h *MediaHandler) GetAssetReferences(c *gin.Context) {
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}

	report, err := h.mediaService.GetAssetReferences(id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	response.Success(c, gin.H{"references": report.References, "total": report.Total})
}

func (h *MediaHandler) ExportCopyrightEvidence(c *gin.Context) {
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}

	archive, filename, err := h.mediaService.ExportCopyrightEvidence(c.Request.Context(), id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/zip", archive)
}

func (h *MediaHandler) ServeAssetFile(c *gin.Context) {
	id, ok := parseMediaAssetID(c)
	if !ok {
		return
	}

	file, err := h.mediaService.OpenAssetFile(c.Request.Context(), id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	if strings.TrimSpace(file.RedirectURL) != "" {
		c.Redirect(http.StatusTemporaryRedirect, file.RedirectURL)
		return
	}
	if file.ReadCloser == nil {
		respondMediaError(c, service.ErrMediaStorageUnavailable)
		return
	}
	defer func() { _ = file.ReadCloser.Close() }()

	mimeType := strings.TrimSpace(file.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	headers := map[string]string{
		"Cache-Control":          "private, no-store",
		"Content-Disposition":    fmt.Sprintf("inline; filename=%q", file.Filename),
		"X-Content-Type-Options": "nosniff",
	}
	c.DataFromReader(http.StatusOK, file.Size, mimeType, file.ReadCloser, headers)
}

package admin

import (
	"fmt"
	"net/http"
	"strings"

	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *MediaHandler) ListAssets(c *gin.Context) {
	params := pagination.ParsePagination(c)
	assets, total, err := h.mediaService.ListAssets(service.MediaAssetListInput{
		LibraryPartition: c.Query("library_partition"), UploadOrigin: c.Query("upload_origin"), CustomerServiceMediaPurpose: c.Query("customer_service_media_purpose"),
		ExcludeConversationAttachments: !canViewAllCustomerServiceAttachmentMetadataInMediaLibrary(c), ReusableAssetsOnly: c.Query("selection_context") == "reusable",
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

	for index := range assets {
		redactCustomerServiceConversationOwnedMediaAssetLibraryDeliveryFields(&assets[index])
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
	if !h.authorizeCustomerServiceMediaAssetMetadataAccess(c, asset) {
		return
	}
	redactCustomerServiceConversationOwnedMediaAssetLibraryDeliveryFields(asset)
	response.Success(c, gin.H{"asset": asset})
}

func (h *MediaHandler) GetAssetReferences(c *gin.Context) {
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

	report, err := h.mediaService.GetAssetReferences(id)
	if err != nil {
		respondMediaError(c, err)
		return
	}
	scoped := permissionScopedMediaReferenceReportForAuthenticatedAdministrator(c, report.References)
	response.Success(c, gin.H{"references": scoped.References, "total": scoped.Total})
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

package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func mediaUsageSummaryReferenceScopeForAuthenticatedAdministrator(c *gin.Context) service.MediaUsageSummaryAccessScope {
	role := auth.NormalizeRole(c.GetString("role"))
	scope := service.MediaUsageSummaryAccessScope{AllowedReferenceTypes: make(map[string]bool)}
	for permission, resourceTypes := range map[auth.Permission][]string{
		auth.PermProductView:       {"product_media", "product_category", "product_brand_logo", "product_rich_text", "product_option_swatch", "product_template_option_swatch"},
		auth.PermContentView:       {"post", "home_visual_tile"},
		auth.PermFAQView:           {"faq"},
		auth.PermGalleryView:       {"gallery_cover", "gallery_image", "showcase"},
		auth.PermReviewView:        {"review"},
		auth.PermSettingsView:      {"website_profile_setting", "site_identity_setting", "site_logo", "site_favicon"},
		auth.PermWorkbenchFeedView: {"workbench_feed_media"},
		auth.PermMediaView:         {"media_asset_registration", "media_derivative_registration"},
	} {
		if role.HasPermission(permission) {
			for _, resourceType := range resourceTypes {
				scope.AllowedReferenceTypes[resourceType] = true
			}
		}
	}
	// Ticket permission alone can be restricted to assigned conversations. A
	// global scanner must not expose their titles, IDs, counts, or attachment data.
	if canViewAllCustomerServiceAttachmentMetadataInMediaLibrary(c) {
		for _, resourceType := range []string{"ticket_message", "ticket_message_attachment_binding", "ticket_auto_reply", "media_upload_session"} {
			scope.AllowedReferenceTypes[resourceType] = true
		}
	}
	// Private customer/order sources stay restricted until their individual
	// record-level authorization can be evaluated here.
	return scope
}

// Keep the older inspection/deletion response on the same presentation boundary
// as the summary. One restricted placeholder preserves the in-use indication
// without disclosing the number, identities or contents of inaccessible rows.
func permissionScopedMediaReferenceReportForAuthenticatedAdministrator(c *gin.Context, references []media.AssetReference) media.AssetReferenceReport {
	scope := mediaUsageSummaryReferenceScopeForAuthenticatedAdministrator(c)
	visible := make([]media.AssetReference, 0)
	restricted := false
	for _, reference := range references {
		if !scope.AllowedReferenceTypes[reference.ResourceType] {
			restricted = true
			continue
		}
		if reference.Category == media.ReferenceCategoryRegistration {
			reference.ResourceID, reference.ParentResourceID = 0, 0
			reference.Label = "其他媒体登记占用同一存储对象，需核查登记"
		}
		visible = append(visible, reference)
	}
	if restricted {
		visible = append(visible, media.AssetReference{Category: "restricted", ResourceType: "restricted_reference", UsageState: "restricted",
			Label: "含受限业务引用，文件仍受保护"})
	}
	return media.AssetReferenceReport{References: visible, Total: len(visible)}
}

func (h *MediaHandler) RebuildReadOnlyMediaUsageSummary(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	report, err := h.mediaService.RebuildReadOnlyMediaUsageSummary(c.Request.Context(), service.MediaAssetListInput{
		LibraryPartition: c.Query("library_partition"), UploadOrigin: c.Query("upload_origin"), CustomerServiceMediaPurpose: c.Query("customer_service_media_purpose"),
		ExcludeConversationAttachments: !canViewAllCustomerServiceAttachmentMetadataInMediaLibrary(c), ReusableAssetsOnly: c.Query("selection_context") == "reusable",
		Search: c.Query("search"), MediaType: c.Query("media_type"), Status: c.Query("status"), Visibility: c.Query("visibility"),
	}, mediaUsageSummaryReferenceScopeForAuthenticatedAdministrator(c))
	if errors.Is(err, repository.ErrMediaUsageSnapshotLimitExceeded) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "使用汇总超过扫描上限，请缩小筛选范围", "code": "media_usage_summary_limit_exceeded"})
		return
	}
	if errors.Is(err, service.ErrMediaUsageSummaryAlreadyRunning) {
		c.JSON(http.StatusConflict, gin.H{"error": "使用汇总正在生成，请稍后重试", "code": "media_usage_summary_running"})
		return
	}
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedMediaLibraryClassification) || errors.Is(err, service.ErrUnsupportedMediaType) ||
			errors.Is(err, service.ErrUnsupportedMediaStatus) || errors.Is(err, service.ErrUnsupportedVisibility) || errors.Is(err, service.ErrMediaAssetForbidden) {
			respondMediaError(c, err)
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "使用汇总未完成，未生成结论，请稍后重试", "code": "media_usage_summary_unavailable"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"report": report}})
}

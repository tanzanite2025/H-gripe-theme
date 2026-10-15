package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
)

func canViewAllCustomerServiceAttachmentMetadataInMediaLibrary(c *gin.Context) bool {
	actorID, canViewAll := adminCustomerServiceScope(c)
	return actorID > 0 && canViewAll && adminCustomerServiceRole(c).HasPermission(auth.PermTicketView)
}

func (h *MediaHandler) authorizeCustomerServiceMediaAssetMetadataAccess(c *gin.Context, asset *media.MediaAsset) bool {
	if service.IsCustomerServiceConversationOwnedMediaAsset(asset) && !canViewAllCustomerServiceAttachmentMetadataInMediaLibrary(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "conversation attachment metadata requires full customer-service access"})
		return false
	}
	return true
}

func redactCustomerServiceConversationOwnedMediaAssetLibraryDeliveryFields(asset *media.MediaAsset) {
	if !service.IsCustomerServiceConversationOwnedMediaAsset(asset) {
		return
	}
	asset.RequiresConversationAccess = true
	asset.URL, asset.AccessURL, asset.StorageKey, asset.ContentSHA256 = "", "", "", ""
	asset.Derivatives = nil
}

func (h *MediaHandler) CountMediaAssetsByLibraryPartition(c *gin.Context) {
	counts, err := h.mediaService.CountMediaAssetsByLibraryPartition(service.MediaAssetListInput{
		Search: c.Query("search"), MediaType: c.Query("media_type"), Status: c.Query("status"), Visibility: c.Query("visibility"),
		UploadOrigin: c.Query("upload_origin"), ExcludeConversationAttachments: !canViewAllCustomerServiceAttachmentMetadataInMediaLibrary(c),
		ReusableAssetsOnly: c.Query("selection_context") == "reusable",
	})
	if err != nil {
		respondMediaError(c, err)
		return
	}
	response.Success(c, gin.H{"partition_counts": counts, "can_view_conversation_attachments": canViewAllCustomerServiceAttachmentMetadataInMediaLibrary(c)})
}

func (h *MediaHandler) rejectMediaLibraryMutationOfConversationOwnedAttachment(c *gin.Context, id uint) bool {
	asset, err := h.mediaService.GetAsset(id)
	if err != nil {
		respondMediaError(c, err)
		return true
	}
	if service.IsCustomerServiceConversationOwnedMediaAsset(asset) {
		c.JSON(http.StatusForbidden, gin.H{"error": "conversation attachments are managed through their conversation lifecycle"})
		return true
	}
	return false
}

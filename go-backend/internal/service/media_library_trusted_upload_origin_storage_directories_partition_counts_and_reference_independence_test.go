package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"
)

func TestMediaLibraryTrustedUploadOriginsPersistOwnershipPurposeAndSeparateStableOriginalDirectories(t *testing.T) {
	for _, expected := range []struct {
		origin    media.MediaAssetUploadOrigin
		purpose   string
		partition media.MediaAssetLibraryPartition
		prefix    string
	}{
		{media.MediaUploadOriginAdministratorCatalog, "product_image", media.MediaLibraryPartitionCatalog, "catalog/product-images/"},
		{media.MediaUploadOriginAdministratorCatalog, "product_description_image", media.MediaLibraryPartitionCatalog, "catalog/product-detail-media/"},
		{media.MediaUploadOriginAdministratorCatalog, "product_variant_swatch", media.MediaLibraryPartitionCatalog, "catalog/product-variant-swatches/"},
		{media.MediaUploadOriginAdministratorContent, "media_library_image", media.MediaLibraryPartitionContent, "content/page-media/"},
		{media.MediaUploadOriginAdministratorShared, "media_library_image", media.MediaLibraryPartitionShared, "shared/reusable-brand-and-instruction-media/"},
		{media.MediaUploadOriginAdministratorReusableCustomerService, "media_library_image", media.MediaLibraryPartitionCustomerService, "customer-service/reusable-reply-media/"},
		{media.MediaUploadOriginCustomerServiceConversation, "customer_service_attachment", media.MediaLibraryPartitionCustomerService, "customer-service/public-conversation-attachments/"},
	} {
		t.Run(string(expected.origin)+expected.purpose, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			mediaService := newMediaDeleteTestService(t, db, t.TempDir())
			asset, err := mediaService.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "material.png", "image/png", mediaUploadTestPNG(t)),
				MediaType: "image", Width: 1, Height: 1, UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42},
				UploadOrigin: expected.origin, UploadPurpose: expected.purpose})
			require.NoError(t, err)
			stored, err := mediaService.repo.FindAssetByID(asset.ID)
			require.NoError(t, err)
			require.Equal(t, expected.partition, stored.LibraryPartition)
			require.Equal(t, expected.origin, stored.UploadOrigin)
			require.Equal(t, expected.purpose, stored.UploadPurpose)
			require.Equal(t, media.MediaClassificationRecordedUploadContext, stored.ClassificationMethod)
			require.True(t, strings.HasPrefix(stored.StorageKey, expected.prefix), stored.StorageKey)
			publicAccess, accessErr := NewPublicUploadAccessService(mediaService, nil).CanServePublicUpload(t.Context(), stored.StorageKey)
			require.NoError(t, accessErr)
			require.True(t, publicAccess, "new public upload directories must retain existing public delivery")
			if expected.origin == media.MediaUploadOriginCustomerServiceConversation {
				require.Equal(t, media.CustomerServiceMediaPurposeConversationAttachment, stored.CustomerServiceMediaPurpose)
				_, err = mediaService.OpenAssetFile(t.Context(), asset.ID)
				require.ErrorIs(t, err, ErrMediaAssetForbidden)
				var claim CopyrightEvidenceClaim
				require.NoError(t, json.Unmarshal([]byte(stored.CopyrightClaimJSON), &claim))
				require.Empty(t, claim.RightsHolder)
			}
		})
	}
}

func TestMediaLibraryPartitionCountsAndReusableSelectionPreserveUnknownHistoryWhileExcludingConversationAndCustomerSubmissions(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{})
	mediaService := newMediaDeleteTestService(t, db, t.TempDir())
	assets := []media.MediaAsset{
		{URL: "https://media.example.test/uploads/product.png", Filename: "product.png", LibraryPartition: media.MediaLibraryPartitionCatalog, Status: "active", Visibility: "public"},
		{URL: "https://media.example.test/uploads/page.png", Filename: "page.png", LibraryPartition: media.MediaLibraryPartitionContent, Status: "active", Visibility: "public"},
		{URL: "https://media.example.test/uploads/shared.png", Filename: "shared.png", LibraryPartition: media.MediaLibraryPartitionShared, Status: "active", Visibility: "public"},
		{URL: "https://media.example.test/uploads/legacy.png", Filename: "legacy.png", Status: "active", Visibility: "public"},
		{URL: "https://media.example.test/uploads/conversation.png", Filename: "conversation.png", LibraryPartition: media.MediaLibraryPartitionCustomerService, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeConversationAttachment, Status: "active", Visibility: "public"},
		{URL: "https://media.example.test/uploads/reusable.png", Filename: "reusable.png", LibraryPartition: media.MediaLibraryPartitionCustomerService, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeReusableReply, Status: "active", Visibility: "public"},
		{URL: "https://media.example.test/uploads/customer.png", Filename: "customer.png", LibraryPartition: media.MediaLibraryPartitionCustomerSubmissions, Status: "active", Visibility: "public"},
		{URL: "https://private.example.test/uploads/customer-service/attachments/originals/legacy.png", Filename: "old-private.png", StorageKey: "customer-service/attachments/originals/legacy.png", Status: "active", Visibility: "private"},
	}
	require.NoError(t, db.Create(&assets).Error)
	_, total, err := mediaService.ListAssets(MediaAssetListInput{LibraryPartition: "operational", PageSize: 1})
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	counts, err := mediaService.CountMediaAssetsByLibraryPartition(MediaAssetListInput{ExcludeConversationAttachments: true, LibraryPartition: "catalog", PageSize: 1})
	require.NoError(t, err)
	require.EqualValues(t, 6, counts["all"])
	require.EqualValues(t, 3, counts["operational"])
	require.EqualValues(t, 1, counts["customer_service"])
	require.EqualValues(t, 1, counts["unclassified"])
	selected, total, err := mediaService.ListAssets(MediaAssetListInput{ReusableAssetsOnly: true, PageSize: 100})
	require.NoError(t, err)
	require.EqualValues(t, 5, total)
	for _, asset := range selected {
		require.False(t, IsCustomerServiceConversationOwnedMediaAsset(&asset))
		require.NotEqual(t, media.MediaLibraryPartitionCustomerSubmissions, asset.LibraryPartition)
	}
	_, _, err = mediaService.ListAssets(MediaAssetListInput{LibraryPartition: "invalid"})
	require.ErrorIs(t, err, ErrUnsupportedMediaLibraryClassification)
	_, _, err = mediaService.ListAssets(MediaAssetListInput{ExcludeConversationAttachments: true, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeConversationAttachment})
	require.ErrorIs(t, err, ErrMediaAssetForbidden)
}

func TestMediaLibraryProductImageChatReuseDoesNotChangeOwnershipOrStorageAndReferencesExplainDeletedParents(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &product.Product{}, &product.ProductMedia{}, &ticket.TicketMessage{})
	mediaService := newMediaDeleteTestService(t, db, t.TempDir())
	asset, err := mediaService.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "product.png", "image/png", mediaUploadTestPNG(t)),
		UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42}, UploadOrigin: media.MediaUploadOriginAdministratorCatalog})
	require.NoError(t, err)
	originalKey := asset.StorageKey
	item := product.Product{SKU: "deleted-media-parent", Slug: "deleted-media-parent", Name: "Retained parent", PriceMinor: 100}
	require.NoError(t, db.Create(&item).Error)
	require.NoError(t, db.Create(&product.ProductMedia{ProductID: item.ID, MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image"}).Error)
	require.NoError(t, db.Model(&item).Update("deleted_at", time.Now()).Error)
	attachments, err := json.Marshal([]string{asset.URL})
	require.NoError(t, err)
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: 10, MessageType: "image", Content: "Product image", Attachments: string(attachments)}).Error)
	report, err := mediaService.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	for _, reference := range report.References {
		if reference.ResourceType == "product_media" {
			require.Equal(t, "retained_deleted_parent", reference.UsageState)
			require.Contains(t, reference.Label, "已删除商品保留")
		}
	}
	stored, err := mediaService.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, media.MediaLibraryPartitionCatalog, stored.LibraryPartition)
	require.Equal(t, originalKey, stored.StorageKey)
	require.ErrorIs(t, mediaService.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	require.NoError(t, db.Unscoped().Delete(&item).Error)
	report, err = mediaService.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	for _, reference := range report.References {
		if reference.ResourceType == "product_media" {
			require.Equal(t, "missing_parent", reference.UsageState)
		}
	}
}

func TestMediaLibraryConversationAttachmentsCannotBypassMetadataRestrictionsThroughImageDimensionPreflight(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	mediaService := newMediaDeleteTestService(t, db, t.TempDir())
	assets := []media.MediaAsset{
		{Filename: "catalog.png", URL: "https://media.example.test/uploads/catalog/product-images/catalog.png", LibraryPartition: media.MediaLibraryPartitionCatalog, Width: 1, Height: 1},
		{Filename: "public-conversation.png", URL: "https://media.example.test/uploads/customer-service/public-conversation-attachments/public-conversation.png", LibraryPartition: media.MediaLibraryPartitionCustomerService, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeConversationAttachment, UploadOrigin: media.MediaUploadOriginCustomerServiceConversation},
		{Filename: "legacy-private-cloud.png", URL: "https://private.example.test/customer-service/attachments/originals/legacy-private-cloud.png"},
		{Filename: "legacy-windows.png", URL: "https://private.example.test/opaque/legacy-windows.png", StorageKey: `customer-service\attachments\originals\legacy-windows.png`},
	}
	require.NoError(t, db.Create(&assets).Error)
	findings, err := mediaService.ListImageDimensionFindings(MediaImageDimensionListInput{State: MediaImageDimensionStateAll, PageSize: 100})
	require.NoError(t, err)
	require.EqualValues(t, 1, findings.Total)
	require.EqualValues(t, 1, findings.Summary.Total)
	require.Len(t, findings.Items, 1)
	require.Equal(t, assets[0].ID, findings.Items[0].Asset.ID)
	for _, asset := range assets[1:] {
		_, err = mediaService.ReconcileImageDimensions(t.Context(), asset.ID)
		require.ErrorIs(t, err, ErrMediaAssetForbidden)
	}
}

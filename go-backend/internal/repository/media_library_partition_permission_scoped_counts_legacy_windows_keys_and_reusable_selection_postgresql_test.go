package repository

import (
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
)

func TestPostgreSQLMediaLibraryPartitionCountsAndReusableSelectionExcludeLegacyConversationNamespaces(t *testing.T) {
	repo := createIsolatedPostgreSQLMediaUploadActorQuotaRepository(t)
	assets := []media.MediaAsset{
		{Filename: "catalog.png", URL: "https://media.example/uploads/catalog/product-images/catalog.png", LibraryPartition: media.MediaLibraryPartitionCatalog},
		{Filename: "historical.png", URL: "https://media.example/uploads/2020/historical.png"},
		{Filename: "public-chat.png", URL: "https://media.example/uploads/customer-service/public-conversation-attachments/public-chat.png", StorageKey: `customer-service\public-conversation-attachments\public-chat.png`},
		{Filename: "private-cloud-chat.png", URL: "https://private.example/customer-service/attachments/originals/private-cloud-chat.png"},
		{Filename: "private-windows-chat.png", URL: "https://private.example/opaque/private-windows-chat.png", StorageKey: `customer-service\attachments\originals\private-windows-chat.png`},
		{Filename: "archived.png", URL: "https://media.example/uploads/shared/archived.png", LibraryPartition: media.MediaLibraryPartitionShared, Status: "archived"},
		{Filename: "customer-submission.png", URL: "https://media.example/uploads/customer-submission.png", LibraryPartition: media.MediaLibraryPartitionCustomerSubmissions},
	}
	for index := range assets {
		// Historical inventory fixtures intentionally have no physical objects.
		require.NoError(t, repo.db.Create(&assets[index]).Error)
	}
	filter := MediaAssetFilter{ExcludeConversationAttachments: true, Page: 1, PageSize: 1}
	visible, total, err := repo.ListAssets(filter)
	require.NoError(t, err)
	require.Len(t, visible, 1)
	require.EqualValues(t, 4, total)
	counts, err := repo.CountMediaAssetsByLibraryPartitionWithinCurrentPermissionScope(filter)
	require.NoError(t, err)
	require.EqualValues(t, total, counts["all"])
	require.EqualValues(t, 2, counts["operational"])
	require.EqualValues(t, 1, counts["unclassified"])

	filter.ReusableAssetsOnly, filter.PageSize = true, 100
	reusable, total, err := repo.ListAssets(filter)
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.ElementsMatch(t, []uint{assets[0].ID, assets[1].ID}, []uint{reusable[0].ID, reusable[1].ID})
	counts, err = repo.CountMediaAssetsByLibraryPartitionWithinCurrentPermissionScope(filter)
	require.NoError(t, err)
	require.EqualValues(t, total, counts["all"])
	require.EqualValues(t, 1, counts["operational"])
	require.Zero(t, counts["customer_submissions"])
}

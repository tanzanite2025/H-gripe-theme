package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestProductInformationTemplateAndCategoryIntroReferencesRetainFilesIncludingDisabledAndDeletedContent(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.ProductInformationTemplate{}, &product.ProductCategory{}, &product.ProductCategoryTranslation{})
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/template-and-category-only.png")
	templates := NewProductInformationTemplateService(repository.NewProductInformationTemplateRepository(db))
	markup := `<p>Exact &amp; retained</p><img src="` + asset.URL + `">`
	template, err := templates.Create(ProductInformationTemplateInput{Kind: "packaging", Name: "Disabled packaging", Slug: "disabled-packaging", Locale: "en", Content: markup, IsEnabled: false})
	require.NoError(t, err)
	require.Contains(t, template.Content, asset.URL)
	category := product.ProductCategory{Name: "Category", Slug: "category-retention", Description: `<p>/uploads/plain.png</p>`, IsEnabled: false}
	require.NoError(t, db.Create(&category).Error)
	categories := NewProductCategoryService(repository.NewProductCategoryRepository(db))
	_, err = categories.UpdateSEO(category.ID, "en", nil, nil, &markup)
	require.NoError(t, err)
	_, err = categories.UpdateSEO(category.ID, "fr", nil, nil, &markup)
	require.NoError(t, err)
	// Deletion outside the normal hard-delete route simulates retained legacy
	// rows; their bytes still must not be reclaimed.
	require.NoError(t, db.Model(template).Update("deleted_at", time.Now().UTC()).Error)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Len(t, report.References, 3)
	types := map[string]string{}
	for _, reference := range report.References {
		types[reference.ResourceType] = reference.Field
	}
	require.Equal(t, "content", types["product_information_template_rich_text"])
	require.Equal(t, "seo_intro", types["product_category_rich_text"])
	require.Equal(t, "seo_intro", types["product_category_translation_rich_text"])
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	configuration := &storage.Config{Type: storage.StorageTypeLocal, LocalPath: "unused-retention-audit", BaseURL: "https://media.example.test"}
	audit := NewHistoricalMediaReadOnlyInventoryService(db, configuration, configuration, "https://shop.example.test")
	inventory, err := audit.AuditHistoricalRegistrationsAndBusinessReferences(t.Context(), defaultHistoricalMediaReadOnlyInventoryRegressionOptions())
	require.NoError(t, err)
	covered := map[string]bool{}
	for _, source := range inventory.Sources {
		if source.Table == "product_information_templates" && source.Field == "content" || (source.Table == "product_categories" || source.Table == "product_category_translations") && source.Field == "seo_intro" {
			require.True(t, source.Complete)
			require.False(t, source.Unavailable)
			require.Equal(t, 1, source.RowsScanned)
			covered[source.Table] = true
		}
	}
	require.Len(t, covered, 3)
	// Text, alt and external suffix matches must not retain an unrelated file.
	plain := `<p>` + asset.URL + `</p><img src="https://foreign.example.test/uploads/` + asset.StorageKey + `" alt="` + asset.URL + `">`
	require.NoError(t, db.Unscoped().Model(template).Update("content", plain).Error)
	require.NoError(t, db.Model(&category).Updates(map[string]any{"seo_intro": plain, "description": markup}).Error)
	require.NoError(t, db.Model(&product.ProductCategoryTranslation{}).Where("product_category_id=?", category.ID).Updates(map[string]any{"seo_intro": plain, "description": markup}).Error)
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Empty(t, report.References)
}

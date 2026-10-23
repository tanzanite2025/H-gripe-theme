package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
)

func TestProductRichTextMediaSourcesAndFileLinksRejectIneligibleOrUnregisteredFilesAtomically(t *testing.T) {
	for _, manager := range []bool{false, true} {
		for _, field := range []string{"description", "short_description"} {
			for _, placement := range []string{"image", "video", "poster", "file link"} {
				t.Run(fmt.Sprintf("manager=%t/%s/%s", manager, field, placement), func(t *testing.T) {
					db, products, _ := createProductMediaTransactionalBindingTestServices(t, manager)
					asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/private-reference.png")
					require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
					markup := `<img src="` + asset.URL + `">`
					switch placement {
					case "video":
						markup = `<video src="` + asset.URL + `"></video>`
					case "poster":
						markup = `<video src="https://external.example.test/video.webm" poster="` + asset.URL + `"></video>`
					case "file link":
						markup = `<a href="` + asset.URL + `">Download</a>`
					}
					input := createProductInputForTransactionalMediaBindingTest()
					if field == "description" {
						input.Description = markup
					} else {
						input.ShortDesc = markup
					}
					_, err := products.CreateAdminProduct(input)
					require.ErrorIs(t, err, ErrProductMediaInvalid)
					for _, model := range []any{&product.Product{}, &product.ProductVariant{}, &outbox.Event{}} {
						var count int64
						require.NoError(t, db.Model(model).Count(&count).Error)
						require.Zero(t, count)
					}
				})
			}
		}
	}
	for _, reference := range []string{"/uploads/missing.png", "/uploads/customer-service/attachments/originals/missing.png"} {
		t.Run(reference, func(t *testing.T) {
			_, products, _ := createProductMediaTransactionalBindingTestServices(t, false)
			input := createProductInputForTransactionalMediaBindingTest()
			input.Description = `<img src="` + reference + `">`
			_, err := products.CreateAdminProduct(input)
			require.ErrorIs(t, err, ErrProductMediaInvalid)
		})
	}
}

func TestProductRichTextCanonicalReferencesProtectOriginalsAndIgnoreTextAndForeignSuffixMatches(t *testing.T) {
	db, products, assets := createProductMediaTransactionalBindingTestServices(t, true)
	image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/images/photo.png")
	video := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/videos/demo.webm")
	require.NoError(t, db.Model(video).Update("media_type", "video").Error)
	derivative := media.MediaAssetDerivative{MediaAssetID: image.ID, Preset: "thumbnail", URL: "https://media.example.test/uploads/derivatives/photo.webp", StorageKey: "derivatives/photo.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	input := createProductInputForTransactionalMediaBindingTest()
	input.Description = `<p>Photo</p><img src="/uploads/derivatives/photo.webp?v=1&amp;test=2"><video src="/uploads/` + video.StorageKey + `" poster="/uploads/` + image.StorageKey + `"></video>`
	input.ShortDesc = `<a href="` + derivative.URL + `#page=2">Download</a>`
	created, err := products.CreateAdminProduct(input)
	require.NoError(t, err)
	require.Contains(t, created.Description, image.URL)
	require.Contains(t, created.Description, video.URL)
	require.NotContains(t, created.Description, derivative.URL)
	require.Contains(t, created.ShortDesc, image.URL+"#page=2")
	report, err := assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Len(t, report.References, 1)
	require.Equal(t, "product_rich_text", report.References[0].ResourceType)
	require.Equal(t, "description, short_description", report.References[0].Field)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
	translated, _, err := products.CopyAdminProductTranslation(created.ID, "fr")
	require.NoError(t, err)
	require.Contains(t, translated.Description, image.URL)

	// Historical HTML still using a derivative/encoded URL is also discoverable.
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", created.ID).Updates(map[string]any{"description": `<img src="https://shop.example.test/uploads/derivatives/photo.webp?a=1&amp;b=2">`, "short_description": ""}).Error)
	report, err = assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Len(t, report.References, 2)
	for _, id := range []uint{created.ID, translated.ID} {
		require.NoError(t, db.Model(&product.Product{}).Where("id = ?", id).Updates(map[string]any{"description": `<p>` + image.URL + `</p><img src="https://foreign.example.test/uploads/` + image.StorageKey + `"><img src="` + image.URL + `.backup">`, "short_description": ""}).Error)
	}
	report, err = assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Zero(t, report.Total)
}

func TestProductRichTextLegacyUploadCompatibilityIsFieldAndOccurrenceScopedAndCannotBeCopied(t *testing.T) {
	db, products, _ := createProductMediaTransactionalBindingTestServices(t, false)
	created, err := products.CreateAdminProduct(createProductInputForTransactionalMediaBindingTest())
	require.NoError(t, err)
	legacy := `<p>Old</p><img src="/uploads/old-unregistered.png">`
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", created.ID).Update("description", legacy).Error)
	retained := strings.Replace(legacy, "Old", "New text", 1)
	_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{Description: &retained})
	require.NoError(t, err)
	duplicate := retained + `<img src="/uploads/old-unregistered.png">`
	_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{Description: &duplicate})
	require.ErrorIs(t, err, ErrProductMediaInvalid)
	_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{ShortDesc: &retained})
	require.ErrorIs(t, err, ErrProductMediaInvalid)
	_, _, err = products.CopyAdminProductTranslation(created.ID, "fr")
	require.ErrorIs(t, err, ErrProductMediaInvalid)

	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/old-private.png")
	require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", created.ID).Update("description", `<img src="`+asset.URL+`">`).Error)
	name := "Ordinary nonmedia edit"
	updated, err := products.UpdateAdminProduct(created.ID, ProductUpdateInput{Name: &name})
	require.NoError(t, err)
	require.Contains(t, updated.Description, asset.URL)
	changed := `<p>Changed</p>` + updated.Description
	_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{Description: &changed})
	require.ErrorIs(t, err, ErrProductMediaInvalid)
}

func TestOmittedProductRichTextFieldsReloadCurrentValuesRatherThanRestoringDetachedReferences(t *testing.T) {
	db, products, assets := createProductMediaTransactionalBindingTestServices(t, false)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/detached.png")
	input := createProductInputForTransactionalMediaBindingTest()
	input.Description, input.ShortDesc = `<img src="`+asset.URL+`">`, `<a href="`+asset.URL+`">Download</a>`
	created, err := products.CreateAdminProduct(input)
	require.NoError(t, err)
	stale, err := products.productRepo.FindByID(created.ID)
	require.NoError(t, err)
	// A second edit detached both fields after the first request loaded them.
	require.NoError(t, db.Model(&product.Product{}).Where("id = ?", created.ID).Updates(map[string]any{"description": "", "short_description": ""}).Error)
	require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	stale.Name = "Concurrent ordinary edit"
	require.NoError(t, products.productRepo.UpdateWithSpecValuesVariantsOptionValuesAndMedia(stale, nil, false, nil, false, nil, false, nil, false, repository.ProductRichTextFieldMutation{}))
	current, err := products.productRepo.FindByID(created.ID)
	require.NoError(t, err)
	require.Empty(t, current.Description)
	require.Empty(t, current.ShortDesc)
	require.Equal(t, stale.Name, current.Name)
}

func createSpecificationTemplateInputWithSwatchForMediaBindingTest(reference string, assetID *uint) ProductSpecificationTemplateInput {
	return ProductSpecificationTemplateInput{Name: "Swatch template", Slug: "swatch-template", IsEnabled: true, SpecDefinitions: []ProductSpecDefinitionInput{{Name: "Finish", Slug: "finish", FieldType: "select", Role: "variant", Presentation: "color", OptionItems: []ProductSpecOptionItemInput{{ValueKey: "ruby", DefaultLabel: "Ruby", SwatchMediaAssetID: assetID, SwatchURL: reference, IsEnabledByDefault: boolPtr(true)}}}}}
}

func TestTemplateSwatchesMaterializeAndTranslateWithoutCopyingAssetsAndProtectBothOwners(t *testing.T) {
	db, products, assets := createProductMediaTransactionalBindingTestServices(t, true)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "shared/swatch.png")
	template, err := products.CreateProductSpecificationTemplate(createSpecificationTemplateInputWithSwatchForMediaBindingTest("", &asset.ID))
	require.NoError(t, err)
	require.Equal(t, asset.URL, template.SpecDefinitions[0].OptionItems[0].SwatchURL)
	input := createProductInputForTransactionalMediaBindingTest()
	input.ProductSpecificationTemplateID = &template.ID
	input.Variants[0].OptionValues = map[string]string{"finish": "ruby"}
	created, err := products.CreateAdminProduct(input)
	require.NoError(t, err)
	require.Equal(t, asset.ID, *created.VariantOptionValues[0].SwatchMediaAssetID)
	require.Equal(t, asset.URL, created.VariantOptionValues[0].SwatchURL)
	translated, _, err := products.CopyAdminProductTranslation(created.ID, "fr")
	require.NoError(t, err)
	require.Equal(t, asset.ID, *translated.VariantOptionValues[0].SwatchMediaAssetID)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Len(t, report.References, 3)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	var count int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	// Disabled options and retained deleted products still hold real references.
	require.NoError(t, db.Model(&product.ProductVariantOptionValue{}).Where("product_id = ?", created.ID).Update("is_enabled", false).Error)
	require.NoError(t, products.productRepo.Delete(created.ID))
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	var retained bool
	for _, reference := range report.References {
		retained = retained || reference.UsageState == "retained_deleted_parent"
	}
	require.True(t, retained)
}

func TestTemplateAndProductSwatchesRejectMismatchedOrIneligibleAssetsAndRevalidateMaterialization(t *testing.T) {
	for _, manager := range []bool{false, true} {
		t.Run(fmt.Sprintf("manager=%t", manager), func(t *testing.T) {
			db, products, _ := createProductMediaTransactionalBindingTestServices(t, manager)
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/swatch.png")
			other := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/other.png")
			_, err := products.CreateProductSpecificationTemplate(createSpecificationTemplateInputWithSwatchForMediaBindingTest(other.URL, &asset.ID))
			require.ErrorIs(t, err, ErrProductSpecInvalid)
			template, err := products.CreateProductSpecificationTemplate(createSpecificationTemplateInputWithSwatchForMediaBindingTest("/uploads/"+asset.StorageKey, nil))
			require.NoError(t, err)
			item := template.SpecDefinitions[0].OptionItems[0]
			require.Equal(t, asset.ID, *item.SwatchMediaAssetID)
			input := createProductInputForTransactionalMediaBindingTest()
			input.ProductSpecificationTemplateID = &template.ID
			input.Variants[0].OptionValues = map[string]string{"finish": "ruby"}
			input.VariantOptionValues = []ProductVariantOptionValueInput{{SpecDefinitionID: template.SpecDefinitions[0].ID, ValueKey: "ruby", Label: "Ruby", SwatchMediaAssetID: &asset.ID, SwatchURL: other.URL}}
			_, err = products.CreateAdminProduct(input)
			require.ErrorIs(t, err, ErrProductMediaInvalid)
			require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
			input.VariantOptionValues = nil // Revalidate the copied template snapshot.
			_, err = products.CreateAdminProduct(input)
			require.ErrorIs(t, err, ErrProductMediaInvalid)
			update := createSpecificationTemplateInputWithSwatchForMediaBindingTest(asset.URL, &asset.ID)
			update.SpecDefinitions[0].ID = template.SpecDefinitions[0].ID
			update.SpecDefinitions[0].OptionItems[0].ID = item.ID
			_, err = products.UpdateProductSpecificationTemplate(template.ID, update)
			require.ErrorIs(t, err, ErrProductSpecInvalid)
		})
	}
}

func TestLegacyTemplateAndProductSwatchesRetainOnlyTheirExistingItemAndCannotPropagate(t *testing.T) {
	db, products, _ := createProductMediaTransactionalBindingTestServices(t, true)
	template, err := products.CreateProductSpecificationTemplate(createSpecificationTemplateInputWithSwatchForMediaBindingTest("/images/static-swatch.png", nil))
	require.NoError(t, err)
	input := createProductInputForTransactionalMediaBindingTest()
	input.ProductSpecificationTemplateID = &template.ID
	input.Variants[0].OptionValues = map[string]string{"finish": "ruby"}
	created, err := products.CreateAdminProduct(input)
	require.NoError(t, err)
	legacy := "/uploads/legacy-unregistered-swatch.png"
	option := created.VariantOptionValues[0]
	require.NoError(t, db.Model(&product.ProductVariantOptionValue{}).Where("id = ?", option.ID).Update("swatch_url", legacy).Error)
	_, err = products.UpdateAdminProduct(created.ID, ProductUpdateInput{UpdateVariantOptionValues: true, VariantOptionValues: []ProductVariantOptionValueInput{{ID: &option.ID, SpecDefinitionID: option.SpecDefinitionID, ValueKey: option.ValueKey, Label: "Updated label", SwatchURL: legacy}}})
	require.NoError(t, err)
	_, _, err = products.CopyAdminProductTranslation(created.ID, "fr")
	require.ErrorIs(t, err, ErrProductMediaInvalid)
	templateItem := template.SpecDefinitions[0].OptionItems[0]
	require.NoError(t, db.Model(&product.ProductSpecOptionItem{}).Where("id = ?", templateItem.ID).Update("swatch_url", legacy).Error)
	update := createSpecificationTemplateInputWithSwatchForMediaBindingTest(legacy, nil)
	update.SpecDefinitions[0].ID = template.SpecDefinitions[0].ID
	update.SpecDefinitions[0].OptionItems[0].ID = templateItem.ID
	_, err = products.UpdateProductSpecificationTemplate(template.ID, update)
	require.NoError(t, err)
	input.Slug, input.Variants[0].SKU = "new-product-from-old-template", "NEW-PRODUCT-FROM-OLD-TEMPLATE"
	_, err = products.CreateAdminProduct(input)
	require.ErrorIs(t, err, ErrProductMediaInvalid)
	update.SpecDefinitions[0].OptionItems[0].ID = 0
	_, err = products.UpdateProductSpecificationTemplate(template.ID, update)
	require.ErrorIs(t, err, ErrProductSpecInvalid)
}

func TestCategoryImagesValidateRegisteredIdentityInsideTransactionAndDetachWithoutDeletingFiles(t *testing.T) {
	db, _, assets := createProductMediaTransactionalBindingTestServices(t, false)
	require.NoError(t, db.AutoMigrate(&product.ProductCategoryTranslation{}))
	categories := NewProductCategoryService(repository.NewProductCategoryRepository(db))
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "shared/category.png")
	input := ProductCategoryInput{Name: "Test category", Slug: "test-category", IsEnabled: true, ImageMediaAssetID: &asset.ID}
	require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
	_, err := categories.Create(input)
	require.ErrorIs(t, err, ErrProductCategoryImageInvalid)
	require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorAdministrator).Error)
	created, err := categories.Create(input)
	require.NoError(t, err)
	require.Equal(t, asset.URL, created.ImageURL)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	input.ImageMediaAssetID = nil
	updated, err := categories.Update(created.ID, input)
	require.NoError(t, err)
	require.Empty(t, updated.ImageURL)
	_, err = assets.GetAsset(asset.ID)
	require.NoError(t, err)
	var events int64
	require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ?", outbox.EventTypeObjectStorageCleanup).Count(&events).Error)
	require.Zero(t, events)
}

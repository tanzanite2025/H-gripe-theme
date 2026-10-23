package service

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

type templateAndCategoryIntroMediaBindingTestEnvironment struct {
	db         *gorm.DB
	assets     *MediaService
	templates  *ProductInformationTemplateService
	categories *ProductCategoryService
}

func configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t *testing.T, db *gorm.DB) templateAndCategoryIntroMediaBindingTestEnvironment {
	t.Helper()
	if db == nil {
		db = newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.BusinessDraftMediaUploadReceipt{}, &outbox.Event{}, &product.ProductInformationTemplate{}, &product.ProductCategory{}, &product.ProductCategoryTranslation{})
	}
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	assets := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", 20<<30)
	assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	assets.siteURL = "https://shop.example.test"
	templates := repository.NewProductInformationTemplateRepository(db)
	templates.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
	categories := repository.NewProductCategoryRepository(db)
	categories.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
	return templateAndCategoryIntroMediaBindingTestEnvironment{db: db, assets: assets, templates: NewProductInformationTemplateService(templates), categories: NewProductCategoryService(categories)}
}

type templateOrCategoryIntroMediaMutationTestTarget struct {
	table, column string
	id            uint
	invalidError  error
	write         func(string) error
}

func createTemplateOrCategoryIntroMediaMutationTestTarget(t *testing.T, env templateAndCategoryIntroMediaBindingTestEnvironment, field string) *templateOrCategoryIntroMediaMutationTestTarget {
	t.Helper()
	identity := strings.ReplaceAll(uuid.NewString(), "-", "")
	target := &templateOrCategoryIntroMediaMutationTestTarget{}
	if strings.HasPrefix(field, "template_") {
		target.table, target.column, target.invalidError = "product_information_templates", "content", ErrProductInformationTemplateInvalid
		input := ProductInformationTemplateInput{Kind: "packaging", Name: "Synthetic template", Slug: "template-" + identity, Locale: "en", IsEnabled: false}
		if field == "template_update" {
			created, err := env.templates.Create(input)
			require.NoError(t, err)
			target.id = created.ID
		}
		target.write = func(content string) error {
			input.Content = content
			var saved *product.ProductInformationTemplate
			var err error
			if field == "template_create" {
				saved, err = env.templates.Create(input)
			} else {
				saved, err = env.templates.Update(target.id, input)
			}
			if err == nil {
				target.id = saved.ID
			}
			return err
		}
		return target
	}
	category := product.ProductCategory{Name: "Synthetic disabled category", Slug: "category-" + identity, Depth: 1, IsEnabled: false}
	require.NoError(t, env.db.Select("*").Create(&category).Error)
	locale := "en"
	target.table, target.column, target.id, target.invalidError = "product_categories", "seo_intro", category.ID, ErrProductCategorySEOInvalid
	if field == "category_translation" || field == "category_translation_first_write" {
		locale = "fr"
		target.table, target.id = "product_category_translations", 0
		if field == "category_translation" {
			_, err := env.categories.UpdateSEO(category.ID, locale, nil, nil, nil)
			require.NoError(t, err)
		}
	}
	if target.table == "product_category_translations" && field != "category_translation_first_write" {
		require.NoError(t, env.db.Table(target.table).Where("product_category_id=? AND locale=?", category.ID, locale).Pluck("id", &target.id).Error)
	}
	target.write = func(content string) error {
		_, err := env.categories.UpdateSEO(category.ID, locale, nil, nil, &content)
		if err == nil && target.table == "product_category_translations" {
			err = env.db.Table(target.table).Where("product_category_id=? AND locale=?", category.ID, locale).Pluck("id", &target.id).Error
		}
		return err
	}
	return target
}

func readTemplateOrCategoryIntroPersistedMediaHTML(t *testing.T, db *gorm.DB, target *templateOrCategoryIntroMediaMutationTestTarget) string {
	t.Helper()
	var content string
	require.NoError(t, db.Table(target.table).Select(target.column).Where("id=?", target.id).Row().Scan(&content))
	return content
}

func TestTemplateAndCategoryIntroNormalSavesRejectInvalidMediaWithoutChangingBusinessRows(t *testing.T) {
	for _, field := range []string{"template_create", "template_update", "category_default", "category_translation", "category_translation_first_write"} {
		for _, scenario := range []string{"private", "customer_owned", "retired", "wrong_image_type", "wrong_video_type", "wrong_poster_type", "unregistered", "conversation_namespace", "dedicated_object", "dedicated_parent_derivative"} {
			t.Run(field+"/"+scenario, func(t *testing.T) {
				env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, nil)
				target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, field)
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/invalid.png")
				markup := `<img src="` + asset.URL + `">`
				switch scenario {
				case "private":
					require.NoError(t, env.db.Model(asset).Update("visibility", "private").Error)
				case "customer_owned":
					require.NoError(t, env.db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
				case "retired":
					require.NoError(t, env.db.Delete(asset).Error)
				case "wrong_image_type", "wrong_poster_type":
					require.NoError(t, env.db.Model(asset).Update("media_type", "video").Error)
					if scenario == "wrong_poster_type" {
						markup = `<video src="https://ordinary.invalid/video.mp4" poster="` + asset.URL + `"></video>`
					}
				case "wrong_video_type":
					markup = `<video src="` + asset.URL + `"></video>`
				case "unregistered":
					markup = `<a href="/uploads/catalog/unregistered-manual">Manual</a>`
				case "conversation_namespace":
					markup = `<img src="/uploads/customer-service/attachments/originals/secret.png">`
				case "dedicated_object", "dedicated_parent_derivative":
					require.NoError(t, env.db.Model(asset).Updates(map[string]any{"storage_key": "visual-showcase/owned.png", "url": "https://media.example.test/uploads/visual-showcase/owned.png"}).Error)
					markup = `<img src="/uploads/visual-showcase/owned.png">`
					if scenario == "dedicated_parent_derivative" {
						derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "legacy", URL: "https://media.example.test/uploads/catalog/owned-derivative.png", StorageKey: "catalog/owned-derivative.png"}
						require.NoError(t, env.db.Create(&derivative).Error)
						markup = `<img src="/uploads/catalog/owned-derivative.png">`
					}
				}
				var before int64
				require.NoError(t, env.db.Table(target.table).Count(&before).Error)
				require.ErrorIs(t, target.write(markup), target.invalidError)
				var after int64
				require.NoError(t, env.db.Table(target.table).Count(&after).Error)
				require.Equal(t, before, after)
				if target.id != 0 {
					require.Empty(t, readTemplateOrCategoryIntroPersistedMediaHTML(t, env.db, target))
				}
			})
		}
	}
}

func TestTemplateAndCategoryIntroNormalSavesCanonicalizeExactAttributesAndRetainExistingLegacyOccurrences(t *testing.T) {
	for _, field := range []string{"template_update", "category_default", "category_translation"} {
		t.Run(field, func(t *testing.T) {
			env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, nil)
			target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, field)
			image := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/image.png")
			video := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/video.mp4")
			document := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/manual")
			require.NoError(t, env.db.Model(video).Update("media_type", "video").Error)
			require.NoError(t, env.db.Model(document).Update("media_type", "document").Error)
			html := `<p>/uploads/not-a-file.png</p><img src="/uploads/catalog/image.png" alt="/uploads/alt.png"><video src="/uploads/catalog/video.mp4" poster="/uploads/catalog/image.png"></video><a href="/uploads/catalog/manual#page=2">Manual</a><a href="https://ordinary.invalid/photo.jpg">External</a>`
			require.NoError(t, target.write(html))
			stored := readTemplateOrCategoryIntroPersistedMediaHTML(t, env.db, target)
			for _, expected := range []string{image.URL, video.URL, document.URL + "#page=2", "/uploads/alt.png", "https://ordinary.invalid/photo.jpg"} {
				require.Contains(t, stored, expected)
			}
			require.ErrorIs(t, env.assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
			const legacy = `<img src="/uploads/catalog/legacy.png">`
			require.NoError(t, env.db.Table(target.table).Where("id=?", target.id).Update(target.column, legacy+legacy).Error)
			require.NoError(t, target.write(`<p>Edited</p>`+legacy))
			require.ErrorIs(t, target.write(legacy+legacy), target.invalidError)
			require.ErrorIs(t, target.write(`<a href="/uploads/catalog/legacy.png">File</a>`), target.invalidError)
			other := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, field)
			require.ErrorIs(t, other.write(legacy), target.invalidError)
			require.NoError(t, target.write(`<p>Detached</p>`))
			require.ErrorIs(t, target.write(legacy), target.invalidError)
		})
	}
}

func TestTemplateAndCategoryIntroBindingReceiptAndBusinessWriteCommitOrRollbackTogether(t *testing.T) {
	for _, field := range []string{"template_create", "template_update", "category_default", "category_translation_first_write"} {
		t.Run(field, func(t *testing.T) {
			drafts := createBusinessDraftMediaReceiptTestEnvironment(t, nil)
			require.NoError(t, drafts.db.AutoMigrate(&product.ProductInformationTemplate{}, &product.ProductCategoryTranslation{}))
			env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, drafts.db)
			env.templates.repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(drafts.service)
			env.categories.repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(drafts.service)
			asset := uploadBusinessDraftMediaReceiptTestAsset(t, drafts)
			target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, field)
			valid := `<img src="` + asset.URL + `">`
			require.ErrorIs(t, target.write(valid+`<img src="/uploads/missing.png">`), target.invalidError)
			var receipt media.BusinessDraftMediaUploadReceipt
			require.NoError(t, env.db.Where("asset_id=?", asset.ID).First(&receipt).Error)
			require.Equal(t, "ready", receipt.Status, "failed validation must roll back the draft binding")
			forcedFailure := errors.New("synthetic business persistence failure")
			callbackName := "reject_template_category_business_write_" + uuid.NewString()
			failWrite := func(tx *gorm.DB) {
				if tx.Statement.Table == target.table {
					tx.AddError(forcedFailure)
				}
			}
			require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callbackName, failWrite))
			require.NoError(t, env.db.Callback().Update().Before("gorm:update").Register(callbackName, failWrite))
			require.ErrorIs(t, target.write(valid), forcedFailure)
			require.NoError(t, env.db.Callback().Create().Remove(callbackName))
			require.NoError(t, env.db.Callback().Update().Remove(callbackName))
			require.NoError(t, env.db.First(&receipt, "id=?", receipt.ID).Error)
			require.Equal(t, "ready", receipt.Status)
			require.NoError(t, target.write(valid))
			require.NoError(t, env.db.First(&receipt, "id=?", receipt.ID).Error)
			require.Equal(t, "bound", receipt.Status)
			assertBusinessDraftMediaReceiptTestObjectsExist(t, drafts, asset, true)
		})
	}
}

func TestCategorySEOUnspecifiedFieldsReloadCommittedValuesAndTranslationNamesPreserveSEOIdentity(t *testing.T) {
	for _, locale := range []string{"en", "fr"} {
		t.Run(locale, func(t *testing.T) {
			env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, nil)
			field := "category_default"
			if locale == "fr" {
				field = "category_translation"
			}
			target := createTemplateOrCategoryIntroMediaMutationTestTarget(t, env, field)
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/detached.png")
			require.NoError(t, target.write(`<img src="`+asset.URL+`">`))
			var categoryID uint
			if locale == "en" {
				categoryID = target.id
			} else {
				require.NoError(t, env.db.Table(target.table).Where("id=?", target.id).Pluck("product_category_id", &categoryID).Error)
			}
			var detached atomic.Bool
			callbackName := "detach_before_category_seo_transaction_" + uuid.NewString()
			require.NoError(t, env.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
				if tx.Statement.Table != target.table || !detached.CompareAndSwap(false, true) {
					return
				}
				if err := env.db.Table(target.table).Where("id=?", target.id).Updates(map[string]any{target.column: "", "meta_description": "Newest committed description"}).Error; err != nil {
					tx.AddError(err)
				}
			}))
			newTitle := "Changed title only"
			_, err := env.categories.UpdateSEO(categoryID, locale, &newTitle, nil, nil)
			require.NoError(t, env.db.Callback().Query().Remove(callbackName))
			require.NoError(t, err)
			require.True(t, detached.Load())
			require.Empty(t, readTemplateOrCategoryIntroPersistedMediaHTML(t, env.db, target))
			view, err := env.categories.GetSEO(categoryID, locale)
			require.NoError(t, err)
			require.Equal(t, "Newest committed description", view.MetaDescription)
			if locale == "fr" {
				require.NoError(t, env.categories.repo.ReplaceTranslationNamesAndDescriptionsPreservingLockedSEO(categoryID, []product.ProductCategoryTranslation{{Locale: "fr", Name: "Updated name", Description: "/uploads/plain-text.png", SEOIntro: `<img src="` + asset.URL + `">`, MetaTitle: "Stale title"}, {Locale: "de", Name: "New translation", SEOIntro: `<img src="` + asset.URL + `">`}}))
				var translations []product.ProductCategoryTranslation
				require.NoError(t, env.db.Where("product_category_id=?", categoryID).Order("locale").Find(&translations).Error)
				require.Len(t, translations, 2)
				require.Empty(t, translations[0].SEOIntro)
				require.Empty(t, translations[1].SEOIntro)
				require.Equal(t, target.id, translations[1].ID)
				require.Equal(t, newTitle, translations[1].MetaTitle)
				require.NoError(t, env.categories.repo.ReplaceTranslationNamesAndDescriptionsPreservingLockedSEO(categoryID, nil))
				var remaining int64
				require.NoError(t, env.db.Model(&product.ProductCategoryTranslation{}).Where("product_category_id=?", categoryID).Count(&remaining).Error)
				require.Zero(t, remaining)
			}
		})
	}
}

func TestCategoryCreationValidatesImageAndNestedIntroAttachmentsInOneTransaction(t *testing.T) {
	env := configureTemplateAndCategoryIntroTransactionalMediaBindingTestEnvironment(t, nil)
	image := createRegisteredPublicImageForProductMediaBindingTest(t, env.db, "catalog/create.png")
	category := product.ProductCategory{Name: "Nested category", Slug: "nested-category", ImageMediaAssetID: &image.ID, SEOIntro: `<img src="` + image.URL + `">`, Translations: []product.ProductCategoryTranslation{{Locale: "fr", Name: "Nested translation", SEOIntro: `<img src="/uploads/missing.png">`}}}
	require.ErrorIs(t, env.categories.repo.Create(&category), repository.ErrProductCategoryImageReferenceInvalid)
	for _, table := range []string{"product_categories", "product_category_translations"} {
		var count int64
		require.NoError(t, env.db.Table(table).Count(&count).Error)
		require.Zero(t, count)
	}
	category.Translations[0].SEOIntro = fmt.Sprintf(`<img src="/uploads/%s">`, image.StorageKey)
	require.NoError(t, env.categories.repo.Create(&category))
	var translation product.ProductCategoryTranslation
	require.NoError(t, env.db.Where("product_category_id=?", category.ID).First(&translation).Error)
	require.Contains(t, translation.SEOIntro, image.URL)
}

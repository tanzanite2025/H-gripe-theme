package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/repository"
)

func createPostTransactionalMediaAndCategoryBindingTestServices(t *testing.T) (*gorm.DB, *PostService, *MediaService) {
	t.Helper()
	db, _ := newTestPostService(t)
	// Post's automatic many-to-many table has only foreign keys. Recreate this
	// empty test table with the explicit production relationship model first.
	require.NoError(t, db.Migrator().DropTable(&post.PostCategory{}))
	require.NoError(t, db.AutoMigrate(&post.Category{}, &post.PostCategory{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}))
	categories := repository.NewBlogCategoryRepository(db)
	posts := repository.NewPostRepository(db)
	posts.ConfigureCategoryRepository(categories)
	articles := NewPostService(posts, nil, 0, categories)
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	assets.siteURL = "https://shop.example.test"
	articles.ConfigureMediaService(assets)
	return db, articles, assets
}

func createPostInputForTransactionalMediaBindingTest(slug string) PostCreateInput {
	return PostCreateInput{Title: "Media article", Slug: slug, Content: "<p>Article</p>", Status: "draft", AuthorID: 42, Locale: "en"}
}

func createPostMarkupForMediaPlacementTest(field, reference string) string {
	switch field {
	case "image":
		return `<img src="` + reference + `">`
	case "video":
		return `<video src="` + reference + `"></video>`
	case "poster":
		return `<video src="https://external.example.test/demo.webm" poster="` + reference + `"></video>`
	default:
		return `<a href="` + reference + `">Download</a>`
	}
}

func TestPostCanonicalMediaReuseAcrossTranslationsDoesNotDuplicateAssetsAndDetachingPreservesAssets(t *testing.T) {
	db, articles, assets := createPostTransactionalMediaAndCategoryBindingTestServices(t)
	image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/article-photo.png")
	video := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/article-demo.webm")
	require.NoError(t, db.Model(video).Update("media_type", "video").Error)
	require.NoError(t, db.Model(image).Update("library_partition", media.MediaLibraryPartitionCatalog).Error)
	derivative := media.MediaAssetDerivative{MediaAssetID: image.ID, Preset: "thumbnail", StorageKey: "derivatives/article.webp", URL: "https://media.example.test/uploads/derivatives/article.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	input := createPostInputForTransactionalMediaBindingTest("canonical")
	input.FeaturedImg = "/uploads/" + derivative.StorageKey + "?v=1"
	input.Content = `<img src="https://shop.example.test/uploads/derivatives/article.webp?a=1&amp;b=2"><video src="/uploads/` + video.StorageKey + `#t=10" poster="/uploads/` + image.StorageKey + `"></video><a href="` + derivative.URL + `#page=2">Download</a>`
	first, err := articles.CreateAdminPost(input)
	require.NoError(t, err)
	require.Equal(t, image.URL, first.FeaturedImg)
	require.Contains(t, first.Content, image.URL+"#page=2")
	require.Contains(t, first.Content, video.URL+"#t=10")
	require.NotContains(t, first.Content, derivative.URL)
	input.Locale, input.Slug, input.TranslationGroupID = "fr", "canonical-fr", &first.ID
	second, err := articles.CreateAdminPost(input)
	require.NoError(t, err)
	var registrations int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Count(&registrations).Error)
	require.EqualValues(t, 2, registrations)
	report, err := assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	require.Equal(t, "featured_img, content", report.References[0].Field)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
	for _, item := range []*post.Post{first, second} {
		empty := ""
		_, err := articles.UpdateAdminPost(item.ID, PostUpdateInput{Content: &empty, FeaturedImg: &empty})
		require.NoError(t, err)
	}
	stored, err := assets.repo.FindAssetByID(image.ID)
	require.NoError(t, err)
	require.Equal(t, media.MediaLibraryPartitionCatalog, stored.LibraryPartition, "cross-domain reuse must not reclassify the original")
	require.NoError(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)))
}

func TestPostNewMediaPlacementsRejectUnregisteredOrIneligibleAssetsWithoutPartialWrites(t *testing.T) {
	for _, field := range []string{"cover", "image", "video", "poster", "file link"} {
		for _, scenario := range []string{"missing", "private", "customer", "conversation", "retired", "inactive", "ambiguous", "private namespace", "wrong type"} {
			if field == "file link" && scenario == "wrong type" {
				continue
			}
			t.Run(field+"/"+scenario, func(t *testing.T) {
				db, articles, _ := createPostTransactionalMediaAndCategoryBindingTestServices(t)
				asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/rejected.png")
				if field == "video" {
					require.NoError(t, db.Model(asset).Update("media_type", "video").Error)
				}
				reference := "/uploads/" + asset.StorageKey
				switch scenario {
				case "missing":
					reference = "/uploads/content/missing.png"
				case "private":
					require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
				case "customer":
					require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
				case "conversation":
					require.NoError(t, db.Model(asset).Update("upload_origin", media.MediaUploadOriginCustomerServiceConversation).Error)
				case "retired":
					require.NoError(t, db.Delete(asset).Error)
				case "inactive":
					require.NoError(t, db.Model(asset).Update("status", "inactive").Error)
				case "ambiguous":
					duplicate := *asset
					duplicate.ID, duplicate.URL = 0, "https://old.example.test/same-object.png"
					require.NoError(t, db.Create(&duplicate).Error)
				case "private namespace":
					reference = "/uploads/customer-service/attachments/originals/missing.png"
				case "wrong type":
					kind := "video"
					if field == "video" {
						kind = "image"
					}
					require.NoError(t, db.Model(asset).Update("media_type", kind).Error)
				}
				input := createPostInputForTransactionalMediaBindingTest("rejected")
				if field == "cover" {
					input.FeaturedImg = reference
				} else {
					input.Content = createPostMarkupForMediaPlacementTest(field, reference)
				}
				_, err := articles.CreateAdminPost(input)
				require.ErrorIs(t, err, ErrPostMediaReferenceInvalid)
				var count int64
				require.NoError(t, db.Unscoped().Model(&post.Post{}).Count(&count).Error)
				require.Zero(t, count)
			})
		}
	}
}

func TestPostHistoricalPlacementsRemainLocalAndOmittedFieldsCannotRestoreDetachedMedia(t *testing.T) {
	db, articles, assets := createPostTransactionalMediaAndCategoryBindingTestServices(t)
	legacy := &post.Post{Title: "Legacy", Slug: "legacy", Locale: "en", Content: `<p>Old</p><img src="/uploads/old-unregistered.png">`, FeaturedImg: "/uploads/old-cover.png"}
	require.NoError(t, db.Create(legacy).Error)
	title, changed := "Ordinary edit", strings.Replace(legacy.Content, "Old", "New", 1)
	_, err := articles.UpdateAdminPost(legacy.ID, PostUpdateInput{Title: &title})
	require.NoError(t, err)
	_, err = articles.UpdateAdminPost(legacy.ID, PostUpdateInput{Content: &changed})
	require.NoError(t, err)
	duplicate := changed + `<img src="/uploads/old-unregistered.png">`
	_, err = articles.UpdateAdminPost(legacy.ID, PostUpdateInput{Content: &duplicate})
	require.ErrorIs(t, err, ErrPostMediaReferenceInvalid)
	moved := "/uploads/old-unregistered.png"
	_, err = articles.UpdateAdminPost(legacy.ID, PostUpdateInput{FeaturedImg: &moved})
	require.ErrorIs(t, err, ErrPostMediaReferenceInvalid)
	clonedValue := createPostInputForTransactionalMediaBindingTest("legacy-fr")
	clonedValue.Locale, clonedValue.Content, clonedValue.FeaturedImg, clonedValue.TranslationGroupID = "fr", changed, legacy.FeaturedImg, &legacy.ID
	_, err = articles.CreateAdminPost(clonedValue)
	require.ErrorIs(t, err, ErrPostMediaReferenceInvalid)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/stale.png")
	input := createPostInputForTransactionalMediaBindingTest("stale")
	input.FeaturedImg, input.Content = asset.URL, `<img src="`+asset.URL+`">`
	created, err := articles.CreateAdminPost(input)
	require.NoError(t, err)
	stale, err := articles.postRepo.FindByID(created.ID)
	require.NoError(t, err)
	empty := ""
	_, err = articles.UpdateAdminPost(created.ID, PostUpdateInput{Content: &empty, FeaturedImg: &empty})
	require.NoError(t, err)
	require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	stale.Title = "Concurrent metadata edit"
	require.NoError(t, articles.postRepo.Update(stale, repository.PostMediaFieldMutation{}))
	current, err := articles.postRepo.FindByID(created.ID)
	require.NoError(t, err)
	require.Empty(t, current.Content)
	require.Empty(t, current.FeaturedImg)
	require.NoError(t, articles.Delete(created.ID))
	require.ErrorIs(t, articles.postRepo.Update(stale), gorm.ErrRecordNotFound)
}

func TestPostCategoryWriteFailureRollsBackArticleMediaAndExistingCategoryRelationships(t *testing.T) {
	for _, mutation := range []string{"create", "update"} {
		t.Run(mutation, func(t *testing.T) {
			db, articles, assets := createPostTransactionalMediaAndCategoryBindingTestServices(t)
			first := post.Category{Name: "First", Slug: "first", Locale: "en"}
			second := post.Category{Name: "Second", Slug: "second", Locale: "en"}
			require.NoError(t, db.Create(&first).Error)
			require.NoError(t, db.Create(&second).Error)
			asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/category-failure.png")
			var existing *post.Post
			if mutation == "update" {
				input := createPostInputForTransactionalMediaBindingTest("category-failure")
				input.CategoryIDs = []uint{first.ID}
				var err error
				existing, err = articles.CreateAdminPost(input)
				require.NoError(t, err)
			}
			failure := errors.New("category relationship insert failed")
			require.NoError(t, db.Callback().Create().Before("gorm:create").Register("fail_post_category_relationship", func(tx *gorm.DB) {
				if tx.Statement.Table == "post_categories" {
					tx.AddError(failure)
				}
			}))
			t.Cleanup(func() { db.Callback().Create().Remove("fail_post_category_relationship") })
			if mutation == "create" {
				input := createPostInputForTransactionalMediaBindingTest("category-failure")
				input.CategoryIDs, input.FeaturedImg = []uint{second.ID}, asset.URL
				_, err := articles.CreateAdminPost(input)
				require.ErrorIs(t, err, failure)
				var count int64
				require.NoError(t, db.Unscoped().Model(&post.Post{}).Count(&count).Error)
				require.Zero(t, count, "failed creation must not leave soft-deleted media placements")
			} else {
				body := `<img src="` + asset.URL + `">`
				_, err := articles.UpdateAdminPost(existing.ID, PostUpdateInput{Content: &body, FeaturedImg: &asset.URL, CategoryIDs: []uint{second.ID}, UpdateCategoryIDs: true})
				require.ErrorIs(t, err, failure)
				current, err := articles.postRepo.FindByID(existing.ID)
				require.NoError(t, err)
				require.Equal(t, existing.Content, current.Content)
				require.Empty(t, current.FeaturedImg)
				require.Len(t, current.Categories, 1)
				require.Equal(t, first.ID, current.Categories[0].ID)
			}
			report, err := assets.GetAssetReferences(asset.ID)
			require.NoError(t, err)
			require.Zero(t, report.Total)
		})
	}
}

func TestPostSEOConcurrentEditUpdatesOnlySEOFieldsWithoutRestoringDetachedMediaOrCategories(t *testing.T) {
	db, articles, assets := createPostTransactionalMediaAndCategoryBindingTestServices(t)
	first := post.Category{Name: "Original category", Slug: "original-category", Locale: "en"}
	second := post.Category{Name: "Concurrent category", Slug: "concurrent-category", Locale: "en"}
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, db.Create(&second).Error)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/seo-stale.png")
	input := createPostInputForTransactionalMediaBindingTest("seo-stale")
	input.FeaturedImg, input.Content = asset.URL, `<img src="`+asset.URL+`">`
	input.CategoryIDs = []uint{first.ID}
	created, err := articles.CreateAdminPost(input)
	require.NoError(t, err)
	injected := false
	require.NoError(t, db.Callback().Query().After("gorm:after_query").Register("detach_post_media_after_seo_snapshot", func(tx *gorm.DB) {
		item, ok := tx.Statement.Dest.(*post.Post)
		if injected || tx.Statement.Table != "posts" || !ok || item.ID != created.ID {
			return
		}
		injected = true
		require.NoError(t, db.Model(&post.Post{}).Where("id = ?", created.ID).Updates(map[string]interface{}{"content": "<p>Detached</p>", "featured_img": "", "title": "Concurrent title", "meta_desc": "Concurrent description"}).Error)
		require.NoError(t, db.Where("post_id = ?", created.ID).Delete(&post.PostCategory{}).Error)
		require.NoError(t, db.Create(&post.PostCategory{PostID: created.ID, CategoryID: second.ID}).Error)
		require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	}))
	defer db.Callback().Query().Remove("detach_post_media_after_seo_snapshot")
	seoTitle := "Updated SEO title"
	current, err := articles.UpdatePostSEO(created.ID, PostSEOUpdateInput{MetaTitle: &seoTitle})
	require.NoError(t, err)
	require.True(t, injected)
	require.Equal(t, "<p>Detached</p>", current.Content)
	require.Empty(t, current.FeaturedImg)
	require.Equal(t, "Concurrent title", current.Title)
	require.Equal(t, "Concurrent description", current.MetaDesc)
	require.Equal(t, seoTitle, current.MetaTitle)
	require.Len(t, current.Categories, 1)
	require.Equal(t, second.ID, current.Categories[0].ID)
}

func TestPostExplicitCategoryReplacementAndStaleAssociationSnapshotsRemainSeparateFromMediaUpdates(t *testing.T) {
	db, articles, _ := createPostTransactionalMediaAndCategoryBindingTestServices(t)
	first := post.Category{Name: "First", Slug: "first", Locale: "en"}
	second := post.Category{Name: "Second", Slug: "second", Locale: "en"}
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, db.Create(&second).Error)
	input := createPostInputForTransactionalMediaBindingTest("explicit-categories")
	input.CategoryIDs = []uint{first.ID, first.ID}
	created, err := articles.CreateAdminPost(input)
	require.NoError(t, err)
	require.Len(t, created.Categories, 1)
	stale, err := articles.postRepo.FindByID(created.ID)
	require.NoError(t, err)
	_, err = articles.UpdateAdminPost(created.ID, PostUpdateInput{CategoryIDs: []uint{second.ID}, UpdateCategoryIDs: true})
	require.NoError(t, err)
	stale.Title, stale.Categories[0].Name = "Ordinary metadata edit", "Stale category name"
	require.NoError(t, articles.postRepo.Update(stale, repository.PostMediaFieldMutation{}))
	current, err := articles.postRepo.FindByID(created.ID)
	require.NoError(t, err)
	require.Len(t, current.Categories, 1)
	require.Equal(t, second.ID, current.Categories[0].ID)
	var unchanged post.Category
	require.NoError(t, db.First(&unchanged, first.ID).Error)
	require.Equal(t, "First", unchanged.Name)
	_, err = articles.UpdateAdminPost(created.ID, PostUpdateInput{UpdateCategoryIDs: true})
	require.NoError(t, err)
	current, err = articles.postRepo.FindByID(created.ID)
	require.NoError(t, err)
	require.Empty(t, current.Categories)
}

func TestPostReferenceScannerMatchesExactRenderedMediaAttributesAndRetainedDeletedArticles(t *testing.T) {
	db, articles, assets := createPostTransactionalMediaAndCategoryBindingTestServices(t)
	asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/scanner.png")
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", StorageKey: "derivatives/scanner.webp", URL: "https://media.example.test/uploads/derivatives/scanner.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	rows := make([]post.Post, 257)
	for index := range rows {
		rows[index] = post.Post{Title: "Non-reference", Slug: fmt.Sprintf("non-reference-%d", index), Locale: "en", Content: `<p>` + asset.URL + `</p><script><img src="` + asset.URL + `"></script><img src="` + asset.URL + `.backup"><img src="https://foreign.example.test/uploads/` + asset.StorageKey + `">`}
	}
	rows[256].FeaturedImg = "/uploads/" + asset.StorageKey + "?v=1"
	rows[256].Content = `<video src="https://external.example.test/video.webm" poster="https://shop.example.test/uploads/derivatives/scanner.webp?a=1&amp;b=2"></video><a href="` + derivative.URL + `#page=2">Download</a>`
	require.NoError(t, db.Create(&rows).Error)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, rows[256].ID, report.References[0].ResourceID)
	require.Equal(t, "featured_img, content", report.References[0].Field)
	require.NoError(t, articles.Delete(rows[256].ID))
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, "retained_deleted_parent", report.References[0].UsageState)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
}

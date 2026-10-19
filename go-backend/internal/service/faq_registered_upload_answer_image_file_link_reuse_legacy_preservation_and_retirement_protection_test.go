package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/testfixtures"
)

func createFAQTransactionalMediaBindingTestServices(t *testing.T) (*gorm.DB, *FAQService, *MediaService) {
	t.Helper()
	db, faqs := newTestFAQService(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.BusinessDraftMediaUploadReceipt{}))
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	assets.siteURL = "https://shop.example.test"
	faqs.ConfigureMediaService(assets)
	for _, locale := range []string{"en", "fr"} {
		require.NoError(t, db.Create(&faq.FAQPage{PageID: "media-test", RoutePath: "/support/media", Locale: locale, Title: "FAQ media", Status: "active"}).Error)
	}
	return db, faqs, assets
}

func createRegistered800PixelWebPAssetForFAQMediaBindingTest(t *testing.T, db *gorm.DB, key string) *media.MediaAsset {
	t.Helper()
	asset := &media.MediaAsset{Filename: "answer.webp", StorageKey: key, URL: "https://media.example.test/uploads/" + key, MediaType: "image", MimeType: "image/webp", Width: 800, Height: 800, Visibility: "public", Status: "active"}
	require.NoError(t, db.Create(asset).Error)
	return asset
}

func createFAQInputWithAnswerImageForMediaBindingTest(reference string) *faq.FAQ {
	return &faq.FAQ{Question: "How does this work?", Answer: "<p>Answer</p>", AnswerImageURL: reference, AnswerImageAlt: "Answer illustration", AnswerImageWidth: 800, AnswerImageHeight: 800, PageID: "media-test", Locale: "en", Status: "draft"}
}

func TestFAQRegisteredUploadReusesOneAssetAcrossAnswersAndDetachingNeverDeletesTheFile(t *testing.T) {
	db, faqs, assets := createFAQTransactionalMediaBindingTestServices(t)
	file := multipartFileHeader(t, "faq-answer.webp", "image/webp", testfixtures.FAQAnswerImageUploadRegistrationAndBinding800By800LosslessWebPFixtureBytes())
	url, err := faqs.UploadAnswerImage(t.Context(), file, media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42}, "Answer illustration")
	require.NoError(t, err)
	var asset media.MediaAsset
	require.NoError(t, db.First(&asset).Error)
	require.Equal(t, media.MediaLibraryPartitionContent, asset.LibraryPartition)
	require.Equal(t, media.MediaUploadOriginAdministratorContent, asset.UploadOrigin)
	require.Equal(t, string(upload.SpecFAQAnswerImage), asset.UploadPurpose)
	require.EqualValues(t, 42, asset.UploaderID)
	require.Equal(t, media.MediaUploadActorAdministrator, asset.UploaderActorType)
	require.True(t, strings.HasPrefix(asset.StorageKey, storage.BusinessDraftMediaObjectPrefix+"/"))
	require.Equal(t, 800, asset.Width)
	require.Equal(t, 800, asset.Height)
	first := createFAQInputWithAnswerImageForMediaBindingTest("/uploads/" + asset.StorageKey + "?v=1")
	first.Answer = `<p><a href="` + url + `#page=2">Download</a></p>`
	require.NoError(t, faqs.Create(first))
	require.Equal(t, url, first.AnswerImageURL)
	require.Contains(t, first.Answer, url+"#page=2")
	second := createFAQInputWithAnswerImageForMediaBindingTest(url)
	second.Locale, second.ParentID = "fr", &first.ID
	require.NoError(t, faqs.Create(second))
	var registrations int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Count(&registrations).Error)
	require.EqualValues(t, 1, registrations, "reuse and translation must not create a second asset")
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	require.Equal(t, "answer_image_url, answer", report.References[0].Field)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
	for _, id := range []uint{first.ID, second.ID} {
		_, err := faqs.UpdateAdminFAQ(id, FAQAdminUpdateInput{Answer: "<p>Only text</p>", AnswerImageSet: true})
		require.NoError(t, err)
	}
	object, err := assets.storage.(storage.ObjectOpener).Open(t.Context(), asset.StorageKey)
	require.NoError(t, err, "detachment must preserve the original file")
	require.Greater(t, object.Size, int64(0))
	require.NoError(t, object.ReadCloser.Close())
	require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
}

func TestFAQNewImageBindingsRejectUnregisteredForeignIneligibleAndIncorrectOriginalsAtomically(t *testing.T) {
	for _, scenario := range []string{"missing", "foreign host", "retired", "private", "customer", "conversation", "video", "wrong MIME", "wrong dimensions", "ambiguous", "800px derivative of a larger original"} {
		t.Run(scenario, func(t *testing.T) {
			db, faqs, _ := createFAQTransactionalMediaBindingTestServices(t)
			asset := createRegistered800PixelWebPAssetForFAQMediaBindingTest(t, db, "content/answer.webp")
			input := createFAQInputWithAnswerImageForMediaBindingTest(asset.URL)
			switch scenario {
			case "missing":
				input.AnswerImageURL = "/uploads/content/missing.webp"
			case "foreign host":
				input.AnswerImageURL = "https://foreign.example.test/uploads/" + asset.StorageKey
			case "retired":
				require.NoError(t, db.Delete(asset).Error)
			case "private":
				require.NoError(t, db.Model(asset).Update("visibility", "private").Error)
			case "customer":
				require.NoError(t, db.Model(asset).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
			case "conversation":
				require.NoError(t, db.Model(asset).Update("upload_origin", media.MediaUploadOriginCustomerServiceConversation).Error)
			case "video":
				require.NoError(t, db.Model(asset).Update("media_type", "video").Error)
			case "wrong MIME":
				require.NoError(t, db.Model(asset).Update("mime_type", "image/png").Error)
			case "wrong dimensions":
				require.NoError(t, db.Model(asset).Update("width", 1600).Error)
			case "ambiguous":
				duplicate := *asset
				duplicate.ID, duplicate.URL = 0, "https://old.example.test/same-object.webp"
				require.NoError(t, db.Create(&duplicate).Error)
			case "800px derivative of a larger original":
				require.NoError(t, db.Model(asset).Update("width", 1600).Error)
				derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", URL: "https://media.example.test/uploads/derivatives/answer.webp", StorageKey: "derivatives/answer.webp", Width: 800, Height: 800}
				require.NoError(t, db.Create(&derivative).Error)
				input.AnswerImageURL = derivative.URL
			}
			require.ErrorIs(t, faqs.Create(input), ErrFAQMediaReferenceInvalid)
			var count int64
			require.NoError(t, db.Model(&faq.FAQ{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
	for _, placement := range []string{"dedicated image", "answer link"} {
		t.Run("invalid replacement preserves old FAQ/"+placement, func(t *testing.T) {
			db, faqs, _ := createFAQTransactionalMediaBindingTestServices(t)
			asset := createRegistered800PixelWebPAssetForFAQMediaBindingTest(t, db, "content/retained.webp")
			item := createFAQInputWithAnswerImageForMediaBindingTest(asset.URL)
			require.NoError(t, faqs.Create(item))
			input := FAQAdminUpdateInput{Question: "Must roll back"}
			if placement == "answer link" {
				input.Answer = `<a href="/uploads/missing.webp">Download</a>`
			} else {
				input.AnswerImageSet, input.AnswerImageURL, input.AnswerImageAlt, input.AnswerImageWidth, input.AnswerImageHeight = true, "/uploads/missing.webp", "Missing", 800, 800
			}
			_, err := faqs.UpdateAdminFAQ(item.ID, input)
			require.ErrorIs(t, err, ErrFAQMediaReferenceInvalid)
			current, err := faqs.GetByID(item.ID)
			require.NoError(t, err)
			require.Equal(t, item.Question, current.Question)
			require.Equal(t, item.AnswerImageURL, current.AnswerImageURL)
		})
	}
}

func TestFAQLegacyImageAndLinkPlacementsStayLocalAndCannotBeCopiedOrRestoredFromStaleEdits(t *testing.T) {
	db, faqs, assets := createFAQTransactionalMediaBindingTestServices(t)
	legacy := createFAQInputWithAnswerImageForMediaBindingTest("/uploads/legacy/answer.webp")
	legacy.Answer = `<p><a href="/uploads/legacy/manual.pdf">Old</a></p>`
	require.NoError(t, db.Create(legacy).Error)
	_, err := faqs.UpdateAdminFAQ(legacy.ID, FAQAdminUpdateInput{Question: "Ordinary edit"})
	require.NoError(t, err)
	_, err = faqs.UpdateAdminFAQ(legacy.ID, FAQAdminUpdateInput{Answer: strings.Replace(legacy.Answer, "Old", "New", 1), AnswerImageSet: true, AnswerImageURL: legacy.AnswerImageURL, AnswerImageAlt: "Still valid", AnswerImageWidth: 800, AnswerImageHeight: 800})
	require.NoError(t, err)
	_, err = faqs.UpdateAdminFAQ(legacy.ID, FAQAdminUpdateInput{Answer: legacy.Answer + legacy.Answer})
	require.ErrorIs(t, err, ErrFAQMediaReferenceInvalid)
	translatedLegacyFAQ := *legacy
	translatedLegacyFAQ.ID, translatedLegacyFAQ.Locale, translatedLegacyFAQ.ParentID = 0, "fr", &legacy.ID
	require.ErrorIs(t, faqs.Create(&translatedLegacyFAQ), ErrFAQMediaReferenceInvalid)
	_, err = faqs.UpdateAdminFAQ(legacy.ID, FAQAdminUpdateInput{Answer: "<p>Detached</p>", AnswerImageSet: true})
	require.NoError(t, err)
	_, err = faqs.UpdateAdminFAQ(legacy.ID, FAQAdminUpdateInput{AnswerImageSet: true, AnswerImageURL: legacy.AnswerImageURL, AnswerImageAlt: "Cannot rebind", AnswerImageWidth: 800, AnswerImageHeight: 800})
	require.ErrorIs(t, err, ErrFAQMediaReferenceInvalid)
	asset := createRegistered800PixelWebPAssetForFAQMediaBindingTest(t, db, "content/stale.webp")
	item := createFAQInputWithAnswerImageForMediaBindingTest(asset.URL)
	item.Answer = `<a href="` + asset.URL + `">Download</a>`
	require.NoError(t, faqs.Create(item))
	stale, err := faqs.GetByID(item.ID)
	require.NoError(t, err)
	_, err = faqs.UpdateAdminFAQ(item.ID, FAQAdminUpdateInput{Answer: "<p>Detached</p>", AnswerImageSet: true})
	require.NoError(t, err)
	require.NoError(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	stale.Question = "Concurrent ordinary edit"
	require.NoError(t, faqs.faqRepo.Update(stale, repository.FAQMediaFieldMutation{}))
	current, err := faqs.GetByID(item.ID)
	require.NoError(t, err)
	require.Equal(t, "<p>Detached</p>", current.Answer)
	require.Empty(t, current.AnswerImageURL)
	require.NoError(t, faqs.Delete(item.ID))
	require.ErrorIs(t, faqs.faqRepo.Update(stale), gorm.ErrRecordNotFound)
}

func TestFAQReferenceScannerMatchesRenderedLinksAndImagesAndProtectsRetainedDeletedAnswers(t *testing.T) {
	db, faqs, assets := createFAQTransactionalMediaBindingTestServices(t)
	asset := createRegistered800PixelWebPAssetForFAQMediaBindingTest(t, db, "content/report.webp")
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", URL: "https://media.example.test/uploads/derivatives/report.webp", StorageKey: "derivatives/report.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	rows := make([]faq.FAQ, 257)
	for index := range rows {
		rows[index] = faq.FAQ{Question: fmt.Sprintf("Non-reference %d", index), Answer: `<p>` + asset.URL + `</p><img src="` + asset.URL + `"><video src="` + asset.URL + `"></video><a href="https://foreign.example.test/uploads/` + asset.StorageKey + `">Foreign</a><a href="` + asset.URL + `.backup">Other</a>`, Locale: "en"}
	}
	rows[256].AnswerImageURL = "/uploads/" + asset.StorageKey + "?cache=1"
	rows[256].Answer = `<a href="https://shop.example.test/uploads/derivatives/report.webp?a=1&amp;b=2#page=2">Download</a>`
	require.NoError(t, db.Create(&rows).Error)
	report, err := assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, rows[256].ID, report.References[0].ResourceID)
	require.Equal(t, "answer_image_url, answer", report.References[0].Field)
	require.NoError(t, faqs.Delete(rows[256].ID))
	report, err = assets.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, "retained_deleted_parent", report.References[0].UsageState)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
}

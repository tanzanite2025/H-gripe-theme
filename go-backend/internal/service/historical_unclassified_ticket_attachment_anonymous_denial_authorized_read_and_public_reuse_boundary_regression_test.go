package service

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

func createHistoricalUnclassifiedPublicTicketAttachmentAccessTestEnvironment(t *testing.T) customerServicePrivateAttachmentAccessTestEnvironment {
	t.Helper()
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	key := "2026/08/20/historical-customer-photo.png"
	asset := media.MediaAsset{Filename: filepath.Base(key), URL: "http://localhost:9200/uploads/" + key, StorageKey: key,
		MediaType: "image", MimeType: "image/png", Status: "active", Visibility: "public", Size: int64(len(env.originalPayload))}
	require.NoError(t, env.db.Create(&asset).Error)
	for _, objectKey := range []string{key, "media-derivatives/historical-customer-photo-thumbnail.png"} {
		filename := filepath.Join(env.publicRoot, filepath.FromSlash(objectKey))
		require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0700))
		require.NoError(t, os.WriteFile(filename, env.originalPayload, 0600))
	}
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", PresetVersion: 1,
		StorageKey: "media-derivatives/historical-customer-photo-thumbnail.png", MimeType: "image/png", Size: asset.Size, Width: 64, Height: 32}
	derivative.URL = "http://localhost:9200/uploads/" + derivative.StorageKey
	require.NoError(t, env.db.Create(&derivative).Error)
	asset.Derivatives = []media.MediaAssetDerivative{derivative}
	env.asset = asset
	attachments, err := json.Marshal([]string{asset.URL})
	require.NoError(t, err)
	env.message.Attachments = string(attachments)
	env.message.AttachmentReferenceFormat = ticket.TicketMessageLegacyURLAttachments
	require.NoError(t, env.db.Save(&env.message).Error)
	return env
}

func TestHistoricalUnclassifiedTicketAttachmentBlocksAnonymousReadsAndNewBindingsButStreamsToConversationOwner(t *testing.T) {
	env := createHistoricalUnclassifiedPublicTicketAttachmentAccessTestEnvironment(t)
	for _, key := range []string{env.asset.StorageKey, env.asset.Derivatives[0].StorageKey} {
		access, err := env.mediaService.PublicUploadAssetAccess(key)
		require.NoError(t, err)
		require.True(t, access.Found)
		require.False(t, access.Allowed)
	}
	for _, reference := range []string{env.asset.URL, "/uploads/" + env.asset.StorageKey} {
		_, err := env.mediaService.CanonicalPublicImageUploadURL(reference)
		require.ErrorIs(t, err, ErrMediaAssetForbidden)
		_, err = env.mediaService.CanonicalPublicMediaAttachmentURL(reference)
		require.ErrorIs(t, err, ErrMediaAssetForbidden)
	}
	projected, err := env.mediaService.GetAsset(env.asset.ID)
	require.NoError(t, err)
	require.True(t, projected.RequiresConversationAccess)
	require.Empty(t, projected.AccessURL)
	_, err = env.mediaService.OpenAssetFile(t.Context(), env.asset.ID)
	require.ErrorIs(t, err, ErrMediaAssetForbidden)
	_, _, err = env.mediaService.ExportCopyrightEvidence(t.Context(), env.asset.ID)
	require.ErrorIs(t, err, ErrMediaAssetForbidden)
	for _, preset := range []string{"", "thumbnail"} {
		file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, preset)
		require.NoError(t, err)
		require.Empty(t, file.RedirectURL)
		payload, err := io.ReadAll(file.ReadCloser)
		require.NoError(t, err)
		require.NoError(t, file.ReadCloser.Close())
		require.Equal(t, env.originalPayload, payload)
	}
	urls, descriptions := env.access.DescribePublicCustomerServiceMessageAttachments(env.message, *env.conversation.ConversationID)
	require.Len(t, descriptions, 1)
	require.Equal(t, []string{descriptions[0].AccessURL}, urls)
	agentURLs, _ := env.access.DescribeAgentCustomerServiceMessageAttachments(env.message)
	require.Contains(t, agentURLs[0], "/api/admin/customer-service/conversations/")
	metadata, err := json.Marshal(map[string]any{"url": env.asset.URL, "nested": []any{map[string]any{"thumbnail": env.asset.Derivatives[0].URL + "?v=1"}}, "caption": "Customer photo"})
	require.NoError(t, err)
	rewritten := env.access.MetadataWithAuthorizedProtectedAttachmentURLs(string(metadata), descriptions)
	require.NotContains(t, rewritten, "localhost:9200")
	require.Contains(t, rewritten, descriptions[0].AccessURL+"?preset=thumbnail")
	require.Contains(t, rewritten, "Customer photo")
	var unchanged media.MediaAsset
	require.NoError(t, env.db.First(&unchanged, env.asset.ID).Error)
	require.Equal(t, media.MediaUploadOriginLegacyUnknown, unchanged.UploadOrigin)
	require.Equal(t, media.MediaLibraryPartitionUnclassified, unchanged.LibraryPartition)
	require.Equal(t, int64(0), unchanged.ClassificationRevision)
	require.Equal(t, "public", unchanged.Visibility)
	for _, key := range []string{unchanged.StorageKey, env.asset.Derivatives[0].StorageKey} {
		payload, err := os.ReadFile(filepath.Join(env.publicRoot, filepath.FromSlash(key)))
		require.NoError(t, err)
		require.Equal(t, env.originalPayload, payload)
	}
}

func TestHistoricalTicketAttachmentProtectionUsesExactJSONReferencesIncludingRetainedDerivatives(t *testing.T) {
	for _, scenario := range []string{"registered old origin", "current trusted alias", "relative derivative", "nested escaped derivative", "query and fragment", "unknown historical origin", "foreign host complete key", "foreign host different key", "prefixed upload path", "filename suffix", "unreferenced", "orphaned message", "retired derivative"} {
		t.Run(scenario, func(t *testing.T) {
			env := createHistoricalUnclassifiedPublicTicketAttachmentAccessTestEnvironment(t)
			value := []any{env.asset.URL}
			wantProtected := true
			switch scenario {
			case "current trusted alias":
				value = []any{"https://shop.example.test/uploads/" + env.asset.StorageKey}
			case "relative derivative":
				value = []any{"/uploads/" + env.asset.Derivatives[0].StorageKey}
			case "nested escaped derivative":
				value = []any{map[string]any{"items": []string{env.asset.Derivatives[0].URL}}}
			case "query and fragment":
				value = []any{env.asset.URL + "?v=1#photo"}
			case "unknown historical origin":
				value = []any{"http://localhost:9199/uploads/" + env.asset.StorageKey}
			case "foreign host complete key":
				value = []any{"https://foreign.example/uploads/" + env.asset.StorageKey}
			case "foreign host different key":
				value, wantProtected = []any{"https://foreign.example/uploads/other/" + env.asset.Filename}, false
			case "prefixed upload path":
				value, wantProtected = []any{"https://foreign.example/proxy/uploads/" + env.asset.StorageKey}, false
			case "filename suffix":
				value, wantProtected = []any{env.asset.URL + ".decoy.png"}, false
			case "unreferenced":
				value, wantProtected = nil, false
			case "orphaned message":
				require.NoError(t, env.db.Model(&env.message).Update("ticket_id", env.conversation.ID+1000).Error)
			case "retired derivative":
				value = []any{env.asset.Derivatives[0].URL}
				require.NoError(t, env.db.Delete(&env.asset.Derivatives[0]).Error)
			}
			encoded, err := json.Marshal(value)
			require.NoError(t, err)
			if scenario == "nested escaped derivative" {
				encoded = []byte(strings.ReplaceAll(string(encoded), "/", `\/`))
			}
			require.NoError(t, env.db.Model(&env.message).Update("attachments", string(encoded)).Error)
			required, err := env.mediaService.repo.HistoricalUnclassifiedMediaRequiresConversationAccess(&env.asset)
			require.NoError(t, err)
			require.Equal(t, wantProtected, required)
		})
	}
}

func TestHistoricalAttachmentReviewedOriginRestoresOnlyExistingConversationReadsWithoutPublicBusinessTrust(t *testing.T) {
	env := createHistoricalUnclassifiedPublicTicketAttachmentAccessTestEnvironment(t)
	reference := "http://localhost:9199/uploads/" + env.asset.StorageKey
	replaceCustomerServiceAttachmentAccessTestMessageReference(t, env, reference)
	require.NoError(t, env.db.First(&env.message, env.message.ID).Error)
	allowed, err := env.mediaService.CanServePublicUpload(env.asset.StorageKey)
	require.NoError(t, err)
	require.False(t, allowed)
	file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, "")
	require.Error(t, err)
	require.Nil(t, file)
	require.Zero(t, env.storage.openCalls)
	require.NoError(t, env.mediaService.ConfigureReviewedHistoricalCustomerServiceAttachmentOrigins("http://localhost:9199"))
	for _, preset := range []string{"", "thumbnail"} {
		file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, preset)
		require.NoError(t, err)
		require.NoError(t, file.ReadCloser.Close())
	}
	urls, descriptions := env.access.DescribePublicCustomerServiceMessageAttachments(env.message, *env.conversation.ConversationID)
	require.Len(t, descriptions, 1)
	require.Equal(t, []string{descriptions[0].AccessURL}, urls)
	metadata, _ := json.Marshal(map[string]any{"url": reference, "thumbnail": "http://localhost:9199/uploads/" + env.asset.Derivatives[0].StorageKey})
	rewritten := env.access.MetadataWithAuthorizedProtectedAttachmentURLs(string(metadata), descriptions)
	require.NotContains(t, rewritten, "localhost:9199")
	require.Contains(t, rewritten, "?preset=thumbnail")
	key, err := env.mediaService.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
	require.NoError(t, err)
	require.Empty(t, key)
	_, err = env.mediaService.CanonicalPublicMediaAttachmentURL(reference)
	require.ErrorIs(t, err, ErrMediaAssetForbidden)
	for _, foreignOrigin := range []string{"https://localhost:9199", "http://localhost:9198", "http://foreign.example:9199"} {
		replaceCustomerServiceAttachmentAccessTestMessageReference(t, env, foreignOrigin+"/uploads/"+env.asset.StorageKey)
		file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, env.message.ID, 0, "")
		require.Error(t, err)
		require.Nil(t, file)
	}
}

func TestHistoricalAttachmentReviewedOriginConfigurationRejectsBroadOrCredentialedAddressesAtomically(t *testing.T) {
	env := createHistoricalUnclassifiedPublicTicketAttachmentAccessTestEnvironment(t)
	require.NoError(t, env.mediaService.ConfigureReviewedHistoricalCustomerServiceAttachmentOrigins("http://localhost:9199,https://old-shop.example.test/"))
	for _, invalid := range []string{"*", "https://*.example.test", "http://:9199", "//old-shop.example.test", "ftp://old-shop.example.test", "https://user:password@old-shop.example.test", "https://old-shop.example.test/uploads", "https://old-shop.example.test?", "https://old-shop.example.test?query=1", "https://old-shop.example.test#fragment", "https://old-shop.example.test#", "https://old-shop.example.test,"} {
		require.Error(t, env.mediaService.ConfigureReviewedHistoricalCustomerServiceAttachmentOrigins(invalid))
		require.Equal(t, env.asset.StorageKey, env.mediaService.reviewedHistoricalCustomerServiceAttachmentObjectKey("http://localhost:9199/uploads/"+env.asset.StorageKey))
	}
	require.NoError(t, env.mediaService.ConfigureReviewedHistoricalCustomerServiceAttachmentOrigins(""))
	require.Empty(t, env.mediaService.reviewedHistoricalCustomerServiceAttachmentObjectKey("http://localhost:9199/uploads/"+env.asset.StorageKey))
}

func TestHistoricalTicketAttachmentPublishedPublicUseAndReviewedOwnershipKeepTheirPublicAccess(t *testing.T) {
	for _, scenario := range []string{"published", "draft", "deleted post", "explicit shared ownership", "same bytes different key"} {
		t.Run(scenario, func(t *testing.T) {
			env := createHistoricalUnclassifiedPublicTicketAttachmentAccessTestEnvironment(t)
			wantAllowed := scenario == "published" || scenario == "explicit shared ownership"
			switch scenario {
			case "explicit shared ownership":
				require.NoError(t, env.db.Model(&env.asset).Updates(map[string]any{"upload_origin": media.MediaUploadOriginAdministratorShared, "library_partition": media.MediaLibraryPartitionShared, "classification_method": media.MediaClassificationRecordedUploadContext}).Error)
			case "same bytes different key":
				other := media.MediaAsset{Filename: "same-bytes.png", URL: "https://shop.example.test/uploads/shared/same-bytes.png", StorageKey: "shared/same-bytes.png", ContentSHA256: strings.Repeat("a", 64)}
				require.NoError(t, env.db.Create(&other).Error)
				require.NoError(t, env.db.Model(&env.asset).Update("content_sha256", other.ContentSHA256).Error)
				allowed, err := env.mediaService.CanServePublicUpload(other.StorageKey)
				require.NoError(t, err)
				require.True(t, allowed)
			default:
				require.NoError(t, env.db.AutoMigrate(&post.Post{}))
				item := post.Post{Title: "Public use", Slug: "public-use", Status: "published", FeaturedImg: env.asset.URL}
				if scenario == "draft" {
					item.Status = "draft"
				}
				require.NoError(t, env.db.Create(&item).Error)
				if scenario == "deleted post" {
					require.NoError(t, env.db.Delete(&item).Error)
				}
			}
			allowed, err := env.mediaService.CanServePublicUpload(env.asset.StorageKey)
			require.NoError(t, err)
			require.Equal(t, wantAllowed, allowed)
			require.NoError(t, env.db.AutoMigrate(&post.Post{}))
			posts := repository.NewPostRepository(env.db)
			posts.ConfigurePublicBusinessMediaReferenceIdentityResolver(env.mediaService)
			newPublicUse := post.Post{Title: "New public use", Slug: "new-public-use", Status: "published", FeaturedImg: env.asset.URL}
			err = posts.Create(&newPublicUse)
			if wantAllowed {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, repository.ErrPostMediaReferenceInvalid)
				var count int64
				require.NoError(t, env.db.Model(&post.Post{}).Where("slug = ?", newPublicUse.Slug).Count(&count).Error)
				require.Zero(t, count)
			}
		})
	}
}

func TestHistoricalAttachmentAuthorizedReadRejectsForeignOwnersOriginsAndAmbiguousRegistrations(t *testing.T) {
	for _, scenario := range []string{"other visitor", "internal message", "foreign URL suffix", "duplicate storage key", "retired duplicate key", "reassigned agent"} {
		t.Run(scenario, func(t *testing.T) {
			env := createHistoricalUnclassifiedPublicTicketAttachmentAccessTestEnvironment(t)
			owner := env.owner
			switch scenario {
			case "other visitor":
				owner.VisitorSessionHash = strings.Repeat("b", 64)
			case "internal message":
				require.NoError(t, env.db.Model(&env.message).Update("is_internal", true).Error)
			case "foreign URL suffix":
				replaceCustomerServiceAttachmentAccessTestMessageReference(t, env, "https://foreign.example/uploads/"+env.asset.StorageKey)
			case "duplicate storage key", "retired duplicate key":
				other := media.MediaAsset{Filename: env.asset.Filename, URL: "https://other.example/uploads/" + env.asset.StorageKey, StorageKey: env.asset.StorageKey}
				require.NoError(t, env.db.Create(&other).Error)
				if scenario == "retired duplicate key" {
					require.NoError(t, env.db.Delete(&other).Error)
				}
				// Trusted aliases may never pick the first registration.
				replaceCustomerServiceAttachmentAccessTestMessageReference(t, env, "/uploads/"+env.asset.StorageKey)
				allowed, err := env.mediaService.CanServePublicUpload(env.asset.StorageKey)
				require.NoError(t, err)
				require.False(t, allowed)
			case "reassigned agent":
				previousAgent := env.conversation.AssignedTo
				require.NoError(t, env.db.Model(&env.conversation).Update("assigned_to", env.conversation.AssignedTo+100).Error)
				file, err := env.access.OpenAgentCustomerServiceMessageAttachment(t.Context(), env.conversation.ID, previousAgent, false, env.message.ID, 0, "")
				require.Error(t, err)
				require.Nil(t, file)
				require.Zero(t, env.storage.openCalls)
				return
			}
			file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, owner, env.message.ID, 0, "")
			require.Error(t, err)
			require.Nil(t, file)
			require.Zero(t, env.storage.openCalls)
		})
	}
}

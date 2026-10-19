package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createRetiredMediaDerivativeWithDurableCleanupEventForTest(t *testing.T, db *gorm.DB, assets *MediaService, root string) (*media.MediaAsset, media.MediaAssetDerivative, media.MediaAssetDerivative, outbox.Event) {
	t.Helper()
	parent := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, assets.storage, "catalog/generation-owner.png")
	require.NoError(t, db.Model(parent).Updates(map[string]any{"upload_origin": media.MediaUploadOriginAdministratorShared, "classification_method": media.MediaClassificationRecordedUploadContext, "library_partition": media.MediaLibraryPartitionShared}).Error)
	oldKey := fmt.Sprintf("media-derivatives/%d/card/v1/previous.png", parent.ID)
	newKey := fmt.Sprintf("media-derivatives/%d/card/v2/current.png", parent.ID)
	old := media.MediaAssetDerivative{MediaAssetID: parent.ID, Preset: "card", PresetVersion: 1, StorageKey: oldKey, URL: assets.storage.GetURL(oldKey), MimeType: "image/png"}
	next := media.MediaAssetDerivative{MediaAssetID: parent.ID, Preset: "card", PresetVersion: 2, StorageKey: newKey, URL: assets.storage.GetURL(newKey), MimeType: "image/png"}
	require.NoError(t, db.Create(&old).Error)
	writeMediaCleanupOwnershipConflictTestObjects(t, root, parent.StorageKey, oldKey, newKey)
	replacements := []media.MediaAssetDerivative{next}
	require.NoError(t, assets.repo.ReplaceAssetDerivativesAndRecordRetiredObjectsWithinTransaction(replacements, assets.recordRetiredMediaDerivativeCleanupEventsWithinReplacementTransaction))
	next = replacements[0]
	event, err := newObjectStorageCleanupEvent(objectCleanupResourceRetiredMediaDerivative, fmt.Sprint(old.ID), "media_asset", fmt.Sprint(parent.ID), []string{oldKey})
	require.NoError(t, err)
	stored, err := repository.NewOutboxRepository(db).FindEventByKey(event.EventKey)
	require.NoError(t, err)
	return parent, old, next, *stored
}

func TestRetiredMediaDerivativeCleanupRetainsExactHistoricalConsumersAndRetriesOriginalEventAfterLastDetachment(t *testing.T) {
	for _, source := range []string{"product-html", "deleted-product-html", "manual-thumbnail", "variant-snapshot", "deleted-variant-snapshot", "legacy-attachment", "structured-message", "same-parent-derivative", "other-original", "retained-url-only-original", "windows-key-original", "conflict-after-first-page"} {
		t.Run(source, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.Product{}, &product.ProductMedia{}, &ticket.TicketMessage{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			assets.siteURL = "https://shop.example.test"
			parent, old, next, event := createRetiredMediaDerivativeWithDurableCleanupEventForTest(t, db, assets, root)
			var detach func()
			switch source {
			case "product-html", "deleted-product-html":
				owner := product.Product{SKU: "retained", Slug: "retained", Name: "Retained", Status: "active", Description: `<img src="` + old.URL + `?size=old#preview">`}
				require.NoError(t, db.Create(&owner).Error)
				if source == "deleted-product-html" {
					require.NoError(t, db.Delete(&owner).Error)
				}
				detach = func() { require.NoError(t, db.Unscoped().Model(&owner).Update("description", "<p>Detached</p>").Error) }
			case "manual-thumbnail", "variant-snapshot", "deleted-variant-snapshot":
				placement := product.ProductMedia{ProductID: 71, MediaType: "image", MediaAssetID: &parent.ID, URL: parent.URL, IsVisible: true}
				if source == "manual-thumbnail" {
					placement.ThumbnailURL = old.URL
				} else {
					placement.ImageVariantData = product.ProductMediaImageVariantsJSON(map[string]product.ProductMediaImageVariant{"card": {URL: "https://shop.example.test/uploads/" + old.StorageKey + "?old=1"}})
				}
				require.NoError(t, db.Create(&placement).Error)
				if source == "deleted-variant-snapshot" {
					require.NoError(t, db.Delete(&placement).Error)
				}
				detach = func() {
					require.NoError(t, db.Unscoped().Model(&placement).Updates(map[string]any{"thumbnail_url": "", "image_variants": "{}"}).Error)
				}
			case "legacy-attachment", "structured-message":
				message := ticket.TicketMessage{TicketID: 71, MessageType: "product", Attachments: "[]", Metadata: "{}"}
				if source == "legacy-attachment" {
					message.Attachments = fmt.Sprintf(`[%q]`, old.URL)
				} else {
					message.Metadata = structuredProductMetadataForMediaTest(t, old.URL)
				}
				require.NoError(t, db.Create(&message).Error)
				detach = func() {
					require.NoError(t, db.Model(&message).Updates(map[string]any{"attachments": "[]", "metadata": "{}"}).Error)
				}
			case "same-parent-derivative":
				conflict := media.MediaAssetDerivative{MediaAssetID: parent.ID, Preset: "another-preset", PresetVersion: 1, StorageKey: old.StorageKey, URL: old.URL + "?second-registration=1"}
				require.NoError(t, db.Create(&conflict).Error)
				detach = func() { require.NoError(t, db.Unscoped().Delete(&conflict).Error) }
			default:
				if source == "conflict-after-first-page" {
					unrelated := make([]media.MediaAsset, 256)
					for index := range unrelated {
						unrelated[index] = media.MediaAsset{Filename: "unrelated", URL: fmt.Sprintf("https://media.example.test/uploads/unrelated/%d.png", index), StorageKey: fmt.Sprintf("unrelated/%d.png", index)}
					}
					require.NoError(t, db.Create(&unrelated).Error)
				}
				conflict := media.MediaAsset{Filename: "conflict.png", URL: old.URL + "?other=1", StorageKey: old.StorageKey}
				if source == "retained-url-only-original" {
					conflict.StorageKey = ""
				} else if source == "windows-key-original" {
					conflict.StorageKey = strings.ReplaceAll(old.StorageKey, "/", "\\")
				}
				require.NoError(t, db.Create(&conflict).Error)
				if source == "retained-url-only-original" {
					require.NoError(t, db.Delete(&conflict).Error)
				}
				detach = func() { require.NoError(t, db.Unscoped().Delete(&conflict).Error) }
			}
			handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
			require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
			if source != "same-parent-derivative" {
				require.ErrorIs(t, assets.DeleteAsset(t.Context(), parent.ID, MediaAssetDeleteConfirmation(parent.ID)), ErrMediaAssetInUse,
					"whole-parent cleanup must not bypass exact-generation retention")
			}
			require.FileExists(t, filepath.Join(root, filepath.FromSlash(old.StorageKey)))
			detach()
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoError(t, handler.Handle(t.Context(), event), "duplicate delivery is idempotent")
			require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(old.StorageKey)))
			require.FileExists(t, filepath.Join(root, filepath.FromSlash(parent.StorageKey)))
			require.FileExists(t, filepath.Join(root, filepath.FromSlash(next.StorageKey)))
			var tombstone media.MediaAssetDerivative
			require.NoError(t, db.Unscoped().First(&tombstone, old.ID).Error)
			require.True(t, tombstone.DeletedAt.Valid)
			var count int64
			require.NoError(t, db.Model(&outbox.Event{}).Where("event_key = ?", event.EventKey).Count(&count).Error)
			require.EqualValues(t, 1, count)
		})
	}
}

func TestRetiredMediaDerivativeReadRequiresExactPublishedConsumerAndNeverReopensNewBinding(t *testing.T) {
	for _, state := range []string{"published", "draft", "deleted-product", "foreign-host", "other-generation", "private-parent", "inactive-parent", "retired-parent", "ambiguous-owner"} {
		t.Run(state, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.Product{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			parent, old, next, _ := createRetiredMediaDerivativeWithDurableCleanupEventForTest(t, db, assets, root)
			reference := old.URL
			if state == "foreign-host" {
				reference = "https://foreign.example.test/uploads/" + old.StorageKey
			}
			if state == "other-generation" {
				reference = next.URL
			}
			owner := product.Product{SKU: "read", Slug: "read", Name: "Read", Status: "active", Description: `<img src="` + reference + `">`}
			if state == "draft" {
				owner.Status = "draft"
			}
			require.NoError(t, db.Create(&owner).Error)
			switch state {
			case "deleted-product":
				require.NoError(t, db.Delete(&owner).Error)
			case "private-parent":
				require.NoError(t, db.Model(parent).Update("visibility", "private").Error)
			case "inactive-parent":
				require.NoError(t, db.Model(parent).Update("status", "inactive").Error)
			case "retired-parent":
				require.NoError(t, db.Delete(parent).Error)
			case "ambiguous-owner":
				require.NoError(t, db.Create(&media.MediaAsset{Filename: "conflict", StorageKey: old.StorageKey, URL: old.URL + "?another=1"}).Error)
			}
			allowed, err := assets.CanServePublicUpload(old.StorageKey)
			require.NoError(t, err)
			require.Equal(t, state == "published", allowed)
			_, _, err = assets.repo.FindUniquelyRegisteredMediaObjectByExactURLOrTrustedKey(old.URL, old.StorageKey)
			require.Error(t, err, "new binding identity remains strict")
			require.NoError(t, db.Unscoped().Model(&owner).Update("description", "Detached").Error)
			allowed, err = assets.CanServePublicUpload(old.StorageKey)
			require.NoError(t, err)
			require.False(t, allowed)
		})
	}
}

func TestMediaDerivativeReplacementCleanupEventFailureRollsBackBothGenerations(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	root := t.TempDir()
	assets := newMediaDeleteTestService(t, db, root)
	parent := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, assets.storage, "catalog/atomic.png")
	oldKey, nextKey := fmt.Sprintf("media-derivatives/%d/card/v1/old.png", parent.ID), fmt.Sprintf("media-derivatives/%d/card/v2/new.png", parent.ID)
	writeMediaCleanupOwnershipConflictTestObjects(t, root, oldKey, nextKey)
	old := media.MediaAssetDerivative{MediaAssetID: parent.ID, Preset: "card", URL: assets.storage.GetURL(oldKey), StorageKey: oldKey, PresetVersion: 1}
	require.NoError(t, db.Create(&old).Error)
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_retired_derivative_cleanup_event", func(tx *gorm.DB) {
		if tx.Statement.Table == "outbox_events" {
			tx.AddError(errors.New("injected outbox persistence failure"))
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Create().Remove("reject_retired_derivative_cleanup_event") })
	err := assets.repo.ReplaceAssetDerivativesAndRecordRetiredObjectsWithinTransaction([]media.MediaAssetDerivative{{MediaAssetID: parent.ID, Preset: "card", PresetVersion: 2, URL: assets.storage.GetURL(nextKey), StorageKey: nextKey}}, assets.recordRetiredMediaDerivativeCleanupEventsWithinReplacementTransaction)
	require.Error(t, err)
	var generations []media.MediaAssetDerivative
	require.NoError(t, db.Unscoped().Find(&generations).Error)
	require.Len(t, generations, 1)
	require.False(t, generations[0].DeletedAt.Valid)
	require.Equal(t, old.ID, generations[0].ID)
	var events int64
	require.NoError(t, db.Model(&outbox.Event{}).Count(&events).Error)
	require.Zero(t, events)
}

type failedMediaUploadCompensationStorageWithInjectedDeleteError struct {
	storage.StorageService
	deleteError error
	deleteCalls int
}

func (s *failedMediaUploadCompensationStorageWithInjectedDeleteError) Delete(ctx context.Context, key string) error {
	s.deleteCalls++
	if s.deleteError != nil {
		return s.deleteError
	}
	return s.StorageService.Delete(ctx, key)
}

func TestFailedMediaUploadCompensationPersistsBeforeDeleteRetainsClaimsAndRetriesProviderFailure(t *testing.T) {
	for _, scenario := range []string{"unclaimed", "registered-original", "same-parent-derivative", "legacy-business-reference", "provider-error", "outbox-error", "cancelled-request"} {
		t.Run(scenario, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.Product{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			provider := &failedMediaUploadCompensationStorageWithInjectedDeleteError{StorageService: assets.storage}
			assets.storage = provider
			key := "catalog/failed.png"
			var parent *media.MediaAsset
			if scenario == "same-parent-derivative" {
				parent = createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, "catalog/parent.png")
				key = fmt.Sprintf("media-derivatives/%d/card/v1/failed.png", parent.ID)
			}
			writeMediaCleanupOwnershipConflictTestObjects(t, root, key)
			var conflict any
			if scenario == "registered-original" {
				owner := &media.MediaAsset{Filename: "claim", StorageKey: key, URL: provider.GetURL(key)}
				require.NoError(t, db.Create(owner).Error)
				conflict = owner
			} else if scenario == "same-parent-derivative" {
				derivative := &media.MediaAssetDerivative{MediaAssetID: parent.ID, Preset: "card", StorageKey: key, URL: provider.GetURL(key)}
				require.NoError(t, db.Create(derivative).Error)
				conflict = derivative
			} else if scenario == "legacy-business-reference" {
				owner := &product.Product{SKU: "legacy", Slug: "legacy", Name: "Legacy", Description: `<img src="` + provider.GetURL(key) + `">`}
				require.NoError(t, db.Create(owner).Error)
				conflict = owner
			}
			if scenario == "provider-error" {
				provider.deleteError = errors.New("injected provider error")
			}
			if scenario == "outbox-error" {
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_failed_upload_cleanup_event", func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" {
						tx.AddError(errors.New("injected event error"))
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Create().Remove("reject_failed_upload_cleanup_event") })
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if scenario == "cancelled-request" {
				cancel()
			}
			err := assets.persistFailedMediaUploadCompensationBeforeAttemptingExactObjectCleanup(ctx, parent, []string{provider.GetURL(key)})
			if scenario == "outbox-error" {
				require.ErrorIs(t, err, ErrObjectStorageCleanupUnavailable)
				require.Zero(t, provider.deleteCalls)
				require.FileExists(t, filepath.Join(root, filepath.FromSlash(key)))
				return
			}
			require.NoError(t, err)
			var event outbox.Event
			require.NoError(t, db.First(&event).Error)
			handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
			if conflict != nil {
				require.Zero(t, provider.deleteCalls)
				require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
				require.FileExists(t, filepath.Join(root, filepath.FromSlash(key)))
				require.NoError(t, db.Unscoped().Delete(conflict).Error)
			}
			if scenario == "provider-error" {
				require.FileExists(t, filepath.Join(root, filepath.FromSlash(key)))
				provider.deleteError = nil
			}
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(key)))
		})
	}
}

func TestFailedMediaUploadCompensationCannotDeleteDedicatedOrForeignDerivativeNamespaces(t *testing.T) {
	for _, key := range []string{"site-logo/same.png", "site-favicon/same.png", "visual-showcase/same.png", "media-derivatives/999/card/v1/same.png", "warranty/same.png", "order-evidence/same.png", "after-sales/same.png"} {
		t.Run(key, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			provider := &failedMediaUploadCompensationStorageWithInjectedDeleteError{StorageService: assets.storage}
			assets.storage = provider
			parent := &media.MediaAsset{ID: 1, Visibility: "public"}
			require.Error(t, assets.persistFailedMediaUploadCompensationBeforeAttemptingExactObjectCleanup(t.Context(), nil, []string{provider.GetURL(key)}))
			require.Error(t, assets.persistFailedMediaUploadCompensationBeforeAttemptingExactObjectCleanup(t.Context(), parent, []string{provider.GetURL(key)}))
			require.Zero(t, provider.deleteCalls)
		})
	}
}

func TestRetiredPrivateStructuredMessageDerivativeKeepsAuthorizedMembershipAndAnonymousDenial(t *testing.T) {
	env := createCustomerServicePrivateAttachmentAccessTestEnvironment(t)
	old := env.asset.Derivatives[1]
	require.NoError(t, env.db.Delete(&old).Error)
	old.ID, old.DeletedAt = 0, gorm.DeletedAt{}
	old.PresetVersion++
	old.StorageKey = strings.Replace(old.StorageKey, "/v1/", "/v2/", 1)
	old.URL = env.storage.GetURL(old.StorageKey)
	writeMediaCleanupOwnershipConflictTestObjects(t, env.privateRoot, old.StorageKey)
	require.NoError(t, env.db.Create(&old).Error)
	metadata := structuredProductMetadataForMediaTest(t, env.asset.Derivatives[1].URL)
	message := ticket.TicketMessage{TicketID: env.conversation.ID, Content: "Historical card", MessageType: "product", Metadata: metadata, Attachments: "[]"}
	require.NoError(t, env.db.Create(&message).Error)
	file, err := env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, env.owner, message.ID, 0, "")
	require.NoError(t, err)
	readPayload, err := io.ReadAll(file.ReadCloser)
	require.NoError(t, err)
	require.NoError(t, file.ReadCloser.Close())
	oldPayload, err := os.ReadFile(filepath.Join(env.privateRoot, filepath.FromSlash(env.asset.Derivatives[1].StorageKey)))
	require.NoError(t, err)
	require.Equal(t, oldPayload, readPayload, "authorized membership must stream the exact historical generation")
	_, err = env.access.OpenPublicCustomerServiceMessageAttachment(t.Context(), *env.conversation.ConversationID, CustomerServiceOwner{VisitorSessionHash: strings.Repeat("b", 64)}, message.ID, 0, "")
	require.Error(t, err)
	allowed, err := env.mediaService.CanServePublicUpload(env.asset.Derivatives[1].StorageKey)
	require.NoError(t, err)
	require.False(t, allowed)
	_, err = os.Stat(filepath.Join(env.privateRoot, filepath.FromSlash(env.asset.Derivatives[1].StorageKey)))
	require.NoError(t, err)
}

func TestFailedDerivativeCleanupAfterParentRemovalUsesOnlyItsRecordedNamespace(t *testing.T) {
	for _, namespace := range []string{mediaDerivativePrefix, storage.CustomerServiceConversationAttachmentDerivativesPrefix} {
		t.Run(namespace, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
			root := t.TempDir()
			assets := newMediaDeleteTestService(t, db, root)
			key := namespace + "/77/card/v1/failed.png"
			objectRoot := root
			if namespace == storage.CustomerServiceConversationAttachmentDerivativesPrefix {
				objectRoot = root + "-private"
			}
			t.Cleanup(func() { _ = os.RemoveAll(objectRoot) })
			writeMediaCleanupOwnershipConflictTestObjects(t, objectRoot, key)
			event, err := newObjectStorageCleanupEvent(objectCleanupResourceUnregisteredMediaDerivative, "77", "media_asset", "77", []string{key})
			require.NoError(t, err)
			handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
			require.NoError(t, handler.Handle(t.Context(), *event))
			require.NoFileExists(t, filepath.Join(objectRoot, filepath.FromSlash(key)))
			event, err = newObjectStorageCleanupEvent(objectCleanupResourceUnregisteredMediaDerivative, "78", "media_asset", "78", []string{key})
			require.NoError(t, err)
			require.Error(t, handler.Handle(t.Context(), *event), "a removed parent never confers another parent's namespace")
		})
	}
}

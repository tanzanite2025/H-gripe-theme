package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createRetiredUnregisteredCustomerServiceOriginalAndDurableCleanupEventForTest(t *testing.T, env customerServicePrivateAttachmentAccessTestEnvironment) (string, outbox.Event) {
	t.Helper()
	key, err := storage.AllocatePrivateObjectKeyForPendingUpload("abandoned-original.png", storage.CustomerServiceConversationAttachmentOriginalsPrefix)
	require.NoError(t, err)
	writeMediaCleanupOwnershipConflictTestObjects(t, env.privateRoot, key)
	session := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
		BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
		ObjectKey: key, Status: "failed", CreatedAt: time.Now().Add(-48 * time.Hour), ExpiresAt: time.Now().Add(-24 * time.Hour)}
	require.NoError(t, env.mediaService.repo.CreateCustomerServicePendingUploadSession(t.Context(), &session))
	retention := configureCustomerServiceManagedAttachmentRetirementTestService(t, env)
	retired, err := retention.retireExpiredUnboundCustomerServiceUploadSession(t.Context(), session, time.Now().UTC())
	require.NoError(t, err)
	require.True(t, retired)
	expected, err := newObjectStorageCleanupEvent(objectCleanupResourceCustomerServiceUnregisteredUpload, session.ID,
		outbox.AggregateTypeCustomerServiceConversation, fmt.Sprint(env.conversation.ID), []string{key})
	require.NoError(t, err)
	event, err := repository.NewOutboxRepository(env.db).FindEventByKey(expected.EventKey)
	require.NoError(t, err)
	return key, *event
}

func TestCustomerServiceUnregisteredOriginalCleanupRetainsExactHistoricalClaimsAndRetriesTheSameEvent(t *testing.T) {
	for _, source := range []string{"original", "deleted-url-only-original", "windows-key-original", "derivative", "retired-url-only-derivative",
		"product-html", "deleted-product-html", "private-origin-product-html", "product-variant-snapshot", "legacy-escaped-attachment", "structured-message",
		"uploading-receipt", "failed-receipt", "ready-receipt", "bound-receipt", "windows-receipt-after-first-page"} {
		t.Run(source, func(t *testing.T) {
			env, _, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			require.NoError(t, env.db.AutoMigrate(&product.Product{}, &product.ProductMedia{}))
			key, event := createRetiredUnregisteredCustomerServiceOriginalAndDurableCleanupEventForTest(t, env)
			reference := env.mediaService.storage.GetURL(key)
			var detach func()
			switch source {
			case "original", "deleted-url-only-original", "windows-key-original":
				claim := media.MediaAsset{Filename: "retained.png", URL: reference + "?retained=1", StorageKey: key, Visibility: "private", Status: "active"}
				if source == "deleted-url-only-original" {
					claim.StorageKey = ""
				} else if source == "windows-key-original" {
					claim.StorageKey = strings.ReplaceAll(key, "/", "\\")
				}
				require.NoError(t, env.db.Create(&claim).Error)
				if source == "deleted-url-only-original" {
					require.NoError(t, env.db.Delete(&claim).Error)
				}
				detach = func() { require.NoError(t, env.db.Unscoped().Delete(&claim).Error) }
			case "derivative", "retired-url-only-derivative":
				claim := media.MediaAssetDerivative{MediaAssetID: env.asset.ID, Preset: "retained-original", PresetVersion: 1, StorageKey: key, URL: reference}
				if source == "retired-url-only-derivative" {
					claim.StorageKey = ""
				}
				require.NoError(t, env.db.Create(&claim).Error)
				if source == "retired-url-only-derivative" {
					require.NoError(t, env.db.Delete(&claim).Error)
				}
				detach = func() { require.NoError(t, env.db.Unscoped().Delete(&claim).Error) }
			case "product-html", "deleted-product-html", "private-origin-product-html":
				if source == "private-origin-product-html" {
					reference = "https://private.example.test/" + key
					_, err := env.mediaService.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
					require.Error(t, err, "retaining a private consumer must not authorize public reuse")
				}
				item := product.Product{SKU: "retained-private-reference", Slug: "retained-private-reference", Name: "Historical reference",
					Description: `<img src="` + reference + `?historical=1#preview">`}
				require.NoError(t, env.db.Create(&item).Error)
				if source == "deleted-product-html" {
					require.NoError(t, env.db.Delete(&item).Error)
				}
				detach = func() { require.NoError(t, env.db.Unscoped().Model(&item).Update("description", "").Error) }
			case "product-variant-snapshot":
				placement := product.ProductMedia{ProductID: 771, MediaType: "image", URL: "https://external.example.test/independent.png",
					ImageVariantData: product.ProductMediaImageVariantsJSON(map[string]product.ProductMediaImageVariant{"card": {URL: reference}})}
				require.NoError(t, env.db.Create(&placement).Error)
				require.NoError(t, env.db.Delete(&placement).Error)
				detach = func() { require.NoError(t, env.db.Unscoped().Model(&placement).Update("image_variants", "{}").Error) }
			case "legacy-escaped-attachment", "structured-message":
				message := ticket.TicketMessage{TicketID: env.conversation.ID, MessageType: "product", Attachments: "[]", Metadata: "{}"}
				if source == "legacy-escaped-attachment" {
					encoded, err := json.Marshal([]string{reference + "?retained=1"})
					require.NoError(t, err)
					message.Attachments = strings.ReplaceAll(string(encoded), "/", `\/`)
				} else {
					message.Metadata = structuredProductMetadataForMediaTest(t, reference)
				}
				require.NoError(t, env.db.Create(&message).Error)
				detach = func() {
					require.NoError(t, env.db.Model(&message).Updates(map[string]any{"attachments": "[]", "metadata": "{}"}).Error)
				}
			default:
				claim := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
					BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
					ObjectKey: key, Status: strings.TrimSuffix(source, "-receipt"), ExpiresAt: time.Now().Add(-time.Hour)}
				if source == "windows-receipt-after-first-page" {
					unrelated := make([]media.MediaUploadSession, 256)
					for index := range unrelated {
						unrelated[index] = claim
						unrelated[index].ID = fmt.Sprintf("00000000-0000-0000-0000-%012d", index)
						unrelated[index].ObjectKey = fmt.Sprintf("%s/unrelated/%d.png", storage.CustomerServiceConversationAttachmentOriginalsPrefix, index)
					}
					require.NoError(t, env.db.Create(&unrelated).Error)
					claim.ID, claim.Status, claim.ObjectKey = "ffffffff-ffff-ffff-ffff-ffffffffffff", "failed", strings.ReplaceAll(key, "/", "\\")
				}
				// Historical rows can predate canonical-key validation.
				require.NoError(t, env.db.Create(&claim).Error)
				detach = func() { require.NoError(t, env.db.Delete(&claim).Error) }
			}
			handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
			require.ErrorIs(t, handler.Handle(t.Context(), event), ErrMediaAssetInUse)
			require.FileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
			stored, err := repository.NewOutboxRepository(env.db).FindEventByKey(event.EventKey)
			require.NoError(t, err)
			require.Equal(t, event.ID, stored.ID)
			require.Equal(t, event.Payload, stored.Payload)
			detach()
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoFileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
			require.FileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(env.asset.StorageKey)))
		})
	}
}

func TestCustomerServiceUploadReceiptAdmissionPrecedesBytesAndRejectsForeignOrNoncanonicalKeys(t *testing.T) {
	env, _, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	key, err := storage.AllocatePrivateObjectKeyForPendingUpload("before-bytes.png", storage.CustomerServiceConversationAttachmentOriginalsPrefix)
	require.NoError(t, err)
	session := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
		BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
		ObjectKey: key, Status: "uploading", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, env.mediaService.repo.CreateCustomerServicePendingUploadSession(t.Context(), &session))
	require.NoFileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
	event, err := newObjectStorageCleanupEvent(objectCleanupResourceCustomerServiceUnregisteredUpload, uuid.NewString(), "test", "test", []string{key})
	require.NoError(t, err)
	require.ErrorIs(t, NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil).Handle(t.Context(), *event), ErrMediaAssetInUse)
	for _, invalid := range []string{"catalog/public.png", "site-logos/dedicated.png", storage.CustomerServiceConversationAttachmentDerivativesPrefix + "/1/card.png",
		strings.ReplaceAll(key, "/", "\\"), "../" + key, "/" + key} {
		session.ID, session.ObjectKey = uuid.NewString(), invalid
		require.ErrorIs(t, env.mediaService.repo.CreateCustomerServicePendingUploadSession(t.Context(), &session), repository.ErrCustomerServiceAttachmentBindingUnavailable)
		require.Error(t, env.mediaService.cleanupRetiredUnregisteredCustomerServiceUploadOriginal(t.Context(), []string{invalid}))
	}
	require.ErrorIs(t, env.mediaService.repo.CreateCustomerServicePendingUploadSession(t.Context(), nil), repository.ErrCustomerServiceAttachmentBindingUnavailable)
	require.Error(t, env.mediaService.cleanupRetiredUnregisteredCustomerServiceUploadOriginal(t.Context(), []string{key, key}))
}

func TestCustomerServiceUnregisteredOriginalCleanupIgnoresForeignHostsNearNamesAndIndependentReceiptPaths(t *testing.T) {
	env, _, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
	key, event := createRetiredUnregisteredCustomerServiceOriginalAndDurableCleanupEventForTest(t, env)
	neighborKey := storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/another-owner/" + filepath.Base(key)
	writeMediaCleanupOwnershipConflictTestObjects(t, env.privateRoot, neighborKey)
	session := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
		BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
		ObjectKey: neighborKey, Status: "uploading", ExpiresAt: time.Now().Add(time.Hour)}
	require.NoError(t, env.mediaService.repo.CreateCustomerServicePendingUploadSession(t.Context(), &session))
	message := ticket.TicketMessage{TicketID: env.conversation.ID, MessageType: "image", Metadata: "{}",
		Attachments: fmt.Sprintf(`[%q,%q]`, "https://foreign.example.test/uploads/"+key, env.mediaService.storage.GetURL(key)+".backup")}
	require.NoError(t, env.db.Create(&message).Error)
	require.NoError(t, NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil).Handle(t.Context(), event))
	require.NoFileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
	require.FileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(neighborKey)))
}

func TestCustomerServiceUnregisteredOriginalCleanupRetainsOriginalEventOnScanProviderFailureOrCanceledRequest(t *testing.T) {
	for _, failure := range []string{"reference-scan", "provider-delete", "canceled-request"} {
		t.Run(failure, func(t *testing.T) {
			env, _, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
			key, event := createRetiredUnregisteredCustomerServiceOriginalAndDurableCleanupEventForTest(t, env)
			counter := &mediaObjectPhysicalCleanupStorageWithAtomicDeletionCount{StorageService: env.mediaService.storage}
			provider := &customerServiceAttachmentCleanupStorageWithInjectedDeleteFailure{StorageService: counter, fail: failure == "provider-delete"}
			env.mediaService.storage = provider
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			injected := errors.New("injected receipt scan failure")
			callbackName := "reject_exact_customer_service_receipt_scan_" + uuid.NewString()
			if failure == "reference-scan" {
				require.NoError(t, env.db.Callback().Query().Before("gorm:query").Register(callbackName, func(tx *gorm.DB) {
					if tx.Statement.Table == "media_upload_sessions" {
						tx.AddError(injected)
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Query().Remove(callbackName) })
			} else if failure == "canceled-request" {
				cancel()
			}
			handler := NewObjectStorageCleanupOutboxHandler(env.mediaService, nil, nil, nil)
			err := handler.Handle(ctx, event)
			require.Error(t, err)
			if failure == "reference-scan" {
				require.ErrorIs(t, err, injected)
				require.NoError(t, env.db.Callback().Query().Remove(callbackName))
			}
			require.Zero(t, counter.deleteCalls.Load(), "failed validation must never reach physical deletion")
			require.FileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
			stored, lookupError := repository.NewOutboxRepository(env.db).FindEventByKey(event.EventKey)
			require.NoError(t, lookupError)
			require.Equal(t, event.Payload, stored.Payload)
			provider.fail = false
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoError(t, handler.Handle(t.Context(), event))
			require.NoFileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
		})
	}
}

func TestCustomerServiceUploadReceiptExactObjectClaimAlsoProtectsWholeAssetRetirementAndCleanup(t *testing.T) {
	for _, originalIdentity := range []string{"storage-key", "url-only", "windows-receipt"} {
		for _, claimTime := range []string{"before-retirement", "after-retirement"} {
			t.Run(originalIdentity+"/"+claimTime, func(t *testing.T) {
				env, _, _ := createCustomerServicePendingPrivateUploadBindingTestEnvironment(t)
				key := storage.CustomerServiceConversationAttachmentOriginalsPrefix + "/whole-asset-receipt-hold.png"
				target := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, env.db, env.mediaService.storage, key)
				changes := map[string]any{"visibility": "private"}
				if originalIdentity == "url-only" {
					changes["storage_key"] = ""
				}
				require.NoError(t, env.db.Model(target).Updates(changes).Error)
				writeMediaCleanupOwnershipConflictTestObjects(t, env.privateRoot, key)
				retire := func() error {
					return env.mediaService.repo.WithinMediaAssetTransactionWithRequestContext(t.Context(), func(repo *repository.MediaRepository, tx *gorm.DB) error {
						current, err := repo.FindAssetByIDUnscopedForUpdate(target.ID)
						if err != nil {
							return err
						}
						return env.mediaService.retireUnreferencedMediaAssetWithinTransaction(repo, tx, current)
					})
				}
				if claimTime == "after-retirement" {
					require.NoError(t, retire())
				}
				receipt := media.MediaUploadSession{ID: uuid.NewString(), BusinessContextType: media.MediaUploadSessionCustomerServiceContext,
					BusinessContextID: env.conversation.ID, UploadActorType: media.MediaUploadActorVisitor, UploadActorVisitorSessionHash: env.owner.VisitorSessionHash,
					ObjectKey: key, Status: "uploading", ExpiresAt: time.Now().Add(time.Hour)}
				if originalIdentity == "windows-receipt" {
					receipt.ObjectKey = strings.ReplaceAll(key, "/", "\\")
				}
				require.NoError(t, env.db.Create(&receipt).Error)
				if claimTime == "before-retirement" {
					require.ErrorIs(t, retire(), ErrMediaAssetInUse)
				} else {
					require.ErrorIs(t, env.mediaService.cleanupDeletedAsset(t.Context(), target.ID), ErrMediaAssetInUse)
				}
				require.FileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
				require.NoError(t, env.db.Delete(&receipt).Error)
				if claimTime == "before-retirement" {
					require.NoError(t, retire())
				}
				require.NoError(t, env.mediaService.cleanupDeletedAsset(t.Context(), target.ID))
				require.NoError(t, env.mediaService.cleanupDeletedAsset(t.Context(), target.ID))
				require.NoFileExists(t, filepath.Join(env.privateRoot, filepath.FromSlash(key)))
			})
		}
	}
}

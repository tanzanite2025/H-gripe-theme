package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion struct {
	storage.StorageService
	uploadReference    string
	uploadError        error
	beforeUploadReturn func()
	beforeDelete       func()
	deleteError        error
	deletedKeys        []string
}

func (s *dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion) UploadWithPrefix(context.Context, *multipart.FileHeader, string) (string, error) {
	if s.beforeUploadReturn != nil {
		s.beforeUploadReturn()
	}
	return s.uploadReference, s.uploadError
}

func (s *dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion) Delete(ctx context.Context, key string) error {
	if s.beforeDelete != nil {
		s.beforeDelete()
	}
	s.deletedKeys = append(s.deletedKeys, key)
	if s.deleteError != nil {
		return s.deleteError
	}
	return s.StorageService.Delete(ctx, key)
}

func (s *dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion) WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key, cacheControl string) (string, error) {
	return s.StorageService.(storage.PreallocatedDedicatedImmediateOwnerObjectWriter).WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(ctx, reader, key, cacheControl)
}

type dedicatedPublicUploadFailureCompensationRegressionEnvironment struct {
	db        *gorm.DB
	provider  *dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion
	logos     *SiteLogoService
	favicons  *SiteFaviconService
	visuals   *HomeVisualTileService
	handler   *ObjectStorageCleanupOutboxHandler
	key       string
	namespace string
}

func createDedicatedPublicUploadFailureCompensationRegressionEnvironment(t *testing.T, namespace string) dedicatedPublicUploadFailureCompensationRegressionEnvironment {
	t.Helper()
	db, underlying, _, oldEvent, key := createDedicatedPublicObjectRetainedHistoryCleanupTestEnvironment(t, namespace)
	require.NoError(t, db.Delete(&oldEvent).Error)
	provider := &dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion{StorageService: underlying, uploadReference: underlying.GetURL(key)}
	logos := NewSiteLogoService(repository.NewSiteLogoRepository(db), provider, "https://shop.example.test")
	favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(db), provider, "https://shop.example.test")
	visuals := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), provider, "https://shop.example.test")
	settings := NewSettingService(repository.NewSettingRepository(db), nil, 60)
	logos.ConfigureSettingService(settings)
	favicons.ConfigureSettingService(settings)
	logos.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	favicons.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	visuals.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	handler := NewObjectStorageCleanupOutboxHandler(nil, logos, visuals, nil)
	handler.ConfigureSiteFaviconService(favicons)
	return dedicatedPublicUploadFailureCompensationRegressionEnvironment{db, provider, logos, favicons, visuals, handler, key, namespace}
}

func (env dedicatedPublicUploadFailureCompensationRegressionEnvironment) uploadThroughDedicatedService(ctx context.Context, t *testing.T) error {
	t.Helper()
	file := createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "upload.webp")
	switch env.namespace {
	case SiteLogoStoragePrefix:
		_, err := env.logos.UploadCurrent(ctx, file, 11)
		return err
	case SiteFaviconStoragePrefix:
		_, err := env.favicons.UploadCurrent(ctx, file, 11)
		return err
	default:
		_, err := env.visuals.UploadAdminImage(ctx, "home-hero", "en", multipartFileHeader(t, "upload.png", "image/png", testOpaquePNG(t, 600, 600)))
		return err
	}
}

func (env dedicatedPublicUploadFailureCompensationRegressionEnvironment) seedCurrentDedicatedOwnerAndReturnDetachment(t *testing.T) func() {
	t.Helper()
	switch env.namespace {
	case SiteLogoStoragePrefix:
		asset := sitelogodomain.Asset{ID: sitelogodomain.CurrentAssetID, StorageKey: env.key, URL: env.provider.uploadReference}
		require.NoError(t, env.db.Create(&asset).Error)
		return func() { require.NoError(t, env.db.Delete(&asset).Error) }
	case SiteFaviconStoragePrefix:
		asset := sitefavicondomain.Asset{ID: sitefavicondomain.CurrentAssetID, StorageKey: env.key, URL: env.provider.uploadReference}
		require.NoError(t, env.db.Create(&asset).Error)
		return func() { require.NoError(t, env.db.Delete(&asset).Error) }
	default:
		item := homevisualtile.Tile{TileSetKey: "home-hero", Locale: "en", StorageKey: env.key, ImageURL: env.provider.uploadReference, DesktopOrder: 1}
		require.NoError(t, env.db.Create(&item).Error)
		return func() { require.NoError(t, env.db.Delete(&item).Error) }
	}
}

func (env dedicatedPublicUploadFailureCompensationRegressionEnvironment) attemptLegacyReturnedURLCompensationForRegression(ctx context.Context) error {
	if env.provider.beforeUploadReturn != nil {
		env.provider.beforeUploadReturn()
	}
	var err error
	switch env.namespace {
	case SiteLogoStoragePrefix:
		err = env.logos.persistFailedUploadCompensationWithOwnerAndHistoricalReferenceProtection(ctx, env.provider.uploadReference)
	case SiteFaviconStoragePrefix:
		err = env.favicons.persistFailedUploadCompensationWithOwnerAndHistoricalReferenceProtection(ctx, env.provider.uploadReference)
	default:
		err = env.visuals.persistFailedUploadCompensationWithOwnerAndHistoricalReferenceProtection(ctx, env.provider.uploadReference)
	}
	return errors.Join(env.provider.uploadError, err)
}

func TestLegacyDedicatedPublicFailedUploadPersistsBeforeDeletionAndRetainsOwnersOrHistoryUntilOriginalTaskRetry(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, HomeVisualTileStoragePrefix} {
		for _, retainedBy := range []string{"none", "current owner", "historical post"} {
			t.Run(namespace+"/"+retainedBy, func(t *testing.T) {
				env := createDedicatedPublicUploadFailureCompensationRegressionEnvironment(t, namespace)
				uploadFailure := errors.New("upload reported a failure after returning an object address")
				env.provider.uploadError = uploadFailure
				detach := func() {}
				switch retainedBy {
				case "current owner":
					detach = env.seedCurrentDedicatedOwnerAndReturnDetachment(t)
				case "historical post":
					detach = insertRetainedHistoricalDedicatedObjectConsumerAndReturnDetachment(t, env.db, env.key, "post_content")
				}
				// Querying the SQLite DB from inside its one-connection cleanup
				// transaction would deadlock. The create callback proves ordering.
				persisted := false
				callback := "observe_persisted_compensation_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "outbox_events" && tx.Error == nil {
						persisted = true
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
				env.provider.beforeDelete = func() { require.True(t, persisted) }
				require.ErrorIs(t, env.attemptLegacyReturnedURLCompensationForRegression(t.Context()), uploadFailure)
				var events []outbox.Event
				require.NoError(t, env.db.Find(&events).Error)
				require.Len(t, events, 1)
				if retainedBy != "none" {
					require.Empty(t, env.provider.deletedKeys)
					requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
				}
				worker := NewOutboxService(repository.NewOutboxRepository(env.db))
				worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, env.handler.Handle)
				if retainedBy == "historical post" {
					_, err := worker.ProcessPending(t.Context(), time.Now().Add(time.Minute), 10)
					require.NoError(t, err)
					var retained outbox.Event
					require.NoError(t, env.db.First(&retained, events[0].ID).Error)
					require.Equal(t, outbox.EventStatusFailed, retained.Status)
					require.Equal(t, events[0].EventKey, retained.EventKey)
				}
				detach()
				require.NoError(t, env.handler.Handle(t.Context(), events[0]))
				requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
			})
		}
	}
}

func requireDedicatedPublicUploadCompensationObjectMissing(t *testing.T, provider storage.ObjectOpener, key string) {
	t.Helper()
	object, err := provider.Open(t.Context(), key)
	if object != nil && object.ReadCloser != nil {
		_ = object.ReadCloser.Close()
	}
	require.Error(t, err)
	require.Nil(t, object)
}

func TestLegacyDedicatedPublicFailedUploadRejectsUntrustedOrOtherNamespaceIdentityWithoutDeletionOrTask(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, HomeVisualTileStoragePrefix} {
		for _, invalidIdentity := range []string{"foreign origin", "another namespace", "namespace without object", "empty address"} {
			t.Run(namespace+"/"+invalidIdentity, func(t *testing.T) {
				env := createDedicatedPublicUploadFailureCompensationRegressionEnvironment(t, namespace)
				switch invalidIdentity {
				case "foreign origin":
					env.provider.uploadReference = "https://foreign.example.test/uploads/" + env.key
				case "another namespace":
					env.provider.uploadReference = env.provider.GetURL("catalog/retained image.png")
				case "namespace without object":
					env.provider.uploadReference = env.provider.GetURL(namespace)
				case "empty address":
					env.provider.uploadReference = ""
				}
				require.Error(t, env.attemptLegacyReturnedURLCompensationForRegression(t.Context()))
				require.Empty(t, env.provider.deletedKeys)
				var count int64
				require.NoError(t, env.db.Model(&outbox.Event{}).Count(&count).Error)
				require.Zero(t, count)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
			})
		}
	}
}

func TestLegacyDedicatedPublicFailedUploadCompensationSurvivesCancellationProviderFailureAndRepeatedKeyButNeverDeletesAfterPersistenceFailure(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, HomeVisualTileStoragePrefix} {
		for _, failureMode := range []string{"request cancellation", "provider delete failure", "outbox persistence failure", "same key after completed compensation"} {
			t.Run(namespace+"/"+failureMode, func(t *testing.T) {
				env := createDedicatedPublicUploadFailureCompensationRegressionEnvironment(t, namespace)
				uploadFailure := errors.New("upload failed")
				env.provider.uploadError = uploadFailure
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				if failureMode == "request cancellation" {
					env.provider.beforeUploadReturn = cancel
				}
				if failureMode == "provider delete failure" {
					env.provider.deleteError = errors.New("object store unavailable")
				}
				if failureMode == "outbox persistence failure" {
					callback := "reject_durable_compensation_" + uuid.NewString()
					require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "outbox_events" {
							tx.AddError(errors.New("outbox unavailable"))
						}
					}))
					t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
				}
				if failureMode == "same key after completed compensation" {
					detach := env.seedCurrentDedicatedOwnerAndReturnDetachment(t)
					require.ErrorIs(t, env.attemptLegacyReturnedURLCompensationForRegression(ctx), uploadFailure)
					worker := NewOutboxService(repository.NewOutboxRepository(env.db))
					worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, env.handler.Handle)
					_, err := worker.ProcessPending(t.Context(), time.Now().Add(time.Minute), 10)
					require.NoError(t, err)
					detach()
					env.provider.deleteError = errors.New("second attempt needs retry")
				}
				err := env.attemptLegacyReturnedURLCompensationForRegression(ctx)
				require.ErrorIs(t, err, uploadFailure)
				var events []outbox.Event
				require.NoError(t, env.db.Order("id ASC").Find(&events).Error)
				if failureMode == "outbox persistence failure" {
					require.ErrorIs(t, err, ErrObjectStorageCleanupUnavailable)
					require.Empty(t, events)
					require.Empty(t, env.provider.deletedKeys)
					requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
					return
				}
				if failureMode == "same key after completed compensation" {
					require.Len(t, events, 2)
					require.NotEqual(t, events[0].EventKey, events[1].EventKey)
					require.Equal(t, outbox.EventStatusProcessed, events[0].Status)
				} else {
					require.Len(t, events, 1)
				}
				require.Equal(t, outbox.EventStatusPending, events[len(events)-1].Status)
				if failureMode == "request cancellation" {
					require.Len(t, env.provider.deletedKeys, 1)
					requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
				} else {
					requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
				}
				env.provider.deleteError = nil
				require.NoError(t, env.handler.Handle(t.Context(), events[len(events)-1]))
				requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
			})
		}
	}
}

func TestDedicatedSiteIdentityOwnerRegistrationRejectsMissingBytesInconsistentKeysAndMissingValidator(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix} {
		for _, invalidRegistration := range []string{"missing bytes", "inconsistent key", "missing validator"} {
			t.Run(namespace+"/"+invalidRegistration, func(t *testing.T) {
				env := createDedicatedPublicUploadFailureCompensationRegressionEnvironment(t, namespace)
				key := env.key
				if invalidRegistration == "missing bytes" {
					require.NoError(t, env.provider.StorageService.Delete(t.Context(), key))
				}
				if invalidRegistration == "inconsistent key" {
					key = namespace + "/different.png"
				}
				var err error
				if namespace == SiteLogoStoragePrefix {
					repo := env.logos.repo
					if invalidRegistration == "missing validator" {
						repo = repository.NewSiteLogoRepository(env.db)
					}
					_, err = repo.WithContext(t.Context()).ReplaceCurrent(&sitelogodomain.Asset{StorageKey: key, URL: env.provider.uploadReference}, nil)
				} else {
					repo := env.favicons.repo
					if invalidRegistration == "missing validator" {
						repo = repository.NewSiteFaviconRepository(env.db)
					}
					err = repo.WithContext(t.Context()).ReplaceCurrent(&sitefavicondomain.Asset{StorageKey: key, URL: env.provider.uploadReference}, nil)
				}
				require.Error(t, err)
				if invalidRegistration == "missing validator" {
					require.ErrorIs(t, err, repository.ErrDedicatedSiteIdentityTransactionalObjectValidationUnavailable)
				}
				var count int64
				require.NoError(t, env.db.Table(map[string]string{SiteLogoStoragePrefix: "site_logo_assets", SiteFaviconStoragePrefix: "site_favicon_assets"}[namespace]).Count(&count).Error)
				require.Zero(t, count)
			})
		}
	}
}

func TestDedicatedSiteIdentityUploadOwnerMirrorFailureLeavesPreallocatedReceiptAndProtectsPreviousObject(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix} {
		for _, reused := range []bool{false, true} {
			t.Run(namespace+"/"+map[bool]string{false: "new object", true: "same key current owner"}[reused], func(t *testing.T) {
				env := createDedicatedPublicUploadFailureCompensationRegressionEnvironment(t, namespace)
				if reused {
					env.seedCurrentDedicatedOwnerAndReturnDetachment(t)
				}
				mirrorFailure := errors.New("mirror persistence failed")
				callback := "reject_identity_setting_mirror_" + uuid.NewString()
				require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "settings" {
						tx.AddError(mirrorFailure)
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
				require.ErrorIs(t, env.uploadThroughDedicatedService(t.Context(), t), mirrorFailure)
				var events []outbox.Event
				require.NoError(t, env.db.Find(&events).Error)
				require.Len(t, events, 1)
				require.Empty(t, env.provider.deletedKeys)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
				var payload outbox.ObjectStorageCleanupPayload
				require.NoError(t, json.Unmarshal(events[0].Payload, &payload))
				require.Len(t, payload.ObjectKeys, 1)
				require.NotEqual(t, env.key, payload.ObjectKeys[0], "new preallocation cannot reuse the old owner's key")
				require.NoError(t, env.handler.Handle(t.Context(), events[0]))
				requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, payload.ObjectKeys[0])
				var settings int64
				require.NoError(t, env.db.Model(&setting.Setting{}).Count(&settings).Error)
				require.Zero(t, settings)
			})
		}
	}
}

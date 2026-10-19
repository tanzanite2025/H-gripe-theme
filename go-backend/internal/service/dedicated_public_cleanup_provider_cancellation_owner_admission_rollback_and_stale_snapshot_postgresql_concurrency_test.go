package service

import (
	"context"
	"database/sql"
	"errors"
	"sync"
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

func createIsolatedPostgreSQLDedicatedPublicCleanupAndOwnerAdmissionTestEnvironment(t *testing.T, namespace string) (dedicatedPublicUploadFailureCompensationRegressionEnvironment, outbox.Event) {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, base.db.AutoMigrate(&sitelogodomain.Asset{}, &sitefavicondomain.Asset{}, &homevisualtile.Tile{}, &setting.Setting{}))
	key := namespace + "/pending/owner-admission.webp"
	writeMediaCleanupOwnershipConflictTestObjects(t, base.publicRoot, key)
	provider := &dedicatedPublicUploadCompensationStorageWithControlledResultAndDeletion{StorageService: base.storage.StorageService, uploadReference: base.storage.GetURL(key)}
	logos := NewSiteLogoService(repository.NewSiteLogoRepository(base.db), provider, "https://shop.example.test")
	favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(base.db), provider, "https://shop.example.test")
	visuals := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(base.db), provider, "https://shop.example.test")
	handler := NewObjectStorageCleanupOutboxHandler(nil, logos, visuals, nil)
	handler.ConfigureSiteFaviconService(favicons)
	resource := map[string]string{SiteLogoStoragePrefix: objectCleanupResourceSiteLogo, SiteFaviconStoragePrefix: objectCleanupResourceSiteFavicon, HomeVisualTileStoragePrefix: objectCleanupResourceHomeVisualTile}[namespace]
	event, err := newObjectStorageCleanupEvent(resource, uuid.NewString(), resource, "owner-admission", []string{key})
	require.NoError(t, err)
	require.NoError(t, repository.NewOutboxRepository(base.db).CreateEvent(event))
	return dedicatedPublicUploadFailureCompensationRegressionEnvironment{base.db, provider, logos, favicons, visuals, handler, key, namespace}, *event
}

func replaceDedicatedPublicOwnerAfterNamespaceLockWithOptionalTransactionalCallback(ctx context.Context, env dedicatedPublicUploadFailureCompensationRegressionEnvironment, db *gorm.DB, afterWrite func(*gorm.DB) error) error {
	callback := func(tx *gorm.DB) error {
		if afterWrite != nil {
			return afterWrite(tx)
		}
		return nil
	}
	switch env.namespace {
	case SiteLogoStoragePrefix:
		owner := NewSiteLogoService(repository.NewSiteLogoRepository(db), env.provider, "https://shop.example.test")
		_, err := owner.repo.WithContext(ctx).ReplaceCurrent(&sitelogodomain.Asset{StorageKey: env.key, URL: env.provider.uploadReference}, func(tx *gorm.DB, _ *sitelogodomain.Asset) error { return callback(tx) })
		return err
	case SiteFaviconStoragePrefix:
		owner := NewSiteFaviconService(repository.NewSiteFaviconRepository(db), env.provider, "https://shop.example.test")
		return owner.repo.WithContext(ctx).ReplaceCurrent(&sitefavicondomain.Asset{StorageKey: env.key, URL: env.provider.uploadReference}, func(tx *gorm.DB, _ *sitefavicondomain.Asset) error { return callback(tx) })
	default:
		owner := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), env.provider, "https://shop.example.test")
		return owner.repo.WithContext(ctx).ReplaceItems("home-hero", "en", []homevisualtile.Tile{{StorageKey: env.key, ImageURL: env.provider.uploadReference, ThumbnailURL: env.provider.uploadReference, Title: "Owner", AltText: "Owner image", DesktopOrder: 1, Width: 600, Height: 600}}, func(tx *gorm.DB, _ []homevisualtile.Tile) error { return callback(tx) })
	}
}

func captureDedicatedPublicCleanupTransactionBackendPIDBeforeProviderDeletion(t *testing.T, env dedicatedPublicUploadFailureCompensationRegressionEnvironment) <-chan int {
	t.Helper()
	acquired := make(chan int, 1)
	var once sync.Once
	callback := "capture_dedicated_public_cleanup_transaction_" + uuid.NewString()
	table := map[string]string{SiteLogoStoragePrefix: "site_logo_assets", SiteFaviconStoragePrefix: "site_favicon_assets", HomeVisualTileStoragePrefix: "visual_showcase_items"}[env.namespace]
	// First() returns ErrRecordNotFound before the first identity owner exists.
	// Capture on entry so that expected absence cannot poison the PID query.
	require.NoError(t, env.db.Callback().Query().Before("gorm:query").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == table {
			once.Do(func() {
				var pid int
				if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
					tx.AddError(err)
				} else {
					acquired <- pid
				}
			})
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Query().Remove(callback) })
	return acquired
}

func TestPostgreSQLDedicatedPublicOwnerAdmissionAndCleanupSerializeInBothOrdersAndReleaseRolledBackOwnership(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, HomeVisualTileStoragePrefix} {
		for _, first := range []string{"owner", "cleanup", "rolled back owner"} {
			t.Run(namespace+"/"+first, func(t *testing.T) {
				env, event := createIsolatedPostgreSQLDedicatedPublicCleanupAndOwnerAdmissionTestEnvironment(t, namespace)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				locked, release := make(chan int, 1), make(chan struct{})
				var once sync.Once
				releaseFirst := func() { once.Do(func() { close(release) }) }
				t.Cleanup(releaseFirst)
				ownerResult, cleanupResult := make(chan error, 1), make(chan error, 1)
				rollback := errors.New("owner transaction deliberately rolled back")
				bind := func() {
					ownerResult <- replaceDedicatedPublicOwnerAfterNamespaceLockWithOptionalTransactionalCallback(ctx, env, env.db, func(tx *gorm.DB) error {
						if first == "cleanup" {
							return nil
						}
						pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, release)
						if first == "rolled back owner" {
							return rollback
						}
						return nil
					})
				}
				cleanup := func() { cleanupResult <- env.handler.Handle(ctx, event) }
				var pid int
				if first == "cleanup" {
					captured := captureDedicatedPublicCleanupTransactionBackendPIDBeforeProviderDeletion(t, env)
					env.provider.beforeDelete = func() {
						select {
						case <-release:
						case <-ctx.Done():
						}
					}
					go cleanup()
					pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, captured)
					go bind()
				} else {
					go bind()
					pid = awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, locked)
					go cleanup()
				}
				requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, env.db, pid)
				releaseFirst()
				ownerError := awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, ownerResult)
				require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult))
				switch first {
				case "owner":
					require.NoError(t, ownerError)
					require.Empty(t, env.provider.deletedKeys)
					requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
				case "rolled back owner":
					require.ErrorIs(t, ownerError, rollback)
					requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
				case "cleanup":
					requireDedicatedPublicOwnerAdmissionRejectedBecauseObjectIsMissing(t, namespace, ownerError)
					requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
				}
			})
		}
	}
}

type dedicatedPublicCleanupStorageThatDeletesOnlyAfterCancellationAndExplicitRelease struct {
	storage.StorageService
	entered chan struct{}
	release <-chan struct{}
}

func (s *dedicatedPublicCleanupStorageThatDeletesOnlyAfterCancellationAndExplicitRelease) Delete(ctx context.Context, key string) error {
	close(s.entered)
	<-s.release
	return s.StorageService.Delete(context.WithoutCancel(ctx), key)
}

func TestPostgreSQLDedicatedPublicCleanupRetainsNamespaceLockUntilCanceledProviderDeletionReturns(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, HomeVisualTileStoragePrefix} {
		t.Run(namespace, func(t *testing.T) {
			env, event := createIsolatedPostgreSQLDedicatedPublicCleanupAndOwnerAdmissionTestEnvironment(t, namespace)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			cleanupContext, cancelCleanup := context.WithCancel(ctx)
			defer cancelCleanup()
			entered, releaseDelete := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseDelete) }) }
			t.Cleanup(release)
			blocking := &dedicatedPublicCleanupStorageThatDeletesOnlyAfterCancellationAndExplicitRelease{StorageService: env.provider.StorageService, entered: entered, release: releaseDelete}
			env.logos.storage, env.favicons.storage, env.visuals.storage = blocking, blocking, blocking
			captured := captureDedicatedPublicCleanupTransactionBackendPIDBeforeProviderDeletion(t, env)
			cleanupResult, ownerResult := make(chan error, 1), make(chan error, 1)
			go func() { cleanupResult <- env.handler.Handle(cleanupContext, event) }()
			pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, captured)
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal("provider deletion did not start")
			}
			go func() {
				ownerResult <- replaceDedicatedPublicOwnerAfterNamespaceLockWithOptionalTransactionalCallback(ctx, env, env.db, nil)
			}()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, env.db, pid)
			cancelCleanup()
			returnedEarly := false
			select {
			case <-ownerResult:
				returnedEarly = true
				t.Error("owner admission returned while a canceled provider could still delete its bytes")
			case <-time.After(200 * time.Millisecond):
			}
			release()
			require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult), context.Canceled)
			if !returnedEarly {
				requireDedicatedPublicOwnerAdmissionRejectedBecauseObjectIsMissing(t, namespace, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, ownerResult))
			}
			requireDedicatedPublicUploadCompensationObjectMissing(t, env.provider, env.key)
			env.logos.storage, env.favicons.storage, env.visuals.storage = env.provider, env.provider, env.provider
			require.NoError(t, env.handler.Handle(ctx, event), "the original task remains safe to retry after cancellation")
		})
	}
}

func requireDedicatedPublicOwnerAdmissionRejectedBecauseObjectIsMissing(t *testing.T, namespace string, err error) {
	t.Helper()
	if namespace == HomeVisualTileStoragePrefix {
		require.ErrorIs(t, err, ErrHomeVisualTileImageInvalid)
	} else {
		require.ErrorIs(t, err, ErrMediaAssetURLUnavailable)
	}
}

func TestPostgreSQLDedicatedPublicOwnershipRejectsStaleSnapshotAndCleanupRejectsExternallyOwnedTransactionLifetime(t *testing.T) {
	for _, namespace := range []string{SiteLogoStoragePrefix, SiteFaviconStoragePrefix, HomeVisualTileStoragePrefix} {
		for _, operation := range []string{"owner in repeatable read", "cleanup in outer transaction"} {
			t.Run(namespace+"/"+operation, func(t *testing.T) {
				env, _ := createIsolatedPostgreSQLDedicatedPublicCleanupAndOwnerAdmissionTestEnvironment(t, namespace)
				err := env.db.Transaction(func(tx *gorm.DB) error {
					if operation == "owner in repeatable read" {
						return replaceDedicatedPublicOwnerAfterNamespaceLockWithOptionalTransactionalCallback(t.Context(), env, tx, nil)
					}
					if namespace == SiteLogoStoragePrefix {
						return repository.NewSiteLogoRepository(tx).WithLockedCurrentSiteLogoForObjectCleanup(t.Context(), func(*sitelogodomain.Asset, *repository.MediaRepository) error {
							t.Error("nested cleanup callback ran")
							return nil
						})
					}
					if namespace == SiteFaviconStoragePrefix {
						return repository.NewSiteFaviconRepository(tx).WithLockedCurrentSiteFaviconForObjectCleanup(t.Context(), func(*sitefavicondomain.Asset, *repository.MediaRepository) error {
							t.Error("nested cleanup callback ran")
							return nil
						})
					}
					return repository.NewHomeVisualTileRepository(tx).WithLockedVisualShowcaseReferencesForObjectCleanup(t.Context(), func(*repository.HomeVisualTileRepository, *repository.MediaRepository) error {
						t.Error("nested cleanup callback ran")
						return nil
					})
				}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
				if operation == "owner in repeatable read" {
					require.ErrorContains(t, err, "requires read committed transaction isolation")
				} else {
					require.ErrorContains(t, err, "requires its own transaction lifetime")
				}
				require.Empty(t, env.provider.deletedKeys)
				requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, env.provider, env.key)
				require.NoError(t, replaceDedicatedPublicOwnerAfterNamespaceLockWithOptionalTransactionalCallback(t.Context(), env, env.db, nil), "a rejected outer transaction must not strand namespace locks")
			})
		}
	}
}

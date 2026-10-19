package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx context.Context, t *testing.T, tx *gorm.DB, acquired chan<- int, release <-chan struct{}) {
	t.Helper()
	var pid int
	if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
		tx.AddError(err)
		return
	}
	acquired <- pid
	select {
	case <-release:
	case <-ctx.Done():
		tx.AddError(ctx.Err())
	}
}

func awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx context.Context, t *testing.T, acquired <-chan int) int {
	t.Helper()
	select {
	case pid := <-acquired:
		return pid
	case <-ctx.Done():
		t.Fatal("media object ownership transaction did not reach its locked boundary")
		return 0
	}
}

func awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx context.Context, t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		t.Fatal("media object ownership operation did not finish")
		return ctx.Err()
	}
}

func TestPostgreSQLMediaObjectRegistrationAndPhysicalCleanupSerializeBothOrdersWithoutDanglingRegistrations(t *testing.T) {
	for _, scenario := range []string{
		"original-registration", "windows-key-registration", "storefront-url-only-registration", "quota-original-registration",
		"create-derivative-registration", "replace-derivative-registration", "associated-derivative-registration",
		"derivative-object-original-registration", "retired-derivative-object-new-derivative-registration",
		"private-unregistered-derivative-original-registration", "private-unregistered-derivative-new-derivative-registration",
	} {
		for _, firstOperation := range []string{"register", "cleanup"} {
			t.Run(scenario+"/"+firstOperation+"-first", func(t *testing.T) {
				env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				db, provider := env.db.WithContext(ctx), env.storage.StorageService
				assets := NewMediaService(repository.NewMediaRepository(db), provider, nil, "https://shop.example.test", 20<<30)
				assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
				root, originalKey := env.publicRoot, "catalog/concurrent-retirement-original.png"
				if strings.HasPrefix(scenario, "private-") {
					root, originalKey = env.privateRoot, storage.CustomerServiceConversationAttachmentOriginalsPrefix+"/concurrent-retirement-original.png"
				}
				target := createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, originalKey)
				selectedKey := originalKey
				if root == env.privateRoot {
					require.NoError(t, db.Model(target).Update("visibility", "private").Error)
					selectedKey = fmt.Sprintf("%s/%d/card/not-in-sql.png", storage.CustomerServiceConversationAttachmentDerivativesPrefix, target.ID)
				} else if strings.HasPrefix(scenario, "derivative-object-") || strings.HasPrefix(scenario, "retired-derivative-object-") {
					selectedKey = "media-derivatives/concurrent-retirement-card.png"
					oldDerivative := media.MediaAssetDerivative{MediaAssetID: target.ID, Preset: "card", PresetVersion: 1, StorageKey: selectedKey, URL: provider.GetURL(selectedKey)}
					require.NoError(t, db.Create(&oldDerivative).Error)
					if strings.HasPrefix(scenario, "retired-") {
						require.NoError(t, db.Delete(&oldDerivative).Error)
					}
				}
				writeMediaCleanupOwnershipConflictTestObjects(t, root, originalKey, selectedKey)
				require.NoError(t, assets.repo.WithTransaction(func(repo *repository.MediaRepository, tx *gorm.DB) error {
					current, err := repo.FindAssetByIDForUpdate(target.ID)
					if err != nil {
						return err
					}
					return assets.retireUnreferencedMediaAssetWithinTransaction(repo, tx, current)
				}))
				var event outbox.Event
				require.NoError(t, db.Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(target.ID)).First(&event).Error)
				candidate := &media.MediaAsset{Filename: "new-registration.png", URL: provider.GetURL(selectedKey) + "?new-registration=1", StorageKey: selectedKey,
					MediaType: "image", MimeType: "image/png", Status: "active", Visibility: "public", Size: 1, UploaderID: 601, UploaderActorType: media.MediaUploadActorAdministrator}
				if scenario == "windows-key-registration" {
					candidate.StorageKey = strings.ReplaceAll(selectedKey, "/", "\\")
				}
				if scenario == "storefront-url-only-registration" {
					candidate.StorageKey, candidate.URL = "", "https://shop.example.test/uploads/"+selectedKey+"?new-registration=1"
				}
				if scenario == "private-unregistered-derivative-original-registration" {
					candidate.StorageKey = ""
				}
				derivativeRegistration := strings.Contains(scenario, "new-derivative-registration") || scenario == "create-derivative-registration" || scenario == "replace-derivative-registration"
				var derivativeOwner *media.MediaAsset
				var derivative media.MediaAssetDerivative
				if derivativeRegistration {
					derivativeOwner = createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, "catalog/new-derivative-owner.png")
					derivative = media.MediaAssetDerivative{MediaAssetID: derivativeOwner.ID, Preset: "card", PresetVersion: 1, URL: candidate.URL, StorageKey: selectedKey}
					if scenario == "replace-derivative-registration" {
						require.NoError(t, db.Create(&media.MediaAssetDerivative{MediaAssetID: derivativeOwner.ID, Preset: "card", PresetVersion: 1,
							URL: provider.GetURL("media-derivatives/previous-card.png"), StorageKey: "media-derivatives/previous-card.png"}).Error)
					}
				}
				if scenario == "associated-derivative-registration" {
					candidate.StorageKey, candidate.URL = "catalog/new-associated-owner.png", provider.GetURL("catalog/new-associated-owner.png")
					writeMediaCleanupOwnershipConflictTestObjects(t, env.publicRoot, candidate.StorageKey)
					candidate.Derivatives = []media.MediaAssetDerivative{{Preset: "card", PresetVersion: 1, URL: provider.GetURL(selectedKey) + "?associated=1", StorageKey: selectedKey}}
				}
				register := func() error {
					switch {
					case scenario == "replace-derivative-registration":
						_, err := assets.repo.ReplaceAssetDerivatives([]media.MediaAssetDerivative{derivative})
						return err
					case derivativeRegistration:
						return assets.repo.CreateAssetDerivatives([]media.MediaAssetDerivative{derivative})
					case scenario == "quota-original-registration":
						return assets.repo.CreateAssetWithinUploadActorStorageQuota(ctx, candidate, 20<<30)
					default:
						return assets.repo.CreateAsset(candidate)
					}
				}
				acquired, releaseFirst := make(chan int, 1), make(chan struct{})
				var releaseOnce, pauseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
				t.Cleanup(release)
				callbackName := "pause_media_object_admission_" + uuid.NewString()
				if firstOperation == "register" {
					table := "media_assets"
					if derivativeRegistration {
						table = "media_asset_derivatives"
					}
					require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
						if tx.Statement.Table == table {
							pauseOnce.Do(func() {
								pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx, t, tx, acquired, releaseFirst)
							})
						}
					}))
					t.Cleanup(func() { _ = db.Callback().Create().Remove(callbackName) })
				} else {
					require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
						if tx.Statement.Table == "media_assets" && tx.Statement.Unscoped {
							pauseOnce.Do(func() {
								pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx, t, tx, acquired, releaseFirst)
							})
						}
					}))
					t.Cleanup(func() { _ = db.Callback().Delete().Remove(callbackName) })
				}
				registerResult, cleanupResult := make(chan error, 1), make(chan error, 1)
				handler := NewObjectStorageCleanupOutboxHandler(assets, nil, nil, nil)
				if firstOperation == "register" {
					go func() { registerResult <- register() }()
				} else {
					go func() { cleanupResult <- handler.Handle(ctx, event) }()
				}
				pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, acquired)
				if firstOperation == "register" {
					go func() { cleanupResult <- handler.Handle(ctx, event) }()
				} else {
					go func() { registerResult <- register() }()
				}
				waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "pg_advisory_xact_lock")
				release()
				registrationError := awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, registerResult)
				cleanupError := awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, cleanupResult)
				if firstOperation == "register" {
					require.NoError(t, registrationError)
					require.ErrorIs(t, cleanupError, ErrMediaAssetInUse)
					require.FileExists(t, filepath.Join(root, filepath.FromSlash(selectedKey)))
					if derivativeRegistration {
						require.NoError(t, db.Unscoped().Where("media_asset_id = ?", derivativeOwner.ID).Delete(&media.MediaAssetDerivative{}).Error)
					} else {
						require.NoError(t, repository.NewMediaRepository(db).HardDeleteAsset(candidate.ID))
					}
					require.NoError(t, handler.Handle(ctx, event), "the original cleanup event must retry after the last conflicting registration is removed")
				} else {
					require.NoError(t, cleanupError)
					require.ErrorIs(t, registrationError, ErrMediaAssetURLUnavailable)
					if derivativeRegistration {
						var rows int64
						require.NoError(t, db.Model(&media.MediaAssetDerivative{}).Where("media_asset_id = ? AND url = ?", derivativeOwner.ID, derivative.URL).Count(&rows).Error)
						require.Zero(t, rows)
						if scenario == "replace-derivative-registration" {
							var previous int64
							previousDerivativeQueryError := db.Model(&media.MediaAssetDerivative{}).Where("media_asset_id = ? AND storage_key = ?", derivativeOwner.ID, "media-derivatives/previous-card.png").Count(&previous).Error
							require.NoError(t, previousDerivativeQueryError)
							require.EqualValues(t, 1, previous, "failed replacement must preserve previous active metadata")
						}
					} else {
						require.Zero(t, candidate.ID)
					}
				}
				require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(selectedKey)))
				require.NoFileExists(t, filepath.Join(root, filepath.FromSlash(originalKey)))
				require.NoError(t, handler.Handle(ctx, event), "duplicate delivery remains idempotent")
				var remainingEvents int64
				require.NoError(t, db.Model(&outbox.Event{}).Where("event_key = ?", event.EventKey).Count(&remainingEvents).Error)
				require.EqualValues(t, 1, remainingEvents)
			})
		}
	}
}

func TestPostgreSQLMediaDerivativeObjectRegistrationUsesOneSortedLockPlanForOppositeObjectOrders(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	db, provider := env.db.WithContext(ctx), env.storage.StorageService
	assets := NewMediaService(repository.NewMediaRepository(db), provider, nil, "", 20<<30)
	keys := []string{"media-derivatives/sorted-a.png", "media-derivatives/sorted-b.png"}
	writeMediaCleanupOwnershipConflictTestObjects(t, env.publicRoot, keys...)
	owners := []*media.MediaAsset{
		createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, "catalog/first-sorted-owner.png"),
		createRegisteredOriginalForMediaCleanupOwnershipConflictTest(t, db, provider, "catalog/second-sorted-owner.png"),
	}
	acquired, releaseFirst := make(chan int, 1), make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
	t.Cleanup(release)
	var observedFirstLock atomic.Bool
	callbackName := "pause_first_media_object_lock_" + uuid.NewString()
	require.NoError(t, db.Callback().Raw().After("gorm:raw").Register(callbackName, func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), "SELECT pg_advisory_xact_lock") && observedFirstLock.CompareAndSwap(false, true) {
			pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(ctx, t, tx, acquired, releaseFirst)
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Raw().Remove(callbackName) })
	register := func(owner uint, first, second string) error {
		return assets.repo.CreateAssetDerivatives([]media.MediaAssetDerivative{
			{MediaAssetID: owner, Preset: "first", PresetVersion: 1, URL: provider.GetURL(first) + fmt.Sprintf("?owner=%d", owner), StorageKey: first},
			{MediaAssetID: owner, Preset: "second", PresetVersion: 1, URL: provider.GetURL(second) + fmt.Sprintf("?owner=%d", owner), StorageKey: second},
		})
	}
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go func() { firstResult <- register(owners[0].ID, keys[0], keys[1]) }()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(ctx, t, acquired)
	go func() { secondResult <- register(owners[1].ID, keys[1], keys[0]) }()
	waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "pg_advisory_xact_lock")
	release()
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, firstResult))
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(ctx, t, secondResult))
}

func TestPostgreSQLMediaObjectRegistrationRejectsStaleSnapshotTransactions(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	provider := env.storage.StorageService
	key := "catalog/repeatable-read-registration.png"
	writeMediaCleanupOwnershipConflictTestObjects(t, env.publicRoot, key)
	err := env.db.Transaction(func(tx *gorm.DB) error {
		assets := NewMediaService(repository.NewMediaRepository(tx), provider, nil, "", 20<<30)
		return assets.repo.CreateAsset(&media.MediaAsset{Filename: "repeatable-read.png", URL: provider.GetURL(key), StorageKey: key})
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	require.ErrorContains(t, err, "requires read committed")
	require.False(t, errors.Is(err, ErrMediaAssetURLUnavailable))
}

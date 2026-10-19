package service

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createIsolatedPostgreSQLBusinessDraftMediaReceiptEnvironment(t *testing.T) businessDraftMediaReceiptTestEnvironment {
	t.Helper()
	base, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	return createBusinessDraftMediaReceiptTestEnvironment(t, base.db)
}

func TestPostgreSQLBusinessDraftMediaRefundPolicySaveAndOrderSnapshotCreationSerializeWithExpiryAndRetirement(t *testing.T) {
	for _, source := range []string{"policy save", "order snapshot"} {
		for _, first := range []string{"binding", "cleanup", "rolled back binding"} {
			t.Run(source+"/"+first, func(t *testing.T) {
				env := createIsolatedPostgreSQLBusinessDraftMediaReceiptEnvironment(t)
				policies, settings := configureBusinessDraftMediaRefundPolicyTestService(t, env)
				_, err := policies.Update("en", refundcancellation.DefaultPolicy())
				require.NoError(t, err)
				asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				bind := func() error {
					_, err := policies.Update("en", businessDraftMediaRefundPolicyTestContent(asset.URL))
					return err
				}
				var expired BusinessDraftMediaUploadExpiryBatchResult
				cleanup := func() error {
					var err error
					expired, err = env.service.ExpireUnboundBusinessDraftMediaUploads(ctx, 10)
					return err
				}
				if source == "order snapshot" {
					_, err := policies.Update("en", businessDraftMediaRefundPolicyTestContent(asset.URL))
					require.NoError(t, err)
					snapshot, err := policies.BuildOrderDisclosure(settings, 91, "en", "", "", nil)
					require.NoError(t, err)
					_, err = policies.Update("en", refundcancellation.DefaultPolicy())
					require.NoError(t, err)
					bind = func() error {
						return repository.NewOrderPolicyDisclosureRepository(env.db.WithContext(ctx)).Create(snapshot)
					}
					cleanup = func() error { return env.service.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID)) }
				} else {
					expireBusinessDraftMediaReceiptTestClock(t, env)
				}
				acquired, release := make(chan int, 1), make(chan struct{})
				var once sync.Once
				releaseFirst := func() { once.Do(func() { close(release) }) }
				t.Cleanup(releaseFirst)
				callback := "pause_business_draft_policy_" + uuid.NewString()
				rollback := errors.New("policy binding intentionally rolled back")
				pause := func(tx *gorm.DB) {
					pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, release)
					if first == "rolled back binding" {
						tx.AddError(rollback)
					}
				}
				if first == "cleanup" {
					require.NoError(t, env.db.Callback().Delete().After("gorm:delete").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "media_assets" {
							pause(tx)
						}
					}))
					t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
				} else if source == "policy save" {
					require.NoError(t, env.db.Callback().Update().After("gorm:update").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "settings" {
							pause(tx)
						}
					}))
					t.Cleanup(func() { _ = env.db.Callback().Update().Remove(callback) })
				} else {
					require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
						if tx.Statement.Table == "order_policy_disclosures" {
							pause(tx)
						}
					}))
					t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
				}
				bindingResult, cleanupResult := make(chan error, 1), make(chan error, 1)
				if first == "cleanup" {
					go func() { cleanupResult <- cleanup() }()
				} else {
					go func() { bindingResult <- bind() }()
				}
				pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, acquired)
				if first == "cleanup" {
					go func() { bindingResult <- bind() }()
				} else {
					go func() { cleanupResult <- cleanup() }()
				}
				requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, pid)
				releaseFirst()
				bindErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, bindingResult)
				cleanupErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, cleanupResult)
				if first == "binding" {
					require.NoError(t, bindErr)
					if source == "order snapshot" {
						require.ErrorIs(t, cleanupErr, ErrMediaAssetInUse)
					} else {
						require.NoError(t, cleanupErr)
						require.Zero(t, expired.Retired)
					}
					assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
				} else {
					require.Error(t, bindErr)
					require.NoError(t, cleanupErr)
					if source == "policy save" {
						require.Equal(t, 1, expired.Retired)
						require.Zero(t, expired.Errors)
						replayBusinessDraftMediaReceiptTestCleanupEvents(t, env)
					}
					assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, false)
				}
			})
		}
	}
}

func TestPostgreSQLBusinessDraftMediaBindingAndExpirySerializeInBothOrdersAndBusinessRollback(t *testing.T) {
	for _, first := range []string{"binding", "expiry", "rolled back binding"} {
		t.Run(first, func(t *testing.T) {
			env := createIsolatedPostgreSQLBusinessDraftMediaReceiptEnvironment(t)
			asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
			expireBusinessDraftMediaReceiptTestClock(t, env)
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			acquired, release := make(chan int, 1), make(chan struct{})
			var once sync.Once
			releaseFirst := func() { once.Do(func() { close(release) }) }
			t.Cleanup(releaseFirst)
			callback := "pause_business_draft_" + uuid.NewString()
			rollback := errors.New("business write intentionally rolled back")
			if first == "expiry" {
				require.NoError(t, env.db.Callback().Delete().After("gorm:delete").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "media_assets" {
						pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, release)
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Delete().Remove(callback) })
			} else {
				require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table == "product_categories" {
						pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, acquired, release)
						if first == "rolled back binding" {
							tx.AddError(rollback)
						}
					}
				}))
				t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callback) })
			}
			repo := repository.NewProductCategoryRepository(env.db.WithContext(ctx))
			repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(env.service)
			category := product.ProductCategory{Name: "Concurrent draft category", Slug: uuid.NewString(), ImageMediaAssetID: &asset.ID}
			bindingResult, expiryResult := make(chan error, 1), make(chan error, 1)
			var expired BusinessDraftMediaUploadExpiryBatchResult
			bind := func() { bindingResult <- repo.Create(&category) }
			expire := func() {
				var err error
				expired, err = env.service.ExpireUnboundBusinessDraftMediaUploads(ctx, 10)
				expiryResult <- err
			}
			if first == "expiry" {
				go expire()
			} else {
				go bind()
			}
			pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, acquired)
			if first == "expiry" {
				go bind()
			} else {
				go expire()
			}
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, pid)
			releaseFirst()
			bindErr := awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, bindingResult)
			require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, expiryResult))
			require.Zero(t, expired.Errors)
			if first == "binding" {
				require.NoError(t, bindErr)
				require.Zero(t, expired.Retired)
				var receipt media.BusinessDraftMediaUploadReceipt
				require.NoError(t, env.db.First(&receipt).Error)
				require.Equal(t, "bound", receipt.Status)
				assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
			} else {
				require.Error(t, bindErr)
				require.Equal(t, 1, expired.Retired)
				replayBusinessDraftMediaReceiptTestCleanupEvents(t, env)
				assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, false)
			}
		})
	}
}

func TestPostgreSQLBusinessDraftMediaExpiryWaitsForCanceledLateProviderWriteBeforeDeletingReceiptAndBytes(t *testing.T) {
	env := createIsolatedPostgreSQLBusinessDraftMediaReceiptEnvironment(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	uploadContext, cancelUpload := context.WithCancel(ctx)
	defer cancelUpload()
	acquired, writeStarted, release := make(chan int, 1), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	releaseProvider := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseProvider)
	createCallback := "short_business_draft_deadline_" + uuid.NewString()
	require.NoError(t, env.db.Callback().Create().Before("gorm:create").Register(createCallback, func(tx *gorm.DB) {
		if receipt, ok := tx.Statement.Dest.(*media.BusinessDraftMediaUploadReceipt); ok {
			receipt.ExpiresAt = time.Now().UTC().Add(400 * time.Millisecond)
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Create().Remove(createCallback) })
	queryCallback := "capture_business_draft_receipt_holder_" + uuid.NewString()
	var captureOnce sync.Once
	require.NoError(t, env.db.Callback().Query().After("gorm:query").Register(queryCallback, func(tx *gorm.DB) {
		if tx.Statement.Table == "business_draft_media_upload_receipts" && tx.Statement.Clauses["FOR"].Expression != nil {
			captureOnce.Do(func() {
				var pid int
				if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
					tx.AddError(err)
				} else {
					acquired <- pid
				}
			})
		}
	}))
	t.Cleanup(func() { _ = env.db.Callback().Query().Remove(queryCallback) })
	provider := &businessDraftMediaPartialWriteTestStorage{StorageService: env.store, ignoreWriteCancellation: true, beforeWrite: func(string) {
		close(writeStarted)
		select {
		case <-release:
		case <-ctx.Done():
		}
	}}
	env.service.storage = provider
	uploadResult := make(chan error, 1)
	input := businessDraftMediaReceiptTestInput(t)
	go func() { _, err := env.service.UploadBusinessDraftMediaAsset(uploadContext, input); uploadResult <- err }()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, acquired)
	select {
	case <-writeStarted:
	case <-ctx.Done():
		t.Fatal("draft upload never reached provider")
	}
	require.Eventually(t, func() bool {
		var expired bool
		return env.db.Raw("SELECT expires_at <= NOW() FROM business_draft_media_upload_receipts LIMIT 1").Scan(&expired).Error == nil && expired
	}, 3*time.Second, 20*time.Millisecond)
	expiryResult := make(chan error, 1)
	var expired BusinessDraftMediaUploadExpiryBatchResult
	go func() {
		var err error
		expired, err = env.service.ExpireUnboundBusinessDraftMediaUploads(ctx, 10)
		expiryResult <- err
	}()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, pid)
	cancelUpload()
	// Cancellation cannot release the transaction while synchronous Put can
	// still finish. The waiting expiry must remain blocked on the same PID.
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, pid)
	releaseProvider()
	require.ErrorIs(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, uploadResult), context.Canceled)
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, expiryResult))
	require.Equal(t, 1, expired.Retired)
	require.Zero(t, expired.Errors)
	replayBusinessDraftMediaReceiptTestCleanupEvents(t, env)
}

func TestPostgreSQLBusinessDraftMediaRejectsStaleExpirySnapshotAndExternallyControlledProviderTransactions(t *testing.T) {
	env := createIsolatedPostgreSQLBusinessDraftMediaReceiptEnvironment(t)
	asset := uploadBusinessDraftMediaReceiptTestAsset(t, env)
	expireBusinessDraftMediaReceiptTestClock(t, env)
	require.NoError(t, env.db.Transaction(func(tx *gorm.DB) error {
		copied := *env.service
		copied.repo = env.service.repo.WithTx(tx)
		result, err := copied.ExpireUnboundBusinessDraftMediaUploads(t.Context(), 10)
		require.ErrorContains(t, err, "own transaction lifetime")
		require.Zero(t, result.Retired)
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead}))
	require.NoError(t, env.db.Transaction(func(tx *gorm.DB) error {
		copied := *env.service
		copied.repo = env.service.repo.WithTx(tx)
		_, err := copied.UploadBusinessDraftMediaAsset(t.Context(), businessDraftMediaReceiptTestInput(t))
		require.ErrorContains(t, err, "own transaction lifetime")
		require.ErrorContains(t, copied.cleanupExpiredBusinessDraftMediaObjectsFromDurableReceipt(t.Context(), uuid.NewString(), []string{asset.StorageKey}), "own transaction lifetime")
		require.ErrorContains(t, copied.cleanupDeletedAsset(t.Context(), asset.ID), "own transaction lifetime")
		return nil
	}))
	assertBusinessDraftMediaReceiptTestObjectsExist(t, env, asset, true)
}

package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createProductViewReceiptIntegrationProduct(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	var id uint
	require.NoError(t, db.Raw(`INSERT INTO products(name,slug,price_minor,currency,status,locale,view_count) VALUES('View receipt integration',?,100,'USD','active','en',0) RETURNING id`, "view-receipt-"+uuid.NewString()).Scan(&id).Error)
	require.Positive(t, id)
	return id
}

func readProductViewReceiptIntegrationSQLCount(t *testing.T, db *gorm.DB, id uint) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Raw(`SELECT view_count FROM products WHERE id=?`, id).Scan(&count).Error)
	return count
}

func TestPostgreSQLAndRedisProductViewCountAtomicReplayReceiptsAndConcurrentBufferAcknowledgement(t *testing.T) {
	if os.Getenv("LOCAL_BACKUP_DISPOSABLE_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_DISPOSABLE_REDIS_TEST_SERVER") != "true" || os.Getenv("LOCAL_BACKUP_INTEGRATION_REDIS_URL") == "" {
		t.Skip("requires explicitly disposable PostgreSQL and Redis servers")
	}
	db := createCompletelyMigratedDisposableTemplateCategoryMediaBindingTestDatabase(t)
	options, err := redis.ParseURL(os.Getenv("LOCAL_BACKUP_INTEGRATION_REDIS_URL"))
	require.NoError(t, err)
	client := redis.NewClient(options)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	require.NoError(t, client.Ping(t.Context()).Err())
	initialKeys, err := client.DBSize(t.Context()).Result()
	require.NoError(t, err)
	require.Zero(t, initialKeys, "never reuse a Redis database that already contains data")
	repo := repository.NewProductRepository(db)
	clearOnlyOwnedViewBufferKeys := func(t *testing.T) {
		t.Helper()
		cleanupContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		require.NoError(t, client.Del(cleanupContext, productViewCountActiveProcessingAndBatchIdentityRedisKeys()...).Err())
	}

	t.Run("migration_down_preserves_nonempty_replay_protection_and_rolls_back_empty_schema_trial", func(t *testing.T) {
		down, err := os.ReadFile(filepath.Join(os.Getenv("LOCAL_BACKUP_INTEGRATION_MIGRATIONS_DIRECTORY"), "436_add_product_view_count_flush_batch_receipts_for_atomic_replay_protection.down.sql"))
		require.NoError(t, err)
		trialRollback := errors.New("restore schema after empty-ledger down trial")
		err = db.Transaction(func(tx *gorm.DB) error {
			if e := tx.Exec(string(down)).Error; e != nil {
				return e
			}
			return trialRollback
		})
		require.ErrorIs(t, err, trialRollback)
		id := createProductViewReceiptIntegrationProduct(t, db)
		_, err = repo.ApplyProductViewCountBatchWithAtomicReplayReceipts(t.Context(), uuid.NewString(), map[uint]int64{id: 2})
		require.NoError(t, err)
		err = db.Transaction(func(tx *gorm.DB) error { return tx.Exec(string(down)).Error })
		require.ErrorContains(t, err, "Cannot remove product view flush replay protection")
		require.Equal(t, int64(2), readProductViewReceiptIntegrationSQLCount(t, db, id))
	})

	t.Run("sixteen_concurrent_SQL_replays_apply_one_increment", func(t *testing.T) {
		id := createProductViewReceiptIntegrationProduct(t, db)
		batchID := uuid.NewString()
		start := make(chan struct{})
		results := make(chan error, 16)
		var applied atomic.Int64
		for worker := 0; worker < 16; worker++ {
			go func() {
				<-start
				count, err := repo.ApplyProductViewCountBatchWithAtomicReplayReceipts(t.Context(), batchID, map[uint]int64{id: 7})
				applied.Add(int64(count))
				results <- err
			}()
		}
		close(start)
		for worker := 0; worker < 16; worker++ {
			require.NoError(t, <-results)
		}
		require.Equal(t, int64(1), applied.Load())
		require.Equal(t, int64(7), readProductViewReceiptIntegrationSQLCount(t, db, id))
		_, err = repo.ApplyProductViewCountBatchWithAtomicReplayReceipts(t.Context(), batchID, map[uint]int64{id: 8})
		require.ErrorContains(t, err, "conflicts with immutable batch delta")
		require.Equal(t, int64(7), readProductViewReceiptIntegrationSQLCount(t, db, id))
	})

	t.Run("SQL_commit_with_lost_Redis_acknowledgement_survives_new_buffer_instance", func(t *testing.T) {
		clearOnlyOwnedViewBufferKeys(t)
		t.Cleanup(func() { clearOnlyOwnedViewBufferKeys(t) })
		id := createProductViewReceiptIntegrationProduct(t, db)
		hook := &rejectProductViewRedisAcknowledgementAfterSQLCommitHook{}
		faultClient := redis.NewClient(options)
		defer faultClient.Close()
		faultClient.AddHook(hook)
		hook.enabled.Store(true)
		buffer := NewProductViewCountBuffer(faultClient, repo, 2)
		require.NoError(t, buffer.Increment(t.Context(), id))
		_, err := buffer.Flush(t.Context())
		require.ErrorContains(t, err, "SQL batch committed")
		require.Equal(t, int64(1), readProductViewReceiptIntegrationSQLCount(t, db, id))
		require.NoError(t, buffer.Increment(t.Context(), id))
		report, err := buffer.InspectReadOnlyProductViewRecoveryBoundary(t.Context())
		require.NoError(t, err)
		require.Equal(t, 1, report.CommittedProcessingReceipts)
		require.Equal(t, int64(1), report.ActiveDeltaTotal)
		require.False(t, report.ReleaseAvailable)
		restarted := NewProductViewCountBuffer(client, repo, 2)
		applied, err := restarted.Flush(t.Context())
		require.NoError(t, err)
		require.Zero(t, applied)
		_, err = restarted.Flush(t.Context())
		require.NoError(t, err)
		require.Equal(t, int64(2), readProductViewReceiptIntegrationSQLCount(t, db, id))
	})

	t.Run("SQL_overflow_rolls_back_earlier_increment_and_all_receipts_without_merging_Redis", func(t *testing.T) {
		clearOnlyOwnedViewBufferKeys(t)
		t.Cleanup(func() { clearOnlyOwnedViewBufferKeys(t) })
		firstID, overflowID := createProductViewReceiptIntegrationProduct(t, db), createProductViewReceiptIntegrationProduct(t, db)
		require.NoError(t, db.Exec(`UPDATE products SET view_count=2147483647 WHERE id=?`, overflowID).Error)
		require.NoError(t, client.HSet(t.Context(), productViewCountRedisKey, strconv.FormatUint(uint64(firstID), 10), "2", strconv.FormatUint(uint64(overflowID), 10), "1").Err())
		buffer := NewProductViewCountBuffer(client, repo, 5)
		_, err := buffer.Flush(t.Context())
		require.Error(t, err)
		batchID, err := client.Get(t.Context(), productViewCountProcessingBatchIdentityRedisKey).Result()
		require.NoError(t, err)
		require.Equal(t, int64(0), readProductViewReceiptIntegrationSQLCount(t, db, firstID))
		var count int64
		require.NoError(t, db.Model(&product.ProductViewCountFlushReceipt{}).Where("batch_id=?", batchID).Count(&count).Error)
		require.Zero(t, count)
		require.Equal(t, int64(2), client.HLen(t.Context(), productViewCountProcessingRedisKey).Val())
		require.NoError(t, db.Exec(`UPDATE products SET view_count=0 WHERE id=?`, overflowID).Error)
		_, err = buffer.Flush(t.Context())
		require.NoError(t, err)
		require.Equal(t, int64(2), readProductViewReceiptIntegrationSQLCount(t, db, firstID))
		require.Equal(t, int64(1), readProductViewReceiptIntegrationSQLCount(t, db, overflowID))
	})

	t.Run("receipt_survives_deleted_product_and_absorbs_replay_after_same_ID_reappears", func(t *testing.T) {
		for _, previouslyApplied := range []bool{false, true} {
			id := createProductViewReceiptIntegrationProduct(t, db)
			batchID := uuid.NewString()
			if !previouslyApplied {
				require.NoError(t, db.Exec(`UPDATE products SET deleted_at=now() WHERE id=?`, id).Error)
			}
			_, err := repo.ApplyProductViewCountBatchWithAtomicReplayReceipts(t.Context(), batchID, map[uint]int64{id: 4})
			require.NoError(t, err)
			require.NoError(t, db.Exec(`DELETE FROM products WHERE id=?`, id).Error)
			var receipt product.ProductViewCountFlushReceipt
			require.NoError(t, db.Where("batch_id=? AND product_id=?", batchID, id).Take(&receipt).Error)
			require.Equal(t, previouslyApplied, receipt.ProductIncrementApplied)
			require.NoError(t, db.Exec(`INSERT INTO products(id,name,slug,price_minor,currency,locale,status,view_count) VALUES(?,'Reappeared product',?,100,'USD','en','active',0)`, id, "reappeared-"+uuid.NewString()).Error)
			applied, err := repo.ApplyProductViewCountBatchWithAtomicReplayReceipts(t.Context(), batchID, map[uint]int64{id: 4})
			require.NoError(t, err)
			require.Zero(t, applied)
			require.Equal(t, int64(0), readProductViewReceiptIntegrationSQLCount(t, db, id))
		}
	})

	t.Run("eight_flushers_and_four_writers_preserve_all_views_across_overlapping_batches", func(t *testing.T) {
		clearOnlyOwnedViewBufferKeys(t)
		t.Cleanup(func() { clearOnlyOwnedViewBufferKeys(t) })
		ids := []uint{createProductViewReceiptIntegrationProduct(t, db), createProductViewReceiptIntegrationProduct(t, db), createProductViewReceiptIntegrationProduct(t, db)}
		start := make(chan struct{})
		errorsChannel := make(chan error, 12)
		var wait sync.WaitGroup
		for worker := 0; worker < 4; worker++ {
			wait.Add(1)
			go func() {
				defer wait.Done()
				<-start
				buffer := NewProductViewCountBuffer(client, repo, 2)
				for round := 0; round < 20; round++ {
					for _, id := range ids {
						if err := buffer.Increment(t.Context(), id); err != nil {
							errorsChannel <- err
							return
						}
					}
				}
				errorsChannel <- nil
			}()
		}
		for worker := 0; worker < 8; worker++ {
			wait.Add(1)
			go func() {
				defer wait.Done()
				<-start
				buffer := NewProductViewCountBuffer(client, repo, 2)
				for round := 0; round < 20; round++ {
					if _, err := buffer.Flush(t.Context()); err != nil {
						errorsChannel <- err
						return
					}
				}
				errorsChannel <- nil
			}()
		}
		close(start)
		wait.Wait()
		close(errorsChannel)
		for err := range errorsChannel {
			require.NoError(t, err)
		}
		buffer := NewProductViewCountBuffer(client, repo, 2)
		for remaining := 0; remaining < 20; remaining++ {
			_, err := buffer.Flush(t.Context())
			require.NoError(t, err)
			report, err := buffer.InspectReadOnlyProductViewRecoveryBoundary(t.Context())
			require.NoError(t, err)
			if report.RedisBuffersEmpty {
				break
			}
		}
		report, err := buffer.InspectReadOnlyProductViewRecoveryBoundary(t.Context())
		require.NoError(t, err)
		require.True(t, report.RedisBuffersEmpty)
		for _, id := range ids {
			require.Equal(t, int64(80), readProductViewReceiptIntegrationSQLCount(t, db, id))
		}
		var receiptsTotal int64
		require.NoError(t, db.Model(&product.ProductViewCountFlushReceipt{}).Select("SUM(delta)").Where("product_id IN ?", ids).Scan(&receiptsTotal).Error)
		require.Equal(t, int64(240), receiptsTotal)
	})

	t.Run("legacy_processing_is_not_assigned_a_fabricated_batch_or_deleted", func(t *testing.T) {
		clearOnlyOwnedViewBufferKeys(t)
		t.Cleanup(func() { clearOnlyOwnedViewBufferKeys(t) })
		id := createProductViewReceiptIntegrationProduct(t, db)
		field := strconv.FormatUint(uint64(id), 10)
		require.NoError(t, client.HSet(t.Context(), productViewCountProcessingRedisKey, field, "3").Err())
		require.NoError(t, client.HSet(t.Context(), productViewCountRedisKey, field, "2").Err())
		_, err := NewProductViewCountBuffer(client, repo, 2).Flush(t.Context())
		require.ErrorContains(t, err, "legacy_product_view_processing_batch")
		require.Equal(t, "3", client.HGet(t.Context(), productViewCountProcessingRedisKey, field).Val())
		require.Equal(t, "2", client.HGet(t.Context(), productViewCountRedisKey, field).Val())
		require.Equal(t, int64(0), readProductViewReceiptIntegrationSQLCount(t, db, id))
		require.Zero(t, client.Exists(t.Context(), productViewCountProcessingBatchIdentityRedisKey).Val())
	})
}

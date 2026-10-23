package localbackup

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/maintenancegate"
)

func verifyStoppedProductDetailCacheExactInvalidationPartialAcknowledgementRetryAndPreservedRedisData(t *testing.T, manager *Manager, taskID string, redisOptions *redis.Options, invokeHTTP func(string, any, int) *httptest.ResponseRecorder) {
	t.Helper()
	ctx := context.Background()
	cfg := manager.Configuration
	route := "product-catalog-stopped-detail-cache-invalidation"
	request := ProductCatalogStoppedSwitchConfirmationRequest{Confirmation: "INVALIDATE STOPPED PRODUCT CACHE " + cfg.SiteID + "/" + cfg.Environment + "/" + taskID}
	before, err := manager.GetTask(taskID)
	require.NoError(t, err)
	switchDigest := before.Journal.ProductCatalogStoppedSwitch.SHA256
	client := redis.NewClient(redisOptions)
	defer client.Close()
	ownedKeys := []string{}
	defer func() {
		if len(ownedKeys) > 0 {
			require.NoError(t, client.Del(ctx, ownedKeys...).Err(), "remove only this test's synthetic keys")
		}
		require.NoError(t, client.Close())
	}()
	cacheValues := map[string]string{"product:1": "synthetic-original-detail", "product:999999": "synthetic-deleted-product-detail", "product:slug:old-slug:EN-us": "synthetic-obsolete-slug", "product:slug:old:slug:": "synthetic-colon-slug", "lock:product:1": "synthetic-reader-token", "lock:product:slug:old-slug:EN-us": "synthetic-slug-reader-token"}
	for index := 0; index < 130; index++ {
		cacheValues[fmt.Sprintf("product:slug:synthetic-old-%03d:en", index)] = fmt.Sprintf("synthetic-value-%03d", index)
	}
	for key, value := range cacheValues {
		ownedKeys = append(ownedKeys, key)
		require.NoError(t, client.Set(ctx, key, value, 10*time.Minute).Err())
	}
	for _, key := range []string{"product:view_counts:synthetic_unrelated_diagnostic", "lock:product:view_counts", "product:01", "product:123:other", "idempotency:synthetic_product_restore"} {
		ownedKeys = append(ownedKeys, key)
		require.NoError(t, client.Set(ctx, key, "synthetic-preserved-fact", 0).Err())
	}
	ownedKeys = append(ownedKeys, "customer_service:synthetic_stream", "jobs:synthetic_pending")
	require.NoError(t, client.XAdd(ctx, &redis.XAddArgs{Stream: "customer_service:synthetic_stream", Values: map[string]interface{}{"message": "synthetic-customer-message"}}).Err())
	require.NoError(t, client.RPush(ctx, "jobs:synthetic_pending", "synthetic-unconsumed-job").Err())
	preserved := map[string]string{}
	for _, key := range ownedKeys {
		if _, isCache := cacheValues[key]; !isCache {
			preserved[key], err = client.Dump(ctx, key).Result()
			require.NoError(t, err)
		}
	}
	marker, err := client.Get(ctx, maintenancegate.ProductViewRedisOwnershipKey).Result()
	require.NoError(t, err)
	// Atomic batch rejection must leave an earlier valid observation untouched.
	digest := sha1.Sum([]byte(cacheValues["product:1"]))
	batchKeys := []string{maintenancegate.ProductViewRedisOwnershipKey, "product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id", "product:1", "product:999999"}
	require.Error(t, deleteStoppedProductDetailCacheKeysOnlyIfObservationsStillMatch.Run(ctx, client, batchKeys, marker, hex.EncodeToString(digest[:]), strings.Repeat("0", 40)).Err())
	require.Equal(t, cacheValues["product:1"], client.Get(ctx, "product:1").Val())
	require.Error(t, deleteStoppedProductDetailCacheKeysOnlyIfObservationsStillMatch.Run(ctx, client, batchKeys[:5], "synthetic-wrong-ownership", hex.EncodeToString(digest[:])).Err())
	require.NoError(t, client.HSet(ctx, "product:view_counts", "synthetic", "1").Err())
	require.Error(t, deleteStoppedProductDetailCacheKeysOnlyIfObservationsStillMatch.Run(ctx, client, batchKeys[:5], marker, hex.EncodeToString(digest[:])).Err())
	require.Equal(t, cacheValues["product:1"], client.Get(ctx, "product:1").Val())
	require.Equal(t, "1", client.HGet(ctx, "product:view_counts", "synthetic").Val())
	require.NoError(t, client.Del(ctx, "product:view_counts").Err())
	require.NoError(t, client.Close())
	// A live ordinary Redis client blocks the operation before durable deletion
	// intent. A failure at this point can be retried in the same stopped window.
	blocker := redis.NewClient(redisOptions)
	require.NoError(t, blocker.Ping(ctx).Err())
	invokeHTTP(route, request, http.StatusOK)
	failed := waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "other Redis clients")
	require.Nil(t, failed.ProductCatalogStoppedDetailCacheInvalidation)
	require.NoError(t, blocker.Close())
	client = redis.NewClient(redisOptions)
	// Wrong Redis types block the entire inventory before any key is deleted.
	wrongType := "product:slug:synthetic-wrong-type:en"
	ownedKeys = append(ownedKeys, wrongType)
	require.NoError(t, client.HSet(ctx, wrongType, "unexpected", "hash").Err())
	require.NoError(t, client.Close())
	invokeHTTP(route, request, http.StatusOK)
	failed = waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "unexpected Redis type")
	require.Nil(t, failed.ProductCatalogStoppedDetailCacheInvalidation)
	client = redis.NewClient(redisOptions)
	require.Equal(t, int64(len(cacheValues)), client.Exists(ctx, sortedSyntheticStoppedProductCacheKeys(cacheValues)...).Val())
	require.NoError(t, client.Del(ctx, wrongType).Err())
	require.NoError(t, client.Close())
	// Save intent, then simulate interruption before any deletion.
	failed, err = manager.GetTask(taskID)
	require.NoError(t, err)
	require.NoError(t, manager.startTaskOperation(failed, "synthetic_cache_intent_failure", func(ctx context.Context, task *Task) error {
		return manager.invalidateProductDetailCacheWithinContinuouslyStoppedSwitch(ctx, task, "synthetic-cache-operator", func(point string) error {
			if point == "cache_intent_persisted" {
				return errors.New("synthetic cache intent interruption")
			}
			return nil
		})
	}))
	failed = waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "synthetic cache intent")
	progress := failed.ProductCatalogStoppedDetailCacheInvalidation
	require.Equal(t, "deleting", progress.Phase)
	require.Equal(t, len(cacheValues)-2, progress.PlannedDetailCacheKeys)
	require.Equal(t, 2, progress.PlannedReadLeaseKeys)
	encodedProgress, err := json.Marshal(progress)
	require.NoError(t, err)
	require.NotContains(t, string(encodedProgress), "synthetic-reader-token")
	require.Equal(t, switchDigest, failed.Journal.ProductCatalogStoppedSwitch.SHA256)
	// New keys cannot be silently adopted by retry.
	newKey := "product:slug:synthetic-new-after-intent:en"
	ownedKeys = append(ownedKeys, newKey)
	client = redis.NewClient(redisOptions)
	require.NoError(t, client.Set(ctx, newKey, "synthetic-new-value", 0).Err())
	require.NoError(t, client.Close())
	invokeHTTP(route, request, http.StatusOK)
	failed = waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "new or changed")
	client = redis.NewClient(redisOptions)
	require.Equal(t, int64(len(cacheValues)), client.Exists(ctx, sortedSyntheticStoppedProductCacheKeys(cacheValues)...).Val())
	require.NoError(t, client.Del(ctx, newKey).Err())
	require.NoError(t, client.Set(ctx, "product:1", "synthetic-changed-value", 0).Err())
	require.NoError(t, client.Close())
	invokeHTTP(route, request, http.StatusOK)
	failed = waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "new or changed")
	client = redis.NewClient(redisOptions)
	require.NoError(t, client.Set(ctx, "product:1", cacheValues["product:1"], 10*time.Minute).Err())
	require.NoError(t, client.Close())
	// Lost reply after the first 128-key atomic batch. Retry accepts missing
	// original observations, and never recaptures the remaining/new values.
	require.NoError(t, manager.startTaskOperation(failed, "synthetic_cache_delete_acknowledgement_loss", func(ctx context.Context, task *Task) error {
		return manager.invalidateProductDetailCacheWithinContinuouslyStoppedSwitch(ctx, task, "synthetic-cache-operator", func(point string) error {
			if point == "cache_batch_deleted_before_acknowledgement" {
				return errors.New("synthetic cache batch reply loss")
			}
			return nil
		})
	}))
	failed = waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "reply loss")
	require.Equal(t, progress.SHA256, failed.ProductCatalogStoppedDetailCacheInvalidation.SHA256)
	client = redis.NewClient(redisOptions)
	require.Equal(t, int64(len(cacheValues)-128), client.Exists(ctx, sortedSyntheticStoppedProductCacheKeys(cacheValues)...).Val())
	require.NoError(t, client.Close())
	invokeHTTP(route, request, http.StatusOK)
	completed := waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "product_module_switched_still_isolated", completed.State, completed.Error)
	require.Equal(t, "verified_absent", completed.ProductCatalogStoppedDetailCacheInvalidation.Phase)
	require.Equal(t, len(cacheValues), completed.ProductCatalogStoppedDetailCacheInvalidation.VerifiedAbsentKeys)
	require.False(t, completed.ProductCatalogRecovery.OnlineOverwriteAvailable)
	require.Equal(t, switchDigest, completed.Journal.ProductCatalogStoppedSwitch.SHA256)
	invokeHTTP(route, request, http.StatusOK)
	completed = waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "product_module_switched_still_isolated", completed.State, completed.Error)
	client = redis.NewClient(redisOptions)
	require.Zero(t, client.Exists(ctx, sortedSyntheticStoppedProductCacheKeys(cacheValues)...).Val())
	for key, value := range preserved {
		require.Equal(t, value, client.Dump(ctx, key).Val(), key)
	}
	require.Equal(t, marker, client.Get(ctx, maintenancegate.ProductViewRedisOwnershipKey).Val())
	require.NoError(t, client.Close())
	// A verified receipt never authorizes consuming subsequently repopulated
	// keys. Its old intent stays intact; the stopped operator must investigate.
	client = redis.NewClient(redisOptions)
	require.NoError(t, client.Set(ctx, newKey, "synthetic-unexpected-reader", 0).Err())
	require.NoError(t, client.Close())
	invokeHTTP(route, request, http.StatusOK)
	failed = waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "new or changed")
	client = redis.NewClient(redisOptions)
	require.Equal(t, "synthetic-unexpected-reader", client.Get(ctx, newKey).Val())
	require.NoError(t, client.Del(ctx, newKey).Err())
	require.NoError(t, client.Close())
	invokeHTTP(route, request, http.StatusOK)
	completed = waitForIndependentRecoveryTask(t, manager, taskID)
	require.Equal(t, "product_module_switched_still_isolated", completed.State, completed.Error)
	client = redis.NewClient(redisOptions)
}

func sortedSyntheticStoppedProductCacheKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func interruptStoppedProductCacheIntentThenVerifyExecutorRestartCannotContinue(t *testing.T, manager *Manager, task *Task) {
	t.Helper()
	require.NoError(t, manager.startTaskOperation(task, "synthetic_cache_intent_before_executor_restart", func(ctx context.Context, task *Task) error {
		return manager.invalidateProductDetailCacheWithinContinuouslyStoppedSwitch(ctx, task, "synthetic-cache-restart-operator", func(point string) error {
			if point == "cache_intent_persisted" {
				return errors.New("synthetic cache intent before executor restart")
			}
			return nil
		})
	}))
	failed := waitForIndependentRecoveryTask(t, manager, task.ID)
	require.Equal(t, "failed", failed.State)
	require.Contains(t, failed.Error, "before executor restart")
	require.NotNil(t, failed.ProductCatalogStoppedDetailCacheInvalidation)
}

func TestStoppedProductDetailCacheHTTPRejectsCallerChosenKeysAndRedisAddresses(t *testing.T) {
	manager := &Manager{Configuration: createIndependentRecoveryTestConfiguration(t)}
	key, err := readProtectedCredential(manager.Configuration.AuthenticationKeyFile)
	require.NoError(t, err)
	token, err := IssueMaintenanceSession(key, manager.Configuration.SiteID, manager.Configuration.Environment, "synthetic-cache-http-operator", []string{PermissionRestore})
	require.NoError(t, err)
	for _, body := range []string{`{"confirmation":"synthetic","keys":["product:view_counts"]}`, `{"confirmation":"synthetic","redis_url":"redis://other.invalid"}`} {
		request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/tasks/"+strings.Repeat("a", 32)+"/product-catalog-stopped-detail-cache-invalidation", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		manager.HTTPHandler().ServeHTTP(response, request)
		require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
	}
}

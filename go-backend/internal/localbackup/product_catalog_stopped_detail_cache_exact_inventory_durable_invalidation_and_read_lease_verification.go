package localbackup

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/service"
)

const maximumStoppedProductDetailCacheKeys = 10000
const maximumStoppedProductDetailCacheInventoryKeyBytes = 1024 * 1024

type ProductCatalogStoppedDetailCacheKeyObservation struct {
	Key       string `json:"key"`
	ReadLease bool   `json:"read_lease"`
	ValueSHA1 string `json:"value_sha1"`
}

// Kept outside RecoveryJournal: this receipt binds the completed switch digest
// without changing that digest or the immutable preparation/preview bindings.
// Missing keys on retry can have expired or been deleted before a lost reply.
// New or changed keys are never adopted into an existing deletion intent.
type ProductCatalogStoppedDetailCacheInvalidationProgress struct {
	Version                   int                                              `json:"version"`
	TaskID                    string                                           `json:"task_id"`
	WindowID                  string                                           `json:"window_id"`
	SwitchJournalSHA256       string                                           `json:"switch_journal_sha256"`
	PostSwitchContractSHA256  string                                           `json:"post_switch_contract_sha256"`
	RedisOwnershipSHA256      string                                           `json:"redis_ownership_sha256"`
	DetailCacheContractSHA256 string                                           `json:"detail_cache_contract_sha256"`
	Operator                  string                                           `json:"operator"`
	Phase                     string                                           `json:"phase"`
	StartedAt                 time.Time                                        `json:"started_at"`
	Keys                      []ProductCatalogStoppedDetailCacheKeyObservation `json:"keys"`
	PlannedDetailCacheKeys    int                                              `json:"planned_detail_cache_keys"`
	PlannedReadLeaseKeys      int                                              `json:"planned_read_lease_keys"`
	VerifiedAbsentKeys        int                                              `json:"verified_absent_keys"`
	VerifiedAt                time.Time                                        `json:"verified_at,omitempty"`
	SHA256                    string                                           `json:"sha256"`
}

// Every batch validates all observations before deleting any key. Only exact
// keys, recognized again in Go, are supplied. These guards never mutate buffers.
var deleteStoppedProductDetailCacheKeysOnlyIfObservationsStillMatch = redis.NewScript(`
if redis.call('TYPE', KEYS[1]).ok ~= 'string' or redis.call('GET', KEYS[1]) ~= ARGV[1] or redis.call('PTTL', KEYS[1]) ~= -1 then
  return redis.error_reply('stopped product cache ownership changed')
end
for i=2,4 do
  if redis.call('EXISTS', KEYS[i]) ~= 0 then return redis.error_reply('stopped product view buffers changed') end
end
for i=5,#KEYS do
  local kind = redis.call('TYPE', KEYS[i]).ok
  if kind ~= 'none' and (kind ~= 'string' or redis.sha1hex(redis.call('GET', KEYS[i])) ~= ARGV[i-3]) then
    return redis.error_reply('stopped product cache observation changed')
  end
end
local removed = 0
for i=5,#KEYS do removed = removed + redis.call('DEL', KEYS[i]) end
return removed
`)

var inspectStoppedProductDetailCacheStringValueDigest = redis.NewScript(`
local kind = redis.call('TYPE', KEYS[1]).ok
if kind == 'none' then return false end
if kind ~= 'string' or redis.call('STRLEN', KEYS[1]) > 16777216 then
  return redis.error_reply('unexpected product cache type or oversized value')
end
return redis.sha1hex(redis.call('GET', KEYS[1]))
`)

func (manager *Manager) saveStoppedProductDetailCacheInvalidationProgress(task *Task) error {
	progress := task.ProductCatalogStoppedDetailCacheInvalidation
	snapshotCopy := *progress
	snapshotCopy.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(snapshotCopy)
	if err != nil {
		return err
	}
	progress.SHA256 = digest
	return manager.saveTask(task)
}

func validateStoppedProductDetailCacheInvalidationBinding(task *Task, report *ProductCatalogStoppedPostSwitchInspection) error {
	progress := task.ProductCatalogStoppedDetailCacheInvalidation
	if progress == nil {
		return nil
	}
	snapshotCopy := *progress
	snapshotCopy.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(snapshotCopy)
	contract := service.DescribeProductCatalogRuntimeProjectionBoundaryForRecovery(true)
	if err != nil || digest != progress.SHA256 || progress.Version != 1 || progress.TaskID != task.ID || progress.WindowID != report.WindowID || progress.SwitchJournalSHA256 != report.SwitchJournalSHA256 || progress.PostSwitchContractSHA256 != report.ContractSHA256 || progress.RedisOwnershipSHA256 != report.RedisOwnershipSHA256 || progress.DetailCacheContractSHA256 != contract.DetailCacheContractSHA256 || strings.TrimSpace(progress.Operator) == "" || progress.StartedAt.IsZero() || (progress.Phase != "deleting" && progress.Phase != "verified_absent") {
		return errors.New("stopped product cache intent identity or digest changed; retain isolation")
	}
	if len(progress.Keys) > maximumStoppedProductDetailCacheKeys {
		return errors.New("stopped product cache inventory exceeds its supported bound")
	}
	detailCount, leaseCount, keyBytes, previous := 0, 0, 0, ""
	for _, observation := range progress.Keys {
		detail, lease := service.RecognizeProductDetailCacheAndReadLeaseKeyForStoppedRecovery(observation.Key)
		valueDigest, err := hex.DecodeString(observation.ValueSHA1)
		if (!detail && !lease) || lease != observation.ReadLease || observation.Key <= previous || err != nil || len(valueDigest) != sha1.Size || hex.EncodeToString(valueDigest) != observation.ValueSHA1 {
			return errors.New("stopped product cache inventory contains an unrecognized or altered observation")
		}
		previous = observation.Key
		keyBytes += len(observation.Key)
		if lease {
			leaseCount++
		} else {
			detailCount++
		}
	}
	if keyBytes > maximumStoppedProductDetailCacheInventoryKeyBytes || detailCount != progress.PlannedDetailCacheKeys || leaseCount != progress.PlannedReadLeaseKeys || (progress.Phase == "deleting" && (!progress.VerifiedAt.IsZero() || progress.VerifiedAbsentKeys != 0)) || (progress.Phase == "verified_absent" && (progress.VerifiedAt.IsZero() || progress.VerifiedAbsentKeys != len(progress.Keys))) {
		return errors.New("stopped product cache inventory counts or completion changed")
	}
	return nil
}

func (manager *Manager) StartStoppedProductDetailCacheInvalidation(id, actor string, request ProductCatalogStoppedSwitchConfirmationRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if request.Confirmation != "INVALIDATE STOPPED PRODUCT CACHE "+manager.Configuration.SiteID+"/"+manager.Configuration.Environment+"/"+id || strings.TrimSpace(actor) == "" {
		return nil, errors.New("explicit stopped product detail cache invalidation confirmation is required")
	}
	if task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) || (task.State != "product_module_switched_still_isolated" && task.State != "failed") {
		return nil, errors.New("an unexpired switched product result in this continuously stopped executor is required")
	}
	if err := manager.validateProductCatalogStoppedSwitchJournalBinding(task); err != nil {
		return nil, err
	}
	if task.Journal.ProductCatalogStoppedSwitch.Phase != "switched_still_isolated" {
		return nil, errors.New("cache invalidation requires the completed isolated product switch")
	}
	if err := manager.startTaskOperation(task, "invalidating_stopped_product_detail_cache_and_read_leases", func(ctx context.Context, task *Task) error {
		return manager.invalidateProductDetailCacheWithinContinuouslyStoppedSwitch(ctx, task, actor, nil)
	}); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

// checkpoint is solely for testing acknowledgement loss; it is not HTTP input.
func (manager *Manager) invalidateProductDetailCacheWithinContinuouslyStoppedSwitch(ctx context.Context, task *Task, actor string, checkpoint func(string) error) error {
	if task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("product package expired before cache processing")
	}
	report, err := manager.inspectStoppedProductCatalogPostSwitchWithHeldLease(ctx, task)
	if err != nil {
		return err
	}
	if err := validateStoppedProductDetailCacheInvalidationBinding(task, report); err != nil {
		return err
	}
	client, ownership, err := manager.connectAndVerifyRuntimeOwnedProductViewRedis(ctx)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(client.Close)
	if err := manager.verifyStoppedProductDetailCacheRedisBoundary(ctx, task, client, ownership); err != nil {
		return err
	}
	observations, err := inspectExactStoppedProductDetailCacheAndReadLeaseKeys(ctx, client)
	if err != nil {
		return err
	}
	progress := task.ProductCatalogStoppedDetailCacheInvalidation
	if progress == nil {
		progress = &ProductCatalogStoppedDetailCacheInvalidationProgress{Version: 1, TaskID: task.ID, WindowID: report.WindowID, SwitchJournalSHA256: report.SwitchJournalSHA256, PostSwitchContractSHA256: report.ContractSHA256, RedisOwnershipSHA256: report.RedisOwnershipSHA256, DetailCacheContractSHA256: service.DescribeProductCatalogRuntimeProjectionBoundaryForRecovery(true).DetailCacheContractSHA256, Operator: actor, Phase: "deleting", StartedAt: time.Now().UTC(), Keys: observations}
		for _, observation := range observations {
			if observation.ReadLease {
				progress.PlannedReadLeaseKeys++
			} else {
				progress.PlannedDetailCacheKeys++
			}
		}
		task.ProductCatalogStoppedDetailCacheInvalidation = progress
		if err := manager.saveStoppedProductDetailCacheInvalidationProgress(task); err != nil {
			return err
		}
		if checkpoint != nil {
			if err := checkpoint("cache_intent_persisted"); err != nil {
				return err
			}
		}
	}
	planned := make(map[string]ProductCatalogStoppedDetailCacheKeyObservation, len(progress.Keys))
	for _, observation := range progress.Keys {
		planned[observation.Key] = observation
	}
	for _, observation := range observations {
		if expected, exists := planned[observation.Key]; !exists || expected != observation || progress.Phase == "verified_absent" {
			return errors.New("new or changed product cache key appeared after durable intent; retain isolation without adopting it")
		}
	}
	for start := 0; start < len(progress.Keys); start += 128 {
		if err := manager.verifyStoppedProductDetailCacheRedisBoundary(ctx, task, client, ownership); err != nil {
			return err
		}
		marker, err := client.Get(ctx, maintenancegate.ProductViewRedisOwnershipKey).Result()
		if err != nil || maintenancegate.FingerprintExternalProjectionConfiguration(marker) != ownership.MarkerSHA256 {
			return errors.New("stopped product cache ownership marker changed")
		}
		keys := []string{maintenancegate.ProductViewRedisOwnershipKey, "product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id"}
		arguments := []interface{}{marker}
		end := start + 128
		if end > len(progress.Keys) {
			end = len(progress.Keys)
		}
		for _, observation := range progress.Keys[start:end] {
			detail, lease := service.RecognizeProductDetailCacheAndReadLeaseKeyForStoppedRecovery(observation.Key)
			if !detail && !lease {
				return errors.New("refusing a non-detail-cache key in the invalidation batch")
			}
			keys = append(keys, observation.Key)
			arguments = append(arguments, observation.ValueSHA1)
		}
		if err := deleteStoppedProductDetailCacheKeysOnlyIfObservationsStillMatch.Run(ctx, client, keys, arguments...).Err(); err != nil {
			return errors.New("stopped product cache batch was not acknowledged; retain durable intent for verified retry")
		}
		if checkpoint != nil {
			if err := checkpoint("cache_batch_deleted_before_acknowledgement"); err != nil {
				return err
			}
		}
	}
	remaining, err := inspectExactStoppedProductDetailCacheAndReadLeaseKeys(ctx, client)
	if err != nil || len(remaining) != 0 {
		return errors.New("product cache or read lease appeared during invalidation; keep application stopped")
	}
	if err := manager.verifyStoppedProductDetailCacheRedisBoundary(ctx, task, client, ownership); err != nil {
		return err
	}
	// Full SQL/media inspection uses its own one-socket Redis client.
	if err := client.Close(); err != nil {
		return err
	}
	finalReport, err := manager.inspectStoppedProductCatalogPostSwitchWithHeldLease(ctx, task)
	if err != nil || finalReport.ContractSHA256 != report.ContractSHA256 {
		return errors.New("product switch boundary changed after cache invalidation; retain isolation")
	}
	finalClient, finalOwnership, err := manager.connectAndVerifyRuntimeOwnedProductViewRedis(ctx)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(finalClient.Close)
	if err := manager.verifyStoppedProductDetailCacheRedisBoundary(ctx, task, finalClient, finalOwnership); err != nil {
		return err
	}
	remaining, err = inspectExactStoppedProductDetailCacheAndReadLeaseKeys(ctx, finalClient)
	if err != nil || len(remaining) != 0 {
		return errors.New("product detail cache changed during final SQL/media inspection; retain isolation")
	}
	progress.Phase, progress.VerifiedAbsentKeys = "verified_absent", len(progress.Keys)
	progress.VerifiedAt = time.Now().UTC()
	task.State, task.Phase = "product_module_switched_still_isolated", "product_detail_cache_and_read_leases_verified_absent_still_isolated"
	return manager.saveStoppedProductDetailCacheInvalidationProgress(task)
}

func (manager *Manager) verifyStoppedProductDetailCacheRedisBoundary(ctx context.Context, task *Task, client redis.UniversalClient, ownership *maintenancegate.ProductViewRedisRuntimeOwnership) error {
	if task.WorkCopyExpired || !time.Now().Before(task.ExpiresAt) {
		return errors.New("product package expired during cache processing; retain isolation for rollback")
	}
	if manager.productCatalogStoppedHandoffLease == nil || manager.productCatalogStoppedHandoffLeaseTaskID != task.ID || !manager.productCatalogStoppedHandoffLease.Locked() || task.ProductViewStoppedDrain == nil || maintenancegate.FingerprintExternalProjectionConfiguration(ownership) != task.ProductViewStoppedDrain.OwnershipSHA256 {
		return errors.New("product cache processing lost this executor's continuous stopped lease or Redis ownership")
	}
	if err := manager.verifyProductCatalogStoppedWindowFenceAndRuntimeBoundaryBytes(task); err != nil {
		return err
	}
	if err := maintenancegate.VerifyRegisteredProductViewRedisRuntimeOwnership(ctx, client, ownership); err != nil {
		return err
	}
	if err := requireNoOtherProductViewRedisLogicalDatabaseClients(ctx, client, manager.Configuration.ProductViewRecoveryRedis.Database); err != nil {
		return err
	}
	count, err := client.Exists(ctx, "product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id").Result()
	if err != nil || count != 0 {
		return errors.New("product view buffers changed before detail cache invalidation")
	}
	return nil
}

func inspectExactStoppedProductDetailCacheAndReadLeaseKeys(ctx context.Context, client redis.UniversalClient) ([]ProductCatalogStoppedDetailCacheKeyObservation, error) {
	observations := map[string]ProductCatalogStoppedDetailCacheKeyObservation{}
	var cursor uint64
	keyBytes, scanned := 0, 0
	for {
		keys, next, err := client.Scan(ctx, cursor, "", 256).Result()
		if err != nil {
			return nil, errors.New("cannot enumerate stopped product detail cache keys")
		}
		scanned += len(keys)
		if scanned > 1000000 {
			return nil, errors.New("redis inventory exceeds the supported stopped cache inspection size")
		}
		for _, key := range keys {
			detail, lease := service.RecognizeProductDetailCacheAndReadLeaseKeyForStoppedRecovery(key)
			if !detail && !lease {
				continue
			}
			valueDigest, err := inspectStoppedProductDetailCacheStringValueDigest.Run(ctx, client, []string{key}).Text()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return nil, errors.New("recognized product cache or read lease has an unreadable or unexpected Redis type")
			}
			observation := ProductCatalogStoppedDetailCacheKeyObservation{Key: key, ReadLease: lease, ValueSHA1: valueDigest}
			if previous, exists := observations[key]; exists {
				if previous != observation {
					return nil, errors.New("product cache changed during Redis inventory")
				}
				continue
			}
			keyBytes += len(key)
			observations[key] = observation
			if len(observations) > maximumStoppedProductDetailCacheKeys || keyBytes > maximumStoppedProductDetailCacheInventoryKeyBytes {
				return nil, errors.New("product detail cache inventory exceeds its supported bound; no keys deleted")
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	result := make([]ProductCatalogStoppedDetailCacheKeyObservation, 0, len(observations))
	for _, observation := range observations {
		result = append(result, observation)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].Key < result[right].Key })
	return result, nil
}

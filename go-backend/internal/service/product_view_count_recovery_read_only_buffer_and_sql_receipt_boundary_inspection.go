package service

import (
	"context"
	"errors"
	"math"
	"reflect"

	"commerce-platform/internal/pkg/maintenancegate"
)

// An observation, never proof that all writers stopped or permission to switch.
// Only hashes and aggregate counts are exposed; raw product IDs stay internal.
type ProductViewCountRecoveryBoundaryInspection struct {
	Version                         int    `json:"version"`
	ActiveProducts                  int    `json:"active_products"`
	ActiveDeltaTotal                int64  `json:"active_delta_total"`
	ProcessingProducts              int    `json:"processing_products"`
	ProcessingDeltaTotal            int64  `json:"processing_delta_total"`
	CommittedProcessingReceipts     int    `json:"committed_processing_receipts"`
	CommittedMissingProductReceipts int    `json:"committed_missing_product_receipts"`
	PendingProcessingProducts       int    `json:"pending_processing_products"`
	RedisBuffersEmpty               bool   `json:"redis_buffers_empty"`
	RedisSnapshotSHA256             string `json:"redis_snapshot_sha256"`
	SQLReceiptsSHA256               string `json:"sql_receipts_sha256"`
	StoppedRevalidationRequired     bool   `json:"stopped_revalidation_required"`
	ReleaseAvailable                bool   `json:"release_available"`
}

func sumPositiveProductViewDeltasWithoutOverflow(counts map[uint]int64) (int64, error) {
	var total int64
	for _, delta := range counts {
		if delta <= 0 || total > math.MaxInt64-delta {
			return 0, errors.New("product view recovery aggregate exceeds capacity; preserve all buffered state")
		}
		total += delta
	}
	return total, nil
}

func (b *ProductViewCountBuffer) InspectReadOnlyProductViewRecoveryBoundary(ctx context.Context) (*ProductViewCountRecoveryBoundaryInspection, error) {
	if b == nil || b.client == nil || b.repo == nil {
		return nil, errors.New("product view recovery inspection requires the configured buffer and SQL repository")
	}
	keys := productViewCountActiveProcessingAndBatchIdentityRedisKeys()
	snapshot, err := productViewCountReadOnlyRecoverySnapshotScript.Run(ctx, b.client, keys, 5000).Slice()
	if err != nil {
		return nil, err
	}
	if len(snapshot) != 3 {
		return nil, errors.New("product view recovery snapshot shape is invalid")
	}
	batchID, ok := snapshot[0].(string)
	if !ok {
		return nil, errors.New("product view recovery snapshot batch identity is invalid")
	}
	active, err := parseCanonicalProductViewCountRedisHashEntries(snapshot[1])
	if err != nil {
		return nil, err
	}
	processing, err := parseCanonicalProductViewCountRedisHashEntries(snapshot[2])
	if err != nil {
		return nil, err
	}
	report := &ProductViewCountRecoveryBoundaryInspection{Version: 1, ActiveProducts: len(active), ProcessingProducts: len(processing), StoppedRevalidationRequired: true}
	report.ActiveDeltaTotal, err = sumPositiveProductViewDeltasWithoutOverflow(active)
	if err != nil {
		return nil, err
	}
	report.ProcessingDeltaTotal, err = sumPositiveProductViewDeltasWithoutOverflow(processing)
	if err != nil {
		return nil, err
	}
	// Verify the store even for an empty buffer: missing migration is never a
	// successful receipt inspection. A nil counts set just reads the table schema.
	receipts, err := b.repo.InspectProductViewCountBatchReplayReceipts(ctx, batchID, processing)
	if err != nil {
		return nil, err
	}
	report.CommittedProcessingReceipts = len(receipts)
	for _, receipt := range receipts {
		if !receipt.ProductIncrementApplied {
			report.CommittedMissingProductReceipts++
		}
	}
	report.PendingProcessingProducts = len(processing) - len(receipts)
	latestReceipts, err := b.repo.InspectProductViewCountBatchReplayReceipts(ctx, batchID, processing)
	if err != nil || !reflect.DeepEqual(receipts, latestReceipts) {
		return nil, errors.New("product view SQL receipts changed during recovery inspection")
	}
	latest, err := productViewCountReadOnlyRecoverySnapshotScript.Run(ctx, b.client, keys, 5000).Slice()
	if err != nil {
		return nil, err
	}
	// Hash canonical maps; HGETALL order is not an identity contract.
	if len(latest) != 3 {
		return nil, errors.New("product view recovery snapshot shape changed during inspection")
	}
	latestActive, parseError := parseCanonicalProductViewCountRedisHashEntries(latest[1])
	if parseError != nil {
		return nil, parseError
	}
	latestProcessing, parseError := parseCanonicalProductViewCountRedisHashEntries(latest[2])
	if parseError != nil {
		return nil, parseError
	}
	if latest[0] != batchID || !reflect.DeepEqual(active, latestActive) || !reflect.DeepEqual(processing, latestProcessing) {
		return nil, errors.New("product view Redis state changed during recovery inspection")
	}
	report.RedisBuffersEmpty = len(active) == 0 && len(processing) == 0
	report.RedisSnapshotSHA256 = maintenancegate.FingerprintExternalProjectionConfiguration(struct {
		BatchID            string
		Active, Processing map[uint]int64
	}{batchID, active, processing})
	report.SQLReceiptsSHA256 = maintenancegate.FingerprintExternalProjectionConfiguration(receipts)
	return report, nil
}

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"
)

var ErrOrderEvidenceStoreUnavailable = errors.New("order evidence store is unavailable")

// OrderEvidenceService creates the initial evidence plan from the immutable
// order snapshot. It does not upload files or decide payment-provider fields.
type OrderEvidenceService struct{}

func NewOrderEvidenceService() *OrderEvidenceService {
	return &OrderEvidenceService{}
}

func (s *OrderEvidenceService) CreateInitialPackage(
	repos repository.TxRepositories,
	snapshot *orderevidence.OrderEvidenceSnapshot,
) (*orderevidence.OrderEvidencePackage, error) {
	if s == nil || repos.OrderEvidence == nil {
		return nil, ErrOrderEvidenceStoreUnavailable
	}
	if snapshot == nil {
		return nil, errors.New("order evidence snapshot is required")
	}
	if snapshot.ID == 0 {
		return nil, errors.New("order evidence snapshot id is required")
	}

	payload, err := orderevidence.ParseOrderEvidenceSnapshotPayload(snapshot)
	if err != nil {
		return nil, err
	}
	now := snapshot.ConfirmedAt.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	pkg := &orderevidence.OrderEvidencePackage{
		OrderID:               snapshot.OrderID,
		SnapshotID:            snapshot.ID,
		PackageVersion:        1,
		Status:                orderevidence.PackageStatusIncomplete,
		OrderTotalUSDSnapshot: snapshot.OrderTotalUSD,
		IsHighValue:           snapshot.IsHighValue,
		HasSpokeTensionQC:     snapshot.HasSpokeTensionQC,
		SchemaVersion:         orderevidence.OrderEvidencePackageSchemaVersion,
	}
	items := buildInitialEvidenceItems(snapshot, payload, now)
	if err := repos.OrderEvidence.CreatePackageWithItems(pkg, items); err != nil {
		return nil, fmt.Errorf("save initial order evidence package: %w", err)
	}
	return pkg, nil
}

func buildInitialEvidenceItems(
	snapshot *orderevidence.OrderEvidenceSnapshot,
	payload orderevidence.OrderEvidenceSnapshotPayload,
	capturedAt time.Time,
) []orderevidence.OrderEvidenceItem {
	items := make([]orderevidence.OrderEvidenceItem, 0, len(payload.Items)*3+2)
	for _, payloadItem := range payload.Items {
		orderItemID := payloadItem.OrderItemID
		itemData, itemHash := evidenceItemDataAndHash(payloadItem)
		snapshotID := snapshot.ID
		items = append(items,
			orderevidence.OrderEvidenceItem{
				OrderID:        snapshot.OrderID,
				OrderItemID:    &orderItemID,
				SnapshotID:     &snapshotID,
				ItemType:       orderevidence.EvidenceItemTypeConfigurationConfirmation,
				Status:         orderevidence.EvidenceItemStatusComplete,
				RequiredReason: orderevidence.EvidenceRequiredReasonBase,
				DataJSON:       itemData,
				CapturedAt:     orderEvidenceTimePtr(capturedAt),
				ContentSHA256:  itemHash,
			},
			orderevidence.OrderEvidenceItem{
				OrderID:        snapshot.OrderID,
				OrderItemID:    &orderItemID,
				ItemType:       orderevidence.EvidenceItemTypeProductIdentity,
				Status:         orderevidence.EvidenceItemStatusMissing,
				RequiredReason: orderevidence.EvidenceRequiredReasonBase,
			},
		)
		if payloadItem.ProductRequirementSnapshot.Required {
			items = append(items, orderevidence.OrderEvidenceItem{
				OrderID:        snapshot.OrderID,
				OrderItemID:    &orderItemID,
				ItemType:       orderevidence.EvidenceItemTypeSpokeQCTension,
				Status:         orderevidence.EvidenceItemStatusMissing,
				RequiredReason: orderevidence.EvidenceRequiredReasonSpokeTensionQC,
			})
		}
	}

	items = append(items,
		orderevidence.OrderEvidenceItem{
			OrderID:        snapshot.OrderID,
			ItemType:       orderevidence.EvidenceItemTypeOutboundWeightPackaging,
			Status:         orderevidence.EvidenceItemStatusMissing,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
		},
		orderevidence.OrderEvidenceItem{
			OrderID:        snapshot.OrderID,
			ItemType:       orderevidence.EvidenceItemTypeSignedPOD,
			Status:         orderevidence.EvidenceItemStatusMissing,
			RequiredReason: orderevidence.EvidenceRequiredReasonBase,
		},
	)
	return items
}

func evidenceItemDataAndHash(value orderevidence.OrderEvidenceSnapshotItem) ([]byte, string) {
	data, err := json.Marshal(value)
	if err != nil {
		return []byte("{}"), ""
	}
	hash := sha256.Sum256(data)
	return data, hex.EncodeToString(hash[:])
}

func orderEvidenceTimePtr(value time.Time) *time.Time {
	value = value.UTC()
	return &value
}

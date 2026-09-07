package service

import (
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	productrequirement "commerce-platform/internal/domain/productrequirement"
	"commerce-platform/internal/repository"
)

var (
	ErrOrderEvidenceSnapshotStoreUnavailable       = errors.New("order evidence snapshot store is unavailable")
	ErrOrderEvidenceSnapshotRequirementUnavailable = errors.New("order evidence requirement store is unavailable")
)

// OrderEvidenceSnapshotService creates the one-time order-time snapshot. It
// deliberately has no attachment, upload, dispute, or admin-page concerns.
type OrderEvidenceSnapshotService struct{}

func NewOrderEvidenceSnapshotService() *OrderEvidenceSnapshotService {
	return &OrderEvidenceSnapshotService{}
}

func (s *OrderEvidenceSnapshotService) CreateForOrder(
	repos repository.TxRepositories,
	orderRecord *order.Order,
) (*orderevidence.OrderEvidenceSnapshot, error) {
	if s == nil || repos.OrderEvidenceSnapshot == nil {
		return nil, ErrOrderEvidenceSnapshotStoreUnavailable
	}
	if repos.ProductQualityRequirement == nil {
		return nil, ErrOrderEvidenceSnapshotRequirementUnavailable
	}
	if orderRecord == nil {
		return nil, errors.New("order evidence snapshot order is required")
	}

	items := make([]orderevidence.SnapshotItemInput, 0, len(orderRecord.Items))
	for _, item := range orderRecord.Items {
		resolution, err := resolveProductRequirementInTx(
			repos.ProductQualityRequirement,
			item.ProductID,
			item.VariantID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"resolve product requirement for order item %d: %w",
				item.ID,
				err,
			)
		}
		items = append(items, orderevidence.SnapshotItemInput{
			Item:               item,
			ProductRequirement: resolution,
		})
	}

	snapshot, err := orderevidence.BuildOrderEvidenceSnapshot(orderRecord, items, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := repos.OrderEvidenceSnapshot.Create(snapshot); err != nil {
		return nil, fmt.Errorf("save order evidence snapshot: %w", err)
	}
	return snapshot, nil
}

func resolveProductRequirementInTx(
	repo *repository.ProductQualityRequirementRepository,
	productID uint,
	variantID *uint,
) (productrequirement.SpokeTensionQCResolution, error) {
	if productID == 0 {
		return productrequirement.SpokeTensionQCResolution{}, errors.New("product_id is required")
	}
	rules, err := repo.ListByProduct(productID)
	if err != nil {
		return productrequirement.SpokeTensionQCResolution{}, err
	}
	return productrequirement.ResolveSpokeTensionQC(productID, variantID, rules)
}

package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/repository"
)

var (
	ErrOrderFulfillmentEvidenceNotConfigured  = errors.New("order fulfillment evidence is not configured")
	ErrOrderFulfillmentEvidencePackageMissing = errors.New("order fulfillment evidence package is missing")
	ErrOrderFulfillmentEvidenceIncomplete     = errors.New("order fulfillment evidence is incomplete")
)

type OrderFulfillmentEvidenceMissing struct {
	ItemID          uint   `json:"item_id,omitempty"`
	OrderItemID     *uint  `json:"order_item_id,omitempty"`
	ItemType        string `json:"item_type"`
	Status          string `json:"status,omitempty"`
	AttachmentCount int    `json:"attachment_count"`
	Reason          string `json:"reason"`
}

type OrderFulfillmentEvidenceCheck struct {
	Ready    bool                              `json:"ready"`
	Total    int                               `json:"total"`
	Complete int                               `json:"complete"`
	Pending  int                               `json:"pending"`
	Missing  []OrderFulfillmentEvidenceMissing `json:"missing,omitempty"`
}

type OrderFulfillmentEvidenceError struct {
	Check OrderFulfillmentEvidenceCheck
}

func (e *OrderFulfillmentEvidenceError) Error() string {
	if e == nil {
		return ErrOrderFulfillmentEvidenceIncomplete.Error()
	}
	if len(e.Check.Missing) == 0 {
		return ErrOrderFulfillmentEvidenceIncomplete.Error()
	}

	labels := make([]string, 0, len(e.Check.Missing))
	for _, missing := range e.Check.Missing {
		label := strings.TrimSpace(missing.ItemType)
		if missing.OrderItemID != nil && *missing.OrderItemID > 0 {
			label = fmt.Sprintf("%s(order item %d)", label, *missing.OrderItemID)
		}
		if label != "" {
			labels = append(labels, label)
		}
	}
	if len(labels) == 0 {
		return ErrOrderFulfillmentEvidenceIncomplete.Error()
	}
	return fmt.Sprintf("%s: %s", ErrOrderFulfillmentEvidenceIncomplete, strings.Join(labels, ", "))
}

func (e *OrderFulfillmentEvidenceError) Unwrap() error {
	return ErrOrderFulfillmentEvidenceIncomplete
}

func (s *OrderEvidenceService) CheckFulfillmentReadiness(
	repos repository.TxRepositories,
	orderRecord *order.Order,
) (*OrderFulfillmentEvidenceCheck, error) {
	if s == nil || repos.OrderEvidence == nil {
		return nil, ErrOrderFulfillmentEvidenceNotConfigured
	}
	if orderRecord == nil || orderRecord.ID == 0 {
		return nil, errors.New("order is required for fulfillment evidence check")
	}

	pkg, err := repos.OrderEvidence.FindLatestPackageByOrderID(orderRecord.ID)
	if repository.IsRecordNotFound(err) {
		return nil, ErrOrderFulfillmentEvidencePackageMissing
	}
	if err != nil {
		return nil, err
	}
	if pkg.IsSuperseded() {
		return nil, ErrOrderFulfillmentEvidencePackageMissing
	}

	requiredTensionOrderItems, err := requiredSpokeTensionOrderItems(repos, pkg)
	if err != nil {
		return nil, err
	}
	check := evaluateOrderFulfillmentEvidence(pkg.Items, orderRecord.Items, requiredTensionOrderItems)
	if !check.Ready {
		return &check, &OrderFulfillmentEvidenceError{Check: check}
	}
	return &check, nil
}

func evaluateOrderFulfillmentEvidence(
	items []orderevidence.OrderEvidenceItem,
	orderItems []order.OrderItem,
	requiredTensionOrderItems map[uint]bool,
) OrderFulfillmentEvidenceCheck {
	check := OrderFulfillmentEvidenceCheck{}
	configurationByOrderItem := make(map[uint]orderevidence.OrderEvidenceItem)
	identityByOrderItem := make(map[uint]orderevidence.OrderEvidenceItem)
	tensionByOrderItem := make(map[uint]orderevidence.OrderEvidenceItem)
	hasOutboundEvidence := false

	for _, item := range items {
		if item.ItemType == orderevidence.EvidenceItemTypeSignedPOD {
			continue
		}
		if item.ItemType == orderevidence.EvidenceItemTypeSpokeQCTension {
			if item.OrderItemID == nil || !requiredTensionOrderItems[*item.OrderItemID] {
				continue
			}
		}

		check.Total++
		if item.Status == orderevidence.EvidenceItemStatusComplete {
			check.Complete++
		} else {
			check.Pending++
		}

		switch item.ItemType {
		case orderevidence.EvidenceItemTypeConfigurationConfirmation:
			if item.OrderItemID != nil {
				configurationByOrderItem[*item.OrderItemID] = item
			}
		case orderevidence.EvidenceItemTypeProductIdentity:
			if item.OrderItemID != nil {
				identityByOrderItem[*item.OrderItemID] = item
			}
		case orderevidence.EvidenceItemTypeSpokeQCTension:
			if item.OrderItemID != nil {
				tensionByOrderItem[*item.OrderItemID] = item
			}
		case orderevidence.EvidenceItemTypeOutboundWeightPackaging:
			hasOutboundEvidence = true
		}

		if item.ItemType == orderevidence.EvidenceItemTypeConfigurationConfirmation {
			if item.Status != orderevidence.EvidenceItemStatusComplete {
				check.Missing = append(check.Missing, missingEvidence(item, "configuration confirmation must be complete"))
			}
			continue
		}
		if item.ItemType == orderevidence.EvidenceItemTypeProductIdentity {
			if item.OrderItemID == nil || *item.OrderItemID == 0 {
				check.Missing = append(check.Missing, missingEvidence(item, "product identity evidence must belong to an order item"))
			} else if item.Status != orderevidence.EvidenceItemStatusComplete {
				check.Missing = append(check.Missing, missingEvidence(item, "product identity evidence must be complete"))
			} else if len(item.Attachments) == 0 {
				check.Missing = append(check.Missing, missingEvidence(item, "product identity evidence requires at least one attachment"))
			}
			continue
		}
		if item.Status != orderevidence.EvidenceItemStatusComplete {
			check.Missing = append(check.Missing, missingEvidence(item, "evidence must be completed before dispatch"))
			continue
		}
		if len(item.Attachments) == 0 {
			check.Missing = append(check.Missing, missingEvidence(item, "at least one attachment is required before dispatch"))
		}
	}

	if !hasOutboundEvidence {
		check.Total++
		check.Pending++
		check.Missing = append(check.Missing, OrderFulfillmentEvidenceMissing{
			ItemType: orderevidence.EvidenceItemTypeOutboundWeightPackaging,
			Reason:   "outbound weight and packaging evidence is missing",
		})
	}
	for _, orderItem := range orderItems {
		if orderItem.ID == 0 {
			continue
		}

		_, ok := configurationByOrderItem[orderItem.ID]
		if !ok {
			check.Total++
			check.Pending++
			check.Missing = append(check.Missing, OrderFulfillmentEvidenceMissing{
				OrderItemID: &orderItem.ID,
				ItemType:    orderevidence.EvidenceItemTypeConfigurationConfirmation,
				Reason:      "configuration confirmation is missing for the order item",
			})
		}

		_, ok = identityByOrderItem[orderItem.ID]
		if !ok {
			check.Total++
			check.Pending++
			check.Missing = append(check.Missing, OrderFulfillmentEvidenceMissing{
				OrderItemID: &orderItem.ID,
				ItemType:    orderevidence.EvidenceItemTypeProductIdentity,
				Reason:      "product identity evidence is missing for the order item",
			})
			continue
		}
		if requiredTensionOrderItems[orderItem.ID] {
			tension, ok := tensionByOrderItem[orderItem.ID]
			if !ok {
				check.Total++
				check.Pending++
				check.Missing = append(check.Missing, OrderFulfillmentEvidenceMissing{
					OrderItemID: &orderItem.ID,
					ItemType:    orderevidence.EvidenceItemTypeSpokeQCTension,
					Reason:      "spoke tension evidence is missing for the order item",
				})
			} else if tension.Status != orderevidence.EvidenceItemStatusComplete || len(tension.Attachments) == 0 {
				check.Missing = append(check.Missing, missingEvidence(tension, "spoke tension evidence needs a complete record and at least one attachment"))
			}
		}
	}
	for requiredOrderItemID := range requiredTensionOrderItems {
		if _, ok := tensionByOrderItem[requiredOrderItemID]; ok {
			continue
		}
		foundOrderItem := false
		for _, orderItem := range orderItems {
			if orderItem.ID == requiredOrderItemID {
				foundOrderItem = true
				break
			}
		}
		if !foundOrderItem {
			check.Total++
			check.Pending++
			check.Missing = append(check.Missing, OrderFulfillmentEvidenceMissing{
				OrderItemID: &requiredOrderItemID,
				ItemType:    orderevidence.EvidenceItemTypeSpokeQCTension,
				Reason:      "spoke tension evidence belongs to an order item that is not present",
			})
		}
	}

	check.Ready = len(check.Missing) == 0
	return check
}

func requiredSpokeTensionOrderItems(
	repos repository.TxRepositories,
	pkg *orderevidence.OrderEvidencePackage,
) (map[uint]bool, error) {
	required := make(map[uint]bool)
	if pkg == nil || !pkg.HasSpokeTensionQC {
		return required, nil
	}
	if repos.OrderEvidenceSnapshot == nil {
		return nil, ErrOrderFulfillmentEvidenceNotConfigured
	}

	snapshot, err := repos.OrderEvidenceSnapshot.FindByOrderID(pkg.OrderID)
	if repository.IsRecordNotFound(err) {
		return nil, ErrOrderFulfillmentEvidencePackageMissing
	}
	if err != nil {
		return nil, err
	}
	if snapshot.ID != pkg.SnapshotID {
		return nil, ErrOrderFulfillmentEvidencePackageMissing
	}
	payload, err := orderevidence.ParseOrderEvidenceSnapshotPayload(snapshot)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid order evidence snapshot: %v", ErrOrderFulfillmentEvidencePackageMissing, err)
	}
	for _, item := range payload.Items {
		if item.ProductRequirementSnapshot.Required && item.OrderItemID > 0 {
			required[item.OrderItemID] = true
		}
	}
	if len(required) == 0 {
		return nil, fmt.Errorf("%w: package declares spoke tension evidence but snapshot has no required order item", ErrOrderFulfillmentEvidencePackageMissing)
	}
	return required, nil
}

func missingEvidence(
	item orderevidence.OrderEvidenceItem,
	reason string,
) OrderFulfillmentEvidenceMissing {
	return OrderFulfillmentEvidenceMissing{
		ItemID:          item.ID,
		OrderItemID:     item.OrderItemID,
		ItemType:        item.ItemType,
		Status:          item.Status,
		AttachmentCount: len(item.Attachments),
		Reason:          reason,
	}
}

package service

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

var (
	ErrOrderEvidenceAssemblyUnavailable = errors.New("order evidence package assembler is unavailable")
	ErrOrderEvidenceAssemblyOrderID     = errors.New("order evidence package assembly order id is required")
)

// OrderEvidenceSourceReference records which durable facts were included in a
// compiled view. It is metadata only; the source records remain authoritative.
type OrderEvidenceSourceReference struct {
	SourceType       string     `json:"source_type"`
	SourceID         uint       `json:"source_id"`
	OrderItemID      *uint      `json:"order_item_id,omitempty"`
	EvidenceItemType string     `json:"evidence_item_type,omitempty"`
	Status           string     `json:"status,omitempty"`
	ContentSHA256    string     `json:"content_sha256,omitempty"`
	CapturedAt       *time.Time `json:"captured_at,omitempty"`
}

// OrderEvidencePackageAssembly is a read-time compiled view for the evidence
// workbench and payment-channel adapters. It deliberately keeps raw shipping
// records private to the service package so provider credentials cannot leak
// through this result.
type OrderEvidencePackageAssembly struct {
	Order             *order.Order                         `json:"order,omitempty"`
	Package           *orderevidence.OrderEvidencePackage  `json:"package,omitempty"`
	Completeness      orderevidence.EvidenceCompleteness   `json:"completeness"`
	TrackingContext   *OrderEvidenceTrackingContext        `json:"tracking_context,omitempty"`
	Sources           []OrderEvidenceSourceReference       `json:"sources"`
	Warnings          []string                             `json:"warnings,omitempty"`
	AssembledAt       time.Time                            `json:"assembled_at"`
	Shipment          *shipping.TrackingShipment           `json:"-"`
	TrackingEvents    []shipping.TrackingEvent             `json:"-"`
}

// OrderEvidencePackageAssembler joins order-time evidence facts with the
// current shipping projection. It does not upload files, mutate source
// records, or translate fields for Stripe/PayPal.
type OrderEvidencePackageAssembler struct {
	orderRepo    *repository.OrderRepository
	evidenceRepo *repository.OrderEvidenceRepository
	shippingRepo *repository.ShippingRepository
}

func NewOrderEvidencePackageAssembler(
	orderRepo *repository.OrderRepository,
	evidenceRepo *repository.OrderEvidenceRepository,
	shippingRepo *repository.ShippingRepository,
) *OrderEvidencePackageAssembler {
	return &OrderEvidencePackageAssembler{
		orderRepo:    orderRepo,
		evidenceRepo: evidenceRepo,
		shippingRepo: shippingRepo,
	}
}

func (s *OrderEvidencePackageAssembler) Assemble(
	orderID uint,
) (*OrderEvidencePackageAssembly, error) {
	if s == nil || s.orderRepo == nil {
		return nil, ErrOrderEvidenceAssemblyUnavailable
	}
	if orderID == 0 {
		return nil, ErrOrderEvidenceAssemblyOrderID
	}

	orderRecord, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	result := &OrderEvidencePackageAssembly{
		Order:          orderRecord,
		Sources:        []OrderEvidenceSourceReference{},
		Warnings:       []string{},
		AssembledAt:    time.Now().UTC(),
		TrackingEvents: []shipping.TrackingEvent{},
	}

	if s.evidenceRepo == nil {
		result.Warnings = append(result.Warnings,
			"order evidence package repository is not configured; fulfillment evidence attachments are unavailable.",
		)
	} else {
		pkg, err := s.evidenceRepo.FindLatestPackageByOrderID(orderID)
		switch {
		case err == nil:
			result.Sources = append(result.Sources, OrderEvidenceSourceReference{
				SourceType: "order_evidence_package",
				SourceID:   pkg.ID,
				Status:     pkg.Status,
			})
			if pkg.IsSuperseded() {
				result.Warnings = append(result.Warnings,
					"the latest order evidence package is superseded and cannot be used as the current fulfillment package.",
				)
			} else {
				result.Package = pkg
				result.Completeness = orderevidence.EvaluateEvidenceCompleteness(pkg.Items)
				result.Sources = append(result.Sources, orderEvidenceItemSources(pkg.Items)...)
			}
		case repository.IsRecordNotFound(err):
			result.Warnings = append(result.Warnings,
				"no order evidence package has been generated for this order.",
			)
		default:
			return nil, err
		}
	}

	if s.shippingRepo == nil {
		result.Warnings = append(result.Warnings,
			"shipping repository is not configured; tracking and delivery evidence are unavailable.",
		)
		return result, nil
	}

	shipment, err := s.shippingRepo.FindTrackingShipmentByOrderID(orderID)
	if err != nil && !repository.IsRecordNotFound(err) {
		return nil, err
	}
	events, err := s.shippingRepo.FindTrackingEventsByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	currentTrackingNumber := ""
	if shipment != nil {
		currentTrackingNumber = strings.TrimSpace(shipment.TrackingNumber)
		result.Shipment = shipment
		result.Sources = append(result.Sources, OrderEvidenceSourceReference{
			SourceType: "tracking_shipment",
			SourceID:   shipment.ID,
			Status:     shipment.SyncStatus,
		})
	} else {
		result.Warnings = append(result.Warnings,
			"no current tracking shipment is recorded for this order.",
		)
	}

	result.TrackingEvents = filterTrackingEventsForShipment(events, currentTrackingNumber)
	for _, event := range result.TrackingEvents {
		result.Sources = append(result.Sources, OrderEvidenceSourceReference{
			SourceType: "tracking_event",
			SourceID:   event.ID,
			Status:     event.Status,
			CapturedAt: assemblyTimePointer(event.EventTime),
		})
	}
	result.TrackingContext = buildOrderEvidenceTrackingContextFromSources(
		shipment,
		result.TrackingEvents,
		assemblyEvidenceItems(result),
	)
	return result, nil
}

func assemblyEvidenceItems(
	assembly *OrderEvidencePackageAssembly,
) []orderevidence.OrderEvidenceItem {
	if assembly == nil || assembly.Package == nil {
		return nil
	}
	return assembly.Package.Items
}

func orderEvidenceItemSources(
	items []orderevidence.OrderEvidenceItem,
) []OrderEvidenceSourceReference {
	sources := make([]OrderEvidenceSourceReference, 0, len(items))
	for _, item := range items {
		sources = append(sources, OrderEvidenceSourceReference{
			SourceType:       "order_evidence_item",
			SourceID:         item.ID,
			OrderItemID:      item.OrderItemID,
			EvidenceItemType: item.ItemType,
			Status:           item.Status,
			ContentSHA256:    strings.TrimSpace(item.ContentSHA256),
			CapturedAt:       item.CapturedAt,
		})
		for _, attachment := range item.Attachments {
			sources = append(sources, OrderEvidenceSourceReference{
				SourceType:    "order_evidence_attachment",
				SourceID:      attachment.ID,
				ContentSHA256: strings.TrimSpace(attachment.SHA256),
				CapturedAt:    assemblyTimePointer(attachment.CreatedAt),
			})
		}
	}
	return sources
}

func assemblyTimePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}

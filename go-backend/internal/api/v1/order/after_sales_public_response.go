package order

import (
	"time"

	"commerce-platform/internal/domain/aftersales"
)

// PublicAfterSalesCaseResponse is the customer-facing after-sales contract.
// Internal database IDs and operator identities are intentionally omitted;
// the order number in the URL is the customer's stable lookup reference.
type PublicAfterSalesCaseResponse struct {
	Type            string                           `json:"type"`
	Status          string                           `json:"status"`
	Reason          string                           `json:"reason"`
	Description     string                           `json:"description"`
	Resolution      string                           `json:"resolution"`
	Items           []PublicAfterSalesCaseItem       `json:"items"`
	ReturnShipments []PublicAfterSalesReturnShipment `json:"return_shipments,omitempty"`
	Events          []PublicAfterSalesCaseEvent      `json:"events"`
	Attachments     []PublicAfterSalesAttachment     `json:"attachments,omitempty"`
	RefundReview    *PublicAfterSalesRefundReview    `json:"refund_review,omitempty"`
	CreatedAt       time.Time                        `json:"created_at"`
	UpdatedAt       time.Time                        `json:"updated_at"`
	ClosedAt        *time.Time                       `json:"closed_at,omitempty"`
}

type PublicAfterSalesCaseItem struct {
	ProductName string `json:"product_name"`
	SKU         string `json:"sku"`
	Quantity    int    `json:"quantity"`
}

type PublicAfterSalesReturnShipment struct {
	WarehouseName    string     `json:"warehouse_name"`
	WarehouseAddress string     `json:"warehouse_address"`
	Carrier          string     `json:"carrier"`
	TrackingNumber   string     `json:"tracking_number"`
	TrackingURL      string     `json:"tracking_url,omitempty"`
	LabelURL         string     `json:"label_url,omitempty"`
	ShippedAt        *time.Time `json:"shipped_at,omitempty"`
	ReceivedAt       *time.Time `json:"received_at,omitempty"`
}

type PublicAfterSalesCaseEvent struct {
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Resolution string    `json:"resolution"`
	CreatedAt  time.Time `json:"created_at"`
}

type PublicAfterSalesAttachment struct {
	Kind        string    `json:"kind"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

type PublicAfterSalesRefundReview struct {
	Status              string     `json:"status"`
	ProposedAmountMinor int64      `json:"proposed_amount_minor"`
	Currency            string     `json:"currency"`
	RequestNotes        string     `json:"request_notes"`
	DecisionNotes       string     `json:"decision_notes"`
	ReviewedAt          *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func publicAfterSalesCaseFromDomain(record aftersales.AfterSalesCase) PublicAfterSalesCaseResponse {
	result := PublicAfterSalesCaseResponse{
		Type:            record.Type,
		Status:          record.Status,
		Reason:          record.Reason,
		Description:     record.Description,
		Resolution:      record.Resolution,
		Items:           make([]PublicAfterSalesCaseItem, 0, len(record.Items)),
		ReturnShipments: make([]PublicAfterSalesReturnShipment, 0, len(record.ReturnShipments)),
		Events:          make([]PublicAfterSalesCaseEvent, 0, len(record.Events)),
		Attachments:     make([]PublicAfterSalesAttachment, 0, len(record.Attachments)),
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
		ClosedAt:        record.ClosedAt,
	}
	for _, shipment := range record.ReturnShipments {
		result.ReturnShipments = append(result.ReturnShipments, PublicAfterSalesReturnShipment{
			WarehouseName:    shipment.WarehouseName,
			WarehouseAddress: shipment.WarehouseAddress,
			Carrier:          shipment.Carrier,
			TrackingNumber:   shipment.TrackingNumber,
			TrackingURL:      shipment.TrackingURL,
			LabelURL:         shipment.LabelURL,
			ShippedAt:        shipment.ShippedAt,
			ReceivedAt:       shipment.ReceivedAt,
		})
	}

	for _, item := range record.Items {
		result.Items = append(result.Items, PublicAfterSalesCaseItem{
			ProductName: item.ProductName,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
		})
	}
	for _, event := range record.Events {
		result.Events = append(result.Events, PublicAfterSalesCaseEvent{
			FromStatus: event.FromStatus,
			ToStatus:   event.ToStatus,
			Resolution: event.Resolution,
			CreatedAt:  event.CreatedAt,
		})
	}
	for _, attachment := range record.Attachments {
		result.Attachments = append(result.Attachments, PublicAfterSalesAttachment{
			Kind:        attachment.Kind,
			Filename:    attachment.Filename,
			ContentType: attachment.ContentType,
			SizeBytes:   attachment.SizeBytes,
			CreatedAt:   attachment.CreatedAt,
		})
	}
	if record.RefundReview != nil {
		result.RefundReview = &PublicAfterSalesRefundReview{
			Status:              record.RefundReview.Status,
			ProposedAmountMinor: record.RefundReview.ProposedAmountMinor,
			Currency:            record.RefundReview.Currency,
			RequestNotes:        record.RefundReview.RequestNotes,
			DecisionNotes:       record.RefundReview.DecisionNotes,
			ReviewedAt:          record.RefundReview.ReviewedAt,
			CreatedAt:           record.RefundReview.CreatedAt,
			UpdatedAt:           record.RefundReview.UpdatedAt,
		}
	}
	return result
}

func publicAfterSalesCasesFromDomain(records []aftersales.AfterSalesCase) []PublicAfterSalesCaseResponse {
	result := make([]PublicAfterSalesCaseResponse, 0, len(records))
	for _, record := range records {
		result = append(result, publicAfterSalesCaseFromDomain(record))
	}
	return result
}

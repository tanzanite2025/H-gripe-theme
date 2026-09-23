package order

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/aftersales"

	"github.com/stretchr/testify/require"
)

func TestPublicAfterSalesCaseResponseExposesProgressWithoutInternalIDs(t *testing.T) {
	caseID := uint(42)
	operatorID := uint(7)
	reviewedAt := time.Date(2026, 9, 12, 1, 2, 3, 0, time.UTC)
	record := aftersales.AfterSalesCase{
		ID:          caseID,
		OrderID:     99,
		Type:        aftersales.TypeReturnRefund,
		Status:      aftersales.StatusReturnInTransit,
		Reason:      "Package arrived damaged",
		Description: "The wheel needs inspection.",
		Resolution:  "Return shipment is in transit.",
		CreatedBy:   operatorID,
		UpdatedBy:   operatorID,
		Items: []aftersales.AfterSalesCaseItem{{
			ID:          11,
			CaseID:      caseID,
			OrderID:     99,
			OrderItemID: 12,
			ProductID:   13,
			ProductName: "Carbon wheelset",
			SKU:         "C50-DT240",
			Quantity:    1,
		}},
		Events: []aftersales.AfterSalesCaseEvent{{
			ID:           21,
			CaseID:       caseID,
			FromStatus:   aftersales.StatusAwaitingReturn,
			ToStatus:     aftersales.StatusReturnInTransit,
			Resolution:   "Return label issued.",
			UpdatedBy:    operatorID,
			OperatorName: "Support",
			CreatedAt:    reviewedAt,
		}},
		ReturnShipments: []aftersales.AfterSalesReturnShipment{{
			ID:               55,
			CaseID:           caseID,
			WarehouseName:    "EU Returns Hub",
			WarehouseAddress: "1 Returns Way, Amsterdam",
			Carrier:          "DHL",
			TrackingNumber:   "DHL-123",
			TrackingURL:      "https://tracking.example/DHL-123",
			LabelURL:         "https://labels.example/DHL-123",
			ShippedAt:        &reviewedAt,
		}},
		RefundReview: &aftersales.AfterSalesRefundReview{
			ID:             31,
			CaseID:         caseID,
			Status:         aftersales.RefundReviewStatusApproved,
			ProposedAmountMinor: 90000,
			Currency:       "USD",
			RequestNotes:   "Refund the paid amount.",
			DecisionNotes:  "Approved.",
			CreatedBy:      operatorID,
			ReviewedByID:   &operatorID,
			ReviewedAt:     &reviewedAt,
		},
	}

	payload, err := json.Marshal(publicAfterSalesCaseFromDomain(record))
	require.NoError(t, err)
	var decoded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &decoded))
	for _, forbidden := range []string{"id", "order_id", "created_by", "updated_by"} {
		if _, exists := decoded[forbidden]; exists {
			t.Fatalf("public after-sales response leaked %s: %s", forbidden, payload)
		}
	}

	var item []map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(decoded["items"], &item))
	for _, forbidden := range []string{"id", "case_id", "order_id", "order_item_id", "product_id"} {
		if _, exists := item[0][forbidden]; exists {
			t.Fatalf("public after-sales item leaked %s: %s", forbidden, payload)
		}
	}

	var event []map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(decoded["events"], &event))
	if _, exists := event[0]["updated_by"]; exists {
		t.Fatalf("public after-sales event leaked updated_by: %s", payload)
	}
	require.Contains(t, string(payload), "Return shipment is in transit.")
	require.Contains(t, string(payload), "Approved.")
	require.Contains(t, string(payload), "DHL-123")
	require.Contains(t, string(payload), "EU Returns Hub")
	for _, forbidden := range []string{"\"id\"", "\"case_id\"", "\"received_by\""} {
		// Internal shipment identity and warehouse operator data must not be
		// exposed through the customer-facing contract.
		if strings.Contains(string(decoded["return_shipments"]), forbidden) {
			t.Fatalf("public return shipment leaked %s: %s", forbidden, payload)
		}
	}
}

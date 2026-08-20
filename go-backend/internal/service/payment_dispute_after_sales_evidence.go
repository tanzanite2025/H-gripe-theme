package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
)

type DisputePolicyDisclosureEvidence struct {
	PolicyKey       string     `json:"policy_key"`
	Locale          string     `json:"locale"`
	RequestedLocale string     `json:"requested_locale"`
	Fallback        bool       `json:"fallback"`
	PolicyVersion   string     `json:"policy_version"`
	PolicyHash      string     `json:"policy_hash"`
	PolicyURL       string     `json:"policy_url"`
	DisclosedAt     time.Time  `json:"disclosed_at"`
	ConsentedAt     *time.Time `json:"consented_at,omitempty"`
	Source          string     `json:"source"`
}

type DisputeRefundLineItemEvidence struct {
	OrderItemID uint    `json:"order_item_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	LineTotal   float64 `json:"line_total"`
	Restock     bool    `json:"restock"`
}

type DisputeRefundEvidence struct {
	ID                      uint                            `json:"id"`
	ProviderRefundID        string                          `json:"provider_refund_id,omitempty"`
	Status                  string                          `json:"status"`
	Amount                  float64                         `json:"amount"`
	RequestedAmount         float64                         `json:"requested_amount"`
	Currency                string                          `json:"currency"`
	Reason                  string                          `json:"reason"`
	CreatedAt               time.Time                       `json:"created_at"`
	CompletedAt             *time.Time                      `json:"completed_at,omitempty"`
	CalculationSnapshotHash string                          `json:"calculation_snapshot_hash,omitempty"`
	LineItems               []DisputeRefundLineItemEvidence `json:"line_items,omitempty"`
}

func buildPolicyDisclosureEvidence(disclosure *orderdomain.PolicyDisclosure) *DisputePolicyDisclosureEvidence {
	if disclosure == nil {
		return nil
	}
	return &DisputePolicyDisclosureEvidence{
		PolicyKey:       disclosure.PolicyKey,
		Locale:          disclosure.Locale,
		RequestedLocale: disclosure.RequestedLocale,
		Fallback:        disclosure.Fallback,
		PolicyVersion:   disclosure.PolicyVersion,
		PolicyHash:      disclosure.PolicyHash,
		PolicyURL:       disclosure.PolicyURL,
		DisclosedAt:     disclosure.DisclosedAt,
		ConsentedAt:     disclosure.ConsentedAt,
		Source:          disclosure.Source,
	}
}

func buildRefundEvidence(refunds []paymentdomain.Refund, currency string) []DisputeRefundEvidence {
	result := make([]DisputeRefundEvidence, 0, len(refunds))
	for _, refund := range refunds {
		item := DisputeRefundEvidence{
			ID:               refund.ID,
			ProviderRefundID: strings.TrimSpace(disputeStringValue(refund.RefundID)),
			Status:           strings.TrimSpace(refund.Status),
			Amount:           refund.Amount,
			RequestedAmount:  refund.RequestedAmount,
			Currency:         strings.TrimSpace(currency),
			Reason:           strings.TrimSpace(refund.Reason),
			CreatedAt:        refund.CreatedAt.UTC(),
			CompletedAt:      refund.CompletedAt,
			LineItems:        []DisputeRefundLineItemEvidence{},
		}
		if snapshot := strings.TrimSpace(refund.CalculationSnapshot); snapshot != "" {
			hash := sha256.Sum256([]byte(snapshot))
			item.CalculationSnapshotHash = hex.EncodeToString(hash[:])
		}
		for _, line := range refund.LineItems {
			item.LineItems = append(item.LineItems, DisputeRefundLineItemEvidence{
				OrderItemID: line.OrderItemID,
				ProductName: strings.TrimSpace(line.ProductName),
				SKU:         strings.TrimSpace(line.SKU),
				Quantity:    line.Quantity,
				LineTotal:   line.LineTotalAmount,
				Restock:     line.Restock,
			})
		}
		result = append(result, item)
	}
	return result
}

func disputeStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func refundEvidenceSummary(refunds []DisputeRefundEvidence) string {
	if len(refunds) == 0 {
		return ""
	}
	lines := []string{fmt.Sprintf("Refund records: %d.", len(refunds))}
	for _, refund := range refunds {
		providerID := strings.TrimSpace(refund.ProviderRefundID)
		if providerID == "" {
			providerID = "not assigned"
		}
		line := fmt.Sprintf(
			"- refund #%d: status=%s; amount=%.2f %s; requested=%.2f; provider_refund_id=%s; created_at=%s",
			refund.ID,
			refund.Status,
			refund.Amount,
			refund.Currency,
			refund.RequestedAmount,
			providerID,
			refund.CreatedAt.UTC().Format(time.RFC3339),
		)
		if refund.CompletedAt != nil {
			line += "; completed_at=" + refund.CompletedAt.UTC().Format(time.RFC3339)
		}
		if refund.Reason != "" {
			line += "; reason=" + refund.Reason
		}
		lines = append(lines, line)
		for _, item := range refund.LineItems {
			lines = append(lines, fmt.Sprintf(
				"  - item %d: %s SKU %s x%d line_total=%.2f restock=%t",
				item.OrderItemID,
				item.ProductName,
				item.SKU,
				item.Quantity,
				item.LineTotal,
				item.Restock,
			))
		}
	}
	return truncateEvidenceText(strings.Join(lines, "\n"), 8000)
}

func latestRefundEvidenceAt(refunds []DisputeRefundEvidence) *time.Time {
	var latest time.Time
	for _, refund := range refunds {
		if refund.CreatedAt.After(latest) {
			latest = refund.CreatedAt
		}
	}
	if latest.IsZero() {
		return nil
	}
	value := latest.UTC()
	return &value
}

func formatOptionalDisputeTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "not recorded"
	}
	return value.UTC().Format(time.RFC3339)
}

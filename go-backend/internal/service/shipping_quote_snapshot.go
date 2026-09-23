package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const shippingQuoteLifetime = 15 * time.Minute

type shippingQuoteHashItem struct {
	ProductID      uint  `json:"product_id"`
	VariantID      uint  `json:"variant_id"`
	TemplateID     uint  `json:"template_id"`
	Quantity       int   `json:"quantity"`
	UnitPriceMinor int64 `json:"unit_price_minor"`
	WeightGrams    int   `json:"weight_grams"`
}

type shippingQuoteHashInput struct {
	Country         string                  `json:"country"`
	PostalCode      string                  `json:"postal_code"`
	Currency        string                  `json:"currency"`
	DisplayCurrency string                  `json:"display_currency"`
	Items           []shippingQuoteHashItem `json:"items"`
}

func (s *ShippingService) persistShippingQuote(input ShippingQuoteInput, quote *ShippingQuote) error {
	if s == nil || s.shippingRepo == nil {
		return errorsWithCause(ErrShippingRateUnavailable, "shipping quote repository is not configured")
	}
	if quote == nil || len(quote.Plans) == 0 {
		return ErrShippingQuotePlanUnavailable
	}
	if strings.TrimSpace(input.SelectedQuotePlanID) != "" {
		return fmt.Errorf("%w: a plan can only be selected from an existing quote", ErrShippingQuotePlanUnavailable)
	}

	quote.ID = uuid.NewString()
	quote.ExpiresAt = time.Now().UTC().Add(shippingQuoteLifetime)
	for index := range quote.Plans {
		quote.Plans[index].ID = uuid.NewString()
	}
	quote.RateVersion = shippingQuoteRateVersion(quote.Plans)
	applyShippingQuotePlan(quote, &quote.Plans[0])

	quoteData, err := json.Marshal(quote)
	if err != nil {
		return fmt.Errorf("encode shipping quote snapshot: %w", err)
	}
	snapshot := &shippingdomain.QuoteSnapshot{
		ID:          quote.ID,
		RequestHash: shippingQuoteRequestHash(input),
		RateVersion: quote.RateVersion,
		QuoteData:   datatypes.JSON(quoteData),
		ExpiresAt:   quote.ExpiresAt,
	}
	if err := s.shippingRepo.CreateQuoteSnapshot(snapshot); err != nil {
		return fmt.Errorf("persist shipping quote snapshot: %w", err)
	}
	return nil
}

func (s *ShippingService) restoreShippingQuote(input ShippingQuoteInput) (*ShippingQuote, error) {
	quoteID := strings.TrimSpace(input.ShippingQuoteID)
	if quoteID == "" {
		return nil, nil
	}
	if s == nil || s.shippingRepo == nil {
		return nil, errorsWithCause(ErrShippingRateUnavailable, "shipping quote repository is not configured")
	}
	snapshot, err := s.shippingRepo.FindQuoteSnapshotByID(quoteID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, fmt.Errorf("%w: quote %s was not found", ErrShippingQuoteExpired, quoteID)
		}
		return nil, fmt.Errorf("load shipping quote snapshot: %w", err)
	}
	if !time.Now().UTC().Before(snapshot.ExpiresAt) {
		return nil, fmt.Errorf("%w: quote %s expired at %s", ErrShippingQuoteExpired, quoteID, snapshot.ExpiresAt.Format(time.RFC3339))
	}
	if snapshot.RequestHash != shippingQuoteRequestHash(input) {
		return nil, fmt.Errorf("%w: quote %s does not match destination, currency, or items", ErrShippingQuoteStale, quoteID)
	}

	var quote ShippingQuote
	if err := json.Unmarshal(snapshot.QuoteData, &quote); err != nil {
		return nil, fmt.Errorf("%w: quote %s snapshot is invalid", ErrShippingRateConfigurationInvalid, quoteID)
	}
	if quote.ID != snapshot.ID || quote.RateVersion != snapshot.RateVersion {
		return nil, fmt.Errorf("%w: quote %s snapshot metadata is inconsistent", ErrShippingRateConfigurationInvalid, quoteID)
	}
	planID := strings.TrimSpace(input.SelectedQuotePlanID)
	if planID == "" && quote.SelectedPlan != nil {
		planID = quote.SelectedPlan.ID
	}
	for index := range quote.Plans {
		if quote.Plans[index].ID == planID {
			applyShippingQuotePlan(&quote, &quote.Plans[index])
			return &quote, nil
		}
	}
	return nil, fmt.Errorf("%w: plan %s does not belong to quote %s", ErrShippingQuotePlanUnavailable, planID, quoteID)
}

func shippingQuoteRequestHash(input ShippingQuoteInput) string {
	items := make([]shippingQuoteHashItem, 0, len(input.Items))
	for _, item := range input.Items {
		variantID := uint(0)
		if item.VariantID != nil {
			variantID = *item.VariantID
		}
		templateID := uint(0)
		if item.ShippingTemplateID != nil {
			templateID = *item.ShippingTemplateID
		}
		unitPriceMinor := item.UnitPriceMinor
		items = append(items, shippingQuoteHashItem{
			ProductID: item.ProductID, VariantID: variantID, TemplateID: templateID,
			Quantity: item.Quantity, UnitPriceMinor: unitPriceMinor, WeightGrams: item.WeightGrams,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].ProductID != items[j].ProductID {
			return items[i].ProductID < items[j].ProductID
		}
		if items[i].VariantID != items[j].VariantID {
			return items[i].VariantID < items[j].VariantID
		}
		if items[i].TemplateID != items[j].TemplateID {
			return items[i].TemplateID < items[j].TemplateID
		}
		if items[i].Quantity != items[j].Quantity {
			return items[i].Quantity < items[j].Quantity
		}
		if items[i].UnitPriceMinor != items[j].UnitPriceMinor {
			return items[i].UnitPriceMinor < items[j].UnitPriceMinor
		}
		return items[i].WeightGrams < items[j].WeightGrams
	})
	payload, _ := json.Marshal(shippingQuoteHashInput{
		Country:         strings.ToUpper(strings.TrimSpace(input.Country)),
		PostalCode:      normalizeShippingPostalCode(input.PostalCode),
		Currency:        strings.ToUpper(strings.TrimSpace(input.Currency)),
		DisplayCurrency: strings.ToUpper(strings.TrimSpace(input.DisplayCurrency)),
		Items:           items,
	})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func shippingQuoteRateVersion(plans []ShippingQuotePlan) string {
	versioned := make([]ShippingQuotePlan, len(plans))
	copy(versioned, plans)
	for index := range versioned {
		versioned[index].ID = ""
	}
	payload, _ := json.Marshal(versioned)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func errorsWithCause(cause error, message string) error {
	return fmt.Errorf("%w: %s", cause, message)
}

package service

import (
	"errors"

	"commerce-platform/internal/domain/currency"
)

type BackendEntryCurrencyMismatchSample struct {
	Kind      string `json:"kind"`
	ID        uint   `json:"id"`
	ProductID uint   `json:"product_id,omitempty"`
	SKU       string `json:"sku"`
	Name      string `json:"name,omitempty"`
	Title     string `json:"title,omitempty"`
	Currency  string `json:"currency"`
}

type BackendEntryCurrencyAudit struct {
	ExpectedCurrency     string                               `json:"expected_currency"`
	ProductMismatchCount int64                                `json:"product_mismatch_count"`
	VariantMismatchCount int64                                `json:"variant_mismatch_count"`
	TotalMismatchCount   int64                                `json:"total_mismatch_count"`
	Samples              []BackendEntryCurrencyMismatchSample `json:"samples"`
}

func (s *ProductService) AuditBackendEntryCurrencyConsistency(expectedCurrency string, sampleLimit int) (BackendEntryCurrencyAudit, error) {
	if s == nil || s.productRepo == nil {
		return BackendEntryCurrencyAudit{}, errors.New("product service is not configured")
	}
	expectedCurrency = currency.NormalizeCode(expectedCurrency)
	if expectedCurrency == "" || !currency.IsCatalogCode(expectedCurrency) {
		expectedCurrency = s.primaryPricingCurrency()
	}
	if sampleLimit <= 0 || sampleLimit > 100 {
		sampleLimit = 10
	}

	productCount, err := s.productRepo.CountProductsWithCurrencyMismatch(expectedCurrency)
	if err != nil {
		return BackendEntryCurrencyAudit{}, err
	}
	variantCount, err := s.productRepo.CountProductVariantsWithCurrencyMismatch(expectedCurrency)
	if err != nil {
		return BackendEntryCurrencyAudit{}, err
	}

	samples := make([]BackendEntryCurrencyMismatchSample, 0, sampleLimit)
	productSamples, err := s.productRepo.ListProductsWithCurrencyMismatch(expectedCurrency, sampleLimit)
	if err != nil {
		return BackendEntryCurrencyAudit{}, err
	}
	for _, sample := range productSamples {
		if len(samples) >= sampleLimit {
			break
		}
		samples = append(samples, BackendEntryCurrencyMismatchSample{
			Kind:     "product",
			ID:       sample.ID,
			SKU:      sample.SKU,
			Name:     sample.Name,
			Currency: currency.NormalizeCode(sample.Currency),
		})
	}

	if len(samples) < sampleLimit {
		variantSamples, err := s.productRepo.ListProductVariantsWithCurrencyMismatch(expectedCurrency, sampleLimit-len(samples))
		if err != nil {
			return BackendEntryCurrencyAudit{}, err
		}
		for _, sample := range variantSamples {
			if len(samples) >= sampleLimit {
				break
			}
			samples = append(samples, BackendEntryCurrencyMismatchSample{
				Kind:      "variant",
				ID:        sample.ID,
				ProductID: sample.ProductID,
				SKU:       sample.SKU,
				Title:     sample.Title,
				Currency:  currency.NormalizeCode(sample.Currency),
			})
		}
	}

	return BackendEntryCurrencyAudit{
		ExpectedCurrency:     expectedCurrency,
		ProductMismatchCount: productCount,
		VariantMismatchCount: variantCount,
		TotalMismatchCount:   productCount + variantCount,
		Samples:              samples,
	}, nil
}

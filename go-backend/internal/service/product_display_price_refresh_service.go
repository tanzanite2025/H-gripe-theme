package service

import (
	"errors"
	"math"
	"sort"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

const displayPriceRefreshBatchSize = 500

type ProductDisplayPriceRefreshResult struct {
	BaseCurrency          string   `json:"base_currency"`
	QuoteCurrencies       []string `json:"quote_currencies"`
	ProductsScanned       int      `json:"products_scanned"`
	ProductsUpdated       int      `json:"products_updated"`
	VariantsScanned       int      `json:"variants_scanned"`
	VariantsUpdated       int      `json:"variants_updated"`
	CurrencyMismatchCount int      `json:"currency_mismatch_count"`
}

// RefreshDisplayPriceSnapshots rebuilds customer-facing display prices from
// the latest cached rates. It never changes source currency or source amounts.
func (s *ProductService) RefreshDisplayPriceSnapshots(
	baseCurrency string,
	quoteCurrencies []string,
	rates []currency.ExchangeRate,
) (ProductDisplayPriceRefreshResult, error) {
	if s == nil || s.productRepo == nil {
		return ProductDisplayPriceRefreshResult{}, errors.New("product service is not configured")
	}

	baseCurrency = currency.NormalizeCode(baseCurrency)
	if !currency.IsCatalogCode(baseCurrency) {
		return ProductDisplayPriceRefreshResult{}, errors.New("display price refresh base currency is invalid")
	}

	quoteCurrencies = normalizeDisplayPriceRefreshQuotes(quoteCurrencies, baseCurrency)
	ratesByQuote := displayPriceRatesByQuote(rates, baseCurrency)
	products, err := s.productRepo.ListProductsForDisplayPriceRefresh()
	if err != nil {
		return ProductDisplayPriceRefreshResult{}, err
	}

	result := ProductDisplayPriceRefreshResult{
		BaseCurrency:    baseCurrency,
		QuoteCurrencies: append([]string(nil), quoteCurrencies...),
		ProductsScanned: len(products),
	}
	updates := make([]repository.ProductDisplayPriceSnapshotUpdate, 0, len(products))
	updatedProductIDs := make([]uint, 0, len(products))

	for _, item := range products {
		result.VariantsScanned += len(item.Variants)

		productCurrency := currency.NormalizeCode(item.Currency)
		if productCurrency != baseCurrency {
			result.CurrencyMismatchCount++
		}

		var productSnapshot datatypes.JSON
		productCanRefresh := productCurrency == baseCurrency
		if productCanRefresh {
			productSnapshot = displayPriceSnapshotJSON(
				item.Price,
				item.SalePrice,
				baseCurrency,
				quoteCurrencies,
				ratesByQuote,
				currency.ParseDisplayPriceSnapshots(item.DisplayPriceData),
			)
			result.ProductsUpdated++
		}

		update := repository.ProductDisplayPriceSnapshotUpdate{
			ProductID: item.ID,
		}
		if productCanRefresh {
			update.UpdateProduct = true
			update.DisplayPriceData = productSnapshot
			updatedProductIDs = append(updatedProductIDs, item.ID)
		}

		for _, variant := range item.Variants {
			variantCurrency := currency.NormalizeCode(variant.Currency)
			if variantCurrency != baseCurrency {
				result.CurrencyMismatchCount++
				continue
			}

			variantSnapshot := displayPriceSnapshotJSON(
				variant.Price,
				variant.SalePrice,
				baseCurrency,
				quoteCurrencies,
				ratesByQuote,
				currency.ParseDisplayPriceSnapshots(variant.DisplayPriceData),
			)
			update.VariantSnapshotUpdates = append(update.VariantSnapshotUpdates, repository.ProductVariantDisplayPriceSnapshotUpdate{
				VariantID:        variant.ID,
				DisplayPriceData: variantSnapshot,
			})
			result.VariantsUpdated++
		}

		if productCanRefresh {
			updates = append(updates, update)
			continue
		}
		if len(update.VariantSnapshotUpdates) > 0 {
			updatedProductIDs = append(updatedProductIDs, item.ID)
			updates = append(updates, update)
		}
	}

	if err := s.productRepo.UpdateDisplayPriceSnapshots(updates); err != nil {
		return result, err
	}

	for start := 0; start < len(updatedProductIDs); start += displayPriceRefreshBatchSize {
		end := start + displayPriceRefreshBatchSize
		if end > len(updatedProductIDs) {
			end = len(updatedProductIDs)
		}
		batch := updatedProductIDs[start:end]
		s.InvalidateProductCacheByIDs(batch)
		if err := s.enqueueProductCacheInvalidationByIDs(batch, "exchange rate display price refresh"); err != nil {
			return result, err
		}
	}
	if len(updatedProductIDs) > 0 {
		s.invalidateStorefrontHTMLCache("exchange rate display price refresh")
	}

	return result, nil
}

func normalizeDisplayPriceRefreshQuotes(values []string, baseCurrency string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		code := currency.NormalizeCode(value)
		if code == "" || code == baseCurrency || !currency.IsCatalogCode(code) {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		result = append(result, code)
	}
	sort.Strings(result)
	return result
}

func displayPriceRatesByQuote(rates []currency.ExchangeRate, baseCurrency string) map[string]float64 {
	result := make(map[string]float64, len(rates))
	for _, rate := range rates {
		if currency.NormalizeCode(rate.BaseCurrency) != baseCurrency {
			continue
		}
		quote := currency.NormalizeCode(rate.QuoteCurrency)
		if quote == "" || quote == baseCurrency || !currency.IsCatalogCode(quote) || rate.Rate <= 0 {
			continue
		}
		result[quote] = rate.Rate
	}
	return result
}

func displayPriceSnapshotJSON(
	price float64,
	salePrice *float64,
	baseCurrency string,
	quoteCurrencies []string,
	ratesByQuote map[string]float64,
	previousSnapshots []currency.DisplayPriceSnapshot,
) datatypes.JSON {
	amount := price
	if salePrice != nil && *salePrice > 0 {
		amount = *salePrice
	}
	if amount <= 0 {
		return datatypes.JSON([]byte("[]"))
	}

	previousByQuote := make(map[string]currency.DisplayPriceSnapshot, len(previousSnapshots))
	for _, snapshot := range previousSnapshots {
		quote := currency.NormalizeCode(snapshot.QuoteCurrency)
		if quote == "" {
			quote = currency.NormalizeCode(snapshot.Currency)
		}
		if quote != "" {
			previousByQuote[quote] = snapshot
		}
	}

	snapshots := make([]currency.DisplayPriceSnapshot, 0, len(quoteCurrencies))
	for _, quote := range quoteCurrencies {
		rate, ok := ratesByQuote[quote]
		if !ok || rate <= 0 {
			if previous, exists := previousByQuote[quote]; exists {
				snapshots = append(snapshots, previous)
			}
			continue
		}
		snapshots = append(snapshots, currency.DisplayPriceSnapshot{
			Amount:        roundDisplayPriceAmount(amount*rate, quote),
			Currency:      quote,
			QuoteCurrency: quote,
			Rate:          rate,
			Source:        "direct_rate",
			Converted:     true,
		})
	}
	return currency.DisplayPriceSnapshotsJSON(snapshots, baseCurrency)
}

func roundDisplayPriceAmount(amount float64, currencyCode string) float64 {
	minorUnits, ok := currency.MinorUnits(currencyCode)
	if !ok {
		minorUnits = 2
	}
	factor := math.Pow10(minorUnits)
	return math.Round(amount*factor) / factor
}

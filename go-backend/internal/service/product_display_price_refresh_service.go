package service

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"sync/atomic"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
)

const displayPriceRefreshBatchSize = 500
const displayPriceRefreshLeaseKey = "product-display-price-refresh"
const displayPriceRefreshLeaseTTL = 30 * time.Minute

var displayPriceRefreshOwnerSequence uint64

var ErrProductDisplayPriceRefreshInProgress = errors.New("product display price refresh is already running")

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
	leaseOwner := fmt.Sprintf("product-display-price:%d:%d", os.Getpid(), atomic.AddUint64(&displayPriceRefreshOwnerSequence, 1))
	if s.displayPriceLeaseRepo != nil {
		acquired, err := s.displayPriceLeaseRepo.TryAcquireSyncLease(displayPriceRefreshLeaseKey, leaseOwner, time.Now().UTC(), displayPriceRefreshLeaseTTL)
		if err != nil {
			return ProductDisplayPriceRefreshResult{}, err
		}
		if !acquired {
			return ProductDisplayPriceRefreshResult{}, ErrProductDisplayPriceRefreshInProgress
		}
		heartbeatDone := make(chan struct{})
		go func() {
			ticker := time.NewTicker(displayPriceRefreshLeaseTTL / 3)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if renewed, renewErr := s.displayPriceLeaseRepo.RenewSyncLease(displayPriceRefreshLeaseKey, leaseOwner, time.Now().UTC(), displayPriceRefreshLeaseTTL); renewErr != nil || !renewed {
						return
					}
				case <-heartbeatDone:
					return
				}
			}
		}()
		defer func() {
			close(heartbeatDone)
			_ = s.displayPriceLeaseRepo.ReleaseSyncLease(displayPriceRefreshLeaseKey, leaseOwner, time.Now().UTC())
		}()
	}
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
		productSourcePrice, productSourceSalePrice, priceErr := productDisplaySourcePrices(&item)
		if priceErr != nil {
			return result, fmt.Errorf("resolve source price for product %d: %w", item.ID, priceErr)
		}

		productCurrency := currency.NormalizeCode(item.Currency)
		if productCurrency != baseCurrency {
			result.CurrencyMismatchCount++
		}

		var productSnapshot datatypes.JSON
		productCanRefresh := productCurrency == baseCurrency
		if productCanRefresh {
			productSnapshot = displayPriceSnapshotJSON(
				productSourcePrice,
				productSourceSalePrice,
				baseCurrency,
				quoteCurrencies,
				ratesByQuote,
				currency.ParseDisplayPriceSnapshots(item.DisplayPriceData),
			)
			result.ProductsUpdated++
		}

		update := repository.ProductDisplayPriceSnapshotUpdate{
			ProductID:      item.ID,
			SourceCurrency: item.Currency,
		}
		if sourceMoney, moneyErr := item.PriceMoney(); moneyErr == nil {
			update.SourcePriceMinor = sourceMoney.AmountMinor()
		}
		if saleMoney, moneyErr := item.SalePriceMoney(); moneyErr == nil && saleMoney != nil {
			minor := saleMoney.AmountMinor()
			update.SourceSalePriceMinor = &minor
		}
		if productCanRefresh {
			update.UpdateProduct = true
			update.DisplayPriceData = productSnapshot
			updatedProductIDs = append(updatedProductIDs, item.ID)
		}

		for _, variant := range item.Variants {
			variantSourcePrice, variantSourceSalePrice, variantPriceErr := productDisplaySourcePricesVariant(&variant)
			if variantPriceErr != nil {
				return result, fmt.Errorf("resolve source price for variant %d: %w", variant.ID, variantPriceErr)
			}
			variantCurrency := currency.NormalizeCode(variant.Currency)
			if variantCurrency != baseCurrency {
				result.CurrencyMismatchCount++
				continue
			}

			variantSnapshot := displayPriceSnapshotJSON(
				variantSourcePrice,
				variantSourceSalePrice,
				baseCurrency,
				quoteCurrencies,
				ratesByQuote,
				currency.ParseDisplayPriceSnapshots(variant.DisplayPriceData),
			)
			variantUpdate := repository.ProductVariantDisplayPriceSnapshotUpdate{
				VariantID:        variant.ID,
				SourceCurrency:   variant.Currency,
				DisplayPriceData: variantSnapshot,
			}
			if sourceMoney, moneyErr := variant.PriceMoney(); moneyErr == nil {
				variantUpdate.SourcePriceMinor = sourceMoney.AmountMinor()
			}
			if saleMoney, moneyErr := variant.SalePriceMoney(); moneyErr == nil && saleMoney != nil {
				minor := saleMoney.AmountMinor()
				variantUpdate.SourceSalePriceMinor = &minor
			}
			update.VariantSnapshotUpdates = append(update.VariantSnapshotUpdates, variantUpdate)
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

func productDisplaySourcePrices(item *product.Product) (float64, *float64, error) {
	priceMoney, err := item.PriceMoney()
	if err != nil {
		return 0, nil, err
	}
	price, err := priceMoney.MajorFloat()
	if err != nil {
		return 0, nil, err
	}
	saleMoney, err := item.SalePriceMoney()
	if err != nil || saleMoney == nil {
		return price, nil, err
	}
	sale, err := saleMoney.MajorFloat()
	if err != nil {
		return 0, nil, err
	}
	return price, &sale, nil
}

func productDisplaySourcePricesVariant(item *product.ProductVariant) (float64, *float64, error) {
	priceMoney, err := item.PriceMoney()
	if err != nil {
		return 0, nil, err
	}
	price, err := priceMoney.MajorFloat()
	if err != nil {
		return 0, nil, err
	}
	saleMoney, err := item.SalePriceMoney()
	if err != nil || saleMoney == nil {
		return price, nil, err
	}
	sale, err := saleMoney.MajorFloat()
	if err != nil {
		return 0, nil, err
	}
	return price, &sale, nil
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
		baseMoney, moneyErr := domainmoney.FromMajorFloat(amount, baseCurrency)
		if moneyErr != nil {
			continue
		}
		convertedMoney, moneyErr := baseMoney.ConvertAtRate(rate, quote)
		if moneyErr != nil {
			continue
		}
		convertedAmount, moneyErr := convertedMoney.MajorFloat()
		if moneyErr != nil || convertedAmount <= 0 {
			continue
		}
		snapshots = append(snapshots, currency.DisplayPriceSnapshot{
			Amount:        convertedAmount,
			Currency:      quote,
			QuoteCurrency: quote,
			Rate:          rate,
			Source:        "direct_rate",
			Converted:     true,
		})
	}
	return currency.DisplayPriceSnapshotsJSON(snapshots, baseCurrency)
}

package service

import (
	"errors"
	"fmt"
	"math"
	"math/big"
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
	ratesByQuote := displayPriceRatesByQuoteRat(rates, baseCurrency)
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
		productSourcePrice, productSourceSalePrice, priceErr := productDisplaySourceMoney(&item)
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
			productSnapshot = displayPriceSnapshotMoneyJSON(
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
			SourceCurrency: productCurrency,
		}
		if sourceVariant := item.StartingPriceVariant(); sourceVariant != nil {
			if sourceMoney, moneyErr := sourceVariant.PriceMoney(); moneyErr == nil {
				update.SourcePriceMinor = sourceMoney.AmountMinor()
			}
			if saleMoney, moneyErr := sourceVariant.SalePriceMoney(); moneyErr == nil && saleMoney != nil {
				minor := saleMoney.AmountMinor()
				update.SourceSalePriceMinor = &minor
			}
		} else if sourceMoney, moneyErr := item.PriceMoney(); moneyErr == nil {
			update.SourcePriceMinor = sourceMoney.AmountMinor()
		}
		if productCanRefresh {
			update.UpdateProduct = true
			update.DisplayPriceData = productSnapshot
			updatedProductIDs = append(updatedProductIDs, item.ID)
		}

		for _, variant := range item.Variants {
			variantSourcePrice, variantSourceSalePrice, variantPriceErr := productDisplaySourceMoneyVariant(&variant)
			if variantPriceErr != nil {
				return result, fmt.Errorf("resolve source price for variant %d: %w", variant.ID, variantPriceErr)
			}
			variantCurrency := currency.NormalizeCode(variant.Currency)
			if variantCurrency != baseCurrency {
				result.CurrencyMismatchCount++
				continue
			}

			variantSnapshot := displayPriceSnapshotMoneyJSON(
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

func productDisplaySourceMoney(item *product.Product) (domainmoney.Money, *domainmoney.Money, error) {
	if item == nil {
		return domainmoney.Money{}, nil, errors.New("product is required")
	}
	if variant := item.StartingPriceVariant(); variant != nil {
		return productDisplaySourceMoneyVariant(variant)
	}
	priceMoney, err := item.PriceMoney()
	if err != nil {
		return domainmoney.Money{}, nil, err
	}
	saleMoney, err := item.SalePriceMoney()
	if err != nil || saleMoney == nil {
		return priceMoney, nil, err
	}
	return priceMoney, saleMoney, nil
}

func productDisplaySourceMoneyVariant(item *product.ProductVariant) (domainmoney.Money, *domainmoney.Money, error) {
	if item == nil {
		return domainmoney.Money{}, nil, errors.New("product variant is required")
	}
	priceMoney, err := item.PriceMoney()
	if err != nil {
		return domainmoney.Money{}, nil, err
	}
	saleMoney, err := item.SalePriceMoney()
	if err != nil || saleMoney == nil {
		return priceMoney, nil, err
	}
	return priceMoney, saleMoney, nil
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

// displayPriceRatesByQuoteRat keeps the refresh calculation exact until the
// final JSON transport boundary.
func displayPriceRatesByQuoteRat(rates []currency.ExchangeRate, baseCurrency string) map[string]*big.Rat {
	result := make(map[string]*big.Rat, len(rates))
	for _, rate := range rates {
		if currency.NormalizeCode(rate.BaseCurrency) != baseCurrency {
			continue
		}
		quote := currency.NormalizeCode(rate.QuoteCurrency)
		if quote == "" || quote == baseCurrency || !currency.IsCatalogCode(quote) {
			continue
		}
		value, err := exchangeRateRatRecord(&rate)
		if err != nil || value.Sign() <= 0 {
			continue
		}
		result[quote] = value
	}
	return result
}

func displayPriceRateFloat(rate *big.Rat) (float64, bool) {
	if rate == nil || rate.Sign() <= 0 {
		return 0, false
	}
	value, _ := rate.Float64()
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}
	return value, true
}

func displayPriceSnapshotMoneyJSON(
	price domainmoney.Money,
	salePrice *domainmoney.Money,
	baseCurrency string,
	quoteCurrencies []string,
	ratesByQuote map[string]*big.Rat,
	previousSnapshots []currency.DisplayPriceSnapshot,
) datatypes.JSON {
	amount := price
	if salePrice != nil && salePrice.AmountMinor() > 0 {
		amount = *salePrice
	}
	if amount.AmountMinor() <= 0 || currency.NormalizeCode(amount.Currency().String()) != currency.NormalizeCode(baseCurrency) {
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
		if !ok || rate == nil || rate.Sign() <= 0 {
			if previous, exists := previousByQuote[quote]; exists {
				snapshots = append(snapshots, previous)
			}
			continue
		}
		convertedMoney, moneyErr := amount.ConvertAtRat(rate, quote)
		if moneyErr != nil || convertedMoney.AmountMinor() <= 0 {
			continue
		}
		convertedAmountText, moneyErr := convertedMoney.FormatMajor()
		rateFloat, rateOK := displayPriceRateFloat(rate)
		if moneyErr != nil || !rateOK {
			continue
		}
		snapshots = append(snapshots, currency.DisplayPriceSnapshot{
			AmountDecimal: convertedAmountText,
			Currency:      quote,
			QuoteCurrency: quote,
			Rate:          rateFloat,
			Source:        "direct_rate",
			Converted:     true,
		})
	}
	return currency.DisplayPriceSnapshotsJSON(snapshots, baseCurrency)
}

package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/gorm"
)

type ExchangeRateRepository struct {
	db *gorm.DB
}

func NewExchangeRateRepository(db *gorm.DB) *ExchangeRateRepository {
	return &ExchangeRateRepository{db: db}
}

func (r *ExchangeRateRepository) WithTx(tx *gorm.DB) *ExchangeRateRepository {
	return &ExchangeRateRepository{db: tx}
}

func (r *ExchangeRateRepository) Find(baseCurrency, quoteCurrency string) (*currency.ExchangeRate, error) {
	var rate currency.ExchangeRate
	if err := r.db.Where(
		"base_currency = ? AND quote_currency = ?",
		strings.ToUpper(strings.TrimSpace(baseCurrency)),
		strings.ToUpper(strings.TrimSpace(quoteCurrency)),
	).First(&rate).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *ExchangeRateRepository) List(baseCurrency string) ([]currency.ExchangeRate, error) {
	var rates []currency.ExchangeRate
	query := r.db.Order("quote_currency ASC")
	baseCurrency = strings.ToUpper(strings.TrimSpace(baseCurrency))
	if baseCurrency != "" {
		query = query.Where("base_currency = ?", baseCurrency)
	}
	if err := query.Find(&rates).Error; err != nil {
		return nil, err
	}
	return rates, nil
}

func (r *ExchangeRateRepository) UpsertRates(rates []currency.ExchangeRate) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range rates {
			rate := rates[i]
			var existing currency.ExchangeRate
			err := tx.Where("base_currency = ? AND quote_currency = ?", rate.BaseCurrency, rate.QuoteCurrency).First(&existing).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := tx.Create(&rate).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}
			updates := map[string]interface{}{
				"rate":       rate.Rate,
				"source":     rate.Source,
				"fetched_at": rate.FetchedAt,
				"expires_at": rate.ExpiresAt,
				"updated_at": time.Now().UTC(),
			}
			if err := tx.Model(&currency.ExchangeRate{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

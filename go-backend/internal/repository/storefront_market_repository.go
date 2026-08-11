package repository

import (
	"strings"

	"commerce-platform/internal/domain/market"

	"gorm.io/gorm"
)

type StorefrontMarketRepository struct {
	db *gorm.DB
}

func NewStorefrontMarketRepository(db *gorm.DB) *StorefrontMarketRepository {
	return &StorefrontMarketRepository{db: db}
}

func (r *StorefrontMarketRepository) List(enabledOnly bool) ([]market.StorefrontMarket, error) {
	var markets []market.StorefrontMarket
	query := r.db.Preload("Countries", func(db *gorm.DB) *gorm.DB {
		return db.Order("code ASC")
	}).Order("priority ASC").Order("code ASC")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	if err := query.Find(&markets).Error; err != nil {
		return nil, err
	}
	return markets, nil
}

func (r *StorefrontMarketRepository) FindByID(id uint) (*market.StorefrontMarket, error) {
	var result market.StorefrontMarket
	if err := r.db.Preload("Countries", func(db *gorm.DB) *gorm.DB {
		return db.Order("code ASC")
	}).First(&result, id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *StorefrontMarketRepository) FindByCode(code string) (*market.StorefrontMarket, error) {
	var result market.StorefrontMarket
	if err := r.db.Preload("Countries", func(db *gorm.DB) *gorm.DB {
		return db.Order("code ASC")
	}).Where("UPPER(code) = ?", strings.ToUpper(strings.TrimSpace(code))).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *StorefrontMarketRepository) FindByCountry(country string) (*market.StorefrontMarket, error) {
	country = strings.ToUpper(strings.TrimSpace(country))
	var countryRecord market.MarketCountry
	if err := r.db.Where("code = ?", country).First(&countryRecord).Error; err != nil {
		return nil, err
	}
	return r.FindByID(countryRecord.MarketID)
}

func (r *StorefrontMarketRepository) Create(input *market.StorefrontMarket) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		enabled := input.Enabled
		countries := append([]market.MarketCountry(nil), input.Countries...)
		input.Countries = nil
		if err := tx.Create(input).Error; err != nil {
			return err
		}
		if !enabled {
			if err := tx.Model(input).Update("enabled", false).Error; err != nil {
				return err
			}
			input.Enabled = false
		}
		if err := replaceMarketCountries(tx, input.ID, countries); err != nil {
			return err
		}
		return tx.Preload("Countries", func(db *gorm.DB) *gorm.DB {
			return db.Order("code ASC")
		}).First(input, input.ID).Error
	})
}

func (r *StorefrontMarketRepository) Update(input *market.StorefrontMarket) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"code":                  input.Code,
			"name":                  input.Name,
			"default_locale":        input.DefaultLocale,
			"supported_locales":     input.SupportedLocales,
			"default_currency":      input.DefaultCurrency,
			"display_currencies":    input.DisplayCurrencies,
			"payment_method_policy": input.PaymentMethodPolicy,
			"logistics_policy":      input.LogisticsPolicy,
			"tax_policy":            input.TaxPolicy,
			"enabled":               input.Enabled,
			"priority":              input.Priority,
		}
		if err := tx.Model(&market.StorefrontMarket{}).Where("id = ?", input.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := replaceMarketCountries(tx, input.ID, input.Countries); err != nil {
			return err
		}
		return tx.Preload("Countries", func(db *gorm.DB) *gorm.DB {
			return db.Order("code ASC")
		}).First(input, input.ID).Error
	})
}

func (r *StorefrontMarketRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("market_id = ?", id).Delete(&market.MarketCountry{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&market.StorefrontMarket{}, id).Error
	})
}

func replaceMarketCountries(tx *gorm.DB, marketID uint, countries []market.MarketCountry) error {
	if err := tx.Unscoped().Where("market_id = ?", marketID).Delete(&market.MarketCountry{}).Error; err != nil {
		return err
	}
	if len(countries) == 0 {
		return nil
	}
	for i := range countries {
		countries[i].ID = 0
		countries[i].MarketID = marketID
	}
	return tx.Create(&countries).Error
}

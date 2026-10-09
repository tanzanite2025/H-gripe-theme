package repository

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/taxrate"

	"gorm.io/gorm"
)

// TaxRateRepository reads tax rules used by checkout and tax calculation.
type TaxRateRepository struct {
	db *gorm.DB
}

func NewTaxRateRepository(db *gorm.DB) *TaxRateRepository {
	return &TaxRateRepository{db: db}
}

func (r *TaxRateRepository) WithTx(tx *gorm.DB) *TaxRateRepository {
	return &TaxRateRepository{db: tx}
}

// FindTaxRateByID 根据ID查找税率
func (r *TaxRateRepository) FindTaxRateByID(id uint) (*taxrate.TaxRate, error) {
	var tr taxrate.TaxRate
	err := r.db.First(&tr, id).Error
	if err != nil {
		return nil, err
	}
	return &tr, nil
}

// FindTaxRateByLocation 根据地区查找税率。
// postalCode 是可选的；查找按精确邮编 -> 州默认 -> 全国默认（state 为空）
// 三级回退，确保只配置全国 VAT 的国家不会被静默按零税率结算。
func (r *TaxRateRepository) FindTaxRateByLocation(country, state string, postalCodes ...string) (*taxrate.TaxRate, error) {
	var tr taxrate.TaxRate

	country = strings.ToUpper(strings.TrimSpace(country))
	state = strings.ToUpper(strings.TrimSpace(state))
	postalCode := ""
	if len(postalCodes) > 0 {
		postalCode = strings.ToUpper(strings.TrimSpace(postalCodes[0]))
	}

	if postalCode != "" {
		err := r.db.Where("country = ? AND COALESCE(state, '') = ? AND enabled = ?", country, state, true).
			Where("postal_code = ?", postalCode).
			Order("priority DESC, id ASC").
			First(&tr).Error
		if err == nil {
			return &tr, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if state != "" {
		err := r.db.Where("country = ? AND state = ? AND enabled = ?", country, state, true).
			Where("COALESCE(postal_code, '') = ''").
			Order("priority DESC, id ASC").
			First(&tr).Error
		if err == nil {
			return &tr, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// National rules are represented by an empty state (and, like state
	// defaults, an empty postal code). Keep the country predicate strict so a
	// missing country can never borrow another country's VAT rate.
	err := r.db.Where("country = ? AND COALESCE(state, '') = '' AND enabled = ?", country, true).
		Where("COALESCE(postal_code, '') = ''").
		Order("priority DESC, id ASC").
		First(&tr).Error
	if err != nil {
		return nil, err
	}
	return &tr, nil
}

// FindAllTaxRates 查找所有税率
func (r *TaxRateRepository) FindAllTaxRates(enabledOnly bool) ([]taxrate.TaxRate, error) {
	var rates []taxrate.TaxRate
	query := r.db.Order("country ASC, state ASC, postal_code ASC, priority DESC, id ASC")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	err := query.Find(&rates).Error
	return rates, err
}

// CreateTaxRate saves a checkout rule and preserves an explicitly disabled
// state despite the database default that activates new rules.
func (r *TaxRateRepository) CreateTaxRate(rate *taxrate.TaxRate) error {
	if rate == nil {
		return errors.New("tax rate is required")
	}
	enabled := rate.Enabled
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rate).Error; err != nil {
			return err
		}
		if enabled {
			return nil
		}
		return tx.Model(&taxrate.TaxRate{}).Where("id = ?", rate.ID).Update("enabled", false).Error
	})
	if err == nil {
		rate.Enabled = enabled
	}
	return err
}

// UpdateTaxRate replaces the editable fields while retaining fields owned by
// other tax workflows and applying the normal tax-rate persistence checks.
func (r *TaxRateRepository) UpdateTaxRate(rate *taxrate.TaxRate) error {
	if rate == nil || rate.ID == 0 {
		return errors.New("tax rate with an ID is required")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var savedRate taxrate.TaxRate
		if err := tx.First(&savedRate, rate.ID).Error; err != nil {
			return err
		}
		savedRate.Name = rate.Name
		savedRate.Country = rate.Country
		savedRate.State = rate.State
		savedRate.PostalCode = rate.PostalCode
		savedRate.RateDecimal = rate.RateDecimal
		savedRate.Priority = rate.Priority
		savedRate.Enabled = rate.Enabled
		if err := tx.Save(&savedRate).Error; err != nil {
			return err
		}
		*rate = savedRate
		return nil
	})
}

// DeleteTaxRate soft-deletes a checkout rule so it remains recoverable for
// audits while disappearing from all active tax lookups.
func (r *TaxRateRepository) DeleteTaxRate(id uint) error {
	result := r.db.Delete(&taxrate.TaxRate{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

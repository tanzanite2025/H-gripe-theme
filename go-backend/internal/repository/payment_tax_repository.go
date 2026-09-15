package repository

import (
	"commerce-platform/internal/domain/payment"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// TaxRate 相关方法

// FindTaxRateByID 根据ID查找税率
func (r *PaymentRepository) FindTaxRateByID(id uint) (*payment.TaxRate, error) {
	var tr payment.TaxRate
	err := r.db.First(&tr, id).Error
	if err != nil {
		return nil, err
	}
	return &tr, nil
}

// FindTaxRateByLocation 根据地区查找税率。
// postalCode 是可选的；查找按精确邮编 -> 州默认 -> 全国默认（state 为空）
// 三级回退，确保只配置全国 VAT 的国家不会被静默按零税率结算。
func (r *PaymentRepository) FindTaxRateByLocation(country, state string, postalCodes ...string) (*payment.TaxRate, error) {
	var tr payment.TaxRate

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
func (r *PaymentRepository) FindAllTaxRates(enabledOnly bool) ([]payment.TaxRate, error) {
	var rates []payment.TaxRate
	query := r.db.Order("country ASC, state ASC")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	err := query.Find(&rates).Error
	return rates, err
}

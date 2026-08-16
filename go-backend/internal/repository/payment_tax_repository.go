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
// postalCode 是可选的；提供时优先匹配精确邮编，找不到再回退到该地区的默认税率。
func (r *PaymentRepository) FindTaxRateByLocation(country, state string, postalCodes ...string) (*payment.TaxRate, error) {
	var tr payment.TaxRate

	country = strings.ToUpper(strings.TrimSpace(country))
	state = strings.ToUpper(strings.TrimSpace(state))
	postalCode := ""
	if len(postalCodes) > 0 {
		postalCode = strings.ToUpper(strings.TrimSpace(postalCodes[0]))
	}

	locationQuery := func() *gorm.DB {
		return r.db.Where("country = ? AND state = ? AND enabled = ?", country, state, true)
	}
	if postalCode != "" {
		err := locationQuery().
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

	err := locationQuery().
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

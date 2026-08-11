package repository

import "commerce-platform/internal/domain/payment"

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

// FindTaxRateByLocation 根据地区查找税率
func (r *PaymentRepository) FindTaxRateByLocation(country, state string) (*payment.TaxRate, error) {
	var tr payment.TaxRate
	err := r.db.Where("country = ? AND state = ? AND enabled = ?", country, state, true).
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

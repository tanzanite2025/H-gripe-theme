package repository

import (
	"tanzanite/internal/domain/payment"

	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// PaymentMethod 相关方法

// CreatePaymentMethod 创建支付方式
func (r *PaymentRepository) CreatePaymentMethod(pm *payment.PaymentMethod) error {
	return r.db.Create(pm).Error
}

// FindPaymentMethodByID 根据ID查找支付方式
func (r *PaymentRepository) FindPaymentMethodByID(id uint) (*payment.PaymentMethod, error) {
	var pm payment.PaymentMethod
	err := r.db.First(&pm, id).Error
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

// FindPaymentMethodByCode 根据代码查找支付方式
func (r *PaymentRepository) FindPaymentMethodByCode(code string) (*payment.PaymentMethod, error) {
	var pm payment.PaymentMethod
	err := r.db.Where("code = ?", code).First(&pm).Error
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

// FindAllPaymentMethods 查找所有支付方式
func (r *PaymentRepository) FindAllPaymentMethods(enabledOnly bool) ([]payment.PaymentMethod, error) {
	var methods []payment.PaymentMethod
	query := r.db.Order("sort_order ASC")

	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	err := query.Find(&methods).Error
	return methods, err
}

// UpdatePaymentMethod 更新支付方式
func (r *PaymentRepository) UpdatePaymentMethod(pm *payment.PaymentMethod) error {
	return r.db.Save(pm).Error
}

// DeletePaymentMethod 删除支付方式
func (r *PaymentRepository) DeletePaymentMethod(id uint) error {
	return r.db.Delete(&payment.PaymentMethod{}, id).Error
}

// TaxRate 相关方法

// CreateTaxRate 创建税率
func (r *PaymentRepository) CreateTaxRate(tr *payment.TaxRate) error {
	return r.db.Create(tr).Error
}

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
func (r *PaymentRepository) FindAllTaxRates() ([]payment.TaxRate, error) {
	var rates []payment.TaxRate
	err := r.db.Order("country ASC, state ASC").Find(&rates).Error
	return rates, err
}

// UpdateTaxRate 更新税率
func (r *PaymentRepository) UpdateTaxRate(tr *payment.TaxRate) error {
	return r.db.Save(tr).Error
}

// DeleteTaxRate 删除税率
func (r *PaymentRepository) DeleteTaxRate(id uint) error {
	return r.db.Delete(&payment.TaxRate{}, id).Error
}

// Transaction 相关方法

// CreateTransaction 创建交易记录
func (r *PaymentRepository) CreateTransaction(t *payment.Transaction) error {
	return r.db.Create(t).Error
}

// FindTransactionByID 根据ID查找交易
func (r *PaymentRepository) FindTransactionByID(id uint) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindTransactionByOrderID 根据订单ID查找交易
func (r *PaymentRepository) FindTransactionByOrderID(orderID uint) ([]payment.Transaction, error) {
	var transactions []payment.Transaction
	err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

// FindTransactionByTransactionID 根据交易ID查找
func (r *PaymentRepository) FindTransactionByTransactionID(transactionID string) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.db.Where("transaction_id = ?", transactionID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTransaction 更新交易
func (r *PaymentRepository) UpdateTransaction(t *payment.Transaction) error {
	return r.db.Save(t).Error
}

// Refund 相关方法

// CreateRefund 创建退款记录
func (r *PaymentRepository) CreateRefund(rf *payment.Refund) error {
	return r.db.Create(rf).Error
}

// FindRefundByID 根据ID查找退款
func (r *PaymentRepository) FindRefundByID(id uint) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.db.First(&rf, id).Error
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

// FindRefundsByOrderID 根据订单ID查找退款
func (r *PaymentRepository) FindRefundsByOrderID(orderID uint) ([]payment.Refund, error) {
	var refunds []payment.Refund
	err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").Find(&refunds).Error
	return refunds, err
}

// UpdateRefund 更新退款
func (r *PaymentRepository) UpdateRefund(rf *payment.Refund) error {
	return r.db.Save(rf).Error
}

package repository

import (
	"tanzanite/internal/domain/payment"
)

// PaymentMethod 相关方法

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

// CreatePaymentMethod 创建支付方式
func (r *PaymentRepository) CreatePaymentMethod(pm *payment.PaymentMethod) error {
	enabled := pm.Enabled
	if err := r.db.Create(pm).Error; err != nil {
		return err
	}
	if !enabled {
		if err := r.db.Model(pm).Update("enabled", false).Error; err != nil {
			return err
		}
		pm.Enabled = false
	}
	return nil
}

// UpdatePaymentMethod 更新支付方式
func (r *PaymentRepository) UpdatePaymentMethod(pm *payment.PaymentMethod) error {
	return r.db.Save(pm).Error
}

// DeletePaymentMethod 删除支付方式
func (r *PaymentRepository) DeletePaymentMethod(id uint) error {
	return r.db.Delete(&payment.PaymentMethod{}, id).Error
}

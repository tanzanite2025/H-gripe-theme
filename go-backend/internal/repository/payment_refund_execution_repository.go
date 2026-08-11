package repository

import (
	paymentdomain "commerce-platform/internal/domain/payment"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRefundExecutionRepository struct {
	db *gorm.DB
}

func NewPaymentRefundExecutionRepository(db *gorm.DB) *PaymentRefundExecutionRepository {
	return &PaymentRefundExecutionRepository{db: db}
}

func (r *PaymentRefundExecutionRepository) WithTx(tx *gorm.DB) *PaymentRefundExecutionRepository {
	return &PaymentRefundExecutionRepository{db: tx}
}

func (r *PaymentRefundExecutionRepository) lockForUpdate(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

func (r *PaymentRefundExecutionRepository) Create(execution *paymentdomain.PaymentRefundExecution) error {
	return r.db.Create(execution).Error
}

func (r *PaymentRefundExecutionRepository) FindByRefundIDForUpdate(refundID uint) (*paymentdomain.PaymentRefundExecution, error) {
	var execution paymentdomain.PaymentRefundExecution
	err := r.lockForUpdate(r.db).Where("refund_id = ?", refundID).First(&execution).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

func (r *PaymentRefundExecutionRepository) Update(execution *paymentdomain.PaymentRefundExecution) error {
	return r.db.Save(execution).Error
}

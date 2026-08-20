package repository

import (
	"commerce-platform/internal/domain/aftersales"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AfterSalesRefundReviewRepository struct {
	db *gorm.DB
}

func NewAfterSalesRefundReviewRepository(db *gorm.DB) *AfterSalesRefundReviewRepository {
	return &AfterSalesRefundReviewRepository{db: db}
}

func (r *AfterSalesRefundReviewRepository) WithTx(tx *gorm.DB) *AfterSalesRefundReviewRepository {
	return &AfterSalesRefundReviewRepository{db: tx}
}

func (r *AfterSalesRefundReviewRepository) Transaction(
	callback func(*AfterSalesRefundReviewRepository) error,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return callback(r.WithTx(tx))
	})
}

func (r *AfterSalesRefundReviewRepository) lockForUpdate(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

func (r *AfterSalesRefundReviewRepository) FindCaseByIDForUpdate(id uint) (*aftersales.AfterSalesCase, error) {
	var record aftersales.AfterSalesCase
	err := r.lockForUpdate(r.db).
		Preload("Items").
		First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *AfterSalesRefundReviewRepository) FindByCaseID(caseID uint) (*aftersales.AfterSalesRefundReview, error) {
	var record aftersales.AfterSalesRefundReview
	err := r.db.Where("case_id = ?", caseID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *AfterSalesRefundReviewRepository) FindByCaseIDForUpdate(caseID uint) (*aftersales.AfterSalesRefundReview, error) {
	var record aftersales.AfterSalesRefundReview
	err := r.lockForUpdate(r.db).Where("case_id = ?", caseID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *AfterSalesRefundReviewRepository) Create(review *aftersales.AfterSalesRefundReview) error {
	return r.db.Create(review).Error
}

func (r *AfterSalesRefundReviewRepository) Update(review *aftersales.AfterSalesRefundReview) error {
	return r.db.Save(review).Error
}

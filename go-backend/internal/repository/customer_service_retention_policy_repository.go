package repository

import (
	"commerce-platform/internal/domain/ticket"
	"errors"

	"gorm.io/gorm"
)

type CustomerServiceRetentionPolicyRepository struct{ db *gorm.DB }

func NewCustomerServiceRetentionPolicyRepository(db *gorm.DB) *CustomerServiceRetentionPolicyRepository {
	return &CustomerServiceRetentionPolicyRepository{db: db}
}

func (r *CustomerServiceRetentionPolicyRepository) Get() (*ticket.CustomerServiceRetentionPolicy, error) {
	var policy ticket.CustomerServiceRetentionPolicy
	err := r.db.Order("id ASC").First(&policy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *CustomerServiceRetentionPolicyRepository) Save(policy *ticket.CustomerServiceRetentionPolicy) error {
	if policy.ID == 0 {
		return r.db.Create(policy).Error
	}
	return r.db.Save(policy).Error
}

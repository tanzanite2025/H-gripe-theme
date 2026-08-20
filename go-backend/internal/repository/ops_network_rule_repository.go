package repository

import (
	"commerce-platform/internal/domain/ops"

	"gorm.io/gorm"
)

type OpsNetworkRuleRepository struct {
	db *gorm.DB
}

func NewOpsNetworkRuleRepository(db *gorm.DB) *OpsNetworkRuleRepository {
	return &OpsNetworkRuleRepository{db: db}
}

func (r *OpsNetworkRuleRepository) ListByEnvironment(environment string) ([]ops.NetworkRule, error) {
	var records []ops.NetworkRule
	query := r.db
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}
	err := query.
		Order("enabled DESC").
		Order("environment ASC").
		Order("managed_by ASC").
		Order("scope ASC").
		Order("name ASC").
		Find(&records).Error
	return records, err
}

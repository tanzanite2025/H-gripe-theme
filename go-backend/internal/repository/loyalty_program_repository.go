package repository

import (
	"commerce-platform/internal/domain/loyalty"

	"gorm.io/gorm"
)

type LoyaltyProgramRepository struct {
	db *gorm.DB
}

func NewLoyaltyProgramRepository(db *gorm.DB) *LoyaltyProgramRepository {
	return &LoyaltyProgramRepository{db: db}
}

func (r *LoyaltyProgramRepository) WithTx(tx *gorm.DB) *LoyaltyProgramRepository {
	return &LoyaltyProgramRepository{db: tx}
}

func (r *LoyaltyProgramRepository) FindActive() (*loyalty.ProgramConfig, error) {
	var config loyalty.ProgramConfig
	err := r.db.Where("status = ?", "active").
		Order("version DESC").
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *LoyaltyProgramRepository) CreateVersion(config *loyalty.ProgramConfig) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if tx.Name() == "postgres" {
			if err := tx.Exec("LOCK TABLE loyalty_program_configs IN EXCLUSIVE MODE").Error; err != nil {
				return err
			}
		}

		var latestVersion int
		if err := tx.Model(&loyalty.ProgramConfig{}).
			Select("COALESCE(MAX(version), 0)").
			Scan(&latestVersion).Error; err != nil {
			return err
		}

		if err := tx.Model(&loyalty.ProgramConfig{}).
			Where("status = ?", "active").
			Update("status", "archived").Error; err != nil {
			return err
		}

		config.Version = latestVersion + 1
		config.Status = "active"
		if err := tx.Create(config).Error; err != nil {
			return err
		}
		return nil
	})
}

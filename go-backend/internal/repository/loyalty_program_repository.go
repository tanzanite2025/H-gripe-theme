package repository

import (
	"tanzanite/internal/domain/loyalty"

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
	err := r.db.
		Preload("RedeemOptions", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Where("status = ?", "active").
		Order("version DESC").
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *LoyaltyProgramRepository) CreateVersion(config *loyalty.ProgramConfig) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("LOCK TABLE loyalty_program_configs, loyalty_program_redeem_options IN EXCLUSIVE MODE").Error; err != nil {
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
		options := config.RedeemOptions
		config.RedeemOptions = nil

		if err := tx.Create(config).Error; err != nil {
			return err
		}

		for index := range options {
			options[index].ID = 0
			options[index].ConfigID = config.ID
			options[index].SortOrder = index
			if err := tx.Create(&options[index]).Error; err != nil {
				return err
			}
		}

		config.RedeemOptions = options
		return nil
	})
}

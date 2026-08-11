package repository

import (
	"errors"
	"commerce-platform/internal/domain/loyalty"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrRedeemOptionOutOfStock = errors.New("redeem gift card option is out of stock")

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
		if tx.Name() == "postgres" {
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

func (r *LoyaltyProgramRepository) ConsumeRedeemOption(configID, optionID uint) (*loyalty.ProgramRedeemOption, error) {
	var option loyalty.ProgramRedeemOption
	if err := r.db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND config_id = ?", optionID, configID).
		First(&option).Error; err != nil {
		return nil, err
	}
	if option.RemainingQuantity() <= 0 {
		return nil, ErrRedeemOptionOutOfStock
	}

	result := r.db.Model(&loyalty.ProgramRedeemOption{}).
		Where("id = ? AND config_id = ? AND redeemed_quantity < stock_quantity", optionID, configID).
		UpdateColumn("redeemed_quantity", gorm.Expr("redeemed_quantity + 1"))
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrRedeemOptionOutOfStock
	}

	option.RedeemedQuantity++
	return &option, nil
}

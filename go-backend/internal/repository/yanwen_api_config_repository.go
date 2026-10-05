package repository

import (
	"errors"

	"commerce-platform/internal/domain/shipping"

	"gorm.io/gorm"
)

type YanwenAPIConfigRepository struct{ db *gorm.DB }

func NewYanwenAPIConfigRepository(db *gorm.DB) *YanwenAPIConfigRepository {
	return &YanwenAPIConfigRepository{db: db}
}

func (r *YanwenAPIConfigRepository) FindYanwenAPIConfigByEnvironment(environment string) (*shipping.YanwenAPIConfig, error) {
	var config shipping.YanwenAPIConfig
	if err := r.db.Where("environment = ?", environment).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *YanwenAPIConfigRepository) SaveYanwenAPIConfig(config *shipping.YanwenAPIConfig) error {
	var existing shipping.YanwenAPIConfig
	err := r.db.Where("environment = ?", config.Environment).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(config).Error
	}
	if err != nil {
		return err
	}
	config.ID = existing.ID
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"endpoint":            config.Endpoint,
		"user_id_encrypted":   config.UserIDEncrypted,
		"api_token_encrypted": config.APITokenEncrypted,
		"enabled":             config.Enabled,
	}).Error
}

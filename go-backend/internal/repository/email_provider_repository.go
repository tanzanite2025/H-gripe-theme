package repository

import (
	"commerce-platform/internal/domain/notification"
	"time"

	"gorm.io/gorm"
)

type EmailProviderRepository struct{ db *gorm.DB }

func NewEmailProviderRepository(db *gorm.DB) *EmailProviderRepository {
	return &EmailProviderRepository{db: db}
}

func (r *EmailProviderRepository) List() ([]notification.EmailProviderConfig, error) {
	var records []notification.EmailProviderConfig
	err := r.db.Order("is_default DESC").Order("is_active DESC").Order("name ASC").Find(&records).Error
	return records, err
}

func (r *EmailProviderRepository) FindByID(id uint) (*notification.EmailProviderConfig, error) {
	var record notification.EmailProviderConfig
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *EmailProviderRepository) FindByCode(code string) (*notification.EmailProviderConfig, error) {
	var record notification.EmailProviderConfig
	if err := r.db.Where("code = ?", code).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *EmailProviderRepository) FindDefaultActive() (*notification.EmailProviderConfig, error) {
	var record notification.EmailProviderConfig
	if err := r.db.Where("is_default = ? AND is_active = ?", true, true).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *EmailProviderRepository) Create(record *notification.EmailProviderConfig) error {
	return r.db.Create(record).Error
}

func (r *EmailProviderRepository) Update(record *notification.EmailProviderConfig) error {
	return r.db.Model(&notification.EmailProviderConfig{}).Where("id = ?", record.ID).Updates(map[string]interface{}{
		"code": record.Code, "name": record.Name, "driver": record.Driver, "host": record.Host,
		"port": record.Port, "username": record.Username, "password_encrypted": record.PasswordEncrypted,
		"from_name": record.FromName, "from_email": record.FromEmail, "reply_to": record.ReplyTo,
		"encryption_type": record.EncryptionType, "is_active": record.IsActive, "is_default": record.IsDefault,
		"last_tested_at": record.LastTestedAt, "last_test_status": record.LastTestStatus, "last_test_error": record.LastTestError,
	}).Error
}

func (r *EmailProviderRepository) SetDefault(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&notification.EmailProviderConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&notification.EmailProviderConfig{}).Where("id = ?", id).Updates(map[string]interface{}{"is_default": true, "is_active": true}).Error
	})
}

func (r *EmailProviderRepository) UpdateTestState(id uint, status, testError string, testedAt time.Time) error {
	return r.db.Model(&notification.EmailProviderConfig{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_tested_at": testedAt, "last_test_status": status, "last_test_error": testError,
	}).Error
}

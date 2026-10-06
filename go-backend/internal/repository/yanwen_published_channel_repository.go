package repository

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"

	"gorm.io/gorm"
)

// YanwenPublishedChannelRepository owns the curated Yanwen collection. It is
// deliberately separate from ShippingRepository so generic shipping data
// access cannot accidentally become an integration with Yanwen's website.
type YanwenPublishedChannelRepository struct {
	db *gorm.DB
}

func NewYanwenPublishedChannelRepository(db *gorm.DB) *YanwenPublishedChannelRepository {
	return &YanwenPublishedChannelRepository{db: db}
}

func (r *YanwenPublishedChannelRepository) FindYanwenPublishedChannelsByEnvironment(
	environment string,
	enabledOnly bool,
) ([]shipping.YanwenPublishedChannel, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("Yanwen published channel repository is not configured")
	}
	var channels []shipping.YanwenPublishedChannel
	query := r.db.Order("display_name ASC").Order("id ASC")
	if environment = strings.TrimSpace(environment); environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	return channels, query.Find(&channels).Error
}

func (r *YanwenPublishedChannelRepository) FindYanwenPublishedChannelByID(id uint) (*shipping.YanwenPublishedChannel, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("Yanwen published channel repository is not configured")
	}
	var channel shipping.YanwenPublishedChannel
	if err := r.db.First(&channel, id).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *YanwenPublishedChannelRepository) CreateYanwenPublishedChannel(channel *shipping.YanwenPublishedChannel) error {
	if r == nil || r.db == nil {
		return errors.New("Yanwen published channel repository is not configured")
	}
	return r.db.Create(channel).Error
}

func (r *YanwenPublishedChannelRepository) UpdateYanwenPublishedChannel(channel *shipping.YanwenPublishedChannel) error {
	if r == nil || r.db == nil {
		return errors.New("Yanwen published channel repository is not configured")
	}
	result := r.db.Model(&shipping.YanwenPublishedChannel{}).
		Where("id = ? AND deleted_at IS NULL", channel.ID).
		UpdateColumns(map[string]interface{}{
			"environment":                 channel.Environment,
			"product_code":                channel.ProductCode,
			"display_name":                channel.DisplayName,
			"countries":                   channel.Countries,
			"package_type":                channel.PackageType,
			"max_weight_grams":            channel.MaxWeightGrams,
			"volumetric_divisor":          channel.VolumetricDivisor,
			"require_receiver_tax_number": channel.RequireReceiverTaxNumber,
			"require_ioss":                channel.RequireIOSS,
			"require_eori":                channel.RequireEORI,
			"notes":                       channel.Notes,
			"enabled":                     channel.Enabled,
			"updated_at":                  time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *YanwenPublishedChannelRepository) DeleteYanwenPublishedChannel(id uint) error {
	if r == nil || r.db == nil {
		return errors.New("Yanwen published channel repository is not configured")
	}
	result := r.db.Delete(&shipping.YanwenPublishedChannel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DisableMissingOfficialProducts stops enabled channels whose product code is
// absent from the latest official catalog for the same environment.
func (r *YanwenPublishedChannelRepository) DisableMissingOfficialProducts(
	environment string,
	productCodes []string,
	updatedAt time.Time,
) (int64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("Yanwen published channel repository is not configured")
	}
	environment = strings.TrimSpace(environment)
	if environment == "" {
		return 0, errors.New("Yanwen environment is required")
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	query := r.db.Model(&shipping.YanwenPublishedChannel{}).
		Where("environment = ? AND enabled = ? AND deleted_at IS NULL", environment, true)
	normalizedProductCodes := make([]string, 0, len(productCodes))
	for _, productCode := range productCodes {
		productCode = strings.TrimSpace(productCode)
		if productCode != "" {
			normalizedProductCodes = append(normalizedProductCodes, productCode)
		}
	}
	if len(normalizedProductCodes) > 0 {
		query = query.Where("product_code NOT IN ?", normalizedProductCodes)
	}
	result := query.UpdateColumns(map[string]interface{}{
		"enabled":    false,
		"updated_at": updatedAt.UTC(),
	})
	return result.RowsAffected, result.Error
}

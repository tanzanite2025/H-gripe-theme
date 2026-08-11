package repository

import (
	"commerce-platform/internal/domain/merchant"
	"time"

	"gorm.io/gorm"
)

type GoogleMerchantRepository struct {
	db *gorm.DB
}

func NewGoogleMerchantRepository(db *gorm.DB) *GoogleMerchantRepository {
	return &GoogleMerchantRepository{db: db}
}

func (r *GoogleMerchantRepository) FindConnection() (*merchant.GoogleMerchantConnection, error) {
	var connection merchant.GoogleMerchantConnection
	if err := r.db.Where("provider = ?", merchant.GoogleMerchantProvider).First(&connection).Error; err != nil {
		return nil, err
	}
	return &connection, nil
}

func (r *GoogleMerchantRepository) SaveConnection(connection *merchant.GoogleMerchantConnection) error {
	return r.db.Save(connection).Error
}

func (r *GoogleMerchantRepository) ConsumeOAuthState(stateHash string, now time.Time) (*merchant.GoogleMerchantConnection, error) {
	var connection merchant.GoogleMerchantConnection
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("provider = ? AND oauth_state_hash = ? AND oauth_state_expires_at > ?", merchant.GoogleMerchantProvider, stateHash, now).
			First(&connection).Error; err != nil {
			return err
		}

		result := tx.Model(&merchant.GoogleMerchantConnection{}).
			Where("id = ? AND oauth_state_hash = ?", connection.ID, stateHash).
			Updates(map[string]interface{}{
				"oauth_state_hash":           "",
				"oauth_state_expires_at":     nil,
				"oauth_initiated_by_user_id": nil,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		connection.OAuthStateHash = ""
		connection.OAuthStateExpiresAt = nil
		connection.OAuthInitiatedByUserID = nil
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &connection, nil
}

func (r *GoogleMerchantRepository) ListOffers() ([]merchant.GoogleMerchantOffer, error) {
	var offers []merchant.GoogleMerchantOffer
	err := r.db.Preload("Product").Preload("Variant").Order("updated_at DESC, id DESC").Find(&offers).Error
	return offers, err
}

func (r *GoogleMerchantRepository) ListOffersByProductID(productID uint) ([]merchant.GoogleMerchantOffer, error) {
	var offers []merchant.GoogleMerchantOffer
	err := r.db.
		Preload("Product").
		Preload("Variant").
		Where("product_id = ?", productID).
		Order("id ASC").
		Find(&offers).Error
	return offers, err
}

func (r *GoogleMerchantRepository) FindOfferByID(id uint) (*merchant.GoogleMerchantOffer, error) {
	var offer merchant.GoogleMerchantOffer
	if err := r.db.Preload("Product").Preload("Variant").First(&offer, id).Error; err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *GoogleMerchantRepository) FindOfferByVariantID(variantID uint) (*merchant.GoogleMerchantOffer, error) {
	var offer merchant.GoogleMerchantOffer
	if err := r.db.Preload("Product").Preload("Variant").Where("variant_id = ?", variantID).First(&offer).Error; err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *GoogleMerchantRepository) CreateOffer(offer *merchant.GoogleMerchantOffer) error {
	return r.db.Create(offer).Error
}

func (r *GoogleMerchantRepository) UpdateOffer(offer *merchant.GoogleMerchantOffer) error {
	return r.db.Save(offer).Error
}

func (r *GoogleMerchantRepository) UpdateOfferSyncState(id uint, status string, lastSyncAt *time.Time, lastError string) error {
	updates := map[string]interface{}{
		"sync_status": status,
		"last_error":  lastError,
	}
	if lastSyncAt != nil {
		updates["last_sync_at"] = lastSyncAt
	}
	return r.db.Model(&merchant.GoogleMerchantOffer{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *GoogleMerchantRepository) CountOffersWithRemoteSync() (int64, error) {
	var count int64
	err := r.db.Model(&merchant.GoogleMerchantOffer{}).
		Where("last_sync_at IS NOT NULL AND sync_status <> ?", "removed").
		Count(&count).Error
	return count, err
}

func (r *GoogleMerchantRepository) DeleteOffer(id uint) error {
	return r.db.Delete(&merchant.GoogleMerchantOffer{}, id).Error
}

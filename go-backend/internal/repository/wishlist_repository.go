package repository

import (
	"tanzanite/internal/domain/wishlist"

	"gorm.io/gorm"
)

type WishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) *WishlistRepository {
	return &WishlistRepository{db: db}
}

func (r *WishlistRepository) ListByUserID(userID uint) ([]wishlist.Item, error) {
	var items []wishlist.Item
	err := r.db.Preload("Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *WishlistRepository) FindByID(id uint) (*wishlist.Item, error) {
	var item wishlist.Item
	err := r.db.Preload("Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *WishlistRepository) FindByUserAndProduct(userID, productID uint) (*wishlist.Item, error) {
	var item wishlist.Item
	err := r.db.Preload("Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *WishlistRepository) Create(item *wishlist.Item) error {
	return r.db.Create(item).Error
}

func (r *WishlistRepository) Delete(id uint) error {
	return r.db.Delete(&wishlist.Item{}, id).Error
}

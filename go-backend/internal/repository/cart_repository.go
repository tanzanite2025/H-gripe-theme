package repository

import (
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
)

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

// FindByUserID 根据用户ID查找购物车
func (r *CartRepository) FindByUserID(userID uint) (*product.Cart, error) {
	var cart product.Cart
	err := r.db.Preload("Items.Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).Preload("Items.Variant").Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// FindBySessionID 根据会话ID查找购物车
func (r *CartRepository) FindBySessionID(sessionID string) (*product.Cart, error) {
	var cart product.Cart
	err := r.db.Preload("Items.Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).Preload("Items.Variant").Where("session_id = ?", sessionID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// Create 创建购物车
func (r *CartRepository) Create(cart *product.Cart) error {
	return r.db.Create(cart).Error
}

// Update 更新购物车
func (r *CartRepository) Update(cart *product.Cart) error {
	return r.db.Save(cart).Error
}

// AddItem 添加商品到购物车
func (r *CartRepository) AddItem(item *product.CartItem) error {
	return r.db.Create(item).Error
}

// UpdateItem 更新购物车项目
func (r *CartRepository) UpdateItem(item *product.CartItem) error {
	return r.db.Save(item).Error
}

// RemoveItem 从购物车移除商品
func (r *CartRepository) RemoveItem(itemID uint) error {
	return r.db.Delete(&product.CartItem{}, itemID).Error
}

// FindItem 查找购物车项目
func (r *CartRepository) FindItem(cartID, productID uint, variantID *uint) (*product.CartItem, error) {
	var item product.CartItem
	query := r.db.Where("cart_id = ? AND product_id = ?", cartID, productID)
	if variantID != nil {
		query = query.Where("variant_id = ?", *variantID)
	} else {
		query = query.Where("variant_id IS NULL")
	}
	err := query.First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ClearCart 清空购物车
func (r *CartRepository) ClearCart(cartID uint) error {
	return r.db.Where("cart_id = ?", cartID).Delete(&product.CartItem{}).Error
}

// GetSummary 获取购物车摘要
func (r *CartRepository) GetSummary(cartID uint) (*product.CartSummary, error) {
	var items []product.CartItem
	err := r.db.Preload("Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).Preload("Variant").Where("cart_id = ?", cartID).Find(&items).Error
	if err != nil {
		return nil, err
	}

	summary := &product.CartSummary{
		ItemCount: 0,
		Total:     0,
		Items:     items,
	}

	for _, item := range items {
		summary.ItemCount += item.Quantity
		summary.Total += item.Price * float64(item.Quantity)
	}

	return summary, nil
}

// BulkUpsertItems 批量插入或更新购物车项目
func (r *CartRepository) BulkUpsertItems(items []product.CartItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			repo := &CartRepository{db: tx}
			existing, err := repo.FindItem(item.CartID, item.ProductID, item.VariantID)
			if err == nil {
				existing.Quantity += item.Quantity
				existing.Price = item.Price
				if err := tx.Save(existing).Error; err != nil {
					return err
				}
				continue
			}
			if !IsRecordNotFound(err) {
				return err
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

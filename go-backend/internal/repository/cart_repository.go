package repository

import (
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	err := r.db.Preload("Items.Product.Images").Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// FindBySessionID 根据会话ID查找购物车
func (r *CartRepository) FindBySessionID(sessionID string) (*product.Cart, error) {
	var cart product.Cart
	err := r.db.Preload("Items.Product.Images").Where("session_id = ?", sessionID).First(&cart).Error
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
func (r *CartRepository) FindItem(cartID, productID uint) (*product.CartItem, error) {
	var item product.CartItem
	err := r.db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
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
	err := r.db.Preload("Product").Preload("Product.Categories").Preload("Product.Images").Where("cart_id = ?", cartID).Find(&items).Error
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
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cart_id"}, {Name: "product_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity": gorm.Expr("cart_items.quantity + EXCLUDED.quantity"),
		}),
	}).Create(&items).Error
}

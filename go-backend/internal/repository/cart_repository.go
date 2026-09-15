package repository

import (
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/product"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

// WithTx reuses the transaction database handle for checkout and rollback
// operations that must commit or roll back with the order.
func (r *CartRepository) WithTx(tx *gorm.DB) *CartRepository {
	return &CartRepository{db: tx}
}

func (r *CartRepository) lockForUpdate(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

// FindByUserID 根据用户ID查找购物车
func (r *CartRepository) FindByUserID(userID uint) (*product.Cart, error) {
	if _, err := r.PruneInvalidVariantItemsByCartIdentity("user_id = ?", userID); err != nil {
		return nil, err
	}
	var cart product.Cart
	err := r.db.Preload("Items.Product.Brand").Preload("Items.Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).Preload("Items.Variant").Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// FindAuthenticatedUserCartByIDForUpdate locks the authenticated user's cart
// row so only one checkout transaction can consume that cart at a time.
func (r *CartRepository) FindAuthenticatedUserCartByIDForUpdate(cartID, userID uint) (*product.Cart, error) {
	var cart product.Cart
	err := r.lockForUpdate(r.db).
		Where("id = ? AND user_id = ?", cartID, userID).
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// FindCheckoutCartItemsByIDForUpdate locks and returns the exact cart-item
// snapshot consumed by a checkout transaction.
func (r *CartRepository) FindCheckoutCartItemsByIDForUpdate(cartID uint) ([]product.CartItem, error) {
	if _, err := r.PruneInvalidVariantItems(cartID); err != nil {
		return nil, err
	}
	var items []product.CartItem
	err := r.lockForUpdate(r.db.Model(&product.CartItem{})).
		Where("cart_id = ?", cartID).
		Order("id ASC").
		Find(&items).Error
	return items, err
}

// FindBySessionID 根据会话ID查找购物车
func (r *CartRepository) FindBySessionID(sessionID string) (*product.Cart, error) {
	if _, err := r.PruneInvalidVariantItemsByCartIdentity("session_id = ?", sessionID); err != nil {
		return nil, err
	}
	var cart product.Cart
	err := r.db.Preload("Items.Product.Brand").Preload("Items.Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).Preload("Items.Variant").Where("session_id = ?", sessionID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// PruneInvalidVariantItems removes cart rows whose variant is no longer
// purchasable. Variant-backed rows are valid only while the variant exists,
// belongs to the cart row's product, is active, and has not been soft-deleted.
// Product-level cart rows (variant_id IS NULL) are intentionally preserved.
func (r *CartRepository) PruneInvalidVariantItems(cartID uint) (int64, error) {
	if cartID == 0 {
		return 0, nil
	}
	result := r.db.
		Where("cart_id = ? AND variant_id IS NOT NULL AND NOT EXISTS (?)", cartID,
			r.db.Model(&product.ProductVariant{}).
				Select("1").
				Where("product_variants.id = cart_items.variant_id").
				Where("product_variants.product_id = cart_items.product_id").
				Where("product_variants.is_active = ?", true)).
		Delete(&product.CartItem{})
	return result.RowsAffected, result.Error
}

func (r *CartRepository) PruneInvalidVariantItemsByCartIdentity(condition string, value interface{}) (int64, error) {
	var cartIDs []uint
	if err := r.db.Model(&product.Cart{}).Where(condition, value).Pluck("id", &cartIDs).Error; err != nil {
		return 0, err
	}
	var removed int64
	for _, cartID := range cartIDs {
		count, err := r.PruneInvalidVariantItems(cartID)
		if err != nil {
			return removed, err
		}
		removed += count
	}
	return removed, nil
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
	return r.FindItemWithConfiguration(cartID, productID, variantID, product.DefaultConfigurationHash)
}

func (r *CartRepository) FindItemWithConfiguration(cartID, productID uint, variantID *uint, configurationHash string) (*product.CartItem, error) {
	var item product.CartItem
	query := r.db.Where("cart_id = ? AND product_id = ?", cartID, productID)
	if variantID != nil {
		query = query.Where("variant_id = ?", *variantID)
	} else {
		query = query.Where("variant_id IS NULL")
	}
	if strings.TrimSpace(configurationHash) == "" {
		configurationHash = product.DefaultConfigurationHash
	}
	query = query.Where("configuration_hash = ? OR (configuration_hash = '' AND ? = ?)", configurationHash, configurationHash, product.DefaultConfigurationHash)
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

// ConsumeLockedCheckoutCartItemRows deletes only the item rows that were
// locked and quoted. The caller must compare the returned row count with the
// locked snapshot before allowing the order transaction to commit.
func (r *CartRepository) ConsumeLockedCheckoutCartItemRows(cartID uint, itemIDs []uint) (int64, error) {
	if len(itemIDs) == 0 {
		return 0, nil
	}
	result := r.db.
		Where("cart_id = ? AND id IN ?", cartID, itemIDs).
		Delete(&product.CartItem{})
	return result.RowsAffected, result.Error
}

// RestoreConsumedOrderItemsToOriginalCheckoutCart returns an unpaid order's
// consumed items to its original cart. Existing matching variant rows are
// incremented by a database-level upsert, so concurrent cart edits cannot
// lose quantities or create duplicate variant rows.
func (r *CartRepository) RestoreConsumedOrderItemsToOriginalCheckoutCart(cartID uint, items []order.OrderItem, currency string) error {
	if cartID == 0 {
		return errors.New("checkout cart id is required to restore order items")
	}
	if len(items) == 0 {
		return nil
	}

	currency = strings.TrimSpace(currency)
	if currency == "" {
		currency = product.DefaultPriceCurrency
	}

	for _, item := range items {
		if item.ProductID == 0 || item.VariantID == nil || *item.VariantID == 0 {
			return errors.New("order item product and variant are required to restore checkout cart")
		}
		cartItem := &product.CartItem{
			CartID:            cartID,
			ProductID:         item.ProductID,
			VariantID:         item.VariantID,
			Quantity:          item.Quantity,
			Currency:          currency,
			ConfigurationData: product.DefaultConfigurationJSONBytes(),
			ConfigurationHash: product.DefaultConfigurationHash,
		}
		priceMoney, priceErr := domainmoney.FromMajorFloat(item.Price, cartItem.Currency)
		if priceErr != nil {
			return fmt.Errorf("order item price: %w", priceErr)
		}
		cartItem.PriceMinor = priceMoney.AmountMinor()
		if item.ConfigurationHash != "" {
			cartItem.ConfigurationHash = item.ConfigurationHash
		} else if len(item.ConfigurationSnapshotData) > 0 && string(item.ConfigurationSnapshotData) != "{}" {
			var snapshot struct {
				ConfigurationHash       string          `json:"configuration_hash"`
				NormalizedConfiguration json.RawMessage `json:"normalized_configuration"`
			}
			if json.Unmarshal(item.ConfigurationSnapshotData, &snapshot) == nil && strings.TrimSpace(snapshot.ConfigurationHash) != "" {
				cartItem.ConfigurationHash = strings.TrimSpace(snapshot.ConfigurationHash)
				if len(snapshot.NormalizedConfiguration) > 0 {
					cartItem.ConfigurationData = append([]byte(nil), snapshot.NormalizedConfiguration...)
				}
			} else {
				cartItem.ConfigurationData = append([]byte(nil), item.ConfigurationSnapshotData...)
				digest := sha256.Sum256(cartItem.ConfigurationData)
				cartItem.ConfigurationHash = fmt.Sprintf("%x", digest[:])
			}
		}

		if err := r.db.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "cart_id"},
					{Name: "product_id"},
					{Name: "variant_id"},
					{Name: "configuration_hash"},
				},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"quantity":    gorm.Expr("quantity + ?", item.Quantity),
					"price_minor": cartItem.PriceMinor,
					"currency":    cartItem.Currency,
					"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
				}),
			}).
			Create(cartItem).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetSummary 获取购物车摘要
func (r *CartRepository) GetSummary(cartID uint) (*product.CartSummary, error) {
	if _, err := r.PruneInvalidVariantItems(cartID); err != nil {
		return nil, err
	}
	var items []product.CartItem
	err := r.db.Preload("Product.Media", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_media.sort_order ASC, product_media.id ASC")
	}).Preload("Variant").Where("cart_id = ?", cartID).Find(&items).Error
	if err != nil {
		return nil, err
	}

	summary := &product.CartSummary{
		ItemCount: 0,
		Items:     items,
	}

	var totalMoney domainmoney.Money
	initialized := false
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("cart item %d quantity must be greater than zero", item.ID)
		}
		unitMoney, err := item.PriceMoney()
		if err != nil {
			return nil, fmt.Errorf("cart item %d price: %w", item.ID, err)
		}
		lineMoney, err := unitMoney.MultiplyInt(int64(item.Quantity))
		if err != nil {
			return nil, fmt.Errorf("calculate cart item %d total: %w", item.ID, err)
		}
		summary.ItemCount += item.Quantity
		if !initialized {
			totalMoney = lineMoney
			initialized = true
		} else if totalMoney, err = totalMoney.Add(lineMoney); err != nil {
			return nil, fmt.Errorf("cart contains mixed currencies: %w", err)
		}
	}
	if initialized {
		summary.TotalMoney = totalMoney
		summary.Total, err = totalMoney.MajorFloat()
		if err != nil {
			return nil, fmt.Errorf("format cart total: %w", err)
		}
	} else {
		summary.TotalMoney = domainmoney.MustNew(0, product.DefaultPriceCurrency)
		summary.Total = 0
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
			if strings.TrimSpace(item.ConfigurationHash) == "" {
				item.ConfigurationHash = product.DefaultConfigurationHash
			}
			if len(item.ConfigurationData) == 0 {
				item.ConfigurationData = product.DefaultConfigurationJSONBytes()
			}
			repo := &CartRepository{db: tx}
			existing, err := repo.FindItemWithConfiguration(item.CartID, item.ProductID, item.VariantID, item.ConfigurationHash)
			if err == nil {
				existing.Quantity += item.Quantity
				existing.PriceMinor = item.PriceMinor
				existing.Currency = item.Currency
				existing.ConfigurationData = item.ConfigurationData
				existing.ConfigurationHash = item.ConfigurationHash
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

// MergeGuestCartOnLogin atomically moves a guest cart into the authenticated
// user's cart. Matching product/variant rows are incremented so a cart that
// was edited in two sessions cannot silently lose quantity.
func (r *CartRepository) MergeGuestCartOnLogin(userID uint, sessionID string) error {
	if userID == 0 || strings.TrimSpace(sessionID) == "" {
		return nil
	}
	sessionID = strings.TrimSpace(sessionID)
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &CartRepository{db: tx}
		var guest product.Cart
		if err := txRepo.lockForUpdate(tx.Where("user_id IS NULL AND session_id = ?", sessionID)).First(&guest).Error; err != nil {
			if IsRecordNotFound(err) {
				return nil
			}
			return err
		}

		var userCart product.Cart
		err := txRepo.lockForUpdate(tx.Where("user_id = ?", userID)).First(&userCart).Error
		if IsRecordNotFound(err) {
			userCart = product.Cart{UserID: &userID}
			if err := tx.Create(&userCart).Error; err != nil {
				if !IsDuplicatedKey(err) {
					return err
				}
				if err := txRepo.lockForUpdate(tx.Where("user_id = ?", userID)).First(&userCart).Error; err != nil {
					return err
				}
			}
		} else if err != nil {
			return err
		}

		var guestItems []product.CartItem
		if err := tx.Where("cart_id = ?", guest.ID).Order("id ASC").Find(&guestItems).Error; err != nil {
			return err
		}
		for _, guestItem := range guestItems {
			if strings.TrimSpace(guestItem.ConfigurationHash) == "" {
				guestItem.ConfigurationHash = product.DefaultConfigurationHash
			}
			if len(guestItem.ConfigurationData) == 0 {
				guestItem.ConfigurationData = product.DefaultConfigurationJSONBytes()
			}
			var existing product.CartItem
			itemQuery := tx.Where("cart_id = ? AND product_id = ?", userCart.ID, guestItem.ProductID)
			if guestItem.VariantID == nil {
				itemQuery = itemQuery.Where("variant_id IS NULL")
			} else {
				itemQuery = itemQuery.Where("variant_id = ?", *guestItem.VariantID)
			}
			configurationHash := strings.TrimSpace(guestItem.ConfigurationHash)
			if configurationHash == "" {
				configurationHash = product.DefaultConfigurationHash
			}
			itemQuery = itemQuery.Where("configuration_hash = ? OR (configuration_hash = '' AND ? = ?)", configurationHash, configurationHash, product.DefaultConfigurationHash)
			findErr := itemQuery.First(&existing).Error
			if findErr == nil {
				existing.Quantity += guestItem.Quantity
				existing.PriceMinor = guestItem.PriceMinor
				existing.Currency = guestItem.Currency
				existing.ConfigurationData = guestItem.ConfigurationData
				existing.ConfigurationHash = guestItem.ConfigurationHash
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				continue
			}
			if !IsRecordNotFound(findErr) {
				return findErr
			}
			guestItem.ID = 0
			guestItem.CartID = userCart.ID
			if err := tx.Create(&guestItem).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("cart_id = ?", guest.ID).Delete(&product.CartItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&guest).Error
	})
}

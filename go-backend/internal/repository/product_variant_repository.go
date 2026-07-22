package repository

import (
	"fmt"
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
)

func (r *ProductRepository) FindVariantBySKU(sku string) (*product.ProductVariant, error) {
	var variant product.ProductVariant
	if err := r.db.Where("sku = ?", sku).First(&variant).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}

func syncProductSummaryFromVariants(p *product.Product, variants []product.ProductVariant) {
	if len(variants) == 0 {
		return
	}

	defaultIndex := -1
	totalStock := 0
	for i, variant := range variants {
		if variant.IsActive {
			totalStock += variant.Stock
		}
		if variant.IsActive && variant.IsDefault {
			defaultIndex = i
		}
	}
	if defaultIndex == -1 {
		for i, variant := range variants {
			if variant.IsActive {
				defaultIndex = i
				break
			}
		}
	}
	if defaultIndex == -1 {
		defaultIndex = 0
	}

	defaultVariant := variants[defaultIndex]
	p.SKU = defaultVariant.SKU
	p.Price = defaultVariant.Price
	p.SalePrice = defaultVariant.SalePrice
	p.Stock = totalStock
}

func replaceProductVariants(tx *gorm.DB, productID uint, variants []product.ProductVariant) error {
	var existingVariants []product.ProductVariant
	if err := tx.Where("product_id = ?", productID).Find(&existingVariants).Error; err != nil {
		return err
	}

	existingByID := make(map[uint]product.ProductVariant, len(existingVariants))
	existingBySKU := make(map[string]product.ProductVariant, len(existingVariants))
	for _, variant := range existingVariants {
		existingByID[variant.ID] = variant
		existingBySKU[variant.SKU] = variant
	}

	keepIDs := make([]uint, 0, len(variants))
	for i := range variants {
		variants[i].ProductID = productID
		if variants[i].ID == 0 {
			if existing, ok := existingBySKU[variants[i].SKU]; ok {
				variants[i].ID = existing.ID
			}
		}

		if variants[i].ID != 0 {
			if _, ok := existingByID[variants[i].ID]; !ok {
				return fmt.Errorf("variant %d does not belong to product %d", variants[i].ID, productID)
			}
			if err := tx.Save(&variants[i]).Error; err != nil {
				return err
			}
			keepIDs = append(keepIDs, variants[i].ID)
			continue
		}

		if err := tx.Create(&variants[i]).Error; err != nil {
			return err
		}
		keepIDs = append(keepIDs, variants[i].ID)
	}

	deleteQuery := tx.Where("product_id = ?", productID)
	if len(keepIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", keepIDs)
	}
	return deleteQuery.Delete(&product.ProductVariant{}).Error
}

func (r *ProductRepository) FindPurchasableVariant(productID uint, variantID *uint) (*product.Product, *product.ProductVariant, error) {
	p, err := r.FindByID(productID)
	if err != nil {
		return nil, nil, err
	}

	activeVariants := p.ActiveVariants()
	if len(activeVariants) == 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}

	if variantID != nil {
		for i := range activeVariants {
			if activeVariants[i].ID == *variantID {
				return p, &activeVariants[i], nil
			}
		}
		return nil, nil, gorm.ErrRecordNotFound
	}

	if variant := p.DefaultVariant(); variant != nil {
		return p, variant, nil
	}

	return nil, nil, gorm.ErrRecordNotFound
}

func (r *ProductRepository) DecrementVariantStocks(items map[uint]int) error {
	if len(items) == 0 {
		return nil
	}

	for variantID, quantity := range items {
		res := r.db.Model(&product.ProductVariant{}).
			Where("id = ? AND is_active = ? AND stock >= ?", variantID, true, quantity).
			UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("insufficient stock for variant %d or variant not found", variantID)
		}
	}

	return r.refreshProductStockForVariants(items)
}

func (r *ProductRepository) IncrementVariantStock(variantID uint, quantity int) error {
	res := r.db.Model(&product.ProductVariant{}).Where("id = ?", variantID).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.refreshProductStockForVariants(map[uint]int{variantID: quantity})
}

func (r *ProductRepository) refreshProductStockForVariants(items map[uint]int) error {
	if len(items) == 0 {
		return nil
	}

	var productIDs []uint
	if err := r.db.Model(&product.ProductVariant{}).
		Where("id IN ?", uintMapKeys(items)).
		Distinct().
		Pluck("product_id", &productIDs).Error; err != nil {
		return err
	}

	for _, productID := range productIDs {
		var totalStock int64
		if err := r.db.Model(&product.ProductVariant{}).
			Where("product_id = ? AND is_active = ? AND deleted_at IS NULL", productID, true).
			Select("COALESCE(SUM(stock), 0)").
			Scan(&totalStock).Error; err != nil {
			return err
		}
		if err := r.db.Model(&product.Product{}).
			Where("id = ?", productID).
			Update("stock", totalStock).Error; err != nil {
			return err
		}
	}

	return nil
}

func uintMapKeys(items map[uint]int) []uint {
	keys := make([]uint, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	return keys
}

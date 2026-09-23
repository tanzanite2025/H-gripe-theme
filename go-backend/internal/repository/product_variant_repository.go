package repository

import (
	"commerce-platform/internal/domain/product"
	"fmt"
	"sort"

	"gorm.io/gorm"
)

func (r *ProductRepository) FindVariantBySKU(sku string) (*product.ProductVariant, error) {
	var variant product.ProductVariant
	if err := r.db.Where("sku = ?", sku).First(&variant).Error; err != nil {
		return nil, err
	}
	if err := r.attachVariantDisplayPriceSnapshot(&variant); err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *ProductRepository) FindVariantByID(id uint) (*product.ProductVariant, error) {
	var variant product.ProductVariant
	if err := r.db.Where("id = ?", id).First(&variant).Error; err != nil {
		return nil, err
	}
	if err := r.attachVariantDisplayPriceSnapshot(&variant); err != nil {
		return nil, err
	}
	return &variant, nil
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
			existing, ok := existingByID[variants[i].ID]
			if !ok {
				return fmt.Errorf("variant %d does not belong to product %d", variants[i].ID, productID)
			}
			// Preserve inventory ownership when an admin edits a translated
			// variant. The API input intentionally does not expose this field.
			if existing.MasterVariantID != nil && *existing.MasterVariantID > 0 {
				variants[i].MasterVariantID = existing.MasterVariantID
				variants[i].Stock = 0
			}
			// Display prices belong to the independent read model. Omit the
			// legacy column so catalog writes cannot reintroduce a stale snapshot.
			if err := tx.Omit("DisplayPriceData").Save(&variants[i]).Error; err != nil {
				return err
			}
			keepIDs = append(keepIDs, variants[i].ID)
			continue
		}

		isActive := variants[i].IsActive
		if err := tx.Omit("DisplayPriceData").Create(&variants[i]).Error; err != nil {
			return err
		}
		if !isActive {
			if err := tx.Model(&variants[i]).Update("is_active", false).Error; err != nil {
				return err
			}
			variants[i].IsActive = false
		}
		keepIDs = append(keepIDs, variants[i].ID)
	}

	deleteQuery := tx.Where("product_id = ?", productID)
	if len(keepIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", keepIDs)
	}
	return deleteQuery.Delete(&product.ProductVariant{}).Error
}

func replaceVariantOptionRules(tx *gorm.DB, variantID uint, groupRules []product.ProductOptionGroupVariantRule, valueRules []product.ProductOptionValueVariantRule) error {
	groupTable := tx.Migrator().HasTable(&product.ProductOptionGroupVariantRule{})
	valueTable := tx.Migrator().HasTable(&product.ProductOptionValueVariantRule{})
	if !groupTable && !valueTable {
		return nil
	}
	if groupTable {
		if err := tx.Where("variant_id = ?", variantID).Delete(&product.ProductOptionGroupVariantRule{}).Error; err != nil {
			return err
		}
	}
	if valueTable {
		if err := tx.Where("variant_id = ?", variantID).Delete(&product.ProductOptionValueVariantRule{}).Error; err != nil {
			return err
		}
	}
	if groupTable {
		for i := range groupRules {
			rule := groupRules[i]
			if err := tx.Model(&product.ProductOptionGroupVariantRule{}).Create(map[string]interface{}{
				"variant_id":              variantID,
				"spec_definition_id":      rule.SpecDefinitionID,
				"is_applicable":           rule.IsApplicable,
				"min_selections_override": rule.MinSelectionsOverride,
				"max_selections_override": rule.MaxSelectionsOverride,
			}).Error; err != nil {
				return err
			}
		}
	}
	if valueTable {
		for i := range valueRules {
			rule := valueRules[i]
			if err := tx.Model(&product.ProductOptionValueVariantRule{}).Create(map[string]interface{}{
				"variant_id":                      variantID,
				"product_variant_option_value_id": rule.ProductVariantOptionValueID,
				"is_enabled":                      rule.IsEnabled,
				"price_delta_minor_override":      rule.PriceDeltaMinorOverride,
				"unavailable_reason":              rule.UnavailableReason,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ProductRepository) FindPurchasableVariant(productID uint, variantID *uint) (*product.Product, *product.ProductVariant, error) {
	p, err := r.FindByID(productID)
	if err != nil {
		return nil, nil, err
	}
	if p.Status != "active" {
		return nil, nil, gorm.ErrRecordNotFound
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

func (r *ProductRepository) DecrementVariantStocks(items map[uint]int) ([]uint, error) {
	if len(items) == 0 {
		return nil, nil
	}

	resolvedItems := make(map[uint]int, len(items))
	for _, variantID := range uintMapKeys(items) {
		quantity := items[variantID]
		if quantity <= 0 {
			return nil, fmt.Errorf("invalid stock quantity %d for variant %d", quantity, variantID)
		}
		masterID, active, err := r.masterVariantID(variantID)
		if err != nil {
			return nil, err
		}
		if !active {
			return nil, fmt.Errorf("insufficient stock for variant %d or variant not found", variantID)
		}
		resolvedItems[masterID] += quantity
	}

	for _, masterID := range uintMapKeys(resolvedItems) {
		quantity := resolvedItems[masterID]
		res := r.db.Model(&product.ProductVariant{}).
			Where("id = ? AND is_active = ? AND stock >= ?", masterID, true, quantity).
			UpdateColumn("stock", gorm.Expr("stock - ?", quantity))
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			return nil, fmt.Errorf("insufficient stock for inventory variant %d or variant not found", masterID)
		}
	}

	return r.findAffectedProductIDsForInventoryVariants(uintMapKeys(resolvedItems))
}

// ValidateVariantStocks checks the inventory owned by the supplied variants
// (including translated variants that point at a master row) without mutating
// it. Callers must still perform the conditional decrement in the order
// transaction because this check is advisory under concurrency.
func (r *ProductRepository) ValidateVariantStocks(items map[uint]int) error {
	if len(items) == 0 {
		return nil
	}
	resolvedItems := make(map[uint]int, len(items))
	for _, variantID := range uintMapKeys(items) {
		quantity := items[variantID]
		if quantity <= 0 {
			return fmt.Errorf("invalid stock quantity %d for variant %d", quantity, variantID)
		}
		masterID, active, err := r.masterVariantID(variantID)
		if err != nil {
			return err
		}
		if !active {
			return fmt.Errorf("insufficient stock for variant %d or variant not found", variantID)
		}
		resolvedItems[masterID] += quantity
	}
	for _, masterID := range uintMapKeys(resolvedItems) {
		var row struct {
			IsActive bool
			Stock    int
		}
		if err := r.db.Model(&product.ProductVariant{}).Select("is_active, stock").Where("id = ?", masterID).First(&row).Error; err != nil {
			return err
		}
		if !row.IsActive || row.Stock < resolvedItems[masterID] {
			return fmt.Errorf("insufficient stock for inventory variant %d or variant not found", masterID)
		}
	}
	return nil
}

// IncrementVariantStocks restores several inventory variants in a stable
// order. Quantities are aggregated first so a component selected by multiple
// option values is restored exactly once.
func (r *ProductRepository) IncrementVariantStocks(items map[uint]int) ([]uint, error) {
	if len(items) == 0 {
		return nil, nil
	}
	resolvedItems := make(map[uint]int, len(items))
	for _, variantID := range uintMapKeys(items) {
		quantity := items[variantID]
		if quantity <= 0 {
			return nil, fmt.Errorf("invalid stock quantity %d for variant %d", quantity, variantID)
		}
		masterID, _, err := r.masterVariantID(variantID)
		if err != nil {
			return nil, err
		}
		resolvedItems[masterID] += quantity
	}
	for _, masterID := range uintMapKeys(resolvedItems) {
		quantity := resolvedItems[masterID]
		res := r.db.Model(&product.ProductVariant{}).Where("id = ?", masterID).
			UpdateColumn("stock", gorm.Expr("stock + ?", quantity))
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			return nil, gorm.ErrRecordNotFound
		}
	}
	return r.findAffectedProductIDsForInventoryVariants(uintMapKeys(resolvedItems))
}

func (r *ProductRepository) IncrementVariantStock(variantID uint, quantity int) ([]uint, error) {
	return r.IncrementVariantStocks(map[uint]int{variantID: quantity})
}

func (r *ProductRepository) masterVariantID(variantID uint) (uint, bool, error) {
	var variant struct {
		ID              uint
		MasterVariantID *uint
		IsActive        bool
	}
	if err := r.db.Model(&product.ProductVariant{}).Select("id, master_variant_id, is_active").Where("id = ?", variantID).First(&variant).Error; err != nil {
		return 0, false, err
	}
	if variant.MasterVariantID != nil && *variant.MasterVariantID > 0 {
		return *variant.MasterVariantID, variant.IsActive, nil
	}
	return variant.ID, variant.IsActive, nil
}

func (r *ProductRepository) findAffectedProductIDsForInventoryVariants(variantIDs []uint) ([]uint, error) {
	if len(variantIDs) == 0 {
		return nil, nil
	}

	var productIDs []uint
	if err := r.db.Model(&product.ProductVariant{}).
		Where("id IN ? OR master_variant_id IN ?", variantIDs, variantIDs).
		Distinct().
		Pluck("product_id", &productIDs).Error; err != nil {
		return nil, err
	}

	sortUintIDs(productIDs)
	return productIDs, nil
}

func uintMapKeys(items map[uint]int) []uint {
	keys := make([]uint, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sortUintIDs(keys)
	return keys
}

func sortUintIDs(ids []uint) {
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
}

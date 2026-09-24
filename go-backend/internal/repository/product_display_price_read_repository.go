package repository

import (
	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/product"

	"gorm.io/datatypes"
)

// attachProductDisplayPriceSnapshots hydrates the storefront read model with
// one snapshot query for the complete product result set. Core catalog rows
// are deliberately ignored, even while their legacy columns still exist.
func (r *ProductRepository) attachProductDisplayPriceSnapshots(products []product.Product) error {
	if len(products) == 0 {
		return nil
	}

	productIDs := make([]uint, 0, len(products))
	for i := range products {
		products[i].DisplayPriceData = datatypes.JSON([]byte("[]"))
		products[i].DisplayPriceSnapshot = nil
		productIDs = append(productIDs, products[i].ID)
		for j := range products[i].Variants {
			products[i].Variants[j].DisplayPriceData = datatypes.JSON([]byte("[]"))
			products[i].Variants[j].DisplayPriceSnapshot = nil
		}
	}

	if !r.db.Migrator().HasTable(&product.ProductDisplayPriceSnapshot{}) {
		return nil
	}

	var snapshots []product.ProductDisplayPriceSnapshot
	if err := r.db.Where("product_id IN ?", productIDs).Find(&snapshots).Error; err != nil {
		return err
	}

	productsByID := make(map[uint]*product.Product, len(products))
	variantsByID := make(map[uint]*product.ProductVariant)
	for i := range products {
		item := &products[i]
		productsByID[item.ID] = item
		for j := range item.Variants {
			variantsByID[item.Variants[j].ID] = &item.Variants[j]
		}
	}

	for i := range snapshots {
		snapshot := &snapshots[i]
		if snapshot.VariantID != nil {
			variant := variantsByID[*snapshot.VariantID]
			if variant != nil && displayPriceSnapshotMatchesSource(snapshot, variant.Currency, variant.PriceMinor, variant.SalePriceMinor) {
				variant.DisplayPriceData = append(datatypes.JSON(nil), snapshot.DisplayPriceData...)
				variant.DisplayPriceSnapshot = snapshot
			}
			continue
		}

		item := productsByID[snapshot.ProductID]
		if item == nil {
			continue
		}
		sourceCurrency, sourcePriceMinor, sourceSalePriceMinor := item.Currency, item.PriceMinor, item.SalePriceMinor
		if variant := item.StartingPriceVariant(); variant != nil {
			sourceCurrency, sourcePriceMinor, sourceSalePriceMinor = variant.Currency, variant.PriceMinor, variant.SalePriceMinor
		}
		if displayPriceSnapshotMatchesSource(snapshot, sourceCurrency, sourcePriceMinor, sourceSalePriceMinor) {
			item.DisplayPriceData = append(datatypes.JSON(nil), snapshot.DisplayPriceData...)
			item.DisplayPriceSnapshot = snapshot
		}
	}
	return nil
}

func (r *ProductRepository) attachProductDisplayPriceSnapshot(item *product.Product) error {
	if item == nil {
		return nil
	}
	items := []product.Product{*item}
	if err := r.attachProductDisplayPriceSnapshots(items); err != nil {
		return err
	}
	*item = items[0]
	return nil
}

func (r *ProductRepository) attachVariantDisplayPriceSnapshot(variant *product.ProductVariant) error {
	if variant == nil {
		return nil
	}
	items := []product.Product{{ID: variant.ProductID, Variants: []product.ProductVariant{*variant}}}
	if err := r.attachProductDisplayPriceSnapshots(items); err != nil {
		return err
	}
	*variant = items[0].Variants[0]
	return nil
}

func displayPriceSnapshotMatchesSource(snapshot *product.ProductDisplayPriceSnapshot, sourceCurrency string, sourcePriceMinor int64, sourceSalePriceMinor *int64) bool {
	return snapshot != nil &&
		snapshot.SourcePriceMinor == sourcePriceMinor &&
		currency.NormalizeCode(snapshot.SourceCurrency) == currency.NormalizeCode(sourceCurrency) &&
		salePriceMinorEqual(snapshot.SourceSalePriceMinor, sourceSalePriceMinor)
}

package product

import (
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProductDisplayPriceSnapshot is a storefront read model. Source pricing
// remains owned by products/product_variants; this table is rebuilt from that
// source and is never part of the transactional product update.
type ProductDisplayPriceSnapshot struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	ScopeKey             string         `gorm:"size:64;uniqueIndex;not null" json:"-"`
	ProductID            uint           `gorm:"not null;index" json:"product_id"`
	VariantID            *uint          `gorm:"index" json:"variant_id,omitempty"`
	SourceCurrency       string         `gorm:"size:3;not null" json:"source_currency"`
	SourcePriceMinor     int64          `gorm:"not null" json:"source_price_minor"`
	SourceSalePriceMinor *int64         `gorm:"column:source_sale_price_minor" json:"source_sale_price_minor,omitempty"`
	DisplayPriceData     datatypes.JSON `gorm:"type:json;not null;default:'[]'" json:"display_prices"`
	CreatedAt            time.Time      `json:"-"`
	UpdatedAt            time.Time      `json:"-"`
}

func (ProductDisplayPriceSnapshot) TableName() string { return "product_display_price_snapshots" }

func productDisplaySnapshotTableReady(tx *gorm.DB) bool {
	return tx != nil && tx.Migrator().HasTable(&ProductDisplayPriceSnapshot{})
}

func loadProductDisplayPriceSnapshot(tx *gorm.DB, productID uint, variantID *uint) (*ProductDisplayPriceSnapshot, error) {
	if !productDisplaySnapshotTableReady(tx) || productID == 0 {
		return nil, nil
	}
	query := tx.Where("product_id = ?", productID)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	var snapshot ProductDisplayPriceSnapshot
	if err := query.First(&snapshot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &snapshot, nil
}

func (p *Product) AfterFind(tx *gorm.DB) error {
	snapshot, err := loadProductDisplayPriceSnapshot(tx, p.ID, nil)
	if err != nil {
		return err
	}
	if snapshot != nil && snapshot.SourcePriceMinor == p.PriceMinor && snapshot.SourceCurrency == p.Currency && sameSalePriceMinor(snapshot.SourceSalePriceMinor, p.SalePriceMinor) {
		p.DisplayPriceData = append(datatypes.JSON(nil), snapshot.DisplayPriceData...)
		p.DisplayPriceSnapshot = snapshot
	}
	return nil
}

func (v *ProductVariant) loadDisplayPriceSnapshot(tx *gorm.DB) error {
	snapshot, err := loadProductDisplayPriceSnapshot(tx, v.ProductID, &v.ID)
	if err != nil {
		return err
	}
	if snapshot != nil && snapshot.SourcePriceMinor == v.PriceMinor && snapshot.SourceCurrency == v.Currency && sameSalePriceMinor(snapshot.SourceSalePriceMinor, v.SalePriceMinor) {
		v.DisplayPriceData = append(datatypes.JSON(nil), snapshot.DisplayPriceData...)
		v.DisplayPriceSnapshot = snapshot
	}
	return nil
}

func sameSalePriceMinor(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

package product

import "time"

// This receipt deliberately survives product deletion and is preserved, rather
// than imported from a product-only package, during partial recovery.
type ProductViewCountFlushReceipt struct {
	BatchID                 string    `gorm:"type:uuid;primaryKey"`
	ProductID               uint      `gorm:"primaryKey;autoIncrement:false"`
	Delta                   int64     `gorm:"not null"`
	ProductIncrementApplied bool      `gorm:"not null"`
	CreatedAt               time.Time `gorm:"not null"`
}

func (ProductViewCountFlushReceipt) TableName() string { return "product_view_count_flush_receipts" }

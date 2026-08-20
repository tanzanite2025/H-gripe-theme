package productiontest

import "time"

// ProductGate marks a product or variant as production-test-only.
type ProductGate struct {
	ID                   uint       `gorm:"primarykey" json:"id"`
	ProductID            uint       `gorm:"not null;index" json:"product_id"`
	VariantID            *uint      `gorm:"index" json:"variant_id,omitempty"`
	IsTestOnly           bool       `gorm:"not null;default:true;index" json:"is_test_only"`
	AllowedTestAccountID *uint      `gorm:"index" json:"allowed_test_account_id,omitempty"`
	Enabled              bool       `gorm:"not null;default:true;index" json:"enabled"`
	HoldFulfillment      bool       `gorm:"not null;default:true" json:"hold_fulfillment"`
	Reason               string     `gorm:"type:text;not null;default:''" json:"reason"`
	StartsAt             *time.Time `gorm:"index" json:"starts_at,omitempty"`
	EndsAt               *time.Time `gorm:"index" json:"ends_at,omitempty"`
	CreatedBy            uint       `gorm:"not null;index" json:"created_by"`
	UpdatedBy            uint       `gorm:"not null;index" json:"updated_by"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (ProductGate) TableName() string {
	return "production_test_product_gates"
}

func (g ProductGate) IsActiveAt(now time.Time) bool {
	if !g.Enabled || !g.IsTestOnly || g.ProductID == 0 {
		return false
	}
	if g.StartsAt != nil && now.Before(*g.StartsAt) {
		return false
	}
	if g.EndsAt != nil && !g.EndsAt.After(now) {
		return false
	}
	return true
}

func (g ProductGate) AppliesTo(item PurchaseItem, now time.Time) bool {
	if !g.IsActiveAt(now) || g.ProductID != item.ProductID {
		return false
	}
	if g.VariantID == nil {
		return true
	}
	return item.VariantID != nil && *g.VariantID == *item.VariantID
}

func (g ProductGate) Allows(account TestAccount) bool {
	if g.AllowedTestAccountID == nil {
		return true
	}
	return account.ID != 0 && account.ID == *g.AllowedTestAccountID
}


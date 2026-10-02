package product

import "time"

// SpokeRepairKitModel is the product-owned compatibility choice shown to a
// buyer on a spoke repair-kit product. Names and lifecycle are snapshots of
// the checked catalog at the time the product is saved.
type SpokeRepairKitModel struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	ProductID         uint      `gorm:"not null;index;uniqueIndex:idx_product_spoke_repair_kit_model" json:"product_id"`
	BrandSlug         string    `gorm:"type:varchar(120);not null;uniqueIndex:idx_product_spoke_repair_kit_model" json:"brand_slug"`
	BrandName         string    `gorm:"type:varchar(160);not null" json:"brand_name"`
	WheelsetModelSlug string    `gorm:"type:varchar(160);not null;uniqueIndex:idx_product_spoke_repair_kit_model" json:"wheelset_model_slug"`
	WheelsetModelName string    `gorm:"type:varchar(255);not null" json:"wheelset_model_name"`
	LifecycleStatus   string    `gorm:"type:varchar(32);not null;default:'current'" json:"lifecycle_status"`
	SourceCheckedAt   string    `gorm:"type:varchar(40)" json:"source_checked_at,omitempty"`
	SortOrder         int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (SpokeRepairKitModel) TableName() string {
	return "product_spoke_repair_kit_models"
}

func (m SpokeRepairKitModel) BuildSpokeRepairKitModelCompatibilityValueKey() string {
	return m.BrandSlug + ":" + m.WheelsetModelSlug
}

func (m SpokeRepairKitModel) Label() string {
	if m.BrandName == "" {
		return m.WheelsetModelName
	}
	return m.BrandName + " / " + m.WheelsetModelName
}

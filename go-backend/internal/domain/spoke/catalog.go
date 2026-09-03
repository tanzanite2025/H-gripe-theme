package spoke

import (
	"time"

	productdomain "commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

type CatalogRimBrand struct {
	ID             uint                       `gorm:"primarykey"`
	Code           string                     `gorm:"size:80;not null;uniqueIndex:idx_spoke_rim_brands_code"`
	Name           string                     `gorm:"size:160;not null"`
	ProductBrandID *uint                      `gorm:"column:product_brand_id;uniqueIndex:idx_spoke_rim_brands_product_brand_id;index" json:"-"`
	ProductBrand   productdomain.ProductBrand `gorm:"foreignKey:ProductBrandID;references:ID" json:"-"`
	SortOrder      int                        `gorm:"not null;default:0;index"`
	IsEnabled      bool                       `gorm:"not null;default:true;index"`
	Models         []CatalogRimModel          `gorm:"foreignKey:BrandID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (CatalogRimBrand) TableName() string {
	return "spoke_rim_brands"
}

type CatalogRimModel struct {
	ID        uint            `gorm:"primarykey"`
	BrandID   uint            `gorm:"not null;index"`
	Code      string          `gorm:"size:120;not null;uniqueIndex:idx_spoke_rim_models_code"`
	Name      string          `gorm:"size:180;not null"`
	ERD       *float64        `gorm:"column:erd_mm"`
	Weight    *float64        `gorm:"column:weight_g"`
	SortOrder int             `gorm:"not null;default:0;index"`
	IsEnabled bool            `gorm:"not null;default:true;index"`
	Brand     CatalogRimBrand `gorm:"foreignKey:BrandID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CatalogRimModel) TableName() string {
	return "spoke_rim_models"
}

type CatalogHubBrand struct {
	ID        uint              `gorm:"primarykey"`
	Code      string            `gorm:"size:80;not null;uniqueIndex:idx_spoke_hub_brands_code"`
	Name      string            `gorm:"size:160;not null"`
	SortOrder int               `gorm:"not null;default:0;index"`
	IsEnabled bool              `gorm:"not null;default:true;index"`
	Models    []CatalogHubModel `gorm:"foreignKey:BrandID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CatalogHubBrand) TableName() string {
	return "spoke_hub_brands"
}

type CatalogHubModel struct {
	ID        uint            `gorm:"primarykey"`
	BrandID   uint            `gorm:"not null;index"`
	Code      string          `gorm:"size:120;not null;uniqueIndex:idx_spoke_hub_models_code"`
	Name      string          `gorm:"size:180;not null"`
	SortOrder int             `gorm:"not null;default:0;index"`
	IsEnabled bool            `gorm:"not null;default:true;index"`
	Brand     CatalogHubBrand `gorm:"foreignKey:BrandID"`

	FitmentHubSpecificationID *uint    `gorm:"column:fitment_hub_specification_id;index" json:"-"`
	FrontLeftFlange           *float64 `gorm:"column:front_left_flange_mm"`
	FrontRightFlange          *float64 `gorm:"column:front_right_flange_mm"`
	FrontLeftFlangePCD        *float64 `gorm:"column:front_left_flange_pcd_mm"`
	FrontRightFlangePCD       *float64 `gorm:"column:front_right_flange_pcd_mm"`
	FrontSpokeHoleDiameter    *float64 `gorm:"column:front_spoke_hole_diameter_mm"`
	RearLeftFlange            *float64 `gorm:"column:rear_left_flange_mm"`
	RearRightFlange           *float64 `gorm:"column:rear_right_flange_mm"`
	RearLeftFlangePCD         *float64 `gorm:"column:rear_left_flange_pcd_mm"`
	RearRightFlangePCD        *float64 `gorm:"column:rear_right_flange_pcd_mm"`
	RearSpokeHoleDiameter     *float64 `gorm:"column:rear_spoke_hole_diameter_mm"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CatalogHubModel) TableName() string {
	return "spoke_hub_models"
}

type CatalogBuildPreset struct {
	ID                     uint            `gorm:"primarykey"`
	Code                   string          `gorm:"size:140;not null;uniqueIndex:idx_spoke_build_presets_code"`
	Name                   string          `gorm:"size:220;not null"`
	Description            string          `gorm:"type:text"`
	KeywordsJSON           string          `gorm:"column:keywords_json;type:text;not null;default:'[]'"`
	RimBrandID             uint            `gorm:"not null;index"`
	RimModelID             uint            `gorm:"not null;index"`
	HubBrandID             uint            `gorm:"not null;index"`
	HubModelID             uint            `gorm:"not null;index"`
	WheelPosition          string          `gorm:"size:16;not null;default:'auto';index"`
	SpokeCount             int             `gorm:"not null"`
	Crossing               int             `gorm:"not null;default:0"`
	NippleType             string          `gorm:"size:20;not null;default:'standard'"`
	NippleLength           *float64        `gorm:"column:nipple_length_mm"`
	ActualFrontLeftLength  *float64        `gorm:"column:actual_front_left_length_mm"`
	ActualFrontRightLength *float64        `gorm:"column:actual_front_right_length_mm"`
	ActualRearLeftLength   *float64        `gorm:"column:actual_rear_left_length_mm"`
	ActualRearRightLength  *float64        `gorm:"column:actual_rear_right_length_mm"`
	ActualLengthNotes      string          `gorm:"column:actual_length_notes;type:text"`
	SortOrder              int             `gorm:"not null;default:0;index"`
	IsEnabled              bool            `gorm:"not null;default:true;index"`
	RimBrand               CatalogRimBrand `gorm:"foreignKey:RimBrandID"`
	RimModel               CatalogRimModel `gorm:"foreignKey:RimModelID"`
	HubBrand               CatalogHubBrand `gorm:"foreignKey:HubBrandID"`
	HubModel               CatalogHubModel `gorm:"foreignKey:HubModelID"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              gorm.DeletedAt `gorm:"index"`
}

func (CatalogBuildPreset) TableName() string {
	return "spoke_build_presets"
}

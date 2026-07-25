package spoke

import (
	"time"

	"gorm.io/gorm"
)

type History struct {
	ID                    uint           `gorm:"primarykey" json:"id"`
	UserID                *uint          `gorm:"index" json:"user_id,omitempty"`
	WheelType             *string        `gorm:"size:40;index" json:"wheel_type"`
	SourceType            *string        `gorm:"size:40;index" json:"source_type"`
	RimBrand              *string        `gorm:"size:120;index" json:"rim_brand"`
	RimModel              *string        `gorm:"size:160;index" json:"rim_model"`
	HubBrand              *string        `gorm:"size:120;index" json:"hub_brand"`
	HubModel              *string        `gorm:"size:160;index" json:"hub_model"`
	ERDMM                 *float64       `json:"erd_mm"`
	LeftFlangePCDMM       *float64       `json:"left_flange_pcd_mm"`
	RightFlangePCDMM      *float64       `json:"right_flange_pcd_mm"`
	LeftFlangeToCenterMM  *float64       `json:"left_flange_to_center_mm"`
	RightFlangeToCenterMM *float64       `json:"right_flange_to_center_mm"`
	SpokeCount            *int           `json:"spoke_count"`
	LacingPattern         *string        `gorm:"size:60" json:"lacing_pattern"`
	NippleType            *string        `gorm:"size:60" json:"nipple_type"`
	LeftLengthMM          *float64       `json:"left_length_mm"`
	RightLengthMM         *float64       `json:"right_length_mm"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (History) TableName() string {
	return "spoke_histories"
}

package fitmentcatalog

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type HubPosition string

const (
	HubPositionFront HubPosition = "front"
	HubPositionRear  HubPosition = "rear"
)

type HubAxleType string

const (
	HubAxleTypeQuickRelease HubAxleType = "quick_release"
	HubAxleTypeThruAxle     HubAxleType = "thru_axle"
	HubAxleTypeBoltOn       HubAxleType = "bolt_on"
	HubAxleTypeOther        HubAxleType = "other"
)

type HubSpecification struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	SpecCode      string         `gorm:"size:80;not null" json:"spec_code"`
	DisplayName   string         `gorm:"size:160;not null" json:"display_name"`
	Position      HubPosition    `gorm:"size:16;not null" json:"position"`
	AxleType      HubAxleType    `gorm:"size:32;not null" json:"axle_type"`
	AxleSpacingMM int            `gorm:"not null" json:"axle_spacing_mm"`
	Notes         string         `gorm:"type:text" json:"notes"`
	IsEnabled     bool           `gorm:"not null;default:false;index" json:"is_enabled"`
	SortOrder     int            `gorm:"not null;default:0;index" json:"sort_order"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	FrameReferenceCount int `gorm:"-" json:"frame_reference_count"`
	ForkReferenceCount  int `gorm:"-" json:"fork_reference_count"`
}

func (HubSpecification) TableName() string {
	return "fitment_hub_specifications"
}

func (specification *HubSpecification) BeforeCreate(tx *gorm.DB) error {
	return specification.Normalize()
}

func (specification *HubSpecification) BeforeSave(tx *gorm.DB) error {
	return specification.Normalize()
}

func (specification *HubSpecification) Normalize() error {
	if specification == nil {
		return fmt.Errorf("hub specification is nil")
	}

	specification.SpecCode = strings.ToUpper(strings.TrimSpace(specification.SpecCode))
	specification.DisplayName = strings.TrimSpace(specification.DisplayName)
	specification.Position = HubPosition(strings.ToLower(strings.TrimSpace(string(specification.Position))))
	specification.AxleType = HubAxleType(strings.ToLower(strings.TrimSpace(string(specification.AxleType))))
	specification.Notes = strings.TrimSpace(specification.Notes)

	return nil
}

func (specification *HubSpecification) Validate() error {
	if specification == nil {
		return fmt.Errorf("hub specification is nil")
	}
	if err := specification.Normalize(); err != nil {
		return err
	}
	if specification.SpecCode == "" {
		return fmt.Errorf("spec_code is required")
	}
	if specification.DisplayName == "" {
		return fmt.Errorf("display_name is required")
	}
	if len(specification.SpecCode) > 80 {
		return fmt.Errorf("spec_code is too long")
	}
	if len(specification.DisplayName) > 160 {
		return fmt.Errorf("display_name is too long")
	}
	if !isSupportedHubPosition(specification.Position) {
		return fmt.Errorf("unsupported position %q", specification.Position)
	}
	if !isSupportedHubAxleType(specification.AxleType) {
		return fmt.Errorf("unsupported axle_type %q", specification.AxleType)
	}
	if specification.AxleSpacingMM <= 0 {
		return fmt.Errorf("axle_spacing_mm must be positive")
	}
	if specification.SortOrder < 0 {
		return fmt.Errorf("sort_order must be non-negative")
	}

	return nil
}

func isSupportedHubPosition(position HubPosition) bool {
	switch position {
	case HubPositionFront, HubPositionRear:
		return true
	default:
		return false
	}
}

func isSupportedHubAxleType(axleType HubAxleType) bool {
	switch axleType {
	case HubAxleTypeQuickRelease, HubAxleTypeThruAxle, HubAxleTypeBoltOn, HubAxleTypeOther:
		return true
	default:
		return false
	}
}

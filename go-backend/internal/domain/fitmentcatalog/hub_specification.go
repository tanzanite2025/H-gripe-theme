package fitmentcatalog

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

var spokeCatalogSpecCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{1,79}$`)

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
	WRMM          *float64       `gorm:"column:wr_mm" json:"wr_mm"`
	WLMM          *float64       `gorm:"column:wl_mm" json:"wl_mm"`
	PCDRMM        *float64       `gorm:"column:pcdr_mm" json:"pcdr_mm"`
	PCDLMM        *float64       `gorm:"column:pcdl_mm" json:"pcdl_mm"`
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
	if !IsValidSpokeCatalogSpecCode(specification.SpecCode) {
		return fmt.Errorf("spec_code must use 2-80 letters, numbers, underscores or hyphens and start with a letter or number")
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
	if err := validateSpokeGeometry(specification); err != nil {
		return err
	}
	if specification.SortOrder < 0 {
		return fmt.Errorf("sort_order must be non-negative")
	}

	return nil
}

func IsValidSpokeCatalogSpecCode(value string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	return spokeCatalogSpecCodePattern.MatchString(normalized)
}

func isSupportedHubPosition(position HubPosition) bool {
	switch position {
	case HubPositionFront, HubPositionRear:
		return true
	default:
		return false
	}
}

func validateSpokeGeometry(specification *HubSpecification) error {
	values := []*float64{
		specification.WRMM,
		specification.WLMM,
		specification.PCDRMM,
		specification.PCDLMM,
	}

	filled := 0
	for _, value := range values {
		if value != nil {
			filled++
		}
	}
	if filled != 0 && filled != len(values) {
		return fmt.Errorf("wr_mm, wl_mm, pcdr_mm and pcdl_mm must be provided together")
	}

	for _, field := range []struct {
		name         string
		value        *float64
		min          float64
		max          float64
		minExclusive bool
	}{
		{name: "wr_mm", value: specification.WRMM, min: 0, max: 100, minExclusive: true},
		{name: "wl_mm", value: specification.WLMM, min: 0, max: 100, minExclusive: true},
		{name: "pcdr_mm", value: specification.PCDRMM, min: 10, max: 150},
		{name: "pcdl_mm", value: specification.PCDLMM, min: 10, max: 150},
	} {
		if field.value == nil {
			continue
		}
		minInvalid := *field.value < field.min
		if field.minExclusive {
			minInvalid = *field.value <= field.min
		}
		if math.IsNaN(*field.value) || math.IsInf(*field.value, 0) || minInvalid || *field.value > field.max {
			if field.minExclusive {
				return fmt.Errorf("%s must be greater than %.0f and no greater than %.0f", field.name, field.min, field.max)
			}
			return fmt.Errorf("%s must be at least %.0f and no greater than %.0f", field.name, field.min, field.max)
		}
	}

	return nil
}

func isSupportedHubAxleType(axleType HubAxleType) bool {
	switch axleType {
	case HubAxleTypeQuickRelease, HubAxleTypeThruAxle, HubAxleTypeBoltOn, HubAxleTypeOther:
		return true
	default:
		return false
	}
}

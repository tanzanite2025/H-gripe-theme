package fitmentcatalog

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ForkFitmentEntry struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	BrandName      string         `gorm:"size:160;not null" json:"brand_name"`
	ModelName      string         `gorm:"size:160;not null" json:"model_name"`
	SeriesName     string         `gorm:"size:160" json:"series_name"`
	GenerationName string         `gorm:"size:160" json:"generation_name"`
	YearMode       YearMode       `gorm:"size:16;not null;default:'unknown'" json:"year_mode"`
	YearFrom       *int           `json:"year_from"`
	YearTo         *int           `json:"year_to"`
	MarketCode     string         `gorm:"size:32" json:"market_code"`
	Notes          string         `gorm:"type:text" json:"notes"`
	IsEnabled      bool           `gorm:"not null;default:false;index" json:"is_enabled"`
	SortOrder      int            `gorm:"not null;default:0;index" json:"sort_order"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	HubSpecifications     []HubSpecification `gorm:"-" json:"hub_specifications"`
	HubSpecificationCount int                `gorm:"-" json:"hub_specification_count"`
}

func (ForkFitmentEntry) TableName() string {
	return "fitment_fork_entries"
}

func (entry *ForkFitmentEntry) BeforeCreate(tx *gorm.DB) error {
	return entry.Normalize()
}

func (entry *ForkFitmentEntry) BeforeSave(tx *gorm.DB) error {
	return entry.Normalize()
}

func (entry *ForkFitmentEntry) Normalize() error {
	if entry == nil {
		return fmt.Errorf("fork fitment entry is nil")
	}

	entry.BrandName = strings.TrimSpace(entry.BrandName)
	entry.ModelName = strings.TrimSpace(entry.ModelName)
	entry.SeriesName = strings.TrimSpace(entry.SeriesName)
	entry.GenerationName = strings.TrimSpace(entry.GenerationName)
	entry.MarketCode = strings.ToUpper(strings.TrimSpace(entry.MarketCode))
	entry.Notes = strings.TrimSpace(entry.Notes)

	if entry.YearMode == "" {
		entry.YearMode = YearModeUnknown
	}

	return nil
}

func (entry *ForkFitmentEntry) Validate() error {
	if entry == nil {
		return fmt.Errorf("fork fitment entry is nil")
	}
	if err := entry.Normalize(); err != nil {
		return err
	}
	if entry.BrandName == "" {
		return fmt.Errorf("brand_name is required")
	}
	if entry.ModelName == "" {
		return fmt.Errorf("model_name is required")
	}
	if len(entry.BrandName) > 160 {
		return fmt.Errorf("brand_name is too long")
	}
	if len(entry.ModelName) > 160 {
		return fmt.Errorf("model_name is too long")
	}
	if len(entry.SeriesName) > 160 {
		return fmt.Errorf("series_name is too long")
	}
	if len(entry.GenerationName) > 160 {
		return fmt.Errorf("generation_name is too long")
	}
	if len(entry.MarketCode) > 32 {
		return fmt.Errorf("market_code is too long")
	}
	if entry.SortOrder < 0 {
		return fmt.Errorf("sort_order must be non-negative")
	}

	switch entry.YearMode {
	case YearModeSingle:
		if entry.YearFrom == nil || entry.YearTo != nil {
			return fmt.Errorf("single year mode requires year_from and forbids year_to")
		}
		if err := validateYear(*entry.YearFrom); err != nil {
			return err
		}
	case YearModeRange:
		if entry.YearFrom == nil || entry.YearTo == nil {
			return fmt.Errorf("range year mode requires year_from and year_to")
		}
		if err := validateYear(*entry.YearFrom); err != nil {
			return err
		}
		if err := validateYear(*entry.YearTo); err != nil {
			return err
		}
		if *entry.YearFrom > *entry.YearTo {
			return fmt.Errorf("year_from must not be greater than year_to")
		}
	case YearModeAll, YearModeUnknown:
		if entry.YearFrom != nil || entry.YearTo != nil {
			return fmt.Errorf("%s year mode forbids year_from and year_to", entry.YearMode)
		}
	default:
		return fmt.Errorf("unsupported year_mode %q", entry.YearMode)
	}

	return nil
}

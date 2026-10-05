package shipping

import (
	"errors"
	"strings"
	"time"
)

// YanwenCountryCatalogEntry caches a country returned by Yanwen's official
// common.country.getlist endpoint for one gateway environment.
type YanwenCountryCatalogEntry struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Environment  string    `gorm:"size:16;not null;uniqueIndex:idx_yanwen_country_environment" json:"environment"`
	CountryID    string    `gorm:"column:country_id;size:80;not null;uniqueIndex:idx_yanwen_country_environment" json:"country_id"`
	CountryCode  string    `gorm:"column:country_code;size:16;not null" json:"country_code"`
	NameChinese  string    `gorm:"column:name_ch;size:200;not null;default:''" json:"name_ch"`
	NameEnglish  string    `gorm:"column:name_en;size:200;not null;default:''" json:"name_en"`
	LastSyncedAt time.Time `gorm:"column:last_synced_at;not null" json:"last_synced_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (YanwenCountryCatalogEntry) TableName() string {
	return "shipping_yanwen_country_catalog_entries"
}

func (entry *YanwenCountryCatalogEntry) Validate() error {
	if entry == nil {
		return errors.New("Yanwen country catalog entry is required")
	}
	entry.Environment = strings.ToLower(strings.TrimSpace(entry.Environment))
	entry.CountryID = strings.TrimSpace(entry.CountryID)
	entry.CountryCode = strings.ToUpper(strings.TrimSpace(entry.CountryCode))
	entry.NameChinese = strings.TrimSpace(entry.NameChinese)
	entry.NameEnglish = strings.TrimSpace(entry.NameEnglish)
	if entry.Environment != "fat" && entry.Environment != "production" {
		return errors.New("Yanwen country environment must be fat or production")
	}
	if entry.CountryID == "" {
		return errors.New("Yanwen country id is required")
	}
	if entry.CountryCode == "" {
		return errors.New("Yanwen country code is required")
	}
	if entry.NameChinese == "" && entry.NameEnglish == "" {
		return errors.New("Yanwen country name is required")
	}
	if entry.LastSyncedAt.IsZero() {
		return errors.New("Yanwen country last synced time is required")
	}
	return nil
}
